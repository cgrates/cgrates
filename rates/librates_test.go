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
				IntervalStart: 0,
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

	sTime := time.Date(2020, time.June, 28, 18, 56, 05, 0, time.UTC)
	usage := 2 * time.Minute
	if ordRts, err := orderRatesOnIntervals(
		allRts, sTime, usage, true, 10); err != nil {
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
		allRts, sTime, usage, true, 10); err != nil {
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
		allRts, sTime, usage, true, 10); err != nil {
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
		rts, sTime, usage, true, 10); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expOrdered, ordRts) {
		t.Errorf("expecting: %s\n, received: %s",
			utils.ToIJSON(expOrdered), utils.ToIJSON(ordRts))
	}
}

func TestOrderRatesOnIntervalsChristmasDay(t *testing.T) {
	rt1 := &engine.Rate{
		ID:     "ALWAYS_RATE",
		Weight: 10,
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
			},
		},
	}
	rtCh1 := &engine.Rate{
		ID:              "CHRISTMAS1",
		ActivationTimes: "* 0-6 24 12 *",
		Weight:          20,
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
			},
		},
	}
	rtCh2 := &engine.Rate{
		ID:              "CHRISTMAS2",
		ActivationTimes: "* 7-12 24 12 *",
		Weight:          20,
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
			},
		},
	}
	rtCh3 := &engine.Rate{
		ID:              "CHRISTMAS3",
		ActivationTimes: "* 13-19 24 12 *",
		Weight:          20,
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
			},
		},
	}
	rtCH4 := &engine.Rate{
		ID:              "CHRISTMAS4",
		ActivationTimes: "* 20-23 24 12 *",
		Weight:          20,
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
			},
		},
	}
	allRates := []*engine.Rate{rt1, rtCh1, rtCh2, rtCh3, rtCH4}
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
			rtCH4,
		},
		{
			26 * time.Hour,
			rt1,
		},
	}
	if ordRts, err := orderRatesOnIntervals(allRates, sTime, usage, true, 10); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(ordRts, expected) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(ordRts))
	}
}

func TestOrderRatesOnIntervalsDoubleRates1(t *testing.T) {
	rt1 := &engine.Rate{
		ID:     "ALWAYS_RATE",
		Weight: 10,
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
			},
		},
	}
	rtCh1 := &engine.Rate{
		ID:              "CHRISTMAS1",
		ActivationTimes: "* * 24 12 *",
		Weight:          20,
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
			},
		},
	}
	rtCh2 := &engine.Rate{
		ID:              "CHRISTMAS2",
		ActivationTimes: "* 18-23 24 12 *",
		Weight:          30,
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
			},
		},
	}
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
	if ordRts, err := orderRatesOnIntervals(allRates, sTime, usage, true, 10); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(ordRts, expected) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(ordRts))
	}
}

func TestOrderRatesOnIntervalsEveryTwentyFiveMins(t *testing.T) {
	rtTwentyFiveMins := &engine.Rate{
		ID:              "TWENTYFIVE_MINS",
		ActivationTimes: "*/25 * * * *",
		Weight:          20,
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
			},
		},
	}
	rt1 := &engine.Rate{
		ID:              "DAY_RATE",
		ActivationTimes: "* * * * 3",
		Weight:          10,
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
			},
		},
	}
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
	if ordRts, err := orderRatesOnIntervals(allRates, sTime, usage, true, 10); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(ordRts, expected) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(ordRts))
	}
}

func TestOrderRatesOnIntervalsOneMinutePause(t *testing.T) {
	rt1 := &engine.Rate{
		ID:              "ALWAYS_RATE",
		ActivationTimes: "26 * * * *",
		Weight:          10,
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
			},
		},
	}
	rtFirstInterval := &engine.Rate{
		ID:              "FIRST_INTERVAL",
		ActivationTimes: "0-25 * * * *",
		Weight:          20,
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
			},
		},
	}
	rtSecondINterval := &engine.Rate{
		ID:              "SECOND_INTERVAL",
		ActivationTimes: "27-59 * * * *",
		Weight:          20,
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
			},
		},
	}
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
	if ordRts, err := orderRatesOnIntervals(allRates, sTime, usage, true, 10); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(ordRts, expected) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(ordRts))
	}
}

