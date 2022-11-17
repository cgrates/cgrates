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
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/utils"
)

func TestRightMargin(t *testing.T) {
	i := &RateInterval{
		Timing: &RITiming{WeekDays: []time.Weekday{time.Monday, time.Tuesday, time.Wednesday, time.Thursday, time.Friday}}}
	t1 := time.Date(2012, time.February, 3, 23, 45, 0, 0, time.UTC)
	t2 := time.Date(2012, time.February, 4, 0, 10, 0, 0, time.UTC)
	ts := &TimeSpan{TimeStart: t1, TimeEnd: t2, ratingInfo: &RatingInfo{}}
	oldDuration := ts.GetDuration()
	nts := ts.SplitByRateInterval(i, false)
	if ts.TimeStart != t1 || ts.TimeEnd != time.Date(2012, time.February, 3, 24, 0, 0, 0, time.UTC) {
		t.Error("Incorrect first half", ts)
	}
	if nts.TimeStart != time.Date(2012, time.February, 4, 0, 0, 0, 0, time.UTC) || nts.TimeEnd != t2 {
		t.Error("Incorrect second half", nts)
	}
	if ts.RateInterval != i {
		t.Error("RateInterval not attached correctly")
	}

	if ts.GetDuration() != 15*time.Minute || nts.GetDuration() != 10*time.Minute {
		t.Error("Wrong durations.for RateIntervals", ts.GetDuration(), ts.GetDuration())
	}

	if ts.GetDuration().Seconds()+nts.GetDuration().Seconds() != oldDuration.Seconds() {
		t.Errorf("The duration has changed: %v + %v != %v", ts.GetDuration().Seconds(), nts.GetDuration().Seconds(), oldDuration.Seconds())
	}
}

func TestSplitMiddle(t *testing.T) {
	i := &RateInterval{
		Timing: &RITiming{
			WeekDays:  utils.WeekDays{time.Monday, time.Tuesday, time.Wednesday, time.Thursday, time.Friday},
			StartTime: "18:00:00",
			EndTime:   "",
		}}
	ts := &TimeSpan{
		TimeStart:  time.Date(2012, 2, 27, 0, 0, 0, 0, time.UTC),
		TimeEnd:    time.Date(2012, 2, 28, 0, 0, 0, 0, time.UTC),
		ratingInfo: &RatingInfo{},
	}

	if !i.Contains(ts.TimeEnd, true) {
		t.Errorf("%+v should contain %+v", i, ts.TimeEnd)
	}

	newTs := ts.SplitByRateInterval(i, false)
	if newTs == nil {
		t.Errorf("Error spliting interval %+v", newTs)
	}
}

func TestRightHourMargin(t *testing.T) {
	i := &RateInterval{Timing: &RITiming{WeekDays: []time.Weekday{time.Monday, time.Tuesday, time.Wednesday, time.Thursday, time.Friday}, EndTime: "17:59:00"}}
	t1 := time.Date(2012, time.February, 3, 17, 30, 0, 0, time.UTC)
	t2 := time.Date(2012, time.February, 3, 18, 00, 0, 0, time.UTC)
	ts := &TimeSpan{TimeStart: t1, TimeEnd: t2, ratingInfo: &RatingInfo{}}
	oldDuration := ts.GetDuration()
	nts := ts.SplitByRateInterval(i, false)
	if ts.TimeStart != t1 || ts.TimeEnd != time.Date(2012, time.February, 3, 17, 59, 0, 0, time.UTC) {
		t.Error("Incorrect first half", ts)
	}
	if nts.TimeStart != time.Date(2012, time.February, 3, 17, 59, 0, 0, time.UTC) || nts.TimeEnd != t2 {
		t.Error("Incorrect second half", nts)
	}
	if ts.RateInterval != i {
		t.Error("RateInterval not attached correctly")
	}

	if ts.GetDuration() != 29*time.Minute || nts.GetDuration() != time.Minute {
		t.Error("Wrong durations.for RateIntervals", ts.GetDuration(), nts.GetDuration())
	}
	if ts.GetDuration().Seconds()+nts.GetDuration().Seconds() != oldDuration.Seconds() {
		t.Errorf("The duration has changed: %v + %v != %v", ts.GetDuration().Seconds(), nts.GetDuration().Seconds(), oldDuration.Seconds())
	}
}

func TestLeftMargin(t *testing.T) {
	i := &RateInterval{Timing: &RITiming{WeekDays: []time.Weekday{time.Monday, time.Tuesday, time.Wednesday, time.Thursday, time.Friday}}}
	t1 := time.Date(2012, time.February, 5, 23, 45, 0, 0, time.UTC)
	t2 := time.Date(2012, time.February, 6, 0, 10, 0, 0, time.UTC)
	ts := &TimeSpan{TimeStart: t1, TimeEnd: t2, ratingInfo: &RatingInfo{}}
	oldDuration := ts.GetDuration()
	nts := ts.SplitByRateInterval(i, false)
	if ts.TimeStart != t1 || ts.TimeEnd != time.Date(2012, time.February, 6, 0, 0, 0, 0, time.UTC) {
		t.Error("Incorrect first half", ts)
	}
	if nts.TimeStart != time.Date(2012, time.February, 6, 0, 0, 0, 0, time.UTC) || nts.TimeEnd != t2 {
		t.Error("Incorrect second half", nts)
	}
	if nts.RateInterval != i {
		t.Error("RateInterval not attached correctly")
	}
	if ts.GetDuration().Seconds() != 15*60 || nts.GetDuration().Seconds() != 10*60 {
		t.Error("Wrong durations.for RateIntervals", ts.GetDuration().Seconds(), nts.GetDuration().Seconds())
	}
	if ts.GetDuration().Seconds()+nts.GetDuration().Seconds() != oldDuration.Seconds() {
		t.Errorf("The duration has changed: %v + %v != %v", ts.GetDuration().Seconds(), nts.GetDuration().Seconds(), oldDuration.Seconds())
	}
}

func TestLeftHourMargin(t *testing.T) {
	i := &RateInterval{Timing: &RITiming{Months: utils.Months{time.December}, MonthDays: utils.MonthDays{1}, StartTime: "09:00:00"}}
	t1 := time.Date(2012, time.December, 1, 8, 45, 0, 0, time.UTC)
	t2 := time.Date(2012, time.December, 1, 9, 20, 0, 0, time.UTC)
	ts := &TimeSpan{TimeStart: t1, TimeEnd: t2, ratingInfo: &RatingInfo{}}
	oldDuration := ts.GetDuration()
	nts := ts.SplitByRateInterval(i, false)
	if ts.TimeStart != t1 || ts.TimeEnd != time.Date(2012, time.December, 1, 9, 0, 0, 0, time.UTC) {
		t.Error("Incorrect first half", ts)
	}
	if nts.TimeStart != time.Date(2012, time.December, 1, 9, 0, 0, 0, time.UTC) || nts.TimeEnd != t2 {
		t.Error("Incorrect second half", nts)
	}
	if nts.RateInterval != i {
		t.Error("RateInterval not attached correctly")
	}
	if ts.GetDuration().Seconds() != 15*60 || nts.GetDuration().Seconds() != 20*60 {
		t.Error("Wrong durations.for RateIntervals", ts.GetDuration().Seconds(), nts.GetDuration().Seconds())
	}
	if ts.GetDuration().Seconds()+nts.GetDuration().Seconds() != oldDuration.Seconds() {
		t.Errorf("The duration has changed: %v + %v != %v", ts.GetDuration().Seconds(), nts.GetDuration().Seconds(), oldDuration.Seconds())
	}
}

func TestEnclosingMargin(t *testing.T) {
	i := &RateInterval{Timing: &RITiming{WeekDays: []time.Weekday{time.Sunday}}}
	t1 := time.Date(2012, time.February, 5, 17, 45, 0, 0, time.UTC)
	t2 := time.Date(2012, time.February, 5, 18, 10, 0, 0, time.UTC)
	ts := &TimeSpan{TimeStart: t1, TimeEnd: t2}
	nts := ts.SplitByRateInterval(i, false)
	if ts.TimeStart != t1 || ts.TimeEnd != t2 || nts != nil {
		t.Error("Incorrect enclosing", ts)
	}
	if ts.RateInterval != i {
		t.Error("RateInterval not attached correctly")
	}
}

