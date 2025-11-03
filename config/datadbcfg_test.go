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
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/utils"
)

func TestDataDbCfgloadFromJsonCfg(t *testing.T) {
	jsonCfg := &DbJsonCfg{
		Db_conns: DbConnsJson{
			utils.MetaDefault: &DbConnJson{
				Db_type:           utils.StringPointer("redis"),
				Db_host:           utils.StringPointer("127.0.0.1"),
				Db_port:           utils.IntPointer(6379),
				Db_name:           utils.StringPointer("10"),
				Db_user:           utils.StringPointer("cgrates"),
				Db_password:       utils.StringPointer("password"),
				Remote_conns:      &[]string{"*conn1"},
				Replication_conns: &[]string{"*conn1"},
				Opts: &DBOptsJson{
					RedisSentinel: utils.StringPointer("sentinel"),
				},
			},
		},
		Items: map[string]*ItemOptsJson{
			utils.MetaAccounts: {
				Replicate: utils.BoolPointer(true),
				Remote:    utils.BoolPointer(true),
			},
		},
	}
	expected := &DbCfg{
		DBConns: DBConns{
			utils.MetaDefault: &DBConn{
				Type:     "redis",
				Host:     "127.0.0.1",
				Port:     "6379",
				Name:     "10",
				User:     "cgrates",
				Password: "password",
				RmtConns: []string{"*conn1"},
				RplConns: []string{"*conn1"},
				Opts: &DBOpts{
					RedisSentinel: "sentinel",
				},
			},
		},
		Items: map[string]*ItemOpts{
			utils.MetaAccounts: {
				DBConn:    utils.MetaDefault,
				Limit:     -1,
				Replicate: true,
				Remote:    true,
			},
		},
	}
	jsnCfg := NewDefaultCGRConfig()
	if err := jsnCfg.dbCfg.loadFromJSONCfg(nil); err != nil {
		t.Error(err)
	} else if err := jsnCfg.dbCfg.loadFromJSONCfg(jsonCfg); err != nil {
		t.Error(err)
	} else {
		if !reflect.DeepEqual(expected.Items[utils.MetaAccounts], jsnCfg.dbCfg.Items[utils.MetaAccounts]) {
			t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected.Items[utils.MetaAccounts]),
				utils.ToJSON(jsnCfg.dbCfg.Items[utils.MetaAccounts]))
		} else if !reflect.DeepEqual(expected.DBConns[utils.MetaDefault].Opts.RedisSentinel, jsnCfg.dbCfg.DBConns[utils.MetaDefault].Opts.RedisSentinel) {
			t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected.DBConns[utils.MetaDefault].Opts.RedisSentinel),
				utils.ToJSON(jsnCfg.dbCfg.DBConns[utils.MetaDefault].Opts.RedisSentinel))
		} else if !reflect.DeepEqual(expected.DBConns[utils.MetaDefault].RplConns, jsnCfg.dbCfg.DBConns[utils.MetaDefault].RplConns) {
			t.Errorf("Expected %+v \n, received %+v", expected.DBConns[utils.MetaDefault].RplConns, jsnCfg.dbCfg.DBConns[utils.MetaDefault].RplConns)
		}
	}
}

func TestDataDbCfgloadFromJsonCfgItemsErr(t *testing.T) {
	jsonCfg := &DbJsonCfg{
		Items: map[string]*ItemOptsJson{

			"Bad Item": {
				Ttl: utils.StringPointer("bad input"),
			},
		},
	}
	expErr := `time: invalid duration "bad input"`
	jsnCfg := NewDefaultCGRConfig()
	jsnCfg.dbCfg.Items = map[string]*ItemOpts{
		"Bad Item": {},
	}
	if err := jsnCfg.dbCfg.loadFromJSONCfg(jsonCfg); err.Error() != expErr {
		t.Errorf("Expected Error <%v>, ]\n Received error <%v>", expErr, err.Error())
	}
}

