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
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

func TestCDRInterfaces(t *testing.T) {
	CDR := new(CDR)
	var _ RawCdr = CDR
	var _ Event = CDR
}

func TestNewCDRFromExternalCDR(t *testing.T) {
	extCdr := &ExternalCDR{CGRID: utils.Sha1("dsafdsaf", time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC).String()), OrderID: 123, ToR: utils.VOICE,
		OriginID: "dsafdsaf", OriginHost: "192.168.1.1", Source: utils.UNIT_TEST, RequestType: utils.META_RATED, Direction: "*out",
		Tenant: "cgrates.org", Category: "call", Account: "1001", Subject: "1001", Destination: "1002", Supplier: "SUPPL1",
		SetupTime: "2013-11-07T08:42:20Z", AnswerTime: "2013-11-07T08:42:26Z", RunID: utils.DEFAULT_RUNID,
		Usage: "0.00000001", PDD: "7.0", ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"}, Cost: 1.01, Rated: true,
	}
	eStorCdr := &CDR{CGRID: utils.Sha1("dsafdsaf", time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC).String()), OrderID: 123, ToR: utils.VOICE,
		OriginID: "dsafdsaf", OriginHost: "192.168.1.1", Source: utils.UNIT_TEST, RequestType: utils.META_RATED, Direction: "*out",
		Tenant: "cgrates.org", Category: "call", Account: "1001", Subject: "1001", Destination: "1002", Supplier: "SUPPL1",
		SetupTime: time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC), AnswerTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC), RunID: utils.DEFAULT_RUNID,
		Usage: time.Duration(10), PDD: time.Duration(7) * time.Second, ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"}, Cost: 1.01, Rated: true,
	}
	if CDR, err := NewCDRFromExternalCDR(extCdr, ""); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eStorCdr, CDR) {
		t.Errorf("Expected: %+v, received: %+v", eStorCdr, CDR)
	}
}

func TestCDRClone(t *testing.T) {
	storCdr := &CDR{CGRID: utils.Sha1("dsafdsaf", time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC).String()), OrderID: 123, ToR: utils.VOICE,
		OriginID: "dsafdsaf", OriginHost: "192.168.1.1", Source: utils.UNIT_TEST, RequestType: utils.META_RATED, Direction: "*out",
		Tenant: "cgrates.org", Category: "call", Account: "1001", Subject: "1001", Destination: "1002", Supplier: "SUPPL1",
		SetupTime: time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC), AnswerTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC), RunID: utils.DEFAULT_RUNID,
		Usage: time.Duration(10), PDD: time.Duration(7) * time.Second, ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"}, Cost: 1.01, Rated: true,
	}
	if clnStorCdr := storCdr.Clone(); !reflect.DeepEqual(storCdr, clnStorCdr) {
		t.Errorf("Expecting: %+v, received: %+v", storCdr, clnStorCdr)
	}
}

