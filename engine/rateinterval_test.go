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

	"github.com/cgrates/cgrates/utils"
)

func TestRateIntervalSimpleContains(t *testing.T) {
	i := &RateInterval{
		Timing: &RITiming{
			WeekDays:  utils.WeekDays{time.Monday, time.Tuesday, time.Wednesday, time.Thursday, time.Friday},
			StartTime: "18:00:00",
			EndTime:   "",
		},
	}
	d := time.Date(2012, time.February, 27, 23, 59, 59, 0, time.UTC)
	if !i.Contains(d, false) {
		t.Errorf("Date %+v shoud be in interval %+v", d, i)
	}
}

func TestRateIntervalMonth(t *testing.T) {
	i := &RateInterval{Timing: &RITiming{Months: utils.Months{time.February}}}
	d := time.Date(2012, time.February, 10, 23, 0, 0, 0, time.UTC)
	d1 := time.Date(2012, time.January, 10, 23, 0, 0, 0, time.UTC)
	if !i.Contains(d, false) {
		t.Errorf("Date %v shoud be in interval %v", d, i)
	}
	if i.Contains(d1, false) {
		t.Errorf("Date %v shoud not be in interval %v", d1, i)
	}
}

func TestRateIntervalMonthDay(t *testing.T) {
	i := &RateInterval{Timing: &RITiming{MonthDays: utils.MonthDays{10}}}
	d := time.Date(2012, time.February, 10, 23, 0, 0, 0, time.UTC)
	d1 := time.Date(2012, time.February, 11, 23, 0, 0, 0, time.UTC)
	if !i.Contains(d, false) {
		t.Errorf("Date %v shoud be in interval %v", d, i)
	}
	if i.Contains(d1, false) {
		t.Errorf("Date %v shoud not be in interval %v", d1, i)
	}
}

func TestRateIntervalMonthAndMonthDay(t *testing.T) {
	i := &RateInterval{Timing: &RITiming{Months: utils.Months{time.February}, MonthDays: utils.MonthDays{10}}}
	d := time.Date(2012, time.February, 10, 23, 0, 0, 0, time.UTC)
	d1 := time.Date(2012, time.February, 11, 23, 0, 0, 0, time.UTC)
	d2 := time.Date(2012, time.January, 10, 23, 0, 0, 0, time.UTC)
	if !i.Contains(d, false) {
		t.Errorf("Date %v shoud be in interval %v", d, i)
	}
	if i.Contains(d1, false) {
		t.Errorf("Date %v shoud not be in interval %v", d1, i)
	}
	if i.Contains(d2, false) {
		t.Errorf("Date %v shoud not be in interval %v", d2, i)
	}
}

func TestRateIntervalWeekDays(t *testing.T) {
	i := &RateInterval{Timing: &RITiming{WeekDays: []time.Weekday{time.Wednesday}}}
	i2 := &RateInterval{Timing: &RITiming{WeekDays: []time.Weekday{time.Wednesday, time.Thursday}}}
	d := time.Date(2012, time.February, 1, 23, 0, 0, 0, time.UTC)
	d1 := time.Date(2012, time.February, 2, 23, 0, 0, 0, time.UTC)
	if !i.Contains(d, false) {
		t.Errorf("Date %v shoud be in interval %v", d, i)
	}
	if i.Contains(d1, false) {
		t.Errorf("Date %v shoud not be in interval %v", d1, i)
	}
	if !i2.Contains(d, false) {
		t.Errorf("Date %v shoud be in interval %v", d, i2)
	}
	if !i2.Contains(d1, false) {
		t.Errorf("Date %v shoud be in interval %v", d1, i2)
	}
}

func TestRateIntervalMonthAndMonthDayAndWeekDays(t *testing.T) {
	i := &RateInterval{Timing: &RITiming{Months: utils.Months{time.February}, MonthDays: utils.MonthDays{1}, WeekDays: []time.Weekday{time.Wednesday}}}
	i2 := &RateInterval{Timing: &RITiming{Months: utils.Months{time.February}, MonthDays: utils.MonthDays{2}, WeekDays: []time.Weekday{time.Wednesday, time.Thursday}}}
	d := time.Date(2012, time.February, 1, 23, 0, 0, 0, time.UTC)
	d1 := time.Date(2012, time.February, 2, 23, 0, 0, 0, time.UTC)
	if !i.Contains(d, false) {
		t.Errorf("Date %v shoud be in interval %v", d, i)
	}
	if i.Contains(d1, false) {
		t.Errorf("Date %v shoud not be in interval %v", d1, i)
	}
	if i2.Contains(d, false) {
		t.Errorf("Date %v shoud be in interval %v", d, i2)
	}
	if !i2.Contains(d1, false) {
		t.Errorf("Date %v shoud be in interval %v", d1, i2)
	}
}

