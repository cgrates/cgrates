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
	"math"
	"testing"
	"time"
)

var hdrCfgFlds = []*config.CgrXmlCfgCdrField{
	&config.CgrXmlCfgCdrField{Name: "TypeOfRecord", Type: CONSTANT, Value: "10", Width: 2},
	&config.CgrXmlCfgCdrField{Name: "Filler1", Type: FILLER, Width: 3},
	&config.CgrXmlCfgCdrField{Name: "DistributorCode", Type: CONSTANT, Value: "VOI", Width: 3},
	&config.CgrXmlCfgCdrField{Name: "FileSeqNr", Type: METATAG, Value: "export_id", Width: 5, Strip: "right", Padding: "zeroleft"},
	&config.CgrXmlCfgCdrField{Name: "LastCdr", Type: METATAG, Value: META_LASTCDRATIME, Width: 12, Layout: "020106150400"},
	&config.CgrXmlCfgCdrField{Name: "FileCreationfTime", Type: METATAG, Value: "time_now", Width: 12, Layout: "020106150400"},
	&config.CgrXmlCfgCdrField{Name: "FileVersion", Type: CONSTANT, Value: "01", Width: 2},
	&config.CgrXmlCfgCdrField{Name: "Filler2", Type: FILLER, Width: 105},
}

var contentCfgFlds = []*config.CgrXmlCfgCdrField{
	&config.CgrXmlCfgCdrField{Name: "TypeOfRecord", Type: CONSTANT, Value: "20", Width: 2},
	&config.CgrXmlCfgCdrField{Name: "Account", Type: CDRFIELD, Value: utils.ACCOUNT, Width: 12, Strip: "left", Padding: "right"},
	&config.CgrXmlCfgCdrField{Name: "Subject", Type: CDRFIELD, Value: utils.SUBJECT, Width: 5, Strip: "right", Padding: "right"},
	&config.CgrXmlCfgCdrField{Name: "CLI", Type: CDRFIELD, Value: "cli", Width: 15, Strip: "xright", Padding: "right"},
	&config.CgrXmlCfgCdrField{Name: "Destination", Type: CDRFIELD, Value: utils.DESTINATION, Width: 24, Strip: "xright", Padding: "right"},
	&config.CgrXmlCfgCdrField{Name: "TOR", Type: CONSTANT, Value: "02", Width: 2},
	&config.CgrXmlCfgCdrField{Name: "SubtypeTOR", Type: CONSTANT, Value: "11", Width: 4, Padding: "right"},
	&config.CgrXmlCfgCdrField{Name: "SetupTime", Type: CDRFIELD, Value: utils.SETUP_TIME, Width: 12, Strip: "right", Padding: "right", Layout: "020106150400"},
	&config.CgrXmlCfgCdrField{Name: "Duration", Type: CDRFIELD, Value: utils.DURATION, Width: 6, Strip: "right", Padding: "right", Layout: utils.SECONDS},
	&config.CgrXmlCfgCdrField{Name: "DataVolume", Type: FILLER, Width: 6},
	&config.CgrXmlCfgCdrField{Name: "TaxCode", Type: CONSTANT, Value: "1", Width: 1},
	&config.CgrXmlCfgCdrField{Name: "OperatorCode", Type: CDRFIELD, Value: "opercode", Width: 2, Strip: "right", Padding: "right"},
	&config.CgrXmlCfgCdrField{Name: "ProductId", Type: CDRFIELD, Value: "productid", Width: 5, Strip: "right", Padding: "right"},
	&config.CgrXmlCfgCdrField{Name: "NetworkId", Type: CONSTANT, Value: "3", Width: 1},
	&config.CgrXmlCfgCdrField{Name: "CallId", Type: CDRFIELD, Value: utils.ACCID, Width: 16, Padding: "right"},
	&config.CgrXmlCfgCdrField{Name: "Filler", Type: FILLER, Width: 8},
	&config.CgrXmlCfgCdrField{Name: "Filler", Type: FILLER, Width: 8},
	&config.CgrXmlCfgCdrField{Name: "TerminationCode", Type: CONCATENATED_CDRFIELD, Value: "operator,product", Width: 5, Strip: "right", Padding: "right"},
	&config.CgrXmlCfgCdrField{Name: "Cost", Type: CDRFIELD, Value: utils.COST, Width: 9, Padding: "zeroleft"},
	&config.CgrXmlCfgCdrField{Name: "DestinationPrivacy", Type: METATAG, Value: META_MASKDESTINATION, Width: 1},
}

