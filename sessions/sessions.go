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

const (
	MaxTimespansInCost = 50
	MaxSessionTTL      = 10000 // maximum session TTL in miliseconds
)

var (
	ErrActiveSession    = errors.New("ACTIVE_SESSION")
	ErrForcedDisconnect = errors.New("FORCED_DISCONNECT")
	debug               bool
)

// NewSReplConns initiates the connections configured for session replication
func NewSReplConns(conns []*config.RemoteHost, reconnects int,
	connTimeout, replyTimeout time.Duration) (sReplConns []*SReplConn, err error) {
	sReplConns = make([]*SReplConn, len(conns))
	for i, replConnCfg := range conns {
		if replCon, err := rpcclient.NewRpcClient("tcp", replConnCfg.Address, replConnCfg.TLS, "", "", "", 0, reconnects,
			connTimeout, replyTimeout, replConnCfg.Transport[1:], nil, true); err != nil {
			return nil, err
		} else {
			sReplConns[i] = &SReplConn{Connection: replCon, Synchronous: replConnCfg.Synchronous}
		}
	}
	return
}

// SReplConn represents one connection to a passive node where we will replicate session data
type SReplConn struct {
	Connection  rpcclient.RpcClientConnection
	Synchronous bool
}

// NewSessionS constructs  a new SessionS instance
func NewSessionS(cgrCfg *config.CGRConfig, ralS, resS, thdS,
	statS, splS, attrS, cdrS, chargerS rpcclient.RpcClientConnection,
	sReplConns []*SReplConn, dm *engine.DataManager, tmz string) *SessionS {
	cgrCfg.SessionSCfg().SessionIndexes[utils.OriginID] = true // Make sure we have indexing for OriginID since it is a requirement on prefix searching
	if chargerS != nil && reflect.ValueOf(chargerS).IsNil() {
		chargerS = nil
	}
	if ralS != nil && reflect.ValueOf(ralS).IsNil() {
		ralS = nil
	}
	if resS != nil && reflect.ValueOf(resS).IsNil() {
		resS = nil
	}
	if thdS != nil && reflect.ValueOf(thdS).IsNil() {
		thdS = nil
	}
	if statS != nil && reflect.ValueOf(statS).IsNil() {
		statS = nil
	}
	if splS != nil && reflect.ValueOf(splS).IsNil() {
		splS = nil
	}
	if attrS != nil && reflect.ValueOf(attrS).IsNil() {
		attrS = nil
	}
	if cdrS != nil && reflect.ValueOf(cdrS).IsNil() {
		cdrS = nil
	}
	return &SessionS{
		cgrCfg:        cgrCfg,
		chargerS:      chargerS,
		ralS:          ralS,
		resS:          resS,
		thdS:          thdS,
		statS:         statS,
		splS:          splS,
		attrS:         attrS,
		cdrS:          cdrS,
		sReplConns:    sReplConns,
		biJClnts:      make(map[rpcclient.RpcClientConnection]string),
		biJIDs:        make(map[string]*biJClient),
		aSessions:     make(map[string]*Session),
		aSessionsIdx:  make(map[string]map[string]map[string]utils.StringMap),
		aSessionsRIdx: make(map[string][]*riFieldNameVal),
		pSessions:     make(map[string]*Session),
		pSessionsIdx:  make(map[string]map[string]map[string]utils.StringMap),
		pSessionsRIdx: make(map[string][]*riFieldNameVal),
		dm:            dm,
	}
}

// biJClient contains info we need to reach back a bidirectional json client
type biJClient struct {
	conn  rpcclient.RpcClientConnection // connection towards BiJ client
	proto float64                       // client protocol version
}

// SessionS represents the session service
type SessionS struct {
	cgrCfg *config.CGRConfig // Separate from smCfg since there can be multiple

	chargerS rpcclient.RpcClientConnection
	ralS     rpcclient.RpcClientConnection // RALs connections
	resS     rpcclient.RpcClientConnection // ResourceS connections
	thdS     rpcclient.RpcClientConnection // ThresholdS connections
	statS    rpcclient.RpcClientConnection // StatS connections
	splS     rpcclient.RpcClientConnection // SupplierS connections
	attrS    rpcclient.RpcClientConnection // AttributeS connections
	cdrS     rpcclient.RpcClientConnection // CDR server connections

	sReplConns []*SReplConn // list of connections where we will replicate our session data

	biJMux   sync.RWMutex                             // mux protecting BI-JSON connections
	biJClnts map[rpcclient.RpcClientConnection]string // index BiJSONConnection so we can sync them later
	biJIDs   map[string]*biJClient                    // identifiers of bidirectional JSON conns, used to call RPC based on connIDs

	aSsMux    sync.RWMutex        // protects aSessions
	aSessions map[string]*Session // group sessions per sessionId, multiple runs based on derived charging

	aSIMux        sync.RWMutex                                     // protects aSessionsIdx
	aSessionsIdx  map[string]map[string]map[string]utils.StringMap // map[fieldName]map[fieldValue][cgrID]utils.StringMap[runID]
	aSessionsRIdx map[string][]*riFieldNameVal                     // reverse indexes for active sessions, used on remove

	pSsMux    sync.RWMutex        // protects pSessions
	pSessions map[string]*Session // group passive sessions based on cgrID

	pSIMux        sync.RWMutex                                     // protects pSessionsIdx
	pSessionsIdx  map[string]map[string]map[string]utils.StringMap // map[fieldName]map[fieldValue][cgrID]utils.StringMap[runID]
	pSessionsRIdx map[string][]*riFieldNameVal                     // reverse indexes for passive sessions, used on remove

	dm *engine.DataManager
}

