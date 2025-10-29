/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/
package migrator

import (
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestCdrsNewV1CDRFromCDRSql(t *testing.T) {
	testCdrSql := &engine.CDRsql{
		ID:          1,
		Cgrid:       "testID",
		RunID:       "testRunID",
		OriginHost:  "testOriginHost",
		Source:      "testSource",
		TOR:         "testTOR",
		RequestType: "testRequestType",
		Tenant:      "cgrates.org",
		Category:    "testCategory",
		Account:     "testAccount",
		Subject:     "testSubject",
		Destination: "testDestination",
		SetupTime:   time.Date(2021, 3, 3, 3, 3, 3, 3, time.UTC),
		AnswerTime:  utils.TimePointer(time.Date(2021, 3, 3, 3, 3, 3, 3, time.UTC)),
		Usage:       1,
		CostSource:  "testSource",
		Cost:        2,
		ExtraInfo:   "testExtraInfo",
		CreatedAt:   time.Date(2021, 3, 3, 3, 3, 3, 3, time.UTC),
		UpdatedAt:   time.Date(2021, 3, 3, 3, 3, 3, 3, time.UTC),
		DeletedAt:   utils.TimePointer(time.Date(2021, 3, 3, 3, 3, 3, 3, time.UTC)),
	}
	expected := &v1Cdrs{
		CGRID:       "testID",
		RunID:       "testRunID",
		OriginHost:  "testOriginHost",
		OrderID:     1,
		Source:      "testSource",
		ToR:         "testTOR",
		RequestType: "testRequestType",
		Tenant:      "cgrates.org",
		Category:    "testCategory",
		Account:     "testAccount",
		Subject:     "testSubject",
		Destination: "testDestination",
		SetupTime:   time.Date(2021, 3, 3, 3, 3, 3, 3, time.UTC),
		AnswerTime:  time.Date(2021, 3, 3, 3, 3, 3, 3, time.UTC),
		Usage:       time.Nanosecond,
		ExtraInfo:   "testExtraInfo",
		Partial:     false,
		Rated:       false,
		CostSource:  "testSource",
		Cost:        2,
	}
	result, err := NewV1CDRFromCDRSql(testCdrSql)
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(expected, result) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestCdrsAsCDRsql(t *testing.T) {
	testV1Cdrs := &v1Cdrs{
		CGRID:       "testID",
		RunID:       "testRunID",
		OriginHost:  "testOriginHost",
		OrderID:     1,
		Source:      "testSource",
		ToR:         "testTOR",
		RequestType: "testRequestType",
		Tenant:      "cgrates.org",
		Category:    "testCategory",
		Account:     "testAccount",
		Subject:     "testSubject",
		Destination: "testDestination",
		SetupTime:   time.Date(2021, 3, 3, 3, 3, 3, 3, time.UTC),
		AnswerTime:  time.Date(2021, 3, 3, 3, 3, 3, 3, time.UTC),
		Usage:       time.Nanosecond,
		ExtraInfo:   "testExtraInfo",
		Partial:     false,
		Rated:       false,
		CostSource:  "testSource",
		Cost:        2,
	}

	expected := &engine.CDRsql{
		ID:          0,
		Cgrid:       "testID",
		RunID:       "testRunID",
		OriginHost:  "testOriginHost",
		Source:      "testSource",
		TOR:         "testTOR",
		RequestType: "testRequestType",
		Tenant:      "cgrates.org",
		Category:    "testCategory",
		Account:     "testAccount",
		Subject:     "testSubject",
		Destination: "testDestination",
		SetupTime:   time.Date(2021, 3, 3, 3, 3, 3, 3, time.UTC),
		AnswerTime:  utils.TimePointer(time.Date(2021, 3, 3, 3, 3, 3, 3, time.UTC)),
		Usage:       1,
		CostSource:  "testSource",
		Cost:        2,
		ExtraInfo:   "testExtraInfo",
		CreatedAt:   time.Date(2021, 3, 3, 3, 3, 3, 3, time.UTC),
		ExtraFields: "",
		CostDetails: "",
	}
	result := testV1Cdrs.AsCDRsql()
	result.CreatedAt = time.Date(2021, 3, 3, 3, 3, 3, 3, time.UTC)
	result.ExtraFields = ""
	result.CostDetails = ""
	if !reflect.DeepEqual(expected, result) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", utils.ToJSON(expected), utils.ToJSON(result))
	}
}

func TestCdrsAsCDRsqlAnswertimeEmpty(t *testing.T) {
	var answTime time.Time
	testV1Cdrs := &v1Cdrs{
		AnswerTime: answTime,
	}

	expected := &engine.CDRsql{
		AnswerTime: nil,
		CreatedAt:  time.Date(2021, 3, 3, 3, 3, 3, 3, time.UTC),
	}
	result := testV1Cdrs.AsCDRsql()
	result.CreatedAt = time.Date(2021, 3, 3, 3, 3, 3, 3, time.UTC)
	result.ExtraFields = ""
	result.CostDetails = ""
	if !reflect.DeepEqual(expected, result) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", utils.ToJSON(expected), utils.ToJSON(result))
	}
}

