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
	"fmt"
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
func TestAttrGetCdrsAsCDRsFilter(t *testing.T) {
	attrGetCdrs := &AttrGetCdrs{
		TimeStart: "2019-04-04T11:45:26.371Z",
		TimeEnd:   "2019-04-04T11:46:26.371Z",
		SkipRated: true,
		CgrIds:    []string{"CGRID"},
		TORs:      []string{MetaVoice},
		Accounts:  []string{"1001"},
		Subjects:  []string{"1001"},
	}
	eOut := &CDRsFilter{
		CGRIDs:              attrGetCdrs.CgrIds,
		RunIDs:              attrGetCdrs.MediationRunIds,
		ToRs:                attrGetCdrs.TORs,
		OriginHosts:         attrGetCdrs.CdrHosts,
		Sources:             attrGetCdrs.CdrSources,
		RequestTypes:        attrGetCdrs.ReqTypes,
		Tenants:             attrGetCdrs.Tenants,
		Categories:          attrGetCdrs.Categories,
		Accounts:            attrGetCdrs.Accounts,
		Subjects:            attrGetCdrs.Subjects,
		DestinationPrefixes: attrGetCdrs.DestinationPrefixes,
		OrderIDStart:        attrGetCdrs.OrderIdStart,
		OrderIDEnd:          attrGetCdrs.OrderIdEnd,
		Paginator:           attrGetCdrs.Paginator,
		OrderBy:             attrGetCdrs.OrderBy,
		MaxCost:             Float64Pointer(-1.0),
	}
	//chekck with an empty struct
	var testStruct *AttrGetCdrs
	rcv, err := testStruct.AsCDRsFilter("")
	if err != nil {
		t.Error(err)
	}
	if rcv != nil {
		t.Errorf("Nil struct expected")
	}
	//check with wrong time-zone
	_, err = attrGetCdrs.AsCDRsFilter("wrongtimezone")
	if err == nil {
		t.Errorf("ParseTimeDetectLayout error")
	}
	//check with wrong TimeStart
	attrGetCdrs.TimeStart = "wrongtimeStart"
	_, err = attrGetCdrs.AsCDRsFilter("")
	if err == nil {
		t.Errorf("Wrong AnswerTimeStart not processed")
	}
	//check with wrong TimeEnd
	attrGetCdrs.TimeStart = "2020-04-04T11:45:26.371Z"
	attrGetCdrs.TimeEnd = "wrongtimeEnd"
	_, err = attrGetCdrs.AsCDRsFilter("")
	if err == nil {
		t.Errorf("Wrong AnswerTimeEnd not processed")
	}

	//check with SkipRated = true & normal timeStar/timeEnd
	attrGetCdrs.TimeStart = "2020-04-04T11:45:26.371Z"
	attrGetCdrs.TimeEnd = "2020-04-04T11:46:26.371Z"
	TimeStart, _ := ParseTimeDetectLayout("2020-04-04T11:45:26.371Z", "")
	eOut.AnswerTimeStart = &TimeStart
	timeEnd, _ := ParseTimeDetectLayout("2020-04-04T11:46:26.371Z", "")
	eOut.AnswerTimeEnd = &timeEnd

	rcv, err = attrGetCdrs.AsCDRsFilter("")
	if err != nil {
		t.Errorf("ParseTimeDetectLayout error")
	}
	if !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expected: %s ,received: %s ", ToJSON(eOut), ToJSON(rcv))
	}

	//check with SkipRated = false
	attrGetCdrs.SkipRated = false
	attrGetCdrs.SkipErrors = true
	eOut.MinCost = Float64Pointer(0.0)
	rcv, err = attrGetCdrs.AsCDRsFilter("")
	if err != nil {
		t.Errorf("ParseTimeDetectLayout error")
	}
	if !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expected: %s ,received: %s ", ToJSON(eOut), ToJSON(rcv))
	}
}

