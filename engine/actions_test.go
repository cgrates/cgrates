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
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/ltcache"

	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

var (
	err error
	//referenceDate = time.Date(2013, 7, 10, 10, 30, 0, 0, time.Local)
	//referenceDate = time.Date(2013, 12, 31, 23, 59, 59, 0, time.Local)
	//referenceDate = time.Date(2011, 1, 1, 0, 0, 0, 1, time.Local)
	referenceDate = time.Now()
	now           = referenceDate
)

func TestActionTimingAlways(t *testing.T) {
	at := &ActionTiming{Timing: &RateInterval{Timing: &RITiming{StartTime: "00:00:00"}}}
	st := at.GetNextStartTime(referenceDate)
	y, m, d := referenceDate.Date()
	expected := time.Date(y, m, d, 0, 0, 0, 0, time.Local).AddDate(0, 0, 1)
	if !st.Equal(expected) {
		t.Errorf("Expected %v was %v", expected, st)
	}
}

func TestActionPlanNothing(t *testing.T) {
	at := &ActionTiming{}
	st := at.GetNextStartTime(referenceDate)
	expected := time.Date(1, 1, 1, 0, 0, 0, 0, time.UTC)
	if !st.Equal(expected) {
		t.Errorf("Expected %v was %v", expected, st)
	}
}

func TestActionTimingMidnight(t *testing.T) {
	at := &ActionTiming{Timing: &RateInterval{Timing: &RITiming{StartTime: "00:00:00"}}}
	y, m, d := referenceDate.Date()
	now := time.Date(y, m, d, 0, 0, 1, 0, time.Local)
	st := at.GetNextStartTime(now)
	expected := time.Date(y, m, d, 0, 0, 0, 0, time.Local).AddDate(0, 0, 1)
	if !st.Equal(expected) {
		t.Errorf("Expected %v was %v", expected, st)
	}
}

func TestActionPlanOnlyHour(t *testing.T) {
	at := &ActionTiming{Timing: &RateInterval{Timing: &RITiming{StartTime: "10:01:00"}}}
	st := at.GetNextStartTime(referenceDate)

	y, m, d := now.Date()
	expected := time.Date(y, m, d, 10, 1, 0, 0, time.Local)
	if referenceDate.After(expected) {
		expected = expected.AddDate(0, 0, 1)
	}
	if !st.Equal(expected) {
		t.Errorf("Expected %v was %v", expected, st)
	}
}

func TestActionPlanHourYear(t *testing.T) {
	at := &ActionTiming{Timing: &RateInterval{Timing: &RITiming{Years: utils.Years{2035}, StartTime: "10:01:00"}}}
	st := at.GetNextStartTime(referenceDate)
	expected := time.Date(2035, 1, 1, 10, 1, 0, 0, time.Local)
	if !st.Equal(expected) {
		t.Errorf("Expected %v was %v", expected, st)
	}
}

func TestActionPlanOnlyWeekdays(t *testing.T) {
	at := &ActionTiming{Timing: &RateInterval{Timing: &RITiming{WeekDays: []time.Weekday{time.Monday}}}}
	st := at.GetNextStartTime(referenceDate)

	y, m, d := now.Date()
	h, min, s := now.Clock()
	e := time.Date(y, m, d, h, min, s, 0, time.Local)
	day := e.Day()
	e = time.Date(e.Year(), e.Month(), day, 0, 0, 0, 0, e.Location())
	for i := 0; i < 8; i++ {
		n := e.AddDate(0, 0, i)
		if n.Weekday() == time.Monday && (n.Equal(now) || n.After(now)) {
			e = n
			break
		}
	}
	if !st.Equal(e) {
		t.Errorf("Expected %v was %v", e, st)
	}
}

func TestActionPlanHourWeekdays(t *testing.T) {
	at := &ActionTiming{Timing: &RateInterval{Timing: &RITiming{
		WeekDays: []time.Weekday{time.Monday}, StartTime: "10:01:00"}}}
	st := at.GetNextStartTime(referenceDate)

	y, m, d := now.Date()
	e := time.Date(y, m, d, 10, 1, 0, 0, time.Local)
	day := e.Day()
	for i := 0; i < 8; i++ {
		e = time.Date(e.Year(), e.Month(), day, e.Hour(),
			e.Minute(), e.Second(), e.Nanosecond(), e.Location())
		n := e.AddDate(0, 0, i)
		if n.Weekday() == time.Monday && (n.Equal(now) || n.After(now)) {
			e = n
			break
		}
	}
	if !st.Equal(e) {
		t.Errorf("Expected %v was %v", e, st)
	}
}

func TestActionPlanOnlyMonthdays(t *testing.T) {

	y, m, d := now.Date()
	tomorrow := time.Date(y, m, d, 0, 0, 0, 0, time.Local).AddDate(0, 0, 1)
	at := &ActionTiming{Timing: &RateInterval{
		Timing: &RITiming{MonthDays: utils.MonthDays{1, 25, 2, tomorrow.Day()}}}}
	st := at.GetNextStartTime(referenceDate)
	expected := tomorrow
	if !st.Equal(expected) {
		t.Errorf("Expected %v was %v", expected, st)
	}
}

func TestActionPlanHourMonthdays(t *testing.T) {

	y, m, d := now.Date()
	testTime := time.Date(y, m, d, 10, 1, 0, 0, time.Local)
	tomorrow := time.Date(y, m, d, 0, 0, 0, 0, time.Local).AddDate(0, 0, 1)
	if now.After(testTime) {
		y, m, d = tomorrow.Date()
	}
	at := &ActionTiming{Timing: &RateInterval{
		Timing: &RITiming{MonthDays: utils.MonthDays{now.Day(), tomorrow.Day()}, StartTime: "10:01:00"}}}
	st := at.GetNextStartTime(referenceDate)
	expected := time.Date(y, m, d, 10, 1, 0, 0, time.Local)
	if !st.Equal(expected) {
		t.Errorf("Expected %v was %v", expected, st)
	}
}

func TestActionPlanOnlyMonths(t *testing.T) {

	y, m, _ := now.Date()
	nextMonth := time.Date(y, m, 1, 0, 0, 0, 0, time.Local).AddDate(0, 1, 0)
	at := &ActionTiming{Timing: &RateInterval{
		Timing: &RITiming{Months: utils.Months{time.February, time.May, nextMonth.Month()}}}}
	st := at.GetNextStartTime(referenceDate)
	expected := time.Date(nextMonth.Year(), nextMonth.Month(), 1, 0, 0, 0, 0, time.Local)
	if !st.Equal(expected) {
		t.Log("NextMonth: ", nextMonth)
		t.Errorf("Expected %v was %v", expected, st)
	}
}

func TestActionPlanHourMonths(t *testing.T) {

	y, m, d := now.Date()
	testTime := time.Date(y, m, d, 10, 1, 0, 0, time.Local)
	nextMonth := time.Date(y, m, 1, 0, 0, 0, 0, time.Local).AddDate(0, 1, 0)
	if now.After(testTime) {
		testTime = testTime.AddDate(0, 0, 1)
		y, m, _ = testTime.Date()
	}
	if now.After(testTime) {
		m = nextMonth.Month()
		y = nextMonth.Year()

	}
	at := &ActionTiming{Timing: &RateInterval{Timing: &RITiming{
		Months:    utils.Months{now.Month(), nextMonth.Month()},
		StartTime: "10:01:00"}}}
	st := at.GetNextStartTime(referenceDate)
	expected := time.Date(y, m, 1, 10, 1, 0, 0, time.Local)
	if referenceDate.After(expected) {
		expected = expected.AddDate(0, 1, 0)
	}
	if !st.Equal(expected) {
		t.Errorf("Expected %v was %v", expected, st)
	}
}

func TestActionPlanHourMonthdaysMonths(t *testing.T) {

	y, m, d := now.Date()
	testTime := time.Date(y, m, d, 10, 1, 0, 0, time.Local)
	nextMonth := time.Date(y, m, 1, 0, 0, 0, 0, time.Local).AddDate(0, 1, 0)
	tomorrow := time.Date(y, m, d, 0, 0, 0, 0, time.Local).AddDate(0, 0, 1)

	if now.After(testTime) {
		y, m, d = tomorrow.Date()
	}
	nextDay := time.Date(y, m, d, 10, 1, 0, 0, time.Local)
	month := nextDay.Month()
	if nextDay.Before(now) {
		if now.After(testTime) {
			month = nextMonth.Month()
		}
	}
	at := &ActionTiming{Timing: &RateInterval{
		Timing: &RITiming{
			Months:    utils.Months{now.Month(), nextMonth.Month()},
			MonthDays: utils.MonthDays{now.Day(), tomorrow.Day()},
			StartTime: "10:01:00",
		},
	}}
	st := at.GetNextStartTime(referenceDate)
	expected := time.Date(y, month, d, 10, 1, 0, 0, time.Local)
	if !st.Equal(expected) {
		t.Errorf("Expected %v was %v", expected, st)
	}
}

func TestActionPlanFirstOfTheMonth(t *testing.T) {

	y, m, _ := now.Date()
	nextMonth := time.Date(y, m, 1, 0, 0, 0, 0, time.Local).AddDate(0, 1, 0)
	at := &ActionTiming{Timing: &RateInterval{
		Timing: &RITiming{
			MonthDays: utils.MonthDays{1},
		},
	}}
	st := at.GetNextStartTime(referenceDate)
	expected := nextMonth
	if !st.Equal(expected) {
		t.Errorf("Expected %v was %v", expected, st)
	}
}

func TestActionPlanOnlyYears(t *testing.T) {
	y, _, _ := referenceDate.Date()
	nextYear := time.Date(y, 1, 1, 0, 0, 0, 0, time.Local).AddDate(1, 0, 0)
	at := &ActionTiming{Timing: &RateInterval{
		Timing: &RITiming{Years: utils.Years{now.Year(), nextYear.Year()}}}}
	st := at.GetNextStartTime(referenceDate)
	expected := nextYear
	if !st.Equal(expected) {
		t.Errorf("Expected %v was %v", expected, st)
	}
}

func TestActionPlanPast(t *testing.T) {
	at := &ActionTiming{Timing: &RateInterval{Timing: &RITiming{Years: utils.Years{2028}}}}
	st := at.GetNextStartTime(referenceDate)
	expected := time.Date(2028, 1, 1, 0, 0, 0, 0, time.Local)
	if !st.Equal(expected) {
		t.Errorf("Expected %v was %v", expected, st)
	}
}

func TestActionPlanHourYears(t *testing.T) {
	at := &ActionTiming{Timing: &RateInterval{Timing: &RITiming{
		Years: utils.Years{referenceDate.Year(), referenceDate.Year() + 1}, StartTime: "10:01:00"}}}
	st := at.GetNextStartTime(referenceDate)
	expected := time.Date(referenceDate.Year(), 1, 1, 10, 1, 0, 0, time.Local)
	if referenceDate.After(expected) {
		expected = expected.AddDate(1, 0, 0)
	}
	if !st.Equal(expected) {
		t.Errorf("Expected %v was %v", expected, st)
	}
}

func TestActionPlanHourMonthdaysYear(t *testing.T) {

	y, m, d := now.Date()
	testTime := time.Date(y, m, d, 10, 1, 0, 0, time.Local)
	tomorrow := time.Date(y, m, d, 10, 1, 0, 0, time.Local).AddDate(0, 0, 1)
	nextYear := time.Date(y, 1, d, 10, 1, 0, 0, time.Local).AddDate(1, 0, 0)
	expected := testTime
	if referenceDate.After(testTime) {
		if referenceDate.After(tomorrow) {
			expected = nextYear
		} else {
			expected = tomorrow
		}
	}
	at := &ActionTiming{Timing: &RateInterval{
		Timing: &RITiming{
			Years:     utils.Years{now.Year(), nextYear.Year()},
			MonthDays: utils.MonthDays{now.Day(), tomorrow.Day()},
			StartTime: "10:01:00",
		},
	}}
	t.Log(at.Timing.Timing.CronString())
	t.Log(time.Now(), referenceDate, referenceDate.After(testTime), referenceDate.After(testTime))
	st := at.GetNextStartTime(referenceDate)
	if !st.Equal(expected) {
		t.Errorf("Expected %v was %v", expected, st)
	}
}

func TestActionPlanHourMonthdaysMonthYear(t *testing.T) {

	y, m, d := now.Date()
	testTime := time.Date(y, m, d, 10, 1, 0, 0, time.Local)
	nextYear := time.Date(y, m, 1, 0, 0, 0, 0, time.Local).AddDate(1, 0, 0)
	nextMonth := time.Date(y, m, 1, 0, 0, 0, 0, time.Local).AddDate(0, 1, 0)
	tomorrow := time.Date(y, m, d, 0, 0, 0, 0, time.Local).AddDate(0, 0, 1)
	day := now.Day()
	if now.After(testTime) {
		day = tomorrow.Day()
	}
	nextDay := time.Date(y, m, day, 10, 1, 0, 0, time.Local)
	month := now.Month()
	if nextDay.Before(now) {
		if now.After(testTime) {
			month = nextMonth.Month()
		}
	}
	nextDay = time.Date(y, month, day, 10, 1, 0, 0, time.Local)
	year := now.Year()
	if nextDay.Before(now) {
		if now.After(testTime) {
			year = nextYear.Year()
		}
	}
	at := &ActionTiming{Timing: &RateInterval{
		Timing: &RITiming{
			Years:     utils.Years{now.Year(), nextYear.Year()},
			Months:    utils.Months{now.Month(), nextMonth.Month()},
			MonthDays: utils.MonthDays{now.Day(), tomorrow.Day()},
			StartTime: "10:01:00",
		},
	}}
	st := at.GetNextStartTime(referenceDate)
	expected := time.Date(year, month, day, 10, 1, 0, 0, time.Local)
	if !st.Equal(expected) {
		t.Errorf("Expected %v was %v", expected, st)
	}
}

func TestActionPlanFirstOfTheYear(t *testing.T) {
	y, _, _ := now.Date()
	nextYear := time.Date(y, 1, 1, 0, 0, 0, 0, time.Local).AddDate(1, 0, 0)
	at := &ActionTiming{Timing: &RateInterval{
		Timing: &RITiming{
			Years:     utils.Years{nextYear.Year()},
			Months:    utils.Months{time.January},
			MonthDays: utils.MonthDays{1},
			StartTime: "00:00:00",
		},
	}}
	st := at.GetNextStartTime(referenceDate)
	expected := nextYear
	if !st.Equal(expected) {
		t.Errorf("Expected %v was %v", expected, st)
	}
}

func TestActionPlanFirstMonthOfTheYear(t *testing.T) {
	y, _, _ := now.Date()
	expected := time.Date(y, 1, 1, 0, 0, 0, 0, time.Local)
	if referenceDate.After(expected) {
		expected = expected.AddDate(1, 0, 0)
	}
	at := &ActionTiming{Timing: &RateInterval{
		Timing: &RITiming{
			Months: utils.Months{time.January},
		},
	}}
	st := at.GetNextStartTime(referenceDate)
	if !st.Equal(expected) {
		t.Errorf("Expected %v was %v", expected, st)
	}
}

func TestActionPlanFirstMonthOfTheYearSecondDay(t *testing.T) {
	y, _, _ := now.Date()
	expected := time.Date(y, 1, 2, 0, 0, 0, 0, time.Local)
	if referenceDate.After(expected) {
		expected = expected.AddDate(1, 0, 0)
	}
	at := &ActionTiming{Timing: &RateInterval{
		Timing: &RITiming{
			Months:    utils.Months{time.January},
			MonthDays: utils.MonthDays{2},
		},
	}}
	st := at.GetNextStartTime(referenceDate)
	if !st.Equal(expected) {
		t.Errorf("Expected %v was %v", expected, st)
	}
}

func TestActionPlanCheckForASAP(t *testing.T) {
	at := &ActionTiming{Timing: &RateInterval{Timing: &RITiming{StartTime: utils.MetaASAP}}}
	if !at.IsASAP() {
		t.Errorf("%v should be asap!", at)
	}
}

func TestActionPlanLogFunction(t *testing.T) {
	a := &Action{
		ActionType: "*log",
		Balance: &BalanceFilter{
			Type:  utils.StringPointer("test"),
			Value: &utils.ValueFormula{Static: 1.1},
		},
	}
	at := &ActionTiming{
		actions: []*Action{a},
	}
	err := at.Execute(nil, "")
	if err != nil {
		t.Errorf("Could not execute LOG action: %v", err)
	}
}

func TestActionPlanFunctionNotAvailable(t *testing.T) {
	a := &Action{
		ActionType: "VALID_FUNCTION_TYPE",
		Balance: &BalanceFilter{
			Type:  utils.StringPointer("test"),
			Value: &utils.ValueFormula{Static: 1.1},
		},
	}
	at := &ActionTiming{
		accountIDs: utils.StringMap{"cgrates.org:dy": true},
		Timing:     &RateInterval{},
		actions:    []*Action{a},
	}
	err := at.Execute(nil, "")
	if err != utils.ErrPartiallyExecuted { // because we want to return err if we can't execute all actions
		t.Errorf("Faild to detect wrong function type: %v", err)
	}
}

func TestActionTimingPriorityListSortByWeight(t *testing.T) {
	at1 := &ActionTiming{Timing: &RateInterval{
		Timing: &RITiming{
			Years: utils.Years{2040},
			Months: utils.Months{time.January, time.February, time.March,
				time.April, time.May, time.June, time.July, time.August, time.September,
				time.October, time.November, time.December},
			MonthDays: utils.MonthDays{1},
			StartTime: "00:00:00",
		},
		Weight: 20,
	}}
	at2 := &ActionTiming{Timing: &RateInterval{
		Timing: &RITiming{
			Years: utils.Years{2040},
			Months: utils.Months{time.January, time.February, time.March,
				time.April, time.May, time.June, time.July, time.August, time.September,
				time.October, time.November, time.December},
			MonthDays: utils.MonthDays{2},
			StartTime: "00:00:00",
		},
		Weight: 10,
	}}
	var atpl ActionTimingPriorityList
	atpl = append(atpl, at2, at1)
	atpl.Sort()
	if atpl[0] != at1 || atpl[1] != at2 {
		t.Errorf("Timing list not sorted correctly: \n %+v, \n %+v \n %+v",
			utils.ToJSON(at1), utils.ToJSON(at2), utils.ToJSON(atpl))
	}
}

func TestActionTimingPriorityListWeight(t *testing.T) {
	at1 := &ActionTiming{
		Timing: &RateInterval{
			Timing: &RITiming{
				Months: utils.Months{time.January, time.February, time.March,
					time.April, time.May, time.June, time.July, time.August, time.September,
					time.October, time.November, time.December},
				MonthDays: utils.MonthDays{1},
				StartTime: "00:00:00",
			},
		},
		Weight: 20,
	}
	at2 := &ActionTiming{
		Timing: &RateInterval{
			Timing: &RITiming{
				Months: utils.Months{time.January, time.February, time.March,
					time.April, time.May, time.June, time.July, time.August, time.September,
					time.October, time.November, time.December},
				MonthDays: utils.MonthDays{1},
				StartTime: "00:00:00",
			},
		},
		Weight: 10,
	}
	var atpl ActionTimingPriorityList
	atpl = append(atpl, at2, at1)
	atpl.Sort()
	if atpl[0] != at1 || atpl[1] != at2 {
		t.Error("Timing list not sorted correctly: ", atpl)
	}
}

func TestActionPlansRemoveMember(t *testing.T) {

	account1 := &Account{ID: "one"}
	account2 := &Account{ID: "two"}

	dm.SetAccount(account1)
	dm.SetAccount(account2)

	ap1 := &ActionPlan{
		Id:         "TestActionPlansRemoveMember1",
		AccountIDs: utils.StringMap{"one": true},
		ActionTimings: []*ActionTiming{
			{
				Uuid: "uuid1",
				Timing: &RateInterval{
					Timing: &RITiming{
						Years:     utils.Years{2012},
						Months:    utils.Months{},
						MonthDays: utils.MonthDays{},
						WeekDays:  utils.WeekDays{},
						StartTime: utils.MetaASAP,
					},
				},
				Weight:    10,
				ActionsID: "MINI",
			},
		},
	}

	ap2 := &ActionPlan{
		Id:         "test2",
		AccountIDs: utils.StringMap{"two": true},
		ActionTimings: []*ActionTiming{
			{
				Uuid: "uuid2",
				Timing: &RateInterval{
					Timing: &RITiming{
						Years:     utils.Years{2012},
						Months:    utils.Months{},
						MonthDays: utils.MonthDays{},
						WeekDays:  utils.WeekDays{},
						StartTime: utils.MetaASAP,
					},
				},
				Weight:    10,
				ActionsID: "MINI",
			},
		},
	}

	if err := dm.SetActionPlan(ap1.Id, ap1, true,
		utils.NonTransactional); err != nil {
		t.Error(err)
	}
	if err = dm.SetActionPlan(ap2.Id, ap2, true,
		utils.NonTransactional); err != nil {
		t.Error(err)
	}
	if err = dm.CacheDataFromDB(utils.ActionPlanPrefix,
		[]string{ap1.Id, ap2.Id}, true); err != nil {
		t.Error(err)
	}
	if err = dm.SetAccountActionPlans(account1.ID,
		[]string{ap1.Id}, false); err != nil {
		t.Error(err)
	}
	if err = dm.CacheDataFromDB(utils.AccountActionPlansPrefix,
		[]string{account1.ID}, true); err != nil {
		t.Error(err)
	}
	dm.GetAccountActionPlans(account1.ID, false, true, utils.NonTransactional) // FixMe: remove here after finishing testing of map
	if err = dm.SetAccountActionPlans(account2.ID,
		[]string{ap2.Id}, false); err != nil {
		t.Error(err)
	}
	if err = dm.CacheDataFromDB(utils.AccountActionPlansPrefix,
		[]string{account2.ID}, false); err != nil {
		t.Error(err)
	}

	actions := []*Action{
		{
			Id:         "REMOVE",
			ActionType: utils.MetaRemoveAccount,
		},
	}

	dm.SetActions(actions[0].Id, actions)

	at := &ActionTiming{
		accountIDs: utils.StringMap{account1.ID: true},
		Timing:     &RateInterval{},
		actions:    actions,
	}

	if err = at.Execute(nil, ""); err != nil {
		t.Errorf("Execute Action: %v", err)
	}

	apr, err1 := dm.GetActionPlan(ap1.Id, true, true, utils.NonTransactional)

	if err1 != nil {
		t.Errorf("Get action plan test: %v", err1)
	}

	if _, exist := apr.AccountIDs[account1.ID]; exist {
		t.Errorf("Account one is not deleted ")
	}

}

func TestActionTriggerMatchNil(t *testing.T) {
	at := &ActionTrigger{
		Balance: &BalanceFilter{
			Type: utils.StringPointer(utils.MetaMonetary),
		},
		ThresholdType:  utils.TriggerMaxBalance,
		ThresholdValue: 2,
	}
	var a *Action
	if !at.Match(a) {
		t.Errorf("Action trigger [%v] does not match action [%v]", at, a)
	}
}

func TestActionTriggerMatchAllBlank(t *testing.T) {
	at := &ActionTrigger{
		Balance: &BalanceFilter{
			Type: utils.StringPointer(utils.MetaMonetary),
		},
		ThresholdType:  utils.TriggerMaxBalance,
		ThresholdValue: 2,
	}
	a := &Action{}
	if !at.Match(a) {
		t.Errorf("Action trigger [%v] does not match action [%v]", at, a)
	}
}

func TestActionTriggerMatchMinuteBucketBlank(t *testing.T) {
	at := &ActionTrigger{
		Balance: &BalanceFilter{
			Type: utils.StringPointer(utils.MetaMonetary),
		},
		ThresholdType:  utils.TriggerMaxBalance,
		ThresholdValue: 2,
	}
	a := &Action{Balance: &BalanceFilter{Type: utils.StringPointer(utils.MetaMonetary)},
		ExtraParameters: `{"BalanceDirections":"*out"}`}
	if !at.Match(a) {
		t.Errorf("Action trigger [%v] does not match action [%v]", at, a)
	}
}

func TestActionTriggerMatchMinuteBucketFull(t *testing.T) {
	at := &ActionTrigger{
		Balance: &BalanceFilter{
			Type: utils.StringPointer(utils.MetaMonetary),
		},
		ThresholdType:  utils.TriggerMaxBalance,
		ThresholdValue: 2,
	}
	a := &Action{ExtraParameters: fmt.Sprintf(`{"ThresholdType":"%v", "ThresholdValue": %v}`,
		utils.TriggerMaxBalance, 2)}
	if !at.Match(a) {
		t.Errorf("Action trigger [%v] does not match action [%v]", at, a)
	}
}

func TestActionTriggerMatchAllFull(t *testing.T) {
	at := &ActionTrigger{
		Balance: &BalanceFilter{
			Type: utils.StringPointer(utils.MetaMonetary),
		},
		ThresholdType:  utils.TriggerMaxBalance,
		ThresholdValue: 2,
	}
	a := &Action{Balance: &BalanceFilter{Type: utils.StringPointer(utils.MetaMonetary)},
		ExtraParameters: fmt.Sprintf(`{"ThresholdType":"%v", "ThresholdValue": %v}`,
			utils.TriggerMaxBalance, 2)}
	if !at.Match(a) {
		t.Errorf("Action trigger [%v] does not match action [%v]", at, a)
	}
}

func TestActionTriggerMatchSomeFalse(t *testing.T) {
	at := &ActionTrigger{
		Balance: &BalanceFilter{
			Type: utils.StringPointer(utils.MetaMonetary),
		},
		ThresholdType:  utils.TriggerMaxBalance,
		ThresholdValue: 2,
	}
	a := &Action{Balance: &BalanceFilter{Type: utils.StringPointer(utils.MetaMonetary)},
		ExtraParameters: fmt.Sprintf(`{"ThresholdType":"%s"}`,
			utils.TriggerMaxBalanceCounter)}
	if at.Match(a) {
		t.Errorf("Action trigger [%v] does not match action [%v]", at, a)
	}
}

func TestActionTriggerMatcBalanceFalse(t *testing.T) {
	at := &ActionTrigger{
		Balance: &BalanceFilter{
			Type: utils.StringPointer(utils.MetaMonetary),
		},
		ThresholdType:  utils.TriggerMaxBalance,
		ThresholdValue: 2,
	}
	a := &Action{Balance: &BalanceFilter{Type: utils.StringPointer(utils.MetaMonetary)},
		ExtraParameters: fmt.Sprintf(`{"GroupID":"%s", "ThresholdType":"%s"}`, "TEST", utils.TriggerMaxBalance)}
	if at.Match(a) {
		t.Errorf("Action trigger [%v] does not match action [%v]", at, a)
	}
}

func TestActionTriggerMatcAllFalse(t *testing.T) {
	at := &ActionTrigger{
		Balance: &BalanceFilter{
			Type: utils.StringPointer(utils.MetaMonetary),
		},
		ThresholdType:  utils.TriggerMaxBalance,
		ThresholdValue: 2,
	}
	a := &Action{Balance: &BalanceFilter{Type: utils.StringPointer(utils.MetaMonetary)},
		ExtraParameters: fmt.Sprintf(`{"UniqueID":"ZIP", "GroupID":"%s", "ThresholdType":"%s"}`, "TEST",
			utils.TriggerMaxBalance)}
	if at.Match(a) {
		t.Errorf("Action trigger [%v] does not match action [%v]", at, a)
	}
}

func TestActionTriggerMatchAll(t *testing.T) {
	at := &ActionTrigger{
		ID:            "TEST",
		UniqueID:      "ZIP",
		ThresholdType: "TT",
		Balance: &BalanceFilter{
			Type:           utils.StringPointer(utils.MetaMonetary),
			RatingSubject:  utils.StringPointer("test1"),
			Value:          &utils.ValueFormula{Static: 2},
			Weight:         utils.Float64Pointer(1.0),
			DestinationIDs: utils.StringMapPointer(utils.NewStringMap("NAT")),
			SharedGroups:   utils.StringMapPointer(utils.NewStringMap("test2")),
		},
	}
	a := &Action{Balance: &BalanceFilter{
		Type:           utils.StringPointer(utils.MetaMonetary),
		RatingSubject:  utils.StringPointer("test1"),
		Value:          &utils.ValueFormula{Static: 2},
		Weight:         utils.Float64Pointer(1.0),
		DestinationIDs: utils.StringMapPointer(utils.NewStringMap("NAT")),
		SharedGroups:   utils.StringMapPointer(utils.NewStringMap("test2")),
	}, ExtraParameters: `{"UniqueID":"ZIP", "GroupID":"TEST", "ThresholdType":"TT"}`}
	if !at.Match(a) {
		t.Errorf("Action trigger [%v] does not match action [%v]", at, a)
	}
}

func TestActionTriggers(t *testing.T) {
	at1 := &ActionTrigger{Weight: 30}
	at2 := &ActionTrigger{Weight: 20}
	at3 := &ActionTrigger{Weight: 10}
	var atpl ActionTriggers
	atpl = append(atpl, at2, at1, at3)
	atpl.Sort()
	if atpl[0] != at1 || atpl[2] != at3 || atpl[1] != at2 {
		t.Error("List not sorted: ", atpl)
	}
}

func TestActionResetTriggres(t *testing.T) {
	ub := &Account{
		ID: "TEST_UB",
		BalanceMap: map[string]Balances{
			utils.MetaMonetary: {
				&Balance{Value: 10},
			},
			utils.MetaVoice: {
				&Balance{Value: 10, Weight: 20, DestinationIDs: utils.NewStringMap("NAT")},
				&Balance{Weight: 10, DestinationIDs: utils.NewStringMap("RET")},
			},
		},
		UnitCounters: UnitCounters{
			utils.MetaMonetary: []*UnitCounter{
				{
					Counters: CounterFilters{
						&CounterFilter{Value: 1},
					},
				},
			},
		},
		ActionTriggers: ActionTriggers{
			&ActionTrigger{
				Balance: &BalanceFilter{
					Type: utils.StringPointer(utils.MetaMonetary)},
				ThresholdValue: 2,
				ActionsID:      "TEST_ACTIONS",
				Executed:       true,
			},
			&ActionTrigger{
				Balance: &BalanceFilter{
					Type: utils.StringPointer(utils.MetaMonetary)},
				ThresholdValue: 2,
				ActionsID:      "TEST_ACTIONS",
				Executed:       true,
			},
		},
	}
	resetTriggersAction(ub, nil, nil, nil, nil, SharedActionsData{}, ActionConnCfg{})
	if ub.ActionTriggers[0].Executed == true || ub.ActionTriggers[1].Executed == true {
		t.Error("Reset triggers action failed!")
	}
}

func TestActionResetTriggresExecutesThem(t *testing.T) {
	ub := &Account{
		ID: "TEST_UB",
		BalanceMap: map[string]Balances{
			utils.MetaMonetary: {
				&Balance{Value: 10},
			},
		},
		UnitCounters: UnitCounters{
			utils.MetaMonetary: []*UnitCounter{
				{Counters: CounterFilters{&CounterFilter{Value: 1}}},
			},
		},
		ActionTriggers: ActionTriggers{
			&ActionTrigger{
				Balance: &BalanceFilter{
					Type: utils.StringPointer(utils.MetaMonetary)},
				ThresholdValue: 2,
				ActionsID:      "TEST_ACTIONS",
				Executed:       true,
			},
		},
	}
	resetTriggersAction(ub, nil, nil, nil, nil, SharedActionsData{}, ActionConnCfg{})
	if ub.ActionTriggers[0].Executed == true || ub.BalanceMap[utils.MetaMonetary][0].GetValue() == 12 {
		t.Error("Reset triggers action failed!")
	}
}

func TestActionResetTriggresActionFilter(t *testing.T) {
	ub := &Account{
		ID: "TEST_UB",
		BalanceMap: map[string]Balances{
			utils.MetaMonetary: {
				&Balance{Value: 10},
			},
			utils.MetaVoice: {
				&Balance{Value: 10, Weight: 20, DestinationIDs: utils.NewStringMap("NAT")},
				&Balance{Weight: 10, DestinationIDs: utils.NewStringMap("RET")},
			},
		},
		UnitCounters: UnitCounters{
			utils.MetaMonetary: []*UnitCounter{
				{Counters: CounterFilters{&CounterFilter{Value: 1}}},
			},
		},
		ActionTriggers: ActionTriggers{
			&ActionTrigger{
				Balance: &BalanceFilter{
					Type: utils.StringPointer(utils.MetaMonetary),
				},
				ThresholdValue: 2,
				ActionsID:      "TEST_ACTIONS",
				Executed:       true},
			&ActionTrigger{
				Balance: &BalanceFilter{
					Type: utils.StringPointer(utils.MetaMonetary),
				},
				ThresholdValue: 2,
				ActionsID:      "TEST_ACTIONS",
				Executed:       true}},
	}
	resetTriggersAction(ub, &Action{Balance: &BalanceFilter{Type: utils.StringPointer(utils.MetaSMS)}}, nil, nil, nil, SharedActionsData{}, ActionConnCfg{})
	if ub.ActionTriggers[0].Executed == false || ub.ActionTriggers[1].Executed == false {
		t.Error("Reset triggers action failed!")
	}
}

