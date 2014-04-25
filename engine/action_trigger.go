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

package engine

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/cgrates/cgrates/utils"
)

type ActionTrigger struct {
	Id             string // uniquely identify the trigger
	BalanceType    string
	Direction      string
	ThresholdType  string //*min_counter, *max_counter, *min_balance, *max_balance
	ThresholdValue float64
	DestinationId  string
	Weight         float64
	ActionsId      string
	Executed       bool
	Recurrent      bool
}

func (at *ActionTrigger) Execute(ub *Account) (err error) {
	if ub.Disabled {
		return fmt.Errorf("User %s is disabled", ub.Id)
	}
	// does NOT need to Lock() because it is triggered from a method that took the Lock
	var aac Actions
	aac, err = accountingStorage.GetActions(at.ActionsId, false)
	aac.Sort()
	if err != nil {
		Logger.Err(fmt.Sprintf("Failed to get actions: %v", err))
		return
	}
	at.Executed = true
	atLeastOneActionExecuted := false
	for _, a := range aac {
		if a.Balance == nil {
			a.Balance = &Balance{}
		}
		a.Balance.ExpirationDate, _ = utils.ParseDate(a.ExpirationString)
		actionFunction, exists := getActionFunc(a.ActionType)
		if !exists {
			Logger.Warning(fmt.Sprintf("Function type %v not available, aborting execution!", a.ActionType))
			return
		}
		go Logger.Info(fmt.Sprintf("Executing %v: %v", ub.Id, a))
		err = actionFunction(ub, a)
		if err == nil {
			atLeastOneActionExecuted = true
		}
	}
	if !atLeastOneActionExecuted || at.Recurrent {
		at.Executed = false
	}
	storageLogger.LogActionTrigger(ub.Id, RATER_SOURCE, at, aac)
	accountingStorage.SetAccount(ub)
	return
}

// returns true if the field of the action timing are equeal to the non empty
// fields of the action
func (at *ActionTrigger) Match(a *Action) bool {
	if a == nil {
		return true
	}
	id := a.BalanceType == "" || at.BalanceType == a.BalanceType
	direction := a.Direction == "" || at.Direction == a.Direction
	thresholdType, thresholdValue := true, true
	if a.ExtraParameters != "" {
		t := struct {
			ThresholdType  string
			ThresholdValue float64
		}{}
		json.Unmarshal([]byte(a.ExtraParameters), &t)
		thresholdType = t.ThresholdType == "" || at.ThresholdType == t.ThresholdType
		thresholdValue = t.ThresholdValue == 0 || at.ThresholdValue == t.ThresholdValue
	}
	return id && direction && thresholdType && thresholdValue
}

// Structure to store actions according to weight
type ActionTriggerPriotityList []*ActionTrigger

func (atpl ActionTriggerPriotityList) Len() int {
	return len(atpl)
}

func (atpl ActionTriggerPriotityList) Swap(i, j int) {
	atpl[i], atpl[j] = atpl[j], atpl[i]
}

func (atpl ActionTriggerPriotityList) Less(i, j int) bool {
	return atpl[i].Weight < atpl[j].Weight
}

func (atpl ActionTriggerPriotityList) Sort() {
	sort.Sort(atpl)
}
