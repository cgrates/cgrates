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
	"encoding/csv"
	"strings"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

func TestCsvCdrWriter(t *testing.T) {
	writer := &bytes.Buffer{}
	cfg, _ := config.NewDefaultCGRConfig()
	storedCdr1 := &CDR{
		CGRID: utils.Sha1("dsafdsaf", time.Unix(1383813745, 0).UTC().String()),
		ToR:   utils.VOICE, OriginID: "dsafdsaf", OriginHost: "192.168.1.1",
		RequestType: utils.META_RATED, Tenant: "cgrates.org", Category: "call",
		Account: "1001", Subject: "1001", Destination: "1002",
		SetupTime:  time.Unix(1383813745, 0).UTC(),
		AnswerTime: time.Unix(1383813746, 0).UTC(),
		Usage:      time.Duration(10) * time.Second,
		RunID:      utils.DEFAULT_RUNID, Cost: 1.01,
		ExtraFields: map[string]string{"extra1": "val_extra1",
			"extra2": "val_extra2", "extra3": "val_extra3"},
	}
	cdre, err := NewCDRExporter([]*CDR{storedCdr1},
		cfg.CdreProfiles["*default"], utils.MetaFileCSV,
		"", "", "firstexport",
		true, 1, ',', map[string]float64{}, 0.0,
		cfg.RoundingDecimals, cfg.HttpSkipTlsVerify, nil)
	if err != nil {
		t.Error("Unexpected error received: ", err)
	}
	if err = cdre.processCDRs(); err != nil {
		t.Error(err)
	}
	csvWriter := csv.NewWriter(writer)
	if err := cdre.writeCsv(csvWriter); err != nil {
		t.Error("Unexpected error: ", err)
	}
	expected := `dbafe9c8614c785a65aabd116dd3959c3c56f7f6,*default,*voice,dsafdsaf,*rated,cgrates.org,call,1001,1001,1002,2013-11-07T08:42:25Z,2013-11-07T08:42:26Z,10s,1.01000`
	result := strings.TrimSpace(writer.String())
	if result != expected {
		t.Errorf("Expected: \n%s received: \n%s.", expected, result)
	}
	if cdre.TotalCost() != 1.01 {
		t.Error("Unexpected TotalCost: ", cdre.TotalCost())
	}
}

func TestAlternativeFieldSeparator(t *testing.T) {
	writer := &bytes.Buffer{}
	cfg, _ := config.NewDefaultCGRConfig()
	storedCdr1 := &CDR{
		CGRID: utils.Sha1("dsafdsaf", time.Unix(1383813745, 0).UTC().String()),
		ToR:   utils.VOICE, OriginID: "dsafdsaf", OriginHost: "192.168.1.1",
		RequestType: utils.META_RATED, Tenant: "cgrates.org", Category: "call",
		Account: "1001", Subject: "1001", Destination: "1002",
		SetupTime:  time.Unix(1383813745, 0).UTC(),
		AnswerTime: time.Unix(1383813746, 0).UTC(),
		Usage:      time.Duration(10) * time.Second,
		RunID:      utils.DEFAULT_RUNID, Cost: 1.01,
		ExtraFields: map[string]string{"extra1": "val_extra1",
			"extra2": "val_extra2", "extra3": "val_extra3"},
	}
	cdre, err := NewCDRExporter([]*CDR{storedCdr1}, cfg.CdreProfiles["*default"],
		utils.MetaFileCSV, "", "", "firstexport",
		true, 1, '|', map[string]float64{}, 0.0,
		cfg.RoundingDecimals, cfg.HttpSkipTlsVerify, nil)
	if err != nil {
		t.Error("Unexpected error received: ", err)
	}
	if err = cdre.processCDRs(); err != nil {
		t.Error(err)
	}
	csvWriter := csv.NewWriter(writer)
	if err := cdre.writeCsv(csvWriter); err != nil {
		t.Error("Unexpected error: ", err)
	}
	expected := `dbafe9c8614c785a65aabd116dd3959c3c56f7f6|*default|*voice|dsafdsaf|*rated|cgrates.org|call|1001|1001|1002|2013-11-07T08:42:25Z|2013-11-07T08:42:26Z|10s|1.01000`
	result := strings.TrimSpace(writer.String())
	if result != expected {
		t.Errorf("Expected: \n%s received: \n%s.", expected, result)
	}
	if cdre.TotalCost() != 1.01 {
		t.Error("Unexpected TotalCost: ", cdre.TotalCost())
	}
}

