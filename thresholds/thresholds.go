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

package thresholds

import (
	"cmp"
	"fmt"
	"maps"
	"runtime"
	"slices"
	"sync"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/attributes"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/guardian"
)

type ThresholdConfig struct {
	FilterIDs        []string
	MaxHits          int
	MinHits          int
	MinSleep         time.Duration
	Blocker          bool                 // blocker flag to stop processing on filters matched
	Weights          utils.DynamicWeights // Weight to sort the thresholds
	ActionProfileIDs []string
	Async            bool
	EeIDs            []string
	AttributeIDs     []string
}

// matchedThreshold is the unit matched by filters
type matchedThreshold struct {
	threshold *utils.Threshold
	profile   *utils.ThresholdProfile
	weight    float64
	lockID    string // ID of the lock used when matching the threshold
}

// NewThresholdService the constructor for ThresholdS service
func NewThresholdService(cfg *config.CGRConfig, dm *engine.DataManager, cache *engine.CacheS, filters *engine.FilterS, cm *engine.ConnManager) *ThresholdS {
	return &ThresholdS{
		cfg:              cfg,
		dm:               dm,
		cache:            cache,
		filters:          filters,
		cm:               cm,
		storedThresholds: make(utils.StringSet),
		stopBackup:       make(chan struct{}),
	}
}

// ThresholdS manages Threshold execution and storing them to DB
type ThresholdS struct {
	cfg     *config.CGRConfig
	dm      *engine.DataManager
	cache   *engine.CacheS
	filters *engine.FilterS
	cm      *engine.ConnManager

	storedMu         sync.Mutex
	storedThresholds utils.StringSet // thresholds that need saving

	stateMu    sync.Mutex // guards stopBackup
	stopBackup chan struct{}
	backupLoop sync.WaitGroup
}

// Reload restarts the backup loop. No-op after Shutdown.
func (s *ThresholdS) Reload(ctx *context.Context) {
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
func (s *ThresholdS) StartLoop(ctx *context.Context) {
	s.backupLoop.Add(1)
	go s.runBackup(ctx)
}

// Shutdown is called to shutdown the service
func (s *ThresholdS) Shutdown(ctx *context.Context) {
	s.stateMu.Lock()
	defer s.stateMu.Unlock()
	if s.stopBackup == nil {
		return
	}
	close(s.stopBackup)
	s.backupLoop.Wait()
	s.stopBackup = nil
	s.storeThresholds(ctx)
}

// backup will regularly store thresholds changed to DB
func (s *ThresholdS) runBackup(ctx *context.Context) {
	defer s.backupLoop.Done()
	storeInterval := s.cfg.ThresholdSCfg().StoreInterval
	if storeInterval <= 0 {
		return
	}
	for {
		s.storeThresholds(ctx)
		select {
		case <-s.stopBackup:
			return
		case <-time.After(storeInterval):
		}
	}
}

// storeThresholds represents one task of complete backup
func (s *ThresholdS) storeThresholds(ctx *context.Context) {
	var failedThresholds []string
	for {
		s.storedMu.Lock()
		tID := s.storedThresholds.GetOne()
		if tID != "" {
			s.storedThresholds.Remove(tID)
		}
		s.storedMu.Unlock()
		if tID == "" {
			break // no more keys, backup completed
		}
		tIf, ok := s.cache.Get(utils.CacheThresholds, tID)
		if !ok || tIf == nil {
			utils.Logger.Warning(fmt.Sprintf("<%s> failed retrieving from cache threshold with ID: %s", utils.ThresholdS, tID))
			continue
		}
		t := tIf.(*utils.Threshold)
		lockID := guardian.Guardian.GuardIDs("",
			s.cfg.GeneralCfg().LockingTimeout,
			utils.ThresholdLockKey(t.Tenant, t.ID))
		if err := s.storeThreshold(ctx, t); err != nil {
			failedThresholds = append(failedThresholds, tID) // record failure so we can schedule it for next backup
		}
		guardian.Guardian.UnguardIDs(lockID)
		// randomize the CPU load and give up thread control
		runtime.Gosched()
	}
	if len(failedThresholds) != 0 { // there were errors on save, schedule the keys for next backup
		s.storedMu.Lock()
		s.storedThresholds.AddSlice(failedThresholds)
		s.storedMu.Unlock()
	}
}

// storeThreshold stores the threshold in DB
func (s *ThresholdS) storeThreshold(ctx *context.Context, t *utils.Threshold) error {
	if err := s.dm.SetThreshold(ctx, t); err != nil {
		utils.Logger.Warning(
			fmt.Sprintf("<%s> failed saving Threshold with tenant: %s and ID: %s, error: %v",
				utils.ThresholdS, t.Tenant, t.ID, err))
		return err
	}
	//since we no longer handle cache in DataManager do here a manual caching
	if tntID := t.TenantID(); s.cache.HasItem(utils.CacheThresholds, tntID) { // only cache if previously there
		if err := s.cache.Set(ctx, utils.CacheThresholds, tntID, t, nil,
			true, utils.NonTransactional); err != nil {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> failed caching Threshold with ID: %s, error: %v",
					utils.ThresholdS, tntID, err))
			return err
		}
	}
	return nil
}