func TestActionResetTriggresActionFilter2(t *testing.T) {
	ub := &Account{
		ID: "TestActionResetTriggresActionFilter2",
		BalanceMap: map[string]Balances{
			utils.MetaMonetary: {
				&Balance{Value: 10},
			},
			utils.MetaVoice: {
				&Balance{Value: 10, Weight: 20, DestinationIDs: utils.NewStringMap("NAT")},
				&Balance{Weight: 10, DestinationIDs: utils.NewStringMap("RET")},
			},
		},
		UnitCounters: UnitCounters{
			utils.MetaMonetary: []*UnitCounter{
				{Counters: CounterFilters{&CounterFilter{Value: 1}}},
			},
		},
		ActionTriggers: ActionTriggers{
			&ActionTrigger{
				Balance: &BalanceFilter{
					Type: utils.StringPointer(utils.MetaMonetary),
				},
				ThresholdValue: 2,
				ActionsID:      "TEST_ACTIONS",
				Executed:       true},
			&ActionTrigger{
				Balance: &BalanceFilter{
					Type: utils.StringPointer(utils.MetaMonetary),
				},
				ThresholdValue: 2,
				ActionsID:      "TEST_ACTIONS",
				Executed:       true}},
	}
	resetTriggersAction(ub, &Action{}, nil, nil, nil, SharedActionsData{}, ActionConnCfg{})
	if ub.ActionTriggers[0].Executed != false && ub.ActionTriggers[1].Executed != false {
		t.Error("Reset triggers action failed!")
	}
}

func TestActionSetPostpaid(t *testing.T) {
	ub := &Account{
		ID: "TEST_UB",
		BalanceMap: map[string]Balances{
			utils.MetaMonetary: {&Balance{Value: 100}},
			utils.MetaVoice: {
				&Balance{Value: 10, Weight: 20, DestinationIDs: utils.NewStringMap("NAT")},
				&Balance{Weight: 10, DestinationIDs: utils.NewStringMap("RET")}}},
		UnitCounters: UnitCounters{utils.MetaMonetary: []*UnitCounter{
			{Counters: CounterFilters{&CounterFilter{Value: 1}}}}},
		ActionTriggers: ActionTriggers{
			&ActionTrigger{Balance: &BalanceFilter{Type: utils.StringPointer(utils.MetaMonetary)},
				ThresholdValue: 2, ActionsID: "TEST_ACTIONS", Executed: true},
			&ActionTrigger{Balance: &BalanceFilter{Type: utils.StringPointer(utils.MetaMonetary)},
				ThresholdValue: 2, ActionsID: "TEST_ACTIONS", Executed: true}},
	}
	allowNegativeAction(ub, nil, nil, nil, nil, SharedActionsData{}, ActionConnCfg{})
	if !ub.AllowNegative {
		t.Error("Set postpaid action failed!")
	}
}

func TestActionSetPrepaid(t *testing.T) {
	ub := &Account{
		ID:            "TEST_UB",
		AllowNegative: true,
		BalanceMap: map[string]Balances{
			utils.MetaMonetary: {&Balance{Value: 100}},
			utils.MetaVoice: {
				&Balance{Value: 10, Weight: 20, DestinationIDs: utils.NewStringMap("NAT")},
				&Balance{Weight: 10, DestinationIDs: utils.NewStringMap("RET")}}},
		UnitCounters: UnitCounters{utils.MetaMonetary: []*UnitCounter{
			{Counters: CounterFilters{&CounterFilter{Value: 1}}}}},
		ActionTriggers: ActionTriggers{
			&ActionTrigger{Balance: &BalanceFilter{Type: utils.StringPointer(utils.MetaMonetary)},
				ThresholdValue: 2, ActionsID: "TEST_ACTIONS", Executed: true},
			&ActionTrigger{Balance: &BalanceFilter{Type: utils.StringPointer(utils.MetaMonetary)},
				ThresholdValue: 2, ActionsID: "TEST_ACTIONS", Executed: true}},
	}
	denyNegativeAction(ub, nil, nil, nil, nil, SharedActionsData{}, ActionConnCfg{})
	if ub.AllowNegative {
		t.Error("Set prepaid action failed!")
	}
}

func TestActionResetPrepaid(t *testing.T) {
	ub := &Account{
		ID:            "TEST_UB",
		AllowNegative: true,
		BalanceMap: map[string]Balances{
			utils.MetaMonetary: {&Balance{Value: 100}},
			utils.MetaVoice: {
				&Balance{Value: 10, Weight: 20, DestinationIDs: utils.NewStringMap("NAT")},
				&Balance{Weight: 10, DestinationIDs: utils.NewStringMap("RET")}}},
		UnitCounters: UnitCounters{
			utils.MetaMonetary: []*UnitCounter{{Counters: CounterFilters{&CounterFilter{Value: 1}}}}},
		ActionTriggers: ActionTriggers{
			&ActionTrigger{Balance: &BalanceFilter{Type: utils.StringPointer(utils.MetaSMS)},
				ThresholdValue: 2, ActionsID: "TEST_ACTIONS", Executed: true},
			&ActionTrigger{Balance: &BalanceFilter{Type: utils.StringPointer(utils.MetaSMS)},
				ThresholdValue: 2, ActionsID: "TEST_ACTIONS", Executed: true}},
	}
	resetAccountAction(ub, nil, nil, nil, nil, SharedActionsData{}, ActionConnCfg{})
	if !ub.AllowNegative ||
		ub.BalanceMap[utils.MetaMonetary].GetTotalValue() != 0 ||
		len(ub.UnitCounters) != 0 ||
		ub.BalanceMap[utils.MetaVoice][0].GetValue() != 0 ||
		ub.ActionTriggers[0].Executed == true || ub.ActionTriggers[1].Executed == true {
		t.Log(ub.BalanceMap)
		t.Error("Reset account action failed!")
	}
}

func TestActionResetPostpaid(t *testing.T) {
	ub := &Account{
		ID: "TEST_UB",
		BalanceMap: map[string]Balances{
			utils.MetaMonetary: {&Balance{Value: 100}},
			utils.MetaVoice: {
				&Balance{Value: 10, Weight: 20, DestinationIDs: utils.NewStringMap("NAT")},
				&Balance{Weight: 10, DestinationIDs: utils.NewStringMap("RET")}}},
		UnitCounters: UnitCounters{
			utils.MetaMonetary: []*UnitCounter{{Counters: CounterFilters{&CounterFilter{Value: 1}}}}},
		ActionTriggers: ActionTriggers{
			&ActionTrigger{Balance: &BalanceFilter{Type: utils.StringPointer(utils.MetaSMS)},
				ThresholdValue: 2, ActionsID: "TEST_ACTIONS", Executed: true},
			&ActionTrigger{Balance: &BalanceFilter{Type: utils.StringPointer(utils.MetaSMS)},
				ThresholdValue: 2, ActionsID: "TEST_ACTIONS", Executed: true}},
	}
	resetAccountAction(ub, nil, nil, nil, nil, SharedActionsData{}, ActionConnCfg{})
	if ub.BalanceMap[utils.MetaMonetary].GetTotalValue() != 0 ||
		len(ub.UnitCounters) != 0 ||
		ub.BalanceMap[utils.MetaVoice][0].GetValue() != 0 ||
		ub.ActionTriggers[0].Executed == true || ub.ActionTriggers[1].Executed == true {
		t.Error("Reset account action failed!")
	}
}

func TestActionTopupResetCredit(t *testing.T) {
	ub := &Account{
		ID: "TEST_UB",
		BalanceMap: map[string]Balances{
			utils.MetaMonetary: {
				&Balance{Value: 100}},
			utils.MetaVoice: {
				&Balance{Value: 10, Weight: 20, DestinationIDs: utils.NewStringMap("NAT")},
				&Balance{Weight: 10, DestinationIDs: utils.NewStringMap("RET")}}},
		UnitCounters: UnitCounters{
			utils.MetaMonetary: []*UnitCounter{
				{Counters: CounterFilters{&CounterFilter{Value: 1}}}}},
		ActionTriggers: ActionTriggers{
			&ActionTrigger{Balance: &BalanceFilter{Type: utils.StringPointer(utils.MetaMonetary)},
				ThresholdValue: 2, ActionsID: "TEST_ACTIONS", Executed: true},
			&ActionTrigger{Balance: &BalanceFilter{Type: utils.StringPointer(utils.MetaMonetary)},
				ThresholdValue: 2, ActionsID: "TEST_ACTIONS", Executed: true}},
	}
	a := &Action{Balance: &BalanceFilter{Type: utils.StringPointer(utils.MetaMonetary),
		Value: &utils.ValueFormula{Static: 10}}}
	topupResetAction(ub, a, nil, nil, nil, SharedActionsData{}, ActionConnCfg{})
	if ub.AllowNegative ||
		ub.BalanceMap[utils.MetaMonetary].GetTotalValue() != 10 ||
		len(ub.UnitCounters) != 0 || // InitCounters finds no counters
		len(ub.BalanceMap[utils.MetaVoice]) != 2 ||
		ub.ActionTriggers[0].Executed != true || ub.ActionTriggers[1].Executed != true {
		t.Errorf("Topup reset action failed: %+s", utils.ToIJSON(ub))
	}
}

func TestActionTopupValueFactors(t *testing.T) {
	ub := &Account{
		ID:         "TEST_UB",
		BalanceMap: map[string]Balances{},
	}
	a := &Action{
		Balance: &BalanceFilter{
			Type:  utils.StringPointer(utils.MetaMonetary),
			Value: &utils.ValueFormula{Static: 10},
		},
		ExtraParameters: `{"*monetary":2}`,
	}
	topupResetAction(ub, a, nil, nil, nil, SharedActionsData{}, ActionConnCfg{})
	if len(ub.BalanceMap) != 1 ||
		ub.BalanceMap[utils.MetaMonetary][0].Factors[utils.MetaMonetary] != 2.0 {
		t.Errorf("Topup reset action failed to set Factors: %+v",
			ub.BalanceMap[utils.MetaMonetary][0].Factors)
	}
}

func TestActionTopupResetCreditId(t *testing.T) {
	ub := &Account{
		ID: "TEST_UB",
		BalanceMap: map[string]Balances{
			utils.MetaMonetary: {
				&Balance{Value: 100},
				&Balance{ID: "TEST_B", Value: 15},
			},
		},
	}
	a := &Action{Balance: &BalanceFilter{
		Type:  utils.StringPointer(utils.MetaMonetary),
		ID:    utils.StringPointer("TEST_B"),
		Value: &utils.ValueFormula{Static: 10}}}
	topupResetAction(ub, a, nil, nil, nil, SharedActionsData{}, ActionConnCfg{})
	if ub.AllowNegative ||
		ub.BalanceMap[utils.MetaMonetary].GetTotalValue() != 110 ||
		len(ub.BalanceMap[utils.MetaMonetary]) != 2 {
		t.Errorf("Topup reset action failed: %+v",
			ub.BalanceMap[utils.MetaMonetary][0])
	}
}

func TestActionTopupResetCreditNoId(t *testing.T) {
	ub := &Account{
		ID: "TEST_UB",
		BalanceMap: map[string]Balances{
			utils.MetaMonetary: {
				&Balance{Value: 100},
				&Balance{ID: "TEST_B", Value: 15},
			},
		},
	}
	a := &Action{Balance: &BalanceFilter{
		Type:  utils.StringPointer(utils.MetaMonetary),
		Value: &utils.ValueFormula{Static: 10}}}
	topupResetAction(ub, a, nil, nil, nil, SharedActionsData{}, ActionConnCfg{})
	if ub.AllowNegative ||
		ub.BalanceMap[utils.MetaMonetary].GetTotalValue() != 20 ||
		len(ub.BalanceMap[utils.MetaMonetary]) != 2 {
		t.Errorf("Topup reset action failed: %+v", ub.BalanceMap[utils.MetaMonetary][1])
	}
}

func TestActionTopupResetMinutes(t *testing.T) {
	ub := &Account{
		ID: "TEST_UB",
		BalanceMap: map[string]Balances{
			utils.MetaMonetary: {&Balance{Value: 100}},
			utils.MetaVoice: {&Balance{Value: 10, Weight: 20,
				DestinationIDs: utils.NewStringMap("NAT")},
				&Balance{Weight: 10, DestinationIDs: utils.NewStringMap("RET")}}},
		UnitCounters: UnitCounters{utils.MetaMonetary: []*UnitCounter{
			{Counters: CounterFilters{&CounterFilter{Value: 1}}}}},
		ActionTriggers: ActionTriggers{
			&ActionTrigger{
				Balance:        &BalanceFilter{Type: utils.StringPointer(utils.MetaMonetary)},
				ThresholdValue: 2, ActionsID: "TEST_ACTIONS", Executed: true},
			&ActionTrigger{
				Balance:        &BalanceFilter{Type: utils.StringPointer(utils.MetaMonetary)},
				ThresholdValue: 2, ActionsID: "TEST_ACTIONS", Executed: true}},
	}
	a := &Action{
		Balance: &BalanceFilter{Type: utils.StringPointer(utils.MetaVoice),
			Value: &utils.ValueFormula{Static: 5}, Weight: utils.Float64Pointer(20),
			DestinationIDs: utils.StringMapPointer(utils.NewStringMap("NAT"))}}
	topupResetAction(ub, a, nil, nil, nil, SharedActionsData{}, ActionConnCfg{})
	if ub.AllowNegative ||
		ub.BalanceMap[utils.MetaVoice].GetTotalValue() != 5 ||
		ub.BalanceMap[utils.MetaMonetary].GetTotalValue() != 100 ||
		len(ub.UnitCounters) != 0 ||
		len(ub.BalanceMap[utils.MetaVoice]) != 2 ||
		ub.ActionTriggers[0].Executed != true ||
		ub.ActionTriggers[1].Executed != true {
		t.Errorf("Topup reset minutes action failed: %+v",
			ub.BalanceMap[utils.MetaVoice][0])
	}
}

func TestActionTopupCredit(t *testing.T) {
	ub := &Account{
		ID: "TEST_UB",
		BalanceMap: map[string]Balances{
			utils.MetaMonetary: {&Balance{Value: 100}},
			utils.MetaVoice: {
				&Balance{Value: 10, Weight: 20,
					DestinationIDs: utils.NewStringMap("NAT")},
				&Balance{Weight: 10,
					DestinationIDs: utils.NewStringMap("RET")}}},
		UnitCounters: UnitCounters{
			utils.MetaMonetary: []*UnitCounter{
				{Counters: CounterFilters{&CounterFilter{Value: 1}}}}},
		ActionTriggers: ActionTriggers{
			&ActionTrigger{Balance: &BalanceFilter{Type: utils.StringPointer(utils.MetaMonetary)},
				ThresholdValue: 2, ActionsID: "TEST_ACTIONS", Executed: true},
			&ActionTrigger{Balance: &BalanceFilter{Type: utils.StringPointer(utils.MetaMonetary)},
				ThresholdValue: 2, ActionsID: "TEST_ACTIONS", Executed: true}},
	}
	a := &Action{Balance: &BalanceFilter{
		Type:  utils.StringPointer(utils.MetaMonetary),
		Value: &utils.ValueFormula{Static: 10}}}
	topupAction(ub, a, nil, nil, nil, SharedActionsData{}, ActionConnCfg{})
	if ub.AllowNegative ||
		ub.BalanceMap[utils.MetaMonetary].GetTotalValue() != 110 ||
		len(ub.UnitCounters) != 0 ||
		len(ub.BalanceMap[utils.MetaVoice]) != 2 ||
		ub.ActionTriggers[0].Executed != true ||
		ub.ActionTriggers[1].Executed != true {
		t.Error("Topup action failed!",
			ub.BalanceMap[utils.MetaMonetary].GetTotalValue())
	}
}

func TestActionTopupMinutes(t *testing.T) {
	ub := &Account{
		ID: "TEST_UB",
		BalanceMap: map[string]Balances{
			utils.MetaMonetary: {
				&Balance{Value: 100}},
			utils.MetaVoice: {&Balance{Value: 10, Weight: 20,
				DestinationIDs: utils.NewStringMap("NAT")},
				&Balance{Weight: 10, DestinationIDs: utils.NewStringMap("RET")}}},
		UnitCounters: UnitCounters{
			utils.MetaMonetary: []*UnitCounter{
				{Counters: CounterFilters{&CounterFilter{Value: 1}}}}},
		ActionTriggers: ActionTriggers{
			&ActionTrigger{Balance: &BalanceFilter{Type: utils.StringPointer(utils.MetaMonetary)},
				ThresholdValue: 2, ActionsID: "TEST_ACTIONS", Executed: true},
			&ActionTrigger{Balance: &BalanceFilter{Type: utils.StringPointer(utils.MetaMonetary)},
				ThresholdValue: 2, ActionsID: "TEST_ACTIONS", Executed: true}},
	}
	a := &Action{Balance: &BalanceFilter{Type: utils.StringPointer(utils.MetaVoice),
		Value: &utils.ValueFormula{Static: 5}, Weight: utils.Float64Pointer(20),
		DestinationIDs: utils.StringMapPointer(utils.NewStringMap("NAT"))}}
	topupAction(ub, a, nil, nil, nil, SharedActionsData{}, ActionConnCfg{})
	if ub.AllowNegative ||
		ub.BalanceMap[utils.MetaVoice].GetTotalValue() != 15 ||
		ub.BalanceMap[utils.MetaMonetary].GetTotalValue() != 100 ||
		len(ub.UnitCounters) != 0 ||
		len(ub.BalanceMap[utils.MetaVoice]) != 2 ||
		ub.ActionTriggers[0].Executed != true || ub.ActionTriggers[1].Executed != true {
		t.Error("Topup minutes action failed!", ub.BalanceMap[utils.MetaVoice])
	}
}

func TestActionDebitCredit(t *testing.T) {
	ub := &Account{
		ID: "TEST_UB",
		BalanceMap: map[string]Balances{
			utils.MetaMonetary: {&Balance{Value: 100}},
			utils.MetaVoice: {
				&Balance{Value: 10, Weight: 20, DestinationIDs: utils.NewStringMap("NAT")},
				&Balance{Weight: 10, DestinationIDs: utils.NewStringMap("RET")}}},
		UnitCounters: UnitCounters{
			utils.MetaMonetary: []*UnitCounter{
				{Counters: CounterFilters{&CounterFilter{Value: 1}}}}},
		ActionTriggers: ActionTriggers{
			&ActionTrigger{Balance: &BalanceFilter{Type: utils.StringPointer(utils.MetaMonetary)},
				ThresholdValue: 2, ActionsID: "TEST_ACTIONS", Executed: true},
			&ActionTrigger{Balance: &BalanceFilter{Type: utils.StringPointer(utils.MetaMonetary)},
				ThresholdValue: 2, ActionsID: "TEST_ACTIONS", Executed: true}},
	}
	a := &Action{Balance: &BalanceFilter{
		Type:  utils.StringPointer(utils.MetaMonetary),
		Value: &utils.ValueFormula{Static: 10}}}
	debitAction(ub, a, nil, nil, nil, SharedActionsData{}, ActionConnCfg{})
	if ub.AllowNegative ||
		ub.BalanceMap[utils.MetaMonetary].GetTotalValue() != 90 ||
		len(ub.UnitCounters) != 0 ||
		len(ub.BalanceMap[utils.MetaVoice]) != 2 ||
		ub.ActionTriggers[0].Executed != true ||
		ub.ActionTriggers[1].Executed != true {
		t.Error("Debit action failed!", utils.ToIJSON(ub))
	}
}

func TestActionDebitMinutes(t *testing.T) {
	ub := &Account{
		ID: "TEST_UB",
		BalanceMap: map[string]Balances{
			utils.MetaMonetary: {&Balance{Value: 100}},
			utils.MetaVoice: {
				&Balance{Value: 10, Weight: 20,
					DestinationIDs: utils.NewStringMap("NAT")},
				&Balance{Weight: 10, DestinationIDs: utils.NewStringMap("RET")}}},
		UnitCounters: UnitCounters{
			utils.MetaMonetary: []*UnitCounter{{Counters: CounterFilters{&CounterFilter{Value: 1}}}}},
		ActionTriggers: ActionTriggers{
			&ActionTrigger{Balance: &BalanceFilter{Type: utils.StringPointer(utils.MetaMonetary)},
				ThresholdValue: 2, ActionsID: "TEST_ACTIONS", Executed: true},
			&ActionTrigger{Balance: &BalanceFilter{Type: utils.StringPointer(utils.MetaMonetary)},
				ThresholdValue: 2, ActionsID: "TEST_ACTIONS", Executed: true}},
	}
	a := &Action{Balance: &BalanceFilter{
		Type:           utils.StringPointer(utils.MetaVoice),
		Value:          &utils.ValueFormula{Static: 5},
		Weight:         utils.Float64Pointer(20),
		DestinationIDs: utils.StringMapPointer(utils.NewStringMap("NAT"))}}
	debitAction(ub, a, nil, nil, nil, SharedActionsData{}, ActionConnCfg{})
	if ub.AllowNegative ||
		ub.BalanceMap[utils.MetaVoice][0].GetValue() != 5 ||
		ub.BalanceMap[utils.MetaMonetary].GetTotalValue() != 100 ||
		len(ub.UnitCounters) != 0 ||
		len(ub.BalanceMap[utils.MetaVoice]) != 2 ||
		ub.ActionTriggers[0].Executed != true || ub.ActionTriggers[1].Executed != true {
		t.Error("Debit minutes action failed!", ub.BalanceMap[utils.MetaVoice][0])
	}
}

func TestActionResetAllCounters(t *testing.T) {
	ub := &Account{
		ID:            "TEST_UB",
		AllowNegative: true,
		BalanceMap: map[string]Balances{
			utils.MetaMonetary: {&Balance{Value: 100}},
			utils.MetaVoice: {
				&Balance{Value: 10, Weight: 20,
					DestinationIDs: utils.NewStringMap("NAT")},
				&Balance{Weight: 10, DestinationIDs: utils.NewStringMap("RET")}}},
		ActionTriggers: ActionTriggers{
			&ActionTrigger{ThresholdType: utils.TriggerMaxEventCounter, ThresholdValue: 2,
				Balance: &BalanceFilter{Type: utils.StringPointer(utils.MetaMonetary),
					DestinationIDs: utils.StringMapPointer(utils.NewStringMap("NAT")),
					Weight:         utils.Float64Pointer(20)},
				ActionsID: "TEST_ACTIONS", Executed: true}},
	}
	ub.InitCounters()
	resetCountersAction(ub, nil, nil, nil, nil, SharedActionsData{}, ActionConnCfg{})
	if !ub.AllowNegative ||
		ub.BalanceMap[utils.MetaMonetary].GetTotalValue() != 100 ||
		len(ub.UnitCounters) != 1 ||
		len(ub.UnitCounters[utils.MetaMonetary][0].Counters) != 1 ||
		len(ub.BalanceMap[utils.MetaMonetary]) != 1 ||
		ub.ActionTriggers[0].Executed != true {
		t.Errorf("Reset counters action failed: %+v %+v %+v", ub.UnitCounters,
			ub.UnitCounters[utils.MetaMonetary][0], ub.UnitCounters[utils.MetaMonetary][0].Counters[0])
	}
	if len(ub.UnitCounters) < 1 {
		t.FailNow()
	}
	c := ub.UnitCounters[utils.MetaMonetary][0].Counters[0]
	if c.Filter.GetWeight() != 20 || c.Value != 0 ||
		c.Filter.GetDestinationIDs()["NAT"] == false {
		t.Errorf("Balance cloned incorrectly: %+v", c)
	}
}

func TestActionResetCounterOnlyDefault(t *testing.T) {
	ub := &Account{
		ID:            "TEST_UB",
		AllowNegative: true,
		BalanceMap: map[string]Balances{
			utils.MetaMonetary: {&Balance{Value: 100}},
			utils.MetaVoice: {&Balance{Value: 10, Weight: 20,
				DestinationIDs: utils.NewStringMap("NAT")}, &Balance{Weight: 10,
				DestinationIDs: utils.NewStringMap("RET")}}},
		ActionTriggers: ActionTriggers{
			&ActionTrigger{Balance: &BalanceFilter{Type: utils.StringPointer(utils.MetaMonetary)},
				ThresholdType: utils.TriggerMaxEventCounter, ThresholdValue: 2,
				ActionsID: "TEST_ACTIONS", Executed: true}},
	}
	a := &Action{Balance: &BalanceFilter{Type: utils.StringPointer(utils.MetaMonetary)}}
	ub.InitCounters()
	resetCountersAction(ub, a, nil, nil, nil, SharedActionsData{}, ActionConnCfg{})
	if !ub.AllowNegative ||
		ub.BalanceMap[utils.MetaMonetary].GetTotalValue() != 100 ||
		len(ub.UnitCounters) != 1 ||
		len(ub.UnitCounters[utils.MetaMonetary][0].Counters) != 1 ||
		len(ub.BalanceMap[utils.MetaVoice]) != 2 ||
		ub.ActionTriggers[0].Executed != true {
		for _, b := range ub.UnitCounters[utils.MetaMonetary][0].Counters {
			t.Logf("B: %+v", b)
		}
		t.Errorf("Reset counters action failed: %+v", ub.UnitCounters)
	}
	if len(ub.UnitCounters) < 1 || len(ub.UnitCounters[utils.MetaMonetary][0].Counters) < 1 {
		t.FailNow()
	}
	c := ub.UnitCounters[utils.MetaMonetary][0].Counters[0]
	if c.Filter.GetWeight() != 0 || c.Value != 0 || len(c.Filter.GetDestinationIDs()) != 0 {
		t.Errorf("Balance cloned incorrectly: %+v!", c)
	}
}

func TestActionResetCounterCredit(t *testing.T) {
	ub := &Account{
		ID:            "TEST_UB",
		AllowNegative: true,
		BalanceMap: map[string]Balances{
			utils.MetaMonetary: {&Balance{Value: 100}},
			utils.MetaVoice: {&Balance{Value: 10, Weight: 20, DestinationIDs: utils.NewStringMap("NAT")},
				&Balance{Weight: 10, DestinationIDs: utils.NewStringMap("RET")}}},
		UnitCounters: UnitCounters{
			utils.MetaMonetary: []*UnitCounter{
				{Counters: CounterFilters{
					&CounterFilter{Value: 1, Filter: new(BalanceFilter)}}}},
			utils.MetaSMS: []*UnitCounter{
				{Counters: CounterFilters{
					&CounterFilter{Value: 1, Filter: new(BalanceFilter)}}}}},
		ActionTriggers: ActionTriggers{
			&ActionTrigger{Balance: &BalanceFilter{Type: utils.StringPointer(utils.MetaMonetary)},
				ThresholdValue: 2, ActionsID: "TEST_ACTIONS", Executed: true}},
	}
	a := &Action{Balance: &BalanceFilter{Type: utils.StringPointer(utils.MetaMonetary)}}
	resetCountersAction(ub, a, nil, nil, nil, SharedActionsData{}, ActionConnCfg{})
	if !ub.AllowNegative ||
		ub.BalanceMap[utils.MetaMonetary].GetTotalValue() != 100 ||
		len(ub.UnitCounters) != 2 ||
		len(ub.BalanceMap[utils.MetaVoice]) != 2 ||
		ub.ActionTriggers[0].Executed != true {
		t.Error("Reset counters action failed!", ub.UnitCounters)
	}
}

func TestActionMakeNegative(t *testing.T) {
	a := &Action{Balance: &BalanceFilter{Value: &utils.ValueFormula{Static: 10}}}
	genericMakeNegative(a)
	if a.Balance.GetValue() > 0 {
		t.Error("Failed to make negative: ", a)
	}
	genericMakeNegative(a)
	if a.Balance.GetValue() > 0 {
		t.Error("Failed to preserve negative: ", a)
	}
}

func TestActionRemove(t *testing.T) {
	if _, err := dm.GetAccount("cgrates.org:remo"); err != nil {
		t.Errorf("account to be removed not found: %v", err)
	}
	a := &Action{
		ActionType: utils.MetaRemoveAccount,
	}

	at := &ActionTiming{
		accountIDs: utils.StringMap{"cgrates.org:remo": true},
		actions:    Actions{a},
	}
	at.Execute(nil, "")
	afterUb, err := dm.GetAccount("cgrates.org:remo")
	if err == nil || afterUb != nil {
		t.Error("error removing account: ", err, afterUb)
	}
}

func TestActionTopup(t *testing.T) {
	initialUb, _ := dm.GetAccount("vdf:minu")
	a := &Action{
		ActionType: utils.MetaTopUp,
		Balance: &BalanceFilter{Type: utils.StringPointer(utils.MetaMonetary), Value: &utils.ValueFormula{Static: 25},
			DestinationIDs: utils.StringMapPointer(utils.NewStringMap("RET")),
			Weight:         utils.Float64Pointer(20)},
	}

	at := &ActionTiming{
		accountIDs: utils.StringMap{"vdf:minu": true},
		actions:    Actions{a},
	}

	at.Execute(nil, "")
	afterUb, _ := dm.GetAccount("vdf:minu")
	initialValue := initialUb.BalanceMap[utils.MetaMonetary].GetTotalValue()
	afterValue := afterUb.BalanceMap[utils.MetaMonetary].GetTotalValue()
	if afterValue != initialValue+25 {
		t.Error("Bad topup before and after: ", initialValue, afterValue)
	}
}

func TestActionTopupLoaded(t *testing.T) {
	initialUb, _ := dm.GetAccount("vdf:minitsboy")
	a := &Action{
		ActionType: utils.MetaTopUp,
		Balance: &BalanceFilter{Type: utils.StringPointer(utils.MetaMonetary),
			Value:          &utils.ValueFormula{Static: 25},
			DestinationIDs: utils.StringMapPointer(utils.NewStringMap("RET")),
			Weight:         utils.Float64Pointer(20)},
	}

	at := &ActionTiming{
		accountIDs: utils.StringMap{"vdf:minitsboy": true},
		actions:    Actions{a},
	}

	at.Execute(nil, "")
	afterUb, _ := dm.GetAccount("vdf:minitsboy")
	initialValue := initialUb.BalanceMap[utils.MetaMonetary].GetTotalValue()
	afterValue := afterUb.BalanceMap[utils.MetaMonetary].GetTotalValue()
	if afterValue != initialValue+25 {
		t.Logf("Initial: %+v", initialUb)
		t.Logf("After: %+v", afterUb)
		t.Error("Bad topup before and after: ", initialValue, afterValue)
	}
}

func TestActionTransactionFuncType(t *testing.T) {
	err := dm.SetAccount(&Account{
		ID: "cgrates.org:trans",
		BalanceMap: map[string]Balances{
			utils.MetaMonetary: {&Balance{
				Value: 10,
			}},
		},
	})
	if err != nil {
		t.Error("Error setting account: ", err)
	}
	at := &ActionTiming{
		accountIDs: utils.StringMap{"cgrates.org:trans": true},
		Timing:     &RateInterval{},
		actions: []*Action{
			{
				ActionType: utils.MetaTopUp,
				Balance: &BalanceFilter{Value: &utils.ValueFormula{Static: 1.1},
					Type: utils.StringPointer(utils.MetaMonetary)},
			},
			{
				ActionType: "VALID_FUNCTION_TYPE",
				Balance: &BalanceFilter{Value: &utils.ValueFormula{Static: 1.1},
					Type: utils.StringPointer("test")},
			},
		},
	}
	at.Execute(nil, "")
	acc, err := dm.GetAccount("cgrates.org:trans")
	if err != nil || acc == nil {
		t.Error("Error getting account: ", acc, err)
	}
	if acc.BalanceMap[utils.MetaMonetary][0].Value != 10 {
		t.Errorf("Transaction didn't work: %v", acc.BalanceMap[utils.MetaMonetary][0].Value)
	}
}

