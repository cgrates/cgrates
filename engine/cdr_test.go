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
		OrderID: 123, ToR: utils.VOICE, OriginID: "dsafdsaf", OriginHost: "192.168.1.1",
		Source: utils.UNIT_TEST, RequestType: utils.META_RATED,
		Tenant: "cgrates.org", Category: "call", Account: "1001", Subject: "1001", Destination: "1002",
		SetupTime: "2013-11-07T08:42:20Z", AnswerTime: "2013-11-07T08:42:26Z", RunID: utils.MetaDefault,
		Usage: "10", Cost: 1.01, PreRated: true,
		ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
	}
	eStorCdr := &CDR{CGRID: utils.Sha1("dsafdsaf", time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC).String()),
		OrderID: 123, ToR: utils.VOICE, OriginID: "dsafdsaf", OriginHost: "192.168.1.1",
		Source: utils.UNIT_TEST, RequestType: utils.META_RATED, RunID: utils.MetaDefault,
		Tenant: "cgrates.org", Category: "call", Account: "1001",
		Subject: "1001", Destination: "1002",
		SetupTime:  time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC),
		AnswerTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC),
		Usage:      time.Duration(10), Cost: 1.01, PreRated: true,
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
		OrderID: 123, ToR: utils.VOICE, OriginID: "dsafdsaf", OriginHost: "192.168.1.1",
		Source: utils.UNIT_TEST, RequestType: utils.META_RATED, Tenant: "cgrates.org",
		Category: "call", Account: "1001", Subject: "1001", Destination: "1002",
		SetupTime:  time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC),
		AnswerTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC),
		RunID:      utils.MetaDefault, Usage: time.Duration(10),
		ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
		Cost:        1.01, PreRated: true,
	}
	if clnStorCdr := storCdr.Clone(); !reflect.DeepEqual(storCdr, clnStorCdr) {
		t.Errorf("Expecting: %+v, received: %+v", storCdr, clnStorCdr)
	}
}