func TestOrderRatesOnIntervalsNewYear(t *testing.T) {
	rt1 := &engine.Rate{
		ID:     "ALWAYS_RATE",
		Weight: 10,
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
			},
		},
	}
	rt1NewYear := &engine.Rate{
		ID:              "NEW_YEAR1",
		ActivationTimes: "* 20-23 * * *",
		Weight:          20,
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
			},
		},
	}
	rt1NewYear2 := &engine.Rate{
		ID:              "NEW_YEAR2",
		ActivationTimes: "0-30 22 * * *",
		Weight:          30,
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
			},
		},
	}
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
	if ordRts, err := orderRatesOnIntervals(allRates, sTime, usage, true, 10); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(ordRts, expected) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(ordRts))
	}
}

func TestOrderRateOnIntervalsEveryHourEveryDay(t *testing.T) {
	rtEveryHour := &engine.Rate{
		ID:              "HOUR_RATE",
		ActivationTimes: "* */1 * * *",
		Weight:          10,
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
			},
		},
	}
	rtEveryDay := &engine.Rate{
		ID:              "DAY_RATE",
		ActivationTimes: "* * 22 * *",
		Weight:          20,
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
			},
		},
	}
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
	if ordRts, err := orderRatesOnIntervals(allRates, sTime, usage, true, 10); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(ordRts, expected) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(ordRts))
	}
}

func TestOrderRatesOnIntervalsOneHourInThreeRates(t *testing.T) {
	rtOneHour1 := &engine.Rate{
		ID:              "HOUR_RATE_1",
		ActivationTimes: "0-19 * * * *",
		Weight:          20,
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
			},
		},
	}
	rtOneHour2 := &engine.Rate{
		ID:              "HOUR_RATE_2",
		ActivationTimes: "20-39 * * * *",
		Weight:          20,
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
			},
		},
	}
	rtOneHour3 := &engine.Rate{
		ID:              "HOUR_RATE_3",
		ActivationTimes: "40-59 * * * *",
		Weight:          20,
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
			},
		},
	}
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
	if ordRts, err := orderRatesOnIntervals(allRates, sTime, usage, true, 10); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(ordRts, expected) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(ordRts))
	}
}

func TestOrderRateOnIntervalsEveryThreeHours(t *testing.T) {
	rtEveryThreeH := &engine.Rate{
		ID:              "EVERY_THREE_RATE",
		ActivationTimes: "* */3 * * *",
		Weight:          10,
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
			},
		},
	}
	rtByDay := &engine.Rate{
		ID:              "DAY_RATE",
		ActivationTimes: "* 15-23 * * *",
		Weight:          20,
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
			},
		},
	}
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
	if ordRts, err := orderRatesOnIntervals(allRates, sTime, usage, true, 10); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(ordRts, expected) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(ordRts))
	}
}

func TestOrderRateOnIntervalsTwoRatesInOne(t *testing.T) {
	rtHalfDay1 := &engine.Rate{
		ID:              "HALF_RATE1",
		ActivationTimes: "* 0-11 22 12 *",
		Weight:          10,
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
			},
		},
	}
	rtHalfDay2 := &engine.Rate{
		ID:              "HALF_RATE2",
		ActivationTimes: "* 12-23 22 12 *",
		Weight:          10,
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
			},
		},
	}
	rtHalfDay2r1 := &engine.Rate{
		ID:              "HALF_RATE2.1",
		ActivationTimes: "* 12-16 22 12 *",
		Weight:          20,
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
			},
		},
	}
	rtHalfDay2r2 := &engine.Rate{
		ID:              "HALF_RATE2.2",
		ActivationTimes: "* 18-23 22 12 *",
		Weight:          20,
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
			},
		},
	}
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
	if ordRts, err := orderRatesOnIntervals(allRates, sTime, usage, true, 10); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(ordRts, expected) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(ordRts))
	}
}

