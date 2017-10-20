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

package engine

import (
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/cgrates/cgrates/utils"
)

type ActionTrigger struct {
	ID            string // original csv tag
	UniqueID      string // individual id
	ThresholdType string //*min_event_counter, *max_event_counter, *min_balance_counter, *max_balance_counter, *min_balance, *max_balance, *balance_expired
	// stats: *min_asr, *max_asr, *min_acd, *max_acd, *min_tcd, *max_tcd, *min_acc, *max_acc, *min_tcc, *max_tcc, *min_ddc, *max_ddc
	ThresholdValue float64
	Recurrent      bool          // reset excuted flag each run
	MinSleep       time.Duration // Minimum duration between two executions in case of recurrent triggers
	ExpirationDate time.Time
	ActivationDate time.Time
	//BalanceType       string // *monetary/*voice etc
	Balance           *BalanceFilter
	Weight            float64
	ActionsID         string
	MinQueuedItems    int // Trigger actions only if this number is hit (stats only)
	Executed          bool
	LastExecutionTime time.Time
}

func (at *ActionTrigger) Execute(ub *Account, sq *CDRStatsQueueTriggered) (err error) {
	// check for min sleep time
	if at.Recurrent && !at.LastExecutionTime.IsZero() && time.Since(at.LastExecutionTime) < at.MinSleep {
		return
	}
	at.LastExecutionTime = time.Now()
	if ub != nil && ub.Disabled {
		return fmt.Errorf("User %s is disabled and there are triggers in action!", ub.ID)
	}
	// does NOT need to Lock() because it is triggered from a method that took the Lock
	var aac Actions
	aac, err = dm.GetActions(at.ActionsID, false, utils.NonTransactional)
	if err != nil {
		utils.Logger.Err(fmt.Sprintf("Failed to get actions: %v", err))
		return
	}
	aac.Sort()
	at.Executed = true
	transactionFailed := false
	removeAccountActionFound := false
	for _, a := range aac {
		// check action filter
		if len(a.Filter) > 0 {
			matched, err := ub.matchActionFilter(a.Filter)
			if err != nil {
				return err
			}
			if !matched {
				continue
			}
		}
		if a.Balance == nil {
			a.Balance = &BalanceFilter{}
		}
		if a.ExpirationString != "" { // if it's *unlimited then it has to be zero time'
			if expDate, parseErr := utils.ParseDate(a.ExpirationString); parseErr == nil {
				a.Balance.ExpirationDate = &time.Time{}
				*a.Balance.ExpirationDate = expDate
			}
		}

		actionFunction, exists := getActionFunc(a.ActionType)
		if !exists {
			utils.Logger.Err(fmt.Sprintf("Function type %v not available, aborting execution!", a.ActionType))
			transactionFailed = false
			break
		}
		//go utils.Logger.Info(fmt.Sprintf("Executing %v, %v: %v", ub, sq, a))
		if err := actionFunction(ub, sq, a, aac); err != nil {
			utils.Logger.Err(fmt.Sprintf("Error executing action %s: %v!", a.ActionType, err))
			transactionFailed = false
			break
		}
		if a.ActionType == REMOVE_ACCOUNT {
			removeAccountActionFound = true
		}
	}
	if transactionFailed || at.Recurrent {
		at.Executed = false
	}
	if !transactionFailed && ub != nil && !removeAccountActionFound {
		Publish(CgrEvent{
			"EventName": utils.EVT_ACTION_TRIGGER_FIRED,
			"Uuid":      at.UniqueID,
			"Id":        at.ID,
			"ActionIds": at.ActionsID,
		})
		dm.DataDB().SetAccount(ub)
	}
	return
}

// returns true if the field of the action timing are equeal to the non empty
// fields of the action
func (at *ActionTrigger) Match(a *Action) bool {
	if a == nil || a.Balance == nil {
		return true
	}
	if a.Balance.Type != nil && a.Balance.GetType() != at.Balance.GetType() {
		return false
	}
	var thresholdType bool
	if a.ExtraParameters != "" {
		t := struct {
			GroupID       string
			UniqueID      string
			ThresholdType string
		}{}
		json.Unmarshal([]byte(a.ExtraParameters), &t)
		// check Ids first
		if t.GroupID != "" {
			return at.ID == t.GroupID
		}
		if t.UniqueID != "" {
			return at.UniqueID == t.UniqueID
		}
		thresholdType = t.ThresholdType == "" || at.ThresholdType == t.ThresholdType
	}

	return thresholdType && at.Balance.CreateBalance().MatchFilter(a.Balance, false)
}

func (at *ActionTrigger) CreateBalance() *Balance {
	b := at.Balance.CreateBalance()
	b.ID = at.UniqueID
	return b
}

// makes a shallow copy of the receiver
func (at *ActionTrigger) Clone() *ActionTrigger {
	clone := new(ActionTrigger)
	*clone = *at
	return clone
}

func (at *ActionTrigger) Equals(oat *ActionTrigger) bool {
	// ids only
	return at.ID == oat.ID && at.UniqueID == oat.UniqueID
}

func (at *ActionTrigger) IsActive(t time.Time) bool {
	return at.ActivationDate.IsZero() || t.After(at.ActivationDate)
}

func (at *ActionTrigger) IsExpired(t time.Time) bool {
	return !at.ExpirationDate.IsZero() && t.After(at.ExpirationDate)
}

// Structure to store actions according to weight
type ActionTriggers []*ActionTrigger

func (atpl ActionTriggers) Len() int {
	return len(atpl)
}

func (atpl ActionTriggers) Swap(i, j int) {
	atpl[i], atpl[j] = atpl[j], atpl[i]
}

//we need higher weights earlyer in the list
func (atpl ActionTriggers) Less(j, i int) bool {
	return atpl[i].Weight < atpl[j].Weight
}

func (atpl ActionTriggers) Sort() {
	sort.Sort(atpl)
}
