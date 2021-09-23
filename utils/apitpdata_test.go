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

	"github.com/ericlagergren/decimal"
)

func TestPaginatorPaginateStringSlice(t *testing.T) {
	//len(in)=0
	eOut := []string{}
	pgnt := new(Paginator)
	rcv := pgnt.PaginateStringSlice(eOut)
	if len(rcv) != 0 {
		t.Errorf("Expecting an empty slice, received: %+v", rcv)
	}
	//offset > len(in)
	eOut = []string{"1"}
	pgnt = &Paginator{Offset: IntPointer(2), Limit: IntPointer(0)}
	rcv = pgnt.PaginateStringSlice(eOut)

	if len(rcv) != 0 {
		t.Errorf("Expecting an empty slice, received: %+v", rcv)
	}
	//offset != 0 && limit != 0
	eOut = []string{"3", "4"}
	pgnt = &Paginator{Offset: IntPointer(2), Limit: IntPointer(0)}
	rcv = pgnt.PaginateStringSlice([]string{"1", "2", "3", "4"})

	if !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting an empty slice, received: %+v", rcv)
	}

	eOut = []string{"1", "2", "3", "4"}
	pgnt = new(Paginator)
	if rcv := pgnt.PaginateStringSlice([]string{"1", "2", "3", "4"}); !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", eOut, rcv)
	}
	eOut = []string{"1", "2", "3"}
	pgnt.Limit = IntPointer(3)
	if rcv := pgnt.PaginateStringSlice([]string{"1", "2", "3", "4"}); !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", eOut, rcv)
	}
	eOut = []string{"2", "3", "4"}
	pgnt.Offset = IntPointer(1)
	if rcv := pgnt.PaginateStringSlice([]string{"1", "2", "3", "4"}); !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", eOut, rcv)
	}
	eOut = []string{}
	pgnt.Offset = IntPointer(4)
	if rcv := pgnt.PaginateStringSlice([]string{"1", "2", "3", "4"}); !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", eOut, rcv)
	}
	eOut = []string{"3"}
	pgnt.Offset = IntPointer(2)
	pgnt.Limit = IntPointer(1)
	if rcv := pgnt.PaginateStringSlice([]string{"1", "2", "3", "4"}); !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", eOut, rcv)
	}
}

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

func TestStartTimeNow(t *testing.T) {
	ev := &CGREvent{
		Tenant: "*req.CGRID",
		ID:     "",
		Event:  map[string]interface{}{},
		APIOpts: map[string]interface{}{
			OptsRatesRateProfileIDs: []string{"123", "456", "789"},
		},
	}
	timpulet1 := time.Now()
	result, err := ev.StartTime(MetaNow, "")
	timpulet2 := time.Now()
	if err != nil {
		t.Errorf("Expected <nil> , received <%+v>", err)
	}
	if result.Before(timpulet1) && result.After(timpulet2) {
		t.Errorf("Expected between <%+v> and <%+v>, received <%+v>", timpulet1, timpulet2, result)
	}
}

func TestStartTime(t *testing.T) {
	ev := &CGREvent{
		Tenant: "*req.CGRID",
		ID:     "",
		Event:  map[string]interface{}{},
		APIOpts: map[string]interface{}{
			"*ratesStartTime":       "2018-01-07T17:00:10Z",
			OptsRatesRateProfileIDs: []string{"123", "456", "789"},
		},
	}
	if result, err := ev.StartTime(MetaNow, ""); err != nil {
		t.Errorf("Expected <nil> , received <%+v>", err)
	} else if !reflect.DeepEqual(result.String(), "2018-01-07 17:00:10 +0000 UTC") {
		t.Errorf("Expected <2018-01-07 17:00:10 +0000 UTC> , received <%+v>", result)
	}
}

func TestStartTime2(t *testing.T) {
	ev := &CGREvent{
		Tenant:  "cgrates.org",
		ID:      "TestEvent",
		Event:   map[string]interface{}{},
		APIOpts: map[string]interface{}{},
	}
	if result, err := ev.StartTime("2018-01-07T17:00:10Z", ""); err != nil {
		t.Errorf("Expected <nil> , received <%+v>", err)
	} else if !reflect.DeepEqual(result.String(), "2018-01-07 17:00:10 +0000 UTC") {
		t.Errorf("Expected <2018-01-07 17:00:10 +0000 UTC> , received <%+v>", result)
	}
}

