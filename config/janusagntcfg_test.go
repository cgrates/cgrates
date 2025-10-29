/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/

package config

import (
	"reflect"
	"testing"

	"github.com/cgrates/cgrates/utils"
)

func TestLoadFromJSONCfgNil(t *testing.T) {
	var jc JanusConn
	err := jc.loadFromJSONCfg(nil)
	if err != nil {
		t.Errorf("Expected %v, received %v", nil, err)
	}

}

func TestJanusConnAsMapInterface(t *testing.T) {
	js := &JanusConn{
		Address: "127.001",
		Type:    "ws",
	}
	exp := map[string]any{
		utils.AddressCfg:       "127.001",
		utils.TypeCfg:          "ws",
		utils.AdminAddressCfg:  "",
		utils.AdminPasswordCfg: "",
	}
	val := js.AsMapInterface()
	if !reflect.DeepEqual(val, exp) {
		t.Errorf("expected %+v received %+v", exp, val)
	}

}
