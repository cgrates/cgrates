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
	eOut := []string{"*out:cgrates.org:*voice:1001", "*out:cgrates.org:*voice:1003", "*out:cgrates.org:*voice:1002"}
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

// now working here
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
		ExtraArgs: map[string]any{
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
		DestinationIDs:           []string{MetaAny},
		ReverseDestinationIDs:    []string{MetaAny},
		RatingPlanIDs:            []string{MetaAny},
		RatingProfileIDs:         []string{MetaAny},
		ActionIDs:                []string{MetaAny},
		ActionPlanIDs:            []string{MetaAny},
		AccountActionPlanIDs:     []string{MetaAny},
		ActionTriggerIDs:         []string{MetaAny},
		SharedGroupIDs:           []string{MetaAny},
		ResourceProfileIDs:       []string{MetaAny},
		ResourceIDs:              []string{MetaAny},
		StatsQueueIDs:            []string{MetaAny},
		StatsQueueProfileIDs:     []string{MetaAny},
		ThresholdIDs:             []string{MetaAny},
		ThresholdProfileIDs:      []string{MetaAny},
		TrendIDs:                 []string{MetaAny},
		TrendProfileIDs:          []string{MetaAny},
		FilterIDs:                []string{MetaAny},
		RouteProfileIDs:          []string{MetaAny},
		AttributeProfileIDs:      []string{MetaAny},
		ChargerProfileIDs:        []string{MetaAny},
		DispatcherProfileIDs:     []string{MetaAny},
		DispatcherHostIDs:        []string{MetaAny},
		Dispatchers:              []string{MetaAny},
		TimingIDs:                []string{MetaAny},
		AttributeFilterIndexIDs:  []string{MetaAny},
		ResourceFilterIndexIDs:   []string{MetaAny},
		StatFilterIndexIDs:       []string{MetaAny},
		ThresholdFilterIndexIDs:  []string{MetaAny},
		RouteFilterIndexIDs:      []string{MetaAny},
		ChargerFilterIndexIDs:    []string{MetaAny},
		DispatcherFilterIndexIDs: []string{MetaAny},
		FilterIndexIDs:           []string{MetaAny},
		RankingIDs:               []string{MetaAny},
		RankingProfileIDs:        []string{MetaAny},
	}
	eMap := NewAttrReloadCacheWithOpts()
	if !reflect.DeepEqual(eMap, newAttrReloadCache) {
		t.Errorf("Expected %+v \n, received %+v", eMap, newAttrReloadCache)
	}
}

