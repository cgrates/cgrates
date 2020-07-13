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
	"fmt"
	"sort"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/guardian"
	"github.com/cgrates/cgrates/utils"
	"github.com/gorhill/cronexpr"
)

const (
	FORMAT = "2006-1-2 15:04:05 MST"
)

type ActionTiming struct {
	Uuid         string
	Timing       *RateInterval
	ActionsID    string
	ExtraData    interface{}
	Weight       float64
	actions      Actions
	accountIDs   utils.StringMap // copy of action plans accounts
	actionPlanID string          // the id of the belonging action plan (info only)
	stCache      time.Time       // cached time of the next start
}

// Tasks converts an ActionTiming into multiple Tasks
func (at *ActionTiming) Tasks() (tsks []*Task) {
	if len(at.accountIDs) == 0 {
		return []*Task{{
			Uuid:      at.Uuid,
			ActionsID: at.ActionsID,
		}}
	}
	tsks = make([]*Task, len(at.accountIDs))
	i := 0
	for acntID := range at.accountIDs {
		tsks[i] = &Task{
			Uuid:      at.Uuid,
			ActionsID: at.ActionsID,
			AccountID: acntID,
		}
		i++
	}
	return
}

type ActionPlan struct {
	Id            string // informative purpose only
	AccountIDs    utils.StringMap
	ActionTimings []*ActionTiming
}

func (apl *ActionPlan) RemoveAccountID(accID string) (found bool) {
	if _, found = apl.AccountIDs[accID]; found {
		delete(apl.AccountIDs, accID)
	}
	return
}

// Clone clones *ActionPlan
func (apl *ActionPlan) Clone() (interface{}, error) {
	cln := &ActionPlan{
		Id:         apl.Id,
		AccountIDs: apl.AccountIDs.Clone(),
	}
	if apl.ActionTimings != nil {
		cln.ActionTimings = make([]*ActionTiming, len(apl.ActionTimings))
		for i, act := range apl.ActionTimings {
			cln.ActionTimings[i] = act.Clone()
		}
	}
	return cln, nil
}

// Clone clones ActionTiming
func (at *ActionTiming) Clone() (cln *ActionTiming) {
	if at == nil {
		return
	}
	cln = &ActionTiming{
		Uuid:      at.Uuid,
		ActionsID: at.ActionsID,
		Weight:    at.Weight,
		ExtraData: at.ExtraData,
		Timing:    at.Timing.Clone(),
	}
	return
}

func (at *ActionTiming) GetNextStartTime(now time.Time) (t time.Time) {
	if !at.stCache.IsZero() {
		return at.stCache
	}
	i := at.Timing
	if i == nil || i.Timing == nil {
		return
	}
	// Normalize
	if i.Timing.StartTime == "" {
		i.Timing.StartTime = "00:00:00"
	}
	if len(i.Timing.Years) > 0 && len(i.Timing.Months) == 0 {
		i.Timing.Months = append(i.Timing.Months, 1)
	}
	if len(i.Timing.Months) > 0 && len(i.Timing.MonthDays) == 0 {
		i.Timing.MonthDays = append(i.Timing.MonthDays, 1)
	}
	at.stCache = cronexpr.MustParse(i.Timing.CronString()).Next(now)
	return at.stCache
}

func (at *ActionTiming) ResetStartTimeCache() {
	at.stCache = time.Date(1, 1, 1, 0, 0, 0, 0, time.UTC)
}

func (at *ActionTiming) SetActions(as Actions) {
	at.actions = as
}

func (at *ActionTiming) SetAccountIDs(accIDs utils.StringMap) {
	at.accountIDs = accIDs
}

func (at *ActionTiming) RemoveAccountID(acntID string) (found bool) {
	if _, found = at.accountIDs[acntID]; found {
		delete(at.accountIDs, acntID)
	}
	return
}

func (at *ActionTiming) GetAccountIDs() utils.StringMap {
	return at.accountIDs
}

func (at *ActionTiming) SetActionPlanID(id string) {
	at.actionPlanID = id
}

func (at *ActionTiming) GetActionPlanID() string {
	return at.actionPlanID
}

func (at *ActionTiming) getActions() (as []*Action, err error) {
	if at.actions == nil {
		at.actions, err = dm.GetActions(at.ActionsID, false, utils.NonTransactional)
	}
	at.actions.Sort()
	return at.actions, err
}

