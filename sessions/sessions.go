/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>
*/

package sessions

import (
	"errors"
	"fmt"
	"math/rand"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/cenkalti/rpc2"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/guardian"

	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

var (
	// ErrForcedDisconnect is used to specify the reason why the session was disconnected
	ErrForcedDisconnect = errors.New("FORCED_DISCONNECT")
)

// NewSessionS constructs  a new SessionS instance
func NewSessionS(cgrCfg *config.CGRConfig,
	dm *engine.DataManager,
	connMgr *engine.ConnManager) *SessionS {
	cgrCfg.SessionSCfg().SessionIndexes[utils.OriginID] = true // Make sure we have indexing for OriginID since it is a requirement on prefix searching

	return &SessionS{
		cgrCfg:        cgrCfg,
		dm:            dm,
		connMgr:       connMgr,
		biJClnts:      make(map[rpcclient.ClientConnector]string),
		biJIDs:        make(map[string]*biJClient),
		aSessions:     make(map[string]*Session),
		aSessionsIdx:  make(map[string]map[string]map[string]utils.StringMap),
		aSessionsRIdx: make(map[string][]*riFieldNameVal),
		pSessions:     make(map[string]*Session),
		pSessionsIdx:  make(map[string]map[string]map[string]utils.StringMap),
		pSessionsRIdx: make(map[string][]*riFieldNameVal),
	}
}

// biJClient contains info we need to reach back a bidirectional json client
type biJClient struct {
	conn  rpcclient.ClientConnector // connection towards BiJ client
	proto float64                   // client protocol version
}

// SessionS represents the session service
type SessionS struct {
	cgrCfg  *config.CGRConfig // Separate from smCfg since there can be multiple
	dm      *engine.DataManager
	connMgr *engine.ConnManager

	biJMux   sync.RWMutex                         // mux protecting BI-JSON connections
	biJClnts map[rpcclient.ClientConnector]string // index BiJSONConnection so we can sync them later
	biJIDs   map[string]*biJClient                // identifiers of bidirectional JSON conns, used to call RPC based on connIDs

	aSsMux    sync.RWMutex        // protects aSessions
	aSessions map[string]*Session // group sessions per sessionId

	aSIMux        sync.RWMutex                                     // protects aSessionsIdx
	aSessionsIdx  map[string]map[string]map[string]utils.StringMap // map[fieldName]map[fieldValue][cgrID]utils.StringMap[runID]sID
	aSessionsRIdx map[string][]*riFieldNameVal                     // reverse indexes for active sessions, used on remove

	pSsMux    sync.RWMutex        // protects pSessions
	pSessions map[string]*Session // group passive sessions based on cgrID

	pSIMux        sync.RWMutex                                     // protects pSessionsIdx
	pSessionsIdx  map[string]map[string]map[string]utils.StringMap // map[fieldName]map[fieldValue][cgrID]utils.StringMap[runID]sID
	pSessionsRIdx map[string][]*riFieldNameVal                     // reverse indexes for passive sessions, used on remove
}

// ListenAndServe starts the service and binds it to the listen loop
func (sS *SessionS) ListenAndServe(exitChan chan bool) (err error) {
	utils.Logger.Info(fmt.Sprintf("<%s> starting <%s> subsystem", utils.CoreS, utils.SessionS))
	if sS.cgrCfg.SessionSCfg().ChannelSyncInterval != 0 {
		go func() {
			for { // Schedule sync channels to run repeately
				select {
				case e := <-exitChan:
					exitChan <- e
					break
				case <-time.After(sS.cgrCfg.SessionSCfg().ChannelSyncInterval):
					sS.syncSessions()
				}
			}

		}()
	}
	e := <-exitChan // block here until shutdown request
	exitChan <- e   // put back for the others listening for shutdown request
	return
}

// Shutdown is called by engine to clear states
func (sS *SessionS) Shutdown() (err error) {
	if len(sS.cgrCfg.SessionSCfg().ReplicationConns) == 0 {
		var hasErr bool
		for _, s := range sS.getSessions("", false) { // Force sessions shutdown
			if err = sS.terminateSession(s, nil, nil, nil, false); err != nil {
				hasErr = true
			}
		}
		if hasErr {
			return utils.ErrPartiallyExecuted
		}
	}
	return
}

// OnBiJSONConnect is called by rpc2.Client on each new connection
func (sS *SessionS) OnBiJSONConnect(c *rpc2.Client) {
	sS.biJMux.Lock()
	nodeID := utils.UUIDSha1Prefix() // connection identifier, should be later updated as login procedure
	sS.biJClnts[c] = nodeID
	sS.biJIDs[nodeID] = &biJClient{
		conn:  c,
		proto: sS.cgrCfg.SessionSCfg().ClientProtocol}
	sS.biJMux.Unlock()
}

// OnBiJSONDisconnect is called by rpc2.Client on each client disconnection
func (sS *SessionS) OnBiJSONDisconnect(c *rpc2.Client) {
	sS.biJMux.Lock()
	if nodeID, has := sS.biJClnts[c]; has {
		delete(sS.biJClnts, c)
		delete(sS.biJIDs, nodeID)
	}
	sS.biJMux.Unlock()
}

// RegisterIntBiJConn is called on internal BiJ connection towards SessionS
func (sS *SessionS) RegisterIntBiJConn(c rpcclient.ClientConnector) {
	sS.biJMux.Lock()
	nodeID := sS.cgrCfg.GeneralCfg().NodeID
	sS.biJClnts[c] = nodeID
	sS.biJIDs[nodeID] = &biJClient{
		conn:  c,
		proto: sS.cgrCfg.SessionSCfg().ClientProtocol}
	sS.biJMux.Unlock()
}

// biJClnt returns a bidirectional JSON client based on connection ID
func (sS *SessionS) biJClnt(connID string) (clnt *biJClient) {
	if connID == "" {
		return
	}
	sS.biJMux.RLock()
	clnt = sS.biJIDs[connID]
	sS.biJMux.RUnlock()
	return
}

// biJClnt returns connection ID based on bidirectional connection received
func (sS *SessionS) biJClntID(c rpcclient.ClientConnector) (clntConnID string) {
	if c == nil {
		return
	}
	sS.biJMux.RLock()
	clntConnID = sS.biJClnts[c]
	sS.biJMux.RUnlock()
	return
}

// biJClnts is a thread-safe method to return the list of active clients for BiJson
func (sS *SessionS) biJClients() (clnts []*biJClient) {
	sS.biJMux.RLock()
	clnts = make([]*biJClient, len(sS.biJIDs))
	i := 0
	for _, clnt := range sS.biJIDs {
		clnts[i] = clnt
		i++
	}
	sS.biJMux.RUnlock()
	return
}

// riFieldNameVal is a reverse index entry
type riFieldNameVal struct {
	fieldName, fieldValue string
}

// sTerminator holds the info needed to force-terminate sessions based on timer
type sTerminator struct {
	timer        *time.Timer
	endChan      chan struct{}
	ttl          time.Duration
	ttlLastUsed  *time.Duration
	ttlUsage     *time.Duration
	ttlLastUsage *time.Duration
}

// setSTerminator installs a new terminator for a session
// setSTerminator is not thread safe, only the goroutine forked from within
func (sS *SessionS) setSTerminator(s *Session) {
	// TTL
	ttl, err := s.EventStart.GetDuration(utils.SessionTTL)
	if err != nil {
		if err != utils.ErrNotFound {
			utils.Logger.Warning(
				fmt.Sprintf("<%s>, cannot extract <%s> from event: <%s>, err: <%s>",
					utils.SessionS, utils.SessionTTL, s.EventStart.String(), err.Error()))
			return
		}
		ttl = sS.cgrCfg.SessionSCfg().SessionTTL
	}
	if ttl == 0 {
		return // nothing to set up
	}
	// random delay computation
	maxDelay, err := s.EventStart.GetDuration(utils.SessionTTLMaxDelay)
	if err != nil {
		if err != utils.ErrNotFound {
			utils.Logger.Warning(
				fmt.Sprintf("<%s>, cannot extract <%s> from event: <%s>, err: <%s>",
					utils.SessionS, utils.SessionTTLMaxDelay, s.EventStart.String(), err.Error()))
			return
		}
		err = nil
		if sS.cgrCfg.SessionSCfg().SessionTTLMaxDelay != nil {
			maxDelay = *sS.cgrCfg.SessionSCfg().SessionTTLMaxDelay
		}
	}
	if maxDelay != 0 {
		rand.Seed(time.Now().Unix())
		ttl += time.Duration(
			rand.Int63n(maxDelay.Nanoseconds()/time.Millisecond.Nanoseconds()) * time.Millisecond.Nanoseconds())
	}
	// LastUsed
	ttlLastUsed, err := s.EventStart.GetDurationPtrOrDefault(
		utils.SessionTTLLastUsed, sS.cgrCfg.SessionSCfg().SessionTTLLastUsed)
	if err != nil {
		utils.Logger.Warning(
			fmt.Sprintf("<%s>, cannot extract <%s> from event: <%s>, err: <%s>",
				utils.SessionS, utils.SessionTTLLastUsed, s.EventStart.String(), err.Error()))
		return
	}
	// TTLUsage
	ttlUsage, err := s.EventStart.GetDurationPtrOrDefault(
		utils.SessionTTLUsage, sS.cgrCfg.SessionSCfg().SessionTTLUsage)
	if err != nil {
		utils.Logger.Warning(
			fmt.Sprintf("<%s>, cannot extract <%s> from event: <%s>, err: <%s>",
				utils.SessionS, utils.SessionTTLUsage, s.EventStart.String(), err.Error()))
		return
	}
	// TTLLastUsage
	ttlLastUsage, err := s.EventStart.GetDurationPtrOrDefault(
		utils.SessionTTLLastUsage, sS.cgrCfg.SessionSCfg().SessionTTLLastUsage)
	if err != nil {
		utils.Logger.Warning(
			fmt.Sprintf("<%s>, cannot extract <%s> from event: <%s>, err: <%s>",
				utils.SessionS, utils.SessionTTLLastUsage, s.EventStart.String(), err.Error()))
		return
	}
	// previously defined, reset
	if s.sTerminator != nil {
		s.sTerminator.ttl = ttl
		if ttlLastUsed != nil {
			s.sTerminator.ttlLastUsed = ttlLastUsed
		}
		if ttlUsage != nil {
			s.sTerminator.ttlUsage = ttlUsage
		}
		if ttlLastUsage != nil {
			s.sTerminator.ttlLastUsage = ttlLastUsage
		}
		s.sTerminator.timer.Reset(s.sTerminator.ttl)
		return
	}
	// new set
	s.sTerminator = &sTerminator{
		timer:        time.NewTimer(ttl),
		endChan:      make(chan struct{}),
		ttl:          ttl,
		ttlLastUsed:  ttlLastUsed,
		ttlUsage:     ttlUsage,
		ttlLastUsage: ttlLastUsage,
	}

	// schedule automatic termination
	go func(endChan chan struct{}, timer *time.Timer) {
		select {
		case <-timer.C:
			s.Lock()
			lastUsage := s.sTerminator.ttl
			if s.sTerminator.ttlLastUsage != nil {
				lastUsage = *s.sTerminator.ttlLastUsage
			}
			sS.forceSTerminate(s, lastUsage,
				s.sTerminator.ttlUsage, s.sTerminator.ttlLastUsed)
			s.Unlock()
		case <-endChan:
			timer.Stop()
		}
	}(s.sTerminator.endChan, s.sTerminator.timer)
	runtime.Gosched() // force context switching
}

// forceSTerminate is called when a session times-out or it is forced from CGRateS side
// not thread safe
func (sS *SessionS) forceSTerminate(s *Session, extraUsage time.Duration, tUsage, lastUsed *time.Duration) (err error) {
	if extraUsage != 0 {
		for i := range s.SRuns {
			if _, err = sS.debitSession(s, i, extraUsage, lastUsed); err != nil {
				utils.Logger.Warning(
					fmt.Sprintf(
						"<%s> failed debitting cgrID %s, sRunIdx: %d, err: %s",
						utils.SessionS, s.cgrID(), i, err.Error()))
			}
		}
	}
	// we apply the correction before
	if err = sS.endSession(s, tUsage, lastUsed, nil, false); err != nil {
		utils.Logger.Warning(
			fmt.Sprintf(
				"<%s> failed force terminating session with ID <%s>, err: <%s>",
				utils.SessionS, s.cgrID(), err.Error()))
	}
	// post the CDRs
	if len(sS.cgrCfg.SessionSCfg().CDRsConns) != 0 {
		if cgrEvs, err := s.asCGREvents(); err != nil {
			utils.Logger.Warning(
				fmt.Sprintf(
					"<%s> failed convering session: %s in CGREvents with err: %s",
					utils.SessionS, utils.ToJSON(s), err.Error()))
		} else {
			var reply string
			for _, cgrEv := range cgrEvs {
				argsProc := &engine.ArgV1ProcessEvent{
					Flags: []string{fmt.Sprintf("%s:false", utils.MetaChargers),
						fmt.Sprintf("%s:false", utils.MetaAttributes)},
					CGREvent:      *cgrEv,
					ArgDispatcher: s.ArgDispatcher,
				}
				if unratedReqs.HasField( // order additional rating for unrated request types
					engine.MapEvent(cgrEv.Event).GetStringIgnoreErrors(utils.RequestType)) {
					argsProc.Flags = append(argsProc.Flags, utils.MetaRALs)
				}
				if err = sS.connMgr.Call(sS.cgrCfg.SessionSCfg().CDRsConns, nil,
					utils.CDRsV1ProcessEvent, argsProc, &reply); err != nil {
					utils.Logger.Warning(
						fmt.Sprintf(
							"<%s> could not post CDR for event %s, err: %s",
							utils.SessionS, utils.ToJSON(cgrEv), err.Error()))
				}
			}
		}
	}
	// release the resources for the session
	if len(sS.cgrCfg.SessionSCfg().ResSConns) != 0 && s.ResourceID != "" {
		var reply string
		argsRU := utils.ArgRSv1ResourceUsage{
			CGREvent: &utils.CGREvent{
				Tenant: s.Tenant,
				ID:     utils.GenUUID(),
				Event:  s.EventStart,
			},
			UsageID:       s.ResourceID,
			Units:         1,
			ArgDispatcher: s.ArgDispatcher,
		}
		if err := sS.connMgr.Call(sS.cgrCfg.SessionSCfg().ResSConns, nil,
			utils.ResourceSv1ReleaseResources,
			argsRU, &reply); err != nil {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> error: %s could not release resource with resourceID: %s",
					utils.SessionS, err.Error(), s.ResourceID))
		}
	}
	sS.replicateSessions(s.CGRID, false, sS.cgrCfg.SessionSCfg().ReplicationConns)
	if clntConn := sS.biJClnt(s.ClientConnID); clntConn != nil {
		go func() {
			var rply string
			if err := clntConn.conn.Call(
				utils.SessionSv1DisconnectSession,
				utils.AttrDisconnectSession{
					EventStart: s.EventStart,
					Reason:     ErrForcedDisconnect.Error()},
				&rply); err != nil {
				if err != utils.ErrNotImplemented {
					utils.Logger.Warning(
						fmt.Sprintf("<%s> err: %s remotely disconnect session with id: %s",
							utils.SessionS, err.Error(), s.CGRID))
				}
			}
		}()
	}
	return
}

// debitSession performs debit for a session run
func (sS *SessionS) debitSession(s *Session, sRunIdx int, dur time.Duration,
	lastUsed *time.Duration) (maxDur time.Duration, err error) {
	if sRunIdx >= len(s.SRuns) {
		err = errors.New("sRunIdx out of range")
		return
	}
	sr := s.SRuns[sRunIdx]
	rDur := sr.debitReserve(dur, lastUsed) // debit out of reserve, rDur is still to be debited
	if rDur == time.Duration(0) {
		return dur, nil // complete debit out of reserve
	}
	dbtRsrv := dur - rDur // the amount debited from reserve
	if sr.CD.LoopIndex > 0 {
		sr.CD.TimeStart = sr.CD.TimeEnd
	}
	sr.CD.TimeEnd = sr.CD.TimeStart.Add(rDur)
	sr.CD.DurationIndex += rDur
	cd := sr.CD.Clone()
	argDsp := s.ArgDispatcher
	cc := new(engine.CallCost)
	if err := sS.connMgr.Call(sS.cgrCfg.SessionSCfg().RALsConns, nil,
		utils.ResponderMaxDebit,
		&engine.CallDescriptorWithArgDispatcher{
			CallDescriptor: cd,
			ArgDispatcher:  argDsp}, cc); err != nil {
		sr.ExtraDuration += dbtRsrv
		return 0, err
	}
	sr.CD.TimeEnd = cc.GetEndTime() // set debited timeEnd
	ccDuration := cc.GetDuration()
	if ccDuration > rDur {
		sr.ExtraDuration = ccDuration - rDur
	}
	if ccDuration >= rDur {
		sr.LastUsage = dur
	} else {
		sr.LastUsage = ccDuration + dbtRsrv
	}
	sr.CD.DurationIndex -= rDur
	sr.CD.DurationIndex += ccDuration
	sr.CD.MaxCostSoFar += cc.Cost
	sr.CD.LoopIndex++
	sr.TotalUsage += sr.LastUsage
	ec := engine.NewEventCostFromCallCost(cc, s.CGRID,
		sr.Event.GetStringIgnoreErrors(utils.RunID))
	if sr.EventCost == nil {
		if ccDuration != time.Duration(0) {
			sr.EventCost = ec
		}
	} else {
		ec.SyncKeys(sr.EventCost)
		sr.EventCost.Merge(ec)
	}
	maxDur = sr.LastUsage
	return
}