func TestActionTransactionBalanceType(t *testing.T) {
	err := dm.SetAccount(&Account{
		ID: "cgrates.org:trans",
		BalanceMap: map[string]Balances{
			utils.MetaMonetary: {&Balance{
				Value: 10,
			}},
		},
	})
	if err != nil {
		t.Error("Error setting account: ", err)
	}
	at := &ActionTiming{
		accountIDs: utils.StringMap{"cgrates.org:trans": true},
		Timing:     &RateInterval{},
		actions: []*Action{
			{
				ActionType: utils.MetaTopUp,
				Balance: &BalanceFilter{Value: &utils.ValueFormula{Static: 1.1},
					Type: utils.StringPointer(utils.MetaMonetary)},
			},
			{
				ActionType: utils.MetaTopUp,
				Balance:    &BalanceFilter{Type: utils.StringPointer("test")},
			},
		},
	}
	err = at.Execute(nil, "")
	if err != nil {
		t.Error(err)
	}
	acc, err := dm.GetAccount("cgrates.org:trans")
	if err != nil || acc == nil {
		t.Error("Error getting account: ", acc, err)
	}
	if acc.BalanceMap[utils.MetaMonetary][0].Value != 11.1 {
		t.Errorf("Transaction didn't work: %v", acc.BalanceMap[utils.MetaMonetary][0].Value)
	}
}

func TestActionTransactionBalanceNotType(t *testing.T) {
	err := dm.SetAccount(&Account{
		ID: "cgrates.org:trans",
		BalanceMap: map[string]Balances{
			utils.MetaMonetary: {&Balance{
				Value: 10,
			}},
		},
	})
	if err != nil {
		t.Error("Error setting account: ", err)
	}
	at := &ActionTiming{
		accountIDs: utils.StringMap{"cgrates.org:trans": true},
		Timing:     &RateInterval{},
		actions: []*Action{
			{
				ActionType: utils.MetaTopUp,
				Balance: &BalanceFilter{Value: &utils.ValueFormula{Static: 1.1},
					Type: utils.StringPointer(utils.MetaVoice)},
			},
			{
				ActionType: utils.MetaTopUp,
				Balance:    &BalanceFilter{Type: utils.StringPointer("test")},
			},
		},
	}
	err = at.Execute(nil, "")
	if err != nil {
		t.Error(err)
	}
	acc, err := dm.GetAccount("cgrates.org:trans")
	if err != nil || acc == nil {
		t.Error("Error getting account: ", acc, err)
	}
	if acc.BalanceMap[utils.MetaMonetary][0].Value != 10.0 {
		t.Errorf("Transaction didn't work: %v", acc.BalanceMap[utils.MetaMonetary][0].Value)
	}
}

func TestActionWithExpireWithoutExpire(t *testing.T) {
	err := dm.SetAccount(&Account{
		ID: "cgrates.org:exp",
		BalanceMap: map[string]Balances{
			utils.MetaMonetary: {&Balance{
				Value: 10,
			}},
		},
	})
	if err != nil {
		t.Error("Error setting account: ", err)
	}
	at := &ActionTiming{
		accountIDs: utils.StringMap{"cgrates.org:exp": true},
		Timing:     &RateInterval{},
		actions: []*Action{
			{
				ActionType: utils.MetaTopUp,
				Balance: &BalanceFilter{
					Type:  utils.StringPointer(utils.MetaVoice),
					Value: &utils.ValueFormula{Static: 15},
				},
			},
			{
				ActionType: utils.MetaTopUp,
				Balance: &BalanceFilter{
					Type:           utils.StringPointer(utils.MetaVoice),
					Value:          &utils.ValueFormula{Static: 30},
					ExpirationDate: utils.TimePointer(time.Date(2025, time.November, 11, 22, 39, 0, 0, time.UTC)),
				},
			},
		},
	}
	err = at.Execute(nil, "")
	if err != nil {
		t.Error(err)
	}
	acc, err := dm.GetAccount("cgrates.org:exp")
	if err != nil || acc == nil {
		t.Errorf("Error getting account: %+v: %v", acc, err)
	}
	if len(acc.BalanceMap) != 2 ||
		len(acc.BalanceMap[utils.MetaVoice]) != 2 {
		t.Errorf("Error debiting expir and unexpire: %+v", acc.BalanceMap[utils.MetaVoice][0])
	}
}

func TestActionRemoveBalance(t *testing.T) {
	err := dm.SetAccount(&Account{
		ID: "cgrates.org:rembal",
		BalanceMap: map[string]Balances{
			utils.MetaMonetary: {
				&Balance{
					Value: 10,
				},
				&Balance{
					Value:          10,
					DestinationIDs: utils.NewStringMap("NAT", "RET"),
					ExpirationDate: time.Date(2025, time.November, 11, 22, 39, 0, 0, time.UTC),
				},
				&Balance{
					Value:          10,
					DestinationIDs: utils.NewStringMap("NAT", "RET"),
				},
			},
		},
	})
	if err != nil {
		t.Error("Error setting account: ", err)
	}
	at := &ActionTiming{
		accountIDs: utils.StringMap{"cgrates.org:rembal": true},
		Timing:     &RateInterval{},
		actions: []*Action{
			{
				ActionType: utils.MetaRemoveBalance,
				Balance: &BalanceFilter{
					Type:           utils.StringPointer(utils.MetaMonetary),
					DestinationIDs: utils.StringMapPointer(utils.NewStringMap("NAT", "RET")),
				},
			},
		},
	}
	err = at.Execute(nil, "")
	if err != nil {
		t.Error(err)
	}
	acc, err := dm.GetAccount("cgrates.org:rembal")
	if err != nil || acc == nil {
		t.Errorf("Error getting account: %+v: %v", acc, err)
	}
	if len(acc.BalanceMap) != 1 ||
		len(acc.BalanceMap[utils.MetaMonetary]) != 1 {
		t.Errorf("Error removing balance: %+v", acc.BalanceMap[utils.MetaMonetary])
	}
}

func TestActionRemoveExpiredBalance(t *testing.T) {
	err := dm.SetAccount(&Account{
		ID: "cgrates.org:rembal2",
		BalanceMap: map[string]Balances{
			utils.MetaMonetary: {
				&Balance{
					Value: 10,
				},
				&Balance{
					Value:          10,
					DestinationIDs: utils.NewStringMap("NAT", "RET"),
					ExpirationDate: time.Date(2025, time.November, 11, 22, 39, 0, 0, time.UTC),
				},
				&Balance{
					Value:          10,
					DestinationIDs: utils.NewStringMap("NAT", "RET"),
					ExpirationDate: time.Date(2010, time.November, 11, 22, 39, 0, 0, time.UTC),
				},
				&Balance{
					Value:          10,
					DestinationIDs: utils.NewStringMap("NAT", "RET"),
					ExpirationDate: time.Date(2012, time.November, 11, 22, 39, 0, 0, time.UTC),
				},
			},
		},
		Disabled: true,
	})
	if err != nil {
		t.Error("Error setting account: ", err)
	}
	at := &ActionTiming{
		accountIDs: utils.StringMap{"cgrates.org:rembal2": true},
		Timing:     &RateInterval{},
		actions: []*Action{
			{
				ActionType: utils.MetaRemoveExpired,
				Balance: &BalanceFilter{
					Type: utils.StringPointer(utils.MetaMonetary),
				},
			},
		},
	}
	err = at.Execute(nil, "")
	if err != nil {
		t.Error(err)
	}
	acc, err := dm.GetAccount("cgrates.org:rembal2")
	if err != nil || acc == nil {
		t.Errorf("Error getting account: %+v: %v", acc, err)
	}
	if len(acc.BalanceMap) != 1 ||
		len(acc.BalanceMap[utils.MetaMonetary]) != 2 {
		t.Errorf("Error removing balance: %+v", utils.ToJSON(acc.BalanceMap[utils.MetaMonetary]))
	}
}

func TestActionTransferMonetaryDefault(t *testing.T) {
	err := dm.SetAccount(
		&Account{
			ID: "cgrates.org:trans",
			BalanceMap: map[string]Balances{
				utils.MetaMonetary: {
					&Balance{
						Uuid:  utils.GenUUID(),
						ID:    utils.MetaDefault,
						Value: 10,
					},
					&Balance{
						Uuid:  utils.GenUUID(),
						Value: 3,
					},
					&Balance{
						Uuid:  utils.GenUUID(),
						Value: 6,
					},
					&Balance{
						Uuid:  utils.GenUUID(),
						Value: -2,
					},
				},
			},
		})
	if err != nil {
		t.Errorf("error setting account: %v", err)
	}

	a := &Action{
		ActionType: utils.MetaTransferMonetaryDefault,
	}

	at := &ActionTiming{
		accountIDs: utils.StringMap{"cgrates.org:trans": true},
		actions:    Actions{a},
	}
	at.Execute(nil, "")

	afterUb, err := dm.GetAccount("cgrates.org:trans")
	if err != nil {
		t.Error("account not found: ", err, afterUb)
	}
	if afterUb.BalanceMap[utils.MetaMonetary].GetTotalValue() != 17 ||
		afterUb.BalanceMap[utils.MetaMonetary][0].Value != 19 ||
		afterUb.BalanceMap[utils.MetaMonetary][1].Value != 0 ||
		afterUb.BalanceMap[utils.MetaMonetary][2].Value != 0 ||
		afterUb.BalanceMap[utils.MetaMonetary][3].Value != -2 {
		for _, b := range afterUb.BalanceMap[utils.MetaMonetary] {
			t.Logf("B: %+v", b)
		}
		t.Error("transfer balance value: ", afterUb.BalanceMap[utils.MetaMonetary].GetTotalValue())
	}
}

func TestActionTransferMonetaryDefaultFilter(t *testing.T) {
	err := dm.SetAccount(
		&Account{
			ID: "cgrates.org:trans",
			BalanceMap: map[string]Balances{
				utils.MetaMonetary: {
					&Balance{
						Uuid:   utils.GenUUID(),
						ID:     utils.MetaDefault,
						Value:  10,
						Weight: 20,
					},
					&Balance{
						Uuid:   utils.GenUUID(),
						Value:  3,
						Weight: 20,
					},
					&Balance{
						Uuid:   utils.GenUUID(),
						Value:  1,
						Weight: 10,
					},
					&Balance{
						Uuid:   utils.GenUUID(),
						Value:  6,
						Weight: 20,
					},
				},
			},
		})
	if err != nil {
		t.Errorf("error setting account: %v", err)
	}

	a := &Action{
		ActionType: utils.MetaTransferMonetaryDefault,
		Balance:    &BalanceFilter{Weight: utils.Float64Pointer(20)},
	}

	at := &ActionTiming{
		accountIDs: utils.StringMap{"cgrates.org:trans": true},
		actions:    Actions{a},
	}
	at.Execute(nil, "")

	afterUb, err := dm.GetAccount("cgrates.org:trans")
	if err != nil {
		t.Error("account not found: ", err, afterUb)
	}
	if afterUb.BalanceMap[utils.MetaMonetary].GetTotalValue() != 20 ||
		afterUb.BalanceMap[utils.MetaMonetary][0].Value != 19 ||
		afterUb.BalanceMap[utils.MetaMonetary][1].Value != 0 ||
		afterUb.BalanceMap[utils.MetaMonetary][2].Value != 1 ||
		afterUb.BalanceMap[utils.MetaMonetary][3].Value != 0 {
		for _, b := range afterUb.BalanceMap[utils.MetaMonetary] {
			t.Logf("B: %+v", b)
		}
		t.Error("transfer balance value: ", afterUb.BalanceMap[utils.MetaMonetary].GetTotalValue())
	}
}

func TestActionConditionalTopup(t *testing.T) {
	err := dm.SetAccount(
		&Account{
			ID: "cgrates.org:cond",
			BalanceMap: map[string]Balances{
				utils.MetaMonetary: {
					&Balance{
						Uuid:   utils.GenUUID(),
						ID:     utils.MetaDefault,
						Value:  10,
						Weight: 20,
					},
					&Balance{
						Uuid:   utils.GenUUID(),
						Value:  3,
						Weight: 20,
					},
					&Balance{
						Uuid:   utils.GenUUID(),
						Value:  1,
						Weight: 10,
					},
					&Balance{
						Uuid:   utils.GenUUID(),
						Value:  6,
						Weight: 20,
					},
				},
			},
		})
	if err != nil {
		t.Errorf("error setting account: %v", err)
	}

	a := &Action{
		ActionType: utils.MetaTopUp,
		Filters:    []string{`*lt:~*req.BalanceMap.*monetary.GetTotalValue:30`},
		Balance: &BalanceFilter{
			Type:   utils.StringPointer(utils.MetaMonetary),
			Value:  &utils.ValueFormula{Static: 11},
			Weight: utils.Float64Pointer(30),
		},
	}

	at := &ActionTiming{
		accountIDs: utils.StringMap{"cgrates.org:cond": true},
		actions:    Actions{a},
	}
	if err = at.Execute(NewFilterS(config.CgrConfig(), nil, nil), ""); err != nil {
		t.Fatal(err)
	}

	afterUb, err := dm.GetAccount("cgrates.org:cond")
	if err != nil {
		t.Error("account not found: ", err, afterUb)
	}
	if len(afterUb.BalanceMap[utils.MetaMonetary]) != 5 ||
		afterUb.BalanceMap[utils.MetaMonetary].GetTotalValue() != 31 ||
		afterUb.BalanceMap[utils.MetaMonetary][4].Value != 11 {
		for _, b := range afterUb.BalanceMap[utils.MetaMonetary] {
			t.Logf("B: %+v", b)
		}
		t.Error("transfer balance value: ", afterUb.BalanceMap[utils.MetaMonetary].GetTotalValue())
	}
}

func TestActionConditionalTopupNoMatch(t *testing.T) {
	err := dm.SetAccount(
		&Account{
			ID: "cgrates.org:cond",
			BalanceMap: map[string]Balances{
				utils.MetaMonetary: {
					&Balance{
						Uuid:   utils.GenUUID(),
						ID:     utils.MetaDefault,
						Value:  10,
						Weight: 20,
					},
					&Balance{
						Uuid:   utils.GenUUID(),
						Value:  3,
						Weight: 20,
					},
					&Balance{
						Uuid:   utils.GenUUID(),
						Value:  1,
						Weight: 10,
					},
					&Balance{
						Uuid:   utils.GenUUID(),
						Value:  6,
						Weight: 20,
					},
				},
			},
		})
	if err != nil {
		t.Errorf("error setting account: %v", err)
	}

	a := &Action{
		ActionType: utils.MetaTopUp,
		Filters:    []string{`*lt:~*req.BalanceMap.*monetary.GetTotalValue:3`},
		Balance: &BalanceFilter{
			Type:   utils.StringPointer(utils.MetaMonetary),
			Value:  &utils.ValueFormula{Static: 11},
			Weight: utils.Float64Pointer(30),
		},
	}

	at := &ActionTiming{
		accountIDs: utils.StringMap{"cgrates.org:cond": true},
		actions:    Actions{a},
	}
	at.Execute(NewFilterS(config.CgrConfig(), nil, nil), "")

	afterUb, err := dm.GetAccount("cgrates.org:cond")
	if err != nil {
		t.Error("account not found: ", err, afterUb)
	}
	if len(afterUb.BalanceMap[utils.MetaMonetary]) != 4 ||
		afterUb.BalanceMap[utils.MetaMonetary].GetTotalValue() != 20 {
		for _, b := range afterUb.BalanceMap[utils.MetaMonetary] {
			t.Logf("B: %+v", b)
		}
		t.Error("transfer balance value: ", afterUb.BalanceMap[utils.MetaMonetary].GetTotalValue())
	}
}

func TestActionConditionalTopupExistingBalance(t *testing.T) {
	err := dm.SetAccount(
		&Account{
			ID: "cgrates.org:cond",
			BalanceMap: map[string]Balances{
				utils.MetaMonetary: {
					&Balance{
						Uuid:   utils.GenUUID(),
						Value:  1,
						Weight: 10,
					},
					&Balance{
						Uuid:   utils.GenUUID(),
						Value:  6,
						Weight: 20,
					},
				},
				utils.MetaVoice: {
					&Balance{
						Uuid:   utils.GenUUID(),
						Value:  10,
						Weight: 10,
					},
					&Balance{
						Uuid:   utils.GenUUID(),
						Value:  100,
						Weight: 20,
					},
				},
			},
		})
	if err != nil {
		t.Errorf("error setting account: %v", err)
	}

	a := &Action{
		ActionType: utils.MetaTopUp,
		Filters:    []string{`*gte:~*req.BalanceMap.*voice.GetTotalValue:100`},
		Balance: &BalanceFilter{
			Type:   utils.StringPointer(utils.MetaMonetary),
			Value:  &utils.ValueFormula{Static: 11},
			Weight: utils.Float64Pointer(10),
		},
	}

	at := &ActionTiming{
		accountIDs: utils.StringMap{"cgrates.org:cond": true},
		actions:    Actions{a},
	}
	at.Execute(NewFilterS(config.CgrConfig(), nil, nil), "")

	afterUb, err := dm.GetAccount("cgrates.org:cond")
	if err != nil {
		t.Error("account not found: ", err, afterUb)
	}
	if len(afterUb.BalanceMap[utils.MetaMonetary]) != 2 ||
		afterUb.BalanceMap[utils.MetaMonetary].GetTotalValue() != 18 {
		for _, b := range afterUb.BalanceMap[utils.MetaMonetary] {
			t.Logf("B: %+v", b)
		}
		t.Error("transfer balance value: ", afterUb.BalanceMap[utils.MetaMonetary].GetTotalValue())
	}
}

func TestActionConditionalDisabledIfNegative(t *testing.T) {
	err := dm.SetAccount(
		&Account{
			ID: "cgrates.org:af",
			BalanceMap: map[string]Balances{
				utils.MetaData: {
					&Balance{
						Uuid:          "fc927edb-1bd6-425e-a2a3-9fd8bafaa524",
						ID:            "for_v3hsillmilld500m_data_500_m",
						Value:         5.242,
						Weight:        10,
						RatingSubject: "for_v3hsillmilld500m_data_forfait",
						Categories: utils.StringMap{
							"data_france": true,
						},
					},
				},
				utils.MetaMonetary: {
					&Balance{
						Uuid:  "9fa1847a-f36a-41a7-8ec0-dfaab370141e",
						ID:    utils.MetaDefault,
						Value: -1.95001,
					},
				},
				utils.MetaSMS: {
					&Balance{
						Uuid:   "d348d15d-2988-4ee4-b847-6a552f94e2ec",
						ID:     "for_v3hsillmilld500m_mms_ill",
						Value:  20000,
						Weight: 10,
						DestinationIDs: utils.StringMap{
							"FRANCE_NATIONAL": true,
						},
						Categories: utils.StringMap{
							"mms_france":  true,
							"tmms_france": true,
							"vmms_france": true,
						},
					},
					&Balance{
						Uuid:   "f4643517-31f6-4199-980f-04cf535471ed",
						ID:     "for_v3hsillmilld500m_sms_ill",
						Value:  20000,
						Weight: 10,
						DestinationIDs: utils.StringMap{
							"FRANCE_NATIONAL": true,
						},
						Categories: utils.StringMap{
							"ms_france": true,
						},
					},
				},
				utils.MetaVoice: {
					&Balance{
						Uuid:   "079ab190-77f4-44f3-9c6f-3a0dd1a59dfd",
						ID:     "for_v3hsillmilld500m_voice_3_h",
						Value:  10800,
						Weight: 10,
						DestinationIDs: utils.StringMap{
							"FRANCE_NATIONAL": true,
						},
						Categories: utils.StringMap{
							"call_france": true,
						},
					},
				},
			},
		})
	if err != nil {
		t.Errorf("error setting account: %v", err)
	}

	a1 := &Action{
		ActionType: utils.MetaSetBalance,
		Filters:    []string{"*string:~*req.BalanceMap.*monetary[0].ID:*default", "*lt:~*req.BalanceMap.*monetary[0].Value:0"},
		Balance: &BalanceFilter{
			Type:     utils.StringPointer("*sms"),
			ID:       utils.StringPointer("for_v3hsillmilld500m_sms_ill"),
			Disabled: utils.BoolPointer(true),
		},
		Weight: 9,
	}
	a2 := &Action{
		ActionType: utils.MetaSetBalance,
		Filters:    []string{"*string:~*req.BalanceMap.*monetary[0].ID:*default", "*lt:~*req.BalanceMap.*monetary[0].Value:0"},
		Balance: &BalanceFilter{
			Type:           utils.StringPointer("*sms"),
			ID:             utils.StringPointer("for_v3hsillmilld500m_mms_ill"),
			DestinationIDs: utils.StringMapPointer(utils.NewStringMap("FRANCE_NATIONAL")),
			Weight:         utils.Float64Pointer(10),
			Disabled:       utils.BoolPointer(true),
		},
		Weight: 8,
	}
	a3 := &Action{
		ActionType: utils.MetaSetBalance,
		Filters:    []string{"*string:~*req.BalanceMap.*monetary[0].ID:*default", "*lt:~*req.BalanceMap.*monetary[0].Value:0"},
		Balance: &BalanceFilter{
			Type:           utils.StringPointer("*sms"),
			ID:             utils.StringPointer("for_v3hsillmilld500m_sms_ill"),
			DestinationIDs: utils.StringMapPointer(utils.NewStringMap("FRANCE_NATIONAL")),
			Weight:         utils.Float64Pointer(10),
			Disabled:       utils.BoolPointer(true),
		},
		Weight: 8,
	}
	a4 := &Action{
		ActionType: utils.MetaSetBalance,
		Filters:    []string{"*string:~*req.BalanceMap.*monetary[0].ID:*default", "*lt:~*req.BalanceMap.*monetary[0].Value:0"},
		Balance: &BalanceFilter{
			Type:          utils.StringPointer("*data"),
			Uuid:          utils.StringPointer("fc927edb-1bd6-425e-a2a3-9fd8bafaa524"),
			RatingSubject: utils.StringPointer("for_v3hsillmilld500m_data_forfait"),
			Weight:        utils.Float64Pointer(10),
			Disabled:      utils.BoolPointer(true),
		},
		Weight: 7,
	}
	a5 := &Action{
		ActionType: utils.MetaSetBalance,
		Filters:    []string{"*string:~*req.BalanceMap.*monetary[0].ID:*default", "*lt:~*req.BalanceMap.*monetary[0].Value:0"},
		Balance: &BalanceFilter{
			Type:           utils.StringPointer("*voice"),
			ID:             utils.StringPointer("for_v3hsillmilld500m_voice_3_h"),
			DestinationIDs: utils.StringMapPointer(utils.NewStringMap("FRANCE_NATIONAL")),
			Weight:         utils.Float64Pointer(10),
			Disabled:       utils.BoolPointer(true),
		},
		Weight: 6,
	}

	at := &ActionTiming{
		accountIDs: utils.StringMap{"cgrates.org:af": true},
		actions:    Actions{a1, a2, a3, a4, a5},
	}
	at.Execute(NewFilterS(config.CgrConfig(), nil, nil), "")

	afterUb, err := dm.GetAccount("cgrates.org:af")
	if err != nil {
		t.Error("account not found: ", err, afterUb)
	}

	for btype, chain := range afterUb.BalanceMap {
		if btype != utils.MetaMonetary {
			for _, b := range chain {
				if b.Disabled != true {
					t.Errorf("Failed to disabled balance (%s): %+v", btype, b)
				}
			}
		}
	}
}

func TestActionSetBalance(t *testing.T) {
	err := dm.SetAccount(
		&Account{
			ID: "cgrates.org:setb",
			BalanceMap: map[string]Balances{
				utils.MetaMonetary: {
					&Balance{
						ID:     "m1",
						Uuid:   utils.GenUUID(),
						Value:  1,
						Weight: 10,
					},
					&Balance{
						ID:     "m2",
						Uuid:   utils.GenUUID(),
						Value:  6,
						Weight: 20,
					},
				},
				utils.MetaVoice: {
					&Balance{
						ID:     "v1",
						Uuid:   utils.GenUUID(),
						Value:  10,
						Weight: 10,
					},
					&Balance{
						ID:     "v2",
						Uuid:   utils.GenUUID(),
						Value:  100,
						Weight: 20,
					},
				},
			},
		})
	if err != nil {
		t.Errorf("error setting account: %v", err)
	}

	a := &Action{
		ActionType: utils.MetaSetBalance,
		Balance: &BalanceFilter{
			ID:     utils.StringPointer("m2"),
			Type:   utils.StringPointer(utils.MetaMonetary),
			Value:  &utils.ValueFormula{Static: 11},
			Weight: utils.Float64Pointer(10),
		},
	}

	at := &ActionTiming{
		accountIDs: utils.StringMap{"cgrates.org:setb": true},
		actions:    Actions{a},
	}
	at.Execute(nil, "")

	afterUb, err := dm.GetAccount("cgrates.org:setb")
	if err != nil {
		t.Error("account not found: ", err, afterUb)
	}
	if len(afterUb.BalanceMap[utils.MetaMonetary]) != 2 ||
		afterUb.BalanceMap[utils.MetaMonetary][1].Value != 11 ||
		afterUb.BalanceMap[utils.MetaMonetary][1].Weight != 10 {
		for _, b := range afterUb.BalanceMap[utils.MetaMonetary] {
			t.Logf("B: %+v", b)
		}
		t.Errorf("Balance: %+v", afterUb.BalanceMap[utils.MetaMonetary][1])
	}
}

func TestActionCSVFilter(t *testing.T) {
	act, err := dm.GetActions("FILTER", false, utils.NonTransactional)
	if err != nil {
		t.Error("error getting actions: ", err)
	}
	if len(act) != 1 || !reflect.DeepEqual(act[0].Filters, []string{"*string:~*req.BalanceMap.*monetary[0].ID:*default", "*lt:~*req.BalanceMap.*monetary[0].Value:0"}) {
		t.Error("Error loading actions: ", act[0].Filters)
	}
}

func TestActionExpirationTime(t *testing.T) {
	a, err := dm.GetActions("EXP", false, utils.NonTransactional)
	if err != nil || a == nil {
		t.Error("Error getting actions: ", err)
	}

	at := &ActionTiming{
		accountIDs: utils.StringMap{"cgrates.org:expo": true},
		actions:    a,
	}
	for rep := 0; rep < 5; rep++ {
		at.Execute(nil, "")
		afterUb, err := dm.GetAccount("cgrates.org:expo")
		if err != nil ||
			len(afterUb.BalanceMap[utils.MetaVoice]) != rep+1 {
			t.Error("error topuping expiration balance: ", utils.ToIJSON(afterUb))
		}
	}
}

func TestActionExpNoExp(t *testing.T) {
	exp, err := dm.GetActions("EXP", false, utils.NonTransactional)
	if err != nil || exp == nil {
		t.Error("Error getting actions: ", err)
	}
	noexp, err := dm.GetActions("NOEXP", false, utils.NonTransactional)
	if err != nil || noexp == nil {
		t.Error("Error getting actions: ", err)
	}
	exp = append(exp, noexp...)
	at := &ActionTiming{
		accountIDs: utils.StringMap{"cgrates.org:expnoexp": true},
		actions:    exp,
	}
	at.Execute(nil, "")
	afterUb, err := dm.GetAccount("cgrates.org:expnoexp")
	if err != nil ||
		len(afterUb.BalanceMap[utils.MetaVoice]) != 2 {
		t.Error("error topuping expiration balance: ", utils.ToIJSON(afterUb))
	}
}

func TestActionTopUpZeroNegative(t *testing.T) {
	account := &Account{
		ID: "cgrates.org:zeroNegative",
		BalanceMap: map[string]Balances{
			utils.MetaMonetary: {
				&Balance{
					ID:    "Bal1",
					Value: -10,
				},
				&Balance{
					ID:    "Bal2",
					Value: 5,
				},
			},
		},
	}
	err := dm.SetAccount(account)
	if err != nil {
		t.Error("Error setting account: ", err)
	}
	at := &ActionTiming{
		accountIDs: utils.StringMap{"cgrates.org:zeroNegative": true},
		Timing:     &RateInterval{},
		actions: []*Action{
			{
				Id:         "ZeroMonetary",
				ActionType: utils.TopUpZeroNegative,
				Balance: &BalanceFilter{
					Type: utils.StringPointer(utils.MetaMonetary),
				},
			},
		},
	}
	err = at.Execute(nil, "")
	if err != nil {
		t.Error(err)
	}
	acc, err := dm.GetAccount("cgrates.org:zeroNegative")
	if err != nil || acc == nil {
		t.Error("Error getting account: ", acc, err)
	}
	//Verify value for first balance(Bal1) should be 0 after execute action TopUpZeroNegative
	if acc.BalanceMap[utils.MetaMonetary][0].Value != 0 {
		t.Errorf("Expecting 0, received: %+v", acc.BalanceMap[utils.MetaMonetary][0].Value)
	}
	//Verify value for secound balance(Bal2) should be the same
	if acc.BalanceMap[utils.MetaMonetary][1].Value != 5 {
		t.Errorf("Expecting 5, received: %+v", acc.BalanceMap[utils.MetaMonetary][1].Value)
	}
}

func TestActionSetExpiry(t *testing.T) {
	timeNowPlus24h := time.Now().Add(24 * time.Hour)
	account := &Account{
		ID: "cgrates.org:zeroNegative",
		BalanceMap: map[string]Balances{
			utils.MetaMonetary: {
				&Balance{
					ID:    "Bal1",
					Value: -10,
				},
				&Balance{
					ID:    "Bal2",
					Value: 5,
				},
			},
		},
	}
	err := dm.SetAccount(account)
	if err != nil {
		t.Error("Error setting account: ", err)
	}
	at := &ActionTiming{
		accountIDs: utils.StringMap{"cgrates.org:zeroNegative": true},
		Timing:     &RateInterval{},
		actions: []*Action{
			{
				Id:         "SetExpiry",
				ActionType: utils.SetExpiry,
				Balance: &BalanceFilter{
					ID:             utils.StringPointer("Bal1"),
					Type:           utils.StringPointer(utils.MetaMonetary),
					ExpirationDate: utils.TimePointer(timeNowPlus24h),
				},
			},
		},
	}
	err = at.Execute(nil, "")
	if err != nil {
		t.Error(err)
	}
	acc, err := dm.GetAccount("cgrates.org:zeroNegative")
	if err != nil || acc == nil {
		t.Error("Error getting account: ", acc, err)
	}
	//Verify ExpirationDate for first balance(Bal1)
	if !acc.BalanceMap[utils.MetaMonetary][0].ExpirationDate.Equal(timeNowPlus24h) {
		t.Errorf("Expecting: %+v, received: %+v", timeNowPlus24h, acc.BalanceMap[utils.MetaMonetary][0].ExpirationDate)
	}
}

type TestRPCParameters struct {
	status string
}

type Attr struct {
	Name    string
	Surname string
	Age     float64
}

func (trpcp *TestRPCParameters) Hopa(ctx *context.Context, in *Attr, out *float64) error {
	trpcp.status = utils.OK
	return nil
}

func TestCgrRpcAction(t *testing.T) {
	trpcp := &TestRPCParameters{}
	utils.RegisterRpcParams("", trpcp)
	a := &Action{
		ExtraParameters: `{"Address": "*internal",
	"Transport": "*gob",
	"Method": "TestRPCParameters.Hopa",
	"Attempts":1,
	"Async" :false,
	"Params": {"Name":"n", "Surname":"s", "Age":10.2}}`,
	}
	if err := cgrRPCAction(nil, a, nil, nil, nil, SharedActionsData{}, ActionConnCfg{}); err != nil {
		t.Error("error executing cgr action: ", err)
	}
	if trpcp.status != utils.OK {
		t.Error("RPC not called!")
	}
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().ReplyTimeout = 1000 * time.Millisecond
	cfg.GeneralCfg().ConnectTimeout = 1000 * time.Millisecond
	config.SetCgrConfig(cfg)
	defer func() {
		config.SetCgrConfig(config.NewDefaultCGRConfig())
	}()
	a = &Action{
		ExtraParameters: `{"Address": "*json*localhost",
	"Transport": "*gob",
	"Method": "TestRPCParameters.Hopa",
	"Attempts":1,
	"Async" :false,
	"Params": {"Name":"n", "Surname":"s", "Age":10.2}}`,
	}
	if err := cgrRPCAction(nil, a, nil, nil, nil, SharedActionsData{}, ActionConnCfg{}); err == nil {
		t.Error("error executing cgr action: ", err)
	}
}

