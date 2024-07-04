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

func TestMigratorCgrCfgloadFromJsonCfg(t *testing.T) {
	cfgJSON := &MigratorCfgJson{
		Out_dataDB_type:     utils.StringPointer(utils.MetaRedis),
		Out_dataDB_host:     utils.StringPointer("127.0.0.1"),
		Out_dataDB_port:     utils.StringPointer("6379"),
		Out_dataDB_name:     utils.StringPointer("10"),
		Out_dataDB_user:     utils.StringPointer(utils.CGRateSLwr),
		Out_dataDB_password: utils.StringPointer(utils.EmptyString),
		Out_dataDB_encoding: utils.StringPointer(utils.MsgPack),
		Out_storDB_type:     utils.StringPointer(utils.MetaMySQL),
		Out_storDB_host:     utils.StringPointer("127.0.0.1"),
		Out_storDB_port:     utils.StringPointer("3306"),
		Out_storDB_name:     utils.StringPointer(utils.CGRateSLwr),
		Out_storDB_user:     utils.StringPointer(utils.CGRateSLwr),
		Out_storDB_password: utils.StringPointer(utils.EmptyString),
		Out_dataDB_opts: &DBOptsJson{
			RedisCluster:            utils.BoolPointer(true),
			RedisClusterSync:        utils.StringPointer("10s"),
			RedisPoolPipelineWindow: utils.StringPointer("5µs"),
		},
		Out_storDB_opts: &DBOptsJson{
			SQLMaxOpenConns: utils.IntPointer(100),
		},
	}
	expected := &MigratorCgrCfg{
		OutDataDBType:     utils.MetaRedis,
		OutDataDBHost:     "127.0.0.1",
		OutDataDBPort:     "6379",
		OutDataDBName:     "10",
		OutDataDBUser:     utils.CGRateSLwr,
		OutDataDBPassword: utils.EmptyString,
		OutDataDBEncoding: utils.MsgPack,
		OutStorDBType:     utils.MetaMySQL,
		OutStorDBHost:     "127.0.0.1",
		OutStorDBPort:     "3306",
		OutStorDBName:     utils.CGRateSLwr,
		OutStorDBUser:     utils.CGRateSLwr,
		OutStorDBPassword: utils.EmptyString,
		OutDataDBOpts: &DataDBOpts{
			MongoConnScheme:         "mongodb",
			RedisMaxConns:           10,
			RedisConnectAttempts:    20,
			RedisSentinel:           utils.EmptyString,
			RedisCluster:            true,
			RedisClusterSync:        10 * time.Second,
			RedisClusterOndownDelay: 0,
			RedisPoolPipelineWindow: 5 * time.Microsecond,
			RedisConnectTimeout:     0,
			RedisReadTimeout:        0,
			RedisWriteTimeout:       0,
			RedisTLS:                false,
		},
		OutStorDBOpts: &StorDBOpts{
			SQLMaxOpenConns: 100,
			MongoConnScheme: "mongodb",
		},
	}
	cfg := NewDefaultCGRConfig()
	if err = cfg.migratorCgrCfg.loadFromJSONCfg(cfgJSON); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, cfg.migratorCgrCfg) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(cfg.migratorCgrCfg))
	}
}

