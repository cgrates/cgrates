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
	"math"
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/utils"
)

func TestOrderRatesOnIntervals11(t *testing.T) {
	rt0 := &utils.Rate{
		ID: "RATE0",
		Weights: utils.DynamicWeights{
			{
				Weight: 0,
			},
		},
		IntervalRates: []*utils.IntervalRate{
			{
				IntervalStart: utils.NewDecimal(0, 0),
			},
		},
	}
	rt0.Compile()
	rtChristmas := &utils.Rate{
		ID:              "RT_CHRISTMAS",
		ActivationTimes: "* * 24 12 *",
		Weights: utils.DynamicWeights{
			{
				Weight: 50,
			},
		},
		IntervalRates: []*utils.IntervalRate{
			{
				IntervalStart: utils.NewDecimal(0, 0),
			},
		},
	}
	rtChristmas.Compile()
	allRts := []*utils.Rate{rt0, rtChristmas}
	expOrdered := []*orderedRate{
		{
			utils.NewDecimal(0, 0).Big,
			rt0,
		},
	}
	wghts := []float64{10, 50}
	sTime := time.Date(2020, time.June, 28, 18, 56, 05, 0, time.UTC)
	usage := utils.NewDecimal(int64(2*time.Minute), 0)
	if ordRts, err := orderRatesOnIntervals(
		allRts, wghts, sTime, usage.Big, true, 10); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expOrdered, ordRts) {
		t.Errorf("expecting: %s\n, received: %s", utils.ToJSON(expOrdered), utils.ToJSON(ordRts))
	}

	expOrdered = []*orderedRate{
		{
			utils.NewDecimal(0, 0).Big,
			rt0,
		},
		{
			utils.NewDecimal(int64(55*time.Second), 0).Big,
			rtChristmas,
		},
	}
	sTime = time.Date(2020, time.December, 23, 23, 59, 05, 0, time.UTC)
	usage = utils.NewDecimal(int64(2*time.Minute), 0)
	if ordRts, err := orderRatesOnIntervals(
		allRts, wghts, sTime, usage.Big, true, 10); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expOrdered, ordRts) {
		t.Errorf("expecting: %s\n, received: %s",
			utils.ToJSON(expOrdered), utils.ToJSON(ordRts))
	}

	expOrdered = []*orderedRate{
		{
			utils.NewDecimal(0, 0).Big,
			rt0,
		},
		{
			utils.NewDecimal(int64(55*time.Second), 0).Big,
			rtChristmas,
		},
		{
			utils.NewDecimal(int64(86455*time.Second), 0).Big,
			rt0,
		},
	}
	sTime = time.Date(2020, time.December, 23, 23, 59, 05, 0, time.UTC)
	usage = utils.NewDecimal(int64(25*time.Hour), 0)
	if ordRts, err := orderRatesOnIntervals(
		allRts, wghts, sTime, usage.Big, true, 10); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expOrdered, ordRts) {
		t.Errorf("expecting: %s\n, received: %s",
			utils.ToJSON(expOrdered), utils.ToJSON(ordRts))
	}

	rts := []*utils.Rate{rtChristmas}
	expOrdered = nil
	sTime = time.Date(2020, time.December, 25, 23, 59, 05, 0, time.UTC)
	usage = utils.NewDecimal(int64(2*time.Minute), 0)
	if ordRts, err := orderRatesOnIntervals(
		rts, wghts, sTime, usage.Big, true, 10); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expOrdered, ordRts) {
		t.Errorf("expecting: %s\n, received: %s",
			utils.ToJSON(expOrdered), utils.ToJSON(ordRts))
	}
}

func TestOrderRatesOnIntervalsChristmasDay(t *testing.T) {
	rt1 := &utils.Rate{
		ID: "ALWAYS_RATE",
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
		IntervalRates: []*utils.IntervalRate{
			{
				IntervalStart: utils.NewDecimal(0, 0),
			},
		},
	}
	rtCh1 := &utils.Rate{
		ID:              "CHRISTMAS1",
		ActivationTimes: "* 0-6 24 12 *",
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
		IntervalRates: []*utils.IntervalRate{
			{
				IntervalStart: utils.NewDecimal(0, 0),
			},
		},
	}
	rtCh2 := &utils.Rate{
		ID:              "CHRISTMAS2",
		ActivationTimes: "* 7-12 24 12 *",
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
		IntervalRates: []*utils.IntervalRate{
			{
				IntervalStart: utils.NewDecimal(0, 0),
			},
		},
	}
	rtCh3 := &utils.Rate{
		ID:              "CHRISTMAS3",
		ActivationTimes: "* 13-19 24 12 *",
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
		IntervalRates: []*utils.IntervalRate{
			{
				IntervalStart: utils.NewDecimal(0, 0),
			},
		},
	}
	rtCh4 := &utils.Rate{
		ID:              "CHRISTMAS4",
		ActivationTimes: "* 20-23 24 12 *",
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
		IntervalRates: []*utils.IntervalRate{
			{
				IntervalStart: utils.NewDecimal(0, 0),
			},
		},
	}
	wghts := []float64{10, 20, 20, 20, 20}
	allRates := []*utils.Rate{rt1, rtCh1, rtCh2, rtCh3, rtCh4}
	for _, idx := range allRates {
		if err := idx.Compile(); err != nil {
			t.Error(err)
		}
	}
	sTime := time.Date(2020, 12, 23, 22, 0, 0, 0, time.UTC)
	usage := utils.NewDecimal(int64(31*time.Hour), 0)
	expected := []*orderedRate{
		{
			utils.NewDecimal(0, 0).Big,
			rt1,
		},
		{
			utils.NewDecimal(int64(2*time.Hour), 0).Big,
			rtCh1,
		},
		{
			utils.NewDecimal(int64(9*time.Hour), 0).Big,
			rtCh2,
		},
		{
			utils.NewDecimal(int64(15*time.Hour), 0).Big,
			rtCh3,
		},
		{
			utils.NewDecimal(int64(22*time.Hour), 0).Big,
			rtCh4,
		},
		{
			utils.NewDecimal(int64(26*time.Hour), 0).Big,
			rt1,
		},
	}
	if ordRts, err := orderRatesOnIntervals(allRates, wghts, sTime, usage.Big, true, 10); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(ordRts, expected) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(ordRts))
	}
}

func TestOrderRatesOnIntervalsDoubleRates1(t *testing.T) {
	rt1 := &utils.Rate{
		ID: "ALWAYS_RATE",
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
		IntervalRates: []*utils.IntervalRate{
			{
				IntervalStart: utils.NewDecimal(0, 0),
			},
		},
	}
	rtCh1 := &utils.Rate{
		ID:              "CHRISTMAS1",
		ActivationTimes: "* * 24 12 *",
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
		IntervalRates: []*utils.IntervalRate{
			{
				IntervalStart: utils.NewDecimal(0, 0),
			},
		},
	}
	rtCh2 := &utils.Rate{
		ID:              "CHRISTMAS2",
		ActivationTimes: "* 18-23 24 12 *",
		Weights: utils.DynamicWeights{
			{
				Weight: 30,
			},
		},
		IntervalRates: []*utils.IntervalRate{
			{
				IntervalStart: utils.NewDecimal(0, 0),
			},
		},
	}
	wghts := []float64{10, 20, 30}
	allRates := []*utils.Rate{rt1, rtCh1, rtCh2}
	for _, idx := range allRates {
		if err := idx.Compile(); err != nil {
			t.Error(err)
		}
	}
	sTime := time.Date(2020, 12, 23, 21, 28, 12, 0, time.UTC)
	usage := utils.NewDecimal(int64(31*time.Hour), 0)
	expected := []*orderedRate{
		{
			utils.NewDecimal(0, 0).Big,
			rt1,
		},
		{
			utils.NewDecimal(int64(2*time.Hour+31*time.Minute+48*time.Second), 0).Big,
			rtCh1,
		},
		{
			utils.NewDecimal(int64(20*time.Hour+31*time.Minute+48*time.Second), 0).Big,
			rtCh2,
		},
		{
			utils.NewDecimal(int64(26*time.Hour+31*time.Minute+48*time.Second), 0).Big,
			rt1,
		},
	}
	if ordRts, err := orderRatesOnIntervals(allRates, wghts, sTime, usage.Big, true, 10); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(ordRts, expected) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(ordRts))
	}
}

func TestOrderRatesOnIntervalsEveryTwentyFiveMins(t *testing.T) {
	rtTwentyFiveMins := &utils.Rate{
		ID:              "TWENTYFIVE_MINS",
		ActivationTimes: "*/25 * * * *",
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
		IntervalRates: []*utils.IntervalRate{
			{
				IntervalStart: utils.NewDecimal(0, 0),
			},
		},
	}
	rt1 := &utils.Rate{
		ID:              "DAY_RATE",
		ActivationTimes: "* * * * 3",
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
		IntervalRates: []*utils.IntervalRate{
			{
				IntervalStart: utils.NewDecimal(0, 0),
			},
		},
	}
	wghts := []float64{10, 20}
	allRates := []*utils.Rate{rt1, rtTwentyFiveMins}
	for _, idx := range allRates {
		if err := idx.Compile(); err != nil {
			t.Error(err)
		}
	}
	sTime := time.Date(2020, 10, 28, 20, 0, 0, 0, time.UTC)
	usage := utils.NewDecimal(int64(time.Hour), 0)
	expected := []*orderedRate{
		{
			utils.NewDecimal(0, 0).Big,
			rtTwentyFiveMins,
		},
		{
			utils.NewDecimal(int64(time.Minute), 0).Big,
			rt1,
		},
		{
			utils.NewDecimal(int64(25*time.Minute), 0).Big,
			rtTwentyFiveMins,
		},
		{
			utils.NewDecimal(int64(26*time.Minute), 0).Big,
			rt1,
		},
		{
			utils.NewDecimal(int64(50*time.Minute), 0).Big,
			rtTwentyFiveMins,
		},
		{
			utils.NewDecimal(int64(51*time.Minute), 0).Big,
			rt1,
		},
	}
	if ordRts, err := orderRatesOnIntervals(allRates, wghts, sTime, usage.Big, true, 10); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(ordRts, expected) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(ordRts))
	}
}

func TestOrderRatesOnIntervalsOneMinutePause(t *testing.T) {
	rt1 := &utils.Rate{
		ID:              "ALWAYS_RATE",
		ActivationTimes: "26 * * * *",
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
		IntervalRates: []*utils.IntervalRate{
			{
				IntervalStart: utils.NewDecimal(0, 0),
			},
		},
	}
	rtFirstInterval := &utils.Rate{
		ID:              "FIRST_INTERVAL",
		ActivationTimes: "0-25 * * * *",
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
		IntervalRates: []*utils.IntervalRate{
			{
				IntervalStart: utils.NewDecimal(0, 0),
			},
		},
	}
	rtSecondINterval := &utils.Rate{
		ID:              "SECOND_INTERVAL",
		ActivationTimes: "27-59 * * * *",
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
		IntervalRates: []*utils.IntervalRate{
			{
				IntervalStart: utils.NewDecimal(0, 0),
			},
		},
	}
	wghts := []float64{10, 20, 20}
	allRates := []*utils.Rate{rt1, rtFirstInterval, rtSecondINterval}
	for _, idx := range allRates {
		if err := idx.Compile(); err != nil {
			t.Error(err)
		}
	}
	sTime := time.Date(2020, 10, 28, 20, 0, 0, 0, time.UTC)
	usage := utils.NewDecimal(int64(time.Hour), 0)
	expected := []*orderedRate{
		{
			utils.NewDecimal(0, 0).Big,
			rtFirstInterval,
		},
		{
			utils.NewDecimal(int64(26*time.Minute), 0).Big,
			rt1,
		},
		{
			utils.NewDecimal(int64(27*time.Minute), 0).Big,
			rtSecondINterval,
		},
	}
	if ordRts, err := orderRatesOnIntervals(allRates, wghts, sTime, usage.Big, true, 10); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(ordRts, expected) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(ordRts))
	}
}

func TestOrderRatesOnIntervalsNewYear(t *testing.T) {
	rt1 := &utils.Rate{
		ID: "ALWAYS_RATE",
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
		IntervalRates: []*utils.IntervalRate{
			{
				IntervalStart: utils.NewDecimal(0, 0),
			},
		},
	}
	rt1NewYear := &utils.Rate{
		ID:              "NEW_YEAR1",
		ActivationTimes: "* 20-23 * * *",
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
		IntervalRates: []*utils.IntervalRate{
			{
				IntervalStart: utils.NewDecimal(0, 0),
			},
		},
	}
	rt1NewYear2 := &utils.Rate{
		ID:              "NEW_YEAR2",
		ActivationTimes: "0-30 22 * * *",
		Weights: utils.DynamicWeights{
			{
				Weight: 30,
			},
		},
		IntervalRates: []*utils.IntervalRate{
			{
				IntervalStart: utils.NewDecimal(0, 0),
			},
		},
	}
	wghts := []float64{10, 20, 30}
	allRates := []*utils.Rate{rt1, rt1NewYear, rt1NewYear2}
	for _, idx := range allRates {
		if err := idx.Compile(); err != nil {
			t.Error(err)
		}
	}
	sTime := time.Date(2020, 12, 30, 23, 0, 0, 0, time.UTC)
	usage := utils.NewDecimal(int64(26*time.Hour), 0)
	expected := []*orderedRate{
		{
			utils.NewDecimal(0, 0).Big,
			rt1NewYear,
		},
		{
			utils.NewDecimal(int64(time.Hour), 0).Big,
			rt1,
		},
		{
			utils.NewDecimal(int64(21*time.Hour), 0).Big,
			rt1NewYear,
		},
		{
			utils.NewDecimal(int64(23*time.Hour), 0).Big,
			rt1NewYear2,
		},
		{
			utils.NewDecimal(int64(23*time.Hour+31*time.Minute), 0).Big,
			rt1NewYear,
		},
		{
			utils.NewDecimal(int64(25*time.Hour), 0).Big,
			rt1,
		},
	}
	if ordRts, err := orderRatesOnIntervals(allRates, wghts, sTime, usage.Big, true, 10); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(ordRts, expected) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(ordRts))
	}
}

func TestOrderRateOnIntervalsEveryHourEveryDay(t *testing.T) {
	rtEveryHour := &utils.Rate{
		ID:              "HOUR_RATE",
		ActivationTimes: "* */1 * * *",
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
		IntervalRates: []*utils.IntervalRate{
			{
				IntervalStart: utils.NewDecimal(0, 0),
			},
		},
	}
	rtEveryDay := &utils.Rate{
		ID:              "DAY_RATE",
		ActivationTimes: "* * 22 * *",
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
		IntervalRates: []*utils.IntervalRate{
			{
				IntervalStart: utils.NewDecimal(0, 0),
			},
		},
	}
	wghts := []float64{10, 20}
	allRates := []*utils.Rate{rtEveryHour, rtEveryDay}
	for _, idx := range allRates {
		if err := idx.Compile(); err != nil {
			t.Error(err)
		}
	}
	sTime := time.Date(2020, 7, 21, 10, 24, 15, 0, time.UTC)
	usage := utils.NewDecimal(int64(49*time.Hour), 0)
	expected := []*orderedRate{
		{
			utils.NewDecimal(0, 0).Big,
			rtEveryHour,
		},
		{
			utils.NewDecimal(int64(13*time.Hour+35*time.Minute+45*time.Second), 0).Big,
			rtEveryDay,
		},
		{
			utils.NewDecimal(int64(37*time.Hour+35*time.Minute+45*time.Second), 0).Big,
			rtEveryHour,
		},
	}
	if ordRts, err := orderRatesOnIntervals(allRates, wghts, sTime, usage.Big, true, 10); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(ordRts, expected) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(ordRts))
	}
}

func TestOrderRatesOnIntervalsOneHourInThreeRates(t *testing.T) {
	rtOneHour1 := &utils.Rate{
		ID:              "HOUR_RATE_1",
		ActivationTimes: "0-19 * * * *",
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
		IntervalRates: []*utils.IntervalRate{
			{
				IntervalStart: utils.NewDecimal(0, 0),
			},
		},
	}
	rtOneHour2 := &utils.Rate{
		ID:              "HOUR_RATE_2",
		ActivationTimes: "20-39 * * * *",
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
		IntervalRates: []*utils.IntervalRate{
			{
				IntervalStart: utils.NewDecimal(0, 0),
			},
		},
	}
	rtOneHour3 := &utils.Rate{
		ID:              "HOUR_RATE_3",
		ActivationTimes: "40-59 * * * *",
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
		IntervalRates: []*utils.IntervalRate{
			{
				IntervalStart: utils.NewDecimal(0, 0),
			},
		},
	}
	wghts := []float64{20, 20, 20}
	allRates := []*utils.Rate{rtOneHour1, rtOneHour2, rtOneHour3}
	for _, idx := range allRates {
		if err := idx.Compile(); err != nil {
			t.Error(err)
		}
	}
	sTime := time.Date(2020, 7, 15, 10, 59, 59, 0, time.UTC)
	usage := utils.NewDecimal(int64(2*time.Hour), 0)
	expected := []*orderedRate{
		{
			utils.NewDecimal(0, 0).Big,
			rtOneHour3,
		},
		{
			utils.NewDecimal(int64(time.Second), 0).Big,
			rtOneHour1,
		},
		{
			utils.NewDecimal(int64(20*time.Minute+time.Second), 0).Big,
			rtOneHour2,
		},
		{
			utils.NewDecimal(int64(40*time.Minute+time.Second), 0).Big,
			rtOneHour3,
		},
		{
			utils.NewDecimal(int64(time.Hour+time.Second), 0).Big,
			rtOneHour1,
		},
		{
			utils.NewDecimal(int64(time.Hour+20*time.Minute+time.Second), 0).Big,
			rtOneHour2,
		},
		{
			utils.NewDecimal(int64(time.Hour+40*time.Minute+time.Second), 0).Big,
			rtOneHour3,
		},
	}
	if ordRts, err := orderRatesOnIntervals(allRates, wghts, sTime, usage.Big, true, 10); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(ordRts, expected) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(ordRts))
	}
}

func TestOrderRateOnIntervalsEveryThreeHours(t *testing.T) {
	rtEveryThreeH := &utils.Rate{
		ID:              "EVERY_THREE_RATE",
		ActivationTimes: "* */3 * * *",
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
		IntervalRates: []*utils.IntervalRate{
			{
				IntervalStart: utils.NewDecimal(0, 0),
			},
		},
	}
	rtByDay := &utils.Rate{
		ID:              "DAY_RATE",
		ActivationTimes: "* 15-23 * * *",
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
		IntervalRates: []*utils.IntervalRate{
			{
				IntervalStart: utils.NewDecimal(0, 0),
			},
		},
	}
	wghts := []float64{10, 20}
	allRates := []*utils.Rate{rtEveryThreeH, rtByDay}
	for _, idx := range allRates {
		if err := idx.Compile(); err != nil {
			t.Error(err)
		}
	}
	sTime := time.Date(2020, 7, 15, 0, 0, 0, 0, time.UTC)
	usage := utils.NewDecimal(int64(24*time.Hour), 0)
	expected := []*orderedRate{
		{
			utils.NewDecimal(0, 0).Big,
			rtEveryThreeH,
		},
		{
			utils.NewDecimal(int64(15*time.Hour), 0).Big,
			rtByDay,
		},
	}
	if ordRts, err := orderRatesOnIntervals(allRates, wghts, sTime, usage.Big, true, 10); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(ordRts, expected) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(ordRts))
	}
}

func TestOrderRateOnIntervalsTwoRatesInOne(t *testing.T) {
	rtHalfDay1 := &utils.Rate{
		ID:              "HALF_RATE1",
		ActivationTimes: "* 0-11 22 12 *",
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
		IntervalRates: []*utils.IntervalRate{
			{
				IntervalStart: utils.NewDecimal(0, 0),
			},
		},
	}
	rtHalfDay2 := &utils.Rate{
		ID:              "HALF_RATE2",
		ActivationTimes: "* 12-23 22 12 *",
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
		IntervalRates: []*utils.IntervalRate{
			{
				IntervalStart: utils.NewDecimal(0, 0),
			},
		},
	}
	rtHalfDay2r1 := &utils.Rate{
		ID:              "HALF_RATE2.1",
		ActivationTimes: "* 12-16 22 12 *",
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
		IntervalRates: []*utils.IntervalRate{
			{
				IntervalStart: utils.NewDecimal(0, 0),
			},
		},
	}
	rtHalfDay2r2 := &utils.Rate{
		ID:              "HALF_RATE2.2",
		ActivationTimes: "* 18-23 22 12 *",
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
		IntervalRates: []*utils.IntervalRate{
			{
				IntervalStart: utils.NewDecimal(0, 0),
			},
		},
	}
	wghts := []float64{10, 10, 20, 20}
	allRates := []*utils.Rate{rtHalfDay1, rtHalfDay2, rtHalfDay2r1, rtHalfDay2r2}
	for _, idx := range allRates {
		if err := idx.Compile(); err != nil {
			t.Error(err)
		}
	}
	sTime := time.Date(2020, 12, 21, 23, 0, 0, 0, time.UTC)
	usage := utils.NewDecimal(int64(25*time.Hour), 0)
	expected := []*orderedRate{
		{
			utils.NewDecimal(int64(time.Hour), 0).Big,
			rtHalfDay1,
		},
		{
			utils.NewDecimal(int64(13*time.Hour), 0).Big,
			rtHalfDay2r1,
		},
		{
			utils.NewDecimal(int64(18*time.Hour), 0).Big,
			rtHalfDay2,
		},
		{
			utils.NewDecimal(int64(19*time.Hour), 0).Big,
			rtHalfDay2r2,
		},
	}
	if ordRts, err := orderRatesOnIntervals(allRates, wghts, sTime, usage.Big, true, 10); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(ordRts, expected) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(ordRts))
	}
}

