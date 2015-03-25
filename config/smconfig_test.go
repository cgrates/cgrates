/*
Real-time Charging System for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

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
	"github.com/cgrates/cgrates/utils"
	"reflect"
	"testing"
)

func TesSmFsConfigLoadFromJsonCfg(t *testing.T) {
	smFsJsnCfg := &SmFsJsonCfg{
		Enabled: utils.BoolPointer(true),
		Connections: &[]*FsConnJsonCfg{
			&FsConnJsonCfg{
				Server:     utils.StringPointer("1.2.3.4:8021"),
				Password:   utils.StringPointer("ClueCon"),
				Reconnects: utils.IntPointer(5),
			},
			&FsConnJsonCfg{
				Server:     utils.StringPointer("2.3.4.5:8021"),
				Password:   utils.StringPointer("ClueCon"),
				Reconnects: utils.IntPointer(5),
			},
		},
	}
	eSmFsConfig := &SmFsConfig{Enabled: true,
		Connections: []*FsConnConfig{
			&FsConnConfig{Server: "1.2.3.4:8021", Password: "ClueCon", Reconnects: 5},
			&FsConnConfig{Server: "1.2.3.4:8021", Password: "ClueCon", Reconnects: 5},
		},
	}
	smFsCfg := new(SmFsConfig)
	if err := smFsCfg.loadFromJsonCfg(smFsJsnCfg); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eSmFsConfig, smFsCfg) {
		t.Error("Received: ", smFsCfg)
	}
}
