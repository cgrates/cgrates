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
	"runtime"
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
		biJClnts:      make(map[birpc.ClientConnector]string),
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
	conn  birpc.ClientConnector // connection towards BiJ client
	proto float64               // client protocol version
}

// SessionS represents the session service
type SessionS struct {
	cgrCfg  *config.CGRConfig // Separate from smCfg since there can be multiple
	dm      *engine.DataManager
	connMgr *engine.ConnManager

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
				sS.syncSessions(context.TODO())
			}
		}
	}
}

// Shutdown is called by engine to clear states
func (sS *SessionS) Shutdown() (err error) {
	if len(sS.cgrCfg.SessionSCfg().ReplicationConns) == 0 {
		var hasErr bool
		for _, s := range sS.getSessions("", false) { // Force sessions shutdown
			if err = sS.terminateSession(context.TODO(), s, nil, nil, nil, false); err != nil {
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
func (sS *SessionS) OnBiJSONConnect(c birpc.ClientConnector) {
	nodeID := utils.UUIDSha1Prefix() // connection identifier, should be later updated as login procedure
	sS.biJMux.Lock()
	sS.biJClnts[c] = nodeID
	sS.biJIDs[nodeID] = &biJClient{
		conn:  c,
		proto: sS.cgrCfg.SessionSCfg().ClientProtocol}
	sS.biJMux.Unlock()
}

// OnBiJSONDisconnect is called by rpc2.Client on each client disconnection
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
	if opts.HasField(utils.OptsSesTTL) {
		ttl, err = opts.GetDuration(utils.OptsSesTTL)
	} else if s.OptsStart.HasField(utils.OptsSesTTL) {
		ttl, err = s.OptsStart.GetDuration(utils.OptsSesTTL)
	} else {
		ttl = sS.cgrCfg.SessionSCfg().SessionTTL
	}
	if err != nil {
		utils.Logger.Warning(
			fmt.Sprintf("<%s>, cannot extract <%s> for session:<%s>, from it's options: <%s>, err: <%s>",
				utils.SessionS, utils.OptsSesTTL, s.CGRID, opts, err))
		return
	}
	if ttl == 0 {
		return // nothing to set up
	}
	// random delay computation
	var maxDelay time.Duration
	if opts.HasField(utils.OptsSesTTLMaxDelay) {
		maxDelay, err = opts.GetDuration(utils.OptsSesTTLMaxDelay)
	} else if s.OptsStart.HasField(utils.OptsSesTTLMaxDelay) {
		maxDelay, err = s.OptsStart.GetDuration(utils.OptsSesTTLMaxDelay)
	} else if sS.cgrCfg.SessionSCfg().SessionTTLMaxDelay != nil {
		maxDelay = *sS.cgrCfg.SessionSCfg().SessionTTLMaxDelay
	}
	if err != nil {
		utils.Logger.Warning(
			fmt.Sprintf("<%s>, cannot extract <%s> for session:<%s>, from it's options: <%s>, err: <%s>",
				utils.SessionS, utils.OptsSesTTLMaxDelay, s.CGRID, opts.String(), err.Error()))
		return
	}
	if maxDelay != 0 {
		rand.Seed(time.Now().Unix())
		ttl += time.Duration(
			rand.Int63n(maxDelay.Milliseconds()) * time.Millisecond.Nanoseconds())
	}
	// LastUsed
	var ttlLastUsed *time.Duration
	if opts.HasField(utils.OptsSesTTLLastUsed) {
		ttlLastUsed, err = opts.GetDurationPtr(utils.OptsSesTTLLastUsed)
	} else if s.OptsStart.HasField(utils.OptsSesTTLLastUsed) {
		ttlLastUsed, err = s.OptsStart.GetDurationPtr(utils.OptsSesTTLLastUsed)
	} else {
		ttlLastUsed = sS.cgrCfg.SessionSCfg().SessionTTLLastUsed
	}
	if err != nil {
		utils.Logger.Warning(
			fmt.Sprintf("<%s>, cannot extract <%s> for session:<%s>, from it's options: <%s>, err: <%s>",
				utils.SessionS, utils.OptsSesTTLLastUsed, s.CGRID, opts.String(), err.Error()))
		return
	}
	// LastUsage
	var ttlLastUsage *time.Duration
	if opts.HasField(utils.OptsSesTTLLastUsage) {
		ttlLastUsage, err = opts.GetDurationPtr(utils.OptsSesTTLLastUsage)
	} else if s.OptsStart.HasField(utils.OptsSesTTLLastUsage) {
		ttlLastUsage, err = s.OptsStart.GetDurationPtr(utils.OptsSesTTLLastUsage)
	} else {
		ttlLastUsage = sS.cgrCfg.SessionSCfg().SessionTTLLastUsage
	}
	if err != nil {
		utils.Logger.Warning(
			fmt.Sprintf("<%s>, cannot extract <%s> for session:<%s>, from it's options: <%s>, err: <%s>",
				utils.SessionS, utils.OptsSesTTLLastUsage, s.CGRID, opts.String(), err.Error()))
		return
	}
	// TTLUsage
	var ttlUsage *time.Duration
	if opts.HasField(utils.OptsSesTTLUsage) {
		ttlUsage, err = opts.GetDurationPtr(utils.OptsSesTTLUsage)
	} else if s.OptsStart.HasField(utils.OptsSesTTLUsage) {
		ttlUsage, err = s.OptsStart.GetDurationPtr(utils.OptsSesTTLUsage)
	} else {
		ttlUsage = sS.cgrCfg.SessionSCfg().SessionTTLUsage
	}
	if err != nil {
		utils.Logger.Warning(
			fmt.Sprintf("<%s>, cannot extract <%s> for session:<%s>, from it's options: <%s>, err: <%s>",
				utils.SessionS, utils.OptsSesTTLUsage, s.CGRID, opts.String(), err.Error()))
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
			sS.forceSTerminate(context.TODO(), s, lastUsage, s.sTerminator.ttlUsage,
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
func (sS *SessionS) forceSTerminate(ctx *context.Context, s *Session, extraUsage time.Duration, tUsage, lastUsed *time.Duration) (err error) {
	if extraUsage != 0 {
		for i := range s.SRuns {
			if _, err = sS.debitSession(ctx, s, i, extraUsage, lastUsed); err != nil {
				utils.Logger.Warning(
					fmt.Sprintf(
						"<%s> failed debitting cgrID %s, sRunIdx: %d, err: %s",
						utils.SessionS, s.cgrID(), i, err.Error()))
			}
		}
	}
	// we apply the correction before
	if err = sS.endSession(ctx, s, tUsage, lastUsed, nil, false); err != nil {
		utils.Logger.Warning(
			fmt.Sprintf(
				"<%s> failed force terminating session with ID <%s>, err: <%s>",
				utils.SessionS, s.cgrID(), err.Error()))
	}
	// post the CDRs
	if len(sS.cgrCfg.SessionSCfg().CDRsConns) != 0 {
		var reply string
		for _, cgrEv := range s.asCGREvents() {
			if cgrEv.APIOpts == nil {
				cgrEv.APIOpts = make(map[string]interface{})
			}
			cgrEv.APIOpts[utils.OptsAttributeS] = false
			cgrEv.APIOpts[utils.OptsChargerS] = false
			if unratedReqs.HasField( // order additional rating for unrated request types
				engine.MapEvent(cgrEv.Event).GetStringIgnoreErrors(utils.RequestType)) {
				// argsProc.Flags = append(argsProc.Flags, utils.MetaRALs)
			}
			cgrEv.SetCloneable(true)
			if err = sS.connMgr.Call(ctx, sS.cgrCfg.SessionSCfg().CDRsConns,
				utils.CDRsV1ProcessEvent, cgrEv, &reply); err != nil {
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
		if s.OptsStart == nil {
			s.OptsStart = make(engine.MapEvent)
		}
		args := &utils.CGREvent{
			Tenant:  s.Tenant,
			ID:      utils.GenUUID(),
			Event:   s.EventStart,
			APIOpts: s.OptsStart,
		}
		args.APIOpts[utils.OptsResourcesUsageID] = s.ResourceID
		args.APIOpts[utils.OptsResourcesUnits] = 1
		if err := sS.connMgr.Call(ctx, sS.cgrCfg.SessionSCfg().ResSConns,
			utils.ResourceSv1ReleaseResources,
			args, &reply); err != nil {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> error: %s could not release resource with resourceID: %s",
					utils.SessionS, err.Error(), s.ResourceID))
		}
	}
	sS.replicateSessions(ctx, s.CGRID, false, sS.cgrCfg.SessionSCfg().ReplicationConns)
	if clntConn := sS.biJClnt(s.ClientConnID); clntConn != nil {
		go func() {
			var rply string
			if err := clntConn.conn.Call(ctx,
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
func (sS *SessionS) debitSession(ctx *context.Context, s *Session, sRunIdx int, dur time.Duration,
	lastUsed *time.Duration) (maxDur time.Duration, err error) {

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
		if maxDebit, err = sS.debitSession(context.TODO(), s, sRunIdx, dbtIvl, nil); err != nil {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> could not complete debit operation on session: <%s>, error: <%s>",
					utils.SessionS, s.cgrID(), err.Error()))
			dscReason := utils.ErrServerError.Error()
			if err.Error() == utils.ErrUnauthorizedDestination.Error() {
				dscReason = err.Error()
			}
			// try to disconect the session n times before we force terminate it on our side
			fib := utils.Fib()
			for i := 0; i < sS.cgrCfg.SessionSCfg().TerminateAttempts; i++ {
				if i != 0 { // not the first iteration
					time.Sleep(time.Duration(fib()) * time.Millisecond)
				}
				if err = sS.disconnectSession(s, dscReason); err == nil {
					s.Unlock()
					return
				}
				utils.Logger.Warning(
					fmt.Sprintf("<%s> could not disconnect session: %s, error: %s",
						utils.SessionS, s.cgrID(), err.Error()))
			}
			if err = sS.forceSTerminate(context.TODO(), s, 0, nil, nil); err != nil {
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
		sS.replicateSessions(context.TODO(), s.CGRID, false, sS.cgrCfg.SessionSCfg().ReplicationConns)
		if maxDebit < dbtIvl { // disconnect faster
			select {
			case <-debitStop: // call was disconnected already
				return
			case <-time.After(maxDebit):
				s.Lock()
				defer s.Unlock()
				// try to disconect the session n times before we force terminate it on our side
				fib := utils.Fib()
				for i := 0; i < sS.cgrCfg.SessionSCfg().TerminateAttempts; i++ {
					if i != 0 { // not the first iteration
						time.Sleep(time.Duration(fib()) * time.Millisecond)
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
				if err = sS.forceSTerminate(context.TODO(), s, 0, nil, nil); err != nil {
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
	if err = clnt.conn.Call(context.TODO(), servMethod,
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
	if err = clnt.conn.Call(context.TODO(), utils.SessionSv1WarnDisconnect,
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
func (sS *SessionS) replicateSessions(ctx *context.Context, cgrID string, psv bool, connIDs []string) {
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
		if err := sS.connMgr.Call(ctx, connIDs,
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
			// ssIndx[fieldName][fieldVal][s.CGRID].Add(sr.CD.RunID)

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

func (sS *SessionS) getIndexedFilters(ctx *context.Context, tenant string, fltrs []string) (
	indexedFltr map[string][]string, unindexedFltr []*engine.FilterRule) {
	indexedFltr = make(map[string][]string)
	for _, fltrID := range fltrs {
		f, err := sS.dm.GetFilter(ctx, tenant, fltrID,
			true, true, utils.NonTransactional)
		if err != nil {
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
func (sS *SessionS) filterSessions(ctx *context.Context, sf *utils.SessionFilter, psv bool) (aSs []*ExternalSession) {
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
	indx, unindx := sS.getIndexedFilters(ctx, tenant, sf.Filters)
	cgrIDs, _ /*matchingSRuns*/ := sS.getSessionIDsMatchingIndexes(indx, psv)
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
			if pass, err = fltr.Pass(ctx, ev); err != nil || !pass {
				pass = false
				return
			}
		}
		return
	}
	for _, s := range ss {
		s.RLock()
		// runIDs := matchingSRuns[s.CGRID]
		for _, sr := range s.SRuns {
			// if len(cgrIDs) != 0 && !runIDs.Has(sr.CD.RunID) {
			// continue
			// }
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
func (sS *SessionS) filterSessionsCount(ctx *context.Context, sf *utils.SessionFilter, psv bool) (count int) {
	count = 0
	if len(sf.Filters) == 0 {
		ss := sS.getSessions(utils.EmptyString, psv)
		for _, s := range ss {
			count += len(s.SRuns)
		}
		return
	}
	tenant := utils.FirstNonEmpty(sf.Tenant, sS.cgrCfg.GeneralCfg().DefaultTenant)
	indx, unindx := sS.getIndexedFilters(ctx, tenant, sf.Filters)
	cgrIDs, _ /* matchingSRuns*/ := sS.getSessionIDsMatchingIndexes(indx, psv)
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
			if pass, err = fltr.Pass(ctx, ev); err != nil || !pass {
				return
			}
		}
		return
	}
	for _, s := range ss {
		s.RLock()
		// runIDs := matchingSRuns[s.CGRID]
		for _, sr := range s.SRuns {
			// if len(cgrIDs) != 0 && !runIDs.Has(sr.CD.RunID) {
			// continue
			// }
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
func (sS *SessionS) newSession(ctx *context.Context, cgrEv *utils.CGREvent, resID, clntConnID string,
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
		ResourceID:    resID,
		EventStart:    evStart.Clone(), // decouple the event from the request so we can avoid concurrency with debit and ttl
		OptsStart:     engine.MapEvent(cgrEv.APIOpts).Clone(),
		ClientConnID:  clntConnID,
		DebitInterval: dbtItval,
	}
	s.chargeable = s.OptsStart.GetBoolOrDefault(utils.OptsSesChargeable, true)
	if !isMsg && sS.isIndexed(s, false) { // check if already exists
		return nil, utils.ErrExists
	}

	var chrgrs []*engine.ChrgSProcessEventReply
	if chrgrs, err = sS.processChargerS(ctx, cgrEv); err != nil {
		return
	}
	s.SRuns = make([]*SRun, len(chrgrs))
	for i, chrgr := range chrgrs {
		me := engine.MapEvent(chrgr.CGREvent.Event)
		s.SRuns[i] = &SRun{
			Event: me,
		}
	}
	return
}

// processChargerS processes the event with chargers and cahces the response based on the requestID
func (sS *SessionS) processChargerS(ctx *context.Context, cgrEv *utils.CGREvent) (chrgrs []*engine.ChrgSProcessEventReply, err error) {
	if x, ok := engine.Cache.Get(utils.CacheEventCharges, cgrEv.ID); ok && x != nil {
		return x.([]*engine.ChrgSProcessEventReply), nil
	}
	if err = sS.connMgr.Call(ctx, sS.cgrCfg.SessionSCfg().ChargerSConns,
		utils.ChargerSv1ProcessEvent, cgrEv, &chrgrs); err != nil {
		err = utils.NewErrChargerS(err)
	}

	if errCh := engine.Cache.Set(ctx, utils.CacheEventCharges, cgrEv.ID, chrgrs, nil,
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
func (sS *SessionS) relocateSession(ctx *context.Context, initOriginID, originID, originHost string) (s *Session) {
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
	sS.replicateSessions(ctx, initCGRID, false, sS.cgrCfg.SessionSCfg().ReplicationConns)
	return
}

// getRelocateSession will relocate a session if it cannot find cgrID and initialOriginID is present
func (sS *SessionS) getRelocateSession(ctx *context.Context, cgrID string, initOriginID,
	originID, originHost string) (s *Session) {
	if s = sS.getActivateSession(cgrID); s != nil ||
		initOriginID == "" {
		return
	}
	return sS.relocateSession(ctx, initOriginID, originID, originHost)
}

// syncSessions synchronizes the active sessions with the one in the clients
// it will force-disconnect the one found in SessionS but not in clients
func (sS *SessionS) syncSessions(ctx *context.Context) {
	sS.aSsMux.RLock()
	asCount := len(sS.aSessions)
	sS.aSsMux.RUnlock()
	if asCount == 0 { // no need to sync the sessions if none is active
		return
	}
	type qReply struct {
		reply []*SessionID
		err   error
	}
	biClnts := sS.biJClients()
	replys := make(chan *qReply, len(biClnts))

	for _, clnt := range biClnts {
		ctx, cancel := context.WithTimeout(ctx, sS.cgrCfg.GeneralCfg().ReplyTimeout)
		defer cancel()
		go func(clnt *biJClient) {
			var reply qReply
			reply.err = clnt.conn.Call(ctx, utils.SessionSv1GetActiveSessionIDs,
				utils.EmptyString, &reply.reply)
			replys <- &reply
		}(clnt)
	}
	queriedCGRIDs := utils.StringSet{}
	for range biClnts {
		reply := <-replys
		if reply.err != nil {
			if reply.err.Error() != utils.ErrNoActiveSession.Error() {
				utils.Logger.Warning(
					fmt.Sprintf("<%s> error <%s> quering session ids", utils.SessionS, reply.err.Error()))
			}
			continue
		}
		for _, sessionID := range reply.reply {
			queriedCGRIDs.Add(sessionID.CGRID())
		}
	}
	var toBeRemoved []string
	sS.aSsMux.RLock()
	for cgrid := range sS.aSessions {
		if !queriedCGRIDs.Has(cgrid) {
			toBeRemoved = append(toBeRemoved, cgrid)
		}
	}
	sS.aSsMux.RUnlock()
	sS.terminateSyncSessions(ctx, toBeRemoved)
}

// Extracted from syncSessions in order to test all cases
func (sS *SessionS) terminateSyncSessions(ctx *context.Context, toBeRemoved []string) {
	for _, cgrID := range toBeRemoved {
		ss := sS.getSessions(cgrID, false)
		if len(ss) == 0 {
			continue
		}
		ss[0].Lock()
		if err := sS.forceSTerminate(ctx, ss[0], 0, nil, nil); err != nil {
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
func (sS *SessionS) authEvent(ctx *context.Context, cgrEv *utils.CGREvent, forceDuration bool) (usage map[string]time.Duration, err error) {
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
	if s, err = sS.newSession(ctx, cgrEv, "", "", 0, forceDuration, true); err != nil {
		return
	}
	usage = make(map[string]time.Duration)
	for _, sr := range s.SRuns {
		var rplyMaxUsage time.Duration
		if !authReqs.HasField(
			sr.Event.GetStringIgnoreErrors(utils.RequestType)) {
			rplyMaxUsage = eventUsage
			// } else if err = sS.connMgr.Call(sS.cgrCfg.SessionSCfg().RALsConns, nil,
			// 	utils.ResponderGetMaxSessionTime,
			// 	&engine.CallDescriptorWithAPIOpts{
			// 		CallDescriptor: sr.CD,
			// 		APIOpts:        s.OptsStart,
			// 	}, &rplyMaxUsage); err != nil {
			// 	err = utils.NewErrRALs(err)
			// 	return
		}
		if rplyMaxUsage > eventUsage {
			rplyMaxUsage = eventUsage
		}
		// usage[sr.CD.RunID] = rplyMaxUsage
	}
	return
}

// initSession handles a new session
// not thread-safe for Session since it is constructed here
func (sS *SessionS) initSession(ctx *context.Context, cgrEv *utils.CGREvent, clntConnID,
	resID string, dbtItval time.Duration, isMsg, forceDuration bool) (s *Session, err error) {
	if s, err = sS.newSession(ctx, cgrEv, resID, clntConnID, dbtItval, forceDuration, isMsg); err != nil {
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
func (sS *SessionS) updateSession(ctx *context.Context, s *Session, updtEv, opts engine.MapEvent, isMsg bool) (maxUsage map[string]time.Duration, err error) {
	if !isMsg {
		defer sS.replicateSessions(ctx, s.CGRID, false, sS.cgrCfg.SessionSCfg().ReplicationConns)
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
	s.chargeable = opts.GetBoolOrDefault(utils.OptsSesChargeable, true)
	//init has no updtEv
	if updtEv == nil {
		updtEv = engine.MapEvent(s.EventStart.Clone())
	}

	var reqMaxUsage time.Duration
	if _, err = updtEv.GetDuration(utils.Usage); err != nil {
		if err != utils.ErrNotFound {
			return
		}
		err = nil
		reqMaxUsage = sS.cgrCfg.SessionSCfg().GetDefaultUsage(updtEv.GetStringIgnoreErrors(utils.ToR))
		updtEv[utils.Usage] = reqMaxUsage
	}
	maxUsage = make(map[string]time.Duration)
	for _, sr := range s.SRuns {
		reqType := sr.Event.GetStringIgnoreErrors(utils.RequestType)
		if reqType == utils.MetaNone {
			//	maxUsage[sr.CD.RunID] = reqMaxUsage
			continue
		}
		// var rplyMaxUsage time.Duration
		// if reqType != utils.MetaPrepaid || s.debitStop != nil {
		// 	rplyMaxUsage = reqMaxUsage
		// } else if rplyMaxUsage, err = sS.debitSession(s, i, reqMaxUsage,
		// 	updtEv.GetDurationPtrIgnoreErrors(utils.LastUsed)); err != nil {
		// 	return
		// }
		// maxUsage[sr.CD.RunID] = rplyMaxUsage
	}
	return
}

// terminateSession will end a session from outside
// calls endSession thread safe
func (sS *SessionS) terminateSession(ctx *context.Context, s *Session, tUsage, lastUsage *time.Duration,
	aTime *time.Time, isMsg bool) (err error) {
	s.Lock()
	err = sS.endSession(ctx, s, tUsage, lastUsage, aTime, isMsg)
	s.Unlock()
	return
}

// endSession will end a session from outside
// this function is not thread safe
func (sS *SessionS) endSession(ctx *context.Context, s *Session, tUsage, lastUsage *time.Duration,
	aTime *time.Time, isMsg bool) (err error) {
	if !isMsg {
		//check if we have replicate connection and close the session there
		defer sS.replicateSessions(ctx, s.CGRID, true, sS.cgrCfg.SessionSCfg().ReplicationConns)
		sS.unregisterSession(s.CGRID, false)
		s.stopSTerminator()
		s.stopDebitLoops()
	}
	for sRunIdx, sr := range s.SRuns {
		// sUsage := sr.TotalUsage
		// if tUsage != nil {
		// sUsage = *tUsage
		// sr.TotalUsage = *tUsage
		// } else if lastUsage != nil &&
		// sr.LastUsage != *lastUsage {
		// sr.TotalUsage -= sr.LastUsage
		// sr.TotalUsage += *lastUsage
		// sUsage = sr.TotalUsage
		// }
		// if sr.EventCost != nil {
		// 	if !isMsg { // in case of one time charge there is no need of corrections
		// 		if notCharged := sUsage - sr.EventCost.GetUsage(); notCharged > 0 { // we did not charge enough, make a manual debit here
		// 			if !s.chargeable {
		// 				sS.pause(sr, notCharged)
		// 			} else {
		// 				if sr.CD.LoopIndex > 0 {
		// 					sr.CD.TimeStart = sr.CD.TimeEnd
		// 				}
		// 				sr.CD.TimeEnd = sr.CD.TimeStart.Add(notCharged)
		// 				sr.CD.DurationIndex += notCharged
		// 				cc := new(engine.CallCost)
		// 				if err = sS.connMgr.Call(sS.cgrCfg.SessionSCfg().RALsConns, nil, utils.ResponderDebit,
		// 					&engine.CallDescriptorWithAPIOpts{
		// 						CallDescriptor: sr.CD,
		// 						APIOpts:        s.OptsStart,
		// 					}, cc); err == nil {
		// 					sr.EventCost.Merge(
		// 						engine.NewEventCostFromCallCost(cc, s.CGRID,
		// 							sr.Event.GetStringIgnoreErrors(utils.RunID)))
		// 				}
		// 			}
		// 		} else if notCharged < 0 { // charged too much, try refund
		// 			if err = sS.refundSession(s, sRunIdx, -notCharged); err != nil {
		// 				utils.Logger.Warning(
		// 					fmt.Sprintf(
		// 						"<%s> failed refunding session: <%s>, srIdx: <%d>, error: <%s>",
		// 						utils.SessionS, s.CGRID, sRunIdx, err.Error()))
		// 			}
		// 		}
		// 		if err := sS.roundCost(s, sRunIdx); err != nil { // will round the cost and refund the extra increment
		// 			utils.Logger.Warning(
		// 				fmt.Sprintf("<%s> failed rounding  session cost for <%s>, srIdx: <%d>, error: <%s>",
		// 					utils.SessionS, s.CGRID, sRunIdx, err.Error()))
		// 		}
		// 	}
		// 	// compute the event cost before saving the SessionCost
		// 	// add here to be applied for messages also
		// 	sr.EventCost.Compute()
		// 	if sS.cgrCfg.SessionSCfg().StoreSCosts {
		// 		if err := sS.storeSCost(s, sRunIdx); err != nil {
		// 			utils.Logger.Warning(
		// 				fmt.Sprintf("<%s> failed storing session cost for <%s>, srIdx: <%d>, error: <%s>",
		// 					utils.SessionS, s.CGRID, sRunIdx, err.Error()))
		// 		}
		// 	}

		// 	// set cost fields
		// 	sr.Event[utils.Cost] = sr.EventCost.GetCost()
		// 	sr.Event[utils.CostDetails] = utils.ToJSON(sr.EventCost) // avoid map[string]interface{} when decoding
		// 	sr.Event[utils.CostSource] = utils.MetaSessionS
		// }
		// Set Usage field
		if sRunIdx == 0 {
			s.EventStart[utils.Usage] = sr.TotalUsage
		}
		sr.Event[utils.Usage] = sr.TotalUsage
		if aTime != nil {
			sr.Event[utils.AnswerTime] = *aTime
		}
	}
	if errCh := engine.Cache.Set(ctx, utils.CacheClosedSessions, s.CGRID, s,
		nil, true, utils.NonTransactional); errCh != nil {
		return errCh
	}
	return
}

// chargeEvent will charge a single event (ie: SMS)
func (sS *SessionS) chargeEvent(ctx *context.Context, cgrEv *utils.CGREvent, forceDuration bool) (maxUsage time.Duration, err error) {
	var s *Session
	if s, err = sS.initSession(ctx, cgrEv, "", "", 0, true, forceDuration); err != nil {
		return
	}
	cgrID := s.CGRID
	var sRunsUsage map[string]time.Duration
	if sRunsUsage, err = sS.updateSession(ctx, s, nil, nil, true); err != nil {
		if errEnd := sS.terminateSession(ctx, s,
			utils.DurationPointer(time.Duration(0)), nil, nil, true); errEnd != nil {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> error when force-ending charged event: <%s>, err: <%s>",
					utils.SessionS, cgrID, errEnd.Error()))
		}
		// err = utils.NewErrRALs(err)
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
	if errEnd := sS.terminateSession(ctx, s, utils.DurationPointer(usage), nil, nil, true); errEnd != nil {
		utils.Logger.Warning(
			fmt.Sprintf("<%s> error when ending charged event: <%s>, err: <%s>",
				utils.SessionS, cgrID, errEnd.Error()))
	}
	return // returns here the maxUsage from update
}

// accounSMaxAbstracts computes the maximum abstract units for the events provided
func (sS *SessionS) accounSMaxAbstracts(ctx *context.Context, cgrEvs []*utils.CGREvent) (maxAbstracts *utils.Decimal, err error) {
	if len(sS.cgrCfg.SessionSCfg().AttrSConns) == 0 {
		return nil, utils.NewErrNotConnected(utils.AccountS)
	}
	for _, cgrEv := range cgrEvs {
		acntCost := new(utils.ExtEventCharges)
		if err = sS.connMgr.Call(ctx, sS.cgrCfg.SessionSCfg().AttrSConns, // Fix Here with AccountS
			utils.AccountSv1DebitAbstracts, cgrEv, &acntCost); err != nil {
			return
		} else if maxAbstracts == nil ||
			maxAbstracts.Compare(utils.NewDecimalFromFloat64(*acntCost.Abstracts)) == 1 { // should compare directly against Decimal
			maxAbstracts = utils.NewDecimalFromFloat64(*acntCost.Abstracts) // did not optimize here since we need to remove floats from acntCost
		}
	}
	return
}

// APIs start here

// BiRPCv1GetActiveSessions returns the list of active sessions based on filter
func (sS *SessionS) BiRPCv1GetActiveSessions(ctx *context.Context,
	args *utils.SessionFilter, reply *[]*ExternalSession) (err error) {
	if args == nil { //protection in case on nil
		args = &utils.SessionFilter{}
	}
	aSs := sS.filterSessions(ctx, args, false)
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
	*reply = sS.filterSessionsCount(ctx, args, false)
	return nil
}

// BiRPCv1GetPassiveSessions returns the passive sessions handled by SessionS
func (sS *SessionS) BiRPCv1GetPassiveSessions(ctx *context.Context,
	args *utils.SessionFilter, reply *[]*ExternalSession) error {
	if args == nil { //protection in case on nil
		args = &utils.SessionFilter{}
	}
	pSs := sS.filterSessions(ctx, args, true)
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
	*reply = sS.filterSessionsCount(ctx, args, true)
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

// BiRPCv1ReplicateSessions will replicate active sessions to either args.Connections or the internal configured ones
// args.Filter is used to filter the sessions which are replicated, CGRID is the only one possible for now
func (sS *SessionS) BiRPCv1ReplicateSessions(ctx *context.Context,
	args ArgsReplicateSessions, reply *string) (err error) {
	sS.replicateSessions(ctx, args.CGRID, args.Passive, args.ConnIDs)
	*reply = utils.OK
	return
}

// BiRPCv1AuthorizeEvent performs authorization for CGREvent based on specific components
func (sS *SessionS) BiRPCv1AuthorizeEvent(ctx *context.Context,
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
		defer engine.Cache.Set(ctx, utils.CacheRPCResponses, cacheKey,
			&utils.CachedRPCResponse{Result: authReply, Error: err},
			nil, true, utils.NonTransactional)
	}
	// end of RPC caching

	if !(args.GetAttributes || utils.OptAsBool(args.APIOpts, utils.OptsAttributeS) ||
		args.GetMaxUsage || utils.OptAsBool(args.APIOpts, utils.OptsSesMaxUsage) ||
		args.AuthorizeResources || utils.OptAsBool(args.APIOpts, utils.OptsSesResourceSAuthorize) ||
		args.GetRoutes || utils.OptAsBool(args.APIOpts, utils.OptsRouteS)) {
		return // Nothing to do
	}
	if args.APIOpts == nil {
		args.APIOpts = make(map[string]interface{})
	}

	if args.GetAttributes ||
		utils.OptAsBool(args.APIOpts, utils.OptsAttributeS) {
		if args.APIOpts == nil {
			args.APIOpts = make(map[string]interface{})
		}
		if args.AttributeIDs != nil {
			args.APIOpts[utils.OptsAttributesAttributeIDs] = args.AttributeIDs
		}
		rplyAttr, err := sS.processAttributes(ctx, args.CGREvent)
		if err == nil {
			args.CGREvent = rplyAttr.CGREvent
			authReply.Attributes = &rplyAttr
		} else if err.Error() != utils.ErrNotFound.Error() {
			return utils.NewErrAttributeS(err)
		}
	}
	if args.GetMaxUsage ||
		utils.OptAsBool(args.APIOpts, utils.OptsSesMaxUsage) {
		var sRunsUsage map[string]time.Duration
		if sRunsUsage, err = sS.authEvent(ctx, args.CGREvent,
			args.ForceDuration || utils.OptAsBool(args.APIOpts, utils.OptsSesForceDuration)); err != nil {
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
	if args.AuthorizeResources ||
		utils.OptAsBool(args.APIOpts, utils.OptsSesResourceSAuthorize) {
		if len(sS.cgrCfg.SessionSCfg().ResSConns) == 0 {
			return utils.NewErrNotConnected(utils.ResourceS)
		}
		originID, _ := args.CGREvent.FieldAsString(utils.OriginID)
		if originID == "" {
			originID = utils.UUIDSha1Prefix()
		}
		args.APIOpts[utils.OptsResourcesUsageID] = originID
		args.APIOpts[utils.OptsResourcesUnits] = 1
		var allocMsg string
		if err = sS.connMgr.Call(ctx, sS.cgrCfg.SessionSCfg().ResSConns, utils.ResourceSv1AuthorizeResources,
			args, &allocMsg); err != nil {
			return utils.NewErrResourceS(err)
		}
		authReply.ResourceAllocation = &allocMsg
	}
	if args.GetRoutes ||
		utils.OptAsBool(args.APIOpts, utils.OptsRouteS) {
		args.APIOpts[utils.OptsRoutesMaxCost] = utils.FirstNonEmpty(args.RoutesMaxCost, utils.IfaceAsString(args.APIOpts[utils.OptsSesRouteSMaxCost]))
		args.APIOpts[utils.OptsRoutesIgnoreErrors] = args.RoutesIgnoreErrors || utils.OptAsBool(args.APIOpts, utils.OptsSesRouteSIgnoreErrors)
		routesReply, err := sS.getRoutes(ctx, args.CGREvent.Clone())
		if err != nil {
			return err
		}
		if routesReply != nil {
			authReply.RouteProfiles = routesReply
		}
	}
	if args.ProcessThresholds ||
		utils.OptAsBool(args.APIOpts, utils.OptsThresholdS) {
		var thIDs []string
		if thIDs, err = utils.OptAsStringSlice(args.APIOpts, utils.OptsSesThresholdIDs); err != nil {
			return
		}
		if thIDs == nil {
			thIDs = args.ThresholdIDs
		}
		tIDs, err := sS.processThreshold(ctx, args.CGREvent, thIDs, true)
		if err != nil && err.Error() != utils.ErrNotFound.Error() {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> error: %s processing event %+v with ThresholdS.",
					utils.SessionS, err.Error(), args.CGREvent))
			withErrors = true
		}
		authReply.ThresholdIDs = &tIDs
	}
	if args.ProcessStats ||
		utils.OptAsBool(args.APIOpts, utils.OptsStatS) {
		var statIDs []string
		if statIDs, err = utils.OptAsStringSlice(args.APIOpts, utils.OptsSesStatIDs); err != nil {
			return
		}
		if statIDs == nil {
			statIDs = args.StatIDs
		}
		sIDs, err := sS.processStats(ctx, args.CGREvent, statIDs, false)
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

// BiRPCv1AuthorizeEventWithDigest performs authorization for CGREvent based on specific components
// returning one level fields instead of multiple ones returned by BiRPCv1AuthorizeEvent
func (sS *SessionS) BiRPCv1AuthorizeEventWithDigest(ctx *context.Context,
	args *V1AuthorizeArgs, authReply *V1AuthorizeReplyWithDigest) (err error) {
	var initAuthRply V1AuthorizeReply
	if err = sS.BiRPCv1AuthorizeEvent(ctx, args, &initAuthRply); err != nil {
		return
	}
	if (args.GetAttributes ||
		utils.OptAsBool(args.APIOpts, utils.OptsAttributeS)) && initAuthRply.Attributes != nil {
		authReply.AttributesDigest = utils.StringPointer(initAuthRply.Attributes.Digest())
	}
	if args.AuthorizeResources ||
		utils.OptAsBool(args.APIOpts, utils.OptsSesResourceSAuthorize) {
		authReply.ResourceAllocation = initAuthRply.ResourceAllocation
	}
	if args.GetMaxUsage ||
		utils.OptAsBool(args.APIOpts, utils.OptsSesMaxUsage) {
		authReply.MaxUsage = initAuthRply.MaxUsage.Seconds()
	}
	if args.GetRoutes ||
		utils.OptAsBool(args.APIOpts, utils.OptsRouteS) {
		authReply.RoutesDigest = utils.StringPointer(initAuthRply.RouteProfiles.Digest())
	}
	if args.ProcessThresholds ||
		utils.OptAsBool(args.APIOpts, utils.OptsThresholdS) {
		authReply.Thresholds = utils.StringPointer(
			strings.Join(*initAuthRply.ThresholdIDs, utils.FieldsSep))
	}
	if args.ProcessStats ||
		utils.OptAsBool(args.APIOpts, utils.OptsStatS) {
		authReply.StatQueues = utils.StringPointer(
			strings.Join(*initAuthRply.StatQueueIDs, utils.FieldsSep))
	}
	return
}

// BiRPCv1InitiateSession initiates a new session
func (sS *SessionS) BiRPCv1InitiateSession(ctx *context.Context,
	args *utils.CGREvent, rply *V1InitSessionReply) (err error) {
	if args == nil {
		return utils.NewErrMandatoryIeMissing(utils.CGREventString)
	}
	var withErrors bool
	if args.ID == "" {
		args.ID = utils.GenUUID()
	}
	if args.Tenant == "" {
		args.Tenant = sS.cgrCfg.GeneralCfg().DefaultTenant
	}

	// RPC caching
	if sS.cgrCfg.CacheCfg().Partitions[utils.CacheRPCResponses].Limit != 0 {
		cacheKey := utils.ConcatenatedKey(utils.SessionSv1InitiateSession, args.ID)
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
		defer engine.Cache.Set(ctx, utils.CacheRPCResponses, cacheKey,
			&utils.CachedRPCResponse{Result: rply, Error: err},
			nil, true, utils.NonTransactional)
	}
	// end of RPC caching
	rply.MaxUsage = utils.DurationPointer(time.Duration(utils.InvalidUsage)) // temp

	attrS := utils.OptAsBool(args.APIOpts, utils.OptsAttributeS)
	initS := utils.OptAsBool(args.APIOpts, utils.OptsSesInitiate)
	resS := utils.OptAsBool(args.APIOpts, utils.OptsSesResourceSAlocate)
	if !(attrS || initS || resS) {
		return // nothing to do
	}
	originID, _ := args.FieldAsString(utils.OriginID)
	if attrS {
		rplyAttr, err := sS.processAttributes(ctx, args)
		if err == nil {
			args = rplyAttr.CGREvent
			rply.Attributes = &rplyAttr
		} else if err.Error() != utils.ErrNotFound.Error() {
			return utils.NewErrAttributeS(err)
		}
	}
	if resS {
		if len(sS.cgrCfg.SessionSCfg().ResSConns) == 0 {
			return utils.NewErrNotConnected(utils.ResourceS)
		}
		if originID == "" {
			return utils.NewErrMandatoryIeMissing(utils.OriginID)
		}
		args.APIOpts[utils.OptsResourcesUsageID] = originID
		args.APIOpts[utils.OptsResourcesUnits] = 1
		var allocMessage string
		if err = sS.connMgr.Call(ctx, sS.cgrCfg.SessionSCfg().ResSConns,
			utils.ResourceSv1AllocateResources, args, &allocMessage); err != nil {
			return utils.NewErrResourceS(err)
		}
		rply.ResourceAllocation = &allocMessage
	}
	if initS {
		var err error
		opts := engine.MapEvent(args.APIOpts)
		dbtItvl := sS.cgrCfg.SessionSCfg().DebitInterval
		if opts.HasField(utils.OptsSesDebitInterval) { // dynamic DebitInterval via CGRDebitInterval
			if dbtItvl, err = opts.GetDuration(utils.OptsSesDebitInterval); err != nil {
				return err //utils.NewErrRALs(err)
			}
		}
		s, err := sS.initSession(ctx, args, sS.biJClntID(ctx.Client), originID, dbtItvl,
			false, utils.OptAsBool(args.APIOpts, utils.OptsSesForceDuration))
		if err != nil {
			return err
		}
		s.RLock() // avoid concurrency with activeDebit
		isPrepaid := s.debitStop != nil
		s.RUnlock()
		if isPrepaid { //active debit
			rply.MaxUsage = utils.DurationPointer(sS.cgrCfg.SessionSCfg().GetDefaultUsage(utils.IfaceAsString(args.Event[utils.ToR])))
		} else {
			var sRunsUsage map[string]time.Duration
			if sRunsUsage, err = sS.updateSession(ctx, s, nil, args.APIOpts, false); err != nil {
				return err //utils.NewErrRALs(err)
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
	if utils.OptAsBool(args.APIOpts, utils.OptsThresholdS) {
		var thIDs []string
		if thIDs, err = utils.OptAsStringSlice(args.APIOpts, utils.OptsSesThresholdIDs); err != nil {
			return
		}
		tIDs, err := sS.processThreshold(ctx, args, thIDs, true)
		if err != nil && err.Error() != utils.ErrNotFound.Error() {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> error: %s processing event %+v with ThresholdS.",
					utils.SessionS, err.Error(), args))
			withErrors = true
		}
		rply.ThresholdIDs = &tIDs
	}
	if utils.OptAsBool(args.APIOpts, utils.OptsStatS) {
		var statIDs []string
		if statIDs, err = utils.OptAsStringSlice(args.APIOpts, utils.OptsSesStatIDs); err != nil {
			return
		}
		sIDs, err := sS.processStats(ctx, args, statIDs, false)
		if err != nil &&
			err.Error() != utils.ErrNotFound.Error() {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> error: %s processing event %+v with StatS.",
					utils.SessionS, err.Error(), args))
			withErrors = true
		}
		rply.StatQueueIDs = &sIDs
	}
	if withErrors {
		err = utils.ErrPartiallyExecuted
	}
	return
}

// BiRPCv1InitiateSessionWithDigest returns the formated result of InitiateSession
func (sS *SessionS) BiRPCv1InitiateSessionWithDigest(ctx *context.Context,
	args *utils.CGREvent, initReply *V1InitReplyWithDigest) (err error) {
	var initSessionRply V1InitSessionReply
	if err = sS.BiRPCv1InitiateSession(ctx, args, &initSessionRply); err != nil {
		return
	}

	if initSessionRply.Attributes != nil {
		initReply.AttributesDigest = utils.StringPointer(initSessionRply.Attributes.Digest())
	}

	initReply.ResourceAllocation = initSessionRply.ResourceAllocation

	//if initSessionRply.MaxUsage != nil {
	//	initReply.MaxUsage = initSessionRply.MaxUsage.Seconds()
	//}
	initReply.MaxUsage = utils.InvalidUsage // temp

	if initSessionRply.ThresholdIDs != nil {
		initReply.Thresholds = utils.StringPointer(
			strings.Join(*initSessionRply.ThresholdIDs, utils.FieldsSep))
	}
	if initSessionRply.StatQueueIDs != nil {
		initReply.StatQueues = utils.StringPointer(
			strings.Join(*initSessionRply.StatQueueIDs, utils.FieldsSep))
	}
	return
}

// BiRPCv1UpdateSession updates an existing session, returning the duration which the session can still last
func (sS *SessionS) BiRPCv1UpdateSession(ctx *context.Context,
	args *utils.CGREvent, rply *V1UpdateSessionReply) (err error) {
	if args == nil {
		return utils.NewErrMandatoryIeMissing(utils.CGREventString)
	}
	if args.ID == utils.EmptyString {
		args.ID = utils.GenUUID()
	}
	if args.Tenant == utils.EmptyString {
		args.Tenant = sS.cgrCfg.GeneralCfg().DefaultTenant
	}

	// RPC caching
	if sS.cgrCfg.CacheCfg().Partitions[utils.CacheRPCResponses].Limit != 0 {
		cacheKey := utils.ConcatenatedKey(utils.SessionSv1UpdateSession, args.ID)
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
		defer engine.Cache.Set(ctx, utils.CacheRPCResponses, cacheKey,
			&utils.CachedRPCResponse{Result: rply, Error: err},
			nil, true, utils.NonTransactional)
	}
	// end of RPC caching
	attrS := utils.OptAsBool(args.APIOpts, utils.OptsAttributeS)
	updS := utils.OptAsBool(args.APIOpts, utils.OptsSesUpdate)
	if !(attrS || updS) {
		return // nothing to do
	}

	if attrS {
		rplyAttr, err := sS.processAttributes(ctx, args)
		if err == nil {
			args = rplyAttr.CGREvent
			rply.Attributes = &rplyAttr
		} else if err.Error() != utils.ErrNotFound.Error() {
			return utils.NewErrAttributeS(err)
		}
	}
	if updS {
		ev := engine.MapEvent(args.Event)
		opts := engine.MapEvent(args.APIOpts)
		dbtItvl := sS.cgrCfg.SessionSCfg().DebitInterval
		if opts.HasField(utils.OptsSesDebitInterval) { // dynamic DebitInterval via CGRDebitInterval
			if dbtItvl, err = opts.GetDuration(utils.OptsSesDebitInterval); err != nil {
				return err //utils.NewErrRALs(err)
			}
		}
		cgrID := GetSetCGRID(ev)
		s := sS.getRelocateSession(ctx, cgrID,
			ev.GetStringIgnoreErrors(utils.InitialOriginID),
			ev.GetStringIgnoreErrors(utils.OriginID),
			ev.GetStringIgnoreErrors(utils.OriginHost))
		if s == nil {
			if s, err = sS.initSession(ctx, args, sS.biJClntID(ctx.Client), ev.GetStringIgnoreErrors(utils.OriginID),
				dbtItvl, false, utils.OptAsBool(args.APIOpts, utils.OptsSesForceDuration)); err != nil {
				return err
			}
		}
		var sRunsUsage map[string]time.Duration
		if sRunsUsage, err = sS.updateSession(ctx, s, ev, args.APIOpts, false); err != nil {
			return err //utils.NewErrRALs(err)
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

// BiRPCv1TerminateSession will stop debit loops as well as release any used resources
func (sS *SessionS) BiRPCv1TerminateSession(ctx *context.Context,
	args *utils.CGREvent, rply *string) (err error) {
	if args == nil {
		return utils.NewErrMandatoryIeMissing(utils.CGREventString)
	}
	var withErrors bool
	if args.ID == "" {
		args.ID = utils.GenUUID()
	}
	if args.Tenant == "" {
		args.Tenant = sS.cgrCfg.GeneralCfg().DefaultTenant
	}
	// RPC caching
	if sS.cgrCfg.CacheCfg().Partitions[utils.CacheRPCResponses].Limit != 0 {
		cacheKey := utils.ConcatenatedKey(utils.SessionSv1TerminateSession, args.ID)
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
		defer engine.Cache.Set(ctx, utils.CacheRPCResponses, cacheKey,
			&utils.CachedRPCResponse{Result: rply, Error: err},
			nil, true, utils.NonTransactional)
	}
	// end of RPC caching
	resS := utils.OptAsBool(args.APIOpts, utils.OptsSesResourceSRelease)
	termS := utils.OptAsBool(args.APIOpts, utils.OptsSesTerminate)
	if !(resS || termS) {
		return // nothing to do
	}

	ev := engine.MapEvent(args.Event)
	opts := engine.MapEvent(args.APIOpts)
	cgrID := GetSetCGRID(ev)
	originID := ev.GetStringIgnoreErrors(utils.OriginID)
	if termS {
		if originID == "" {
			return utils.NewErrMandatoryIeMissing(utils.OriginID)
		}
		dbtItvl := sS.cgrCfg.SessionSCfg().DebitInterval
		if opts.HasField(utils.OptsSesDebitInterval) { // dynamic DebitInterval via CGRDebitInterval
			if dbtItvl, err = opts.GetDuration(utils.OptsSesDebitInterval); err != nil {
				return err //utils.NewErrRALs(err)
			}
		}
		var s *Session
		fib := utils.Fib()
		var isMsg bool // one time charging, do not perform indexing and sTerminator
		for i := 0; i < sS.cgrCfg.SessionSCfg().TerminateAttempts; i++ {
			if s = sS.getRelocateSession(ctx, cgrID,
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
			if s, err = sS.initSession(ctx, args, sS.biJClntID(ctx.Client), ev.GetStringIgnoreErrors(utils.OriginID),
				dbtItvl, isMsg, utils.OptAsBool(args.APIOpts, utils.OptsSesForceDuration)); err != nil {
				return err //utils.NewErrRALs(err)
			}
			if _, err = sS.updateSession(ctx, s, ev, opts, isMsg); err != nil {
				return err
			}
			break
		}
		if !isMsg {
			s.UpdateSRuns(ev, sS.cgrCfg.SessionSCfg().AlterableFields)
		}
		s.Lock()
		s.chargeable = opts.GetBoolOrDefault(utils.OptsSesChargeable, true)
		s.Unlock()
		if err = sS.terminateSession(ctx, s,
			ev.GetDurationPtrIgnoreErrors(utils.Usage),
			ev.GetDurationPtrIgnoreErrors(utils.LastUsed),
			ev.GetTimePtrIgnoreErrors(utils.AnswerTime, utils.EmptyString),
			isMsg); err != nil {
			return err //utils.NewErrRALs(err)
		}
	}
	if resS {
		if len(sS.cgrCfg.SessionSCfg().ResSConns) == 0 {
			return utils.NewErrNotConnected(utils.ResourceS)
		}
		if originID == "" {
			return utils.NewErrMandatoryIeMissing(utils.OriginID)
		}
		args.APIOpts[utils.OptsResourcesUsageID] = originID
		args.APIOpts[utils.OptsResourcesUnits] = 1
		var reply string
		if err = sS.connMgr.Call(ctx, sS.cgrCfg.SessionSCfg().ResSConns, utils.ResourceSv1ReleaseResources,
			args, &reply); err != nil {
			return utils.NewErrResourceS(err)
		}
	}
	if utils.OptAsBool(args.APIOpts, utils.OptsThresholdS) {
		var thIDs []string
		if thIDs, err = utils.OptAsStringSlice(args.APIOpts, utils.OptsSesThresholdIDs); err != nil {
			return
		}
		_, err := sS.processThreshold(ctx, args, thIDs, true)
		if err != nil &&
			err.Error() != utils.ErrNotFound.Error() {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> error: %s processing event %+v with ThresholdS.",
					utils.SessionS, err.Error(), args))
			withErrors = true
		}
	}
	if utils.OptAsBool(args.APIOpts, utils.OptsStatS) {
		var statIDs []string
		if statIDs, err = utils.OptAsStringSlice(args.APIOpts, utils.OptsSesStatIDs); err != nil {
			return
		}
		_, err := sS.processStats(ctx, args, statIDs, false)
		if err != nil &&
			err.Error() != utils.ErrNotFound.Error() {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> error: %s processing event %+v with StatS.",
					utils.SessionS, err.Error(), args))
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
		defer engine.Cache.Set(ctx, utils.CacheRPCResponses, cacheKey,
			&utils.CachedRPCResponse{Result: rply, Error: err},
			nil, true, utils.NonTransactional)
	}
	// end of RPC caching
	// in case that source don't exist add it
	if _, has := cgrEv.Event[utils.Source]; !has {
		cgrEv.Event[utils.Source] = utils.MetaSessionS
	}

	return sS.processCDR(ctx, cgrEv, rply)
}

// BiRPCv1ProcessMessage processes one event with the right subsystems based on arguments received
func (sS *SessionS) BiRPCv1ProcessMessage(ctx *context.Context,
	args *utils.CGREvent, rply *V1ProcessMessageReply) (err error) {
	if args == nil {
		return utils.NewErrMandatoryIeMissing(utils.CGREventString)
	}
	var withErrors bool
	if args.ID == utils.EmptyString {
		args.ID = utils.GenUUID()
	}
	if args.Tenant == utils.EmptyString {
		args.Tenant = sS.cgrCfg.GeneralCfg().DefaultTenant
	}

	// RPC caching
	if sS.cgrCfg.CacheCfg().Partitions[utils.CacheRPCResponses].Limit != 0 {
		cacheKey := utils.ConcatenatedKey(utils.SessionSv1ProcessMessage, args.ID)
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
		defer engine.Cache.Set(ctx, utils.CacheRPCResponses, cacheKey,
			&utils.CachedRPCResponse{Result: rply, Error: err},
			nil, true, utils.NonTransactional)
	}
	// end of RPC caching

	me := engine.MapEvent(args.Event)
	originID := me.GetStringIgnoreErrors(utils.OriginID)

	if utils.OptAsBool(args.APIOpts, utils.OptsAttributeS) {
		rplyAttr, err := sS.processAttributes(ctx, args)
		if err == nil {
			args = rplyAttr.CGREvent
			rply.Attributes = &rplyAttr
		} else if err.Error() != utils.ErrNotFound.Error() {
			return utils.NewErrAttributeS(err)
		}
	}
	if utils.OptAsBool(args.APIOpts, utils.OptsSesResourceSAlocate) {
		if len(sS.cgrCfg.SessionSCfg().ResSConns) == 0 {
			return utils.NewErrNotConnected(utils.ResourceS)
		}
		if originID == "" {
			return utils.NewErrMandatoryIeMissing(utils.OriginID)
		}
		args.APIOpts[utils.OptsResourcesUsageID] = originID
		args.APIOpts[utils.OptsResourcesUnits] = 1
		var allocMessage string
		if err = sS.connMgr.Call(ctx, sS.cgrCfg.SessionSCfg().ResSConns, utils.ResourceSv1AllocateResources,
			args, &allocMessage); err != nil {
			return utils.NewErrResourceS(err)
		}
		rply.ResourceAllocation = &allocMessage
	}
	if utils.OptAsBool(args.APIOpts, utils.OptsRouteS) {
		routesReply, err := sS.getRoutes(ctx, args.Clone())
		if err != nil {
			return err
		}
		if routesReply != nil {
			rply.RouteProfiles = routesReply
		}
	}
	if utils.OptAsBool(args.APIOpts, utils.OptsSesMessage) {
		var maxUsage time.Duration
		if maxUsage, err = sS.chargeEvent(ctx, args, utils.OptAsBool(args.APIOpts, utils.OptsSesForceDuration)); err != nil {
			return err
		}
		rply.MaxUsage = &maxUsage
	}
	if utils.OptAsBool(args.APIOpts, utils.OptsThresholdS) {
		var thIDs []string
		if thIDs, err = utils.OptAsStringSlice(args.APIOpts, utils.OptsSesThresholdIDs); err != nil {
			return
		}
		tIDs, err := sS.processThreshold(ctx, args, thIDs, true)
		if err != nil && err.Error() != utils.ErrNotFound.Error() {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> error: %s processing event %+v with ThresholdS.",
					utils.SessionS, err.Error(), args))
			withErrors = true
		}
		rply.ThresholdIDs = &tIDs
	}
	if utils.OptAsBool(args.APIOpts, utils.OptsStatS) {
		var stIDs []string
		if stIDs, err = utils.OptAsStringSlice(args.APIOpts, utils.OptsSesStatIDs); err != nil {
			return
		}
		sIDs, err := sS.processStats(ctx, args, stIDs, false)
		if err != nil &&
			err.Error() != utils.ErrNotFound.Error() {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> error: %s processing event %+v with StatS.",
					utils.SessionS, err.Error(), args))
			withErrors = true
		}
		rply.StatQueueIDs = &sIDs
	}
	if withErrors {
		err = utils.ErrPartiallyExecuted
	}

	return
}

// BiRPCv1ProcessEvent processes one event with the right subsystems based on arguments received
func (sS *SessionS) BiRPCv1ProcessEvent(ctx *context.Context,
	args *utils.CGREvent, rply *V1ProcessEventReply) (err error) {
	if args == nil {
		return utils.NewErrMandatoryIeMissing(utils.CGREventString)
	}
	var withErrors bool
	if args.ID == "" {
		args.ID = utils.GenUUID()
	}
	if args.Tenant == "" {
		args.Tenant = sS.cgrCfg.GeneralCfg().DefaultTenant
	}

	// RPC caching
	if sS.cgrCfg.CacheCfg().Partitions[utils.CacheRPCResponses].Limit != 0 {
		cacheKey := utils.ConcatenatedKey(utils.SessionSv1ProcessEvent, args.ID)
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
		defer engine.Cache.Set(ctx, utils.CacheRPCResponses, cacheKey,
			&utils.CachedRPCResponse{Result: rply, Error: err},
			nil, true, utils.NonTransactional)
	}
	// end of RPC caching

	blockError := utils.OptAsBool(args.APIOpts, utils.OptsSesBlockerError)
	events := map[string]*utils.CGREvent{
		utils.MetaRaw: args,
	}

	if utils.OptAsBool(args.APIOpts, utils.OptsChargerS) {
		var chrgrs []*engine.ChrgSProcessEventReply
		if chrgrs, err = sS.processChargerS(ctx, args); err != nil {
			return
		}
		for _, chrgr := range chrgrs {
			events[utils.IfaceAsString(chrgr.CGREvent.Event[utils.RunID])] = chrgr.CGREvent
		}
	}

	// check for *attribute
	if utils.OptAsBool(args.APIOpts, utils.OptsAttributeS) {
		rply.Attributes = make(map[string]*engine.AttrSProcessEventReply)

		for runID, cgrEv := range getDerivedEvents(events, utils.OptAsBool(args.APIOpts, utils.OptsSesAttributeSDerivedReply)) {
			rplyAttr, err := sS.processAttributes(ctx, cgrEv)
			if err != nil {
				if err.Error() != utils.ErrNotFound.Error() {
					return utils.NewErrAttributeS(err)
				}
			} else {
				*cgrEv = *rplyAttr.CGREvent
				rply.Attributes[runID] = &rplyAttr
			}
		}
		args = events[utils.MetaRaw]
	}

	// get routes if required
	if utils.OptAsBool(args.APIOpts, utils.OptsRouteS) {
		rply.RouteProfiles = make(map[string]engine.SortedRoutesList)
		// check in case we have options for suppliers
		for runID, cgrEv := range getDerivedEvents(events, utils.OptAsBool(args.APIOpts, utils.OptsSesRouteSDerivedReply)) {
			routesReply, err := sS.getRoutes(ctx, cgrEv.Clone())
			if err != nil {
				return err
			}
			if routesReply != nil {
				rply.RouteProfiles[runID] = routesReply
			}
		}
	}

	// process thresholds if required
	if utils.OptAsBool(args.APIOpts, utils.OptsThresholdS) {
		rply.ThresholdIDs = make(map[string][]string)
		var thIDs []string
		if thIDs, err = utils.OptAsStringSlice(args.APIOpts, utils.OptsSesThresholdIDs); err != nil {
			return
		}
		for runID, cgrEv := range getDerivedEvents(events, utils.OptAsBool(args.APIOpts, utils.OptsSesThresholdSDerivedReply)) {
			tIDs, err := sS.processThreshold(ctx, cgrEv, thIDs, true)
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
	if utils.OptAsBool(args.APIOpts, utils.OptsStatS) {
		rply.StatQueueIDs = make(map[string][]string)
		var stIDs []string
		if stIDs, err = utils.OptAsStringSlice(args.APIOpts, utils.OptsSesStatIDs); err != nil {
			return
		}
		for runID, cgrEv := range getDerivedEvents(events, utils.OptAsBool(args.APIOpts, utils.OptsSesStatSDerivedReply)) {
			sIDs, err := sS.processStats(ctx, cgrEv, stIDs, true)
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

	if utils.OptAsBool(args.APIOpts, utils.OptsSesSTIRAuthenticate) {
		for _, cgrEv := range getDerivedEvents(events, utils.OptAsBool(args.APIOpts, utils.OptsSesSTIRDerivedReply)) {
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
			if err = AuthStirShaken(ctx, opts.GetStringIgnoreErrors(utils.OptsStirIdentity),
				utils.FirstNonEmpty(opts.GetStringIgnoreErrors(utils.OptsStirOriginatorTn), ev.GetStringIgnoreErrors(utils.AccountField)),
				opts.GetStringIgnoreErrors(utils.OptsStirOriginatorURI),
				utils.FirstNonEmpty(opts.GetStringIgnoreErrors(utils.OptsStirDestinationTn), ev.GetStringIgnoreErrors(utils.Destination)),
				opts.GetStringIgnoreErrors(utils.OptsStirDestinationURI),
				attest, stirMaxDur); err != nil {
				return utils.NewSTIRError(err.Error())
			}
		}
	} else if utils.OptAsBool(args.APIOpts, utils.OptsSesSTIRInitiate) {
		rply.STIRIdentity = make(map[string]string)
		for runID, cgrEv := range getDerivedEvents(events, utils.OptAsBool(args.APIOpts, utils.OptsSesSTIRDerivedReply)) {
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
			if rply.STIRIdentity[runID], err = NewSTIRIdentity(ctx, header, payload, prvkeyPath, sS.cgrCfg.GeneralCfg().ReplyTimeout); err != nil {
				return utils.NewSTIRError(err.Error())
			}
		}
	}

	// check for *resources
	if opt, has := args.APIOpts[utils.OptsResourceS]; has {
		if len(sS.cgrCfg.SessionSCfg().ResSConns) == 0 {
			return utils.NewErrNotConnected(utils.ResourceS)
		}
		var method string
		// check what we need to do for resources (*authorization/*allocation)
		switch optStr := utils.IfaceAsString(opt); optStr {
		case utils.MetaAuthorize:
			method = utils.ResourceSv1AuthorizeResources
		case utils.MetaAllocate:
			method = utils.ResourceSv1AllocateResources
		case utils.MetaRelease:
			method = utils.ResourceSv1ReleaseResources
		default:
			return fmt.Errorf("unsuported value for %s option: %q ", utils.OptsResourceS, optStr)
		}
		rply.ResourceAllocation = make(map[string]string)
		for runID, cgrEv := range getDerivedEvents(events, utils.OptAsBool(args.APIOpts, utils.OptsSesResourceSDerivedReply)) {
			originID := engine.MapEvent(cgrEv.Event).GetStringIgnoreErrors(utils.OriginID)
			if originID == "" {
				return utils.NewErrMandatoryIeMissing(utils.OriginID)
			}

			cgrEv.APIOpts[utils.OptsResourcesUsageID] = originID
			cgrEv.APIOpts[utils.OptsResourcesUnits] = 1
			var resMessage string
			if err = sS.connMgr.Call(ctx, sS.cgrCfg.SessionSCfg().ResSConns, method,
				cgrEv, &resMessage); err != nil {
				if blockError {
					return utils.NewErrResourceS(err)
				}
				utils.Logger.Warning(
					fmt.Sprintf("<%s> error: <%s> processing event %+v for RunID <%s>  with ResourceS.",
						utils.SessionS, err.Error(), cgrEv, runID))
				withErrors = true
			}
			rply.ResourceAllocation[runID] = resMessage
		}
	}

	// check what we need to do for RALs (*authorize/*initiate/*update/*terminate)
	// dbtItvl := sS.cgrCfg.SessionSCfg().DebitInterval
	// if argsFlagsWithParams.GetBool(utils.MetaRALs) {
	// 	if ralsOpts := argsFlagsWithParams[utils.MetaRALs]; len(ralsOpts) != 0 {
	// 		//check for subflags and convert them into utils.FlagsWithParams
	// 		// check for *cost
	// 		if ralsOpts.Has(utils.MetaCost) {
	// 			rply.Cost = make(map[string]float64)
	// 			for runID, cgrEv := range getDerivedEvents(events, ralsOpts.Has(utils.MetaDerivedReply)) {
	// 				ev := engine.MapEvent(cgrEv.Event)
	// 				//compose the CallDescriptor with Args
	// 				startTime := ev.GetTimeIgnoreErrors(utils.AnswerTime,
	// 					sS.cgrCfg.GeneralCfg().DefaultTimezone)
	// 				if startTime.IsZero() { // AnswerTime not parsable, try SetupTime
	// 					startTime = ev.GetTimeIgnoreErrors(utils.SetupTime,
	// 						sS.cgrCfg.GeneralCfg().DefaultTimezone)
	// 				}
	// 				category := ev.GetStringIgnoreErrors(utils.Category)
	// 				if len(category) == 0 {
	// 					category = sS.cgrCfg.GeneralCfg().DefaultCategory
	// 				}
	// 				subject := ev.GetStringIgnoreErrors(utils.Subject)
	// 				if len(subject) == 0 {
	// 					subject = ev.GetStringIgnoreErrors(utils.AccountField)
	// 				}

	// 				cd := &engine.CallDescriptor{
	// 					CgrID:         cgrEv.ID,
	// 					RunID:         ev.GetStringIgnoreErrors(utils.RunID),
	// 					ToR:           ev.GetStringIgnoreErrors(utils.ToR),
	// 					Tenant:        cgrEv.Tenant,
	// 					Category:      category,
	// 					Subject:       subject,
	// 					Account:       ev.GetStringIgnoreErrors(utils.AccountField),
	// 					Destination:   ev.GetStringIgnoreErrors(utils.Destination),
	// 					TimeStart:     startTime,
	// 					TimeEnd:       startTime.Add(ev.GetDurationIgnoreErrors(utils.Usage)),
	// 					ForceDuration: ralsOpts.Has(utils.MetaFD),
	// 				}
	// 				var cc engine.CallCost
	// 				if err = sS.connMgr.Call(sS.cgrCfg.SessionSCfg().RALsConns, nil,
	// 					utils.ResponderGetCost,
	// 					&engine.CallDescriptorWithAPIOpts{
	// 						CallDescriptor: cd,
	// 						APIOpts:        cgrEv.APIOpts,
	// 					}, &cc); err != nil {
	// 					return err
	// 				}
	// 				rply.Cost[runID] = cc.Cost
	// 			}
	// 		}
	// 		opts := engine.MapEvent(args.APIOpts)
	// 		ev := engine.MapEvent(args.CGREvent.Event)
	// 		originID := ev.GetStringIgnoreErrors(utils.OriginID)
	// 		switch {
	// 		//check for auth session
	// 		case ralsOpts.Has(utils.MetaAuthorize):
	// 			var sRunsMaxUsage map[string]time.Duration
	// 			if sRunsMaxUsage, err = sS.authEvent(args.CGREvent, ralsOpts.Has(utils.MetaFD)); err != nil {
	// 				return err
	// 			}
	// 			rply.MaxUsage = getDerivedMaxUsage(sRunsMaxUsage, ralsOpts.Has(utils.MetaDerivedReply))
	// 		// check for init session
	// 		case ralsOpts.Has(utils.MetaInitiate):
	// 			if opts.HasField(utils.OptsDebitInterval) { // dynamic DebitInterval via CGRDebitInterval
	// 				if dbtItvl, err = opts.GetDuration(utils.OptsDebitInterval); err != nil {
	// 					return err //utils.NewErrRALs(err)
	// 				}
	// 			}
	// 			s, err := sS.initSession(args.CGREvent, sS.biJClntID(clnt), originID, dbtItvl, false,
	// 				ralsOpts.Has(utils.MetaFD))
	// 			if err != nil {
	// 				return err
	// 			}
	// 			sRunsMaxUsage := make(map[string]time.Duration)
	// 			s.RLock()
	// 			isPrepaid := s.debitStop != nil
	// 			s.RUnlock()
	// 			if isPrepaid { //active debit
	// 				for _, sr := range s.SRuns {
	// 					sRunsMaxUsage[sr.CD.RunID] = sS.cgrCfg.SessionSCfg().GetDefaultUsage(ev.GetStringIgnoreErrors(utils.ToR))
	// 				}
	// 			} else if sRunsMaxUsage, err = sS.updateSession(s, nil, args.APIOpts, false); err != nil {
	// 				return err //utils.NewErrRALs(err)
	// 			}
	// 			rply.MaxUsage = getDerivedMaxUsage(sRunsMaxUsage, ralsOpts.Has(utils.MetaDerivedReply))
	// 			//check for update session
	// 		case ralsOpts.Has(utils.MetaUpdate):
	// 			if opts.HasField(utils.OptsDebitInterval) { // dynamic DebitInterval via CGRDebitInterval
	// 				if dbtItvl, err = opts.GetDuration(utils.OptsDebitInterval); err != nil {
	// 					return err //utils.NewErrRALs(err)
	// 				}
	// 			}
	// 			s := sS.getRelocateSession(GetSetCGRID(ev),
	// 				ev.GetStringIgnoreErrors(utils.InitialOriginID),
	// 				ev.GetStringIgnoreErrors(utils.OriginID),
	// 				ev.GetStringIgnoreErrors(utils.OriginHost))
	// 			if s == nil {
	// 				if s, err = sS.initSession(args.CGREvent, sS.biJClntID(clnt), ev.GetStringIgnoreErrors(utils.OriginID),
	// 					dbtItvl, false, ralsOpts.Has(utils.MetaFD)); err != nil {
	// 					return err
	// 				}
	// 			}
	// 			var sRunsMaxUsage map[string]time.Duration
	// 			if sRunsMaxUsage, err = sS.updateSession(s, ev, args.APIOpts, false); err != nil {
	// 				return err //utils.NewErrRALs(err)
	// 			}
	// 			rply.MaxUsage = getDerivedMaxUsage(sRunsMaxUsage, ralsOpts.Has(utils.MetaDerivedReply))
	// 			// check for terminate session
	// 		case ralsOpts.Has(utils.MetaTerminate):
	// 			if opts.HasField(utils.OptsDebitInterval) { // dynamic DebitInterval via CGRDebitInterval
	// 				if dbtItvl, err = opts.GetDuration(utils.OptsDebitInterval); err != nil {
	// 					return err //utils.NewErrRALs(err)
	// 				}
	// 			}
	// 			s := sS.getRelocateSession(GetSetCGRID(ev),
	// 				ev.GetStringIgnoreErrors(utils.InitialOriginID),
	// 				ev.GetStringIgnoreErrors(utils.OriginID),
	// 				ev.GetStringIgnoreErrors(utils.OriginHost))
	// 			if s == nil {
	// 				if s, err = sS.initSession(args.CGREvent, sS.biJClntID(clnt), ev.GetStringIgnoreErrors(utils.OriginID),
	// 					dbtItvl, false, ralsOpts.Has(utils.MetaFD)); err != nil {
	// 					return err
	// 				}
	// 			} else {
	// 				s.Lock()
	// 				s.chargeable = opts.GetBoolOrDefault(utils.OptsChargeable, true)
	// 				s.Unlock()
	// 			}
	// 			if err = sS.terminateSession(s,
	// 				ev.GetDurationPtrIgnoreErrors(utils.Usage),
	// 				ev.GetDurationPtrIgnoreErrors(utils.LastUsed),
	// 				ev.GetTimePtrIgnoreErrors(utils.AnswerTime, utils.EmptyString),
	// 				false); err != nil {
	// 				return err //utils.NewErrRALs(err)
	// 			}
	// 		}
	// 	}
	// }

	if utils.OptAsBool(args.APIOpts, utils.OptsCDRs) {
		if len(sS.cgrCfg.SessionSCfg().CDRsConns) == 0 {
			return utils.NewErrNotConnected(utils.CDRs)
		}
		var cdrRply string
		for _, cgrEv := range getDerivedEvents(events, utils.OptAsBool(args.APIOpts, utils.OptsSesCDRsDerivedReply)) {
			if err := sS.processCDR(ctx, cgrEv, &cdrRply); err != nil {
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

// BiRPCv1SyncSessions will sync sessions on demand
func (sS *SessionS) BiRPCv1SyncSessions(ctx *context.Context,
	ignParam *utils.TenantWithAPIOpts, reply *string) error {
	sS.syncSessions(ctx)
	*reply = utils.OK
	return nil
}

// BiRPCv1ForceDisconnect will force disconnecting sessions matching sessions
func (sS *SessionS) BiRPCv1ForceDisconnect(ctx *context.Context,
	args *utils.SessionFilter, reply *string) (err error) {
	if args == nil { //protection in case on nil
		args = &utils.SessionFilter{}
	}
	if len(args.Filters) != 0 && sS.dm == nil {
		return utils.ErrNoDatabaseConn
	}
	aSs := sS.filterSessions(ctx, args, false)
	if len(aSs) == 0 {
		return utils.ErrNotFound
	}
	for _, as := range aSs {
		ss := sS.getSessions(as.CGRID, false)
		if len(ss) == 0 {
			continue
		}
		ss[0].Lock()
		if errTerm := sS.forceSTerminate(ctx, ss[0], 0, nil, nil); errTerm != nil {
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
func (sS *SessionS) BiRPCv1RegisterInternalBiJSONConn(ctx *context.Context,
	connID string, reply *string) error {
	sS.RegisterIntBiJConn(ctx.Client, connID)
	*reply = utils.OK
	return nil
}

// BiRPCv1ActivateSessions is called to activate a list/all sessions
// returns utils.ErrPartiallyExecuted in case of errors
func (sS *SessionS) BiRPCv1ActivateSessions(ctx *context.Context,
	sIDs *utils.SessionIDsWithAPIOpts, reply *string) (err error) {
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
func (sS *SessionS) BiRPCv1DeactivateSessions(ctx *context.Context,
	sIDs *utils.SessionIDsWithAPIOpts, reply *string) (err error) {
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

func (sS *SessionS) processCDR(ctx *context.Context, cgrEv *utils.CGREvent, rply *string) (err error) {
	ev := engine.MapEvent(cgrEv.Event)
	cgrID := GetSetCGRID(ev)
	s := sS.getRelocateSession(ctx, cgrID,
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
		return sS.connMgr.Call(ctx, sS.cgrCfg.SessionSCfg().CDRsConns, utils.CDRsV1ProcessEvent,
			cgrEv, rply)
	}

	// Use previously stored Session to generate CDRs
	s.updateSRuns(ev, sS.cgrCfg.SessionSCfg().AlterableFields)
	// create one CGREvent for each session run
	var withErrors bool
	for _, cgrEv := range s.asCGREvents() {
		if cgrEv.APIOpts == nil {
			cgrEv.APIOpts = make(map[string]interface{})
		}
		cgrEv.APIOpts[utils.OptsAttributeS] = false
		cgrEv.APIOpts[utils.OptsChargerS] = false
		if mp := engine.MapEvent(cgrEv.Event); unratedReqs.HasField(mp.GetStringIgnoreErrors(utils.RequestType)) { // order additional rating for unrated request types
			// argsProc.Flags = append(argsProc.Flags, fmt.Sprintf("%s:true", utils.MetaRALs))
		}
		if err = sS.connMgr.Call(ctx, sS.cgrCfg.SessionSCfg().CDRsConns, utils.CDRsV1ProcessEvent,
			cgrEv, rply); err != nil {
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
func (sS *SessionS) processThreshold(ctx *context.Context, cgrEv *utils.CGREvent, thIDs []string, clnb bool) (tIDs []string, err error) {
	if len(sS.cgrCfg.SessionSCfg().ThreshSConns) == 0 {
		return tIDs, utils.NewErrNotConnected(utils.ThresholdS)
	}
	// check if we have thresholdIDs
	if len(thIDs) != 0 {
		cgrEv.APIOpts[utils.OptsThresholdsThresholdIDs] = thIDs
	}
	cgrEv.SetCloneable(clnb)
	//initialize the returned variable
	err = sS.connMgr.Call(ctx, sS.cgrCfg.SessionSCfg().ThreshSConns, utils.ThresholdSv1ProcessEvent, cgrEv, &tIDs)
	return
}

// processStats will receive the event and send it to StatS to be processed
func (sS *SessionS) processStats(ctx *context.Context, cgrEv *utils.CGREvent, stsIDs []string, clnb bool) (sIDs []string, err error) {
	if len(sS.cgrCfg.SessionSCfg().StatSConns) == 0 {
		return sIDs, utils.NewErrNotConnected(utils.StatS)
	}

	statArgs := &engine.StatsArgsProcessEvent{
		CGREvent: cgrEv,
	}
	// check in case we have StatIDs inside flags
	if len(stsIDs) != 0 {
		cgrEv.APIOpts[utils.OptsStatsStatIDs] = stsIDs
	}
	statArgs.SetCloneable(clnb)
	//initialize the returned variable
	err = sS.connMgr.Call(ctx, sS.cgrCfg.SessionSCfg().StatSConns, utils.StatSv1ProcessEvent, statArgs, &sIDs)
	return
}

// getRoutes will receive the event and send it to SupplierS to find the suppliers
func (sS *SessionS) getRoutes(ctx *context.Context, cgrEv *utils.CGREvent) (routesReply engine.SortedRoutesList, err error) {
	if len(sS.cgrCfg.SessionSCfg().RouteSConns) == 0 {
		return routesReply, utils.NewErrNotConnected(utils.RouteS)
	}
	if acd, has := cgrEv.Event[utils.ACD]; has {
		cgrEv.Event[utils.Usage] = acd
	}

	if err = sS.connMgr.Call(ctx, sS.cgrCfg.SessionSCfg().RouteSConns, utils.RouteSv1GetRoutes,
		cgrEv, &routesReply); err != nil {
		return routesReply, utils.NewErrRouteS(err)
	}
	return
}

// processAttributes will receive the event and send it to AttributeS to be processed
func (sS *SessionS) processAttributes(ctx *context.Context, cgrEv *utils.CGREvent) (rplyEv engine.AttrSProcessEventReply, err error) {
	if len(sS.cgrCfg.SessionSCfg().AttrSConns) == 0 {
		return rplyEv, utils.NewErrNotConnected(utils.AttributeS)
	}
	if cgrEv.APIOpts == nil {
		cgrEv.APIOpts = make(engine.MapEvent)
	}
	cgrEv.APIOpts[utils.Subsys] = utils.MetaSessionS
	cgrEv.APIOpts[utils.OptsContext] = utils.FirstNonEmpty(
		utils.IfaceAsString(cgrEv.APIOpts[utils.OptsContext]),
		utils.MetaSessionS)
	err = sS.connMgr.Call(ctx, sS.cgrCfg.SessionSCfg().AttrSConns, utils.AttributeSv1ProcessEvent,
		cgrEv, &rplyEv)
	return
}

func (sS *SessionS) sendRar(ctx *context.Context, s *Session) (err error) {
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
	if err = clnt.conn.Call(ctx, utils.SessionSv1ReAuthorize, originID, &rply); err == utils.ErrNotImplemented {
		err = nil
	}
	return
}

// BiRPCv1ReAuthorize sends a RAR for the matching sessions
func (sS *SessionS) BiRPCv1ReAuthorize(ctx *context.Context,
	args *utils.SessionFilter, reply *string) (err error) {
	if args == nil { //protection in case on nil
		args = &utils.SessionFilter{}
	}
	aSs := sS.filterSessions(ctx, args, false)
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
		if errTerm := sS.sendRar(ctx, ss[0]); errTerm != nil {
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
		if err = clnt.conn.Call(ctx, utils.SessionSv1DisconnectPeer, args, reply); err != nil && err != utils.ErrNotImplemented {
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
	if err = AuthStirShaken(ctx, args.Identity, args.OriginatorTn, args.OriginatorURI,
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
		ctx,
		utils.NewPASSporTHeader(utils.FirstNonEmpty(args.PublicKeyPath,
			sS.cgrCfg.SessionSCfg().STIRCfg.PublicKeyPath)),
		args.Payload, utils.FirstNonEmpty(args.PrivateKeyPath,
			sS.cgrCfg.SessionSCfg().STIRCfg.PrivateKeyPath),
		sS.cgrCfg.GeneralCfg().ReplyTimeout); err != nil {
		return utils.NewSTIRError(err.Error())
	}
	return
}
