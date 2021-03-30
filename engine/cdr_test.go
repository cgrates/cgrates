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
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

func TestNewCDRFromExternalCDR(t *testing.T) {
	extCdr := &ExternalCDR{
		CGRID:   utils.Sha1("dsafdsaf", time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC).String()),
		OrderID: 123, ToR: utils.MetaVoice, OriginID: "dsafdsaf", OriginHost: "192.168.1.1",
		Source: utils.UnitTest, RequestType: utils.MetaRated,
		Tenant: "cgrates.org", Category: "call", Account: "1001", Subject: "1001", Destination: "1002",
		SetupTime: "2013-11-07T08:42:20Z", AnswerTime: "2013-11-07T08:42:26Z", RunID: utils.MetaDefault,
		Usage: "10", Cost: 1.01, PreRated: true,
		ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
	}
	eStorCdr := &CDR{CGRID: utils.Sha1("dsafdsaf", time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC).String()),
		OrderID: 123, ToR: utils.MetaVoice, OriginID: "dsafdsaf", OriginHost: "192.168.1.1",
		Source: utils.UnitTest, RequestType: utils.MetaRated, RunID: utils.MetaDefault,
		Tenant: "cgrates.org", Category: "call", Account: "1001",
		Subject: "1001", Destination: "1002",
		SetupTime:  time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC),
		AnswerTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC),
		Usage:      10, Cost: 1.01, PreRated: true,
		ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
	}
	if CDR, err := NewCDRFromExternalCDR(extCdr, ""); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eStorCdr, CDR) {
		t.Errorf("Expected: %+v, received: %+v", eStorCdr, CDR)
	}
}

func TestCDRClone(t *testing.T) {
	storCdr := &CDR{CGRID: utils.Sha1("dsafdsaf", time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC).String()),
		OrderID: 123, ToR: utils.MetaVoice, OriginID: "dsafdsaf", OriginHost: "192.168.1.1",
		Source: utils.UnitTest, RequestType: utils.MetaRated, Tenant: "cgrates.org",
		Category: "call", Account: "1001", Subject: "1001", Destination: "1002",
		SetupTime:  time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC),
		AnswerTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC),
		RunID:      utils.MetaDefault, Usage: 10,
		ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
		Cost:        1.01, PreRated: true,
	}
	if clnStorCdr := storCdr.Clone(); !reflect.DeepEqual(storCdr, clnStorCdr) {
		t.Errorf("Expecting: %+v, received: %+v", storCdr, clnStorCdr)
	}
}