// ListenAndServe starts the service and binds it to the listen loop
func (sS *SessionS) ListenAndServe(exitChan chan bool) (err error) {
	if sS.cgrCfg.SessionSCfg().ChannelSyncInterval != 0 {
		go func() {
			for { // Schedule sync channels to run repetately
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
	for _, s := range sS.getSessions("", false) { // Force sessions shutdown
		sS.endSession(s, nil, nil, nil)
	}
	return
}

// OnBiJSONConnect is called by rpc2.Client on each new connection
func (sS *SessionS) OnBiJSONConnect(c *rpc2.Client) {
	sS.biJMux.Lock()
	nodeID := utils.UUIDSha1Prefix() // connection identifier, should be later updated as login procedure
	sS.biJClnts[c] = nodeID
	sS.biJIDs[nodeID] = &biJClient{conn: c,
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

// RegisterIntBiJConn is called on each internal BiJ connection towards SessionS
func (sS *SessionS) RegisterIntBiJConn(c rpcclient.RpcClientConnection) {
	sS.biJMux.Lock()
	nodeID := sS.cgrCfg.GeneralCfg().NodeID
	sS.biJClnts[c] = nodeID
	sS.biJIDs[nodeID] = &biJClient{conn: c,
		proto: sS.cgrCfg.SessionSCfg().ClientProtocol}
	sS.biJMux.Unlock()
}

// biJClnt returns a bidirectional JSON client based on connection ID
func (sS *SessionS) biJClnt(connID string) (clnt *biJClient) {
	if connID == "" {
		return nil
	}
	sS.biJMux.RLock()
	clnt = sS.biJIDs[connID]
	sS.biJMux.RUnlock()
	return
}

// biJClnt returns connection ID based on bidirectional connection received
func (sS *SessionS) biJClntID(c rpcclient.RpcClientConnection) (clntConnID string) {
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
	timer       *time.Timer
	endChan     chan struct{}
	ttl         time.Duration
	ttlLastUsed *time.Duration
	ttlUsage    *time.Duration
}

// setSTerminator installs a new terminator for a session
// assumes the Session is locked in this stage
func (sS *SessionS) setSTerminator(s *Session) {
	ttl, err := s.EventStart.GetDuration(utils.SessionTTL)
	switch err {
	case nil: // all good
	case utils.ErrNotFound:
		ttl = sS.cgrCfg.SessionSCfg().SessionTTL
	default: // not nil
		utils.Logger.Warning(
			fmt.Sprintf("<%s>, cannot extract %s from event: %s, err: %s",
				utils.SessionS, utils.SessionTTL, s.EventStart.String(), err.Error()))
	}
	s.Lock()
	defer s.Unlock()
	if ttl == 0 && s.sTerminator == nil {
		return // nothing to set up
	}
	// random delay computation
	maxDelay, err := s.EventStart.GetDuration(utils.SessionTTLMaxDelay)
	switch err {
	case nil: // all good
	case utils.ErrNotFound:
		if sS.cgrCfg.SessionSCfg().SessionTTLMaxDelay != nil {
			maxDelay = *sS.cgrCfg.SessionSCfg().SessionTTLMaxDelay
		}
	default: // not nil
		utils.Logger.Warning(
			fmt.Sprintf("<%s>, cannot extract %s from event %s, err: %s",
				utils.SessionS, utils.SessionTTLMaxDelay, s.EventStart.String(), err.Error()))
		return
	}
	sTTLMaxDelay := maxDelay.Nanoseconds() / time.Millisecond.Nanoseconds() // Milliseconds precision for randomness
	if sTTLMaxDelay != 0 {
		rand.Seed(time.Now().Unix())
		ttl += time.Duration(rand.Int63n(sTTLMaxDelay) * time.Millisecond.Nanoseconds())
	}
	ttlLastUsed, err := s.EventStart.GetDurationPtrOrDefault(utils.SessionTTLLastUsed, sS.cgrCfg.SessionSCfg().SessionTTLLastUsed)
	if err != nil {
		utils.Logger.Warning(
			fmt.Sprintf("<%s>, cannot extract %s from event, disabling session timeout for event: <%s>",
				utils.SessionS, utils.SessionTTLLastUsed, s.EventStart.String()))
		return
	}
	ttlUsage, err := s.EventStart.GetDurationPtrOrDefault(utils.SessionTTLUsage, sS.cgrCfg.SessionSCfg().SessionTTLUsage)
	if err != nil {
		utils.Logger.Warning(
			fmt.Sprintf("<%s>, cannot extract %s from event, disabling session timeout for event: <%s>",
				utils.SessionS, utils.SessionTTLUsage, s.EventStart.String()))
		return
	}
	if s.sTerminator != nil {
		if ttl != 0 { // only change if different than 0
			s.sTerminator.ttl = ttl
			if ttlLastUsed != nil {
				s.sTerminator.ttlLastUsed = ttlLastUsed
			}
			if ttlUsage != nil {
				s.sTerminator.ttlUsage = ttlUsage
			}
			s.sTerminator.timer.Reset(s.sTerminator.ttl)
		}
		return
	}
	timer := time.NewTimer(ttl)
	endChan := make(chan struct{})
	s.sTerminator = &sTerminator{
		timer:       timer,
		endChan:     endChan,
		ttl:         ttl,
		ttlLastUsed: ttlLastUsed,
		ttlUsage:    ttlUsage,
	}
	go func() { // schedule automatic termination
		select {
		case <-timer.C:
			eUsage := s.sTerminator.ttl
			if s.sTerminator.ttlUsage != nil {
				eUsage = *s.sTerminator.ttlUsage
			}
			sS.forceSTerminate(s, eUsage,
				s.sTerminator.ttlLastUsed)
		case <-endChan:
			timer.Stop()
		}
		s.Lock()
		s.sTerminator = nil
		s.Unlock()
	}()
}

// forceSTerminate is called when a session times-out or it is forced from CGRateS side
func (sS *SessionS) forceSTerminate(s *Session, extraDebit time.Duration, lastUsed *time.Duration) (err error) {
	if extraDebit != 0 {
		for i := range s.SRuns {
			if _, err = sS.debitSession(s, i, extraDebit, lastUsed); err != nil {
				utils.Logger.Warning(
					fmt.Sprintf(
						"<%s> failed debitting cgrID %s, sRunIdx: %d, err: %s",
						utils.SessionS, s.CGRid(), i, err.Error()))
			}
		}
	}
	// we apply the correction before
	if err = sS.endSession(s, nil, nil, nil); err != nil {
		utils.Logger.Warning(
			fmt.Sprintf(
				"<%s> failed force terminating session with ID <%s>, err: <%s>",
				utils.SessionS, s.CGRid(), err.Error()))
	}
	// post the CDRs
	if sS.cdrS != nil {
		if cgrEvs, err := s.asCGREvents(); err != nil {
			utils.Logger.Warning(
				fmt.Sprintf(
					"<%s> failed convering session: %s in CGREvents with err: %s",
					utils.SessionS, utils.ToJSON(s), err.Error()))
		} else {
			var reply string
			for _, cgrEv := range cgrEvs {
				argsProc := &engine.ArgV1ProcessEvent{
					CGREvent:      *cgrEv,
					ChargerS:      utils.BoolPointer(false),
					AttributeS:    utils.BoolPointer(false),
					ArgDispatcher: s.ArgDispatcher,
				}
				if unratedReqs.HasField( // order additional rating for unrated request types
					engine.NewMapEvent(cgrEv.Event).GetStringIgnoreErrors(utils.RequestType)) {
					argsProc.RALs = utils.BoolPointer(true)
				}
				if err = sS.cdrS.Call(utils.CDRsV1ProcessEvent, argsProc, &reply); err != nil {
					utils.Logger.Warning(
						fmt.Sprintf(
							"<%s> could not post CDR for event %s, err: %s",
							utils.SessionS, utils.ToJSON(cgrEv), err.Error()))
				}
			}
		}
	}
	// release the resources for the session
	if sS.resS != nil && s.ResourceID != "" {
		var reply string
		argsRU := utils.ArgRSv1ResourceUsage{
			CGREvent: &utils.CGREvent{
				Tenant: s.Tenant,
				ID:     utils.GenUUID(),
				Event:  s.EventStart.AsMapInterface(),
			},
			UsageID:       s.ResourceID,
			Units:         1,
			ArgDispatcher: s.ArgDispatcher,
		}
		if err := sS.resS.Call(utils.ResourceSv1ReleaseResources,
			argsRU, &reply); err != nil {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> error: %s could not release resource with resourceID: %s",
					utils.SessionS, err.Error(), s.ResourceID))
		}
	}
	sS.replicateSessions(s.CGRID, false, sS.sReplConns)
	if clntConn := sS.biJClnt(s.ClientConnID); clntConn != nil {
		go func() {
			var rply string
			if err := clntConn.conn.Call(utils.SessionSv1DisconnectSession,
				utils.AttrDisconnectSession{
					EventStart: s.EventStart.AsMapInterface(),
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

	s.Lock()
	if sRunIdx >= len(s.SRuns) {
		err = errors.New("sRunIdx out of range")
		s.Unlock()
		return
	}

	sr := s.SRuns[sRunIdx]

	rDur := sr.debitReserve(dur, lastUsed) // debit out of reserve, rDur is still to be debited
	if rDur == time.Duration(0) {
		s.Unlock()
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
	s.Unlock()
	cc := new(engine.CallCost)
	if err := sS.ralS.Call(utils.ResponderMaxDebit,
		&engine.CallDescriptorWithArgDispatcher{CallDescriptor: cd,
			ArgDispatcher: argDsp}, cc); err != nil {
		s.Lock()
		sr.ExtraDuration += dbtRsrv
		s.Unlock()
		return 0, err
	}
	s.Lock()
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
	sr.CD.LoopIndex += 1
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
	s.Unlock()
	return
}

// debitLoopSession will periodically debit sessions, ie: automatic prepaid
func (sS *SessionS) debitLoopSession(s *Session, sRunIdx int,
	dbtIvl time.Duration) (maxDur time.Duration, err error) {

	s.RLock()
	lenSRuns := len(s.SRuns)
	s.RUnlock()
	if sRunIdx >= lenSRuns {
		err = errors.New("sRunIdx out of range")
		return
	}

	for {
		var maxDebit time.Duration
		if maxDebit, err = sS.debitSession(s, sRunIdx, dbtIvl, nil); err != nil {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> could not complete debit operation on session: <%s>, error: <%s>",
					utils.SessionS, s.CGRid(), err.Error()))
			dscReason := utils.ErrServerError.Error()
			if err.Error() == utils.ErrUnauthorizedDestination.Error() {
				dscReason = err.Error()
			}
			if err = sS.disconnectSession(s, dscReason); err != nil {
				utils.Logger.Warning(
					fmt.Sprintf("<%s> could not disconnect session: %s, error: %s",
						utils.SessionS, s.CGRid(), err.Error()))
			}
			return
		} else if maxDebit < dbtIvl {
			go func() { // schedule sending disconnect command
				select {
				case <-s.debitStop: // call was disconnected already
					return
				case <-time.After(maxDebit):
					if err := sS.disconnectSession(s, utils.ErrInsufficientCredit.Error()); err != nil {
						utils.Logger.Warning(
							fmt.Sprintf("<%s> could not command disconnect session: %s, error: %s",
								utils.SessionS, s.CGRid(), err.Error()))
					}
				}
			}()
		}
		select {
		case <-s.debitStop:
			return
		case <-time.After(dbtIvl):
			continue
		}
	}

	return
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
	sCC := srplsEC.AsCallCost()
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
		TOR:         sr.CD.TOR,
		Increments:  incrmts,
	}
	var acnt engine.Account
	if err = sS.ralS.Call(utils.ResponderRefundIncrements,
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
	if sRunIdx >= len(s.SRuns) {
		return errors.New("sRunIdx out of range")
	}
	sr := s.SRuns[sRunIdx]
	if sr.EventCost == nil {
		return // no costs to save, ignore the operation
	}
	smCost := &engine.V2SMCost{
		CGRID:       s.CGRID,
		CostSource:  utils.MetaSessionS,
		RunID:       sr.Event.GetStringIgnoreErrors(utils.RunID),
		OriginHost:  s.EventStart.GetStringIgnoreErrors(utils.OriginHost),
		OriginID:    s.EventStart.GetStringIgnoreErrors(utils.OriginID),
		Usage:       sr.TotalUsage,
		CostDetails: sr.EventCost,
	}
	argSmCost := &engine.ArgsV2CDRSStoreSMCost{
		Cost:           smCost,
		CheckDuplicate: true,
		ArgDispatcher:  s.ArgDispatcher,
		TenantArg: &utils.TenantArg{
			Tenant: s.Tenant,
		},
	}
	var reply string
	if err := sS.cdrS.Call(utils.CDRsV2StoreSessionCost,
		argSmCost, &reply); err != nil {
		if err == utils.ErrExists {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> refunding session: <%s> error: <%s>",
					utils.SessionS, s.CGRID, err.Error()))
			if err = sS.refundSession(s, sRunIdx, sr.CD.GetDuration()); err != nil { // refund entire duration
				utils.Logger.Warning(
					fmt.Sprintf(
						"<%s> failed refunding session: <%s>, srIdx: <%d>, error: <%s>",
						utils.SessionS, s.CGRID, sRunIdx, err.Error()))
			}
		} else {
			return err
		}
	}
	return nil
}

// disconnectSession will send disconnect from SessionS to clients
func (sS *SessionS) disconnectSession(s *Session, rsn string) (err error) {
	clnt := sS.biJClnt(s.ClientConnID)
	if clnt == nil {
		return fmt.Errorf("calling %s requires bidirectional JSON connection", utils.SessionSv1DisconnectSession)
	}
	s.EventStart.Set(utils.Usage, s.TotalUsage()) // Set the usage to total one debitted
	sEv := s.EventStart.AsMapInterface()
	servMethod := utils.SessionSv1DisconnectSession
	if clnt.proto == 0 { // compatibility with OpenSIPS 2.3
		servMethod = "SMGClientV1.DisconnectSession"
	}
	var rply string
	if err := clnt.conn.Call(servMethod,
		utils.AttrDisconnectSession{EventStart: sEv,
			Reason: rsn}, &rply); err != nil {
		if err != utils.ErrNotImplemented {
			return err
		}
		err = nil
	}
	return
}

// replicateSessions will replicate sessions with or without cgrID specified
func (sS *SessionS) replicateSessions(cgrID string, psv bool, rplConns []*SReplConn) (err error) {
	if len(rplConns) == 0 {
		return
	}
	ss := sS.getSessions(cgrID, psv)
	if len(ss) == 0 {
		// session scheduled to be removed from remote (initiate also the EventStart to avoid the panic)
		ss = []*Session{&Session{CGRID: cgrID, EventStart: engine.NewSafEvent(nil)}}
	}
	var wg sync.WaitGroup
	for _, rplConn := range rplConns {
		if rplConn.Synchronous {
			wg.Add(1)
		}
		go func(conn rpcclient.RpcClientConnection, sync bool, ss []*Session) {
			for _, s := range ss {
				sCln := s.Clone()
				var rply string
				if err := conn.Call(utils.SessionSv1SetPassiveSession,
					sCln, &rply); err != nil {
					utils.Logger.Warning(
						fmt.Sprintf("<%s> cannot replicate session with id <%s>, err: %s",
							utils.SessionS, sCln.CGRID, err.Error()))
				}
			}
			if sync {
				wg.Done()
			}
		}(rplConn.Connection, rplConn.Synchronous, ss)
	}
	wg.Wait() // wait for synchronous replication to finish
	return
}

// registerSession will register an active or passive Session
// called on init or relocate
func (sS *SessionS) registerSession(s *Session, passive bool) {
	sMux := &sS.aSsMux
	sMp := sS.aSessions
	if passive {
		sMux = &sS.pSsMux
		sMp = sS.pSessions
	}
	sMux.Lock()
	sMp[s.CGRID] = s
	sS.indexSession(s, passive)
	if !passive {
		sS.setSTerminator(s)
	}
	sMux.Unlock()
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
	s, found := sMp[cgrID]
	if !found {
		sMux.Unlock()
		return false
	}
	delete(sMp, cgrID)
	sS.unindexSession(cgrID, passive)
	if !passive {
		if s.sTerminator != nil &&
			s.sTerminator.endChan != nil {
			close(s.sTerminator.endChan)
			s.sTerminator.endChan = nil
			time.Sleep(1) // ensure context switching so that the goroutine can remove old terminator
		}
	}
	sMux.Unlock()
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
			fieldVal, err := sr.Event.GetString(fieldName)
			if err != nil {
				if err == utils.ErrNotFound {
					fieldVal = utils.NOT_AVAILABLE
				} else {
					utils.Logger.Err(fmt.Sprintf("<%s> retrieving field: %s from event: %+v, err: <%s>", utils.SessionS, fieldName, s.EventStart, err))
					continue
				}
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

func (sS *SessionS) getIndexedFilters(tenant string, fltrs []string) (indexedFltr map[string][]string, unindexedFltr []*engine.FilterRule) {
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
			if fltr.Type != utils.MetaString ||
				!sS.cgrCfg.SessionSCfg().SessionIndexes.HasKey(strings.TrimPrefix(fltr.FieldName, utils.DynamicDataPrefix)) {
				unindexedFltr = append(unindexedFltr, fltr)
				continue
			}
			indexedFltr[fltr.FieldName] = fltr.Values
		}
	}
	return
}

// matchedIndexes returns map[matchedFieldName]possibleMatchedFieldVal so we optimize further to avoid checking them
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
		fltrName = strings.TrimPrefix(fltrName, utils.DynamicDataPrefix)
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
		checkNr += 1
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
	ss := sS.getSessionsFromCGRIDs(psv, cgrIDs...)
	pass := func(filterRules []*engine.FilterRule,
		me engine.MapEvent) (pass bool) {
		pass = true
		if len(filterRules) == 0 {
			return
		}
		var err error
		ev := config.NewNavigableMap(me)
		for _, fltr := range filterRules {
			if pass, err = fltr.Pass(ev, sS.statS, tenant); err != nil || !pass {
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

func (sS *SessionS) filterSessionsCount(sf *utils.SessionFilter, psv bool) (count int) {
	count = 0
	if len(sf.Filters) == 0 {
		ss := sS.getSessions(utils.EmptyString, psv)
		for _, s := range ss {
			s.RLock()
			count += len(s.SRuns)
			s.RUnlock()
		}
		return
	}
	tenant := utils.FirstNonEmpty(sf.Tenant, sS.cgrCfg.GeneralCfg().DefaultTenant)
	indx, unindx := sS.getIndexedFilters(tenant, sf.Filters)
	cgrIDs, matchingSRuns := sS.getSessionIDsMatchingIndexes(indx, psv)
	ss := sS.getSessionsFromCGRIDs(psv, cgrIDs...)
	pass := func(filterRules []*engine.FilterRule,
		me engine.MapEvent) (pass bool) {
		pass = true
		if len(filterRules) == 0 {
			return
		}
		var err error
		ev := config.NewNavigableMap(me)
		for _, fltr := range filterRules {
			if pass, err = fltr.Pass(ev, sS.statS, tenant); err != nil || !pass {
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
				count += 1
			}
		}
		s.RUnlock()
	}
	return
}

// forkSession will populate SRuns within a Session based on ChargerS output
// forSession can only be called once per Session
// not thread-safe since it should be called in init where there is no concurrency
func (sS *SessionS) forkSession(s *Session) (err error) {
	if sS.chargerS == nil {
		return errors.New("ChargerS is disabled")
	}
	if len(s.SRuns) != 0 {
		return errors.New("already forked")
	}
	cgrEv := &utils.CGREventWithArgDispatcher{
		CGREvent: &utils.CGREvent{
			Tenant: s.Tenant,
			ID:     utils.UUIDSha1Prefix(),
			Event:  s.EventStart.AsMapInterface(),
		},
		ArgDispatcher: s.ArgDispatcher,
	}
	var chrgrs []*engine.ChrgSProcessEventReply
	if err = sS.chargerS.Call(utils.ChargerSv1ProcessEvent,
		cgrEv, &chrgrs); err != nil {
		if err.Error() == utils.ErrNotFound.Error() {
			return utils.ErrNoActiveSession
		}
		return
	}
	s.SRuns = make([]*SRun, len(chrgrs))
	for i, chrgr := range chrgrs {
		me := engine.NewMapEvent(chrgr.CGREvent.Event)
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
				CgrID:       s.CGRID,
				RunID:       me.GetStringIgnoreErrors(utils.RunID),
				TOR:         me.GetStringIgnoreErrors(utils.ToR),
				Tenant:      s.Tenant,
				Category:    category,
				Subject:     subject,
				Account:     me.GetStringIgnoreErrors(utils.Account),
				Destination: me.GetStringIgnoreErrors(utils.Destination),
				TimeStart:   startTime,
				TimeEnd:     startTime.Add(s.EventStart.GetDurationIgnoreErrors(utils.Usage)),
				ExtraFields: me.AsMapStringIgnoreErrors(utils.NewStringMap(utils.MainCDRFields...)),
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
func (sS *SessionS) transitSState(cgrID string, psv bool) (ss []*Session) {
	ss = sS.getSessions(cgrID, !psv)
	for _, s := range ss {
		s.RLock() // protect the sTerminator
		sS.unregisterSession(cgrID, !psv)
		s.RUnlock()
		sS.registerSession(s, psv)
		// ToDo: activate prepaid debits
	}
	return
}

// getActivateSessions returns the sessions from active list or moves from passive
func (sS *SessionS) getActivateSessions(cgrID string) (ss []*Session) {
	ss = sS.getSessions(cgrID, false)
	if len(ss) == 0 {
		ss = sS.transitSState(cgrID, false)
	}
	return
}

// relocateSession will change the CGRID of a session (ie: prefix based session group)
func (sS *SessionS) relocateSessions(initOriginID, originID, originHost string) (ss []*Session) {
	if initOriginID == "" {
		return
	}
	initCGRID := utils.Sha1(initOriginID, originHost)
	newCGRID := utils.Sha1(originID, originHost)
	ss = sS.getActivateSessions(initCGRID)
	for _, s := range ss {
		s.Lock()
		sS.unregisterSession(s.CGRID, false)
		s.CGRID = newCGRID
		// Overwrite initial CGRID with new one
		s.EventStart.Set(utils.CGRID, newCGRID)    // Overwrite CGRID for final CDR
		s.EventStart.Set(utils.OriginID, originID) // Overwrite OriginID for session indexing
		for _, sRun := range s.SRuns {
			sRun.Event[utils.CGRID] = newCGRID // needed for CDR generation
			sRun.Event[utils.OriginID] = originID
		}
		s.Unlock()
		sS.registerSession(s, false)
		sS.replicateSessions(initCGRID, false, sS.sReplConns)
	}
	return
}

// getRelocateSessions will relocate a session if it cannot find cgrID and initialOriginID is present
func (sS *SessionS) getRelocateSessions(cgrID string, initOriginID,
	originID, originHost string) (ss []*Session) {
	if ss = sS.getActivateSessions(cgrID); len(ss) != 0 ||
		initOriginID == "" {
		return
	}
	return sS.relocateSessions(initOriginID, originID, originHost)
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
					fmt.Sprintf("<%s> error quering session ids : %+v", utils.SessionS, err))
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
		if err := sS.forceSTerminate(ss[0], 0, nil); err != nil {
			utils.Logger.Warning(fmt.Sprintf("<%s> failed force-terminating session: <%s>, err: <%s>", utils.SessionS, cgrID, err))
		}
	}
}

// initSessionDebitLoops will init the debit loops for a session
// not thread-safe, it should be protected in another layer
func (sS *SessionS) initSessionDebitLoops(s *Session) {
	if s.debitStop != nil { // already initialized
		return
	}
	s.debitStop = make(chan struct{})
	for i, sr := range s.SRuns {
		if s.DebitInterval != 0 &&
			sr.Event.GetStringIgnoreErrors(utils.RequestType) == utils.META_PREPAID {
			go sS.debitLoopSession(s, i, s.DebitInterval)
			time.Sleep(1) // allow the goroutine to be executed
		}
	}
}

// authSession calculates maximum usage allowed for given session
func (sS *SessionS) authSession(tnt string, evStart *engine.SafEvent) (maxUsage time.Duration, err error) {
	cgrID := GetSetCGRID(evStart)
	if _, err = evStart.GetDuration(utils.Usage); err != nil {
		if err != utils.ErrNotFound {
			return
		}
		evStart.Set(utils.Usage, sS.cgrCfg.SessionSCfg().MaxCallDuration) // will be used in CD
		err = nil
	}
	s := &Session{
		CGRID:      cgrID,
		Tenant:     tnt,
		EventStart: evStart,
	}
	//check if we have APIKey in event and in case it has add it in ArgDispatcher
	apiKeyIface, errApiKey := evStart.FieldAsString([]string{utils.MetaApiKey})
	if errApiKey == nil {
		s.ArgDispatcher = &utils.ArgDispatcher{
			APIKey: utils.StringPointer(apiKeyIface),
		}
	}
	//check if we have RouteID in event and in case it has add it in ArgDispatcher
	routeIDIface, errRouteID := evStart.FieldAsString([]string{utils.MetaRouteID})
	if errRouteID == nil {
		if errApiKey.Error() == utils.ErrNotFound.Error() { //in case we don't have APIKey, but we have RouteID we need to initialize the struct
			s.ArgDispatcher = &utils.ArgDispatcher{
				RouteID: utils.StringPointer(routeIDIface),
			}
		} else {
			s.ArgDispatcher.RouteID = utils.StringPointer(routeIDIface)
		}
	}
	if err = sS.forkSession(s); err != nil {
		return
	}

	var maxUsageSet bool // so we know if we have set the 0 on purpose
	prepaidReqs := []string{utils.META_PREPAID, utils.META_PSEUDOPREPAID}
	for _, sr := range s.SRuns {
		var rplyMaxUsage time.Duration
		if !utils.IsSliceMember(prepaidReqs,
			sr.Event.GetStringIgnoreErrors(utils.RequestType)) {
			rplyMaxUsage = time.Duration(-1)
		} else if err = sS.ralS.Call(utils.ResponderGetMaxSessionTime,
			&engine.CallDescriptorWithArgDispatcher{CallDescriptor: sr.CD,
				ArgDispatcher: s.ArgDispatcher}, &rplyMaxUsage); err != nil {
			return
		}
		if !maxUsageSet ||
			maxUsage == time.Duration(-1) ||
			(rplyMaxUsage < maxUsage && rplyMaxUsage != time.Duration(-1)) {
			maxUsage = rplyMaxUsage
			maxUsageSet = true
		}
	}
	return
}

// initSession handles a new session
func (sS *SessionS) initSession(tnt string, evStart *engine.SafEvent, clntConnID string,
	resID string, dbtItval time.Duration, argDisp *utils.ArgDispatcher) (s *Session, err error) {
	cgrID := GetSetCGRID(evStart)
	s = &Session{
		CGRID:         cgrID,
		Tenant:        tnt,
		ResourceID:    resID,
		EventStart:    evStart,
		ClientConnID:  clntConnID,
		DebitInterval: dbtItval,
		ArgDispatcher: argDisp,
	}
	if err = sS.forkSession(s); err != nil {
		return nil, err
	}
	sS.initSessionDebitLoops(s)
	sS.registerSession(s, false) // make the session available to the rest of the system
	return
}

// updateSession will reset terminator, perform debits and replicate sessions
func (sS *SessionS) updateSession(s *Session, updtEv engine.MapEvent) (maxUsage time.Duration, err error) {
	defer sS.replicateSessions(s.CGRID, false, sS.sReplConns)
	// update fields from new event
	for k, v := range updtEv {
		if protectedSFlds.HasField(k) {
			continue
		}
		s.EventStart.Set(k, v) // update previoius field with new one
	}
	sS.setSTerminator(s) // reset the terminator
	//init has no updtEv
	if updtEv == nil {
		updtEv = engine.NewMapEvent(s.EventStart.AsMapInterface())
	}
	var reqMaxUsage time.Duration
	if reqMaxUsage, err = updtEv.GetDuration(utils.Usage); err != nil {
		if err != utils.ErrNotFound {
			return
		}
		reqMaxUsage = sS.cgrCfg.SessionSCfg().MaxCallDuration
		err = nil
	}
	var maxUsageSet bool // so we know if we have set the 0 on purpose
	prepaidReqs := []string{utils.META_PREPAID, utils.META_PSEUDOPREPAID}
	for i, sr := range s.SRuns {
		var rplyMaxUsage time.Duration
		if !utils.IsSliceMember(prepaidReqs,
			sr.Event.GetStringIgnoreErrors(utils.RequestType)) {
			rplyMaxUsage = time.Duration(-1)
		} else if rplyMaxUsage, err = sS.debitSession(s, i, reqMaxUsage,
			updtEv.GetDurationPtrIgnoreErrors(utils.LastUsed)); err != nil {
			return
		}
		if !maxUsageSet ||
			maxUsage == time.Duration(-1) ||
			(rplyMaxUsage < maxUsage && rplyMaxUsage != time.Duration(-1)) {
			maxUsage = rplyMaxUsage
			maxUsageSet = true
		}
	}

	return
}

// endSession will end a session from outside
func (sS *SessionS) endSession(s *Session, tUsage, lastUsage *time.Duration, aTime *time.Time) (err error) {
	//check if we have replicate connection and close the session there
	defer sS.replicateSessions(s.CGRID, true, sS.sReplConns)

	s.Lock() // no need to release it untill end since the session should be anyway closed
	sS.unregisterSession(s.CGRID, false)
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
		if s.debitStop != nil {
			close(s.debitStop) // Stop automatic debits
			s.debitStop = nil
		}
		if sr.EventCost != nil {
			if notCharged := sUsage - sr.EventCost.GetUsage(); notCharged > 0 { // we did not charge enough, make a manual debit here
				if sr.CD.LoopIndex > 0 {
					sr.CD.TimeStart = sr.CD.TimeEnd
				}
				sr.CD.TimeEnd = sr.CD.TimeStart.Add(notCharged)
				sr.CD.DurationIndex += notCharged
				cc := new(engine.CallCost)
				if err = sS.ralS.Call(utils.ResponderDebit,
					&engine.CallDescriptorWithArgDispatcher{CallDescriptor: sr.CD,
						ArgDispatcher: s.ArgDispatcher}, cc); err == nil {
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
			// set cost fields
			sr.Event[utils.Cost] = sr.EventCost.GetCost()
			sr.Event[utils.CostDetails] = utils.ToJSON(sr.EventCost) // avoid map[string]interface{} when decoding
			sr.Event[utils.CostSource] = utils.MetaSessionS
		}
		// Set Usage field
		if sRunIdx == 0 {
			s.EventStart.Set(utils.Usage, sr.TotalUsage)
		}
		sr.Event[utils.Usage] = sr.TotalUsage
		if aTime != nil {
			sr.Event[utils.AnswerTime] = *aTime
		}
		if sS.cgrCfg.SessionSCfg().StoreSCosts {
			if err := sS.storeSCost(s, sRunIdx); err != nil {
				utils.Logger.Warning(
					fmt.Sprintf("<%s> failed storing session cost for <%s>, srIdx: <%d>, error: <%s>",
						utils.SessionS, s.CGRID, sRunIdx, err.Error()))
			}
		}

	}
	engine.Cache.Set(utils.CacheClosedSessions, s.CGRID, s,
		nil, true, utils.NonTransactional)
	s.Unlock()
	return
}

// chargeEvent will charge a single event (ie: SMS)
func (sS *SessionS) chargeEvent(tnt string, ev *engine.SafEvent, argDisp *utils.ArgDispatcher) (maxUsage time.Duration, err error) {
	cgrID := GetSetCGRID(ev)
	var s *Session
	if s, err = sS.initSession(tnt, ev, "", "", 0, argDisp); err != nil {
		return
	}
	if maxUsage, err = sS.updateSession(s, ev.AsMapInterface()); err != nil {
		if errEnd := sS.endSession(s, utils.DurationPointer(time.Duration(0)), nil, nil); errEnd != nil {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> error when force-ending charged event: <%s>, err: <%s>",
					utils.SessionS, cgrID, err.Error()))
		}
		return
	}
	usage := maxUsage
	if utils.IsSliceMember(utils.PostPaidRatedSlice, ev.GetStringIgnoreErrors(utils.RequestType)) {
		usage = ev.GetDurationIgnoreErrors(utils.Usage)
	}
	//in case of postpaid and rated maxUsage = usage from event
	if errEnd := sS.endSession(s, utils.DurationPointer(usage), nil, nil); errEnd != nil {
		utils.Logger.Warning(
			fmt.Sprintf("<%s> error when ending charged event: <%s>, err: <%s>",
				utils.SessionS, cgrID, err.Error()))
	}
	return // returns here the maxUsage from update
}

// APIs start here

// Call is part of RpcClientConnection interface
func (sS *SessionS) Call(serviceMethod string, args interface{}, reply interface{}) error {
	return sS.CallBiRPC(nil, serviceMethod, args, reply)
}

// CallBiRPC is part of utils.BiRPCServer interface to help internal connections do calls over rpcclient.RpcClientConnection interface
func (sS *SessionS) CallBiRPC(clnt rpcclient.RpcClientConnection,
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
func (sS *SessionS) BiRPCv1GetActiveSessions(clnt rpcclient.RpcClientConnection,
	args *utils.SessionFilter, reply *[]*ExternalSession) (err error) {
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
	*reply = aSs
	return nil
}

// BiRPCv1GetActiveSessionsCount counts the active sessions
func (sS *SessionS) BiRPCv1GetActiveSessionsCount(clnt rpcclient.RpcClientConnection,
	args *utils.SessionFilter, reply *int) error {
	if args == nil { //protection in case on nil
		args = &utils.SessionFilter{}
	}
	if len(args.Filters) != 0 && sS.dm == nil {
		return utils.ErrNoDatabaseConn
	}
	*reply = sS.filterSessionsCount(args, false)
	return nil
}

// BiRPCv1GetPassiveSessions returns the passive sessions handled by SessionS
func (sS *SessionS) BiRPCv1GetPassiveSessions(clnt rpcclient.RpcClientConnection,
	args *utils.SessionFilter, reply *[]*ExternalSession) error {
	if args == nil { //protection in case on nil
		args = &utils.SessionFilter{}
	}
	if len(args.Filters) != 0 && sS.dm == nil {
		return utils.ErrNoDatabaseConn
	}
	pSs := sS.filterSessions(args, true)
	if len(pSs) == 0 {
		return utils.ErrNotFound
	}
	*reply = pSs
	return nil
}

// BiRPCv1GetPassiveSessionsCount counts the passive sessions handled by the system
func (sS *SessionS) BiRPCv1GetPassiveSessionsCount(clnt rpcclient.RpcClientConnection,
	args *utils.SessionFilter, reply *int) error {
	if args == nil { //protection in case on nil
		args = &utils.SessionFilter{}
	}
	if len(args.Filters) != 0 && sS.dm == nil {
		return utils.ErrNoDatabaseConn
	}
	*reply = sS.filterSessionsCount(args, true)
	return nil
}

// BiRPCv1SetPassiveSessions used for replicating Sessions
func (sS *SessionS) BiRPCv1SetPassiveSession(clnt rpcclient.RpcClientConnection,
	s *Session, reply *string) (err error) {
	if s.CGRID == "" {
		return utils.NewErrMandatoryIeMissing(utils.CGRID)
	}
	if s.EventStart == nil { // remove instead of
		s.RLock()
		removed := sS.unregisterSession(s.CGRID, true)
		s.RUnlock()
		if !removed {
			err = utils.ErrServerError
			return
		}
	} else {
		//if we have an active session with the same CGRID
		//we unregister it first then regiser the new one
		s.Lock()
		if len(sS.getSessions(s.CGRID, false)) != 0 {
			sS.unregisterSession(s.CGRID, false)
		}

		sS.initSessionDebitLoops(s)
		s.Unlock()
		sS.registerSession(s, true)
	}
	*reply = utils.OK
	return
}

type ArgsReplicateSessions struct {
	CGRID       string
	Passive     bool
	Connections []*config.RemoteHost
}

// BiRPCv1ReplicateSessions will replicate active sessions to either args.Connections or the internal configured ones
// args.Filter is used to filter the sessions which are replicated, CGRID is the only one possible for now
func (sS *SessionS) BiRPCv1ReplicateSessions(clnt rpcclient.RpcClientConnection,
	args ArgsReplicateSessions, reply *string) (err error) {
	sSConns := sS.sReplConns
	if len(args.Connections) != 0 {
		if sSConns, err = NewSReplConns(args.Connections,
			sS.cgrCfg.GeneralCfg().Reconnects,
			sS.cgrCfg.GeneralCfg().ConnectTimeout,
			sS.cgrCfg.GeneralCfg().ReplyTimeout); err != nil {
			return utils.NewErrServerError(err)
		}
	}
	if err = sS.replicateSessions(args.CGRID, args.Passive, sSConns); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = utils.OK
	return
}

// NewV1AuthorizeArgs is a constructor for V1AuthorizeArgs
func NewV1AuthorizeArgs(attrs, res, maxUsage, thrslds,
	statQueues, suppls, supplsIgnoreErrs, supplsEventCost bool,
	cgrEv *utils.CGREvent, argDisp *utils.ArgDispatcher,
	supplierPaginator utils.Paginator) (args *V1AuthorizeArgs) {
	args = &V1AuthorizeArgs{
		GetAttributes:         attrs,
		AuthorizeResources:    res,
		GetMaxUsage:           maxUsage,
		ProcessThresholds:     thrslds,
		ProcessStats:          statQueues,
		SuppliersIgnoreErrors: supplsIgnoreErrs,
		GetSuppliers:          suppls,
		CGREvent:              cgrEv,
	}
	if supplsEventCost {
		args.SuppliersMaxCost = utils.MetaSuppliersEventCost
	}
	args.ArgDispatcher = argDisp
	args.Paginator = supplierPaginator
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
	SuppliersMaxCost      string
	SuppliersIgnoreErrors bool
	*utils.CGREvent
	utils.Paginator
	*utils.ArgDispatcher
}

// V1AuthorizeReply are options available in auth reply
type V1AuthorizeReply struct {
	Attributes         *engine.AttrSProcessEventReply
	ResourceAllocation *string
	MaxUsage           *time.Duration
	Suppliers          *engine.SortedSuppliers
	ThresholdIDs       *[]string
	StatQueueIDs       *[]string
}

// AsNavigableMap is part of engine.NavigableMapper interface
func (v1AuthReply *V1AuthorizeReply) AsNavigableMap(
	ignr []*config.FCTemplate) (*config.NavigableMap, error) {
	cgrReply := make(map[string]interface{})
	if v1AuthReply != nil {
		if v1AuthReply.Attributes != nil {
			attrs := make(map[string]interface{})
			for _, fldName := range v1AuthReply.Attributes.AlteredFields {
				if v1AuthReply.Attributes.CGREvent.HasField(fldName) {
					attrs[fldName] = v1AuthReply.Attributes.CGREvent.Event[fldName]
				}
			}
			cgrReply[utils.CapAttributes] = attrs
		}
		if v1AuthReply.ResourceAllocation != nil {
			cgrReply[utils.CapResourceAllocation] = *v1AuthReply.ResourceAllocation
		}
		if v1AuthReply.MaxUsage != nil {
			cgrReply[utils.CapMaxUsage] = *v1AuthReply.MaxUsage
		}
		if v1AuthReply.Suppliers != nil {
			cgrReply[utils.CapSuppliers] = v1AuthReply.Suppliers.AsNavigableMap()
		}
		if v1AuthReply.ThresholdIDs != nil {
			cgrReply[utils.CapThresholds] = *v1AuthReply.ThresholdIDs
		}
		if v1AuthReply.StatQueueIDs != nil {
			cgrReply[utils.CapStatQueues] = *v1AuthReply.StatQueueIDs
		}
	}
	return config.NewNavigableMap(cgrReply), nil
}

// BiRPCv1AuthorizeEvent performs authorization for CGREvent based on specific components
func (sS *SessionS) BiRPCv1AuthorizeEvent(clnt rpcclient.RpcClientConnection,
	args *V1AuthorizeArgs, authReply *V1AuthorizeReply) (err error) {
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
		if sS.attrS == nil {
			return utils.NewErrNotConnected(utils.AttributeS)
		}
		attrArgs := &engine.AttrArgsProcessEvent{
			Context:       utils.StringPointer(utils.MetaSessionS),
			CGREvent:      args.CGREvent,
			ArgDispatcher: args.ArgDispatcher,
		}
		var rplyEv engine.AttrSProcessEventReply
		if err := sS.attrS.Call(utils.AttributeSv1ProcessEvent,
			attrArgs, &rplyEv); err == nil {
			args.CGREvent = rplyEv.CGREvent
			if tntIface, has := args.CGREvent.Event[utils.MetaTenant]; has {
				// special case when we want to overwrite the tenant
				args.CGREvent.Tenant = tntIface.(string)
				delete(args.CGREvent.Event, utils.MetaTenant)
			}
			authReply.Attributes = &rplyEv
		} else if err.Error() != utils.ErrNotFound.Error() {
			return utils.NewErrAttributeS(err)
		}
	}
	if args.GetMaxUsage {
		maxUsage, err := sS.authSession(args.CGREvent.Tenant,
			engine.NewSafEvent(args.CGREvent.Event))
		if err != nil {
			return utils.NewErrRALs(err)
		}
		authReply.MaxUsage = &maxUsage
	}
	if args.AuthorizeResources {
		if sS.resS == nil {
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
		if err = sS.resS.Call(utils.ResourceSv1AuthorizeResources,
			attrRU, &allocMsg); err != nil {
			return utils.NewErrResourceS(err)
		}
		authReply.ResourceAllocation = &allocMsg
	}
	if args.GetSuppliers {
		if sS.splS == nil {
			return utils.NewErrNotConnected(utils.SupplierS)
		}
		cgrEv := args.CGREvent.Clone()
		if acd, has := cgrEv.Event[utils.ACD]; has {
			cgrEv.Event[utils.Usage] = acd
		}
		var splsReply engine.SortedSuppliers
		sArgs := &engine.ArgsGetSuppliers{
			IgnoreErrors:  args.SuppliersIgnoreErrors,
			MaxCost:       args.SuppliersMaxCost,
			CGREvent:      cgrEv,
			Paginator:     args.Paginator,
			ArgDispatcher: args.ArgDispatcher,
		}
		if err = sS.splS.Call(utils.SupplierSv1GetSuppliers,
			sArgs, &splsReply); err != nil {
			return utils.NewErrSupplierS(err)
		}
		if splsReply.SortedSuppliers != nil {
			authReply.Suppliers = &splsReply
		}
	}
	if args.ProcessThresholds {
		if sS.thdS == nil {
			return utils.NewErrNotConnected(utils.ThresholdS)
		}
		var tIDs []string
		thEv := &engine.ArgsProcessEvent{
			CGREvent:      args.CGREvent,
			ArgDispatcher: args.ArgDispatcher,
		}
		if err := sS.thdS.Call(utils.ThresholdSv1ProcessEvent, thEv, &tIDs); err != nil &&
			err.Error() != utils.ErrNotFound.Error() {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> error: %s processing event %+v with ThresholdS.",
					utils.SessionS, err.Error(), thEv))
		}
		authReply.ThresholdIDs = &tIDs
	}
	if args.ProcessStats {
		if sS.statS == nil {
			return utils.NewErrNotConnected(utils.StatService)
		}
		statArgs := &engine.StatsArgsProcessEvent{
			CGREvent:      args.CGREvent,
			ArgDispatcher: args.ArgDispatcher,
		}
		var statReply []string
		if err := sS.statS.Call(utils.StatSv1ProcessEvent,
			statArgs, &statReply); err != nil &&
			err.Error() != utils.ErrNotFound.Error() {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> error: %s processing event %+v with StatS.",
					utils.SessionS, err.Error(), args.CGREvent))
		}
		authReply.StatQueueIDs = &statReply
	}
	return nil
}

// V1AuthorizeReplyWithDigest contains return options for auth with digest
type V1AuthorizeReplyWithDigest struct {
	AttributesDigest   *string
	ResourceAllocation *string
	MaxUsage           *float64 // special treat returning time.Duration.Seconds()
	SuppliersDigest    *string
	Thresholds         *string
	StatQueues         *string
}

// BiRPCv1AuthorizeEventWithDigest performs authorization for CGREvent based on specific components
// returning one level fields instead of multiple ones returned by BiRPCv1AuthorizeEvent
func (sS *SessionS) BiRPCv1AuthorizeEventWithDigest(clnt rpcclient.RpcClientConnection,
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
		authReply.MaxUsage = utils.Float64Pointer(-1.0)
		if *initAuthRply.MaxUsage != time.Duration(-1) {
			authReply.MaxUsage = utils.Float64Pointer(initAuthRply.MaxUsage.Seconds())
		}
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
func NewV1InitSessionArgs(attrs, resrc, acnt, thrslds, stats bool,
	cgrEv *utils.CGREvent, argDisp *utils.ArgDispatcher) (args *V1InitSessionArgs) {
	args = &V1InitSessionArgs{
		GetAttributes:     attrs,
		AllocateResources: resrc,
		InitSession:       acnt,
		ProcessThresholds: thrslds,
		ProcessStats:      stats,
		CGREvent:          cgrEv,
		ArgDispatcher:     argDisp,
	}
	return
}

// V1InitSessionArgs are options for session initialization request
type V1InitSessionArgs struct {
	GetAttributes     bool
	AllocateResources bool
	InitSession       bool
	ProcessThresholds bool
	ProcessStats      bool
	*utils.CGREvent
	*utils.ArgDispatcher
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
func (v1Rply *V1InitSessionReply) AsNavigableMap(
	ignr []*config.FCTemplate) (*config.NavigableMap, error) {
	cgrReply := make(map[string]interface{})
	if v1Rply != nil {
		if v1Rply.Attributes != nil {
			attrs := make(map[string]interface{})
			for _, fldName := range v1Rply.Attributes.AlteredFields {
				if v1Rply.Attributes.CGREvent.HasField(fldName) {
					attrs[fldName] = v1Rply.Attributes.CGREvent.Event[fldName]
				}
			}
			cgrReply[utils.CapAttributes] = attrs
		}
		if v1Rply.ResourceAllocation != nil {
			cgrReply[utils.CapResourceAllocation] = *v1Rply.ResourceAllocation
		}
		if v1Rply.MaxUsage != nil {
			cgrReply[utils.CapMaxUsage] = *v1Rply.MaxUsage
		}
		if v1Rply.ThresholdIDs != nil {
			cgrReply[utils.CapThresholds] = *v1Rply.ThresholdIDs
		}
		if v1Rply.StatQueueIDs != nil {
			cgrReply[utils.CapStatQueues] = *v1Rply.StatQueueIDs
		}
	}
	return config.NewNavigableMap(cgrReply), nil
}

// BiRPCv1InitiateSession initiates a new session
func (sS *SessionS) BiRPCv1InitiateSession(clnt rpcclient.RpcClientConnection,
	args *V1InitSessionArgs, rply *V1InitSessionReply) (err error) {
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
		if sS.attrS == nil {
			return utils.NewErrNotConnected(utils.AttributeS)
		}
		attrArgs := &engine.AttrArgsProcessEvent{
			Context:       utils.StringPointer(utils.MetaSessionS),
			CGREvent:      args.CGREvent,
			ArgDispatcher: args.ArgDispatcher,
		}
		var rplyEv engine.AttrSProcessEventReply
		if err := sS.attrS.Call(utils.AttributeSv1ProcessEvent,
			attrArgs, &rplyEv); err == nil {
			args.CGREvent = rplyEv.CGREvent
			if tntIface, has := args.CGREvent.Event[utils.MetaTenant]; has {
				// special case when we want to overwrite the tenant
				args.CGREvent.Tenant = tntIface.(string)
				delete(args.CGREvent.Event, utils.MetaTenant)
			}
			rply.Attributes = &rplyEv
		} else if err.Error() != utils.ErrNotFound.Error() {
			return utils.NewErrAttributeS(err)
		}
	}
	if args.AllocateResources {
		if sS.resS == nil {
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
		if err = sS.resS.Call(utils.ResourceSv1AllocateResources,
			attrRU, &allocMessage); err != nil {
			return utils.NewErrResourceS(err)
		}
		rply.ResourceAllocation = &allocMessage
	}
	if args.InitSession {
		var err error
		ev := engine.NewSafEvent(args.CGREvent.Event)
		dbtItvl := sS.cgrCfg.SessionSCfg().DebitInterval
		if ev.HasField(utils.CGRDebitInterval) { // dynamic DebitInterval via CGRDebitInterval
			if dbtItvl, err = ev.GetDuration(utils.CGRDebitInterval); err != nil {
				return utils.NewErrRALs(err)
			}
		}
		s, err := sS.initSession(args.CGREvent.Tenant, ev,
			sS.biJClntID(clnt), originID, dbtItvl, args.ArgDispatcher)
		if err != nil {
			return utils.NewErrRALs(err)
		}
		if dbtItvl > 0 { //active debit
			rply.MaxUsage = utils.DurationPointer(time.Duration(-1))
		} else {
			if maxUsage, err := sS.updateSession(s, nil); err != nil {
				return utils.NewErrRALs(err)
			} else {
				rply.MaxUsage = &maxUsage
			}
		}
	}
	if args.ProcessThresholds {
		if sS.thdS == nil {
			return utils.NewErrNotConnected(utils.ThresholdS)
		}
		var tIDs []string
		thEv := &engine.ArgsProcessEvent{
			CGREvent:      args.CGREvent,
			ArgDispatcher: args.ArgDispatcher,
		}
		if err := sS.thdS.Call(utils.ThresholdSv1ProcessEvent,
			thEv, &tIDs); err != nil &&
			err.Error() != utils.ErrNotFound.Error() {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> error: %s processing event %+v with ThresholdS.",
					utils.SessionS, err.Error(), thEv))
		}
		rply.ThresholdIDs = &tIDs
	}
	if args.ProcessStats {
		if sS.statS == nil {
			return utils.NewErrNotConnected(utils.StatService)
		}
		var statReply []string
		statArgs := &engine.StatsArgsProcessEvent{
			CGREvent:      args.CGREvent,
			ArgDispatcher: args.ArgDispatcher,
		}
		if err := sS.statS.Call(utils.StatSv1ProcessEvent,
			statArgs, &statReply); err != nil &&
			err.Error() != utils.ErrNotFound.Error() {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> error: %s processing event %+v with StatS.",
					utils.SessionS, err.Error(), args.CGREvent))
		}
		rply.StatQueueIDs = &statReply
	}
	return
}

type V1InitReplyWithDigest struct {
	AttributesDigest   *string
	ResourceAllocation *string
	MaxUsage           *float64
	Thresholds         *string
	StatQueues         *string
}

func (sS *SessionS) BiRPCv1InitiateSessionWithDigest(clnt rpcclient.RpcClientConnection,
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
		initReply.MaxUsage = utils.Float64Pointer(-1.0)
		if *initSessionRply.MaxUsage != time.Duration(-1) {
			initReply.MaxUsage = utils.Float64Pointer(initSessionRply.MaxUsage.Seconds())
		}
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
func NewV1UpdateSessionArgs(attrs, acnts bool,
	cgrEv *utils.CGREvent, argDisp *utils.ArgDispatcher) (args *V1UpdateSessionArgs) {
	args = &V1UpdateSessionArgs{
		GetAttributes: attrs,
		UpdateSession: acnts,
		CGREvent:      cgrEv,
		ArgDispatcher: argDisp,
	}
	return
}

// V1UpdateSessionArgs contains options for session update
type V1UpdateSessionArgs struct {
	GetAttributes bool
	UpdateSession bool
	*utils.CGREvent
	*utils.ArgDispatcher
}

// V1UpdateSessionReply contains options for session update reply
type V1UpdateSessionReply struct {
	Attributes *engine.AttrSProcessEventReply
	MaxUsage   *time.Duration
}

// AsNavigableMap is part of engine.NavigableMapper interface
func (v1Rply *V1UpdateSessionReply) AsNavigableMap(
	ignr []*config.FCTemplate) (*config.NavigableMap, error) {
	cgrReply := make(map[string]interface{})
	if v1Rply != nil {
		if v1Rply.Attributes != nil {
			attrs := make(map[string]interface{})
			for _, fldName := range v1Rply.Attributes.AlteredFields {
				if v1Rply.Attributes.CGREvent.HasField(fldName) {
					attrs[fldName] = v1Rply.Attributes.CGREvent.Event[fldName]
				}
			}
			cgrReply[utils.CapAttributes] = attrs
		}
		if v1Rply.MaxUsage != nil {
			cgrReply[utils.CapMaxUsage] = *v1Rply.MaxUsage
		}
	}
	return config.NewNavigableMap(cgrReply), nil
}

// BiRPCv1UpdateSession updates an existing session, returning the duration which the session can still last
func (sS *SessionS) BiRPCv1UpdateSession(clnt rpcclient.RpcClientConnection,
	args *V1UpdateSessionArgs, rply *V1UpdateSessionReply) (err error) {
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
		if sS.attrS == nil {
			return utils.NewErrNotConnected(utils.AttributeS)
		}
		attrArgs := &engine.AttrArgsProcessEvent{
			Context:       utils.StringPointer(utils.MetaSessionS),
			CGREvent:      args.CGREvent,
			ArgDispatcher: args.ArgDispatcher,
		}
		var rplyEv engine.AttrSProcessEventReply
		if err := sS.attrS.Call(utils.AttributeSv1ProcessEvent,
			attrArgs, &rplyEv); err == nil {
			args.CGREvent = rplyEv.CGREvent
			if tntIface, has := args.CGREvent.Event[utils.MetaTenant]; has {
				// special case when we want to overwrite the tenant
				args.CGREvent.Tenant = tntIface.(string)
				delete(args.CGREvent.Event, utils.MetaTenant)
			}
			rply.Attributes = &rplyEv
		} else if err.Error() != utils.ErrNotFound.Error() {
			return utils.NewErrAttributeS(err)
		}
	}
	if args.UpdateSession {
		me := engine.NewMapEvent(args.CGREvent.Event)
		dbtItvl := sS.cgrCfg.SessionSCfg().DebitInterval
		if me.HasField(utils.CGRDebitInterval) { // dynamic DebitInterval via CGRDebitInterval
			if dbtItvl, err = me.GetDuration(utils.CGRDebitInterval); err != nil {
				return utils.NewErrRALs(err)
			}
		}
		ev := engine.NewSafEvent(args.CGREvent.Event)
		cgrID := GetSetCGRID(ev)
		ss := sS.getRelocateSessions(cgrID,
			me.GetStringIgnoreErrors(utils.InitialOriginID),
			me.GetStringIgnoreErrors(utils.OriginID),
			me.GetStringIgnoreErrors(utils.OriginHost))
		var s *Session
		if len(ss) == 0 {
			if s, err = sS.initSession(args.CGREvent.Tenant,
				ev, sS.biJClntID(clnt),
				me.GetStringIgnoreErrors(utils.OriginID), dbtItvl, args.ArgDispatcher); err != nil {
				return utils.NewErrRALs(err)
			}
		} else {
			s = ss[0]
		}
		if maxUsage, err := sS.updateSession(s, ev.AsMapInterface()); err != nil {
			return utils.NewErrRALs(err)
		} else {
			rply.MaxUsage = &maxUsage
		}
	}
	return
}

func NewV1TerminateSessionArgs(acnts, resrc, thrds, stats bool,
	cgrEv *utils.CGREvent, argDisp *utils.ArgDispatcher) (args *V1TerminateSessionArgs) {
	args = &V1TerminateSessionArgs{
		TerminateSession:  acnts,
		ReleaseResources:  resrc,
		ProcessThresholds: thrds,
		ProcessStats:      stats,
		CGREvent:          cgrEv,
		ArgDispatcher:     argDisp,
	}
	return
}

type V1TerminateSessionArgs struct {
	TerminateSession  bool
	ReleaseResources  bool
	ProcessThresholds bool
	ProcessStats      bool
	*utils.CGREvent
	*utils.ArgDispatcher
}

// BiRPCV1TerminateSession will stop debit loops as well as release any used resources
func (sS *SessionS) BiRPCv1TerminateSession(clnt rpcclient.RpcClientConnection,
	args *V1TerminateSessionArgs, rply *string) (err error) {
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
	ev := engine.NewSafEvent(args.CGREvent.Event)
	me := engine.NewMapEvent(args.CGREvent.Event) // used for easy access to fields within the event
	cgrID := GetSetCGRID(ev)
	originID := me.GetStringIgnoreErrors(utils.OriginID)
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
		ss := sS.getRelocateSessions(cgrID,
			me.GetStringIgnoreErrors(utils.InitialOriginID),
			me.GetStringIgnoreErrors(utils.OriginID),
			me.GetStringIgnoreErrors(utils.OriginHost))
		var s *Session
		if len(ss) == 0 {
			if s, err = sS.initSession(args.CGREvent.Tenant,
				ev, sS.biJClntID(clnt),
				me.GetStringIgnoreErrors(utils.OriginID), dbtItvl, args.ArgDispatcher); err != nil {
				return utils.NewErrRALs(err)
			}
		} else {
			s = ss[0]
		}
		if err = sS.endSession(s,
			me.GetDurationPtrIgnoreErrors(utils.Usage),
			me.GetDurationPtrIgnoreErrors(utils.LastUsed),
			utils.TimePointer(me.GetTimeIgnoreErrors(utils.AnswerTime, utils.EmptyString))); err != nil {
			return utils.NewErrRALs(err)
		}
	}
	if args.ReleaseResources {
		if sS.resS == nil {
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
		if err = sS.resS.Call(utils.ResourceSv1ReleaseResources,
			argsRU, &reply); err != nil {
			return utils.NewErrResourceS(err)
		}
	}
	if args.ProcessThresholds {
		if sS.thdS == nil {
			return utils.NewErrNotConnected(utils.ThresholdS)
		}
		var tIDs []string
		thEv := &engine.ArgsProcessEvent{
			CGREvent:      args.CGREvent,
			ArgDispatcher: args.ArgDispatcher,
		}
		if err := sS.thdS.Call(utils.ThresholdSv1ProcessEvent, thEv, &tIDs); err != nil &&
			err.Error() != utils.ErrNotFound.Error() {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> error: %s processing event %+v with ThresholdS.",
					utils.SessionS, err.Error(), thEv))
		}
	}
	if args.ProcessStats {
		if sS.statS == nil {
			return utils.NewErrNotConnected(utils.StatS)
		}
		var statReply []string
		statArgs := &engine.StatsArgsProcessEvent{
			CGREvent:      args.CGREvent,
			ArgDispatcher: args.ArgDispatcher,
		}
		if err := sS.statS.Call(utils.StatSv1ProcessEvent,
			statArgs, &statReply); err != nil &&
			err.Error() != utils.ErrNotFound.Error() {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> error: %s processing event %+v with StatS.",
					utils.SessionS, err.Error(), args.CGREvent))
		}
	}
	*rply = utils.OK
	return
}

