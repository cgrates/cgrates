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
		RunId:            "test1",
		ReqTypeField:     "reqtype1",
		DirectionField:   "direction1",
		TenantField:      "tenant1",
		CategoryField:    "tor1",
		AccountField:     "account1",
		SubjectField:     "subject1",
		DestinationField: "destination1",
		SetupTimeField:   "setuptime1",
		AnswerTimeField:  "answertime1",
		DurationField:    "duration1",
	}
	if dc1, err := NewDerivedCharger("test1", "reqtype1", "direction1", "tenant1", "tor1", "account1", "subject1", "destination1",
		"setuptime1", "answertime1", "duration1"); err != nil {
		t.Error("Unexpected error", err.Error)
	} else if !reflect.DeepEqual(edc1, dc1) {
		t.Errorf("Expecting: %v, received: %v", edc1, dc1)
	}
	edc2 := &DerivedCharger{
		RunId:            "test2",
		ReqTypeField:     "~reqtype2:s/sip:(.+)/$1/",
		DirectionField:   "~direction2:s/sip:(.+)/$1/",
		TenantField:      "~tenant2:s/sip:(.+)/$1/",
		CategoryField:    "~tor2:s/sip:(.+)/$1/",
		AccountField:     "~account2:s/sip:(.+)/$1/",
		SubjectField:     "~subject2:s/sip:(.+)/$1/",
		DestinationField: "~destination2:s/sip:(.+)/$1/",
		SetupTimeField:   "~setuptime2:s/sip:(.+)/$1/",
		AnswerTimeField:  "~answertime2:s/sip:(.+)/$1/",
		DurationField:    "~duration2:s/sip:(.+)/$1/",
	}
	edc2.rsrReqTypeField, _ = NewRSRField("~reqtype2:s/sip:(.+)/$1/")
	edc2.rsrDirectionField, _ = NewRSRField("~direction2:s/sip:(.+)/$1/")
	edc2.rsrTenantField, _ = NewRSRField("~tenant2:s/sip:(.+)/$1/")
	edc2.rsrCategoryField, _ = NewRSRField("~tor2:s/sip:(.+)/$1/")
	edc2.rsrAccountField, _ = NewRSRField("~account2:s/sip:(.+)/$1/")
	edc2.rsrSubjectField, _ = NewRSRField("~subject2:s/sip:(.+)/$1/")
	edc2.rsrDestinationField, _ = NewRSRField("~destination2:s/sip:(.+)/$1/")
	edc2.rsrSetupTimeField, _ = NewRSRField("~setuptime2:s/sip:(.+)/$1/")
	edc2.rsrAnswerTimeField, _ = NewRSRField("~answertime2:s/sip:(.+)/$1/")
	edc2.rsrDurationField, _ = NewRSRField("~duration2:s/sip:(.+)/$1/")
	if dc2, err := NewDerivedCharger("test2",
		"~reqtype2:s/sip:(.+)/$1/",
		"~direction2:s/sip:(.+)/$1/",
		"~tenant2:s/sip:(.+)/$1/",
		"~tor2:s/sip:(.+)/$1/",
		"~account2:s/sip:(.+)/$1/",
		"~subject2:s/sip:(.+)/$1/",
		"~destination2:s/sip:(.+)/$1/",
		"~setuptime2:s/sip:(.+)/$1/",
		"~answertime2:s/sip:(.+)/$1/",
		"~duration2:s/sip:(.+)/$1/"); err != nil {
		t.Error("Unexpected error", err.Error)
	} else if !reflect.DeepEqual(edc2, dc2) {
		t.Errorf("Expecting: %v, received: %v", edc2, dc2)
	}
}

func TestDerivedChargersKey(t *testing.T) {
	if dcKey := DerivedChargersKey("cgrates.org", "call", "*out", "dan", "dan"); dcKey != "cgrates.org:call:*out:dan:dan" {
		t.Error("Unexpected derived chargers key: ", dcKey)
	}
}

func TestAppendDefaultRun(t *testing.T) {
	var dc1 DerivedChargers
	dcDf := &DerivedCharger{RunId: DEFAULT_RUNID, ReqTypeField: META_DEFAULT, DirectionField: META_DEFAULT,
		TenantField: META_DEFAULT, CategoryField: META_DEFAULT, AccountField: META_DEFAULT, SubjectField: META_DEFAULT,
		DestinationField: META_DEFAULT, SetupTimeField: META_DEFAULT, AnswerTimeField: META_DEFAULT, DurationField: META_DEFAULT}
	eDc1 := DerivedChargers{dcDf}
	if dc1, _ = dc1.AppendDefaultRun(); !reflect.DeepEqual(dc1, eDc1) {
		t.Error("Unexpected result.")
	}
	dc2 := DerivedChargers{
		&DerivedCharger{RunId: "extra1", ReqTypeField: "^prepaid", DirectionField: "*default", TenantField: "*default", CategoryField: "*default",
			AccountField: "rif", SubjectField: "rif", DestinationField: "*default", SetupTimeField: "*default", AnswerTimeField: "*default", DurationField: "*default"},
		&DerivedCharger{RunId: "extra2", ReqTypeField: "*default", DirectionField: "*default", TenantField: "*default", CategoryField: "*default",
			AccountField: "ivo", SubjectField: "ivo", DestinationField: "*default", SetupTimeField: "*default", AnswerTimeField: "*default", DurationField: "*default"},
	}
	eDc2 := append(dc2, dcDf)
	if dc2, _ = dc2.AppendDefaultRun(); !reflect.DeepEqual(dc2, eDc2) {
		t.Error("Unexpected result.")
	}
}
