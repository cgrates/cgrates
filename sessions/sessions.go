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
	cgrCfg.SessionSCfg().SessionIndexes.Add(utils.OriginID) // Make sure we have indexing for OriginID since it is a requirement on prefix searching

	return &SessionS{
		cgrCfg:        cgrCfg,
		dm:            dm,
		connMgr:       connMgr,
		biJClnts:      make(map[rpcclient.ClientConnector]string),
		biJIDs:        make(map[string]*biJClient),
		aSessions:     make(map[string]*Session),
		aSessionsIdx:  make(map[string]map[string]map[string]utils.StringSet),
		aSessionsRIdx: make(map[string][]*riFieldNameVal),
		pSessions:     make(map[string]*Session),
		pSessionsIdx:  make(map[string]map[string]map[string]utils.StringSet),
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
	aSessionsIdx  map[string]map[string]map[string]utils.StringSet // map[fieldName]map[fieldValue][cgrID]utils.StringSet[runID]sID
	aSessionsRIdx map[string][]*riFieldNameVal                     // reverse indexes for active sessions, used on remove

	pSsMux    sync.RWMutex        // protects pSessions
	pSessions map[string]*Session // group passive sessions based on cgrID

	pSIMux        sync.RWMutex                                     // protects pSessionsIdx
	pSessionsIdx  map[string]map[string]map[string]utils.StringSet // map[fieldName]map[fieldValue][cgrID]utils.StringSet[runID]sID
	pSessionsRIdx map[string][]*riFieldNameVal                     // reverse indexes for passive sessions, used on remove
}

// ListenAndServe starts the service and binds it to the listen loop
func (sS *SessionS) ListenAndServe(stopChan chan struct{}) {
	utils.Logger.Info(fmt.Sprintf("<%s> starting <%s> subsystem", utils.CoreS, utils.SessionS))
	if sS.cgrCfg.SessionSCfg().ChannelSyncInterval != 0 {
		for { // Schedule sync channels to run repeately
			select {
			case <-stopChan:
				return
			case <-time.After(sS.cgrCfg.SessionSCfg().ChannelSyncInterval):
				sS.syncSessions()
			}
		}
	}
	return
}

