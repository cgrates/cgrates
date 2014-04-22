/*
Rating system for Telecom Environments
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

package config

import (
	"github.com/cgrates/cgrates/utils"
	"reflect"
	"testing"
)

func TestAppendDerivedChargers(t *testing.T) {
	var err error
	dcs := make(DerivedChargers, 0)
	if _, err := dcs.Append(&DerivedCharger{RunId: utils.DEFAULT_RUNID}); err == nil {
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
		TorField:         "tor1",
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
		TorField:         "~tor2:s/sip:(.+)/$1/",
		AccountField:     "~account2:s/sip:(.+)/$1/",
		SubjectField:     "~subject2:s/sip:(.+)/$1/",
		DestinationField: "~destination2:s/sip:(.+)/$1/",
		SetupTimeField:   "~setuptime2:s/sip:(.+)/$1/",
		AnswerTimeField:  "~answertime2:s/sip:(.+)/$1/",
		DurationField:    "~duration2:s/sip:(.+)/$1/",
	}
	edc2.rsrReqTypeField, _ = utils.NewRSRField("~reqtype2:s/sip:(.+)/$1/")
	edc2.rsrDirectionField, _ = utils.NewRSRField("~direction2:s/sip:(.+)/$1/")
	edc2.rsrTenantField, _ = utils.NewRSRField("~tenant2:s/sip:(.+)/$1/")
	edc2.rsrTorField, _ = utils.NewRSRField("~tor2:s/sip:(.+)/$1/")
	edc2.rsrAccountField, _ = utils.NewRSRField("~account2:s/sip:(.+)/$1/")
	edc2.rsrSubjectField, _ = utils.NewRSRField("~subject2:s/sip:(.+)/$1/")
	edc2.rsrDestinationField, _ = utils.NewRSRField("~destination2:s/sip:(.+)/$1/")
	edc2.rsrSetupTimeField, _ = utils.NewRSRField("~setuptime2:s/sip:(.+)/$1/")
	edc2.rsrAnswerTimeField, _ = utils.NewRSRField("~answertime2:s/sip:(.+)/$1/")
	edc2.rsrDurationField, _ = utils.NewRSRField("~duration2:s/sip:(.+)/$1/")
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

func TestParseCfgDerivedCharging(t *testing.T) {
	eFieldsCfg := []byte(`[derived_charging]
run_ids = run1, run2
reqtype_fields = test1, test2 
direction_fields = test1, test2
tenant_fields = test1, test2
tor_fields = test1, test2
account_fields = test1, test2
subject_fields = test1, test2
destination_fields = test1, test2
setup_time_fields = test1, test2
answer_time_fields = test1, test2
duration_fields = test1, test2
`)
	edcs := DerivedChargers{
		&DerivedCharger{RunId: "run1", ReqTypeField: "test1", DirectionField: "test1", TenantField: "test1", TorField: "test1",
			AccountField: "test1", SubjectField: "test1", DestinationField: "test1", SetupTimeField: "test1", AnswerTimeField: "test1", DurationField: "test1"},
		&DerivedCharger{RunId: "run2", ReqTypeField: "test2", DirectionField: "test2", TenantField: "test2", TorField: "test2",
			AccountField: "test2", SubjectField: "test2", DestinationField: "test2", SetupTimeField: "test2", AnswerTimeField: "test2", DurationField: "test2"}}
	if cfg, err := NewCGRConfigFromBytes(eFieldsCfg); err != nil {
		t.Error("Could not parse the config", err.Error())
	} else if !reflect.DeepEqual(cfg.DerivedChargers, edcs) {
		t.Errorf("Expecting: %v, received: %v", edcs, cfg.DerivedChargers)
	}
}
