/*
Real-time Charging System for Telecom & ISP environments
Copyright (C) 2012-2014 ITsysCOM GmbH

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

package utils

import (
	"reflect"
	"testing"
	"time"
)

func TestStoredCdrInterfaces(t *testing.T) {
	storedCdr := new(StoredCdr)
	var _ RawCdr = storedCdr
}

func TestFieldAsString(t *testing.T) {
	cdr := StoredCdr{CgrId: Sha1("dsafdsaf", time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC).String()), OrderId: 123, TOR: VOICE, AccId: "dsafdsaf", CdrHost: "192.168.1.1", CdrSource: "test", ReqType: "rated", Direction: "*out", Tenant: "cgrates.org",
		Category: "call", Account: "1001", Subject: "1001", Destination: "1002", SetupTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC), AnswerTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC), MediationRunId: DEFAULT_RUNID,
		Usage: time.Duration(10) * time.Second, ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"}, Cost: 1.01, RatedAccount: "dan", RatedSubject: "dans",
	}
	if cdr.FieldAsString(&RSRField{Id: CGRID}) != cdr.CgrId ||
		cdr.FieldAsString(&RSRField{Id: ORDERID}) != "123" ||
		cdr.FieldAsString(&RSRField{Id: TOR}) != VOICE ||
		cdr.FieldAsString(&RSRField{Id: ACCID}) != cdr.AccId ||
		cdr.FieldAsString(&RSRField{Id: CDRHOST}) != cdr.CdrHost ||
		cdr.FieldAsString(&RSRField{Id: CDRSOURCE}) != cdr.CdrSource ||
		cdr.FieldAsString(&RSRField{Id: REQTYPE}) != cdr.ReqType ||
		cdr.FieldAsString(&RSRField{Id: DIRECTION}) != cdr.Direction ||
		cdr.FieldAsString(&RSRField{Id: CATEGORY}) != cdr.Category ||
		cdr.FieldAsString(&RSRField{Id: ACCOUNT}) != cdr.Account ||
		cdr.FieldAsString(&RSRField{Id: SUBJECT}) != cdr.Subject ||
		cdr.FieldAsString(&RSRField{Id: DESTINATION}) != cdr.Destination ||
		cdr.FieldAsString(&RSRField{Id: SETUP_TIME}) != cdr.SetupTime.String() ||
		cdr.FieldAsString(&RSRField{Id: ANSWER_TIME}) != cdr.AnswerTime.String() ||
		cdr.FieldAsString(&RSRField{Id: USAGE}) != "10" ||
		cdr.FieldAsString(&RSRField{Id: MEDI_RUNID}) != cdr.MediationRunId ||
		cdr.FieldAsString(&RSRField{Id: COST}) != "1.01" ||
		cdr.FieldAsString(&RSRField{Id: RATED_ACCOUNT}) != "dan" ||
		cdr.FieldAsString(&RSRField{Id: RATED_SUBJECT}) != "dans" ||
		cdr.FieldAsString(&RSRField{Id: "field_extr1"}) != cdr.ExtraFields["field_extr1"] ||
		cdr.FieldAsString(&RSRField{Id: "fieldextr2"}) != cdr.ExtraFields["fieldextr2"] ||
		cdr.FieldAsString(&RSRField{Id: "dummy_field"}) != "" {
		t.Error("Unexpected filed value received",
			cdr.FieldAsString(&RSRField{Id: CGRID}) != cdr.CgrId,
			cdr.FieldAsString(&RSRField{Id: ORDERID}) != "123",
			cdr.FieldAsString(&RSRField{Id: TOR}) != VOICE,
			cdr.FieldAsString(&RSRField{Id: ACCID}) != cdr.AccId,
			cdr.FieldAsString(&RSRField{Id: CDRHOST}) != cdr.CdrHost,
			cdr.FieldAsString(&RSRField{Id: CDRSOURCE}) != cdr.CdrSource,
			cdr.FieldAsString(&RSRField{Id: REQTYPE}) != cdr.ReqType,
			cdr.FieldAsString(&RSRField{Id: DIRECTION}) != cdr.Direction,
			cdr.FieldAsString(&RSRField{Id: CATEGORY}) != cdr.Category,
			cdr.FieldAsString(&RSRField{Id: ACCOUNT}) != cdr.Account,
			cdr.FieldAsString(&RSRField{Id: SUBJECT}) != cdr.Subject,
			cdr.FieldAsString(&RSRField{Id: DESTINATION}) != cdr.Destination,
			cdr.FieldAsString(&RSRField{Id: SETUP_TIME}) != cdr.SetupTime.String(),
			cdr.FieldAsString(&RSRField{Id: ANSWER_TIME}) != cdr.AnswerTime.String(),
			cdr.FieldAsString(&RSRField{Id: USAGE}) != "10",
			cdr.FieldAsString(&RSRField{Id: MEDI_RUNID}) != cdr.MediationRunId,
			cdr.FieldAsString(&RSRField{Id: RATED_ACCOUNT}) != "dan",
			cdr.FieldAsString(&RSRField{Id: RATED_SUBJECT}) != "dans",
			cdr.FieldAsString(&RSRField{Id: COST}) != "1.01",
			cdr.FieldAsString(&RSRField{Id: "field_extr1"}) != cdr.ExtraFields["field_extr1"],
			cdr.FieldAsString(&RSRField{Id: "fieldextr2"}) != cdr.ExtraFields["fieldextr2"],
			cdr.FieldAsString(&RSRField{Id: "dummy_field"}) != "")
	}
}

func TestPassesFieldFilter(t *testing.T) {
	cdr := &StoredCdr{CgrId: Sha1("dsafdsaf", time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC).String()), OrderId: 123, TOR: VOICE, AccId: "dsafdsaf",
		CdrHost: "192.168.1.1", CdrSource: "test", ReqType: "rated", Direction: "*out", Tenant: "cgrates.org",
		Category: "call", Account: "1001", Subject: "1001", Destination: "1002", SetupTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC),
		AnswerTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC), MediationRunId: DEFAULT_RUNID,
		Usage: time.Duration(10) * time.Second, ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"}, Cost: 1.01,
	}
	if pass, _ := cdr.PassesFieldFilter(nil); !pass {
		t.Error("Not passing filter")
	}
	acntPrefxFltr, _ := NewRSRField(`~account:s/(.+)/1001/`)
	if pass, _ := cdr.PassesFieldFilter(acntPrefxFltr); !pass {
		t.Error("Not passing filter")
	}
	acntPrefxFltr, _ = NewRSRField(`~account:s/^(10)\d\d$/10/`)
	if pass, _ := cdr.PassesFieldFilter(acntPrefxFltr); !pass {
		t.Error("Not passing valid filter")
	}
	acntPrefxFltr, _ = NewRSRField(`~account:s/^\d(10)\d$/10/`)
	if pass, _ := cdr.PassesFieldFilter(acntPrefxFltr); pass {
		t.Error("Passing filter")
	}
	acntPrefxFltr, _ = NewRSRField(`~account:s/^(10)\d\d$/010/`)
	if pass, _ := cdr.PassesFieldFilter(acntPrefxFltr); pass {
		t.Error("Passing filter")
	}
	acntPrefxFltr, _ = NewRSRField(`~account:s/^1010$/1010/`)
	if pass, _ := cdr.PassesFieldFilter(acntPrefxFltr); pass {
		t.Error("Passing filter")
	}
	torFltr, _ := NewRSRField(`^tor/*voice/`)
	if pass, _ := cdr.PassesFieldFilter(torFltr); !pass {
		t.Error("Not passing filter")
	}
	torFltr, _ = NewRSRField(`^tor/*data/`)
	if pass, _ := cdr.PassesFieldFilter(torFltr); pass {
		t.Error("Passing filter")
	}
	inexistentFieldFltr, _ := NewRSRField(`^fakefield/fakevalue/`)
	if pass, _ := cdr.PassesFieldFilter(inexistentFieldFltr); pass {
		t.Error("Passing filter")
	}
}

func TestPassesFieldFilterDn1(t *testing.T) {
	cdr := &StoredCdr{CgrId: Sha1("dsafdsaf", time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC).String()), Account: "futurem0005",
		ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"}, Cost: 1.01,
	}
	acntPrefxFltr, _ := NewRSRField(`~account:s/^\w+[shmp]\d{4}$//`)
	if pass, _ := cdr.PassesFieldFilter(acntPrefxFltr); !pass {
		t.Error("Not passing valid filter")
	}

	cdr = &StoredCdr{CgrId: Sha1("dsafdsaf", time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC).String()), Account: "futurem00005",
		ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"}, Cost: 1.01,
	}
	if pass, _ := cdr.PassesFieldFilter(acntPrefxFltr); pass {
		t.Error("Should not pass filter")
	}
	cdr = &StoredCdr{CgrId: Sha1("dsafdsaf", time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC).String()), Account: "0402129281",
		ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"}, Cost: 1.01,
	}
	acntPrefxFltr, _ = NewRSRField(`~account:s/^0\d{9}$//`)
	if pass, _ := cdr.PassesFieldFilter(acntPrefxFltr); !pass {
		t.Error("Not passing valid filter")
	}
	acntPrefxFltr, _ = NewRSRField(`~account:s/^0(\d{9})$/placeholder/`)
	if pass, _ := cdr.PassesFieldFilter(acntPrefxFltr); pass {
		t.Error("Should not pass filter")
	}
	cdr = &StoredCdr{CgrId: Sha1("dsafdsaf", time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC).String()), Account: "04021292812",
		ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"}, Cost: 1.01,
	}
	if pass, _ := cdr.PassesFieldFilter(acntPrefxFltr); pass {
		t.Error("Should not pass filter")
	}
	cdr = &StoredCdr{CgrId: Sha1("dsafdsaf", time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC).String()), Account: "0162447222",
		ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"}, Cost: 1.01,
	}
	if acntPrefxFltr, err := NewRSRField(`~account:s/^0\d{9}$//`); err != nil {
		t.Error("Unexpected parse error", err)
	} else if acntPrefxFltr == nil {
		t.Error("Failed parsing rule")
	} else if pass, _ := cdr.PassesFieldFilter(acntPrefxFltr); !pass {
		t.Error("Not passing valid filter")
	}
	if acntPrefxFltr, err := NewRSRField(`~account:s/^\w+[shmp]\d{4}$//`); err != nil {
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
		t.Error("Unexpected cost after multiply: %v", cdr.Cost)
	}
	cdr = StoredCdr{Cost: 1.01}
	if cdr.CostMultiply(1000, 0); cdr.Cost != 1010 {
		t.Error("Unexpected cost after multiply: %v", cdr.Cost)
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
	if cdr.FormatUsage(SECONDS) != "10" {
		t.Error("Wrong usage format: ", cdr.FormatUsage(SECONDS))
	}
	if cdr.FormatUsage("default") != "10" {
		t.Error("Wrong usage format: ", cdr.FormatUsage("default"))
	}
	cdr = StoredCdr{TOR: DATA, Usage: time.Duration(1640113000000000)}
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
	storCdr := StoredCdr{CgrId: Sha1("dsafdsaf", time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC).String()), OrderId: 123, TOR: VOICE, AccId: "dsafdsaf", CdrHost: "192.168.1.1",
		CdrSource: UNIT_TEST, ReqType: "rated", Direction: "*out", Tenant: "cgrates.org", Category: "call", Account: "1001", Subject: "1001", Destination: "1002",
		SetupTime: time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC), AnswerTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC), MediationRunId: DEFAULT_RUNID,
		Usage: time.Duration(10) * time.Second, ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"}, RatedSubject: "dans", Cost: 1.01,
	}
	cdrForm := storCdr.AsHttpForm()
	if cdrForm.Get(TOR) != VOICE {
		t.Errorf("Expected: %s, received: %s", VOICE, cdrForm.Get(TOR))
	}
	if cdrForm.Get(ACCID) != "dsafdsaf" {
		t.Errorf("Expected: %s, received: %s", "dsafdsaf", cdrForm.Get(ACCID))
	}
	if cdrForm.Get(CDRHOST) != "192.168.1.1" {
		t.Errorf("Expected: %s, received: %s", "192.168.1.1", cdrForm.Get(CDRHOST))
	}
	if cdrForm.Get(CDRSOURCE) != UNIT_TEST {
		t.Errorf("Expected: %s, received: %s", UNIT_TEST, cdrForm.Get(CDRSOURCE))
	}
	if cdrForm.Get(REQTYPE) != "rated" {
		t.Errorf("Expected: %s, received: %s", "rated", cdrForm.Get(REQTYPE))
	}
	if cdrForm.Get(DIRECTION) != "*out" {
		t.Errorf("Expected: %s, received: %s", "*out", cdrForm.Get(DIRECTION))
	}
	if cdrForm.Get(TENANT) != "cgrates.org" {
		t.Errorf("Expected: %s, received: %s", "cgrates.org", cdrForm.Get(TENANT))
	}
	if cdrForm.Get(CATEGORY) != "call" {
		t.Errorf("Expected: %s, received: %s", "call", cdrForm.Get(CATEGORY))
	}
	if cdrForm.Get(ACCOUNT) != "1001" {
		t.Errorf("Expected: %s, received: %s", "1001", cdrForm.Get(ACCOUNT))
	}
	if cdrForm.Get(SUBJECT) != "1001" {
		t.Errorf("Expected: %s, received: %s", "1001", cdrForm.Get(SUBJECT))
	}
	if cdrForm.Get(DESTINATION) != "1002" {
		t.Errorf("Expected: %s, received: %s", "1002", cdrForm.Get(DESTINATION))
	}
	if cdrForm.Get(SETUP_TIME) != "2013-11-07 08:42:20 +0000 UTC" {
		t.Errorf("Expected: %s, received: %s", "2013-11-07 08:42:20 +0000 UTC", cdrForm.Get(SETUP_TIME))
	}
	if cdrForm.Get(ANSWER_TIME) != "2013-11-07 08:42:26 +0000 UTC" {
		t.Errorf("Expected: %s, received: %s", "2013-11-07 08:42:26 +0000 UTC", cdrForm.Get(ANSWER_TIME))
	}
	if cdrForm.Get(USAGE) != "10" {
		t.Errorf("Expected: %s, received: %s", "10", cdrForm.Get(USAGE))
	}
	if cdrForm.Get("field_extr1") != "val_extr1" {
		t.Errorf("Expected: %s, received: %s", "val_extr1", cdrForm.Get("field_extr1"))
	}
	if cdrForm.Get("fieldextr2") != "valextr2" {
		t.Errorf("Expected: %s, received: %s", "valextr2", cdrForm.Get("fieldextr2"))
	}
}

func TestStoredCdrForkCdr(t *testing.T) {
	storCdr := StoredCdr{CgrId: Sha1("dsafdsaf", time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC).String()), OrderId: 123, TOR: VOICE, AccId: "dsafdsaf", CdrHost: "192.168.1.1",
		CdrSource: UNIT_TEST, ReqType: "rated", Direction: "*out", Tenant: "cgrates.org", Category: "call", Account: "1001", Subject: "1001", Destination: "1002",
		SetupTime: time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC), AnswerTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC), MediationRunId: DEFAULT_RUNID,
		Usage: time.Duration(10) * time.Second, ExtraFields: map[string]string{"field_extr1": "val_extr1", "field_extr2": "valextr2"}, Cost: 1.01, RatedSubject: "dans",
	}
	rtSampleCdrOut, err := storCdr.ForkCdr("sample_run1", &RSRField{Id: REQTYPE}, &RSRField{Id: DIRECTION}, &RSRField{Id: TENANT}, &RSRField{Id: CATEGORY},
		&RSRField{Id: ACCOUNT}, &RSRField{Id: SUBJECT}, &RSRField{Id: DESTINATION}, &RSRField{Id: SETUP_TIME}, &RSRField{Id: ANSWER_TIME}, &RSRField{Id: USAGE},
		[]*RSRField{&RSRField{Id: "field_extr1"}, &RSRField{Id: "field_extr2"}}, true)
	if err != nil {
		t.Error("Unexpected error received", err)
	}
	expctSplRatedCdr := &StoredCdr{CgrId: storCdr.CgrId, TOR: VOICE, AccId: "dsafdsaf", CdrHost: "192.168.1.1", CdrSource: UNIT_TEST, ReqType: "rated",
		Direction: "*out", Tenant: "cgrates.org", Category: "call", Account: "1001", Subject: "1001", Destination: "1002",
		SetupTime: time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC), AnswerTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC),
		Usage: time.Duration(10) * time.Second, ExtraFields: map[string]string{"field_extr1": "val_extr1", "field_extr2": "valextr2"}, MediationRunId: "sample_run1", Cost: -1}
	if !reflect.DeepEqual(expctSplRatedCdr, rtSampleCdrOut) {
		t.Errorf("Expected: %v, received: %v", expctSplRatedCdr, rtSampleCdrOut)
	}
}

func TestStoredCdrForkCdrStaticVals(t *testing.T) {
	storCdr := StoredCdr{CgrId: Sha1("dsafdsaf", time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC).String()), OrderId: 123, TOR: VOICE, AccId: "dsafdsaf", CdrHost: "192.168.1.1",
		CdrSource: UNIT_TEST, ReqType: "rated", Direction: "*out", Tenant: "cgrates.org", Category: "call", Account: "1001", Subject: "1001", Destination: "1002",
		SetupTime: time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC), AnswerTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC), MediationRunId: DEFAULT_RUNID,
		Usage: time.Duration(10) * time.Second, ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"}, Cost: 1.01,
	}
	rsrStPostpaid, _ := NewRSRField("^postpaid")
	rsrStIn, _ := NewRSRField("^*in")
	rsrStCgr, _ := NewRSRField("^cgrates.com")
	rsrStPC, _ := NewRSRField("^premium_call")
	rsrStFA, _ := NewRSRField("^first_account")
	rsrStFS, _ := NewRSRField("^first_subject")
	rsrStST, _ := NewRSRField("^2013-12-07T08:42:24Z")
	rsrStAT, _ := NewRSRField("^2013-12-07T08:42:26Z")
	rsrStDur, _ := NewRSRField("^12s")
	rtCdrOut2, err := storCdr.ForkCdr("wholesale_run", rsrStPostpaid, rsrStIn, rsrStCgr, rsrStPC, rsrStFA, rsrStFS, &RSRField{Id: "destination"}, rsrStST, rsrStAT, rsrStDur,
		[]*RSRField{}, true)

	if err != nil {
		t.Error("Unexpected error received", err)
	}
	expctRatedCdr2 := &StoredCdr{CgrId: storCdr.CgrId, TOR: VOICE, AccId: "dsafdsaf", CdrHost: "192.168.1.1", CdrSource: UNIT_TEST, ReqType: "postpaid",
		Direction: "*in", Tenant: "cgrates.com", Category: "premium_call", Account: "first_account", Subject: "first_subject", Destination: "1002",
		SetupTime:  time.Date(2013, 12, 7, 8, 42, 24, 0, time.UTC),
		AnswerTime: time.Date(2013, 12, 7, 8, 42, 26, 0, time.UTC), Usage: time.Duration(12) * time.Second,
		ExtraFields: map[string]string{}, MediationRunId: "wholesale_run", Cost: -1}
	if !reflect.DeepEqual(rtCdrOut2, expctRatedCdr2) {
		t.Errorf("Received: %v, expected: %v", rtCdrOut2, expctRatedCdr2)
	}
	_, err = storCdr.ForkCdr("wholesale_run", &RSRField{Id: "dummy_header"}, &RSRField{Id: "direction"}, &RSRField{Id: "tenant"}, &RSRField{Id: "tor"}, &RSRField{Id: "account"},
		&RSRField{Id: "subject"}, &RSRField{Id: "destination"}, &RSRField{Id: "setup_time"}, &RSRField{Id: "answer_time"}, &RSRField{Id: "duration"},
		[]*RSRField{}, true)
	if err == nil {
		t.Error("Failed to detect missing header")
	}
}

func TestStoredCdrForkCdrFromMetaDefaults(t *testing.T) {
	storCdr := StoredCdr{CgrId: Sha1("dsafdsaf", time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC).String()), OrderId: 123, TOR: VOICE, AccId: "dsafdsaf", CdrHost: "192.168.1.1",
		CdrSource: UNIT_TEST, ReqType: "rated", Direction: "*out", Tenant: "cgrates.org", Category: "call", Account: "1001", Subject: "1001", Destination: "1002",
		SetupTime: time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC), AnswerTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC), MediationRunId: DEFAULT_RUNID,
		Usage: time.Duration(10) * time.Second, ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"}, Cost: 1.01,
	}
	expctCdr := &StoredCdr{CgrId: storCdr.CgrId, TOR: VOICE, AccId: "dsafdsaf", CdrHost: "192.168.1.1", CdrSource: UNIT_TEST, ReqType: "rated",
		Direction: "*out", Tenant: "cgrates.org", Category: "call", Account: "1001", Subject: "1001", Destination: "1002",
		SetupTime: time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC), AnswerTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC),
		Usage:       time.Duration(10) * time.Second,
		ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"}, MediationRunId: "wholesale_run", Cost: -1}
	cdrOut, err := storCdr.ForkCdr("wholesale_run", &RSRField{Id: META_DEFAULT}, &RSRField{Id: META_DEFAULT}, &RSRField{Id: META_DEFAULT}, &RSRField{Id: META_DEFAULT},
		&RSRField{Id: META_DEFAULT}, &RSRField{Id: META_DEFAULT}, &RSRField{Id: META_DEFAULT},
		&RSRField{Id: META_DEFAULT}, &RSRField{Id: META_DEFAULT}, &RSRField{Id: META_DEFAULT}, []*RSRField{&RSRField{Id: "field_extr1"}, &RSRField{Id: "fieldextr2"}}, true)
	if err != nil {
		t.Fatal("Unexpected error received", err)
	}

	if !reflect.DeepEqual(expctCdr, cdrOut) {
		t.Errorf("Expected: %v, received: %v", expctCdr, cdrOut)
	}
	// Should also accept nil as defaults
	if cdrOut, err := storCdr.ForkCdr("wholesale_run", nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
		[]*RSRField{&RSRField{Id: "field_extr1"}, &RSRField{Id: "fieldextr2"}}, true); err != nil {
		t.Fatal("Unexpected error received", err)
	} else if !reflect.DeepEqual(expctCdr, cdrOut) {
		t.Errorf("Expected: %v, received: %v", expctCdr, cdrOut)
	}
}

func TestStoredCdrAsCgrCdrOut(t *testing.T) {
	storCdr := StoredCdr{CgrId: Sha1("dsafdsaf", time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC).String()), OrderId: 123, TOR: VOICE, AccId: "dsafdsaf", CdrHost: "192.168.1.1",
		CdrSource: UNIT_TEST, ReqType: "rated", Direction: "*out", Tenant: "cgrates.org", Category: "call", Account: "1001", Subject: "1001", Destination: "1002",
		SetupTime: time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC), AnswerTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC), MediationRunId: DEFAULT_RUNID,
		Usage: time.Duration(10), ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"}, Cost: 1.01, RatedAccount: "dan", RatedSubject: "dans",
	}
	expectOutCdr := &CgrCdrOut{CgrId: Sha1("dsafdsaf", time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC).String()), OrderId: 123, TOR: VOICE, AccId: "dsafdsaf", CdrHost: "192.168.1.1",
		CdrSource: UNIT_TEST, ReqType: "rated", Direction: "*out", Tenant: "cgrates.org", Category: "call", Account: "1001", Subject: "1001", Destination: "1002",
		SetupTime: time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC), AnswerTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC), MediationRunId: DEFAULT_RUNID,
		Usage: 0.00000001, ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"}, Cost: 1.01, RatedAccount: "dan", RatedSubject: "dans",
	}
	if cdrOut := storCdr.AsCgrCdrOut(); !reflect.DeepEqual(expectOutCdr, cdrOut) {
		t.Errorf("Expected: %+v, received: %+v", expectOutCdr, cdrOut)
	}
}
