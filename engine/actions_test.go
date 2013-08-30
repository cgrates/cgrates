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
	"github.com/cgrates/cgrates/utils"
	"testing"
	"time"
)

func TestActionTimingNothing(t *testing.T) {
	at := &ActionTiming{}
	st := at.GetNextStartTime()
	expected := time.Date(1, 1, 1, 0, 0, 0, 0, time.UTC)
	if !st.Equal(expected) {
		t.Errorf("Expected %v was %v", expected, st)
	}
}

func TestActionTimingOnlyHour(t *testing.T) {
	at := &ActionTiming{Timing: &Interval{StartTime: "10:01:00"}}
	st := at.GetNextStartTime()
	now := time.Now()
	y, m, d := now.Date()
	expected := time.Date(y, m, d, 10, 1, 0, 0, time.Local)
	if !st.Equal(expected) {
		t.Errorf("Expected %v was %v", expected, st)
	}
}

func TestActionTimingOnlyWeekdays(t *testing.T) {
	at := &ActionTiming{Timing: &Interval{WeekDays: []time.Weekday{time.Monday}}}
	st := at.GetNextStartTime()
	now := time.Now()
	y, m, d := now.Date()
	h, min, s := now.Clock()
	e := time.Date(y, m, d, h, min, s, 0, time.Local)
	day := e.Day()
	for _, i := range []int{0, 1, 2, 3, 4, 5, 6, 7} {
		e = time.Date(e.Year(), e.Month(), day, e.Hour(), e.Minute(), e.Second(), e.Nanosecond(), e.Location()).AddDate(0, 0, i)
		if e.Weekday() == time.Monday && (e.Equal(now) || e.After(now)) {
			break
		}
	}
	if !st.Equal(e) {
		t.Errorf("Expected %v was %v", e, st)
	}
}

func TestActionTimingHourWeekdays(t *testing.T) {
	at := &ActionTiming{Timing: &Interval{WeekDays: []time.Weekday{time.Monday}, StartTime: "10:01:00"}}
	st := at.GetNextStartTime()
	now := time.Now()
	y, m, d := now.Date()
	e := time.Date(y, m, d, 10, 1, 0, 0, time.Local)
	day := e.Day()
	for _, i := range []int{0, 1, 2, 3, 4, 5, 6, 7} {
		e = time.Date(e.Year(), e.Month(), day, e.Hour(), e.Minute(), e.Second(), e.Nanosecond(), e.Location()).AddDate(0, 0, i)
		if e.Weekday() == time.Monday && (e.Equal(now) || e.After(now)) {
			break
		}
	}
	if !st.Equal(e) {
		t.Errorf("Expected %v was %v", e, st)
	}
}

func TestActionTimingOnlyMonthdays(t *testing.T) {
	now := time.Now()
	y, m, d := now.Date()
	tomorrow := time.Date(y, m, d, 0, 0, 0, 0, time.Local).AddDate(0, 0, 1)
	at := &ActionTiming{Timing: &Interval{MonthDays: MonthDays{1, 25, 2, tomorrow.Day()}}}
	st := at.GetNextStartTime()
	expected := time.Date(y, m, tomorrow.Day(), 0, 0, 0, 0, time.Local)
	if !st.Equal(expected) {
		t.Errorf("Expected %v was %v", expected, st)
	}
}

func TestActionTimingHourMonthdays(t *testing.T) {
	now := time.Now()
	y, m, d := now.Date()
	testTime := time.Date(y, m, d, 10, 1, 0, 0, time.Local)
	tomorrow := time.Date(y, m, d, 0, 0, 0, 0, time.Local).AddDate(0, 0, 1)
	day := now.Day()
	if now.After(testTime) {
		day = tomorrow.Day()
	}
	at := &ActionTiming{Timing: &Interval{MonthDays: MonthDays{now.Day(), tomorrow.Day()}, StartTime: "10:01:00"}}
	st := at.GetNextStartTime()
	expected := time.Date(y, m, day, 10, 1, 0, 0, time.Local)
	if !st.Equal(expected) {
		t.Errorf("Expected %v was %v", expected, st)
	}
}

func TestActionTimingOnlyMonths(t *testing.T) {
	now := time.Now()
	y, m, d := now.Date()
	nextMonth := time.Date(y, m, d, 0, 0, 0, 0, time.Local).AddDate(0, 1, 0)
	at := &ActionTiming{Timing: &Interval{Months: Months{time.February, time.May, nextMonth.Month()}}}
	st := at.GetNextStartTime()
	expected := time.Date(y, nextMonth.Month(), 1, 0, 0, 0, 0, time.Local)
	if !st.Equal(expected) {
		t.Errorf("Expected %v was %v", expected, st)
	}
}

func TestActionTimingHourMonths(t *testing.T) {
	now := time.Now()
	y, m, d := now.Date()
	testTime := time.Date(y, m, d, 10, 1, 0, 0, time.Local)
	nextMonth := time.Date(y, m, d, 0, 0, 0, 0, time.Local).AddDate(0, 1, 0)
	month := now.Month()
	if now.After(testTime) {
		month = nextMonth.Month()
	}
	at := &ActionTiming{Timing: &Interval{Months: Months{now.Month(), nextMonth.Month()}, StartTime: "10:01:00"}}
	st := at.GetNextStartTime()
	expected := time.Date(y, month, d, 10, 1, 0, 0, time.Local)
	if !st.Equal(expected) {
		t.Errorf("Expected %v was %v", expected, st)
	}
}

func TestActionTimingHourMonthdaysMonths(t *testing.T) {
	now := time.Now()
	y, m, d := now.Date()
	testTime := time.Date(y, m, d, 10, 1, 0, 0, time.Local)
	nextMonth := time.Date(y, m, d, 0, 0, 0, 0, time.Local).AddDate(0, 1, 0)
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
	at := &ActionTiming{Timing: &Interval{
		Months:    Months{now.Month(), nextMonth.Month()},
		MonthDays: MonthDays{now.Day(), tomorrow.Day()},
		StartTime: "10:01:00",
	}}
	st := at.GetNextStartTime()
	expected := time.Date(y, month, day, 10, 1, 0, 0, time.Local)
	if !st.Equal(expected) {
		t.Errorf("Expected %v was %v", expected, st)
	}
}

