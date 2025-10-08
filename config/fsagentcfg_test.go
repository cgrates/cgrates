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

func TestFsConnCfgLoadFromJSONCfg(t *testing.T) {
	fs := &FsConnCfg{
		MaxReconnectInterval: time.Duration(4),
	}
	jsnCfg := &FsConnJsonCfg{
		MaxReconnectInterval: utils.StringPointer("invalid time"),
	}
	expErr := `time: invalid duration "invalid time"`
	if err := fs.loadFromJSONCfg(jsnCfg); err.Error() != expErr {
		t.Errorf("Expected error <%v>, Received error <%v>", expErr, err.Error())

	}
}

func TestDiffFsConnJsonCfgMaxReconnInterval(t *testing.T) {
	v1 := &FsConnCfg{MaxReconnectInterval: time.Duration(3)}

	v2 := &FsConnCfg{MaxReconnectInterval: time.Duration(2)}

	expected := &FsConnJsonCfg{MaxReconnectInterval: utils.StringPointer("2ns")}

	rcv := diffFsConnJsonCfg(v1, v2)
	if !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}

	v1 = v2
	expected = &FsConnJsonCfg{}

	rcv = diffFsConnJsonCfg(v1, v2)
	if !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}
