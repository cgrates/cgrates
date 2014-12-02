/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2012-2014 ITsysCOM

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
	"strconv"
	"time"

	"github.com/cgrates/cgrates/utils"
	"github.com/gorhill/cronexpr"
)

const (
	FORMAT = "2006-1-2 15:04:05 MST"
	ASAP   = "*asap"
)

type ActionTiming struct {
	Uuid       string // uniquely identify the timing
	Id         string // informative purpose only
	AccountIds []string
	Timing     *RateInterval
	Weight     float64
	ActionsId  string
	actions    Actions
	stCache    time.Time // cached time of the next start
}

type ActionPlan []*ActionTiming

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
	if i.Timing.StartTime != ASAP {
		l := fmt.Sprintf("%d-%d-%d %s %s", y, m, d, i.Timing.StartTime, z)
		var err error
		t, err = time.Parse(FORMAT, l)
		if err != nil {
			Logger.Err(fmt.Sprintf("Cannot parse action timing's StartTime %v", l))
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

func (at *ActionTiming) resetStartTimeCache() {
	at.stCache = time.Date(1, 1, 1, 0, 0, 0, 0, time.UTC)
}

func (at *ActionTiming) SetActions(as Actions) {
	at.actions = as
}

func (at *ActionTiming) getActions() (as []*Action, err error) {
	if at.actions == nil {
		at.actions, err = accountingStorage.GetActions(at.ActionsId, false)
	}
	at.actions.Sort()
	return at.actions, err
}

func (at *ActionTiming) Execute() (err error) {
	if len(at.AccountIds) == 0 { // nothing to do if no accounts set
		return
	}
	at.resetStartTimeCache()
	aac, err := at.getActions()
	if err != nil {
		Logger.Err(fmt.Sprintf("Failed to get actions for %s: %s", at.ActionsId, err))
		return
	}
	for _, a := range aac {
		if expDate, parseErr := utils.ParseDate(a.ExpirationString); a.Balance.ExpirationDate.IsZero() && parseErr == nil && !expDate.IsZero() {
			a.Balance.ExpirationDate = expDate
		}
		actionFunction, exists := getActionFunc(a.ActionType)
		if !exists {
			// do not allow the action timing to be rescheduled
			at.Timing = nil
			Logger.Crit(fmt.Sprintf("Function type %v not available, aborting execution!", a.ActionType))
			return
		}
		for _, ubId := range at.AccountIds {
			_, err := AccLock.Guard(ubId, func() (float64, error) {
				ub, err := accountingStorage.GetAccount(ubId)
				if err != nil {
					Logger.Warning(fmt.Sprintf("Could not get user balances for this id: %s. Skipping!", ubId))
					return 0, err
				} else if ub.Disabled && a.ActionType != ENABLE_ACCOUNT {
					return 0, fmt.Errorf("Account %s is disabled", ubId)
				}
				//Logger.Info(fmt.Sprintf("Executing %v on %+v", a.ActionType, ub))
				err = actionFunction(ub, nil, a)
				//Logger.Info(fmt.Sprintf("After execute, account: %+v", ub))
				accountingStorage.SetAccount(ub)
				return 0, nil
			})
			if err != nil {
				Logger.Warning(fmt.Sprintf("Error executing action timing: %v", err))
			}
		}
	}
	storageLogger.LogActionTiming(SCHED_SOURCE, at, aac)
	return
}

func (at *ActionTiming) IsASAP() bool {
	return at.Timing.Timing.StartTime == ASAP
}

// Structure to store actions according to weight
type ActionTimingPriotityList []*ActionTiming

func (atpl ActionTimingPriotityList) Len() int {
	return len(atpl)
}

func (atpl ActionTimingPriotityList) Swap(i, j int) {
	atpl[i], atpl[j] = atpl[j], atpl[i]
}

func (atpl ActionTimingPriotityList) Less(i, j int) bool {
	if atpl[i].GetNextStartTime(time.Now()).Equal(atpl[j].GetNextStartTime(time.Now())) {
		return atpl[i].Weight < atpl[j].Weight
	}
	return atpl[i].GetNextStartTime(time.Now()).Before(atpl[j].GetNextStartTime(time.Now()))
}

func (atpl ActionTimingPriotityList) Sort() {
	sort.Sort(atpl)
}

func (at *ActionTiming) String_DISABLED() string {
	return at.Id + " " + at.GetNextStartTime(time.Now()).String() + ",w: " + strconv.FormatFloat(at.Weight, 'f', -1, 64)
}

// Helper to remove ActionTiming members based on specific filters, empty data means no always match
func RemActionTiming(ats ActionPlan, actionTimingId, balanceId string) ActionPlan {
	for idx, at := range ats {
		if len(actionTimingId) != 0 && at.Uuid != actionTimingId { // No Match for ActionTimingId, no need to move further
			continue
		}
		if len(balanceId) == 0 { // No account defined, considered match for complete removal
			if len(ats) == 1 { // Removing last item, by init empty
				return make([]*ActionTiming, 0)
			}
			ats[idx], ats = ats[len(ats)-1], ats[:len(ats)-1]
			continue
		}
		for iBlnc, blncId := range at.AccountIds {
			if blncId == balanceId {
				if len(at.AccountIds) == 1 { // Only one balance, remove complete at
					if len(ats) == 1 { // Removing last item, by init empty
						return make([]*ActionTiming, 0)
					}
					ats[idx], ats = ats[len(ats)-1], ats[:len(ats)-1]
				} else {
					at.AccountIds[iBlnc], at.AccountIds = at.AccountIds[len(at.AccountIds)-1], at.AccountIds[:len(at.AccountIds)-1]
				}
			}
		}
	}
	return ats
}
