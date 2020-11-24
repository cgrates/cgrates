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

	"github.com/ericlagergren/decimal"

	"github.com/cgrates/cron"

	"github.com/cgrates/cgrates/utils"
)

func TestRateProfileSort(t *testing.T) {
	rPrf := &RateProfile{
		Tenant: "cgrates.org",
		ID:     "RP1",
		Rates: map[string]*Rate{
			"RT_WEEK": {
				ID:              "RT_WEEK",
				Weight:          0,
				ActivationTimes: "* * * * 1-5",
				IntervalRates: []*IntervalRate{
					{
						IntervalStart: 0,
						RecurrentFee:  0.12,
						Unit:          time.Minute,
						Increment:     time.Minute,
					},
					{
						IntervalStart: time.Minute,
						RecurrentFee:  0.06,
						Unit:          time.Minute,
						Increment:     time.Second,
					},
				},
			},
			"RT_Custom": {
				ID:              "RT_Custom",
				Weight:          0,
				ActivationTimes: "* * * * 1-5",
				IntervalRates: []*IntervalRate{
					{
						IntervalStart: time.Second,
						RecurrentFee:  0.12,
						Unit:          time.Minute,
						Increment:     time.Minute,
					},
					{
						IntervalStart: time.Second,
						RecurrentFee:  0.19,
						Unit:          time.Minute,
						Increment:     time.Second,
					},
					{
						IntervalStart: 15 * time.Second,
						RecurrentFee:  0.4,
						Unit:          time.Minute,
						Increment:     time.Second,
					},
					{
						IntervalStart: 10 * time.Second,
						RecurrentFee:  0.27,
						Unit:          time.Minute,
						Increment:     time.Second,
					},
				},
			},
			"RT_WEEKEND": {
				ID:              "RT_WEEKEND",
				Weight:          10,
				ActivationTimes: "* * * * 0,6",
				IntervalRates: []*IntervalRate{
					{
						IntervalStart: 10 * time.Second,
						RecurrentFee:  0.06,
						Unit:          time.Minute,
						Increment:     time.Second,
					},
					{
						IntervalStart: time.Minute,
						RecurrentFee:  0.18,
						Unit:          time.Minute,
						Increment:     time.Second,
					},
					{
						IntervalStart: 18 * time.Second,
						RecurrentFee:  0.12,
						Unit:          time.Minute,
						Increment:     time.Second,
					},
				},
			},
			"RT_CHRISTMAS": {
				ID:              "RT_CHRISTMAS",
				Weight:          30,
				ActivationTimes: "* * 24 12 *",
				IntervalRates: []*IntervalRate{
					{
						IntervalStart: 0,
						RecurrentFee:  0.06,
						Unit:          time.Minute,
						Increment:     time.Second,
					},
				},
			},
		},
	}
	exp := &RateProfile{
		Tenant: "cgrates.org",
		ID:     "RP1",
		Rates: map[string]*Rate{
			"RT_WEEK": {
				ID:              "RT_WEEK",
				Weight:          0,
				ActivationTimes: "* * * * 1-5",
				IntervalRates: []*IntervalRate{
					{
						IntervalStart: 0,
						RecurrentFee:  0.12,
						Unit:          time.Minute,
						Increment:     time.Minute,
					},
					{
						IntervalStart: time.Minute,
						RecurrentFee:  0.06,
						Unit:          time.Minute,
						Increment:     time.Second,
					},
				},
			},
			"RT_WEEKEND": {
				ID:              "RT_WEEKEND",
				Weight:          10,
				ActivationTimes: "* * * * 0,6",
				IntervalRates: []*IntervalRate{
					{
						IntervalStart: 10 * time.Second,
						RecurrentFee:  0.06,
						Unit:          time.Minute,
						Increment:     time.Second,
					},
					{
						IntervalStart: 18 * time.Second,
						RecurrentFee:  0.12,
						Unit:          time.Minute,
						Increment:     time.Second,
					},
					{
						IntervalStart: time.Minute,
						RecurrentFee:  0.18,
						Unit:          time.Minute,
						Increment:     time.Second,
					},
				},
			},
			"RT_Custom": {
				ID:              "RT_Custom",
				Weight:          0,
				ActivationTimes: "* * * * 1-5",
				IntervalRates: []*IntervalRate{
					{
						IntervalStart: time.Second,
						RecurrentFee:  0.12,
						Unit:          time.Minute,
						Increment:     time.Minute,
					},
					{
						IntervalStart: time.Second,
						RecurrentFee:  0.19,
						Unit:          time.Minute,
						Increment:     time.Second,
					},
					{
						IntervalStart: 10 * time.Second,
						RecurrentFee:  0.27,
						Unit:          time.Minute,
						Increment:     time.Second,
					},
					{
						IntervalStart: 15 * time.Second,
						RecurrentFee:  0.4,
						Unit:          time.Minute,
						Increment:     time.Second,
					},
				},
			},
			"RT_CHRISTMAS": {
				ID:              "RT_CHRISTMAS",
				Weight:          30,
				ActivationTimes: "* * 24 12 *",
				IntervalRates: []*IntervalRate{
					{
						IntervalStart: 0,
						RecurrentFee:  0.06,
						Unit:          time.Minute,
						Increment:     time.Second,
					},
				},
			},
		},
	}
	rPrf.Sort()
	if !reflect.DeepEqual(rPrf, exp) {
		t.Errorf("Expected: %v,\n received: %v", utils.ToJSON(exp), utils.ToJSON(rPrf))
	}
}

