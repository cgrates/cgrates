/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2012-2015 ITsysCOM

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

package engine

import (
	"github.com/cgrates/cgrates/utils"
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
	cgrCdr := CgrCdr{utils.TOR: utils.VOICE, utils.ACCID: "dsafdsaf", utils.CDRHOST: "192.168.1.1", utils.CDRSOURCE: "internal_test", utils.REQTYPE: utils.META_RATED,
		utils.DIRECTION: utils.OUT,
		utils.TENANT:    "cgrates.org", utils.CATEGORY: "call",
		utils.ACCOUNT: "1001", utils.SUBJECT: "1001", utils.DESTINATION: "1002", utils.SETUP_TIME: "2013-11-07T08:42:20Z", utils.ANSWER_TIME: "2013-11-07T08:42:26Z",
		utils.USAGE: "10", utils.SUPPLIER: "SUPPL1", "field_extr1": "val_extr1", "fieldextr2": "valextr2"}
	setupTime, _ := utils.ParseTimeDetectLayout(cgrCdr["setup_time"], "")
	expctRtCdr := &StoredCdr{CgrId: utils.Sha1(cgrCdr["accid"], setupTime.String()), TOR: utils.VOICE, AccId: cgrCdr["accid"], CdrHost: cgrCdr["cdrhost"], CdrSource: cgrCdr["cdrsource"],
		ReqType:   cgrCdr["reqtype"],
		Direction: cgrCdr[utils.DIRECTION], Tenant: cgrCdr["tenant"], Category: cgrCdr[utils.CATEGORY], Account: cgrCdr["account"], Subject: cgrCdr["subject"],
		Destination: cgrCdr["destination"], SetupTime: time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC), AnswerTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC),
		Usage: time.Duration(10) * time.Second, Supplier: "SUPPL1",
		ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"}, Cost: -1}
	if storedCdr := cgrCdr.AsStoredCdr(""); !reflect.DeepEqual(expctRtCdr, storedCdr) {
		t.Errorf("Expecting %v, received: %v", expctRtCdr, storedCdr)
	}
}

// Make sure the replicated CDR matches the expected StoredCdr
func TestReplicatedCgrCdrAsStoredCdr(t *testing.T) {
	cgrCdr := CgrCdr{utils.CGRID: "164b0422fdc6a5117031b427439482c6a4f90e41", utils.TOR: utils.VOICE, utils.ACCID: "dsafdsaf", utils.CDRHOST: "192.168.1.1",
		utils.CDRSOURCE: "internal_test", utils.REQTYPE: utils.META_RATED,
		utils.DIRECTION: utils.OUT, utils.TENANT: "cgrates.org", utils.CATEGORY: "call",
		utils.ACCOUNT: "1001", utils.SUBJECT: "1001", utils.DESTINATION: "1002", utils.SETUP_TIME: "2013-11-07T08:42:20Z", utils.PDD: "0.200", utils.ANSWER_TIME: "2013-11-07T08:42:26Z",
		utils.USAGE: "10", utils.SUPPLIER: "SUPPL1", utils.DISCONNECT_CAUSE: "NORMAL_CLEARING", utils.COST: "0.12", utils.RATED: "true", "field_extr1": "val_extr1", "fieldextr2": "valextr2"}
	expctRtCdr := &StoredCdr{CgrId: cgrCdr[utils.CGRID],
		TOR:             cgrCdr[utils.TOR],
		AccId:           cgrCdr[utils.ACCID],
		CdrHost:         cgrCdr[utils.CDRHOST],
		CdrSource:       cgrCdr[utils.CDRSOURCE],
		ReqType:         cgrCdr[utils.REQTYPE],
		Direction:       cgrCdr[utils.DIRECTION],
		Tenant:          cgrCdr[utils.TENANT],
		Category:        cgrCdr[utils.CATEGORY],
		Account:         cgrCdr[utils.ACCOUNT],
		Subject:         cgrCdr[utils.SUBJECT],
		Destination:     cgrCdr[utils.DESTINATION],
		SetupTime:       time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC),
		Pdd:             time.Duration(200) * time.Millisecond,
		AnswerTime:      time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC),
		Usage:           time.Duration(10) * time.Second,
		Supplier:        cgrCdr[utils.SUPPLIER],
		DisconnectCause: cgrCdr[utils.DISCONNECT_CAUSE],
		ExtraFields:     map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
		Cost:            0.12,
		Rated:           true,
	}
	if storedCdr := cgrCdr.AsStoredCdr(""); !reflect.DeepEqual(expctRtCdr, storedCdr) {
		t.Errorf("Expecting %v, received: %v", expctRtCdr, storedCdr)
	}
}
