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
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/cgrates/cgrates/utils"
)

const (
	FORMAT     = "2006-1-2 15:04:05 MST"
	ASAP       = "*asap"
	ASAP_DELAY = "1m"
)

type ActionTiming struct {
	Id             string // uniquely identify the timing
	Tag            string // informative purpose only
	UserBalanceIds []string
	Timing         *RateInterval
	Weight         float64
	ActionsId      string
	actions        Actions
	stCache        time.Time // cached time of the next start
}

type ActionTimings []*ActionTiming

func (at *ActionTiming) GetNextStartTime() (t time.Time) {
	if !at.stCache.IsZero() {
		return at.stCache
	}
	i := at.Timing
	if i == nil {
		return
	}
	now := time.Now()
	y, m, d := now.Date()
	z, _ := now.Zone()
	if i.Timing.StartTime != "" && i.Timing.StartTime != ASAP {
		l := fmt.Sprintf("%d-%d-%d %s %s", y, m, d, i.Timing.StartTime, z)
		var err error
		t, err = time.Parse(FORMAT, l)
		if err != nil {
			Logger.Err(fmt.Sprintf("Cannot parse action timing's StartTime %v", l))
			at.stCache = t
			return
		}
	}
	// weekdays
	if i.Timing.WeekDays != nil && len(i.Timing.WeekDays) > 0 {
		i.Timing.WeekDays.Sort()
		if t.IsZero() {
			t = time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), now.Second(), 0, now.Location())
		}
		d := t.Day()
		for _, j := range []int{0, 1, 2, 3, 4, 5, 6, 7} {
			t = time.Date(t.Year(), t.Month(), d, t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), t.Location()).AddDate(0, 0, j)
			for _, wd := range i.Timing.WeekDays {
				if t.Weekday() == wd && (t.Equal(now) || t.After(now)) {
					at.stCache = t
					return
				}
			}
		}
	}
	// monthdays
	if i.Timing.MonthDays != nil && len(i.Timing.MonthDays) > 0 {
		i.Timing.MonthDays.Sort()
		now := time.Now()
		year := now.Year()
		month := now.Month()
		x := sort.SearchInts(i.Timing.MonthDays, now.Day())
		d = i.Timing.MonthDays[0]
		if x < len(i.Timing.MonthDays) {
			if i.Timing.MonthDays[x] == now.Day() {
				if t.Equal(now) || t.After(now) {
					h, m, s := t.Clock()
					t = time.Date(now.Year(), now.Month(), now.Day(), h, m, s, 0, time.Local)
					goto MONTHS
				}
				if x+1 < len(i.Timing.MonthDays) { // today was found in the list, jump to the next grater day
					d = i.Timing.MonthDays[x+1]
				} else { // jump to next month
					month = now.AddDate(0, 1, 0).Month()
				}
			} else { // today was not found in the list, x is the first greater day
				d = i.Timing.MonthDays[x]
			}
		} else {
			if len(i.Timing.Months) == 0 {
				t = time.Date(now.Year(), month, d, 0, 0, 0, 0, time.Local).AddDate(0, 1, 0)
				year = t.Year()
				month = t.Month()
			}
		}
		h, m, s := t.Clock()
		t = time.Date(year, month, d, h, m, s, 0, time.Local)
	}
MONTHS:
	if i.Timing.Months != nil && len(i.Timing.Months) > 0 {
		i.Timing.Months.Sort()
		now := time.Now()
		year := now.Year()
		x := sort.Search(len(i.Timing.Months), func(x int) bool { return i.Timing.Months[x] >= now.Month() })
		m = i.Timing.Months[0]
		if x < len(i.Timing.Months) {
			if i.Timing.Months[x] == now.Month() {
				if t.Equal(now) || t.After(now) {
					h, m, s := t.Clock()
					t = time.Date(now.Year(), now.Month(), t.Day(), h, m, s, 0, time.Local)
					goto YEARS
				}
				if x+1 < len(i.Timing.Months) { // this month was found in the list so jump to next available month
					m = i.Timing.Months[x+1]
					// reset the monthday
					if i.Timing.MonthDays != nil {
						t = time.Date(t.Year(), t.Month(), i.Timing.MonthDays[0], t.Hour(), t.Minute(), t.Second(), 0, t.Location())
					}
				} else { // jump to next year
					year = now.AddDate(1, 0, 0).Year()
				}
			} else { // this month was not found in the list, x is the first greater month
				m = i.Timing.Months[x]
				// reset the monthday
				if i.Timing.MonthDays != nil {
					t = time.Date(t.Year(), t.Month(), i.Timing.MonthDays[0], t.Hour(), t.Minute(), t.Second(), 0, t.Location())
				}
			}
		} else {
			if len(i.Timing.Years) == 0 {
				t = time.Date(year, m, t.Day(), 0, 0, 0, 0, time.Local).AddDate(1, 0, 0)
				year = t.Year()
			}
		}
		h, min, s := t.Clock()
		t = time.Date(year, m, t.Day(), h, min, s, 0, time.Local)
	}
