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

package utils

import (
	"math"
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cron"
	"github.com/ericlagergren/decimal"
)

func TestLibratesTenantID(t *testing.T) {
	rp := &RateProfile{
		Tenant: "tenant",
		ID:     "testID",
	}
	expected := rp.Tenant + ":" + rp.ID
	received := rp.TenantID()
	if received != expected {
		t.Errorf("\nExpected: <%v>,\nReceived: <%v>", expected, received)
	}
}

func TestLibratesCompilerp(t *testing.T) {
	// empty struct
	rp := &RateProfile{}
	err := rp.Compile()
	if err != nil {
		t.Errorf("\nExpected: <%v>, \nReceived: <%v>", nil, err)
	}

	// non-empty
	fail := "shouldfail"
	rp.ID = "test"
	rp.Tenant = "tenant"
	rp.Rates = map[string]*Rate{
		"testKey1": {
			ID:              "ID1",
			ActivationTimes: fail,
		},
		"testKey2": {
			ID: "ID2",
		},
	}

	expected := "expected exactly 5 fields, found 1: [" + fail + "]"
	err = rp.Compile()

	if err == nil || err.Error() != expected {
		t.Errorf("\nExpected: <%v>, \nReceived: <%v>", expected, err)
	}
}

func TestLibratesUID(t *testing.T) {
	rt := &Rate{
		uID: "testString",
	}

	expected := "testString"
	received := rt.UID()

	if received != expected {
		t.Errorf("\nExpected: %q, \nReceived: %q", expected, received)
	}

}

func TestLibratesCompilert(t *testing.T) {
	rt := &Rate{
		ActivationTimes: EmptyString,
	}

	err := rt.Compile()

	if err != nil {
		t.Errorf("\nExpected: <%+v>, \nReceived: <%+v>", nil, err)
	}
}

func TestLibratesRunTimes(t *testing.T) {
	var (
		sTime     time.Time
		eTime     time.Time
		verbosity int
	)

	// memory leak test
	verbosity = 0

	rt := &Rate{}

	received, err := rt.RunTimes(sTime, eTime, verbosity)
	var expected [][]time.Time

	if err == nil || err != ErrMaxIterationsReached {
		t.Errorf("\nExpected: <%+v>, \nReceived: <%+v>", ErrMaxIterationsReached, err)
	}

	if !reflect.DeepEqual(received, expected) {
		t.Errorf("\nExpected: <%+v>, \nReceived: <%+v>", expected, received)
	}

	// aTime after eTime test
	schd, err := cron.ParseStandard("* * * * *")
	if err != nil {
		t.Errorf("\ndidn't expect error, got %v", err)
	}

	rt.sched = schd
	verbosity = 10
	eTime = sTime.Add(10 * time.Minute)

	received, err = rt.RunTimes(sTime, eTime, verbosity)

	if err != nil {
		t.Errorf("\ndidn't expect error, got %v", err)
	}

	if !reflect.DeepEqual(received, expected) {
		t.Errorf("\nExpected: <%+v>, \nReceived: <%+v>", expected, received)
	}

	// eTime before iTime test
	schd, err = cron.ParseStandard("* * 12 3 *")
	if err != nil {
		t.Errorf("\ndidn't expect error, got %v", err)
	}

	rt.sched = schd
	sTime, err = time.Parse(time.RFC3339, "2022-03-11T15:04:05Z")
	if err != nil {
		t.Errorf("\ndidn't expect error, got %v", err)
	}
	eTime = sTime.Add(24 * time.Hour)

	if received, err = rt.RunTimes(sTime, eTime, verbosity); err != nil {
		t.Error(err)
	}

	aT1, err := time.Parse(time.RFC3339, "2022-03-12T00:00:00Z")
	if err != nil {
		t.Errorf("\ndidn't expect error, got %v", err)
	}

	aT2, err := time.Parse(time.RFC3339, "2022-03-13T00:00:00Z")
	if err != nil {
		t.Errorf("\ndidn't expect error, got %v", err)
	}

	aTsl := make([]time.Time, 0)
	aTsl = append(aTsl, aT1, aT2)
	expected = append(expected, aTsl)

	if !reflect.DeepEqual(received, expected) {
		t.Errorf("\nExpected: <%+v>, \nReceived: <%+v>", expected, received)
	}

	//eTime after iTime
	schd, err = cron.ParseStandard("* * 12 3 *")
	if err != nil {
		t.Errorf("\ndidn't expect error, got %v", err)
	}

	rt.sched = schd
	sTime, err = time.Parse(time.RFC3339, "2022-03-11T15:04:05Z")
	if err != nil {
		t.Errorf("\ndidn't expect error, got %v", err)
	}
	eTime = sTime.Add(48 * time.Hour)

	received, err = rt.RunTimes(sTime, eTime, verbosity)

	if err != nil {
		t.Errorf("\ndidn't expect error, got %v", err)
	}

	if !reflect.DeepEqual(received, expected) {
		t.Errorf("\nExpected: <%+v>, \nReceived: <%+v>", expected, received)
	}
}

