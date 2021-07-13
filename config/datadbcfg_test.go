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

func TestDataDBRemoteReplication(t *testing.T) {
	var dbcfg, expected DataDbCfg
	cfgJSONStr := `{
"data_db": {								// database used to store runtime data (eg: accounts, cdr stats)
	"db_type": "*redis",					// data_db type: <*redis|*mongo|*internal>
	"db_host": "127.0.0.1",					// data_db host address
	"db_port": -1,	 						// data_db port to reach the database
	"db_name": "10", 						// data_db database name to connect to
	"db_user": "cgrates", 					// username to use when connecting to data_db
	"db_password": "password",				// password to use when connecting to data_db
	"redis_sentinel":"sentinel",			// redis_sentinel is the name of sentinel
	"remote_conns":["Conn1"],
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
		RmtConns:           []string{"Conn1"},
	}
	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnDataDbCfg, err := jsnCfg.DbJsonCfg(DATADB_JSN); err != nil {
		t.Error(err)
	} else if err = dbcfg.loadFromJsonCfg(jsnDataDbCfg); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, dbcfg) {
		t.Errorf("Expected: %+v ,\n recived: %+v", utils.ToJSON(expected), utils.ToJSON(dbcfg))
	}

	cfgJSONStr = `{
"data_db": {								// database used to store runtime data (eg: accounts, cdr stats)
	"db_type": "*internal",					// data_db type: <*redis|*mongo|*internal>
	"remote_conns":["Conn1"],
	"replication_conns":["Conn2"],
	}
}`

	expected = DataDbCfg{
		DataDbType:         utils.INTERNAL,
		DataDbHost:         "127.0.0.1",
		DataDbPort:         "6379",
		DataDbName:         "10",
		DataDbUser:         "cgrates",
		DataDbPass:         "password",
		DataDbSentinelName: "sentinel",
		RmtConns:           []string{"Conn1"},
		RplConns:           []string{"Conn2"},
	}
	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnDataDbCfg, err := jsnCfg.DbJsonCfg(DATADB_JSN); err != nil {
		t.Error(err)
	} else if err = dbcfg.loadFromJsonCfg(jsnDataDbCfg); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, dbcfg) {
		t.Errorf("Expected: %+v ,\n recived: %+v", utils.ToJSON(expected), utils.ToJSON(dbcfg))
	}

	cfgJSONStr = `{
"data_db": {								// database used to store runtime data (eg: accounts, cdr stats)
	"db_type": "*internal",					// data_db type: <*redis|*mongo|*internal>
	"remote_conns":["Conn1","Conn2","Conn3"],
	"replication_conns":["Conn4","Conn5"],
	}
}`

	expected = DataDbCfg{
		DataDbType:         utils.INTERNAL,
		DataDbHost:         "127.0.0.1",
		DataDbPort:         "6379",
		DataDbName:         "10",
		DataDbUser:         "cgrates",
		DataDbPass:         "password",
		DataDbSentinelName: "sentinel",
		RmtConns:           []string{"Conn1", "Conn2", "Conn3"},
		RplConns:           []string{"Conn4", "Conn5"},
	}
	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnDataDbCfg, err := jsnCfg.DbJsonCfg(DATADB_JSN); err != nil {
		t.Error(err)
	} else if err = dbcfg.loadFromJsonCfg(jsnDataDbCfg); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, dbcfg) {
		t.Errorf("Expected: %+v ,\n recived: %+v", utils.ToJSON(expected), utils.ToJSON(dbcfg))
	}
}

func TestDataDbCfgloadFromJsonCfgItems(t *testing.T) {
	var dbcfg, expected DataDbCfg
	cfgJSONStr := `{
"data_db": {								// database used to store runtime data (eg: accounts, cdr stats)
	"db_type": "*redis",					// data_db type: <*redis|*mongo|*internal>
	"db_host": "127.0.0.1",					// data_db host address
	"db_port": -1,	 						// data_db port to reach the database
	"db_name": "10", 						// data_db database name to connect to
	"db_user": "cgrates", 					// username to use when connecting to data_db
	"db_password": "password",				// password to use when connecting to data_db
	"redis_sentinel":"sentinel",			// redis_sentinel is the name of sentinel
	"remote_conns":["Conn1"],
    "items":{
		"*accounts":{"replicate":true, "limit": 5,"ttl": "6"},
		"*reverse_destinations": {"replicate":false},
		"*destinations": {"replicate":false},
	  }	
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
		RmtConns:           []string{"Conn1"},
		Items: map[string]*ItemOpt{
			utils.MetaAccounts: &ItemOpt{
				Replicate: true,
				TTL:       6,
				Limit:     5,
			},
			utils.MetaReverseDestinations: &ItemOpt{
				Replicate: false,
			},
			utils.MetaDestinations: &ItemOpt{
				Replicate: false,
			},
		},
	}
	dbcfg.Items = make(map[string]*ItemOpt)
	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnDataDbCfg, err := jsnCfg.DbJsonCfg(DATADB_JSN); err != nil {
		t.Error(err)
	} else if err = dbcfg.loadFromJsonCfg(jsnDataDbCfg); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, dbcfg) {
		t.Errorf("Expected: %+v ,\n recived: %+v", utils.ToJSON(expected), utils.ToJSON(dbcfg))
	}

	cfgJSONStr = `{
		"data_db": {								// database used to store runtime data (eg: accounts, cdr stats)
			"db_type": "*redis",					// data_db type: <*redis|*mongo|*internal>
			"db_host": "127.0.0.1",					// data_db host address
			"db_port": -1,	 						// data_db port to reach the database
			"db_name": "10", 						// data_db database name to connect to
			"db_user": "cgrates", 					// username to use when connecting to data_db
			"db_password": "password",				// password to use when connecting to data_db
			"redis_sentinel":"sentinel",			// redis_sentinel is the name of sentinel
			"remote_conns":["Conn1"],
			"items":{
				"*dispatcher_hosts":{"remote":true, "replicate":true, "limit": -1, "ttl": "", "static_ttl": true}, 
				"*filter_indexes" :{"remote":true, "replicate":true, "limit": -1, "ttl": "", "static_ttl": true}, 
				"*load_ids":{"remote":true, "replicate":true, "limit": -1, "ttl": "", "static_ttl": true}, 
			
			  }	
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
		RmtConns:           []string{"Conn1"},
		Items: map[string]*ItemOpt{
			utils.MetaDispatcherHosts: &ItemOpt{
				Remote:    true,
				Replicate: true,
				Limit:     -1,
				StaticTTL: true,
			},
			utils.MetaFilterIndexes: &ItemOpt{
				Remote:    true,
				Replicate: true,
				Limit:     -1,
				StaticTTL: true,
			},
			utils.MetaLoadIDs: &ItemOpt{
				Remote:    true,
				Replicate: true,
				Limit:     -1,
				StaticTTL: true,
			},
		},
	}
	dbcfg.Items = make(map[string]*ItemOpt)
	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnDataDbCfg, err := jsnCfg.DbJsonCfg(DATADB_JSN); err != nil {
		t.Error(err)
	} else if err = dbcfg.loadFromJsonCfg(jsnDataDbCfg); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, dbcfg) {
		t.Errorf("Expected: %+v ,\n recived: %+v", utils.ToJSON(expected), utils.ToJSON(dbcfg))
	}

	cfgJSONStr = `{
		"data_db": {								// database used to store runtime data (eg: accounts, cdr stats)
			"db_type": "*redis",					// data_db type: <*redis|*mongo|*internal>
			"db_host": "127.0.0.1",					// data_db host address
			"db_port": -1,	 						// data_db port to reach the database
			"db_name": "10", 						// data_db database name to connect to
			"db_user": "cgrates", 					// username to use when connecting to data_db
			"db_password": "password",				// password to use when connecting to data_db
			"redis_sentinel":"sentinel",			// redis_sentinel is the name of sentinel
			"remote_conns":["Conn1"],
			"items":{
				"*timings": {"remote":false, "replicate":false, "limit": 9, "ttl": "8", "static_ttl": true}, 
				"*resource_profiles":{"remote":false, "replicate":false, "limit": 9, "ttl": "8", "static_ttl": true}, 
				"*resources":{"remote":false, "replicate":false, "limit": 9, "ttl": "8", "static_ttl": true}, 
				"*statqueue_profiles": {"remote":false, "replicate":false, "limit": 9, "ttl": "8", "static_ttl": true}, 
			  }	
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
		RmtConns:           []string{"Conn1"},
		Items: map[string]*ItemOpt{
			utils.MetaTimings: &ItemOpt{
				Limit:     9,
				TTL:       8,
				StaticTTL: true,
			},
			utils.MetaResourceProfile: &ItemOpt{
				Limit:     9,
				TTL:       8,
				StaticTTL: true,
			},
			utils.MetaResources: &ItemOpt{
				Limit:     9,
				TTL:       8,
				StaticTTL: true,
			},
			utils.MetaStatQueueProfiles: &ItemOpt{
				Limit:     9,
				TTL:       8,
				StaticTTL: true,
			},
		},
	}
	dbcfg.Items = make(map[string]*ItemOpt)
	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnDataDbCfg, err := jsnCfg.DbJsonCfg(DATADB_JSN); err != nil {
		t.Error(err)
	} else if err = dbcfg.loadFromJsonCfg(jsnDataDbCfg); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, dbcfg) {
		t.Errorf("Expected: %+v ,\n recived: %+v", utils.ToJSON(expected), utils.ToJSON(dbcfg))
	}
}