var trailerCfgFlds = []*config.CgrXmlCfgCdrField{
	&config.CgrXmlCfgCdrField{Name: "TypeOfRecord", Type: CONSTANT, Value: "90", Width: 2},
	&config.CgrXmlCfgCdrField{Name: "Filler1", Type: FILLER, Width: 3},
	&config.CgrXmlCfgCdrField{Name: "DistributorCode", Type: CONSTANT, Value: "VOI", Width: 3},
	&config.CgrXmlCfgCdrField{Name: "FileSeqNr", Type: METATAG, Value: META_EXPORTID, Width: 5, Strip: "right", Padding: "zeroleft"},
	&config.CgrXmlCfgCdrField{Name: "NumberOfRecords", Type: METATAG, Value: META_NRCDRS, Width: 6, Padding: "zeroleft"},
	&config.CgrXmlCfgCdrField{Name: "CdrsDuration", Type: METATAG, Value: META_DURCDRS, Width: 8, Padding: "zeroleft"},
	&config.CgrXmlCfgCdrField{Name: "FirstCdrTime", Type: METATAG, Value: META_FIRSTCDRATIME, Width: 12, Layout: "020106150400"},
	&config.CgrXmlCfgCdrField{Name: "LastCdrTime", Type: METATAG, Value: META_LASTCDRATIME, Width: 12, Layout: "020106150400"},
	&config.CgrXmlCfgCdrField{Name: "Filler2", Type: FILLER, Width: 93},
}

// Write one CDR and test it's results only for content buffer
func TestWriteCdr(t *testing.T) {
	wrBuf := &bytes.Buffer{}
	exportTpl := &config.CgrXmlCdreFwCfg{Header: &config.CgrXmlCfgCdrHeader{Fields: hdrCfgFlds},
		Content: &config.CgrXmlCfgCdrContent{Fields: contentCfgFlds},
		Trailer: &config.CgrXmlCfgCdrTrailer{Fields: trailerCfgFlds},
	}
	fwWriter := FixedWidthCdrWriter{writer: wrBuf, exportTemplate: exportTpl, roundDecimals: 4, header: &bytes.Buffer{}, content: &bytes.Buffer{}, trailer: &bytes.Buffer{}}
	cdr := &utils.StoredCdr{CgrId: utils.Sha1("dsafdsaf", time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC).String()), OrderId: 1, AccId: "dsafdsaf", CdrHost: "192.168.1.1", ReqType: "rated", Direction: "*out", Tenant: "cgrates.org",
		Category: "call", Account: "1001", Subject: "1001", Destination: "1002",
		SetupTime:  time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC),
		AnswerTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC),
		Duration:   time.Duration(10) * time.Second, MediationRunId: utils.DEFAULT_RUNID, Cost: 2.34567,
		ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
	}
	if err := fwWriter.WriteCdr(cdr); err != nil {
		t.Error(err)
	}
	eContentOut := "201001        1001                1002                    0211  07111308420010          1       3dsafdsaf                             0002.34570\n"
	contentOut := fwWriter.content.String()
	if len(contentOut) != 145 {
		t.Error("Unexpected content length", len(contentOut))
	} else if contentOut != eContentOut {
		t.Errorf("Content out different than expected. Have <%s>, expecting: <%s>", contentOut, eContentOut)
	}
	eHeader := "10   VOI0000007111308420024031415390001                                                                                                         \n"
	eTrailer := "90   VOI0000000000100000010071113084260071113084200                                                                                             \n"
	outBeforeWrite := ""
	if wrBuf.String() != outBeforeWrite {
		t.Errorf("Output buffer should be empty before write")
	}
	fwWriter.Close()
	allOut := wrBuf.String()
	eAllOut := eHeader + eContentOut + eTrailer
	if math.Mod(float64(len(allOut)), 145) != 0 {
		t.Error("Unexpected export content length", len(allOut))
	} else if len(allOut) != len(eAllOut) {
		t.Errorf("Output does not match expected length. Have output %q, expecting: %q", allOut, eAllOut)
	}
	// Test stats
	if !fwWriter.firstCdrATime.Equal(cdr.AnswerTime) {
		t.Error("Unexpected firstCdrATime in stats: ", fwWriter.firstCdrATime)
	} else if !fwWriter.lastCdrATime.Equal(cdr.AnswerTime) {
		t.Error("Unexpected lastCdrATime in stats: ", fwWriter.lastCdrATime)
	} else if fwWriter.numberOfRecords != 1 {
		t.Error("Unexpected number of records in the stats: ", fwWriter.numberOfRecords)
	} else if fwWriter.totalDuration != cdr.Duration {
		t.Error("Unexpected total duration in the stats: ", fwWriter.totalDuration)
	} else if fwWriter.totalCost != utils.Round(cdr.Cost, fwWriter.roundDecimals, utils.ROUNDING_MIDDLE) {
		t.Error("Unexpected total cost in the stats: ", fwWriter.totalCost)
	}
	if fwWriter.FirstOrderId() != 1 {
		t.Error("Unexpected FirstOrderId", fwWriter.FirstOrderId())
	}
	if fwWriter.LastOrderId() != 1 {
		t.Error("Unexpected LastOrderId", fwWriter.LastOrderId())
	}
	if fwWriter.TotalCost() != utils.Round(cdr.Cost, fwWriter.roundDecimals, utils.ROUNDING_MIDDLE) {
		t.Error("Unexpected TotalCost: ", fwWriter.TotalCost())
	}
}