// debitLoopSession will periodically debit sessions, ie: automatic prepaid
// threadSafe since it will run into it's own goroutine
func (sS *SessionS) debitLoopSession(s *Session, sRunIdx int,
	dbtIvl time.Duration) (maxDur time.Duration, err error) {
	// NextAutoDebit works in tandem with session replication
	now := time.Now()
	if s.SRuns[sRunIdx].NextAutoDebit != nil &&
		now.Before(*s.SRuns[sRunIdx].NextAutoDebit) {
		time.Sleep(now.Sub(*s.SRuns[sRunIdx].NextAutoDebit))
	}
	for {
		s.Lock()
		if s.debitStop == nil {
			// session already closed (most probably from sessionEnd), fixes concurrency
			s.Unlock()
			return
		}
		var maxDebit time.Duration
		if maxDebit, err = sS.debitSession(s, sRunIdx, dbtIvl, nil); err != nil {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> could not complete debit operation on session: <%s>, error: <%s>",
					utils.SessionS, s.cgrID(), err.Error()))
			dscReason := utils.ErrServerError.Error()
			if err.Error() == utils.ErrUnauthorizedDestination.Error() {
				dscReason = err.Error()
			}
			// try to disconect the session n times before we force terminate it on our side
			for i := 0; i < sS.cgrCfg.SessionSCfg().TerminateAttempts; i++ {
				if err = sS.disconnectSession(s, dscReason); err == nil {
					s.Unlock()
					return
				}
				utils.Logger.Warning(
					fmt.Sprintf("<%s> could not disconnect session: %s, error: %s",
						utils.SessionS, s.cgrID(), err.Error()))
			}
			if err = sS.forceSTerminate(s, 0, nil, nil); err != nil {
				utils.Logger.Warning(fmt.Sprintf("<%s> failed force-terminating session: <%s>, err: <%s>", utils.SessionS, s.cgrID(), err))
			}
			s.Unlock()
			return
		}
		debitStop := s.debitStop // avoid concurrency with endSession
		s.SRuns[sRunIdx].NextAutoDebit = utils.TimePointer(time.Now().Add(dbtIvl))
		s.Unlock()
		sS.replicateSessions(s.CGRID, false, sS.cgrCfg.SessionSCfg().ReplicationConns)
		if maxDebit < dbtIvl { // disconnect faster
			select {
			case <-debitStop: // call was disconnected already
				return
			case <-time.After(maxDebit):
				s.Lock()
				defer s.Unlock()
				// try to disconect the session n times before we force terminate it on our side
				for i := 0; i < sS.cgrCfg.SessionSCfg().TerminateAttempts; i++ {
					if err = sS.disconnectSession(s, utils.ErrInsufficientCredit.Error()); err == nil {
						return
					}
				}
				utils.Logger.Warning(
					fmt.Sprintf("<%s> could not disconnect session: <%s>, error: <%s>",
						utils.SessionS, s.cgrID(), err.Error()))
				if err = sS.forceSTerminate(s, 0, nil, nil); err != nil {
					utils.Logger.Warning(fmt.Sprintf("<%s> failed force-terminating session: <%s>, err: <%s>",
						utils.SessionS, s.cgrID(), err))
				}
			}
			return
		}
		select {
		case <-debitStop:
			return
		case <-time.After(dbtIvl):
			continue
		}
	}
}

// refundSession will refund the extra usage debitted by the end of session
// not thread-safe so the locks need to be done in a layer above
// rUsage represents the amount of usage to be refunded
func (sS *SessionS) refundSession(s *Session, sRunIdx int, rUsage time.Duration) (err error) {
	if sRunIdx >= len(s.SRuns) {
		return errors.New("sRunIdx out of range")
	}
	sr := s.SRuns[sRunIdx]
	if sr.EventCost == nil {
		return errors.New("no event cost")
	}
	srplsEC, err := sr.EventCost.Trim(sr.EventCost.GetUsage() - rUsage)
	if err != nil {
		return err
	} else if srplsEC == nil {
		return
	}
	sCC := srplsEC.AsCallCost(sr.CD.ToR)
	var incrmts engine.Increments
	for _, tmspn := range sCC.Timespans {
		for _, incr := range tmspn.Increments {
			if incr.BalanceInfo == nil ||
				(incr.BalanceInfo.Unit == nil &&
					incr.BalanceInfo.Monetary == nil) {
				continue // not enough information for refunds, most probably free units uncounted
			}
			for i := 0; i < tmspn.CompressFactor; i++ {
				incrmts = append(incrmts, incr)
			}
		}
	}
	cd := &engine.CallDescriptor{
		CgrID:       s.CGRID,
		RunID:       sr.Event.GetStringIgnoreErrors(utils.RunID),
		Category:    sr.CD.Category,
		Tenant:      sr.CD.Tenant,
		Subject:     sr.CD.Subject,
		Account:     sr.CD.Account,
		Destination: sr.CD.Destination,
		ToR:         utils.FirstNonEmpty(sr.CD.ToR, utils.VOICE),
		Increments:  incrmts,
	}
	var acnt engine.Account
	if err = sS.connMgr.Call(sS.cgrCfg.SessionSCfg().RALsConns, nil, utils.ResponderRefundIncrements,
		&engine.CallDescriptorWithArgDispatcher{CallDescriptor: cd,
			ArgDispatcher: s.ArgDispatcher}, &acnt); err != nil {
		return
	}
	if acnt.ID != "" { // Account info updated, update also cached AccountSummary
		sr.EventCost.AccountSummary = acnt.AsAccountSummary()
	}
	return
}

// storeSCost will post the session cost to CDRs
// not thread safe, need to be handled in a layer above
func (sS *SessionS) storeSCost(s *Session, sRunIdx int) (err error) {
	sr := s.SRuns[sRunIdx]
	smCost := &engine.SMCost{
		CGRID:       s.CGRID,
		CostSource:  utils.MetaSessionS,
		RunID:       sr.Event.GetStringIgnoreErrors(utils.RunID),
		OriginHost:  s.EventStart.GetStringIgnoreErrors(utils.OriginHost),
		OriginID:    s.EventStart.GetStringIgnoreErrors(utils.OriginID),
		Usage:       sr.TotalUsage,
		CostDetails: sr.EventCost,
	}
	argSmCost := &engine.AttrCDRSStoreSMCost{
		Cost:           smCost,
		CheckDuplicate: true,
		ArgDispatcher:  s.ArgDispatcher,
		TenantArg: &utils.TenantArg{
			Tenant: s.Tenant,
		},
	}
	var reply string
	// use the v1 because it doesn't do rounding refund
	if err := sS.connMgr.Call(sS.cgrCfg.SessionSCfg().CDRsConns, nil, utils.CDRsV1StoreSessionCost,
		argSmCost, &reply); err != nil && err == utils.ErrExists {
		utils.Logger.Warning(
			fmt.Sprintf("<%s> refunding session: <%s> error: <%s>",
				utils.SessionS, s.CGRID, err.Error()))
		if err = sS.refundSession(s, sRunIdx, sr.CD.GetDuration()); err != nil { // refund entire duration
			utils.Logger.Warning(
				fmt.Sprintf(
					"<%s> failed refunding session: <%s>, srIdx: <%d>, error: <%s>",
					utils.SessionS, s.CGRID, sRunIdx, err.Error()))
		}
		err = nil
	}
	return err
}

// roundCost will round the EventCost and will refund the extra debited increments
// should be called only at the endSession
// not thread safe, need to be handled in a layer above
func (sS *SessionS) roundCost(s *Session, sRunIdx int) (err error) {
	sr := s.SRuns[sRunIdx]
	runID := sr.Event.GetStringIgnoreErrors(utils.RunID)
	cc := sr.EventCost.AsCallCost(utils.EmptyString)
	cc.Round()
	if roundIncrements := cc.GetRoundIncrements(); len(roundIncrements) != 0 {
		cd := cc.CreateCallDescriptor()
		cd.CgrID = s.CGRID
		cd.RunID = runID
		cd.Increments = roundIncrements
		var response float64
		if err = sS.connMgr.Call(sS.cgrCfg.SessionSCfg().RALsConns, nil,
			utils.ResponderRefundRounding,
			&engine.CallDescriptorWithArgDispatcher{CallDescriptor: cd},
			&response); err != nil {
			return
		}
	}
	sr.EventCost = engine.NewEventCostFromCallCost(cc, s.CGRID, runID)
	return
}

// disconnectSession will send disconnect from SessionS to clients
// not thread safe, it considers that the session is already stopped by this time
func (sS *SessionS) disconnectSession(s *Session, rsn string) (err error) {
	clnt := sS.biJClnt(s.ClientConnID)
	if clnt == nil {
		return fmt.Errorf("calling %s requires bidirectional JSON connection, connID: <%s>",
			utils.SessionSv1DisconnectSession, s.ClientConnID)
	}
	s.EventStart[utils.Usage] = s.totalUsage() // Set the usage to total one debitted
	servMethod := utils.SessionSv1DisconnectSession
	if clnt.proto == 0 { // compatibility with OpenSIPS 2.3
		servMethod = "SMGClientV1.DisconnectSession"
	}
	var rply string
	if err = clnt.conn.Call(servMethod,
		utils.AttrDisconnectSession{
			EventStart: s.EventStart,
			Reason:     rsn}, &rply); err != nil {
		if err != utils.ErrNotImplemented {
			return err
		}
		err = nil
	}
	return
}

// replicateSessions will replicate sessions with or without cgrID specified
func (sS *SessionS) replicateSessions(cgrID string, psv bool, connIDs []string) (err error) {
	if len(connIDs) == 0 {
		return
	}
	ss := sS.getSessions(cgrID, psv)
	if len(ss) == 0 {
		// session scheduled to be removed from remote (initiate also the EventStart to avoid the panic)
		ss = []*Session{{
			CGRID:      cgrID,
			EventStart: make(engine.MapEvent),
		}}
	}
	for _, s := range ss {
		sCln := s.Clone()
		var rply string
		if err := sS.connMgr.Call(connIDs, nil,
			utils.SessionSv1SetPassiveSession,
			sCln, &rply); err != nil {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> cannot replicate session with id <%s>, err: %s",
					utils.SessionS, sCln.CGRID, err.Error()))
		}
	}
	return
}

// registerSession will register an active or passive Session
// called on init or relocate
// not thread safe for the Session
func (sS *SessionS) registerSession(s *Session, passive bool) {
	sMux := &sS.aSsMux
	sMp := sS.aSessions
	if passive {
		sMux = &sS.pSsMux
		sMp = sS.pSessions
	}
	sMux.Lock()
	sMp[s.CGRID] = s
	sMux.Unlock()
	sS.indexSession(s, passive)
}

// isIndexed returns if the session is indexed
func (sS *SessionS) isIndexed(s *Session, passive bool) (has bool) {
	sMux := &sS.aSsMux
	sMp := sS.aSessions
	if passive {
		sMux = &sS.pSsMux
		sMp = sS.pSessions
	}
	sMux.Lock()
	_, has = sMp[s.CGRID]
	sMux.Unlock()
	return
}

// uregisterSession will unregister an active or passive session based on it's CGRID
// called on session terminate or relocate
func (sS *SessionS) unregisterSession(cgrID string, passive bool) bool {
	sMux := &sS.aSsMux
	sMp := sS.aSessions
	if passive {
		sMux = &sS.pSsMux
		sMp = sS.pSessions
	}
	sMux.Lock()
	if _, has := sMp[cgrID]; !has {
		sMux.Unlock()
		return false
	}
	delete(sMp, cgrID)
	sMux.Unlock()
	sS.unindexSession(cgrID, passive)
	return true
}

// indexSession will index an active or passive Session based on configuration
func (sS *SessionS) indexSession(s *Session, pSessions bool) {
	idxMux := &sS.aSIMux // pointer to original mux since will have no effect if we copy it
	ssIndx := sS.aSessionsIdx
	ssRIdx := sS.aSessionsRIdx
	if pSessions {
		idxMux = &sS.pSIMux
		ssIndx = sS.pSessionsIdx
		ssRIdx = sS.pSessionsRIdx
	}
	idxMux.Lock()
	defer idxMux.Unlock()
	for fieldName := range sS.cgrCfg.SessionSCfg().SessionIndexes {
		for _, sr := range s.SRuns {
			fieldVal, err := sr.Event.GetString(fieldName) // the only error from GetString is ErrNotFound
			if err != nil {
				fieldVal = utils.NOT_AVAILABLE
			}
			if fieldVal == "" {
				fieldVal = utils.MetaEmpty
			}
			if _, hasFieldName := ssIndx[fieldName]; !hasFieldName { // Init it here
				ssIndx[fieldName] = make(map[string]map[string]utils.StringMap)
			}
			if _, hasFieldVal := ssIndx[fieldName][fieldVal]; !hasFieldVal {
				ssIndx[fieldName][fieldVal] = make(map[string]utils.StringMap)
			}
			if _, hasCGRID := ssIndx[fieldName][fieldVal][s.CGRID]; !hasCGRID {
				ssIndx[fieldName][fieldVal][s.CGRID] = make(utils.StringMap)
			}
			ssIndx[fieldName][fieldVal][s.CGRID][sr.CD.RunID] = true

			// reverse index
			if _, hasIt := ssRIdx[s.CGRID]; !hasIt {
				ssRIdx[s.CGRID] = make([]*riFieldNameVal, 0)
			}
			ssRIdx[s.CGRID] = append(ssRIdx[s.CGRID], &riFieldNameVal{fieldName: fieldName, fieldValue: fieldVal})
		}
	}
	return
}

// unindexASession removes an active or passive session from indexes
// called on terminate or relocate
func (sS *SessionS) unindexSession(cgrID string, pSessions bool) bool {
	idxMux := &sS.aSIMux
	ssIndx := sS.aSessionsIdx
	ssRIdx := sS.aSessionsRIdx
	if pSessions {
		idxMux = &sS.pSIMux
		ssIndx = sS.pSessionsIdx
		ssRIdx = sS.pSessionsRIdx
	}
	idxMux.Lock()
	defer idxMux.Unlock()
	if _, hasIt := ssRIdx[cgrID]; !hasIt {
		return false
	}
	for _, riFNV := range ssRIdx[cgrID] {
		delete(ssIndx[riFNV.fieldName][riFNV.fieldValue], cgrID)
		if len(ssIndx[riFNV.fieldName][riFNV.fieldValue]) == 0 {
			delete(ssIndx[riFNV.fieldName], riFNV.fieldValue)
		}
		if len(ssIndx[riFNV.fieldName]) == 0 {
			delete(ssIndx, riFNV.fieldName)
		}
	}
	delete(ssRIdx, cgrID)
	return true
}

func (sS *SessionS) getIndexedFilters(tenant string, fltrs []string) (
	indexedFltr map[string][]string, unindexedFltr []*engine.FilterRule) {
	indexedFltr = make(map[string][]string)
	for _, fltrID := range fltrs {
		f, err := engine.GetFilter(sS.dm, tenant, fltrID,
			true, true, utils.NonTransactional)
		if err != nil {
			if err == utils.ErrNotFound {
				err = utils.ErrPrefixNotFound(fltrID)
			}
			continue
		}
		if f.ActivationInterval != nil &&
			!f.ActivationInterval.IsActiveAtTime(time.Now()) { // not active
			continue
		}
		for _, fltr := range f.Rules {
			fldName := strings.TrimPrefix(fltr.Element, utils.DynamicDataPrefix+utils.MetaReq+utils.NestingSep) // remove ~req. prefix
			if fltr.Type != utils.MetaString ||
				!sS.cgrCfg.SessionSCfg().SessionIndexes.HasKey(fldName) {
				unindexedFltr = append(unindexedFltr, fltr)
				continue
			}
			indexedFltr[fldName] = fltr.Values
		}
	}
	return
}

// getSessionIDsMatchingIndexes returns map[matchedFieldName]possibleMatchedFieldVal so we optimize further to avoid checking them
func (sS *SessionS) getSessionIDsMatchingIndexes(fltrs map[string][]string,
	pSessions bool) ([]string, map[string]utils.StringMap) {
	idxMux := &sS.aSIMux
	ssIndx := sS.aSessionsIdx
	if pSessions {
		idxMux = &sS.pSIMux
		ssIndx = sS.pSessionsIdx
	}
	idxMux.RLock()
	defer idxMux.RUnlock()
	matchingSessions := make(map[string]utils.StringMap)
	checkNr := 0
	getMatchingIndexes := func(fltrName string, values []string) (matchingSessionsbyValue map[string]utils.StringMap) {
		matchingSessionsbyValue = make(map[string]utils.StringMap)
		// fltrName = strings.TrimPrefix(fltrName, utils.DynamicDataPrefix+utils.MetaReq+utils.NestingSep) //
		if _, hasFldName := ssIndx[fltrName]; !hasFldName {
			return
		}
		for _, fltrVal := range values {
			if _, hasFldVal := ssIndx[fltrName][fltrVal]; !hasFldVal {
				continue
			}
			for cgrID, runIDs := range ssIndx[fltrName][fltrVal] {
				if _, hasCGRID := matchingSessionsbyValue[cgrID]; !hasCGRID {
					matchingSessionsbyValue[cgrID] = utils.NewStringMap()
				}
				for runID := range runIDs {
					matchingSessionsbyValue[cgrID][runID] = true
				}
			}
		}
		return matchingSessionsbyValue
	}
	for fltrName, fltrVals := range fltrs {
		matchedIndx := getMatchingIndexes(fltrName, fltrVals)
		checkNr++
		if checkNr == 1 { // First run will init the MatchingSessions
			matchingSessions = matchedIndx
			continue
		}
		// Higher run, takes out non matching indexes
		for cgrID, runIDs := range matchingSessions {
			if matchedRunIDs, hasCGRID := matchedIndx[cgrID]; !hasCGRID {
				delete(matchingSessions, cgrID)
				continue
			} else {
				for runID := range runIDs {
					if !matchedRunIDs.HasKey(runID) {
						delete(matchingSessions[cgrID], runID)
					}
				}
			}
		}
		if len(matchingSessions) == 0 {
			return make([]string, 0), make(map[string]utils.StringMap)
		}
	}
	cgrIDs := []string{}
	for cgrID := range matchingSessions {
		cgrIDs = append(cgrIDs, cgrID)

	}
	return cgrIDs, matchingSessions
}

