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

package stats

import (
	"cmp"
	"fmt"
	"maps"
	"runtime"
	"slices"
	"sync"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/guardian"
)

// matchedStatQueue is the unit matched by filters
type matchedStatQueue struct {
	statQueue *utils.StatQueue
	profile   *utils.StatQueueProfile
	ttl       *time.Duration // timeToLeave, picked on each init
	weight    float64
	lockID    string // ID of the lock used when matching the stat
}

// NewStatService initializes a StatService
func NewStatService(cfg *config.CGRConfig, dm *engine.DataManager, filters *engine.FilterS, cm *engine.ConnManager) *StatS {
	return &StatS{
		cfg:              cfg,
		dm:               dm,
		filters:          filters,
		cm:               cm,
		storedStatQueues: make(utils.StringSet),
		stopBackup:       make(chan struct{}),
	}
}

// StatS builds stats for events
type StatS struct {
	cfg     *config.CGRConfig
	dm      *engine.DataManager
	filters *engine.FilterS
	cm      *engine.ConnManager

	storedMu         sync.Mutex
	storedStatQueues utils.StringSet // stat queues that need saving

	stateMu    sync.Mutex // guards stopBackup
	stopBackup chan struct{}
	backupLoop sync.WaitGroup
}

// Reload restarts the backup loop. No-op after Shutdown.
func (s *StatS) Reload(ctx *context.Context) {
	s.stateMu.Lock()
	defer s.stateMu.Unlock()
	if s.stopBackup == nil {
		return
	}
	close(s.stopBackup)
	s.backupLoop.Wait()
	s.stopBackup = make(chan struct{})
	s.StartLoop(ctx)
}

// StartLoop starts the goroutine with the backup loop
func (s *StatS) StartLoop(ctx *context.Context) {
	s.backupLoop.Add(1)
	go s.runBackup(ctx)
}

// Shutdown is called to shutdown the service
func (s *StatS) Shutdown(ctx *context.Context) {
	s.stateMu.Lock()
	defer s.stateMu.Unlock()
	if s.stopBackup == nil {
		return
	}
	close(s.stopBackup)
	s.backupLoop.Wait()
	s.stopBackup = nil
	s.storeStats(ctx)
}

// runBackup will regularly store statQueues changed to DB
func (s *StatS) runBackup(ctx *context.Context) {
	defer s.backupLoop.Done()
	storeInterval := s.cfg.StatSCfg().StoreInterval
	if storeInterval <= 0 {
		return
	}
	for {
		s.storeStats(ctx)
		select {
		case <-s.stopBackup:
			return
		case <-time.After(storeInterval):
		}
	}
}

// storeStats represents one task of complete backup
func (s *StatS) storeStats(ctx *context.Context) {
	var failedSqIDs []string
	for { // don't stop untill we store all dirty statQueues
		s.storedMu.Lock()
		sID := s.storedStatQueues.GetOne()
		if sID != "" {
			s.storedStatQueues.Remove(sID)
		}
		s.storedMu.Unlock()
		if sID == "" {
			break // no more keys, backup completed
		}
		sqIf, ok := engine.Cache.Get(utils.CacheStatQueues, sID)
		if !ok || sqIf == nil {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> failed retrieving from cache stat queue with ID: %s",
					utils.StatService, sID))
			continue
		}
		sq := sqIf.(*utils.StatQueue)
		lockID := guardian.Guardian.GuardIDs("",
			s.cfg.GeneralCfg().LockingTimeout,
			utils.StatQueueLockKey(sq.Tenant, sq.ID))
		if err := s.StoreStatQueue(ctx, sq); err != nil {
			failedSqIDs = append(failedSqIDs, sID) // record failure so we can schedule it for next backup
		}
		guardian.Guardian.UnguardIDs(lockID)
		// randomize the CPU load and give up thread control
		runtime.Gosched()
	}
	if len(failedSqIDs) != 0 { // there were errors on save, schedule the keys for next backup
		s.storedMu.Lock()
		s.storedStatQueues.AddSlice(failedSqIDs)
		s.storedMu.Unlock()
	}
}

