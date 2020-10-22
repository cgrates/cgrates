/*
Real-time Online/Offline Charging System (OerS) for Telecom & ISP environments
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

package rates

import (
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/utils"

	"github.com/cgrates/cgrates/engine"
)

func TestOrderRatesOnIntervals(t *testing.T) {
	rt0 := &engine.Rate{
		ID:     "RATE0",
		Weight: 0,
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: time.Duration(0),
			},
		},
	}
	rt0.Compile()
	rtChristmas := &engine.Rate{
		ID:              "RT_CHRISTMAS",
		ActivationTimes: "* * 24 12 *",
		Weight:          50,
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: time.Duration(0),
			},
		},
	}
	rtChristmas.Compile()
	allRts := []*engine.Rate{rt0, rtChristmas}
	expOrdered := []*orderedRate{
		{
			time.Duration(0),
			rt0,
		},
	}

	sTime := time.Date(2020, time.June, 28, 18, 56, 05, 0, time.UTC)
	usage := time.Duration(2 * time.Minute)
	if ordRts, err := orderRatesOnIntervals(
		allRts, sTime, usage, true, 10); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expOrdered, ordRts) {
		t.Errorf("expecting: %s\n, received: %s",
			utils.ToIJSON(expOrdered), utils.ToIJSON(ordRts))
	}

	expOrdered = []*orderedRate{
		{
			time.Duration(0),
			rt0,
		},
		{
			time.Duration(55 * time.Second),
			rtChristmas,
		},
	}
	sTime = time.Date(2020, time.December, 23, 23, 59, 05, 0, time.UTC)
	usage = time.Duration(2 * time.Minute)
	if ordRts, err := orderRatesOnIntervals(
		allRts, sTime, usage, true, 10); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expOrdered, ordRts) {
		t.Errorf("expecting: %s\n, received: %s",
			utils.ToIJSON(expOrdered), utils.ToIJSON(ordRts))
	}

	expOrdered = []*orderedRate{
		{
			time.Duration(0),
			rt0,
		},
		{
			time.Duration(55 * time.Second),
			rtChristmas,
		},
		{
			time.Duration(86455 * time.Second),
			rt0,
		},
	}
	sTime = time.Date(2020, time.December, 23, 23, 59, 05, 0, time.UTC)
	usage = time.Duration(25 * time.Hour)
	if ordRts, err := orderRatesOnIntervals(
		allRts, sTime, usage, true, 10); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expOrdered, ordRts) {
		t.Errorf("expecting: %s\n, received: %s",
			utils.ToIJSON(expOrdered), utils.ToIJSON(ordRts))
	}
	if ordRts, err := orderRatesOnIntervals(
		allRts, sTime, usage, true, 10); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expOrdered, ordRts) {
		t.Errorf("expecting: %s\n, received: %s",
			utils.ToIJSON(expOrdered), utils.ToIJSON(ordRts))
	}

	rts := []*engine.Rate{rtChristmas}
	expOrdered = nil
	sTime = time.Date(2020, time.December, 25, 23, 59, 05, 0, time.UTC)
	usage = time.Duration(2 * time.Minute)
	if ordRts, err := orderRatesOnIntervals(
		rts, sTime, usage, true, 10); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expOrdered, ordRts) {
		t.Errorf("expecting: %s\n, received: %s",
			utils.ToIJSON(expOrdered), utils.ToIJSON(ordRts))
	}
}

func TestNewRatesWithWinner(t *testing.T) {
	rt := &rateWithTimes{
		uId: "randomID",
	}
	expected := &ratesWithWinner{
		rts: map[string]*rateWithTimes{
			"randomID": rt,
		},
		wnr: rt,
	}
	if !reflect.DeepEqual(expected, newRatesWithWinner(rt)) {
		t.Errorf("Expected %+v, received %+v", expected, newRatesWithWinner(rt))
	}
}

func TestOrderRatesOnIntervalCaseMaxIterations(t *testing.T) {
	rt1 := &engine.Rate{
		ID:              "RT_1",
		ActivationTimes: "1 * * * *",
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: time.Duration(0),
			},
		},
	}
	err := rt1.Compile()
	if err != nil {
		t.Error(err)
	}
	aRts := []*engine.Rate{rt1}
	sTime := time.Date(2020, 01, 02, 0, 1, 0, 0, time.UTC)
	usage := 96 * time.Hour
	expectedErr := "maximum iterations reached"
	if _, err := orderRatesOnIntervals(aRts, sTime, usage, false, 1); err == nil || err.Error() != expectedErr {
		t.Errorf("Expected %+v, received %+v", expectedErr, err)
	}
}

func TestOrderRatesOnIntervalIsDirectionFalse(t *testing.T) {
	rt1 := &engine.Rate{
		ID:              "RT_1",
		ActivationTimes: "* * 27 02 *",
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: time.Duration(0),
			},
		},
	}
	err := rt1.Compile()
	if err != nil {
		t.Error(err)
	}
	expected := []*engine.RateSInterval{
		{
			UsageStart: 0,
			Increments: []*engine.RateSIncrement{
				{
					Rate: &engine.Rate{
						ID:              "RT_1",
						ActivationTimes: "* * 27 02 *",
						IntervalRates: []*engine.IntervalRate{
							{
								IntervalStart: time.Duration(0),
							},
						},
					},
				},
			},
		},
	}
	err = expected[0].Increments[0].Rate.Compile()
	if err != nil {
		t.Error(err)
	}
	aRts := []*engine.Rate{rt1}
	sTime := time.Date(0001, 02, 27, 0, 0, 0, 0, time.UTC)
	usage := 48 * time.Hour
	if ordRts, err := orderRatesOnIntervals(aRts, sTime, usage, false, 5); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(ordRts, expected) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(expected), utils.ToJSON(ordRts))
	}
}

func TestOrderRatesOnIntervalWinnNill(t *testing.T) {
	rt1 := &engine.Rate{
		ID:              "RT_1",
		ActivationTimes: "* * 1 * *",
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: time.Duration(0),
			},
		},
	}
	err := rt1.Compile()
	if err != nil {
		t.Error(err)
	}
	expected := []*engine.RateSInterval{
		{
			UsageStart: 0,
			Increments: []*engine.RateSIncrement{
				{
					Rate: &engine.Rate{
						ID:              "RT_1",
						ActivationTimes: "* * 1 * *",
						IntervalRates: []*engine.IntervalRate{
							{
								IntervalStart: time.Duration(0),
							},
						},
					},
				},
			},
		},
	}
	err = expected[0].Increments[0].Rate.Compile()
	if err != nil {
		t.Error(err)
	}
	aRts := []*engine.Rate{rt1}
	sTime := time.Date(2020, 12, 1, 0, 0, 0, 0, time.UTC)
	usage := 96 * time.Hour
	if ordRts, err := orderRatesOnIntervals(aRts, sTime, usage, true, 4); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(ordRts, expected) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(expected), utils.ToJSON(ordRts))
	}
}

func TestOrderRatesOnIntervalIntervalStartHigherThanEndIdx(t *testing.T) {
	rt1 := &engine.Rate{
		ID:              "RT_1",
		ActivationTimes: "* * 1 * *",
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 48 * time.Hour,
			},
		},
	}
	err := rt1.Compile()
	if err != nil {
		t.Error(err)
	}
	expected := []*engine.RateSInterval{
		{
			UsageStart: 0,
			Increments: []*engine.RateSIncrement{
				{
					Rate: &engine.Rate{
						ID:              "RT_1",
						ActivationTimes: "* * 1 * *",
						IntervalRates: []*engine.IntervalRate{
							{
								IntervalStart: 48 * time.Hour,
							},
						},
					},
					IntervalRateIndex: 1,
				},
			},
		},
	}
	err = expected[0].Increments[0].Rate.Compile()
	if err != nil {
		t.Error(err)
	}
	aRts := []*engine.Rate{rt1}
	sTime := time.Date(2020, 12, 1, 0, 0, 0, 0, time.UTC)
	usage := 48 * time.Hour
	if _, err := orderRatesOnIntervals(aRts, sTime, usage, false, 4); err != nil {
		t.Error(err)
	}
}

func TestOrderRatesOnIntervalStartLowerThanEndIdx(t *testing.T) {
	rt1 := &engine.Rate{
		ID:              "RT_1",
		ActivationTimes: "* * 1 * *",
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 23 * time.Hour,
			},
			{
				IntervalStart: -time.Hour,
			},
		},
	}
	err := rt1.Compile()
	if err != nil {
		t.Error(err)
	}
	expected := []*engine.RateSInterval{
		{
			UsageStart: 0,
			Increments: []*engine.RateSIncrement{
				{
					Rate: &engine.Rate{
						ID:              "RT_1",
						ActivationTimes: "* * 1 * *",
						IntervalRates: []*engine.IntervalRate{
							{
								IntervalStart: 23 * time.Hour,
							},
							{
								IntervalStart: -time.Hour,
							},
						},
					},
					IntervalRateIndex: 1,
				},
			},
		},
	}
	err = expected[0].Increments[0].Rate.Compile()
	if err != nil {
		t.Error(err)
	}
	aRts := []*engine.Rate{rt1}
	sTime := time.Date(2020, 12, 1, 0, 0, 0, 0, time.UTC)
	usage := 48 * time.Hour
	if ordRts, err := orderRatesOnIntervals(aRts, sTime, usage, false, 4); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(ordRts, expected) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(expected), utils.ToJSON(ordRts))
	}
}
