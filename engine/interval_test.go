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
	"testing"
	"time"
)

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
