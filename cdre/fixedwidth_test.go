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

package cdre

import (
	"bytes"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	"testing"
	"time"
)

var contentCfgFlds = []*config.CgrXmlCfgCdrField{
	&config.CgrXmlCfgCdrField{Name: "RecordType", Type: CONSTANT, Value: "20", Width: 2},
	&config.CgrXmlCfgCdrField{Name: "SIPTrunkID", Type: CDRFIELD, Value: utils.ACCOUNT, Width: 12, Strip: "left", Padding: "right"},
	&config.CgrXmlCfgCdrField{Name: "ConnectionNumber", Type: CDRFIELD, Value: utils.SUBJECT, Width: 5, Strip: "right", Padding: "right"},
	&config.CgrXmlCfgCdrField{Name: "ANumber", Type: CDRFIELD, Value: "cli", Width: 15, Strip: "xright", Padding: "right"},
	&config.CgrXmlCfgCdrField{Name: "CalledNumber", Type: CDRFIELD, Value: "destination", Width: 24, Strip: "xright", Padding: "right"},
	&config.CgrXmlCfgCdrField{Name: "ServiceType", Type: CONSTANT, Value: "02", Width: 2},
	&config.CgrXmlCfgCdrField{Name: "ServiceIdentification", Type: CONSTANT, Value: "11", Width: 4, Padding: "right"},
	&config.CgrXmlCfgCdrField{Name: "StartChargingDateTime", Type: CDRFIELD, Value: utils.SETUP_TIME, Width: 12, Strip: "right", Padding: "right", Layout: "020106150400"},
	&config.CgrXmlCfgCdrField{Name: "ChargeableTime", Type: CDRFIELD, Value: utils.DURATION, Width: 6, Strip: "right", Padding: "right"},
	&config.CgrXmlCfgCdrField{Name: "DataVolume", Type: FILLER, Width: 6, Padding: "right"},
	&config.CgrXmlCfgCdrField{Name: "TaxCode", Type: CONSTANT, Value: "1", Width: 1},
	&config.CgrXmlCfgCdrField{Name: "OperatorTAPCode", Type: CDRFIELD, Value: "opertapcode", Width: 2, Strip: "right", Padding: "right"},
	&config.CgrXmlCfgCdrField{Name: "ProductNumber", Type: CDRFIELD, Value: "productnumber", Width: 5, Strip: "right", Padding: "right"},
	&config.CgrXmlCfgCdrField{Name: "NetworkSubtype", Type: CONSTANT, Value: "3", Width: 1},
	&config.CgrXmlCfgCdrField{Name: "SessionID", Type: CDRFIELD, Value: utils.ACCID, Width: 16, Padding: "right"},
	&config.CgrXmlCfgCdrField{Name: "VolumeUP", Type: FILLER, Width: 8, Padding: "right"},
	&config.CgrXmlCfgCdrField{Name: "VolumeDown", Type: FILLER, Width: 8, Padding: "right"},
	&config.CgrXmlCfgCdrField{Name: "TerminatingOperator", Type: CONCATENATED_CDRFIELD, Value: "tapcode,operatorcode", Width: 5, Strip: "right", Padding: "right"},
	&config.CgrXmlCfgCdrField{Name: "EndCharge", Type: CDRFIELD, Value: utils.COST, Width: 9, Padding: "zeroleft"},
	&config.CgrXmlCfgCdrField{Name: "CallMaskingIndicator", Type: CDRFIELD, Value: "calledmask", Width: 1, Strip: "right", Padding: "right"},
}

// Write one CDR and test it's results only for content buffer
func TestWriteCdr(t *testing.T) {
	wrBuf := &bytes.Buffer{}
	exportTpl := &config.CgrXmlCdreFwCfg{Content: &config.CgrXmlCfgCdrContent{Fields: contentCfgFlds}}
	fwWriter := FixedWidthCdrWriter{writer: wrBuf, exportTemplate: exportTpl, roundDecimals: 4, header: &bytes.Buffer{}, content: &bytes.Buffer{}, trailer: &bytes.Buffer{}}
	cdr := &utils.StoredCdr{CgrId: utils.FSCgrId("dsafdsaf"), AccId: "dsafdsaf", CdrHost: "192.168.1.1", ReqType: "rated", Direction: "*out", Tenant: "cgrates.org",
		TOR: "call", Account: "1001", Subject: "1001", Destination: "1002",
		SetupTime:  time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC),
		AnswerTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC),
		Duration:   time.Duration(10) * time.Second, MediationRunId: utils.DEFAULT_RUNID, Cost: 2.34567,
		ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
	}
	if err := fwWriter.WriteCdr(cdr); err != nil {
		t.Error(err)
	}
	contentOut := fwWriter.content.String()
	if len(contentOut) != 145 {
		t.Error("Unexpected content length", len(contentOut))
	}
	eOut := "201001        1001                1002                    0211  07111308420010          1       3dsafdsaf                             0002.3457 \n"
	if contentOut != eOut {
		t.Errorf("Content out different than expected. Have <%s>, expecting: <%s>", contentOut, eOut)
	}
	outBeforeWrite := ""
	if wrBuf.String() != outBeforeWrite {
		t.Errorf("Output buffer should be empty before write")
	}
	fwWriter.Close()
	if wrBuf.String() != eOut {
		t.Errorf("Output buffer does not contain expected info. Have <%s>, expecting: <%s>", wrBuf.String(), eOut)
	}
}
