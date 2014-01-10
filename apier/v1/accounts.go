/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2013 ITsysCOM

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

package apier

import (
	"errors"
	"fmt"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/cgrates/engine"
	"time"
)

type AttrAcntAction struct {
	Tenant    string
	Account   string
	Direction string
}

type AccountActionTiming struct {
	Id              string    // The id to reference this particular ActionTiming
	ActionTimingsId string    // The id of the ActionTimings profile attached to the account
	ActionsId       string    // The id of actions which will be executed
	NextExecTime    time.Time // Next execution time
}

func (self *ApierV1) GetAccountActionTimings(attrs AttrAcntAction, reply *[]*AccountActionTiming) error {
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
			if utils.IsSliceMember(at.UserBalanceIds, utils.BalanceKey(attrs.Tenant, attrs.Account, attrs.Direction)) {
				accountATs = append(accountATs, &AccountActionTiming{Id: at.Id, ActionTimingsId: at.Tag, ActionsId: at.ActionsId, NextExecTime: at.GetNextStartTime()})
			}
		}
	}
	*reply = accountATs
	return nil
}

type AttrRemActionTiming struct {
	ActionTimingsId string // Id identifying the ActionTimings profile
	ActionTimingId  string // Internal CGR id identifying particular ActionTiming, *all for all user related ActionTimings to be canceled
	Tenant          string // Tenant he account belongs to
	Account         string // Account name
	Direction       string // Traffic direction
	ReloadScheduler bool  // If set it will reload the scheduler after adding
}

// Removes an ActionTimings or parts of it depending on filters being set
func (self *ApierV1) RemActionTiming(attrs AttrRemActionTiming, reply *string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"ActionTimingsId"}); len(missing) != 0 { // Only mandatory ActionTimingsId
		return fmt.Errorf("%s:%v", utils.ERR_MANDATORY_IE_MISSING, missing)
	}
	if len(attrs.Account) != 0 { // Presence of Account requires complete account details to be provided
		if missing := utils.MissingStructFields(&attrs, []string{"Tenant", "Account", "Direction"}); len(missing) != 0 {
			return fmt.Errorf("%s:%v", utils.ERR_MANDATORY_IE_MISSING, missing)
		}
	}
	_, err := engine.AccLock.Guard(engine.ACTION_TIMING_PREFIX, func() (float64, error) { // ToDo: Expand the scheduler to consider the locks also
		ats, err := self.AccountDb.GetActionTimings(attrs.ActionTimingsId)
		if err != nil {
			return 0, err
		} else if len(ats) == 0 {
			return 0, errors.New(utils.ERR_NOT_FOUND)
		}
		ats = engine.RemActionTiming(ats, attrs.ActionTimingId, utils.BalanceKey(attrs.Tenant, attrs.Account, attrs.Direction))
		if err := self.AccountDb.SetActionTimings(attrs.ActionTimingsId, ats); err != nil {
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
	if balance, err := self.AccountDb.GetUserBalance(utils.BalanceKey(attrs.Tenant, attrs.Account, attrs.Direction)); err != nil {
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
	balanceId := utils.BalanceKey(attrs.Tenant, attrs.Account, attrs.Direction)
	_, err := engine.AccLock.Guard(balanceId, func() (float64, error) {
		ub, err := self.AccountDb.GetUserBalance(balanceId)
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
		if err := self.AccountDb.SetUserBalance(ub); err != nil {
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
