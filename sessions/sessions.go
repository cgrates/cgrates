// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package sessions

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/guardian"
)

var (
	// ErrForcedDisconnect is used to specify the reason why the session was disconnected
	ErrForcedDisconnect = errors.New("FORCED_DISCONNECT")
)

// NewSessionS constructs  a new SessionS instance
func NewSessionS(cgrCfg *config.CGRConfig, dm *engine.DataManager, cache *engine.CacheS, fltrS *engine.FilterS,
	connMgr *engine.ConnManager) *SessionS {
	cgrCfg.SessionSCfg().SessionIndexes.Add(utils.OriginID) // Make sure we have indexing for OriginID since it is a requirement on prefix searching

	return &SessionS{
		cfg:            cgrCfg,
		dm:             dm,
		cache:          cache,
		locker:         guardian.New(),
		storedSessions: make(utils.StringSet),
		fltrS:          fltrS,
		connMgr:        connMgr,
		biJClnts:       make(map[birpc.ClientConnector]string),
		biJIDs:         make(map[string]*biJClient),
		aSessions:      make(map[string]*session),
		aSessionsIdx:   make(map[string]map[string]map[string]utils.StringSet),
		aSessionsRIdx:  make(map[string][]*riFieldNameVal),
		pSessions:      make(map[string]*session),
		pSessionsIdx:   make(map[string]map[string]map[string]utils.StringSet),
		pSessionsRIdx:  make(map[string][]*riFieldNameVal),
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
	cache   *engine.CacheS
	locker  *guardian.Locker
	fltrS   *engine.FilterS
	connMgr *engine.ConnManager

	storedMu       sync.Mutex
	storedSessions utils.StringSet

	biJMux   sync.RWMutex                     // mux protecting BI-JSON connections
	biJClnts map[birpc.ClientConnector]string // index BiJSONConnection so we can sync them later
	biJIDs   map[string]*biJClient            // identifiers of bidirectional JSON conns, used to call RPC based on connIDs

	aSsMux    sync.RWMutex        // protects aSessions
	aSessions map[string]*session // group sessions per sessionId

	aSIMux        sync.RWMutex                                     // protects aSessionsIdx
	aSessionsIdx  map[string]map[string]map[string]utils.StringSet // map[fieldName]map[fieldValue][originID]utils.StringSet[runID]sID
	aSessionsRIdx map[string][]*riFieldNameVal                     // reverse indexes for active sessions, used on remove

	pSsMux    sync.RWMutex        // protects pSessions
	pSessions map[string]*session // group passive sessions based on originID

	pSIMux        sync.RWMutex                                     // protects pSessionsIdx
	pSessionsIdx  map[string]map[string]map[string]utils.StringSet // map[fieldName]map[fieldValue][originID]utils.StringSet[runID]sID
	pSessionsRIdx map[string][]*riFieldNameVal                     // reverse indexes for passive sessions, used on remove
}

// ListenAndServe starts the service and binds it to the listen loop
func (sS *SessionS) ListenAndServe(stopChan chan struct{}) {
	var syncC, storeC <-chan time.Time
	if interval := sS.cfg.SessionSCfg().ChannelSyncInterval; interval > 0 {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		syncC = ticker.C
	}
	if interval := sS.cfg.SessionSCfg().StoreInterval; interval > 0 {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		storeC = ticker.C
	}
	for {
		select {
		case <-stopChan:
			return
		case <-syncC:
			sS.syncSessions(context.TODO())
		case <-storeC:
			sS.storeSessions()
		}
	}
}

func (sS *SessionS) RestoreSessions(ctx context.Context) error {
	if sS.cfg.SessionSCfg().StoreInterval == 0 {
		return nil
	}
	sessions, err := sS.dm.GetSessions(ctx)
	if err != nil {
		return fmt.Errorf("restore sessions: %w", err)
	}
	for _, saved := range sessions {
		s := &session{session: saved}
		s.lk.Lock()
		sS.registerSession(s, saved.Passive)
		if !saved.Passive {
			sS.activateSession(s)
		}
		s.lk.Unlock()
	}
	return nil
}

func (sS *SessionS) removeStoredSession(ctx context.Context, id string) error {
	interval := sS.cfg.SessionSCfg().StoreInterval
	if interval == 0 {
		return nil
	}
	if interval > 0 {
		sS.scheduleSession(id)
		return nil
	}
	err := sS.dm.RemoveSession(ctx, id)
	if err != nil {
		logSessionStoreError(id, err)
	}
	return err
}

func logSessionStoreError(id string, err error) {
	utils.Logger.Warning(fmt.Sprintf("<%s> session storage error for CGRID %q: %v", utils.SessionS, id, err))
}

func (sS *SessionS) scheduleSession(id string) {
	sS.storedMu.Lock()
	sS.storedSessions.Add(id)
	sS.storedMu.Unlock()
}

func (sS *SessionS) storeSession(ctx context.Context, s *session) error {
	s.lk.RLock()
	defer s.lk.RUnlock()
	return sS.dm.SetSession(ctx, s.session)
}

func (sS *SessionS) storeSessions() {
	if sS.cfg.SessionSCfg().StoreInterval == 0 {
		return
	}
	// Leave IDs queued during this pass for the next save.
	sS.storedMu.Lock()
	ids := sS.storedSessions
	sS.storedSessions = make(utils.StringSet)
	sS.storedMu.Unlock()
	ctx := context.Background()
	for id := range ids {
		unlock := sS.locker.Lock(id)
		// A stopped session may already have been replaced under this CGRID.
		ss := sS.getSessions(id, false)
		if len(ss) == 0 {
			ss = sS.getSessions(id, true)
		}
		var err error
		if len(ss) != 0 {
			err = sS.storeSession(ctx, ss[0])
		} else {
			err = sS.dm.RemoveSession(ctx, id)
		}
		unlock()
		if err != nil {
			logSessionStoreError(id, err)
			sS.scheduleSession(id)
		}
	}
}

// Shutdown is called by engine to clear states
func (sS *SessionS) Shutdown() (err error) {
	if sS.cfg.SessionSCfg().StoreInterval != 0 {
		for _, s := range sS.getSessions("", false) {
			s.lk.Lock()
			s.stopSTerminator()
			s.lk.Unlock()
		}
		sS.storeSessions()
		return nil
	}
	replConns, err := engine.GetConnIDs(context.TODO(), sS.cfg.SessionSCfg().Conns, utils.MetaReplication, utils.MetaAny, utils.MapStorage{}, nil, sS.fltrS)
	if len(replConns) == 0 {
		var hasErr bool
		for _, s := range sS.getSessions("", false) { // Force sessions shutdown
			if err = sS.terminateSessionNew(context.TODO(), s); err != nil {
				hasErr = true
				utils.Logger.Warning(
					fmt.Sprintf("<%s> error: %s terminating session: %+v",
						utils.SessionS, err.Error(), s))
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
	timer   *time.Timer
	endChan chan struct{}
}

// setSTerminator must run with s.lk held so a save cannot copy the TTL while it is being updated.
func (sS *SessionS) setSTerminator(ctx context.Context, s *session, opts engine.MapEvent) {
	var err error
	origin := s.session.OriginCGREvent

	// TTL
	var ttl time.Duration
	if ttl, err = engine.GetDurationOptsFromMultipleMaps(ctx, origin.Tenant, origin.Event, opts, origin.APIOpts, sS.fltrS,
		sS.cfg.SessionSCfg().Opts.TTL, config.SessionsTTLDftOpt, utils.OptsSesTTL); err != nil {
		utils.Logger.Warning(
			fmt.Sprintf("<%s>, cannot extract <%s> for session:<%s>, from its options: <%s>, err: <%s>",
				utils.SessionS, utils.OptsSesTTL, origin.APIOpts[utils.MetaOriginID], opts, err))
		return
	}
	if ttl == 0 {
		return // nothing to set up
	}
	// random delay computation
	var maxDelay time.Duration
	if maxDelay, err = engine.GetDurationOptsFromMultipleMaps(ctx, origin.Tenant, origin.Event, opts, origin.APIOpts, sS.fltrS,
		sS.cfg.SessionSCfg().Opts.TTLMaxDelay, config.SessionsTTLMaxDelayDftOpt, utils.OptsSesTTLMaxDelay); err != nil {
		utils.Logger.Warning(
			fmt.Sprintf("<%s>, cannot extract <%s> for session:<%s>, from its options: <%s>, err: <%s>",
				utils.SessionS, utils.OptsSesTTLMaxDelay, origin.APIOpts[utils.MetaOriginID], opts.String(), err.Error()))
		return
	}
	if maxDelay != 0 {
		ttl += time.Duration(
			rand.Int63n(maxDelay.Milliseconds()) * time.Millisecond.Nanoseconds())
	}
	// LastUsed
	var ttlLastUsed *time.Duration
	if ttlLastUsed, err = engine.GetDurationPointerOptsFromMultipleMaps(ctx, origin.Tenant, origin.Event, opts, origin.APIOpts, sS.fltrS,
		sS.cfg.SessionSCfg().Opts.TTLLastUsed, utils.OptsSesTTLLastUsed); err != nil {
		utils.Logger.Warning(
			fmt.Sprintf("<%s>, cannot extract <%s> for session:<%s>, from its options: <%s>, err: <%s>",
				utils.SessionS, utils.OptsSesTTLLastUsed, origin.APIOpts[utils.MetaOriginID], opts.String(), err.Error()))
		return
	}
	// LastUsage
	var ttlLastUsage *time.Duration
	if ttlLastUsage, err = engine.GetDurationPointerOptsFromMultipleMaps(ctx, origin.Tenant, origin.Event, opts, origin.APIOpts, sS.fltrS,
		sS.cfg.SessionSCfg().Opts.TTLLastUsage, utils.OptsSesTTLLastUsage); err != nil {
		utils.Logger.Warning(
			fmt.Sprintf("<%s>, cannot extract <%s> for session:<%s>, from its options: <%s>, err: <%s>",
				utils.SessionS, utils.OptsSesTTLLastUsage, origin.APIOpts[utils.MetaOriginID], opts.String(), err.Error()))
		return
	}
	// TTLUsage
	var ttlUsage *time.Duration
	if ttlUsage, err = engine.GetDurationPointerOptsFromMultipleMaps(ctx, origin.Tenant, origin.Event, opts, origin.APIOpts, sS.fltrS,
		sS.cfg.SessionSCfg().Opts.TTLUsage, utils.OptsSesTTLUsage); err != nil {
		utils.Logger.Warning(
			fmt.Sprintf("<%s>, cannot extract <%s> for session:<%s>, from its options: <%s>, err: <%s>",
				utils.SessionS, utils.OptsSesTTLUsage, origin.APIOpts[utils.MetaOriginID], opts.String(), err.Error()))
		return
	}
	if s.session.TTL == nil {
		s.session.TTL = new(utils.SessionTTL)
	}
	t := s.session.TTL
	t.Deadline = time.Now().Add(ttl)
	t.Duration = ttl
	if ttlLastUsed != nil {
		t.LastUsed = ttlLastUsed
	}
	if ttlUsage != nil {
		t.Usage = ttlUsage
	}
	if ttlLastUsage != nil {
		t.LastUsage = ttlLastUsage
	}
	sS.scheduleTTLExpiry(ctx, s)
}

func (sS *SessionS) activateSession(s *session) {
	// Call after registration so immediate expiry can remove the session.
	if s.session.TTL != nil {
		sS.scheduleTTLExpiry(context.Background(), s)
	}
}

func (sS *SessionS) scheduleTTLExpiry(ctx context.Context, s *session) {
	delay := time.Until(s.session.TTL.Deadline)
	if s.sTerminator != nil && s.sTerminator.endChan != nil {
		s.sTerminator.timer.Reset(delay)
		return
	}
	s.sTerminator = &sTerminator{timer: time.NewTimer(delay), endChan: make(chan struct{})}

	go func(id string, endChan chan struct{}, timer *time.Timer) {
		select {
		case <-timer.C:
			var unlock func()
			if sS.cfg.SessionSCfg().StoreInterval != 0 {
				unlock = sS.locker.Lock(id)
			}
			s.lk.Lock()
			// The timer may have been cancelled or replaced while acquiring the locks.
			if s.sTerminator.endChan != endChan {
				s.lk.Unlock()
				if unlock != nil {
					unlock()
				}
				return
			}
			ttl := s.session.TTL
			lastUsage := ttl.Duration
			if ttl.LastUsage != nil {
				lastUsage = *ttl.LastUsage
			}
			_ = sS.forceSTerminate(ctx, s, lastUsage, ttl.Usage, ttl.LastUsed, unlock)
			s.lk.Unlock()
		case <-endChan:
			timer.Stop()
		}
	}(s.session.ID, s.sTerminator.endChan, s.sTerminator.timer)
	runtime.Gosched() // force context switching
}

// forceSTerminate is called when a session times-out or it is forced from CGRateS side
// not thread safe
func (sS *SessionS) forceSTerminate(ctx context.Context, s *session, extraUsage time.Duration, tUsage, lastUsed *time.Duration, unlock func()) (err error) {
	if extraUsage != 0 {
		// for i := range s.SRuns {
		// 	if _, err = sS.debitSession(ctx, s, i, extraUsage, lastUsed); err != nil {
		// 		utils.Logger.Warning(
		// 			fmt.Sprintf(
		// 				"<%s> failed debitting originID %s, sRunIdx: %d, err: %s",
		// 				utils.SessionS, s.ID, i, err.Error()))
		// 	}
		// }
	}
	// we apply the correction before
	if err = sS.endSession(ctx, s, tUsage, lastUsed, nil, false); err != nil {
		utils.Logger.Warning(
			fmt.Sprintf(
				"<%s> failed force terminating session with ID <%s>, err: <%s>",
				utils.SessionS, s.session.ID, err.Error()))
	}
	if unlock != nil {
		unlock()
	}
	tenant := s.session.OriginCGREvent.Tenant
	dP := s.session.OriginCGREvent.AsDataProvider()
	// post the URs using ees
	if ees, errEEs := engine.GetBoolOpts(ctx, tenant, dP, nil,
		sS.fltrS, sS.cfg.SessionSCfg().Opts.EEs,
		utils.MetaEEs); errEEs != nil {
		utils.Logger.Warning(
			fmt.Sprintf("<%s> error: %s processing event: %+v flag for %s",
				utils.SessionS, errEEs.Error(), s.asCGREventsMap(), utils.EEs))
	} else if ees {
		for _, cgrEv := range s.asCGREventsMap() {
			if cgrEv.APIOpts == nil {
				cgrEv.APIOpts = make(map[string]any)
			}
			cgrEv.APIOpts[utils.MetaAttributes] = false
			cgrEv.APIOpts[utils.MetaChargers] = false
			cgrEv.SetCloneable(true)
			if _, err = sS.eesProcessEvent(ctx, cgrEv); err != nil {
				utils.Logger.Warning(
					fmt.Sprintf("<%s> error: %s processing event: %+v with %s",
						utils.SessionS, err.Error(), cgrEv, utils.EEs))
			}
		}
	}
	// release the resources for the session
	if resSConns, errR := engine.GetConnIDs(ctx, sS.cfg.SessionSCfg().Conns, utils.MetaResources, tenant, dP, nil, sS.fltrS); errR != nil {
		utils.Logger.Warning(
			fmt.Sprintf("<%s> error resolving ResourceS connections: %s", utils.SessionS, errR.Error()))
	} else if len(resSConns) != 0 {
		var reply string
		args := s.session.OriginCGREvent.Clone()
		args.ID = utils.UUIDSha1Prefix()
		args.APIOpts[utils.OptsResourcesUsageID] = s.session.ID
		args.APIOpts[utils.OptsResourcesUnits] = 1
		if err := sS.connMgr.Call(ctx, resSConns,
			utils.ResourceSv1ReleaseResources,
			args, &reply); err != nil {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> error: %s could not release resource with resourceID: %s",
					utils.SessionS, err.Error(), s.session.ID))
		}
	}
	// release the ips for the session
	if ipsConns, errI := engine.GetConnIDs(ctx, sS.cfg.SessionSCfg().Conns, utils.MetaIPs, tenant, dP, nil, sS.fltrS); errI != nil {
		utils.Logger.Warning(
			fmt.Sprintf("<%s> error resolving IPs connections: %s", utils.SessionS, errI.Error()))
	} else if len(ipsConns) != 0 {
		var reply string
		args := s.session.OriginCGREvent.Clone()
		args.ID = utils.UUIDSha1Prefix()
		args.APIOpts[utils.OptsIPsAllocationID] = s.session.ID
		if err := sS.connMgr.Call(ctx, ipsConns,
			utils.IPsV1ReleaseIP,
			args, &reply); err != nil {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> error: %s could not release IP %q",
					utils.SessionS, err.Error(), s.session.ID))
		}
	}
	if errRpl := sS.replicateSessions(ctx, s.session.ID, false); errRpl != nil {
		utils.Logger.Warning(
			fmt.Sprintf("<%s> error: <%s> replicating session: <%s>",
				utils.SessionS, errRpl.Error(), s.session.ID))
	}
	if clnt := sS.biJClnt(s.session.ClientConnID); clnt != nil {
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
						utils.SessionS, err.Error(), s.session.ID))
			}
		}()
	}
	return
}