func TestActionTimingFirstOfTheMonth(t *testing.T) {
	now := time.Now()
	y, m, _ := now.Date()
	nextMonth := time.Date(y, m, 1, 0, 0, 0, 0, time.Local).AddDate(0, 1, 0)
	at := &ActionTiming{Timing: &Interval{
		Months:    Months{time.January, time.February, time.March, time.April, time.May, time.June, time.July, time.August, time.September, time.October, time.November, time.December},
		MonthDays: MonthDays{1},
	}}
	st := at.GetNextStartTime()
	expected := nextMonth
	if !st.Equal(expected) {
		t.Errorf("Expected %v was %v", expected, st)
	}
}

func TestActionTimingOnlyYears(t *testing.T) {
	now := time.Now()
	y, m, d := now.Date()
	nextYear := time.Date(y, m, d, 0, 0, 0, 0, time.Local).AddDate(1, 0, 0)
	at := &ActionTiming{Timing: &Interval{Years: Years{now.Year(), nextYear.Year()}}}
	st := at.GetNextStartTime()
	expected := time.Date(nextYear.Year(), 1, 1, 0, 0, 0, 0, time.Local)
	if !st.Equal(expected) {
		t.Errorf("Expected %v was %v", expected, st)
	}
}

func TestActionTimingHourYears(t *testing.T) {
	now := time.Now()
	y, m, d := now.Date()
	testTime := time.Date(y, m, d, 10, 1, 0, 0, time.Local)
	nextYear := time.Date(y, m, d, 0, 0, 0, 0, time.Local).AddDate(1, 0, 0)
	year := now.Year()
	if now.After(testTime) {
		year = nextYear.Year()
	}
	at := &ActionTiming{Timing: &Interval{Years: Years{now.Year(), nextYear.Year()}, StartTime: "10:01:00"}}
	st := at.GetNextStartTime()
	expected := time.Date(year, m, d, 10, 1, 0, 0, time.Local)
	if !st.Equal(expected) {
		t.Errorf("Expected %v was %v", expected, st)
	}
}

func TestActionTimingHourMonthdaysYear(t *testing.T) {
	now := time.Now()
	y, m, d := now.Date()
	testTime := time.Date(y, m, d, 10, 1, 0, 0, time.Local)
	nextYear := time.Date(y, m, d, 0, 0, 0, 0, time.Local).AddDate(1, 0, 0)
	tomorrow := time.Date(y, m, d, 0, 0, 0, 0, time.Local).AddDate(0, 0, 1)
	day := now.Day()
	if now.After(testTime) {
		day = tomorrow.Day()
	}
	nextDay := time.Date(y, m, day, 10, 1, 0, 0, time.Local)
	year := now.Year()
	if nextDay.Before(now) {
		if now.After(testTime) {
			year = nextYear.Year()
		}
	}
	at := &ActionTiming{Timing: &Interval{
		Years:     Years{now.Year(), nextYear.Year()},
		MonthDays: MonthDays{now.Day(), tomorrow.Day()},
		StartTime: "10:01:00",
	}}
	st := at.GetNextStartTime()
	expected := time.Date(year, m, day, 10, 1, 0, 0, time.Local)
	if !st.Equal(expected) {
		t.Errorf("Expected %v was %v", expected, st)
	}
}

func TestActionTimingHourMonthdaysMonthYear(t *testing.T) {
	now := time.Now()
	y, m, d := now.Date()
	testTime := time.Date(y, m, d, 10, 1, 0, 0, time.Local)
	nextYear := time.Date(y, m, d, 0, 0, 0, 0, time.Local).AddDate(1, 0, 0)
	nextMonth := time.Date(y, m, d, 0, 0, 0, 0, time.Local).AddDate(0, 1, 0)
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
	at := &ActionTiming{Timing: &Interval{
		Years:     Years{now.Year(), nextYear.Year()},
		Months:    Months{now.Month(), nextMonth.Month()},
		MonthDays: MonthDays{now.Day(), tomorrow.Day()},
		StartTime: "10:01:00",
	}}
	st := at.GetNextStartTime()
	expected := time.Date(year, month, day, 10, 1, 0, 0, time.Local)
	if !st.Equal(expected) {
		t.Errorf("Expected %v was %v", expected, st)
	}
}

func TestActionTimingFirstOfTheYear(t *testing.T) {
	now := time.Now()
	y, _, _ := now.Date()
	nextYear := time.Date(y, 1, 1, 0, 0, 0, 0, time.Local).AddDate(1, 0, 0)
	at := &ActionTiming{Timing: &Interval{
		Years:     Years{nextYear.Year()},
		Months:    Months{time.January},
		MonthDays: MonthDays{1},
		StartTime: "00:00:00",
	}}
	st := at.GetNextStartTime()
	expected := nextYear
	if !st.Equal(expected) {
		t.Errorf("Expected %v was %v", expected, st)
	}
}

func TestActionTimingFirstMonthOfTheYear(t *testing.T) {
	now := time.Now()
	y, _, _ := now.Date()
	nextYear := time.Date(y, 1, 1, 0, 0, 0, 0, time.Local).AddDate(1, 0, 0)
	at := &ActionTiming{Timing: &Interval{
		Months: Months{time.January},
	}}
	st := at.GetNextStartTime()
	expected := nextYear
	if !st.Equal(expected) {
		t.Errorf("Expected %v was %v", expected, st)
	}
}

func TestActionTimingCheckForASAP(t *testing.T) {
	at := &ActionTiming{Timing: &Interval{StartTime: ASAP}}
	if !at.CheckForASAP() {
		t.Errorf("%v should be asap!", at)
	}
}

