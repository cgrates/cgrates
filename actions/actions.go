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
	"fmt"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
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

/*
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
		delete(ap.AccountIDs, sActs.apID)
		if len(ap.AccountIDs) == 0 {
			gErr = aS.dm.RemoveActionProfile(sActs.tenant, sActs.apID, utils.NonTransactional, true)
		} else {
			gErr = aS.dm.SetActionProfile(ap, true)
		}
		if gErr != nil {
			utils.Logger.Warning(
				fmt.Sprintf(
					"<%s> saving ActionProfile with id: <%s:%s>, error: <%s>",
					utils.ActionS, sActs.tenant, sActs.apID, err))
		}
		return
	}, aS.cfg.GeneralCfg().LockingTimeout, utils.ActionProfilePrefix+sActs.apID)
	return
}

// scheduleActions is responsible for scheduling the actions needing execution
func (aS *ActionS) scheduleActions(tnt string, aPrflIDs []string, cgrEv *utils.CGREventWithOpts) (err error) {
	var aPfs engine.ActionProfiles
	if aPfs, err = aS.matchingActionProfilesForEvent(tnt, aPrflIDs, cgrEv); err != nil {
		return
	}
	for _, aPf := range aPfs {
		ctx := context.Background()
		var acts []actioner
		// actsExec will be used bellow as common code block
		actsExec := func(acntID string) (errExec error) {
			if len(acts) == 0 { // not yet initialized
				if acts, errExec = newActionersFromActions(aS.cfg, aS.fltrS, aS.dm, aPf.Actions); errExec != nil {
					return
				}
			}
			sActs := newScheduledActs(aPf.Tenant, aPf.ID, acntID, ctx,
				&ActData{cgrEv.CGREvent.Event, cgrEv.Opts}, acts)
			if aPf.Schedule == utils.ASAP {
				go aS.asapExecuteActions(sActs)
				return
			}
			if _, errExec = aS.crn.AddFunc(aPf.Schedule, sActs.ScheduledExecute); errExec != nil {
				utils.Logger.Warning(
					fmt.Sprintf(
						"<%s> scheduling ActionProfile with id: <%s:%s>, error: <%s>",
						utils.ActionS, sActs.tenant, sActs.apID, errExec))
				errExec = nil
			}
			return
		}
		if len(aPf.AccountIDs) == 0 { // no accounts, other acts
			if err = actsExec(utils.EmptyString); err != nil {
				return err
			}
			continue
		}
		for acntID := range aPf.AccountIDs {
			if err = actsExec(acntID); err != nil {
				return err
			}
		}
	}
	return
}

type ArgActionSv1ExecuteActions struct {
	*utils.CGREventWithOpts
	ActionProfileIDs []string
}

// V1ExecuteActions will be called to execute ASAP action profiles, ignoring their Schedule field
func (aS *ActionS) V1ExecuteActions(args *ArgActionSv1ExecuteActions, rpl *string) (err error) {
	var aPfs engine.ActionProfiles
	if aPfs, err = aS.matchingActionProfilesForEvent(args.Tenant, args.ActionProfileIDs,
		args.CGREventWithOpts); err != nil {
		return utils.NewErrServerError(err)
	}
	var partExec bool
	for _, aPf := range aPfs {
		ctx := context.Background()
		var acts []actioner
		// actsExec will be used bellow as common code block
		actsExec := func(acntID string) (errExec error) {
			if len(acts) == 0 { // not yet initialized
				if acts, errExec = newActionersFromActions(aS.cfg, aS.fltrS, aS.dm, aPf.Actions); errExec != nil {
					utils.Logger.Warning(
						fmt.Sprintf(
							"<%s> creating actions for ActionProfile with id: <%s:%s>, error: <%s>",
							utils.ActionS, args.Tenant, aPf.ID, errExec))
					partExec = true
					return
				}
			}
			sActs := newScheduledActs(aPf.Tenant, aPf.ID, acntID, ctx,
				&ActData{args.CGREvent.Event, args.Opts}, acts)
			if errExec = aS.asapExecuteActions(sActs); errExec != nil {
				utils.Logger.Warning(
					fmt.Sprintf(
						"<%s> executing ActionProfile with id: <%s:%s>, error: <%s>",
						utils.ActionS, sActs.tenant, sActs.apID, errExec))
				partExec = true
				return
			}
			return
		}
		if len(aPf.AccountIDs) == 0 { // no accounts, other acts
			actsExec(utils.EmptyString)
			continue
		}
		for acntID := range aPf.AccountIDs {
			actsExec(acntID)
		}
	}
	if partExec {
		return utils.ErrPartiallyExecuted
	}
	*rpl = utils.OK
	return
}
*/
