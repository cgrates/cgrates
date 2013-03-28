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

package rater

import (
	"reflect"
	"testing"
	"time"
)

func TestIntervalStoreRestore(t *testing.T) {
	i := &Interval{
		Months:         Months{time.January, time.February, time.March, time.April, time.May, time.June, time.July, time.August, time.September, time.October, time.November, time.December},
		MonthDays:      MonthDays{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31},
		WeekDays:       WeekDays{time.Monday, time.Tuesday, time.Wednesday, time.Thursday, time.Friday},
		StartTime:      "18:00:00",
		EndTime:        "00:00:00",
		Weight:         10.0,
		ConnectFee:     0.0,
		Price:          1.0,
		PricedUnits:    60,
		RateIncrements: 1,
	}
	r := i.store()
	if string(r) != ";1,2,3,4,5,6,7,8,9,10,11,12;1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31;1,2,3,4,5;18:00:00;00:00:00;10;0;1;60;1" {
		t.Errorf("Error serializing interval: %v", string(r))
	}
	o := &Interval{}
	o.restore(r)
	if !reflect.DeepEqual(o, i) {
		t.Errorf("Expected %v was  %v", i, o)
	}
}

func TestIntervalRestoreFromString(t *testing.T) {
	s := ";1,2,3,4,5,6,7,8,9,10,11,12;1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31;1,2,3,4,5,6,0;00:00:00;;10;0;0.2;60;1"
	i := Interval{}
	i.restore(s)
	if i.Price != 0.2 {
		t.Errorf("Error restoring inteval period from string %+v", i)
	}
}

func TestIntervalMonth(t *testing.T) {
	i := &Interval{Months: Months{time.February}}
	d := time.Date(2012, time.February, 10, 23, 0, 0, 0, time.UTC)
	d1 := time.Date(2012, time.January, 10, 23, 0, 0, 0, time.UTC)
	if !i.Contains(d) {
		t.Errorf("Date %v shoud be in interval %v", d, i)
	}
	if i.Contains(d1) {
		t.Errorf("Date %v shoud not be in interval %v", d1, i)
	}
}

func TestIntervalMonthDay(t *testing.T) {
	i := &Interval{MonthDays: MonthDays{10}}
	d := time.Date(2012, time.February, 10, 23, 0, 0, 0, time.UTC)
	d1 := time.Date(2012, time.February, 11, 23, 0, 0, 0, time.UTC)
	if !i.Contains(d) {
		t.Errorf("Date %v shoud be in interval %v", d, i)
	}
	if i.Contains(d1) {
		t.Errorf("Date %v shoud not be in interval %v", d1, i)
	}
}

func TestIntervalMonthAndMonthDay(t *testing.T) {
	i := &Interval{Months: Months{time.February}, MonthDays: MonthDays{10}}
	d := time.Date(2012, time.February, 10, 23, 0, 0, 0, time.UTC)
	d1 := time.Date(2012, time.February, 11, 23, 0, 0, 0, time.UTC)
	d2 := time.Date(2012, time.January, 10, 23, 0, 0, 0, time.UTC)
	if !i.Contains(d) {
		t.Errorf("Date %v shoud be in interval %v", d, i)
	}
	if i.Contains(d1) {
		t.Errorf("Date %v shoud not be in interval %v", d1, i)
	}
	if i.Contains(d2) {
		t.Errorf("Date %v shoud not be in interval %v", d2, i)
	}
}

func TestIntervalWeekDays(t *testing.T) {
	i := &Interval{WeekDays: []time.Weekday{time.Wednesday}}
	i2 := &Interval{WeekDays: []time.Weekday{time.Wednesday, time.Thursday}}
	d := time.Date(2012, time.February, 1, 23, 0, 0, 0, time.UTC)
	d1 := time.Date(2012, time.February, 2, 23, 0, 0, 0, time.UTC)
	if !i.Contains(d) {
		t.Errorf("Date %v shoud be in interval %v", d, i)
	}
	if i.Contains(d1) {
		t.Errorf("Date %v shoud not be in interval %v", d1, i)
	}
	if !i2.Contains(d) {
		t.Errorf("Date %v shoud be in interval %v", d, i2)
	}
	if !i2.Contains(d1) {
		t.Errorf("Date %v shoud be in interval %v", d1, i2)
	}
}

