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
	"strings"
	"testing"
	"time"

	"github.com/ericlagergren/decimal"
)

func TestTPDistinctIdsString(t *testing.T) {
	eIn1 := []string{"1", "2", "3", "4"}
	eIn2 := TPDistinctIds(eIn1)
	expected := strings.Join(eIn1, FieldsSep)
	received := eIn2.String()

	if !reflect.DeepEqual(expected, received) {
		t.Errorf("Expecting: %+v, received: %+v", expected, received)
	}
}

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

func TestNewRateSlot(t *testing.T) {
	var err error
	eOut := &RateSlot{
		ConnectFee:            1,
		Rate:                  1.01,
		RateUnit:              "1",
		RateIncrement:         "1",
		GroupIntervalStart:    "1",
		rateUnitDur:           1,
		rateIncrementDur:      1,
		groupIntervalStartDur: 1,
		tag:                   "",
	}
	rcv, err := NewRateSlot(eOut.ConnectFee, eOut.Rate, eOut.RateUnit, eOut.RateIncrement, eOut.GroupIntervalStart)
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expected: %+v ,received: %+v ", eOut, rcv)
	}
	eOut.RateUnit = "a"
	_, err = NewRateSlot(eOut.ConnectFee, eOut.Rate, eOut.RateUnit, eOut.RateIncrement, eOut.GroupIntervalStart)
	//must receive from time an error: "invalid duration a"
	if err == nil || err.Error() != "time: invalid duration \"a\"" {
		t.Error(err)
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

func TestSetDurations(t *testing.T) {
	rs := &RateSlot{
		RateUnit:           "1",
		RateIncrement:      "1",
		GroupIntervalStart: "1",
	}
	//must receive "nil" with
	if err := rs.SetDurations(); err != nil {
		t.Error(err)
	}
	rs.RateUnit = "a"
	//at RateUnit if, must receive from time an error: "invalid duration a"
	if err := rs.SetDurations(); err == nil || err.Error() != "time: invalid duration \"a\"" {
		t.Error(err)
	}
	rs.RateUnit = "1"
	rs.RateIncrement = "a"
	//at RateIncrement, must receive from time an error: "invalid duration a"
	if err := rs.SetDurations(); err == nil || err.Error() != "time: invalid duration \"a\"" {
		t.Error(err)
	}
	rs.RateIncrement = "1"
	rs.GroupIntervalStart = "a"
	//at GroupIntervalStart, must receive from time an error: "invalid duration a"
	if err := rs.SetDurations(); err == nil || err.Error() != "time: invalid duration \"a\"" {
		t.Error(err)
	}
}

func TestRateUnitDuration(t *testing.T) {
	eOut := &RateSlot{
		rateUnitDur:           1,
		rateIncrementDur:      1,
		groupIntervalStartDur: 1,
	}
	rcv := eOut.RateUnitDuration()
	if rcv != eOut.rateUnitDur {
		t.Errorf("Expected %+v, received %+v", eOut.rateUnitDur, rcv)
	}
	rcv = eOut.RateIncrementDuration()
	if rcv != eOut.rateIncrementDur {
		t.Errorf("Expected %+v, received %+v", eOut.rateIncrementDur, rcv)
	}
	rcv = eOut.GroupIntervalStartDuration()
	if rcv != eOut.groupIntervalStartDur {
		t.Errorf("Expected %+v, received %+v", eOut.groupIntervalStartDur, rcv)
	}
}

func TestNewTiming(t *testing.T) {
	eOut := &TPTiming{
		ID:        "1",
		Years:     Years{2020},
		Months:    []time.Month{time.April},
		MonthDays: MonthDays{18},
		WeekDays:  WeekDays{06},
		StartTime: "00:00:00",
		EndTime:   "11:11:11",
	}
	rcv := NewTiming("1", "2020", "04", "18", "06", "00:00:00;11:11:11")
	if !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expected %+v, received %+v", eOut, rcv)
	}
	//without endtime, check if .Split method works (only one timestamp)
	rcv = NewTiming("1", "2020", "04", "18", "06", "00:00:00")
	eOut.EndTime = ""
	if !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expected %+v, received %+v", eOut, rcv)
	}
	//check if .Split method works (ignoring the last timestamp)
	rcv = NewTiming("1", "2020", "04", "18", "06", "00:00:00;11:11:11;22:22:22")
	eOut.EndTime = "11:11:11"
	if !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expected %+v, received %+v", eOut, rcv)
	}
}

