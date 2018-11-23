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
		SetupTime: "2013-11-07T08:42:20Z", AnswerTime: "2013-11-07T08:42:26Z", RunID: utils.DEFAULT_RUNID,
		Usage: "10", Cost: 1.01, PreRated: true,
		ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
	}
	eStorCdr := &CDR{CGRID: utils.Sha1("dsafdsaf", time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC).String()),
		OrderID: 123, ToR: utils.VOICE, OriginID: "dsafdsaf", OriginHost: "192.168.1.1",
		Source: utils.UNIT_TEST, RequestType: utils.META_RATED, RunID: utils.DEFAULT_RUNID,
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
		RunID:      utils.DEFAULT_RUNID, Usage: time.Duration(10),
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
		AnswerTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC), RunID: utils.DEFAULT_RUNID,
		Usage: time.Duration(10) * time.Second, Cost: 1.01,
		ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
	}
	prsr := config.NewRSRParserMustCompile(utils.DynamicDataPrefix+utils.CGRID, true)
	eFldVal := cdr.CGRID
	if fldVal, err := cdr.FieldAsString(prsr); err != nil {
		t.Error(err)
	} else if fldVal != eFldVal {
		t.Errorf("field: <%v>, expected: <%v>, received: <%v>", prsr, eFldVal, fldVal)
	}
	prsr = config.NewRSRParserMustCompile(utils.DynamicDataPrefix+utils.OrderID, true)
	eFldVal = strconv.FormatInt(cdr.OrderID, 10)
	if fldVal, err := cdr.FieldAsString(prsr); err != nil {
		t.Error(err)
	} else if fldVal != eFldVal {
		t.Errorf("field: <%v>, expected: <%v>, received: <%v>", prsr, eFldVal, fldVal)
	}
	prsr = config.NewRSRParserMustCompile(utils.DynamicDataPrefix+utils.ToR, true)
	eFldVal = cdr.ToR
	if fldVal, err := cdr.FieldAsString(prsr); err != nil {
		t.Error(err)
	} else if fldVal != eFldVal {
		t.Errorf("field: <%v>, expected: <%v>, received: <%v>", prsr, eFldVal, fldVal)
	}
	prsr = config.NewRSRParserMustCompile(utils.DynamicDataPrefix+utils.OriginID, true)
	eFldVal = cdr.OriginID
	if fldVal, err := cdr.FieldAsString(prsr); err != nil {
		t.Error(err)
	} else if fldVal != eFldVal {
		t.Errorf("field: <%v>, expected: <%v>, received: <%v>", prsr, eFldVal, fldVal)
	}
	prsr = config.NewRSRParserMustCompile(utils.DynamicDataPrefix+utils.OriginHost, true)
	eFldVal = cdr.OriginHost
	if fldVal, err := cdr.FieldAsString(prsr); err != nil {
		t.Error(err)
	} else if fldVal != eFldVal {
		t.Errorf("field: <%v>, expected: <%v>, received: <%v>", prsr, eFldVal, fldVal)
	}
	prsr = config.NewRSRParserMustCompile(utils.DynamicDataPrefix+utils.Source, true)
	eFldVal = cdr.Source
	if fldVal, err := cdr.FieldAsString(prsr); err != nil {
		t.Error(err)
	} else if fldVal != eFldVal {
		t.Errorf("field: <%v>, expected: <%v>, received: <%v>", prsr, eFldVal, fldVal)
	}
	prsr = config.NewRSRParserMustCompile(utils.DynamicDataPrefix+utils.RequestType, true)
	eFldVal = cdr.RequestType
	if fldVal, err := cdr.FieldAsString(prsr); err != nil {
		t.Error(err)
	} else if fldVal != eFldVal {
		t.Errorf("field: <%v>, expected: <%v>, received: <%v>", prsr, eFldVal, fldVal)
	}
	prsr = config.NewRSRParserMustCompile(utils.DynamicDataPrefix+utils.Category, true)
	eFldVal = cdr.Category
	if fldVal, err := cdr.FieldAsString(prsr); err != nil {
		t.Error(err)
	} else if fldVal != eFldVal {
		t.Errorf("field: <%v>, expected: <%v>, received: <%v>", prsr, eFldVal, fldVal)
	}
	prsr = config.NewRSRParserMustCompile(utils.DynamicDataPrefix+utils.Account, true)
	eFldVal = cdr.Account
	if fldVal, err := cdr.FieldAsString(prsr); err != nil {
		t.Error(err)
	} else if fldVal != eFldVal {
		t.Errorf("field: <%v>, expected: <%v>, received: <%v>", prsr, eFldVal, fldVal)
	}
	prsr = config.NewRSRParserMustCompile(utils.DynamicDataPrefix+utils.Subject, true)
	eFldVal = cdr.Subject
	if fldVal, err := cdr.FieldAsString(prsr); err != nil {
		t.Error(err)
	} else if fldVal != eFldVal {
		t.Errorf("field: <%v>, expected: <%v>, received: <%v>", prsr, eFldVal, fldVal)
	}
	prsr = config.NewRSRParserMustCompile(utils.DynamicDataPrefix+utils.Destination, true)
	eFldVal = cdr.Destination
	if fldVal, err := cdr.FieldAsString(prsr); err != nil {
		t.Error(err)
	} else if fldVal != eFldVal {
		t.Errorf("field: <%v>, expected: <%v>, received: <%v>", prsr, eFldVal, fldVal)
	}
	prsr = config.NewRSRParserMustCompile(utils.DynamicDataPrefix+utils.SetupTime, true)
	eFldVal = cdr.SetupTime.Format(time.RFC3339)
	if fldVal, err := cdr.FieldAsString(prsr); err != nil {
		t.Error(err)
	} else if fldVal != eFldVal {
		t.Errorf("expected: <%s>, received: <%s>", eFldVal, fldVal)
	}
	prsr = config.NewRSRParserMustCompile(utils.DynamicDataPrefix+utils.AnswerTime, true)
	eFldVal = cdr.AnswerTime.Format(time.RFC3339)
	if fldVal, err := cdr.FieldAsString(prsr); err != nil {
		t.Error(err)
	} else if fldVal != eFldVal {
		t.Errorf("expected: <%s>, received: <%s>", eFldVal, fldVal)
	}
	prsr = config.NewRSRParserMustCompile(utils.DynamicDataPrefix+utils.Usage, true)
	eFldVal = "10s"
	if fldVal, err := cdr.FieldAsString(prsr); err != nil {
		t.Error(err)
	} else if fldVal != eFldVal {
		t.Errorf("field: <%v>, expected: <%v>, received: <%v>", prsr, eFldVal, fldVal)
	}
	prsr = config.NewRSRParserMustCompile(utils.DynamicDataPrefix+utils.RunID, true)
	eFldVal = cdr.RunID
	if fldVal, err := cdr.FieldAsString(prsr); err != nil {
		t.Error(err)
	} else if fldVal != eFldVal {
		t.Errorf("field: <%v>, expected: <%v>, received: <%v>", prsr, eFldVal, fldVal)
	}
	prsr = config.NewRSRParserMustCompile(utils.DynamicDataPrefix+utils.Cost, true)
	eFldVal = strconv.FormatFloat(cdr.Cost, 'f', -1, 64)
	if fldVal, err := cdr.FieldAsString(prsr); err != nil {
		t.Error(err)
	} else if fldVal != eFldVal {
		t.Errorf("field: <%v>, expected: <%v>, received: <%v>", prsr, eFldVal, fldVal)
	}
	prsr = config.NewRSRParserMustCompile(utils.DynamicDataPrefix+"field_extr1", true)
	eFldVal = cdr.ExtraFields["field_extr1"]
	if fldVal, err := cdr.FieldAsString(prsr); err != nil {
		t.Error(err)
	} else if fldVal != eFldVal {
		t.Errorf("field: <%v>, expected: <%v>, received: <%v>", prsr, eFldVal, fldVal)
	}
	prsr = config.NewRSRParserMustCompile(utils.DynamicDataPrefix+"fieldextr2", true)
	eFldVal = cdr.ExtraFields["fieldextr2"]
	if fldVal, err := cdr.FieldAsString(prsr); err != nil {
		t.Error(err)
	} else if fldVal != eFldVal {
		t.Errorf("field: <%v>, expected: <%v>, received: <%v>", prsr, eFldVal, fldVal)
	}
	prsr = config.NewRSRParserMustCompile(utils.DynamicDataPrefix+"dummy_field", true)
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
		RunID:      utils.DEFAULT_RUNID, Usage: time.Duration(10) * time.Second,
		ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"}, Cost: 1.01,
	}
	eVal := "call_from_1001"
	if val := cdr.FieldsAsString(
		config.NewRSRParsersMustCompile("~Category;_from_;~Account", true, utils.INFIELD_SEP)); val != eVal {
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
	if cdrForm.Get(utils.DIRECTION) != "*out" {
		t.Errorf("Expected: %s, received: %s", "*out", cdrForm.Get(utils.DIRECTION))
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

func TestCDRForkCdr(t *testing.T) {
	storCdr := CDR{CGRID: utils.Sha1("dsafdsaf", time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC).String()),
		OrderID: 123, ToR: utils.VOICE, OriginID: "dsafdsaf", OriginHost: "192.168.1.1", Source: utils.UNIT_TEST,
		RequestType: utils.META_RATED, Tenant: "cgrates.org", Category: "call", Account: "1001",
		Subject: "1001", Destination: "1002", RunID: utils.DEFAULT_RUNID,
		SetupTime:  time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC),
		AnswerTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC),
		Usage:      time.Duration(10) * time.Second, Cost: 1.01,
		ExtraFields: map[string]string{"field_extr1": "val_extr1", "field_extr2": "valextr2"}}
	rtSampleCdrOut, err := storCdr.ForkCdr("sample_run1", &utils.RSRField{Id: utils.RequestType}, &utils.RSRField{Id: utils.Tenant},
		&utils.RSRField{Id: utils.Category}, &utils.RSRField{Id: utils.Account}, &utils.RSRField{Id: utils.Subject}, &utils.RSRField{Id: utils.Destination},
		&utils.RSRField{Id: utils.SetupTime}, &utils.RSRField{Id: utils.AnswerTime}, &utils.RSRField{Id: utils.Usage},
		&utils.RSRField{Id: utils.PreRated}, &utils.RSRField{Id: utils.COST},
		[]*utils.RSRField{&utils.RSRField{Id: "field_extr1"}, &utils.RSRField{Id: "field_extr2"}}, true, "")
	if err != nil {
		t.Error("Unexpected error received", err)
	}
	expctSplPreRatedCdr := &CDR{CGRID: storCdr.CGRID, ToR: utils.VOICE, OriginID: "dsafdsaf",
		OriginHost: "192.168.1.1", Source: utils.UNIT_TEST, RequestType: utils.META_RATED,
		Tenant: "cgrates.org", Category: "call", Account: "1001", Subject: "1001", Destination: "1002",
		SetupTime:   time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC),
		AnswerTime:  time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC),
		Usage:       time.Duration(10) * time.Second,
		ExtraFields: map[string]string{"field_extr1": "val_extr1", "field_extr2": "valextr2"},
		RunID:       "sample_run1", PreRated: false, Cost: 1.01}
	if !reflect.DeepEqual(expctSplPreRatedCdr, rtSampleCdrOut) {
		t.Errorf("Expected: %v, received: %v", expctSplPreRatedCdr, rtSampleCdrOut)
	}
}