func TestFieldAsString(t *testing.T) {
	cdr := CDR{CGRID: utils.Sha1("dsafdsaf", time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC).String()),
		OrderID: 123, ToR: utils.MetaVoice, OriginID: "dsafdsaf",
		OriginHost: "192.168.1.1", Source: "test", RequestType: utils.MetaRated,
		Tenant: "cgrates.org", Category: "call", Account: "1001", Subject: "1001",
		Destination: "1002", SetupTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC),
		AnswerTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC), RunID: utils.MetaDefault,
		Usage: 10 * time.Second, Cost: 1.01,
		ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
	}
	prsr := config.NewRSRParserMustCompile(utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.CGRID)
	eFldVal := cdr.CGRID
	if fldVal, err := cdr.FieldAsString(prsr); err != nil {
		t.Error(err)
	} else if fldVal != eFldVal {
		t.Errorf("field: <%v>, expected: <%v>, received: <%v>", prsr, eFldVal, fldVal)
	}
	prsr = config.NewRSRParserMustCompile(utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.OrderID)
	eFldVal = strconv.FormatInt(cdr.OrderID, 10)
	if fldVal, err := cdr.FieldAsString(prsr); err != nil {
		t.Error(err)
	} else if fldVal != eFldVal {
		t.Errorf("field: <%v>, expected: <%v>, received: <%v>", prsr, eFldVal, fldVal)
	}
	prsr = config.NewRSRParserMustCompile(utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.ToR)
	eFldVal = cdr.ToR
	if fldVal, err := cdr.FieldAsString(prsr); err != nil {
		t.Error(err)
	} else if fldVal != eFldVal {
		t.Errorf("field: <%v>, expected: <%v>, received: <%v>", prsr, eFldVal, fldVal)
	}
	prsr = config.NewRSRParserMustCompile(utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.OriginID)
	eFldVal = cdr.OriginID
	if fldVal, err := cdr.FieldAsString(prsr); err != nil {
		t.Error(err)
	} else if fldVal != eFldVal {
		t.Errorf("field: <%v>, expected: <%v>, received: <%v>", prsr, eFldVal, fldVal)
	}
	prsr = config.NewRSRParserMustCompile(utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.OriginHost)
	eFldVal = cdr.OriginHost
	if fldVal, err := cdr.FieldAsString(prsr); err != nil {
		t.Error(err)
	} else if fldVal != eFldVal {
		t.Errorf("field: <%v>, expected: <%v>, received: <%v>", prsr, eFldVal, fldVal)
	}
	prsr = config.NewRSRParserMustCompile(utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Source)
	eFldVal = cdr.Source
	if fldVal, err := cdr.FieldAsString(prsr); err != nil {
		t.Error(err)
	} else if fldVal != eFldVal {
		t.Errorf("field: <%v>, expected: <%v>, received: <%v>", prsr, eFldVal, fldVal)
	}
	prsr = config.NewRSRParserMustCompile(utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.RequestType)
	eFldVal = cdr.RequestType
	if fldVal, err := cdr.FieldAsString(prsr); err != nil {
		t.Error(err)
	} else if fldVal != eFldVal {
		t.Errorf("field: <%v>, expected: <%v>, received: <%v>", prsr, eFldVal, fldVal)
	}
	prsr = config.NewRSRParserMustCompile(utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Category)
	eFldVal = cdr.Category
	if fldVal, err := cdr.FieldAsString(prsr); err != nil {
		t.Error(err)
	} else if fldVal != eFldVal {
		t.Errorf("field: <%v>, expected: <%v>, received: <%v>", prsr, eFldVal, fldVal)
	}
	prsr = config.NewRSRParserMustCompile(utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.AccountField)
	eFldVal = cdr.Account
	if fldVal, err := cdr.FieldAsString(prsr); err != nil {
		t.Error(err)
	} else if fldVal != eFldVal {
		t.Errorf("field: <%v>, expected: <%v>, received: <%v>", prsr, eFldVal, fldVal)
	}
	prsr = config.NewRSRParserMustCompile(utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Subject)
	eFldVal = cdr.Subject
	if fldVal, err := cdr.FieldAsString(prsr); err != nil {
		t.Error(err)
	} else if fldVal != eFldVal {
		t.Errorf("field: <%v>, expected: <%v>, received: <%v>", prsr, eFldVal, fldVal)
	}
	prsr = config.NewRSRParserMustCompile(utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Destination)
	eFldVal = cdr.Destination
	if fldVal, err := cdr.FieldAsString(prsr); err != nil {
		t.Error(err)
	} else if fldVal != eFldVal {
		t.Errorf("field: <%v>, expected: <%v>, received: <%v>", prsr, eFldVal, fldVal)
	}
	prsr = config.NewRSRParserMustCompile(utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.SetupTime)
	eFldVal = cdr.SetupTime.Format(time.RFC3339)
	if fldVal, err := cdr.FieldAsString(prsr); err != nil {
		t.Error(err)
	} else if fldVal != eFldVal {
		t.Errorf("expected: <%s>, received: <%s>", eFldVal, fldVal)
	}
	prsr = config.NewRSRParserMustCompile(utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.AnswerTime)
	eFldVal = cdr.AnswerTime.Format(time.RFC3339)
	if fldVal, err := cdr.FieldAsString(prsr); err != nil {
		t.Error(err)
	} else if fldVal != eFldVal {
		t.Errorf("expected: <%s>, received: <%s>", eFldVal, fldVal)
	}
	prsr = config.NewRSRParserMustCompile(utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Usage)
	eFldVal = "10s"
	if fldVal, err := cdr.FieldAsString(prsr); err != nil {
		t.Error(err)
	} else if fldVal != eFldVal {
		t.Errorf("field: <%v>, expected: <%v>, received: <%v>", prsr, eFldVal, fldVal)
	}
	prsr = config.NewRSRParserMustCompile(utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.RunID)
	eFldVal = cdr.RunID
	if fldVal, err := cdr.FieldAsString(prsr); err != nil {
		t.Error(err)
	} else if fldVal != eFldVal {
		t.Errorf("field: <%v>, expected: <%v>, received: <%v>", prsr, eFldVal, fldVal)
	}
	prsr = config.NewRSRParserMustCompile(utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Cost)
	eFldVal = strconv.FormatFloat(cdr.Cost, 'f', -1, 64)
	if fldVal, err := cdr.FieldAsString(prsr); err != nil {
		t.Error(err)
	} else if fldVal != eFldVal {
		t.Errorf("field: <%v>, expected: <%v>, received: <%v>", prsr, eFldVal, fldVal)
	}
	prsr = config.NewRSRParserMustCompile(utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + "field_extr1")
	eFldVal = cdr.ExtraFields["field_extr1"]
	if fldVal, err := cdr.FieldAsString(prsr); err != nil {
		t.Error(err)
	} else if fldVal != eFldVal {
		t.Errorf("field: <%v>, expected: <%v>, received: <%v>", prsr, eFldVal, fldVal)
	}
	prsr = config.NewRSRParserMustCompile(utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + "fieldextr2")
	eFldVal = cdr.ExtraFields["fieldextr2"]
	if fldVal, err := cdr.FieldAsString(prsr); err != nil {
		t.Error(err)
	} else if fldVal != eFldVal {
		t.Errorf("field: <%v>, expected: <%v>, received: <%v>", prsr, eFldVal, fldVal)
	}
	prsr = config.NewRSRParserMustCompile(utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + "dummy_field")
	if fldVal, err := cdr.FieldAsString(prsr); err != utils.ErrNotFound {
		t.Error(err)
	} else if fldVal != utils.EmptyString {
		t.Errorf("field: <%v>, expected: <%v>, received: <%v>", prsr, utils.EmptyString, fldVal)
	}
}

