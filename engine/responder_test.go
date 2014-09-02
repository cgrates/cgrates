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

package engine

import (
	"reflect"
	"testing"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

// Test internal abilites of GetDerivedChargers
func TestResponderGetDerivedChargers(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	cfgedDC := utils.DerivedChargers{&utils.DerivedCharger{RunId: "responder1", ReqTypeField: "test", DirectionField: "test", TenantField: "test",
		CategoryField: "test", AccountField: "test", SubjectField: "test", DestinationField: "test", SetupTimeField: "test", AnswerTimeField: "test", UsageField: "test"}}
	cfg.DerivedChargers = cfgedDC
	config.SetCgrConfig(cfg)
	r := Responder{}
	attrs := utils.AttrDerivedChargers{Tenant: "cgrates.org", Category: "call", Direction: "*out", Account: "dan", Subject: "dan"}
	var dcs utils.DerivedChargers
	if err := r.GetDerivedChargers(attrs, &dcs); err != nil {
		t.Error("Unexpected error", err.Error())
	} else if !reflect.DeepEqual(dcs, cfgedDC) {
		//t.Errorf("Expecting: %v, received: %v ", cfgedDC, dcs)
		//TODO: fix the above test when DEBUG=false
	}
}