// matchingThresholdsForEvent returns ordered list of matching thresholds which are active for an Event
func (s *ThresholdS) matchingThresholdsForEvent(ctx *context.Context, tnt string,
	args *utils.CGREvent) (ts []*matchedThreshold, unlock func(), err error) {
	unlockAll := func() {
		for _, mt := range ts {
			guardian.Guardian.UnguardIDs(mt.lockID)
		}
	}

	evNm := utils.MapStorage{
		utils.MetaReq:  args.Event,
		utils.MetaOpts: args.APIOpts,
	}
	thIDs, err := engine.GetStringSliceOpts(ctx, tnt, evNm, nil, s.filters, s.cfg.ThresholdSCfg().Opts.ProfileIDs,
		config.ThresholdsProfileIDsDftOpt, utils.OptsThresholdsProfileIDs)
	if err != nil {
		return nil, nil, err
	}
	ignFilters, err := engine.GetBoolOpts(ctx, tnt, evNm, nil, s.filters, s.cfg.ThresholdSCfg().Opts.ProfileIgnoreFilters,
		utils.MetaProfileIgnoreFilters)
	if err != nil {
		return nil, nil, err
	}

	tIDs := utils.NewStringSet(thIDs)
	if len(tIDs) == 0 {
		ignFilters = false
		tIDs, err = engine.MatchingItemIDsForEvent(ctx, evNm,
			s.cfg.ThresholdSCfg().StringIndexedFields,
			s.cfg.ThresholdSCfg().PrefixIndexedFields,
			s.cfg.ThresholdSCfg().SuffixIndexedFields,
			s.cfg.ThresholdSCfg().ExistsIndexedFields,
			s.cfg.ThresholdSCfg().NotExistsIndexedFields,
			s.dm, utils.CacheThresholdFilterIndexes, tnt,
			s.cfg.ThresholdSCfg().IndexedSelects,
			s.cfg.ThresholdSCfg().NestedFields,
		)
		if err != nil {
			return nil, nil, err
		}
	}

	// Lock items in sorted order to prevent AB-BA deadlock.
	itemIDs := slices.Sorted(maps.Keys(tIDs))

	ts = make([]*matchedThreshold, 0, len(itemIDs))
	for _, id := range itemIDs {
		lockID := guardian.Guardian.GuardIDs("",
			s.cfg.GeneralCfg().LockingTimeout,
			utils.ThresholdLockKey(tnt, id))
		profile, err := s.dm.GetThresholdProfile(ctx, tnt, id, true, true, utils.NonTransactional)
		if err != nil {
			guardian.Guardian.UnguardIDs(lockID)
			if err == utils.ErrNotFound {
				err = nil
				continue
			}
			unlockAll()
			return nil, nil, err
		}
		if !ignFilters {
			var pass bool
			if pass, err = s.filters.Pass(ctx, tnt, profile.FilterIDs,
				evNm); err != nil {
				guardian.Guardian.UnguardIDs(lockID)
				unlockAll()
				return nil, nil, err
			} else if !pass {
				guardian.Guardian.UnguardIDs(lockID)
				continue
			}
		}
		threshold, err := s.dm.GetThreshold(ctx, profile.Tenant, profile.ID, true, true, "")
		if err != nil {
			guardian.Guardian.UnguardIDs(lockID)
			if err == utils.ErrNotFound { // corner case where the threshold was removed due to MaxHits
				err = nil
				continue
			}
			unlockAll()
			return nil, nil, err
		}
		weight, err := engine.WeightFromDynamics(ctx, profile.Weights,
			s.filters, tnt, evNm)
		if err != nil {
			guardian.Guardian.UnguardIDs(lockID)
			unlockAll()
			return nil, nil, err
		}
		ts = append(ts, &matchedThreshold{
			threshold: threshold,
			profile:   profile,
			weight:    weight,
			lockID:    lockID,
		})
	}
	if len(ts) == 0 {
		unlockAll()
		return nil, nil, utils.ErrNotFound
	}

	// Sort by weight (higher values first).
	slices.SortFunc(ts, func(a, b *matchedThreshold) int {
		return cmp.Compare(b.weight, a.weight)
	})

	for i, mt := range ts {
		if mt.profile.Blocker && i != len(ts)-1 { // blocker will stop processing and we are not at last index
			for _, dropped := range ts[i+1:] {
				guardian.Guardian.UnguardIDs(dropped.lockID)
			}
			ts = ts[:i+1]
			break
		}
	}
	return ts, unlockAll, nil
}

