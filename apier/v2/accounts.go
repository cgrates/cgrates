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

package v2

import (
	"errors"
	"math"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/guardian"
	"github.com/cgrates/cgrates/utils"
)

func (apiv2 *APIerSv2) GetAccounts(attr *utils.AttrGetAccounts, reply *[]*engine.Account) error {
	tnt := attr.Tenant
	if tnt == utils.EmptyString {
		tnt = apiv2.Config.GeneralCfg().DefaultTenant
	}
	var accountKeys []string
	var err error
	if len(attr.AccountIDs) == 0 {
		if accountKeys, err = apiv2.DataManager.DataDB().GetKeysForPrefix(utils.AccountPrefix + tnt); err != nil {
			return err
		}
	} else {
		for _, acntID := range attr.AccountIDs {
			if len(acntID) == 0 { // Source of error returned from redis (key not found)
				continue
			}
			accountKeys = append(accountKeys, utils.AccountPrefix+utils.ConcatenatedKey(tnt, acntID))
		}
	}
	if len(accountKeys) == 0 {
		return nil
	}
	if attr.Offset > len(accountKeys) {
		attr.Offset = len(accountKeys)
	}
	if attr.Offset < 0 {
		attr.Offset = 0
	}
	var limitedAccounts []string
	if attr.Limit != 0 {
		max := math.Min(float64(attr.Offset+attr.Limit), float64(len(accountKeys)))
		limitedAccounts = accountKeys[attr.Offset:int(max)]
	} else {
		limitedAccounts = accountKeys[attr.Offset:]
	}
	retAccounts := make([]*engine.Account, 0)
	for _, acntKey := range limitedAccounts {
		if acnt, err := apiv2.DataManager.GetAccount(acntKey[len(utils.AccountPrefix):]); err != nil && err != utils.ErrNotFound { // Not found is not an error here
			return err
		} else if acnt != nil {
			if alNeg, has := attr.Filter[utils.AllowNegative]; has && alNeg != acnt.AllowNegative {
				continue
			}
			if dis, has := attr.Filter[utils.Disabled]; has && dis != acnt.Disabled {
				continue
			}
			retAccounts = append(retAccounts, acnt)
		}
	}
	*reply = retAccounts
	return nil
}

// GetAccountsCount sets in reply var the total number of accounts registered for the received tenant
// returns ErrNotFound in case of 0 accounts
func (apiv2 *APIerSv2) GetAccountsCount(attr *utils.AttrGetAccountsCount, reply *int) (err error) {
	tnt := attr.Tenant
	if tnt == utils.EmptyString {
		tnt = apiv2.Config.GeneralCfg().DefaultTenant
	}
	var accountKeys []string
	if accountKeys, err = apiv2.DataManager.DataDB().GetKeysForPrefix(utils.AccountPrefix + tnt); err != nil {
		return err
	}
	if len(accountKeys) == 0 {
		return utils.ErrNotFound
	}
	*reply = len(accountKeys)

	return
}

// Get balance
func (apiv2 *APIerSv2) GetAccount(attr *utils.AttrGetAccount, reply *engine.Account) error {
	tnt := attr.Tenant
	if tnt == utils.EmptyString {
		tnt = apiv2.Config.GeneralCfg().DefaultTenant
	}
	tag := utils.ConcatenatedKey(tnt, attr.Account)
	account, err := apiv2.DataManager.GetAccount(tag)
	if err != nil {
		return err
	}
	*reply = *account
	return nil
}

type AttrSetAccount struct {
	Tenant                 string
	Account                string
	ActionPlanIDs          []string
	ActionPlansOverwrite   bool
	ActionTriggerIDs       []string
	ActionTriggerOverwrite bool
	ExtraOptions           map[string]bool
	ReloadScheduler        bool
}

