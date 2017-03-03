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
	"fmt"
	"math"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/guardian"
	"github.com/cgrates/cgrates/utils"
)

func (self *ApierV2) GetAccounts(attr utils.AttrGetAccounts, reply *[]*engine.Account) error {
	if len(attr.Tenant) == 0 {
		return utils.NewErrMandatoryIeMissing("Tenant")
	}
	var accountKeys []string
	var err error
	if len(attr.AccountIds) == 0 {
		if accountKeys, err = self.DataDB.GetKeysForPrefix(utils.ACCOUNT_PREFIX + attr.Tenant); err != nil {
			return err
		}
	} else {
		for _, acntId := range attr.AccountIds {
			if len(acntId) == 0 { // Source of error returned from redis (key not found)
				continue
			}
			accountKeys = append(accountKeys, utils.ACCOUNT_PREFIX+utils.ConcatenatedKey(attr.Tenant, acntId))
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
		if acnt, err := self.DataDB.GetAccount(acntKey[len(utils.ACCOUNT_PREFIX):]); err != nil && err != utils.ErrNotFound { // Not found is not an error here
			return err
		} else if acnt != nil {
			retAccounts = append(retAccounts, acnt)
		}
	}
	*reply = retAccounts
	return nil
}

// Get balance
func (self *ApierV2) GetAccount(attr *utils.AttrGetAccount, reply *engine.Account) error {
	tag := fmt.Sprintf("%s:%s", attr.Tenant, attr.Account)
	account, err := self.DataDB.GetAccount(tag)
	if err != nil {
		return err
	}

	*reply = *account
	return nil
}

type AttrSetAccount struct {
	Tenant                 string
	Account                string
	ActionPlanIDs          *[]string
	ActionPlansOverwrite   bool
	ActionTriggerIDs       *[]string
	ActionTriggerOverwrite bool
	AllowNegative          *bool
	Disabled               *bool
	ReloadScheduler        bool
}

func (self *ApierV2) SetAccount(attr AttrSetAccount, reply *string) error {
	if missing := utils.MissingStructFields(&attr, []string{"Tenant", "Account"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	accID := utils.AccountKey(attr.Tenant, attr.Account)
	dirtyActionPlans := make(map[string]*engine.ActionPlan)
	var ub *engine.Account
	_, err := guardian.Guardian.Guard(func() (interface{}, error) {
		if bal, _ := self.DataDB.GetAccount(accID); bal != nil {
			ub = bal
		} else { // Not found in db, create it here
			ub = &engine.Account{
				ID: accID,
			}
		}
		if attr.ActionPlanIDs != nil {
			_, err := guardian.Guardian.Guard(func() (interface{}, error) {
				acntAPids, err := self.DataDB.GetAccountActionPlans(accID, false, utils.NonTransactional)
				if err != nil && err != utils.ErrNotFound {
					return 0, err
				}
				if attr.ActionPlansOverwrite {
					// clean previous action plans
					for i := 0; i < len(acntAPids); {
						apID := acntAPids[i]
						if utils.IsSliceMember(*attr.ActionPlanIDs, apID) {
							i++      // increase index since we don't remove from slice
							continue // not removing the ones where
						}
						ap, err := self.DataDB.GetActionPlan(apID, false, utils.NonTransactional)
						if err != nil {
							return 0, err
						}
						delete(ap.AccountIDs, accID)
						dirtyActionPlans[apID] = ap
						acntAPids = append(acntAPids[:i], acntAPids[i+1:]...) // remove the item from the list so we can overwrite the real list
					}
				}
				for _, apID := range *attr.ActionPlanIDs {
					if utils.IsSliceMember(acntAPids, apID) {
						continue // Already there
					}
					ap, err := self.DataDB.GetActionPlan(apID, false, utils.NonTransactional)
					if err != nil {
						return 0, err
					}
					if ap.AccountIDs == nil {
						ap.AccountIDs = make(utils.StringMap)
					}
					ap.AccountIDs[accID] = true
					dirtyActionPlans[apID] = ap
					acntAPids = append(acntAPids, apID)
					// create tasks
					for _, at := range ap.ActionTimings {
						if at.IsASAP() {
							t := &engine.Task{
								Uuid:      utils.GenUUID(),
								AccountID: accID,
								ActionsID: at.ActionsID,
							}
							if err = self.DataDB.PushTask(t); err != nil {
								return 0, err
							}
						}
					}
				}
				apIDs := make([]string, len(dirtyActionPlans))
				i := 0
				for actionPlanID, ap := range dirtyActionPlans {
					if err := self.DataDB.SetActionPlan(actionPlanID, ap, true, utils.NonTransactional); err != nil {
						return 0, err
					}
					apIDs[i] = actionPlanID
					i++
				}
				if err := self.DataDB.CacheDataFromDB(utils.ACTION_PLAN_PREFIX, apIDs, true); err != nil {
					return 0, err
				}
				if err := self.DataDB.SetAccountActionPlans(accID, acntAPids, true); err != nil {
					return 0, err
				}
				if err = self.DataDB.CacheDataFromDB(utils.AccountActionPlansPrefix, []string{accID}, true); err != nil {
					return 0, err
				}
				return 0, nil
			}, 0, utils.ACTION_PLAN_PREFIX)
			if err != nil {
				return 0, err
			}
		}

		if attr.ActionTriggerIDs != nil {
			if attr.ActionTriggerOverwrite {
				ub.ActionTriggers = make(engine.ActionTriggers, 0)
			}
			for _, actionTriggerID := range *attr.ActionTriggerIDs {
				atrs, err := self.DataDB.GetActionTriggers(actionTriggerID, false, utils.NonTransactional)
				if err != nil {
					return 0, err
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
		}
		ub.InitCounters()
		if attr.AllowNegative != nil {
			ub.AllowNegative = *attr.AllowNegative
		}
		if attr.Disabled != nil {
			ub.Disabled = *attr.Disabled
		}
		// All prepared, save account
		if err := self.DataDB.SetAccount(ub); err != nil {
			return 0, err
		}
		return 0, nil
	}, 0, accID)
	if err != nil {
		return utils.NewErrServerError(err)
	}
	if attr.ReloadScheduler && len(dirtyActionPlans) != 0 {
		sched := self.ServManager.GetScheduler()
		if sched == nil {
			return errors.New(utils.SchedulerNotRunningCaps)
		}
		sched.Reload()
	}
	*reply = utils.OK // This will mark saving of the account, error still can show up in actionTimingsId
	return nil
}
