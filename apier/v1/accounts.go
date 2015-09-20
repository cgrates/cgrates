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

package v1

import (
	"math"
	"regexp"
	"strings"
	"time"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

type AttrAcntAction struct {
	Tenant    string
	Account   string
	Direction string
}

type AccountActionTiming struct {
	ActionPlanId string    // The id of the ActionPlanId profile attached to the account
	Uuid         string    // The id to reference this particular ActionTiming
	ActionsId    string    // The id of actions which will be executed
	NextExecTime time.Time // Next execution time
}

func (self *ApierV1) GetAccountActionPlan(attrs AttrAcntAction, reply *[]*AccountActionTiming) error {
	if missing := utils.MissingStructFields(&attrs, []string{"Tenant", "Account", "Direction"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(strings.Join(missing, ","), "")
	}
	accountATs := make([]*AccountActionTiming, 0)
	allATs, err := self.RatingDb.GetAllActionPlans()
	if err != nil {
		return utils.NewErrServerError(err)
	}
	for _, ats := range allATs {
		for _, at := range ats {
			if utils.IsSliceMember(at.AccountIds, utils.AccountKey(attrs.Tenant, attrs.Account, attrs.Direction)) {
				accountATs = append(accountATs, &AccountActionTiming{Uuid: at.Uuid, ActionPlanId: at.Id, ActionsId: at.ActionsId, NextExecTime: at.GetNextStartTime(time.Now())})
			}
		}
	}
	*reply = accountATs
	return nil
}

type AttrRemActionTiming struct {
	ActionPlanId    string // Id identifying the ActionTimings profile
	ActionTimingId  string // Internal CGR id identifying particular ActionTiming, *all for all user related ActionTimings to be canceled
	Tenant          string // Tenant he account belongs to
	Account         string // Account name
	Direction       string // Traffic direction
	ReloadScheduler bool   // If set it will reload the scheduler after adding
}

// Removes an ActionTimings or parts of it depending on filters being set
func (self *ApierV1) RemActionTiming(attrs AttrRemActionTiming, reply *string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"ActionPlanId"}); len(missing) != 0 { // Only mandatory ActionPlanId
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if len(attrs.Account) != 0 { // Presence of Account requires complete account details to be provided
		if missing := utils.MissingStructFields(&attrs, []string{"Tenant", "Account", "Direction"}); len(missing) != 0 {
			return utils.NewErrMandatoryIeMissing(missing...)
		}
	}
	_, err := engine.Guardian.Guard(func() (interface{}, error) {
		ats, err := self.RatingDb.GetActionPlans(attrs.ActionPlanId)
		if err != nil {
			return 0, err
		} else if len(ats) == 0 {
			return 0, utils.ErrNotFound
		}
		ats = engine.RemActionPlan(ats, attrs.ActionTimingId, utils.AccountKey(attrs.Tenant, attrs.Account, attrs.Direction))
		if err := self.RatingDb.SetActionPlans(attrs.ActionPlanId, ats); err != nil {
			return 0, err
		}
		return 0, nil
	}, 0, utils.ACTION_TIMING_PREFIX)
	if err != nil {
		return utils.NewErrServerError(err)
	}
	if attrs.ReloadScheduler && self.Sched != nil {
		self.Sched.LoadActionPlans(self.RatingDb)
		self.Sched.Restart()
	}
	*reply = OK
	return nil
}

// Returns a list of ActionTriggers on an account
func (self *ApierV1) GetAccountActionTriggers(attrs AttrAcntAction, reply *engine.ActionTriggers) error {
	if missing := utils.MissingStructFields(&attrs, []string{"Tenant", "Account", "Direction"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if balance, err := self.AccountDb.GetAccount(utils.AccountKey(attrs.Tenant, attrs.Account, attrs.Direction)); err != nil {
		return utils.NewErrServerError(err)
	} else {
		*reply = balance.ActionTriggers
	}
	return nil
}

type AttrRemAcntActionTriggers struct {
	Tenant           string // Tenant he account belongs to
	Account          string // Account name
	Direction        string // Traffic direction
	ActionTriggersId string // Id filtering only specific id to remove (can be regexp pattern)
}

// Returns a list of ActionTriggers on an account
func (self *ApierV1) RemAccountActionTriggers(attrs AttrRemAcntActionTriggers, reply *string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"Tenant", "Account", "Direction"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	balanceId := utils.AccountKey(attrs.Tenant, attrs.Account, attrs.Direction)
	_, err := engine.Guardian.Guard(func() (interface{}, error) {
		ub, err := self.AccountDb.GetAccount(balanceId)
		if err != nil {
			return 0, err
		}
		nactrs := make(engine.ActionTriggers, 0)
		for _, actr := range ub.ActionTriggers {
			match, _ := regexp.MatchString(attrs.ActionTriggersId, actr.Id)
			if len(attrs.ActionTriggersId) != 0 && !match {
				nactrs = append(nactrs, actr)
			}
		}
		ub.ActionTriggers = nactrs
		if err := self.AccountDb.SetAccount(ub); err != nil {
			return 0, err
		}
		return 0, nil
	}, 0, balanceId)
	if err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = OK
	return nil
}

// Ads a new account into dataDb. If already defined, returns success.
func (self *ApierV1) SetAccount(attr utils.AttrSetAccount, reply *string) error {
	if missing := utils.MissingStructFields(&attr, []string{"Tenant", "Direction", "Account"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	balanceId := utils.AccountKey(attr.Tenant, attr.Account, attr.Direction)
	var ub *engine.Account
	var ats engine.ActionPlans
	_, err := engine.Guardian.Guard(func() (interface{}, error) {
		if bal, _ := self.AccountDb.GetAccount(balanceId); bal != nil {
			ub = bal
		} else { // Not found in db, create it here
			ub = &engine.Account{
				Id: balanceId,
			}
		}

		if len(attr.ActionPlanId) != 0 {
			var err error
			ats, err = self.RatingDb.GetActionPlans(attr.ActionPlanId)
			if err != nil {
				return 0, err
			}
			for _, at := range ats {
				at.AccountIds = append(at.AccountIds, balanceId)
			}
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
	}, 0, balanceId)
	if err != nil {
		return utils.NewErrServerError(err)
	}
	if len(ats) != 0 {
		_, err := engine.Guardian.Guard(func() (interface{}, error) { // ToDo: Try locking it above on read somehow
			if err := self.RatingDb.SetActionPlans(attr.ActionPlanId, ats); err != nil {
				return 0, err
			}
			return 0, nil
		}, 0, utils.ACTION_TIMING_PREFIX)
		if err != nil {
			return utils.NewErrServerError(err)
		}
		if self.Sched != nil {
			self.Sched.LoadActionPlans(self.RatingDb)
			self.Sched.Restart()
		}
	}
	*reply = OK // This will mark saving of the account, error still can show up in actionTimingsId
	return nil
}

func (self *ApierV1) RemoveAccount(attr utils.AttrRemoveAccount, reply *string) error {
	if missing := utils.MissingStructFields(&attr, []string{"Tenant", "Direction", "Account"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	accountId := utils.AccountKey(attr.Tenant, attr.Account, attr.Direction)
	_, err := engine.Guardian.Guard(func() (interface{}, error) {
		if err := self.AccountDb.RemoveAccount(accountId); err != nil {
			return 0, err
		}
		return 0, nil
	}, 0, accountId)
	if err != nil {
		return utils.NewErrServerError(err)
	}

	*reply = OK
	return nil
}

func (self *ApierV1) GetAccounts(attr utils.AttrGetAccounts, reply *[]*engine.Account) error {
	if len(attr.Tenant) == 0 {
		return utils.NewErrMandatoryIeMissing("Tenanat")
	}
	if len(attr.Direction) == 0 {
		attr.Direction = utils.OUT
	}
	var accountKeys []string
	var err error
	if len(attr.AccountIds) == 0 {
		if accountKeys, err = self.AccountDb.GetKeysForPrefix(utils.ACCOUNT_PREFIX + utils.ConcatenatedKey(attr.Direction, attr.Tenant)); err != nil {
			return err
		}
	} else {
		for _, acntId := range attr.AccountIds {
			if len(acntId) == 0 { // Source of error returned from redis (key not found)
				continue
			}
			accountKeys = append(accountKeys, utils.ACCOUNT_PREFIX+utils.ConcatenatedKey(attr.Direction, attr.Tenant, acntId))
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
