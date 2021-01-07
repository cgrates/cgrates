/*
Real-time Online/Offline Charging System (OerS) for Telecom & ISP environments
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

package actions

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/guardian"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/cron"
)

// NewActionS instantiates the ActionS
func NewActionS(cfg *config.CGRConfig, fltrS *engine.FilterS, dm *engine.DataManager) *ActionS {
	return &ActionS{
		cfg:   cfg,
		fltrS: fltrS,
		dm:    dm,
		crnLk: new(sync.RWMutex),
	}
}

// ActionS manages exection of Actions
type ActionS struct {
	cfg   *config.CGRConfig
	fltrS *engine.FilterS
	dm    *engine.DataManager
	crn   *cron.Cron
	crnLk *sync.RWMutex
}

// ListenAndServe keeps the service alive
func (aS *ActionS) ListenAndServe(stopChan, cfgRld chan struct{}) {
	utils.Logger.Info(fmt.Sprintf("<%s> starting <%s>",
		utils.CoreS, utils.ActionS))
	aS.schedInit() // initialize cron and schedule actions
	for {
		select {
		case <-stopChan:
			return
		case rld := <-cfgRld: // configuration was reloaded
			cfgRld <- rld
		}
	}
}

// Shutdown is called to shutdown the service
func (aS *ActionS) Shutdown() (err error) {
	utils.Logger.Info(fmt.Sprintf("<%s> shutdown <%s>", utils.CoreS, utils.ActionS))
	aS.crnLk.RLock()
	aS.crn.Stop()
	aS.crnLk.RUnlock()
	return
}

// Call implements rpcclient.ClientConnector interface for internal RPC
func (aS *ActionS) Call(serviceMethod string, args interface{}, reply interface{}) error {
	return utils.RPCCall(aS, serviceMethod, args, reply)
}

// schedInit is called at service start
func (aS *ActionS) schedInit() {
	utils.Logger.Info(fmt.Sprintf("<%s> initializing scheduler.", utils.ActionS))
	tnts := []string{aS.cfg.GeneralCfg().DefaultTenant}
	if aS.cfg.ActionSCfg().Tenants != nil {
		tnts = *aS.cfg.ActionSCfg().Tenants
	}
	cgrEvs := make([]*utils.CGREventWithOpts, len(tnts))
	for i, tnt := range tnts {
		cgrEvs[i] = &utils.CGREventWithOpts{
			CGREvent: &utils.CGREvent{
				Tenant: tnt,
				ID:     utils.GenUUID(),
				Time:   utils.TimePointer(time.Now()),
			},
			Opts: map[string]interface{}{
				utils.EventType: utils.SchedulerInit,
				utils.NodeID:    aS.cfg.GeneralCfg().NodeID,
			},
		}
	}
	aS.scheduleActions(cgrEvs, nil, true)
}

// scheduleActions will set up cron and load the matching data
func (aS *ActionS) scheduleActions(cgrEvs []*utils.CGREventWithOpts, aPrflIDs []string, crnReset bool) (err error) {
	aS.crnLk.Lock() // make sure we don't have parallel processes running  setu
	defer aS.crnLk.Unlock()
	crn := aS.crn
	if crnReset {
		crn = cron.New()
	}
	var partExec bool
	for _, cgrEv := range cgrEvs {
		var schedActSet []*scheduledActs
		if schedActSet, err = aS.scheduledActions(cgrEv.Tenant, cgrEv, aPrflIDs, false); err != nil {
			utils.Logger.Warning(
				fmt.Sprintf(
					"<%s> scheduler init, ignoring tenant: <%s>, error: <%s>",
					utils.ActionS, cgrEv.Tenant, err))
			partExec = true
			continue
		}
		for _, sActs := range schedActSet {
			if sActs.schedule == utils.ASAP {
				go aS.asapExecuteActions(sActs)
				continue
			}
			if _, err = crn.AddFunc(sActs.schedule, sActs.ScheduledExecute); err != nil {
				utils.Logger.Warning(
					fmt.Sprintf(
						"<%s> scheduling ActionProfile with id: <%s:%s>, error: <%s>",
						utils.ActionS, sActs.tenant, sActs.apID, err))
				partExec = true
				continue
			}
		}
	}
	if partExec {
		err = utils.ErrPartiallyExecuted
	}
	if crnReset {
		if aS.crn != nil {
			aS.crn.Stop()
		}
		aS.crn = crn
		aS.crn.Start()
	}
	return
}

// matchingActionProfilesForEvent returns the matched ActionProfiles for the given event
func (aS *ActionS) matchingActionProfilesForEvent(tnt string,
	cgrEv *utils.CGREventWithOpts, aPrflIDs []string) (aPfs engine.ActionProfiles, err error) {
	evNm := utils.MapStorage{
		utils.MetaReq:  cgrEv.CGREvent.Event,
		utils.MetaOpts: cgrEv.Opts,
	}
	if len(aPrflIDs) == 0 {
		var aPfIDMp utils.StringSet
		if aPfIDMp, err = engine.MatchingItemIDsForEvent(
			evNm,
			aS.cfg.ActionSCfg().StringIndexedFields,
			aS.cfg.ActionSCfg().PrefixIndexedFields,
			aS.cfg.ActionSCfg().SuffixIndexedFields,
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
	for _, aPfID := range aPrflIDs {
		var aPf *engine.ActionProfile
		if aPf, err = aS.dm.GetActionProfile(tnt, aPfID,
			true, true, utils.NonTransactional); err != nil {
			if err == utils.ErrNotFound {
				err = nil
				continue
			}
			return
		}
		if aPf.ActivationInterval != nil && cgrEv.Time != nil &&
			!aPf.ActivationInterval.IsActiveAtTime(*cgrEv.Time) { // not active
			continue
		}
		var pass bool
		if pass, err = aS.fltrS.Pass(tnt, aPf.FilterIDs, evNm); err != nil {
			return
		} else if !pass {
			continue
		}
		aPfs = append(aPfs, aPf)
	}
	if len(aPfs) == 0 {
		return nil, utils.ErrNotFound
	}
	aPfs.Sort()
	return
}

// scheduledActions is responsible for scheduling the action profiles matching cgrEv
func (aS *ActionS) scheduledActions(tnt string, cgrEv *utils.CGREventWithOpts, aPrflIDs []string,
	forceASAP bool) (schedActs []*scheduledActs, err error) {
	var partExec bool
	var aPfs engine.ActionProfiles
	if aPfs, err = aS.matchingActionProfilesForEvent(tnt, cgrEv, aPrflIDs); err != nil {
		return
	}

	for _, aPf := range aPfs {
		ctx := context.Background()
		var trgActs map[string][]actioner // build here the list of actioners based on the trgKey
		var trgKey string
		for _, aCfg := range aPf.Actions { // create actioners and attach them to the right target
			if trgTyp := actionTarget(aCfg.Type); trgTyp != utils.MetaNone ||
				trgKey == utils.EmptyString {
				trgKey = trgTyp
			}
			if act, errAct := newActioner(aS.cfg, aS.fltrS, aS.dm, aCfg); errAct != nil {
				utils.Logger.Warning(
					fmt.Sprintf(
						"<%s> ignoring ActionProfile with id: <%s:%s> creating action: <%s>, error: <%s>",
						utils.ActionS, aPf.Tenant, aPf.ID, aCfg.ID, errAct))
				partExec = true
				break
			} else {
				trgActs[trgKey] = append(trgActs[trgKey], act)
			}
		}
		if partExec {
			continue // skip this profile from processing further
		}
		for trg, acts := range trgActs {
			if trg == utils.MetaNone { // only one scheduledActs set
				schedActs = append(schedActs, newScheduledActs(aPf.Tenant, aPf.ID, trg, utils.EmptyString, aPf.Schedule,
					ctx, &ActData{cgrEv.CGREvent.Event, cgrEv.Opts}, acts))
				continue
			}
			if len(aPf.Targets[trg]) == 0 {
				continue // no items selected
			}
			for trgID := range aPf.Targets[trg] {
				schedActs = append(schedActs, newScheduledActs(aPf.Tenant, aPf.ID, trg, trgID, aPf.Schedule,
					ctx, &ActData{cgrEv.CGREvent.Event, cgrEv.Opts}, acts))
			}
		}
	}
	if partExec {
		err = utils.ErrPartiallyExecuted
	}
	return
}

// asapExecuteActions executes the scheduledActs and removes the executed from database
// uses locks to avoid concurrent access
func (aS *ActionS) asapExecuteActions(sActs *scheduledActs) (err error) {
	_, err = guardian.Guardian.Guard(func() (gRes interface{}, gErr error) {
		var ap *engine.ActionProfile
		if ap, gErr = aS.dm.GetActionProfile(sActs.tenant, sActs.apID, true, true, utils.NonTransactional); gErr != nil {
			utils.Logger.Warning(
				fmt.Sprintf(
					"<%s> querying ActionProfile with id: <%s:%s>, error: <%s>",
					utils.ActionS, sActs.tenant, sActs.apID, err))
			return
		}
		if gErr = sActs.Execute(); gErr != nil { // cannot remove due to errors on execution
			return
		}
		delete(ap.Targets[sActs.trgTyp], sActs.trgID)
		if gErr = aS.dm.SetActionProfile(ap, true); gErr != nil {
			utils.Logger.Warning(
				fmt.Sprintf(
					"<%s> saving ActionProfile with id: <%s:%s>, error: <%s>",
					utils.ActionS, sActs.tenant, sActs.apID, err))
		}
		return
	}, aS.cfg.GeneralCfg().LockingTimeout, utils.ActionProfilePrefix+sActs.apID)
	return
}

// V1ScheduleActions will be called to schedule actions matching the arguments
func (aS *ActionS) V1ScheduleActions(args *utils.ArgActionSv1ScheduleActions, rpl *string) (err error) {
	if err = aS.scheduleActions([]*utils.CGREventWithOpts{args.CGREventWithOpts},
		args.ActionProfileIDs, false); err != nil {
		return
	}
	*rpl = utils.OK
	return
}

// V1ExecuteActions will be called to execute ASAP action profiles, ignoring their Schedule field
func (aS *ActionS) V1ExecuteActions(args *utils.ArgActionSv1ScheduleActions, rpl *string) (err error) {
	var schedActSet []*scheduledActs
	if schedActSet, err = aS.scheduledActions(args.CGREventWithOpts.Tenant,
		args.CGREventWithOpts, args.ActionProfileIDs, true); err != nil {
		return
	}
	var partExec bool
	// execute the actions
	for _, sActs := range schedActSet {
		if err = aS.asapExecuteActions(sActs); err != nil {
			partExec = true
		}
	}
	if partExec {
		err = utils.ErrPartiallyExecuted
		return
	}
	*rpl = utils.OK
	return
}
