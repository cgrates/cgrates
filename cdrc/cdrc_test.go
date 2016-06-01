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

/*
func TestNewPartialFlatstoreRecord(t *testing.T) {
	ePr := &PartialFlatstoreRecord{Method: "INVITE", AccId: "dd0c4c617a9919d29a6175cdff223a9e@0:0:0:0:0:0:0:02daec40c548625ac", Timestamp: time.Date(2015, 7, 9, 15, 6, 48, 0, time.UTC),
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
*/

/*
func TestOsipsFlatstoreCdrs(t *testing.T) {
	flatstoreCdrs := `
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
			CgrId:           "e61034c34148a7c4f40623e00ca5e551d1408bf3",
			TOR:             utils.VOICE,
			AccId:           "dd0c4c617a9919d29a6175cdff223a9e@0:0:0:0:0:0:0:02daec40c548625ac",
			CdrHost:         "0.0.0.0",
			CdrSource:       "TEST_CDRC",
			ReqType:         utils.META_PREPAID,
			Direction:       "*out",
			Tenant:          "cgrates.org",
			Category:        "call",
			Account:         "1001",
			Subject:         "1001",
			Destination:     "1002",
			SetupTime:       time.Date(2015, 7, 9, 15, 06, 48, 0, time.UTC),
			AnswerTime:      time.Date(2015, 7, 9, 15, 06, 48, 0, time.UTC),
			Usage:           time.Duration(2) * time.Second,
			DisconnectCause: "200 OK",
			ExtraFields: map[string]string{
				"DialogIdentifier": "3401:2069362475",
			},
			Cost: -1,
		},
		&engine.StoredCdr{
			CgrId:           "3ed64a28190e20ac8a6fd8fd48cb23efbfeb7a17",
			TOR:             utils.VOICE,
			AccId:           "214d8f52b566e33a9349b184e72a4cca@0:0:0:0:0:0:0:0f9d3d5c3c863a6e3",
			CdrHost:         "0.0.0.0",
			CdrSource:       "TEST_CDRC",
			ReqType:         utils.META_POSTPAID,
			Direction:       "*out",
			Tenant:          "cgrates.org",
			Category:        "call",
			Account:         "1002",
			Subject:         "1002",
			Destination:     "1001",
			SetupTime:       time.Date(2015, 7, 9, 15, 10, 47, 0, time.UTC),
			AnswerTime:      time.Date(2015, 7, 9, 15, 10, 47, 0, time.UTC),
			Usage:           time.Duration(4) * time.Second,
			DisconnectCause: "200 OK",
			ExtraFields: map[string]string{
				"DialogIdentifier": "1877:893549741",
			},
			Cost: -1,
		},
		&engine.StoredCdr{
			CgrId:           "f2f8d9341adfbbe1836b22f75182142061ef3d20",
			TOR:             utils.VOICE,
			AccId:           "3a63321dd3b325eec688dc2aefb6ac2d@0:0:0:0:0:0:0:036e39a542d996f9",
			CdrHost:         "0.0.0.0",
			CdrSource:       "TEST_CDRC",
			ReqType:         utils.META_PREPAID,
			Direction:       "*out",
			Tenant:          "cgrates.org",
			Category:        "call",
			Account:         "1001",
			Subject:         "1001",
			Destination:     "1002",
			SetupTime:       time.Date(2015, 7, 9, 15, 10, 57, 0, time.UTC),
			AnswerTime:      time.Date(2015, 7, 9, 15, 10, 57, 0, time.UTC),
			Usage:           time.Duration(4) * time.Second,
			DisconnectCause: "200 OK",
			ExtraFields: map[string]string{
				"DialogIdentifier": "2407:1884881533",
			},
			Cost: -1,
		},
		&engine.StoredCdr{
			CgrId:           "ccf05e7e3b9db9d2370bcbe316817447dba7df54",
			TOR:             utils.VOICE,
			AccId:           "a58ebaae40d08d6757d8424fb09c4c54@0:0:0:0:0:0:0:03111f3c949ca4c42",
			CdrHost:         "0.0.0.0",
			CdrSource:       "TEST_CDRC",
			ReqType:         utils.META_PREPAID,
			Direction:       "*out",
			Tenant:          "cgrates.org",
			Category:        "call",
			Account:         "1001",
			Subject:         "1001",
			Destination:     "1002",
			SetupTime:       time.Date(2015, 7, 9, 15, 11, 30, 0, time.UTC), //2015-07-09T17:11:30+02:00
			AnswerTime:      time.Date(2015, 7, 9, 15, 11, 30, 0, time.UTC),
			Usage:           time.Duration(2) * time.Second,
			DisconnectCause: "200 OK",
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
		&config.CfgCdrField{Tag: "DisconnectCause", Type: utils.CDRFIELD, CdrFieldId: utils.DISCONNECT_CAUSE, Value: utils.ParseRSRFieldsMustCompile("4;^ ;5", utils.INFIELD_SEP), Mandatory: true},
		&config.CfgCdrField{Tag: "DialogId", Type: utils.CDRFIELD, CdrFieldId: "DialogIdentifier", Value: utils.ParseRSRFieldsMustCompile("11", utils.INFIELD_SEP)},
	}}
	cdrc := &Cdrc{CdrFormat: utils.OSIPS_FLATSTORE, cdrSourceIds: []string{"TEST_CDRC"}, failedCallsPrefix: "missed_calls",
		cdrFields: cdrFields, partialRecords: make(map[string]map[string]*PartialFlatstoreRecord),
		guard: engine.Guardian}
	cdrsContent := bytes.NewReader([]byte(flatstoreCdrs))
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

func TestOsipsFlatstoreMissedCdrs(t *testing.T) {
	flatstoreCdrs := `
INVITE|ef6c6256|da501581|0bfdd176d1b93e7df3de5c6f4873ee04@0:0:0:0:0:0:0:0|487|Request Terminated|1436454643|*prepaid|1001|1002||1224:339382783
INVITE|7905e511||81880da80a94bda81b425b09009e055c@0:0:0:0:0:0:0:0|404|Not Found|1436454668|*prepaid|1001|1002||1980:1216490844
INVITE|324cb497|d4af7023|8deaadf2ae9a17809a391f05af31afb0@0:0:0:0:0:0:0:0|486|Busy here|1436454687|*postpaid|1002|1001||474:130115066
`
	eCdrs := []*engine.StoredCdr{
		&engine.StoredCdr{
			CgrId:           "1c20aa6543a5a30d26b2354ae79e1f5fb720e8e5",
			TOR:             utils.VOICE,
			AccId:           "0bfdd176d1b93e7df3de5c6f4873ee04@0:0:0:0:0:0:0:0ef6c6256da501581",
			CdrHost:         "0.0.0.0",
			CdrSource:       "TEST_CDRC",
			ReqType:         utils.META_PREPAID,
			Direction:       "*out",
			Tenant:          "cgrates.org",
			Category:        "call",
			Account:         "1001",
			Subject:         "1001",
			Destination:     "1002",
			SetupTime:       time.Date(2015, 7, 9, 15, 10, 43, 0, time.UTC),
			AnswerTime:      time.Date(2015, 7, 9, 15, 10, 43, 0, time.UTC),
			Usage:           0,
			DisconnectCause: "487 Request Terminated",
			ExtraFields: map[string]string{
				"DialogIdentifier": "1224:339382783",
			},
			Cost: -1,
		},
		&engine.StoredCdr{
			CgrId:           "054ab7c6c7fe6dc4a72f34e270027fa2aa930a58",
			TOR:             utils.VOICE,
			AccId:           "81880da80a94bda81b425b09009e055c@0:0:0:0:0:0:0:07905e511",
			CdrHost:         "0.0.0.0",
			CdrSource:       "TEST_CDRC",
			ReqType:         utils.META_PREPAID,
			Direction:       "*out",
			Tenant:          "cgrates.org",
			Category:        "call",
			Account:         "1001",
			Subject:         "1001",
			Destination:     "1002",
			SetupTime:       time.Date(2015, 7, 9, 15, 11, 8, 0, time.UTC),
			AnswerTime:      time.Date(2015, 7, 9, 15, 11, 8, 0, time.UTC),
			Usage:           0,
			DisconnectCause: "404 Not Found",
			ExtraFields: map[string]string{
				"DialogIdentifier": "1980:1216490844",
			},
			Cost: -1,
		},
		&engine.StoredCdr{
			CgrId:           "d49ea63d1655b15149336004629f1cadd1434b89",
			TOR:             utils.VOICE,
			AccId:           "8deaadf2ae9a17809a391f05af31afb0@0:0:0:0:0:0:0:0324cb497d4af7023",
			CdrHost:         "0.0.0.0",
			CdrSource:       "TEST_CDRC",
			ReqType:         utils.META_POSTPAID,
			Direction:       "*out",
			Tenant:          "cgrates.org",
			Category:        "call",
			Account:         "1002",
			Subject:         "1002",
			Destination:     "1001",
			SetupTime:       time.Date(2015, 7, 9, 15, 11, 27, 0, time.UTC),
			AnswerTime:      time.Date(2015, 7, 9, 15, 11, 27, 0, time.UTC),
			Usage:           0,
			DisconnectCause: "486 Busy here",
			ExtraFields: map[string]string{
				"DialogIdentifier": "474:130115066",
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
		&config.CfgCdrField{Tag: "Usage", Type: utils.CDRFIELD, CdrFieldId: utils.USAGE, Mandatory: true},
		&config.CfgCdrField{Tag: "DisconnectCause", Type: utils.CDRFIELD, CdrFieldId: utils.DISCONNECT_CAUSE, Value: utils.ParseRSRFieldsMustCompile("4;^ ;5", utils.INFIELD_SEP), Mandatory: true},
		&config.CfgCdrField{Tag: "DialogId", Type: utils.CDRFIELD, CdrFieldId: "DialogIdentifier", Value: utils.ParseRSRFieldsMustCompile("11", utils.INFIELD_SEP)},
	}}
	cdrc := &Cdrc{CdrFormat: utils.OSIPS_FLATSTORE, cdrSourceIds: []string{"TEST_CDRC"}, failedCallsPrefix: "missed_calls",
		cdrFields: cdrFields, partialRecords: make(map[string]map[string]*PartialFlatstoreRecord),
		guard: engine.Guardian}
	cdrsContent := bytes.NewReader([]byte(flatstoreCdrs))
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
		record, err := cdrc.processPartialRecord(cdrCsv, "missed_calls_1.log")
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
*/
