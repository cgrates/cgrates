/*
Real-time Charging System for Telecom & ISP environments
Copyright (C) 2012-2014 ITsysCOM GmbH

This program is free software: you can Storagetribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITH*out ANY WARRANTY; without even the implied warranty of
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

func TestStoredCdrInterfaces(t *testing.T) {
	ratedCdr := new(StoredCdr)
	var _ RawCDR = ratedCdr
}

func TestNewStoredCdrFromRawCDR(t *testing.T) {
	cgrCdr := CgrCdr{"accid": "dsafdsaf", "cdrhost": "192.168.1.1", "cdrsource": "internal_test", "reqtype": "rated", "direction": "*out", "tenant": "cgrates.org", "tor": "call",
		"account": "1001", "subject": "1001", "destination": "1002", "setup_time": "2013-11-07T08:42:20Z", "answer_time": "2013-11-07T08:42:26Z", "duration": "10",
		"field_extr1": "val_extr1", "fieldextr2": "valextr2"}
	setupTime, _ := ParseTimeDetectLayout(cgrCdr["setup_time"])
	expctRtCdr := &StoredCdr{CgrId: Sha1(cgrCdr["accid"], setupTime.String()), AccId: cgrCdr["accid"], CdrHost: cgrCdr["cdrhost"], CdrSource: cgrCdr["cdrsource"], ReqType: cgrCdr["reqtype"],
		Direction: cgrCdr["direction"], Tenant: cgrCdr["tenant"], Category: cgrCdr["tor"], Account: cgrCdr["account"], Subject: cgrCdr["subject"],
		Destination: cgrCdr["destination"], SetupTime: time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC), AnswerTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC), Duration: time.Duration(10) * time.Second,
		ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"}, MediationRunId: DEFAULT_RUNID, Cost: -1}
	if rt, err := NewStoredCdrFromRawCDR(cgrCdr); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rt, expctRtCdr) {
		t.Errorf("Received %v, expected: %v", rt, expctRtCdr)
	}
}

func TestStoredCdrFields(t *testing.T) {
	ratedCdr := StoredCdr{CgrId: Sha1("dsafdsaf", time.Unix(1383813746, 0).String()), AccId: "dsafdsaf", CdrHost: "192.168.1.1", ReqType: "rated", Direction: "*out", Tenant: "cgrates.org",
		Category: "call", Account: "1001", Subject: "1001", Destination: "1002", SetupTime: time.Unix(1383813746, 0), AnswerTime: time.Unix(1383813746, 0), Duration: 10,
		ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"}, Cost: 1.01,
	}
	if ratedCdr.GetCgrId() != Sha1("dsafdsaf", time.Unix(1383813746, 0).String()) {
		t.Error("Error parsing cdr: ", ratedCdr)
	}
	if ratedCdr.GetAccId() != "dsafdsaf" {
		t.Error("Error parsing cdr: ", ratedCdr)
	}
	if ratedCdr.GetCdrHost() != "192.168.1.1" {
		t.Error("Error parsing cdr: ", ratedCdr)
	}
	if ratedCdr.GetDirection() != "*out" {
		t.Error("Error parsing cdr: ", ratedCdr)
	}
	if ratedCdr.GetSubject() != "1001" {
		t.Error("Error parsing cdr: ", ratedCdr)
	}
	if ratedCdr.GetAccount() != "1001" {
		t.Error("Error parsing cdr: ", ratedCdr)
	}
	if ratedCdr.GetDestination() != "1002" {
		t.Error("Error parsing cdr: ", ratedCdr)
	}
	if ratedCdr.GetTOR() != "call" {
		t.Error("Error parsing cdr: ", ratedCdr)
	}
	if ratedCdr.GetTenant() != "cgrates.org" {
		t.Error("Error parsing cdr: ", ratedCdr)
	}
	if ratedCdr.GetReqType() != RATED {
		t.Error("Error parsing cdr: ", ratedCdr)
	}
	setupTime, _ := ratedCdr.GetSetupTime()
	expectedSTime, _ := time.Parse(time.RFC3339, "2013-11-07T08:42:26Z")
	if setupTime.UTC() != expectedSTime {
		t.Error("Error parsing cdr: ", ratedCdr)
	}
	answerTime, _ := ratedCdr.GetAnswerTime()
	expectedATime, _ := time.Parse(time.RFC3339, "2013-11-07T08:42:26Z")
	if answerTime.UTC() != expectedATime {
		t.Error("Error parsing cdr: ", ratedCdr)
	}
	if ratedCdr.GetDuration() != 10 {
		t.Error("Error parsing cdr: ", ratedCdr)
	}
	extraFields := ratedCdr.GetExtraFields()
	if len(extraFields) != 2 {
		t.Error("Error parsing extra fields: ", extraFields)
	}
	if extraFields["field_extr1"] != "val_extr1" {
		t.Error("Error parsing extra fields: ", extraFields)
	}
	if ratedCdr.Cost != 1.01 {
		t.Error("Error parsing cdr: ", ratedCdr)
	}
}

func TestAsRawCdrHttpForm(t *testing.T) {
	ratedCdr := StoredCdr{CgrId: Sha1("dsafdsaf", time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC).String()), AccId: "dsafdsaf", CdrHost: "192.168.1.1", CdrSource: "test", ReqType: "rated", Direction: "*out", Tenant: "cgrates.org",
		Category: "call", Account: "1001", Subject: "1001", Destination: "1002", SetupTime: time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC), AnswerTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC),
		Duration: time.Duration(10) * time.Second, ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"}, Cost: 1.01,
	}
	cdrForm := ratedCdr.AsRawCdrHttpForm()
	if cdrForm.Get(ACCID) != ratedCdr.AccId {
		t.Errorf("Expected: %s, received: %s", ratedCdr.AccId, cdrForm.Get(ACCID))
	}
	if cdrForm.Get(CDRHOST) != ratedCdr.CdrHost {
		t.Errorf("Expected: %s, received: %s", ratedCdr.CdrHost, cdrForm.Get(CDRHOST))
	}
	if cdrForm.Get(CDRSOURCE) != ratedCdr.CdrSource {
		t.Errorf("Expected: %s, received: %s", ratedCdr.CdrSource, cdrForm.Get(CDRSOURCE))
	}
	if cdrForm.Get(REQTYPE) != ratedCdr.ReqType {
		t.Errorf("Expected: %s, received: %s", ratedCdr.ReqType, cdrForm.Get(REQTYPE))
	}
	if cdrForm.Get(DIRECTION) != ratedCdr.Direction {
		t.Errorf("Expected: %s, received: %s", ratedCdr.Direction, cdrForm.Get(DIRECTION))
	}
	if cdrForm.Get(TENANT) != ratedCdr.Tenant {
		t.Errorf("Expected: %s, received: %s", ratedCdr.Tenant, cdrForm.Get(TENANT))
	}
	if cdrForm.Get(Category) != ratedCdr.Category {
		t.Errorf("Expected: %s, received: %s", ratedCdr.Category, cdrForm.Get(Category))
	}
	if cdrForm.Get(ACCOUNT) != ratedCdr.Account {
		t.Errorf("Expected: %s, received: %s", ratedCdr.Account, cdrForm.Get(ACCOUNT))
	}
	if cdrForm.Get(SUBJECT) != ratedCdr.Subject {
		t.Errorf("Expected: %s, received: %s", ratedCdr.Subject, cdrForm.Get(SUBJECT))
	}
	if cdrForm.Get(DESTINATION) != ratedCdr.Destination {
		t.Errorf("Expected: %s, received: %s", ratedCdr.Destination, cdrForm.Get(DESTINATION))
	}
	if cdrForm.Get(SETUP_TIME) != "2013-11-07 08:42:20 +0000 UTC" {
		t.Errorf("Expected: %s, received: %s", "2013-11-07 08:42:26 +0000 UTC", cdrForm.Get(SETUP_TIME))
	}
	if cdrForm.Get(ANSWER_TIME) != "2013-11-07 08:42:26 +0000 UTC" {
		t.Errorf("Expected: %s, received: %s", "2013-11-07 08:42:26 +0000 UTC", cdrForm.Get(ANSWER_TIME))
	}
	if cdrForm.Get(DURATION) != "10" {
		t.Errorf("Expected: %s, received: %s", "10", cdrForm.Get(DURATION))
	}
	if cdrForm.Get("field_extr1") != ratedCdr.ExtraFields["field_extr1"] {
		t.Errorf("Expected: %s, received: %s", ratedCdr.ExtraFields["field_extr1"], cdrForm.Get("field_extr1"))
	}
	if cdrForm.Get("fieldextr2") != ratedCdr.ExtraFields["fieldextr2"] {
		t.Errorf("Expected: %s, received: %s", ratedCdr.ExtraFields["fieldextr2"], cdrForm.Get("fieldextr2"))
	}
}

func TestExportFieldValue(t *testing.T) {
	cdr := StoredCdr{CgrId: Sha1("dsafdsaf", time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC).String()), OrderId: 123, AccId: "dsafdsaf", CdrHost: "192.168.1.1", CdrSource: "test", ReqType: "rated", Direction: "*out", Tenant: "cgrates.org",
		Category: "call", Account: "1001", Subject: "1001", Destination: "1002", SetupTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC), AnswerTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC), MediationRunId: DEFAULT_RUNID,
		Duration: time.Duration(10) * time.Second, ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"}, Cost: 1.01,
	}
	if cdr.ExportFieldValue(CGRID) != cdr.CgrId ||
		cdr.ExportFieldValue(ORDERID) != "123" ||
		cdr.ExportFieldValue(ACCID) != cdr.AccId ||
		cdr.ExportFieldValue(CDRHOST) != cdr.CdrHost ||
		cdr.ExportFieldValue(CDRSOURCE) != cdr.CdrSource ||
		cdr.ExportFieldValue(REQTYPE) != cdr.ReqType ||
		cdr.ExportFieldValue(DIRECTION) != cdr.Direction ||
		cdr.ExportFieldValue(TENANT) != cdr.Tenant ||
		cdr.ExportFieldValue(Category) != cdr.Category ||
		cdr.ExportFieldValue(ACCOUNT) != cdr.Account ||
		cdr.ExportFieldValue(SUBJECT) != cdr.Subject ||
		cdr.ExportFieldValue(DESTINATION) != cdr.Destination ||
		cdr.ExportFieldValue(SETUP_TIME) != cdr.SetupTime.String() ||
		cdr.ExportFieldValue(ANSWER_TIME) != cdr.AnswerTime.String() ||
		cdr.ExportFieldValue(DURATION) != "10" ||
		cdr.ExportFieldValue(MEDI_RUNID) != cdr.MediationRunId ||
		cdr.ExportFieldValue(COST) != "1.01" ||
		cdr.ExportFieldValue("field_extr1") != cdr.ExtraFields["field_extr1"] ||
		cdr.ExportFieldValue("fieldextr2") != cdr.ExtraFields["fieldextr2"] ||
		cdr.ExportFieldValue("dummy_field") != "" {
		t.Error("Unexpected filed value received")
	}
}

func TestFormatCost(t *testing.T) {
	cdr := StoredCdr{Cost: 1.01}
	if cdr.FormatCost(0, 4) != "1.0100" {
		t.Error("Unexpected format of the cost: ", cdr.FormatCost(0, 4))
	}
	cdr = StoredCdr{Cost: 1.01001}
	if cdr.FormatCost(0, 4) != "1.0100" {
		t.Error("Unexpected format of the cost: ", cdr.FormatCost(0, 4))
	}
	if cdr.FormatCost(2, 0) != "101" {
		t.Error("Unexpected format of the cost: ", cdr.FormatCost(2, 0))
	}
	if cdr.FormatCost(1, 0) != "10" {
		t.Error("Unexpected format of the cost: ", cdr.FormatCost(1, 0))
	}
	if cdr.FormatCost(2, 3) != "101.001" {
		t.Error("Unexpected format of the cost: ", cdr.FormatCost(2, 3))
	}
}