// Shutdown is called by engine to clear states
func (sS *SessionS) Shutdown() (err error) {
	var hasErr bool
	for _, s := range sS.getSessions("", false) { // Force sessions shutdown
		if err = sS.terminateSession(s, nil, nil, nil, false); err != nil {
			hasErr = true
		}
	}
	if hasErr {
		return utils.ErrPartiallyExecuted
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
func (sS *SessionS) setSTerminator(s *Session, opts engine.MapEvent) {
	var err error
	// TTL
	var ttl time.Duration
	if opts.HasField(utils.OptsSessionTTL) {
		ttl, err = opts.GetDuration(utils.OptsSessionTTL)
	} else if s.OptsStart.HasField(utils.OptsSessionTTL) {
		ttl, err = s.OptsStart.GetDuration(utils.OptsSessionTTL)
	} else {
		ttl = sS.cgrCfg.SessionSCfg().SessionTTL
	}
	if err != nil {
		utils.Logger.Warning(
			fmt.Sprintf("<%s>, cannot extract <%s> for session:<%s>, from it's options: <%s>, err: <%s>",
				utils.SessionS, utils.OptsSessionTTL, s.CGRID, opts, err))
		return
	}
	if ttl == 0 {
		return // nothing to set up
	}
	// random delay computation
	var maxDelay time.Duration
	if opts.HasField(utils.OptsSessionTTLMaxDelay) {
		maxDelay, err = opts.GetDuration(utils.OptsSessionTTLMaxDelay)
	} else if s.OptsStart.HasField(utils.OptsSessionTTLMaxDelay) {
		maxDelay, err = s.OptsStart.GetDuration(utils.OptsSessionTTLMaxDelay)
	} else if sS.cgrCfg.SessionSCfg().SessionTTLMaxDelay != nil {
		maxDelay = *sS.cgrCfg.SessionSCfg().SessionTTLMaxDelay
	}
	if err != nil {
		utils.Logger.Warning(
			fmt.Sprintf("<%s>, cannot extract <%s> for session:<%s>, from it's options: <%s>, err: <%s>",
				utils.SessionS, utils.OptsSessionTTLMaxDelay, s.CGRID, opts.String(), err.Error()))
		return
	}
	if maxDelay != 0 {
		rand.Seed(time.Now().Unix())
		ttl += time.Duration(
			rand.Int63n(maxDelay.Milliseconds()) * time.Millisecond.Nanoseconds())
	}
	// LastUsed
	var ttlLastUsed *time.Duration
	if opts.HasField(utils.OptsSessionTTLLastUsed) {
		ttlLastUsed, err = opts.GetDurationPtr(utils.OptsSessionTTLLastUsed)
	} else if s.OptsStart.HasField(utils.OptsSessionTTLLastUsed) {
		ttlLastUsed, err = s.OptsStart.GetDurationPtr(utils.OptsSessionTTLLastUsed)
	} else {
		ttlLastUsed = sS.cgrCfg.SessionSCfg().SessionTTLLastUsed
	}
	if err != nil {
		utils.Logger.Warning(
			fmt.Sprintf("<%s>, cannot extract <%s> for session:<%s>, from it's options: <%s>, err: <%s>",
				utils.SessionS, utils.OptsSessionTTLLastUsed, s.CGRID, opts.String(), err.Error()))
		return
	}
	// LastUsage
	var ttlLastUsage *time.Duration
	if opts.HasField(utils.OptsSessionTTLLastUsage) {
		ttlLastUsage, err = opts.GetDurationPtr(utils.OptsSessionTTLLastUsage)
	} else if s.OptsStart.HasField(utils.OptsSessionTTLLastUsage) {
		ttlLastUsage, err = s.OptsStart.GetDurationPtr(utils.OptsSessionTTLLastUsage)
	} else {
		ttlLastUsage = sS.cgrCfg.SessionSCfg().SessionTTLLastUsage
	}
	if err != nil {
		utils.Logger.Warning(
			fmt.Sprintf("<%s>, cannot extract <%s> for session:<%s>, from it's options: <%s>, err: <%s>",
				utils.SessionS, utils.OptsSessionTTLLastUsage, s.CGRID, opts.String(), err.Error()))
		return
	}
	// TTLUsage
	var ttlUsage *time.Duration
	if opts.HasField(utils.OptsSessionTTLUsage) {
		ttlUsage, err = opts.GetDurationPtr(utils.OptsSessionTTLUsage)
	} else if s.OptsStart.HasField(utils.OptsSessionTTLUsage) {
		ttlUsage, err = s.OptsStart.GetDurationPtr(utils.OptsSessionTTLUsage)
	} else {
		ttlUsage = sS.cgrCfg.SessionSCfg().SessionTTLUsage
	}
	if err != nil {
		utils.Logger.Warning(
			fmt.Sprintf("<%s>, cannot extract <%s> for session:<%s>, from it's options: <%s>, err: <%s>",
				utils.SessionS, utils.OptsSessionTTLUsage, s.CGRID, opts.String(), err.Error()))
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
			sS.forceSTerminate(s, lastUsage, s.sTerminator.ttlUsage,
				s.sTerminator.ttlLastUsed)
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
		var reply string
		for _, cgrEv := range s.asCGREvents() {
			argsProc := &engine.ArgV1ProcessEvent{
				Flags: []string{fmt.Sprintf("%s:false", utils.MetaChargers),
					fmt.Sprintf("%s:false", utils.MetaAttributes)},
				CGREvent: *cgrEv,
			}
			if unratedReqs.HasField( // order additional rating for unrated request types
				engine.MapEvent(cgrEv.Event).GetStringIgnoreErrors(utils.RequestType)) {
				argsProc.Flags = append(argsProc.Flags, utils.MetaRALs)
			}
			argsProc.SetCloneable(true)
			if err = sS.connMgr.Call(sS.cgrCfg.SessionSCfg().CDRsConns, nil,
				utils.CDRsV1ProcessEvent, argsProc, &reply); err != nil {
				utils.Logger.Warning(
					fmt.Sprintf(
						"<%s> could not post CDR for event %s, err: %s",
						utils.SessionS, utils.ToJSON(cgrEv), err.Error()))
			}
		}
	}
	// release the resources for the session
	if len(sS.cgrCfg.SessionSCfg().ResSConns) != 0 && s.ResourceID != "" {
		var reply string
		argsRU := &utils.ArgRSv1ResourceUsage{
			CGREvent: &utils.CGREvent{
				Tenant: s.Tenant,
				ID:     utils.GenUUID(),
				Event:  s.EventStart,
				Opts:   s.OptsStart,
			},
			UsageID: s.ResourceID,
			Units:   1,
		}
		argsRU.SetCloneable(true)
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
	if !s.chargeable {
		sS.pause(sr, dur)
		sr.TotalUsage += sr.LastUsage
		return dur, nil
	}
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
	cc := new(engine.CallCost)
	err = sS.connMgr.Call(sS.cgrCfg.SessionSCfg().RALsConns, nil,
		utils.ResponderMaxDebit,
		&engine.CallDescriptorWithOpts{
			CallDescriptor: cd,
			Opts:           s.OptsStart,
		}, cc)
	if err != nil {
		// verify in case of *dynaprepaid RequestType
		if err.Error() == utils.ErrAccountNotFound.Error() &&
			sr.Event.GetStringIgnoreErrors(utils.RequestType) == utils.MetaDynaprepaid {
			var reply string
			// execute the actionPlan configured in RalS
			if err = sS.connMgr.Call(sS.cgrCfg.SessionSCfg().SchedulerConns, nil,
				utils.SchedulerSv1ExecuteActionPlans, &utils.AttrsExecuteActionPlans{
					ActionPlanIDs: sS.cgrCfg.RalsCfg().DynaprepaidActionPlans,
					Tenant:        cd.Tenant, AccountID: cd.Account},
				&reply); err != nil {
				return
			}
			// execute again the MaxDebit operation
			err = sS.connMgr.Call(sS.cgrCfg.SessionSCfg().RALsConns, nil,
				utils.ResponderMaxDebit,
				&engine.CallDescriptorWithOpts{
					CallDescriptor: cd,
					Opts:           s.OptsStart,
				}, cc)
		}
		if err != nil {
			sr.ExtraDuration += dbtRsrv
			return 0, err
		}
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

func (sS *SessionS) pause(sr *SRun, dur time.Duration) {
	if sr.CD.LoopIndex > 0 {
		sr.CD.TimeStart = sr.CD.TimeEnd
	}
	sr.CD.TimeEnd = sr.CD.TimeStart.Add(dur)
	sr.CD.DurationIndex += dur
	ec := engine.NewFreeEventCost(sr.CD.CgrID, sr.CD.RunID, sr.CD.Account, sr.CD.TimeStart, dur)
	sr.LastUsage = dur
	sr.CD.LoopIndex++
	if sr.EventCost == nil { // is the first increment
		// when we start the call with debit interval 0
		// but later we update this value with one greater than 0
		sr.EventCost = ec
	} else { // we already debited something
		// copy the old AccountSummary as in Merge the old one is overwriten by the new one
		ec.AccountSummary = sr.EventCost.AccountSummary
		// similar to the debit merge the event costs
		sr.EventCost.Merge(ec)
	}
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
		if maxDebit < dbtIvl && sS.cgrCfg.SessionSCfg().MinDurLowBalance != time.Duration(0) { // warn client for low balance
			if sS.cgrCfg.SessionSCfg().MinDurLowBalance >= dbtIvl {
				utils.Logger.Warning(fmt.Sprintf("<%s> can not run warning for the session: <%s> since the remaining time:<%s> is higher than the debit interval:<%s>.",
					utils.SessionS, s.cgrID(), sS.cgrCfg.SessionSCfg().MinDurLowBalance, dbtIvl))
			} else if maxDebit <= sS.cgrCfg.SessionSCfg().MinDurLowBalance {
				go sS.warnSession(s.ClientConnID, s.EventStart.Clone())
			}
		}
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
		ToR:         utils.FirstNonEmpty(sr.CD.ToR, utils.MetaVoice),
		Increments:  incrmts,
	}
	var acnt engine.Account
	if err = sS.connMgr.Call(sS.cgrCfg.SessionSCfg().RALsConns, nil, utils.ResponderRefundIncrements,
		&engine.CallDescriptorWithOpts{
			CallDescriptor: cd,
			Opts:           s.OptsStart,
		}, &acnt); err != nil {
		return
	}
	if acnt.ID != "" { // Account info updated, update also cached AccountSummary
		acntSummary := acnt.AsAccountSummary()
		acntSummary.UpdateInitialValue(sr.EventCost.AccountSummary)
		sr.EventCost.AccountSummary = acntSummary
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
		Opts:           s.OptsStart,
		Tenant:         s.Tenant,
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
			&engine.CallDescriptorWithOpts{CallDescriptor: cd},
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

// warnSession will send warning from SessionS to clients
// regarding low balance
func (sS *SessionS) warnSession(connID string, ev map[string]interface{}) (err error) {
	clnt := sS.biJClnt(connID)
	if clnt == nil {
		return fmt.Errorf("calling %s requires bidirectional JSON connection, connID: <%s>",
			utils.SessionSv1WarnDisconnect, connID)
	}
	var rply string
	if err = clnt.conn.Call(utils.SessionSv1WarnDisconnect,
		ev, &rply); err != nil {
		if err != utils.ErrNotImplemented {
			utils.Logger.Warning(fmt.Sprintf("<%s> failed to warn session: <%s>, err: <%s>",
				utils.SessionS, ev[utils.CGRID], err))
			return
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
				fieldVal = utils.NotAvailable
			}
			if fieldVal == "" {
				fieldVal = utils.MetaEmpty
			}
			if _, hasFieldName := ssIndx[fieldName]; !hasFieldName { // Init it here
				ssIndx[fieldName] = make(map[string]map[string]utils.StringSet)
			}
			if _, hasFieldVal := ssIndx[fieldName][fieldVal]; !hasFieldVal {
				ssIndx[fieldName][fieldVal] = make(map[string]utils.StringSet)
			}
			if _, hasCGRID := ssIndx[fieldName][fieldVal][s.CGRID]; !hasCGRID {
				ssIndx[fieldName][fieldVal][s.CGRID] = make(utils.StringSet)
			}
			ssIndx[fieldName][fieldVal][s.CGRID].Add(sr.CD.RunID)

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
		f, err := sS.dm.GetFilter(tenant, fltrID,
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
				!sS.cgrCfg.SessionSCfg().SessionIndexes.Has(fldName) {
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
	pSessions bool) ([]string, map[string]utils.StringSet) {
	idxMux := &sS.aSIMux
	ssIndx := sS.aSessionsIdx
	if pSessions {
		idxMux = &sS.pSIMux
		ssIndx = sS.pSessionsIdx
	}
	idxMux.RLock()
	defer idxMux.RUnlock()
	matchingSessions := make(map[string]utils.StringSet)
	checkNr := 0
	getMatchingIndexes := func(fltrName string, values []string) (matchingSessionsbyValue map[string]utils.StringSet) {
		matchingSessionsbyValue = make(map[string]utils.StringSet)
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
					matchingSessionsbyValue[cgrID] = utils.StringSet{}
				}
				for runID := range runIDs {
					matchingSessionsbyValue[cgrID].Add(runID)
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
					if !matchedRunIDs.Has(runID) {
						delete(matchingSessions[cgrID], runID)
					}
				}
			}
		}
		if len(matchingSessions) == 0 {
			return make([]string, 0), make(map[string]utils.StringSet)
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
			if pass, err = fltr.Pass(ev); err != nil || !pass {
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
			if len(cgrIDs) != 0 && !runIDs.Has(sr.CD.RunID) {
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
			if pass, err = fltr.Pass(ev); err != nil || !pass {
				return
			}
		}
		return
	}
	for _, s := range ss {
		s.RLock()
		runIDs := matchingSRuns[s.CGRID]
		for _, sr := range s.SRuns {
			if len(cgrIDs) != 0 && !runIDs.Has(sr.CD.RunID) {
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

// newSession will populate SRuns within a Session based on ChargerS output
// forSession can only be called once per Session
// not thread-safe since it should be called in init where there is no concurrency
func (sS *SessionS) newSession(cgrEv *utils.CGREvent, resID, clntConnID string,
	dbtItval time.Duration, forceDuration, isMsg bool) (s *Session, err error) {
	if len(sS.cgrCfg.SessionSCfg().ChargerSConns) == 0 {
		err = errors.New("ChargerS is disabled")
		return
	}
	cgrID := GetSetCGRID(cgrEv.Event)
	s = &Session{
		CGRID:         cgrID,
		Tenant:        cgrEv.Tenant,
		ResourceID:    resID,
		EventStart:    engine.MapEvent(cgrEv.Event).Clone(), // decouple the event from the request so we can avoid concurrency with debit and ttl
		OptsStart:     engine.MapEvent(cgrEv.Opts).Clone(),
		ClientConnID:  clntConnID,
		DebitInterval: dbtItval,
	}
	s.chargeable = s.OptsStart.GetBoolOrDefault(utils.OptsChargeable, true)
	if !isMsg && sS.isIndexed(s, false) { // check if already exists
		return nil, utils.ErrExists
	}

	var chrgrs []*engine.ChrgSProcessEventReply
	if chrgrs, err = sS.processChargerS(cgrEv); err != nil {
		return
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
			subject = me.GetStringIgnoreErrors(utils.AccountField)
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
				Account:       me.GetStringIgnoreErrors(utils.AccountField),
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

// processChargerS processes the event with chargers and cahces the response based on the requestID
func (sS *SessionS) processChargerS(cgrEv *utils.CGREvent) (chrgrs []*engine.ChrgSProcessEventReply, err error) {
	if x, ok := engine.Cache.Get(utils.CacheEventCharges, cgrEv.ID); ok && x != nil {
		return x.([]*engine.ChrgSProcessEventReply), nil
	}
	if err = sS.connMgr.Call(sS.cgrCfg.SessionSCfg().ChargerSConns, nil,
		utils.ChargerSv1ProcessEvent, cgrEv, &chrgrs); err != nil {
		err = utils.NewErrChargerS(err)
	}

	if errCh := engine.Cache.Set(utils.CacheEventCharges, cgrEv.ID, chrgrs, nil,
		true, utils.NonTransactional); errCh != nil {
		return nil, errCh
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
	sS.terminateSyncSessions(toBeRemoved)
}

// Extracted from syncSessions in order to test all cases
func (sS *SessionS) terminateSyncSessions(toBeRemoved []string) {
	for _, cgrID := range toBeRemoved {
		ss := sS.getSessions(cgrID, false)
		if len(ss) == 0 {
			continue
		}
		ss[0].Lock()
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
		if s.DebitInterval > 0 &&
			sr.Event.GetStringIgnoreErrors(utils.RequestType) == utils.MetaPrepaid {
			if s.debitStop == nil { // init the debitStop only for the first sRun with DebitInterval and RequestType MetaPrepaid
				s.debitStop = make(chan struct{})
			}
			go sS.debitLoopSession(s, i, s.DebitInterval)
			runtime.Gosched() // allow the goroutine to be executed
		}
	}
}

// authEvent calculates maximum usage allowed for the given event
func (sS *SessionS) authEvent(cgrEv *utils.CGREvent, forceDuration bool) (usage map[string]time.Duration, err error) {
	evStart := engine.MapEvent(cgrEv.Event)
	var eventUsage time.Duration
	if eventUsage, err = evStart.GetDuration(utils.Usage); err != nil {
		if err != utils.ErrNotFound {
			return
		}
		err = nil
		eventUsage = sS.cgrCfg.GeneralCfg().MaxCallDuration
		evStart[utils.Usage] = eventUsage // will be used in CD
	}
	var s *Session
	if s, err = sS.newSession(cgrEv, "", "", 0, forceDuration, true); err != nil {
		return
	}
	usage = make(map[string]time.Duration)
	for _, sr := range s.SRuns {
		var rplyMaxUsage time.Duration
		if !authReqs.HasField(
			sr.Event.GetStringIgnoreErrors(utils.RequestType)) {
			rplyMaxUsage = eventUsage
		} else if err = sS.connMgr.Call(sS.cgrCfg.SessionSCfg().RALsConns, nil,
			utils.ResponderGetMaxSessionTime,
			&engine.CallDescriptorWithOpts{
				CallDescriptor: sr.CD,
				Opts:           s.OptsStart,
			}, &rplyMaxUsage); err != nil {
			err = utils.NewErrRALs(err)
			return
		}
		if rplyMaxUsage > eventUsage {
			rplyMaxUsage = eventUsage
		}
		usage[sr.CD.RunID] = rplyMaxUsage
	}
	return
}

// initSession handles a new session
// not thread-safe for Session since it is constructed here
func (sS *SessionS) initSession(cgrEv *utils.CGREvent, clntConnID,
	resID string, dbtItval time.Duration, isMsg, forceDuration bool) (s *Session, err error) {
	if s, err = sS.newSession(cgrEv, resID, clntConnID, dbtItval, forceDuration, isMsg); err != nil {
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
func (sS *SessionS) updateSession(s *Session, updtEv, opts engine.MapEvent, isMsg bool) (maxUsage map[string]time.Duration, err error) {
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
		sS.setSTerminator(s, opts) // reset the terminator
	}
	s.chargeable = opts.GetBoolOrDefault(utils.OptsChargeable, true)
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
		reqMaxUsage = sS.cgrCfg.GeneralCfg().MaxCallDuration
		updtEv[utils.Usage] = reqMaxUsage
	}
	maxUsage = make(map[string]time.Duration)
	for i, sr := range s.SRuns {
		reqType := sr.Event.GetStringIgnoreErrors(utils.RequestType)
		if reqType == utils.MetaNone {
			continue
		}
		var rplyMaxUsage time.Duration
		if reqType != utils.MetaPrepaid || s.debitStop != nil {
			rplyMaxUsage = reqMaxUsage
		} else if rplyMaxUsage, err = sS.debitSession(s, i, reqMaxUsage,
			updtEv.GetDurationPtrIgnoreErrors(utils.LastUsed)); err != nil {
			return
		}
		maxUsage[sr.CD.RunID] = rplyMaxUsage
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
					if !s.chargeable {
						sS.pause(sr, notCharged)
					} else {
						if sr.CD.LoopIndex > 0 {
							sr.CD.TimeStart = sr.CD.TimeEnd
						}
						sr.CD.TimeEnd = sr.CD.TimeStart.Add(notCharged)
						sr.CD.DurationIndex += notCharged
						cc := new(engine.CallCost)
						if err = sS.connMgr.Call(sS.cgrCfg.SessionSCfg().RALsConns, nil, utils.ResponderDebit,
							&engine.CallDescriptorWithOpts{
								CallDescriptor: sr.CD,
								Opts:           s.OptsStart,
							}, cc); err == nil {
							sr.EventCost.Merge(
								engine.NewEventCostFromCallCost(cc, s.CGRID,
									sr.Event.GetStringIgnoreErrors(utils.RunID)))
						}
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
	if errCh := engine.Cache.Set(utils.CacheClosedSessions, s.CGRID, s,
		nil, true, utils.NonTransactional); errCh != nil {
		return errCh
	}
	return
}

// chargeEvent will charge a single event (ie: SMS)
func (sS *SessionS) chargeEvent(cgrEv *utils.CGREvent, forceDuration bool) (maxUsage time.Duration, err error) {
	var s *Session
	if s, err = sS.initSession(cgrEv, "", "", 0, true, forceDuration); err != nil {
		return
	}
	cgrID := s.CGRID
	var sRunsUsage map[string]time.Duration
	if sRunsUsage, err = sS.updateSession(s, nil, nil, true); err != nil {
		if errEnd := sS.terminateSession(s,
			utils.DurationPointer(time.Duration(0)), nil, nil, true); errEnd != nil {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> error when force-ending charged event: <%s>, err: <%s>",
					utils.SessionS, cgrID, errEnd.Error()))
		}
		err = utils.NewErrRALs(err)
		return
	}
	var maxUsageSet bool // so we know if we have set the 0 on purpose
	for _, rplyMaxUsage := range sRunsUsage {
		if !maxUsageSet || rplyMaxUsage < maxUsage {
			maxUsage = rplyMaxUsage
			maxUsageSet = true
		}
	}
	ev := engine.MapEvent(cgrEv.Event)
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
	res, maxUsage, routes, routesIgnoreErrs, routesEventCost bool,
	cgrEv *utils.CGREvent, routePaginator utils.Paginator,
	forceDuration bool, routesMaxCost string) (args *V1AuthorizeArgs) {
	args = &V1AuthorizeArgs{
		GetAttributes:      attrs,
		AuthorizeResources: res,
		GetMaxUsage:        maxUsage,
		ProcessThresholds:  thrslds,
		ProcessStats:       statQueues,
		RoutesIgnoreErrors: routesIgnoreErrs,
		GetRoutes:          routes,
		CGREvent:           cgrEv,
		ForceDuration:      forceDuration,
	}
	if routesEventCost {
		args.RoutesMaxCost = utils.MetaEventCost
	} else {
		args.RoutesMaxCost = routesMaxCost
	}
	args.Paginator = routePaginator
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
	GetAttributes      bool
	AuthorizeResources bool
	GetMaxUsage        bool
	ForceDuration      bool
	ProcessThresholds  bool
	ProcessStats       bool
	GetRoutes          bool
	RoutesMaxCost      string
	RoutesIgnoreErrors bool
	AttributeIDs       []string
	ThresholdIDs       []string
	StatIDs            []string
	*utils.CGREvent
	utils.Paginator
}

// ParseFlags will populate the V1AuthorizeArgs flags
func (args *V1AuthorizeArgs) ParseFlags(flags string) {
	for _, subsystem := range strings.Split(flags, utils.FieldsSep) {
		switch {
		case subsystem == utils.MetaAccounts:
			args.GetMaxUsage = true
		case subsystem == utils.MetaResources:
			args.AuthorizeResources = true
		case subsystem == utils.MetaRoutes:
			args.GetRoutes = true
		case subsystem == utils.MetaRoutesIgnoreErrors:
			args.RoutesIgnoreErrors = true
		case subsystem == utils.MetaRoutesEventCost:
			args.RoutesMaxCost = utils.MetaEventCost
		case strings.HasPrefix(subsystem, utils.MetaRoutesMaxCost):
			args.RoutesMaxCost = strings.TrimPrefix(subsystem, utils.MetaRoutesMaxCost+utils.InInFieldSep)
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
	args.Paginator, _ = utils.GetRoutePaginatorFromOpts(args.Opts)
	return
}

// V1AuthorizeReply are options available in auth reply
type V1AuthorizeReply struct {
	Attributes         *engine.AttrSProcessEventReply
	ResourceAllocation *string
	MaxUsage           *time.Duration
	Routes             *engine.SortedRoutes
	ThresholdIDs       *[]string
	StatQueueIDs       *[]string
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
	if v1AuthReply.MaxUsage != nil {
		cgrReply[utils.CapMaxUsage] = utils.NewNMData(*v1AuthReply.MaxUsage)
	}
	if v1AuthReply.Routes != nil {
		cgrReply[utils.CapRoutes] = v1AuthReply.Routes.AsNavigableMap()
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
	if args.CGREvent.Tenant == "" {
		args.CGREvent.Tenant = sS.cgrCfg.GeneralCfg().DefaultTenant
	}

	// RPC caching
	if sS.cgrCfg.CacheCfg().Partitions[utils.CacheRPCResponses].Limit != 0 {
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
		!args.GetMaxUsage && !args.GetRoutes {
		return utils.NewErrMandatoryIeMissing("subsystems")
	}
	if args.GetAttributes {
		rplyAttr, err := sS.processAttributes(args.CGREvent, args.AttributeIDs, false)
		if err == nil {
			args.CGREvent = rplyAttr.CGREvent
			authReply.Attributes = &rplyAttr
		} else if err.Error() != utils.ErrNotFound.Error() {
			return utils.NewErrAttributeS(err)
		}
	}
	if args.GetMaxUsage {
		var sRunsUsage map[string]time.Duration
		if sRunsUsage, err = sS.authEvent(args.CGREvent, args.ForceDuration); err != nil {
			return err
		}

		var maxUsage time.Duration
		var maxUsageSet bool // so we know if we have set the 0 on purpose
		for _, rplyMaxUsage := range sRunsUsage {
			if !maxUsageSet || rplyMaxUsage < maxUsage {
				maxUsage = rplyMaxUsage
				maxUsageSet = true
			}
		}
		authReply.MaxUsage = &maxUsage
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
		attrRU := &utils.ArgRSv1ResourceUsage{
			CGREvent: args.CGREvent,
			UsageID:  originID,
			Units:    1,
		}
		if err = sS.connMgr.Call(sS.cgrCfg.SessionSCfg().ResSConns, nil, utils.ResourceSv1AuthorizeResources,
			attrRU, &allocMsg); err != nil {
			return utils.NewErrResourceS(err)
		}
		authReply.ResourceAllocation = &allocMsg
	}
	if args.GetRoutes {
		routesReply, err := sS.getRoutes(args.CGREvent.Clone(), args.Paginator,
			args.RoutesIgnoreErrors, args.RoutesMaxCost, false)
		if err != nil {
			return err
		}
		if routesReply.SortedRoutes != nil {
			authReply.Routes = &routesReply
		}
	}
	if args.ProcessThresholds {
		tIDs, err := sS.processThreshold(args.CGREvent, args.ThresholdIDs, true)
		if err != nil && err.Error() != utils.ErrNotFound.Error() {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> error: %s processing event %+v with ThresholdS.",
					utils.SessionS, err.Error(), args.CGREvent))
			withErrors = true
		}
		authReply.ThresholdIDs = &tIDs
	}
	if args.ProcessStats {
		sIDs, err := sS.processStats(args.CGREvent, args.StatIDs, false)
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
	RoutesDigest       *string
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
	if args.GetRoutes {
		authReply.RoutesDigest = utils.StringPointer(initAuthRply.Routes.Digest())
	}
	if args.ProcessThresholds {
		authReply.Thresholds = utils.StringPointer(
			strings.Join(*initAuthRply.ThresholdIDs, utils.FieldsSep))
	}
	if args.ProcessStats {
		authReply.StatQueues = utils.StringPointer(
			strings.Join(*initAuthRply.StatQueueIDs, utils.FieldsSep))
	}
	return nil
}

// NewV1InitSessionArgs is a constructor for V1InitSessionArgs
func NewV1InitSessionArgs(attrs bool, attributeIDs []string,
	thrslds bool, thresholdIDs []string, stats bool, statIDs []string,
	resrc, acnt bool, cgrEv *utils.CGREvent, forceDuration bool) (args *V1InitSessionArgs) {
	args = &V1InitSessionArgs{
		GetAttributes:     attrs,
		AllocateResources: resrc,
		InitSession:       acnt,
		ProcessThresholds: thrslds,
		ProcessStats:      stats,
		CGREvent:          cgrEv,
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
}

// ParseFlags will populate the V1InitSessionArgs flags
func (args *V1InitSessionArgs) ParseFlags(flags string) {
	for _, subsystem := range strings.Split(flags, utils.FieldsSep) {
		switch {
		case subsystem == utils.MetaAccounts:
			args.InitSession = true
		case subsystem == utils.MetaResources:
			args.AllocateResources = true
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
}

// V1InitSessionReply are options for initialization reply
type V1InitSessionReply struct {
	Attributes         *engine.AttrSProcessEventReply
	ResourceAllocation *string
	MaxUsage           *time.Duration
	ThresholdIDs       *[]string
	StatQueueIDs       *[]string
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
	if v1Rply.MaxUsage != nil {
		cgrReply[utils.CapMaxUsage] = utils.NewNMData(*v1Rply.MaxUsage)
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
	if args.CGREvent.Tenant == "" {
		args.CGREvent.Tenant = sS.cgrCfg.GeneralCfg().DefaultTenant
	}

	// RPC caching
	if sS.cgrCfg.CacheCfg().Partitions[utils.CacheRPCResponses].Limit != 0 {
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
		rplyAttr, err := sS.processAttributes(args.CGREvent, args.AttributeIDs, false)
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
		attrRU := &utils.ArgRSv1ResourceUsage{
			CGREvent: args.CGREvent,
			UsageID:  originID,
			Units:    1,
		}
		var allocMessage string
		if err = sS.connMgr.Call(sS.cgrCfg.SessionSCfg().ResSConns, nil,
			utils.ResourceSv1AllocateResources, attrRU, &allocMessage); err != nil {
			return utils.NewErrResourceS(err)
		}
		rply.ResourceAllocation = &allocMessage
	}
	if args.InitSession {
		var err error
		opts := engine.MapEvent(args.Opts)
		dbtItvl := sS.cgrCfg.SessionSCfg().DebitInterval
		if opts.HasField(utils.OptsDebitInterval) { // dynamic DebitInterval via CGRDebitInterval
			if dbtItvl, err = opts.GetDuration(utils.OptsDebitInterval); err != nil {
				return utils.NewErrRALs(err)
			}
		}
		s, err := sS.initSession(args.CGREvent, sS.biJClntID(clnt), originID, dbtItvl,
			false, args.ForceDuration)
		if err != nil {
			return err
		}
		s.RLock() // avoid concurrency with activeDebit
		isPrepaid := s.debitStop != nil
		s.RUnlock()
		if isPrepaid { //active debit
			rply.MaxUsage = &sS.cgrCfg.GeneralCfg().MaxCallDuration
		} else {
			var sRunsUsage map[string]time.Duration
			if sRunsUsage, err = sS.updateSession(s, nil, args.Opts, false); err != nil {
				return utils.NewErrRALs(err)
			}

			var maxUsage time.Duration
			var maxUsageSet bool // so we know if we have set the 0 on purpose
			for _, rplyMaxUsage := range sRunsUsage {
				if !maxUsageSet || rplyMaxUsage < maxUsage {
					maxUsage = rplyMaxUsage
					maxUsageSet = true
				}
			}
			rply.MaxUsage = &maxUsage
		}
	}
	if args.ProcessThresholds {
		tIDs, err := sS.processThreshold(args.CGREvent, args.ThresholdIDs, true)
		if err != nil && err.Error() != utils.ErrNotFound.Error() {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> error: %s processing event %+v with ThresholdS.",
					utils.SessionS, err.Error(), args.CGREvent))
			withErrors = true
		}
		rply.ThresholdIDs = &tIDs
	}
	if args.ProcessStats {
		sIDs, err := sS.processStats(args.CGREvent, args.StatIDs, false)
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
			strings.Join(*initSessionRply.ThresholdIDs, utils.FieldsSep))
	}
	if args.ProcessStats {
		initReply.StatQueues = utils.StringPointer(
			strings.Join(*initSessionRply.StatQueueIDs, utils.FieldsSep))
	}
	return nil
}

// NewV1UpdateSessionArgs is a constructor for update session arguments
func NewV1UpdateSessionArgs(attrs bool, attributeIDs []string,
	acnts bool, cgrEv *utils.CGREvent, forceDuration bool) (args *V1UpdateSessionArgs) {
	args = &V1UpdateSessionArgs{
		GetAttributes: attrs,
		UpdateSession: acnts,
		CGREvent:      cgrEv,
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
}

// V1UpdateSessionReply contains options for session update reply
type V1UpdateSessionReply struct {
	Attributes *engine.AttrSProcessEventReply
	MaxUsage   *time.Duration
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
	if v1Rply.MaxUsage != nil {
		cgrReply[utils.CapMaxUsage] = utils.NewNMData(*v1Rply.MaxUsage)
	}
	return cgrReply
}

// BiRPCv1UpdateSession updates an existing session, returning the duration which the session can still last
func (sS *SessionS) BiRPCv1UpdateSession(clnt rpcclient.ClientConnector,
	args *V1UpdateSessionArgs, rply *V1UpdateSessionReply) (err error) {
	if args.CGREvent == nil {
		return utils.NewErrMandatoryIeMissing(utils.CGREventString)
	}
	if args.CGREvent.ID == utils.EmptyString {
		args.CGREvent.ID = utils.GenUUID()
	}
	if args.CGREvent.Tenant == utils.EmptyString {
		args.CGREvent.Tenant = sS.cgrCfg.GeneralCfg().DefaultTenant
	}

	// RPC caching
	if sS.cgrCfg.CacheCfg().Partitions[utils.CacheRPCResponses].Limit != 0 {
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

	if args.GetAttributes {
		rplyAttr, err := sS.processAttributes(args.CGREvent, args.AttributeIDs, false)
		if err == nil {
			args.CGREvent = rplyAttr.CGREvent
			rply.Attributes = &rplyAttr
		} else if err.Error() != utils.ErrNotFound.Error() {
			return utils.NewErrAttributeS(err)
		}
	}
	if args.UpdateSession {
		ev := engine.MapEvent(args.CGREvent.Event)
		opts := engine.MapEvent(args.Opts)
		dbtItvl := sS.cgrCfg.SessionSCfg().DebitInterval
		if opts.HasField(utils.OptsDebitInterval) { // dynamic DebitInterval via CGRDebitInterval
			if dbtItvl, err = opts.GetDuration(utils.OptsDebitInterval); err != nil {
				return utils.NewErrRALs(err)
			}
		}
		cgrID := GetSetCGRID(ev)
		s := sS.getRelocateSession(cgrID,
			ev.GetStringIgnoreErrors(utils.InitialOriginID),
			ev.GetStringIgnoreErrors(utils.OriginID),
			ev.GetStringIgnoreErrors(utils.OriginHost))
		if s == nil {
			if s, err = sS.initSession(args.CGREvent, sS.biJClntID(clnt), ev.GetStringIgnoreErrors(utils.OriginID),
				dbtItvl, false, args.ForceDuration); err != nil {
				return err
			}
		}
		var sRunsUsage map[string]time.Duration
		if sRunsUsage, err = sS.updateSession(s, ev, args.Opts, false); err != nil {
			return utils.NewErrRALs(err)
		}
		var maxUsage time.Duration
		var maxUsageSet bool // so we know if we have set the 0 on purpose
		for _, rplyMaxUsage := range sRunsUsage {
			if !maxUsageSet || rplyMaxUsage < maxUsage {
				maxUsage = rplyMaxUsage
				maxUsageSet = true
			}
		}
		rply.MaxUsage = &maxUsage
	}
	return
}

// NewV1TerminateSessionArgs creates a new V1TerminateSessionArgs using the given arguments
func NewV1TerminateSessionArgs(acnts, resrc,
	thrds bool, thresholdIDs []string, stats bool,
	statIDs []string, cgrEv *utils.CGREvent, forceDuration bool) (args *V1TerminateSessionArgs) {
	args = &V1TerminateSessionArgs{
		TerminateSession:  acnts,
		ReleaseResources:  resrc,
		ProcessThresholds: thrds,
		ProcessStats:      stats,
		CGREvent:          cgrEv,
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
}

// ParseFlags will populate the V1TerminateSessionArgs flags
func (args *V1TerminateSessionArgs) ParseFlags(flags string) {
	for _, subsystem := range strings.Split(flags, utils.FieldsSep) {
		switch {
		case subsystem == utils.MetaAccounts:
			args.TerminateSession = true
		case subsystem == utils.MetaResources:
			args.ReleaseResources = true
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
	if args.CGREvent.Tenant == "" {
		args.CGREvent.Tenant = sS.cgrCfg.GeneralCfg().DefaultTenant
	}
	// RPC caching
	if sS.cgrCfg.CacheCfg().Partitions[utils.CacheRPCResponses].Limit != 0 {
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

	ev := engine.MapEvent(args.CGREvent.Event)
	opts := engine.MapEvent(args.Opts)
	cgrID := GetSetCGRID(ev)
	originID := ev.GetStringIgnoreErrors(utils.OriginID)
	if args.TerminateSession {
		if originID == "" {
			return utils.NewErrMandatoryIeMissing(utils.OriginID)
		}
		dbtItvl := sS.cgrCfg.SessionSCfg().DebitInterval
		if opts.HasField(utils.OptsDebitInterval) { // dynamic DebitInterval via CGRDebitInterval
			if dbtItvl, err = opts.GetDuration(utils.OptsDebitInterval); err != nil {
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
			if s, err = sS.initSession(args.CGREvent, sS.biJClntID(clnt), ev.GetStringIgnoreErrors(utils.OriginID),
				dbtItvl, isMsg, args.ForceDuration); err != nil {
				return utils.NewErrRALs(err)
			}
			if _, err = sS.updateSession(s, ev, opts, isMsg); err != nil {
				return err
			}
			break
		}
		if !isMsg {
			s.UpdateSRuns(ev, sS.cgrCfg.SessionSCfg().AlterableFields)
		}
		s.Lock()
		s.chargeable = opts.GetBoolOrDefault(utils.OptsChargeable, true)
		s.Unlock()
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
		argsRU := &utils.ArgRSv1ResourceUsage{
			CGREvent: args.CGREvent,
			UsageID:  originID, // same ID should be accepted by first group since the previous resource should be expired
			Units:    1,
		}
		if err = sS.connMgr.Call(sS.cgrCfg.SessionSCfg().ResSConns, nil, utils.ResourceSv1ReleaseResources,
			argsRU, &reply); err != nil {
			return utils.NewErrResourceS(err)
		}
	}
	if args.ProcessThresholds {
		_, err := sS.processThreshold(args.CGREvent, args.ThresholdIDs, true)
		if err != nil &&
			err.Error() != utils.ErrNotFound.Error() {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> error: %s processing event %+v with ThresholdS.",
					utils.SessionS, err.Error(), args.CGREvent))
			withErrors = true
		}
	}
	if args.ProcessStats {
		_, err := sS.processStats(args.CGREvent, args.StatIDs, false)
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
	cgrEv *utils.CGREvent, rply *string) (err error) {
	if cgrEv.ID == utils.EmptyString {
		cgrEv.ID = utils.GenUUID()
	}
	if cgrEv.Tenant == utils.EmptyString {
		cgrEv.Tenant = sS.cgrCfg.GeneralCfg().DefaultTenant
	}

	// RPC caching
	if sS.cgrCfg.CacheCfg().Partitions[utils.CacheRPCResponses].Limit != 0 {
		cacheKey := utils.ConcatenatedKey(utils.SessionSv1ProcessCDR, cgrEv.ID)
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
	if _, has := cgrEv.Event[utils.Source]; !has {
		cgrEv.Event[utils.Source] = utils.MetaSessionS
	}

	return sS.processCDR(cgrEv, []string{utils.MetaRALs}, rply, false)
}

// NewV1ProcessMessageArgs is a constructor for MessageArgs used by ProcessMessage
func NewV1ProcessMessageArgs(attrs bool, attributeIDs []string,
	thds bool, thresholdIDs []string, stats bool, statIDs []string, resrc, acnts,
	routes, routesIgnoreErrs, routesEventCost bool, cgrEv *utils.CGREvent,
	routePaginator utils.Paginator, forceDuration bool, routesMaxCost string) (args *V1ProcessMessageArgs) {
	args = &V1ProcessMessageArgs{
		AllocateResources:  resrc,
		Debit:              acnts,
		GetAttributes:      attrs,
		ProcessThresholds:  thds,
		ProcessStats:       stats,
		RoutesIgnoreErrors: routesIgnoreErrs,
		GetRoutes:          routes,
		CGREvent:           cgrEv,
		ForceDuration:      forceDuration,
	}
	if routesEventCost {
		args.RoutesMaxCost = utils.MetaEventCost
	} else {
		args.RoutesMaxCost = routesMaxCost
	}
	args.Paginator = routePaginator
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
	GetAttributes      bool
	AllocateResources  bool
	Debit              bool
	ForceDuration      bool
	ProcessThresholds  bool
	ProcessStats       bool
	GetRoutes          bool
	RoutesMaxCost      string
	RoutesIgnoreErrors bool
	AttributeIDs       []string
	ThresholdIDs       []string
	StatIDs            []string
	*utils.CGREvent
	utils.Paginator
}

// ParseFlags will populate the V1ProcessMessageArgs flags
func (args *V1ProcessMessageArgs) ParseFlags(flags string) {
	for _, subsystem := range strings.Split(flags, utils.FieldsSep) {
		switch {
		case subsystem == utils.MetaAccounts:
			args.Debit = true
		case subsystem == utils.MetaResources:
			args.AllocateResources = true
		case subsystem == utils.MetaRoutes:
			args.GetRoutes = true
		case subsystem == utils.MetaRoutesIgnoreErrors:
			args.RoutesIgnoreErrors = true
		case subsystem == utils.MetaRoutesEventCost:
			args.RoutesMaxCost = utils.MetaEventCost
		case strings.HasPrefix(subsystem, utils.MetaRoutesMaxCost):
			args.RoutesMaxCost = strings.TrimPrefix(subsystem, utils.MetaRoutesMaxCost+utils.InInFieldSep)
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
	args.Paginator, _ = utils.GetRoutePaginatorFromOpts(args.Opts)
	return
}

// V1ProcessMessageReply is the reply for the ProcessMessage API
type V1ProcessMessageReply struct {
	MaxUsage           *time.Duration
	ResourceAllocation *string
	Attributes         *engine.AttrSProcessEventReply
	Routes             *engine.SortedRoutes
	ThresholdIDs       *[]string
	StatQueueIDs       *[]string
}

// AsNavigableMap is part of engine.NavigableMapper interface
func (v1Rply *V1ProcessMessageReply) AsNavigableMap() utils.NavigableMap2 {
	cgrReply := make(utils.NavigableMap2)
	if v1Rply.MaxUsage != nil {
		cgrReply[utils.CapMaxUsage] = utils.NewNMData(*v1Rply.MaxUsage)
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
	if v1Rply.Routes != nil {
		cgrReply[utils.CapRoutes] = v1Rply.Routes.AsNavigableMap()
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
	if args.CGREvent.ID == utils.EmptyString {
		args.CGREvent.ID = utils.GenUUID()
	}
	if args.CGREvent.Tenant == utils.EmptyString {
		args.CGREvent.Tenant = sS.cgrCfg.GeneralCfg().DefaultTenant
	}

	// RPC caching
	if sS.cgrCfg.CacheCfg().Partitions[utils.CacheRPCResponses].Limit != 0 {
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

	me := engine.MapEvent(args.CGREvent.Event)
	originID := me.GetStringIgnoreErrors(utils.OriginID)

	if args.GetAttributes {
		rplyAttr, err := sS.processAttributes(args.CGREvent, args.AttributeIDs, false)
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
		attrRU := &utils.ArgRSv1ResourceUsage{
			CGREvent: args.CGREvent,
			UsageID:  originID,
			Units:    1,
		}
		var allocMessage string
		if err = sS.connMgr.Call(sS.cgrCfg.SessionSCfg().ResSConns, nil, utils.ResourceSv1AllocateResources,
			attrRU, &allocMessage); err != nil {
			return utils.NewErrResourceS(err)
		}
		rply.ResourceAllocation = &allocMessage
	}
	if args.GetRoutes {
		routesReply, err := sS.getRoutes(args.CGREvent.Clone(), args.Paginator,
			args.RoutesIgnoreErrors, args.RoutesMaxCost, false)
		if err != nil {
			return err
		}
		if routesReply.SortedRoutes != nil {
			rply.Routes = &routesReply
		}
	}
	if args.Debit {
		var maxUsage time.Duration
		if maxUsage, err = sS.chargeEvent(args.CGREvent, args.ForceDuration); err != nil {
			return err
		}
		rply.MaxUsage = &maxUsage
	}
	if args.ProcessThresholds {
		tIDs, err := sS.processThreshold(args.CGREvent, args.ThresholdIDs, true)
		if err != nil && err.Error() != utils.ErrNotFound.Error() {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> error: %s processing event %+v with ThresholdS.",
					utils.SessionS, err.Error(), args.CGREvent))
			withErrors = true
		}
		rply.ThresholdIDs = &tIDs
	}
	if args.ProcessStats {
		sIDs, err := sS.processStats(args.CGREvent, args.StatIDs, false)
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
}

// V1ProcessEventReply is the reply for the ProcessEvent API
type V1ProcessEventReply struct {
	MaxUsage           map[string]time.Duration
	Cost               map[string]float64 // Cost is the cost received from Rater, ignoring accounting part
	ResourceAllocation map[string]string
	Attributes         map[string]*engine.AttrSProcessEventReply
	Routes             map[string]*engine.SortedRoutes
	ThresholdIDs       map[string][]string
	StatQueueIDs       map[string][]string
	STIRIdentity       map[string]string
}

// AsNavigableMap is part of engine.NavigableMapper interface
func (v1Rply *V1ProcessEventReply) AsNavigableMap() utils.NavigableMap2 {
	cgrReply := make(utils.NavigableMap2)
	if v1Rply.MaxUsage != nil {
		usage := make(utils.NavigableMap2)
		for k, v := range v1Rply.MaxUsage {
			usage[k] = utils.NewNMData(v)
		}
		cgrReply[utils.CapMaxUsage] = usage
	}
	if v1Rply.ResourceAllocation != nil {
		res := make(utils.NavigableMap2)
		for k, v := range v1Rply.ResourceAllocation {
			res[k] = utils.NewNMData(v)
		}
		cgrReply[utils.CapResourceAllocation] = res
	}
	if v1Rply.Attributes != nil {
		atts := make(utils.NavigableMap2)
		for k, att := range v1Rply.Attributes {
			attrs := make(utils.NavigableMap2)
			for _, fldName := range att.AlteredFields {
				fldName = strings.TrimPrefix(fldName, utils.MetaReq+utils.NestingSep)
				if att.CGREvent.HasField(fldName) {
					attrs[fldName] = utils.NewNMData(att.CGREvent.Event[fldName])
				}
			}
			atts[k] = attrs
		}
		cgrReply[utils.CapAttributes] = atts
	}
	if v1Rply.Routes != nil {
		routes := make(utils.NavigableMap2)
		for k, route := range v1Rply.Routes {
			routes[k] = route.AsNavigableMap()
		}
		cgrReply[utils.CapRoutes] = routes
	}
	if v1Rply.ThresholdIDs != nil {
		th := make(utils.NavigableMap2)
		for k, thr := range v1Rply.ThresholdIDs {
			thIDs := make(utils.NMSlice, len(thr))
			for i, v := range thr {
				thIDs[i] = utils.NewNMData(v)
			}
			th[k] = &thIDs
		}
		cgrReply[utils.CapThresholds] = th
	}
	if v1Rply.StatQueueIDs != nil {
		st := make(utils.NavigableMap2)
		for k, sts := range v1Rply.StatQueueIDs {
			stIDs := make(utils.NMSlice, len(sts))
			for i, v := range sts {
				stIDs[i] = utils.NewNMData(v)
			}
			st[k] = &stIDs
		}
		cgrReply[utils.CapStatQueues] = st
	}
	if v1Rply.Cost != nil {
		costs := make(utils.NavigableMap2)
		for k, cost := range v1Rply.Cost {
			costs[k] = utils.NewNMData(cost)
		}
	}
	if v1Rply.STIRIdentity != nil {
		stir := make(utils.NavigableMap2)
		for k, v := range v1Rply.STIRIdentity {
			stir[k] = utils.NewNMData(v)
		}
		cgrReply[utils.OptsStirIdentity] = stir
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
	if args.CGREvent.Tenant == "" {
		args.CGREvent.Tenant = sS.cgrCfg.GeneralCfg().DefaultTenant
	}

	// RPC caching
	if sS.cgrCfg.CacheCfg().Partitions[utils.CacheRPCResponses].Limit != 0 {
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

	//convert from Flags []string to utils.FlagsWithParams
	argsFlagsWithParams := utils.FlagsWithParamsFromSlice(args.Flags)

	blockError := argsFlagsWithParams.GetBool(utils.MetaBlockerError)
	events := map[string]*utils.CGREvent{
		utils.MetaRaw: args.CGREvent,
	}
	if argsFlagsWithParams.GetBool(utils.MetaChargers) {
		var chrgrs []*engine.ChrgSProcessEventReply
		if chrgrs, err = sS.processChargerS(args.CGREvent); err != nil {
			return
		}
		for _, chrgr := range chrgrs {
			events[utils.IfaceAsString(chrgr.CGREvent.Event[utils.RunID])] = chrgr.CGREvent
		}
	}

	// check for *attribute
	if argsFlagsWithParams.GetBool(utils.MetaAttributes) {
		attrIDs := argsFlagsWithParams.ParamsSlice(utils.MetaAttributes, utils.MetaIDs)
		rply.Attributes = make(map[string]*engine.AttrSProcessEventReply)

		for runID, cgrEv := range getDerivedEvents(events, argsFlagsWithParams[utils.MetaAttributes].Has(utils.MetaDerivedReply)) {
			rplyAttr, err := sS.processAttributes(cgrEv, attrIDs, false)
			if err != nil {
				if err.Error() != utils.ErrNotFound.Error() {
					return utils.NewErrAttributeS(err)
				}
			} else {
				*cgrEv = *rplyAttr.CGREvent
				rply.Attributes[runID] = &rplyAttr
			}
		}
		args.CGREvent = events[utils.MetaRaw]
	}

	// get routes if required
	if argsFlagsWithParams.GetBool(utils.MetaRoutes) {
		rply.Routes = make(map[string]*engine.SortedRoutes)
		// check in case we have options for suppliers
		flags := argsFlagsWithParams[utils.MetaRoutes]
		ignoreErrors := flags.Has(utils.MetaIgnoreErrors)
		var maxCost string
		if flags.Has(utils.MetaEventCost) {
			maxCost = utils.MetaEventCost
		} else {
			maxCost = flags.ParamValue(utils.MetaMaxCost)
		}
		for runID, cgrEv := range getDerivedEvents(events, flags.Has(utils.MetaDerivedReply)) {
			routesReply, err := sS.getRoutes(cgrEv.Clone(), args.Paginator, ignoreErrors, maxCost, false)
			if err != nil {
				return err
			}
			if routesReply.SortedRoutes != nil {
				rply.Routes[runID] = &routesReply
			}
		}
	}

	// process thresholds if required
	if argsFlagsWithParams.GetBool(utils.MetaThresholds) {
		rply.ThresholdIDs = make(map[string][]string)
		thIDs := argsFlagsWithParams.ParamsSlice(utils.MetaThresholds, utils.MetaIDs)
		for runID, cgrEv := range getDerivedEvents(events, argsFlagsWithParams[utils.MetaThresholds].Has(utils.MetaDerivedReply)) {
			tIDs, err := sS.processThreshold(cgrEv, thIDs, true)
			if err != nil && err.Error() != utils.ErrNotFound.Error() {
				if blockError {
					return utils.NewErrThresholdS(err)
				}
				utils.Logger.Warning(
					fmt.Sprintf("<%s> error: %s processing event %+v for RunID <%s>  with ThresholdS.",
						utils.SessionS, err.Error(), cgrEv, runID))
				withErrors = true
			}
			rply.ThresholdIDs[runID] = tIDs
		}
	}

	// process stats if required
	if argsFlagsWithParams.GetBool(utils.MetaStats) {
		rply.StatQueueIDs = make(map[string][]string)
		stIDs := argsFlagsWithParams.ParamsSlice(utils.MetaStats, utils.MetaIDs)
		for runID, cgrEv := range getDerivedEvents(events, argsFlagsWithParams[utils.MetaStats].Has(utils.MetaDerivedReply)) {
			sIDs, err := sS.processStats(cgrEv, stIDs, true)
			if err != nil &&
				err.Error() != utils.ErrNotFound.Error() {
				if blockError {
					return utils.NewErrStatS(err)
				}
				utils.Logger.Warning(
					fmt.Sprintf("<%s> error: %s processing event %+v for RunID <%s> with StatS.",
						utils.SessionS, err.Error(), cgrEv, runID))
				withErrors = true
			}
			rply.StatQueueIDs[runID] = sIDs
		}
	}

	if argsFlagsWithParams.GetBool(utils.MetaSTIRAuthenticate) {
		for _, cgrEv := range getDerivedEvents(events, argsFlagsWithParams[utils.MetaSTIRAuthenticate].Has(utils.MetaDerivedReply)) {
			ev := engine.MapEvent(cgrEv.Event)
			opts := engine.MapEvent(cgrEv.Opts)
			attest := sS.cgrCfg.SessionSCfg().STIRCfg.AllowedAttest
			if uattest := opts.GetStringIgnoreErrors(utils.OptsStirATest); uattest != utils.EmptyString {
				attest = utils.NewStringSet(strings.Split(uattest, utils.InfieldSep))
			}
			var stirMaxDur time.Duration
			if stirMaxDur, err = opts.GetDuration(utils.OptsStirPayloadMaxDuration); err != nil {
				stirMaxDur = sS.cgrCfg.SessionSCfg().STIRCfg.PayloadMaxduration
			}
			if err = AuthStirShaken(opts.GetStringIgnoreErrors(utils.OptsStirIdentity),
				utils.FirstNonEmpty(opts.GetStringIgnoreErrors(utils.OptsStirOriginatorTn), ev.GetStringIgnoreErrors(utils.AccountField)),
				opts.GetStringIgnoreErrors(utils.OptsStirOriginatorURI),
				utils.FirstNonEmpty(opts.GetStringIgnoreErrors(utils.OptsStirDestinationTn), ev.GetStringIgnoreErrors(utils.Destination)),
				opts.GetStringIgnoreErrors(utils.OptsStirDestinationURI),
				attest, stirMaxDur); err != nil {
				return utils.NewSTIRError(err.Error())
			}
		}
	} else if argsFlagsWithParams.GetBool(utils.MetaSTIRInitiate) {
		rply.STIRIdentity = make(map[string]string)
		for runID, cgrEv := range getDerivedEvents(events, argsFlagsWithParams[utils.MetaSTIRInitiate].Has(utils.MetaDerivedReply)) {
			ev := engine.MapEvent(cgrEv.Event)
			opts := engine.MapEvent(cgrEv.Opts)
			attest := sS.cgrCfg.SessionSCfg().STIRCfg.DefaultAttest
			if uattest := opts.GetStringIgnoreErrors(utils.OptsStirATest); uattest != utils.EmptyString {
				attest = uattest
			}

			destURI := opts.GetStringIgnoreErrors(utils.OptsStirDestinationTn)
			destTn := utils.FirstNonEmpty(opts.GetStringIgnoreErrors(utils.OptsStirDestinationTn), ev.GetStringIgnoreErrors(utils.Destination))

			dest := utils.NewPASSporTDestinationsIdentity(strings.Split(destTn, utils.InfieldSep), strings.Split(destURI, utils.InfieldSep))

			var orig *utils.PASSporTOriginsIdentity
			if origURI := opts.GetStringIgnoreErrors(utils.OptsStirOriginatorURI); origURI != utils.EmptyString {
				orig = utils.NewPASSporTOriginsIdentity(utils.EmptyString, origURI)
			} else {
				orig = utils.NewPASSporTOriginsIdentity(
					utils.FirstNonEmpty(opts.GetStringIgnoreErrors(utils.OptsStirOriginatorTn),
						ev.GetStringIgnoreErrors(utils.AccountField)),
					utils.EmptyString)
			}
			pubkeyPath := utils.FirstNonEmpty(opts.GetStringIgnoreErrors(utils.OptsStirPublicKeyPath), sS.cgrCfg.SessionSCfg().STIRCfg.PublicKeyPath)
			prvkeyPath := utils.FirstNonEmpty(opts.GetStringIgnoreErrors(utils.OptsStirPrivateKeyPath), sS.cgrCfg.SessionSCfg().STIRCfg.PrivateKeyPath)

			payload := utils.NewPASSporTPayload(attest, cgrEv.ID, *dest, *orig)
			header := utils.NewPASSporTHeader(pubkeyPath)
			if rply.STIRIdentity[runID], err = NewSTIRIdentity(header, payload, prvkeyPath, sS.cgrCfg.GeneralCfg().ReplyTimeout); err != nil {
				return utils.NewSTIRError(err.Error())
			}
		}
	}

	// check for *resources
	if argsFlagsWithParams.GetBool(utils.MetaResources) {
		if len(sS.cgrCfg.SessionSCfg().ResSConns) == 0 {
			return utils.NewErrNotConnected(utils.ResourceS)
		}
		rply.ResourceAllocation = make(map[string]string)
		if resOpt := argsFlagsWithParams[utils.MetaResources]; len(resOpt) != 0 {
			for runID, cgrEv := range getDerivedEvents(events, resOpt.Has(utils.MetaDerivedReply)) {
				originID := engine.MapEvent(cgrEv.Event).GetStringIgnoreErrors(utils.OriginID)
				if originID == "" {
					return utils.NewErrMandatoryIeMissing(utils.OriginID)
				}

				attrRU := &utils.ArgRSv1ResourceUsage{
					CGREvent: cgrEv,
					UsageID:  originID,
					Units:    1,
				}
				attrRU.SetCloneable(true)
				var resMessage string
				// check what we need to do for resources (*authorization/*allocation)
				//check for subflags and convert them into utils.FlagsWithParams
				switch {
				case resOpt.Has(utils.MetaAuthorize):
					if err = sS.connMgr.Call(sS.cgrCfg.SessionSCfg().ResSConns, nil, utils.ResourceSv1AuthorizeResources,
						attrRU, &resMessage); err != nil {
						if blockError {
							return utils.NewErrResourceS(err)
						}
						utils.Logger.Warning(
							fmt.Sprintf("<%s> error: <%s> processing event %+v for RunID <%s>  with ResourceS.",
								utils.SessionS, err.Error(), cgrEv, runID))
						withErrors = true
					}
				case resOpt.Has(utils.MetaAllocate):
					if err = sS.connMgr.Call(sS.cgrCfg.SessionSCfg().ResSConns, nil, utils.ResourceSv1AllocateResources,
						attrRU, &resMessage); err != nil {
						if blockError {
							return utils.NewErrResourceS(err)
						}
						utils.Logger.Warning(
							fmt.Sprintf("<%s> error: <%s> processing event %+v for RunID <%s>  with ResourceS.",
								utils.SessionS, err.Error(), cgrEv, runID))
						withErrors = true
					}
				case resOpt.Has(utils.MetaRelease):
					if err = sS.connMgr.Call(sS.cgrCfg.SessionSCfg().ResSConns, nil, utils.ResourceSv1ReleaseResources,
						attrRU, &resMessage); err != nil {
						if blockError {
							return utils.NewErrResourceS(err)
						}
						utils.Logger.Warning(
							fmt.Sprintf("<%s> error: <%s> processing event %+v for RunID <%s>  with ResourceS.",
								utils.SessionS, err.Error(), cgrEv, runID))
						withErrors = true
					}
				}
				rply.ResourceAllocation[runID] = resMessage
			}
		}
	}

	// check what we need to do for RALs (*authorize/*initiate/*update/*terminate)
	dbtItvl := sS.cgrCfg.SessionSCfg().DebitInterval
	if argsFlagsWithParams.GetBool(utils.MetaRALs) {
		if ralsOpts := argsFlagsWithParams[utils.MetaRALs]; len(ralsOpts) != 0 {
			//check for subflags and convert them into utils.FlagsWithParams
			// check for *cost
			if ralsOpts.Has(utils.MetaCost) {
				rply.Cost = make(map[string]float64)
				for runID, cgrEv := range getDerivedEvents(events, ralsOpts.Has(utils.MetaDerivedReply)) {
					ev := engine.MapEvent(cgrEv.Event)
					//compose the CallDescriptor with Args
					startTime := ev.GetTimeIgnoreErrors(utils.AnswerTime,
						sS.cgrCfg.GeneralCfg().DefaultTimezone)
					if startTime.IsZero() { // AnswerTime not parsable, try SetupTime
						startTime = ev.GetTimeIgnoreErrors(utils.SetupTime,
							sS.cgrCfg.GeneralCfg().DefaultTimezone)
					}
					category := ev.GetStringIgnoreErrors(utils.Category)
					if len(category) == 0 {
						category = sS.cgrCfg.GeneralCfg().DefaultCategory
					}
					subject := ev.GetStringIgnoreErrors(utils.Subject)
					if len(subject) == 0 {
						subject = ev.GetStringIgnoreErrors(utils.AccountField)
					}

					cd := &engine.CallDescriptor{
						CgrID:         cgrEv.ID,
						RunID:         ev.GetStringIgnoreErrors(utils.RunID),
						ToR:           ev.GetStringIgnoreErrors(utils.ToR),
						Tenant:        cgrEv.Tenant,
						Category:      category,
						Subject:       subject,
						Account:       ev.GetStringIgnoreErrors(utils.AccountField),
						Destination:   ev.GetStringIgnoreErrors(utils.Destination),
						TimeStart:     startTime,
						TimeEnd:       startTime.Add(ev.GetDurationIgnoreErrors(utils.Usage)),
						ForceDuration: ralsOpts.Has(utils.MetaFD),
					}
					var cc engine.CallCost
					if err = sS.connMgr.Call(sS.cgrCfg.SessionSCfg().RALsConns, nil,
						utils.ResponderGetCost,
						&engine.CallDescriptorWithOpts{
							CallDescriptor: cd,
							Opts:           cgrEv.Opts,
						}, &cc); err != nil {
						return err
					}
					rply.Cost[runID] = cc.Cost
				}
			}
			opts := engine.MapEvent(args.Opts)
			ev := engine.MapEvent(args.CGREvent.Event)
			originID := ev.GetStringIgnoreErrors(utils.OriginID)
			switch {
			//check for auth session
			case ralsOpts.Has(utils.MetaAuthorize):
				var sRunsMaxUsage map[string]time.Duration
				if sRunsMaxUsage, err = sS.authEvent(args.CGREvent, ralsOpts.Has(utils.MetaFD)); err != nil {
					return err
				}
				rply.MaxUsage = getDerivedMaxUsage(sRunsMaxUsage, ralsOpts.Has(utils.MetaDerivedReply))
			// check for init session
			case ralsOpts.Has(utils.MetaInitiate):
				if opts.HasField(utils.OptsDebitInterval) { // dynamic DebitInterval via CGRDebitInterval
					if dbtItvl, err = opts.GetDuration(utils.OptsDebitInterval); err != nil {
						return utils.NewErrRALs(err)
					}
				}
				s, err := sS.initSession(args.CGREvent, sS.biJClntID(clnt), originID, dbtItvl, false,
					ralsOpts.Has(utils.MetaFD))
				if err != nil {
					return err
				}
				sRunsMaxUsage := make(map[string]time.Duration)
				s.RLock()
				isPrepaid := s.debitStop != nil
				s.RUnlock()
				if isPrepaid { //active debit
					for _, sr := range s.SRuns {
						sRunsMaxUsage[sr.CD.RunID] = sS.cgrCfg.GeneralCfg().MaxCallDuration
					}
				} else if sRunsMaxUsage, err = sS.updateSession(s, nil, args.Opts, false); err != nil {
					return utils.NewErrRALs(err)
				}
				rply.MaxUsage = getDerivedMaxUsage(sRunsMaxUsage, ralsOpts.Has(utils.MetaDerivedReply))
				//check for update session
			case ralsOpts.Has(utils.MetaUpdate):
				if opts.HasField(utils.OptsDebitInterval) { // dynamic DebitInterval via CGRDebitInterval
					if dbtItvl, err = opts.GetDuration(utils.OptsDebitInterval); err != nil {
						return utils.NewErrRALs(err)
					}
				}
				s := sS.getRelocateSession(GetSetCGRID(ev),
					ev.GetStringIgnoreErrors(utils.InitialOriginID),
					ev.GetStringIgnoreErrors(utils.OriginID),
					ev.GetStringIgnoreErrors(utils.OriginHost))
				if s == nil {
					if s, err = sS.initSession(args.CGREvent, sS.biJClntID(clnt), ev.GetStringIgnoreErrors(utils.OriginID),
						dbtItvl, false, ralsOpts.Has(utils.MetaFD)); err != nil {
						return err
					}
				}
				var sRunsMaxUsage map[string]time.Duration
				if sRunsMaxUsage, err = sS.updateSession(s, ev, args.Opts, false); err != nil {
					return utils.NewErrRALs(err)
				}
				rply.MaxUsage = getDerivedMaxUsage(sRunsMaxUsage, ralsOpts.Has(utils.MetaDerivedReply))
				// check for terminate session
			case ralsOpts.Has(utils.MetaTerminate):
				if opts.HasField(utils.OptsDebitInterval) { // dynamic DebitInterval via CGRDebitInterval
					if dbtItvl, err = opts.GetDuration(utils.OptsDebitInterval); err != nil {
						return utils.NewErrRALs(err)
					}
				}
				s := sS.getRelocateSession(GetSetCGRID(ev),
					ev.GetStringIgnoreErrors(utils.InitialOriginID),
					ev.GetStringIgnoreErrors(utils.OriginID),
					ev.GetStringIgnoreErrors(utils.OriginHost))
				if s == nil {
					if s, err = sS.initSession(args.CGREvent, sS.biJClntID(clnt), ev.GetStringIgnoreErrors(utils.OriginID),
						dbtItvl, false, ralsOpts.Has(utils.MetaFD)); err != nil {
						return err
					}
				} else {
					s.Lock()
					s.chargeable = opts.GetBoolOrDefault(utils.OptsChargeable, true)
					s.Unlock()
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

	if argsFlagsWithParams.GetBool(utils.MetaCDRs) {
		if len(sS.cgrCfg.SessionSCfg().CDRsConns) == 0 {
			return utils.NewErrNotConnected(utils.CDRs)
		}
		flgs := argsFlagsWithParams[utils.MetaCDRs].SliceFlags()
		var cdrRply string
		for _, cgrEv := range getDerivedEvents(events, argsFlagsWithParams[utils.MetaCDRs].Has(utils.MetaDerivedReply)) {
			if err := sS.processCDR(cgrEv, flgs, &cdrRply, false); err != nil {
				if blockError {
					return utils.NewErrCDRS(err)
				}
				utils.Logger.Warning(
					fmt.Sprintf("<%s> error: <%s> processing event %+v with CDRs.",
						utils.SessionS, err.Error(), cgrEv))
				withErrors = true
			}
		}
	}
	if withErrors {
		err = utils.ErrPartiallyExecuted
	}
	return
}

// V1GetCostReply is the reply for the GetCost API
type V1GetCostReply struct {
	Attributes *engine.AttrSProcessEventReply
	EventCost  *engine.EventCost
}

// BiRPCv1GetCost processes one event with the right subsystems based on arguments received
func (sS *SessionS) BiRPCv1GetCost(clnt rpcclient.ClientConnector,
	args *V1ProcessEventArgs, rply *V1GetCostReply) (err error) {
	if args.CGREvent == nil {
		return utils.NewErrMandatoryIeMissing(utils.CGREventString)
	}
	var withErrors bool
	if args.CGREvent.ID == "" {
		args.CGREvent.ID = utils.GenUUID()
	}
	if args.CGREvent.Tenant == "" {
		args.CGREvent.Tenant = sS.cgrCfg.GeneralCfg().DefaultTenant
	}

	// RPC caching
	if sS.cgrCfg.CacheCfg().Partitions[utils.CacheRPCResponses].Limit != 0 {
		cacheKey := utils.ConcatenatedKey(utils.SessionSv1GetCost, args.CGREvent.ID)
		refID := guardian.Guardian.GuardIDs("",
			sS.cgrCfg.GeneralCfg().LockingTimeout, cacheKey) // RPC caching needs to be atomic
		defer guardian.Guardian.UnguardIDs(refID)

		if itm, has := engine.Cache.Get(utils.CacheRPCResponses, cacheKey); has {
			cachedResp := itm.(*utils.CachedRPCResponse)
			if cachedResp.Error == nil {
				*rply = *cachedResp.Result.(*V1GetCostReply)
			}
			return cachedResp.Error
		}
		defer engine.Cache.Set(utils.CacheRPCResponses, cacheKey,
			&utils.CachedRPCResponse{Result: rply, Error: err},
			nil, true, utils.NonTransactional)
	}
	// end of RPC caching

	//convert from Flags []string to utils.FlagsWithParams
	argsFlagsWithParams := utils.FlagsWithParamsFromSlice(args.Flags)
	// check for *attribute
	if argsFlagsWithParams.Has(utils.MetaAttributes) {
		rplyAttr, err := sS.processAttributes(args.CGREvent,
			argsFlagsWithParams.ParamsSlice(utils.MetaAttributes, utils.MetaIDs), false)
		if err == nil {
			args.CGREvent = rplyAttr.CGREvent
			rply.Attributes = &rplyAttr
		} else if err.Error() != utils.ErrNotFound.Error() {
			return utils.NewErrAttributeS(err)
		}
	}
	//compose the CallDescriptor with Args
	me := engine.MapEvent(args.CGREvent.Event)
	startTime := me.GetTimeIgnoreErrors(utils.AnswerTime,
		sS.cgrCfg.GeneralCfg().DefaultTimezone)
	if startTime.IsZero() { // AnswerTime not parsable, try SetupTime
		startTime = me.GetTimeIgnoreErrors(utils.SetupTime,
			sS.cgrCfg.GeneralCfg().DefaultTimezone)
	}
	category := me.GetStringIgnoreErrors(utils.Category)
	if len(category) == 0 {
		category = sS.cgrCfg.GeneralCfg().DefaultCategory
	}
	subject := me.GetStringIgnoreErrors(utils.Subject)
	if len(subject) == 0 {
		subject = me.GetStringIgnoreErrors(utils.AccountField)
	}

	cd := &engine.CallDescriptor{
		CgrID:       args.CGREvent.ID,
		RunID:       me.GetStringIgnoreErrors(utils.RunID),
		ToR:         me.GetStringIgnoreErrors(utils.ToR),
		Tenant:      args.CGREvent.Tenant,
		Category:    category,
		Subject:     subject,
		Account:     me.GetStringIgnoreErrors(utils.AccountField),
		Destination: me.GetStringIgnoreErrors(utils.Destination),
		TimeStart:   startTime,
		TimeEnd:     startTime.Add(me.GetDurationIgnoreErrors(utils.Usage)),
	}
	var cc engine.CallCost
	if err = sS.connMgr.Call(sS.cgrCfg.SessionSCfg().RALsConns, nil,
		utils.ResponderGetCost,
		&engine.CallDescriptorWithOpts{
			CallDescriptor: cd,
			Opts:           args.Opts,
		}, &cc); err != nil {
		return
	}
	ec := engine.NewEventCostFromCallCost(&cc, args.CGREvent.ID, me.GetStringIgnoreErrors(utils.RunID))
	ec.Compute()
	rply.EventCost = ec
	if withErrors {
		err = utils.ErrPartiallyExecuted
	}
	return
}

// BiRPCv1SyncSessions will sync sessions on demand
func (sS *SessionS) BiRPCv1SyncSessions(clnt rpcclient.ClientConnector,
	ignParam *utils.TenantWithOpts, reply *string) error {
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
		ss[0].Lock()
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
	sIDs *utils.SessionIDsWithArgsDispatcher, reply *string) (err error) {
	if len(sIDs.IDs) == 0 {
		sS.pSsMux.RLock()
		i := 0
		sIDs.IDs = make([]string, len(sS.pSessions))
		for sID := range sS.pSessions {
			sIDs.IDs[i] = sID
			i++
		}
		sS.pSsMux.RUnlock()
	}
	for _, sID := range sIDs.IDs {
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
	sIDs *utils.SessionIDsWithArgsDispatcher, reply *string) (err error) {
	if len(sIDs.IDs) == 0 {
		sS.aSsMux.RLock()
		i := 0
		sIDs.IDs = make([]string, len(sS.aSessions))
		for sID := range sS.aSessions {
			sIDs.IDs[i] = sID
			i++
		}
		sS.aSsMux.RUnlock()
	}
	for _, sID := range sIDs.IDs {
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

func (sS *SessionS) processCDR(cgrEv *utils.CGREvent, flags []string, rply *string, clnb bool) (err error) {
	ev := engine.MapEvent(cgrEv.Event)
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
		argsProc := &engine.ArgV1ProcessEvent{
			Flags:    flags,
			CGREvent: *cgrEv,
		}
		argsProc.SetCloneable(clnb)
		return sS.connMgr.Call(sS.cgrCfg.SessionSCfg().CDRsConns, nil, utils.CDRsV1ProcessEvent,
			argsProc, rply)
	}

	// Use previously stored Session to generate CDRs
	s.updateSRuns(ev, sS.cgrCfg.SessionSCfg().AlterableFields)
	// create one CGREvent for each session run
	var withErrors bool
	for _, cgrEv := range s.asCGREvents() {
		argsProc := &engine.ArgV1ProcessEvent{
			Flags: []string{fmt.Sprintf("%s:false", utils.MetaChargers),
				fmt.Sprintf("%s:false", utils.MetaAttributes)},
			CGREvent: *cgrEv,
		}
		argsProc.SetCloneable(clnb)
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

// processThreshold will receive the event and send it to ThresholdS to be processed
func (sS *SessionS) processThreshold(cgrEv *utils.CGREvent, thIDs []string, clnb bool) (tIDs []string, err error) {
	if len(sS.cgrCfg.SessionSCfg().ThreshSConns) == 0 {
		return tIDs, utils.NewErrNotConnected(utils.ThresholdS)
	}
	thEv := &engine.ThresholdsArgsProcessEvent{
		CGREvent: cgrEv,
	}
	// check if we have thresholdIDs
	if len(thIDs) != 0 {
		thEv.ThresholdIDs = thIDs
	}
	thEv.SetCloneable(clnb)
	//initialize the returned variable
	err = sS.connMgr.Call(sS.cgrCfg.SessionSCfg().ThreshSConns, nil, utils.ThresholdSv1ProcessEvent, thEv, &tIDs)
	return
}

// processStats will receive the event and send it to StatS to be processed
func (sS *SessionS) processStats(cgrEv *utils.CGREvent, stsIDs []string, clnb bool) (sIDs []string, err error) {
	if len(sS.cgrCfg.SessionSCfg().StatSConns) == 0 {
		return sIDs, utils.NewErrNotConnected(utils.StatS)
	}

	statArgs := &engine.StatsArgsProcessEvent{
		CGREvent: cgrEv,
	}
	// check in case we have StatIDs inside flags
	if len(stsIDs) != 0 {
		statArgs.StatIDs = stsIDs
	}
	statArgs.SetCloneable(clnb)
	//initialize the returned variable
	err = sS.connMgr.Call(sS.cgrCfg.SessionSCfg().StatSConns, nil, utils.StatSv1ProcessEvent, statArgs, &sIDs)
	return
}

// getRoutes will receive the event and send it to SupplierS to find the suppliers
func (sS *SessionS) getRoutes(cgrEv *utils.CGREvent, pag utils.Paginator, ignoreErrors bool,
	maxCost string, clnb bool) (routesReply engine.SortedRoutes, err error) {
	if len(sS.cgrCfg.SessionSCfg().RouteSConns) == 0 {
		return routesReply, utils.NewErrNotConnected(utils.RouteS)
	}
	if acd, has := cgrEv.Event[utils.ACD]; has {
		cgrEv.Event[utils.Usage] = acd
	}
	sArgs := &engine.ArgsGetRoutes{
		CGREvent:     cgrEv,
		Paginator:    pag,
		IgnoreErrors: ignoreErrors,
		MaxCost:      maxCost,
	}
	sArgs.SetCloneable(clnb)
	if err = sS.connMgr.Call(sS.cgrCfg.SessionSCfg().RouteSConns, nil, utils.RouteSv1GetRoutes,
		sArgs, &routesReply); err != nil {
		return routesReply, utils.NewErrRouteS(err)
	}
	return
}

// processAttributes will receive the event and send it to AttributeS to be processed
func (sS *SessionS) processAttributes(cgrEv *utils.CGREvent, attrIDs []string,
	clnb bool) (rplyEv engine.AttrSProcessEventReply, err error) {
	if len(sS.cgrCfg.SessionSCfg().AttrSConns) == 0 {
		return rplyEv, utils.NewErrNotConnected(utils.AttributeS)
	}
	if cgrEv.Opts == nil {
		cgrEv.Opts = make(engine.MapEvent)
	}
	cgrEv.Opts[utils.Subsys] = utils.MetaSessionS
	var processRuns *int
	if val, has := cgrEv.Opts[utils.OptsAttributesProcessRuns]; has {
		if v, err := utils.IfaceAsTInt64(val); err == nil {
			processRuns = utils.IntPointer(int(v))
		}
	}
	attrArgs := &engine.AttrArgsProcessEvent{
		Context: utils.StringPointer(utils.FirstNonEmpty(
			utils.IfaceAsString(cgrEv.Opts[utils.OptsContext]),
			utils.MetaSessionS)),
		CGREvent:     cgrEv,
		AttributeIDs: attrIDs,
		ProcessRuns:  processRuns,
	}
	attrArgs.SetCloneable(clnb)
	err = sS.connMgr.Call(sS.cgrCfg.SessionSCfg().AttrSConns, nil, utils.AttributeSv1ProcessEvent,
		attrArgs, &rplyEv)
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
				Event: ev,
			},
		},
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
				Event: ev,
			},
		},
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
				Event: ev,
			},
		},
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
				Event: ev,
			},
		},
		rply)
}

// BiRPCV1ProcessCDR should send the CDR to CDRS
// DEPRECATED, it will be removed in future versions
// Kept for compatibility with OpenSIPS 2.3
func (sS *SessionS) BiRPCV1ProcessCDR(clnt rpcclient.ClientConnector,
	ev engine.MapEvent, rply *string) (err error) {
	return sS.BiRPCv1ProcessCDR(
		clnt,
		&utils.CGREvent{
			Tenant: utils.FirstNonEmpty(
				ev.GetStringIgnoreErrors(utils.Tenant),
				sS.cgrCfg.GeneralCfg().DefaultTenant),
			ID:    utils.UUIDSha1Prefix(),
			Event: ev},
		rply)
}

func (sS *SessionS) sendRar(s *Session) (err error) {
	clnt := sS.biJClnt(s.ClientConnID)
	if clnt == nil {
		return fmt.Errorf("calling %s requires bidirectional JSON connection, connID: <%s>",
			utils.SessionSv1ReAuthorize, s.ClientConnID)
	}
	var originID string
	if originID, err = s.EventStart.GetString(utils.OriginID); err != nil {
		return
	}
	var rply string
	if err = clnt.conn.Call(utils.SessionSv1ReAuthorize, originID, &rply); err == utils.ErrNotImplemented {
		err = nil
	}
	return
}

// BiRPCv1ReAuthorize sends a RAR for the matching sessions
func (sS *SessionS) BiRPCv1ReAuthorize(clnt rpcclient.ClientConnector,
	args *utils.SessionFilter, reply *string) (err error) {
	if args == nil { //protection in case on nil
		args = &utils.SessionFilter{}
	}
	aSs := sS.filterSessions(args, false)
	if len(aSs) == 0 {
		return utils.ErrNotFound
	}
	cache := utils.NewStringSet(nil)
	for _, as := range aSs {
		if cache.Has(as.CGRID) {
			continue
		}
		cache.Add(as.CGRID)
		ss := sS.getSessions(as.CGRID, false)
		if len(ss) == 0 {
			continue
		}
		if errTerm := sS.sendRar(ss[0]); errTerm != nil {
			utils.Logger.Warning(
				fmt.Sprintf(
					"<%s> failed sending RAR for session with id: <%s>, err: <%s>",
					utils.SessionS, ss[0].cgrID(), errTerm.Error()))
			err = utils.ErrPartiallyExecuted
		}
	}
	if err != nil {
		return
	}
	*reply = utils.OK
	return
}

// BiRPCv1DisconnectPeer sends a DPR for the given OriginHost and OriginRealm
func (sS *SessionS) BiRPCv1DisconnectPeer(clnt rpcclient.ClientConnector,
	args *utils.DPRArgs, reply *string) (err error) {
	hasErrors := false
	clients := make(map[string]*biJClient)
	sS.biJMux.RLock()
	for ID, clnt := range sS.biJIDs {
		clients[ID] = clnt
	}
	sS.biJMux.RUnlock()
	for ID, clnt := range clients {
		if err = clnt.conn.Call(utils.SessionSv1DisconnectPeer, args, reply); err != nil && err != utils.ErrNotImplemented {
			utils.Logger.Warning(
				fmt.Sprintf(
					"<%s> failed sending DPR for connection with id: <%s>, err: <%s>",
					utils.SessionS, ID, err))
			hasErrors = true
		}
	}
	if hasErrors {
		return utils.ErrPartiallyExecuted
	}
	*reply = utils.OK
	return nil
}

// BiRPCv1STIRAuthenticate the API for STIR checking
func (sS *SessionS) BiRPCv1STIRAuthenticate(clnt rpcclient.ClientConnector,
	args *V1STIRAuthenticateArgs, reply *string) (err error) {
	attest := sS.cgrCfg.SessionSCfg().STIRCfg.AllowedAttest
	if len(args.Attest) != 0 {
		attest = utils.NewStringSet(args.Attest)
	}
	stirMaxDur := sS.cgrCfg.SessionSCfg().STIRCfg.PayloadMaxduration
	if args.PayloadMaxDuration != utils.EmptyString {
		if stirMaxDur, err = utils.ParseDurationWithNanosecs(args.PayloadMaxDuration); err != nil {
			return
		}
	}
	if err = AuthStirShaken(args.Identity, args.OriginatorTn, args.OriginatorURI,
		args.DestinationTn, args.DestinationURI, attest, stirMaxDur); err != nil {
		return utils.NewSTIRError(err.Error())
	}
	*reply = utils.OK
	return
}

// BiRPCv1STIRIdentity the API for STIR header creation
func (sS *SessionS) BiRPCv1STIRIdentity(clnt rpcclient.ClientConnector,
	args *V1STIRIdentityArgs, identity *string) (err error) {
	if args.Payload.ATTest == utils.EmptyString {
		args.Payload.ATTest = sS.cgrCfg.SessionSCfg().STIRCfg.DefaultAttest
	}
	if args.OverwriteIAT {
		args.Payload.IAT = time.Now().Unix()
	}
	if *identity, err = NewSTIRIdentity(
		utils.NewPASSporTHeader(utils.FirstNonEmpty(args.PublicKeyPath,
			sS.cgrCfg.SessionSCfg().STIRCfg.PublicKeyPath)),
		args.Payload, utils.FirstNonEmpty(args.PrivateKeyPath,
			sS.cgrCfg.SessionSCfg().STIRCfg.PrivateKeyPath),
		sS.cgrCfg.GeneralCfg().ReplyTimeout); err != nil {
		return utils.NewSTIRError(err.Error())
	}
	return
}