func TestTPTimingSetTiming(t *testing.T) {
	tpTiming := &TPTiming{
		ID:        "1",
		Years:     Years{2020},
		Months:    []time.Month{time.April},
		MonthDays: MonthDays{18},
		WeekDays:  WeekDays{06},
		StartTime: "00:00:00",
		EndTime:   "11:11:11",
	}
	tpRatingPlanBinding := new(TPRatingPlanBinding)
	tpRatingPlanBinding.SetTiming(tpTiming)
	if !reflect.DeepEqual(tpTiming, tpRatingPlanBinding.timing) {
		t.Errorf("Expected %+v, received %+v", tpTiming, tpRatingPlanBinding.timing)
	}
	rcv := tpRatingPlanBinding.Timing()
	if !reflect.DeepEqual(tpTiming, rcv) {
		t.Errorf("Expected %+v, received %+v", tpTiming, rcv)
	}
}

func TestTPRatingProfileKeys(t *testing.T) {
	//empty check -> KeyId
	tpRatingProfile := new(TPRatingProfile)
	eOut := "*out:::"
	rcv := tpRatingProfile.KeyId()
	if !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expected %+v, received %+v", eOut, rcv)
	}
	//empty check -> GetId
	eOut = ":*out:::"
	rcv = tpRatingProfile.GetId()
	if !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expected %+v, received %+v", eOut, rcv)
	}
	// test check -> KeyId
	tpRatingProfile.Tenant = "test"
	tpRatingProfile.Category = "test"
	tpRatingProfile.Subject = "test"
	eOut = "*out:test:test:test"
	rcv = tpRatingProfile.KeyId()
	if !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expected %+v, received %+v", eOut, rcv)
	}
	//test check -> GetId
	eOut = "test:*out:test:test:test"
	tpRatingProfile.TPid = "test"
	tpRatingProfile.LoadId = "test"

	rcv = tpRatingProfile.GetId()
	if !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expected %+v, received %+v", eOut, rcv)
	}
}

func TestTPRatingProfileSetRatingProfilesId(t *testing.T) {
	//empty check
	tpRatingProfile := new(TPRatingProfile)
	tpRatingProfile.SetRatingProfileID("")
	eOut := new(TPRatingProfile)
	if !reflect.DeepEqual(eOut, tpRatingProfile) {
		t.Errorf("Expected %+v, received %+v", eOut, tpRatingProfile)
	}
	//test check
	tpRatingProfile.SetRatingProfileID("1:3:4:5")
	eOut.LoadId = "1"
	eOut.Tenant = "3"
	eOut.Category = "4"
	eOut.Subject = "5"
	if !reflect.DeepEqual(eOut, tpRatingProfile) {
		t.Errorf("Expected %+v, received %+v", eOut, tpRatingProfile)
	}
	//wrong TPRatingProfile sent
	err := tpRatingProfile.SetRatingProfileID("1:2:3:4:5:6")
	if err == nil {
		t.Error("Wrong TPRatingProfileId sent and no error received")
	}

}

func TestAttrGetRatingProfileGetID(t *testing.T) {
	//empty check
	attrGetRatingProfile := &AttrGetRatingProfile{
		Tenant:   "",
		Category: "",
		Subject:  "",
	}
	rcv := attrGetRatingProfile.GetID()
	eOut := "*out:::"
	if !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expected %+v, received %+v", eOut, rcv)
	}
	//test check
	attrGetRatingProfile.Tenant = "cgrates"
	attrGetRatingProfile.Category = "cgrates"
	attrGetRatingProfile.Subject = "cgrates"
	rcv = attrGetRatingProfile.GetID()
	eOut = "*out:cgrates:cgrates:cgrates"
	if !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expected %+v, received %+v", eOut, rcv)
	}
}