func TestActionTimingIsOneTimeRun(t *testing.T) {
	at := &ActionTiming{Timing: &Interval{StartTime: ASAP}}
	if !at.CheckForASAP() {
		t.Errorf("%v should be asap!", at)
	}
	if !at.IsOneTimeRun() {
		t.Errorf("%v should be one time run!", at)
	}
}

func TestActionTimingOneTimeRun(t *testing.T) {
	at := &ActionTiming{Timing: &Interval{StartTime: ASAP}}
	at.CheckForASAP()
	nextRun := at.GetNextStartTime()
	if nextRun.IsZero() {
		t.Error("next time failed for asap")
	}
}

func TestActionTimingLogFunction(t *testing.T) {
	a := &Action{
		ActionType:   "*log",
		BalanceId:    "test",
		Units:        1.1,
		MinuteBucket: &MinuteBucket{},
	}
	at := &ActionTiming{
		actions: []*Action{a},
	}
	err := at.Execute()
	if err != nil {
		t.Errorf("Could not execute LOG action: %v", err)
	}
}

func TestActionTimingPriotityListSortByWeight(t *testing.T) {
	at1 := &ActionTiming{Timing: &Interval{
		Years:     Years{2100},
		Months:    Months{time.January, time.February, time.March, time.April, time.May, time.June, time.July, time.August, time.September, time.October, time.November, time.December},
		MonthDays: MonthDays{1},
		StartTime: "00:00:00",
		Weight:    20,
	}}
	at2 := &ActionTiming{Timing: &Interval{
		Years:     Years{2100},
		Months:    Months{time.January, time.February, time.March, time.April, time.May, time.June, time.July, time.August, time.September, time.October, time.November, time.December},
		MonthDays: MonthDays{2},
		StartTime: "00:00:00",
		Weight:    10,
	}}
	var atpl ActionTimingPriotityList
	atpl = append(atpl, at2, at1)
	atpl.Sort()
	if atpl[0] != at1 || atpl[1] != at2 {
		t.Error("Timing list not sorted correctly: ", at1, at2, atpl)
	}
}

func TestActionTimingPriotityListWeight(t *testing.T) {
	at1 := &ActionTiming{
		Timing: &Interval{
			Months:    Months{time.January, time.February, time.March, time.April, time.May, time.June, time.July, time.August, time.September, time.October, time.November, time.December},
			MonthDays: MonthDays{1},
			StartTime: "00:00:00",
		},
		Weight: 10.0,
	}
	at2 := &ActionTiming{
		Timing: &Interval{
			Months:    Months{time.January, time.February, time.March, time.April, time.May, time.June, time.July, time.August, time.September, time.October, time.November, time.December},
			MonthDays: MonthDays{1},
			StartTime: "00:00:00",
		},
		Weight: 20.0,
	}
	var atpl ActionTimingPriotityList
	atpl = append(atpl, at2, at1)
	atpl.Sort()
	if atpl[0] != at1 || atpl[1] != at2 {
		t.Error("Timing list not sorted correctly: ", atpl)
	}
}

func TestActionTriggerMatchNil(t *testing.T) {
	at := &ActionTrigger{
		Direction:      OUTBOUND,
		BalanceId:      CREDIT,
		ThresholdType:  TRIGGER_MAX_BALANCE,
		ThresholdValue: 2,
	}
	var a *Action
	if !at.Match(a) {
		t.Errorf("Action trigger [%v] does not match action [%v]", at, a)
	}
}

func TestActionTriggerMatchAllBlank(t *testing.T) {
	at := &ActionTrigger{
		Direction:      OUTBOUND,
		BalanceId:      CREDIT,
		ThresholdType:  TRIGGER_MAX_BALANCE,
		ThresholdValue: 2,
	}
	a := &Action{}
	if !at.Match(a) {
		t.Errorf("Action trigger [%v] does not match action [%v]", at, a)
	}
}

func TestActionTriggerMatchMinuteBucketBlank(t *testing.T) {
	at := &ActionTrigger{
		Direction:      OUTBOUND,
		BalanceId:      CREDIT,
		ThresholdType:  TRIGGER_MAX_BALANCE,
		ThresholdValue: 2,
	}
	a := &Action{Direction: OUTBOUND, BalanceId: CREDIT}
	if !at.Match(a) {
		t.Errorf("Action trigger [%v] does not match action [%v]", at, a)
	}
}

func TestActionTriggerMatchMinuteBucketFull(t *testing.T) {
	at := &ActionTrigger{
		Direction:      OUTBOUND,
		BalanceId:      CREDIT,
		ThresholdType:  TRIGGER_MAX_BALANCE,
		ThresholdValue: 2,
	}
	a := &Action{MinuteBucket: &MinuteBucket{PriceType: TRIGGER_MAX_BALANCE, Price: 2}}
	if !at.Match(a) {
		t.Errorf("Action trigger [%v] does not match action [%v]", at, a)
	}
}

func TestActionTriggerMatchAllFull(t *testing.T) {
	at := &ActionTrigger{
		Direction:      OUTBOUND,
		BalanceId:      CREDIT,
		ThresholdType:  TRIGGER_MAX_BALANCE,
		ThresholdValue: 2,
	}
	a := &Action{Direction: OUTBOUND, BalanceId: CREDIT, MinuteBucket: &MinuteBucket{PriceType: TRIGGER_MAX_BALANCE, Price: 2}}
	if !at.Match(a) {
		t.Errorf("Action trigger [%v] does not match action [%v]", at, a)
	}
}

func TestActionTriggerMatchSomeFalse(t *testing.T) {
	at := &ActionTrigger{
		Direction:      OUTBOUND,
		BalanceId:      CREDIT,
		ThresholdType:  TRIGGER_MAX_BALANCE,
		ThresholdValue: 2,
	}
	a := &Action{Direction: INBOUND, BalanceId: CREDIT, MinuteBucket: &MinuteBucket{PriceType: TRIGGER_MAX_BALANCE, Price: 2}}
	if at.Match(a) {
		t.Errorf("Action trigger [%v] does not match action [%v]", at, a)
	}
}