func TestOrderRateOnIntervalsEvery1Hour30Mins(t *testing.T) {
	rateEvery1H := &engine.Rate{
		ID:              "HOUR_RATE",
		ActivationTimes: "* */1 * * *",
		Weight:          10,
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
			},
		},
	}
	rateEvery30Mins := &engine.Rate{
		ID:              "MINUTES_RATE",
		ActivationTimes: "*/30 * * * *",
		Weight:          20,
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
			},
		},
	}
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
	if ordRts, err := orderRatesOnIntervals(allRates, sTime, usage, true, 10); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(ordRts, expected) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(ordRts))
	}
}

func TestOrderRatesOnIntervalsOnePrinciapalRateCase1(t *testing.T) {
	rtPrincipal := &engine.Rate{
		ID:              "PRINCIPAL_RATE",
		ActivationTimes: "* 10-22 * * *",
		Weight:          10,
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
			},
		},
	}
	rt1 := &engine.Rate{
		ID:              "RT1",
		ActivationTimes: "* 10-18 * * *",
		Weight:          20,
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
			},
		},
	}
	rt2 := &engine.Rate{
		ID:              "RT2",
		ActivationTimes: "* 10-16 * * *",
		Weight:          30,
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
			},
		},
	}
	rt3 := &engine.Rate{
		ID:              "RT3",
		ActivationTimes: "* 10-14 * * *",
		Weight:          40,
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
			},
		},
	}
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
	if ordRts, err := orderRatesOnIntervals(allRates, sTime, usage, true, 10); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(ordRts, expected) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(ordRts))
	}
}

func TestOrderRatesOnIntervalsOnePrinciapalRateCase2(t *testing.T) {
	rtPrincipal := &engine.Rate{
		ID:              "PRINCIPAL_RATE",
		ActivationTimes: "* 10-22 * * *",
		Weight:          10,
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
			},
		},
	}
	rt1 := &engine.Rate{
		ID:              "RT1",
		ActivationTimes: "* 18-22 * * *",
		Weight:          40,
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
			},
		},
	}
	rt2 := &engine.Rate{
		ID:              "RT2",
		ActivationTimes: "* 16-22 * * *",
		Weight:          30,
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
			},
		},
	}
	rt3 := &engine.Rate{
		ID:              "RT3",
		ActivationTimes: "* 14-22 * * *",
		Weight:          20,
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
			},
		},
	}
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
	if ordRts, err := orderRatesOnIntervals(allRates, sTime, usage, true, 10); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(ordRts, expected) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(ordRts))
	}
}

func TestOrderRatesOnIntervalsEvenOddMinutes(t *testing.T) {
	rtOddMInutes := &engine.Rate{
		ID:              "ODD_RATE",
		ActivationTimes: "*/1 * * * *",
		Weight:          10,
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
			},
		},
	}
	rtEvenMinutes := &engine.Rate{
		ID:              "EVEN_RATE",
		ActivationTimes: "*/2 * * * *",
		Weight:          20,
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
			},
		},
	}
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
	if ordRts, err := orderRatesOnIntervals(allRates, sTime, usage, true, 10); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(ordRts, expected) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(ordRts))
	}
}

func TestOrderRatesOnIntervalsDoubleRates2(t *testing.T) {
	rt1 := &engine.Rate{
		ID:     "ALWAYS_RATE",
		Weight: 10,
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
			},
		},
	}
	rtCh1 := &engine.Rate{
		ID:              "CHRISTMAS1",
		ActivationTimes: "* * 24 12 *",
		Weight:          20,
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
			},
		},
	}
	rtCh2 := &engine.Rate{
		ID:              "CHRISTMAS2",
		ActivationTimes: "* 10-12 24 12 *",
		Weight:          30,
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
			},
		},
	}
	rtCh3 := &engine.Rate{
		ID:              "CHRISTMAS3",
		ActivationTimes: "* 20-22 24 12 *",
		Weight:          30,
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
			},
		},
	}
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
	if ordRts, err := orderRatesOnIntervals(allRates, sTime, usage, true, 10); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(ordRts, expected) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(ordRts))
	}
}

