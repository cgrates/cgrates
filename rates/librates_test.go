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
		ID: "RATE0",
		Weights: utils.DynamicWeights{
			{
				Weight: 0,
			},
		},
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
			},
		},
	}
	rt0.Compile()
	rtChristmas := &engine.Rate{
		ID:              "RT_CHRISTMAS",
		ActivationTimes: "* * 24 12 *",
		Weights: utils.DynamicWeights{
			{
				Weight: 50,
			},
		},
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
			},
		},
	}
	rtChristmas.Compile()
	allRts := []*engine.Rate{rt0, rtChristmas}
	expOrdered := []*orderedRate{
		{
			0,
			rt0,
		},
	}
	wghts := []float64{10, 50}
	sTime := time.Date(2020, time.June, 28, 18, 56, 05, 0, time.UTC)
	usage := 2 * time.Minute
	if ordRts, err := orderRatesOnIntervals(
		allRts, wghts, sTime, usage, true, 10); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expOrdered, ordRts) {
		t.Errorf("expecting: %s\n, received: %s",
			utils.ToIJSON(expOrdered), utils.ToIJSON(ordRts))
	}

	expOrdered = []*orderedRate{
		{
			0,
			rt0,
		},
		{
			55 * time.Second,
			rtChristmas,
		},
	}
	sTime = time.Date(2020, time.December, 23, 23, 59, 05, 0, time.UTC)
	usage = 2 * time.Minute
	if ordRts, err := orderRatesOnIntervals(
		allRts, wghts, sTime, usage, true, 10); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expOrdered, ordRts) {
		t.Errorf("expecting: %s\n, received: %s",
			utils.ToIJSON(expOrdered), utils.ToIJSON(ordRts))
	}

	expOrdered = []*orderedRate{
		{
			0,
			rt0,
		},
		{
			55 * time.Second,
			rtChristmas,
		},
		{
			86455 * time.Second,
			rt0,
		},
	}
	sTime = time.Date(2020, time.December, 23, 23, 59, 05, 0, time.UTC)
	usage = 25 * time.Hour
	if ordRts, err := orderRatesOnIntervals(
		allRts, wghts, sTime, usage, true, 10); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expOrdered, ordRts) {
		t.Errorf("expecting: %s\n, received: %s",
			utils.ToIJSON(expOrdered), utils.ToIJSON(ordRts))
	}

	rts := []*engine.Rate{rtChristmas}
	expOrdered = nil
	sTime = time.Date(2020, time.December, 25, 23, 59, 05, 0, time.UTC)
	usage = 2 * time.Minute
	if ordRts, err := orderRatesOnIntervals(
		rts, wghts, sTime, usage, true, 10); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expOrdered, ordRts) {
		t.Errorf("expecting: %s\n, received: %s",
			utils.ToIJSON(expOrdered), utils.ToIJSON(ordRts))
	}
}

func TestOrderRatesOnIntervalsChristmasDay(t *testing.T) {
	rt1 := &engine.Rate{
		ID: "ALWAYS_RATE",
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
			},
		},
	}
	rtCh1 := &engine.Rate{
		ID:              "CHRISTMAS1",
		ActivationTimes: "* 0-6 24 12 *",
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
			},
		},
	}
	rtCh2 := &engine.Rate{
		ID:              "CHRISTMAS2",
		ActivationTimes: "* 7-12 24 12 *",
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
			},
		},
	}
	rtCh3 := &engine.Rate{
		ID:              "CHRISTMAS3",
		ActivationTimes: "* 13-19 24 12 *",
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
			},
		},
	}
	rtCh4 := &engine.Rate{
		ID:              "CHRISTMAS4",
		ActivationTimes: "* 20-23 24 12 *",
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
			},
		},
	}
	wghts := []float64{10, 20, 20, 20, 20}
	allRates := []*engine.Rate{rt1, rtCh1, rtCh2, rtCh3, rtCh4}
	for _, idx := range allRates {
		if err := idx.Compile(); err != nil {
			t.Error(err)
		}
	}
	sTime := time.Date(2020, 12, 23, 22, 0, 0, 0, time.UTC)
	usage := 31 * time.Hour
	expected := []*orderedRate{
		{
			0,
			rt1,
		},
		{
			2 * time.Hour,
			rtCh1,
		},
		{
			9 * time.Hour,
			rtCh2,
		},
		{
			15 * time.Hour,
			rtCh3,
		},
		{
			22 * time.Hour,
			rtCh4,
		},
		{
			26 * time.Hour,
			rt1,
		},
	}
	if ordRts, err := orderRatesOnIntervals(allRates, wghts, sTime, usage, true, 10); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(ordRts, expected) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(ordRts))
	}
}

func TestOrderRatesOnIntervalsDoubleRates1(t *testing.T) {
	rt1 := &engine.Rate{
		ID: "ALWAYS_RATE",
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
			},
		},
	}
	rtCh1 := &engine.Rate{
		ID:              "CHRISTMAS1",
		ActivationTimes: "* * 24 12 *",
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
			},
		},
	}
	rtCh2 := &engine.Rate{
		ID:              "CHRISTMAS2",
		ActivationTimes: "* 18-23 24 12 *",
		Weights: utils.DynamicWeights{
			{
				Weight: 30,
			},
		},
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
			},
		},
	}
	wghts := []float64{10, 20, 30}
	allRates := []*engine.Rate{rt1, rtCh1, rtCh2}
	for _, idx := range allRates {
		if err := idx.Compile(); err != nil {
			t.Error(err)
		}
	}
	sTime := time.Date(2020, 12, 23, 21, 28, 12, 0, time.UTC)
	usage := 31 * time.Hour
	expected := []*orderedRate{
		{
			0,
			rt1,
		},
		{
			2*time.Hour + 31*time.Minute + 48*time.Second,
			rtCh1,
		},
		{
			20*time.Hour + 31*time.Minute + 48*time.Second,
			rtCh2,
		},
		{
			26*time.Hour + 31*time.Minute + 48*time.Second,
			rt1,
		},
	}
	if ordRts, err := orderRatesOnIntervals(allRates, wghts, sTime, usage, true, 10); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(ordRts, expected) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(ordRts))
	}
}

func TestOrderRatesOnIntervalsEveryTwentyFiveMins(t *testing.T) {
	rtTwentyFiveMins := &engine.Rate{
		ID:              "TWENTYFIVE_MINS",
		ActivationTimes: "*/25 * * * *",
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
			},
		},
	}
	rt1 := &engine.Rate{
		ID:              "DAY_RATE",
		ActivationTimes: "* * * * 3",
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
			},
		},
	}
	wghts := []float64{10, 20}
	allRates := []*engine.Rate{rt1, rtTwentyFiveMins}
	for _, idx := range allRates {
		if err := idx.Compile(); err != nil {
			t.Error(err)
		}
	}
	sTime := time.Date(2020, 10, 28, 20, 0, 0, 0, time.UTC)
	usage := time.Hour
	expected := []*orderedRate{
		{
			0,
			rtTwentyFiveMins,
		},
		{
			time.Minute,
			rt1,
		},
		{
			25 * time.Minute,
			rtTwentyFiveMins,
		},
		{
			26 * time.Minute,
			rt1,
		},
		{
			50 * time.Minute,
			rtTwentyFiveMins,
		},
		{
			51 * time.Minute,
			rt1,
		},
	}
	if ordRts, err := orderRatesOnIntervals(allRates, wghts, sTime, usage, true, 10); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(ordRts, expected) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(ordRts))
	}
}

func TestOrderRatesOnIntervalsOneMinutePause(t *testing.T) {
	rt1 := &engine.Rate{
		ID:              "ALWAYS_RATE",
		ActivationTimes: "26 * * * *",
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
			},
		},
	}
	rtFirstInterval := &engine.Rate{
		ID:              "FIRST_INTERVAL",
		ActivationTimes: "0-25 * * * *",
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
			},
		},
	}
	rtSecondINterval := &engine.Rate{
		ID:              "SECOND_INTERVAL",
		ActivationTimes: "27-59 * * * *",
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
			},
		},
	}
	wghts := []float64{10, 20, 20}
	allRates := []*engine.Rate{rt1, rtFirstInterval, rtSecondINterval}
	for _, idx := range allRates {
		if err := idx.Compile(); err != nil {
			t.Error(err)
		}
	}
	sTime := time.Date(2020, 10, 28, 20, 0, 0, 0, time.UTC)
	usage := time.Hour
	expected := []*orderedRate{
		{
			0,
			rtFirstInterval,
		},
		{
			26 * time.Minute,
			rt1,
		},
		{
			27 * time.Minute,
			rtSecondINterval,
		},
	}
	if ordRts, err := orderRatesOnIntervals(allRates, wghts, sTime, usage, true, 10); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(ordRts, expected) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(ordRts))
	}
}

func TestOrderRatesOnIntervalsNewYear(t *testing.T) {
	rt1 := &engine.Rate{
		ID: "ALWAYS_RATE",
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
			},
		},
	}
	rt1NewYear := &engine.Rate{
		ID:              "NEW_YEAR1",
		ActivationTimes: "* 20-23 * * *",
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
			},
		},
	}
	rt1NewYear2 := &engine.Rate{
		ID:              "NEW_YEAR2",
		ActivationTimes: "0-30 22 * * *",
		Weights: utils.DynamicWeights{
			{
				Weight: 30,
			},
		},
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
			},
		},
	}
	wghts := []float64{10, 20, 30}
	allRates := []*engine.Rate{rt1, rt1NewYear, rt1NewYear2}
	for _, idx := range allRates {
		if err := idx.Compile(); err != nil {
			t.Error(err)
		}
	}
	sTime := time.Date(2020, 12, 30, 23, 0, 0, 0, time.UTC)
	usage := 26 * time.Hour
	expected := []*orderedRate{
		{
			0,
			rt1NewYear,
		},
		{
			time.Hour,
			rt1,
		},
		{
			21 * time.Hour,
			rt1NewYear,
		},
		{
			23 * time.Hour,
			rt1NewYear2,
		},
		{
			23*time.Hour + 31*time.Minute,
			rt1NewYear,
		},
		{
			25 * time.Hour,
			rt1,
		},
	}
	if ordRts, err := orderRatesOnIntervals(allRates, wghts, sTime, usage, true, 10); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(ordRts, expected) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(ordRts))
	}
}

func TestOrderRateOnIntervalsEveryHourEveryDay(t *testing.T) {
	rtEveryHour := &engine.Rate{
		ID:              "HOUR_RATE",
		ActivationTimes: "* */1 * * *",
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
			},
		},
	}
	rtEveryDay := &engine.Rate{
		ID:              "DAY_RATE",
		ActivationTimes: "* * 22 * *",
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
			},
		},
	}
	wghts := []float64{10, 20}
	allRates := []*engine.Rate{rtEveryHour, rtEveryDay}
	for _, idx := range allRates {
		if err := idx.Compile(); err != nil {
			t.Error(err)
		}
	}
	sTime := time.Date(2020, 7, 21, 10, 24, 15, 0, time.UTC)
	usage := 49 * time.Hour
	expected := []*orderedRate{
		{
			0,
			rtEveryHour,
		},
		{
			13*time.Hour + 35*time.Minute + 45*time.Second,
			rtEveryDay,
		},
		{
			37*time.Hour + 35*time.Minute + 45*time.Second,
			rtEveryHour,
		},
	}
	if ordRts, err := orderRatesOnIntervals(allRates, wghts, sTime, usage, true, 10); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(ordRts, expected) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(ordRts))
	}
}

func TestOrderRatesOnIntervalsOneHourInThreeRates(t *testing.T) {
	rtOneHour1 := &engine.Rate{
		ID:              "HOUR_RATE_1",
		ActivationTimes: "0-19 * * * *",
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
			},
		},
	}
	rtOneHour2 := &engine.Rate{
		ID:              "HOUR_RATE_2",
		ActivationTimes: "20-39 * * * *",
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
			},
		},
	}
	rtOneHour3 := &engine.Rate{
		ID:              "HOUR_RATE_3",
		ActivationTimes: "40-59 * * * *",
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
			},
		},
	}
	wghts := []float64{20, 20, 20}
	allRates := []*engine.Rate{rtOneHour1, rtOneHour2, rtOneHour3}
	for _, idx := range allRates {
		if err := idx.Compile(); err != nil {
			t.Error(err)
		}
	}
	sTime := time.Date(2020, 7, 15, 10, 59, 59, 0, time.UTC)
	usage := 2 * time.Hour
	expected := []*orderedRate{
		{
			0,
			rtOneHour3,
		},
		{
			time.Second,
			rtOneHour1,
		},
		{
			20*time.Minute + time.Second,
			rtOneHour2,
		},
		{
			40*time.Minute + time.Second,
			rtOneHour3,
		},
		{
			time.Hour + time.Second,
			rtOneHour1,
		},
		{
			time.Hour + 20*time.Minute + time.Second,
			rtOneHour2,
		},
		{
			time.Hour + 40*time.Minute + time.Second,
			rtOneHour3,
		},
	}
	if ordRts, err := orderRatesOnIntervals(allRates, wghts, sTime, usage, true, 10); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(ordRts, expected) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(ordRts))
	}
}

func TestOrderRateOnIntervalsEveryThreeHours(t *testing.T) {
	rtEveryThreeH := &engine.Rate{
		ID:              "EVERY_THREE_RATE",
		ActivationTimes: "* */3 * * *",
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
			},
		},
	}
	rtByDay := &engine.Rate{
		ID:              "DAY_RATE",
		ActivationTimes: "* 15-23 * * *",
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
			},
		},
	}
	wghts := []float64{10, 20}
	allRates := []*engine.Rate{rtEveryThreeH, rtByDay}
	for _, idx := range allRates {
		if err := idx.Compile(); err != nil {
			t.Error(err)
		}
	}
	sTime := time.Date(2020, 7, 15, 0, 0, 0, 0, time.UTC)
	usage := 24 * time.Hour
	expected := []*orderedRate{
		{
			0,
			rtEveryThreeH,
		},
		{
			15 * time.Hour,
			rtByDay,
		},
	}
	if ordRts, err := orderRatesOnIntervals(allRates, wghts, sTime, usage, true, 10); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(ordRts, expected) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(ordRts))
	}
}