// filterSessions will return a list of sessions in external format based on filters passed
// is thread safe for the Sessions
func (sS *SessionS) filterSessions(sf *utils.SessionFilter, psv bool) (aSs []*ExternalSession) {
	if len(sf.Filters) == 0 {
		ss := sS.getSessions(utils.EmptyString, psv)
		for _, s := range ss {
			aSs = append(aSs,
				s.AsExternalSessions(sS.cgrCfg.GeneralCfg().DefaultTimezone,
					sS.cgrCfg.GeneralCfg().NodeID)...) // Expensive for large number of sessions
			if sf.Limit != nil && *sf.Limit > 0 && *sf.Limit < len(aSs) {
				return aSs[:*sf.Limit]
			}
		}
		return
	}
	tenant := utils.FirstNonEmpty(sf.Tenant, sS.cgrCfg.GeneralCfg().DefaultTenant)
	indx, unindx := sS.getIndexedFilters(tenant, sf.Filters)
	cgrIDs, matchingSRuns := sS.getSessionIDsMatchingIndexes(indx, psv)
	if len(indx) != 0 && len(cgrIDs) == 0 { // no sessions matched the indexed filters
		return
	}
	ss := sS.getSessionsFromCGRIDs(psv, cgrIDs...)
	pass := func(filterRules []*engine.FilterRule,
		me engine.MapEvent) (pass bool) {
		pass = true
		if len(filterRules) == 0 {
			return
		}
		var err error
		ev := utils.MapStorage{utils.MetaReq: me.Data()}
		for _, fltr := range filterRules {
			// we don't know how many values we have so we need to build the fieldValues DataProvider
			fieldValuesDP := make([]utils.DataProvider, len(fltr.Values))
			for i := range fltr.Values {
				fieldValuesDP[i] = ev
			}
			if pass, err = fltr.Pass(ev, fieldValuesDP); err != nil || !pass {
				pass = false
				return
			}
		}
		return
	}
	for _, s := range ss {
		s.RLock()
		runIDs := matchingSRuns[s.CGRID]
		for _, sr := range s.SRuns {
			if len(cgrIDs) != 0 && !runIDs.HasKey(sr.CD.RunID) {
				continue
			}
			if pass(unindx, sr.Event) {
				aSs = append(aSs,
					s.AsExternalSession(sr, sS.cgrCfg.GeneralCfg().DefaultTimezone,
						sS.cgrCfg.GeneralCfg().NodeID)) // Expensive for large number of sessions
				if sf.Limit != nil && *sf.Limit > 0 && *sf.Limit < len(aSs) {
					s.RUnlock()
					return aSs[:*sf.Limit]
				}
			}
		}
		s.RUnlock()
	}
	return
}

// filterSessionsCount re
func (sS *SessionS) filterSessionsCount(sf *utils.SessionFilter, psv bool) (count int) {
	count = 0
	if len(sf.Filters) == 0 {
		ss := sS.getSessions(utils.EmptyString, psv)
		for _, s := range ss {
			count += len(s.SRuns)
		}
		return
	}
	tenant := utils.FirstNonEmpty(sf.Tenant, sS.cgrCfg.GeneralCfg().DefaultTenant)
	indx, unindx := sS.getIndexedFilters(tenant, sf.Filters)
	cgrIDs, matchingSRuns := sS.getSessionIDsMatchingIndexes(indx, psv)
	if len(indx) != 0 && len(cgrIDs) == 0 { // no sessions matched the indexed filters
		return
	}
	ss := sS.getSessionsFromCGRIDs(psv, cgrIDs...)
	pass := func(filterRules []*engine.FilterRule,
		me engine.MapEvent) (pass bool) {
		pass = true
		if len(filterRules) == 0 {
			return
		}
		var err error
		ev := utils.MapStorage{utils.MetaReq: me.Data()}
		for _, fltr := range filterRules {
			// we don't know how many values we have so we need to build the fieldValues DataProvider
			fieldValuesDP := make([]utils.DataProvider, len(fltr.Values))
			for i := range fltr.Values {
				fieldValuesDP[i] = ev
			}
			if pass, err = fltr.Pass(ev, fieldValuesDP); err != nil || !pass {
				return
			}
		}
		return
	}
	for _, s := range ss {
		s.RLock()
		runIDs := matchingSRuns[s.CGRID]
		for _, sr := range s.SRuns {
			if len(cgrIDs) != 0 && !runIDs.HasKey(sr.CD.RunID) {
				continue
			}
			if pass(unindx, sr.Event) {
				count++
			}
		}
		s.RUnlock()
	}
	return
}

// forkSession will populate SRuns within a Session based on ChargerS output
// forSession can only be called once per Session
// not thread-safe since it should be called in init where there is no concurrency
func (sS *SessionS) forkSession(s *Session, forceDuration bool) (err error) {
	if len(sS.cgrCfg.SessionSCfg().ChargerSConns) == 0 {
		return errors.New("ChargerS is disabled")
	}
	if len(s.SRuns) != 0 {
		return errors.New("already forked")
	}
	cgrEv := &utils.CGREventWithArgDispatcher{
		CGREvent: &utils.CGREvent{
			Tenant: s.Tenant,
			ID:     utils.UUIDSha1Prefix(),
			Event:  s.EventStart,
		},
		ArgDispatcher: s.ArgDispatcher,
	}
	var chrgrs []*engine.ChrgSProcessEventReply
	if err = sS.connMgr.Call(sS.cgrCfg.SessionSCfg().ChargerSConns, nil,
		utils.ChargerSv1ProcessEvent, cgrEv, &chrgrs); err != nil {
		return utils.NewErrChargerS(err)
	}
	s.SRuns = make([]*SRun, len(chrgrs))
	for i, chrgr := range chrgrs {
		me := engine.MapEvent(chrgr.CGREvent.Event)
		startTime := me.GetTimeIgnoreErrors(utils.AnswerTime,
			sS.cgrCfg.GeneralCfg().DefaultTimezone)
		if startTime.IsZero() { // AnswerTime not parsable, try SetupTime
			startTime = s.EventStart.GetTimeIgnoreErrors(utils.SetupTime,
				sS.cgrCfg.GeneralCfg().DefaultTimezone)
		}
		category := me.GetStringIgnoreErrors(utils.Category)
		if len(category) == 0 {
			category = sS.cgrCfg.GeneralCfg().DefaultCategory
		}
		subject := me.GetStringIgnoreErrors(utils.Subject)
		if len(subject) == 0 {
			subject = me.GetStringIgnoreErrors(utils.Account)
		}
		s.SRuns[i] = &SRun{
			Event: me,
			CD: &engine.CallDescriptor{
				CgrID:         s.CGRID,
				RunID:         me.GetStringIgnoreErrors(utils.RunID),
				ToR:           me.GetStringIgnoreErrors(utils.ToR),
				Tenant:        s.Tenant,
				Category:      category,
				Subject:       subject,
				Account:       me.GetStringIgnoreErrors(utils.Account),
				Destination:   me.GetStringIgnoreErrors(utils.Destination),
				TimeStart:     startTime,
				TimeEnd:       startTime.Add(s.EventStart.GetDurationIgnoreErrors(utils.Usage)),
				ExtraFields:   me.AsMapString(utils.MainCDRFields),
				ForceDuration: forceDuration,
			},
		}
	}
	return
}

// getSessions is used to return in a thread-safe manner active or passive sessions
func (sS *SessionS) getSessions(cgrID string, pSessions bool) (ss []*Session) {
	ssMux := &sS.aSsMux  // get the pointer so we don't copy, otherwise locks will not work
	ssMp := sS.aSessions // reference it so we don't overwrite the new map without protection
	if pSessions {
		ssMux = &sS.pSsMux
		ssMp = sS.pSessions
	}
	ssMux.RLock()
	defer ssMux.RUnlock()
	if len(cgrID) == 0 {
		ss = make([]*Session, len(ssMp))
		var i int
		for _, s := range ssMp {
			ss[i] = s
			i++
		}
		return
	}
	if s, hasCGRID := ssMp[cgrID]; hasCGRID {
		ss = []*Session{s}
	}
	return
}

// getSessions is used to return in a thread-safe manner active or passive sessions
func (sS *SessionS) getSessionsFromCGRIDs(pSessions bool, cgrIDs ...string) (ss []*Session) {
	ssMux := &sS.aSsMux  // get the pointer so we don't copy, otherwise locks will not work
	ssMp := sS.aSessions // reference it so we don't overwrite the new map without protection
	if pSessions {
		ssMux = &sS.pSsMux
		ssMp = sS.pSessions
	}
	ssMux.RLock()
	defer ssMux.RUnlock()
	if len(cgrIDs) == 0 {
		ss = make([]*Session, len(ssMp))
		var i int
		for _, s := range ssMp {
			ss[i] = s
			i++
		}
		return
	}
	ss = make([]*Session, len(cgrIDs))
	for i, cgrID := range cgrIDs {
		if s, hasCGRID := ssMp[cgrID]; hasCGRID {
			ss[i] = s
		}
	}
	return
}

// transitSState will transit the sessions from one state (active/passive) to another (passive/active)
func (sS *SessionS) transitSState(cgrID string, psv bool) (s *Session) {
	ss := sS.getSessions(cgrID, !psv)
	if len(ss) == 0 {
		return
	}
	s = ss[0]
	s.Lock()
	sS.unregisterSession(cgrID, !psv)
	sS.registerSession(s, psv)
	if !psv {
		sS.initSessionDebitLoops(s)
	} else { // transit from active with possible STerminator and DebitLoops
		s.stopSTerminator()
		s.stopDebitLoops()
	}
	s.Unlock()
	return
}

// getActivateSession returns the session from active list or moves from passive
func (sS *SessionS) getActivateSession(cgrID string) (s *Session) {
	ss := sS.getSessions(cgrID, false)
	if len(ss) != 0 {
		return ss[0]
	}
	return sS.transitSState(cgrID, false)
}

// relocateSession will change the CGRID of a session (ie: prefix based session group)
func (sS *SessionS) relocateSession(initOriginID, originID, originHost string) (s *Session) {
	if initOriginID == "" {
		return
	}
	initCGRID := utils.Sha1(initOriginID, originHost)
	newCGRID := utils.Sha1(originID, originHost)
	s = sS.getActivateSession(initCGRID)
	if s == nil {
		return
	}
	sS.unregisterSession(s.CGRID, false)
	s.Lock()
	s.CGRID = newCGRID
	// Overwrite initial CGRID with new one
	s.EventStart[utils.CGRID] = newCGRID    // Overwrite CGRID for final CDR
	s.EventStart[utils.OriginID] = originID // Overwrite OriginID for session indexing
	for _, sRun := range s.SRuns {
		sRun.Event[utils.CGRID] = newCGRID // needed for CDR generation
		sRun.Event[utils.OriginID] = originID
	}
	s.Unlock()
	sS.registerSession(s, false)
	sS.replicateSessions(initCGRID, false, sS.cgrCfg.SessionSCfg().ReplicationConns)
	return
}

// getRelocateSession will relocate a session if it cannot find cgrID and initialOriginID is present
func (sS *SessionS) getRelocateSession(cgrID string, initOriginID,
	originID, originHost string) (s *Session) {
	if s = sS.getActivateSession(cgrID); s != nil ||
		initOriginID == "" {
		return
	}
	return sS.relocateSession(initOriginID, originID, originHost)
}

// syncSessions synchronizes the active sessions with the one in the clients
// it will force-disconnect the one found in SessionS but not in clients
func (sS *SessionS) syncSessions() {
	queriedCGRIDs := engine.NewSafEvent(nil) // need this to be
	var err error
	for _, clnt := range sS.biJClients() {
		errChan := make(chan error)
		go func() {
			var queriedSessionIDs []*SessionID
			if err := clnt.conn.Call(utils.SessionSv1GetActiveSessionIDs,
				utils.EmptyString, &queriedSessionIDs); err != nil {
				errChan <- err
			}
			for _, sessionID := range queriedSessionIDs {
				queriedCGRIDs.Set(sessionID.CGRID(), struct{}{})
			}
			errChan <- nil
		}()
		select {
		case err = <-errChan:
			if err != nil {
				utils.Logger.Warning(
					fmt.Sprintf("<%s> error <%s> quering session ids", utils.SessionS, err.Error()))
			}
		case <-time.After(sS.cgrCfg.GeneralCfg().ReplyTimeout):
			utils.Logger.Warning(
				fmt.Sprintf("<%s> timeout quering session ids ", utils.SessionS))
		}

	}
	var toBeRemoved []string
	sS.aSsMux.RLock()
	for cgrid := range sS.aSessions {
		if !queriedCGRIDs.HasField(cgrid) {
			toBeRemoved = append(toBeRemoved, cgrid)
		}
	}
	sS.aSsMux.RUnlock()
	for _, cgrID := range toBeRemoved {
		ss := sS.getSessions(cgrID, false)
		if len(ss) == 0 {
			continue
		}
		ss[0].Lock() // protect forceSTerminate as it is not thread safe for the session
		if err := sS.forceSTerminate(ss[0], 0, nil, nil); err != nil {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> failed force-terminating session: <%s>, err: <%s>",
					utils.SessionS, cgrID, err.Error()))
		}
		ss[0].Unlock()
	}
}

// initSessionDebitLoops will init the debit loops for a session
// not thread-safe, it should be protected in another layer
func (sS *SessionS) initSessionDebitLoops(s *Session) {
	if s.debitStop != nil { // already initialized
		return
	}
	for i, sr := range s.SRuns {
		if s.DebitInterval != 0 &&
			sr.Event.GetStringIgnoreErrors(utils.RequestType) == utils.META_PREPAID {
			if s.debitStop == nil { // init the debitStop only for the first sRun with DebitInterval and RequestType META_PREPAID
				s.debitStop = make(chan struct{})
			}
			go sS.debitLoopSession(s, i, s.DebitInterval)
			runtime.Gosched() // allow the goroutine to be executed
		}
	}
}

// authEvent calculates maximum usage allowed for the given event
func (sS *SessionS) authEvent(tnt string, evStart engine.MapEvent, forceDuration bool) (maxUsage time.Duration, err error) {
	cgrID := GetSetCGRID(evStart)
	var eventUsage time.Duration
	if eventUsage, err = evStart.GetDuration(utils.Usage); err != nil {
		if err != utils.ErrNotFound {
			return
		}
		err = nil
		eventUsage = sS.cgrCfg.SessionSCfg().GetDefaultUsage(evStart.GetStringIgnoreErrors(utils.ToR))
		evStart[utils.Usage] = eventUsage // will be used in CD
	}
	s := &Session{
		CGRID:      cgrID,
		Tenant:     tnt,
		EventStart: evStart,
	}
	//check if we have APIKey in event and in case it has add it in ArgDispatcher
	apiKey, errAPIKey := evStart.GetString(utils.MetaApiKey)
	if errAPIKey == nil {
		s.ArgDispatcher = &utils.ArgDispatcher{
			APIKey: utils.StringPointer(apiKey),
		}
	}
	//check if we have RouteID in event and in case it has add it in ArgDispatcher
	if routeID, err := evStart.GetString(utils.MetaRouteID); err == nil {
		if errAPIKey == utils.ErrNotFound { //in case we don't have APIKey, but we have RouteID we need to initialize the struct
			s.ArgDispatcher = &utils.ArgDispatcher{
				RouteID: utils.StringPointer(routeID),
			}
		} else {
			s.ArgDispatcher.RouteID = utils.StringPointer(routeID)
		}
	}
	if err = sS.forkSession(s, forceDuration); err != nil {
		return
	}
	var maxUsageSet bool // so we know if we have set the 0 on purpose
	for _, sr := range s.SRuns {
		var rplyMaxUsage time.Duration
		if !authReqs.HasField(
			sr.Event.GetStringIgnoreErrors(utils.RequestType)) {
			rplyMaxUsage = eventUsage
		} else if err = sS.connMgr.Call(sS.cgrCfg.SessionSCfg().RALsConns, nil,
			utils.ResponderGetMaxSessionTime,
			&engine.CallDescriptorWithArgDispatcher{CallDescriptor: sr.CD,
				ArgDispatcher: s.ArgDispatcher}, &rplyMaxUsage); err != nil {
			err = utils.NewErrRALs(err)
			return
		}
		if rplyMaxUsage > eventUsage {
			rplyMaxUsage = eventUsage
		}
		if !maxUsageSet || rplyMaxUsage < maxUsage {
			maxUsage = rplyMaxUsage
			maxUsageSet = true
		}
	}
	return
}

// initSession handles a new session
// not thread-safe for Session since it is constructed here
func (sS *SessionS) initSession(tnt string, evStart engine.MapEvent, clntConnID string,
	resID string, dbtItval time.Duration, argDisp *utils.ArgDispatcher, isMsg, forceDuration bool) (s *Session, err error) {
	cgrID := GetSetCGRID(evStart)
	if !evStart.HasField(utils.Usage) && evStart.HasField(utils.LastUsed) {
		evStart[utils.Usage] = evStart[utils.LastUsed]
	}
	s = &Session{
		CGRID:         cgrID,
		Tenant:        tnt,
		ResourceID:    resID,
		EventStart:    evStart.Clone(),
		ClientConnID:  clntConnID,
		DebitInterval: dbtItval,
		ArgDispatcher: argDisp,
	}
	if !isMsg && sS.isIndexed(s, false) { // check if already exists
		return nil, utils.ErrExists
	}
	if err = sS.forkSession(s, forceDuration); err != nil {
		return nil, err
	}
	if !isMsg {
		s.Lock() // avoid endsession before initialising
		sS.initSessionDebitLoops(s)
		sS.registerSession(s, false)
		s.Unlock()
	}
	return
}

// updateSession will reset terminator, perform debits and replicate sessions
func (sS *SessionS) updateSession(s *Session, updtEv engine.MapEvent, isMsg, forceDuration bool) (maxUsage time.Duration, err error) {
	if !isMsg {
		defer sS.replicateSessions(s.CGRID, false, sS.cgrCfg.SessionSCfg().ReplicationConns)
		s.Lock()
		defer s.Unlock()

		// update fields from new event
		for k, v := range updtEv {
			if utils.ProtectedSFlds.Has(k) {
				continue
			}
			s.EventStart[k] = v // update previoius field with new one
		}
		s.updateSRuns(updtEv, sS.cgrCfg.SessionSCfg().AlterableFields)
		sS.setSTerminator(s) // reset the terminator
	}
	//init has no updtEv
	if updtEv == nil {
		updtEv = engine.MapEvent(s.EventStart.Clone())
	}

	var reqMaxUsage time.Duration
	if reqMaxUsage, err = updtEv.GetDuration(utils.Usage); err != nil {
		if err != utils.ErrNotFound {
			return
		}
		err = nil
		reqMaxUsage = sS.cgrCfg.SessionSCfg().GetDefaultUsage(updtEv.GetStringIgnoreErrors(utils.ToR))
		updtEv[utils.Usage] = reqMaxUsage
	}
	var maxUsageSet bool // so we know if we have set the 0 on purpose
	for i, sr := range s.SRuns {
		reqType := sr.Event.GetStringIgnoreErrors(utils.RequestType)
		if reqType == utils.META_NONE {
			continue
		}
		var rplyMaxUsage time.Duration
		if reqType != utils.META_PREPAID || s.debitStop != nil {
			rplyMaxUsage = reqMaxUsage
		} else if rplyMaxUsage, err = sS.debitSession(s, i, reqMaxUsage,
			updtEv.GetDurationPtrIgnoreErrors(utils.LastUsed)); err != nil {
			return
		}
		if !maxUsageSet || rplyMaxUsage < maxUsage {
			maxUsage = rplyMaxUsage
			maxUsageSet = true
		}
	}
	return
}