func TestValueFormulaDebit(t *testing.T) {
	if _, err := dm.GetAccount("cgrates.org:vf"); err != nil {
		t.Errorf("account to be removed not found: %v", err)
	}

	at := &ActionTiming{
		accountIDs: utils.StringMap{"cgrates.org:vf": true},
		ActionsID:  "VF",
	}
	at.Execute(nil, "")
	afterUb, err := dm.GetAccount("cgrates.org:vf")
	// not an exact value, depends of month
	v := afterUb.BalanceMap[utils.MetaMonetary].GetTotalValue()
	if err != nil || v > -0.30 || v < -0.36 {
		t.Error("error debiting account: ", err, utils.ToIJSON(afterUb), v)
	}
}

func TestClonedAction(t *testing.T) {

	var a *Action
	if clone := a.Clone(); clone != nil {
		t.Errorf("Expected nil but got %v", clone)
	}
	a = &Action{
		Id:         "test1",
		ActionType: utils.MetaTopUp,
		Balance: &BalanceFilter{
			ID:    utils.StringPointer(utils.MetaDefault),
			Value: &utils.ValueFormula{Static: 1},
			Type:  utils.StringPointer(utils.MetaMonetary),
		},
		Weight: float64(10),
	}
	if clone := a.Clone(); !reflect.DeepEqual(a, clone) {
		t.Error("error cloning action: ", utils.ToIJSON(clone))
	}
}

func TestClonedActions(t *testing.T) {
	actions := Actions{
		&Action{
			Id:         "RECUR_FOR_V3HSILLMILLD1G",
			ActionType: utils.MetaTopUp,
			Balance: &BalanceFilter{
				ID:    utils.StringPointer(utils.MetaDefault),
				Value: &utils.ValueFormula{Static: 1},
				Type:  utils.StringPointer(utils.MetaMonetary),
			},
			Weight: float64(30),
		},
		&Action{
			Id:         "RECUR_FOR_V3HSILLMILLD5G",
			ActionType: utils.MetaDebit,
			Balance: &BalanceFilter{
				ID:    utils.StringPointer(utils.MetaDefault),
				Value: &utils.ValueFormula{Static: 2},
				Type:  utils.StringPointer(utils.MetaMonetary),
			},
			Weight: float64(20),
		},
	}
	clone := actions.Clone()
	if !reflect.DeepEqual(actions, clone) {
		t.Errorf("Expecting %+v, received: %+v", utils.ToIJSON(actions), utils.ToIJSON(clone))
	}

}

func TestCacheGetClonedActions(t *testing.T) {
	actions := Actions{
		&Action{
			Id:         "RECUR_FOR_V3HSILLMILLD1G",
			ActionType: utils.MetaTopUp,
			Balance: &BalanceFilter{
				ID:    utils.StringPointer(utils.MetaDefault),
				Value: &utils.ValueFormula{Static: 1},
				Type:  utils.StringPointer(utils.MetaMonetary),
			},
			Weight: float64(30),
		},
		&Action{
			Id:         "REACT_FOR_V3HSILLMILL",
			ActionType: utils.MetaSetBalance,
			Balance: &BalanceFilter{
				ID:    utils.StringPointer("for_v3hsillmill_sms_ill"),
				Type:  utils.StringPointer(utils.MetaSMS),
				Value: &utils.ValueFormula{Static: 20000},
				DestinationIDs: &utils.StringMap{
					"FRANCE_NATIONAL":      true,
					"FRANCE_NATIONAL_FREE": false,
					"ZONE1":                false},
				Categories: &utils.StringMap{
					"sms_eurotarif": true,
					"sms_france":    true},
				Disabled: utils.BoolPointer(false),
				Blocker:  utils.BoolPointer(false),
			},
			Weight: float64(10),
		},
	}
	if err := Cache.Set(utils.CacheActions, "MYTEST", actions, nil, true, ""); err != nil {
		t.Errorf("Expecting: nil, received: %s", err)
	}
	clned, ok := Cache.Get(utils.CacheActions, "MYTEST")
	if !ok {
		t.Error(ltcache.ErrNotFound)
	}
	aCloned := clned.(Actions).Clone()
	if !reflect.DeepEqual(actions, aCloned) {
		t.Errorf("Expecting: %+v, received: %+v", actions[1].Balance, aCloned[1].Balance)
	}
}

// TestCdrLogAction
type RPCMock struct {
	args *ArgV1ProcessEvents
}

func (r *RPCMock) Call(ctx *context.Context, method string, args any, rply any) error {
	if method != utils.CDRsV1ProcessEvents {
		return rpcclient.ErrUnsupporteServiceMethod
	}
	if r.args != nil {
		return fmt.Errorf("There should be only one call to this function")
	}
	r.args = args.(*ArgV1ProcessEvents)
	rp := rply.(*string)
	*rp = utils.OK
	return nil
}

func TestCdrLogAction(t *testing.T) {
	mock := RPCMock{}

	dfltCfg := config.NewDefaultCGRConfig()
	if err != nil {
		t.Fatal(err)
	}
	dfltCfg.SchedulerCfg().CDRsConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCDRs)}
	config.SetCgrConfig(dfltCfg)

	internalChan := make(chan birpc.ClientConnector, 1)
	internalChan <- &mock

	NewConnManager(dfltCfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCDRs): internalChan,
	})

	extraData := map[string]any{
		"test": "val",
	}
	acc := &Account{
		ID: "cgrates.org:1001",
		BalanceMap: map[string]Balances{
			utils.MetaMonetary: {
				&Balance{Value: 20},
			},
		},
		UnitCounters: UnitCounters{
			utils.MetaMonetary: []*UnitCounter{
				{
					Counters: CounterFilters{
						&CounterFilter{Value: 1},
					},
				},
			},
		},
	}
	a := &Action{
		Id:              "CDRLog1",
		ActionType:      utils.CDRLog,
		ExtraParameters: "{\"BalanceID\":\"~*acnt.BalanceID\",\"ActionID\":\"~*act.ActionID\",\"BalanceValue\":\"~*acnt.BalanceValue\"}",
		Weight:          50,
	}
	acs := Actions{
		a,
		&Action{
			Id:         "CdrDebit",
			ActionType: "*debit",
			Balance: &BalanceFilter{
				ID:     utils.StringPointer(utils.MetaDefault),
				Value:  &utils.ValueFormula{Static: 9.95},
				Type:   utils.StringPointer(utils.MetaMonetary),
				Weight: utils.Float64Pointer(0),
			},
			Weight:       float64(90),
			balanceValue: 10,
		},
	}
	if err := cdrLogAction(acc, a, acs, nil, extraData, SharedActionsData{}, ActionConnCfg{}); err != nil {
		t.Fatal(err)
	}
	if mock.args == nil {
		t.Fatalf("Expected a call to %s", utils.CDRsV1ProcessEvent)
	}
	expCgrEv := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     mock.args.CGREvents[0].ID,
		Event: map[string]any{
			"Account":      "1001",
			"ActionID":     "CdrDebit",
			"AnswerTime":   mock.args.CGREvents[0].Event["AnswerTime"],
			"BalanceID":    "*default",
			"BalanceValue": "10",
			"CGRID":        mock.args.CGREvents[0].Event["CGRID"],
			"Category":     "",
			"Cost":         9.95,
			"CostSource":   "",
			"Destination":  "",
			"ExtraInfo":    "",
			"OrderID":      mock.args.CGREvents[0].Event["OrderID"],
			"OriginHost":   "127.0.0.1",
			"OriginID":     mock.args.CGREvents[0].Event["OriginID"],
			"Partial":      false,
			"PreRated":     true,
			"RequestType":  "*none",
			"RunID":        "*debit",
			"SetupTime":    mock.args.CGREvents[0].Event["SetupTime"],
			"Source":       utils.CDRLog,
			"Subject":      "1001",
			"Tenant":       "cgrates.org",
			"ToR":          "*monetary",
			"Usage":        mock.args.CGREvents[0].Event["Usage"],
			"test":         "val",
		},
		APIOpts: map[string]any{},
	}
	if !reflect.DeepEqual(expCgrEv, mock.args.CGREvents[0]) {
		t.Errorf("Expected: %+v \n,received: %+v", utils.ToJSON(expCgrEv), utils.ToJSON(mock.args.CGREvents[0]))
	}
}

func TestRemoteSetAccountAction(t *testing.T) {
	expError := `Post "127.1.0.11//": unsupported protocol scheme ""`
	if err = remoteSetAccount(nil, &Action{ExtraParameters: "127.1.0.11//"}, nil, nil, nil, SharedActionsData{}, ActionConnCfg{}); err == nil ||
		err.Error() != expError {
		t.Fatalf("Expected error: %s, received: %v", expError, err)
	}
	expError = `json: unsupported type: func()`
	if err = remoteSetAccount(&Account{
		ActionTriggers: ActionTriggers{{
			Balance: &BalanceFilter{
				Value: &utils.ValueFormula{
					Params: map[string]any{utils.MetaVoice: func() {}},
				},
			},
		}},
	}, &Action{ExtraParameters: "127.1.0.11//"}, nil, nil, nil, SharedActionsData{}, ActionConnCfg{}); err == nil ||
		err.Error() != expError {
		t.Fatalf("Expected error: %s, received: %v", expError, err)
	}

	ts := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) { rw.Write([]byte("5")) }))
	acc := &Account{ID: "1001"}
	expError = `json: cannot unmarshal number into Go value of type engine.Account`
	if err = remoteSetAccount(acc, &Action{ExtraParameters: ts.URL}, nil, nil, nil, SharedActionsData{}, ActionConnCfg{}); err == nil ||
		err.Error() != expError {
		t.Fatalf("Expected error: %s, received: %v", expError, err)
	}
	exp := &Account{ID: "1001"}
	if !reflect.DeepEqual(exp, acc) {
		t.Errorf("Expected: %s,received: %s", utils.ToJSON(exp), utils.ToJSON(acc))
	}
	ts.Close()

	acc = &Account{ID: "1001"}
	exp = &Account{
		ID: "1001",
		BalanceMap: map[string]Balances{
			utils.MetaVoice: {{
				ID:    "money",
				Value: 15,
			}},
		},
	}
	ts = httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		accStr := utils.ToJSON(acc) + "\n"
		val, err := io.ReadAll(r.Body)
		r.Body.Close()
		if err != nil {
			t.Error(err)
			return
		}
		if string(val) != accStr {
			t.Errorf("Expected %q,received: %q", accStr, string(val))
			return
		}
		rw.Write([]byte(utils.ToJSON(exp)))
	}))
	if err = remoteSetAccount(acc, &Action{ExtraParameters: ts.URL}, nil, nil, nil, SharedActionsData{}, ActionConnCfg{}); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(exp, acc) {
		t.Errorf("Expected: %s,received: %s", utils.ToJSON(exp), utils.ToJSON(acc))
	}
	ts.Close()
}

/**************** Benchmarks ********************************/

func BenchmarkUUID(b *testing.B) {
	m := make(map[string]int, 1000)
	for i := 0; i < b.N; i++ {
		uuid := utils.GenUUID()
		if len(uuid) == 0 {
			b.Fatalf("GenUUID error %s", uuid)
		}
		b.StopTimer()
		c := m[uuid]
		if c > 0 {
			b.Fatalf("duplicate uuid[%s] count %d", uuid, c)
		}
		m[uuid] = c + 1
		b.StartTimer()
	}
}

func TestResetAccountCDR(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	defer func() {
		config.SetCgrConfig(config.NewDefaultCGRConfig())
		SetCdrStorage(cdrStorage)
	}()
	idb, dErr := NewInternalDB(nil, nil, true, nil, cfg.DataDbCfg().Items)
	if dErr != nil {
		t.Error(dErr)
	}
	dm := NewDataManager(idb, cfg.CacheCfg(), nil)
	fltrs := NewFilterS(cfg, nil, dm)
	SetCdrStorage(idb)
	var extraData any
	acc := &Account{
		ID: "cgrates.org:1001",
		BalanceMap: map[string]Balances{
			utils.MetaMonetary: {
				&Balance{Value: 20},
			},
		},
		UnitCounters: UnitCounters{
			utils.MetaMonetary: []*UnitCounter{
				{
					Counters: CounterFilters{
						&CounterFilter{Value: 1},
					},
				},
			},
		},
	}
	a := &Action{
		Id:              "CDRLog1",
		ActionType:      utils.CDRLog,
		ExtraParameters: "{\"BalanceID\":\"~*acnt.BalanceID\",\"ActionID\":\"~*act.ActionID\",\"BalanceValue\":\"~*acnt.BalanceValue\"}",
		Weight:          50,
	}
	acs := Actions{
		a,
		&Action{
			Id:         "CdrDebit",
			ActionType: "*debit",
			Balance: &BalanceFilter{
				ID:     utils.StringPointer(utils.MetaDefault),
				Value:  &utils.ValueFormula{Static: 9.95},
				Type:   utils.StringPointer(utils.MetaMonetary),
				Weight: utils.Float64Pointer(0),
			},
			Weight:       float64(90),
			balanceValue: 10,
		},
	}
	if err := resetAccountCDR(nil, a, acs, fltrs, extraData, SharedActionsData{}, ActionConnCfg{}); err == nil || err.Error() != "nil account" {
		t.Errorf("expected <nil account> ,received <%+v>", err)
	} else if err = resetAccountCDR(acc, a, acs, fltrs, extraData, SharedActionsData{}, ActionConnCfg{}); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	}
	SetCdrStorage(nil)
	if err := resetAccountCDR(acc, a, acs, fltrs, extraData, SharedActionsData{}, ActionConnCfg{}); err == nil || err.Error() != fmt.Sprintf("nil cdrStorage for %s action", utils.ToJSON(a)) {
		t.Error(err)
	}

}

func TestSetRecurrentAction(t *testing.T) {

	ub := &Account{
		ID: "ACCID",
		ActionTriggers: ActionTriggers{
			&ActionTrigger{
				ID:        "acTrigger",
				UniqueID:  "uuid_acc",
				Recurrent: false,
			},
			&ActionTrigger{
				ID:        "acTrigger1",
				UniqueID:  "uuid_acc1",
				Recurrent: false,
			},
		},
	}
	ac := &Action{
		Id: "acTrigger",
	}
	if err = setRecurrentAction(ub, ac, nil, nil, nil, SharedActionsData{}, ActionConnCfg{}); err != nil {
		t.Error(err)
	}
}

func TestActionSetDDestinations(t *testing.T) {
	tmpDm := dm
	defer func() {
		dm = tmpDm
	}()
	cfg := config.NewDefaultCGRConfig()
	cfg.RalsCfg().StatSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.StatSConnsCfg)}
	cfg.GeneralCfg().DefaultTenant = "cgrates.org"
	cfg.DataDbCfg().Items = map[string]*config.ItemOpt{
		utils.CacheDestinations: {
			Limit:     3,
			StaticTTL: true,
		},
		utils.MetaReverseDestinations: {
			Limit:     3,
			Replicate: false,
		},
	}
	ccMock := &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.StatSv1GetStatQueue: func(ctx *context.Context, args, reply any) error {
				rpl := &StatQueue{
					Tenant: "cgrates",
					ID:     "id",
					SQMetrics: map[string]StatMetric{
						utils.MetaDDC: &StatDDC{
							FilterIDs: []string{"filters"},
							Count:     7,
						},
					},
					SQItems: []SQItem{
						{
							EventID: "event1",
						}, {
							EventID: "event2",
						},
					},
				}
				*reply.(*StatQueue) = *rpl
				return nil
			},
		},
	}
	clientconn := make(chan birpc.ClientConnector, 1)
	clientconn <- ccMock
	connMgr := NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.StatSConnsCfg): clientconn,
	})
	SetConnManager(connMgr)
	config.SetCgrConfig(cfg)
	ub := &Account{
		ID: "ACCID",
		ActionTriggers: ActionTriggers{
			&ActionTrigger{
				ID:        "acTrigger",
				UniqueID:  "uuid_acc",
				Recurrent: false,
			},
			&ActionTrigger{
				ID:        "acTrigger1",
				UniqueID:  "uuid_acc1",
				Recurrent: false,
			},
		},
		BalanceMap: map[string]Balances{
			utils.MetaMonetary: {

				&Balance{Value: 10,
					DestinationIDs: utils.StringMap{

						"*ddc:fr":  true,
						"*ddc:ger": false,
					}},
			},
			utils.MetaVoice: {
				&Balance{Value: 10, Weight: 20, DestinationIDs: utils.NewStringMap("NAT")},
				&Balance{Weight: 10, DestinationIDs: utils.NewStringMap("RET")},
			},
		},
		UnitCounters: UnitCounters{
			utils.MetaMonetary: []*UnitCounter{
				{
					Counters: CounterFilters{
						&CounterFilter{Value: 1},
					},
				},
			},
		},
	}
	a := &Action{
		Id:              "CDRLog1",
		ActionType:      utils.CDRLog,
		ExtraParameters: "{\"BalanceID\";\"~*acnt.BalanceID\";\"ActionID\";\"~*act.ActionID\";\"BalanceValue\";\"~*acnt.BalanceValue\"}",
		Weight:          50,
	}
	acs := Actions{
		a,
		&Action{
			Id:         "CdrDebit",
			ActionType: "*debit",
			Balance: &BalanceFilter{
				ID:     utils.StringPointer(utils.MetaDefault),
				Value:  &utils.ValueFormula{Static: 9.95},
				Type:   utils.StringPointer(utils.MetaMonetary),
				Weight: utils.Float64Pointer(0),
			},
			Weight:       float64(90),
			balanceValue: 10,
		},
	}
	if err := dm.dataDB.SetDestinationDrv(&Destination{
		Id: "*ddc:fr",
	}, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	if err := dm.dataDB.SetDestinationDrv(&Destination{
		Id: "*ddc:ger",
	}, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	SetDataStorage(dm)

	if err := setddestinations(ub, a, acs, nil, nil, SharedActionsData{}, ActionConnCfg{}); err != nil {
		t.Error(err)
	}

}

func TestActionPublishAccount(t *testing.T) {
	utils.Logger.SetLogLevel(4)
	utils.Logger.SetSyslog(nil)
	buf := new(bytes.Buffer)
	log.SetOutput(buf)
	cfg := config.NewDefaultCGRConfig()
	tmpCfg := cfg
	defer func() {
		utils.Logger.SetLogLevel(0)
		log.SetOutput(os.Stderr)
		config.SetCgrConfig(tmpCfg)
		SetConnManager(nil)
	}()
	cfg.RalsCfg().ThresholdSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.ThreshSConnsCfg)}
	cfg.RalsCfg().StatSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.StatSConnsCfg)}
	clientConn := make(chan birpc.ClientConnector, 1)
	clientConn <- &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.ThresholdSv1ProcessEvent: func(ctx *context.Context, args, reply any) error {
				*reply.(*[]string) = []string{"*thr"}
				return errors.New("Can't publish!")
			},
			utils.StatSv1ProcessEvent: func(ctx *context.Context, args, reply any) error {
				*reply.(*[]string) = []string{"*stat"}
				return errors.New("Can't publish!")
			},
		},
	}
	connMgr := NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.ThreshSConnsCfg): clientConn,
		utils.ConcatenatedKey(utils.MetaInternal, utils.StatSConnsCfg):   clientConn,
	})

	SetConnManager(connMgr)
	config.SetCgrConfig(cfg)
	ub := &Account{
		ID: "ACCID",
		ActionTriggers: ActionTriggers{
			&ActionTrigger{
				ID:        "acTrigger",
				UniqueID:  "uuid_acc",
				Recurrent: false,
			},
			&ActionTrigger{
				ID:        "acTrigger1",
				UniqueID:  "uuid_acc1",
				Recurrent: false,
			},
		},
		BalanceMap: map[string]Balances{
			utils.MetaMonetary: {
				&Balance{Value: 10,
					DestinationIDs: utils.StringMap{

						"*ddc_dest": true,
						"*dest":     false,
					}},
			},
			utils.MetaVoice: {
				&Balance{Value: 10, Weight: 20, DestinationIDs: utils.NewStringMap("NAT")},
				&Balance{Weight: 10, DestinationIDs: utils.NewStringMap("RET")},
			},
		},
		UnitCounters: UnitCounters{
			utils.MetaMonetary: []*UnitCounter{
				{
					Counters: CounterFilters{
						&CounterFilter{Value: 1},
					},
				},
			},
		},
	}
	a := &Action{
		Id:              "CDRLog1",
		ActionType:      utils.CDRLog,
		ExtraParameters: "{\"BalanceID\":\"~*acnt.BalanceID\",\"ActionID\":\"~*act.ActionID\",\"BalanceValue\":\"~*acnt.BalanceValue\"}",
		Weight:          50,
	}
	acs := Actions{
		a,
		&Action{
			Id:         "CdrDebit",
			ActionType: "*debit",
			Balance: &BalanceFilter{
				ID:     utils.StringPointer(utils.MetaDefault),
				Value:  &utils.ValueFormula{Static: 9.95},
				Type:   utils.StringPointer(utils.MetaMonetary),
				Weight: utils.Float64Pointer(0),
			},
			Weight:       float64(90),
			balanceValue: 10,
		},
	}
	expLog := ` with ThresholdS`
	expLog2 := `with StatS.`
	if err := publishAccount(ub, a, acs, nil, nil, SharedActionsData{}, ActionConnCfg{}); err != nil {
		t.Errorf("received %v", err)
	} else if rcvLog := buf.String(); !strings.Contains(rcvLog, expLog) {
		t.Errorf("Logger %v doesn't contain %v", rcvLog, expLog)
	} else if rcvLog := buf.String(); !strings.Contains(rcvLog, expLog2) {
		t.Errorf("Logger %v doesn't contain %v", rcvLog, expLog2)
	}
}

func TestExportAction(t *testing.T) {

	cfg := config.NewDefaultCGRConfig()
	cfg.ApierCfg().EEsConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.EEsConnsCfg)}
	config.SetCgrConfig(cfg)
	ccMock := &ccMock{
		calls: map[string]func(ctx *context.Context, args, reply any) error{
			utils.EeSv1ProcessEvent: func(ctx *context.Context, args, reply any) error {
				rpl := &map[string]map[string]any{}
				*reply.(*map[string]map[string]any) = *rpl

				return nil
			},
		},
	}
	clientconn := make(chan birpc.ClientConnector, 1)
	clientconn <- ccMock
	connMgr := NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.EEsConnsCfg): clientconn,
	})
	SetConnManager(connMgr)
	ub := &Account{
		ID: "ACCID",
		ActionTriggers: ActionTriggers{
			&ActionTrigger{
				ID:        "acTrigger",
				UniqueID:  "uuid_acc",
				Recurrent: false,
			},
			&ActionTrigger{
				ID:        "acTrigger1",
				UniqueID:  "uuid_acc1",
				Recurrent: false,
			},
		},
		BalanceMap: map[string]Balances{
			utils.MetaMonetary: {

				&Balance{Value: 10,
					DestinationIDs: utils.StringMap{

						"*ddc_dest": true,
						"*dest":     false,
					}},
			},
			utils.MetaVoice: {
				&Balance{Value: 10, Weight: 20, DestinationIDs: utils.NewStringMap("NAT")},
				&Balance{Weight: 10, DestinationIDs: utils.NewStringMap("RET")},
			},
		},
		UnitCounters: UnitCounters{
			utils.MetaMonetary: []*UnitCounter{
				{
					Counters: CounterFilters{
						&CounterFilter{Value: 1},
					},
				},
			},
		},
	}
	a := &Action{
		Id:              "CDRLog1",
		ActionType:      utils.CDRLog,
		ExtraParameters: "{\"BalanceID\":\"~*acnt.BalanceID\",\"ActionID\":\"~*act.ActionID\",\"BalanceValue\":\"~*acnt.BalanceValue\"}",
		Weight:          50,
	}
	acs := Actions{
		a,
		&Action{
			Id:         "CdrDebit",
			ActionType: "*debit",
			Balance: &BalanceFilter{
				ID:     utils.StringPointer(utils.MetaDefault),
				Value:  &utils.ValueFormula{Static: 9.95},
				Type:   utils.StringPointer(utils.MetaMonetary),
				Weight: utils.Float64Pointer(0),
			},
			Weight:       float64(90),
			balanceValue: 10,
		},
	}
	extraData := &utils.CGREvent{
		Tenant:  "tenant",
		ID:      "id1",
		Time:    utils.TimePointer(time.Date(2022, 12, 1, 1, 0, 0, 0, time.UTC)),
		Event:   map[string]any{},
		APIOpts: map[string]any{},
	}
	if err := export(ub, a, acs, nil, nil, SharedActionsData{}, ActionConnCfg{}); err != nil {
		t.Errorf("received %v", err)
	} else if err = export(nil, a, acs, nil, extraData, SharedActionsData{}, ActionConnCfg{}); err != nil {
		t.Errorf("received %v", err)
	} else if err = export(nil, a, acs, nil, "test", SharedActionsData{}, ActionConnCfg{}); err != nil {
		t.Error(err)
	} else if err = export(nil, a, acs, nil, nil, SharedActionsData{}, ActionConnCfg{}); err != nil {
		t.Error(err)
	}
}
func TestResetStatQueue(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.SchedulerCfg().StatSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.StatSConnsCfg)}
	ccMock := &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.StatSv1ResetStatQueue: func(ctx *context.Context, args, reply any) error {
				rpl := "reset"
				*reply.(*string) = rpl
				return nil
			},
		},
	}
	clientconn := make(chan birpc.ClientConnector, 1)
	clientconn <- ccMock
	connMgr := NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.StatSConnsCfg): clientconn,
	})
	SetConnManager(connMgr)
	config.SetCgrConfig(cfg)
	ub := &Account{}
	a := &Action{
		ExtraParameters: "cgrates.org:id",
	}
	acs := Actions{}
	if err := resetStatQueue(ub, a, acs, nil, nil, SharedActionsData{}, ActionConnCfg{}); err == nil {
		t.Errorf("received <%+v>", err)
	}

}

func TestResetTreshold(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()

	cfg.SchedulerCfg().ThreshSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.ThreshSConnsCfg)}
	ccMock := &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.ThresholdSv1ResetThreshold: func(ctx *context.Context, args, reply any) error {
				rpl := "threshold_reset"
				*reply.(*string) = rpl
				return nil
			},
		},
	}
	clientconn := make(chan birpc.ClientConnector, 1)
	clientconn <- ccMock
	connMgr := NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.ThreshSConnsCfg): clientconn,
	})
	SetConnManager(connMgr)
	config.SetCgrConfig(cfg)
	ub := &Account{}
	a := &Action{
		ExtraParameters: "cgrates.org:id",
	}
	acs := Actions{}
	if err := resetThreshold(ub, a, acs, nil, nil, SharedActionsData{}, ActionConnCfg{}); err == nil {
		t.Errorf("received <%+v>", err)
	}

}

func TestEnableDisableAccountAction(t *testing.T) {

	var acc *Account
	expErr := "nil account"
	if err := enableAccountAction(acc, nil, nil, nil, nil, SharedActionsData{}, ActionConnCfg{}); err == nil || err.Error() != expErr {
		t.Errorf("expected %+v ,received %v", expErr, err)
	} else if err = disableAccountAction(acc, nil, nil, nil, nil, SharedActionsData{}, ActionConnCfg{}); err == nil || err.Error() != expErr {
		t.Errorf("expected %+v ,received %v", expErr, err)
	} else if err = genericDebit(acc, nil, true, nil); err == nil || err.Error() != expErr {
		t.Errorf("expected %+v ,received %v", expErr, err)
	} else if err = resetCountersAction(acc, nil, nil, nil, nil, SharedActionsData{}, ActionConnCfg{}); err == nil || err.Error() != expErr {
		t.Errorf("expected %+v ,received %v", expErr, err)
	} else if err = debitAction(acc, nil, nil, nil, nil, SharedActionsData{}, ActionConnCfg{}); err == nil || err.Error() != expErr {
		t.Errorf("expected %+v ,received %v", expErr, err)
	} else if err = debitResetAction(acc, nil, nil, nil, nil, SharedActionsData{}, ActionConnCfg{}); err == nil || err.Error() != expErr {
		t.Errorf("expected %+v ,received %v", expErr, err)
	} else if err = topupAction(acc, nil, nil, nil, nil, SharedActionsData{}, ActionConnCfg{}); err == nil || err.Error() != expErr {
		t.Errorf("expected %+v ,received %v", expErr, err)
	} else if err = topupResetAction(acc, nil, nil, nil, nil, SharedActionsData{}, ActionConnCfg{}); err == nil || err.Error() != expErr {
		t.Errorf("expected %+v ,received %v", expErr, err)
	} else if err = resetAccountAction(acc, nil, nil, nil, nil, SharedActionsData{}, ActionConnCfg{}); err == nil || err.Error() != expErr {
		t.Errorf("expected %+v ,received %v", expErr, err)
	} else if err = denyNegativeAction(acc, nil, nil, nil, nil, SharedActionsData{}, ActionConnCfg{}); err == nil || err.Error() != expErr {
		t.Errorf("expected %+v ,received %v", expErr, err)
	} else if err = allowNegativeAction(acc, nil, nil, nil, nil, SharedActionsData{}, ActionConnCfg{}); err == nil || err.Error() != expErr {
		t.Errorf("expected %+v ,received %v", expErr, err)
	} else if err = unsetRecurrentAction(acc, nil, nil, nil, nil, SharedActionsData{}, ActionConnCfg{}); err == nil || err.Error() != expErr {
		t.Errorf("expected %+v ,received %v", expErr, err)
	} else if err = setRecurrentAction(acc, nil, nil, nil, nil, SharedActionsData{}, ActionConnCfg{}); err == nil || err.Error() != expErr {
		t.Errorf("expected %+v ,received %v", expErr, err)
	} else if err = resetTriggersAction(acc, nil, nil, nil, nil, SharedActionsData{}, ActionConnCfg{}); err == nil || err.Error() != expErr {
		t.Errorf("expected %+v ,received %v", expErr, err)
	}
}

func TestResetAccountCDRSuccesful(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	idb, dErr := NewInternalDB(nil, nil, true, nil, cfg.DataDbCfg().Items)
	if dErr != nil {
		t.Error(dErr)
	}
	dm := NewDataManager(idb, cfg.CacheCfg(), nil)
	fltrs := NewFilterS(cfg, nil, dm)
	cdr := &CDR{
		CGRID:       "Cdr1",
		OrderID:     123,
		ToR:         utils.MetaVoice,
		OriginID:    "OriginCDR1",
		OriginHost:  "192.168.1.1",
		Source:      "test",
		RequestType: utils.MetaRated,
		Category:    "call",
		Account:     "1001",
		Subject:     "1001",
		Destination: "+4986517174963",
		RunID:       utils.MetaDefault,
		Usage:       time.Duration(0),
		ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
		Cost:        1.01,
		CostDetails: &EventCost{
			CGRID:     "ecId",
			RunID:     "ecRunId",
			StartTime: time.Date(2022, 12, 1, 12, 0, 0, 0, time.UTC),
			Usage:     utils.DurationPointer(1 * time.Hour),
			Cost:      utils.Float64Pointer(12.1),
			Charges:   []*ChargingInterval{},
			AccountSummary: &AccountSummary{
				Tenant: "cgrates.org",
				ID:     "acc_Id",
				BalanceSummaries: BalanceSummaries{
					{
						UUID:     "uuid",
						ID:       "summary_id",
						Type:     "type",
						Initial:  1,
						Value:    12,
						Disabled: true,
					}, {},
				},
				AllowNegative: true,
				Disabled:      false,
			},
			Accounting:    Accounting{},
			RatingFilters: RatingFilters{},
			Rates:         ChargedRates{},
		},
	}
	if err := idb.SetCDR(cdr, true); err != nil {
		t.Error(err)
	}

	SetCdrStorage(idb)
	var extraData any
	acc := &Account{
		ID: "cgrates.org:1001",
		BalanceMap: map[string]Balances{
			utils.MetaMonetary: {
				&Balance{Value: 20},
			},
		},
		UnitCounters: UnitCounters{
			utils.MetaMonetary: []*UnitCounter{
				{
					Counters: CounterFilters{
						&CounterFilter{Value: 1},
					},
				},
			},
		},
	}
	a := &Action{
		Id:              "CDRLog1",
		ActionType:      utils.CDRLog,
		ExtraParameters: "{\"BalanceID\":\"~*acnt.BalanceID\",\"ActionID\":\"~*act.ActionID\",\"BalanceValue\":\"~*acnt.BalanceValue\"}",
		Weight:          50,
	}
	acs := Actions{
		a,
		&Action{
			Id:         "CdrDebit",
			ActionType: "*debit",
			Balance: &BalanceFilter{
				ID:     utils.StringPointer(utils.MetaDefault),
				Value:  &utils.ValueFormula{Static: 9.95},
				Type:   utils.StringPointer(utils.MetaMonetary),
				Weight: utils.Float64Pointer(0),
			},
			Weight:       float64(90),
			balanceValue: 10,
		},
	}
	if err = resetAccountCDR(acc, a, acs, fltrs, extraData, SharedActionsData{}, ActionConnCfg{}); err != nil {
		t.Error(err)
	}

}