func TestOrderRateOnIntervalsTwoRatesInOne(t *testing.T) {
	rtHalfDay1 := &engine.Rate{
		ID:              "HALF_RATE1",
		ActivationTimes: "* 0-11 22 12 *",
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
			},
		},
	}
	rtHalfDay2 := &engine.Rate{
		ID:              "HALF_RATE2",
		ActivationTimes: "* 12-23 22 12 *",
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
			},
		},
	}
	rtHalfDay2r1 := &engine.Rate{
		ID:              "HALF_RATE2.1",
		ActivationTimes: "* 12-16 22 12 *",
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
			},
		},
	}
	rtHalfDay2r2 := &engine.Rate{
		ID:              "HALF_RATE2.2",
		ActivationTimes: "* 18-23 22 12 *",
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
			},
		},
	}
	wghts := []float64{10, 10, 20, 20}
	allRates := []*engine.Rate{rtHalfDay1, rtHalfDay2, rtHalfDay2r1, rtHalfDay2r2}
	for _, idx := range allRates {
		if err := idx.Compile(); err != nil {
			t.Error(err)
		}
	}
	sTime := time.Date(2020, 12, 21, 23, 0, 0, 0, time.UTC)
	usage := 25 * time.Hour
	expected := []*orderedRate{
		{
			time.Hour,
			rtHalfDay1,
		},
		{
			13 * time.Hour,
			rtHalfDay2r1,
		},
		{
			18 * time.Hour,
			rtHalfDay2,
		},
		{
			19 * time.Hour,
			rtHalfDay2r2,
		},
	}
	if ordRts, err := orderRatesOnIntervals(allRates, wghts, sTime, usage, true, 10); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(ordRts, expected) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(ordRts))
	}
}

func TestOrderRateOnIntervalsEvery1Hour30Mins(t *testing.T) {
	rateEvery1H := &engine.Rate{
		ID:              "HOUR_RATE",
		ActivationTimes: "* */1 * * *",
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
			},
		},
	}
	rateEvery30Mins := &engine.Rate{
		ID:              "MINUTES_RATE",
		ActivationTimes: "*/30 * * * *",
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
			},
		},
	}
	wghts := []float64{10, 20}
	allRates := []*engine.Rate{rateEvery1H, rateEvery30Mins}
	for _, idx := range allRates {
		if err := idx.Compile(); err != nil {
			t.Error(err)
		}
	}
	sTime := time.Date(2020, 9, 20, 10, 0, 0, 0, time.UTC)
	usage := time.Hour + time.Second
	expected := []*orderedRate{
		{
			0,
			rateEvery30Mins,
		},
		{
			time.Minute,
			rateEvery1H,
		},
		{
			30 * time.Minute,
			rateEvery30Mins,
		},
		{
			30*time.Minute + time.Minute,
			rateEvery1H,
		},
		{
			time.Hour,
			rateEvery30Mins,
		},
	}
	if ordRts, err := orderRatesOnIntervals(allRates, wghts, sTime, usage, true, 10); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(ordRts, expected) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(ordRts))
	}
}

func TestOrderRatesOnIntervalsOnePrinciapalRateCase1(t *testing.T) {
	rtPrincipal := &engine.Rate{
		ID:              "PRINCIPAL_RATE",
		ActivationTimes: "* 10-22 * * *",
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
			},
		},
	}
	rt1 := &engine.Rate{
		ID:              "RT1",
		ActivationTimes: "* 10-18 * * *",
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
			},
		},
	}
	rt2 := &engine.Rate{
		ID:              "RT2",
		ActivationTimes: "* 10-16 * * *",
		Weights: utils.DynamicWeights{
			{
				Weight: 30,
			},
		},
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
			},
		},
	}
	rt3 := &engine.Rate{
		ID:              "RT3",
		ActivationTimes: "* 10-14 * * *",
		Weights: utils.DynamicWeights{
			{
				Weight: 40,
			},
		},
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
			},
		},
	}
	wghts := []float64{10, 20, 30, 40}
	allRates := []*engine.Rate{rtPrincipal, rt1, rt2, rt3}
	for _, idx := range allRates {
		if err := idx.Compile(); err != nil {
			t.Error(err)
		}
	}
	sTime := time.Date(2020, 7, 21, 10, 0, 0, 0, time.UTC)
	usage := 13 * time.Hour
	expected := []*orderedRate{
		{
			0,
			rt3,
		},
		{
			5 * time.Hour,
			rt2,
		},
		{
			7 * time.Hour,
			rt1,
		},
		{
			9 * time.Hour,
			rtPrincipal,
		},
	}
	if ordRts, err := orderRatesOnIntervals(allRates, wghts, sTime, usage, true, 10); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(ordRts, expected) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(ordRts))
	}
}

func TestOrderRatesOnIntervalsOnePrinciapalRateCase2(t *testing.T) {
	rtPrincipal := &engine.Rate{
		ID:              "PRINCIPAL_RATE",
		ActivationTimes: "* 10-22 * * *",
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
			},
		},
	}
	rt1 := &engine.Rate{
		ID:              "RT1",
		ActivationTimes: "* 18-22 * * *",
		Weights: utils.DynamicWeights{
			{
				Weight: 40,
			},
		},
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
			},
		},
	}
	rt2 := &engine.Rate{
		ID:              "RT2",
		ActivationTimes: "* 16-22 * * *",
		Weights: utils.DynamicWeights{
			{
				Weight: 30,
			},
		},
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
			},
		},
	}
	rt3 := &engine.Rate{
		ID:              "RT3",
		ActivationTimes: "* 14-22 * * *",
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
			},
		},
	}
	wght := []float64{10, 40, 30, 20}
	allRates := []*engine.Rate{rtPrincipal, rt1, rt2, rt3}
	for _, idx := range allRates {
		if err := idx.Compile(); err != nil {
			t.Error(err)
		}
	}
	sTime := time.Date(2020, 7, 21, 10, 0, 0, 0, time.UTC)
	usage := 13 * time.Hour
	expected := []*orderedRate{
		{
			0,
			rtPrincipal,
		},
		{
			4 * time.Hour,
			rt3,
		},
		{
			6 * time.Hour,
			rt2,
		},
		{
			8 * time.Hour,
			rt1,
		},
	}
	if ordRts, err := orderRatesOnIntervals(allRates, wght, sTime, usage, true, 10); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(ordRts, expected) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(ordRts))
	}
}

func TestOrderRatesOnIntervalsEvenOddMinutes(t *testing.T) {
	rtOddMInutes := &engine.Rate{
		ID:              "ODD_RATE",
		ActivationTimes: "*/1 * * * *",
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
			},
		},
	}
	rtEvenMinutes := &engine.Rate{
		ID:              "EVEN_RATE",
		ActivationTimes: "*/2 * * * *",
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
			},
		},
	}
	wghts := []float64{10, 20}
	allRates := []*engine.Rate{rtOddMInutes, rtEvenMinutes}
	for _, idx := range allRates {
		if err := idx.Compile(); err != nil {
			t.Error(err)
		}
	}
	sTime := time.Date(2020, 12, 23, 22, 0, 0, 0, time.UTC)
	usage := 5*time.Minute + time.Second
	expected := []*orderedRate{
		{
			0,
			rtEvenMinutes,
		},
		{
			time.Minute,
			rtOddMInutes,
		},
		{
			2 * time.Minute,
			rtEvenMinutes,
		},
		{
			3 * time.Minute,
			rtOddMInutes,
		},
		{
			4 * time.Minute,
			rtEvenMinutes,
		},
		{
			5 * time.Minute,
			rtOddMInutes,
		},
	}
	if ordRts, err := orderRatesOnIntervals(allRates, wghts, sTime, usage, true, 10); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(ordRts, expected) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(ordRts))
	}
}

func TestOrderRatesOnIntervalsDoubleRates2(t *testing.T) {
	rt1 := &engine.Rate{
		ID: "ALWAYS_RATE",
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
			},
		},
	}
	rtCh1 := &engine.Rate{
		ID:              "CHRISTMAS1",
		ActivationTimes: "* * 24 12 *",
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
			},
		},
	}
	rtCh2 := &engine.Rate{
		ID:              "CHRISTMAS2",
		ActivationTimes: "* 10-12 24 12 *",
		Weights: utils.DynamicWeights{
			{
				Weight: 30,
			},
		},
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
			},
		},
	}
	rtCh3 := &engine.Rate{
		ID:              "CHRISTMAS3",
		ActivationTimes: "* 20-22 24 12 *",
		Weights: utils.DynamicWeights{
			{
				Weight: 30,
			},
		},
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
			},
		},
	}
	wghts := []float64{10, 20, 30, 30}
	allRates := []*engine.Rate{rt1, rtCh1, rtCh2, rtCh3}
	for _, idx := range allRates {
		if err := idx.Compile(); err != nil {
			t.Error(err)
		}
	}
	sTime := time.Date(2020, 12, 23, 22, 0, 0, 0, time.UTC)
	usage := 36 * time.Hour
	expected := []*orderedRate{
		{
			0,
			rt1,
		},
		{
			2 * time.Hour,
			rtCh1,
		},
		{
			12 * time.Hour,
			rtCh2,
		},
		{
			15 * time.Hour,
			rtCh1,
		},
		{
			22 * time.Hour,
			rtCh3,
		},
		{
			25 * time.Hour,
			rtCh1,
		},
		{
			26 * time.Hour,
			rt1,
		},
	}
	if ordRts, err := orderRatesOnIntervals(allRates, wghts, sTime, usage, true, 10); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(ordRts, expected) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(ordRts))
	}
}

func TestOrderOnRatesIntervalsEveryTwoHours(t *testing.T) {
	rt1 := &engine.Rate{
		ID: "ALWAYS_RATE",
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
			},
		},
	}
	rtEvTwoHours := &engine.Rate{
		ID: "EVERY_TWO_HOURS",
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
		ActivationTimes: "* */2 * * *",
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
			},
		},
	}
	wghts := []float64{10, 20}
	allRates := []*engine.Rate{rt1, rtEvTwoHours}
	for _, idx := range allRates {
		if err := idx.Compile(); err != nil {
			t.Error(err)
		}
	}
	sTime := time.Date(2020, 7, 21, 12, 10, 0, 0, time.UTC)
	usage := 4 * time.Hour
	expected := []*orderedRate{
		{
			0,
			rtEvTwoHours,
		},
		{
			50 * time.Minute,
			rt1,
		},
		{
			time.Hour + 50*time.Minute,
			rtEvTwoHours,
		},
		{
			2*time.Hour + 50*time.Minute,
			rt1,
		},
		{
			3*time.Hour + 50*time.Minute,
			rtEvTwoHours,
		},
	}
	if ordRts, err := orderRatesOnIntervals(allRates, wghts, sTime, usage, true, 10); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(ordRts, expected) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(ordRts))
	}
}

func TestOrderRatesOnIntervalsEveryTwoDays(t *testing.T) {
	rt1 := &engine.Rate{
		ID: "ALWAYS_RATE",
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
			},
		},
	}
	rtEveryTwoDays := &engine.Rate{
		ID:              "RATE_EVERY_DAY",
		ActivationTimes: "* * */2 * *",
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
			},
		},
	}
	wghts := []float64{10, 20}
	allRates := []*engine.Rate{rt1, rtEveryTwoDays}
	for _, idx := range allRates {
		if err := idx.Compile(); err != nil {
			t.Error(err)
		}
	}
	sTime := time.Date(2020, 7, 21, 23, 59, 59, 0, time.UTC)
	usage := 96*time.Hour + 2*time.Second
	expected := []*orderedRate{
		{
			0,
			rtEveryTwoDays,
		},
		{
			time.Second,
			rt1,
		},
		{
			24*time.Hour + time.Second,
			rtEveryTwoDays,
		},
		{
			48*time.Hour + time.Second,
			rt1,
		},
		{
			72*time.Hour + time.Second,
			rtEveryTwoDays,
		},
		{
			96*time.Hour + time.Second,
			rt1,
		},
	}
	if ordRts, err := orderRatesOnIntervals(allRates, wghts, sTime, usage, true, 10); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(ordRts, expected) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(ordRts))
	}
}

func TestOrderRatesOnIntervalsSpecialHour(t *testing.T) {
	rtRestricted := &engine.Rate{
		ID:              "RESTRICTED",
		ActivationTimes: "* 10-22 * * *",
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
			},
		},
	}
	rtWayRestricted := &engine.Rate{
		ID:              "WAY_RESTRICTED",
		ActivationTimes: "* 12-14 * * *",
		Weights: utils.DynamicWeights{
			{
				Weight: 30,
			},
		},
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
			},
		},
	}
	rtSpecialHour := &engine.Rate{
		ID:              "SPECIAL_HOUR",
		ActivationTimes: "* 13 * * *",
		Weights: utils.DynamicWeights{
			{
				Weight: 40,
			},
		},
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
			},
		},
	}
	wghts := []float64{20, 30, 40}
	allRts := []*engine.Rate{rtRestricted, rtWayRestricted, rtSpecialHour}
	for _, idx := range allRts {
		if err := idx.Compile(); err != nil {
			t.Error(err)
		}
	}
	sTime := time.Date(2020, 7, 21, 9, 0, 0, 0, time.UTC)
	usage := 11 * time.Hour
	expected := []*orderedRate{
		{
			time.Hour,
			rtRestricted,
		},
		{
			3 * time.Hour,
			rtWayRestricted,
		},
		{
			4 * time.Hour,
			rtSpecialHour,
		},
		{
			5 * time.Hour,
			rtWayRestricted,
		},
		{
			6 * time.Hour,
			rtRestricted,
		},
	}
	if ordRts, err := orderRatesOnIntervals(allRts, wghts, sTime, usage, true, 10); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(ordRts, expected) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(ordRts))
	}
}

func TestOrderRateIntervalsRateEveryTenMinutes(t *testing.T) {
	rt1 := &engine.Rate{
		ID:              "DAY_RATE",
		ActivationTimes: "* * 21 7 *",
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
			},
		},
	}
	rtEveryTenMin := &engine.Rate{
		ID:              "EVERY_TEN_MIN",
		ActivationTimes: "*/20 * * * *",
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
			},
		},
	}
	wghts := []float64{10, 20}
	allRts := []*engine.Rate{rt1, rtEveryTenMin}
	for _, idx := range allRts {
		if err := idx.Compile(); err != nil {
			t.Error(err)
		}
	}
	sTime := time.Date(2020, 7, 21, 10, 05, 0, 0, time.UTC)
	usage := 40 * time.Minute
	expected := []*orderedRate{
		{
			0,
			rt1,
		},
		{
			15 * time.Minute,
			rtEveryTenMin,
		},
		{
			16 * time.Minute,
			rt1,
		},
		{
			35 * time.Minute,
			rtEveryTenMin,
		},
		{
			36 * time.Minute,
			rt1,
		},
	}
	if ordRts, err := orderRatesOnIntervals(allRts, wghts, sTime, usage, true, 10); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(ordRts, expected) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(ordRts))
	}
}