// StoreStatQueue stores the statQueue in DB
func (s *StatS) StoreStatQueue(ctx *context.Context, sq *utils.StatQueue) (err error) {
	if err = s.dm.SetStatQueue(ctx, sq); err != nil {
		utils.Logger.Warning(
			fmt.Sprintf("<StatS> failed saving StatQueue with ID: %s, error: %s",
				sq.TenantID(), err.Error()))
		return
	}
	//since we no longer handle cache in DataManager do here a manual caching
	if tntID := sq.TenantID(); engine.Cache.HasItem(utils.CacheStatQueues, tntID) { // only cache if previously there
		if err = engine.Cache.Set(ctx, utils.CacheStatQueues, tntID, sq, nil,
			true, utils.NonTransactional); err != nil {
			utils.Logger.Warning(
				fmt.Sprintf("<StatS> failed caching StatQueue with ID: %s, error: %s",
					tntID, err.Error()))
			return
		}
	}
	return
}

// matchingStatQueuesForEvent returns ordered list of matching statQueues which are active by the time of the call
func (s *StatS) matchingStatQueuesForEvent(ctx *context.Context, tnt string,
	args *utils.CGREvent) (sqs []*matchedStatQueue, unlock func(), err error) {
	unlockAll := func() {
		for _, m := range sqs {
			guardian.Guardian.UnguardIDs(m.lockID)
		}
	}

	evNm := args.AsDataProvider()
	statsIDs, err := engine.GetStringSliceOpts(ctx, tnt, evNm, nil, s.filters, s.cfg.StatSCfg().Opts.ProfileIDs,
		config.StatsProfileIDsDftOpt, utils.OptsStatsProfileIDs)
	if err != nil {
		return nil, nil, err
	}
	ignoreFilters, err := engine.GetBoolOpts(ctx, tnt, evNm, nil, s.filters, s.cfg.StatSCfg().Opts.ProfileIgnoreFilters,
		utils.MetaProfileIgnoreFilters)
	if err != nil {
		return nil, nil, err
	}

	sqIDs := utils.NewStringSet(statsIDs)
	if len(sqIDs) == 0 {
		ignoreFilters = false
		sqIDs, err = engine.MatchingItemIDsForEvent(ctx, evNm,
			s.cfg.StatSCfg().StringIndexedFields,
			s.cfg.StatSCfg().PrefixIndexedFields,
			s.cfg.StatSCfg().SuffixIndexedFields,
			s.cfg.StatSCfg().ExistsIndexedFields,
			s.cfg.StatSCfg().NotExistsIndexedFields,
			s.dm, utils.CacheStatFilterIndexes, tnt,
			s.cfg.StatSCfg().IndexedSelects,
			s.cfg.StatSCfg().NestedFields,
		)
		if err != nil {
			return nil, nil, err
		}
	}

	// Lock items in sorted order to prevent AB-BA deadlock.
	itemIDs := slices.Sorted(maps.Keys(sqIDs))

	sqs = make([]*matchedStatQueue, 0, len(itemIDs))
	for _, id := range itemIDs {
		lockID := guardian.Guardian.GuardIDs("",
			s.cfg.GeneralCfg().LockingTimeout,
			utils.StatQueueLockKey(tnt, id))
		sqPrfl, err := s.dm.GetStatQueueProfile(ctx, tnt, id, true, true, utils.NonTransactional)
		if err != nil {
			guardian.Guardian.UnguardIDs(lockID)
			if err == utils.ErrNotFound {
				continue
			}
			unlockAll()
			return nil, nil, err
		}
		if !ignoreFilters {
			pass, err := s.filters.Pass(ctx, tnt, sqPrfl.FilterIDs,
				evNm)
			if err != nil {
				guardian.Guardian.UnguardIDs(lockID)
				unlockAll()
				return nil, nil, err
			} else if !pass {
				guardian.Guardian.UnguardIDs(lockID)
				continue
			}
		}
		sq, err := s.dm.GetStatQueue(ctx, sqPrfl.Tenant, sqPrfl.ID, true, true, utils.EmptyString)
		if err != nil {
			guardian.Guardian.UnguardIDs(lockID)
			unlockAll()
			return nil, nil, err
		}
		var ttl *time.Duration
		if sqPrfl.TTL > 0 {
			ttl = utils.DurationPointer(sqPrfl.TTL)
		}
		if sqPrfl.TTL == -1 && sqPrfl.QueueLength == -1 {
			ttl = utils.DurationPointer(sqPrfl.TTL)
		}
		weight, err := engine.WeightFromDynamics(ctx, sqPrfl.Weights, s.filters, tnt, evNm)
		if err != nil {
			guardian.Guardian.UnguardIDs(lockID)
			unlockAll()
			return nil, nil, err
		}
		sqs = append(sqs, &matchedStatQueue{
			statQueue: sq,
			profile:   sqPrfl,
			ttl:       ttl,
			weight:    weight,
			lockID:    lockID,
		})
	}
	if len(sqs) == 0 {
		return nil, nil, utils.ErrNotFound
	}

	// Sort by weight (higher values first).
	slices.SortFunc(sqs, func(a, b *matchedStatQueue) int {
		return cmp.Compare(b.weight, a.weight)
	})

	return sqs, unlockAll, nil
}