// debitSession performs debit for a session run
func (sS *SessionS) debitSession(ctx context.Context, s *session, sRunIdx int, dur time.Duration,
	lastUsed *time.Duration) (maxDur time.Duration, err error) {

	return
}

// debitLoopSession will periodically debit sessions, ie: automatic prepaid
// threadSafe since it will run into it's own goroutine
func (sS *SessionS) debitLoopSession(s *session, sRunIdx int,
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
func (sS *SessionS) disconnectSession(s *session, rsn string) (err error) {
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

// replicateSessions will replicate sessions with or without sID specified
func (sS *SessionS) replicateSessions(ctx context.Context, sID string, psv bool) (err error) {
	ss := sS.getSessions(sID, psv)
	if len(ss) == 0 {
		// emulate the Session, so it can be also removed from remote
		ss = []*session{{session: &utils.Session{ID: sID}}}
	}
	for _, s := range ss {
		sCln := s.Clone()
		var conns []string // ToDo: cache per session
		tnt := utils.MetaAny
		dP := utils.MapStorage{}
		if sCln.OriginCGREvent != nil {
			tnt = sCln.OriginCGREvent.Tenant
			dP = sCln.OriginCGREvent.AsDataProvider()
		}
		if conns, err = engine.GetConnIDs(ctx, sS.cfg.SessionSCfg().Conns, utils.MetaReplication,
			tnt, dP, nil, sS.fltrS); err != nil {
			return
		}
		if len(conns) == 0 {
			err = utils.NewErrNotConnected(utils.MetaReplication)
			return
		}
		var rply string
		if err = sS.connMgr.Call(ctx, conns,
			utils.SessionSv1SetPassiveSession,
			sCln, &rply); err != nil {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> cannot replicate session with id <%s>, err: %s",
					utils.SessionS, sCln.ID, err.Error()))
		}
	}
	return
}