// terminateSession will end a session from outside
// calls endSession thread safe
func (sS *SessionS) terminateSession(s *Session, tUsage, lastUsage *time.Duration,
	aTime *time.Time, isMsg bool) (err error) {
	s.Lock()
	err = sS.endSession(s, tUsage, lastUsage, aTime, isMsg)
	s.Unlock()
	return
}

// endSession will end a session from outside
// this function is not thread safe
func (sS *SessionS) endSession(s *Session, tUsage, lastUsage *time.Duration,
	aTime *time.Time, isMsg bool) (err error) {
	if !isMsg {
		//check if we have replicate connection and close the session there
		defer sS.replicateSessions(s.CGRID, true, sS.cgrCfg.SessionSCfg().ReplicationConns)
		sS.unregisterSession(s.CGRID, false)
		s.stopSTerminator()
		s.stopDebitLoops()
	}
	for sRunIdx, sr := range s.SRuns {
		sUsage := sr.TotalUsage
		if tUsage != nil {
			sUsage = *tUsage
			sr.TotalUsage = *tUsage
		} else if lastUsage != nil &&
			sr.LastUsage != *lastUsage {
			sr.TotalUsage -= sr.LastUsage
			sr.TotalUsage += *lastUsage
			sUsage = sr.TotalUsage
		}
		if sr.EventCost != nil {
			if !isMsg { // in case of one time charge there is no need of corrections
				if notCharged := sUsage - sr.EventCost.GetUsage(); notCharged > 0 { // we did not charge enough, make a manual debit here
					if sr.CD.LoopIndex > 0 {
						sr.CD.TimeStart = sr.CD.TimeEnd
					}
					sr.CD.TimeEnd = sr.CD.TimeStart.Add(notCharged)
					sr.CD.DurationIndex += notCharged
					cc := new(engine.CallCost)
					if err = sS.connMgr.Call(sS.cgrCfg.SessionSCfg().RALsConns, nil, utils.ResponderDebit,
						&engine.CallDescriptorWithArgDispatcher{
							CallDescriptor: sr.CD,
							ArgDispatcher:  s.ArgDispatcher}, cc); err == nil {
						sr.EventCost.Merge(
							engine.NewEventCostFromCallCost(cc, s.CGRID,
								sr.Event.GetStringIgnoreErrors(utils.RunID)))
					}
				} else if notCharged < 0 { // charged too much, try refund
					if err = sS.refundSession(s, sRunIdx, -notCharged); err != nil {
						utils.Logger.Warning(
							fmt.Sprintf(
								"<%s> failed refunding session: <%s>, srIdx: <%d>, error: <%s>",
								utils.SessionS, s.CGRID, sRunIdx, err.Error()))
					}
				}
				if err := sS.roundCost(s, sRunIdx); err != nil { // will round the cost and refund the extra increment
					utils.Logger.Warning(
						fmt.Sprintf("<%s> failed rounding  session cost for <%s>, srIdx: <%d>, error: <%s>",
							utils.SessionS, s.CGRID, sRunIdx, err.Error()))
				}
			}
			// compute the event cost before saving the SessionCost
			// add here to be applied for messages also
			sr.EventCost.Compute()
			if sS.cgrCfg.SessionSCfg().StoreSCosts {
				if err := sS.storeSCost(s, sRunIdx); err != nil {
					utils.Logger.Warning(
						fmt.Sprintf("<%s> failed storing session cost for <%s>, srIdx: <%d>, error: <%s>",
							utils.SessionS, s.CGRID, sRunIdx, err.Error()))
				}
			}
			// set cost fields
			sr.Event[utils.Cost] = sr.EventCost.GetCost()
			sr.Event[utils.CostDetails] = utils.ToJSON(sr.EventCost) // avoid map[string]interface{} when decoding
			sr.Event[utils.CostSource] = utils.MetaSessionS
		}
		// Set Usage field
		if sRunIdx == 0 {
			s.EventStart[utils.Usage] = sr.TotalUsage
		}
		sr.Event[utils.Usage] = sr.TotalUsage
		if aTime != nil {
			sr.Event[utils.AnswerTime] = *aTime
		}
	}
	engine.Cache.Set(utils.CacheClosedSessions, s.CGRID, s,
		nil, true, utils.NonTransactional)
	return
}

// chargeEvent will charge a single event (ie: SMS)
func (sS *SessionS) chargeEvent(tnt string, ev engine.MapEvent,
	argDisp *utils.ArgDispatcher, forceDuration bool) (maxUsage time.Duration, err error) {
	cgrID := GetSetCGRID(ev)
	var s *Session
	if s, err = sS.initSession(tnt, ev, "", "", 0, argDisp, true, forceDuration); err != nil {
		return
	}
	if maxUsage, err = sS.updateSession(s, nil, true, forceDuration); err != nil {
		if errEnd := sS.terminateSession(s,
			utils.DurationPointer(time.Duration(0)), nil, nil, true); errEnd != nil {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> error when force-ending charged event: <%s>, err: <%s>",
					utils.SessionS, cgrID, errEnd.Error()))
		}
		err = utils.NewErrRALs(err)
		return
	}
	usage := maxUsage
	if utils.SliceHasMember(utils.PostPaidRatedSlice, ev.GetStringIgnoreErrors(utils.RequestType)) {
		usage = ev.GetDurationIgnoreErrors(utils.Usage)
	}
	//in case of postpaid and rated maxUsage = usage from event
	if errEnd := sS.terminateSession(s, utils.DurationPointer(usage), nil, nil, true); errEnd != nil {
		utils.Logger.Warning(
			fmt.Sprintf("<%s> error when ending charged event: <%s>, err: <%s>",
				utils.SessionS, cgrID, errEnd.Error()))
	}
	return // returns here the maxUsage from update
}

// APIs start here

// Call is part of RpcClientConnection interface
func (sS *SessionS) Call(serviceMethod string, args interface{}, reply interface{}) error {
	return sS.CallBiRPC(nil, serviceMethod, args, reply)
}

// CallBiRPC is part of utils.BiRPCServer interface to help internal connections do calls over rpcclient.ClientConnector interface
func (sS *SessionS) CallBiRPC(clnt rpcclient.ClientConnector,
	serviceMethod string, args interface{}, reply interface{}) error {
	parts := strings.Split(serviceMethod, ".")
	if len(parts) != 2 {
		return rpcclient.ErrUnsupporteServiceMethod
	}
	// get method BiRPCV1.Method
	method := reflect.ValueOf(sS).MethodByName(
		"BiRPC" + parts[0][len(parts[0])-2:] + parts[1]) // Inherit the version V1 in the method name and add prefix
	if !method.IsValid() {
		return rpcclient.ErrUnsupporteServiceMethod
	}
	// construct the params
	var clntVal reflect.Value
	if clnt == nil {
		clntVal = reflect.New(
			reflect.TypeOf(new(utils.BiRPCInternalClient))).Elem() // Kinda cheat since we make up a type here
	} else {
		clntVal = reflect.ValueOf(clnt)
	}
	params := []reflect.Value{clntVal, reflect.ValueOf(args),
		reflect.ValueOf(reply)}
	ret := method.Call(params)
	if len(ret) != 1 {
		return utils.ErrServerError
	}
	if ret[0].Interface() == nil {
		return nil
	}
	err, ok := ret[0].Interface().(error)
	if !ok {
		return utils.ErrServerError
	}
	return err
}

// BiRPCv1GetActiveSessions returns the list of active sessions based on filter
func (sS *SessionS) BiRPCv1GetActiveSessions(clnt rpcclient.ClientConnector,
	args *utils.SessionFilter, reply *[]*ExternalSession) (err error) {
	if args == nil { //protection in case on nil
		args = &utils.SessionFilter{}
	}
	aSs := sS.filterSessions(args, false)
	if len(aSs) == 0 {
		return utils.ErrNotFound
	}
	*reply = aSs
	return nil
}

// BiRPCv1GetActiveSessionsCount counts the active sessions
func (sS *SessionS) BiRPCv1GetActiveSessionsCount(clnt rpcclient.ClientConnector,
	args *utils.SessionFilter, reply *int) error {
	if args == nil { //protection in case on nil
		args = &utils.SessionFilter{}
	}
	*reply = sS.filterSessionsCount(args, false)
	return nil
}

// BiRPCv1GetPassiveSessions returns the passive sessions handled by SessionS
func (sS *SessionS) BiRPCv1GetPassiveSessions(clnt rpcclient.ClientConnector,
	args *utils.SessionFilter, reply *[]*ExternalSession) error {
	if args == nil { //protection in case on nil
		args = &utils.SessionFilter{}
	}
	pSs := sS.filterSessions(args, true)
	if len(pSs) == 0 {
		return utils.ErrNotFound
	}
	*reply = pSs
	return nil
}

// BiRPCv1GetPassiveSessionsCount counts the passive sessions handled by the system
func (sS *SessionS) BiRPCv1GetPassiveSessionsCount(clnt rpcclient.ClientConnector,
	args *utils.SessionFilter, reply *int) error {
	if args == nil { //protection in case on nil
		args = &utils.SessionFilter{}
	}
	*reply = sS.filterSessionsCount(args, true)
	return nil
}

// BiRPCv1SetPassiveSession used for replicating Sessions
func (sS *SessionS) BiRPCv1SetPassiveSession(clnt rpcclient.ClientConnector,
	s *Session, reply *string) (err error) {
	if s.CGRID == "" {
		return utils.NewErrMandatoryIeMissing(utils.CGRID)
	}
	if s.EventStart == nil { // remove
		if ureg := sS.unregisterSession(s.CGRID, true); !ureg {
			return utils.ErrNotFound
		}
		*reply = utils.OK
		return
	}
	if aSs := sS.getSessions(s.CGRID, false); len(aSs) != 0 { // found active session, transit to passive
		aSs[0].Lock()
		sS.unregisterSession(s.CGRID, false)
		aSs[0].stopSTerminator()
		aSs[0].stopDebitLoops()
		aSs[0].Unlock()
	}
	sS.registerSession(s, true)

	*reply = utils.OK
	return
}

// ArgsReplicateSessions used to specify wich Session to replicate over the given connections
type ArgsReplicateSessions struct {
	CGRID   string
	Passive bool
	ConnIDs []string
}

// BiRPCv1ReplicateSessions will replicate active sessions to either args.Connections or the internal configured ones
// args.Filter is used to filter the sessions which are replicated, CGRID is the only one possible for now
func (sS *SessionS) BiRPCv1ReplicateSessions(clnt rpcclient.ClientConnector,
	args ArgsReplicateSessions, reply *string) (err error) {
	if err = sS.replicateSessions(args.CGRID, args.Passive, args.ConnIDs); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = utils.OK
	return
}

// NewV1AuthorizeArgs is a constructor for V1AuthorizeArgs
func NewV1AuthorizeArgs(attrs bool, attributeIDs []string,
	thrslds bool, thresholdIDs []string, statQueues bool, statIDs []string,
	res, maxUsage, suppls, supplsIgnoreErrs, supplsEventCost bool,
	cgrEv *utils.CGREvent, argDisp *utils.ArgDispatcher,
	supplierPaginator utils.Paginator, forceDuration bool, supMaxCost string) (args *V1AuthorizeArgs) {
	args = &V1AuthorizeArgs{
		GetAttributes:         attrs,
		AuthorizeResources:    res,
		GetMaxUsage:           maxUsage,
		ProcessThresholds:     thrslds,
		ProcessStats:          statQueues,
		SuppliersIgnoreErrors: supplsIgnoreErrs,
		GetSuppliers:          suppls,
		CGREvent:              cgrEv,
		ForceDuration:         forceDuration,
	}
	if supplsEventCost {
		args.SuppliersMaxCost = utils.MetaEventCost
	} else {
		args.SuppliersMaxCost = supMaxCost
	}
	args.ArgDispatcher = argDisp
	args.Paginator = supplierPaginator
	if len(attributeIDs) != 0 {
		args.AttributeIDs = attributeIDs
	}
	if len(thresholdIDs) != 0 {
		args.ThresholdIDs = thresholdIDs
	}
	if len(statIDs) != 0 {
		args.StatIDs = statIDs
	}

	return
}

// V1AuthorizeArgs are options available in auth request
type V1AuthorizeArgs struct {
	GetAttributes         bool
	AuthorizeResources    bool
	GetMaxUsage           bool
	ProcessThresholds     bool
	ProcessStats          bool
	GetSuppliers          bool
	ForceDuration         bool
	SuppliersMaxCost      string
	SuppliersIgnoreErrors bool
	AttributeIDs          []string
	ThresholdIDs          []string
	StatIDs               []string
	*utils.CGREvent
	utils.Paginator
	*utils.ArgDispatcher
}

// ParseFlags will populate the V1AuthorizeArgs flags
func (args *V1AuthorizeArgs) ParseFlags(flags string) {
	dispatcherFlag := false
	for _, subsystem := range strings.Split(flags, utils.FIELDS_SEP) {
		switch {
		case subsystem == utils.MetaAccounts:
			args.GetMaxUsage = true
		case subsystem == utils.MetaResources:
			args.AuthorizeResources = true
		case subsystem == utils.MetaDispatchers:
			dispatcherFlag = true
		case subsystem == utils.MetaSuppliers:
			args.GetSuppliers = true
		case subsystem == utils.MetaSuppliersIgnoreErrors:
			args.SuppliersIgnoreErrors = true
		case subsystem == utils.MetaSuppliersEventCost:
			args.SuppliersMaxCost = utils.MetaEventCost
		case strings.HasPrefix(subsystem, utils.MetaSuppliersMaxCost):
			args.SuppliersMaxCost = strings.TrimPrefix(subsystem, utils.MetaSuppliersMaxCost+utils.InInFieldSep)
		case strings.HasPrefix(subsystem, utils.MetaAttributes):
			args.GetAttributes = true
			args.AttributeIDs = getFlagIDs(subsystem)
		case strings.HasPrefix(subsystem, utils.MetaThresholds):
			args.ProcessThresholds = true
			args.ThresholdIDs = getFlagIDs(subsystem)
		case strings.HasPrefix(subsystem, utils.MetaStats):
			args.ProcessStats = true
			args.StatIDs = getFlagIDs(subsystem)
		case subsystem == utils.MetaFD:
			args.ForceDuration = true
		}
	}
	cgrArgs := args.CGREvent.ExtractArgs(dispatcherFlag, true)
	args.ArgDispatcher = cgrArgs.ArgDispatcher
	args.Paginator = *cgrArgs.SupplierPaginator
}

// V1AuthorizeReply are options available in auth reply
type V1AuthorizeReply struct {
	Attributes         *engine.AttrSProcessEventReply
	ResourceAllocation *string
	MaxUsage           time.Duration
	Suppliers          *engine.SortedSuppliers
	ThresholdIDs       *[]string
	StatQueueIDs       *[]string
	getMaxUsage        bool // used by AsNavigableMap to know if MaxUsage is needed
}

// SetMaxUsageNeeded used by agent that use the reply as NavigableMapper
func (v1AuthReply *V1AuthorizeReply) SetMaxUsageNeeded(getMaxUsage bool) {
	if v1AuthReply == nil {
		return
	}
	v1AuthReply.getMaxUsage = getMaxUsage
}

// AsNavigableMap is part of engine.NavigableMapper interface
func (v1AuthReply *V1AuthorizeReply) AsNavigableMap() utils.NavigableMap2 {
	cgrReply := make(utils.NavigableMap2)
	if v1AuthReply.Attributes != nil {
		attrs := make(utils.NavigableMap2)
		for _, fldName := range v1AuthReply.Attributes.AlteredFields {
			fldName = strings.TrimPrefix(fldName, utils.MetaReq+utils.NestingSep)
			if v1AuthReply.Attributes.CGREvent.HasField(fldName) {
				attrs[fldName] = utils.NewNMData(v1AuthReply.Attributes.CGREvent.Event[fldName])
			}
		}
		cgrReply[utils.CapAttributes] = attrs
	}
	if v1AuthReply.ResourceAllocation != nil {
		cgrReply[utils.CapResourceAllocation] = utils.NewNMData(*v1AuthReply.ResourceAllocation)
	}
	if v1AuthReply.getMaxUsage {
		cgrReply[utils.CapMaxUsage] = utils.NewNMData(v1AuthReply.MaxUsage)
	}
	if v1AuthReply.Suppliers != nil {
		cgrReply[utils.CapSuppliers] = v1AuthReply.Suppliers.AsNavigableMap()
	}
	if v1AuthReply.ThresholdIDs != nil {
		thIDs := make(utils.NMSlice, len(*v1AuthReply.ThresholdIDs))
		for i, v := range *v1AuthReply.ThresholdIDs {
			thIDs[i] = utils.NewNMData(v)
		}
		cgrReply[utils.CapThresholds] = &thIDs
	}
	if v1AuthReply.StatQueueIDs != nil {
		stIDs := make(utils.NMSlice, len(*v1AuthReply.StatQueueIDs))
		for i, v := range *v1AuthReply.StatQueueIDs {
			stIDs[i] = utils.NewNMData(v)
		}
		cgrReply[utils.CapStatQueues] = &stIDs
	}
	return cgrReply
}