func TestOrderOnRatesIntervalsEveryTwoHours(t *testing.T) {
	rt1 := &engine.Rate{
		ID:     "ALWAYS_RATE",
		Weight: 10,
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
			},
		},
	}
	rtEvTwoHours := &engine.Rate{
		ID:              "EVERY_TWO_HOURS",
		Weight:          20,
		ActivationTimes: "* */2 * * *",
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
			},
		},
	}
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
	if ordRts, err := orderRatesOnIntervals(allRates, sTime, usage, true, 10); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(ordRts, expected) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(ordRts))
	}
}

func TestOrderRatesOnIntervalsEveryTwoDays(t *testing.T) {
	rt1 := &engine.Rate{
		ID:     "ALWAYS_RATE",
		Weight: 10,
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
			},
		},
	}
	rtEveryTwoDays := &engine.Rate{
		ID:              "RATE_EVERY_DAY",
		ActivationTimes: "* * */2 * *",
		Weight:          20,
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
			},
		},
	}
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
	if ordRts, err := orderRatesOnIntervals(allRates, sTime, usage, true, 10); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(ordRts, expected) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(ordRts))
	}
}

func TestOrderRatesOnIntervalsSpecialHour(t *testing.T) {
	rtRestricted := &engine.Rate{
		ID:              "RESTRICTED",
		ActivationTimes: "* 10-22 * * *",
		Weight:          20,
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
			},
		},
	}
	rtWayRestricted := &engine.Rate{
		ID:              "WAY_RESTRICTED",
		ActivationTimes: "* 12-14 * * *",
		Weight:          30,
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
			},
		},
	}
	rtSpecialHour := &engine.Rate{
		ID:              "SPECIAL_HOUR",
		ActivationTimes: "* 13 * * *",
		Weight:          40,
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
			},
		},
	}
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
	if ordRts, err := orderRatesOnIntervals(allRts, sTime, usage, true, 10); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(ordRts, expected) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(ordRts))
	}
}

func TestOrderRateIntervalsRateEveryTenMinutes(t *testing.T) {
	rt1 := &engine.Rate{
		ID:              "DAY_RATE",
		ActivationTimes: "* * 21 7 *",
		Weight:          10,
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
			},
		},
	}
	rtEveryTenMin := &engine.Rate{
		ID:              "EVERY_TEN_MIN",
		ActivationTimes: "*/20 * * * *",
		Weight:          20,
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
			},
		},
	}
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
	if ordRts, err := orderRatesOnIntervals(allRts, sTime, usage, true, 10); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(ordRts, expected) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(ordRts))
	}
}

func TestOrderRatesOnIntervalsDayOfTheWeek(t *testing.T) {
	rt1 := &engine.Rate{
		ID:     "ALWAYS_RATE",
		Weight: 10,
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
			},
		},
	}
	rtDay := &engine.Rate{
		ID:              "DAY_RATE",
		ActivationTimes: "* * 21 7 2",
		Weight:          20,
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
			},
		},
	}
	rtDay1 := &engine.Rate{
		ID:              "DAY_RATE1",
		ActivationTimes: "* 15 21 7 2",
		Weight:          30,
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
			},
		},
	}
	rtDay2 := &engine.Rate{
		ID:              "DAY_RATE2",
		ActivationTimes: "* 18 21 7 2",
		Weight:          30,
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
			},
		},
	}
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
	if ordRts, err := orderRatesOnIntervals(allRates, sTime, usage, true, 10); err != nil {
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
	sTime := time.Date(2020, 12, 1, 0, 0, 0, 0, time.UTC)
	usage := 48 * time.Hour
	if ordRts, err := orderRatesOnIntervals(aRts, sTime, usage, false, 4); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(ordRts, expected) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(expected), utils.ToJSON(ordRts))
	}
}

