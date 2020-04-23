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
	"testing"

	"github.com/cgrates/cgrates/utils"
)

func TestStoreDbCfgloadFromJsonCfg(t *testing.T) {
	var dbcfg, expected StorDbCfg
	if err := dbcfg.loadFromJsonCfg(nil); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(dbcfg, expected) {
		t.Errorf("Expected: %+v ,recived: %+v", expected, dbcfg)
	}
	if err := dbcfg.loadFromJsonCfg(new(DbJsonCfg)); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(dbcfg, expected) {
		t.Errorf("Expected: %+v ,recived: %+v", expected, dbcfg)
	}
	cfgJSONStr := `{
"stor_db": {								// database used to store offline tariff plans and CDRs
	"db_type": "*mysql",					// stor database type to use: <*mongo|*mysql|*postgres|*internal>
	"db_host": "127.0.0.1",					// the host to connect to
	"db_port": 3306,						// the port to reach the stordb
	"db_name": "cgrates",					// stor database name
	"db_user": "cgrates",					// username to use when connecting to stordb
	"db_password": "password",				// password to use when connecting to stordb
	"max_open_conns": 100,					// maximum database connections opened, not applying for mongo
	"max_idle_conns": 10,					// maximum database connections idle, not applying for mongo
	"conn_max_lifetime": 0, 				// maximum amount of time in seconds a connection may be reused (0 for unlimited), not applying for mongo
	"string_indexed_fields": [],			// indexes on cdrs table to speed up queries, used in case of *mongo and *internal
	"prefix_indexed_fields":[],				// prefix indexes on cdrs table to speed up queries, used in case of *internal
	}
}`
	expected = StorDbCfg{
		Type:                "mysql",
		Host:                "127.0.0.1",
		Port:                "3306",
		Name:                "cgrates",
		User:                "cgrates",
		Password:            "password",
		MaxOpenConns:        100,
		MaxIdleConns:        10,
		ConnMaxLifetime:     0,
		StringIndexedFields: []string{},
		PrefixIndexedFields: []string{},
	}
	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnStoreDbCfg, err := jsnCfg.DbJsonCfg(STORDB_JSN); err != nil {
		t.Error(err)
	} else if err = dbcfg.loadFromJsonCfg(jsnStoreDbCfg); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, dbcfg) {
		t.Errorf("Expected: %+v , recived: %+v", expected, dbcfg)
	}
}
func TestStoreDbCfgloadFromJsonCfgPort(t *testing.T) {
	var dbcfg StorDbCfg
	cfgJSONStr := `{
"stor_db": {
	"db_type": "mongo",
	}
}`
	expected := StorDbCfg{
		Type: "mongo",
	}
	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnDataDbCfg, err := jsnCfg.DbJsonCfg(STORDB_JSN); err != nil {
		t.Error(err)
	} else if err = dbcfg.loadFromJsonCfg(jsnDataDbCfg); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, dbcfg) {
		t.Errorf("Expected: %+v , recived: %+v", expected, dbcfg)
	}
	cfgJSONStr = `{
"stor_db": {
	"db_type": "mongo",
	"db_port": -1,
	}
}`
	expected = StorDbCfg{
		Type: "mongo",
		Port: "27017",
	}
	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnDataDbCfg, err := jsnCfg.DbJsonCfg(STORDB_JSN); err != nil {
		t.Error(err)
	} else if err = dbcfg.loadFromJsonCfg(jsnDataDbCfg); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, dbcfg) {
		t.Errorf("Expected: %+v , recived: %+v", expected, dbcfg)
	}
	cfgJSONStr = `{
"stor_db": {
	"db_type": "*internal",
	"db_port": -1,
	}
}`
	expected = StorDbCfg{
		Type: "internal",
		Port: "internal",
	}
	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnDataDbCfg, err := jsnCfg.DbJsonCfg(STORDB_JSN); err != nil {
		t.Error(err)
	} else if err = dbcfg.loadFromJsonCfg(jsnDataDbCfg); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, dbcfg) {
		t.Errorf("Expected: %+v , recived: %+v", expected, dbcfg)
	}
}

func TestStorDbCfgAsMapInterface(t *testing.T) {
	var dbcfg StorDbCfg
	cfgJSONStr := `{
		"stor_db": {								
			"db_type": "*mysql",					
			"db_host": "127.0.0.1",					
			"db_port": -1,						
			"db_name": "cgrates",					
			"db_user": "cgrates",					
			"db_password": "",						
			"max_open_conns": 100,					
			"max_idle_conns": 10,					
			"conn_max_lifetime": 0, 				
			"string_indexed_fields": [],			
			"prefix_indexed_fields":[],				
			"query_timeout":"10s",
			"sslmode":"disable",					
			"items":{
				"session_costs": {"limit": -1, "ttl": "", "static_ttl": false}, 
				"cdrs": {"limit": -1, "ttl": "", "static_ttl": false}, 		
			},
		},
}`

	eMap := map[string]interface{}{
		"db_type":               "*mysql",
		"db_host":               "127.0.0.1",
		"db_port":               3306,
		"db_name":               "cgrates",
		"db_user":               "cgrates",
		"db_password":           "",
		"max_open_conns":        100,
		"max_idle_conns":        10,
		"conn_max_lifetime":     0,
		"string_indexed_fields": []string{},
		"prefix_indexed_fields": []string{},
		"query_timeout":         "10s",
		"sslmode":               "disable",
		"items": map[string]interface{}{
			"session_costs": map[string]interface{}{"limit": -1, "ttl": "", "static_ttl": false, "remote": false, "replicate": false},
			"cdrs":          map[string]interface{}{"limit": -1, "ttl": "", "static_ttl": false, "remote": false, "replicate": false},
		},
	}
	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnStoreDbCfg, err := jsnCfg.DbJsonCfg(STORDB_JSN); err != nil {
		t.Error(err)
	} else if err = dbcfg.loadFromJsonCfg(jsnStoreDbCfg); err != nil {
		t.Error(err)
	} else if rcv := dbcfg.AsMapInterface(); !reflect.DeepEqual(eMap, rcv) {
		t.Errorf("Expected: %+v ,\n recived: %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}
}