func TestActionTriggerMatcMinuteBucketFalse(t *testing.T) {
	at := &ActionTrigger{
		Direction:      OUTBOUND,
		BalanceId:      CREDIT,
		ThresholdType:  TRIGGER_MAX_BALANCE,
		ThresholdValue: 2,
	}
	a := &Action{Direction: OUTBOUND, BalanceId: CREDIT, MinuteBucket: &MinuteBucket{PriceType: TRIGGER_MAX_BALANCE, Price: 3}}
	if at.Match(a) {
		t.Errorf("Action trigger [%v] does not match action [%v]", at, a)
	}
}

func TestActionTriggerMatcAllFalse(t *testing.T) {
	at := &ActionTrigger{
		Direction:      OUTBOUND,
		BalanceId:      CREDIT,
		ThresholdType:  TRIGGER_MAX_BALANCE,
		ThresholdValue: 2,
	}
	a := &Action{Direction: INBOUND, BalanceId: MINUTES, MinuteBucket: &MinuteBucket{PriceType: TRIGGER_MAX_COUNTER, Price: 3}}
	if at.Match(a) {
		t.Errorf("Action trigger [%v] does not match action [%v]", at, a)
	}
}

func TestActionTriggerPriotityList(t *testing.T) {
	at1 := &ActionTrigger{Weight: 10}
	at2 := &ActionTrigger{Weight: 20}
	at3 := &ActionTrigger{Weight: 30}
	var atpl ActionTriggerPriotityList
	atpl = append(atpl, at2, at1, at3)
	atpl.Sort()
	if atpl[0] != at1 || atpl[2] != at3 || atpl[1] != at2 {
		t.Error("List not sorted: ", atpl)
	}
}

/*func TestActionLog(t *testing.T) {
	a := &Action{
		ActionType:   "TEST",
		BalanceId:    "BALANCE",
		Units:        10,
		Weight:       11,
		MinuteBucket: &MinuteBucket{},
	}
	logAction(nil, a)
}*/

func TestActionResetTriggres(t *testing.T) {
	ub := &UserBalance{
		Id:             "TEST_UB",
		BalanceMap:     map[string]BalanceChain{CREDIT: BalanceChain{&Balance{Value: 10}}},
		UnitCounters:   []*UnitsCounter{&UnitsCounter{BalanceId: CREDIT, Units: 1}},
		MinuteBuckets:  []*MinuteBucket{&MinuteBucket{Seconds: 10, Weight: 20, Price: 1, DestinationId: "NAT"}, &MinuteBucket{Weight: 10, Price: 10, PriceType: PRICE_ABSOLUTE, DestinationId: "RET"}},
		ActionTriggers: ActionTriggerPriotityList{&ActionTrigger{BalanceId: CREDIT, ThresholdValue: 2, ActionsId: "TEST_ACTIONS", Executed: true}, &ActionTrigger{BalanceId: CREDIT, ThresholdValue: 2, ActionsId: "TEST_ACTIONS", Executed: true}},
	}
	resetTriggersAction(ub, nil)
	if ub.ActionTriggers[0].Executed == true || ub.ActionTriggers[1].Executed == true {
		t.Error("Reset triggers action failed!")
	}
}

func TestActionResetTriggresExecutesThem(t *testing.T) {
	ub := &UserBalance{
		Id:             "TEST_UB",
		BalanceMap:     map[string]BalanceChain{CREDIT: BalanceChain{&Balance{Value: 10}}},
		UnitCounters:   []*UnitsCounter{&UnitsCounter{BalanceId: CREDIT, Units: 1}},
		ActionTriggers: ActionTriggerPriotityList{&ActionTrigger{BalanceId: CREDIT, ThresholdValue: 2, ActionsId: "TEST_ACTIONS", Executed: true}},
	}
	resetTriggersAction(ub, nil)
	if ub.ActionTriggers[0].Executed == true || ub.BalanceMap[CREDIT][0].Value == 12 {
		t.Error("Reset triggers action failed!")
	}
}

func TestActionResetTriggresActionFilter(t *testing.T) {
	ub := &UserBalance{
		Id:             "TEST_UB",
		BalanceMap:     map[string]BalanceChain{CREDIT: BalanceChain{&Balance{Value: 10}}},
		UnitCounters:   []*UnitsCounter{&UnitsCounter{BalanceId: CREDIT, Units: 1}},
		MinuteBuckets:  []*MinuteBucket{&MinuteBucket{Seconds: 10, Weight: 20, Price: 1, DestinationId: "NAT"}, &MinuteBucket{Weight: 10, Price: 10, PriceType: PRICE_ABSOLUTE, DestinationId: "RET"}},
		ActionTriggers: ActionTriggerPriotityList{&ActionTrigger{BalanceId: CREDIT, ThresholdValue: 2, ActionsId: "TEST_ACTIONS", Executed: true}, &ActionTrigger{BalanceId: CREDIT, ThresholdValue: 2, ActionsId: "TEST_ACTIONS", Executed: true}},
	}
	resetTriggersAction(ub, &Action{BalanceId: SMS})
	if ub.ActionTriggers[0].Executed == false || ub.ActionTriggers[1].Executed == false {
		t.Error("Reset triggers action failed!")
	}
}

func TestActionSetPostpaid(t *testing.T) {
	ub := &UserBalance{
		Id:             "TEST_UB",
		Type:           UB_TYPE_PREPAID,
		BalanceMap:     map[string]BalanceChain{CREDIT: BalanceChain{&Balance{Value: 100}}},
		UnitCounters:   []*UnitsCounter{&UnitsCounter{BalanceId: CREDIT, Units: 1}},
		MinuteBuckets:  []*MinuteBucket{&MinuteBucket{Seconds: 10, Weight: 20, Price: 1, DestinationId: "NAT"}, &MinuteBucket{Weight: 10, Price: 10, PriceType: PRICE_ABSOLUTE, DestinationId: "RET"}},
		ActionTriggers: ActionTriggerPriotityList{&ActionTrigger{BalanceId: CREDIT, ThresholdValue: 2, ActionsId: "TEST_ACTIONS", Executed: true}, &ActionTrigger{BalanceId: CREDIT, ThresholdValue: 2, ActionsId: "TEST_ACTIONS", Executed: true}},
	}
	setPostpaidAction(ub, nil)
	if ub.Type != UB_TYPE_POSTPAID {
		t.Error("Set postpaid action failed!")
	}
}

