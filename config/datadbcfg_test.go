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
	jsonCfg := &DbJsonCfg{
		Db_type:           utils.StringPointer("redis"),
		Db_host:           utils.StringPointer("127.0.0.1"),
		Db_port:           utils.IntPointer(6379),
		Db_name:           utils.StringPointer("10"),
		Db_user:           utils.StringPointer("cgrates"),
		Db_password:       utils.StringPointer("password"),
		Remote_conns:      &[]string{"*conn1"},
		Replication_conns: &[]string{"*conn1"},
		Items: map[string]*ItemOptJson{
			utils.MetaAccounts: {
				Replicate: utils.BoolPointer(true),
				Remote:    utils.BoolPointer(true),
			},
		},
		Opts: map[string]interface{}{
			utils.RedisSentinelNameCfg: "sentinel",
		},
	}
	expected := &DataDbCfg{
		Type:     "redis",
		Host:     "127.0.0.1",
		Port:     "6379",
		Name:     "10",
		User:     "cgrates",
		Password: "password",
		RmtConns: []string{"*conn1"},
		RplConns: []string{"*conn1"},
		Items: map[string]*ItemOpt{
			utils.MetaAccounts: {
				Replicate: true,
				Remote:    true,
			},
		},
		Opts: map[string]interface{}{
			utils.RedisSentinelNameCfg: "sentinel",
		},
	}
	jsnCfg := NewDefaultCGRConfig()
	if err = jsnCfg.dataDbCfg.loadFromJSONCfg(nil); err != nil {
		t.Error(err)
	} else if err = jsnCfg.dataDbCfg.loadFromJSONCfg(jsonCfg); err != nil {
		t.Error(err)
	} else {
		if !reflect.DeepEqual(expected.Items[utils.MetaAccounts], jsnCfg.dataDbCfg.Items[utils.MetaAccounts]) {
			t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected.Items[utils.MetaAccounts]),
				utils.ToJSON(jsnCfg.dataDbCfg.Items[utils.MetaAccounts]))
		} else if !reflect.DeepEqual(expected.Opts[utils.RedisSentinelNameCfg], jsnCfg.dataDbCfg.Opts[utils.RedisSentinelNameCfg]) {
			t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected.Opts[utils.RedisSentinelNameCfg]),
				utils.ToJSON(jsnCfg.dataDbCfg.Opts[utils.RedisSentinelNameCfg]))
		} else if !reflect.DeepEqual(expected.RplConns, jsnCfg.dataDbCfg.RplConns) {
			t.Errorf("Expected %+v \n, received %+v", expected.RplConns, jsnCfg.dataDbCfg.RplConns)
		}
	}
}

func TestConnsloadFromJsonCfg(t *testing.T) {
	jsonCfg := &DbJsonCfg{
		Remote_conns: &[]string{"*internal"},
	}
	expectedErrRmt := "Remote connection ID needs to be different than <*internal> "
	jsnCfg := NewDefaultCGRConfig()
	if err = jsnCfg.dataDbCfg.loadFromJSONCfg(jsonCfg); err == nil || err.Error() != expectedErrRmt {
		t.Errorf("Expected %+v, received %+v", expectedErrRmt, err)
	}

	jsonCfg = &DbJsonCfg{
		Replication_conns: &[]string{"*internal"},
	}
	expectedErrRpl := "Remote connection ID needs to be different than <*internal> "
	jsnCfg = NewDefaultCGRConfig()
	if err = jsnCfg.dataDbCfg.loadFromJSONCfg(jsonCfg); err == nil || err.Error() != expectedErrRpl {
		t.Errorf("Expected %+v, received %+v", expectedErrRpl, err)
	}
}

func TestItemCfgloadFromJson(t *testing.T) {
	jsonCfg := &ItemOptJson{
		Remote:    utils.BoolPointer(true),
		Replicate: utils.BoolPointer(true),
		Api_key:   utils.StringPointer("randomVal"),
		Route_id:  utils.StringPointer("randomID"),
	}
	expected := &ItemOpt{
		Remote:    true,
		Replicate: true,
		APIKey:    "randomVal",
		RouteID:   "randomID",
	}
	rcv := new(ItemOpt)
	rcv.loadFromJSONCfg(nil)
	rcv.loadFromJSONCfg(jsonCfg)
	if !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}