func TestFieldAsString(t *testing.T) {
	cdr := CDR{CGRID: utils.Sha1("dsafdsaf", time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC).String()),
		OrderID: 123, ToR: utils.VOICE, OriginID: "dsafdsaf",
		OriginHost: "192.168.1.1", Source: "test", RequestType: utils.META_RATED,
		Tenant: "cgrates.org", Category: "call", Account: "1001", Subject: "1001",
		Destination: "1002", SetupTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC),
		AnswerTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC), RunID: utils.MetaDefault,
		Usage: time.Duration(10) * time.Second, Cost: 1.01,
		ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
	}
	prsr := config.NewRSRParserMustCompile(utils.DynamicDataPrefix+utils.MetaReq+utils.NestingSep+utils.CGRID, true)
	eFldVal := cdr.CGRID
	if fldVal, err := cdr.FieldAsString(prsr); err != nil {
		t.Error(err)
	} else if fldVal != eFldVal {
		t.Errorf("field: <%v>, expected: <%v>, received: <%v>", prsr, eFldVal, fldVal)
	}
	prsr = config.NewRSRParserMustCompile(utils.DynamicDataPrefix+utils.MetaReq+utils.NestingSep+utils.OrderID, true)
	eFldVal = strconv.FormatInt(cdr.OrderID, 10)
	if fldVal, err := cdr.FieldAsString(prsr); err != nil {
		t.Error(err)
	} else if fldVal != eFldVal {
		t.Errorf("field: <%v>, expected: <%v>, received: <%v>", prsr, eFldVal, fldVal)
	}
	prsr = config.NewRSRParserMustCompile(utils.DynamicDataPrefix+utils.MetaReq+utils.NestingSep+utils.ToR, true)
	eFldVal = cdr.ToR
	if fldVal, err := cdr.FieldAsString(prsr); err != nil {
		t.Error(err)
	} else if fldVal != eFldVal {
		t.Errorf("field: <%v>, expected: <%v>, received: <%v>", prsr, eFldVal, fldVal)
	}
	prsr = config.NewRSRParserMustCompile(utils.DynamicDataPrefix+utils.MetaReq+utils.NestingSep+utils.OriginID, true)
	eFldVal = cdr.OriginID
	if fldVal, err := cdr.FieldAsString(prsr); err != nil {
		t.Error(err)
	} else if fldVal != eFldVal {
		t.Errorf("field: <%v>, expected: <%v>, received: <%v>", prsr, eFldVal, fldVal)
	}
	prsr = config.NewRSRParserMustCompile(utils.DynamicDataPrefix+utils.MetaReq+utils.NestingSep+utils.OriginHost, true)
	eFldVal = cdr.OriginHost
	if fldVal, err := cdr.FieldAsString(prsr); err != nil {
		t.Error(err)
	} else if fldVal != eFldVal {
		t.Errorf("field: <%v>, expected: <%v>, received: <%v>", prsr, eFldVal, fldVal)
	}
	prsr = config.NewRSRParserMustCompile(utils.DynamicDataPrefix+utils.MetaReq+utils.NestingSep+utils.Source, true)
	eFldVal = cdr.Source
	if fldVal, err := cdr.FieldAsString(prsr); err != nil {
		t.Error(err)
	} else if fldVal != eFldVal {
		t.Errorf("field: <%v>, expected: <%v>, received: <%v>", prsr, eFldVal, fldVal)
	}
	prsr = config.NewRSRParserMustCompile(utils.DynamicDataPrefix+utils.MetaReq+utils.NestingSep+utils.RequestType, true)
	eFldVal = cdr.RequestType
	if fldVal, err := cdr.FieldAsString(prsr); err != nil {
		t.Error(err)
	} else if fldVal != eFldVal {
		t.Errorf("field: <%v>, expected: <%v>, received: <%v>", prsr, eFldVal, fldVal)
	}
	prsr = config.NewRSRParserMustCompile(utils.DynamicDataPrefix+utils.MetaReq+utils.NestingSep+utils.Category, true)
	eFldVal = cdr.Category
	if fldVal, err := cdr.FieldAsString(prsr); err != nil {
		t.Error(err)
	} else if fldVal != eFldVal {
		t.Errorf("field: <%v>, expected: <%v>, received: <%v>", prsr, eFldVal, fldVal)
	}
	prsr = config.NewRSRParserMustCompile(utils.DynamicDataPrefix+utils.MetaReq+utils.NestingSep+utils.Account, true)
	eFldVal = cdr.Account
	if fldVal, err := cdr.FieldAsString(prsr); err != nil {
		t.Error(err)
	} else if fldVal != eFldVal {
		t.Errorf("field: <%v>, expected: <%v>, received: <%v>", prsr, eFldVal, fldVal)
	}
	prsr = config.NewRSRParserMustCompile(utils.DynamicDataPrefix+utils.MetaReq+utils.NestingSep+utils.Subject, true)
	eFldVal = cdr.Subject
	if fldVal, err := cdr.FieldAsString(prsr); err != nil {
		t.Error(err)
	} else if fldVal != eFldVal {
		t.Errorf("field: <%v>, expected: <%v>, received: <%v>", prsr, eFldVal, fldVal)
	}
	prsr = config.NewRSRParserMustCompile(utils.DynamicDataPrefix+utils.MetaReq+utils.NestingSep+utils.Destination, true)
	eFldVal = cdr.Destination
	if fldVal, err := cdr.FieldAsString(prsr); err != nil {
		t.Error(err)
	} else if fldVal != eFldVal {
		t.Errorf("field: <%v>, expected: <%v>, received: <%v>", prsr, eFldVal, fldVal)
	}
	prsr = config.NewRSRParserMustCompile(utils.DynamicDataPrefix+utils.MetaReq+utils.NestingSep+utils.SetupTime, true)
	eFldVal = cdr.SetupTime.Format(time.RFC3339)
	if fldVal, err := cdr.FieldAsString(prsr); err != nil {
		t.Error(err)
	} else if fldVal != eFldVal {
		t.Errorf("expected: <%s>, received: <%s>", eFldVal, fldVal)
	}
	prsr = config.NewRSRParserMustCompile(utils.DynamicDataPrefix+utils.MetaReq+utils.NestingSep+utils.AnswerTime, true)
	eFldVal = cdr.AnswerTime.Format(time.RFC3339)
	if fldVal, err := cdr.FieldAsString(prsr); err != nil {
		t.Error(err)
	} else if fldVal != eFldVal {
		t.Errorf("expected: <%s>, received: <%s>", eFldVal, fldVal)
	}
	prsr = config.NewRSRParserMustCompile(utils.DynamicDataPrefix+utils.MetaReq+utils.NestingSep+utils.Usage, true)
	eFldVal = "10s"
	if fldVal, err := cdr.FieldAsString(prsr); err != nil {
		t.Error(err)
	} else if fldVal != eFldVal {
		t.Errorf("field: <%v>, expected: <%v>, received: <%v>", prsr, eFldVal, fldVal)
	}
	prsr = config.NewRSRParserMustCompile(utils.DynamicDataPrefix+utils.MetaReq+utils.NestingSep+utils.RunID, true)
	eFldVal = cdr.RunID
	if fldVal, err := cdr.FieldAsString(prsr); err != nil {
		t.Error(err)
	} else if fldVal != eFldVal {
		t.Errorf("field: <%v>, expected: <%v>, received: <%v>", prsr, eFldVal, fldVal)
	}
	prsr = config.NewRSRParserMustCompile(utils.DynamicDataPrefix+utils.MetaReq+utils.NestingSep+utils.Cost, true)
	eFldVal = strconv.FormatFloat(cdr.Cost, 'f', -1, 64)
	if fldVal, err := cdr.FieldAsString(prsr); err != nil {
		t.Error(err)
	} else if fldVal != eFldVal {
		t.Errorf("field: <%v>, expected: <%v>, received: <%v>", prsr, eFldVal, fldVal)
	}
	prsr = config.NewRSRParserMustCompile(utils.DynamicDataPrefix+utils.MetaReq+utils.NestingSep+"field_extr1", true)
	eFldVal = cdr.ExtraFields["field_extr1"]
	if fldVal, err := cdr.FieldAsString(prsr); err != nil {
		t.Error(err)
	} else if fldVal != eFldVal {
		t.Errorf("field: <%v>, expected: <%v>, received: <%v>", prsr, eFldVal, fldVal)
	}
	prsr = config.NewRSRParserMustCompile(utils.DynamicDataPrefix+utils.MetaReq+utils.NestingSep+"fieldextr2", true)
	eFldVal = cdr.ExtraFields["fieldextr2"]
	if fldVal, err := cdr.FieldAsString(prsr); err != nil {
		t.Error(err)
	} else if fldVal != eFldVal {
		t.Errorf("field: <%v>, expected: <%v>, received: <%v>", prsr, eFldVal, fldVal)
	}
	prsr = config.NewRSRParserMustCompile(utils.DynamicDataPrefix+utils.MetaReq+utils.NestingSep+"dummy_field", true)
	if fldVal, err := cdr.FieldAsString(prsr); err != utils.ErrNotFound {
		t.Error(err)
	} else if fldVal != utils.EmptyString {
		t.Errorf("field: <%v>, expected: <%v>, received: <%v>", prsr, utils.EmptyString, fldVal)
	}
}

func TestFieldsAsString(t *testing.T) {
	cdr := CDR{CGRID: utils.Sha1("dsafdsaf", time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC).String()),
		OrderID: 123, ToR: utils.VOICE, OriginID: "dsafdsaf", OriginHost: "192.168.1.1", Source: "test",
		RequestType: utils.META_RATED, Tenant: "cgrates.org",
		Category: "call", Account: "1001", Subject: "1001", Destination: "1002",
		SetupTime:  time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC),
		AnswerTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC),
		RunID:      utils.MetaDefault, Usage: time.Duration(10) * time.Second,
		ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"}, Cost: 1.01,
	}
	eVal := "call_from_1001"
	if val := cdr.FieldsAsString(
		config.NewRSRParsersMustCompile("~*req.Category;_from_;~*req.Account", true, utils.INFIELD_SEP)); val != eVal {
		t.Errorf("Expecting : %s, received: %q", eVal, val)
	}
}