func TestOrderRatesOnIntervalsDayOfTheWeek(t *testing.T) {
	rt1 := &engine.Rate{
		ID: "ALWAYS_RATE",
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
			},
		},
	}
	rtDay := &engine.Rate{
		ID:              "DAY_RATE",
		ActivationTimes: "* * 21 7 2",
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
			},
		},
	}
	rtDay1 := &engine.Rate{
		ID:              "DAY_RATE1",
		ActivationTimes: "* 15 21 7 2",
		Weights: utils.DynamicWeights{
			{
				Weight: 30,
			},
		},
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
			},
		},
	}
	rtDay2 := &engine.Rate{
		ID:              "DAY_RATE2",
		ActivationTimes: "* 18 21 7 2",
		Weights: utils.DynamicWeights{
			{
				Weight: 30,
			},
		},
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
			},
		},
	}
	wghts := []float64{10, 20, 30, 30}
	allRates := []*engine.Rate{rt1, rtDay, rtDay1, rtDay2}
	for _, idx := range allRates {
		if err := idx.Compile(); err != nil {
			t.Error(err)
		}
	}
	sTime := time.Date(2020, 7, 20, 23, 0, 0, 0, time.UTC)
	usage := 30 * time.Hour
	expected := []*orderedRate{
		{
			0,
			rt1,
		},
		{
			time.Hour,
			rtDay,
		},
		{
			16 * time.Hour,
			rtDay1,
		},
		{
			17 * time.Hour,
			rtDay,
		},
		{
			19 * time.Hour,
			rtDay2,
		},
		{
			20 * time.Hour,
			rtDay,
		},
		{
			25 * time.Hour,
			rt1,
		},
	}
	if ordRts, err := orderRatesOnIntervals(allRates, wghts, sTime, usage, true, 10); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(ordRts, expected) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(ordRts))
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
				IntervalStart: 0,
			},
		},
	}
	wghts := []float64{0}
	err := rt1.Compile()
	if err != nil {
		t.Error(err)
	}
	aRts := []*engine.Rate{rt1}
	sTime := time.Date(2020, 01, 02, 0, 1, 0, 0, time.UTC)
	usage := 96 * time.Hour
	expectedErr := "maximum iterations reached"
	if _, err := orderRatesOnIntervals(aRts, wghts, sTime, usage, false, 1); err == nil || err.Error() != expectedErr {
		t.Errorf("Expected %+v, received %+v", expectedErr, err)
	}
}

func TestOrderRatesOnIntervalIsDirectionFalse(t *testing.T) {
	rt1 := &engine.Rate{
		ID:              "RT_1",
		ActivationTimes: "* * 27 02 *",
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
			},
		},
	}
	err := rt1.Compile()
	if err != nil {
		t.Error(err)
	}
	expected := []*orderedRate{
		{
			0,
			rt1,
		},
	}
	err = expected[0].Rate.Compile()
	if err != nil {
		t.Error(err)
	}
	wghts := []float64{0}
	aRts := []*engine.Rate{rt1}
	sTime := time.Date(0001, 02, 27, 0, 0, 0, 0, time.UTC)
	usage := 48 * time.Hour
	if ordRts, err := orderRatesOnIntervals(aRts, wghts, sTime, usage, false, 5); err != nil {
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
				IntervalStart: 0,
			},
		},
	}
	err := rt1.Compile()
	if err != nil {
		t.Error(err)
	}
	expected := []*orderedRate{
		{
			0,
			rt1,
		},
	}
	err = expected[0].Rate.Compile()
	if err != nil {
		t.Error(err)
	}
	wghts := []float64{0}
	aRts := []*engine.Rate{rt1}
	sTime := time.Date(2020, 12, 1, 0, 0, 0, 0, time.UTC)
	usage := 96 * time.Hour
	if ordRts, err := orderRatesOnIntervals(aRts, wghts, sTime, usage, true, 4); err != nil {
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
	expected := []*orderedRate{
		{
			0,
			rt1,
		},
	}
	err = expected[0].Rate.Compile()
	if err != nil {
		t.Error(err)
	}
	wghts := []float64{0}
	aRts := []*engine.Rate{rt1}
	sTime := time.Date(2020, 12, 1, 0, 0, 0, 0, time.UTC)
	usage := 48 * time.Hour
	if _, err := orderRatesOnIntervals(aRts, wghts, sTime, usage, false, 4); err != nil {
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
	expected := []*orderedRate{
		{
			0,
			rt1,
		},
	}
	err = expected[0].Rate.Compile()
	if err != nil {
		t.Error(err)
	}
	aRts := []*engine.Rate{rt1}
	wghts := []float64{0}
	sTime := time.Date(2020, 12, 1, 0, 0, 0, 0, time.UTC)
	usage := 48 * time.Hour
	if ordRts, err := orderRatesOnIntervals(aRts, wghts, sTime, usage, false, 4); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(ordRts, expected) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(expected), utils.ToJSON(ordRts))
	}
}

func TestComputeRateSIntervals(t *testing.T) {
	minDecimal, err := utils.NewDecimalFromUsage("1m")
	if err != nil {
		t.Error(err)
	}
	secDecimal, err := utils.NewDecimalFromUsage("1s")
	if err != nil {
		t.Error(err)
	}
	rt0 := &engine.Rate{
		ID: "RATE0",
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
				RecurrentFee:  utils.NewDecimal(10, 1),
				Unit:          minDecimal,
				Increment:     minDecimal,
			},
			{
				IntervalStart: 60 * time.Second,
				RecurrentFee:  utils.NewDecimal(5, 3),
				Unit:          minDecimal,
				Increment:     secDecimal,
			},
		},
	}

	rt1 := &engine.Rate{
		ID: "RATE1",
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
				RecurrentFee:  utils.NewDecimal(2, 1),
				Unit:          minDecimal,
				Increment:     secDecimal,
			},
			{
				IntervalStart: 2 * time.Minute,
				RecurrentFee:  utils.NewDecimal(15, 2),
				Unit:          minDecimal,
				Increment:     secDecimal,
			},
		},
	}

	rp := &engine.RateProfile{
		Tenant: "cgrates.org",
		ID:     "RATE_PROFILE",
		Rates: map[string]*engine.Rate{
			rt0.ID: rt0,
			rt1.ID: rt1,
		},
	}
	if err := rp.Compile(); err != nil {
		t.Fatal(err)
	}

	rts := []*orderedRate{
		{
			0,
			rt0,
		},
		{
			90 * time.Second,
			rt1,
		},
	}

	eRtIvls := []*engine.RateSInterval{
		{
			UsageStart: 0,
			Increments: []*engine.RateSIncrement{
				{
					UsageStart:        0,
					Usage:             time.Minute,
					Rate:              rt0,
					IntervalRateIndex: 0,
					CompressFactor:    1,
				},
				{
					UsageStart:        time.Minute,
					Usage:             30 * time.Second,
					Rate:              rt0,
					IntervalRateIndex: 1,
					CompressFactor:    30,
				},
			},
			CompressFactor: 1,
		},
		{
			UsageStart: 90 * time.Second,
			Increments: []*engine.RateSIncrement{
				{
					UsageStart:        90 * time.Second,
					Usage:             30 * time.Second,
					Rate:              rt1,
					IntervalRateIndex: 0,
					CompressFactor:    30,
				},
				{
					UsageStart:        2 * time.Minute,
					Usage:             10 * time.Second,
					Rate:              rt1,
					IntervalRateIndex: 1,
					CompressFactor:    10,
				},
			},
			CompressFactor: 1,
		},
	}
	if rtIvls, err := computeRateSIntervals(rts,
		0, 130*time.Second); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eRtIvls, rtIvls) {
		t.Errorf("expecting: %+v \n,received: %+v", utils.ToJSON(eRtIvls), utils.ToJSON(rtIvls))
	}

	rts = []*orderedRate{
		{
			0,
			rt0,
		},
		{
			90 * time.Second,
			rt1,
		},
	}

	eRtIvls = []*engine.RateSInterval{
		{
			UsageStart: time.Minute,
			Increments: []*engine.RateSIncrement{
				{
					UsageStart:        time.Minute,
					Usage:             30 * time.Second,
					Rate:              rt0,
					IntervalRateIndex: 1,
					CompressFactor:    30,
				},
			},
			CompressFactor: 1,
		},
		{
			UsageStart: 90 * time.Second,
			Increments: []*engine.RateSIncrement{
				{
					UsageStart:        90 * time.Second,
					Usage:             30 * time.Second,
					Rate:              rt1,
					IntervalRateIndex: 0,
					CompressFactor:    30,
				},
				{
					UsageStart:        2 * time.Minute,
					Usage:             10 * time.Second,
					Rate:              rt1,
					IntervalRateIndex: 1,
					CompressFactor:    10,
				},
			},
			CompressFactor: 1,
		},
	}
	if rtIvls, err := computeRateSIntervals(rts,
		time.Minute, 70*time.Second); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eRtIvls, rtIvls) {
		t.Errorf("expecting: %+v, received: %+v", utils.ToIJSON(eRtIvls), utils.ToIJSON(rtIvls))
	}
}

func TestComputeRateSIntervals1(t *testing.T) {
	minDecimal, err := utils.NewDecimalFromUsage("1m")
	if err != nil {
		t.Error(err)
	}
	secDecimal, err := utils.NewDecimalFromUsage("1s")
	if err != nil {
		t.Error(err)
	}
	tsecDecimal, err := utils.NewDecimalFromUsage("30s")
	if err != nil {
		t.Error(err)
	}
	rt0 := &engine.Rate{
		ID: "RATE0",
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
				RecurrentFee:  utils.NewDecimal(2, 1),
				Unit:          tsecDecimal,
				Increment:     tsecDecimal,
			},
			{
				IntervalStart: 30 * time.Second,
				RecurrentFee:  utils.NewDecimal(15, 2),
				Unit:          minDecimal,
				Increment:     secDecimal,
			},
		},
	}

	rt1 := &engine.Rate{
		ID: "RATE1",
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
				RecurrentFee:  utils.NewDecimal(2, 1),
				Unit:          minDecimal,
				Increment:     secDecimal,
			},
			{
				IntervalStart: 2 * time.Minute,
				RecurrentFee:  utils.NewDecimal(15, 2),
				Unit:          minDecimal,
				Increment:     secDecimal,
			},
		},
	}

	rp := &engine.RateProfile{
		Tenant: "cgrates.org",
		ID:     "RATE_PROFILE",
		Rates: map[string]*engine.Rate{
			rt0.ID: rt0,
			rt1.ID: rt1,
		},
	}
	if err := rp.Compile(); err != nil {
		t.Fatal(err)
	}

	ordRts := []*orderedRate{
		{
			0,
			rt0,
		},
		{
			time.Minute + 10*time.Second,
			rt1,
		},
	}

	eRtIvls := []*engine.RateSInterval{
		{
			UsageStart: 30 * time.Second,
			Increments: []*engine.RateSIncrement{
				{
					UsageStart:        30 * time.Second,
					Usage:             40 * time.Second,
					Rate:              rt0,
					IntervalRateIndex: 1,
					CompressFactor:    40,
				},
			},
			CompressFactor: 1,
		},
		{
			UsageStart: time.Minute + 10*time.Second,
			Increments: []*engine.RateSIncrement{
				{
					UsageStart:        time.Minute + 10*time.Second,
					Usage:             50 * time.Second,
					Rate:              rt1,
					IntervalRateIndex: 0,
					CompressFactor:    50,
				},
				{
					UsageStart:        2 * time.Minute,
					Usage:             90 * time.Second,
					Rate:              rt1,
					IntervalRateIndex: 1,
					CompressFactor:    90,
				},
			},
			CompressFactor: 1,
		},
	}
	if rtIvls, err := computeRateSIntervals(ordRts, 30*time.Second, 3*time.Minute); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rtIvls, eRtIvls) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(eRtIvls), utils.ToJSON(rtIvls))
	}
}

func TestComputeRateSIntervalsWIthFixedFee(t *testing.T) {
	minDecimal, err := utils.NewDecimalFromUsage("1m")
	if err != nil {
		t.Error(err)
	}
	secDecimal, err := utils.NewDecimalFromUsage("1s")
	if err != nil {
		t.Error(err)
	}
	tsecDecimal, err := utils.NewDecimalFromUsage("30s")
	if err != nil {
		t.Error(err)
	}
	rt0 := &engine.Rate{
		ID: "RATE0",
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
				FixedFee:      utils.NewDecimal(123, 3),
				RecurrentFee:  utils.NewDecimal(2, 1),
				Unit:          tsecDecimal,
				Increment:     tsecDecimal,
			},
			{
				IntervalStart: 30 * time.Second,
				RecurrentFee:  utils.NewDecimal(15, 2),
				Unit:          minDecimal,
				Increment:     secDecimal,
			},
		},
	}

	rt1 := &engine.Rate{
		ID: "RATE1",
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
				FixedFee:      utils.NewDecimal(567, 3),
				RecurrentFee:  utils.NewDecimal(2, 1),
				Unit:          minDecimal,
				Increment:     secDecimal,
			},
			{
				IntervalStart: 2 * time.Minute,
				RecurrentFee:  utils.NewDecimal(15, 2),
				Unit:          minDecimal,
				Increment:     secDecimal,
			},
		},
	}

	rp := &engine.RateProfile{
		Tenant: "cgrates.org",
		ID:     "RATE_PROFILE",
		Rates: map[string]*engine.Rate{
			rt0.ID: rt0,
			rt1.ID: rt1,
		},
	}
	if err := rp.Compile(); err != nil {
		t.Fatal(err)
	}

	ordRts := []*orderedRate{
		{
			0,
			rt0,
		},
		{
			time.Minute + 10*time.Second,
			rt1,
		},
	}

	eRtIvls := []*engine.RateSInterval{
		{
			UsageStart: 0,
			Increments: []*engine.RateSIncrement{
				{
					UsageStart:        0,
					Rate:              rt0,
					IntervalRateIndex: 0,
					CompressFactor:    1,
					Usage:             -1,
				},
				{
					UsageStart:        0,
					Rate:              rt0,
					IntervalRateIndex: 0,
					CompressFactor:    1,
					Usage:             30 * time.Second,
				},
				{
					UsageStart:        30 * time.Second,
					Rate:              rt0,
					IntervalRateIndex: 1,
					CompressFactor:    40,
					Usage:             40 * time.Second,
				},
			},
			CompressFactor: 1,
		},
		{
			UsageStart: time.Minute + 10*time.Second,
			Increments: []*engine.RateSIncrement{
				{
					UsageStart:        time.Minute + 10*time.Second,
					Rate:              rt1,
					IntervalRateIndex: 0,
					CompressFactor:    1,
					Usage:             -1,
				},
				{
					UsageStart:        time.Minute + 10*time.Second,
					Rate:              rt1,
					IntervalRateIndex: 0,
					CompressFactor:    50,
					Usage:             50 * time.Second,
				},
				{
					UsageStart:        2 * time.Minute,
					Rate:              rt1,
					IntervalRateIndex: 1,
					CompressFactor:    60,
					Usage:             60 * time.Second,
				},
			},
			CompressFactor: 1,
		},
	}
	if rtIvls, err := computeRateSIntervals(ordRts, 0, 3*time.Minute); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rtIvls, eRtIvls) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(eRtIvls), utils.ToJSON(rtIvls))
	}
}

