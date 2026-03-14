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
	"strings"
	"sync"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/attributes"
	"github.com/cgrates/cgrates/chargers"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/routes"

	"github.com/cgrates/cgrates/utils"
)

var (
	// ErrForcedDisconnect is used to specify the reason why the session was disconnected
	ErrForcedDisconnect = errors.New("FORCED_DISCONNECT")
)

// NewSessionS constructs  a new SessionS instance
func NewSessionS(cgrCfg *config.CGRConfig, dm *engine.DataManager, fltrS *engine.FilterS,
	connMgr *engine.ConnManager) *SessionS {
	cgrCfg.SessionSCfg().SessionIndexes.Add(utils.OriginID) // Make sure we have indexing for OriginID since it is a requirement on prefix searching

	return &SessionS{
		cfg:           cgrCfg,
		dm:            dm,
		fltrS:         fltrS,
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
	cfg     *config.CGRConfig // Separate from smCfg since there can be multiple
	dm      *engine.DataManager
	fltrS   *engine.FilterS
	connMgr *engine.ConnManager

	biJMux   sync.RWMutex                     // mux protecting BI-JSON connections
	biJClnts map[birpc.ClientConnector]string // index BiJSONConnection so we can sync them later
	biJIDs   map[string]*biJClient            // identifiers of bidirectional JSON conns, used to call RPC based on connIDs

	aSsMux    sync.RWMutex        // protects aSessions
	aSessions map[string]*Session // group sessions per sessionId

	aSIMux        sync.RWMutex                                     // protects aSessionsIdx
	aSessionsIdx  map[string]map[string]map[string]utils.StringSet // map[fieldName]map[fieldValue][originID]utils.StringSet[runID]sID
	aSessionsRIdx map[string][]*riFieldNameVal                     // reverse indexes for active sessions, used on remove

	pSsMux    sync.RWMutex        // protects pSessions
	pSessions map[string]*Session // group passive sessions based on originID

	pSIMux        sync.RWMutex                                     // protects pSessionsIdx
	pSessionsIdx  map[string]map[string]map[string]utils.StringSet // map[fieldName]map[fieldValue][originID]utils.StringSet[runID]sID
	pSessionsRIdx map[string][]*riFieldNameVal                     // reverse indexes for passive sessions, used on remove
}

// ListenAndServe starts the service and binds it to the listen loop
func (sS *SessionS) ListenAndServe(stopChan chan struct{}) {
	if sS.cfg.SessionSCfg().ChannelSyncInterval != 0 {
		for { // Schedule sync channels to run repeately
			select {
			case <-stopChan:
				return
			case <-time.After(sS.cfg.SessionSCfg().ChannelSyncInterval):
				sS.syncSessions(context.TODO())
			}
		}
	}
}

// Shutdown is called by engine to clear states
func (sS *SessionS) Shutdown() (err error) {
	replConns, err := engine.GetConnIDs(context.TODO(), sS.cfg.SessionSCfg().Conns[utils.MetaReplication], utils.MetaAny, utils.MapStorage{}, sS.fltrS)
	if len(replConns) == 0 {
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
		proto: sS.cfg.SessionSCfg().ClientProtocol}
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
		nodeID = sS.cfg.GeneralCfg().NodeID
	}
	sS.biJMux.Lock()
	sS.biJClnts[c] = nodeID
	sS.biJIDs[nodeID] = &biJClient{
		conn:  c,
		proto: sS.cfg.SessionSCfg().ClientProtocol}
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
func (sS *SessionS) setSTerminator(ctx *context.Context, s *Session, opts engine.MapEvent) {
	var err error

	// TTL
	var ttl time.Duration
	if ttl, err = engine.GetDurationOptsFromMultipleMaps(ctx, s.OriginCGREvent.Tenant, s.OriginCGREvent.Event, opts, s.OriginCGREvent.APIOpts, sS.fltrS,
		sS.cfg.SessionSCfg().Opts.TTL, config.SessionsTTLDftOpt, utils.OptsSesTTL); err != nil {
		utils.Logger.Warning(
			fmt.Sprintf("<%s>, cannot extract <%s> for session:<%s>, from its options: <%s>, err: <%s>",
				utils.SessionS, utils.OptsSesTTL, s.OriginCGREvent.APIOpts[utils.MetaOriginID], opts, err))
		return
	}
	if ttl == 0 {
		return // nothing to set up
	}
	// random delay computation
	var maxDelay time.Duration
	if maxDelay, err = engine.GetDurationOptsFromMultipleMaps(ctx, s.OriginCGREvent.Tenant, s.OriginCGREvent.Event, opts, s.OriginCGREvent.APIOpts, sS.fltrS,
		sS.cfg.SessionSCfg().Opts.TTLMaxDelay, config.SessionsTTLMaxDelayDftOpt, utils.OptsSesTTLMaxDelay); err != nil {
		utils.Logger.Warning(
			fmt.Sprintf("<%s>, cannot extract <%s> for session:<%s>, from its options: <%s>, err: <%s>",
				utils.SessionS, utils.OptsSesTTLMaxDelay, s.OriginCGREvent.APIOpts[utils.MetaOriginID], opts.String(), err.Error()))
		return
	}
	if maxDelay != 0 {
		ttl += time.Duration(
			rand.Int63n(maxDelay.Milliseconds()) * time.Millisecond.Nanoseconds())
	}
	// LastUsed
	var ttlLastUsed *time.Duration
	if ttlLastUsed, err = engine.GetDurationPointerOptsFromMultipleMaps(ctx, s.OriginCGREvent.Tenant, s.OriginCGREvent.Event, opts, s.OriginCGREvent.APIOpts, sS.fltrS,
		sS.cfg.SessionSCfg().Opts.TTLLastUsed, utils.OptsSesTTLLastUsed); err != nil {
		utils.Logger.Warning(
			fmt.Sprintf("<%s>, cannot extract <%s> for session:<%s>, from its options: <%s>, err: <%s>",
				utils.SessionS, utils.OptsSesTTLLastUsed, s.OriginCGREvent.APIOpts[utils.MetaOriginID], opts.String(), err.Error()))
		return
	}
	// LastUsage
	var ttlLastUsage *time.Duration
	if ttlLastUsage, err = engine.GetDurationPointerOptsFromMultipleMaps(ctx, s.OriginCGREvent.Tenant, s.OriginCGREvent.Event, opts, s.OriginCGREvent.APIOpts, sS.fltrS,
		sS.cfg.SessionSCfg().Opts.TTLLastUsage, utils.OptsSesTTLLastUsage); err != nil {
		utils.Logger.Warning(
			fmt.Sprintf("<%s>, cannot extract <%s> for session:<%s>, from its options: <%s>, err: <%s>",
				utils.SessionS, utils.OptsSesTTLLastUsage, s.OriginCGREvent.APIOpts[utils.MetaOriginID], opts.String(), err.Error()))
		return
	}
	// TTLUsage
	var ttlUsage *time.Duration
	if ttlUsage, err = engine.GetDurationPointerOptsFromMultipleMaps(ctx, s.OriginCGREvent.Tenant, s.OriginCGREvent.Event, opts, s.OriginCGREvent.APIOpts, sS.fltrS,
		sS.cfg.SessionSCfg().Opts.TTLUsage, utils.OptsSesTTLUsage); err != nil {
		utils.Logger.Warning(
			fmt.Sprintf("<%s>, cannot extract <%s> for session:<%s>, from its options: <%s>, err: <%s>",
				utils.SessionS, utils.OptsSesTTLUsage, s.OriginCGREvent.APIOpts[utils.MetaOriginID], opts.String(), err.Error()))
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
			s.lk.Lock()
			lastUsage := s.sTerminator.ttl
			if s.sTerminator.ttlLastUsage != nil {
				lastUsage = *s.sTerminator.ttlLastUsage
			}
			sS.forceSTerminate(ctx, s, lastUsage, s.sTerminator.ttlUsage,
				s.sTerminator.ttlLastUsed)
			s.lk.Unlock()
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
						"<%s> failed debitting originID %s, sRunIdx: %d, err: %s",
						utils.SessionS, s.ID, i, err.Error()))
			}
		}
	}
	// we apply the correction before
	if err = sS.endSession(ctx, s, tUsage, lastUsed, nil, false); err != nil {
		utils.Logger.Warning(
			fmt.Sprintf(
				"<%s> failed force terminating session with ID <%s>, err: <%s>",
				utils.SessionS, s.ID, err.Error()))
	}
	tenant := s.OriginCGREvent.Tenant
	dP := s.OriginCGREvent.AsDataProvider()
	// post the CDRs
	if cdrsConns, errC := engine.GetConnIDs(ctx, sS.cfg.SessionSCfg().Conns[utils.MetaCDRs], tenant, dP, sS.fltrS); errC != nil {
		utils.Logger.Warning(
			fmt.Sprintf("<%s> error resolving CDRs connections: %s", utils.SessionS, errC.Error()))
	} else if len(cdrsConns) != 0 {
		var reply string
		for _, cgrEv := range s.asCGREvents() {
			if cgrEv.APIOpts == nil {
				cgrEv.APIOpts = make(map[string]any)
			}
			cgrEv.APIOpts[utils.MetaAttributes] = false
			cgrEv.APIOpts[utils.MetaChargers] = false
			if unratedReqs.HasField( // order additional rating for unrated request types
				engine.MapEvent(cgrEv.Event).GetStringIgnoreErrors(utils.RequestType)) {
				// argsProc.Flags = append(argsProc.Flags, utils.MetaRALs)
			}
			cgrEv.SetCloneable(true)
			if err = sS.connMgr.Call(ctx, cdrsConns,
				utils.CDRsV1ProcessEvent, cgrEv, &reply); err != nil {
				utils.Logger.Warning(
					fmt.Sprintf(
						"<%s> could not post CDR for event %s, err: %s",
						utils.SessionS, utils.ToJSON(cgrEv), err.Error()))
			}
		}
	}
	// release the resources for the session
	if resSConns, errR := engine.GetConnIDs(ctx, sS.cfg.SessionSCfg().Conns[utils.MetaResources], tenant, dP, sS.fltrS); errR != nil {
		utils.Logger.Warning(
			fmt.Sprintf("<%s> error resolving ResourceS connections: %s", utils.SessionS, errR.Error()))
	} else if len(resSConns) != 0 {
		var reply string
		args := s.OriginCGREvent.Clone()
		args.ID = utils.UUIDSha1Prefix()
		args.APIOpts[utils.OptsResourcesUsageID] = s.ID
		args.APIOpts[utils.OptsResourcesUnits] = 1
		if err := sS.connMgr.Call(ctx, resSConns,
			utils.ResourceSv1ReleaseResources,
			args, &reply); err != nil {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> error: %s could not release resource with resourceID: %s",
					utils.SessionS, err.Error(), s.ID))
		}
	}
	// release the ips for the session
	if ipsConns, errI := engine.GetConnIDs(ctx, sS.cfg.SessionSCfg().Conns[utils.MetaIPs], tenant, dP, sS.fltrS); errI != nil {
		utils.Logger.Warning(
			fmt.Sprintf("<%s> error resolving IPs connections: %s", utils.SessionS, errI.Error()))
	} else if len(ipsConns) != 0 {
		var reply string
		args := s.OriginCGREvent.Clone()
		args.ID = utils.UUIDSha1Prefix()
		args.APIOpts[utils.OptsIPsAllocationID] = s.ID
		if err := sS.connMgr.Call(ctx, ipsConns,
			utils.IPsV1ReleaseIP,
			args, &reply); err != nil {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> error: %s could not release IP %q",
					utils.SessionS, err.Error(), s.ID))
		}
	}
	replConns, err := engine.GetConnIDs(ctx, sS.cfg.SessionSCfg().Conns[utils.MetaReplication], utils.MetaAny, utils.MapStorage{}, sS.fltrS)
	if err != nil {
		return
	}
	sS.replicateSessions(ctx, s.ID, false, replConns)
	if clnt := sS.biJClnt(s.ClientConnID); clnt != nil {
		go func() {
			var rply string
			if err := clnt.conn.Call(ctx,
				utils.AgentV1DisconnectSession,
				utils.CGREvent{
					ID: utils.GenUUID(),
					// EventStart: s.OriginCGREvent.Event,
					// Reason:     ErrForcedDisconnect.Error(),
				},
				&rply); err != nil && err != utils.ErrNotImplemented {
				utils.Logger.Warning(
					fmt.Sprintf("<%s> err: %s remotely disconnect session with id: %s",
						utils.SessionS, err.Error(), s.ID))
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
	/*
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
						utils.SessionS, s.originID(), err.Error()))
				dscReason := utils.ErrServerError.Error()
				if err.Error() == utils.ErrUnauthorizedDestination.Error() {
					dscReason = err.Error()
				}
				// try to disconect the session n times before we force terminate it on our side
				fib := utils.FibDuration(time.Millisecond, 0)
				for i := 0; i < sS.cfg.SessionSCfg().TerminateAttempts; i++ {
					if i != 0 { // not the first iteration
						time.Sleep(fib())
					}
					if err = sS.disconnectSession(s, dscReason); err == nil {
						s.Unlock()
						return
					}
					utils.Logger.Warning(
						fmt.Sprintf("<%s> could not disconnect session: %s, error: %s",
							utils.SessionS, s.originID(), err.Error()))
				}
				if err = sS.forceSTerminate(context.TODO(), s, 0, nil, nil); err != nil {
					utils.Logger.Warning(fmt.Sprintf("<%s> failed force-terminating session: <%s>, err: <%s>", utils.SessionS, s.originID(), err))
				}
				s.Unlock()
				return
			}
			debitStop := s.debitStop // avoid concurrency with endSession
			s.SRuns[sRunIdx].NextAutoDebit = utils.TimePointer(time.Now().Add(dbtIvl))
			if maxDebit < dbtIvl && sS.cfg.SessionSCfg().MinDurLowBalance != time.Duration(0) { // warn client for low balance
				if sS.cfg.SessionSCfg().MinDurLowBalance >= dbtIvl {
					utils.Logger.Warning(fmt.Sprintf("<%s> can not run warning for the session: <%s> since the remaining time:<%s> is higher than the debit interval:<%s>.",
						utils.SessionS, s.originID(), sS.cfg.SessionSCfg().MinDurLowBalance, dbtIvl))
				} else if maxDebit <= sS.cfg.SessionSCfg().MinDurLowBalance {
					go sS.warnSession(s.ClientConnID, s.EventStart.Clone(), s.OptsStart.Clone())
				}
			}
			s.Unlock()
			sS.replicateSessions(context.TODO(), utils.IfaceAsString(s.OptsStart[utils.MetaOriginID]), false, sS.cfg.SessionSCfg().ReplicationConns)
			if maxDebit < dbtIvl { // disconnect faster
				select {
				case <-debitStop: // call was disconnected already
					return
				case <-time.After(maxDebit):
					s.Lock()
					defer s.Unlock()
					// try to disconect the session n times before we force terminate it on our side
					fib := utils.FibDuration(time.Millisecond, 0)
					for i := 0; i < sS.cfg.SessionSCfg().TerminateAttempts; i++ {
						if i != 0 { // not the first iteration
							time.Sleep(fib())
						}
						if err = sS.disconnectSession(s, utils.ErrInsufficientCredit.Error()); err == nil {
							return
						}
						utils.Logger.Warning(
							fmt.Sprintf("<%s> could not disconnect session: %s, error: %s",
								utils.SessionS, s.originID(), err.Error()))
					}
					utils.Logger.Warning(
						fmt.Sprintf("<%s> could not disconnect session: <%s>, error: <%s>",
							utils.SessionS, s.originID(), err.Error()))
					if err = sS.forceSTerminate(context.TODO(), s, 0, nil, nil); err != nil {
						utils.Logger.Warning(fmt.Sprintf("<%s> failed force-terminating session: <%s>, err: <%s>",
							utils.SessionS, s.originID(), err))
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
	*/
	return
}

// disconnectSession will send disconnect from SessionS to clients
// not thread safe, it considers that the session is already stopped by this time
func (sS *SessionS) disconnectSession(s *Session, rsn string) (err error) {
	/*
		clnt := sS.biJClnt(s.ClientConnID)
		if clnt == nil {
			return fmt.Errorf("calling %s requires bidirectional JSON connection, connID: <%s>",
				utils.SessionSv1DisconnectSession, s.ClientConnID)
		}
		s.CGR[utils.Usage] = s.totalUsage() // Set the usage to total one debitted
		servMethod := utils.SessionSv1DisconnectSession
		if clnt.proto == 0 { // compatibility with OpenSIPS 2.3
			servMethod = "SMGClientV1.DisconnectSession"
		}
		var rply string
		if err = clnt.conn.Call(context.TODO(), servMethod,
			utils.AttrDisconnectSession{
				EventStart: s.OriginCGREvent.Event,
				Reason:     rsn}, &rply); err != nil {
			if err != utils.ErrNotImplemented {
				return err
			}
			err = nil
		}
	*/
	return
}

// replicateSessions will replicate sessions with or without originID specified
func (sS *SessionS) replicateSessions(ctx *context.Context, originID string, psv bool, connIDs []string) {
	if len(connIDs) == 0 {
		return
	}
	ss := sS.getSessions(originID, psv)
	if len(ss) == 0 {
		// emulate the Session, so it can be also removed from remote
		ss = []*Session{{ID: originID}}
	}
	for _, s := range ss {
		sCln := s.Clone()
		var rply string
		if err := sS.connMgr.Call(ctx, connIDs,
			utils.SessionSv1SetPassiveSession,
			sCln, &rply); err != nil {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> cannot replicate session with id <%s>, err: %s",
					utils.SessionS, sCln.ID, err.Error()))
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
	sMp[s.ID] = s
	sMux.Unlock()
	sS.indexSession(s, passive)
}

// isSessionRegistered returns if the session is present within the active/passive sessions map
func (sS *SessionS) isSessionRegistered(sID string, passive bool) (has bool) {
	sMux := &sS.aSsMux
	sMp := sS.aSessions
	if passive {
		sMux = &sS.pSsMux
		sMp = sS.pSessions
	}
	sMux.Lock()
	_, has = sMp[utils.IfaceAsString(sID)]
	sMux.Unlock()
	return
}

// uregisterSession will unregister an active or passive session based on it's originID
// called on session terminate or relocate
func (sS *SessionS) unregisterSession(originID string, passive bool) bool {
	sMux := &sS.aSsMux
	sMp := sS.aSessions
	if passive {
		sMux = &sS.pSsMux
		sMp = sS.pSessions
	}
	sMux.Lock()
	if _, has := sMp[originID]; !has {
		sMux.Unlock()
		return false
	}
	delete(sMp, originID)
	sMux.Unlock()
	sS.unindexSession(originID, passive)
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
	for fieldName := range sS.cfg.SessionSCfg().SessionIndexes {
		for _, sr := range s.SRuns {
			fieldVal, err := sr.CGREvent.FieldAsString(fieldName) // the only error from GetString is ErrNotFound
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
			if _, hasID := ssIndx[fieldName][fieldVal][s.ID]; !hasID {
				ssIndx[fieldName][fieldVal][s.ID] = make(utils.StringSet) // we index runs under session id here
			}
			ssIndx[fieldName][fieldVal][s.ID].Add(sr.ID)

			// reverse index
			if _, hasIt := ssRIdx[s.ID]; !hasIt {
				ssRIdx[s.ID] = make([]*riFieldNameVal, 0)
			}
			ssRIdx[s.ID] = append(ssRIdx[s.ID], &riFieldNameVal{fieldName: fieldName, fieldValue: fieldVal})
		}
	}
}

// unindexASession removes an active or passive session from indexes
// called on terminate or relocate
func (sS *SessionS) unindexSession(originID string, pSessions bool) bool {
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
	if _, hasIt := ssRIdx[originID]; !hasIt {
		return false
	}
	for _, riFNV := range ssRIdx[originID] {
		delete(ssIndx[riFNV.fieldName][riFNV.fieldValue], originID)
		if len(ssIndx[riFNV.fieldName][riFNV.fieldValue]) == 0 {
			delete(ssIndx[riFNV.fieldName], riFNV.fieldValue)
		}
		if len(ssIndx[riFNV.fieldName]) == 0 {
			delete(ssIndx, riFNV.fieldName)
		}
	}
	delete(ssRIdx, originID)
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
				!sS.cfg.SessionSCfg().SessionIndexes.Has(fldName) {
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
			for originID, runIDs := range ssIndx[fltrName][fltrVal] {
				if _, hasoriginID := matchingSessionsbyValue[originID]; !hasoriginID {
					matchingSessionsbyValue[originID] = utils.StringSet{}
				}
				for runID := range runIDs {
					matchingSessionsbyValue[originID].Add(runID)
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
		for originID, runIDs := range matchingSessions {
			if matchedRunIDs, hasoriginID := matchedIndx[originID]; !hasoriginID {
				delete(matchingSessions, originID)
				continue
			} else {
				for runID := range runIDs {
					if !matchedRunIDs.Has(runID) {
						delete(matchingSessions[originID], runID)
					}
				}
			}
		}
		if len(matchingSessions) == 0 {
			return make([]string, 0), make(map[string]utils.StringSet)
		}
	}
	originIDs := []string{}
	for originID := range matchingSessions {
		originIDs = append(originIDs, originID)

	}
	return originIDs, matchingSessions
}

// filterSessions will return a list of sessions in external format based on filters passed
// is thread safe for the Sessions
func (sS *SessionS) filterSessions(ctx *context.Context, sf *utils.SessionFilter, psv bool) (aSs []*ExternalSession) {
	if len(sf.Filters) == 0 {
		ss := sS.getSessions(utils.EmptyString, psv)
		for _, s := range ss {
			aSs = append(aSs,
				s.AsExternalSessions(sS.cfg.GeneralCfg().DefaultTimezone,
					sS.cfg.GeneralCfg().NodeID)...) // Expensive for large number of sessions
			if sf.Limit != nil && *sf.Limit > 0 && *sf.Limit < len(aSs) {
				return aSs[:*sf.Limit]
			}
		}
		return
	}
	tenant := utils.FirstNonEmpty(sf.Tenant, sS.cfg.GeneralCfg().DefaultTenant)
	indx, unindx := sS.getIndexedFilters(ctx, tenant, sf.Filters)
	originIDs, _ /*matchingSRuns*/ := sS.getSessionIDsMatchingIndexes(indx, psv)
	if len(indx) != 0 && len(originIDs) == 0 { // no sessions matched the indexed filters
		return
	}
	ss := sS.getSessionsFromOriginIDs(psv, originIDs...)
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
		s.lk.RLock()
		// runIDs := matchingSRuns[s.OptsStart[utils.MetaOriginID]]
		for i, sr := range s.SRuns {
			// if len(originIDs) != 0 && !runIDs.Has(sr.CD.RunID) {
			// continue
			// }
			if pass(unindx, sr.CGREvent.Event) {
				aSs = append(aSs,
					s.AsExternalSession(i, sS.cfg.GeneralCfg().NodeID)) // Expensive for large number of sessions
				if sf.Limit != nil && *sf.Limit > 0 && *sf.Limit < len(aSs) {
					s.lk.RUnlock()
					return aSs[:*sf.Limit]
				}
			}
		}
		s.lk.RUnlock()
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
	tenant := utils.FirstNonEmpty(sf.Tenant, sS.cfg.GeneralCfg().DefaultTenant)
	indx, unindx := sS.getIndexedFilters(ctx, tenant, sf.Filters)
	originIDs, _ /* matchingSRuns*/ := sS.getSessionIDsMatchingIndexes(indx, psv)
	if len(indx) != 0 && len(originIDs) == 0 { // no sessions matched the indexed filters
		return
	}
	ss := sS.getSessionsFromOriginIDs(psv, originIDs...)
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
		s.lk.RLock()
		// runIDs := matchingSRuns[s.OptsStart[utils.MetaOriginID]]
		for _, sr := range s.SRuns {
			// if len(originIDs) != 0 && !runIDs.Has(sr.CD.RunID) {
			// continue
			// }
			if pass(unindx, sr.CGREvent.Event) {
				count++
			}
		}
		s.lk.RUnlock()
	}
	return
}

// newSession will populate SRuns within a Session based on ChargerS output
// forSession can only be called once per Session
// not thread-safe since it should be called in init where there is no concurrency
func (sS *SessionS) newSession(ctx *context.Context, cgrEv *utils.CGREvent,
	clntConnID string) (s *Session, err error) {
	s = &Session{
		ID:             utils.IfaceAsString(cgrEv.APIOpts[utils.MetaOriginID]),
		OriginCGREvent: cgrEv,
		ClientConnID:   clntConnID,
	}

	var chrgS bool
	if chrgS, err = engine.GetBoolOpts(ctx, cgrEv.Tenant, cgrEv.AsDataProvider(), nil,
		sS.fltrS, sS.cfg.SessionSCfg().Opts.Chargers,
		utils.MetaChargers); err != nil {
		return
	}
	if chrgS {
		var chrgrs []*chargers.ChrgSProcessEventReply
		if chrgrs, err = sS.processChargerS(ctx, cgrEv); err != nil {
			return
		}
		s.SRuns = make([]*SRun, len(chrgrs))
		for i, chrgr := range chrgrs {
			s.SRuns[i] = NewSRun(chrgr.CGREvent)
		}
	}

	return
}

// processChargerS processes the event with chargers and caches the response based on the requestID
func (sS *SessionS) processChargerS(ctx *context.Context, cgrEv *utils.CGREvent) (chrgrs []*chargers.ChrgSProcessEventReply, err error) {
	var conns []string
	if conns, err = engine.GetConnIDs(ctx, sS.cfg.SessionSCfg().Conns[utils.MetaChargers], cgrEv.Tenant, cgrEv.AsDataProvider(), sS.fltrS); err != nil {
		return
	}
	if len(conns) == 0 {
		err = errors.New("ChargerS is disabled")
		return
	}
	if x, ok := engine.Cache.Get(utils.CacheEventCharges, cgrEv.ID); ok && x != nil {
		return x.([]*chargers.ChrgSProcessEventReply), nil
	}
	if err = sS.connMgr.Call(ctx, conns,
		utils.ChargerSv1ProcessEvent, cgrEv, &chrgrs); err != nil {
		err = utils.NewErrChargerS(err)
	}

	if errCh := engine.Cache.Set(ctx, utils.CacheEventCharges, cgrEv.ID, chrgrs, nil,
		true, utils.NonTransactional); errCh != nil {
		return nil, errCh
	}
	return
}

// ipsAuthorize will authorize the event with the IPs subsystem
func (sS *SessionS) ipsAuthorize(ctx *context.Context, cgrEv *utils.CGREvent) (rply *utils.AllocatedIP, err error) {
	var conns []string
	if conns, err = engine.GetConnIDs(ctx, sS.cfg.SessionSCfg().Conns[utils.MetaIPs], cgrEv.Tenant, cgrEv.AsDataProvider(), sS.fltrS); err != nil {
		return
	}
	if len(conns) == 0 {
		return nil, utils.NewErrNotConnected(utils.IPs)
	}
	var alcIP utils.AllocatedIP
	if err = sS.connMgr.Call(ctx, conns,
		utils.IPsV1AuthorizeIP,
		cgrEv, &alcIP); err != nil {
		utils.Logger.Warning(
			fmt.Sprintf("<%s> error: %s could not authorize IP for event: %+v",
				utils.SessionS, err.Error(), cgrEv))
	}
	return &alcIP, nil
}

// resourcesAuthorize will authorize the event with the Resources subsystem
func (sS *SessionS) resourcesAuthorize(ctx *context.Context, cgrEv *utils.CGREvent) (resID string, err error) {
	var conns []string
	if conns, err = engine.GetConnIDs(ctx, sS.cfg.SessionSCfg().Conns[utils.MetaResources], cgrEv.Tenant, cgrEv.AsDataProvider(), sS.fltrS); err != nil {
		return
	}
	if len(conns) == 0 {
		return utils.EmptyString, utils.NewErrNotConnected(utils.ResourceS)
	}
	var resUsageID string
	if resUsageID, err = engine.GetStringOpts(ctx, cgrEv.Tenant, cgrEv.AsDataProvider(),
		nil, sS.fltrS, sS.cfg.SessionSCfg().Opts.ResourcesUsageID,
		utils.OptsResourcesUsageID); err != nil {
		return
	}
	cgrEv.APIOpts[utils.OptsResourcesUsageID] = resUsageID
	var resUnits int
	if resUnits, err = engine.GetIntOpts(ctx, cgrEv.Tenant, cgrEv.AsDataProvider(),
		nil, sS.fltrS, sS.cfg.SessionSCfg().Opts.ResourcesUnits,
		utils.OptsResourcesUnits); err != nil {
		return
	}
	cgrEv.APIOpts[utils.OptsResourcesUnits] = resUnits
	var resMessage string
	if err = sS.connMgr.Call(ctx, conns,
		utils.ResourceSv1AuthorizeResources,
		cgrEv, &resMessage); err != nil {
	}
	return resMessage, nil
}

// accountsMaxAbstracts will query the AccountS cost for Event
func (sS *SessionS) accountsMaxAbstracts(ctx *context.Context, cgrEv *utils.CGREvent) (rply *utils.EventCharges, err error) {
	var conns []string
	if conns, err = engine.GetConnIDs(ctx, sS.cfg.SessionSCfg().Conns[utils.MetaAccounts],
		cgrEv.Tenant, cgrEv.AsDataProvider(), sS.fltrS); err != nil {
		return
	}
	if len(conns) == 0 {
		err = utils.NewErrNotConnected(utils.AccountS)
		return
	}
	var acntCost utils.EventCharges
	if err = sS.connMgr.Call(ctx, conns,
		utils.AccountSv1MaxAbstracts, cgrEv, &acntCost); err != nil {
		return
	}
	return &acntCost, nil
}

// accountSDebitEvent will debit the abstracts for the provided event
func (sS *SessionS) accountSDebitEvent(ctx *context.Context, cgrEv *utils.CGREvent) (eEc *utils.EventCharges, err error) {
	var conns []string
	if conns, err = engine.GetConnIDs(ctx, sS.cfg.SessionSCfg().Conns[utils.MetaAccounts],
		cgrEv.Tenant, cgrEv.AsDataProvider(), sS.fltrS); err != nil {
		return
	}
	if len(conns) == 0 {
		err = utils.NewErrNotConnected(utils.AccountS)
		return
	}
	var reply utils.EventCharges
	if err = sS.connMgr.Call(ctx, conns,
		utils.AccountSv1DebitAbstracts, cgrEv, &reply); err != nil {
		return
	}
	return &reply, nil
}

// ratesCost will query the RateS cost for Event
func (sS *SessionS) ratesCost(ctx *context.Context, cgrEv *utils.CGREvent) (cost *utils.Decimal, err error) {
	var rateConns []string
	rateConns, err = engine.GetConnIDs(ctx, sS.cfg.SessionSCfg().Conns[utils.MetaRates], cgrEv.Tenant, cgrEv.AsDataProvider(), sS.fltrS)
	if err != nil {
		return
	}
	if len(rateConns) == 0 {
		err = errors.New("RateS is disabled")
		return
	}
	var rtsCost utils.RateProfileCost
	if err = sS.connMgr.Call(ctx, rateConns,
		utils.RateSv1CostForEvent, cgrEv, &rtsCost); err != nil {
		return
	}
	return rtsCost.Cost, nil
}

// getSessions is used to return in a thread-safe manner active or passive sessions
func (sS *SessionS) getSessions(originID string, pSessions bool) (ss []*Session) {
	ssMux := &sS.aSsMux  // get the pointer so we don't copy, otherwise locks will not work
	ssMp := sS.aSessions // reference it so we don't overwrite the new map without protection
	if pSessions {
		ssMux = &sS.pSsMux
		ssMp = sS.pSessions
	}
	ssMux.RLock()
	defer ssMux.RUnlock()
	if len(originID) == 0 {
		ss = make([]*Session, len(ssMp))
		var i int
		for _, s := range ssMp {
			ss[i] = s
			i++
		}
		return
	}
	if s, hasOptOriginID := ssMp[originID]; hasOptOriginID {
		ss = []*Session{s}
	}
	return
}

// getSessions is used to return in a thread-safe manner active or passive sessions
func (sS *SessionS) getSessionsFromOriginIDs(pSessions bool, originIDs ...string) (ss []*Session) {
	ssMux := &sS.aSsMux  // get the pointer so we don't copy, otherwise locks will not work
	ssMp := sS.aSessions // reference it so we don't overwrite the new map without protection
	if pSessions {
		ssMux = &sS.pSsMux
		ssMp = sS.pSessions
	}
	ssMux.RLock()
	defer ssMux.RUnlock()
	if len(originIDs) == 0 {
		ss = make([]*Session, len(ssMp))
		var i int
		for _, s := range ssMp {
			ss[i] = s
			i++
		}
		return
	}
	ss = make([]*Session, len(originIDs))
	for i, originID := range originIDs {
		if s, hasoriginID := ssMp[originID]; hasoriginID {
			ss[i] = s
		}
	}
	return
}

// transitSState will transit the sessions from one state (active/passive) to another (passive/active)
func (sS *SessionS) transitSState(originID string, psv bool) (s *Session) {
	ss := sS.getSessions(originID, !psv)
	if len(ss) == 0 {
		return
	}
	s = ss[0]
	s.lk.Lock()
	sS.unregisterSession(originID, !psv)
	sS.registerSession(s, psv)
	if !psv {
		sS.initSessionDebitLoops(s)
	} else { // transit from active with possible STerminator and DebitLoops
		s.stopSTerminator()
		s.stopDebitLoops()
	}
	s.lk.Unlock()
	return
}

// getActivateSession returns the session from active list or moves from passive
func (sS *SessionS) getActivateSession(originID string) (s *Session) {
	ss := sS.getSessions(originID, false)
	if len(ss) != 0 {
		return ss[0]
	}
	return sS.transitSState(originID, false)
}

// relocateSession will change the originID of a session (ie: prefix based session group)
func (sS *SessionS) relocateSession(ctx *context.Context, initOriginID, originID, originHost string) (s *Session) {
	/*
		if initOriginID == "" {
			return
		}
		initOptOriginID := utils.Sha1(initOriginID, originHost)
		newOptOriginID := utils.Sha1(originID, originHost)
		s = sS.getActivateSession(initOptOriginID)
		if s == nil {
			return
		}
		sS.unregisterSession(utils.IfaceAsString(s.ID), false)
		s.lk.Lock()
		// Overwrite initial originID with new one
		s.ID = newOptOriginID
		s.OriginCGREvent.APIOpts[utils.MetaOriginID] = newOptOriginID // Overwrite optOriginID for final CDR
		for _, sRun := range s.SRuns {
			//sRun.Event[utils.MetaOriginID] = newOptOriginID // needed for CDR generation
			sRun.Event[utils.OriginID] = originID
		}
		s.Unlock()
		sS.registerSession(s, false)
		sS.replicateSessions(ctx, initOptOriginID, false, sS.cfg.SessionSCfg().ReplicationConns)
	*/
	return
}

// getRelocateSession will relocate a session if it cannot find originID and initialOriginID is present
func (sS *SessionS) getRelocateSession(ctx *context.Context, optOriginID string, initOriginID,
	originID, originHost string) (s *Session) {
	if s = sS.getActivateSession(optOriginID); s != nil ||
		initOriginID == "" {
		return
	}
	return sS.relocateSession(ctx, initOriginID, originID, originHost)
}

// syncSessions synchronizes the active sessions with the one in the clients
// it will force-disconnect the one found in SessionS but not in clients
func (sS *SessionS) syncSessions(ctx *context.Context) {
	/*
		sS.aSsMux.RLock()
		asCount := len(sS.aSessions)
		sS.aSsMux.RUnlock()
		if asCount == 0 { // no need to sync the sessions if none is active
			return
		}
		type qReply struct {
			reply []string
			err   error
		}
		biClnts := sS.biJClients()
		replys := make(chan *qReply, len(biClnts))

		for _, clnt := range biClnts {
			ctx, cancel := context.WithTimeout(ctx, sS.cfg.GeneralCfg().ReplyTimeout)
			defer cancel()
			go func(clnt *biJClient) {
				var reply qReply
				reply.err = clnt.conn.Call(ctx, utils.SessionSv1GetActiveSessionIDs,
					utils.EmptyString, &reply.reply)
				replys <- &reply
			}(clnt)
		}
		queriedOriginIDs := utils.StringSet{}
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
				queriedOriginIDs.Add(sessionID.OptsOriginID())
			}
		}
		var toBeRemoved []string
		sS.aSsMux.RLock()
		for optOriginID := range sS.aSessions {
			if !queriedOriginIDs.Has(optOriginID) {
				toBeRemoved = append(toBeRemoved, optOriginID)
			}
		}
		sS.aSsMux.RUnlock()
		sS.terminateSyncSessions(ctx, toBeRemoved)
	*/
}

// Extracted from syncSessions in order to test all cases
func (sS *SessionS) terminateSyncSessions(ctx *context.Context, toBeRemoved []string) {
	/*
		for _, optOriginID := range toBeRemoved {
			ss := sS.getSessions(optOriginID, false)
			if len(ss) == 0 {
				continue
			}
			ss[0].Lock()
			if err := sS.forceSTerminate(ctx, ss[0], 0, nil, nil); err != nil {
				utils.Logger.Warning(
					fmt.Sprintf("<%s> failed force-terminating session: <%s>, err: <%s>",
						utils.SessionS, optOriginID, err.Error()))
			}
			ss[0].Unlock()
		}
	*/
}

// initSessionDebitLoops will init the debit loops for a session
// not thread-safe, it should be protected in another layer
func (sS *SessionS) initSessionDebitLoops(s *Session) {
	/*
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
	*/
}

// initSession handles a new session
// not thread-safe for Session since it is constructed here
func (sS *SessionS) initSession(ctx *context.Context, cgrEv *utils.CGREvent,
	clntConnID string, isInstantEvent bool) (s *Session, err error) {
	sID := utils.IfaceAsString(cgrEv.APIOpts[utils.MetaOriginID])
	if !isInstantEvent && sS.isSessionRegistered(sID, false) { // check if already exists
		return nil, utils.ErrExists
	}
	if s, err = sS.newSession(ctx, cgrEv, clntConnID); err != nil {
		return
	}
	if !isInstantEvent {
		s.lk.Lock() // avoid endsession before initialising
		sS.registerSession(s, false)
		s.lk.Unlock()
	}
	return
}

// updateSession will reset terminator, perform debits and replicate sessions
func (sS *SessionS) updateSession(ctx *context.Context, s *Session, updtEv, opts engine.MapEvent,
	dbtItvl time.Duration) (maxUsage map[string]time.Duration, err error) {
	defer func() {
		replConns, _ := engine.GetConnIDs(ctx, sS.cfg.SessionSCfg().Conns[utils.MetaReplication], utils.MetaAny, utils.MapStorage{}, sS.fltrS)
		sS.replicateSessions(ctx, s.ID, false, replConns)
	}()
	s.lk.Lock()
	defer s.lk.Unlock()

	// update fields from new event
	for k, v := range updtEv {
		if utils.ProtectedSFlds.Has(k) {
			continue
		}
		s.OriginCGREvent.Event[k] = v // update previoius field with new one
	}
	s.updateSRuns(updtEv, sS.cfg.SessionSCfg().AlterableFields)
	sS.setSTerminator(ctx, s, opts) // reset the terminator

	// TODO: Chargeable functionality not yet available in Session struct
	// event := &utils.CGREvent{
	//	Tenant:  s.OriginCGREvent.Tenant,
	//	Event:   updtEv,
	//	APIOpts: opts,
	// }
	// if s.Chargeable, err = engine.GetBoolOpts(ctx, event.Tenant, event.AsDataProvider(), nil, sS.fltrS, sS.cfg.SessionSCfg().Opts.Chargeable,
	//	utils.MetaChargeable); err != nil {
	//	return
	// }
	//init has no updtEv
	if updtEv == nil {
		updtEv = engine.MapEvent(s.OriginCGREvent.Event).Clone()
	}

	var reqMaxUsage time.Duration
	if _, err = updtEv.GetDuration(utils.Usage); err != nil {
		if err != utils.ErrNotFound {
			return
		}
		err = nil
		reqMaxUsage = sS.cfg.SessionSCfg().GetDefaultUsage(updtEv.GetStringIgnoreErrors(utils.ToR))
		updtEv[utils.Usage] = reqMaxUsage
	}
	maxUsage = make(map[string]time.Duration)
	for _, sr := range s.SRuns {
		reqType := engine.MapEvent(sr.CGREvent.Event).GetStringIgnoreErrors(utils.RequestType)
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
	aTime *time.Time, isInstantEvent bool) (err error) {
	s.lk.Lock()
	err = sS.endSession(ctx, s, tUsage, lastUsage, aTime, isInstantEvent)
	s.lk.Unlock()
	return
}

// endSession will end a session from outside
// this function is not thread safe
func (sS *SessionS) endSession(ctx *context.Context, s *Session, tUsage, lastUsage *time.Duration,
	aTime *time.Time, isInstantEvent bool) (err error) {
	if !isInstantEvent {
		//check if we have replicate connection and close the session there
		defer func() {
			replConns, _ := engine.GetConnIDs(ctx, sS.cfg.SessionSCfg().Conns[utils.MetaReplication], utils.MetaAny, utils.MapStorage{}, sS.fltrS)
			sS.replicateSessions(ctx, utils.IfaceAsString(s.OriginCGREvent.APIOpts[utils.MetaOriginID]), true, replConns)
		}()
		sS.unregisterSession(utils.IfaceAsString(s.OriginCGREvent.APIOpts[utils.MetaOriginID]), false)
		s.stopSTerminator()
		//s.stopDebitLoops()  // TODO: debit loops functionality will be implemented in future versions
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
		// 	if !isInstantEvent { // in case of one time charge there is no need of corrections
		// 		if notCharged := sUsage - sr.EventCost.GetUsage(); notCharged > 0 { // we did not charge enough, make a manual debit here
		// 			if !s.Chargeable {
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
		// 						engine.NewEventCostFromCallCost(cc, s.OptsStart[utils.MetaOriginID],
		// 							sr.Event.GetStringIgnoreErrors(utils.RunID)))
		// 				}
		// 			}
		// 		} else if notCharged < 0 { // charged too much, try refund
		// 			if err = sS.refundSession(s, sRunIdx, -notCharged); err != nil {
		// 				utils.Logger.Warning(
		// 					fmt.Sprintf(
		// 						"<%s> failed refunding session: <%s>, srIdx: <%d>, error: <%s>",
		// 						utils.SessionS, s.OptsStart[utils.MetaOriginID], sRunIdx, err.Error()))
		// 			}
		// 		}
		// 		if err := sS.roundCost(s, sRunIdx); err != nil { // will round the cost and refund the extra increment
		// 			utils.Logger.Warning(
		// 				fmt.Sprintf("<%s> failed rounding  session cost for <%s>, srIdx: <%d>, error: <%s>",
		// 					utils.SessionS, s.OptsStart[utils.MetaOriginID], sRunIdx, err.Error()))
		// 		}
		// 	}
		// 	// compute the event cost before saving the SessionCost
		// 	// add here to be applied for messages also
		// 	sr.EventCost.Compute()
		// 	if sS.cgrCfg.SessionSCfg().StoreSCosts {
		// 		if err := sS.storeSCost(s, sRunIdx); err != nil {
		// 			utils.Logger.Warning(
		// 				fmt.Sprintf("<%s> failed storing session cost for <%s>, srIdx: <%d>, error: <%s>",
		// 					utils.SessionS, s.OptsStart[utils.MetaOriginID], sRunIdx, err.Error()))
		// 		}
		// 	}

		// 	// set cost fields
		// 	sr.Event[utils.Cost] = sr.EventCost.GetCost()
		// 	sr.Event[utils.CostDetails] = utils.ToJSON(sr.EventCost) // avoid map[string]any when decoding
		// 	sr.Event[utils.CostSource] = utils.MetaSessionS
		// }
		// Set Usage field
		if sRunIdx == 0 {
			s.OriginCGREvent.Event[utils.Usage] = sr.TotalUsage
		}
		sr.CGREvent.Event[utils.Usage] = sr.TotalUsage
		if aTime != nil {
			sr.CGREvent.Event[utils.AnswerTime] = *aTime
		}
	}
	if errCh := engine.Cache.Set(ctx, utils.CacheClosedSessions, utils.IfaceAsString(s.OriginCGREvent.APIOpts[utils.MetaOriginID]), s,
		nil, true, utils.NonTransactional); errCh != nil {
		return errCh
	}
	return
}

// accountSMaxAbstracts computes the maximum abstract units for the events received
func (sS *SessionS) accountSMaxAbstracts(ctx *context.Context, cgrEvs map[string]*utils.CGREvent) (maxAbstracts map[string]*utils.Decimal, err error) {
	// resolve AccountS connections using the first event for tenant/dP context
	var acctConns []string
	for _, cgrEv := range cgrEvs {
		if acctConns, err = engine.GetConnIDs(ctx, sS.cfg.SessionSCfg().Conns[utils.MetaAccounts], cgrEv.Tenant, cgrEv.AsDataProvider(), sS.fltrS); err != nil {
			return
		}
		break
	}
	if len(acctConns) == 0 {
		return nil, utils.NewErrNotConnected(utils.AccountS)
	}
	maxAbstracts = make(map[string]*utils.Decimal)
	for runID, cgrEv := range cgrEvs {
		var acntCost utils.EventCharges
		if err = sS.connMgr.Call(ctx, acctConns,
			utils.AccountSv1MaxAbstracts, cgrEv, &acntCost); err != nil {
			return
		}
		maxAbstracts[runID] = acntCost.Abstracts // did not optimize here since we need to remove floats from acntCost
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
	/*
		if _, has := s.OptsStart[utils.MetaOriginID]; has {
			return utils.NewErrMandatoryIeMissing(utils.MetaOriginID)
		}
		if s.EventStart == nil { // remove
			if ureg := sS.unregisterSession(utils.IfaceAsString(s.OptsStart[utils.MetaOriginID]), true); !ureg {
				return utils.ErrNotFound
			}
			*reply = utils.OK
			return
		}
		if aSs := sS.getSessions(utils.IfaceAsString(s.OptsStart[utils.MetaOriginID]), false); len(aSs) != 0 { // found active session, transit to passive
			aSs[0].Lock()
			sS.unregisterSession(utils.IfaceAsString(s.OptsStart[utils.MetaOriginID]), false)
			aSs[0].stopSTerminator()
			aSs[0].stopDebitLoops()
			aSs[0].Unlock()
		}
		sS.registerSession(s, true)
	*/
	*reply = utils.OK
	return
}

// BiRPCv1ReplicateSessions will replicate active sessions to either args.Connections or the internal configured ones
// args.Filter is used to filter the sessions which are replicated, originID is the only one possible for now
func (sS *SessionS) BiRPCv1ReplicateSessions(ctx *context.Context,
	args ArgsReplicateSessions, reply *string) (err error) {
	sS.replicateSessions(ctx, utils.IfaceAsString(args.APIOpts[utils.MetaOriginID]), args.Passive, args.ConnIDs)
	*reply = utils.OK
	return
}

// BiRPCv1InitiateSessionWithDigest returns the formated result of InitiateSession
func (sS *SessionS) BiRPCv1InitiateSessionWithDigest(ctx *context.Context,
	args *utils.CGREvent, initReply *V1InitReplyWithDigest) (err error) {
	return
}

// BiRPCv1ProcessMessage processes one event with the right subsystems based on arguments received
func (sS *SessionS) BiRPCv1ProcessMessage(ctx *context.Context,
	args *utils.CGREvent, rply *V1ProcessMessageReply) (err error) {
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
	// for _, as := range aSs {
	// 	ss := sS.getSessions(as.CGRID, false)
	// 	if len(ss) == 0 {
	// 		continue
	// 	}
	// 	ss[0].Lock()
	// 	if errTerm := sS.forceSTerminate(ctx, ss[0], 0, nil, nil); errTerm != nil {
	// 		utils.Logger.Warning(
	// 			fmt.Sprintf(
	// 				"<%s> failed force-terminating session with id: <%s>, err: <%s>",
	// 				utils.SessionS, ss[0].originID(), errTerm.Error()))
	// 		err = utils.ErrPartiallyExecuted
	// 	}
	// 	ss[0].Unlock()
	// }
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
	var cdrsConns []string
	if cdrsConns, err = engine.GetConnIDs(ctx, sS.cfg.SessionSCfg().Conns[utils.MetaCDRs], cgrEv.Tenant, cgrEv.AsDataProvider(), sS.fltrS); err != nil {
		return
	}
	if len(cdrsConns) == 0 {
		return utils.NewErrNotConnected(utils.CDRs)
	}
	ev := engine.MapEvent(cgrEv.Event)
	originID := GetSetOptsOriginID(ev, cgrEv.APIOpts)
	s := sS.getRelocateSession(ctx, originID,
		ev.GetStringIgnoreErrors(utils.InitialOriginID),
		ev.GetStringIgnoreErrors(utils.OriginID),
		ev.GetStringIgnoreErrors(utils.OriginHost))
	if s != nil {
		utils.Logger.Warning(
			fmt.Sprintf("<%s> ProcessCDR called for active session with originID: <%s>",
				utils.SessionS, originID))
		s.lk.Lock() // events update session panic
		defer s.lk.Unlock()
	} else if sIface, has := engine.Cache.Get(utils.CacheClosedSessions, originID); has {
		// found in cache
		s = sIface.(*Session)
	} else { // no cached session, CDR will be handled by CDRs
		return sS.connMgr.Call(ctx, cdrsConns, utils.CDRsV1ProcessEvent,
			cgrEv, rply)
	}

	// Use previously stored Session to generate CDRs
	s.updateSRuns(ev, sS.cfg.SessionSCfg().AlterableFields)
	// create one CGREvent for each session run
	var withErrors bool
	for _, cgrEv := range s.asCGREvents() {
		if cgrEv.APIOpts == nil {
			cgrEv.APIOpts = make(map[string]any)
		}
		cgrEv.APIOpts[utils.MetaAttributes] = false
		cgrEv.APIOpts[utils.MetaChargers] = false
		if mp := engine.MapEvent(cgrEv.Event); unratedReqs.HasField(mp.GetStringIgnoreErrors(utils.RequestType)) { // order additional rating for unrated request types
			// argsProc.Flags = append(argsProc.Flags, fmt.Sprintf("%s:true", utils.MetaRALs))
		}
		if err = sS.connMgr.Call(ctx, cdrsConns, utils.CDRsV1ProcessEvent,
			cgrEv, rply); err != nil {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> error <%s> posting CDR with originID: <%s>",
					utils.SessionS, err.Error(), originID))
			withErrors = true
		}
	}
	if withErrors {
		err = utils.ErrPartiallyExecuted
	}
	return

}

// processThreshold will receive the event and send it to ThresholdS to be processed
func (sS *SessionS) processThreshold(ctx *context.Context, cgrEv *utils.CGREvent, clnb bool) (tIDs []string, err error) {
	var conns []string
	if conns, err = engine.GetConnIDs(ctx, sS.cfg.SessionSCfg().Conns[utils.MetaThresholds], cgrEv.Tenant, cgrEv.AsDataProvider(), sS.fltrS); err != nil {
		return
	}
	if len(conns) == 0 {
		return tIDs, utils.NewErrNotConnected(utils.ThresholdS)
	}
	cgrEv.SetCloneable(clnb)
	//initialize the returned variable
	err = sS.connMgr.Call(ctx, conns, utils.ThresholdSv1ProcessEvent, cgrEv, &tIDs)
	return
}

// processStats will receive the event and send it to StatS to be processed
func (sS *SessionS) processStats(ctx *context.Context, cgrEv *utils.CGREvent, clnb bool) (sIDs []string, err error) {
	var conns []string
	if conns, err = engine.GetConnIDs(ctx, sS.cfg.SessionSCfg().Conns[utils.MetaStats], cgrEv.Tenant, cgrEv.AsDataProvider(), sS.fltrS); err != nil {
		return
	}
	if len(conns) == 0 {
		return sIDs, utils.NewErrNotConnected(utils.StatS)
	}
	cgrEv.SetCloneable(clnb)
	//initialize the returned variable
	err = sS.connMgr.Call(ctx, conns, utils.StatSv1ProcessEvent, cgrEv, &sIDs)
	return
}

// getRoutes will receive the event and send it to SupplierS to find the suppliers
func (sS *SessionS) getRoutes(ctx *context.Context, cgrEv *utils.CGREvent) (routesReply routes.SortedRoutesList, err error) {
	var conns []string
	if conns, err = engine.GetConnIDs(ctx, sS.cfg.SessionSCfg().Conns[utils.MetaRoutes], cgrEv.Tenant, cgrEv.AsDataProvider(), sS.fltrS); err != nil {
		return
	}
	if len(conns) == 0 {
		return routesReply, utils.NewErrNotConnected(utils.RouteS)
	}
	if acd, has := cgrEv.Event[utils.ACD]; has {
		cgrEv.Event[utils.Usage] = acd
	}
	if err = sS.connMgr.Call(ctx, conns, utils.RouteSv1GetRoutes,
		cgrEv, &routesReply); err != nil {
		return routesReply, utils.NewErrRouteS(err)
	}
	return
}

// processAttributes will receive the event and send it to AttributeS to be processed
func (sS *SessionS) processAttributes(ctx *context.Context, cgrEv *utils.CGREvent) (rplyEv *attributes.AttrSProcessEventReply, err error) {
	var conns []string
	if conns, err = engine.GetConnIDs(ctx, sS.cfg.SessionSCfg().Conns[utils.MetaAttributes], cgrEv.Tenant, cgrEv.AsDataProvider(), sS.fltrS); err != nil {
		return
	}
	if len(conns) == 0 {
		return rplyEv, utils.NewErrNotConnected(utils.AttributeS)
	}
	if cgrEv.APIOpts == nil {
		cgrEv.APIOpts = make(engine.MapEvent)
	}
	cgrEv.APIOpts[utils.MetaSubsys] = utils.MetaSessionS
	cgrEv.APIOpts[utils.OptsContext] = utils.FirstNonEmpty(
		utils.IfaceAsString(cgrEv.APIOpts[utils.OptsContext]),
		utils.MetaSessionS)
	err = sS.connMgr.Call(ctx, conns, utils.AttributeSv1ProcessEvent,
		cgrEv, &rplyEv)
	return
}

// BiRPCv1AlterSession sends a RAR for the matching sessions
func (sS *SessionS) BiRPCv1AlterSession(ctx *context.Context,
	args utils.SessionFilterWithEvent, reply *string) (err error) {
	if args.SessionFilter == nil { //protection in case on nil
		args.SessionFilter = &utils.SessionFilter{}
	}
	aSs := sS.filterSessions(ctx, args.SessionFilter, false)
	if len(aSs) == 0 {
		return utils.ErrNotFound
	}
	// uniqueSIDs := utils.NewStringSet(nil)
	// for _, as := range aSs {
	// if uniqueSIDs.Has(as.CGRID) {
	// 	continue
	// }
	// uniqueSIDs.Add(as.CGRID)
	// ss := sS.getSessions(as.CGRID, false)
	// if len(ss) == 0 {
	// 	continue
	// }
	// if errTerm := sS.alterSession(ctx, ss[0], args.APIOpts, args.Event); errTerm != nil {
	// 	utils.Logger.Warning(
	// 		fmt.Sprintf(
	// 			"<%s> altering session with id '%s' failed: <%v>",
	// 			utils.SessionS, ss[0].cgrID(), errTerm))
	// 	err = utils.ErrPartiallyExecuted
	// }
	// }
	// if err != nil {
	// 	return
	// }
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
	attest := sS.cfg.SessionSCfg().STIRCfg.AllowedAttest
	if len(args.Attest) != 0 {
		attest = utils.NewStringSet(args.Attest)
	}
	stirMaxDur := sS.cfg.SessionSCfg().STIRCfg.PayloadMaxduration
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
		args.Payload.ATTest = sS.cfg.SessionSCfg().STIRCfg.DefaultAttest
	}
	if args.OverwriteIAT {
		args.Payload.IAT = time.Now().Unix()
	}
	if *identity, err = NewSTIRIdentity(
		ctx,
		utils.NewPASSporTHeader(utils.FirstNonEmpty(args.PublicKeyPath,
			sS.cfg.SessionSCfg().STIRCfg.PublicKeyPath)),
		args.Payload, utils.FirstNonEmpty(args.PrivateKeyPath,
			sS.cfg.SessionSCfg().STIRCfg.PrivateKeyPath),
		sS.cfg.GeneralCfg().ReplyTimeout); err != nil {
		return utils.NewSTIRError(err.Error())
	}
	return
}