func TestMigratorCgrCfgAsMapInterface(t *testing.T) {
	cfgJSONStr := `{
	"migrator": {
		"out_datadb_host": "127.0.0.19",
		"out_datadb_port": "8865",
		"out_datadb_name": "12",
		"out_stordb_host": "127.0.0.19",
		"out_stordb_port": "1234",
        "users_filters":["users","filters","Account"],
        "out_datadb_opts":{	
		   "redisCluster": true,					
		   "redisClusterSync": "2s",					
		   "redisClusterOndownDelay": "1",
		   "redisPoolPipelineWindow": "150µs",
		   "redisReadTimeout": "3s",
		   "redisWriteTimeout": "3s",	
		   "redisMaxConns": 5,
		   "redisConnectAttempts": 15,
		},
		"out_stordb_opts":{	
			"mysqlDSNParams": {
				"key": "value",
			},
			"pgSSLMode": "disable"					
		 },
	},
}`
	eMap := map[string]any{
		utils.OutDataDBTypeCfg:     "*redis",
		utils.OutDataDBHostCfg:     "127.0.0.19",
		utils.OutDataDBPortCfg:     "8865",
		utils.OutDataDBNameCfg:     "12",
		utils.OutDataDBUserCfg:     "cgrates",
		utils.OutDataDBPasswordCfg: "",
		utils.OutDataDBEncodingCfg: "msgpack",
		utils.OutStorDBTypeCfg:     "*mysql",
		utils.OutStorDBHostCfg:     "127.0.0.19",
		utils.OutStorDBPortCfg:     "1234",
		utils.OutStorDBNameCfg:     "cgrates",
		utils.OutStorDBUserCfg:     "cgrates",
		utils.OutStorDBPasswordCfg: "",
		utils.UsersFiltersCfg:      []string{"users", "filters", "Account"},
		utils.OutStorDBOptsCfg: map[string]any{
			utils.MongoQueryTimeoutCfg: "0s",
			utils.MongoConnSchemeCfg:   "mongodb",
			utils.MYSQLDSNParams: map[string]string{
				"key": "value",
			},
			utils.MysqlLocation:         utils.EmptyString,
			utils.PgSSLModeCfg:          "disable",
			utils.SQLConnMaxLifetimeCfg: "0s",
			utils.SQLMaxIdleConnsCfg:    0,
			utils.SQLMaxOpenConnsCfg:    0,
		},
		utils.OutDataDBOptsCfg: map[string]any{
			utils.MongoQueryTimeoutCfg:       "0s",
			utils.MongoConnSchemeCfg:         "mongodb",
			utils.RedisMaxConnsCfg:           5,
			utils.RedisConnectAttemptsCfg:    15,
			utils.RedisSentinelNameCfg:       "",
			utils.RedisClusterCfg:            true,
			utils.RedisClusterSyncCfg:        "2s",
			utils.RedisClusterOnDownDelayCfg: "1ns",
			utils.RedisPoolPipelineWindowCfg: "150µs",
			utils.RedisPoolPipelineLimitCfg:  0,
			utils.RedisConnectTimeoutCfg:     "0s",
			utils.RedisReadTimeoutCfg:        "3s",
			utils.RedisWriteTimeoutCfg:       "3s",
			utils.RedisTLS:                   false,
			utils.RedisClientCertificate:     "",
			utils.RedisClientKey:             "",
			utils.RedisCACertificate:         "",
		},
	}
	if cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cgrCfg.migratorCgrCfg.AsMapInterface(); !reflect.DeepEqual(eMap, rcv) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}
}
func TestMigratorCgrCfgAsMapInterface1(t *testing.T) {
	cfgJSONStr := `{
		"migrator": {
			"out_stordb_password": "out_stordb_password",
			"users_filters":["users","filters","Account"],
			"out_datadb_opts": {
				"redisSentinel": "out_datadb_redis_sentinel",
				"redisConnectTimeout": "5s",
				"redisPoolPipelineWindow": "1ms",
				"redisPoolPipelineLimit": 3
			},
		},
	}`
	eMap := map[string]any{
		utils.OutDataDBTypeCfg:     "*redis",
		utils.OutDataDBHostCfg:     "127.0.0.1",
		utils.OutDataDBPortCfg:     "6379",
		utils.OutDataDBNameCfg:     "10",
		utils.OutDataDBUserCfg:     "cgrates",
		utils.OutDataDBPasswordCfg: "",
		utils.OutDataDBEncodingCfg: "msgpack",
		utils.OutStorDBTypeCfg:     "*mysql",
		utils.OutStorDBHostCfg:     "127.0.0.1",
		utils.OutStorDBPortCfg:     "3306",
		utils.OutStorDBNameCfg:     "cgrates",
		utils.OutStorDBUserCfg:     "cgrates",
		utils.OutStorDBPasswordCfg: "out_stordb_password",
		utils.UsersFiltersCfg:      []string{"users", "filters", "Account"},
		utils.OutStorDBOptsCfg: map[string]any{
			utils.MongoQueryTimeoutCfg:  "0s",
			utils.MongoConnSchemeCfg:    "mongodb",
			utils.MYSQLDSNParams:        map[string]string(nil),
			utils.MysqlLocation:         utils.EmptyString,
			utils.PgSSLModeCfg:          utils.EmptyString,
			utils.SQLConnMaxLifetimeCfg: "0s",
			utils.SQLMaxIdleConnsCfg:    0,
			utils.SQLMaxOpenConnsCfg:    0,
		},
		utils.OutDataDBOptsCfg: map[string]any{
			utils.MongoQueryTimeoutCfg:       "0s",
			utils.MongoConnSchemeCfg:         "mongodb",
			utils.RedisMaxConnsCfg:           10,
			utils.RedisConnectAttemptsCfg:    20,
			utils.RedisSentinelNameCfg:       "out_datadb_redis_sentinel",
			utils.RedisClusterCfg:            false,
			utils.RedisClusterSyncCfg:        "5s",
			utils.RedisClusterOnDownDelayCfg: "0s",
			utils.RedisPoolPipelineWindowCfg: "1ms",
			utils.RedisPoolPipelineLimitCfg:  3,
			utils.RedisConnectTimeoutCfg:     "5s",
			utils.RedisReadTimeoutCfg:        "0s",
			utils.RedisWriteTimeoutCfg:       "0s",
			utils.RedisTLS:                   false,
			utils.RedisClientCertificate:     "",
			utils.RedisClientKey:             "",
			utils.RedisCACertificate:         "",
		},
	}
	if cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cgrCfg.migratorCgrCfg.AsMapInterface(); !reflect.DeepEqual(eMap, rcv) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}
}