func (s *StatS) getStatQueue(ctx *context.Context, tnt, id string) (sq *utils.StatQueue, err error) {
	if sq, err = s.dm.GetStatQueue(ctx, tnt, id, true, true, utils.EmptyString); err != nil {
		return
	}
	if _, err = remExpired(sq); err != nil {
		return nil, err
	}
	return
}

// processThresholds will pass the event for statQueue to ThresholdS
func (s *StatS) processThresholds(ctx *context.Context, sQs []*matchedStatQueue, opts map[string]any, tnt string, dP utils.DataProvider) (err error) {
	threshConns, err := engine.GetConnIDs(ctx, s.cfg.StatSCfg().Conns[utils.MetaThresholds], tnt, dP, s.filters)
	if err != nil {
		return
	}
	if len(threshConns) == 0 {
		return
	}
	if opts == nil {
		opts = make(map[string]any)
	}
	opts[utils.MetaEventType] = utils.StatUpdate
	var withErrs bool
	for _, m := range sQs {
		if len(m.profile.ThresholdIDs) == 1 &&
			m.profile.ThresholdIDs[0] == utils.MetaNone {
			continue
		}
		opts[utils.OptsThresholdsProfileIDs] = m.profile.ThresholdIDs
		thEv := &utils.CGREvent{
			Tenant: m.statQueue.Tenant,
			ID:     utils.GenUUID(),
			Event: map[string]any{
				utils.EventType: utils.StatUpdate,
				utils.StatID:    m.statQueue.ID,
			},
			APIOpts: opts,
		}
		for metricID, metric := range m.statQueue.SQMetrics {
			thEv.Event[metricID] = metric.GetValue()
		}

		var tIDs []string
		if err := s.cm.Call(ctx, threshConns,
			utils.ThresholdSv1ProcessEvent, thEv, &tIDs); err != nil &&
			(len(m.profile.ThresholdIDs) != 0 || err.Error() != utils.ErrNotFound.Error()) {
			utils.Logger.Warning(
				fmt.Sprintf("<StatS> error: %s processing event %+v with ThresholdS.", err.Error(), thEv))
			withErrs = true
		}
	}
	if withErrs {
		err = utils.ErrPartiallyExecuted
	}
	return
}

// processEEs will pass the event for statQueue to EEs
func (s *StatS) processEEs(ctx *context.Context, sQs []*matchedStatQueue, opts map[string]any, tnt string, dP utils.DataProvider) (err error) {
	eesConns, err := engine.GetConnIDs(ctx, s.cfg.StatSCfg().Conns[utils.MetaEEs], tnt, dP, s.filters)
	if err != nil {
		return
	}
	if len(eesConns) == 0 {
		return
	}
	var withErrs bool
	if opts == nil {
		opts = make(map[string]any)
	}
	for _, m := range sQs {
		metrics := make(map[string]any)
		for metricID, metric := range m.statQueue.SQMetrics {
			metrics[metricID] = metric.GetValue()
		}
		cgrEv := &utils.CGREvent{
			Tenant: m.statQueue.Tenant,
			ID:     utils.GenUUID(),
			Event: map[string]any{
				utils.EventType: utils.StatUpdate,
				utils.StatID:    m.statQueue.ID,
				utils.Metrics:   metrics,
			},
			APIOpts: opts,
		}

		cgrEventWithID := &utils.CGREventWithEeIDs{
			CGREvent: cgrEv,
			EeIDs:    s.cfg.StatSCfg().EEsExporterIDs,
		}
		var reply map[string]map[string]any
		if err := s.cm.Call(ctx, eesConns,
			utils.EeSv1ProcessEvent,
			&cgrEventWithID, &reply); err != nil &&
			err.Error() != utils.ErrNotFound.Error() {
			utils.Logger.Warning(
				fmt.Sprintf("<StatS> error: %s processing event %+v with EEs.", err.Error(), cgrEv))
			withErrs = true
		}
	}
	if withErrs {
		err = utils.ErrPartiallyExecuted
	}
	return
}

