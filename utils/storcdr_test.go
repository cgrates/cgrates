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
	"time"
	"testing"
)

func TestStorCDRInterfaces(t *testing.T) {
	storCdr := new(StorCDR)
	var _ CDR = storCdr
}

func TestStorCdrFields(t *testing.T) {
	storCdr := StorCDR{ CgrId: FSCgrId("dsafdsaf"), AccId:"dsafdsaf", CdrHost:"192.168.1.1", ReqType:"rated", Direction:"*out", Tenant:"cgrates.org",  
			TOR:"call", Account:"1001", Subject:"1001", Destination:"1002", AnswerTime:time.Unix(1383813746,0), Duration:10, 
			ExtraFields:map[string]string{"field_extr1":"val_extr1", "fieldextr2":"valextr2"},
			}
	if storCdr.GetCgrId() != "b18944ef4dc618569f24c27b9872827a242bad0c" {
		t.Error("Error parsing cdr: ", storCdr)
	}
	if storCdr.GetAccId() != "dsafdsaf" {
		t.Error("Error parsing cdr: ", storCdr)
	}
	if storCdr.GetCdrHost() != "192.168.1.1" {
		t.Error("Error parsing cdr: ", storCdr)
	}
	if storCdr.GetDirection() != "*out" {
		t.Error("Error parsing cdr: ", storCdr)
	}
	if storCdr.GetSubject() != "1001" {
		t.Error("Error parsing cdr: ", storCdr)
	}
	if storCdr.GetAccount() != "1001" {
		t.Error("Error parsing cdr: ", storCdr)
	}
	if storCdr.GetDestination() != "1002" {
		t.Error("Error parsing cdr: ", storCdr)
	}
	if storCdr.GetTOR() != "call" {
		t.Error("Error parsing cdr: ", storCdr)
	}
	if storCdr.GetTenant() != "cgrates.org" {
		t.Error("Error parsing cdr: ", storCdr)
	}
	if storCdr.GetReqType() != RATED {
		t.Error("Error parsing cdr: ", storCdr)
	}
	answerTime, _ := storCdr.GetAnswerTime()
	expectedATime, _ := time.Parse(time.RFC3339, "2013-11-07T08:42:26Z")
	if answerTime.UTC() != expectedATime {
		t.Error("Error parsing cdr: ", storCdr)
	}
	if storCdr.GetDuration() != 10 {
		t.Error("Error parsing cdr: ", storCdr)
	}
	extraFields := storCdr.GetExtraFields()
	if len(extraFields) != 2 {
		t.Error("Error parsing extra fields: ", extraFields)
	}
	if extraFields["field_extr1"] != "val_extr1" {
		t.Error("Error parsing extra fields: ", extraFields)
	}
}