func TestDataDbCfgloadFromJsonCfgPort(t *testing.T) {
	var dbcfg DataDbCfg
	dbcfg.Opts = make(map[string]interface{})
	cfgJSONStr := `{
"data_db": {
	"db_type": "mongo",
	}
}`
	expected := DataDbCfg{
		Type: "mongo",
		Opts: make(map[string]interface{}),
	}
	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnDataDbCfg, err := jsnCfg.DbJsonCfg(DataDBJSON); err != nil {
		t.Error(err)
	} else if err = dbcfg.loadFromJSONCfg(jsnDataDbCfg); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, dbcfg) {
		t.Errorf("Expected: %+v , received: %+v", expected, dbcfg)
	}
	cfgJSONStr = `{
"data_db": {
	"db_type": "mongo",
	"db_port": -1,
	}
}`
	expected = DataDbCfg{
		Type: "mongo",
		Port: "27017",
		Opts: make(map[string]interface{}),
	}
	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnDataDbCfg, err := jsnCfg.DbJsonCfg(DataDBJSON); err != nil {
		t.Error(err)
	} else if err = dbcfg.loadFromJSONCfg(jsnDataDbCfg); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, dbcfg) {
		t.Errorf("Expected: %+v , received: %+v", expected, dbcfg)
	}
	cfgJSONStr = `{
"data_db": {
	"db_type": "*internal",
	"db_port": -1,
	}
}`
	expected = DataDbCfg{
		Type: "internal",
		Port: "internal",
		Opts: make(map[string]interface{}),
	}
	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnDataDbCfg, err := jsnCfg.DbJsonCfg(DataDBJSON); err != nil {
		t.Error(err)
	} else if err = dbcfg.loadFromJSONCfg(jsnDataDbCfg); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, dbcfg) {
		t.Errorf("Expected: %+v , received: %+v", expected, dbcfg)
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
	"opts":{
		"redisSentinel":"sentinel",			// redisSentinel is the name of sentinel
	},
	"remote_conns":["Conn1"],
	}
}`

	dbcfg.Opts = make(map[string]interface{})
	expected = DataDbCfg{
		Type:     "redis",
		Host:     "127.0.0.1",
		Port:     "6379",
		Name:     "10",
		User:     "cgrates",
		Password: "password",
		RmtConns: []string{"Conn1"},
		Opts: map[string]interface{}{
			utils.RedisSentinelNameCfg: "sentinel",
		},
	}
	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnDataDbCfg, err := jsnCfg.DbJsonCfg(DataDBJSON); err != nil {
		t.Error(err)
	} else if err = dbcfg.loadFromJSONCfg(jsnDataDbCfg); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, dbcfg) {
		t.Errorf("Expected: %+v ,\n received: %+v", utils.ToJSON(expected), utils.ToJSON(dbcfg))
	}

	cfgJSONStr = `{
"data_db": {								// database used to store runtime data (eg: accounts, cdr stats)
	"db_type": "*internal",					// data_db type: <*redis|*mongo|*internal>
	"remote_conns":["Conn1"],
	"replication_conns":["Conn2"],
	}
}`

	expected = DataDbCfg{
		Type:     utils.Internal,
		Host:     "127.0.0.1",
		Port:     "6379",
		Name:     "10",
		User:     "cgrates",
		Password: "password",
		Opts: map[string]interface{}{
			utils.RedisSentinelNameCfg: "sentinel",
		},
		RmtConns: []string{"Conn1"},
		RplConns: []string{"Conn2"},
	}
	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnDataDbCfg, err := jsnCfg.DbJsonCfg(DataDBJSON); err != nil {
		t.Error(err)
	} else if err = dbcfg.loadFromJSONCfg(jsnDataDbCfg); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, dbcfg) {
		t.Errorf("Expected: %+v ,\n received: %+v", utils.ToJSON(expected), utils.ToJSON(dbcfg))
	}

	cfgJSONStr = `{
"data_db": {								// database used to store runtime data (eg: accounts, cdr stats)
	"db_type": "*internal",					// data_db type: <*redis|*mongo|*internal>
	"remote_conns":["Conn1","Conn2","Conn3"],
	"replication_conns":["Conn4","Conn5"],
	}
}`

	expected = DataDbCfg{
		Type:     utils.Internal,
		Host:     "127.0.0.1",
		Port:     "6379",
		Name:     "10",
		User:     "cgrates",
		Password: "password",
		Opts: map[string]interface{}{
			utils.RedisSentinelNameCfg: "sentinel",
		},
		RmtConns: []string{"Conn1", "Conn2", "Conn3"},
		RplConns: []string{"Conn4", "Conn5"},
	}
	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnDataDbCfg, err := jsnCfg.DbJsonCfg(DataDBJSON); err != nil {
		t.Error(err)
	} else if err = dbcfg.loadFromJSONCfg(jsnDataDbCfg); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, dbcfg) {
		t.Errorf("Expected: %+v ,\n received: %+v", utils.ToJSON(expected), utils.ToJSON(dbcfg))
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
	"remote_conns":["Conn1"],
    "items":{
		"*accounts":{"replicate":true},
	  }	,
	"opts": {
		"redisSentinel":"sentinel",			// redisSentinel is the name of sentinel
	  }
	},
}`

	expected = DataDbCfg{
		Type:     "redis",
		Host:     "127.0.0.1",
		Port:     "6379",
		Name:     "10",
		User:     "cgrates",
		Password: "password",
		RmtConns: []string{"Conn1"},
		Items: map[string]*ItemOpt{
			utils.MetaAccounts: {
				Replicate: true,
			},
		},
		Opts: map[string]interface{}{
			utils.RedisSentinelNameCfg: "sentinel",
		},
	}
	dbcfg.Items = make(map[string]*ItemOpt)
	dbcfg.Opts = make(map[string]interface{})
	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnDataDbCfg, err := jsnCfg.DbJsonCfg(DataDBJSON); err != nil {
		t.Error(err)
	} else if err = dbcfg.loadFromJSONCfg(jsnDataDbCfg); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, dbcfg) {
		t.Errorf("Expected: %+v ,\n received: %+v", utils.ToJSON(expected), utils.ToJSON(dbcfg))
	}

	cfgJSONStr = `{
		"data_db": {								// database used to store runtime data (eg: accounts, cdr stats)
			"db_type": "*redis",					// data_db type: <*redis|*mongo|*internal>
			"db_host": "127.0.0.1",					// data_db host address
			"db_port": -1,	 						// data_db port to reach the database
			"db_name": "10", 						// data_db database name to connect to
			"db_user": "cgrates", 					// username to use when connecting to data_db
			"db_password": "password",				// password to use when connecting to data_db
			"opts": {
				"redisSentinel":"sentinel",			// redisSentinel is the name of sentinel
			},
			"remote_conns":["Conn1"],
			"items":{
				"*dispatcher_hosts":{"remote":true, "replicate":true}, 
				"*indexes" :{"remote":true, "replicate":true}, 
				"*load_ids":{"remote":true, "replicate":true}, 
			
			  }	
			}
		}`

	expected = DataDbCfg{
		Type:     "redis",
		Host:     "127.0.0.1",
		Port:     "6379",
		Name:     "10",
		User:     "cgrates",
		Password: "password",
		Opts: map[string]interface{}{
			utils.RedisSentinelNameCfg: "sentinel",
		},
		RmtConns: []string{"Conn1"},
		Items: map[string]*ItemOpt{
			utils.MetaDispatcherHosts: {
				Remote:    true,
				Replicate: true,
			},
			utils.MetaIndexes: {
				Remote:    true,
				Replicate: true,
			},
			utils.MetaLoadIDs: {
				Remote:    true,
				Replicate: true,
			},
		},
	}
	dbcfg.Items = make(map[string]*ItemOpt)
	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnDataDbCfg, err := jsnCfg.DbJsonCfg(DataDBJSON); err != nil {
		t.Error(err)
	} else if err = dbcfg.loadFromJSONCfg(jsnDataDbCfg); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, dbcfg) {
		t.Errorf("Expected: %+v ,\n received: %+v", utils.ToJSON(expected), utils.ToJSON(dbcfg))
	}

	cfgJSONStr = `{
		"data_db": {								// database used to store runtime data (eg: accounts, cdr stats)
			"db_type": "*redis",					// data_db type: <*redis|*mongo|*internal>
			"db_host": "127.0.0.1",					// data_db host address
			"db_port": -1,	 						// data_db port to reach the database
			"db_name": "10", 						// data_db database name to connect to
			"db_user": "cgrates", 					// username to use when connecting to data_db
			"db_password": "password",				// password to use when connecting to data_db
			"opts": {
				"redisSentinel":"sentinel",			// redisSentinel is the name of sentinel
			},
			"remote_conns":["Conn1"],
			"items":{
				"*resource_profiles":{"remote":false, "replicate":false}, 
				"*resources":{"remote":false, "replicate":false}, 
				"*statqueue_profiles": {"remote":false, "replicate":false}, 
			  }	
			}
		}`

	expected = DataDbCfg{
		Type:     "redis",
		Host:     "127.0.0.1",
		Port:     "6379",
		Name:     "10",
		User:     "cgrates",
		Password: "password",
		Opts: map[string]interface{}{
			utils.RedisSentinelNameCfg: "sentinel",
		},
		RmtConns: []string{"Conn1"},
		Items: map[string]*ItemOpt{
			utils.MetaResourceProfile:   {},
			utils.MetaResources:         {},
			utils.MetaStatQueueProfiles: {},
		},
	}
	dbcfg.Items = make(map[string]*ItemOpt)
	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnDataDbCfg, err := jsnCfg.DbJsonCfg(DataDBJSON); err != nil {
		t.Error(err)
	} else if err = dbcfg.loadFromJSONCfg(jsnDataDbCfg); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, dbcfg) {
		t.Errorf("Expected: %+v ,\n received: %+v", utils.ToJSON(expected), utils.ToJSON(dbcfg))
	}
}