func TestDataDbLoadFromJsonCfgOpt(t *testing.T) {
	dbOpts := &DBOpts{}
	if err := dbOpts.loadFromJSONCfg(nil); err != nil {
		t.Error(err)
	}
	errExpect := `time: unknown unit "c" in duration "2c"`
	jsnCfg := &DBOptsJson{
		RedisClusterSync: utils.StringPointer("2c"),
	}

	if err := dbOpts.loadFromJSONCfg(jsnCfg); err == nil || err.Error() != errExpect {
		t.Errorf("Expecting %v \n but received \n %v", errExpect, err.Error())
	}

	jsnCfg = &DBOptsJson{
		RedisClusterOndownDelay: utils.StringPointer("2c"),
	}

	if err := dbOpts.loadFromJSONCfg(jsnCfg); err == nil || err.Error() != errExpect {
		t.Errorf("Expecting %v \n but received \n %v", errExpect, err.Error())
	}
	jsnCfg = &DBOptsJson{
		MongoQueryTimeout: utils.StringPointer("2c"),
	}

	if err := dbOpts.loadFromJSONCfg(jsnCfg); err == nil || err.Error() != errExpect {
		t.Errorf("Expecting %v \n but received \n %v", errExpect, err.Error())
	}

}

func TestDataDbLoadFromJsonCfgRedisConnTimeOut(t *testing.T) {
	dbOpts := &DBOpts{}
	jsnCfg := &DBOptsJson{
		RedisConnectTimeout: utils.StringPointer("2c"),
	}
	errExpect := `time: unknown unit "c" in duration "2c"`
	if err := dbOpts.loadFromJSONCfg(jsnCfg); err == nil || err.Error() != errExpect {
		t.Errorf("Expecting %v \n but received \n %v", errExpect, err.Error())
	}
}
func TestDataDbLoadFromJsonCfgRedisReadTimeOut(t *testing.T) {
	dbOpts := &DBOpts{}
	jsnCfg := &DBOptsJson{
		RedisReadTimeout: utils.StringPointer("2c"),
	}
	errExpect := `time: unknown unit "c" in duration "2c"`
	if err := dbOpts.loadFromJSONCfg(jsnCfg); err == nil || err.Error() != errExpect {
		t.Errorf("Expecting %v \n but received \n %v", errExpect, err.Error())
	}
}

func TestDataDbLoadFromJsonCfgRedisWriteTimeout(t *testing.T) {
	dbOpts := &DBOpts{}
	jsnCfg := &DBOptsJson{
		RedisWriteTimeout: utils.StringPointer("2c"),
	}
	errExpect := `time: unknown unit "c" in duration "2c"`
	if err := dbOpts.loadFromJSONCfg(jsnCfg); err == nil || err.Error() != errExpect {
		t.Errorf("Expecting %v \n but received \n %v", errExpect, err.Error())
	}
}
func TestConnsloadFromJsonCfg(t *testing.T) {
	jsonCfg := &DbJsonCfg{
		Db_conns: DbConnsJson{
			utils.MetaDefault: &DbConnJson{
				Remote_conns: &[]string{"*internal"},
			},
		},
	}
	expectedErrRmt := "remote connection ID needs to be different than <*internal> "
	jsnCfg := NewDefaultCGRConfig()
	if err := jsnCfg.dbCfg.loadFromJSONCfg(jsonCfg); err == nil || err.Error() != expectedErrRmt {
		t.Errorf("Expected %+v, received %+v", expectedErrRmt, err)
	}

	jsonCfg = &DbJsonCfg{
		Db_conns: DbConnsJson{
			utils.MetaDefault: &DbConnJson{
				Replication_conns: &[]string{"*internal"},
			},
		},
	}
	expectedErrRpl := "remote connection ID needs to be different than <*internal> "
	jsnCfg = NewDefaultCGRConfig()
	if err := jsnCfg.dbCfg.loadFromJSONCfg(jsonCfg); err == nil || err.Error() != expectedErrRpl {
		t.Errorf("Expected %+v, received %+v", expectedErrRpl, err)
	}
}

func TestItemCfgloadFromJson(t *testing.T) {
	jsonCfg := &ItemOptsJson{
		Remote:    utils.BoolPointer(true),
		Replicate: utils.BoolPointer(true),
		Api_key:   utils.StringPointer("randomVal"),
		Route_id:  utils.StringPointer("randomID"),
	}
	expected := &ItemOpts{
		Remote:    true,
		Replicate: true,
		APIKey:    "randomVal",
		RouteID:   "randomID",
	}
	rcv := new(ItemOpts)
	rcv.loadFromJSONCfg(nil)
	rcv.loadFromJSONCfg(jsonCfg)
	if !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}

