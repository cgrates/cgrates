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
		Items: &map[string]*ItemOptJson{
			utils.MetaAccounts: {
				Replicate: utils.BoolPointer(true),
				Remote:    utils.BoolPointer(true),
			},
			utils.MetaReverseDestinations: {
				Replicate: utils.BoolPointer(true),
			},
			utils.MetaDestinations: {
				Replicate: utils.BoolPointer(false),
			},
		},
		Opts: &DBOptsJson{
			RedisSentinel: utils.StringPointer("sentinel"),
		},
	}
	expected := &DataDbCfg{
		Type:     "*redis",
		Host:     "127.0.0.1",
		Port:     "6379",
		Name:     "10",
		User:     "cgrates",
		Password: "password",
		RmtConns: []string{"*conn1"},
		RplConns: []string{"*conn1"},
		Items: map[string]*ItemOpt{
			utils.MetaAccounts: {
				Limit:     -1,
				Replicate: true,
				Remote:    true,
			},
			utils.MetaReverseDestinations: {
				Limit:     -1,
				Replicate: true,
			},
			utils.MetaDestinations: {
				Limit:     -1,
				Replicate: false,
			},
		},
		Opts: &DataDBOpts{
			RedisSentinel: "sentinel",
		},
	}
	jsnCfg := NewDefaultCGRConfig()
	if err := jsnCfg.dataDbCfg.loadFromJSONCfg(nil); err != nil {
		t.Error(err)
	} else if err := jsnCfg.dataDbCfg.loadFromJSONCfg(jsonCfg); err != nil {
		t.Error(err)
	} else {
		if !reflect.DeepEqual(expected.Items[utils.MetaAccounts], jsnCfg.dataDbCfg.Items[utils.MetaAccounts]) {
			t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected.Items[utils.MetaAccounts]),
				utils.ToJSON(jsnCfg.dataDbCfg.Items[utils.MetaAccounts]))
		} else if !reflect.DeepEqual(expected.Opts.RedisSentinel, jsnCfg.dataDbCfg.Opts.RedisSentinel) {
			t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected.Opts.RedisSentinel),
				utils.ToJSON(jsnCfg.dataDbCfg.Opts.RedisSentinel))
		} else if !reflect.DeepEqual(expected.RplConns, jsnCfg.dataDbCfg.RplConns) {
			t.Errorf("Expected %+v \n, received %+v", expected.RplConns, jsnCfg.dataDbCfg.RplConns)
		}
	}

	if err := jsnCfg.dataDbCfg.Opts.loadFromJSONCfg(nil); err != nil {
		t.Error(err)
	} else if err := jsnCfg.dataDbCfg.loadFromJSONCfg(&DbJsonCfg{
		Opts: &DBOptsJson{
			RedisClusterSync: utils.StringPointer("test"),
		}}); err == nil {
		t.Error(err)
	} else if err := jsnCfg.dataDbCfg.loadFromJSONCfg(&DbJsonCfg{
		Opts: &DBOptsJson{
			RedisClusterOndownDelay: utils.StringPointer("test2"),
		},
	}); err == nil {
		t.Error(err)
	} else if err := jsnCfg.dataDbCfg.loadFromJSONCfg(&DbJsonCfg{
		Opts: &DBOptsJson{
			RedisConnectTimeout: utils.StringPointer("test3"),
		},
	}); err == nil {
		t.Error(err)
	} else if err := jsnCfg.dataDbCfg.loadFromJSONCfg(&DbJsonCfg{
		Opts: &DBOptsJson{
			RedisReadTimeout: utils.StringPointer("test4"),
		},
	}); err == nil {
		t.Error(err)
	} else if err := jsnCfg.dataDbCfg.loadFromJSONCfg(&DbJsonCfg{
		Opts: &DBOptsJson{
			RedisWriteTimeout: utils.StringPointer("test5"),
		},
	}); err == nil {
		t.Error(err)
	} else if err := jsnCfg.dataDbCfg.loadFromJSONCfg(&DbJsonCfg{
		Opts: &DBOptsJson{
			MongoQueryTimeout: utils.StringPointer("test4"),
		},
	}); err == nil {
		t.Error(err)
	} else if err := jsnCfg.dataDbCfg.loadFromJSONCfg(&DbJsonCfg{
		Items: &map[string]*ItemOptJson{
			utils.MetaAccounts: {
				Ttl: utils.StringPointer("test5"),
			}}}); err == nil {
		t.Error(err)
	}

}

