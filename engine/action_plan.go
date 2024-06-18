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
	ExtraData    any
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
func (apl *ActionPlan) Clone() (any, error) {
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

// getDayOrEndOfMonth returns the day if is a valid date relative to t1 month
func getDayOrEndOfMonth(day int, t1 time.Time) int {
	if lastDay := utils.GetEndOfMonth(t1).Day(); lastDay <= day { // clamp the day to last day of month in order to corectly compare the time
		day = lastDay
	}
	return day
}

func (at *ActionTiming) GetNextStartTime(t1 time.Time) (t time.Time) {
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
	at.stCache = cronexpr.MustParse(i.Timing.CronString()).Next(t1)
	if i.Timing.ID == utils.MetaMonthlyEstimated {
		// substract a month from at.stCache only if we skip 2 months
		// or we skip a month because mentioned MonthDay is after the last day of the current month
		if at.stCache.Month() == t1.Month()+2 ||
			(utils.GetEndOfMonth(t1).Day() < at.Timing.Timing.MonthDays[0] &&
				at.stCache.Month() == t1.Month()+1) {
			lastDay := utils.GetEndOfMonth(at.stCache).Day()
			// only change the time if the new one is after t1
			if tmp := at.stCache.AddDate(0, 0, -lastDay); tmp.After(t1) {
				at.stCache = tmp
			}
		}
	}
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
func (at *ActionTiming) Execute(fltrS *FilterS, originService string) (err error) {
	at.ResetStartTimeCache()
	acts, err := at.getActions()
	if err != nil {
		utils.Logger.Err(fmt.Sprintf("Failed to get actions for %s: %s", at.ActionsID, err))
		return
	}
	var partialyExecuted bool
	for accID := range at.accountIDs {
		_ = guardian.Guardian.Guard(func() error {
			acc, err := dm.GetAccount(accID)
			if err != nil { // create account
				if err != utils.ErrNotFound {
					utils.Logger.Warning(fmt.Sprintf("Could not get account id: %s. Skipping!", accID))
					return err
				}
				err = nil
				acc = &Account{
					ID: accID,
				}
			}
			transactionFailed := false
			removeAccountActionFound := false
			sharedData := NewSharedActionsData(acts)
			for i, act := range acts {
				// check action filter
				if len(act.Filters) > 0 {
					if pass, err := fltrS.Pass(utils.NewTenantID(accID).Tenant, act.Filters,
						utils.MapStorage{utils.MetaReq: acc}); err != nil {
						return err
					} else if !pass {
						continue
					}
				}
				if act.Balance == nil {
					act.Balance = &BalanceFilter{}
				}
				if act.ExpirationString != "" { // if it's *unlimited then it has to be zero time
					if expDate, parseErr := utils.ParseTimeDetectLayout(act.ExpirationString,
						config.CgrConfig().GeneralCfg().DefaultTimezone); parseErr == nil {
						act.Balance.ExpirationDate = &time.Time{}
						*act.Balance.ExpirationDate = expDate
					}
				}

				actionFunction, exists := getActionFunc(act.ActionType)
				if !exists {
					// do not allow the action plan to be rescheduled
					at.Timing = nil
					utils.Logger.Err(
						fmt.Sprintf("Function type %v not available, aborting execution!",
							act.ActionType))
					partialyExecuted = true
					transactionFailed = true
					break
				}
				sharedData.idx = i // set the current action index in shared data
				if err := actionFunction(acc, act, acts, fltrS, at.ExtraData, sharedData,
					newActionConnCfg(originService, act.ActionType, config.CgrConfig())); err != nil {
					utils.Logger.Err(
						fmt.Sprintf("Error executing action %s: %v!",
							act.ActionType, err))
					partialyExecuted = true
					transactionFailed = true
					break
				}
				if act.ActionType == utils.MetaRemoveAccount {
					removeAccountActionFound = true
				}
			}
			if !transactionFailed && !removeAccountActionFound {
				dm.SetAccount(acc)
			}
			return nil
		}, config.CgrConfig().GeneralCfg().LockingTimeout, utils.AccountPrefix+accID)
	}
	//reset the error in case that the account is not found
	err = nil
	if len(at.accountIDs) == 0 { // action timing executing without accounts
		for _, act := range acts {
			if expDate, parseErr := utils.ParseTimeDetectLayout(act.ExpirationString,
				config.CgrConfig().GeneralCfg().DefaultTimezone); (act.Balance == nil || act.Balance.EmptyExpirationDate()) &&
				parseErr == nil && !expDate.IsZero() {
				act.Balance.ExpirationDate = &time.Time{}
				*act.Balance.ExpirationDate = expDate
			}

			actionFunction, exists := getActionFunc(act.ActionType)
			if !exists {
				// do not allow the action plan to be rescheduled
				at.Timing = nil
				utils.Logger.Err(
					fmt.Sprintf("Function type %v not available, aborting execution!",
						act.ActionType))
				partialyExecuted = true
				break
			}
			if err := actionFunction(nil, act, acts, fltrS, at.ExtraData, SharedActionsData{},
				newActionConnCfg(originService, act.ActionType, config.CgrConfig())); err != nil {
				utils.Logger.Err(
					fmt.Sprintf("Error executing accountless action %s: %v!",
						act.ActionType, err))
				partialyExecuted = true
				break
			}
		}
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
	return at.Timing.Timing.StartTime == utils.MetaASAP
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
