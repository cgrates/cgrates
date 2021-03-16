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
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"

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
	at := &ActionTiming{Timing: &RateInterval{Timing: &RITiming{Years: utils.Years{2022}, StartTime: "10:01:00"}}}
	st := at.GetNextStartTime(referenceDate)
	expected := time.Date(2022, 1, 1, 10, 1, 0, 0, time.Local)
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
		y, m, d = testTime.Date()
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
	at := &ActionTiming{Timing: &RateInterval{Timing: &RITiming{Years: utils.Years{2023}}}}
	st := at.GetNextStartTime(referenceDate)
	expected := time.Date(2023, 1, 1, 0, 0, 0, 0, time.Local)
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
	err := at.Execute(nil, nil)
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
	err := at.Execute(nil, nil)
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
	dm.GetAccountActionPlans(account1.ID, true, utils.NonTransactional) // FixMe: remove here after finishing testing of map
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

	dm.SetActions(actions[0].Id, actions, utils.NonTransactional)

	at := &ActionTiming{
		accountIDs: utils.StringMap{account1.ID: true},
		Timing:     &RateInterval{},
		actions:    actions,
	}

	if err = at.Execute(nil, nil); err != nil {
		t.Errorf("Execute Action: %v", err)
	}

	apr, err1 := dm.GetActionPlan(ap1.Id, false, utils.NonTransactional)

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
	}, ExtraParameters: fmt.Sprintf(`{"UniqueID":"ZIP", "GroupID":"TEST", "ThresholdType":"TT"}`)}
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
	resetTriggersAction(ub, nil, nil, nil)
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
	resetTriggersAction(ub, nil, nil, nil)
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
	resetTriggersAction(ub, &Action{Balance: &BalanceFilter{Type: utils.StringPointer(utils.MetaSMS)}}, nil, nil)
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
	resetTriggersAction(ub, &Action{}, nil, nil)
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
	allowNegativeAction(ub, nil, nil, nil)
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
	denyNegativeAction(ub, nil, nil, nil)
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
	resetAccountAction(ub, nil, nil, nil)
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
	resetAccountAction(ub, nil, nil, nil)
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
	topupResetAction(ub, a, nil, nil)
	if ub.AllowNegative ||
		ub.BalanceMap[utils.MetaMonetary].GetTotalValue() != 10 ||
		len(ub.UnitCounters) != 0 || // InitCounters finds no counters
		len(ub.BalanceMap[utils.MetaVoice]) != 2 ||
		ub.ActionTriggers[0].Executed != true || ub.ActionTriggers[1].Executed != true {
		t.Errorf("Topup reset action failed: %+s", utils.ToIJSON(ub))
	}
}

