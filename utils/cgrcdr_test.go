/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2013 ITsysCOM

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
)

/*
curl --data "accid=asbfdsaf&cdrhost=192.168.1.1&reqtype=rated&direction=*out&tenant=cgrates.org&tor=call&account=1001&subject=1001&destination=1002&time_answer=1383813746&duration=10&field_extr1=val_extr1&fieldextr2=valextr2" http://ipbxdev:2080/cgr
*/

func TestCgrCdrFields(t *testing.T) {
	cgrCdr := CgrCdr{"accid": "dsafdsaf", "cdrhost": "192.168.1.1", "reqtype": "rated", "direction": "*out", "tenant": "cgrates.org", "tor": "call",
		"account": "1001", "subject": "1001", "destination": "1002", "setup_time": "2013-11-07T08:42:20Z", "answer_time": "2013-11-07T08:42:26Z", "duration": "10",
		"field_extr1": "val_extr1", "fieldextr2": "valextr2"}
	setupTime, _ := ParseTimeDetectLayout("2013-11-07T08:42:20Z")
	if cgrCdr.GetCgrId() != Sha1("dsafdsaf", setupTime.String()) {
		t.Error("Error parsing cdr: ", cgrCdr)
	}
	if cgrCdr.GetAccId() != "dsafdsaf" {
		t.Error("Error parsing cdr: ", cgrCdr)
	}
	if cgrCdr.GetCdrHost() != "192.168.1.1" {
		t.Error("Error parsing cdr: ", cgrCdr)
	}
	if cgrCdr.GetDirection() != "*out" {
		t.Error("Error parsing cdr: ", cgrCdr)
	}
	if cgrCdr.GetSubject() != "1001" {
		t.Error("Error parsing cdr: ", cgrCdr)
	}
	if cgrCdr.GetAccount() != "1001" {
		t.Error("Error parsing cdr: ", cgrCdr)
	}
	if cgrCdr.GetDestination() != "1002" {
		t.Error("Error parsing cdr: ", cgrCdr)
	}
	if cgrCdr.GetTOR() != "call" {
		t.Error("Error parsing cdr: ", cgrCdr)
	}
	if cgrCdr.GetTenant() != "cgrates.org" {
		t.Error("Error parsing cdr: ", cgrCdr)
	}
	if cgrCdr.GetReqType() != RATED {
		t.Error("Error parsing cdr: ", cgrCdr)
	}
	expectedSTime, _ := time.Parse(time.RFC3339, "2013-11-07T08:42:20Z")
	if setupTime.UTC() != expectedSTime {
		t.Error("Error parsing cdr: ", cgrCdr)
	}
	answerTime, _ := cgrCdr.GetAnswerTime()
	expectedATime, _ := time.Parse(time.RFC3339, "2013-11-07T08:42:26Z")
	if answerTime.UTC() != expectedATime {
		t.Error("Error parsing cdr: ", cgrCdr)
	}
	dur, _ := cgrCdr.GetDuration()
	if dur != time.Duration(10)*time.Second {
		t.Error("Error parsing cdr: ", cgrCdr)
	}
	extraFields := cgrCdr.GetExtraFields()
	if len(extraFields) != 2 {
		t.Error("Error parsing extra fields: ", extraFields)
	}
	if extraFields["field_extr1"] != "val_extr1" {
		t.Error("Error parsing extra fields: ", extraFields)
	}
}