func TestRPCCDRsFilterAsCDRsFilter(t *testing.T) {
	var testStruct *RPCCDRsFilter
	rcv, err := testStruct.AsCDRsFilter("")
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(rcv, new(CDRsFilter)) {
		t.Errorf("Empty struct expected")
	}
	rpcCDRsFilter := &RPCCDRsFilter{
		CGRIDs:                 []string{"CGRIDs"},
		NotCGRIDs:              []string{"NotCGRIDs"},
		RunIDs:                 []string{"RunIDs"},
		NotRunIDs:              []string{"NotRunIDs"},
		OriginIDs:              []string{"OriginIDs"},
		NotOriginIDs:           []string{"NotOriginIDs"},
		OriginHosts:            []string{"OriginHosts"},
		NotOriginHosts:         []string{"NotOriginHosts"},
		Sources:                []string{"Sources"},
		NotSources:             []string{"NotSources"},
		ToRs:                   []string{"ToRs"},
		NotToRs:                []string{"NotToRs"},
		RequestTypes:           []string{"RequestTypes"},
		NotRequestTypes:        []string{"NotRequestTypes"},
		Tenants:                []string{"Tenants"},
		NotTenants:             []string{"NotTenants"},
		Categories:             []string{"Categories"},
		NotCategories:          []string{"NotCategories"},
		Accounts:               []string{"Accounts"},
		NotAccounts:            []string{"NotAccounts"},
		Subjects:               []string{"Subjects"},
		NotSubjects:            []string{"NotSubjects"},
		DestinationPrefixes:    []string{"DestinationPrefixes"},
		NotDestinationPrefixes: []string{"NotDestinationPrefixes"},
		Costs:                  []float64{0.1, 0.2},
		NotCosts:               []float64{0.3, 0.4},
		ExtraFields:            map[string]string{},
		NotExtraFields:         map[string]string{},
		SetupTimeStart:         "2020-04-18T11:46:26.371Z",
		SetupTimeEnd:           "2020-04-18T11:46:26.371Z",
		AnswerTimeStart:        "2020-04-18T11:46:26.371Z",
		AnswerTimeEnd:          "2020-04-18T11:46:26.371Z",
		CreatedAtStart:         "2020-04-18T11:46:26.371Z",
		CreatedAtEnd:           "2020-04-18T11:46:26.371Z",
		UpdatedAtStart:         "2020-04-18T11:46:26.371Z",
		UpdatedAtEnd:           "2020-04-18T11:46:26.371Z",
		MinUsage:               "MinUsage",
		MaxUsage:               "MaxUsage",
		OrderBy:                "OrderBy",
		ExtraArgs: map[string]interface{}{
			OrderIDStart: 0,
			OrderIDEnd:   0,
			MinCost:      0.,
			MaxCost:      0.,
		},
	}
	eOut := &CDRsFilter{
		CGRIDs:                 rpcCDRsFilter.CGRIDs,
		NotCGRIDs:              rpcCDRsFilter.NotCGRIDs,
		RunIDs:                 rpcCDRsFilter.RunIDs,
		NotRunIDs:              rpcCDRsFilter.NotRunIDs,
		OriginIDs:              rpcCDRsFilter.OriginIDs,
		NotOriginIDs:           rpcCDRsFilter.NotOriginIDs,
		ToRs:                   rpcCDRsFilter.ToRs,
		NotToRs:                rpcCDRsFilter.NotToRs,
		OriginHosts:            rpcCDRsFilter.OriginHosts,
		NotOriginHosts:         rpcCDRsFilter.NotOriginHosts,
		Sources:                rpcCDRsFilter.Sources,
		NotSources:             rpcCDRsFilter.NotSources,
		RequestTypes:           rpcCDRsFilter.RequestTypes,
		NotRequestTypes:        rpcCDRsFilter.NotRequestTypes,
		Tenants:                rpcCDRsFilter.Tenants,
		NotTenants:             rpcCDRsFilter.NotTenants,
		Categories:             rpcCDRsFilter.Categories,
		NotCategories:          rpcCDRsFilter.NotCategories,
		Accounts:               rpcCDRsFilter.Accounts,
		NotAccounts:            rpcCDRsFilter.NotAccounts,
		Subjects:               rpcCDRsFilter.Subjects,
		NotSubjects:            rpcCDRsFilter.NotSubjects,
		DestinationPrefixes:    rpcCDRsFilter.DestinationPrefixes,
		NotDestinationPrefixes: rpcCDRsFilter.NotDestinationPrefixes,
		Costs:                  rpcCDRsFilter.Costs,
		NotCosts:               rpcCDRsFilter.NotCosts,
		ExtraFields:            rpcCDRsFilter.ExtraFields,
		NotExtraFields:         rpcCDRsFilter.NotExtraFields,
		OrderIDStart:           Int64Pointer(0),
		OrderIDEnd:             Int64Pointer(0),
		MinUsage:               rpcCDRsFilter.MinUsage,
		MaxUsage:               rpcCDRsFilter.MaxUsage,
		MinCost:                Float64Pointer(0.),
		MaxCost:                Float64Pointer(0.),
		Paginator:              rpcCDRsFilter.Paginator,
		OrderBy:                rpcCDRsFilter.OrderBy,
	}
	tTime, _ := ParseTimeDetectLayout("2020-04-18T11:46:26.371Z", "")
	eOut.AnswerTimeEnd = &tTime
	eOut.UpdatedAtEnd = &tTime
	eOut.UpdatedAtStart = &tTime
	eOut.CreatedAtEnd = &tTime
	eOut.CreatedAtStart = &tTime
	eOut.AnswerTimeEnd = &tTime
	eOut.AnswerTimeStart = &tTime
	eOut.SetupTimeEnd = &tTime
	eOut.SetupTimeStart = &tTime

	if rcv, err = rpcCDRsFilter.AsCDRsFilter(""); err != nil {
		t.Errorf("ParseTimeDetectLayout error")
	} else if !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expected: %s ,received: %s ", ToJSON(eOut), ToJSON(rcv))
	}

	rpcCDRsFilter.ExtraArgs[MaxCost] = "notFloat64"
	if _, err = rpcCDRsFilter.AsCDRsFilter(""); err == nil {
		t.Errorf("MaxCost should not be processed")
	}

	rpcCDRsFilter.ExtraArgs[MinCost] = "notFloat64"
	if _, err = rpcCDRsFilter.AsCDRsFilter(""); err == nil {
		t.Errorf("MinCost should not be processed")
	}
	rpcCDRsFilter.ExtraArgs[OrderIDEnd] = "notInt64"
	if _, err = rpcCDRsFilter.AsCDRsFilter(""); err == nil {
		t.Errorf("OrderIDEnd should not be processed")
	}
	rpcCDRsFilter.ExtraArgs[OrderIDStart] = "notInt64"
	if _, err = rpcCDRsFilter.AsCDRsFilter(""); err == nil {
		t.Errorf("OrderIDStart should not be processed")
	}

	rpcCDRsFilter.UpdatedAtEnd = "wrongUpdatedAtEnd"
	if _, err = rpcCDRsFilter.AsCDRsFilter(""); err == nil {
		t.Errorf("UpdatedAtEnd should not be processed")
	}
	rpcCDRsFilter.UpdatedAtStart = "wrongUpdatedAtStart"
	if _, err = rpcCDRsFilter.AsCDRsFilter(""); err == nil {
		t.Errorf("UpdatedAtStart should not be processed")
	}
	rpcCDRsFilter.CreatedAtEnd = "wrongCreatedAtEnd"
	if _, err = rpcCDRsFilter.AsCDRsFilter(""); err == nil {
		t.Errorf("CreatedAtEnd should not be processed")
	}
	rpcCDRsFilter.CreatedAtStart = "wrongCreatedAtStart"
	if _, err = rpcCDRsFilter.AsCDRsFilter(""); err == nil {
		t.Errorf("CreatedAtStart should not be processed")
	}
	rpcCDRsFilter.AnswerTimeEnd = "wrongAnswerTimeEnd"
	if _, err = rpcCDRsFilter.AsCDRsFilter(""); err == nil {
		t.Errorf("AnswerTimeEnd should not be processed")
	}
	rpcCDRsFilter.AnswerTimeStart = "wrongAnswerTimeStart"
	if _, err = rpcCDRsFilter.AsCDRsFilter(""); err == nil {
		t.Errorf("AnswerTimeStart should not be processed")
	}
	rpcCDRsFilter.SetupTimeEnd = "wrongSetupTimeEnd"
	if _, err = rpcCDRsFilter.AsCDRsFilter(""); err == nil {
		t.Errorf("SetupTimeEnd should not be processed")
	}
	rpcCDRsFilter.SetupTimeStart = "wrongSetupTimeStart"
	if _, err = rpcCDRsFilter.AsCDRsFilter(""); err == nil {
		t.Errorf("SetupTimeStart should not be processed")
	}
}

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

