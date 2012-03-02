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

func TestRightMargin(t *testing.T) {
	i := &Interval{WeekDays: []time.Weekday{time.Monday, time.Tuesday, time.Wednesday, time.Thursday, time.Friday}}
	t1 := time.Date(2012, time.February, 3, 23, 45, 0, 0, time.UTC)
	t2 := time.Date(2012, time.February, 4, 0, 10, 0, 0, time.UTC)
	ts := &TimeSpan{TimeStart: t1, TimeEnd: t2}
	oldDuration := ts.GetDuration()
	nts := ts.SplitByInterval(i)
	if ts.TimeStart != t1 || ts.TimeEnd != time.Date(2012, time.February, 3, 23, 59, 59, 0, time.UTC) {
		t.Error("Incorrect first half", ts)
	}
	if nts.TimeStart != time.Date(2012, time.February, 3, 23, 59, 59, 0, time.UTC) || nts.TimeEnd != t2 {
		t.Error("Incorrect second half", nts)
	}
	if ts.Interval != i {
		t.Error("Interval not attached correctly")
	}

	if ts.GetDuration().Seconds() != 15*60-1 || nts.GetDuration().Seconds() != 10*60+1 {
		t.Error("Wrong durations.for Intervals", ts.GetDuration().Seconds(), ts.GetDuration().Seconds())
	}

	if ts.GetDuration().Seconds()+nts.GetDuration().Seconds() != oldDuration.Seconds() {
		t.Errorf("The duration has changed: %v + %v != %v", ts.GetDuration().Seconds(), nts.GetDuration().Seconds(), oldDuration.Seconds())
	}
}

func TestRightHourMargin(t *testing.T) {
	i := &Interval{WeekDays: []time.Weekday{time.Monday, time.Tuesday, time.Wednesday, time.Thursday, time.Friday}, EndTime: "17:59:00"}
	t1 := time.Date(2012, time.February, 3, 17, 30, 0, 0, time.UTC)
	t2 := time.Date(2012, time.February, 3, 18, 00, 0, 0, time.UTC)
	ts := &TimeSpan{TimeStart: t1, TimeEnd: t2}
	oldDuration := ts.GetDuration()
	nts := ts.SplitByInterval(i)
	if ts.TimeStart != t1 || ts.TimeEnd != time.Date(2012, time.February, 3, 17, 59, 00, 0, time.UTC) {
		t.Error("Incorrect first half", ts)
	}
	if nts.TimeStart != time.Date(2012, time.February, 3, 17, 59, 00, 0, time.UTC) || nts.TimeEnd != t2 {
		t.Error("Incorrect second half", nts)
	}
	if ts.Interval != i {
		t.Error("Interval not attached correctly")
	}

	if ts.GetDuration().Seconds() != 29*60 || nts.GetDuration().Seconds() != 1*60 {
		t.Error("Wrong durations.for Intervals", ts.GetDuration().Seconds(), nts.GetDuration().Seconds())
	}
	if ts.GetDuration().Seconds()+nts.GetDuration().Seconds() != oldDuration.Seconds() {
		t.Errorf("The duration has changed: %v + %v != %v", ts.GetDuration().Seconds(), nts.GetDuration().Seconds(), oldDuration.Seconds())
	}
}

func TestLeftMargin(t *testing.T) {
	i := &Interval{WeekDays: []time.Weekday{time.Monday, time.Tuesday, time.Wednesday, time.Thursday, time.Friday}}
	t1 := time.Date(2012, time.February, 5, 23, 45, 0, 0, time.UTC)
	t2 := time.Date(2012, time.February, 6, 0, 10, 0, 0, time.UTC)
	ts := &TimeSpan{TimeStart: t1, TimeEnd: t2}
	oldDuration := ts.GetDuration()
	nts := ts.SplitByInterval(i)
	if ts.TimeStart != t1 || ts.TimeEnd != time.Date(2012, time.February, 6, 0, 0, 0, 0, time.UTC) {
		t.Error("Incorrect first half", ts)
	}
	if nts.TimeStart != time.Date(2012, time.February, 6, 0, 0, 0, 0, time.UTC) || nts.TimeEnd != t2 {
		t.Error("Incorrect second half", nts)
	}
	if nts.Interval != i {
		t.Error("Interval not attached correctly")
	}
	if ts.GetDuration().Seconds() != 15*60 || nts.GetDuration().Seconds() != 10*60 {
		t.Error("Wrong durations.for Intervals", ts.GetDuration().Seconds(), nts.GetDuration().Seconds())
	}
	if ts.GetDuration().Seconds()+nts.GetDuration().Seconds() != oldDuration.Seconds() {
		t.Errorf("The duration has changed: %v + %v != %v", ts.GetDuration().Seconds(), nts.GetDuration().Seconds(), oldDuration.Seconds())
	}
}

