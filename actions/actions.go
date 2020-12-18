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
		crn:   cron.New(),
	}
}

// ActionS manages exection of Actions
type ActionS struct {
	cfg   *config.CGRConfig
	fltrS *engine.FilterS
	dm    *engine.DataManager
	crn   *cron.Cron
}

// ListenAndServe keeps the service alive
func (aS *ActionS) ListenAndServe(stopChan, cfgRld chan struct{}) {
	utils.Logger.Info(fmt.Sprintf("<%s> starting <%s>",
		utils.CoreS, utils.ActionS))
	aS.crn.Start()
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
	aS.crn.Stop()
	return
}

// Call implements rpcclient.ClientConnector interface for internal RPC
func (aS *ActionS) Call(serviceMethod string, args interface{}, reply interface{}) error {
	return utils.RPCCall(aS, serviceMethod, args, reply)
}

// matchingActionProfilesForEvent returns the matched ActionProfiles for the given event
func (aS *ActionS) matchingActionProfilesForEvent(tnt string, aPrflIDs []string,
	cgrEv *utils.CGREventWithOpts) (aPfs engine.ActionProfiles, err error) {
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

// scheduleActions is responsible for scheduling the action profiles matching cgrEv
func (aS *ActionS) scheduleActions(tnt string, aPrflIDs []string, cgrEv *utils.CGREventWithOpts, forceASAP bool) (err error) {
	var partExec bool
	var aPfs engine.ActionProfiles
	if aPfs, err = aS.matchingActionProfilesForEvent(tnt, aPrflIDs, cgrEv); err != nil {
		return
	}

	for _, aPf := range aPfs {
		ctx := context.Background()
		var trgActs map[string][]actioner // build here the list of actioners based on the trgKey
		var trgKey string
		for _, aCfg := range aPf.Actions { // create actioners and attach them to the right target
			if trgTyp := actionTarget(aCfg.Type); trgTyp != utils.META_NONE ||
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
		// build schedActSet
		var schedActSet []*scheduledActs
		for trg, acts := range trgActs {
			if trg == utils.META_NONE { // only one scheduledActs set
				schedActSet = append(schedActSet, newScheduledActs(aPf.Tenant, aPf.ID, trg, utils.EmptyString,
					ctx, &ActData{cgrEv.CGREvent.Event, cgrEv.Opts}, acts))
				continue
			}
			if len(aPf.Targets[trg]) == 0 {
				continue // no items selected
			}
			for trgID := range aPf.Targets[trg] {
				schedActSet = append(schedActSet, newScheduledActs(aPf.Tenant, aPf.ID, trg, trgID,
					ctx, &ActData{cgrEv.CGREvent.Event, cgrEv.Opts}, acts))
			}
		}
		// execute the actions
		for _, sActs := range schedActSet {
			if aPf.Schedule == utils.ASAP || forceASAP {
				go aS.asapExecuteActions(sActs)
				continue
			}
			if _, errExec := aS.crn.AddFunc(aPf.Schedule, sActs.ScheduledExecute); errExec != nil {
				utils.Logger.Warning(
					fmt.Sprintf(
						"<%s> scheduling ActionProfile with id: <%s:%s>, error: <%s>",
						utils.ActionS, sActs.tenant, sActs.apID, errExec))
				partExec = true
				continue
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

type ArgActionSv1ExecuteActions struct {
	*utils.CGREventWithOpts
	ActionProfileIDs []string
}

// V1ExecuteActions will be called to execute ASAP action profiles, ignoring their Schedule field
func (aS *ActionS) V1ExecuteActions(args *ArgActionSv1ExecuteActions, rpl *string) (err error) {
	if err = aS.scheduleActions(args.CGREventWithOpts.Tenant, args.ActionProfileIDs,
		args.CGREventWithOpts, true); err != nil {
		return
	}
	*rpl = utils.OK
	return
}