func TestCDRsFilterPrepare(t *testing.T) {
	fltr := &CDRsFilter{
		CGRIDs:          []string{"5", "6", "1", "3", "2", "4"},
		NotCGRIDs:       []string{"5", "6", "1", "3", "2", "4"},
		RunIDs:          []string{"5", "6", "1", "3", "2", "4"},
		NotRunIDs:       []string{"5", "6", "1", "3", "2", "4"},
		OriginIDs:       []string{"5", "6", "1", "3", "2", "4"},
		NotOriginIDs:    []string{"5", "6", "1", "3", "2", "4"},
		OriginHosts:     []string{"5", "6", "1", "3", "2", "4"},
		NotOriginHosts:  []string{"5", "6", "1", "3", "2", "4"},
		Sources:         []string{"5", "6", "1", "3", "2", "4"},
		NotSources:      []string{"5", "6", "1", "3", "2", "4"},
		ToRs:            []string{"5", "6", "1", "3", "2", "4"},
		NotToRs:         []string{"5", "6", "1", "3", "2", "4"},
		RequestTypes:    []string{"5", "6", "1", "3", "2", "4"},
		NotRequestTypes: []string{"5", "6", "1", "3", "2", "4"},
		Tenants:         []string{"5", "6", "1", "3", "2", "4"},
		NotTenants:      []string{"5", "6", "1", "3", "2", "4"},
		Categories:      []string{"5", "6", "1", "3", "2", "4"},
		NotCategories:   []string{"5", "6", "1", "3", "2", "4"},
		Accounts:        []string{"5", "6", "1", "3", "2", "4"},
		NotAccounts:     []string{"5", "6", "1", "3", "2", "4"},
		Subjects:        []string{"5", "6", "1", "3", "2", "4"},
		NotSubjects:     []string{"5", "6", "1", "3", "2", "4"},
		Costs:           []float64{5, 6, 1, 3, 2, 4},
		NotCosts:        []float64{5, 6, 1, 3, 2, 4},
	}
	exp := &CDRsFilter{
		CGRIDs:          []string{"1", "2", "3", "4", "5", "6"},
		NotCGRIDs:       []string{"1", "2", "3", "4", "5", "6"},
		RunIDs:          []string{"1", "2", "3", "4", "5", "6"},
		NotRunIDs:       []string{"1", "2", "3", "4", "5", "6"},
		OriginIDs:       []string{"1", "2", "3", "4", "5", "6"},
		NotOriginIDs:    []string{"1", "2", "3", "4", "5", "6"},
		OriginHosts:     []string{"1", "2", "3", "4", "5", "6"},
		NotOriginHosts:  []string{"1", "2", "3", "4", "5", "6"},
		Sources:         []string{"1", "2", "3", "4", "5", "6"},
		NotSources:      []string{"1", "2", "3", "4", "5", "6"},
		ToRs:            []string{"1", "2", "3", "4", "5", "6"},
		NotToRs:         []string{"1", "2", "3", "4", "5", "6"},
		RequestTypes:    []string{"1", "2", "3", "4", "5", "6"},
		NotRequestTypes: []string{"1", "2", "3", "4", "5", "6"},
		Tenants:         []string{"1", "2", "3", "4", "5", "6"},
		NotTenants:      []string{"1", "2", "3", "4", "5", "6"},
		Categories:      []string{"1", "2", "3", "4", "5", "6"},
		NotCategories:   []string{"1", "2", "3", "4", "5", "6"},
		Accounts:        []string{"1", "2", "3", "4", "5", "6"},
		NotAccounts:     []string{"1", "2", "3", "4", "5", "6"},
		Subjects:        []string{"1", "2", "3", "4", "5", "6"},
		NotSubjects:     []string{"1", "2", "3", "4", "5", "6"},
		Costs:           []float64{1, 2, 3, 4, 5, 6},
		NotCosts:        []float64{1, 2, 3, 4, 5, 6},
	}
	if fltr.Prepare(); !reflect.DeepEqual(exp, fltr) {
		t.Errorf("Expected %s,received %s", ToJSON(exp), ToJSON(fltr))
	}
}