YEARS:
	if i.Timing.Years != nil && len(i.Timing.Years) > 0 {
		i.Timing.Years.Sort()
		now := time.Now()
		x := sort.Search(len(i.Timing.Years), func(x int) bool { return i.Timing.Years[x] >= now.Year() })
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
					if i.Timing.MonthDays != nil {
						t = time.Date(t.Year(), t.Month(), i.Timing.MonthDays[0], t.Hour(), t.Minute(), t.Second(), 0, t.Location())
					}
				}
			} else { // this year was not found in the list, x is the first greater year
				y = i.Timing.Years[x]
				// reset the month
				if i.Timing.Months != nil {
					t = time.Date(t.Year(), i.Timing.Months[0], t.Day(), t.Hour(), t.Minute(), t.Second(), 0, t.Location())
				}
				// reset the monthday
				if i.Timing.MonthDays != nil {
					t = time.Date(t.Year(), t.Month(), i.Timing.MonthDays[0], t.Hour(), t.Minute(), t.Second(), 0, t.Location())
				}
			}
		}
		h, min, s := t.Clock()
		t = time.Date(y, t.Month(), t.Day(), h, min, s, 0, time.Local)
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
	at.resetStartTimeCache()
	aac, err := at.getActions()
	if err != nil {
		Logger.Err(fmt.Sprintf("Failed to get actions for %s: %s", at.ActionsId, err))
		return
	}
	for _, a := range aac {
		a.Balance.ExpirationDate, _ = utils.ParseDate(a.ExpirationString)
		actionFunction, exists := getActionFunc(a.ActionType)
		if !exists {
			Logger.Crit(fmt.Sprintf("Function type %v not available, aborting execution!", a.ActionType))
			return
		}
		for _, ubId := range at.UserBalanceIds {
			AccLock.Guard(ubId, func() (float64, error) {
				ub, err := accountingStorage.GetUserBalance(ubId)
				if err != nil {
					Logger.Warning(fmt.Sprintf("Could not get user balances for this id: %s. Skipping!", ubId))
					return 0, err
				}

				Logger.Info(fmt.Sprintf("Executing %v on %v", a.ActionType, ub.Id))
				err = actionFunction(ub, a)
				accountingStorage.SetUserBalance(ub)
				return 0, nil
			})
		}
	}
	storageLogger.LogActionTiming(SCHED_SOURCE, at, aac)
	return
}

// checks for *asap string as start time and replaces it wit an actual time in the newar future
// returns true if the *asap string was found
func (at *ActionTiming) CheckForASAP() bool {
	if at.Timing.Timing.StartTime == ASAP {
		delay, _ := time.ParseDuration(ASAP_DELAY)
		timeTokens := strings.Split(time.Now().Add(delay).Format(time.Stamp), " ")
		at.Timing.Timing.StartTime = timeTokens[len(timeTokens)-1]
		return true
	}
	return false
}

// returns true if only the starting time was is filled in the Timing field
func (at *ActionTiming) IsOneTimeRun() bool {
	return len(at.Timing.Timing.Years) == 0 &&
		len(at.Timing.Timing.Months) == 0 &&
		len(at.Timing.Timing.MonthDays) == 0 &&
		len(at.Timing.Timing.WeekDays) == 0 &&
		len(at.Timing.Timing.StartTime) != 0
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
	if atpl[i].GetNextStartTime().Equal(atpl[j].GetNextStartTime()) {
		return atpl[i].Weight < atpl[j].Weight
	}
	return atpl[i].GetNextStartTime().Before(atpl[j].GetNextStartTime())
}

func (atpl ActionTimingPriotityList) Sort() {
	sort.Sort(atpl)
}

func (at *ActionTiming) String_DISABLED() string {
	return at.Tag + " " + at.GetNextStartTime().String() + ",w: " + strconv.FormatFloat(at.Weight, 'f', -1, 64)
}
