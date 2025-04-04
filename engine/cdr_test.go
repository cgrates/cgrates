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
	"encoding/json"
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

func TestNewCDRFromExternalCDRErrors(t *testing.T) {
	extCdr := &ExternalCDR{
		CGRID:       utils.Sha1("dsafdsaf", time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC).String()),
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
		Destination: "1002",
		SetupTime:   "2013-11-07T08:42:20Z",
		AnswerTime:  "2013-11-07T08:42:26Z",
		RunID:       utils.MetaDefault,
		Usage:       "10",
		Cost:        1.01,
		PreRated:    true,
		ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
	}

	///
	extCdr.SetupTime = "invalid"
	errExpect := "Unsupported time format"
	if _, err := NewCDRFromExternalCDR(extCdr, ""); err == nil || err.Error() != errExpect {
		t.Errorf("Expected %v but received %v", errExpect, err)
	}
	extCdr.SetupTime = "2013-11-07T08:42:20Z"

	///
	extCdr.AnswerTime = "invalid"
	if _, err := NewCDRFromExternalCDR(extCdr, ""); err == nil || err.Error() != errExpect {
		t.Errorf("Expected %v but received %v", errExpect, err)
	}
	extCdr.AnswerTime = "2013-11-07T08:42:20Z"

	///

	extCdr.CGRID = ""
	if _, err := NewCDRFromExternalCDR(extCdr, ""); err != nil {
		t.Error(err)
	}
	extCdr.CGRID = utils.Sha1("dsafdsaf", time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC).String())

	///
	extCdr.Usage = "invalid"
	errExpect = `time: invalid duration "invalid"`
	if _, err := NewCDRFromExternalCDR(extCdr, ""); err == nil || err.Error() != errExpect {
		t.Error(err)
	}
	extCdr.Usage = "10"

	///
	extCdr.CostDetails = "1"
	errExpect = `json: cannot unmarshal number into Go value of type engine.EventCost`
	if _, err := NewCDRFromExternalCDR(extCdr, ""); err == nil || err.Error() != errExpect {
		t.Error(err)
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
				{
					UUID:  "f9be602747f4",
					ID:    "monetary",
					Type:  utils.MetaMonetary,
					Value: 0.5,
				},
				{
					UUID:  "2e02510ab90a",
					ID:    "voice",
					Type:  utils.MetaVoice,
					Value: 10,
				},
			},
		},
	}

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
		CostDetails: NewEventCostFromCallCost(cc, "TestCDRTestCDRAsMapStringIface2", utils.MetaDefault),
	}

	prsr := config.NewRSRParserMustCompile("~*req.CostDetails.CGRID")
	eFldVal := "TestCDRTestCDRAsMapStringIface2"
	if fldVal, err := cdr.FieldAsString(prsr); err != nil {
		t.Error(err)
	} else if fldVal != eFldVal {
		t.Errorf("field: <%v>, expected: <%v>, received: <%v>", prsr, eFldVal, fldVal)
	}

	prsr = config.NewRSRParserMustCompile("~*req.CostDetails.AccountSummary.ID")
	eFldVal = "AccountFromAccountSummary"
	if fldVal, err := cdr.FieldAsString(prsr); err != nil {
		t.Error(err)
	} else if fldVal != eFldVal {
		t.Errorf("field: <%v>, expected: <%v>, received: <%v>", prsr, eFldVal, fldVal)
	}

	prsr = config.NewRSRParserMustCompile("~*req.CostDetails.AccountSummary.BalanceSummaries[1].ID")
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
	storCdr := CDR{CGRID: utils.Sha1("dsafdsaf", time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC).String()), OrderID: 123, ToR: utils.MetaVoice, OriginID: "dsafdsaf",
		OriginHost: "192.168.1.1", Source: utils.UNIT_TEST, RequestType: utils.MetaRated,
		Tenant: "cgrates.org", Category: "call", Account: "1001", Subject: "1001", Destination: "1002",
		SetupTime: time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC), AnswerTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC), RunID: utils.MetaDefault,
		Usage: 10 * time.Second, Supplier: "SUPPL1",
		ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"}, Cost: 1.01,
	}
	cdrForm := storCdr.AsHttpForm()
	if cdrForm.Get(utils.ToR) != utils.MetaVoice {
		t.Errorf("Expected: %s, received: %s", utils.MetaVoice, cdrForm.Get(utils.ToR))
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
	if cdrForm.Get(utils.RequestType) != utils.MetaRated {
		t.Errorf("Expected: %s, received: %s", utils.MetaRated, cdrForm.Get(utils.RequestType))
	}
	if cdrForm.Get(utils.Tenant) != "cgrates.org" {
		t.Errorf("Expected: %s, received: %s", "cgrates.org", cdrForm.Get(utils.Tenant))
	}
	if cdrForm.Get(utils.Category) != "call" {
		t.Errorf("Expected: %s, received: %s", "call", cdrForm.Get(utils.Category))
	}
	if cdrForm.Get(utils.AccountField) != "1001" {
		t.Errorf("Expected: %s, received: %s", "1001", cdrForm.Get(utils.AccountField))
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
	if cdrForm.Get(utils.Route) != "SUPPL1" {
		t.Errorf("Expected: %s, received: %s", "1001", cdrForm.Get(utils.Route))
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
	expectOutCdr.ToR, storCdr.ToR, expectOutCdr.Usage = "other", "other", "10000000000"
	if cdrOut := storCdr.AsExternalCDR(); !reflect.DeepEqual(expectOutCdr, cdrOut) {
		t.Errorf("Expected: %+v, received: %+v", expectOutCdr, cdrOut)
	}

}

func TestUsageReqAsCD(t *testing.T) {
	req := &UsageRecord{ToR: utils.MetaVoice, RequestType: utils.MetaRated,
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
		TimeEnd:             time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC).Add(10),
		DenyNegativeAccount: true}
	if cd, err := req.AsCallDescriptor("", true); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCD, cd) {
		t.Errorf("Expected: %+v, received: %+v", eCD, cd)
	}

}
func TestUsageReqAsCDNil(t *testing.T) {
	req := &UsageRecord{
		ToR: utils.MetaVoice, RequestType: utils.MetaRated,
		Tenant: "cgrates.org", Category: "call",
		Account: "1001", Subject: "1001", Destination: "1002",
		Usage:     "test",
		SetupTime: "2013-11-07T08:42:26Z",
		ExtraFields: map[string]string{
			"ExtraField1": "extraVal",
		},
	}

	if _, err := req.AsCallDescriptor("/time.zone", true); err == nil {
		t.Error(err)
	} else if _, err = req.AsCallDescriptor("", true); err == nil {
		t.Error(err)
	}
	req.Usage = "10"
	if _, err := req.AsCallDescriptor("", true); err != nil {
		t.Error(err)
	}

}