func TestOrderRateOnIntervalsEvery1Hour30Mins(t *testing.T) {
	rateEvery1H := &utils.Rate{
		ID:              "HOUR_RATE",
		ActivationTimes: "* */1 * * *",
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
		IntervalRates: []*utils.IntervalRate{
			{
				IntervalStart: utils.NewDecimal(0, 0),
			},
		},
	}
	rateEvery30Mins := &utils.Rate{
		ID:              "MINUTES_RATE",
		ActivationTimes: "*/30 * * * *",
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
		IntervalRates: []*utils.IntervalRate{
			{
				IntervalStart: utils.NewDecimal(0, 0),
			},
		},
	}
	wghts := []float64{10, 20}
	allRates := []*utils.Rate{rateEvery1H, rateEvery30Mins}
	for _, idx := range allRates {
		if err := idx.Compile(); err != nil {
			t.Error(err)
		}
	}
	sTime := time.Date(2020, 9, 20, 10, 0, 0, 0, time.UTC)
	usage := utils.NewDecimal(int64(time.Hour+time.Second), 0)
	expected := []*orderedRate{
		{
			utils.NewDecimal(0, 0).Big,
			rateEvery30Mins,
		},
		{
			utils.NewDecimal(int64(time.Minute), 0).Big,
			rateEvery1H,
		},
		{
			utils.NewDecimal(int64(30*time.Minute), 0).Big,
			rateEvery30Mins,
		},
		{
			utils.NewDecimal(int64(30*time.Minute+time.Minute), 0).Big,
			rateEvery1H,
		},
		{
			utils.NewDecimal(int64(time.Hour), 0).Big,
			rateEvery30Mins,
		},
	}
	if ordRts, err := orderRatesOnIntervals(allRates, wghts, sTime, usage.Big, true, 10); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(ordRts, expected) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(ordRts))
	}
}

func TestOrderRatesOnIntervalsOnePrinciapalRateCase1(t *testing.T) {
	rtPrincipal := &utils.Rate{
		ID:              "PRINCIPAL_RATE",
		ActivationTimes: "* 10-22 * * *",
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
		IntervalRates: []*utils.IntervalRate{
			{
				IntervalStart: utils.NewDecimal(0, 0),
			},
		},
	}
	rt1 := &utils.Rate{
		ID:              "RT1",
		ActivationTimes: "* 10-18 * * *",
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
		IntervalRates: []*utils.IntervalRate{
			{
				IntervalStart: utils.NewDecimal(0, 0),
			},
		},
	}
	rt2 := &utils.Rate{
		ID:              "RT2",
		ActivationTimes: "* 10-16 * * *",
		Weights: utils.DynamicWeights{
			{
				Weight: 30,
			},
		},
		IntervalRates: []*utils.IntervalRate{
			{
				IntervalStart: utils.NewDecimal(0, 0),
			},
		},
	}
	rt3 := &utils.Rate{
		ID:              "RT3",
		ActivationTimes: "* 10-14 * * *",
		Weights: utils.DynamicWeights{
			{
				Weight: 40,
			},
		},
		IntervalRates: []*utils.IntervalRate{
			{
				IntervalStart: utils.NewDecimal(0, 0),
			},
		},
	}
	wghts := []float64{10, 20, 30, 40}
	allRates := []*utils.Rate{rtPrincipal, rt1, rt2, rt3}
	for _, idx := range allRates {
		if err := idx.Compile(); err != nil {
			t.Error(err)
		}
	}
	sTime := time.Date(2020, 7, 21, 10, 0, 0, 0, time.UTC)
	usage := utils.NewDecimal(int64(13*time.Hour), 0)
	expected := []*orderedRate{
		{
			utils.NewDecimal(0, 0).Big,
			rt3,
		},
		{
			utils.NewDecimal(int64(5*time.Hour), 0).Big,
			rt2,
		},
		{
			utils.NewDecimal(int64(7*time.Hour), 0).Big,
			rt1,
		},
		{
			utils.NewDecimal(int64(9*time.Hour), 0).Big,
			rtPrincipal,
		},
	}
	if ordRts, err := orderRatesOnIntervals(allRates, wghts, sTime, usage.Big, true, 10); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(ordRts, expected) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(ordRts))
	}
}

func TestOrderRatesOnIntervalsOnePrinciapalRateCase2(t *testing.T) {
	rtPrincipal := &utils.Rate{
		ID:              "PRINCIPAL_RATE",
		ActivationTimes: "* 10-22 * * *",
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
		IntervalRates: []*utils.IntervalRate{
			{
				IntervalStart: utils.NewDecimal(0, 0),
			},
		},
	}
	rt1 := &utils.Rate{
		ID:              "RT1",
		ActivationTimes: "* 18-22 * * *",
		Weights: utils.DynamicWeights{
			{
				Weight: 40,
			},
		},
		IntervalRates: []*utils.IntervalRate{
			{
				IntervalStart: utils.NewDecimal(0, 0),
			},
		},
	}
	rt2 := &utils.Rate{
		ID:              "RT2",
		ActivationTimes: "* 16-22 * * *",
		Weights: utils.DynamicWeights{
			{
				Weight: 30,
			},
		},
		IntervalRates: []*utils.IntervalRate{
			{
				IntervalStart: utils.NewDecimal(0, 0),
			},
		},
	}
	rt3 := &utils.Rate{
		ID:              "RT3",
		ActivationTimes: "* 14-22 * * *",
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
		IntervalRates: []*utils.IntervalRate{
			{
				IntervalStart: utils.NewDecimal(0, 0),
			},
		},
	}
	wght := []float64{10, 40, 30, 20}
	allRates := []*utils.Rate{rtPrincipal, rt1, rt2, rt3}
	for _, idx := range allRates {
		if err := idx.Compile(); err != nil {
			t.Error(err)
		}
	}
	sTime := time.Date(2020, 7, 21, 10, 0, 0, 0, time.UTC)
	usage := utils.NewDecimal(int64(13*time.Hour), 0)
	expected := []*orderedRate{
		{
			utils.NewDecimal(0, 0).Big,
			rtPrincipal,
		},
		{
			utils.NewDecimal(int64(4*time.Hour), 0).Big,
			rt3,
		},
		{
			utils.NewDecimal(int64(6*time.Hour), 0).Big,
			rt2,
		},
		{
			utils.NewDecimal(int64(8*time.Hour), 0).Big,
			rt1,
		},
	}
	if ordRts, err := orderRatesOnIntervals(allRates, wght, sTime, usage.Big, true, 10); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(ordRts, expected) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(ordRts))
	}
}

func TestOrderRatesOnIntervalsEvenOddMinutes(t *testing.T) {
	rtOddMInutes := &utils.Rate{
		ID:              "ODD_RATE",
		ActivationTimes: "*/1 * * * *",
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
		IntervalRates: []*utils.IntervalRate{
			{
				IntervalStart: utils.NewDecimal(0, 0),
			},
		},
	}
	rtEvenMinutes := &utils.Rate{
		ID:              "EVEN_RATE",
		ActivationTimes: "*/2 * * * *",
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
		IntervalRates: []*utils.IntervalRate{
			{
				IntervalStart: utils.NewDecimal(0, 0),
			},
		},
	}
	wghts := []float64{10, 20}
	allRates := []*utils.Rate{rtOddMInutes, rtEvenMinutes}
	for _, idx := range allRates {
		if err := idx.Compile(); err != nil {
			t.Error(err)
		}
	}
	sTime := time.Date(2020, 12, 23, 22, 0, 0, 0, time.UTC)
	usage := utils.NewDecimal(int64(5*time.Minute+time.Second), 0)
	expected := []*orderedRate{
		{
			utils.NewDecimal(0, 0).Big,
			rtEvenMinutes,
		},
		{
			utils.NewDecimal(int64(time.Minute), 0).Big,
			rtOddMInutes,
		},
		{
			utils.NewDecimal(int64(2*time.Minute), 0).Big,
			rtEvenMinutes,
		},
		{
			utils.NewDecimal(int64(3*time.Minute), 0).Big,
			rtOddMInutes,
		},
		{
			utils.NewDecimal(int64(4*time.Minute), 0).Big,
			rtEvenMinutes,
		},
		{
			utils.NewDecimal(int64(5*time.Minute), 0).Big,
			rtOddMInutes,
		},
	}
	if ordRts, err := orderRatesOnIntervals(allRates, wghts, sTime, usage.Big, true, 10); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(ordRts, expected) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(ordRts))
	}
}

func TestOrderRatesOnIntervalsDoubleRates2(t *testing.T) {
	rt1 := &utils.Rate{
		ID: "ALWAYS_RATE",
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
		IntervalRates: []*utils.IntervalRate{
			{
				IntervalStart: utils.NewDecimal(0, 0),
			},
		},
	}
	rtCh1 := &utils.Rate{
		ID:              "CHRISTMAS1",
		ActivationTimes: "* * 24 12 *",
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
		IntervalRates: []*utils.IntervalRate{
			{
				IntervalStart: utils.NewDecimal(0, 0),
			},
		},
	}
	rtCh2 := &utils.Rate{
		ID:              "CHRISTMAS2",
		ActivationTimes: "* 10-12 24 12 *",
		Weights: utils.DynamicWeights{
			{
				Weight: 30,
			},
		},
		IntervalRates: []*utils.IntervalRate{
			{
				IntervalStart: utils.NewDecimal(0, 0),
			},
		},
	}
	rtCh3 := &utils.Rate{
		ID:              "CHRISTMAS3",
		ActivationTimes: "* 20-22 24 12 *",
		Weights: utils.DynamicWeights{
			{
				Weight: 30,
			},
		},
		IntervalRates: []*utils.IntervalRate{
			{
				IntervalStart: utils.NewDecimal(0, 0),
			},
		},
	}
	wghts := []float64{10, 20, 30, 30}
	allRates := []*utils.Rate{rt1, rtCh1, rtCh2, rtCh3}
	for _, idx := range allRates {
		if err := idx.Compile(); err != nil {
			t.Error(err)
		}
	}
	sTime := time.Date(2020, 12, 23, 22, 0, 0, 0, time.UTC)
	usage := utils.NewDecimal(int64(36*time.Hour), 0)
	expected := []*orderedRate{
		{
			utils.NewDecimal(0, 0).Big,
			rt1,
		},
		{
			utils.NewDecimal(int64(2*time.Hour), 0).Big,
			rtCh1,
		},
		{
			utils.NewDecimal(int64(12*time.Hour), 0).Big,
			rtCh2,
		},
		{
			utils.NewDecimal(int64(15*time.Hour), 0).Big,
			rtCh1,
		},
		{
			utils.NewDecimal(int64(22*time.Hour), 0).Big,
			rtCh3,
		},
		{
			utils.NewDecimal(int64(25*time.Hour), 0).Big,
			rtCh1,
		},
		{
			utils.NewDecimal(int64(26*time.Hour), 0).Big,
			rt1,
		},
	}
	if ordRts, err := orderRatesOnIntervals(allRates, wghts, sTime, usage.Big, true, 10); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(ordRts, expected) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(ordRts))
	}
}

func TestOrderOnRatesIntervalsEveryTwoHours(t *testing.T) {
	rt1 := &utils.Rate{
		ID: "ALWAYS_RATE",
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
		IntervalRates: []*utils.IntervalRate{
			{
				IntervalStart: utils.NewDecimal(0, 0),
			},
		},
	}
	rtEvTwoHours := &utils.Rate{
		ID: "EVERY_TWO_HOURS",
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
		ActivationTimes: "* */2 * * *",
		IntervalRates: []*utils.IntervalRate{
			{
				IntervalStart: utils.NewDecimal(0, 0),
			},
		},
	}
	wghts := []float64{10, 20}
	allRates := []*utils.Rate{rt1, rtEvTwoHours}
	for _, idx := range allRates {
		if err := idx.Compile(); err != nil {
			t.Error(err)
		}
	}
	sTime := time.Date(2020, 7, 21, 12, 10, 0, 0, time.UTC)
	usage := utils.NewDecimal(int64(4*time.Hour), 0)
	expected := []*orderedRate{
		{
			utils.NewDecimal(0, 0).Big,
			rtEvTwoHours,
		},
		{
			utils.NewDecimal(int64(50*time.Minute), 0).Big,
			rt1,
		},
		{
			utils.NewDecimal(int64(time.Hour+50*time.Minute), 0).Big,
			rtEvTwoHours,
		},
		{
			utils.NewDecimal(int64(2*time.Hour+50*time.Minute), 0).Big,
			rt1,
		},
		{
			utils.NewDecimal(int64(3*time.Hour+50*time.Minute), 0).Big,
			rtEvTwoHours,
		},
	}
	if ordRts, err := orderRatesOnIntervals(allRates, wghts, sTime, usage.Big, true, 10); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(ordRts, expected) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(ordRts))
	}
}

func TestOrderRatesOnIntervalsEveryTwoDays(t *testing.T) {
	rt1 := &utils.Rate{
		ID: "ALWAYS_RATE",
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
		IntervalRates: []*utils.IntervalRate{
			{
				IntervalStart: utils.NewDecimal(0, 0),
			},
		},
	}
	rtEveryTwoDays := &utils.Rate{
		ID:              "RATE_EVERY_DAY",
		ActivationTimes: "* * */2 * *",
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
		IntervalRates: []*utils.IntervalRate{
			{
				IntervalStart: utils.NewDecimal(0, 0),
			},
		},
	}
	wghts := []float64{10, 20}
	allRates := []*utils.Rate{rt1, rtEveryTwoDays}
	for _, idx := range allRates {
		if err := idx.Compile(); err != nil {
			t.Error(err)
		}
	}
	sTime := time.Date(2020, 7, 21, 23, 59, 59, 0, time.UTC)
	usage := utils.NewDecimal(int64(96*time.Hour+2*time.Second), 0)
	expected := []*orderedRate{
		{
			utils.NewDecimal(0, 0).Big,
			rtEveryTwoDays,
		},
		{
			utils.NewDecimal(int64(time.Second), 0).Big,
			rt1,
		},
		{
			utils.NewDecimal(int64(24*time.Hour+time.Second), 0).Big,
			rtEveryTwoDays,
		},
		{
			utils.NewDecimal(int64(48*time.Hour+time.Second), 0).Big,
			rt1,
		},
		{
			utils.NewDecimal(int64(72*time.Hour+time.Second), 0).Big,
			rtEveryTwoDays,
		},
		{
			utils.NewDecimal(int64(96*time.Hour+time.Second), 0).Big,
			rt1,
		},
	}
	if ordRts, err := orderRatesOnIntervals(allRates, wghts, sTime, usage.Big, true, 10); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(ordRts, expected) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(ordRts))
	}
}

func TestOrderRatesOnIntervalsSpecialHour(t *testing.T) {
	rtRestricted := &utils.Rate{
		ID:              "RESTRICTED",
		ActivationTimes: "* 10-22 * * *",
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
		IntervalRates: []*utils.IntervalRate{
			{
				IntervalStart: utils.NewDecimal(0, 0),
			},
		},
	}
	rtWayRestricted := &utils.Rate{
		ID:              "WAY_RESTRICTED",
		ActivationTimes: "* 12-14 * * *",
		Weights: utils.DynamicWeights{
			{
				Weight: 30,
			},
		},
		IntervalRates: []*utils.IntervalRate{
			{
				IntervalStart: utils.NewDecimal(0, 0),
			},
		},
	}
	rtSpecialHour := &utils.Rate{
		ID:              "SPECIAL_HOUR",
		ActivationTimes: "* 13 * * *",
		Weights: utils.DynamicWeights{
			{
				Weight: 40,
			},
		},
		IntervalRates: []*utils.IntervalRate{
			{
				IntervalStart: utils.NewDecimal(0, 0),
			},
		},
	}
	wghts := []float64{20, 30, 40}
	allRts := []*utils.Rate{rtRestricted, rtWayRestricted, rtSpecialHour}
	for _, idx := range allRts {
		if err := idx.Compile(); err != nil {
			t.Error(err)
		}
	}
	sTime := time.Date(2020, 7, 21, 9, 0, 0, 0, time.UTC)
	usage := utils.NewDecimal(int64(11*time.Hour), 0)
	expected := []*orderedRate{
		{
			utils.NewDecimal(int64(time.Hour), 0).Big,
			rtRestricted,
		},
		{
			utils.NewDecimal(int64(3*time.Hour), 0).Big,
			rtWayRestricted,
		},
		{
			utils.NewDecimal(int64(4*time.Hour), 0).Big,
			rtSpecialHour,
		},
		{
			utils.NewDecimal(int64(5*time.Hour), 0).Big,
			rtWayRestricted,
		},
		{
			utils.NewDecimal(int64(6*time.Hour), 0).Big,
			rtRestricted,
		},
	}
	if ordRts, err := orderRatesOnIntervals(allRts, wghts, sTime, usage.Big, true, 10); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(ordRts, expected) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(ordRts))
	}
}

func TestOrderRateIntervalsRateEveryTenMinutes(t *testing.T) {
	rt1 := &utils.Rate{
		ID:              "DAY_RATE",
		ActivationTimes: "* * 21 7 *",
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
		IntervalRates: []*utils.IntervalRate{
			{
				IntervalStart: utils.NewDecimal(0, 0),
			},
		},
	}
	rtEveryTenMin := &utils.Rate{
		ID:              "EVERY_TEN_MIN",
		ActivationTimes: "*/20 * * * *",
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
		IntervalRates: []*utils.IntervalRate{
			{
				IntervalStart: utils.NewDecimal(0, 0),
			},
		},
	}
	wghts := []float64{10, 20}
	allRts := []*utils.Rate{rt1, rtEveryTenMin}
	for _, idx := range allRts {
		if err := idx.Compile(); err != nil {
			t.Error(err)
		}
	}
	sTime := time.Date(2020, 7, 21, 10, 05, 0, 0, time.UTC)
	usage := utils.NewDecimal(int64(40*time.Minute), 0)
	expected := []*orderedRate{
		{
			utils.NewDecimal(0, 0).Big,
			rt1,
		},
		{
			utils.NewDecimal(int64(15*time.Minute), 0).Big,
			rtEveryTenMin,
		},
		{
			utils.NewDecimal(int64(16*time.Minute), 0).Big,
			rt1,
		},
		{
			utils.NewDecimal(int64(35*time.Minute), 0).Big,
			rtEveryTenMin,
		},
		{
			utils.NewDecimal(int64(36*time.Minute), 0).Big,
			rt1,
		},
	}
	if ordRts, err := orderRatesOnIntervals(allRts, wghts, sTime, usage.Big, true, 10); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(ordRts, expected) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(ordRts))
	}
}

func TestOrderRatesOnIntervalsDayOfTheWeek(t *testing.T) {
	rt1 := &utils.Rate{
		ID: "ALWAYS_RATE",
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
		IntervalRates: []*utils.IntervalRate{
			{
				IntervalStart: utils.NewDecimal(0, 0),
			},
		},
	}
	rtDay := &utils.Rate{
		ID:              "DAY_RATE",
		ActivationTimes: "* * 21 7 2",
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
		IntervalRates: []*utils.IntervalRate{
			{
				IntervalStart: utils.NewDecimal(0, 0),
			},
		},
	}
	rtDay1 := &utils.Rate{
		ID:              "DAY_RATE1",
		ActivationTimes: "* 15 21 7 2",
		Weights: utils.DynamicWeights{
			{
				Weight: 30,
			},
		},
		IntervalRates: []*utils.IntervalRate{
			{
				IntervalStart: utils.NewDecimal(0, 0),
			},
		},
	}
	rtDay2 := &utils.Rate{
		ID:              "DAY_RATE2",
		ActivationTimes: "* 18 21 7 2",
		Weights: utils.DynamicWeights{
			{
				Weight: 30,
			},
		},
		IntervalRates: []*utils.IntervalRate{
			{
				IntervalStart: utils.NewDecimal(0, 0),
			},
		},
	}
	wghts := []float64{10, 20, 30, 30}
	allRates := []*utils.Rate{rt1, rtDay, rtDay1, rtDay2}
	for _, idx := range allRates {
		if err := idx.Compile(); err != nil {
			t.Error(err)
		}
	}
	sTime := time.Date(2020, 7, 20, 23, 0, 0, 0, time.UTC)
	usage := utils.NewDecimal(int64(30*time.Hour), 0)
	expected := []*orderedRate{
		{
			utils.NewDecimal(0, 0).Big,
			rt1,
		},
		{
			utils.NewDecimal(int64(time.Hour), 0).Big,
			rtDay,
		},
		{
			utils.NewDecimal(int64(16*time.Hour), 0).Big,
			rtDay1,
		},
		{
			utils.NewDecimal(int64(17*time.Hour), 0).Big,
			rtDay,
		},
		{
			utils.NewDecimal(int64(19*time.Hour), 0).Big,
			rtDay2,
		},
		{
			utils.NewDecimal(int64(20*time.Hour), 0).Big,
			rtDay,
		},
		{
			utils.NewDecimal(int64(25*time.Hour), 0).Big,
			rt1,
		},
	}
	if ordRts, err := orderRatesOnIntervals(allRates, wghts, sTime, usage.Big, true, 10); err != nil {
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
	rt1 := &utils.Rate{
		ID:              "RT_1",
		ActivationTimes: "1 * * * *",
		IntervalRates: []*utils.IntervalRate{
			{
				IntervalStart: utils.NewDecimal(0, 0),
			},
		},
	}
	wghts := []float64{0}
	err := rt1.Compile()
	if err != nil {
		t.Error(err)
	}
	aRts := []*utils.Rate{rt1}
	sTime := time.Date(2020, 01, 02, 0, 1, 0, 0, time.UTC)
	usage := utils.NewDecimal(int64(96*time.Hour), 0)
	expectedErr := "maximum iterations reached"
	if _, err := orderRatesOnIntervals(aRts, wghts, sTime, usage.Big, false, 1); err == nil || err.Error() != expectedErr {
		t.Errorf("Expected %+v, received %+v", expectedErr, err)
	}
}

func TestOrderRatesOnIntervalIsDirectionFalse(t *testing.T) {
	rt1 := &utils.Rate{
		ID:              "RT_1",
		ActivationTimes: "* * 27 02 *",
		IntervalRates: []*utils.IntervalRate{
			{
				IntervalStart: utils.NewDecimal(0, 0),
			},
		},
	}
	err := rt1.Compile()
	if err != nil {
		t.Error(err)
	}
	expected := []*orderedRate{
		{
			utils.NewDecimal(0, 0).Big,
			rt1,
		},
	}
	err = expected[0].Rate.Compile()
	if err != nil {
		t.Error(err)
	}
	wghts := []float64{0}
	aRts := []*utils.Rate{rt1}
	sTime := time.Date(0001, 02, 27, 0, 0, 0, 0, time.UTC)
	usage := utils.NewDecimal(int64(48*time.Hour), 0)
	if ordRts, err := orderRatesOnIntervals(aRts, wghts, sTime, usage.Big, false, 5); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(ordRts, expected) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(expected), utils.ToJSON(ordRts))
	}
}

