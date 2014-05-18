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

func TestCgrCdrAsStoredCdr(t *testing.T) {
	cgrCdr := CgrCdr{TOR: VOICE, ACCID: "dsafdsaf", CDRHOST: "192.168.1.1", CDRSOURCE: "internal_test", REQTYPE: "rated", DIRECTION: "*out", TENANT: "cgrates.org", CATEGORY: "call",
		ACCOUNT: "1001", SUBJECT: "1001", DESTINATION: "1002", SETUP_TIME: "2013-11-07T08:42:20Z", ANSWER_TIME: "2013-11-07T08:42:26Z", USAGE: "10",
		"field_extr1": "val_extr1", "fieldextr2": "valextr2"}
	setupTime, _ := ParseTimeDetectLayout(cgrCdr["setup_time"])
	expctRtCdr := &StoredCdr{CgrId: Sha1(cgrCdr["accid"], setupTime.String()), TOR: VOICE, AccId: cgrCdr["accid"], CdrHost: cgrCdr["cdrhost"], CdrSource: cgrCdr["cdrsource"], ReqType: cgrCdr["reqtype"],
		Direction: cgrCdr[DIRECTION], Tenant: cgrCdr["tenant"], Category: cgrCdr[CATEGORY], Account: cgrCdr["account"], Subject: cgrCdr["subject"],
		Destination: cgrCdr["destination"], SetupTime: time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC), AnswerTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC), Duration: time.Duration(10) * time.Second,
		ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"}, Cost: -1}
	if storedCdr := cgrCdr.AsStoredCdr(); !reflect.DeepEqual(expctRtCdr, storedCdr) {
		t.Errorf("Expecting %v, received: %v", expctRtCdr, storedCdr)
	}
}
