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
)

func TestClonePaginator(t *testing.T) {
	expectedPaginator := Paginator{
		Limit:  IntPointer(2),
		Offset: IntPointer(0),
	}
	clonedPaginator := expectedPaginator.Clone()
	if !reflect.DeepEqual(expectedPaginator, clonedPaginator) {
		t.Errorf("Expected %+v, received %+v", ToJSON(expectedPaginator), ToJSON(clonedPaginator))
	}
}

//now working here

func TestTPActivationIntervalAsActivationInterval(t *testing.T) {
	tPActivationInterval := &TPActivationInterval{
		ActivationTime: "2019-04-04T11:45:26.371Z",
		ExpiryTime:     "2019-04-04T11:46:26.371Z",
	}
	eOut := new(ActivationInterval)

	tTime, _ := ParseTimeDetectLayout("2019-04-04T11:45:26.371Z", "")
	eOut.ActivationTime = tTime
	tTime, _ = ParseTimeDetectLayout("2019-04-04T11:46:26.371Z", "")
	eOut.ExpiryTime = tTime

	rcv, err := tPActivationInterval.AsActivationInterval("")
	if err != nil {
		t.Errorf("ParseTimeDetectLayout error")
	}
	if !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expected: %s ,received: %s ", ToJSON(eOut), ToJSON(rcv))
	}
	//check with wrong time
	tPActivationInterval.ExpiryTime = "wrongExpiryTime"
	_, err = tPActivationInterval.AsActivationInterval("")
	if err == nil {
		t.Errorf("Wrong ExpiryTime not processed")
	}
	tPActivationInterval.ActivationTime = "wrongActivationTime"
	_, err = tPActivationInterval.AsActivationInterval("")
	if err == nil {
		t.Errorf("Wrong ActivationTimes not processed")
	}
}

func TestActivationIntervalIsActiveAtTime(t *testing.T) {
	activationInterval := new(ActivationInterval)

	//case ActivationTimes = Expiry = 0001-01-01 00:00:00 +0000 UTC
	activationInterval.ActivationTime = time.Time{}
	activationInterval.ExpiryTime = time.Time{}
	rcv := activationInterval.IsActiveAtTime(time.Time{})
	if !rcv {
		t.Errorf("ActivationTimes = Expiry = time.Time{}, expecting 0 ")
	}
	activationInterval.ActivationTime = time.Date(2018, time.April, 18, 23, 0, 0, 0, time.UTC)
	activationInterval.ExpiryTime = time.Date(2020, time.April, 18, 23, 0, 0, 0, time.UTC)

	//atTime < ActivationTimes
	atTime := time.Date(2017, time.April, 18, 23, 0, 0, 0, time.UTC)
	rcv = activationInterval.IsActiveAtTime(atTime)
	if rcv {
		t.Errorf("atTime < ActivationTimes, expecting 0 ")
	}

	//atTime > ExpiryTime
	atTime = time.Date(2021, time.April, 18, 23, 0, 0, 0, time.UTC) //tTime
	rcv = activationInterval.IsActiveAtTime(atTime)
	if rcv {
		t.Errorf("atTime > Expiry, expecting 0 ")
	}

	//ideal case
	atTime = time.Date(2019, time.April, 18, 23, 0, 0, 0, time.UTC) //tTime
	rcv = activationInterval.IsActiveAtTime(atTime)
	if !rcv {
		t.Errorf("ActivationTimes < atTime < ExpiryTime. Expecting 1 ")
	}
	//ActivationTimes > ExpiryTime
	activationInterval.ActivationTime = time.Date(2020, time.April, 18, 23, 0, 0, 0, time.UTC) //tTime
	activationInterval.ExpiryTime = time.Date(2018, time.April, 18, 23, 0, 0, 0, time.UTC)     //tTime
	rcv = activationInterval.IsActiveAtTime(atTime)
	if rcv {
		t.Errorf("ActivationTimes > ExpiryTime. Expecting 0 ")
	}
}

func TestNewAttrReloadCacheWithOpts(t *testing.T) {
	newAttrReloadCache := &AttrReloadCacheWithAPIOpts{
		ResourceProfileIDs:           []string{MetaAny},
		ResourceIDs:                  []string{MetaAny},
		IPProfileIDs:                 []string{MetaAny},
		IPIDs:                        []string{MetaAny},
		StatsQueueIDs:                []string{MetaAny},
		StatsQueueProfileIDs:         []string{MetaAny},
		ThresholdIDs:                 []string{MetaAny},
		ThresholdProfileIDs:          []string{MetaAny},
		TrendIDs:                     []string{MetaAny},
		TrendProfileIDs:              []string{MetaAny},
		FilterIDs:                    []string{MetaAny},
		RouteProfileIDs:              []string{MetaAny},
		AttributeProfileIDs:          []string{MetaAny},
		ChargerProfileIDs:            []string{MetaAny},
		RateProfileIDs:               []string{MetaAny},
		AttributeFilterIndexIDs:      []string{MetaAny},
		ResourceFilterIndexIDs:       []string{MetaAny},
		IPFilterIndexIDs:             []string{MetaAny},
		StatFilterIndexIDs:           []string{MetaAny},
		ThresholdFilterIndexIDs:      []string{MetaAny},
		RouteFilterIndexIDs:          []string{MetaAny},
		ChargerFilterIndexIDs:        []string{MetaAny},
		RateProfilesFilterIndexIDs:   []string{MetaAny},
		RateFilterIndexIDs:           []string{MetaAny},
		FilterIndexIDs:               []string{MetaAny},
		ActionProfileIDs:             []string{MetaAny},
		AccountIDs:                   []string{MetaAny},
		ActionProfilesFilterIndexIDs: []string{MetaAny},
		AccountsFilterIndexIDs:       []string{MetaAny},
		RankingProfileIDs:            []string{MetaAny},
		RankingIDs:                   []string{MetaAny},
	}
	eMap := NewAttrReloadCacheWithOpts()
	if !reflect.DeepEqual(eMap, newAttrReloadCache) {
		t.Errorf("Expected %+v \n, received %+v", eMap, newAttrReloadCache)
	}
}