func TestNewAttrReloadCacheWithOpts(t *testing.T) {
	newAttrReloadCache := &AttrReloadCacheWithAPIOpts{
		ResourceProfileIDs:           []string{MetaAny},
		ResourceIDs:                  []string{MetaAny},
		StatsQueueIDs:                []string{MetaAny},
		StatsQueueProfileIDs:         []string{MetaAny},
		ThresholdIDs:                 []string{MetaAny},
		ThresholdProfileIDs:          []string{MetaAny},
		FilterIDs:                    []string{MetaAny},
		RouteProfileIDs:              []string{MetaAny},
		AttributeProfileIDs:          []string{MetaAny},
		ChargerProfileIDs:            []string{MetaAny},
		DispatcherProfileIDs:         []string{MetaAny},
		DispatcherHostIDs:            []string{MetaAny},
		RateProfileIDs:               []string{MetaAny},
		AttributeFilterIndexIDs:      []string{MetaAny},
		ResourceFilterIndexIDs:       []string{MetaAny},
		StatFilterIndexIDs:           []string{MetaAny},
		ThresholdFilterIndexIDs:      []string{MetaAny},
		RouteFilterIndexIDs:          []string{MetaAny},
		ChargerFilterIndexIDs:        []string{MetaAny},
		DispatcherFilterIndexIDs:     []string{MetaAny},
		RateProfilesFilterIndexIDs:   []string{MetaAny},
		RateFilterIndexIDs:           []string{MetaAny},
		FilterIndexIDs:               []string{MetaAny},
		ActionProfileIDs:             []string{MetaAny},
		AccountIDs:                   []string{MetaAny},
		ActionProfilesFilterIndexIDs: []string{MetaAny},
		AccountsFilterIndexIDs:       []string{MetaAny},
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

func TestAPITPDataPaginate(t *testing.T) {
	var in []string

	if rcv, err := Paginate(in, 2, 2, 6); err != nil {
		t.Error(err)
	} else if rcv != nil {
		t.Error("expected nil return")
	}

	in = []string{"FLTR_1", "FLTR_2", "FLTR_3", "FLTR_4", "FLTR_5", "FLTR_6", "FLTR_7",
		"FLTR_8", "FLTR_9", "FLTR_10", "FLTR_11", "FLTR_12", "FLTR_13", "FLTR_14",
		"FLTR_15", "FLTR_16", "FLTR_17", "FLTR_18", "FLTR_19", "FLTR_20"}

	exp := []string{"FLTR_7", "FLTR_8"}
	if rcv, err := Paginate(in, 2, 6, 9); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", exp, rcv)
	}

	exp = []string{"FLTR_19", "FLTR_20"}
	if rcv, err := Paginate(in, 0, 18, 50); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", exp, rcv)
	}

	if rcv, err := Paginate(in, 0, 0, 50); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, in) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", in, rcv)
	}

	experr := `SERVER_ERROR: maximum number of items exceeded`
	if _, err := Paginate(in, 0, 0, 19); err == nil || err.Error() != experr {
		t.Error(err)
	}

	if rcv, err := Paginate(in, 25, 18, 50); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", exp, rcv)
	}

	var expOut []string
	if rcv, err := Paginate(in, 2, 22, 50); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, expOut) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", expOut, rcv)
	}

	if _, err := Paginate(in, 2, 4, 5); err == nil || err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	if _, err := Paginate(in, 0, 18, 19); err == nil || err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}

}

