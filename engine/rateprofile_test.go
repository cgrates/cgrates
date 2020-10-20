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
						IntervalStart: time.Duration(0 * time.Second),
						Value:         0.12,
						Unit:          time.Duration(1 * time.Minute),
						Increment:     time.Duration(1 * time.Minute),
					},
					{
						IntervalStart: time.Duration(1 * time.Minute),
						Value:         0.06,
						Unit:          time.Duration(1 * time.Minute),
						Increment:     time.Duration(1 * time.Second),
					},
				},
			},
			"RT_Custom": {
				ID:              "RT_Custom",
				Weight:          0,
				ActivationTimes: "* * * * 1-5",
				IntervalRates: []*IntervalRate{
					{
						IntervalStart: time.Duration(1 * time.Second),
						Value:         0.12,
						Unit:          time.Duration(1 * time.Minute),
						Increment:     time.Duration(1 * time.Minute),
					},
					{
						IntervalStart: time.Duration(1 * time.Second),
						Value:         0.19,
						Unit:          time.Duration(1 * time.Minute),
						Increment:     time.Duration(1 * time.Second),
					},
					{
						IntervalStart: time.Duration(15 * time.Second),
						Value:         0.4,
						Unit:          time.Duration(1 * time.Minute),
						Increment:     time.Duration(1 * time.Second),
					},
					{
						IntervalStart: time.Duration(10 * time.Second),
						Value:         0.27,
						Unit:          time.Duration(1 * time.Minute),
						Increment:     time.Duration(1 * time.Second),
					},
				},
			},
			"RT_WEEKEND": {
				ID:              "RT_WEEKEND",
				Weight:          10,
				ActivationTimes: "* * * * 0,6",
				IntervalRates: []*IntervalRate{
					{
						IntervalStart: time.Duration(10 * time.Second),
						Value:         0.06,
						Unit:          time.Duration(1 * time.Minute),
						Increment:     time.Duration(1 * time.Second),
					},
					{
						IntervalStart: time.Duration(1 * time.Minute),
						Value:         0.18,
						Unit:          time.Duration(1 * time.Minute),
						Increment:     time.Duration(1 * time.Second),
					},
					{
						IntervalStart: time.Duration(18 * time.Second),
						Value:         0.12,
						Unit:          time.Duration(1 * time.Minute),
						Increment:     time.Duration(1 * time.Second),
					},
				},
			},
			"RT_CHRISTMAS": {
				ID:              "RT_CHRISTMAS",
				Weight:          30,
				ActivationTimes: "* * 24 12 *",
				IntervalRates: []*IntervalRate{
					{
						IntervalStart: time.Duration(0 * time.Second),
						Value:         0.06,
						Unit:          time.Duration(1 * time.Minute),
						Increment:     time.Duration(1 * time.Second),
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
						IntervalStart: time.Duration(0 * time.Second),
						Value:         0.12,
						Unit:          time.Duration(1 * time.Minute),
						Increment:     time.Duration(1 * time.Minute),
					},
					{
						IntervalStart: time.Duration(1 * time.Minute),
						Value:         0.06,
						Unit:          time.Duration(1 * time.Minute),
						Increment:     time.Duration(1 * time.Second),
					},
				},
			},
			"RT_WEEKEND": {
				ID:              "RT_WEEKEND",
				Weight:          10,
				ActivationTimes: "* * * * 0,6",
				IntervalRates: []*IntervalRate{
					{
						IntervalStart: time.Duration(10 * time.Second),
						Value:         0.06,
						Unit:          time.Duration(1 * time.Minute),
						Increment:     time.Duration(1 * time.Second),
					},
					{
						IntervalStart: time.Duration(18 * time.Second),
						Value:         0.12,
						Unit:          time.Duration(1 * time.Minute),
						Increment:     time.Duration(1 * time.Second),
					},
					{
						IntervalStart: time.Duration(1 * time.Minute),
						Value:         0.18,
						Unit:          time.Duration(1 * time.Minute),
						Increment:     time.Duration(1 * time.Second),
					},
				},
			},
			"RT_Custom": {
				ID:              "RT_Custom",
				Weight:          0,
				ActivationTimes: "* * * * 1-5",
				IntervalRates: []*IntervalRate{
					{
						IntervalStart: time.Duration(1 * time.Second),
						Value:         0.12,
						Unit:          time.Duration(1 * time.Minute),
						Increment:     time.Duration(1 * time.Minute),
					},
					{
						IntervalStart: time.Duration(1 * time.Second),
						Value:         0.19,
						Unit:          time.Duration(1 * time.Minute),
						Increment:     time.Duration(1 * time.Second),
					},
					{
						IntervalStart: time.Duration(10 * time.Second),
						Value:         0.27,
						Unit:          time.Duration(1 * time.Minute),
						Increment:     time.Duration(1 * time.Second),
					},
					{
						IntervalStart: time.Duration(15 * time.Second),
						Value:         0.4,
						Unit:          time.Duration(1 * time.Minute),
						Increment:     time.Duration(1 * time.Second),
					},
				},
			},
			"RT_CHRISTMAS": {
				ID:              "RT_CHRISTMAS",
				Weight:          30,
				ActivationTimes: "* * 24 12 *",
				IntervalRates: []*IntervalRate{
					{
						IntervalStart: time.Duration(0 * time.Second),
						Value:         0.06,
						Unit:          time.Duration(1 * time.Minute),
						Increment:     time.Duration(1 * time.Second),
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

func TestRateCompile(t *testing.T) {
	rt := &RateProfile{
		Rates: map[string]*Rate{
			"randomVal1": {
				ID:              "RT_CHRISTMAS",
				Weight:          30,
				ActivationTimes: "* * 24 12 *",
			},
			"randomVal2": {
				ID:              "RT_CHRISTMAS",
				Weight:          30,
				ActivationTimes: utils.EmptyString,
			},
		},
	}
	if err := rt.Compile(); err != nil {
		t.Error(err)
	}

	rt.Rates["randomVal1"].ActivationTimes = "* * * *"
	rt.Rates["randomVal2"].ActivationTimes = "* * * *"
	expectedErr := "expected exactly 5 fields, found 4: [* * * *]"
	if err := rt.Compile(); err == nil || err.Error() != expectedErr {
		t.Errorf("Expected %+v, received %+v ", expectedErr, err)
	}
}

func TestRateProfileCompile(t *testing.T) {
	rt := &Rate{
		ID:              "RT_CHRISTMAS",
		Weight:          30,
		ActivationTimes: "* * 24 12 *",
	}
	if err := rt.Compile(); err != nil {
		t.Error(err)
	}

	rt = &Rate{
		ID:              "RT_CHRISTMAS",
		Weight:          30,
		ActivationTimes: utils.EmptyString,
	}
	if err := rt.Compile(); err != nil {
		t.Error(err)
	}

	rt = &Rate{
		ID:              "RT_CHRISTMAS",
		Weight:          30,
		ActivationTimes: "error",
	}
	if err := rt.Compile(); err == nil || err.Error() != "expected exactly 5 fields, found 1: [error]" {
		t.Error(err)
	}
}

func TestRateProfileRunTimes(t *testing.T) {
	rt := &Rate{
		ID: "RATE0",
		IntervalRates: []*IntervalRate{
			{
				IntervalStart: time.Duration(0),
			},
		},
	}
	rt.Compile()

	sTime := time.Date(2020, time.June, 28, 18, 56, 05, 0, time.UTC)
	eTime := sTime.Add(time.Duration(2 * time.Minute))
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
	eTime = sTime.Add(time.Duration(time.Hour))
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
	eTime = sTime.Add(time.Duration(time.Hour))
	eRTimes = nil // cannot cover full interval
	if rTimes, err := rt.RunTimes(sTime, eTime, 10); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eRTimes, rTimes) {
		t.Errorf("expecting: %+v, received: %+v", eRTimes, rTimes)
	}

	// sTime smaller than activation time but first aTime inside, eTime inside activation interval
	sTime = time.Date(2020, 12, 23, 23, 59, 59, 0, time.UTC)
	eTime = sTime.Add(time.Duration(time.Hour))
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
