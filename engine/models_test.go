/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or56
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>
*/

package engine

import (
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/utils"
)

func TestModelsTimingMdlTableName(t *testing.T) {
	testStruct := TimingMdl{}
	exp := utils.TBLTPTimings
	result := testStruct.TableName()
	if !reflect.DeepEqual(exp, result) {
		t.Errorf("\nExpected <%+v>,\nreceived <%+v>", exp, result)
	}
}

func TestModelsDestinationMdlTableName(t *testing.T) {
	testStruct := DestinationMdl{}
	exp := utils.TBLTPDestinations
	result := testStruct.TableName()
	if !reflect.DeepEqual(exp, result) {
		t.Errorf("\nExpected <%+v>,\nreceived <%+v>", exp, result)
	}
}

func TestModelsRateMdlTableName(t *testing.T) {
	testStruct := RateMdl{}
	exp := utils.TBLTPRates
	result := testStruct.TableName()
	if !reflect.DeepEqual(exp, result) {
		t.Errorf("\nExpected <%+v>,\nreceived <%+v>", exp, result)
	}
}

func TestModelsDestinationRateMdlTableName(t *testing.T) {
	testStruct := DestinationRateMdl{}
	exp := utils.TBLTPDestinationRates
	result := testStruct.TableName()
	if !reflect.DeepEqual(exp, result) {
		t.Errorf("\nExpected <%+v>,\nreceived <%+v>", exp, result)
	}
}

func TestModelsRatingPlanMdlTableName(t *testing.T) {
	testStruct := RatingPlanMdl{}
	exp := utils.TBLTPRatingPlans
	result := testStruct.TableName()
	if !reflect.DeepEqual(exp, result) {
		t.Errorf("\nExpected <%+v>,\nreceived <%+v>", exp, result)
	}
}

func TestModelsRatingProfileMdlTableName(t *testing.T) {
	testStruct := RatingProfileMdl{}
	exp := utils.TBLTPRatingProfiles
	result := testStruct.TableName()
	if !reflect.DeepEqual(exp, result) {
		t.Errorf("\nExpected <%+v>,\nreceived <%+v>", exp, result)
	}
}

func TestModelsActionMdlTableName(t *testing.T) {
	testStruct := ActionMdl{}
	exp := utils.TBLTPActions
	result := testStruct.TableName()
	if !reflect.DeepEqual(exp, result) {
		t.Errorf("\nExpected <%+v>,\nreceived <%+v>", exp, result)
	}
}

func TestModelsActionPlanMdlTableName(t *testing.T) {
	testStruct := ActionPlanMdl{}
	exp := utils.TBLTPActionPlans
	result := testStruct.TableName()
	if !reflect.DeepEqual(exp, result) {
		t.Errorf("\nExpected <%+v>,\nreceived <%+v>", exp, result)
	}
}

func TestModelsActionTriggerMdlTableName(t *testing.T) {
	testStruct := ActionTriggerMdl{}
	exp := utils.TBLTPActionTriggers
	result := testStruct.TableName()
	if !reflect.DeepEqual(exp, result) {
		t.Errorf("\nExpected <%+v>,\nreceived <%+v>", exp, result)
	}
}

func TestModelsAccountActionMdlTableName(t *testing.T) {
	testStruct := AccountActionMdl{}
	exp := utils.TBLTPAccountActions
	result := testStruct.TableName()
	if !reflect.DeepEqual(exp, result) {
		t.Errorf("\nExpected <%+v>,\nreceived <%+v>", exp, result)
	}
}

func TestModelsSetAccountActionId(t *testing.T) {
	testStruct := &AccountActionMdl{
		Id:      0,
		Loadid:  "",
		Tenant:  "",
		Account: "",
	}

	err := testStruct.SetAccountActionId("id1:id2:id3")
	if err != nil {
		t.Errorf("\nExpected <%+v>,\nreceived <%+v>", nil, err)
	}
	if !reflect.DeepEqual("id1", testStruct.Loadid) {
		t.Errorf("\nExpected <%+v>,\nreceived <%+v>", "id1", testStruct.Loadid)
	}
	if !reflect.DeepEqual("id2", testStruct.Tenant) {
		t.Errorf("\nExpected <%+v>,\nreceived <%+v>", "id2", testStruct.Tenant)
	}
	if !reflect.DeepEqual("id3", testStruct.Account) {
		t.Errorf("\nExpected <%+v>,\nreceived <%+v>", "id3", testStruct.Account)
	}

}

