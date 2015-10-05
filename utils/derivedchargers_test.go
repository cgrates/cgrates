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

package utils

import (
	"reflect"
	"testing"
)

func TestAppendDerivedChargers(t *testing.T) {
	var err error
	dcs := make(DerivedChargers, 0)
	if _, err := dcs.Append(&DerivedCharger{RunId: DEFAULT_RUNID}); err == nil {
		t.Error("Failed to detect using of the default runid")
	}
	if dcs, err = dcs.Append(&DerivedCharger{RunId: "FIRST_RUNID"}); err != nil {
		t.Error("Failed to add runid")
	} else if len(dcs) != 1 {
		t.Error("Unexpected number of items inside DerivedChargers configuration", len(dcs))
	}
	if dcs, err = dcs.Append(&DerivedCharger{RunId: "SECOND_RUNID"}); err != nil {
		t.Error("Failed to add runid")
	} else if len(dcs) != 2 {
		t.Error("Unexpected number of items inside DerivedChargers configuration", len(dcs))
	}
	if _, err := dcs.Append(&DerivedCharger{RunId: "SECOND_RUNID"}); err == nil {
		t.Error("Failed to detect duplicate runid")
	}
}

func TestNewDerivedCharger(t *testing.T) {
	edc1 := &DerivedCharger{
		RunId:                "test1",
		RunFilters:           "",
		ReqTypeField:         "reqtype1",
		DirectionField:       "direction1",
		TenantField:          "tenant1",
		CategoryField:        "tor1",
		AccountField:         "account1",
		SubjectField:         "subject1",
		DestinationField:     "destination1",
		SetupTimeField:       "setuptime1",
		PddField:             "pdd1",
		AnswerTimeField:      "answertime1",
		UsageField:           "duration1",
		SupplierField:        "supplier1",
		DisconnectCauseField: "NORMAL_CLEARING",
		RatedField:           "rated1",
		CostField:            "cost1",
	}
	if dc1, err := NewDerivedCharger("test1", "", "reqtype1", "direction1", "tenant1", "tor1", "account1", "subject1", "destination1",
		"setuptime1", "pdd1", "answertime1", "duration1", "supplier1", "NORMAL_CLEARING", "rated1", "cost1"); err != nil {
		t.Error("Unexpected error", err.Error)
	} else if !reflect.DeepEqual(edc1, dc1) {
		t.Errorf("Expecting: %v, received: %v", edc1, dc1)
	}
	edc2 := &DerivedCharger{
		RunId:                "test2",
		RunFilters:           "^cdr_source/tdm_cdrs/",
		ReqTypeField:         "~reqtype2:s/sip:(.+)/$1/",
		DirectionField:       "~direction2:s/sip:(.+)/$1/",
		TenantField:          "~tenant2:s/sip:(.+)/$1/",
		CategoryField:        "~tor2:s/sip:(.+)/$1/",
		AccountField:         "~account2:s/sip:(.+)/$1/",
		SubjectField:         "~subject2:s/sip:(.+)/$1/",
		DestinationField:     "~destination2:s/sip:(.+)/$1/",
		SetupTimeField:       "~setuptime2:s/sip:(.+)/$1/",
		PddField:             "~pdd:s/sip:(.+)/$1/",
		AnswerTimeField:      "~answertime2:s/sip:(.+)/$1/",
		UsageField:           "~duration2:s/sip:(.+)/$1/",
		SupplierField:        "~supplier2:s/(.+)/$1/",
		DisconnectCauseField: "~cgr_disconnect:s/(.+)/$1/",
		CostField:            "~cgr_cost:s/(.+)/$1/",
		RatedField:           "~cgr_rated:s/(.+)/$1/",
	}
	edc2.rsrRunFilters, _ = ParseRSRFields("^cdr_source/tdm_cdrs/", INFIELD_SEP)
	edc2.rsrReqTypeField, _ = NewRSRField("~reqtype2:s/sip:(.+)/$1/")
	edc2.rsrDirectionField, _ = NewRSRField("~direction2:s/sip:(.+)/$1/")
	edc2.rsrTenantField, _ = NewRSRField("~tenant2:s/sip:(.+)/$1/")
	edc2.rsrCategoryField, _ = NewRSRField("~tor2:s/sip:(.+)/$1/")
	edc2.rsrAccountField, _ = NewRSRField("~account2:s/sip:(.+)/$1/")
	edc2.rsrSubjectField, _ = NewRSRField("~subject2:s/sip:(.+)/$1/")
	edc2.rsrDestinationField, _ = NewRSRField("~destination2:s/sip:(.+)/$1/")
	edc2.rsrSetupTimeField, _ = NewRSRField("~setuptime2:s/sip:(.+)/$1/")
	edc2.rsrPddField, _ = NewRSRField("~pdd:s/sip:(.+)/$1/")
	edc2.rsrAnswerTimeField, _ = NewRSRField("~answertime2:s/sip:(.+)/$1/")
	edc2.rsrUsageField, _ = NewRSRField("~duration2:s/sip:(.+)/$1/")
	edc2.rsrSupplierField, _ = NewRSRField("~supplier2:s/(.+)/$1/")
	edc2.rsrDisconnectCauseField, _ = NewRSRField("~cgr_disconnect:s/(.+)/$1/")
	edc2.rsrCostField, _ = NewRSRField("~cgr_cost:s/(.+)/$1/")
	edc2.rsrRatedField, _ = NewRSRField("~cgr_rated:s/(.+)/$1/")
	if dc2, err := NewDerivedCharger("test2",
		"^cdr_source/tdm_cdrs/",
		"~reqtype2:s/sip:(.+)/$1/",
		"~direction2:s/sip:(.+)/$1/",
		"~tenant2:s/sip:(.+)/$1/",
		"~tor2:s/sip:(.+)/$1/",
		"~account2:s/sip:(.+)/$1/",
		"~subject2:s/sip:(.+)/$1/",
		"~destination2:s/sip:(.+)/$1/",
		"~setuptime2:s/sip:(.+)/$1/",
		"~pdd:s/sip:(.+)/$1/",
		"~answertime2:s/sip:(.+)/$1/",
		"~duration2:s/sip:(.+)/$1/",
		"~supplier2:s/(.+)/$1/",
		"~cgr_disconnect:s/(.+)/$1/",
		"~cgr_rated:s/(.+)/$1/",
		"~cgr_cost:s/(.+)/$1/"); err != nil {
		t.Error("Unexpected error", err)
	} else if !reflect.DeepEqual(edc2, dc2) {
		t.Errorf("Expecting: %v, received: %v", edc2, dc2)
	}
}

