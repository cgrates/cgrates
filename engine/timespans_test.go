/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2012-2015 ITsysCOM

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

	"github.com/cgrates/cgrates/utils"
)

func TestTSRightMargin(t *testing.T) {
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

func TestTSSplitMiddle(t *testing.T) {
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

func TestTSRightHourMargin(t *testing.T) {
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

	if ts.GetDuration() != 29*time.Minute || nts.GetDuration() != 1*time.Minute {
		t.Error("Wrong durations.for RateIntervals", ts.GetDuration(), nts.GetDuration())
	}
	if ts.GetDuration().Seconds()+nts.GetDuration().Seconds() != oldDuration.Seconds() {
		t.Errorf("The duration has changed: %v + %v != %v", ts.GetDuration().Seconds(), nts.GetDuration().Seconds(), oldDuration.Seconds())
	}
}

func TestTSLeftMargin(t *testing.T) {
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

func TestTSLeftHourMargin(t *testing.T) {
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

func TestTSEnclosingMargin(t *testing.T) {
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

func TestTSOutsideMargin(t *testing.T) {
	i := &RateInterval{Timing: &RITiming{WeekDays: []time.Weekday{time.Monday}}}
	t1 := time.Date(2012, time.February, 5, 17, 45, 0, 0, time.UTC)
	t2 := time.Date(2012, time.February, 5, 18, 10, 0, 0, time.UTC)
	ts := &TimeSpan{TimeStart: t1, TimeEnd: t2}
	result := ts.SplitByRateInterval(i, false)
	if result != nil {
		t.Error("RateInterval not split correctly")
	}
}

func TestTSContains(t *testing.T) {
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

func TestTSSplitByRatingPlan(t *testing.T) {
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

func TestTSTimespanGetCost(t *testing.T) {
	t1 := time.Date(2012, time.February, 5, 17, 45, 0, 0, time.UTC)
	t2 := time.Date(2012, time.February, 5, 17, 55, 0, 0, time.UTC)
	ts1 := TimeSpan{TimeStart: t1, TimeEnd: t2}
	if ts1.calculateCost() != 0 {
		t.Error("No interval and still kicking")
	}
	ts1.SetRateInterval(
		&RateInterval{
			Timing: &RITiming{},
			Rating: &RIRate{Rates: RateGroups{&Rate{0, 1.0, 1 * time.Second, 1 * time.Second}}},
		},
	)
	if ts1.calculateCost() != 600 {
		t.Error("Expected 10 got ", ts1.Cost)
	}
	ts1.RateInterval = nil
	ts1.SetRateInterval(&RateInterval{Rating: &RIRate{Rates: RateGroups{&Rate{0, 1.0, 1 * time.Second, 60 * time.Second}}}})
	if ts1.calculateCost() != 10 {
		t.Error("Expected 6000 got ", ts1.Cost)
	}
}

func TestTSTimespanGetCostIntervals(t *testing.T) {
	ts := &TimeSpan{}
	ts.Increments = make(Increments, 11)
	for i := 0; i < 11; i++ {
		ts.Increments[i] = &Increment{Cost: 0.02}
	}
	if ts.calculateCost() != 0.22 {
		t.Error("Error caclulating timespan cost: ", ts.calculateCost())
	}
}

func TestTSSetRateInterval(t *testing.T) {
	i1 := &RateInterval{
		Timing: &RITiming{},
		Rating: &RIRate{Rates: RateGroups{&Rate{0, 1.0, 1 * time.Second, 1 * time.Second}}},
	}
	ts1 := TimeSpan{RateInterval: i1}
	i2 := &RateInterval{
		Timing: &RITiming{},
		Rating: &RIRate{Rates: RateGroups{&Rate{0, 2.0, 1 * time.Second, 1 * time.Second}}},
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

func TestTSTimespanSplitGroupedRates(t *testing.T) {
	i := &RateInterval{
		Timing: &RITiming{
			EndTime: "17:59:00",
		},
		Rating: &RIRate{
			Rates: RateGroups{&Rate{0, 2, 1 * time.Second, 1 * time.Second}, &Rate{900 * time.Second, 1, 1 * time.Second, 1 * time.Second}},
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

func TestTSTimespanSplitGroupedRatesIncrements(t *testing.T) {
	i := &RateInterval{
		Timing: &RITiming{
			EndTime: "17:59:00",
		},
		Rating: &RIRate{
			Rates: RateGroups{
				&Rate{
					GroupIntervalStart: 0,
					Value:              2,
					RateIncrement:      time.Second,
					RateUnit:           time.Second},
				&Rate{
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

	if ts.GetDuration().Seconds() != 0.5*60 || nts.GetDuration().Seconds() != 1*60 {
		t.Error("Wrong durations.for RateIntervals", ts.GetDuration().Seconds(), nts.GetDuration().Seconds())
	}
	if ts.GetDuration()+nts.GetDuration() != oldDuration+30*time.Second {
		t.Errorf("The duration has changed: %v + %v != %v", ts.GetDuration().Seconds(), nts.GetDuration().Seconds(), oldDuration.Seconds())
	}
}

func TestTSTimespanSplitRightHourMarginBeforeGroup(t *testing.T) {
	i := &RateInterval{
		Timing: &RITiming{
			EndTime: "17:00:30",
		},
		Rating: &RIRate{
			Rates: RateGroups{&Rate{0, 2, 1 * time.Second, 1 * time.Second}, &Rate{60 * time.Second, 1, 60 * time.Second, 1 * time.Second}},
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

func TestTSTimespanSplitGroupSecondSplit(t *testing.T) {
	i := &RateInterval{
		Timing: &RITiming{
			EndTime: "17:03:30",
		},
		Rating: &RIRate{
			Rates: RateGroups{&Rate{0, 2, 1 * time.Second, 1 * time.Second}, &Rate{60 * time.Second, 1, 1 * time.Second, 1 * time.Second}}},
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

func TestTSTimespanSplitLong(t *testing.T) {
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

func TestTSTimespanSplitMultipleGroup(t *testing.T) {
	i := &RateInterval{
		Timing: &RITiming{
			EndTime: "17:05:00",
		},
		Rating: &RIRate{
			Rates: RateGroups{&Rate{0, 2, 1 * time.Second, 1 * time.Second}, &Rate{60 * time.Second, 1, 1 * time.Second, 1 * time.Second}, &Rate{180 * time.Second, 1, 1 * time.Second, 1 * time.Second}}},
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

func TestTSTimespanExpandingPastEnd(t *testing.T) {
	timespans := []*TimeSpan{
		&TimeSpan{
			TimeStart: time.Date(2013, 9, 10, 14, 30, 0, 0, time.UTC),
			TimeEnd:   time.Date(2013, 9, 10, 14, 30, 30, 0, time.UTC),
			RateInterval: &RateInterval{Rating: &RIRate{Rates: RateGroups{
				&Rate{RateIncrement: 60 * time.Second},
			}}},
		},
		&TimeSpan{
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

func TestTSTimespanExpandingDurationIndex(t *testing.T) {
	timespans := []*TimeSpan{
		&TimeSpan{
			TimeStart: time.Date(2013, 9, 10, 14, 30, 0, 0, time.UTC),
			TimeEnd:   time.Date(2013, 9, 10, 14, 30, 30, 0, time.UTC),
			RateInterval: &RateInterval{Rating: &RIRate{Rates: RateGroups{
				&Rate{RateIncrement: 60 * time.Second},
			}}},
		},
		&TimeSpan{
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

func TestTSTimespanExpandingRoundingPastEnd(t *testing.T) {
	timespans := []*TimeSpan{
		&TimeSpan{
			TimeStart: time.Date(2013, 9, 10, 14, 30, 0, 0, time.UTC),
			TimeEnd:   time.Date(2013, 9, 10, 14, 30, 20, 0, time.UTC),
			RateInterval: &RateInterval{Rating: &RIRate{Rates: RateGroups{
				&Rate{RateIncrement: 15 * time.Second},
			}}},
		},
		&TimeSpan{
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

func TestTSTimespanExpandingPastEndMultiple(t *testing.T) {
	timespans := []*TimeSpan{
		&TimeSpan{
			TimeStart: time.Date(2013, 9, 10, 14, 30, 0, 0, time.UTC),
			TimeEnd:   time.Date(2013, 9, 10, 14, 30, 30, 0, time.UTC),
			RateInterval: &RateInterval{Rating: &RIRate{Rates: RateGroups{
				&Rate{RateIncrement: 60 * time.Second},
			}}},
		},
		&TimeSpan{
			TimeStart: time.Date(2013, 9, 10, 14, 30, 30, 0, time.UTC),
			TimeEnd:   time.Date(2013, 9, 10, 14, 30, 40, 0, time.UTC),
		},
		&TimeSpan{
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

func TestTSTimespanExpandingPastEndMultipleEqual(t *testing.T) {
	timespans := []*TimeSpan{
		&TimeSpan{
			TimeStart: time.Date(2013, 9, 10, 14, 30, 0, 0, time.UTC),
			TimeEnd:   time.Date(2013, 9, 10, 14, 30, 30, 0, time.UTC),
			RateInterval: &RateInterval{Rating: &RIRate{Rates: RateGroups{
				&Rate{RateIncrement: 60 * time.Second},
			}}},
		},
		&TimeSpan{
			TimeStart: time.Date(2013, 9, 10, 14, 30, 30, 0, time.UTC),
			TimeEnd:   time.Date(2013, 9, 10, 14, 30, 40, 0, time.UTC),
		},
		&TimeSpan{
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

func TestTSTimespanExpandingBeforeEnd(t *testing.T) {
	timespans := []*TimeSpan{
		&TimeSpan{
			TimeStart: time.Date(2013, 9, 10, 14, 30, 0, 0, time.UTC),
			TimeEnd:   time.Date(2013, 9, 10, 14, 30, 30, 0, time.UTC),
			RateInterval: &RateInterval{Rating: &RIRate{Rates: RateGroups{
				&Rate{RateIncrement: 45 * time.Second},
			}}},
		},
		&TimeSpan{
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

func TestTSTimespanExpandingBeforeEndMultiple(t *testing.T) {
	timespans := []*TimeSpan{
		&TimeSpan{
			TimeStart: time.Date(2013, 9, 10, 14, 30, 0, 0, time.UTC),
			TimeEnd:   time.Date(2013, 9, 10, 14, 30, 30, 0, time.UTC),
			RateInterval: &RateInterval{Rating: &RIRate{Rates: RateGroups{
				&Rate{RateIncrement: 45 * time.Second},
			}}},
		},
		&TimeSpan{
			TimeStart: time.Date(2013, 9, 10, 14, 30, 30, 0, time.UTC),
			TimeEnd:   time.Date(2013, 9, 10, 14, 30, 50, 0, time.UTC),
		},
		&TimeSpan{
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

func TestTSTimespanCreateSecondsSlice(t *testing.T) {
	ts := &TimeSpan{
		TimeStart: time.Date(2013, 9, 10, 14, 30, 0, 0, time.UTC),
		TimeEnd:   time.Date(2013, 9, 10, 14, 30, 30, 0, time.UTC),
		RateInterval: &RateInterval{Rating: &RIRate{Rates: RateGroups{
			&Rate{Value: 2.0},
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

func TestTSTimespanCreateIncrements(t *testing.T) {
	ts := &TimeSpan{
		TimeStart: time.Date(2013, 9, 10, 14, 30, 0, 0, time.UTC),
		TimeEnd:   time.Date(2013, 9, 10, 14, 30, 30, 100000000, time.UTC),
		RateInterval: &RateInterval{
			Rating: &RIRate{
				RoundingMethod:   utils.ROUNDING_MIDDLE,
				RoundingDecimals: 2,
				Rates: RateGroups{
					&Rate{
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
	if len(ts.Increments) < 3 || ts.Increments[2].Cost != 20.07 {
		t.Error("Wrong second slice: ", ts.Increments[2].Cost)
	}
}

func TestTSTimespanSplitByIncrement(t *testing.T) {
	ts := &TimeSpan{
		TimeStart:     time.Date(2013, 9, 19, 18, 30, 0, 0, time.UTC),
		TimeEnd:       time.Date(2013, 9, 19, 18, 31, 00, 0, time.UTC),
		DurationIndex: 60 * time.Second,
		ratingInfo:    &RatingInfo{},
		RateInterval: &RateInterval{
			Rating: &RIRate{
				RoundingMethod:   utils.ROUNDING_MIDDLE,
				RoundingDecimals: 2,
				Rates: RateGroups{
					&Rate{
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

func TestTSTimespanSplitByIncrementStart(t *testing.T) {
	ts := &TimeSpan{
		TimeStart:     time.Date(2013, 9, 19, 18, 30, 0, 0, time.UTC),
		TimeEnd:       time.Date(2013, 9, 19, 18, 31, 00, 0, time.UTC),
		DurationIndex: 60 * time.Second,
		RateInterval: &RateInterval{
			Rating: &RIRate{
				RoundingMethod:   utils.ROUNDING_MIDDLE,
				RoundingDecimals: 2,
				Rates: RateGroups{
					&Rate{
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

func TestTSTimespanSplitByIncrementEnd(t *testing.T) {
	ts := &TimeSpan{
		TimeStart:     time.Date(2013, 9, 19, 18, 30, 0, 0, time.UTC),
		TimeEnd:       time.Date(2013, 9, 19, 18, 31, 00, 0, time.UTC),
		DurationIndex: 60 * time.Second,
		RateInterval: &RateInterval{
			Rating: &RIRate{
				RoundingMethod:   utils.ROUNDING_MIDDLE,
				RoundingDecimals: 2,
				Rates: RateGroups{
					&Rate{
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

func TestTSTimespanSplitByDuration(t *testing.T) {
	ts := &TimeSpan{
		TimeStart:     time.Date(2013, 9, 19, 18, 30, 0, 0, time.UTC),
		TimeEnd:       time.Date(2013, 9, 19, 18, 31, 00, 0, time.UTC),
		DurationIndex: 60 * time.Second,
		ratingInfo:    &RatingInfo{},
		RateInterval: &RateInterval{
			Rating: &RIRate{
				RoundingMethod:   utils.ROUNDING_MIDDLE,
				RoundingDecimals: 2,
				Rates: RateGroups{
					&Rate{
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

func TestTSRemoveOverlapedFromIndexMiddle(t *testing.T) {
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

func TestTSRemoveOverlapedFromIndexMiddleNonBounds(t *testing.T) {
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

func TestTSRemoveOverlapedFromIndexMiddleNonBoundsOver(t *testing.T) {
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

func TestTSRemoveOverlapedFromIndexEnd(t *testing.T) {
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

func TestTSRemoveOverlapedFromIndexEndPast(t *testing.T) {
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

func TestTSRemoveOverlapedFromIndexAll(t *testing.T) {
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

func TestTSRemoveOverlapedFromIndexNone(t *testing.T) {
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

func TestTSRemoveOverlapedFromIndexOne(t *testing.T) {
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

func TestTSRemoveOverlapedFromIndexTwo(t *testing.T) {
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

func TestTSOverlapWithTimeSpansMiddleLong(t *testing.T) {
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

func TestTSOverlapWithTimeSpansMiddleMedium(t *testing.T) {
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

func TestTSOverlapWithTimeSpansMiddleShort(t *testing.T) {
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

func TestTSOverlapWithTimeSpansStart(t *testing.T) {
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

func TestTSOverlapWithTimeSpansAlmostEnd(t *testing.T) {
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

func TestTSOverlapWithTimeSpansEnd(t *testing.T) {
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

func TestTSOverlapWithTimeSpansPastEnd(t *testing.T) {
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

func TestTSOverlapWithTimeSpansAll(t *testing.T) {
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

func TestTSOverlapWithTimeSpansAllPast(t *testing.T) {
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

func TestTSOverlapWithTimeSpansOne(t *testing.T) {
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

func TestTSIncrementsCompressDecompress(t *testing.T) {
	tss := TimeSpans{
		&TimeSpan{
			Increments: Increments{
				&Increment{
					Duration:            time.Minute,
					Cost:                10.4,
					BalanceInfo:         &BalanceInfo{"1", "2", "3"},
					BalanceRateInterval: &RateInterval{Rating: &RIRate{Rates: RateGroups{&Rate{GroupIntervalStart: 0, Value: 100, RateIncrement: 10 * time.Second, RateUnit: time.Second}}}},
					UnitInfo:            &UnitInfo{"1", 2.3, utils.VOICE},
				},
				&Increment{
					Duration:            time.Minute,
					Cost:                10.4,
					BalanceInfo:         &BalanceInfo{"1", "2", "3"},
					BalanceRateInterval: &RateInterval{Rating: &RIRate{Rates: RateGroups{&Rate{GroupIntervalStart: 0, Value: 100, RateIncrement: 10 * time.Second, RateUnit: time.Second}}}},
					UnitInfo:            &UnitInfo{"1", 2.3, utils.VOICE},
				},
				&Increment{
					Duration:            time.Minute,
					Cost:                10.4,
					BalanceInfo:         &BalanceInfo{"1", "2", "3"},
					BalanceRateInterval: &RateInterval{Rating: &RIRate{Rates: RateGroups{&Rate{GroupIntervalStart: 0, Value: 100, RateIncrement: 10 * time.Second, RateUnit: time.Second}}}},
					UnitInfo:            &UnitInfo{"1", 2.3, utils.VOICE},
				},
				&Increment{
					Duration:            time.Minute,
					Cost:                10.4,
					BalanceInfo:         &BalanceInfo{"1", "2", "3"},
					BalanceRateInterval: &RateInterval{Rating: &RIRate{Rates: RateGroups{&Rate{GroupIntervalStart: 0, Value: 100, RateIncrement: 1111 * time.Second, RateUnit: time.Second}}}},
					UnitInfo:            &UnitInfo{"1", 2.3, utils.VOICE},
				},
				&Increment{
					Duration:            time.Minute,
					Cost:                10.4,
					BalanceInfo:         &BalanceInfo{"1", "2", "3"},
					BalanceRateInterval: &RateInterval{Rating: &RIRate{Rates: RateGroups{&Rate{GroupIntervalStart: 0, Value: 100, RateIncrement: 10 * time.Second, RateUnit: time.Second}}}},
					UnitInfo:            &UnitInfo{"1", 2.3, utils.VOICE},
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

func TestTSMultipleIncrementsCompressDecompress(t *testing.T) {
	tss := TimeSpans{
		&TimeSpan{
			Increments: Increments{
				&Increment{
					Duration:            time.Minute,
					Cost:                10.4,
					BalanceInfo:         &BalanceInfo{"1", "2", "3"},
					BalanceRateInterval: &RateInterval{Rating: &RIRate{Rates: RateGroups{&Rate{GroupIntervalStart: 0, Value: 100, RateIncrement: 10 * time.Second, RateUnit: time.Second}}}},
					UnitInfo:            &UnitInfo{"1", 2.3, utils.VOICE},
				},
				&Increment{
					Duration:            time.Minute,
					Cost:                10.4,
					BalanceInfo:         &BalanceInfo{"1", "2", "3"},
					BalanceRateInterval: &RateInterval{Rating: &RIRate{Rates: RateGroups{&Rate{GroupIntervalStart: 0, Value: 100, RateIncrement: 10 * time.Second, RateUnit: time.Second}}}},
					UnitInfo:            &UnitInfo{"1", 2.3, utils.VOICE},
				},
				&Increment{
					Duration:            time.Minute,
					Cost:                10.4,
					BalanceInfo:         &BalanceInfo{"1", "2", "3"},
					BalanceRateInterval: &RateInterval{Rating: &RIRate{Rates: RateGroups{&Rate{GroupIntervalStart: 0, Value: 100, RateIncrement: 10 * time.Second, RateUnit: time.Second}}}},
					UnitInfo:            &UnitInfo{"1", 2.3, utils.VOICE},
				},
				&Increment{
					Duration:            time.Minute,
					Cost:                10.4,
					BalanceInfo:         &BalanceInfo{"1", "2", "3"},
					BalanceRateInterval: &RateInterval{Rating: &RIRate{Rates: RateGroups{&Rate{GroupIntervalStart: 0, Value: 100, RateIncrement: 1111 * time.Second, RateUnit: time.Second}}}},
					UnitInfo:            &UnitInfo{"1", 2.3, utils.VOICE},
				},
				&Increment{
					Duration:            time.Minute,
					Cost:                10.4,
					BalanceInfo:         &BalanceInfo{"1", "2", "3"},
					BalanceRateInterval: &RateInterval{Rating: &RIRate{Rates: RateGroups{&Rate{GroupIntervalStart: 0, Value: 100, RateIncrement: 10 * time.Second, RateUnit: time.Second}}}},
					UnitInfo:            &UnitInfo{"1", 2.3, utils.VOICE},
				},
			},
		},
	}
	tss.Compress()
	tss.Compress()
	if len(tss[0].Increments) != 3 {
		t.Error("Error compressing timespan: ", tss[0].Increments)
	}
	tss.Decompress()
	tss.Decompress()
	if len(tss[0].Increments) != 5 {
		t.Error("Error decompressing timespans: ", tss[0].Increments)
	}
}

func TestTSBetterIntervalZero(t *testing.T) {
	ts := &TimeSpan{
		TimeStart:    time.Date(2015, 4, 27, 8, 0, 0, 0, time.UTC),
		RateInterval: &RateInterval{Timing: &RITiming{StartTime: "08:00:00"}},
	}

	interval := &RateInterval{Timing: &RITiming{StartTime: "00:00:00"}}
	if !ts.hasBetterRateIntervalThan(interval) {
		t.Error("Wrong better rate interval!")
	}
}

func TestTSBetterIntervalBefore(t *testing.T) {
	ts := &TimeSpan{
		TimeStart:    time.Date(2015, 4, 27, 8, 0, 0, 0, time.UTC),
		RateInterval: &RateInterval{Timing: &RITiming{StartTime: "08:00:00"}},
	}

	interval := &RateInterval{Timing: &RITiming{StartTime: "07:00:00"}}
	if !ts.hasBetterRateIntervalThan(interval) {
		t.Error("Wrong better rate interval!")
	}
}

func TestTSBetterIntervalEqual(t *testing.T) {
	ts := &TimeSpan{
		TimeStart:    time.Date(2015, 4, 27, 8, 0, 0, 0, time.UTC),
		RateInterval: &RateInterval{Timing: &RITiming{StartTime: "08:00:00"}},
	}

	interval := &RateInterval{Timing: &RITiming{StartTime: "08:00:00"}}
	if !ts.hasBetterRateIntervalThan(interval) {
		t.Error("Wrong better rate interval!")
	}
}

func TestTSBetterIntervalAfter(t *testing.T) {
	ts := &TimeSpan{
		TimeStart:    time.Date(2015, 4, 27, 8, 0, 0, 0, time.UTC),
		RateInterval: &RateInterval{Timing: &RITiming{StartTime: "08:00:00"}},
	}

	interval := &RateInterval{Timing: &RITiming{StartTime: "13:00:00"}}
	if !ts.hasBetterRateIntervalThan(interval) {
		t.Error("Wrong better rate interval!")
	}
}

func TestTSBetterIntervalBetter(t *testing.T) {
	ts := &TimeSpan{
		TimeStart:    time.Date(2015, 4, 27, 8, 0, 0, 0, time.UTC),
		RateInterval: &RateInterval{Timing: &RITiming{StartTime: "00:00:00"}},
	}

	interval := &RateInterval{Timing: &RITiming{StartTime: "08:00:00"}}
	if ts.hasBetterRateIntervalThan(interval) {
		t.Error("Wrong better rate interval!")
	}
}

func TestTSBetterIntervalBetterHour(t *testing.T) {
	ts := &TimeSpan{
		TimeStart:    time.Date(2015, 4, 27, 8, 0, 0, 0, time.UTC),
		RateInterval: &RateInterval{Timing: &RITiming{StartTime: "00:00:00"}},
	}

	interval := &RateInterval{Timing: &RITiming{StartTime: "06:00:00"}}
	if ts.hasBetterRateIntervalThan(interval) {
		t.Error("Wrong better rate interval!")
	}
}

func TestTSBetterIntervalAgainAfter(t *testing.T) {
	ts := &TimeSpan{
		TimeStart:    time.Date(2015, 4, 27, 8, 0, 0, 0, time.UTC),
		RateInterval: &RateInterval{Timing: &RITiming{StartTime: "00:00:00"}},
	}

	interval := &RateInterval{Timing: &RITiming{StartTime: "13:00:00"}}
	if !ts.hasBetterRateIntervalThan(interval) {
		t.Error("Wrong better rate interval!")
	}
}

func TestTSCompressDecompress(t *testing.T) {
	tss := TimeSpans{
		&TimeSpan{
			TimeStart:     time.Date(2015, 1, 9, 16, 18, 0, 0, time.UTC),
			TimeEnd:       time.Date(2015, 1, 9, 16, 19, 0, 0, time.UTC),
			DurationIndex: 1 * time.Minute,
		},
		&TimeSpan{
			TimeStart:     time.Date(2015, 1, 9, 16, 19, 0, 0, time.UTC),
			TimeEnd:       time.Date(2015, 1, 9, 16, 20, 0, 0, time.UTC),
			DurationIndex: 2 * time.Minute,
		},
		&TimeSpan{
			TimeStart:     time.Date(2015, 1, 9, 16, 20, 0, 0, time.UTC),
			TimeEnd:       time.Date(2015, 1, 9, 16, 21, 0, 0, time.UTC),
			DurationIndex: 3 * time.Minute,
		},
		&TimeSpan{
			TimeStart:     time.Date(2015, 1, 9, 16, 21, 0, 0, time.UTC),
			TimeEnd:       time.Date(2015, 1, 9, 16, 22, 0, 0, time.UTC),
			DurationIndex: 4 * time.Minute,
		},
	}
	tss.Compress()
	if len(tss) != 1 ||
		!tss[0].TimeStart.Equal(time.Date(2015, 1, 9, 16, 18, 0, 0, time.UTC)) ||
		!tss[0].TimeEnd.Equal(time.Date(2015, 1, 9, 16, 22, 0, 0, time.UTC)) ||
		tss[0].DurationIndex != 4*time.Minute {
		for _, ts := range tss {
			t.Logf("TS: %v", ts)
		}
		t.Error("Error compressing timespans: ", tss)
	}
	tss.Decompress()
	if len(tss) != 4 ||
		!tss[0].TimeStart.Equal(time.Date(2015, 1, 9, 16, 18, 0, 0, time.UTC)) ||
		!tss[0].TimeEnd.Equal(time.Date(2015, 1, 9, 16, 19, 0, 0, time.UTC)) ||
		tss[0].DurationIndex != 1*time.Minute ||
		!tss[1].TimeStart.Equal(time.Date(2015, 1, 9, 16, 19, 0, 0, time.UTC)) ||
		!tss[1].TimeEnd.Equal(time.Date(2015, 1, 9, 16, 20, 0, 0, time.UTC)) ||
		tss[1].DurationIndex != 2*time.Minute ||
		!tss[2].TimeStart.Equal(time.Date(2015, 1, 9, 16, 20, 0, 0, time.UTC)) ||
		!tss[2].TimeEnd.Equal(time.Date(2015, 1, 9, 16, 21, 0, 0, time.UTC)) ||
		tss[2].DurationIndex != 3*time.Minute ||
		!tss[3].TimeStart.Equal(time.Date(2015, 1, 9, 16, 21, 0, 0, time.UTC)) ||
		!tss[3].TimeEnd.Equal(time.Date(2015, 1, 9, 16, 22, 0, 0, time.UTC)) ||
		tss[3].DurationIndex != 4*time.Minute {
		for i, ts := range tss {
			t.Logf("TS(%d): %v", i, ts)
		}
		t.Error("Error decompressing timespans: ", tss)
	}
}