func TestOutsideMargin(t *testing.T) {
	i := &RateInterval{Timing: &RITiming{WeekDays: []time.Weekday{time.Monday}}}
	t1 := time.Date(2012, time.February, 5, 17, 45, 0, 0, time.UTC)
	t2 := time.Date(2012, time.February, 5, 18, 10, 0, 0, time.UTC)
	ts := &TimeSpan{TimeStart: t1, TimeEnd: t2}
	result := ts.SplitByRateInterval(i, false)
	if result != nil {
		t.Error("RateInterval not split correctly")
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

func TestSplitByRatingPlan(t *testing.T) {
	t1 := time.Date(2012, time.February, 5, 17, 45, 0, 0, time.UTC)
	t2 := time.Date(2012, time.February, 5, 17, 55, 0, 0, time.UTC)
	t3 := time.Date(2012, time.February, 5, 17, 50, 0, 0, time.UTC)
	ts := TimeSpan{TimeStart: t1, TimeEnd: t2, ratingInfo: &RatingInfo{}}
	ap1 := &RatingInfo{ActivationTime: t1}
	ap2 := &RatingInfo{ActivationTime: t2}
	ap3 := &RatingInfo{ActivationTime: t3}

	if ts.SplitByRatingPlan(ap1) != nil {
		t.Error("Error spliting on left margin")
	}
	if ts.SplitByRatingPlan(ap2) != nil {
		t.Error("Error spliting on right margin")
	}
	result := ts.SplitByRatingPlan(ap3)
	if result.TimeStart != t3 || result.TimeEnd != t2 {
		t.Error("Error spliting on interior")
	}
}

func TestTimespanGetCost(t *testing.T) {
	t1 := time.Date(2012, time.February, 5, 17, 45, 0, 0, time.UTC)
	t2 := time.Date(2012, time.February, 5, 17, 55, 0, 0, time.UTC)
	ts1 := TimeSpan{TimeStart: t1, TimeEnd: t2}
	if ts1.CalculateCost() != 0 {
		t.Error("No interval and still kicking")
	}
	ts1.SetRateInterval(
		&RateInterval{
			Timing: &RITiming{},
			Rating: &RIRate{Rates: RateGroups{&RGRate{0, 1.0, time.Second, time.Second}}},
		},
	)
	if ts1.CalculateCost() != 600 {
		t.Error("Expected 10 got ", ts1.Cost)
	}
	ts1.RateInterval = nil
	ts1.SetRateInterval(&RateInterval{Rating: &RIRate{Rates: RateGroups{&RGRate{0, 1.0, time.Second, 60 * time.Second}}}})
	if ts1.CalculateCost() != 10 {
		t.Error("Expected 6000 got ", ts1.Cost)
	}
}

func TestTimespanGetCostIntervals(t *testing.T) {
	ts := &TimeSpan{}
	ts.Increments = make(Increments, 11)
	for i := 0; i < 11; i++ {
		ts.Increments[i] = &Increment{Cost: 0.02}
	}
	if ts.CalculateCost() != 0.22 {
		t.Error("Error caclulating timespan cost: ", ts.CalculateCost())
	}
}

func TestSetRateInterval(t *testing.T) {
	i1 := &RateInterval{
		Timing: &RITiming{},
		Rating: &RIRate{Rates: RateGroups{&RGRate{0, 1.0, time.Second, time.Second}}},
	}
	ts1 := TimeSpan{RateInterval: i1}
	i2 := &RateInterval{
		Timing: &RITiming{},
		Rating: &RIRate{Rates: RateGroups{&RGRate{0, 2.0, time.Second, time.Second}}},
	}
	if !ts1.hasBetterRateIntervalThan(i2) {
		ts1.SetRateInterval(i2)
	}
	if ts1.RateInterval != i1 {
		t.Error("Smaller price interval should win")
	}
	i2.Weight = 1
	ts1.SetRateInterval(i2)
	if ts1.RateInterval != i2 {
		t.Error("Bigger ponder interval should win")
	}
}

func TestTimespanSplitGroupedRates(t *testing.T) {
	i := &RateInterval{
		Timing: &RITiming{
			EndTime: "17:59:00",
		},
		Rating: &RIRate{
			Rates: RateGroups{&RGRate{0, 2, time.Second, time.Second}, &RGRate{900 * time.Second, 1, time.Second, time.Second}},
		},
	}
	t1 := time.Date(2012, time.February, 3, 17, 30, 0, 0, time.UTC)
	t2 := time.Date(2012, time.February, 3, 18, 00, 0, 0, time.UTC)
	ts := &TimeSpan{TimeStart: t1, TimeEnd: t2, DurationIndex: 1800 * time.Second, ratingInfo: &RatingInfo{}}
	oldDuration := ts.GetDuration()
	nts := ts.SplitByRateInterval(i, false)
	splitTime := time.Date(2012, time.February, 3, 17, 45, 00, 0, time.UTC)
	if ts.TimeStart != t1 || ts.TimeEnd != splitTime {
		t.Error("Incorrect first half", ts.TimeStart, ts.TimeEnd)
	}
	if nts.TimeStart != splitTime || nts.TimeEnd != t2 {
		t.Error("Incorrect second half", nts)
	}
	if ts.RateInterval != i {
		t.Error("RateInterval not attached correctly")
	}
	c1 := ts.RateInterval.GetCost(ts.GetDuration(), ts.GetGroupStart())
	c2 := nts.RateInterval.GetCost(nts.GetDuration(), nts.GetGroupStart())
	if c1 != 1800 || c2 != 900 {
		t.Error("Wrong costs: ", c1, c2)
	}

	if ts.GetDuration().Seconds() != 15*60 || nts.GetDuration().Seconds() != 15*60 {
		t.Error("Wrong durations.for RateIntervals", ts.GetDuration().Seconds(), nts.GetDuration().Seconds())
	}
	if ts.GetDuration().Seconds()+nts.GetDuration().Seconds() != oldDuration.Seconds() {
		t.Errorf("The duration has changed: %v + %v != %v", ts.GetDuration().Seconds(), nts.GetDuration().Seconds(), oldDuration.Seconds())
	}
}

func TestTimespanSplitGroupedRatesIncrements(t *testing.T) {
	i := &RateInterval{
		Timing: &RITiming{
			EndTime: "17:59:00",
		},
		Rating: &RIRate{
			Rates: RateGroups{
				&RGRate{
					GroupIntervalStart: 0,
					Value:              2,
					RateIncrement:      time.Second,
					RateUnit:           time.Second},
				&RGRate{
					GroupIntervalStart: 30 * time.Second,
					Value:              1,
					RateIncrement:      time.Minute,
					RateUnit:           time.Second,
				}}},
	}
	t1 := time.Date(2012, time.February, 3, 17, 30, 0, 0, time.UTC)
	t2 := time.Date(2012, time.February, 3, 17, 31, 0, 0, time.UTC)
	ts := &TimeSpan{TimeStart: t1, TimeEnd: t2, DurationIndex: 60 * time.Second, ratingInfo: &RatingInfo{}}
	oldDuration := ts.GetDuration()
	nts := ts.SplitByRateInterval(i, false)
	cd := &CallDescriptor{}
	timespans := cd.roundTimeSpansToIncrement([]*TimeSpan{ts, nts})
	if len(timespans) != 2 {
		t.Error("Error rounding timespans: ", timespans)
	}
	ts = timespans[0]
	nts = timespans[1]
	splitTime := time.Date(2012, time.February, 3, 17, 30, 30, 0, time.UTC)
	if ts.TimeStart != t1 || ts.TimeEnd != splitTime {
		t.Error("Incorrect first half", ts)
	}
	t3 := time.Date(2012, time.February, 3, 17, 31, 30, 0, time.UTC)
	if nts.TimeStart != splitTime || nts.TimeEnd != t3 {
		t.Error("Incorrect second half", nts.TimeStart, nts.TimeEnd)
	}
	if ts.RateInterval != i {
		t.Error("RateInterval not attached correctly")
	}
	c1 := ts.RateInterval.GetCost(ts.GetDuration(), ts.GetGroupStart())
	c2 := nts.RateInterval.GetCost(nts.GetDuration(), nts.GetGroupStart())
	if c1 != 60 || c2 != 60 {
		t.Error("Wrong costs: ", c1, c2)
	}

	if ts.GetDuration().Seconds() != 0.5*60 || nts.GetDuration().Seconds() != 60 {
		t.Error("Wrong durations.for RateIntervals", ts.GetDuration().Seconds(), nts.GetDuration().Seconds())
	}
	if ts.GetDuration()+nts.GetDuration() != oldDuration+30*time.Second {
		t.Errorf("The duration has changed: %v + %v != %v", ts.GetDuration().Seconds(), nts.GetDuration().Seconds(), oldDuration.Seconds())
	}
}

func TestTimespanSplitRightHourMarginBeforeGroup(t *testing.T) {
	i := &RateInterval{
		Timing: &RITiming{
			EndTime: "17:00:30",
		},
		Rating: &RIRate{
			Rates: RateGroups{&RGRate{0, 2, time.Second, time.Second}, &RGRate{60 * time.Second, 1, 60 * time.Second, time.Second}},
		},
	}
	t1 := time.Date(2012, time.February, 3, 17, 00, 0, 0, time.UTC)
	t2 := time.Date(2012, time.February, 3, 17, 01, 0, 0, time.UTC)
	ts := &TimeSpan{TimeStart: t1, TimeEnd: t2, ratingInfo: &RatingInfo{}}
	oldDuration := ts.GetDuration()
	nts := ts.SplitByRateInterval(i, false)
	splitTime := time.Date(2012, time.February, 3, 17, 00, 30, 0, time.UTC)
	if ts.TimeStart != t1 || ts.TimeEnd != splitTime {
		t.Error("Incorrect first half", ts)
	}
	if nts.TimeStart != splitTime || nts.TimeEnd != t2 {
		t.Error("Incorrect second half", nts)
	}
	if ts.RateInterval != i {
		t.Error("RateInterval not attached correctly")
	}

	if ts.GetDuration().Seconds() != 30 || nts.GetDuration().Seconds() != 30 {
		t.Error("Wrong durations.for RateIntervals", ts.GetDuration().Seconds(), nts.GetDuration().Seconds())
	}
	if ts.GetDuration().Seconds()+nts.GetDuration().Seconds() != oldDuration.Seconds() {
		t.Errorf("The duration has changed: %v + %v != %v", ts.GetDuration().Seconds(), nts.GetDuration().Seconds(), oldDuration.Seconds())
	}
	nnts := nts.SplitByRateInterval(i, false)
	if nnts != nil {
		t.Error("Bad new split", nnts)
	}
}

func TestTimespanSplitGroupSecondSplit(t *testing.T) {
	i := &RateInterval{
		Timing: &RITiming{
			EndTime: "17:03:30",
		},
		Rating: &RIRate{
			Rates: RateGroups{&RGRate{0, 2, time.Second, time.Second}, &RGRate{60 * time.Second, 1, time.Second, time.Second}}},
	}
	t1 := time.Date(2012, time.February, 3, 17, 00, 0, 0, time.UTC)
	t2 := time.Date(2012, time.February, 3, 17, 04, 0, 0, time.UTC)
	ts := &TimeSpan{TimeStart: t1, TimeEnd: t2, DurationIndex: 240 * time.Second, ratingInfo: &RatingInfo{}}
	oldDuration := ts.GetDuration()
	nts := ts.SplitByRateInterval(i, false)
	splitTime := time.Date(2012, time.February, 3, 17, 01, 00, 0, time.UTC)
	if ts.TimeStart != t1 || ts.TimeEnd != splitTime {
		t.Error("Incorrect first half", nts)
	}
	if nts.TimeStart != splitTime || nts.TimeEnd != t2 {
		t.Error("Incorrect second half", nts)
	}
	if ts.RateInterval != i {
		t.Error("RateInterval not attached correctly")
	}

	if ts.GetDuration().Seconds() != 60 || nts.GetDuration().Seconds() != 180 {
		t.Error("Wrong durations.for RateIntervals", ts.GetDuration().Seconds(), nts.GetDuration().Seconds())
	}
	if ts.GetDuration().Seconds()+nts.GetDuration().Seconds() != oldDuration.Seconds() {
		t.Errorf("The duration has changed: %v + %v != %v", ts.GetDuration().Seconds(), nts.GetDuration().Seconds(), oldDuration.Seconds())
	}
	nnts := nts.SplitByRateInterval(i, false)
	nsplitTime := time.Date(2012, time.February, 3, 17, 03, 30, 0, time.UTC)
	if nts.TimeStart != splitTime || nts.TimeEnd != nsplitTime {
		t.Error("Incorrect first half", nts)
	}
	if nnts.TimeStart != nsplitTime || nnts.TimeEnd != t2 {
		t.Error("Incorrect second half", nnts)
	}
	if nts.RateInterval != i {
		t.Error("RateInterval not attached correctly")
	}

	if nts.GetDuration().Seconds() != 150 || nnts.GetDuration().Seconds() != 30 {
		t.Error("Wrong durations.for RateIntervals", nts.GetDuration().Seconds(), nnts.GetDuration().Seconds())
	}
}

func TestTimespanSplitLong(t *testing.T) {
	i := &RateInterval{
		Timing: &RITiming{
			StartTime: "18:00:00",
		},
	}
	t1 := time.Date(2013, time.October, 9, 9, 0, 0, 0, time.UTC)
	t2 := time.Date(2013, time.October, 10, 20, 0, 0, 0, time.UTC)
	ts := &TimeSpan{TimeStart: t1, TimeEnd: t2, DurationIndex: t2.Sub(t1), ratingInfo: &RatingInfo{}}
	oldDuration := ts.GetDuration()
	nts := ts.SplitByRateInterval(i, false)
	splitTime := time.Date(2013, time.October, 9, 18, 0, 0, 0, time.UTC)
	if ts.TimeStart != t1 || ts.TimeEnd != splitTime {
		t.Error("Incorrect first half", nts)
	}
	if nts.TimeStart != splitTime || nts.TimeEnd != t2 {
		t.Error("Incorrect second half", nts)
	}
	if nts.RateInterval != i {
		t.Error("RateInterval not attached correctly")
	}

	if ts.GetDuration() != 9*time.Hour || nts.GetDuration() != 26*time.Hour {
		t.Error("Wrong durations.for RateIntervals", ts.GetDuration(), nts.GetDuration())
	}
	if ts.GetDuration()+nts.GetDuration() != oldDuration {
		t.Errorf("The duration has changed: %v + %v != %v", ts.GetDuration(), nts.GetDuration(), oldDuration)
	}
}

func TestTimespanSplitMultipleGroup(t *testing.T) {
	i := &RateInterval{
		Timing: &RITiming{
			EndTime: "17:05:00",
		},
		Rating: &RIRate{
			Rates: RateGroups{&RGRate{0, 2, time.Second, time.Second}, &RGRate{60 * time.Second, 1, time.Second, time.Second}, &RGRate{180 * time.Second, 1, time.Second, time.Second}}},
	}
	t1 := time.Date(2012, time.February, 3, 17, 00, 0, 0, time.UTC)
	t2 := time.Date(2012, time.February, 3, 17, 04, 0, 0, time.UTC)
	ts := &TimeSpan{TimeStart: t1, TimeEnd: t2, DurationIndex: 240 * time.Second, ratingInfo: &RatingInfo{}}
	oldDuration := ts.GetDuration()
	nts := ts.SplitByRateInterval(i, false)
	splitTime := time.Date(2012, time.February, 3, 17, 01, 00, 0, time.UTC)
	if ts.TimeStart != t1 || ts.TimeEnd != splitTime {
		t.Error("Incorrect first half", nts)
	}
	if nts.TimeStart != splitTime || nts.TimeEnd != t2 {
		t.Error("Incorrect second half", nts)
	}
	if ts.RateInterval != i {
		t.Error("RateInterval not attached correctly")
	}

	if ts.GetDuration().Seconds() != 60 || nts.GetDuration().Seconds() != 180 {
		t.Error("Wrong durations.for RateIntervals", ts.GetDuration().Seconds(), nts.GetDuration().Seconds())
	}
	if ts.GetDuration().Seconds()+nts.GetDuration().Seconds() != oldDuration.Seconds() {
		t.Errorf("The duration has changed: %v + %v != %v", ts.GetDuration().Seconds(), nts.GetDuration().Seconds(), oldDuration.Seconds())
	}
	nnts := nts.SplitByRateInterval(i, false)
	nsplitTime := time.Date(2012, time.February, 3, 17, 03, 00, 0, time.UTC)
	if nts.TimeStart != splitTime || nts.TimeEnd != nsplitTime {
		t.Error("Incorrect first half", nts)
	}
	if nnts.TimeStart != nsplitTime || nnts.TimeEnd != t2 {
		t.Error("Incorrect second half", nnts)
	}
	if nts.RateInterval != i {
		t.Error("RateInterval not attached correctly")
	}

	if nts.GetDuration().Seconds() != 120 || nnts.GetDuration().Seconds() != 60 {
		t.Error("Wrong durations.for RateIntervals", nts.GetDuration().Seconds(), nnts.GetDuration().Seconds())
	}
}

func TestTimespanExpandingPastEnd(t *testing.T) {
	timespans := []*TimeSpan{
		{
			TimeStart: time.Date(2013, 9, 10, 14, 30, 0, 0, time.UTC),
			TimeEnd:   time.Date(2013, 9, 10, 14, 30, 30, 0, time.UTC),
			RateInterval: &RateInterval{Rating: &RIRate{Rates: RateGroups{
				&RGRate{RateIncrement: 60 * time.Second},
			}}},
		},
		{
			TimeStart: time.Date(2013, 9, 10, 14, 30, 30, 0, time.UTC),
			TimeEnd:   time.Date(2013, 9, 10, 14, 30, 45, 0, time.UTC),
		},
	}
	cd := &CallDescriptor{}
	timespans = cd.roundTimeSpansToIncrement(timespans)
	if len(timespans) != 1 {
		t.Error("Error removing overlaped intervals: ", timespans)
	}
	if !timespans[0].TimeEnd.Equal(time.Date(2013, 9, 10, 14, 31, 0, 0, time.UTC)) {
		t.Errorf("Error expanding timespan: %+v", timespans[0])
	}
}

func TestTimespanExpandingDurationIndex(t *testing.T) {
	timespans := []*TimeSpan{
		{
			TimeStart: time.Date(2013, 9, 10, 14, 30, 0, 0, time.UTC),
			TimeEnd:   time.Date(2013, 9, 10, 14, 30, 30, 0, time.UTC),
			RateInterval: &RateInterval{Rating: &RIRate{Rates: RateGroups{
				&RGRate{RateIncrement: 60 * time.Second},
			}}},
		},
		{
			TimeStart: time.Date(2013, 9, 10, 14, 30, 30, 0, time.UTC),
			TimeEnd:   time.Date(2013, 9, 10, 14, 30, 45, 0, time.UTC),
		},
	}
	cd := &CallDescriptor{}
	timespans = cd.roundTimeSpansToIncrement(timespans)

	if len(timespans) != 1 || timespans[0].GetDuration() != time.Minute {
		t.Error("Error setting call duration: ", timespans[0])
	}
}

func TestTimespanExpandingRoundingPastEnd(t *testing.T) {
	timespans := []*TimeSpan{
		{
			TimeStart: time.Date(2013, 9, 10, 14, 30, 0, 0, time.UTC),
			TimeEnd:   time.Date(2013, 9, 10, 14, 30, 20, 0, time.UTC),
			RateInterval: &RateInterval{Rating: &RIRate{Rates: RateGroups{
				&RGRate{RateIncrement: 15 * time.Second},
			}}},
		},
		{
			TimeStart: time.Date(2013, 9, 10, 14, 30, 20, 0, time.UTC),
			TimeEnd:   time.Date(2013, 9, 10, 14, 30, 40, 0, time.UTC),
		},
	}
	cd := &CallDescriptor{}
	timespans = cd.roundTimeSpansToIncrement(timespans)
	if len(timespans) != 2 {
		t.Error("Error removing overlaped intervals: ", timespans[0])
	}
	if !timespans[0].TimeEnd.Equal(time.Date(2013, 9, 10, 14, 30, 30, 0, time.UTC)) {
		t.Error("Error expanding timespan: ", timespans[0])
	}
}

func TestTimespanExpandingPastEndMultiple(t *testing.T) {
	timespans := []*TimeSpan{
		{
			TimeStart: time.Date(2013, 9, 10, 14, 30, 0, 0, time.UTC),
			TimeEnd:   time.Date(2013, 9, 10, 14, 30, 30, 0, time.UTC),
			RateInterval: &RateInterval{Rating: &RIRate{Rates: RateGroups{
				&RGRate{RateIncrement: 60 * time.Second},
			}}},
		},
		{
			TimeStart: time.Date(2013, 9, 10, 14, 30, 30, 0, time.UTC),
			TimeEnd:   time.Date(2013, 9, 10, 14, 30, 40, 0, time.UTC),
		},
		{
			TimeStart: time.Date(2013, 9, 10, 14, 30, 40, 0, time.UTC),
			TimeEnd:   time.Date(2013, 9, 10, 14, 30, 50, 0, time.UTC),
		},
	}
	cd := &CallDescriptor{}
	timespans = cd.roundTimeSpansToIncrement(timespans)
	if len(timespans) != 1 {
		t.Error("Error removing overlaped intervals: ", timespans)
	}
	if !timespans[0].TimeEnd.Equal(time.Date(2013, 9, 10, 14, 31, 0, 0, time.UTC)) {
		t.Error("Error expanding timespan: ", timespans[0])
	}
}

func TestTimespanExpandingPastEndMultipleEqual(t *testing.T) {
	timespans := []*TimeSpan{
		{
			TimeStart: time.Date(2013, 9, 10, 14, 30, 0, 0, time.UTC),
			TimeEnd:   time.Date(2013, 9, 10, 14, 30, 30, 0, time.UTC),
			RateInterval: &RateInterval{Rating: &RIRate{Rates: RateGroups{
				&RGRate{RateIncrement: 60 * time.Second},
			}}},
		},
		{
			TimeStart: time.Date(2013, 9, 10, 14, 30, 30, 0, time.UTC),
			TimeEnd:   time.Date(2013, 9, 10, 14, 30, 40, 0, time.UTC),
		},
		{
			TimeStart: time.Date(2013, 9, 10, 14, 30, 40, 0, time.UTC),
			TimeEnd:   time.Date(2013, 9, 10, 14, 31, 00, 0, time.UTC),
		},
	}
	cd := &CallDescriptor{}
	timespans = cd.roundTimeSpansToIncrement(timespans)
	if len(timespans) != 1 {
		t.Error("Error removing overlaped intervals: ", timespans)
	}
	if !timespans[0].TimeEnd.Equal(time.Date(2013, 9, 10, 14, 31, 0, 0, time.UTC)) {
		t.Error("Error expanding timespan: ", timespans[0])
	}
}

func TestTimespanExpandingBeforeEnd(t *testing.T) {
	timespans := []*TimeSpan{
		{
			TimeStart: time.Date(2013, 9, 10, 14, 30, 0, 0, time.UTC),
			TimeEnd:   time.Date(2013, 9, 10, 14, 30, 30, 0, time.UTC),
			RateInterval: &RateInterval{Rating: &RIRate{Rates: RateGroups{
				&RGRate{RateIncrement: 45 * time.Second},
			}}},
		},
		{
			TimeStart: time.Date(2013, 9, 10, 14, 30, 30, 0, time.UTC),
			TimeEnd:   time.Date(2013, 9, 10, 14, 31, 0, 0, time.UTC),
		},
	}
	cd := &CallDescriptor{}
	timespans = cd.roundTimeSpansToIncrement(timespans)
	if len(timespans) != 2 {
		t.Error("Error removing overlaped intervals: ", timespans)
	}
	if !timespans[0].TimeEnd.Equal(time.Date(2013, 9, 10, 14, 30, 45, 0, time.UTC)) ||
		!timespans[1].TimeStart.Equal(time.Date(2013, 9, 10, 14, 30, 45, 0, time.UTC)) ||
		!timespans[1].TimeEnd.Equal(time.Date(2013, 9, 10, 14, 31, 0, 0, time.UTC)) {
		t.Error("Error expanding timespan: ", timespans[0])
	}
}

func TestTimespanExpandingBeforeEndMultiple(t *testing.T) {
	timespans := []*TimeSpan{
		{
			TimeStart: time.Date(2013, 9, 10, 14, 30, 0, 0, time.UTC),
			TimeEnd:   time.Date(2013, 9, 10, 14, 30, 30, 0, time.UTC),
			RateInterval: &RateInterval{Rating: &RIRate{Rates: RateGroups{
				&RGRate{RateIncrement: 45 * time.Second},
			}}},
		},
		{
			TimeStart: time.Date(2013, 9, 10, 14, 30, 30, 0, time.UTC),
			TimeEnd:   time.Date(2013, 9, 10, 14, 30, 50, 0, time.UTC),
		},
		{
			TimeStart: time.Date(2013, 9, 10, 14, 30, 50, 0, time.UTC),
			TimeEnd:   time.Date(2013, 9, 10, 14, 31, 00, 0, time.UTC),
		},
	}
	cd := &CallDescriptor{}
	timespans = cd.roundTimeSpansToIncrement(timespans)
	if len(timespans) != 3 {
		t.Error("Error removing overlaped intervals: ", timespans)
	}
	if !timespans[0].TimeEnd.Equal(time.Date(2013, 9, 10, 14, 30, 45, 0, time.UTC)) ||
		!timespans[1].TimeStart.Equal(time.Date(2013, 9, 10, 14, 30, 45, 0, time.UTC)) ||
		!timespans[1].TimeEnd.Equal(time.Date(2013, 9, 10, 14, 30, 50, 0, time.UTC)) {
		t.Error("Error expanding timespan: ", timespans[0])
	}
}

func TestTimespanCreateSecondsSlice(t *testing.T) {
	ts := &TimeSpan{
		TimeStart: time.Date(2013, 9, 10, 14, 30, 0, 0, time.UTC),
		TimeEnd:   time.Date(2013, 9, 10, 14, 30, 30, 0, time.UTC),
		RateInterval: &RateInterval{Rating: &RIRate{Rates: RateGroups{
			&RGRate{Value: 2.0},
		}}},
	}
	ts.createIncrementsSlice()
	if len(ts.Increments) != 30 {
		t.Error("Error creating second slice: ", ts.Increments)
	}
	if ts.Increments[0].Cost != 2.0 {
		t.Error("Wrong second slice: ", ts.Increments[0])
	}
}

func TestTimespanCreateIncrements(t *testing.T) {
	ts := &TimeSpan{
		TimeStart: time.Date(2013, 9, 10, 14, 30, 0, 0, time.UTC),
		TimeEnd:   time.Date(2013, 9, 10, 14, 30, 30, 100000000, time.UTC),
		RateInterval: &RateInterval{
			Rating: &RIRate{
				RoundingMethod:   utils.MetaRoundingMiddle,
				RoundingDecimals: 2,
				Rates: RateGroups{
					&RGRate{
						Value:         2.0,
						RateIncrement: 10 * time.Second,
					},
				},
			},
		},
	}
	ts.createIncrementsSlice()
	if len(ts.Increments) != 3 {
		t.Error("Error creating increment slice: ", len(ts.Increments))
	}
	if len(ts.Increments) < 3 || ts.Increments[2].Cost != 20.066667 {
		t.Error("Wrong second slice: ", ts.Increments[2].Cost)
	}
}

func TestTimespanSplitByIncrement(t *testing.T) {
	ts := &TimeSpan{
		TimeStart:     time.Date(2013, 9, 19, 18, 30, 0, 0, time.UTC),
		TimeEnd:       time.Date(2013, 9, 19, 18, 31, 00, 0, time.UTC),
		DurationIndex: 60 * time.Second,
		ratingInfo:    &RatingInfo{},
		RateInterval: &RateInterval{
			Rating: &RIRate{
				RoundingMethod:   utils.MetaRoundingMiddle,
				RoundingDecimals: 2,
				Rates: RateGroups{
					&RGRate{
						Value:         2.0,
						RateIncrement: 10 * time.Second,
					},
				},
			},
		},
	}
	ts.createIncrementsSlice()
	if len(ts.Increments) != 6 {
		t.Error("Error creating increment slice: ", len(ts.Increments))
	}
	newTs := ts.SplitByIncrement(5)
	if ts.GetDuration() != 50*time.Second || newTs.GetDuration() != 10*time.Second {
		t.Error("Error spliting by increment: ", ts.GetDuration(), newTs.GetDuration())
	}
	if ts.DurationIndex != 50*time.Second || newTs.DurationIndex != 60*time.Second {
		t.Error("Error spliting by increment at setting call duration: ", ts.DurationIndex, newTs.DurationIndex)
	}
	if len(ts.Increments) != 5 || len(newTs.Increments) != 1 {
		t.Error("Error spliting increments: ", ts.Increments, newTs.Increments)
	}
}

func TestTimespanSplitByIncrementStart(t *testing.T) {
	ts := &TimeSpan{
		TimeStart:     time.Date(2013, 9, 19, 18, 30, 0, 0, time.UTC),
		TimeEnd:       time.Date(2013, 9, 19, 18, 31, 00, 0, time.UTC),
		DurationIndex: 60 * time.Second,
		RateInterval: &RateInterval{
			Rating: &RIRate{
				RoundingMethod:   utils.MetaRoundingMiddle,
				RoundingDecimals: 2,
				Rates: RateGroups{
					&RGRate{
						Value:         2.0,
						RateIncrement: 10 * time.Second,
					},
				},
			},
		},
	}
	ts.createIncrementsSlice()
	if len(ts.Increments) != 6 {
		t.Error("Error creating increment slice: ", len(ts.Increments))
	}
	newTs := ts.SplitByIncrement(0)
	if ts.GetDuration() != 60*time.Second || newTs != nil {
		t.Error("Error spliting by increment: ", ts.GetDuration())
	}
	if ts.DurationIndex != 60*time.Second {
		t.Error("Error spliting by incrementat setting call duration: ", ts.DurationIndex)
	}
	if len(ts.Increments) != 6 {
		t.Error("Error spliting increments: ", ts.Increments)
	}
}

func TestTimespanSplitByIncrementEnd(t *testing.T) {
	ts := &TimeSpan{
		TimeStart:     time.Date(2013, 9, 19, 18, 30, 0, 0, time.UTC),
		TimeEnd:       time.Date(2013, 9, 19, 18, 31, 00, 0, time.UTC),
		DurationIndex: 60 * time.Second,
		RateInterval: &RateInterval{
			Rating: &RIRate{
				RoundingMethod:   utils.MetaRoundingMiddle,
				RoundingDecimals: 2,
				Rates: RateGroups{
					&RGRate{
						Value:         2.0,
						RateIncrement: 10 * time.Second,
					},
				},
			},
		},
	}
	ts.createIncrementsSlice()
	if len(ts.Increments) != 6 {
		t.Error("Error creating increment slice: ", len(ts.Increments))
	}
	newTs := ts.SplitByIncrement(6)
	if ts.GetDuration() != 60*time.Second || newTs != nil {
		t.Error("Error spliting by increment: ", ts.GetDuration())
	}
	if ts.DurationIndex != 60*time.Second {
		t.Error("Error spliting by increment at setting call duration: ", ts.DurationIndex)
	}
	if len(ts.Increments) != 6 {
		t.Error("Error spliting increments: ", ts.Increments)
	}
}

func TestTimespanSplitByDuration(t *testing.T) {
	ts := &TimeSpan{
		TimeStart:     time.Date(2013, 9, 19, 18, 30, 0, 0, time.UTC),
		TimeEnd:       time.Date(2013, 9, 19, 18, 31, 00, 0, time.UTC),
		DurationIndex: 60 * time.Second,
		ratingInfo:    &RatingInfo{},
		RateInterval: &RateInterval{
			Rating: &RIRate{
				RoundingMethod:   utils.MetaRoundingMiddle,
				RoundingDecimals: 2,
				Rates: RateGroups{
					&RGRate{
						Value:         2.0,
						RateIncrement: 10 * time.Second,
					},
				},
			},
		},
	}
	ts.createIncrementsSlice()
	if len(ts.Increments) != 6 {
		t.Error("Error creating increment slice: ", len(ts.Increments))
	}
	newTs := ts.SplitByDuration(46 * time.Second)
	if ts.GetDuration() != 46*time.Second || newTs.GetDuration() != 14*time.Second {
		t.Error("Error spliting by duration: ", ts.GetDuration(), newTs.GetDuration())
	}
	if ts.DurationIndex != 46*time.Second || newTs.DurationIndex != 60*time.Second {
		t.Error("Error spliting by duration at setting call duration: ", ts.DurationIndex, newTs.DurationIndex)
	}
	if len(ts.Increments) != 5 || len(newTs.Increments) != 2 {
		t.Error("Error spliting increments: ", ts.Increments, newTs.Increments)
	}
	if ts.Increments[4].Duration != 6*time.Second || newTs.Increments[0].Duration != 4*time.Second {
		t.Error("Error spliting increment: ", ts.Increments[4], newTs.Increments[0])
	}
}

func TestRemoveOverlapedFromIndexMiddle(t *testing.T) {
	tss := TimeSpans{
		&TimeSpan{
			TimeStart: time.Date(2013, 12, 5, 15, 45, 0, 0, time.UTC),
			TimeEnd:   time.Date(2013, 12, 5, 15, 46, 0, 0, time.UTC),
		},
		&TimeSpan{
			TimeStart: time.Date(2013, 12, 5, 15, 46, 0, 0, time.UTC),
			TimeEnd:   time.Date(2013, 12, 5, 15, 48, 0, 0, time.UTC),
		},
		&TimeSpan{
			TimeStart: time.Date(2013, 12, 5, 15, 47, 0, 0, time.UTC),
			TimeEnd:   time.Date(2013, 12, 5, 15, 48, 0, 0, time.UTC),
		},
		&TimeSpan{
			TimeStart: time.Date(2013, 12, 5, 15, 48, 0, 0, time.UTC),
			TimeEnd:   time.Date(2013, 12, 5, 15, 49, 0, 0, time.UTC),
		},
	}
	(&tss).RemoveOverlapedFromIndex(1)
	if len(tss) != 3 ||
		tss[0].TimeEnd != time.Date(2013, 12, 5, 15, 46, 0, 0, time.UTC) ||
		tss[1].TimeEnd != time.Date(2013, 12, 5, 15, 48, 0, 0, time.UTC) ||
		tss[2].TimeEnd != time.Date(2013, 12, 5, 15, 49, 0, 0, time.UTC) {
		for _, ts := range tss {
			t.Logf("TS: %+v", ts)
		}
		t.Error("Error removing overlaped timespans: ", tss)
	}
}

func TestRemoveOverlapedFromIndexMiddleNonBounds(t *testing.T) {
	tss := TimeSpans{
		&TimeSpan{
			TimeStart: time.Date(2013, 12, 5, 15, 45, 0, 0, time.UTC),
			TimeEnd:   time.Date(2013, 12, 5, 15, 46, 0, 0, time.UTC),
		},
		&TimeSpan{
			TimeStart: time.Date(2013, 12, 5, 15, 46, 0, 0, time.UTC),
			TimeEnd:   time.Date(2013, 12, 5, 15, 47, 30, 0, time.UTC),
		},
		&TimeSpan{
			TimeStart: time.Date(2013, 12, 5, 15, 47, 0, 0, time.UTC),
			TimeEnd:   time.Date(2013, 12, 5, 15, 48, 0, 0, time.UTC),
		},
		&TimeSpan{
			TimeStart: time.Date(2013, 12, 5, 15, 48, 0, 0, time.UTC),
			TimeEnd:   time.Date(2013, 12, 5, 15, 49, 0, 0, time.UTC),
		},
	}
	(&tss).RemoveOverlapedFromIndex(1)
	if len(tss) != 4 ||
		tss[0].TimeEnd != time.Date(2013, 12, 5, 15, 46, 0, 0, time.UTC) ||
		tss[1].TimeEnd != time.Date(2013, 12, 5, 15, 47, 30, 0, time.UTC) ||
		tss[2].TimeStart != time.Date(2013, 12, 5, 15, 47, 30, 0, time.UTC) ||
		tss[2].TimeEnd != time.Date(2013, 12, 5, 15, 48, 0, 0, time.UTC) ||
		tss[3].TimeEnd != time.Date(2013, 12, 5, 15, 49, 0, 0, time.UTC) {
		for _, ts := range tss {
			t.Logf("TS: %+v", ts)
		}
		t.Error("Error removing overlaped timespans: ", tss)
	}
}

func TestRemoveOverlapedFromIndexMiddleNonBoundsOver(t *testing.T) {
	tss := TimeSpans{
		&TimeSpan{
			TimeStart: time.Date(2013, 12, 5, 15, 45, 0, 0, time.UTC),
			TimeEnd:   time.Date(2013, 12, 5, 15, 46, 0, 0, time.UTC),
		},
		&TimeSpan{
			TimeStart: time.Date(2013, 12, 5, 15, 46, 0, 0, time.UTC),
			TimeEnd:   time.Date(2013, 12, 5, 15, 48, 30, 0, time.UTC),
		},
		&TimeSpan{
			TimeStart: time.Date(2013, 12, 5, 15, 47, 0, 0, time.UTC),
			TimeEnd:   time.Date(2013, 12, 5, 15, 48, 0, 0, time.UTC),
		},
		&TimeSpan{
			TimeStart: time.Date(2013, 12, 5, 15, 48, 0, 0, time.UTC),
			TimeEnd:   time.Date(2013, 12, 5, 15, 49, 0, 0, time.UTC),
		},
	}
	(&tss).RemoveOverlapedFromIndex(1)
	if len(tss) != 3 ||
		tss[0].TimeEnd != time.Date(2013, 12, 5, 15, 46, 0, 0, time.UTC) ||
		tss[1].TimeEnd != time.Date(2013, 12, 5, 15, 48, 30, 0, time.UTC) ||
		tss[2].TimeStart != time.Date(2013, 12, 5, 15, 48, 30, 0, time.UTC) ||
		tss[2].TimeEnd != time.Date(2013, 12, 5, 15, 49, 0, 0, time.UTC) {
		for _, ts := range tss {
			t.Logf("TS: %+v", ts)
		}
		t.Error("Error removing overlaped timespans: ", tss)
	}
}

func TestRemoveOverlapedFromIndexEnd(t *testing.T) {
	tss := TimeSpans{
		&TimeSpan{
			TimeStart: time.Date(2013, 12, 5, 15, 45, 0, 0, time.UTC),
			TimeEnd:   time.Date(2013, 12, 5, 15, 46, 0, 0, time.UTC),
		},
		&TimeSpan{
			TimeStart: time.Date(2013, 12, 5, 15, 46, 0, 0, time.UTC),
			TimeEnd:   time.Date(2013, 12, 5, 15, 49, 0, 0, time.UTC),
		},
		&TimeSpan{
			TimeStart: time.Date(2013, 12, 5, 15, 47, 0, 0, time.UTC),
			TimeEnd:   time.Date(2013, 12, 5, 15, 48, 0, 0, time.UTC),
		},
		&TimeSpan{
			TimeStart: time.Date(2013, 12, 5, 15, 48, 0, 0, time.UTC),
			TimeEnd:   time.Date(2013, 12, 5, 15, 49, 0, 0, time.UTC),
		},
	}
	(&tss).RemoveOverlapedFromIndex(1)
	if len(tss) != 2 ||
		tss[0].TimeEnd != time.Date(2013, 12, 5, 15, 46, 0, 0, time.UTC) ||
		tss[1].TimeEnd != time.Date(2013, 12, 5, 15, 49, 0, 0, time.UTC) {
		for _, ts := range tss {
			t.Logf("TS: %+v", ts)
		}
		t.Error("Error removing overlaped timespans: ", tss)
	}
}

func TestRemoveOverlapedFromIndexEndPast(t *testing.T) {
	tss := TimeSpans{
		&TimeSpan{
			TimeStart: time.Date(2013, 12, 5, 15, 45, 0, 0, time.UTC),
			TimeEnd:   time.Date(2013, 12, 5, 15, 46, 0, 0, time.UTC),
		},
		&TimeSpan{
			TimeStart: time.Date(2013, 12, 5, 15, 46, 0, 0, time.UTC),
			TimeEnd:   time.Date(2013, 12, 5, 15, 50, 0, 0, time.UTC),
		},
		&TimeSpan{
			TimeStart: time.Date(2013, 12, 5, 15, 47, 0, 0, time.UTC),
			TimeEnd:   time.Date(2013, 12, 5, 15, 48, 0, 0, time.UTC),
		},
		&TimeSpan{
			TimeStart: time.Date(2013, 12, 5, 15, 48, 0, 0, time.UTC),
			TimeEnd:   time.Date(2013, 12, 5, 15, 49, 0, 0, time.UTC),
		},
	}
	(&tss).RemoveOverlapedFromIndex(1)
	if len(tss) != 2 ||
		tss[0].TimeEnd != time.Date(2013, 12, 5, 15, 46, 0, 0, time.UTC) ||
		tss[1].TimeEnd != time.Date(2013, 12, 5, 15, 50, 0, 0, time.UTC) {
		for _, ts := range tss {
			t.Logf("TS: %+v", ts)
		}
		t.Error("Error removing overlaped timespans: ", tss)
	}
}

func TestRemoveOverlapedFromIndexAll(t *testing.T) {
	tss := TimeSpans{
		&TimeSpan{
			TimeStart: time.Date(2013, 12, 5, 15, 45, 0, 0, time.UTC),
			TimeEnd:   time.Date(2013, 12, 5, 15, 49, 0, 0, time.UTC),
		},
		&TimeSpan{
			TimeStart: time.Date(2013, 12, 5, 15, 46, 0, 0, time.UTC),
			TimeEnd:   time.Date(2013, 12, 5, 15, 47, 0, 0, time.UTC),
		},
		&TimeSpan{
			TimeStart: time.Date(2013, 12, 5, 15, 47, 0, 0, time.UTC),
			TimeEnd:   time.Date(2013, 12, 5, 15, 48, 0, 0, time.UTC),
		},
		&TimeSpan{
			TimeStart: time.Date(2013, 12, 5, 15, 48, 0, 0, time.UTC),
			TimeEnd:   time.Date(2013, 12, 5, 15, 49, 0, 0, time.UTC),
		},
	}
	(&tss).RemoveOverlapedFromIndex(0)
	if len(tss) != 1 ||
		tss[0].TimeEnd != time.Date(2013, 12, 5, 15, 49, 0, 0, time.UTC) {
		for _, ts := range tss {
			t.Logf("TS: %+v", ts)
		}
		t.Error("Error removing overlaped timespans: ", tss)
	}
}

func TestRemoveOverlapedFromIndexNone(t *testing.T) {
	tss := TimeSpans{
		&TimeSpan{
			TimeStart: time.Date(2013, 12, 5, 15, 45, 0, 0, time.UTC),
			TimeEnd:   time.Date(2013, 12, 5, 15, 46, 0, 0, time.UTC),
		},
		&TimeSpan{
			TimeStart: time.Date(2013, 12, 5, 15, 46, 0, 0, time.UTC),
			TimeEnd:   time.Date(2013, 12, 5, 15, 47, 0, 0, time.UTC),
		},
		&TimeSpan{
			TimeStart: time.Date(2013, 12, 5, 15, 47, 0, 0, time.UTC),
			TimeEnd:   time.Date(2013, 12, 5, 15, 48, 0, 0, time.UTC),
		},
		&TimeSpan{
			TimeStart: time.Date(2013, 12, 5, 15, 48, 0, 0, time.UTC),
			TimeEnd:   time.Date(2013, 12, 5, 15, 49, 0, 0, time.UTC),
		},
	}
	(&tss).RemoveOverlapedFromIndex(0)
	if len(tss) != 4 ||
		tss[0].TimeEnd != time.Date(2013, 12, 5, 15, 46, 0, 0, time.UTC) ||
		tss[1].TimeEnd != time.Date(2013, 12, 5, 15, 47, 0, 0, time.UTC) ||
		tss[2].TimeEnd != time.Date(2013, 12, 5, 15, 48, 0, 0, time.UTC) ||
		tss[3].TimeEnd != time.Date(2013, 12, 5, 15, 49, 0, 0, time.UTC) {
		for _, ts := range tss {
			t.Logf("TS: %+v", ts)
		}
		t.Error("Error removing overlaped timespans: ", tss)
	}
}

func TestRemoveOverlapedFromIndexOne(t *testing.T) {
	tss := TimeSpans{
		&TimeSpan{
			TimeStart: time.Date(2013, 12, 5, 15, 45, 0, 0, time.UTC),
			TimeEnd:   time.Date(2013, 12, 5, 15, 46, 0, 0, time.UTC),
		},
	}
	(&tss).RemoveOverlapedFromIndex(0)
	if len(tss) != 1 ||
		tss[0].TimeEnd != time.Date(2013, 12, 5, 15, 46, 0, 0, time.UTC) {
		for _, ts := range tss {
			t.Logf("TS: %+v", ts)
		}
		t.Error("Error removing overlaped timespans: ", tss)
	}
}

func TestRemoveOverlapedFromIndexTwo(t *testing.T) {
	tss := TimeSpans{
		&TimeSpan{
			TimeStart: time.Date(2013, 12, 5, 15, 45, 0, 0, time.UTC),
			TimeEnd:   time.Date(2013, 12, 5, 15, 50, 0, 0, time.UTC),
		},
		&TimeSpan{
			TimeStart: time.Date(2013, 12, 5, 15, 46, 0, 0, time.UTC),
			TimeEnd:   time.Date(2013, 12, 5, 15, 47, 0, 0, time.UTC),
		},
	}
	(&tss).RemoveOverlapedFromIndex(0)
	if len(tss) != 1 ||
		tss[0].TimeEnd != time.Date(2013, 12, 5, 15, 50, 0, 0, time.UTC) {
		for _, ts := range tss {
			t.Logf("TS: %+v", ts)
		}
		t.Error("Error removing overlaped timespans: ", tss)
	}
}

func TestOverlapWithTimeSpansMiddleLong(t *testing.T) {
	tss := TimeSpans{
		&TimeSpan{
			TimeStart: time.Date(2013, 12, 5, 15, 45, 0, 0, time.UTC),
			TimeEnd:   time.Date(2013, 12, 5, 15, 46, 0, 0, time.UTC),
		},
		&TimeSpan{
			TimeStart: time.Date(2013, 12, 5, 15, 46, 0, 0, time.UTC),
			TimeEnd:   time.Date(2013, 12, 5, 15, 47, 0, 0, time.UTC),
		},
		&TimeSpan{
			TimeStart: time.Date(2013, 12, 5, 15, 47, 0, 0, time.UTC),
			TimeEnd:   time.Date(2013, 12, 5, 15, 48, 0, 0, time.UTC),
		},
		&TimeSpan{
			TimeStart: time.Date(2013, 12, 5, 15, 48, 0, 0, time.UTC),
			TimeEnd:   time.Date(2013, 12, 5, 15, 49, 0, 0, time.UTC),
		},
	}
	newTss := TimeSpans{
		&TimeSpan{
			TimeStart: time.Date(2013, 12, 5, 15, 46, 0, 0, time.UTC),
			TimeEnd:   time.Date(2013, 12, 5, 15, 48, 30, 0, time.UTC),
		},
	}
	(&tss).OverlapWithTimeSpans(newTss, nil, 1)
	if len(tss) != 3 ||
		tss[0].TimeEnd != time.Date(2013, 12, 5, 15, 46, 0, 0, time.UTC) ||
		tss[1].TimeEnd != time.Date(2013, 12, 5, 15, 48, 30, 0, time.UTC) ||
		tss[2].TimeEnd != time.Date(2013, 12, 5, 15, 49, 0, 0, time.UTC) {
		for _, ts := range tss {
			t.Logf("TS: %v", ts)
		}
		t.Error("Error overlaping with timespans timespans: ", tss)
	}
}

func TestOverlapWithTimeSpansMiddleMedium(t *testing.T) {
	tss := TimeSpans{
		&TimeSpan{
			TimeStart: time.Date(2013, 12, 5, 15, 45, 0, 0, time.UTC),
			TimeEnd:   time.Date(2013, 12, 5, 15, 46, 0, 0, time.UTC),
		},
		&TimeSpan{
			TimeStart: time.Date(2013, 12, 5, 15, 46, 0, 0, time.UTC),
			TimeEnd:   time.Date(2013, 12, 5, 15, 47, 0, 0, time.UTC),
		},
		&TimeSpan{
			TimeStart: time.Date(2013, 12, 5, 15, 47, 0, 0, time.UTC),
			TimeEnd:   time.Date(2013, 12, 5, 15, 48, 0, 0, time.UTC),
		},
		&TimeSpan{
			TimeStart: time.Date(2013, 12, 5, 15, 48, 0, 0, time.UTC),
			TimeEnd:   time.Date(2013, 12, 5, 15, 49, 0, 0, time.UTC),
		},
	}
	newTss := TimeSpans{
		&TimeSpan{
			TimeStart: time.Date(2013, 12, 5, 15, 46, 0, 0, time.UTC),
			TimeEnd:   time.Date(2013, 12, 5, 15, 47, 30, 0, time.UTC),
		},
	}
	(&tss).OverlapWithTimeSpans(newTss, nil, 1)
	if len(tss) != 4 ||
		tss[0].TimeEnd != time.Date(2013, 12, 5, 15, 46, 0, 0, time.UTC) ||
		tss[1].TimeEnd != time.Date(2013, 12, 5, 15, 47, 30, 0, time.UTC) ||
		tss[2].TimeEnd != time.Date(2013, 12, 5, 15, 48, 0, 0, time.UTC) ||
		tss[3].TimeEnd != time.Date(2013, 12, 5, 15, 49, 0, 0, time.UTC) {
		for _, ts := range tss {
			t.Logf("TS: %v", ts)
		}
		t.Error("Error overlaping with timespans timespans: ", tss)
	}
}

func TestOverlapWithTimeSpansMiddleShort(t *testing.T) {
	tss := TimeSpans{
		&TimeSpan{
			TimeStart: time.Date(2013, 12, 5, 15, 45, 0, 0, time.UTC),
			TimeEnd:   time.Date(2013, 12, 5, 15, 46, 0, 0, time.UTC),
		},
		&TimeSpan{
			TimeStart: time.Date(2013, 12, 5, 15, 46, 0, 0, time.UTC),
			TimeEnd:   time.Date(2013, 12, 5, 15, 47, 0, 0, time.UTC),
		},
		&TimeSpan{
			TimeStart: time.Date(2013, 12, 5, 15, 47, 0, 0, time.UTC),
			TimeEnd:   time.Date(2013, 12, 5, 15, 48, 0, 0, time.UTC),
		},
		&TimeSpan{
			TimeStart: time.Date(2013, 12, 5, 15, 48, 0, 0, time.UTC),
			TimeEnd:   time.Date(2013, 12, 5, 15, 49, 0, 0, time.UTC),
		},
	}
	newTss := TimeSpans{
		&TimeSpan{
			TimeStart: time.Date(2013, 12, 5, 15, 46, 0, 0, time.UTC),
			TimeEnd:   time.Date(2013, 12, 5, 15, 46, 30, 0, time.UTC),
		},
	}
	(&tss).OverlapWithTimeSpans(newTss, nil, 1)
	if len(tss) != 5 ||
		tss[0].TimeEnd != time.Date(2013, 12, 5, 15, 46, 0, 0, time.UTC) ||
		tss[1].TimeEnd != time.Date(2013, 12, 5, 15, 46, 30, 0, time.UTC) ||
		tss[2].TimeEnd != time.Date(2013, 12, 5, 15, 47, 0, 0, time.UTC) ||
		tss[3].TimeEnd != time.Date(2013, 12, 5, 15, 48, 0, 0, time.UTC) ||
		tss[4].TimeEnd != time.Date(2013, 12, 5, 15, 49, 0, 0, time.UTC) {
		for _, ts := range tss {
			t.Logf("TS: %v", ts)
		}
		t.Error("Error overlaping with timespans timespans: ", tss)
	}
}

func TestOverlapWithTimeSpansStart(t *testing.T) {
	tss := TimeSpans{
		&TimeSpan{
			TimeStart: time.Date(2013, 12, 5, 15, 45, 0, 0, time.UTC),
			TimeEnd:   time.Date(2013, 12, 5, 15, 46, 0, 0, time.UTC),
		},
		&TimeSpan{
			TimeStart: time.Date(2013, 12, 5, 15, 46, 0, 0, time.UTC),
			TimeEnd:   time.Date(2013, 12, 5, 15, 47, 0, 0, time.UTC),
		},
		&TimeSpan{
			TimeStart: time.Date(2013, 12, 5, 15, 47, 0, 0, time.UTC),
			TimeEnd:   time.Date(2013, 12, 5, 15, 48, 0, 0, time.UTC),
		},
		&TimeSpan{
			TimeStart: time.Date(2013, 12, 5, 15, 48, 0, 0, time.UTC),
			TimeEnd:   time.Date(2013, 12, 5, 15, 49, 0, 0, time.UTC),
		},
	}
	newTss := TimeSpans{
		&TimeSpan{
			TimeStart: time.Date(2013, 12, 5, 15, 45, 0, 0, time.UTC),
			TimeEnd:   time.Date(2013, 12, 5, 15, 47, 30, 0, time.UTC),
		},
	}
	(&tss).OverlapWithTimeSpans(newTss, nil, 0)
	if len(tss) != 3 ||
		tss[0].TimeEnd != time.Date(2013, 12, 5, 15, 47, 30, 0, time.UTC) ||
		tss[1].TimeEnd != time.Date(2013, 12, 5, 15, 48, 0, 0, time.UTC) ||
		tss[2].TimeEnd != time.Date(2013, 12, 5, 15, 49, 0, 0, time.UTC) {
		for _, ts := range tss {
			t.Logf("TS: %v", ts)
		}
		t.Error("Error overlaping with timespans timespans: ", tss)
	}
}

func TestOverlapWithTimeSpansAlmostEnd(t *testing.T) {
	tss := TimeSpans{
		&TimeSpan{
			TimeStart: time.Date(2013, 12, 5, 15, 45, 0, 0, time.UTC),
			TimeEnd:   time.Date(2013, 12, 5, 15, 46, 0, 0, time.UTC),
		},
		&TimeSpan{
			TimeStart: time.Date(2013, 12, 5, 15, 46, 0, 0, time.UTC),
			TimeEnd:   time.Date(2013, 12, 5, 15, 47, 0, 0, time.UTC),
		},
		&TimeSpan{
			TimeStart: time.Date(2013, 12, 5, 15, 47, 0, 0, time.UTC),
			TimeEnd:   time.Date(2013, 12, 5, 15, 48, 0, 0, time.UTC),
		},
		&TimeSpan{
			TimeStart: time.Date(2013, 12, 5, 15, 48, 0, 0, time.UTC),
			TimeEnd:   time.Date(2013, 12, 5, 15, 49, 0, 0, time.UTC),
		},
	}
	newTss := TimeSpans{
		&TimeSpan{
			TimeStart: time.Date(2013, 12, 5, 15, 48, 0, 0, time.UTC),
			TimeEnd:   time.Date(2013, 12, 5, 15, 48, 30, 0, time.UTC),
		},
	}
	(&tss).OverlapWithTimeSpans(newTss, nil, 3)
	if len(tss) != 5 ||
		tss[0].TimeEnd != time.Date(2013, 12, 5, 15, 46, 0, 0, time.UTC) ||
		tss[1].TimeEnd != time.Date(2013, 12, 5, 15, 47, 0, 0, time.UTC) ||
		tss[2].TimeEnd != time.Date(2013, 12, 5, 15, 48, 0, 0, time.UTC) ||
		tss[3].TimeEnd != time.Date(2013, 12, 5, 15, 48, 30, 0, time.UTC) ||
		tss[4].TimeEnd != time.Date(2013, 12, 5, 15, 49, 0, 0, time.UTC) {
		for _, ts := range tss {
			t.Logf("TS: %v", ts)
		}
		t.Error("Error overlaping with timespans timespans: ", tss)
	}
}

func TestOverlapWithTimeSpansEnd(t *testing.T) {
	tss := TimeSpans{
		&TimeSpan{
			TimeStart: time.Date(2013, 12, 5, 15, 45, 0, 0, time.UTC),
			TimeEnd:   time.Date(2013, 12, 5, 15, 46, 0, 0, time.UTC),
		},
		&TimeSpan{
			TimeStart: time.Date(2013, 12, 5, 15, 46, 0, 0, time.UTC),
			TimeEnd:   time.Date(2013, 12, 5, 15, 47, 0, 0, time.UTC),
		},
		&TimeSpan{
			TimeStart: time.Date(2013, 12, 5, 15, 47, 0, 0, time.UTC),
			TimeEnd:   time.Date(2013, 12, 5, 15, 48, 0, 0, time.UTC),
		},
		&TimeSpan{
			TimeStart: time.Date(2013, 12, 5, 15, 48, 0, 0, time.UTC),
			TimeEnd:   time.Date(2013, 12, 5, 15, 49, 0, 0, time.UTC),
		},
	}
	newTss := TimeSpans{
		&TimeSpan{
			TimeStart: time.Date(2013, 12, 5, 15, 48, 0, 0, time.UTC),
			TimeEnd:   time.Date(2013, 12, 5, 15, 49, 0, 0, time.UTC),
		},
	}
	(&tss).OverlapWithTimeSpans(newTss, nil, 3)
	if len(tss) != 4 ||
		tss[0].TimeEnd != time.Date(2013, 12, 5, 15, 46, 0, 0, time.UTC) ||
		tss[1].TimeEnd != time.Date(2013, 12, 5, 15, 47, 0, 0, time.UTC) ||
		tss[2].TimeEnd != time.Date(2013, 12, 5, 15, 48, 0, 0, time.UTC) ||
		tss[3].TimeEnd != time.Date(2013, 12, 5, 15, 49, 0, 0, time.UTC) {
		for _, ts := range tss {
			t.Logf("TS: %v", ts)
		}
		t.Error("Error overlaping with timespans timespans: ", tss)
	}
}

func TestOverlapWithTimeSpansPastEnd(t *testing.T) {
	tss := TimeSpans{
		&TimeSpan{
			TimeStart: time.Date(2013, 12, 5, 15, 45, 0, 0, time.UTC),
			TimeEnd:   time.Date(2013, 12, 5, 15, 46, 0, 0, time.UTC),
		},
		&TimeSpan{
			TimeStart: time.Date(2013, 12, 5, 15, 46, 0, 0, time.UTC),
			TimeEnd:   time.Date(2013, 12, 5, 15, 47, 0, 0, time.UTC),
		},
		&TimeSpan{
			TimeStart: time.Date(2013, 12, 5, 15, 47, 0, 0, time.UTC),
			TimeEnd:   time.Date(2013, 12, 5, 15, 48, 0, 0, time.UTC),
		},
		&TimeSpan{
			TimeStart: time.Date(2013, 12, 5, 15, 48, 0, 0, time.UTC),
			TimeEnd:   time.Date(2013, 12, 5, 15, 49, 0, 0, time.UTC),
		},
	}
	newTss := TimeSpans{
		&TimeSpan{
			TimeStart: time.Date(2013, 12, 5, 15, 48, 0, 0, time.UTC),
			TimeEnd:   time.Date(2013, 12, 5, 15, 49, 30, 0, time.UTC),
		},
	}
	(&tss).OverlapWithTimeSpans(newTss, nil, 3)
	if len(tss) != 4 ||
		tss[0].TimeEnd != time.Date(2013, 12, 5, 15, 46, 0, 0, time.UTC) ||
		tss[1].TimeEnd != time.Date(2013, 12, 5, 15, 47, 0, 0, time.UTC) ||
		tss[2].TimeEnd != time.Date(2013, 12, 5, 15, 48, 0, 0, time.UTC) ||
		tss[3].TimeEnd != time.Date(2013, 12, 5, 15, 49, 30, 0, time.UTC) {
		for _, ts := range tss {
			t.Logf("TS: %v", ts)
		}
		t.Error("Error overlaping with timespans timespans: ", tss)
	}
}

func TestOverlapWithTimeSpansAll(t *testing.T) {
	tss := TimeSpans{
		&TimeSpan{
			TimeStart: time.Date(2013, 12, 5, 15, 45, 0, 0, time.UTC),
			TimeEnd:   time.Date(2013, 12, 5, 15, 46, 0, 0, time.UTC),
		},
		&TimeSpan{
			TimeStart: time.Date(2013, 12, 5, 15, 46, 0, 0, time.UTC),
			TimeEnd:   time.Date(2013, 12, 5, 15, 47, 0, 0, time.UTC),
		},
		&TimeSpan{
			TimeStart: time.Date(2013, 12, 5, 15, 47, 0, 0, time.UTC),
			TimeEnd:   time.Date(2013, 12, 5, 15, 48, 0, 0, time.UTC),
		},
		&TimeSpan{
			TimeStart: time.Date(2013, 12, 5, 15, 48, 0, 0, time.UTC),
			TimeEnd:   time.Date(2013, 12, 5, 15, 49, 0, 0, time.UTC),
		},
	}
	newTss := TimeSpans{
		&TimeSpan{
			TimeStart: time.Date(2013, 12, 5, 15, 45, 0, 0, time.UTC),
			TimeEnd:   time.Date(2013, 12, 5, 15, 49, 0, 0, time.UTC),
		},
	}
	(&tss).OverlapWithTimeSpans(newTss, nil, 0)
	if len(tss) != 1 ||
		tss[0].TimeStart != time.Date(2013, 12, 5, 15, 45, 0, 0, time.UTC) ||
		tss[0].TimeEnd != time.Date(2013, 12, 5, 15, 49, 0, 0, time.UTC) {
		for _, ts := range tss {
			t.Logf("TS: %v", ts)
		}
		t.Error("Error overlaping with timespans timespans: ", tss)
	}
}

func TestOverlapWithTimeSpansAllPast(t *testing.T) {
	tss := TimeSpans{
		&TimeSpan{
			TimeStart: time.Date(2013, 12, 5, 15, 45, 0, 0, time.UTC),
			TimeEnd:   time.Date(2013, 12, 5, 15, 46, 0, 0, time.UTC),
		},
		&TimeSpan{
			TimeStart: time.Date(2013, 12, 5, 15, 46, 0, 0, time.UTC),
			TimeEnd:   time.Date(2013, 12, 5, 15, 47, 0, 0, time.UTC),
		},
		&TimeSpan{
			TimeStart: time.Date(2013, 12, 5, 15, 47, 0, 0, time.UTC),
			TimeEnd:   time.Date(2013, 12, 5, 15, 48, 0, 0, time.UTC),
		},
		&TimeSpan{
			TimeStart: time.Date(2013, 12, 5, 15, 48, 0, 0, time.UTC),
			TimeEnd:   time.Date(2013, 12, 5, 15, 49, 0, 0, time.UTC),
		},
	}
	newTss := TimeSpans{
		&TimeSpan{
			TimeStart: time.Date(2013, 12, 5, 15, 45, 0, 0, time.UTC),
			TimeEnd:   time.Date(2013, 12, 5, 15, 49, 30, 0, time.UTC),
		},
	}
	(&tss).OverlapWithTimeSpans(newTss, nil, 0)
	if len(tss) != 1 ||
		tss[0].TimeStart != time.Date(2013, 12, 5, 15, 45, 0, 0, time.UTC) ||
		tss[0].TimeEnd != time.Date(2013, 12, 5, 15, 49, 30, 0, time.UTC) {
		for _, ts := range tss {
			t.Logf("TS: %v", ts)
		}
		t.Error("Error overlaping with timespans timespans: ", tss)
	}
}

func TestOverlapWithTimeSpansOne(t *testing.T) {
	tss := TimeSpans{
		&TimeSpan{
			TimeStart: time.Date(2013, 12, 5, 15, 45, 0, 0, time.UTC),
			TimeEnd:   time.Date(2013, 12, 5, 15, 49, 0, 0, time.UTC),
		},
	}
	newTss := TimeSpans{
		&TimeSpan{
			TimeStart: time.Date(2013, 12, 5, 15, 45, 0, 0, time.UTC),
			TimeEnd:   time.Date(2013, 12, 5, 15, 47, 30, 0, time.UTC),
		},
	}
	(&tss).OverlapWithTimeSpans(newTss, nil, 0)
	if len(tss) != 2 ||
		tss[0].TimeEnd != time.Date(2013, 12, 5, 15, 47, 30, 0, time.UTC) ||
		tss[1].TimeEnd != time.Date(2013, 12, 5, 15, 49, 0, 0, time.UTC) {
		for _, ts := range tss {
			t.Logf("TS: %v", ts)
		}
		t.Error("Error overlaping with timespans timespans: ", tss)
	}
}

func TestIncrementsCompressDecompress(t *testing.T) {
	tss := TimeSpans{
		&TimeSpan{
			Increments: Increments{
				&Increment{
					Duration: time.Minute,
					Cost:     2,
					BalanceInfo: &DebitInfo{
						Unit:      &UnitInfo{UUID: "1", Value: 25, DestinationID: "1", Consumed: 1, ToR: utils.MetaVoice, RateInterval: &RateInterval{Rating: &RIRate{Rates: RateGroups{&RGRate{GroupIntervalStart: 0, Value: 100, RateIncrement: 10 * time.Second, RateUnit: time.Second}}}}},
						Monetary:  &MonetaryInfo{UUID: "2", Value: 98},
						AccountID: "3"},
				},
				&Increment{
					Duration: time.Minute,
					Cost:     2,
					BalanceInfo: &DebitInfo{
						Unit:      &UnitInfo{UUID: "1", Value: 24, DestinationID: "1", Consumed: 1, ToR: utils.MetaVoice, RateInterval: &RateInterval{Rating: &RIRate{Rates: RateGroups{&RGRate{GroupIntervalStart: 0, Value: 100, RateIncrement: 10 * time.Second, RateUnit: time.Second}}}}},
						Monetary:  &MonetaryInfo{UUID: "2", Value: 96},
						AccountID: "3"},
				},
				&Increment{
					Duration: time.Minute,
					Cost:     2,
					BalanceInfo: &DebitInfo{
						Unit:      &UnitInfo{UUID: "1", Value: 23, DestinationID: "1", Consumed: 1, ToR: utils.MetaVoice, RateInterval: &RateInterval{Rating: &RIRate{Rates: RateGroups{&RGRate{GroupIntervalStart: 0, Value: 100, RateIncrement: 10 * time.Second, RateUnit: time.Second}}}}},
						Monetary:  &MonetaryInfo{UUID: "2", Value: 94},
						AccountID: "3"},
				},
				&Increment{
					Duration: time.Minute,
					Cost:     2,
					BalanceInfo: &DebitInfo{
						Unit:      &UnitInfo{UUID: "1", Value: 22, DestinationID: "1", Consumed: 1, ToR: utils.MetaVoice, RateInterval: &RateInterval{Rating: &RIRate{Rates: RateGroups{&RGRate{GroupIntervalStart: 0, Value: 100, RateIncrement: 1111 * time.Second, RateUnit: time.Second}}}}},
						Monetary:  &MonetaryInfo{UUID: "2", Value: 92},
						AccountID: "3"},
				},
				&Increment{
					Duration: time.Minute,
					Cost:     2,
					BalanceInfo: &DebitInfo{
						Unit:      &UnitInfo{UUID: "1", Value: 21, DestinationID: "1", Consumed: 1, ToR: utils.MetaVoice, RateInterval: &RateInterval{Rating: &RIRate{Rates: RateGroups{&RGRate{GroupIntervalStart: 0, Value: 100, RateIncrement: 10 * time.Second, RateUnit: time.Second}}}}},
						Monetary:  &MonetaryInfo{UUID: "2", Value: 90},
						AccountID: "3"},
				},
			},
		},
	}
	tss.Compress()
	if len(tss[0].Increments) != 3 {
		t.Error("Error compressing timespan: ", utils.ToIJSON(tss[0]))
	}
	tss.Decompress()
	if len(tss[0].Increments) != 5 {
		t.Error("Error decompressing timespans: ", utils.ToIJSON(tss[0]))
	}
}

func TestMultipleIncrementsCompressDecompress(t *testing.T) {
	tss := TimeSpans{
		&TimeSpan{
			Increments: Increments{
				&Increment{
					Duration: time.Minute,
					Cost:     10.4,
					BalanceInfo: &DebitInfo{
						Unit:      &UnitInfo{UUID: "1", DestinationID: "1", Consumed: 2.3, ToR: utils.MetaVoice, RateInterval: &RateInterval{Rating: &RIRate{Rates: RateGroups{&RGRate{GroupIntervalStart: 0, Value: 100, RateIncrement: 10 * time.Second, RateUnit: time.Second}}}}},
						Monetary:  &MonetaryInfo{UUID: "2"},
						AccountID: "3"},
				},
				&Increment{
					Duration: time.Minute,
					Cost:     10.4,
					BalanceInfo: &DebitInfo{
						Unit:      &UnitInfo{UUID: "1", DestinationID: "1", Consumed: 2.3, ToR: utils.MetaVoice, RateInterval: &RateInterval{Rating: &RIRate{Rates: RateGroups{&RGRate{GroupIntervalStart: 0, Value: 100, RateIncrement: 10 * time.Second, RateUnit: time.Second}}}}},
						Monetary:  &MonetaryInfo{UUID: "2"},
						AccountID: "3"},
				},
				&Increment{
					Duration: time.Minute,
					Cost:     10.4,
					BalanceInfo: &DebitInfo{
						Unit:      &UnitInfo{UUID: "1", DestinationID: "1", Consumed: 2.3, ToR: utils.MetaVoice, RateInterval: &RateInterval{Rating: &RIRate{Rates: RateGroups{&RGRate{GroupIntervalStart: 0, Value: 100, RateIncrement: 10 * time.Second, RateUnit: time.Second}}}}},
						Monetary:  &MonetaryInfo{UUID: "2"},
						AccountID: "3"},
				},
				&Increment{
					Duration: time.Minute,
					Cost:     10.4,
					BalanceInfo: &DebitInfo{
						Unit:      &UnitInfo{UUID: "1", DestinationID: "1", Consumed: 2.3, ToR: utils.MetaVoice, RateInterval: &RateInterval{Rating: &RIRate{Rates: RateGroups{&RGRate{GroupIntervalStart: 0, Value: 100, RateIncrement: 1111 * time.Second, RateUnit: time.Second}}}}},
						Monetary:  &MonetaryInfo{UUID: "2"},
						AccountID: "3"},
				},
				&Increment{
					Duration: time.Minute,
					Cost:     10.4,
					BalanceInfo: &DebitInfo{
						Unit:      &UnitInfo{UUID: "1", DestinationID: "1", Consumed: 2.3, ToR: utils.MetaVoice, RateInterval: &RateInterval{Rating: &RIRate{Rates: RateGroups{&RGRate{GroupIntervalStart: 0, Value: 100, RateIncrement: 10 * time.Second, RateUnit: time.Second}}}}},
						Monetary:  &MonetaryInfo{UUID: "2"},
						AccountID: "3"},
				},
			},
		},
	}
	tss.Compress()
	if len(tss[0].Increments) != 3 {
		t.Error("Error compressing timespan: ", tss[0].Increments)
	}
	tss.Decompress()
	if len(tss[0].Increments) != 5 {
		t.Error("Error decompressing timespans: ", tss[0].Increments)
	}
}

func TestBetterIntervalZero(t *testing.T) {
	ts := &TimeSpan{
		TimeStart:    time.Date(2015, 4, 27, 8, 0, 0, 0, time.UTC),
		RateInterval: &RateInterval{Timing: &RITiming{StartTime: "08:00:00"}},
	}

	interval := &RateInterval{Timing: &RITiming{StartTime: "00:00:00"}}
	if !ts.hasBetterRateIntervalThan(interval) {
		t.Error("Wrong better rate interval!")
	}
}

func TestBetterIntervalBefore(t *testing.T) {
	ts := &TimeSpan{
		TimeStart:    time.Date(2015, 4, 27, 8, 0, 0, 0, time.UTC),
		RateInterval: &RateInterval{Timing: &RITiming{StartTime: "08:00:00"}},
	}

	interval := &RateInterval{Timing: &RITiming{StartTime: "07:00:00"}}
	if !ts.hasBetterRateIntervalThan(interval) {
		t.Error("Wrong better rate interval!")
	}
}

func TestBetterIntervalEqual(t *testing.T) {
	ts := &TimeSpan{
		TimeStart:    time.Date(2015, 4, 27, 8, 0, 0, 0, time.UTC),
		RateInterval: &RateInterval{Timing: &RITiming{StartTime: "08:00:00"}},
	}

	interval := &RateInterval{Timing: &RITiming{StartTime: "08:00:00"}}
	if !ts.hasBetterRateIntervalThan(interval) {
		t.Error("Wrong better rate interval!")
	}
}

func TestBetterIntervalAfter(t *testing.T) {
	ts := &TimeSpan{
		TimeStart:    time.Date(2015, 4, 27, 8, 0, 0, 0, time.UTC),
		RateInterval: &RateInterval{Timing: &RITiming{StartTime: "08:00:00"}},
	}

	interval := &RateInterval{Timing: &RITiming{StartTime: "13:00:00"}}
	if !ts.hasBetterRateIntervalThan(interval) {
		t.Error("Wrong better rate interval!")
	}
}

func TestBetterIntervalBetter(t *testing.T) {
	ts := &TimeSpan{
		TimeStart:    time.Date(2015, 4, 27, 8, 0, 0, 0, time.UTC),
		RateInterval: &RateInterval{Timing: &RITiming{StartTime: "00:00:00"}},
	}

	interval := &RateInterval{Timing: &RITiming{StartTime: "08:00:00"}}
	if ts.hasBetterRateIntervalThan(interval) {
		t.Error("Wrong better rate interval!")
	}
}

func TestBetterIntervalBetterHour(t *testing.T) {
	ts := &TimeSpan{
		TimeStart:    time.Date(2015, 4, 27, 8, 0, 0, 0, time.UTC),
		RateInterval: &RateInterval{Timing: &RITiming{StartTime: "00:00:00"}},
	}

	interval := &RateInterval{Timing: &RITiming{StartTime: "06:00:00"}}
	if ts.hasBetterRateIntervalThan(interval) {
		t.Error("Wrong better rate interval!")
	}
}

func TestBetterIntervalAgainAfter(t *testing.T) {
	ts := &TimeSpan{
		TimeStart:    time.Date(2015, 4, 27, 8, 0, 0, 0, time.UTC),
		RateInterval: &RateInterval{Timing: &RITiming{StartTime: "00:00:00"}},
	}

	interval := &RateInterval{Timing: &RITiming{StartTime: "13:00:00"}}
	if !ts.hasBetterRateIntervalThan(interval) {
		t.Error("Wrong better rate interval!")
	}
}

func TestCompressDecompress(t *testing.T) {
	tss := TimeSpans{
		&TimeSpan{
			TimeStart:     time.Date(2015, 1, 9, 16, 18, 0, 0, time.UTC),
			TimeEnd:       time.Date(2015, 1, 9, 16, 19, 0, 0, time.UTC),
			Cost:          1.2,
			DurationIndex: time.Minute,
		},
		&TimeSpan{
			TimeStart:     time.Date(2015, 1, 9, 16, 19, 0, 0, time.UTC),
			TimeEnd:       time.Date(2015, 1, 9, 16, 20, 0, 0, time.UTC),
			Cost:          1.2,
			DurationIndex: 2 * time.Minute,
		},
		&TimeSpan{
			TimeStart:     time.Date(2015, 1, 9, 16, 20, 0, 0, time.UTC),
			TimeEnd:       time.Date(2015, 1, 9, 16, 21, 0, 0, time.UTC),
			Cost:          1.2,
			DurationIndex: 3 * time.Minute,
		},
		&TimeSpan{
			TimeStart:     time.Date(2015, 1, 9, 16, 21, 0, 0, time.UTC),
			TimeEnd:       time.Date(2015, 1, 9, 16, 22, 0, 0, time.UTC),
			Cost:          1.2,
			DurationIndex: 4 * time.Minute,
		},
	}
	tss.Compress()
	if len(tss) != 1 ||
		!tss[0].TimeStart.Equal(time.Date(2015, 1, 9, 16, 18, 0, 0, time.UTC)) ||
		!tss[0].TimeEnd.Equal(time.Date(2015, 1, 9, 16, 22, 0, 0, time.UTC)) ||
		tss[0].DurationIndex != 4*time.Minute ||
		tss[0].Cost != 4.8 ||
		tss[0].CompressFactor != 4 {
		for _, ts := range tss {
			t.Logf("TS: %+v", ts)
		}
		t.Error("Error compressing timespans: ", tss)
	}
	tss.Decompress()
	if len(tss) != 4 ||
		!tss[0].TimeStart.Equal(time.Date(2015, 1, 9, 16, 18, 0, 0, time.UTC)) ||
		!tss[0].TimeEnd.Equal(time.Date(2015, 1, 9, 16, 19, 0, 0, time.UTC)) ||
		tss[0].DurationIndex != time.Minute ||
		tss[0].CompressFactor != 1 ||
		tss[0].Cost != 1.2 ||
		!tss[1].TimeStart.Equal(time.Date(2015, 1, 9, 16, 19, 0, 0, time.UTC)) ||
		!tss[1].TimeEnd.Equal(time.Date(2015, 1, 9, 16, 20, 0, 0, time.UTC)) ||
		tss[1].DurationIndex != 2*time.Minute ||
		tss[1].CompressFactor != 1 ||
		tss[1].Cost != 1.2 ||
		!tss[2].TimeStart.Equal(time.Date(2015, 1, 9, 16, 20, 0, 0, time.UTC)) ||
		!tss[2].TimeEnd.Equal(time.Date(2015, 1, 9, 16, 21, 0, 0, time.UTC)) ||
		tss[2].DurationIndex != 3*time.Minute ||
		tss[2].CompressFactor != 1 ||
		tss[2].Cost != 1.2 ||
		!tss[3].TimeStart.Equal(time.Date(2015, 1, 9, 16, 21, 0, 0, time.UTC)) ||
		!tss[3].TimeEnd.Equal(time.Date(2015, 1, 9, 16, 22, 0, 0, time.UTC)) ||
		tss[3].DurationIndex != 4*time.Minute ||
		tss[3].CompressFactor != 1 ||
		tss[3].Cost != 1.2 {
		for i, ts := range tss {
			t.Logf("TS(%d): %+v", i, ts)
		}
		t.Error("Error decompressing timespans: ", tss)
	}
}

func TestDifferentCompressDecompress(t *testing.T) {
	tss := TimeSpans{
		&TimeSpan{
			TimeStart:     time.Date(2015, 1, 9, 16, 18, 0, 0, time.UTC),
			TimeEnd:       time.Date(2015, 1, 9, 16, 19, 0, 0, time.UTC),
			RateInterval:  &RateInterval{Weight: 1},
			Cost:          1.2,
			DurationIndex: time.Minute,
		},
		&TimeSpan{
			TimeStart:     time.Date(2015, 1, 9, 16, 19, 0, 0, time.UTC),
			TimeEnd:       time.Date(2015, 1, 9, 16, 20, 0, 0, time.UTC),
			RateInterval:  &RateInterval{Weight: 2},
			Cost:          1.2,
			DurationIndex: 2 * time.Minute,
		},
		&TimeSpan{
			TimeStart:     time.Date(2015, 1, 9, 16, 20, 0, 0, time.UTC),
			TimeEnd:       time.Date(2015, 1, 9, 16, 21, 0, 0, time.UTC),
			RateInterval:  &RateInterval{Weight: 1},
			Cost:          1.2,
			DurationIndex: 3 * time.Minute,
		},
		&TimeSpan{
			TimeStart:     time.Date(2015, 1, 9, 16, 21, 0, 0, time.UTC),
			TimeEnd:       time.Date(2015, 1, 9, 16, 22, 0, 0, time.UTC),
			RateInterval:  &RateInterval{Weight: 1},
			Cost:          1.2,
			DurationIndex: 4 * time.Minute,
		},
	}
	tss.Compress()
	if len(tss) != 3 ||
		!tss[0].TimeStart.Equal(time.Date(2015, 1, 9, 16, 18, 0, 0, time.UTC)) ||
		!tss[0].TimeEnd.Equal(time.Date(2015, 1, 9, 16, 19, 0, 0, time.UTC)) ||
		tss[0].DurationIndex != time.Minute ||
		tss[0].Cost != 1.2 ||
		!tss[1].TimeStart.Equal(time.Date(2015, 1, 9, 16, 19, 0, 0, time.UTC)) ||
		!tss[1].TimeEnd.Equal(time.Date(2015, 1, 9, 16, 20, 0, 0, time.UTC)) ||
		tss[1].DurationIndex != 2*time.Minute ||
		tss[1].Cost != 1.2 ||
		!tss[2].TimeStart.Equal(time.Date(2015, 1, 9, 16, 20, 0, 0, time.UTC)) ||
		!tss[2].TimeEnd.Equal(time.Date(2015, 1, 9, 16, 22, 0, 0, time.UTC)) ||
		tss[2].DurationIndex != 4*time.Minute ||
		tss[2].Cost != 2.4 {
		for _, ts := range tss {
			t.Logf("TS: %+v", ts)
		}
		t.Error("Error compressing timespans: ", tss)
	}
	tss.Decompress()
	if len(tss) != 4 ||
		!tss[0].TimeStart.Equal(time.Date(2015, 1, 9, 16, 18, 0, 0, time.UTC)) ||
		!tss[0].TimeEnd.Equal(time.Date(2015, 1, 9, 16, 19, 0, 0, time.UTC)) ||
		tss[0].DurationIndex != time.Minute ||
		tss[0].CompressFactor != 1 ||
		tss[0].Cost != 1.2 ||
		!tss[1].TimeStart.Equal(time.Date(2015, 1, 9, 16, 19, 0, 0, time.UTC)) ||
		!tss[1].TimeEnd.Equal(time.Date(2015, 1, 9, 16, 20, 0, 0, time.UTC)) ||
		tss[1].DurationIndex != 2*time.Minute ||
		tss[1].CompressFactor != 1 ||
		tss[1].Cost != 1.2 ||
		!tss[2].TimeStart.Equal(time.Date(2015, 1, 9, 16, 20, 0, 0, time.UTC)) ||
		!tss[2].TimeEnd.Equal(time.Date(2015, 1, 9, 16, 21, 0, 0, time.UTC)) ||
		tss[2].DurationIndex != 3*time.Minute ||
		tss[2].CompressFactor != 1 ||
		tss[2].Cost != 1.2 ||
		!tss[3].TimeStart.Equal(time.Date(2015, 1, 9, 16, 21, 0, 0, time.UTC)) ||
		!tss[3].TimeEnd.Equal(time.Date(2015, 1, 9, 16, 22, 0, 0, time.UTC)) ||
		tss[3].DurationIndex != 4*time.Minute ||
		tss[3].CompressFactor != 1 ||
		tss[3].Cost != 1.2 {
		for i, ts := range tss {
			t.Logf("TS(%d): %+v", i, ts)
		}
		t.Error("Error decompressing timespans: ", tss)
	}
}

func TestMerge(t *testing.T) {
	tss1 := &TimeSpan{
		TimeStart: time.Date(2015, 1, 9, 16, 18, 0, 0, time.UTC),
		TimeEnd:   time.Date(2015, 1, 9, 16, 19, 0, 0, time.UTC),
		RateInterval: &RateInterval{
			Rating: &RIRate{
				RoundingMethod:   utils.MetaRoundingMiddle,
				RoundingDecimals: 2,
				Rates: RateGroups{
					&RGRate{
						Value:         2.0,
						RateIncrement: 10 * time.Second,
					},
				},
			},
		},
		Cost:          3,
		DurationIndex: time.Minute,
		Increments: Increments{
			&Increment{
				Duration: time.Minute,
				Cost:     1,
				BalanceInfo: &DebitInfo{
					Unit:      &UnitInfo{UUID: "1", DestinationID: "1", Consumed: 2.3, ToR: utils.MetaVoice, RateInterval: &RateInterval{Rating: &RIRate{Rates: RateGroups{&RGRate{GroupIntervalStart: 0, Value: 100, RateIncrement: 10 * time.Second, RateUnit: time.Second}}}}},
					Monetary:  &MonetaryInfo{UUID: "2"},
					AccountID: "3"},
				CompressFactor: 3,
			},
		},
	}
	tss2 := &TimeSpan{
		TimeStart: time.Date(2015, 1, 9, 16, 19, 0, 0, time.UTC),
		TimeEnd:   time.Date(2015, 1, 9, 16, 20, 0, 0, time.UTC),
		RateInterval: &RateInterval{
			Rating: &RIRate{
				RoundingMethod:   utils.MetaRoundingMiddle,
				RoundingDecimals: 2,
				Rates: RateGroups{
					&RGRate{
						Value:         2.0,
						RateIncrement: 10 * time.Second,
					},
				},
			},
		},
		Cost:          2,
		DurationIndex: 2 * time.Minute,
		Increments: Increments{
			&Increment{
				Duration: time.Minute,
				Cost:     1,
				BalanceInfo: &DebitInfo{
					Unit:      &UnitInfo{UUID: "1", DestinationID: "1", Consumed: 2.3, ToR: utils.MetaVoice, RateInterval: &RateInterval{Rating: &RIRate{Rates: RateGroups{&RGRate{GroupIntervalStart: 0, Value: 100, RateIncrement: 10 * time.Second, RateUnit: time.Second}}}}},
					Monetary:  &MonetaryInfo{UUID: "2"},
					AccountID: "3"},
			},
			&Increment{
				Duration: time.Minute,
				Cost:     1,
				BalanceInfo: &DebitInfo{
					Unit:      &UnitInfo{UUID: "1", DestinationID: "1", Consumed: 2.3, ToR: utils.MetaVoice, RateInterval: &RateInterval{Rating: &RIRate{Rates: RateGroups{&RGRate{GroupIntervalStart: 0, Value: 100, RateIncrement: 10 * time.Second, RateUnit: time.Second}}}}},
					Monetary:  &MonetaryInfo{UUID: "2"},
					AccountID: "3"},
			},
		},
	}
	eMergedTSS := &TimeSpan{
		TimeStart: time.Date(2015, 1, 9, 16, 18, 0, 0, time.UTC),
		TimeEnd:   time.Date(2015, 1, 9, 16, 20, 0, 0, time.UTC),
		RateInterval: &RateInterval{
			Rating: &RIRate{
				RoundingMethod:   utils.MetaRoundingMiddle,
				RoundingDecimals: 2,
				Rates: RateGroups{
					&RGRate{
						Value:         2.0,
						RateIncrement: 10 * time.Second,
					},
				},
			},
		},
		Cost:          5,
		DurationIndex: 2 * time.Minute,
		Increments: Increments{
			&Increment{
				Duration: time.Minute,
				Cost:     1,
				BalanceInfo: &DebitInfo{
					Unit:      &UnitInfo{UUID: "1", DestinationID: "1", Consumed: 2.3, ToR: utils.MetaVoice, RateInterval: &RateInterval{Rating: &RIRate{Rates: RateGroups{&RGRate{GroupIntervalStart: 0, Value: 100, RateIncrement: 10 * time.Second, RateUnit: time.Second}}}}},
					Monetary:  &MonetaryInfo{UUID: "2"},
					AccountID: "3"},
				CompressFactor: 3,
			},
			&Increment{
				Duration: time.Minute,
				Cost:     1,
				BalanceInfo: &DebitInfo{
					Unit:      &UnitInfo{UUID: "1", DestinationID: "1", Consumed: 2.3, ToR: utils.MetaVoice, RateInterval: &RateInterval{Rating: &RIRate{Rates: RateGroups{&RGRate{GroupIntervalStart: 0, Value: 100, RateIncrement: 10 * time.Second, RateUnit: time.Second}}}}},
					Monetary:  &MonetaryInfo{UUID: "2"},
					AccountID: "3"},
				CompressFactor: 1,
			},
			&Increment{
				Duration: time.Minute,
				Cost:     1,
				BalanceInfo: &DebitInfo{
					Unit:      &UnitInfo{UUID: "1", DestinationID: "1", Consumed: 2.3, ToR: utils.MetaVoice, RateInterval: &RateInterval{Rating: &RIRate{Rates: RateGroups{&RGRate{GroupIntervalStart: 0, Value: 100, RateIncrement: 10 * time.Second, RateUnit: time.Second}}}}},
					Monetary:  &MonetaryInfo{UUID: "2"},
					AccountID: "3"},
				CompressFactor: 1,
			},
		},
		CompressFactor: 1,
	}
	if merged := tss1.Merge(tss2); !merged {
		t.Error("Not merged")
	} else if !tss1.Equal(eMergedTSS) {
		t.Errorf("Expecting: %+v, received: %+v", eMergedTSS, tss1)
	}
	tss1.TimeEnd = time.Date(2015, 1, 9, 16, 20, 0, 0, time.UTC)
	if merged := tss1.Merge(tss2); merged {
		t.Error("expected false")
	}
	tss1.MatchedSubject = "match_subj1"
	if merged := tss1.Merge(tss2); merged {
		t.Error("expected false")
	}
}

func TestIncrementClone(t *testing.T) {
	incr := &Increment{}
	eOut := &Increment{}
	if clone := incr.Clone(); !reflect.DeepEqual(eOut, clone) {
		t.Errorf("Expecting %+v, received: %+v", utils.ToJSON(eOut), utils.ToJSON(clone))
	}
	incr = &Increment{
		Duration:       10,
		Cost:           0.7,
		CompressFactor: 10,
		BalanceInfo:    &DebitInfo{AccountID: "AccountID_test"},
	}
	eOut = &Increment{
		Duration:       10,
		Cost:           0.7,
		CompressFactor: 10,
		BalanceInfo:    &DebitInfo{AccountID: "AccountID_test"},
	}
	if clone := incr.Clone(); !reflect.DeepEqual(eOut, clone) {
		t.Errorf("Expecting %+v, received: %+v", utils.ToJSON(eOut), utils.ToJSON(clone))
	}
}

func TestTimeSpansMerge(t *testing.T) {

	tss := &TimeSpans{
		{
			CompressFactor: 4,
			DurationIndex:  1 * time.Hour,
			Cost:           12,
			TimeStart:      time.Date(2022, 12, 1, 18, 0, 0, 0, time.UTC),
			TimeEnd:        time.Date(2022, 12, 1, 19, 0, 0, 0, time.UTC),
			Increments: Increments{
				&Increment{
					Duration: 2 * time.Minute,
					Cost:     23.22,
				},
				{
					Duration: 5 * time.Minute,
					Cost:     12.32,
				},
			},
			RateInterval: &RateInterval{
				Timing: &RITiming{
					ID:    "id",
					Years: utils.Years{2, 1, 3},
				},
				Rating: &RIRate{
					ConnectFee:       12.11,
					RoundingMethod:   "method",
					RoundingDecimals: 13,
					MaxCost:          494,
				},
				Weight: 2.3,
			},
			MatchedSubject: "subject",
			MatchedPrefix:  "match_prefix",
			MatchedDestId:  "dest_id",
			RatingPlanId:   "rate_id",
		},
		{
			CompressFactor: 4,
			DurationIndex:  1 * time.Hour,
			Cost:           12,
			TimeStart:      time.Date(2022, 12, 1, 19, 0, 0, 0, time.UTC),
			TimeEnd:        time.Date(2022, 12, 1, 20, 0, 0, 0, time.UTC),
			Increments: Increments{
				&Increment{
					Duration: 2 * time.Minute,
					Cost:     23.22,
				},
				{
					Duration: 5 * time.Minute,
					Cost:     12.32,
				},
			},
			RateInterval: &RateInterval{
				Timing: &RITiming{
					ID:    "id",
					Years: utils.Years{2, 1, 3},
				},
				Rating: &RIRate{
					ConnectFee:       12.11,
					RoundingMethod:   "method",
					RoundingDecimals: 13,
					MaxCost:          494,
				},
				Weight: 2.3,
			},
			MatchedSubject: "subject",
			MatchedPrefix:  "match_prefix",
			MatchedDestId:  "dest_id",
			RatingPlanId:   "rate_id",
		}, {
			Cost:           11,
			DurationIndex:  1 * time.Hour,
			CompressFactor: 4,
			TimeStart:      time.Date(2022, 12, 1, 20, 0, 0, 0, time.UTC),
			TimeEnd:        time.Date(2022, 12, 1, 21, 0, 0, 0, time.UTC),
			Increments: Increments{
				&Increment{
					Duration: 2 * time.Minute,
					Cost:     23.22,
				},
				{
					Duration: 5 * time.Minute,
					Cost:     12.32,
				},
			},
			RateInterval: &RateInterval{
				Timing: &RITiming{
					ID:    "id",
					Years: utils.Years{2, 1, 3},
				},
				Rating: &RIRate{
					ConnectFee:       12.11,
					RoundingMethod:   "method",
					RoundingDecimals: 13,
					MaxCost:          494,
				},
				Weight: 2.3,
			},
			MatchedSubject: "subject",
			MatchedPrefix:  "match_prefix",
			MatchedDestId:  "dest_id",
			RatingPlanId:   "rate_id",
		},
	}
	expMerge := &TimeSpans{{
		CompressFactor: 4,
		DurationIndex:  1 * time.Hour,
		Cost:           35,
		TimeStart:      time.Date(2022, 12, 1, 18, 0, 0, 0, time.UTC),
		TimeEnd:        time.Date(2022, 12, 1, 21, 0, 0, 0, time.UTC),
		Increments: Increments{
			{
				Duration: 2 * time.Minute,
				Cost:     23.22,
			},
			{
				Duration: 5 * time.Minute,
				Cost:     12.32,
			},
			{
				Duration: 2 * time.Minute,
				Cost:     23.22,
			},
			{
				Duration: 5 * time.Minute,
				Cost:     12.32,
			},
			{
				Duration: 2 * time.Minute,
				Cost:     23.22,
			},
			{
				Duration: 5 * time.Minute,
				Cost:     12.32,
			},
		},
		RateInterval: &RateInterval{
			Timing: &RITiming{
				ID:    "id",
				Years: utils.Years{2, 1, 3},
			},
			Rating: &RIRate{
				ConnectFee:       12.11,
				RoundingMethod:   "method",
				RoundingDecimals: 13,
				MaxCost:          494,
			},
			Weight: 2.3,
		},
		MatchedSubject: "subject",
		MatchedPrefix:  "match_prefix",
		MatchedDestId:  "dest_id",
		RatingPlanId:   "rate_id",
	}}
	if tss.Merge(); !reflect.DeepEqual(tss, expMerge) {
		t.Errorf("expected %v ,recived %v", utils.ToJSON(expMerge), utils.ToJSON(tss))
	}
}

func TestMIUIEqualFalse(t *testing.T) {
	mi := &MonetaryInfo{
		UUID:  "uuid",
		ID:    "id",
		Value: 23.1,
	}
	ui := &UnitInfo{
		UUID:          "uuid",
		ID:            "id",
		Value:         12.2,
		DestinationID: "destId",
	}
	if val := mi.Equal(nil); val {
		t.Errorf("expected false ,received %+v", val)
	} else if val = ui.Equal(nil); val {
		t.Errorf("expected false,received %+v", val)
	}

}