func TestNewTPBalanceCostIncrement(t *testing.T) {
	incrementStr := "20"
	fixedFeeStr := "10"
	recurentFeeStr := "0.4"
	filterStr := "*string:*Account:1001"
	expected := &TPBalanceCostIncrement{
		FilterIDs:    []string{"*string:*Account:1001"},
		Increment:    "20",
		FixedFee:     Float64Pointer(10),
		RecurrentFee: Float64Pointer(0.4),
	}
	if rcv, err := NewTPBalanceCostIncrement(filterStr, incrementStr, fixedFeeStr, recurentFeeStr); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Expected %+v, received %+v", expected, rcv)
	}
}

func TestNewTPBalanceCostIncrementErrors(t *testing.T) {
	invalidStr := "not_float64"
	expectedErr := "strconv.ParseFloat: parsing \"not_float64\": invalid syntax"
	if _, err := NewTPBalanceCostIncrement(EmptyString, EmptyString, invalidStr, EmptyString); err == nil || err.Error() != expectedErr {
		t.Errorf("Expected %+v, received %+v", expectedErr, err)
	}
	if _, err := NewTPBalanceCostIncrement(EmptyString, EmptyString, EmptyString, invalidStr); err == nil || err.Error() != expectedErr {
		t.Errorf("Expected %+v, received %+v", expectedErr, err)
	}
}

func TestTPBalanceCostIncrementAsString(t *testing.T) {
	costIncr := &TPBalanceCostIncrement{
		FilterIDs:    []string{"*string:*Account:1001"},
		Increment:    "20",
		FixedFee:     Float64Pointer(10),
		RecurrentFee: Float64Pointer(0.4),
	}
	expected := "*string:*Account:1001;20;10;0.4"
	if rcv := costIncr.AsString(); expected != rcv {
		t.Errorf("Expected %+v, received %+v", expected, rcv)
	}
}

func TestNewBalanceUnitFactor(t *testing.T) {
	factorStr := "1.7"
	filterStr := "*string:*Account:1001"
	expected := &TPBalanceUnitFactor{
		FilterIDs: []string{"*string:*Account:1001"},
		Factor:    1.7,
	}
	if rcv, err := NewTPBalanceUnitFactor(filterStr, factorStr); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rcv) {
		t.Errorf("Expected %+v, received %+v", expected, rcv)
	}
}

func TestNewBalanceUnitFactorError(t *testing.T) {
	invalidStr := "not_float64"
	expectedErr := "strconv.ParseFloat: parsing \"not_float64\": invalid syntax"
	if _, err := NewTPBalanceUnitFactor(EmptyString, invalidStr); err == nil || err.Error() != expectedErr {
		t.Errorf("Expected %+v, received %+v", expectedErr, err)
	}
}

func TestBalanceUnitFactor(t *testing.T) {
	unitFctr := &TPBalanceUnitFactor{
		FilterIDs: []string{"*string:*Account:1001"},
		Factor:    1.7,
	}
	expected := "*string:*Account:1001;1.7"
	if rcv := unitFctr.AsString(); expected != rcv {
		t.Errorf("Expected %+v, received %+v", expected, rcv)
	}
}

func TestActivationIntervalEquals(t *testing.T) {
	aI := &ActivationInterval{
		ActivationTime: time.Time{},
		ExpiryTime:     time.Date(2021, 5, 13, 0, 0, 0, 0, time.UTC),
	}

	actInt := &ActivationInterval{
		ActivationTime: time.Date(2021, 5, 13, 0, 0, 0, 0, time.UTC),
		ExpiryTime:     time.Date(2021, 5, 13, 0, 0, 0, 0, time.UTC),
	}

	if aI.Equals(actInt) {
		t.Error("ActivationInervals should not match")
	}

	aI.ActivationTime = time.Time{}
	actInt.ActivationTime = time.Time{}
	if !aI.Equals(actInt) {
		t.Error("Expected both activation interval to be equal")
	}
}

func TestNewAttrReloadCacheWithOptsFromMap(t *testing.T) {
	excluded := NewStringSet([]string{MetaAPIBan, MetaLoadIDs})
	mp := make(map[string][]string)
	for k := range CacheInstanceToPrefix {
		if !excluded.Has(k) {
			mp[k] = []string{MetaAny}
		}
	}

	exp := NewAttrReloadCacheWithOpts()
	rply := NewAttrReloadCacheWithOptsFromMap(mp, "", nil)
	if !reflect.DeepEqual(exp, rply) {
		t.Errorf("Expected %+v \n, received %+v", ToJSON(exp), ToJSON(rply))
	}
	rplyM := rply.Map()
	if !reflect.DeepEqual(mp, rplyM) {
		t.Errorf("Expected %+v \n, received %+v", ToJSON(mp), ToJSON(rplyM))
	}
}
