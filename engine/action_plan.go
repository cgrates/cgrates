/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/

package engine

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
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

// CacheClone returns a clone of ActionPlan used by ltcache CacheCloner
func (apl *ActionPlan) CacheClone() any {
	return apl.Clone()
}

func (apl *ActionPlan) RemoveAccountID(accID string) (found bool) {
	if _, found = apl.AccountIDs[accID]; found {
		delete(apl.AccountIDs, accID)
	}
	return
}

// Clone clones *ActionPlan
func (apl *ActionPlan) Clone() *ActionPlan {
	if apl == nil {
		return nil
	}
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
	return cln
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

func (at *ActionTiming) GetNextStartTime(refTime time.Time) time.Time {
	if !at.stCache.IsZero() {
		return at.stCache
	}
	// Put action schedule time in stCache for 1 time actions
	if at.Timing != nil && at.Timing.Timing != nil &&
		strings.HasPrefix(at.Timing.Timing.StartTime, utils.PlusChar) {
		tmStrTmp, _ := time.ParseDuration(strings.TrimPrefix(
			at.Timing.Timing.StartTime, utils.PlusChar))
		at.stCache = time.Now().Add(tmStrTmp)
		return at.stCache
	}
	rateIvl := at.Timing
	if rateIvl == nil || rateIvl.Timing == nil {
		return time.Time{}
	}
	// Normalize
	if rateIvl.Timing.StartTime == "" {
		rateIvl.Timing.StartTime = "00:00:00"
	}
	if len(rateIvl.Timing.Years) > 0 && len(rateIvl.Timing.Months) == 0 {
		rateIvl.Timing.Months = append(rateIvl.Timing.Months, 1)
	}
	if len(rateIvl.Timing.Months) > 0 && len(rateIvl.Timing.MonthDays) == 0 {
		rateIvl.Timing.MonthDays = append(rateIvl.Timing.MonthDays, 1)
	}
	at.stCache = cronexpr.MustParse(rateIvl.Timing.CronString()).Next(refTime)
	if rateIvl.Timing.ID == utils.MetaMonthlyEstimated {
		// When target day doesn't exist in a month, fall back to that month's last day
		// instead of skipping to next occurrence.
		currentMonth := refTime.Month()
		targetMonthDay := rateIvl.Timing.MonthDays[0]
		oneMonthSkip := utils.GetEndOfMonth(refTime).Day() < targetMonthDay &&
			at.stCache.Month() == currentMonth+1
		twoMonthSkip := at.stCache.Month() == currentMonth+2
		if oneMonthSkip || twoMonthSkip {
			daysToSubtract := utils.GetEndOfMonth(at.stCache).Day()

			// When transitioning from Jan to Feb, subtract the
			// actual desired day instead of Mar's last day.
			// This fixes cases like:
			// - Jan 29 -> Mar 29 (should be Feb 28 in non-leap years)
			// - Jan 30 -> Mar 30 (should be Feb 28/29)
			if currentMonth == time.January {
				daysToSubtract = targetMonthDay
			}

			adjustedTime := at.stCache.AddDate(0, 0, -daysToSubtract)
			if adjustedTime.After(refTime) {
				at.stCache = adjustedTime
			}
		}
	}
	return at.stCache
}

func (at *ActionTiming) ResetStartTimeCache() {
	if at.Timing != nil && at.Timing.Timing != nil && strings.HasPrefix(at.Timing.Timing.StartTime,
		utils.PlusChar) {
		return // dont reset time for 1 time action plans starting with "+"
	}
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
			if accountActions[act.ActionType] {
				continue
			}
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

type AttrSetActionPlanAccounts struct {
	Id              string            // Profile id
	ActionPlan      []*AttrActionPlan // Set of actions this Actions profile will perform
	AccountIDs      []string          // Set of accounts which the actions can be performed on
	Overwrite       bool              // If previously defined, will be overwritten
	ReloadScheduler bool              // Enables automatic reload of the scheduler (eg: useful when adding a single action timing)
}

type AttrSetActionPlan struct {
	Id              string            // Profile id
	ActionPlan      []*AttrActionPlan // Set of actions this Actions profile will perform
	Overwrite       bool              // If previously defined, will be overwritten
	ReloadScheduler bool              // Enables automatic reload of the scheduler (eg: useful when adding a single action timing)
}

type AttrActionPlan struct {
	ActionsId string  // Actions id
	TimingID  string  // timingID is used to specify the ID of the timing for a corner case ( e.g. *monthly_estimated )
	Years     string  // semicolon separated list of years this timing is valid on, *any or empty supported
	Months    string  // semicolon separated list of months this timing is valid on, *any or empty supported
	MonthDays string  // semicolon separated list of month's days this timing is valid on, *any or empty supported
	WeekDays  string  // semicolon separated list of week day names this timing is valid on *any or empty supported
	Time      string  // String representing the time this timing starts on, *asap supported
	Weight    float64 // Binding's weight
}

func (attr *AttrActionPlan) GetRITiming(dm *DataManager) (timing *RITiming, err error) {
	if dfltTiming, isDefault := checkDefaultTiming(attr.Time); isDefault {
		return dfltTiming, nil
	}
	timing = new(RITiming)

	if attr.TimingID != utils.EmptyString &&
		!strings.HasPrefix(attr.TimingID, utils.Meta) { // in case of dynamic timing
		if dbTiming, err := dm.GetTiming(attr.TimingID, false, utils.NonTransactional); err != nil {
			if err != utils.ErrNotFound { // if not found let the user to populate all the timings values
				return nil, err
			}
		} else {
			timing.ID = dbTiming.ID
			timing.Years = dbTiming.Years
			timing.Months = dbTiming.Months
			timing.MonthDays = dbTiming.MonthDays
			timing.WeekDays = dbTiming.WeekDays
			timing.StartTime = dbTiming.StartTime
			timing.EndTime = dbTiming.EndTime
		}
	}
	timing.ID = attr.TimingID
	timing.Years.Parse(attr.Years, ";")
	timing.Months.Parse(attr.Months, ";")
	timing.MonthDays.Parse(attr.MonthDays, ";")
	timing.WeekDays.Parse(attr.WeekDays, ";")
	if !verifyFormat(attr.Time) {
		err = fmt.Errorf("%s:%s", utils.ErrUnsupportedFormat.Error(), attr.Time)
		return
	}
	timing.StartTime = attr.Time
	return
}

// checkDefaultTiming will check the tStr if it's of the the default timings ( the same as in TPReader )
// and will compute it properly
func checkDefaultTiming(tStr string) (rTm *RITiming, isDefault bool) {
	currentTime := time.Now()
	fmtTime := currentTime.Format("15:04:05")
	switch tStr {
	case utils.MetaEveryMinute:
		return &RITiming{
			ID:        utils.MetaEveryMinute,
			Years:     utils.Years{},
			Months:    utils.Months{},
			MonthDays: utils.MonthDays{},
			WeekDays:  utils.WeekDays{},
			StartTime: utils.ConcatenatedKey(utils.Meta, utils.Meta, strconv.Itoa(currentTime.Second())),
			EndTime:   "",
		}, true
	case utils.MetaHourly:
		return &RITiming{
			ID:        utils.MetaHourly,
			Years:     utils.Years{},
			Months:    utils.Months{},
			MonthDays: utils.MonthDays{},
			WeekDays:  utils.WeekDays{},
			StartTime: utils.ConcatenatedKey(utils.Meta, strconv.Itoa(currentTime.Minute()), strconv.Itoa(currentTime.Second())),
			EndTime:   "",
		}, true
	case utils.MetaDaily:
		return &RITiming{
			ID:        utils.MetaDaily,
			Years:     utils.Years{},
			Months:    utils.Months{},
			MonthDays: utils.MonthDays{},
			WeekDays:  utils.WeekDays{},
			StartTime: fmtTime,
			EndTime:   ""}, true
	case utils.MetaWeekly:
		return &RITiming{
			ID:        utils.MetaWeekly,
			Years:     utils.Years{},
			Months:    utils.Months{},
			MonthDays: utils.MonthDays{},
			WeekDays:  utils.WeekDays{currentTime.Weekday()},
			StartTime: fmtTime,
			EndTime:   "",
		}, true
	case utils.MetaMonthly:
		return &RITiming{
			ID:        utils.MetaMonthly,
			Years:     utils.Years{},
			Months:    utils.Months{},
			MonthDays: utils.MonthDays{currentTime.Day()},
			WeekDays:  utils.WeekDays{},
			StartTime: fmtTime,
			EndTime:   "",
		}, true
	case utils.MetaMonthlyEstimated:
		return &RITiming{
			ID:        utils.MetaMonthlyEstimated,
			Years:     utils.Years{},
			Months:    utils.Months{},
			MonthDays: utils.MonthDays{currentTime.Day()},
			WeekDays:  utils.WeekDays{},
			StartTime: fmtTime,
			EndTime:   "",
		}, true
	case utils.MetaMonthEnd:
		return &RITiming{
			ID:        utils.MetaMonthEnd,
			Years:     utils.Years{},
			Months:    utils.Months{},
			MonthDays: utils.MonthDays{-1},
			WeekDays:  utils.WeekDays{},
			StartTime: fmtTime,
			EndTime:   "",
		}, true
	case utils.MetaYearly:
		return &RITiming{
			ID:        utils.MetaYearly,
			Years:     utils.Years{},
			Months:    utils.Months{currentTime.Month()},
			MonthDays: utils.MonthDays{currentTime.Day()},
			WeekDays:  utils.WeekDays{},
			StartTime: fmtTime,
			EndTime:   "",
		}, true
	default:
		return nil, false
	}
}

func verifyFormat(tStr string) bool {
	if tStr == utils.EmptyString || tStr == utils.MetaASAP ||
		strings.HasPrefix(tStr, utils.PlusChar) {
		return true
	}

	if len(tStr) > 8 { // hh:mm:ss
		return false
	}
	if a := strings.Split(tStr, utils.InInFieldSep); len(a) != 3 {
		return false
	} else {
		if _, err := strconv.Atoi(a[0]); err != nil {
			return false
		} else if _, err := strconv.Atoi(a[1]); err != nil {
			return false
		} else if _, err := strconv.Atoi(a[2]); err != nil {
			return false
		}
	}
	return true
}