func TestComputeRateSIntervals(t *testing.T) {
	rt0 := &engine.Rate{
		ID: "RATE0",
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
				RecurrentFee:  0.10,
				Unit:          1 * time.Minute,
				Increment:     1 * time.Minute,
			},
			{
				IntervalStart: 60 * time.Second,
				RecurrentFee:  0.05,
				Unit:          1 * time.Minute,
				Increment:     1 * time.Second,
			},
		},
	}
	rt0.Compile()

	rt1 := &engine.Rate{
		ID: "RATE1",
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
				RecurrentFee:  0.20,
				Unit:          1 * time.Minute,
				Increment:     1 * time.Second,
			},
			{
				IntervalStart: 2 * time.Minute,
				RecurrentFee:  0.15,
				Unit:          1 * time.Minute,
				Increment:     1 * time.Second,
			},
		},
	}
	rt1.Compile()

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
	rt0 := &engine.Rate{
		ID: "RATE0",
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
				RecurrentFee:  0.20,
				Unit:          30 * time.Second,
				Increment:     30 * time.Second,
			},
			{
				IntervalStart: 30 * time.Second,
				RecurrentFee:  0.15,
				Unit:          1 * time.Minute,
				Increment:     1 * time.Second,
			},
		},
	}
	rt0.Compile()

	rt1 := &engine.Rate{
		ID: "RATE1",
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
				RecurrentFee:  0.20,
				Unit:          1 * time.Minute,
				Increment:     1 * time.Second,
			},
			{
				IntervalStart: 2 * time.Minute,
				RecurrentFee:  0.15,
				Unit:          time.Minute,
				Increment:     1 * time.Second,
			},
		},
	}
	rt1.Compile()

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
	rt0 := &engine.Rate{
		ID: "RATE0",
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
				FixedFee:      0.123,
				RecurrentFee:  0.20,
				Unit:          30 * time.Second,
				Increment:     30 * time.Second,
			},
			{
				IntervalStart: 30 * time.Second,
				RecurrentFee:  0.15,
				Unit:          1 * time.Minute,
				Increment:     1 * time.Second,
			},
		},
	}
	rt0.Compile()

	rt1 := &engine.Rate{
		ID: "RATE1",
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
				FixedFee:      0.567,
				RecurrentFee:  0.20,
				Unit:          1 * time.Minute,
				Increment:     1 * time.Second,
			},
			{
				IntervalStart: 2 * time.Minute,
				RecurrentFee:  0.15,
				Unit:          time.Minute,
				Increment:     1 * time.Second,
			},
		},
	}
	rt1.Compile()

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
	rt0 := &engine.Rate{
		ID:              "RATE0",
		ActivationTimes: "* * * * *",
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
				RecurrentFee:  1.00,
				Unit:          time.Minute,
				Increment:     time.Minute,
			},
			{
				IntervalStart: 50 * time.Minute,
				RecurrentFee:  0.50,
				Unit:          time.Minute,
				Increment:     time.Minute,
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
				RecurrentFee:  0.15,
				Unit:          time.Minute,
				Increment:     time.Minute,
			},
			{
				IntervalStart: 45 * time.Minute,
				RecurrentFee:  0.20,
				Unit:          time.Minute,
				Increment:     time.Minute,
			},
		},
	}
	rt1.Compile()
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
	if rcvOrdRts, err := orderRatesOnIntervals(allRates, sTime, usage, true, 10); err != nil {
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
	rt1 := &engine.Rate{
		ID: "RATE1",
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
				RecurrentFee:  0.10,
				Unit:          30 * time.Second,
				Increment:     time.Second,
			},
			{
				IntervalStart: time.Minute,
				RecurrentFee:  0.20,
				Unit:          30 * time.Second,
				Increment:     time.Second,
			},
			{
				IntervalStart: 2 * time.Minute,
				RecurrentFee:  0.30,
				Unit:          30 * time.Second,
				Increment:     time.Second,
			},
		},
	}
	rt1.Compile()

	rt2 := &engine.Rate{
		ID: "RATE2",
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 30 * time.Second,
				RecurrentFee:  0.15,
				Unit:          30 * time.Second,
				Increment:     time.Second,
			},
			{
				IntervalStart: time.Minute + 30*time.Second,
				RecurrentFee:  0.25,
				Unit:          30 * time.Second,
				Increment:     time.Second,
			},
			{
				IntervalStart: 2*time.Minute + 30*time.Second,
				RecurrentFee:  0.35,
				Unit:          30 * time.Second,
				Increment:     time.Second,
			},
		},
	}
	rt1.Compile()

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
	rt1 := &engine.Rate{
		ID: "RATE1",
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: time.Minute,
				RecurrentFee:  0.10,
				Unit:          30 * time.Second,
				Increment:     30 * time.Second,
			},
		},
	}
	rt1.Compile()

	rt2 := &engine.Rate{
		ID: "RATE2",
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 2 * time.Minute,
				RecurrentFee:  0.20,
				Unit:          30 * time.Second,
				Increment:     30 * time.Second,
			},
		},
	}
	rt2.Compile()
	ordRts := []*orderedRate{
		{
			0,
			rt1,
		},
	}

	expected := "intervalStart for rate: <> higher than usage: 0s"
	if _, err := computeRateSIntervals(ordRts, 0, 3*time.Minute); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, \nreceived %+v", expected, err)
	}
}