func TestCgrCdrForkCdr(t *testing.T) {
	sampleCdr1 := &CgrCdr{"accid": "dsafdsaf", "cdrhost": "192.168.1.1", "cdrsource": "source_test", "reqtype": "rated", "direction": "*out", "tenant": "cgrates.org", "tor": "call",
		"account": "1001", "subject": "1001", "destination": "1002", "setup_time": "2013-11-07T08:42:24Z", "answer_time": "2013-11-07T08:42:26Z", "duration": "10",
		"field_extr1": "val_extr1", "fieldextr2": "valextr2"}
	rtSampleCdrOut, err := sampleCdr1.ForkCdr("sample_run1", "reqtype", "direction", "tenant", "tor", "account", "subject", "destination", "setup_time", "answer_time", "duration",
		[]string{}, true)
	if err != nil {
		t.Error("Unexpected error received", err)
	}
	setupTime1 := time.Date(2013, 11, 7, 8, 42, 24, 0, time.UTC)
	expctSplRatedCdr := &StoredCdr{CgrId: Sha1("dsafdsaf", setupTime1.String()), AccId: "dsafdsaf", CdrHost: "192.168.1.1", CdrSource: "source_test", ReqType: "rated",
		Direction: "*out", Tenant: "cgrates.org", TOR: "call", Account: "1001", Subject: "1001", Destination: "1002",
		SetupTime: setupTime1, AnswerTime: time.Unix(1383813746, 0).UTC(),
		Duration: 10000000000, ExtraFields: map[string]string{}, MediationRunId: "sample_run1", Cost: -1}
	if !reflect.DeepEqual(expctSplRatedCdr, rtSampleCdrOut) {
		t.Errorf("Expected: %v, received: %v", expctSplRatedCdr, rtSampleCdrOut)
	}
	cgrCdr := &CgrCdr{"accid": "dsafdsaf", "cdrhost": "192.168.1.1", "cdrsource": "source_test", "reqtype": "rated", "direction": "*out", "tenant": "cgrates.org", "tor": "call",
		"account": "1001", "subject": "1001", "destination": "1002", "setup_time": "2013-11-07T08:42:24Z", "answer_time": "2013-11-07T08:42:26Z", "duration": "10",
		"field_extr1": "val_extr1", "fieldextr2": "valextr2"}
	rtCdrOut, err := cgrCdr.ForkCdr("wholesale_run", "reqtype", "direction", "tenant", "tor", "account", "subject", "destination", "setup_time", "answer_time", "duration",
		[]string{"field_extr1", "fieldextr2"}, true)
	if err != nil {
		t.Error("Unexpected error received", err)
	}
	setupTime, _ := ParseTimeDetectLayout("2013-11-07T08:42:24Z")
	expctRatedCdr := &StoredCdr{CgrId: Sha1("dsafdsaf", setupTime.String()), AccId: "dsafdsaf", CdrHost: "192.168.1.1", CdrSource: "source_test", ReqType: "rated",
		Direction: "*out", Tenant: "cgrates.org", TOR: "call", Account: "1001", Subject: "1001", Destination: "1002",
		SetupTime: time.Unix(1383813744, 0).UTC(), AnswerTime: time.Unix(1383813746, 0).UTC(),
		Duration: 10000000000, ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"}, MediationRunId: "wholesale_run", Cost: -1}
	if !reflect.DeepEqual(rtCdrOut, expctRatedCdr) {
		t.Errorf("Received: %v, expected: %v", rtCdrOut, expctRatedCdr)
	}
	rtCdrOut2, err := cgrCdr.ForkCdr("wholesale_run", "^postpaid", "^*in", "^cgrates.com", "^premium_call", "^first_account", "^first_subject", "destination",
		"^2013-12-07T08:42:24Z", "^2013-12-07T08:42:26Z", "^12s", []string{"field_extr1", "fieldextr2"}, true)
	if err != nil {
		t.Error("Unexpected error received", err)
	}
	expctRatedCdr2 := &StoredCdr{CgrId: Sha1("dsafdsaf", time.Date(2013, 12, 7, 8, 42, 24, 0, time.UTC).String()), AccId: "dsafdsaf", CdrHost: "192.168.1.1", CdrSource: "source_test", ReqType: "postpaid",
		Direction: "*in", Tenant: "cgrates.com", TOR: "premium_call", Account: "first_account", Subject: "first_subject", Destination: "1002",
		SetupTime:  time.Date(2013, 12, 7, 8, 42, 24, 0, time.UTC),
		AnswerTime: time.Date(2013, 12, 7, 8, 42, 26, 0, time.UTC), Duration: time.Duration(12) * time.Second,
		ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"}, MediationRunId: "wholesale_run", Cost: -1}
	if !reflect.DeepEqual(rtCdrOut2, expctRatedCdr2) {
		t.Errorf("Received: %v, expected: %v", rtCdrOut2, expctRatedCdr2)
	}
	_, err = cgrCdr.ForkCdr("wholesale_run", "dummy_header", "direction", "tenant", "tor", "account", "subject", "destination", "setup_time", "answer_time", "duration",
		[]string{"field_extr1", "fieldextr2"}, true)
	if err == nil {
		t.Error("Failed to detect missing header")
	}
}

func TestCgrCdrForkCdrFromMetaDefaults(t *testing.T) {
	cgrCdr := &CgrCdr{"accid": "dsafdsaf", "cdrhost": "192.168.1.1", "cdrsource": "source_test", "reqtype": "rated", "direction": "*out", "tenant": "cgrates.org", "tor": "call",
		"account": "1001", "subject": "1001", "destination": "1002", "setup_time": "2013-11-07T08:42:24Z", "answer_time": "2013-11-07T08:42:26Z", "duration": "10",
		"field_extr1": "val_extr1", "fieldextr2": "valextr2"}
	setupTime := time.Date(2013, 11, 7, 8, 42, 24, 0, time.UTC)
	expctCdr := &StoredCdr{CgrId: Sha1("dsafdsaf", setupTime.String()), AccId: "dsafdsaf", CdrHost: "192.168.1.1", CdrSource: "source_test", ReqType: "rated",
		Direction: "*out", Tenant: "cgrates.org", TOR: "call", Account: "1001", Subject: "1001", Destination: "1002",
		SetupTime: setupTime, AnswerTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC),
		Duration:    time.Duration(10) * time.Second,
		ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"}, MediationRunId: "wholesale_run", Cost: -1}
	cdrOut, err := cgrCdr.ForkCdr("wholesale_run", META_DEFAULT, META_DEFAULT, META_DEFAULT, META_DEFAULT, META_DEFAULT, META_DEFAULT, META_DEFAULT,
		META_DEFAULT, META_DEFAULT, META_DEFAULT, []string{"field_extr1", "fieldextr2"}, true)
	if err != nil {
		t.Fatal("Unexpected error received", err)
	}

	if !reflect.DeepEqual(expctCdr, cdrOut) {
		t.Errorf("Expected: %v, received: %v", expctCdr, cdrOut)
	}
}