func TestFieldsAsString(t *testing.T) {
	cdr := CDR{CGRID: utils.Sha1("dsafdsaf", time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC).String()),
		OrderID: 123, ToR: utils.MetaVoice, OriginID: "dsafdsaf", OriginHost: "192.168.1.1", Source: "test",
		RequestType: utils.MetaRated, Tenant: "cgrates.org",
		Category: "call", Account: "1001", Subject: "1001", Destination: "1002",
		SetupTime:  time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC),
		AnswerTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC),
		RunID:      utils.MetaDefault, Usage: 10 * time.Second,
		ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"}, Cost: 1.01,
	}
	eVal := "call_from_1001"
	if val := cdr.FieldsAsString(
		config.NewRSRParsersMustCompile("~*req.Category;_from_;~*req.Account", utils.InfieldSep)); val != eVal {
		t.Errorf("Expecting : %s, received: %q", eVal, val)
	}
}

func TestFormatCost(t *testing.T) {
	cdr := CDR{Cost: 1.01}
	if cdr.FormatCost(0, 4) != "1.0100" {
		t.Error("Unexpected format of the cost: ", cdr.FormatCost(0, 4))
	}
	cdr = CDR{Cost: 1.01001}
	if cdr.FormatCost(0, 4) != "1.0100" {
		t.Error("Unexpected format of the cost: ", cdr.FormatCost(0, 4))
	}
	if cdr.FormatCost(2, 0) != "101" {
		t.Error("Unexpected format of the cost: ", cdr.FormatCost(2, 0))
	}
	if cdr.FormatCost(1, 0) != "10" {
		t.Error("Unexpected format of the cost: ", cdr.FormatCost(1, 0))
	}
	if cdr.FormatCost(2, 3) != "101.001" {
		t.Error("Unexpected format of the cost: ", cdr.FormatCost(2, 3))
	}
}

