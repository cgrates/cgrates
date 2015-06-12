/*
Real-time Charging System for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

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

package v1

import (
	"reflect"
	"testing"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var apierDcT *ApierV1

func init() {
	dataStorage, _ := engine.NewMapStorage()
	cfg, _ := config.NewDefaultCGRConfig()
	apierDcT = &ApierV1{RatingDb: engine.RatingStorage(dataStorage), Config: cfg}
}

/*
func TestGetEmptyDC(t *testing.T) {
	attrs := utils.AttrDerivedChargers{Tenant: "cgrates.org", Category: "call", Direction: "*out", Account: "dan", Subject: "dan"}
	var dcs utils.DerivedChargers
	if err := apierDcT.GetDerivedChargers(attrs, &dcs); err != nil {
		t.Error("Unexpected error", err.Error())
	} else if !reflect.DeepEqual(dcs, apierDcT.Config.DerivedChargers) {
		t.Error("Returned DerivedChargers not matching the configured ones")
	}
}
*/

func TestSetDC(t *testing.T) {
	dcs1 := utils.DerivedChargers{
		&utils.DerivedCharger{RunId: "extra1", ReqTypeField: "^prepaid", DirectionField: "*default", TenantField: "*default", CategoryField: "*default",
			AccountField: "rif", SubjectField: "rif", DestinationField: "*default", SetupTimeField: "*default", AnswerTimeField: "*default", UsageField: "*default"},
		&utils.DerivedCharger{RunId: "extra2", ReqTypeField: "*default", DirectionField: "*default", TenantField: "*default", CategoryField: "*default",
			AccountField: "ivo", SubjectField: "ivo", DestinationField: "*default", SetupTimeField: "*default", AnswerTimeField: "*default", UsageField: "*default"},
	}
	attrs := AttrSetDerivedChargers{Direction: "*out", Tenant: "cgrates.org", Category: "call", Account: "dan", Subject: "dan", DerivedChargers: dcs1}
	var reply string
	if err := apierDcT.SetDerivedChargers(attrs, &reply); err != nil {
		t.Error("Unexpected error", err.Error())
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}
}

func TestGetDC(t *testing.T) {
	attrs := utils.AttrDerivedChargers{Tenant: "cgrates.org", Category: "call", Direction: "*out", Account: "dan", Subject: "dan"}
	eDcs := utils.DerivedChargers{
		&utils.DerivedCharger{RunId: "extra1", ReqTypeField: "^prepaid", DirectionField: "*default", TenantField: "*default", CategoryField: "*default",
			AccountField: "rif", SubjectField: "rif", DestinationField: "*default", SetupTimeField: "*default", AnswerTimeField: "*default", UsageField: "*default"},
		&utils.DerivedCharger{RunId: "extra2", ReqTypeField: "*default", DirectionField: "*default", TenantField: "*default", CategoryField: "*default",
			AccountField: "ivo", SubjectField: "ivo", DestinationField: "*default", SetupTimeField: "*default", AnswerTimeField: "*default", UsageField: "*default"},
	}
	var dcs utils.DerivedChargers
	if err := apierDcT.GetDerivedChargers(attrs, &dcs); err != nil {
		t.Error("Unexpected error", err.Error())
	} else if !reflect.DeepEqual(dcs, eDcs) {
		t.Errorf("Expecting: %v, received: %v", eDcs, dcs)
	}
}

func TestRemDC(t *testing.T) {
	attrs := AttrRemDerivedChargers{Direction: "*out", Tenant: "cgrates.org", Category: "call", Account: "dan", Subject: "dan"}
	var reply string
	if err := apierDcT.RemDerivedChargers(attrs, &reply); err != nil {
		t.Error("Unexpected error", err.Error())
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}
}

/*
func TestGetEmptyDC2(t *testing.T) {
	attrs := utils.AttrDerivedChargers{Tenant: "cgrates.org", Category: "call", Direction: "*out", Account: "dan", Subject: "dan"}
	var dcs utils.DerivedChargers
	if err := apierDcT.GetDerivedChargers(attrs, &dcs); err != nil {
		t.Error("Unexpected error", err.Error())
	} else if !reflect.DeepEqual(dcs, apierDcT.Config.DerivedChargers) {
		for _, dc := range dcs {
			fmt.Printf("Got dc: %v\n", dc)
		}
		t.Error("Returned DerivedChargers not matching the configured ones")
	}
}
*/