func TestRateProfileCompile(t *testing.T) {
	rt := &RateProfile{
		Rates: map[string]*Rate{
			"randomVal1": {
				ID:              "RT_CHRISTMAS",
				Weight:          30,
				ActivationTimes: "* * 24 12 *",
			},
		},
		Tenant: "cgrates.org",
		ID:     "RTP1",
	}
	expectedATime, err := cron.ParseStandard("* * 24 12 *")
	if err != nil {
		t.Fatal(err)
	}
	expRt := &RateProfile{
		Rates: map[string]*Rate{
			"randomVal1": {
				ID:              "RT_CHRISTMAS",
				Weight:          30,
				ActivationTimes: "* * 24 12 *",
				sched:           expectedATime,
				uID:             utils.ConcatenatedKey(rt.Tenant, rt.ID, "RT_CHRISTMAS"),
			},
		},
		Tenant:  "cgrates.org",
		ID:      "RTP1",
		minCost: new(decimal.Big).SetFloat64(rt.MinCost),
		maxCost: new(decimal.Big).SetFloat64(rt.MaxCost),
	}
	if err := rt.Compile(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rt, expRt) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(expRt), utils.ToJSON(rt))
	}
}

func TestRateUID(t *testing.T) {
	rt := &RateProfile{
		Rates: map[string]*Rate{
			"randomVal1": {
				ID:              "RT_CHRISTMAS",
				Weight:          30,
				ActivationTimes: "* * 24 12 *",
				uID:             "randomID",
			},
		},
	}
	expected := "randomID"
	if newID := rt.Rates["randomVal1"].UID(); !reflect.DeepEqual(newID, expected) {
		t.Errorf("Expected %+v, received %+v", expected, newID)
	}
}

func TestRateProfileCompileError(t *testing.T) {
	rt := &RateProfile{
		Rates: map[string]*Rate{
			"randomVal": {
				ID:              "RT_CHRISTMAS",
				Weight:          30,
				ActivationTimes: "* * * *",
			},
		},
	}
	expectedErr := "expected exactly 5 fields, found 4: [* * * *]"
	if err := rt.Compile(); err == nil || err.Error() != expectedErr {
		t.Errorf("Expected %+v, received %+v ", expectedErr, err)
	}
}

