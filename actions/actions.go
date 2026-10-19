// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package actions

import (
	"cmp"
	"errors"
	"fmt"
	"slices"
	"sync"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/cron"
)

// NewActionS instantiates the ActionS
func NewActionS(cfg *config.CGRConfig, cache *engine.CacheS, fltrS *engine.FilterS, dm *engine.DataManager, connMgr *engine.ConnManager) (aS *ActionS) {
	aS = &ActionS{
		cfg:     cfg,
		cache:   cache,
		connMgr: connMgr,
		fltrS:   fltrS,
		dm:      dm,
		crnLk:   new(sync.RWMutex),
	}
	aS.schedInit() // initialize cron and schedule actions
	return aS
}

// ActionS manages exection of Actions
type ActionS struct {
	cfg              *config.CGRConfig
	cache            *engine.CacheS
	connMgr          *engine.ConnManager
	fltrS            *engine.FilterS
	dm               *engine.DataManager
	crn              *cron.Cron
	crnLk            *sync.RWMutex
	profileSchedules map[string]*profileSchedule
}

type profileSchedule struct {
	entryIDs []cron.EntryID
	targets  map[string]utils.StringSet
}

// Shutdown is called to shutdown the service
func (aS *ActionS) Shutdown() {
	aS.crnLk.RLock()
	aS.crn.Stop()
	aS.crnLk.RUnlock()
}

// schedInit is called at service start
func (aS *ActionS) schedInit() {
	utils.Logger.Info(fmt.Sprintf("<%s> initializing scheduler.", utils.ActionS))
	tnts := []string{aS.cfg.GeneralCfg().DefaultTenant}
	if aS.cfg.ActionSCfg().Tenants != nil {
		tnts = *aS.cfg.ActionSCfg().Tenants
	}
	cgrEvs := make([]*utils.CGREvent, len(tnts))
	for i, tnt := range tnts {
		cgrEvs[i] = &utils.CGREvent{
			Tenant: tnt,
			APIOpts: map[string]any{
				utils.EventType: utils.SchedulerInit,
				utils.NodeID:    aS.cfg.GeneralCfg().NodeID,
			},
		}
	}
	aS.scheduleActions(context.Background(), cgrEvs, nil, false, false, true)
}

// scheduleActions will set up cron and load the matching data
func (aS *ActionS) scheduleActions(ctx *context.Context, cgrEvs []*utils.CGREvent, aPrflIDs []string, ignFilters, overwrite, crnReset bool) error {
	aS.crnLk.Lock()
	defer aS.crnLk.Unlock()
	crn := aS.crn
	if crnReset {
		crn = cron.New()
		aS.profileSchedules = make(map[string]*profileSchedule)
	}
	ignFilters = ignFilters && len(aPrflIDs) != 0
	var partExec bool
	for _, cgrEv := range cgrEvs {
		evNm := cgrEv.AsDataProvider()
		evNm[utils.MetaCfg] = aS.cfg.GetDataProvider()
		profiles, err := aS.matchingActionProfilesForEvent(ctx, cgrEv.Tenant, evNm, aPrflIDs, ignFilters)
		if err != nil {
			utils.Logger.Warning(fmt.Sprintf(
				"<%s> scheduler init, ignoring tenant: <%s>, error: <%s>",
				utils.ActionS, cgrEv.Tenant, err))
			partExec = true
			continue
		}
		for _, profile := range profiles {
			schedActs, err := aS.scheduledActionsForProfile(ctx, cgrEv.Tenant, cgrEv, profile, evNm, ignFilters)
			if err != nil {
				continue
			}
			key := profile.TenantID()
			scheduled := aS.profileSchedules[key]
			if !overwrite && scheduled != nil {
				schedActs = slices.DeleteFunc(schedActs, func(sActs *scheduledActs) bool {
					return scheduled.targets[sActs.trgTyp].Has(sActs.trgID)
				})
			}
			var schedule cron.Schedule
			if len(schedActs) != 0 && profile.Schedule != utils.MetaASAP {
				if schedule, err = cron.ParseStandard(profile.Schedule); err != nil {
					utils.Logger.Warning(fmt.Sprintf(
						"<%s> scheduling ActionProfile with id: <%s:%s>, error: <%s>",
						utils.ActionS, profile.Tenant, profile.ID, err))
					partExec = true
					continue
				}
			}
			if overwrite && scheduled != nil {
				for _, entryID := range scheduled.entryIDs {
					crn.Remove(entryID)
				}
				delete(aS.profileSchedules, key)
				scheduled = nil
			}
			if schedule != nil {
				if scheduled == nil {
					scheduled = &profileSchedule{targets: make(map[string]utils.StringSet)}
					aS.profileSchedules[key] = scheduled
				}
				entryID := crn.Schedule(schedule, cron.FuncJob(func() { aS.cronExecuteActions(schedActs) }))
				scheduled.entryIDs = append(scheduled.entryIDs, entryID)
				for _, sActs := range schedActs {
					if scheduled.targets[sActs.trgTyp] == nil {
						scheduled.targets[sActs.trgTyp] = make(utils.StringSet)
					}
					scheduled.targets[sActs.trgTyp].Add(sActs.trgID)
				}
			}
			if profile.Schedule == utils.MetaASAP {
				for _, sActs := range schedActs {
					go aS.asapExecuteActions(context.Background(), sActs)
				}
			}
		}
	}
	if crnReset {
		if aS.crn != nil {
			aS.crn.Stop()
		}
		aS.crn = crn
		aS.crn.Start()
	}
	if partExec {
		return utils.ErrPartiallyExecuted
	}
	return nil
}

