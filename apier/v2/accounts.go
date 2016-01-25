/*
Real-time Charging System for Telecom & ISP environments
Copyright (C) 2012-2015 ITsysCOM GmbH

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
	"fmt"
	"math"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func (self *ApierV2) GetAccounts(attr utils.AttrGetAccounts, reply *[]*engine.Account) error {
	if len(attr.Tenant) == 0 {
		return utils.NewErrMandatoryIeMissing("Tenanat")
	}
	var accountKeys []string
	var err error
	if len(attr.AccountIds) == 0 {
		if accountKeys, err = self.AccountDb.GetKeysForPrefix(utils.ACCOUNT_PREFIX + utils.ConcatenatedKey(attr.Tenant)); err != nil {
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
	var limitedAccounts []string
	if attr.Limit != 0 {
		max := math.Min(float64(attr.Offset+attr.Limit), float64(len(accountKeys)))
		limitedAccounts = accountKeys[attr.Offset:int(max)]
	} else {
		limitedAccounts = accountKeys[attr.Offset:]
	}
	retAccounts := make([]*engine.Account, 0)
	for _, acntKey := range limitedAccounts {
		if acnt, err := self.AccountDb.GetAccount(acntKey[len(utils.ACCOUNT_PREFIX):]); err != nil && err != utils.ErrNotFound { // Not found is not an error here
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
	account, err := self.AccountDb.GetAccount(tag)
	if err != nil {
		return err
	}

	*reply = *account
	return nil
}

type AttrSetAccount struct {
	Tenant                 string
	Account                string
	ActionPlanIds          string
	ActionPlansOverwrite   bool
	ActionTriggersIds      string
	ActionTriggerOverwrite bool
	AllowNegative          *bool
	Disabled               *bool
	ReloadScheduler        bool
}

func (self *ApierV2) SetAccount(attr AttrSetAccount, reply *string) error {
	if missing := utils.MissingStructFields(&attr, []string{"Tenant", "Account"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	var schedulerReloadNeeded = false
	accID := utils.AccountKey(attr.Tenant, attr.Account)
	var ub *engine.Account
	_, err := engine.Guardian.Guard(func() (interface{}, error) {
		if bal, _ := self.AccountDb.GetAccount(accID); bal != nil {
			ub = bal
		} else { // Not found in db, create it here
			ub = &engine.Account{
				Id: accID,
			}
		}
		if len(attr.ActionPlanIds) != 0 {
			_, err := engine.Guardian.Guard(func() (interface{}, error) {
				var ap *engine.ActionPlan
				var err error
				ap, err = self.RatingDb.GetActionPlan(attr.ActionPlanIds, false)
				if err != nil {
					return 0, err
				}
				if _, exists := ap.AccountIDs[accID]; !exists {
					if ap.AccountIDs == nil {
						ap.AccountIDs = make(utils.StringMap)
					}
					ap.AccountIDs[accID] = true
					schedulerReloadNeeded = true
					// create tasks
					for _, at := range ap.ActionTimings {
						if at.IsASAP() {
							t := &engine.Task{
								Uuid:      utils.GenUUID(),
								AccountID: accID,
								ActionsID: at.ActionsID,
							}
							if err = self.RatingDb.PushTask(t); err != nil {
								return 0, err
							}
						}
					}
					if err := self.RatingDb.SetActionPlan(attr.ActionPlanIds, ap); err != nil {
						return 0, err
					}
					// update cache
					self.RatingDb.CacheRatingPrefixValues(map[string][]string{utils.ACTION_PLAN_PREFIX: []string{utils.ACTION_PLAN_PREFIX + attr.ActionPlanIds}})
				}
				return 0, nil
			}, 0, utils.ACTION_PLAN_PREFIX)
			if err != nil {
				return 0, err
			}
		}

		if len(attr.ActionTriggersIds) != 0 {
			atrs, err := self.RatingDb.GetActionTriggers(attr.ActionTriggersIds)
			if err != nil {
				return 0, err
			}
			ub.ActionTriggers = atrs
			ub.InitCounters()
		}
		if attr.AllowNegative != nil {
			ub.AllowNegative = *attr.AllowNegative
		}
		if attr.Disabled != nil {
			ub.Disabled = *attr.Disabled
		}
		// All prepared, save account
		if err := self.AccountDb.SetAccount(ub); err != nil {
			return 0, err
		}
		return 0, nil
	}, 0, accID)
	if err != nil {
		return utils.NewErrServerError(err)
	}
	if attr.ReloadScheduler && schedulerReloadNeeded {
		// reload scheduler
		if self.Sched != nil {
			self.Sched.Reload(true)
		}
	}
	*reply = utils.OK // This will mark saving of the account, error still can show up in actionTimingsId
	return nil
}
