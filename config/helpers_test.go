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
	"reflect"
	"regexp"
	"testing"

	"github.com/cgrates/cgrates/utils"
)

func TestParseRSRFields(t *testing.T) {
	fields := `host,~sip_redirected_to:s/sip:\+49(\d+)@/0$1/,destination`
	expectParsedFields := []*utils.RSRField{&utils.RSRField{Id: "host"},
		&utils.RSRField{Id: "sip_redirected_to", RSRules: []*utils.ReSearchReplace{&utils.ReSearchReplace{regexp.MustCompile(`sip:\+49(\d+)@`), "0$1"}}},
		&utils.RSRField{Id: "destination"}}
	if parsedFields, err := ParseRSRFields(fields); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if !reflect.DeepEqual(parsedFields, expectParsedFields) {
		t.Errorf("Unexpected value of parsed fields")
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
	edcs := utils.DerivedChargers{
		&utils.DerivedCharger{RunId: "run1", ReqTypeField: "test1", DirectionField: "test1", TenantField: "test1", TorField: "test1",
			AccountField: "test1", SubjectField: "test1", DestinationField: "test1", SetupTimeField: "test1", AnswerTimeField: "test1", DurationField: "test1"},
		&utils.DerivedCharger{RunId: "run2", ReqTypeField: "test2", DirectionField: "test2", TenantField: "test2", TorField: "test2",
			AccountField: "test2", SubjectField: "test2", DestinationField: "test2", SetupTimeField: "test2", AnswerTimeField: "test2", DurationField: "test2"}}
	if cfg, err := NewCGRConfigFromBytes(eFieldsCfg); err != nil {
		t.Error("Could not parse the config", err.Error())
	} else if !reflect.DeepEqual(cfg.DerivedChargers, edcs) {
		t.Errorf("Expecting: %v, received: %v", edcs, cfg.DerivedChargers)
	}
}