// matchingActionProfilesForEvent returns the matched ActionProfiles for the given event
func (aS *ActionS) matchingActionProfilesForEvent(ctx *context.Context, tnt string,
	evNm utils.MapStorage, aPrflIDs []string, ignoreFilters bool) (aPfs []*utils.ActionProfile, err error) {
	if len(aPrflIDs) == 0 {
		var aPfIDMp utils.StringSet
		if aPfIDMp, err = engine.MatchingItemIDsForEvent(
			ctx,
			evNm,
			aS.cfg.ActionSCfg().StringIndexedFields,
			aS.cfg.ActionSCfg().PrefixIndexedFields,
			aS.cfg.ActionSCfg().SuffixIndexedFields,
			aS.cfg.ActionSCfg().ExistsIndexedFields,
			aS.cfg.ActionSCfg().NotExistsIndexedFields,
			aS.dm,
			utils.CacheActionProfilesFilterIndexes,
			tnt,
			aS.cfg.ActionSCfg().IndexedSelects,
			aS.cfg.ActionSCfg().NestedFields,
		); err != nil {
			return
		}
		aPrflIDs = aPfIDMp.AsSlice()
	}
	weights := make(map[string]float64) // stores sorting weights by profile ID
	for _, aPfID := range aPrflIDs {
		var aPf *utils.ActionProfile
		if aPf, err = aS.dm.GetActionProfile(ctx, tnt, aPfID,
			true, true, utils.NonTransactional); err != nil {
			if err == utils.ErrNotFound {
				err = nil
				continue
			}
			return
		}
		if !ignoreFilters {
			var pass bool
			if pass, err = aS.fltrS.Pass(ctx, tnt, aPf.FilterIDs, evNm); err != nil {
				return
			} else if !pass {
				continue
			}
		}
		weight, err := engine.WeightFromDynamics(ctx, aPf.Weights, aS.fltrS, tnt, evNm)
		if err != nil {
			return nil, err
		}
		weights[aPf.ID] = weight
		aPfs = append(aPfs, aPf)
	}
	if len(aPfs) == 0 {
		return nil, utils.ErrNotFound
	}

	// Sort by weight (higher values first).
	slices.SortFunc(aPfs, func(a, b *utils.ActionProfile) int {
		return cmp.Compare(weights[b.ID], weights[a.ID])
	})

	for i, aPf := range aPfs {
		var blocker bool
		if blocker, err = engine.BlockerFromDynamics(ctx, aPf.Blockers, aS.fltrS, tnt, evNm); err != nil {
			return
		}
		if blocker {
			aPfs = aPfs[0 : i+1]
			break
		}
	}
	return
}

// scheduledActions is responsible for scheduling the action profiles matching cgrEv
func (aS *ActionS) scheduledActions(ctx *context.Context, tnt string, cgrEv *utils.CGREvent, aPrflIDs []string,
	ignoreFilters bool) (schedActs []*scheduledActs, err error) {
	evNm := cgrEv.AsDataProvider()
	evNm[utils.MetaCfg] = aS.cfg.GetDataProvider() // add *cfg dataprovider for *cfg filters
	ignoreFilters = ignoreFilters && len(aPrflIDs) != 0
	aPfs, err := aS.matchingActionProfilesForEvent(ctx, tnt, evNm, aPrflIDs, ignoreFilters)
	if err != nil {
		return
	}
	for _, aPf := range aPfs {
		sActs, err := aS.scheduledActionsForProfile(ctx, tnt, cgrEv, aPf, evNm, ignoreFilters)
		if err != nil {
			continue
		}
		schedActs = append(schedActs, sActs...)
	}
	return
}

