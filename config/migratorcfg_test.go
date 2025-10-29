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

func TestMigratorCgrCfgloadFromJsonCfg(t *testing.T) {
	var migcfg, expected MigratorCgrCfg
	if err := migcfg.loadFromJsonCfg(nil); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(migcfg, expected) {
		t.Errorf("Expected: %+v ,received: %+v", expected, migcfg)
	}
	if err := migcfg.loadFromJsonCfg(new(MigratorCfgJson)); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(migcfg, expected) {
		t.Errorf("Expected: %+v ,received: %+v", expected, migcfg)
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
		OutDataDBType:     "*redis",
		OutDataDBHost:     "127.0.0.1",
		OutDataDBPort:     "6379",
		OutDataDBName:     "10",
		OutDataDBUser:     "cgrates",
		OutDataDBPassword: "",
		OutDataDBEncoding: "msgpack",
		OutStorDBType:     "*mysql",
		OutStorDBHost:     "127.0.0.1",
		OutStorDBPort:     "3306",
		OutStorDBName:     "cgrates",
		OutStorDBUser:     "cgrates",
		OutStorDBPassword: "",
	}
	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnThSCfg, err := jsnCfg.MigratorCfgJson(); err != nil {
		t.Error(err)
	} else if err = migcfg.loadFromJsonCfg(jsnThSCfg); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, migcfg) {
		t.Errorf("Expected: %+v , received: %+v", expected, migcfg)
	}
}

func TestMigratorCgrCfgAsMapInterface(t *testing.T) {
	var migcfg MigratorCgrCfg
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
		"users_filters":[],
	},
}`
	var users_filters []string
	eMap := map[string]any{
		"out_datadb_type":           "*redis",
		"out_datadb_host":           "127.0.0.1",
		"out_datadb_port":           "6379",
		"out_datadb_name":           "10",
		"out_datadb_user":           "cgrates",
		"out_datadb_password":       "",
		"out_datadb_encoding":       "msgpack",
		"out_stordb_type":           "*mysql",
		"out_stordb_host":           "127.0.0.1",
		"out_stordb_port":           "3306",
		"out_stordb_name":           "cgrates",
		"out_stordb_user":           "cgrates",
		"out_stordb_password":       "",
		"users_filters":             users_filters,
		"out_datadb_redis_sentinel": "",
	}
	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnThSCfg, err := jsnCfg.MigratorCfgJson(); err != nil {
		t.Error(err)
	} else if err = migcfg.loadFromJsonCfg(jsnThSCfg); err != nil {
		t.Error(err)
	} else if rcv := migcfg.AsMapInterface(); !reflect.DeepEqual(eMap, rcv) {
		t.Errorf("\nExpected: %+v\nReceived: %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}
	cfgJSONStr = `{
		"migrator": {
			"out_datadb_type": "redis",
			"out_datadb_host": "127.0.0.1",
			"out_datadb_port": "6379",
			"out_datadb_name": "10",
			"out_datadb_user": "cgrates",
			"out_datadb_password": "out_datadb_password",
			"out_datadb_encoding" : "msgpack",
			"out_stordb_type": "mysql",
			"out_stordb_host": "127.0.0.1",
			"out_stordb_port": "3306",
			"out_stordb_name": "cgrates",
			"out_stordb_user": "cgrates",
			"out_stordb_password": "out_stordb_password",
			"users_filters":["users","filters","Account"],
			"out_datadb_redis_sentinel": "out_datadb_redis_sentinel",
		},
	}`

	eMap = map[string]any{
		"out_datadb_type":           "*redis",
		"out_datadb_host":           "127.0.0.1",
		"out_datadb_port":           "6379",
		"out_datadb_name":           "10",
		"out_datadb_user":           "cgrates",
		"out_datadb_password":       "out_datadb_password",
		"out_datadb_encoding":       "msgpack",
		"out_stordb_type":           "*mysql",
		"out_stordb_host":           "127.0.0.1",
		"out_stordb_port":           "3306",
		"out_stordb_name":           "cgrates",
		"out_stordb_user":           "cgrates",
		"out_stordb_password":       "out_stordb_password",
		"users_filters":             []string{"users", "filters", "Account"},
		"out_datadb_redis_sentinel": "out_datadb_redis_sentinel",
	}
	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnThSCfg, err := jsnCfg.MigratorCfgJson(); err != nil {
		t.Error(err)
	} else if err = migcfg.loadFromJsonCfg(jsnThSCfg); err != nil {
		t.Error(err)
	} else if rcv := migcfg.AsMapInterface(); !reflect.DeepEqual(eMap, rcv) {
		t.Errorf("\nExpected: %+v\nReceived: %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}

}

func TestMigratorCgrCfgloadFromJsonCfg2(t *testing.T) {
	str := "*test"
	mg := &MigratorCgrCfg{}
	jsnCfg := &MigratorCfgJson{
		Out_dataDB_type: &str,
		Out_storDB_type: &str,
	}

	err := mg.loadFromJsonCfg(jsnCfg)
	if err != nil {
		t.Error(err)
	}

	exp := &MigratorCgrCfg{
		OutDataDBType: str,
		OutStorDBType: str,
	}

	if !reflect.DeepEqual(exp, mg) {
		t.Errorf("\nexpected: %s \nreceived: %s\n", utils.ToJSON(exp), utils.ToJSON(mg))
	}
}