// registerSession will register an active or passive Session
// called on init or relocate
// not thread safe for the Session
func (sS *SessionS) registerSession(s *session, passive bool) {
	s.session.Passive = passive
	sMux := &sS.aSsMux
	sMp := sS.aSessions
	if passive {
		sMux = &sS.pSsMux
		sMp = sS.pSessions
	}
	sMux.Lock()
	sMp[s.session.ID] = s
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
	_, has = sMp[sID]
	sMux.Unlock()
	return
}

// uregisterSession will unregister an active or passive session based on it's originID
// called on session terminate or relocate
func (sS *SessionS) unregisterSession(sID string, passive bool) bool {
	sMux := &sS.aSsMux
	sMp := sS.aSessions
	if passive {
		sMux = &sS.pSsMux
		sMp = sS.pSessions
	}
	sMux.Lock()
	if _, has := sMp[sID]; !has {
		sMux.Unlock()
		return false
	}
	delete(sMp, sID)
	sMux.Unlock()
	sS.unindexSession(sID, passive)
	return true
}

// indexSession will index an active or passive Session based on configuration
func (sS *SessionS) indexSession(s *session, pSessions bool) {
	srs := make([]*utils.SRun, 0, len(s.session.SRuns))
	for _, sr := range s.session.SRuns {
		srs = append(srs, sr)
	}
	sS.indexSRuns(s.session.ID, pSessions, srs...)
}