func TestDataDbCfgloadFromJsonCfgPort(t *testing.T) {
	var dbcfg DbCfg
	dbcfg.DBConns = DBConns{}
	cfgJSONStr := `{
"db": {
	"db_conns": {
		"*default": {	
			"db_type": "mongo",
			},
		},
	}
}`
	cfg := NewDefaultCGRConfig()
	expected := DbCfg{
		DBConns: DBConns{
			utils.MetaDefault: &DBConn{
				Type: utils.MetaMongo,
			},
		},
	}
	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if err := dbcfg.Load(context.Background(), jsnCfg, cfg); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected.DBConns[utils.MetaDefault].Type, dbcfg.DBConns[utils.MetaDefault].Type) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(dbcfg))
	}
	cfgJSONStr = `{
"db": {
	"db_conns": {
		"*default": {	
			"db_type": "mongo",
			"db_port": -1,
			},
		},
	}
}`
	expected = DbCfg{
		DBConns: DBConns{
			utils.MetaDefault: &DBConn{
				Type: utils.MetaMongo,
				Port: "27017",
			},
		},
	}
	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if err := dbcfg.Load(context.Background(), jsnCfg, cfg); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected.DBConns[utils.MetaDefault].Port, dbcfg.DBConns[utils.MetaDefault].Port) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(dbcfg))
	}
	cfgJSONStr = `{
"db": {
	"db_conns": {
		"*default": {	
			"db_type": "*internal",
			"db_port": -1,
				"opts":{
				"internalDBRewriteInterval": "0s",
				"internalDBDumpInterval": "0s"
					}
			},
		},
	},
}`
	expected = DbCfg{
		DBConns: DBConns{
			utils.MetaDefault: &DBConn{
				Type: utils.MetaInternal,
				Port: "internal",
				Opts: &DBOpts{
					InternalDBDumpInterval:    0,
					InternalDBRewriteInterval: 0,
				},
			},
		},
	}
	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if err := dbcfg.Load(context.Background(), jsnCfg, cfg); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected.DBConns[utils.MetaDefault].Opts.InternalDBDumpInterval, dbcfg.DBConns[utils.MetaDefault].Opts.InternalDBDumpInterval) {
		t.Errorf("Expected: %+v , received: %+v", expected, dbcfg)
	}
}

