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

var (
	ACTIVE_TIME = time.Now()
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

func TestRateIntervalEqualWeight(t *testing.T) {
	i1 := &RateInterval{Weight: 1}
	i2 := &RateInterval{Weight: 2}
	if i1.Equal(i2) {
		t.Errorf("%v and %v should not be equal", i1, i2)
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
			&RGRate{
				GroupIntervalStart: time.Hour,
				Value:              0.17,
				RateIncrement:      time.Second,
				RateUnit:           time.Minute,
			},
			&RGRate{
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
			&RGRate{
				GroupIntervalStart: time.Hour,
				Value:              0.17,
				RateIncrement:      time.Second,
				RateUnit:           time.Minute,
			},
			&RGRate{
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
	r1 := &RGRate{
		GroupIntervalStart: time.Hour,
		Value:              0.17,
		RateUnit:           time.Minute,
	}
	r2 := &RGRate{
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

func TestRITimingCronEveryX(t *testing.T) {
	rit := &RITiming{
		StartTime: utils.MetaEveryMinute,
	}
	eCronStr := "0 * * * * * *"
	if cronStr := rit.CronString(); cronStr != eCronStr {
		t.Errorf("Expecting: <%s>, received: <%s>", eCronStr, cronStr)
	}
	rit = &RITiming{
		StartTime: utils.MetaHourly,
	}
	eCronStr = "0 0 * * * * *"
	if cronStr := rit.CronString(); cronStr != eCronStr {
		t.Errorf("Expecting: <%s>, received: <%s>", eCronStr, cronStr)
	}
}

func TestRateIntervalCost(t *testing.T) {
	ri := &RateInterval{
		Rating: &RIRate{
			Rates: RateGroups{
				&RGRate{
					Value:         0.1,
					RateIncrement: time.Second,
					RateUnit:      60 * time.Second,
				},
			},
		},
	}
	x := ri.GetCost(60*time.Second, 0)
	if x != 0.1 {
		t.Error("expected 0.1 was: ", x)
	}
}

func TestRateGroupsEquals(t *testing.T) {
	rg1 := RateGroups{
		&RGRate{
			GroupIntervalStart: 0,
			Value:              0.1,
			RateIncrement:      time.Minute,
			RateUnit:           60 * time.Second,
		},
		&RGRate{
			GroupIntervalStart: 60 * time.Second,
			Value:              0.05,
			RateIncrement:      time.Second,
			RateUnit:           time.Second,
		},
	}
	rg2 := RateGroups{
		&RGRate{
			GroupIntervalStart: 0,
			Value:              0.1,
			RateIncrement:      time.Minute,
			RateUnit:           60 * time.Second,
		},
		&RGRate{
			GroupIntervalStart: 60 * time.Second,
			Value:              0.05,
			RateIncrement:      time.Second,
			RateUnit:           time.Second,
		},
	}
	if !rg1.Equals(rg2) {
		t.Error("not equal")
	}
	rg2 = RateGroups{
		&RGRate{
			GroupIntervalStart: 0,
			Value:              0.1,
			RateIncrement:      time.Minute,
			RateUnit:           60 * time.Second,
		},
		&RGRate{
			GroupIntervalStart: 60 * time.Second,
			Value:              0.3,
			RateIncrement:      time.Second,
			RateUnit:           time.Second,
		},
	}
	if rg1.Equals(rg2) {
		t.Error("equals")
	}
	rg2 = RateGroups{
		&RGRate{
			GroupIntervalStart: 0,
			Value:              0.1,
			RateIncrement:      time.Minute,
			RateUnit:           60 * time.Second,
		},
	}
	if rg1.Equals(rg2) {
		t.Error("equals")
	}
}

func TestRateIntervalClone(t *testing.T) {
	var i, cln RateInterval
	if cloned := i.Clone(); !reflect.DeepEqual(cln, *cloned) {
		t.Errorf("Expecting: %+v, received: %+v", cln, *cloned)
	}
	i = RateInterval{
		Weight: 0.7,
	}
	cln = RateInterval{
		Weight: 0.7,
	}
	if cloned := i.Clone(); !reflect.DeepEqual(cln, *cloned) {
		t.Errorf("Expecting: %+v, received: %+v", cln, *cloned)
	}
}

func TestRITimingClone(t *testing.T) {
	var rit, cln RITiming
	if cloned := rit.Clone(); !reflect.DeepEqual(cln, *cloned) {
		t.Errorf("Expecting: %+v, received: %+v", cln, *cloned)
	}
	rit = RITiming{
		Years:     utils.Years{2019},
		Months:    utils.Months{4},
		MonthDays: utils.MonthDays{18},
		WeekDays:  utils.WeekDays{5},
		StartTime: "StartTime_test",
		EndTime:   "EndTime_test",
	}
	cln = RITiming{
		Years:     utils.Years{2019},
		Months:    utils.Months{4},
		MonthDays: utils.MonthDays{18},
		WeekDays:  utils.WeekDays{5},
		StartTime: "StartTime_test",
		EndTime:   "EndTime_test",
	}
	cloned := rit.Clone()
	if !reflect.DeepEqual(cln, *cloned) {
		t.Errorf("Expecting: %+v, received: %+v", cln, *cloned)
	}
	rit.Years = utils.Years{2020}
	if cloned.Years[0] != cln.Years[0] {
		t.Errorf("Expecting: 2019, received: %+v", cloned.Years)
	}
}

func TestRITimingClone2(t *testing.T) {
	var rit, cln RITiming
	rit = RITiming{
		Years:     utils.Years{2019},
		Months:    utils.Months{4},
		MonthDays: utils.MonthDays{18},
		WeekDays:  utils.WeekDays{5},
		StartTime: "StartTime_test",
		EndTime:   "EndTime_test",
	}
	cln = RITiming{
		Years:     utils.Years{2019},
		Months:    utils.Months{4},
		MonthDays: utils.MonthDays{18},
		WeekDays:  utils.WeekDays{5},
		StartTime: "StartTime_test",
		EndTime:   "EndTime_test",
	}
	cloned := rit.Clone()
	if !reflect.DeepEqual(cln, *cloned) {
		t.Errorf("Expecting: %+v, received: %+v", cln, *cloned)
	}
	rit.Years[0] = 2020
	if cloned.Years[0] != cln.Years[0] {
		t.Errorf("Expecting: 2019, received: %+v", cloned.Years)
	}
}

func TestRIRateClone(t *testing.T) {
	var rit, cln RIRate
	if cloned := rit.Clone(); !reflect.DeepEqual(cln, *cloned) {
		t.Errorf("Expecting: %+v,\n received: %+v", cln, *cloned)
	}
	rit = RIRate{
		ConnectFee:       0.7,
		RoundingMethod:   "RoundingMethod_test",
		RoundingDecimals: 7,
		MaxCost:          0.7,
		MaxCostStrategy:  "MaxCostStrategy_test",
		Rates: RateGroups{
			&RGRate{
				GroupIntervalStart: 10,
				Value:              0.7,
				RateIncrement:      10,
				RateUnit:           10,
			},
		},
	}
	cln = RIRate{
		ConnectFee:       0.7,
		RoundingMethod:   "RoundingMethod_test",
		RoundingDecimals: 7,
		MaxCost:          0.7,
		MaxCostStrategy:  "MaxCostStrategy_test",
		Rates: RateGroups{
			&RGRate{
				GroupIntervalStart: 10,
				Value:              0.7,
				RateIncrement:      10,
				RateUnit:           10,
			},
		},
	}
	cloned := rit.Clone()
	if !reflect.DeepEqual(cln, *cloned) {
		t.Errorf("Expecting: %+v, received: %+v", cln, *cloned)
	}
	rit.Rates[0].GroupIntervalStart = 7
	if cloned.Rates[0].GroupIntervalStart != 10 {
		t.Errorf("Expecting: 10,\n received: %+v", cloned.Rates[0].GroupIntervalStart)
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

func TestRateIntervalCronStringDefault(t *testing.T) {
	rit := &RITiming{
		StartTime: "223000",
	}
	cron := rit.CronString()
	if !reflect.DeepEqual(cron, "* * * * * * *") {
		t.Errorf("\nExpecting: <* * * * * * *>,\n Received: <%+v>", cron)
	}
}

func TestRateIntervalCronStringMonthDayNegative(t *testing.T) {
	rit := &RITiming{
		StartTime: "223000",
		MonthDays: utils.MonthDays{-1},
	}
	cron := rit.CronString()
	if !reflect.DeepEqual(cron, "* * * L * * *") {
		t.Errorf("\nExpecting: <* * * L * * *>,\n Received: <%+v>", cron)
	}
}

func TestRateIntervalIsActiveAt(t *testing.T) {
	rit := &RITiming{}
	cronActive := rit.IsActive()
	if !reflect.DeepEqual(cronActive, true) {
		t.Errorf("\nExpecting: <true>,\n Received: <%+v>", cronActive)
	}
}

func TestRateIntervalIsActiveAtNot(t *testing.T) {
	rit := &RITiming{
		Years: utils.Years{1000},
	}
	cronActive := rit.IsActive()
	if !reflect.DeepEqual(cronActive, false) {
		t.Errorf("\nExpecting: <false>,\n Received: <%+v>", cronActive)
	}
}

func TestRateIntervalFieldAsInterfaceError(t *testing.T) {
	rateTest := &RGRate{
		Value: 2.2,
	}
	_, err := rateTest.FieldAsInterface([]string{"FALSE"})
	if err == nil && err.Error() != "unsupported field prefix: <FALSE>" {
		t.Errorf("\nExpecting: <unsupported field prefix: <FALSE>>,\n Received: <%+v>", err)
	}
}

func TestRateIntervalFieldAsInterfaceError2(t *testing.T) {
	rateTest := &RGRate{}
	_, err := rateTest.FieldAsInterface([]string{"value1", "value2"})

	if err == nil && err != utils.ErrNotFound {
		t.Errorf("\nExpecting: <NOT_FOUND>,\n Received: <%+v>", err)
	}
}

func TestRateIntervalFieldAsInterfaceRateIncrement(t *testing.T) {
	rateTest := &RGRate{
		RateIncrement: time.Second,
	}
	if result, err := rateTest.FieldAsInterface([]string{"RateIncrement"}); err != nil {
		t.Errorf("\nExpecting: <nil>,\n Received: <%+v>", err)
	} else if !reflect.DeepEqual(result, time.Second) {
		t.Errorf("\nExpecting: <1s>,\n Received: <%+v>", result)
	}

}

func TestRateIntervalFieldAsInterfaceGroupIntervalStart(t *testing.T) {
	rateTest := &RGRate{
		GroupIntervalStart: time.Second,
	}
	if result, err := rateTest.FieldAsInterface([]string{"GroupIntervalStart"}); err != nil {
		t.Errorf("\nExpecting: <nil>,\n Received: <%+v>", err)
	} else if !reflect.DeepEqual(result, time.Second) {
		t.Errorf("\nExpecting: <1s>,\n Received: <%+v>", result)
	}

}

func TestRateIntervalFieldAsInterfaceRateUnit(t *testing.T) {
	rateTest := &RGRate{
		RateUnit: time.Second,
	}
	if result, err := rateTest.FieldAsInterface([]string{"RateUnit"}); err != nil {
		t.Errorf("\nExpecting: <nil>,\n Received: <%+v>", err)
	} else if !reflect.DeepEqual(result, time.Second) {
		t.Errorf("\nExpecting: <1s>,\n Received: <%+v>", result)
	}

}

func TestRateGroupsEqual(t *testing.T) {
	rateGroupOG := RateGroups{&RGRate{
		Value: 2.2,
	}}
	rateGroupOther := RateGroups{&RGRate{
		Value: 2.2,
	}}
	result := rateGroupOG.Equal(rateGroupOther)
	if !reflect.DeepEqual(result, true) {
		t.Errorf("\nExpecting: <true>,\n Received: <%+v>", result)
	}
}

func TestRateGroupsEqualFalse(t *testing.T) {
	rateGroupOG := RateGroups{&RGRate{
		Value: 2.2,
	}}
	rateGroupOther := RateGroups{&RGRate{
		Value: 2.5,
	}}
	result := rateGroupOG.Equal(rateGroupOther)
	if !reflect.DeepEqual(result, false) {
		t.Errorf("\nExpecting: <false>,\n Received: <%+v>", result)
	}
}

func TestRateGroupsUnEqual(t *testing.T) {
	rateGroupOG := RateGroups{&RGRate{
		Value: 2.2,
	},
		&RGRate{},
	}
	rateGroupOther := RateGroups{&RGRate{
		Value: 2.5,
	}}
	result := rateGroupOG.Equal(rateGroupOther)
	if !reflect.DeepEqual(result, false) {
		t.Errorf("\nExpecting: <false>,\n Received: <%+v>", result)
	}
}