func TestModelsSetAccountActionIdError(t *testing.T) {
	testStruct := &AccountActionMdl{
		Id:      0,
		Loadid:  "",
		Tenant:  "",
		Account: "",
	}

	err := testStruct.SetAccountActionId("id1;id2;id3")
	if err == nil || err.Error() != "Wrong TP Account Action Id: id1;id2;id3" {
		t.Errorf("\nExpected <Wrong TP Account Action Id: id1;id2;id3>,\nreceived <%+v>", err)
	}

}

func TestModelGetAccountActionId(t *testing.T) {
	testStruct := &AccountActionMdl{
		Id:      0,
		Tpid:    "",
		Loadid:  "",
		Tenant:  "id1",
		Account: "id2",
	}
	result := testStruct.GetAccountActionId()
	exp := utils.ConcatenatedKey(testStruct.Tenant, testStruct.Account)
	if !reflect.DeepEqual(exp, result) {
		t.Errorf("\nExpected <%+v> ,\nreceived <%+v>", exp, result)
	}
}

func TestModelsAsMapStringInterface(t *testing.T) {
	testCdrSql := CDRsql{
		ID:          1,
		Cgrid:       "testCgrID1",
		RunID:       "testRunID",
		OriginHost:  "testOriginHost",
		Source:      "testSource",
		OriginID:    "testOriginId",
		TOR:         "testTOR",
		RequestType: "testRequestType",
		Tenant:      "cgrates.org",
		Category:    "testCategory",
		Account:     "testAccount",
		Subject:     "testSubject",
		Destination: "testDestination",
		SetupTime:   time.Date(2021, 3, 3, 3, 3, 3, 3, time.UTC),
		AnswerTime:  utils.TimePointer(time.Date(2021, 3, 3, 3, 3, 3, 3, time.UTC)),
		Usage:       2,
		ExtraFields: "extraFields",
		CostSource:  "testCostSource",
		Cost:        2,
		CostDetails: "testCostDetails",
		ExtraInfo:   "testExtraInfo",
		CreatedAt:   time.Date(2021, 3, 3, 3, 3, 3, 3, time.UTC),
		UpdatedAt:   time.Date(2021, 3, 3, 3, 3, 3, 3, time.UTC),
		DeletedAt:   utils.TimePointer(time.Date(2021, 3, 3, 3, 3, 3, 3, time.UTC)),
	}
	expected := map[string]any{
		"cgrid":        testCdrSql.Cgrid,
		"run_id":       testCdrSql.RunID,
		"origin_host":  testCdrSql.OriginHost,
		"source":       testCdrSql.Source,
		"origin_id":    testCdrSql.OriginID,
		"tor":          testCdrSql.TOR,
		"request_type": testCdrSql.RequestType,
		"tenant":       testCdrSql.Tenant,
		"category":     testCdrSql.Category,
		"account":      testCdrSql.Account,
		"subject":      testCdrSql.Subject,
		"destination":  testCdrSql.Destination,
		"setup_time":   testCdrSql.SetupTime,
		"answer_time":  testCdrSql.AnswerTime,
		"usage":        testCdrSql.Usage,
		"extra_fields": testCdrSql.ExtraFields,
		"cost_source":  testCdrSql.CostSource,
		"cost":         testCdrSql.Cost,
		"cost_details": testCdrSql.CostDetails,
		"extra_info":   testCdrSql.ExtraInfo,
		"created_at":   testCdrSql.CreatedAt,
		"updated_at":   testCdrSql.UpdatedAt,
	}
	result := testCdrSql.AsMapStringInterface()
	if !reflect.DeepEqual(expected, result) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestEngineSharedGroupMdlTableName(t *testing.T) {
	testStruct := SharedGroupMdl{}
	exp := utils.TBLTPSharedGroups
	result := testStruct.TableName()
	if !reflect.DeepEqual(exp, result) {
		t.Errorf("\nExpected: <%+v>\nReceived: <%+v>", exp, result)
	}
}

func TestEngineResourceMdlTableName(t *testing.T) {
	testStruct := ResourceMdl{}
	exp := utils.TBLTPResources
	result := testStruct.TableName()
	if !reflect.DeepEqual(exp, result) {
		t.Errorf("\nExpected: <%+v>\nReceived: <%+v>", exp, result)
	}
}

func TestEngineStatMdlTableName(t *testing.T) {
	testStruct := StatMdl{}
	exp := utils.TBLTPStats
	result := testStruct.TableName()
	if !reflect.DeepEqual(exp, result) {
		t.Errorf("\nExpected: <%+v>\nReceived: <%+v>", exp, result)
	}
}

func TestEngineThresholdMdlTableName(t *testing.T) {
	testStruct := ThresholdMdl{}
	exp := utils.TBLTPThresholds
	result := testStruct.TableName()
	if !reflect.DeepEqual(exp, result) {
		t.Errorf("\nExpected: <%+v>\nReceived: <%+v>", exp, result)
	}
}

func TestEngineTableNameTPfilters(t *testing.T) {
	want := "tp_filters"
	got := FilterMdl{}.TableName()
	if got != want {
		t.Errorf("Expected TableName(): %s, got: %s", want, got)
	}
}

func TestEngineTableNameCdrs(t *testing.T) {
	want := "cdrs"
	got := CDRsql{}.TableName()
	if got != want {
		t.Errorf("Expected TableName(): %s, got: %s", want, got)
	}
}

func TestEngineTableNameSessionCosts(t *testing.T) {
	want := "session_costs"
	got := SessionCostsSQL{}.TableName()
	if got != want {
		t.Errorf("Expected TableName(): %s, got: %s", want, got)
	}
}

func TestEngineTableNameTBLVersions(t *testing.T) {
	expected := "versions"
	got := TBLVersion{}.TableName()
	if got != expected {
		t.Errorf("Expected TableName(): %s, got: %s", expected, got)
	}
}

func TestEngineTableNameTbltpRoutes(t *testing.T) {
	want := "tp_routes"
	got := RouteMdl{}.TableName()
	if got != want {
		t.Errorf("Expected TableName(): %s, got: %s", want, got)
	}
}

func TestEngineTableNameTPAttributeMdl(t *testing.T) {
	want := "tp_attributes"
	got := AttributeMdl{}.TableName()
	if got != want {
		t.Errorf("Expected TableName(): %s, got: %s", want, got)
	}
}

func TestEngineTableNameChargerMdl(t *testing.T) {
	want := "tp_chargers"
	got := ChargerMdl{}.TableName()
	if got != want {
		t.Errorf("Expected TableName(): %s, got: %s", want, got)
	}
}

func TestEngineTableNameDispatcherProfileMdl(t *testing.T) {
	want := "tp_dispatcher_profiles"
	got := DispatcherProfileMdl{}.TableName()
	if got != want {
		t.Errorf("Expected TableName(): %s, got: %s", want, got)
	}
}

func TestEngineTableNameDispatcherHostMdl(t *testing.T) {
	want := "tp_dispatcher_hosts"
	got := DispatcherHostMdl{}.TableName()
	if got != want {
		t.Errorf("Expected TableName(): %s, got: %s", got, got)
	}
}

func TestTableName(t *testing.T) {
	expectedTableName := utils.TBLTPRankings
	rankingsMdl := &RankingsMdl{}
	actualTableName := rankingsMdl.TableName()
	if actualTableName != expectedTableName {
		t.Errorf("expected %s, but got %s", expectedTableName, actualTableName)
	}
}

func TestTrendsTableName(t *testing.T) {
	expectedTableName := utils.TBLTPTrends
	trendsMdl := &TrendsMdl{}
	actualTableName := trendsMdl.TableName()
	if actualTableName != expectedTableName {
		t.Errorf("expected %s, but got %s", expectedTableName, actualTableName)
	}
}