func TestRateCompileChristmasTime(t *testing.T) {
	rt := &Rate{
		ID:              "RT_CHRISTMAS",
		Weight:          30,
		ActivationTimes: "* * 24 12 *",
	}
	expTime, err := cron.ParseStandard("* * 24 12 *")
	if err != nil {
		t.Error(err)
	}
	expectedRt := &Rate{
		ID:              "RT_CHRISTMAS",
		Weight:          30,
		ActivationTimes: "* * 24 12 *",
		sched:           expTime,
	}
	if err := rt.Compile(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedRt, rt) {
		t.Errorf("Expected %+v, received %+v", expectedRt, rt)
	}
}

func TestRateCompileEmptyActivationTime(t *testing.T) {
	rt := &Rate{
		ID:              "RT_CHRISTMAS",
		Weight:          30,
		ActivationTimes: utils.EmptyString,
	}
	expTime, err := cron.ParseStandard("* * * * *")
	if err != nil {
		t.Error(err)
	}
	expectedRt := &Rate{
		ID:              "RT_CHRISTMAS",
		Weight:          30,
		ActivationTimes: utils.EmptyString,
		sched:           expTime,
	}
	if err := rt.Compile(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rt, expectedRt) {
		t.Errorf("Expected %+v, received %+v", expectedRt, rt)
	}
}

func TestRateProfileRunTimes(t *testing.T) {
	rt := &Rate{
		ID: "RATE0",
		IntervalRates: []*IntervalRate{
			{
				IntervalStart: 0,
			},
		},
	}
	rt.Compile()

	sTime := time.Date(2020, time.June, 28, 18, 56, 05, 0, time.UTC)
	eTime := sTime.Add(2 * time.Minute)
	eRTimes := [][]time.Time{
		{time.Date(2020, time.June, 28, 18, 56, 0, 0, time.UTC),
			time.Time{}},
	}

	if rTimes, err := rt.RunTimes(sTime, eTime, 10); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eRTimes, rTimes) {
		t.Errorf("expecting: %+v, received: %+v", eRTimes, rTimes)
	}

	rt = &Rate{
		ID:              "RT_CHRISTMAS",
		Weight:          30,
		ActivationTimes: "* * 24 12 *",
	}
	rt.Compile()

	// sTime and eTime inside the activation interval
	sTime = time.Date(2020, 12, 24, 12, 0, 0, 0, time.UTC)
	eTime = sTime.Add(time.Hour)
	eRTimes = [][]time.Time{
		{time.Date(2020, 12, 24, 12, 0, 0, 0, time.UTC), time.Date(2020, 12, 25, 0, 0, 0, 0, time.UTC)},
	}
	if rTimes, err := rt.RunTimes(sTime, eTime, 10); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eRTimes, rTimes) {
		t.Errorf("expecting: %+v, received: %+v", eRTimes, rTimes)
	}
	// sTime smaller than activation time, eTime equals aTime
	sTime = time.Date(2020, 12, 23, 23, 0, 0, 0, time.UTC)
	eTime = sTime.Add(time.Hour)
	eRTimes = nil // cannot cover full interval
	if rTimes, err := rt.RunTimes(sTime, eTime, 10); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eRTimes, rTimes) {
		t.Errorf("expecting: %+v, received: %+v", eRTimes, rTimes)
	}

	// sTime smaller than activation time but first aTime inside, eTime inside activation interval
	sTime = time.Date(2020, 12, 23, 23, 59, 59, 0, time.UTC)
	eTime = sTime.Add(time.Hour)
	eRTimes = [][]time.Time{
		{time.Date(2020, 12, 24, 0, 0, 0, 0, time.UTC), time.Date(2020, 12, 25, 0, 0, 0, 0, time.UTC)},
	}
	if rTimes, err := rt.RunTimes(sTime, eTime, 1000000); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eRTimes, rTimes) {
		t.Errorf("expecting: %+v, received: %+v", eRTimes, rTimes)
	}

	// sTime way before aTime, eTime inside aInterval
	sTime = time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)
	eTime = time.Date(2021, 12, 24, 0, 1, 0, 0, time.UTC)
	eRTimes = [][]time.Time{
		{time.Date(2021, 12, 24, 0, 0, 0, 0, time.UTC), time.Date(2021, 12, 25, 0, 0, 0, 0, time.UTC)},
	}
	if rTimes, err := rt.RunTimes(sTime, eTime, 1000000); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eRTimes, rTimes) {
		t.Errorf("expecting: %+v, received: %+v", eRTimes, rTimes)
	}
}