func TestCDRForkCdrStaticVals(t *testing.T) {
	storCdr := CDR{
		CGRID:   utils.Sha1("dsafdsaf", time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC).String()),
		OrderID: 123, ToR: utils.VOICE, OriginID: "dsafdsaf",
		OriginHost: "192.168.1.1", Source: utils.UNIT_TEST, RequestType: utils.META_RATED,
		Tenant: "cgrates.org", Category: "call", Account: "1001", Subject: "1001", Destination: "1002",
		SetupTime:  time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC),
		AnswerTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC), RunID: utils.DEFAULT_RUNID,
		Usage: time.Duration(10) * time.Second, Cost: 1.01,
		ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
	}
	rsrStPostpaid, _ := utils.NewRSRField("^" + utils.META_POSTPAID)
	rsrStCgr, _ := utils.NewRSRField("^cgrates.com")
	rsrStPC, _ := utils.NewRSRField("^premium_call")
	rsrStFA, _ := utils.NewRSRField("^first_account")
	rsrStFS, _ := utils.NewRSRField("^first_subject")
	rsrStST, _ := utils.NewRSRField("^2013-12-07T08:42:24Z")
	rsrStAT, _ := utils.NewRSRField("^2013-12-07T08:42:26Z")
	rsrStDur, _ := utils.NewRSRField("^12s")
	rsrStPreRated, _ := utils.NewRSRField("^true")
	rsrStCost, _ := utils.NewRSRField("^1.2")
	rtCdrOut2, err := storCdr.ForkCdr("wholesale_run", rsrStPostpaid, rsrStCgr, rsrStPC, rsrStFA, rsrStFS, &utils.RSRField{Id: utils.Destination},
		rsrStST, rsrStAT, rsrStDur, rsrStPreRated, rsrStCost, []*utils.RSRField{}, true, "")
	if err != nil {
		t.Error("Unexpected error received", err)
	}
	expctPreRatedCdr2 := &CDR{CGRID: storCdr.CGRID, ToR: utils.VOICE, OriginID: "dsafdsaf",
		OriginHost: "192.168.1.1", Source: utils.UNIT_TEST, RequestType: utils.META_POSTPAID,
		Tenant: "cgrates.com", Category: "premium_call", Account: "first_account",
		Subject: "first_subject", Destination: "1002",
		SetupTime:  time.Date(2013, 12, 7, 8, 42, 24, 0, time.UTC),
		AnswerTime: time.Date(2013, 12, 7, 8, 42, 26, 0, time.UTC),
		Usage:      time.Duration(12) * time.Second, PreRated: true, Cost: 1.2,
		ExtraFields: map[string]string{}, RunID: "wholesale_run"}
	if !reflect.DeepEqual(rtCdrOut2, expctPreRatedCdr2) {
		t.Errorf("Received: %v, expected: %v", rtCdrOut2, expctPreRatedCdr2)
	}
	_, err = storCdr.ForkCdr("wholesale_run", &utils.RSRField{Id: "dummy_header"},
		&utils.RSRField{Id: utils.Tenant}, &utils.RSRField{Id: utils.ToR},
		&utils.RSRField{Id: utils.Account}, &utils.RSRField{Id: utils.Subject},
		&utils.RSRField{Id: utils.Destination}, &utils.RSRField{Id: utils.SetupTime},
		&utils.RSRField{Id: utils.AnswerTime}, &utils.RSRField{Id: utils.Usage},
		&utils.RSRField{Id: utils.PreRated}, &utils.RSRField{Id: utils.COST},
		[]*utils.RSRField{}, true, "")
	if err == nil {
		t.Error("Failed to detect missing header")
	}
}

