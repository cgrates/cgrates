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
package engine

import (
	"bytes"
	"math"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

var hdrJsnCfgFlds = []*config.FcTemplateJsonCfg{
	{
		Tag:   utils.StringPointer("TypeOfRecord"),
		Type:  utils.StringPointer(utils.META_CONSTANT),
		Value: utils.StringPointer("10"),
		Width: utils.IntPointer(2)},
	{
		Tag:   utils.StringPointer("Filler1"),
		Type:  utils.StringPointer(utils.META_FILLER),
		Width: utils.IntPointer(3)},
	{
		Tag:   utils.StringPointer("DistributorCode"),
		Type:  utils.StringPointer(utils.META_CONSTANT),
		Value: utils.StringPointer("VOI"),
		Width: utils.IntPointer(3)},
	{
		Tag:     utils.StringPointer("FileSeqNr"),
		Type:    utils.StringPointer(utils.META_HANDLER),
		Value:   utils.StringPointer(META_EXPORTID),
		Width:   utils.IntPointer(5),
		Strip:   utils.StringPointer("right"),
		Padding: utils.StringPointer("zeroleft")},
	{
		Tag:    utils.StringPointer("LastCdr"),
		Type:   utils.StringPointer(utils.META_HANDLER),
		Width:  utils.IntPointer(12),
		Value:  utils.StringPointer(META_LASTCDRATIME),
		Layout: utils.StringPointer("020106150400")},
	{
		Tag:    utils.StringPointer("FileCreationfTime"),
		Type:   utils.StringPointer(utils.META_HANDLER),
		Value:  utils.StringPointer(META_TIMENOW),
		Width:  utils.IntPointer(12),
		Layout: utils.StringPointer("020106150400")},
	{
		Tag:   utils.StringPointer("FileVersion"),
		Type:  utils.StringPointer(utils.META_CONSTANT),
		Value: utils.StringPointer("01"),
		Width: utils.IntPointer(2)},
	{
		Tag:   utils.StringPointer("Filler2"),
		Type:  utils.StringPointer(utils.META_FILLER),
		Width: utils.IntPointer(105)},
}

var contentJsnCfgFlds = []*config.FcTemplateJsonCfg{
	{
		Tag:   utils.StringPointer("TypeOfRecord"),
		Type:  utils.StringPointer(utils.META_CONSTANT),
		Value: utils.StringPointer("20"),
		Width: utils.IntPointer(2)},
	{
		Tag:     utils.StringPointer("Account"),
		Type:    utils.StringPointer(utils.META_COMPOSED),
		Value:   utils.StringPointer("~" + utils.Account),
		Width:   utils.IntPointer(12),
		Strip:   utils.StringPointer("left"),
		Padding: utils.StringPointer("right")},
	{
		Tag:     utils.StringPointer("Subject"),
		Type:    utils.StringPointer(utils.META_COMPOSED),
		Value:   utils.StringPointer("~" + utils.Subject),
		Width:   utils.IntPointer(5),
		Strip:   utils.StringPointer("right"),
		Padding: utils.StringPointer("right")},
	{
		Tag:     utils.StringPointer("CLI"),
		Type:    utils.StringPointer(utils.META_COMPOSED),
		Width:   utils.IntPointer(15),
		Value:   utils.StringPointer("cli"),
		Strip:   utils.StringPointer("xright"),
		Padding: utils.StringPointer("right")},
	{
		Tag:     utils.StringPointer("Destination"),
		Type:    utils.StringPointer(utils.META_COMPOSED),
		Value:   utils.StringPointer("~" + utils.Destination),
		Width:   utils.IntPointer(24),
		Strip:   utils.StringPointer("xright"),
		Padding: utils.StringPointer("right")},
	{
		Tag:   utils.StringPointer("ToR"),
		Type:  utils.StringPointer(utils.META_CONSTANT),
		Value: utils.StringPointer("02"),
		Width: utils.IntPointer(2)},
	{
		Tag:     utils.StringPointer("SubtypeTOR"),
		Type:    utils.StringPointer(utils.META_CONSTANT),
		Value:   utils.StringPointer("11"),
		Padding: utils.StringPointer("right"),
		Width:   utils.IntPointer(4)},
	{
		Tag:     utils.StringPointer("SetupTime"),
		Type:    utils.StringPointer(utils.META_COMPOSED),
		Value:   utils.StringPointer("~" + utils.SetupTime),
		Width:   utils.IntPointer(12),
		Strip:   utils.StringPointer("right"),
		Padding: utils.StringPointer("right"),
		Layout:  utils.StringPointer("020106150400")},
	{
		Tag:     utils.StringPointer("Duration"),
		Type:    utils.StringPointer(utils.META_COMPOSED),
		Value:   utils.StringPointer("~" + utils.Usage),
		Width:   utils.IntPointer(6),
		Strip:   utils.StringPointer("right"),
		Padding: utils.StringPointer("right"),
		Layout:  utils.StringPointer(utils.SECONDS)},
	{
		Tag:   utils.StringPointer("DataVolume"),
		Type:  utils.StringPointer(utils.META_FILLER),
		Width: utils.IntPointer(6)},
	{
		Tag:   utils.StringPointer("TaxCode"),
		Type:  utils.StringPointer(utils.META_CONSTANT),
		Value: utils.StringPointer("1"),
		Width: utils.IntPointer(1)},
	{
		Tag:     utils.StringPointer("OperatorCode"),
		Type:    utils.StringPointer(utils.META_COMPOSED),
		Value:   utils.StringPointer("opercode"),
		Width:   utils.IntPointer(2),
		Strip:   utils.StringPointer("right"),
		Padding: utils.StringPointer("right")},
	{
		Tag:     utils.StringPointer("ProductId"),
		Type:    utils.StringPointer(utils.META_COMPOSED),
		Value:   utils.StringPointer("~productid"),
		Width:   utils.IntPointer(5),
		Strip:   utils.StringPointer("right"),
		Padding: utils.StringPointer("right")},
	{
		Tag:   utils.StringPointer("NetworkId"),
		Type:  utils.StringPointer(utils.META_CONSTANT),
		Value: utils.StringPointer("3"),
		Width: utils.IntPointer(1)},
	{
		Tag:     utils.StringPointer("CallId"),
		Type:    utils.StringPointer(utils.META_COMPOSED),
		Value:   utils.StringPointer("~" + utils.OriginID),
		Width:   utils.IntPointer(16),
		Padding: utils.StringPointer("right")},
	{
		Tag:   utils.StringPointer("Filler"),
		Type:  utils.StringPointer(utils.META_FILLER),
		Width: utils.IntPointer(8)},
	{
		Tag:   utils.StringPointer("Filler"),
		Type:  utils.StringPointer(utils.META_FILLER),
		Width: utils.IntPointer(8)},
	{
		Tag:     utils.StringPointer("TerminationCode"),
		Type:    utils.StringPointer(utils.META_COMPOSED),
		Value:   utils.StringPointer("~operator;~product"),
		Width:   utils.IntPointer(5),
		Strip:   utils.StringPointer("right"),
		Padding: utils.StringPointer("right")},
	{
		Tag:     utils.StringPointer("Cost"),
		Type:    utils.StringPointer(utils.META_COMPOSED),
		Width:   utils.IntPointer(9),
		Value:   utils.StringPointer("~" + utils.COST),
		Padding: utils.StringPointer("zeroleft")},
	{
		Tag:   utils.StringPointer("DestinationPrivacy"),
		Type:  utils.StringPointer(utils.MetaMaskedDestination),
		Width: utils.IntPointer(1)},
}

var trailerJsnCfgFlds = []*config.FcTemplateJsonCfg{
	{
		Tag:   utils.StringPointer("TypeOfRecord"),
		Type:  utils.StringPointer(utils.META_CONSTANT),
		Value: utils.StringPointer("90"),
		Width: utils.IntPointer(2)},
	{
		Tag:   utils.StringPointer("Filler1"),
		Type:  utils.StringPointer(utils.META_FILLER),
		Width: utils.IntPointer(3)},
	{
		Tag:   utils.StringPointer("DistributorCode"),
		Type:  utils.StringPointer(utils.META_CONSTANT),
		Value: utils.StringPointer("VOI"),
		Width: utils.IntPointer(3)},
	{
		Tag:     utils.StringPointer("FileSeqNr"),
		Type:    utils.StringPointer(utils.META_HANDLER),
		Value:   utils.StringPointer(META_EXPORTID),
		Width:   utils.IntPointer(5),
		Strip:   utils.StringPointer("right"),
		Padding: utils.StringPointer("zeroleft")},
	{
		Tag:     utils.StringPointer("NumberOfRecords"),
		Type:    utils.StringPointer(utils.META_HANDLER),
		Value:   utils.StringPointer(META_NRCDRS),
		Width:   utils.IntPointer(6),
		Padding: utils.StringPointer("zeroleft")},
	{
		Tag:     utils.StringPointer("CdrsDuration"),
		Type:    utils.StringPointer(utils.META_HANDLER),
		Value:   utils.StringPointer(META_DURCDRS),
		Width:   utils.IntPointer(8),
		Padding: utils.StringPointer("zeroleft"),
		Layout:  utils.StringPointer(utils.SECONDS)},
	{
		Tag:    utils.StringPointer("FirstCdrTime"),
		Type:   utils.StringPointer(utils.META_HANDLER),
		Width:  utils.IntPointer(12),
		Value:  utils.StringPointer(META_FIRSTCDRATIME),
		Layout: utils.StringPointer("020106150400")},
	{
		Tag:    utils.StringPointer("LastCdrTime"),
		Type:   utils.StringPointer(utils.META_HANDLER),
		Width:  utils.IntPointer(12),
		Value:  utils.StringPointer(META_LASTCDRATIME),
		Layout: utils.StringPointer("020106150400")},
	{
		Tag:   utils.StringPointer("Filler2"),
		Type:  utils.StringPointer(utils.META_FILLER),
		Width: utils.IntPointer(93)},
}

var hdrCfgFlds, contentCfgFlds, trailerCfgFlds []*config.FCTemplate

// Write one CDR and test it's results only for content buffer
func TestWriteCdr(t *testing.T) {
	var err error
	wrBuf := &bytes.Buffer{}
	cfg, _ := config.NewDefaultCGRConfig()
	if hdrCfgFlds, err = config.FCTemplatesFromFCTemplatesJsonCfg(hdrJsnCfgFlds, utils.INFIELD_SEP); err != nil {
		t.Error(err)
	}
	if contentCfgFlds, err = config.FCTemplatesFromFCTemplatesJsonCfg(contentJsnCfgFlds, utils.INFIELD_SEP); err != nil {
		t.Error(err)
	}
	if trailerCfgFlds, err = config.FCTemplatesFromFCTemplatesJsonCfg(trailerJsnCfgFlds, utils.INFIELD_SEP); err != nil {
		t.Error(err)
	}
	cdreCfg := &config.CdreCfg{
		ExportFormat:  utils.MetaFileFWV,
		HeaderFields:  hdrCfgFlds,
		ContentFields: contentCfgFlds,
		TrailerFields: trailerCfgFlds,
	}
	cdr := &CDR{
		CGRID: utils.Sha1("dsafdsaf", time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC).String()),
		ToR:   utils.VOICE, OrderID: 1, OriginID: "dsafdsaf", OriginHost: "192.168.1.1",
		RequestType: utils.META_RATED, Tenant: "cgrates.org", Category: "call",
		Account: "1001", Subject: "1001", Destination: "1002",
		SetupTime:  time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC),
		AnswerTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC),
		Usage:      time.Duration(10) * time.Second,
		RunID:      utils.DEFAULT_RUNID, Cost: 2.34567,
		ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
	}

	cdre, err := NewCDRExporter([]*CDR{cdr}, cdreCfg, utils.MetaFileFWV, "", "", "fwv_1",
		true, 1, '|', map[string]float64{}, 0.0, cfg.GeneralCfg().RoundingDecimals,
		cfg.GeneralCfg().HttpSkipTlsVerify, nil, nil)
	if err != nil {
		t.Error(err)
	}
	if err = cdre.processCDRs(); err != nil {
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
		t.Errorf("Unexpected export content length %d, have output \n%q, \n expecting: \n%q",
			len(allOut), allOut, eAllOut)
	} else if len(allOut) != len(eAllOut) {
		t.Errorf("Output does not match expected length. Have output \n%q, \n expecting: \n%q",
			allOut, eAllOut)
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
	} else if cdre.totalCost != utils.Round(cdr.Cost, cdre.roundingDecimals, utils.ROUNDING_MIDDLE) {
		t.Error("Unexpected total cost in the stats: ", cdre.totalCost)
	}

	if cdre.FirstOrderId() != 1 {
		t.Error("Unexpected FirstOrderId", cdre.FirstOrderId())
	}
	if cdre.LastOrderId() != 1 {
		t.Error("Unexpected LastOrderId", cdre.LastOrderId())
	}
	if cdre.TotalCost() != utils.Round(cdr.Cost, cdre.roundingDecimals, utils.ROUNDING_MIDDLE) {
		t.Error("Unexpected TotalCost: ", cdre.TotalCost())
	}
}