// BiRPCv1ProcessCDR sends the CDR to CDRs
func (sS *SessionS) BiRPCv1ProcessCDR(clnt rpcclient.RpcClientConnection,
	cgrEvWithArgDisp *utils.CGREventWithArgDispatcher, rply *string) (err error) {
	if cgrEvWithArgDisp.ID == "" {
		cgrEvWithArgDisp.ID = utils.GenUUID()
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

	ev := engine.NewSafEvent(cgrEvWithArgDisp.Event)
	cgrID := GetSetCGRID(ev)
	ss := sS.getRelocateSessions(cgrID,
		ev.GetStringIgnoreErrors(utils.InitialOriginID),
		ev.GetStringIgnoreErrors(utils.OriginID),
		ev.GetStringIgnoreErrors(utils.OriginHost))
	var s *Session
	if len(ss) != 0 {
		utils.Logger.Warning(
			fmt.Sprintf("<%s> ProcessCDR called for active session with CGRID: <%s>",
				utils.SessionS, cgrID))
		s = ss[0]
	} else { // try retrieving from closed_sessions within cache
		if sIface, has := engine.Cache.Get(utils.CacheClosedSessions, cgrID); has {
			s = sIface.(*Session)
		}
	}
	if s == nil { // no cached session, CDR will be handled by CDRs
		return sS.cdrS.Call(utils.CDRsV1ProcessEvent,
			&engine.ArgV1ProcessEvent{
				CGREvent:      *cgrEvWithArgDisp.CGREvent,
				ArgDispatcher: cgrEvWithArgDisp.ArgDispatcher}, rply)
	}

	// Use previously stored Session to generate CDRs

	// update stored event with fields out of CDR
	for k, v := range ev.Me {
		if protectedSFlds.HasField(k) {
			continue
		}
		s.EventStart.Set(k, v) // update previoius field with new one
	}
	// create one CGREvent for each session run plus *raw one
	var cgrEvs []*utils.CGREvent
	if cgrEvs, err = s.asCGREvents(); err != nil {
		return utils.NewErrServerError(err)
	}

	var withErrors bool
	for _, cgrEv := range cgrEvs {
		argsProc := &engine.ArgV1ProcessEvent{
			CGREvent:      *cgrEv,
			ChargerS:      utils.BoolPointer(false),
			AttributeS:    utils.BoolPointer(false),
			ArgDispatcher: cgrEvWithArgDisp.ArgDispatcher,
		}
		if mp := engine.NewMapEvent(cgrEv.Event); mp.GetStringIgnoreErrors(utils.RunID) != utils.MetaRaw && // check if is *raw
			unratedReqs.HasField(mp.GetStringIgnoreErrors(utils.RequestType)) { // order additional rating for unrated request types
			argsProc.RALs = utils.BoolPointer(true)
		}
		if err = sS.cdrS.Call(utils.CDRsV1ProcessEvent,
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

// NewV1ProcessEventArgs is a constructor for EventArgs used by ProcessEvent
func NewV1ProcessEventArgs(resrc, acnts, attrs, thds, stats,
	suppls, supplsIgnoreErrs, supplsEventCost bool,
	cgrEv *utils.CGREvent, argDisp *utils.ArgDispatcher,
	supplierPaginator utils.Paginator) (args *V1ProcessEventArgs) {
	args = &V1ProcessEventArgs{
		AllocateResources:     resrc,
		Debit:                 acnts,
		GetAttributes:         attrs,
		ProcessThresholds:     thds,
		ProcessStats:          stats,
		SuppliersIgnoreErrors: supplsIgnoreErrs,
		GetSuppliers:          suppls,
		CGREvent:              cgrEv,
		ArgDispatcher:         argDisp,
	}
	if supplsEventCost {
		args.SuppliersMaxCost = utils.MetaSuppliersEventCost
	}
	args.Paginator = supplierPaginator
	return
}

// V1ProcessEventArgs are the options passed to ProcessEvent API
type V1ProcessEventArgs struct {
	GetAttributes         bool
	AllocateResources     bool
	Debit                 bool
	ProcessThresholds     bool
	ProcessStats          bool
	GetSuppliers          bool
	SuppliersMaxCost      string
	SuppliersIgnoreErrors bool
	*utils.CGREvent
	utils.Paginator
	*utils.ArgDispatcher
}

// V1ProcessEventReply is the reply for the ProcessEvent API
type V1ProcessEventReply struct {
	MaxUsage           *time.Duration
	ResourceAllocation *string
	Attributes         *engine.AttrSProcessEventReply
	Suppliers          *engine.SortedSuppliers
	ThresholdIDs       *[]string
	StatQueueIDs       *[]string
}

// AsNavigableMap is part of engine.NavigableMapper interface
func (v1Rply *V1ProcessEventReply) AsNavigableMap(
	ignr []*config.FCTemplate) (*config.NavigableMap, error) {
	cgrReply := make(map[string]interface{})
	if v1Rply != nil {
		if v1Rply.MaxUsage != nil {
			cgrReply[utils.CapMaxUsage] = *v1Rply.MaxUsage
		}
		if v1Rply.ResourceAllocation != nil {
			cgrReply[utils.CapResourceAllocation] = *v1Rply.ResourceAllocation
		}
		if v1Rply.Attributes != nil {
			attrs := make(map[string]interface{})
			for _, fldName := range v1Rply.Attributes.AlteredFields {
				if v1Rply.Attributes.CGREvent.HasField(fldName) {
					attrs[fldName] = v1Rply.Attributes.CGREvent.Event[fldName]
				}
			}
			cgrReply[utils.CapAttributes] = attrs
		}
		if v1Rply.Suppliers != nil {
			cgrReply[utils.CapSuppliers] = v1Rply.Suppliers.AsNavigableMap()
		}
		if v1Rply.ThresholdIDs != nil {
			cgrReply[utils.CapThresholds] = *v1Rply.ThresholdIDs
		}
		if v1Rply.StatQueueIDs != nil {
			cgrReply[utils.CapStatQueues] = *v1Rply.StatQueueIDs
		}
	}
	return config.NewNavigableMap(cgrReply), nil
}

// BiRPCv1ProcessEvent processes one event with the right subsystems based on arguments received
func (sS *SessionS) BiRPCv1ProcessEvent(clnt rpcclient.RpcClientConnection,
	args *V1ProcessEventArgs, rply *V1ProcessEventReply) (err error) {
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
	me := engine.NewMapEvent(args.CGREvent.Event)
	originID := me.GetStringIgnoreErrors(utils.OriginID)

	if args.GetAttributes {
		if sS.attrS == nil {
			return utils.NewErrNotConnected(utils.AttributeS)
		}
		attrArgs := &engine.AttrArgsProcessEvent{
			Context:       utils.StringPointer(utils.MetaSessionS),
			CGREvent:      args.CGREvent,
			ArgDispatcher: args.ArgDispatcher,
		}
		var rplyEv engine.AttrSProcessEventReply
		if err := sS.attrS.Call(utils.AttributeSv1ProcessEvent,
			attrArgs, &rplyEv); err == nil {
			args.CGREvent = rplyEv.CGREvent
			if tntIface, has := args.CGREvent.Event[utils.MetaTenant]; has {
				// special case when we want to overwrite the tenant
				args.CGREvent.Tenant = tntIface.(string)
				delete(args.CGREvent.Event, utils.MetaTenant)
			}
			rply.Attributes = &rplyEv
		} else if err.Error() != utils.ErrNotFound.Error() {
			return utils.NewErrAttributeS(err)
		}
	}
	if args.AllocateResources {
		if sS.resS == nil {
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
		if err = sS.resS.Call(utils.ResourceSv1AllocateResources,
			attrRU, &allocMessage); err != nil {
			return utils.NewErrResourceS(err)
		}
		rply.ResourceAllocation = &allocMessage
	}
	if args.GetSuppliers {
		if sS.splS == nil {
			return utils.NewErrNotConnected(utils.SupplierS)
		}
		cgrEv := args.CGREvent.Clone()
		if acd, has := cgrEv.Event[utils.ACD]; has {
			cgrEv.Event[utils.Usage] = acd
		}
		var splsReply engine.SortedSuppliers
		sArgs := &engine.ArgsGetSuppliers{
			IgnoreErrors:  args.SuppliersIgnoreErrors,
			MaxCost:       args.SuppliersMaxCost,
			CGREvent:      cgrEv,
			Paginator:     args.Paginator,
			ArgDispatcher: args.ArgDispatcher,
		}
		if err = sS.splS.Call(utils.SupplierSv1GetSuppliers,
			sArgs, &splsReply); err != nil {
			return utils.NewErrSupplierS(err)
		}
		if splsReply.SortedSuppliers != nil {
			rply.Suppliers = &splsReply
		}
	}
	if args.Debit {
		if maxUsage, err := sS.chargeEvent(args.CGREvent.Tenant,
			engine.NewSafEvent(args.CGREvent.Event), args.ArgDispatcher); err != nil {
			return utils.NewErrRALs(err)
		} else {
			rply.MaxUsage = &maxUsage
		}
	}
	if args.ProcessThresholds {
		if sS.thdS == nil {
			return utils.NewErrNotConnected(utils.ThresholdS)
		}
		var tIDs []string
		thEv := &engine.ArgsProcessEvent{
			CGREvent:      args.CGREvent,
			ArgDispatcher: args.ArgDispatcher,
		}
		if err := sS.thdS.Call(utils.ThresholdSv1ProcessEvent,
			thEv, &tIDs); err != nil &&
			err.Error() != utils.ErrNotFound.Error() {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> error: %s processing event %+v with ThresholdS.",
					utils.SessionS, err.Error(), thEv))
		}
		rply.ThresholdIDs = &tIDs
	}
	if args.ProcessStats {
		if sS.statS == nil {
			return utils.NewErrNotConnected(utils.StatS)
		}
		var statReply []string
		statArgs := &engine.StatsArgsProcessEvent{
			CGREvent:      args.CGREvent,
			ArgDispatcher: args.ArgDispatcher,
		}
		if err := sS.statS.Call(utils.StatSv1ProcessEvent,
			statArgs, &statReply); err != nil &&
			err.Error() != utils.ErrNotFound.Error() {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> error: %s processing event %+v with StatS.",
					utils.SessionS, err.Error(), args.CGREvent))
		}
		rply.ThresholdIDs = &statReply
	}
	return nil
}

// BiRPCv1SyncSessions will sync sessions on demand
func (sS *SessionS) BiRPCv1SyncSessions(clnt rpcclient.RpcClientConnection,
	ignParam string, reply *string) error {
	sS.syncSessions()
	*reply = utils.OK
	return nil
}

// BiRPCv1ForceDisconnect will force disconnecting sessions matching sessions
func (sS *SessionS) BiRPCv1ForceDisconnect(clnt rpcclient.RpcClientConnection,
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
		if errTerm := sS.forceSTerminate(ss[0], 0, nil); errTerm != nil {
			utils.Logger.Warning(
				fmt.Sprintf(
					"<%s> failed force-terminating session with id: <%s>, err: <%s>",
					utils.SessionS, ss[0].CGRid(), errTerm.Error()))
			err = utils.ErrPartiallyExecuted
		}
	}
	if err == nil {
		*reply = utils.OK
	} else {
		*reply = err.Error()
	}
	return nil
}