func TestRateProfileRunTimesMaxIterations(t *testing.T) {
	rt := &Rate{
		ID: "RATE0",
		IntervalRates: []*IntervalRate{
			{
				IntervalStart: 0,
			},
		},
		ActivationTimes: "* * 24 12 *",
	}
	err := rt.Compile()
	if err != nil {
		t.Error(err)
	}
	sTime := time.Date(2020, 12, 24, 23, 30, 0, 0, time.UTC)
	eTime := time.Date(2021, 12, 25, 23, 30, 0, 0, time.UTC)
	expectedErr := "maximum iterations reached"
	if _, err := rt.RunTimes(sTime, eTime, 2); err == nil || err.Error() != expectedErr {
		t.Errorf("Expected %+v, received %+v", expectedErr, err)
	}
}

func TestRateProfileRunTimesPassingActivationTIme(t *testing.T) {
	rt := &Rate{
		ID: "RATE0",
		IntervalRates: []*IntervalRate{
			{
				IntervalStart: 0,
			},
		},
		ActivationTimes: "* * 24 * *",
	}
	err := rt.Compile()
	if err != nil {
		t.Error(err)
	}
	sTime := time.Date(2020, 12, 23, 0, 0, 0, 0, time.UTC)
	eTime := time.Date(2020, 12, 27, 0, 0, 0, 0, time.UTC)
	expectedTime := [][]time.Time{
		{time.Date(2020, 12, 24, 0, 0, 0, 0, time.UTC), time.Date(2020, 12, 25, 0, 0, 0, 0, time.UTC)},
	}
	if rTimes, err := rt.RunTimes(sTime, eTime, 2); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedTime, rTimes) {
		t.Errorf("Expected %+v, received %+v", expectedTime, rTimes)
	}
}

func TestCostForIntervals(t *testing.T) {
	rt0 := &Rate{
		ID: "RATE0",
		IntervalRates: []*IntervalRate{
			{
				IntervalStart: time.Duration(0),
				Unit:          time.Duration(1 * time.Minute),
				Increment:     time.Duration(1 * time.Minute),
				RecurrentFee:  2.4,
			},
			{
				IntervalStart: time.Duration(60 * time.Second),
				Unit:          time.Duration(1 * time.Minute),
				Increment:     time.Duration(1 * time.Second),
				RecurrentFee:  2.4,
			},
		},
	}
	rt0.Compile()
	rt1 := &Rate{
		ID: "RATE1",
		IntervalRates: []*IntervalRate{
			{

				IntervalStart: time.Duration(0),
				Unit:          time.Duration(1 * time.Minute),
				Increment:     time.Duration(1 * time.Second),
				RecurrentFee:  1.2,
			},
			{

				IntervalStart: time.Duration(2 * time.Minute),
				Unit:          time.Duration(1 * time.Minute),
				Increment:     time.Duration(1 * time.Second),
				RecurrentFee:  0.6,
			},
		},
	}
	rt1.Compile()
	rtIvls := []*RateSInterval{
		{
			UsageStart: time.Duration(0),
			Increments: []*RateSIncrement{
				{
					UsageStart:        time.Duration(0),
					Usage:             time.Duration(time.Minute),
					Rate:              rt0,
					IntervalRateIndex: 0,
					CompressFactor:    1,
				},
				{
					UsageStart:        time.Duration(time.Minute),
					Usage:             time.Duration(30 * time.Second),
					Rate:              rt0,
					IntervalRateIndex: 1,
					CompressFactor:    30,
				},
			},
			CompressFactor: 1,
		},
		{
			UsageStart: time.Duration(90 * time.Second),
			Increments: []*RateSIncrement{
				{
					UsageStart:        time.Duration(90 * time.Second),
					Usage:             time.Duration(30 * time.Second),
					Rate:              rt1,
					IntervalRateIndex: 0,
					CompressFactor:    30,
				},
				{
					UsageStart:        time.Duration(2 * time.Minute),
					Usage:             time.Duration(10 * time.Second),
					Rate:              rt1,
					IntervalRateIndex: 1,
					CompressFactor:    10,
				},
			},
			CompressFactor: 1,
		},
	}
	eDcml, _ := new(decimal.Big).SetFloat64(4.3).Float64()
	if cost, _ := CostForIntervals(rtIvls).Float64(); cost != eDcml {
		t.Errorf("eDcml: %f, received: %+v", eDcml, cost)
	}
}