func TestActionSetPrepaid(t *testing.T) {
	ub := &UserBalance{
		Id:             "TEST_UB",
		Type:           UB_TYPE_POSTPAID,
		BalanceMap:     map[string]BalanceChain{CREDIT: BalanceChain{&Balance{Value: 100}}},
		UnitCounters:   []*UnitsCounter{&UnitsCounter{BalanceId: CREDIT, Units: 1}},
		MinuteBuckets:  []*MinuteBucket{&MinuteBucket{Seconds: 10, Weight: 20, Price: 1, DestinationId: "NAT"}, &MinuteBucket{Weight: 10, Price: 10, PriceType: PRICE_ABSOLUTE, DestinationId: "RET"}},
		ActionTriggers: ActionTriggerPriotityList{&ActionTrigger{BalanceId: CREDIT, ThresholdValue: 2, ActionsId: "TEST_ACTIONS", Executed: true}, &ActionTrigger{BalanceId: CREDIT, ThresholdValue: 2, ActionsId: "TEST_ACTIONS", Executed: true}},
	}
	setPrepaidAction(ub, nil)
	if ub.Type != UB_TYPE_PREPAID {
		t.Error("Set prepaid action failed!")
	}
}

func TestActionResetPrepaid(t *testing.T) {
	ub := &UserBalance{
		Id:             "TEST_UB",
		Type:           UB_TYPE_POSTPAID,
		BalanceMap:     map[string]BalanceChain{CREDIT: BalanceChain{&Balance{Value: 100}}},
		UnitCounters:   []*UnitsCounter{&UnitsCounter{BalanceId: CREDIT, Units: 1}},
		MinuteBuckets:  []*MinuteBucket{&MinuteBucket{Seconds: 10, Weight: 20, Price: 1, DestinationId: "NAT"}, &MinuteBucket{Weight: 10, Price: 10, PriceType: PRICE_ABSOLUTE, DestinationId: "RET"}},
		ActionTriggers: ActionTriggerPriotityList{&ActionTrigger{BalanceId: SMS, ThresholdValue: 2, ActionsId: "TEST_ACTIONS", Executed: true}, &ActionTrigger{BalanceId: SMS, ThresholdValue: 2, ActionsId: "TEST_ACTIONS", Executed: true}},
	}
	resetPrepaidAction(ub, nil)
	if ub.Type != UB_TYPE_PREPAID ||
		ub.BalanceMap[CREDIT].GetTotalValue() != 0 ||
		len(ub.UnitCounters) != 0 ||
		len(ub.MinuteBuckets) != 0 ||
		ub.ActionTriggers[0].Executed == true || ub.ActionTriggers[1].Executed == true {
		t.Error("Reset prepaid action failed!")
	}
}

func TestActionResetPostpaid(t *testing.T) {
	ub := &UserBalance{
		Id:             "TEST_UB",
		Type:           UB_TYPE_PREPAID,
		BalanceMap:     map[string]BalanceChain{CREDIT: BalanceChain{&Balance{Value: 100}}},
		UnitCounters:   []*UnitsCounter{&UnitsCounter{BalanceId: CREDIT, Units: 1}},
		MinuteBuckets:  []*MinuteBucket{&MinuteBucket{Seconds: 10, Weight: 20, Price: 1, DestinationId: "NAT"}, &MinuteBucket{Weight: 10, Price: 10, PriceType: PRICE_ABSOLUTE, DestinationId: "RET"}},
		ActionTriggers: ActionTriggerPriotityList{&ActionTrigger{BalanceId: SMS, ThresholdValue: 2, ActionsId: "TEST_ACTIONS", Executed: true}, &ActionTrigger{BalanceId: SMS, ThresholdValue: 2, ActionsId: "TEST_ACTIONS", Executed: true}},
	}
	resetPostpaidAction(ub, nil)
	if ub.Type != UB_TYPE_POSTPAID ||
		ub.BalanceMap[CREDIT].GetTotalValue() != 0 ||
		len(ub.UnitCounters) != 0 ||
		len(ub.MinuteBuckets) != 0 ||
		ub.ActionTriggers[0].Executed == true || ub.ActionTriggers[1].Executed == true {
		t.Error("Reset postpaid action failed!")
	}
}

func TestActionTopupResetCredit(t *testing.T) {
	ub := &UserBalance{
		Id:             "TEST_UB",
		Type:           UB_TYPE_PREPAID,
		BalanceMap:     map[string]BalanceChain{CREDIT + OUTBOUND: BalanceChain{&Balance{Value: 100}}},
		UnitCounters:   []*UnitsCounter{&UnitsCounter{BalanceId: CREDIT, Direction: OUTBOUND, Units: 1}},
		MinuteBuckets:  []*MinuteBucket{&MinuteBucket{Seconds: 10, Weight: 20, Price: 1, DestinationId: "NAT"}, &MinuteBucket{Weight: 10, Price: 10, PriceType: PRICE_ABSOLUTE, DestinationId: "RET"}},
		ActionTriggers: ActionTriggerPriotityList{&ActionTrigger{BalanceId: CREDIT, Direction: OUTBOUND, ThresholdValue: 2, ActionsId: "TEST_ACTIONS", Executed: true}, &ActionTrigger{BalanceId: CREDIT, Direction: OUTBOUND, ThresholdValue: 2, ActionsId: "TEST_ACTIONS", Executed: true}},
	}
	a := &Action{BalanceId: CREDIT, Direction: OUTBOUND, Units: 10}
	topupResetAction(ub, a)
	if ub.Type != UB_TYPE_PREPAID ||
		ub.BalanceMap[CREDIT+OUTBOUND].GetTotalValue() != 10 ||
		len(ub.UnitCounters) != 1 ||
		len(ub.MinuteBuckets) != 2 ||
		ub.ActionTriggers[0].Executed != true || ub.ActionTriggers[1].Executed != true {
		t.Errorf("Topup reset action failed: %#v", ub.BalanceMap[CREDIT+OUTBOUND].GetTotalValue())
	}
}

