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

func TestMonth(t *testing.T) {
	i := &Interval{Month: time.February}
	d := time.Date(2012, time.February, 10, 23, 0, 0, 0, time.UTC)
	d1 := time.Date(2012, time.January, 10, 23, 0, 0, 0, time.UTC)
	if !i.Contains(d) {
		t.Errorf("Date %v shoud be in interval %v", d, i)
	}
	if i.Contains(d1) {
		t.Errorf("Date %v shoud not be in interval %v", d1, i)
	}
}

func TestMonthDay(t *testing.T) {
	i := &Interval{MonthDay: 10}
	d := time.Date(2012, time.February, 10, 23, 0, 0, 0, time.UTC)
	d1 := time.Date(2012, time.February, 11, 23, 0, 0, 0, time.UTC)
	if !i.Contains(d) {
		t.Errorf("Date %v shoud be in interval %v", d, i)
	}
	if i.Contains(d1) {
		t.Errorf("Date %v shoud not be in interval %v", d1, i)
	}
}

func TestMonthAndMonthDay(t *testing.T) {
	i := &Interval{Month: time.February, MonthDay: 10}
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

func TestWeekDays(t *testing.T) {
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

func TestMonthAndMonthDayAndWeekDays(t *testing.T) {
	i := &Interval{Month: time.February, MonthDay: 1, WeekDays: []time.Weekday{time.Wednesday}}
	i2 := &Interval{Month: time.February, MonthDay: 2, WeekDays: []time.Weekday{time.Wednesday, time.Thursday}}
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

func TestHours(t *testing.T) {
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

func TestEverything(t *testing.T) {
	i := &Interval{Month: time.February,
		MonthDay:  1,
		WeekDays:  []time.Weekday{time.Wednesday, time.Thursday},
		StartTime: "14:30:00",
		EndTime:   "15:00:00"}
	d := time.Date(2012, time.February, 1, 14, 30, 1, 0, time.UTC)
	d1 := time.Date(2012, time.February, 1, 14, 29, 1, 0, time.UTC)
	d2 := time.Date(2012, time.February, 1, 15, 00, 00, 0, time.UTC)
	d3 := time.Date(2012, time.February, 1, 15, 0, 1, 0, time.UTC)
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

func BenchmarkIntervalContainsDate(b *testing.B) {
	i := &Interval{Month: time.February, MonthDay: 1, WeekDays: []time.Weekday{time.Wednesday, time.Thursday}, StartTime: "14:30:00", EndTime: "15:00:00"}
	d := time.Date(2012, time.February, 1, 14, 30, 0, 0, time.UTC)
	for x := 0; x < b.N; x++ {
		i.Contains(d)
	}
}