func TestRemoveSessionCost(t *testing.T) {
	tmp := Cache
	tmpCdr := cdrStorage
	utils.Logger.SetLogLevel(4)
	utils.Logger.SetSyslog(nil)
	buf := new(bytes.Buffer)
	log.SetOutput(buf)
	defer func() {
		utils.Logger.SetLogLevel(0)
		log.SetOutput(os.Stderr)
		Cache = tmp
		cdrStorage = tmpCdr
	}()
	Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	action := &Action{
		ExtraParameters: "*acnt.BalanceID;*act.ActionID",
	}
	Cache.Set(utils.CacheFilters, utils.ConcatenatedKey(cfg.GeneralCfg().DefaultTenant, "*acnt.BalanceID"), &Filter{
		Tenant: "tnt",
		Rules: []*FilterRule{
			{
				Values:  []string{"val1,val2"},
				Type:    utils.MetaString,
				Element: utils.MetaScPrefix + utils.CGRID},
			{
				Values:  []string{"val1,val2"},
				Type:    utils.MetaString,
				Element: "test"},
		},
	}, []string{"grpId"}, true, utils.NonTransactional)

	expLog := `for filter`
	expLog2 := `in action:`
	if err := removeSessionCosts(nil, action, nil, nil, nil, SharedActionsData{}, ActionConnCfg{}); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	} else if rcvLog := buf.String(); !strings.Contains(rcvLog, expLog) {
		t.Errorf("expected %v,received %v", expLog, rcvLog)
	} else if rcvLog := buf.String(); !strings.Contains(rcvLog, expLog2) {
		t.Errorf("expected %v,received %v", expLog, rcvLog)
	}

}

func TestLogAction(t *testing.T) {
	acc := &Account{
		ID: "cgrates.org:1001",
		BalanceMap: map[string]Balances{
			utils.MetaMonetary: {
				&Balance{Value: 20},
			},
		},
		UnitCounters: UnitCounters{
			utils.MetaMonetary: []*UnitCounter{
				{
					Counters: CounterFilters{
						&CounterFilter{Value: 1},
					},
				},
			},
		},
	}
	extraData := map[string]any{
		"field1": "value",
		"field2": "second",
	}
	if err := logAction(acc, nil, nil, nil, nil, SharedActionsData{}, ActionConnCfg{}); err != nil {
		t.Error(err)
	} else if err = logAction(nil, nil, nil, nil, extraData, SharedActionsData{}, ActionConnCfg{}); err != nil {
		t.Error(err)
	}

}

func TestCdrLogProviderFieldAsInterface(t *testing.T) {
	acc := &Account{
		ID: "ACCID",
		ActionTriggers: ActionTriggers{
			&ActionTrigger{
				ID:        "acTrigger",
				UniqueID:  "uuid_acc",
				Recurrent: false,
			},
			&ActionTrigger{
				ID:        "acTrigger1",
				UniqueID:  "uuid_acc1",
				Recurrent: false,
			},
		},
		BalanceMap: map[string]Balances{
			utils.MetaMonetary: {

				&Balance{Value: 10,
					DestinationIDs: utils.StringMap{

						"*ddc_dest": true,
						"*dest":     false,
					}},
			},
			utils.MetaVoice: {
				&Balance{Value: 10, Weight: 20, DestinationIDs: utils.NewStringMap("NAT")},
				&Balance{Weight: 10, DestinationIDs: utils.NewStringMap("RET")},
			},
		},
		UnitCounters: UnitCounters{
			utils.MetaMonetary: []*UnitCounter{
				{
					Counters: CounterFilters{
						&CounterFilter{Value: 1},
					},
				},
			},
		},
	}
	a := &Action{

		Id:              "CDRLog1",
		ActionType:      utils.CDRLog,
		ExtraParameters: "{\"BalanceID\":\"~*acnt.BalanceID\",\"ActionID\":\"~*act.ActionID\",\"BalanceValue\":\"~*acnt.BalanceValue\"}",
		Weight:          50,
		Balance: &BalanceFilter{
			Uuid:          utils.StringPointer("uuid113"),
			ID:            utils.StringPointer("b_id_22"),
			Type:          utils.StringPointer("*prepaid"),
			RatingSubject: utils.StringPointer("rate"),
			DestinationIDs: &utils.StringMap{
				"dest1": true,
			},
			Value: &utils.ValueFormula{
				Static: 3.0,
			},
			Categories: &utils.StringMap{
				"category": true,
			},
			SharedGroups: &utils.StringMap{
				"group1": true,
			},
			ExpirationDate: utils.TimePointer(time.Date(2022, 1, 1, 2, 0, 0, 0, time.UTC)),
			Weight:         utils.Float64Pointer(323.0),
		},
	}
	cd := &cdrLogProvider{action: a,
		acnt: acc,
		cache: utils.MapStorage{
			"field": "val",
		}}
	if val, err := cd.FieldAsInterface([]string{utils.MetaAcnt, utils.AccountID}); err != nil {
		t.Error(err)
	} else if val != acc.ID {
		t.Errorf("expected %v,received %v", acc.ID, val)
	}
	if _, has := cd.cache[utils.MetaAcnt]; !has {
		t.Error("field does not exist")
	}
	if val, err := cd.FieldAsInterface([]string{utils.MetaAcnt, utils.BalanceUUID}); err != nil {
		t.Error(err)
	} else if val != *a.Balance.Uuid {
		t.Errorf("expected %v,received %v", *a.Balance.Uuid, val)
	}
	if _, has := cd.cache[utils.MetaAcnt]; !has {
		t.Error("field does not exist")
	}

	if val, err := cd.FieldAsInterface([]string{utils.MetaAcnt, utils.DestinationIDs}); err != nil {
		t.Error(err)
	} else if val != a.Balance.DestinationIDs.String() {
		t.Errorf("expected %v,received %v", *a.Balance.Uuid, val)
	}
	if _, has := cd.cache[utils.MetaAcnt]; !has {
		t.Error("field does not exist")
	}

	if val, err := cd.FieldAsInterface([]string{utils.MetaAcnt, utils.ExtraParameters}); err != nil {
		t.Error()
	} else if val != a.ExtraParameters {
		t.Errorf("expected %v,received %v", *a.Balance.Uuid, val)
	}
	if _, has := cd.cache[utils.MetaAcnt]; !has {
		t.Error("field does not exist")
	}

	if val, err := cd.FieldAsInterface([]string{utils.MetaAcnt, utils.RatingSubject}); err != nil {
		t.Error()
	} else if val != *a.Balance.RatingSubject {
		t.Errorf("expected %v,received %v", *a.Balance.Uuid, val)
	}
	if _, has := cd.cache[utils.MetaAcnt]; !has {
		t.Error("field does not exist")
	}

	if val, err := cd.FieldAsInterface([]string{utils.MetaAcnt, utils.Category}); err != nil {
		t.Error()
	} else if val != a.Balance.Categories.String() {
		t.Errorf("expected %v,received %v", *a.Balance.Uuid, val)
	}
	if _, has := cd.cache[utils.MetaAcnt]; !has {
		t.Error("field does not exist")
	}

	if val, err := cd.FieldAsInterface([]string{utils.MetaAcnt, utils.SharedGroups}); err != nil {
		t.Error()
	} else if val != a.Balance.SharedGroups.String() {
		t.Errorf("expected %v,received %v", *a.Balance.Uuid, val)
	}
	if _, has := cd.cache[utils.MetaAcnt]; !has {
		t.Error("field does not exist")
	}

	if val, err := cd.FieldAsInterface([]string{utils.MetaAct, utils.ActionType}); err != nil {
		t.Error()
	} else if val != a.ActionType {
		t.Errorf("expected %v,received %v", *a.Balance.Uuid, val)
	}
	if _, has := cd.cache[utils.MetaAct]; !has {
		t.Error("field does not exist")
	}
	if val, err := cd.FieldAsInterface([]string{"val"}); err != nil {
		t.Error()
	} else if val != "val" {
		t.Errorf("expected %v,received %v", *a.Balance.Uuid, val)
	}
	if _, has := cd.cache["val"]; !has {
		t.Error("field does not exist")
	}
}

func TestRemoveAccountAcc(t *testing.T) {
	a := &Action{
		Id:              "CDRLog1",
		ActionType:      utils.CDRLog,
		ExtraParameters: "{\"BalanceID\":\"~*acnt.BalanceID\",\"ActionID\":\"~*act.ActionID\",\"BalanceValue\":\"~*acnt.BalanceValue\"}",
		Weight:          50,
	}
	acs := Actions{
		a,
		&Action{
			Id:         "CdrDebit",
			ActionType: "*debit",
			Balance: &BalanceFilter{
				ID:     utils.StringPointer(utils.MetaDefault),
				Value:  &utils.ValueFormula{Static: 9.95},
				Type:   utils.StringPointer(utils.MetaMonetary),
				Weight: utils.Float64Pointer(0),
			},
			Weight:       float64(90),
			balanceValue: 10,
		},
	}
	extraData := &utils.CGREvent{
		Tenant:  "tenant",
		ID:      "id1",
		Time:    utils.TimePointer(time.Date(2022, 12, 1, 1, 0, 0, 0, time.UTC)),
		Event:   map[string]any{},
		APIOpts: map[string]any{},
	}
	if err := removeAccountAction(nil, a, acs, nil, extraData, SharedActionsData{}, ActionConnCfg{}); err != nil {
		t.Error(err)
	}
}

func TestRemoveAccountActionErr(t *testing.T) {
	tmp := Cache
	tmpDm := dm

	setLogger := func(buf *bytes.Buffer) {
		utils.Logger.SetLogLevel(4)
		utils.Logger.SetSyslog(nil)
		log.SetOutput(buf)
	}
	removeLogger := func() {
		utils.Logger.SetLogLevel(0)
		log.SetOutput(os.Stderr)
	}
	buf := new(bytes.Buffer)
	setLogger(buf)
	cfg := config.NewDefaultCGRConfig()
	defer func() {
		removeLogger()
		Cache = tmp
		SetDataStorage(tmpDm)

		config.SetCgrConfig(config.NewDefaultCGRConfig())
	}()
	cfg.DataDbCfg().Items = map[string]*config.ItemOpt{
		utils.CacheAccounts: {
			Limit:  3,
			TTL:    2 * time.Minute,
			Remote: true,
		},
		utils.CacheAccountActionPlans: {
			Limit:     3,
			StaticTTL: true,
			Remote:    true,
		},
		utils.CacheActionPlans: {
			Remote:    true,
			Limit:     3,
			StaticTTL: true,
		},
	}
	cfg.DataDbCfg().RmtConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator)}
	clientConn := make(chan birpc.ClientConnector, 1)
	clientConn <- &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.ReplicatorSv1GetAccountActionPlans: func(ctx *context.Context, args, reply any) error {

				return errors.New("ActionPlans not found")
			},
			utils.ReplicatorSv1GetActionPlan: func(ctx *context.Context, args, reply any) error {
				return errors.New("ActionPlan not found")
			},
		},
	}
	connMgr := NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator): clientConn,
	})
	a := &Action{
		Id:              "CDRLog1",
		ActionType:      utils.CDRLog,
		ExtraParameters: "{\"BalanceID\":\"~*acnt.BalanceID\",\"ActionID\":\"~*act.ActionID\",\"BalanceValue\":\"~*acnt.BalanceValue\"}",
		Weight:          50,
	}
	acs := Actions{
		a,
		&Action{
			Id:         "CdrDebit",
			ActionType: "*debit",
			Balance: &BalanceFilter{
				ID:     utils.StringPointer(utils.MetaDefault),
				Value:  &utils.ValueFormula{Static: 9.95},
				Type:   utils.StringPointer(utils.MetaMonetary),
				Weight: utils.Float64Pointer(0),
			},
			Weight:       float64(90),
			balanceValue: 10,
		},
	}
	extraData := &utils.CGREvent{
		Tenant:  "tenant",
		ID:      "id1",
		Time:    utils.TimePointer(time.Date(2022, 12, 1, 1, 0, 0, 0, time.UTC)),
		Event:   map[string]any{},
		APIOpts: map[string]any{},
	}
	ub := &Account{
		BalanceMap: map[string]Balances{
			utils.MetaMonetary: {&Balance{
				Value: 10,
			}},
		},
	}
	db, dErr := NewInternalDB(nil, nil, true, nil, cfg.DataDbCfg().Items)
	if dErr != nil {
		t.Error(dErr)
	}
	dm := NewDataManager(db, cfg.CacheCfg(), connMgr)
	SetDataStorage(nil)
	if err := removeAccountAction(ub, a, acs, nil, extraData, SharedActionsData{}, ActionConnCfg{}); err == nil || err != utils.ErrInvalidKey {
		t.Error(err)
	}
	ub.ID = "cgrates.org:exp"
	expLog := `[ERROR] Could not remove account Id: cgrates.org:exp: NO_DATABASE_CONNECTION`
	if err := removeAccountAction(ub, a, acs, nil, extraData, SharedActionsData{}, ActionConnCfg{}); err == nil {
		t.Error(err)
	} else if rcvLog := buf.String(); !strings.Contains(rcvLog, expLog) {
		t.Errorf("expected log <%+v> to be included in: <%+v>",
			expLog, rcvLog)
	}
	SetDataStorage(dm)
	config.SetCgrConfig(cfg)
	ub.ID = "acc_id"
	if err := dm.SetAccount(ub); err != nil {
		t.Error(err)
	}
	removeLogger()
	buf2 := new(bytes.Buffer)
	setLogger(buf2)
	expLog = `Could not get action plans`
	if err := removeAccountAction(ub, a, acs, nil, extraData, SharedActionsData{}, ActionConnCfg{}); err == nil {
		t.Error(err)
	} else if rcvLog := buf2.String(); !strings.Contains(rcvLog, expLog) {
		t.Errorf("Logger %v doesn't contain %v", rcvLog, expLog)
	}
	removeLogger()
	buf3 := new(bytes.Buffer)
	setLogger(buf3)
	expLog = `Could not retrieve action plan:`
	dm.SetAccountActionPlans(ub.ID, []string{"acc1"}, true)
	if err := removeAccountAction(ub, a, acs, nil, extraData, SharedActionsData{}, ActionConnCfg{}); err == nil {
		t.Error(err)
	} else if rcvLog := buf3.String(); !strings.Contains(rcvLog, expLog) {
		t.Errorf("Logger %v doesn't contain %v", rcvLog, expLog)
	}
	// if err := dm.SetActionPlan("acc1", &ActionPlan{
	// 	ActionTimings: []*ActionTiming{
	// 		{ActionsID: "ENABLE_ACNT"},
	// 	},
	// }, true, utils.NonTransactional); err != nil {
	// 	t.Error(err)
	// }

}

func TestRemoveExpiredErrs(t *testing.T) {
	var acc *Account
	action := &Action{
		Id:               "MINI",
		ActionType:       utils.MetaTopUpReset,
		ExpirationString: utils.MetaUnlimited,
		ExtraParameters:  "",
		Weight:           10,
		Balance: &BalanceFilter{
			Type:           utils.StringPointer(utils.MetaMonetary),
			Uuid:           utils.StringPointer("uuid"),
			Value:          &utils.ValueFormula{Static: 10},
			Weight:         utils.Float64Pointer(10),
			DestinationIDs: nil,
			TimingIDs:      nil,
			SharedGroups:   nil,
			Categories:     nil,
			Disabled:       utils.BoolPointer(false),
			Blocker:        utils.BoolPointer(false),
		},
	}
	if err := removeExpired(acc, action, nil, nil, nil, SharedActionsData{}, ActionConnCfg{}); err == nil || err.Error() != fmt.Sprintf("nil account for %s action", utils.ToJSON(action)) {
		t.Error(err)
	}
	acc = &Account{
		ID:         "cgrates.org:rembal2",
		BalanceMap: map[string]Balances{},
		Disabled:   true,
	}
	if err = removeExpired(acc, action, nil, nil, nil, SharedActionsData{}, ActionConnCfg{}); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	}
	acc.BalanceMap = map[string]Balances{
		utils.MetaMonetary: {
			&Balance{
				Value: 10,
			},
			&Balance{
				Value:          10,
				DestinationIDs: utils.NewStringMap("NAT", "RET"),
				ExpirationDate: time.Date(2050, time.November, 11, 22, 39, 0, 0, time.UTC),
			},
			&Balance{
				Value:          10,
				DestinationIDs: utils.NewStringMap("NAT", "RET"),
				ExpirationDate: time.Date(2050, time.November, 15, 22, 39, 0, 0, time.UTC),
			},
			&Balance{
				Value:          10,
				DestinationIDs: utils.NewStringMap("NAT", "RET"),
				ExpirationDate: time.Date(2050, time.November, 11, 22, 39, 0, 0, time.UTC),
			},
		},
	}
	if err := removeExpired(acc, action, nil, nil, nil, SharedActionsData{}, ActionConnCfg{}); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	}
}

func TestTransferMonetaryDefaultAction(t *testing.T) {
	utils.Logger.SetLogLevel(3)
	utils.Logger.SetSyslog(nil)
	buf := new(bytes.Buffer)
	log.SetOutput(buf)
	defer func() {
		utils.Logger.SetLogLevel(0)
		log.SetOutput(os.Stderr)
	}()
	a := &Action{
		Id:              "CDRLog1",
		ActionType:      utils.CDRLog,
		ExtraParameters: "{\"BalanceID\":\"~*acnt.BalanceID\",\"ActionID\":\"~*act.ActionID\",\"BalanceValue\":\"~*acnt.BalanceValue\"}",
		Weight:          50,
	}
	acs := Actions{
		&Action{
			Id:           "CdrDebit",
			ActionType:   "*debit",
			Weight:       float64(90),
			balanceValue: 10,
		},
	}
	expLog := `*transfer_monetary_default called without account`
	if err := transferMonetaryDefaultAction(nil, a, acs, nil, "data", SharedActionsData{}, ActionConnCfg{}); err == nil || err != utils.ErrAccountNotFound {
		t.Errorf("expected <%v>,received <%v>", utils.ErrAccountNotFound, err)
	} else if rcvLog := buf.String(); !strings.Contains(rcvLog, expLog) {
		t.Errorf("expected log <%+v> to be included in: <%+v>",
			expLog, rcvLog)
	}
	ub := &Account{}
	if err := transferMonetaryDefaultAction(ub, a, acs, nil, "data", SharedActionsData{}, ActionConnCfg{}); err == nil || err != utils.ErrNotFound {
		t.Errorf("expected <%v>,received <%v>", utils.ErrNotFound, err)
	}
}

func TestRemoveBalanceActionErr(t *testing.T) {
	acc := &Account{
		ID: "vdf:minu",
		BalanceMap: map[string]Balances{
			utils.MetaMonetary: {&Balance{Value: 50}},
			utils.MetaVoice: {
				&Balance{Value: 200 * float64(time.Second),
					ExpirationDate: time.Date(2022, 11, 22, 2, 0, 0, 0, time.UTC),
					DestinationIDs: utils.NewStringMap("NAT"), Weight: 10},
				&Balance{Value: 100 * float64(time.Second),
					DestinationIDs: utils.NewStringMap("RET"), Weight: 20},
			},
		},
	}
	acs := &Action{
		Balance: &BalanceFilter{
			ExpirationDate: utils.TimePointer(time.Date(2022, 11, 12, 2, 0, 0, 0, time.UTC)),
			Type:           utils.StringPointer(utils.MetaMonetary),
			Value:          &utils.ValueFormula{Static: 10},
		},
	}
	if err := removeBalanceAction(nil, acs, nil, nil, nil, SharedActionsData{}, ActionConnCfg{}); err == nil {
		t.Error(err)
	}
	if err := removeBalanceAction(acc, acs, nil, nil, nil, SharedActionsData{}, ActionConnCfg{}); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	}
}

func TestDebitResetAction(t *testing.T) {

	ub := &Account{
		ID: "OUT:CUSTOMER_1:rif",
		BalanceMap: map[string]Balances{
			utils.MetaVoice:    {&Balance{Value: 21}},
			utils.MetaMonetary: {&Balance{Value: 21}},
		},
	}
	a := &Action{
		Id:               "MINI",
		ActionType:       utils.MetaTopUpReset,
		ExpirationString: utils.MetaUnlimited,
		ExtraParameters:  "",
		Weight:           10,
		Balance: &BalanceFilter{
			Type:           utils.StringPointer(utils.MetaMonetary),
			Uuid:           utils.StringPointer("uuid"),
			Value:          &utils.ValueFormula{Static: 10},
			Weight:         utils.Float64Pointer(10),
			DestinationIDs: nil,
			TimingIDs:      nil,
			SharedGroups:   nil,
			Categories:     nil,
			Disabled:       utils.BoolPointer(false),
			Blocker:        utils.BoolPointer(false),
		},
	}
	if err := debitResetAction(ub, a, nil, nil, nil, SharedActionsData{}, ActionConnCfg{}); err != nil {
		t.Error(err)
	}
}

func TestSetDestinationsErr(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	tempConn := connMgr
	tmpDm := dm
	defer func() {
		config.SetCgrConfig(config.NewDefaultCGRConfig())
		SetConnManager(tempConn)
		dm = tmpDm
	}()
	cfg.RalsCfg().StatSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaStats)}
	cfg.DataDbCfg().RplConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator)}
	cfg.DataDbCfg().Items = map[string]*config.ItemOpt{
		utils.MetaReverseDestinations: {
			Replicate: true,
		},
		utils.MetaDestinations: {
			Replicate: false,
		},
	}
	db, dErr := NewInternalDB(nil, nil, true, nil, cfg.DataDbCfg().Items)
	if dErr != nil {
		t.Error(dErr)
	}
	clientConn := make(chan birpc.ClientConnector, 1)
	clientConn <- &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.StatSv1GetStatQueue: func(ctx *context.Context, args, reply any) error {
				rpl := StatQueue{
					Tenant: "cgrates.org",
					ID:     "StatsID",
					SQItems: []SQItem{{
						EventID: "ev1",
					}},
					SQMetrics: map[string]StatMetric{
						utils.MetaDDC: &StatDDC{
							FilterIDs: []string{"Test_Filter_ID"},
							FieldValues: map[string]utils.StringSet{
								"1001": {
									"EVENT_1": {},
								},
								"1002": {
									"EVENT_3": {},
								},
							},
							Events: map[string]map[string]int64{
								"Event1": {
									"FieldValue1": 1,
								},
								"Event2": {},
							},
							MinItems: 3,
							Count:    3,
						},
					},
				}
				*reply.(*StatQueue) = rpl
				return nil
			},
			utils.ReplicatorSv1SetReverseDestination: func(ctx *context.Context, args, reply any) error {
				return utils.ErrNotImplemented
			},
		},
	}
	connMgr := NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaStats):      clientConn,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator): clientConn,
	})
	dm := NewDataManager(db, cfg.CacheCfg(), connMgr)
	ub := &Account{
		ID: "cgrates.org:1001",
		BalanceMap: map[string]Balances{
			utils.MetaVoice: {
				{
					ID:    "1001",
					Value: 0,
					DestinationIDs: utils.StringMap{
						utils.MetaDDC + "TEST": true,
					},
				},
			},
			utils.MetaMonetary: {
				{
					ID:    utils.MetaDefault,
					Value: -1.8,
				},
			},
		},
	}
	a := &Action{
		Id:               "TOPUP_RST_GNR_1000",
		ActionType:       utils.MetaTopUpReset,
		ExtraParameters:  `{"*voice": 60.0,"*data":1024.0,"*sms":1.0}`,
		Weight:           10,
		ExpirationString: utils.MetaUnlimited,
	}
	acs := Actions{
		a,
	}
	dest := &Destination{Id: utils.MetaDDC + "TEST", Prefixes: []string{"+491", "+492", "+493"}}
	if err := dm.SetDestination(dest, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	config.SetCgrConfig(cfg)
	SetConnManager(connMgr)
	SetDataStorage(dm)
	if err := setddestinations(ub, a, acs, nil, nil, SharedActionsData{}, ActionConnCfg{}); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	}
	ub = &Account{
		ID: "cgrates.org:1001",
		BalanceMap: map[string]Balances{
			utils.MetaMonetary: {
				&Balance{Value: 20},
			},
		},
	}
	if err := setddestinations(ub, a, acs, nil, nil, SharedActionsData{}, ActionConnCfg{}); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	}
}

func TestRemoveAccountActionLogg(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	tempConn := connMgr
	tmpDm := dm
	tmpCache := Cache
	defer func() {
		config.SetCgrConfig(config.NewDefaultCGRConfig())
		SetConnManager(tempConn)
		dm = tmpDm
		Cache = tmpCache
	}()
	Cache.Clear(nil)
	cfg.DataDbCfg().Items = map[string]*config.ItemOpt{
		utils.CacheAccountActionPlans: {
			Remote: true,
		},
		utils.MetaAccounts: {
			Replicate: false,
		},
		utils.MetaActionPlans: {
			Remote:    true,
			Replicate: true,
		},
	}
	cfg.DataDbCfg().RmtConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator)}
	cfg.DataDbCfg().RplConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator)}
	db, dErr := NewInternalDB(nil, nil, true, nil, cfg.DataDbCfg().Items)
	if dErr != nil {
		t.Error(dErr)
	}
	clientConn := make(chan birpc.ClientConnector, 1)
	clientConn <- &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.ReplicatorSv1GetAccountActionPlans: func(ctx *context.Context, args, reply any) error {
				rpl := []string{"PACKAGE_10_SHARED_A_5"}
				*reply.(*[]string) = rpl
				return nil
			},
			utils.ReplicatorSv1GetActionPlan: func(ctx *context.Context, args, reply any) error {
				rpl := ActionPlan{
					Id: "PACKAGE_10_SHARED_A_5",
					AccountIDs: utils.StringMap{
						"cgrates.org:1001": true,
					},
				}
				*reply.(**ActionPlan) = &rpl
				return nil
			},
			utils.ReplicatorSv1SetActionPlan: func(ctx *context.Context, args, reply any) error {
				return utils.ErrNotFound
			},
		}}
	connMgr := NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator): clientConn,
	})
	dm := NewDataManager(db, cfg.CacheCfg(), connMgr)
	ub := &Account{
		ID: "cgrates.org:1001",
		BalanceMap: map[string]Balances{
			utils.MetaVoice: {
				{
					ID:    "1001",
					Value: 0,
					DestinationIDs: utils.StringMap{
						utils.MetaDDC + "TEST": true,
					},
				},
			},
			utils.MetaMonetary: {
				{
					ID:    utils.MetaDefault,
					Value: -1.8,
				},
			},
		},
	}
	a := &Action{
		Id:               "TOPUP_RST_GNR_1000",
		ActionType:       utils.MetaTopUpReset,
		ExtraParameters:  `{"*voice": 60.0,"*data":1024.0,"*sms":1.0}`,
		Weight:           10,
		ExpirationString: utils.MetaUnlimited,
	}
	acs := Actions{
		a,
	}
	SetDataStorage(dm)
	config.SetCgrConfig(cfg)
	if err := removeAccountAction(ub, a, acs, nil, nil, SharedActionsData{}, ActionConnCfg{}); err != nil {
		t.Error(err)
	}

}