func (apiv2 *APIerSv2) SetAccount(attr *AttrSetAccount, reply *string) error {
	if missing := utils.MissingStructFields(attr, []string{utils.AccountField}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	tnt := attr.Tenant
	if tnt == utils.EmptyString {
		tnt = apiv2.Config.GeneralCfg().DefaultTenant
	}
	accID := utils.ConcatenatedKey(tnt, attr.Account)
	dirtyActionPlans := make(map[string]*engine.ActionPlan)
	var ub *engine.Account
	var schedNeedsReload bool
	err := guardian.Guardian.Guard(func() error {
		if bal, _ := apiv2.DataManager.GetAccount(accID); bal != nil {
			ub = bal
		} else { // Not found in db, create it here
			ub = &engine.Account{
				ID: accID,
			}
		}
		err := guardian.Guardian.Guard(func() error {
			acntAPids, err := apiv2.DataManager.GetAccountActionPlans(accID, true, true, utils.NonTransactional)
			if err != nil && err != utils.ErrNotFound {
				return err
			}
			if attr.ActionPlansOverwrite {
				// clean previous action plans
				nAcntAPids := make([]string, 0, len(acntAPids))
				for _, apID := range acntAPids {
					if utils.IsSliceMember(attr.ActionPlanIDs, apID) {
						nAcntAPids = append(nAcntAPids, apID)
						continue // not removing the ones where
					}
					ap, err := apiv2.DataManager.GetActionPlan(apID, true, true, utils.NonTransactional)
					if err != nil {
						return err
					}
					delete(ap.AccountIDs, accID)
					dirtyActionPlans[apID] = ap
				}
				acntAPids = nAcntAPids
			}
			for _, apID := range attr.ActionPlanIDs {
				ap, err := apiv2.DataManager.GetActionPlan(apID, true, true, utils.NonTransactional)
				if err != nil {
					return err
				}
				// create tasks
				var schedTasks int // keep count on the number of scheduled tasks so we can compare with actions needed
				for _, at := range ap.ActionTimings {
					if at.IsASAP() {
						t := &engine.Task{
							Uuid:      utils.GenUUID(),
							AccountID: accID,
							ActionsID: at.ActionsID,
						}
						if err = apiv2.DataManager.DataDB().PushTask(t); err != nil {
							return err
						}
						schedTasks++
					}
				}
				if schedTasks != 0 && !schedNeedsReload {
					schedNeedsReload = true
				}
				if schedTasks == len(ap.ActionTimings) || // scheduled all actions, no need to add account to AP
					utils.IsSliceMember(acntAPids, apID) {
					continue // No need to reschedule since already there
				}
				if ap.AccountIDs == nil {
					ap.AccountIDs = make(utils.StringMap)
				}
				ap.AccountIDs[accID] = true
				dirtyActionPlans[apID] = ap
				acntAPids = append(acntAPids, apID)
			}
			if len(dirtyActionPlans) != 0 && !schedNeedsReload {
				schedNeedsReload = true
			}
			apIDs := make([]string, 0, len(dirtyActionPlans))
			for actionPlanID, ap := range dirtyActionPlans {
				if err := apiv2.DataManager.SetActionPlan(actionPlanID, ap, true, utils.NonTransactional); err != nil {
					return err
				}
				apIDs = append(apIDs, actionPlanID)
			}
			if err := apiv2.DataManager.SetAccountActionPlans(accID, acntAPids, true); err != nil {
				return err
			}
			return apiv2.ConnMgr.Call(apiv2.Config.ApierCfg().CachesConns, nil,
				utils.CacheSv1ReloadCache, &utils.AttrReloadCacheWithAPIOpts{
					AccountActionPlanIDs: []string{accID},
					ActionPlanIDs:        apIDs,
				}, reply)
		}, config.CgrConfig().GeneralCfg().LockingTimeout, utils.ActionPlanPrefix)
		if err != nil {
			return err
		}

		if attr.ActionTriggerOverwrite {
			ub.ActionTriggers = make(engine.ActionTriggers, 0)
		}
		for _, actionTriggerID := range attr.ActionTriggerIDs {
			atrs, err := apiv2.DataManager.GetActionTriggers(actionTriggerID, false, utils.NonTransactional)
			if err != nil {
				return err
			}
			for _, at := range atrs {
				var found bool
				for _, existingAt := range ub.ActionTriggers {
					if existingAt.Equals(at) {
						found = true
						break
					}
				}
				if !found {
					ub.ActionTriggers = append(ub.ActionTriggers, at)
				}
			}
		}

		ub.InitCounters()
		if alNeg, has := attr.ExtraOptions[utils.AllowNegative]; has {
			ub.AllowNegative = alNeg
		}
		if dis, has := attr.ExtraOptions[utils.Disabled]; has {
			ub.Disabled = dis
		}
		// All prepared, save account
		return apiv2.DataManager.SetAccount(ub)
	}, config.CgrConfig().GeneralCfg().LockingTimeout, utils.AccountPrefix+accID)
	if err != nil {
		return utils.NewErrServerError(err)
	}
	if attr.ReloadScheduler && schedNeedsReload {
		sched := apiv2.SchedulerService.GetScheduler()
		if sched == nil {
			return errors.New(utils.SchedulerNotRunningCaps)
		}
		sched.Reload()
	}
	*reply = utils.OK // This will mark saving of the account, error still can show up in actionTimingsId
	return nil
}