func TestDataDBRemoteReplication(t *testing.T) {
	var dbcfg, expected DbCfg
	cfgJSONStr := `{
"db": {								// database used to store runtime data (eg: accounts, cdr stats)
	"db_conns": {
		"*default": {	
			"db_type": "*redis",					// data_db type: <*redis|*mongo|*internal>
			"db_host": "127.0.0.1",					// data_db host address
			"db_port": -1,	 						// data_db port to reach the database
			"db_name": "10", 						// data_db database name to connect to
			"db_user": "cgrates", 					// username to use when connecting to data_db
			"db_password": "password",				// password to use when connecting to data_db
			"remote_conns":["Conn1"],
			"opts":{
			"redisSentinel":"sentinel",			// redisSentinel is the name of sentinel
	      	},
			},
		},
	},
}`
	cfg := NewDefaultCGRConfig()
	dbcfg.DBConns = DBConns{}
	expected = DbCfg{
		DBConns: DBConns{
			utils.MetaDefault: &DBConn{
				Type:     "*redis",
				Host:     "127.0.0.1",
				Port:     "6379",
				Name:     "10",
				User:     "cgrates",
				Password: "password",
				RmtConns: []string{"Conn1"},
				Opts: &DBOpts{
					RedisSentinel: "sentinel",
				},
			},
		},
	}
	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if err := dbcfg.Load(context.Background(), jsnCfg, cfg); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, dbcfg) {
		t.Errorf("Expected: %+v ,\n received: %+v", utils.ToJSON(expected), utils.ToJSON(dbcfg))
	}
	cfgJSONStr = `{
"db": {								// database used to store runtime data (eg: accounts, cdr stats)
	"db_conns": {
		"*default": {	
			"db_type": "*internal",					// data_db type: <*redis|*mongo|*internal>
			"remote_conns":["Conn1"],
			"replication_conns":["Conn2"],
			"opts":{
	        	"internalDBRewriteInterval": "0s",
	         	"internalDBDumpInterval": "0s"
           	}
			},
		},
	},
	
}`

	expected = DbCfg{
		DBConns: DBConns{
			utils.MetaDefault: &DBConn{
				Type:     utils.MetaInternal,
				RmtConns: []string{"Conn1"},
				RplConns: []string{"Conn2"},
				Opts: &DBOpts{
					InternalDBDumpInterval:    0,
					InternalDBRewriteInterval: 0,
				},
			},
		},
	}
	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if err := dbcfg.Load(context.Background(), jsnCfg, cfg); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected.DBConns[utils.MetaDefault].Opts.InternalDBRewriteInterval, dbcfg.DBConns[utils.MetaDefault].Opts.InternalDBRewriteInterval) {
		t.Errorf("Expected: %+v ,\n received: %+v", utils.ToJSON(expected), utils.ToJSON(dbcfg))
	} else if !reflect.DeepEqual(expected.DBConns[utils.MetaDefault].Opts.InternalDBDumpInterval, dbcfg.DBConns[utils.MetaDefault].Opts.InternalDBDumpInterval) {
		t.Errorf("Expected: %+v ,\n received: %+v", utils.ToJSON(expected), utils.ToJSON(dbcfg))
	}

	cfgJSONStr = `{
"db": {								// database used to store runtime data (eg: accounts, cdr stats)
	"db_conns": {
		"*default": {	
			"db_type": "*internal",					// data_db type: <*redis|*mongo|*internal>
			"remote_conns":["Conn1","Conn2","Conn3"],
			"replication_conns":["Conn4","Conn5"],
				"opts":{
					"internalDBRewriteInterval": "0s",
					"internalDBDumpInterval": "0s"
					}
			},
		},
	},
}`

	expected = DbCfg{
		DBConns: DBConns{
			utils.MetaDefault: &DBConn{
				Type:     utils.MetaInternal,
				RmtConns: []string{"Conn1", "Conn2", "Conn3"},
				RplConns: []string{"Conn4", "Conn5"},
				Opts: &DBOpts{
					InternalDBDumpInterval:    0,
					InternalDBRewriteInterval: 0,
				},
			},
		},
	}
	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if err := dbcfg.Load(context.Background(), jsnCfg, cfg); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected.DBConns[utils.MetaDefault].Opts.InternalDBDumpInterval, dbcfg.DBConns[utils.MetaDefault].Opts.InternalDBRewriteInterval) {
		t.Errorf("Expected: %+v ,\n received: %+v", utils.ToJSON(expected), utils.ToJSON(dbcfg))
	}
}