func TestFallbackSubjKeys(t *testing.T) {
	//call with empty slice -> expect an empty slice
	rcv := FallbackSubjKeys("", "", "")
	if rcv != nil {
		t.Errorf("Expected an empty slice")
	}
	//check with test vars
	eOut := []string{"*out:cgrates.org:*voice:1001", "*out:cgrates.org:*voice:1002", "*out:cgrates.org:*voice:1003"}
	rcv = FallbackSubjKeys("cgrates.org", MetaVoice, "1001;1003;1002")
	if !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expected %+v, received %+v", eOut, rcv)
	}
}

func TestTPAccountActionsKeyId(t *testing.T) {
	//empty check KeyIs()
	tPAccountActions := new(TPAccountActions)
	rcv := tPAccountActions.KeyId()
	eOut := ":"
	if !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expected %+v, received %+v", eOut, rcv)
	}
	//empty check GetId()
	rcv = tPAccountActions.GetId()
	eOut = "::"
	if !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expected %+v, received %+v", eOut, rcv)
	}
	//empty check SetAccountActionsId()
	err := tPAccountActions.SetAccountActionsId("")
	if err == nil {
		t.Errorf("Wrong TP Account Action Id not received")
	}
	//test check SetAccountActionsId()
	err = tPAccountActions.SetAccountActionsId("loadID:cgrates.org:cgrates")
	if err != nil {
		t.Errorf("Expected nil")
	}
	expectedOut := &TPAccountActions{
		LoadId:  "loadID",
		Tenant:  "cgrates.org",
		Account: "cgrates",
	}
	if !reflect.DeepEqual(expectedOut, tPAccountActions) {
		t.Errorf("Expected %+v, received %+v", ToJSON(expectedOut), ToJSON(tPAccountActions))
	}
	//test check KeyIs() *Tenant, Account setted above via SetAccountActionsId
	rcv = tPAccountActions.KeyId()
	eOut = "cgrates.org:cgrates"
	if !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expected %+v, received %+v", eOut, rcv)
	}
	//tect check GetID() *LoadId, tenant, account setted above via SetAccountActionsId
	rcv = tPAccountActions.GetId()
	eOut = "loadID:cgrates.org:cgrates"
	if !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expected %+v, received %+v", eOut, rcv)
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