// BiRPCv1AuthorizeEvent performs authorization for CGREvent based on specific components
func (sS *SessionS) BiRPCv1AuthorizeEvent(clnt rpcclient.ClientConnector,
	args *V1AuthorizeArgs, authReply *V1AuthorizeReply) (err error) {
	if args.CGREvent == nil {
		return utils.NewErrMandatoryIeMissing(utils.CGREventString)
	}
	var withErrors bool
	if args.CGREvent.ID == "" {
		args.CGREvent.ID = utils.GenUUID()
	}
	// RPC caching
	if sS.cgrCfg.CacheCfg()[utils.CacheRPCResponses].Limit != 0 {
		cacheKey := utils.ConcatenatedKey(utils.SessionSv1AuthorizeEvent, args.CGREvent.ID)
		refID := guardian.Guardian.GuardIDs("",
			sS.cgrCfg.GeneralCfg().LockingTimeout, cacheKey) // RPC caching needs to be atomic
		defer guardian.Guardian.UnguardIDs(refID)

		if itm, has := engine.Cache.Get(utils.CacheRPCResponses, cacheKey); has {
			cachedResp := itm.(*utils.CachedRPCResponse)
			if cachedResp.Error == nil {
				*authReply = *cachedResp.Result.(*V1AuthorizeReply)
			}
			return cachedResp.Error
		}
		defer engine.Cache.Set(utils.CacheRPCResponses, cacheKey,
			&utils.CachedRPCResponse{Result: authReply, Error: err},
			nil, true, utils.NonTransactional)
	}
	// end of RPC caching

	if !args.GetAttributes && !args.AuthorizeResources &&
		!args.GetMaxUsage && !args.GetSuppliers {
		return utils.NewErrMandatoryIeMissing("subsystems")
	}
	if args.CGREvent.Tenant == "" {
		args.CGREvent.Tenant = sS.cgrCfg.GeneralCfg().DefaultTenant
	}
	if args.GetAttributes {
		rplyAttr, err := sS.processAttributes(args.CGREvent, args.ArgDispatcher,
			args.AttributeIDs)
		if err == nil {
			args.CGREvent = rplyAttr.CGREvent
			authReply.Attributes = &rplyAttr
		} else if err.Error() != utils.ErrNotFound.Error() {
			return utils.NewErrAttributeS(err)
		}
	}
	if args.GetMaxUsage {
		if authReply.MaxUsage, err = sS.authEvent(args.CGREvent.Tenant,
			args.CGREvent.Event, args.ForceDuration); err != nil {
			return err
		}
	}
	if args.AuthorizeResources {
		if len(sS.cgrCfg.SessionSCfg().ResSConns) == 0 {
			return utils.NewErrNotConnected(utils.ResourceS)
		}
		originID, _ := args.CGREvent.FieldAsString(utils.OriginID)
		if originID == "" {
			originID = utils.UUIDSha1Prefix()
		}
		var allocMsg string
		attrRU := utils.ArgRSv1ResourceUsage{
			CGREvent:      args.CGREvent,
			UsageID:       originID,
			Units:         1,
			ArgDispatcher: args.ArgDispatcher,
		}
		if err = sS.connMgr.Call(sS.cgrCfg.SessionSCfg().ResSConns, nil, utils.ResourceSv1AuthorizeResources,
			attrRU, &allocMsg); err != nil {
			return utils.NewErrResourceS(err)
		}
		authReply.ResourceAllocation = &allocMsg
	}
	if args.GetSuppliers {
		splsReply, err := sS.getSuppliers(args.CGREvent.Clone(), args.ArgDispatcher,
			args.Paginator, args.SuppliersIgnoreErrors, args.SuppliersMaxCost)
		if err != nil {
			return err
		}
		if splsReply.SortedSuppliers != nil {
			authReply.Suppliers = &splsReply
		}
	}
	if args.ProcessThresholds {
		tIDs, err := sS.processThreshold(args.CGREvent, args.ArgDispatcher,
			args.ThresholdIDs)
		if err != nil && err.Error() != utils.ErrNotFound.Error() {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> error: %s processing event %+v with ThresholdS.",
					utils.SessionS, err.Error(), args.CGREvent))
			withErrors = true
		}
		authReply.ThresholdIDs = &tIDs
	}
	if args.ProcessStats {
		sIDs, err := sS.processStats(args.CGREvent, args.ArgDispatcher,
			args.StatIDs)
		if err != nil &&
			err.Error() != utils.ErrNotFound.Error() {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> error: %s processing event %+v with StatS.",
					utils.SessionS, err.Error(), args.CGREvent))
			withErrors = true
		}
		authReply.StatQueueIDs = &sIDs
	}
	if withErrors {
		err = utils.ErrPartiallyExecuted
	}
	return
}

// V1AuthorizeReplyWithDigest contains return options for auth with digest
type V1AuthorizeReplyWithDigest struct {
	AttributesDigest   *string
	ResourceAllocation *string
	MaxUsage           float64 // special treat returning time.Duration.Seconds()
	SuppliersDigest    *string
	Thresholds         *string
	StatQueues         *string
}

// BiRPCv1AuthorizeEventWithDigest performs authorization for CGREvent based on specific components
// returning one level fields instead of multiple ones returned by BiRPCv1AuthorizeEvent
func (sS *SessionS) BiRPCv1AuthorizeEventWithDigest(clnt rpcclient.ClientConnector,
	args *V1AuthorizeArgs, authReply *V1AuthorizeReplyWithDigest) (err error) {
	var initAuthRply V1AuthorizeReply
	if err = sS.BiRPCv1AuthorizeEvent(clnt, args, &initAuthRply); err != nil {
		return
	}
	if args.GetAttributes && initAuthRply.Attributes != nil {
		authReply.AttributesDigest = utils.StringPointer(initAuthRply.Attributes.Digest())
	}
	if args.AuthorizeResources {
		authReply.ResourceAllocation = initAuthRply.ResourceAllocation
	}
	if args.GetMaxUsage {
		authReply.MaxUsage = initAuthRply.MaxUsage.Seconds()
	}
	if args.GetSuppliers {
		authReply.SuppliersDigest = utils.StringPointer(initAuthRply.Suppliers.Digest())
	}
	if args.ProcessThresholds {
		authReply.Thresholds = utils.StringPointer(
			strings.Join(*initAuthRply.ThresholdIDs, utils.FIELDS_SEP))
	}
	if args.ProcessStats {
		authReply.StatQueues = utils.StringPointer(
			strings.Join(*initAuthRply.StatQueueIDs, utils.FIELDS_SEP))
	}
	return nil
}

// NewV1InitSessionArgs is a constructor for V1InitSessionArgs
func NewV1InitSessionArgs(attrs bool, attributeIDs []string,
	thrslds bool, thresholdIDs []string, stats bool, statIDs []string,
	resrc, acnt bool, cgrEv *utils.CGREvent,
	argDisp *utils.ArgDispatcher, forceDuration bool) (args *V1InitSessionArgs) {
	args = &V1InitSessionArgs{
		GetAttributes:     attrs,
		AllocateResources: resrc,
		InitSession:       acnt,
		ProcessThresholds: thrslds,
		ProcessStats:      stats,
		CGREvent:          cgrEv,
		ArgDispatcher:     argDisp,
		ForceDuration:     forceDuration,
	}
	if len(attributeIDs) != 0 {
		args.AttributeIDs = attributeIDs
	}
	if len(thresholdIDs) != 0 {
		args.ThresholdIDs = thresholdIDs
	}
	if len(statIDs) != 0 {
		args.StatIDs = statIDs
	}
	return
}

// V1InitSessionArgs are options for session initialization request
type V1InitSessionArgs struct {
	GetAttributes     bool
	AllocateResources bool
	InitSession       bool
	ForceDuration     bool
	ProcessThresholds bool
	ProcessStats      bool
	AttributeIDs      []string
	ThresholdIDs      []string
	StatIDs           []string
	*utils.CGREvent
	*utils.ArgDispatcher
}

// ParseFlags will populate the V1InitSessionArgs flags
func (args *V1InitSessionArgs) ParseFlags(flags string) {
	dispatcherFlag := false
	for _, subsystem := range strings.Split(flags, utils.FIELDS_SEP) {
		switch {
		case subsystem == utils.MetaAccounts:
			args.InitSession = true
		case subsystem == utils.MetaResources:
			args.AllocateResources = true
		case subsystem == utils.MetaDispatchers:
			dispatcherFlag = true
		case strings.HasPrefix(subsystem, utils.MetaAttributes):
			args.GetAttributes = true
			args.AttributeIDs = getFlagIDs(subsystem)
		case strings.HasPrefix(subsystem, utils.MetaThresholds):
			args.ProcessThresholds = true
			args.ThresholdIDs = getFlagIDs(subsystem)
		case strings.HasPrefix(subsystem, utils.MetaStats):
			args.ProcessStats = true
			args.StatIDs = getFlagIDs(subsystem)
		case subsystem == utils.MetaFD:
			args.ForceDuration = true
		}
	}
	cgrArgs := args.CGREvent.ExtractArgs(dispatcherFlag, false)
	args.ArgDispatcher = cgrArgs.ArgDispatcher
}

// V1InitSessionReply are options for initialization reply
type V1InitSessionReply struct {
	Attributes         *engine.AttrSProcessEventReply
	ResourceAllocation *string
	MaxUsage           time.Duration
	ThresholdIDs       *[]string
	StatQueueIDs       *[]string
	getMaxUsage        bool
}

// SetMaxUsageNeeded used by agent that use the reply as NavigableMapper
func (v1Rply *V1InitSessionReply) SetMaxUsageNeeded(getMaxUsage bool) {
	if v1Rply == nil {
		return
	}
	v1Rply.getMaxUsage = getMaxUsage
}

// AsNavigableMap is part of engine.NavigableMapper interface
func (v1Rply *V1InitSessionReply) AsNavigableMap() utils.NavigableMap2 {
	cgrReply := make(utils.NavigableMap2)
	if v1Rply.Attributes != nil {
		attrs := make(utils.NavigableMap2)
		for _, fldName := range v1Rply.Attributes.AlteredFields {
			fldName = strings.TrimPrefix(fldName, utils.MetaReq+utils.NestingSep)
			if v1Rply.Attributes.CGREvent.HasField(fldName) {
				attrs[fldName] = utils.NewNMData(v1Rply.Attributes.CGREvent.Event[fldName])
			}
		}
		cgrReply[utils.CapAttributes] = attrs
	}
	if v1Rply.ResourceAllocation != nil {
		cgrReply[utils.CapResourceAllocation] = utils.NewNMData(*v1Rply.ResourceAllocation)
	}
	if v1Rply.getMaxUsage {
		cgrReply[utils.CapMaxUsage] = utils.NewNMData(v1Rply.MaxUsage)
	}

	if v1Rply.ThresholdIDs != nil {
		thIDs := make(utils.NMSlice, len(*v1Rply.ThresholdIDs))
		for i, v := range *v1Rply.ThresholdIDs {
			thIDs[i] = utils.NewNMData(v)
		}
		cgrReply[utils.CapThresholds] = &thIDs
	}
	if v1Rply.StatQueueIDs != nil {
		stIDs := make(utils.NMSlice, len(*v1Rply.StatQueueIDs))
		for i, v := range *v1Rply.StatQueueIDs {
			stIDs[i] = utils.NewNMData(v)
		}
		cgrReply[utils.CapStatQueues] = &stIDs
	}
	return cgrReply
}

// BiRPCv1InitiateSession initiates a new session
func (sS *SessionS) BiRPCv1InitiateSession(clnt rpcclient.ClientConnector,
	args *V1InitSessionArgs, rply *V1InitSessionReply) (err error) {
	if args.CGREvent == nil {
		return utils.NewErrMandatoryIeMissing(utils.CGREventString)
	}
	var withErrors bool
	if args.CGREvent.ID == "" {
		args.CGREvent.ID = utils.GenUUID()
	}

	// RPC caching
	if sS.cgrCfg.CacheCfg()[utils.CacheRPCResponses].Limit != 0 {
		cacheKey := utils.ConcatenatedKey(utils.SessionSv1InitiateSession, args.CGREvent.ID)
		refID := guardian.Guardian.GuardIDs("",
			sS.cgrCfg.GeneralCfg().LockingTimeout, cacheKey) // RPC caching needs to be atomic
		defer guardian.Guardian.UnguardIDs(refID)

		if itm, has := engine.Cache.Get(utils.CacheRPCResponses, cacheKey); has {
			cachedResp := itm.(*utils.CachedRPCResponse)
			if cachedResp.Error == nil {
				*rply = *cachedResp.Result.(*V1InitSessionReply)
			}
			return cachedResp.Error
		}
		defer engine.Cache.Set(utils.CacheRPCResponses, cacheKey,
			&utils.CachedRPCResponse{Result: rply, Error: err},
			nil, true, utils.NonTransactional)
	}
	// end of RPC caching

	if !args.GetAttributes && !args.AllocateResources && !args.InitSession {
		return utils.NewErrMandatoryIeMissing("subsystems")
	}
	if args.CGREvent.Tenant == "" {
		args.CGREvent.Tenant = sS.cgrCfg.GeneralCfg().DefaultTenant
	}
	originID, _ := args.CGREvent.FieldAsString(utils.OriginID)
	if args.GetAttributes {
		rplyAttr, err := sS.processAttributes(args.CGREvent, args.ArgDispatcher,
			args.AttributeIDs)
		if err == nil {
			args.CGREvent = rplyAttr.CGREvent
			rply.Attributes = &rplyAttr
		} else if err.Error() != utils.ErrNotFound.Error() {
			return utils.NewErrAttributeS(err)
		}
	}
	if args.AllocateResources {
		if len(sS.cgrCfg.SessionSCfg().ResSConns) == 0 {
			return utils.NewErrNotConnected(utils.ResourceS)
		}
		if originID == "" {
			return utils.NewErrMandatoryIeMissing(utils.OriginID)
		}
		attrRU := utils.ArgRSv1ResourceUsage{
			CGREvent:      args.CGREvent,
			UsageID:       originID,
			Units:         1,
			ArgDispatcher: args.ArgDispatcher,
		}
		var allocMessage string
		if err = sS.connMgr.Call(sS.cgrCfg.SessionSCfg().ResSConns, nil, utils.ResourceSv1AllocateResources,
			attrRU, &allocMessage); err != nil {
			return utils.NewErrResourceS(err)
		}
		rply.ResourceAllocation = &allocMessage
	}
	if args.InitSession {
		var err error
		ev := engine.MapEvent(args.CGREvent.Event)
		dbtItvl := sS.cgrCfg.SessionSCfg().DebitInterval
		if ev.HasField(utils.CGRDebitInterval) { // dynamic DebitInterval via CGRDebitInterval
			if dbtItvl, err = ev.GetDuration(utils.CGRDebitInterval); err != nil {
				return utils.NewErrRALs(err)
			}
		}
		s, err := sS.initSession(args.CGREvent.Tenant, ev,
			sS.biJClntID(clnt), originID, dbtItvl, args.ArgDispatcher, false, args.ForceDuration)
		if err != nil {
			return err
		}
		s.RLock() // avoid concurrency with activeDebit
		isPrepaid := s.debitStop != nil
		s.RUnlock()
		if isPrepaid { //active debit
			rply.MaxUsage = sS.cgrCfg.SessionSCfg().GetDefaultUsage(ev.GetStringIgnoreErrors(utils.ToR))
		} else {
			var maxUsage time.Duration
			if maxUsage, err = sS.updateSession(s, nil, false, args.ForceDuration); err != nil {
				return utils.NewErrRALs(err)
			}
			rply.MaxUsage = maxUsage
		}
	}
	if args.ProcessThresholds {
		tIDs, err := sS.processThreshold(args.CGREvent, args.ArgDispatcher,
			args.ThresholdIDs)
		if err != nil && err.Error() != utils.ErrNotFound.Error() {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> error: %s processing event %+v with ThresholdS.",
					utils.SessionS, err.Error(), args.CGREvent))
			withErrors = true
		}
		rply.ThresholdIDs = &tIDs
	}
	if args.ProcessStats {
		sIDs, err := sS.processStats(args.CGREvent, args.ArgDispatcher,
			args.StatIDs)
		if err != nil &&
			err.Error() != utils.ErrNotFound.Error() {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> error: %s processing event %+v with StatS.",
					utils.SessionS, err.Error(), args.CGREvent))
			withErrors = true
		}
		rply.StatQueueIDs = &sIDs
	}
	if withErrors {
		err = utils.ErrPartiallyExecuted
	}
	return
}

// V1InitReplyWithDigest is the formated reply
type V1InitReplyWithDigest struct {
	AttributesDigest   *string
	ResourceAllocation *string
	MaxUsage           float64
	Thresholds         *string
	StatQueues         *string
}

// BiRPCv1InitiateSessionWithDigest returns the formated result of InitiateSession
func (sS *SessionS) BiRPCv1InitiateSessionWithDigest(clnt rpcclient.ClientConnector,
	args *V1InitSessionArgs, initReply *V1InitReplyWithDigest) (err error) {
	var initSessionRply V1InitSessionReply
	if err = sS.BiRPCv1InitiateSession(clnt, args, &initSessionRply); err != nil {
		return
	}

	if args.GetAttributes &&
		initSessionRply.Attributes != nil {
		initReply.AttributesDigest = utils.StringPointer(initSessionRply.Attributes.Digest())
	}

	if args.AllocateResources {
		initReply.ResourceAllocation = initSessionRply.ResourceAllocation
	}

	if args.InitSession {
		initReply.MaxUsage = initSessionRply.MaxUsage.Seconds()
	}

	if args.ProcessThresholds {
		initReply.Thresholds = utils.StringPointer(
			strings.Join(*initSessionRply.ThresholdIDs, utils.FIELDS_SEP))
	}
	if args.ProcessStats {
		initReply.StatQueues = utils.StringPointer(
			strings.Join(*initSessionRply.StatQueueIDs, utils.FIELDS_SEP))
	}
	return nil
}

// NewV1UpdateSessionArgs is a constructor for update session arguments
func NewV1UpdateSessionArgs(attrs bool, attributeIDs []string,
	acnts bool, cgrEv *utils.CGREvent,
	argDisp *utils.ArgDispatcher, forceDuration bool) (args *V1UpdateSessionArgs) {
	args = &V1UpdateSessionArgs{
		GetAttributes: attrs,
		UpdateSession: acnts,
		CGREvent:      cgrEv,
		ArgDispatcher: argDisp,
		ForceDuration: forceDuration,
	}
	if len(attributeIDs) != 0 {
		args.AttributeIDs = attributeIDs
	}
	return
}

// V1UpdateSessionArgs contains options for session update
type V1UpdateSessionArgs struct {
	GetAttributes bool
	UpdateSession bool
	ForceDuration bool
	AttributeIDs  []string
	*utils.CGREvent
	*utils.ArgDispatcher
}

// V1UpdateSessionReply contains options for session update reply
type V1UpdateSessionReply struct {
	Attributes  *engine.AttrSProcessEventReply
	MaxUsage    time.Duration
	getMaxUsage bool
}

// SetMaxUsageNeeded used by agent that use the reply as NavigableMapper
func (v1Rply *V1UpdateSessionReply) SetMaxUsageNeeded(getMaxUsage bool) {
	if v1Rply == nil {
		return
	}
	v1Rply.getMaxUsage = getMaxUsage
}

// AsNavigableMap is part of engine.NavigableMapper interface
func (v1Rply *V1UpdateSessionReply) AsNavigableMap() utils.NavigableMap2 {
	cgrReply := make(utils.NavigableMap2)
	if v1Rply.Attributes != nil {
		attrs := make(utils.NavigableMap2)
		for _, fldName := range v1Rply.Attributes.AlteredFields {
			fldName = strings.TrimPrefix(fldName, utils.MetaReq+utils.NestingSep)
			if v1Rply.Attributes.CGREvent.HasField(fldName) {
				attrs[fldName] = utils.NewNMData(v1Rply.Attributes.CGREvent.Event[fldName])
			}
		}
		cgrReply[utils.CapAttributes] = attrs
	}
	if v1Rply.getMaxUsage {
		cgrReply[utils.CapMaxUsage] = utils.NewNMData(v1Rply.MaxUsage)
	}
	return cgrReply
}