func TestDataDbCfgAsMapInterface(t *testing.T) {
	cfgJSONStr := `{
	"data_db": {								
		"db_type": "*redis",					
		"db_host": "127.0.0.1",					
		"db_port": 6379, 						
		"db_name": "10", 						
		"db_user": "cgrates", 					
		"db_password": "", 						
		"opts": {
			"redisSentinel":"",					
			"mongoQueryTimeout":"10s",
		},
		"remote_conns":[],
		"replication_conns":[],
		"items":{
			"*accounts":{"remote":true, "replicate":false, "api_key": "randomVal", "route_id": "randomVal"}, 					
			"*reverse_destinations": {"remote":false, "replicate":false, "api_key": "randomVal", "route_id": "randomVal"},
		},
	},		
}`
	eMap := map[string]interface{}{
		utils.DataDbTypeCfg: "*redis",
		utils.DataDbHostCfg: "127.0.0.1",
		utils.DataDbPortCfg: 6379,
		utils.DataDbNameCfg: "10",
		utils.DataDbUserCfg: "cgrates",
		utils.DataDbPassCfg: "",
		utils.OptsCfg: map[string]interface{}{
			utils.RedisSentinelNameCfg: "",
			utils.MongoQueryTimeoutCfg: "10s",
		},
		utils.RemoteConnsCfg:      []string{},
		utils.ReplicationConnsCfg: []string{},
		utils.ItemsCfg: map[string]interface{}{
			utils.MetaAccounts: map[string]interface{}{utils.RemoteCfg: true, utils.ReplicateCfg: false, utils.APIKeyCfg: "randomVal", utils.RouteIDCfg: "randomVal"},
		},
	}
	if cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else {
		rcv := cgrCfg.dataDbCfg.AsMapInterface()
		if !reflect.DeepEqual(eMap[utils.ItemsCfg].(map[string]interface{})[utils.MetaAccounts],
			rcv[utils.ItemsCfg].(map[string]interface{})[utils.MetaAccounts]) {
			t.Errorf("Expected %+v, received %+v", eMap[utils.ItemsCfg].(map[string]interface{})[utils.MetaAccounts],
				rcv[utils.ItemsCfg].(map[string]interface{})[utils.MetaAccounts])
		}
	}
}

