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
	"fmt"
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

func TestRateIntervalCost(t *testing.T) {
	ri := &RateInterval{
		Rating: &RIRate{
			Rates: RateGroups{
				&Rate{
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
		&Rate{
			GroupIntervalStart: time.Duration(0),
			Value:              0.1,
			RateIncrement:      time.Minute,
			RateUnit:           60 * time.Second,
		},
		&Rate{
			GroupIntervalStart: time.Duration(60 * time.Second),
			Value:              0.05,
			RateIncrement:      time.Second,
			RateUnit:           time.Second,
		},
	}
	rg2 := RateGroups{
		&Rate{
			GroupIntervalStart: time.Duration(0),
			Value:              0.1,
			RateIncrement:      time.Minute,
			RateUnit:           60 * time.Second,
		},
		&Rate{
			GroupIntervalStart: time.Duration(60 * time.Second),
			Value:              0.05,
			RateIncrement:      time.Second,
			RateUnit:           time.Second,
		},
	}
	if !rg1.Equals(rg2) {
		t.Error("not equal")
	}
	rg2 = RateGroups{
		&Rate{
			GroupIntervalStart: time.Duration(0),
			Value:              0.1,
			RateIncrement:      time.Minute,
			RateUnit:           60 * time.Second,
		},
		&Rate{
			GroupIntervalStart: time.Duration(60 * time.Second),
			Value:              0.3,
			RateIncrement:      time.Second,
			RateUnit:           time.Second,
		},
	}
	if rg1.Equals(rg2) {
		t.Error("equals")
	}
	rg2 = RateGroups{
		&Rate{
			GroupIntervalStart: time.Duration(0),
			Value:              0.1,
			RateIncrement:      time.Minute,
			RateUnit:           60 * time.Second,
		},
	}
	if rg1.Equals(rg2) {
		t.Error("equals")
	}
}

func TestRateInterval_CronStringTT(t *testing.T) {
	testCases := []struct {
		name     string
		ri       *RITiming
		wantCron string
	}{
		{
			name: "SingleYearMonthMonthDayWeekDay",
			ri: &RITiming{
				Years:     utils.Years{2012},
				Months:    utils.Months{time.February},
				MonthDays: utils.MonthDays{1},
				WeekDays:  []time.Weekday{time.Sunday},
				StartTime: "14:30:00",
			},
			wantCron: "0 30 14 1 2 0 2012",
		},
		{
			name: "MultipleYearsMonthsMonthsDaysWeekDays",
			ri: &RITiming{
				Years:     utils.Years{2012, 2014},
				Months:    utils.Months{time.February, time.January},
				MonthDays: utils.MonthDays{15, 16},
				WeekDays:  []time.Weekday{time.Sunday, time.Monday},
				StartTime: "14:30:00",
			},
			wantCron: "0 30 14 15,16 2,1 0,1 2012,2014",
		},
		{
			name: "WildcardStartTime",
			ri: &RITiming{
				StartTime: "*:30:00",
			},
			wantCron: "0 30 * * * * *",
		},
		{
			name:     "AllFieldsEmpty",
			ri:       &RITiming{},
			wantCron: "* * * * * * *",
		},
		{
			name: "EveryMinute",
			ri: &RITiming{
				StartTime: utils.MetaEveryMinute,
			},
			wantCron: "0 * * * * * *",
		},
		{
			name: "Hourly",
			ri: &RITiming{
				StartTime: utils.MetaHourly,
			},
			wantCron: "0 0 * * * * *",
		},
		{
			name: "SingleDigit0Hour",
			ri: &RITiming{
				StartTime: "0:49:25",
			},
			wantCron: "25 49 0 * * * *",
		},
		{
			name: "SingleDigit0Minute",
			ri: &RITiming{
				StartTime: "*:0:49",
			},
			wantCron: "49 0 * * * * *",
		},
		{
			name: "SingleDigit0Second",
			ri: &RITiming{
				StartTime: "*:49:0",
			},
			wantCron: "0 49 * * * * *",
		},
		{
			name: "DoubleDigit0Hour",
			ri: &RITiming{
				StartTime: "00:49:25",
			},
			wantCron: "25 49 0 * * * *",
		},
		{
			name: "DoubleDigit0Minute",
			ri: &RITiming{
				StartTime: "*:00:49",
			},
			wantCron: "49 0 * * * * *",
		},
		{
			name: "DoubleDigit0Second",
			ri: &RITiming{
				StartTime: "*:49:00",
			},
			wantCron: "0 49 * * * * *",
		},
		{
			name: "InvalidStartTimeFormat",
			ri: &RITiming{
				StartTime: "223000",
			},
			wantCron: "* * * * * * *",
		},
		{
			name: "PatternWithZeros",
			ri: &RITiming{
				Years:     utils.Years{2020},
				Months:    utils.Months{1},
				MonthDays: utils.MonthDays{0},
				WeekDays:  []time.Weekday{time.Monday},
				StartTime: "00:00:00",
			},
			wantCron: "0 0 0 0 1 1 2020",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := tc.ri.CronString()
			if got != tc.wantCron {
				t.Errorf("RITiming=%s\nCronString()=%q, want %q", utils.ToJSON(tc.ri), got, tc.wantCron)
			}
		})
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

func TestRIRateClone(t *testing.T) {
	var rit, cln RIRate
	if cloned := rit.Clone(); !reflect.DeepEqual(cln, *cloned) {
		t.Errorf("\nExpecting: %+v,\n received: %+v", cln, *cloned)
	}
	rit = RIRate{
		ConnectFee:       0.7,
		RoundingMethod:   "RoundingMethod_test",
		RoundingDecimals: 7,
		MaxCost:          0.7,
		MaxCostStrategy:  "MaxCostStrategy_test",
		Rates: RateGroups{
			&Rate{
				GroupIntervalStart: time.Duration(10),
				Value:              0.7,
				RateIncrement:      time.Duration(10),
				RateUnit:           time.Duration(10),
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
			&Rate{
				GroupIntervalStart: time.Duration(10),
				Value:              0.7,
				RateIncrement:      time.Duration(10),
				RateUnit:           time.Duration(10),
			},
		},
	}
	cloned := rit.Clone()
	if !reflect.DeepEqual(cln, *cloned) {
		t.Errorf("Expecting: %+v, received: %+v", cln, *cloned)
	}
	rit.Rates[0].GroupIntervalStart = time.Duration(7)
	if cloned.Rates[0].GroupIntervalStart != time.Duration(10) {
		t.Errorf("\nExpecting: 10,\n received: %+v", cloned.Rates[0].GroupIntervalStart)
	}
}

func TestRateIntervalIsActive(t *testing.T) {
	rit := &RITiming{
		Years:      utils.Years{2021},
		Months:     utils.Months{9},
		MonthDays:  utils.MonthDays{2},
		WeekDays:   utils.WeekDays{2},
		StartTime:  "00:00:00",
		EndTime:    "02:02:02",
		cronString: str,
		tag:        str,
	}

	rcv := rit.IsActive()

	if rcv != false {
		t.Error(rcv)
	}
}

func TestRateIntervalFieldAsInterface(t *testing.T) {
	r := &Rate{
		GroupIntervalStart: 1 * time.Millisecond,
		Value:              fl,
		RateIncrement:      1 * time.Millisecond,
		RateUnit:           1 * time.Millisecond,
	}

	tests := []struct {
		name string
		arg  []string
		exp  any
		err  string
	}{
		{
			name: "empty field path",
			arg:  []string{},
			exp:  nil,
			err:  "NOT_FOUND",
		},
		{
			name: "default case",
			arg:  []string{str},
			exp:  nil,
			err:  "unsupported field prefix: <test>",
		},
		{
			name: "GroupIntervalStart case",
			arg:  []string{"GroupIntervalStart"},
			exp:  1 * time.Millisecond,
			err:  "",
		},
		{
			name: "RateIncrement case",
			arg:  []string{"RateIncrement"},
			exp:  1 * time.Millisecond,
			err:  "",
		},
		{
			name: "RateUnit case",
			arg:  []string{"RateUnit"},
			exp:  1 * time.Millisecond,
			err:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rcv, err := r.FieldAsInterface(tt.arg)

			if err != nil {
				if err.Error() != tt.err {
					t.Fatal(err)
				}
			}

			if rcv != tt.exp {
				t.Error(rcv)
			}
		})
	}
}

func TestRateIntervalRateGroupsEqual(t *testing.T) {
	r := &Rate{
		GroupIntervalStart: 1 * time.Millisecond,
		Value:              fl,
		RateIncrement:      1 * time.Millisecond,
		RateUnit:           1 * time.Millisecond,
	}
	r2 := &Rate{
		GroupIntervalStart: 1 * time.Millisecond,
		Value:              3.5,
		RateIncrement:      1 * time.Millisecond,
		RateUnit:           1 * time.Millisecond,
	}
	pg := RateGroups{r}
	of := RateGroups{}
	of2 := RateGroups{r2}
	of3 := RateGroups{r}

	tests := []struct {
		name string
		arg  RateGroups
		exp  bool
	}{
		{
			name: "different lengths",
			arg:  of,
			exp:  false,
		},
		{
			name: "not equal",
			arg:  of2,
			exp:  false,
		},
		{
			name: "equal",
			arg:  of3,
			exp:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rcv := pg.Equal(tt.arg)

			if rcv != tt.exp {
				t.Error(rcv)
			}
		})
	}
}

func TestRateIntervalAddRate(t *testing.T) {
	r := &Rate{
		GroupIntervalStart: 1 * time.Millisecond,
		Value:              fl,
		RateIncrement:      1 * time.Millisecond,
		RateUnit:           1 * time.Millisecond,
	}
	r2 := &Rate{
		GroupIntervalStart: 1 * time.Millisecond,
		Value:              3.5,
		RateIncrement:      1 * time.Millisecond,
		RateUnit:           1 * time.Millisecond,
	}
	pg := RateGroups{r}

	pg.AddRate(r2)
	exp := RateGroups{r, r2}

	if !reflect.DeepEqual(pg, exp) {
		t.Errorf("expected %s, received %s", utils.ToJSON(exp), utils.ToJSON(pg))
	}

	pg.AddRate(r2)

	if !reflect.DeepEqual(pg, exp) {
		t.Errorf("expected %s, received %s", utils.ToJSON(exp), utils.ToJSON(pg))
	}
}

func TestRateIntervalString_DISABLED(t *testing.T) {
	rit := &RITiming{
		Years:      utils.Years{2021},
		Months:     utils.Months{9},
		MonthDays:  utils.MonthDays{2},
		WeekDays:   utils.WeekDays{2},
		StartTime:  "00:00:00",
		EndTime:    "02:02:02",
		cronString: str,
		tag:        str,
	}
	i := &RateInterval{
		Timing: rit,
	}

	rcv := i.String_DISABLED()
	exp := fmt.Sprintf("%v %v %v %v %v %v", i.Timing.Years, i.Timing.Months, i.Timing.MonthDays, i.Timing.WeekDays, i.Timing.StartTime, i.Timing.EndTime)

	if rcv != exp {
		t.Errorf("expected %s, received %s", exp, rcv)
	}
}

func TestRateIntervalEqual2(t *testing.T) {
	rit := &RITiming{
		Years:      utils.Years{2021},
		Months:     utils.Months{9},
		MonthDays:  utils.MonthDays{2},
		WeekDays:   utils.WeekDays{2},
		StartTime:  "00:00:00",
		EndTime:    "02:02:02",
		cronString: str,
		tag:        str,
	}
	i := &RateInterval{
		Timing: rit,
	}

	rcv := i.Equal(nil)

	if rcv != false {
		t.Error(rcv)
	}
}

func TestRateIntervalGetMaxCost(t *testing.T) {
	var ri RateInterval

	fl, str := ri.GetMaxCost()

	if fl != 0.0 {
		t.Error(fl)
	}

	if str != "" {
		t.Error(str)
	}
}

func TestRateIntervalRateClone(t *testing.T) {
	var r *Rate

	rcv := r.Clone()

	if rcv != nil {
		t.Error(rcv)
	}
}

func TestRateIntervalGetRateParameters(t *testing.T) {
	r := &Rate{
		GroupIntervalStart: 1 * time.Millisecond,
		Value:              fl,
		RateIncrement:      1 * time.Millisecond,
		RateUnit:           1 * time.Millisecond,
	}
	i := &RateInterval{
		Rating: &RIRate{
			Rates: RateGroups{r},
		},
	}

	fl, ri, ru := i.GetRateParameters(0 * time.Millisecond)

	if fl != -1 || ri != -1 || ru != -1 {
		t.Error(fl, ri, ru)
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