func TestFieldAsStringForCostDetails(t *testing.T) {
	cc := &CallCost{
		Category:    "generic",
		Tenant:      "cgrates.org",
		Subject:     "1001",
		Account:     "1001",
		Destination: "data",
		ToR:         "*data",
		Cost:        0,
		AccountSummary: &AccountSummary{
			Tenant: "cgrates.org",
			ID:     "AccountFromAccountSummary",
			BalanceSummaries: []*BalanceSummary{
				&BalanceSummary{
					UUID:  "f9be602747f4",
					ID:    "monetary",
					Type:  utils.MONETARY,
					Value: 0.5,
				},
				&BalanceSummary{
					UUID:  "2e02510ab90a",
					ID:    "voice",
					Type:  utils.VOICE,
					Value: 10,
				},
			},
		},
	}

	cdr := &CDR{
		CGRID:       utils.Sha1("dsafdsaf", time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC).String()),
		OrderID:     123,
		ToR:         utils.VOICE,
		OriginID:    "dsafdsaf",
		OriginHost:  "192.168.1.1",
		Source:      utils.UNIT_TEST,
		RequestType: utils.META_RATED,
		Tenant:      "cgrates.org",
		Category:    "call",
		Account:     "1002",
		Subject:     "1001",
		Destination: "+4986517174963",
		SetupTime:   time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC),
		AnswerTime:  time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC),
		RunID:       utils.MetaDefault,
		Usage:       time.Duration(10) * time.Second,
		ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
		Cost:        1.01,
		CostDetails: NewEventCostFromCallCost(cc, "TestCDRTestCDRAsMapStringIface2", utils.MetaDefault),
	}

	prsr := config.NewRSRParserMustCompile("~*req.CostDetails.CGRID", true)
	eFldVal := "TestCDRTestCDRAsMapStringIface2"
	if fldVal, err := cdr.FieldAsString(prsr); err != nil {
		t.Error(err)
	} else if fldVal != eFldVal {
		t.Errorf("field: <%v>, expected: <%v>, received: <%v>", prsr, eFldVal, fldVal)
	}

	prsr = config.NewRSRParserMustCompile("~*req.CostDetails.AccountSummary.ID", true)
	eFldVal = "AccountFromAccountSummary"
	if fldVal, err := cdr.FieldAsString(prsr); err != nil {
		t.Error(err)
	} else if fldVal != eFldVal {
		t.Errorf("field: <%v>, expected: <%v>, received: <%v>", prsr, eFldVal, fldVal)
	}

	prsr = config.NewRSRParserMustCompile("~*req.CostDetails.AccountSummary.BalanceSummaries[1].ID", true)
	eFldVal = "voice"
	if fldVal, err := cdr.FieldAsString(prsr); err != nil {
		t.Error(err)
	} else if fldVal != eFldVal {
		t.Errorf("field: <%v>, expected: <%v>, received: <%v>", prsr, eFldVal, fldVal)
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

/*
func TestCDRAsHttpForm(t *testing.T) {
	storCdr := CDR{CGRID: utils.Sha1("dsafdsaf", time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC).String()), OrderID: 123, ToR: utils.VOICE, OriginID: "dsafdsaf",
		OriginHost: "192.168.1.1", Source: utils.UNIT_TEST, RequestType: utils.META_RATED,
		Tenant: "cgrates.org", Category: "call", Account: "1001", Subject: "1001", Destination: "1002",
		SetupTime: time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC), AnswerTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC), RunID: utils.MetaDefault,
		Usage: time.Duration(10) * time.Second, Supplier: "SUPPL1",
		ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"}, Cost: 1.01,
	}
	cdrForm := storCdr.AsHttpForm()
	if cdrForm.Get(utils.ToR) != utils.VOICE {
		t.Errorf("Expected: %s, received: %s", utils.VOICE, cdrForm.Get(utils.ToR))
	}
	if cdrForm.Get(utils.OriginID) != "dsafdsaf" {
		t.Errorf("Expected: %s, received: %s", "dsafdsaf", cdrForm.Get(utils.OriginID))
	}
	if cdrForm.Get(utils.OriginHost) != "192.168.1.1" {
		t.Errorf("Expected: %s, received: %s", "192.168.1.1", cdrForm.Get(utils.OriginHost))
	}
	if cdrForm.Get(utils.Source) != utils.UNIT_TEST {
		t.Errorf("Expected: %s, received: %s", utils.UNIT_TEST, cdrForm.Get(utils.Source))
	}
	if cdrForm.Get(utils.RequestType) != utils.META_RATED {
		t.Errorf("Expected: %s, received: %s", utils.META_RATED, cdrForm.Get(utils.RequestType))
	}
	if cdrForm.Get(utils.Tenant) != "cgrates.org" {
		t.Errorf("Expected: %s, received: %s", "cgrates.org", cdrForm.Get(utils.Tenant))
	}
	if cdrForm.Get(utils.Category) != "call" {
		t.Errorf("Expected: %s, received: %s", "call", cdrForm.Get(utils.Category))
	}
	if cdrForm.Get(utils.ACCOUNT) != "1001" {
		t.Errorf("Expected: %s, received: %s", "1001", cdrForm.Get(utils.ACCOUNT))
	}
	if cdrForm.Get(utils.Subject) != "1001" {
		t.Errorf("Expected: %s, received: %s", "1001", cdrForm.Get(utils.Subject))
	}
	if cdrForm.Get(utils.Destination) != "1002" {
		t.Errorf("Expected: %s, received: %s", "1002", cdrForm.Get(utils.Destination))
	}
	if cdrForm.Get(utils.SetupTime) != "2013-11-07T08:42:20Z" {
		t.Errorf("Expected: %s, received: %s", "2013-11-07T08:42:20Z", cdrForm.Get(utils.SetupTime))
	}
	if cdrForm.Get(utils.AnswerTime) != "2013-11-07T08:42:26Z" {
		t.Errorf("Expected: %s, received: %s", "2013-11-07T08:42:26Z", cdrForm.Get(utils.AnswerTime))
	}
	if cdrForm.Get(utils.Usage) != "10" {
		t.Errorf("Expected: %s, received: %s", "10", cdrForm.Get(utils.Usage))
	}
	if cdrForm.Get(utils.SUPPLIER) != "SUPPL1" {
		t.Errorf("Expected: %s, received: %s", "1001", cdrForm.Get(utils.SUPPLIER))
	}
	if cdrForm.Get("field_extr1") != "val_extr1" {
		t.Errorf("Expected: %s, received: %s", "val_extr1", cdrForm.Get("field_extr1"))
	}
	if cdrForm.Get("fieldextr2") != "valextr2" {
		t.Errorf("Expected: %s, received: %s", "valextr2", cdrForm.Get("fieldextr2"))
	}
}
*/

func TestCDRAsExternalCDR(t *testing.T) {
	storCdr := CDR{CGRID: utils.Sha1("dsafdsaf", time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC).String()),
		OrderID: 123, ToR: utils.VOICE, OriginID: "dsafdsaf", OriginHost: "192.168.1.1",
		Source: utils.UNIT_TEST, RequestType: utils.META_RATED,
		Tenant: "cgrates.org", Category: "call", Account: "1001", Subject: "1001", Destination: "1002",
		SetupTime:  time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC),
		AnswerTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC),
		RunID:      utils.MetaDefault, Usage: time.Duration(10 * time.Second), Cost: 1.01,
		ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"}}
	expectOutCdr := &ExternalCDR{
		CGRID:   utils.Sha1("dsafdsaf", time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC).String()),
		OrderID: 123, ToR: utils.VOICE, OriginID: "dsafdsaf", OriginHost: "192.168.1.1",
		Source: utils.UNIT_TEST, RequestType: utils.META_RATED,
		Tenant: "cgrates.org", Category: "call", Account: "1001", Subject: "1001", Destination: "1002",
		SetupTime: "2013-11-07T08:42:20Z", AnswerTime: "2013-11-07T08:42:26Z", RunID: utils.MetaDefault,
		Usage: "10s", Cost: 1.01, CostDetails: "null",
		ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"}}
	if cdrOut := storCdr.AsExternalCDR(); !reflect.DeepEqual(expectOutCdr, cdrOut) {
		t.Errorf("Expected: %+v, received: %+v", expectOutCdr, cdrOut)
	}
}

