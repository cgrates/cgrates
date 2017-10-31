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

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/guardian"
	"github.com/cgrates/cgrates/utils"
)

type AttrSetAccountActionTriggers struct {
	Tenant                string
	Account               string
	GroupID               *string
	UniqueID              *string
	ThresholdType         *string
	ThresholdValue        *float64
	Recurrent             *bool
	Executed              *bool
	MinSleep              *string
	ExpirationDate        *string
	ActivationDate        *string
	BalanceID             *string
	BalanceType           *string
	BalanceDirections     *[]string
	BalanceDestinationIds *[]string
	BalanceWeight         *float64
	BalanceExpirationDate *string
	BalanceTimingTags     *[]string
	BalanceRatingSubject  *string
	BalanceCategories     *[]string
	BalanceSharedGroups   *[]string
	BalanceBlocker        *bool
	BalanceDisabled       *bool
	MinQueuedItems        *int
	ActionsID             *string
}

func (attr *AttrSetAccountActionTriggers) UpdateActionTrigger(at *engine.ActionTrigger, timezone string) (updated bool, err error) {
	if at == nil {
		return false, errors.New("Empty ActionTrigger")
	}
	if at.ID == "" { // New AT, update it's data
		if missing := utils.MissingStructFields(attr, []string{"GroupID", "ThresholdType", "ThresholdValue"}); len(missing) != 0 {
			return false, utils.NewErrMandatoryIeMissing(missing...)
		}
		at.ID = *attr.GroupID
		if attr.UniqueID != nil {
			at.UniqueID = *attr.UniqueID
		}
	}
	if attr.GroupID != nil && *attr.GroupID != at.ID {
		return
	}
	if attr.UniqueID != nil && *attr.UniqueID != at.UniqueID {
		return
	}
	// at matches
	updated = true
	if attr.ThresholdType != nil {
		at.ThresholdType = *attr.ThresholdType
	}
	if attr.ThresholdValue != nil {
		at.ThresholdValue = *attr.ThresholdValue
	}
	if attr.Recurrent != nil {
		at.Recurrent = *attr.Recurrent
	}
	if attr.Executed != nil {
		at.Executed = *attr.Executed
	}
	if attr.MinSleep != nil {
		if at.MinSleep, err = utils.ParseDurationWithNanosecs(*attr.MinSleep); err != nil {
			return
		}
	}
	if attr.ExpirationDate != nil {
		if at.ExpirationDate, err = utils.ParseTimeDetectLayout(*attr.ExpirationDate, timezone); err != nil {
			return
		}
	}
	if attr.ActivationDate != nil {
		if at.ActivationDate, err = utils.ParseTimeDetectLayout(*attr.ActivationDate, timezone); err != nil {
			return
		}
	}
	if at.Balance == nil {
		at.Balance = &engine.BalanceFilter{}
	}
	if attr.BalanceID != nil {
		at.Balance.ID = attr.BalanceID
	}
	if attr.BalanceType != nil {
		at.Balance.Type = attr.BalanceType
	}
	if attr.BalanceDirections != nil {
		at.Balance.Directions = utils.StringMapPointer(utils.NewStringMap(*attr.BalanceDirections...))
	}
	if attr.BalanceDestinationIds != nil {
		at.Balance.DestinationIDs = utils.StringMapPointer(utils.NewStringMap(*attr.BalanceDestinationIds...))
	}
	if attr.BalanceWeight != nil {
		at.Balance.Weight = attr.BalanceWeight
	}
	if attr.BalanceExpirationDate != nil {
		balanceExpTime, err := utils.ParseDate(*attr.BalanceExpirationDate)
		if err != nil {
			return false, err
		}
		at.Balance.ExpirationDate = &balanceExpTime
	}
	if attr.BalanceTimingTags != nil {
		at.Balance.TimingIDs = utils.StringMapPointer(utils.NewStringMap(*attr.BalanceTimingTags...))
	}
	if attr.BalanceRatingSubject != nil {
		at.Balance.RatingSubject = attr.BalanceRatingSubject
	}
	if attr.BalanceCategories != nil {
		at.Balance.Categories = utils.StringMapPointer(utils.NewStringMap(*attr.BalanceCategories...))
	}
	if attr.BalanceSharedGroups != nil {
		at.Balance.SharedGroups = utils.StringMapPointer(utils.NewStringMap(*attr.BalanceSharedGroups...))
	}
	if attr.BalanceBlocker != nil {
		at.Balance.Blocker = attr.BalanceBlocker
	}
	if attr.BalanceDisabled != nil {
		at.Balance.Disabled = attr.BalanceDisabled
	}
	if attr.MinQueuedItems != nil {
		at.MinQueuedItems = *attr.MinQueuedItems
	}
	if attr.ActionsID != nil {
		at.ActionsID = *attr.ActionsID
	}
	return
}

// SetAccountActionTriggers Updates or Creates ActionTriggers for an Account
func (self *ApierV2) SetAccountActionTriggers(attr AttrSetAccountActionTriggers, reply *string) error {
	if missing := utils.MissingStructFields(&attr, []string{"Tenant", "Account"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	accID := utils.AccountKey(attr.Tenant, attr.Account)
	var account *engine.Account
	_, err := guardian.Guardian.Guard(func() (interface{}, error) {
		if acc, err := self.DataManager.DataDB().GetAccount(accID); err == nil {
			account = acc
		} else {
			return 0, err
		}
		var foundOne bool
		for _, at := range account.ActionTriggers {
			if updated, err := attr.UpdateActionTrigger(at, self.Config.DefaultTimezone); err != nil {
				return 0, err
			} else if updated && !foundOne {
				foundOne = true
			}
		}
		if !foundOne { // Did not find one to update, create a new AT
			at := new(engine.ActionTrigger)
			if updated, err := attr.UpdateActionTrigger(at, self.Config.DefaultTimezone); err != nil {
				return 0, err
			} else if updated { // Adding a new AT
				account.ActionTriggers = append(account.ActionTriggers, at)
			}
		}
		account.ExecuteActionTriggers(nil)
		if err := self.DataManager.DataDB().SetAccount(account); err != nil {
			return 0, err
		}
		return 0, nil
	}, 0, accID)
	if err != nil {
		*reply = err.Error()
		return err
	}
	*reply = utils.OK
	return nil
}
