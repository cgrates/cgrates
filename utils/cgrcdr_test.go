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

func TestCgrCdrInterfaces(t *testing.T) {
	var _ RawCdr = make(CgrCdr)
}

func TestCgrCdrFields(t *testing.T) {
	cgrCdr := CgrCdr{ACCID: "dsafdsaf", CDRHOST: "192.168.1.1", REQTYPE: "rated", DIRECTION: "*out", TENANT: "cgrates.org", CATEGORY: "call",
		ACCOUNT: "1001", SUBJECT: "1001", DESTINATION: "1002", SETUP_TIME: "2013-11-07T08:42:20Z", ANSWER_TIME: "2013-11-07T08:42:26Z", DURATION: "10",
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
	if cgrCdr.GetCategory() != "call" {
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

func TestCgrCdrAsStoredCdr(t *testing.T) {
	cgrCdr := CgrCdr{ACCID: "dsafdsaf", CDRHOST: "192.168.1.1", CDRSOURCE: "internal_test", REQTYPE: "rated", DIRECTION: "*out", TENANT: "cgrates.org", CATEGORY: "call",
		ACCOUNT: "1001", SUBJECT: "1001", DESTINATION: "1002", SETUP_TIME: "2013-11-07T08:42:20Z", ANSWER_TIME: "2013-11-07T08:42:26Z", DURATION: "10",
		"field_extr1": "val_extr1", "fieldextr2": "valextr2"}
	setupTime, _ := ParseTimeDetectLayout(cgrCdr["setup_time"])
	expctRtCdr := &StoredCdr{CgrId: Sha1(cgrCdr["accid"], setupTime.String()), AccId: cgrCdr["accid"], CdrHost: cgrCdr["cdrhost"], CdrSource: cgrCdr["cdrsource"], ReqType: cgrCdr["reqtype"],
		Direction: cgrCdr[DIRECTION], Tenant: cgrCdr["tenant"], Category: cgrCdr[CATEGORY], Account: cgrCdr["account"], Subject: cgrCdr["subject"],
		Destination: cgrCdr["destination"], SetupTime: time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC), AnswerTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC), Duration: time.Duration(10) * time.Second,
		ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"}, Cost: -1}
	if storedCdr := cgrCdr.AsStoredCdr(); !reflect.DeepEqual(expctRtCdr, storedCdr) {
		t.Errorf("Expecting %v, received: %v", expctRtCdr, storedCdr)
	}
}