func TesUsageReqAsCDR(t *testing.T) {
	setupReq := &UsageRecord{ToR: utils.VOICE, RequestType: utils.META_RATED,
		Tenant: "cgrates.org", Category: "call",
		Account: "1001", Subject: "1001", Destination: "1002",
		SetupTime: "2013-11-07T08:42:20Z", AnswerTime: "2013-11-07T08:42:26Z",
		Usage: "0.00000001"}
	eStorCdr := &CDR{ToR: utils.VOICE, RequestType: utils.META_RATED,
		Tenant: "cgrates.org", Category: "call", Account: "1001",
		Subject: "1001", Destination: "1002",
		SetupTime:  time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC),
		AnswerTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC),
		Usage:      time.Duration(10)}
	if CDR, err := setupReq.AsCDR(""); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eStorCdr, CDR) {
		t.Errorf("Expected: %+v, received: %+v", eStorCdr, CDR)
	}
}

func TestUsageReqAsCD(t *testing.T) {
	req := &UsageRecord{ToR: utils.VOICE, RequestType: utils.META_RATED,
		Tenant: "cgrates.org", Category: "call",
		Account: "1001", Subject: "1001", Destination: "1002",
		SetupTime: "2013-11-07T08:42:20Z", AnswerTime: "2013-11-07T08:42:26Z",
		Usage: "10",
	}
	eCD := &CallDescriptor{CgrID: "c4630df20b2a0c5b11311e4b5a8c3178cf314344", ToR: req.ToR,
		Tenant:   req.Tenant,
		Category: req.Category, Account: req.Account,
		Subject: req.Subject, Destination: req.Destination,
		TimeStart:           time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC),
		TimeEnd:             time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC).Add(time.Duration(10)),
		DenyNegativeAccount: true}
	if cd, err := req.AsCallDescriptor("", true); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCD, cd) {
		t.Errorf("Expected: %+v, received: %+v", eCD, cd)
	}
}

func TestCdrClone(t *testing.T) {
	cdr := &CDR{}
	eOut := &CDR{}
	if rcv := cdr.Clone(); !reflect.DeepEqual(rcv, eOut) {
		t.Errorf("Expecting: %+v, received: %+v", eOut, rcv)
	}
	cdr = &CDR{
		CGRID:     "CGRID_test",
		OrderID:   18,
		SetupTime: time.Date(2020, time.April, 18, 23, 0, 4, 0, time.UTC),
		Usage:     time.Duration(10),
		ExtraFields: map[string]string{
			"test1": "_test1_",
			"test2": "_test2_",
		},
		Partial: true,
		Cost:    0.74,
		CostDetails: &EventCost{
			CGRID: "EventCost_CGRID",
			Cost:  utils.Float64Pointer(0.74),
		},
	}
	eOut = &CDR{
		CGRID:     "CGRID_test",
		OrderID:   18,
		SetupTime: time.Date(2020, time.April, 18, 23, 0, 4, 0, time.UTC),
		Usage:     time.Duration(10),
		ExtraFields: map[string]string{
			"test1": "_test1_",
			"test2": "_test2_",
		},
		Partial: true,
		Cost:    0.74,
		CostDetails: &EventCost{
			CGRID: "EventCost_CGRID",
			Cost:  utils.Float64Pointer(0.74),
		},
	}
	eOut.CostDetails.initCache()
	if rcv := cdr.Clone(); !reflect.DeepEqual(rcv, eOut) {
		t.Errorf("Expecting: %+v,\n received: %+v", eOut, rcv)
	}

}
func TestCDRAsMapStringIface(t *testing.T) {
	cdr := &CDR{
		CGRID:       utils.Sha1("dsafdsaf", time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC).String()),
		OrderID:     123,
		ToR:         utils.VOICE,
		OriginID:    "dsafdsaf",
		OriginHost:  "192.168.1.1",
		Source:      utils.UNIT_TEST,
		RequestType: utils.META_RATED,
		Tenant:      "cgrates.org",
		Category:    "call",
		Account:     "1002",
		Subject:     "1001",
		Destination: "+4986517174963",
		SetupTime:   time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC),
		AnswerTime:  time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC),
		RunID:       utils.MetaDefault,
		Usage:       time.Duration(10) * time.Second,
		ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
		Cost:        1.01,
	}

	mp := map[string]interface{}{
		"field_extr1":     "val_extr1",
		"fieldextr2":      "valextr2",
		utils.CGRID:       cdr.CGRID,
		utils.RunID:       utils.MetaDefault,
		utils.OrderID:     cdr.OrderID,
		utils.OriginHost:  "192.168.1.1",
		utils.Source:      utils.UNIT_TEST,
		utils.OriginID:    "dsafdsaf",
		utils.ToR:         utils.VOICE,
		utils.RequestType: utils.META_RATED,
		utils.Tenant:      "cgrates.org",
		utils.Category:    "call",
		utils.Account:     "1002",
		utils.Subject:     "1001",
		utils.Destination: "+4986517174963",
		utils.SetupTime:   time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC),
		utils.AnswerTime:  time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC),
		utils.Usage:       time.Duration(10) * time.Second,
		utils.CostSource:  cdr.CostSource,
		utils.Cost:        1.01,
		utils.PreRated:    false,
		utils.Partial:     false,
		utils.ExtraInfo:   cdr.ExtraInfo,
	}
	if cdrMp := cdr.AsMapStringIface(); !reflect.DeepEqual(mp, cdrMp) {
		t.Errorf("Expecting: %+v, received: %+v", mp, cdrMp)
	}
}

