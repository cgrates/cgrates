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
		RunID:      utils.MetaDefault, Cost: 1.01,
		ExtraFields: map[string]string{"extra1": "val_extra1",
			"extra2": "val_extra2", "extra3": "val_extra3"},
	}
	cdre, err := NewCDRExporter([]*CDR{storedCdr1},
		cfg.CdreProfiles[utils.MetaDefault], utils.MetaFileCSV, "", "", "firstexport",
		true, 1, utils.CSV_SEP, cfg.GeneralCfg().HttpSkipTlsVerify, nil, nil)
	if err != nil {
		t.Error("Unexpected error received: ", err)
	}
	if err = cdre.processCDRs(); err != nil {
		t.Error(err)
	}
	if err = cdre.composeHeader(); err != nil {
		t.Error(err)
	}
	if err = cdre.composeTrailer(); err != nil {
		t.Error(err)
	}
	csvWriter := csv.NewWriter(writer)
	if err := cdre.writeCsv(csvWriter); err != nil {
		t.Error("Unexpected error: ", err)
	}
	expected := `dbafe9c8614c785a65aabd116dd3959c3c56f7f6,*default,*voice,dsafdsaf,*rated,cgrates.org,call,1001,1001,1002,2013-11-07T08:42:25Z,2013-11-07T08:42:26Z,10s,1.0100`
	result := strings.TrimSpace(writer.String())
	if result != expected {
		t.Errorf("Expected: \n%s \n received: \n%s.", expected, result)
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
		RunID:      utils.MetaDefault, Cost: 1.01,
		ExtraFields: map[string]string{"extra1": "val_extra1",
			"extra2": "val_extra2", "extra3": "val_extra3"},
	}
	cdre, err := NewCDRExporter([]*CDR{storedCdr1}, cfg.CdreProfiles[utils.MetaDefault],
		utils.MetaFileCSV, "", "", "firstexport", true, 1, '|',
		cfg.GeneralCfg().HttpSkipTlsVerify, nil, nil)
	if err != nil {
		t.Error("Unexpected error received: ", err)
	}
	if err = cdre.processCDRs(); err != nil {
		t.Error(err)
	}
	if err = cdre.composeHeader(); err != nil {
		t.Error(err)
	}
	if err = cdre.composeTrailer(); err != nil {
		t.Error(err)
	}
	csvWriter := csv.NewWriter(writer)
	if err := cdre.writeCsv(csvWriter); err != nil {
		t.Error("Unexpected error: ", err)
	}
	expected := `dbafe9c8614c785a65aabd116dd3959c3c56f7f6|*default|*voice|dsafdsaf|*rated|cgrates.org|call|1001|1001|1002|2013-11-07T08:42:25Z|2013-11-07T08:42:26Z|10s|1.0100`
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
	cdreCfg := cfg.CdreProfiles[utils.MetaDefault]
	cdreCfg.Fields = []*config.FCTemplate{
		{
			Tag:   "*exp.ToR",
			Path:  utils.PathItems{{Field: utils.MetaExp}, {Field: "ToR"}},
			Type:  "*composed",
			Value: config.NewRSRParsersMustCompile(utils.DynamicDataPrefix+utils.MetaReq+utils.NestingSep+"ToR", true, utils.INFIELD_SEP)},
		{
			Tag:   "*exp.OriginID",
			Path:  utils.PathItems{{Field: utils.MetaExp}, {Field: "OriginID"}},
			Type:  "*composed",
			Value: config.NewRSRParsersMustCompile(utils.DynamicDataPrefix+utils.MetaReq+utils.NestingSep+"OriginID", true, utils.INFIELD_SEP)},
		{
			Tag:   "*exp.RequestType",
			Path:  utils.PathItems{{Field: utils.MetaExp}, {Field: "RequestType"}},
			Type:  "*composed",
			Value: config.NewRSRParsersMustCompile(utils.DynamicDataPrefix+utils.MetaReq+utils.NestingSep+"RequestType", true, utils.INFIELD_SEP)},
		{
			Tag:   "*exp.Tenant",
			Path:  utils.PathItems{{Field: utils.MetaExp}, {Field: "Tenant"}},
			Type:  "*composed",
			Value: config.NewRSRParsersMustCompile(utils.DynamicDataPrefix+utils.MetaReq+utils.NestingSep+"Tenant", true, utils.INFIELD_SEP)},
		{
			Tag:   "*exp.Category",
			Path:  utils.PathItems{{Field: utils.MetaExp}, {Field: "Category"}},
			Type:  "*composed",
			Value: config.NewRSRParsersMustCompile(utils.DynamicDataPrefix+utils.MetaReq+utils.NestingSep+"Category", true, utils.INFIELD_SEP)},
		{
			Tag:   "*exp.Account",
			Path:  utils.PathItems{{Field: utils.MetaExp}, {Field: "Account"}},
			Type:  "*composed",
			Value: config.NewRSRParsersMustCompile(utils.DynamicDataPrefix+utils.MetaReq+utils.NestingSep+"Account", true, utils.INFIELD_SEP)},
		{
			Tag:   "*exp.Destination",
			Path:  utils.PathItems{{Field: utils.MetaExp}, {Field: "Destination"}},
			Type:  "*composed",
			Value: config.NewRSRParsersMustCompile(utils.DynamicDataPrefix+utils.MetaReq+utils.NestingSep+"Destination", true, utils.INFIELD_SEP)},
		{
			Tag:    "*exp.AnswerTime",
			Path:   utils.PathItems{{Field: utils.MetaExp}, {Field: "AnswerTime"}},
			Type:   "*composed",
			Value:  config.NewRSRParsersMustCompile(utils.DynamicDataPrefix+utils.MetaReq+utils.NestingSep+"AnswerTime", true, utils.INFIELD_SEP),
			Layout: "2006-01-02T15:04:05Z07:00"},
		{
			Tag:     "*exp.UsageVoice",
			Path:    utils.PathItems{{Field: utils.MetaExp}, {Field: "UsageVoice"}},
			Type:    "*composed",
			Filters: []string{"*string:~*req.ToR:*voice"},
			Value:   config.NewRSRParsersMustCompile("~*req.Usage{*duration_seconds}", true, utils.INFIELD_SEP)},
		{
			Tag:     "*exp.UsageData",
			Path:    utils.PathItems{{Field: utils.MetaExp}, {Field: "UsageData"}},
			Type:    "*composed",
			Filters: []string{"*string:~*req.ToR:*data"},
			Value:   config.NewRSRParsersMustCompile("~*req.Usage{*duration_nanoseconds}", true, utils.INFIELD_SEP)},
		{
			Tag:     "*exp.UsageSMS",
			Path:    utils.PathItems{{Field: utils.MetaExp}, {Field: "UsageSMS"}},
			Type:    "*composed",
			Filters: []string{"*string:~*req.ToR:*sms"},
			Value:   config.NewRSRParsersMustCompile("~*req.Usage{*duration_nanoseconds}", true, utils.INFIELD_SEP)},
		{
			Tag:              "*exp.Cost",
			Path:             utils.PathItems{{Field: utils.MetaExp}, {Field: "Cost"}},
			Type:             "*composed",
			Value:            config.NewRSRParsersMustCompile(utils.DynamicDataPrefix+utils.MetaReq+utils.NestingSep+"Cost", true, utils.INFIELD_SEP),
			RoundingDecimals: 5},
	}
	cdrVoice := &CDR{
		CGRID: utils.Sha1("dsafdsaf", time.Unix(1383813745, 0).UTC().String()),
		ToR:   utils.VOICE, OriginID: "dsafdsaf", OriginHost: "192.168.1.1",
		RequestType: utils.META_RATED, Tenant: "cgrates.org", Category: "call",
		Account: "1001", Subject: "1001", Destination: "1002",
		SetupTime:  time.Unix(1383813745, 0).UTC(),
		AnswerTime: time.Unix(1383813746, 0).UTC(),
		Usage:      time.Duration(10) * time.Second,
		RunID:      utils.MetaDefault, Cost: 1.01,
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
		RunID:      utils.MetaDefault, Cost: 0.012,
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
		RunID:      utils.MetaDefault, Cost: 0.15,
		ExtraFields: map[string]string{"extra1": "val_extra1",
			"extra2": "val_extra2", "extra3": "val_extra3"},
	}
	cdre, err := NewCDRExporter([]*CDR{cdrVoice, cdrData, cdrSMS}, cdreCfg,
		utils.MetaFileCSV, "", "", "firstexport",
		true, 1, '|', true, nil, &FilterS{cfg: cfg})
	if err != nil {
		t.Error("Unexpected error received: ", err)
	}
	if err = cdre.processCDRs(); err != nil {
		t.Error(err)
	}
	if err = cdre.composeHeader(); err != nil {
		t.Error(err)
	}
	if err = cdre.composeTrailer(); err != nil {
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

func TestExportWithFilter(t *testing.T) {
	writer := &bytes.Buffer{}
	cfg, _ := config.NewDefaultCGRConfig()
	cdreCfg := cfg.CdreProfiles[utils.MetaDefault]
	cdreCfg.Filters = []string{"*string:~*req.Tenant:cgrates.org"}
	cdreCfg.Fields = []*config.FCTemplate{
		{
			Tag:   "*exp.ToR",
			Path:  utils.PathItems{{Field: utils.MetaExp}, {Field: "ToR"}},
			Type:  "*composed",
			Value: config.NewRSRParsersMustCompile(utils.DynamicDataPrefix+utils.MetaReq+utils.NestingSep+"ToR", true, utils.INFIELD_SEP)},
		{
			Tag:   "*exp.OriginID",
			Path:  utils.PathItems{{Field: utils.MetaExp}, {Field: "OriginID"}},
			Type:  "*composed",
			Value: config.NewRSRParsersMustCompile(utils.DynamicDataPrefix+utils.MetaReq+utils.NestingSep+"OriginID", true, utils.INFIELD_SEP)},
		{
			Tag:   "*exp.RequestType",
			Path:  utils.PathItems{{Field: utils.MetaExp}, {Field: "RequestType"}},
			Type:  "*composed",
			Value: config.NewRSRParsersMustCompile(utils.DynamicDataPrefix+utils.MetaReq+utils.NestingSep+"RequestType", true, utils.INFIELD_SEP)},
		{
			Tag:   "*exp.Tenant",
			Path:  utils.PathItems{{Field: utils.MetaExp}, {Field: "Tenant"}},
			Type:  "*composed",
			Value: config.NewRSRParsersMustCompile(utils.DynamicDataPrefix+utils.MetaReq+utils.NestingSep+"Tenant", true, utils.INFIELD_SEP)},
		{
			Tag:   "*exp.Category",
			Path:  utils.PathItems{{Field: utils.MetaExp}, {Field: "Category"}},
			Type:  "*composed",
			Value: config.NewRSRParsersMustCompile(utils.DynamicDataPrefix+utils.MetaReq+utils.NestingSep+"Category", true, utils.INFIELD_SEP)},
		{
			Tag:   "*exp.Account",
			Path:  utils.PathItems{{Field: utils.MetaExp}, {Field: "Account"}},
			Type:  "*composed",
			Value: config.NewRSRParsersMustCompile(utils.DynamicDataPrefix+utils.MetaReq+utils.NestingSep+"Account", true, utils.INFIELD_SEP)},
		{
			Tag:   "*exp.Destination",
			Path:  utils.PathItems{{Field: utils.MetaExp}, {Field: "Destination"}},
			Type:  "*composed",
			Value: config.NewRSRParsersMustCompile(utils.DynamicDataPrefix+utils.MetaReq+utils.NestingSep+"Destination", true, utils.INFIELD_SEP)},
		{
			Tag:    "*exp.AnswerTime",
			Path:   utils.PathItems{{Field: utils.MetaExp}, {Field: "AnswerTime"}},
			Type:   "*composed",
			Value:  config.NewRSRParsersMustCompile(utils.DynamicDataPrefix+utils.MetaReq+utils.NestingSep+"AnswerTime", true, utils.INFIELD_SEP),
			Layout: "2006-01-02T15:04:05Z07:00"},
		{
			Tag:     "*exp.UsageVoice",
			Path:    utils.PathItems{{Field: utils.MetaExp}, {Field: "UsageVoice"}},
			Type:    "*composed",
			Filters: []string{"*string:~*req.ToR:*voice"},
			Value:   config.NewRSRParsersMustCompile("~*req.Usage{*duration_seconds}", true, utils.INFIELD_SEP)},
		{
			Tag:     "*exp.UsageData",
			Path:    utils.PathItems{{Field: utils.MetaExp}, {Field: "UsageData"}},
			Type:    "*composed",
			Filters: []string{"*string:~*req.ToR:*data"},
			Value:   config.NewRSRParsersMustCompile("~*req.Usage{*duration_nanoseconds}", true, utils.INFIELD_SEP)},
		{
			Tag:     "*exp.UsageSMS",
			Path:    utils.PathItems{{Field: utils.MetaExp}, {Field: "UsageSMS"}},
			Type:    "*composed",
			Filters: []string{"*string:~*req.ToR:*sms"},
			Value:   config.NewRSRParsersMustCompile("~*req.Usage{*duration_nanoseconds}", true, utils.INFIELD_SEP)},
		{
			Tag:              "*exp.Cost",
			Path:             utils.PathItems{{Field: utils.MetaExp}, {Field: "Cost"}},
			Type:             "*composed",
			Value:            config.NewRSRParsersMustCompile(utils.DynamicDataPrefix+utils.MetaReq+utils.NestingSep+"Cost", true, utils.INFIELD_SEP),
			RoundingDecimals: 5},
	}
	cdrVoice := &CDR{
		CGRID: utils.Sha1("dsafdsaf", time.Unix(1383813745, 0).UTC().String()),
		ToR:   utils.VOICE, OriginID: "dsafdsaf", OriginHost: "192.168.1.1",
		RequestType: utils.META_RATED, Tenant: "cgrates.org", Category: "call",
		Account: "1001", Subject: "1001", Destination: "1002",
		SetupTime:  time.Unix(1383813745, 0).UTC(),
		AnswerTime: time.Unix(1383813746, 0).UTC(),
		Usage:      time.Duration(10) * time.Second,
		RunID:      utils.MetaDefault, Cost: 1.01,
		ExtraFields: map[string]string{"extra1": "val_extra1",
			"extra2": "val_extra2", "extra3": "val_extra3"},
	}
	cdrData := &CDR{
		CGRID: utils.Sha1("abcdef", time.Unix(1383813745, 0).UTC().String()),
		ToR:   utils.DATA, OriginID: "abcdef", OriginHost: "192.168.1.1",
		RequestType: utils.META_RATED, Tenant: "AnotherTenant", Category: "call", //for data CDR use different Tenant
		Account: "1001", Subject: "1001", Destination: "1002",
		SetupTime:  time.Unix(1383813745, 0).UTC(),
		AnswerTime: time.Unix(1383813746, 0).UTC(),
		Usage:      time.Duration(10) * time.Nanosecond,
		RunID:      utils.MetaDefault, Cost: 0.012,
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
		RunID:      utils.MetaDefault, Cost: 0.15,
		ExtraFields: map[string]string{"extra1": "val_extra1",
			"extra2": "val_extra2", "extra3": "val_extra3"},
	}
	cdre, err := NewCDRExporter([]*CDR{cdrVoice, cdrData, cdrSMS}, cdreCfg,
		utils.MetaFileCSV, "", "", "firstexport",
		true, 1, '|', true, nil, &FilterS{cfg: cfg})
	if err != nil {
		t.Error("Unexpected error received: ", err)
	}
	if err = cdre.processCDRs(); err != nil {
		t.Error(err)
	}
	if err = cdre.composeHeader(); err != nil {
		t.Error(err)
	}
	if err = cdre.composeTrailer(); err != nil {
		t.Error(err)
	}
	csvWriter := csv.NewWriter(writer)
	if err := cdre.writeCsv(csvWriter); err != nil {
		t.Error("Unexpected error: ", err)
	}
	expected := `*sms|sdfwer|*rated|cgrates.org|call|1001|1002|2013-11-07T08:42:26Z|1|0.15000
*voice|dsafdsaf|*rated|cgrates.org|call|1001|1002|2013-11-07T08:42:26Z|10|1.01000`
	result := strings.TrimSpace(writer.String())
	if len(result) != len(expected) { // export is async, cannot check order
		t.Errorf("expected: \n%s received: \n%s.", expected, result)
	}
	if cdre.TotalCost() != 1.16 {
		t.Error("unexpected TotalCost: ", cdre.TotalCost())
	}
}

func TestExportWithFilter2(t *testing.T) {
	writer := &bytes.Buffer{}
	cfg, _ := config.NewDefaultCGRConfig()
	cdreCfg := cfg.CdreProfiles[utils.MetaDefault]
	cdreCfg.Filters = []string{"*string:~*req.Tenant:cgrates.org", "*lte:~*req.Cost:0.5"}
	cdreCfg.Fields = []*config.FCTemplate{
		{
			Tag:   "*exp.ToR",
			Path:  utils.PathItems{{Field: utils.MetaExp}, {Field: "ToR"}},
			Type:  "*composed",
			Value: config.NewRSRParsersMustCompile(utils.DynamicDataPrefix+utils.MetaReq+utils.NestingSep+"ToR", true, utils.INFIELD_SEP)},
		{
			Tag:   "*exp.OriginID",
			Path:  utils.PathItems{{Field: utils.MetaExp}, {Field: "OriginID"}},
			Type:  "*composed",
			Value: config.NewRSRParsersMustCompile(utils.DynamicDataPrefix+utils.MetaReq+utils.NestingSep+"OriginID", true, utils.INFIELD_SEP)},
		{
			Tag:   "*exp.RequestType",
			Path:  utils.PathItems{{Field: utils.MetaExp}, {Field: "RequestType"}},
			Type:  "*composed",
			Value: config.NewRSRParsersMustCompile(utils.DynamicDataPrefix+utils.MetaReq+utils.NestingSep+"RequestType", true, utils.INFIELD_SEP)},
		{
			Tag:   "*exp.Tenant",
			Path:  utils.PathItems{{Field: utils.MetaExp}, {Field: "Tenant"}},
			Type:  "*composed",
			Value: config.NewRSRParsersMustCompile(utils.DynamicDataPrefix+utils.MetaReq+utils.NestingSep+"Tenant", true, utils.INFIELD_SEP)},
		{
			Tag:   "*exp.Category",
			Path:  utils.PathItems{{Field: utils.MetaExp}, {Field: "Category"}},
			Type:  "*composed",
			Value: config.NewRSRParsersMustCompile(utils.DynamicDataPrefix+utils.MetaReq+utils.NestingSep+"Category", true, utils.INFIELD_SEP)},
		{
			Tag:   "*exp.Account",
			Path:  utils.PathItems{{Field: utils.MetaExp}, {Field: "Account"}},
			Type:  "*composed",
			Value: config.NewRSRParsersMustCompile(utils.DynamicDataPrefix+utils.MetaReq+utils.NestingSep+"Account", true, utils.INFIELD_SEP)},
		{
			Tag:   "*exp.Destination",
			Path:  utils.PathItems{{Field: utils.MetaExp}, {Field: "Destination"}},
			Type:  "*composed",
			Value: config.NewRSRParsersMustCompile(utils.DynamicDataPrefix+utils.MetaReq+utils.NestingSep+"Destination", true, utils.INFIELD_SEP)},
		{
			Tag:    "*exp.AnswerTime",
			Path:   utils.PathItems{{Field: utils.MetaExp}, {Field: "AnswerTime"}},
			Type:   "*composed",
			Value:  config.NewRSRParsersMustCompile(utils.DynamicDataPrefix+utils.MetaReq+utils.NestingSep+"AnswerTime", true, utils.INFIELD_SEP),
			Layout: "2006-01-02T15:04:05Z07:00"},
		{
			Tag:     "*exp.UsageVoice",
			Path:    utils.PathItems{{Field: utils.MetaExp}, {Field: "UsageVoice"}},
			Type:    "*composed",
			Filters: []string{"*string:~*req.ToR:*voice"},
			Value:   config.NewRSRParsersMustCompile("~*req.Usage{*duration_seconds}", true, utils.INFIELD_SEP)},
		{
			Tag:     "*exp.UsageData",
			Path:    utils.PathItems{{Field: utils.MetaExp}, {Field: "UsageData"}},
			Type:    "*composed",
			Filters: []string{"*string:~*req.ToR:*data"},
			Value:   config.NewRSRParsersMustCompile("~*req.Usage{*duration_nanoseconds}", true, utils.INFIELD_SEP)},
		{
			Tag:     "*exp.UsageSMS",
			Path:    utils.PathItems{{Field: utils.MetaExp}, {Field: "UsageSMS"}},
			Type:    "*composed",
			Filters: []string{"*string:~*req.ToR:*sms"},
			Value:   config.NewRSRParsersMustCompile("~*req.Usage{*duration_nanoseconds}", true, utils.INFIELD_SEP)},
		{
			Tag:              "*exp.Cost",
			Path:             utils.PathItems{{Field: utils.MetaExp}, {Field: "Cost"}},
			Type:             "*composed",
			Value:            config.NewRSRParsersMustCompile(utils.DynamicDataPrefix+utils.MetaReq+utils.NestingSep+"Cost", true, utils.INFIELD_SEP),
			RoundingDecimals: 5},
	}
	cdrVoice := &CDR{
		CGRID: utils.Sha1("dsafdsaf", time.Unix(1383813745, 0).UTC().String()),
		ToR:   utils.VOICE, OriginID: "dsafdsaf", OriginHost: "192.168.1.1",
		RequestType: utils.META_RATED, Tenant: "cgrates.org", Category: "call",
		Account: "1001", Subject: "1001", Destination: "1002",
		SetupTime:  time.Unix(1383813745, 0).UTC(),
		AnswerTime: time.Unix(1383813746, 0).UTC(),
		Usage:      time.Duration(10) * time.Second,
		RunID:      utils.MetaDefault, Cost: 1.01,
		ExtraFields: map[string]string{"extra1": "val_extra1",
			"extra2": "val_extra2", "extra3": "val_extra3"},
	}
	cdrData := &CDR{
		CGRID: utils.Sha1("abcdef", time.Unix(1383813745, 0).UTC().String()),
		ToR:   utils.DATA, OriginID: "abcdef", OriginHost: "192.168.1.1",
		RequestType: utils.META_RATED, Tenant: "AnotherTenant", Category: "call", //for data CDR use different Tenant
		Account: "1001", Subject: "1001", Destination: "1002",
		SetupTime:  time.Unix(1383813745, 0).UTC(),
		AnswerTime: time.Unix(1383813746, 0).UTC(),
		Usage:      time.Duration(10) * time.Nanosecond,
		RunID:      utils.MetaDefault, Cost: 0.012,
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
		RunID:      utils.MetaDefault, Cost: 0.15,
		ExtraFields: map[string]string{"extra1": "val_extra1",
			"extra2": "val_extra2", "extra3": "val_extra3"},
	}
	cdre, err := NewCDRExporter([]*CDR{cdrVoice, cdrData, cdrSMS}, cdreCfg,
		utils.MetaFileCSV, "", "", "firstexport",
		true, 1, '|', true, nil, &FilterS{cfg: cfg})
	if err != nil {
		t.Error("Unexpected error received: ", err)
	}
	if err = cdre.processCDRs(); err != nil {
		t.Error(err)
	}
	if err = cdre.composeHeader(); err != nil {
		t.Error(err)
	}
	if err = cdre.composeTrailer(); err != nil {
		t.Error(err)
	}
	csvWriter := csv.NewWriter(writer)
	if err := cdre.writeCsv(csvWriter); err != nil {
		t.Error("Unexpected error: ", err)
	}
	expected := `*sms|sdfwer|*rated|cgrates.org|call|1001|1002|2013-11-07T08:42:26Z|1|0.15000`
	result := strings.TrimSpace(writer.String())
	if len(result) != len(expected) { // export is async, cannot check order
		t.Errorf("expected: \n%s received: \n%s.", expected, result)
	}
	if cdre.TotalCost() != 0.15 {
		t.Error("unexpected TotalCost: ", cdre.TotalCost())
	}
}