func TestCDRAsExternalCDR(t *testing.T) {
	storCdr := CDR{CGRID: utils.Sha1("dsafdsaf", time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC).String()),
		OrderID: 123, ToR: utils.MetaVoice, OriginID: "dsafdsaf", OriginHost: "192.168.1.1",
		Source: utils.UnitTest, RequestType: utils.MetaRated,
		Tenant: "cgrates.org", Category: "call", Account: "1001", Subject: "1001", Destination: "1002",
		SetupTime:  time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC),
		AnswerTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC),
		RunID:      utils.MetaDefault, Usage: 10 * time.Second, Cost: 1.01,
		ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"}}
	expectOutCdr := &ExternalCDR{
		CGRID:   utils.Sha1("dsafdsaf", time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC).String()),
		OrderID: 123, ToR: utils.MetaVoice, OriginID: "dsafdsaf", OriginHost: "192.168.1.1",
		Source: utils.UnitTest, RequestType: utils.MetaRated,
		Tenant: "cgrates.org", Category: "call", Account: "1001", Subject: "1001", Destination: "1002",
		SetupTime: "2013-11-07T08:42:20Z", AnswerTime: "2013-11-07T08:42:26Z", RunID: utils.MetaDefault,
		Usage: "10s", Cost: 1.01, CostDetails: "null",
		ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"}}
	if cdrOut := storCdr.AsExternalCDR(); !reflect.DeepEqual(expectOutCdr, cdrOut) {
		t.Errorf("Expected: %+v, received: %+v", expectOutCdr, cdrOut)
	}
}

func TestCDRAsMapStringIface(t *testing.T) {
	cdr := &CDR{
		CGRID:       utils.Sha1("dsafdsaf", time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC).String()),
		OrderID:     123,
		ToR:         utils.MetaVoice,
		OriginID:    "dsafdsaf",
		OriginHost:  "192.168.1.1",
		Source:      utils.UnitTest,
		RequestType: utils.MetaRated,
		Tenant:      "cgrates.org",
		Category:    "call",
		Account:     "1002",
		Subject:     "1001",
		Destination: "+4986517174963",
		SetupTime:   time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC),
		AnswerTime:  time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC),
		RunID:       utils.MetaDefault,
		Usage:       10 * time.Second,
		ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
		Cost:        1.01,
	}

	mp := map[string]interface{}{
		"field_extr1":      "val_extr1",
		"fieldextr2":       "valextr2",
		utils.CGRID:        cdr.CGRID,
		utils.RunID:        utils.MetaDefault,
		utils.OrderID:      cdr.OrderID,
		utils.OriginHost:   "192.168.1.1",
		utils.Source:       utils.UnitTest,
		utils.OriginID:     "dsafdsaf",
		utils.ToR:          utils.MetaVoice,
		utils.RequestType:  utils.MetaRated,
		utils.Tenant:       "cgrates.org",
		utils.Category:     "call",
		utils.AccountField: "1002",
		utils.Subject:      "1001",
		utils.Destination:  "+4986517174963",
		utils.SetupTime:    time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC),
		utils.AnswerTime:   time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC),
		utils.Usage:        10 * time.Second,
		utils.CostSource:   cdr.CostSource,
		utils.Cost:         1.01,
		utils.PreRated:     false,
		utils.Partial:      false,
		utils.ExtraInfo:    cdr.ExtraInfo,
	}
	if cdrMp := cdr.AsMapStringIface(); !reflect.DeepEqual(mp, cdrMp) {
		t.Errorf("Expecting: %+v, received: %+v", mp, cdrMp)
	}
}

func TestCDRNewCDRFromSQL(t *testing.T) {
	extraFields := map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"}
	cdrSQL := &CDRsql{
		ID:          123,
		Cgrid:       "abecd993d06672714c4218a6dcf8278e0589a171",
		RunID:       utils.MetaDefault,
		OriginID:    "dsafdsaf",
		TOR:         utils.MetaVoice,
		Source:      utils.UnitTest,
		Tenant:      "cgrates.org",
		Category:    "call",
		Account:     "1001",
		Subject:     "1001",
		Destination: "+4986517174963",
		SetupTime:   time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC),
		AnswerTime:  time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC),
		Usage:       10000000000,
		Cost:        1.01,
		RequestType: utils.MetaRated,
		OriginHost:  "192.168.1.1",
		ExtraFields: utils.ToJSON(extraFields),
	}

	cdr := &CDR{CGRID: utils.Sha1("dsafdsaf", time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC).String()),
		OrderID:     123,
		ToR:         utils.MetaVoice,
		OriginID:    "dsafdsaf",
		OriginHost:  "192.168.1.1",
		Source:      utils.UnitTest,
		RequestType: utils.MetaRated,
		Tenant:      "cgrates.org",
		Category:    "call",
		Account:     "1001",
		Subject:     "1001",
		Destination: "+4986517174963",
		SetupTime:   time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC),
		AnswerTime:  time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC),
		RunID:       utils.MetaDefault,
		Usage:       10 * time.Second,
		Cost:        1.01,
		ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
	}

	if eCDR, err := NewCDRFromSQL(cdrSQL); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(cdr, eCDR) {
		t.Errorf("Expecting: %+v, received: %+v", cdr, eCDR)
	}

}

