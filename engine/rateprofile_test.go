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
				ID:             "RT_WEEK",
				Weight:         0,
				ActivationTime: "* * * * 1-5",
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
				ID:             "RT_Custom",
				Weight:         0,
				ActivationTime: "* * * * 1-5",
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
				ID:             "RT_WEEKEND",
				Weight:         10,
				ActivationTime: "* * * * 0,6",
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
				ID:             "RT_CHRISTMAS",
				Weight:         30,
				ActivationTime: "* * 24 12 *",
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
				ID:             "RT_WEEK",
				Weight:         0,
				ActivationTime: "* * * * 1-5",
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
				ID:             "RT_WEEKEND",
				Weight:         10,
				ActivationTime: "* * * * 0,6",
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
				ID:             "RT_Custom",
				Weight:         0,
				ActivationTime: "* * * * 1-5",
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
				ID:             "RT_CHRISTMAS",
				Weight:         30,
				ActivationTime: "* * 24 12 *",
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

func TestRateProfileCompile(t *testing.T) {
	rt := &Rate{
		ID:             "RT_CHRISTMAS",
		Weight:         30,
		ActivationTime: "* * 24 12 *",
	}
	if err := rt.Compile(); err != nil {
		t.Error(err)
	}

	refTime := time.Date(2018, time.January, 7, 17, 0, 0, 0, time.UTC)
	exp := time.Date(2018, time.December, 24, 0, 0, 0, 0, time.UTC)
	if rcv := rt.NextActivationTime(refTime); !rcv.Equal(exp) {
		t.Errorf("Expected: %v,\n received: %v", exp, rcv)
	}
	rt = &Rate{
		ID:             "RT_CHRISTMAS",
		Weight:         30,
		ActivationTime: utils.EmptyString,
	}
	if err := rt.Compile(); err != nil {
		t.Error(err)
	}

	exp = time.Date(2018, time.January, 7, 17, 1, 0, 0, time.UTC)
	if rcv := rt.NextActivationTime(refTime); !rcv.Equal(exp) {
		t.Errorf("Expected: %v,\n received: %v", exp, rcv)
	}

	rt = &Rate{
		ID:             "RT_CHRISTMAS",
		Weight:         30,
		ActivationTime: "error",
	}
	if err := rt.Compile(); err == nil || err.Error() != "expected exactly 5 fields, found 1: [error]" {
		t.Error(err)
	}
}