func TestCdrClone(t *testing.T) {

	var cdr *CDR
	if val := cdr.Clone(); val != nil {
		t.Errorf("expected nil , received %v ", val)
	}
	cdr = &CDR{}
	eOut := &CDR{}
	if rcv := cdr.Clone(); !reflect.DeepEqual(rcv, eOut) {
		t.Errorf("Expecting: %+v, received: %+v", eOut, rcv)
	}
	cdr = &CDR{
		CGRID:     "CGRID_test",
		OrderID:   18,
		SetupTime: time.Date(2020, time.April, 18, 23, 0, 4, 0, time.UTC),
		Usage:     10,
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
		Usage:     10,
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

	mp := map[string]any{
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
		CostDetails: NewEventCostFromCallCost(cc, "TestCDRTestCDRAsMapStringIface2", utils.MetaDefault),
	}

	mp := map[string]any{
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
		utils.CostDetails:  cdr.CostDetails,
	}
	if cdrMp := cdr.AsMapStringIface(); !reflect.DeepEqual(mp, cdrMp) {
		t.Errorf("Expecting: %+v, received: %+v", mp, cdrMp)
	}
}

func TestCDRAsCDRsql(t *testing.T) {
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
	eCDR, err := cdr.AsCDRsql(&JSONMarshaler{})
	if err != nil {
		t.Error(err)
	}
	eCDRSql := &CDRsql{
		Cgrid:       cdr.CGRID,
		RunID:       cdr.RunID,
		OriginID:    "dsafdsaf",
		TOR:         utils.MetaVoice,
		Source:      utils.UnitTest,
		Tenant:      "cgrates.org",
		Category:    "call",
		Account:     "1001",
		Subject:     "1001",
		Destination: "+4986517174963",
		SetupTime:   time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC),
		AnswerTime:  utils.TimePointer(time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC)),
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
		AnswerTime:  utils.TimePointer(time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC)),
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

	if eCDR, err := NewCDRFromSQL(cdrSQL, &JSONMarshaler{}); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(cdr, eCDR) {
		t.Errorf("Expecting: %+v, received: %+v", cdr, eCDR)
	}
	cdrSQL.CostDetails = "val"
	if _, err = NewCDRFromSQL(cdrSQL, &JSONMarshaler{}); err == nil {
		t.Error(err)
	}
	cdrSQL.ExtraFields = "test"
	if _, err = NewCDRFromSQL(cdrSQL, &JSONMarshaler{}); err == nil {
		t.Error(err)
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
		Event: map[string]any{
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

func TestEngineCDRString(t *testing.T) {
	TestCDR := CDR{
		CGRID:       "grid1",
		RunID:       "runid123",
		OrderID:     12345,
		OriginHost:  "10.1.1.1",
		Source:      "testsrouce",
		OriginID:    "origin123",
		ToR:         "tortest",
		RequestType: "prepaid",
		Tenant:      "cgrates.org",
		Category:    "rating",
		Account:     "testaccount",
		Subject:     "user",
		Destination: "+1234567890",
		SetupTime:   time.Now(),
		AnswerTime:  time.Now().Add(time.Minute * 5),
		Usage:       time.Minute * 3,
		ExtraFields: map[string]string{"key1": "value1"},
		ExtraInfo:   "some additional information",
		Partial:     false,
		PreRated:    false,
		CostSource:  "rating_engine",
		Cost:        1.50,
	}
	want, _ := json.Marshal(TestCDR)
	got := TestCDR.String()
	if string(got) != string(want) {
		t.Errorf("Expected CDR.String() to return %s, got %s", want, got)
	}
}

func TestCompressedCDR(t *testing.T) {
	tmpCfg := config.CgrConfig()
	defer func() {
		config.SetCgrConfig(tmpCfg)
	}()
	config.CgrConfig().CdrsCfg().CompressStoredCost = true
	cdr := CDR{
		CGRID:       utils.Sha1("dsafdsaf", time.Date(2024, 03, 7, 8, 42, 26, 0, time.UTC).String()),
		OrderID:     123,
		ToR:         utils.MetaVoice,
		OriginID:    "dsafdsaf",
		OriginHost:  "192.168.1.1",
		Source:      "test",
		RequestType: utils.MetaRated,
		Tenant:      "cgrates.org",
		Category:    "call",
		Account:     "1001",
		Subject:     "1001",
		Destination: "1002",
		SetupTime:   time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC),
		AnswerTime:  time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC),
		RunID:       utils.MetaDefault,
		Usage:       10 * time.Second,
		ExtraInfo:   "extraInfo",
		Partial:     false,
		PreRated:    true,
		CostSource:  "cost source",
		CostDetails: &EventCost{
			CGRID:     "test1",
			RunID:     utils.MetaDefault,
			StartTime: time.Date(2017, 1, 9, 16, 18, 21, 0, time.UTC),
			Usage:     utils.DurationPointer(3 * time.Minute),
			Cost:      utils.Float64Pointer(2.3),
			Charges: []*ChargingInterval{
				{
					RatingID: "RatingID2",
					Increments: []*ChargingIncrement{
						{
							Usage:          2 * time.Minute,
							Cost:           2.0,
							AccountingID:   "a012888",
							CompressFactor: 1,
						},
						{
							Usage:          time.Second,
							Cost:           0.005,
							AccountingID:   "44d6c02",
							CompressFactor: 60,
						},
					},
					CompressFactor: 1,
				},
			},
			AccountSummary: &AccountSummary{
				Tenant: "cgrates.org",
				ID:     "1001",
				BalanceSummaries: []*BalanceSummary{
					{
						UUID:  "uuid1",
						Type:  utils.MetaMonetary,
						Value: 50,
					},
				},
				AllowNegative: false,
				Disabled:      false,
			},
			Rating: Rating{
				"c1a5ab9": &RatingUnit{
					ConnectFee:       0.1,
					RoundingMethod:   "*up",
					RoundingDecimals: 5,
					RatesID:          "ec1a177",
					RatingFiltersID:  "43e77dc",
				},
			},
			Accounting: Accounting{
				"a012888": &BalanceCharge{
					AccountID:   "cgrates.org:1001",
					BalanceUUID: "uuid1",
					Units:       120.7,
				},
				"44d6c02": &BalanceCharge{
					AccountID:   "cgrates.org:1001",
					BalanceUUID: "uuid1",
					Units:       120.7,
				},
			},
			Rates: ChargedRates{
				"ec1a177": RateGroups{
					&RGRate{
						GroupIntervalStart: 0,
						Value:              0.01,
						RateIncrement:      time.Minute,
						RateUnit:           time.Second},
				},
			},
		},
		Cost: 1.01,
	}

	cdrCompres, err := cdr.AsCompressedCostCDR(&JSONMarshaler{})
	if err != nil {
		t.Error(err)
	}
	cdr2, err := NewCDRFromCompressedCDR(cdrCompres, &JSONMarshaler{})
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(cdr.CostDetails, cdr2.CostDetails) {
		t.Errorf("expected %v,received %v", utils.ToJSON(cdr.CostDetails), utils.ToJSON(cdr2.CostDetails))
	}

	cdrsql, err := cdr.AsCDRsql(&JSONMarshaler{})
	if err != nil {
		t.Error(err)
	}
	cdr3, err := NewCDRFromSQL(cdrsql, &JSONMarshaler{})
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(cdr.CostDetails, cdr3.CostDetails) {
		t.Errorf("expected %v,received %v\n", utils.ToJSON(cdr.CostDetails), utils.ToJSON(cdr3.CostDetails))
	}

	idb := NewInternalDB(nil, nil, true, false, config.CgrConfig().DataDbCfg().Items)
	if err := idb.SetCDR(&cdr, true); err != nil {
		t.Error(err)
	}
	if cdrs, _, err := idb.GetCDRs(&utils.CDRsFilter{}, false); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(cdrs[0].CostDetails, cdr.CostDetails) {
		t.Errorf("expected %v,\nreceived %v\n", utils.ToJSON(cdr), utils.ToJSON(cdrs[0]))
	}
}
