/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
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
	"reflect"
	"strings"
	"testing"
)

func TestMigratorCgrCfgloadFromJsonCfg(t *testing.T) {
	var migcfg, expected MigratorCgrCfg
	if err := migcfg.loadFromJsonCfg(nil); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(migcfg, expected) {
		t.Errorf("Expected: %+v ,recived: %+v", expected, migcfg)
	}
	if err := migcfg.loadFromJsonCfg(new(MigratorCfgJson)); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(migcfg, expected) {
		t.Errorf("Expected: %+v ,recived: %+v", expected, migcfg)
	}
	cfgJSONStr := `{
"migrator": {
	"out_datadb_type": "redis",
	"out_datadb_host": "127.0.0.1",
	"out_datadb_port": "6379",
	"out_datadb_name": "10",
	"out_datadb_user": "cgrates",
	"out_datadb_password": "",
	"out_datadb_encoding" : "msgpack",
	"out_stordb_type": "mysql",
	"out_stordb_host": "127.0.0.1",
	"out_stordb_port": "3306",
	"out_stordb_name": "cgrates",
	"out_stordb_user": "cgrates",
	"out_stordb_password": "",
},	
}`
	expected = MigratorCgrCfg{
		OutDataDBType:     "redis",
		OutDataDBHost:     "127.0.0.1",
		OutDataDBPort:     "6379",
		OutDataDBName:     "10",
		OutDataDBUser:     "cgrates",
		OutDataDBPassword: "",
		OutDataDBEncoding: "msgpack",
		OutStorDBType:     "mysql",
		OutStorDBHost:     "127.0.0.1",
		OutStorDBPort:     "3306",
		OutStorDBName:     "cgrates",
		OutStorDBUser:     "cgrates",
		OutStorDBPassword: "",
	}
	if jsnCfg, err := NewCgrJsonCfgFromReader(strings.NewReader(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnThSCfg, err := jsnCfg.MigratorCfgJson(); err != nil {
		t.Error(err)
	} else if err = migcfg.loadFromJsonCfg(jsnThSCfg); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, migcfg) {
		t.Errorf("Expected: %+v , recived: %+v", expected, migcfg)
	}
}
