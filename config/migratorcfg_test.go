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
		Out_dataDB_opts: &DBOptsJson{
			RedisCluster:     utils.BoolPointer(true),
			RedisClusterSync: utils.StringPointer("10s"),
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

		OutDataDBOpts: &DataDBOpts{
			RedisMaxConns:           10,
			RedisConnectAttempts:    20,
			RedisCluster:            true,
			RedisClusterSync:        10 * time.Second,
			RedisPoolPipelineWindow: 150 * time.Microsecond,
			MongoConnScheme:         "mongodb",
			MongoQueryTimeout:       10 * time.Second,
		},
	}
	cfg := NewDefaultCGRConfig()
	if err := cfg.migratorCgrCfg.loadFromJSONCfg(cfgJSON); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, cfg.migratorCgrCfg) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(cfg.migratorCgrCfg))
	}
	cfgJSON = nil
	if err := cfg.migratorCgrCfg.loadFromJSONCfg(cfgJSON); err != nil {
		t.Error(err)
	}
}

func TestMigratorCgrCfgAsMapInterface(t *testing.T) {
	cfgJSONStr := `{
	"migrator": {
		"out_datadb_host": "127.0.0.19",
		"out_datadb_port": "8865",
		"out_datadb_name": "12",

        "users_filters":["users","filters","Account"],
        "out_datadb_opts":{	
		   "redisCluster": true,
		   "redisClusterSync": "2s",
		   "redisClusterOndownDelay": "1",
		   "redisConnectTimeout": "5s",
		   "redisReadTimeout": "5s",
		   "redisWriteTimeout": "5s",
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

		utils.UsersFiltersCfg: []string{"users", "filters", "Account"},

		utils.OutDataDBOptsCfg: map[string]any{
			utils.RedisMaxConnsCfg:           10,
			utils.RedisConnectAttemptsCfg:    20,
			utils.RedisSentinelNameCfg:       "",
			utils.RedisClusterCfg:            true,
			utils.RedisClusterSyncCfg:        "2s",
			utils.RedisClusterOnDownDelayCfg: "1ns",
			utils.RedisConnectTimeoutCfg:     "5s",
			utils.RedisReadTimeoutCfg:        "5s",
			utils.RedisWriteTimeoutCfg:       "5s",
			utils.RedisPoolPipelineLimitCfg:  0,
			utils.RedisPoolPipelineWindowCfg: "150µs",
			utils.RedisTLSCfg:                false,
			utils.RedisClientCertificateCfg:  "",
			utils.RedisClientKeyCfg:          "",
			utils.RedisCACertificateCfg:      "",
			utils.MongoQueryTimeoutCfg:       "10s",
			utils.MongoConnSchemeCfg:         "mongodb",
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
			"users_filters":["users","filters","Account"],
			"out_datadb_opts": {
				"redisSentinel": "out_datadb_redis_sentinel",
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

		utils.UsersFiltersCfg: []string{"users", "filters", "Account"},

		utils.OutDataDBOptsCfg: map[string]any{
			utils.RedisMaxConnsCfg:           10,
			utils.RedisConnectAttemptsCfg:    20,
			utils.RedisSentinelNameCfg:       "out_datadb_redis_sentinel",
			utils.RedisClusterCfg:            false,
			utils.RedisClusterSyncCfg:        "5s",
			utils.RedisClusterOnDownDelayCfg: "0s",
			utils.RedisConnectTimeoutCfg:     "0s",
			utils.RedisReadTimeoutCfg:        "0s",
			utils.RedisWriteTimeoutCfg:       "0s",
			utils.RedisPoolPipelineLimitCfg:  0,
			utils.RedisPoolPipelineWindowCfg: "150µs",
			utils.RedisTLSCfg:                false,
			utils.RedisClientCertificateCfg:  "",
			utils.RedisClientKeyCfg:          "",
			utils.RedisCACertificateCfg:      "",
			utils.MongoQueryTimeoutCfg:       "10s",
			utils.MongoConnSchemeCfg:         "mongodb",
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

		utils.UsersFiltersCfg: []string(nil),

		utils.OutDataDBOptsCfg: map[string]any{
			utils.RedisMaxConnsCfg:           10,
			utils.RedisConnectAttemptsCfg:    20,
			utils.RedisSentinelNameCfg:       "",
			utils.RedisClusterCfg:            false,
			utils.RedisClusterSyncCfg:        "5s",
			utils.RedisClusterOnDownDelayCfg: "0s",
			utils.RedisTLSCfg:                false,
			utils.RedisClientCertificateCfg:  "",
			utils.RedisConnectTimeoutCfg:     "0s",
			utils.RedisReadTimeoutCfg:        "0s",
			utils.RedisWriteTimeoutCfg:       "0s",
			utils.RedisPoolPipelineLimitCfg:  0,
			utils.RedisPoolPipelineWindowCfg: "150µs",
			utils.RedisClientKeyCfg:          "",
			utils.RedisCACertificateCfg:      "",
			utils.MongoQueryTimeoutCfg:       "10s",
			utils.MongoConnSchemeCfg:         "mongodb",
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
		UsersFilters:      []string{utils.AccountField},
		OutDataDBOpts: &DataDBOpts{
			RedisCluster:     true,
			RedisClusterSync: 10 * time.Second,
		},
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
}

func TestDiffMigratorCfgJson(t *testing.T) {
	var d *MigratorCfgJson

	v1 := &MigratorCgrCfg{
		OutDataDBType:     "postgres",
		OutDataDBHost:     "127.0.0.1",
		OutDataDBPort:     "8080",
		OutDataDBName:     "cgrates",
		OutDataDBUser:     "cgrates_user",
		OutDataDBPassword: "CGRateS.org",
		OutDataDBEncoding: "utf-8",
		OutDataDBOpts:     &DataDBOpts{},

		UsersFilters: []string{},
	}

	v2 := &MigratorCgrCfg{

		OutDataDBEncoding: "utf-16",
		OutDataDBType:     "redis",
		OutDataDBHost:     "0.0.0.0",
		OutDataDBPort:     "4037",
		OutDataDBName:     "cgrates_redis",
		OutDataDBUser:     "cgrates_redis_user",
		OutDataDBPassword: "CGRateS.org_redis",
		OutDataDBOpts: &DataDBOpts{
			RedisCluster: true,
		},
		UsersFilters: []string{"cgrates_redis_user"},
	}

	expected := &MigratorCfgJson{

		Out_dataDB_encoding: utils.StringPointer("utf-16"),
		Out_dataDB_type:     utils.StringPointer("redis"),
		Out_dataDB_host:     utils.StringPointer("0.0.0.0"),
		Out_dataDB_port:     utils.StringPointer("4037"),
		Out_dataDB_name:     utils.StringPointer("cgrates_redis"),
		Out_dataDB_user:     utils.StringPointer("cgrates_redis_user"),
		Out_dataDB_password: utils.StringPointer("CGRateS.org_redis"),
		Out_dataDB_opts: &DBOptsJson{
			RedisCluster: utils.BoolPointer(true),
		},

		Users_filters: &[]string{"cgrates_redis_user"},
	}

	rcv := diffMigratorCfgJson(d, v1, v2)
	if !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}

	v1 = v2
	expected = &MigratorCfgJson{
		Out_dataDB_opts: &DBOptsJson{},
	}
	rcv = diffMigratorCfgJson(d, v1, v2)
	if !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}

func TestMigratorCloneSection(t *testing.T) {
	mgrCfg := &MigratorCgrCfg{
		OutDataDBType:     "postgres",
		OutDataDBHost:     "127.0.0.1",
		OutDataDBPort:     "8080",
		OutDataDBName:     "cgrates",
		OutDataDBUser:     "cgrates_user",
		OutDataDBPassword: "CGRateS.org",
		OutDataDBEncoding: "utf-8",

		OutDataDBOpts: &DataDBOpts{},

		UsersFilters: []string{},
	}

	exp := &MigratorCgrCfg{
		OutDataDBType:     "postgres",
		OutDataDBHost:     "127.0.0.1",
		OutDataDBPort:     "8080",
		OutDataDBName:     "cgrates",
		OutDataDBUser:     "cgrates_user",
		OutDataDBPassword: "CGRateS.org",
		OutDataDBEncoding: "utf-8",

		OutDataDBOpts: &DataDBOpts{},

		UsersFilters: []string{},
	}

	rcv := mgrCfg.CloneSection()
	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(rcv))
	}
}