func TestCdrsNewV1CDRFromCDRSqlAnswerTimeNil(t *testing.T) {
	testCdrSql := &engine.CDRsql{
		ID:          1,
		Cgrid:       "testID",
		RunID:       "testRunID",
		OriginHost:  "testOriginHost",
		Source:      "testSource",
		TOR:         "testTOR",
		RequestType: "testRequestType",
		Tenant:      "cgrates.org",
		Category:    "testCategory",
		Account:     "testAccount",
		Subject:     "testSubject",
		Destination: "testDestination",
		SetupTime:   time.Date(2021, 3, 3, 3, 3, 3, 3, time.UTC),
		AnswerTime:  nil,
		Usage:       1,
		CostSource:  "testSource",
		Cost:        2,
		ExtraInfo:   "testExtraInfo",
		CreatedAt:   time.Date(2021, 3, 3, 3, 3, 3, 3, time.UTC),
		UpdatedAt:   time.Date(2021, 3, 3, 3, 3, 3, 3, time.UTC),
		DeletedAt:   utils.TimePointer(time.Date(2021, 3, 3, 3, 3, 3, 3, time.UTC)),
	}
	var answTime time.Time
	expected := &v1Cdrs{
		CGRID:       "testID",
		RunID:       "testRunID",
		OriginHost:  "testOriginHost",
		OrderID:     1,
		Source:      "testSource",
		ToR:         "testTOR",
		RequestType: "testRequestType",
		Tenant:      "cgrates.org",
		Category:    "testCategory",
		Account:     "testAccount",
		Subject:     "testSubject",
		Destination: "testDestination",
		SetupTime:   time.Date(2021, 3, 3, 3, 3, 3, 3, time.UTC),
		AnswerTime:  answTime,
		Usage:       time.Nanosecond,
		ExtraInfo:   "testExtraInfo",
		Partial:     false,
		Rated:       false,
		CostSource:  "testSource",
		Cost:        2,
	}
	result, err := NewV1CDRFromCDRSql(testCdrSql)
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(expected, result) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestV1toV2Cdr(t *testing.T) {
	var testCallCost = &engine.CallCost{Cost: 10.0}

	v1Cdr := &v1Cdrs{
		CGRID:       "1234",
		RunID:       "1234",
		OrderID:     1,
		OriginHost:  "192.168.1.1",
		Source:      "TestSource",
		OriginID:    "12345",
		ToR:         "*voice",
		RequestType: "postpaid",
		Tenant:      "tenant1",
		Category:    "category1",
		Account:     "account1",
		Subject:     "subject1",
		Destination: "destination1",
		SetupTime:   time.Now(),
		AnswerTime:  time.Now().Add(5 * time.Minute),
		Usage:       5 * time.Minute,
		ExtraFields: map[string]string{"field1": "value1"},
		ExtraInfo:   "No errors",
		Partial:     false,
		Rated:       true,
		CostSource:  "TestSource",
		Cost:        15.0,
		CostDetails: testCallCost,
	}

	v2Cdr := v1Cdr.V1toV2Cdr()

	if v2Cdr.CGRID != v1Cdr.CGRID {
		t.Fatalf("CGRID mismatch: got %s, want %s", v2Cdr.CGRID, v1Cdr.CGRID)
	}
	if v2Cdr.RunID != v1Cdr.RunID {
		t.Fatalf("RunID mismatch: got %s, want %s", v2Cdr.RunID, v1Cdr.RunID)
	}
	if v2Cdr.OrderID != v1Cdr.OrderID {
		t.Fatalf("OrderID mismatch: got %d, want %d", v2Cdr.OrderID, v1Cdr.OrderID)
	}
	if v2Cdr.OriginHost != v1Cdr.OriginHost {
		t.Fatalf("OriginHost mismatch: got %s, want %s", v2Cdr.OriginHost, v1Cdr.OriginHost)
	}
	if v2Cdr.Source != v1Cdr.Source {
		t.Fatalf("Source mismatch: got %s, want %s", v2Cdr.Source, v1Cdr.Source)
	}
	if v2Cdr.OriginID != v1Cdr.OriginID {
		t.Fatalf("OriginID mismatch: got %s, want %s", v2Cdr.OriginID, v1Cdr.OriginID)
	}
	if v2Cdr.ToR != v1Cdr.ToR {
		t.Fatalf("ToR mismatch: got %s, want %s", v2Cdr.ToR, v1Cdr.ToR)
	}
	if v2Cdr.RequestType != v1Cdr.RequestType {
		t.Fatalf("RequestType mismatch: got %s, want %s", v2Cdr.RequestType, v1Cdr.RequestType)
	}
	if v2Cdr.Tenant != v1Cdr.Tenant {
		t.Fatalf("Tenant mismatch: got %s, want %s", v2Cdr.Tenant, v1Cdr.Tenant)
	}
	if v2Cdr.Category != v1Cdr.Category {
		t.Fatalf("Category mismatch: got %s, want %s", v2Cdr.Category, v1Cdr.Category)
	}
	if v2Cdr.Account != v1Cdr.Account {
		t.Fatalf("Account mismatch: got %s, want %s", v2Cdr.Account, v1Cdr.Account)
	}
	if v2Cdr.Subject != v1Cdr.Subject {
		t.Fatalf("Subject mismatch: got %s, want %s", v2Cdr.Subject, v1Cdr.Subject)
	}
	if v2Cdr.Destination != v1Cdr.Destination {
		t.Fatalf("Destination mismatch: got %s, want %s", v2Cdr.Destination, v1Cdr.Destination)
	}
	if !v2Cdr.SetupTime.Equal(v1Cdr.SetupTime) {
		t.Fatalf("SetupTime mismatch: got %v, want %v", v2Cdr.SetupTime, v1Cdr.SetupTime)
	}
	if !v2Cdr.AnswerTime.Equal(v1Cdr.AnswerTime) {
		t.Fatalf("AnswerTime mismatch: got %v, want %v", v2Cdr.AnswerTime, v1Cdr.AnswerTime)
	}
	if v2Cdr.Usage != v1Cdr.Usage {
		t.Fatalf("Usage mismatch: got %v, want %v", v2Cdr.Usage, v1Cdr.Usage)
	}
	if v2Cdr.ExtraFields["field1"] != v1Cdr.ExtraFields["field1"] {
		t.Fatalf("ExtraFields mismatch: got %s, want %s", v2Cdr.ExtraFields["field1"], v1Cdr.ExtraFields["field1"])
	}
	if v2Cdr.ExtraInfo != v1Cdr.ExtraInfo {
		t.Fatalf("ExtraInfo mismatch: got %s, want %s", v2Cdr.ExtraInfo, v1Cdr.ExtraInfo)
	}
	if v2Cdr.Partial != v1Cdr.Partial {
		t.Fatalf("Partial mismatch: got %t, want %t", v2Cdr.Partial, v1Cdr.Partial)
	}
	if v2Cdr.PreRated != v1Cdr.Rated {
		t.Fatalf("Rated mismatch: got %t, want %t", v2Cdr.PreRated, v1Cdr.Rated)
	}
	if v2Cdr.CostSource != v1Cdr.CostSource {
		t.Fatalf("CostSource mismatch: got %s, want %s", v2Cdr.CostSource, v1Cdr.CostSource)
	}
	if v2Cdr.Cost != v1Cdr.Cost {
		t.Fatalf("Cost mismatch: got %f, want %f", v2Cdr.Cost, v1Cdr.Cost)
	}

	if v2Cdr.CostDetails == nil {
		t.Fatalf("v2Cdr.CostDetails is nil")
	} else if v1Cdr.CostDetails == nil {
		t.Fatalf("v1Cdr.CostDetails is nil")
	}
}