func (aS *ActionS) scheduledActionsForProfile(ctx *context.Context, tnt string, cgrEv *utils.CGREvent,
	profile *utils.ActionProfile, evNm utils.MapStorage, ignoreFilters bool) ([]*scheduledActs, error) {
	trgActs := make(map[string][]actioner)
	for _, aCfg := range profile.Actions {
		act, err := newActioner(ctx, cgrEv, aS.cfg, aS.cache, aS.fltrS, aS.dm, aS.connMgr, aCfg, tnt)
		if err != nil {
			utils.Logger.Warning(fmt.Sprintf(
				"<%s> ignoring ActionProfile with id: <%s:%s> creating action: <%s>, error: <%s>",
				utils.ActionS, profile.Tenant, profile.ID, aCfg.ID, err))
			return nil, err
		}
		trgTyp := actionTarget(aCfg.Type)
		trgActs[trgTyp] = append(trgActs[trgTyp], act)
	}
	var schedActs []*scheduledActs
	for trg, acts := range trgActs {
		if trg == utils.MetaNone {
			schedActs = append(schedActs, newScheduledActs(ctx, profile.Tenant, profile.ID, trg, "",
				ignoreFilters, evNm, acts))
			continue
		}
		for trgID := range profile.Targets[trg] {
			schedActs = append(schedActs, newScheduledActs(ctx, profile.Tenant, profile.ID, trg, trgID,
				ignoreFilters, evNm, acts))
		}
	}
	return schedActs, nil
}

// asapExecuteActions executes the scheduledActs and removes the executed from database
// uses locks to avoid concurrent access
func (aS *ActionS) asapExecuteActions(ctx *context.Context, sActs *scheduledActs) error {
	return aS.dm.Guard(ctx, func(ctx *context.Context) (err error) {
		var ap *utils.ActionProfile
		if ap, err = aS.dm.GetActionProfile(ctx, sActs.tenant, sActs.apID, true, true, utils.NonTransactional); err != nil {
			utils.Logger.Warning(
				fmt.Sprintf(
					"<%s> querying ActionProfile with id: <%s:%s>, error: <%s>",
					utils.ActionS, sActs.tenant, sActs.apID, err))
			return
		}
		if err = sActs.Execute(ctx); err != nil { // cannot remove due to errors on execution
			return
		}
		delete(ap.Targets[sActs.trgTyp], sActs.trgID)
		if err = aS.dm.SetActionProfile(ctx, ap, true); err != nil {
			utils.Logger.Warning(
				fmt.Sprintf(
					"<%s> saving ActionProfile with id: <%s:%s>, error: <%s>",
					utils.ActionS, sActs.tenant, sActs.apID, err))
		}
		return
	}, utils.ActionProfilePrefix+utils.ConcatenatedKey(sActs.tenant, sActs.apID))
}

// cronExecuteActions runs a batch of scheduled actions under the profile lock.
func (aS *ActionS) cronExecuteActions(schedActs []*scheduledActs) {
	ctx := context.Background()
	first := schedActs[0]
	err := aS.dm.Guard(ctx, func(ctx *context.Context) error {
		if !first.ignFilters {
			ap, err := aS.dm.GetActionProfile(ctx, first.tenant, first.apID, true, true, utils.NonTransactional)
			if err != nil {
				return err
			}
			pass, err := aS.fltrS.Pass(ctx, first.tenant, ap.FilterIDs, first.data)
			if err != nil || !pass {
				return err
			}
		}
		var err error
		for _, sActs := range schedActs {
			err = errors.Join(err, sActs.Execute(ctx))
		}
		return err
	}, utils.ActionProfilePrefix+utils.ConcatenatedKey(first.tenant, first.apID))
	if err != nil {
		utils.Logger.Warning(fmt.Sprintf(
			"<%s> executing scheduled ActionProfile <%s:%s>, error: <%s>",
			utils.ActionS, first.tenant, first.apID, err))
	}
}
