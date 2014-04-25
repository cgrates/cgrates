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

package apier

import (
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"reflect"
	"testing"
)

var apierDcT *ApierV1

func init() {
	dataStorage, _ := engine.NewMapStorage()
	cfg, _ := config.NewDefaultCGRConfig()
	apierDcT = &ApierV1{AccountDb: engine.AccountingStorage(dataStorage), Config: cfg}
}

func TestGetEmptyDC(t *testing.T) {
	attrs := utils.AttrDerivedChargers{Tenant: "cgrates.org", Tor: "call", Direction: "*out", Account: "dan", Subject: "dan"}
	var dcs utils.DerivedChargers
	if err := apierDcT.DerivedChargers(attrs, &dcs); err != nil {
		t.Error("Unexpected error", err.Error())
	} else if !reflect.DeepEqual(dcs, apierDcT.Config.DerivedChargers) {
		t.Error("Returned DerivedChargers not matching the configured ones")
	}
}