func TestStartTimeError(t *testing.T) {
	ev := &CGREvent{
		Tenant: "*req.CGRID",
		ID:     "",
		Event:  map[string]interface{}{},
		APIOpts: map[string]interface{}{
			"*ratesStartTime":       "start",
			OptsRatesRateProfileIDs: []string{"123", "456", "789"},
		},
	}
	_, err := ev.StartTime(MetaNow, "")
	if err == nil && err.Error() != "received <Unsupported time format" {
		t.Errorf("Expected <nil> , received <%+v>", err)
	}
}

func TestUsageMinute(t *testing.T) {
	ev := &CGREvent{
		Tenant: "*req.CGRID",
		ID:     "",
		Event:  map[string]interface{}{},
		APIOpts: map[string]interface{}{
			OptsRatesRateProfileIDs: []string{"123", "456", "789"},
		},
	}
	if rcv, err := ev.OptsAsDecimal(decimal.New(int64(60*time.Second), 0), OptsRatesUsage, MetaUsage); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(decimal.New(int64(time.Minute), 0), rcv) {
		t.Errorf("Expected %+v, received %+v", decimal.New(int64(time.Minute), 0), rcv)
	}
}

func TestUsageError(t *testing.T) {
	ev := &CGREvent{
		Tenant: "*req.CGRID",
		ID:     "",
		Event:  map[string]interface{}{},
		APIOpts: map[string]interface{}{
			"*ratesUsage":           "start",
			OptsRatesRateProfileIDs: []string{"123", "456", "789"},
		},
	}
	_, err := ev.OptsAsDecimal(decimal.New(int64(time.Minute), 0), OptsRatesUsage, MetaUsage)
	if err == nil && err.Error() != "received <Unsupported time format" {
		t.Errorf("Expected <nil> , received <%+v>", err)
	}
}

func TestUsage(t *testing.T) {
	ev := &CGREvent{
		Tenant: "*req.CGRID",
		ID:     "",
		Event:  map[string]interface{}{},
		APIOpts: map[string]interface{}{
			"*ratesUsage":           "2m10s",
			OptsRatesRateProfileIDs: []string{"123", "456", "789"},
		},
	}

	if result, err := ev.OptsAsDecimal(decimal.New(int64(time.Minute), 0), OptsRatesUsage, MetaUsage); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(result.String(), "130000000000") {
		t.Errorf("Expected <130000000000> , received <%+v>", result.String())
	}
}

func TestNewTPBalanceCostIncrement(t *testing.T) {
	incrementStr := "20"
	fixedFeeStr := "10"
	recurentFeeStr := "0.4"
	filterStr := "*string:*Account:1001"
	expected := &TPBalanceCostIncrement{
		FilterIDs:    []string{"*string:*Account:1001"},
		Increment:    Float64Pointer(20),
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
	if _, err := NewTPBalanceCostIncrement(EmptyString, invalidStr, EmptyString, EmptyString); err == nil || err.Error() != expectedErr {
		t.Errorf("Expected %+v, received %+v", expectedErr, err)
	}
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
		Increment:    Float64Pointer(20),
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

func TestATDUsage(t *testing.T) {
	ev := &CGREvent{
		ID: "testID",
		APIOpts: map[string]interface{}{
			OptsRatesUsage: true,
		},
	}

	_, err := ev.OptsAsDecimal(decimal.New(int64(time.Minute), 0), OptsRatesUsage, MetaUsage)
	expected := "cannot convert field: bool to decimal.Big"
	if err == nil || err.Error() != expected {
		t.Errorf("\nExpected: <%+v>, \nReceived: <%+v>", expected, err.Error())
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
}

func TestIntervalStart(t *testing.T) {
	args := &CGREvent{
		APIOpts: map[string]interface{}{
			OptsRatesIntervalStart:  "1ns",
			OptsRatesRateProfileIDs: []string{"RP_1001"},
		},
	}
	rcv, err := args.OptsAsDecimal(decimal.New(0, 0), OptsRatesIntervalStart)
	exp := new(decimal.Big).SetUint64(1)
	if err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("Expected %v but received %v", rcv, exp)
	}
}

func TestIntervalStartDefault(t *testing.T) {
	args := &CGREvent{
		APIOpts: map[string]interface{}{
			OptsRatesRateProfileIDs: []string{"RP_1001"},
		},
	}
	rcv, err := args.OptsAsDecimal(decimal.New(0, 0), OptsRatesIntervalStart)
	exp := new(decimal.Big).SetUint64(0)
	if err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("Expected %v but received %v", rcv, exp)
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