func TestFieldAsString(t *testing.T) {
	cdr := CDR{CGRID: utils.Sha1("dsafdsaf", time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC).String()), OrderID: 123, ToR: utils.VOICE, OriginID: "dsafdsaf",
		OriginHost: "192.168.1.1", Source: "test", RequestType: utils.META_RATED, Direction: "*out", Tenant: "cgrates.org",
		Category: "call", Account: "1001", Subject: "1001", Destination: "1002", SetupTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC),
		AnswerTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC), RunID: utils.DEFAULT_RUNID,
		Usage: time.Duration(10) * time.Second, PDD: time.Duration(5) * time.Second, Supplier: "SUPPL1",
		ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
		Cost:        1.01,
	}
	if cdr.FieldAsString(&utils.RSRField{Id: utils.CGRID}) != cdr.CGRID ||
		cdr.FieldAsString(&utils.RSRField{Id: utils.ORDERID}) != "123" ||
		cdr.FieldAsString(&utils.RSRField{Id: utils.TOR}) != utils.VOICE ||
		cdr.FieldAsString(&utils.RSRField{Id: utils.ACCID}) != cdr.OriginID ||
		cdr.FieldAsString(&utils.RSRField{Id: utils.CDRHOST}) != cdr.OriginHost ||
		cdr.FieldAsString(&utils.RSRField{Id: utils.CDRSOURCE}) != cdr.Source ||
		cdr.FieldAsString(&utils.RSRField{Id: utils.REQTYPE}) != cdr.RequestType ||
		cdr.FieldAsString(&utils.RSRField{Id: utils.DIRECTION}) != cdr.Direction ||
		cdr.FieldAsString(&utils.RSRField{Id: utils.CATEGORY}) != cdr.Category ||
		cdr.FieldAsString(&utils.RSRField{Id: utils.ACCOUNT}) != cdr.Account ||
		cdr.FieldAsString(&utils.RSRField{Id: utils.SUBJECT}) != cdr.Subject ||
		cdr.FieldAsString(&utils.RSRField{Id: utils.DESTINATION}) != cdr.Destination ||
		cdr.FieldAsString(&utils.RSRField{Id: utils.SETUP_TIME}) != cdr.SetupTime.Format(time.RFC3339) ||
		cdr.FieldAsString(&utils.RSRField{Id: utils.ANSWER_TIME}) != cdr.AnswerTime.Format(time.RFC3339) ||
		cdr.FieldAsString(&utils.RSRField{Id: utils.USAGE}) != "10" ||
		cdr.FieldAsString(&utils.RSRField{Id: utils.PDD}) != "5" ||
		cdr.FieldAsString(&utils.RSRField{Id: utils.SUPPLIER}) != cdr.Supplier ||
		cdr.FieldAsString(&utils.RSRField{Id: utils.MEDI_RUNID}) != cdr.RunID ||
		cdr.FieldAsString(&utils.RSRField{Id: utils.COST}) != "1.01" ||
		cdr.FieldAsString(&utils.RSRField{Id: "field_extr1"}) != cdr.ExtraFields["field_extr1"] ||
		cdr.FieldAsString(&utils.RSRField{Id: "fieldextr2"}) != cdr.ExtraFields["fieldextr2"] ||
		cdr.FieldAsString(&utils.RSRField{Id: "dummy_field"}) != "" {
		t.Error("Unexpected filed value received",
			cdr.FieldAsString(&utils.RSRField{Id: utils.CGRID}) != cdr.CGRID,
			cdr.FieldAsString(&utils.RSRField{Id: utils.ORDERID}) != "123",
			cdr.FieldAsString(&utils.RSRField{Id: utils.TOR}) != utils.VOICE,
			cdr.FieldAsString(&utils.RSRField{Id: utils.ACCID}) != cdr.OriginID,
			cdr.FieldAsString(&utils.RSRField{Id: utils.CDRHOST}) != cdr.OriginHost,
			cdr.FieldAsString(&utils.RSRField{Id: utils.CDRSOURCE}) != cdr.Source,
			cdr.FieldAsString(&utils.RSRField{Id: utils.REQTYPE}) != cdr.RequestType,
			cdr.FieldAsString(&utils.RSRField{Id: utils.DIRECTION}) != cdr.Direction,
			cdr.FieldAsString(&utils.RSRField{Id: utils.CATEGORY}) != cdr.Category,
			cdr.FieldAsString(&utils.RSRField{Id: utils.ACCOUNT}) != cdr.Account,
			cdr.FieldAsString(&utils.RSRField{Id: utils.SUBJECT}) != cdr.Subject,
			cdr.FieldAsString(&utils.RSRField{Id: utils.DESTINATION}) != cdr.Destination,
			cdr.FieldAsString(&utils.RSRField{Id: utils.SETUP_TIME}) != cdr.SetupTime.Format(time.RFC3339),
			cdr.FieldAsString(&utils.RSRField{Id: utils.ANSWER_TIME}) != cdr.AnswerTime.Format(time.RFC3339),
			cdr.FieldAsString(&utils.RSRField{Id: utils.USAGE}) != "10",
			cdr.FieldAsString(&utils.RSRField{Id: utils.PDD}) != "5",
			cdr.FieldAsString(&utils.RSRField{Id: utils.SUPPLIER}) != cdr.Supplier,
			cdr.FieldAsString(&utils.RSRField{Id: utils.MEDI_RUNID}) != cdr.RunID,
			cdr.FieldAsString(&utils.RSRField{Id: utils.COST}) != "1.01",
			cdr.FieldAsString(&utils.RSRField{Id: "field_extr1"}) != cdr.ExtraFields["field_extr1"],
			cdr.FieldAsString(&utils.RSRField{Id: "fieldextr2"}) != cdr.ExtraFields["fieldextr2"],
			cdr.FieldAsString(&utils.RSRField{Id: "dummy_field"}) != "")
	}
}

func TestFieldsAsString(t *testing.T) {
	cdr := CDR{CGRID: utils.Sha1("dsafdsaf", time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC).String()), OrderID: 123, ToR: utils.VOICE, OriginID: "dsafdsaf",
		OriginHost: "192.168.1.1", Source: "test", RequestType: utils.META_RATED, Direction: "*out", Tenant: "cgrates.org",
		Category: "call", Account: "1001", Subject: "1001", Destination: "1002", SetupTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC),
		AnswerTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC), RunID: utils.DEFAULT_RUNID,
		Usage: time.Duration(10) * time.Second, PDD: time.Duration(5) * time.Second, Supplier: "SUPPL1",
		ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"}, Cost: 1.01,
	}
	eVal := "call_from_1001"
	if val := cdr.FieldsAsString(utils.ParseRSRFieldsMustCompile("Category;^_from_;Account", utils.INFIELD_SEP)); val != eVal {
		t.Errorf("Expecting : %s, received: %s", eVal, val)
	}
}

func TestUsageMultiply(t *testing.T) {
	cdr := CDR{Usage: time.Duration(10) * time.Second}
	if cdr.UsageMultiply(1024.0, 0); cdr.Usage != time.Duration(10240)*time.Second {
		t.Errorf("Unexpected usage after multiply: %v", cdr.Usage.Nanoseconds())
	}
	cdr = CDR{Usage: time.Duration(10240) * time.Second} // Simulate conversion back, gives out a bit odd result but this can be rounded on export
	expectDuration, _ := time.ParseDuration("10.000005120s")
	if cdr.UsageMultiply(0.000976563, 0); cdr.Usage != expectDuration {
		t.Errorf("Unexpected usage after multiply: %v", cdr.Usage.Nanoseconds())
	}
}