// BiRPCv1UpdateSession updates an existing session, returning the duration which the session can still last
func (sS *SessionS) BiRPCv1UpdateSession(clnt rpcclient.ClientConnector,
	args *V1UpdateSessionArgs, rply *V1UpdateSessionReply) (err error) {
	if args.CGREvent == nil {
		return utils.NewErrMandatoryIeMissing(utils.CGREventString)
	}
	if args.CGREvent.ID == "" {
		args.CGREvent.ID = utils.GenUUID()
	}

	// RPC caching
	if sS.cgrCfg.CacheCfg()[utils.CacheRPCResponses].Limit != 0 {
		cacheKey := utils.ConcatenatedKey(utils.SessionSv1UpdateSession, args.CGREvent.ID)
		refID := guardian.Guardian.GuardIDs("",
			sS.cgrCfg.GeneralCfg().LockingTimeout, cacheKey) // RPC caching needs to be atomic
		defer guardian.Guardian.UnguardIDs(refID)

		if itm, has := engine.Cache.Get(utils.CacheRPCResponses, cacheKey); has {
			cachedResp := itm.(*utils.CachedRPCResponse)
			if cachedResp.Error == nil {
				*rply = *cachedResp.Result.(*V1UpdateSessionReply)
			}
			return cachedResp.Error
		}
		defer engine.Cache.Set(utils.CacheRPCResponses, cacheKey,
			&utils.CachedRPCResponse{Result: rply, Error: err},
			nil, true, utils.NonTransactional)
	}
	// end of RPC caching

	if !args.GetAttributes && !args.UpdateSession {
		return utils.NewErrMandatoryIeMissing("subsystems")
	}
	if args.CGREvent.Tenant == "" {
		args.CGREvent.Tenant = sS.cgrCfg.GeneralCfg().DefaultTenant
	}
	if args.GetAttributes {
		rplyAttr, err := sS.processAttributes(args.CGREvent, args.ArgDispatcher,
			args.AttributeIDs)
		if err == nil {
			args.CGREvent = rplyAttr.CGREvent
			rply.Attributes = &rplyAttr
		} else if err.Error() != utils.ErrNotFound.Error() {
			return utils.NewErrAttributeS(err)
		}
	}
	if args.UpdateSession {
		ev := engine.MapEvent(args.CGREvent.Event)
		dbtItvl := sS.cgrCfg.SessionSCfg().DebitInterval
		if ev.HasField(utils.CGRDebitInterval) { // dynamic DebitInterval via CGRDebitInterval
			if dbtItvl, err = ev.GetDuration(utils.CGRDebitInterval); err != nil {
				return utils.NewErrRALs(err)
			}
		}
		cgrID := GetSetCGRID(ev)
		s := sS.getRelocateSession(cgrID,
			ev.GetStringIgnoreErrors(utils.InitialOriginID),
			ev.GetStringIgnoreErrors(utils.OriginID),
			ev.GetStringIgnoreErrors(utils.OriginHost))
		if s == nil {
			if s, err = sS.initSession(args.CGREvent.Tenant,
				ev, sS.biJClntID(clnt),
				ev.GetStringIgnoreErrors(utils.OriginID),
				dbtItvl, args.ArgDispatcher, false, args.ForceDuration); err != nil {
				return err
			}
		}
		if rply.MaxUsage, err = sS.updateSession(s, ev.Clone(), false, args.ForceDuration); err != nil {
			return utils.NewErrRALs(err)
		}
	}
	return
}

// NewV1TerminateSessionArgs creates a new V1TerminateSessionArgs using the given arguments
func NewV1TerminateSessionArgs(acnts, resrc,
	thrds bool, thresholdIDs []string, stats bool,
	statIDs []string, cgrEv *utils.CGREvent,
	argDisp *utils.ArgDispatcher, forceDuration bool) (args *V1TerminateSessionArgs) {
	args = &V1TerminateSessionArgs{
		TerminateSession:  acnts,
		ReleaseResources:  resrc,
		ProcessThresholds: thrds,
		ProcessStats:      stats,
		CGREvent:          cgrEv,
		ArgDispatcher:     argDisp,
		ForceDuration:     forceDuration,
	}
	if len(thresholdIDs) != 0 {
		args.ThresholdIDs = thresholdIDs
	}
	if len(statIDs) != 0 {
		args.StatIDs = statIDs
	}
	return
}

// V1TerminateSessionArgs is used as argumen for TerminateSession
type V1TerminateSessionArgs struct {
	TerminateSession  bool
	ForceDuration     bool
	ReleaseResources  bool
	ProcessThresholds bool
	ProcessStats      bool
	ThresholdIDs      []string
	StatIDs           []string
	*utils.CGREvent
	*utils.ArgDispatcher
}

// ParseFlags will populate the V1TerminateSessionArgs flags
func (args *V1TerminateSessionArgs) ParseFlags(flags string) {
	dispatcherFlag := false
	for _, subsystem := range strings.Split(flags, utils.FIELDS_SEP) {
		switch {
		case subsystem == utils.MetaAccounts:
			args.TerminateSession = true
		case subsystem == utils.MetaResources:
			args.ReleaseResources = true
		case subsystem == utils.MetaDispatchers:
			dispatcherFlag = true
		case strings.Index(subsystem, utils.MetaThresholds) != -1:
			args.ProcessThresholds = true
			args.ThresholdIDs = getFlagIDs(subsystem)
		case strings.Index(subsystem, utils.MetaStats) != -1:
			args.ProcessStats = true
			args.StatIDs = getFlagIDs(subsystem)
		case subsystem == utils.MetaFD:
			args.ForceDuration = true
		}
	}
	cgrArgs := args.CGREvent.ExtractArgs(dispatcherFlag, false)
	args.ArgDispatcher = cgrArgs.ArgDispatcher
}

// BiRPCv1TerminateSession will stop debit loops as well as release any used resources
func (sS *SessionS) BiRPCv1TerminateSession(clnt rpcclient.ClientConnector,
	args *V1TerminateSessionArgs, rply *string) (err error) {
	if args.CGREvent == nil {
		return utils.NewErrMandatoryIeMissing(utils.CGREventString)
	}
	var withErrors bool
	if args.CGREvent.ID == "" {
		args.CGREvent.ID = utils.GenUUID()
	}
	// RPC caching
	if sS.cgrCfg.CacheCfg()[utils.CacheRPCResponses].Limit != 0 {
		cacheKey := utils.ConcatenatedKey(utils.SessionSv1TerminateSession, args.CGREvent.ID)
		refID := guardian.Guardian.GuardIDs("",
			sS.cgrCfg.GeneralCfg().LockingTimeout, cacheKey) // RPC caching needs to be atomic
		defer guardian.Guardian.UnguardIDs(refID)

		if itm, has := engine.Cache.Get(utils.CacheRPCResponses, cacheKey); has {
			cachedResp := itm.(*utils.CachedRPCResponse)
			if cachedResp.Error == nil {
				*rply = *cachedResp.Result.(*string)
			}
			return cachedResp.Error
		}
		defer engine.Cache.Set(utils.CacheRPCResponses, cacheKey,
			&utils.CachedRPCResponse{Result: rply, Error: err},
			nil, true, utils.NonTransactional)
	}
	// end of RPC caching
	if !args.TerminateSession && !args.ReleaseResources {
		return utils.NewErrMandatoryIeMissing("subsystems")
	}
	if args.CGREvent.Tenant == "" {
		args.CGREvent.Tenant = sS.cgrCfg.GeneralCfg().DefaultTenant
	}
	ev := engine.MapEvent(args.CGREvent.Event)
	cgrID := GetSetCGRID(ev)
	originID := ev.GetStringIgnoreErrors(utils.OriginID)
	if args.TerminateSession {
		if originID == "" {
			return utils.NewErrMandatoryIeMissing(utils.OriginID)
		}
		dbtItvl := sS.cgrCfg.SessionSCfg().DebitInterval
		if ev.HasField(utils.CGRDebitInterval) { // dynamic DebitInterval via CGRDebitInterval
			if dbtItvl, err = ev.GetDuration(utils.CGRDebitInterval); err != nil {
				return utils.NewErrRALs(err)
			}
		}
		var s *Session
		fib := utils.Fib()
		var isMsg bool // one time charging, do not perform indexing and sTerminator
		for i := 0; i < sS.cgrCfg.SessionSCfg().TerminateAttempts; i++ {
			if s = sS.getRelocateSession(cgrID,
				ev.GetStringIgnoreErrors(utils.InitialOriginID),
				ev.GetStringIgnoreErrors(utils.OriginID),
				ev.GetStringIgnoreErrors(utils.OriginHost)); s != nil {
				break
			}
			if i+1 < sS.cgrCfg.SessionSCfg().TerminateAttempts { // not last iteration
				time.Sleep(time.Duration(fib()) * time.Millisecond)
				continue
			}
			isMsg = true
			if s, err = sS.initSession(args.CGREvent.Tenant,
				ev, sS.biJClntID(clnt),
				ev.GetStringIgnoreErrors(utils.OriginID), dbtItvl,
				args.ArgDispatcher, isMsg, args.ForceDuration); err != nil {
				return utils.NewErrRALs(err)
			}
			if _, err = sS.updateSession(s, ev, isMsg, args.ForceDuration); err != nil {
				return err
			}
			break
		}
		if !isMsg {
			s.UpdateSRuns(ev, sS.cgrCfg.SessionSCfg().AlterableFields)
		}
		if err = sS.terminateSession(s,
			ev.GetDurationPtrIgnoreErrors(utils.Usage),
			ev.GetDurationPtrIgnoreErrors(utils.LastUsed),
			ev.GetTimePtrIgnoreErrors(utils.AnswerTime, utils.EmptyString),
			isMsg); err != nil {
			return utils.NewErrRALs(err)
		}
	}
	if args.ReleaseResources {
		if len(sS.cgrCfg.SessionSCfg().ResSConns) == 0 {
			return utils.NewErrNotConnected(utils.ResourceS)
		}
		if originID == "" {
			return utils.NewErrMandatoryIeMissing(utils.OriginID)
		}
		var reply string
		argsRU := utils.ArgRSv1ResourceUsage{
			CGREvent:      args.CGREvent,
			UsageID:       originID, // same ID should be accepted by first group since the previous resource should be expired
			Units:         1,
			ArgDispatcher: args.ArgDispatcher,
		}
		if err = sS.connMgr.Call(sS.cgrCfg.SessionSCfg().ResSConns, nil, utils.ResourceSv1ReleaseResources,
			argsRU, &reply); err != nil {
			return utils.NewErrResourceS(err)
		}
	}
	if args.ProcessThresholds {
		_, err := sS.processThreshold(args.CGREvent, args.ArgDispatcher,
			args.ThresholdIDs)
		if err != nil &&
			err.Error() != utils.ErrNotFound.Error() {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> error: %s processing event %+v with ThresholdS.",
					utils.SessionS, err.Error(), args.CGREvent))
			withErrors = true
		}
	}
	if args.ProcessStats {
		_, err := sS.processStats(args.CGREvent, args.ArgDispatcher,
			args.StatIDs)
		if err != nil &&
			err.Error() != utils.ErrNotFound.Error() {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> error: %s processing event %+v with StatS.",
					utils.SessionS, err.Error(), args.CGREvent))
			withErrors = true
		}
	}
	if withErrors {
		err = utils.ErrPartiallyExecuted
	}
	*rply = utils.OK
	return
}

// BiRPCv1ProcessCDR sends the CDR to CDRs
func (sS *SessionS) BiRPCv1ProcessCDR(clnt rpcclient.ClientConnector,
	cgrEvWithArgDisp *utils.CGREventWithArgDispatcher, rply *string) (err error) {
	if cgrEvWithArgDisp.ID == "" {
		cgrEvWithArgDisp.ID = utils.GenUUID()
	}
	if cgrEvWithArgDisp.Tenant == utils.EmptyString {
		cgrEvWithArgDisp.Tenant = sS.cgrCfg.GeneralCfg().DefaultTenant
	}

	// RPC caching
	if sS.cgrCfg.CacheCfg()[utils.CacheRPCResponses].Limit != 0 {
		cacheKey := utils.ConcatenatedKey(utils.SessionSv1ProcessCDR, cgrEvWithArgDisp.ID)
		refID := guardian.Guardian.GuardIDs("",
			sS.cgrCfg.GeneralCfg().LockingTimeout, cacheKey) // RPC caching needs to be atomic
		defer guardian.Guardian.UnguardIDs(refID)

		if itm, has := engine.Cache.Get(utils.CacheRPCResponses, cacheKey); has {
			cachedResp := itm.(*utils.CachedRPCResponse)
			if cachedResp.Error == nil {
				*rply = *cachedResp.Result.(*string)
			}
			return cachedResp.Error
		}
		defer engine.Cache.Set(utils.CacheRPCResponses, cacheKey,
			&utils.CachedRPCResponse{Result: rply, Error: err},
			nil, true, utils.NonTransactional)
	}
	// end of RPC caching
	// in case that source don't exist add it
	if _, has := cgrEvWithArgDisp.Event[utils.Source]; !has {
		cgrEvWithArgDisp.Event[utils.Source] = utils.MetaSessionS
	}

	ev := engine.MapEvent(cgrEvWithArgDisp.Event)
	cgrID := GetSetCGRID(ev)
	s := sS.getRelocateSession(cgrID,
		ev.GetStringIgnoreErrors(utils.InitialOriginID),
		ev.GetStringIgnoreErrors(utils.OriginID),
		ev.GetStringIgnoreErrors(utils.OriginHost))
	if s != nil {
		utils.Logger.Warning(
			fmt.Sprintf("<%s> ProcessCDR called for active session with CGRID: <%s>",
				utils.SessionS, cgrID))
		s.Lock() // events update session panic
		defer s.Unlock()
	} else if sIface, has := engine.Cache.Get(utils.CacheClosedSessions, cgrID); has {
		// found in cache
		s = sIface.(*Session)
	} else { // no cached session, CDR will be handled by CDRs
		return sS.connMgr.Call(sS.cgrCfg.SessionSCfg().CDRsConns, nil, utils.CDRsV1ProcessEvent,
			&engine.ArgV1ProcessEvent{
				Flags:         []string{utils.MetaRALs},
				CGREvent:      *cgrEvWithArgDisp.CGREvent,
				ArgDispatcher: cgrEvWithArgDisp.ArgDispatcher}, rply)
	}

	// Use previously stored Session to generate CDRs
	s.updateSRuns(ev, sS.cgrCfg.SessionSCfg().AlterableFields)
	// create one CGREvent for each session run
	var cgrEvs []*utils.CGREvent
	if cgrEvs, err = s.asCGREvents(); err != nil {
		return utils.NewErrServerError(err)
	}
	var withErrors bool
	for _, cgrEv := range cgrEvs {
		argsProc := &engine.ArgV1ProcessEvent{
			Flags: []string{fmt.Sprintf("%s:false", utils.MetaChargers),
				fmt.Sprintf("%s:false", utils.MetaAttributes)},
			CGREvent:      *cgrEv,
			ArgDispatcher: cgrEvWithArgDisp.ArgDispatcher,
		}
		if mp := engine.MapEvent(cgrEv.Event); unratedReqs.HasField(mp.GetStringIgnoreErrors(utils.RequestType)) { // order additional rating for unrated request types
			argsProc.Flags = append(argsProc.Flags, fmt.Sprintf("%s:true", utils.MetaRALs))
		}
		if err = sS.connMgr.Call(sS.cgrCfg.SessionSCfg().CDRsConns, nil, utils.CDRsV1ProcessEvent,
			argsProc, rply); err != nil {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> error <%s> posting CDR with CGRID: <%s>",
					utils.SessionS, err.Error(), cgrID))
			withErrors = true
		}
	}
	if withErrors {
		err = utils.ErrPartiallyExecuted
	}
	return
}

// NewV1ProcessMessageArgs is a constructor for MessageArgs used by ProcessMessage
func NewV1ProcessMessageArgs(attrs bool, attributeIDs []string,
	thds bool, thresholdIDs []string, stats bool, statIDs []string, resrc, acnts,
	suppls, supplsIgnoreErrs, supplsEventCost bool,
	cgrEv *utils.CGREvent, argDisp *utils.ArgDispatcher,
	supplierPaginator utils.Paginator, forceDuration bool, supMaxCost string) (args *V1ProcessMessageArgs) {
	args = &V1ProcessMessageArgs{
		AllocateResources:     resrc,
		Debit:                 acnts,
		GetAttributes:         attrs,
		ProcessThresholds:     thds,
		ProcessStats:          stats,
		SuppliersIgnoreErrors: supplsIgnoreErrs,
		GetSuppliers:          suppls,
		CGREvent:              cgrEv,
		ArgDispatcher:         argDisp,
		ForceDuration:         forceDuration,
	}
	if supplsEventCost {
		args.SuppliersMaxCost = utils.MetaEventCost
	} else {
		args.SuppliersMaxCost = supMaxCost
	}
	args.Paginator = supplierPaginator
	if len(attributeIDs) != 0 {
		args.AttributeIDs = attributeIDs
	}
	if len(thresholdIDs) != 0 {
		args.ThresholdIDs = thresholdIDs
	}
	if len(statIDs) != 0 {
		args.StatIDs = statIDs
	}
	return
}

// V1ProcessMessageArgs are the options passed to ProcessMessage API
type V1ProcessMessageArgs struct {
	GetAttributes         bool
	AllocateResources     bool
	Debit                 bool
	ForceDuration         bool
	ProcessThresholds     bool
	ProcessStats          bool
	GetSuppliers          bool
	SuppliersMaxCost      string
	SuppliersIgnoreErrors bool
	AttributeIDs          []string
	ThresholdIDs          []string
	StatIDs               []string
	*utils.CGREvent
	utils.Paginator
	*utils.ArgDispatcher
}