// processEvent processes a new event, dispatching to matching thresholds
func (s *ThresholdS) processEvent(ctx *context.Context, tnt string, args *utils.CGREvent) (thIDs []string, err error) {
	matchTs, unlock, err := s.matchingThresholdsForEvent(ctx, tnt, args)
	if err != nil {
		return nil, err
	}
	defer unlock()
	var withErrors bool
	thIDs = make([]string, 0, len(matchTs))
	for _, mt := range matchTs {
		thIDs = append(thIDs, mt.threshold.ID)
		if mt.profile.MaxHits != -1 && mt.threshold.Hits >= mt.profile.MaxHits { // threshold already reached max hits
			continue
		}
		mt.threshold.Hits++
		if time.Now().After(mt.threshold.Snooze) &&
			mt.threshold.Hits >= mt.profile.MinHits &&
			(mt.profile.MaxHits != 1 || mt.threshold.Hits <= mt.profile.MaxHits) { // threshold is active
			if args.APIOpts == nil {
				args.APIOpts = make(map[string]any)
			}
			args.APIOpts[utils.OptsActionsProfileIDs] = mt.profile.ActionProfileIDs
			evNm := utils.MapStorage{
				utils.MetaReq:  args.Event,
				utils.MetaOpts: args.APIOpts,
			}
			if len(mt.profile.AttributeIDs) != 0 && !slices.Contains(mt.profile.AttributeIDs, utils.MetaNone) {
				rplyAttrS, err := s.processAttributeS(ctx, tnt, mt, args)
				if err != nil {
					withErrors = true
					utils.Logger.Warning(fmt.Sprintf("<%s> failed processing event with Attributes: %s, error: %v", utils.ThresholdS, mt.threshold.TenantID(), err))
				}
				if rplyAttrS != nil && len(rplyAttrS.AlteredFields) != 0 {
					args = rplyAttrS.CGREvent
				}
			}
			actionConns, err := engine.GetConnIDs(ctx, s.cfg.ThresholdSCfg().Conns, utils.MetaActions, tnt, evNm, nil, s.filters)
			if err != nil {
				withErrors = true
				utils.Logger.Warning(fmt.Sprintf("<%s> failed resolving action connections for threshold: %s, error: %v", utils.ThresholdS, mt.threshold.TenantID(), err))
				continue
			}
			var reply string
			if !mt.profile.Async {
				if err = s.cm.Call(ctx, actionConns, utils.ActionSv1ExecuteActions, args, &reply); err != nil {
					withErrors = true
					utils.Logger.Warning(fmt.Sprintf("<%s> failed executing actions for threshold: %s, error: %v", utils.ThresholdS, mt.threshold.TenantID(), err))
				}
			} else {
				go func() {
					if errExec := s.cm.Call(context.Background(), actionConns, utils.ActionSv1ExecuteActions,
						args, &reply); errExec != nil {
						utils.Logger.Warning(fmt.Sprintf("<%s> failed executing actions for threshold: %s, error: %v", utils.ThresholdS, mt.threshold.TenantID(), errExec))
					}
				}()
			}
			if !withErrors {
				mt.threshold.Snooze = time.Now().Add(mt.profile.MinSleep)
			}
			if err := s.processEEs(ctx, args.APIOpts, mt, tnt, evNm); err != nil {
				utils.Logger.Warning(
					fmt.Sprintf("<%s> received error: %v when processing with EEs.", utils.ThresholdS, err))
				withErrors = true
			}
		}
		if s.cfg.ThresholdSCfg().StoreInterval == -1 {
			if err := s.storeThreshold(ctx, mt.threshold); err != nil {
				withErrors = true
			}
		} else {
			s.storedMu.Lock()
			s.storedThresholds.Add(mt.threshold.TenantID())
			s.storedMu.Unlock()
		}
	}
	if withErrors {
		err = utils.ErrPartiallyExecuted
	}
	return thIDs, err
}

