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

func TestModelsAsMapStringInterface(t *testing.T) {
	testCdrSql := CDRsql{
		ID: 1,
		// Cgrid:       "testCgrID1",
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
		// "cgrid":        testCdrSql.Cgrid,
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

func TestResourceMdlTableName(t *testing.T) {
	model := ResourceMdl{
		PK:                2,
		Tpid:              "tpid",
		Tenant:            "tenant",
		ID:                "id",
		FilterIDs:         "fltrId",
		Weights:           "weights",
		UsageTTL:          "useageTTL",
		Limit:             "limit",
		AllocationMessage: "message",
		Blocker:           true,
		Stored:            true,
		ThresholdIDs:      "thresholdid",
		CreatedAt:         time.Now(),
	}
	rcv := model.TableName()
	if !reflect.DeepEqual(rcv, utils.TBLTPResources) {
		t.Errorf("Expected <%v>, Received <%v>", utils.TBLTPResources, rcv)
	}
}
func TestStatMdlTableName(t *testing.T) {
	model := StatMdl{
		PK:              2,
		Tpid:            "tpid",
		Tenant:          "tenant",
		ID:              "id",
		FilterIDs:       "fltrId",
		Weights:         "weights",
		Blockers:        "blockers",
		QueueLength:     2,
		TTL:             "TTL",
		MinItems:        3,
		Stored:          true,
		ThresholdIDs:    "thresholdid",
		MetricIDs:       "metricIds",
		MetricFilterIDs: "metricfltrIds",
		MetricBlockers:  "metricBlockrs",
		CreatedAt:       time.Now(),
	}
	rcv := model.TableName()
	if !reflect.DeepEqual(rcv, utils.TBLTPStats) {
		t.Errorf("Expected <%v>, Received <%v>", utils.TBLTPStats, rcv)
	}
}
func TestThresholdMdlTableName(t *testing.T) {
	model := ThresholdMdl{
		PK:               2,
		Tpid:             "tpid",
		Tenant:           "tenant",
		ID:               "id",
		FilterIDs:        "fltrId",
		Weights:          "weights",
		MaxHits:          44,
		MinHits:          1,
		MinSleep:         "2s",
		Blocker:          true,
		ActionProfileIDs: "actProfileIds",
		CreatedAt:        time.Now(),
	}
	rcv := model.TableName()
	if !reflect.DeepEqual(rcv, utils.TBLTPThresholds) {
		t.Errorf("Expected <%v>, Received <%v>", utils.TBLTPThresholds, rcv)
	}
}
func TestFilterMdlTableName(t *testing.T) {
	model := FilterMdl{
		PK:        2,
		Tpid:      "tpid",
		Tenant:    "tenant",
		ID:        "id",
		Type:      "type",
		Element:   "elm",
		Values:    "vals",
		CreatedAt: time.Now(),
	}
	rcv := model.TableName()
	if !reflect.DeepEqual(rcv, utils.TBLTPFilters) {
		t.Errorf("Expected <%v>, Received <%v>", utils.TBLTPFilters, rcv)
	}
}
func TestCDRsqlTableName(t *testing.T) {
	cdrSql := &CDRsql{
		ID:          1,
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
	rcv := cdrSql.TableName()
	if !reflect.DeepEqual(rcv, utils.CDRsTBL) {
		t.Errorf("Expected <%v>, Received <%v>", utils.CDRsTBL, rcv)
	}
}
func TestSessionCostsSQLTableName(t *testing.T) {
	sessCostSql := &SessionCostsSQL{
		ID:          1,
		RunID:       "testRunID",
		OriginHost:  "testOriginHost",
		OriginID:    "testOriginId",
		CostSource:  "testCostSource",
		Usage:       2,
		CostDetails: "testCostDetails",
		CreatedAt:   time.Date(2021, 3, 3, 3, 3, 3, 3, time.UTC),
		DeletedAt:   utils.TimePointer(time.Date(2021, 3, 3, 3, 3, 3, 3, time.UTC)),
	}
	rcv := sessCostSql.TableName()
	if !reflect.DeepEqual(rcv, utils.SessionCostsTBL) {
		t.Errorf("Expected <%v>, Received <%v>", utils.SessionCostsTBL, rcv)
	}
}
func TestTBLVersionTableName(t *testing.T) {
	tblVer := &TBLVersion{
		ID:      1,
		Item:    "testItem",
		Version: 14,
	}
	rcv := tblVer.TableName()
	if !reflect.DeepEqual(rcv, utils.TBLVersions) {
		t.Errorf("Expected <%v>, Received <%v>", utils.TBLVersions, rcv)
	}
}
func TestRouteMdlTableName(t *testing.T) {
	model := RouteMdl{
		PK:                  2,
		Tpid:                "tpid",
		Tenant:              "tenant",
		ID:                  "id",
		FilterIDs:           "fltrId",
		Weights:             "weights",
		Blockers:            "blockers",
		Sorting:             "testSorting",
		SortingParameters:   "testSortingParams",
		RouteID:             "testRouteId",
		RouteFilterIDs:      "testRouteFltrId",
		RouteAccountIDs:     "testRouteAccIds",
		RouteRateProfileIDs: "testRouteRateProfileIds",
		RouteResourceIDs:    "testRouteResrcIds",
		RouteStatIDs:        "testRouteStatIds",
		RouteWeights:        "testRouteWeights",
		RouteBlockers:       "testRouteBlckrs",
		RouteParameters:     "testRouteParams",
		CreatedAt:           time.Now(),
	}
	rcv := model.TableName()
	if !reflect.DeepEqual(rcv, utils.TBLTPRoutes) {
		t.Errorf("Expected <%v>, Received <%v>", utils.TBLTPRoutes, rcv)
	}
}
func TestAttributeMdlTableName(t *testing.T) {
	model := AttributeMdl{
		PK:                 2,
		Tpid:               "tpid",
		Tenant:             "tenant",
		ID:                 "id",
		FilterIDs:          "fltrId",
		Weights:            "weights",
		Blockers:           "blockers",
		AttributeFilterIDs: "testAttrFltrIds",
		AttributeBlockers:  "testAttrBlckrs",
		Path:               "testPath",
		Type:               "testType",
		Value:              "testValue",
		CreatedAt:          time.Now(),
	}
	rcv := model.TableName()
	if !reflect.DeepEqual(rcv, utils.TBLTPAttributes) {
		t.Errorf("Expected <%v>, Received <%v>", utils.TBLTPAttributes, rcv)
	}
}
func TestChargerMdlTableName(t *testing.T) {
	model := ChargerMdl{
		PK:           2,
		Tpid:         "tpid",
		Tenant:       "tenant",
		ID:           "id",
		FilterIDs:    "fltrId",
		Weights:      "weights",
		Blockers:     "blockers",
		RunID:        "testRunId",
		AttributeIDs: "testAttrIds",
		CreatedAt:    time.Now(),
	}
	rcv := model.TableName()
	if !reflect.DeepEqual(rcv, utils.TBLTPChargers) {
		t.Errorf("Expected <%v>, Received <%v>", utils.TBLTPChargers, rcv)
	}
}

func TestRateProfileMdlTableName(t *testing.T) {
	model := RateProfileMdl{
		PK:                  2,
		Tpid:                "tpid",
		Tenant:              "tenant",
		ID:                  "id",
		FilterIDs:           "testFiltrIds",
		Weights:             "testWeights",
		MinCost:             1,
		MaxCost:             99,
		MaxCostStrategy:     "testMaxCostStrat",
		RateID:              "testRateId",
		RateFilterIDs:       "testRateFltrIds",
		RateActivationTimes: "testRateActiveTime",
		RateWeights:         "testRateWeights",
		RateBlocker:         true,
		RateIntervalStart:   "testRateIntervalStart",
		RateFixedFee:        22,
		RateRecurrentFee:    55,
		RateUnit:            "testRateUnit",
		RateIncrement:       "testRateIncrement",
		CreatedAt:           time.Now(),
	}
	rcv := model.TableName()
	if !reflect.DeepEqual(rcv, utils.TBLTPRateProfiles) {
		t.Errorf("Expected <%v>, Received <%v>", utils.TBLTPRateProfiles, rcv)
	}
}
func TestActionProfileMdlTableName(t *testing.T) {
	model := ActionProfileMdl{
		PK:                     2,
		Tpid:                   "tpid",
		Tenant:                 "tenant",
		ID:                     "id",
		FilterIDs:              "testFiltrIds",
		Weights:                "testWeights",
		Blockers:               "testBlockers",
		Schedule:               "testSchedule",
		TargetType:             "testTargetType",
		TargetIDs:              "testTargetIds",
		ActionID:               "testActionId",
		ActionFilterIDs:        "testActionFltrIds",
		ActionTTL:              "testActionTTL",
		ActionType:             "testActionType",
		ActionOpts:             "testActionOpts",
		ActionWeights:          "testActionWeights",
		ActionBlockers:         "testActionBlockers",
		ActionDiktatsID:        "testActionDiktatsID",
		ActionDiktatsFilterIDs: "testActionDiktatsFilterIDs",
		ActionDiktatsOpts:      "testActionDiktatsOpts",
		ActionDiktatsWeights:   "testActionDiktatsWeights",
		ActionDiktatsBlockers:  "testActionDiktatsBlockers",
		CreatedAt:              time.Now(),
	}
	rcv := model.TableName()
	if !reflect.DeepEqual(rcv, utils.TBLTPActionProfiles) {
		t.Errorf("Expected <%v>, Received <%v>", utils.TBLTPActionProfiles, rcv)
	}
}
func TestAccountMdlTableName(t *testing.T) {
	model := AccountMdl{
		PK:                    2,
		Tpid:                  "tpid",
		Tenant:                "tenant",
		ID:                    "id",
		FilterIDs:             "testFiltrIds",
		Weights:               "testWeights",
		Blockers:              "testBlockers",
		Opts:                  "testOpts",
		BalanceID:             "testBalId",
		BalanceFilterIDs:      "testBalFltrIds",
		BalanceWeights:        "testBalWeights",
		BalanceBlockers:       "testBalBlockrs",
		BalanceType:           "testBalType",
		BalanceUnits:          "testBalUnits",
		BalanceUnitFactors:    "testBalUnitFactors",
		BalanceOpts:           "testBalOpts",
		BalanceCostIncrements: "testBalCostIncr",
		BalanceAttributeIDs:   "testBalAttrIds",
		BalanceRateProfileIDs: "testBalRateProfileIds",
		ThresholdIDs:          "testThresholdIds",
		CreatedAt:             time.Now(),
	}
	rcv := model.TableName()
	if !reflect.DeepEqual(rcv, utils.TBLTPAccounts) {
		t.Errorf("Expected <%v>, Received <%v>", utils.TBLTPAccounts, rcv)
	}
}