func TestRateIntervalHours(t *testing.T) {
	i := &RateInterval{Timing: &RITiming{StartTime: "14:30:00", EndTime: "15:00:00"}}
	d := time.Date(2012, time.February, 10, 14, 30, 1, 0, time.UTC)
	d1 := time.Date(2012, time.January, 10, 14, 29, 0, 0, time.UTC)
	d2 := time.Date(2012, time.January, 10, 14, 59, 0, 0, time.UTC)
	d3 := time.Date(2012, time.January, 10, 15, 01, 0, 0, time.UTC)
	if !i.Contains(d, false) {
		t.Errorf("Date %v shoud be in interval %v", d, i)
	}
	if i.Contains(d1, false) {
		t.Errorf("Date %v shoud not be in interval %v", d1, i)
	}
	if !i.Contains(d2, false) {
		t.Errorf("Date %v shoud be in interval %v", d2, i)
	}
	if i.Contains(d3, false) {
		t.Errorf("Date %v shoud not be in interval %v", d3, i)
	}
}

func TestRateIntervalEverything(t *testing.T) {
	i := &RateInterval{
		Timing: &RITiming{
			Months:    utils.Months{time.February},
			Years:     utils.Years{2012},
			MonthDays: utils.MonthDays{1},
			WeekDays:  []time.Weekday{time.Wednesday, time.Thursday},
			StartTime: "14:30:00",
			EndTime:   "15:00:00"}}
	d := time.Date(2012, time.February, 1, 14, 30, 1, 0, time.UTC)
	d1 := time.Date(2012, time.February, 1, 14, 29, 1, 0, time.UTC)
	d2 := time.Date(2012, time.February, 1, 15, 00, 00, 0, time.UTC)
	d3 := time.Date(2012, time.February, 1, 15, 0, 1, 0, time.UTC)
	d4 := time.Date(2011, time.February, 1, 15, 00, 00, 0, time.UTC)
	if !i.Contains(d, false) {
		t.Errorf("Date %v shoud be in interval %v", d, i)
	}
	if i.Contains(d1, false) {
		t.Errorf("Date %v shoud not be in interval %v", d1, i)
	}
	if !i.Contains(d2, false) {
		t.Errorf("Date %v shoud be in interval %v", d2, i)
	}
	if i.Contains(d3, false) {
		t.Errorf("Date %v shoud not be in interval %v", d3, i)
	}
	if i.Contains(d4, false) {
		t.Errorf("Date %v shoud not be in interval %v", d3, i)
	}
}

func TestRateIntervalEqual(t *testing.T) {
	i1 := &RateInterval{
		Timing: &RITiming{
			Months:    utils.Months{time.February},
			MonthDays: utils.MonthDays{1},
			WeekDays:  []time.Weekday{time.Wednesday, time.Thursday},
			StartTime: "14:30:00",
			EndTime:   "15:00:00"}}
	i2 := &RateInterval{Timing: &RITiming{
		Months:    utils.Months{time.February},
		MonthDays: utils.MonthDays{1},
		WeekDays:  []time.Weekday{time.Wednesday, time.Thursday},
		StartTime: "14:30:00",
		EndTime:   "15:00:00"}}
	if !i1.Equal(i2) || !i2.Equal(i1) {
		t.Errorf("%v and %v are not equal", i1, i2)
	}
}

func TestRateIntervalNotEqual(t *testing.T) {
	i1 := &RateInterval{
		Timing: &RITiming{
			Months:    utils.Months{time.February},
			MonthDays: utils.MonthDays{1},
			WeekDays:  []time.Weekday{time.Wednesday},
			StartTime: "14:30:00",
			EndTime:   "15:00:00"}}
	i2 := &RateInterval{Timing: &RITiming{
		Months:    utils.Months{time.February},
		MonthDays: utils.MonthDays{1},
		WeekDays:  []time.Weekday{time.Wednesday, time.Thursday},
		StartTime: "14:30:00",
		EndTime:   "15:00:00"}}
	if i1.Equal(i2) || i2.Equal(i1) {
		t.Errorf("%v and %v not equal", i1, i2)
	}
}