func TestComputeRateSIntervals2(t *testing.T) {
	minDecimal, err := utils.NewDecimalFromUsage("1m")
	if err != nil {
		t.Error(err)
	}
	rt0 := &engine.Rate{
		ID:              "RATE0",
		ActivationTimes: "* * * * *",
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
				RecurrentFee:  utils.NewDecimal(1, 0),
				Unit:          minDecimal,
				Increment:     minDecimal,
			},
			{
				IntervalStart: 50 * time.Minute,
				RecurrentFee:  utils.NewDecimal(5, 1),
				Unit:          minDecimal,
				Increment:     minDecimal,
			},
		},
	}
	rt0.Compile()

	rt1 := &engine.Rate{
		ID:              "RATE1",
		ActivationTimes: "45-49 * * * *",
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
				RecurrentFee:  utils.NewDecimal(15, 2),
				Unit:          minDecimal,
				Increment:     minDecimal,
			},
			{
				IntervalStart: 45 * time.Minute,
				RecurrentFee:  utils.NewDecimal(2, 1),
				Unit:          minDecimal,
				Increment:     minDecimal,
			},
		},
	}
	rt1.Compile()
	wghts := []float64{0, 0}
	allRates := []*engine.Rate{rt0, rt1}

	ordRts := []*orderedRate{
		{
			0,
			rt0,
		},
		{
			45 * time.Minute,
			rt1,
		},
		{
			50 * time.Minute,
			rt0,
		},
	}

	eRtIvls := []*engine.RateSInterval{
		{
			UsageStart: 0,
			Increments: []*engine.RateSIncrement{
				{
					UsageStart:        0,
					Rate:              rt0,
					IntervalRateIndex: 0,
					CompressFactor:    45,
					Usage:             45 * time.Minute,
				},
			},
			CompressFactor: 1,
		},
		{
			UsageStart: 45 * time.Minute,
			Increments: []*engine.RateSIncrement{
				{
					UsageStart:        45 * time.Minute,
					Rate:              rt1,
					IntervalRateIndex: 1,
					CompressFactor:    5,
					Usage:             5 * time.Minute,
				},
			},
			CompressFactor: 1,
		},
		{
			UsageStart: 50 * time.Minute,
			Increments: []*engine.RateSIncrement{
				{
					UsageStart:        50 * time.Minute,
					Rate:              rt0,
					IntervalRateIndex: 1,
					CompressFactor:    10,
					Usage:             10 * time.Minute,
				},
			},
			CompressFactor: 1,
		},
	}
	sTime := time.Date(2020, 7, 21, 0, 0, 0, 0, time.UTC)
	usage := time.Hour
	if rcvOrdRts, err := orderRatesOnIntervals(allRates, wghts, sTime, usage, true, 10); err != nil {
		t.Error(eRtIvls)
	} else if !reflect.DeepEqual(ordRts, rcvOrdRts) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(ordRts), utils.ToJSON(rcvOrdRts))
	} else if rcveRtIvls, err := computeRateSIntervals(rcvOrdRts, 0, usage); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcveRtIvls, eRtIvls) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(eRtIvls), utils.ToJSON(rcveRtIvls))
	}
}

func TestComputeRateSIntervalsEvery30Seconds(t *testing.T) {
	tsecDecimal, err := utils.NewDecimalFromUsage("30s")
	if err != nil {
		t.Error(err)
	}
	secDecimal, err := utils.NewDecimalFromUsage("1s")
	if err != nil {
		t.Error(err)
	}
	rt1 := &engine.Rate{
		ID: "RATE1",
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
				RecurrentFee:  utils.NewDecimal(1, 1),
				Unit:          tsecDecimal,
				Increment:     secDecimal,
			},
			{
				IntervalStart: time.Minute,
				RecurrentFee:  utils.NewDecimal(2, 1),
				Unit:          tsecDecimal,
				Increment:     secDecimal,
			},
			{
				IntervalStart: 2 * time.Minute,
				RecurrentFee:  utils.NewDecimal(3, 1),
				Unit:          tsecDecimal,
				Increment:     secDecimal,
			},
		},
	}

	rt2 := &engine.Rate{
		ID: "RATE2",
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 30 * time.Second,
				RecurrentFee:  utils.NewDecimal(15, 2),
				Unit:          tsecDecimal,
				Increment:     secDecimal,
			},
			{
				IntervalStart: time.Minute + 30*time.Second,
				RecurrentFee:  utils.NewDecimal(25, 2),
				Unit:          tsecDecimal,
				Increment:     secDecimal,
			},
			{
				IntervalStart: 2*time.Minute + 30*time.Second,
				RecurrentFee:  utils.NewDecimal(35, 2),
				Unit:          tsecDecimal,
				Increment:     secDecimal,
			},
		},
	}

	rp := &engine.RateProfile{
		Tenant: "cgrates.org",
		ID:     "RATE_PROFILE",
		Rates: map[string]*engine.Rate{
			rt1.ID: rt1,
			rt2.ID: rt2,
		},
	}
	if err := rp.Compile(); err != nil {
		t.Fatal(err)
	}

	ordRts := []*orderedRate{
		{
			0,
			rt1,
		},
		{
			30 * time.Second,
			rt2,
		},
		{
			time.Minute,
			rt1,
		},
		{
			time.Minute + 30*time.Second,
			rt2,
		},
		{
			2 * time.Minute,
			rt1,
		},
		{
			2*time.Minute + 30*time.Second,
			rt2,
		},
	}

	expOrdRates := []*engine.RateSInterval{
		{
			UsageStart: 0,
			Increments: []*engine.RateSIncrement{
				{
					UsageStart:        0,
					Rate:              rt1,
					IntervalRateIndex: 0,
					CompressFactor:    30,
					Usage:             30 * time.Second,
				},
			},
			CompressFactor: 1,
		},
		{
			UsageStart: 30 * time.Second,
			Increments: []*engine.RateSIncrement{
				{
					UsageStart:        30 * time.Second,
					Rate:              rt2,
					IntervalRateIndex: 0,
					CompressFactor:    30,
					Usage:             30 * time.Second,
				},
			},
			CompressFactor: 1,
		},
		{
			UsageStart: time.Minute,
			Increments: []*engine.RateSIncrement{
				{
					UsageStart:        time.Minute,
					Rate:              rt1,
					IntervalRateIndex: 1,
					CompressFactor:    30,
					Usage:             30 * time.Second,
				},
			},
			CompressFactor: 1,
		},
		{
			UsageStart: time.Minute + 30*time.Second,
			Increments: []*engine.RateSIncrement{
				{
					UsageStart:        time.Minute + 30*time.Second,
					Rate:              rt2,
					IntervalRateIndex: 1,
					CompressFactor:    30,
					Usage:             30 * time.Second,
				},
			},
			CompressFactor: 1,
		},
		{
			UsageStart: 2 * time.Minute,
			Increments: []*engine.RateSIncrement{
				{
					UsageStart:        2 * time.Minute,
					Rate:              rt1,
					IntervalRateIndex: 2,
					CompressFactor:    30,
					Usage:             30 * time.Second,
				},
			},
			CompressFactor: 1,
		},
		{
			UsageStart: 2*time.Minute + 30*time.Second,
			Increments: []*engine.RateSIncrement{
				{
					UsageStart:        2*time.Minute + 30*time.Second,
					Rate:              rt2,
					IntervalRateIndex: 2,
					CompressFactor:    30,
					Usage:             30 * time.Second,
				},
			},
			CompressFactor: 1,
		},
	}
	if rcvOrdRates, err := computeRateSIntervals(ordRts, 0, 3*time.Minute); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcvOrdRates, expOrdRates) {
		t.Errorf("Expected %+v,\nreceived %+v", utils.ToJSON(expOrdRates), utils.ToJSON(rcvOrdRates))
	}
}

func TestComputeRateSIntervalsStartHigherThanUsage(t *testing.T) {
	tsecDecimal, err := utils.NewDecimalFromUsage("30s")
	if err != nil {
		t.Error(err)
	}
	rt1 := &engine.Rate{
		ID: "RATE1",
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: time.Minute,
				RecurrentFee:  utils.NewDecimal(1, 1),
				Unit:          tsecDecimal,
				Increment:     tsecDecimal,
			},
		},
	}

	rt2 := &engine.Rate{
		ID: "RATE2",
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 2 * time.Minute,
				RecurrentFee:  utils.NewDecimal(2, 1),
				Unit:          tsecDecimal,
				Increment:     tsecDecimal,
			},
		},
	}

	rp := &engine.RateProfile{
		Tenant: "cgrates.org",
		ID:     "RATE_PROFILE",
		Rates: map[string]*engine.Rate{
			rt1.ID: rt1,
			rt2.ID: rt2,
		},
	}
	if err := rp.Compile(); err != nil {
		t.Fatal(err)
	}

	ordRts := []*orderedRate{
		{
			0,
			rt1,
		},
	}

	expected := "intervalStart for rate: <cgrates.org:RATE_PROFILE:RATE1> higher than usage: 0s"
	if _, err := computeRateSIntervals(ordRts, 0, 3*time.Minute); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, \nreceived %+q", expected, err)
	}
}

func TestComputeRateSIntervalsZeroIncrement(t *testing.T) {
	tsecDecimal, err := utils.NewDecimalFromUsage("30s")
	if err != nil {
		t.Error(err)
	}
	zeroDecimal, err := utils.NewDecimalFromUsage("0s")
	if err != nil {
		t.Error(err)
	}
	rt1 := &engine.Rate{
		ID: "RATE1",
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
				RecurrentFee:  utils.NewDecimal(1, 1),
				Unit:          tsecDecimal,
				Increment:     zeroDecimal,
			},
		},
	}
	rt1.Compile()

	ordRts := []*orderedRate{
		{
			0 * time.Second,
			rt1,
		},
	}

	expected := "zero increment to be charged within rate: <>"
	if _, err := computeRateSIntervals(ordRts, 33*time.Second, 3*time.Minute); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, \nreceived %+q", expected, err)
	}
}

func TestComputeRateSIntervalsCeilingCmpFactor(t *testing.T) {
	tsecDecimal, err := utils.NewDecimalFromUsage("30s")
	if err != nil {
		t.Error(err)
	}
	secDecimal, err := utils.NewDecimalFromUsage("1s")
	if err != nil {
		t.Error(err)
	}
	ssecDecimal, err := utils.NewDecimalFromUsage("7s")
	if err != nil {
		t.Error(err)
	}
	minDecimal, err := utils.NewDecimalFromUsage("1m")
	if err != nil {
		t.Error(err)
	}
	rt1 := &engine.Rate{
		ID: "RATE1",
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
				RecurrentFee:  utils.NewDecimal(1, 1),
				Unit:          tsecDecimal,
				Increment:     secDecimal,
			},
			{
				IntervalStart: 30 * time.Second,
				RecurrentFee:  utils.NewDecimal(25, 3),
				Unit:          minDecimal,
				Increment:     ssecDecimal,
			},
		},
	}
	rt1.Compile()

	ordRts := []*orderedRate{
		{
			0,
			rt1,
		},
	}
	expOrdRts := []*engine.RateSInterval{
		{
			UsageStart: 0,
			Increments: []*engine.RateSIncrement{
				{
					UsageStart:        0,
					Rate:              rt1,
					IntervalRateIndex: 0,
					Usage:             30 * time.Second,
					CompressFactor:    30,
				},
				{
					UsageStart:        30 * time.Second,
					Rate:              rt1,
					IntervalRateIndex: 1,
					Usage:             40 * time.Second,
					CompressFactor:    6,
				},
			},
			CompressFactor: 1,
		},
	}
	if rcvOrdRts, err := computeRateSIntervals(ordRts, 0, time.Minute+10*time.Second); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcvOrdRts, expOrdRts) {
		t.Errorf("Expected %+v, \nreceived %+v", utils.ToJSON(expOrdRts), utils.ToJSON(rcvOrdRts))
	}
}