func TestActionTopupResetMinutes(t *testing.T) {
	ub := &UserBalance{
		Id:             "TEST_UB",
		Type:           UB_TYPE_PREPAID,
		BalanceMap:     map[string]BalanceChain{CREDIT + OUTBOUND: BalanceChain{&Balance{Value: 100}}},
		UnitCounters:   []*UnitsCounter{&UnitsCounter{BalanceId: CREDIT, Direction: OUTBOUND, Units: 1}},
		MinuteBuckets:  []*MinuteBucket{&MinuteBucket{Seconds: 10, Weight: 20, Price: 1, DestinationId: "NAT"}, &MinuteBucket{Weight: 10, Price: 10, PriceType: PRICE_ABSOLUTE, DestinationId: "RET"}},
		ActionTriggers: ActionTriggerPriotityList{&ActionTrigger{BalanceId: CREDIT, Direction: OUTBOUND, ThresholdValue: 2, ActionsId: "TEST_ACTIONS", Executed: true}, &ActionTrigger{BalanceId: CREDIT, Direction: OUTBOUND, ThresholdValue: 2, ActionsId: "TEST_ACTIONS", Executed: true}},
	}
	a := &Action{BalanceId: MINUTES, Direction: OUTBOUND, MinuteBucket: &MinuteBucket{Seconds: 5, Weight: 20, Price: 1, DestinationId: "NAT"}}
	topupResetAction(ub, a)
	if ub.Type != UB_TYPE_PREPAID ||
		ub.MinuteBuckets[0].Seconds != 5 ||
		ub.BalanceMap[CREDIT+OUTBOUND].GetTotalValue() != 100 ||
		len(ub.UnitCounters) != 1 ||
		len(ub.MinuteBuckets) != 1 ||
		ub.ActionTriggers[0].Executed != true || ub.ActionTriggers[1].Executed != true {
		t.Error("Topup reset minutes action failed!", ub.MinuteBuckets[0])
	}
}

func TestActionTopupCredit(t *testing.T) {
	ub := &UserBalance{
		Id:             "TEST_UB",
		Type:           UB_TYPE_PREPAID,
		BalanceMap:     map[string]BalanceChain{CREDIT + OUTBOUND: BalanceChain{&Balance{Value: 100}}},
		UnitCounters:   []*UnitsCounter{&UnitsCounter{BalanceId: CREDIT, Direction: OUTBOUND, Units: 1}},
		MinuteBuckets:  []*MinuteBucket{&MinuteBucket{Seconds: 10, Weight: 20, Price: 1, DestinationId: "NAT"}, &MinuteBucket{Weight: 10, Price: 10, PriceType: PRICE_ABSOLUTE, DestinationId: "RET"}},
		ActionTriggers: ActionTriggerPriotityList{&ActionTrigger{BalanceId: CREDIT, Direction: OUTBOUND, ThresholdValue: 2, ActionsId: "TEST_ACTIONS", Executed: true}, &ActionTrigger{BalanceId: CREDIT, Direction: OUTBOUND, ThresholdValue: 2, ActionsId: "TEST_ACTIONS", Executed: true}},
	}
	a := &Action{BalanceId: CREDIT, Direction: OUTBOUND, Units: 10}
	topupAction(ub, a)
	if ub.Type != UB_TYPE_PREPAID ||
		ub.BalanceMap[CREDIT+OUTBOUND].GetTotalValue() != 110 ||
		len(ub.UnitCounters) != 1 ||
		len(ub.MinuteBuckets) != 2 ||
		ub.ActionTriggers[0].Executed != true || ub.ActionTriggers[1].Executed != true {
		t.Error("Topup action failed!", ub.BalanceMap[CREDIT+OUTBOUND].GetTotalValue())
	}
}

func TestActionTopupMinutes(t *testing.T) {
	ub := &UserBalance{
		Id:             "TEST_UB",
		Type:           UB_TYPE_PREPAID,
		BalanceMap:     map[string]BalanceChain{CREDIT: BalanceChain{&Balance{Value: 100}}},
		UnitCounters:   []*UnitsCounter{&UnitsCounter{BalanceId: CREDIT, Units: 1}},
		MinuteBuckets:  []*MinuteBucket{&MinuteBucket{Seconds: 10, Weight: 20, Price: 1, DestinationId: "NAT"}, &MinuteBucket{Weight: 10, Price: 10, PriceType: PRICE_ABSOLUTE, DestinationId: "RET"}},
		ActionTriggers: ActionTriggerPriotityList{&ActionTrigger{BalanceId: CREDIT, ThresholdValue: 2, ActionsId: "TEST_ACTIONS", Executed: true}, &ActionTrigger{BalanceId: CREDIT, ThresholdValue: 2, ActionsId: "TEST_ACTIONS", Executed: true}},
	}
	a := &Action{BalanceId: MINUTES, MinuteBucket: &MinuteBucket{Seconds: 5, Weight: 20, Price: 1, DestinationId: "NAT"}}
	topupAction(ub, a)
	if ub.Type != UB_TYPE_PREPAID ||
		ub.MinuteBuckets[0].Seconds != 15 ||
		ub.BalanceMap[CREDIT].GetTotalValue() != 100 ||
		len(ub.UnitCounters) != 1 ||
		len(ub.MinuteBuckets) != 2 ||
		ub.ActionTriggers[0].Executed != true || ub.ActionTriggers[1].Executed != true {
		t.Error("Topup minutes action failed!", ub.MinuteBuckets[0])
	}
}