func TestNewTAFromAccountKey(t *testing.T) {
	//check with empty string
	eOut := &TenantAccount{
		Tenant:  "",
		Account: "",
	}
	rcv, err := NewTAFromAccountKey(":")
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expected: %s ,received: %s ", ToJSON(eOut), ToJSON(rcv))
	}
	//check with test string
	eOut = &TenantAccount{
		Tenant:  "cgrates.org",
		Account: "1001",
	}
	rcv, err = NewTAFromAccountKey("cgrates.org:1001")
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expected: %s ,received: %s ", ToJSON(eOut), ToJSON(rcv))
	}
	//check with wrong TenantAccount
	_, err = NewTAFromAccountKey("")
	if err == nil {
		t.Errorf("Unsupported format not processed")
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

func TestArgRSv1ResourceUsageCloneCase1(t *testing.T) {
	expectedArgRSv1 := &ArgRSv1ResourceUsage{
		clnb: true,
	}
	newArgRSv1 := new(ArgRSv1ResourceUsage)
	newArgRSv1.SetCloneable(true)
	if !reflect.DeepEqual(expectedArgRSv1, newArgRSv1) {
		t.Errorf("Expected %+v, received %+v", expectedArgRSv1, newArgRSv1)
	}
}

func TestArgRSv1ResourceUsageCloneCase2(t *testing.T) {
	newArgRSv1 := &ArgRSv1ResourceUsage{
		CGREvent: &CGREvent{
			Tenant:  "*req.CGRID",
			APIOpts: map[string]interface{}{},
		},
		UsageID:  "randomID",
		UsageTTL: DurationPointer(2),
		Units:    1.0,
	}
	if replyArgRsv1, err := newArgRSv1.RPCClone(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(newArgRSv1, replyArgRsv1) {
		t.Errorf("Expected %+v \n, received %+v", ToJSON(newArgRSv1), ToJSON(replyArgRsv1))
	}
}

func TestArgRSv1ResourceUsageCloneCase3(t *testing.T) {
	newArgRSv1 := &ArgRSv1ResourceUsage{
		CGREvent: &CGREvent{
			Tenant:  "*req.CGRID",
			Event:   map[string]interface{}{},
			APIOpts: map[string]interface{}{},
		},
		UsageID:  "randomID",
		UsageTTL: DurationPointer(2),
		Units:    1.0,
		clnb:     true,
	}
	if replyArgRsv1, err := newArgRSv1.RPCClone(); err != nil {
		t.Error(err)
	} else if newArgRSv1.clnb = false; !reflect.DeepEqual(newArgRSv1, replyArgRsv1) {
		t.Errorf("Expected %+v \n, received %+v", ToJSON(newArgRSv1), ToJSON(replyArgRsv1))
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

func TestAppendToSMCostFilter(t *testing.T) {
	var err error
	smfltr := new(SMCostFilter)
	expected := &SMCostFilter{
		CGRIDs: []string{"CGRID1", "CGRID2"},
	}
	if smfltr, err = AppendToSMCostFilter(smfltr, MetaString, MetaScPrefix+CGRID, []string{"CGRID1", "CGRID2"}, ""); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(smfltr, expected) {
		t.Errorf("Expected: %s ,received: %s ", ToJSON(expected), ToJSON(smfltr))
	}
	expected.NotCGRIDs = []string{"CGRID3", "CGRID4"}
	if smfltr, err = AppendToSMCostFilter(smfltr, "*notstring", MetaScPrefix+CGRID, []string{"CGRID3", "CGRID4"}, ""); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(smfltr, expected) {
		t.Errorf("Expected: %s ,received: %s ", ToJSON(expected), ToJSON(smfltr))
	}

	expected.RunIDs = []string{"RunID1", "RunID2"}
	if smfltr, err = AppendToSMCostFilter(smfltr, MetaString, MetaScPrefix+RunID, []string{"RunID1", "RunID2"}, ""); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(smfltr, expected) {
		t.Errorf("Expected: %s ,received: %s ", ToJSON(expected), ToJSON(smfltr))
	}
	expected.NotRunIDs = []string{"RunID3", "RunID4"}
	if smfltr, err = AppendToSMCostFilter(smfltr, "*notstring", MetaScPrefix+RunID, []string{"RunID3", "RunID4"}, ""); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(smfltr, expected) {
		t.Errorf("Expected: %s ,received: %s ", ToJSON(expected), ToJSON(smfltr))
	}

	expected.OriginHosts = []string{"OriginHost1", "OriginHost2"}
	if smfltr, err = AppendToSMCostFilter(smfltr, MetaString, MetaScPrefix+OriginHost, []string{"OriginHost1", "OriginHost2"}, ""); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(smfltr, expected) {
		t.Errorf("Expected: %s ,received: %s ", ToJSON(expected), ToJSON(smfltr))
	}
	expected.NotOriginHosts = []string{"OriginHost3", "OriginHost4"}
	if smfltr, err = AppendToSMCostFilter(smfltr, "*notstring", MetaScPrefix+OriginHost, []string{"OriginHost3", "OriginHost4"}, ""); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(smfltr, expected) {
		t.Errorf("Expected: %s ,received: %s ", ToJSON(expected), ToJSON(smfltr))
	}

	expected.OriginIDs = []string{"OriginID1", "OriginID2"}
	if smfltr, err = AppendToSMCostFilter(smfltr, MetaString, MetaScPrefix+OriginID, []string{"OriginID1", "OriginID2"}, ""); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(smfltr, expected) {
		t.Errorf("Expected: %s ,received: %s ", ToJSON(expected), ToJSON(smfltr))
	}
	expected.NotOriginIDs = []string{"OriginID3", "OriginID4"}
	if smfltr, err = AppendToSMCostFilter(smfltr, "*notstring", MetaScPrefix+OriginID, []string{"OriginID3", "OriginID4"}, ""); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(smfltr, expected) {
		t.Errorf("Expected: %s ,received: %s ", ToJSON(expected), ToJSON(smfltr))
	}

	expected.CostSources = []string{"CostSource1", "CostSource2"}
	if smfltr, err = AppendToSMCostFilter(smfltr, MetaString, MetaScPrefix+CostSource, []string{"CostSource1", "CostSource2"}, ""); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(smfltr, expected) {
		t.Errorf("Expected: %s ,received: %s ", ToJSON(expected), ToJSON(smfltr))
	}
	expected.NotCostSources = []string{"CostSource3", "CostSource4"}
	if smfltr, err = AppendToSMCostFilter(smfltr, "*notstring", MetaScPrefix+CostSource, []string{"CostSource3", "CostSource4"}, ""); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(smfltr, expected) {
		t.Errorf("Expected: %s ,received: %s ", ToJSON(expected), ToJSON(smfltr))
	}

	if smfltr, err = AppendToSMCostFilter(smfltr, "*prefix", MetaScPrefix+CGRID, []string{"CGRID1", "CGRID2"}, ""); err == nil || err.Error() != "FilterType: \"*prefix\" not supported for FieldName: \"~*sc.CGRID\"" {
		t.Errorf("Expected error: FilterType: \"*prefix\" not supported for FieldName: \"~*sc.CGRID\" ,received %v", err)
	}
	if !reflect.DeepEqual(smfltr, expected) {
		t.Errorf("Expected: %s ,received: %s ", ToJSON(expected), ToJSON(smfltr))
	}
	if smfltr, err = AppendToSMCostFilter(smfltr, "*prefix", MetaScPrefix+RunID, []string{"CGRID1", "CGRID2"}, ""); err == nil || err.Error() != "FilterType: \"*prefix\" not supported for FieldName: \"~*sc.RunID\"" {
		t.Errorf("Expected error: FilterType: \"*prefix\" not supported for FieldName: \"~*sc.RunID\" ,received %v", err)
	}
	if !reflect.DeepEqual(smfltr, expected) {
		t.Errorf("Expected: %s ,received: %s ", ToJSON(expected), ToJSON(smfltr))
	}
	if smfltr, err = AppendToSMCostFilter(smfltr, "*prefix", MetaScPrefix+OriginHost, []string{"CGRID1", "CGRID2"}, ""); err == nil || err.Error() != "FilterType: \"*prefix\" not supported for FieldName: \"~*sc.OriginHost\"" {
		t.Errorf("Expected error: FilterType: \"*prefix\" not supported for FieldName: \"~*sc.OriginHost\" ,received %v", err)
	}
	if !reflect.DeepEqual(smfltr, expected) {
		t.Errorf("Expected: %s ,received: %s ", ToJSON(expected), ToJSON(smfltr))
	}
	if smfltr, err = AppendToSMCostFilter(smfltr, "*prefix", MetaScPrefix+OriginID, []string{"CGRID1", "CGRID2"}, ""); err == nil || err.Error() != "FilterType: \"*prefix\" not supported for FieldName: \"~*sc.OriginID\"" {
		t.Errorf("Expected error: FilterType: \"*prefix\" not supported for FieldName: \"~*sc.OriginID\" ,received %v", err)
	}
	if !reflect.DeepEqual(smfltr, expected) {
		t.Errorf("Expected: %s ,received: %s ", ToJSON(expected), ToJSON(smfltr))
	}
	if smfltr, err = AppendToSMCostFilter(smfltr, "*prefix", MetaScPrefix+CostSource, []string{"CGRID1", "CGRID2"}, ""); err == nil || err.Error() != "FilterType: \"*prefix\" not supported for FieldName: \"~*sc.CostSource\"" {
		t.Errorf("Expected error: FilterType: \"*prefix\" not supported for FieldName: \"~*sc.CostSource\" ,received %v", err)
	}
	if !reflect.DeepEqual(smfltr, expected) {
		t.Errorf("Expected: %s ,received: %s ", ToJSON(expected), ToJSON(smfltr))
	}
	if smfltr, err = AppendToSMCostFilter(smfltr, MetaString, CGRID, []string{"CGRID1", "CGRID2"}, ""); err == nil || err.Error() != "FieldName: \"CGRID\" not supported" {
		t.Errorf("Expected error: FieldName: \"CGRID\" not supported ,received %v", err)
	}
	if !reflect.DeepEqual(smfltr, expected) {
		t.Errorf("Expected: %s ,received: %s ", ToJSON(expected), ToJSON(smfltr))
	}
	if smfltr, err = AppendToSMCostFilter(smfltr, "*prefix", CGRID, []string{"CGRID1", "CGRID2"}, ""); err == nil || err.Error() != "FieldName: \"CGRID\" not supported" {
		t.Errorf("Expected error: FieldName: \"CGRID\" not supported ,received %v", err)
	}
	if !reflect.DeepEqual(smfltr, expected) {
		t.Errorf("Expected: %s ,received: %s ", ToJSON(expected), ToJSON(smfltr))

	}
	expected.Usage.Min = DurationPointer(time.Second)
	if smfltr, err = AppendToSMCostFilter(smfltr, "*gte", MetaScPrefix+Usage, []string{"1s", "2s"}, ""); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(smfltr, expected) {
		t.Errorf("Expected: %s ,received: %s ", ToJSON(expected), ToJSON(smfltr))
	}
	expected.Usage.Max = DurationPointer(3 * time.Second)
	if smfltr, err = AppendToSMCostFilter(smfltr, "*lt", MetaScPrefix+Usage, []string{"3s", "4s"}, ""); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(smfltr, expected) {
		t.Errorf("Expected: %s ,received: %s ", ToJSON(expected), ToJSON(smfltr))
	}

	if smfltr, err = AppendToSMCostFilter(smfltr, "*gte", MetaScPrefix+Usage, []string{"one second"}, ""); err == nil || err.Error() != "Error when converting field: \"*gte\"  value: \"~*sc.Usage\" in time.Duration " {
		t.Errorf("Expected error: Error when converting field: \"*gte\"  value: \"~*sc.Usage\" in time.Duration ,received %v", err)
	}
	if !reflect.DeepEqual(smfltr, expected) {
		t.Errorf("Expected: %s ,received: %s ", ToJSON(expected), ToJSON(smfltr))
	}

	if smfltr, err = AppendToSMCostFilter(smfltr, "*lt", MetaScPrefix+Usage, []string{"one second"}, ""); err == nil || err.Error() != "Error when converting field: \"*lt\"  value: \"~*sc.Usage\" in time.Duration " {
		t.Errorf("Expected error: Error when converting field: \"*lt\"  value: \"~*sc.Usage\" in time.Duration ,received %v", err)
	}
	if !reflect.DeepEqual(smfltr, expected) {
		t.Errorf("Expected: %s ,received: %s ", ToJSON(expected), ToJSON(smfltr))
	}
	if smfltr, err = AppendToSMCostFilter(smfltr, "*prefix", MetaScPrefix+Usage, []string{"CGRID1", "CGRID2"}, ""); err == nil || err.Error() != "FilterType: \"*prefix\" not supported for FieldName: \"~*sc.Usage\"" {
		t.Errorf("Expected error: FilterType: \"*prefix\" not supported for FieldName: \"~*sc.Usage\" ,received %v", err)
	}
	if !reflect.DeepEqual(smfltr, expected) {
		t.Errorf("Expected: %s ,received: %s ", ToJSON(expected), ToJSON(smfltr))
	}

	now := time.Now().UTC().Round(time.Second)
	strNow := now.Format("2006-01-02T15:04:05")

	expected.CreatedAt.Begin = &now
	if smfltr, err = AppendToSMCostFilter(smfltr, "*gte", MetaScPrefix+CreatedAt, []string{strNow}, ""); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(smfltr, expected) {
		t.Errorf("Expected: %s ,received: %s ", ToJSON(expected), ToJSON(smfltr))
	}

	expected.CreatedAt.End = &now
	if smfltr, err = AppendToSMCostFilter(smfltr, "*lt", MetaScPrefix+CreatedAt, []string{strNow}, ""); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(smfltr, expected) {
		t.Errorf("Expected: %s ,received: %s ", ToJSON(expected), ToJSON(smfltr))
	}

	if smfltr, err = AppendToSMCostFilter(smfltr, "*gte", MetaScPrefix+CreatedAt, []string{time.Now().String()}, ""); err == nil || err.Error() != "Error when converting field: \"*gte\"  value: \"~*sc.CreatedAt\" in time.Time " {
		t.Errorf("Expected error: Error when converting field: \"*gte\"  value: \"~*sc.CreatedAt\" in time.Time ,received %v", err)
	}
	if !reflect.DeepEqual(smfltr, expected) {
		t.Errorf("Expected: %s ,received: %s ", ToJSON(expected), ToJSON(smfltr))
	}
	if smfltr, err = AppendToSMCostFilter(smfltr, "*lt", MetaScPrefix+CreatedAt, []string{time.Now().String()}, ""); err == nil || err.Error() != "Error when converting field: \"*lt\"  value: \"~*sc.CreatedAt\" in time.Time " {
		t.Errorf("Expected error: Error when converting field: \"*lt\"  value: \"~*sc.CreatedAt\" in time.Time ,received %v", err)
	}
	if !reflect.DeepEqual(smfltr, expected) {
		t.Errorf("Expected: %s ,received: %s ", ToJSON(expected), ToJSON(smfltr))
	}
	if smfltr, err = AppendToSMCostFilter(smfltr, "*prefix", MetaScPrefix+CreatedAt, []string{"CGRID1", "CGRID2"}, ""); err == nil || err.Error() != "FilterType: \"*prefix\" not supported for FieldName: \"~*sc.CreatedAt\"" {
		t.Errorf("Expected error: FilterType: \"*prefix\" not supported for FieldName: \"~*sc.CreatedAt\" ,received %v", err)
	}
	if !reflect.DeepEqual(smfltr, expected) {
		t.Errorf("Expected: %s ,received: %s ", ToJSON(expected), ToJSON(smfltr))
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
		ArgsCache: map[string][]string{
			DestinationIDs:             nil,
			ReverseDestinationIDs:      nil,
			RatingPlanIDs:              nil,
			RatingProfileIDs:           nil,
			ActionIDs:                  nil,
			ActionPlanIDs:              nil,
			AccountActionPlanIDs:       nil,
			ActionTriggerIDs:           nil,
			SharedGroupIDs:             nil,
			ResourceProfileIDs:         nil,
			ResourceIDs:                nil,
			StatsQueueIDs:              nil,
			StatsQueueProfileIDs:       nil,
			ThresholdIDs:               nil,
			ThresholdProfileIDs:        nil,
			FilterIDs:                  nil,
			RouteProfileIDs:            nil,
			AttributeProfileIDs:        nil,
			ChargerProfileIDs:          nil,
			DispatcherProfileIDs:       nil,
			DispatcherHostIDs:          nil,
			RateProfileIDs:             nil,
			TimingIDs:                  nil,
			AttributeFilterIndexIDs:    nil,
			ResourceFilterIndexIDs:     nil,
			StatFilterIndexIDs:         nil,
			ThresholdFilterIndexIDs:    nil,
			RouteFilterIndexIDs:        nil,
			ChargerFilterIndexIDs:      nil,
			DispatcherFilterIndexIDs:   nil,
			RateProfilesFilterIndexIDs: nil,
			RateFilterIndexIDs:         nil,
			FilterIndexIDs:             nil,
		},
	}
	eMap := NewAttrReloadCacheWithOpts()
	if !reflect.DeepEqual(eMap, newAttrReloadCache) {
		t.Errorf("Expected %+v \n, received %+v", eMap, newAttrReloadCache)
	}
}

func TestStartTimeNow(t *testing.T) {
	testCostEventStruct := &ArgsCostForEvent{
		RateProfileIDs: []string{"123", "456", "789"},
		CGREvent: &CGREvent{
			Tenant:  "*req.CGRID",
			ID:      "",
			Event:   map[string]interface{}{},
			APIOpts: map[string]interface{}{},
		},
	}
	timpulet1 := time.Now()
	result, err := testCostEventStruct.StartTime("")
	timpulet2 := time.Now()
	if err != nil {
		t.Errorf("Expected <nil> , received <%+v>", err)
	}
	if result.Before(timpulet1) && result.After(timpulet2) {
		t.Errorf("Expected between <%+v> and <%+v>, received <%+v>", timpulet1, timpulet2, result)
	}
}

func TestStartTime(t *testing.T) {
	testCostEventStruct := &ArgsCostForEvent{
		RateProfileIDs: []string{"123", "456", "789"},
		CGREvent: &CGREvent{
			Tenant:  "*req.CGRID",
			ID:      "",
			Event:   map[string]interface{}{},
			APIOpts: map[string]interface{}{"*ratesStartTime": "2018-01-07T17:00:10Z"},
		},
	}
	if result, err := testCostEventStruct.StartTime(""); err != nil {
		t.Errorf("Expected <nil> , received <%+v>", err)
	} else if !reflect.DeepEqual(result.String(), "2018-01-07 17:00:10 +0000 UTC") {
		t.Errorf("Expected <2018-01-07 17:00:10 +0000 UTC> , received <%+v>", result)
	}
}

func TestStartTimeError(t *testing.T) {
	testCostEventStruct := &ArgsCostForEvent{
		RateProfileIDs: []string{"123", "456", "789"},
		CGREvent: &CGREvent{
			Tenant:  "*req.CGRID",
			ID:      "",
			Event:   map[string]interface{}{},
			APIOpts: map[string]interface{}{"*ratesStartTime": "start"},
		},
	}
	_, err := testCostEventStruct.StartTime("")
	if err == nil && err.Error() != "received <Unsupported time format" {
		t.Errorf("Expected <nil> , received <%+v>", err)
	}
}

func TestUsageMinute(t *testing.T) {
	testCostEventStruct := &ArgsCostForEvent{
		RateProfileIDs: []string{"123", "456", "789"},
		CGREvent: &CGREvent{
			Tenant:  "*req.CGRID",
			ID:      "",
			Event:   map[string]interface{}{},
			APIOpts: map[string]interface{}{},
		},
	}
	if rcv, err := testCostEventStruct.Usage(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(decimal.New(int64(time.Minute), 0), rcv) {
		t.Errorf("Expected %+v, received %+v", decimal.New(int64(time.Minute), 0), rcv)
	}
}

func TestUsageError(t *testing.T) {
	testCostEventStruct := &ArgsCostForEvent{
		RateProfileIDs: []string{"123", "456", "789"},
		CGREvent: &CGREvent{
			Tenant:  "*req.CGRID",
			ID:      "",
			Event:   map[string]interface{}{},
			APIOpts: map[string]interface{}{"*ratesUsage": "start"},
		},
	}
	_, err := testCostEventStruct.Usage()
	if err == nil && err.Error() != "received <Unsupported time format" {
		t.Errorf("Expected <nil> , received <%+v>", err)
	}
}

func TestUsage(t *testing.T) {
	testCostEventStruct := &ArgsCostForEvent{
		RateProfileIDs: []string{"123", "456", "789"},
		CGREvent: &CGREvent{
			Tenant:  "*req.CGRID",
			ID:      "",
			Event:   map[string]interface{}{},
			APIOpts: map[string]interface{}{"*ratesUsage": "2m10s"},
		},
	}

	if result, err := testCostEventStruct.Usage(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(result.String(), "130000000000") {
		t.Errorf("Expected <130000000000> , received <%+v>", result.String())
	}
}

func TestATDUsage(t *testing.T) {
	args := &ArgsCostForEvent{
		CGREvent: &CGREvent{
			ID: "testID",
			Event: map[string]interface{}{
				Usage: true,
			},
		},
	}

	_, err := args.Usage()
	expected := "cannot convert field: true to time.Duration"
	if err == nil || err.Error() != expected {
		t.Errorf("\nExpected: <%+v>, \nReceived: <%+v>", expected, err.Error())
	}
}