func TestCDRTestCDRAsMapStringIface2(t *testing.T) {
	cc := &CallCost{
		Category:    "generic",
		Tenant:      "cgrates.org",
		Subject:     "1001",
		Account:     "1001",
		Destination: "data",
		ToR:         "*data",
		Cost:        0,
	}

	cdr := &CDR{
		CGRID:       utils.Sha1("dsafdsaf", time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC).String()),
		OrderID:     123,
		ToR:         utils.VOICE,
		OriginID:    "dsafdsaf",
		OriginHost:  "192.168.1.1",
		Source:      utils.UNIT_TEST,
		RequestType: utils.META_RATED,
		Tenant:      "cgrates.org",
		Category:    "call",
		Account:     "1002",
		Subject:     "1001",
		Destination: "+4986517174963",
		SetupTime:   time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC),
		AnswerTime:  time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC),
		RunID:       utils.MetaDefault,
		Usage:       time.Duration(10) * time.Second,
		ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
		Cost:        1.01,
		CostDetails: NewEventCostFromCallCost(cc, "TestCDRTestCDRAsMapStringIface2", utils.MetaDefault),
	}

	mp := map[string]interface{}{
		"field_extr1":     "val_extr1",
		"fieldextr2":      "valextr2",
		utils.CGRID:       cdr.CGRID,
		utils.RunID:       utils.MetaDefault,
		utils.OrderID:     cdr.OrderID,
		utils.OriginHost:  "192.168.1.1",
		utils.Source:      utils.UNIT_TEST,
		utils.OriginID:    "dsafdsaf",
		utils.ToR:         utils.VOICE,
		utils.RequestType: utils.META_RATED,
		utils.Tenant:      "cgrates.org",
		utils.Category:    "call",
		utils.Account:     "1002",
		utils.Subject:     "1001",
		utils.Destination: "+4986517174963",
		utils.SetupTime:   time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC),
		utils.AnswerTime:  time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC),
		utils.Usage:       time.Duration(10) * time.Second,
		utils.CostSource:  cdr.CostSource,
		utils.Cost:        1.01,
		utils.PreRated:    false,
		utils.Partial:     false,
		utils.ExtraInfo:   cdr.ExtraInfo,
		utils.CostDetails: cdr.CostDetails,
	}
	if cdrMp := cdr.AsMapStringIface(); !reflect.DeepEqual(mp, cdrMp) {
		t.Errorf("Expecting: %+v, received: %+v", mp, cdrMp)
	}
}