func TestComputeRateSIntervalsSwitchingRates(t *testing.T) {
	tsecDecimal, err := utils.NewDecimalFromUsage("30s")
	if err != nil {
		t.Error(err)
	}
	secDecimal, err := utils.NewDecimalFromUsage("1s")
	if err != nil {
		t.Error(err)
	}
	ssecDecimal, err := utils.NewDecimalFromUsage("7s")
	if err != nil {
		t.Error(err)
	}
	fsecDecimal, err := utils.NewDecimalFromUsage("5s")
	if err != nil {
		t.Error(err)
	}
	fssecDecimal, err := utils.NewDecimalFromUsage("25s")
	if err != nil {
		t.Error(err)
	}
	minDecimal, err := utils.NewDecimalFromUsage("1m")
	if err != nil {
		t.Error(err)
	}
	rt1 := &engine.Rate{
		ID: "RATE1",
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
				RecurrentFee:  utils.NewDecimal(1, 1),
				Unit:          tsecDecimal,
				Increment:     secDecimal,
			},
			{
				IntervalStart: 25 * time.Second,
				RecurrentFee:  utils.NewDecimal(25, 2),
				Unit:          minDecimal,
				Increment:     ssecDecimal,
			},
		},
	}

	rt2 := &engine.Rate{
		ID: "RATE2",
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 45 * time.Second,
				RecurrentFee:  utils.NewDecimal(15, 2),
				Unit:          tsecDecimal,
				Increment:     secDecimal,
			},
			{
				IntervalStart: 55 * time.Second,
				RecurrentFee:  utils.NewDecimal(3, 1),
				Unit:          tsecDecimal,
				Increment:     fsecDecimal,
			},
		},
	}

	rt3 := &engine.Rate{
		ID: "RATE3",
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 30 * time.Second,
				RecurrentFee:  utils.NewDecimal(1, 1),
				Unit:          fssecDecimal,
				Increment:     fsecDecimal,
			},
			{
				IntervalStart: time.Minute,
				RecurrentFee:  utils.NewDecimal(5, 3),
				Unit:          tsecDecimal,
				Increment:     fsecDecimal,
			},
		},
	}

	rp := &engine.RateProfile{
		Tenant: "cgrates.org",
		ID:     "RATE_PROFILE",
		Rates: map[string]*engine.Rate{
			rt1.ID: rt1,
			rt2.ID: rt2,
			rt3.ID: rt3,
		},
	}
	if err := rp.Compile(); err != nil {
		t.Fatal(err)
	}

	ordRts := []*orderedRate{
		{
			0,
			rt1,
		},
		{
			35 * time.Second,
			rt3,
		},
		{
			46 * time.Second,
			rt2,
		},
		{
			time.Minute,
			rt3,
		},
	}

	expOrdRts := []*engine.RateSInterval{
		{
			UsageStart: 0,
			Increments: []*engine.RateSIncrement{
				{
					UsageStart:        0,
					Rate:              rt1,
					IntervalRateIndex: 0,
					Usage:             25 * time.Second,
					CompressFactor:    25,
				},
				{
					UsageStart:        25 * time.Second,
					Rate:              rt1,
					IntervalRateIndex: 1,
					Usage:             10 * time.Second,
					CompressFactor:    2,
				},
			},
			CompressFactor: 1,
		},
		{
			UsageStart: 35 * time.Second,
			Increments: []*engine.RateSIncrement{
				{
					UsageStart:        35 * time.Second,
					Rate:              rt3,
					IntervalRateIndex: 0,
					Usage:             11 * time.Second,
					CompressFactor:    3,
				},
			},
			CompressFactor: 1,
		},
		{
			UsageStart: 46 * time.Second,
			Increments: []*engine.RateSIncrement{
				{
					UsageStart:        46 * time.Second,
					Rate:              rt2,
					IntervalRateIndex: 0,
					Usage:             9 * time.Second,
					CompressFactor:    9,
				},
				{
					UsageStart:        55 * time.Second,
					Rate:              rt2,
					IntervalRateIndex: 1,
					Usage:             5 * time.Second,
					CompressFactor:    1,
				},
			},
			CompressFactor: 1,
		},
		{
			UsageStart: time.Minute,
			Increments: []*engine.RateSIncrement{
				{
					UsageStart:        time.Minute,
					Rate:              rt3,
					IntervalRateIndex: 1,
					Usage:             10 * time.Second,
					CompressFactor:    2,
				},
			},
			CompressFactor: 1,
		},
	}

	if rcvOrdRts, err := computeRateSIntervals(ordRts, 0, time.Minute+10*time.Second); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcvOrdRts, expOrdRts) {
		t.Errorf("Expected %+v, \nreceived %+v", utils.ToJSON(expOrdRts), utils.ToJSON(rcvOrdRts))
	}
}

func TestComputeRatesIntervalsAllInOne(t *testing.T) {
	tsecDecimal, err := utils.NewDecimalFromUsage("30s")
	if err != nil {
		t.Error(err)
	}
	secDecimal, err := utils.NewDecimalFromUsage("1s")
	if err != nil {
		t.Error(err)
	}
	ssecDecimal, err := utils.NewDecimalFromUsage("7s")
	if err != nil {
		t.Error(err)
	}
	minDecimal, err := utils.NewDecimalFromUsage("1m")
	if err != nil {
		t.Error(err)
	}
	rt1 := &engine.Rate{
		ID: "RATE1",
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: time.Minute,
				RecurrentFee:  utils.NewDecimal(3, 1),
				Unit:          tsecDecimal,
				Increment:     secDecimal,
			},
			{
				IntervalStart: 3 * time.Minute,
				RecurrentFee:  utils.NewDecimal(3, 1),
				Unit:          minDecimal,
				Increment:     ssecDecimal,
			},
		},
	}

	rt2 := &engine.Rate{
		ID: "RATE2",
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: time.Minute + 30*time.Second,
				RecurrentFee:  utils.NewDecimal(2, 1),
				Unit:          tsecDecimal,
				Increment:     secDecimal,
			},
			{
				IntervalStart: 2*time.Minute + 30*time.Second,
				RecurrentFee:  utils.NewDecimal(2, 1),
				Unit:          tsecDecimal,
				Increment:     ssecDecimal,
			},
		},
	}

	rt3 := &engine.Rate{
		ID: "RATE3",
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 2 * time.Minute,
				RecurrentFee:  utils.NewDecimal(1, 1),
				Unit:          tsecDecimal,
				Increment:     tsecDecimal,
			},
		},
	}

	rp := &engine.RateProfile{
		Tenant: "cgrates.org",
		ID:     "RATE_PROFILE",
		Rates: map[string]*engine.Rate{
			rt1.ID: rt1,
			rt2.ID: rt2,
			rt3.ID: rt3,
		},
	}
	if err := rp.Compile(); err != nil {
		t.Fatal(err)
	}

	ordRates := []*orderedRate{
		{
			0,
			rt1,
		},
		{
			time.Minute,
			rt1,
		},
		{
			time.Minute + 30*time.Second,
			rt2,
		},
		{
			2 * time.Minute,
			rt3,
		},
		{
			2*time.Minute + 30*time.Second,
			rt2,
		},
		{
			3 * time.Minute,
			rt1,
		},
	}
	expOrdRts := []*engine.RateSInterval{
		{
			UsageStart: time.Minute,
			Increments: []*engine.RateSIncrement{
				{
					UsageStart:        time.Minute,
					Rate:              rt1,
					IntervalRateIndex: 0,
					Usage:             30 * time.Second,
					CompressFactor:    30,
				},
			},
			CompressFactor: 1,
		},
		{
			UsageStart: time.Minute + 30*time.Second,
			Increments: []*engine.RateSIncrement{
				{
					UsageStart:        time.Minute + 30*time.Second,
					Rate:              rt2,
					IntervalRateIndex: 0,
					Usage:             30 * time.Second,
					CompressFactor:    30,
				},
			},
			CompressFactor: 1,
		},
		{
			UsageStart: 2 * time.Minute,
			Increments: []*engine.RateSIncrement{
				{
					UsageStart:        2 * time.Minute,
					Rate:              rt3,
					IntervalRateIndex: 0,
					Usage:             30 * time.Second,
					CompressFactor:    1,
				},
			},
			CompressFactor: 1,
		},
		{
			UsageStart: 2*time.Minute + 30*time.Second,
			Increments: []*engine.RateSIncrement{
				{
					UsageStart:        2*time.Minute + 30*time.Second,
					Rate:              rt2,
					IntervalRateIndex: 1,
					Usage:             30 * time.Second,
					CompressFactor:    5,
				},
			},
			CompressFactor: 1,
		},
		{
			UsageStart: 3 * time.Minute,
			Increments: []*engine.RateSIncrement{
				{
					UsageStart:        3 * time.Minute,
					Rate:              rt1,
					IntervalRateIndex: 1,
					Usage:             4 * time.Minute,
					CompressFactor:    35,
				},
			},
			CompressFactor: 1,
		},
	}

	if rcvOrdRts, err := computeRateSIntervals(ordRates, time.Minute, 6*time.Minute); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcvOrdRts, expOrdRts) {
		t.Errorf("Expected %+v, \nreceived %+v", utils.ToJSON(expOrdRts), utils.ToJSON(rcvOrdRts))
	}
}

func TestOrderRatesIntervalsFullDay(t *testing.T) {
	hourDecimal, err := utils.NewDecimalFromUsage("1h")
	if err != nil {
		t.Error(err)
	}
	tminDecimal, err := utils.NewDecimalFromUsage("3m")
	if err != nil {
		t.Error(err)
	}
	rt1 := &engine.Rate{
		ID: "RATE1",
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
				RecurrentFee:  utils.NewDecimal(5, 1),
				Unit:          hourDecimal,
				Increment:     tminDecimal,
			},
			{
				IntervalStart: 4 * time.Hour,
				RecurrentFee:  utils.NewDecimal(35, 2),
				Unit:          hourDecimal,
				Increment:     tminDecimal,
			},
		},
	}

	rt2 := &engine.Rate{
		ID: "RATE2",
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 13 * time.Hour,
				RecurrentFee:  utils.NewDecimal(4, 1),
				Unit:          hourDecimal,
				Increment:     tminDecimal,
			},
			{
				IntervalStart: 16 * time.Hour,
				RecurrentFee:  utils.NewDecimal(35, 2),
				Unit:          hourDecimal,
				Increment:     tminDecimal,
			},
		},
	}
	thourDecimal, err := utils.NewDecimalFromUsage("30h")
	if err != nil {
		t.Error(err)
	}
	dminDecimal, err := utils.NewDecimalFromUsage("2m")
	if err != nil {
		t.Error(err)
	}
	fminDecimal, err := utils.NewDecimalFromUsage("5m")
	if err != nil {
		t.Error(err)
	}
	rtGH := &engine.Rate{
		ID: "RATE_GOLDEN_HOUR",
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 12 * time.Hour,
				RecurrentFee:  utils.NewDecimal(2, 1),
				Unit:          thourDecimal,
				Increment:     dminDecimal,
			},
			{
				IntervalStart: 12*time.Hour + 30*time.Minute,
				RecurrentFee:  utils.NewDecimal(1, 1),
				Unit:          hourDecimal,
				Increment:     fminDecimal,
			},
		},
	}

	rp := &engine.RateProfile{
		Tenant: "cgrates.org",
		ID:     "RATE_PROFIOLE",
		Rates: map[string]*engine.Rate{
			rt1.ID:  rt1,
			rt2.ID:  rt2,
			rtGH.ID: rtGH,
		},
	}
	if err := rp.Compile(); err != nil {
		t.Fatal(err)
	}

	ordRts := []*orderedRate{
		{
			0,
			rt1,
		},
		{
			6 * time.Hour,
			rt1,
		},
		{
			12 * time.Hour,
			rtGH,
		},
		{
			13 * time.Hour,
			rt2,
		},
	}

	expOrdRts := []*engine.RateSInterval{
		{
			UsageStart: 0,
			Increments: []*engine.RateSIncrement{
				{
					UsageStart:        0,
					Rate:              rt1,
					IntervalRateIndex: 0,
					Usage:             4 * time.Hour,
					CompressFactor:    80,
				},
				{
					UsageStart:        4 * time.Hour,
					Rate:              rt1,
					IntervalRateIndex: 1,
					Usage:             2 * time.Hour,
					CompressFactor:    40,
				},
			},
			CompressFactor: 1,
		},
		{
			UsageStart: 6 * time.Hour,
			Increments: []*engine.RateSIncrement{
				{
					UsageStart:        6 * time.Hour,
					Rate:              rt1,
					IntervalRateIndex: 1,
					Usage:             6 * time.Hour,
					CompressFactor:    120,
				},
			},
			CompressFactor: 1,
		},
		{
			UsageStart: 12 * time.Hour,
			Increments: []*engine.RateSIncrement{
				{
					UsageStart:        12 * time.Hour,
					Rate:              rtGH,
					IntervalRateIndex: 0,
					Usage:             30 * time.Minute,
					CompressFactor:    15,
				},
				{
					UsageStart:        12*time.Hour + 30*time.Minute,
					Rate:              rtGH,
					IntervalRateIndex: 1,
					Usage:             30 * time.Minute,
					CompressFactor:    6,
				},
			},
			CompressFactor: 1,
		},
		{
			UsageStart: 13 * time.Hour,
			Increments: []*engine.RateSIncrement{
				{
					UsageStart:        13 * time.Hour,
					Rate:              rt2,
					IntervalRateIndex: 0,
					Usage:             3 * time.Hour,
					CompressFactor:    60,
				},
				{
					UsageStart:        16 * time.Hour,
					Rate:              rt2,
					IntervalRateIndex: 1,
					Usage:             9 * time.Hour,
					CompressFactor:    180,
				},
			},
			CompressFactor: 1,
		},
	}

	if rcvOrdRts, err := computeRateSIntervals(ordRts, 0, 25*time.Hour); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcvOrdRts, expOrdRts) {
		t.Errorf("Expected %+v, \nreceived %+v", utils.ToJSON(expOrdRts), utils.ToJSON(rcvOrdRts))
	}
}