func TestDataDbCfgloadFromJsonCfgItems(t *testing.T) {
	var dbcfg, expected DbCfg
	cfgJSONStr := `{
"db": {								// database used to store runtime data (eg: accounts, cdr stats)
	"db_conns": {
		"*default": {	
			"db_type": "*redis",					// data_db type: <*redis|*mongo|*internal>
			"db_host": "127.0.0.1",					// data_db host address
			"db_port": -1,	 						// data_db port to reach the database
			"db_name": "10", 						// data_db database name to connect to
			"db_user": "cgrates", 					// username to use when connecting to data_db
			"db_password": "password",				// password to use when connecting to data_db
			"remote_conns":["Conn1"],
			"opts": {
				"redisSentinel":"sentinel",			// redisSentinel is the name of sentinel
	  		}
		},
	},
    "items":{
		"*accounts":{"replicate":true},
	  }	,
	},
}`

	expected = DbCfg{
		DBConns: DBConns{
			utils.MetaDefault: &DBConn{
				Type:     "*redis",
				Host:     "127.0.0.1",
				Port:     "6379",
				Name:     "10",
				User:     "cgrates",
				Password: "password",
				RmtConns: []string{"Conn1"},
				Opts: &DBOpts{
					RedisSentinel: "sentinel",
				},
			},
		},
		Items: map[string]*ItemOpts{
			utils.MetaAccounts: {
				Limit:     -1,
				Replicate: true,
			},
		},
	}
	dbcfg.Items = make(map[string]*ItemOpts)
	dbcfg.DBConns = DBConns{}
	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if err := dbcfg.Load(context.Background(), jsnCfg, NewDefaultCGRConfig()); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, dbcfg) {
		t.Errorf("Expected: %+v ,\n received: %+v", utils.ToJSON(expected), utils.ToJSON(dbcfg))
	}

	cfgJSONStr = `{
		"db": {								// database used to store runtime data (eg: accounts, cdr stats)
			"db_conns": {
				"*default": {	
					"db_type": "*redis",					// data_db type: <*redis|*mongo|*internal>
					"db_host": "127.0.0.1",					// data_db host address
					"db_port": -1,	 						// data_db port to reach the database
					"db_name": "10", 						// data_db database name to connect to
					"db_user": "cgrates", 					// username to use when connecting to data_db
					"db_password": "password",				// password to use when connecting to data_db
					"remote_conns":["Conn1"],
					"opts": {
				    "redisSentinel":"sentinel",			// redisSentinel is the name of sentinel
		            },
				},
			},
			"items":{ 
				"*load_ids":{"remote":true, "replicate":true}, 
			
			  }	
			}
		}`

	expected = DbCfg{
		DBConns: DBConns{
			utils.MetaDefault: &DBConn{
				Type:     "*redis",
				Host:     "127.0.0.1",
				Port:     "6379",
				Name:     "10",
				User:     "cgrates",
				Password: "password",
				RmtConns: []string{"Conn1"},
				Opts: &DBOpts{
					RedisSentinel: "sentinel",
				},
			},
		},
		Items: map[string]*ItemOpts{
			utils.MetaLoadIDs: {
				Limit:     -1,
				Remote:    true,
				Replicate: true,
			},
		},
	}
	dbcfg.Items = make(map[string]*ItemOpts)
	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if err := dbcfg.Load(context.Background(), jsnCfg, NewDefaultCGRConfig()); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, dbcfg) {
		t.Errorf("Expected: %+v ,\n received: %+v", utils.ToJSON(expected), utils.ToJSON(dbcfg))
	}

	cfgJSONStr = `{
		"db": {								// database used to store runtime data (eg: accounts, cdr stats)
			"db_conns": {
				"*default": {	
					"db_type": "*redis",					// data_db type: <*redis|*mongo|*internal>
					"db_host": "127.0.0.1",					// data_db host address
					"db_port": -1,	 						// data_db port to reach the database
					"db_name": "10", 						// data_db database name to connect to
					"db_user": "cgrates", 					// username to use when connecting to data_db
					"db_password": "password",				// password to use when connecting to data_db
					"remote_conns":["Conn1"],
					"opts": {
						"redisSentinel":"sentinel",			// redisSentinel is the name of sentinel
					},
				},
			},
			"items":{
				"*resource_profiles":{"remote":false, "replicate":false}, 
				"*resources":{"remote":false, "replicate":false}, 
				"*statqueue_profiles": {"remote":false, "replicate":false}, 
			  }	
			}
		}`

	expected = DbCfg{
		DBConns: DBConns{
			utils.MetaDefault: &DBConn{
				Type:     "*redis",
				Host:     "127.0.0.1",
				Port:     "6379",
				Name:     "10",
				User:     "cgrates",
				Password: "password",
				RmtConns: []string{"Conn1"},
				Opts: &DBOpts{
					RedisSentinel: "sentinel",
				},
			},
		},
		Items: map[string]*ItemOpts{
			utils.MetaResourceProfiles:  {Limit: -1},
			utils.MetaResources:         {Limit: -1},
			utils.MetaStatQueueProfiles: {Limit: -1},
		},
	}
	dbcfg.Items = make(map[string]*ItemOpts)
	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if err := dbcfg.Load(context.Background(), jsnCfg, NewDefaultCGRConfig()); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, dbcfg) {
		t.Errorf("Expected: %+v ,\n received: %+v", utils.ToJSON(expected), utils.ToJSON(dbcfg))
	}
}