func TestConnsloadFromJsonCfg(t *testing.T) {
	jsonCfg := &DbJsonCfg{
		Remote_conns: &[]string{"*internal"},
	}
	expectedErrRmt := "Remote connection ID needs to be different than *internal"
	jsnCfg := NewDefaultCGRConfig()
	if err := jsnCfg.dataDbCfg.loadFromJSONCfg(jsonCfg); err == nil || err.Error() != expectedErrRmt {
		t.Errorf("Expected %+v, received %+v", expectedErrRmt, err)
	}

	jsonCfg = &DbJsonCfg{
		Replication_conns: &[]string{"*internal"},
	}
	expectedErrRpl := "Replication connection ID needs to be different than *internal"
	jsnCfg = NewDefaultCGRConfig()
	if err := jsnCfg.dataDbCfg.loadFromJSONCfg(jsonCfg); err == nil || err.Error() != expectedErrRpl {
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
	dbcfg.Opts = &DataDBOpts{}
	cfgJSONStr := `{
"data_db": {
	"db_type": "mongo",
	}
}`
	expected := DataDbCfg{
		Type: "*mongo",
		Opts: &DataDBOpts{},
	}
	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnDataDbCfg, err := jsnCfg.DbJsonCfg(DATADB_JSN); err != nil {
		t.Error(err)
	} else if err = dbcfg.loadFromJSONCfg(jsnDataDbCfg); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, dbcfg) {
		t.Errorf("Expected: %+v , received: %+v", utils.ToJSON(expected), utils.ToJSON(dbcfg))
	}
	cfgJSONStr = `{
"data_db": {
	"db_type": "mongo",
	"db_port": -1,
	}
}`
	expected = DataDbCfg{
		Type: "*mongo",
		Port: "27017",
		Opts: &DataDBOpts{},
	}
	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnDataDbCfg, err := jsnCfg.DbJsonCfg(DATADB_JSN); err != nil {
		t.Error(err)
	} else if err = dbcfg.loadFromJSONCfg(jsnDataDbCfg); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, dbcfg) {
		t.Errorf("Expected: %+v , received: %+v", utils.ToJSON(expected), utils.ToJSON(dbcfg))
	}
	cfgJSONStr = `{
"data_db": {
	"db_type": "*internal",
	"db_port": -1,
	}
}`
	expected = DataDbCfg{
		Type: "*internal",
		Port: "internal",
		Opts: &DataDBOpts{},
	}
	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnDataDbCfg, err := jsnCfg.DbJsonCfg(DATADB_JSN); err != nil {
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

	dbcfg.Opts = &DataDBOpts{}
	expected = DataDbCfg{
		Type:     "*redis",
		Host:     "127.0.0.1",
		Port:     "6379",
		Name:     "10",
		User:     "cgrates",
		Password: "password",
		RmtConns: []string{"Conn1"},
		Opts: &DataDBOpts{
			RedisSentinel: "sentinel",
		},
	}
	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnDataDbCfg, err := jsnCfg.DbJsonCfg(DATADB_JSN); err != nil {
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
		Type:     utils.MetaInternal,
		Host:     "127.0.0.1",
		Port:     "6379",
		Name:     "10",
		User:     "cgrates",
		Password: "password",
		Opts: &DataDBOpts{
			RedisSentinel: "sentinel",
		},
		RmtConns: []string{"Conn1"},
		RplConns: []string{"Conn2"},
	}
	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnDataDbCfg, err := jsnCfg.DbJsonCfg(DATADB_JSN); err != nil {
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
		Type:     utils.MetaInternal,
		Host:     "127.0.0.1",
		Port:     "6379",
		Name:     "10",
		User:     "cgrates",
		Password: "password",
		Opts: &DataDBOpts{
			RedisSentinel: "sentinel",
		},
		RmtConns: []string{"Conn1", "Conn2", "Conn3"},
		RplConns: []string{"Conn4", "Conn5"},
	}
	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnDataDbCfg, err := jsnCfg.DbJsonCfg(DATADB_JSN); err != nil {
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
		"*reverse_destinations": {"replicate":false},
		"*destinations": {"replicate":false},
	  }	,
	"opts": {
		"redisSentinel":"sentinel",			// redisSentinel is the name of sentinel
	  }
	},
}`

	expected = DataDbCfg{
		Type:     "*redis",
		Host:     "127.0.0.1",
		Port:     "6379",
		Name:     "10",
		User:     "cgrates",
		Password: "password",
		RmtConns: []string{"Conn1"},
		Items: map[string]*ItemOpt{
			utils.MetaAccounts: {
				Replicate: true,
				Limit:     -1,
			},
			utils.MetaReverseDestinations: {
				Replicate: false,
				Limit:     -1,
			},
			utils.MetaDestinations: {
				Replicate: false,
				Limit:     -1,
			},
		},
		Opts: &DataDBOpts{
			RedisSentinel: "sentinel",
		},
	}
	dbcfg.Items = make(map[string]*ItemOpt)
	dbcfg.Opts = &DataDBOpts{}
	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnDataDbCfg, err := jsnCfg.DbJsonCfg(DATADB_JSN); err != nil {
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
				"*load_ids":{"remote":true, "replicate":true}, 
			
			  }	
			}
		}`

	expected = DataDbCfg{
		Type:     "*redis",
		Host:     "127.0.0.1",
		Port:     "6379",
		Name:     "10",
		User:     "cgrates",
		Password: "password",
		Opts: &DataDBOpts{
			RedisSentinel: "sentinel",
		},
		RmtConns: []string{"Conn1"},
		Items: map[string]*ItemOpt{
			utils.MetaDispatcherHosts: {
				Remote:    true,
				Replicate: true,
				Limit:     -1,
			},
			utils.MetaLoadIDs: {
				Remote:    true,
				Replicate: true,
				Limit:     -1,
			},
		},
	}
	dbcfg.Items = make(map[string]*ItemOpt)
	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnDataDbCfg, err := jsnCfg.DbJsonCfg(DATADB_JSN); err != nil {
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
				"*timings": {"remote":false, "replicate":false}, 
				"*resource_profiles":{"remote":false, "replicate":false}, 
				"*resources":{"remote":false, "replicate":false}, 
				"*statqueue_profiles": {"remote":false, "replicate":false}, 
			  }	
			}
		}`

	expected = DataDbCfg{
		Type:     "*redis",
		Host:     "127.0.0.1",
		Port:     "6379",
		Name:     "10",
		User:     "cgrates",
		Password: "password",
		Opts: &DataDBOpts{
			RedisSentinel: "sentinel",
		},
		RmtConns: []string{"Conn1"},
		Items: map[string]*ItemOpt{
			utils.MetaTimings:           {Limit: -1},
			utils.MetaResourceProfile:   {Limit: -1},
			utils.MetaResources:         {Limit: -1},
			utils.MetaStatQueueProfiles: {Limit: -1},
		},
	}
	dbcfg.Items = make(map[string]*ItemOpt)
	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnDataDbCfg, err := jsnCfg.DbJsonCfg(DATADB_JSN); err != nil {
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
			"*accounts":{"remote":true, "replicate":false, "api_key": "randomVal", "route_id": "randomVal","ttl":"1"}, 					
			"*reverse_destinations": {"remote":false, "replicate":false, "api_key": "randomVal", "route_id": "randomVal","ttl":"1"},
		},
	},		
}`
	eMap := map[string]any{
		utils.DataDbTypeCfg: "*redis",
		utils.DataDbHostCfg: "127.0.0.1",
		utils.DataDbPortCfg: 6379,
		utils.DataDbNameCfg: "10",
		utils.DataDbUserCfg: "cgrates",
		utils.DataDbPassCfg: "",
		utils.OptsCfg: map[string]any{
			utils.RedisSentinelNameCfg: "",
			utils.MongoQueryTimeoutCfg: "10s",
		},
		utils.RemoteConnsCfg:      []string{},
		utils.ReplicationConnsCfg: []string{},
		utils.ItemsCfg: map[string]any{
			utils.MetaAccounts:            map[string]any{utils.RemoteCfg: true, utils.ReplicateCfg: false, utils.APIKeyCfg: "randomVal", utils.RouteIDCfg: "randomVal", utils.LimitCfg: -1, utils.StaticTTLCfg: false, utils.TTLCfg: "1ns"},
			utils.MetaReverseDestinations: map[string]any{utils.RemoteCfg: false, utils.ReplicateCfg: false, utils.APIKeyCfg: "randomVal", utils.RouteIDCfg: "randomVal", utils.LimitCfg: -1, utils.StaticTTLCfg: false, utils.TTLCfg: "1ns"},
		},
	}
	if cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else {
		rcv := cgrCfg.dataDbCfg.AsMapInterface()
		if err != nil {
			t.Error(err)
		}
		if !reflect.DeepEqual(eMap[utils.ItemsCfg].(map[string]any)[utils.MetaAccounts],
			rcv[utils.ItemsCfg].(map[string]any)[utils.MetaAccounts]) {
			t.Errorf("Expected %+v, received %+v", eMap[utils.ItemsCfg].(map[string]any)[utils.MetaAccounts],
				rcv[utils.ItemsCfg].(map[string]any)[utils.MetaAccounts])
		} else if !reflect.DeepEqual(eMap[utils.ItemsCfg].(map[string]any)[utils.MetaReverseDestinations],
			rcv[utils.ItemsCfg].(map[string]any)[utils.MetaReverseDestinations]) {
			t.Errorf("Expected %+v, received %+v", eMap[utils.ItemsCfg].(map[string]any)[utils.MetaReverseDestinations],
				rcv[utils.ItemsCfg].(map[string]any)[utils.MetaReverseDestinations])
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
		Items: &map[string]*ItemOptJson{
			utils.MetaAccounts: {
				Replicate: utils.BoolPointer(true),
				Remote:    utils.BoolPointer(true),
			},
			utils.MetaReverseDestinations: {
				Replicate: utils.BoolPointer(true),
			},
			utils.MetaDestinations: {
				Replicate: utils.BoolPointer(false),
			},
		},
		Opts: &DBOptsJson{
			RedisSentinel: utils.StringPointer("sentinel"),
		},
	}
	jsnCfg := NewDefaultCGRConfig()
	if err := jsnCfg.dataDbCfg.loadFromJSONCfg(jsonCfg); err != nil {
		t.Error(err)
	} else {
		rcv := jsnCfg.dataDbCfg.Clone()
		if !reflect.DeepEqual(rcv.Items[utils.MetaAccounts], jsnCfg.dataDbCfg.Items[utils.MetaAccounts]) {
			t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(rcv.Items[utils.MetaAccounts]),
				utils.ToJSON(jsnCfg.dataDbCfg.Items[utils.MetaAccounts]))
		} else if !reflect.DeepEqual(rcv.Opts.RedisSentinel, jsnCfg.dataDbCfg.Opts.RedisSentinel) {
			t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(rcv.Opts.RedisSentinel),
				utils.ToJSON(jsnCfg.dataDbCfg.Opts.RedisSentinel))
		} else if !reflect.DeepEqual(rcv.RplConns, jsnCfg.dataDbCfg.RplConns) {
			t.Errorf("Expected %+v \n, received %+v", rcv.RplConns, jsnCfg.dataDbCfg.RplConns)
		}
	}
}