func TestDerivedChargersKey(t *testing.T) {
	if dcKey := DerivedChargersKey("*out", "cgrates.org", "call", "dan", "dan"); dcKey != "*out:cgrates.org:call:dan:dan" {
		t.Error("Unexpected derived chargers key: ", dcKey)
	}
}

func TestAppendDefaultRun(t *testing.T) {
	var dc1 DerivedChargers
	dcDf := &DerivedCharger{RunId: DEFAULT_RUNID, RunFilters: "", ReqTypeField: META_DEFAULT, DirectionField: META_DEFAULT,
		TenantField: META_DEFAULT, CategoryField: META_DEFAULT, AccountField: META_DEFAULT, SubjectField: META_DEFAULT,
		DestinationField: META_DEFAULT, SetupTimeField: META_DEFAULT, PddField: META_DEFAULT, AnswerTimeField: META_DEFAULT, UsageField: META_DEFAULT, SupplierField: META_DEFAULT,
		DisconnectCauseField: META_DEFAULT, CostField: META_DEFAULT, RatedField: META_DEFAULT}
	eDc1 := DerivedChargers{dcDf}
	if dc1, _ = dc1.AppendDefaultRun(); !reflect.DeepEqual(dc1, eDc1) {
		t.Errorf("Expecting: %+v, received: %+v", eDc1[0], dc1[0])
	}
	dc2 := DerivedChargers{
		&DerivedCharger{RunId: "extra1", RunFilters: "", ReqTypeField: "reqtype2", DirectionField: META_DEFAULT, TenantField: META_DEFAULT, CategoryField: META_DEFAULT,
			AccountField: "rif", SubjectField: "rif", DestinationField: META_DEFAULT, SetupTimeField: META_DEFAULT, PddField: META_DEFAULT, AnswerTimeField: META_DEFAULT, UsageField: META_DEFAULT,
			DisconnectCauseField: META_DEFAULT},
		&DerivedCharger{RunId: "extra2", ReqTypeField: META_DEFAULT, DirectionField: META_DEFAULT, TenantField: META_DEFAULT, CategoryField: META_DEFAULT,
			AccountField: "ivo", SubjectField: "ivo", DestinationField: META_DEFAULT, SetupTimeField: META_DEFAULT, PddField: META_DEFAULT, AnswerTimeField: META_DEFAULT,
			UsageField: META_DEFAULT, SupplierField: META_DEFAULT, DisconnectCauseField: META_DEFAULT},
	}
	eDc2 := append(dc2, dcDf)
	if dc2, _ = dc2.AppendDefaultRun(); !reflect.DeepEqual(dc2, eDc2) {
		t.Errorf("Expecting: %+v, received: %+v", eDc2, dc2)
	}
}