func TestExportVoiceWithConvert(t *testing.T) {
	writer := &bytes.Buffer{}
	cfg, _ := config.NewDefaultCGRConfig()
	cdreCfg := cfg.CdreProfiles["*default"]
	cdreCfg.ContentFields = []*config.CfgCdrField{
		&config.CfgCdrField{Tag: "ToR", Type: "*composed",
			Value: utils.ParseRSRFieldsMustCompile("ToR", utils.INFIELD_SEP)},
		&config.CfgCdrField{Tag: "OriginID", Type: "*composed",
			Value: utils.ParseRSRFieldsMustCompile("OriginID", utils.INFIELD_SEP)},
		&config.CfgCdrField{Tag: "RequestType", Type: "*composed",
			Value: utils.ParseRSRFieldsMustCompile("RequestType", utils.INFIELD_SEP)},
		&config.CfgCdrField{Tag: "Tenant", Type: "*composed",
			Value: utils.ParseRSRFieldsMustCompile("Tenant", utils.INFIELD_SEP)},
		&config.CfgCdrField{Tag: "Category", Type: "*composed",
			Value: utils.ParseRSRFieldsMustCompile("Category", utils.INFIELD_SEP)},
		&config.CfgCdrField{Tag: "Account", Type: "*composed",
			Value: utils.ParseRSRFieldsMustCompile("Account", utils.INFIELD_SEP)},
		&config.CfgCdrField{Tag: "Destination", Type: "*composed",
			Value: utils.ParseRSRFieldsMustCompile("Destination", utils.INFIELD_SEP)},
		&config.CfgCdrField{Tag: "AnswerTime", Type: "*composed",
			Value:  utils.ParseRSRFieldsMustCompile("AnswerTime", utils.INFIELD_SEP),
			Layout: "2006-01-02T15:04:05Z07:00"},
		&config.CfgCdrField{Tag: "UsageVoice", Type: "*composed",
			FieldFilter: utils.ParseRSRFieldsMustCompile("ToR(*voice)", utils.INFIELD_SEP),
			Value:       utils.ParseRSRFieldsMustCompile("Usage{*duration_seconds}", utils.INFIELD_SEP)},
		&config.CfgCdrField{Tag: "UsageData", Type: "*composed",
			FieldFilter: utils.ParseRSRFieldsMustCompile("ToR(*data)", utils.INFIELD_SEP),
			Value:       utils.ParseRSRFieldsMustCompile("Usage{*duration_nanoseconds}", utils.INFIELD_SEP)},
		&config.CfgCdrField{Tag: "UsageSMS", Type: "*composed",
			FieldFilter: utils.ParseRSRFieldsMustCompile("ToR(*sms)", utils.INFIELD_SEP),
			Value:       utils.ParseRSRFieldsMustCompile("Usage{*duration_nanoseconds}", utils.INFIELD_SEP)},
		&config.CfgCdrField{Tag: "Cost", Type: "*composed",
			Value:            utils.ParseRSRFieldsMustCompile("Cost", utils.INFIELD_SEP),
			RoundingDecimals: 4},
	}
	cdrVoice := &CDR{
		CGRID: utils.Sha1("dsafdsaf", time.Unix(1383813745, 0).UTC().String()),
		ToR:   utils.VOICE, OriginID: "dsafdsaf", OriginHost: "192.168.1.1",
		RequestType: utils.META_RATED, Tenant: "cgrates.org", Category: "call",
		Account: "1001", Subject: "1001", Destination: "1002",
		SetupTime:  time.Unix(1383813745, 0).UTC(),
		AnswerTime: time.Unix(1383813746, 0).UTC(),
		Usage:      time.Duration(10) * time.Second,
		RunID:      utils.DEFAULT_RUNID, Cost: 1.01,
		ExtraFields: map[string]string{"extra1": "val_extra1",
			"extra2": "val_extra2", "extra3": "val_extra3"},
	}
	cdrData := &CDR{
		CGRID: utils.Sha1("abcdef", time.Unix(1383813745, 0).UTC().String()),
		ToR:   utils.DATA, OriginID: "abcdef", OriginHost: "192.168.1.1",
		RequestType: utils.META_RATED, Tenant: "cgrates.org", Category: "call",
		Account: "1001", Subject: "1001", Destination: "1002",
		SetupTime:  time.Unix(1383813745, 0).UTC(),
		AnswerTime: time.Unix(1383813746, 0).UTC(),
		Usage:      time.Duration(10) * time.Nanosecond,
		RunID:      utils.DEFAULT_RUNID, Cost: 0.012,
		ExtraFields: map[string]string{"extra1": "val_extra1",
			"extra2": "val_extra2", "extra3": "val_extra3"},
	}
	cdrSMS := &CDR{
		CGRID: utils.Sha1("sdfwer", time.Unix(1383813745, 0).UTC().String()),
		ToR:   utils.SMS, OriginID: "sdfwer", OriginHost: "192.168.1.1",
		RequestType: utils.META_RATED, Tenant: "cgrates.org", Category: "call",
		Account: "1001", Subject: "1001", Destination: "1002",
		SetupTime:  time.Unix(1383813745, 0).UTC(),
		AnswerTime: time.Unix(1383813746, 0).UTC(),
		Usage:      time.Duration(1),
		RunID:      utils.DEFAULT_RUNID, Cost: 0.15,
		ExtraFields: map[string]string{"extra1": "val_extra1",
			"extra2": "val_extra2", "extra3": "val_extra3"},
	}
	cdre, err := NewCDRExporter([]*CDR{cdrVoice, cdrData, cdrSMS}, cdreCfg,
		utils.MetaFileCSV, "", "", "firstexport",
		true, 1, '|', map[string]float64{}, 0.0,
		5, true, nil)
	if err != nil {
		t.Error("Unexpected error received: ", err)
	}
	if err = cdre.processCDRs(); err != nil {
		t.Error(err)
	}
	csvWriter := csv.NewWriter(writer)
	if err := cdre.writeCsv(csvWriter); err != nil {
		t.Error("Unexpected error: ", err)
	}
	expected := `*sms|sdfwer|*rated|cgrates.org|call|1001|1002|2013-11-07T08:42:26Z|1|0.15000
*voice|dsafdsaf|*rated|cgrates.org|call|1001|1002|2013-11-07T08:42:26Z|10|1.01000
*data|abcdef|*rated|cgrates.org|call|1001|1002|2013-11-07T08:42:26Z|10|0.01200`
	result := strings.TrimSpace(writer.String())
	if len(result) != len(expected) { // export is async, cannot check order
		t.Errorf("expected: \n%s received: \n%s.", expected, result)
	}
	if cdre.TotalCost() != 1.172 {
		t.Error("unexpected TotalCost: ", cdre.TotalCost())
	}
}