func TestCDRAsExportRecord(t *testing.T) {
	cc := &CallCost{
		Category:    "generic",
		Tenant:      "cgrates.org",
		Subject:     "1001",
		Account:     "1001",
		Destination: "data",
		ToR:         "*data",
		Cost:        0,
		AccountSummary: &AccountSummary{
			Tenant: "cgrates.org",
			ID:     "AccountFromAccountSummary",
		},
	}
	eventCost := NewEventCostFromCallCost(cc, "TestCDRTestCDRAsMapStringIface2", utils.MetaDefault)
	eventCost.RatingFilters = RatingFilters{
		"3d99c91": RatingMatchedFilters{
			"DestinationID":     "CustomDestination",
			"DestinationPrefix": "26377",
			"RatingPlanID":      "RP_ZW_v1",
		},
	}

	cdr := &CDR{
		CGRID: utils.Sha1("dsafdsaf",
			time.Unix(1383813745, 0).UTC().String()),
		ToR: utils.VOICE, OriginID: "dsafdsaf",
		OriginHost:  "192.168.1.1",
		RequestType: utils.META_RATED,
		Tenant:      "cgrates.org",
		Category:    "call",
		Account:     "1001",
		Subject:     "1001",
		Destination: "+4986517174963",
		SetupTime:   time.Unix(1383813745, 0).UTC(),
		AnswerTime:  time.Unix(1383813746, 0).UTC(),
		Usage:       time.Duration(10) * time.Second,
		RunID:       utils.MetaDefault, Cost: 1.01,
		ExtraFields: map[string]string{"stop_time": "2014-06-11 19:19:00 +0000 UTC", "fieldextr2": "valextr2"},
		CostDetails: eventCost,
	}

	prsr := config.NewRSRParsersMustCompile(utils.DynamicDataPrefix+utils.MetaReq+utils.NestingSep+utils.Destination, true, utils.INFIELD_SEP)
	cfgCdrFld := &config.FCTemplate{
		Tag:       "destination",
		Path:      utils.PathItems{{Field: utils.MetaExp}, {Field: "Destination"}},
		PathSlice: []string{utils.MetaExp, "Destination"},
		Type:      utils.META_COMPOSED,
		Value:     prsr,
		Timezone:  "UTC",
	}
	if expRecord, err := cdr.AsExportRecord([]*config.FCTemplate{cfgCdrFld}, false, nil, nil); err != nil {
		t.Error(err)
	} else if expRecord[0] != cdr.Destination {
		t.Errorf("Expecting:\n%s\nReceived:\n%s", cdr.Destination, expRecord)
	}
	if err := dm.SetReverseDestination(&Destination{Id: "MASKED_DESTINATIONS", Prefixes: []string{"+4986517174963"}},
		utils.NonTransactional); err != nil {
		t.Error(err)
	}

	cfgCdrFld = &config.FCTemplate{
		Tag:        "Destination",
		Path:       utils.PathItems{{Field: utils.MetaExp}, {Field: "Destination"}},
		PathSlice:  []string{utils.MetaExp, "Destination"},
		Type:       utils.META_COMPOSED,
		Value:      prsr,
		MaskDestID: "MASKED_DESTINATIONS",
		MaskLen:    3,
	}
	eDst := "+4986517174***"
	if expRecord, err := cdr.AsExportRecord([]*config.FCTemplate{cfgCdrFld}, false, nil, nil); err != nil {
		t.Error(err)
	} else if expRecord[0] != eDst {
		t.Errorf("Expecting:\n%s\nReceived:\n%s", eDst, expRecord[0])
	}

	cfgCdrFld = &config.FCTemplate{
		Tag:        "MaskedDest",
		Path:       utils.PathItems{{Field: utils.MetaExp}, {Field: "MaskedDest"}},
		PathSlice:  []string{utils.MetaExp, "MaskedDest"},
		Type:       utils.MetaMaskedDestination,
		Value:      prsr,
		MaskDestID: "MASKED_DESTINATIONS",
	}
	if expRecord, err := cdr.AsExportRecord([]*config.FCTemplate{cfgCdrFld}, false, nil, nil); err != nil {
		t.Error(err)
	} else if expRecord[0] != "1" {
		t.Errorf("Expecting:\n%s\nReceived:\n%s", "1", expRecord[0])
	}
	defaultCfg, _ := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, true, defaultCfg.DataDbCfg().Items)
	dmForCDR := NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	cfgCdrFld = &config.FCTemplate{
		Tag:       "destination",
		Path:      utils.PathItems{{Field: utils.MetaExp}, {Field: "Destination"}},
		PathSlice: []string{utils.MetaExp, "Destination"},
		Type:      utils.META_COMPOSED,
		Value:     prsr,
		Filters:   []string{"*string:~*req.Tenant:itsyscom.com"},
		Timezone:  "UTC",
	}
	if rcrd, err := cdr.AsExportRecord([]*config.FCTemplate{cfgCdrFld}, false, nil, &FilterS{dm: dmForCDR, cfg: defaultCfg}); err != nil {
		t.Error(err)
	} else if len(rcrd) != 0 {
		t.Error("failed using filter")
	}

	// Test MetaDateTime
	prsr = config.NewRSRParsersMustCompile("~*req.stop_time", true, utils.INFIELD_SEP)
	layout := "2006-01-02 15:04:05"
	cfgCdrFld = &config.FCTemplate{
		Tag:       "stop_time",
		Type:      utils.MetaDateTime,
		Path:      utils.PathItems{{Field: utils.MetaExp}, {Field: "stop_time"}},
		PathSlice: []string{utils.MetaExp, "stop_time"},
		Value:     prsr,
		Layout:    layout,
		Timezone:  "UTC",
	}
	if expRecord, err := cdr.AsExportRecord([]*config.FCTemplate{cfgCdrFld}, false, nil, &FilterS{dm: dmForCDR, cfg: defaultCfg}); err != nil {
		t.Error(err)
	} else if expRecord[0] != "2014-06-11 19:19:00" {
		t.Error("Expecting: 2014-06-11 19:19:00, got: ", expRecord[0])
	}

	// Test filter
	cfgCdrFld = &config.FCTemplate{
		Tag:       "stop_time",
		Type:      utils.MetaDateTime,
		Path:      utils.PathItems{{Field: utils.MetaExp}, {Field: "stop_time"}},
		PathSlice: []string{utils.MetaExp, "stop_time"},
		Value:     prsr,
		Filters:   []string{"*string:~*req.Tenant:itsyscom.com"},
		Layout:    layout,
		Timezone:  "UTC",
	}
	if rcrd, err := cdr.AsExportRecord([]*config.FCTemplate{cfgCdrFld}, false, nil, &FilterS{dm: dmForCDR, cfg: defaultCfg}); err != nil {
		t.Error(err)
	} else if len(rcrd) != 0 {
		t.Error("failed using filter")
	}

	prsr = config.NewRSRParsersMustCompile("~*req.fieldextr2", true, utils.INFIELD_SEP)
	cfgCdrFld = &config.FCTemplate{
		Tag:       "stop_time",
		Type:      utils.MetaDateTime,
		Path:      utils.PathItems{{Field: utils.MetaExp}, {Field: "stop_time"}},
		PathSlice: []string{utils.MetaExp, "stop_time"},
		Value:     prsr,
		Layout:    layout,
		Timezone:  "UTC"}
	// Test time parse error
	if _, err := cdr.AsExportRecord([]*config.FCTemplate{cfgCdrFld}, false, nil, nil); err == nil {
		t.Error("Should give error here, got none.")
	}

	prsr = config.NewRSRParsersMustCompile("~*req.CostDetails.CGRID", true, utils.INFIELD_SEP)
	cfgCdrFld = &config.FCTemplate{
		Tag:       "CGRIDFromCostDetails",
		Type:      utils.META_COMPOSED,
		Path:      utils.PathItems{{Field: utils.MetaExp}, {Field: "CGRIDFromCostDetails"}},
		PathSlice: []string{utils.MetaExp, "CGRIDFromCostDetails"},
		Value:     prsr,
	}
	if expRecord, err := cdr.AsExportRecord([]*config.FCTemplate{cfgCdrFld}, false, nil, nil); err != nil {
		t.Error(err)
	} else if expRecord[0] != cdr.CostDetails.CGRID {
		t.Errorf("Expecting:\n%s\nReceived:\n%s", cdr.CostDetails.CGRID, expRecord)
	}
	prsr = config.NewRSRParsersMustCompile("~*req.CostDetails.AccountSummary.ID", true, utils.INFIELD_SEP)
	cfgCdrFld = &config.FCTemplate{
		Tag:       "AccountID",
		Type:      utils.META_COMPOSED,
		Path:      utils.PathItems{{Field: utils.MetaExp}, {Field: "CustomAccountID"}},
		PathSlice: []string{utils.MetaExp, "CustomAccountID"},
		Value:     prsr,
	}
	if expRecord, err := cdr.AsExportRecord([]*config.FCTemplate{cfgCdrFld}, false, nil, nil); err != nil {
		t.Error(err)
	} else if expRecord[0] != cdr.CostDetails.AccountSummary.ID {
		t.Errorf("Expecting:\n%s\nReceived:\n%s", cdr.CostDetails.AccountSummary.ID, expRecord)
	}

	expected := `{"3d99c91":{"DestinationID":"CustomDestination","DestinationPrefix":"26377","RatingPlanID":"RP_ZW_v1"}}`
	prsr = config.NewRSRParsersMustCompile("~*req.CostDetails.RatingFilters", true, utils.INFIELD_SEP)
	cfgCdrFld = &config.FCTemplate{
		Tag:       "DestinationID",
		Type:      utils.META_COMPOSED,
		Path:      utils.PathItems{{Field: utils.MetaExp}, {Field: "CustomDestinationID"}},
		PathSlice: []string{utils.MetaExp, "CustomDestinationID"},
		Value:     prsr,
	}
	if expRecord, err := cdr.AsExportRecord([]*config.FCTemplate{cfgCdrFld}, false, nil, nil); err != nil {
		t.Error(err)
	} else if expRecord[0] != expected {
		t.Errorf("Expecting: <%q>,\n Received: <%q>", expected, expRecord[0])
	}

	expected = "RP_ZW_v1"
	prsr = config.NewRSRParsersMustCompile("~*req.CostDetails.RatingFilters:s/RatingPlanID\"\\s?\\:\\s?\"([^\"]*)\".*/$1/", true, utils.INFIELD_SEP)
	cfgCdrFld = &config.FCTemplate{
		Tag:       "DestinationID",
		Type:      utils.META_COMPOSED,
		Path:      utils.PathItems{{Field: utils.MetaExp}, {Field: "CustomDestinationID"}},
		PathSlice: []string{utils.MetaExp, "CustomDestinationID"},
		Value:     prsr,
	}
	if expRecord, err := cdr.AsExportRecord([]*config.FCTemplate{cfgCdrFld}, false, nil, nil); err != nil {
		t.Error(err)
	} else if expRecord[0] != expected {
		t.Errorf("Expecting: <%q>,\n Received: <%q>", expected, expRecord[0])
	}

}