func TestCDRForkCdrFromMetaDefaults(t *testing.T) {
	storCdr := CDR{CGRID: utils.Sha1("dsafdsaf", time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC).String()),
		OrderID: 123, ToR: utils.VOICE, OriginID: "dsafdsaf", OriginHost: "192.168.1.1", Source: utils.UNIT_TEST,
		RequestType: utils.META_RATED, Tenant: "cgrates.org", Category: "call",
		Account: "1001", Subject: "1001", Destination: "1002", RunID: utils.DEFAULT_RUNID,
		SetupTime:  time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC),
		AnswerTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC),
		Usage:      time.Duration(10) * time.Second, Cost: 1.01,
		ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
	}
	expctCdr := &CDR{CGRID: storCdr.CGRID, ToR: utils.VOICE, OriginID: "dsafdsaf",
		OriginHost: "192.168.1.1", Source: utils.UNIT_TEST, RequestType: utils.META_RATED,
		Tenant: "cgrates.org", Category: "call", Account: "1001", Subject: "1001", Destination: "1002",
		SetupTime:  time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC),
		AnswerTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC),
		Usage:      time.Duration(10) * time.Second, Cost: 1.01, RunID: "wholesale_run",
		ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"}}
	cdrOut, err := storCdr.ForkCdr("wholesale_run", &utils.RSRField{Id: utils.META_DEFAULT}, &utils.RSRField{Id: utils.META_DEFAULT},
		&utils.RSRField{Id: utils.META_DEFAULT}, &utils.RSRField{Id: utils.META_DEFAULT}, &utils.RSRField{Id: utils.META_DEFAULT},
		&utils.RSRField{Id: utils.META_DEFAULT}, &utils.RSRField{Id: utils.META_DEFAULT}, &utils.RSRField{Id: utils.META_DEFAULT},
		&utils.RSRField{Id: utils.META_DEFAULT}, &utils.RSRField{Id: utils.META_DEFAULT}, &utils.RSRField{Id: utils.META_DEFAULT},
		[]*utils.RSRField{&utils.RSRField{Id: "field_extr1"}, &utils.RSRField{Id: "fieldextr2"}}, true, "")
	if err != nil {
		t.Fatal("Unexpected error received", err)
	}

	if !reflect.DeepEqual(expctCdr, cdrOut) {
		t.Errorf("Expected: %v, received: %v", expctCdr, cdrOut)
	}
	// Should also accept nil as defaults
	if cdrOut, err := storCdr.ForkCdr("wholesale_run", nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
		[]*utils.RSRField{&utils.RSRField{Id: "field_extr1"}, &utils.RSRField{Id: "fieldextr2"}}, true, ""); err != nil {
		t.Fatal("Unexpected error received", err)
	} else if !reflect.DeepEqual(expctCdr, cdrOut) {
		t.Errorf("Expected: %v, received: %v", expctCdr, cdrOut)
	}
}