func TestDataDbCfgAsMapInterface(t *testing.T) {
	var dbcfg DataDbCfg
	dbcfg.Items = make(map[string]*ItemOpt)
	cfgJSONStr := `{
	"data_db": {								
		"db_type": "*redis",					
		"db_host": "127.0.0.1",					
		"db_port": 6379, 						
		"db_name": "10", 						
		"db_user": "cgrates", 					
		"db_password": "", 						
		"redis_sentinel":"",					
		"query_timeout":"10s",
		"remote_conns":[],
		"replication_conns":[],
		"items":{
			"*accounts":{"remote":true, "replicate":false, "limit": -1, "ttl": "", "static_ttl": false}, 					
			"*reverse_destinations": {"remote":false, "replicate":false, "limit": 7, "ttl": "", "static_ttl": true},
		},
	},		
}`
	eMap := map[string]interface{}{
		"db_type":              "*redis",
		"db_host":              "127.0.0.1",
		"db_port":              6379,
		"db_name":              "10",
		"db_user":              "cgrates",
		"db_password":          "",
		"redis_sentinel":       "",
		"query_timeout":        "10s",
		"remote_conns":         []string{},
		"replication_conns":    []string{},
		"replication_filtered": false,
		"items": map[string]interface{}{
			"*accounts":             map[string]interface{}{"remote": true, "replicate": false, "limit": -1, "ttl": "", "static_ttl": false},
			"*reverse_destinations": map[string]interface{}{"remote": false, "replicate": false, "limit": 7, "ttl": "", "static_ttl": true},
		},
	}
	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnDataDbCfg, err := jsnCfg.DbJsonCfg(DATADB_JSN); err != nil {
		t.Error(err)
	} else if err = dbcfg.loadFromJsonCfg(jsnDataDbCfg); err != nil {
		t.Error(err)
	} else if rcv := dbcfg.AsMapInterface(); !reflect.DeepEqual(eMap, rcv) {
		t.Errorf("Expected: %+v ,\n recived: %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}
}
