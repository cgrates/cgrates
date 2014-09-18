/*
Real-time Charging System for Telecom & ISP environments
Copyright (C) 2012-2014 ITsysCOM GmbH

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
	"errors"
	"fmt"
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
		return fmt.Errorf("%s:%v", utils.ERR_MANDATORY_IE_MISSING, missing)
	}
	accountATs := make([]*AccountActionTiming, 0)
	allATs, err := self.AccountDb.GetAllActionTimings()
	if err != nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
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
		return fmt.Errorf("%s:%v", utils.ERR_MANDATORY_IE_MISSING, missing)
	}
	if len(attrs.Account) != 0 { // Presence of Account requires complete account details to be provided
		if missing := utils.MissingStructFields(&attrs, []string{"Tenant", "Account", "Direction"}); len(missing) != 0 {
			return fmt.Errorf("%s:%v", utils.ERR_MANDATORY_IE_MISSING, missing)
		}
	}
	_, err := engine.AccLock.Guard(engine.ACTION_TIMING_PREFIX, func() (float64, error) {
		ats, err := self.AccountDb.GetActionTimings(attrs.ActionPlanId)
		if err != nil {
			return 0, err
		} else if len(ats) == 0 {
			return 0, errors.New(utils.ERR_NOT_FOUND)
		}
		ats = engine.RemActionTiming(ats, attrs.ActionTimingId, utils.AccountKey(attrs.Tenant, attrs.Account, attrs.Direction))
		if err := self.AccountDb.SetActionTimings(attrs.ActionPlanId, ats); err != nil {
			return 0, err
		}
		return 0, nil
	})
	if err != nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
	}
	if attrs.ReloadScheduler && self.Sched != nil {
		self.Sched.LoadActionTimings(self.AccountDb)
		self.Sched.Restart()
	}
	*reply = OK
	return nil
}

// Returns a list of ActionTriggers on an account
func (self *ApierV1) GetAccountActionTriggers(attrs AttrAcntAction, reply *engine.ActionTriggerPriotityList) error {
	if missing := utils.MissingStructFields(&attrs, []string{"Tenant", "Account", "Direction"}); len(missing) != 0 {
		return fmt.Errorf("%s:%v", utils.ERR_MANDATORY_IE_MISSING, missing)
	}
	if balance, err := self.AccountDb.GetAccount(utils.AccountKey(attrs.Tenant, attrs.Account, attrs.Direction)); err != nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
	} else {
		*reply = balance.ActionTriggers
	}
	return nil
}

type AttrRemAcntActionTriggers struct {
	Tenant          string // Tenant he account belongs to
	Account         string // Account name
	Direction       string // Traffic direction
	ActionTriggerId string // Id filtering only specific id to remove
}

// Returns a list of ActionTriggers on an account
func (self *ApierV1) RemAccountActionTriggers(attrs AttrRemAcntActionTriggers, reply *string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"Tenant", "Account", "Direction"}); len(missing) != 0 {
		return fmt.Errorf("%s:%v", utils.ERR_MANDATORY_IE_MISSING, missing)
	}
	balanceId := utils.AccountKey(attrs.Tenant, attrs.Account, attrs.Direction)
	_, err := engine.AccLock.Guard(balanceId, func() (float64, error) {
		ub, err := self.AccountDb.GetAccount(balanceId)
		if err != nil {
			return 0, err
		}
		for idx, actr := range ub.ActionTriggers {
			if len(attrs.ActionTriggerId) != 0 && actr.Id != attrs.ActionTriggerId { // Empty actionTriggerId will match always
				continue
			}
			if len(ub.ActionTriggers) != 1 { // Remove by index
				ub.ActionTriggers[idx], ub.ActionTriggers = ub.ActionTriggers[len(ub.ActionTriggers)-1], ub.ActionTriggers[:len(ub.ActionTriggers)-1]
			} else { // For last item, simply reinit the slice
				ub.ActionTriggers = make(engine.ActionTriggerPriotityList, 0)
			}
		}
		if err := self.AccountDb.SetAccount(ub); err != nil {
			return 0, err
		}
		return 0, nil
	})
	if err != nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
	}
	*reply = OK
	return nil
}

type AttrSetAccount struct {
	Tenant        string
	Direction     string
	Account       string
	ActionPlanId  string
	AllowNegative bool
}

// Ads a new account into dataDb. If already defined, returns success.
func (self *ApierV1) SetAccount(attr AttrSetAccount, reply *string) error {
	if missing := utils.MissingStructFields(&attr, []string{"Tenant", "Direction", "Account"}); len(missing) != 0 {
		return fmt.Errorf("%s:%v", utils.ERR_MANDATORY_IE_MISSING, missing)
	}
	balanceId := utils.AccountKey(attr.Tenant, attr.Account, attr.Direction)
	var ub *engine.Account
	var ats engine.ActionPlan
	_, err := engine.AccLock.Guard(balanceId, func() (float64, error) {
		if bal, _ := self.AccountDb.GetAccount(balanceId); bal != nil {
			ub = bal
		} else { // Not found in db, create it here
			ub = &engine.Account{
				Id:            balanceId,
				AllowNegative: attr.AllowNegative,
			}
		}

		if len(attr.ActionPlanId) != 0 {
			var err error
			ats, err = self.AccountDb.GetActionTimings(attr.ActionPlanId)
			if err != nil {
				return 0, err
			}
			for _, at := range ats {
				at.AccountIds = append(at.AccountIds, balanceId)
			}
		}
		// All prepared, save account
		if err := self.AccountDb.SetAccount(ub); err != nil {
			return 0, err
		}
		return 0, nil
	})
	if err != nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
	}
	if len(ats) != 0 {
		_, err := engine.AccLock.Guard(engine.ACTION_TIMING_PREFIX, func() (float64, error) { // ToDo: Try locking it above on read somehow
			if err := self.AccountDb.SetActionTimings(attr.ActionPlanId, ats); err != nil {
				return 0, err
			}
			return 0, nil
		})
		if err != nil {
			return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
		}
		if self.Sched != nil {
			self.Sched.LoadActionTimings(self.AccountDb)
			self.Sched.Restart()
		}
	}
	*reply = OK // This will mark saving of the account, error still can show up in actionTimingsId
	return nil
}