// processAttributeS will process the event with AttributeS
func (s *ThresholdS) processAttributeS(ctx *context.Context, tnt string, mt *matchedThreshold, cgrEv *utils.CGREvent) (*attributes.ProcessEventReply, error) {
	attrConns, err := engine.GetConnIDs(ctx, s.cfg.ThresholdSCfg().Conns, utils.MetaAttributes, tnt, cgrEv.AsDataProvider(), nil, s.filters)
	if err != nil {
		return nil, err
	}
	cgrEv.APIOpts[utils.OptsThresholdsProfileIDs] = []string{mt.profile.ID}
	cgrEv.APIOpts[utils.OptsAttributesProfileIDs] = mt.profile.AttributeIDs
	cgrEv.APIOpts[utils.OptsContext] = utils.FirstNonEmpty(
		utils.IfaceAsString(cgrEv.APIOpts[utils.OptsContext]),
		utils.MetaThresholds)
	var rplyAttr attributes.ProcessEventReply
	if err = s.cm.Call(ctx, attrConns, utils.AttributeSv1ProcessEvent, cgrEv, &rplyAttr); err != nil {
		if err.Error() != utils.ErrNotFound.Error() {
			return nil, err
		}
	}
	return &rplyAttr, nil
}

func (s *ThresholdS) processEEs(ctx *context.Context, opts map[string]any, mt *matchedThreshold, tnt string, dP utils.DataProvider) error {
	var targetEeIDs []string
	if len(mt.profile.EeIDs) > 0 {
		targetEeIDs = mt.profile.EeIDs
		if isNone := slices.Contains(mt.profile.EeIDs, utils.MetaNone); isNone {
			targetEeIDs = []string{}
		}
	} else {
		targetEeIDs = s.cfg.ThresholdSCfg().EEsExporterIDs
	}
	eesConns, err := engine.GetConnIDs(ctx, s.cfg.ThresholdSCfg().Conns, utils.MetaEEs, tnt, dP, nil, s.filters)
	if err != nil {
		return err
	}
	if len(targetEeIDs) > 0 {
		if len(eesConns) == 0 {
			return utils.NewErrNotConnected(utils.EEs)
		}
	} else {
		return nil // no EEs to process
	}
	if opts == nil {
		opts = make(map[string]any)
	}
	sortedFilterIDs := slices.Clone(mt.profile.FilterIDs)
	slices.Sort(sortedFilterIDs)
	opts[utils.MetaEventType] = utils.ThresholdHit
	cgrEv := &utils.CGREvent{
		Tenant: mt.threshold.Tenant,
		ID:     utils.GenUUID(),
		Event: map[string]any{
			utils.EventType: utils.ThresholdHit,
			utils.ID:        mt.threshold.ID,
			utils.Hits:      mt.threshold.Hits,
			utils.Snooze:    mt.threshold.Snooze,
			utils.ThresholdConfig: ThresholdConfig{
				FilterIDs:        sortedFilterIDs,
				MaxHits:          mt.profile.MaxHits,
				MinHits:          mt.profile.MinHits,
				MinSleep:         mt.profile.MinSleep,
				Blocker:          mt.profile.Blocker,
				Weights:          mt.profile.Weights,
				ActionProfileIDs: mt.profile.ActionProfileIDs,
				Async:            mt.profile.Async,
				EeIDs:            mt.profile.EeIDs,
			},
		},
		APIOpts: opts,
	}
	cgrEventWithID := &utils.CGREventWithEeIDs{
		CGREvent: cgrEv,
		EeIDs:    targetEeIDs,
	}
	var reply map[string]map[string]any
	if mt.profile.Async {
		go func() {
			if errExec := s.cm.Call(context.TODO(), eesConns,
				utils.EeSv1ProcessEvent,
				cgrEventWithID, &reply); errExec != nil &&
				errExec.Error() != utils.ErrNotFound.Error() {
				utils.Logger.Warning(
					fmt.Sprintf("<%s> error: %v processing event %+v with EEs.", utils.ThresholdS, errExec, cgrEv))
			}
		}()
	} else if err := s.cm.Call(context.TODO(), eesConns,
		utils.EeSv1ProcessEvent,
		cgrEventWithID, &reply); err != nil &&
		err.Error() != utils.ErrNotFound.Error() {
		return err
	}
	return nil
}