func TestComputeRatesIntervalsEveryTwoSeconds(t *testing.T) {
	tsecDecimal, err := utils.NewDecimalFromUsage("10s")
	if err != nil {
		t.Error(err)
	}
	secDecimal, err := utils.NewDecimalFromUsage("1s")
	if err != nil {
		t.Error(err)
	}
	twsecDecimal, err := utils.NewDecimalFromUsage("2s")
	if err != nil {
		t.Error(err)
	}
	rt1 := &engine.Rate{
		ID: "RATE1",
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
				RecurrentFee:  utils.NewDecimal(5, 1),
				Unit:          tsecDecimal,
				Increment:     twsecDecimal,
			},
			{
				IntervalStart: 4 * time.Second,
				RecurrentFee:  utils.NewDecimal(48, 2),
				Unit:          tsecDecimal,
				Increment:     twsecDecimal,
			},
			{
				IntervalStart: 8 * time.Second,
				RecurrentFee:  utils.NewDecimal(45, 2),
				Unit:          tsecDecimal,
				Increment:     twsecDecimal,
			},
		},
	}

	rt2 := &engine.Rate{
		ID: "RATE2",
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
				RecurrentFee:  utils.NewDecimal(5, 1),
				Unit:          twsecDecimal,
				Increment:     secDecimal,
			},
			{
				IntervalStart: 2 * time.Second,
				RecurrentFee:  utils.NewDecimal(48, 2),
				Unit:          twsecDecimal,
				Increment:     secDecimal,
			},
			{
				IntervalStart: 6 * time.Second,
				RecurrentFee:  utils.NewDecimal(45, 2),
				Unit:          twsecDecimal,
				Increment:     secDecimal,
			},
		},
	}

	rp := &engine.RateProfile{
		Tenant: "cgrates.org",
		ID:     "RATE_PROFILE",
		Rates: map[string]*engine.Rate{
			rt1.ID: rt1,
			rt2.ID: rt2,
		},
	}
	if err := rp.Compile(); err != nil {
		t.Fatal(err)
	}

	ordRts := []*orderedRate{
		{
			0,
			rt1,
		},
		{
			time.Second,
			rt2,
		},
		{
			2 * time.Second,
			rt1,
		},
		{
			3 * time.Second,
			rt2,
		},
		{
			5 * time.Second,
			rt1,
		},
		{
			7 * time.Second,
			rt2,
		},
	}

	expOrdRts := []*engine.RateSInterval{
		{
			UsageStart: 0,
			Increments: []*engine.RateSIncrement{
				{
					UsageStart:        0,
					Rate:              rt1,
					IntervalRateIndex: 0,
					CompressFactor:    1,
					Usage:             time.Second,
				},
			},
			CompressFactor: 1,
		},
		{
			UsageStart: time.Second,
			Increments: []*engine.RateSIncrement{
				{
					UsageStart:        time.Second,
					Rate:              rt2,
					IntervalRateIndex: 0,
					CompressFactor:    1,
					Usage:             time.Second,
				},
			},
			CompressFactor: 1,
		},
		{
			UsageStart: 2 * time.Second,
			Increments: []*engine.RateSIncrement{
				{
					UsageStart:        2 * time.Second,
					Rate:              rt1,
					IntervalRateIndex: 0,
					CompressFactor:    1,
					Usage:             time.Second,
				},
			},
			CompressFactor: 1,
		},
		{
			UsageStart: 3 * time.Second,
			Increments: []*engine.RateSIncrement{
				{
					UsageStart:        3 * time.Second,
					Rate:              rt2,
					IntervalRateIndex: 1,
					CompressFactor:    2,
					Usage:             2 * time.Second,
				},
			},
			CompressFactor: 1,
		},
		{
			UsageStart: 5 * time.Second,
			Increments: []*engine.RateSIncrement{
				{
					UsageStart:        5 * time.Second,
					Rate:              rt1,
					IntervalRateIndex: 1,
					CompressFactor:    1,
					Usage:             2 * time.Second,
				},
			},
			CompressFactor: 1,
		},
		{
			UsageStart: 7 * time.Second,
			Increments: []*engine.RateSIncrement{
				{
					UsageStart:        7 * time.Second,
					Rate:              rt2,
					IntervalRateIndex: 2,
					CompressFactor:    3,
					Usage:             3 * time.Second,
				},
			},
			CompressFactor: 1,
		},
	}

	if rcvOrdRts, err := computeRateSIntervals(ordRts, 0, 10*time.Second); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcvOrdRts, expOrdRts) {
		t.Errorf("Expected %+v, \nreceived %+v", utils.ToJSON(expOrdRts), utils.ToJSON(rcvOrdRts))
	}
}

func TestComputeRateSIntervalsOneHourRate(t *testing.T) {
	minDecimal, err := utils.NewDecimalFromUsage("15m")
	if err != nil {
		t.Error(err)
	}
	fminDeminal, err := utils.NewDecimalFromUsage("5m")
	if err != nil {
		t.Error(err)
	}
	cDecimal, err := utils.NewDecimalFromUsage("1m30s")
	if err != nil {
		t.Error(err)
	}
	rt1 := &engine.Rate{
		ID: "RATE1",
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
				RecurrentFee:  utils.NewDecimal(5, 3),
				Unit:          minDecimal,
				Increment:     minDecimal,
			},
			{
				IntervalStart: time.Hour,
				RecurrentFee:  utils.NewDecimal(5, 1),
				Unit:          fminDeminal,
				Increment:     cDecimal,
			},
		},
	}
	tminDeminal, err := utils.NewDecimalFromUsage("10m")
	if err != nil {
		t.Error(err)
	}
	ominDeminal, err := utils.NewDecimalFromUsage("1m")
	if err != nil {
		t.Error(err)
	}
	rt2 := &engine.Rate{
		ID: "RATE2",
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 15 * time.Minute,
				RecurrentFee:  utils.NewDecimal(1, 1),
				Unit:          tminDeminal,
				Increment:     ominDeminal,
			},
			{
				IntervalStart: 30 * time.Minute,
				RecurrentFee:  utils.NewDecimal(15, 2),
				Unit:          tminDeminal,
				Increment:     ominDeminal,
			},
			{
				IntervalStart: 45 * time.Minute,
				RecurrentFee:  utils.NewDecimal(2, 1),
				Unit:          tminDeminal,
				Increment:     ominDeminal,
			},
		},
	}

	rp := &engine.RateProfile{
		Tenant: "cgrates.org",
		ID:     "RATE_PROFILE",
		Rates: map[string]*engine.Rate{
			rt1.ID: rt1,
			rt2.ID: rt2,
		},
	}
	if err := rp.Compile(); err != nil {
		t.Fatal(err)
	}

	ordRts := []*orderedRate{
		{
			0,
			rt1,
		},
		{
			20 * time.Minute,
			rt2,
		},
		{
			time.Hour + time.Minute,
			rt1,
		},
	}

	expOrdRts := []*engine.RateSInterval{
		{
			UsageStart: 0,
			Increments: []*engine.RateSIncrement{
				{
					UsageStart:        0,
					Rate:              rt1,
					IntervalRateIndex: 0,
					CompressFactor:    2,
					Usage:             20 * time.Minute,
				},
			},
			CompressFactor: 1,
		},
		{
			UsageStart: 20 * time.Minute,
			Increments: []*engine.RateSIncrement{
				{
					UsageStart:        20 * time.Minute,
					Rate:              rt2,
					IntervalRateIndex: 0,
					CompressFactor:    10,
					Usage:             10 * time.Minute,
				},
				{
					UsageStart:        30 * time.Minute,
					Rate:              rt2,
					IntervalRateIndex: 1,
					CompressFactor:    15,
					Usage:             15 * time.Minute,
				},
				{
					UsageStart:        45 * time.Minute,
					Rate:              rt2,
					IntervalRateIndex: 2,
					CompressFactor:    16,
					Usage:             16 * time.Minute,
				},
			},
			CompressFactor: 1,
		},
		{
			UsageStart: time.Hour + time.Minute,
			Increments: []*engine.RateSIncrement{
				{
					UsageStart:        time.Hour + time.Minute,
					Rate:              rt1,
					IntervalRateIndex: 1,
					CompressFactor:    6,
					Usage:             9 * time.Minute,
				},
			},
			CompressFactor: 1,
		},
	}

	if rcvOrdRts, err := computeRateSIntervals(ordRts, 0, time.Hour+10*time.Minute); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcvOrdRts, expOrdRts) {
		t.Errorf("Expected %+v, \nreceived %+v", utils.ToJSON(expOrdRts), utils.ToJSON(rcvOrdRts))
	}
}

func TestComputeRateSIntervalsCompressIncrements(t *testing.T) {
	minDecimal, err := utils.NewDecimalFromUsage("1m")
	if err != nil {
		t.Error(err)
	}
	tsecDecimal, err := utils.NewDecimalFromUsage("30s")
	if err != nil {
		t.Error(err)
	}

	rt1 := &engine.Rate{
		ID: "RATE1",
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
				RecurrentFee:  utils.NewDecimal(2, 1),
				Unit:          minDecimal,
				Increment:     tsecDecimal,
			},
			{
				IntervalStart: 30 * time.Minute,
				RecurrentFee:  utils.NewDecimal(2, 1),
				Unit:          minDecimal,
				Increment:     tsecDecimal,
			},
		},
	}

	rt2 := &engine.Rate{
		ID: "RATE2",
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
				RecurrentFee:  utils.NewDecimal(2, 1),
				Unit:          minDecimal,
				Increment:     tsecDecimal,
			},
			{
				IntervalStart: 30 * time.Minute,
				RecurrentFee:  utils.NewDecimal(2, 1),
				Unit:          minDecimal,
				Increment:     tsecDecimal,
			},
		},
	}

	rp := &engine.RateProfile{
		Tenant: "cgrates.org",
		ID:     "RATE_PROFILE",
		Rates: map[string]*engine.Rate{
			rt1.ID: rt1,
			rt2.ID: rt2,
		},
	}
	if err := rp.Compile(); err != nil {
		t.Fatal(err)
	}

	ordRts := []*orderedRate{
		{
			0,
			rt1,
		},
		{
			25 * time.Minute,
			rt1,
		},
	}

	expOrdRts := []*engine.RateSInterval{
		{
			UsageStart: 0,
			Increments: []*engine.RateSIncrement{
				{
					UsageStart:        0,
					Rate:              rt1,
					IntervalRateIndex: 0,
					CompressFactor:    50,
					Usage:             25 * time.Minute,
				},
			},
			CompressFactor: 1,
		},
		{
			UsageStart: 25 * time.Minute,
			Increments: []*engine.RateSIncrement{
				{
					UsageStart:        25 * time.Minute,
					Rate:              rt1,
					IntervalRateIndex: 0,
					CompressFactor:    10,
					Usage:             5 * time.Minute,
				},
				{
					UsageStart:        30 * time.Minute,
					Rate:              rt1,
					IntervalRateIndex: 1,
					CompressFactor:    60,
					Usage:             30 * time.Minute,
				},
			},
			CompressFactor: 1,
		},
	}

	if rcvOrdRts, err := computeRateSIntervals(ordRts, 0, time.Hour); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcvOrdRts, expOrdRts) {
		t.Errorf("Expected %+v, \nreceived %+v", utils.ToJSON(expOrdRts), utils.ToJSON(rcvOrdRts))
	}
}

func TestComputeRateSIntervalsStartAfterIntervalStartDifferentRates(t *testing.T) {
	tsecDecimal, err := utils.NewDecimalFromUsage("30s")
	if err != nil {
		t.Error(err)
	}

	tminDecimal, err := utils.NewDecimalFromUsage("10m")
	if err != nil {
		t.Error(err)
	}
	minDecimal, err := utils.NewDecimalFromUsage("1m")
	if err != nil {
		t.Error(err)
	}
	secDecimal, err := utils.NewDecimalFromUsage("1s")
	if err != nil {
		t.Error(err)
	}
	tensecDecimal, err := utils.NewDecimalFromUsage("10s")
	if err != nil {
		t.Error(err)
	}

	rt1 := &engine.Rate{
		ID: "RATE1",
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
				RecurrentFee:  utils.NewDecimal(22, 2),
				Unit:          minDecimal,
				Increment:     secDecimal,
			},
		},
	}

	rt2 := &engine.Rate{
		ID: "RATE2",
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 60 * time.Minute,
				RecurrentFee:  utils.NewDecimal(2, 1),
				Unit:          tminDecimal,
				Increment:     tensecDecimal,
			},
		},
	}
	twminDecimal, err := utils.NewDecimalFromUsage("20m")
	if err != nil {
		t.Error(err)
	}
	twsecDecimal, err := utils.NewDecimalFromUsage("20s")
	if err != nil {
		t.Error(err)
	}

	rt3 := &engine.Rate{
		ID: "RATE3",
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 120 * time.Minute,
				RecurrentFee:  utils.NewDecimal(18, 2),
				Unit:          twminDecimal,
				Increment:     twsecDecimal,
			},
		},
	}
	trminDecimal, err := utils.NewDecimalFromUsage("30m")
	if err != nil {
		t.Error(err)
	}

	rt4 := &engine.Rate{
		ID: "RATE4",
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 180 * time.Minute,
				RecurrentFee:  utils.NewDecimal(16, 2),
				Unit:          trminDecimal,
				Increment:     tsecDecimal,
			},
		},
	}

	rp := &engine.RateProfile{
		Tenant: "cgrates.org",
		ID:     "RATE_PROFILE",
		Rates: map[string]*engine.Rate{
			rt1.ID: rt1,
			rt2.ID: rt2,
			rt3.ID: rt3,
			rt4.ID: rt4,
		},
	}
	if err := rp.Compile(); err != nil {
		t.Fatal(err)
	}

	ordRts := []*orderedRate{
		{
			20 * time.Minute,
			rt1,
		},
		{
			80 * time.Minute,
			rt2,
		},
		{
			140 * time.Minute,
			rt3,
		},
		{
			200 * time.Minute,
			rt4,
		},
	}

	expOrdRts := []*engine.RateSInterval{
		{
			UsageStart: 20 * time.Minute,
			Increments: []*engine.RateSIncrement{
				{
					UsageStart:        20 * time.Minute,
					Rate:              rt1,
					IntervalRateIndex: 0,
					CompressFactor:    3600,
					Usage:             time.Hour,
				},
			},
			CompressFactor: 1,
		},
		{
			UsageStart: 80 * time.Minute,
			Increments: []*engine.RateSIncrement{
				{
					UsageStart:        80 * time.Minute,
					Rate:              rt2,
					IntervalRateIndex: 0,
					CompressFactor:    360,
					Usage:             time.Hour,
				},
			},
			CompressFactor: 1,
		},
		{
			UsageStart: 140 * time.Minute,
			Increments: []*engine.RateSIncrement{
				{
					UsageStart:        140 * time.Minute,
					Rate:              rt3,
					IntervalRateIndex: 0,
					CompressFactor:    180,
					Usage:             time.Hour,
				},
			},
			CompressFactor: 1,
		},
		{
			UsageStart: 200 * time.Minute,
			Increments: []*engine.RateSIncrement{
				{
					UsageStart:        200 * time.Minute,
					Rate:              rt4,
					IntervalRateIndex: 0,
					CompressFactor:    60,
					Usage:             30 * time.Minute,
				},
			},
			CompressFactor: 1,
		},
	}

	if rcvOrdRts, err := computeRateSIntervals(ordRts, 20*time.Minute, 210*time.Minute); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcvOrdRts, expOrdRts) {
		t.Errorf("Expected %+v, \nreceived %+v", utils.ToJSON(expOrdRts), utils.ToJSON(rcvOrdRts))
	}
}