func TestComputeRateSIntervalsZeroIncrement(t *testing.T) {
	rt1 := &engine.Rate{
		ID: "RATE1",
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
				RecurrentFee:  0.10,
				Unit:          30 * time.Second,
				Increment:     0 * time.Second,
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
	rt1 := &engine.Rate{
		ID: "RATE1",
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
				RecurrentFee:  0.10,
				Unit:          30 * time.Second,
				Increment:     1 * time.Second,
			},
			{
				IntervalStart: 30 * time.Second,
				RecurrentFee:  0.25,
				Unit:          time.Minute,
				Increment:     7 * time.Second,
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
	rt1 := &engine.Rate{
		ID: "RATE1",
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
				RecurrentFee:  0.10,
				Unit:          30 * time.Second,
				Increment:     1 * time.Second,
			},
			{
				IntervalStart: 25 * time.Second,
				RecurrentFee:  0.25,
				Unit:          time.Minute,
				Increment:     7 * time.Second,
			},
		},
	}
	rt1.Compile()

	rt2 := &engine.Rate{
		ID: "RATE2",
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 45 * time.Second,
				RecurrentFee:  0.15,
				Unit:          30 * time.Second,
				Increment:     1 * time.Second,
			},
			{
				IntervalStart: 55 * time.Second,
				RecurrentFee:  0.30,
				Unit:          30 * time.Minute,
				Increment:     5 * time.Second,
			},
		},
	}
	rt2.Compile()

	rt3 := &engine.Rate{
		ID: "RATE3",
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 30 * time.Second,
				RecurrentFee:  0.10,
				Unit:          25 * time.Second,
				Increment:     5 * time.Second,
			},
			{
				IntervalStart: time.Minute,
				RecurrentFee:  0.05,
				Unit:          10 * time.Second,
				Increment:     5 * time.Second,
			},
		},
	}
	rt3.Compile()

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
	rt1 := &engine.Rate{
		ID: "RATE1",
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: time.Minute,
				RecurrentFee:  0.30,
				Unit:          30 * time.Second,
				Increment:     1 * time.Second,
			},
			{
				IntervalStart: 3 * time.Minute,
				RecurrentFee:  0.30,
				Unit:          time.Minute,
				Increment:     7 * time.Second,
			},
		},
	}
	rt1.Compile()

	rt2 := &engine.Rate{
		ID: "RATE2",
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: time.Minute + 30*time.Second,
				RecurrentFee:  0.20,
				Unit:          30 * time.Second,
				Increment:     1 * time.Second,
			},
			{
				IntervalStart: 2*time.Minute + 30*time.Second,
				RecurrentFee:  0.20,
				Unit:          30 * time.Second,
				Increment:     7 * time.Second,
			},
		},
	}
	rt2.Compile()

	rt3 := &engine.Rate{
		ID: "RATE3",
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 2 * time.Minute,
				RecurrentFee:  0.10,
				Unit:          30 * time.Second,
				Increment:     30 * time.Second,
			},
		},
	}
	rt3.Compile()

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
	rt1 := &engine.Rate{
		ID: "RATE1",
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
				RecurrentFee:  0.50,
				Unit:          time.Hour,
				Increment:     3 * time.Minute,
			},
			{
				IntervalStart: 4 * time.Hour,
				RecurrentFee:  0.35,
				Unit:          time.Hour,
				Increment:     3 * time.Minute,
			},
		},
	}
	rt1.Compile()

	rt2 := &engine.Rate{
		ID: "RATE2",
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 13 * time.Hour,
				RecurrentFee:  0.40,
				Unit:          time.Hour,
				Increment:     3 * time.Minute,
			},
			{
				IntervalStart: 16 * time.Hour,
				RecurrentFee:  0.35,
				Unit:          time.Hour,
				Increment:     3 * time.Minute,
			},
		},
	}
	rt1.Compile()

	rtGH := &engine.Rate{
		ID: "RATE_GOLDEN_HOUR",
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 12 * time.Hour,
				RecurrentFee:  0.20,
				Unit:          30 * time.Hour,
				Increment:     2 * time.Minute,
			},
			{
				IntervalStart: 12*time.Hour + 30*time.Minute,
				RecurrentFee:  0.10,
				Unit:          time.Hour,
				Increment:     5 * time.Minute,
			},
		},
	}
	rtGH.Compile()

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
	rt1 := &engine.Rate{
		ID: "RATE1",
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
				RecurrentFee:  0.50,
				Unit:          10 * time.Second,
				Increment:     2 * time.Second,
			},
			{
				IntervalStart: 4 * time.Second,
				RecurrentFee:  0.48,
				Unit:          10 * time.Second,
				Increment:     2 * time.Second,
			},
			{
				IntervalStart: 8 * time.Second,
				RecurrentFee:  0.45,
				Unit:          10 * time.Second,
				Increment:     2 * time.Second,
			},
		},
	}
	rt1.Compile()

	rt2 := &engine.Rate{
		ID: "RATE2",
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
				RecurrentFee:  0.50,
				Unit:          2 * time.Second,
				Increment:     time.Second,
			},
			{
				IntervalStart: 2 * time.Second,
				RecurrentFee:  0.48,
				Unit:          2 * time.Second,
				Increment:     time.Second,
			},
			{
				IntervalStart: 6 * time.Second,
				RecurrentFee:  0.45,
				Unit:          2 * time.Second,
				Increment:     time.Second,
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
	rt1 := &engine.Rate{
		ID: "RATE1",
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
				RecurrentFee:  0.05,
				Unit:          15 * time.Minute,
				Increment:     15 * time.Minute,
			},
			{
				IntervalStart: time.Hour,
				RecurrentFee:  0.50,
				Unit:          5 * time.Minute,
				Increment:     time.Minute + 30*time.Second,
			},
		},
	}
	rt1.Compile()

	rt2 := &engine.Rate{
		ID: "RATE2",
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 15 * time.Minute,
				RecurrentFee:  0.10,
				Unit:          10 * time.Minute,
				Increment:     time.Minute,
			},
			{
				IntervalStart: 30 * time.Minute,
				RecurrentFee:  0.15,
				Unit:          10 * time.Minute,
				Increment:     time.Minute,
			},
			{
				IntervalStart: 45 * time.Minute,
				RecurrentFee:  0.20,
				Unit:          10 * time.Minute,
				Increment:     time.Minute,
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
	rt1 := &engine.Rate{
		ID: "RATE1",
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
				RecurrentFee:  0.20,
				Unit:          time.Minute,
				Increment:     30 * time.Second,
			},
			{
				IntervalStart: 30 * time.Minute,
				RecurrentFee:  0.20,
				Unit:          time.Minute,
				Increment:     30 * time.Second,
			},
		},
	}
	rt1.Compile()

	rt2 := &engine.Rate{
		ID: "RATE2",
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
				RecurrentFee:  0.20,
				Unit:          time.Minute,
				Increment:     30 * time.Second,
			},
			{
				IntervalStart: 30 * time.Minute,
				RecurrentFee:  0.20,
				Unit:          time.Minute,
				Increment:     30 * time.Second,
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