func TestRitStrigyfy(t *testing.T) {
	rit1 := &RITiming{
		Years:     utils.Years{},
		Months:    utils.Months{time.January, time.February},
		MonthDays: utils.MonthDays{},
		StartTime: "00:00:00",
	}
	rit2 := &RITiming{
		Years:     utils.Years{},
		Months:    utils.Months{time.January, time.February},
		MonthDays: utils.MonthDays{},
		StartTime: "00:00:00",
	}
	if rit1.Stringify() != rit2.Stringify() {
		t.Error("Error in rir stringify: ", rit1.Stringify(), rit2.Stringify())
	}
}

func TestRirStrigyfy(t *testing.T) {
	rir1 := &RIRate{
		ConnectFee: 0.1,
		Rates: RateGroups{
			&Rate{
				GroupIntervalStart: time.Hour,
				Value:              0.17,
				RateIncrement:      time.Second,
				RateUnit:           time.Minute,
			},
			&Rate{
				GroupIntervalStart: 0,
				Value:              0.7,
				RateIncrement:      time.Second,
				RateUnit:           time.Minute,
			},
		},
		RoundingMethod:   utils.ROUNDING_MIDDLE,
		RoundingDecimals: 4,
	}
	rir2 := &RIRate{
		ConnectFee: 0.1,
		Rates: RateGroups{
			&Rate{
				GroupIntervalStart: time.Hour,
				Value:              0.17,
				RateIncrement:      time.Second,
				RateUnit:           time.Minute,
			},
			&Rate{
				GroupIntervalStart: 0,
				Value:              0.7,
				RateIncrement:      time.Second,
				RateUnit:           time.Minute,
			},
		},
		RoundingMethod:   utils.ROUNDING_MIDDLE,
		RoundingDecimals: 4,
	}
	if rir1.Stringify() != rir2.Stringify() {
		t.Error("Error in rate stringify: ", rir1.Stringify(), rir2.Stringify())
	}
}

func TestRateStrigyfy(t *testing.T) {
	r1 := &Rate{
		GroupIntervalStart: time.Hour,
		Value:              0.17,
		RateUnit:           time.Minute,
	}
	r2 := &Rate{
		GroupIntervalStart: time.Hour,
		Value:              0.17,
		RateUnit:           time.Minute,
	}
	if r1.Stringify() != r2.Stringify() {
		t.Error("Error in rate stringify: ", r1.Stringify(), r2.Stringify())
	}
}

func TestRateIntervalCronAll(t *testing.T) {
	rit := &RITiming{
		Years:     utils.Years{2012},
		Months:    utils.Months{time.February},
		MonthDays: utils.MonthDays{1},
		WeekDays:  []time.Weekday{time.Sunday},
		StartTime: "14:30:00",
	}
	expected := "0 30 14 1 2 0 2012"
	cron := rit.CronString()
	if cron != expected {
		t.Errorf("Expected %s was %s", expected, cron)
	}
}

func TestRateIntervalCronMultiple(t *testing.T) {
	rit := &RITiming{
		Years:     utils.Years{2012, 2014},
		Months:    utils.Months{time.February, time.January},
		MonthDays: utils.MonthDays{15, 16},
		WeekDays:  []time.Weekday{time.Sunday, time.Monday},
		StartTime: "14:30:00",
	}
	expected := "0 30 14 15,16 2,1 0,1 2012,2014"
	cron := rit.CronString()

	if cron != expected {
		t.Errorf("Expected %s was %s", expected, cron)
	}
}

func TestRateIntervalCronStar(t *testing.T) {
	rit := &RITiming{
		StartTime: "*:30:00",
	}
	expected := "0 30 * * * * *"
	cron := rit.CronString()

	if cron != expected {
		t.Errorf("Expected %s was %s", expected, cron)
	}
}

func TestRateIntervalCronEmpty(t *testing.T) {
	rit := &RITiming{}
	expected := "* * * * * * *"
	cron := rit.CronString()

	if cron != expected {
		t.Errorf("Expected %s was %s", expected, cron)
	}
}

/*********************************Benchmarks**************************************/

func BenchmarkRateIntervalContainsDate(b *testing.B) {
	i := &RateInterval{Timing: &RITiming{Months: utils.Months{time.February}, MonthDays: utils.MonthDays{1}, WeekDays: []time.Weekday{time.Wednesday, time.Thursday}, StartTime: "14:30:00", EndTime: "15:00:00"}}
	d := time.Date(2012, time.February, 1, 14, 30, 0, 0, time.UTC)
	for x := 0; x < b.N; x++ {
		i.Contains(d, false)
	}
}
