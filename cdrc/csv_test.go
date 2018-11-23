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
package cdrc

import (
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestCsvRecordToCDR(t *testing.T) {
	cgrConfig, _ := config.NewDefaultCGRConfig()
	cdrcConfig := cgrConfig.CdrcProfiles["/var/spool/cgrates/cdrc/in"][0]
	cdrcConfig.CdrSourceId = "TEST_CDRC"
	cdrcConfig.ContentFields = append(cdrcConfig.ContentFields, &config.FCTemplate{
		Tag: utils.RunID, Type: utils.META_COMPOSED, FieldId: utils.RunID,
		Value: config.NewRSRParsersMustCompile("*default", true, utils.INFIELD_SEP)})
	csvProcessor := &CsvRecordsProcessor{dfltCdrcCfg: cdrcConfig, cdrcCfgs: []*config.CdrcCfg{cdrcConfig}}
	cdrRow := []string{"firstField", "secondField"}
	_, err := csvProcessor.recordToStoredCdr(cdrRow, cdrcConfig, "cgrates.org")
	if err == nil {
		t.Error("Failed to corectly detect missing fields from record")
	}
	cdrRow = []string{"ignored", "ignored", utils.VOICE, "acc1", utils.META_PREPAID, "*out", "cgrates.org",
		"call", "1001", "1001", "+4986517174963", "2013-02-03 19:50:00", "2013-02-03 19:54:00",
		"62s", "supplier1", "172.16.1.1", "NORMAL_DISCONNECT"}
	rtCdr, err := csvProcessor.recordToStoredCdr(cdrRow, cdrcConfig, "cgrates.org")
	if err != nil {
		t.Error("Failed to parse CDR in rated cdr", err)
	}
	expectedCdr := &engine.CDR{
		CGRID:       utils.Sha1(cdrRow[3], time.Date(2013, 2, 3, 19, 50, 0, 0, time.UTC).String()),
		RunID:       "*default",
		ToR:         cdrRow[2],
		OriginID:    cdrRow[3],
		OriginHost:  "0.0.0.0", // Got it over internal interface
		Source:      "TEST_CDRC",
		RequestType: cdrRow[4],
		Tenant:      cdrRow[6],
		Category:    cdrRow[7],
		Account:     cdrRow[8],
		Subject:     cdrRow[9],
		Destination: cdrRow[10],
		SetupTime:   time.Date(2013, 2, 3, 19, 50, 0, 0, time.UTC),
		AnswerTime:  time.Date(2013, 2, 3, 19, 54, 0, 0, time.UTC),
		Usage:       time.Duration(62) * time.Second,
		ExtraFields: map[string]string{},
		Cost:        -1,
	}
	if !reflect.DeepEqual(expectedCdr, rtCdr) {
		t.Errorf("Expected: \n%v, \nreceived: \n%v", expectedCdr, rtCdr)
	}
}

func TestCsvDataMultiplyFactor(t *testing.T) {
	cgrConfig, _ := config.NewDefaultCGRConfig()
	cdrcConfig := cgrConfig.CdrcProfiles["/var/spool/cgrates/cdrc/in"][0]
	cdrcConfig.CdrSourceId = "TEST_CDRC"
	cdrcConfig.ContentFields = []*config.FCTemplate{
		{Tag: "TORField", Type: utils.META_COMPOSED, FieldId: utils.ToR,
			Value: config.NewRSRParsersMustCompile("~0", true, utils.INFIELD_SEP)},
		{Tag: "UsageField", Type: utils.META_COMPOSED, FieldId: utils.Usage,
			Value: config.NewRSRParsersMustCompile("~1", true, utils.INFIELD_SEP)},
	}
	csvProcessor := &CsvRecordsProcessor{dfltCdrcCfg: cdrcConfig, cdrcCfgs: []*config.CdrcCfg{cdrcConfig}}
	csvProcessor.cdrcCfgs[0].DataUsageMultiplyFactor = 0
	cdrRow := []string{"*data", "1"}
	rtCdr, err := csvProcessor.recordToStoredCdr(cdrRow, cdrcConfig, "cgrates.org")
	if err != nil {
		t.Error("Failed to parse CDR in rated cdr", err)
	}
	var sTime time.Time
	expectedCdr := &engine.CDR{
		CGRID:       utils.Sha1("", sTime.String()),
		ToR:         cdrRow[0],
		OriginHost:  "0.0.0.0",
		Source:      "TEST_CDRC",
		Usage:       time.Duration(1),
		ExtraFields: map[string]string{},
		Cost:        -1,
	}
	if !reflect.DeepEqual(expectedCdr, rtCdr) {
		t.Errorf("Expected: \n%v, \nreceived: \n%v", expectedCdr, rtCdr)
	}
	csvProcessor.cdrcCfgs[0].DataUsageMultiplyFactor = 1024
	expectedCdr = &engine.CDR{
		CGRID:       utils.Sha1("", sTime.String()),
		ToR:         cdrRow[0],
		OriginHost:  "0.0.0.0",
		Source:      "TEST_CDRC",
		Usage:       time.Duration(1024),
		ExtraFields: map[string]string{},
		Cost:        -1,
	}
	if rtCdr, _ := csvProcessor.recordToStoredCdr(cdrRow,
		cdrcConfig, "cgrates.org"); !reflect.DeepEqual(expectedCdr, rtCdr) {
		t.Errorf("Expected: \n%v, \nreceived: \n%v", expectedCdr, rtCdr)
	}
	cdrRow = []string{"*voice", "1s"}
	expectedCdr = &engine.CDR{
		CGRID:       utils.Sha1("", sTime.String()),
		ToR:         cdrRow[0],
		OriginHost:  "0.0.0.0",
		Source:      "TEST_CDRC",
		Usage:       time.Duration(1) * time.Second,
		ExtraFields: map[string]string{},
		Cost:        -1,
	}
	if rtCdr, _ := csvProcessor.recordToStoredCdr(cdrRow,
		cdrcConfig, "cgrates.org"); !reflect.DeepEqual(expectedCdr, rtCdr) {
		t.Errorf("Expected: \n%v, \nreceived: \n%v", expectedCdr, rtCdr)
	}
}

func TestCsvPairToRecord(t *testing.T) {
	eRecord := []string{"INVITE", "2daec40c", "548625ac",
		"dd0c4c617a9919d29a6175cdff223a9e@0:0:0:0:0:0:0:0", "200", "OK", "1436454408",
		"*prepaid", "1001", "1002", "", "3401:2069362475", "2"}
	invPr := &UnpairedRecord{Method: "INVITE",
		Timestamp: time.Date(2015, 7, 9, 15, 6, 48, 0, time.UTC),
		Values: []string{"INVITE", "2daec40c", "548625ac",
			"dd0c4c617a9919d29a6175cdff223a9e@0:0:0:0:0:0:0:0", "200", "OK",
			"1436454408", "*prepaid", "1001", "1002", "", "3401:2069362475"}}
	byePr := &UnpairedRecord{Method: "BYE", Timestamp: time.Date(2015, 7, 9, 15, 6, 50, 0, time.UTC),
		Values: []string{"BYE", "2daec40c", "548625ac",
			"dd0c4c617a9919d29a6175cdff223a9e@0:0:0:0:0:0:0:0", "200", "OK",
			"1436454410", "", "", "", "", "3401:2069362475"}}
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
	byePr.Values = []string{"BYE", "2daec40c", "548625ac",
		"dd0c4c617a9919d29a6175cdff223a9e@0:0:0:0:0:0:0:0", "200", "OK",
		"1436454410", "", "", "", "3401:2069362475"} // Took one value out
	if _, err := pairToRecord(invPr, byePr); err == nil || err.Error() != "INCONSISTENT_VALUES_LENGTH" {
		t.Error(err)
	}
}