func TestActionTopupValueFactor(t *testing.T) {
	ub := &Account{
		ID:         "TEST_UB",
		BalanceMap: map[string]Balances{},
	}
	a := &Action{
		Balance: &BalanceFilter{
			Type:  utils.StringPointer(utils.MetaMonetary),
			Value: &utils.ValueFormula{Static: 10},
		},
		ExtraParameters: `{"*monetary":2.0}`,
	}
	topupResetAction(ub, a, nil, nil)
	if len(ub.BalanceMap) != 1 ||
		ub.BalanceMap[utils.MetaMonetary][0].Factor[utils.MetaMonetary] != 2.0 {
		t.Errorf("Topup reset action failed to set Factor: %+v",
			ub.BalanceMap[utils.MetaMonetary][0].Factor)
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
	topupResetAction(ub, a, nil, nil)
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
	topupResetAction(ub, a, nil, nil)
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
	topupResetAction(ub, a, nil, nil)
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
	topupAction(ub, a, nil, nil)
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
	topupAction(ub, a, nil, nil)
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
	debitAction(ub, a, nil, nil)
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
	debitAction(ub, a, nil, nil)
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
	resetCountersAction(ub, nil, nil, nil)
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
	resetCountersAction(ub, a, nil, nil)
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
	resetCountersAction(ub, a, nil, nil)
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
	at.Execute(nil, nil)
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

	at.Execute(nil, nil)
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

	at.Execute(nil, nil)
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
	err = at.Execute(nil, nil)
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
	err = at.Execute(nil, nil)
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
	err = at.Execute(nil, nil)
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
	err = at.Execute(nil, nil)
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
	err = at.Execute(nil, nil)
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
	err = at.Execute(nil, nil)
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
	at.Execute(nil, nil)

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
		t.Error("ransfer balance value: ", afterUb.BalanceMap[utils.MetaMonetary].GetTotalValue())
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
	at.Execute(nil, nil)

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
		t.Error("ransfer balance value: ", afterUb.BalanceMap[utils.MetaMonetary].GetTotalValue())
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
		Filter:     `{"Type":"*monetary","Value":1,"Weight":10}`,
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
	at.Execute(nil, nil)

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
		t.Error("ransfer balance value: ", afterUb.BalanceMap[utils.MetaMonetary].GetTotalValue())
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
		Filter:     `{"Type":"*monetary","Value":2,"Weight":10}`,
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
	at.Execute(nil, nil)

	afterUb, err := dm.GetAccount("cgrates.org:cond")
	if err != nil {
		t.Error("account not found: ", err, afterUb)
	}
	if len(afterUb.BalanceMap[utils.MetaMonetary]) != 4 ||
		afterUb.BalanceMap[utils.MetaMonetary].GetTotalValue() != 20 {
		for _, b := range afterUb.BalanceMap[utils.MetaMonetary] {
			t.Logf("B: %+v", b)
		}
		t.Error("ransfer balance value: ", afterUb.BalanceMap[utils.MetaMonetary].GetTotalValue())
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
		Filter:     `{"Type":"*voice","Value":{"*gte":100}}`,
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
	at.Execute(nil, nil)

	afterUb, err := dm.GetAccount("cgrates.org:cond")
	if err != nil {
		t.Error("account not found: ", err, afterUb)
	}
	if len(afterUb.BalanceMap[utils.MetaMonetary]) != 2 ||
		afterUb.BalanceMap[utils.MetaMonetary].GetTotalValue() != 18 {
		for _, b := range afterUb.BalanceMap[utils.MetaMonetary] {
			t.Logf("B: %+v", b)
		}
		t.Error("ransfer balance value: ", afterUb.BalanceMap[utils.MetaMonetary].GetTotalValue())
	}
}

func TestActionConditionalDisabledIfNegative(t *testing.T) {
	err := dm.SetAccount(
		&Account{
			ID: "cgrates.org:af",
			BalanceMap: map[string]Balances{
				"*data": {
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
				"*monetary": {
					&Balance{
						Uuid:  "9fa1847a-f36a-41a7-8ec0-dfaab370141e",
						ID:    utils.MetaDefault,
						Value: -1.95001,
					},
				},
				"*sms": {
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
				"*voice": {
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
		Filter:     "{\"*and\":[{\"Value\":{\"*lt\":0}},{\"ID\":{\"*eq\":\"*default\"}}]}",
		Balance: &BalanceFilter{
			Type:     utils.StringPointer("*sms"),
			ID:       utils.StringPointer("for_v3hsillmilld500m_sms_ill"),
			Disabled: utils.BoolPointer(true),
		},
		Weight: 9,
	}
	a2 := &Action{
		ActionType: utils.MetaSetBalance,
		Filter:     "{\"*and\":[{\"Value\":{\"*lt\":0}},{\"ID\":{\"*eq\":\"*default\"}}]}",
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
		Filter:     "{\"*and\":[{\"Value\":{\"*lt\":0}},{\"ID\":{\"*eq\":\"*default\"}}]}",
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
		Filter:     "{\"*and\":[{\"Value\":{\"*lt\":0}},{\"ID\":{\"*eq\":\"*default\"}}]}",
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
		Filter:     "{\"*and\":[{\"Value\":{\"*lt\":0}},{\"ID\":{\"*eq\":\"*default\"}}]}",
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
	at.Execute(nil, nil)

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
	at.Execute(nil, nil)

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
	if len(act) != 1 || act[0].Filter != `{"*and":[{"Value":{"*lt":0}},{"Id":{"*eq":"*default"}}]}` {
		t.Error("Error loading actions: ", act[0].Filter)
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
		at.Execute(nil, nil)
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
	at.Execute(nil, nil)
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
	err = at.Execute(nil, nil)
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
	err = at.Execute(nil, nil)
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

func (trpcp *TestRPCParameters) Hopa(in Attr, out *float64) error {
	trpcp.status = utils.OK
	return nil
}

func (trpcp *TestRPCParameters) Call(serviceMethod string, args interface{}, reply interface{}) error {
	parts := strings.Split(serviceMethod, ".")
	if len(parts) != 2 {
		return utils.ErrNotImplemented
	}
	// get method
	method := reflect.ValueOf(trpcp).MethodByName(parts[1])
	if !method.IsValid() {
		return utils.ErrNotImplemented
	}

	// construct the params
	params := []reflect.Value{reflect.ValueOf(args).Elem(), reflect.ValueOf(reply)}

	ret := method.Call(params)
	if len(ret) != 1 {
		return utils.ErrServerError
	}
	if ret[0].Interface() == nil {
		return nil
	}
	err, ok := ret[0].Interface().(error)
	if !ok {
		return utils.ErrServerError
	}
	return err
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
	if err := cgrRPCAction(nil, a, nil, nil); err != nil {
		t.Error("error executing cgr action: ", err)
	}
	if trpcp.status != utils.OK {
		t.Error("RPC not called!")
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
	at.Execute(nil, nil)
	afterUb, err := dm.GetAccount("cgrates.org:vf")
	// not an exact value, depends of month
	v := afterUb.BalanceMap[utils.MetaMonetary].GetTotalValue()
	if err != nil || v > -0.30 || v < -0.36 {
		t.Error("error debiting account: ", err, utils.ToIJSON(afterUb), v)
	}
}

func TestClonedAction(t *testing.T) {
	a := &Action{
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
	if clone, err := actions.Clone(); err != nil {
		t.Error("error cloning actions: ", err)
	} else if !reflect.DeepEqual(actions, clone) {
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
	clned, err := Cache.GetCloned(utils.CacheActions, "MYTEST")
	if err != nil {
		t.Error(err)
	}
	aCloned := clned.(Actions)
	if !reflect.DeepEqual(actions, aCloned) {
		t.Errorf("Expecting: %+v, received: %+v", actions[1].Balance, aCloned[1].Balance)
	}
}

// TestCdrLogAction
type RPCMock struct {
	args *ArgV1ProcessEvent
}

func (r *RPCMock) Call(method string, args interface{}, rply interface{}) error {
	if method != utils.CDRsV1ProcessEvent {
		return rpcclient.ErrUnsupporteServiceMethod
	}
	if r.args != nil {
		return fmt.Errorf("There should be only one call to this function")
	}
	r.args = args.(*ArgV1ProcessEvent)
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

	internalChan := make(chan rpcclient.ClientConnector, 1)
	internalChan <- &mock

	NewConnManager(dfltCfg, map[string]chan rpcclient.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCDRs): internalChan,
	})

	var extraData interface{}
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
	if err := cdrLogAction(acc, a, acs, extraData); err != nil {
		t.Fatal(err)
	}
	if mock.args == nil {
		t.Fatalf("Expected a call to %s", utils.CDRsV1ProcessEvent)
	}
	expCgrEv := utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     mock.args.CGREvent.ID,
		Event: map[string]interface{}{
			"Account":      "1001",
			"ActionID":     "CdrDebit",
			"AnswerTime":   mock.args.CGREvent.Event["AnswerTime"],
			"BalanceID":    "*default",
			"BalanceValue": "10",
			"CGRID":        mock.args.CGREvent.Event["CGRID"],
			"Category":     "",
			"Cost":         9.95,
			"CostSource":   "",
			"Destination":  "",
			"ExtraInfo":    "",
			"OrderID":      mock.args.CGREvent.Event["OrderID"],
			"OriginHost":   "127.0.0.1",
			"OriginID":     mock.args.CGREvent.Event["OriginID"],
			"Partial":      false,
			"PreRated":     true,
			"RequestType":  "*none",
			"RunID":        "*debit",
			"SetupTime":    mock.args.CGREvent.Event["SetupTime"],
			"Source":       utils.CDRLog,
			"Subject":      "1001",
			"Tenant":       "cgrates.org",
			"ToR":          "*monetary",
			"Usage":        mock.args.CGREvent.Event["Usage"],
		},
		Opts: map[string]interface{}{},
	}
	if !reflect.DeepEqual(expCgrEv, mock.args.CGREvent) {
		t.Errorf("Expected: %+v \n,received: %+v", expCgrEv, mock.args.CGREvent)
	}
}

func TestRemoteSetAccountAction(t *testing.T) {
	expError := `Post "127.1.0.11//": unsupported protocol scheme ""`
	if err = remoteSetAccount(nil, &Action{ExtraParameters: "127.1.0.11//"}, nil, nil); err == nil ||
		err.Error() != expError {
		t.Fatalf("Expected error: %s, received: %v", expError, err)
	}
	expError = `json: unsupported type: func()`
	if err = remoteSetAccount(&Account{
		ActionTriggers: ActionTriggers{{
			Balance: &BalanceFilter{
				Value: &utils.ValueFormula{
					Params: map[string]interface{}{utils.MetaVoice: func() {}},
				},
			},
		}},
	}, &Action{ExtraParameters: "127.1.0.11//"}, nil, nil); err == nil ||
		err.Error() != expError {
		t.Fatalf("Expected error: %s, received: %v", expError, err)
	}

	ts := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) { rw.Write([]byte("5")) }))
	acc := &Account{ID: "1001"}
	expError = `json: cannot unmarshal number into Go value of type engine.Account`
	if err = remoteSetAccount(acc, &Action{ExtraParameters: ts.URL}, nil, nil); err == nil ||
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
	if err = remoteSetAccount(acc, &Action{ExtraParameters: ts.URL}, nil, nil); err != nil {
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