/*
func TestLibratesCorrectCost(t *testing.T) {

	//CorrectCost does nothing
	rPc := &RateProfileCost{
		Cost:    NewDecimal(1234, 3),
		MinCost: NewDecimal(1, 0),
		MaxCost: NewDecimal(2, 0),
		Altered: []string{},
	}

	expected := &RateProfileCost{
		Cost:    NewDecimal(1234, 3),
		MinCost: NewDecimal(1, 0),
		MaxCost: NewDecimal(2, 0),
		Altered: []string{},
	}
	rPc.CorrectCost(nil, "")

	if !reflect.DeepEqual(rPc, expected) {
		t.Errorf("\nExpected: <%+v>, \nReceived: <%+v>", expected, rPc)
	}

	//CorrectCost rounds the cost
	expected = &RateProfileCost{
		Cost:    NewDecimal(124, 2),
		MinCost: NewDecimal(1, 0),
		MaxCost: NewDecimal(2, 0),
		Altered: []string{RoundingDecimals},
	}

	rPc.CorrectCost(IntPointer(2), MetaRoundingUp)

	if !reflect.DeepEqual(rPc, expected) {
		t.Errorf("\nExpected: <%+v>, \nReceived: <%+v>", ToJSON(expected), ToJSON(rPc))
	}

	//CorrectCost assigns MaxCost to Cost when Cost > MaxCost

	expected = &RateProfileCost{
		Cost:    NewDecimal(2, 0),
		MinCost: NewDecimal(1, 0),
		MaxCost: NewDecimal(2, 0),
		Altered: []string{RoundingDecimals, MaxCost},
	}
	rPc.Cost = NewDecimal(234, 2)
	rPc.CorrectCost(nil, "")

	if !reflect.DeepEqual(rPc, expected) {
		t.Errorf("\nExpected: <%+v>, \nReceived: <%+v>", expected, rPc)
	}

	//CorrectCost assigns MinCost to Cost when Cost < MinCost

	expected = &RateProfileCost{
		Cost:    NewDecimal(1, 0),
		MinCost: NewDecimal(1, 0),
		MaxCost: NewDecimal(2, 0),
		Altered: []string{RoundingDecimals, MaxCost, MinCost},
	}
	rPc.Cost = NewDecimal(12, 2)
	rPc.CorrectCost(nil, "")

	if !reflect.DeepEqual(rPc, expected) {
		t.Errorf("\nExpected: <%+v>, \nReceived: <%+v>", expected, rPc)
	}
}

*/