// processEvent processes a new event, dispatching to matching queues.
// Queues matching are also cached to speed up
func (s *StatS) processEvent(ctx *context.Context, tnt string, args *utils.CGREvent) (statQueueIDs []string, err error) {
	matchSQs, unlock, err := s.matchingStatQueuesForEvent(ctx, tnt, args)
	if err != nil {
		return nil, err
	}
	defer unlock()

	evNm := args.AsDataProvider()
	statQueueIDs = getStatQueueIDs(matchSQs)
	var withErrors bool
	for idx, m := range matchSQs {
		if err = m.processEvent(ctx, tnt, args.ID, s.filters, evNm); err != nil {
			utils.Logger.Warning(
				fmt.Sprintf("<StatS> Queue: %s, ignoring event: %s, error: %s",
					m.statQueue.TenantID(), utils.ConcatenatedKey(tnt, args.ID), err.Error()))
			withErrors = true
		}
		if s.cfg.StatSCfg().StoreInterval != 0 && m.profile.Stored {
			if s.cfg.StatSCfg().StoreInterval == -1 {
				s.StoreStatQueue(ctx, m.statQueue)
			} else {
				s.storedMu.Lock()
				s.storedStatQueues.Add(m.statQueue.TenantID())
				s.storedMu.Unlock()
			}
		}
		// verify the Blockers from the profiles
		// get the dynamic blocker from the profile and check if it pass trough its filters
		var blocker bool
		if blocker, err = engine.BlockerFromDynamics(ctx, m.profile.Blockers, s.filters, tnt, evNm); err != nil {
			return
		}
		if blocker && idx != len(matchSQs)-1 { // blocker will stop processing and we are not at last index
			break
		}

	}

	if s.processThresholds(ctx, matchSQs, args.APIOpts, tnt, evNm) != nil || s.processEEs(ctx, matchSQs, args.APIOpts, tnt, evNm) != nil || withErrors {
		err = utils.ErrPartiallyExecuted
	}
	return
}

// processEvent processes a utils.CGREvent, returns true if processed
func (m *matchedStatQueue) processEvent(ctx *context.Context, tnt, evID string, filterS *engine.FilterS, evNm utils.MapStorage) (err error) {

	//processing metrics without storing in the queue
	if oneEv := m.isOneEvent(); oneEv {
		return m.addStatOneEvent(ctx, tnt, filterS, evNm)
	}
	if _, err = remExpired(m.statQueue); err != nil {
		return
	}
	if err = m.remOnQueueLength(); err != nil {
		return
	}
	return m.addStatEvent(ctx, tnt, evID, filterS, evNm)
}

// remEventWithID removes an event from metrics
func remEventWithID(sq *utils.StatQueue, evID string) (err error) {
	for metricID, metric := range sq.SQMetrics {
		if err = metric.RemEvent(evID); err != nil {
			if err.Error() == utils.ErrNotFound.Error() {
				err = nil
				continue
			}
			utils.Logger.Warning(fmt.Sprintf("<StatQueue> metricID: %s, remove eventID: %s, error: %s", metricID, evID, err.Error()))
			return
		}
	}
	return
}

