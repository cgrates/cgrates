/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/

package sessions

import (
	"errors"
	"fmt"
	"math/rand"
	"runtime"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/guardian"

	"github.com/cgrates/cgrates/utils"
)

// NewSessionS constructs  a new SessionS instance
func NewSessionS(cgrCfg *config.CGRConfig,
	dm *engine.DataManager,
	connMgr *engine.ConnManager) *SessionS {
	cgrCfg.SessionSCfg().SessionIndexes.Add(utils.OriginID) // Make sure we have indexing for OriginID since it is a requirement on prefix searching

	return &SessionS{
		cgrCfg:         cgrCfg,
		dm:             dm,
		connMgr:        connMgr,
		biJClnts:       make(map[birpc.ClientConnector]string),
		biJIDs:         make(map[string]*biJClient),
		aSessions:      make(map[string]*Session),
		aSessionsIdx:   make(map[string]map[string]map[string]utils.StringSet),
		aSessionsRIdx:  make(map[string][]*riFieldNameVal),
		pSessions:      make(map[string]*Session),
		pSessionsIdx:   make(map[string]map[string]map[string]utils.StringSet),
		pSessionsRIdx:  make(map[string][]*riFieldNameVal),
		bkpSessionIDs:  make(utils.StringSet),
		removeSsCGRIDs: make(utils.StringSet),
	}
}

// PopulateCtx inserts the ctx parameter to the sessions ctx
func (sS *SessionS) PopulateCtx(ctx *context.Context) {
	sS.ctx = ctx
}

// biJClient contains info we need to reach back a bidirectional json client
type biJClient struct {
	conn  birpc.ClientConnector // connection towards BiJ client
	proto float64               // client protocol version
}

// SessionS represents the session service
type SessionS struct {
	cgrCfg  *config.CGRConfig // Separate from smCfg since there can be multiple
	dm      *engine.DataManager
	connMgr *engine.ConnManager
	ctx     *context.Context

	biJMux   sync.RWMutex                     // mux protecting BI-JSON connections
	biJClnts map[birpc.ClientConnector]string // index BiJSONConnection so we can sync them later
	biJIDs   map[string]*biJClient            // identifiers of bidirectional JSON conns, used to call RPC based on connIDs

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

	bkpSessionIDs     utils.StringSet // keep a record of session cgrids to be stored in dataDB backup
	bkpSessionIDsMux  sync.RWMutex    // prevent concurrency when adding/deleting CGRIDs from map
	removeSsCGRIDs    utils.StringSet // keep a record of session cgrids to be removed from dataDB backup
	removeSsCGRIDsMux sync.RWMutex    // prevent concurrency when adding/deleting CGRIDs from map
	storeSessMux      sync.RWMutex    // protects storeSessions
}

// SyncSessions starts the service and binds it to the listen loop
func (sS *SessionS) SyncSessions(stopChan chan struct{}) {
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
}

// Shutdown is called by engine to clear states
func (sS *SessionS) Shutdown() (err error) {
	if sS.cgrCfg.SessionSCfg().BackupInterval != 0 {
		if _, err := sS.storeSessions(); err != nil {
			utils.Logger.Err(fmt.Sprintf("Backup Sessions error on shutdown: <%v>", err))
		}
	}
	return
}

// OnBiJSONConnect handles new client connections.
func (sS *SessionS) OnBiJSONConnect(c birpc.ClientConnector) {
	nodeID := utils.UUIDSha1Prefix() // connection identifier, should be later updated as login procedure
	sS.biJMux.Lock()
	sS.biJClnts[c] = nodeID
	sS.biJIDs[nodeID] = &biJClient{
		conn:  c,
		proto: sS.cgrCfg.SessionSCfg().ClientProtocol}
	sS.biJMux.Unlock()
}

// OnBiJSONDisconnect handles client disconnects.
func (sS *SessionS) OnBiJSONDisconnect(c birpc.ClientConnector) {
	sS.biJMux.Lock()
	if nodeID, has := sS.biJClnts[c]; has {
		delete(sS.biJClnts, c)
		delete(sS.biJIDs, nodeID)
	}
	sS.biJMux.Unlock()
}

// RegisterIntBiJConn is called on internal BiJ connection towards SessionS
func (sS *SessionS) RegisterIntBiJConn(c birpc.ClientConnector, nodeID string) {
	if nodeID == utils.EmptyString {
		nodeID = sS.cgrCfg.GeneralCfg().NodeID
	}
	sS.biJMux.Lock()
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
func (sS *SessionS) biJClntID(c birpc.ClientConnector) (clntConnID string) {
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
	if opts.HasField(utils.OptsSessionsTTL) {
		ttl, err = opts.GetDuration(utils.OptsSessionsTTL)
	} else if s.OptsStart.HasField(utils.OptsSessionsTTL) {
		ttl, err = s.OptsStart.GetDuration(utils.OptsSessionsTTL)
	} else {
		ttl = sS.cgrCfg.SessionSCfg().SessionTTL
	}
	if err != nil {
		utils.Logger.Warning(
			fmt.Sprintf("<%s>, cannot extract <%s> for session:<%s>, from it's options: <%s>, err: <%s>",
				utils.SessionS, utils.OptsSessionsTTL, s.CGRID, opts, err))
		return
	}
	if ttl == 0 {
		return // nothing to set up
	}
	// random delay computation
	var maxDelay time.Duration
	if opts.HasField(utils.OptsSessionsTTLMaxDelay) {
		maxDelay, err = opts.GetDuration(utils.OptsSessionsTTLMaxDelay)
	} else if s.OptsStart.HasField(utils.OptsSessionsTTLMaxDelay) {
		maxDelay, err = s.OptsStart.GetDuration(utils.OptsSessionsTTLMaxDelay)
	} else if sS.cgrCfg.SessionSCfg().SessionTTLMaxDelay != nil {
		maxDelay = *sS.cgrCfg.SessionSCfg().SessionTTLMaxDelay
	}
	if err != nil {
		utils.Logger.Warning(
			fmt.Sprintf("<%s>, cannot extract <%s> for session:<%s>, from it's options: <%s>, err: <%s>",
				utils.SessionS, utils.OptsSessionsTTLMaxDelay, s.CGRID, opts.String(), err.Error()))
		return
	}
	if maxDelay != 0 {
		ttl += time.Duration(
			rand.Int63n(maxDelay.Milliseconds()) * time.Millisecond.Nanoseconds())
	}
	// LastUsed
	var ttlLastUsed *time.Duration
	if opts.HasField(utils.OptsSessionsTTLLastUsed) {
		ttlLastUsed, err = opts.GetDurationPtr(utils.OptsSessionsTTLLastUsed)
	} else if s.OptsStart.HasField(utils.OptsSessionsTTLLastUsed) {
		ttlLastUsed, err = s.OptsStart.GetDurationPtr(utils.OptsSessionsTTLLastUsed)
	} else {
		ttlLastUsed = sS.cgrCfg.SessionSCfg().SessionTTLLastUsed
	}
	if err != nil {
		utils.Logger.Warning(
			fmt.Sprintf("<%s>, cannot extract <%s> for session:<%s>, from it's options: <%s>, err: <%s>",
				utils.SessionS, utils.OptsSessionsTTLLastUsed, s.CGRID, opts.String(), err.Error()))
		return
	}
	// LastUsage
	var ttlLastUsage *time.Duration
	if opts.HasField(utils.OptsSessionsTTLLastUsage) {
		ttlLastUsage, err = opts.GetDurationPtr(utils.OptsSessionsTTLLastUsage)
	} else if s.OptsStart.HasField(utils.OptsSessionsTTLLastUsage) {
		ttlLastUsage, err = s.OptsStart.GetDurationPtr(utils.OptsSessionsTTLLastUsage)
	} else {
		ttlLastUsage = sS.cgrCfg.SessionSCfg().SessionTTLLastUsage
	}
	if err != nil {
		utils.Logger.Warning(
			fmt.Sprintf("<%s>, cannot extract <%s> for session:<%s>, from it's options: <%s>, err: <%s>",
				utils.SessionS, utils.OptsSessionsTTLLastUsage, s.CGRID, opts.String(), err.Error()))
		return
	}
	// TTLUsage
	var ttlUsage *time.Duration
	if opts.HasField(utils.OptsSessionsTTLUsage) {
		ttlUsage, err = opts.GetDurationPtr(utils.OptsSessionsTTLUsage)
	} else if s.OptsStart.HasField(utils.OptsSessionsTTLUsage) {
		ttlUsage, err = s.OptsStart.GetDurationPtr(utils.OptsSessionsTTLUsage)
	} else {
		ttlUsage = sS.cgrCfg.SessionSCfg().SessionTTLUsage
	}
	if err != nil {
		utils.Logger.Warning(
			fmt.Sprintf("<%s>, cannot extract <%s> for session:<%s>, from it's options: <%s>, err: <%s>",
				utils.SessionS, utils.OptsSessionsTTLUsage, s.CGRID, opts.String(), err.Error()))
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
				s.sTerminator.ttlLastUsed, nil,
				map[string]any{
					utils.DisconnectCause: utils.SessionTimeout})
			s.Unlock()
		case <-endChan:
			timer.Stop()
		}
	}(s.sTerminator.endChan, s.sTerminator.timer)
	runtime.Gosched() // force context switching
}

