/*
Real-time Charging System for Telecom & ISP environments
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

package cdrc

import (
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestCsvRecordForkCdr(t *testing.T) {
	cgrConfig, _ := config.NewDefaultCGRConfig()
	cdrcConfig := cgrConfig.CdrcProfiles["/var/log/cgrates/cdrc/in"][utils.META_DEFAULT]
	cdrcConfig.CdrSourceId = "TEST_CDRC"
	cdrcConfig.ContentFields = append(cdrcConfig.ContentFields, &config.CfgCdrField{Tag: "SupplierTest", Type: utils.META_COMPOSED, FieldId: utils.SUPPLIER, Value: []*utils.RSRField{&utils.RSRField{Id: "14"}}})
	cdrcConfig.ContentFields = append(cdrcConfig.ContentFields, &config.CfgCdrField{Tag: "DisconnectCauseTest", Type: utils.META_COMPOSED, FieldId: utils.DISCONNECT_CAUSE,
		Value: []*utils.RSRField{&utils.RSRField{Id: "16"}}})
	//
	csvProcessor := &CsvRecordsProcessor{dfltCdrcCfg: cdrcConfig, cdrcCfgs: map[string]*config.CdrcConfig{"*default": cdrcConfig}}
	cdrRow := []string{"firstField", "secondField"}
	_, err := csvProcessor.recordToStoredCdr(cdrRow, "*default")
	if err == nil {
		t.Error("Failed to corectly detect missing fields from record")
	}
	cdrRow = []string{"ignored", "ignored", utils.VOICE, "acc1", utils.META_PREPAID, "*out", "cgrates.org", "call", "1001", "1001", "+4986517174963",
		"2013-02-03 19:50:00", "2013-02-03 19:54:00", "62", "supplier1", "172.16.1.1", "NORMAL_DISCONNECT"}
	rtCdr, err := csvProcessor.recordToStoredCdr(cdrRow, "*default")
	if err != nil {
		t.Error("Failed to parse CDR in rated cdr", err)
	}
	expectedCdr := &engine.StoredCdr{
		CgrId:           utils.Sha1(cdrRow[3], time.Date(2013, 2, 3, 19, 50, 0, 0, time.UTC).String()),
		TOR:             cdrRow[2],
		AccId:           cdrRow[3],
		CdrHost:         "0.0.0.0", // Got it over internal interface
		CdrSource:       "TEST_CDRC",
		ReqType:         cdrRow[4],
		Direction:       cdrRow[5],
		Tenant:          cdrRow[6],
		Category:        cdrRow[7],
		Account:         cdrRow[8],
		Subject:         cdrRow[9],
		Destination:     cdrRow[10],
		SetupTime:       time.Date(2013, 2, 3, 19, 50, 0, 0, time.UTC),
		AnswerTime:      time.Date(2013, 2, 3, 19, 54, 0, 0, time.UTC),
		Usage:           time.Duration(62) * time.Second,
		Supplier:        "supplier1",
		DisconnectCause: "NORMAL_DISCONNECT",
		ExtraFields:     map[string]string{},
		Cost:            -1,
	}
	if !reflect.DeepEqual(expectedCdr, rtCdr) {
		t.Errorf("Expected: \n%v, \nreceived: \n%v", expectedCdr, rtCdr)
	}
}

func TestCsvDataMultiplyFactor(t *testing.T) {
	cgrConfig, _ := config.NewDefaultCGRConfig()
	cdrcConfig := cgrConfig.CdrcProfiles["/var/log/cgrates/cdrc/in"][utils.META_DEFAULT]
	cdrcConfig.CdrSourceId = "TEST_CDRC"
	cdrcConfig.ContentFields = []*config.CfgCdrField{&config.CfgCdrField{Tag: "TORField", Type: utils.META_COMPOSED, FieldId: utils.TOR, Value: []*utils.RSRField{&utils.RSRField{Id: "0"}}},
		&config.CfgCdrField{Tag: "UsageField", Type: utils.META_COMPOSED, FieldId: utils.USAGE, Value: []*utils.RSRField{&utils.RSRField{Id: "1"}}}}
	csvProcessor := &CsvRecordsProcessor{dfltCdrcCfg: cdrcConfig, cdrcCfgs: map[string]*config.CdrcConfig{"*default": cdrcConfig}}
	csvProcessor.cdrcCfgs["*default"].DataUsageMultiplyFactor = 0
	cdrRow := []string{"*data", "1"}
	rtCdr, err := csvProcessor.recordToStoredCdr(cdrRow, "*default")
	if err != nil {
		t.Error("Failed to parse CDR in rated cdr", err)
	}
	var sTime time.Time
	expectedCdr := &engine.StoredCdr{
		CgrId:       utils.Sha1("", sTime.String()),
		TOR:         cdrRow[0],
		CdrHost:     "0.0.0.0",
		CdrSource:   "TEST_CDRC",
		Usage:       time.Duration(1) * time.Second,
		ExtraFields: map[string]string{},
		Cost:        -1,
	}
	if !reflect.DeepEqual(expectedCdr, rtCdr) {
		t.Errorf("Expected: \n%v, \nreceived: \n%v", expectedCdr, rtCdr)
	}
	csvProcessor.cdrcCfgs["*default"].DataUsageMultiplyFactor = 1024
	expectedCdr = &engine.StoredCdr{
		CgrId:       utils.Sha1("", sTime.String()),
		TOR:         cdrRow[0],
		CdrHost:     "0.0.0.0",
		CdrSource:   "TEST_CDRC",
		Usage:       time.Duration(1024) * time.Second,
		ExtraFields: map[string]string{},
		Cost:        -1,
	}
	if rtCdr, _ := csvProcessor.recordToStoredCdr(cdrRow, "*default"); !reflect.DeepEqual(expectedCdr, rtCdr) {
		t.Errorf("Expected: \n%v, \nreceived: \n%v", expectedCdr, rtCdr)
	}
	cdrRow = []string{"*voice", "1"}
	expectedCdr = &engine.StoredCdr{
		CgrId:       utils.Sha1("", sTime.String()),
		TOR:         cdrRow[0],
		CdrHost:     "0.0.0.0",
		CdrSource:   "TEST_CDRC",
		Usage:       time.Duration(1) * time.Second,
		ExtraFields: map[string]string{},
		Cost:        -1,
	}
	if rtCdr, _ := csvProcessor.recordToStoredCdr(cdrRow, "*default"); !reflect.DeepEqual(expectedCdr, rtCdr) {
		t.Errorf("Expected: \n%v, \nreceived: \n%v", expectedCdr, rtCdr)
	}
}

func TestCsvPairToRecord(t *testing.T) {
	eRecord := []string{"INVITE", "2daec40c", "548625ac", "dd0c4c617a9919d29a6175cdff223a9e@0:0:0:0:0:0:0:0", "200", "OK", "1436454408", "*prepaid", "1001", "1002", "", "3401:2069362475", "2"}
	invPr := &PartialFlatstoreRecord{Method: "INVITE", Timestamp: time.Date(2015, 7, 9, 15, 6, 48, 0, time.UTC),
		Values: []string{"INVITE", "2daec40c", "548625ac", "dd0c4c617a9919d29a6175cdff223a9e@0:0:0:0:0:0:0:0", "200", "OK", "1436454408", "*prepaid", "1001", "1002", "", "3401:2069362475"}}
	byePr := &PartialFlatstoreRecord{Method: "BYE", Timestamp: time.Date(2015, 7, 9, 15, 6, 50, 0, time.UTC),
		Values: []string{"BYE", "2daec40c", "548625ac", "dd0c4c617a9919d29a6175cdff223a9e@0:0:0:0:0:0:0:0", "200", "OK", "1436454410", "", "", "", "", "3401:2069362475"}}
	if rec, err := pairToRecord(invPr, byePr); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eRecord, rec) {
		t.Errorf("Expected: %+v, received: %+v", eRecord, rec)
	}
	if rec, err := pairToRecord(byePr, invPr); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eRecord, rec) {
		t.Errorf("Expected: %+v, received: %+v", eRecord, rec)
	}
	if _, err := pairToRecord(byePr, byePr); err == nil || err.Error() != "MISSING_INVITE" {
		t.Error(err)
	}
	if _, err := pairToRecord(invPr, invPr); err == nil || err.Error() != "MISSING_BYE" {
		t.Error(err)
	}
	byePr.Values = []string{"BYE", "2daec40c", "548625ac", "dd0c4c617a9919d29a6175cdff223a9e@0:0:0:0:0:0:0:0", "200", "OK", "1436454410", "", "", "", "3401:2069362475"} // Took one value out
	if _, err := pairToRecord(invPr, byePr); err == nil || err.Error() != "INCONSISTENT_VALUES_LENGTH" {
		t.Error(err)
	}
}