func TestRateProfileSort(t *testing.T) {
	minDecimal, err := NewDecimalFromUsage("1m")
	if err != nil {
		t.Error(err)
	}
	secDecimal, err := NewDecimalFromUsage("1s")
	if err != nil {
		t.Error(err)
	}
	rPrf := &RateProfile{
		Tenant: "cgrates.org",
		ID:     "RP1",
		Rates: map[string]*Rate{
			"RT_WEEK": {
				ID: "RT_WEEK",
				Weights: DynamicWeights{
					{
						Weight: 0,
					},
				},
				ActivationTimes: "* * * * 1-5",
				IntervalRates: []*IntervalRate{
					{
						IntervalStart: NewDecimal(0, 0),
						RecurrentFee:  NewDecimal(12, 2),
						Unit:          minDecimal,
						Increment:     minDecimal,
					},
					{
						IntervalStart: NewDecimal(int64(time.Minute), 0),
						RecurrentFee:  NewDecimal(6, 2),
						Unit:          minDecimal,
						Increment:     secDecimal,
					},
				},
			},
			"RT_Custom": {
				ID: "RT_Custom",
				Weights: DynamicWeights{
					{
						Weight: 0,
					},
				},
				ActivationTimes: "* * * * 1-5",
				IntervalRates: []*IntervalRate{
					{
						IntervalStart: NewDecimal(int64(time.Second), 0),
						RecurrentFee:  NewDecimal(12, 2),
						Unit:          minDecimal,
						Increment:     minDecimal,
					},
					{
						IntervalStart: NewDecimal(int64(time.Second), 0),
						RecurrentFee:  NewDecimal(19, 2),
						Unit:          minDecimal,
						Increment:     secDecimal,
					},
					{
						IntervalStart: NewDecimal(int64(15*time.Second), 0),
						RecurrentFee:  NewDecimal(4, 1),
						Unit:          minDecimal,
						Increment:     secDecimal,
					},
					{
						IntervalStart: NewDecimal(int64(10*time.Second), 0),
						RecurrentFee:  NewDecimal(27, 2),
						Unit:          minDecimal,
						Increment:     secDecimal,
					},
				},
			},
			"RT_WEEKEND": {
				ID: "RT_WEEKEND",
				Weights: DynamicWeights{
					{
						Weight: 10,
					},
				},
				ActivationTimes: "* * * * 0,6",
				IntervalRates: []*IntervalRate{
					{
						IntervalStart: NewDecimal(int64(10*time.Second), 0),
						RecurrentFee:  NewDecimal(6, 2),
						Unit:          minDecimal,
						Increment:     secDecimal,
					},
					{
						IntervalStart: NewDecimal(int64(time.Minute), 0),
						RecurrentFee:  NewDecimal(18, 2),
						Unit:          minDecimal,
						Increment:     secDecimal,
					},
					{
						IntervalStart: NewDecimal(int64(18*time.Second), 0),
						RecurrentFee:  NewDecimal(12, 2),
						Unit:          minDecimal,
						Increment:     secDecimal,
					},
				},
			},
			"RT_CHRISTMAS": {
				ID: "RT_CHRISTMAS",
				Weights: DynamicWeights{
					{
						Weight: 30,
					},
				},
				ActivationTimes: "* * 24 12 *",
				IntervalRates: []*IntervalRate{
					{
						IntervalStart: NewDecimal(0, 0),
						RecurrentFee:  NewDecimal(6, 2),
						Unit:          minDecimal,
						Increment:     secDecimal,
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
				ID: "RT_WEEK",
				Weights: DynamicWeights{
					{
						Weight: 0,
					},
				},
				ActivationTimes: "* * * * 1-5",
				IntervalRates: []*IntervalRate{
					{
						IntervalStart: NewDecimal(0, 0),
						RecurrentFee:  NewDecimal(12, 2),
						Unit:          minDecimal,
						Increment:     minDecimal,
					},
					{
						IntervalStart: NewDecimal(int64(time.Minute), 0),
						RecurrentFee:  NewDecimal(6, 2),
						Unit:          minDecimal,
						Increment:     secDecimal,
					},
				},
			},
			"RT_WEEKEND": {
				ID: "RT_WEEKEND",
				Weights: DynamicWeights{
					{
						Weight: 10,
					},
				},
				ActivationTimes: "* * * * 0,6",
				IntervalRates: []*IntervalRate{
					{
						IntervalStart: NewDecimal(int64(10*time.Second), 0),
						RecurrentFee:  NewDecimal(6, 2),
						Unit:          minDecimal,
						Increment:     secDecimal,
					},
					{
						IntervalStart: NewDecimal(int64(18*time.Second), 0),
						RecurrentFee:  NewDecimal(12, 2),
						Unit:          minDecimal,
						Increment:     secDecimal,
					},
					{
						IntervalStart: NewDecimal(int64(time.Minute), 0),
						RecurrentFee:  NewDecimal(18, 2),
						Unit:          minDecimal,
						Increment:     secDecimal,
					},
				},
			},
			"RT_Custom": {
				ID: "RT_Custom",
				Weights: DynamicWeights{
					{
						Weight: 0,
					},
				},
				ActivationTimes: "* * * * 1-5",
				IntervalRates: []*IntervalRate{
					{
						IntervalStart: NewDecimal(int64(time.Second), 0),
						RecurrentFee:  NewDecimal(12, 2),
						Unit:          minDecimal,
						Increment:     minDecimal,
					},
					{
						IntervalStart: NewDecimal(int64(time.Second), 0),
						RecurrentFee:  NewDecimal(19, 2),
						Unit:          minDecimal,
						Increment:     secDecimal,
					},
					{
						IntervalStart: NewDecimal(int64(10*time.Second), 0),
						RecurrentFee:  NewDecimal(27, 2),
						Unit:          minDecimal,
						Increment:     secDecimal,
					},
					{
						IntervalStart: NewDecimal(int64(15*time.Second), 0),
						RecurrentFee:  NewDecimal(4, 1),
						Unit:          minDecimal,
						Increment:     secDecimal,
					},
				},
			},
			"RT_CHRISTMAS": {
				ID: "RT_CHRISTMAS",
				Weights: DynamicWeights{
					{
						Weight: 30,
					},
				},
				ActivationTimes: "* * 24 12 *",
				IntervalRates: []*IntervalRate{
					{
						IntervalStart: NewDecimal(0, 0),
						RecurrentFee:  NewDecimal(6, 2),
						Unit:          minDecimal,
						Increment:     secDecimal,
					},
				},
			},
		},
	}
	rPrf.Sort()
	if !reflect.DeepEqual(rPrf, exp) {
		t.Errorf("Expected: %v,\n received: %v", ToJSON(exp), ToJSON(rPrf))
	}
}

func TestRateProfileCompile(t *testing.T) {
	rt := &RateProfile{
		Rates: map[string]*Rate{
			"randomVal1": {
				ID: "RT_CHRISTMAS",
				Weights: DynamicWeights{
					{
						Weight: 30,
					},
				},
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
				ID: "RT_CHRISTMAS",
				Weights: DynamicWeights{
					{
						Weight: 30,
					},
				},
				ActivationTimes: "* * 24 12 *",
				sched:           expectedATime,
				uID:             ConcatenatedKey(rt.Tenant, rt.ID, "RT_CHRISTMAS"),
			},
		},
		Tenant: "cgrates.org",
		ID:     "RTP1",
	}
	if err := rt.Compile(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rt, expRt) {
		t.Errorf("Expected %+v, received %+v", ToJSON(expRt), ToJSON(rt))
	}
}

func TestRateUID(t *testing.T) {
	rt := &RateProfile{
		Rates: map[string]*Rate{
			"randomVal1": {
				ID: "RT_CHRISTMAS",
				Weights: DynamicWeights{
					{
						Weight: 30,
					},
				},
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
				ID: "RT_CHRISTMAS",
				Weights: DynamicWeights{
					{
						Weight: 30,
					},
				},
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
		ID: "RT_CHRISTMAS",
		Weights: DynamicWeights{
			{
				Weight: 30,
			},
		},
		ActivationTimes: "* * 24 12 *",
	}
	expTime, err := cron.ParseStandard("* * 24 12 *")
	if err != nil {
		t.Error(err)
	}
	expectedRt := &Rate{
		ID: "RT_CHRISTMAS",
		Weights: DynamicWeights{
			{
				Weight: 30,
			},
		},
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
		ID: "RT_CHRISTMAS",
		Weights: DynamicWeights{
			{
				Weight: 30,
			},
		},
		ActivationTimes: EmptyString,
	}
	expTime, err := cron.ParseStandard("* * * * *")
	if err != nil {
		t.Error(err)
	}
	expectedRt := &Rate{
		ID: "RT_CHRISTMAS",
		Weights: DynamicWeights{
			{
				Weight: 30,
			},
		},
		ActivationTimes: EmptyString,
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
				IntervalStart: NewDecimal(0, 0),
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
		ID: "RT_CHRISTMAS",
		Weights: DynamicWeights{
			{
				Weight: 30,
			},
		},
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
				IntervalStart: NewDecimal(0, 0),
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
				IntervalStart: NewDecimal(0, 0),
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
	minDecimal, err := NewDecimalFromUsage("1m")
	if err != nil {
		t.Error(err)
	}
	secDecimal, err := NewDecimalFromUsage("1s")
	if err != nil {
		t.Error(err)
	}
	rt0 := &Rate{
		ID: "RATE0",
		IntervalRates: []*IntervalRate{
			{
				IntervalStart: NewDecimal(0, 0),
				Unit:          minDecimal,
				Increment:     minDecimal,
				RecurrentFee:  NewDecimal(24, 1),
			},
			{
				IntervalStart: NewDecimal(int64(60*time.Second), 0),
				Unit:          minDecimal,
				Increment:     secDecimal,
				RecurrentFee:  NewDecimal(24, 1),
			},
		},
	}
	rt0.Compile()
	rt1 := &Rate{
		ID: "RATE1",
		IntervalRates: []*IntervalRate{
			{

				IntervalStart: NewDecimal(0, 0),
				Unit:          minDecimal,
				Increment:     secDecimal,
				RecurrentFee:  NewDecimal(12, 1),
			},
			{

				IntervalStart: NewDecimal(int64(2*time.Minute), 0),
				Unit:          minDecimal,
				Increment:     secDecimal,
				RecurrentFee:  NewDecimal(6, 1),
			},
		},
	}
	rt1.Compile()
	rtsMap := map[string]*IntervalRate{
		"RATE0_0": rt0.IntervalRates[0],
		"RATE0_1": rt0.IntervalRates[1],
		"RATE1_0": rt1.IntervalRates[0],
		"RATE1_1": rt1.IntervalRates[1],
	}

	rtIvls := []*RateSIntervalCost{
		{
			IntervalStart: NewDecimal(0, 0),
			Increments: []*RateSIncrementCost{
				{
					IncrementStart:    NewDecimal(0, 0),
					Usage:             NewDecimal(int64(time.Minute), 0),
					IntervalRateIndex: 0,
					RateID:            "RATE0_0",
					CompressFactor:    1,
				},
				{
					IncrementStart:    NewDecimal(int64(time.Minute), 0),
					Usage:             NewDecimal(int64(30*time.Second), 0),
					IntervalRateIndex: 1,
					RateID:            "RATE0_1",
					CompressFactor:    30,
				},
			},
			CompressFactor: 1,
		},
		{
			IntervalStart: NewDecimal(int64(90*time.Second), 0),
			Increments: []*RateSIncrementCost{
				{
					IncrementStart:    NewDecimal(int64(90*time.Second), 0),
					Usage:             NewDecimal(int64(30*time.Second), 0),
					IntervalRateIndex: 0,
					RateID:            "RATE1_0",
					CompressFactor:    30,
				},
				{
					IncrementStart:    NewDecimal(int64(2*time.Minute), 0),
					Usage:             NewDecimal(int64(10*time.Minute), 0),
					IntervalRateIndex: 1,
					RateID:            "RATE1_1",
					CompressFactor:    10,
				},
			},
			CompressFactor: 1,
		},
	}
	eDcml, _ := new(decimal.Big).SetFloat64(4.3).Float64()
	cost := CostForIntervals(rtIvls, rtsMap)
	if costFlt, _ := cost.Float64(); costFlt != eDcml {
		t.Errorf("eDcml: %f, received: %+v", eDcml, cost)
	}
}

func TestCostForIntervalsWIthFixedFee(t *testing.T) {
	minDecimal, err := NewDecimalFromUsage("1m")
	if err != nil {
		t.Error(err)
	}
	secDecimal, err := NewDecimalFromUsage("1s")
	if err != nil {
		t.Error(err)
	}
	rt0 := &Rate{
		ID: "RATE0",
		IntervalRates: []*IntervalRate{
			{
				IntervalStart: NewDecimal(0, 0),
				FixedFee:      NewDecimal(4, 1),
				RecurrentFee:  NewDecimal(24, 1),
				Unit:          minDecimal,
				Increment:     minDecimal,
			},
			{
				IntervalStart: NewDecimal(int64(60*time.Second), 0),
				RecurrentFee:  NewDecimal(24, 1),
				Unit:          minDecimal,
				Increment:     secDecimal,
			},
		},
	}
	rt0.Compile()
	rt1 := &Rate{
		ID: "RATE1",
		IntervalRates: []*IntervalRate{
			//{},
			{
				IntervalStart: NewDecimal(0, 0),
				FixedFee:      NewDecimal(2, 1),
				RecurrentFee:  NewDecimal(12, 1),
				Unit:          minDecimal,
				Increment:     secDecimal,
			},
			{
				IntervalStart: NewDecimal(int64(2*time.Minute), 0),
				RecurrentFee:  NewDecimal(6, 1),
				Unit:          minDecimal,
				Increment:     secDecimal,
			},
		},
	}
	rt1.Compile()
	rtsMap := map[string]*IntervalRate{
		"RATE0_0": rt0.IntervalRates[0],
		"RATE0_1": rt0.IntervalRates[1],
		"RATE1_0": rt1.IntervalRates[0],
		"RATE1_1": rt1.IntervalRates[1],
	}

	rtIvls := []*RateSIntervalCost{
		{
			IntervalStart: NewDecimal(0, 0),
			Increments: []*RateSIncrementCost{
				{ // cost 0,4
					IncrementStart:    NewDecimal(0, 0),
					IntervalRateIndex: 0,
					RateID:            "RATE0_0",
					CompressFactor:    1,
					Usage:             NewDecimal(-1, 0),
				},
				{ // cost 2,4
					IncrementStart:    NewDecimal(0, 0),
					IntervalRateIndex: 0,
					RateID:            "RATE0_0",
					CompressFactor:    1,
					Usage:             NewDecimal(int64(time.Minute), 0),
				},
				{ // cost 1,2
					IncrementStart:    NewDecimal(int64(time.Minute), 0),
					IntervalRateIndex: 1,
					RateID:            "RATE0_1",
					CompressFactor:    30,
					Usage:             NewDecimal(int64(30*time.Second), 0),
				},
			},
			CompressFactor: 1,
		},
		{
			IntervalStart: NewDecimal(int64(90*time.Second), 0),
			Increments: []*RateSIncrementCost{
				{ // cost 0,2
					IncrementStart:    NewDecimal(int64(90*time.Second), 0),
					IntervalRateIndex: 0,
					RateID:            "RATE1_0",
					CompressFactor:    1,
					Usage:             NewDecimal(-1, 0),
				},
				{ // cost 0,6
					IncrementStart:    NewDecimal(int64(90*time.Second), 0),
					IntervalRateIndex: 0,
					RateID:            "RATE1_0",
					CompressFactor:    30,
					Usage:             NewDecimal(int64(30*time.Second), 0),
				},
				{ // cost 0,1
					IncrementStart:    NewDecimal(int64(2*time.Minute), 0),
					IntervalRateIndex: 1,
					RateID:            "RATE1_1",
					CompressFactor:    10,
					Usage:             NewDecimal(int64(10*time.Second), 0),
				},
			},
			CompressFactor: 1,
		},
	}
	eDcml, _ := new(decimal.Big).SetFloat64(4.9).Float64()
	cost := CostForIntervals(rtIvls, rtsMap)
	if costFlt, _ := cost.Float64(); costFlt != eDcml {
		t.Errorf("eDcml: %f, received: %+v", eDcml, costFlt)
	}
}

/*
func TestRateProfileCostCorrectCost(t *testing.T) {
	rPrfCost := &RateProfileCost{
		ID:   "Test1",
		Cost: NewDecimal(234, 3),
	}
	rPrfCost.CorrectCost(IntPointer(2), MetaRoundingUp)
	if rPrfCost.Cost != NewDecimal(24, 2) {
		t.Errorf("Expected: %+v, received: %+v", 0.24, rPrfCost.Cost)
	}
	if !reflect.DeepEqual(rPrfCost.Altered, []string{RoundingDecimals}) {
		t.Errorf("Expected: %+v, received: %+v", []string{RoundingDecimals}, rPrfCost.Altered)
	}

}

func TestRateProfileCostCorrectCostMinCost(t *testing.T) {
	testRPC := &RateProfileCost{
		Cost:    NewDecimal(5, 1),
		MinCost: NewDecimal(15, 1),
	}
	testRPC.CorrectCost(IntPointer(2), "")
	if testRPC.Cost != NewDecimal(15, 1) {
		t.Errorf("\nExpecting: <1.5>,\n Received: <%+v>", testRPC.Cost)
	}
}

func TestRateProfileCostCorrectCostMaxCost(t *testing.T) {
	testRPC := &RateProfileCost{
		Cost:    NewDecimal(25, 1),
		MaxCost: NewDecimal(15, 1),
	}
	testRPC.CorrectCost(IntPointer(2), "")
	if testRPC.Cost != NewDecimal(15, 1) {
		t.Errorf("\nExpecting: <1.5>,\n Received: <%+v>", testRPC.Cost)
	}
}

*/

func TestRateSIncrementCompressEquals(t *testing.T) {
	inCr1 := &RateSIncrement{
		IncrementStart:    NewDecimal(0, 0),
		Usage:             NewDecimal(int64(time.Minute), 0),
		IntervalRateIndex: 0,
		CompressFactor:    1,
	}
	inCr2 := &RateSIncrement{
		IncrementStart:    NewDecimal(0, 0),
		Usage:             NewDecimal(int64(time.Minute), 0),
		IntervalRateIndex: 0,
		CompressFactor:    1,
	}
	result := inCr1.CompressEquals(inCr2)
	if result != true {
		t.Errorf("\nExpecting: <true>,\n Received: <%+v>", result)
	}
}

func TestRateSIncrementCompressEqualsCase1(t *testing.T) {
	inCr1 := &RateSIncrement{
		IncrementStart:    NewDecimal(0, 0),
		Usage:             NewDecimal(int64(time.Minute), 0),
		IntervalRateIndex: 1,
		CompressFactor:    1,
	}
	inCr2 := &RateSIncrement{
		IncrementStart:    NewDecimal(0, 0),
		Usage:             NewDecimal(int64(time.Minute), 0),
		IntervalRateIndex: 0,
		CompressFactor:    1,
	}
	result := inCr1.CompressEquals(inCr2)
	if result != false {
		t.Errorf("\nExpecting: <false>,\n Received: <%+v>", result)
	}
}

func TestRateSIncrementCompressEqualsCase2(t *testing.T) {
	inCr1 := &RateSIncrement{
		IncrementStart:    NewDecimal(0, 0),
		Usage:             NewDecimal(int64(time.Minute), 0),
		IntervalRateIndex: 0,
		CompressFactor:    1,
	}
	inCr2 := &RateSIncrement{
		IncrementStart:    NewDecimal(0, 0),
		Usage:             NewDecimal(int64(time.Minute), 0),
		IntervalRateIndex: 1,
		CompressFactor:    1,
	}
	result := inCr1.CompressEquals(inCr2)
	if result != false {
		t.Errorf("\nExpecting: <false>,\n Received: <%+v>", result)
	}
}

func TestRateSIncrementCompressEqualsCase3(t *testing.T) {
	inCr1 := &RateSIncrement{
		IncrementStart:    NewDecimal(0, 0),
		Usage:             NewDecimal(int64(time.Second), 0),
		IntervalRateIndex: 1,
		CompressFactor:    1,
	}
	inCr2 := &RateSIncrement{
		IncrementStart:    NewDecimal(0, 0),
		Usage:             NewDecimal(int64(time.Minute), 0),
		IntervalRateIndex: 1,
		CompressFactor:    1,
	}
	result := inCr1.CompressEquals(inCr2)
	if result != false {
		t.Errorf("\nExpecting: <false>,\n Received: <%+v>", result)
	}
}

func TestRateSIntervalCompressEqualsCase1(t *testing.T) {
	rateSintrv1 := &RateSInterval{
		IntervalStart: NewDecimal(0, 0),
		Increments: []*RateSIncrement{
			{
				IncrementStart:    NewDecimal(0, 0),
				Usage:             NewDecimal(int64(time.Minute), 0),
				IntervalRateIndex: 0,
				CompressFactor:    1,
			},
			{
				IncrementStart:    NewDecimal(int64(time.Minute), 0),
				Usage:             NewDecimal(int64(time.Minute+10*time.Second), 0),
				IntervalRateIndex: 1,
				CompressFactor:    70,
			},
		},
		CompressFactor: 1,
	}

	rateSintrv2 := &RateSInterval{
		IntervalStart: NewDecimal(0, 0),
		Increments: []*RateSIncrement{
			{
				IncrementStart:    NewDecimal(0, 0),
				Usage:             NewDecimal(int64(time.Minute), 0),
				IntervalRateIndex: 0,
				CompressFactor:    1,
			},
		},
		CompressFactor: 1,
	}
	result := rateSintrv1.CompressEquals(rateSintrv2)
	if result != false {
		t.Errorf("\nExpecting <false>,\nReceived <%+v>", result)
	}
}

func TestRateSIntervalCompressEqualsCase2(t *testing.T) {
	rateSintrv1 := &RateSInterval{
		IntervalStart: NewDecimal(0, 0),
		Increments: []*RateSIncrement{
			{
				IncrementStart:    NewDecimal(0, 0),
				Usage:             NewDecimal(int64(time.Minute), 0),
				IntervalRateIndex: 0,
				CompressFactor:    1,
			},
			{
				IncrementStart:    NewDecimal(0, 0),
				Usage:             NewDecimal(int64(time.Second), 0),
				IntervalRateIndex: 0,
				CompressFactor:    1,
			},
		},
		CompressFactor: 1,
	}

	rateSintrv2 := &RateSInterval{
		IntervalStart: NewDecimal(1, 0),
		Increments: []*RateSIncrement{
			{
				IncrementStart:    NewDecimal(0, 0),
				Usage:             NewDecimal(int64(time.Minute), 0),
				IntervalRateIndex: 0,
				CompressFactor:    1,
			},
			{
				IncrementStart:    NewDecimal(0, 0),
				Usage:             NewDecimal(int64(time.Minute), 0),
				IntervalRateIndex: 2,
				CompressFactor:    1,
			},
		},
		CompressFactor: 0,
	}
	result := rateSintrv1.CompressEquals(rateSintrv2)
	if result != false {
		t.Errorf("\nExpecting <false>,\nReceived <%+v>", result)
	}
}

func TestRateSIntervalCompressEqualsCase3(t *testing.T) {
	rateSintrv1 := &RateSInterval{
		IntervalStart: NewDecimal(0, 0),
		Increments: []*RateSIncrement{
			{
				IncrementStart:    NewDecimal(0, 0),
				Usage:             NewDecimal(int64(time.Minute), 0),
				IntervalRateIndex: 0,
				CompressFactor:    1,
			},
		},
		CompressFactor: 1,
	}

	rateSintrv2 := &RateSInterval{
		IntervalStart: NewDecimal(0, 0),
		Increments: []*RateSIncrement{
			{
				IncrementStart:    NewDecimal(0, 0),
				Usage:             NewDecimal(int64(time.Minute), 0),
				IntervalRateIndex: 0,
				CompressFactor:    1,
			},
		},
		CompressFactor: 1,
	}
	result := rateSintrv1.CompressEquals(rateSintrv2)
	if result != true {
		t.Errorf("\nExpecting <true>,\nReceived <%+v>", result)
	}
}

func TestLibratesAsRateProfile(t *testing.T) {
	// Invalid DynamicWeights string
	ext := &APIRateProfile{
		Weights: "testWeight",
	}
	rp := &RateProfile{
		Tenant:          ext.Tenant,
		ID:              ext.ID,
		FilterIDs:       ext.FilterIDs,
		MaxCostStrategy: ext.MaxCostStrategy,
	}

	received, err := ext.AsRateProfile()
	experr := "invalid DynamicWeight format for string <testWeight>"

	if err == nil || err.Error() != experr {
		t.Errorf("\nExpected: <%+v>, \nReceived: <%+v>", experr, err)
	}

	if received != nil {
		t.Errorf("\nExpected: <%+v>, \nReceived: <%+v>", nil, received)
	}

	// No changes
	ext.Weights = EmptyString

	expected := rp
	received, err = ext.AsRateProfile()

	if err != nil {
		t.Errorf("\nExpected nil, got <%+v>", err)
	}

	if !reflect.DeepEqual(received, expected) {
		t.Errorf("\nExpected: <%+v>, \nReceived: <%+v>", expected, received)
	}

	// assign MinCost to rp
	ext.MinCost = Float64Pointer(1)

	expected.MinCost = NewDecimal(1, 0)
	received, err = ext.AsRateProfile()

	if err != nil {
		t.Errorf("\nExpected nil, got <%+v>", err)
	}

	if !reflect.DeepEqual(received, expected) {
		t.Errorf("\nExpected: <%+v>, \nReceived: <%+v>", expected, received)
	}

	// assign MaxCost to rp
	ext.MaxCost = Float64Pointer(2)

	expected.MaxCost = NewDecimal(2, 0)
	received, err = ext.AsRateProfile()

	if err != nil {
		t.Errorf("\nExpected nil, got <%+v>", err)
	}

	if !reflect.DeepEqual(received, expected) {
		t.Errorf("\nExpected: <%+v>, \nReceived: <%+v>", expected, received)
	}
}

func TestLibratesAsRateProfileNon0Len(t *testing.T) {
	id := "testID"
	ext := &APIRateProfile{
		Rates: map[string]*APIRate{
			"testKey": {
				ID:      id,
				Weights: "testWeight",
			},
		},
	}
	rp := &RateProfile{
		Tenant:          ext.Tenant,
		ID:              ext.ID,
		FilterIDs:       ext.FilterIDs,
		MaxCostStrategy: ext.MaxCostStrategy,
	}

	expected := rp
	expected.Rates = map[string]*Rate{
		"testKey": nil,
	}
	received, err := ext.AsRateProfile()

	if err.Error() != "invalid DynamicWeight format for string <testWeight>" {
		t.Errorf("\nExpected nil, got <%+v>", err)
	}

	if !reflect.DeepEqual(received, expected) {
		t.Errorf("\nExpected: <%v>, \nReceived: <%v>", expected, received)
	}
}

func TestRatesIntervalEquals(t *testing.T) {
	rtInt1 := &RateSInterval{
		IntervalStart: NewDecimal(int64(10*time.Second), 0),
		Increments: []*RateSIncrement{
			{
				IncrementStart:    NewDecimal(int64(time.Second), 0),
				IntervalRateIndex: 1,
				RateID:            "ID1",
				CompressFactor:    2,
				Usage:             NewDecimal(int64(5*time.Second), 0),
			},
		},
		CompressFactor: 2,
	}
	rtInt2 := &RateSInterval{
		IntervalStart: NewDecimal(int64(10*time.Second), 0),
		Increments: []*RateSIncrement{
			{
				IncrementStart:    NewDecimal(int64(time.Second), 0),
				IntervalRateIndex: 1,
				RateID:            "ID1",
				CompressFactor:    2,
				Usage:             NewDecimal(int64(5*time.Second), 0),
			},
		},
		CompressFactor: 2,
	}
	riRef := map[string]*IntervalRate{
		"ID1": {
			IntervalStart: NewDecimal(0, 0),
			RecurrentFee:  NewDecimal(12, 1),
		},
	}
	nRiRef := map[string]*IntervalRate{
		"ID1": {
			IntervalStart: NewDecimal(0, 0),
			RecurrentFee:  NewDecimal(12, 1),
		},
	}

	// equals is looking for compressFactor
	if !rtInt1.Equals(rtInt2, riRef, nRiRef) {
		t.Errorf("Intervals %+v and %+v are not equal", ToJSON(rtInt1), ToJSON(rtInt2))
	}

	// not equals for IntervalStart
	rtInt1.IntervalStart = NewDecimal(int64(20*time.Second), 0)
	if rtInt1.Equals(rtInt2, riRef, nRiRef) {
		t.Errorf("Intervals %+v and %+v are equal", ToJSON(rtInt1), ToJSON(rtInt2))
	}
	rtInt1.IntervalStart = NewDecimal(int64(10*time.Second), 0)

	rtInt2.IntervalStart = NewDecimal(int64(20*time.Second), 0)
	if rtInt1.Equals(rtInt2, riRef, nRiRef) {
		t.Errorf("Intervals %+v and %+v are equal", ToJSON(rtInt1), ToJSON(rtInt2))
	}
	rtInt2.IntervalStart = NewDecimal(int64(10*time.Second), 0)

	// not equals for CompressFactor
	rtInt1.CompressFactor = 5
	if rtInt1.Equals(rtInt2, riRef, nRiRef) {
		t.Errorf("Intervals %+v and %+v are equal", ToJSON(rtInt1), ToJSON(rtInt2))
	}
	rtInt1.CompressFactor = 2

	rtInt2.CompressFactor = 8
	if rtInt1.Equals(rtInt2, riRef, nRiRef) {
		t.Errorf("Intervals %+v and %+v are equal", ToJSON(rtInt1), ToJSON(rtInt2))
	}
	rtInt2.CompressFactor = 2

	//not equals for Increments and their length
	rtInt1.Increments[0].Usage = NewDecimal(int64(90*time.Second), 0)
	if rtInt1.Equals(rtInt2, riRef, nRiRef) {
		t.Errorf("Intervals %+v and %+v are equal", ToJSON(rtInt1), ToJSON(rtInt2))
	}
	rtInt1.Increments[0].Usage = NewDecimal(int64(5*time.Second), 0)

	rtInt2.Increments[0].Usage = NewDecimal(int64(80*time.Second), 0)
	if rtInt1.Equals(rtInt2, riRef, nRiRef) {
		t.Errorf("Intervals %+v and %+v are equal", ToJSON(rtInt1), ToJSON(rtInt2))
	}
	rtInt2.Increments[0].Usage = NewDecimal(int64(5*time.Second), 0)

	rtInt1 = &RateSInterval{
		IntervalStart: NewDecimal(int64(10*time.Second), 0),
		Increments: []*RateSIncrement{
			{
				IncrementStart:    NewDecimal(int64(time.Second), 0),
				IntervalRateIndex: 1,
				CompressFactor:    2,
				Usage:             NewDecimal(int64(5*time.Second), 0),
			},
			{
				IncrementStart: NewDecimal(int64(time.Second), 0),
				Usage:          NewDecimal(int64(5*time.Second), 0),
			},
		},
		CompressFactor: 2,
	}
	if rtInt1.Equals(rtInt2, riRef, nRiRef) {
		t.Errorf("Intervals %+v and %+v are equal", ToJSON(rtInt1), ToJSON(rtInt2))
	}
}

func TestRatesIncrementEquals(t *testing.T) {
	incr1 := &RateSIncrement{
		IncrementStart:    NewDecimal(int64(time.Second), 0),
		IntervalRateIndex: 1,
		RateID:            "ID1",
		CompressFactor:    2,
		Usage:             NewDecimal(int64(5*time.Second), 0),
	}
	incr2 := &RateSIncrement{
		IncrementStart:    NewDecimal(int64(time.Second), 0),
		IntervalRateIndex: 1,
		RateID:            "ID1",
		CompressFactor:    2,
		Usage:             NewDecimal(int64(5*time.Second), 0),
	}
	riRef := map[string]*IntervalRate{
		"ID1": {
			IntervalStart: NewDecimal(0, 0),
			RecurrentFee:  NewDecimal(12, 1),
		},
	}
	nRiRef := map[string]*IntervalRate{
		"ID1": {
			IntervalStart: NewDecimal(0, 0),
			RecurrentFee:  NewDecimal(12, 1),
		},
	}

	// equals is not looking for compress factor
	if !incr1.Equals(incr2, riRef, nRiRef) {
		t.Errorf("Intervals %+v and %+v are not equal", ToJSON(incr1), ToJSON(incr2))
	}

	// not equals by IncrementStart
	incr1.IncrementStart = NewDecimal(int64(10*time.Second), 0)
	if incr1.Equals(incr2, riRef, nRiRef) {
		t.Errorf("Intervals %+v and %+v are equal", ToJSON(incr1), ToJSON(incr2))
	}
	incr1.IncrementStart = NewDecimal(int64(time.Second), 0)

	incr2.IncrementStart = NewDecimal(int64(10*time.Second), 0)
	if incr1.Equals(incr2, riRef, nRiRef) {
		t.Errorf("Intervals %+v and %+v are equal", ToJSON(incr1), ToJSON(incr2))
	}
	incr2.IncrementStart = NewDecimal(int64(time.Second), 0)

	// not equals by IntervalRateIndex
	incr1.IntervalRateIndex = 5
	if incr1.Equals(incr2, riRef, nRiRef) {
		t.Errorf("Intervals %+v and %+v are equal", ToJSON(incr1), ToJSON(incr2))
	}
	incr1.IntervalRateIndex = 1

	incr2.IntervalRateIndex = 5
	if incr1.Equals(incr2, riRef, nRiRef) {
		t.Errorf("Intervals %+v and %+v are equal", ToJSON(incr1), ToJSON(incr2))
	}
	incr2.IntervalRateIndex = 1

	// not equals by CompressFactor
	incr1.CompressFactor = 0
	if incr1.Equals(incr2, riRef, nRiRef) {
		t.Errorf("Intervals %+v and %+v are equal", ToJSON(incr1), ToJSON(incr2))
	}
	incr1.CompressFactor = 2

	incr2.CompressFactor = 9
	if incr1.Equals(incr2, riRef, nRiRef) {
		t.Errorf("Intervals %+v and %+v are equal", ToJSON(incr1), ToJSON(incr2))
	}
	incr2.CompressFactor = 2
}

func TestAsExtRateSInterval(t *testing.T) {
	rI := &RateSInterval{
		IntervalStart: NewDecimal(int64(time.Second), 0),
		Increments: []*RateSIncrement{
			{
				IncrementStart:    NewDecimal(int64(time.Nanosecond), 0),
				Usage:             NewDecimal(int64(time.Minute), 0),
				IntervalRateIndex: 0,
				CompressFactor:    1,
				cost:              NewDecimal(1000, 0).Big,
			},
			{
				IncrementStart:    NewDecimal(int64(time.Minute), 0),
				Usage:             NewDecimal(int64(2*time.Minute), 0),
				IntervalRateIndex: 2,
				CompressFactor:    5,
			},
		},
		CompressFactor: 1,
		cost:           NewDecimal(1234, 1).Big,
	}

	expRi := &ExtRateSInterval{
		IntervalStart: Float64Pointer(float64(time.Second)),
		Increments: []*ExtRateSIncrement{
			{
				IncrementStart:    Float64Pointer(float64(time.Nanosecond)),
				Usage:             Float64Pointer(float64(time.Minute)),
				IntervalRateIndex: 0,
				CompressFactor:    1,
				cost:              Float64Pointer(1000),
			},
			{
				IncrementStart:    Float64Pointer(float64(time.Minute)),
				Usage:             Float64Pointer(float64(2 * time.Minute)),
				IntervalRateIndex: 2,
				CompressFactor:    5,
			},
		},
		CompressFactor: 1,
		cost:           Float64Pointer(123.4),
	}

	if rcv, err := rI.AsExtRateSInterval(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, expRi) {
		t.Errorf("Expected %+v \n, received %+v", ToJSON(expRi), ToJSON(rcv))
	}
}

func TestAsExtRateSIntervalErrorsCheck(t *testing.T) {
	rI := &RateSInterval{
		IntervalStart: NewDecimal(int64(math.Inf(1))-1, 0),
		Increments: []*RateSIncrement{
			{
				IncrementStart:    NewDecimal(int64(time.Nanosecond), 0),
				Usage:             NewDecimal(int64(time.Minute), 0),
				IntervalRateIndex: 0,
				CompressFactor:    1,
				cost:              NewDecimal(1000, 0).Big,
			},
		},
		CompressFactor: 1,
		cost:           NewDecimal(1234, 1).Big,
	}

	expErr := "Cannot convert decimal IntervalStart into float64 "
	if _, err := rI.AsExtRateSInterval(); err == nil || err.Error() != expErr {
		t.Errorf("Expected %+v, received %+v", expErr, err)
	}
	rI.IntervalStart = NewDecimal(0, 0)

	rI.cost = NewDecimal(int64(math.Inf(1))-1, 0).Big
	expErr = "Cannot convert decimal cost into float64 "
	if _, err := rI.AsExtRateSInterval(); err == nil || err.Error() != expErr {
		t.Errorf("Expected %+q, received %+q", expErr, err)
	}
	rI.cost = NewDecimal(0, 0).Big

	rI.Increments[0].IncrementStart = NewDecimal(int64(math.Inf(1))-1, 0)
	expErr = "Cannot convert decimal IncrementStart into float64 "
	if _, err := rI.AsExtRateSInterval(); err == nil || err.Error() != expErr {
		t.Errorf("Expected %+q, received %+q", expErr, err)
	}
	rI.Increments[0].IncrementStart = NewDecimal(0, 0)

	rI.Increments[0].Usage = NewDecimal(int64(math.Inf(1))-1, 0)
	expErr = "Cannot convert decimal Usage into float64 "
	if _, err := rI.AsExtRateSInterval(); err == nil || err.Error() != expErr {
		t.Errorf("Expected %+q, received %+q", expErr, err)
	}
	rI.Increments[0].Usage = NewDecimal(0, 0)

	rI.Increments[0].cost = NewDecimal(int64(math.Inf(1))-1, 0).Big
	expErr = "Cannot convert decimal cost into float64 "
	if _, err := rI.AsExtRateSInterval(); err == nil || err.Error() != expErr {
		t.Errorf("Expected %+q, received %+q", expErr, err)
	}
	rI.Increments[0].cost = NewDecimal(0, 0).Big
}

func TestCostForIntervalsWithPartialIntervals(t *testing.T) {
	rt0 := &Rate{
		ID: "RT_2",
		IntervalRates: []*IntervalRate{
			{
				IntervalStart: NewDecimal(0, 0),
				RecurrentFee:  NewDecimal(2, 2),
				Unit:          NewDecimal(int64(time.Second), 0),
				Increment:     NewDecimal(int64(time.Second), 0),
			},
			{
				IntervalStart: NewDecimal(int64(time.Minute), 0),
				FixedFee:      NewDecimal(2, 2),
				Unit:          NewDecimal(int64(time.Second), 0),
				Increment:     NewDecimal(int64(time.Second), 0),
			},
		},
	}

	rtsMap := map[string]*IntervalRate{
		"RT_2_0": rt0.IntervalRates[0],
		"RT_2_1": rt0.IntervalRates[1],
	}

	rtIvls := []*RateSIntervalCost{
		{
			IntervalStart:  NewDecimal(int64(2*time.Minute), 0),
			CompressFactor: 1,
			Increments: []*RateSIncrementCost{
				{
					IncrementStart:    NewDecimal(int64(2*time.Minute), 0),
					IntervalRateIndex: 1,
					RateID:            "RT_2_1",
					CompressFactor:    1,
					Usage:             NewDecimal(-1, 0),
				},
			},
		},
	}

	if cost := CostForIntervals(rtIvls, rtsMap); cost.Cmp(decimal.New(2, 2)) != 0 {
		t.Errorf("received cost: %s", cost)
	}
}