func TestCostForIntervalsWIthFixedFee(t *testing.T) {
	rt0 := &Rate{
		ID: "RATE0",
		IntervalRates: []*IntervalRate{
			{
				IntervalStart: time.Duration(0),
				FixedFee:      0.4,
				RecurrentFee:  2.4,
				Unit:          time.Duration(1 * time.Minute),
				Increment:     time.Duration(1 * time.Minute),
			},
			{
				IntervalStart: time.Duration(60 * time.Second),
				RecurrentFee:  2.4,
				Unit:          time.Duration(1 * time.Minute),
				Increment:     time.Duration(1 * time.Second),
			},
		},
	}
	rt0.Compile()
	rt1 := &Rate{
		ID: "RATE1",
		IntervalRates: []*IntervalRate{
			{

				IntervalStart: time.Duration(0),
				FixedFee:      0.2,
				RecurrentFee:  1.2,
				Unit:          time.Duration(1 * time.Minute),
				Increment:     time.Duration(1 * time.Second),
			},
			{

				IntervalStart: time.Duration(2 * time.Minute),
				RecurrentFee:  0.6,
				Unit:          time.Duration(1 * time.Minute),
				Increment:     time.Duration(1 * time.Second),
			},
		},
	}
	rt1.Compile()
	rtIvls := []*RateSInterval{
		{
			UsageStart: time.Duration(0),
			Increments: []*RateSIncrement{
				{ // cost 0,4
					UsageStart:        time.Duration(0),
					Rate:              rt0,
					IntervalRateIndex: 0,
					CompressFactor:    1,
					Usage:             utils.InvalidDuration,
				},
				{ // cost 2,4
					UsageStart:        time.Duration(0),
					Rate:              rt0,
					IntervalRateIndex: 0,
					CompressFactor:    1,
					Usage:             time.Duration(time.Minute),
				},
				{ // cost 1,2
					UsageStart:        time.Duration(time.Minute),
					Rate:              rt0,
					IntervalRateIndex: 1,
					CompressFactor:    30,
					Usage:             time.Duration(30 * time.Second),
				},
			},
			CompressFactor: 1,
		},
		{
			UsageStart: time.Duration(90 * time.Second),
			Increments: []*RateSIncrement{
				{ // cost 0,2
					UsageStart:        time.Duration(90 * time.Second),
					Rate:              rt1,
					IntervalRateIndex: 0,
					CompressFactor:    1,
					Usage:             utils.InvalidDuration,
				},
				{ // cost 0,6
					UsageStart:        time.Duration(90 * time.Second),
					Rate:              rt1,
					IntervalRateIndex: 0,
					CompressFactor:    30,
					Usage:             time.Duration(30 * time.Second),
				},
				{ // cost 0,1
					UsageStart:        time.Duration(2 * time.Minute),
					Rate:              rt1,
					IntervalRateIndex: 1,
					CompressFactor:    10,
					Usage:             time.Duration(10 * time.Second),
				},
			},
			CompressFactor: 1,
		},
	}
	eDcml, _ := new(decimal.Big).SetFloat64(4.9).Float64()
	if cost, _ := CostForIntervals(rtIvls).Float64(); cost != eDcml {
		t.Errorf("eDcml: %f, received: %+v", eDcml, cost)
	}
}