func TestActionsTransferBalance(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	var srcAcc *Account
	var destAcc *Account
	db := &DataDBMock{
		SetAccountDrvF: func(acc *Account) error {
			switch {
			case strings.HasSuffix(acc.ID, "_FAIL"):
				return utils.ErrServerError
			case srcAcc.ID == acc.ID:
				srcAcc = acc
			case destAcc.ID == acc.ID:
				destAcc = acc
			default:
				return utils.ErrAccountNotFound
			}
			return nil
		},
		GetAccountDrvF: func(id string) (*Account, error) {
			if destAcc == nil || destAcc.ID != id {
				return nil, utils.ErrNotFound
			}
			return destAcc, nil
		},
	}
	mockedDM := NewDataManager(db, cfg.CacheCfg(), nil)
	originalDM := dm
	defer func() { dm = originalDM }()
	dm = mockedDM

	testcases := []struct {
		name                string
		srcAcc              *Account
		destAcc             *Account
		act                 *Action
		expectedSrcBalance  float64
		expectedDestBalance float64
		expectedErr         string
	}{
		{
			name: "SuccessfulTransfer",
			srcAcc: &Account{
				ID: "cgrates.org:ACC_SRC",
				BalanceMap: map[string]Balances{
					utils.MetaMonetary: {
						{
							ID:    "BALANCE_SRC",
							Value: 10,
						},
					},
				},
			},
			destAcc: &Account{
				ID: "cgrates.org:ACC_DEST",
				BalanceMap: map[string]Balances{
					utils.MetaMonetary: {
						{
							ID:    "BALANCE_DEST",
							Value: 5,
						},
					},
				},
			},
			act: &Action{
				Id:         "ACT_TRANSFER_BALANCE",
				ActionType: utils.MetaTransferBalance,
				ExtraParameters: `{
					"DestinationAccountID": "cgrates.org:ACC_DEST",
					"DestinationBalanceID": "BALANCE_DEST"
				}`,
				Balance: &BalanceFilter{
					ID: utils.StringPointer("BALANCE_SRC"),
					Value: &utils.ValueFormula{
						Static: 3,
					},
				},
			},
			expectedSrcBalance:  7,
			expectedDestBalance: 8,
		},
		{
			name: "SuccessfulTransferNegativeSrcBalance",
			srcAcc: &Account{
				ID: "cgrates.org:ACC_SRC",
				BalanceMap: map[string]Balances{
					utils.MetaMonetary: {
						{
							ID:    "*default",
							Value: 2,
						},
					},
				},
			},
			destAcc: &Account{
				ID: "cgrates.org:ACC_DEST",
				BalanceMap: map[string]Balances{
					utils.MetaMonetary: {
						{
							ID:    "BALANCE_DEST",
							Value: 5,
						},
					},
				},
			},
			act: &Action{
				Id:         "ACT_TRANSFER_BALANCE",
				ActionType: utils.MetaTransferBalance,
				ExtraParameters: `{
					"DestinationAccountID": "cgrates.org:ACC_DEST",
					"DestinationBalanceID": "BALANCE_DEST"
				}`,
				Balance: &BalanceFilter{
					ID: utils.StringPointer("*default"),
					Value: &utils.ValueFormula{
						Static: 3,
					},
				},
			},
			expectedSrcBalance:  -1,
			expectedDestBalance: 8,
		},
		{
			name: "SuccessfulTransferNegativeDestBalance",
			srcAcc: &Account{
				ID: "cgrates.org:ACC_SRC",
				BalanceMap: map[string]Balances{
					utils.MetaMonetary: {
						{
							ID:    "BALANCE_SRC",
							Value: 5,
						},
					},
				},
			},
			destAcc: &Account{
				ID: "cgrates.org:ACC_DEST",
				BalanceMap: map[string]Balances{
					utils.MetaMonetary: {
						{
							ID:    "*default",
							Value: 2,
						},
					},
				},
			},
			act: &Action{
				Id:         "ACT_TRANSFER_BALANCE",
				ActionType: utils.MetaTransferBalance,
				ExtraParameters: `{
					"DestinationAccountID": "cgrates.org:ACC_DEST",
					"DestinationBalanceID": "*default"
				}`,
				Balance: &BalanceFilter{
					ID: utils.StringPointer("BALANCE_SRC"),
					Value: &utils.ValueFormula{
						Static: -3,
					},
				},
			},
			expectedSrcBalance:  8,
			expectedDestBalance: -1,
		},
		{
			name: "SuccessfulTransferRefNegativeSrcBalance",
			srcAcc: &Account{
				ID: "cgrates.org:ACC_SRC",
				BalanceMap: map[string]Balances{
					utils.MetaMonetary: {
						{
							ID:    "*default",
							Value: 2,
						},
					},
				},
			},
			destAcc: &Account{
				ID: "cgrates.org:ACC_DEST",
				BalanceMap: map[string]Balances{
					utils.MetaMonetary: {
						{
							ID:    "BALANCE_DEST",
							Value: 5,
						},
					},
				},
			},
			act: &Action{
				Id:         "ACT_TRANSFER_BALANCE",
				ActionType: utils.MetaTransferBalance,
				ExtraParameters: `{
					"DestinationAccountID": "cgrates.org:ACC_DEST",
					"DestinationBalanceID": "BALANCE_DEST",
					"DestinationReferenceValue": 8
				}`,
				Balance: &BalanceFilter{
					ID: utils.StringPointer("*default"),
				},
			},
			expectedSrcBalance:  -1,
			expectedDestBalance: 8,
		},
		{
			name: "SuccessfulTransferRefNegativeDestBalance",
			srcAcc: &Account{
				ID: "cgrates.org:ACC_SRC",
				BalanceMap: map[string]Balances{
					utils.MetaMonetary: {
						{
							ID:    "BALANCE_SRC",
							Value: 10,
						},
					},
				},
			},
			destAcc: &Account{
				ID: "cgrates.org:ACC_DEST",
				BalanceMap: map[string]Balances{
					utils.MetaMonetary: {
						{
							ID:    "*default",
							Value: 5,
						},
					},
				},
			},
			act: &Action{
				Id:         "ACT_TRANSFER_BALANCE",
				ActionType: utils.MetaTransferBalance,
				ExtraParameters: `{
					"DestinationAccountID": "cgrates.org:ACC_DEST",
					"DestinationBalanceID": "*default",
					"DestinationReferenceValue": -0.01
				}`,
				Balance: &BalanceFilter{
					ID: utils.StringPointer("BALANCE_SRC"),
					Value: &utils.ValueFormula{
						Static: 3,
					},
				},
			},
			expectedSrcBalance:  15.01,
			expectedDestBalance: -0.01,
		},
		{
			name:        "NilAccount",
			expectedErr: "source account is nil",
		},
		{
			name: "UnspecifiedBalanceID",
			srcAcc: &Account{
				ID: "cgrates.org:ACC_SRC",
			},
			act: &Action{
				Id:      "ACT_TRANSFER_BALANCE",
				Balance: &BalanceFilter{},
			},
			expectedErr: "source balance ID is missing",
		},
		{
			name: "MissingDestinationParameters",
			srcAcc: &Account{
				ID: "cgrates.org:ACC_SRC",
			},
			act: &Action{
				Id: "ACT_TRANSFER_BALANCE",
				Balance: &BalanceFilter{
					ID: utils.StringPointer("BALANCE_SRC"),
				},
			},
			expectedErr: "ExtraParameters used to identify the destination balance are missing",
		},
		{
			name: "MissingSourceAccountBalances",
			srcAcc: &Account{
				ID: "cgrates.org:ACC_SRC",
			},
			act: &Action{
				Id:              "ACT_TRANSFER_BALANCE",
				ExtraParameters: "invalid_params",
				Balance: &BalanceFilter{
					ID: utils.StringPointer("BALANCE_SRC"),
				},
			},
			expectedErr: "account cgrates.org:ACC_SRC has no balances to transfer from",
		},
		{
			name: "SourceAccountBalanceNotFound",
			srcAcc: &Account{
				ID: "cgrates.org:ACC_SRC",
				BalanceMap: map[string]Balances{
					utils.MetaMonetary: {
						{
							ID:    "*default",
							Value: 10,
						},
					},
				},
			},
			act: &Action{
				Id:              "ACT_TRANSFER_BALANCE",
				ExtraParameters: "invalid_params",
				Balance: &BalanceFilter{
					ID: utils.StringPointer("BALANCE_SRC"),
				},
			},
			expectedErr: "source balance not found or expired",
		},
		{
			name: "InvalidExtraParameters",
			srcAcc: &Account{
				ID: "cgrates.org:ACC_SRC",
				BalanceMap: map[string]Balances{
					utils.MetaMonetary: {
						{
							ID:    "BALANCE_SRC",
							Value: 10,
						},
					},
				},
			},
			act: &Action{
				Id:              "ACT_TRANSFER_BALANCE",
				ExtraParameters: "invalid_params",
				Balance: &BalanceFilter{
					ID: utils.StringPointer("BALANCE_SRC"),
					Value: &utils.ValueFormula{
						Static: 3,
					},
				},
			},
			expectedErr: "invalid character 'i' looking for beginning of value",
		},
		{
			name: "DestinationAccountNotFound",
			srcAcc: &Account{
				ID: "cgrates.org:ACC_SRC",
				BalanceMap: map[string]Balances{
					utils.MetaMonetary: {
						{
							ID:    "BALANCE_SRC",
							Value: 10,
						},
					},
				},
			},
			act: &Action{
				Id: "ACT_TRANSFER_BALANCE",
				ExtraParameters: `{
					"DestinationAccountID": "cgrates.org:ACC_DEST",
					"DestinationBalanceID": "BALANCE_DEST"
				}`,
				Balance: &BalanceFilter{
					ID: utils.StringPointer("BALANCE_SRC"),
					Value: &utils.ValueFormula{
						Static: 3,
					},
				},
			},
			expectedErr: "retrieving destination account failed: NOT_FOUND",
		},
		{
			name: "DestinationBalanceNotFound",
			srcAcc: &Account{
				ID: "cgrates.org:ACC_SRC",
				BalanceMap: map[string]Balances{
					utils.MetaMonetary: {
						{
							ID:    "BALANCE_SRC",
							Value: 10,
						},
					},
				},
			},
			destAcc: &Account{
				ID: "cgrates.org:ACC_DEST",
			},
			act: &Action{
				Id: "ACT_TRANSFER_BALANCE",
				ExtraParameters: `{
					"DestinationAccountID": "cgrates.org:ACC_DEST",
					"DestinationBalanceID": "BALANCE_DEST"
				}`,
				Balance: &BalanceFilter{
					ID: utils.StringPointer("BALANCE_SRC"),
					Value: &utils.ValueFormula{
						Static: 3,
					},
				},
			},
			expectedSrcBalance:  7,
			expectedDestBalance: 3,
		},
		{
			name: "TransferUnitsNotSpecifiedOr0",
			srcAcc: &Account{
				ID: "cgrates.org:ACC_SRC",
				BalanceMap: map[string]Balances{
					utils.MetaMonetary: {
						{
							ID:    "BALANCE_SRC",
							Value: 1,
						},
					},
				},
			},
			destAcc: &Account{
				ID: "cgrates.org:ACC_DEST",
				BalanceMap: map[string]Balances{
					utils.MetaMonetary: {
						{
							ID:    "BALANCE_DEST",
							Value: 5,
						},
					},
				},
			},
			act: &Action{
				Id: "ACT_TRANSFER_BALANCE",
				ExtraParameters: `{
					"DestinationAccountID": "cgrates.org:ACC_DEST",
					"DestinationBalanceID": "BALANCE_DEST"
				}`,
				Balance: &BalanceFilter{
					ID: utils.StringPointer("BALANCE_SRC"),
					Value: &utils.ValueFormula{
						Static: 0,
					},
				},
			},
			expectedErr: "transfer amount is missing or 0",
		},
		{
			name: "NotEnoughFundsInSourceBalance",
			srcAcc: &Account{
				ID: "cgrates.org:ACC_SRC",
				BalanceMap: map[string]Balances{
					utils.MetaMonetary: {
						{
							ID:    "BALANCE_SRC",
							Value: 1,
						},
					},
				},
			},
			destAcc: &Account{
				ID: "cgrates.org:ACC_DEST",
				BalanceMap: map[string]Balances{
					utils.MetaMonetary: {
						{
							ID:    "BALANCE_DEST",
							Value: 5,
						},
					},
				},
			},
			act: &Action{
				Id: "ACT_TRANSFER_BALANCE",
				ExtraParameters: `{
					"DestinationAccountID": "cgrates.org:ACC_DEST",
					"DestinationBalanceID": "BALANCE_DEST"
				}`,
				Balance: &BalanceFilter{
					ID: utils.StringPointer("BALANCE_SRC"),
					Value: &utils.ValueFormula{
						Static: 3,
					},
				},
			},
			expectedErr: `insufficient credits in source balance "BALANCE_SRC" (account "cgrates.org:ACC_SRC") for transfer of 3.00 units`,
		},
		{
			name: "NotEnoughFundsInDestBalance",
			srcAcc: &Account{
				ID: "cgrates.org:ACC_SRC",
				BalanceMap: map[string]Balances{
					utils.MetaMonetary: {
						{
							ID:    "BALANCE_SRC",
							Value: 1,
						},
					},
				},
			},
			destAcc: &Account{
				ID: "cgrates.org:ACC_DEST",
				BalanceMap: map[string]Balances{
					utils.MetaMonetary: {
						{
							ID:    "BALANCE_DEST",
							Value: 5,
						},
					},
				},
			},
			act: &Action{
				Id: "ACT_TRANSFER_BALANCE",
				ExtraParameters: `{
					"DestinationAccountID": "cgrates.org:ACC_DEST",
					"DestinationBalanceID": "BALANCE_DEST",
					"DestinationReferenceValue": -0.01
				}`,
				Balance: &BalanceFilter{
					ID: utils.StringPointer("BALANCE_SRC"),
					Value: &utils.ValueFormula{ // should be ignored (overwritten by the ReferenceValue)
						Static: 3,
					},
				},
			},
			expectedErr: `insufficient credits in destination balance "BALANCE_DEST" (account "cgrates.org:ACC_DEST") for transfer of -5.01 units`,
		},
		{
			name: "DestinationBalanceFailedUpdate",
			srcAcc: &Account{
				ID: "cgrates.org:ACC_SRC",
				BalanceMap: map[string]Balances{
					utils.MetaMonetary: {
						{
							ID:    "BALANCE_SRC",
							Value: 10,
						},
					},
				},
			},
			destAcc: &Account{
				ID: "cgrates.org:ACC_DEST_FAIL",
				BalanceMap: map[string]Balances{
					utils.MetaMonetary: {
						{
							ID:    "BALANCE_DEST",
							Value: 5,
						},
					},
				},
			},
			act: &Action{
				Id: "ACT_TRANSFER_BALANCE",
				ExtraParameters: `{
					"DestinationAccountID": "cgrates.org:ACC_DEST_FAIL",
					"DestinationBalanceID": "BALANCE_DEST"
				}`,
				Balance: &BalanceFilter{
					ID: utils.StringPointer("BALANCE_SRC"),
					Value: &utils.ValueFormula{
						Static: 3,
					},
				},
			},
			expectedErr: "updating destination account failed: SERVER_ERROR",
		},
	}

	verifyBalance := func(t *testing.T, acc *Account, expected float64, accType, balanceType string, balanceIdx int) {
		t.Helper()
		balance := acc.BalanceMap[balanceType]
		balanceValue := balance[balanceIdx].Value
		if balanceValue != expected {
			t.Errorf("received wrong %s balance value: expected %v, received %v",
				accType, expected, balanceValue)
		}
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			srcAcc = tc.srcAcc
			destAcc = tc.destAcc
			err := transferBalanceAction(srcAcc, tc.act, nil, nil, nil, SharedActionsData{}, ActionConnCfg{})
			if tc.expectedErr != "" {
				if err == nil || err.Error() != tc.expectedErr {
					t.Errorf("expected error %v, received %v", tc.expectedErr, err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			verifyBalance(t, srcAcc, tc.expectedSrcBalance, "source", utils.MetaMonetary, 0)
			verifyBalance(t, destAcc, tc.expectedDestBalance, "destination", utils.MetaMonetary, 0)
		})
	}
}

func TestActionsAlterAndDisconnectSessions(t *testing.T) {
	tempConn := connMgr
	tmpDm := dm
	tmpCache := Cache
	defer func() {
		config.SetCgrConfig(config.NewDefaultCGRConfig())
		SetConnManager(tempConn)
		dm = tmpDm
		Cache = tmpCache
	}()
	Cache.Clear(nil)
	var alterSessionsRequest string
	var disconnectSessionsRequest string
	ccMock := &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.SessionSv1AlterSessions: func(ctx *context.Context, args, reply any) error {
				alterSessionsRequest = utils.ToJSON(args)
				return nil
			},
			utils.SessionSv1ForceDisconnect: func(ctx *context.Context, args, reply any) error {
				disconnectSessionsRequest = utils.ToJSON(args)
				return nil
			},
		},
	}
	connID := utils.ConcatenatedKey(utils.MetaInternal, utils.MetaSessionS)
	clientconn := make(chan birpc.ClientConnector, 1)
	clientconn <- ccMock
	NewConnManager(config.NewDefaultCGRConfig(), map[string]chan birpc.ClientConnector{
		connID: clientconn,
	})
	testcases := []struct {
		name            string
		extraParams     string
		connIDs         []string
		expectedRequest string
		expectedErr     string
	}{
		{
			name:            "SuccessfulRequest",
			connIDs:         []string{connID},
			expectedRequest: `{"Limit":1,"Filters":["*string:~*req.Account:1001","*prefix:~*req.Destination:+40"],"Tenant":"tenant.com","APIOpts":{"*radCoATemplate":"mytemplate","secondopt":"secondval"},"Event":{"Account":"1002","Destination":"+40123456"}}`,
			extraParams:     "tenant.com;*string:~*req.Account:1001&*prefix:~*req.Destination:+40;1;*radCoATemplate:mytemplate&secondopt:secondval;Account:1002&Destination:+40123456",
		},
		{
			name:            "FailedRequest",
			expectedRequest: `{"Limit":1,"Filters":["*string:~*req.Account:1001","*prefix:~*req.Destination:+40"],"Tenant":"tenant.com","APIOpts":{"*radCoATemplate":"mytemplate","secondopt":"secondval"},"Event":{"Account":"1002","Destination":"+40123456"}}`,
			extraParams:     "tenant.com;*string:~*req.Account:1001&*prefix:~*req.Destination:+40;1;*radCoATemplate:mytemplate&secondopt:secondval;Account:1002&Destination:+40123456",
			expectedErr:     "MANDATORY_IE_MISSING: [connIDs]",
		},
		{
			name:        "WrongNumberOfParams",
			extraParams: "tenant;;1;",
			expectedErr: "invalid number of parameters <4> expected 5",
		},
		{
			name:        "InvalidEventMap",
			extraParams: "tenant;;1;opt:value;key",
			expectedErr: "invalid key-value pair: key",
		},
		{
			name:        "InvalidOptsMap",
			extraParams: "tenant;;1;opt;key:value",
			expectedErr: "invalid key-value pair: opt",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			action := &Action{ExtraParameters: tc.extraParams}
			t.Cleanup(func() {
				alterSessionsRequest = ""
				disconnectSessionsRequest = ""
			})
			err := alterSessionsAction(nil, action, nil, nil, nil,
				SharedActionsData{}, ActionConnCfg{
					ConnIDs: tc.connIDs,
				})
			if tc.expectedErr != "" {
				if err == nil || err.Error() != tc.expectedErr {
					t.Errorf("expected error %v, received %v", tc.expectedErr, err)
				}
			} else if err != nil {
				t.Error(err)
			} else if alterSessionsRequest != tc.expectedRequest {
				t.Errorf("expected: %v\nreceived: %v", tc.expectedRequest, alterSessionsRequest)
			}
			err = forceDisconnectSessionsAction(nil, action, nil, nil, nil,
				SharedActionsData{}, ActionConnCfg{
					ConnIDs: tc.connIDs,
				})
			if tc.expectedErr != "" {
				if err == nil || err.Error() != tc.expectedErr {
					t.Errorf("expected error %v, received %v", tc.expectedErr, err)
				}
			} else if err != nil {
				t.Error(err)
			} else if disconnectSessionsRequest != tc.expectedRequest {
				t.Errorf("expected: %v\nreceived: %v", tc.expectedRequest, disconnectSessionsRequest)
			}
		})
	}
}

func TestActionsStringMethod(t *testing.T) {
	cdrProvider := &cdrLogProvider{}
	expectedJSON := "{}"
	jsonStr := cdrProvider.String()
	if jsonStr != expectedJSON {
		t.Errorf("String() method result does not match expected JSON. Got: %s, Expected: %s", jsonStr, expectedJSON)
	}
}

func TestActionsClone(t *testing.T) {
	actions := Actions{}
	cloned := actions.Clone()
	if len(cloned) != len(actions) {
		t.Errorf("Expected cloned Actions length %d, got %d", len(actions), len(cloned))
	}
	for i := range actions {
		if !reflect.DeepEqual(actions[i], cloned[i]) {
			t.Errorf("Expected cloned Action[%d] to match original, got %+v", i, cloned[i])
		}
	}
	cloned = (Actions)(nil).Clone()
	if cloned != nil {
		t.Errorf("Expected cloned result to be nil when apl is nil, got %+v", cloned)
	}
}

func TestActionsNewActionConnCfgThresholdSNonSessionAction(t *testing.T) {
	Config := &config.CGRConfig{}
	act := newActionConnCfg(utils.ThresholdS, "someOtherAction", Config)
	if len(act.ConnIDs) != 0 {
		t.Errorf("Expected ConnIDs length 0, got %d", len(act.ConnIDs))
	}
}

func TestDynamicThreshold(t *testing.T) {
	tempConn := connMgr
	tmpDm := dm
	tmpCache := Cache
	defer func() {
		config.SetCgrConfig(config.NewDefaultCGRConfig())
		SetConnManager(tempConn)
		dm = tmpDm
		Cache = tmpCache
	}()
	Cache.Clear(nil)
	var tpwo *ThresholdProfileWithAPIOpts
	ccMock := &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.APIerSv1SetThresholdProfile: func(ctx *context.Context, args, reply any) error {
				var canCast bool
				if tpwo, canCast = args.(*ThresholdProfileWithAPIOpts); !canCast {
					return fmt.Errorf("couldnt cast")
				}
				return nil
			},
		},
	}
	connID := utils.ConcatenatedKey(utils.MetaInternal, utils.MetaApier)
	clientconn := make(chan birpc.ClientConnector, 1)
	clientconn <- ccMock
	NewConnManager(config.NewDefaultCGRConfig(), map[string]chan birpc.ClientConnector{
		connID: clientconn,
	})
	testcases := []struct {
		name        string
		extraParams string
		connIDs     []string
		expTpwo     *ThresholdProfileWithAPIOpts
		expectedErr string
	}{
		{
			name:    "SuccessfulRequest",
			connIDs: []string{connID},
			expTpwo: &ThresholdProfileWithAPIOpts{
				ThresholdProfile: &ThresholdProfile{
					Tenant:    "cgrates.org",
					ID:        "THD_ACNT_1001",
					FilterIDs: []string{"*string:~*req.Account:1001"},
					ActivationInterval: &utils.ActivationInterval{
						ActivationTime: time.Date(2014, 7, 29, 15, 0, 0, 0, time.UTC),
					},
					MaxHits:   1,
					MinHits:   1,
					MinSleep:  1 * time.Second,
					Blocker:   true,
					Weight:    10,
					ActionIDs: []string{"ACT_LOG_WARNING"},
					Async:     true,
					EeIDs:     []string{"eeID1", "eeID2"},
				},
				APIOpts: map[string]any{
					"key": "value",
				},
			},
			extraParams: "cgrates.org;THD_ACNT_1001;*string:~*req.Account:1001;2014-07-29T15:00:00Z;1;1;1s;true;10;ACT_LOG_WARNING;true;eeID1&eeID2;key:value",
		},
		{
			name:    "SuccessfulRequestWithDynamicPaths",
			connIDs: []string{connID},
			expTpwo: &ThresholdProfileWithAPIOpts{
				ThresholdProfile: &ThresholdProfile{
					Tenant:    "cgrates.org",
					ID:        "THD_ACNT_1001",
					FilterIDs: []string{"*string:~*req.Account:1001"},
					ActivationInterval: &utils.ActivationInterval{
						ActivationTime: time.Now(),
						ExpiryTime:     time.Date(3000, 7, 29, 15, 0, 0, 0, time.UTC),
					},
					MaxHits:   1,
					MinHits:   1,
					MinSleep:  1 * time.Second,
					Blocker:   true,
					Weight:    10,
					ActionIDs: []string{"ACT_LOG_WARNING"},
					Async:     true,
					EeIDs:     []string{"eeID1", "eeID2"},
				},
				APIOpts: map[string]any{
					"key": "value",
				},
			},
			extraParams: "*tenant;THD_ACNT_<~*req.Account>;*string:~*req.Account:<~*req.Account>;*now&3000-07-29T15:00:00Z;1;1;1s;true;10;ACT_LOG_WARNING;true;eeID1&eeID2;key:value",
		},
		{
			name:    "SuccessfulRequestEmptyFields",
			connIDs: []string{connID},
			expTpwo: &ThresholdProfileWithAPIOpts{
				ThresholdProfile: &ThresholdProfile{
					Tenant:             "cgrates.org",
					ID:                 "THD_ACNT_1001",
					FilterIDs:          nil,
					ActivationInterval: &utils.ActivationInterval{},
					MaxHits:            0,
					MinHits:            0,
					MinSleep:           0,
					Blocker:            false,
					Weight:             0,
					ActionIDs:          nil,
					Async:              false,
					EeIDs:              nil,
				},
				APIOpts: map[string]any{},
			},
			extraParams: "cgrates.org;THD_ACNT_1001;;;;;;;;;;;",
		},
		{
			name:        "MissingConns",
			extraParams: "cgrates.org;THD_ACNT_1001;FLTR_ACNT_1001;2014-07-29T15:00:00Z;1;1;1s;false;10;ACT_LOG_WARNING;true;eeID1&eeID2;",
			expectedErr: "MANDATORY_IE_MISSING: [connIDs]",
		},
		{
			name:        "WrongNumberOfParams",
			extraParams: "tenant;;1;",
			expectedErr: "invalid number of parameters <4> expected 13",
		},
		{
			name:        "ActivationIntervalLengthFail",
			extraParams: "cgrates.org;THD_ACNT_1001;FLTR_ACNT_1001;2014-07-29T15:00:00Z&&;1;1;1s;false;10;ACT_LOG_WARNING;true;eeID1&eeID2;",
			expectedErr: utils.ErrUnsupportedFormat.Error(),
		},
		{
			name:        "ActivationIntervalBadStringFail",
			extraParams: "cgrates.org;THD_ACNT_1001;FLTR_ACNT_1001;bad String;1;1;1s;false;10;ACT_LOG_WARNING;true;eeID1&eeID2;",
			expectedErr: `parsing time "bad String" as "2006-01-02T15:04:05Z07:00": cannot parse "bad String" as "2006"`,
		},
		{
			name:        "ActivationIntervalBadStringFail2",
			extraParams: "cgrates.org;THD_ACNT_1001;FLTR_ACNT_1001;2014-07-29T15:00:00Z&bad String;1;1;1s;false;10;ACT_LOG_WARNING;true;eeID1&eeID2;",
			expectedErr: `parsing time "bad String" as "2006-01-02T15:04:05Z07:00": cannot parse "bad String" as "2006"`,
		},
		{
			name:        "MaxHitsFail",
			extraParams: "cgrates.org;THD_ACNT_1001;FLTR_ACNT_1001;2014-07-29T15:00:00Z;BadString;1;1s;false;10;ACT_LOG_WARNING;true;eeID1&eeID2;",
			expectedErr: `strconv.Atoi: parsing "BadString": invalid syntax`,
		},
		{
			name:        "MinHitsFail",
			extraParams: "cgrates.org;THD_ACNT_1001;FLTR_ACNT_1001;2014-07-29T15:00:00Z;1;BadString;1s;false;10;ACT_LOG_WARNING;true;eeID1&eeID2;",
			expectedErr: `strconv.Atoi: parsing "BadString": invalid syntax`,
		},
		{
			name:        "MinSleepFail",
			extraParams: "cgrates.org;THD_ACNT_1001;FLTR_ACNT_1001;2014-07-29T15:00:00Z;1;1;BadString;false;10;ACT_LOG_WARNING;true;eeID1&eeID2;",
			expectedErr: `time: invalid duration "BadString"`,
		},
		{
			name:        "BlockerFail",
			extraParams: "cgrates.org;THD_ACNT_1001;FLTR_ACNT_1001;2014-07-29T15:00:00Z;1;1;1s;BadString;10;ACT_LOG_WARNING;true;eeID1&eeID2;",
			expectedErr: `strconv.ParseBool: parsing "BadString": invalid syntax`,
		},
		{
			name:        "WeightFail",
			extraParams: "cgrates.org;THD_ACNT_1001;FLTR_ACNT_1001;2014-07-29T15:00:00Z;1;1;1s;false;BadString;ACT_LOG_WARNING;true;eeID1&eeID2;",
			expectedErr: `strconv.ParseFloat: parsing "BadString": invalid syntax`,
		},
		{
			name:        "AsyncFail",
			extraParams: "cgrates.org;THD_ACNT_1001;FLTR_ACNT_1001;2014-07-29T15:00:00Z;1;1;1s;false;10;ACT_LOG_WARNING;BadString;eeID1&eeID2;",
			expectedErr: `strconv.ParseBool: parsing "BadString": invalid syntax`,
		},
		{
			name:        "InvalidOptsMap",
			extraParams: "cgrates.org;THD_ACNT_1001;FLTR_ACNT_1001;2014-07-29T15:00:00Z;1;1;1s;false;10;ACT_LOG_WARNING;true;eeID1&eeID2;opt",
			expectedErr: "invalid key-value pair: opt",
		},
	}

	for i, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			action := &Action{ExtraParameters: tc.extraParams}
			ev := &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "evID",
				Time:   &time.Time{},
				Event: map[string]any{
					utils.AccountField: "1001",
				},
			}
			t.Cleanup(func() {
				tpwo = nil
			})
			err := dynamicThreshold(nil, action, nil, nil, ev,
				SharedActionsData{}, ActionConnCfg{
					ConnIDs: tc.connIDs,
				})
			if tc.expectedErr != "" {
				if err == nil || err.Error() != tc.expectedErr {
					t.Errorf("expected error <%v>, received <%v>", tc.expectedErr, err)
				}
			} else if err != nil {
				t.Error(err)
			} else if !reflect.DeepEqual(tpwo, tc.expTpwo) {
				if i != 1 {
					t.Errorf("Expected <%v>\nReceived\n<%v>", utils.ToJSON(tc.expTpwo), utils.ToJSON(tpwo))
				} else {
					// Get the absolute difference between the times
					diff := tpwo.ActivationInterval.ActivationTime.Sub(tc.expTpwo.ActivationInterval.ActivationTime)
					if diff < 0 {
						diff = -diff // Make sure it's positive
					}
					// Check if difference is less than or equal to 1 second
					if diff <= time.Second {
						tc.expTpwo.ActivationInterval.ActivationTime = tpwo.ActivationInterval.ActivationTime
						if !reflect.DeepEqual(tpwo, tc.expTpwo) {
							t.Errorf("Expected <%v>\nReceived\n<%v>", utils.ToJSON(tc.expTpwo), utils.ToJSON(tpwo))
						}
					} else {
						t.Errorf("Expected <%v>\nReceived\n<%v>", utils.ToJSON(tc.expTpwo), utils.ToJSON(tpwo))
					}
				}
			}
		})
	}
}

func TestDynamicStats(t *testing.T) {
	tempConn := connMgr
	tmpDm := dm
	tmpCache := Cache
	defer func() {
		config.SetCgrConfig(config.NewDefaultCGRConfig())
		SetConnManager(tempConn)
		dm = tmpDm
		Cache = tmpCache
	}()
	Cache.Clear(nil)
	var sqpwo *StatQueueProfileWithAPIOpts
	ccMock := &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.APIerSv1SetStatQueueProfile: func(ctx *context.Context, args, reply any) error {
				var canCast bool
				if sqpwo, canCast = args.(*StatQueueProfileWithAPIOpts); !canCast {
					return fmt.Errorf("couldnt cast")
				}
				return nil
			},
		},
	}
	connID := utils.ConcatenatedKey(utils.MetaInternal, utils.MetaApier)
	clientconn := make(chan birpc.ClientConnector, 1)
	clientconn <- ccMock
	NewConnManager(config.NewDefaultCGRConfig(), map[string]chan birpc.ClientConnector{
		connID: clientconn,
	})
	testcases := []struct {
		name        string
		extraParams string
		connIDs     []string
		expSqpwo    *StatQueueProfileWithAPIOpts
		expectedErr string
	}{
		{
			name:    "SuccessfulRequest",
			connIDs: []string{connID},
			expSqpwo: &StatQueueProfileWithAPIOpts{
				StatQueueProfile: &StatQueueProfile{
					Tenant:    "cgrates.org",
					ID:        "Stat_1",
					FilterIDs: []string{"FLTR_STAT_1"},
					ActivationInterval: &utils.ActivationInterval{
						ActivationTime: time.Date(2014, 7, 29, 15, 0, 0, 0, time.UTC),
					},
					QueueLength: 100,
					TTL:         10 * time.Second,
					MinItems:    0,
					Metrics: []*MetricWithFilters{
						{
							FilterIDs: []string{"Metric_FLTR"},
							MetricID:  "*acd",
						},
						{
							FilterIDs: []string{"Metric_FLTR"},
							MetricID:  "*tcd",
						},
						{
							FilterIDs: []string{"Metric_FLTR"},
							MetricID:  "*asr",
						},
					},
					Stored:       false,
					Blocker:      true,
					Weight:       30,
					ThresholdIDs: []string{"*none"},
				},
				APIOpts: map[string]any{
					"key": "value",
				},
			},
			extraParams: "cgrates.org;Stat_1;FLTR_STAT_1;2014-07-29T15:00:00Z;100;10s;0;*acd&*tcd&*asr;Metric_FLTR;false;true;30;*none;key:value",
		},
		{
			name:    "SuccessfulRequestWithDynamicPaths",
			connIDs: []string{connID},
			expSqpwo: &StatQueueProfileWithAPIOpts{
				StatQueueProfile: &StatQueueProfile{
					Tenant:    "cgrates.org",
					ID:        "Stat_1001",
					FilterIDs: []string{"*string:~*req.Account:1001"},
					ActivationInterval: &utils.ActivationInterval{
						ActivationTime: time.Now(),
						ExpiryTime:     time.Date(3000, 7, 29, 15, 0, 0, 0, time.UTC),
					},
					QueueLength: 100,
					TTL:         10 * time.Second,
					MinItems:    0,
					Metrics: []*MetricWithFilters{
						{
							FilterIDs: []string{"Metric_FLTR"},
							MetricID:  "*acd",
						},
						{
							FilterIDs: []string{"Metric_FLTR"},
							MetricID:  "*tcd",
						},
						{
							FilterIDs: []string{"Metric_FLTR"},
							MetricID:  "*asr",
						},
					},
					Stored:       false,
					Blocker:      true,
					Weight:       30,
					ThresholdIDs: []string{"*none"},
				},
				APIOpts: map[string]any{
					"key": "value",
				},
			},
			extraParams: "*tenant;Stat_<~*req.Account>;*string:~*req.Account:<~*req.Account>;*now&3000-07-29T15:00:00Z;100;10s;0;*acd&*tcd&*asr;Metric_FLTR;false;true;30;*none;key:value",
		},
		{
			name:    "SuccessfulRequestEmptyFields",
			connIDs: []string{connID},
			expSqpwo: &StatQueueProfileWithAPIOpts{
				StatQueueProfile: &StatQueueProfile{
					Tenant:             "cgrates.org",
					ID:                 "Stat_1",
					FilterIDs:          nil,
					ActivationInterval: &utils.ActivationInterval{},
					QueueLength:        0,
					TTL:                0,
					MinItems:           0,
					Metrics:            nil,
					Stored:             false,
					Blocker:            false,
					Weight:             0,
					ThresholdIDs:       nil,
				},
				APIOpts: map[string]any{},
			},
			extraParams: "cgrates.org;Stat_1;;;;;;;;;;;;",
		},
		{
			name:        "MissingConns",
			extraParams: "cgrates.org;Stat_1;FLTR_STAT_1;2014-07-29T15:00:00Z;100;10s;0;*acd&*tcd&*asr;Metric_FLTR;false;true;30;*none;key:value",
			expectedErr: "MANDATORY_IE_MISSING: [connIDs]",
		},
		{
			name:        "WrongNumberOfParams",
			extraParams: "tenant;;1;",
			expectedErr: "invalid number of parameters <4> expected 14",
		},
		{
			name:        "ActivationIntervalLengthFail",
			extraParams: "cgrates.org;Stat_1;FLTR_STAT_1;2014-07-29T15:00:00Z&&;100;10s;0;*acd&*tcd&*asr;Metric_FLTR;false;true;30;*none;key:value",
			expectedErr: utils.ErrUnsupportedFormat.Error(),
		},
		{
			name:        "ActivationIntervalBadStringFail",
			extraParams: "cgrates.org;Stat_1;FLTR_STAT_1;bad String;100;10s;0;*acd&*tcd&*asr;Metric_FLTR;false;true;30;*none;key:value",
			expectedErr: `parsing time "bad String" as "2006-01-02T15:04:05Z07:00": cannot parse "bad String" as "2006"`,
		},
		{
			name:        "ActivationIntervalBadStringFail2",
			extraParams: "cgrates.org;Stat_1;FLTR_STAT_1;2014-07-29T15:00:00Z&bad String;100;10s;0;*acd&*tcd&*asr;Metric_FLTR;false;true;30;*none;key:value",
			expectedErr: `parsing time "bad String" as "2006-01-02T15:04:05Z07:00": cannot parse "bad String" as "2006"`,
		},
		{
			name:        "QueueLengthFail",
			extraParams: "cgrates.org;Stat_1;FLTR_STAT_1;2014-07-29T15:00:00Z;BadString;10s;0;*acd&*tcd&*asr;Metric_FLTR;false;true;30;*none;key:value",
			expectedErr: `strconv.Atoi: parsing "BadString": invalid syntax`,
		},
		{
			name:        "TTLFail",
			extraParams: "cgrates.org;Stat_1;FLTR_STAT_1;2014-07-29T15:00:00Z;100;BadString;0;*acd&*tcd&*asr;Metric_FLTR;false;true;30;*none;key:value",
			expectedErr: `time: invalid duration "BadString"`,
		},
		{
			name:        "MinItemsFail",
			extraParams: "cgrates.org;Stat_1;FLTR_STAT_1;2014-07-29T15:00:00Z;100;10s;BadString;*acd&*tcd&*asr;Metric_FLTR;false;true;30;*none;key:value",
			expectedErr: `strconv.Atoi: parsing "BadString": invalid syntax`,
		},
		{
			name:        "StoredFail",
			extraParams: "cgrates.org;Stat_1;FLTR_STAT_1;2014-07-29T15:00:00Z;100;10s;0;*acd&*tcd&*asr;Metric_FLTR;BadString;true;30;*none;key:value",
			expectedErr: `strconv.ParseBool: parsing "BadString": invalid syntax`,
		},
		{
			name:        "BlockerFail",
			extraParams: "cgrates.org;Stat_1;FLTR_STAT_1;2014-07-29T15:00:00Z;100;10s;0;*acd&*tcd&*asr;Metric_FLTR;false;BadString;30;*none;key:value",
			expectedErr: `strconv.ParseBool: parsing "BadString": invalid syntax`,
		},
		{
			name:        "WeightFail",
			extraParams: "cgrates.org;Stat_1;FLTR_STAT_1;2014-07-29T15:00:00Z;100;10s;0;*acd&*tcd&*asr;Metric_FLTR;false;true;BadString;*none;key:value",
			expectedErr: `strconv.ParseFloat: parsing "BadString": invalid syntax`,
		},
		{
			name:        "InvalidOptsMap",
			extraParams: "cgrates.org;Stat_1;FLTR_STAT_1;2014-07-29T15:00:00Z;100;10s;0;*acd&*tcd&*asr;Metric_FLTR;false;true;30;*none;opt",
			expectedErr: "invalid key-value pair: opt",
		},
	}

	for i, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			action := &Action{ExtraParameters: tc.extraParams}
			ev := &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "evID",
				Time:   &time.Time{},
				Event: map[string]any{
					utils.AccountField: "1001",
				},
			}
			t.Cleanup(func() {
				sqpwo = nil
			})
			err := dynamicStats(nil, action, nil, nil, ev,
				SharedActionsData{}, ActionConnCfg{
					ConnIDs: tc.connIDs,
				})
			if tc.expectedErr != "" {
				if err == nil || err.Error() != tc.expectedErr {
					t.Errorf("expected error <%v>, received <%v>", tc.expectedErr, err)
				}
			} else if err != nil {
				t.Error(err)
			} else if !reflect.DeepEqual(sqpwo, tc.expSqpwo) {
				if i != 1 {
					t.Errorf("Expected <%v>\nReceived\n<%v>", utils.ToJSON(tc.expSqpwo), utils.ToJSON(sqpwo))
				} else {
					// Get the absolute difference between the times
					diff := sqpwo.ActivationInterval.ActivationTime.Sub(tc.expSqpwo.ActivationInterval.ActivationTime)
					if diff < 0 {
						diff = -diff // Make sure it's positive
					}
					// Check if difference is less than or equal to 1 second
					if diff <= time.Second {
						tc.expSqpwo.ActivationInterval.ActivationTime = sqpwo.ActivationInterval.ActivationTime
						if !reflect.DeepEqual(sqpwo, tc.expSqpwo) {
							t.Errorf("Expected <%v>\nReceived\n<%v>", utils.ToJSON(tc.expSqpwo), utils.ToJSON(sqpwo))
						}
					} else {
						t.Errorf("Expected <%v>\nReceived\n<%v>", utils.ToJSON(tc.expSqpwo), utils.ToJSON(sqpwo))
					}
				}
			}
		})
	}
}