func TestCDRAsExternalCDR(t *testing.T) {
	storCdr := CDR{CGRID: utils.Sha1("dsafdsaf", time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC).String()),
		OrderID: 123, ToR: utils.VOICE, OriginID: "dsafdsaf", OriginHost: "192.168.1.1",
		Source: utils.UNIT_TEST, RequestType: utils.META_RATED,
		Tenant: "cgrates.org", Category: "call", Account: "1001", Subject: "1001", Destination: "1002",
		SetupTime:  time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC),
		AnswerTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC),
		RunID:      utils.DEFAULT_RUNID, Usage: time.Duration(10 * time.Second), Cost: 1.01,
		ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"}}
	expectOutCdr := &ExternalCDR{
		CGRID:   utils.Sha1("dsafdsaf", time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC).String()),
		OrderID: 123, ToR: utils.VOICE, OriginID: "dsafdsaf", OriginHost: "192.168.1.1",
		Source: utils.UNIT_TEST, RequestType: utils.META_RATED,
		Tenant: "cgrates.org", Category: "call", Account: "1001", Subject: "1001", Destination: "1002",
		SetupTime: "2013-11-07T08:42:20Z", AnswerTime: "2013-11-07T08:42:26Z", RunID: utils.DEFAULT_RUNID,
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
	eCD := &CallDescriptor{CgrID: "c4630df20b2a0c5b11311e4b5a8c3178cf314344", TOR: req.ToR,
		Direction: utils.OUT, Tenant: req.Tenant,
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

func TestCDRParseFieldValue(t *testing.T) {
	cdr := new(CDR)
	if err := cdr.ParseFieldValue(utils.Partial, "true", ""); err != nil {
		t.Error(err)
	} else if !cdr.Partial {
		t.Errorf("Received cdr: %+v", cdr)
	}
	if err := cdr.ParseFieldValue(utils.OrderID, "5", ""); err != nil {
		t.Error(err)
	} else if cdr.OrderID != 5 {
		t.Errorf("Received cdr: %+v", cdr)
	}
	if err := cdr.ParseFieldValue(utils.RunID, "*default", ""); err != nil {
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
		Tenant:      "cgrates.org",
		Category:    "call",
		Account:     "1002",
		Subject:     "1001",
		Destination: "+4986517174963",
		SetupTime:   time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC),
		AnswerTime:  time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC),
		RunID:       utils.DEFAULT_RUNID,
		Usage:       time.Duration(10) * time.Second,
		ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
		Cost:        1.01,
	}

	mp := map[string]interface{}{
		"field_extr1":     "val_extr1",
		"fieldextr2":      "valextr2",
		utils.CGRID:       cdr.CGRID,
		utils.RunID:       utils.DEFAULT_RUNID,
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
		utils.CostDetails: cdr.CostDetails,
		utils.PreRated:    false,
		utils.Partial:     false,
		utils.ExtraInfo:   cdr.ExtraInfo,
	}
	if cdrMp := cdr.AsMapStringIface(); !reflect.DeepEqual(mp, cdrMp) {
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
		Tenant:      "cgrates.org",
		Category:    "call",
		Account:     "1001",
		Subject:     "1001",
		Destination: "+4986517174963",
		SetupTime:   time.Unix(1383813745, 0).UTC(),
		AnswerTime:  time.Unix(1383813746, 0).UTC(),
		Usage:       time.Duration(10) * time.Second,
		RunID:       utils.DEFAULT_RUNID, Cost: 1.01,
		ExtraFields: map[string]string{"stop_time": "2014-06-11 19:19:00 +0000 UTC", "fieldextr2": "valextr2"}}

	prsr := config.NewRSRParsersMustCompile(utils.DynamicDataPrefix+utils.Destination, true, utils.INFIELD_SEP)
	cfgCdrFld := &config.FCTemplate{Tag: "destination", Type: utils.META_COMPOSED,
		FieldId: utils.Destination, Value: prsr, Timezone: "UTC"}
	if expRecord, err := cdr.AsExportRecord([]*config.FCTemplate{cfgCdrFld}, false, nil, 0, nil); err != nil {
		t.Error(err)
	} else if expRecord[0] != cdr.Destination {
		t.Errorf("Expecting:\n%s\nReceived:\n%s", cdr.Destination, expRecord)
	}
	if err := dm.DataDB().SetReverseDestination(&Destination{Id: "MASKED_DESTINATIONS", Prefixes: []string{"+4986517174963"}},
		utils.NonTransactional); err != nil {
		t.Error(err)
	}

	cfgCdrFld = &config.FCTemplate{Tag: "Destination", Type: utils.META_COMPOSED,
		FieldId: utils.Destination, Value: prsr, MaskDestID: "MASKED_DESTINATIONS", MaskLen: 3}
	eDst := "+4986517174***"
	if expRecord, err := cdr.AsExportRecord([]*config.FCTemplate{cfgCdrFld}, false, nil, 0, nil); err != nil {
		t.Error(err)
	} else if expRecord[0] != eDst {
		t.Errorf("Expecting:\n%s\nReceived:\n%s", eDst, expRecord[0])
	}

	cfgCdrFld = &config.FCTemplate{Tag: "MaskedDest", Type: utils.MetaMaskedDestination,
		Value: prsr, MaskDestID: "MASKED_DESTINATIONS"}
	if expRecord, err := cdr.AsExportRecord([]*config.FCTemplate{cfgCdrFld}, false, nil, 0, nil); err != nil {
		t.Error(err)
	} else if expRecord[0] != "1" {
		t.Errorf("Expecting:\n%s\nReceived:\n%s", "1", expRecord[0])
	}
	data, _ := NewMapStorage()
	dmForCDR := NewDataManager(data)
	defaultCfg, err := config.NewDefaultCGRConfig()
	if err != nil {
		t.Errorf("Error: %+v", err)
	}

	cfgCdrFld = &config.FCTemplate{Tag: "destination", Type: utils.META_COMPOSED,
		FieldId: utils.Destination, Value: prsr, Filters: []string{"*string:Tenant:itsyscom.com"}, Timezone: "UTC"}
	if rcrd, err := cdr.AsExportRecord([]*config.FCTemplate{cfgCdrFld}, false, nil, 0, &FilterS{dm: dmForCDR, cfg: defaultCfg}); err != nil {
		t.Error(err)
	} else if len(rcrd) != 0 {
		t.Error("failed using filter")
	}

	// Test MetaDateTime
	prsr = config.NewRSRParsersMustCompile(utils.DynamicDataPrefix+"stop_time", true, utils.INFIELD_SEP)
	layout := "2006-01-02 15:04:05"
	cfgCdrFld = &config.FCTemplate{Tag: "stop_time", Type: utils.MetaDateTime,
		FieldId: "stop_time", Value: prsr, Layout: layout, Timezone: "UTC"}
	if expRecord, err := cdr.AsExportRecord([]*config.FCTemplate{cfgCdrFld}, false, nil, 0, &FilterS{dm: dmForCDR, cfg: defaultCfg}); err != nil {
		t.Error(err)
	} else if expRecord[0] != "2014-06-11 19:19:00" {
		t.Error("Expecting: 2014-06-11 19:19:00, got: ", expRecord[0])
	}

	// Test filter
	cfgCdrFld = &config.FCTemplate{Tag: "stop_time", Type: utils.MetaDateTime,
		FieldId: "stop_time", Value: prsr, Filters: []string{"*string:Tenant:itsyscom.com"},
		Layout: layout, Timezone: "UTC"}
	if rcrd, err := cdr.AsExportRecord([]*config.FCTemplate{cfgCdrFld}, false, nil, 0, &FilterS{dm: dmForCDR, cfg: defaultCfg}); err != nil {
		t.Error(err)
	} else if len(rcrd) != 0 {
		t.Error("failed using filter")
	}

	prsr = config.NewRSRParsersMustCompile(utils.DynamicDataPrefix+"fieldextr2", true, utils.INFIELD_SEP)
	cfgCdrFld = &config.FCTemplate{Tag: "stop_time", Type: utils.MetaDateTime,
		FieldId: "stop_time", Value: prsr, Layout: layout, Timezone: "UTC"}
	// Test time parse error
	if _, err := cdr.AsExportRecord([]*config.FCTemplate{cfgCdrFld}, false, nil, 0, nil); err == nil {
		t.Error("Should give error here, got none.")
	}
}