func TestCDRAsCGREvent(t *testing.T) {
	cdr := &CDR{
		CGRID:       utils.Sha1("dsafdsaf", time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC).String()),
		OrderID:     123,
		ToR:         utils.MetaVoice,
		OriginID:    "dsafdsaf",
		OriginHost:  "192.168.1.1",
		Source:      utils.UnitTest,
		RequestType: utils.MetaRated,
		Tenant:      "cgrates.org",
		Category:    "call",
		Account:     "1001",
		Subject:     "1001",
		Destination: "+4986517174963",
		SetupTime:   time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC),
		AnswerTime:  time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC),
		RunID:       utils.MetaDefault,
		Usage:       10 * time.Second,
		Cost:        1.01,
		ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
	}
	eCGREvent := utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "GenePreRated",
		Event: map[string]interface{}{
			"Account":     "1001",
			"AnswerTime":  time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC),
			"CGRID":       cdr.CGRID,
			"Category":    "call",
			"Cost":        1.01,
			"CostSource":  "",
			"Destination": "+4986517174963",
			"ExtraInfo":   "",
			"OrderID":     int64(123),
			"OriginHost":  "192.168.1.1",
			"OriginID":    "dsafdsaf",
			"Partial":     false,
			"RequestType": utils.MetaRated,
			"RunID":       utils.MetaDefault,
			"SetupTime":   time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC),
			"Source":      "UNIT_TEST",
			"Subject":     "1001",
			"Tenant":      "cgrates.org",
			"ToR":         "*voice",
			"Usage":       10 * time.Second,
			"field_extr1": "val_extr1",
			"fieldextr2":  "valextr2",
			"PreRated":    false,
		},
	}
	cgrEvent := cdr.AsCGREvent()
	if !reflect.DeepEqual(eCGREvent.Tenant, cgrEvent.Tenant) {
		t.Errorf("Expecting: %+v, received: %+v", eCGREvent.Tenant, cgrEvent.Tenant)
	}
	for fldName, fldVal := range eCGREvent.Event {
		if _, has := cgrEvent.Event[fldName]; !has {
			t.Errorf("Expecting: %+v, received: %+v", fldName, nil)
		} else if fldVal != cgrEvent.Event[fldName] {
			t.Errorf("Expecting: %s:%+v, received: %s:%+v",
				fldName, eCGREvent.Event[fldName], fldName, cgrEvent.Event[fldName])
		}
	}
}

func TestCDRAddDefaults(t *testing.T) {
	cdr := &CDR{
		OriginID:   "dsafdsaf",
		OriginHost: "192.168.1.2",
		Account:    "1001",
	}
	cfg := config.NewDefaultCGRConfig()

	eCDR := &CDR{
		CGRID:       "bf736fb56ce586357ab2f286b777187a1612c6e6",
		ToR:         utils.MetaVoice,
		RunID:       utils.MetaDefault,
		Subject:     "1001",
		RequestType: utils.MetaRated,
		Tenant:      "cgrates.org",
		Category:    utils.Call,
		OriginID:    "dsafdsaf",
		OriginHost:  "192.168.1.2",
		Account:     "1001",
	}
	cdr.AddDefaults(cfg)
	if !reflect.DeepEqual(cdr, eCDR) {
		t.Errorf("Expecting: %+v, received: %+v", eCDR, cdr)
	}
}