// remExpired expires items in queue
func remExpired(sq *utils.StatQueue) (removed int, err error) {
	var expIdx *int // index of last item to be expired
	for i, item := range sq.SQItems {
		if item.ExpiryTime == nil {
			break // items are ordered, so no need to look further
		}
		if item.ExpiryTime.After(time.Now()) {
			break
		}
		if err = remEventWithID(sq, item.EventID); err != nil {
			return
		}
		expIdx = utils.IntPointer(i)
	}
	if expIdx == nil {
		return
	}
	removed = *expIdx + 1
	sq.SQItems = sq.SQItems[removed:]
	return
}

// remOnQueueLength removes elements based on QueueLength setting
func (m *matchedStatQueue) remOnQueueLength() (err error) {
	if m.profile.QueueLength <= 0 { // infinite length
		return
	}
	if len(m.statQueue.SQItems) == m.profile.QueueLength { // reached limit, rem first element
		item := m.statQueue.SQItems[0]
		if err = remEventWithID(m.statQueue, item.EventID); err != nil {
			return
		}
		m.statQueue.SQItems = m.statQueue.SQItems[1:]
	}
	return
}

// addStatEvent computes metrics for an event
func (m *matchedStatQueue) addStatEvent(ctx *context.Context, tnt, evID string, filterS *engine.FilterS, evNm utils.MapStorage) (err error) {
	var expTime *time.Time
	if m.ttl != nil {
		expTime = utils.TimePointer(time.Now().Add(*m.ttl))
	}
	m.statQueue.SQItems = append(m.statQueue.SQItems, utils.SQItem{EventID: evID, ExpiryTime: expTime})
	var pass bool
	// recreate the request without *opts
	metricEvNm := utils.MapStorage{utils.MetaReq: evNm[utils.MetaReq], utils.MetaOpts: evNm[utils.MetaOpts]}

	dDP := engine.NewDynamicDP(ctx, config.CgrConfig(), tnt, metricEvNm, filterS)
	for idx, metricCfg := range m.profile.Metrics {
		if pass, err = filterS.Pass(ctx, tnt, metricCfg.FilterIDs,
			evNm); err != nil {
			return
		} else if !pass {
			continue
		}
		// in case of # metrics type
		if err = m.statQueue.SQMetrics[metricCfg.MetricID].AddEvent(evID, dDP); err != nil {
			utils.Logger.Warning(fmt.Sprintf("<StatQueue>: metric: %s, add eventID: %s, error: %s", metricCfg.MetricID,
				evID, err.Error()))
			return
		}
		// every metric has a blocker, verify them
		var blocker bool
		if blocker, err = engine.BlockerFromDynamics(ctx, metricCfg.Blockers, filterS, tnt, evNm); err != nil {
			return
		}
		if blocker && idx != len(m.profile.Metrics)-1 {
			break
		}
	}
	return
}

func (m *matchedStatQueue) isOneEvent() bool {
	return m.ttl != nil && *m.ttl == -1
}

func (m *matchedStatQueue) addStatOneEvent(ctx *context.Context, tnt string, filterS *engine.FilterS, evNm utils.MapStorage) (err error) {
	var pass bool

	metricEvNm := utils.MapStorage{utils.MetaReq: evNm[utils.MetaReq], utils.MetaOpts: evNm[utils.MetaOpts]}
	dDP := engine.NewDynamicDP(ctx, config.CgrConfig(), tnt, metricEvNm, filterS)

	for idx, metricCfg := range m.profile.Metrics {
		if pass, err = filterS.Pass(ctx, tnt, metricCfg.FilterIDs,
			evNm); err != nil {
			return
		} else if !pass {
			continue
		}

		if err = m.statQueue.SQMetrics[metricCfg.MetricID].AddOneEvent(dDP); err != nil {
			utils.Logger.Warning(fmt.Sprintf("<StatQueue>: metric: %s, error: %s", metricCfg.MetricID, err.Error()))
			return
		}

		var blocker bool
		if blocker, err = engine.BlockerFromDynamics(ctx, metricCfg.Blockers, filterS, tnt, evNm); err != nil {
			return
		}
		if blocker && idx != len(m.profile.Metrics)-1 {
			break
		}
	}
	return
}

// getStatQueueIDs returns a slice of IDs from the given matched StatQueues
func getStatQueueIDs(ms []*matchedStatQueue) []string {
	ids := make([]string, len(ms))
	for i, m := range ms {
		ids[i] = m.statQueue.ID
	}
	return ids
}
