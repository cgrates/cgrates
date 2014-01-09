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

package cdrexporter

import (
	"bytes"
	"github.com/cgrates/cgrates/utils"
	"strings"
	"testing"
	"time"
)

func TestCsvCdrWriter(t *testing.T) {
	writer := &bytes.Buffer{}
	csvCdrWriter := NewCsvCdrWriter(writer, 4, []string{"extra3", "extra1"})
	ratedCdr := &utils.RatedCDR{CgrId: utils.FSCgrId("dsafdsaf"), AccId: "dsafdsaf", CdrHost: "192.168.1.1", ReqType: "rated", Direction: "*out", Tenant: "cgrates.org",
		TOR: "call", Account: "1001", Subject: "1001", Destination: "1002", AnswerTime: time.Unix(1383813746, 0).UTC(), Duration: 10, MediationRunId: utils.DEFAULT_RUNID,
		ExtraFields: map[string]string{"extra1": "val_extra1", "extra2": "val_extra2", "extra3": "val_extra3"}, Cost: 1.01,
	}
	csvCdrWriter.Write(ratedCdr)
	csvCdrWriter.Close()
	expected := "b18944ef4dc618569f24c27b9872827a242bad0c,default,dsafdsaf,192.168.1.1,rated,*out,cgrates.org,call,1001,1001,1002,2013-11-07 08:42:26 +0000 UTC,10,1.0100,val_extra3,val_extra1"
	result := strings.TrimSpace(writer.String())
	if result != expected {
		t.Errorf("Expected %s received %s.", expected, result)
	}
}