func TestNewAttrReloadCacheWithOptsFromMap(t *testing.T) {
	excluded := NewStringSet([]string{MetaAPIBan, MetaSentryPeer, MetaAccounts, MetaLoadIDs})
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

func TestIsActiveAt(t *testing.T) {
	tests := []struct {
		name      string
		timing    TPTiming
		checkTime time.Time
		expected  bool
	}{
		{
			name: "Active timing",
			timing: TPTiming{
				Years:     Years{2024},
				Months:    Months{time.January},
				MonthDays: MonthDays{15},
				WeekDays:  WeekDays{time.Monday},
				StartTime: "09:00:00",
				EndTime:   "17:00:00",
			},
			checkTime: time.Date(2024, time.January, 15, 10, 0, 0, 0, time.UTC),
			expected:  true,
		},
		{
			name: "Inactive year",
			timing: TPTiming{
				Years:     Years{2023},
				Months:    Months{time.January},
				MonthDays: MonthDays{15},
				WeekDays:  WeekDays{time.Monday},
				StartTime: "09:00:00",
				EndTime:   "17:00:00",
			},
			checkTime: time.Date(2024, time.January, 15, 10, 0, 0, 0, time.UTC),
			expected:  false,
		},
		{
			name: "Inactive month",
			timing: TPTiming{
				Years:     Years{2024},
				Months:    Months{time.February},
				MonthDays: MonthDays{15},
				WeekDays:  WeekDays{time.Monday},
				StartTime: "09:00:00",
				EndTime:   "17:00:00",
			},
			checkTime: time.Date(2024, time.January, 15, 10, 0, 0, 0, time.UTC),
			expected:  false,
		},
		{
			name: "Inactive day",
			timing: TPTiming{
				Years:     Years{2024},
				Months:    Months{time.January},
				MonthDays: MonthDays{16},
				WeekDays:  WeekDays{time.Monday},
				StartTime: "09:00:00",
				EndTime:   "17:00:00",
			},
			checkTime: time.Date(2024, time.January, 15, 10, 0, 0, 0, time.UTC),
			expected:  false,
		},
		{
			name: "Inactive weekday",
			timing: TPTiming{
				Years:     Years{2024},
				Months:    Months{time.January},
				MonthDays: MonthDays{15},
				WeekDays:  WeekDays{time.Wednesday},
				StartTime: "09:00:00",
				EndTime:   "17:00:00",
			},
			checkTime: time.Date(2024, time.January, 15, 10, 0, 0, 0, time.UTC),
			expected:  false,
		},
		{
			name: "Before start time",
			timing: TPTiming{
				Years:     Years{2024},
				Months:    Months{time.January},
				MonthDays: MonthDays{15},
				WeekDays:  WeekDays{time.Monday},
				StartTime: "12:00:00",
				EndTime:   "17:00:00",
			},
			checkTime: time.Date(2024, time.January, 15, 11, 0, 0, 0, time.UTC),
			expected:  false,
		},
		{
			name: "After end time",
			timing: TPTiming{
				Years:     Years{2024},
				Months:    Months{time.January},
				MonthDays: MonthDays{15},
				WeekDays:  WeekDays{time.Monday},
				StartTime: "09:00:00",
				EndTime:   "12:00:00",
			},
			checkTime: time.Date(2024, time.January, 15, 13, 0, 0, 0, time.UTC),
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.timing.IsActiveAt(tt.checkTime)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestGetRightMargin(t *testing.T) {
	tests := []struct {
		name      string
		timing    TPTiming
		checkTime time.Time
		expected  time.Time
	}{
		{
			name: "With specific end time",
			timing: TPTiming{
				EndTime: "15:30:00",
			},
			checkTime: time.Date(2024, time.January, 15, 10, 0, 0, 0, time.UTC),
			expected:  time.Date(2024, time.January, 15, 15, 30, 0, 0, time.UTC),
		},
		{
			name: "With default end of the day",
			timing: TPTiming{
				EndTime: "",
			},
			checkTime: time.Date(2024, time.January, 15, 10, 0, 0, 0, time.UTC),
			expected:  time.Date(2024, time.January, 15, 23, 59, 59, 0, time.UTC).Add(time.Second),
		},
		{
			name: "With second specific end time",
			timing: TPTiming{
				EndTime: "12:00:00",
			},
			checkTime: time.Date(2024, time.January, 15, 10, 0, 0, 0, time.UTC),
			expected:  time.Date(2024, time.January, 15, 12, 0, 0, 0, time.UTC),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.timing.getRightMargin(tt.checkTime)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestIsTimeFormatedTrue(t *testing.T) {
	timeString := "12:12:12"
	if !IsTimeFormated(timeString) {
		t.Error("expected time to be formated, but returned false")
	}
}

func TestIsTimeFormatedFalse(t *testing.T) {
	timeString := "*any"
	if IsTimeFormated(timeString) {
		t.Error("expected invalid time format, but returned true")
	}
}

func TestTPDestinationClone(t *testing.T) {
	tests := []struct {
		name string
		tpd  *TPDestination
		want *TPDestination
	}{
		{
			name: "nil instance",
			tpd:  nil,
			want: nil,
		},
		{
			name: "empty prefixes",
			tpd: &TPDestination{
				TPid:     "TP1",
				ID:       "DEST1",
				Prefixes: nil,
			},
			want: &TPDestination{
				TPid:     "TP1",
				ID:       "DEST1",
				Prefixes: nil,
			},
		},
		{
			name: "with prefixes",
			tpd: &TPDestination{
				TPid:     "TP1",
				ID:       "DEST1",
				Prefixes: []string{"+49", "+1"},
			},
			want: &TPDestination{
				TPid:     "TP1",
				ID:       "DEST1",
				Prefixes: []string{"+49", "+1"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clone := tt.tpd.Clone()
			if !reflect.DeepEqual(clone, tt.want) {
				t.Errorf("TPDestination.Clone() = %v, want %v", clone, tt.want)
			}

			if tt.tpd != nil && tt.tpd.Prefixes != nil {
				if &tt.tpd.Prefixes[0] == &clone.Prefixes[0] {
					t.Errorf("TPDestination.Clone() did not perform a deep copy of Prefixes")
				}

				originalPrefixes := make([]string, len(tt.tpd.Prefixes))
				copy(originalPrefixes, tt.tpd.Prefixes)
				tt.tpd.Prefixes[0] = "modified"

				if !reflect.DeepEqual(clone.Prefixes, originalPrefixes) {
					t.Errorf("Clone's Prefixes changed when original was modified; clone: %v, original: %v",
						clone.Prefixes, originalPrefixes)
				}
			}
		})
	}
}

func TestTPRateRALsClone(t *testing.T) {
	tests := []struct {
		name string
		tpr  *TPRateRALs
		want *TPRateRALs
	}{
		{
			name: "nil instance",
			tpr:  nil,
			want: nil,
		},
		{
			name: "empty rate slots",
			tpr: &TPRateRALs{
				TPid:      "TP1",
				ID:        "RATE1",
				RateSlots: nil,
			},
			want: &TPRateRALs{
				TPid:      "TP1",
				ID:        "RATE1",
				RateSlots: nil,
			},
		},
		{
			name: "with rate slots",
			tpr: &TPRateRALs{
				TPid: "TP1",
				ID:   "RATE1",
				RateSlots: []*RateSlot{
					{
						ConnectFee:            0.5,
						Rate:                  0.01,
						RateUnit:              "60s",
						RateIncrement:         "1s",
						GroupIntervalStart:    "0s",
						rateUnitDur:           60 * time.Second,
						rateIncrementDur:      time.Second,
						groupIntervalStartDur: 0,
						tag:                   "test",
					},
				},
			},
			want: &TPRateRALs{
				TPid: "TP1",
				ID:   "RATE1",
				RateSlots: []*RateSlot{
					{
						ConnectFee:            0.5,
						Rate:                  0.01,
						RateUnit:              "60s",
						RateIncrement:         "1s",
						GroupIntervalStart:    "0s",
						rateUnitDur:           60 * time.Second,
						rateIncrementDur:      time.Second,
						groupIntervalStartDur: 0,
						tag:                   "test",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clone := tt.tpr.Clone()
			if !reflect.DeepEqual(clone, tt.want) {
				t.Errorf("TPRateRALs.Clone() = %v, want %v", clone, tt.want)
			}

			if tt.tpr != nil && tt.tpr.RateSlots != nil && len(tt.tpr.RateSlots) > 0 {
				if &tt.tpr.RateSlots[0] == &clone.RateSlots[0] {
					t.Errorf("TPRateRALs.Clone() did not perform a deep copy of RateSlots")
				}

				originalRate := tt.tpr.RateSlots[0].Rate
				originalConnectFee := tt.tpr.RateSlots[0].ConnectFee

				tt.tpr.RateSlots[0].Rate = 999.99
				tt.tpr.RateSlots[0].ConnectFee = 888.88

				if clone.RateSlots[0].Rate != originalRate || clone.RateSlots[0].ConnectFee != originalConnectFee {
					t.Errorf("Clone's RateSlots changed when original was modified; clone: %+v, original rates: %v, %v",
						clone.RateSlots[0], originalRate, originalConnectFee)
				}
			}
		})
	}
}

func TestTPRatingProfileClone(t *testing.T) {
	tests := []struct {
		name string
		rpf  *TPRatingProfile
		want *TPRatingProfile
	}{
		{
			name: "nil instance",
			rpf:  nil,
			want: nil,
		},
		{
			name: "empty rating plan activations",
			rpf: &TPRatingProfile{
				TPid:                  "TP1",
				LoadId:                "LOAD1",
				Tenant:                "cgrates.org",
				Category:              "call",
				Subject:               "1001",
				RatingPlanActivations: nil,
			},
			want: &TPRatingProfile{
				TPid:                  "TP1",
				LoadId:                "LOAD1",
				Tenant:                "cgrates.org",
				Category:              "call",
				Subject:               "1001",
				RatingPlanActivations: nil,
			},
		},
		{
			name: "with rating plan activations",
			rpf: &TPRatingProfile{
				TPid:     "TP1",
				LoadId:   "LOAD1",
				Tenant:   "cgrates.org",
				Category: "call",
				Subject:  "1001",
				RatingPlanActivations: []*TPRatingActivation{
					{
						ActivationTime:   "2022-01-01T00:00:00Z",
						RatingPlanId:     "RP_1001",
						FallbackSubjects: "1002;1003",
					},
					{
						ActivationTime:   "2022-02-01T00:00:00Z",
						RatingPlanId:     "RP_1002",
						FallbackSubjects: "1004;1005",
					},
				},
			},
			want: &TPRatingProfile{
				TPid:     "TP1",
				LoadId:   "LOAD1",
				Tenant:   "cgrates.org",
				Category: "call",
				Subject:  "1001",
				RatingPlanActivations: []*TPRatingActivation{
					{
						ActivationTime:   "2022-01-01T00:00:00Z",
						RatingPlanId:     "RP_1001",
						FallbackSubjects: "1002;1003",
					},
					{
						ActivationTime:   "2022-02-01T00:00:00Z",
						RatingPlanId:     "RP_1002",
						FallbackSubjects: "1004;1005",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clone := tt.rpf.Clone()
			if !reflect.DeepEqual(clone, tt.want) {
				t.Errorf("TPRatingProfile.Clone() = %v, want %v", clone, tt.want)
			}

			if tt.rpf != nil && tt.rpf.RatingPlanActivations != nil && len(tt.rpf.RatingPlanActivations) > 0 {
				if &tt.rpf.RatingPlanActivations[0] == &clone.RatingPlanActivations[0] {
					t.Errorf("TPRatingProfile.Clone() did not perform a deep copy of RatingPlanActivations")
				}

				originalActivationTime := tt.rpf.RatingPlanActivations[0].ActivationTime
				originalRatingPlanId := tt.rpf.RatingPlanActivations[0].RatingPlanId
				originalFallbackSubjects := tt.rpf.RatingPlanActivations[0].FallbackSubjects

				tt.rpf.RatingPlanActivations[0].ActivationTime = "2030-01-01T00:00:00Z"
				tt.rpf.RatingPlanActivations[0].RatingPlanId = "MODIFIED_RP"
				tt.rpf.RatingPlanActivations[0].FallbackSubjects = "modified"

				if clone.RatingPlanActivations[0].ActivationTime != originalActivationTime ||
					clone.RatingPlanActivations[0].RatingPlanId != originalRatingPlanId ||
					clone.RatingPlanActivations[0].FallbackSubjects != originalFallbackSubjects {
					t.Errorf("Clone's RatingPlanActivations changed when original was modified; clone: %+v, original values: %v, %v, %v",
						clone.RatingPlanActivations[0], originalActivationTime, originalRatingPlanId, originalFallbackSubjects)
				}
			}
		})
	}
}

func TestTPRateRALsCacheClone(t *testing.T) {
	tests := []struct {
		name string
		tpr  *TPRateRALs
		want *TPRateRALs
	}{
		{
			name: "Empty TPRateRALs",
			tpr:  &TPRateRALs{},
			want: &TPRateRALs{},
		},
		{
			name: "TPRateRALs with ID only",
			tpr:  &TPRateRALs{ID: "RT 1001"},
			want: &TPRateRALs{ID: "RT 1001"},
		},
		{
			name: "TPRateRALs with TPid only",
			tpr:  &TPRateRALs{TPid: "TP1"},
			want: &TPRateRALs{TPid: "TP1"},
		},
		{
			name: "TPRateRALs with empty RateSlots",
			tpr:  &TPRateRALs{TPid: "TP1", ID: "RT 1001", RateSlots: []*RateSlot{}},
			want: &TPRateRALs{TPid: "TP1", ID: "RT 1001", RateSlots: []*RateSlot{}},
		},
		{
			name: "TPRateRALs with single RateSlot",
			tpr: &TPRateRALs{
				TPid: "TP1",
				ID:   "RT 1001",
				RateSlots: []*RateSlot{
					{
						ConnectFee:         0.1,
						Rate:               0.2,
						RateUnit:           "60s",
						RateIncrement:      "1s",
						GroupIntervalStart: "0s",
						rateUnitDur:        60 * time.Second,
						rateIncrementDur:   time.Second,
						tag:                "test",
					},
				},
			},
			want: &TPRateRALs{
				TPid: "TP1",
				ID:   "RT 1001",
				RateSlots: []*RateSlot{
					{
						ConnectFee:         0.1,
						Rate:               0.2,
						RateUnit:           "60s",
						RateIncrement:      "1s",
						GroupIntervalStart: "0s",
						rateUnitDur:        60 * time.Second,
						rateIncrementDur:   time.Second,
						tag:                "test",
					},
				},
			},
		},
		{
			name: "Complete TPRateRALs with multiple RateSlots",
			tpr: &TPRateRALs{
				TPid: "TP1",
				ID:   "RT 1001",
				RateSlots: []*RateSlot{
					{
						ConnectFee:            0.1,
						Rate:                  0.2,
						RateUnit:              "60s",
						RateIncrement:         "1s",
						GroupIntervalStart:    "0s",
						rateUnitDur:           60 * time.Second,
						rateIncrementDur:      time.Second,
						groupIntervalStartDur: 0,
						tag:                   "first",
					},
					{
						ConnectFee:            0.0,
						Rate:                  0.1,
						RateUnit:              "60s",
						RateIncrement:         "30s",
						GroupIntervalStart:    "60s",
						rateUnitDur:           60 * time.Second,
						rateIncrementDur:      30 * time.Second,
						groupIntervalStartDur: 60 * time.Second,
						tag:                   "second",
					},
				},
			},
			want: &TPRateRALs{
				TPid: "TP1",
				ID:   "RT 1001",
				RateSlots: []*RateSlot{
					{
						ConnectFee:            0.1,
						Rate:                  0.2,
						RateUnit:              "60s",
						RateIncrement:         "1s",
						GroupIntervalStart:    "0s",
						rateUnitDur:           60 * time.Second,
						rateIncrementDur:      time.Second,
						groupIntervalStartDur: 0,
						tag:                   "first",
					},
					{
						ConnectFee:            0.0,
						Rate:                  0.1,
						RateUnit:              "60s",
						RateIncrement:         "30s",
						GroupIntervalStart:    "60s",
						rateUnitDur:           60 * time.Second,
						rateIncrementDur:      30 * time.Second,
						groupIntervalStartDur: 60 * time.Second,
						tag:                   "second",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.tpr.CacheClone()
			gotTPR, ok := got.(*TPRateRALs)
			if !ok {
				t.Errorf("CacheClone() returned type %T, want *TPRateRALs", got)
				return
			}

			if gotTPR.TPid != tt.want.TPid {
				t.Errorf("CacheClone().TPid = %v, want %v", gotTPR.TPid, tt.want.TPid)
			}

			if gotTPR.ID != tt.want.ID {
				t.Errorf("CacheClone().ID = %v, want %v", gotTPR.ID, tt.want.ID)
			}

			if len(gotTPR.RateSlots) != len(tt.want.RateSlots) {
				t.Errorf("CacheClone().RateSlots length = %v, want %v", len(gotTPR.RateSlots), len(tt.want.RateSlots))
				return
			}

			for i, rs := range tt.want.RateSlots {
				gotRS := gotTPR.RateSlots[i]

				if gotRS.ConnectFee != rs.ConnectFee {
					t.Errorf("CacheClone().RateSlots[%d].ConnectFee = %v, want %v", i, gotRS.ConnectFee, rs.ConnectFee)
				}

				if gotRS.Rate != rs.Rate {
					t.Errorf("CacheClone().RateSlots[%d].Rate = %v, want %v", i, gotRS.Rate, rs.Rate)
				}

				if gotRS.RateUnit != rs.RateUnit {
					t.Errorf("CacheClone().RateSlots[%d].RateUnit = %v, want %v", i, gotRS.RateUnit, rs.RateUnit)
				}

				if gotRS.RateIncrement != rs.RateIncrement {
					t.Errorf("CacheClone().RateSlots[%d].RateIncrement = %v, want %v", i, gotRS.RateIncrement, rs.RateIncrement)
				}

				if gotRS.GroupIntervalStart != rs.GroupIntervalStart {
					t.Errorf("CacheClone().RateSlots[%d].GroupIntervalStart = %v, want %v", i, gotRS.GroupIntervalStart, rs.GroupIntervalStart)
				}

				if gotRS.rateUnitDur != rs.rateUnitDur {
					t.Errorf("CacheClone().RateSlots[%d].rateUnitDur = %v, want %v", i, gotRS.rateUnitDur, rs.rateUnitDur)
				}

				if gotRS.rateIncrementDur != rs.rateIncrementDur {
					t.Errorf("CacheClone().RateSlots[%d].rateIncrementDur = %v, want %v", i, gotRS.rateIncrementDur, rs.rateIncrementDur)
				}

				if gotRS.groupIntervalStartDur != rs.groupIntervalStartDur {
					t.Errorf("CacheClone().RateSlots[%d].groupIntervalStartDur = %v, want %v", i, gotRS.groupIntervalStartDur, rs.groupIntervalStartDur)
				}

				if gotRS.tag != rs.tag {
					t.Errorf("CacheClone().RateSlots[%d].tag = %v, want %v", i, gotRS.tag, rs.tag)
				}
			}

			if tt.tpr.RateSlots != nil && len(tt.tpr.RateSlots) > 0 {
				originalRate := tt.tpr.RateSlots[0].Rate
				tt.tpr.RateSlots[0].Rate = 999.99

				if gotTPR.RateSlots[0].Rate != originalRate {
					t.Errorf("CacheClone() did not create a deep copy of RateSlots")
				}

				tt.tpr.RateSlots[0].Rate = originalRate
			}
		})
	}
}

func TestTPRankingProfileClone(t *testing.T) {
	tests := []struct {
		name string
		trp  TPRankingProfile
		want *TPRankingProfile
	}{
		{
			name: "Test with empty profile",
			trp:  TPRankingProfile{},
			want: &TPRankingProfile{},
		},
		{
			name: "Test with fully populated profile",
			trp: TPRankingProfile{
				TPid:              "TPR1",
				Tenant:            "cgrates.org",
				ID:                "profile1",
				Schedule:          "* * * * *",
				StatIDs:           []string{"stat1", "stat2"},
				MetricIDs:         []string{"metric1", "metric2"},
				Sorting:           "asc",
				SortingParameters: []string{"param1", "param2"},
				Stored:            true,
				ThresholdIDs:      []string{"threshold1", "threshold2"},
			},
			want: &TPRankingProfile{
				TPid:              "TPR1",
				Tenant:            "cgrates.org",
				ID:                "profile1",
				Schedule:          "* * * * *",
				StatIDs:           []string{"stat1", "stat2"},
				MetricIDs:         []string{"metric1", "metric2"},
				Sorting:           "asc",
				SortingParameters: []string{"param1", "param2"},
				Stored:            true,
				ThresholdIDs:      []string{"threshold1", "threshold2"},
			},
		},
		{
			name: "Test with some nil slices",
			trp: TPRankingProfile{
				TPid:              "TPR2",
				Tenant:            "cgrates2.org",
				ID:                "profile2",
				Schedule:          "* * * * *",
				StatIDs:           nil,
				MetricIDs:         []string{"metric1"},
				Sorting:           "desc",
				SortingParameters: nil,
				Stored:            false,
				ThresholdIDs:      []string{"threshold1"},
			},
			want: &TPRankingProfile{
				TPid:              "TPR2",
				Tenant:            "cgrates2.org",
				ID:                "profile2",
				Schedule:          "* * * * *",
				StatIDs:           nil,
				MetricIDs:         []string{"metric1"},
				Sorting:           "desc",
				SortingParameters: nil,
				Stored:            false,
				ThresholdIDs:      []string{"threshold1"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.trp.Clone()

			if reflect.TypeOf(got) != reflect.TypeOf(tt.want) {
				t.Errorf("Clone() returned wrong type: got %T, want %T", got, tt.want)
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Clone() = %v, want %v", got, tt.want)
			}

			if tt.trp.StatIDs != nil && len(tt.trp.StatIDs) > 0 {
				originalSlice := tt.trp.StatIDs
				clonedSlice := got.StatIDs
				clonedSlice[0] = "modified"
				if originalSlice[0] == clonedSlice[0] {
					t.Errorf("Clone() did not create a deep copy of StatIDs")
				}
			}

			if tt.trp.MetricIDs != nil && len(tt.trp.MetricIDs) > 0 {
				originalSlice := tt.trp.MetricIDs
				clonedSlice := got.MetricIDs
				clonedSlice[0] = "modified"
				if originalSlice[0] == clonedSlice[0] {
					t.Errorf("Clone() did not create a deep copy of MetricIDs")
				}
			}

			if tt.trp.SortingParameters != nil && len(tt.trp.SortingParameters) > 0 {
				originalSlice := tt.trp.SortingParameters
				clonedSlice := got.SortingParameters
				clonedSlice[0] = "modified"
				if originalSlice[0] == clonedSlice[0] {
					t.Errorf("Clone() did not create a deep copy of SortingParameters")
				}
			}

			if tt.trp.ThresholdIDs != nil && len(tt.trp.ThresholdIDs) > 0 {
				originalSlice := tt.trp.ThresholdIDs
				clonedSlice := got.ThresholdIDs
				clonedSlice[0] = "modified"
				if originalSlice[0] == clonedSlice[0] {
					t.Errorf("Clone() did not create a deep copy of ThresholdIDs")
				}
			}
		})
	}
	t.Run("Test with nil receiver", func(t *testing.T) {
		var trp *TPRankingProfile = nil
		got := trp.Clone()
		if got != nil {
			t.Errorf("Clone() with nil receiver = %v, want nil", got)
		}
	})
}

func TestTPDestinationCacheClone(t *testing.T) {
	tests := []struct {
		name string
		tpd  *TPDestination
		want *TPDestination
	}{
		{
			name: "Empty TPDestination",
			tpd:  &TPDestination{},
			want: &TPDestination{},
		},
		{
			name: "TPDestination with ID only",
			tpd:  &TPDestination{ID: "DST 1001"},
			want: &TPDestination{ID: "DST 1001"},
		},
		{
			name: "TPDestination with TPid only",
			tpd:  &TPDestination{TPid: "TP1"},
			want: &TPDestination{TPid: "TP1"},
		},
		{
			name: "TPDestination with empty Prefixes",
			tpd:  &TPDestination{TPid: "TP1", ID: "DST 1001", Prefixes: []string{}},
			want: &TPDestination{TPid: "TP1", ID: "DST 1001", Prefixes: []string{}},
		},
		{
			name: "TPDestination with single Prefix",
			tpd:  &TPDestination{TPid: "TP1", ID: "DST 1001", Prefixes: []string{"49"}},
			want: &TPDestination{TPid: "TP1", ID: "DST 1001", Prefixes: []string{"49"}},
		},
		{
			name: "Complete TPDestination with multiple Prefixes",
			tpd:  &TPDestination{TPid: "TP1", ID: "DST 1001", Prefixes: []string{"49", "41", "43"}},
			want: &TPDestination{TPid: "TP1", ID: "DST 1001", Prefixes: []string{"49", "41", "43"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.tpd.CacheClone()

			gotTPD, ok := got.(*TPDestination)
			if !ok {
				t.Errorf("CacheClone() returned type %T, want *TPDestination", got)
				return
			}

			if gotTPD.TPid != tt.want.TPid {
				t.Errorf("CacheClone().TPid = %v, want %v", gotTPD.TPid, tt.want.TPid)
			}

			if gotTPD.ID != tt.want.ID {
				t.Errorf("CacheClone().ID = %v, want %v", gotTPD.ID, tt.want.ID)
			}

			if len(gotTPD.Prefixes) != len(tt.want.Prefixes) {
				t.Errorf("CacheClone().Prefixes length = %v, want %v", len(gotTPD.Prefixes), len(tt.want.Prefixes))
				return
			}

			for i, prefix := range tt.want.Prefixes {
				if gotTPD.Prefixes[i] != prefix {
					t.Errorf("CacheClone().Prefixes[%d] = %v, want %v", i, gotTPD.Prefixes[i], prefix)
				}
			}
			if tt.tpd.Prefixes != nil && len(tt.tpd.Prefixes) > 0 {
				originalPrefix := tt.tpd.Prefixes[0]
				tt.tpd.Prefixes[0] = "modified"

				if gotTPD.Prefixes[0] != originalPrefix {
					t.Errorf("CacheClone() did not create a deep copy of Prefixes slice")
				}
				tt.tpd.Prefixes[0] = originalPrefix
			}
		})
	}
}

func TestTPDestinationRateClone(t *testing.T) {
	tests := []struct {
		name string
		tpdr *TPDestinationRate
		want *TPDestinationRate
	}{
		{
			name: "Nil TPDestinationRate",
			tpdr: nil,
			want: nil,
		},
		{
			name: "Empty TPDestinationRate",
			tpdr: &TPDestinationRate{},
			want: &TPDestinationRate{},
		},
		{
			name: "TPDestinationRate with ID only",
			tpdr: &TPDestinationRate{ID: "DR1001"},
			want: &TPDestinationRate{ID: "DR1001"},
		},
		{
			name: "TPDestinationRate with TPid only",
			tpdr: &TPDestinationRate{TPid: "TP1"},
			want: &TPDestinationRate{TPid: "TP1"},
		},
		{
			name: "TPDestinationRate with TPid and ID",
			tpdr: &TPDestinationRate{TPid: "TP1", ID: "DR1001"},
			want: &TPDestinationRate{TPid: "TP1", ID: "DR1001"},
		},
		{
			name: "TPDestinationRate with empty DestinationRates",
			tpdr: &TPDestinationRate{TPid: "TP1", ID: "DR1001", DestinationRates: []*DestinationRate{}},
			want: &TPDestinationRate{TPid: "TP1", ID: "DR1001", DestinationRates: []*DestinationRate{}},
		},
		{
			name: "TPDestinationRate with single DestinationRate",
			tpdr: &TPDestinationRate{
				TPid: "TP1",
				ID:   "DR1001",
				DestinationRates: []*DestinationRate{
					{
						DestinationId:    "DST1",
						RateId:           "RT1",
						RoundingMethod:   "*up",
						RoundingDecimals: 4,
						MaxCost:          0.60,
						MaxCostStrategy:  "*disconnect",
					},
				},
			},
			want: &TPDestinationRate{
				TPid: "TP1",
				ID:   "DR1001",
				DestinationRates: []*DestinationRate{
					{
						DestinationId:    "DST1",
						RateId:           "RT1",
						RoundingMethod:   "*up",
						RoundingDecimals: 4,
						MaxCost:          0.60,
						MaxCostStrategy:  "*disconnect",
					},
				},
			},
		},
		{
			name: "TPDestinationRate with multiple DestinationRates",
			tpdr: &TPDestinationRate{
				TPid: "TP1",
				ID:   "DR1001",
				DestinationRates: []*DestinationRate{
					{
						DestinationId:    "DST1",
						RateId:           "RT1",
						RoundingMethod:   "*up",
						RoundingDecimals: 4,
						MaxCost:          0.60,
						MaxCostStrategy:  "*disconnect",
					},
					{
						DestinationId:    "DST2",
						RateId:           "RT2",
						RoundingMethod:   "*down",
						RoundingDecimals: 2,
						MaxCost:          1.00,
						MaxCostStrategy:  "*disconnect",
					},
				},
			},
			want: &TPDestinationRate{
				TPid: "TP1",
				ID:   "DR1001",
				DestinationRates: []*DestinationRate{
					{
						DestinationId:    "DST1",
						RateId:           "RT1",
						RoundingMethod:   "*up",
						RoundingDecimals: 4,
						MaxCost:          0.60,
						MaxCostStrategy:  "*disconnect",
					},
					{
						DestinationId:    "DST2",
						RateId:           "RT2",
						RoundingMethod:   "*down",
						RoundingDecimals: 2,
						MaxCost:          1.00,
						MaxCostStrategy:  "*disconnect",
					},
				},
			},
		},
		{
			name: "TPDestinationRate with Rate field",
			tpdr: &TPDestinationRate{
				TPid: "TP1",
				ID:   "DR1001",
				DestinationRates: []*DestinationRate{
					{
						DestinationId:    "DST1",
						RateId:           "RT1",
						RoundingMethod:   "*up",
						RoundingDecimals: 4,
						MaxCost:          0.60,
						MaxCostStrategy:  "*disconnect",
						Rate: &TPRateRALs{
							TPid: "TP1",
							ID:   "RT1",
							RateSlots: []*RateSlot{
								{
									ConnectFee:         0.10,
									Rate:               0.05,
									RateUnit:           "60s",
									RateIncrement:      "1s",
									GroupIntervalStart: "0s",
								},
							},
						},
					},
				},
			},
			want: &TPDestinationRate{
				TPid: "TP1",
				ID:   "DR1001",
				DestinationRates: []*DestinationRate{
					{
						DestinationId:    "DST1",
						RateId:           "RT1",
						RoundingMethod:   "*up",
						RoundingDecimals: 4,
						MaxCost:          0.60,
						MaxCostStrategy:  "*disconnect",
						Rate: &TPRateRALs{
							TPid: "TP1",
							ID:   "RT1",
							RateSlots: []*RateSlot{
								{
									ConnectFee:         0.10,
									Rate:               0.05,
									RateUnit:           "60s",
									RateIncrement:      "1s",
									GroupIntervalStart: "0s",
								},
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.tpdr.Clone()

			if got == nil && tt.want == nil {
				return
			}

			if (got == nil && tt.want != nil) || (got != nil && tt.want == nil) {
				t.Errorf("Clone() = %v, want %v", got, tt.want)
				return
			}

			if got.TPid != tt.want.TPid {
				t.Errorf("Clone().TPid = %v, want %v", got.TPid, tt.want.TPid)
			}

			if got.ID != tt.want.ID {
				t.Errorf("Clone().ID = %v, want %v", got.ID, tt.want.ID)
			}

			if len(got.DestinationRates) != len(tt.want.DestinationRates) {
				t.Errorf("Clone().DestinationRates length = %v, want %v",
					len(got.DestinationRates), len(tt.want.DestinationRates))
				return
			}

			for i, destRate := range tt.want.DestinationRates {
				if got.DestinationRates[i].DestinationId != destRate.DestinationId {
					t.Errorf("Clone().DestinationRates[%d].DestinationId = %v, want %v",
						i, got.DestinationRates[i].DestinationId, destRate.DestinationId)
				}
				if got.DestinationRates[i].RateId != destRate.RateId {
					t.Errorf("Clone().DestinationRates[%d].RateId = %v, want %v",
						i, got.DestinationRates[i].RateId, destRate.RateId)
				}
				if got.DestinationRates[i].RoundingMethod != destRate.RoundingMethod {
					t.Errorf("Clone().DestinationRates[%d].RoundingMethod = %v, want %v",
						i, got.DestinationRates[i].RoundingMethod, destRate.RoundingMethod)
				}
				if got.DestinationRates[i].RoundingDecimals != destRate.RoundingDecimals {
					t.Errorf("Clone().DestinationRates[%d].RoundingDecimals = %v, want %v",
						i, got.DestinationRates[i].RoundingDecimals, destRate.RoundingDecimals)
				}
				if got.DestinationRates[i].MaxCost != destRate.MaxCost {
					t.Errorf("Clone().DestinationRates[%d].MaxCost = %v, want %v",
						i, got.DestinationRates[i].MaxCost, destRate.MaxCost)
				}
				if got.DestinationRates[i].MaxCostStrategy != destRate.MaxCostStrategy {
					t.Errorf("Clone().DestinationRates[%d].MaxCostStrategy = %v, want %v",
						i, got.DestinationRates[i].MaxCostStrategy, destRate.MaxCostStrategy)
				}

				if destRate.Rate != nil {
					if got.DestinationRates[i].Rate == nil {
						t.Errorf("Clone().DestinationRates[%d].Rate is nil, want non-nil", i)
						continue
					}

					if got.DestinationRates[i].Rate.TPid != destRate.Rate.TPid {
						t.Errorf("Clone().DestinationRates[%d].Rate.TPid = %v, want %v",
							i, got.DestinationRates[i].Rate.TPid, destRate.Rate.TPid)
					}

					if got.DestinationRates[i].Rate.ID != destRate.Rate.ID {
						t.Errorf("Clone().DestinationRates[%d].Rate.ID = %v, want %v",
							i, got.DestinationRates[i].Rate.ID, destRate.Rate.ID)
					}

					if len(got.DestinationRates[i].Rate.RateSlots) != len(destRate.Rate.RateSlots) {
						t.Errorf("Clone().DestinationRates[%d].Rate.RateSlots length = %v, want %v",
							i, len(got.DestinationRates[i].Rate.RateSlots), len(destRate.Rate.RateSlots))
						continue
					}

					for j, rateSlot := range destRate.Rate.RateSlots {
						gotSlot := got.DestinationRates[i].Rate.RateSlots[j]
						if !reflect.DeepEqual(gotSlot, rateSlot) {
							t.Errorf("Clone().DestinationRates[%d].Rate.RateSlots[%d] = %v, want %v",
								i, j, gotSlot, rateSlot)
						}
					}
				}
			}

			if tt.tpdr != nil && len(tt.tpdr.DestinationRates) > 0 {
				originalDestID := tt.tpdr.DestinationRates[0].DestinationId
				tt.tpdr.DestinationRates[0].DestinationId = "modified"

				if got.DestinationRates[0].DestinationId != originalDestID {
					t.Errorf("Clone() did not create a deep copy of DestinationRates")
				}

				tt.tpdr.DestinationRates[0].DestinationId = originalDestID

				if tt.tpdr.DestinationRates[0].Rate != nil &&
					len(tt.tpdr.DestinationRates[0].Rate.RateSlots) > 0 {

					originalRate := tt.tpdr.DestinationRates[0].Rate.RateSlots[0].Rate
					tt.tpdr.DestinationRates[0].Rate.RateSlots[0].Rate = 999.99

					if got.DestinationRates[0].Rate.RateSlots[0].Rate != originalRate {
						t.Errorf("Clone() did not create a deep copy of Rate.RateSlots")
					}
					tt.tpdr.DestinationRates[0].Rate.RateSlots[0].Rate = originalRate
				}
			}
		})
	}
}
