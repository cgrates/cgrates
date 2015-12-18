/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2012-2015 ITsysCOM

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
	"regexp"
	"sort"
	"time"

	"github.com/cgrates/cgrates/utils"
)

type ActionTrigger struct {
	ID            string // original csv tag
	UniqueID      string // individual id
	ThresholdType string //*min_event_counter, *max_event_counter, *min_balance_counter, *max_balance_counter, *min_balance, *max_balance, *exp_balance
	// stats: *min_asr, *max_asr, *min_acd, *max_acd, *min_tcd, *max_tcd, *min_acc, *max_acc, *min_tcc, *max_tcc, *min_ddc, *max_ddc
	ThresholdValue        float64
	Recurrent             bool          // reset eexcuted flag each run
	MinSleep              time.Duration // Minimum duration between two executions in case of recurrent triggers
	BalanceId             string
	BalanceType           string          // *monetary/*voice etc
	BalanceDirections     utils.StringMap // filter for balance
	BalanceDestinationIds utils.StringMap // filter for balance
	BalanceWeight         float64         // filter for balance
	BalanceExpirationDate time.Time       // filter for balance
	BalanceTimingTags     utils.StringMap // filter for balance
	BalanceRatingSubject  string          // filter for balance
	BalanceCategories     utils.StringMap // filter for balance
	BalanceSharedGroups   utils.StringMap // filter for balance
	BalanceDisabled       bool            // filter for balance
	Weight                float64
	ActionsId             string
	MinQueuedItems        int // Trigger actions only if this number is hit (stats only)
	Executed              bool
	lastExecutionTime     time.Time
}

func (at *ActionTrigger) Execute(ub *Account, sq *StatsQueueTriggered) (err error) {
	// check for min sleep time
	if at.Recurrent && !at.lastExecutionTime.IsZero() && time.Since(at.lastExecutionTime) < at.MinSleep {
		return
	}
	at.lastExecutionTime = time.Now()
	if ub != nil && ub.Disabled {
		return fmt.Errorf("User %s is disabled and there are triggers in action!", ub.Id)
	}
	// does NOT need to Lock() because it is triggered from a method that took the Lock
	var aac Actions
	aac, err = ratingStorage.GetActions(at.ActionsId, false)
	aac.Sort()
	if err != nil {
		utils.Logger.Err(fmt.Sprintf("Failed to get actions: %v", err))
		return
	}
	at.Executed = true
	transactionFailed := false
	removeAccountActionFound := false
	for _, a := range aac {
		if a.Balance == nil {
			a.Balance = &Balance{}
		}
		a.Balance.ExpirationDate, _ = utils.ParseDate(a.ExpirationString)
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
		storageLogger.LogActionTrigger(ub.Id, utils.RATER_SOURCE, at, aac)
		accountingStorage.SetAccount(ub)
	}
	return
}

// returns true if the field of the action timing are equeal to the non empty
// fields of the action
func (at *ActionTrigger) Match(a *Action) bool {
	if a == nil {
		return true
	}
	// if we have Id than we can draw an early conclusion
	if a.Id != "" {
		match, _ := regexp.MatchString(a.Id, at.ID)
		return match
	}
	id := a.BalanceType == "" || at.BalanceType == a.BalanceType
	thresholdType, thresholdValue, direction, destinationId, weight, ratingSubject, categories, sharedGroup, timings, disabled := true, true, true, true, true, true, true, true, true, true
	if a.ExtraParameters != "" {
		t := struct {
			ThresholdType        string
			ThresholdValue       float64
			DestinationIds       string
			BalanceDirections    string
			BalanceWeight        float64
			BalanceRatingSubject string
			BalanceCategories    string
			BalanceSharedGroups  string
			BalanceTimingTags    string
			BalanceDisabled      bool
		}{}
		json.Unmarshal([]byte(a.ExtraParameters), &t)
		thresholdType = t.ThresholdType == "" || at.ThresholdType == t.ThresholdType
		thresholdValue = t.ThresholdValue == 0 || at.ThresholdValue == t.ThresholdValue
		direction = len(t.BalanceDirections) == 0 || at.BalanceDirections.Equal(utils.ParseStringMap(t.BalanceDirections))
		destinationId = len(t.DestinationIds) == 0 || at.BalanceDestinationIds.Equal(utils.ParseStringMap(t.DestinationIds))
		categories = len(t.BalanceCategories) == 0 || at.BalanceCategories.Equal(utils.ParseStringMap(t.BalanceCategories))
		timings = len(t.BalanceTimingTags) == 0 || at.BalanceTimingTags.Equal(utils.ParseStringMap(t.BalanceTimingTags))
		sharedGroup = len(t.BalanceSharedGroups) == 0 || at.BalanceSharedGroups.Equal(utils.ParseStringMap(t.BalanceSharedGroups))
		weight = t.BalanceWeight == 0 || at.BalanceWeight == t.BalanceWeight
		ratingSubject = t.BalanceRatingSubject == "" || at.BalanceRatingSubject == t.BalanceRatingSubject
		disabled = at.BalanceDisabled == t.BalanceDisabled
	}
	return id && direction && thresholdType && thresholdValue && destinationId && weight && ratingSubject && categories && sharedGroup && timings && disabled
}

// makes a shallow copy of the receiver
func (at *ActionTrigger) Clone() *ActionTrigger {
	clone := new(ActionTrigger)
	*clone = *at
	return clone
}

func (at *ActionTrigger) CreateBalance() *Balance {
	return &Balance{
		Id:             at.UniqueID,
		Directions:     at.BalanceDirections,
		ExpirationDate: at.BalanceExpirationDate,
		DestinationIds: at.BalanceDestinationIds,
		RatingSubject:  at.BalanceRatingSubject,
		Categories:     at.BalanceCategories,
		SharedGroups:   at.BalanceSharedGroups,
		TimingIDs:      at.BalanceTimingTags,
		Disabled:       at.BalanceDisabled,
		Weight:         at.BalanceWeight,
	}
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