func TestCostMultiply(t *testing.T) {
	cdr := CDR{Cost: 1.01}
	if cdr.CostMultiply(1.19, 4); cdr.Cost != 1.2019 {
		t.Errorf("Unexpected cost after multiply: %v", cdr.Cost)
	}
	cdr = CDR{Cost: 1.01}
	if cdr.CostMultiply(1000, 0); cdr.Cost != 1010 {
		t.Errorf("Unexpected cost after multiply: %v", cdr.Cost)
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

func TestFormatUsage(t *testing.T) {
	cdr := CDR{Usage: time.Duration(10) * time.Second}
	if cdr.FormatUsage(utils.SECONDS) != "10" {
		t.Error("Wrong usage format: ", cdr.FormatUsage(utils.SECONDS))
	}
	if cdr.FormatUsage("default") != "10" {
		t.Error("Wrong usage format: ", cdr.FormatUsage("default"))
	}
	cdr = CDR{ToR: utils.DATA, Usage: time.Duration(1640113000000000)}
	if cdr.FormatUsage("default") != "1640113" {
		t.Error("Wrong usage format: ", cdr.FormatUsage("default"))
	}
	cdr = CDR{Usage: time.Duration(2) * time.Millisecond}
	if cdr.FormatUsage("default") != "0.002" {
		t.Error("Wrong usage format: ", cdr.FormatUsage("default"))
	}
	cdr = CDR{Usage: time.Duration(1002) * time.Millisecond}
	if cdr.FormatUsage("default") != "1.002" {
		t.Error("Wrong usage format: ", cdr.FormatUsage("default"))
	}
}

/*
func TestCDRAsHttpForm(t *testing.T) {
	storCdr := CDR{CGRID: utils.Sha1("dsafdsaf", time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC).String()), OrderID: 123, ToR: utils.VOICE, OriginID: "dsafdsaf",
		OriginHost: "192.168.1.1", Source: utils.UNIT_TEST, RequestType: utils.META_RATED, Direction: "*out",
		Tenant: "cgrates.org", Category: "call", Account: "1001", Subject: "1001", Destination: "1002",
		SetupTime: time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC), AnswerTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC), RunID: utils.DEFAULT_RUNID,
		Usage: time.Duration(10) * time.Second, Supplier: "SUPPL1",
		ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"}, Cost: 1.01,
	}
	cdrForm := storCdr.AsHttpForm()
	if cdrForm.Get(utils.TOR) != utils.VOICE {
		t.Errorf("Expected: %s, received: %s", utils.VOICE, cdrForm.Get(utils.TOR))
	}
	if cdrForm.Get(utils.ACCID) != "dsafdsaf" {
		t.Errorf("Expected: %s, received: %s", "dsafdsaf", cdrForm.Get(utils.ACCID))
	}
	if cdrForm.Get(utils.CDRHOST) != "192.168.1.1" {
		t.Errorf("Expected: %s, received: %s", "192.168.1.1", cdrForm.Get(utils.CDRHOST))
	}
	if cdrForm.Get(utils.CDRSOURCE) != utils.UNIT_TEST {
		t.Errorf("Expected: %s, received: %s", utils.UNIT_TEST, cdrForm.Get(utils.CDRSOURCE))
	}
	if cdrForm.Get(utils.REQTYPE) != utils.META_RATED {
		t.Errorf("Expected: %s, received: %s", utils.META_RATED, cdrForm.Get(utils.REQTYPE))
	}
	if cdrForm.Get(utils.DIRECTION) != "*out" {
		t.Errorf("Expected: %s, received: %s", "*out", cdrForm.Get(utils.DIRECTION))
	}
	if cdrForm.Get(utils.TENANT) != "cgrates.org" {
		t.Errorf("Expected: %s, received: %s", "cgrates.org", cdrForm.Get(utils.TENANT))
	}
	if cdrForm.Get(utils.CATEGORY) != "call" {
		t.Errorf("Expected: %s, received: %s", "call", cdrForm.Get(utils.CATEGORY))
	}
	if cdrForm.Get(utils.ACCOUNT) != "1001" {
		t.Errorf("Expected: %s, received: %s", "1001", cdrForm.Get(utils.ACCOUNT))
	}
	if cdrForm.Get(utils.SUBJECT) != "1001" {
		t.Errorf("Expected: %s, received: %s", "1001", cdrForm.Get(utils.SUBJECT))
	}
	if cdrForm.Get(utils.DESTINATION) != "1002" {
		t.Errorf("Expected: %s, received: %s", "1002", cdrForm.Get(utils.DESTINATION))
	}
	if cdrForm.Get(utils.SETUP_TIME) != "2013-11-07T08:42:20Z" {
		t.Errorf("Expected: %s, received: %s", "2013-11-07T08:42:20Z", cdrForm.Get(utils.SETUP_TIME))
	}
	if cdrForm.Get(utils.ANSWER_TIME) != "2013-11-07T08:42:26Z" {
		t.Errorf("Expected: %s, received: %s", "2013-11-07T08:42:26Z", cdrForm.Get(utils.ANSWER_TIME))
	}
	if cdrForm.Get(utils.USAGE) != "10" {
		t.Errorf("Expected: %s, received: %s", "10", cdrForm.Get(utils.USAGE))
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

func TestCDRForkCdr(t *testing.T) {
	storCdr := CDR{CGRID: utils.Sha1("dsafdsaf", time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC).String()), OrderID: 123, ToR: utils.VOICE,
		OriginID: "dsafdsaf", OriginHost: "192.168.1.1", Source: utils.UNIT_TEST, RequestType: utils.META_RATED, Direction: "*out",
		Tenant: "cgrates.org", Category: "call", Account: "1001", Subject: "1001", Destination: "1002",
		SetupTime: time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC), PDD: time.Duration(200) * time.Millisecond,
		AnswerTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC), RunID: utils.DEFAULT_RUNID,
		Usage: time.Duration(10) * time.Second, Supplier: "suppl1",
		ExtraFields: map[string]string{"field_extr1": "val_extr1", "field_extr2": "valextr2"}, Cost: 1.01}
	rtSampleCdrOut, err := storCdr.ForkCdr("sample_run1", &utils.RSRField{Id: utils.REQTYPE}, &utils.RSRField{Id: utils.DIRECTION}, &utils.RSRField{Id: utils.TENANT},
		&utils.RSRField{Id: utils.CATEGORY}, &utils.RSRField{Id: utils.ACCOUNT}, &utils.RSRField{Id: utils.SUBJECT}, &utils.RSRField{Id: utils.DESTINATION},
		&utils.RSRField{Id: utils.SETUP_TIME}, &utils.RSRField{Id: utils.PDD}, &utils.RSRField{Id: utils.ANSWER_TIME}, &utils.RSRField{Id: utils.USAGE},
		&utils.RSRField{Id: utils.SUPPLIER}, &utils.RSRField{Id: utils.DISCONNECT_CAUSE}, &utils.RSRField{Id: utils.RATED_FLD}, &utils.RSRField{Id: utils.COST},
		[]*utils.RSRField{&utils.RSRField{Id: "field_extr1"}, &utils.RSRField{Id: "field_extr2"}}, true, "")
	if err != nil {
		t.Error("Unexpected error received", err)
	}
	expctSplRatedCdr := &CDR{CGRID: storCdr.CGRID, ToR: utils.VOICE, OriginID: "dsafdsaf", OriginHost: "192.168.1.1", Source: utils.UNIT_TEST, RequestType: utils.META_RATED,
		Direction: "*out", Tenant: "cgrates.org", Category: "call", Account: "1001", Subject: "1001", Destination: "1002",
		SetupTime: time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC), PDD: time.Duration(200) * time.Millisecond, AnswerTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC),
		Usage: time.Duration(10) * time.Second, Supplier: "suppl1", ExtraFields: map[string]string{"field_extr1": "val_extr1", "field_extr2": "valextr2"},
		RunID: "sample_run1", Rated: false, Cost: 1.01}
	if !reflect.DeepEqual(expctSplRatedCdr, rtSampleCdrOut) {
		t.Errorf("Expected: %v, received: %v", expctSplRatedCdr, rtSampleCdrOut)
	}
}

func TestCDRForkCdrStaticVals(t *testing.T) {
	storCdr := CDR{CGRID: utils.Sha1("dsafdsaf", time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC).String()), OrderID: 123, ToR: utils.VOICE, OriginID: "dsafdsaf",
		OriginHost: "192.168.1.1", Source: utils.UNIT_TEST, RequestType: utils.META_RATED, Direction: "*out",
		Tenant: "cgrates.org", Category: "call", Account: "1001", Subject: "1001", Destination: "1002",
		SetupTime: time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC), AnswerTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC), RunID: utils.DEFAULT_RUNID,
		Usage: time.Duration(10) * time.Second, ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"}, Cost: 1.01,
	}
	rsrStPostpaid, _ := utils.NewRSRField("^" + utils.META_POSTPAID)
	rsrStIn, _ := utils.NewRSRField("^*in")
	rsrStCgr, _ := utils.NewRSRField("^cgrates.com")
	rsrStPC, _ := utils.NewRSRField("^premium_call")
	rsrStFA, _ := utils.NewRSRField("^first_account")
	rsrStFS, _ := utils.NewRSRField("^first_subject")
	rsrStST, _ := utils.NewRSRField("^2013-12-07T08:42:24Z")
	rsrStAT, _ := utils.NewRSRField("^2013-12-07T08:42:26Z")
	rsrStDur, _ := utils.NewRSRField("^12s")
	rsrStSuppl, _ := utils.NewRSRField("^supplier1")
	rsrStDCause, _ := utils.NewRSRField("^HANGUP_COMPLETE")
	rsrPDD, _ := utils.NewRSRField("^3")
	rsrStRated, _ := utils.NewRSRField("^true")
	rsrStCost, _ := utils.NewRSRField("^1.2")
	rtCdrOut2, err := storCdr.ForkCdr("wholesale_run", rsrStPostpaid, rsrStIn, rsrStCgr, rsrStPC, rsrStFA, rsrStFS, &utils.RSRField{Id: utils.DESTINATION},
		rsrStST, rsrPDD, rsrStAT, rsrStDur, rsrStSuppl, rsrStDCause, rsrStRated, rsrStCost, []*utils.RSRField{}, true, "")
	if err != nil {
		t.Error("Unexpected error received", err)
	}
	expctRatedCdr2 := &CDR{CGRID: storCdr.CGRID, ToR: utils.VOICE, OriginID: "dsafdsaf", OriginHost: "192.168.1.1", Source: utils.UNIT_TEST, RequestType: utils.META_POSTPAID,
		Direction: "*in", Tenant: "cgrates.com", Category: "premium_call", Account: "first_account", Subject: "first_subject", Destination: "1002",
		SetupTime:  time.Date(2013, 12, 7, 8, 42, 24, 0, time.UTC),
		AnswerTime: time.Date(2013, 12, 7, 8, 42, 26, 0, time.UTC), Usage: time.Duration(12) * time.Second, PDD: time.Duration(3) * time.Second,
		Supplier: "supplier1", DisconnectCause: "HANGUP_COMPLETE", Rated: true, Cost: 1.2,
		ExtraFields: map[string]string{}, RunID: "wholesale_run"}
	if !reflect.DeepEqual(rtCdrOut2, expctRatedCdr2) {
		t.Errorf("Received: %v, expected: %v", rtCdrOut2, expctRatedCdr2)
	}
	_, err = storCdr.ForkCdr("wholesale_run", &utils.RSRField{Id: "dummy_header"}, &utils.RSRField{Id: utils.DIRECTION}, &utils.RSRField{Id: utils.TENANT},
		&utils.RSRField{Id: utils.TOR}, &utils.RSRField{Id: utils.ACCOUNT}, &utils.RSRField{Id: utils.SUBJECT}, &utils.RSRField{Id: utils.DESTINATION},
		&utils.RSRField{Id: utils.SETUP_TIME}, &utils.RSRField{Id: utils.PDD}, &utils.RSRField{Id: utils.ANSWER_TIME}, &utils.RSRField{Id: utils.USAGE},
		&utils.RSRField{Id: utils.SUPPLIER},
		&utils.RSRField{Id: utils.DISCONNECT_CAUSE}, &utils.RSRField{Id: utils.RATED_FLD}, &utils.RSRField{Id: utils.COST}, []*utils.RSRField{}, true, "")
	if err == nil {
		t.Error("Failed to detect missing header")
	}
}

func TestCDRForkCdrFromMetaDefaults(t *testing.T) {
	storCdr := CDR{CGRID: utils.Sha1("dsafdsaf", time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC).String()), OrderID: 123, ToR: utils.VOICE, OriginID: "dsafdsaf",
		OriginHost: "192.168.1.1", Source: utils.UNIT_TEST, RequestType: utils.META_RATED, Direction: "*out", Tenant: "cgrates.org",
		Category: "call", Account: "1001", Subject: "1001", Destination: "1002",
		SetupTime: time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC), AnswerTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC), RunID: utils.DEFAULT_RUNID,
		Usage: time.Duration(10) * time.Second, PDD: time.Duration(4) * time.Second, Supplier: "SUPPL3",
		ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"}, Cost: 1.01,
	}
	expctCdr := &CDR{CGRID: storCdr.CGRID, ToR: utils.VOICE, OriginID: "dsafdsaf", OriginHost: "192.168.1.1", Source: utils.UNIT_TEST, RequestType: utils.META_RATED,
		Direction: "*out", Tenant: "cgrates.org", Category: "call", Account: "1001", Subject: "1001", Destination: "1002",
		SetupTime: time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC), AnswerTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC),
		Usage: time.Duration(10) * time.Second, PDD: time.Duration(4) * time.Second, Supplier: "SUPPL3", Cost: 1.01,
		ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"}, RunID: "wholesale_run"}
	cdrOut, err := storCdr.ForkCdr("wholesale_run", &utils.RSRField{Id: utils.META_DEFAULT}, &utils.RSRField{Id: utils.META_DEFAULT}, &utils.RSRField{Id: utils.META_DEFAULT},
		&utils.RSRField{Id: utils.META_DEFAULT}, &utils.RSRField{Id: utils.META_DEFAULT}, &utils.RSRField{Id: utils.META_DEFAULT}, &utils.RSRField{Id: utils.META_DEFAULT},
		&utils.RSRField{Id: utils.META_DEFAULT}, &utils.RSRField{Id: utils.META_DEFAULT}, &utils.RSRField{Id: utils.META_DEFAULT}, &utils.RSRField{Id: utils.META_DEFAULT},
		&utils.RSRField{Id: utils.META_DEFAULT}, &utils.RSRField{Id: utils.META_DEFAULT}, &utils.RSRField{Id: utils.META_DEFAULT}, &utils.RSRField{Id: utils.META_DEFAULT},
		[]*utils.RSRField{&utils.RSRField{Id: "field_extr1"}, &utils.RSRField{Id: "fieldextr2"}}, true, "")
	if err != nil {
		t.Fatal("Unexpected error received", err)
	}

	if !reflect.DeepEqual(expctCdr, cdrOut) {
		t.Errorf("Expected: %v, received: %v", expctCdr, cdrOut)
	}
	// Should also accept nil as defaults
	if cdrOut, err := storCdr.ForkCdr("wholesale_run", nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
		[]*utils.RSRField{&utils.RSRField{Id: "field_extr1"}, &utils.RSRField{Id: "fieldextr2"}}, true, ""); err != nil {
		t.Fatal("Unexpected error received", err)
	} else if !reflect.DeepEqual(expctCdr, cdrOut) {
		t.Errorf("Expected: %v, received: %v", expctCdr, cdrOut)
	}
}

func TestCDRAsExternalCDR(t *testing.T) {
	storCdr := CDR{CGRID: utils.Sha1("dsafdsaf", time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC).String()), OrderID: 123, ToR: utils.VOICE,
		OriginID: "dsafdsaf", OriginHost: "192.168.1.1", Source: utils.UNIT_TEST, RequestType: utils.META_RATED, Direction: "*out",
		Tenant: "cgrates.org", Category: "call", Account: "1001", Subject: "1001", Destination: "1002",
		SetupTime: time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC), AnswerTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC), RunID: utils.DEFAULT_RUNID,
		Usage: time.Duration(10), PDD: time.Duration(7) * time.Second, Supplier: "SUPPL1",
		ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"}, Cost: 1.01}
	expectOutCdr := &ExternalCDR{CGRID: utils.Sha1("dsafdsaf", time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC).String()), OrderID: 123, ToR: utils.VOICE,
		OriginID: "dsafdsaf", OriginHost: "192.168.1.1", Source: utils.UNIT_TEST, RequestType: utils.META_RATED, Direction: "*out",
		Tenant: "cgrates.org", Category: "call", Account: "1001", Subject: "1001", Destination: "1002",
		SetupTime: "2013-11-07T08:42:20Z", AnswerTime: "2013-11-07T08:42:26Z", RunID: utils.DEFAULT_RUNID,
		Usage: "0.00000001", PDD: "7", Supplier: "SUPPL1",
		ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"}, Cost: 1.01, CostDetails: "null"}
	if cdrOut := storCdr.AsExternalCDR(); !reflect.DeepEqual(expectOutCdr, cdrOut) {
		t.Errorf("Expected: %+v, received: %+v", expectOutCdr, cdrOut)
	}
}

func TestCDREventFields(t *testing.T) {
	cdr := &CDR{CGRID: utils.Sha1("dsafdsaf", time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC).String()), OrderID: 123, ToR: utils.VOICE, OriginID: "dsafdsaf",
		OriginHost: "192.168.1.1", Source: "test", RequestType: utils.META_RATED, Direction: "*out", Tenant: "cgrates.org", Category: "call", Account: "dan", Subject: "dans",
		Destination: "1002", SetupTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC), AnswerTime: time.Date(2013, 11, 7, 8, 42, 27, 0, time.UTC),
		RunID: utils.DEFAULT_RUNID, Usage: time.Duration(10) * time.Second, Supplier: "suppl1",
		ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"}, Cost: 1.01}

	if ev := cdr.AsEvent(""); ev != Event(cdr) {
		t.Error("Received: ", ev)
	}
	if res := cdr.GetName(); res != "test" {
		t.Error("Received: ", res)
	}
	if res := cdr.GetCgrId(""); res != utils.Sha1("dsafdsaf", time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC).String()) {
		t.Error("Received: ", res)
	}
	if res := cdr.GetUUID(); res != "dsafdsaf" {
		t.Error("Received: ", res)
	}
	if res := cdr.GetDirection(utils.META_DEFAULT); res != "*out" {
		t.Error("Received: ", res)
	}
	if res := cdr.GetSubject(utils.META_DEFAULT); res != "dans" {
		t.Error("Received: ", res)
	}
	if res := cdr.GetAccount(utils.META_DEFAULT); res != "dan" {
		t.Error("Received: ", res)
	}
	if res := cdr.GetDestination(utils.META_DEFAULT); res != "1002" {
		t.Error("Received: ", res)
	}
	if res := cdr.GetCallDestNr(utils.META_DEFAULT); res != "1002" {
		t.Error("Received: ", res)
	}
	if res := cdr.GetCategory(utils.META_DEFAULT); res != "call" {
		t.Error("Received: ", res)
	}
	if res := cdr.GetTenant(utils.META_DEFAULT); res != "cgrates.org" {
		t.Error("Received: ", res)
	}
	if res := cdr.GetReqType(utils.META_DEFAULT); res != utils.META_RATED {
		t.Error("Received: ", res)
	}
	if st, _ := cdr.GetSetupTime(utils.META_DEFAULT, ""); st != time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC) {
		t.Error("Received: ", st)
	}
	if at, _ := cdr.GetAnswerTime(utils.META_DEFAULT, ""); at != time.Date(2013, 11, 7, 8, 42, 27, 0, time.UTC) {
		t.Error("Received: ", at)
	}
	if et, _ := cdr.GetEndTime(utils.META_DEFAULT, ""); et != time.Date(2013, 11, 7, 8, 42, 37, 0, time.UTC) {
		t.Error("Received: ", et)
	}
	if dur, _ := cdr.GetDuration(utils.META_DEFAULT); dur != cdr.Usage {
		t.Error("Received: ", dur)
	}
	if suppl := cdr.GetSupplier(utils.META_DEFAULT); suppl != cdr.Supplier {
		t.Error("Received: ", suppl)
	}
	if res := cdr.GetOriginatorIP(utils.META_DEFAULT); res != cdr.OriginHost {
		t.Error("Received: ", res)
	}
	if extraFlds := cdr.GetExtraFields(); !reflect.DeepEqual(cdr.ExtraFields, extraFlds) {
		t.Error("Received: ", extraFlds)
	}
}

func TesUsageReqAsCDR(t *testing.T) {
	setupReq := &UsageRecord{ToR: utils.VOICE, RequestType: utils.META_RATED, Direction: "*out", Tenant: "cgrates.org", Category: "call",
		Account: "1001", Subject: "1001", Destination: "1002",
		SetupTime: "2013-11-07T08:42:20Z", AnswerTime: "2013-11-07T08:42:26Z", Usage: "0.00000001",
	}
	eStorCdr := &CDR{ToR: utils.VOICE, RequestType: utils.META_RATED, Direction: "*out",
		Tenant: "cgrates.org", Category: "call", Account: "1001", Subject: "1001", Destination: "1002",
		SetupTime: time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC), AnswerTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC), Usage: time.Duration(10)}
	if CDR, err := setupReq.AsCDR(""); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eStorCdr, CDR) {
		t.Errorf("Expected: %+v, received: %+v", eStorCdr, CDR)
	}
}

func TestUsageReqAsCD(t *testing.T) {
	req := &UsageRecord{ToR: utils.VOICE, RequestType: utils.META_RATED, Direction: "*out", Tenant: "cgrates.org", Category: "call",
		Account: "1001", Subject: "1001", Destination: "1002",
		SetupTime: "2013-11-07T08:42:20Z", AnswerTime: "2013-11-07T08:42:26Z", Usage: "0.00000001",
	}
	eCD := &CallDescriptor{CgrID: "9473e7b2e075d168b9da10ae957ee68fe5a217e4", TOR: req.ToR, Direction: req.Direction, Tenant: req.Tenant,
		Category: req.Category, Account: req.Account, Subject: req.Subject, Destination: req.Destination,
		TimeStart: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC), TimeEnd: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC).Add(time.Duration(10)), DenyNegativeAccount: true}
	if cd, err := req.AsCallDescriptor("", true); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCD, cd) {
		t.Errorf("Expected: %+v, received: %+v", eCD, cd)
	}
}

func TestCDRParseFieldValue(t *testing.T) {
	cdr := new(CDR)
	if err := cdr.ParseFieldValue(utils.PartialField, "true", ""); err != nil {
		t.Error(err)
	} else if !cdr.Partial {
		t.Errorf("Received cdr: %+v", cdr)
	}
	if err := cdr.ParseFieldValue(utils.ORDERID, "5", ""); err != nil {
		t.Error(err)
	} else if cdr.OrderID != 5 {
		t.Errorf("Received cdr: %+v", cdr)
	}
	if err := cdr.ParseFieldValue(utils.MEDI_RUNID, "*default", ""); err != nil {
		t.Error(err)
	} else if cdr.RunID != "*default" {
		t.Errorf("Received cdr: %+v", cdr)
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
		Direction:   "*out",
		Tenant:      "cgrates.org",
		Category:    "call",
		Account:     "1002",
		Subject:     "1001",
		Destination: "+4986517174963",
		SetupTime:   time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC),
		AnswerTime:  time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC),
		RunID:       utils.DEFAULT_RUNID,
		Usage:       time.Duration(10) * time.Second,
		Supplier:    "SUPPL1",
		ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
		Cost:        1.01,
	}

	mp := map[string]interface{}{
		"field_extr1":          "val_extr1",
		"fieldextr2":           "valextr2",
		utils.CGRID:            cdr.CGRID,
		utils.MEDI_RUNID:       utils.DEFAULT_RUNID,
		utils.ORDERID:          cdr.OrderID,
		utils.CDRHOST:          "192.168.1.1",
		utils.CDRSOURCE:        utils.UNIT_TEST,
		utils.ACCID:            "dsafdsaf",
		utils.TOR:              utils.VOICE,
		utils.REQTYPE:          utils.META_RATED,
		utils.DIRECTION:        "*out",
		utils.TENANT:           "cgrates.org",
		utils.CATEGORY:         "call",
		utils.ACCOUNT:          "1002",
		utils.SUBJECT:          "1001",
		utils.DESTINATION:      "+4986517174963",
		utils.SETUP_TIME:       time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC),
		utils.PDD:              time.Duration(0) * time.Second,
		utils.ANSWER_TIME:      time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC),
		utils.USAGE:            time.Duration(10) * time.Second,
		utils.SUPPLIER:         "SUPPL1",
		utils.DISCONNECT_CAUSE: cdr.DisconnectCause,
		utils.CostSource:       cdr.CostSource,
		utils.COST:             1.01,
		utils.COST_DETAILS:     cdr.CostDetails,
		utils.RATED:            false,
		utils.PartialField:     false,
		utils.ExtraInfo:        cdr.ExtraInfo,
	}
	if cdrMp, err := cdr.AsMapStringIface(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(mp, cdrMp) {
		t.Errorf("Expecting: %+v, received: %+v", mp, cdrMp)
	}

}

func TestCDRAsExportRecord(t *testing.T) {
	cdr := &CDR{
		CGRID: utils.Sha1("dsafdsaf",
			time.Unix(1383813745, 0).UTC().String()),
		ToR: utils.VOICE, OriginID: "dsafdsaf",
		OriginHost:  "192.168.1.1",
		RequestType: utils.META_RATED,
		Direction:   "*out", Tenant: "cgrates.org",
		Category:    "call",
		Account:     "1001",
		Subject:     "1001",
		Destination: "+4986517174963",
		SetupTime:   time.Unix(1383813745, 0).UTC(),
		AnswerTime:  time.Unix(1383813746, 0).UTC(),
		Usage:       time.Duration(10) * time.Second,
		RunID:       utils.DEFAULT_RUNID, Cost: 1.01,
		ExtraFields: map[string]string{"stop_time": "2014-06-11 19:19:00 +0000 UTC", "fieldextr2": "valextr2"}}

	val, _ := utils.ParseRSRFields(utils.DESTINATION, utils.INFIELD_SEP)
	cfgCdrFld := &config.CfgCdrField{Tag: "destination", Type: utils.META_COMPOSED, FieldId: utils.DESTINATION, Value: val, Timezone: "UTC"}
	if expRecord, err := cdr.AsExportRecord([]*config.CfgCdrField{cfgCdrFld}, false, nil, 0); err != nil {
		t.Error(err)
	} else if expRecord[0] != cdr.Destination {
		t.Errorf("Expecting:\n%s\nReceived:\n%s", cdr.Destination, expRecord[0])
	}
	if err := dm.DataDB().SetReverseDestination(&Destination{Id: "MASKED_DESTINATIONS", Prefixes: []string{"+4986517174963"}},
		utils.NonTransactional); err != nil {
		t.Error(err)
	}
	cfgCdrFld = &config.CfgCdrField{Tag: "destination", Type: utils.META_COMPOSED, FieldId: utils.DESTINATION, Value: val, MaskDestID: "MASKED_DESTINATIONS", MaskLen: 3}
	eDst := "+4986517174***"
	if expRecord, err := cdr.AsExportRecord([]*config.CfgCdrField{cfgCdrFld}, false, nil, 0); err != nil {
		t.Error(err)
	} else if expRecord[0] != eDst {
		t.Errorf("Expecting:\n%s\nReceived:\n%s", eDst, expRecord[0])
	}
	cfgCdrFld = &config.CfgCdrField{Tag: "MaskedDest", Type: utils.MetaMaskedDestination, Value: val, MaskDestID: "MASKED_DESTINATIONS"}
	if expRecord, err := cdr.AsExportRecord([]*config.CfgCdrField{cfgCdrFld}, false, nil, 0); err != nil {
		t.Error(err)
	} else if expRecord[0] != "1" {
		t.Errorf("Expecting:\n%s\nReceived:\n%s", "1", expRecord[0])
	}
	fltr, _ := utils.ParseRSRFields("Tenant(itsyscom.com)", utils.INFIELD_SEP)
	cfgCdrFld = &config.CfgCdrField{Tag: "destination", Type: utils.META_COMPOSED, FieldId: utils.DESTINATION, Value: val, FieldFilter: fltr, Timezone: "UTC"}
	if _, err := cdr.AsExportRecord([]*config.CfgCdrField{cfgCdrFld}, false, nil, 0); err == nil {
		t.Error("Failed to use filter")
	}
	// Test MetaDateTime
	val, _ = utils.ParseRSRFields("stop_time", utils.INFIELD_SEP)
	layout := "2006-01-02 15:04:05"
	cfgCdrFld = &config.CfgCdrField{Tag: "stop_time", Type: utils.MetaDateTime, FieldId: "stop_time", Value: val, Layout: layout, Timezone: "UTC"}
	if expRecord, err := cdr.AsExportRecord([]*config.CfgCdrField{cfgCdrFld}, false, nil, 0); err != nil {
		t.Error(err)
	} else if expRecord[0] != "2014-06-11 19:19:00" {
		t.Error("Expecting: 2014-06-11 19:19:00, got: ", expRecord[0])
	}
	// Test filter
	fltr, _ = utils.ParseRSRFields("Tenant(itsyscom.com)", utils.INFIELD_SEP)
	cfgCdrFld = &config.CfgCdrField{Tag: "stop_time", Type: utils.MetaDateTime, FieldId: "stop_time", Value: val, FieldFilter: fltr, Layout: layout, Timezone: "UTC"}
	if _, err := cdr.AsExportRecord([]*config.CfgCdrField{cfgCdrFld}, false, nil, 0); err == nil {
		t.Error("Received empty error", err)
	}
	val, _ = utils.ParseRSRFields("fieldextr2", utils.INFIELD_SEP)
	cfgCdrFld = &config.CfgCdrField{Tag: "stop_time", Type: utils.MetaDateTime, FieldId: "stop_time", Value: val, Layout: layout, Timezone: "UTC"}
	// Test time parse error
	if _, err := cdr.AsExportRecord([]*config.CfgCdrField{cfgCdrFld}, false, nil, 0); err == nil {
		t.Error("Should give error here, got none.")
	}
}

func TestCDRAsExportMap(t *testing.T) {
	cdr := &CDR{CGRID: utils.Sha1("dsafdsaf", time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC).String()), OrderID: 123, ToR: utils.VOICE, OriginID: "dsafdsaf",
		OriginHost: "192.168.1.1", Source: utils.UNIT_TEST, RequestType: utils.META_RATED, Direction: "*out",
		Tenant: "cgrates.org", Category: "call", Account: "1001", Subject: "1001", Destination: "+4986517174963",
		SetupTime: time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC), AnswerTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC), RunID: utils.DEFAULT_RUNID,
		Usage: time.Duration(10) * time.Second, Supplier: "SUPPL1",
		ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"}, Cost: 1.01,
	}
	eCDRMp := map[string]string{
		utils.CGRID:       cdr.CGRID,
		utils.DESTINATION: "004986517174963",
		"FieldExtra1":     "val_extr1",
	}
	expFlds := []*config.CfgCdrField{
		&config.CfgCdrField{FieldId: utils.CGRID, Type: utils.META_COMPOSED, Value: utils.ParseRSRFieldsMustCompile(utils.CGRID, utils.INFIELD_SEP)},
		&config.CfgCdrField{FieldId: utils.DESTINATION, Type: utils.META_COMPOSED, Value: utils.ParseRSRFieldsMustCompile("~Destination:s/^\\+(\\d+)$/00${1}/", utils.INFIELD_SEP)},
		&config.CfgCdrField{FieldId: "FieldExtra1", Type: utils.META_COMPOSED, Value: utils.ParseRSRFieldsMustCompile("field_extr1", utils.INFIELD_SEP)},
	}
	if cdrMp, err := cdr.AsExportMap(expFlds, false, nil, 0); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCDRMp, cdrMp) {
		t.Errorf("Expecting: %+v, received: %+v", eCDRMp, cdrMp)
	}
}