func TestDynamicAttribute(t *testing.T) {
	tempConn := connMgr
	tmpDm := dm
	tmpCache := Cache
	defer func() {
		config.SetCgrConfig(config.NewDefaultCGRConfig())
		SetConnManager(tempConn)
		dm = tmpDm
		Cache = tmpCache
	}()
	Cache.Clear(nil)
	var apwo *AttributeProfileWithAPIOpts
	ccMock := &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.APIerSv1SetAttributeProfile: func(ctx *context.Context, args, reply any) error {
				var canCast bool
				if apwo, canCast = args.(*AttributeProfileWithAPIOpts); !canCast {
					return fmt.Errorf("couldnt cast")
				}
				return nil
			},
		},
	}
	connID := utils.ConcatenatedKey(utils.MetaInternal, utils.MetaApier)
	clientconn := make(chan birpc.ClientConnector, 1)
	clientconn <- ccMock
	NewConnManager(config.NewDefaultCGRConfig(), map[string]chan birpc.ClientConnector{
		connID: clientconn,
	})
	testcases := []struct {
		name        string
		extraParams string
		connIDs     []string
		expApwo     *AttributeProfileWithAPIOpts
		expectedErr string
	}{
		{
			name:    "SuccessfulRequest",
			connIDs: []string{connID},
			expApwo: &AttributeProfileWithAPIOpts{
				AttributeProfile: &AttributeProfile{
					Tenant:    "cgrates.org",
					ID:        "Attr_1",
					Contexts:  []string{"*sessions", "*chargers"},
					FilterIDs: []string{"FLTR_ATTR_1", "FLTR_ATTR_2"},
					ActivationInterval: &utils.ActivationInterval{
						ActivationTime: time.Date(2014, 7, 29, 15, 0, 0, 0, time.UTC),
					},
					Attributes: []*Attribute{
						{
							FilterIDs: []string{"AttrFltr_1", "AttrFltr2"},
							Path:      "*req.Subject",
							Type:      "*constant",
							Value:     config.NewRSRParsersMustCompile("SUPPLIER1&SUPPLIER2", "&"),
						},
					},
					Blocker: true,
					Weight:  10,
				},
				APIOpts: map[string]any{
					"key": "value",
				},
			},
			extraParams: "cgrates.org;Attr_1;*sessions&*chargers;FLTR_ATTR_1&FLTR_ATTR_2;2014-07-29T15:00:00Z;AttrFltr_1&AttrFltr2;*req.Subject;*constant;SUPPLIER1&SUPPLIER2;true;10;key:value",
		},
		{
			name:    "SuccessfulRequestWithDynamicPaths",
			connIDs: []string{connID},
			expApwo: &AttributeProfileWithAPIOpts{
				AttributeProfile: &AttributeProfile{
					Tenant:    "cgrates.org",
					ID:        "Attr_1001",
					Contexts:  []string{"*any"},
					FilterIDs: []string{"*string:~*req.Account:1001"},
					ActivationInterval: &utils.ActivationInterval{
						ActivationTime: time.Now(),
						ExpiryTime:     time.Date(3000, 7, 29, 15, 0, 0, 0, time.UTC),
					},
					Attributes: []*Attribute{
						{
							FilterIDs: []string{"AttrFltr_1", "AttrFltr2"},
							Path:      "*req.Subject",
							Type:      "*constant",
							Value:     config.NewRSRParsersMustCompile("SUPPLIER1", "&"),
						},
					},
					Blocker: true,
					Weight:  10,
				},
				APIOpts: map[string]any{
					"key": "value",
				},
			},
			extraParams: "*tenant;Attr_<~*req.Account>;*any;*string:~*req.Account:<~*req.Account>;*now&3000-07-29T15:00:00Z;AttrFltr_1&AttrFltr2;*req.Subject;*constant;SUPPLIER1;true;10;key:value",
		},
		{
			name:    "SuccessfulRequestEmptyFields",
			connIDs: []string{connID},
			expApwo: &AttributeProfileWithAPIOpts{
				AttributeProfile: &AttributeProfile{
					Tenant:             "cgrates.org",
					ID:                 "Attr_1",
					Contexts:           nil,
					FilterIDs:          nil,
					ActivationInterval: &utils.ActivationInterval{},
					Attributes:         nil,
					Blocker:            false,
					Weight:             0,
				},
				APIOpts: map[string]any{},
			},
			extraParams: "cgrates.org;Attr_1;;;;;;;;;;",
		},
		{
			name:        "MissingConns",
			extraParams: "cgrates.org;Attr_1;*sessions&*chargers;FLTR_ATTR_1&FLTR_ATTR_2;2014-07-29T15:00:00Z;AttrFltr_1&AttrFltr2;*req.Subject;*constant;SUPPLIER1;true;10;key:value",
			expectedErr: "MANDATORY_IE_MISSING: [connIDs]",
		},
		{
			name:        "WrongNumberOfParams",
			extraParams: "tenant;;1;",
			expectedErr: "invalid number of parameters <4> expected 12",
		},
		{
			name:        "ActivationIntervalLengthFail",
			extraParams: "cgrates.org;Attr_1;*sessions&*chargers;FLTR_ATTR_1&FLTR_ATTR_2;2014-07-29T15:00:00Z&&;AttrFltr_1&AttrFltr2;*req.Subject;*constant;SUPPLIER1;true;10;key:value",
			expectedErr: utils.ErrUnsupportedFormat.Error(),
		},
		{
			name:        "ActivationIntervalBadStringFail",
			extraParams: "cgrates.org;Attr_1;*sessions&*chargers;FLTR_ATTR_1&FLTR_ATTR_2;bad String;AttrFltr_1&AttrFltr2;*req.Subject;*constant;SUPPLIER1;true;10;key:value",
			expectedErr: `parsing time "bad String" as "2006-01-02T15:04:05Z07:00": cannot parse "bad String" as "2006"`,
		},
		{
			name:        "ActivationIntervalBadStringFail2",
			extraParams: "cgrates.org;Attr_1;*sessions&*chargers;FLTR_ATTR_1&FLTR_ATTR_2;2014-07-29T15:00:00Z&bad String;AttrFltr_1&AttrFltr2;*req.Subject;*constant;SUPPLIER1;true;10;key:value",
			expectedErr: `parsing time "bad String" as "2006-01-02T15:04:05Z07:00": cannot parse "bad String" as "2006"`,
		},
		{
			name:        "AttributesValueFail",
			extraParams: "cgrates.org;Attr_1;*sessions&*chargers;FLTR_ATTR_1&FLTR_ATTR_2;2014-07-29T15:00:00Z;AttrFltr_1&AttrFltr2;*req.Subject;*constant;&&&&;true;10;key:value",
			expectedErr: `emtpy RSRParser in rule: <>`,
		},
		{
			name:        "BlockerFail",
			extraParams: "cgrates.org;Attr_1;*sessions&*chargers;FLTR_ATTR_1&FLTR_ATTR_2;2014-07-29T15:00:00Z;AttrFltr_1&AttrFltr2;*req.Subject;*constant;SUPPLIER1;BadString;10;key:value",
			expectedErr: `strconv.ParseBool: parsing "BadString": invalid syntax`,
		},
		{
			name:        "WeightFail",
			extraParams: "cgrates.org;Attr_1;*sessions&*chargers;FLTR_ATTR_1&FLTR_ATTR_2;2014-07-29T15:00:00Z;AttrFltr_1&AttrFltr2;*req.Subject;*constant;SUPPLIER1;true;BadString;key:value",
			expectedErr: `strconv.ParseFloat: parsing "BadString": invalid syntax`,
		},
		{
			name:        "InvalidOptsMap",
			extraParams: "cgrates.org;Attr_1;*sessions&*chargers;FLTR_ATTR_1&FLTR_ATTR_2;2014-07-29T15:00:00Z;AttrFltr_1&AttrFltr2;*req.Subject;*constant;SUPPLIER1;true;10;opt",
			expectedErr: "invalid key-value pair: opt",
		},
	}

	for i, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			action := &Action{ExtraParameters: tc.extraParams}
			ev := &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "evID",
				Time:   &time.Time{},
				Event: map[string]any{
					utils.AccountField: "1001",
				},
			}
			t.Cleanup(func() {
				apwo = nil
			})
			err := dynamicAttribute(nil, action, nil, nil, ev,
				SharedActionsData{}, ActionConnCfg{
					ConnIDs: tc.connIDs,
				})
			if tc.expectedErr != "" {
				if err == nil || err.Error() != tc.expectedErr {
					t.Errorf("expected error <%v>, received <%v>", tc.expectedErr, err)
				}
			} else if err != nil {
				t.Error(err)
			} else if !reflect.DeepEqual(apwo, tc.expApwo) {
				if i != 1 {
					t.Errorf("Expected <%v>\nReceived\n<%v>", utils.ToJSON(tc.expApwo), utils.ToJSON(apwo))
				} else {
					// Get the absolute difference between the times
					diff := apwo.ActivationInterval.ActivationTime.Sub(tc.expApwo.ActivationInterval.ActivationTime)
					if diff < 0 {
						diff = -diff // Make sure it's positive
					}
					// Check if difference is less than or equal to 1 second
					if diff <= time.Second {
						tc.expApwo.ActivationInterval.ActivationTime = apwo.ActivationInterval.ActivationTime
						if !reflect.DeepEqual(apwo, tc.expApwo) {
							t.Errorf("Expected <%v>\nReceived\n<%v>", utils.ToJSON(tc.expApwo), utils.ToJSON(apwo))
						}
					} else {
						t.Errorf("Expected <%v>\nReceived\n<%v>", utils.ToJSON(tc.expApwo), utils.ToJSON(apwo))
					}
				}
			}
		})
	}
}

func TestDynamicActionPlan(t *testing.T) {
	tempConn := connMgr
	tmpDm := dm
	tmpCache := Cache
	defer func() {
		config.SetCgrConfig(config.NewDefaultCGRConfig())
		SetConnManager(tempConn)
		dm = tmpDm
		Cache = tmpCache
	}()
	Cache.Clear(nil)
	var ap *AttrSetActionPlan
	ccMock := &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.APIerSv1SetActionPlan: func(ctx *context.Context, args, reply any) error {
				var canCast bool
				if ap, canCast = args.(*AttrSetActionPlan); !canCast {
					return fmt.Errorf("couldnt cast AttrSetActionPlan")
				}
				return nil
			},
			utils.APIerSv1GetActions: func(ctx *context.Context, args2, reply any) error {
				return nil
			},
			utils.APIerSv1GetTiming: func(ctx *context.Context, args3, reply any) error {
				var canCast bool
				if args3, canCast = args3.(*utils.ArgsGetTimingID); !canCast {
					return fmt.Errorf("couldnt cast ArgsGetTimingID")
				}
				var exp *utils.TPTiming
				if utils.ToJSON(args3) == utils.ToJSON(&utils.ArgsGetTimingID{ID: "Timing_1"}) {
					exp = &utils.TPTiming{
						ID:        "Timing_1",
						Years:     utils.Years{2025},
						Months:    utils.Months{2},
						MonthDays: utils.MonthDays{3},
						WeekDays:  utils.WeekDays{1, 2, 3},
						StartTime: "00:12:12",
						EndTime:   "12:12:12",
					}
				}
				if utils.ToJSON(args3) == utils.ToJSON(&utils.ArgsGetTimingID{ID: "*asap"}) {
					exp = &utils.TPTiming{
						ID:        "*asap",
						Years:     utils.Years{2025},
						Months:    utils.Months{},
						MonthDays: utils.MonthDays{},
						WeekDays:  utils.WeekDays{},
						StartTime: "*asap",
					}
				}
				if exp == nil {
					return utils.ErrNotFound
				}
				*reply.(*utils.TPTiming) = *exp
				return nil
			},
		},
	}
	connID := utils.ConcatenatedKey(utils.MetaInternal, utils.MetaApier)
	clientconn := make(chan birpc.ClientConnector, 1)
	clientconn <- ccMock
	NewConnManager(config.NewDefaultCGRConfig(), map[string]chan birpc.ClientConnector{
		connID: clientconn,
	})
	testcases := []struct {
		name        string
		extraParams string
		connIDs     []string
		expAp       *AttrSetActionPlan
		expectedErr string
	}{
		{
			name:    "SuccessfulRequest",
			connIDs: []string{connID},
			expAp: &AttrSetActionPlan{
				Id: "ActPl_1",
				ActionPlan: []*AttrActionPlan{
					{
						ActionsId: "Action_1",
						TimingID:  "Timing_1",
						Years:     "2025",
						Months:    "2",
						MonthDays: "3",
						WeekDays:  "1;2;3",
						Time:      "00:12:12;12:12:12",
						Weight:    10,
					},
				},
				Overwrite:       true,
				ReloadScheduler: true,
			},
			extraParams: "ActPl_1;Action_1;Timing_1;10;true",
		},
		{
			name:    "SuccessfulRequestWithDynamicPaths",
			connIDs: []string{connID},
			expAp: &AttrSetActionPlan{
				Id: "ActPl_1",
				ActionPlan: []*AttrActionPlan{
					{
						ActionsId: "Action_1001",
						TimingID:  "*asap",
						Years:     "2025",
						Months:    "*any",
						MonthDays: "*any",
						WeekDays:  "*any",
						Time:      "*asap",
						Weight:    10,
					},
				},
				Overwrite:       false,
				ReloadScheduler: true,
			},
			extraParams: "ActPl_1;Action_<~*req.Account>;*asap;10;",
		},
		{
			name:    "SuccessfulRequestEmptyFields",
			connIDs: []string{connID},
			expAp: &AttrSetActionPlan{
				Id: "ActPl_1",
				ActionPlan: []*AttrActionPlan{
					{
						ActionsId: "Action_1",
						TimingID:  "",
						Weight:    0,
					},
				},
				Overwrite:       false,
				ReloadScheduler: true,
			},
			extraParams: "ActPl_1;Action_1;;;",
		},
		{
			name:        "MissingConns",
			extraParams: "ActPl_1;Action_1;*asap;10;",
			expectedErr: "MANDATORY_IE_MISSING: [connIDs]",
		},
		{
			name:        "WrongNumberOfParams",
			extraParams: "ActPl_1;Action_1;*asap;10;;;",
			expectedErr: "invalid number of parameters <7> expected 5",
		},
		{
			name:        "ActionIdEmptyFail",
			extraParams: "ActPl_1;;*asap;10;",
			expectedErr: `empty ActionsId for <ActPl_1> dynamic_action_plan`,
		},
		{
			name:        "WeightFail",
			connIDs:     []string{connID},
			extraParams: "ActPl_1;Action_1;*asap;BadString;",
			expectedErr: `strconv.ParseFloat: parsing "BadString": invalid syntax`,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			action := &Action{ExtraParameters: tc.extraParams}
			ev := &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "evID",
				Time:   &time.Time{},
				Event: map[string]any{
					utils.AccountField: "1001",
				},
			}
			t.Cleanup(func() {
				ap = nil
			})
			err := dynamicActionPlan(nil, action, nil, nil, ev,
				SharedActionsData{}, ActionConnCfg{
					ConnIDs: tc.connIDs,
				})
			if tc.expectedErr != "" {
				if err == nil || err.Error() != tc.expectedErr {
					t.Errorf("expected error <%v>, received <%v>", tc.expectedErr, err)
				}
			} else if err != nil {
				t.Error(err)
			} else if utils.ToJSON(ap) != utils.ToJSON(tc.expAp) {
				t.Errorf("Expected <%v>\nReceived\n<%v>", utils.ToJSON(tc.expAp), utils.ToJSON(ap))
			}
		})
	}
}

func TestDynamicActionPlanAccount(t *testing.T) {
	tempConn := connMgr
	tmpDm := dm
	tmpCache := Cache
	defer func() {
		config.SetCgrConfig(config.NewDefaultCGRConfig())
		SetConnManager(tempConn)
		dm = tmpDm
		Cache = tmpCache
	}()
	Cache.Clear(nil)
	var ap *AttrSetActionPlanAccounts
	ccMock := &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.APIerSv1SetActionPlanAccounts: func(ctx *context.Context, args, reply any) error {
				var canCast bool
				if ap, canCast = args.(*AttrSetActionPlanAccounts); !canCast {
					return fmt.Errorf("couldnt cast AttrSetActionPlanAccounts")
				}
				return nil
			},
			utils.APIerSv1GetActions: func(ctx *context.Context, args2, reply any) error {
				return nil
			},
			utils.APIerSv1GetTiming: func(ctx *context.Context, args3, reply any) error {
				var canCast bool
				if args3, canCast = args3.(*utils.ArgsGetTimingID); !canCast {
					return fmt.Errorf("couldnt cast ArgsGetTimingID")
				}
				var exp *utils.TPTiming
				if utils.ToJSON(args3) == utils.ToJSON(&utils.ArgsGetTimingID{ID: "Timing_1"}) {
					exp = &utils.TPTiming{
						ID:        "Timing_1",
						Years:     utils.Years{2025},
						Months:    utils.Months{2},
						MonthDays: utils.MonthDays{3},
						WeekDays:  utils.WeekDays{1, 2, 3},
						StartTime: "00:12:12",
						EndTime:   "12:12:12",
					}
				}
				if utils.ToJSON(args3) == utils.ToJSON(&utils.ArgsGetTimingID{ID: "*asap"}) {
					exp = &utils.TPTiming{
						ID:        "*asap",
						Years:     utils.Years{2025},
						Months:    utils.Months{},
						MonthDays: utils.MonthDays{},
						WeekDays:  utils.WeekDays{},
						StartTime: "*asap",
					}
				}
				if exp == nil {
					return utils.ErrNotFound
				}
				*reply.(*utils.TPTiming) = *exp
				return nil
			},
		},
	}
	connID := utils.ConcatenatedKey(utils.MetaInternal, utils.MetaApier)
	clientconn := make(chan birpc.ClientConnector, 1)
	clientconn <- ccMock
	NewConnManager(config.NewDefaultCGRConfig(), map[string]chan birpc.ClientConnector{
		connID: clientconn,
	})
	testcases := []struct {
		name        string
		extraParams string
		connIDs     []string
		expAp       *AttrSetActionPlanAccounts
		expectedErr string
	}{
		{
			name:    "SuccessfulRequest",
			connIDs: []string{connID},
			expAp: &AttrSetActionPlanAccounts{
				Id: "ActPl_1",
				ActionPlan: []*AttrActionPlan{
					{
						ActionsId: "Action_1",
						TimingID:  "Timing_1",
						Years:     "2025",
						Months:    "2",
						MonthDays: "3",
						WeekDays:  "1;2;3",
						Time:      "00:12:12;12:12:12",
						Weight:    10,
					},
				},
				AccountIDs:      []string{"cgrates.org:1001"},
				Overwrite:       true,
				ReloadScheduler: true,
			},
			extraParams: "ActPl_1;Action_1;Timing_1;10;true;cgrates.org:1001",
		},
		{
			name:    "SuccessfulRequestWithDynamicPaths",
			connIDs: []string{connID},
			expAp: &AttrSetActionPlanAccounts{
				Id: "ActPl_1",
				ActionPlan: []*AttrActionPlan{
					{
						ActionsId: "Action_1001",
						TimingID:  "*asap",
						Years:     "2025",
						Months:    "*any",
						MonthDays: "*any",
						WeekDays:  "*any",
						Time:      "*asap",
						Weight:    10,
					},
				},
				AccountIDs:      []string{"cgrates.org:1001"},
				Overwrite:       false,
				ReloadScheduler: true,
			},
			extraParams: "ActPl_1;Action_<~*req.Account>;*asap;10;;<*tenant+:+~*req.Account>",
		},
		{
			name:    "SuccessfulRequestEmptyFields",
			connIDs: []string{connID},
			expAp: &AttrSetActionPlanAccounts{
				Id: "ActPl_1",
				ActionPlan: []*AttrActionPlan{
					{
						ActionsId: "Action_1",
						TimingID:  "",
						Weight:    0,
					},
				},
				Overwrite:       false,
				ReloadScheduler: true,
			},
			extraParams: "ActPl_1;Action_1;;;;",
		},
		{
			name:        "MissingConns",
			extraParams: "ActPl_1;Action_1;*asap;10;;",
			expectedErr: "MANDATORY_IE_MISSING: [connIDs]",
		},
		{
			name:        "WrongNumberOfParams",
			extraParams: "ActPl_1;Action_1;*asap;10;;;",
			expectedErr: "invalid number of parameters <7> expected 6",
		},
		{
			name:        "ActionIdEmptyFail",
			extraParams: "ActPl_1;;*asap;10;;",
			expectedErr: `empty ActionsId for <ActPl_1> dynamic_action_plan`,
		},
		{
			name:        "WeightFail",
			connIDs:     []string{connID},
			extraParams: "ActPl_1;Action_1;*asap;BadString;;",
			expectedErr: `strconv.ParseFloat: parsing "BadString": invalid syntax`,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			action := &Action{ExtraParameters: tc.extraParams}
			ev := &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "evID",
				Time:   &time.Time{},
				Event: map[string]any{
					utils.AccountField: "1001",
				},
			}
			t.Cleanup(func() {
				ap = nil
			})
			err := dynamicActionPlanAccount(nil, action, nil, nil, ev,
				SharedActionsData{}, ActionConnCfg{
					ConnIDs: tc.connIDs,
				})
			if tc.expectedErr != "" {
				if err == nil || err.Error() != tc.expectedErr {
					t.Errorf("expected error <%v>, received <%v>", tc.expectedErr, err)
				}
			} else if err != nil {
				t.Error(err)
			} else if utils.ToJSON(ap) != utils.ToJSON(tc.expAp) {
				t.Errorf("Expected <%v>\nReceived\n<%v>", utils.ToJSON(tc.expAp), utils.ToJSON(ap))
			}
		})
	}
}

func TestDynamicActionAll(t *testing.T) {
	tempConn := connMgr
	tmpDm := dm
	tmpCache := Cache
	defer func() {
		config.SetCgrConfig(config.NewDefaultCGRConfig())
		SetConnManager(tempConn)
		dm = tmpDm
		Cache = tmpCache
	}()
	Cache.Clear(nil)
	var a *utils.AttrSetActions
	ccMock := &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.APIerSv2SetActions: func(ctx *context.Context, args, reply any) error {
				var canCast bool
				if a, canCast = args.(*utils.AttrSetActions); !canCast {
					return fmt.Errorf("couldnt cast")
				}
				return nil
			},
		},
	}
	connID := utils.ConcatenatedKey(utils.MetaInternal, utils.MetaApier)
	clientconn := make(chan birpc.ClientConnector, 1)
	clientconn <- ccMock
	NewConnManager(config.NewDefaultCGRConfig(), map[string]chan birpc.ClientConnector{
		connID: clientconn,
	})
	testcases := []struct {
		name        string
		extraParams string
		connIDs     []string
		expA        *utils.AttrSetActions
		expectedErr string
	}{
		{
			name:    "SuccessfulRequest",
			connIDs: []string{connID},
			expA: &utils.AttrSetActions{
				ActionsId: "Alter_Session_10",
				Actions: []*utils.TPAction{
					{
						Identifier:      utils.MetaAlterSessions,
						ExtraParameters: "cgrates.org;*string:~*req.Account:1001;1;*radCoATemplate:mycoa;CustomFilter:mycustomvalue",
						Filters:         "*string:~*req.Account:1001;filter2",
						BalanceId:       "badID",
						BalanceType:     utils.MetaMonetary,
						Categories:      "call;data",
						DestinationIds:  "1002;1003",
						RatingSubject:   "SPECIAL_1002",
						SharedGroups:    "SHARED_A;SHARED_B",
						ExpiryTime:      "*unlimited",
						TimingTags:      "weekdays;offpeak",
						Units:           "10",
						BalanceWeight:   "10",
						BalanceBlocker:  "true",
						BalanceDisabled: "true",
						Weight:          10,
					},
				},
				Overwrite: false,
			},
			extraParams: "Alter_Session_10;*alter_sessions;\fcgrates.org;*string:~*req.Account:1001;1;*radCoATemplate:mycoa;CustomFilter:mycustomvalue\f;*string:~*req.Account:1001&filter2;badID;*monetary;call&data;1002&1003;SPECIAL_1002;SHARED_A&SHARED_B;*unlimited;weekdays&offpeak;10;10;true;true;10",
		},
		{
			name:    "SuccessfulRequestWithDynamicPaths",
			connIDs: []string{connID},
			expA: &utils.AttrSetActions{
				ActionsId: "CDR_Log_1001",
				Actions: []*utils.TPAction{
					{
						Identifier:      utils.CDRLog,
						ExtraParameters: "{\"Account\":\"1001\",\"RequestType\":\"*pseudoprepaid\",\"Subject\":\"DifferentThanAccount\", \"ToR\":\"~ActionType:s/^\\*(.*)$/did_$1/\"}",
						Filters:         "*string:~*req.Account:1001;filter2",
						BalanceId:       "badID",
						BalanceType:     utils.MetaMonetary,
						Categories:      "call;data",
						DestinationIds:  "1002;1003",
						RatingSubject:   "SPECIAL_1002",
						SharedGroups:    "SHARED_A;SHARED_B",
						ExpiryTime:      "*unlimited",
						TimingTags:      "weekdays;offpeak",
						Units:           "10",
						BalanceWeight:   "10",
						BalanceBlocker:  "true",
						BalanceDisabled: "true",
						Weight:          10,
					},
				},
				Overwrite: false,
			},
			extraParams: "CDR_Log_<~*req.Account>;*cdrlog;\f{\"Account\":\"<~*req.Account>\",\"RequestType\":\"*pseudoprepaid\",\"Subject\":\"DifferentThanAccount\", \"ToR\":\"~ActionType:s/^\\*(.*)$/did_$1/\"}\f;*string:~*req.Account:<~*req.Account>&filter2;badID;*monetary;call&data;1002&1003;SPECIAL_1002;SHARED_A&SHARED_B;*unlimited;weekdays&offpeak;10;10;true;true;10",
		},
		{
			name:    "SuccessfulRequestEmptyFields",
			connIDs: []string{connID},
			expA: &utils.AttrSetActions{
				ActionsId: "DISABLE_ACC",
				Actions: []*utils.TPAction{
					{
						Identifier: utils.MetaDisableAccount,
					},
				},
			},
			extraParams: "DISABLE_ACC;*disable_account;;;;;;;;;;;;;;;",
		},
		{
			name:        "MissingConns",
			extraParams: "TOPUP_MONETARY_10;*topup;;*string:~*req.Account:1001&filter2;badID;*monetary;call&data;1002&1003;SPECIAL_1002;SHARED_A&SHARED_B;*unlimited;weekdays&offpeak;10;10;true;true;10",
			expectedErr: "MANDATORY_IE_MISSING: [connIDs]",
		},
		{
			name:        "WrongNumberOfParams",
			extraParams: "TOPUP_MONETARY_10;*topup;;*string:~*req.Account:1001&filter2;badID;*monetary;call&data;1002&1003;SPECIAL_1002;SHARED_A&SHARED_B;*unlimited;weekdays&offpeak;10;10;true;true;10;;;",
			expectedErr: "invalid number of parameters <20> expected 17",
		},
		{
			name:        "ActionIdEmptyFail",
			extraParams: "DISABLE_ACC;;;;;;;;;;;;;;;;",
			expectedErr: `empty Action for <DISABLE_ACC> dynamic_action`,
		},
		{
			name:        "ActionIdEmptyFail",
			extraParams: ";;;;;;;;;;;;;;;;",
			expectedErr: `empty ActionsId for dynamic_action`,
		},
		{
			name:        "WeightFail",
			extraParams: "TOPUP_MONETARY_10;*topup;;*string:~*req.Account:1001&filter2;badID;*monetary;call&data;1002&1003;SPECIAL_1002;SHARED_A&SHARED_B;*unlimited;weekdays&offpeak;10;10;true;true;BadString",
			expectedErr: `strconv.ParseFloat: parsing "BadString": invalid syntax`,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			action := &Action{ExtraParameters: tc.extraParams}
			ev := &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "evID",
				Time:   &time.Time{},
				Event: map[string]any{
					utils.AccountField: "1001",
				},
			}
			t.Cleanup(func() {
				a = nil
			})
			err := dynamicAction(nil, action, nil, nil, ev,
				SharedActionsData{}, ActionConnCfg{
					ConnIDs: tc.connIDs,
				})
			if tc.expectedErr != "" {
				if err == nil || err.Error() != tc.expectedErr {
					t.Errorf("expected error <%v>, received <%v>", tc.expectedErr, err)
				}
			} else if err != nil {
				t.Error(err)
			} else if utils.ToJSON(a) != utils.ToJSON(tc.expA) {
				t.Errorf("Expected <%v>\nReceived\n<%v>", utils.ToJSON(tc.expA), utils.ToJSON(a))
			}
		})
	}
}

