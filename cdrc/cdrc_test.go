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

package cdrc

import (
	"bytes"
	"encoding/csv"
	"io"
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestRecordForkCdr(t *testing.T) {
	cgrConfig, _ := config.NewDefaultCGRConfig()
	cdrcConfig := cgrConfig.CdrcProfiles["/var/log/cgrates/cdrc/in"][utils.META_DEFAULT]
	cdrcConfig.CdrFields = append(cdrcConfig.CdrFields, &config.CfgCdrField{Tag: "SupplierTest", Type: utils.CDRFIELD, CdrFieldId: "supplier", Value: []*utils.RSRField{&utils.RSRField{Id: "14"}}})
	cdrcConfig.CdrFields = append(cdrcConfig.CdrFields, &config.CfgCdrField{Tag: "DisconnectCauseTest", Type: utils.CDRFIELD, CdrFieldId: utils.DISCONNECT_CAUSE,
		Value: []*utils.RSRField{&utils.RSRField{Id: "16"}}})
	cdrc := &Cdrc{CdrFormat: CSV, cdrSourceIds: []string{"TEST_CDRC"}, cdrFields: [][]*config.CfgCdrField{cdrcConfig.CdrFields}}
	cdrRow := []string{"firstField", "secondField"}
	_, err := cdrc.recordToStoredCdr(cdrRow, 0)
	if err == nil {
		t.Error("Failed to corectly detect missing fields from record")
	}
	cdrRow = []string{"ignored", "ignored", utils.VOICE, "acc1", utils.META_PREPAID, "*out", "cgrates.org", "call", "1001", "1001", "+4986517174963",
		"2013-02-03 19:50:00", "2013-02-03 19:54:00", "62", "supplier1", "172.16.1.1", "NORMAL_DISCONNECT"}
	rtCdr, err := cdrc.recordToStoredCdr(cdrRow, 0)
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

func TestDataMultiplyFactor(t *testing.T) {
	cdrFields := []*config.CfgCdrField{&config.CfgCdrField{Tag: "TORField", Type: utils.CDRFIELD, CdrFieldId: "tor", Value: []*utils.RSRField{&utils.RSRField{Id: "0"}}},
		&config.CfgCdrField{Tag: "UsageField", Type: utils.CDRFIELD, CdrFieldId: "usage", Value: []*utils.RSRField{&utils.RSRField{Id: "1"}}}}
	cdrc := &Cdrc{CdrFormat: CSV, cdrSourceIds: []string{"TEST_CDRC"}, duMultiplyFactors: []float64{0}, cdrFields: [][]*config.CfgCdrField{cdrFields}}
	cdrRow := []string{"*data", "1"}
	rtCdr, err := cdrc.recordToStoredCdr(cdrRow, 0)
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
	cdrc.duMultiplyFactors = []float64{1024}
	expectedCdr = &engine.StoredCdr{
		CgrId:       utils.Sha1("", sTime.String()),
		TOR:         cdrRow[0],
		CdrHost:     "0.0.0.0",
		CdrSource:   "TEST_CDRC",
		Usage:       time.Duration(1024) * time.Second,
		ExtraFields: map[string]string{},
		Cost:        -1,
	}
	if rtCdr, _ := cdrc.recordToStoredCdr(cdrRow, 0); !reflect.DeepEqual(expectedCdr, rtCdr) {
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
	if rtCdr, _ := cdrc.recordToStoredCdr(cdrRow, 0); !reflect.DeepEqual(expectedCdr, rtCdr) {
		t.Errorf("Expected: \n%v, \nreceived: \n%v", expectedCdr, rtCdr)
	}
}

/*
func TestDnTdmCdrs(t *testing.T) {
	tdmCdrs := `
49773280254,0049LN130676000285,N_IP_0676_00-Internet 0676 WRAP 13,02.07.2014 15:24:40,02.07.2014 15:24:40,1,25,Peak,0.000000,49DE13
49893252121,0049651515477,N_MO_MRAP_00-WRAP Mobile,02.07.2014 15:24:41,02.07.2014 15:24:41,1,8,Peak,0.003920,49651
49497361022,0049LM0409005226,N_MO_MTMB_00-RW-Mobile,02.07.2014 15:24:41,02.07.2014 15:24:41,1,43,Peak,0.021050,49MTMB
`
	cgrConfig, _ := config.NewDefaultCGRConfig()
	eCdrs := []*engine.StoredCdr{
		&engine.StoredCdr{
			CgrId:       utils.Sha1("49773280254", time.Date(2014, 7, 2, 15, 24, 40, 0, time.UTC).String()),
			TOR:         utils.VOICE,
			AccId:       "49773280254",
			CdrHost:     "0.0.0.0",
			CdrSource:   cgrConfig.CdrcSourceId,
			ReqType:     utils.META_RATED,
			Direction:   "*out",
			Tenant:      "cgrates.org",
			Category:    "call",
			Account:     "+49773280254",
			Subject:     "+49773280254",
			Destination: "+49676000285",
			SetupTime:   time.Date(2014, 7, 2, 15, 24, 40, 0, time.UTC),
			AnswerTime:  time.Date(2014, 7, 2, 15, 24, 40, 0, time.UTC),
			Usage:       time.Duration(25) * time.Second,
			Cost:        -1,
		},
		&engine.StoredCdr{
			CgrId:       utils.Sha1("49893252121", time.Date(2014, 7, 2, 15, 24, 41, 0, time.UTC).String()),
			TOR:         utils.VOICE,
			AccId:       "49893252121",
			CdrHost:     "0.0.0.0",
			CdrSource:   cgrConfig.CdrcSourceId,
			ReqType:     utils.META_RATED,
			Direction:   "*out",
			Tenant:      "cgrates.org",
			Category:    "call",
			Account:     "+49893252121",
			Subject:     "+49893252121",
			Destination: "+49651515477",
			SetupTime:   time.Date(2014, 7, 2, 15, 24, 41, 0, time.UTC),
			AnswerTime:  time.Date(2014, 7, 2, 15, 24, 41, 0, time.UTC),
			Usage:       time.Duration(8) * time.Second,
			Cost:        -1,
		},
		&engine.StoredCdr{
			CgrId:       utils.Sha1("49497361022", time.Date(2014, 7, 2, 15, 24, 41, 0, time.UTC).String()),
			TOR:         utils.VOICE,
			AccId:       "49497361022",
			CdrHost:     "0.0.0.0",
			CdrSource:   cgrConfig.CdrcSourceId,
			ReqType:     utils.META_RATED,
			Direction:   "*out",
			Tenant:      "cgrates.org",
			Category:    "call",
			Account:     "+49497361022",
			Subject:     "+49497361022",
			Destination: "+499005226",
			SetupTime:   time.Date(2014, 7, 2, 15, 24, 41, 0, time.UTC),
			AnswerTime:  time.Date(2014, 7, 2, 15, 24, 41, 0, time.UTC),
			Usage:       time.Duration(43) * time.Second,
			Cost:        -1,
		},
	}
	torFld, _ := utils.NewRSRField("^*voice")
	acntFld, _ := utils.NewRSRField(`~0:s/^([1-9]\d+)$/+$1/`)
	reqTypeFld, _ := utils.NewRSRField("^rated")
	dirFld, _ := utils.NewRSRField("^*out")
	tenantFld, _ := utils.NewRSRField("^cgrates.org")
	categFld, _ := utils.NewRSRField("^call")
	dstFld, _ := utils.NewRSRField(`~1:s/^00(\d+)(?:[a-zA-Z].{3})*0*([1-9]\d+)$/+$1$2/`)
	usageFld, _ := utils.NewRSRField(`~6:s/^(\d+)$/${1}s/`)
	cgrConfig.CdrcCdrFields = map[string]*utils.RSRField{
		utils.TOR:         torFld,
		utils.ACCID:       &utils.RSRField{Id: "0"},
		utils.REQTYPE:     reqTypeFld,
		utils.DIRECTION:   dirFld,
		utils.TENANT:      tenantFld,
		utils.CATEGORY:    categFld,
		utils.ACCOUNT:     acntFld,
		utils.SUBJECT:     acntFld,
		utils.DESTINATION: dstFld,
		utils.SETUP_TIME:  &utils.RSRField{Id: "4"},
		utils.ANSWER_TIME: &utils.RSRField{Id: "4"},
		utils.USAGE:       usageFld,
	}
	cdrc := &Cdrc{cgrConfig.CdrcCdrs, cgrConfig.CdrcCdrFormat, cgrConfig.CdrcCdrInDir, cgrConfig.CdrcCdrOutDir, cgrConfig.CdrcSourceId, cgrConfig.CdrcRunDelay, ',',
		cgrConfig.CdrcCdrFields, new(cdrs.CDRS), nil}
	cdrsContent := bytes.NewReader([]byte(tdmCdrs))
	csvReader := csv.NewReader(cdrsContent)
	cdrs := make([]*engine.StoredCdr, 0)
	for {
		cdrCsv, err := csvReader.Read()
		if err != nil && err == io.EOF {
			break // End of file
		} else if err != nil {
			t.Error("Unexpected error:", err)
		}
		if cdr, err := cdrc.recordToStoredCdr(cdrCsv); err != nil {
			t.Error("Unexpected error: ", err)
		} else {
			cdrs = append(cdrs, cdr)
		}
	}
	if !reflect.DeepEqual(eCdrs, cdrs) {
		for _, ecdr := range eCdrs {
			fmt.Printf("Cdr expected: %+v\n", ecdr)
		}
		for _, cdr := range cdrs {
			fmt.Printf("Cdr processed: %+v\n", cdr)
		}
		t.Errorf("Expecting: %+v, received: %+v", eCdrs, cdrs)
	}

}
*/

func TestNewPartialFlatstoreRecord(t *testing.T) {
	ePr := &PartialFlatstoreRecord{Method: "INVITE", AccId: "dd0c4c617a9919d29a6175cdff223a9e@0:0:0:0:0:0:0:02daec40c548625ac", Timestamp: time.Date(2015, 7, 9, 17, 6, 48, 0, time.Local),
		Values: []string{"INVITE", "2daec40c", "548625ac", "dd0c4c617a9919d29a6175cdff223a9e@0:0:0:0:0:0:0:0", "200", "OK", "1436454408", "*prepaid", "1001", "1002", "", "3401:2069362475"}}
	if pr, err := NewPartialFlatstoreRecord(ePr.Values); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(ePr, pr) {
		t.Errorf("Expecting: %+v, received: %+v", ePr, pr)
	}
	if _, err := NewPartialFlatstoreRecord([]string{"INVITE", "2daec40c", "548625ac", "dd0c4c617a9919d29a6175cdff223a9e@0:0:0:0:0:0:0:0", "200", "OK"}); err == nil || err.Error() != "MISSING_IE" {
		t.Error(err)
	}
}

func TestPairToRecord(t *testing.T) {
	eRecord := []string{"INVITE", "2daec40c", "548625ac", "dd0c4c617a9919d29a6175cdff223a9e@0:0:0:0:0:0:0:0", "200", "OK", "1436454408", "*prepaid", "1001", "1002", "", "3401:2069362475", "2"}
	invPr := &PartialFlatstoreRecord{Method: "INVITE", Timestamp: time.Date(2015, 7, 9, 17, 6, 48, 0, time.Local),
		Values: []string{"INVITE", "2daec40c", "548625ac", "dd0c4c617a9919d29a6175cdff223a9e@0:0:0:0:0:0:0:0", "200", "OK", "1436454408", "*prepaid", "1001", "1002", "", "3401:2069362475"}}
	byePr := &PartialFlatstoreRecord{Method: "BYE", Timestamp: time.Date(2015, 7, 9, 17, 6, 50, 0, time.Local),
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

func TestOsipsFlatstoreCdrs(t *testing.T) {
	osipsCdrs := `
INVITE|2daec40c|548625ac|dd0c4c617a9919d29a6175cdff223a9e@0:0:0:0:0:0:0:0|200|OK|1436454408|*prepaid|1001|1002||3401:2069362475
BYE|2daec40c|548625ac|dd0c4c617a9919d29a6175cdff223a9e@0:0:0:0:0:0:0:0|200|OK|1436454410|||||3401:2069362475
INVITE|f9d3d5c3|c863a6e3|214d8f52b566e33a9349b184e72a4cca@0:0:0:0:0:0:0:0|200|OK|1436454647|*postpaid|1002|1001||1877:893549741
BYE|f9d3d5c3|c863a6e3|214d8f52b566e33a9349b184e72a4cca@0:0:0:0:0:0:0:0|200|OK|1436454651|||||1877:893549741
INVITE|36e39a5|42d996f9|3a63321dd3b325eec688dc2aefb6ac2d@0:0:0:0:0:0:0:0|200|OK|1436454657|*prepaid|1001|1002||2407:1884881533
BYE|36e39a5|42d996f9|3a63321dd3b325eec688dc2aefb6ac2d@0:0:0:0:0:0:0:0|200|OK|1436454661|||||2407:1884881533
INVITE|3111f3c9|49ca4c42|a58ebaae40d08d6757d8424fb09c4c54@0:0:0:0:0:0:0:0|200|OK|1436454690|*prepaid|1001|1002||3099:1909036290
BYE|3111f3c9|49ca4c42|a58ebaae40d08d6757d8424fb09c4c54@0:0:0:0:0:0:0:0|200|OK|1436454692|||||3099:1909036290
`

	eCdrs := []*engine.StoredCdr{
		&engine.StoredCdr{
			CgrId:       "e61034c34148a7c4f40623e00ca5e551d1408bf3",
			TOR:         utils.VOICE,
			AccId:       "dd0c4c617a9919d29a6175cdff223a9e@0:0:0:0:0:0:0:02daec40c548625ac",
			CdrHost:     "0.0.0.0",
			CdrSource:   "TEST_CDRC",
			ReqType:     utils.META_PREPAID,
			Direction:   "*out",
			Tenant:      "cgrates.org",
			Category:    "call",
			Account:     "1001",
			Subject:     "1001",
			Destination: "1002",
			SetupTime:   time.Date(2015, 7, 9, 17, 06, 48, 0, time.Local),
			AnswerTime:  time.Date(2015, 7, 9, 17, 06, 48, 0, time.Local),
			Usage:       time.Duration(2) * time.Second,
			ExtraFields: map[string]string{
				"DialogIdentifier": "3401:2069362475",
			},
			Cost: -1,
		},
		&engine.StoredCdr{
			CgrId:       "3ed64a28190e20ac8a6fd8fd48cb23efbfeb7a17",
			TOR:         utils.VOICE,
			AccId:       "214d8f52b566e33a9349b184e72a4cca@0:0:0:0:0:0:0:0f9d3d5c3c863a6e3",
			CdrHost:     "0.0.0.0",
			CdrSource:   "TEST_CDRC",
			ReqType:     utils.META_POSTPAID,
			Direction:   "*out",
			Tenant:      "cgrates.org",
			Category:    "call",
			Account:     "1002",
			Subject:     "1002",
			Destination: "1001",
			SetupTime:   time.Date(2015, 7, 9, 17, 10, 47, 0, time.Local),
			AnswerTime:  time.Date(2015, 7, 9, 17, 10, 47, 0, time.Local),
			Usage:       time.Duration(4) * time.Second,
			ExtraFields: map[string]string{
				"DialogIdentifier": "1877:893549741",
			},
			Cost: -1,
		},
		&engine.StoredCdr{
			CgrId:       "f2f8d9341adfbbe1836b22f75182142061ef3d20",
			TOR:         utils.VOICE,
			AccId:       "3a63321dd3b325eec688dc2aefb6ac2d@0:0:0:0:0:0:0:036e39a542d996f9",
			CdrHost:     "0.0.0.0",
			CdrSource:   "TEST_CDRC",
			ReqType:     utils.META_PREPAID,
			Direction:   "*out",
			Tenant:      "cgrates.org",
			Category:    "call",
			Account:     "1001",
			Subject:     "1001",
			Destination: "1002",
			SetupTime:   time.Date(2015, 7, 9, 17, 10, 57, 0, time.Local),
			AnswerTime:  time.Date(2015, 7, 9, 17, 10, 57, 0, time.Local),
			Usage:       time.Duration(4) * time.Second,
			ExtraFields: map[string]string{
				"DialogIdentifier": "2407:1884881533",
			},
			Cost: -1,
		},
		&engine.StoredCdr{
			CgrId:       "ccf05e7e3b9db9d2370bcbe316817447dba7df54",
			TOR:         utils.VOICE,
			AccId:       "a58ebaae40d08d6757d8424fb09c4c54@0:0:0:0:0:0:0:03111f3c949ca4c42",
			CdrHost:     "0.0.0.0",
			CdrSource:   "TEST_CDRC",
			ReqType:     utils.META_PREPAID,
			Direction:   "*out",
			Tenant:      "cgrates.org",
			Category:    "call",
			Account:     "1001",
			Subject:     "1001",
			Destination: "1002",
			SetupTime:   time.Date(2015, 7, 9, 17, 11, 30, 0, time.Local), //2015-07-09T17:11:30+02:00
			AnswerTime:  time.Date(2015, 7, 9, 17, 11, 30, 0, time.Local),
			Usage:       time.Duration(2) * time.Second,
			ExtraFields: map[string]string{
				"DialogIdentifier": "3099:1909036290",
			},
			Cost: -1,
		},
	}

	cdrFields := [][]*config.CfgCdrField{[]*config.CfgCdrField{
		&config.CfgCdrField{Tag: "Tor", Type: utils.CDRFIELD, CdrFieldId: utils.TOR, Value: utils.ParseRSRFieldsMustCompile("^*voice", utils.INFIELD_SEP), Mandatory: true},
		&config.CfgCdrField{Tag: "AccId", Type: utils.CDRFIELD, CdrFieldId: utils.ACCID, Mandatory: true},
		&config.CfgCdrField{Tag: "ReqType", Type: utils.CDRFIELD, CdrFieldId: utils.REQTYPE, Value: utils.ParseRSRFieldsMustCompile("7", utils.INFIELD_SEP), Mandatory: true},
		&config.CfgCdrField{Tag: "Direction", Type: utils.CDRFIELD, CdrFieldId: utils.DIRECTION, Value: utils.ParseRSRFieldsMustCompile("^*out", utils.INFIELD_SEP), Mandatory: true},
		&config.CfgCdrField{Tag: "Direction", Type: utils.CDRFIELD, CdrFieldId: utils.DIRECTION, Value: utils.ParseRSRFieldsMustCompile("^*out", utils.INFIELD_SEP), Mandatory: true},
		&config.CfgCdrField{Tag: "Tenant", Type: utils.CDRFIELD, CdrFieldId: utils.TENANT, Value: utils.ParseRSRFieldsMustCompile("^cgrates.org", utils.INFIELD_SEP), Mandatory: true},
		&config.CfgCdrField{Tag: "Category", Type: utils.CDRFIELD, CdrFieldId: utils.CATEGORY, Value: utils.ParseRSRFieldsMustCompile("^call", utils.INFIELD_SEP), Mandatory: true},
		&config.CfgCdrField{Tag: "Account", Type: utils.CDRFIELD, CdrFieldId: utils.ACCOUNT, Value: utils.ParseRSRFieldsMustCompile("8", utils.INFIELD_SEP), Mandatory: true},
		&config.CfgCdrField{Tag: "Subject", Type: utils.CDRFIELD, CdrFieldId: utils.SUBJECT, Value: utils.ParseRSRFieldsMustCompile("8", utils.INFIELD_SEP), Mandatory: true},
		&config.CfgCdrField{Tag: "Destination", Type: utils.CDRFIELD, CdrFieldId: utils.DESTINATION, Value: utils.ParseRSRFieldsMustCompile("9", utils.INFIELD_SEP), Mandatory: true},
		&config.CfgCdrField{Tag: "SetupTime", Type: utils.CDRFIELD, CdrFieldId: utils.SETUP_TIME, Value: utils.ParseRSRFieldsMustCompile("6", utils.INFIELD_SEP), Mandatory: true},
		&config.CfgCdrField{Tag: "AnswerTime", Type: utils.CDRFIELD, CdrFieldId: utils.ANSWER_TIME, Value: utils.ParseRSRFieldsMustCompile("6", utils.INFIELD_SEP), Mandatory: true},
		&config.CfgCdrField{Tag: "Duration", Type: utils.CDRFIELD, CdrFieldId: utils.USAGE, Mandatory: true},
		&config.CfgCdrField{Tag: "DialogId", Type: utils.CDRFIELD, CdrFieldId: "DialogIdentifier", Value: utils.ParseRSRFieldsMustCompile("11", utils.INFIELD_SEP)},
	}}
	cdrc := &Cdrc{CdrFormat: utils.OSIPS_FLATSTORE, cdrSourceIds: []string{"TEST_CDRC"}, cdrFields: cdrFields, partialRecords: make(map[string]map[string]*PartialFlatstoreRecord)}
	cdrsContent := bytes.NewReader([]byte(osipsCdrs))
	csvReader := csv.NewReader(cdrsContent)
	csvReader.Comma = '|'
	cdrs := make([]*engine.StoredCdr, 0)
	recNrs := 0
	for {
		recNrs++
		cdrCsv, err := csvReader.Read()
		if err != nil && err == io.EOF {
			break // End of file
		} else if err != nil {
			t.Error("Unexpected error:", err)
		}
		record, err := cdrc.processPartialRecord(cdrCsv, "dummyfilename")
		if err != nil {
			t.Error(err)
		}
		if record == nil {
			continue // Partial record
		}
		if storedCdr, err := cdrc.recordToStoredCdr(record, 0); err != nil {
			t.Error(err)
		} else if storedCdr != nil {
			cdrs = append(cdrs, storedCdr)
		}
	}
	if !reflect.DeepEqual(eCdrs, cdrs) {
		t.Errorf("Expecting: %+v, received: %+v", eCdrs, cdrs)
	}

}

/*
func TestOsipsFlatstoreMissedCdrs(t *testing.T) {
	osipsCdrs := `
INVITE|ef6c6256|da501581|0bfdd176d1b93e7df3de5c6f4873ee04@0:0:0:0:0:0:0:0|487|Request Terminated|1436454643|*prepaid|1001|1002||1224:339382783
INVITE|7905e511||81880da80a94bda81b425b09009e055c@0:0:0:0:0:0:0:0|404|Not Found|1436454668|*prepaid|1001|1002||1980:1216490844
INVITE|25f7def5||9984b9ec535a7d317b542744d48d0ed6@0:0:0:0:0:0:0:0|404|Not Found|1436454669|*prepaid|1001|1002||14:1595110662
INVITE|ae0a7f6c||02f7fa9334db7aa4130bbf7627370621@0:0:0:0:0:0:0:0|404|Not Found|1436454670|*prepaid|1001|1002||176:1975670970
INVITE|32b97104||3154c2a80294f538991a88d86f4e1085@0:0:0:0:0:0:0:0|404|Not Found|1436454678|*prepaid|1001|1002||2607:1024337552
INVITE|324cb497|d4af7023|8deaadf2ae9a17809a391f05af31afb0@0:0:0:0:0:0:0:0|486|Busy here|1436454687|*postpaid|1002|1002||474:130115066
INVITE|167ac4db|c53c85e5|4b3885cb78dde44dc7936abd2fa281e1@0:0:0:0:0:0:0:0|487|Request Terminated|1436454695|*postpaid|1002|1002||1922:549002535
`
}
*/