// indexSRuns will index the SRuns
func (sS *SessionS) indexSRuns(sID string, pSessions bool, srs ...*utils.SRun) {
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
		for _, sr := range srs {
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
			if _, hasID := ssIndx[fieldName][fieldVal][sID]; !hasID {
				ssIndx[fieldName][fieldVal][sID] = make(utils.StringSet) // we index runs under session id here
			}
			ssIndx[fieldName][fieldVal][sID].Add(sr.ID)

			// reverse index
			if _, hasIt := ssRIdx[sID]; !hasIt {
				ssRIdx[sID] = make([]*riFieldNameVal, 0)
			}
			ssRIdx[sID] = append(ssRIdx[sID], &riFieldNameVal{fieldName: fieldName, fieldValue: fieldVal})
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

func (sS *SessionS) getIndexedFilters(ctx context.Context, tenant string, fltrs []string) (
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
func (sS *SessionS) filterSessions(ctx context.Context, sf *utils.SessionFilter, psv bool) (aSs []*ExternalSession) {
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
	originIDs, matchingSRuns := sS.getSessionIDsMatchingIndexes(indx, psv)
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
		runIDs := matchingSRuns[s.session.ID]
		for runID, sr := range s.session.SRuns {
			if len(originIDs) != 0 && !runIDs.Has(runID) {
				continue
			}
			if pass(unindx, sr.CGREvent.Event) {
				aSs = append(aSs,
					s.AsExternalSession(runID, sS.cfg.GeneralCfg().NodeID))
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
func (sS *SessionS) filterSessionsCount(ctx context.Context, sf *utils.SessionFilter, psv bool) (count int) {
	count = 0
	if len(sf.Filters) == 0 {
		ss := sS.getSessions(utils.EmptyString, psv)
		for _, s := range ss {
			s.lk.RLock()
			count += len(s.session.SRuns)
			s.lk.RUnlock()
		}
		return
	}
	tenant := utils.FirstNonEmpty(sf.Tenant, sS.cfg.GeneralCfg().DefaultTenant)
	indx, unindx := sS.getIndexedFilters(ctx, tenant, sf.Filters)
	originIDs, matchingSRuns := sS.getSessionIDsMatchingIndexes(indx, psv)
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
		runIDs := matchingSRuns[s.session.ID]
		for runID, sr := range s.session.SRuns {
			if len(originIDs) != 0 && !runIDs.Has(runID) {
				continue
			}
			if pass(unindx, sr.CGREvent.Event) {
				count++
			}
		}
		s.lk.RUnlock()
	}
	return
}

// newSessionOutEvent is the new approach to creating session
// FixMe: rename as soon as we will remove the old newSession
func (sS *SessionS) newSessionOutEvent(ctx context.Context, sID string, cgrEv *utils.CGREvent,
	clntConnID string) (s *session, err error) {
	s = &session{session: &utils.Session{
		ID:             sID,
		OriginCGREvent: cgrEv,
		ClientConnID:   clntConnID,
		SRuns:          make(map[string]*utils.SRun),
	}}
	return
}

// setSession is called on each processEvent with *session flag to either create or update a session
// cch contain the preprocessed opts
func (sS *SessionS) setSession(ctx context.Context, cgrEv *utils.CGREvent,
	cch map[string]any, clientConnID string) (s *session, err error) {
	if cgrIDVal, has := cch[utils.MetaCGRid]; !has ||
		cgrIDVal.(string) == utils.EmptyString {
		return nil, utils.NewErrMandatoryIeMissing(utils.MetaCGRid)
	}
	id := cch[utils.MetaCGRid].(string)
	if s = sS.getActivateSession(id); s == nil {
		if s, err = sS.newSessionOutEvent(ctx, id, cgrEv, clientConnID); err != nil {
			return nil, err
		}
		sS.registerSession(s, false)
	}
	s.lk.Lock()
	sS.setSTerminator(ctx, s, cgrEv.APIOpts) // start termination timer
	s.lk.Unlock()
	return
}

// terminateSessionNew is the new way to terminate a session
func (sS *SessionS) terminateSessionNew(ctx context.Context, s *session) (err error) {
	s.lk.Lock()
	defer s.lk.Unlock()
	s.stopSTerminator()
	for _, run := range s.runStates {
		sS.DisableAutoChargeSRun(run)
	}
	for _, sRun := range s.session.SRuns {
		if sRun.Charges == nil {
			continue
		}
		sRun.CGREvent.APIOpts[utils.MetaAccountsCost] = sRun.Charges

		// corrections don't make much sense without previous charges
		if sRun.Charges.Abstracts == nil ||
			sRun.Charges.Abstracts.Compare(utils.NewDecimal(0, 0)) != 1 {
			continue
		}

		if sRun.UsageAdjustment != nil && sRun.UsageAdjustment.Compare(utils.NewDecimal(0, 0)) != 0 {
			switch sRun.UsageAdjustment.Compare(utils.NewDecimal(0, 0)) {
			case 1: // refund needs to be performed
				var rfndEC *utils.EventCharges
				if rfndEC, err = sRun.Charges.Truncate(utils.SubstractDecimal(sRun.Charges.Abstracts, sRun.UsageAdjustment)); err != nil {
					return
				}
				if err = sS.accountSRefundCharges(ctx, rfndEC, s.session.OriginCGREvent); err != nil {
					return
				}
			case -1: // debit more
				sRun.CGREvent.APIOpts[utils.MetaUsage] = utils.AbsoluteDecimal(sRun.UsageAdjustment)
				var acntCost *utils.EventCharges
				if acntCost, err = sS.accountSDebitEvent(ctx, sRun.CGREvent); err != nil {
					return
				}
				sRun.Charges.Merge(acntCost)
			}
		}
	}
	sS.unregisterSession(s.session.ID, false)
	return
}

// ensureDisconnect makes sure the session will be disconnected, attempting normal than forceDisconnect
func (sS *SessionS) ensureDisconnect(s *session, funcTag string, errIn error) (err error) {
	utils.Logger.Warning(
		fmt.Sprintf("<%s> funcTag: <%s> error: %s, session: %s, disconnecting session ...",
			utils.SessionS, funcTag, errIn.Error(), utils.ToJSON(s.session)))
	if err = sS.disconnectSession(s, utils.ErrServerError.Error()); err != nil {
		utils.Logger.Warning(
			fmt.Sprintf("<%s> error: %s attempted to disconnect session: %s, will forceDisconnect ...",
				utils.SessionS, err.Error(), utils.ToJSON(s.session)))
		if err = sS.forceSTerminate(context.TODO(), s, time.Duration(0), nil, nil, nil); err != nil {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> error: %s attempted to forceDisconnect session: %s",
					utils.SessionS, err.Error(), utils.ToJSON(s.session)))
		}
	}
	return
}

// EnableAutoChargeSRun will enable autocharging loop on a SRun
func (sS *SessionS) EnableAutoChargeSRun(s *session, sRunID string, autoChargeInterval time.Duration) {
	var autoChargeUsage, aCMinGranted *utils.Decimal
	var eEc *utils.EventCharges
	var err error
	var origAChrIval *time.Duration // populated if the original changes
	sr := s.session.SRuns[sRunID]
	state := s.runStates[sRunID]
	sr.AutoChargeInterval = &autoChargeInterval // small chance of concurrency, maybe add lock
	for {
		state.lk.RLock()
		cgrEv := sr.CGREvent.Clone()
		state.lk.RUnlock()

		if aCMinGranted, err = engine.GetDecimalOpts(context.TODO(), cgrEv.Tenant, cgrEv.AsDataProvider(), nil,
			sS.fltrS, sS.cfg.SessionSCfg().Opts.AutoChargeMinGranted,
			utils.MetaAutoChargeMinGranted); err != nil {
			sS.ensureDisconnect(s, utils.MetaAutoChargeMinGranted, err)
			return
		}
		if autoChargeUsage, err = engine.GetDecimalOpts(context.TODO(), cgrEv.Tenant, cgrEv.AsDataProvider(), nil,
			sS.fltrS, sS.cfg.SessionSCfg().Opts.AutoChargeUsage,
			utils.MetaAutoChargeUsageCfg); err != nil {
			sS.ensureDisconnect(s, utils.MetaAutoChargeUsageCfg, err)
			return
		}
		isACLoopUsage := utils.NewDecimal(int64(autoChargeInterval), 0).Compare(autoChargeUsage) == 0 // special case when we are looping for the same usage, ie: voice debit
		cgrEv.APIOpts[utils.MetaUsage] = autoChargeUsage                                              // dynamically change the next usage
		if eEc, err = sS.accountSDebitEvent(context.TODO(), cgrEv); err != nil {
			sS.ensureDisconnect(s, utils.AutoCharge, err)
			return
		}
		state.lk.Lock()
		sr.Charges.Merge(eEc)
		state.lk.Unlock()

		if eEc.Abstracts.Compare(aCMinGranted) == -1 { // allowed Abstracts are smaller than MinGranted setting
			sS.ensureDisconnect(s, utils.AutoCharge, utils.ErrInsufficientCredit)
			return
		}

		if origAChrIval != nil { // restore it from previous
			autoChargeInterval = *origAChrIval
			origAChrIval = nil
		}

		if eEc.Abstracts.Compare(autoChargeUsage) == -1 && isACLoopUsage { // granted smaller than requested, for loopUsage we try again when that empties
			origAChrIval = &autoChargeInterval
			if eAbs, isDur := eEc.Abstracts.Duration(); !isDur {
				sS.ensureDisconnect(s, utils.AutoCharge, utils.ErrServerError)
				return
			} else {
				autoChargeInterval = eAbs
			}
		}

		state.lk.Lock()
		sr.NextAutoCharge = utils.TimePointer(time.Now().Add(autoChargeInterval))
		state.lk.Unlock()

		select {
		case <-state.autoChargeStop:
			return
		case <-time.After(autoChargeInterval):
			continue
		}
	}
}

// DisableAutoChargeSRun will disable autocharging loop on a SRun
func (sS *SessionS) DisableAutoChargeSRun(sRun *sRun) {
	sRun.lk.Lock()
	if sRun.autoChargeStop != nil {
		close(sRun.autoChargeStop)
		sRun.autoChargeStop = nil
		sRun.run.AutoChargeInterval = nil
	}
	sRun.lk.Unlock()
}

// ipsAuthorize will authorize the event with the IPs subsystem
func (sS *SessionS) ipsAuthorize(ctx context.Context, cgrEv *utils.CGREvent) (rply *utils.AllocatedIP, err error) {
	var conns []string
	if conns, err = engine.GetConnIDs(ctx, sS.cfg.SessionSCfg().Conns, utils.MetaIPs,
		cgrEv.Tenant, cgrEv.AsDataProvider(), nil, sS.fltrS); err != nil {
		return
	}
	if len(conns) == 0 {
		return nil, utils.NewErrNotConnected(utils.IPs)
	}
	originID, _ := cgrEv.OptAsString(utils.MetaOriginID)
	if originID == "" {
		originID = utils.UUIDSha1Prefix()
	}
	cgrEv.APIOpts[utils.OptsIPsAllocationID] = originID
	var alcIP utils.AllocatedIP
	if err = sS.connMgr.Call(ctx, conns,
		utils.IPsV1AuthorizeIP,
		cgrEv, &alcIP); err != nil {
		utils.Logger.Warning(
			fmt.Sprintf("<%s> error: %s could not authorize IP for event: %+v",
				utils.SessionS, err.Error(), cgrEv))
	}
	return &alcIP, err
}

func (sS *SessionS) ipsAllocate(ctx context.Context, cgrEv *utils.CGREvent) (*utils.AllocatedIP, error) {
	conns, err := engine.GetConnIDs(ctx, sS.cfg.SessionSCfg().Conns, utils.MetaIPs, cgrEv.Tenant, cgrEv.AsDataProvider(), nil, sS.fltrS)
	if err != nil {
		return nil, err
	}
	if len(conns) == 0 {
		return nil, utils.NewErrNotConnected(utils.IPs)
	}
	originID, err := cgrEv.OptAsString(utils.MetaOriginID)
	if err != nil {
		return nil, err
	}
	if originID == "" {
		return nil, utils.NewErrMandatoryIeMissing(utils.MetaOriginID)
	}
	cgrEv.APIOpts[utils.OptsIPsAllocationID] = originID
	var allocatedIP utils.AllocatedIP
	if err := sS.connMgr.Call(ctx, conns, utils.IPsV1AllocateIP, cgrEv, &allocatedIP); err != nil {
		return nil, err
	}
	return &allocatedIP, nil
}

func (sS *SessionS) ipsRelease(ctx context.Context, cgrEv *utils.CGREvent) error {
	conns, err := engine.GetConnIDs(ctx, sS.cfg.SessionSCfg().Conns, utils.MetaIPs, cgrEv.Tenant, cgrEv.AsDataProvider(), nil, sS.fltrS)
	if err != nil {
		return err
	}
	if len(conns) == 0 {
		return utils.NewErrNotConnected(utils.IPs)
	}
	originID, err := cgrEv.OptAsString(utils.MetaOriginID)
	if err != nil {
		return err
	}
	if originID == "" {
		return utils.NewErrMandatoryIeMissing(utils.MetaOriginID)
	}
	cgrEv.APIOpts[utils.OptsIPsAllocationID] = originID
	var reply string
	return sS.connMgr.Call(ctx, conns, utils.IPsV1ReleaseIP, cgrEv, &reply)
}

// resourcesAuthorize will authorize the event with the Resources subsystem
func (sS *SessionS) resourcesAuthorize(ctx context.Context, cgrEv *utils.CGREvent) (resID string, err error) {
	var conns []string
	if conns, err = engine.GetConnIDs(ctx, sS.cfg.SessionSCfg().Conns, utils.MetaResources, cgrEv.Tenant, cgrEv.AsDataProvider(), nil, sS.fltrS); err != nil {
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

func (sS *SessionS) resourcesAllocate(ctx context.Context, cgrEv *utils.CGREvent) (string, error) {
	conns, err := engine.GetConnIDs(ctx, sS.cfg.SessionSCfg().Conns, utils.MetaResources, cgrEv.Tenant, cgrEv.AsDataProvider(), nil, sS.fltrS)
	if err != nil {
		return "", err
	}
	if len(conns) == 0 {
		return "", utils.NewErrNotConnected(utils.ResourceS)
	}
	resUsageID, err := engine.GetStringOpts(ctx, cgrEv.Tenant, cgrEv.AsDataProvider(),
		nil, sS.fltrS, sS.cfg.SessionSCfg().Opts.ResourcesUsageID,
		utils.OptsResourcesUsageID)
	if err != nil {
		return "", err
	}
	cgrEv.APIOpts[utils.OptsResourcesUsageID] = resUsageID
	resUnits, err := engine.GetIntOpts(ctx, cgrEv.Tenant, cgrEv.AsDataProvider(),
		nil, sS.fltrS, sS.cfg.SessionSCfg().Opts.ResourcesUnits,
		utils.OptsResourcesUnits)
	if err != nil {
		return "", err
	}
	cgrEv.APIOpts[utils.OptsResourcesUnits] = resUnits
	var resMessage string
	if err := sS.connMgr.Call(ctx, conns,
		utils.ResourceSv1AllocateResources,
		cgrEv, &resMessage); err != nil {
		return "", err
	}
	return resMessage, nil
}

func (sS *SessionS) resourcesRelease(ctx context.Context, cgrEv *utils.CGREvent) error {
	conns, err := engine.GetConnIDs(ctx, sS.cfg.SessionSCfg().Conns, utils.MetaResources, cgrEv.Tenant, cgrEv.AsDataProvider(), nil, sS.fltrS)
	if err != nil {
		return err
	}
	if len(conns) == 0 {
		return utils.NewErrNotConnected(utils.ResourceS)
	}
	resUsageID, err := engine.GetStringOpts(ctx, cgrEv.Tenant, cgrEv.AsDataProvider(),
		nil, sS.fltrS, sS.cfg.SessionSCfg().Opts.ResourcesUsageID,
		utils.OptsResourcesUsageID)
	if err != nil {
		return err
	}
	cgrEv.APIOpts[utils.OptsResourcesUsageID] = resUsageID
	var reply string
	return sS.connMgr.Call(ctx, conns,
		utils.ResourceSv1ReleaseResources,
		cgrEv, &reply)
}

// accountsMaxAbstracts will query the AccountS cost for Event
func (sS *SessionS) accountsMaxAbstracts(ctx context.Context, cgrEv *utils.CGREvent) (rply *utils.EventCharges, err error) {
	if _, hasUsage := cgrEv.APIOpts[utils.MetaUsage]; !hasUsage {
		return nil, utils.NewErrMandatoryIeMissing(utils.MetaUsage)
	}
	var conns []string
	if conns, err = engine.GetConnIDs(ctx, sS.cfg.SessionSCfg().Conns, utils.MetaAccounts,
		cgrEv.Tenant, cgrEv.AsDataProvider(), nil, sS.fltrS); err != nil {
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

// ratesCost will query the RateS cost for Event
func (sS *SessionS) ratesCost(ctx context.Context, cgrEv *utils.CGREvent) (*utils.RateProfileCost, error) {
	rateConns, err := engine.GetConnIDs(ctx, sS.cfg.SessionSCfg().Conns, utils.MetaRates, cgrEv.Tenant, cgrEv.AsDataProvider(), nil, sS.fltrS)
	if err != nil {
		return nil, err
	}
	if len(rateConns) == 0 {
		return nil, utils.NewErrNotConnected(utils.RateS)
	}
	var rtsCost utils.RateProfileCost
	if err := sS.connMgr.Call(ctx, rateConns,
		utils.RateSv1CostForEvent, cgrEv, &rtsCost); err != nil {
		return nil, err
	}
	return &rtsCost, nil
}

// accountSDebitEvent will debit the abstracts for the provided event
// if session is provided, it will try to debit first out of reserved balance
func (sS *SessionS) accountSDebitEvent(ctx context.Context, cgrEv *utils.CGREvent) (eEc *utils.EventCharges, err error) {
	if usageIface, hasUsage := cgrEv.APIOpts[utils.MetaUsage]; !hasUsage {
		return nil, utils.NewErrMandatoryIeMissing(utils.MetaUsage)
	} else if usage, _ := utils.IfaceAsDuration(usageIface); usage == time.Duration(0) {
		eEc = utils.NewEventCharges()
		eEc.Abstracts = utils.NewDecimal(0, 0)
		return
	}
	var conns []string
	if conns, err = engine.GetConnIDs(ctx, sS.cfg.SessionSCfg().Conns, utils.MetaAccounts,
		cgrEv.Tenant, cgrEv.AsDataProvider(), nil, sS.fltrS); err != nil {
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

// accountSRefundCharges will refund the charges specified into rfndEc
// cgrEv is provided as argument for filtering connections and the needed tenant
func (sS *SessionS) accountSRefundCharges(ctx context.Context, rfndEc *utils.EventCharges, cgrEv *utils.CGREvent) (err error) {
	if rfndEc == nil { // no refund to be performed
		return
	}
	var conns []string
	if conns, err = engine.GetConnIDs(ctx, sS.cfg.SessionSCfg().Conns, utils.MetaAccounts,
		cgrEv.Tenant, cgrEv.AsDataProvider(), nil, sS.fltrS); err != nil {
		return
	}
	if len(conns) == 0 {
		return utils.NewErrNotConnected(utils.AccountS)
	}
	var ign string
	return sS.connMgr.Call(ctx, conns,
		utils.AccountSv1RefundCharges,
		&utils.APIEventCharges{Tenant: cgrEv.Tenant, EventCharges: rfndEc}, &ign)
}

// getSessions is used to return in a thread-safe manner active or passive sessions
func (sS *SessionS) getSessions(originID string, pSessions bool) (ss []*session) {
	ssMux := &sS.aSsMux  // get the pointer so we don't copy, otherwise locks will not work
	ssMp := sS.aSessions // reference it so we don't overwrite the new map without protection
	if pSessions {
		ssMux = &sS.pSsMux
		ssMp = sS.pSessions
	}
	ssMux.RLock()
	defer ssMux.RUnlock()
	if len(originID) == 0 {
		ss = make([]*session, len(ssMp))
		var i int
		for _, s := range ssMp {
			ss[i] = s
			i++
		}
		return
	}
	if s, hasOptOriginID := ssMp[originID]; hasOptOriginID {
		ss = []*session{s}
	}
	return
}

// getSessions is used to return in a thread-safe manner active or passive sessions
func (sS *SessionS) getSessionsFromOriginIDs(pSessions bool, originIDs ...string) (ss []*session) {
	ssMux := &sS.aSsMux  // get the pointer so we don't copy, otherwise locks will not work
	ssMp := sS.aSessions // reference it so we don't overwrite the new map without protection
	if pSessions {
		ssMux = &sS.pSsMux
		ssMp = sS.pSessions
	}
	ssMux.RLock()
	defer ssMux.RUnlock()
	if len(originIDs) == 0 {
		ss = make([]*session, len(ssMp))
		var i int
		for _, s := range ssMp {
			ss[i] = s
			i++
		}
		return
	}
	ss = make([]*session, len(originIDs))
	for i, originID := range originIDs {
		if s, hasoriginID := ssMp[originID]; hasoriginID {
			ss[i] = s
		}
	}
	return
}

// transitSState will transit the sessions from one state (active/passive) to another (passive/active)
func (sS *SessionS) transitSState(sID string, psv bool) (s *session) {
	ss := sS.getSessions(sID, !psv)
	if len(ss) == 0 {
		return
	}
	s = ss[0]
	s.lk.Lock()
	sS.unregisterSession(sID, !psv)
	sS.registerSession(s, psv)
	if !psv {
		sS.activateSession(s)
	} else { // transit from active with possible STerminator and DebitLoops
		s.stopSTerminator()
		//s.stopDebitLoops()
	}
	s.lk.Unlock()
	return
}

// relocateSession will change the SessionID of an active session
func (sS *SessionS) relocateSession(ctx context.Context, oldSID, newSID string) (s *session) {
	ss := sS.getSessions(oldSID, false)
	if len(ss) == 0 {
		return
	}
	s = ss[0]
	s.lk.Lock()
	sS.unregisterSession(oldSID, false)
	s.session.ID = newSID
	for _, sRun := range s.session.SRuns {
		sRun.CGREvent.APIOpts[utils.MetaCGRid] = newSID
	}
	sS.registerSession(s, false)
	s.lk.Unlock()
	sS.replicateSessions(ctx, oldSID, false)
	sS.replicateSessions(ctx, newSID, false)
	return
}

// getActivateSession returns the session from active list or moves from passive
func (sS *SessionS) getActivateSession(sID string) (s *session) {
	ss := sS.getSessions(sID, false)
	if len(ss) != 0 {
		return ss[0]
	}
	return sS.transitSState(sID, false)
}

// getRelocateSession will relocate a session if it cannot find originID and initialOriginID is present
func (sS *SessionS) getRelocateSession(ctx context.Context, sID string, oldSID string) (s *session) {
	if s = sS.getActivateSession(sID); s != nil ||
		oldSID == utils.EmptyString {
		return
	}
	return sS.relocateSession(ctx, oldSID, sID)
}

// syncSessions synchronizes the active sessions with the one in the clients
// it will force-disconnect the one found in SessionS but not in clients
func (sS *SessionS) syncSessions(ctx context.Context) {
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
func (sS *SessionS) terminateSyncSessions(ctx context.Context, toBeRemoved []string) {
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

// endSession will end a session from outside
// this function is not thread safe
func (sS *SessionS) endSession(ctx context.Context, s *session, tUsage, lastUsage *time.Duration,
	aTime *time.Time, isInstantEvent bool) (err error) {
	if !isInstantEvent {
		//check if we have replicate connection and close the session there
		defer func() {
			_ = sS.replicateSessions(ctx, s.session.ID, false)
		}()
		sS.unregisterSession(s.session.ID, false)
		s.stopSTerminator()
		_ = sS.removeStoredSession(context.Background(), s.session.ID)
		//s.stopDebitLoops()  // TODO: debit loops functionality will be implemented in future versions
	}
	//for _, sr := range s.SRuns {
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

	// 	// set cost fields
	// 	sr.Event[utils.Cost] = sr.EventCost.GetCost()
	// 	sr.Event[utils.CostDetails] = utils.ToJSON(sr.EventCost) // avoid map[string]any when decoding
	// 	sr.Event[utils.CostSource] = utils.MetaSessionS
	// }
	// Set Usage field
	// 	if sRunIdx == 0 {
	// 		s.OriginCGREvent.Event[utils.Usage] = sr.TotalUsage
	// 	}
	// 	sr.CGREvent.Event[utils.Usage] = sr.TotalUsage
	// 	if aTime != nil {
	// 		sr.CGREvent.Event[utils.AnswerTime] = *aTime
	// 	}
	// }
	if errCh := sS.cache.Set(ctx, utils.CacheClosedSessions, utils.IfaceAsString(s.session.OriginCGREvent.APIOpts[utils.MetaOriginID]), s.session,
		nil, true, utils.NonTransactional); errCh != nil {
		return errCh
	}
	return
}

// accountSMaxAbstracts computes the maximum abstract units for the events received
func (sS *SessionS) accountSMaxAbstracts(ctx context.Context, cgrEvs map[string]*utils.CGREvent) (maxAbstracts map[string]*utils.Decimal, err error) {
	// resolve AccountS connections using the first event for tenant/dP context
	var acctConns []string
	for _, cgrEv := range cgrEvs {
		if acctConns, err = engine.GetConnIDs(ctx, sS.cfg.SessionSCfg().Conns, utils.MetaAccounts, cgrEv.Tenant, cgrEv.AsDataProvider(), nil, sS.fltrS); err != nil {
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
func (sS *SessionS) BiRPCv1GetActiveSessions(ctx context.Context,
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
func (sS *SessionS) BiRPCv1GetActiveSessionsCount(ctx context.Context,
	args *utils.SessionFilter, reply *int) error {
	if args == nil { //protection in case on nil
		args = &utils.SessionFilter{}
	}
	*reply = sS.filterSessionsCount(ctx, args, false)
	return nil
}

// BiRPCv1GetPassiveSessions returns the passive sessions handled by SessionS
func (sS *SessionS) BiRPCv1GetPassiveSessions(ctx context.Context,
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
func (sS *SessionS) BiRPCv1GetPassiveSessionsCount(ctx context.Context,
	args *utils.SessionFilter, reply *int) error {
	if args == nil { //protection in case on nil
		args = &utils.SessionFilter{}
	}
	*reply = sS.filterSessionsCount(ctx, args, true)
	return nil
}

// BiRPCv1SetPassiveSession used for replicating Sessions
func (sS *SessionS) BiRPCv1SetPassiveSession(ctx context.Context,
	s *utils.Session, reply *string) (err error) {
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
func (sS *SessionS) BiRPCv1ReplicateSessions(ctx context.Context,
	args ArgsReplicateSessions, reply *string) (err error) {
	sS.replicateSessions(ctx, utils.IfaceAsString(args.APIOpts[utils.MetaCGRid]), args.Passive)
	*reply = utils.OK
	return
}

// BiRPCv1ProcessMessage processes one event with the right subsystems based on arguments received
func (sS *SessionS) BiRPCv1ProcessMessage(ctx context.Context,
	args *utils.CGREvent, rply *V1ProcessMessageReply) (err error) {
	return
}

// BiRPCv1SyncSessions will sync sessions on demand
func (sS *SessionS) BiRPCv1SyncSessions(ctx context.Context,
	ignParam *utils.TenantWithAPIOpts, reply *string) error {
	sS.syncSessions(ctx)
	*reply = utils.OK
	return nil
}

// BiRPCv1ForceDisconnect will force disconnecting sessions matching sessions
func (sS *SessionS) BiRPCv1ForceDisconnect(ctx context.Context,
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
func (sS *SessionS) BiRPCv1RegisterInternalBiJSONConn(ctx context.Context,
	connID string, reply *string) error {
	sS.RegisterIntBiJConn(birpc.ClientFromContext(ctx), connID)
	*reply = utils.OK
	return nil
}

func (sS *SessionS) setSessionState(ctx context.Context, id string, passive bool) error {
	interval := sS.cfg.SessionSCfg().StoreInterval
	if interval != 0 {
		defer sS.locker.Lock(id)()
	}
	s := sS.transitSState(id, passive)
	if s == nil {
		state := "passive"
		if passive {
			state = "active"
		}
		utils.Logger.Warning(fmt.Sprintf("<%s> no %s session with id: <%s>", utils.SessionS, state, id))
		return utils.ErrNotFound
	}
	if interval == 0 {
		return nil
	}
	if interval > 0 {
		sS.scheduleSession(id)
		return nil
	}
	err := sS.storeSession(ctx, s)
	if err != nil {
		logSessionStoreError(id, err)
	}
	return err
}

// BiRPCv1ActivateSessions is called to activate a list/all sessions
// returns utils.ErrPartiallyExecuted in case of errors
func (sS *SessionS) BiRPCv1ActivateSessions(ctx context.Context,
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
		if sS.setSessionState(ctx, sID, false) != nil {
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
func (sS *SessionS) BiRPCv1DeactivateSessions(ctx context.Context,
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
		if sS.setSessionState(ctx, sID, true) != nil {
			err = utils.ErrPartiallyExecuted
		}
	}
	if err == nil {
		*reply = utils.OK
	}
	return
}

// hasSession will prove if a session with sID exists in either active, passive or closed cached sessions
func (sS *SessionS) hasSession(ctx context.Context, sID string) bool {
	return false
}

// processThreshold will receive the event and send it to ThresholdS to be processed
func (sS *SessionS) processThreshold(ctx context.Context, cgrEv *utils.CGREvent, clnb bool) (tIDs []string, err error) {
	var conns []string
	if conns, err = engine.GetConnIDs(ctx, sS.cfg.SessionSCfg().Conns, utils.MetaThresholds, cgrEv.Tenant, cgrEv.AsDataProvider(), nil, sS.fltrS); err != nil {
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
func (sS *SessionS) processStats(ctx context.Context, cgrEv *utils.CGREvent, clnb bool) (sIDs []string, err error) {
	var conns []string
	if conns, err = engine.GetConnIDs(ctx, sS.cfg.SessionSCfg().Conns, utils.MetaStats,
		cgrEv.Tenant, cgrEv.AsDataProvider(), nil, sS.fltrS); err != nil {
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
func (sS *SessionS) getRoutes(ctx context.Context, cgrEv *utils.CGREvent) (routesReply utils.SortedRoutesList, err error) {
	var conns []string
	if conns, err = engine.GetConnIDs(ctx, sS.cfg.SessionSCfg().Conns, utils.MetaRoutes,
		cgrEv.Tenant, cgrEv.AsDataProvider(), nil, sS.fltrS); err != nil {
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

// eesProcessEvent will send the CGREvent to EEs
func (sS *SessionS) eesProcessEvent(ctx context.Context, cgrEv *utils.CGREvent) (eesIDs []string, err error) {
	var conns []string
	if conns, err = engine.GetConnIDs(ctx, sS.cfg.SessionSCfg().Conns, utils.MetaEEs,
		cgrEv.Tenant, cgrEv.AsDataProvider(), nil, sS.fltrS); err != nil {
		return
	}
	if len(conns) == 0 {
		return eesIDs, utils.NewErrNotConnected(utils.EEs)
	}
	var procEEs map[string]map[string]any
	if err = sS.connMgr.Call(ctx, conns, utils.EeSv1ProcessEvent,
		&utils.CGREventWithEeIDs{CGREvent: cgrEv}, &procEEs); err != nil {
		return eesIDs, utils.NewErrEEs(err)
	}
	eesIDs = make([]string, 0, len(procEEs))
	for eeID := range procEEs {
		eesIDs = append(eesIDs, eeID)
	}
	return
}

// BiRPCv1AlterSession sends a RAR for the matching sessions
func (sS *SessionS) BiRPCv1AlterSession(ctx context.Context,
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
func (sS *SessionS) BiRPCv1DisconnectPeer(ctx context.Context,
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
func (sS *SessionS) BiRPCv1STIRAuthenticate(ctx context.Context,
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
	if err = AuthStirShaken(ctx, sS.cache, args.Identity, args.OriginatorTn, args.OriginatorURI,
		args.DestinationTn, args.DestinationURI, attest, stirMaxDur, sS.cfg.GeneralCfg().ReplyTimeout); err != nil {
		return utils.NewSTIRError(err.Error())
	}
	*reply = utils.OK
	return
}

// BiRPCv1STIRIdentity the API for STIR header creation
func (sS *SessionS) BiRPCv1STIRIdentity(ctx context.Context,
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
		sS.cache,
		utils.NewPASSporTHeader(utils.FirstNonEmpty(args.PublicKeyPath,
			sS.cfg.SessionSCfg().STIRCfg.PublicKeyPath)),
		args.Payload, utils.FirstNonEmpty(args.PrivateKeyPath,
			sS.cfg.SessionSCfg().STIRCfg.PrivateKeyPath),
		sS.cfg.GeneralCfg().ReplyTimeout); err != nil {
		return utils.NewSTIRError(err.Error())
	}
	return
}