func TestLeftHourMargin(t *testing.T) {
	i := &Interval{Month: time.December, MonthDay: 1, StartTime: "09:00:00"}
	t1 := time.Date(2012, time.December, 1, 8, 45, 0, 0, time.UTC)
	t2 := time.Date(2012, time.December, 1, 9, 20, 0, 0, time.UTC)
	ts := &TimeSpan{TimeStart: t1, TimeEnd: t2}
	oldDuration := ts.GetDuration()
	nts := ts.SplitByInterval(i)
	if ts.TimeStart != t1 || ts.TimeEnd != time.Date(2012, time.December, 1, 9, 0, 0, 0, time.UTC) {
		t.Error("Incorrect first half", ts)
	}
	if nts.TimeStart != time.Date(2012, time.December, 1, 9, 0, 0, 0, time.UTC) || nts.TimeEnd != t2 {
		t.Error("Incorrect second half", nts)
	}
	if nts.Interval != i {
		t.Error("Interval not attached correctly")
	}
	if ts.GetDuration().Seconds() != 15*60 || nts.GetDuration().Seconds() != 20*60 {
		t.Error("Wrong durations.for Intervals", ts.GetDuration().Seconds(), nts.GetDuration().Seconds())
	}
	if ts.GetDuration().Seconds()+nts.GetDuration().Seconds() != oldDuration.Seconds() {
		t.Errorf("The duration has changed: %v + %v != %v", ts.GetDuration().Seconds(), nts.GetDuration().Seconds(), oldDuration.Seconds())
	}
}

func TestEnclosingMargin(t *testing.T) {
	i := &Interval{WeekDays: []time.Weekday{time.Sunday}}
	t1 := time.Date(2012, time.February, 5, 17, 45, 0, 0, time.UTC)
	t2 := time.Date(2012, time.February, 5, 18, 10, 0, 0, time.UTC)
	ts := &TimeSpan{TimeStart: t1, TimeEnd: t2}
	nts := ts.SplitByInterval(i)
	if ts.TimeStart != t1 || ts.TimeEnd != t2 || nts != nil {
		t.Error("Incorrect enclosing", ts)
	}
	if ts.Interval != i {
		t.Error("Interval not attached correctly")
	}
}

func TestOutsideMargin(t *testing.T) {
	i := &Interval{WeekDays: []time.Weekday{time.Monday}}
	t1 := time.Date(2012, time.February, 5, 17, 45, 0, 0, time.UTC)
	t2 := time.Date(2012, time.February, 5, 18, 10, 0, 0, time.UTC)
	ts := &TimeSpan{TimeStart: t1, TimeEnd: t2}
	result := ts.SplitByInterval(i)
	if result != nil {
		t.Error("Interval not split correctly")
	}
}

func TestContains(t *testing.T) {
	t1 := time.Date(2012, time.February, 5, 17, 45, 0, 0, time.UTC)
	t2 := time.Date(2012, time.February, 5, 17, 55, 0, 0, time.UTC)
	t3 := time.Date(2012, time.February, 5, 17, 50, 0, 0, time.UTC)
	ts := TimeSpan{TimeStart: t1, TimeEnd: t2}
	if ts.Contains(t1) {
		t.Error("It should NOT contain ", t1)
	}
	if ts.Contains(t2) {
		t.Error("It should NOT contain ", t1)
	}
	if !ts.Contains(t3) {
		t.Error("It should contain ", t3)
	}
}

func TestSplitByActivationTime(t *testing.T) {
	t1 := time.Date(2012, time.February, 5, 17, 45, 0, 0, time.UTC)
	t2 := time.Date(2012, time.February, 5, 17, 55, 0, 0, time.UTC)
	t3 := time.Date(2012, time.February, 5, 17, 50, 0, 0, time.UTC)
	ts := TimeSpan{TimeStart: t1, TimeEnd: t2}
	ap1 := &ActivationPeriod{ActivationTime: t1}
	ap2 := &ActivationPeriod{ActivationTime: t2}
	ap3 := &ActivationPeriod{ActivationTime: t3}

	if ts.SplitByActivationPeriod(ap1) != nil {
		t.Error("Error spliting on left margin")
	}
	if ts.SplitByActivationPeriod(ap2) != nil {
		t.Error("Error spliting on right margin")
	}
	result := ts.SplitByActivationPeriod(ap3)
	if result.TimeStart != t3 || result.TimeEnd != t2 {
		t.Error("Error spliting on interior")
	}
}

func TestTimespanGetCost(t *testing.T) {
	getter, _ := NewKyotoStorage("../data/test.kch")
	defer getter.Close()
	t1 := time.Date(2012, time.February, 5, 17, 45, 0, 0, time.UTC)
	t2 := time.Date(2012, time.February, 5, 17, 55, 0, 0, time.UTC)
	ts1 := TimeSpan{TimeStart: t1, TimeEnd: t2}
	cd := &CallDescriptor{Subject: "other", storageGetter: getter}
	if ts1.GetCost(cd) != 0 {
		t.Error("No interval and still kicking")
	}
	ts1.Interval = &Interval{Price: 1}
	if ts1.GetCost(cd) != 600 {
		t.Error("Expected 10 got ", ts1.GetCost(cd))
	}
	ts1.Interval.BillingUnit = .1
	if ts1.GetCost(cd) != 6000 {
		t.Error("Expected 6000 got ", ts1.GetCost(cd))
	}
}

func TestSetInterval(t *testing.T) {
	i1 := &Interval{Price: 1}
	ts1 := TimeSpan{Interval: i1}
	i2 := &Interval{Price: 2}
	ts1.SetInterval(i2)
	if ts1.Interval != i1 {
		t.Error("Smaller price interval should win")
	}
	i2.Ponder = 1
	ts1.SetInterval(i2)
	if ts1.Interval != i2 {
		t.Error("Bigger ponder interval should win")
	}
}
