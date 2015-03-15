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
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

func TestRecordForkCdr(t *testing.T) {
	cgrConfig, _ := config.NewDefaultCGRConfig()
	cdrcConfig := cgrConfig.CdrcProfiles["/var/log/cgrates/cdrc/in"][utils.META_DEFAULT]
	cdrcConfig.CdrFields = append(cdrcConfig.CdrFields, &config.CfgCdrField{Tag: "SupplierTest", Type: utils.CDRFIELD, CdrFieldId: "supplier", Value: []*utils.RSRField{&utils.RSRField{Id: "14"}}})
	cdrc := &Cdrc{CdrFormat: CSV, cdrSourceId: "TEST_CDRC", cdrFields: [][]*config.CfgCdrField{cdrcConfig.CdrFields}}
	cdrRow := []string{"firstField", "secondField"}
	_, err := cdrc.recordToStoredCdr(cdrRow, cdrc.cdrFields[0])
	if err == nil {
		t.Error("Failed to corectly detect missing fields from record")
	}
	cdrRow = []string{"ignored", "ignored", utils.VOICE, "acc1", utils.META_PREPAID, "*out", "cgrates.org", "call", "1001", "1001", "+4986517174963",
		"2013-02-03 19:50:00", "2013-02-03 19:54:00", "62", "supplier1", "172.16.1.1"}
	rtCdr, err := cdrc.recordToStoredCdr(cdrRow, cdrc.cdrFields[0])
	if err != nil {
		t.Error("Failed to parse CDR in rated cdr", err)
	}
	expectedCdr := &utils.StoredCdr{
		CgrId:       utils.Sha1(cdrRow[3], time.Date(2013, 2, 3, 19, 50, 0, 0, time.UTC).String()),
		TOR:         cdrRow[2],
		AccId:       cdrRow[3],
		CdrHost:     "0.0.0.0", // Got it over internal interface
		CdrSource:   "TEST_CDRC",
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

func TestDataMultiplyFactor(t *testing.T) {
	cdrFields := []*config.CfgCdrField{&config.CfgCdrField{Tag: "TORField", Type: utils.CDRFIELD, CdrFieldId: "tor", Value: []*utils.RSRField{&utils.RSRField{Id: "0"}}},
		&config.CfgCdrField{Tag: "UsageField", Type: utils.CDRFIELD, CdrFieldId: "usage", Value: []*utils.RSRField{&utils.RSRField{Id: "1"}}}}
	cdrc := &Cdrc{CdrFormat: CSV, cdrSourceId: "TEST_CDRC", cdrFields: [][]*config.CfgCdrField{cdrFields}}
	cdrRow := []string{"*data", "1"}
	rtCdr, err := cdrc.recordToStoredCdr(cdrRow, cdrc.cdrFields[0])
	if err != nil {
		t.Error("Failed to parse CDR in rated cdr", err)
	}
	var sTime time.Time
	expectedCdr := &utils.StoredCdr{
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
	cdrc.duMultiplyFactor = 1024
	expectedCdr = &utils.StoredCdr{
		CgrId:       utils.Sha1("", sTime.String()),
		TOR:         cdrRow[0],
		CdrHost:     "0.0.0.0",
		CdrSource:   "TEST_CDRC",
		Usage:       time.Duration(1024) * time.Second,
		ExtraFields: map[string]string{},
		Cost:        -1,
	}
	if rtCdr, _ := cdrc.recordToStoredCdr(cdrRow, cdrc.cdrFields[0]); !reflect.DeepEqual(expectedCdr, rtCdr) {
		t.Errorf("Expected: \n%v, \nreceived: \n%v", expectedCdr, rtCdr)
	}
	cdrRow = []string{"*voice", "1"}
	expectedCdr = &utils.StoredCdr{
		CgrId:       utils.Sha1("", sTime.String()),
		TOR:         cdrRow[0],
		CdrHost:     "0.0.0.0",
		CdrSource:   "TEST_CDRC",
		Usage:       time.Duration(1) * time.Second,
		ExtraFields: map[string]string{},
		Cost:        -1,
	}
	if rtCdr, _ := cdrc.recordToStoredCdr(cdrRow, cdrc.cdrFields[0]); !reflect.DeepEqual(expectedCdr, rtCdr) {
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
	eCdrs := []*utils.StoredCdr{
		&utils.StoredCdr{
			CgrId:       utils.Sha1("49773280254", time.Date(2014, 7, 2, 15, 24, 40, 0, time.UTC).String()),
			TOR:         utils.VOICE,
			AccId:       "49773280254",
			CdrHost:     "0.0.0.0",
			CdrSource:   cgrConfig.CdrcSourceId,
			ReqType:     utils.META_RATED,
			Direction:   "*out",
			Tenant:      "sip.test.deanconnect.nl",
			Category:    "call",
			Account:     "+49773280254",
			Subject:     "+49773280254",
			Destination: "+49676000285",
			SetupTime:   time.Date(2014, 7, 2, 15, 24, 40, 0, time.UTC),
			AnswerTime:  time.Date(2014, 7, 2, 15, 24, 40, 0, time.UTC),
			Usage:       time.Duration(25) * time.Second,
			Cost:        -1,
		},
		&utils.StoredCdr{
			CgrId:       utils.Sha1("49893252121", time.Date(2014, 7, 2, 15, 24, 41, 0, time.UTC).String()),
			TOR:         utils.VOICE,
			AccId:       "49893252121",
			CdrHost:     "0.0.0.0",
			CdrSource:   cgrConfig.CdrcSourceId,
			ReqType:     utils.META_RATED,
			Direction:   "*out",
			Tenant:      "sip.test.deanconnect.nl",
			Category:    "call",
			Account:     "+49893252121",
			Subject:     "+49893252121",
			Destination: "+49651515477",
			SetupTime:   time.Date(2014, 7, 2, 15, 24, 41, 0, time.UTC),
			AnswerTime:  time.Date(2014, 7, 2, 15, 24, 41, 0, time.UTC),
			Usage:       time.Duration(8) * time.Second,
			Cost:        -1,
		},
		&utils.StoredCdr{
			CgrId:       utils.Sha1("49497361022", time.Date(2014, 7, 2, 15, 24, 41, 0, time.UTC).String()),
			TOR:         utils.VOICE,
			AccId:       "49497361022",
			CdrHost:     "0.0.0.0",
			CdrSource:   cgrConfig.CdrcSourceId,
			ReqType:     utils.META_RATED,
			Direction:   "*out",
			Tenant:      "sip.test.deanconnect.nl",
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
	tenantFld, _ := utils.NewRSRField("^sip.test.deanconnect.nl")
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
	cdrs := make([]*utils.StoredCdr, 0)
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