func TestCloneDataDB(t *testing.T) {
	jsonCfg := &DbJsonCfg{
		Db_type:           utils.StringPointer("redis"),
		Db_host:           utils.StringPointer("127.0.0.1"),
		Db_port:           utils.IntPointer(6379),
		Db_name:           utils.StringPointer("10"),
		Db_user:           utils.StringPointer("cgrates"),
		Db_password:       utils.StringPointer("password"),
		Remote_conns:      &[]string{"*conn1"},
		Replication_conns: &[]string{"*conn1"},
		Items: map[string]*ItemOptJson{
			utils.MetaAccounts: {
				Replicate: utils.BoolPointer(true),
				Remote:    utils.BoolPointer(true),
			},
		},
		Opts: map[string]interface{}{
			utils.RedisSentinelNameCfg: "sentinel",
		},
	}
	jsnCfg := NewDefaultCGRConfig()
	if err = jsnCfg.dataDbCfg.loadFromJSONCfg(jsonCfg); err != nil {
		t.Error(err)
	} else {
		rcv := jsnCfg.dataDbCfg.Clone()
		if !reflect.DeepEqual(rcv.Items[utils.MetaAccounts], jsnCfg.dataDbCfg.Items[utils.MetaAccounts]) {
			t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(rcv.Items[utils.MetaAccounts]),
				utils.ToJSON(jsnCfg.dataDbCfg.Items[utils.MetaAccounts]))
		} else if !reflect.DeepEqual(rcv.Opts[utils.RedisSentinelNameCfg], jsnCfg.dataDbCfg.Opts[utils.RedisSentinelNameCfg]) {
			t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(rcv.Opts[utils.RedisSentinelNameCfg]),
				utils.ToJSON(jsnCfg.dataDbCfg.Opts[utils.RedisSentinelNameCfg]))
		} else if !reflect.DeepEqual(rcv.RplConns, jsnCfg.dataDbCfg.RplConns) {
			t.Errorf("Expected %+v \n, received %+v", rcv.RplConns, jsnCfg.dataDbCfg.RplConns)
		}
	}
}