// forceSTerminate is called when a session times-out or it is forced from CGRateS side
// not thread safe
func (sS *SessionS) forceSTerminate(s *Session, extraUsage time.Duration, tUsage, lastUsed *time.Duration,
	apiOpts, event map[string]any) (err error) {
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
			if err = sS.connMgr.Call(context.TODO(), sS.cgrCfg.SessionSCfg().CDRsConns,
				utils.CDRsV1ProcessEvent, argsProc, &reply); err != nil {
				utils.Logger.Warning(
					fmt.Sprintf(
						"<%s> could not post CDR for event %s, err: %s",
						utils.SessionS, utils.ToJSON(cgrEv), err.Error()))
			}
		}
	}
	// release the resources for the session
	if len(sS.cgrCfg.SessionSCfg().ResourceSConns) != 0 && s.ResourceID != "" {
		var reply string
		cgrEv := &utils.CGREvent{
			Tenant:  s.Tenant,
			ID:      utils.GenUUID(),
			Event:   s.EventStart,
			APIOpts: s.OptsStart,
		}
		if cgrEv.APIOpts == nil {
			cgrEv.APIOpts = make(map[string]any)
		}
		cgrEv.APIOpts[utils.OptsResourcesUsageID] = s.ResourceID
		cgrEv.APIOpts[utils.OptsResourcesUnits] = 1
		cgrEv.SetCloneable(true)
		if err := sS.connMgr.Call(context.TODO(), sS.cgrCfg.SessionSCfg().ResourceSConns,
			utils.ResourceSv1ReleaseResources,
			cgrEv, &reply); err != nil {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> error: %s could not release resource with resourceID: %s",
					utils.SessionS, err.Error(), s.ResourceID))
		}
	}
	if len(sS.cgrCfg.SessionSCfg().IPsConns) != 0 && s.IPAllocID != "" {
		var reply string
		cgrEv := &utils.CGREvent{
			Tenant:  s.Tenant,
			ID:      utils.GenUUID(),
			Event:   s.EventStart,
			APIOpts: s.OptsStart,
		}
		if cgrEv.APIOpts == nil {
			cgrEv.APIOpts = make(map[string]any)
		}
		cgrEv.APIOpts[utils.OptsIPsAllocationID] = s.IPAllocID
		cgrEv.SetCloneable(true)
		if err := sS.connMgr.Call(context.TODO(), sS.cgrCfg.SessionSCfg().IPsConns,
			utils.IPsV1ReleaseIP,
			cgrEv, &reply); err != nil {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> could not release IP allocation %q: %v",
					utils.SessionS, s.IPAllocID, err))
		}
	}
	sS.replicateSessions(s.CGRID, false, sS.cgrCfg.SessionSCfg().ReplicationConns)
	if clnt := sS.biJClnt(s.ClientConnID); clnt != nil {
		go func() {
			// Merge parameter event with the session event. Losing the EventStart OriginID
			// could create unwanted behaviour.
			if event == nil {
				event = make(map[string]any)
			}
			for key, val := range s.EventStart {
				if _, has := event[key]; !has {
					event[key] = val
				}
			}

			// Determine the service method based on the client's protocol version.
			var servMethod string
			switch clnt.proto {
			case 0: // ensure compatibility with OpenSIPS 2.3
				servMethod = "SMGClientV1.DisconnectSession"
			case 1.0: // ensure compatibility with OpenSIPS 3.x versions
				servMethod = "SessionSv1.DisconnectSession"
			case 2.0:
				servMethod = utils.AgentV1DisconnectSession
			}

			// Prepare args based on the client's protocol version.
			var dscErr error
			var rply string
			switch {
			case clnt.proto < 2.0:
				rsn := utils.IfaceAsString(event[utils.DisconnectCause])
				dscArgs := struct {
					EventStart map[string]any
					Reason     string
				}{
					EventStart: s.EventStart,
					Reason:     rsn,
				}
				dscErr = clnt.conn.Call(context.TODO(), servMethod, dscArgs, &rply)
			default:
				dscArgs := utils.CGREvent{
					ID:      utils.GenUUID(),
					Time:    utils.TimePointer(time.Now()),
					APIOpts: apiOpts,
					Event:   event,
				}
				dscErr = clnt.conn.Call(context.TODO(), servMethod, dscArgs, &rply)
			}
			if dscErr != nil && dscErr != utils.ErrNotImplemented {
				utils.Logger.Warning(fmt.Sprintf(
					"<%s> remotely disconnecting session with id <%s> failed: %v",
					utils.SessionS, s.CGRID, dscErr))
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
	if !s.Chargeable {
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
	err = sS.connMgr.Call(context.TODO(), sS.cgrCfg.SessionSCfg().RALsConns,
		utils.ResponderMaxDebit,
		&engine.CallDescriptorWithAPIOpts{
			CallDescriptor: cd,
			APIOpts:        s.OptsStart,
		}, cc)
	if err != nil {
		// verify in case of *dynaprepaid RequestType
		if err.Error() == utils.ErrAccountNotFound.Error() &&
			sr.Event.GetStringIgnoreErrors(utils.RequestType) == utils.MetaDynaprepaid {
			var reply string
			// execute the actionPlan configured in Scheduler
			if err = sS.connMgr.Call(context.TODO(), sS.cgrCfg.SessionSCfg().SchedulerConns,
				utils.SchedulerSv1ExecuteActionPlans, &utils.AttrsExecuteActionPlans{
					ActionPlanIDs: sS.cgrCfg.SchedulerCfg().DynaprepaidActionPlans,
					Tenant:        cd.Tenant, AccountID: cd.Account},
				&reply); err != nil {
				return
			}
			// execute again the MaxDebit operation
			err = sS.connMgr.Call(context.TODO(), sS.cgrCfg.SessionSCfg().RALsConns,
				utils.ResponderMaxDebit,
				&engine.CallDescriptorWithAPIOpts{
					CallDescriptor: cd,
					APIOpts:        s.OptsStart,
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
			// try to disconnect the session n times before we force terminate it on our side
			fib := utils.FibDuration(time.Millisecond, 0)
			for i := 0; i < sS.cgrCfg.SessionSCfg().TerminateAttempts; i++ {
				if i != 0 { // not the first iteration
					time.Sleep(fib())
				}
				if err = sS.disconnectSession(s, dscReason); err == nil {
					s.Unlock()
					return
				}
				utils.Logger.Warning(
					fmt.Sprintf("<%s> could not disconnect session: %s, error: %s",
						utils.SessionS, s.cgrID(), err.Error()))
			}
			if err = sS.forceSTerminate(s, 0, nil, nil, nil,
				map[string]any{utils.DisconnectCause: utils.ForcedDisconnect}); err != nil {
				utils.Logger.Warning(
					fmt.Sprintf("<%s> failed force-terminating session: <%s>, err: <%s>",
						utils.SessionS, s.cgrID(), err))
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
				// try to disconnect the session n times before we force terminate it on our side
				fib := utils.FibDuration(time.Millisecond, 0)
				for i := 0; i < sS.cgrCfg.SessionSCfg().TerminateAttempts; i++ {
					if i != 0 { // not the first iteration
						time.Sleep(fib())
					}
					if err = sS.disconnectSession(s, utils.ErrInsufficientCredit.Error()); err == nil {
						return
					}
					utils.Logger.Warning(
						fmt.Sprintf("<%s> could not disconnect session: %s, error: %s",
							utils.SessionS, s.cgrID(), err.Error()))
				}
				utils.Logger.Warning(
					fmt.Sprintf("<%s> could not disconnect session: <%s>, error: <%s>",
						utils.SessionS, s.cgrID(), err.Error()))
				if err = sS.forceSTerminate(s, 0, nil, nil, nil,
					map[string]any{utils.DisconnectCause: utils.ForcedDisconnect}); err != nil {
					utils.Logger.Warning(
						fmt.Sprintf("<%s> failed force-terminating session: <%s>, err: <%s>",
							utils.SessionS, s.cgrID(), err))
				}
			}
			return
		}
		s.Lock()
		s.UpdatedAt = time.Now()
		s.Unlock()
		if sS.cgrCfg.SessionSCfg().BackupInterval > 0 {
			sS.bkpSessionIDsMux.Lock()
			sS.bkpSessionIDs.Add(s.CGRID)
			sS.bkpSessionIDsMux.Unlock()
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
	if err = sS.connMgr.Call(context.TODO(), sS.cgrCfg.SessionSCfg().RALsConns,
		utils.ResponderRefundIncrements,
		&engine.CallDescriptorWithAPIOpts{
			CallDescriptor: cd,
			APIOpts:        s.OptsStart,
		}, &acnt); err != nil {
		return
	}
	if acnt.ID != "" { // Account info updated, update also cached AccountSummary
		acntSummary := acnt.AsAccountSummary()
		acntSummary.UpdateInitialValue(sr.EventCost.AccountSummary)
		sr.EventCost.AccountSummary = acntSummary
	}
	s.UpdatedAt = time.Now()
	if sS.cgrCfg.SessionSCfg().BackupInterval > 0 {
		sS.bkpSessionIDsMux.Lock()
		sS.bkpSessionIDs.Add(s.CGRID)
		sS.bkpSessionIDsMux.Unlock()
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
		APIOpts:        s.OptsStart,
		Tenant:         s.Tenant,
	}
	var reply string
	// use the v1 because it doesn't do rounding refund
	if err = sS.connMgr.Call(context.TODO(), sS.cgrCfg.SessionSCfg().CDRsConns,
		utils.CDRsV1StoreSessionCost,
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
	}
	return
}

// roundCost will round the EventCost and will refund the extra debited increments
// should be called only at the endSession
// not thread safe, need to be handled in a layer above
func (sS *SessionS) roundCost(s *Session, sRunIdx int) (err error) {
	sr := s.SRuns[sRunIdx]
	runID := sr.Event.GetStringIgnoreErrors(utils.RunID)
	cc := sr.EventCost.AsCallCost(utils.EmptyString)
	if sr.CD != nil {
		cc.Category = sr.CD.Category
		cc.Subject = sr.CD.Subject
		cc.Tenant = sr.CD.Tenant
		cc.Account = sr.CD.Account
		cc.Destination = sr.CD.Destination
		cc.ToR = sr.CD.ToR
	}
	cc.Round()
	if roundIncrements := cc.GetRoundIncrements(); len(roundIncrements) != 0 {
		cd := cc.CreateCallDescriptor()
		cd.CgrID = s.CGRID
		cd.RunID = runID
		cd.Increments = roundIncrements
		response := new(engine.Account)
		if err = sS.connMgr.Call(context.TODO(), sS.cgrCfg.SessionSCfg().RALsConns,
			utils.ResponderRefundRounding,
			&engine.CallDescriptorWithAPIOpts{CallDescriptor: cd},
			response); err != nil {
			return
		}
		accSum := response.AsAccountSummary()
		accSum.UpdateInitialValue(cc.AccountSummary)
		cc.AccountSummary = accSum
	}
	sr.EventCost = engine.NewEventCostFromCallCost(cc, s.CGRID, runID)
	return
}

// disconnectSession sends a disconnect request to the client associated with the given session.
// Note: This function is not thread-safe and assumes the session is already stopped.
func (s *SessionS) disconnectSession(sess *Session, rsn string) (err error) {
	clnt := s.biJClnt(sess.ClientConnID)
	if clnt == nil {
		return fmt.Errorf("calling %s requires bidirectional JSON connection, connID: <%s>",
			utils.AgentV1DisconnectSession, sess.ClientConnID)
	}
	sess.EventStart[utils.Usage] = sess.totalUsage() // Set the usage to total one debitted

	// Determine the service method based on the client's protocol version.
	var servMethod string
	switch clnt.proto {
	case 0: // compatibility with OpenSIPS 2.3
		servMethod = "SMGClientV1.DisconnectSession"
	case 1.0: // compatibility with OpenSIPS 3.x versions
		servMethod = "SessionSv1.DisconnectSession"
	case 2.0:
		servMethod = utils.AgentV1DisconnectSession
	}

	// Prepare args based on the client's protocol version.
	var rply string
	switch {
	case clnt.proto < 2.0:
		dscArgs := struct {
			EventStart map[string]any
			Reason     string
		}{
			EventStart: sess.EventStart,
			Reason:     rsn,
		}
		err = clnt.conn.Call(context.TODO(), servMethod, dscArgs, &rply)
	case clnt.proto == 2.0:
		dscArgs := utils.CGREvent{
			ID:    utils.GenUUID(),
			Time:  utils.TimePointer(time.Now()),
			Event: sess.EventStart,
		}
		dscArgs.Event[utils.DisconnectCause] = rsn
		err = clnt.conn.Call(context.TODO(), servMethod, dscArgs, &rply)
	}
	if err != nil && err != utils.ErrNotImplemented {
		return err
	}
	return nil
}

// warnSession will send warning from SessionS to clients
// regarding low balance
func (sS *SessionS) warnSession(connID string, ev map[string]any) (err error) {
	clnt := sS.biJClnt(connID)
	if clnt == nil {
		return fmt.Errorf("calling %s requires bidirectional JSON connection, connID: <%s>",
			utils.AgentV1WarnDisconnect, connID)
	}
	var rply string
	if err = clnt.conn.Call(context.TODO(), utils.AgentV1WarnDisconnect,
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
func (sS *SessionS) replicateSessions(cgrID string, psv bool, connIDs []string) {
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
		if err := sS.connMgr.Call(context.TODO(), connIDs,
			utils.SessionSv1SetPassiveSession,
			sCln, &rply); err != nil {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> cannot replicate session with id <%s>, err: %s",
					utils.SessionS, sCln.CGRID, err.Error()))
		}
	}
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
	s.UpdatedAt = time.Now()
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
	if !passive && sS.cgrCfg.SessionSCfg().BackupInterval > 0 {
		sS.bkpSessionIDsMux.Lock()
		delete(sS.bkpSessionIDs, cgrID) // in case not yet in backup, dont needlessly store session
		sS.bkpSessionIDsMux.Unlock()
		sS.removeSsCGRIDsMux.Lock()
		sS.removeSsCGRIDs.Add(cgrID)
		sS.removeSsCGRIDsMux.Unlock()
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
			splitFieldName := utils.SplitPath(fieldName, utils.NestingSep[0], -1)
			fieldName := splitFieldName[len(splitFieldName)-1] // take only the last field name from the slice
			fieldVal, err := sr.Event.GetString(fieldName)     // the only error from GetString is ErrNotFound
			if err != nil {
				fieldVal = utils.NotAvailable
			}
			if fieldVal == utils.EmptyString {
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
func (sS *SessionS) newSession(cgrEv *utils.CGREvent, originID, clntConnID string,
	dbtItval time.Duration, forceDuration, isMsg bool) (s *Session, err error) {
	if len(sS.cgrCfg.SessionSCfg().ChargerSConns) == 0 {
		err = errors.New("ChargerS is disabled")
		return
	}
	cgrID := GetSetCGRID(cgrEv.Event)
	evStart := engine.MapEvent(cgrEv.Event)
	if !evStart.HasField(utils.Usage) && evStart.HasField(utils.LastUsed) {
		evStart[utils.Usage] = evStart[utils.LastUsed]
	}
	s = &Session{
		CGRID:         cgrID,
		Tenant:        cgrEv.Tenant,
		IPAllocID:     originID,
		ResourceID:    originID,
		EventStart:    evStart.Clone(), // decouple the event from the request so we can avoid concurrency with debit and ttl
		OptsStart:     engine.MapEvent(cgrEv.APIOpts).Clone(),
		ClientConnID:  clntConnID,
		DebitInterval: dbtItval,
	}
	s.Chargeable = s.OptsStart.GetBoolOrDefault(utils.OptsChargeable, true)
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
				Tenant:        chrgr.CGREvent.Tenant,
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
	if err = sS.connMgr.Call(context.TODO(), sS.cgrCfg.SessionSCfg().ChargerSConns,
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
	s = sS.transitSState(cgrID, false)
	if len(ss) != 0 && sS.cgrCfg.SessionSCfg().BackupInterval > 0 {
		sS.bkpSessionIDsMux.Lock()
		sS.bkpSessionIDs.Add(s.CGRID)
		sS.bkpSessionIDsMux.Unlock()
	}
	return
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
	if sS.cgrCfg.SessionSCfg().BackupInterval > 0 {
		sS.bkpSessionIDsMux.Lock()
		sS.bkpSessionIDs.Add(s.CGRID)
		sS.bkpSessionIDsMux.Unlock()
	}
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
	sS.aSsMux.RLock()
	asCount := len(sS.aSessions)
	sS.aSsMux.RUnlock()
	if asCount == 0 { // no need to sync the sessions if none is active
		return
	}
	queriedCGRIDs := engine.NewSafEvent(nil) // will populate from all goroutines at once
	var wg sync.WaitGroup

	for _, clnt := range sS.biJClients() {
		wg.Add(1)
		go func() {
			// query all connections at once
			servMethod := utils.AgentV1GetActiveSessionIDs
			if clnt.proto < 2.0 {
				// ensure compatibility with OpenSIPS
				servMethod = "SessionSv1.GetActiveSessionIDs"
			}

			var queriedSessionIDs []*SessionID
			if err := clnt.conn.Call(context.TODO(), servMethod, utils.EmptyString,
				&queriedSessionIDs); err != nil &&
				err.Error() != utils.ErrNoActiveSession.Error() {
				utils.Logger.Warning(fmt.Sprintf(
					"<%s> error <%v> querying session ids", utils.SessionS, err))
			}

			for _, sessionID := range queriedSessionIDs {
				queriedCGRIDs.Set(sessionID.CGRID(), struct{}{})
			}
			wg.Done()
		}()
	}

	wg.Wait() // wait for all clients to finish in one way or another
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
		var eUsage time.Duration
		if sS.cgrCfg.SessionSCfg().StaleChanMaxExtraUsage > 0 { // add extra usage
			eUsage += time.Duration(
				rand.Int63n(sS.cgrCfg.SessionSCfg().StaleChanMaxExtraUsage.Milliseconds()) * time.Millisecond.Nanoseconds())
		}
		ss[0].Lock()
		if err := sS.forceSTerminate(ss[0], eUsage, nil, nil, nil,
			map[string]any{utils.DisconnectCause: utils.ForcedDisconnect}); err != nil {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> failed force-terminating session: <%s>, err: <%v>",
					utils.SessionS, cgrID, err))
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
		if (s.DebitInterval > 0 &&
			sr.Event.GetStringIgnoreErrors(utils.RequestType) == utils.MetaPrepaid) ||
			(s.DebitInterval > 0 && sr.Event.GetStringIgnoreErrors(utils.RequestType) ==
				utils.MetaDynaprepaid) {
			if s.debitStop == nil { // init the debitStop only for the first sRun with DebitInterval and RequestType MetaPrepaids.DebitInterval > 0 &&
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
		eventUsage = sS.cgrCfg.SessionSCfg().GetDefaultUsage(evStart.GetStringIgnoreErrors(utils.ToR))
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
		} else if err = sS.connMgr.Call(context.TODO(), sS.cgrCfg.SessionSCfg().RALsConns,
			utils.ResponderGetMaxSessionTime,
			&engine.CallDescriptorWithAPIOpts{
				CallDescriptor: sr.CD,
				APIOpts:        s.OptsStart,
			}, &rplyMaxUsage); err != nil {
			if err.Error() == utils.ErrAccountNotFound.Error() &&
				sr.Event.GetStringIgnoreErrors(utils.RequestType) == utils.MetaDynaprepaid {
				var reply string
				// execute the actionPlan configured in Scheduler
				if err = sS.connMgr.Call(context.TODO(), sS.cgrCfg.SessionSCfg().SchedulerConns,
					utils.SchedulerSv1ExecuteActionPlans, &utils.AttrsExecuteActionPlans{
						ActionPlanIDs: sS.cgrCfg.SchedulerCfg().DynaprepaidActionPlans,
						Tenant:        sr.CD.Tenant,
						AccountID:     sr.CD.Account},
					&reply); err != nil {
					return
				}
				// execute again the GetMaxSessionTime operation
				err = sS.connMgr.Call(context.TODO(), sS.cgrCfg.SessionSCfg().RALsConns,
					utils.ResponderGetMaxSessionTime,
					&engine.CallDescriptorWithAPIOpts{
						CallDescriptor: sr.CD,
						APIOpts:        s.OptsStart,
					}, &rplyMaxUsage)
			}
			if err != nil {
				err = utils.NewErrRALs(err)
				return
			}
		}
		if rplyMaxUsage > eventUsage {
			rplyMaxUsage = eventUsage
		}
		usage[sr.CD.RunID] = rplyMaxUsage
	}
	return
}

// restoreSessions reinitiates sessions stored on dataDB backup
// no session protection needed since it runs only once at start of service,
// before the start modifying/creating sessions
func (sS *SessionS) restoreSessions(sessions []*Session) {
	for _, s := range sessions {
		tor, _ := s.EventStart[utils.ToR].(string)
		if tor == utils.EmptyString {
			tor = utils.MetaVoice
		}
		if time.Since(s.UpdatedAt) <= sS.cgrCfg.SessionSCfg().DefaultUsage[tor] {
			sS.initSessionDebitLoops(s)
			sS.registerSession(s, false)
		} else { // remove expired sessions from dataDB
			sS.removeSsCGRIDsMux.Lock()
			sS.removeSsCGRIDs.Add(s.CGRID)
			sS.removeSsCGRIDsMux.Unlock()
		}
	}
}

// initSession handles a new session
// not thread-safe for Session since it is constructed here
func (sS *SessionS) initSession(cgrEv *utils.CGREvent, clntConnID, originID string,
	dbtItval time.Duration, isMsg, forceDuration bool) (s *Session, err error) {
	if s, err = sS.newSession(cgrEv, originID, clntConnID, dbtItval, forceDuration, isMsg); err != nil {
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
	s.Chargeable = opts.GetBoolOrDefault(utils.OptsChargeable, true)
	s.UpdatedAt = time.Now()
	//init has no updtEv
	if updtEv == nil {
		updtEv = engine.MapEvent(s.EventStart.Clone())
	}

	var reqMaxUsage time.Duration
	if reqMaxUsage, err = updtEv.GetDuration(utils.Usage); err != nil {
		if err != utils.ErrNotFound {
			return
		}

		reqMaxUsage = sS.cgrCfg.SessionSCfg().GetDefaultUsage(updtEv.GetStringIgnoreErrors(utils.ToR))
		updtEv[utils.Usage] = reqMaxUsage
	}
	lastUsed := updtEv.GetDurationPtrIgnoreErrors(utils.LastUsed)

	var totalUsage time.Duration
	if totalUsage, err = updtEv.GetDuration(utils.TotalUsage); err != nil {
		if err != utils.ErrNotFound {
			return
		}
		err = nil
	} else {
		reqMaxUsage, lastUsed = s.midSessionUsage(totalUsage)
	}

	maxUsage = make(map[string]time.Duration)
	for i, sr := range s.SRuns {
		reqType := sr.Event.GetStringIgnoreErrors(utils.RequestType)
		var rplyMaxUsage time.Duration
		switch reqType {
		case utils.MetaPrepaid, utils.MetaDynaprepaid:
			if s.debitStop == nil {
				if rplyMaxUsage, err = sS.debitSession(s, i,
					reqMaxUsage, lastUsed); err != nil {
					return
				}
				break
			}
			rplyMaxUsage = reqMaxUsage
		case utils.MetaPseudoPrepaid:
			if err = sS.connMgr.Call(context.TODO(), sS.cgrCfg.SessionSCfg().RALsConns,
				utils.ResponderGetMaxSessionTime,
				&engine.CallDescriptorWithAPIOpts{
					CallDescriptor: sr.CD,
					APIOpts:        s.OptsStart,
				}, &rplyMaxUsage); err != nil {
				return
			}
		default:
			rplyMaxUsage = reqMaxUsage
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
			// if !isMsg { // in case of one time charge there is no need of corrections
			if notCharged := sUsage - sr.EventCost.GetUsage(); notCharged > 0 { // we did not charge enough, make a manual debit here
				if !s.Chargeable {
					sS.pause(sr, notCharged)
				} else {
					if sr.CD.LoopIndex > 0 {
						sr.CD.TimeStart = sr.CD.TimeEnd
					}
					sr.CD.TimeEnd = sr.CD.TimeStart.Add(notCharged)
					sr.CD.DurationIndex += notCharged
					cc := new(engine.CallCost)
					if err = sS.connMgr.Call(context.TODO(), sS.cgrCfg.SessionSCfg().RALsConns,
						utils.ResponderDebit,
						&engine.CallDescriptorWithAPIOpts{
							CallDescriptor: sr.CD,
							APIOpts:        s.OptsStart,
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
			// }
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
			sr.Event[utils.CostDetails] = utils.ToJSON(sr.EventCost) // avoid map[string]any when decoding
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
	if sS.cgrCfg.SessionSCfg().BackupInterval > 0 {
		sS.bkpSessionIDsMux.Lock()
		delete(sS.bkpSessionIDs, s.CGRID) // in case not yet in backup, dont needlessly store session
		sS.bkpSessionIDsMux.Unlock()
		sS.removeSsCGRIDsMux.Lock()
		sS.removeSsCGRIDs.Add(s.CGRID)
		sS.removeSsCGRIDsMux.Unlock()
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
	if slices.Contains(utils.PostPaidRatedSlice, ev.GetStringIgnoreErrors(utils.RequestType)) {
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

// BiRPCv1GetActiveSessions returns the list of active sessions based on filter
func (sS *SessionS) BiRPCv1GetActiveSessions(ctx *context.Context,
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
func (sS *SessionS) BiRPCv1GetActiveSessionsCount(ctx *context.Context,
	args *utils.SessionFilter, reply *int) error {
	if args == nil { //protection in case on nil
		args = &utils.SessionFilter{}
	}
	*reply = sS.filterSessionsCount(args, false)
	return nil
}

// BiRPCv1GetPassiveSessions returns the passive sessions handled by SessionS
func (sS *SessionS) BiRPCv1GetPassiveSessions(ctx *context.Context,
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
func (sS *SessionS) BiRPCv1GetPassiveSessionsCount(ctx *context.Context,
	args *utils.SessionFilter, reply *int) error {
	if args == nil { //protection in case on nil
		args = &utils.SessionFilter{}
	}
	*reply = sS.filterSessionsCount(args, true)
	return nil
}

// BiRPCv1SetPassiveSession used for replicating Sessions
func (sS *SessionS) BiRPCv1SetPassiveSession(ctx *context.Context,
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
func (sS *SessionS) BiRPCv1ReplicateSessions(ctx *context.Context,
	args ArgsReplicateSessions, reply *string) (err error) {
	sS.replicateSessions(args.CGRID, args.Passive, args.ConnIDs)
	*reply = utils.OK
	return
}

// NewV1AuthorizeArgs is a constructor for V1AuthorizeArgs
func NewV1AuthorizeArgs(attrs bool, attributeIDs []string,
	thrslds bool, thresholdIDs []string, statQueues bool, statIDs []string,
	ips, res, maxUsage, routes, routesIgnoreErrs, routesEventCost bool,
	cgrEv *utils.CGREvent, routePaginator utils.Paginator,
	forceDuration bool, routesMaxCost string) (args *V1AuthorizeArgs) {
	args = &V1AuthorizeArgs{
		GetAttributes:      attrs,
		AuthorizeIP:        ips,
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
	AuthorizeIP        bool
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

func (V1AuthorizeArgs) RPCClone() {}

// ParseFlags will populate the V1AuthorizeArgs flags
func (args *V1AuthorizeArgs) ParseFlags(flags, sep string) {
	for _, subsystem := range strings.Split(flags, sep) {
		switch {
		case subsystem == utils.MetaAccounts:
			args.GetMaxUsage = true
		case subsystem == utils.MetaIPs:
			args.AuthorizeIP = true
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
	args.Paginator, _ = utils.GetRoutePaginatorFromOpts(args.APIOpts)
}

// V1AuthorizeReply are options available in auth reply
type V1AuthorizeReply struct {
	Attributes         *engine.AttrSProcessEventReply `json:",omitempty"`
	AllocatedIP        *engine.AllocatedIP            `json:",omitempty"`
	ResourceAllocation *string                        `json:",omitempty"`
	MaxUsage           *time.Duration                 `json:",omitempty"`
	RouteProfiles      engine.SortedRoutesList        `json:",omitempty"`
	ThresholdIDs       *[]string                      `json:",omitempty"`
	StatQueueIDs       *[]string                      `json:",omitempty"`

	needsMaxUsage bool // for gob encoding only
}

// SetMaxUsageNeeded used by agent that use the reply as NavigableMapper
// only used for gob encoding
func (r *V1AuthorizeReply) SetMaxUsageNeeded(getMaxUsage bool) {
	if r == nil {
		return
	}
	r.needsMaxUsage = getMaxUsage
}

// AsNavigableMap is part of engine.NavigableMapper interface
func (r *V1AuthorizeReply) AsNavigableMap() map[string]*utils.DataNode {
	cgrReply := make(map[string]*utils.DataNode)
	if r.Attributes != nil {
		attrs := &utils.DataNode{Type: utils.NMMapType, Map: make(map[string]*utils.DataNode)}
		for _, fldName := range r.Attributes.AlteredFields {
			fldName = strings.TrimPrefix(fldName, utils.MetaReq+utils.NestingSep)
			if r.Attributes.CGREvent.HasField(fldName) {
				attrs.Map[fldName] = utils.NewLeafNode(r.Attributes.CGREvent.Event[fldName])
			}
		}
		cgrReply[utils.CapAttributes] = attrs
	}
	if r.AllocatedIP != nil {
		cgrReply[utils.CapAllocatedIP] = &utils.DataNode{
			Type: utils.NMMapType,
			Map:  r.AllocatedIP.AsNavigableMap(),
		}
	}
	if r.ResourceAllocation != nil {
		cgrReply[utils.CapResourceAllocation] = utils.NewLeafNode(*r.ResourceAllocation)
	}
	if r.MaxUsage != nil {
		cgrReply[utils.CapMaxUsage] = utils.NewLeafNode(*r.MaxUsage)
	} else if r.needsMaxUsage {
		cgrReply[utils.CapMaxUsage] = utils.NewLeafNode(0)
	}

	if r.RouteProfiles != nil {
		nm := r.RouteProfiles.AsNavigableMap()
		cgrReply[utils.CapRouteProfiles] = nm
	}
	if r.ThresholdIDs != nil {
		thIDs := &utils.DataNode{Type: utils.NMSliceType, Slice: make([]*utils.DataNode, len(*r.ThresholdIDs))}
		for i, v := range *r.ThresholdIDs {
			thIDs.Slice[i] = utils.NewLeafNode(v)
		}
		cgrReply[utils.CapThresholds] = thIDs
	}
	if r.StatQueueIDs != nil {
		stIDs := &utils.DataNode{Type: utils.NMSliceType, Slice: make([]*utils.DataNode, len(*r.StatQueueIDs))}
		for i, v := range *r.StatQueueIDs {
			stIDs.Slice[i] = utils.NewLeafNode(v)
		}
		cgrReply[utils.CapStatQueues] = stIDs
	}
	return cgrReply
}

// BiRPCv1AuthorizeEvent performs authorization for CGREvent based on specific components
func (sS *SessionS) BiRPCv1AuthorizeEvent(ctx *context.Context,
	args *V1AuthorizeArgs, authReply *V1AuthorizeReply) (err error) {
	if args.CGREvent == nil {
		return utils.NewErrMandatoryIeMissing(utils.CGREventString)
	}
	if args.CGREvent.Event == nil {
		return utils.NewErrMandatoryIeMissing(utils.Event)
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

	if !args.GetAttributes && !args.AuthorizeResources && !args.AuthorizeIP &&
		!args.GetMaxUsage && !args.GetRoutes {
		return utils.NewErrMandatoryIeMissing(utils.Subsystems)
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
		if len(sS.cgrCfg.SessionSCfg().ResourceSConns) == 0 {
			return utils.NewErrNotConnected(utils.ResourceS)
		}
		originID, _ := args.CGREvent.FieldAsString(utils.OriginID)
		if originID == "" {
			originID = utils.UUIDSha1Prefix()
		}
		var allocMsg string
		if args.APIOpts == nil {
			args.APIOpts = make(map[string]any)
		}
		args.CGREvent.APIOpts[utils.OptsResourcesUsageID] = originID
		args.CGREvent.APIOpts[utils.OptsResourcesUnits] = 1
		if err = sS.connMgr.Call(context.TODO(), sS.cgrCfg.SessionSCfg().ResourceSConns, utils.ResourceSv1AuthorizeResources,
			args.CGREvent, &allocMsg); err != nil {
			return utils.NewErrResourceS(err)
		}
		authReply.ResourceAllocation = &allocMsg
	}
	if args.AuthorizeIP {
		if len(sS.cgrCfg.SessionSCfg().IPsConns) == 0 {
			return utils.NewErrNotConnected(utils.IPs)
		}
		originID, _ := args.CGREvent.FieldAsString(utils.OriginID)
		if originID == "" {
			originID = utils.UUIDSha1Prefix()
		}
		var allocIP engine.AllocatedIP
		if args.APIOpts == nil {
			args.APIOpts = make(map[string]any)
		}
		args.CGREvent.APIOpts[utils.OptsIPsAllocationID] = originID
		if err = sS.connMgr.Call(context.TODO(), sS.cgrCfg.SessionSCfg().IPsConns, utils.IPsV1AuthorizeIP,
			args.CGREvent, &allocIP); err != nil {
			return utils.NewErrIPs(err)
		}
		authReply.AllocatedIP = &allocIP
	}
	if args.GetRoutes {
		routesReply, err := sS.getRoutes(args.CGREvent.Clone(), args.Paginator,
			args.RoutesIgnoreErrors, args.RoutesMaxCost, false)
		if err != nil {
			return err
		}
		if routesReply != nil {
			authReply.RouteProfiles = routesReply
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
	AllocatedIP        *string
	ResourceAllocation *string
	MaxUsage           float64 // special treat returning time.Duration.Seconds()
	RoutesDigest       *string
	Thresholds         *string
	StatQueues         *string
}

// BiRPCv1AuthorizeEventWithDigest performs authorization for CGREvent based on specific components
// returning one level fields instead of multiple ones returned by BiRPCv1AuthorizeEvent
func (sS *SessionS) BiRPCv1AuthorizeEventWithDigest(ctx *context.Context,
	args *V1AuthorizeArgs, authReply *V1AuthorizeReplyWithDigest) (err error) {
	var initAuthRply V1AuthorizeReply
	if err = sS.BiRPCv1AuthorizeEvent(ctx, args, &initAuthRply); err != nil {
		return
	}
	if args.GetAttributes && initAuthRply.Attributes != nil {
		authReply.AttributesDigest = utils.StringPointer(initAuthRply.Attributes.Digest())
	}
	if args.AuthorizeResources {
		authReply.ResourceAllocation = initAuthRply.ResourceAllocation
	}
	if args.AuthorizeIP {
		authReply.AllocatedIP = utils.StringPointer(initAuthRply.AllocatedIP.Digest())
	}
	if args.GetMaxUsage {
		authReply.MaxUsage = initAuthRply.MaxUsage.Seconds()
	}
	if args.GetRoutes {
		authReply.RoutesDigest = utils.StringPointer(initAuthRply.RouteProfiles.Digest(false))
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
	resrc, ips, acnt bool, cgrEv *utils.CGREvent, forceDuration bool) (args *V1InitSessionArgs) {
	args = &V1InitSessionArgs{
		GetAttributes:     attrs,
		AllocateIP:        ips,
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
	AllocateIP        bool
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

func (V1InitSessionArgs) RPCClone() {}

// ParseFlags will populate the V1InitSessionArgs flags
func (args *V1InitSessionArgs) ParseFlags(flags, sep string) {
	for _, subsystem := range strings.Split(flags, sep) {
		switch {
		case subsystem == utils.MetaAccounts:
			args.InitSession = true
		case subsystem == utils.MetaResources:
			args.AllocateResources = true
		case subsystem == utils.MetaIPs:
			args.AllocateIP = true
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
	Attributes         *engine.AttrSProcessEventReply `json:",omitempty"`
	AllocatedIP        *engine.AllocatedIP            `json:",omitempty"`
	ResourceAllocation *string                        `json:",omitempty"`
	MaxUsage           *time.Duration                 `json:",omitempty"`
	ThresholdIDs       *[]string                      `json:",omitempty"`
	StatQueueIDs       *[]string                      `json:",omitempty"`

	needsMaxUsage bool // for gob encoding only
}

// SetMaxUsageNeeded used by agent that use the reply as NavigableMapper
// only used for gob encoding
func (v1Rply *V1InitSessionReply) SetMaxUsageNeeded(getMaxUsage bool) {
	if v1Rply == nil {
		return
	}
	v1Rply.needsMaxUsage = getMaxUsage
}

// AsNavigableMap is part of engine.NavigableMapper interface
func (r *V1InitSessionReply) AsNavigableMap() map[string]*utils.DataNode {
	cgrReply := make(map[string]*utils.DataNode)
	if r.Attributes != nil {
		attrs := &utils.DataNode{Type: utils.NMMapType, Map: make(map[string]*utils.DataNode)}
		for _, fldName := range r.Attributes.AlteredFields {
			fldName = strings.TrimPrefix(fldName, utils.MetaReq+utils.NestingSep)
			if r.Attributes.CGREvent.HasField(fldName) {
				attrs.Map[fldName] = utils.NewLeafNode(r.Attributes.CGREvent.Event[fldName])
			}
		}
		cgrReply[utils.CapAttributes] = attrs
	}
	if r.AllocatedIP != nil {
		cgrReply[utils.CapAllocatedIP] = &utils.DataNode{
			Type: utils.NMMapType,
			Map:  r.AllocatedIP.AsNavigableMap(),
		}
	}
	if r.ResourceAllocation != nil {
		cgrReply[utils.CapResourceAllocation] = utils.NewLeafNode(*r.ResourceAllocation)
	}
	if r.MaxUsage != nil {
		cgrReply[utils.CapMaxUsage] = utils.NewLeafNode(*r.MaxUsage)
	} else if r.needsMaxUsage {
		cgrReply[utils.CapMaxUsage] = utils.NewLeafNode(0)
	}

	if r.ThresholdIDs != nil {
		thIDs := &utils.DataNode{Type: utils.NMSliceType, Slice: make([]*utils.DataNode, len(*r.ThresholdIDs))}
		for i, v := range *r.ThresholdIDs {
			thIDs.Slice[i] = utils.NewLeafNode(v)
		}
		cgrReply[utils.CapThresholds] = thIDs
	}
	if r.StatQueueIDs != nil {
		stIDs := &utils.DataNode{Type: utils.NMSliceType, Slice: make([]*utils.DataNode, len(*r.StatQueueIDs))}
		for i, v := range *r.StatQueueIDs {
			stIDs.Slice[i] = utils.NewLeafNode(v)
		}
		cgrReply[utils.CapStatQueues] = stIDs
	}
	return cgrReply
}

// BiRPCv1InitiateSession initiates a new session
func (sS *SessionS) BiRPCv1InitiateSession(ctx *context.Context,
	args *V1InitSessionArgs, rply *V1InitSessionReply) (err error) {
	if args.CGREvent == nil {
		return utils.NewErrMandatoryIeMissing(utils.CGREventString)
	}
	if args.CGREvent.Event == nil {
		return utils.NewErrMandatoryIeMissing(utils.Event)
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

	if !args.GetAttributes && !args.AllocateResources && !args.AllocateIP && !args.InitSession {
		return utils.NewErrMandatoryIeMissing(utils.Subsystems)
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
		if len(sS.cgrCfg.SessionSCfg().ResourceSConns) == 0 {
			return utils.NewErrNotConnected(utils.ResourceS)
		}
		if originID == "" {
			return utils.NewErrMandatoryIeMissing(utils.OriginID)
		}
		if args.APIOpts == nil {
			args.APIOpts = make(map[string]any)
		}
		args.CGREvent.APIOpts[utils.OptsResourcesUsageID] = originID
		args.CGREvent.APIOpts[utils.OptsResourcesUnits] = 1
		var allocMessage string
		if err = sS.connMgr.Call(context.TODO(), sS.cgrCfg.SessionSCfg().ResourceSConns,
			utils.ResourceSv1AllocateResources, args.CGREvent, &allocMessage); err != nil {
			return utils.NewErrResourceS(err)
		}
		rply.ResourceAllocation = &allocMessage
	}
	if args.AllocateIP {
		if len(sS.cgrCfg.SessionSCfg().IPsConns) == 0 {
			return utils.NewErrNotConnected(utils.IPs)
		}
		if originID == "" {
			return utils.NewErrMandatoryIeMissing(utils.OriginID)
		}
		if args.APIOpts == nil {
			args.APIOpts = make(map[string]any)
		}
		args.CGREvent.APIOpts[utils.OptsIPsAllocationID] = originID
		var allocIP engine.AllocatedIP
		if err = sS.connMgr.Call(context.TODO(), sS.cgrCfg.SessionSCfg().IPsConns,
			utils.IPsV1AllocateIP, args.CGREvent, &allocIP); err != nil {
			return utils.NewErrIPs(err)
		}
		rply.AllocatedIP = &allocIP
	}
	if args.InitSession {
		var err error
		opts := engine.MapEvent(args.APIOpts)
		dbtItvl := sS.cgrCfg.SessionSCfg().DebitInterval
		if opts.HasField(utils.OptsDebitInterval) { // dynamic DebitInterval via CGRDebitInterval
			if dbtItvl, err = opts.GetDuration(utils.OptsDebitInterval); err != nil {
				return utils.NewErrRALs(err)
			}
		}
		s, err := sS.initSession(args.CGREvent, sS.biJClntID(ctx.Client), originID, dbtItvl,
			false, args.ForceDuration)
		if err != nil {
			return err
		}
		s.RLock() // avoid concurrency with activeDebit
		hasDebitLoops := s.debitStop != nil
		s.RUnlock()
		if hasDebitLoops { //active debit
			rply.MaxUsage = utils.DurationPointer(-1)
		} else {
			var sRunsUsage map[string]time.Duration
			if sRunsUsage, err = sS.updateSession(s, nil, args.APIOpts, false); err != nil {
				return utils.NewErrRALs(err)
			}
			if sS.cgrCfg.SessionSCfg().BackupInterval > 0 {
				sS.bkpSessionIDsMux.Lock()
				sS.bkpSessionIDs.Add(s.CGRID)
				sS.bkpSessionIDsMux.Unlock()
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
	AllocatedIP        *string
	ResourceAllocation *string
	MaxUsage           float64
	Thresholds         *string
	StatQueues         *string
}

// BiRPCv1InitiateSessionWithDigest returns the formated result of InitiateSession
func (sS *SessionS) BiRPCv1InitiateSessionWithDigest(ctx *context.Context,
	args *V1InitSessionArgs, initReply *V1InitReplyWithDigest) (err error) {
	var initSessionRply V1InitSessionReply
	if err = sS.BiRPCv1InitiateSession(ctx, args, &initSessionRply); err != nil {
		return
	}

	if args.GetAttributes &&
		initSessionRply.Attributes != nil {
		initReply.AttributesDigest = utils.StringPointer(initSessionRply.Attributes.Digest())
	}

	if args.AllocateResources {
		initReply.ResourceAllocation = initSessionRply.ResourceAllocation
	}

	if args.AllocateIP {
		initReply.AllocatedIP = utils.StringPointer(initSessionRply.AllocatedIP.Digest())
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
func NewV1UpdateSessionArgs(attrs, thresholds, stats bool, attributeIDs, thresholdIDs, statIDs []string,
	acnts bool, cgrEv *utils.CGREvent, forceDuration bool) (args *V1UpdateSessionArgs) {
	args = &V1UpdateSessionArgs{
		GetAttributes:     attrs,
		UpdateSession:     acnts,
		CGREvent:          cgrEv,
		ForceDuration:     forceDuration,
		ProcessThresholds: thresholds,
		ProcessStats:      stats,
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

// V1UpdateSessionArgs contains options for session update
type V1UpdateSessionArgs struct {
	GetAttributes     bool
	UpdateSession     bool
	ForceDuration     bool
	ProcessThresholds bool
	ProcessStats      bool
	AttributeIDs      []string
	ThresholdIDs      []string
	StatIDs           []string
	*utils.CGREvent
}

func (V1UpdateSessionArgs) RPCClone() {}

// V1UpdateSessionReply contains options for session update reply
type V1UpdateSessionReply struct {
	Attributes    *engine.AttrSProcessEventReply `json:",omitempty"`
	MaxUsage      *time.Duration                 `json:",omitempty"`
	ThresholdIDs  *[]string                      `json:",omitempty"`
	StatQueueIDs  *[]string                      `json:",omitempty"`
	needsMaxUsage bool                           // for gob encoding only
}

// SetMaxUsageNeeded used by agent that use the reply as NavigableMapper
// only used for gob encoding
func (v1Rply *V1UpdateSessionReply) SetMaxUsageNeeded(getMaxUsage bool) {
	if v1Rply == nil {
		return
	}
	v1Rply.needsMaxUsage = getMaxUsage
}

// AsNavigableMap is part of engine.NavigableMapper interface
func (v1Rply *V1UpdateSessionReply) AsNavigableMap() map[string]*utils.DataNode {
	cgrReply := make(map[string]*utils.DataNode)
	if v1Rply.Attributes != nil {
		attrs := &utils.DataNode{Type: utils.NMMapType, Map: make(map[string]*utils.DataNode)}
		for _, fldName := range v1Rply.Attributes.AlteredFields {
			fldName = strings.TrimPrefix(fldName, utils.MetaReq+utils.NestingSep)
			if v1Rply.Attributes.CGREvent.HasField(fldName) {
				attrs.Map[fldName] = utils.NewLeafNode(v1Rply.Attributes.CGREvent.Event[fldName])
			}
		}
		cgrReply[utils.CapAttributes] = attrs
	}
	if v1Rply.MaxUsage != nil {
		cgrReply[utils.CapMaxUsage] = utils.NewLeafNode(*v1Rply.MaxUsage)
	} else if v1Rply.needsMaxUsage {
		cgrReply[utils.CapMaxUsage] = utils.NewLeafNode(0)
	}
	return cgrReply
}

// BiRPCv1UpdateSession updates an existing session, returning the duration which the session can still last
func (sS *SessionS) BiRPCv1UpdateSession(ctx *context.Context,
	args *V1UpdateSessionArgs, rply *V1UpdateSessionReply) (err error) {
	if args.CGREvent == nil {
		return utils.NewErrMandatoryIeMissing(utils.CGREventString)
	}
	if args.CGREvent.Event == nil {
		return utils.NewErrMandatoryIeMissing(utils.Event)
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
		return utils.NewErrMandatoryIeMissing(utils.Subsystems)
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
		opts := engine.MapEvent(args.APIOpts)
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
			if s, err = sS.initSession(args.CGREvent, sS.biJClntID(ctx.Client), ev.GetStringIgnoreErrors(utils.OriginID),
				dbtItvl, false, args.ForceDuration); err != nil {
				return err
			}
		}
		var sRunsUsage map[string]time.Duration
		if sRunsUsage, err = sS.updateSession(s, ev, args.APIOpts, false); err != nil {
			return utils.NewErrRALs(err)
		}
		if sS.cgrCfg.SessionSCfg().BackupInterval > 0 {
			sS.bkpSessionIDsMux.Lock()
			sS.bkpSessionIDs.Add(s.CGRID)
			sS.bkpSessionIDsMux.Unlock()
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
	if args.ProcessThresholds {
		tIDs, err := sS.processThreshold(args.CGREvent, args.ThresholdIDs, true)
		if err == nil {
			rply.ThresholdIDs = &tIDs
		} else if err.Error() != utils.ErrNotFound.Error() {
			return utils.NewErrThresholdS(err)
		}
	}
	if args.ProcessStats {
		sIDs, err := sS.processStats(args.CGREvent, args.StatIDs, false)
		if err == nil {
			rply.StatQueueIDs = &sIDs
		} else if err.Error() != utils.ErrNotFound.Error() {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> error: %s processing event %+v with StatS.",
					utils.SessionS, err.Error(), args.CGREvent))
		}
	}
	return
}

// NewV1TerminateSessionArgs creates a new V1TerminateSessionArgs using the given arguments
func NewV1TerminateSessionArgs(acnts, resrc, ips,
	thrds bool, thresholdIDs []string, stats bool,
	statIDs []string, cgrEv *utils.CGREvent, forceDuration bool) (args *V1TerminateSessionArgs) {
	args = &V1TerminateSessionArgs{
		TerminateSession:  acnts,
		ReleaseResources:  resrc,
		ReleaseIP:         ips,
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
	ReleaseIP         bool
	ReleaseResources  bool
	ProcessThresholds bool
	ProcessStats      bool
	ThresholdIDs      []string
	StatIDs           []string
	*utils.CGREvent
}

func (V1TerminateSessionArgs) RPCClone() {}

// ParseFlags will populate the V1TerminateSessionArgs flags
func (args *V1TerminateSessionArgs) ParseFlags(flags, sep string) {
	for _, subsystem := range strings.Split(flags, sep) {
		switch {
		case subsystem == utils.MetaAccounts:
			args.TerminateSession = true
		case subsystem == utils.MetaResources:
			args.ReleaseResources = true
		case subsystem == utils.MetaIPs:
			args.ReleaseIP = true
		case strings.Contains(subsystem, utils.MetaThresholds):
			args.ProcessThresholds = true
			args.ThresholdIDs = getFlagIDs(subsystem)
		case strings.Contains(subsystem, utils.MetaStats):
			args.ProcessStats = true
			args.StatIDs = getFlagIDs(subsystem)
		case subsystem == utils.MetaFD:
			args.ForceDuration = true
		}
	}
}

// BiRPCv1TerminateSession will stop debit loops as well as release any used resources
func (sS *SessionS) BiRPCv1TerminateSession(ctx *context.Context,
	args *V1TerminateSessionArgs, rply *string) (err error) {
	if args.CGREvent == nil {
		return utils.NewErrMandatoryIeMissing(utils.CGREventString)
	}
	if args.CGREvent.Event == nil {
		return utils.NewErrMandatoryIeMissing(utils.Event)
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
	if !args.TerminateSession && !args.ReleaseResources && !args.ReleaseIP {
		return utils.NewErrMandatoryIeMissing(utils.Subsystems)
	}

	ev := engine.MapEvent(args.CGREvent.Event)
	opts := engine.MapEvent(args.APIOpts)
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
		fib := utils.FibDuration(time.Millisecond, 0)
		var isMsg bool // one time charging, do not perform indexing and sTerminator
		for i := 0; i < sS.cgrCfg.SessionSCfg().TerminateAttempts; i++ {
			if s = sS.getRelocateSession(cgrID,
				ev.GetStringIgnoreErrors(utils.InitialOriginID),
				ev.GetStringIgnoreErrors(utils.OriginID),
				ev.GetStringIgnoreErrors(utils.OriginHost)); s != nil {
				break
			}
			if i+1 < sS.cgrCfg.SessionSCfg().TerminateAttempts { // not last iteration
				time.Sleep(fib())
				continue
			}
			isMsg = true
			if s, err = sS.initSession(args.CGREvent, sS.biJClntID(ctx.Client), ev.GetStringIgnoreErrors(utils.OriginID),
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
		s.Chargeable = opts.GetBoolOrDefault(utils.OptsChargeable, true)
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
		if len(sS.cgrCfg.SessionSCfg().ResourceSConns) == 0 {
			return utils.NewErrNotConnected(utils.ResourceS)
		}
		if originID == "" {
			return utils.NewErrMandatoryIeMissing(utils.OriginID)
		}
		var reply string
		if args.APIOpts == nil {
			args.APIOpts = make(map[string]any)
		}
		args.CGREvent.APIOpts[utils.OptsResourcesUsageID] = originID // same ID should be accepted by first group since the previous resource should be expired
		args.CGREvent.APIOpts[utils.OptsResourcesUnits] = 1
		if err = sS.connMgr.Call(context.TODO(), sS.cgrCfg.SessionSCfg().ResourceSConns, utils.ResourceSv1ReleaseResources,
			args.CGREvent, &reply); err != nil {
			return utils.NewErrResourceS(err)
		}
	}
	if args.ReleaseIP {
		if len(sS.cgrCfg.SessionSCfg().IPsConns) == 0 {
			return utils.NewErrNotConnected(utils.IPs)
		}
		if originID == "" {
			return utils.NewErrMandatoryIeMissing(utils.OriginID)
		}
		var reply string
		if args.APIOpts == nil {
			args.APIOpts = make(map[string]any)
		}
		args.CGREvent.APIOpts[utils.OptsIPsAllocationID] = originID
		if err = sS.connMgr.Call(context.TODO(), sS.cgrCfg.SessionSCfg().IPsConns, utils.IPsV1ReleaseIP,
			args.CGREvent, &reply); err != nil {
			return utils.NewErrIPs(err)
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
func (sS *SessionS) BiRPCv1ProcessCDR(ctx *context.Context,
	cgrEv *utils.CGREvent, rply *string) (err error) {
	if cgrEv.Event == nil {
		return utils.NewErrMandatoryIeMissing(utils.Event)
	}
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
	thds bool, thresholdIDs []string, stats bool, statIDs []string, resrc, ips, acnts,
	routes, routesIgnoreErrs, routesEventCost bool, cgrEv *utils.CGREvent,
	routePaginator utils.Paginator, forceDuration bool, routesMaxCost string) (args *V1ProcessMessageArgs) {
	args = &V1ProcessMessageArgs{
		AllocateResources:  resrc,
		AllocateIP:         ips,
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
	AllocateIP         bool
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

func (V1ProcessMessageArgs) RPCClone() {}

// ParseFlags will populate the V1ProcessMessageArgs flags
func (args *V1ProcessMessageArgs) ParseFlags(flags, sep string) {
	for _, subsystem := range strings.Split(flags, sep) {
		switch {
		case subsystem == utils.MetaAccounts:
			args.Debit = true
		case subsystem == utils.MetaResources:
			args.AllocateResources = true
		case subsystem == utils.MetaIPs:
			args.AllocateIP = true
		case subsystem == utils.MetaRoutes:
			args.GetRoutes = true
		case subsystem == utils.MetaRoutesIgnoreErrors:
			args.RoutesIgnoreErrors = true
		case subsystem == utils.MetaRoutesEventCost:
			args.RoutesMaxCost = utils.MetaEventCost
		case strings.HasPrefix(subsystem, utils.MetaRoutesMaxCost):
			args.RoutesMaxCost = strings.TrimPrefix(subsystem, utils.MetaRoutesMaxCost+utils.InInFieldSep)
		case strings.Contains(subsystem, utils.MetaAttributes):
			args.GetAttributes = true
			args.AttributeIDs = getFlagIDs(subsystem)
		case strings.Contains(subsystem, utils.MetaThresholds):
			args.ProcessThresholds = true
			args.ThresholdIDs = getFlagIDs(subsystem)
		case strings.Contains(subsystem, utils.MetaStats):
			args.ProcessStats = true
			args.StatIDs = getFlagIDs(subsystem)
		case subsystem == utils.MetaFD:
			args.ForceDuration = true
		}
	}
	args.Paginator, _ = utils.GetRoutePaginatorFromOpts(args.APIOpts)
}

// V1ProcessMessageReply is the reply for the ProcessMessage API
type V1ProcessMessageReply struct {
	MaxUsage           *time.Duration                 `json:",omitempty"`
	AllocatedIP        *engine.AllocatedIP            `json:",omitempty"`
	ResourceAllocation *string                        `json:",omitempty"`
	Attributes         *engine.AttrSProcessEventReply `json:",omitempty"`
	RouteProfiles      engine.SortedRoutesList        `json:",omitempty"`
	ThresholdIDs       *[]string                      `json:",omitempty"`
	StatQueueIDs       *[]string                      `json:",omitempty"`

	needsMaxUsage bool // for gob encoding only
}

// SetMaxUsageNeeded used by agent that use the reply as NavigableMapper
// only used for gob encoding
func (r *V1ProcessMessageReply) SetMaxUsageNeeded(getMaxUsage bool) {
	if r == nil {
		return
	}
	r.needsMaxUsage = getMaxUsage
}

// AsNavigableMap is part of engine.NavigableMapper interface
func (r *V1ProcessMessageReply) AsNavigableMap() map[string]*utils.DataNode {
	cgrReply := make(map[string]*utils.DataNode)
	if r.MaxUsage != nil {
		cgrReply[utils.CapMaxUsage] = utils.NewLeafNode(*r.MaxUsage)
	} else if r.needsMaxUsage {
		cgrReply[utils.CapMaxUsage] = utils.NewLeafNode(0)
	}
	if r.ResourceAllocation != nil {
		cgrReply[utils.CapResourceAllocation] = utils.NewLeafNode(*r.ResourceAllocation)
	}
	if r.AllocatedIP != nil {
		cgrReply[utils.CapAllocatedIP] = &utils.DataNode{
			Type: utils.NMMapType,
			Map:  r.AllocatedIP.AsNavigableMap(),
		}
	}
	if r.Attributes != nil {
		attrs := &utils.DataNode{Type: utils.NMMapType, Map: make(map[string]*utils.DataNode)}
		for _, fldName := range r.Attributes.AlteredFields {
			fldName = strings.TrimPrefix(fldName, utils.MetaReq+utils.NestingSep)
			if r.Attributes.CGREvent.HasField(fldName) {
				attrs.Map[fldName] = utils.NewLeafNode(r.Attributes.CGREvent.Event[fldName])
			}
		}
		cgrReply[utils.CapAttributes] = attrs
	}
	if r.RouteProfiles != nil {
		cgrReply[utils.CapRouteProfiles] = r.RouteProfiles.AsNavigableMap()
	}
	if r.ThresholdIDs != nil {
		thIDs := &utils.DataNode{Type: utils.NMSliceType, Slice: make([]*utils.DataNode, len(*r.ThresholdIDs))}
		for i, v := range *r.ThresholdIDs {
			thIDs.Slice[i] = utils.NewLeafNode(v)
		}
		cgrReply[utils.CapThresholds] = thIDs
	}
	if r.StatQueueIDs != nil {
		stIDs := &utils.DataNode{Type: utils.NMSliceType, Slice: make([]*utils.DataNode, len(*r.StatQueueIDs))}
		for i, v := range *r.StatQueueIDs {
			stIDs.Slice[i] = utils.NewLeafNode(v)
		}
		cgrReply[utils.CapStatQueues] = stIDs
	}
	return cgrReply
}

// BiRPCv1ProcessMessage processes one event with the right subsystems based on arguments received
func (sS *SessionS) BiRPCv1ProcessMessage(ctx *context.Context,
	args *V1ProcessMessageArgs, rply *V1ProcessMessageReply) (err error) {
	if args.CGREvent == nil {
		return utils.NewErrMandatoryIeMissing(utils.CGREventString)
	}
	if args.CGREvent.Event == nil {
		return utils.NewErrMandatoryIeMissing(utils.Event)
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
		if len(sS.cgrCfg.SessionSCfg().ResourceSConns) == 0 {
			return utils.NewErrNotConnected(utils.ResourceS)
		}
		if originID == "" {
			return utils.NewErrMandatoryIeMissing(utils.OriginID)
		}
		if args.APIOpts == nil {
			args.APIOpts = make(map[string]any)
		}
		args.CGREvent.APIOpts[utils.OptsResourcesUsageID] = originID
		args.CGREvent.APIOpts[utils.OptsResourcesUnits] = 1
		var allocMessage string
		if err = sS.connMgr.Call(context.TODO(), sS.cgrCfg.SessionSCfg().ResourceSConns, utils.ResourceSv1AllocateResources,
			args.CGREvent, &allocMessage); err != nil {
			return utils.NewErrResourceS(err)
		}
		rply.ResourceAllocation = &allocMessage
	}
	if args.AllocateIP {
		if len(sS.cgrCfg.SessionSCfg().IPsConns) == 0 {
			return utils.NewErrNotConnected(utils.IPs)
		}
		if originID == "" {
			return utils.NewErrMandatoryIeMissing(utils.OriginID)
		}
		if args.APIOpts == nil {
			args.APIOpts = make(map[string]any)
		}
		args.CGREvent.APIOpts[utils.OptsIPsAllocationID] = originID
		var allocIP engine.AllocatedIP
		if err = sS.connMgr.Call(context.TODO(), sS.cgrCfg.SessionSCfg().IPsConns, utils.IPsV1AllocateIP,
			args.CGREvent, &allocIP); err != nil {
			return utils.NewErrIPs(err)
		}
		rply.AllocatedIP = &allocIP
	}
	if args.GetRoutes {
		routesReply, err := sS.getRoutes(args.CGREvent.Clone(), args.Paginator,
			args.RoutesIgnoreErrors, args.RoutesMaxCost, false)
		if err != nil {
			return err
		}
		if routesReply != nil {
			rply.RouteProfiles = routesReply
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

// RPCClone implements rpcclient.RPCCloner interface
func (V1ProcessEventArgs) RPCClone() {}

// V1ProcessEventReply is the reply for the ProcessEvent API
type V1ProcessEventReply struct {
	MaxUsage           map[string]time.Duration                  `json:",omitempty"`
	Cost               map[string]float64                        `json:",omitempty"` // Cost is the cost received from Rater, ignoring accounting part
	ResourceAllocation map[string]string                         `json:",omitempty"`
	AllocatedIP        map[string]*engine.AllocatedIP            `json:",omitempty"`
	Attributes         map[string]*engine.AttrSProcessEventReply `json:",omitempty"`
	RouteProfiles      map[string]engine.SortedRoutesList        `json:",omitempty"`
	ThresholdIDs       map[string][]string                       `json:",omitempty"`
	StatQueueIDs       map[string][]string                       `json:",omitempty"`
	STIRIdentity       map[string]string                         `json:",omitempty"`
}

// AsNavigableMap is part of engine.NavigableMapper interface
func (r *V1ProcessEventReply) AsNavigableMap() map[string]*utils.DataNode {
	cgrReply := make(map[string]*utils.DataNode)
	if r.MaxUsage != nil {
		usage := &utils.DataNode{Type: utils.NMMapType, Map: make(map[string]*utils.DataNode)}
		for k, v := range r.MaxUsage {
			usage.Map[k] = utils.NewLeafNode(v)
		}
		cgrReply[utils.CapMaxUsage] = usage
	}
	if r.ResourceAllocation != nil {
		res := &utils.DataNode{
			Type: utils.NMMapType,
			Map:  make(map[string]*utils.DataNode),
		}
		for k, v := range r.ResourceAllocation {
			res.Map[k] = utils.NewLeafNode(v)
		}
		cgrReply[utils.CapResourceAllocation] = res
	}
	if r.AllocatedIP != nil {
		ips := &utils.DataNode{
			Type: utils.NMMapType,
			Map:  make(map[string]*utils.DataNode),
		}
		for k, v := range r.AllocatedIP {
			ips.Map[k] = &utils.DataNode{
				Type: utils.NMMapType,
				Map:  v.AsNavigableMap(),
			}
		}
		cgrReply[utils.CapAllocatedIP] = ips
	}
	if r.Attributes != nil {
		atts := &utils.DataNode{Type: utils.NMMapType, Map: make(map[string]*utils.DataNode)}
		for k, att := range r.Attributes {
			attrs := &utils.DataNode{Type: utils.NMMapType, Map: make(map[string]*utils.DataNode)}
			for _, fldName := range att.AlteredFields {
				fldName = strings.TrimPrefix(fldName, utils.MetaReq+utils.NestingSep)
				if att.CGREvent.HasField(fldName) {
					attrs.Map[fldName] = utils.NewLeafNode(att.CGREvent.Event[fldName])
				}
			}
			atts.Map[k] = attrs
		}
		cgrReply[utils.CapAttributes] = atts
	}
	if r.RouteProfiles != nil {
		routes := &utils.DataNode{Type: utils.NMMapType, Map: make(map[string]*utils.DataNode)}
		for k, route := range r.RouteProfiles {
			routes.Map[k] = route.AsNavigableMap()
		}
		cgrReply[utils.CapRouteProfiles] = routes
	}
	if r.ThresholdIDs != nil {
		th := &utils.DataNode{Type: utils.NMMapType, Map: make(map[string]*utils.DataNode)}
		for k, thr := range r.ThresholdIDs {
			thIDs := &utils.DataNode{Type: utils.NMSliceType, Slice: make([]*utils.DataNode, len(thr))}
			for i, v := range thr {
				thIDs.Slice[i] = utils.NewLeafNode(v)
			}
			th.Map[k] = thIDs
		}
		cgrReply[utils.CapThresholds] = th
	}
	if r.StatQueueIDs != nil {
		st := &utils.DataNode{Type: utils.NMMapType, Map: make(map[string]*utils.DataNode)}
		for k, sts := range r.StatQueueIDs {
			stIDs := &utils.DataNode{Type: utils.NMSliceType, Slice: make([]*utils.DataNode, len(sts))}
			for i, v := range sts {
				stIDs.Slice[i] = utils.NewLeafNode(v)
			}
			st.Map[k] = stIDs
		}
		cgrReply[utils.CapStatQueues] = st
	}
	if r.Cost != nil {
		costs := &utils.DataNode{Type: utils.NMMapType, Map: make(map[string]*utils.DataNode)}
		for k, cost := range r.Cost {
			costs.Map[k] = utils.NewLeafNode(cost)
		}
	}
	if r.STIRIdentity != nil {
		stir := &utils.DataNode{Type: utils.NMMapType, Map: make(map[string]*utils.DataNode)}
		for k, v := range r.STIRIdentity {
			stir.Map[k] = utils.NewLeafNode(v)
		}
		cgrReply[utils.OptsStirIdentity] = stir
	}
	return cgrReply
}

// BiRPCv1ProcessEvent processes one event with the right subsystems based on arguments received
func (sS *SessionS) BiRPCv1ProcessEvent(ctx *context.Context,
	args *V1ProcessEventArgs, rply *V1ProcessEventReply) (err error) {
	if args.CGREvent == nil {
		return utils.NewErrMandatoryIeMissing(utils.CGREventString)
	}
	if args.CGREvent.Event == nil {
		return utils.NewErrMandatoryIeMissing(utils.Event)
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
		rply.RouteProfiles = make(map[string]engine.SortedRoutesList)
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
			if routesReply != nil {
				rply.RouteProfiles[runID] = routesReply
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
			opts := engine.MapEvent(cgrEv.APIOpts)
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
			opts := engine.MapEvent(cgrEv.APIOpts)
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
		if len(sS.cgrCfg.SessionSCfg().ResourceSConns) == 0 {
			return utils.NewErrNotConnected(utils.ResourceS)
		}
		rply.ResourceAllocation = make(map[string]string)
		if resOpt := argsFlagsWithParams[utils.MetaResources]; len(resOpt) != 0 {
			for runID, cgrEv := range getDerivedEvents(events, resOpt.Has(utils.MetaDerivedReply)) {
				originID := engine.MapEvent(cgrEv.Event).GetStringIgnoreErrors(utils.OriginID)
				if originID == "" {
					return utils.NewErrMandatoryIeMissing(utils.OriginID)
				}
				if args.APIOpts == nil {
					args.APIOpts = make(map[string]any)
				}
				args.CGREvent.APIOpts[utils.OptsResourcesUsageID] = originID
				args.CGREvent.APIOpts[utils.OptsResourcesUnits] = 1
				cgrEv.SetCloneable(true)
				var resMessage string
				// check what we need to do for resources (*authorization/*allocation)
				//check for subflags and convert them into utils.FlagsWithParams
				switch {
				case resOpt.Has(utils.MetaAuthorize):
					if err = sS.connMgr.Call(context.TODO(), sS.cgrCfg.SessionSCfg().ResourceSConns, utils.ResourceSv1AuthorizeResources,
						args.CGREvent, &resMessage); err != nil {
						if blockError {
							return utils.NewErrResourceS(err)
						}
						utils.Logger.Warning(
							fmt.Sprintf("<%s> error: <%s> processing event %+v for RunID <%s>  with ResourceS.",
								utils.SessionS, err.Error(), cgrEv, runID))
						withErrors = true
					}
				case resOpt.Has(utils.MetaAllocate):
					if err = sS.connMgr.Call(context.TODO(), sS.cgrCfg.SessionSCfg().ResourceSConns, utils.ResourceSv1AllocateResources,
						args.CGREvent, &resMessage); err != nil {
						if blockError {
							return utils.NewErrResourceS(err)
						}
						utils.Logger.Warning(
							fmt.Sprintf("<%s> error: <%s> processing event %+v for RunID <%s>  with ResourceS.",
								utils.SessionS, err.Error(), cgrEv, runID))
						withErrors = true
					}
				case resOpt.Has(utils.MetaRelease):
					if err = sS.connMgr.Call(context.TODO(), sS.cgrCfg.SessionSCfg().ResourceSConns, utils.ResourceSv1ReleaseResources,
						args.CGREvent, &resMessage); err != nil {
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

	// check for *ips
	if argsFlagsWithParams.GetBool(utils.MetaIPs) {
		if len(sS.cgrCfg.SessionSCfg().IPsConns) == 0 {
			return utils.NewErrNotConnected(utils.IPs)
		}
		rply.AllocatedIP = make(map[string]*engine.AllocatedIP)
		if ipsOpt := argsFlagsWithParams[utils.MetaIPs]; len(ipsOpt) != 0 {
			for runID, cgrEv := range getDerivedEvents(events, ipsOpt.Has(utils.MetaDerivedReply)) {
				originID := engine.MapEvent(cgrEv.Event).GetStringIgnoreErrors(utils.OriginID)
				if originID == "" {
					return utils.NewErrMandatoryIeMissing(utils.OriginID)
				}
				if args.APIOpts == nil {
					args.APIOpts = make(map[string]any)
				}
				args.CGREvent.APIOpts[utils.OptsIPsAllocationID] = originID
				cgrEv.SetCloneable(true)
				var allocIP engine.AllocatedIP
				switch {
				case ipsOpt.Has(utils.MetaAuthorize):
					if err = sS.connMgr.Call(context.TODO(), sS.cgrCfg.SessionSCfg().IPsConns, utils.IPsV1AuthorizeIP,
						args.CGREvent, &allocIP); err != nil {
						if blockError {
							return utils.NewErrIPs(err)
						}
						utils.Logger.Warning(
							fmt.Sprintf("<%s> error: <%s> processing event %+v for RunID <%s> with IPs.",
								utils.SessionS, err.Error(), cgrEv, runID))
						withErrors = true
					}
				case ipsOpt.Has(utils.MetaAllocate):
					if err = sS.connMgr.Call(context.TODO(), sS.cgrCfg.SessionSCfg().IPsConns, utils.IPsV1AllocateIP,
						args.CGREvent, &allocIP); err != nil {
						if blockError {
							return utils.NewErrIPs(err)
						}
						utils.Logger.Warning(
							fmt.Sprintf("<%s> error: <%s> processing event %+v for RunID <%s> with IPs.",
								utils.SessionS, err.Error(), cgrEv, runID))
						withErrors = true
					}
				case ipsOpt.Has(utils.MetaRelease):
					var reply string
					if err = sS.connMgr.Call(context.TODO(), sS.cgrCfg.SessionSCfg().IPsConns, utils.IPsV1ReleaseIP,
						args.CGREvent, &reply); err != nil {
						if blockError {
							return utils.NewErrIPs(err)
						}
						utils.Logger.Warning(
							fmt.Sprintf("<%s> error: <%s> processing event %+v for RunID <%s> with IPs.",
								utils.SessionS, err.Error(), cgrEv, runID))
						withErrors = true
					}
				}
				rply.AllocatedIP[runID] = &allocIP
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
					if err = sS.connMgr.Call(context.TODO(), sS.cgrCfg.SessionSCfg().RALsConns,
						utils.ResponderGetCost,
						&engine.CallDescriptorWithAPIOpts{
							CallDescriptor: cd,
							APIOpts:        cgrEv.APIOpts,
						}, &cc); err != nil {
						return err
					}
					rply.Cost[runID] = cc.Cost
				}
			}
			opts := engine.MapEvent(args.APIOpts)
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
				s, err := sS.initSession(args.CGREvent, sS.biJClntID(ctx.Client), originID, dbtItvl, false,
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
						sRunsMaxUsage[sr.CD.RunID] = sS.cgrCfg.SessionSCfg().GetDefaultUsage(ev.GetStringIgnoreErrors(utils.ToR))
					}
				} else if sRunsMaxUsage, err = sS.updateSession(s, nil, args.APIOpts, false); err != nil {
					return utils.NewErrRALs(err)
				}
				if sS.cgrCfg.SessionSCfg().BackupInterval > 0 {
					sS.bkpSessionIDsMux.Lock()
					sS.bkpSessionIDs.Add(s.CGRID)
					sS.bkpSessionIDsMux.Unlock()
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
					if s, err = sS.initSession(args.CGREvent, sS.biJClntID(ctx.Client), ev.GetStringIgnoreErrors(utils.OriginID),
						dbtItvl, false, ralsOpts.Has(utils.MetaFD)); err != nil {
						return err
					}
				}
				var sRunsMaxUsage map[string]time.Duration
				if sRunsMaxUsage, err = sS.updateSession(s, ev, args.APIOpts, false); err != nil {
					return utils.NewErrRALs(err)
				}
				if sS.cgrCfg.SessionSCfg().BackupInterval > 0 {
					sS.bkpSessionIDsMux.Lock()
					sS.bkpSessionIDs.Add(s.CGRID)
					sS.bkpSessionIDsMux.Unlock()
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
					if s, err = sS.initSession(args.CGREvent, sS.biJClntID(ctx.Client), ev.GetStringIgnoreErrors(utils.OriginID),
						dbtItvl, false, ralsOpts.Has(utils.MetaFD)); err != nil {
						return err
					}
					if _, err = sS.updateSession(s, ev, opts, false); err != nil {
						return err
					}
				} else {
					s.Lock()
					s.Chargeable = opts.GetBoolOrDefault(utils.OptsChargeable, true)
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
func (sS *SessionS) BiRPCv1GetCost(ctx *context.Context,
	args *V1ProcessEventArgs, rply *V1GetCostReply) (err error) {
	if args.CGREvent == nil {
		return utils.NewErrMandatoryIeMissing(utils.CGREventString)
	}
	if args.CGREvent.Event == nil {
		return utils.NewErrMandatoryIeMissing(utils.Event)
	}
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
	if err = sS.connMgr.Call(context.TODO(), sS.cgrCfg.SessionSCfg().RALsConns,
		utils.ResponderGetCost,
		&engine.CallDescriptorWithAPIOpts{
			CallDescriptor: cd,
			APIOpts:        args.APIOpts,
		}, &cc); err != nil {
		return
	}
	ec := engine.NewEventCostFromCallCost(&cc, args.CGREvent.ID, me.GetStringIgnoreErrors(utils.RunID))
	ec.Compute()
	rply.EventCost = ec
	return
}

// BiRPCv1SyncSessions will sync sessions on demand
func (sS *SessionS) BiRPCv1SyncSessions(ctx *context.Context,
	ignParam *utils.TenantWithAPIOpts, reply *string) error {
	sS.syncSessions()
	*reply = utils.OK
	return nil
}

// BiRPCv1ForceDisconnect will force disconnecting sessions matching sessions
func (sS *SessionS) BiRPCv1ForceDisconnect(ctx *context.Context,
	args utils.SessionFilterWithEvent, reply *string) (err error) {
	if args.SessionFilter == nil { //protection in case on nil
		args.SessionFilter = &utils.SessionFilter{}
	}
	if len(args.Filters) != 0 && sS.dm == nil {
		return utils.ErrNoDatabaseConn
	}
	aSs := sS.filterSessions(args.SessionFilter, false)
	if len(aSs) == 0 {
		return utils.ErrNotFound
	}
	for _, as := range aSs {
		ss := sS.getSessions(as.CGRID, false)
		if len(ss) == 0 {
			continue
		}
		ss[0].Lock()
		if errTerm := sS.forceSTerminate(ss[0], 0, nil, nil,
			args.APIOpts, args.Event); errTerm != nil {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> failed force-terminating session with id: <%s>, err: <%v>",
					utils.SessionS, ss[0].cgrID(), errTerm))
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
func (sS *SessionS) BiRPCv1RegisterInternalBiJSONConn(ctx *context.Context,
	connID string, reply *string) error {
	sS.RegisterIntBiJConn(ctx.Client, connID)
	*reply = utils.OK
	return nil
}

// BiRPCv1ActivateSessions is called to activate a list/all sessions
// returns utils.ErrPartiallyExecuted in case of errors
func (sS *SessionS) BiRPCv1ActivateSessions(ctx *context.Context,
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
		} else {
			if sS.cgrCfg.SessionSCfg().BackupInterval > 0 {
				sS.bkpSessionIDsMux.Lock()
				sS.bkpSessionIDs.Add(sID)
				sS.bkpSessionIDsMux.Unlock()
			}
		}
	}
	if err == nil {
		*reply = utils.OK
	}
	return
}

// BiRPCv1DeactivateSessions is called to deactivate a list/all active sessios
// returns utils.ErrPartiallyExecuted in case of errors
func (sS *SessionS) BiRPCv1DeactivateSessions(ctx *context.Context,
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
		} else {
			if sS.cgrCfg.SessionSCfg().BackupInterval > 0 {
				sS.removeSsCGRIDsMux.Lock()
				sS.removeSsCGRIDs.Add(sID)
				sS.removeSsCGRIDsMux.Unlock()
			}
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
		return sS.connMgr.Call(context.TODO(), sS.cgrCfg.SessionSCfg().CDRsConns, utils.CDRsV1ProcessEvent,
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
		if err = sS.connMgr.Call(context.TODO(), sS.cgrCfg.SessionSCfg().CDRsConns, utils.CDRsV1ProcessEvent,
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
	if len(sS.cgrCfg.SessionSCfg().ThresholdSConns) == 0 {
		return tIDs, utils.NewErrNotConnected(utils.ThresholdS)
	}
	if cgrEv.APIOpts == nil {
		cgrEv.APIOpts = make(map[string]any)
	}
	// check if we have thresholdIDs
	if len(thIDs) != 0 {
		cgrEv.APIOpts[utils.OptsThresholdsProfileIDs] = thIDs
	}
	cgrEv.SetCloneable(clnb)
	//initialize the returned variable
	err = sS.connMgr.Call(context.TODO(), sS.cgrCfg.SessionSCfg().ThresholdSConns, utils.ThresholdSv1ProcessEvent, cgrEv, &tIDs)
	return
}

// processStats will receive the event and send it to StatS to be processed
func (sS *SessionS) processStats(cgrEv *utils.CGREvent, stsIDs []string, clnb bool) (sIDs []string, err error) {
	if len(sS.cgrCfg.SessionSCfg().StatSConns) == 0 {
		return sIDs, utils.NewErrNotConnected(utils.StatS)
	}
	if cgrEv.APIOpts == nil {
		cgrEv.APIOpts = make(map[string]any)
	}
	// check in case we have StatIDs inside flags
	if len(stsIDs) != 0 {
		cgrEv.APIOpts[utils.OptsStatsProfileIDs] = stsIDs
	}
	cgrEv.SetCloneable(clnb)
	//initialize the returned variable
	err = sS.connMgr.Call(context.TODO(), sS.cgrCfg.SessionSCfg().StatSConns, utils.StatSv1ProcessEvent, cgrEv, &sIDs)
	return
}

// getRoutes will receive the event and send it to SupplierS to find the suppliers
func (sS *SessionS) getRoutes(cgrEv *utils.CGREvent, pag utils.Paginator, ignoreErrors bool,
	maxCost string, clnb bool) (routesReply engine.SortedRoutesList, err error) {
	if len(sS.cgrCfg.SessionSCfg().RouteSConns) == 0 {
		return routesReply, utils.NewErrNotConnected(utils.RouteS)
	}
	if acd, has := cgrEv.Event[utils.ACD]; has {
		cgrEv.Event[utils.Usage] = acd
	}
	if maxCost != utils.EmptyString {
		cgrEv.APIOpts[utils.OptsRoutesMaxCost] = maxCost
	}
	if ignoreErrors {
		cgrEv.APIOpts[utils.OptsRoutesIgnoreErrors] = ignoreErrors
	}
	cgrEv.SetCloneable(clnb)
	if err = sS.connMgr.Call(context.TODO(), sS.cgrCfg.SessionSCfg().RouteSConns, utils.RouteSv1GetRoutes,
		cgrEv, &routesReply); err != nil {
		return routesReply, utils.NewErrRouteS(err)
	}
	return
}

// processAttributes will receive the event and send it to AttributeS to be processed
func (sS *SessionS) processAttributes(cgrEv *utils.CGREvent, attrIDs []string,
	clnb bool) (rplyEv engine.AttrSProcessEventReply, err error) {
	if len(sS.cgrCfg.SessionSCfg().AttributeSConns) == 0 {
		return rplyEv, utils.NewErrNotConnected(utils.AttributeS)
	}
	if cgrEv.APIOpts == nil {
		cgrEv.APIOpts = make(engine.MapEvent)
	}
	cgrEv.APIOpts[utils.MetaSubsys] = utils.MetaSessionS
	cgrEv.APIOpts[utils.OptsAttributesProfileIDs] = attrIDs
	ctx, has := cgrEv.APIOpts[utils.OptsContext]
	cgrEv.APIOpts[utils.OptsContext] = utils.FirstNonEmpty(
		utils.IfaceAsString(ctx),
		utils.MetaSessionS)
	cgrEv.SetCloneable(clnb)
	err = sS.connMgr.Call(context.TODO(), sS.cgrCfg.SessionSCfg().AttributeSConns, utils.AttributeSv1ProcessEvent,
		cgrEv, &rplyEv)
	if err == nil && !has && utils.IfaceAsString(rplyEv.APIOpts[utils.OptsContext]) == utils.MetaSessionS {
		delete(rplyEv.APIOpts, utils.OptsContext)
	}
	return
}

// BiRPCV1GetMaxUsage returns the maximum usage as seconds, compatible with OpenSIPS 2.3
// DEPRECATED, it will be removed in future versions
func (sS *SessionS) BiRPCV1GetMaxUsage(ctx *context.Context,
	ev engine.MapEvent, maxUsage *float64) (err error) {
	var rply *V1AuthorizeReply
	if err = sS.BiRPCv1AuthorizeEvent(
		ctx,
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
func (sS *SessionS) BiRPCV1InitiateSession(ctx *context.Context,
	ev engine.MapEvent, maxUsage *float64) (err error) {
	var rply *V1InitSessionReply
	if err = sS.BiRPCv1InitiateSession(
		ctx,
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
func (sS *SessionS) BiRPCV1UpdateSession(ctx *context.Context,
	ev engine.MapEvent, maxUsage *float64) (err error) {
	var rply *V1UpdateSessionReply
	if err = sS.BiRPCv1UpdateSession(
		ctx,
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
func (sS *SessionS) BiRPCV1TerminateSession(ctx *context.Context,
	ev engine.MapEvent, rply *string) (err error) {
	return sS.BiRPCv1TerminateSession(
		ctx,
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
func (sS *SessionS) BiRPCV1ProcessCDR(ctx *context.Context,
	ev engine.MapEvent, rply *string) (err error) {
	return sS.BiRPCv1ProcessCDR(
		ctx,
		&utils.CGREvent{
			Tenant: utils.FirstNonEmpty(
				ev.GetStringIgnoreErrors(utils.Tenant),
				sS.cgrCfg.GeneralCfg().DefaultTenant),
			ID:    utils.UUIDSha1Prefix(),
			Event: ev},
		rply)
}

func (sS *SessionS) alterSession(ctx *context.Context, s *Session, apiOpts map[string]any, event map[string]any) (err error) {
	clnt := sS.biJClnt(s.ClientConnID)
	if clnt == nil {
		return fmt.Errorf("calling %s requires bidirectional JSON connection, connID: <%s>",
			utils.AgentV1AlterSession, s.ClientConnID)
	}

	// Merge parameter event with the session event. Losing the EventStart OriginID
	// could create unwanted behaviour.
	if event == nil {
		event = make(map[string]any)
	}
	for key, val := range s.EventStart {
		if _, has := event[key]; !has {
			event[key] = val
		}
	}
	args := utils.CGREvent{
		ID:      utils.GenUUID(),
		Time:    utils.TimePointer(time.Now()),
		APIOpts: apiOpts,
		Event:   event,
	}

	var rply string
	if err = clnt.conn.Call(ctx, utils.AgentV1AlterSession, args, &rply); err == utils.ErrNotImplemented {
		err = nil
	}
	return
}

// BiRPCv1AlterSessions sends a RAR for the matching sessions
func (sS *SessionS) BiRPCv1AlterSessions(ctx *context.Context,
	args utils.SessionFilterWithEvent, reply *string) (err error) {
	if args.SessionFilter == nil { //protection in case on nil
		args.SessionFilter = &utils.SessionFilter{}
	}
	aSs := sS.filterSessions(args.SessionFilter, false)
	if len(aSs) == 0 {
		return utils.ErrNotFound
	}
	uniqueSIDs := utils.NewStringSet(nil)
	for _, as := range aSs {
		if uniqueSIDs.Has(as.CGRID) {
			continue
		}
		uniqueSIDs.Add(as.CGRID)
		ss := sS.getSessions(as.CGRID, false)
		if len(ss) == 0 {
			continue
		}
		if errTerm := sS.alterSession(ctx, ss[0], args.APIOpts, args.Event); errTerm != nil {
			utils.Logger.Warning(
				fmt.Sprintf(
					"<%s> altering session with id '%s' failed: <%v>",
					utils.SessionS, ss[0].cgrID(), errTerm))
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
func (sS *SessionS) BiRPCv1DisconnectPeer(ctx *context.Context,
	args *utils.DPRArgs, reply *string) (err error) {
	hasErrors := false
	clients := make(map[string]*biJClient)
	sS.biJMux.RLock()
	for ID, clnt := range sS.biJIDs {
		clients[ID] = clnt
	}
	sS.biJMux.RUnlock()
	for ID, clnt := range clients {
		if err = clnt.conn.Call(ctx, utils.AgentV1DisconnectPeer, args, reply); err != nil && err != utils.ErrNotImplemented {
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
func (sS *SessionS) BiRPCv1STIRAuthenticate(ctx *context.Context,
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
func (sS *SessionS) BiRPCv1STIRIdentity(ctx *context.Context,
	args *V1STIRIdentityArgs, identity *string) (err error) {
	if args == nil || args.Payload == nil {
		return utils.NewErrMandatoryIeMissing("Payload")
	}
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

// BiRPCv1STIRIdentity the API for STIR header creation
func (sS *SessionS) BiRPCv1CapsError(ctx *context.Context,
	args any, identity *string) (err error) {
	return utils.ErrMaxConcurrentRPCExceededNoCaps
}

// BiRPCv1Sleep mimics a request whose process takes the given amount of time to process
func (ssv1 *SessionS) BiRPCv1Sleep(ctx *context.Context, args *utils.DurationArgs,
	reply *string) (err error) {
	time.Sleep(args.Duration)
	*reply = utils.OK
	return nil
}

// RestoreAndBackupSessions will restore previuos backup sessions and start backup looping
func (sS *SessionS) RestoreAndBackupSessions(stopChan chan struct{}) error {
	var restoredSess []*Session //holds the restored sessions gotten from datadb
	storedSessions, err := sS.dm.GetSessionsBackup(sS.cgrCfg.GeneralCfg().NodeID,
		sS.cgrCfg.GeneralCfg().DefaultTenant)
	if err != nil && err != utils.ErrNoBackupFound { // if backup is not found we still want to start the backup loop
		return err
	} else {
		for _, storSess := range storedSessions {
			storSess := newSessionFromStoredSession(storSess)
			restoredSess = append(restoredSess, storSess)
		}
		sS.restoreSessions(restoredSess)
	}

	go sS.runBackup(stopChan)
	return nil
}

// Start running backup loop
func (sS *SessionS) runBackup(stopChan chan struct{}) {
	backupInterval := sS.cgrCfg.SessionSCfg().BackupInterval
	if backupInterval > 0 {
		for {
			if err := sS.storeSessionsMarked(); err != nil {
				utils.Logger.Err(fmt.Sprintf("Backup Sessions error: <%v>", err))
			}
			select {
			case <-stopChan:
				return
			case <-time.After(backupInterval):
			}
		}
	}
}

// storeSessionsMarked stores only marked active sessions for backup in DataDB, and removes inactive sessions from it
func (sS *SessionS) storeSessionsMarked() (err error) {
	sS.bkpSessionIDsMux.Lock()
	var storedSessions []*engine.StoredSession // hold the converted active marked sessions
	for cgrID := range sS.bkpSessionIDs {
		activeSess := sS.getSessions(cgrID, false)
		if len(activeSess) == 0 {
			utils.Logger.Warning("<SessionS> Couldn't backup session with CGRID <" + cgrID + ">. Session is not active")
			delete(sS.bkpSessionIDs, cgrID) // remove inactive session cgrids from the map
			continue
		}
		activeSess[0].lk.RLock()
		storedSessions = append(storedSessions, activeSess[0].asStoredSession())
		activeSess[0].lk.RUnlock()
	}
	if len(storedSessions) != 0 {
		if err := sS.dm.SetBackupSessions(sS.cgrCfg.GeneralCfg().NodeID,
			sS.cgrCfg.GeneralCfg().DefaultTenant, storedSessions); err != nil {
			sS.bkpSessionIDsMux.Unlock()
			return err
		}
	}
	for _, sess := range storedSessions {
		delete(sS.bkpSessionIDs, sess.CGRID)
	}
	sS.bkpSessionIDsMux.Unlock()
	sS.removeSsCGRIDsMux.Lock()
	defer sS.removeSsCGRIDsMux.Unlock()
	for cgrID := range sS.removeSsCGRIDs {
		if err := sS.dm.RemoveSessionsBackup(sS.cgrCfg.GeneralCfg().NodeID,
			sS.cgrCfg.GeneralCfg().DefaultTenant, cgrID); err != nil {
			return err
		}
		delete(sS.removeSsCGRIDs, cgrID)
	}
	return nil
}

// storeSessions clears current sessions stored in datadb, and stores active sessions for backup in it
func (sS *SessionS) storeSessions() (sessStored int, err error) {
	sS.storeSessMux.Lock() // prevents concurrent execution of the function
	defer sS.storeSessMux.Unlock()
	activeSess := sS.getSessions(utils.EmptyString, false)
	// remove all sessions from dataDB backup if any
	if err := sS.dm.RemoveSessionsBackup(sS.cgrCfg.GeneralCfg().NodeID,
		sS.cgrCfg.GeneralCfg().DefaultTenant, utils.EmptyString); err != nil {
		return 0, err
	}
	if len(activeSess) == 0 {
		return
	}
	var storedSessions []*engine.StoredSession
	for _, sess := range activeSess {
		activeSess[0].lk.RLock()
		storedSessions = append(storedSessions, sess.asStoredSession())
		activeSess[0].lk.RUnlock()
	}
	if err := sS.dm.SetBackupSessions(sS.cgrCfg.GeneralCfg().NodeID,
		sS.cgrCfg.GeneralCfg().DefaultTenant, storedSessions); err != nil {
		return 0, err
	}
	return len(activeSess), nil
}

// BiRPCv1BackupActiveSessions will store all active sessions in dataDB and reply with the amount of sessions it stored
func (sS *SessionS) BiRPCv1BackupActiveSessions(ctx *context.Context,
	args string, reply *int) error {
	if sessCount, err := sS.storeSessions(); err != nil {
		return err
	} else {
		*reply = sessCount
	}
	return nil
}

// BiRPCv1TestSessToThresh is used to test birpc calls. unfinished remove later
func (sS *SessionS) BiRPCv1TestSessToThresh(_ *context.Context,
	args *string, rply *[]string) (err error) {
	if err = sS.connMgr.Call(sS.ctx, sS.cgrCfg.SessionSCfg().ThresholdSConns,
		utils.ThresholdSv1GetThresholdIDs, &utils.TenantWithAPIOpts{
			Tenant: sS.cgrCfg.GeneralCfg().DefaultTenant,
		}, rply); err != nil {
		utils.Logger.Warning(
			fmt.Sprintf(
				"<%s> could not test and get thresholdsIDs. reply <%+v> err: %s",
				utils.SessionS, rply, err.Error()))
	}
	return
}
