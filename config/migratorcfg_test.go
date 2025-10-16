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
		FromItems: map[string]*FromItemJson{
			utils.CacheVersions: {utils.StringPointer("ConnID1")},
		},
		Out_db_opts: &DBOptsJson{
			RedisCluster:     utils.BoolPointer(true),
			RedisClusterSync: utils.StringPointer("10s"),
		},
	}
	expected := &MigratorCgrCfg{
		FromItems: map[string]*MigratorFromItem{
			utils.MetaAccounts:          {DBConn: utils.MetaDefault},
			utils.MetaChargerProfiles:   {DBConn: utils.MetaDefault},
			utils.MetaFilters:           {DBConn: utils.MetaDefault},
			utils.MetaLoadIDs:           {DBConn: utils.MetaDefault},
			utils.MetaStatQueueProfiles: {DBConn: utils.MetaDefault},
			utils.CacheVersions:         {DBConn: "ConnID1"},
		},
		OutDBOpts: &DBOpts{
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
		"fromItems": {
			"*versions": {"dbConn": "someDBID"},
		},
        "users_filters":["users","filters","Account"],
        "out_db_opts":{	
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

		utils.UsersFiltersCfg: []string{"users", "filters", "Account"},
		utils.FromItemsCfg: map[string]any{
			utils.MetaAccounts:          map[string]any{utils.DBConnCfg: utils.MetaDefault},
			utils.MetaChargerProfiles:   map[string]any{utils.DBConnCfg: utils.MetaDefault},
			utils.MetaFilters:           map[string]any{utils.DBConnCfg: utils.MetaDefault},
			utils.MetaLoadIDs:           map[string]any{utils.DBConnCfg: utils.MetaDefault},
			utils.MetaStatQueueProfiles: map[string]any{utils.DBConnCfg: utils.MetaDefault},
			utils.CacheVersions:         map[string]any{utils.DBConnCfg: "someDBID"},
		},
		utils.OutDBOptsCfg: map[string]any{
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
			"fromItems": {
				"*versions": {"dbConn": "someDBID"},
			},
			"out_db_opts": {
				"redisSentinel": "out_datadb_redis_sentinel",
			},
		},
	}`
	eMap := map[string]any{

		utils.UsersFiltersCfg: []string{"users", "filters", "Account"},
		utils.FromItemsCfg: map[string]any{
			utils.MetaAccounts:          map[string]any{utils.DBConnCfg: utils.MetaDefault},
			utils.MetaChargerProfiles:   map[string]any{utils.DBConnCfg: utils.MetaDefault},
			utils.MetaFilters:           map[string]any{utils.DBConnCfg: utils.MetaDefault},
			utils.MetaLoadIDs:           map[string]any{utils.DBConnCfg: utils.MetaDefault},
			utils.MetaStatQueueProfiles: map[string]any{utils.DBConnCfg: utils.MetaDefault},
			utils.CacheVersions:         map[string]any{utils.DBConnCfg: "someDBID"},
		},
		utils.OutDBOptsCfg: map[string]any{
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

		utils.UsersFiltersCfg: []string(nil),
		utils.FromItemsCfg: map[string]any{
			utils.MetaAccounts:          map[string]any{utils.DBConnCfg: utils.MetaDefault},
			utils.MetaChargerProfiles:   map[string]any{utils.DBConnCfg: utils.MetaDefault},
			utils.MetaFilters:           map[string]any{utils.DBConnCfg: utils.MetaDefault},
			utils.MetaLoadIDs:           map[string]any{utils.DBConnCfg: utils.MetaDefault},
			utils.MetaStatQueueProfiles: map[string]any{utils.DBConnCfg: utils.MetaDefault},
			utils.CacheVersions:         map[string]any{utils.DBConnCfg: utils.MetaDefault},
		},
		utils.OutDBOptsCfg: map[string]any{
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
		UsersFilters: []string{utils.AccountField},
		FromItems: map[string]*MigratorFromItem{
			utils.CacheVersions: {DBConn: utils.MetaDefault},
		},
		OutDBOpts: &DBOpts{
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
	if rcv.OutDBOpts.RedisSentinel = "1"; sa.OutDBOpts.RedisSentinel != "" {
		t.Errorf("Expected clone to not modify the cloned")
	}
}

func TestDiffMigratorCfgJson(t *testing.T) {
	var d *MigratorCfgJson

	v1 := &MigratorCgrCfg{
		OutDBOpts:    &DBOpts{},
		FromItems:    map[string]*MigratorFromItem{},
		UsersFilters: []string{},
	}

	v2 := &MigratorCgrCfg{

		OutDBOpts: &DBOpts{
			RedisCluster: true,
		},
		FromItems:    map[string]*MigratorFromItem{},
		UsersFilters: []string{"cgrates_redis_user"},
	}

	expected := &MigratorCfgJson{

		Out_db_opts: &DBOptsJson{
			RedisCluster: utils.BoolPointer(true),
		},
		FromItems:     map[string]*FromItemJson{},
		Users_filters: &[]string{"cgrates_redis_user"},
	}

	rcv := diffMigratorCfgJson(d, v1, v2)
	if !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}

	v1 = v2
	expected = &MigratorCfgJson{
		FromItems:   map[string]*FromItemJson{},
		Out_db_opts: &DBOptsJson{},
	}
	rcv = diffMigratorCfgJson(d, v1, v2)
	if !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}

func TestMigratorCloneSection(t *testing.T) {
	mgrCfg := &MigratorCgrCfg{

		OutDBOpts:    &DBOpts{},
		FromItems:    map[string]*MigratorFromItem{},
		UsersFilters: []string{},
	}

	exp := &MigratorCgrCfg{

		OutDBOpts:    &DBOpts{},
		FromItems:    map[string]*MigratorFromItem{},
		UsersFilters: []string{},
	}

	rcv := mgrCfg.CloneSection()
	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(rcv))
	}
}