func TestDataDbEqualsTrue(t *testing.T) {
	itm := &ItemOpt{
		Remote:    true,
		Replicate: false,
		RouteID:   "RouteID",
		APIKey:    "APIKey",
	}

	itm2 := &ItemOpt{
		Remote:    true,
		Replicate: false,
		RouteID:   "RouteID",
		APIKey:    "APIKey",
	}

	if !itm.Equals(itm2) {
		t.Error("Items should match")
	}
}

func TestDataDbEqualsFalse(t *testing.T) {
	itm := &ItemOpt{
		Remote:    true,
		Replicate: false,
		RouteID:   "RouteID",
		APIKey:    "APIKey",
	}

	itm2 := &ItemOpt{
		Remote:    false,
		Replicate: true,
		RouteID:   "RouteID2",
		APIKey:    "APIKey2",
	}

	if itm.Equals(itm2) {
		t.Error("Items should not match")
	}
}

func TestDiffItemOptJson(t *testing.T) {
	var d *ItemOptJson

	v1 := &ItemOpt{
		Remote:    true,
		Replicate: false,
		RouteID:   "RouteID",
		APIKey:    "APIKey",
	}

	v2 := &ItemOpt{
		Remote:    false,
		Replicate: true,
		RouteID:   "RouteID2",
		APIKey:    "APIKey2",
	}

	expected := &ItemOptJson{
		Remote:    utils.BoolPointer(false),
		Replicate: utils.BoolPointer(true),
		Route_id:  utils.StringPointer("RouteID2"),
		Api_key:   utils.StringPointer("APIKey2"),
	}

	rcv := diffItemOptJson(d, v1, v2)
	if !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}

	v2_2 := v1
	expected2 := &ItemOptJson{}
	rcv = diffItemOptJson(d, v1, v2_2)
	if !reflect.DeepEqual(rcv, expected2) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected2), utils.ToJSON(rcv))
	}
}