func TestCDRAsExportMap(t *testing.T) {
	cdr := &CDR{CGRID: utils.Sha1("dsafdsaf", time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC).String()),
		OrderID: 123, ToR: utils.VOICE, OriginID: "dsafdsaf",
		OriginHost: "192.168.1.1", Source: utils.UNIT_TEST, RequestType: utils.META_RATED,
		Tenant: "cgrates.org", Category: "call", Account: "1001",
		Subject: "1001", Destination: "+4986517174963",
		SetupTime:  time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC),
		AnswerTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC), RunID: utils.DEFAULT_RUNID,
		Usage: time.Duration(10) * time.Second, Cost: 1.01,
		ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
	}
	eCDRMp := map[string]string{
		utils.CGRID:       cdr.CGRID,
		utils.Destination: "004986517174963",
		"FieldExtra1":     "val_extr1",
	}
	expFlds := []*config.FCTemplate{
		&config.FCTemplate{FieldId: utils.CGRID, Type: utils.META_COMPOSED,
			Value: config.NewRSRParsersMustCompile(utils.DynamicDataPrefix+utils.CGRID, true, utils.INFIELD_SEP)},
		&config.FCTemplate{FieldId: utils.Destination, Type: utils.META_COMPOSED,
			Value: config.NewRSRParsersMustCompile("~Destination:s/^\\+(\\d+)$/00${1}/", true, utils.INFIELD_SEP)},
		&config.FCTemplate{FieldId: "FieldExtra1", Type: utils.META_COMPOSED,
			Value: config.NewRSRParsersMustCompile(utils.DynamicDataPrefix+"field_extr1", true, utils.INFIELD_SEP)},
	}
	if cdrMp, err := cdr.AsExportMap(expFlds, false, nil, 0, nil); err != nil {
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
		RunID:       utils.DEFAULT_RUNID,
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
		RunID:       "*default",
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
		RunID:       utils.DEFAULT_RUNID,
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
		RunID:       utils.DEFAULT_RUNID,
		Usage:       time.Duration(10) * time.Second,
		Cost:        1.01,
		ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
	}
	var costdetails *EventCost
	eCGREvent := utils.CGREvent{
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
			"OriginHost":  "192.168.1.1",
			"OriginID":    "dsafdsaf",
			"Partial":     false,
			"RequestType": utils.META_RATED,
			"RunID":       "*default",
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
		RunID:       utils.DEFAULT_RUNID,
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
			"RunID":       "*default",
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
		RunID:       utils.DEFAULT_RUNID,
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

func TestCDRParseFieldValue2(t *testing.T) {
	cdr := new(CDR)
	if err := cdr.ParseFieldValue(utils.RunID, "*default", ""); err != nil {
		t.Error(err)
	} else if cdr.RunID != "*default" {
		t.Errorf("Received cdr: %+v", cdr)
	}
	if err := cdr.ParseFieldValue(utils.OriginID, "FirstID", ""); err != nil {
		t.Error(err)
	} else if cdr.OriginID != "FirstID" {
		t.Errorf("Received cdr: %+v", cdr)
	}
	if err := cdr.ParseFieldValue(utils.OriginID, "SecondID", ""); err != nil {
		t.Error(err)
	} else if cdr.OriginID != "SecondID" {
		t.Errorf("Received cdr: %+v", cdr)
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
		RunID:       utils.MetaRaw,
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