// ParseFlags will populate the V1ProcessMessageArgs flags
func (args *V1ProcessMessageArgs) ParseFlags(flags string) {
	dispatcherFlag := false
	for _, subsystem := range strings.Split(flags, utils.FIELDS_SEP) {
		switch {
		case subsystem == utils.MetaAccounts:
			args.Debit = true
		case subsystem == utils.MetaResources:
			args.AllocateResources = true
		case subsystem == utils.MetaDispatchers:
			dispatcherFlag = true
		case subsystem == utils.MetaSuppliers:
			args.GetSuppliers = true
		case subsystem == utils.MetaSuppliersIgnoreErrors:
			args.SuppliersIgnoreErrors = true
		case subsystem == utils.MetaSuppliersEventCost:
			args.SuppliersMaxCost = utils.MetaEventCost
		case strings.HasPrefix(subsystem, utils.MetaSuppliersMaxCost):
			args.SuppliersMaxCost = strings.TrimPrefix(subsystem, utils.MetaSuppliersMaxCost+utils.InInFieldSep)
		case strings.Index(subsystem, utils.MetaAttributes) != -1:
			args.GetAttributes = true
			args.AttributeIDs = getFlagIDs(subsystem)
		case strings.Index(subsystem, utils.MetaThresholds) != -1:
			args.ProcessThresholds = true
			args.ThresholdIDs = getFlagIDs(subsystem)
		case strings.Index(subsystem, utils.MetaStats) != -1:
			args.ProcessStats = true
			args.StatIDs = getFlagIDs(subsystem)
		case subsystem == utils.MetaFD:
			args.ForceDuration = true
		}
	}
	cgrArgs := args.CGREvent.ExtractArgs(dispatcherFlag, true)
	args.ArgDispatcher = cgrArgs.ArgDispatcher
	args.Paginator = *cgrArgs.SupplierPaginator
}

// V1ProcessMessageReply is the reply for the ProcessMessage API
type V1ProcessMessageReply struct {
	MaxUsage           time.Duration
	ResourceAllocation *string
	Attributes         *engine.AttrSProcessEventReply
	Suppliers          *engine.SortedSuppliers
	ThresholdIDs       *[]string
	StatQueueIDs       *[]string
	getMaxUsage        bool
}

// SetMaxUsageNeeded used by agent that use the reply as NavigableMapper
func (v1Rply *V1ProcessMessageReply) SetMaxUsageNeeded(getMaxUsage bool) {
	if v1Rply == nil {
		return
	}
	v1Rply.getMaxUsage = getMaxUsage
}

// AsNavigableMap is part of engine.NavigableMapper interface
func (v1Rply *V1ProcessMessageReply) AsNavigableMap() utils.NavigableMap2 {
	cgrReply := make(utils.NavigableMap2)
	if v1Rply.getMaxUsage {
		cgrReply[utils.CapMaxUsage] = utils.NewNMData(v1Rply.MaxUsage)
	}
	if v1Rply.ResourceAllocation != nil {
		cgrReply[utils.CapResourceAllocation] = utils.NewNMData(*v1Rply.ResourceAllocation)
	}
	if v1Rply.Attributes != nil {
		attrs := make(utils.NavigableMap2)
		for _, fldName := range v1Rply.Attributes.AlteredFields {
			fldName = strings.TrimPrefix(fldName, utils.MetaReq+utils.NestingSep)
			if v1Rply.Attributes.CGREvent.HasField(fldName) {
				attrs[fldName] = utils.NewNMData(v1Rply.Attributes.CGREvent.Event[fldName])
			}
		}
		cgrReply[utils.CapAttributes] = attrs
	}
	if v1Rply.Suppliers != nil {
		cgrReply[utils.CapSuppliers] = v1Rply.Suppliers.AsNavigableMap()
	}
	if v1Rply.ThresholdIDs != nil {
		thIDs := make(utils.NMSlice, len(*v1Rply.ThresholdIDs))
		for i, v := range *v1Rply.ThresholdIDs {
			thIDs[i] = utils.NewNMData(v)
		}
		cgrReply[utils.CapThresholds] = &thIDs
	}
	if v1Rply.StatQueueIDs != nil {
		stIDs := make(utils.NMSlice, len(*v1Rply.StatQueueIDs))
		for i, v := range *v1Rply.StatQueueIDs {
			stIDs[i] = utils.NewNMData(v)
		}
		cgrReply[utils.CapStatQueues] = &stIDs
	}
	return cgrReply
}

// BiRPCv1ProcessMessage processes one event with the right subsystems based on arguments received
func (sS *SessionS) BiRPCv1ProcessMessage(clnt rpcclient.ClientConnector,
	args *V1ProcessMessageArgs, rply *V1ProcessMessageReply) (err error) {
	if args.CGREvent == nil {
		return utils.NewErrMandatoryIeMissing(utils.CGREventString)
	}
	var withErrors bool
	if args.CGREvent.ID == "" {
		args.CGREvent.ID = utils.GenUUID()
	}

	// RPC caching
	if sS.cgrCfg.CacheCfg()[utils.CacheRPCResponses].Limit != 0 {
		cacheKey := utils.ConcatenatedKey(utils.SessionSv1ProcessMessage, args.CGREvent.ID)
		refID := guardian.Guardian.GuardIDs("",
			sS.cgrCfg.GeneralCfg().LockingTimeout, cacheKey) // RPC caching needs to be atomic
		defer guardian.Guardian.UnguardIDs(refID)

		if itm, has := engine.Cache.Get(utils.CacheRPCResponses, cacheKey); has {
			cachedResp := itm.(*utils.CachedRPCResponse)
			if cachedResp.Error == nil {
				*rply = *cachedResp.Result.(*V1ProcessMessageReply)
			}
			return cachedResp.Error
		}
		defer engine.Cache.Set(utils.CacheRPCResponses, cacheKey,
			&utils.CachedRPCResponse{Result: rply, Error: err},
			nil, true, utils.NonTransactional)
	}
	// end of RPC caching

	if args.CGREvent.Tenant == "" {
		args.CGREvent.Tenant = sS.cgrCfg.GeneralCfg().DefaultTenant
	}
	me := engine.MapEvent(args.CGREvent.Event)
	originID := me.GetStringIgnoreErrors(utils.OriginID)

	if args.GetAttributes {
		rplyAttr, err := sS.processAttributes(args.CGREvent, args.ArgDispatcher,
			args.AttributeIDs)
		if err == nil {
			args.CGREvent = rplyAttr.CGREvent
			rply.Attributes = &rplyAttr
		} else if err.Error() != utils.ErrNotFound.Error() {
			return utils.NewErrAttributeS(err)
		}
	}
	if args.AllocateResources {
		if len(sS.cgrCfg.SessionSCfg().ResSConns) == 0 {
			return utils.NewErrNotConnected(utils.ResourceS)
		}
		if originID == "" {
			return utils.NewErrMandatoryIeMissing(utils.OriginID)
		}
		attrRU := utils.ArgRSv1ResourceUsage{
			CGREvent:      args.CGREvent,
			UsageID:       originID,
			Units:         1,
			ArgDispatcher: args.ArgDispatcher,
		}
		var allocMessage string
		if err = sS.connMgr.Call(sS.cgrCfg.SessionSCfg().ResSConns, nil, utils.ResourceSv1AllocateResources,
			attrRU, &allocMessage); err != nil {
			return utils.NewErrResourceS(err)
		}
		rply.ResourceAllocation = &allocMessage
	}
	if args.GetSuppliers {
		splsReply, err := sS.getSuppliers(args.CGREvent.Clone(), args.ArgDispatcher,
			args.Paginator, args.SuppliersIgnoreErrors, args.SuppliersMaxCost)
		if err != nil {
			return err
		}
		if splsReply.SortedSuppliers != nil {
			rply.Suppliers = &splsReply
		}
	}
	if args.Debit {
		var maxUsage time.Duration
		if maxUsage, err = sS.chargeEvent(args.CGREvent.Tenant,
			engine.MapEvent(args.CGREvent.Event), args.ArgDispatcher, args.ForceDuration); err != nil {
			return err
		}
		rply.MaxUsage = maxUsage
	}
	if args.ProcessThresholds {
		tIDs, err := sS.processThreshold(args.CGREvent, args.ArgDispatcher,
			args.ThresholdIDs)
		if err != nil && err.Error() != utils.ErrNotFound.Error() {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> error: %s processing event %+v with ThresholdS.",
					utils.SessionS, err.Error(), args.CGREvent))
			withErrors = true
		}
		rply.ThresholdIDs = &tIDs
	}
	if args.ProcessStats {
		sIDs, err := sS.processStats(args.CGREvent, args.ArgDispatcher,
			args.StatIDs)
		if err != nil &&
			err.Error() != utils.ErrNotFound.Error() {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> error: %s processing event %+v with StatS.",
					utils.SessionS, err.Error(), args.CGREvent))
			withErrors = true
		}
		rply.StatQueueIDs = &sIDs
	}
	if withErrors {
		err = utils.ErrPartiallyExecuted
	}

	return
}

// V1ProcessEventArgs are the options passed to ProcessEvent API
type V1ProcessEventArgs struct {
	Flags []string
	*utils.CGREvent
	utils.Paginator
	*utils.ArgDispatcher
}

// V1ProcessEventReply is the reply for the ProcessEvent API
type V1ProcessEventReply struct {
	MaxUsage           time.Duration
	ResourceAllocation *string
	Attributes         *engine.AttrSProcessEventReply
	Suppliers          *engine.SortedSuppliers
	ThresholdIDs       *[]string
	StatQueueIDs       *[]string
	getMaxUsage        bool
}

// SetMaxUsageNeeded used by agent that use the reply as NavigableMapper
func (v1Rply *V1ProcessEventReply) SetMaxUsageNeeded(getMaxUsage bool) {
	if v1Rply == nil {
		return
	}
	v1Rply.getMaxUsage = getMaxUsage
}

// AsNavigableMap is part of engine.NavigableMapper interface
func (v1Rply *V1ProcessEventReply) AsNavigableMap() utils.NavigableMap2 {
	cgrReply := make(utils.NavigableMap2)
	if v1Rply.getMaxUsage {
		cgrReply[utils.CapMaxUsage] = utils.NewNMData(v1Rply.MaxUsage)
	}
	if v1Rply.ResourceAllocation != nil {
		cgrReply[utils.CapResourceAllocation] = utils.NewNMData(*v1Rply.ResourceAllocation)
	}
	if v1Rply.Attributes != nil {
		attrs := make(utils.NavigableMap2)
		for _, fldName := range v1Rply.Attributes.AlteredFields {
			fldName = strings.TrimPrefix(fldName, utils.MetaReq+utils.NestingSep)
			if v1Rply.Attributes.CGREvent.HasField(fldName) {
				attrs[fldName] = utils.NewNMData(v1Rply.Attributes.CGREvent.Event[fldName])
			}
		}
		cgrReply[utils.CapAttributes] = attrs
	}
	if v1Rply.Suppliers != nil {
		cgrReply[utils.CapSuppliers] = v1Rply.Suppliers.AsNavigableMap()
	}
	if v1Rply.ThresholdIDs != nil {
		thIDs := make(utils.NMSlice, len(*v1Rply.ThresholdIDs))
		for i, v := range *v1Rply.ThresholdIDs {
			thIDs[i] = utils.NewNMData(v)
		}
		cgrReply[utils.CapThresholds] = &thIDs
	}
	if v1Rply.StatQueueIDs != nil {
		stIDs := make(utils.NMSlice, len(*v1Rply.StatQueueIDs))
		for i, v := range *v1Rply.StatQueueIDs {
			stIDs[i] = utils.NewNMData(v)
		}
		cgrReply[utils.CapStatQueues] = &stIDs
	}
	return cgrReply
}

// BiRPCv1ProcessEvent processes one event with the right subsystems based on arguments received
func (sS *SessionS) BiRPCv1ProcessEvent(clnt rpcclient.ClientConnector,
	args *V1ProcessEventArgs, rply *V1ProcessEventReply) (err error) {
	if args.CGREvent == nil {
		return utils.NewErrMandatoryIeMissing(utils.CGREventString)
	}
	var withErrors bool
	if args.CGREvent.ID == "" {
		args.CGREvent.ID = utils.GenUUID()
	}

	// RPC caching
	if sS.cgrCfg.CacheCfg()[utils.CacheRPCResponses].Limit != 0 {
		cacheKey := utils.ConcatenatedKey(utils.SessionSv1ProcessEvent, args.CGREvent.ID)
		refID := guardian.Guardian.GuardIDs("",
			sS.cgrCfg.GeneralCfg().LockingTimeout, cacheKey) // RPC caching needs to be atomic
		defer guardian.Guardian.UnguardIDs(refID)

		if itm, has := engine.Cache.Get(utils.CacheRPCResponses, cacheKey); has {
			cachedResp := itm.(*utils.CachedRPCResponse)
			if cachedResp.Error == nil {
				*rply = *cachedResp.Result.(*V1ProcessEventReply)
			}
			return cachedResp.Error
		}
		defer engine.Cache.Set(utils.CacheRPCResponses, cacheKey,
			&utils.CachedRPCResponse{Result: rply, Error: err},
			nil, true, utils.NonTransactional)
	}
	// end of RPC caching

	if args.CGREvent.Tenant == "" {
		args.CGREvent.Tenant = sS.cgrCfg.GeneralCfg().DefaultTenant
	}
	ev := engine.MapEvent(args.CGREvent.Event)
	originID := ev.GetStringIgnoreErrors(utils.OriginID)
	dbtItvl := sS.cgrCfg.SessionSCfg().DebitInterval

	//convert from Flags []string to utils.FlagsWithParams
	var argsFlagsWithParams utils.FlagsWithParams
	if argsFlagsWithParams, err = utils.FlagsWithParamsFromSlice(args.Flags); err != nil {
		return
	}
	// check for *attribute
	if argsFlagsWithParams.HasKey(utils.MetaAttributes) {
		rplyAttr, err := sS.processAttributes(args.CGREvent, args.ArgDispatcher,
			argsFlagsWithParams.ParamsSlice(utils.MetaAttributes))
		if err == nil {
			args.CGREvent = rplyAttr.CGREvent
			rply.Attributes = &rplyAttr
		} else if err.Error() != utils.ErrNotFound.Error() {
			return utils.NewErrAttributeS(err)
		}
	}
	// check for *resources
	if argsFlagsWithParams.HasKey(utils.MetaResources) {
		if len(sS.cgrCfg.SessionSCfg().ResSConns) == 0 {
			return utils.NewErrNotConnected(utils.ResourceS)
		}
		if originID == "" {
			return utils.NewErrMandatoryIeMissing(utils.OriginID)
		}
		attrRU := utils.ArgRSv1ResourceUsage{
			CGREvent:      args.CGREvent,
			UsageID:       originID,
			Units:         1,
			ArgDispatcher: args.ArgDispatcher,
		}
		var resMessage string
		// check what we need to do for resources (*authorization/*allocation)
		if resOpt := argsFlagsWithParams.ParamsSlice(utils.MetaResources); len(resOpt) != 0 {
			//check for subflags and convert them into utils.FlagsWithParams
			resourceFlagsWithParams, err := utils.FlagsWithParamsFromSlice(resOpt)
			if err != nil {
				return err
			}
			if resourceFlagsWithParams.HasKey(utils.MetaAuthorize) {
				if err = sS.connMgr.Call(sS.cgrCfg.SessionSCfg().ResSConns, nil, utils.ResourceSv1AuthorizeResources,
					attrRU, &resMessage); err != nil {
					return utils.NewErrResourceS(err)
				}
				rply.ResourceAllocation = &resMessage
			}
			if resourceFlagsWithParams.HasKey(utils.MetaAllocate) {
				if err = sS.connMgr.Call(sS.cgrCfg.SessionSCfg().ResSConns, nil, utils.ResourceSv1AllocateResources,
					attrRU, &resMessage); err != nil {
					return utils.NewErrResourceS(err)
				}
				rply.ResourceAllocation = &resMessage
			}
			if resourceFlagsWithParams.HasKey(utils.MetaRelease) {
				if err = sS.connMgr.Call(sS.cgrCfg.SessionSCfg().ResSConns, nil, utils.ResourceSv1ReleaseResources,
					attrRU, &resMessage); err != nil {
					return utils.NewErrResourceS(err)
				}
				rply.ResourceAllocation = &resMessage
			}
		}
	}
	// check what we need to do for RALs (*auth/*init/*update/*terminate)
	if argsFlagsWithParams.HasKey(utils.MetaRALs) {
		if ralsOpts := argsFlagsWithParams.ParamsSlice(utils.MetaRALs); len(ralsOpts) != 0 {
			//check for subflags and convert them into utils.FlagsWithParams
			ralsFlagsWithParams, err := utils.FlagsWithParamsFromSlice(ralsOpts)
			//for the moment only the the flag will be executed
			switch {
			//check for auth session
			case ralsFlagsWithParams.HasKey(utils.MetaAuthorize):
				maxUsage, err := sS.authEvent(args.CGREvent.Tenant,
					engine.MapEvent(args.CGREvent.Event), ralsFlagsWithParams.HasKey(utils.MetaFD))
				if err != nil {
					return err
				}
				rply.MaxUsage = maxUsage
			// check for init session
			case ralsFlagsWithParams.HasKey(utils.MetaInitiate):
				if ev.HasField(utils.CGRDebitInterval) { // dynamic DebitInterval via CGRDebitInterval
					if dbtItvl, err = ev.GetDuration(utils.CGRDebitInterval); err != nil {
						return err
					}
				}
				s, err := sS.initSession(args.CGREvent.Tenant, ev,
					sS.biJClntID(clnt), originID, dbtItvl, args.ArgDispatcher, false, ralsFlagsWithParams.HasKey(utils.MetaFD))
				if err != nil {
					return err
				}
				s.RLock()
				isPrepaid := s.debitStop != nil
				s.RUnlock()
				if isPrepaid { //active debit
					rply.MaxUsage = sS.cgrCfg.SessionSCfg().GetDefaultUsage(ev.GetStringIgnoreErrors(utils.ToR))
				} else {
					var maxUsage time.Duration
					if maxUsage, err = sS.updateSession(s, nil, false, ralsFlagsWithParams.HasKey(utils.MetaFD)); err != nil {
						return utils.NewErrRALs(err)
					}
					rply.MaxUsage = maxUsage
				}
			//check for update session
			case ralsFlagsWithParams.HasKey(utils.MetaUpdate):
				if ev.HasField(utils.CGRDebitInterval) { // dynamic DebitInterval via CGRDebitInterval
					if dbtItvl, err = ev.GetDuration(utils.CGRDebitInterval); err != nil {
						return utils.NewErrRALs(err)
					}
				}
				ev := engine.MapEvent(args.CGREvent.Event)
				cgrID := GetSetCGRID(ev)
				s := sS.getRelocateSession(cgrID,
					ev.GetStringIgnoreErrors(utils.InitialOriginID),
					ev.GetStringIgnoreErrors(utils.OriginID),
					ev.GetStringIgnoreErrors(utils.OriginHost))
				if s == nil {
					if s, err = sS.initSession(args.CGREvent.Tenant,
						ev, sS.biJClntID(clnt),
						ev.GetStringIgnoreErrors(utils.OriginID), dbtItvl, args.ArgDispatcher,
						false, ralsFlagsWithParams.HasKey(utils.MetaFD)); err != nil {
						return err
					}
				}
				var maxUsage time.Duration
				if maxUsage, err = sS.updateSession(s, ev, false, ralsFlagsWithParams.HasKey(utils.MetaFD)); err != nil {
					return utils.NewErrRALs(err)
				}
				rply.MaxUsage = maxUsage
			// check for terminate session
			case ralsFlagsWithParams.HasKey(utils.MetaTerminate):
				if ev.HasField(utils.CGRDebitInterval) { // dynamic DebitInterval via CGRDebitInterval
					if dbtItvl, err = ev.GetDuration(utils.CGRDebitInterval); err != nil {
						return utils.NewErrRALs(err)
					}
				}
				cgrID := GetSetCGRID(ev)
				s := sS.getRelocateSession(cgrID,
					ev.GetStringIgnoreErrors(utils.InitialOriginID),
					ev.GetStringIgnoreErrors(utils.OriginID),
					ev.GetStringIgnoreErrors(utils.OriginHost))
				if s == nil {
					if s, err = sS.initSession(args.CGREvent.Tenant,
						ev, sS.biJClntID(clnt),
						ev.GetStringIgnoreErrors(utils.OriginID), dbtItvl,
						args.ArgDispatcher, false, ralsFlagsWithParams.HasKey(utils.MetaFD)); err != nil {
						return err
					}
				}
				if err = sS.terminateSession(s,
					ev.GetDurationPtrIgnoreErrors(utils.Usage),
					ev.GetDurationPtrIgnoreErrors(utils.LastUsed),
					ev.GetTimePtrIgnoreErrors(utils.AnswerTime, utils.EmptyString),
					false); err != nil {
					return utils.NewErrRALs(err)
				}
			}
		}
	}
	// get suppliers if required
	if argsFlagsWithParams.HasKey(utils.MetaSuppliers) {
		var ignoreErrors bool
		// check in case we have options for suppliers
		maxCost := argsFlagsWithParams.ParamValue(utils.MetaSuppliersMaxCost)
		if splOpts := argsFlagsWithParams.ParamsSlice(utils.MetaSuppliers); len(splOpts) != 0 {
			//check for subflags and convert them into utils.FlagsWithParams
			splsFlagsWithParams, err := utils.FlagsWithParamsFromSlice(splOpts)
			if err != nil {
				return err
			}
			if splsFlagsWithParams.HasKey(utils.MetaIgnoreErrors) {
				ignoreErrors = true
			}
			if splsFlagsWithParams.HasKey(utils.MetaEventCost) {
				maxCost = utils.MetaEventCost
			}
		}
		splsReply, err := sS.getSuppliers(args.CGREvent.Clone(), args.ArgDispatcher,
			args.Paginator, ignoreErrors, maxCost)
		if err != nil {
			return err
		}
		if splsReply.SortedSuppliers != nil {
			rply.Suppliers = &splsReply
		}
	}
	// process thresholds if required
	if argsFlagsWithParams.HasKey(utils.MetaThresholds) {
		tIDs, err := sS.processThreshold(args.CGREvent, args.ArgDispatcher,
			argsFlagsWithParams.ParamsSlice(utils.MetaThresholds))
		if err != nil && err.Error() != utils.ErrNotFound.Error() {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> error: %s processing event %+v with ThresholdS.",
					utils.SessionS, err.Error(), args.CGREvent))
			withErrors = true
		}
		rply.ThresholdIDs = &tIDs
	}
	// process stats if required
	if argsFlagsWithParams.HasKey(utils.MetaStats) {
		sIDs, err := sS.processStats(args.CGREvent, args.ArgDispatcher,
			argsFlagsWithParams.ParamsSlice(utils.MetaStats))
		if err != nil &&
			err.Error() != utils.ErrNotFound.Error() {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> error: %s processing event %+v with StatS.",
					utils.SessionS, err.Error(), args.CGREvent))
			withErrors = true
		}
		rply.StatQueueIDs = &sIDs
	}
	if argsFlagsWithParams.HasKey(utils.MetaCDRs) {
		var rplyCDR string
		cgrID := GetSetCGRID(ev)
		s := sS.getRelocateSession(cgrID,
			ev.GetStringIgnoreErrors(utils.InitialOriginID),
			ev.GetStringIgnoreErrors(utils.OriginID),
			ev.GetStringIgnoreErrors(utils.OriginHost))
		if s != nil {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> ProcessCDR called for active session with CGRID: <%s>",
					utils.SessionS, cgrID))
			s.Lock() // events update session panic
			defer s.Unlock()
		} else if sIface, has := engine.Cache.Get(utils.CacheClosedSessions, cgrID); has {
			// found in cache
			s = sIface.(*Session)
		} else { // no cached session, CDR will be handled by CDRs
			return sS.connMgr.Call(sS.cgrCfg.SessionSCfg().CDRsConns, nil, utils.CDRsV1ProcessEvent,
				&engine.ArgV1ProcessEvent{
					Flags:         []string{utils.MetaRALs},
					CGREvent:      *args.CGREvent,
					ArgDispatcher: args.ArgDispatcher}, &rplyCDR)
		}

		// Use previously stored Session to generate CDRs
		s.updateSRuns(ev, sS.cgrCfg.SessionSCfg().AlterableFields)
		// create one CGREvent for each session run
		var cgrEvs []*utils.CGREvent
		if cgrEvs, err = s.asCGREvents(); err != nil {
			return utils.NewErrServerError(err)
		}
		var withErrors bool
		for _, cgrEv := range cgrEvs {
			argsProc := &engine.ArgV1ProcessEvent{
				Flags: []string{utils.MetaChargers + ":false",
					utils.MetaAttributes + ":false"},
				CGREvent:      *cgrEv,
				ArgDispatcher: args.ArgDispatcher,
			}
			if mp := engine.MapEvent(cgrEv.Event); unratedReqs.HasField(mp.GetStringIgnoreErrors(utils.RequestType)) { // order additional rating for unrated request types
				argsProc.Flags = append(argsProc.Flags, fmt.Sprintf("%s:true", utils.MetaRALs))
			}
			if err = sS.connMgr.Call(sS.cgrCfg.SessionSCfg().CDRsConns, nil, utils.CDRsV1ProcessEvent,
				argsProc, &rplyCDR); err != nil {
				utils.Logger.Warning(
					fmt.Sprintf("<%s> error <%s> posting CDR with CGRID: <%s>",
						utils.SessionS, err.Error(), cgrID))
				withErrors = true
			}
		}
		if withErrors {
			err = utils.ErrPartiallyExecuted
		}
	}
	if withErrors {
		err = utils.ErrPartiallyExecuted
	}
	return
}