// Execute will execute all actions in an action plan
// Reports on success/fail via channel if != nil
func (at *ActionTiming) Execute(successActions, failedActions chan *Action) (err error) {
	at.ResetStartTimeCache()
	aac, err := at.getActions()
	if err != nil {
		utils.Logger.Err(fmt.Sprintf("Failed to get actions for %s: %s", at.ActionsID, err))
		return
	}
	var partialyExecuted bool
	for accID := range at.accountIDs {
		_, err = guardian.Guardian.Guard(func() (interface{}, error) {
			acc, err := dm.GetAccount(accID)
			if err != nil {
				utils.Logger.Warning(fmt.Sprintf("Could not get account id: %s. Skipping!", accID))
				return 0, err
			}
			transactionFailed := false
			removeAccountActionFound := false
			for _, a := range aac {
				// check action filter
				if len(a.Filter) > 0 {
					matched, err := acc.matchActionFilter(a.Filter)
					//log.Print("Checkng: ", a.Filter, matched)
					if err != nil {
						return 0, err
					}
					if !matched {
						continue
					}
				}
				if a.Balance == nil {
					a.Balance = &BalanceFilter{}
				}
				if a.ExpirationString != "" { // if it's *unlimited then it has to be zero time
					if expDate, parseErr := utils.ParseTimeDetectLayout(a.ExpirationString,
						config.CgrConfig().GeneralCfg().DefaultTimezone); parseErr == nil {
						a.Balance.ExpirationDate = &time.Time{}
						*a.Balance.ExpirationDate = expDate
					}
				}

				actionFunction, exists := getActionFunc(a.ActionType)
				if !exists {
					// do not allow the action plan to be rescheduled
					at.Timing = nil
					utils.Logger.Err(fmt.Sprintf("Function type %v not available, aborting execution!", a.ActionType))
					partialyExecuted = true
					transactionFailed = true
					break
				}
				if err := actionFunction(acc, a, aac, at.ExtraData); err != nil {
					utils.Logger.Err(fmt.Sprintf("Error executing action %s: %v!", a.ActionType, err))
					partialyExecuted = true
					transactionFailed = true
					if failedActions != nil {
						go func() { failedActions <- a }()
					}
					break
				}
				if successActions != nil {
					go func() { successActions <- a }()
				}
				if a.ActionType == utils.REMOVE_ACCOUNT {
					removeAccountActionFound = true
				}
			}
			if !transactionFailed && !removeAccountActionFound {
				dm.SetAccount(acc)
			}
			return 0, nil
		}, config.CgrConfig().GeneralCfg().LockingTimeout, utils.ACCOUNT_PREFIX+accID)
	}
	//reset the error in case that the account is not found
	err = nil
	if len(at.accountIDs) == 0 { // action timing executing without accounts
		for _, a := range aac {
			if expDate, parseErr := utils.ParseTimeDetectLayout(a.ExpirationString,
				config.CgrConfig().GeneralCfg().DefaultTimezone); (a.Balance == nil || a.Balance.EmptyExpirationDate()) &&
				parseErr == nil && !expDate.IsZero() {
				a.Balance.ExpirationDate = &time.Time{}
				*a.Balance.ExpirationDate = expDate
			}

			actionFunction, exists := getActionFunc(a.ActionType)
			if !exists {
				// do not allow the action plan to be rescheduled
				at.Timing = nil
				utils.Logger.Err(fmt.Sprintf("Function type %v not available, aborting execution!", a.ActionType))
				partialyExecuted = true
				if failedActions != nil {
					go func() { failedActions <- a }()
				}
				break
			}
			if err := actionFunction(nil, a, aac, at.ExtraData); err != nil {
				utils.Logger.Err(fmt.Sprintf("Error executing accountless action %s: %v!", a.ActionType, err))
				partialyExecuted = true
				if failedActions != nil {
					go func() { failedActions <- a }()
				}
				break
			}
			if successActions != nil {
				go func() { successActions <- a }()
			}
		}
	}
	if err != nil {
		utils.Logger.Warning(fmt.Sprintf("Error executing action plan: %v", err))
		return err
	}
	if partialyExecuted {
		return utils.ErrPartiallyExecuted
	}
	return
}

func (at *ActionTiming) IsASAP() bool {
	if at.Timing == nil {
		return false
	}
	return at.Timing.Timing.StartTime == utils.ASAP
}

// Structure to store actions according to execution time and weight
type ActionTimingPriorityList []*ActionTiming

func (atpl ActionTimingPriorityList) Len() int {
	return len(atpl)
}

func (atpl ActionTimingPriorityList) Swap(i, j int) {
	atpl[i], atpl[j] = atpl[j], atpl[i]
}

func (atpl ActionTimingPriorityList) Less(i, j int) bool {
	if atpl[i].GetNextStartTime(time.Now()).Equal(atpl[j].GetNextStartTime(time.Now())) {
		// higher weights earlyer in the list
		return atpl[i].Weight > atpl[j].Weight
	}
	return atpl[i].GetNextStartTime(time.Now()).Before(atpl[j].GetNextStartTime(time.Now()))
}

func (atpl ActionTimingPriorityList) Sort() {
	sort.Sort(atpl)
}

// Structure to store actions according to weight
type ActionTimingWeightOnlyPriorityList []*ActionTiming

func (atpl ActionTimingWeightOnlyPriorityList) Len() int {
	return len(atpl)
}

func (atpl ActionTimingWeightOnlyPriorityList) Swap(i, j int) {
	atpl[i], atpl[j] = atpl[j], atpl[i]
}

func (atpl ActionTimingWeightOnlyPriorityList) Less(i, j int) bool {
	return atpl[i].Weight > atpl[j].Weight
}

func (atpl ActionTimingWeightOnlyPriorityList) Sort() {
	sort.Sort(atpl)
}
