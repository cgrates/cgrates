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

type Task struct {
	Uuid      string
	AccountID string
	ActionsID string
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

func (apl *ActionPlan) Clone() (interface{}, error) {
	cln := new(ActionPlan)
	if err := utils.Clone(*apl, cln); err != nil {
		return nil, err
	}
	return cln, nil
}

func (t *Task) Execute() error {
	return (&ActionTiming{
		Uuid:       t.Uuid,
		ActionsID:  t.ActionsID,
		accountIDs: utils.StringMap{t.AccountID: true},
	}).Execute(nil, nil)
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

// To be deleted after the above solution proves reliable
func (at *ActionTiming) GetNextStartTimeOld(now time.Time) (t time.Time) {
	if !at.stCache.IsZero() {
		return at.stCache
	}
	i := at.Timing
	if i == nil {
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
	y, m, d := now.Date()
	z, _ := now.Zone()
	if i.Timing.StartTime != utils.ASAP {
		l := fmt.Sprintf("%d-%d-%d %s %s", y, m, d, i.Timing.StartTime, z)
		var err error
		t, err = time.Parse(FORMAT, l)
		if err != nil {
			utils.Logger.Err(fmt.Sprintf("Cannot parse action plan's StartTime %v", l))
			at.stCache = t
			return
		}
		if now.After(t) || now.Equal(t) { // Set it to next day this time
			t = t.AddDate(0, 0, 1)
		}
	}
	// weekdays
	if i.Timing.WeekDays != nil && len(i.Timing.WeekDays) > 0 {
		i.Timing.WeekDays.Sort()
		if t.IsZero() {
			t = time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), now.Second(), 0, now.Location())
		}
		for j := 0; j < 8; j++ {
			n := t.AddDate(0, 0, j)
			for _, wd := range i.Timing.WeekDays {
				if n.Weekday() == wd && (n.Equal(now) || n.After(now)) {
					at.stCache = n
					t = n
					return
				}
			}
		}
	}
	// monthdays
	if i.Timing.MonthDays != nil && len(i.Timing.MonthDays) > 0 {
		i.Timing.MonthDays.Sort()
		year := t.Year()
		month := t
		x := sort.SearchInts(i.Timing.MonthDays, t.Day())
		d = i.Timing.MonthDays[0]
		if x < len(i.Timing.MonthDays) {
			if i.Timing.MonthDays[x] == t.Day() {
				if t.Equal(now) || t.After(now) {
					goto MONTHS
				}
				if x+1 < len(i.Timing.MonthDays) { // today was found in the list, jump to the next grater day
					d = i.Timing.MonthDays[x+1]
				} else { // jump to next month
					//not using now to make sure the next month has the the 1 date
					//(if today is 31) next month may not have it
					tmp := time.Date(year, month.Month(), 1, 0, 0, 0, 0, time.Local)
					month = tmp.AddDate(0, 1, 0)
				}
			} else { // today was not found in the list, x is the first greater day
				d = i.Timing.MonthDays[x]
			}
		}
		h, m, s := t.Clock()
		t = time.Date(month.Year(), month.Month(), d, h, m, s, 0, time.Local)
	}
MONTHS:
	if i.Timing.Months != nil && len(i.Timing.Months) > 0 {
		i.Timing.Months.Sort()
		year := t.Year()
		x := sort.Search(len(i.Timing.Months), func(x int) bool { return i.Timing.Months[x] >= t.Month() })
		m = i.Timing.Months[0]
		if x < len(i.Timing.Months) {
			if i.Timing.Months[x] == t.Month() {
				if t.Equal(now) || t.After(now) {
					goto YEARS
				}
				if x+1 < len(i.Timing.Months) { // this month was found in the list so jump to next available month
					m = i.Timing.Months[x+1]
					// reset the monthday
					t = time.Date(t.Year(), t.Month(), i.Timing.MonthDays[0], t.Hour(), t.Minute(), t.Second(), 0, t.Location())
				} else { // jump to next year
					//not using now to make sure the next year has the the 1 date
					//(if today is 31) next month may not have it
					tmp := time.Date(year, 1, 1, 0, 0, 0, 0, time.Local)
					year = tmp.AddDate(1, 0, 0).Year()
				}
			} else { // this month was not found in the list, x is the first greater month
				m = i.Timing.Months[x]
				// reset the monthday
				t = time.Date(t.Year(), t.Month(), i.Timing.MonthDays[0], t.Hour(), t.Minute(), t.Second(), 0, t.Location())
			}
		}
		h, min, s := t.Clock()
		t = time.Date(year, m, t.Day(), h, min, s, 0, time.Local)
	} else {
		if now.After(t) {
			t = t.AddDate(0, 1, 0)
		}
	}
YEARS:
	if i.Timing.Years != nil && len(i.Timing.Years) > 0 {
		i.Timing.Years.Sort()
		x := sort.Search(len(i.Timing.Years), func(x int) bool { return i.Timing.Years[x] >= t.Year() })
		y = i.Timing.Years[0]
		if x < len(i.Timing.Years) {
			if i.Timing.Years[x] == now.Year() {
				if t.Equal(now) || t.After(now) {
					h, m, s := t.Clock()
					t = time.Date(now.Year(), t.Month(), t.Day(), h, m, s, 0, time.Local)
					at.stCache = t
					return
				}
				if x+1 < len(i.Timing.Years) { // this year was found in the list so jump to next available year
					y = i.Timing.Years[x+1]
					// reset the month
					if i.Timing.Months != nil {
						t = time.Date(t.Year(), i.Timing.Months[0], t.Day(), t.Hour(), t.Minute(), t.Second(), 0, t.Location())
					}
					// reset the monthday
					t = time.Date(t.Year(), t.Month(), i.Timing.MonthDays[0], t.Hour(), t.Minute(), t.Second(), 0, t.Location())
				}
			} else { // this year was not found in the list, x is the first greater year
				y = i.Timing.Years[x]
				// reset the month/monthday
				t = time.Date(t.Year(), i.Timing.Months[0], i.Timing.MonthDays[0], t.Hour(), t.Minute(), t.Second(), 0, t.Location())
			}
		}
		h, min, s := t.Clock()
		t = time.Date(y, t.Month(), t.Day(), h, min, s, 0, time.Local)
	} else {
		if now.After(t) {
			t = t.AddDate(1, 0, 0)
		}
	}
	at.stCache = t
	return
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
	for accID, _ := range at.accountIDs {
		_, err = guardian.Guardian.Guard(func() (interface{}, error) {
			acc, err := dm.DataDB().GetAccount(accID)
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
					transactionFailed = true
					break
				}
				if err := actionFunction(acc, a, aac, at.ExtraData); err != nil {
					utils.Logger.Err(fmt.Sprintf("Error executing action %s: %v!", a.ActionType, err))
					transactionFailed = true
					if failedActions != nil {
						go func() { failedActions <- a }()
					}
					break
				}
				if successActions != nil {
					go func() { successActions <- a }()
				}
				if a.ActionType == REMOVE_ACCOUNT {
					removeAccountActionFound = true
				}
			}
			if !transactionFailed && !removeAccountActionFound {
				dm.DataDB().SetAccount(acc)
			}
			return 0, nil
		}, 0, accID)
	}
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
				if failedActions != nil {
					go func() { failedActions <- a }()
				}
				break
			}
			if err := actionFunction(nil, a, aac, at.ExtraData); err != nil {
				utils.Logger.Err(fmt.Sprintf("Error executing accountless action %s: %v!", a.ActionType, err))
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
	Publish(CgrEvent{
		"EventName": utils.EVT_ACTION_TIMING_FIRED,
		"Uuid":      at.Uuid,
		"Id":        at.actionPlanID,
		"ActionIds": at.ActionsID,
	})
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