func TestActionDebitCredit(t *testing.T) {
	ub := &UserBalance{
		Id:             "TEST_UB",
		Type:           UB_TYPE_PREPAID,
		BalanceMap:     map[string]BalanceChain{CREDIT + OUTBOUND: BalanceChain{&Balance{Value: 100}}},
		UnitCounters:   []*UnitsCounter{&UnitsCounter{BalanceId: CREDIT, Direction: OUTBOUND, Units: 1}},
		MinuteBuckets:  []*MinuteBucket{&MinuteBucket{Seconds: 10, Weight: 20, Price: 1, DestinationId: "NAT"}, &MinuteBucket{Weight: 10, Price: 10, PriceType: PRICE_ABSOLUTE, DestinationId: "RET"}},
		ActionTriggers: ActionTriggerPriotityList{&ActionTrigger{BalanceId: CREDIT, Direction: OUTBOUND, ThresholdValue: 2, ActionsId: "TEST_ACTIONS", Executed: true}, &ActionTrigger{BalanceId: CREDIT, Direction: OUTBOUND, ThresholdValue: 2, ActionsId: "TEST_ACTIONS", Executed: true}},
	}
	a := &Action{BalanceId: CREDIT, Direction: OUTBOUND, Units: 10}
	debitAction(ub, a)
	if ub.Type != UB_TYPE_PREPAID ||
		ub.BalanceMap[CREDIT+OUTBOUND].GetTotalValue() != 90 ||
		len(ub.UnitCounters) != 1 ||
		len(ub.MinuteBuckets) != 2 ||
		ub.ActionTriggers[0].Executed != true || ub.ActionTriggers[1].Executed != true {
		t.Error("Debit action failed!", ub.BalanceMap[CREDIT+OUTBOUND].GetTotalValue())
	}
}

func TestActionDebitMinutes(t *testing.T) {
	ub := &UserBalance{
		Id:             "TEST_UB",
		Type:           UB_TYPE_PREPAID,
		BalanceMap:     map[string]BalanceChain{CREDIT: BalanceChain{&Balance{Value: 100}}},
		UnitCounters:   []*UnitsCounter{&UnitsCounter{BalanceId: CREDIT, Units: 1}},
		MinuteBuckets:  []*MinuteBucket{&MinuteBucket{Seconds: 10, Weight: 20, Price: 1, DestinationId: "NAT"}, &MinuteBucket{Weight: 10, Price: 10, PriceType: PRICE_ABSOLUTE, DestinationId: "RET"}},
		ActionTriggers: ActionTriggerPriotityList{&ActionTrigger{BalanceId: CREDIT, ThresholdValue: 2, ActionsId: "TEST_ACTIONS", Executed: true}, &ActionTrigger{BalanceId: CREDIT, ThresholdValue: 2, ActionsId: "TEST_ACTIONS", Executed: true}},
	}
	a := &Action{BalanceId: MINUTES, MinuteBucket: &MinuteBucket{Seconds: 5, Weight: 20, Price: 1, DestinationId: "NAT"}}
	debitAction(ub, a)
	if ub.Type != UB_TYPE_PREPAID ||
		ub.MinuteBuckets[0].Seconds != 5 ||
		ub.BalanceMap[CREDIT].GetTotalValue() != 100 ||
		len(ub.UnitCounters) != 1 ||
		len(ub.MinuteBuckets) != 2 ||
		ub.ActionTriggers[0].Executed != true || ub.ActionTriggers[1].Executed != true {
		t.Error("Debit minutes action failed!", ub.MinuteBuckets[0])
	}
}

func TestActionResetAllCounters(t *testing.T) {
	ub := &UserBalance{
		Id:             "TEST_UB",
		Type:           UB_TYPE_POSTPAID,
		BalanceMap:     map[string]BalanceChain{CREDIT: BalanceChain{&Balance{Value: 100}}},
		UnitCounters:   []*UnitsCounter{&UnitsCounter{BalanceId: CREDIT, Units: 1}},
		MinuteBuckets:  []*MinuteBucket{&MinuteBucket{Seconds: 10, Weight: 20, Price: 1, DestinationId: "NAT"}, &MinuteBucket{Weight: 10, Price: 10, PriceType: PRICE_ABSOLUTE, DestinationId: "RET"}},
		ActionTriggers: ActionTriggerPriotityList{&ActionTrigger{BalanceId: CREDIT, ThresholdValue: 2, ActionsId: "TEST_ACTIONS", Executed: true}},
	}
	resetCountersAction(ub, nil)
	if ub.Type != UB_TYPE_POSTPAID ||
		ub.BalanceMap[CREDIT].GetTotalValue() != 100 ||
		len(ub.UnitCounters) != 1 ||
		len(ub.UnitCounters[0].MinuteBuckets) != 1 ||
		len(ub.MinuteBuckets) != 2 ||
		ub.ActionTriggers[0].Executed != true {
		t.Error("Reset counters action failed!")
	}
	if len(ub.UnitCounters) < 1 {
		t.FailNow()
	}
	mb := ub.UnitCounters[0].MinuteBuckets[0]
	if mb.Weight != 20 || mb.Price != 1 || mb.Seconds != 10 || mb.DestinationId != "NAT" {
		t.Errorf("Minute bucked cloned incorrectly: %v!", mb)
	}
}