func TestOrderRatesOnIntervalWinnNill(t *testing.T) {
	rt1 := &utils.Rate{
		ID:              "RT_1",
		ActivationTimes: "* * 1 * *",
		IntervalRates: []*utils.IntervalRate{
			{
				IntervalStart: utils.NewDecimal(0, 0),
			},
		},
	}
	err := rt1.Compile()
	if err != nil {
		t.Error(err)
	}
	expected := []*orderedRate{
		{
			utils.NewDecimal(0, 0).Big,
			rt1,
		},
	}
	err = expected[0].Rate.Compile()
	if err != nil {
		t.Error(err)
	}
	wghts := []float64{0}
	aRts := []*utils.Rate{rt1}
	sTime := time.Date(2020, 12, 1, 0, 0, 0, 0, time.UTC)
	usage := utils.NewDecimal(int64(96*time.Hour), 0)
	if ordRts, err := orderRatesOnIntervals(aRts, wghts, sTime, usage.Big, true, 4); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(ordRts, expected) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(expected), utils.ToJSON(ordRts))
	}
}

func TestOrderRatesOnIntervalIntervalStartHigherThanEndIdx(t *testing.T) {
	rt1 := &utils.Rate{
		ID:              "RT_1",
		ActivationTimes: "* * 1 * *",
		IntervalRates: []*utils.IntervalRate{
			{
				IntervalStart: utils.NewDecimal(int64(48*time.Hour), 0),
			},
		},
	}
	err := rt1.Compile()
	if err != nil {
		t.Error(err)
	}
	expected := []*orderedRate{
		{
			utils.NewDecimal(0, 0).Big,
			rt1,
		},
	}
	err = expected[0].Rate.Compile()
	if err != nil {
		t.Error(err)
	}
	wghts := []float64{0}
	aRts := []*utils.Rate{rt1}
	sTime := time.Date(2020, 12, 1, 0, 0, 0, 0, time.UTC)
	usage := utils.NewDecimal(int64(48*time.Hour), 0)
	if _, err := orderRatesOnIntervals(aRts, wghts, sTime, usage.Big, false, 4); err != nil {
		t.Error(err)
	}
}

