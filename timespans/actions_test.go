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
	"testing"
	"time"
)

func TestActionTimingNothing(t *testing.T) {
	at := &ActionTiming{}
	st, err := at.GetNextStartTime()
	if err != nil {
		t.Error(err)
	}
	expected := time.Date(1, 1, 1, 0, 0, 0, 0, time.UTC)
	if !st.Equal(expected) {
		t.Errorf("Expected %v was %v", expected, st)
	}
}

func TestActionTimingOnlyHour(t *testing.T) {
	at := &ActionTiming{Timing: &Interval{StartTime: "10:01:00"}}
	st, err := at.GetNextStartTime()
	if err != nil {
		t.Error(err)
	}
	now := time.Now()
	y, m, d := now.Date()
	expected := time.Date(y, m, d, 10, 1, 0, 0, time.Local)
	if !st.Equal(expected) {
		t.Errorf("Expected %v was %v", expected, st)
	}
}

/*func TestActionTimingOnlyWeekdays(t *testing.T) {
	at := &ActionTiming{Timing: &Interval{WeekDays: []time.Weekday{time.Monday}}}
	st, err := at.GetNextStartTime()
	if err != nil {
		t.Error(err)
	}
	now := time.Now()
	y, m, d := now.Date()
	expected := time.Date(y, m, d, 10, 1, 0, 0, time.Local)
	if !st.Equal(expected) {
		t.Errorf("Expected %v was %v", expected, st)
	}
}

func TestActionTimingWeekdaysHour(t *testing.T) {
	at := &ActionTiming{Timing: &Interval{WeekDays: []time.Weekday{time.Monday}, StartTime: "10:01:00"}}
	st, err := at.GetNextStartTime()
	if err != nil {
		t.Error(err)
	}
	now := time.Now()
	y, m, d := now.Date()
	expected := time.Date(y, m, d, 10, 1, 0, 0, time.Local)
	if !st.Equal(expected) {
		t.Errorf("Expected %v was %v", expected, st)
	}
}*/

func TestActionTimingOnlyMonthdays(t *testing.T) {
	now := time.Now()
	y, m, d := now.Date()
	tomorrow := time.Date(y, m, d+1, 0, 0, 0, 0, time.Local)
	at := &ActionTiming{Timing: &Interval{MonthDays: MonthDays{1, 25, 2, tomorrow.Day()}}}
	st, err := at.GetNextStartTime()
	if err != nil {
		t.Error(err)
	}
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
	st, err := at.GetNextStartTime()
	if err != nil {
		t.Error(err)
	}
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
	st, err := at.GetNextStartTime()
	if err != nil {
		t.Error(err)
	}
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
	st, err := at.GetNextStartTime()
	if err != nil {
		t.Error(err)
	}
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
	month := now.Month()
	if now.After(testTime) {
		month = nextMonth.Month()
	}
	at := &ActionTiming{Timing: &Interval{
		Months:    Months{now.Month(), nextMonth.Month()},
		MonthDays: MonthDays{now.Day(), tomorrow.Day()},
		StartTime: "10:01:00",
	}}
	st, err := at.GetNextStartTime()
	if err != nil {
		t.Error(err)
	}
	expected := time.Date(y, month, day, 10, 1, 0, 0, time.Local)
	if !st.Equal(expected) {
		t.Errorf("Expected %v was %v", expected, st)
	}
}