func TestActionResetCounterMinutes(t *testing.T) {
	ub := &UserBalance{
		Id:             "TEST_UB",
		Type:           UB_TYPE_POSTPAID,
		BalanceMap:     map[string]BalanceChain{CREDIT: BalanceChain{&Balance{Value: 100}}},
		UnitCounters:   []*UnitsCounter{&UnitsCounter{BalanceId: CREDIT, Units: 1}},
		MinuteBuckets:  []*MinuteBucket{&MinuteBucket{Seconds: 10, Weight: 20, Price: 1, DestinationId: "NAT"}, &MinuteBucket{Weight: 10, Price: 10, PriceType: PRICE_ABSOLUTE, DestinationId: "RET"}},
		ActionTriggers: ActionTriggerPriotityList{&ActionTrigger{BalanceId: CREDIT, ThresholdValue: 2, ActionsId: "TEST_ACTIONS", Executed: true}},
	}
	a := &Action{BalanceId: MINUTES}
	resetCounterAction(ub, a)
	if ub.Type != UB_TYPE_POSTPAID ||
		ub.BalanceMap[CREDIT].GetTotalValue() != 100 ||
		len(ub.UnitCounters) != 2 ||
		len(ub.UnitCounters[1].MinuteBuckets) != 1 ||
		len(ub.MinuteBuckets) != 2 ||
		ub.ActionTriggers[0].Executed != true {
		t.Error("Reset counters action failed!", ub.UnitCounters[1].MinuteBuckets)
	}
	if len(ub.UnitCounters) < 2 || len(ub.UnitCounters[1].MinuteBuckets) < 1 {
		t.FailNow()
	}
	mb := ub.UnitCounters[1].MinuteBuckets[0]
	if mb.Weight != 20 || mb.Price != 1 || mb.Seconds != 10 || mb.DestinationId != "NAT" {
		t.Errorf("Minute bucked cloned incorrectly: %v!", mb)
	}
}

func TestActionResetCounterCREDIT(t *testing.T) {
	ub := &UserBalance{
		Id:             "TEST_UB",
		Type:           UB_TYPE_POSTPAID,
		BalanceMap:     map[string]BalanceChain{CREDIT: BalanceChain{&Balance{Value: 100}}},
		UnitCounters:   []*UnitsCounter{&UnitsCounter{BalanceId: CREDIT, Direction: OUTBOUND, Units: 1}, &UnitsCounter{BalanceId: SMS, Direction: OUTBOUND, Units: 1}},
		MinuteBuckets:  []*MinuteBucket{&MinuteBucket{Seconds: 10, Weight: 20, Price: 1, DestinationId: "NAT"}, &MinuteBucket{Weight: 10, Price: 10, PriceType: PRICE_ABSOLUTE, DestinationId: "RET"}},
		ActionTriggers: ActionTriggerPriotityList{&ActionTrigger{BalanceId: CREDIT, Direction: OUTBOUND, ThresholdValue: 2, ActionsId: "TEST_ACTIONS", Executed: true}},
	}
	a := &Action{BalanceId: CREDIT, Direction: OUTBOUND}
	resetCounterAction(ub, a)
	if ub.Type != UB_TYPE_POSTPAID ||
		ub.BalanceMap[CREDIT].GetTotalValue() != 100 ||
		len(ub.UnitCounters) != 2 ||
		len(ub.MinuteBuckets) != 2 ||
		ub.ActionTriggers[0].Executed != true {
		t.Error("Reset counters action failed!", ub.UnitCounters)
	}
}

func TestActionTriggerLogging(t *testing.T) {
	at := &ActionTrigger{
		Id:             "some_uuid",
		BalanceId:      CREDIT,
		Direction:      OUTBOUND,
		ThresholdValue: 100.0,
		DestinationId:  "NAT",
		Weight:         10.0,
		ActionsId:      "TEST_ACTIONS",
	}
	as, err := storageGetter.GetActions(at.ActionsId)
	if err != nil {
		t.Error("Error getting actions for the action timing: ", err)
	}
	storageGetter.LogActionTrigger("rif", RATER_SOURCE, at, as)
	//expected := "rif*some_uuid;MONETARY;OUT;NAT;TEST_ACTIONS;100;10;false*|TOPUP|MONETARY|OUT|10|0"
	var key string
	atMap, _ := storageGetter.GetAllActionTimings()
	for k, v := range atMap {
		_ = k
		_ = v
		/*if strings.Contains(k, LOG_ACTION_TRIGGER_PREFIX) && strings.Contains(v, expected) {
			key = k
			break
		}*/
	}
	if key != "" {
		t.Error("Action timing was not logged")
	}
}

func TestActionTimingLogging(t *testing.T) {
	i := &Interval{
		Months:     Months{time.January, time.February, time.March, time.April, time.May, time.June, time.July, time.August, time.September, time.October, time.November, time.December},
		MonthDays:  MonthDays{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31},
		WeekDays:   WeekDays{time.Monday, time.Tuesday, time.Wednesday, time.Thursday, time.Friday},
		StartTime:  "18:00:00",
		EndTime:    "00:00:00",
		Weight:     10.0,
		ConnectFee: 0.0,
		Prices:     PriceGroups{&Price{0, 1.0, 1 * time.Second, 60 * time.Second}},
	}
	at := &ActionTiming{
		Id:             "some uuid",
		Tag:            "test",
		UserBalanceIds: []string{"one", "two", "three"},
		Timing:         i,
		Weight:         10.0,
		ActionsId:      "TEST_ACTIONS",
	}
	as, err := storageGetter.GetActions(at.ActionsId)
	if err != nil {
		t.Error("Error getting actions for the action trigger: ", err)
	}
	storageGetter.LogActionTiming(SCHED_SOURCE, at, as)
	//expected := "some uuid|test|one,two,three|;1,2,3,4,5,6,7,8,9,10,11,12;1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31;1,2,3,4,5;18:00:00;00:00:00;10;0;1;60;1|10|TEST_ACTIONS*|TOPUP|MONETARY|OUT|10|0"
	var key string
	atMap, _ := storageGetter.GetAllActionTimings()
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
	a := &Action{Units: 10}
	genericMakeNegative(a)
	if a.Units > 0 {
		t.Error("Failed to make negative: ", a)
	}
	genericMakeNegative(a)
	if a.Units > 0 {
		t.Error("Failed to preserve negative: ", a)
	}
}

/********************************** Benchmarks ********************************/

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