type pag struct {
	limit    int
	offset   int
	maxItems int
}

func testName(p pag) string {
	return fmt.Sprintf("limit:<%d>, offset:<%d>, maxItems:<%d>", p.limit, p.offset, p.maxItems)
}

func TestPagination(t *testing.T) {
	in := []string{"FLTR_1", "FLTR_2", "FLTR_3", "FLTR_4", "FLTR_5"}
	experr := "SERVER_ERROR: maximum number of items exceeded"
	cases := []struct {
		p   pag
		err string
	}{
		{pag{limit: 0, offset: 0, maxItems: 1}, experr},
		{pag{limit: 1, offset: 0, maxItems: 1}, ""},
		{pag{limit: 0, offset: 1, maxItems: 1}, experr},
		{pag{limit: 0, offset: 0, maxItems: 2}, experr},
		{pag{limit: 0, offset: 1, maxItems: 2}, experr},
		{pag{limit: 0, offset: 2, maxItems: 2}, experr},
		{pag{limit: 1, offset: 0, maxItems: 2}, ""},
		{pag{limit: 1, offset: 1, maxItems: 2}, ""},
		{pag{limit: 2, offset: 0, maxItems: 2}, ""},
		{pag{limit: 0, offset: 0, maxItems: 3}, experr},
		{pag{limit: 0, offset: 1, maxItems: 3}, experr},
		{pag{limit: 0, offset: 2, maxItems: 3}, experr},
		{pag{limit: 0, offset: 3, maxItems: 3}, experr},
		{pag{limit: 1, offset: 0, maxItems: 3}, ""},
		{pag{limit: 1, offset: 1, maxItems: 3}, ""},
		{pag{limit: 1, offset: 2, maxItems: 3}, ""},
		{pag{limit: 2, offset: 0, maxItems: 3}, ""},
		{pag{limit: 2, offset: 1, maxItems: 3}, ""},
		{pag{limit: 3, offset: 0, maxItems: 3}, ""},
		{pag{limit: 0, offset: 0, maxItems: 4}, experr},
		{pag{limit: 0, offset: 1, maxItems: 4}, experr},
		{pag{limit: 0, offset: 2, maxItems: 4}, experr},
		{pag{limit: 0, offset: 3, maxItems: 4}, experr},
		{pag{limit: 0, offset: 4, maxItems: 4}, experr},
		{pag{limit: 1, offset: 0, maxItems: 4}, ""},
		{pag{limit: 1, offset: 1, maxItems: 4}, ""},
		{pag{limit: 1, offset: 2, maxItems: 4}, ""},
		{pag{limit: 1, offset: 3, maxItems: 4}, ""},
		{pag{limit: 2, offset: 0, maxItems: 4}, ""},
		{pag{limit: 2, offset: 1, maxItems: 4}, ""},
		{pag{limit: 2, offset: 2, maxItems: 4}, ""},
		{pag{limit: 3, offset: 0, maxItems: 4}, ""},
		{pag{limit: 3, offset: 1, maxItems: 4}, ""},
		{pag{limit: 4, offset: 0, maxItems: 4}, ""},
		{pag{limit: 0, offset: 0, maxItems: 5}, ""},
		{pag{limit: 0, offset: 1, maxItems: 5}, ""},
		{pag{limit: 0, offset: 2, maxItems: 5}, ""},
		{pag{limit: 0, offset: 3, maxItems: 5}, ""},
		{pag{limit: 0, offset: 4, maxItems: 5}, ""},
		{pag{limit: 0, offset: 5, maxItems: 5}, ""},
		{pag{limit: 1, offset: 0, maxItems: 5}, ""},
		{pag{limit: 1, offset: 1, maxItems: 5}, ""},
		{pag{limit: 1, offset: 2, maxItems: 5}, ""},
		{pag{limit: 1, offset: 3, maxItems: 5}, ""},
		{pag{limit: 1, offset: 4, maxItems: 5}, ""},
		{pag{limit: 2, offset: 0, maxItems: 5}, ""},
		{pag{limit: 2, offset: 1, maxItems: 5}, ""},
		{pag{limit: 2, offset: 2, maxItems: 5}, ""},
		{pag{limit: 2, offset: 3, maxItems: 5}, ""},
		{pag{limit: 3, offset: 0, maxItems: 5}, ""},
		{pag{limit: 3, offset: 1, maxItems: 5}, ""},
		{pag{limit: 3, offset: 2, maxItems: 5}, ""},
		{pag{limit: 4, offset: 0, maxItems: 5}, ""},
		{pag{limit: 4, offset: 1, maxItems: 5}, ""},
		{pag{limit: 5, offset: 0, maxItems: 5}, ""},
	}

	for _, c := range cases {
		t.Run(testName(c.p), func(t *testing.T) {
			_, err := Paginate(in, c.p.limit, c.p.offset, c.p.maxItems)
			if err != nil {
				if c.err == "" {
					t.Error("did not expect error")
				}
			} else if c.err != "" {
				t.Errorf("expected error")
			}
		})
	}
}

func TestAPITPDataGetPaginateOpts(t *testing.T) {
	opts := map[string]interface{}{
		PageLimitOpt:    1.3,
		PageOffsetOpt:   4,
		PageMaxItemsOpt: "5",
	}

	if limit, offset, maxItems, err := GetPaginateOpts(opts); err != nil {
		t.Error(err)
	} else if limit != 1 {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", 1, limit)
	} else if offset != 4 {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", 4, offset)
	} else if maxItems != 5 {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", 5, maxItems)
	}

	opts[PageMaxItemsOpt] = false
	experr := `cannot convert field<bool>: false to int`
	if _, _, _, err := GetPaginateOpts(opts); err == nil || err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	opts[PageOffsetOpt] = struct{}{}
	experr = `cannot convert field<struct {}>: {} to int`
	if _, _, _, err := GetPaginateOpts(opts); err == nil || err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	opts[PageLimitOpt] = true
	experr = `cannot convert field<bool>: true to int`
	if _, _, _, err := GetPaginateOpts(opts); err == nil || err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}
}
