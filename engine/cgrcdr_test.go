/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

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
curl --data "OriginID=asbfdsaf&OriginHost=192.168.1.1&RequestType=rated&direction=*out&tenant=cgrates.org&tor=call&account=1001&subject=1001&destination=1002&time_answer=1383813746&duration=10&field_extr1=val_extr1&fieldextr2=valextr2" http://ipbxdev:2080/cgr
*/

func TestCgrCdrInterfaces(t *testing.T) {
	var _ RawCdr = make(CgrCdr)
}

func TestCgrCdrAsCDR(t *testing.T) {
	cgrCdr := CgrCdr{utils.ToR: utils.VOICE, utils.OriginID: "dsafdsaf",
		utils.OriginHost: "192.168.1.1", utils.Source: "internal_test",
		utils.RequestType: utils.META_RATED,
		utils.Tenant:      "cgrates.org", utils.Category: "call",
		utils.Account: "1001", utils.Subject: "1001", utils.Destination: "1002",
		utils.SetupTime: "2013-11-07T08:42:20Z", utils.AnswerTime: "2013-11-07T08:42:26Z",
		utils.Usage: "10s", "field_extr1": "val_extr1", "fieldextr2": "valextr2"}
	setupTime, _ := utils.ParseTimeDetectLayout(cgrCdr[utils.SetupTime], "")
	expctRtCdr := &CDR{CGRID: utils.Sha1(cgrCdr[utils.OriginID], setupTime.String()),
		ToR: utils.VOICE, OriginID: cgrCdr[utils.OriginID],
		OriginHost:  cgrCdr[utils.OriginHost],
		Source:      cgrCdr[utils.Source],
		RequestType: cgrCdr[utils.RequestType],
		Tenant:      cgrCdr[utils.Tenant], Category: cgrCdr[utils.Category],
		Account: cgrCdr[utils.Account], Subject: cgrCdr[utils.Subject],
		Destination: cgrCdr[utils.Destination],
		SetupTime:   time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC),
		AnswerTime:  time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC),
		Usage:       time.Duration(10) * time.Second, Cost: -1,
		ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"}}
	if CDR := cgrCdr.AsCDR(""); !reflect.DeepEqual(expctRtCdr, CDR) {
		t.Errorf("Expecting %v, received: %v", expctRtCdr, CDR)
	}
}

// Make sure the replicated CDR matches the expected CDR
func TestReplicatedCgrCdrAsCDR(t *testing.T) {
	cgrCdr := CgrCdr{
		utils.CGRID: "164b0422fdc6a5117031b427439482c6a4f90e41",
		utils.ToR:   utils.VOICE, utils.OriginID: "dsafdsaf",
		utils.OriginHost:  "192.168.1.1",
		utils.Source:      "internal_test",
		utils.RequestType: utils.META_RATED,
		utils.Tenant:      "cgrates.org", utils.Category: "call",
		utils.Account: "1001", utils.Subject: "1001",
		utils.Destination: "1002",
		utils.SetupTime:   "2013-11-07T08:42:20Z",
		utils.AnswerTime:  "2013-11-07T08:42:26Z",
		utils.Usage:       "10s", utils.COST: "0.12",
		utils.RATED: "true", "field_extr1": "val_extr1",
		"fieldextr2": "valextr2"}
	expctRtCdr := &CDR{
		CGRID:       cgrCdr[utils.CGRID],
		ToR:         cgrCdr[utils.ToR],
		OriginID:    cgrCdr[utils.OriginID],
		OriginHost:  cgrCdr[utils.OriginHost],
		Source:      cgrCdr[utils.Source],
		RequestType: cgrCdr[utils.RequestType],
		Tenant:      cgrCdr[utils.Tenant],
		Category:    cgrCdr[utils.Category],
		Account:     cgrCdr[utils.Account],
		Subject:     cgrCdr[utils.Subject],
		Destination: cgrCdr[utils.Destination],
		SetupTime:   time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC),
		AnswerTime:  time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC),
		Usage:       time.Duration(10) * time.Second,
		ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
		Cost:        0.12, PreRated: true,
	}
	if CDR := cgrCdr.AsCDR(""); !reflect.DeepEqual(expctRtCdr, CDR) {
		t.Errorf("Expecting %v, received: %v", expctRtCdr, CDR)
	}
}