func TestOrderRatesOnIntervalStartLowerThanEndIdx(t *testing.T) {
	rt1 := &utils.Rate{
		ID:              "RT_1",
		ActivationTimes: "* * 1 * *",
		IntervalRates: []*utils.IntervalRate{
			{
				IntervalStart: utils.NewDecimal(int64(23*time.Hour), 0),
			},
			{
				IntervalStart: utils.NewDecimal(int64(-time.Hour), 0),
			},
		},
	}
	err := rt1.Compile()
	if err != nil {
		t.Error(err)
	}
	expected := []*orderedRate{
		{
			utils.NewDecimal(0, 0).Big,
			rt1,
		},
	}
	err = expected[0].Rate.Compile()
	if err != nil {
		t.Error(err)
	}
	aRts := []*utils.Rate{rt1}
	wghts := []float64{0}
	sTime := time.Date(2020, 12, 1, 0, 0, 0, 0, time.UTC)
	usage := utils.NewDecimal(int64(48*time.Hour), 0)
	if ordRts, err := orderRatesOnIntervals(aRts, wghts, sTime, usage.Big, false, 4); err != nil {
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
	rt0 := &utils.Rate{
		ID: "RATE0",
		IntervalRates: []*utils.IntervalRate{
			{
				IntervalStart: utils.NewDecimal(0, 0),
				RecurrentFee:  utils.NewDecimal(1, 0),
				Unit:          minDecimal,
				Increment:     minDecimal,
			},
			{
				IntervalStart: utils.NewDecimal(int64(60*time.Second), 0),
				RecurrentFee:  utils.NewDecimal(5, 3),
				Unit:          minDecimal,
				Increment:     secDecimal,
			},
		},
	}

	rt1 := &utils.Rate{
		ID: "RATE1",
		IntervalRates: []*utils.IntervalRate{
			{
				IntervalStart: utils.NewDecimal(0, 0),
				RecurrentFee:  utils.NewDecimal(2, 1),
				Unit:          minDecimal,
				Increment:     secDecimal,
			},
			{
				IntervalStart: utils.NewDecimal(int64(2*time.Minute), 0),
				RecurrentFee:  utils.NewDecimal(15, 2),
				Unit:          minDecimal,
				Increment:     secDecimal,
			},
		},
	}

	rp := &utils.RateProfile{
		Tenant: "cgrates.org",
		ID:     "RATE_PROFILE",
		Rates: map[string]*utils.Rate{
			rt0.ID: rt0,
			rt1.ID: rt1,
		},
	}
	if err := rp.Compile(); err != nil {
		t.Fatal(err)
	}

	rts := []*orderedRate{
		{
			utils.NewDecimal(0, 0).Big,
			rt0,
		},
		{
			utils.NewDecimal(int64(90*time.Second), 0).Big,
			rt1,
		},
	}

	eRtIvls := []*utils.RateSInterval{
		{
			IntervalStart: utils.NewDecimal(0, 0),
			Increments: []*utils.RateSIncrement{
				{
					IncrementStart:    utils.NewDecimal(0, 0),
					Usage:             utils.NewDecimal(int64(time.Minute), 0),
					RateIntervalIndex: 0,
					RateID:            "UUID00",
					CompressFactor:    1,
				},
				{
					IncrementStart:    utils.NewDecimal(int64(time.Minute), 0),
					Usage:             utils.NewDecimal(int64(30*time.Second), 0),
					RateIntervalIndex: 1,
					RateID:            "UUID01",
					CompressFactor:    30,
				},
			},
			CompressFactor: 1,
		},
		{
			IntervalStart: utils.NewDecimal(int64(90*time.Second), 0),
			Increments: []*utils.RateSIncrement{
				{
					IncrementStart:    utils.NewDecimal(int64(90*time.Second), 0),
					Usage:             utils.NewDecimal(int64(30*time.Second), 0),
					RateIntervalIndex: 0,
					RateID:            "UUID10",
					CompressFactor:    30,
				},
				{
					IncrementStart:    utils.NewDecimal(int64(2*time.Minute), 0),
					Usage:             utils.NewDecimal(int64(10*time.Second), 0),
					RateIntervalIndex: 1,
					RateID:            "UUID11",
					CompressFactor:    10,
				},
			},
			CompressFactor: 1,
		},
	}
	expCstRts := map[string]*utils.IntervalRate{
		"UUID00": {
			IntervalStart: utils.NewDecimal(0, 0),
			RecurrentFee:  utils.NewDecimal(1, 0),
			Unit:          utils.NewDecimal(int64(time.Minute), 0),
			Increment:     utils.NewDecimal(int64(time.Minute), 0),
		},
		"UUID01": {
			IntervalStart: utils.NewDecimal(int64(time.Minute), 0),
			RecurrentFee:  utils.NewDecimal(5, 3),
			Unit:          utils.NewDecimal(int64(time.Minute), 0),
			Increment:     utils.NewDecimal(int64(1*time.Second), 0),
		},
		"UUID10": {
			IntervalStart: utils.NewDecimal(0, 0),
			RecurrentFee:  utils.NewDecimal(2, 1),
			Unit:          utils.NewDecimal(int64(time.Minute), 0),
			Increment:     utils.NewDecimal(int64(1*time.Second), 0),
		},
		"UUID11": {
			IntervalStart: utils.NewDecimal(int64(2*time.Minute), 0),
			RecurrentFee:  utils.NewDecimal(15, 2),
			Unit:          utils.NewDecimal(int64(time.Minute), 0),
			Increment:     utils.NewDecimal(int64(1*time.Second), 0),
		},
	}
	cstRts := make(map[string]*utils.IntervalRate)
	if rtIvls, err := computeRateSIntervals(rts, utils.NewDecimal(0, 0).Big,
		utils.NewDecimal(int64(130*time.Second), 0).Big, cstRts); err != nil {
		t.Error(err)
	} else {
		for idx, val := range rtIvls {
			if !val.Equals(eRtIvls[idx], cstRts, expCstRts) {
				t.Errorf("expecting: %+v \n,received: %+v", utils.ToJSON(expCstRts), utils.ToJSON(cstRts))
			}
		}
	}

	rts = []*orderedRate{
		{
			utils.NewDecimal(0, 0).Big,
			rt0,
		},
		{
			utils.NewDecimal(int64(90*time.Second), 0).Big,
			rt1,
		},
	}

	eRtIvls = []*utils.RateSInterval{
		{
			IntervalStart: utils.NewDecimal(int64(time.Minute), 0),
			Increments: []*utils.RateSIncrement{
				{
					IncrementStart:    utils.NewDecimal(int64(time.Minute), 0),
					Usage:             utils.NewDecimal(int64(30*time.Second), 0),
					RateIntervalIndex: 1,
					RateID:            "UUID01",
					CompressFactor:    30,
				},
			},
			CompressFactor: 1,
		},
		{
			IntervalStart: utils.NewDecimal(int64(90*time.Second), 0),
			Increments: []*utils.RateSIncrement{
				{
					IncrementStart:    utils.NewDecimal(int64(90*time.Second), 0),
					Usage:             utils.NewDecimal(int64(30*time.Second), 0),
					RateIntervalIndex: 0,
					RateID:            "UUID10",
					CompressFactor:    30,
				},
				{
					IncrementStart:    utils.NewDecimal(int64(2*time.Minute), 0),
					Usage:             utils.NewDecimal(int64(10*time.Second), 0),
					RateIntervalIndex: 1,
					RateID:            "UUID11",
					CompressFactor:    10,
				},
			},
			CompressFactor: 1,
		},
	}
	expCstRts = map[string]*utils.IntervalRate{
		"UUID01": {
			IntervalStart: utils.NewDecimal(int64(time.Minute), 0),
			RecurrentFee:  utils.NewDecimal(5, 3),
			Unit:          utils.NewDecimal(int64(time.Minute), 0),
			Increment:     utils.NewDecimal(int64(1*time.Second), 0),
		},
		"UUID10": {
			IntervalStart: utils.NewDecimal(0, 0),
			RecurrentFee:  utils.NewDecimal(2, 1),
			Unit:          utils.NewDecimal(int64(time.Minute), 0),
			Increment:     utils.NewDecimal(int64(1*time.Second), 0),
		},
		"UUID11": {
			IntervalStart: utils.NewDecimal(int64(2*time.Minute), 0),
			RecurrentFee:  utils.NewDecimal(15, 2),
			Unit:          utils.NewDecimal(int64(time.Minute), 0),
			Increment:     utils.NewDecimal(int64(1*time.Second), 0),
		},
	}
	if rtIvls, err := computeRateSIntervals(rts, utils.NewDecimal(int64(time.Minute), 0).Big,
		utils.NewDecimal(int64(70*time.Second), 0).Big, cstRts); err != nil {
		t.Error(err)
	} else {
		for idx, val := range rtIvls {
			if !val.Equals(eRtIvls[idx], cstRts, expCstRts) {
				t.Errorf("expecting: %+v \n,received: %+v", utils.ToJSON(expCstRts), utils.ToJSON(cstRts))
				t.Fatalf("expecting: %+v \n,received: %+v", utils.ToJSON(eRtIvls), utils.ToJSON(rtIvls))
			}
		}
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
	rt0 := &utils.Rate{
		ID: "RATE0",
		IntervalRates: []*utils.IntervalRate{
			{
				IntervalStart: utils.NewDecimal(0, 0),
				RecurrentFee:  utils.NewDecimal(2, 1),
				Unit:          tsecDecimal,
				Increment:     tsecDecimal,
			},
			{
				IntervalStart: utils.NewDecimal(int64(30*time.Second), 0),
				RecurrentFee:  utils.NewDecimal(15, 2),
				Unit:          minDecimal,
				Increment:     secDecimal,
			},
		},
	}

	rt1 := &utils.Rate{
		ID: "RATE1",
		IntervalRates: []*utils.IntervalRate{
			{
				IntervalStart: utils.NewDecimal(0, 0),
				RecurrentFee:  utils.NewDecimal(2, 1),
				Unit:          minDecimal,
				Increment:     secDecimal,
			},
			{
				IntervalStart: utils.NewDecimal(int64(2*time.Minute), 0),
				RecurrentFee:  utils.NewDecimal(15, 2),
				Unit:          minDecimal,
				Increment:     secDecimal,
			},
		},
	}

	rp := &utils.RateProfile{
		Tenant: "cgrates.org",
		ID:     "RATE_PROFILE",
		Rates: map[string]*utils.Rate{
			rt0.ID: rt0,
			rt1.ID: rt1,
		},
	}
	if err := rp.Compile(); err != nil {
		t.Fatal(err)
	}

	ordRts := []*orderedRate{
		{
			utils.NewDecimal(0, 0).Big,
			rt0,
		},
		{
			utils.NewDecimal(int64(time.Minute+10*time.Second), 0).Big,
			rt1,
		},
	}

	eRtIvls := []*utils.RateSInterval{
		{
			IntervalStart: utils.NewDecimal(int64(30*time.Second), 0),
			Increments: []*utils.RateSIncrement{
				{
					IncrementStart:    utils.NewDecimal(int64(30*time.Second), 0),
					Usage:             utils.NewDecimal(int64(40*time.Second), 0),
					RateIntervalIndex: 1,
					RateID:            "UUID01",
					CompressFactor:    40,
				},
			},
			CompressFactor: 1,
		},
		{
			IntervalStart: utils.NewDecimal(int64(time.Minute+10*time.Second), 0),
			Increments: []*utils.RateSIncrement{
				{
					IncrementStart:    utils.NewDecimal(int64(time.Minute+10*time.Second), 0),
					Usage:             utils.NewDecimal(int64(50*time.Second), 0),
					RateIntervalIndex: 0,
					RateID:            "UUID10",
					CompressFactor:    50,
				},
				{
					IncrementStart:    utils.NewDecimal(int64(2*time.Minute), 0),
					Usage:             utils.NewDecimal(int64(90*time.Second), 0),
					RateIntervalIndex: 1,
					RateID:            "UUID11",
					CompressFactor:    90,
				},
			},
			CompressFactor: 1,
		},
	}
	expCstRts := map[string]*utils.IntervalRate{
		"UUID01": {
			IntervalStart: utils.NewDecimal(int64(30*time.Second), 0),
			RecurrentFee:  utils.NewDecimal(15, 2),
			Unit:          utils.NewDecimal(int64(time.Minute), 0),
			Increment:     utils.NewDecimal(int64(time.Second), 0),
		},
		"UUID10": {
			IntervalStart: utils.NewDecimal(0, 0),
			RecurrentFee:  utils.NewDecimal(2, 1),
			Unit:          utils.NewDecimal(int64(time.Minute), 0),
			Increment:     utils.NewDecimal(int64(1*time.Second), 0),
		},
		"UUID11": {
			IntervalStart: utils.NewDecimal(int64(2*time.Minute), 0),
			RecurrentFee:  utils.NewDecimal(15, 2),
			Unit:          utils.NewDecimal(int64(time.Minute), 0),
			Increment:     utils.NewDecimal(int64(1*time.Second), 0),
		},
	}
	cstRts := make(map[string]*utils.IntervalRate)
	if rtIvls, err := computeRateSIntervals(ordRts, utils.NewDecimal(int64(30*time.Second), 0).Big,
		utils.NewDecimal(int64(3*time.Minute), 0).Big, cstRts); err != nil {
		t.Error(err)
	} else {
		for idx, val := range rtIvls {
			if !val.Equals(eRtIvls[idx], cstRts, expCstRts) {
				t.Errorf("expecting: %+v \n,received: %+v", utils.ToJSON(expCstRts), utils.ToJSON(cstRts))
				t.Fatalf("expecting: %+v \n,received: %+v", utils.ToJSON(eRtIvls), utils.ToJSON(rtIvls))
			}
		}
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
	rt0 := &utils.Rate{
		ID: "RATE0",
		IntervalRates: []*utils.IntervalRate{
			{
				IntervalStart: utils.NewDecimal(0, 0),
				FixedFee:      utils.NewDecimal(123, 3),
				RecurrentFee:  utils.NewDecimal(2, 1),
				Unit:          tsecDecimal,
				Increment:     tsecDecimal,
			},
			{
				IntervalStart: utils.NewDecimal(int64(30*time.Second), 0),
				RecurrentFee:  utils.NewDecimal(15, 2),
				Unit:          minDecimal,
				Increment:     secDecimal,
			},
		},
	}

	rt1 := &utils.Rate{
		ID: "RATE1",
		IntervalRates: []*utils.IntervalRate{
			{
				IntervalStart: utils.NewDecimal(0, 0),
				FixedFee:      utils.NewDecimal(567, 3),
				RecurrentFee:  utils.NewDecimal(2, 1),
				Unit:          minDecimal,
				Increment:     secDecimal,
			},
			{
				IntervalStart: utils.NewDecimal(int64(2*time.Minute), 0),
				RecurrentFee:  utils.NewDecimal(15, 2),
				Unit:          minDecimal,
				Increment:     secDecimal,
			},
		},
	}

	rp := &utils.RateProfile{
		Tenant: "cgrates.org",
		ID:     "RATE_PROFILE",
		Rates: map[string]*utils.Rate{
			rt0.ID: rt0,
			rt1.ID: rt1,
		},
	}
	if err := rp.Compile(); err != nil {
		t.Fatal(err)
	}

	ordRts := []*orderedRate{
		{
			utils.NewDecimal(0, 0).Big,
			rt0,
		},
		{
			utils.NewDecimal(int64(time.Minute+10*time.Second), 0).Big,
			rt1,
		},
	}

	eRtIvls := []*utils.RateSInterval{
		{
			IntervalStart: utils.NewDecimal(0, 0),
			Increments: []*utils.RateSIncrement{
				{
					IncrementStart:    utils.NewDecimal(0, 0),
					RateIntervalIndex: 0,
					RateID:            "UUID00",
					CompressFactor:    1,
					Usage:             utils.NewDecimal(-1, 0),
				},
				{
					IncrementStart:    utils.NewDecimal(0, 0),
					RateIntervalIndex: 0,
					RateID:            "UUID00",
					CompressFactor:    1,
					Usage:             utils.NewDecimal(int64(30*time.Second), 0),
				},
				{
					IncrementStart:    utils.NewDecimal(int64(30*time.Second), 0),
					RateIntervalIndex: 1,
					RateID:            "UUID01",
					CompressFactor:    40,
					Usage:             utils.NewDecimal(int64(40*time.Second), 0),
				},
			},
			CompressFactor: 1,
		},
		{
			IntervalStart: utils.NewDecimal(int64(time.Minute+10*time.Second), 0),
			Increments: []*utils.RateSIncrement{
				{
					IncrementStart:    utils.NewDecimal(int64(time.Minute+10*time.Second), 0),
					RateIntervalIndex: 0,
					RateID:            "UUID10",
					CompressFactor:    1,
					Usage:             utils.NewDecimal(-1, 0),
				},
				{
					IncrementStart:    utils.NewDecimal(int64(time.Minute+10*time.Second), 0),
					RateIntervalIndex: 0,
					RateID:            "UUID10",
					CompressFactor:    50,
					Usage:             utils.NewDecimal(int64(50*time.Second), 0),
				},
				{
					IncrementStart:    utils.NewDecimal(int64(2*time.Minute), 0),
					RateIntervalIndex: 1,
					RateID:            "UUID11",
					CompressFactor:    60,
					Usage:             utils.NewDecimal(int64(60*time.Second), 0),
				},
			},
			CompressFactor: 1,
		},
	}
	expCstRts := map[string]*utils.IntervalRate{
		"UUID00": {
			IntervalStart: utils.NewDecimal(0, 0),
			FixedFee:      utils.NewDecimal(123, 3),
			RecurrentFee:  utils.NewDecimal(2, 1),
			Unit:          tsecDecimal,
			Increment:     tsecDecimal,
		},
		"UUID01": {
			IntervalStart: utils.NewDecimal(int64(30*time.Second), 0),
			RecurrentFee:  utils.NewDecimal(15, 2),
			Unit:          minDecimal,
			Increment:     secDecimal,
		},
		"UUID10": {
			IntervalStart: utils.NewDecimal(0, 0),
			FixedFee:      utils.NewDecimal(567, 3),
			RecurrentFee:  utils.NewDecimal(2, 1),
			Unit:          minDecimal,
			Increment:     secDecimal,
		},
		"UUID11": {
			IntervalStart: utils.NewDecimal(int64(2*time.Minute), 0),
			RecurrentFee:  utils.NewDecimal(15, 2),
			Unit:          minDecimal,
			Increment:     secDecimal,
		},
	}
	cstRts := make(map[string]*utils.IntervalRate)
	if rtIvls, err := computeRateSIntervals(ordRts, utils.NewDecimal(0, 0).Big,
		utils.NewDecimal(int64(3*time.Minute), 0).Big, cstRts); err != nil {
		t.Error(err)
	} else {
		for idx, val := range rtIvls {
			if !val.Equals(eRtIvls[idx], cstRts, expCstRts) {
				t.Errorf("expecting: %+v \n,received: %+v", utils.ToJSON(expCstRts), utils.ToJSON(cstRts))
				t.Fatalf("expecting: %+v \n,received: %+v", utils.ToJSON(eRtIvls), utils.ToJSON(rtIvls))
			}
		}
	}
}

func TestComputeRateSIntervals2(t *testing.T) {
	minDecimal, err := utils.NewDecimalFromUsage("1m")
	if err != nil {
		t.Error(err)
	}
	rt0 := &utils.Rate{
		ID:              "RATE0",
		ActivationTimes: "* * * * *",
		IntervalRates: []*utils.IntervalRate{
			{
				IntervalStart: utils.NewDecimal(0, 0),
				RecurrentFee:  utils.NewDecimal(1, 0),
				Unit:          minDecimal,
				Increment:     minDecimal,
			},
			{
				IntervalStart: utils.NewDecimal(int64(50*time.Minute), 0),
				RecurrentFee:  utils.NewDecimal(5, 1),
				Unit:          minDecimal,
				Increment:     minDecimal,
			},
		},
	}
	rt0.Compile()

	rt1 := &utils.Rate{
		ID:              "RATE1",
		ActivationTimes: "45-49 * * * *",
		IntervalRates: []*utils.IntervalRate{
			{
				IntervalStart: utils.NewDecimal(0, 0),
				RecurrentFee:  utils.NewDecimal(15, 2),
				Unit:          minDecimal,
				Increment:     minDecimal,
			},
			{
				IntervalStart: utils.NewDecimal(int64(45*time.Minute), 0),
				RecurrentFee:  utils.NewDecimal(2, 1),
				Unit:          minDecimal,
				Increment:     minDecimal,
			},
		},
	}
	rt1.Compile()
	wghts := []float64{0, 0}
	allRates := []*utils.Rate{rt0, rt1}

	ordRts := []*orderedRate{
		{
			utils.NewDecimal(0, 0).Big,
			rt0,
		},
		{
			utils.NewDecimal(int64(45*time.Minute), 0).Big,
			rt1,
		},
		{
			utils.NewDecimal(int64(50*time.Minute), 0).Big,
			rt0,
		},
	}

	eRtIvls := []*utils.RateSInterval{
		{
			IntervalStart: utils.NewDecimal(0, 0),
			Increments: []*utils.RateSIncrement{
				{
					IncrementStart:    utils.NewDecimal(0, 0),
					RateIntervalIndex: 0,
					RateID:            "UUID00",
					CompressFactor:    45,
					Usage:             utils.NewDecimal(int64(45*time.Minute), 0),
				},
			},
			CompressFactor: 1,
		},
		{
			IntervalStart: utils.NewDecimal(int64(45*time.Minute), 0),
			Increments: []*utils.RateSIncrement{
				{
					IncrementStart:    utils.NewDecimal(int64(45*time.Minute), 0),
					RateIntervalIndex: 1,
					RateID:            "UUID10",
					CompressFactor:    5,
					Usage:             utils.NewDecimal(int64(5*time.Minute), 0),
				},
			},
			CompressFactor: 1,
		},
		{
			IntervalStart: utils.NewDecimal(int64(50*time.Minute), 0),
			Increments: []*utils.RateSIncrement{
				{
					IncrementStart:    utils.NewDecimal(int64(50*time.Minute), 0),
					RateIntervalIndex: 1,
					RateID:            "UUID01",
					CompressFactor:    10,
					Usage:             utils.NewDecimal(int64(10*time.Minute), 0),
				},
			},
			CompressFactor: 1,
		},
	}
	cstRts := make(map[string]*utils.IntervalRate)
	expCstRts := map[string]*utils.IntervalRate{
		"UUID00": {
			IntervalStart: utils.NewDecimal(0, 0),
			RecurrentFee:  utils.NewDecimal(1, 0),
			Unit:          minDecimal,
			Increment:     minDecimal,
		},
		"UUID10": {
			IntervalStart: utils.NewDecimal(int64(45*time.Minute), 0),
			RecurrentFee:  utils.NewDecimal(2, 1),
			Unit:          minDecimal,
			Increment:     minDecimal,
		},
		"UUID01": {
			IntervalStart: utils.NewDecimal(int64(50*time.Minute), 0),
			RecurrentFee:  utils.NewDecimal(5, 1),
			Unit:          minDecimal,
			Increment:     minDecimal,
		},
	}
	sTime := time.Date(2020, 7, 21, 0, 0, 0, 0, time.UTC)
	usage := utils.NewDecimal(int64(time.Hour), 0)
	if rcvOrdRts, err := orderRatesOnIntervals(allRates, wghts, sTime, usage.Big, true, 10); err != nil {
		t.Error(eRtIvls)
	} else if !reflect.DeepEqual(ordRts, rcvOrdRts) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(ordRts), utils.ToJSON(rcvOrdRts))
	} else if rcveRtIvls, err := computeRateSIntervals(rcvOrdRts, utils.NewDecimal(0, 0).Big,
		usage.Big, cstRts); err != nil {
		t.Error(err)
	} else {
		for idx, val := range rcveRtIvls {
			if !val.Equals(eRtIvls[idx], cstRts, expCstRts) {
				t.Errorf("expecting: %+v \n,received: %+v", utils.ToJSON(expCstRts), utils.ToJSON(cstRts))
				t.Fatalf("expecting: %+v \n,received: %+v", utils.ToJSON(eRtIvls), utils.ToJSON(rcveRtIvls))
			}
		}
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
	rt1 := &utils.Rate{
		ID: "RATE1",
		IntervalRates: []*utils.IntervalRate{
			{
				IntervalStart: utils.NewDecimal(0, 0),
				RecurrentFee:  utils.NewDecimal(1, 1),
				Unit:          tsecDecimal,
				Increment:     secDecimal,
			},
			{
				IntervalStart: utils.NewDecimal(int64(time.Minute), 0),
				RecurrentFee:  utils.NewDecimal(2, 1),
				Unit:          tsecDecimal,
				Increment:     secDecimal,
			},
			{
				IntervalStart: utils.NewDecimal(int64(2*time.Minute), 0),
				RecurrentFee:  utils.NewDecimal(3, 1),
				Unit:          tsecDecimal,
				Increment:     secDecimal,
			},
		},
	}

	rt2 := &utils.Rate{
		ID: "RATE2",
		IntervalRates: []*utils.IntervalRate{
			{
				IntervalStart: utils.NewDecimal(int64(30*time.Second), 0),
				RecurrentFee:  utils.NewDecimal(15, 2),
				Unit:          tsecDecimal,
				Increment:     secDecimal,
			},
			{
				IntervalStart: utils.NewDecimal(int64(time.Minute+30*time.Second), 0),
				RecurrentFee:  utils.NewDecimal(25, 2),
				Unit:          tsecDecimal,
				Increment:     secDecimal,
			},
			{
				IntervalStart: utils.NewDecimal(int64(2*time.Minute+30*time.Second), 0),
				RecurrentFee:  utils.NewDecimal(35, 2),
				Unit:          tsecDecimal,
				Increment:     secDecimal,
			},
		},
	}

	rp := &utils.RateProfile{
		Tenant: "cgrates.org",
		ID:     "RATE_PROFILE",
		Rates: map[string]*utils.Rate{
			rt1.ID: rt1,
			rt2.ID: rt2,
		},
	}
	if err := rp.Compile(); err != nil {
		t.Fatal(err)
	}

	ordRts := []*orderedRate{
		{
			utils.NewDecimal(0, 0).Big,
			rt1,
		},
		{
			utils.NewDecimal(int64(30*time.Second), 0).Big,
			rt2,
		},
		{
			utils.NewDecimal(int64(time.Minute), 0).Big,
			rt1,
		},
		{
			utils.NewDecimal(int64(time.Minute+30*time.Second), 0).Big,
			rt2,
		},
		{
			utils.NewDecimal(int64(2*time.Minute), 0).Big,
			rt1,
		},
		{
			utils.NewDecimal(int64(2*time.Minute+30*time.Second), 0).Big,
			rt2,
		},
	}

	expOrdRates := []*utils.RateSInterval{
		{
			IntervalStart: utils.NewDecimal(0, 0),
			Increments: []*utils.RateSIncrement{
				{
					IncrementStart:    utils.NewDecimal(0, 0),
					RateIntervalIndex: 0,
					RateID:            "UUID00",
					CompressFactor:    30,
					Usage:             utils.NewDecimal(int64(30*time.Second), 0),
				},
			},
			CompressFactor: 1,
		},
		{
			IntervalStart: utils.NewDecimal(int64(30*time.Second), 0),
			Increments: []*utils.RateSIncrement{
				{
					IncrementStart:    utils.NewDecimal(int64(30*time.Second), 0),
					RateIntervalIndex: 0,
					RateID:            "UUID10",
					CompressFactor:    30,
					Usage:             utils.NewDecimal(int64(30*time.Second), 0),
				},
			},
			CompressFactor: 1,
		},
		{
			IntervalStart: utils.NewDecimal(int64(time.Minute), 0),
			Increments: []*utils.RateSIncrement{
				{
					IncrementStart:    utils.NewDecimal(int64(time.Minute), 0),
					RateIntervalIndex: 1,
					RateID:            "UUID01",
					CompressFactor:    30,
					Usage:             utils.NewDecimal(int64(30*time.Second), 0),
				},
			},
			CompressFactor: 1,
		},
		{
			IntervalStart: utils.NewDecimal(int64(time.Minute+30*time.Second), 0),
			Increments: []*utils.RateSIncrement{
				{
					IncrementStart:    utils.NewDecimal(int64(time.Minute+30*time.Second), 0),
					RateIntervalIndex: 1,
					RateID:            "UUID11",
					CompressFactor:    30,
					Usage:             utils.NewDecimal(int64(30*time.Second), 0),
				},
			},
			CompressFactor: 1,
		},
		{
			IntervalStart: utils.NewDecimal(int64(2*time.Minute), 0),
			Increments: []*utils.RateSIncrement{
				{
					IncrementStart:    utils.NewDecimal(int64(2*time.Minute), 0),
					RateIntervalIndex: 2,
					RateID:            "UUID02",
					CompressFactor:    30,
					Usage:             utils.NewDecimal(int64(30*time.Second), 0),
				},
			},
			CompressFactor: 1,
		},
		{
			IntervalStart: utils.NewDecimal(int64(2*time.Minute+30*time.Second), 0),
			Increments: []*utils.RateSIncrement{
				{
					IncrementStart:    utils.NewDecimal(int64(2*time.Minute+30*time.Second), 0),
					RateIntervalIndex: 2,
					RateID:            "UUID12",
					CompressFactor:    30,
					Usage:             utils.NewDecimal(int64(30*time.Second), 0),
				},
			},
			CompressFactor: 1,
		},
	}
	expCstRts := map[string]*utils.IntervalRate{
		"UUID00": {
			IntervalStart: utils.NewDecimal(0, 0),
			RecurrentFee:  utils.NewDecimal(1, 1),
			Unit:          tsecDecimal,
			Increment:     secDecimal,
		},
		"UUID10": {
			IntervalStart: utils.NewDecimal(int64(30*time.Second), 0),
			RecurrentFee:  utils.NewDecimal(15, 2),
			Unit:          tsecDecimal,
			Increment:     secDecimal,
		},
		"UUID01": {
			IntervalStart: utils.NewDecimal(int64(time.Minute), 0),
			RecurrentFee:  utils.NewDecimal(2, 1),
			Unit:          tsecDecimal,
			Increment:     secDecimal,
		},
		"UUID11": {
			IntervalStart: utils.NewDecimal(int64(time.Minute+30*time.Second), 0),
			RecurrentFee:  utils.NewDecimal(25, 2),
			Unit:          tsecDecimal,
			Increment:     secDecimal,
		},
		"UUID02": {
			IntervalStart: utils.NewDecimal(int64(2*time.Minute), 0),
			RecurrentFee:  utils.NewDecimal(3, 1),
			Unit:          tsecDecimal,
			Increment:     secDecimal,
		},
		"UUID12": {
			IntervalStart: utils.NewDecimal(int64(2*time.Minute+30*time.Second), 0),
			RecurrentFee:  utils.NewDecimal(35, 2),
			Unit:          tsecDecimal,
			Increment:     secDecimal,
		},
	}
	cstRts := make(map[string]*utils.IntervalRate)
	if rtIvls, err := computeRateSIntervals(ordRts, utils.NewDecimal(0, 0).Big,
		utils.NewDecimal(int64(3*time.Minute), 0).Big, cstRts); err != nil {
		t.Error(err)
	} else {
		for idx, val := range rtIvls {
			if !val.Equals(expOrdRates[idx], cstRts, expCstRts) {
				t.Errorf("expecting: %+v \n,received: %+v", utils.ToJSON(expCstRts), utils.ToJSON(cstRts))
				t.Fatalf("expecting: %+v \n,received: %+v", utils.ToJSON(expOrdRates), utils.ToJSON(rtIvls))
			}
		}
	}
}

func TestComputeRateSIntervalsStartHigherThanUsage(t *testing.T) {
	tsecDecimal, err := utils.NewDecimalFromUsage("30s")
	if err != nil {
		t.Error(err)
	}
	rt1 := &utils.Rate{
		ID: "RATE1",
		IntervalRates: []*utils.IntervalRate{
			{
				IntervalStart: utils.NewDecimal(int64(time.Minute), 0),
				RecurrentFee:  utils.NewDecimal(1, 1),
				Unit:          tsecDecimal,
				Increment:     tsecDecimal,
			},
		},
	}

	rt2 := &utils.Rate{
		ID: "RATE2",
		IntervalRates: []*utils.IntervalRate{
			{
				IntervalStart: utils.NewDecimal(int64(2*time.Minute), 0),
				RecurrentFee:  utils.NewDecimal(2, 1),
				Unit:          tsecDecimal,
				Increment:     tsecDecimal,
			},
		},
	}

	rp := &utils.RateProfile{
		Tenant: "cgrates.org",
		ID:     "RATE_PROFILE",
		Rates: map[string]*utils.Rate{
			rt1.ID: rt1,
			rt2.ID: rt2,
		},
	}
	if err := rp.Compile(); err != nil {
		t.Fatal(err)
	}

	ordRts := []*orderedRate{
		{
			utils.NewDecimal(0, 0).Big,
			rt1,
		},
	}
	cstRts := make(map[string]*utils.IntervalRate)
	expected := "intervalStart for rate: <cgrates.org:RATE_PROFILE:RATE1> higher than usage: 0"
	if _, err := computeRateSIntervals(ordRts, utils.NewDecimal(0, 0).Big,
		utils.NewDecimal(int64(3*time.Minute), 0).Big, cstRts); err == nil || err.Error() != expected {
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
	rt1 := &utils.Rate{
		ID: "RATE1",
		IntervalRates: []*utils.IntervalRate{
			{
				IntervalStart: utils.NewDecimal(0, 0),
				RecurrentFee:  utils.NewDecimal(1, 1),
				Unit:          tsecDecimal,
				Increment:     zeroDecimal,
			},
		},
	}
	rt1.Compile()

	ordRts := []*orderedRate{
		{
			utils.NewDecimal(int64(0*time.Second), 0).Big,
			rt1,
		},
	}

	cstRts := make(map[string]*utils.IntervalRate)
	expected := "zero increment to be charged within rate: <>"
	if _, err := computeRateSIntervals(ordRts, utils.NewDecimal(int64(33*time.Second), 0).Big,
		utils.NewDecimal(int64(3*time.Minute), 0).Big, cstRts); err == nil || err.Error() != expected {
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
	rt1 := &utils.Rate{
		ID: "RATE1",
		IntervalRates: []*utils.IntervalRate{
			{
				IntervalStart: utils.NewDecimal(0, 0),
				RecurrentFee:  utils.NewDecimal(1, 1),
				Unit:          tsecDecimal,
				Increment:     secDecimal,
			},
			{
				IntervalStart: utils.NewDecimal(int64(30*time.Second), 0),
				RecurrentFee:  utils.NewDecimal(25, 3),
				Unit:          minDecimal,
				Increment:     ssecDecimal,
			},
		},
	}
	rt1.Compile()

	ordRts := []*orderedRate{
		{
			utils.NewDecimal(0, 0).Big,
			rt1,
		},
	}
	expOrdRts := []*utils.RateSInterval{
		{
			IntervalStart: utils.NewDecimal(0, 0),
			Increments: []*utils.RateSIncrement{
				{
					IncrementStart:    utils.NewDecimal(0, 0),
					RateIntervalIndex: 0,
					RateID:            "UUID00",
					Usage:             utils.NewDecimal(int64(30*time.Second), 0),
					CompressFactor:    30,
				},
				{
					IncrementStart:    utils.NewDecimal(int64(30*time.Second), 0),
					RateIntervalIndex: 1,
					RateID:            "UUID01",
					Usage:             utils.NewDecimal(int64(40*time.Second), 0),
					CompressFactor:    6,
				},
			},
			CompressFactor: 1,
		},
	}
	expCstRts := map[string]*utils.IntervalRate{
		"UUID00": {
			IntervalStart: utils.NewDecimal(0, 0),
			RecurrentFee:  utils.NewDecimal(1, 1),
			Unit:          tsecDecimal,
			Increment:     secDecimal,
		},
		"UUID01": {
			IntervalStart: utils.NewDecimal(int64(30*time.Second), 0),
			RecurrentFee:  utils.NewDecimal(25, 3),
			Unit:          minDecimal,
			Increment:     ssecDecimal,
		},
	}
	cstRts := make(map[string]*utils.IntervalRate)
	if rtIvls, err := computeRateSIntervals(ordRts, utils.NewDecimal(0, 0).Big,
		utils.NewDecimal(int64(time.Minute+10*time.Second), 0).Big, cstRts); err != nil {
		t.Error(err)
	} else {
		for idx, val := range rtIvls {
			if !val.Equals(expOrdRts[idx], cstRts, expCstRts) {
				t.Errorf("expecting: %+v \n,received: %+v", utils.ToJSON(expCstRts), utils.ToJSON(cstRts))
				t.Fatalf("expecting: %+v \n,received: %+v", utils.ToJSON(expCstRts), utils.ToJSON(rtIvls))
			}
		}
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
	rt1 := &utils.Rate{
		ID: "RATE1",
		IntervalRates: []*utils.IntervalRate{
			{
				IntervalStart: utils.NewDecimal(0, 0),
				RecurrentFee:  utils.NewDecimal(1, 1),
				Unit:          tsecDecimal,
				Increment:     secDecimal,
			},
			{
				IntervalStart: utils.NewDecimal(int64(25*time.Second), 0),
				RecurrentFee:  utils.NewDecimal(25, 2),
				Unit:          minDecimal,
				Increment:     ssecDecimal,
			},
		},
	}

	rt2 := &utils.Rate{
		ID: "RATE2",
		IntervalRates: []*utils.IntervalRate{
			{
				IntervalStart: utils.NewDecimal(int64(45*time.Second), 0),
				RecurrentFee:  utils.NewDecimal(15, 2),
				Unit:          tsecDecimal,
				Increment:     secDecimal,
			},
			{
				IntervalStart: utils.NewDecimal(int64(55*time.Second), 0),
				RecurrentFee:  utils.NewDecimal(3, 1),
				Unit:          tsecDecimal,
				Increment:     fsecDecimal,
			},
		},
	}

	rt3 := &utils.Rate{
		ID: "RATE3",
		IntervalRates: []*utils.IntervalRate{
			{
				IntervalStart: utils.NewDecimal(int64(30*time.Second), 0),
				RecurrentFee:  utils.NewDecimal(1, 1),
				Unit:          fssecDecimal,
				Increment:     fsecDecimal,
			},
			{
				IntervalStart: utils.NewDecimal(int64(time.Minute), 0),
				RecurrentFee:  utils.NewDecimal(5, 3),
				Unit:          tsecDecimal,
				Increment:     fsecDecimal,
			},
		},
	}

	rp := &utils.RateProfile{
		Tenant: "cgrates.org",
		ID:     "RATE_PROFILE",
		Rates: map[string]*utils.Rate{
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
			utils.NewDecimal(0, 0).Big,
			rt1,
		},
		{
			utils.NewDecimal(int64(35*time.Second), 0).Big,
			rt3,
		},
		{
			utils.NewDecimal(int64(46*time.Second), 0).Big,
			rt2,
		},
		{
			utils.NewDecimal(int64(time.Minute), 0).Big,
			rt3,
		},
	}

	expOrdRts := []*utils.RateSInterval{
		{
			IntervalStart: utils.NewDecimal(0, 0),
			Increments: []*utils.RateSIncrement{
				{
					IncrementStart:    utils.NewDecimal(0, 0),
					RateIntervalIndex: 0,
					RateID:            "UUID00",
					Usage:             utils.NewDecimal(int64(25*time.Second), 0),
					CompressFactor:    25,
				},
				{
					IncrementStart:    utils.NewDecimal(int64(25*time.Second), 0),
					RateIntervalIndex: 1,
					RateID:            "UUID01",
					Usage:             utils.NewDecimal(int64(10*time.Second), 0),
					CompressFactor:    2,
				},
			},
			CompressFactor: 1,
		},
		{
			IntervalStart: utils.NewDecimal(int64(35*time.Second), 0),
			Increments: []*utils.RateSIncrement{
				{
					IncrementStart:    utils.NewDecimal(int64(35*time.Second), 0),
					RateIntervalIndex: 0,
					RateID:            "UUID30",
					Usage:             utils.NewDecimal(int64(11*time.Second), 0),
					CompressFactor:    3,
				},
			},
			CompressFactor: 1,
		},
		{
			IntervalStart: utils.NewDecimal(int64(46*time.Second), 0),
			Increments: []*utils.RateSIncrement{
				{
					IncrementStart:    utils.NewDecimal(int64(46*time.Second), 0),
					RateIntervalIndex: 0,
					RateID:            "UUID20",
					Usage:             utils.NewDecimal(int64(9*time.Second), 0),
					CompressFactor:    9,
				},
				{
					IncrementStart:    utils.NewDecimal(int64(55*time.Second), 0),
					RateIntervalIndex: 1,
					RateID:            "UUID21",
					Usage:             utils.NewDecimal(int64(5*time.Second), 0),
					CompressFactor:    1,
				},
			},
			CompressFactor: 1,
		},
		{
			IntervalStart: utils.NewDecimal(int64(time.Minute), 0),
			Increments: []*utils.RateSIncrement{
				{
					IncrementStart:    utils.NewDecimal(int64(time.Minute), 0),
					RateIntervalIndex: 1,
					RateID:            "UUID31",
					Usage:             utils.NewDecimal(int64(10*time.Second), 0),
					CompressFactor:    2,
				},
			},
			CompressFactor: 1,
		},
	}
	expCstRts := map[string]*utils.IntervalRate{
		"UUID00": {
			IntervalStart: utils.NewDecimal(0, 0),
			RecurrentFee:  utils.NewDecimal(1, 1),
			Unit:          tsecDecimal,
			Increment:     secDecimal,
		},
		"UUID01": {
			IntervalStart: utils.NewDecimal(int64(25*time.Second), 0),
			RecurrentFee:  utils.NewDecimal(25, 2),
			Unit:          minDecimal,
			Increment:     ssecDecimal,
		},
		"UUID30": {
			IntervalStart: utils.NewDecimal(int64(30*time.Second), 0),
			RecurrentFee:  utils.NewDecimal(1, 1),
			Unit:          fssecDecimal,
			Increment:     fsecDecimal,
		},
		"UUID20": {
			IntervalStart: utils.NewDecimal(int64(45*time.Second), 0),
			RecurrentFee:  utils.NewDecimal(15, 2),
			Unit:          tsecDecimal,
			Increment:     secDecimal,
		},
		"UUID21": {
			IntervalStart: utils.NewDecimal(int64(55*time.Second), 0),
			RecurrentFee:  utils.NewDecimal(3, 1),
			Unit:          tsecDecimal,
			Increment:     fsecDecimal,
		},
		"UUID31": {
			IntervalStart: utils.NewDecimal(int64(time.Minute), 0),
			RecurrentFee:  utils.NewDecimal(5, 3),
			Unit:          tsecDecimal,
			Increment:     fsecDecimal,
		},
	}
	cstRts := make(map[string]*utils.IntervalRate)
	if rtIvls, err := computeRateSIntervals(ordRts, utils.NewDecimal(0, 0).Big,
		utils.NewDecimal(int64(time.Minute+10*time.Second), 0).Big, cstRts); err != nil {
		t.Error(err)
	} else {
		for idx, val := range rtIvls {
			if !val.Equals(expOrdRts[idx], cstRts, expCstRts) {
				t.Errorf("expecting: %+v \n,received: %+v", utils.ToJSON(expCstRts), utils.ToJSON(cstRts))
				t.Fatalf("expecting: %+v \n,received: %+v", utils.ToJSON(expOrdRts), utils.ToJSON(rtIvls))
			}
		}
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
	rt1 := &utils.Rate{
		ID: "RATE1",
		IntervalRates: []*utils.IntervalRate{
			{
				IntervalStart: utils.NewDecimal(int64(time.Minute), 0),
				RecurrentFee:  utils.NewDecimal(3, 1),
				Unit:          tsecDecimal,
				Increment:     secDecimal,
			},
			{
				IntervalStart: utils.NewDecimal(int64(3*time.Minute), 0),
				RecurrentFee:  utils.NewDecimal(3, 1),
				Unit:          minDecimal,
				Increment:     ssecDecimal,
			},
		},
	}

	rt2 := &utils.Rate{
		ID: "RATE2",
		IntervalRates: []*utils.IntervalRate{
			{
				IntervalStart: utils.NewDecimal(int64(time.Minute+30*time.Second), 0),
				RecurrentFee:  utils.NewDecimal(2, 1),
				Unit:          tsecDecimal,
				Increment:     secDecimal,
			},
			{
				IntervalStart: utils.NewDecimal(int64(2*time.Minute+30*time.Second), 0),
				RecurrentFee:  utils.NewDecimal(2, 1),
				Unit:          tsecDecimal,
				Increment:     ssecDecimal,
			},
		},
	}

	rt3 := &utils.Rate{
		ID: "RATE3",
		IntervalRates: []*utils.IntervalRate{
			{
				IntervalStart: utils.NewDecimal(int64(2*time.Minute), 0),
				RecurrentFee:  utils.NewDecimal(1, 1),
				Unit:          tsecDecimal,
				Increment:     tsecDecimal,
			},
		},
	}

	rp := &utils.RateProfile{
		Tenant: "cgrates.org",
		ID:     "RATE_PROFILE",
		Rates: map[string]*utils.Rate{
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
			utils.NewDecimal(0, 0).Big,
			rt1,
		},
		{
			utils.NewDecimal(int64(time.Minute), 0).Big,
			rt1,
		},
		{
			utils.NewDecimal(int64(time.Minute+30*time.Second), 0).Big,
			rt2,
		},
		{
			utils.NewDecimal(int64(2*time.Minute), 0).Big,
			rt3,
		},
		{
			utils.NewDecimal(int64(2*time.Minute+30*time.Second), 0).Big,
			rt2,
		},
		{
			utils.NewDecimal(int64(3*time.Minute), 0).Big,
			rt1,
		},
	}
	expOrdRts := []*utils.RateSInterval{
		{
			IntervalStart: utils.NewDecimal(int64(time.Minute), 0),
			Increments: []*utils.RateSIncrement{
				{
					IncrementStart:    utils.NewDecimal(int64(time.Minute), 0),
					RateIntervalIndex: 0,
					RateID:            "UUID00",
					Usage:             utils.NewDecimal(int64(30*time.Second), 0),
					CompressFactor:    30,
				},
			},
			CompressFactor: 1,
		},
		{
			IntervalStart: utils.NewDecimal(int64(time.Minute+30*time.Second), 0),
			Increments: []*utils.RateSIncrement{
				{
					IncrementStart:    utils.NewDecimal(int64(time.Minute+30*time.Second), 0),
					RateIntervalIndex: 0,
					RateID:            "UUID10",
					Usage:             utils.NewDecimal(int64(30*time.Second), 0),
					CompressFactor:    30,
				},
			},
			CompressFactor: 1,
		},
		{
			IntervalStart: utils.NewDecimal(int64(2*time.Minute), 0),
			Increments: []*utils.RateSIncrement{
				{
					IncrementStart:    utils.NewDecimal(int64(2*time.Minute), 0),
					RateIntervalIndex: 0,
					RateID:            "UUID20",
					Usage:             utils.NewDecimal(int64(30*time.Second), 0),
					CompressFactor:    1,
				},
			},
			CompressFactor: 1,
		},
		{
			IntervalStart: utils.NewDecimal(int64(2*time.Minute+30*time.Second), 0),
			Increments: []*utils.RateSIncrement{
				{
					IncrementStart:    utils.NewDecimal(int64(2*time.Minute+30*time.Second), 0),
					RateIntervalIndex: 1,
					RateID:            "UUID11",
					Usage:             utils.NewDecimal(int64(30*time.Second), 0),
					CompressFactor:    5,
				},
			},
			CompressFactor: 1,
		},
		{
			IntervalStart: utils.NewDecimal(int64(3*time.Minute), 0),
			Increments: []*utils.RateSIncrement{
				{
					IncrementStart:    utils.NewDecimal(int64(3*time.Minute), 0),
					RateIntervalIndex: 1,
					RateID:            "UUID01",
					Usage:             utils.NewDecimal(int64(4*time.Minute), 0),
					CompressFactor:    35,
				},
			},
			CompressFactor: 1,
		},
	}
	expCstRts := map[string]*utils.IntervalRate{
		"UUID00": {
			IntervalStart: utils.NewDecimal(int64(time.Minute), 0),
			RecurrentFee:  utils.NewDecimal(3, 1),
			Unit:          tsecDecimal,
			Increment:     secDecimal,
		},
		"UUID10": {
			IntervalStart: utils.NewDecimal(int64(time.Minute+30*time.Second), 0),
			RecurrentFee:  utils.NewDecimal(2, 1),
			Unit:          tsecDecimal,
			Increment:     secDecimal,
		},
		"UUID20": {
			IntervalStart: utils.NewDecimal(int64(2*time.Minute), 0),
			RecurrentFee:  utils.NewDecimal(1, 1),
			Unit:          tsecDecimal,
			Increment:     tsecDecimal,
		},
		"UUID11": {
			IntervalStart: utils.NewDecimal(int64(2*time.Minute+30*time.Second), 0),
			RecurrentFee:  utils.NewDecimal(2, 1),
			Unit:          tsecDecimal,
			Increment:     ssecDecimal,
		},
		"UUID01": {
			IntervalStart: utils.NewDecimal(int64(3*time.Minute), 0),
			RecurrentFee:  utils.NewDecimal(3, 1),
			Unit:          minDecimal,
			Increment:     ssecDecimal,
		},
	}
	cstRts := make(map[string]*utils.IntervalRate)
	if rtIvls, err := computeRateSIntervals(ordRates, utils.NewDecimal(int64(time.Minute), 0).Big,
		utils.NewDecimal(int64(6*time.Minute), 0).Big, cstRts); err != nil {
		t.Error(err)
	} else {
		for idx, val := range rtIvls {
			if !val.Equals(expOrdRts[idx], cstRts, expCstRts) {
				t.Errorf("expecting: %+v \n,received: %+v", utils.ToJSON(expCstRts), utils.ToJSON(cstRts))
				t.Fatalf("expecting: %+v \n,received: %+v", utils.ToJSON(expOrdRts), utils.ToJSON(rtIvls))
			}
		}
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
	rt1 := &utils.Rate{
		ID: "RATE1",
		IntervalRates: []*utils.IntervalRate{
			{
				IntervalStart: utils.NewDecimal(0, 0),
				RecurrentFee:  utils.NewDecimal(5, 1),
				Unit:          hourDecimal,
				Increment:     tminDecimal,
			},
			{
				IntervalStart: utils.NewDecimal(int64(4*time.Hour), 0),
				RecurrentFee:  utils.NewDecimal(35, 2),
				Unit:          hourDecimal,
				Increment:     tminDecimal,
			},
		},
	}

	rt2 := &utils.Rate{
		ID: "RATE2",
		IntervalRates: []*utils.IntervalRate{
			{
				IntervalStart: utils.NewDecimal(int64(13*time.Hour), 0),
				RecurrentFee:  utils.NewDecimal(4, 1),
				Unit:          hourDecimal,
				Increment:     tminDecimal,
			},
			{
				IntervalStart: utils.NewDecimal(int64(16*time.Hour), 0),
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
	rtGH := &utils.Rate{
		ID: "RATE_GOLDEN_HOUR",
		IntervalRates: []*utils.IntervalRate{
			{
				IntervalStart: utils.NewDecimal(int64(12*time.Hour), 0),
				RecurrentFee:  utils.NewDecimal(2, 1),
				Unit:          thourDecimal,
				Increment:     dminDecimal,
			},
			{
				IntervalStart: utils.NewDecimal(int64(12*time.Hour+30*time.Minute), 0),
				RecurrentFee:  utils.NewDecimal(1, 1),
				Unit:          hourDecimal,
				Increment:     fminDecimal,
			},
		},
	}

	rp := &utils.RateProfile{
		Tenant: "cgrates.org",
		ID:     "RATE_PROFIOLE",
		Rates: map[string]*utils.Rate{
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
			utils.NewDecimal(0, 0).Big,
			rt1,
		},
		{
			utils.NewDecimal(int64(6*time.Hour), 0).Big,
			rt1,
		},
		{
			utils.NewDecimal(int64(12*time.Hour), 0).Big,
			rtGH,
		},
		{
			utils.NewDecimal(int64(13*time.Hour), 0).Big,
			rt2,
		},
	}

	expOrdRts := []*utils.RateSInterval{
		{
			IntervalStart: utils.NewDecimal(0, 0),
			Increments: []*utils.RateSIncrement{
				{
					IncrementStart:    utils.NewDecimal(0, 0),
					RateIntervalIndex: 0,
					RateID:            "UUID00",
					Usage:             utils.NewDecimal(int64(4*time.Hour), 0),
					CompressFactor:    80,
				},
				{
					IncrementStart:    utils.NewDecimal(int64(4*time.Hour), 0),
					RateIntervalIndex: 1,
					RateID:            "UUID01",
					Usage:             utils.NewDecimal(int64(2*time.Hour), 0),
					CompressFactor:    40,
				},
			},
			CompressFactor: 1,
		},
		{
			IntervalStart: utils.NewDecimal(int64(6*time.Hour), 0),
			Increments: []*utils.RateSIncrement{
				{
					IncrementStart:    utils.NewDecimal(int64(6*time.Hour), 0),
					RateIntervalIndex: 1,
					RateID:            "UUID01",
					Usage:             utils.NewDecimal(int64(6*time.Hour), 0),
					CompressFactor:    120,
				},
			},
			CompressFactor: 1,
		},
		{
			IntervalStart: utils.NewDecimal(int64(12*time.Hour), 0),
			Increments: []*utils.RateSIncrement{
				{
					IncrementStart:    utils.NewDecimal(int64(12*time.Hour), 0),
					RateIntervalIndex: 0,
					RateID:            "UUIDGH0",
					Usage:             utils.NewDecimal(int64(30*time.Minute), 0),
					CompressFactor:    15,
				},
				{
					IncrementStart:    utils.NewDecimal(int64(12*time.Hour+30*time.Minute), 0),
					RateIntervalIndex: 1,
					RateID:            "UUIDGH1",
					Usage:             utils.NewDecimal(int64(30*time.Minute), 0),
					CompressFactor:    6,
				},
			},
			CompressFactor: 1,
		},
		{
			IntervalStart: utils.NewDecimal(int64(13*time.Hour), 0),
			Increments: []*utils.RateSIncrement{
				{
					IncrementStart:    utils.NewDecimal(int64(13*time.Hour), 0),
					RateIntervalIndex: 0,
					RateID:            "UUID10",
					Usage:             utils.NewDecimal(int64(3*time.Hour), 0),
					CompressFactor:    60,
				},
				{
					IncrementStart:    utils.NewDecimal(int64(16*time.Hour), 0),
					RateIntervalIndex: 1,
					RateID:            "UUID11",
					Usage:             utils.NewDecimal(int64(9*time.Hour), 0),
					CompressFactor:    180,
				},
			},
			CompressFactor: 1,
		},
	}
	expCstRts := map[string]*utils.IntervalRate{
		"UUID00": {
			IntervalStart: utils.NewDecimal(0, 0),
			RecurrentFee:  utils.NewDecimal(5, 1),
			Unit:          hourDecimal,
			Increment:     tminDecimal,
		},
		"UUID01": {
			IntervalStart: utils.NewDecimal(int64(4*time.Hour), 0),
			RecurrentFee:  utils.NewDecimal(35, 2),
			Unit:          hourDecimal,
			Increment:     tminDecimal,
		},
		"UUIDGH0": {
			IntervalStart: utils.NewDecimal(int64(12*time.Hour), 0),
			RecurrentFee:  utils.NewDecimal(2, 1),
			Unit:          thourDecimal,
			Increment:     dminDecimal,
		},
		"UUIDGH1": {
			IntervalStart: utils.NewDecimal(int64(12*time.Hour+30*time.Minute), 0),
			RecurrentFee:  utils.NewDecimal(1, 1),
			Unit:          hourDecimal,
			Increment:     fminDecimal,
		},
		"UUID10": {
			IntervalStart: utils.NewDecimal(int64(13*time.Hour), 0),
			RecurrentFee:  utils.NewDecimal(4, 1),
			Unit:          hourDecimal,
			Increment:     tminDecimal,
		},
		"UUID11": {
			IntervalStart: utils.NewDecimal(int64(16*time.Hour), 0),
			RecurrentFee:  utils.NewDecimal(35, 2),
			Unit:          hourDecimal,
			Increment:     tminDecimal,
		},
	}
	cstRts := make(map[string]*utils.IntervalRate)
	if rtIvls, err := computeRateSIntervals(ordRts, utils.NewDecimal(0, 0).Big,
		utils.NewDecimal(int64(25*time.Hour), 0).Big, cstRts); err != nil {
		t.Error(err)
	} else {
		for idx, val := range rtIvls {
			if !val.Equals(expOrdRts[idx], cstRts, expCstRts) {
				t.Fatalf("expecting: %+v \n,received: %+v", utils.ToJSON(expOrdRts), utils.ToJSON(rtIvls))
			}
		}
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
	rt1 := &utils.Rate{
		ID: "RATE1",
		IntervalRates: []*utils.IntervalRate{
			{
				IntervalStart: utils.NewDecimal(0, 0),
				RecurrentFee:  utils.NewDecimal(5, 1),
				Unit:          tsecDecimal,
				Increment:     twsecDecimal,
			},
			{
				IntervalStart: utils.NewDecimal(int64(4*time.Second), 0),
				RecurrentFee:  utils.NewDecimal(48, 2),
				Unit:          tsecDecimal,
				Increment:     twsecDecimal,
			},
			{
				IntervalStart: utils.NewDecimal(int64(8*time.Second), 0),
				RecurrentFee:  utils.NewDecimal(45, 2),
				Unit:          tsecDecimal,
				Increment:     twsecDecimal,
			},
		},
	}

	rt2 := &utils.Rate{
		ID: "RATE2",
		IntervalRates: []*utils.IntervalRate{
			{
				IntervalStart: utils.NewDecimal(0, 0),
				RecurrentFee:  utils.NewDecimal(5, 1),
				Unit:          twsecDecimal,
				Increment:     secDecimal,
			},
			{
				IntervalStart: utils.NewDecimal(int64(2*time.Second), 0),
				RecurrentFee:  utils.NewDecimal(48, 2),
				Unit:          twsecDecimal,
				Increment:     secDecimal,
			},
			{
				IntervalStart: utils.NewDecimal(int64(6*time.Second), 0),
				RecurrentFee:  utils.NewDecimal(45, 2),
				Unit:          twsecDecimal,
				Increment:     secDecimal,
			},
		},
	}

	rp := &utils.RateProfile{
		Tenant: "cgrates.org",
		ID:     "RATE_PROFILE",
		Rates: map[string]*utils.Rate{
			rt1.ID: rt1,
			rt2.ID: rt2,
		},
	}
	if err := rp.Compile(); err != nil {
		t.Fatal(err)
	}

	ordRts := []*orderedRate{
		{
			utils.NewDecimal(0, 0).Big,
			rt1,
		},
		{
			utils.NewDecimal(int64(time.Second), 0).Big,
			rt2,
		},
		{
			utils.NewDecimal(int64(2*time.Second), 0).Big,
			rt1,
		},
		{
			utils.NewDecimal(int64(3*time.Second), 0).Big,
			rt2,
		},
		{
			utils.NewDecimal(int64(5*time.Second), 0).Big,
			rt1,
		},
		{
			utils.NewDecimal(int64(7*time.Second), 0).Big,
			rt2,
		},
	}

	expOrdRts := []*utils.RateSInterval{
		{
			IntervalStart: utils.NewDecimal(0, 0),
			Increments: []*utils.RateSIncrement{
				{
					IncrementStart:    utils.NewDecimal(0, 0),
					RateIntervalIndex: 0,
					RateID:            "UUID00",
					CompressFactor:    1,
					Usage:             utils.NewDecimal(int64(time.Second), 0),
				},
			},
			CompressFactor: 1,
		},
		{
			IntervalStart: utils.NewDecimal(int64(time.Second), 0),
			Increments: []*utils.RateSIncrement{
				{
					IncrementStart:    utils.NewDecimal(int64(time.Second), 0),
					RateIntervalIndex: 0,
					RateID:            "UUID10",
					CompressFactor:    1,
					Usage:             utils.NewDecimal(int64(time.Second), 0),
				},
			},
			CompressFactor: 1,
		},
		{
			IntervalStart: utils.NewDecimal(int64(2*time.Second), 0),
			Increments: []*utils.RateSIncrement{
				{
					IncrementStart:    utils.NewDecimal(int64(2*time.Second), 0),
					RateIntervalIndex: 0,
					RateID:            "UUID01",
					CompressFactor:    1,
					Usage:             utils.NewDecimal(int64(time.Second), 0),
				},
			},
			CompressFactor: 1,
		},
		{
			IntervalStart: utils.NewDecimal(int64(3*time.Second), 0),
			Increments: []*utils.RateSIncrement{
				{
					IncrementStart:    utils.NewDecimal(int64(3*time.Second), 0),
					RateIntervalIndex: 1,
					RateID:            "UUID11",
					CompressFactor:    2,
					Usage:             utils.NewDecimal(int64(2*time.Second), 0),
				},
			},
			CompressFactor: 1,
		},
		{
			IntervalStart: utils.NewDecimal(int64(5*time.Second), 0),
			Increments: []*utils.RateSIncrement{
				{
					IncrementStart:    utils.NewDecimal(int64(5*time.Second), 0),
					RateIntervalIndex: 1,
					RateID:            "UUID02",
					CompressFactor:    1,
					Usage:             utils.NewDecimal(int64(2*time.Second), 0),
				},
			},
			CompressFactor: 1,
		},
		{
			IntervalStart: utils.NewDecimal(int64(7*time.Second), 0),
			Increments: []*utils.RateSIncrement{
				{
					IncrementStart:    utils.NewDecimal(int64(7*time.Second), 0),
					RateIntervalIndex: 2,
					RateID:            "UUID12",
					CompressFactor:    3,
					Usage:             utils.NewDecimal(int64(3*time.Second), 0),
				},
			},
			CompressFactor: 1,
		},
	}
	expCstRts := map[string]*utils.IntervalRate{
		"UUID00": {
			IntervalStart: utils.NewDecimal(0, 0),
			RecurrentFee:  utils.NewDecimal(5, 1),
			Unit:          tsecDecimal,
			Increment:     twsecDecimal,
		},
		"UUID10": {
			IntervalStart: utils.NewDecimal(0, 0),
			RecurrentFee:  utils.NewDecimal(5, 1),
			Unit:          twsecDecimal,
			Increment:     secDecimal,
		},
		"UUID01": {
			IntervalStart: utils.NewDecimal(0, 0),
			RecurrentFee:  utils.NewDecimal(5, 1),
			Unit:          tsecDecimal,
			Increment:     twsecDecimal,
		},
		"UUID11": {
			IntervalStart: utils.NewDecimal(int64(2*time.Second), 0),
			RecurrentFee:  utils.NewDecimal(48, 2),
			Unit:          twsecDecimal,
			Increment:     secDecimal,
		},
		"UUID02": {
			IntervalStart: utils.NewDecimal(int64(4*time.Second), 0),
			RecurrentFee:  utils.NewDecimal(48, 2),
			Unit:          tsecDecimal,
			Increment:     twsecDecimal,
		},
		"UUID12": {
			IntervalStart: utils.NewDecimal(int64(6*time.Second), 0),
			RecurrentFee:  utils.NewDecimal(45, 2),
			Unit:          twsecDecimal,
			Increment:     secDecimal,
		},
	}
	cstRts := make(map[string]*utils.IntervalRate)
	if rtIvls, err := computeRateSIntervals(ordRts, utils.NewDecimal(0, 0).Big,
		utils.NewDecimal(int64(10*time.Second), 0).Big, cstRts); err != nil {
		t.Error(err)
	} else {
		for idx, val := range rtIvls {
			if !val.Equals(expOrdRts[idx], cstRts, expCstRts) {
				t.Errorf("expecting: %+v \n,received: %+v", utils.ToJSON(expCstRts), utils.ToJSON(cstRts))
				t.Fatalf("expecting: %+v \n,received: %+v", utils.ToJSON(expOrdRts), utils.ToJSON(rtIvls))
			}
		}
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
	rt1 := &utils.Rate{
		ID: "RATE1",
		IntervalRates: []*utils.IntervalRate{
			{
				IntervalStart: utils.NewDecimal(0, 0),
				RecurrentFee:  utils.NewDecimal(5, 3),
				Unit:          minDecimal,
				Increment:     minDecimal,
			},
			{
				IntervalStart: utils.NewDecimal(int64(time.Hour), 0),
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
	rt2 := &utils.Rate{
		ID: "RATE2",
		IntervalRates: []*utils.IntervalRate{
			{
				IntervalStart: utils.NewDecimal(int64(15*time.Minute), 0),
				RecurrentFee:  utils.NewDecimal(1, 1),
				Unit:          tminDeminal,
				Increment:     ominDeminal,
			},
			{
				IntervalStart: utils.NewDecimal(int64(30*time.Minute), 0),
				RecurrentFee:  utils.NewDecimal(15, 2),
				Unit:          tminDeminal,
				Increment:     ominDeminal,
			},
			{
				IntervalStart: utils.NewDecimal(int64(45*time.Minute), 0),
				RecurrentFee:  utils.NewDecimal(2, 1),
				Unit:          tminDeminal,
				Increment:     ominDeminal,
			},
		},
	}

	rp := &utils.RateProfile{
		Tenant: "cgrates.org",
		ID:     "RATE_PROFILE",
		Rates: map[string]*utils.Rate{
			rt1.ID: rt1,
			rt2.ID: rt2,
		},
	}
	if err := rp.Compile(); err != nil {
		t.Fatal(err)
	}

	ordRts := []*orderedRate{
		{
			utils.NewDecimal(0, 0).Big,
			rt1,
		},
		{
			utils.NewDecimal(int64(20*time.Minute), 0).Big,
			rt2,
		},
		{
			utils.NewDecimal(int64(time.Hour+time.Minute), 0).Big,
			rt1,
		},
	}

	expOrdRts := []*utils.RateSInterval{
		{
			IntervalStart: utils.NewDecimal(0, 0),
			Increments: []*utils.RateSIncrement{
				{
					IncrementStart:    utils.NewDecimal(0, 0),
					RateIntervalIndex: 0,
					RateID:            "UUID00",
					CompressFactor:    2,
					Usage:             utils.NewDecimal(int64(20*time.Minute), 0),
				},
			},
			CompressFactor: 1,
		},
		{
			IntervalStart: utils.NewDecimal(int64(20*time.Minute), 0),
			Increments: []*utils.RateSIncrement{
				{
					IncrementStart:    utils.NewDecimal(int64(20*time.Minute), 0),
					RateIntervalIndex: 0,
					RateID:            "UUID10",
					CompressFactor:    10,
					Usage:             utils.NewDecimal(int64(10*time.Minute), 0),
				},
				{
					IncrementStart:    utils.NewDecimal(int64(30*time.Minute), 0),
					RateIntervalIndex: 1,
					RateID:            "UUID11",
					CompressFactor:    15,
					Usage:             utils.NewDecimal(int64(15*time.Minute), 0),
				},
				{
					IncrementStart:    utils.NewDecimal(int64(45*time.Minute), 0),
					RateIntervalIndex: 2,
					RateID:            "UUID12",
					CompressFactor:    16,
					Usage:             utils.NewDecimal(int64(16*time.Minute), 0),
				},
			},
			CompressFactor: 1,
		},
		{
			IntervalStart: utils.NewDecimal(int64(time.Hour+time.Minute), 0),
			Increments: []*utils.RateSIncrement{
				{
					IncrementStart:    utils.NewDecimal(int64(time.Hour+time.Minute), 0),
					RateIntervalIndex: 1,
					RateID:            "UUID01",
					CompressFactor:    6,
					Usage:             utils.NewDecimal(int64(9*time.Minute), 0),
				},
			},
			CompressFactor: 1,
		},
	}
	expCstRts := map[string]*utils.IntervalRate{
		"UUID00": {
			IntervalStart: utils.NewDecimal(0, 0),
			RecurrentFee:  utils.NewDecimal(5, 3),
			Unit:          minDecimal,
			Increment:     minDecimal,
		},
		"UUID10": {
			IntervalStart: utils.NewDecimal(int64(15*time.Minute), 0),
			RecurrentFee:  utils.NewDecimal(1, 1),
			Unit:          tminDeminal,
			Increment:     ominDeminal,
		},
		"UUID11": {
			IntervalStart: utils.NewDecimal(int64(30*time.Minute), 0),
			RecurrentFee:  utils.NewDecimal(15, 2),
			Unit:          tminDeminal,
			Increment:     ominDeminal,
		},
		"UUID12": {
			IntervalStart: utils.NewDecimal(int64(45*time.Minute), 0),
			RecurrentFee:  utils.NewDecimal(2, 1),
			Unit:          tminDeminal,
			Increment:     ominDeminal,
		},
		"UUID01": {
			IntervalStart: utils.NewDecimal(int64(time.Hour), 0),
			RecurrentFee:  utils.NewDecimal(5, 1),
			Unit:          fminDeminal,
			Increment:     cDecimal,
		},
	}
	cstRts := make(map[string]*utils.IntervalRate)
	if rtIvls, err := computeRateSIntervals(ordRts, utils.NewDecimal(0, 0).Big,
		utils.NewDecimal(int64(time.Hour+10*time.Minute), 0).Big, cstRts); err != nil {
		t.Error(err)
	} else {
		for idx, val := range rtIvls {
			if !val.Equals(expOrdRts[idx], cstRts, expCstRts) {
				t.Errorf("expecting: %+v \n,received: %+v", utils.ToJSON(expCstRts), utils.ToJSON(cstRts))
				t.Fatalf("expecting: %+v \n,received: %+v", utils.ToJSON(expOrdRts), utils.ToJSON(rtIvls))
			}
		}
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

	rt1 := &utils.Rate{
		ID: "RATE1",
		IntervalRates: []*utils.IntervalRate{
			{
				IntervalStart: utils.NewDecimal(0, 0),
				RecurrentFee:  utils.NewDecimal(2, 1),
				Unit:          minDecimal,
				Increment:     tsecDecimal,
			},
			{
				IntervalStart: utils.NewDecimal(int64(30*time.Minute), 0),
				RecurrentFee:  utils.NewDecimal(2, 1),
				Unit:          minDecimal,
				Increment:     tsecDecimal,
			},
		},
	}

	rt2 := &utils.Rate{
		ID: "RATE2",
		IntervalRates: []*utils.IntervalRate{
			{
				IntervalStart: utils.NewDecimal(0, 0),
				RecurrentFee:  utils.NewDecimal(2, 1),
				Unit:          minDecimal,
				Increment:     tsecDecimal,
			},
			{
				IntervalStart: utils.NewDecimal(int64(30*time.Minute), 0),
				RecurrentFee:  utils.NewDecimal(2, 1),
				Unit:          minDecimal,
				Increment:     tsecDecimal,
			},
		},
	}

	rp := &utils.RateProfile{
		Tenant: "cgrates.org",
		ID:     "RATE_PROFILE",
		Rates: map[string]*utils.Rate{
			rt1.ID: rt1,
			rt2.ID: rt2,
		},
	}
	if err := rp.Compile(); err != nil {
		t.Fatal(err)
	}

	ordRts := []*orderedRate{
		{
			utils.NewDecimal(0, 0).Big,
			rt1,
		},
		{
			utils.NewDecimal(int64(25*time.Minute), 0).Big,
			rt1,
		},
	}

	expOrdRts := []*utils.RateSInterval{
		{
			IntervalStart: utils.NewDecimal(0, 0),
			Increments: []*utils.RateSIncrement{
				{
					IncrementStart:    utils.NewDecimal(0, 0),
					RateIntervalIndex: 0,
					RateID:            "UUID00",
					CompressFactor:    50,
					Usage:             utils.NewDecimal(int64(25*time.Minute), 0),
				},
			},
			CompressFactor: 1,
		},
		{
			IntervalStart: utils.NewDecimal(int64(25*time.Minute), 0),
			Increments: []*utils.RateSIncrement{
				{
					IncrementStart:    utils.NewDecimal(int64(25*time.Minute), 0),
					RateIntervalIndex: 0,
					RateID:            "UUID00",
					CompressFactor:    10,
					Usage:             utils.NewDecimal(int64(5*time.Minute), 0),
				},
				{
					IncrementStart:    utils.NewDecimal(int64(30*time.Minute), 0),
					RateIntervalIndex: 1,
					RateID:            "UUID01",
					CompressFactor:    60,
					Usage:             utils.NewDecimal(int64(30*time.Minute), 0),
				},
			},
			CompressFactor: 1,
		},
	}
	expCstRts := map[string]*utils.IntervalRate{
		"UUID00": {
			IntervalStart: utils.NewDecimal(0, 0),
			RecurrentFee:  utils.NewDecimal(2, 1),
			Unit:          minDecimal,
			Increment:     tsecDecimal,
		},
		"UUID01": {
			IntervalStart: utils.NewDecimal(int64(30*time.Minute), 0),
			RecurrentFee:  utils.NewDecimal(2, 1),
			Unit:          minDecimal,
			Increment:     tsecDecimal,
		},
	}
	cstRts := make(map[string]*utils.IntervalRate)
	if rtIvls, err := computeRateSIntervals(ordRts, utils.NewDecimal(0, 0).Big,
		utils.NewDecimal(int64(time.Hour), 0).Big, cstRts); err != nil {
		t.Error(err)
	} else {
		for idx, val := range rtIvls {
			if !val.Equals(expOrdRts[idx], cstRts, expCstRts) {
				t.Errorf("expecting: %+v \n,received: %+v", utils.ToJSON(expCstRts), utils.ToJSON(cstRts))
				t.Fatalf("expecting: %+v \n,received: %+v", utils.ToJSON(expOrdRts), utils.ToJSON(rtIvls))
			}
		}
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

	rt1 := &utils.Rate{
		ID: "RATE1",
		IntervalRates: []*utils.IntervalRate{
			{
				IntervalStart: utils.NewDecimal(0, 0),
				RecurrentFee:  utils.NewDecimal(22, 2),
				Unit:          minDecimal,
				Increment:     secDecimal,
			},
		},
	}

	rt2 := &utils.Rate{
		ID: "RATE2",
		IntervalRates: []*utils.IntervalRate{
			{
				IntervalStart: utils.NewDecimal(int64(60*time.Minute), 0),
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

	rt3 := &utils.Rate{
		ID: "RATE3",
		IntervalRates: []*utils.IntervalRate{
			{
				IntervalStart: utils.NewDecimal(int64(120*time.Minute), 0),
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

	rt4 := &utils.Rate{
		ID: "RATE4",
		IntervalRates: []*utils.IntervalRate{
			{
				IntervalStart: utils.NewDecimal(int64(180*time.Minute), 0),
				RecurrentFee:  utils.NewDecimal(16, 2),
				Unit:          trminDecimal,
				Increment:     tsecDecimal,
			},
		},
	}

	rp := &utils.RateProfile{
		Tenant: "cgrates.org",
		ID:     "RATE_PROFILE",
		Rates: map[string]*utils.Rate{
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
			utils.NewDecimal(int64(20*time.Minute), 0).Big,
			rt1,
		},
		{
			utils.NewDecimal(int64(80*time.Minute), 0).Big,
			rt2,
		},
		{
			utils.NewDecimal(int64(140*time.Minute), 0).Big,
			rt3,
		},
		{
			utils.NewDecimal(int64(200*time.Minute), 0).Big,
			rt4,
		},
	}

	expOrdRts := []*utils.RateSInterval{
		{
			IntervalStart: utils.NewDecimal(int64(20*time.Minute), 0),
			Increments: []*utils.RateSIncrement{
				{
					IncrementStart:    utils.NewDecimal(int64(20*time.Minute), 0),
					RateIntervalIndex: 0,
					RateID:            "UUID00",
					CompressFactor:    3600,
					Usage:             utils.NewDecimal(int64(time.Hour), 0),
				},
			},
			CompressFactor: 1,
		},
		{
			IntervalStart: utils.NewDecimal(int64(80*time.Minute), 0),
			Increments: []*utils.RateSIncrement{
				{
					IncrementStart:    utils.NewDecimal(int64(80*time.Minute), 0),
					RateIntervalIndex: 0,
					RateID:            "UUID10",
					CompressFactor:    360,
					Usage:             utils.NewDecimal(int64(60*time.Minute), 0),
				},
			},
			CompressFactor: 1,
		},
		{
			IntervalStart: utils.NewDecimal(int64(140*time.Minute), 0),
			Increments: []*utils.RateSIncrement{
				{
					IncrementStart:    utils.NewDecimal(int64(140*time.Minute), 0),
					RateIntervalIndex: 0,
					RateID:            "UUID20",
					CompressFactor:    180,
					Usage:             utils.NewDecimal(int64(60*time.Minute), 0),
				},
			},
			CompressFactor: 1,
		},
		{
			IntervalStart: utils.NewDecimal(int64(200*time.Minute), 0),
			Increments: []*utils.RateSIncrement{
				{
					IncrementStart:    utils.NewDecimal(int64(200*time.Minute), 0),
					RateIntervalIndex: 0,
					RateID:            "UUID30",
					CompressFactor:    60,
					Usage:             utils.NewDecimal(int64(30*time.Minute), 0),
				},
			},
			CompressFactor: 1,
		},
	}

	expCstRts := map[string]*utils.IntervalRate{
		"UUID00": {
			IntervalStart: utils.NewDecimal(0, 0),
			RecurrentFee:  utils.NewDecimal(22, 2),
			Unit:          minDecimal,
			Increment:     secDecimal,
		},
		"UUID10": {
			IntervalStart: utils.NewDecimal(int64(60*time.Minute), 0),
			RecurrentFee:  utils.NewDecimal(2, 1),
			Unit:          tminDecimal,
			Increment:     tensecDecimal,
		},
		"UUID20": {
			IntervalStart: utils.NewDecimal(int64(120*time.Minute), 0),
			RecurrentFee:  utils.NewDecimal(18, 2),
			Unit:          twminDecimal,
			Increment:     twsecDecimal,
		},
		"UUID30": {
			IntervalStart: utils.NewDecimal(int64(180*time.Minute), 0),
			RecurrentFee:  utils.NewDecimal(16, 2),
			Unit:          trminDecimal,
			Increment:     tsecDecimal,
		},
	}
	cstRts := make(map[string]*utils.IntervalRate)
	if rtIvls, err := computeRateSIntervals(ordRts, utils.NewDecimal(int64(20*time.Minute), 0).Big,
		utils.NewDecimal(int64(210*time.Minute), 0).Big, cstRts); err != nil {
		t.Error(err)
	} else {
		for idx, val := range rtIvls {
			if !val.Equals(expOrdRts[idx], cstRts, expCstRts) {
				t.Errorf("expecting: %+v \n,received: %+v", utils.ToJSON(expCstRts), utils.ToJSON(cstRts))
				t.Fatalf("expecting: %+v \n,received: %+v", utils.ToJSON(expOrdRts), utils.ToJSON(rtIvls))
			}
		}
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
	rt1 := &utils.Rate{
		ID: "RATE1",
		IntervalRates: []*utils.IntervalRate{
			{
				IntervalStart: utils.NewDecimal(0, 0),
				RecurrentFee:  utils.NewDecimal(22, 2),
				Unit:          minDecimal,
				Increment:     secDecimal,
			},
			{
				IntervalStart: utils.NewDecimal(int64(60*time.Minute), 0),
				RecurrentFee:  utils.NewDecimal(2, 1),
				Unit:          tminDecimal,
				Increment:     tssecDecimal,
			},
			{
				IntervalStart: utils.NewDecimal(int64(120*time.Minute), 0),
				RecurrentFee:  utils.NewDecimal(18, 2),
				Unit:          twminDecimal,
				Increment:     twsecDecimal,
			},
			{
				IntervalStart: utils.NewDecimal(int64(180*time.Minute), 0),
				RecurrentFee:  utils.NewDecimal(16, 2),
				Unit:          thminDecimal,
				Increment:     tsecDecimal,
			},
		},
	}
	rt1.Compile()

	ordRts := []*orderedRate{
		{
			utils.NewDecimal(int64(20*time.Minute), 0).Big,
			rt1,
		},
		{
			utils.NewDecimal(int64(80*time.Minute), 0).Big,
			rt1,
		},
		{
			utils.NewDecimal(int64(140*time.Minute), 0).Big,
			rt1,
		},
		{
			utils.NewDecimal(int64(200*time.Minute), 0).Big,
			rt1,
		},
	}
	expOrdRts := []*utils.RateSInterval{
		{
			IntervalStart: utils.NewDecimal(int64(20*time.Minute), 0),
			Increments: []*utils.RateSIncrement{
				{
					IncrementStart:    utils.NewDecimal(int64(20*time.Minute), 0),
					RateIntervalIndex: 0,
					RateID:            "UUID00",
					CompressFactor:    2400,
					Usage:             utils.NewDecimal(int64(40*time.Minute), 0),
				},
				{
					IncrementStart:    utils.NewDecimal(int64(60*time.Minute), 0),
					RateIntervalIndex: 1,
					RateID:            "UUID01",
					CompressFactor:    120,
					Usage:             utils.NewDecimal(int64(20*time.Minute), 0),
				},
			},
			CompressFactor: 1,
		},
		{
			IntervalStart: utils.NewDecimal(int64(80*time.Minute), 0),
			Increments: []*utils.RateSIncrement{
				{
					IncrementStart:    utils.NewDecimal(int64(80*time.Minute), 0),
					RateIntervalIndex: 1,
					RateID:            "UUID01",
					CompressFactor:    240,
					Usage:             utils.NewDecimal(int64(40*time.Minute), 0),
				},
				{
					IncrementStart:    utils.NewDecimal(int64(120*time.Minute), 0),
					RateIntervalIndex: 2,
					RateID:            "UUID02",
					CompressFactor:    60,
					Usage:             utils.NewDecimal(int64(20*time.Minute), 0),
				},
			},
			CompressFactor: 1,
		},
		{
			IntervalStart: utils.NewDecimal(int64(140*time.Minute), 0),
			Increments: []*utils.RateSIncrement{
				{
					IncrementStart:    utils.NewDecimal(int64(140*time.Minute), 0),
					RateIntervalIndex: 2,
					RateID:            "UUID02",
					CompressFactor:    120,
					Usage:             utils.NewDecimal(int64(40*time.Minute), 0),
				},
				{
					IncrementStart:    utils.NewDecimal(int64(180*time.Minute), 0),
					RateIntervalIndex: 3,
					RateID:            "UUID03",
					CompressFactor:    40,
					Usage:             utils.NewDecimal(int64(20*time.Minute), 0),
				},
			},
			CompressFactor: 1,
		},
		{
			IntervalStart: utils.NewDecimal(int64(200*time.Minute), 0),
			Increments: []*utils.RateSIncrement{
				{
					IncrementStart:    utils.NewDecimal(int64(200*time.Minute), 0),
					RateIntervalIndex: 3,
					RateID:            "UUID03",
					CompressFactor:    60,
					Usage:             utils.NewDecimal(int64(30*time.Minute), 0),
				},
			},
			CompressFactor: 1,
		},
	}
	expCstRts := map[string]*utils.IntervalRate{
		"UUID00": {
			IntervalStart: utils.NewDecimal(0, 0),
			RecurrentFee:  utils.NewDecimal(22, 2),
			Unit:          minDecimal,
			Increment:     secDecimal,
		},
		"UUID01": {
			IntervalStart: utils.NewDecimal(int64(60*time.Minute), 0),
			RecurrentFee:  utils.NewDecimal(2, 1),
			Unit:          tminDecimal,
			Increment:     tssecDecimal,
		},
		"UUID02": {
			IntervalStart: utils.NewDecimal(int64(120*time.Minute), 0),
			RecurrentFee:  utils.NewDecimal(18, 2),
			Unit:          twminDecimal,
			Increment:     twsecDecimal,
		},
		"UUID03": {
			IntervalStart: utils.NewDecimal(int64(180*time.Minute), 0),
			RecurrentFee:  utils.NewDecimal(16, 2),
			Unit:          thminDecimal,
			Increment:     tsecDecimal,
		},
	}
	cstRts := make(map[string]*utils.IntervalRate)
	if rtIvls, err := computeRateSIntervals(ordRts, utils.NewDecimal(int64(20*time.Minute), 0).Big,
		utils.NewDecimal(int64(210*time.Minute), 0).Big, cstRts); err != nil {
		t.Error(err)
	} else {
		for idx, val := range rtIvls {
			if !val.Equals(expOrdRts[idx], cstRts, expCstRts) {
				t.Errorf("expecting: %+v \n,received: %+v", utils.ToJSON(expCstRts), utils.ToJSON(cstRts))
				t.Fatalf("expecting: %+v \n,received: %+v", utils.ToJSON(expOrdRts), utils.ToJSON(rtIvls))
			}
		}
	}

	expOrdRts = []*utils.RateSInterval{
		{
			IntervalStart: utils.NewDecimal(0, 0),
			Increments: []*utils.RateSIncrement{
				{
					IncrementStart:    utils.NewDecimal(0, 0),
					RateIntervalIndex: 0,
					RateID:            "UUID00",
					CompressFactor:    3600,
					Usage:             utils.NewDecimal(int64(60*time.Minute), 0),
				},
				{
					IncrementStart:    utils.NewDecimal(int64(60*time.Minute), 0),
					RateIntervalIndex: 1,
					RateID:            "UUID01",
					CompressFactor:    120,
					Usage:             utils.NewDecimal(int64(20*time.Minute), 0),
				},
			},
			CompressFactor: 1,
		},
		{
			IntervalStart: utils.NewDecimal(int64(80*time.Minute), 0),
			Increments: []*utils.RateSIncrement{
				{
					IncrementStart:    utils.NewDecimal(int64(80*time.Minute), 0),
					RateIntervalIndex: 1,
					RateID:            "UUID01",
					CompressFactor:    240,
					Usage:             utils.NewDecimal(int64(40*time.Minute), 0),
				},
				{
					IncrementStart:    utils.NewDecimal(int64(120*time.Minute), 0),
					RateIntervalIndex: 2,
					RateID:            "UUID02",
					CompressFactor:    60,
					Usage:             utils.NewDecimal(int64(20*time.Minute), 0),
				},
			},
			CompressFactor: 1,
		},
		{
			IntervalStart: utils.NewDecimal(int64(140*time.Minute), 0),
			Increments: []*utils.RateSIncrement{
				{
					IncrementStart:    utils.NewDecimal(int64(140*time.Minute), 0),
					RateIntervalIndex: 2,
					RateID:            "UUID02",
					CompressFactor:    120,
					Usage:             utils.NewDecimal(int64(40*time.Minute), 0),
				},
				{
					IncrementStart:    utils.NewDecimal(int64(180*time.Minute), 0),
					RateIntervalIndex: 3,
					RateID:            "UUID03",
					CompressFactor:    40,
					Usage:             utils.NewDecimal(int64(20*time.Minute), 0),
				},
			},
			CompressFactor: 1,
		},
		{
			IntervalStart: utils.NewDecimal(int64(200*time.Minute), 0),
			Increments: []*utils.RateSIncrement{
				{
					IncrementStart:    utils.NewDecimal(int64(200*time.Minute), 0),
					RateIntervalIndex: 3,
					RateID:            "UUID03",
					CompressFactor:    60,
					Usage:             utils.NewDecimal(int64(30*time.Minute), 0),
				},
			},
			CompressFactor: 1,
		},
	}
	cstRts = make(map[string]*utils.IntervalRate)
	if rtIvls, err := computeRateSIntervals(ordRts, utils.NewDecimal(int64(0*time.Minute), 0).Big,
		utils.NewDecimal(int64(230*time.Minute), 0).Big, cstRts); err != nil {
		t.Error(err)
	} else {
		for idx, val := range rtIvls {
			if !val.Equals(expOrdRts[idx], cstRts, expCstRts) {
				t.Errorf("expecting: %+v \n,received: %+v", utils.ToJSON(expCstRts), utils.ToJSON(cstRts))
				t.Fatalf("expecting: %+v \n,received: %+v", utils.ToJSON(expOrdRts), utils.ToJSON(rtIvls))
			}
		}
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
	rt1 := &utils.Rate{
		ID: "RATE1",
		IntervalRates: []*utils.IntervalRate{
			{
				IntervalStart: utils.NewDecimal(0, 0),
				RecurrentFee:  utils.NewDecimal(5, 1),
				Unit:          twminDecimal,
				Increment:     minDecimal,
			},
			{
				IntervalStart: utils.NewDecimal(int64(24*time.Hour), 0),
				RecurrentFee:  utils.NewDecimal(1, 1),
				Unit:          twminDecimal,
				Increment:     secDecimal,
			},
		},
	}

	rt2 := &utils.Rate{
		ID: "RATE2",
		IntervalRates: []*utils.IntervalRate{
			{
				IntervalStart: utils.NewDecimal(int64(4*time.Hour+30*time.Minute), 0),
				RecurrentFee:  utils.NewDecimal(45, 2),
				Unit:          twminDecimal,
				Increment:     nminDecimal,
			},
			{
				IntervalStart: utils.NewDecimal(int64(8*time.Hour+30*time.Minute), 0),
				RecurrentFee:  utils.NewDecimal(4, 1),
				Unit:          twminDecimal,
				Increment:     nminDecimal,
			},
			{
				IntervalStart: utils.NewDecimal(int64(16*time.Hour+30*time.Minute), 0),
				RecurrentFee:  utils.NewDecimal(35, 2),
				Unit:          twminDecimal,
				Increment:     nminDecimal,
			},
			{
				IntervalStart: utils.NewDecimal(int64(20*time.Hour+30*time.Minute), 0),
				RecurrentFee:  utils.NewDecimal(3, 1),
				Unit:          twminDecimal,
				Increment:     nminDecimal,
			},
		},
	}

	rt3 := &utils.Rate{
		ID: "RATE_SPECIAL",
		IntervalRates: []*utils.IntervalRate{
			{
				IntervalStart: utils.NewDecimal(int64(12*time.Hour), 0),
				RecurrentFee:  utils.NewDecimal(2, 1),
				Unit:          twminDecimal,
				Increment:     secDecimal,
			},
		},
	}

	rp := &utils.RateProfile{
		Tenant: "cgrates.org",
		ID:     "RATE_PROFILE",
		Rates: map[string]*utils.Rate{
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
			utils.NewDecimal(0, 0).Big,
			rt1,
		},
		{
			utils.NewDecimal(int64(4*time.Hour+31*time.Minute), 0).Big,
			rt2,
		},
		{
			utils.NewDecimal(int64(12*time.Hour), 0).Big,
			rt3,
		},
		{
			utils.NewDecimal(int64(13*time.Hour), 0).Big,
			rt2,
		},
		{
			utils.NewDecimal(int64(24*time.Hour), 0).Big,
			rt1,
		},
	}

	expOrdRts := []*utils.RateSInterval{
		{
			IntervalStart: utils.NewDecimal(0, 0),
			Increments: []*utils.RateSIncrement{
				{
					IncrementStart:    utils.NewDecimal(0, 0),
					RateIntervalIndex: 0,
					RateID:            "UUID00",
					CompressFactor:    271,
					Usage:             utils.NewDecimal(int64(4*time.Hour+31*time.Minute), 0),
				},
			},
			CompressFactor: 1,
		},
		{
			IntervalStart: utils.NewDecimal(int64(4*time.Hour+31*time.Minute), 0),
			Increments: []*utils.RateSIncrement{
				{
					IncrementStart:    utils.NewDecimal(int64(4*time.Hour+31*time.Minute), 0),
					RateIntervalIndex: 0,
					RateID:            "UUID10",
					CompressFactor:    27,
					Usage:             utils.NewDecimal(int64(3*time.Hour+59*time.Minute), 0),
				},
				{
					IncrementStart:    utils.NewDecimal(int64(8*time.Hour+30*time.Minute), 0),
					RateIntervalIndex: 1,
					RateID:            "UUID11",
					CompressFactor:    24,
					Usage:             utils.NewDecimal(int64(3*time.Hour+30*time.Minute), 0),
				},
			},
			CompressFactor: 1,
		},
		{
			IntervalStart: utils.NewDecimal(int64(12*time.Hour), 0),
			Increments: []*utils.RateSIncrement{
				{
					IncrementStart:    utils.NewDecimal(int64(12*time.Hour), 0),
					RateIntervalIndex: 0,
					RateID:            "UUID_SPECIAL",
					CompressFactor:    3600,
					Usage:             utils.NewDecimal(int64(time.Hour), 0),
				},
			},
			CompressFactor: 1,
		},
		{
			IntervalStart: utils.NewDecimal(int64(13*time.Hour), 0),
			Increments: []*utils.RateSIncrement{
				{
					IncrementStart:    utils.NewDecimal(int64(13*time.Hour), 0),
					RateIntervalIndex: 1,
					RateID:            "UUID11",
					CompressFactor:    24,
					Usage:             utils.NewDecimal(int64(3*time.Hour+30*time.Minute), 0),
				},
				{
					IncrementStart:    utils.NewDecimal(int64(16*time.Hour+30*time.Minute), 0),
					RateIntervalIndex: 2,
					RateID:            "UUID12",
					CompressFactor:    27,
					Usage:             utils.NewDecimal(int64(4*time.Hour), 0),
				},
				{
					IncrementStart:    utils.NewDecimal(int64(20*time.Hour+30*time.Minute), 0),
					RateIntervalIndex: 3,
					RateID:            "UUID13",
					CompressFactor:    24,
					Usage:             utils.NewDecimal(int64(3*time.Hour+30*time.Minute), 0),
				},
			},
			CompressFactor: 1,
		},
		{
			IntervalStart: utils.NewDecimal(int64(24*time.Hour), 0),
			Increments: []*utils.RateSIncrement{
				{
					IncrementStart:    utils.NewDecimal(int64(24*time.Hour), 0),
					RateIntervalIndex: 1,
					RateID:            "UUID01",
					CompressFactor:    3600,
					Usage:             utils.NewDecimal(int64(time.Hour), 0),
				},
			},
			CompressFactor: 1,
		},
	}
	expCstRts := map[string]*utils.IntervalRate{
		"UUID00": {
			IntervalStart: utils.NewDecimal(0, 0),
			RecurrentFee:  utils.NewDecimal(5, 1),
			Unit:          twminDecimal,
			Increment:     minDecimal,
		},
		"UUID10": {
			IntervalStart: utils.NewDecimal(int64(4*time.Hour+30*time.Minute), 0),
			RecurrentFee:  utils.NewDecimal(45, 2),
			Unit:          twminDecimal,
			Increment:     nminDecimal,
		},
		"UUID11": {
			IntervalStart: utils.NewDecimal(int64(8*time.Hour+30*time.Minute), 0),
			RecurrentFee:  utils.NewDecimal(4, 1),
			Unit:          twminDecimal,
			Increment:     nminDecimal,
		},
		"UUID_SPECIAL": {
			IntervalStart: utils.NewDecimal(int64(12*time.Hour), 0),
			RecurrentFee:  utils.NewDecimal(2, 1),
			Unit:          twminDecimal,
			Increment:     secDecimal,
		},
		"UUID12": {
			IntervalStart: utils.NewDecimal(int64(16*time.Hour+30*time.Minute), 0),
			RecurrentFee:  utils.NewDecimal(35, 2),
			Unit:          twminDecimal,
			Increment:     nminDecimal,
		},
		"UUID13": {
			IntervalStart: utils.NewDecimal(int64(20*time.Hour+30*time.Minute), 0),
			RecurrentFee:  utils.NewDecimal(3, 1),
			Unit:          twminDecimal,
			Increment:     nminDecimal,
		},
		"UUID01": {
			IntervalStart: utils.NewDecimal(int64(24*time.Hour), 0),
			RecurrentFee:  utils.NewDecimal(1, 1),
			Unit:          twminDecimal,
			Increment:     secDecimal,
		},
	}
	cstRts := make(map[string]*utils.IntervalRate)
	if rtIvls, err := computeRateSIntervals(ordRts, utils.NewDecimal(0, 0).Big,
		utils.NewDecimal(int64(25*time.Hour), 0).Big, cstRts); err != nil {
		t.Error(err)
	} else {
		for idx, val := range rtIvls {
			if !val.Equals(expOrdRts[idx], cstRts, expCstRts) {
				t.Errorf("expecting: %+v \n,received: %+v", utils.ToJSON(expCstRts), utils.ToJSON(cstRts))
				t.Fatalf("expecting: %+v \n,received: %+v", utils.ToJSON(expOrdRts), utils.ToJSON(rtIvls))
			}
		}
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
	rt1 := &utils.Rate{
		ID: "RATE1",
		IntervalRates: []*utils.IntervalRate{
			{
				IntervalStart: utils.NewDecimal(0, 0),
			},
			{
				IntervalStart: utils.NewDecimal(int64(15*time.Minute), 0),
				RecurrentFee:  utils.NewDecimal(5, 1),
				Unit:          fminDecimal,
				Increment:     eminDecimal,
			},
			{
				IntervalStart: utils.NewDecimal(int64(30*time.Minute), 0),
				RecurrentFee:  utils.NewDecimal(4, 1),
				Unit:          fminDecimal,
				Increment:     nminDecimal,
			},
		},
	}

	rt2 := &utils.Rate{
		ID: "RATE2",
		IntervalRates: []*utils.IntervalRate{
			{
				IntervalStart: utils.NewDecimal(int64(45*time.Minute), 0),
				RecurrentFee:  utils.NewDecimal(3, 1),
				Unit:          fminDecimal,
				Increment:     sminDecimal,
			},
			{
				IntervalStart: utils.NewDecimal(int64(time.Hour), 0),
				RecurrentFee:  utils.NewDecimal(2, 1),
				Unit:          fminDecimal,
				Increment:     fvminDecimal,
			},
		},
	}

	rp := &utils.RateProfile{
		Tenant: "cgrates.org",
		ID:     "RATE_PROFILE",
		Rates: map[string]*utils.Rate{
			rt1.ID: rt1,
			rt2.ID: rt2,
		},
	}
	if err := rp.Compile(); err != nil {
		t.Error(err)
	}

	ordRts := []*orderedRate{
		{
			utils.NewDecimal(0, 0).Big,
			rt1,
		},
		{
			utils.NewDecimal(int64(45*time.Minute), 0).Big,
			rt2,
		},
	}

	expOrdRts := []*utils.RateSInterval{
		{
			IntervalStart: utils.NewDecimal(int64(15*time.Minute), 0),
			Increments: []*utils.RateSIncrement{
				{
					IncrementStart:    utils.NewDecimal(int64(15*time.Minute), 0),
					RateIntervalIndex: 1,
					RateID:            "UUID01",
					CompressFactor:    2,
					Usage:             utils.NewDecimal(int64(15*time.Minute), 0),
				},
				{
					IncrementStart:    utils.NewDecimal(int64(30*time.Minute), 0),
					RateIntervalIndex: 2,
					RateID:            "UUID02",
					CompressFactor:    2,
					Usage:             utils.NewDecimal(int64(15*time.Minute), 0),
				},
			},
			CompressFactor: 1,
		},
		{
			IntervalStart: utils.NewDecimal(int64(45*time.Minute), 0),
			Increments: []*utils.RateSIncrement{
				{
					IncrementStart:    utils.NewDecimal(int64(45*time.Minute), 0),
					RateIntervalIndex: 0,
					RateID:            "UUID10",
					CompressFactor:    3,
					Usage:             utils.NewDecimal(int64(15*time.Minute), 0),
				},
				{
					IncrementStart:    utils.NewDecimal(int64(time.Hour), 0),
					RateIntervalIndex: 1,
					RateID:            "UUID11",
					CompressFactor:    6,
					Usage:             utils.NewDecimal(int64(30*time.Minute), 0),
				},
			},
			CompressFactor: 1,
		},
	}
	expCstRts := map[string]*utils.IntervalRate{
		"UUID01": {
			IntervalStart: utils.NewDecimal(int64(15*time.Minute), 0),
			RecurrentFee:  utils.NewDecimal(5, 1),
			Unit:          fminDecimal,
			Increment:     eminDecimal,
		},
		"UUID02": {
			IntervalStart: utils.NewDecimal(int64(30*time.Minute), 0),
			RecurrentFee:  utils.NewDecimal(4, 1),
			Unit:          fminDecimal,
			Increment:     nminDecimal,
		},
		"UUID10": {
			IntervalStart: utils.NewDecimal(int64(45*time.Minute), 0),
			RecurrentFee:  utils.NewDecimal(3, 1),
			Unit:          fminDecimal,
			Increment:     sminDecimal,
		},
		"UUID11": {
			IntervalStart: utils.NewDecimal(int64(time.Hour), 0),
			RecurrentFee:  utils.NewDecimal(2, 1),
			Unit:          fminDecimal,
			Increment:     fvminDecimal,
		},
	}
	cstRts := make(map[string]*utils.IntervalRate)
	if rtIvls, err := computeRateSIntervals(ordRts, utils.NewDecimal(int64(15*time.Minute), 0).Big,
		utils.NewDecimal(int64(time.Hour+15*time.Minute), 0).Big, cstRts); err != nil {
		t.Error(err)
	} else {
		for idx, val := range rtIvls {
			if !val.Equals(expOrdRts[idx], cstRts, expCstRts) {
				t.Errorf("expecting: %+v \n,received: %+v", utils.ToJSON(expCstRts), utils.ToJSON(cstRts))
				t.Fatalf("expecting: %+v \n,received: %+v", utils.ToJSON(expOrdRts), utils.ToJSON(rtIvls))
			}
		}
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
	rt1 := &utils.Rate{
		ID: "RATE1",
		IntervalRates: []*utils.IntervalRate{
			{
				IntervalStart: utils.NewDecimal(0, 0),
				RecurrentFee:  utils.NewDecimal(5, 1),
				Unit:          tminDecimal,
				Increment:     tsecDecimal,
			},
			{
				IntervalStart: utils.NewDecimal(int64(12*time.Minute+35*time.Second), 0),
				RecurrentFee:  utils.NewDecimal(4, 1),
				Unit:          sminDecimal,
				Increment:     secDecimal,
			},
			{
				IntervalStart: utils.NewDecimal(int64(time.Hour+37*time.Minute+19*time.Second), 0),
				RecurrentFee:  utils.NewDecimal(2, 1),
				Unit:          nminDecimal,
				Increment:     sminDecimal,
			},
		},
	}
	rt1.Compile()

	rt2 := &utils.Rate{
		ID: "RATE2",
		IntervalRates: []*utils.IntervalRate{
			{
				IntervalStart: utils.NewDecimal(int64(38*time.Minute+15*time.Second), 0),
				RecurrentFee:  utils.NewDecimal(1, 1),
				Unit:          eminDecimal,
				Increment:     secDecimal,
			},
		},
	}
	rt2.Compile()

	ordRts := []*orderedRate{
		{
			utils.NewDecimal(0, 0).Big,
			rt1,
		},
		{
			utils.NewDecimal(int64(39*time.Minute), 0).Big,
			rt2,
		},
		{
			utils.NewDecimal(int64(time.Hour+37*time.Minute+19*time.Second), 0).Big,
			rt1,
		},
	}

	expOrdRts := []*utils.RateSInterval{
		{
			IntervalStart: utils.NewDecimal(0, 0),
			Increments: []*utils.RateSIncrement{
				{
					IncrementStart:    utils.NewDecimal(0, 0),
					RateIntervalIndex: 0,
					RateID:            "UUID00",
					CompressFactor:    378,
					Usage:             utils.NewDecimal(int64(12*time.Minute+35*time.Second), 0),
				},
				{
					IncrementStart:    utils.NewDecimal(int64(12*time.Minute+35*time.Second), 0),
					RateIntervalIndex: 1,
					RateID:            "UUID01",
					CompressFactor:    1585,
					Usage:             utils.NewDecimal(int64(26*time.Minute+25*time.Second), 0),
				},
			},
			CompressFactor: 1,
		},
		{
			IntervalStart: utils.NewDecimal(int64(39*time.Minute), 0),
			Increments: []*utils.RateSIncrement{
				{
					IncrementStart:    utils.NewDecimal(int64(39*time.Minute), 0),
					RateIntervalIndex: 0,
					RateID:            "UUID10",
					CompressFactor:    3499,
					Usage:             utils.NewDecimal(int64(58*time.Minute+19*time.Second), 0),
				},
			},
			CompressFactor: 1,
		},
		{
			IntervalStart: utils.NewDecimal(int64(time.Hour+37*time.Minute+19*time.Second), 0),
			Increments: []*utils.RateSIncrement{
				{
					IncrementStart:    utils.NewDecimal(int64(time.Hour+37*time.Minute+19*time.Second), 0),
					RateIntervalIndex: 2,
					RateID:            "UUID02",
					CompressFactor:    1,
					Usage:             utils.NewDecimal(int64(41*time.Second), 0),
				},
			},
			CompressFactor: 1,
		},
	}
	expCstRts := map[string]*utils.IntervalRate{
		"UUID00": {
			IntervalStart: utils.NewDecimal(0, 0),
			RecurrentFee:  utils.NewDecimal(5, 1),
			Unit:          tminDecimal,
			Increment:     tsecDecimal,
		},
		"UUID01": {
			IntervalStart: utils.NewDecimal(int64(12*time.Minute+35*time.Second), 0),
			RecurrentFee:  utils.NewDecimal(4, 1),
			Unit:          sminDecimal,
			Increment:     secDecimal,
		},
		"UUID10": {
			IntervalStart: utils.NewDecimal(int64(38*time.Minute+15*time.Second), 0),
			RecurrentFee:  utils.NewDecimal(1, 1),
			Unit:          eminDecimal,
			Increment:     secDecimal,
		},
		"UUID02": {
			IntervalStart: utils.NewDecimal(int64(time.Hour+37*time.Minute+19*time.Second), 0),
			RecurrentFee:  utils.NewDecimal(2, 1),
			Unit:          nminDecimal,
			Increment:     sminDecimal,
		},
	}
	cstRts := make(map[string]*utils.IntervalRate)
	if rtIvls, err := computeRateSIntervals(ordRts, utils.NewDecimal(0, 0).Big,
		utils.NewDecimal(int64(time.Hour+38*time.Minute), 0).Big, cstRts); err != nil {
		t.Error(err)
	} else {
		for idx, val := range rtIvls {
			if !val.Equals(expOrdRts[idx], cstRts, expCstRts) {
				t.Errorf("expecting: %+v \n,received: %+v", utils.ToJSON(expCstRts), utils.ToJSON(cstRts))
				t.Fatalf("expecting: %+v \n,received: %+v", utils.ToJSON(expOrdRts), utils.ToJSON(rtIvls))
			}
		}
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
	rt1 := &utils.Rate{
		ID: "RATE1",
		IntervalRates: []*utils.IntervalRate{
			{
				IntervalStart: utils.NewDecimal(0, 0),
				RecurrentFee:  utils.NewDecimal(5, 1),
				Unit:          tminDecimal,
				Increment:     tsecDecimal,
			},
		},
	}
	rt1.Compile()

	rt2 := &utils.Rate{
		ID: "RATE2",
		IntervalRates: []*utils.IntervalRate{
			{
				IntervalStart: utils.NewDecimal(int64(20*time.Minute), 0),
				RecurrentFee:  utils.NewDecimal(4, 1),
				Unit:          ttminDecimal,
				Increment:     tsecDecimal,
			},
			{
				IntervalStart: utils.NewDecimal(int64(40*time.Minute), 0),
				RecurrentFee:  utils.NewDecimal(3, 1),
				Unit:          ttminDecimal,
				Increment:     fsecDecimal,
			},
			{
				IntervalStart: utils.NewDecimal(int64(time.Hour), 0),
				RecurrentFee:  utils.NewDecimal(2, 1),
				Unit:          ttminDecimal,
				Increment:     tminDecimal,
			},
			{
				IntervalStart: utils.NewDecimal(int64(2*time.Hour), 0),
				RecurrentFee:  utils.NewDecimal(1, 1),
				Unit:          ttminDecimal,
				Increment:     fminDecimal,
			},
		},
	}
	rt2.Compile()

	ordRts := []*orderedRate{
		{
			utils.NewDecimal(0, 0).Big,
			rt1,
		},
		{
			utils.NewDecimal(int64(20*time.Minute), 0).Big,
			rt2,
		},
		{
			utils.NewDecimal(int64(21*time.Minute), 0).Big,
			rt1,
		},
		{
			utils.NewDecimal(int64(40*time.Minute), 0).Big,
			rt2,
		},
		{
			utils.NewDecimal(int64(41*time.Minute), 0).Big,
			rt1,
		},
		{
			utils.NewDecimal(int64(time.Hour), 0).Big,
			rt2,
		},
		{
			utils.NewDecimal(int64(time.Hour+time.Minute), 0).Big,
			rt1,
		},
		{
			utils.NewDecimal(int64(2*time.Hour), 0).Big,
			rt2,
		},
		{
			utils.NewDecimal(int64(2*time.Hour+time.Minute), 0).Big,
			rt1,
		},
	}

	expOrdRts := []*utils.RateSInterval{
		{
			IntervalStart: utils.NewDecimal(0, 0),
			Increments: []*utils.RateSIncrement{
				{
					IncrementStart:    utils.NewDecimal(0, 0),
					RateIntervalIndex: 0,
					RateID:            "UUID00",
					CompressFactor:    600,
					Usage:             utils.NewDecimal(int64(20*time.Minute), 0),
				},
			},
			CompressFactor: 1,
		},
		{
			IntervalStart: utils.NewDecimal(int64(20*time.Minute), 0),
			Increments: []*utils.RateSIncrement{
				{
					IncrementStart:    utils.NewDecimal(int64(20*time.Minute), 0),
					RateIntervalIndex: 0,
					RateID:            "UUID10",
					CompressFactor:    30,
					Usage:             utils.NewDecimal(int64(time.Minute), 0),
				},
			},
			CompressFactor: 1,
		},
		{
			IntervalStart: utils.NewDecimal(int64(21*time.Minute), 0),
			Increments: []*utils.RateSIncrement{
				{
					IncrementStart:    utils.NewDecimal(int64(21*time.Minute), 0),
					RateIntervalIndex: 0,
					RateID:            "UUID00",
					CompressFactor:    570,
					Usage:             utils.NewDecimal(int64(19*time.Minute), 0),
				},
			},
			CompressFactor: 1,
		},
		{
			IntervalStart: utils.NewDecimal(int64(40*time.Minute), 0),
			Increments: []*utils.RateSIncrement{
				{
					IncrementStart:    utils.NewDecimal(int64(40*time.Minute), 0),
					RateIntervalIndex: 1,
					RateID:            "UUID11",
					CompressFactor:    15,
					Usage:             utils.NewDecimal(int64(time.Minute), 0),
				},
			},
			CompressFactor: 1,
		},
		{
			IntervalStart: utils.NewDecimal(int64(41*time.Minute), 0),
			Increments: []*utils.RateSIncrement{
				{
					IncrementStart:    utils.NewDecimal(int64(41*time.Minute), 0),
					RateIntervalIndex: 0,
					RateID:            "UUID00",
					CompressFactor:    570,
					Usage:             utils.NewDecimal(int64(19*time.Minute), 0),
				},
			},
			CompressFactor: 1,
		},
		{
			IntervalStart: utils.NewDecimal(int64(time.Hour), 0),
			Increments: []*utils.RateSIncrement{
				{
					IncrementStart:    utils.NewDecimal(int64(time.Hour), 0),
					RateIntervalIndex: 2,
					RateID:            "UUID12",
					CompressFactor:    1,
					Usage:             utils.NewDecimal(int64(time.Minute), 0),
				},
			},
			CompressFactor: 1,
		},
		{
			IntervalStart: utils.NewDecimal(int64(time.Hour+time.Minute), 0),
			Increments: []*utils.RateSIncrement{
				{
					IncrementStart:    utils.NewDecimal(int64(time.Hour+time.Minute), 0),
					RateIntervalIndex: 0,
					RateID:            "UUID00",
					CompressFactor:    1770,
					Usage:             utils.NewDecimal(int64(59*time.Minute), 0),
				},
			},
			CompressFactor: 1,
		},
		{
			IntervalStart: utils.NewDecimal(int64(2*time.Hour), 0),
			Increments: []*utils.RateSIncrement{
				{
					IncrementStart:    utils.NewDecimal(int64(2*time.Hour), 0),
					RateIntervalIndex: 3,
					RateID:            "UUID13",
					CompressFactor:    1,
					Usage:             utils.NewDecimal(int64(time.Minute), 0),
				},
			},
			CompressFactor: 1,
		},
		{
			IntervalStart: utils.NewDecimal(int64(2*time.Hour+time.Minute), 0),
			Increments: []*utils.RateSIncrement{
				{
					IncrementStart:    utils.NewDecimal(int64(2*time.Hour+time.Minute), 0),
					RateIntervalIndex: 0,
					RateID:            "UUID00",
					CompressFactor:    30,
					Usage:             utils.NewDecimal(int64(time.Minute), 0),
				},
			},
			CompressFactor: 1,
		},
	}
	expCstRts := map[string]*utils.IntervalRate{
		"UUID00": {
			IntervalStart: utils.NewDecimal(0, 0),
			RecurrentFee:  utils.NewDecimal(5, 1),
			Unit:          tminDecimal,
			Increment:     tsecDecimal,
		},
		"UUID10": {
			IntervalStart: utils.NewDecimal(int64(20*time.Minute), 0),
			RecurrentFee:  utils.NewDecimal(4, 1),
			Unit:          ttminDecimal,
			Increment:     tsecDecimal,
		},
		"UUID11": {
			IntervalStart: utils.NewDecimal(int64(40*time.Minute), 0),
			RecurrentFee:  utils.NewDecimal(3, 1),
			Unit:          ttminDecimal,
			Increment:     fsecDecimal,
		},
		"UUID12": {
			IntervalStart: utils.NewDecimal(int64(time.Hour), 0),
			RecurrentFee:  utils.NewDecimal(2, 1),
			Unit:          ttminDecimal,
			Increment:     tminDecimal,
		},
		"UUID13": {
			IntervalStart: utils.NewDecimal(int64(2*time.Hour), 0),
			RecurrentFee:  utils.NewDecimal(1, 1),
			Unit:          ttminDecimal,
			Increment:     fminDecimal,
		},
	}
	cstRts := make(map[string]*utils.IntervalRate)
	if rtIvls, err := computeRateSIntervals(ordRts, utils.NewDecimal(0, 0).Big,
		utils.NewDecimal(int64(2*time.Hour+2*time.Minute), 0).Big, cstRts); err != nil {
		t.Error(err)
	} else {
		for idx, val := range rtIvls {
			if !val.Equals(expOrdRts[idx], cstRts, expCstRts) {
				t.Errorf("expecting: %+v \n,received: %+v", utils.ToJSON(expCstRts), utils.ToJSON(cstRts))
				t.Fatalf("expecting: %+v \n,received: %+v", utils.ToJSON(expOrdRts), utils.ToJSON(rtIvls))
			}
		}
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
	rt1 := &utils.Rate{
		ID: "RATE1",
		IntervalRates: []*utils.IntervalRate{
			{
				IntervalStart: utils.NewDecimal(0, 0),
				RecurrentFee:  utils.NewDecimal(5, 1),
				Unit:          hourDecimal,
				Increment:     fminDecimal,
			},
		},
	}
	rt1.Compile()

	rt2 := &utils.Rate{
		ID: "RATE2",
		IntervalRates: []*utils.IntervalRate{
			{
				IntervalStart: utils.NewDecimal(int64(168*time.Hour), 0),
				RecurrentFee:  utils.NewDecimal(5, 1),
				Unit:          minDecimal,
				Increment:     minDecimal,
			},
		},
	}
	rt2.Compile()

	ordRts := []*orderedRate{
		{
			utils.NewDecimal(0, 0).Big,
			rt1,
		},
		{
			utils.NewDecimal(int64(168*time.Hour), 0).Big,
			rt2,
		},
	}

	expOrdRts := []*utils.RateSInterval{
		{
			IntervalStart: utils.NewDecimal(0, 0),
			Increments: []*utils.RateSIncrement{
				{
					IncrementStart:    utils.NewDecimal(0, 0),
					RateIntervalIndex: 0,
					RateID:            "UUID00",
					CompressFactor:    593,
					Usage:             utils.NewDecimal(int64(168*time.Hour), 0),
				},
			},
			CompressFactor: 1,
		},
		{
			IntervalStart: utils.NewDecimal(int64(168*time.Hour), 0),
			Increments: []*utils.RateSIncrement{
				{
					IncrementStart:    utils.NewDecimal(int64(168*time.Hour), 0),
					RateIntervalIndex: 0,
					RateID:            "UUID10",
					CompressFactor:    60,
					Usage:             utils.NewDecimal(int64(time.Hour), 0),
				},
			},
			CompressFactor: 1,
		},
	}
	expCstRts := map[string]*utils.IntervalRate{
		"UUID00": {
			IntervalStart: utils.NewDecimal(0, 0),
			RecurrentFee:  utils.NewDecimal(5, 1),
			Unit:          hourDecimal,
			Increment:     fminDecimal,
		},
		"UUID10": {
			IntervalStart: utils.NewDecimal(int64(168*time.Hour), 0),
			RecurrentFee:  utils.NewDecimal(5, 1),
			Unit:          minDecimal,
			Increment:     minDecimal,
		},
	}
	cstRts := make(map[string]*utils.IntervalRate)
	if rtIvls, err := computeRateSIntervals(ordRts, utils.NewDecimal(0, 0).Big,
		utils.NewDecimal(int64(169*time.Hour), 0).Big, cstRts); err != nil {
		t.Error(err)
	} else {
		for idx, val := range rtIvls {
			if !val.Equals(expOrdRts[idx], cstRts, expCstRts) {
				t.Errorf("expecting: %+v \n,received: %+v", utils.ToJSON(expCstRts), utils.ToJSON(cstRts))
				t.Fatalf("expecting: %+v \n,received: %+v", utils.ToJSON(expOrdRts), utils.ToJSON(rtIvls))
			}
		}
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
	rt1 := &utils.Rate{
		ID: "RATE1",
		IntervalRates: []*utils.IntervalRate{
			{
				IntervalStart: utils.NewDecimal(0, 0),
				RecurrentFee:  utils.NewDecimal(2, 1),
				Unit:          fminDecimal,
				Increment:     minDecimal,
			},
			{
				IntervalStart: utils.NewDecimal(int64(20*time.Minute), 0),
				RecurrentFee:  utils.NewDecimal(4, 1),
				Unit:          minDecimal,
				Increment:     secDecimal,
			},
		},
	}

	rt2 := &utils.Rate{
		ID: "RATE2",
		IntervalRates: []*utils.IntervalRate{
			{
				IntervalStart: utils.NewDecimal(int64(50*time.Minute), 0),
				RecurrentFee:  utils.NewDecimal(1, 1),
				Unit:          tminDecimal,
				Increment:     minDecimal,
			},
			{
				IntervalStart: utils.NewDecimal(int64(time.Hour), 0),
				RecurrentFee:  utils.NewDecimal(5, 3),
				Unit:          fminDecimal,
				Increment:     secDecimal,
			},
		},
	}

	ordRts := []*orderedRate{
		{
			utils.NewDecimal(0, 0).Big,
			rt1,
		},
		{
			utils.NewDecimal(int64(time.Hour), 0).Big,
			rt2,
		},
	}

	expOrdRts := []*utils.RateSInterval{
		{
			IntervalStart: utils.NewDecimal(0, 0),
			Increments: []*utils.RateSIncrement{
				{
					IncrementStart:    utils.NewDecimal(0, 0),
					RateIntervalIndex: 0,
					RateID:            "UUID00",
					CompressFactor:    20,
					Usage:             utils.NewDecimal(int64(20*time.Minute), 0),
				},
				{
					IncrementStart:    utils.NewDecimal(int64(20*time.Minute), 0),
					RateIntervalIndex: 1,
					RateID:            "UUID01",
					CompressFactor:    2400,
					Usage:             utils.NewDecimal(int64(40*time.Minute), 0),
				},
			},
			CompressFactor: 1,
		},
		{
			IntervalStart: utils.NewDecimal(int64(time.Hour), 0),
			Increments: []*utils.RateSIncrement{
				{
					IncrementStart:    utils.NewDecimal(int64(time.Hour), 0),
					RateIntervalIndex: 1,
					RateID:            "UUID10",
					CompressFactor:    1200,
					Usage:             utils.NewDecimal(int64(20*time.Minute), 0),
				},
			},
			CompressFactor: 1,
		},
	}
	expCstRts := map[string]*utils.IntervalRate{
		"UUID00": {
			IntervalStart: utils.NewDecimal(0, 0),
			RecurrentFee:  utils.NewDecimal(2, 1),
			Unit:          fminDecimal,
			Increment:     minDecimal,
		},
		"UUID01": {
			IntervalStart: utils.NewDecimal(int64(20*time.Minute), 0),
			RecurrentFee:  utils.NewDecimal(4, 1),
			Unit:          minDecimal,
			Increment:     secDecimal,
		},
		"UUID10": {
			IntervalStart: utils.NewDecimal(int64(time.Hour), 0),
			RecurrentFee:  utils.NewDecimal(5, 3),
			Unit:          fminDecimal,
			Increment:     secDecimal,
		},
	}
	cstRts := make(map[string]*utils.IntervalRate)
	if rtIvls, err := computeRateSIntervals(ordRts, utils.NewDecimal(0, 0).Big,
		utils.NewDecimal(int64(time.Hour+20*time.Minute), 0).Big, cstRts); err != nil {
		t.Error(err)
	} else {
		for idx, val := range rtIvls {
			if !val.Equals(expOrdRts[idx], cstRts, expCstRts) {
				t.Errorf("expecting: %+v \n,received: %+v", utils.ToJSON(expCstRts), utils.ToJSON(cstRts))
				t.Fatalf("expecting: %+v \n,received: %+v", utils.ToJSON(expOrdRts), utils.ToJSON(rtIvls))
			}
		}
	}
}

func TestOrderRatesOnIntervalsErrorConvert(t *testing.T) {
	rt1 := &utils.Rate{
		ID:              "RT_1",
		ActivationTimes: "* * 1 * *",
		IntervalRates: []*utils.IntervalRate{
			{
				IntervalStart: utils.NewDecimal(int64(48*time.Hour), 0),
			},
		},
	}
	err := rt1.Compile()
	if err != nil {
		t.Error(err)
	}

	wghts := []float64{0}
	aRts := []*utils.Rate{rt1}
	sTime := time.Date(2020, 12, 1, 0, 0, 0, 0, time.UTC)
	usage := utils.NewDecimalFromFloat64(math.Inf(1))
	expected := "<RateS> cannot convert <&{Context:{MaxScale:0 MinScale:0 Precision:0 Traps: Conditions: RoundingMode:ToNearestEven OperatingMode:GDA} unscaled:{neg:false abs:[]} compact:0 exp:0 precision:0 form:2}> increment to Int64"
	_, err = orderRatesOnIntervals(aRts, wghts, sTime, usage.Big, false, 4)
	if err == nil || err.Error() != expected {
		t.Error(err)
	}
}

func TestComputeRateSIntervalsRecurrentFee(t *testing.T) {
	tsecDecimal, err := utils.NewDecimalFromUsage("30s")
	if err != nil {
		t.Error(err)
	}
	secDecimal, err := utils.NewDecimalFromUsage("1s")
	if err != nil {
		t.Error(err)
	}
	rt1 := &utils.Rate{
		ID: "RATE1",
		IntervalRates: []*utils.IntervalRate{
			{
				IntervalStart: utils.NewDecimal(0, 0),
				RecurrentFee:  nil,
				Unit:          tsecDecimal,
				Increment:     secDecimal,
			},
		},
	}
	rt1.Compile()

	ordRts := []*orderedRate{
		{
			utils.NewDecimal(0, 0).Big,
			rt1,
		},
	}
	expOrdRts := []*utils.RateSInterval{
		{
			IntervalStart: utils.NewDecimal(0, 0),
			Increments: []*utils.RateSIncrement{
				{
					IncrementStart:    utils.NewDecimal(0, 0),
					RateIntervalIndex: 0,
					RateID:            "UUID00",
					CompressFactor:    20,
					Usage:             utils.NewDecimal(int64(time.Minute+10*time.Second), 0),
				},
			},
			CompressFactor: 1,
		},
	}
	expCstRts := map[string]*utils.IntervalRate{
		"UUID00": {
			IntervalStart: utils.NewDecimal(0, 0),
			RecurrentFee:  nil,
			Unit:          tsecDecimal,
			Increment:     secDecimal,
		},
	}
	cstRts := make(map[string]*utils.IntervalRate)
	if rtIvls, err := computeRateSIntervals(ordRts, utils.NewDecimal(0, 0).Big,
		utils.NewDecimal(int64(time.Minute+10*time.Second), 0).Big, cstRts); err != nil {
		t.Error(err)
	} else {
		for idx, val := range rtIvls {
			if !val.Equals(expOrdRts[idx], cstRts, expCstRts) {
				t.Errorf("expecting: %+v \n,received: %+v", utils.ToJSON(expCstRts), utils.ToJSON(cstRts))
				t.Fatalf("expecting: %+v \n,received: %+v", utils.ToJSON(expOrdRts), utils.ToJSON(rtIvls))
			}
		}
	}
}

func TestComputeRateSIntervalsRecurrentFeeCmpFactorIntInvalidError(t *testing.T) {
	fminDecimal, err := utils.NewDecimalFromUsage("5m")
	if err != nil {
		t.Error(err)
	}
	rt1 := &utils.Rate{
		ID: "RATE1",
		IntervalRates: []*utils.IntervalRate{
			{
				IntervalStart: utils.NewDecimal(0, 0),
				RecurrentFee:  utils.NewDecimal(2, 1),
				Unit:          fminDecimal,
				Increment:     utils.NewDecimal(1, 0),
			},
		},
	}

	ordRts := []*orderedRate{
		{
			utils.NewDecimal(0, 0).Big,
			rt1,
		},
	}

	expected := "<RateS> cannot convert <&{Context:{MaxScale:0 MinScale:0 Precision:0 Traps: Conditions: RoundingMode:ToNearestEven OperatingMode:GDA} unscaled:{neg:false abs:[]} compact:0 exp:0 precision:0 form:2}> increment to Int64"
	_, err = computeRateSIntervals(ordRts, utils.NewDecimal(0, 0).Big,
		utils.NewDecimalFromFloat64(math.Inf(1)).Big, make(map[string]*utils.IntervalRate))
	if err == nil || err.Error() != expected {
		t.Error(err)
	}
}