// BiRPCv1SyncSessions will sync sessions on demand
func (sS *SessionS) BiRPCv1SyncSessions(clnt rpcclient.ClientConnector,
	ignParam string, reply *string) error {
	sS.syncSessions()
	*reply = utils.OK
	return nil
}

// BiRPCv1ForceDisconnect will force disconnecting sessions matching sessions
func (sS *SessionS) BiRPCv1ForceDisconnect(clnt rpcclient.ClientConnector,
	args *utils.SessionFilter, reply *string) (err error) {
	if args == nil { //protection in case on nil
		args = &utils.SessionFilter{}
	}
	if len(args.Filters) != 0 && sS.dm == nil {
		return utils.ErrNoDatabaseConn
	}
	aSs := sS.filterSessions(args, false)
	if len(aSs) == 0 {
		return utils.ErrNotFound
	}
	for _, as := range aSs {
		ss := sS.getSessions(as.CGRID, false)
		if len(ss) == 0 {
			continue
		}
		ss[0].Lock() // protect forceSTerminate as it is not thread safe for the session
		if errTerm := sS.forceSTerminate(ss[0], 0, nil, nil); errTerm != nil {
			utils.Logger.Warning(
				fmt.Sprintf(
					"<%s> failed force-terminating session with id: <%s>, err: <%s>",
					utils.SessionS, ss[0].cgrID(), errTerm.Error()))
			err = utils.ErrPartiallyExecuted
		}
		ss[0].Unlock()
	}
	if err == nil {
		*reply = utils.OK
	} else {
		*reply = err.Error()
	}
	return nil
}

// BiRPCv1RegisterInternalBiJSONConn will register the client for a bidirectional comunication
func (sS *SessionS) BiRPCv1RegisterInternalBiJSONConn(clnt rpcclient.ClientConnector,
	ign string, reply *string) error {
	sS.RegisterIntBiJConn(clnt)
	*reply = utils.OK
	return nil
}

// BiRPCv1ActivateSessions is called to activate a list/all sessions
// returns utils.ErrPartiallyExecuted in case of errors
func (sS *SessionS) BiRPCv1ActivateSessions(clnt rpcclient.ClientConnector,
	sIDs []string, reply *string) (err error) {
	if len(sIDs) == 0 {
		sS.pSsMux.RLock()
		i := 0
		sIDs = make([]string, len(sS.pSessions))
		for sID := range sS.pSessions {
			sIDs[i] = sID
			i++
		}
		sS.pSsMux.RUnlock()
	}
	for _, sID := range sIDs {
		if s := sS.transitSState(sID, false); s == nil {
			utils.Logger.Warning(fmt.Sprintf("<%s> no passive session with id: <%s>", utils.SessionS, sID))
			err = utils.ErrPartiallyExecuted
		}
	}
	if err == nil {
		*reply = utils.OK
	}
	return
}

// BiRPCv1DeactivateSessions is called to deactivate a list/all active sessios
// returns utils.ErrPartiallyExecuted in case of errors
func (sS *SessionS) BiRPCv1DeactivateSessions(clnt rpcclient.ClientConnector,
	sIDs []string, reply *string) (err error) {
	if len(sIDs) == 0 {
		sS.aSsMux.RLock()
		i := 0
		sIDs = make([]string, len(sS.aSessions))
		for sID := range sS.aSessions {
			sIDs[i] = sID
			i++
		}
		sS.aSsMux.RUnlock()
	}
	for _, sID := range sIDs {
		if s := sS.transitSState(sID, true); s == nil {
			utils.Logger.Warning(fmt.Sprintf("<%s> no active session with id: <%s>", utils.SessionS, sID))
			err = utils.ErrPartiallyExecuted
		}
	}
	if err == nil {
		*reply = utils.OK
	}
	return
}

// processThreshold will receive the event and send it to ThresholdS to be processed
func (sS *SessionS) processThreshold(cgrEv *utils.CGREvent, argDisp *utils.ArgDispatcher, thIDs []string) (tIDs []string, err error) {
	if len(sS.cgrCfg.SessionSCfg().ThreshSConns) == 0 {
		return tIDs, utils.NewErrNotConnected(utils.ThresholdS)
	}
	thEv := &engine.ArgsProcessEvent{
		CGREvent:      cgrEv,
		ArgDispatcher: argDisp,
	}
	// check if we have thresholdIDs
	if len(thIDs) != 0 {
		thEv.ThresholdIDs = thIDs
	}
	//initialize the returned variable
	tIDs = make([]string, 0)
	err = sS.connMgr.Call(sS.cgrCfg.SessionSCfg().ThreshSConns, nil, utils.ThresholdSv1ProcessEvent, thEv, &tIDs)
	return
}

// processStats will receive the event and send it to StatS to be processed
func (sS *SessionS) processStats(cgrEv *utils.CGREvent, argDisp *utils.ArgDispatcher, stsIDs []string) (sIDs []string, err error) {
	if len(sS.cgrCfg.SessionSCfg().StatSConns) == 0 {
		return sIDs, utils.NewErrNotConnected(utils.StatS)
	}

	statArgs := &engine.StatsArgsProcessEvent{
		CGREvent:      cgrEv,
		ArgDispatcher: argDisp,
	}
	// check in case we have StatIDs inside flags
	if len(stsIDs) != 0 {
		statArgs.StatIDs = stsIDs
	}
	//initialize the returned variable
	sIDs = make([]string, 0)
	err = sS.connMgr.Call(sS.cgrCfg.SessionSCfg().StatSConns, nil, utils.StatSv1ProcessEvent, statArgs, &sIDs)
	return
}

// getSuppliers will receive the event and send it to SupplierS to find the suppliers
func (sS *SessionS) getSuppliers(cgrEv *utils.CGREvent, argDisp *utils.ArgDispatcher, pag utils.Paginator,
	ignoreErrors bool, maxCost string) (splsReply engine.SortedSuppliers, err error) {
	if len(sS.cgrCfg.SessionSCfg().SupplSConns) == 0 {
		return splsReply, utils.NewErrNotConnected(utils.SupplierS)
	}
	if acd, has := cgrEv.Event[utils.ACD]; has {
		cgrEv.Event[utils.Usage] = acd
	}
	sArgs := &engine.ArgsGetSuppliers{
		CGREvent:      cgrEv,
		Paginator:     pag,
		ArgDispatcher: argDisp,
		IgnoreErrors:  ignoreErrors,
		MaxCost:       maxCost,
	}
	if err = sS.connMgr.Call(sS.cgrCfg.SessionSCfg().SupplSConns, nil, utils.SupplierSv1GetSuppliers,
		sArgs, &splsReply); err != nil {
		return splsReply, utils.NewErrSupplierS(err)
	}
	return
}

// processAttributes will receive the event and send it to AttributeS to be processed
func (sS *SessionS) processAttributes(cgrEv *utils.CGREvent, argDisp *utils.ArgDispatcher,
	attrIDs []string) (rplyEv engine.AttrSProcessEventReply, err error) {
	if len(sS.cgrCfg.SessionSCfg().AttrSConns) == 0 {
		return rplyEv, utils.NewErrNotConnected(utils.AttributeS)
	}
	attrArgs := &engine.AttrArgsProcessEvent{
		Context: utils.StringPointer(utils.FirstNonEmpty(
			utils.IfaceAsString(cgrEv.Event[utils.Context]),
			utils.MetaSessionS)),
		CGREvent:      cgrEv,
		ArgDispatcher: argDisp,
	}
	if len(attrIDs) != 0 {
		attrArgs.AttributeIDs = attrIDs
	}
	if err = sS.connMgr.Call(sS.cgrCfg.SessionSCfg().AttrSConns, nil, utils.AttributeSv1ProcessEvent,
		attrArgs, &rplyEv); err != nil {
		return
	}
	return
}

// BiRPCV1GetMaxUsage returns the maximum usage as seconds, compatible with OpenSIPS 2.3
// DEPRECATED, it will be removed in future versions
func (sS *SessionS) BiRPCV1GetMaxUsage(clnt rpcclient.ClientConnector,
	ev engine.MapEvent, maxUsage *float64) (err error) {
	var rply *V1AuthorizeReply
	if err = sS.BiRPCv1AuthorizeEvent(
		clnt,
		&V1AuthorizeArgs{
			GetMaxUsage: true,
			CGREvent: &utils.CGREvent{
				Tenant: utils.FirstNonEmpty(
					ev.GetStringIgnoreErrors(utils.Tenant),
					sS.cgrCfg.GeneralCfg().DefaultTenant),
				ID:    utils.UUIDSha1Prefix(),
				Event: ev}},
		rply); err != nil {
		return
	}
	*maxUsage = rply.MaxUsage.Seconds()
	return nil
}

// BiRPCV1InitiateSession is called on session start, returns the maximum number of seconds the session can last
// DEPRECATED, it will be removed in future versions
// Kept for compatibility with OpenSIPS 2.3
func (sS *SessionS) BiRPCV1InitiateSession(clnt rpcclient.ClientConnector,
	ev engine.MapEvent, maxUsage *float64) (err error) {
	var rply *V1InitSessionReply
	if err = sS.BiRPCv1InitiateSession(
		clnt,
		&V1InitSessionArgs{
			InitSession: true,
			CGREvent: &utils.CGREvent{
				Tenant: utils.FirstNonEmpty(
					ev.GetStringIgnoreErrors(utils.Tenant),
					sS.cgrCfg.GeneralCfg().DefaultTenant),
				ID:    utils.UUIDSha1Prefix(),
				Event: ev}},
		rply); err != nil {
		return
	}
	*maxUsage = rply.MaxUsage.Seconds()
	return
}

// BiRPCV1UpdateSession processes interim updates, returns remaining duration from the RALs
// DEPRECATED, it will be removed in future versions
// Kept for compatibility with OpenSIPS 2.3
func (sS *SessionS) BiRPCV1UpdateSession(clnt rpcclient.ClientConnector,
	ev engine.MapEvent, maxUsage *float64) (err error) {
	var rply *V1UpdateSessionReply
	if err = sS.BiRPCv1UpdateSession(
		clnt,
		&V1UpdateSessionArgs{
			UpdateSession: true,
			CGREvent: &utils.CGREvent{
				Tenant: utils.FirstNonEmpty(
					ev.GetStringIgnoreErrors(utils.Tenant),
					sS.cgrCfg.GeneralCfg().DefaultTenant),
				ID:    utils.UUIDSha1Prefix(),
				Event: ev}},
		rply); err != nil {
		return
	}
	*maxUsage = rply.MaxUsage.Seconds()
	return
}

// BiRPCV1TerminateSession is called on session end, should stop debit loop
// DEPRECATED, it will be removed in future versions
// Kept for compatibility with OpenSIPS 2.3
func (sS *SessionS) BiRPCV1TerminateSession(clnt rpcclient.ClientConnector,
	ev engine.MapEvent, rply *string) (err error) {
	return sS.BiRPCv1TerminateSession(
		clnt,
		&V1TerminateSessionArgs{
			TerminateSession: true,
			CGREvent: &utils.CGREvent{
				Tenant: utils.FirstNonEmpty(
					ev.GetStringIgnoreErrors(utils.Tenant),
					sS.cgrCfg.GeneralCfg().DefaultTenant),
				ID:    utils.UUIDSha1Prefix(),
				Event: ev}},
		rply)
}

// BiRPCV1ProcessCDR should send the CDR to CDRS
// DEPRECATED, it will be removed in future versions
// Kept for compatibility with OpenSIPS 2.3
func (sS *SessionS) BiRPCV1ProcessCDR(clnt rpcclient.ClientConnector,
	ev engine.MapEvent, rply *string) (err error) {
	return sS.BiRPCv1ProcessCDR(
		clnt,
		&utils.CGREventWithArgDispatcher{
			CGREvent: &utils.CGREvent{
				Tenant: utils.FirstNonEmpty(
					ev.GetStringIgnoreErrors(utils.Tenant),
					sS.cgrCfg.GeneralCfg().DefaultTenant),
				ID:    utils.UUIDSha1Prefix(),
				Event: ev}},
		rply)
}