func TestCDRexportFieldValue(t *testing.T) {
	cdr := &CDR{
		CGRID:       utils.Sha1("dsafdsaf", time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC).String()),
		OrderID:     123,
		ToR:         utils.MetaVoice,
		OriginID:    "dsafdsaf",
		OriginHost:  "192.168.1.1",
		Source:      utils.UnitTest,
		RequestType: utils.MetaRated,
		Tenant:      "cgrates.org",
		Category:    "call",
		Account:     "1001",
		Subject:     "1001",
		Destination: "+4986517174963",
		SetupTime:   time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC),
		AnswerTime:  time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC),
		RunID:       utils.MetaDefault,
		Usage:       10 * time.Second,
		Cost:        1.01,
		ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
	}

	cfgCdrFld := &config.FCTemplate{Path: "*exp.SetupTime", Type: utils.MetaComposed,
		Value: config.NewRSRParsersMustCompile("~SetupTime", utils.InfieldSep), Layout: time.RFC3339}

	eVal := "2013-11-07T08:42:20Z"
	if val, err := cdr.exportFieldValue(cfgCdrFld, nil); err != nil {
		t.Error(err)
	} else if val != eVal {
		t.Errorf("Expecting: %+v, received: %+v", eVal, utils.ToJSON(val))
	}

}

func TestCDReRoundingDecimals(t *testing.T) {
	cdr := &CDR{
		CGRID:       utils.Sha1("dsafdsaf", time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC).String()),
		OrderID:     123,
		ToR:         utils.MetaVoice,
		OriginID:    "dsafdsaf",
		OriginHost:  "192.168.1.1",
		Source:      utils.UnitTest,
		RequestType: utils.MetaRated,
		Tenant:      "cgrates.org",
		Category:    "call",
		Account:     "1001",
		Subject:     "1001",
		Destination: "+4986517174963",
		SetupTime:   time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC),
		AnswerTime:  time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC),
		RunID:       utils.MetaDefault,
		Usage:       10 * time.Second,
		Cost:        1.32165,
		ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
	}

	cfgCdrFld := &config.FCTemplate{
		Path:  "*exp.Cost",
		Type:  utils.MetaComposed,
		Value: config.NewRSRParsersMustCompile("~SetupTime", utils.InfieldSep),
	}

	//5 is the default value for rounding decimals
	eVal := "1.32165"
	if val, err := cdr.exportFieldValue(cfgCdrFld, nil); err != nil {
		t.Error(err)
	} else if val != eVal {
		t.Errorf("Expecting: %+v, received: %+v", eVal, val)
	}

	config.CgrConfig().GeneralCfg().RoundingDecimals = 4
	eVal = "1.3216"
	if val, err := cdr.exportFieldValue(cfgCdrFld, nil); err != nil {
		t.Error(err)
	} else if val != eVal {
		t.Errorf("Expecting: %+v, received: %+v", eVal, val)
	}

	config.CgrConfig().GeneralCfg().RoundingDecimals = 2
	eVal = "1.32"
	if val, err := cdr.exportFieldValue(cfgCdrFld, nil); err != nil {
		t.Error(err)
	} else if val != eVal {
		t.Errorf("Expecting: %+v, received: %+v", eVal, val)
	}

	config.CgrConfig().GeneralCfg().RoundingDecimals = 3
	eVal = "1.322"
	if val, err := cdr.exportFieldValue(cfgCdrFld, nil); err != nil {
		t.Error(err)
	} else if val != eVal {
		t.Errorf("Expecting: %+v, received: %+v", eVal, val)
	}

	config.CgrConfig().GeneralCfg().RoundingDecimals = 1
	eVal = "1.3"
	if val, err := cdr.exportFieldValue(cfgCdrFld, nil); err != nil {
		t.Error(err)
	} else if val != eVal {
		t.Errorf("Expecting: %+v, received: %+v", eVal, val)
	}

	config.CgrConfig().GeneralCfg().RoundingDecimals = 2
	eVal = "1.32"
	if val, err := cdr.exportFieldValue(cfgCdrFld, nil); err != nil {
		t.Error(err)
	} else if val != eVal {
		t.Errorf("Expecting: %+v, received: %+v", eVal, val)
	}

	config.CgrConfig().GeneralCfg().RoundingDecimals = 4
	eVal = "1.3216"
	if val, err := cdr.exportFieldValue(cfgCdrFld, nil); err != nil {
		t.Error(err)
	} else if val != eVal {
		t.Errorf("Expecting: %+v, received: %+v", eVal, val)
	}

	config.CgrConfig().GeneralCfg().RoundingDecimals = 1
	eVal = "1.3"
	if val, err := cdr.exportFieldValue(cfgCdrFld, nil); err != nil {
		t.Error(err)
	} else if val != eVal {
		t.Errorf("Expecting: %+v, received: %+v", eVal, val)
	}

	config.CgrConfig().GeneralCfg().RoundingDecimals = 4
	eVal = "1.3216"
	if val, err := cdr.exportFieldValue(cfgCdrFld, nil); err != nil {
		t.Error(err)
	} else if val != eVal {
		t.Errorf("Expecting: %+v, received: %+v", eVal, val)
	}

	eVal = "1.3216"
	if val, err := cdr.exportFieldValue(cfgCdrFld, nil); err != nil {
		t.Error(err)
	} else if val != eVal {
		t.Errorf("Expecting: %+v, received: %+v", eVal, val)
	}
	config.CgrConfig().GeneralCfg().RoundingDecimals = 3
	eVal = "1.322"
	if val, err := cdr.exportFieldValue(cfgCdrFld, nil); err != nil {
		t.Error(err)
	} else if val != eVal {
		t.Errorf("Expecting: %+v, received: %+v", eVal, val)
	}

	config.CgrConfig().GeneralCfg().RoundingDecimals = 4
	eVal = "1.3216"
	if val, err := cdr.exportFieldValue(cfgCdrFld, nil); err != nil {
		t.Error(err)
	} else if val != eVal {
		t.Errorf("Expecting: %+v, received: %+v", eVal, val)
	}

	//resetore roundingdecimals value
	config.CgrConfig().GeneralCfg().RoundingDecimals = 5
}

