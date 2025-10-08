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
	"time"

	"github.com/cgrates/cgrates/utils"
)

func TestAsteriskConnCfgloadFromJSONCfg(t *testing.T) {
	aConnCfg := &AsteriskConnCfg{
		MaxReconnectInterval: time.Duration(5),
	}

	jsnCfg := &AstConnJsonCfg{

		Max_reconnect_interval: utils.StringPointer("return error"),
	}
	expErr := "time: invalid duration \"return error\""
	if err := aConnCfg.loadFromJSONCfg(jsnCfg); err == nil || err.Error() != expErr {
		t.Errorf("Expected error <%v>, Received error <%v>", expErr, err)
	}

}

func TestDiffAstConnJsonCfgMaxReconn(t *testing.T) {
	v1 := &AsteriskConnCfg{
		MaxReconnectInterval: time.Duration(4),
	}

	v2 := &AsteriskConnCfg{
		MaxReconnectInterval: time.Duration(5),
	}

	expected := &AstConnJsonCfg{
		Max_reconnect_interval: utils.StringPointer("5ns"),
	}

	rcv := diffAstConnJsonCfg(v1, v2)
	if !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}

}