func TestDataDbCfgAsMapInterface(t *testing.T) {
	cfgJSONStr := `{
	"db": {			
		"db_conns": {
			"*default": {	
				"db_type": "*redis",					
				"db_host": "127.0.0.1",					
				"db_port": 6379, 						
				"db_name": "10", 						
				"db_user": "cgrates", 					
				"db_password": "", 						
				"remote_conns":[],
				"replication_conns":[],
				"opts": {
					"redisSentinel":"",					
					"mongoQueryTimeout":"10s",
				},
			},
		},					
		"items":{
			"*accounts":{"remote":true, "replicate":false, "api_key": "randomVal", "route_id": "randomVal"}, 					
			"*reverse_destinations": {"remote":false, "replicate":false, "api_key": "randomVal", "route_id": "randomVal"},
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
			utils.MetaAccounts: map[string]any{utils.RemoteCfg: true, utils.ReplicateCfg: false, utils.APIKeyCfg: "randomVal", utils.RouteIDCfg: "randomVal", utils.LimitCfg: -1, utils.StaticTTLCfg: false, utils.DBConnCfg: utils.MetaDefault},
		},
	}
	if cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else {
		rcv := cgrCfg.dbCfg.AsMapInterface().(map[string]any)
		if !reflect.DeepEqual(eMap[utils.ItemsCfg].(map[string]any)[utils.MetaAccounts],
			rcv[utils.ItemsCfg].(map[string]any)[utils.MetaAccounts]) {
			t.Errorf("Expected %+v, received %+v", eMap[utils.ItemsCfg].(map[string]any)[utils.MetaAccounts],
				rcv[utils.ItemsCfg].(map[string]any)[utils.MetaAccounts])
		}
	}
}

func TestCloneDataDB(t *testing.T) {
	jsonCfg := &DbJsonCfg{
		Db_conns: DbConnsJson{
			utils.MetaDefault: &DbConnJson{
				Db_type:           utils.StringPointer("redis"),
				Db_host:           utils.StringPointer("127.0.0.1"),
				Db_port:           utils.IntPointer(6379),
				Db_name:           utils.StringPointer("10"),
				Db_user:           utils.StringPointer("cgrates"),
				Db_password:       utils.StringPointer("password"),
				Remote_conns:      &[]string{"*conn1"},
				Replication_conns: &[]string{"*conn1"},
				Opts: &DBOptsJson{
					RedisSentinel: utils.StringPointer("sentinel"),
				},
			},
		},
		Items: map[string]*ItemOptsJson{
			utils.MetaAccounts: {
				Replicate: utils.BoolPointer(true),
				Remote:    utils.BoolPointer(true),
			},
		},
	}
	jsnCfg := NewDefaultCGRConfig()
	if err := jsnCfg.dbCfg.loadFromJSONCfg(jsonCfg); err != nil {
		t.Error(err)
	} else {
		rcv := jsnCfg.dbCfg.Clone()
		if !reflect.DeepEqual(rcv.Items[utils.MetaAccounts], jsnCfg.dbCfg.Items[utils.MetaAccounts]) {
			t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(rcv.Items[utils.MetaAccounts]),
				utils.ToJSON(jsnCfg.dbCfg.Items[utils.MetaAccounts]))
		} else if !reflect.DeepEqual(rcv.DBConns[utils.MetaDefault].Opts.RedisSentinel, jsnCfg.dbCfg.DBConns[utils.MetaDefault].Opts.RedisSentinel) {
			t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(rcv.DBConns[utils.MetaDefault].Opts.RedisSentinel),
				utils.ToJSON(jsnCfg.dbCfg.DBConns[utils.MetaDefault].Opts.RedisSentinel))
		} else if !reflect.DeepEqual(rcv.DBConns[utils.MetaDefault].RplConns, jsnCfg.dbCfg.DBConns[utils.MetaDefault].RplConns) {
			t.Errorf("Expected %+v \n, received %+v", rcv.DBConns[utils.MetaDefault].RplConns, jsnCfg.dbCfg.DBConns[utils.MetaDefault].RplConns)
		}
	}
}

func TestDataDbEqualsTrue(t *testing.T) {
	itm := &ItemOpts{
		Remote:    true,
		Replicate: false,
		RouteID:   "RouteID",
		APIKey:    "APIKey",
	}

	itm2 := &ItemOpts{
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
	itm := &ItemOpts{
		Remote:    true,
		Replicate: false,
		RouteID:   "RouteID",
		APIKey:    "APIKey",
	}

	itm2 := &ItemOpts{
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
	var d *ItemOptsJson

	v1 := &ItemOpts{
		Remote:    true,
		Replicate: false,
		RouteID:   "RouteID",
		APIKey:    "APIKey",
		Limit:     1,
		StaticTTL: true,
		TTL:       2,
	}

	v2 := &ItemOpts{
		Remote:    false,
		Replicate: true,
		RouteID:   "RouteID2",
		APIKey:    "APIKey2",
		Limit:     2,
		StaticTTL: false,
		TTL:       3,
	}

	expected := &ItemOptsJson{
		Remote:     utils.BoolPointer(false),
		Replicate:  utils.BoolPointer(true),
		Route_id:   utils.StringPointer("RouteID2"),
		Api_key:    utils.StringPointer("APIKey2"),
		Limit:      utils.IntPointer(2),
		Ttl:        utils.StringPointer("3ns"),
		Static_ttl: utils.BoolPointer(false),
	}

	rcv := diffItemOptJson(d, v1, v2)
	if !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}

	v2_2 := v1
	expected2 := &ItemOptsJson{}
	rcv = diffItemOptJson(d, v1, v2_2)
	if !reflect.DeepEqual(rcv, expected2) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected2), utils.ToJSON(rcv))
	}
}

func TestDiffMapItemOptJson(t *testing.T) {
	var d map[string]*ItemOptsJson

	v1 := map[string]*ItemOpts{
		"ITEM_OPT1": {
			Remote:    true,
			Replicate: false,
			RouteID:   "RouteID",
			APIKey:    "APIKey",
		},
	}

	v2 := map[string]*ItemOpts{
		"ITEM_OPT1": {
			Remote:    false,
			Replicate: true,
			RouteID:   "RouteID2",
			APIKey:    "APIKey2",
		},
	}

	expected := map[string]*ItemOptsJson{
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
	expected2 := map[string]*ItemOptsJson{}
	rcv = diffMapItemOptJson(d, v1, v2_2)
	if !reflect.DeepEqual(rcv, expected2) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected2), utils.ToJSON(rcv))
	}
}

func TestDiffDataDbJsonCfg(t *testing.T) {
	var d *DbJsonCfg

	v1 := &DbCfg{
		DBConns: DBConns{
			utils.MetaDefault: &DBConn{
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
				Opts: &DBOpts{
					RedisSentinel: "sentinel1",
				},
			},
		},
		Items: map[string]*ItemOpts{},
	}

	v2 := &DbCfg{
		DBConns: DBConns{
			utils.MetaDefault: &DBConn{
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
				Opts: &DBOpts{
					RedisSentinel: "sentinel2",
				},
			},
		},
		Items: map[string]*ItemOpts{
			"ITEM_1": {
				Remote:    true,
				Replicate: true,
				RouteID:   "RouteID2",
				APIKey:    "APIKey2",
			},
		},
	}

	expected := &DbJsonCfg{
		Db_conns: DbConnsJson{
			utils.MetaDefault: &DbConnJson{
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
				Opts: &DBOptsJson{
					RedisSentinel: utils.StringPointer("sentinel2"),
				},
			},
		},
		Items: map[string]*ItemOptsJson{
			"ITEM_1": {
				Remote:    utils.BoolPointer(true),
				Replicate: utils.BoolPointer(true),
				Route_id:  utils.StringPointer("RouteID2"),
				Api_key:   utils.StringPointer("APIKey2"),
			},
		},
	}

	rcv := diffDataDBJsonCfg(d, v1, v2)
	if !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}

	v2_2 := v1
	expected2 := &DbJsonCfg{
		Db_conns: DbConnsJson{},
		Items:    map[string]*ItemOptsJson{},
	}
	rcv = diffDataDBJsonCfg(d, v1, v2_2)
	if !reflect.DeepEqual(rcv, expected2) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected2), utils.ToJSON(rcv))
	}
}

func TestDataDbDiffOptsJson(t *testing.T) {
	var d *DBOptsJson

	v1 := &DBOpts{
		RedisSentinel:           "sentinel",
		RedisCluster:            false,
		RedisClusterSync:        1 * time.Second,
		RedisClusterOndownDelay: 1 * time.Second,
		MongoQueryTimeout:       1 * time.Second,
		RedisTLS:                false,
		RedisClientCertificate:  "",
		RedisClientKey:          "",
		RedisCACertificate:      "",
		RedisMaxConns:           2,
		RedisConnectAttempts:    3,
		RedisConnectTimeout:     3,
		RedisReadTimeout:        2,
		RedisWriteTimeout:       2,
	}

	v2 := &DBOpts{
		RedisSentinel:           "sentinel2",
		RedisCluster:            true,
		RedisClusterSync:        2 * time.Second,
		RedisClusterOndownDelay: 2 * time.Second,
		MongoQueryTimeout:       2 * time.Second,
		RedisTLS:                true,
		RedisClientCertificate:  "1",
		RedisClientKey:          "1",
		RedisCACertificate:      "1",
		RedisMaxConns:           3,
		RedisConnectAttempts:    4,
		RedisConnectTimeout:     4,
		RedisReadTimeout:        3,
		RedisWriteTimeout:       3,
	}

	exp := &DBOptsJson{
		RedisSentinel:           utils.StringPointer("sentinel2"),
		RedisCluster:            utils.BoolPointer(true),
		RedisClusterSync:        utils.StringPointer("2s"),
		RedisClusterOndownDelay: utils.StringPointer("2s"),
		MongoQueryTimeout:       utils.StringPointer("2s"),
		RedisTLS:                utils.BoolPointer(true),
		RedisClientCertificate:  utils.StringPointer("1"),
		RedisClientKey:          utils.StringPointer("1"),
		RedisCACertificate:      utils.StringPointer("1"),
		RedisWriteTimeout:       utils.StringPointer("3ns"),
		RedisReadTimeout:        utils.StringPointer("3ns"),
		RedisConnectTimeout:     utils.StringPointer("4ns"),
		RedisConnectAttempts:    utils.IntPointer(4),
		RedisMaxConns:           utils.IntPointer(3),
	}

	rcv := diffDataDBOptsJsonCfg(d, v1, v2)
	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(rcv))
	}
}
func TestDataDbDefaultDBPort(t *testing.T) {
	port := defaultDBPort(utils.MetaPostgres, utils.MetaDynamic)
	if port != "5432" {
		t.Errorf("Expected %v \n but received \n %v", "5432", port)
	}
}
func TestDataDbDefaultDBPortMySQL(t *testing.T) {
	port := defaultDBPort(utils.MetaMySQL, utils.MetaDynamic)
	if port != "3306" {
		t.Errorf("Expected %v \n but received \n %v", "3306", port)
	}
}
func TestDataDbDiff(t *testing.T) {
	dataDbCfg := &DbCfg{
		DBConns: DBConns{
			utils.MetaDefault: &DBConn{
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
				Opts: &DBOpts{
					RedisSentinel: "sentinel1",
				},
			},
		},
		Items: map[string]*ItemOpts{},
	}

	exp := &DbCfg{
		DBConns: DBConns{
			utils.MetaDefault: &DBConn{
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
				Opts: &DBOpts{
					RedisSentinel: "sentinel1",
				},
			},
		},
		Items: map[string]*ItemOpts{},
	}

	rcv := dataDbCfg.CloneSection()
	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(rcv))
	}
}

func TestItemOptsAsMapInterface(t *testing.T) {
	itm := &ItemOpts{
		TTL: 1,
	}

	exp := map[string]any{
		"dbConn":     "",
		"limit":      0,
		"remote":     false,
		"replicate":  false,
		"static_ttl": false,
		"ttl":        "1ns",
	}

	if rcv := itm.AsMapInterface(); !reflect.DeepEqual(rcv, exp) {
		t.Errorf("Expected <%+v>, \nReceived <%+v>", utils.ToJSON(exp), utils.ToJSON(rcv))
	}

}