func TestWriteCdrs(t *testing.T) {
	wrBuf := &bytes.Buffer{}
	exportTpl := &config.CgrXmlCdreFwCfg{Header: &config.CgrXmlCfgCdrHeader{Fields: hdrCfgFlds},
		Content: &config.CgrXmlCfgCdrContent{Fields: contentCfgFlds},
		Trailer: &config.CgrXmlCfgCdrTrailer{Fields: trailerCfgFlds},
	}
	fwWriter := FixedWidthCdrWriter{writer: wrBuf, exportTemplate: exportTpl, roundDecimals: 4, header: &bytes.Buffer{}, content: &bytes.Buffer{}, trailer: &bytes.Buffer{}}
	cdr1 := &utils.StoredCdr{CgrId: utils.Sha1("aaa1", time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC).String()), OrderId: 2, AccId: "aaa1", CdrHost: "192.168.1.1", ReqType: "rated", Direction: "*out", Tenant: "cgrates.org",
		Category: "call", Account: "1001", Subject: "1001", Destination: "1010",
		SetupTime:  time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC),
		AnswerTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC),
		Duration:   time.Duration(10) * time.Second, MediationRunId: utils.DEFAULT_RUNID, Cost: 2.25,
		ExtraFields: map[string]string{"productnumber": "12341", "fieldextr2": "valextr2"},
	}
	cdr2 := &utils.StoredCdr{CgrId: utils.Sha1("aaa2", time.Date(2013, 11, 7, 7, 42, 20, 0, time.UTC).String()), OrderId: 4, AccId: "aaa2", CdrHost: "192.168.1.2", ReqType: "prepaid", Direction: "*out", Tenant: "cgrates.org",
		Category: "call", Account: "1002", Subject: "1002", Destination: "1011",
		SetupTime:  time.Date(2013, 11, 7, 7, 42, 20, 0, time.UTC),
		AnswerTime: time.Date(2013, 11, 7, 7, 42, 26, 0, time.UTC),
		Duration:   time.Duration(5) * time.Minute, MediationRunId: utils.DEFAULT_RUNID, Cost: 1.40001,
		ExtraFields: map[string]string{"productnumber": "12342", "fieldextr2": "valextr2"},
	}
	cdr3 := &utils.StoredCdr{}
	cdr4 := &utils.StoredCdr{CgrId: utils.Sha1("aaa3", time.Date(2013, 11, 7, 9, 42, 18, 0, time.UTC).String()), OrderId: 3, AccId: "aaa4", CdrHost: "192.168.1.4", ReqType: "postpaid", Direction: "*out", Tenant: "cgrates.org",
		Category: "call", Account: "1004", Subject: "1004", Destination: "1013",
		SetupTime:  time.Date(2013, 11, 7, 9, 42, 18, 0, time.UTC),
		AnswerTime: time.Date(2013, 11, 7, 9, 42, 26, 0, time.UTC),
		Duration:   time.Duration(20) * time.Second, MediationRunId: utils.DEFAULT_RUNID, Cost: 2.34567,
		ExtraFields: map[string]string{"productnumber": "12344", "fieldextr2": "valextr2"},
	}
	for _, cdr := range []*utils.StoredCdr{cdr1, cdr2, cdr3, cdr4} {
		if err := fwWriter.WriteCdr(cdr); err != nil {
			t.Error(err)
		}
		contentOut := fwWriter.content.String()
		if math.Mod(float64(len(contentOut)), 145) != 0 { // Rest must be 0 always, so content is always multiple of 145 which is our row fixLength
			t.Error("Unexpected content length", len(contentOut))
		}
	}
	if len(wrBuf.String()) != 0 {
		t.Errorf("Output buffer should be empty before write")
	}
	fwWriter.Close()
	if len(wrBuf.String()) != 725 {
		t.Error("Output buffer does not contain expected info. Expecting len: 725, got: ", len(wrBuf.String()))
	}
	// Test stats
	if !fwWriter.firstCdrATime.Equal(cdr2.AnswerTime) {
		t.Error("Unexpected firstCdrATime in stats: ", fwWriter.firstCdrATime)
	}
	if !fwWriter.lastCdrATime.Equal(cdr4.AnswerTime) {
		t.Error("Unexpected lastCdrATime in stats: ", fwWriter.lastCdrATime)
	}
	if fwWriter.numberOfRecords != 3 {
		t.Error("Unexpected number of records in the stats: ", fwWriter.numberOfRecords)
	}
	if fwWriter.totalDuration != time.Duration(330)*time.Second {
		t.Error("Unexpected total duration in the stats: ", fwWriter.totalDuration)
	}
	if fwWriter.totalCost != 5.9957 {
		t.Error("Unexpected total cost in the stats: ", fwWriter.totalCost)
	}
	if fwWriter.FirstOrderId() != 2 {
		t.Error("Unexpected FirstOrderId", fwWriter.FirstOrderId())
	}
	if fwWriter.LastOrderId() != 4 {
		t.Error("Unexpected LastOrderId", fwWriter.LastOrderId())
	}
	if fwWriter.TotalCost() != 5.9957 {
		t.Error("Unexpected TotalCost: ", fwWriter.TotalCost())
	}
}
