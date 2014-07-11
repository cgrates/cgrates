/*
Real-time Charging System for Telecom & ISP environments
Copyright (C) 2012-2014 ITsysCOM GmbH

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

package config

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/cgrates/cgrates/utils"
)

func TestConfigSlice(t *testing.T) {
	eCS := []string{"", ""}
	if cs, err := ConfigSlice(" , "); err != nil {
		t.Error("Unexpected error: ", err)
	} else if !reflect.DeepEqual(eCS, cs) {
		t.Errorf("Expecting: %v, received: %v", eCS, cs)
	}
}

func TestParseCfgDerivedCharging(t *testing.T) {
	eFieldsCfg := []byte(`[derived_charging]
run_ids = run1, run2
run_filters =,
reqtype_fields = test1, test2 
direction_fields = test1, test2
tenant_fields = test1, test2
category_fields = test1, test2
account_fields = test1, test2
subject_fields = test1, test2
destination_fields = test1, test2
setup_time_fields = test1, test2
answer_time_fields = test1, test2
usage_fields = test1, test2
`)
	edcs := utils.DerivedChargers{
		&utils.DerivedCharger{RunId: "run1", ReqTypeField: "test1", DirectionField: "test1", TenantField: "test1", CategoryField: "test1",
			AccountField: "test1", SubjectField: "test1", DestinationField: "test1", SetupTimeField: "test1", AnswerTimeField: "test1", UsageField: "test1"},
		&utils.DerivedCharger{RunId: "run2", ReqTypeField: "test2", DirectionField: "test2", TenantField: "test2", CategoryField: "test2",
			AccountField: "test2", SubjectField: "test2", DestinationField: "test2", SetupTimeField: "test2", AnswerTimeField: "test2", UsageField: "test2"}}
	if cfg, err := NewCGRConfigFromBytes(eFieldsCfg); err != nil {
		t.Error("Could not parse the config", err.Error())
	} else if !reflect.DeepEqual(cfg.DerivedChargers, edcs) {
		t.Errorf("Expecting: %v, received: %v", edcs, cfg.DerivedChargers)
	}
}

func TestParseCfgDerivedChargingDn1(t *testing.T) {
	eFieldsCfg := []byte(`[derived_charging]
run_ids = run1, run2
run_filters =~account:s/^\w+[mpls]\d{6}$//,~account:s/^0\d{9}$//;^account/value/
reqtype_fields = test1, test2 
direction_fields = test1, test2
tenant_fields = test1, test2
category_fields = test1, test2
account_fields = test1, test2
subject_fields = test1, test2
destination_fields = test1, test2
setup_time_fields = test1, test2
answer_time_fields = test1, test2
usage_fields = test1, test2
`)
	eDcs := make(utils.DerivedChargers, 2)
	if dc, err := utils.NewDerivedCharger("run1", `~account:s/^\w+[mpls]\d{6}$//`, "test1", "test1", "test1",
		"test1", "test1", "test1", "test1", "test1", "test1", "test1"); err != nil {
		t.Error("Unexpected error: ", err)
	} else {
		eDcs[0] = dc
	}
	if dc, err := utils.NewDerivedCharger("run2", `~account:s/^0\d{9}$//;^account/value/`, "test2", "test2", "test2",
		"test2", "test2", "test2", "test2", "test2", "test2", "test2"); err != nil {
		t.Error("Unexpected error: ", err)
	} else {
		eDcs[1] = dc
	}

	if cfg, err := NewCGRConfigFromBytes(eFieldsCfg); err != nil {
		t.Error("Could not parse the config", err.Error())
	} else if !reflect.DeepEqual(cfg.DerivedChargers, eDcs) {
		dcsJson, _ := json.Marshal(cfg.DerivedChargers)
		t.Errorf("Received: %s", string(dcsJson))
	}
}

func TestParseCdrcCdrFields(t *testing.T) {
	eFieldsCfg := []byte(`[cdrc]
cdr_type = test
tor_field = tor1
accid_field = accid1
reqtype_field = reqtype1
direction_field = direction1
tenant_field = tenant1
category_field = category1
account_field = account1
subject_field = subject1
destination_field = destination1
setup_time_field = setuptime1
answer_time_field = answertime1
usage_field = duration1
extra_fields = extra1:extraval1,extra2:extraval1
`)
	eCdrcCdrFlds := map[string]*utils.RSRField{
		utils.TOR:         &utils.RSRField{Id: "tor1"},
		utils.ACCID:       &utils.RSRField{Id: "accid1"},
		utils.REQTYPE:     &utils.RSRField{Id: "reqtype1"},
		utils.DIRECTION:   &utils.RSRField{Id: "direction1"},
		utils.TENANT:      &utils.RSRField{Id: "tenant1"},
		utils.CATEGORY:    &utils.RSRField{Id: "category1"},
		utils.ACCOUNT:     &utils.RSRField{Id: "account1"},
		utils.SUBJECT:     &utils.RSRField{Id: "subject1"},
		utils.DESTINATION: &utils.RSRField{Id: "destination1"},
		utils.SETUP_TIME:  &utils.RSRField{Id: "setuptime1"},
		utils.ANSWER_TIME: &utils.RSRField{Id: "answertime1"},
		utils.USAGE:       &utils.RSRField{Id: "duration1"},
		"extra1":          &utils.RSRField{Id: "extraval1"},
		"extra2":          &utils.RSRField{Id: "extraval1"},
	}
	if cfg, err := NewCGRConfigFromBytes(eFieldsCfg); err != nil {
		t.Error("Could not parse the config", err.Error())
	} else if !reflect.DeepEqual(cfg.CdrcCdrFields, eCdrcCdrFlds) {
		t.Errorf("Expecting: %v, received: %v, tor: %v", eCdrcCdrFlds, cfg.CdrcCdrFields, cfg.CdrcCdrFields[utils.TOR])
	}
}