func TestMigratorCgrCfgAsMapInterface2(t *testing.T) {
	cfgJSONStr := `{
		"migrator": {},
	}`
	eMap := map[string]any{
		utils.OutDataDBTypeCfg:     "*redis",
		utils.OutDataDBHostCfg:     "127.0.0.1",
		utils.OutDataDBPortCfg:     "6379",
		utils.OutDataDBNameCfg:     "10",
		utils.OutDataDBUserCfg:     "cgrates",
		utils.OutDataDBPasswordCfg: "",
		utils.OutDataDBEncodingCfg: "msgpack",
		utils.OutStorDBTypeCfg:     "*mysql",
		utils.OutStorDBHostCfg:     "127.0.0.1",
		utils.OutStorDBPortCfg:     "3306",
		utils.OutStorDBNameCfg:     "cgrates",
		utils.OutStorDBUserCfg:     "cgrates",
		utils.OutStorDBPasswordCfg: "",
		utils.UsersFiltersCfg:      []string(nil),
		utils.OutStorDBOptsCfg: map[string]any{
			utils.MongoQueryTimeoutCfg:  "0s",
			utils.MongoConnSchemeCfg:    "mongodb",
			utils.MYSQLDSNParams:        map[string]string(nil),
			utils.MysqlLocation:         utils.EmptyString,
			utils.PgSSLModeCfg:          utils.EmptyString,
			utils.SQLConnMaxLifetimeCfg: "0s",
			utils.SQLMaxIdleConnsCfg:    0,
			utils.SQLMaxOpenConnsCfg:    0,
		},
		utils.OutDataDBOptsCfg: map[string]any{
			utils.MongoQueryTimeoutCfg:       "0s",
			utils.MongoConnSchemeCfg:         "mongodb",
			utils.RedisMaxConnsCfg:           10,
			utils.RedisConnectAttemptsCfg:    20,
			utils.RedisSentinelNameCfg:       "",
			utils.RedisClusterCfg:            false,
			utils.RedisClusterSyncCfg:        "5s",
			utils.RedisClusterOnDownDelayCfg: "0s",
			utils.RedisPoolPipelineWindowCfg: "150µs",
			utils.RedisPoolPipelineLimitCfg:  0,
			utils.RedisConnectTimeoutCfg:     "0s",
			utils.RedisReadTimeoutCfg:        "0s",
			utils.RedisWriteTimeoutCfg:       "0s",
			utils.RedisTLS:                   false,
			utils.RedisClientCertificate:     "",
			utils.RedisClientKey:             "",
			utils.RedisCACertificate:         "",
		},
	}
	if cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cgrCfg.migratorCgrCfg.AsMapInterface(); !reflect.DeepEqual(eMap, rcv) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}

}

func TestMigratorCgrCfgClone(t *testing.T) {
	sa := &MigratorCgrCfg{
		OutDataDBType:     utils.MetaRedis,
		OutDataDBHost:     "127.0.0.1",
		OutDataDBPort:     "6379",
		OutDataDBName:     "10",
		OutDataDBUser:     utils.CGRateSLwr,
		OutDataDBPassword: utils.EmptyString,
		OutDataDBEncoding: utils.MsgPack,
		OutStorDBType:     utils.MetaMySQL,
		OutStorDBHost:     "127.0.0.1",
		OutStorDBPort:     "3306",
		OutStorDBName:     utils.CGRateSLwr,
		OutStorDBUser:     utils.CGRateSLwr,
		OutStorDBPassword: utils.EmptyString,
		UsersFilters:      []string{utils.AccountField},
		OutDataDBOpts: &DataDBOpts{
			RedisMaxConns:           10,
			RedisConnectAttempts:    20,
			RedisSentinel:           utils.EmptyString,
			RedisCluster:            true,
			RedisClusterSync:        10 * time.Second,
			RedisClusterOndownDelay: 0,
			RedisPoolPipelineWindow: 150 * time.Microsecond,
			RedisConnectTimeout:     0,
			RedisReadTimeout:        0,
			RedisWriteTimeout:       0,
			MongoQueryTimeout:       10 * time.Second,
			RedisTLS:                false,
		},
		OutStorDBOpts: &StorDBOpts{},
	}
	rcv := sa.Clone()
	if !reflect.DeepEqual(sa, rcv) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(sa), utils.ToJSON(rcv))
	}
	if rcv.UsersFilters[0] = ""; sa.UsersFilters[0] != utils.AccountField {
		t.Errorf("Expected clone to not modify the cloned")
	}
	if rcv.OutDataDBOpts.RedisSentinel = "1"; sa.OutDataDBOpts.RedisSentinel != "" {
		t.Errorf("Expected clone to not modify the cloned")
	}
	if rcv.OutStorDBOpts.PgSSLMode = "1"; sa.OutStorDBOpts.PgSSLMode != "" {
		t.Errorf("Expected clone to not modify the cloned")
	}
}