func TestCDRcombimedCdrFieldVal(t *testing.T) {
	cdr := &CDR{
		CGRID:       utils.Sha1("dsafdsaf", time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC).String()),
		OrderID:     123,
		ToR:         utils.MetaVoice,
		OriginID:    "dsafdsaf",
		OriginHost:  "192.168.1.1",
		Source:      utils.UnitTest,
		RequestType: utils.MetaRated,
		Tenant:      "cgrates.org",
		Category:    "call",
		Account:     "1001",
		Subject:     "1001",
		Destination: "+4986517174963",
		SetupTime:   time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC),
		AnswerTime:  time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC),
		RunID:       utils.MetaDefault,
		Usage:       10 * time.Second,
		Cost:        1.32165,
		ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
	}
	groupCDRs := []*CDR{
		cdr,
		{
			CGRID:       utils.Sha1("dsafdsaf", time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC).String()),
			OrderID:     124,
			ToR:         utils.MetaVoice,
			OriginID:    "dsafdsaf",
			OriginHost:  "192.168.1.1",
			Source:      utils.UnitTest,
			RequestType: utils.MetaRated,
			Tenant:      "cgrates.org",
			Category:    "call",
			Account:     "1001",
			Subject:     "1001",
			Destination: "+4986517174963",
			SetupTime:   time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC),
			AnswerTime:  time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC),
			RunID:       "testRun1",
			Usage:       10 * time.Second,
			Cost:        1.22,
		},
		{
			CGRID:       utils.Sha1("dsafdsaf", time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC).String()),
			OrderID:     125,
			ToR:         utils.MetaVoice,
			OriginID:    "dsafdsaf",
			OriginHost:  "192.168.1.1",
			Source:      utils.UnitTest,
			RequestType: utils.MetaRated,
			Tenant:      "cgrates.org",
			Category:    "call",
			Account:     "1001",
			Subject:     "1001",
			Destination: "+4986517174963",
			SetupTime:   time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC),
			AnswerTime:  time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC),
			RunID:       "testRun2",
			Usage:       10 * time.Second,
			Cost:        1.764,
		},
	}

	tpFld := &config.FCTemplate{
		Tag:     "TestCombiMed",
		Type:    utils.MetaCombimed,
		Filters: []string{"*string:~*req.RunID:testRun1"},
		Value:   config.NewRSRParsersMustCompile("~*req.Cost", utils.InfieldSep),
	}
	cfg := config.NewDefaultCGRConfig()

	if out, err := cdr.combimedCdrFieldVal(tpFld, groupCDRs, &FilterS{cfg: cfg}); err != nil {
		t.Error(err)
	} else if out != "1.22" {
		t.Errorf("Expected : %+v, received: %+v", "1.22", out)
	}

}
