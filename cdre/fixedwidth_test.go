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

package cdre

import (
	"bytes"
	"math"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var hdrJsnCfgFlds = []*config.CdrFieldJsonCfg{
	&config.CdrFieldJsonCfg{Tag: utils.StringPointer("TypeOfRecord"), Type: utils.StringPointer(utils.CONSTANT), Value: utils.StringPointer("10"), Width: utils.IntPointer(2)},
	&config.CdrFieldJsonCfg{Tag: utils.StringPointer("Filler1"), Type: utils.StringPointer(utils.FILLER), Width: utils.IntPointer(3)},
	&config.CdrFieldJsonCfg{Tag: utils.StringPointer("DistributorCode"), Type: utils.StringPointer(utils.CONSTANT), Value: utils.StringPointer("VOI"), Width: utils.IntPointer(3)},
	&config.CdrFieldJsonCfg{Tag: utils.StringPointer("FileSeqNr"), Type: utils.StringPointer(utils.METATAG), Value: utils.StringPointer(META_EXPORTID),
		Width: utils.IntPointer(5), Strip: utils.StringPointer("right"), Padding: utils.StringPointer("zeroleft")},
	&config.CdrFieldJsonCfg{Tag: utils.StringPointer("LastCdr"), Type: utils.StringPointer(utils.METATAG), Value: utils.StringPointer(META_LASTCDRATIME),
		Width: utils.IntPointer(12), Layout: utils.StringPointer("020106150400")},
	&config.CdrFieldJsonCfg{Tag: utils.StringPointer("FileCreationfTime"), Type: utils.StringPointer(utils.METATAG), Value: utils.StringPointer(META_TIMENOW),
		Width: utils.IntPointer(12), Layout: utils.StringPointer("020106150400")},
	&config.CdrFieldJsonCfg{Tag: utils.StringPointer("FileVersion"), Type: utils.StringPointer(utils.CONSTANT), Value: utils.StringPointer("01"), Width: utils.IntPointer(2)},
	&config.CdrFieldJsonCfg{Tag: utils.StringPointer("Filler2"), Type: utils.StringPointer(utils.FILLER), Width: utils.IntPointer(105)},
}

var contentJsnCfgFlds = []*config.CdrFieldJsonCfg{
	&config.CdrFieldJsonCfg{Tag: utils.StringPointer("TypeOfRecord"), Type: utils.StringPointer(utils.CONSTANT), Value: utils.StringPointer("20"), Width: utils.IntPointer(2)},
	&config.CdrFieldJsonCfg{Tag: utils.StringPointer("Account"), Type: utils.StringPointer(utils.CDRFIELD), Value: utils.StringPointer(utils.ACCOUNT), Width: utils.IntPointer(12),
		Strip: utils.StringPointer("left"), Padding: utils.StringPointer("right")},
	&config.CdrFieldJsonCfg{Tag: utils.StringPointer("Subject"), Type: utils.StringPointer(utils.CDRFIELD), Value: utils.StringPointer(utils.SUBJECT), Width: utils.IntPointer(5),
		Strip: utils.StringPointer("right"), Padding: utils.StringPointer("right")},
	&config.CdrFieldJsonCfg{Tag: utils.StringPointer("CLI"), Type: utils.StringPointer(utils.CDRFIELD), Value: utils.StringPointer("cli"), Width: utils.IntPointer(15),
		Strip: utils.StringPointer("xright"), Padding: utils.StringPointer("right")},
	&config.CdrFieldJsonCfg{Tag: utils.StringPointer("Destination"), Type: utils.StringPointer(utils.CDRFIELD), Value: utils.StringPointer(utils.DESTINATION), Width: utils.IntPointer(24),
		Strip: utils.StringPointer("xright"), Padding: utils.StringPointer("right")},
	&config.CdrFieldJsonCfg{Tag: utils.StringPointer("TOR"), Type: utils.StringPointer(utils.CONSTANT), Value: utils.StringPointer("02"), Width: utils.IntPointer(2)},
	&config.CdrFieldJsonCfg{Tag: utils.StringPointer("SubtypeTOR"), Type: utils.StringPointer(utils.CONSTANT), Value: utils.StringPointer("11"), Width: utils.IntPointer(4),
		Padding: utils.StringPointer("right")},
	&config.CdrFieldJsonCfg{Tag: utils.StringPointer("SetupTime"), Type: utils.StringPointer(utils.CDRFIELD), Value: utils.StringPointer(utils.SETUP_TIME), Width: utils.IntPointer(12),
		Strip: utils.StringPointer("right"), Padding: utils.StringPointer("right"), Layout: utils.StringPointer("020106150400")},
	&config.CdrFieldJsonCfg{Tag: utils.StringPointer("Duration"), Type: utils.StringPointer(utils.CDRFIELD), Value: utils.StringPointer(utils.USAGE), Width: utils.IntPointer(6),
		Strip: utils.StringPointer("right"), Padding: utils.StringPointer("right"), Layout: utils.StringPointer(utils.SECONDS)},
	&config.CdrFieldJsonCfg{Tag: utils.StringPointer("DataVolume"), Type: utils.StringPointer(utils.FILLER), Width: utils.IntPointer(6)},
	&config.CdrFieldJsonCfg{Tag: utils.StringPointer("TaxCode"), Type: utils.StringPointer(utils.CONSTANT), Value: utils.StringPointer("1"), Width: utils.IntPointer(1)},
	&config.CdrFieldJsonCfg{Tag: utils.StringPointer("OperatorCode"), Type: utils.StringPointer(utils.CDRFIELD), Value: utils.StringPointer("opercode"), Width: utils.IntPointer(2),
		Strip: utils.StringPointer("right"), Padding: utils.StringPointer("right")},
	&config.CdrFieldJsonCfg{Tag: utils.StringPointer("ProductId"), Type: utils.StringPointer(utils.CDRFIELD), Value: utils.StringPointer("productid"), Width: utils.IntPointer(5),
		Strip: utils.StringPointer("right"), Padding: utils.StringPointer("right")},
	&config.CdrFieldJsonCfg{Tag: utils.StringPointer("NetworkId"), Type: utils.StringPointer(utils.CONSTANT), Value: utils.StringPointer("3"), Width: utils.IntPointer(1)},
	&config.CdrFieldJsonCfg{Tag: utils.StringPointer("CallId"), Type: utils.StringPointer(utils.CDRFIELD), Value: utils.StringPointer(utils.ACCID), Width: utils.IntPointer(16),
		Padding: utils.StringPointer("right")},
	&config.CdrFieldJsonCfg{Tag: utils.StringPointer("Filler"), Type: utils.StringPointer(utils.FILLER), Width: utils.IntPointer(8)},
	&config.CdrFieldJsonCfg{Tag: utils.StringPointer("Filler"), Type: utils.StringPointer(utils.FILLER), Width: utils.IntPointer(8)},
	&config.CdrFieldJsonCfg{Tag: utils.StringPointer("TerminationCode"), Type: utils.StringPointer(utils.CDRFIELD), Value: utils.StringPointer("operator;product"),
		Width: utils.IntPointer(5), Strip: utils.StringPointer("right"), Padding: utils.StringPointer("right")},
	&config.CdrFieldJsonCfg{Tag: utils.StringPointer("Cost"), Type: utils.StringPointer(utils.CDRFIELD), Value: utils.StringPointer(utils.COST), Width: utils.IntPointer(9),
		Padding: utils.StringPointer("zeroleft")},
	&config.CdrFieldJsonCfg{Tag: utils.StringPointer("DestinationPrivacy"), Type: utils.StringPointer(utils.METATAG), Value: utils.StringPointer(META_MASKDESTINATION),
		Width: utils.IntPointer(1)},
}

var trailerJsnCfgFlds = []*config.CdrFieldJsonCfg{
	&config.CdrFieldJsonCfg{Tag: utils.StringPointer("TypeOfRecord"), Type: utils.StringPointer(utils.CONSTANT), Value: utils.StringPointer("90"), Width: utils.IntPointer(2)},
	&config.CdrFieldJsonCfg{Tag: utils.StringPointer("Filler1"), Type: utils.StringPointer(utils.FILLER), Width: utils.IntPointer(3)},
	&config.CdrFieldJsonCfg{Tag: utils.StringPointer("DistributorCode"), Type: utils.StringPointer(utils.CONSTANT), Value: utils.StringPointer("VOI"), Width: utils.IntPointer(3)},
	&config.CdrFieldJsonCfg{Tag: utils.StringPointer("FileSeqNr"), Type: utils.StringPointer(utils.METATAG), Value: utils.StringPointer(META_EXPORTID), Width: utils.IntPointer(5),
		Strip: utils.StringPointer("right"), Padding: utils.StringPointer("zeroleft")},
	&config.CdrFieldJsonCfg{Tag: utils.StringPointer("NumberOfRecords"), Type: utils.StringPointer(utils.METATAG), Value: utils.StringPointer(META_NRCDRS),
		Width: utils.IntPointer(6), Padding: utils.StringPointer("zeroleft")},
	&config.CdrFieldJsonCfg{Tag: utils.StringPointer("CdrsDuration"), Type: utils.StringPointer(utils.METATAG), Value: utils.StringPointer(META_DURCDRS),
		Width: utils.IntPointer(8), Padding: utils.StringPointer("zeroleft"), Layout: utils.StringPointer(utils.SECONDS)},
	&config.CdrFieldJsonCfg{Tag: utils.StringPointer("FirstCdrTime"), Type: utils.StringPointer(utils.METATAG), Value: utils.StringPointer(META_FIRSTCDRATIME),
		Width: utils.IntPointer(12), Layout: utils.StringPointer("020106150400")},
	&config.CdrFieldJsonCfg{Tag: utils.StringPointer("LastCdrTime"), Type: utils.StringPointer(utils.METATAG), Value: utils.StringPointer(META_LASTCDRATIME),
		Width: utils.IntPointer(12), Layout: utils.StringPointer("020106150400")},
	&config.CdrFieldJsonCfg{Tag: utils.StringPointer("Filler2"), Type: utils.StringPointer(utils.FILLER), Width: utils.IntPointer(93)},
}

var hdrCfgFlds, contentCfgFlds, trailerCfgFlds []*config.CfgCdrField

// Write one CDR and test it's results only for content buffer
func TestWriteCdr(t *testing.T) {
	var err error
	wrBuf := &bytes.Buffer{}
	cfg, _ := config.NewDefaultCGRConfig()
	if hdrCfgFlds, err = config.CfgCdrFieldsFromCdrFieldsJsonCfg(hdrJsnCfgFlds); err != nil {
		t.Error(err)
	}
	if contentCfgFlds, err = config.CfgCdrFieldsFromCdrFieldsJsonCfg(contentJsnCfgFlds); err != nil {
		t.Error(err)
	}
	if trailerCfgFlds, err = config.CfgCdrFieldsFromCdrFieldsJsonCfg(trailerJsnCfgFlds); err != nil {
		t.Error(err)
	}
	cdreCfg := &config.CdreConfig{
		CdrFormat:     utils.CDRE_FIXED_WIDTH,
		HeaderFields:  hdrCfgFlds,
		ContentFields: contentCfgFlds,
		TrailerFields: trailerCfgFlds,
	}
	cdr := &engine.StoredCdr{CgrId: utils.Sha1("dsafdsaf", time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC).String()),
		TOR: utils.VOICE, OrderId: 1, AccId: "dsafdsaf", CdrHost: "192.168.1.1",
		ReqType: utils.META_RATED, Direction: "*out", Tenant: "cgrates.org",
		Category: "call", Account: "1001", Subject: "1001", Destination: "1002",
		SetupTime:  time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC),
		AnswerTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC),
		Usage:      time.Duration(10) * time.Second, MediationRunId: utils.DEFAULT_RUNID, Cost: 2.34567,
		ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
	}
	cdre, err := NewCdrExporter([]*engine.StoredCdr{cdr}, nil, cdreCfg, utils.CDRE_FIXED_WIDTH, ',', "fwv_1", 0.0, 0.0, 0.0, 0.0, 0, 4,
		cfg.RoundingDecimals, "", -1, cfg.HttpSkipTlsVerify, "")
	if err != nil {
		t.Error(err)
	}
	eHeader := "10   VOIfwv_107111308420018011511340001                                                                                                         \n"
	eContentOut := "201001        1001                1002                    0211  07111308420010          1       3dsafdsaf                             0002.34570\n"
	eTrailer := "90   VOIfwv_100000100000010071113084200071113084200                                                                                             \n"
	if err := cdre.writeOut(wrBuf); err != nil {
		t.Error(err)
	}
	allOut := wrBuf.String()
	eAllOut := eHeader + eContentOut + eTrailer
	if math.Mod(float64(len(allOut)), 145) != 0 {
		t.Error("Unexpected export content length", len(allOut))
	} else if len(allOut) != len(eAllOut) {
		t.Errorf("Output does not match expected length. Have output %q, expecting: %q", allOut, eAllOut)
	}
	// Test stats
	if !cdre.firstCdrATime.Equal(cdr.AnswerTime) {
		t.Error("Unexpected firstCdrATime in stats: ", cdre.firstCdrATime)
	} else if !cdre.lastCdrATime.Equal(cdr.AnswerTime) {
		t.Error("Unexpected lastCdrATime in stats: ", cdre.lastCdrATime)
	} else if cdre.numberOfRecords != 1 {
		t.Error("Unexpected number of records in the stats: ", cdre.numberOfRecords)
	} else if cdre.totalDuration != cdr.Usage {
		t.Error("Unexpected total duration in the stats: ", cdre.totalDuration)
	} else if cdre.totalCost != utils.Round(cdr.Cost, cdre.roundDecimals, utils.ROUNDING_MIDDLE) {
		t.Error("Unexpected total cost in the stats: ", cdre.totalCost)
	}
	if cdre.FirstOrderId() != 1 {
		t.Error("Unexpected FirstOrderId", cdre.FirstOrderId())
	}
	if cdre.LastOrderId() != 1 {
		t.Error("Unexpected LastOrderId", cdre.LastOrderId())
	}
	if cdre.TotalCost() != utils.Round(cdr.Cost, cdre.roundDecimals, utils.ROUNDING_MIDDLE) {
		t.Error("Unexpected TotalCost: ", cdre.TotalCost())
	}
}