func TestIntervalMonthAndMonthDayAndWeekDays(t *testing.T) {
	i := &Interval{Months: Months{time.February}, MonthDays: MonthDays{1}, WeekDays: []time.Weekday{time.Wednesday}}
	i2 := &Interval{Months: Months{time.February}, MonthDays: MonthDays{2}, WeekDays: []time.Weekday{time.Wednesday, time.Thursday}}
	d := time.Date(2012, time.February, 1, 23, 0, 0, 0, time.UTC)
	d1 := time.Date(2012, time.February, 2, 23, 0, 0, 0, time.UTC)
	if !i.Contains(d) {
		t.Errorf("Date %v shoud be in interval %v", d, i)
	}
	if i.Contains(d1) {
		t.Errorf("Date %v shoud not be in interval %v", d1, i)
	}
	if i2.Contains(d) {
		t.Errorf("Date %v shoud be in interval %v", d, i2)
	}
	if !i2.Contains(d1) {
		t.Errorf("Date %v shoud be in interval %v", d1, i2)
	}
}

func TestIntervalHours(t *testing.T) {
	i := &Interval{StartTime: "14:30:00", EndTime: "15:00:00"}
	d := time.Date(2012, time.February, 10, 14, 30, 1, 0, time.UTC)
	d1 := time.Date(2012, time.January, 10, 14, 29, 0, 0, time.UTC)
	d2 := time.Date(2012, time.January, 10, 14, 59, 0, 0, time.UTC)
	d3 := time.Date(2012, time.January, 10, 15, 01, 0, 0, time.UTC)
	if !i.Contains(d) {
		t.Errorf("Date %v shoud be in interval %v", d, i)
	}
	if i.Contains(d1) {
		t.Errorf("Date %v shoud not be in interval %v", d1, i)
	}
	if !i.Contains(d2) {
		t.Errorf("Date %v shoud be in interval %v", d2, i)
	}
	if i.Contains(d3) {
		t.Errorf("Date %v shoud not be in interval %v", d3, i)
	}
}

func TestIntervalEverything(t *testing.T) {
	i := &Interval{Months: Months{time.February},
		Years:     Years{2012},
		MonthDays: MonthDays{1},
		WeekDays:  []time.Weekday{time.Wednesday, time.Thursday},
		StartTime: "14:30:00",
		EndTime:   "15:00:00"}
	d := time.Date(2012, time.February, 1, 14, 30, 1, 0, time.UTC)
	d1 := time.Date(2012, time.February, 1, 14, 29, 1, 0, time.UTC)
	d2 := time.Date(2012, time.February, 1, 15, 00, 00, 0, time.UTC)
	d3 := time.Date(2012, time.February, 1, 15, 0, 1, 0, time.UTC)
	d4 := time.Date(2011, time.February, 1, 15, 00, 00, 0, time.UTC)
	if !i.Contains(d) {
		t.Errorf("Date %v shoud be in interval %v", d, i)
	}
	if i.Contains(d1) {
		t.Errorf("Date %v shoud not be in interval %v", d1, i)
	}
	if !i.Contains(d2) {
		t.Errorf("Date %v shoud be in interval %v", d2, i)
	}
	if i.Contains(d3) {
		t.Errorf("Date %v shoud not be in interval %v", d3, i)
	}
	if i.Contains(d4) {
		t.Errorf("Date %v shoud not be in interval %v", d3, i)
	}
}

func TestIntervalEqual(t *testing.T) {
	i1 := &Interval{Months: Months{time.February},
		MonthDays: MonthDays{1},
		WeekDays:  []time.Weekday{time.Wednesday, time.Thursday},
		StartTime: "14:30:00",
		EndTime:   "15:00:00"}
	i2 := &Interval{Months: Months{time.February},
		MonthDays: MonthDays{1},
		WeekDays:  []time.Weekday{time.Wednesday, time.Thursday},
		StartTime: "14:30:00",
		EndTime:   "15:00:00"}
	if !i1.Equal(i2) || !i2.Equal(i1) {
		t.Errorf("%v and %v are not equal", i1, i2)
	}
}

func TestIntervalNotEqual(t *testing.T) {
	i1 := &Interval{Months: Months{time.February},
		MonthDays: MonthDays{1},
		WeekDays:  []time.Weekday{time.Wednesday},
		StartTime: "14:30:00",
		EndTime:   "15:00:00"}
	i2 := &Interval{Months: Months{time.February},
		MonthDays: MonthDays{1},
		WeekDays:  []time.Weekday{time.Wednesday, time.Thursday},
		StartTime: "14:30:00",
		EndTime:   "15:00:00"}
	if i1.Equal(i2) || i2.Equal(i1) {
		t.Errorf("%v and %v not equal", i1, i2)
	}
}

func BenchmarkIntervalContainsDate(b *testing.B) {
	i := &Interval{Months: Months{time.February}, MonthDays: MonthDays{1}, WeekDays: []time.Weekday{time.Wednesday, time.Thursday}, StartTime: "14:30:00", EndTime: "15:00:00"}
	d := time.Date(2012, time.February, 1, 14, 30, 0, 0, time.UTC)
	for x := 0; x < b.N; x++ {
		i.Contains(d)
	}
}
