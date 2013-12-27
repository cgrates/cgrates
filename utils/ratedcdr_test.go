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
	"testing"
	"time"
	"reflect"
)

func TestRatedCDRInterfaces(t *testing.T) {
	ratedCdr := new(RatedCDR)
	var _ RawCDR = ratedCdr
}

func TestNewRatedCDRFromRawCDR(t *testing.T) {
	cgrCdr := CgrCdr{"accid": "dsafdsaf", "cdrhost": "192.168.1.1", "cdrsource": "internal_test", "reqtype": "rated", "direction": "*out", "tenant": "cgrates.org", "tor": "call",
		"account": "1001", "subject": "1001", "destination": "1002", "answer_time": "2013-11-07T08:42:26Z", "duration": "10",
		"field_extr1": "val_extr1", "fieldextr2": "valextr2"}
	expctRtCdr := &RatedCDR{CgrId: FSCgrId(cgrCdr["accid"]), AccId: cgrCdr["accid"], CdrHost: cgrCdr["cdrhost"], CdrSource: cgrCdr["cdrsource"], ReqType: cgrCdr["reqtype"], 
		Direction: cgrCdr["direction"], Tenant: cgrCdr["tenant"], TOR: cgrCdr["tor"], Account: cgrCdr["account"], Subject: cgrCdr["subject"],
		Destination: cgrCdr["destination"], AnswerTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC), Duration: time.Duration(10)*time.Second,
		ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"}, MediationRunId: DEFAULT_RUNID, Cost: -1}
	if rt, err := NewRatedCDRFromRawCDR(cgrCdr); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rt, expctRtCdr) {
		t.Errorf("Received %v, expected: %v", rt, expctRtCdr)
	}
}

func TestRatedCdrFields(t *testing.T) {
	ratedCdr := RatedCDR{CgrId: FSCgrId("dsafdsaf"), AccId: "dsafdsaf", CdrHost: "192.168.1.1", ReqType: "rated", Direction: "*out", Tenant: "cgrates.org",
		TOR: "call", Account: "1001", Subject: "1001", Destination: "1002", AnswerTime: time.Unix(1383813746, 0), Duration: 10,
		ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"}, Cost: 1.01,
		}
	if ratedCdr.GetCgrId() != "b18944ef4dc618569f24c27b9872827a242bad0c" {
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