func TestCDRAsExportMap(t *testing.T) {
	cdr := &CDR{CGRID: utils.Sha1("dsafdsaf", time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC).String()),
		OrderID: 123, ToR: utils.VOICE, OriginID: "dsafdsaf",
		OriginHost: "192.168.1.1", Source: utils.UNIT_TEST, RequestType: utils.META_RATED,
		Tenant: "cgrates.org", Category: "call", Account: "1001",
		Subject: "1001", Destination: "+4986517174963",
		SetupTime:  time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC),
		AnswerTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC), RunID: utils.MetaDefault,
		Usage: time.Duration(10) * time.Second, Cost: 1.01,
		ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
	}
	eCDRMp := map[string]string{
		utils.CGRID:       cdr.CGRID,
		utils.Destination: "004986517174963",
		"FieldExtra1":     "val_extr1",
	}
	expFlds := []*config.FCTemplate{
		&config.FCTemplate{Path: utils.PathItems{{Field: utils.MetaExp}, {Field: utils.CGRID}}, Type: utils.META_COMPOSED,
			PathSlice: []string{utils.MetaExp, utils.CGRID},
			Value:     config.NewRSRParsersMustCompile(utils.DynamicDataPrefix+utils.MetaReq+utils.NestingSep+utils.CGRID, true, utils.INFIELD_SEP)},
		&config.FCTemplate{Path: utils.PathItems{{Field: utils.MetaExp}, {Field: utils.Destination}}, Type: utils.META_COMPOSED,
			PathSlice: []string{utils.MetaExp, utils.Destination},
			Value:     config.NewRSRParsersMustCompile("~*req.Destination:s/^\\+(\\d+)$/00${1}/", true, utils.INFIELD_SEP)},
		&config.FCTemplate{Path: utils.PathItems{{Field: utils.MetaExp}, {Field: "FieldExtra1"}}, Type: utils.META_COMPOSED,
			PathSlice: []string{utils.MetaExp, "FieldExtra1"},
			Value:     config.NewRSRParsersMustCompile("~*req.field_extr1", true, utils.INFIELD_SEP)},
	}
	if cdrMp, err := cdr.AsExportMap(expFlds, false, nil, nil); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCDRMp, cdrMp) {
		t.Errorf("Expecting: %+v, received: %+v", eCDRMp, cdrMp)
	}
}

func TestCDRAsCDRsql(t *testing.T) {
	cdr := &CDR{CGRID: utils.Sha1("dsafdsaf", time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC).String()),
		OrderID:     123,
		ToR:         utils.VOICE,
		OriginID:    "dsafdsaf",
		OriginHost:  "192.168.1.1",
		Source:      utils.UNIT_TEST,
		RequestType: utils.META_RATED,
		Tenant:      "cgrates.org",
		Category:    "call",
		Account:     "1001",
		Subject:     "1001",
		Destination: "+4986517174963",
		SetupTime:   time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC),
		AnswerTime:  time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC),
		RunID:       utils.MetaDefault,
		Usage:       time.Duration(10) * time.Second,
		Cost:        1.01,
		ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
	}
	eCDR := cdr.AsCDRsql()
	eCDRSql := &CDRsql{
		Cgrid:       cdr.CGRID,
		RunID:       cdr.RunID,
		OriginID:    "dsafdsaf",
		TOR:         utils.VOICE,
		Source:      utils.UNIT_TEST,
		Tenant:      "cgrates.org",
		Category:    "call",
		Account:     "1001",
		Subject:     "1001",
		Destination: "+4986517174963",
		SetupTime:   time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC),
		AnswerTime:  time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC),
		Usage:       cdr.Usage.Nanoseconds(),
		Cost:        cdr.Cost,
		ExtraFields: utils.ToJSON(cdr.ExtraFields),
		RequestType: cdr.RequestType,
		OriginHost:  cdr.OriginHost,
		CostDetails: utils.ToJSON(cdr.CostDetails),
		CreatedAt:   eCDR.CreatedAt,
	}
	if !reflect.DeepEqual(eCDR, eCDRSql) {
		t.Errorf("Expecting: %+v, received: %+v", eCDR, eCDRSql)
	}

}

func TestCDRNewCDRFromSQL(t *testing.T) {
	extraFields := map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"}
	cdrSql := &CDRsql{
		ID:          123,
		Cgrid:       "abecd993d06672714c4218a6dcf8278e0589a171",
		RunID:       utils.MetaDefault,
		OriginID:    "dsafdsaf",
		TOR:         utils.VOICE,
		Source:      utils.UNIT_TEST,
		Tenant:      "cgrates.org",
		Category:    "call",
		Account:     "1001",
		Subject:     "1001",
		Destination: "+4986517174963",
		SetupTime:   time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC),
		AnswerTime:  time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC),
		Usage:       10000000000,
		Cost:        1.01,
		RequestType: utils.META_RATED,
		OriginHost:  "192.168.1.1",
		ExtraFields: utils.ToJSON(extraFields),
	}

	cdr := &CDR{CGRID: utils.Sha1("dsafdsaf", time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC).String()),
		OrderID:     123,
		ToR:         utils.VOICE,
		OriginID:    "dsafdsaf",
		OriginHost:  "192.168.1.1",
		Source:      utils.UNIT_TEST,
		RequestType: utils.META_RATED,
		Tenant:      "cgrates.org",
		Category:    "call",
		Account:     "1001",
		Subject:     "1001",
		Destination: "+4986517174963",
		SetupTime:   time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC),
		AnswerTime:  time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC),
		RunID:       utils.MetaDefault,
		Usage:       time.Duration(10) * time.Second,
		Cost:        1.01,
		ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
	}

	if eCDR, err := NewCDRFromSQL(cdrSql); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(cdr, eCDR) {
		t.Errorf("Expecting: %+v, received: %+v", cdr, eCDR)
	}

}

