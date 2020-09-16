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
	"opts": {
		"max_open_conns": 100,					// maximum database connections opened, not applying for mongo
		"max_idle_conns": 10,					// maximum database connections idle, not applying for mongo
		"conn_max_lifetime": 0, 				// maximum amount of time in seconds a connection may be reused (0 for unlimited), not applying for mongo
	},
	"string_indexed_fields": [],			// indexes on cdrs table to speed up queries, used in case of *mongo and *internal
	"prefix_indexed_fields":[],				// prefix indexes on cdrs table to speed up queries, used in case of *internal
	}
}`
	expected = StorDbCfg{
		Type:     "mysql",
		Host:     "127.0.0.1",
		Port:     "3306",
		Name:     "cgrates",
		User:     "cgrates",
		Password: "password",
		Opts: map[string]interface{}{
			utils.MaxOpenConnsCfg:    100.,
			utils.MaxIdleConnsCfg:    10.,
			utils.ConnMaxLifetimeCfg: 0.,
		},
		StringIndexedFields: []string{},
		PrefixIndexedFields: []string{},
	}
	dbcfg.Opts = make(map[string]interface{})
	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnStoreDbCfg, err := jsnCfg.DbJsonCfg(STORDB_JSN); err != nil {
		t.Error(err)
	} else if err = dbcfg.loadFromJsonCfg(jsnStoreDbCfg); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, dbcfg) {
		t.Errorf("Expected: %+v , recived: %+v", utils.ToJSON(expected), utils.ToJSON(dbcfg))
	}
}
func TestStoreDbCfgloadFromJsonCfgPort(t *testing.T) {
	var dbcfg StorDbCfg
	cfgJSONStr := `{
"stor_db": {
	"db_type": "mongo",
	}
}`
	dbcfg.Opts = make(map[string]interface{})
	expected := StorDbCfg{
		Type: "mongo",
		Opts: make(map[string]interface{}),
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
		Opts: make(map[string]interface{}),
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
		Opts: make(map[string]interface{}),
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
	dbcfg.Items = make(map[string]*ItemOpt)
	dbcfg.Opts = make(map[string]interface{})
	cfgJSONStr := `{
		"stor_db": {								
			"db_type": "*mysql",					
			"db_host": "127.0.0.1",					
			"db_port": -1,						
			"db_name": "cgrates",					
			"db_user": "cgrates",					
			"db_password": "",						
			"string_indexed_fields": [],			
			"prefix_indexed_fields":[],	
			"opts": {	
				"max_open_conns": 100,					
				"max_idle_conns": 10,					
				"conn_max_lifetime": 0, 			
				"query_timeout":"10s",
				"sslmode":"disable",					
			},
			"items":{
				"session_costs": {}, 
				"cdrs": {}, 		
			},
		},
}`

	eMap := map[string]interface{}{
		utils.DataDbTypeCfg:          "*mysql",
		utils.DataDbHostCfg:          "127.0.0.1",
		utils.DataDbPortCfg:          3306,
		utils.DataDbNameCfg:          "cgrates",
		utils.DataDbUserCfg:          "cgrates",
		utils.DataDbPassCfg:          "",
		utils.StringIndexedFieldsCfg: []string{},
		utils.PrefixIndexedFieldsCfg: []string{},
		utils.OptsCfg: map[string]interface{}{
			utils.MaxOpenConnsCfg:    100.,
			utils.MaxIdleConnsCfg:    10.,
			utils.ConnMaxLifetimeCfg: 0.,
			utils.QueryTimeoutCfg:    "10s",
			utils.SSLModeCfg:         "disable",
		},
		utils.ItemsCfg: map[string]interface{}{
			utils.SessionCostsTBL: map[string]interface{}{utils.RemoteCfg: false, utils.ReplicateCfg: false},
			utils.CdrsCfg:         map[string]interface{}{utils.RemoteCfg: false, utils.ReplicateCfg: false},
		},
	}
	if cfgCgr, err := NewCGRConfigFromJsonStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else {
		rcv := cfgCgr.storDbCfg.AsMapInterface()
		if !reflect.DeepEqual(eMap[utils.ItemsCfg].(map[string]interface{})[utils.SessionSConnsCfg],
			rcv[utils.ItemsCfg].(map[string]interface{})[utils.SessionSConnsCfg]) {
			t.Errorf("Expected %+v, received %+v", eMap[utils.ItemsCfg].(map[string]interface{})[utils.SessionSConnsCfg],
				rcv[utils.ItemsCfg].(map[string]interface{})[utils.SessionSConnsCfg])
		} else if !reflect.DeepEqual(eMap[utils.OptsCfg], rcv[utils.OptsCfg]) {
			t.Errorf("Expected %+v \n, received %+v", eMap[utils.OptsCfg], rcv[utils.OptsCfg])
		} else if !reflect.DeepEqual(eMap[utils.PrefixIndexedFieldsCfg], rcv[utils.PrefixIndexedFieldsCfg]) {
			t.Errorf("Expected %+v \n, received %+v", eMap[utils.PrefixIndexedFieldsCfg], rcv[utils.PrefixIndexedFieldsCfg])
		}
	}
}
