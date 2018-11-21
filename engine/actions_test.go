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
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/cgrates/cgrates/utils"
)

var err error
var (
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
	at := &ActionTiming{Timing: &RateInterval{Timing: &RITiming{StartTime: utils.ASAP}}}
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
	if err != nil {
		t.Errorf("Faild to detect wrong function type: %v", err)
	}
}

func TestActionTimingPriorityListSortByWeight(t *testing.T) {
	at1 := &ActionTiming{Timing: &RateInterval{
		Timing: &RITiming{
			Years: utils.Years{2020},
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
			Years: utils.Years{2020},
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
		t.Error("Timing list not sorted correctly: ", at1, at2, atpl)
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

	dm.DataDB().SetAccount(account1)
	dm.DataDB().SetAccount(account2)

	ap1 := &ActionPlan{
		Id:         "TestActionPlansRemoveMember1",
		AccountIDs: utils.StringMap{"one": true},
		ActionTimings: []*ActionTiming{
			&ActionTiming{
				Uuid: "uuid1",
				Timing: &RateInterval{
					Timing: &RITiming{
						Years:     utils.Years{2012},
						Months:    utils.Months{},
						MonthDays: utils.MonthDays{},
						WeekDays:  utils.WeekDays{},
						StartTime: utils.ASAP,
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
			&ActionTiming{
				Uuid: "uuid2",
				Timing: &RateInterval{
					Timing: &RITiming{
						Years:     utils.Years{2012},
						Months:    utils.Months{},
						MonthDays: utils.MonthDays{},
						WeekDays:  utils.WeekDays{},
						StartTime: utils.ASAP,
					},
				},
				Weight:    10,
				ActionsID: "MINI",
			},
		},
	}

	if err := dm.DataDB().SetActionPlan(ap1.Id, ap1, true,
		utils.NonTransactional); err != nil {
		t.Error(err)
	}
	if err = dm.DataDB().SetActionPlan(ap2.Id, ap2, true,
		utils.NonTransactional); err != nil {
		t.Error(err)
	}
	if err = dm.CacheDataFromDB(utils.ACTION_PLAN_PREFIX,
		[]string{ap1.Id, ap2.Id}, true); err != nil {
		t.Error(err)
	}
	if err = dm.DataDB().SetAccountActionPlans(account1.ID,
		[]string{ap1.Id}, false); err != nil {
		t.Error(err)
	}
	if err = dm.CacheDataFromDB(utils.AccountActionPlansPrefix,
		[]string{account1.ID}, true); err != nil {
		t.Error(err)
	}
	dm.DataDB().GetAccountActionPlans(account1.ID, true, utils.NonTransactional) // FixMe: remove here after finishing testing of map
	if err = dm.DataDB().SetAccountActionPlans(account2.ID,
		[]string{ap2.Id}, false); err != nil {
		t.Error(err)
	}
	if err = dm.CacheDataFromDB(utils.AccountActionPlansPrefix,
		[]string{account2.ID}, false); err != nil {
		t.Error(err)
	}

	actions := []*Action{
		&Action{
			Id:         "REMOVE",
			ActionType: REMOVE_ACCOUNT,
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

	apr, err1 := dm.DataDB().GetActionPlan(ap1.Id, false, utils.NonTransactional)

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
			Type:       utils.StringPointer(utils.MONETARY),
			Directions: utils.StringMapPointer(utils.NewStringMap(utils.OUT)),
		},
		ThresholdType:  utils.TRIGGER_MAX_BALANCE,
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
			Type:       utils.StringPointer(utils.MONETARY),
			Directions: utils.StringMapPointer(utils.NewStringMap(utils.OUT)),
		},
		ThresholdType:  utils.TRIGGER_MAX_BALANCE,
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
			Type:       utils.StringPointer(utils.MONETARY),
			Directions: utils.StringMapPointer(utils.NewStringMap(utils.OUT)),
		},
		ThresholdType:  utils.TRIGGER_MAX_BALANCE,
		ThresholdValue: 2,
	}
	a := &Action{Balance: &BalanceFilter{Type: utils.StringPointer(utils.MONETARY)},
		ExtraParameters: `{"BalanceDirections":"*out"}`}
	if !at.Match(a) {
		t.Errorf("Action trigger [%v] does not match action [%v]", at, a)
	}
}

func TestActionTriggerMatchMinuteBucketFull(t *testing.T) {
	at := &ActionTrigger{
		Balance: &BalanceFilter{
			Type:       utils.StringPointer(utils.MONETARY),
			Directions: utils.StringMapPointer(utils.NewStringMap(utils.OUT)),
		},
		ThresholdType:  utils.TRIGGER_MAX_BALANCE,
		ThresholdValue: 2,
	}
	a := &Action{ExtraParameters: fmt.Sprintf(`{"ThresholdType":"%v", "ThresholdValue": %v}`,
		utils.TRIGGER_MAX_BALANCE, 2)}
	if !at.Match(a) {
		t.Errorf("Action trigger [%v] does not match action [%v]", at, a)
	}
}

func TestActionTriggerMatchAllFull(t *testing.T) {
	at := &ActionTrigger{
		Balance: &BalanceFilter{
			Type:       utils.StringPointer(utils.MONETARY),
			Directions: utils.StringMapPointer(utils.NewStringMap(utils.OUT)),
		},
		ThresholdType:  utils.TRIGGER_MAX_BALANCE,
		ThresholdValue: 2,
	}
	a := &Action{Balance: &BalanceFilter{Type: utils.StringPointer(utils.MONETARY)},
		ExtraParameters: fmt.Sprintf(`{"ThresholdType":"%v", "ThresholdValue": %v, "BalanceDirections":"*out"}`,
			utils.TRIGGER_MAX_BALANCE, 2)}
	if !at.Match(a) {
		t.Errorf("Action trigger [%v] does not match action [%v]", at, a)
	}
}

func TestActionTriggerMatchSomeFalse(t *testing.T) {
	at := &ActionTrigger{
		Balance: &BalanceFilter{
			Type:       utils.StringPointer(utils.MONETARY),
			Directions: utils.StringMapPointer(utils.NewStringMap(utils.OUT)),
		},
		ThresholdType:  utils.TRIGGER_MAX_BALANCE,
		ThresholdValue: 2,
	}
	a := &Action{Balance: &BalanceFilter{Type: utils.StringPointer(utils.MONETARY)},
		ExtraParameters: fmt.Sprintf(`{"ThresholdType":"%s"}`,
			utils.TRIGGER_MAX_BALANCE_COUNTER)}
	if at.Match(a) {
		t.Errorf("Action trigger [%v] does not match action [%v]", at, a)
	}
}

func TestActionTriggerMatcBalanceFalse(t *testing.T) {
	at := &ActionTrigger{
		Balance: &BalanceFilter{
			Type:       utils.StringPointer(utils.MONETARY),
			Directions: utils.StringMapPointer(utils.NewStringMap(utils.OUT)),
		},
		ThresholdType:  utils.TRIGGER_MAX_BALANCE,
		ThresholdValue: 2,
	}
	a := &Action{Balance: &BalanceFilter{Type: utils.StringPointer(utils.MONETARY),
		Directions: utils.StringMapPointer(utils.NewStringMap(utils.OUT))},
		ExtraParameters: fmt.Sprintf(`{"GroupID":"%s", "ThresholdType":"%s"}`, "TEST", utils.TRIGGER_MAX_BALANCE)}
	if at.Match(a) {
		t.Errorf("Action trigger [%v] does not match action [%v]", at, a)
	}
}

func TestActionTriggerMatcAllFalse(t *testing.T) {
	at := &ActionTrigger{
		Balance: &BalanceFilter{
			Type:       utils.StringPointer(utils.MONETARY),
			Directions: utils.StringMapPointer(utils.NewStringMap(utils.OUT)),
		},
		ThresholdType:  utils.TRIGGER_MAX_BALANCE,
		ThresholdValue: 2,
	}
	a := &Action{Balance: &BalanceFilter{Type: utils.StringPointer(utils.MONETARY)},
		ExtraParameters: fmt.Sprintf(`{"UniqueID":"ZIP", "GroupID":"%s", "ThresholdType":"%s"}`, "TEST", utils.TRIGGER_MAX_BALANCE)}
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
			Type:           utils.StringPointer(utils.MONETARY),
			RatingSubject:  utils.StringPointer("test1"),
			Directions:     utils.StringMapPointer(utils.NewStringMap(utils.OUT)),
			Value:          &utils.ValueFormula{Static: 2},
			Weight:         utils.Float64Pointer(1.0),
			DestinationIDs: utils.StringMapPointer(utils.NewStringMap("NAT")),
			SharedGroups:   utils.StringMapPointer(utils.NewStringMap("test2")),
		},
	}
	a := &Action{Balance: &BalanceFilter{
		Type:           utils.StringPointer(utils.MONETARY),
		RatingSubject:  utils.StringPointer("test1"),
		Directions:     utils.StringMapPointer(utils.NewStringMap(utils.OUT)),
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
		BalanceMap: map[string]Balances{utils.MONETARY: Balances{&Balance{Value: 10}},
			utils.VOICE: Balances{&Balance{Value: 10, Weight: 20, DestinationIDs: utils.NewStringMap("NAT")},
				&Balance{Weight: 10, DestinationIDs: utils.NewStringMap("RET")}}},
		UnitCounters: UnitCounters{utils.MONETARY: []*UnitCounter{&UnitCounter{Counters: CounterFilters{&CounterFilter{Value: 1}}}}},
		ActionTriggers: ActionTriggers{&ActionTrigger{Balance: &BalanceFilter{
			Type: utils.StringPointer(utils.MONETARY)}, ThresholdValue: 2, ActionsID: "TEST_ACTIONS", Executed: true},
			&ActionTrigger{Balance: &BalanceFilter{
				Type: utils.StringPointer(utils.MONETARY)}, ThresholdValue: 2, ActionsID: "TEST_ACTIONS", Executed: true}},
	}
	resetTriggersAction(ub, nil, nil, nil)
	if ub.ActionTriggers[0].Executed == true || ub.ActionTriggers[1].Executed == true {
		t.Error("Reset triggers action failed!")
	}
}

func TestActionResetTriggresExecutesThem(t *testing.T) {
	ub := &Account{
		ID:         "TEST_UB",
		BalanceMap: map[string]Balances{utils.MONETARY: Balances{&Balance{Value: 10}}},
		UnitCounters: UnitCounters{utils.MONETARY: []*UnitCounter{
			&UnitCounter{Counters: CounterFilters{&CounterFilter{Value: 1}}}}},
		ActionTriggers: ActionTriggers{&ActionTrigger{Balance: &BalanceFilter{
			Type: utils.StringPointer(utils.MONETARY)}, ThresholdValue: 2, ActionsID: "TEST_ACTIONS", Executed: true}},
	}
	resetTriggersAction(ub, nil, nil, nil)
	if ub.ActionTriggers[0].Executed == true || ub.BalanceMap[utils.MONETARY][0].GetValue() == 12 {
		t.Error("Reset triggers action failed!")
	}
}

func TestActionResetTriggresActionFilter(t *testing.T) {
	ub := &Account{
		ID: "TEST_UB",
		BalanceMap: map[string]Balances{utils.MONETARY: Balances{
			&Balance{Value: 10}}, utils.VOICE: Balances{
			&Balance{Value: 10, Weight: 20, DestinationIDs: utils.NewStringMap("NAT")},
			&Balance{Weight: 10, DestinationIDs: utils.NewStringMap("RET")}}},
		UnitCounters: UnitCounters{utils.MONETARY: []*UnitCounter{
			&UnitCounter{Counters: CounterFilters{&CounterFilter{Value: 1}}}}},
		ActionTriggers: ActionTriggers{
			&ActionTrigger{Balance: &BalanceFilter{Type: utils.StringPointer(utils.MONETARY)},
				ThresholdValue: 2, ActionsID: "TEST_ACTIONS", Executed: true},
			&ActionTrigger{Balance: &BalanceFilter{Type: utils.StringPointer(utils.MONETARY)},
				ThresholdValue: 2, ActionsID: "TEST_ACTIONS", Executed: true}},
	}
	resetTriggersAction(ub, &Action{Balance: &BalanceFilter{Type: utils.StringPointer(utils.SMS)}}, nil, nil)
	if ub.ActionTriggers[0].Executed == false || ub.ActionTriggers[1].Executed == false {
		t.Error("Reset triggers action failed!")
	}
}

func TestActionSetPostpaid(t *testing.T) {
	ub := &Account{
		ID: "TEST_UB",
		BalanceMap: map[string]Balances{
			utils.MONETARY: Balances{&Balance{Value: 100}},
			utils.VOICE: Balances{
				&Balance{Value: 10, Weight: 20, DestinationIDs: utils.NewStringMap("NAT")},
				&Balance{Weight: 10, DestinationIDs: utils.NewStringMap("RET")}}},
		UnitCounters: UnitCounters{utils.MONETARY: []*UnitCounter{
			&UnitCounter{Counters: CounterFilters{&CounterFilter{Value: 1}}}}},
		ActionTriggers: ActionTriggers{
			&ActionTrigger{Balance: &BalanceFilter{Type: utils.StringPointer(utils.MONETARY)},
				ThresholdValue: 2, ActionsID: "TEST_ACTIONS", Executed: true},
			&ActionTrigger{Balance: &BalanceFilter{Type: utils.StringPointer(utils.MONETARY)},
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
			utils.MONETARY: Balances{&Balance{Value: 100}},
			utils.VOICE: Balances{
				&Balance{Value: 10, Weight: 20, DestinationIDs: utils.NewStringMap("NAT")},
				&Balance{Weight: 10, DestinationIDs: utils.NewStringMap("RET")}}},
		UnitCounters: UnitCounters{utils.MONETARY: []*UnitCounter{
			&UnitCounter{Counters: CounterFilters{&CounterFilter{Value: 1}}}}},
		ActionTriggers: ActionTriggers{
			&ActionTrigger{Balance: &BalanceFilter{Type: utils.StringPointer(utils.MONETARY)},
				ThresholdValue: 2, ActionsID: "TEST_ACTIONS", Executed: true},
			&ActionTrigger{Balance: &BalanceFilter{Type: utils.StringPointer(utils.MONETARY)},
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
			utils.MONETARY: Balances{&Balance{Value: 100}},
			utils.VOICE: Balances{
				&Balance{Value: 10, Weight: 20, DestinationIDs: utils.NewStringMap("NAT")},
				&Balance{Weight: 10, DestinationIDs: utils.NewStringMap("RET")}}},
		UnitCounters: UnitCounters{
			utils.MONETARY: []*UnitCounter{&UnitCounter{Counters: CounterFilters{&CounterFilter{Value: 1}}}}},
		ActionTriggers: ActionTriggers{
			&ActionTrigger{Balance: &BalanceFilter{Type: utils.StringPointer(utils.SMS)},
				ThresholdValue: 2, ActionsID: "TEST_ACTIONS", Executed: true},
			&ActionTrigger{Balance: &BalanceFilter{Type: utils.StringPointer(utils.SMS)},
				ThresholdValue: 2, ActionsID: "TEST_ACTIONS", Executed: true}},
	}
	resetAccountAction(ub, nil, nil, nil)
	if !ub.AllowNegative ||
		ub.BalanceMap[utils.MONETARY].GetTotalValue() != 0 ||
		len(ub.UnitCounters) != 0 ||
		ub.BalanceMap[utils.VOICE][0].GetValue() != 0 ||
		ub.ActionTriggers[0].Executed == true || ub.ActionTriggers[1].Executed == true {
		t.Log(ub.BalanceMap)
		t.Error("Reset account action failed!")
	}
}

func TestActionResetPostpaid(t *testing.T) {
	ub := &Account{
		ID: "TEST_UB",
		BalanceMap: map[string]Balances{
			utils.MONETARY: Balances{&Balance{Value: 100}},
			utils.VOICE: Balances{
				&Balance{Value: 10, Weight: 20, DestinationIDs: utils.NewStringMap("NAT")},
				&Balance{Weight: 10, DestinationIDs: utils.NewStringMap("RET")}}},
		UnitCounters: UnitCounters{
			utils.MONETARY: []*UnitCounter{&UnitCounter{Counters: CounterFilters{&CounterFilter{Value: 1}}}}},
		ActionTriggers: ActionTriggers{
			&ActionTrigger{Balance: &BalanceFilter{Type: utils.StringPointer(utils.SMS)},
				ThresholdValue: 2, ActionsID: "TEST_ACTIONS", Executed: true},
			&ActionTrigger{Balance: &BalanceFilter{Type: utils.StringPointer(utils.SMS)},
				ThresholdValue: 2, ActionsID: "TEST_ACTIONS", Executed: true}},
	}
	resetAccountAction(ub, nil, nil, nil)
	if ub.BalanceMap[utils.MONETARY].GetTotalValue() != 0 ||
		len(ub.UnitCounters) != 0 ||
		ub.BalanceMap[utils.VOICE][0].GetValue() != 0 ||
		ub.ActionTriggers[0].Executed == true || ub.ActionTriggers[1].Executed == true {
		t.Error("Reset account action failed!")
	}
}

func TestActionTopupResetCredit(t *testing.T) {
	ub := &Account{
		ID: "TEST_UB",
		BalanceMap: map[string]Balances{
			utils.MONETARY: Balances{&Balance{Directions: utils.NewStringMap(utils.OUT), Value: 100}},
			utils.VOICE: Balances{
				&Balance{Value: 10, Weight: 20, DestinationIDs: utils.NewStringMap("NAT")},
				&Balance{Weight: 10, DestinationIDs: utils.NewStringMap("RET")}}},
		UnitCounters: UnitCounters{
			utils.MONETARY: []*UnitCounter{
				&UnitCounter{Counters: CounterFilters{&CounterFilter{Value: 1,
					Filter: &BalanceFilter{Directions: utils.StringMapPointer(utils.NewStringMap(utils.OUT))}}}}}},
		ActionTriggers: ActionTriggers{
			&ActionTrigger{Balance: &BalanceFilter{Type: utils.StringPointer(utils.MONETARY),
				Directions: utils.StringMapPointer(utils.NewStringMap(utils.OUT))},
				ThresholdValue: 2, ActionsID: "TEST_ACTIONS", Executed: true},
			&ActionTrigger{Balance: &BalanceFilter{Type: utils.StringPointer(utils.MONETARY),
				Directions: utils.StringMapPointer(utils.NewStringMap(utils.OUT))},
				ThresholdValue: 2, ActionsID: "TEST_ACTIONS", Executed: true}},
	}
	a := &Action{Balance: &BalanceFilter{Type: utils.StringPointer(utils.MONETARY),
		Value: &utils.ValueFormula{Static: 10}, Directions: utils.StringMapPointer(utils.NewStringMap(utils.OUT))}}
	topupResetAction(ub, a, nil, nil)
	if ub.AllowNegative ||
		ub.BalanceMap[utils.MONETARY].GetTotalValue() != 10 ||
		len(ub.UnitCounters) != 0 || // InitCounters finds no counters
		len(ub.BalanceMap[utils.VOICE]) != 2 ||
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
			Type:       utils.StringPointer(utils.MONETARY),
			Value:      &utils.ValueFormula{Static: 10},
			Directions: utils.StringMapPointer(utils.NewStringMap(utils.OUT)),
		},
		ExtraParameters: `{"*monetary":2.0}`,
	}
	topupResetAction(ub, a, nil, nil)
	if len(ub.BalanceMap) != 1 || ub.BalanceMap[utils.MONETARY][0].Factor[utils.MONETARY] != 2.0 {
		t.Errorf("Topup reset action failed to set Factor: %+v", ub.BalanceMap[utils.MONETARY][0].Factor)
	}
}

func TestActionTopupResetCreditId(t *testing.T) {
	ub := &Account{
		ID: "TEST_UB",
		BalanceMap: map[string]Balances{
			utils.MONETARY: Balances{
				&Balance{Value: 100},
				&Balance{ID: "TEST_B", Value: 15},
			},
		},
	}
	a := &Action{Balance: &BalanceFilter{Type: utils.StringPointer(utils.MONETARY), ID: utils.StringPointer("TEST_B"),
		Value: &utils.ValueFormula{Static: 10}, Directions: utils.StringMapPointer(utils.NewStringMap(utils.OUT))}}
	topupResetAction(ub, a, nil, nil)
	if ub.AllowNegative ||
		ub.BalanceMap[utils.MONETARY].GetTotalValue() != 110 ||
		len(ub.BalanceMap[utils.MONETARY]) != 2 {
		t.Errorf("Topup reset action failed: %+v", ub.BalanceMap[utils.MONETARY][0])
	}
}

func TestActionTopupResetCreditNoId(t *testing.T) {
	ub := &Account{
		ID: "TEST_UB",
		BalanceMap: map[string]Balances{
			utils.MONETARY: Balances{
				&Balance{Value: 100, Directions: utils.NewStringMap(utils.OUT)},
				&Balance{ID: "TEST_B", Value: 15, Directions: utils.NewStringMap(utils.OUT)},
			},
		},
	}
	a := &Action{Balance: &BalanceFilter{Type: utils.StringPointer(utils.MONETARY),
		Value: &utils.ValueFormula{Static: 10}, Directions: utils.StringMapPointer(utils.NewStringMap(utils.OUT))}}
	topupResetAction(ub, a, nil, nil)
	if ub.AllowNegative ||
		ub.BalanceMap[utils.MONETARY].GetTotalValue() != 20 ||
		len(ub.BalanceMap[utils.MONETARY]) != 2 {
		t.Errorf("Topup reset action failed: %+v", ub.BalanceMap[utils.MONETARY][1])
	}
}

func TestActionTopupResetMinutes(t *testing.T) {
	ub := &Account{
		ID: "TEST_UB",
		BalanceMap: map[string]Balances{
			utils.MONETARY: Balances{&Balance{Value: 100}},
			utils.VOICE: Balances{&Balance{Value: 10, Weight: 20, DestinationIDs: utils.NewStringMap("NAT"),
				Directions: utils.NewStringMap(utils.OUT)}, &Balance{Weight: 10, DestinationIDs: utils.NewStringMap("RET")}}},
		UnitCounters: UnitCounters{utils.MONETARY: []*UnitCounter{&UnitCounter{Counters: CounterFilters{&CounterFilter{Value: 1,
			Filter: &BalanceFilter{Directions: utils.StringMapPointer(utils.NewStringMap(utils.OUT))}}}}}},
		ActionTriggers: ActionTriggers{
			&ActionTrigger{Balance: &BalanceFilter{Type: utils.StringPointer(utils.MONETARY),
				Directions: utils.StringMapPointer(utils.NewStringMap(utils.OUT))},
				ThresholdValue: 2, ActionsID: "TEST_ACTIONS", Executed: true},
			&ActionTrigger{Balance: &BalanceFilter{Type: utils.StringPointer(utils.MONETARY),
				Directions: utils.StringMapPointer(utils.NewStringMap(utils.OUT))},
				ThresholdValue: 2, ActionsID: "TEST_ACTIONS", Executed: true}},
	}
	a := &Action{Balance: &BalanceFilter{Type: utils.StringPointer(utils.VOICE),
		Value: &utils.ValueFormula{Static: 5}, Weight: utils.Float64Pointer(20),
		DestinationIDs: utils.StringMapPointer(utils.NewStringMap("NAT")),
		Directions:     utils.StringMapPointer(utils.NewStringMap(utils.OUT))}}
	topupResetAction(ub, a, nil, nil)
	if ub.AllowNegative ||
		ub.BalanceMap[utils.VOICE].GetTotalValue() != 5 ||
		ub.BalanceMap[utils.MONETARY].GetTotalValue() != 100 ||
		len(ub.UnitCounters) != 0 ||
		len(ub.BalanceMap[utils.VOICE]) != 2 ||
		ub.ActionTriggers[0].Executed != true || ub.ActionTriggers[1].Executed != true {
		t.Errorf("Topup reset minutes action failed: %+v", ub.BalanceMap[utils.VOICE][0])
	}
}

func TestActionTopupCredit(t *testing.T) {
	ub := &Account{
		ID: "TEST_UB",
		BalanceMap: map[string]Balances{
			utils.MONETARY: Balances{&Balance{Value: 100}},
			utils.VOICE: Balances{&Balance{Value: 10, Weight: 20, DestinationIDs: utils.NewStringMap("NAT"),
				Directions: utils.NewStringMap(utils.OUT)}, &Balance{Weight: 10,
				DestinationIDs: utils.NewStringMap("RET")}}},
		UnitCounters: UnitCounters{
			utils.MONETARY: []*UnitCounter{
				&UnitCounter{Counters: CounterFilters{&CounterFilter{Value: 1,
					Filter: &BalanceFilter{Directions: utils.StringMapPointer(utils.NewStringMap(utils.OUT))}}}}}},
		ActionTriggers: ActionTriggers{
			&ActionTrigger{Balance: &BalanceFilter{Type: utils.StringPointer(utils.MONETARY),
				Directions: utils.StringMapPointer(utils.NewStringMap(utils.OUT))},
				ThresholdValue: 2, ActionsID: "TEST_ACTIONS", Executed: true},
			&ActionTrigger{Balance: &BalanceFilter{Type: utils.StringPointer(utils.MONETARY),
				Directions: utils.StringMapPointer(utils.NewStringMap(utils.OUT))},
				ThresholdValue: 2, ActionsID: "TEST_ACTIONS", Executed: true}},
	}
	a := &Action{Balance: &BalanceFilter{Type: utils.StringPointer(utils.MONETARY),
		Value:      &utils.ValueFormula{Static: 10},
		Directions: utils.StringMapPointer(utils.NewStringMap(utils.OUT))}}
	topupAction(ub, a, nil, nil)
	if ub.AllowNegative ||
		ub.BalanceMap[utils.MONETARY].GetTotalValue() != 110 ||
		len(ub.UnitCounters) != 0 ||
		len(ub.BalanceMap[utils.VOICE]) != 2 ||
		ub.ActionTriggers[0].Executed != true || ub.ActionTriggers[1].Executed != true {
		t.Error("Topup action failed!", ub.BalanceMap[utils.MONETARY].GetTotalValue())
	}
}

func TestActionTopupMinutes(t *testing.T) {
	ub := &Account{
		ID: "TEST_UB",
		BalanceMap: map[string]Balances{
			utils.MONETARY: Balances{&Balance{Value: 100}},
			utils.VOICE: Balances{&Balance{Value: 10, Weight: 20, DestinationIDs: utils.NewStringMap("NAT"),
				Directions: utils.NewStringMap(utils.OUT)}, &Balance{Weight: 10, DestinationIDs: utils.NewStringMap("RET")}}},
		UnitCounters: UnitCounters{
			utils.MONETARY: []*UnitCounter{&UnitCounter{Counters: CounterFilters{&CounterFilter{Value: 1}}}}},
		ActionTriggers: ActionTriggers{
			&ActionTrigger{Balance: &BalanceFilter{Type: utils.StringPointer(utils.MONETARY)},
				ThresholdValue: 2, ActionsID: "TEST_ACTIONS", Executed: true},
			&ActionTrigger{Balance: &BalanceFilter{Type: utils.StringPointer(utils.MONETARY)},
				ThresholdValue: 2, ActionsID: "TEST_ACTIONS", Executed: true}},
	}
	a := &Action{Balance: &BalanceFilter{Type: utils.StringPointer(utils.VOICE),
		Value: &utils.ValueFormula{Static: 5}, Weight: utils.Float64Pointer(20),
		DestinationIDs: utils.StringMapPointer(utils.NewStringMap("NAT")),
		Directions:     utils.StringMapPointer(utils.NewStringMap(utils.OUT))}}
	topupAction(ub, a, nil, nil)
	if ub.AllowNegative ||
		ub.BalanceMap[utils.VOICE].GetTotalValue() != 15 ||
		ub.BalanceMap[utils.MONETARY].GetTotalValue() != 100 ||
		len(ub.UnitCounters) != 0 ||
		len(ub.BalanceMap[utils.VOICE]) != 2 ||
		ub.ActionTriggers[0].Executed != true || ub.ActionTriggers[1].Executed != true {
		t.Error("Topup minutes action failed!", ub.BalanceMap[utils.VOICE])
	}
}

func TestActionDebitCredit(t *testing.T) {
	ub := &Account{
		ID: "TEST_UB",
		BalanceMap: map[string]Balances{
			utils.MONETARY: Balances{&Balance{Value: 100}},
			utils.VOICE: Balances{
				&Balance{Value: 10, Weight: 20, DestinationIDs: utils.NewStringMap("NAT")},
				&Balance{Weight: 10, DestinationIDs: utils.NewStringMap("RET")}}},
		UnitCounters: UnitCounters{
			utils.MONETARY: []*UnitCounter{&UnitCounter{Counters: CounterFilters{&CounterFilter{Value: 1,
				Filter: &BalanceFilter{Directions: utils.StringMapPointer(utils.NewStringMap(utils.OUT))}}}}}},
		ActionTriggers: ActionTriggers{
			&ActionTrigger{Balance: &BalanceFilter{Type: utils.StringPointer(utils.MONETARY),
				Directions: utils.StringMapPointer(utils.NewStringMap(utils.OUT))},
				ThresholdValue: 2, ActionsID: "TEST_ACTIONS", Executed: true},
			&ActionTrigger{Balance: &BalanceFilter{Type: utils.StringPointer(utils.MONETARY),
				Directions: utils.StringMapPointer(utils.NewStringMap(utils.OUT))},
				ThresholdValue: 2, ActionsID: "TEST_ACTIONS", Executed: true}},
	}
	a := &Action{Balance: &BalanceFilter{Type: utils.StringPointer(utils.MONETARY),
		Value:      &utils.ValueFormula{Static: 10},
		Directions: utils.StringMapPointer(utils.NewStringMap(utils.OUT))}}
	debitAction(ub, a, nil, nil)
	if ub.AllowNegative ||
		ub.BalanceMap[utils.MONETARY].GetTotalValue() != 90 ||
		len(ub.UnitCounters) != 0 ||
		len(ub.BalanceMap[utils.VOICE]) != 2 ||
		ub.ActionTriggers[0].Executed != true || ub.ActionTriggers[1].Executed != true {
		t.Error("Debit action failed!", utils.ToIJSON(ub))
	}
}

func TestActionDebitMinutes(t *testing.T) {
	ub := &Account{
		ID: "TEST_UB",
		BalanceMap: map[string]Balances{
			utils.MONETARY: Balances{&Balance{Value: 100}},
			utils.VOICE: Balances{
				&Balance{Value: 10, Weight: 20,
					DestinationIDs: utils.NewStringMap("NAT"), Directions: utils.NewStringMap(utils.OUT)},
				&Balance{Weight: 10, DestinationIDs: utils.NewStringMap("RET")}}},
		UnitCounters: UnitCounters{
			utils.MONETARY: []*UnitCounter{&UnitCounter{Counters: CounterFilters{&CounterFilter{Value: 1}}}}},
		ActionTriggers: ActionTriggers{
			&ActionTrigger{Balance: &BalanceFilter{Type: utils.StringPointer(utils.MONETARY)},
				ThresholdValue: 2, ActionsID: "TEST_ACTIONS", Executed: true},
			&ActionTrigger{Balance: &BalanceFilter{Type: utils.StringPointer(utils.MONETARY)},
				ThresholdValue: 2, ActionsID: "TEST_ACTIONS", Executed: true}},
	}
	a := &Action{Balance: &BalanceFilter{Type: utils.StringPointer(utils.VOICE),
		Value: &utils.ValueFormula{Static: 5}, Weight: utils.Float64Pointer(20),
		DestinationIDs: utils.StringMapPointer(utils.NewStringMap("NAT")),
		Directions:     utils.StringMapPointer(utils.NewStringMap(utils.OUT))}}
	debitAction(ub, a, nil, nil)
	if ub.AllowNegative ||
		ub.BalanceMap[utils.VOICE][0].GetValue() != 5 ||
		ub.BalanceMap[utils.MONETARY].GetTotalValue() != 100 ||
		len(ub.UnitCounters) != 0 ||
		len(ub.BalanceMap[utils.VOICE]) != 2 ||
		ub.ActionTriggers[0].Executed != true || ub.ActionTriggers[1].Executed != true {
		t.Error("Debit minutes action failed!", ub.BalanceMap[utils.VOICE][0])
	}
}

func TestActionResetAllCounters(t *testing.T) {
	ub := &Account{
		ID:            "TEST_UB",
		AllowNegative: true,
		BalanceMap: map[string]Balances{
			utils.MONETARY: Balances{&Balance{Value: 100}},
			utils.VOICE: Balances{
				&Balance{Value: 10, Weight: 20, DestinationIDs: utils.NewStringMap("NAT"),
					Directions: utils.NewStringMap(utils.OUT)},
				&Balance{Weight: 10, DestinationIDs: utils.NewStringMap("RET"),
					Directions: utils.NewStringMap(utils.OUT)}}},

		ActionTriggers: ActionTriggers{
			&ActionTrigger{ThresholdType: utils.TRIGGER_MAX_EVENT_COUNTER, ThresholdValue: 2,
				Balance: &BalanceFilter{Type: utils.StringPointer(utils.MONETARY),
					DestinationIDs: utils.StringMapPointer(utils.NewStringMap("NAT")),
					Weight:         utils.Float64Pointer(20)}, ActionsID: "TEST_ACTIONS", Executed: true}},
	}
	ub.InitCounters()
	resetCountersAction(ub, nil, nil, nil)
	if !ub.AllowNegative ||
		ub.BalanceMap[utils.MONETARY].GetTotalValue() != 100 ||
		len(ub.UnitCounters) != 1 ||
		len(ub.UnitCounters[utils.MONETARY][0].Counters) != 1 ||
		len(ub.BalanceMap[utils.MONETARY]) != 1 ||
		ub.ActionTriggers[0].Executed != true {
		t.Errorf("Reset counters action failed: %+v %+v %+v", ub.UnitCounters,
			ub.UnitCounters[utils.MONETARY][0], ub.UnitCounters[utils.MONETARY][0].Counters[0])
	}
	if len(ub.UnitCounters) < 1 {
		t.FailNow()
	}
	c := ub.UnitCounters[utils.MONETARY][0].Counters[0]
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
			utils.MONETARY: Balances{&Balance{Value: 100}},
			utils.VOICE: Balances{&Balance{Value: 10, Weight: 20,
				DestinationIDs: utils.NewStringMap("NAT")}, &Balance{Weight: 10,
				DestinationIDs: utils.NewStringMap("RET")}}},
		ActionTriggers: ActionTriggers{
			&ActionTrigger{Balance: &BalanceFilter{Type: utils.StringPointer(utils.MONETARY)},
				ThresholdType: utils.TRIGGER_MAX_EVENT_COUNTER, ThresholdValue: 2,
				ActionsID: "TEST_ACTIONS", Executed: true}},
	}
	a := &Action{Balance: &BalanceFilter{Type: utils.StringPointer(utils.MONETARY)}}
	ub.InitCounters()
	resetCountersAction(ub, a, nil, nil)
	if !ub.AllowNegative ||
		ub.BalanceMap[utils.MONETARY].GetTotalValue() != 100 ||
		len(ub.UnitCounters) != 1 ||
		len(ub.UnitCounters[utils.MONETARY][0].Counters) != 1 ||
		len(ub.BalanceMap[utils.VOICE]) != 2 ||
		ub.ActionTriggers[0].Executed != true {
		for _, b := range ub.UnitCounters[utils.MONETARY][0].Counters {
			t.Logf("B: %+v", b)
		}
		t.Errorf("Reset counters action failed: %+v", ub.UnitCounters)
	}
	if len(ub.UnitCounters) < 1 || len(ub.UnitCounters[utils.MONETARY][0].Counters) < 1 {
		t.FailNow()
	}
	c := ub.UnitCounters[utils.MONETARY][0].Counters[0]
	if c.Filter.GetWeight() != 0 || c.Value != 0 || len(c.Filter.GetDestinationIDs()) != 0 {
		t.Errorf("Balance cloned incorrectly: %+v!", c)
	}
}

func TestActionResetCounterCredit(t *testing.T) {
	ub := &Account{
		ID:            "TEST_UB",
		AllowNegative: true,
		BalanceMap: map[string]Balances{
			utils.MONETARY: Balances{&Balance{Value: 100}},
			utils.VOICE: Balances{&Balance{Value: 10, Weight: 20, DestinationIDs: utils.NewStringMap("NAT")},
				&Balance{Weight: 10, DestinationIDs: utils.NewStringMap("RET")}}},
		UnitCounters: UnitCounters{
			utils.MONETARY: []*UnitCounter{&UnitCounter{Counters: CounterFilters{
				&CounterFilter{Value: 1, Filter: &BalanceFilter{
					Directions: utils.StringMapPointer(utils.NewStringMap(utils.OUT))}}}}},
			utils.SMS: []*UnitCounter{&UnitCounter{
				Counters: CounterFilters{&CounterFilter{Value: 1,
					Filter: &BalanceFilter{Directions: utils.StringMapPointer(utils.NewStringMap(utils.OUT))}}}}}},
		ActionTriggers: ActionTriggers{
			&ActionTrigger{Balance: &BalanceFilter{Type: utils.StringPointer(utils.MONETARY),
				Directions: utils.StringMapPointer(utils.NewStringMap(utils.OUT))},
				ThresholdValue: 2, ActionsID: "TEST_ACTIONS", Executed: true}},
	}
	a := &Action{Balance: &BalanceFilter{Type: utils.StringPointer(utils.MONETARY)}}
	resetCountersAction(ub, a, nil, nil)
	if !ub.AllowNegative ||
		ub.BalanceMap[utils.MONETARY].GetTotalValue() != 100 ||
		len(ub.UnitCounters) != 2 ||
		len(ub.BalanceMap[utils.VOICE]) != 2 ||
		ub.ActionTriggers[0].Executed != true {
		t.Error("Reset counters action failed!", ub.UnitCounters)
	}
}

func TestActionTriggerLogging(t *testing.T) {
	at := &ActionTrigger{
		ID: "some_uuid",
		Balance: &BalanceFilter{
			Type:           utils.StringPointer(utils.MONETARY),
			Directions:     utils.StringMapPointer(utils.NewStringMap(utils.OUT)),
			DestinationIDs: utils.StringMapPointer(utils.NewStringMap("NAT")),
		},
		ThresholdValue: 100.0,
		Weight:         10.0,
		ActionsID:      "TEST_ACTIONS",
	}
	as, err := dm.GetActions(at.ActionsID, false, utils.NonTransactional)
	if err != nil {
		t.Error("Error getting actions for the action timing: ", as, err)
	}
	Publish(CgrEvent{
		"EventName": utils.EVT_ACTION_TRIGGER_FIRED,
		"Uuid":      at.UniqueID,
		"Id":        at.ID,
		"ActionIds": at.ActionsID,
	})
	//expected := "rif*some_uuid;MONETARY;OUT;NAT;TEST_ACTIONS;100;10;false*|TOPUP|MONETARY|OUT|10|0"
	var key string
	atMap, _ := dm.DataDB().GetAllActionPlans()
	for k, v := range atMap {
		_ = k
		_ = v
		/*if strings.Contains(k, LOG_ACTION_utils.TRIGGER_PREFIX) && strings.Contains(v, expected) {
		    key = k
		    break
		}*/
	}
	if key != "" {
		t.Error("Action timing was not logged")
	}
}

func TestActionPlanLogging(t *testing.T) {
	i := &RateInterval{
		Timing: &RITiming{
			Months: utils.Months{time.January, time.February, time.March, time.April, time.May, time.June,
				time.July, time.August, time.September, time.October, time.November, time.December},
			MonthDays: utils.MonthDays{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
				16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31},
			WeekDays:  utils.WeekDays{time.Monday, time.Tuesday, time.Wednesday, time.Thursday, time.Friday},
			StartTime: "18:00:00",
			EndTime:   "00:00:00",
		},
		Weight: 10.0,
		Rating: &RIRate{
			ConnectFee: 0.0,
			Rates:      RateGroups{&Rate{0, 1.0, 1 * time.Second, 60 * time.Second}},
		},
	}
	at := &ActionTiming{
		accountIDs: utils.StringMap{"one": true, "two": true, "three": true},
		Timing:     i,
		Weight:     10.0,
		ActionsID:  "TEST_ACTIONS",
	}
	if err != nil {
		t.Error("Error getting actions for the action trigger: ", err)
	}
	Publish(CgrEvent{
		"EventName": utils.EVT_ACTION_TIMING_FIRED,
		"Uuid":      at.Uuid,
		"Id":        at.actionPlanID,
		"ActionIds": at.ActionsID,
	})
	//expected := "some uuid|test|one,two,three|;1,2,3,4,5,6,7,8,9,10,11,12;1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31;1,2,3,4,5;18:00:00;00:00:00;10;0;1;60;1|10|TEST_ACTIONS*|TOPUP|MONETARY|OUT|10|0"
	var key string
	atMap, _ := dm.DataDB().GetAllActionPlans()
	for k, v := range atMap {
		_ = k
		_ = v
		/*if strings.Contains(k, LOG_ACTION_TIMMING_PREFIX) && strings.Contains(string(v), expected) {
		    key = k
		}*/
	}
	if key != "" {
		t.Error("Action trigger was not logged")
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

func TestRemoveAction(t *testing.T) {
	if _, err := dm.DataDB().GetAccount("cgrates.org:remo"); err != nil {
		t.Errorf("account to be removed not found: %v", err)
	}
	a := &Action{
		ActionType: REMOVE_ACCOUNT,
	}

	at := &ActionTiming{
		accountIDs: utils.StringMap{"cgrates.org:remo": true},
		actions:    Actions{a},
	}
	at.Execute(nil, nil)
	afterUb, err := dm.DataDB().GetAccount("cgrates.org:remo")
	if err == nil || afterUb != nil {
		t.Error("error removing account: ", err, afterUb)
	}
}

func TestTopupAction(t *testing.T) {
	initialUb, _ := dm.DataDB().GetAccount("vdf:minu")
	a := &Action{
		ActionType: TOPUP,
		Balance: &BalanceFilter{Type: utils.StringPointer(utils.MONETARY), Value: &utils.ValueFormula{Static: 25},
			DestinationIDs: utils.StringMapPointer(utils.NewStringMap("RET")),
			Directions:     utils.StringMapPointer(utils.NewStringMap(utils.OUT)), Weight: utils.Float64Pointer(20)},
	}

	at := &ActionTiming{
		accountIDs: utils.StringMap{"vdf:minu": true},
		actions:    Actions{a},
	}

	at.Execute(nil, nil)
	afterUb, _ := dm.DataDB().GetAccount("vdf:minu")
	initialValue := initialUb.BalanceMap[utils.MONETARY].GetTotalValue()
	afterValue := afterUb.BalanceMap[utils.MONETARY].GetTotalValue()
	if afterValue != initialValue+25 {
		t.Error("Bad topup before and after: ", initialValue, afterValue)
	}
}

func TestTopupActionLoaded(t *testing.T) {
	initialUb, _ := dm.DataDB().GetAccount("vdf:minitsboy")
	a := &Action{
		ActionType: TOPUP,
		Balance: &BalanceFilter{Type: utils.StringPointer(utils.MONETARY),
			Value: &utils.ValueFormula{Static: 25}, DestinationIDs: utils.StringMapPointer(utils.NewStringMap("RET")),
			Directions: utils.StringMapPointer(utils.NewStringMap(utils.OUT)), Weight: utils.Float64Pointer(20)},
	}

	at := &ActionTiming{
		accountIDs: utils.StringMap{"vdf:minitsboy": true},
		actions:    Actions{a},
	}

	at.Execute(nil, nil)
	afterUb, _ := dm.DataDB().GetAccount("vdf:minitsboy")
	initialValue := initialUb.BalanceMap[utils.MONETARY].GetTotalValue()
	afterValue := afterUb.BalanceMap[utils.MONETARY].GetTotalValue()
	if afterValue != initialValue+25 {
		t.Logf("Initial: %+v", initialUb)
		t.Logf("After: %+v", afterUb)
		t.Error("Bad topup before and after: ", initialValue, afterValue)
	}
}

/*
Need to be reviewed with extra data instead of cdrstats
func TestActionSetDDestination(t *testing.T) {
	acc := &Account{BalanceMap: map[string]Balances{
		utils.MONETARY: Balances{&Balance{DestinationIDs: utils.NewStringMap("*ddc_test")}}}}
	origD := &Destination{Id: "*ddc_test", Prefixes: []string{"111", "222"}}
	dm.DataDB().SetDestination(origD, utils.NonTransactional)
	dm.DataDB().SetReverseDestination(origD, utils.NonTransactional)
	// check redis and cache
	if d, err := dm.DataDB().GetDestination("*ddc_test", false, utils.NonTransactional); err != nil || !reflect.DeepEqual(d, origD) {
		t.Error("Error storing destination: ", d, err)
	}
	dm.DataDB().GetReverseDestination("111", false, utils.NonTransactional)
	x1, found := Cache.Get(utils.CacheReverseDestinations, "111")
	if !found || len(x1.([]string)) != 1 {
		t.Error("Error cacheing destination: ", x1)
	}
	dm.DataDB().GetReverseDestination("222", false, utils.NonTransactional)
	x1, found = Cache.Get(utils.CacheReverseDestinations, "222")
	if !found || len(x1.([]string)) != 1 {
		t.Error("Error cacheing destination: ", x1)
	}
	setddestinations(acc, &CDRStatsQueueTriggered{Metrics: map[string]float64{"333": 1, "666": 1}}, nil, nil)
	d, err := dm.DataDB().GetDestination("*ddc_test", false, utils.NonTransactional)
	if err != nil ||
		d.Id != origD.Id ||
		len(d.Prefixes) != 2 ||
		!utils.IsSliceMember(d.Prefixes, "333") ||
		!utils.IsSliceMember(d.Prefixes, "666") {
		t.Error("Error storing destination: ", d, err)
	}

	var ok bool
	x1, ok = Cache.Get(utils.CacheReverseDestinations, "111")
	if ok {
		t.Error("Error cacheing destination: ", x1)
	}
	x1, ok = Cache.Get(utils.CacheReverseDestinations, "222")
	if ok {
		t.Error("Error cacheing destination: ", x1)
	}
	dm.DataDB().GetReverseDestination("333", false, utils.NonTransactional)
	x1, found = Cache.Get(utils.CacheReverseDestinations, "333")
	if !found || len(x1.([]string)) != 1 {
		t.Error("Error cacheing destination: ", x1)
	}
	dm.DataDB().GetReverseDestination("666", false, utils.NonTransactional)
	x1, found = Cache.Get(utils.CacheReverseDestinations, "666")
	if !found || len(x1.([]string)) != 1 {
		t.Error("Error cacheing destination: ", x1)
	}
}
*/

func TestActionTransactionFuncType(t *testing.T) {
	err := dm.DataDB().SetAccount(&Account{
		ID: "cgrates.org:trans",
		BalanceMap: map[string]Balances{
			utils.MONETARY: Balances{&Balance{
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
			&Action{
				ActionType: TOPUP,
				Balance: &BalanceFilter{Value: &utils.ValueFormula{Static: 1.1},
					Type: utils.StringPointer(utils.MONETARY)},
			},
			&Action{
				ActionType: "VALID_FUNCTION_TYPE",
				Balance: &BalanceFilter{Value: &utils.ValueFormula{Static: 1.1},
					Type: utils.StringPointer("test")},
			},
		},
	}
	err = at.Execute(nil, nil)
	acc, err := dm.DataDB().GetAccount("cgrates.org:trans")
	if err != nil || acc == nil {
		t.Error("Error getting account: ", acc, err)
	}
	if acc.BalanceMap[utils.MONETARY][0].Value != 10 {
		t.Errorf("Transaction didn't work: %v", acc.BalanceMap[utils.MONETARY][0].Value)
	}
}

func TestActionTransactionBalanceType(t *testing.T) {
	err := dm.DataDB().SetAccount(&Account{
		ID: "cgrates.org:trans",
		BalanceMap: map[string]Balances{
			utils.MONETARY: Balances{&Balance{
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
			&Action{
				ActionType: TOPUP,
				Balance: &BalanceFilter{Value: &utils.ValueFormula{Static: 1.1},
					Type: utils.StringPointer(utils.MONETARY)},
			},
			&Action{
				ActionType: TOPUP,
				Balance:    &BalanceFilter{Type: utils.StringPointer("test")},
			},
		},
	}
	err = at.Execute(nil, nil)
	acc, err := dm.DataDB().GetAccount("cgrates.org:trans")
	if err != nil || acc == nil {
		t.Error("Error getting account: ", acc, err)
	}
	if acc.BalanceMap[utils.MONETARY][0].Value != 11.1 {
		t.Errorf("Transaction didn't work: %v", acc.BalanceMap[utils.MONETARY][0].Value)
	}
}

func TestActionTransactionBalanceNotType(t *testing.T) {
	err := dm.DataDB().SetAccount(&Account{
		ID: "cgrates.org:trans",
		BalanceMap: map[string]Balances{
			utils.MONETARY: Balances{&Balance{
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
			&Action{
				ActionType: TOPUP,
				Balance: &BalanceFilter{Value: &utils.ValueFormula{Static: 1.1},
					Type: utils.StringPointer(utils.VOICE)},
			},
			&Action{
				ActionType: TOPUP,
				Balance:    &BalanceFilter{Type: utils.StringPointer("test")},
			},
		},
	}
	err = at.Execute(nil, nil)
	acc, err := dm.DataDB().GetAccount("cgrates.org:trans")
	if err != nil || acc == nil {
		t.Error("Error getting account: ", acc, err)
	}
	if acc.BalanceMap[utils.MONETARY][0].Value != 10.0 {
		t.Errorf("Transaction didn't work: %v", acc.BalanceMap[utils.MONETARY][0].Value)
	}
}

func TestActionWithExpireWithoutExpire(t *testing.T) {
	err := dm.DataDB().SetAccount(&Account{
		ID: "cgrates.org:exp",
		BalanceMap: map[string]Balances{
			utils.MONETARY: Balances{&Balance{
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
			&Action{
				ActionType: TOPUP,
				Balance: &BalanceFilter{
					Type:  utils.StringPointer(utils.VOICE),
					Value: &utils.ValueFormula{Static: 15},
				},
			},
			&Action{
				ActionType: TOPUP,
				Balance: &BalanceFilter{
					Type:           utils.StringPointer(utils.VOICE),
					Value:          &utils.ValueFormula{Static: 30},
					ExpirationDate: utils.TimePointer(time.Date(2025, time.November, 11, 22, 39, 0, 0, time.UTC)),
				},
			},
		},
	}
	err = at.Execute(nil, nil)
	acc, err := dm.DataDB().GetAccount("cgrates.org:exp")
	if err != nil || acc == nil {
		t.Errorf("Error getting account: %+v: %v", acc, err)
	}
	if len(acc.BalanceMap) != 2 ||
		len(acc.BalanceMap[utils.VOICE]) != 2 {
		t.Errorf("Error debiting expir and unexpire: %+v", acc.BalanceMap[utils.VOICE][0])
	}
}

func TestActionRemoveBalance(t *testing.T) {
	err := dm.DataDB().SetAccount(&Account{
		ID: "cgrates.org:rembal",
		BalanceMap: map[string]Balances{
			utils.MONETARY: Balances{
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
			&Action{
				ActionType: REMOVE_BALANCE,
				Balance: &BalanceFilter{
					Type:           utils.StringPointer(utils.MONETARY),
					DestinationIDs: utils.StringMapPointer(utils.NewStringMap("NAT", "RET")),
				},
			},
		},
	}
	err = at.Execute(nil, nil)
	acc, err := dm.DataDB().GetAccount("cgrates.org:rembal")
	if err != nil || acc == nil {
		t.Errorf("Error getting account: %+v: %v", acc, err)
	}
	if len(acc.BalanceMap) != 1 ||
		len(acc.BalanceMap[utils.MONETARY]) != 1 {
		t.Errorf("Error removing balance: %+v", acc.BalanceMap[utils.MONETARY])
	}
}

func TestActionTransferMonetaryDefault(t *testing.T) {
	err := dm.DataDB().SetAccount(
		&Account{
			ID: "cgrates.org:trans",
			BalanceMap: map[string]Balances{
				utils.MONETARY: Balances{
					&Balance{
						Uuid:  utils.GenUUID(),
						ID:    utils.META_DEFAULT,
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
		ActionType: TRANSFER_MONETARY_DEFAULT,
	}

	at := &ActionTiming{
		accountIDs: utils.StringMap{"cgrates.org:trans": true},
		actions:    Actions{a},
	}
	at.Execute(nil, nil)

	afterUb, err := dm.DataDB().GetAccount("cgrates.org:trans")
	if err != nil {
		t.Error("account not found: ", err, afterUb)
	}
	if afterUb.BalanceMap[utils.MONETARY].GetTotalValue() != 17 ||
		afterUb.BalanceMap[utils.MONETARY][0].Value != 19 ||
		afterUb.BalanceMap[utils.MONETARY][1].Value != 0 ||
		afterUb.BalanceMap[utils.MONETARY][2].Value != 0 ||
		afterUb.BalanceMap[utils.MONETARY][3].Value != -2 {
		for _, b := range afterUb.BalanceMap[utils.MONETARY] {
			t.Logf("B: %+v", b)
		}
		t.Error("ransfer balance value: ", afterUb.BalanceMap[utils.MONETARY].GetTotalValue())
	}
}

func TestActionTransferMonetaryDefaultFilter(t *testing.T) {
	err := dm.DataDB().SetAccount(
		&Account{
			ID: "cgrates.org:trans",
			BalanceMap: map[string]Balances{
				utils.MONETARY: Balances{
					&Balance{
						Uuid:   utils.GenUUID(),
						ID:     utils.META_DEFAULT,
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
		ActionType: TRANSFER_MONETARY_DEFAULT,
		Balance:    &BalanceFilter{Weight: utils.Float64Pointer(20)},
	}

	at := &ActionTiming{
		accountIDs: utils.StringMap{"cgrates.org:trans": true},
		actions:    Actions{a},
	}
	at.Execute(nil, nil)

	afterUb, err := dm.DataDB().GetAccount("cgrates.org:trans")
	if err != nil {
		t.Error("account not found: ", err, afterUb)
	}
	if afterUb.BalanceMap[utils.MONETARY].GetTotalValue() != 20 ||
		afterUb.BalanceMap[utils.MONETARY][0].Value != 19 ||
		afterUb.BalanceMap[utils.MONETARY][1].Value != 0 ||
		afterUb.BalanceMap[utils.MONETARY][2].Value != 1 ||
		afterUb.BalanceMap[utils.MONETARY][3].Value != 0 {
		for _, b := range afterUb.BalanceMap[utils.MONETARY] {
			t.Logf("B: %+v", b)
		}
		t.Error("ransfer balance value: ", afterUb.BalanceMap[utils.MONETARY].GetTotalValue())
	}
}

func TestActionConditionalTopup(t *testing.T) {
	err := dm.DataDB().SetAccount(
		&Account{
			ID: "cgrates.org:cond",
			BalanceMap: map[string]Balances{
				utils.MONETARY: Balances{
					&Balance{
						Uuid:   utils.GenUUID(),
						ID:     utils.META_DEFAULT,
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
		ActionType: TOPUP,
		Filter:     `{"Type":"*monetary","Value":1,"Weight":10}`,
		Balance: &BalanceFilter{
			Type:   utils.StringPointer(utils.MONETARY),
			Value:  &utils.ValueFormula{Static: 11},
			Weight: utils.Float64Pointer(30),
		},
	}

	at := &ActionTiming{
		accountIDs: utils.StringMap{"cgrates.org:cond": true},
		actions:    Actions{a},
	}
	at.Execute(nil, nil)

	afterUb, err := dm.DataDB().GetAccount("cgrates.org:cond")
	if err != nil {
		t.Error("account not found: ", err, afterUb)
	}
	if len(afterUb.BalanceMap[utils.MONETARY]) != 5 ||
		afterUb.BalanceMap[utils.MONETARY].GetTotalValue() != 31 ||
		afterUb.BalanceMap[utils.MONETARY][4].Value != 11 {
		for _, b := range afterUb.BalanceMap[utils.MONETARY] {
			t.Logf("B: %+v", b)
		}
		t.Error("ransfer balance value: ", afterUb.BalanceMap[utils.MONETARY].GetTotalValue())
	}
}

func TestActionConditionalTopupNoMatch(t *testing.T) {
	err := dm.DataDB().SetAccount(
		&Account{
			ID: "cgrates.org:cond",
			BalanceMap: map[string]Balances{
				utils.MONETARY: Balances{
					&Balance{
						Uuid:   utils.GenUUID(),
						ID:     utils.META_DEFAULT,
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
		ActionType: TOPUP,
		Filter:     `{"Type":"*monetary","Value":2,"Weight":10}`,
		Balance: &BalanceFilter{
			Type:   utils.StringPointer(utils.MONETARY),
			Value:  &utils.ValueFormula{Static: 11},
			Weight: utils.Float64Pointer(30),
		},
	}

	at := &ActionTiming{
		accountIDs: utils.StringMap{"cgrates.org:cond": true},
		actions:    Actions{a},
	}
	at.Execute(nil, nil)

	afterUb, err := dm.DataDB().GetAccount("cgrates.org:cond")
	if err != nil {
		t.Error("account not found: ", err, afterUb)
	}
	if len(afterUb.BalanceMap[utils.MONETARY]) != 4 ||
		afterUb.BalanceMap[utils.MONETARY].GetTotalValue() != 20 {
		for _, b := range afterUb.BalanceMap[utils.MONETARY] {
			t.Logf("B: %+v", b)
		}
		t.Error("ransfer balance value: ", afterUb.BalanceMap[utils.MONETARY].GetTotalValue())
	}
}

func TestActionConditionalTopupExistingBalance(t *testing.T) {
	err := dm.DataDB().SetAccount(
		&Account{
			ID: "cgrates.org:cond",
			BalanceMap: map[string]Balances{
				utils.MONETARY: Balances{
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
				utils.VOICE: Balances{
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
		ActionType: TOPUP,
		Filter:     `{"Type":"*voice","Value":{"*gte":100}}`,
		Balance: &BalanceFilter{
			Type:   utils.StringPointer(utils.MONETARY),
			Value:  &utils.ValueFormula{Static: 11},
			Weight: utils.Float64Pointer(10),
		},
	}

	at := &ActionTiming{
		accountIDs: utils.StringMap{"cgrates.org:cond": true},
		actions:    Actions{a},
	}
	at.Execute(nil, nil)

	afterUb, err := dm.DataDB().GetAccount("cgrates.org:cond")
	if err != nil {
		t.Error("account not found: ", err, afterUb)
	}
	if len(afterUb.BalanceMap[utils.MONETARY]) != 2 ||
		afterUb.BalanceMap[utils.MONETARY].GetTotalValue() != 18 {
		for _, b := range afterUb.BalanceMap[utils.MONETARY] {
			t.Logf("B: %+v", b)
		}
		t.Error("ransfer balance value: ", afterUb.BalanceMap[utils.MONETARY].GetTotalValue())
	}
}

func TestActionConditionalDisabledIfNegative(t *testing.T) {
	err := dm.DataDB().SetAccount(
		&Account{
			ID: "cgrates.org:af",
			BalanceMap: map[string]Balances{
				"*data": Balances{
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
				"*monetary": Balances{
					&Balance{
						Uuid:  "9fa1847a-f36a-41a7-8ec0-dfaab370141e",
						ID:    "*default",
						Value: -1.95001,
					},
				},
				"*sms": Balances{
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
				"*voice": Balances{
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
		ActionType: "*set_balance",
		Filter:     "{\"*and\":[{\"Value\":{\"*lt\":0}},{\"ID\":{\"*eq\":\"*default\"}}]}",
		Balance: &BalanceFilter{
			Type:     utils.StringPointer("*sms"),
			ID:       utils.StringPointer("for_v3hsillmilld500m_sms_ill"),
			Disabled: utils.BoolPointer(true),
		},
		Weight: 9,
	}
	a2 := &Action{
		ActionType: "*set_balance",
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
		ActionType: "*set_balance",
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
		ActionType: "*set_balance",
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
		ActionType: "*set_balance",
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

	afterUb, err := dm.DataDB().GetAccount("cgrates.org:af")
	if err != nil {
		t.Error("account not found: ", err, afterUb)
	}

	for btype, chain := range afterUb.BalanceMap {
		if btype != utils.MONETARY {
			for _, b := range chain {
				if b.Disabled != true {
					t.Errorf("Failed to disabled balance (%s): %+v", btype, b)
				}
			}
		}
	}
}

func TestActionSetBalance(t *testing.T) {
	err := dm.DataDB().SetAccount(
		&Account{
			ID: "cgrates.org:setb",
			BalanceMap: map[string]Balances{
				utils.MONETARY: Balances{
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
				utils.VOICE: Balances{
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
		ActionType: SET_BALANCE,
		Balance: &BalanceFilter{
			ID:     utils.StringPointer("m2"),
			Type:   utils.StringPointer(utils.MONETARY),
			Value:  &utils.ValueFormula{Static: 11},
			Weight: utils.Float64Pointer(10),
		},
	}

	at := &ActionTiming{
		accountIDs: utils.StringMap{"cgrates.org:setb": true},
		actions:    Actions{a},
	}
	at.Execute(nil, nil)

	afterUb, err := dm.DataDB().GetAccount("cgrates.org:setb")
	if err != nil {
		t.Error("account not found: ", err, afterUb)
	}
	if len(afterUb.BalanceMap[utils.MONETARY]) != 2 ||
		afterUb.BalanceMap[utils.MONETARY][1].Value != 11 ||
		afterUb.BalanceMap[utils.MONETARY][1].Weight != 10 {
		for _, b := range afterUb.BalanceMap[utils.MONETARY] {
			t.Logf("B: %+v", b)
		}
		t.Errorf("Balance: %+v", afterUb.BalanceMap[utils.MONETARY][1])
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
		afterUb, err := dm.DataDB().GetAccount("cgrates.org:expo")
		if err != nil ||
			len(afterUb.BalanceMap[utils.VOICE]) != rep+1 {
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
	afterUb, err := dm.DataDB().GetAccount("cgrates.org:expnoexp")
	if err != nil ||
		len(afterUb.BalanceMap[utils.VOICE]) != 2 {
		t.Error("error topuping expiration balance: ", utils.ToIJSON(afterUb))
	}
}

func TestActionTopUpZeroNegative(t *testing.T) {
	account := &Account{
		ID: "cgrates.org:zeroNegative",
		BalanceMap: map[string]Balances{
			utils.MONETARY: Balances{
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
	err := dm.DataDB().SetAccount(account)
	if err != nil {
		t.Error("Error setting account: ", err)
	}
	at := &ActionTiming{
		accountIDs: utils.StringMap{"cgrates.org:zeroNegative": true},
		Timing:     &RateInterval{},
		actions: []*Action{
			&Action{
				Id:         "ZeroMonetary",
				ActionType: TopUpZeroNegative,
				Balance: &BalanceFilter{
					Type: utils.StringPointer(utils.MONETARY),
				},
			},
		},
	}
	err = at.Execute(nil, nil)
	acc, err := dm.DataDB().GetAccount("cgrates.org:zeroNegative")
	if err != nil || acc == nil {
		t.Error("Error getting account: ", acc, err)
	}
	//Verify value for first balance(Bal1) should be 0 after execute action TopUpZeroNegative
	if acc.BalanceMap[utils.MONETARY][0].Value != 0 {
		t.Errorf("Expecting 0, received: %+v", acc.BalanceMap[utils.MONETARY][0].Value)
	}
	//Verify value for secound balance(Bal2) should be the same
	if acc.BalanceMap[utils.MONETARY][1].Value != 5 {
		t.Errorf("Expecting 5, received: %+v", acc.BalanceMap[utils.MONETARY][1].Value)
	}
}

func TestActionSetExpiry(t *testing.T) {
	var cloneTimeNowPlus24h time.Time
	timeNowPlus24h := time.Now().Add(time.Duration(24 * time.Hour))
	//Need clone because time.Now adds extra information that DeepEqual doesn't like
	if err := utils.Clone(timeNowPlus24h, &cloneTimeNowPlus24h); err != nil {
		t.Error(err)
	}
	account := &Account{
		ID: "cgrates.org:zeroNegative",
		BalanceMap: map[string]Balances{
			utils.MONETARY: Balances{
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
	err := dm.DataDB().SetAccount(account)
	if err != nil {
		t.Error("Error setting account: ", err)
	}
	at := &ActionTiming{
		accountIDs: utils.StringMap{"cgrates.org:zeroNegative": true},
		Timing:     &RateInterval{},
		actions: []*Action{
			&Action{
				Id:         "SetExpiry",
				ActionType: SetExpiry,
				Balance: &BalanceFilter{
					ID:             utils.StringPointer("Bal1"),
					Type:           utils.StringPointer(utils.MONETARY),
					ExpirationDate: utils.TimePointer(cloneTimeNowPlus24h),
				},
			},
		},
	}
	err = at.Execute(nil, nil)
	acc, err := dm.DataDB().GetAccount("cgrates.org:zeroNegative")
	if err != nil || acc == nil {
		t.Error("Error getting account: ", acc, err)
	}
	//Verify ExpirationDate for first balance(Bal1)
	if !acc.BalanceMap[utils.MONETARY][0].ExpirationDate.Equal(cloneTimeNowPlus24h) {
		t.Errorf("Expecting: %+v, received: %+v", cloneTimeNowPlus24h, acc.BalanceMap[utils.MONETARY][0].ExpirationDate)
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
	if _, err := dm.DataDB().GetAccount("cgrates.org:vf"); err != nil {
		t.Errorf("account to be removed not found: %v", err)
	}

	at := &ActionTiming{
		accountIDs: utils.StringMap{"cgrates.org:vf": true},
		ActionsID:  "VF",
	}
	at.Execute(nil, nil)
	afterUb, err := dm.DataDB().GetAccount("cgrates.org:vf")
	// not an exact value, depends of month
	v := afterUb.BalanceMap[utils.MONETARY].GetTotalValue()
	if err != nil || v > -0.30 || v < -0.36 {
		t.Error("error debiting account: ", err, utils.ToIJSON(afterUb), v)
	}
}

func TestClonedAction(t *testing.T) {
	a := &Action{
		Id:         "test1",
		ActionType: TOPUP,
		Balance: &BalanceFilter{
			ID:    utils.StringPointer("*default"),
			Value: &utils.ValueFormula{Static: 1},
			Type:  utils.StringPointer(utils.MONETARY),
		},
		Weight: float64(10),
	}

	clone := a.Clone()

	if !reflect.DeepEqual(a, clone) {
		t.Error("error cloning action: ", utils.ToIJSON(clone))
	}
}

func TestClonedActions(t *testing.T) {
	actions := Actions{
		&Action{
			Id:         "RECUR_FOR_V3HSILLMILLD1G",
			ActionType: TOPUP,
			Balance: &BalanceFilter{
				ID:    utils.StringPointer("*default"),
				Value: &utils.ValueFormula{Static: 1},
				Type:  utils.StringPointer(utils.MONETARY),
			},
			Weight: float64(30),
		},
		&Action{
			Id:         "RECUR_FOR_V3HSILLMILLD5G",
			ActionType: DEBIT,
			Balance: &BalanceFilter{
				ID:    utils.StringPointer("*default"),
				Value: &utils.ValueFormula{Static: 2},
				Type:  utils.StringPointer(utils.MONETARY),
			},
			Weight: float64(20),
		},
	}

	clone, err := actions.Clone()

	if err != nil {
		t.Error("error cloning actions: ", err)
	}

	if !reflect.DeepEqual(actions, clone) {
		t.Error("error cloning actions: ", utils.ToIJSON(clone))
	}

}

func TestCacheGetClonedActions(t *testing.T) {
	actions := Actions{
		&Action{
			Id:         "RECUR_FOR_V3HSILLMILLD1G",
			ActionType: TOPUP,
			Balance: &BalanceFilter{
				ID:    utils.StringPointer("*default"),
				Value: &utils.ValueFormula{Static: 1},
				Type:  utils.StringPointer(utils.MONETARY),
			},
			Weight: float64(30),
		},
		&Action{
			Id:         "REACT_FOR_V3HSILLMILL",
			ActionType: SET_BALANCE,
			Balance: &BalanceFilter{
				ID:    utils.StringPointer("for_v3hsillmill_sms_ill"),
				Type:  utils.StringPointer(utils.SMS),
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
	Cache.Set(utils.CacheActions, "MYTEST", actions, nil, true, "")
	clned, err := Cache.GetCloned(utils.CacheActions, "MYTEST")
	if err != nil {
		t.Error(err)
	}
	aCloned := clned.(Actions)
	if !reflect.DeepEqual(actions, aCloned) {
		t.Errorf("Expecting: %+v, received: %+v", actions[1].Balance, aCloned[1].Balance)
	}
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
