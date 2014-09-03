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

package engine

import (
	"reflect"
	"testing"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

var cfgDcT *config.CGRConfig
var acntDb AccountingStorage

func init() {
	cfgDcT, _ = config.NewDefaultCGRConfig()
	acntDb = accountingStorage
	acntDb.CacheAccounting(nil, nil, nil, nil)
}

// Accounting db has no DerivedChargers nor configured defaults
func TestHandleGetEmptyDC(t *testing.T) {
	attrs := utils.AttrDerivedChargers{Tenant: "cgrates.org", Category: "call", Direction: "*out", Account: "test2", Subject: "test2"}
	if dcs, err := HandleGetDerivedChargers(acntDb, cfgDcT, attrs); err != nil {
		t.Error("Unexpected error", err.Error())
	} else if !reflect.DeepEqual(dcs, cfgDcT.DerivedChargers) {
		t.Error("Returned DerivedChargers not matching the configured ones")
	}
}

// Accounting db has no DerivedChargers, configured defaults
func TestHandleGetConfiguredDC(t *testing.T) {
	cfgedDC := utils.DerivedChargers{&utils.DerivedCharger{RunId: "responder1", ReqTypeField: "test", DirectionField: "test", TenantField: "test",
		CategoryField: "test", AccountField: "test", SubjectField: "test", DestinationField: "test", SetupTimeField: "test", AnswerTimeField: "test", UsageField: "test"}}
	cfgDcT.DerivedChargers = cfgedDC
	attrs := utils.AttrDerivedChargers{Tenant: "cgrates.org", Category: "call", Direction: "*out", Account: "test3", Subject: "test3"}
	if dcs, err := HandleGetDerivedChargers(acntDb, cfgDcT, attrs); err != nil {
		t.Error("Unexpected error", err.Error())
	} else if !reflect.DeepEqual(dcs, cfgedDC) {
		t.Error("Returned DerivedChargers not matching the configured ones")
	}
}

// Receive composed derived chargers
func TestHandleGetStoredDC(t *testing.T) {
	keyCharger1 := utils.DerivedChargersKey("*out", "cgrates.org", "call", "rif", "rif")
	charger1 := utils.DerivedChargers{
		&utils.DerivedCharger{RunId: "extra1", ReqTypeField: "^prepaid", DirectionField: "*default", TenantField: "*default", CategoryField: "*default",
			AccountField: "rif", SubjectField: "rif", DestinationField: "*default", SetupTimeField: "*default", AnswerTimeField: "*default", UsageField: "*default"},
		&utils.DerivedCharger{RunId: "extra2", ReqTypeField: "*default", DirectionField: "*default", TenantField: "*default", CategoryField: "*default",
			AccountField: "ivo", SubjectField: "ivo", DestinationField: "*default", SetupTimeField: "*default", AnswerTimeField: "*default", UsageField: "*default"},
	}
	if err := acntDb.SetDerivedChargers(keyCharger1, charger1); err != nil {
		t.Error("Error on setDerivedChargers", err.Error())
	}
	// Expected Charger should have default configured values added
	expCharger1 := append(charger1, &utils.DerivedCharger{RunId: "responder1", ReqTypeField: "test", DirectionField: "test", TenantField: "test",
		CategoryField: "test", AccountField: "test", SubjectField: "test", DestinationField: "test", SetupTimeField: "test", AnswerTimeField: "test", UsageField: "test"})
	acntDb.CacheAccounting(nil, nil, nil, nil)
	attrs := utils.AttrDerivedChargers{Tenant: "cgrates.org", Category: "call", Direction: "*out", Account: "rif", Subject: "rif"}
	if dcs, err := HandleGetDerivedChargers(acntDb, cfgDcT, attrs); err != nil {
		t.Error("Unexpected error", err.Error())
	} else if !reflect.DeepEqual(dcs, expCharger1) {
		t.Error("Returned DerivedChargers not matching the configured ones")
	}
	cfgDcT.CombinedDerivedChargers = false
	if dcs, err := HandleGetDerivedChargers(acntDb, cfgDcT, attrs); err != nil {
		t.Error("Unexpected error", err.Error())
	} else if !reflect.DeepEqual(dcs, charger1) {
		t.Error("Returned DerivedChargers not matching the configured ones")
	}
}
