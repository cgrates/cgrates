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

package cdrc

import (
	"github.com/cgrates/cgrates/cdrs"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	"reflect"
	"testing"
	"time"
	"unicode/utf8"
)

func TestRecordForkCdr(t *testing.T) {
	cgrConfig, _ := config.NewDefaultCGRConfig()
	cgrConfig.CdrcCdrFields["supplier"] = &utils.RSRField{Id: "14"}
	csvSepRune, _ := utf8.DecodeRune([]byte(cgrConfig.CdrcCsvSep))
	cdrc := &Cdrc{cgrConfig.CdrcCdrs, cgrConfig.CdrcCdrType, cgrConfig.CdrcCdrInDir, cgrConfig.CdrcCdrOutDir, cgrConfig.CdrcSourceId, cgrConfig.CdrcRunDelay, csvSepRune,
		cgrConfig.CdrcCdrFields, new(cdrs.CDRS), nil}
	cdrRow := []string{"firstField", "secondField"}
	_, err := cdrc.recordForkCdr(cdrRow)
	if err == nil {
		t.Error("Failed to corectly detect missing fields from record")
	}
	cdrRow = []string{"ignored", "ignored", utils.VOICE, "acc1", "prepaid", "*out", "cgrates.org", "call", "1001", "1001", "+4986517174963",
		"2013-02-03 19:50:00", "2013-02-03 19:54:00", "62000000000", "supplier1", "172.16.1.1"}
	rtCdr, err := cdrc.recordForkCdr(cdrRow)
	if err != nil {
		t.Error("Failed to parse CDR in rated cdr", err)
	}
	expectedCdr := &utils.StoredCdr{
		CgrId:       utils.Sha1(cdrRow[3], time.Date(2013, 2, 3, 19, 50, 0, 0, time.UTC).String()),
		TOR:         cdrRow[2],
		AccId:       cdrRow[3],
		CdrHost:     "0.0.0.0", // Got it over internal interface
		CdrSource:   cgrConfig.CdrcSourceId,
		ReqType:     cdrRow[4],
		Direction:   cdrRow[5],
		Tenant:      cdrRow[6],
		Category:    cdrRow[7],
		Account:     cdrRow[8],
		Subject:     cdrRow[9],
		Destination: cdrRow[10],
		SetupTime:   time.Date(2013, 2, 3, 19, 50, 0, 0, time.UTC),
		AnswerTime:  time.Date(2013, 2, 3, 19, 54, 0, 0, time.UTC),
		Usage:       time.Duration(62) * time.Second,
		ExtraFields: map[string]string{"supplier": "supplier1"},
		Cost:        -1,
	}
	if !reflect.DeepEqual(expectedCdr, rtCdr) {
		t.Errorf("Expected: \n%v, \nreceived: \n%v", expectedCdr, rtCdr)
	}
}