func TestWriteCdrs(t *testing.T) {
	wrBuf := &bytes.Buffer{}
	cdreCfg := &config.CdreCfg{
		ExportFormat:  utils.MetaFileFWV,
		HeaderFields:  hdrCfgFlds,
		ContentFields: contentCfgFlds,
		TrailerFields: trailerCfgFlds,
	}
	cdr1 := &CDR{CGRID: utils.Sha1("aaa1", time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC).String()),
		ToR: utils.VOICE, OrderID: 2, OriginID: "aaa1", OriginHost: "192.168.1.1",
		RequestType: utils.META_RATED, Tenant: "cgrates.org",
		Category: "call", Account: "1001", Subject: "1001", Destination: "1010",
		SetupTime:  time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC),
		AnswerTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC),
		Usage:      time.Duration(10) * time.Second, RunID: utils.DEFAULT_RUNID, Cost: 2.25,
		ExtraFields: map[string]string{"productnumber": "12341", "fieldextr2": "valextr2"},
	}
	cdr2 := &CDR{CGRID: utils.Sha1("aaa2", time.Date(2013, 11, 7, 7, 42, 20, 0, time.UTC).String()),
		ToR: utils.VOICE, OrderID: 4, OriginID: "aaa2", OriginHost: "192.168.1.2",
		RequestType: utils.META_PREPAID, Tenant: "cgrates.org",
		Category: "call", Account: "1002", Subject: "1002", Destination: "1011",
		SetupTime:  time.Date(2013, 11, 7, 7, 42, 20, 0, time.UTC),
		AnswerTime: time.Date(2013, 11, 7, 7, 42, 26, 0, time.UTC),
		Usage:      time.Duration(5) * time.Minute,
		RunID:      utils.DEFAULT_RUNID, Cost: 1.40001,
		ExtraFields: map[string]string{"productnumber": "12342", "fieldextr2": "valextr2"},
	}
	cdr3 := &CDR{}
	cdr4 := &CDR{CGRID: utils.Sha1("aaa3", time.Date(2013, 11, 7, 9, 42, 18, 0, time.UTC).String()),
		ToR: utils.VOICE, OrderID: 3, OriginID: "aaa4", OriginHost: "192.168.1.4",
		RequestType: utils.META_POSTPAID, Tenant: "cgrates.org",
		Category: "call", Account: "1004", Subject: "1004", Destination: "1013",
		SetupTime:  time.Date(2013, 11, 7, 9, 42, 18, 0, time.UTC),
		AnswerTime: time.Date(2013, 11, 7, 9, 42, 26, 0, time.UTC),
		Usage:      time.Duration(20) * time.Second,
		RunID:      utils.DEFAULT_RUNID, Cost: 2.34567,
		ExtraFields: map[string]string{"productnumber": "12344", "fieldextr2": "valextr2"},
	}
	cfg, _ := config.NewDefaultCGRConfig()
	cdre, err := NewCDRExporter([]*CDR{cdr1, cdr2, cdr3, cdr4}, cdreCfg,
		utils.MetaFileFWV, "", "", "fwv_1", true, 1, ',', map[string]float64{},
		0.0, cfg.GeneralCfg().RoundingDecimals,
		cfg.GeneralCfg().HttpSkipTlsVerify, nil, nil)
	if err != nil {
		t.Error(err)
	}
	if err = cdre.processCDRs(); err != nil {
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
	if cdre.totalCost != 5.99568 {
		t.Error("Unexpected total cost in the stats: ", cdre.totalCost)
	}
	if cdre.FirstOrderId() != 2 {
		t.Error("Unexpected FirstOrderId", cdre.FirstOrderId())
	}
	if cdre.LastOrderId() != 4 {
		t.Error("Unexpected LastOrderId", cdre.LastOrderId())
	}
	if cdre.TotalCost() != 5.99568 {
		t.Error("Unexpected TotalCost: ", cdre.TotalCost())
	}
}
