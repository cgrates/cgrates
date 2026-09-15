// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package engine

import (
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/utils"
)

func TestModelsAsMapStringInterface(t *testing.T) {
	testUrSql := URsql{
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
		// "cgrid":        testUrSql.Cgrid,
		"run_id":       testUrSql.RunID,
		"originHost":   testUrSql.OriginHost,
		"source":       testUrSql.Source,
		"origin_id":    testUrSql.OriginID,
		"tor":          testUrSql.TOR,
		"request_type": testUrSql.RequestType,
		"tenant":       testUrSql.Tenant,
		"category":     testUrSql.Category,
		"account":      testUrSql.Account,
		"subject":      testUrSql.Subject,
		"destination":  testUrSql.Destination,
		"setup_time":   testUrSql.SetupTime,
		"answer_time":  testUrSql.AnswerTime,
		"usage":        testUrSql.Usage,
		"extraFields":  testUrSql.ExtraFields,
		"cost_source":  testUrSql.CostSource,
		"cost":         testUrSql.Cost,
		"cost_details": testUrSql.CostDetails,
		"extra_info":   testUrSql.ExtraInfo,
		"created_at":   testUrSql.CreatedAt,
		"updated_at":   testUrSql.UpdatedAt,
	}
	result := testUrSql.AsMapStringInterface()
	if !reflect.DeepEqual(expected, result) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestURsqlTableName(t *testing.T) {
	urSql := &URsql{
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
	rcv := urSql.TableName()
	if !reflect.DeepEqual(rcv, utils.URsTBL) {
		t.Errorf("Expected <%v>, Received <%v>", utils.URsTBL, rcv)
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