func (sS *SessionS) BiRPCv1RegisterInternalBiJSONConn(clnt rpcclient.RpcClientConnection,
	ign string, reply *string) error {
	sS.RegisterIntBiJConn(clnt)
	*reply = utils.OK
	return nil
}

// BiRPCV1GetMaxUsage returns the maximum usage as seconds, compatible with OpenSIPS 2.3
// DEPRECATED, it will be removed in future versions
func (sS *SessionS) BiRPCV1GetMaxUsage(clnt rpcclient.RpcClientConnection,
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
	if *rply.MaxUsage == time.Duration(-1) {
		*maxUsage = -1.0
	} else {
		*maxUsage = rply.MaxUsage.Seconds()
	}
	return nil
}

// BiRPCV1InitiateSession is called on session start, returns the maximum number of seconds the session can last
// DEPRECATED, it will be removed in future versions
// Kept for compatibility with OpenSIPS 2.3
func (sS *SessionS) BiRPCV1InitiateSession(clnt rpcclient.RpcClientConnection,
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
	if *rply.MaxUsage == time.Duration(-1) {
		*maxUsage = -1.0
	} else {
		*maxUsage = rply.MaxUsage.Seconds()
	}
	return
}

// BiRPCV1UpdateSession processes interim updates, returns remaining duration from the RALs
// DEPRECATED, it will be removed in future versions
// Kept for compatibility with OpenSIPS 2.3
func (sS *SessionS) BiRPCV1UpdateSession(clnt rpcclient.RpcClientConnection,
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
	if *rply.MaxUsage == time.Duration(-1) {
		*maxUsage = -1.0
	} else {
		*maxUsage = rply.MaxUsage.Seconds()
	}
	return
}

// BiRPCV1TerminateSession is called on session end, should stop debit loop
// DEPRECATED, it will be removed in future versions
// Kept for compatibility with OpenSIPS 2.3
func (sS *SessionS) BiRPCV1TerminateSession(clnt rpcclient.RpcClientConnection,
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
func (sS *SessionS) BiRPCV1ProcessCDR(clnt rpcclient.RpcClientConnection,
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