func TestWriteCdrs(t *testing.T) {
	wrBuf := &bytes.Buffer{}
	cdreCfg := &config.CdreConfig{
		CdrFormat:     utils.CDRE_FIXED_WIDTH,
		HeaderFields:  hdrCfgFlds,
		ContentFields: contentCfgFlds,
		TrailerFields: trailerCfgFlds,
	}
	cdr1 := &engine.StoredCdr{CgrId: utils.Sha1("aaa1", time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC).String()),
		TOR: utils.VOICE, OrderId: 2, AccId: "aaa1", CdrHost: "192.168.1.1", ReqType: utils.META_RATED, Direction: "*out", Tenant: "cgrates.org",
		Category: "call", Account: "1001", Subject: "1001", Destination: "1010",
		SetupTime:  time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC),
		AnswerTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC),
		Usage:      time.Duration(10) * time.Second, MediationRunId: utils.DEFAULT_RUNID, Cost: 2.25,
		ExtraFields: map[string]string{"productnumber": "12341", "fieldextr2": "valextr2"},
	}
	cdr2 := &engine.StoredCdr{CgrId: utils.Sha1("aaa2", time.Date(2013, 11, 7, 7, 42, 20, 0, time.UTC).String()),
		TOR: utils.VOICE, OrderId: 4, AccId: "aaa2", CdrHost: "192.168.1.2", ReqType: utils.META_PREPAID, Direction: "*out", Tenant: "cgrates.org",
		Category: "call", Account: "1002", Subject: "1002", Destination: "1011",
		SetupTime:  time.Date(2013, 11, 7, 7, 42, 20, 0, time.UTC),
		AnswerTime: time.Date(2013, 11, 7, 7, 42, 26, 0, time.UTC),
		Usage:      time.Duration(5) * time.Minute, MediationRunId: utils.DEFAULT_RUNID, Cost: 1.40001,
		ExtraFields: map[string]string{"productnumber": "12342", "fieldextr2": "valextr2"},
	}
	cdr3 := &engine.StoredCdr{}
	cdr4 := &engine.StoredCdr{CgrId: utils.Sha1("aaa3", time.Date(2013, 11, 7, 9, 42, 18, 0, time.UTC).String()),
		TOR: utils.VOICE, OrderId: 3, AccId: "aaa4", CdrHost: "192.168.1.4", ReqType: utils.META_POSTPAID, Direction: "*out", Tenant: "cgrates.org",
		Category: "call", Account: "1004", Subject: "1004", Destination: "1013",
		SetupTime:  time.Date(2013, 11, 7, 9, 42, 18, 0, time.UTC),
		AnswerTime: time.Date(2013, 11, 7, 9, 42, 26, 0, time.UTC),
		Usage:      time.Duration(20) * time.Second, MediationRunId: utils.DEFAULT_RUNID, Cost: 2.34567,
		ExtraFields: map[string]string{"productnumber": "12344", "fieldextr2": "valextr2"},
	}
	cfg, _ := config.NewDefaultCGRConfig()
	cdre, err := NewCdrExporter([]*engine.StoredCdr{cdr1, cdr2, cdr3, cdr4}, nil, cdreCfg, utils.CDRE_FIXED_WIDTH, ',',
		"fwv_1", 0.0, 0.0, 0.0, 0.0, 0, 4, cfg.RoundingDecimals, "", -1, cfg.HttpSkipTlsVerify, "")
	if err != nil {
		t.Error(err)
	}
	if err := cdre.writeOut(wrBuf); err != nil {
		t.Error(err)
	}
	if len(wrBuf.String()) != 725 {
		t.Error("Output buffer does not contain expected info. Expecting len: 725, got: ", len(wrBuf.String()))
	}
	// Test stats
	if !cdre.firstCdrATime.Equal(cdr2.AnswerTime) {
		t.Error("Unexpected firstCdrATime in stats: ", cdre.firstCdrATime)
	}
	if !cdre.lastCdrATime.Equal(cdr4.AnswerTime) {
		t.Error("Unexpected lastCdrATime in stats: ", cdre.lastCdrATime)
	}
	if cdre.numberOfRecords != 3 {
		t.Error("Unexpected number of records in the stats: ", cdre.numberOfRecords)
	}
	if cdre.totalDuration != time.Duration(330)*time.Second {
		t.Error("Unexpected total duration in the stats: ", cdre.totalDuration)
	}
	if cdre.totalCost != 5.9957 {
		t.Error("Unexpected total cost in the stats: ", cdre.totalCost)
	}
	if cdre.FirstOrderId() != 2 {
		t.Error("Unexpected FirstOrderId", cdre.FirstOrderId())
	}
	if cdre.LastOrderId() != 4 {
		t.Error("Unexpected LastOrderId", cdre.LastOrderId())
	}
	if cdre.TotalCost() != 5.9957 {
		t.Error("Unexpected TotalCost: ", cdre.TotalCost())
	}
}
