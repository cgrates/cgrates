/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2012  Radu Ioan Fericean

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

package timespans

import (
	"reflect"
	"testing"
	"time"
)

func init() {
	storageGetter, _ = NewRedisStorage("tcp:127.0.0.1:6379", 10)
	SetStorageGetter(storageGetter)
	populateTestActions()
}

func populateTestActions() {
	/*ats := []*Action{
		&Action{ActionType: "TOPUP", BalanceId: CREDIT, Units: 10},
		&Action{ActionType: "TOPUP", BalanceId: MINUTES, MinuteBucket: &MinuteBucket{Seconds: 10, DestinationId: "NAT"}},
	}
	storageGetter.SetActions("TEST_ACTIONS", ats)
	ats1 := []*Action{
		&Action{ActionType: "TOPUP", BalanceId: CREDIT, Units: 10, Weight: 20},
		&Action{ActionType: "RESET_PREPAID", Weight: 10},
	}
	storageGetter.SetActions("TEST_ACTIONS_ORDER", ats1)*/
}

func TestActionTimingStoreRestore(t *testing.T) {
	i := &Interval{
		Months:      Months{time.January, time.February, time.March, time.April, time.May, time.June, time.July, time.August, time.September, time.October, time.November, time.December},
		MonthDays:   MonthDays{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31},
		WeekDays:    WeekDays{time.Monday, time.Tuesday, time.Wednesday, time.Thursday, time.Friday},
		StartTime:   "18:00:00",
		EndTime:     "00:00:00",
		Weight:      10.0,
		ConnectFee:  0.0,
		Price:       1.0,
		BillingUnit: 1.0,
	}
	at := &ActionTiming{
		Tag:            "test",
		UserBalanceIds: []string{"one", "two", "three"},
		Timing:         i,
		Weight:         10.0,
		ActionsId:      "Commando",
	}
	r := at.store()
	if string(r) != "test|one,two,three|1,2,3,4,5,6,7,8,9,10,11,12;1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31;1,2,3,4,5;18:00:00;00:00:00;10;0;1;1|10|Commando" {
		t.Errorf("Error serializing action timing: %v", string(r))
	}
	o := &ActionTiming{}
	o.restore(r)
	if !reflect.DeepEqual(o, at) {
		t.Errorf("Expected %v was  %v", at, o)
	}
}

func TestActionTriggerStoreRestore(t *testing.T) {
	at := &ActionTrigger{
		BalanceId:      CREDIT,
		ThresholdValue: 100.0,
		DestinationId:  "NAT",
		Weight:         10.0,
		ActionsId:      "Commando",
	}
	r := at.store()
	if string(r) != "MONETARY;NAT;Commando;100;10;false" {
		t.Errorf("Error serializing action trigger: %v", string(r))
	}
	o := &ActionTrigger{}
	o.restore(r)
	if !reflect.DeepEqual(o, at) {
		t.Errorf("Expected %v was  %v", at, o)
	}
}

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
		e = time.Date(e.Year(), e.Month(), day+i, e.Hour(), e.Minute(), e.Second(), e.Nanosecond(), e.Location())
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
		e = time.Date(e.Year(), e.Month(), day+i, e.Hour(), e.Minute(), e.Second(), e.Nanosecond(), e.Location())
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
	tomorrow := time.Date(y, m, d+1, 0, 0, 0, 0, time.Local)
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
	tomorrow := time.Date(y, m, d+1, 0, 0, 0, 0, time.Local)
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
	nextMonth := time.Date(y, m+1, d, 0, 0, 0, 0, time.Local)
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
	nextMonth := time.Date(y, m+1, d, 0, 0, 0, 0, time.Local)
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
	nextMonth := time.Date(y, m+1, d, 0, 0, 0, 0, time.Local)
	tomorrow := time.Date(y, m, d+1, 0, 0, 0, 0, time.Local)
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

func TestActionTimingFisrtOfTheMonth(t *testing.T) {
	now := time.Now()
	y, m, _ := now.Date()
	nextMonth := time.Date(y, m+1, 1, 0, 0, 0, 0, time.Local)
	at := &ActionTiming{Timing: &Interval{
		Months:    Months{time.January, time.February, time.March, time.April, time.May, time.June, time.July, time.August, time.September, time.October, time.November, time.December},
		MonthDays: MonthDays{1},
		StartTime: "00:00:00",
	}}
	st := at.GetNextStartTime()
	expected := nextMonth
	if !st.Equal(expected) {
		t.Errorf("Expected %v was %v", expected, st)
	}
}

func TestActionTimingIsOneTimeRunNoInterval(t *testing.T) {
	at := &ActionTiming{}
	if !at.IsOneTimeRun() {
		t.Errorf("%v should be one time run!", at)
	}
}

func TestActionTimingIsOneTimeRunNothing(t *testing.T) {
	at := &ActionTiming{Timing: &Interval{}}
	if !at.IsOneTimeRun() {
		t.Errorf("%v should be one time run!", at)
	}
}

func TestActionTimingIsOneTimeRunStartTime(t *testing.T) {
	at := &ActionTiming{Timing: &Interval{
		StartTime: "00:00:00",
	}}
	if !at.IsOneTimeRun() {
		t.Errorf("%v should be one time run!", at)
	}
}

func TestActionTimingIsOneTimeRunWeekDay(t *testing.T) {
	at := &ActionTiming{Timing: &Interval{
		WeekDays: WeekDays{time.Monday},
	}}
	if at.IsOneTimeRun() {
		t.Errorf("%v should NOT be one time run!", at)
	}
}

func TestActionTimingLogFunction(t *testing.T) {
	a := &Action{
		ActionType:   "LOG",
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

func TestActionTimingPriotityList(t *testing.T) {
	at1 := &ActionTiming{Timing: &Interval{
		Months:    Months{time.January, time.February, time.March, time.April, time.May, time.June, time.July, time.August, time.September, time.October, time.November, time.December},
		MonthDays: MonthDays{1},
		StartTime: "00:00:00",
		Weight:    20,
	}}
	at2 := &ActionTiming{Timing: &Interval{
		Months:    Months{time.January, time.February, time.March, time.April, time.May, time.June, time.July, time.August, time.September, time.October, time.November, time.December},
		MonthDays: MonthDays{2},
		StartTime: "00:00:00",
		Weight:    10,
	}}
	var atpl ActionTimingPriotityList
	atpl = append(atpl, at2, at1)
	t.Log(atpl)
	atpl.Sort()
	if atpl[0] != at1 || atpl[1] != at2 {
		t.Error("Timing list not sorted correctly: ", atpl)
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
