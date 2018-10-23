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
	"cdrs_indexes": [],						// indexes on cdrs table to speed up queries, used only in case of mongo
	}
}`
	expected = StorDbCfg{
		StorDBType:            "mysql",
		StorDBHost:            "127.0.0.1",
		StorDBPort:            "3306",
		StorDBName:            "cgrates",
		StorDBUser:            "cgrates",
		StorDBPass:            "password",
		StorDBMaxOpenConns:    100,
		StorDBMaxIdleConns:    10,
		StorDBConnMaxLifetime: 0,
		StorDBCDRSIndexes:     []string{},
	}
	if jsnCfg, err := NewCgrJsonCfgFromReader(strings.NewReader(cfgJSONStr)); err != nil {
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
		StorDBType: "mongo",
	}
	if jsnCfg, err := NewCgrJsonCfgFromReader(strings.NewReader(cfgJSONStr)); err != nil {
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
		StorDBType: "mongo",
		StorDBPort: "27017",
	}
	if jsnCfg, err := NewCgrJsonCfgFromReader(strings.NewReader(cfgJSONStr)); err != nil {
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
		StorDBType: "internal",
		StorDBPort: "internal",
	}
	if jsnCfg, err := NewCgrJsonCfgFromReader(strings.NewReader(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnDataDbCfg, err := jsnCfg.DbJsonCfg(STORDB_JSN); err != nil {
		t.Error(err)
	} else if err = dbcfg.loadFromJsonCfg(jsnDataDbCfg); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, dbcfg) {
		t.Errorf("Expected: %+v , recived: %+v", expected, dbcfg)
	}
}