func TestCDRAsCGREvent(t *testing.T) {
	cdr := &CDR{
		CGRID:       utils.Sha1("dsafdsaf", time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC).String()),
		OrderID:     123,
		ToR:         utils.VOICE,
		OriginID:    "dsafdsaf",
		OriginHost:  "192.168.1.1",
		Source:      utils.UNIT_TEST,
		RequestType: utils.META_RATED,
		Tenant:      "cgrates.org",
		Category:    "call",
		Account:     "1001",
		Subject:     "1001",
		Destination: "+4986517174963",
		SetupTime:   time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC),
		AnswerTime:  time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC),
		RunID:       utils.MetaDefault,
		Usage:       time.Duration(10) * time.Second,
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
			"RequestType": utils.META_RATED,
			"RunID":       utils.MetaDefault,
			"SetupTime":   time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC),
			"Source":      "UNIT_TEST",
			"Subject":     "1001",
			"Tenant":      "cgrates.org",
			"ToR":         "*voice",
			"Usage":       time.Duration(10) * time.Second,
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

func TestCDRUpdateFromCGREvent(t *testing.T) {
	cdr := &CDR{
		CGRID:       utils.Sha1("dsafdsaf", time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC).String()),
		OrderID:     123,
		ToR:         utils.VOICE,
		OriginID:    "dsafdsaf",
		OriginHost:  "192.168.1.1",
		Source:      utils.UNIT_TEST,
		RequestType: utils.META_RATED,
		Tenant:      "cgrates.org",
		Category:    "call",
		Account:     "1001",
		Subject:     "1001",
		Destination: "+4986517174963",
		SetupTime:   time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC),
		AnswerTime:  time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC),
		RunID:       utils.MetaDefault,
		Usage:       time.Duration(10) * time.Second,
		Cost:        1.01,
		ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
	}
	var costdetails *CallCost
	cgrEvent := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "GenePreRated",
		Event: map[string]interface{}{
			"Account":     "1001",
			"AnswerTime":  time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC),
			"CGRID":       cdr.CGRID,
			"Category":    "call",
			"Cost":        1.01,
			"CostDetails": costdetails,
			"CostSource":  "",
			"Destination": "+4986517174963",
			"ExtraInfo":   "",
			"OrderID":     int64(123),
			"OriginHost":  "192.168.1.2",
			"OriginID":    "dsafdsaf",
			"Partial":     false,
			"RequestType": "*PreRated",
			"RunID":       utils.MetaDefault,
			"SetupTime":   time.Date(2013, 11, 7, 8, 42, 23, 0, time.UTC),
			"Source":      "UNIT_TEST",
			"Subject":     "1001",
			"Tenant":      "cgrates.org",
			"ToR":         "*voice",
			"Usage":       time.Duration(13) * time.Second,
			"field_extr1": "val_extr1",
			"fieldextr2":  "valextr2",
			"PreRated":    false,
		},
	}
	eCDR := &CDR{
		CGRID:       cdr.CGRID,
		OrderID:     123,
		ToR:         utils.VOICE,
		OriginID:    "dsafdsaf",
		OriginHost:  "192.168.1.2",
		Source:      utils.UNIT_TEST,
		RequestType: utils.META_RATED,
		Tenant:      "cgrates.org",
		Category:    "call",
		Account:     "1001",
		Subject:     "1001",
		Destination: "+4986517174963",
		SetupTime:   time.Date(2013, 11, 7, 8, 42, 23, 0, time.UTC),
		AnswerTime:  time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC),
		RunID:       utils.MetaDefault,
		Usage:       time.Duration(13) * time.Second,
		Cost:        1.01,
		ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
	}
	if err := cdr.UpdateFromCGREvent(cgrEvent, []string{"OriginHost", "SetupTime", "Usage"}); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(cdr, eCDR) {
		t.Errorf("Expecting: %+v, received: %+v", cdr, eCDR)
	}
}

func TestCDRAddDefaults(t *testing.T) {
	cdr := &CDR{
		OriginID:   "dsafdsaf",
		OriginHost: "192.168.1.2",
		Account:    "1001",
	}
	cfg, err := config.NewDefaultCGRConfig()
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	eCDR := &CDR{
		CGRID:       "bf736fb56ce586357ab2f286b777187a1612c6e6",
		ToR:         utils.VOICE,
		RunID:       utils.MetaDefault,
		Subject:     "1001",
		RequestType: utils.META_RATED,
		Tenant:      "cgrates.org",
		Category:    utils.CALL,
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
		ToR:         utils.VOICE,
		OriginID:    "dsafdsaf",
		OriginHost:  "192.168.1.1",
		Source:      utils.UNIT_TEST,
		RequestType: utils.META_RATED,
		Tenant:      "cgrates.org",
		Category:    "call",
		Account:     "1001",
		Subject:     "1001",
		Destination: "+4986517174963",
		SetupTime:   time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC),
		AnswerTime:  time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC),
		RunID:       utils.MetaDefault,
		Usage:       time.Duration(10) * time.Second,
		Cost:        1.01,
		ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
	}

	cfgCdrFld := &config.FCTemplate{Path: utils.PathItems{{Field: utils.MetaExp}, {Field: "SetupTime"}}, Type: utils.META_COMPOSED,
		PathSlice: []string{utils.MetaExp, "SetupTime"},
		Value:     config.NewRSRParsersMustCompile("~SetupTime", true, utils.INFIELD_SEP)}

	eVal := "2013-11-07T08:42:20Z"
	if val, err := cdr.exportFieldValue(cfgCdrFld, nil); err != nil {
		t.Error(err)
	} else if val != eVal {
		t.Errorf("Expecting: %+v, received: %+v", eVal, val)
	}

}
