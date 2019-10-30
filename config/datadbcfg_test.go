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
	"time"

	"github.com/cgrates/cgrates/utils"
)

func TestDataDbCfgloadFromJsonCfg(t *testing.T) {
	var dbcfg, expected DataDbCfg
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
"data_db": {								// database used to store runtime data (eg: accounts, cdr stats)
	"db_type": "*redis",					// data_db type: <*redis|*mongo|*internal>
	"db_host": "127.0.0.1",					// data_db host address
	"db_port": -1,	 						// data_db port to reach the database
	"db_name": "10", 						// data_db database name to connect to
	"db_user": "cgrates", 					// username to use when connecting to data_db
	"db_password": "password",				// password to use when connecting to data_db
	"redis_sentinel":"sentinel",			// redis_sentinel is the name of sentinel
	}
}`
	expected = DataDbCfg{
		DataDbType:         "redis",
		DataDbHost:         "127.0.0.1",
		DataDbPort:         "6379",
		DataDbName:         "10",
		DataDbUser:         "cgrates",
		DataDbPass:         "password",
		DataDbSentinelName: "sentinel",
	}
	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnDataDbCfg, err := jsnCfg.DbJsonCfg(DATADB_JSN); err != nil {
		t.Error(err)
	} else if err = dbcfg.loadFromJsonCfg(jsnDataDbCfg); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, dbcfg) {
		t.Errorf("Expected: %+v , recived: %+v", expected, dbcfg)
	}
}
func TestDataDbCfgloadFromJsonCfgPort(t *testing.T) {
	var dbcfg DataDbCfg
	cfgJSONStr := `{
"data_db": {
	"db_type": "mongo",
	}
}`
	expected := DataDbCfg{
		DataDbType: "mongo",
	}
	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnDataDbCfg, err := jsnCfg.DbJsonCfg(DATADB_JSN); err != nil {
		t.Error(err)
	} else if err = dbcfg.loadFromJsonCfg(jsnDataDbCfg); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, dbcfg) {
		t.Errorf("Expected: %+v , recived: %+v", expected, dbcfg)
	}
	cfgJSONStr = `{
"data_db": {
	"db_type": "mongo",
	"db_port": -1,
	}
}`
	expected = DataDbCfg{
		DataDbType: "mongo",
		DataDbPort: "27017",
	}
	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnDataDbCfg, err := jsnCfg.DbJsonCfg(DATADB_JSN); err != nil {
		t.Error(err)
	} else if err = dbcfg.loadFromJsonCfg(jsnDataDbCfg); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, dbcfg) {
		t.Errorf("Expected: %+v , recived: %+v", expected, dbcfg)
	}
	cfgJSONStr = `{
"data_db": {
	"db_type": "*internal",
	"db_port": -1,
	}
}`
	expected = DataDbCfg{
		DataDbType: "internal",
		DataDbPort: "internal",
	}
	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnDataDbCfg, err := jsnCfg.DbJsonCfg(DATADB_JSN); err != nil {
		t.Error(err)
	} else if err = dbcfg.loadFromJsonCfg(jsnDataDbCfg); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, dbcfg) {
		t.Errorf("Expected: %+v , recived: %+v", expected, dbcfg)
	}
}

func TestDataDbNewDataDbFromUrl(t *testing.T) {
	if _, err := newDataDBCfgFromUrl(utils.EmptyString); err != utils.ErrMandatoryIeMissing {
		t.Errorf("Expected: %+v , recived: %+v", utils.ErrMandatoryIeMissing, err)
	}

	url := "127.0.0.1:1234"
	expected := &DataDbCfg{
		DataDbType:         utils.REDIS,
		DataDbHost:         "127.0.0.1",
		DataDbPort:         "1234",
		DataDbName:         "10",
		DataDbUser:         "cgrates",
		DataDbPass:         "",
		DataDbSentinelName: "",
		QueryTimeout:       10 * time.Second,
		RmtDataDBCfgs:      []*DataDbCfg{},
		RplDataDBCfgs:      []*DataDbCfg{},
	}
	if rcv, err := newDataDBCfgFromUrl(url); err != nil || !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Error: %+v \n, expected: %+v ,\n recived: %+v", err, utils.ToJSON(expected), utils.ToJSON(rcv))
	}

	url = "127.0.0.1:1234/?user=test&pass=test"
	expected = &DataDbCfg{
		DataDbType:         utils.REDIS,
		DataDbHost:         "127.0.0.1",
		DataDbPort:         "1234",
		DataDbName:         "10",
		DataDbUser:         "test",
		DataDbPass:         "test",
		DataDbSentinelName: "",
		QueryTimeout:       10 * time.Second,
		RmtDataDBCfgs:      []*DataDbCfg{},
		RplDataDBCfgs:      []*DataDbCfg{},
	}
	if rcv, err := newDataDBCfgFromUrl(url); err != nil || !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Error: %+v , expected: %+v , recived: %+v", err, utils.ToJSON(expected), utils.ToJSON(rcv))
	}

	url = "0.0.0.0:1234/?type=*mongo"
	expected = &DataDbCfg{
		DataDbType:         utils.MONGO,
		DataDbHost:         "0.0.0.0",
		DataDbPort:         "1234",
		DataDbName:         "10",
		DataDbUser:         "cgrates",
		DataDbPass:         "",
		DataDbSentinelName: "",
		QueryTimeout:       10 * time.Second,
		RmtDataDBCfgs:      []*DataDbCfg{},
		RplDataDBCfgs:      []*DataDbCfg{},
	}
	if rcv, err := newDataDBCfgFromUrl(url); err != nil || !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Error: %+v , expected: %+v , recived: %+v", err, utils.ToJSON(expected), utils.ToJSON(rcv))
	}

	url = "0.0.0.0:1234/?type=*mongo&query=error"
	if _, err := newDataDBCfgFromUrl(url); err == nil || err.Error() != "time: invalid duration error" {
		t.Errorf("Expected:<time: invalid duration error> , recived: <%+v>", err)
	}
}
