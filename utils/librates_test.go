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

func TestLibratesCorrectCost(t *testing.T) {
	//CorrectCost does nothing in this case
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

	if !rPc.Equals(expected) {
		t.Errorf("\nExpected: <%+v>, \nReceived: <%+v>", expected, rPc)
	}

	//CorrectCost rounds the cost
	expected = &RateProfileCost{
		Cost:    NewDecimal(12, 1),
		MinCost: NewDecimal(1, 0),
		MaxCost: NewDecimal(2, 0),
		Altered: []string{RoundingDecimals},
	}

	rPc.CorrectCost(IntPointer(2), MetaRoundingUp)

	if !rPc.Equals(expected) {
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

	if !rPc.Equals(expected) {
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

	if !rPc.Equals(expected) {
		t.Errorf("\nExpected: <%+v>, \nReceived: <%+v>", expected, rPc)
	}
}

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

func TestRateProfileCostCorrectCostMinCost(t *testing.T) {
	testRPC := &RateProfileCost{
		Cost:    NewDecimal(5, 1),
		MinCost: NewDecimal(15, 1),
	}
	testRPC.CorrectCost(IntPointer(2), "")
	if testRPC.Cost.Compare(NewDecimal(15, 1)) != 0 {
		t.Errorf("\nExpecting: <1.5>,\n Received: <%+v>", testRPC.Cost)
	}
}

func TestRateProfileCostCorrectCostMaxCost(t *testing.T) {
	testRPC := &RateProfileCost{
		Cost:    NewDecimal(25, 1),
		MaxCost: NewDecimal(15, 1),
	}
	testRPC.CorrectCost(IntPointer(2), "")
	if testRPC.Cost.Compare(NewDecimal(15, 1)) != 0 {
		t.Errorf("\nExpecting: <1.5>,\n Received: <%+v>", ToJSON(testRPC.Cost))
	}
}

func TestRateSIncrementCompressEquals(t *testing.T) {
	inCr1 := &RateSIncrement{
		IncrementStart:    NewDecimal(0, 0),
		Usage:             NewDecimal(int64(time.Minute), 0),
		RateIntervalIndex: 0,
		CompressFactor:    1,
	}
	inCr2 := &RateSIncrement{
		IncrementStart:    NewDecimal(0, 0),
		Usage:             NewDecimal(int64(time.Minute), 0),
		RateIntervalIndex: 0,
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
		RateIntervalIndex: 1,
		CompressFactor:    1,
	}
	inCr2 := &RateSIncrement{
		IncrementStart:    NewDecimal(0, 0),
		Usage:             NewDecimal(int64(time.Minute), 0),
		RateIntervalIndex: 0,
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
		RateIntervalIndex: 0,
		CompressFactor:    1,
	}
	inCr2 := &RateSIncrement{
		IncrementStart:    NewDecimal(0, 0),
		Usage:             NewDecimal(int64(time.Minute), 0),
		RateIntervalIndex: 1,
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
		RateIntervalIndex: 1,
		CompressFactor:    1,
	}
	inCr2 := &RateSIncrement{
		IncrementStart:    NewDecimal(0, 0),
		Usage:             NewDecimal(int64(time.Minute), 0),
		RateIntervalIndex: 1,
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
				RateIntervalIndex: 0,
				CompressFactor:    1,
			},
			{
				IncrementStart:    NewDecimal(int64(time.Minute), 0),
				Usage:             NewDecimal(int64(time.Minute+10*time.Second), 0),
				RateIntervalIndex: 1,
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
				RateIntervalIndex: 0,
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
				RateIntervalIndex: 0,
				CompressFactor:    1,
			},
			{
				IncrementStart:    NewDecimal(0, 0),
				Usage:             NewDecimal(int64(time.Second), 0),
				RateIntervalIndex: 0,
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
				RateIntervalIndex: 0,
				CompressFactor:    1,
			},
			{
				IncrementStart:    NewDecimal(0, 0),
				Usage:             NewDecimal(int64(time.Minute), 0),
				RateIntervalIndex: 2,
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
				RateIntervalIndex: 0,
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
				RateIntervalIndex: 0,
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

func TestRatesIntervalEquals(t *testing.T) {
	rtInt1 := &RateSInterval{
		IntervalStart: NewDecimal(int64(10*time.Second), 0),
		Increments: []*RateSIncrement{
			{
				IncrementStart:    NewDecimal(int64(time.Second), 0),
				RateIntervalIndex: 1,
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
				RateIntervalIndex: 1,
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
				RateIntervalIndex: 1,
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
		RateIntervalIndex: 1,
		RateID:            "ID1",
		CompressFactor:    2,
		Usage:             NewDecimal(int64(5*time.Second), 0),
	}
	incr2 := &RateSIncrement{
		IncrementStart:    NewDecimal(int64(time.Second), 0),
		RateIntervalIndex: 1,
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

	// not equals by RateIntervalIndex
	incr1.RateIntervalIndex = 5
	if incr1.Equals(incr2, riRef, nRiRef) {
		t.Errorf("Intervals %+v and %+v are equal", ToJSON(incr1), ToJSON(incr2))
	}
	incr1.RateIntervalIndex = 1

	incr2.RateIntervalIndex = 5
	if incr1.Equals(incr2, riRef, nRiRef) {
		t.Errorf("Intervals %+v and %+v are equal", ToJSON(incr1), ToJSON(incr2))
	}
	incr2.RateIntervalIndex = 1

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

func TestRateProfileCostEquals(t *testing.T) {
	rtPrfCost := &RateProfileCost{
		ID:              "RATE_1",
		Cost:            NewDecimal(2, 1),
		MinCost:         NewDecimal(1, 2),
		MaxCost:         NewDecimal(15, 0),
		MaxCostStrategy: "*round",
		CostIntervals: []*RateSIntervalCost{
			{
				Increments: []*RateSIncrementCost{
					{
						RateIntervalIndex: 0,
						RateID:            "RATE1",
						CompressFactor:    1,
						Usage:             NewDecimal(int64(time.Minute), 0),
					},
				},
				CompressFactor: 1,
			},
		},
		Rates: map[string]*IntervalRate{
			"RATE1": {
				IntervalStart: NewDecimal(0, 0),
				RecurrentFee:  NewDecimal(2, 1),
				Unit:          NewDecimal(int64(time.Second), 1),
				Increment:     NewDecimal(int64(time.Second), 1),
			},
		},
		Altered: []string{MetaRoundingDown},
	}

	expectedRT := &RateProfileCost{
		ID:              "RATE_1",
		Cost:            NewDecimal(2, 1),
		MinCost:         NewDecimal(1, 2),
		MaxCost:         NewDecimal(15, 0),
		MaxCostStrategy: "*round",
		CostIntervals: []*RateSIntervalCost{
			{
				Increments: []*RateSIncrementCost{
					{
						RateIntervalIndex: 0,
						RateID:            "RATE1",
						CompressFactor:    1,
						Usage:             NewDecimal(int64(time.Minute), 0),
					},
				},
				CompressFactor: 1,
			},
		},
		Rates: map[string]*IntervalRate{
			"RATE1": {
				IntervalStart: NewDecimal(0, 0),
				RecurrentFee:  NewDecimal(2, 1),
				Unit:          NewDecimal(int64(time.Second), 1),
				Increment:     NewDecimal(int64(time.Second), 1),
			},
		},
		Altered: []string{MetaRoundingDown},
	}
	if !rtPrfCost.Equals(expectedRT) {
		t.Errorf("%v and \n%v are not equals", ToJSON(rtPrfCost), ToJSON(expectedRT))
	}
}

func TestRateProfileCostNotEquals(t *testing.T) {
	rtPrfCost := &RateProfileCost{
		ID:              "RATE_1",
		Cost:            NewDecimal(4, 1),
		MinCost:         NewDecimal(1, 2),
		MaxCost:         NewDecimal(15, 0),
		MaxCostStrategy: "*round",
		CostIntervals: []*RateSIntervalCost{
			{
				Increments: []*RateSIncrementCost{
					{
						RateIntervalIndex: 0,
						RateID:            "RATE1",
						CompressFactor:    1,
						Usage:             NewDecimal(int64(time.Minute), 0),
					},
				},
				CompressFactor: 1,
			},
		},
		Rates: map[string]*IntervalRate{
			"RATE1": {
				IntervalStart: NewDecimal(0, 0),
				RecurrentFee:  NewDecimal(2, 1),
				Unit:          NewDecimal(int64(time.Second), 1),
				Increment:     NewDecimal(int64(time.Second), 1),
			},
		},
		Altered: []string{MetaRoundingDown},
	}
	expectedRT := &RateProfileCost{
		ID:              "RATE_1",
		Cost:            NewDecimal(2, 1),
		MinCost:         NewDecimal(1, 2),
		MaxCost:         NewDecimal(15, 0),
		MaxCostStrategy: "*round",
		CostIntervals: []*RateSIntervalCost{
			{
				Increments: []*RateSIncrementCost{
					{
						RateIntervalIndex: 0,
						RateID:            "RATE1",
						CompressFactor:    1,
						Usage:             NewDecimal(int64(time.Minute), 0),
					},
				},
				CompressFactor: 1,
			},
		},
		Rates: map[string]*IntervalRate{
			"RATE1": {
				IntervalStart: NewDecimal(0, 0),
				RecurrentFee:  NewDecimal(2, 1),
				Unit:          NewDecimal(int64(time.Second), 1),
				Increment:     NewDecimal(int64(time.Second), 1),
			},
		},
		Altered: []string{MetaRoundingDown},
	}
	if rtPrfCost.Equals(expectedRT) {
		t.Errorf("%v and \n%v are equals", ToJSON(rtPrfCost), ToJSON(expectedRT))
	}
}

func TestRateProfileCostAlteredNotEquals(t *testing.T) {
	rtPrfCost := &RateProfileCost{
		ID:              "RATE_1",
		Cost:            NewDecimal(2, 1),
		MinCost:         NewDecimal(1, 2),
		MaxCost:         NewDecimal(15, 0),
		MaxCostStrategy: "*round",
		CostIntervals: []*RateSIntervalCost{
			{
				Increments: []*RateSIncrementCost{
					{
						RateIntervalIndex: 0,
						RateID:            "RATE1",
						CompressFactor:    1,
						Usage:             NewDecimal(int64(time.Minute), 0),
					},
				},
				CompressFactor: 1,
			},
		},
		Rates: map[string]*IntervalRate{
			"RATE1": {
				IntervalStart: NewDecimal(0, 0),
				RecurrentFee:  NewDecimal(2, 1),
				Unit:          NewDecimal(int64(time.Second), 1),
				Increment:     NewDecimal(int64(time.Second), 1),
			},
		},
		Altered: []string{MetaRoundingDown},
	}
	expectedRT := &RateProfileCost{
		ID:              "RATE_1",
		Cost:            NewDecimal(2, 1),
		MinCost:         NewDecimal(1, 2),
		MaxCost:         NewDecimal(15, 0),
		MaxCostStrategy: "*round",
		CostIntervals: []*RateSIntervalCost{
			{
				Increments: []*RateSIncrementCost{
					{
						RateIntervalIndex: 0,
						RateID:            "RATE1",
						CompressFactor:    1,
						Usage:             NewDecimal(int64(time.Minute), 0),
					},
				},
				CompressFactor: 1,
			},
		},
		Rates: map[string]*IntervalRate{
			"RATE1": {
				IntervalStart: NewDecimal(0, 0),
				RecurrentFee:  NewDecimal(2, 1),
				Unit:          NewDecimal(int64(time.Second), 1),
				Increment:     NewDecimal(int64(time.Second), 1),
			},
		},
		Altered: []string{},
	}
	if rtPrfCost.Equals(expectedRT) {
		t.Errorf("%v and \n%v are equals", ToJSON(rtPrfCost), ToJSON(expectedRT))
	}
}

func TestRateProfileCostCINotEquals(t *testing.T) {
	rtPrfCost := &RateProfileCost{
		ID:              "RATE_1",
		Cost:            NewDecimal(2, 1),
		MinCost:         NewDecimal(1, 2),
		MaxCost:         NewDecimal(15, 0),
		MaxCostStrategy: "*round",
		CostIntervals: []*RateSIntervalCost{
			{
				Increments: []*RateSIncrementCost{
					{
						RateIntervalIndex: 0,
						RateID:            "RATE1",
						CompressFactor:    1,
						Usage:             NewDecimal(int64(time.Minute), 0),
					},
				},
				CompressFactor: 1,
			},
		},
		Rates: map[string]*IntervalRate{
			"RATE1": {
				IntervalStart: NewDecimal(0, 0),
				RecurrentFee:  NewDecimal(2, 1),
				Unit:          NewDecimal(int64(time.Second), 1),
				Increment:     NewDecimal(int64(time.Second), 1),
			},
		},
		Altered: []string{MetaRoundingDown},
	}
	expectedRT := &RateProfileCost{
		ID:              "RATE_1",
		Cost:            NewDecimal(2, 1),
		MinCost:         NewDecimal(1, 2),
		MaxCost:         NewDecimal(15, 0),
		MaxCostStrategy: "*round",
		CostIntervals: []*RateSIntervalCost{
			{
				Increments: []*RateSIncrementCost{
					{
						RateIntervalIndex: 1,
						RateID:            "RATE2",
						CompressFactor:    2,
						Usage:             NewDecimal(int64(time.Minute), 0),
					},
				},
				CompressFactor: 1,
			},
		},
		Rates: map[string]*IntervalRate{
			"RATE1": {
				IntervalStart: NewDecimal(0, 0),
				RecurrentFee:  NewDecimal(2, 1),
				Unit:          NewDecimal(int64(time.Second), 1),
				Increment:     NewDecimal(int64(time.Second), 1),
			},
		},
		Altered: []string{MetaRoundingDown},
	}
	if rtPrfCost.Equals(expectedRT) {
		t.Errorf("%v and \n%v are equals", ToJSON(rtPrfCost), ToJSON(expectedRT))
	}
}

func TestAsRatesIntervalsCost(t *testing.T) {
	rtsIntrvl := &RateSInterval{
		IntervalStart: NewDecimal(0, 0),
		Increments: []*RateSIncrement{
			{
				IncrementStart:    NewDecimal(0, 0),
				RateIntervalIndex: 0,
				RateID:            "RATE1",
				CompressFactor:    1,
				Usage:             NewDecimal(int64(time.Minute), 0),
			},
			{
				IncrementStart:    NewDecimal(int64(time.Minute), 0),
				RateIntervalIndex: 1,
				RateID:            "RATE1",
				CompressFactor:    5,
				Usage:             NewDecimal(int64(2*time.Minute), 0),
			},
		},
		CompressFactor: 1,
	}
	expRtsIntCost := &RateSIntervalCost{
		Increments: []*RateSIncrementCost{
			{
				RateIntervalIndex: 0,
				RateID:            "RATE1",
				CompressFactor:    1,
				Usage:             NewDecimal(int64(time.Minute), 0),
			},
			{
				RateIntervalIndex: 1,
				RateID:            "RATE1",
				CompressFactor:    5,
				Usage:             NewDecimal(int64(2*time.Minute), 0),
			},
		},
		CompressFactor: 1,
	}
	if rcv := rtsIntrvl.AsRatesIntervalsCost(); !reflect.DeepEqual(rcv, expRtsIntCost) {
		t.Errorf("Expected %+v, received %+v", ToJSON(expRtsIntCost), ToJSON(rcv))
	} else if !rcv.Equals(expRtsIntCost, nil, nil) {
		t.Errorf("Expected %+v, received %+v", ToJSON(expRtsIntCost), ToJSON(rcv))
	}
}

func TestRateSIncrementCost(t *testing.T) {
	rIc := &RateSIncrement{
		IncrementStart:    NewDecimal(int64(2), 0),
		RateIntervalIndex: 0,
		RateID:            "RI1",
		CompressFactor:    int64(3),
		Usage:             NewDecimal(int64(30), 0),
	}

	rts := map[string]*IntervalRate{
		"RI1": {
			IntervalStart: NewDecimal(int64(2), 0),
			FixedFee:      NewDecimal(int64(10), 0),
			RecurrentFee:  NewDecimal(int64(5), 0),
		},
	}

	cost := rIc.Cost(rts)
	exp := decimal.WithContext(DecimalContext).SetUint64(15)
	if !reflect.DeepEqual(cost, exp) {
		t.Errorf("Expected %T \n but received \n %T", exp, cost)
	}
}

func TestRateSIncrementCostNotNil(t *testing.T) {
	rIc := &RateSIncrement{
		IncrementStart:    NewDecimal(int64(2), 0),
		RateIntervalIndex: 0,
		RateID:            "RI1",
		CompressFactor:    int64(3),
		Usage:             NewDecimal(int64(30), 0),
		cost:              decimal.WithContext(DecimalContext).SetUint64(15),
	}

	rts := map[string]*IntervalRate{
		"RI1": {
			IntervalStart: NewDecimal(int64(2), 0),
			FixedFee:      NewDecimal(int64(10), 0),
			RecurrentFee:  NewDecimal(int64(5), 0),
		},
	}

	cost := rIc.Cost(rts)
	exp := decimal.WithContext(DecimalContext).SetUint64(15)
	if !reflect.DeepEqual(cost, exp) {
		t.Errorf("Expected %T \n but received \n %T", exp, cost)
	}
}

func TestRateSIncrementCostNoID(t *testing.T) {
	rIc := &RateSIncrement{
		IncrementStart:    NewDecimal(int64(2), 0),
		RateIntervalIndex: 0,
		RateID:            "RI1",
		CompressFactor:    int64(3),
		Usage:             NewDecimal(int64(30), 0),
	}

	rts := map[string]*IntervalRate{
		"not_RI1": {
			IntervalStart: NewDecimal(int64(2), 0),
			FixedFee:      NewDecimal(int64(10), 0),
			RecurrentFee:  NewDecimal(int64(5), 0),
		},
	}

	cost := rIc.Cost(rts)
	if cost != nil {
		t.Errorf("expected cost to be nil, received <%+v>", cost)
	}
}

func TestRateSIncrementCostFixedFee(t *testing.T) {
	rIc := &RateSIncrement{
		IncrementStart:    NewDecimal(int64(2), 0),
		RateIntervalIndex: 0,
		RateID:            "RI1",
		CompressFactor:    int64(3),
		Usage:             NewDecimal(int64(-1), 0),
	}

	rts := map[string]*IntervalRate{
		"RI1": {
			IntervalStart: NewDecimal(int64(2), 0),
			FixedFee:      NewDecimal(int64(10), 0),
			RecurrentFee:  NewDecimal(int64(5), 0),
		},
	}

	cost := rIc.Cost(rts)
	exp := decimal.WithContext(DecimalContext).SetUint64(10)
	if !reflect.DeepEqual(cost, exp) {
		t.Errorf("Expected %v \n but received \n %v", exp, cost)
	}
}

func TestRateSIncrementCostDiffUnitIncrement(t *testing.T) {
	rIc := &RateSIncrement{
		IncrementStart:    NewDecimal(int64(2), 0),
		RateIntervalIndex: 0,
		RateID:            "RI1",
		CompressFactor:    int64(1),
		Usage:             NewDecimal(int64(2), 0),
	}

	rts := map[string]*IntervalRate{
		"RI1": {
			IntervalStart: NewDecimal(int64(2), 0),
			FixedFee:      NewDecimal(int64(10), 0),
			RecurrentFee:  NewDecimal(int64(2), 0),
			Unit:          NewDecimal(int64(2), 0),
			Increment:     NewDecimal(int64(3), 0),
		},
	}

	cost := rIc.Cost(rts)
	exp := decimal.WithContext(DecimalContext).SetUint64(3)
	if !reflect.DeepEqual(cost, exp) {
		t.Errorf("Expected %v \n but received \n %v", exp, cost)
	}
}

func TestAPIIntervalRateAsIR(t *testing.T) {
	ext := &APIIntervalRate{
		IntervalStart: "2",
		FixedFee:      Float64Pointer(10),
		RecurrentFee:  Float64Pointer(2),
		Unit:          Float64Pointer(2),
		Increment:     Float64Pointer(3),
	}

	exp := &IntervalRate{
		IntervalStart: NewDecimal(int64(2), 0),
		FixedFee:      NewDecimal(int64(10), 0),
		RecurrentFee:  NewDecimal(int64(2), 0),
		Unit:          NewDecimal(int64(2), 0),
		Increment:     NewDecimal(int64(3), 0),
	}

	rcv, err := ext.AsIntervalRate()
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(exp, rcv) {
		t.Errorf("Expected %v \n but received \n %v", exp, rcv)
	}
}

func TestAPIIntervalRateAsIRError(t *testing.T) {
	ext := &APIIntervalRate{
		IntervalStart: "not_a_decimal",
		FixedFee:      Float64Pointer(10),
		RecurrentFee:  Float64Pointer(2),
		Unit:          Float64Pointer(2),
		Increment:     Float64Pointer(3),
	}

	exp := "can't convert <not_a_decimal> to decimal"

	_, err := ext.AsIntervalRate()
	if !reflect.DeepEqual(exp, err.Error()) {
		t.Errorf("Expected %v \n but received \n %v", exp, err.Error())
	}
}

func TestAPIRateAsRate(t *testing.T) {
	aR := &APIRate{
		ID:              "rate_id1",
		FilterIDs:       []string{"fltr1"},
		ActivationTimes: "1 1 3",
		Weights:         ";10",
		Blocker:         false,
		IntervalRates: []*APIIntervalRate{
			{
				IntervalStart: "2",
				FixedFee:      Float64Pointer(10),
				RecurrentFee:  Float64Pointer(2),
				Unit:          Float64Pointer(2),
				Increment:     Float64Pointer(3),
			},
		},
	}

	exp := &Rate{
		ID:              "rate_id1",
		FilterIDs:       []string{"fltr1"},
		ActivationTimes: "1 1 3",
		Weights: DynamicWeights{
			{
				Weight: 10,
			},
		},
		Blocker: false,
		IntervalRates: []*IntervalRate{
			{
				IntervalStart: NewDecimal(int64(2), 0),
				FixedFee:      NewDecimal(int64(10), 0),
				RecurrentFee:  NewDecimal(int64(2), 0),
				Unit:          NewDecimal(int64(2), 0),
				Increment:     NewDecimal(int64(3), 0),
			},
		},
	}

	rcv, err := aR.AsRate()
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(exp, rcv) {
		t.Errorf("Expected %v \n but received \n %v", exp, rcv)
	}
}

func TestAPIRateAsRateError(t *testing.T) {
	aR := &APIRate{
		ID:              "rate_id1",
		FilterIDs:       []string{"fltr1"},
		ActivationTimes: "1 1 3",
		Weights:         ";10",
		Blocker:         false,
		IntervalRates: []*APIIntervalRate{
			{
				IntervalStart: "not_a_decimal",
				FixedFee:      Float64Pointer(10),
				RecurrentFee:  Float64Pointer(2),
				Unit:          Float64Pointer(2),
				Increment:     Float64Pointer(3),
			},
		},
	}

	exp := "can't convert <not_a_decimal> to decimal"

	_, err := aR.AsRate()
	if !reflect.DeepEqual(exp, err.Error()) {
		t.Errorf("Expected %v \n but received \n %v", exp, err.Error())
	}
}

func TestIntervalRateEqualsNilIR(t *testing.T) {
	var iR *IntervalRate
	iR = nil
	iR.Equals(nil)

	iR = &IntervalRate{
		Unit: NewDecimal(int64(2), 0),
	}
	iR.Equals(nil)
}

func TestRateSIntervalCostEquals(t *testing.T) {
	rIC := &RateSIntervalCost{
		Increments: nil,
	}
	nRIc := &RateSIntervalCost{
		Increments: []*RateSIncrementCost{
			{
				RateIntervalIndex: 0,
				RateID:            "RI1",
				CompressFactor:    int64(1),
				Usage:             NewDecimal(int64(2), 0),
			},
		},
	}
	if rIC.Equals(nRIc, nil, nil) {
		t.Error("Shouldn't match")
	}
	rIC.Increments = []*RateSIncrementCost{
		{
			RateIntervalIndex: 3,
			RateID:            "RI2",
			CompressFactor:    int64(3),
			Usage:             NewDecimal(int64(5), 0),
		},
	}
	if rIC.Equals(nRIc, nil, nil) {
		t.Error("Shouldn't match")
	}
}
func TestRateSIntervalCost(t *testing.T) {
	rIv := &RateSInterval{
		Increments: []*RateSIncrement{
			{
				RateID: "ir1",
				Usage:  NewDecimal(int64(-1), 0),
			},
		},
	}

	rts := map[string]*IntervalRate{
		"ir1": {
			FixedFee: NewDecimal(int64(2), 0),
		},
	}

	rcv := rIv.Cost(rts)
	exp := decimal.WithContext(DecimalContext).SetUint64(2)
	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("expected <%v>,\nreceived <%v>", exp, rcv)
	}
}

func TestRateProfile(t *testing.T) {
	rp := RateProfile{Rates: make(map[string]*Rate)}
	exp := RateProfile{
		Tenant:    "cgrates.org",
		ID:        ID,
		FilterIDs: []string{"fltr1", "*string:~*req.Account:1001"},
		Weights: DynamicWeights{
			{
				Weight: 10,
			},
		},
		MinCost:         NewDecimal(10, 0),
		MaxCost:         NewDecimal(10, 0),
		MaxCostStrategy: "strategy",
		Rates: map[string]*Rate{
			"rat1": {
				ID:        "rat1",
				FilterIDs: []string{"fltr1"},
				Weights: DynamicWeights{
					{
						Weight: 20,
					},
				},
				ActivationTimes: "* * * * *",
				Blocker:         true,
				IntervalRates: []*IntervalRate{{
					IntervalStart: NewDecimal(10, 0),
					FixedFee:      NewDecimal(10, 0),
					RecurrentFee:  NewDecimal(10, 0),
					Unit:          NewDecimal(10, 0),
					Increment:     NewDecimal(10, 0),
				}},
			},
		},
	}
	if err := rp.Set([]string{}, "", false); err != ErrWrongPath {
		t.Error(err)
	}
	if err := rp.Set([]string{"NotAField"}, "", false); err != ErrWrongPath {
		t.Error(err)
	}
	if err := rp.Set([]string{"NotAField", "1"}, "", false); err != ErrWrongPath {
		t.Error(err)
	}

	if err := rp.Set([]string{Tenant}, "cgrates.org", false); err != nil {
		t.Error(err)
	}
	if err := rp.Set([]string{ID}, ID, false); err != nil {
		t.Error(err)
	}
	if err := rp.Set([]string{FilterIDs}, "fltr1;*string:~*req.Account:1001", false); err != nil {
		t.Error(err)
	}
	if err := rp.Set([]string{Weights}, ";10", false); err != nil {
		t.Error(err)
	}
	if err := rp.Set([]string{MinCost}, "10", false); err != nil {
		t.Error(err)
	}
	if err := rp.Set([]string{MaxCost}, "10", false); err != nil {
		t.Error(err)
	}
	if err := rp.Set([]string{MaxCostStrategy}, "strategy", false); err != nil {
		t.Error(err)
	}

	if err := rp.Set([]string{Rates + "[rat1]", ID}, "rat1", false); err != nil {
		t.Error(err)
	}

	if err := rp.Set([]string{Rates, "rat1", FilterIDs}, "fltr1", false); err != nil {
		t.Error(err)
	}
	if err := rp.Set([]string{Rates, "rat1", Weights}, ";20", false); err != nil {
		t.Error(err)
	}
	if err := rp.Set([]string{Rates, "rat1", ActivationTimes}, "* * * * *", false); err != nil {
		t.Error(err)
	}
	if err := rp.Set([]string{Rates, "rat1", Blocker}, "true", false); err != nil {
		t.Error(err)
	}

	if err := rp.Set([]string{Rates, "rat1", "Wrong"}, "true", false); err != ErrWrongPath {
		t.Error(err)
	}
	if err := rp.Set([]string{Rates, "rat1", "Wrong", "Path"}, "true", false); err != ErrWrongPath {
		t.Error(err)
	}
	if err := rp.Set([]string{Rates, "rat1", "Wrong", "Path", "2"}, "true", false); err != ErrWrongPath {
		t.Error(err)
	}
	if err := rp.Set([]string{Rates, "rat1", IntervalRates, "Wrong"}, "true", false); err != ErrWrongPath {
		t.Error(err)
	}

	if err := rp.Set([]string{Rates, "rat1", IntervalRates, IntervalStart}, "10", false); err != nil {
		t.Error(err)
	}
	if err := rp.Set([]string{Rates, "rat1", IntervalRates, FixedFee}, "10", false); err != nil {
		t.Error(err)
	}
	if err := rp.Set([]string{Rates, "rat1", IntervalRates, RecurrentFee}, "10", false); err != nil {
		t.Error(err)
	}
	if err := rp.Set([]string{Rates, "rat1", IntervalRates, Unit}, "10", false); err != nil {
		t.Error(err)
	}
	if err := rp.Set([]string{Rates, "rat1", IntervalRates, Increment}, "10", false); err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(exp, rp) {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(rp))
	}
}

func TestRateProfileFieldAsInterface(t *testing.T) {
	rp := RateProfile{
		Tenant:          "cgrates.org",
		ID:              ID,
		FilterIDs:       []string{"fltr1", "*string:~*req.Account:1001"},
		Weights:         DynamicWeights{{}},
		MinCost:         NewDecimal(10, 0),
		MaxCost:         NewDecimal(10, 0),
		MaxCostStrategy: "strategy",
		Rates: map[string]*Rate{
			"rat1": {
				ID:              "rat1",
				FilterIDs:       []string{"fltr1"},
				Weights:         DynamicWeights{{}},
				ActivationTimes: "* * * * *",
				Blocker:         true,
				IntervalRates: []*IntervalRate{{
					IntervalStart: NewDecimal(10, 0),
					FixedFee:      NewDecimal(10, 0),
					RecurrentFee:  NewDecimal(10, 0),
					Unit:          NewDecimal(10, 0),
					Increment:     NewDecimal(10, 0),
				}},
			},
		},
	}
	if _, err := rp.FieldAsInterface(nil); err != ErrNotFound {
		t.Fatal(err)
	}
	if _, err := rp.FieldAsInterface([]string{"field"}); err != ErrNotFound {
		t.Fatal(err)
	}
	if _, err := rp.FieldAsInterface([]string{"field", ""}); err != ErrNotFound {
		t.Fatal(err)
	}
	if _, err := rp.FieldAsInterface([]string{Opts + "[f]"}); err != ErrNotFound {
		t.Fatal(err)
	}
	if _, err := rp.FieldAsInterface([]string{Opts + "[f]", ""}); err != ErrNotFound {
		t.Fatal(err)
	}
	if val, err := rp.FieldAsInterface([]string{Tenant}); err != nil {
		t.Fatal(err)
	} else if exp := "cgrates.org"; exp != val {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, err := rp.FieldAsInterface([]string{ID}); err != nil {
		t.Fatal(err)
	} else if exp := ID; exp != val {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, err := rp.FieldAsInterface([]string{Weights}); err != nil {
		t.Fatal(err)
	} else if exp := ";0"; exp != val {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, err := rp.FieldAsInterface([]string{MaxCostStrategy}); err != nil {
		t.Fatal(err)
	} else if exp := "strategy"; exp != val {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, err := rp.FieldAsInterface([]string{FilterIDs}); err != nil {
		t.Fatal(err)
	} else if exp := rp.FilterIDs; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, err := rp.FieldAsInterface([]string{FilterIDs + "[0]"}); err != nil {
		t.Fatal(err)
	} else if exp := rp.FilterIDs[0]; exp != val {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, err := rp.FieldAsInterface([]string{Rates}); err != nil {
		t.Fatal(err)
	} else if exp := rp.Rates; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	expErrMsg := `strconv.Atoi: parsing "a": invalid syntax`
	if _, err := rp.FieldAsInterface([]string{FilterIDs + "[a]"}); err == nil || err.Error() != expErrMsg {
		t.Errorf("Expeceted: %v, received: %v", expErrMsg, err)
	}
	if val, err := rp.FieldAsInterface([]string{Rates + "[rat1]"}); err != nil {
		t.Fatal(err)
	} else if exp := rp.Rates["rat1"]; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, err := rp.FieldAsInterface([]string{Rates, "rat1"}); err != nil {
		t.Fatal(err)
	} else if exp := rp.Rates["rat1"]; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, err := rp.FieldAsInterface([]string{MinCost}); err != nil {
		t.Fatal(err)
	} else if exp := rp.MinCost; exp.Cmp(val.(*Decimal).Big) != 0 {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, err := rp.FieldAsInterface([]string{MaxCost}); err != nil {
		t.Fatal(err)
	} else if exp := rp.MaxCost; exp.Cmp(val.(*Decimal).Big) != 0 {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if _, err := rp.FieldAsInterface([]string{Rates, "rat2"}); err != ErrNotFound {
		t.Fatal(err)
	}
	if _, err := rp.FieldAsInterface([]string{Rates, "rat1", ""}); err != ErrNotFound {
		t.Fatal(err)
	}
	if _, err := rp.FieldAsInterface([]string{Rates, "rat1", "", ""}); err != ErrNotFound {
		t.Fatal(err)
	}
	if _, err := rp.FieldAsInterface([]string{Rates, "rat1", "", "", ""}); err != ErrNotFound {
		t.Fatal(err)
	}

	if val, err := rp.FieldAsInterface([]string{Rates + "[rat1]", ID}); err != nil {
		t.Fatal(err)
	} else if exp := rp.Rates["rat1"].ID; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, err := rp.FieldAsInterface([]string{Rates + "[rat1]", FilterIDs}); err != nil {
		t.Fatal(err)
	} else if exp := rp.Rates["rat1"].FilterIDs; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, err := rp.FieldAsInterface([]string{Rates + "[rat1]", IntervalRates}); err != nil {
		t.Fatal(err)
	} else if exp := rp.Rates["rat1"].IntervalRates; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, err := rp.FieldAsInterface([]string{Rates + "[rat1]", Blocker}); err != nil {
		t.Fatal(err)
	} else if exp := rp.Rates["rat1"].Blocker; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, err := rp.FieldAsInterface([]string{Rates + "[rat1]", ActivationTimes}); err != nil {
		t.Fatal(err)
	} else if exp := rp.Rates["rat1"].ActivationTimes; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, err := rp.FieldAsInterface([]string{Rates + "[rat1]", Weights}); err != nil {
		t.Fatal(err)
	} else if exp := ";0"; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, err := rp.FieldAsInterface([]string{Rates + "[rat1]", FilterIDs + "[0]"}); err != nil {
		t.Fatal(err)
	} else if exp := rp.Rates["rat1"].FilterIDs[0]; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, err := rp.FieldAsInterface([]string{Rates + "[rat1]", IntervalRates + "[0]"}); err != nil {
		t.Fatal(err)
	} else if exp := rp.Rates["rat1"].IntervalRates[0]; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if _, err := rp.FieldAsInterface([]string{Rates + "[rat1]", IntervalRates + "[0]", ""}); err != ErrNotFound {
		t.Fatal(err)
	}
	if _, err := rp.Rates["rat1"].IntervalRates[0].FieldAsInterface([]string{"", ""}); err != ErrNotFound {
		t.Fatal(err)
	}
	if val, err := rp.FieldAsInterface([]string{Rates + "[rat1]", IntervalRates + "[0]", IntervalStart}); err != nil {
		t.Fatal(err)
	} else if exp := rp.Rates["rat1"].IntervalRates[0].IntervalStart; exp.Cmp(val.(*Decimal).Big) != 0 {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, err := rp.FieldAsInterface([]string{Rates + "[rat1]", IntervalRates + "[0]", FixedFee}); err != nil {
		t.Fatal(err)
	} else if exp := rp.Rates["rat1"].IntervalRates[0].FixedFee; exp.Cmp(val.(*Decimal).Big) != 0 {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, err := rp.FieldAsInterface([]string{Rates + "[rat1]", IntervalRates + "[0]", RecurrentFee}); err != nil {
		t.Fatal(err)
	} else if exp := rp.Rates["rat1"].IntervalRates[0].RecurrentFee; exp.Cmp(val.(*Decimal).Big) != 0 {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, err := rp.FieldAsInterface([]string{Rates + "[rat1]", IntervalRates + "[0]", Unit}); err != nil {
		t.Fatal(err)
	} else if exp := rp.Rates["rat1"].IntervalRates[0].Unit; exp.Cmp(val.(*Decimal).Big) != 0 {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, err := rp.FieldAsInterface([]string{Rates + "[rat1]", IntervalRates + "[0]", Increment}); err != nil {
		t.Fatal(err)
	} else if exp := rp.Rates["rat1"].IntervalRates[0].Increment; exp.Cmp(val.(*Decimal).Big) != 0 {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}

	if _, err := rp.FieldAsString([]string{""}); err != ErrNotFound {
		t.Fatal(err)
	}
	if val, err := rp.FieldAsString([]string{Tenant}); err != nil {
		t.Fatal(err)
	} else if exp := "cgrates.org"; exp != val {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, exp := rp.String(), ToJSON(rp); exp != val {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}

	if _, err := rp.Rates["rat1"].FieldAsString([]string{""}); err != ErrNotFound {
		t.Fatal(err)
	}
	if val, err := rp.Rates["rat1"].FieldAsString([]string{ID}); err != nil {
		t.Fatal(err)
	} else if exp := "rat1"; exp != val {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, exp := rp.Rates["rat1"].String(), ToJSON(rp.Rates["rat1"]); exp != val {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}

	if _, err := rp.Rates["rat1"].IntervalRates[0].FieldAsString([]string{""}); err != ErrNotFound {
		t.Fatal(err)
	}
	if val, err := rp.Rates["rat1"].IntervalRates[0].FieldAsString([]string{IntervalStart}); err != nil {
		t.Fatal(err)
	} else if exp := "10"; exp != val {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, exp := rp.Rates["rat1"].IntervalRates[0].String(), ToJSON(rp.Rates["rat1"].IntervalRates[0]); exp != val {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
}

func TestRateProfileMerge(t *testing.T) {
	acc := &RateProfile{
		Rates: map[string]*Rate{
			"rat1": {},
			"rat2": {Blocker: true},
		},
	}
	exp := &RateProfile{
		Tenant:          "cgrates.org",
		ID:              "ID",
		FilterIDs:       []string{"fltr1"},
		Weights:         DynamicWeights{{}},
		MinCost:         NewDecimal(10, 0),
		MaxCost:         NewDecimal(10, 0),
		MaxCostStrategy: "strategy",
		Rates: map[string]*Rate{
			"rat1": {
				ID:              "rat1",
				FilterIDs:       []string{"fltr1"},
				ActivationTimes: "* * * * *",
				Weights:         DynamicWeights{{}},
				Blocker:         true,
				IntervalRates:   []*IntervalRate{{}},
			},
			"rat2": {
				ID:      "rat2",
				Blocker: true,
			},
			"rat3": {},
		},
	}
	if acc.Merge(&RateProfile{
		Tenant:          "cgrates.org",
		ID:              "ID",
		FilterIDs:       []string{"fltr1"},
		Weights:         DynamicWeights{{}},
		MinCost:         NewDecimal(10, 0),
		MaxCost:         NewDecimal(10, 0),
		MaxCostStrategy: "strategy",
		Rates: map[string]*Rate{
			"rat1": {
				ID:              "rat1",
				FilterIDs:       []string{"fltr1"},
				ActivationTimes: "* * * * *",
				Weights:         DynamicWeights{{}},
				Blocker:         true,
				IntervalRates:   []*IntervalRate{{}},
			},
			"rat2": {
				ID: "rat2",
			},
			"rat3": {},
		},
	}); !reflect.DeepEqual(exp, acc) {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(acc))
	}

}

func TestLibratesIntervalRateCloneEmpty(t *testing.T) {
	iR := &IntervalRate{}
	if rcv := iR.Clone(); !reflect.DeepEqual(rcv, iR) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>",
			ToJSON(iR), ToJSON(rcv))
	}
}

func TestLibratesIntervalRateClone(t *testing.T) {
	iR := &IntervalRate{
		IntervalStart: NewDecimalFromFloat64(1.2),
		FixedFee:      NewDecimalFromFloat64(1.234),
		RecurrentFee:  NewDecimalFromFloat64(0.5),
		Unit:          NewDecimalFromFloat64(7.1),
		Increment:     NewDecimalFromFloat64(-321),
	}
	if rcv := iR.Clone(); !reflect.DeepEqual(rcv, iR) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>",
			ToJSON(iR), ToJSON(rcv))
	}
}

func TestLibratesRateSIntervalCloneEmpty(t *testing.T) {
	ri := &RateSInterval{}
	if rcv := ri.Clone(); !reflect.DeepEqual(rcv, ri) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>",
			ToJSON(ri), ToJSON(rcv))
	}
}

func TestLibratesRateSIntervalClone(t *testing.T) {
	ri := &RateSInterval{
		IntervalStart: NewDecimalFromFloat64(1.234),
		Increments: []*RateSIncrement{
			{
				IncrementStart:    NewDecimalFromFloat64(1.234),
				RateIntervalIndex: 1,
				RateID:            "Rate1",
				CompressFactor:    1,
				Usage:             NewDecimalFromFloat64(-321),
				cost:              decimal.New(4321, 5),
			},
			{
				IncrementStart:    NewDecimalFromFloat64(4.321),
				RateIntervalIndex: 1,
				RateID:            "Rate2",
				CompressFactor:    1,
				Usage:             NewDecimalFromFloat64(-123),
				cost:              decimal.New(123, 1),
			},
		},
		CompressFactor: 1,
		cost:           decimal.New(4321, 5),
	}
	if rcv := ri.Clone(); !reflect.DeepEqual(rcv, ri) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>",
			ToJSON(ri), ToJSON(rcv))
	}
}

func TestLibratesRateSIncrementCloneEmpty(t *testing.T) {
	ri := &RateSIncrement{}
	if rcv := ri.Clone(); !reflect.DeepEqual(rcv, ri) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>",
			ToJSON(ri), ToJSON(rcv))
	}
}

func TestLibratesRateSIncrementClone(t *testing.T) {
	ri := &RateSIncrement{
		IncrementStart:    NewDecimalFromFloat64(1.234),
		RateIntervalIndex: 1,
		RateID:            "testRateID",
		CompressFactor:    1,
		Usage:             NewDecimalFromFloat64(-321),
		cost:              decimal.New(4321, 5),
	}
	if rcv := ri.Clone(); !reflect.DeepEqual(rcv, ri) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>",
			ToJSON(ri), ToJSON(rcv))
	}
}

func TestAsRateAPIConvert(t *testing.T) {
	ext := APIRate{

		Weights: "notEmpty",
	}
	expErr := "invalid DynamicWeight format for string <notEmpty>"

	if _, err := ext.AsRate(); err == nil || err.Error() != expErr {
		t.Errorf("%+v", err)
	}
}

func TestCloneRate(t *testing.T) {
	rt := &Rate{
		FilterIDs: []string{"id1", "id2"},
		Weights: DynamicWeights{
			{
				Weight: 0,
			},
		},
		IntervalRates: []*IntervalRate{
			{
				IntervalStart: NewDecimal(0, 0),
				RecurrentFee:  NewDecimal(12, 2),
			},
		},
	}

	exp := &Rate{
		FilterIDs: []string{"id1", "id2"},
		Weights: DynamicWeights{
			{
				Weight: 0,
			},
		},
		IntervalRates: []*IntervalRate{
			{
				IntervalStart: NewDecimal(0, 0),
				RecurrentFee:  NewDecimal(12, 2),
			},
		},
	}
	if rcv := rt.Clone(); !reflect.DeepEqual(exp, rcv) {

		t.Errorf("Expected <%v>, Received <%v>", exp, rcv)
	}

}

func TestEqualsNRpCt(t *testing.T) {
	rtPrfCost := &RateProfileCost{
		ID:              "RATE_1",
		Cost:            NewDecimal(2, 1),
		MinCost:         NewDecimal(1, 2),
		MaxCost:         NewDecimal(15, 0),
		MaxCostStrategy: "*round",
		CostIntervals: []*RateSIntervalCost{
			{
				Increments: []*RateSIncrementCost{
					{
						RateIntervalIndex: 0,
						RateID:            "RATE1",
						CompressFactor:    1,
						Usage:             NewDecimal(int64(time.Minute), 0),
					},
				},
				CompressFactor: 1,
			},
		},
		Rates: map[string]*IntervalRate{
			"RATE1": {
				IntervalStart: NewDecimal(0, 0),
				RecurrentFee:  NewDecimal(2, 1),
				Unit:          NewDecimal(int64(time.Second), 1),
				Increment:     NewDecimal(int64(time.Second), 1),
			},
		},
		Altered: []string{MetaRoundingDown},
	}
	expectedRT := &RateProfileCost{
		ID:              "RATE_1",
		Cost:            NewDecimal(2, 1),
		MinCost:         NewDecimal(1, 2),
		MaxCost:         NewDecimal(15, 0),
		MaxCostStrategy: "*round",
		CostIntervals: []*RateSIntervalCost{
			{
				Increments: []*RateSIncrementCost{
					{
						RateIntervalIndex: 0,
						RateID:            "RATE1",
						CompressFactor:    1,
						Usage:             NewDecimal(int64(time.Minute), 0),
					},
				},
				CompressFactor: 1,
			},
		},
		Rates: map[string]*IntervalRate{
			"RATE1": {
				IntervalStart: NewDecimal(0, 0),
				RecurrentFee:  NewDecimal(2, 1),
				Unit:          NewDecimal(int64(time.Second), 1),
				Increment:     NewDecimal(int64(time.Second), 1),
			},
		},
		Altered: []string{MetaRoundingUp},
	}
	if rtPrfCost.Equals(expectedRT) {
		t.Errorf("%v and \n%v are equals", ToJSON(rtPrfCost), ToJSON(expectedRT))
	}
}

func TestMergeRate(t *testing.T) {
	rt := &Rate{ID: "rate_id1",
		FilterIDs:       []string{"fltr1"},
		ActivationTimes: "1 1 3",
		Weights: DynamicWeights{
			{
				Weight: 10,
			},
		},
		Blocker: false,
		IntervalRates: []*IntervalRate{
			{
				IntervalStart: NewDecimal(int64(1), 0),
				FixedFee:      NewDecimal(int64(10), 0),
				RecurrentFee:  NewDecimal(int64(2), 0),
				Unit:          NewDecimal(int64(2), 0),
				Increment:     NewDecimal(int64(3), 0),
			},
		}}
	vi := &Rate{
		ID:              "rate_id1",
		FilterIDs:       []string{"fltr1"},
		ActivationTimes: "1 1 3",
		Weights: DynamicWeights{
			{
				Weight: 10,
			},
		},
		Blocker: false,
		IntervalRates: []*IntervalRate{
			{
				IntervalStart: NewDecimal(int64(1), 0),
				FixedFee:      NewDecimal(int64(10), 0),
				RecurrentFee:  NewDecimal(int64(2), 0),
				Unit:          NewDecimal(int64(2), 0),
				Increment:     NewDecimal(int64(3), 0),
			},
		},
	}

	rt.Merge(vi)

}

type MockMarshaler struct{}

func (MockMarshaler) Marshal(any) ([]byte, error) {
	return nil, ErrNotImplemented
}

func (MockMarshaler) Unmarshal([]byte, any) error {
	return ErrNotImplemented
}

func TestAsDataDBMap(t *testing.T) {

	ms := MockMarshaler{}
	rp := &RateProfile{
		FilterIDs: []string{"fltr1", "fltr2"},
		Weights: DynamicWeights{{

			Weight: 10,
		}},
		MinCost:         NewDecimal(10, 0),
		MaxCost:         NewDecimal(10, 0),
		MaxCostStrategy: "strategy",
		Rates: map[string]*Rate{
			"rat1": {ID: "rat1"}, "rat2": {ID: "rat2"},
		},
	}

	exp := "{\"FilterIDs\":\"fltr1;fltr2\",\"MaxCost\":\"10\",\"MaxCostStrategy\":\"strategy\",\"MinCost\":\"10\",\"Rates:rat1\":\"{\\\"ID\\\":\\\"rat1\\\",\\\"FilterIDs\\\":null,\\\"ActivationTimes\\\":\\\"\\\",\\\"Weights\\\":null,\\\"Blocker\\\":false,\\\"IntervalRates\\\":null}\",\"Rates:rat2\":\"{\\\"ID\\\":\\\"rat2\\\",\\\"FilterIDs\\\":null,\\\"ActivationTimes\\\":\\\"\\\",\\\"Weights\\\":null,\\\"Blocker\\\":false,\\\"IntervalRates\\\":null}\",\"Weights\":\";10\"}"

	if rcv, err := rp.AsDataDBMap(JSONMarshaler{}); err != nil {
		t.Error(err)

	} else if !reflect.DeepEqual(exp, ToJSON(rcv)) {
		t.Errorf("Expected <%v %T>, \nReceived <%v %T>", exp, exp, ToJSON(rcv), ToJSON(rcv))
	}

	if _, err := rp.AsDataDBMap(ms); err == nil || err != ErrNotImplemented {
		t.Errorf("Expected <%v>, Received <%v>", ErrNotImplemented, err)
	}
}

func TestNewRateProfileFromMapDataDBMap(t *testing.T) {
	mapRP := map[string]any{
		"FilterIDs":  "fltrID1",
		"Weights":    "fltrID1;20",
		"MinCost":    "2",
		"MaxCost":    "10",
		"Rates:rat1": "{\"ID\":\"rat1\",\"FilterIDs\":null,\"ActivationTimes\":\"\",\"Weights\":null,\"Blocker\":false,\"IntervalRates\":null}",
	}
	if rcv, err := NewRateProfileFromMapDataDBMap("cgrates.org", "ExID", mapRP, JSONMarshaler{}); err != nil {
		t.Error(rcv, err)
	}

	mapRP = map[string]any{
		"FilterIDs": "fltrID1",
		"Weights":   "wrong",
	}
	expErr := "invalid DynamicWeight format for string <wrong>"
	if _, err := NewRateProfileFromMapDataDBMap("cgrates.org", "ExID", mapRP, JSONMarshaler{}); err == nil || err.Error() != expErr {
		t.Errorf("Expected error <%v>, received <%v>", expErr, err)
	}
	mapRP = map[string]any{
		"FilterIDs": "fltrID1",
		"MinCost":   "wrong",
	}
	expErr = "can't convert <wrong> to decimal"
	if _, err := NewRateProfileFromMapDataDBMap("cgrates.org", "ExID", mapRP, JSONMarshaler{}); err == nil || err.Error() != expErr {
		t.Errorf("Expected error <%v>, received <%v>", expErr, err)
	}
	mapRP = map[string]any{
		"FilterIDs": "fltrID1",
		"MaxCost":   "wrong",
	}
	expErr = "can't convert <wrong> to decimal"
	if _, err := NewRateProfileFromMapDataDBMap("cgrates.org", "ExID", mapRP, JSONMarshaler{}); err == nil || err.Error() != expErr {
		t.Errorf("Expected error <%v>, received <%v>", expErr, err)
	}
	mapRP = map[string]any{
		"FilterIDs":  "fltrID1",
		"Rates:rat1": "{\"ID\":\"rat1\",\"FilterIDs\":null,\"ActivationTimes\":\"\",\"Weights\":null,\"Blocker\":false,\"IntervalRates\":null}\",\"Rates:rat2\":\"{\"ID\":\"rat2\",\"FilterIDs\":null,\"ActivationTimes\":\"\",\"Weights\":null,\"Blocker\":false,\"IntervalRates\":null}\",\"Weights\":\";10\"}",
	}
	expErr = "invalid character '\"' after top-level value"
	if _, err := NewRateProfileFromMapDataDBMap("cgrates.org", "ExID", mapRP, JSONMarshaler{}); err == nil || err.Error() != expErr {
		t.Errorf("Expected error <%v>, received <%v>", expErr, err)
	}
}
