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

/*
curl --data "accid=asbfdsaf&cdrhost=192.168.1.1&reqtype=rated&direction=*out&tenant=cgrates.org&tor=call&account=1001&subject=1001&destination=1002&time_answer=1383813746&duration=10&field_extr1=val_extr1&fieldextr2=valextr2" http://ipbxdev:2022/cgr
*/

func TestCgrCdrFields(t *testing.T) {
	cgrCdr := CgrCdr{"accid": "dsafdsaf", "cdrhost": "192.168.1.1", "reqtype": "rated", "direction": "*out", "tenant": "cgrates.org", "tor": "call",
		"account": "1001", "subject": "1001", "destination": "1002", "answer_time": "2013-11-07T08:42:26Z", "duration": "10", "field_extr1": "val_extr1", "fieldextr2": "valextr2"}
	if cgrCdr.GetCgrId() != FSCgrId("dsafdsaf") {
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
	answerTime, _ := cgrCdr.GetAnswerTime()
	expectedATime, _ := time.Parse(time.RFC3339, "2013-11-07T08:42:26Z")
	if answerTime.UTC() != expectedATime {
		t.Error("Error parsing cdr: ", cgrCdr)
	}
	if cgrCdr.GetDuration() != time.Duration(10)*time.Second {
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

func TestCgrCdrAsRatedCdr(t *testing.T) {
	cgrCdr := &CgrCdr{"accid": "dsafdsaf", "cdrhost": "192.168.1.1", "cdrsource": "source_test", "reqtype": "rated", "direction": "*out", "tenant": "cgrates.org", "tor": "call",
		"account": "1001", "subject": "1001", "destination": "1002", "answer_time": "2013-11-07T08:42:26Z", "duration": "10", 
		"field_extr1": "val_extr1", "fieldextr2": "valextr2"}
	rtCdrOut, err := cgrCdr.AsRatedCdr("wholesale_run", "reqtype", "direction", "tenant", "tor", "account", "subject", "destination", "answer_time", "duration", []string{"field_extr1","fieldextr2"}, true)
	if err != nil {
		t.Error("Unexpected error received", err)
	}
	expctRatedCdr := &RatedCDR{CgrId: FSCgrId("dsafdsaf"), AccId: "dsafdsaf", CdrHost: "192.168.1.1", CdrSource: "source_test", ReqType: "rated", 
		Direction: "*out", Tenant: "cgrates.org", TOR: "call", Account: "1001", Subject: "1001", Destination: "1002", AnswerTime: time.Unix(1383813746,0).UTC(),
		Duration: 10000000000, ExtraFields: map[string]string{"field_extr1":"val_extr1", "fieldextr2": "valextr2"}, MediationRunId:"wholesale_run", Cost: -1}
	if !reflect.DeepEqual(rtCdrOut, expctRatedCdr) {
		t.Errorf("Received: %v, expected: %v", rtCdrOut, expctRatedCdr)
	}
}