func TestDynamicDestination(t *testing.T) {
	tempConn := connMgr
	tmpDm := dm
	tmpCache := Cache
	defer func() {
		config.SetCgrConfig(config.NewDefaultCGRConfig())
		SetConnManager(tempConn)
		dm = tmpDm
		Cache = tmpCache
	}()
	Cache.Clear(nil)
	var dest *utils.AttrSetDestination
	ccMock := &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.APIerSv1SetDestination: func(ctx *context.Context, args, reply any) error {
				var canCast bool
				if dest, canCast = args.(*utils.AttrSetDestination); !canCast {
					return fmt.Errorf("couldnt cast")
				}
				return nil
			},
		},
	}
	connID := utils.ConcatenatedKey(utils.MetaInternal, utils.MetaApier)
	clientconn := make(chan birpc.ClientConnector, 1)
	clientconn <- ccMock
	NewConnManager(config.NewDefaultCGRConfig(), map[string]chan birpc.ClientConnector{
		connID: clientconn,
	})
	testcases := []struct {
		name        string
		extraParams string
		connIDs     []string
		expDest     *utils.AttrSetDestination
		expectedErr string
	}{
		{
			name:    "SuccessfulRequest",
			connIDs: []string{connID},
			expDest: &utils.AttrSetDestination{
				Id:        "DST_1002&1003",
				Prefixes:  []string{"1002", "1003"},
				Overwrite: false,
			},
			extraParams: "DST_1002&1003;1002&1003",
		},
		{
			name:    "SuccessfulRequestWithDynamicPaths",
			connIDs: []string{connID},
			expDest: &utils.AttrSetDestination{
				Id:        "DST_1001&1003",
				Prefixes:  []string{"1001", "1003"},
				Overwrite: false,
			},
			extraParams: "DST_1001&1003;<~*req.Account>&1003",
		},
		{
			name:    "SuccessfulRequestEmptyFields",
			connIDs: []string{connID},
			expDest: &utils.AttrSetDestination{
				Id:        "DST_Any",
				Overwrite: false,
			},
			extraParams: "DST_Any;",
		},
		{
			name:        "MissingConns",
			extraParams: "DST_1002&1003;1002&1003",
			expectedErr: "MANDATORY_IE_MISSING: [connIDs]",
		},
		{
			name:        "WrongNumberOfParams",
			extraParams: "DST_1002&1003;1002&1003;;;",
			expectedErr: "invalid number of parameters <5> expected 2",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			action := &Action{ExtraParameters: tc.extraParams}
			ev := &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "evID",
				Time:   &time.Time{},
				Event: map[string]any{
					utils.AccountField: "1001",
				},
			}
			t.Cleanup(func() {
				dest = nil
			})
			err := dynamicDestination(nil, action, nil, nil, ev,
				SharedActionsData{}, ActionConnCfg{
					ConnIDs: tc.connIDs,
				})
			if tc.expectedErr != "" {
				if err == nil || err.Error() != tc.expectedErr {
					t.Errorf("expected error <%v>, received <%v>", tc.expectedErr, err)
				}
			} else if err != nil {
				t.Error(err)
			} else if utils.ToJSON(dest) != utils.ToJSON(tc.expDest) {
				t.Errorf("Expected <%v>\nReceived\n<%v>", utils.ToJSON(tc.expDest), utils.ToJSON(dest))
			}
		})
	}
}

func TestDynamicFilter(t *testing.T) {
	tempConn := connMgr
	tmpDm := dm
	tmpCache := Cache
	defer func() {
		config.SetCgrConfig(config.NewDefaultCGRConfig())
		SetConnManager(tempConn)
		dm = tmpDm
		Cache = tmpCache
	}()
	Cache.Clear(nil)
	var fwo *FilterWithAPIOpts
	ccMock := &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.APIerSv1SetFilter: func(ctx *context.Context, args, reply any) error {
				var canCast bool
				if fwo, canCast = args.(*FilterWithAPIOpts); !canCast {
					return fmt.Errorf("couldnt cast")
				}
				return nil
			},
		},
	}
	connID := utils.ConcatenatedKey(utils.MetaInternal, utils.MetaApier)
	clientconn := make(chan birpc.ClientConnector, 1)
	clientconn <- ccMock
	NewConnManager(config.NewDefaultCGRConfig(), map[string]chan birpc.ClientConnector{
		connID: clientconn,
	})
	testcases := []struct {
		name        string
		extraParams string
		connIDs     []string
		expFwo      *FilterWithAPIOpts
		expectedErr string
	}{
		{
			name:    "SuccessfulRequest",
			connIDs: []string{connID},
			expFwo: &FilterWithAPIOpts{
				Filter: &Filter{
					Tenant: "cgrates.org",
					ID:     "Fltr_1",
					Rules: []*FilterRule{
						{
							Type:    "*string",
							Element: "~*req.Account",
							Values:  []string{"1001", "1002"},
						},
					},
					ActivationInterval: &utils.ActivationInterval{
						ActivationTime: time.Date(2014, 7, 29, 15, 0, 0, 0, time.UTC),
					},
				},
				APIOpts: map[string]any{
					"key": "value",
				},
			},
			extraParams: "cgrates.org;Fltr_1;*string;~*req.Account;1001&1002;2014-07-29T15:00:00Z;key:value",
		},
		{
			name:    "SuccessfulRequestWithDynamicPaths",
			connIDs: []string{connID},
			expFwo: &FilterWithAPIOpts{
				Filter: &Filter{
					Tenant: "cgrates.org",
					ID:     "Fltr_1001",
					Rules: []*FilterRule{
						{
							Type:    "*string",
							Element: "~*req.Account",
							Values:  []string{"1001", "1002"},
						},
					},
					ActivationInterval: &utils.ActivationInterval{
						ActivationTime: time.Now(),
						ExpiryTime:     time.Date(3000, 7, 29, 15, 0, 0, 0, time.UTC),
					},
				},
				APIOpts: map[string]any{
					"key": "value",
				},
			},
			extraParams: "*tenant;Fltr_<~*req.Account>;*string;~*req.<~*req.ExtraInfo>;<~*req.Account>&1002;*now&3000-07-29T15:00:00Z;key:value",
		},
		{
			name:    "SuccessfulRequestEmptyFields",
			connIDs: []string{connID},
			expFwo: &FilterWithAPIOpts{
				Filter: &Filter{
					Tenant: "cgrates.org",
					ID:     "Fltr_1001",
					Rules: []*FilterRule{
						{
							Type:    "",
							Element: "",
							Values:  nil,
						},
					},
					ActivationInterval: &utils.ActivationInterval{},
				},
				APIOpts: map[string]any{},
			},
			extraParams: "*tenant;Fltr_1001;;;;;",
		},
		{
			name:        "MissingConns",
			extraParams: "cgrates.org;Fltr_1;*string;~*req.Account;1001&1002;2014-07-29T15:00:00Z;key:value",
			expectedErr: "MANDATORY_IE_MISSING: [connIDs]",
		},
		{
			name:        "WrongNumberOfParams",
			extraParams: "tenant;;;",
			expectedErr: "invalid number of parameters <4> expected 7",
		},
		{
			name:        "ActivationIntervalLengthFail",
			extraParams: "cgrates.org;Fltr_1;*string;~*req.Account;1001&1002;2014-07-29T15:00:00Z&&;key:value",
			expectedErr: utils.ErrUnsupportedFormat.Error(),
		},
		{
			name:        "ActivationIntervalBadStringFail",
			extraParams: "cgrates.org;Fltr_1;*string;~*req.Account;1001&1002;bad String;key:value",
			expectedErr: `parsing time "bad String" as "2006-01-02T15:04:05Z07:00": cannot parse "bad String" as "2006"`,
		},
		{
			name:        "ActivationIntervalBadStringFail2",
			extraParams: "cgrates.org;Fltr_1;*string;~*req.Account;1001&1002;2014-07-29T15:00:00Z&bad String;key:value",
			expectedErr: `parsing time "bad String" as "2006-01-02T15:04:05Z07:00": cannot parse "bad String" as "2006"`,
		},
		{
			name:        "InvalidOptsMap",
			extraParams: "cgrates.org;Fltr_1;*string;~*req.Account;1001&1002;2014-07-29T15:00:00Z;opt",
			expectedErr: "invalid key-value pair: opt",
		},
	}

	for i, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			action := &Action{ExtraParameters: tc.extraParams}
			ev := &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "evID",
				Time:   &time.Time{},
				Event: map[string]any{
					utils.AccountField: "1001",
					utils.ExtraInfo:    utils.AccountField,
				},
			}
			t.Cleanup(func() {
				fwo = nil
			})
			err := dynamicFilter(nil, action, nil, nil, ev,
				SharedActionsData{}, ActionConnCfg{
					ConnIDs: tc.connIDs,
				})
			if tc.expectedErr != "" {
				if err == nil || err.Error() != tc.expectedErr {
					t.Errorf("expected error <%v>, received <%v>", tc.expectedErr, err)
				}
			} else if err != nil {
				t.Error(err)
			} else if !reflect.DeepEqual(fwo, tc.expFwo) {
				if i != 1 {
					t.Errorf("Expected <%v>\nReceived\n<%v>", utils.ToJSON(tc.expFwo), utils.ToJSON(fwo))
				} else {
					// Get the absolute difference between the times
					diff := fwo.ActivationInterval.ActivationTime.Sub(tc.expFwo.ActivationInterval.ActivationTime)
					if diff < 0 {
						diff = -diff // Make sure it's positive
					}
					// Check if difference is less than or equal to 1 second
					if diff <= time.Second {
						tc.expFwo.ActivationInterval.ActivationTime = fwo.ActivationInterval.ActivationTime
						if !reflect.DeepEqual(fwo, tc.expFwo) {
							t.Errorf("Expected <%v>\nReceived\n<%v>", utils.ToJSON(tc.expFwo), utils.ToJSON(fwo))
						}
					} else {
						t.Errorf("Expected <%v>\nReceived\n<%v>", utils.ToJSON(tc.expFwo), utils.ToJSON(fwo))
					}
				}
			}
		})
	}
}

func TestDynamicRoute(t *testing.T) {
	tempConn := connMgr
	tmpDm := dm
	tmpCache := Cache
	defer func() {
		config.SetCgrConfig(config.NewDefaultCGRConfig())
		SetConnManager(tempConn)
		dm = tmpDm
		Cache = tmpCache
	}()
	Cache.Clear(nil)
	var rwo *RouteWithAPIOpts
	ccMock := &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.APIerSv1SetRouteProfile: func(ctx *context.Context, args, reply any) error {
				var canCast bool
				if rwo, canCast = args.(*RouteWithAPIOpts); !canCast {
					return fmt.Errorf("couldnt cast")
				}
				return nil
			},
		},
	}
	connID := utils.ConcatenatedKey(utils.MetaInternal, utils.MetaApier)
	clientconn := make(chan birpc.ClientConnector, 1)
	clientconn <- ccMock
	NewConnManager(config.NewDefaultCGRConfig(), map[string]chan birpc.ClientConnector{
		connID: clientconn,
	})
	testcases := []struct {
		name        string
		extraParams string
		connIDs     []string
		expRwo      *RouteWithAPIOpts
		expectedErr string
	}{
		{
			name:    "SuccessfulRequest",
			connIDs: []string{connID},
			expRwo: &RouteWithAPIOpts{
				RouteProfile: &RouteProfile{
					Tenant:    "cgrates.org",
					ID:        "RTP_ACNT_1001",
					FilterIDs: []string{"*string:~*req.Account:1001", "*string:~*req.Account:1002"},
					ActivationInterval: &utils.ActivationInterval{
						ActivationTime: time.Date(2014, 7, 29, 15, 0, 0, 0, time.UTC),
					},
					SortingParameters: []string{"*acd", "*tcc"},
					Routes: []*Route{
						{
							ID:              "route1",
							FilterIDs:       []string{"*string:~*req.Account:1001", "*string:~*req.Account:1002"},
							AccountIDs:      []string{"1001", "1002"},
							RatingPlanIDs:   []string{"RP1", "RP2"},
							ResourceIDs:     []string{"RS1", "RS2"},
							StatIDs:         []string{"Stat_1", "Stat_1_1"},
							Weight:          10,
							Blocker:         true,
							RouteParameters: "param",
						},
					},
					Sorting: utils.MetaWeight,
					Weight:  10,
				},
				APIOpts: map[string]any{
					"key": "value",
				},
			},
			extraParams: "cgrates.org;RTP_ACNT_1001;*string:~*req.Account:1001&*string:~*req.Account:1002;2014-07-29T15:00:00Z;*weight;*acd&*tcc;route1;*string:~*req.Account:1001&*string:~*req.Account:1002;1001&1002;RP1&RP2;RS1&RS2;Stat_1&Stat_1_1;10;true;param;10;key:value",
		},
		{
			name:    "SuccessfulRequestWithDynamicPaths",
			connIDs: []string{connID},
			expRwo: &RouteWithAPIOpts{
				RouteProfile: &RouteProfile{
					Tenant:    "cgrates.org",
					ID:        "RTP_ACNT_1001",
					FilterIDs: []string{"*string:~*req.Account:1001", "*string:~*req.Account:1002"},
					ActivationInterval: &utils.ActivationInterval{
						ActivationTime: time.Now(),
						ExpiryTime:     time.Date(3000, 7, 29, 15, 0, 0, 0, time.UTC),
					},
					SortingParameters: []string{"*acd", "*tcc"},
					Routes: []*Route{
						{
							ID:              "route1",
							FilterIDs:       []string{"*string:~*req.Account:1001", "*string:~*req.Account:1002"},
							AccountIDs:      []string{"1001", "1002"},
							RatingPlanIDs:   []string{"RP1", "RP2"},
							ResourceIDs:     []string{"RS1", "RS2"},
							StatIDs:         []string{"Stat_1", "Stat_1_1"},
							Weight:          10,
							Blocker:         true,
							RouteParameters: "param",
						},
					},
					Sorting: utils.MetaWeight,
					Weight:  10,
				},
				APIOpts: map[string]any{
					"key": "value",
				},
			},
			extraParams: "*tenant;RTP_ACNT_<~*req.Account>;*string:~*req.Account:<~*req.Account>&*string:~*req.Account:1002;*now&3000-07-29T15:00:00Z;*weight;*acd&*tcc;route1;*string:~*req.Account:<~*req.Account>&*string:~*req.Account:1002;<~*req.Account>&1002;RP1&RP2;RS1&RS2;Stat_1&Stat_1_1;10;true;param;10;key:value",
		},
		{
			name:    "SuccessfulRequestEmptyFields",
			connIDs: []string{connID},
			expRwo: &RouteWithAPIOpts{
				RouteProfile: &RouteProfile{
					Tenant:             "cgrates.org",
					ID:                 "RTP_ACNT_1001",
					FilterIDs:          nil,
					ActivationInterval: &utils.ActivationInterval{},
					SortingParameters:  nil,
					Routes: []*Route{
						{
							ID:              "route1",
							FilterIDs:       nil,
							AccountIDs:      nil,
							RatingPlanIDs:   nil,
							ResourceIDs:     nil,
							StatIDs:         nil,
							Weight:          0,
							Blocker:         false,
							RouteParameters: "",
						},
					},
					Sorting: "",
					Weight:  0,
				},
				APIOpts: map[string]any{},
			},
			extraParams: "cgrates.org;RTP_ACNT_1001;;;;;route1;;;;;;;;;;",
		},
		{
			name:        "MissingConns",
			extraParams: "cgrates.org;RTP_ACNT_1001;*string:~*req.Account:1001&*string:~*req.Account:1002;2014-07-29T15:00:00Z;*weight;*acd&*tcc;route1;*string:~*req.Account:1001&*string:~*req.Account:1002;1001&1002;RP1&RP2;RS1&RS2;Stat_1&Stat_1_1;10;true;param;10;key:value",
			expectedErr: "MANDATORY_IE_MISSING: [connIDs]",
		},
		{
			name:        "WrongNumberOfParams",
			extraParams: "tenant;RTP1;;",
			expectedErr: "invalid number of parameters <4> expected 17",
		},
		{
			name:        "ActivationIntervalLengthFail",
			extraParams: "cgrates.org;RTP_ACNT_1001;*string:~*req.Account:1001&*string:~*req.Account:1002;2014-07-29T15:00:00Z&&;*weight;*acd&*tcc;route1;*string:~*req.Account:1001&*string:~*req.Account:1002;1001&1002;RP1&RP2;RS1&RS2;Stat_1&Stat_1_1;10;true;param;10;key:value",
			expectedErr: utils.ErrUnsupportedFormat.Error(),
		},
		{
			name:        "ActivationIntervalBadStringFail",
			extraParams: "cgrates.org;RTP_ACNT_1001;*string:~*req.Account:1001&*string:~*req.Account:1002;bad String;*weight;*acd&*tcc;route1;*string:~*req.Account:1001&*string:~*req.Account:1002;1001&1002;RP1&RP2;RS1&RS2;Stat_1&Stat_1_1;10;true;param;10;key:value",
			expectedErr: `parsing time "bad String" as "2006-01-02T15:04:05Z07:00": cannot parse "bad String" as "2006"`,
		},
		{
			name:        "ActivationIntervalBadStringFail2",
			extraParams: "cgrates.org;RTP_ACNT_1001;*string:~*req.Account:1001&*string:~*req.Account:1002;2014-07-29T15:00:00Z&bad String;*weight;*acd&*tcc;route1;*string:~*req.Account:1001&*string:~*req.Account:1002;1001&1002;RP1&RP2;RS1&RS2;Stat_1&Stat_1_1;10;true;param;10;key:value",
			expectedErr: `parsing time "bad String" as "2006-01-02T15:04:05Z07:00": cannot parse "bad String" as "2006"`,
		},
		{
			name:        "RouteWeightFail",
			extraParams: "cgrates.org;RTP_ACNT_1001;*string:~*req.Account:1001&*string:~*req.Account:1002;2014-07-29T15:00:00Z;*weight;*acd&*tcc;route1;*string:~*req.Account:1001&*string:~*req.Account:1002;1001&1002;RP1&RP2;RS1&RS2;Stat_1&Stat_1_1;BadString;true;param;10;key:value",
			expectedErr: `strconv.ParseFloat: parsing "BadString": invalid syntax`,
		},
		{
			name:        "RouteBlockerFail",
			extraParams: "cgrates.org;RTP_ACNT_1001;*string:~*req.Account:1001&*string:~*req.Account:1002;2014-07-29T15:00:00Z;*weight;*acd&*tcc;route1;*string:~*req.Account:1001&*string:~*req.Account:1002;1001&1002;RP1&RP2;RS1&RS2;Stat_1&Stat_1_1;10;BadString;param;10;key:value",
			expectedErr: `strconv.ParseBool: parsing "BadString": invalid syntax`,
		},
		{
			name:        "WeightFail",
			extraParams: "cgrates.org;RTP_ACNT_1001;*string:~*req.Account:1001&*string:~*req.Account:1002;2014-07-29T15:00:00Z;*weight;*acd&*tcc;route1;*string:~*req.Account:1001&*string:~*req.Account:1002;1001&1002;RP1&RP2;RS1&RS2;Stat_1&Stat_1_1;10;true;param;BadString;key:value",
			expectedErr: `strconv.ParseFloat: parsing "BadString": invalid syntax`,
		},
		{
			name:        "InvalidOptsMap",
			extraParams: "cgrates.org;RTP_ACNT_1001;*string:~*req.Account:1001&*string:~*req.Account:1002;2014-07-29T15:00:00Z;*weight;*acd&*tcc;route1;*string:~*req.Account:1001&*string:~*req.Account:1002;1001&1002;RP1&RP2;RS1&RS2;Stat_1&Stat_1_1;10;true;param;10;opt",
			expectedErr: "invalid key-value pair: opt",
		},
	}

	for i, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			action := &Action{ExtraParameters: tc.extraParams}
			ev := &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "evID",
				Time:   &time.Time{},
				Event: map[string]any{
					utils.AccountField: "1001",
				},
			}
			t.Cleanup(func() {
				rwo = nil
			})
			err := dynamicRoute(nil, action, nil, nil, ev,
				SharedActionsData{}, ActionConnCfg{
					ConnIDs: tc.connIDs,
				})
			if tc.expectedErr != "" {
				if err == nil || err.Error() != tc.expectedErr {
					t.Errorf("expected error <%v>, received <%v>", tc.expectedErr, err)
				}
			} else if err != nil {
				t.Error(err)
			} else if !reflect.DeepEqual(rwo, tc.expRwo) {
				if i != 1 {
					t.Errorf("Expected <%v>\nReceived\n<%v>", utils.ToJSON(tc.expRwo), utils.ToJSON(rwo))
				} else {
					// Get the absolute difference between the times
					diff := rwo.ActivationInterval.ActivationTime.Sub(tc.expRwo.ActivationInterval.ActivationTime)
					if diff < 0 {
						diff = -diff // Make sure it's positive
					}
					// Check if difference is less than or equal to 1 second
					if diff <= time.Second {
						tc.expRwo.ActivationInterval.ActivationTime = rwo.ActivationInterval.ActivationTime
						if !reflect.DeepEqual(rwo, tc.expRwo) {
							t.Errorf("Expected <%v>\nReceived\n<%v>", utils.ToJSON(tc.expRwo), utils.ToJSON(rwo))
						}
					} else {
						t.Errorf("Expected <%v>\nReceived\n<%v>", utils.ToJSON(tc.expRwo), utils.ToJSON(rwo))
					}
				}
			}
		})
	}
}

func TestDynamicRanking(t *testing.T) {
	tempConn := connMgr
	tmpDm := dm
	tmpCache := Cache
	defer func() {
		config.SetCgrConfig(config.NewDefaultCGRConfig())
		SetConnManager(tempConn)
		dm = tmpDm
		Cache = tmpCache
	}()
	Cache.Clear(nil)
	var rpwo *RankingProfileWithAPIOpts
	ccMock := &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.APIerSv1SetRankingProfile: func(ctx *context.Context, args, reply any) error {
				var canCast bool
				if rpwo, canCast = args.(*RankingProfileWithAPIOpts); !canCast {
					return fmt.Errorf("couldnt cast")
				}
				return nil
			},
		},
	}
	connID := utils.ConcatenatedKey(utils.MetaInternal, utils.MetaApier)
	clientconn := make(chan birpc.ClientConnector, 1)
	clientconn <- ccMock
	NewConnManager(config.NewDefaultCGRConfig(), map[string]chan birpc.ClientConnector{
		connID: clientconn,
	})
	testcases := []struct {
		name        string
		extraParams string
		connIDs     []string
		expRpwo     *RankingProfileWithAPIOpts
		expectedErr string
	}{
		{
			name:    "SuccessfulRequest",
			connIDs: []string{connID},
			expRpwo: &RankingProfileWithAPIOpts{
				RankingProfile: &RankingProfile{
					Tenant:            "cgrates.org",
					ID:                "RKNG_ACNT_1001",
					Schedule:          "@every 15m",
					StatIDs:           []string{"Stats2", "Stats3", "Stats4"},
					MetricIDs:         []string{"Metric1", "Metric3"},
					Sorting:           "*asc",
					SortingParameters: []string{"metricA:true", "metricB:false"},
					Stored:            true,
					ThresholdIDs:      []string{"THD1", "THD2"},
				},
				APIOpts: map[string]any{
					"key": "value",
				},
			},
			extraParams: "cgrates.org;RKNG_ACNT_1001;@every 15m;Stats2&Stats3&Stats4;Metric1&Metric3;*asc;metricA:true&metricB:false;true;THD1&THD2;key:value",
		},
		{
			name:    "SuccessfulRequestWithDynamicPaths",
			connIDs: []string{connID},
			expRpwo: &RankingProfileWithAPIOpts{
				RankingProfile: &RankingProfile{
					Tenant:            "cgrates.org",
					ID:                "RKNG_ACNT_1001",
					Schedule:          "@every 15m",
					StatIDs:           []string{"Stats2", "Stats3", "Stats4"},
					MetricIDs:         []string{"Metric1", "Metric3"},
					Sorting:           "*asc",
					SortingParameters: []string{"metricA:true", "metricB:false"},
					Stored:            true,
					ThresholdIDs:      []string{"THD1", "THD2"},
				},
				APIOpts: map[string]any{
					"key": "value",
				},
			},
			extraParams: "*tenant;RKNG_ACNT_<~*req.Account>;@every 15m;Stats2&Stats3&Stats4;Metric1&Metric3;*asc;metricA:true&metricB:false;true;THD1&THD2;key:value",
		},
		{
			name:    "SuccessfulRequestEmptyFields",
			connIDs: []string{connID},
			expRpwo: &RankingProfileWithAPIOpts{
				RankingProfile: &RankingProfile{
					Tenant:            "cgrates.org",
					ID:                "RKNG_ACNT_1001",
					Schedule:          "@every 15m",
					StatIDs:           nil,
					MetricIDs:         nil,
					Sorting:           "",
					SortingParameters: nil,
					Stored:            false,
					ThresholdIDs:      nil,
				},
				APIOpts: map[string]any{},
			},
			extraParams: "cgrates.org;RKNG_ACNT_1001;@every 15m;;;;;;;",
		},
		{
			name:        "MissingConns",
			extraParams: "cgrates.org;RKNG_ACNT_1001;@every 15m;Stats2&Stats3&Stats4;Metric1&Metric3;*asc;metricA:true&metricB:false;true;THD1&THD2;key:value",
			expectedErr: "MANDATORY_IE_MISSING: [connIDs]",
		},
		{
			name:        "WrongNumberOfParams",
			extraParams: "tenant;RKNG;;",
			expectedErr: "invalid number of parameters <4> expected 10",
		},
		{
			name:        "RankingStoredFail",
			extraParams: "cgrates.org;RKNG_ACNT_1001;@every 15m;Stats2&Stats3&Stats4;Metric1&Metric3;*asc;metricA:true&metricB:false;BadString;THD1&THD2;key:value",
			expectedErr: `strconv.ParseBool: parsing "BadString": invalid syntax`,
		},
		{
			name:        "InvalidOptsMap",
			extraParams: "cgrates.org;RKNG_ACNT_1001;@every 15m;Stats2&Stats3&Stats4;Metric1&Metric3;*asc;metricA:true&metricB:false;true;THD1&THD2;opt",
			expectedErr: "invalid key-value pair: opt",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			action := &Action{ExtraParameters: tc.extraParams}
			ev := &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "evID",
				Time:   &time.Time{},
				Event: map[string]any{
					utils.AccountField: "1001",
				},
			}
			t.Cleanup(func() {
				rpwo = nil
			})
			err := dynamicRanking(nil, action, nil, nil, ev,
				SharedActionsData{}, ActionConnCfg{
					ConnIDs: tc.connIDs,
				})
			if tc.expectedErr != "" {
				if err == nil || err.Error() != tc.expectedErr {
					t.Errorf("expected error <%v>, received <%v>", tc.expectedErr, err)
				}
			} else if err != nil {
				t.Error(err)
			} else if !reflect.DeepEqual(rpwo, tc.expRpwo) {
				t.Errorf("Expected <%v>\nReceived\n<%v>", utils.ToJSON(tc.expRpwo), utils.ToJSON(rpwo))
			}
		})
	}
}

func TestDynamicRatingProfile(t *testing.T) {
	tempConn := connMgr
	tmpDm := dm
	tmpCache := Cache
	defer func() {
		config.SetCgrConfig(config.NewDefaultCGRConfig())
		SetConnManager(tempConn)
		dm = tmpDm
		Cache = tmpCache
	}()
	Cache.Clear(nil)
	var rpwo *utils.AttrSetRatingProfile
	ccMock := &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.APIerSv1SetRatingProfile: func(ctx *context.Context, args, reply any) error {
				var canCast bool
				if rpwo, canCast = args.(*utils.AttrSetRatingProfile); !canCast {
					return fmt.Errorf("couldnt cast")
				}
				return nil
			},
		},
	}
	connID := utils.ConcatenatedKey(utils.MetaInternal, utils.MetaApier)
	clientconn := make(chan birpc.ClientConnector, 1)
	clientconn <- ccMock
	NewConnManager(config.NewDefaultCGRConfig(), map[string]chan birpc.ClientConnector{
		connID: clientconn,
	})
	testcases := []struct {
		name        string
		extraParams string
		connIDs     []string
		expRpwo     *utils.AttrSetRatingProfile
		expectedErr string
	}{
		{
			name:    "SuccessfulRequest",
			connIDs: []string{connID},
			expRpwo: &utils.AttrSetRatingProfile{
				Tenant:   "cgrates.org",
				Category: "call",
				Subject:  utils.MetaAny,
				RatingPlanActivations: []*utils.TPRatingActivation{
					{
						ActivationTime:   "2014-07-29T15:00:00Z",
						RatingPlanId:     "RP_TESTIT1",
						FallbackSubjects: "RP_TEST",
					},
				},
				APIOpts: map[string]any{
					"key": "value",
				},
			},
			extraParams: "cgrates.org;call;*any;2014-07-29T15:00:00Z;RP_TESTIT1;RP_TEST;key:value",
		},
		{
			name:    "SuccessfulRequestWithDynamicPaths",
			connIDs: []string{connID},
			expRpwo: &utils.AttrSetRatingProfile{
				Tenant:   "cgrates.org",
				Category: "call",
				Subject:  "1001",
				RatingPlanActivations: []*utils.TPRatingActivation{
					{
						ActivationTime:   "*now",
						RatingPlanId:     "RP_TEST1001",
						FallbackSubjects: "RP_TEST",
					},
				},
				APIOpts: map[string]any{
					"key": "value",
				},
			},
			extraParams: "*tenant;call;~*req.Account;*now;RP_TEST<~*req.Account>;RP_TEST;key:value",
		},
		{
			name:        "MissingConns",
			extraParams: "cgrates.org;call;*any;2014-07-29T15:00:00Z;RP_TESTIT1;RP_TEST;key:value",
			expectedErr: "MANDATORY_IE_MISSING: [connIDs]",
		},
		{
			name:        "WrongNumberOfParams",
			extraParams: "cgrates.org;;;",
			expectedErr: "invalid number of parameters <4> expected 7",
		},
		{
			name:        "InvalidOptsMap",
			extraParams: "cgrates.org;call;*any;2014-07-29T15:00:00Z;RP_TESTIT1;RP_TEST;opt",
			expectedErr: "invalid key-value pair: opt",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			action := &Action{ExtraParameters: tc.extraParams}
			ev := &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "evID",
				Time:   &time.Time{},
				Event: map[string]any{
					utils.AccountField: "1001",
				},
			}
			t.Cleanup(func() {
				rpwo = nil
			})
			err := dynamicRatingProfile(nil, action, nil, nil, ev,
				SharedActionsData{}, ActionConnCfg{
					ConnIDs: tc.connIDs,
				})
			if tc.expectedErr != "" {
				if err == nil || err.Error() != tc.expectedErr {
					t.Errorf("expected error <%v>, received <%v>", tc.expectedErr, err)
				}
			} else if err != nil {
				t.Error(err)
			} else if !reflect.DeepEqual(rpwo, tc.expRpwo) {
				t.Errorf("Expected <%v>\nReceived\n<%v>", utils.ToJSON(tc.expRpwo), utils.ToJSON(rpwo))
			}
		})
	}
}