func TestDiffMapItemOptJson(t *testing.T) {
	var d map[string]*ItemOptJson

	v1 := map[string]*ItemOpt{
		"ITEM_OPT1": {
			Remote:    true,
			Replicate: false,
			RouteID:   "RouteID",
			APIKey:    "APIKey",
		},
	}

	v2 := map[string]*ItemOpt{
		"ITEM_OPT1": {
			Remote:    false,
			Replicate: true,
			RouteID:   "RouteID2",
			APIKey:    "APIKey2",
		},
	}

	expected := map[string]*ItemOptJson{
		"ITEM_OPT1": {
			Remote:    utils.BoolPointer(false),
			Replicate: utils.BoolPointer(true),
			Route_id:  utils.StringPointer("RouteID2"),
			Api_key:   utils.StringPointer("APIKey2"),
		},
	}

	rcv := diffMapItemOptJson(d, v1, v2)
	if !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}

	v2_2 := v1
	expected2 := map[string]*ItemOptJson{}
	rcv = diffMapItemOptJson(d, v1, v2_2)
	if !reflect.DeepEqual(rcv, expected2) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected2), utils.ToJSON(rcv))
	}
}

func TestDiffDataDbJsonCfg(t *testing.T) {
	var d *DbJsonCfg

	v1 := &DataDbCfg{
		Type:        "mysql",
		Host:        "/host",
		Port:        "8080",
		Name:        "cgrates.org",
		User:        "cgrates",
		Password:    "CGRateSPassword",
		RmtConns:    []string{"itsyscom.com"},
		RmtConnID:   "connID",
		RplConns:    []string{},
		RplFiltered: true,
		RplCache:    "RplCache",
		Items:       map[string]*ItemOpt{},
		Opts:        map[string]interface{}{},
	}

	v2 := &DataDbCfg{
		Type:        "postgres",
		Host:        "/host2",
		Port:        "8037",
		Name:        "itsyscom.com",
		User:        "itsyscom",
		Password:    "ITsysCOMPassword",
		RmtConns:    []string{"cgrates.org"},
		RmtConnID:   "connID2",
		RplConns:    []string{"RplConn1"},
		RplFiltered: false,
		RplCache:    "RplCache2",
		Items: map[string]*ItemOpt{
			"ITEM_1": {
				Remote:    true,
				Replicate: true,
				RouteID:   "RouteID2",
				APIKey:    "APIKey2",
			},
		},
		Opts: map[string]interface{}{
			"OPT_1": "OptValue",
		},
	}

	expected := &DbJsonCfg{
		Db_type:              utils.StringPointer("postgres"),
		Db_host:              utils.StringPointer("/host2"),
		Db_port:              utils.IntPointer(8037),
		Db_name:              utils.StringPointer("itsyscom.com"),
		Db_user:              utils.StringPointer("itsyscom"),
		Db_password:          utils.StringPointer("ITsysCOMPassword"),
		Remote_conns:         &[]string{"cgrates.org"},
		Remote_conn_id:       utils.StringPointer("connID2"),
		Replication_conns:    &[]string{"RplConn1"},
		Replication_filtered: utils.BoolPointer(false),
		Replication_cache:    utils.StringPointer("RplCache2"),
		Items: map[string]*ItemOptJson{
			"ITEM_1": {
				Remote:    utils.BoolPointer(true),
				Replicate: utils.BoolPointer(true),
				Route_id:  utils.StringPointer("RouteID2"),
				Api_key:   utils.StringPointer("APIKey2"),
			},
		},
		Opts: map[string]interface{}{
			"OPT_1": "OptValue",
		},
	}

	rcv := diffDataDbJsonCfg(d, v1, v2)
	if !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}

	v2_2 := v1
	expected2 := &DbJsonCfg{
		Items: map[string]*ItemOptJson{},
		Opts:  map[string]interface{}{},
	}
	rcv = diffDataDbJsonCfg(d, v1, v2_2)
	if !reflect.DeepEqual(rcv, expected2) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected2), utils.ToJSON(rcv))
	}
}
