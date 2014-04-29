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
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	"reflect"
	"testing"
	"time"
)

func TestParseFieldsConfig(t *testing.T) {
	// Test default config
	cgrConfig, _ := config.NewDefaultCGRConfig()
	// Test primary field index definition
	cgrConfig.CdrcAccIdField = "detect_me"
	cdrc := &Cdrc{cgrCfg: cgrConfig}
	if err := cdrc.parseFieldsConfig(); err == nil {
		t.Error("Failed detecting error in accounting id definition", err)
	}
	cgrConfig.CdrcAccIdField = "^static_val"
	cgrConfig.CdrcSubjectField = "1"
	cdrc = &Cdrc{cgrCfg: cgrConfig}
	if err := cdrc.parseFieldsConfig(); err != nil {
		t.Error("Failed to corectly parse primary fields %v", cdrc.cfgCdrFields)
	}
	cgrConfig.CdrcExtraFields = []string{"^static_val:orig_ip"}
	// Test extra field index definition
	cgrConfig.CdrcAccIdField = "0" // Put back as int
	cgrConfig.CdrcExtraFields = []string{"supplier1", "orig_ip:11"}
	cdrc = &Cdrc{cgrCfg: cgrConfig}
	if err := cdrc.parseFieldsConfig(); err == nil {
		t.Error("Failed detecting error in extra fields definition", err)
	}
	cgrConfig.CdrcExtraFields = []string{"supplier1:^top_supplier", "orig_ip:11"}
	cdrc = &Cdrc{cgrCfg: cgrConfig}
	if err := cdrc.parseFieldsConfig(); err != nil {
		t.Errorf("Failed to corectly parse extra fields %v", cdrc.cfgCdrFields)
	}
}

func TestRecordForkCdr(t *testing.T) {
	cgrConfig, _ := config.NewDefaultCGRConfig()
	cgrConfig.CdrcExtraFields = []string{"supplier:11"}
	cdrc := &Cdrc{cgrCfg: cgrConfig}
	if err := cdrc.parseFieldsConfig(); err != nil {
		t.Error("Failed parsing default fieldIndexesFromConfig", err)
	}
	cdrRow := []string{"firstField", "secondField"}
	_, err := cdrc.recordForkCdr(cdrRow)
	if err == nil {
		t.Error("Failed to corectly detect missing fields from record")
	}
	cdrRow = []string{"acc1", "prepaid", "*out", "cgrates.org", "call", "1001", "1001", "+4986517174963", "2013-02-03 19:50:00", "2013-02-03 19:54:00", "62",
		"supplier1", "172.16.1.1"}
	rtCdr, err := cdrc.recordForkCdr(cdrRow)
	if err != nil {
		t.Error("Failed to parse CDR in rated cdr", err)
	}
	expectedCdr := &utils.StoredCdr{
		CgrId:       utils.Sha1(cdrRow[0], time.Date(2013, 2, 3, 19, 50, 0, 0, time.UTC).String()),
		AccId:       cdrRow[0],
		CdrSource:   cgrConfig.CdrcSourceId,
		ReqType:     cdrRow[1],
		Direction:   cdrRow[2],
		Tenant:      cdrRow[3],
		Category:    cdrRow[4],
		Account:     cdrRow[5],
		Subject:     cdrRow[6],
		Destination: cdrRow[7],
		SetupTime:   time.Date(2013, 2, 3, 19, 50, 0, 0, time.UTC),
		AnswerTime:  time.Date(2013, 2, 3, 19, 54, 0, 0, time.UTC),
		Duration:    time.Duration(62) * time.Second,
		ExtraFields: map[string]string{"supplier": "supplier1"},
		Cost:        -1,
	}
	if !reflect.DeepEqual(expectedCdr, rtCdr) {
		t.Errorf("Expected: \n%v, \nreceived: \n%v", expectedCdr, rtCdr)
	}
	/*
		if cdrAsForm.Get(utils.CDRSOURCE) != cgrConfig.CdrcSourceId {
			t.Error("Unexpected cdrsource received", cdrAsForm.Get(utils.CDRSOURCE))
		}
		if cdrAsForm.Get(utils.REQTYPE) != "prepaid" {
			t.Error("Unexpected CDR value received", cdrAsForm.Get(utils.REQTYPE))
		}
		if cdrAsForm.Get("supplier") != "supplier1" {
			t.Error("Unexpected CDR value received", cdrAsForm.Get("supplier"))
		}
	*/
}
