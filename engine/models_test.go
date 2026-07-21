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
		"originHost":   testCdrSql.OriginHost,
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
		"extraFields":  testCdrSql.ExtraFields,
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
