/*
Real-time Charging System for Telecom & ISP environments
Copyright (C) 2012-2015 ITsysCOM GmbH

This program is free software: you can Storagetribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITH*out ANY WARRANTY; without even the implied warranty of
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

	"github.com/cgrates/cgrates/utils"
)

func TestStoredCdrInterfaces(t *testing.T) {
	storedCdr := new(StoredCdr)
	var _ RawCdr = storedCdr
	var _ Event = storedCdr
}

func TestNewStoredCdrFromExternalCdr(t *testing.T) {
	extCdr := &ExternalCdr{CgrId: utils.Sha1("dsafdsaf", time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC).String()), OrderId: 123, TOR: utils.VOICE,
		AccId: "dsafdsaf", CdrHost: "192.168.1.1", CdrSource: utils.UNIT_TEST, ReqType: utils.META_RATED, Direction: "*out",
		Tenant: "cgrates.org", Category: "call", Account: "1001", Subject: "1001", Destination: "1002", Supplier: "SUPPL1",
		SetupTime: "2013-11-07T08:42:20Z", AnswerTime: "2013-11-07T08:42:26Z", MediationRunId: utils.DEFAULT_RUNID,
		Usage: "0.00000001", Pdd: "7.0", ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"}, Cost: 1.01, RatedAccount: "dan", RatedSubject: "dans", Rated: true,
	}
	eStorCdr := &StoredCdr{CgrId: utils.Sha1("dsafdsaf", time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC).String()), OrderId: 123, TOR: utils.VOICE,
		AccId: "dsafdsaf", CdrHost: "192.168.1.1", CdrSource: utils.UNIT_TEST, ReqType: utils.META_RATED, Direction: "*out",
		Tenant: "cgrates.org", Category: "call", Account: "1001", Subject: "1001", Destination: "1002", Supplier: "SUPPL1",
		SetupTime: time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC), AnswerTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC), MediationRunId: utils.DEFAULT_RUNID,
		Usage: time.Duration(10), Pdd: time.Duration(7) * time.Second, ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"}, Cost: 1.01, RatedAccount: "dan", RatedSubject: "dans", Rated: true,
	}
	if storedCdr, err := NewStoredCdrFromExternalCdr(extCdr, ""); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eStorCdr, storedCdr) {
		t.Errorf("Expected: %+v, received: %+v", eStorCdr, storedCdr)
	}
}

func TestStoredCdrClone(t *testing.T) {
	storCdr := &StoredCdr{CgrId: utils.Sha1("dsafdsaf", time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC).String()), OrderId: 123, TOR: utils.VOICE,
		AccId: "dsafdsaf", CdrHost: "192.168.1.1", CdrSource: utils.UNIT_TEST, ReqType: utils.META_RATED, Direction: "*out",
		Tenant: "cgrates.org", Category: "call", Account: "1001", Subject: "1001", Destination: "1002", Supplier: "SUPPL1",
		SetupTime: time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC), AnswerTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC), MediationRunId: utils.DEFAULT_RUNID,
		Usage: time.Duration(10), Pdd: time.Duration(7) * time.Second, ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"}, Cost: 1.01, RatedAccount: "dan", RatedSubject: "dans", Rated: true,
	}
	if clnStorCdr := storCdr.Clone(); !reflect.DeepEqual(storCdr, clnStorCdr) {
		t.Errorf("Expecting: %+v, received: %+v", storCdr, clnStorCdr)
	}
}

func TestFieldAsString(t *testing.T) {
	cdr := StoredCdr{CgrId: utils.Sha1("dsafdsaf", time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC).String()), OrderId: 123, TOR: utils.VOICE, AccId: "dsafdsaf",
		CdrHost: "192.168.1.1", CdrSource: "test", ReqType: utils.META_RATED, Direction: "*out", Tenant: "cgrates.org",
		Category: "call", Account: "1001", Subject: "1001", Destination: "1002", SetupTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC),
		AnswerTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC), MediationRunId: utils.DEFAULT_RUNID,
		Usage: time.Duration(10) * time.Second, Pdd: time.Duration(5) * time.Second, Supplier: "SUPPL1", ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
		Cost: 1.01, RatedAccount: "dan", RatedSubject: "dans",
	}
	if cdr.FieldAsString(&utils.RSRField{Id: utils.CGRID}) != cdr.CgrId ||
		cdr.FieldAsString(&utils.RSRField{Id: utils.ORDERID}) != "123" ||
		cdr.FieldAsString(&utils.RSRField{Id: utils.TOR}) != utils.VOICE ||
		cdr.FieldAsString(&utils.RSRField{Id: utils.ACCID}) != cdr.AccId ||
		cdr.FieldAsString(&utils.RSRField{Id: utils.CDRHOST}) != cdr.CdrHost ||
		cdr.FieldAsString(&utils.RSRField{Id: utils.CDRSOURCE}) != cdr.CdrSource ||
		cdr.FieldAsString(&utils.RSRField{Id: utils.REQTYPE}) != cdr.ReqType ||
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
		cdr.FieldAsString(&utils.RSRField{Id: utils.MEDI_RUNID}) != cdr.MediationRunId ||
		cdr.FieldAsString(&utils.RSRField{Id: utils.COST}) != "1.01" ||
		cdr.FieldAsString(&utils.RSRField{Id: utils.RATED_ACCOUNT}) != "dan" ||
		cdr.FieldAsString(&utils.RSRField{Id: utils.RATED_SUBJECT}) != "dans" ||
		cdr.FieldAsString(&utils.RSRField{Id: "field_extr1"}) != cdr.ExtraFields["field_extr1"] ||
		cdr.FieldAsString(&utils.RSRField{Id: "fieldextr2"}) != cdr.ExtraFields["fieldextr2"] ||
		cdr.FieldAsString(&utils.RSRField{Id: "dummy_field"}) != "" {
		t.Error("Unexpected filed value received",
			cdr.FieldAsString(&utils.RSRField{Id: utils.CGRID}) != cdr.CgrId,
			cdr.FieldAsString(&utils.RSRField{Id: utils.ORDERID}) != "123",
			cdr.FieldAsString(&utils.RSRField{Id: utils.TOR}) != utils.VOICE,
			cdr.FieldAsString(&utils.RSRField{Id: utils.ACCID}) != cdr.AccId,
			cdr.FieldAsString(&utils.RSRField{Id: utils.CDRHOST}) != cdr.CdrHost,
			cdr.FieldAsString(&utils.RSRField{Id: utils.CDRSOURCE}) != cdr.CdrSource,
			cdr.FieldAsString(&utils.RSRField{Id: utils.REQTYPE}) != cdr.ReqType,
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
			cdr.FieldAsString(&utils.RSRField{Id: utils.MEDI_RUNID}) != cdr.MediationRunId,
			cdr.FieldAsString(&utils.RSRField{Id: utils.RATED_ACCOUNT}) != "dan",
			cdr.FieldAsString(&utils.RSRField{Id: utils.RATED_SUBJECT}) != "dans",
			cdr.FieldAsString(&utils.RSRField{Id: utils.COST}) != "1.01",
			cdr.FieldAsString(&utils.RSRField{Id: "field_extr1"}) != cdr.ExtraFields["field_extr1"],
			cdr.FieldAsString(&utils.RSRField{Id: "fieldextr2"}) != cdr.ExtraFields["fieldextr2"],
			cdr.FieldAsString(&utils.RSRField{Id: "dummy_field"}) != "")
	}
}

func TestFieldsAsString(t *testing.T) {
	cdr := StoredCdr{CgrId: utils.Sha1("dsafdsaf", time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC).String()), OrderId: 123, TOR: utils.VOICE, AccId: "dsafdsaf",
		CdrHost: "192.168.1.1", CdrSource: "test", ReqType: utils.META_RATED, Direction: "*out", Tenant: "cgrates.org",
		Category: "call", Account: "1001", Subject: "1001", Destination: "1002", SetupTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC),
		AnswerTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC), MediationRunId: utils.DEFAULT_RUNID,
		Usage: time.Duration(10) * time.Second, Pdd: time.Duration(5) * time.Second, Supplier: "SUPPL1", ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
		Cost: 1.01, RatedAccount: "dan", RatedSubject: "dans",
	}
	eVal := "call_from_1001"
	if val := cdr.FieldsAsString(utils.ParseRSRFieldsMustCompile("Category;^_from_;Account", utils.INFIELD_SEP)); val != eVal {
		t.Errorf("Expecting : %s, received: %s", eVal, val)
	}
}

func TestPassesFieldFilter(t *testing.T) {
	cdr := &StoredCdr{CgrId: utils.Sha1("dsafdsaf", time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC).String()), OrderId: 123, TOR: utils.VOICE, AccId: "dsafdsaf",
		CdrHost: "192.168.1.1", CdrSource: "test", ReqType: utils.META_RATED, Direction: "*out", Tenant: "cgrates.org",
		Category: "call", Account: "1001", Subject: "1001", Destination: "1002", SetupTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC),
		AnswerTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC), MediationRunId: utils.DEFAULT_RUNID,
		Usage: time.Duration(10) * time.Second, ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"}, Cost: 1.01,
	}
	if pass, _ := cdr.PassesFieldFilter(nil); !pass {
		t.Error("Not passing filter")
	}
	acntPrefxFltr, _ := utils.NewRSRField(`~Account:s/(.+)/1001/`)
	if pass, _ := cdr.PassesFieldFilter(acntPrefxFltr); !pass {
		t.Error("Not passing filter")
	}
	acntPrefxFltr, _ = utils.NewRSRField(`~Account:s/^(10)\d\d$/10/`)
	if pass, _ := cdr.PassesFieldFilter(acntPrefxFltr); !pass {
		t.Error("Not passing valid filter")
	}
	acntPrefxFltr, _ = utils.NewRSRField(`~Account:s/^\d(10)\d$/10/`)
	if pass, _ := cdr.PassesFieldFilter(acntPrefxFltr); pass {
		t.Error("Passing filter")
	}
	acntPrefxFltr, _ = utils.NewRSRField(`~Account:s/^(10)\d\d$/010/`)
	if pass, _ := cdr.PassesFieldFilter(acntPrefxFltr); pass {
		t.Error("Passing filter")
	}
	acntPrefxFltr, _ = utils.NewRSRField(`~Account:s/^1010$/1010/`)
	if pass, _ := cdr.PassesFieldFilter(acntPrefxFltr); pass {
		t.Error("Passing filter")
	}
	torFltr, _ := utils.NewRSRField(`^TOR::*voice/`)
	if pass, _ := cdr.PassesFieldFilter(torFltr); !pass {
		t.Error("Not passing filter")
	}
	torFltr, _ = utils.NewRSRField(`^TOR/*data/`)
	if pass, _ := cdr.PassesFieldFilter(torFltr); pass {
		t.Error("Passing filter")
	}
	inexistentFieldFltr, _ := utils.NewRSRField(`^fakefield/fakevalue/`)
	if pass, _ := cdr.PassesFieldFilter(inexistentFieldFltr); pass {
		t.Error("Passing filter")
	}
}

func TestPassesFieldFilterDn1(t *testing.T) {
	cdr := &StoredCdr{CgrId: utils.Sha1("dsafdsaf", time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC).String()), Account: "futurem0005",
		ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"}, Cost: 1.01,
	}
	acntPrefxFltr, _ := utils.NewRSRField(`~Account:s/^\w+[shmp]\d{4}$//`)
	if pass, _ := cdr.PassesFieldFilter(acntPrefxFltr); !pass {
		t.Error("Not passing valid filter")
	}

	cdr = &StoredCdr{CgrId: utils.Sha1("dsafdsaf", time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC).String()), Account: "futurem00005",
		ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"}, Cost: 1.01,
	}
	if pass, _ := cdr.PassesFieldFilter(acntPrefxFltr); pass {
		t.Error("Should not pass filter")
	}
	cdr = &StoredCdr{CgrId: utils.Sha1("dsafdsaf", time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC).String()), Account: "0402129281",
		ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"}, Cost: 1.01,
	}
	acntPrefxFltr, _ = utils.NewRSRField(`~Account:s/^0\d{9}$//`)
	if pass, _ := cdr.PassesFieldFilter(acntPrefxFltr); !pass {
		t.Error("Not passing valid filter")
	}
	acntPrefxFltr, _ = utils.NewRSRField(`~Account:s/^0(\d{9})$/placeholder/`)
	if pass, _ := cdr.PassesFieldFilter(acntPrefxFltr); pass {
		t.Error("Should not pass filter")
	}
	cdr = &StoredCdr{CgrId: utils.Sha1("dsafdsaf", time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC).String()), Account: "04021292812",
		ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"}, Cost: 1.01,
	}
	if pass, _ := cdr.PassesFieldFilter(acntPrefxFltr); pass {
		t.Error("Should not pass filter")
	}
	cdr = &StoredCdr{CgrId: utils.Sha1("dsafdsaf", time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC).String()), Account: "0162447222",
		ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"}, Cost: 1.01,
	}
	if acntPrefxFltr, err := utils.NewRSRField(`~Account:s/^0\d{9}$//`); err != nil {
		t.Error("Unexpected parse error", err)
	} else if acntPrefxFltr == nil {
		t.Error("Failed parsing rule")
	} else if pass, _ := cdr.PassesFieldFilter(acntPrefxFltr); !pass {
		t.Error("Not passing valid filter")
	}
	if acntPrefxFltr, err := utils.NewRSRField(`~Account:s/^\w+[shmp]\d{4}$//`); err != nil {
		t.Error("Unexpected parse error", err)
	} else if acntPrefxFltr == nil {
		t.Error("Failed parsing rule")
	} else if pass, _ := cdr.PassesFieldFilter(acntPrefxFltr); pass {
		t.Error("Should not pass filter")
	}
}

func TestUsageMultiply(t *testing.T) {
	cdr := StoredCdr{Usage: time.Duration(10) * time.Second}
	if cdr.UsageMultiply(1024.0, 0); cdr.Usage != time.Duration(10240)*time.Second {
		t.Errorf("Unexpected usage after multiply: %v", cdr.Usage.Nanoseconds())
	}
	cdr = StoredCdr{Usage: time.Duration(10240) * time.Second} // Simulate conversion back, gives out a bit odd result but this can be rounded on export
	expectDuration, _ := time.ParseDuration("10.000005120s")
	if cdr.UsageMultiply(0.000976563, 0); cdr.Usage != expectDuration {
		t.Errorf("Unexpected usage after multiply: %v", cdr.Usage.Nanoseconds())
	}
}

func TestCostMultiply(t *testing.T) {
	cdr := StoredCdr{Cost: 1.01}
	if cdr.CostMultiply(1.19, 4); cdr.Cost != 1.2019 {
		t.Errorf("Unexpected cost after multiply: %v", cdr.Cost)
	}
	cdr = StoredCdr{Cost: 1.01}
	if cdr.CostMultiply(1000, 0); cdr.Cost != 1010 {
		t.Errorf("Unexpected cost after multiply: %v", cdr.Cost)
	}
}

func TestFormatCost(t *testing.T) {
	cdr := StoredCdr{Cost: 1.01}
	if cdr.FormatCost(0, 4) != "1.0100" {
		t.Error("Unexpected format of the cost: ", cdr.FormatCost(0, 4))
	}
	cdr = StoredCdr{Cost: 1.01001}
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
	cdr := StoredCdr{Usage: time.Duration(10) * time.Second}
	if cdr.FormatUsage(utils.SECONDS) != "10" {
		t.Error("Wrong usage format: ", cdr.FormatUsage(utils.SECONDS))
	}
	if cdr.FormatUsage("default") != "10" {
		t.Error("Wrong usage format: ", cdr.FormatUsage("default"))
	}
	cdr = StoredCdr{TOR: utils.DATA, Usage: time.Duration(1640113000000000)}
	if cdr.FormatUsage("default") != "1640113" {
		t.Error("Wrong usage format: ", cdr.FormatUsage("default"))
	}
	cdr = StoredCdr{Usage: time.Duration(2) * time.Millisecond}
	if cdr.FormatUsage("default") != "0.002" {
		t.Error("Wrong usage format: ", cdr.FormatUsage("default"))
	}
	cdr = StoredCdr{Usage: time.Duration(1002) * time.Millisecond}
	if cdr.FormatUsage("default") != "1.002" {
		t.Error("Wrong usage format: ", cdr.FormatUsage("default"))
	}
}

func TestStoredCdrAsHttpForm(t *testing.T) {
	storCdr := StoredCdr{CgrId: utils.Sha1("dsafdsaf", time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC).String()), OrderId: 123, TOR: utils.VOICE, AccId: "dsafdsaf",
		CdrHost: "192.168.1.1", CdrSource: utils.UNIT_TEST, ReqType: utils.META_RATED, Direction: "*out",
		Tenant: "cgrates.org", Category: "call", Account: "1001", Subject: "1001", Destination: "1002",
		SetupTime: time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC), AnswerTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC), MediationRunId: utils.DEFAULT_RUNID,
		Usage: time.Duration(10) * time.Second, Supplier: "SUPPL1",
		ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"}, RatedSubject: "dans", Cost: 1.01,
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

func TestStoredCdrForkCdr(t *testing.T) {
	storCdr := StoredCdr{CgrId: utils.Sha1("dsafdsaf", time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC).String()), OrderId: 123, TOR: utils.VOICE,
		AccId: "dsafdsaf", CdrHost: "192.168.1.1", CdrSource: utils.UNIT_TEST, ReqType: utils.META_RATED, Direction: "*out",
		Tenant: "cgrates.org", Category: "call", Account: "1001", Subject: "1001", Destination: "1002",
		SetupTime: time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC), Pdd: time.Duration(200) * time.Millisecond, AnswerTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC), MediationRunId: utils.DEFAULT_RUNID,
		Usage: time.Duration(10) * time.Second, Supplier: "suppl1", ExtraFields: map[string]string{"field_extr1": "val_extr1", "field_extr2": "valextr2"},
		Cost: 1.01, RatedSubject: "dans"}
	rtSampleCdrOut, err := storCdr.ForkCdr("sample_run1", &utils.RSRField{Id: utils.REQTYPE}, &utils.RSRField{Id: utils.DIRECTION}, &utils.RSRField{Id: utils.TENANT},
		&utils.RSRField{Id: utils.CATEGORY}, &utils.RSRField{Id: utils.ACCOUNT}, &utils.RSRField{Id: utils.SUBJECT}, &utils.RSRField{Id: utils.DESTINATION},
		&utils.RSRField{Id: utils.SETUP_TIME}, &utils.RSRField{Id: utils.PDD}, &utils.RSRField{Id: utils.ANSWER_TIME}, &utils.RSRField{Id: utils.USAGE},
		&utils.RSRField{Id: utils.SUPPLIER}, &utils.RSRField{Id: utils.DISCONNECT_CAUSE}, &utils.RSRField{Id: utils.RATED_FLD}, &utils.RSRField{Id: utils.COST},
		[]*utils.RSRField{&utils.RSRField{Id: "field_extr1"}, &utils.RSRField{Id: "field_extr2"}}, true, "")
	if err != nil {
		t.Error("Unexpected error received", err)
	}
	expctSplRatedCdr := &StoredCdr{CgrId: storCdr.CgrId, TOR: utils.VOICE, AccId: "dsafdsaf", CdrHost: "192.168.1.1", CdrSource: utils.UNIT_TEST, ReqType: utils.META_RATED,
		Direction: "*out", Tenant: "cgrates.org", Category: "call", Account: "1001", Subject: "1001", Destination: "1002",
		SetupTime: time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC), Pdd: time.Duration(200) * time.Millisecond, AnswerTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC),
		Usage: time.Duration(10) * time.Second, Supplier: "suppl1", ExtraFields: map[string]string{"field_extr1": "val_extr1", "field_extr2": "valextr2"},
		MediationRunId: "sample_run1", Rated: false, Cost: 1.01}
	if !reflect.DeepEqual(expctSplRatedCdr, rtSampleCdrOut) {
		t.Errorf("Expected: %v, received: %v", expctSplRatedCdr, rtSampleCdrOut)
	}
}

func TestStoredCdrForkCdrStaticVals(t *testing.T) {
	storCdr := StoredCdr{CgrId: utils.Sha1("dsafdsaf", time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC).String()), OrderId: 123, TOR: utils.VOICE, AccId: "dsafdsaf",
		CdrHost: "192.168.1.1", CdrSource: utils.UNIT_TEST, ReqType: utils.META_RATED, Direction: "*out",
		Tenant: "cgrates.org", Category: "call", Account: "1001", Subject: "1001", Destination: "1002",
		SetupTime: time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC), AnswerTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC), MediationRunId: utils.DEFAULT_RUNID,
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
	rsrPdd, _ := utils.NewRSRField("^3")
	rsrStRated, _ := utils.NewRSRField("^true")
	rsrStCost, _ := utils.NewRSRField("^1.2")
	rtCdrOut2, err := storCdr.ForkCdr("wholesale_run", rsrStPostpaid, rsrStIn, rsrStCgr, rsrStPC, rsrStFA, rsrStFS, &utils.RSRField{Id: utils.DESTINATION},
		rsrStST, rsrPdd, rsrStAT, rsrStDur, rsrStSuppl, rsrStDCause, rsrStRated, rsrStCost, []*utils.RSRField{}, true, "")
	if err != nil {
		t.Error("Unexpected error received", err)
	}
	expctRatedCdr2 := &StoredCdr{CgrId: storCdr.CgrId, TOR: utils.VOICE, AccId: "dsafdsaf", CdrHost: "192.168.1.1", CdrSource: utils.UNIT_TEST, ReqType: utils.META_POSTPAID,
		Direction: "*in", Tenant: "cgrates.com", Category: "premium_call", Account: "first_account", Subject: "first_subject", Destination: "1002",
		SetupTime:  time.Date(2013, 12, 7, 8, 42, 24, 0, time.UTC),
		AnswerTime: time.Date(2013, 12, 7, 8, 42, 26, 0, time.UTC), Usage: time.Duration(12) * time.Second, Pdd: time.Duration(3) * time.Second,
		Supplier: "supplier1", DisconnectCause: "HANGUP_COMPLETE", Rated: true, Cost: 1.2,
		ExtraFields: map[string]string{}, MediationRunId: "wholesale_run"}
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

func TestStoredCdrForkCdrFromMetaDefaults(t *testing.T) {
	storCdr := StoredCdr{CgrId: utils.Sha1("dsafdsaf", time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC).String()), OrderId: 123, TOR: utils.VOICE, AccId: "dsafdsaf",
		CdrHost: "192.168.1.1", CdrSource: utils.UNIT_TEST, ReqType: utils.META_RATED, Direction: "*out", Tenant: "cgrates.org",
		Category: "call", Account: "1001", Subject: "1001", Destination: "1002",
		SetupTime: time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC), AnswerTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC), MediationRunId: utils.DEFAULT_RUNID,
		Usage: time.Duration(10) * time.Second, Pdd: time.Duration(4) * time.Second, Supplier: "SUPPL3",
		ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"}, Cost: 1.01,
	}
	expctCdr := &StoredCdr{CgrId: storCdr.CgrId, TOR: utils.VOICE, AccId: "dsafdsaf", CdrHost: "192.168.1.1", CdrSource: utils.UNIT_TEST, ReqType: utils.META_RATED,
		Direction: "*out", Tenant: "cgrates.org", Category: "call", Account: "1001", Subject: "1001", Destination: "1002",
		SetupTime: time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC), AnswerTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC),
		Usage: time.Duration(10) * time.Second, Pdd: time.Duration(4) * time.Second, Supplier: "SUPPL3", Cost: 1.01,
		ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"}, MediationRunId: "wholesale_run"}
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

func TestStoredCdrAsExternalCdr(t *testing.T) {
	storCdr := StoredCdr{CgrId: utils.Sha1("dsafdsaf", time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC).String()), OrderId: 123, TOR: utils.VOICE,
		AccId: "dsafdsaf", CdrHost: "192.168.1.1", CdrSource: utils.UNIT_TEST, ReqType: utils.META_RATED, Direction: "*out",
		Tenant: "cgrates.org", Category: "call", Account: "1001", Subject: "1001", Destination: "1002",
		SetupTime: time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC), AnswerTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC), MediationRunId: utils.DEFAULT_RUNID,
		Usage: time.Duration(10), Pdd: time.Duration(7) * time.Second, Supplier: "SUPPL1", ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
		Cost: 1.01, RatedAccount: "dan", RatedSubject: "dans",
	}
	expectOutCdr := &ExternalCdr{CgrId: utils.Sha1("dsafdsaf", time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC).String()), OrderId: 123, TOR: utils.VOICE,
		AccId: "dsafdsaf", CdrHost: "192.168.1.1", CdrSource: utils.UNIT_TEST, ReqType: utils.META_RATED, Direction: "*out",
		Tenant: "cgrates.org", Category: "call", Account: "1001", Subject: "1001", Destination: "1002",
		SetupTime: "2013-11-07T08:42:20Z", AnswerTime: "2013-11-07T08:42:26Z", MediationRunId: utils.DEFAULT_RUNID,
		Usage: "0.00000001", Pdd: "7", Supplier: "SUPPL1", ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
		Cost: 1.01, RatedAccount: "dan", RatedSubject: "dans",
	}
	if cdrOut := storCdr.AsExternalCdr(); !reflect.DeepEqual(expectOutCdr, cdrOut) {
		t.Errorf("Expected: %+v, received: %+v", expectOutCdr, cdrOut)
	}
}

func TestStoredCdrEventFields(t *testing.T) {
	cdr := &StoredCdr{CgrId: utils.Sha1("dsafdsaf", time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC).String()), OrderId: 123, TOR: utils.VOICE, AccId: "dsafdsaf",
		CdrHost: "192.168.1.1", CdrSource: "test", ReqType: utils.META_RATED, Direction: "*out", Tenant: "cgrates.org", Category: "call", Account: "dan", Subject: "dans",
		Destination: "1002", SetupTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC), AnswerTime: time.Date(2013, 11, 7, 8, 42, 27, 0, time.UTC),
		MediationRunId: utils.DEFAULT_RUNID, Usage: time.Duration(10) * time.Second, Supplier: "suppl1",
		ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"}, Cost: 1.01, RatedAccount: "dan", RatedSubject: "dan"}

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
	if res := cdr.GetOriginatorIP(utils.META_DEFAULT); res != cdr.CdrHost {
		t.Error("Received: ", res)
	}
	if extraFlds := cdr.GetExtraFields(); !reflect.DeepEqual(cdr.ExtraFields, extraFlds) {
		t.Error("Received: ", extraFlds)
	}
}

func TesUsageReqAsStoredCdr(t *testing.T) {
	setupReq := &UsageRecord{TOR: utils.VOICE, ReqType: utils.META_RATED, Direction: "*out", Tenant: "cgrates.org", Category: "call",
		Account: "1001", Subject: "1001", Destination: "1002",
		SetupTime: "2013-11-07T08:42:20Z", AnswerTime: "2013-11-07T08:42:26Z", Usage: "0.00000001",
	}
	eStorCdr := &StoredCdr{TOR: utils.VOICE, ReqType: utils.META_RATED, Direction: "*out",
		Tenant: "cgrates.org", Category: "call", Account: "1001", Subject: "1001", Destination: "1002",
		SetupTime: time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC), AnswerTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC), Usage: time.Duration(10)}
	if storedCdr, err := setupReq.AsStoredCdr(""); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eStorCdr, storedCdr) {
		t.Errorf("Expected: %+v, received: %+v", eStorCdr, storedCdr)
	}
}

func TestUsageReqAsCD(t *testing.T) {
	req := &UsageRecord{TOR: utils.VOICE, ReqType: utils.META_RATED, Direction: "*out", Tenant: "cgrates.org", Category: "call",
		Account: "1001", Subject: "1001", Destination: "1002",
		SetupTime: "2013-11-07T08:42:20Z", AnswerTime: "2013-11-07T08:42:26Z", Usage: "0.00000001",
	}
	eCD := &CallDescriptor{TOR: req.TOR, Direction: req.Direction,
		Tenant: req.Tenant, Category: req.Category, Account: req.Account, Subject: req.Subject, Destination: req.Destination,
		TimeStart: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC), TimeEnd: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC).Add(time.Duration(10))}
	if cd, err := req.AsCallDescriptor(""); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCD, cd) {
		t.Errorf("Expected: %+v, received: %+v", eCD, cd)
	}
}