func TestComputeRateSIntervalsStartAfterIntervalStartSameRate(t *testing.T) {
	tsecDecimal, err := utils.NewDecimalFromUsage("30s")
	if err != nil {
		t.Error(err)
	}
	minDecimal, err := utils.NewDecimalFromUsage("1m")
	if err != nil {
		t.Error(err)
	}
	secDecimal, err := utils.NewDecimalFromUsage("1s")
	if err != nil {
		t.Error(err)
	}
	twsecDecimal, err := utils.NewDecimalFromUsage("20s")
	if err != nil {
		t.Error(err)
	}
	tssecDecimal, err := utils.NewDecimalFromUsage("10s")
	if err != nil {
		t.Error(err)
	}
	tminDecimal, err := utils.NewDecimalFromUsage("10m")
	if err != nil {
		t.Error(err)
	}
	twminDecimal, err := utils.NewDecimalFromUsage("20m")
	if err != nil {
		t.Error(err)
	}
	thminDecimal, err := utils.NewDecimalFromUsage("30m")
	if err != nil {
		t.Error(err)
	}
	rt1 := &engine.Rate{
		ID: "RATE1",
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
				RecurrentFee:  utils.NewDecimal(22, 2),
				Unit:          minDecimal,
				Increment:     secDecimal,
			},
			{
				IntervalStart: 60 * time.Minute,
				RecurrentFee:  utils.NewDecimal(2, 1),
				Unit:          tminDecimal,
				Increment:     tssecDecimal,
			},
			{
				IntervalStart: 120 * time.Minute,
				RecurrentFee:  utils.NewDecimal(18, 2),
				Unit:          twminDecimal,
				Increment:     twsecDecimal,
			},
			{
				IntervalStart: 180 * time.Minute,
				RecurrentFee:  utils.NewDecimal(16, 2),
				Unit:          thminDecimal,
				Increment:     tsecDecimal,
			},
		},
	}
	rt1.Compile()

	ordRts := []*orderedRate{
		{
			20 * time.Minute,
			rt1,
		},
		{
			80 * time.Minute,
			rt1,
		},
		{
			140 * time.Minute,
			rt1,
		},
		{
			200 * time.Minute,
			rt1,
		},
	}
	expOrdRts := []*engine.RateSInterval{
		{
			UsageStart: 20 * time.Minute,
			Increments: []*engine.RateSIncrement{
				{
					UsageStart:        20 * time.Minute,
					Rate:              rt1,
					IntervalRateIndex: 0,
					CompressFactor:    2400,
					Usage:             40 * time.Minute,
				},
				{
					UsageStart:        60 * time.Minute,
					Rate:              rt1,
					IntervalRateIndex: 1,
					CompressFactor:    120,
					Usage:             20 * time.Minute,
				},
			},
			CompressFactor: 1,
		},
		{
			UsageStart: 80 * time.Minute,
			Increments: []*engine.RateSIncrement{
				{
					UsageStart:        80 * time.Minute,
					Rate:              rt1,
					IntervalRateIndex: 1,
					CompressFactor:    240,
					Usage:             40 * time.Minute,
				},
				{
					UsageStart:        120 * time.Minute,
					Rate:              rt1,
					IntervalRateIndex: 2,
					CompressFactor:    60,
					Usage:             20 * time.Minute,
				},
			},
			CompressFactor: 1,
		},
		{
			UsageStart: 140 * time.Minute,
			Increments: []*engine.RateSIncrement{
				{
					UsageStart:        140 * time.Minute,
					Rate:              rt1,
					IntervalRateIndex: 2,
					CompressFactor:    120,
					Usage:             40 * time.Minute,
				},
				{
					UsageStart:        180 * time.Minute,
					Rate:              rt1,
					IntervalRateIndex: 3,
					CompressFactor:    40,
					Usage:             20 * time.Minute,
				},
			},
			CompressFactor: 1,
		},
		{
			UsageStart: 200 * time.Minute,
			Increments: []*engine.RateSIncrement{
				{
					UsageStart:        200 * time.Minute,
					Rate:              rt1,
					IntervalRateIndex: 3,
					CompressFactor:    60,
					Usage:             30 * time.Minute,
				},
			},
			CompressFactor: 1,
		},
	}
	if rcvOrdRts, err := computeRateSIntervals(ordRts, 20*time.Minute, 210*time.Minute); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcvOrdRts, expOrdRts) {
		t.Errorf("Expected %+v, \nreceived %+v", utils.ToJSON(expOrdRts), utils.ToJSON(rcvOrdRts))
	}

	expOrdRts = []*engine.RateSInterval{
		{
			UsageStart: 0,
			Increments: []*engine.RateSIncrement{
				{
					UsageStart:        0,
					Rate:              rt1,
					IntervalRateIndex: 0,
					CompressFactor:    3600,
					Usage:             60 * time.Minute,
				},
				{
					UsageStart:        60 * time.Minute,
					Rate:              rt1,
					IntervalRateIndex: 1,
					CompressFactor:    120,
					Usage:             20 * time.Minute,
				},
			},
			CompressFactor: 1,
		},
		{
			UsageStart: 80 * time.Minute,
			Increments: []*engine.RateSIncrement{
				{
					UsageStart:        80 * time.Minute,
					Rate:              rt1,
					IntervalRateIndex: 1,
					CompressFactor:    240,
					Usage:             40 * time.Minute,
				},
				{
					UsageStart:        120 * time.Minute,
					Rate:              rt1,
					IntervalRateIndex: 2,
					CompressFactor:    60,
					Usage:             20 * time.Minute,
				},
			},
			CompressFactor: 1,
		},
		{
			UsageStart: 140 * time.Minute,
			Increments: []*engine.RateSIncrement{
				{
					UsageStart:        140 * time.Minute,
					Rate:              rt1,
					IntervalRateIndex: 2,
					CompressFactor:    120,
					Usage:             40 * time.Minute,
				},
				{
					UsageStart:        180 * time.Minute,
					Rate:              rt1,
					IntervalRateIndex: 3,
					CompressFactor:    40,
					Usage:             20 * time.Minute,
				},
			},
			CompressFactor: 1,
		},
		{
			UsageStart: 200 * time.Minute,
			Increments: []*engine.RateSIncrement{
				{
					UsageStart:        200 * time.Minute,
					Rate:              rt1,
					IntervalRateIndex: 3,
					CompressFactor:    60,
					Usage:             30 * time.Minute,
				},
			},
			CompressFactor: 1,
		},
	}

	if rcvOrdRts, err := computeRateSIntervals(ordRts, 0*time.Minute, 230*time.Minute); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcvOrdRts, expOrdRts) {
		t.Errorf("Expected %+v, \nreceived %+v", utils.ToJSON(expOrdRts), utils.ToJSON(rcvOrdRts))
	}
}

func TestComputeRateSIntervalsHalfDayIntervals(t *testing.T) {
	twminDecimal, err := utils.NewDecimalFromUsage("30m")
	if err != nil {
		t.Error(err)
	}

	nminDecimal, err := utils.NewDecimalFromUsage("9m")
	if err != nil {
		t.Error(err)
	}
	minDecimal, err := utils.NewDecimalFromUsage("1m")
	if err != nil {
		t.Error(err)
	}
	secDecimal, err := utils.NewDecimalFromUsage("1s")
	if err != nil {
		t.Error(err)
	}
	rt1 := &engine.Rate{
		ID: "RATE1",
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
				RecurrentFee:  utils.NewDecimal(5, 1),
				Unit:          twminDecimal,
				Increment:     minDecimal,
			},
			{
				IntervalStart: 24 * time.Hour,
				RecurrentFee:  utils.NewDecimal(1, 1),
				Unit:          twminDecimal,
				Increment:     secDecimal,
			},
		},
	}

	rt2 := &engine.Rate{
		ID: "RATE2",
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 4*time.Hour + 30*time.Minute,
				RecurrentFee:  utils.NewDecimal(45, 2),
				Unit:          twminDecimal,
				Increment:     nminDecimal,
			},
			{
				IntervalStart: 8*time.Hour + 30*time.Minute,
				RecurrentFee:  utils.NewDecimal(4, 1),
				Unit:          twminDecimal,
				Increment:     nminDecimal,
			},
			{
				IntervalStart: 16*time.Hour + 30*time.Minute,
				RecurrentFee:  utils.NewDecimal(35, 2),
				Unit:          twminDecimal,
				Increment:     nminDecimal,
			},
			{
				IntervalStart: 20*time.Hour + 30*time.Minute,
				RecurrentFee:  utils.NewDecimal(3, 1),
				Unit:          twminDecimal,
				Increment:     nminDecimal,
			},
		},
	}

	rt3 := &engine.Rate{
		ID: "RATE_SPECIAL",
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 12 * time.Hour,
				RecurrentFee:  utils.NewDecimal(2, 1),
				Unit:          twminDecimal,
				Increment:     secDecimal,
			},
		},
	}

	rp := &engine.RateProfile{
		Tenant: "cgrates.org",
		ID:     "RATE_PROFILE",
		Rates: map[string]*engine.Rate{
			rt1.ID: rt1,
			rt2.ID: rt2,
			rt3.ID: rt3,
		},
	}
	if err := rp.Compile(); err != nil {
		t.Error(err)
	}

	ordRts := []*orderedRate{
		{
			0,
			rt1,
		},
		{
			4*time.Hour + 31*time.Minute,
			rt2,
		},
		{
			12 * time.Hour,
			rt3,
		},
		{
			13 * time.Hour,
			rt2,
		},
		{
			24 * time.Hour,
			rt1,
		},
	}

	expOrdRts := []*engine.RateSInterval{
		{
			UsageStart: 0,
			Increments: []*engine.RateSIncrement{
				{
					UsageStart:        0,
					Rate:              rt1,
					IntervalRateIndex: 0,
					CompressFactor:    271,
					Usage:             4*time.Hour + 31*time.Minute,
				},
			},
			CompressFactor: 1,
		},
		{
			UsageStart: 4*time.Hour + 31*time.Minute,
			Increments: []*engine.RateSIncrement{
				{
					UsageStart:        4*time.Hour + 31*time.Minute,
					Rate:              rt2,
					IntervalRateIndex: 0,
					CompressFactor:    27,
					Usage:             3*time.Hour + 59*time.Minute,
				},
				{
					UsageStart:        8*time.Hour + 30*time.Minute,
					Rate:              rt2,
					IntervalRateIndex: 1,
					CompressFactor:    24,
					Usage:             3*time.Hour + 30*time.Minute,
				},
			},
			CompressFactor: 1,
		},
		{
			UsageStart: 12 * time.Hour,
			Increments: []*engine.RateSIncrement{
				{
					UsageStart:        12 * time.Hour,
					Rate:              rt3,
					IntervalRateIndex: 0,
					CompressFactor:    3600,
					Usage:             time.Hour,
				},
			},
			CompressFactor: 1,
		},
		{
			UsageStart: 13 * time.Hour,
			Increments: []*engine.RateSIncrement{
				{
					UsageStart:        13 * time.Hour,
					Rate:              rt2,
					IntervalRateIndex: 1,
					CompressFactor:    24,
					Usage:             3*time.Hour + 30*time.Minute,
				},
				{
					UsageStart:        16*time.Hour + 30*time.Minute,
					Rate:              rt2,
					IntervalRateIndex: 2,
					CompressFactor:    27,
					Usage:             4 * time.Hour,
				},
				{
					UsageStart:        20*time.Hour + 30*time.Minute,
					Rate:              rt2,
					IntervalRateIndex: 3,
					CompressFactor:    24,
					Usage:             3*time.Hour + 30*time.Minute,
				},
			},
			CompressFactor: 1,
		},
		{
			UsageStart: 24 * time.Hour,
			Increments: []*engine.RateSIncrement{
				{
					UsageStart:        24 * time.Hour,
					Rate:              rt1,
					IntervalRateIndex: 1,
					CompressFactor:    3600,
					Usage:             time.Hour,
				},
			},
			CompressFactor: 1,
		},
	}

	if rcvOrdRts, err := computeRateSIntervals(ordRts, 0, 25*time.Hour); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcvOrdRts, expOrdRts) {
		t.Errorf("Expected %+v, \nreceived %+v", utils.ToJSON(expOrdRts), utils.ToJSON(rcvOrdRts))
	}
}

func TestComputeRateSIntervalsConsecutiveRates(t *testing.T) {
	fminDecimal, err := utils.NewDecimalFromUsage("15m")
	if err != nil {
		t.Error(err)
	}
	eminDecimal, err := utils.NewDecimalFromUsage("11m")
	if err != nil {
		t.Error(err)
	}
	nminDecimal, err := utils.NewDecimalFromUsage("9m")
	if err != nil {
		t.Error(err)
	}
	sminDecimal, err := utils.NewDecimalFromUsage("7m")
	if err != nil {
		t.Error(err)
	}
	fvminDecimal, err := utils.NewDecimalFromUsage("5m")
	if err != nil {
		t.Error(err)
	}
	rt1 := &engine.Rate{
		ID: "RATE1",
		IntervalRates: []*engine.IntervalRate{
			{},
			{
				IntervalStart: 15 * time.Minute,
				RecurrentFee:  utils.NewDecimal(5, 1),
				Unit:          fminDecimal,
				Increment:     eminDecimal,
			},
			{
				IntervalStart: 30 * time.Minute,
				RecurrentFee:  utils.NewDecimal(4, 1),
				Unit:          fminDecimal,
				Increment:     nminDecimal,
			},
		},
	}

	rt2 := &engine.Rate{
		ID: "RATE2",
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 45 * time.Minute,
				RecurrentFee:  utils.NewDecimal(3, 1),
				Unit:          fminDecimal,
				Increment:     sminDecimal,
			},
			{
				IntervalStart: time.Hour,
				RecurrentFee:  utils.NewDecimal(2, 1),
				Unit:          fminDecimal,
				Increment:     fvminDecimal,
			},
		},
	}

	rp := &engine.RateProfile{
		Tenant: "cgrates.org",
		ID:     "RATE_PROFILE",
		Rates: map[string]*engine.Rate{
			rt1.ID: rt1,
			rt2.ID: rt2,
		},
	}
	if err := rp.Compile(); err != nil {
		t.Error(err)
	}

	ordRts := []*orderedRate{
		{
			0,
			rt1,
		},
		{
			45 * time.Minute,
			rt2,
		},
	}

	expOrdRts := []*engine.RateSInterval{
		{
			UsageStart: 15 * time.Minute,
			Increments: []*engine.RateSIncrement{
				{
					UsageStart:        15 * time.Minute,
					Rate:              rt1,
					IntervalRateIndex: 1,
					CompressFactor:    2,
					Usage:             15 * time.Minute,
				},
				{
					UsageStart:        30 * time.Minute,
					Rate:              rt1,
					IntervalRateIndex: 2,
					CompressFactor:    2,
					Usage:             15 * time.Minute,
				},
			},
			CompressFactor: 1,
		},
		{
			UsageStart: 45 * time.Minute,
			Increments: []*engine.RateSIncrement{
				{
					UsageStart:        45 * time.Minute,
					Rate:              rt2,
					IntervalRateIndex: 0,
					CompressFactor:    3,
					Usage:             15 * time.Minute,
				},
				{
					UsageStart:        time.Hour,
					Rate:              rt2,
					IntervalRateIndex: 1,
					CompressFactor:    6,
					Usage:             30 * time.Minute,
				},
			},
			CompressFactor: 1,
		},
	}

	if rcvOrdRts, err := computeRateSIntervals(ordRts, 15*time.Minute, time.Hour+15*time.Minute); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcvOrdRts, expOrdRts) {
		t.Errorf("Expected %+v, \nreceived %+v", utils.ToJSON(expOrdRts), utils.ToJSON(rcvOrdRts))
	}
}

func TestComputeRateSIntervalsRatesByMinutes(t *testing.T) {
	tsecDecimal, err := utils.NewDecimalFromUsage("2s")
	if err != nil {
		t.Error(err)
	}
	secDecimal, err := utils.NewDecimalFromUsage("1s")
	if err != nil {
		t.Error(err)
	}
	tminDecimal, err := utils.NewDecimalFromUsage("2m")
	if err != nil {
		t.Error(err)
	}
	sminDecimal, err := utils.NewDecimalFromUsage("6m")
	if err != nil {
		t.Error(err)
	}
	nminDecimal, err := utils.NewDecimalFromUsage("9m")
	if err != nil {
		t.Error(err)
	}
	eminDecimal, err := utils.NewDecimalFromUsage("8m")
	if err != nil {
		t.Error(err)
	}
	rt1 := &engine.Rate{
		ID: "RATE1",
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
				RecurrentFee:  utils.NewDecimal(5, 1),
				Unit:          tminDecimal,
				Increment:     tsecDecimal,
			},
			{
				IntervalStart: 12*time.Minute + 35*time.Second,
				RecurrentFee:  utils.NewDecimal(4, 1),
				Unit:          sminDecimal,
				Increment:     secDecimal,
			},
			{
				IntervalStart: time.Hour + 37*time.Minute + 19*time.Second,
				RecurrentFee:  utils.NewDecimal(2, 1),
				Unit:          nminDecimal,
				Increment:     sminDecimal,
			},
		},
	}
	rt1.Compile()

	rt2 := &engine.Rate{
		ID: "RATE2",
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 38*time.Minute + 15*time.Second,
				RecurrentFee:  utils.NewDecimal(1, 1),
				Unit:          eminDecimal,
				Increment:     secDecimal,
			},
		},
	}
	rt2.Compile()

	ordRts := []*orderedRate{
		{
			0,
			rt1,
		},
		{
			39 * time.Minute,
			rt2,
		},
		{
			time.Hour + 37*time.Minute + 19*time.Second,
			rt1,
		},
	}

	expOrdRts := []*engine.RateSInterval{
		{
			UsageStart: 0,
			Increments: []*engine.RateSIncrement{
				{
					UsageStart:        0,
					Rate:              rt1,
					IntervalRateIndex: 0,
					CompressFactor:    378,
					Usage:             12*time.Minute + 35*time.Second,
				},
				{
					UsageStart:        12*time.Minute + 35*time.Second,
					Rate:              rt1,
					IntervalRateIndex: 1,
					CompressFactor:    1585,
					Usage:             26*time.Minute + 25*time.Second,
				},
			},
			CompressFactor: 1,
		},
		{
			UsageStart: 39 * time.Minute,
			Increments: []*engine.RateSIncrement{
				{
					UsageStart:        39 * time.Minute,
					Rate:              rt2,
					IntervalRateIndex: 0,
					CompressFactor:    3499,
					Usage:             58*time.Minute + 19*time.Second,
				},
			},
			CompressFactor: 1,
		},
		{
			UsageStart: time.Hour + 37*time.Minute + 19*time.Second,
			Increments: []*engine.RateSIncrement{
				{
					UsageStart:        time.Hour + 37*time.Minute + 19*time.Second,
					Rate:              rt1,
					IntervalRateIndex: 2,
					CompressFactor:    1,
					Usage:             41 * time.Second,
				},
			},
			CompressFactor: 1,
		},
	}

	if rcvOrdRts, err := computeRateSIntervals(ordRts, 0, time.Hour+38*time.Minute); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcvOrdRts, expOrdRts) {
		t.Errorf("Expected %+v, \nreceived %+v", utils.ToJSON(expOrdRts), utils.ToJSON(rcvOrdRts))
	}
}

func TestComputeRateSIntervalsSwitchingRates2(t *testing.T) {
	tsecDecimal, err := utils.NewDecimalFromUsage("2s")
	if err != nil {
		t.Error(err)
	}
	tminDecimal, err := utils.NewDecimalFromUsage("2m")
	if err != nil {
		t.Error(err)
	}
	ttminDecimal, err := utils.NewDecimalFromUsage("10m")
	if err != nil {
		t.Error(err)
	}
	fsecDecimal, err := utils.NewDecimalFromUsage("4s")
	if err != nil {
		t.Error(err)
	}
	fminDecimal, err := utils.NewDecimalFromUsage("4m")
	if err != nil {
		t.Error(err)
	}
	rt1 := &engine.Rate{
		ID: "RATE1",
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
				RecurrentFee:  utils.NewDecimal(5, 1),
				Unit:          tminDecimal,
				Increment:     tsecDecimal,
			},
		},
	}
	rt1.Compile()

	rt2 := &engine.Rate{
		ID: "RATE2",
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 20 * time.Minute,
				RecurrentFee:  utils.NewDecimal(4, 1),
				Unit:          ttminDecimal,
				Increment:     tsecDecimal,
			},
			{
				IntervalStart: 40 * time.Minute,
				RecurrentFee:  utils.NewDecimal(3, 1),
				Unit:          ttminDecimal,
				Increment:     fsecDecimal,
			},
			{
				IntervalStart: time.Hour,
				RecurrentFee:  utils.NewDecimal(2, 1),
				Unit:          ttminDecimal,
				Increment:     tminDecimal,
			},
			{
				IntervalStart: 2 * time.Hour,
				RecurrentFee:  utils.NewDecimal(1, 1),
				Unit:          ttminDecimal,
				Increment:     fminDecimal,
			},
		},
	}
	rt2.Compile()

	ordRts := []*orderedRate{
		{
			0,
			rt1,
		},
		{
			20 * time.Minute,
			rt2,
		},
		{
			21 * time.Minute,
			rt1,
		},
		{
			40 * time.Minute,
			rt2,
		},
		{
			41 * time.Minute,
			rt1,
		},
		{
			time.Hour,
			rt2,
		},
		{
			time.Hour + time.Minute,
			rt1,
		},
		{
			2 * time.Hour,
			rt2,
		},
		{
			2*time.Hour + time.Minute,
			rt1,
		},
	}

	expOrdRts := []*engine.RateSInterval{
		{
			UsageStart: 0,
			Increments: []*engine.RateSIncrement{
				{
					UsageStart:        0,
					Rate:              rt1,
					IntervalRateIndex: 0,
					CompressFactor:    600,
					Usage:             20 * time.Minute,
				},
			},
			CompressFactor: 1,
		},
		{
			UsageStart: 20 * time.Minute,
			Increments: []*engine.RateSIncrement{
				{
					UsageStart:        20 * time.Minute,
					Rate:              rt2,
					IntervalRateIndex: 0,
					CompressFactor:    30,
					Usage:             time.Minute,
				},
			},
			CompressFactor: 1,
		},
		{
			UsageStart: 21 * time.Minute,
			Increments: []*engine.RateSIncrement{
				{
					UsageStart:        21 * time.Minute,
					Rate:              rt1,
					IntervalRateIndex: 0,
					CompressFactor:    570,
					Usage:             19 * time.Minute,
				},
			},
			CompressFactor: 1,
		},
		{
			UsageStart: 40 * time.Minute,
			Increments: []*engine.RateSIncrement{
				{
					UsageStart:        40 * time.Minute,
					Rate:              rt2,
					IntervalRateIndex: 1,
					CompressFactor:    15,
					Usage:             time.Minute,
				},
			},
			CompressFactor: 1,
		},
		{
			UsageStart: 41 * time.Minute,
			Increments: []*engine.RateSIncrement{
				{
					UsageStart:        41 * time.Minute,
					Rate:              rt1,
					IntervalRateIndex: 0,
					CompressFactor:    570,
					Usage:             19 * time.Minute,
				},
			},
			CompressFactor: 1,
		},
		{
			UsageStart: time.Hour,
			Increments: []*engine.RateSIncrement{
				{
					UsageStart:        time.Hour,
					Rate:              rt2,
					IntervalRateIndex: 2,
					CompressFactor:    1,
					Usage:             time.Minute,
				},
			},
			CompressFactor: 1,
		},
		{
			UsageStart: time.Hour + time.Minute,
			Increments: []*engine.RateSIncrement{
				{
					UsageStart:        time.Hour + time.Minute,
					Rate:              rt1,
					IntervalRateIndex: 0,
					CompressFactor:    1770,
					Usage:             59 * time.Minute,
				},
			},
			CompressFactor: 1,
		},
		{
			UsageStart: 2 * time.Hour,
			Increments: []*engine.RateSIncrement{
				{
					UsageStart:        2 * time.Hour,
					Rate:              rt2,
					IntervalRateIndex: 3,
					CompressFactor:    1,
					Usage:             time.Minute,
				},
			},
			CompressFactor: 1,
		},
		{
			UsageStart: 2*time.Hour + time.Minute,
			Increments: []*engine.RateSIncrement{
				{
					UsageStart:        2*time.Hour + time.Minute,
					Rate:              rt1,
					IntervalRateIndex: 0,
					CompressFactor:    30,
					Usage:             time.Minute,
				},
			},
			CompressFactor: 1,
		},
	}

	if rcvOrdRts, err := computeRateSIntervals(ordRts, 0, 2*time.Hour+2*time.Minute); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcvOrdRts, expOrdRts) {
		t.Errorf("Expected %+v, \nreceived %+v", utils.ToJSON(expOrdRts), utils.ToJSON(rcvOrdRts))
	}
}

func TestComputeRateSIntervalsSOneWeekCall(t *testing.T) {
	minDecimal, err := utils.NewDecimalFromUsage("1m")
	if err != nil {
		t.Error(err)
	}
	hourDecimal, err := utils.NewDecimalFromUsage("1h")
	if err != nil {
		t.Error(err)
	}
	fminDecimal, err := utils.NewDecimalFromUsage("17m")
	if err != nil {
		t.Error(err)
	}
	rt1 := &engine.Rate{
		ID: "RATE1",
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
				RecurrentFee:  utils.NewDecimal(5, 1),
				Unit:          hourDecimal,
				Increment:     fminDecimal,
			},
		},
	}
	rt1.Compile()

	rt2 := &engine.Rate{
		ID: "RATE2",
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 168 * time.Hour,
				RecurrentFee:  utils.NewDecimal(5, 1),
				Unit:          minDecimal,
				Increment:     minDecimal,
			},
		},
	}
	rt2.Compile()

	ordRts := []*orderedRate{
		{
			0,
			rt1,
		},
		{
			168 * time.Hour,
			rt2,
		},
	}

	expOrdRts := []*engine.RateSInterval{
		{
			UsageStart: 0,
			Increments: []*engine.RateSIncrement{
				{
					UsageStart:        0,
					Rate:              rt1,
					IntervalRateIndex: 0,
					CompressFactor:    593,
					Usage:             168 * time.Hour,
				},
			},
			CompressFactor: 1,
		},
		{
			UsageStart: 168 * time.Hour,
			Increments: []*engine.RateSIncrement{
				{
					UsageStart:        168 * time.Hour,
					Rate:              rt2,
					IntervalRateIndex: 0,
					CompressFactor:    60,
					Usage:             time.Hour,
				},
			},
			CompressFactor: 1,
		},
	}

	if rcvOrdRts, err := computeRateSIntervals(ordRts, 0, 169*time.Hour); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcvOrdRts, expOrdRts) {
		t.Errorf("Expected %+v, \nreceived %+v", utils.ToJSON(expOrdRts), utils.ToJSON(rcvOrdRts))
	}
}

func TestComputeRateSIntervalsPauseBetweenRates(t *testing.T) {
	minDecimal, err := utils.NewDecimalFromUsage("1m")
	if err != nil {
		t.Error(err)
	}
	secDecimal, err := utils.NewDecimalFromUsage("1s")
	if err != nil {
		t.Error(err)
	}
	fminDecimal, err := utils.NewDecimalFromUsage("5m")
	if err != nil {
		t.Error(err)
	}
	tminDecimal, err := utils.NewDecimalFromUsage("10m")
	if err != nil {
		t.Error(err)
	}
	rt1 := &engine.Rate{
		ID: "RATE1",
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
				RecurrentFee:  utils.NewDecimal(2, 1),
				Unit:          fminDecimal,
				Increment:     minDecimal,
			},
			{
				IntervalStart: 20 * time.Minute,
				RecurrentFee:  utils.NewDecimal(4, 1),
				Unit:          minDecimal,
				Increment:     secDecimal,
			},
		},
	}

	rt2 := &engine.Rate{
		ID: "RATE2",
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 50 * time.Minute,
				RecurrentFee:  utils.NewDecimal(1, 1),
				Unit:          tminDecimal,
				Increment:     minDecimal,
			},
			{
				IntervalStart: time.Hour,
				RecurrentFee:  utils.NewDecimal(5, 3),
				Unit:          fminDecimal,
				Increment:     secDecimal,
			},
		},
	}

	ordRts := []*orderedRate{
		{
			0,
			rt1,
		},
		{
			time.Hour,
			rt2,
		},
	}

	expOrdRts := []*engine.RateSInterval{
		{
			UsageStart: 0,
			Increments: []*engine.RateSIncrement{
				{
					UsageStart:        0,
					Rate:              rt1,
					IntervalRateIndex: 0,
					CompressFactor:    20,
					Usage:             20 * time.Minute,
				},
				{
					UsageStart:        20 * time.Minute,
					Rate:              rt1,
					IntervalRateIndex: 1,
					CompressFactor:    2400,
					Usage:             40 * time.Minute,
				},
			},
			CompressFactor: 1,
		},
		{
			UsageStart: time.Hour,
			Increments: []*engine.RateSIncrement{
				{
					UsageStart:        time.Hour,
					Rate:              rt2,
					IntervalRateIndex: 1,
					CompressFactor:    1200,
					Usage:             20 * time.Minute,
				},
			},
			CompressFactor: 1,
		},
	}

	if rcvOrdRts, err := computeRateSIntervals(ordRts, 0, time.Hour+20*time.Minute); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcvOrdRts, expOrdRts) {
		t.Errorf("Expected %+v, \nreceived %+v", utils.ToJSON(expOrdRts), utils.ToJSON(rcvOrdRts))
	}
}
