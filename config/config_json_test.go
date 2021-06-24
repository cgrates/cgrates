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
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

func TestDfGeneralJsonCfg(t *testing.T) {
	eCfg := &GeneralJsonCfg{
		Node_id:              utils.StringPointer(""),
		Logger:               utils.StringPointer(utils.MetaSysLog),
		Log_level:            utils.IntPointer(utils.LOGLEVEL_INFO),
		Rounding_decimals:    utils.IntPointer(5),
		Dbdata_encoding:      utils.StringPointer("*msgpack"),
		Tpexport_dir:         utils.StringPointer("/var/spool/cgrates/tpe"),
		Poster_attempts:      utils.IntPointer(3),
		Failed_posts_dir:     utils.StringPointer("/var/spool/cgrates/failed_posts"),
		Failed_posts_ttl:     utils.StringPointer("5s"),
		Default_request_type: utils.StringPointer(utils.MetaRated),
		Default_category:     utils.StringPointer("call"),
		Default_tenant:       utils.StringPointer("cgrates.org"),
		Default_caching:      utils.StringPointer(utils.MetaReload),
		Default_timezone:     utils.StringPointer("Local"),
		Connect_attempts:     utils.IntPointer(5),
		Reconnects:           utils.IntPointer(-1),
		Connect_timeout:      utils.StringPointer("1s"),
		Reply_timeout:        utils.StringPointer("2s"),
		Locking_timeout:      utils.StringPointer("0"),
		Digest_separator:     utils.StringPointer(","),
		Digest_equal:         utils.StringPointer(":"),
		Rsr_separator:        utils.StringPointer(";"),
		Max_parallel_conns:   utils.IntPointer(100),
	}
	dfCgrJSONCfg, err := NewCgrJsonCfgFromBytes([]byte(CGRATES_CFG_JSON))
	if err != nil {
		t.Error(err)
	}
	if gCfg, err := dfCgrJSONCfg.GeneralJsonCfg(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCfg, gCfg) {
		t.Errorf("expecting: %s, \nreceived: %s", utils.ToIJSON(eCfg), utils.ToIJSON(gCfg))
	}
}

func TestDfCoreSJsonCfg(t *testing.T) {
	eCfg := &CoreSJsonCfg{
		Caps:                utils.IntPointer(0),
		Caps_strategy:       utils.StringPointer(utils.MetaBusy),
		Caps_stats_interval: utils.StringPointer("0"),
		Shutdown_timeout:    utils.StringPointer("1s"),
	}
	dfCgrJSONCfg, err := NewCgrJsonCfgFromBytes([]byte(CGRATES_CFG_JSON))
	if err != nil {
		t.Error(err)
	}
	if gCfg, err := dfCgrJSONCfg.CoreSJSON(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCfg, gCfg) {
		t.Errorf("expecting: %s, \nreceived: %s", utils.ToIJSON(eCfg), utils.ToIJSON(gCfg))
	}
}

func TestCacheJsonCfg(t *testing.T) {
	eCfg := &CacheJsonCfg{
		Partitions: map[string]*CacheParamJsonCfg{
			utils.CacheResourceProfiles: {Limit: utils.IntPointer(-1),
				Ttl: utils.StringPointer(""), Static_ttl: utils.BoolPointer(false),
				Precache: utils.BoolPointer(false), Replicate: utils.BoolPointer(false)},
			utils.CacheResources: {Limit: utils.IntPointer(-1),
				Ttl: utils.StringPointer(""), Static_ttl: utils.BoolPointer(false),
				Precache: utils.BoolPointer(false), Replicate: utils.BoolPointer(false)},
			utils.CacheEventResources: {Limit: utils.IntPointer(-1),
				Ttl: utils.StringPointer(""), Static_ttl: utils.BoolPointer(false),
				Replicate: utils.BoolPointer(false)},
			utils.CacheStatQueueProfiles: {Limit: utils.IntPointer(-1),
				Ttl: utils.StringPointer(""), Static_ttl: utils.BoolPointer(false),
				Precache: utils.BoolPointer(false), Replicate: utils.BoolPointer(false)},
			utils.CacheStatQueues: {Limit: utils.IntPointer(-1),
				Ttl: utils.StringPointer(""), Static_ttl: utils.BoolPointer(false),
				Precache: utils.BoolPointer(false), Replicate: utils.BoolPointer(false)},
			utils.CacheThresholdProfiles: {Limit: utils.IntPointer(-1),
				Ttl: utils.StringPointer(""), Static_ttl: utils.BoolPointer(false),
				Precache: utils.BoolPointer(false), Replicate: utils.BoolPointer(false)},
			utils.CacheThresholds: {Limit: utils.IntPointer(-1),
				Ttl: utils.StringPointer(""), Static_ttl: utils.BoolPointer(false),
				Precache: utils.BoolPointer(false), Replicate: utils.BoolPointer(false)},
			utils.CacheFilters: {Limit: utils.IntPointer(-1),
				Ttl: utils.StringPointer(""), Static_ttl: utils.BoolPointer(false),
				Precache: utils.BoolPointer(false), Replicate: utils.BoolPointer(false)},
			utils.CacheRouteProfiles: {Limit: utils.IntPointer(-1),
				Ttl: utils.StringPointer(""), Static_ttl: utils.BoolPointer(false),
				Precache: utils.BoolPointer(false), Replicate: utils.BoolPointer(false)},
			utils.CacheAttributeProfiles: {Limit: utils.IntPointer(-1),
				Ttl: utils.StringPointer(""), Static_ttl: utils.BoolPointer(false),
				Precache: utils.BoolPointer(false), Replicate: utils.BoolPointer(false)},
			utils.CacheChargerProfiles: {Limit: utils.IntPointer(-1),
				Ttl: utils.StringPointer(""), Static_ttl: utils.BoolPointer(false),
				Precache: utils.BoolPointer(false), Replicate: utils.BoolPointer(false)},
			utils.CacheDispatcherProfiles: {Limit: utils.IntPointer(-1),
				Ttl: utils.StringPointer(""), Static_ttl: utils.BoolPointer(false),
				Precache: utils.BoolPointer(false), Replicate: utils.BoolPointer(false)},
			utils.CacheRateProfiles: {Limit: utils.IntPointer(-1),
				Ttl: utils.StringPointer(""), Static_ttl: utils.BoolPointer(false),
				Precache: utils.BoolPointer(false), Replicate: utils.BoolPointer(false)},
			utils.CacheActionProfiles: {Limit: utils.IntPointer(-1),
				Ttl: utils.StringPointer(""), Static_ttl: utils.BoolPointer(false),
				Precache: utils.BoolPointer(false), Replicate: utils.BoolPointer(false)},
			utils.CacheAccounts: {Limit: utils.IntPointer(-1),
				Ttl: utils.StringPointer(""), Static_ttl: utils.BoolPointer(false),
				Precache: utils.BoolPointer(false), Replicate: utils.BoolPointer(false)},
			utils.CacheDispatcherHosts: {Limit: utils.IntPointer(-1),
				Ttl: utils.StringPointer(""), Static_ttl: utils.BoolPointer(false),
				Precache: utils.BoolPointer(false), Replicate: utils.BoolPointer(false)},
			utils.CacheResourceFilterIndexes: {Limit: utils.IntPointer(-1),
				Ttl: utils.StringPointer(""), Static_ttl: utils.BoolPointer(false),
				Replicate: utils.BoolPointer(false)},
			utils.CacheStatFilterIndexes: {Limit: utils.IntPointer(-1),
				Ttl: utils.StringPointer(""), Static_ttl: utils.BoolPointer(false),
				Replicate: utils.BoolPointer(false)},
			utils.CacheThresholdFilterIndexes: {Limit: utils.IntPointer(-1),
				Ttl: utils.StringPointer(""), Static_ttl: utils.BoolPointer(false),
				Replicate: utils.BoolPointer(false)},
			utils.CacheRouteFilterIndexes: {Limit: utils.IntPointer(-1),
				Ttl: utils.StringPointer(""), Static_ttl: utils.BoolPointer(false),
				Replicate: utils.BoolPointer(false)},
			utils.CacheAttributeFilterIndexes: {Limit: utils.IntPointer(-1),
				Ttl: utils.StringPointer(""), Static_ttl: utils.BoolPointer(false),
				Replicate: utils.BoolPointer(false)},
			utils.CacheChargerFilterIndexes: {Limit: utils.IntPointer(-1),
				Ttl: utils.StringPointer(""), Static_ttl: utils.BoolPointer(false),
				Replicate: utils.BoolPointer(false)},
			utils.CacheDispatcherFilterIndexes: {Limit: utils.IntPointer(-1),
				Ttl: utils.StringPointer(""), Static_ttl: utils.BoolPointer(false),
				Replicate: utils.BoolPointer(false)},
			utils.CacheRateProfilesFilterIndexes: {Limit: utils.IntPointer(-1),
				Ttl: utils.StringPointer(""), Static_ttl: utils.BoolPointer(false),
				Replicate: utils.BoolPointer(false)},
			utils.CacheRateFilterIndexes: {Limit: utils.IntPointer(-1),
				Ttl: utils.StringPointer(""), Static_ttl: utils.BoolPointer(false),
				Replicate: utils.BoolPointer(false)},
			utils.CacheActionProfilesFilterIndexes: {Limit: utils.IntPointer(-1),
				Ttl: utils.StringPointer(""), Static_ttl: utils.BoolPointer(false),
				Replicate: utils.BoolPointer(false)},
			utils.CacheAccountsFilterIndexes: {Limit: utils.IntPointer(-1),
				Ttl: utils.StringPointer(""), Static_ttl: utils.BoolPointer(false),
				Replicate: utils.BoolPointer(false)},
			utils.CacheReverseFilterIndexes: {Limit: utils.IntPointer(-1),
				Ttl: utils.StringPointer(""), Static_ttl: utils.BoolPointer(false),
				Replicate: utils.BoolPointer(false)},
			utils.CacheDispatcherRoutes: {Limit: utils.IntPointer(-1),
				Ttl: utils.StringPointer(""), Static_ttl: utils.BoolPointer(false),
				Replicate: utils.BoolPointer(false)},
			utils.CacheDispatcherLoads: {Limit: utils.IntPointer(-1),
				Ttl: utils.StringPointer(""), Static_ttl: utils.BoolPointer(false),
				Replicate: utils.BoolPointer(false)},
			utils.CacheDispatchers: {Limit: utils.IntPointer(-1),
				Ttl: utils.StringPointer(""), Static_ttl: utils.BoolPointer(false),
				Replicate: utils.BoolPointer(false)},
			utils.CacheDiameterMessages: {Limit: utils.IntPointer(-1),
				Ttl: utils.StringPointer("3h"), Static_ttl: utils.BoolPointer(false),
				Replicate: utils.BoolPointer(false)},
			utils.CacheRPCResponses: {Limit: utils.IntPointer(0),
				Ttl: utils.StringPointer("2s"), Static_ttl: utils.BoolPointer(false),
				Replicate: utils.BoolPointer(false)},
			utils.CacheClosedSessions: {Limit: utils.IntPointer(-1),
				Ttl: utils.StringPointer("10s"), Static_ttl: utils.BoolPointer(false),
				Replicate: utils.BoolPointer(false)},
			utils.CacheEventCharges: {Limit: utils.IntPointer(-1),
				Ttl: utils.StringPointer("10s"), Static_ttl: utils.BoolPointer(false),
				Replicate: utils.BoolPointer(false)},
			utils.CacheCDRIDs: {Limit: utils.IntPointer(-1),
				Ttl: utils.StringPointer("10m"), Static_ttl: utils.BoolPointer(false),
				Replicate: utils.BoolPointer(false)},
			utils.CacheLoadIDs: {Limit: utils.IntPointer(-1),
				Ttl: utils.StringPointer(""), Static_ttl: utils.BoolPointer(false),
				Precache: utils.BoolPointer(false), Replicate: utils.BoolPointer(false)},
			utils.CacheRPCConnections: {Limit: utils.IntPointer(-1),
				Ttl: utils.StringPointer(""), Static_ttl: utils.BoolPointer(false),
				Replicate: utils.BoolPointer(false)},
			utils.CacheUCH: {Limit: utils.IntPointer(-1),
				Ttl: utils.StringPointer("3h"), Static_ttl: utils.BoolPointer(false),
				Replicate: utils.BoolPointer(false)},
			utils.CacheSTIR: {Limit: utils.IntPointer(-1),
				Ttl: utils.StringPointer("3h"), Static_ttl: utils.BoolPointer(false),
				Replicate: utils.BoolPointer(false)},
			utils.CacheCapsEvents: {Limit: utils.IntPointer(-1),
				Ttl: utils.StringPointer(""), Static_ttl: utils.BoolPointer(false),
				Replicate: utils.BoolPointer(false),
			},
			utils.CacheVersions: {Limit: utils.IntPointer(-1),
				Ttl: utils.StringPointer(""), Static_ttl: utils.BoolPointer(false),
				Replicate: utils.BoolPointer(false)},
			utils.CacheTBLTPResources: {Limit: utils.IntPointer(-1),
				Ttl: utils.StringPointer(""), Static_ttl: utils.BoolPointer(false),
				Replicate: utils.BoolPointer(false)},
			utils.CacheTBLTPThresholds: {Limit: utils.IntPointer(-1),
				Ttl: utils.StringPointer(""), Static_ttl: utils.BoolPointer(false),
				Replicate: utils.BoolPointer(false)},
			utils.CacheTBLTPFilters: {Limit: utils.IntPointer(-1),
				Ttl: utils.StringPointer(""), Static_ttl: utils.BoolPointer(false),
				Replicate: utils.BoolPointer(false)},
			utils.CacheSessionCostsTBL: {Limit: utils.IntPointer(-1),
				Ttl: utils.StringPointer(""), Static_ttl: utils.BoolPointer(false),
				Replicate: utils.BoolPointer(false)},
			utils.CacheCDRsTBL: {Limit: utils.IntPointer(-1),
				Ttl: utils.StringPointer(""), Static_ttl: utils.BoolPointer(false),
				Replicate: utils.BoolPointer(false)},
			utils.CacheTBLTPRoutes: {Limit: utils.IntPointer(-1),
				Ttl: utils.StringPointer(""), Static_ttl: utils.BoolPointer(false),
				Replicate: utils.BoolPointer(false)},
			utils.CacheTBLTPStats: {Limit: utils.IntPointer(-1),
				Ttl: utils.StringPointer(""), Static_ttl: utils.BoolPointer(false),
				Replicate: utils.BoolPointer(false)},
			utils.CacheTBLTPAttributes: {Limit: utils.IntPointer(-1),
				Ttl: utils.StringPointer(""), Static_ttl: utils.BoolPointer(false),
				Replicate: utils.BoolPointer(false)},
			utils.CacheTBLTPChargers: {Limit: utils.IntPointer(-1),
				Ttl: utils.StringPointer(""), Static_ttl: utils.BoolPointer(false),
				Replicate: utils.BoolPointer(false)},
			utils.CacheTBLTPDispatchers: {Limit: utils.IntPointer(-1),
				Ttl: utils.StringPointer(""), Static_ttl: utils.BoolPointer(false),
				Replicate: utils.BoolPointer(false)},
			utils.CacheTBLTPDispatcherHosts: {Limit: utils.IntPointer(-1),
				Ttl: utils.StringPointer(""), Static_ttl: utils.BoolPointer(false),
				Replicate: utils.BoolPointer(false)},
			utils.CacheTBLTPRateProfiles: {Limit: utils.IntPointer(-1),
				Ttl: utils.StringPointer(""), Static_ttl: utils.BoolPointer(false),
				Replicate: utils.BoolPointer(false)},
			utils.CacheTBLTPActionProfiles: {Limit: utils.IntPointer(-1),
				Ttl: utils.StringPointer(""), Static_ttl: utils.BoolPointer(false),
				Replicate: utils.BoolPointer(false)},
			utils.CacheTBLTPAccounts: {Limit: utils.IntPointer(-1),
				Ttl: utils.StringPointer(""), Static_ttl: utils.BoolPointer(false),
				Replicate: utils.BoolPointer(false)},
			utils.MetaAPIBan: {Limit: utils.IntPointer(-1),
				Ttl: utils.StringPointer("2m"), Static_ttl: utils.BoolPointer(false),
				Replicate: utils.BoolPointer(false)},
			utils.CacheReplicationHosts: {Limit: utils.IntPointer(0),
				Ttl: utils.StringPointer(""), Static_ttl: utils.BoolPointer(false),
				Replicate: utils.BoolPointer(false)},
		},
		Replication_conns: &[]string{},
	}
	dfCgrJSONCfg, err := NewCgrJsonCfgFromBytes([]byte(CGRATES_CFG_JSON))
	if err != nil {
		t.Error(err)
	}

	if gCfg, err := dfCgrJSONCfg.CacheJsonCfg(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCfg, gCfg) {
		t.Errorf("expected: %s\n, received: %s", utils.ToJSON(eCfg), utils.ToJSON(gCfg))
	}
}

func TestDfListenJsonCfg(t *testing.T) {
	eCfg := &ListenJsonCfg{
		Rpc_json:     utils.StringPointer("127.0.0.1:2012"),
		Rpc_gob:      utils.StringPointer("127.0.0.1:2013"),
		Http:         utils.StringPointer("127.0.0.1:2080"),
		Rpc_json_tls: utils.StringPointer("127.0.0.1:2022"),
		Rpc_gob_tls:  utils.StringPointer("127.0.0.1:2023"),
		Http_tls:     utils.StringPointer("127.0.0.1:2280"),
	}
	dfCgrJSONCfg, err := NewCgrJsonCfgFromBytes([]byte(CGRATES_CFG_JSON))
	if err != nil {
		t.Error(err)
	}
	if cfg, err := dfCgrJSONCfg.ListenJsonCfg(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCfg, cfg) {
		t.Error("Received: ", cfg)
	}
}

func TestDfDataDbJsonCfg(t *testing.T) {
	eCfg := &DbJsonCfg{
		Db_type:              utils.StringPointer("*redis"),
		Db_host:              utils.StringPointer("127.0.0.1"),
		Db_port:              utils.IntPointer(6379),
		Db_name:              utils.StringPointer("10"),
		Db_user:              utils.StringPointer("cgrates"),
		Db_password:          utils.StringPointer(""),
		Replication_conns:    &[]string{},
		Remote_conns:         &[]string{},
		Replication_filtered: utils.BoolPointer(false),
		Remote_conn_id:       utils.StringPointer(""),
		Replication_cache:    utils.StringPointer(""),
		Opts: map[string]interface{}{
			utils.RedisSentinelNameCfg:       "",
			utils.MongoQueryTimeoutCfg:       "10s",
			utils.RedisClusterCfg:            false,
			utils.RedisClusterOnDownDelayCfg: "0",
			utils.RedisClusterSyncCfg:        "5s",
			utils.RedisTLS:                   false,
			utils.RedisClientCertificate:     "",
			utils.RedisClientKey:             "",
			utils.RedisCACertificate:         "",
		},
		Items: map[string]*ItemOptJson{
			utils.MetaAccounts: {
				Replicate: utils.BoolPointer(false),
				Remote:    utils.BoolPointer(false),
			},
			utils.MetaActions: {
				Replicate: utils.BoolPointer(false),
				Remote:    utils.BoolPointer(false),
			},
			utils.MetaResourceProfile: {
				Replicate: utils.BoolPointer(false),
				Remote:    utils.BoolPointer(false),
			},
			utils.MetaStatQueues: {
				Replicate: utils.BoolPointer(false),
				Remote:    utils.BoolPointer(false),
			},
			utils.MetaResources: {
				Replicate: utils.BoolPointer(false),
				Remote:    utils.BoolPointer(false),
			},
			utils.MetaStatQueueProfiles: {
				Replicate: utils.BoolPointer(false),
				Remote:    utils.BoolPointer(false),
			},
			utils.MetaThresholds: {
				Replicate: utils.BoolPointer(false),
				Remote:    utils.BoolPointer(false),
			},
			utils.MetaThresholdProfiles: {
				Replicate: utils.BoolPointer(false),
				Remote:    utils.BoolPointer(false),
			},
			utils.MetaFilters: {
				Replicate: utils.BoolPointer(false),
				Remote:    utils.BoolPointer(false),
			},
			utils.MetaRouteProfiles: {
				Replicate: utils.BoolPointer(false),
				Remote:    utils.BoolPointer(false),
			},
			utils.MetaAttributeProfiles: {
				Replicate: utils.BoolPointer(false),
				Remote:    utils.BoolPointer(false),
			},
			utils.MetaDispatcherHosts: {
				Replicate: utils.BoolPointer(false),
				Remote:    utils.BoolPointer(false),
			},
			utils.MetaRateProfiles: {
				Replicate: utils.BoolPointer(false),
				Remote:    utils.BoolPointer(false),
			},
			utils.MetaActionProfiles: {
				Replicate: utils.BoolPointer(false),
				Remote:    utils.BoolPointer(false),
			},
			utils.MetaChargerProfiles: {
				Replicate: utils.BoolPointer(false),
				Remote:    utils.BoolPointer(false),
			},
			utils.MetaDispatcherProfiles: {
				Replicate: utils.BoolPointer(false),
				Remote:    utils.BoolPointer(false),
			},
			utils.MetaIndexes: {
				Replicate: utils.BoolPointer(false),
				Remote:    utils.BoolPointer(false),
			},
			utils.MetaLoadIDs: {
				Replicate: utils.BoolPointer(false),
				Remote:    utils.BoolPointer(false),
			},
		},
	}
	dfCgrJSONCfg, err := NewCgrJsonCfgFromBytes([]byte(CGRATES_CFG_JSON))
	if err != nil {
		t.Error(err)
	}
	if cfg, err := dfCgrJSONCfg.DbJsonCfg(DataDBJSON); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCfg, cfg) {
		t.Errorf("expecting: %s, \nreceived: %s", utils.ToJSON(eCfg), utils.ToJSON(cfg))
	}
}

func TestDfStorDBJsonCfg(t *testing.T) {
	eCfg := &DbJsonCfg{
		Db_type:               utils.StringPointer("*mysql"),
		Db_host:               utils.StringPointer("127.0.0.1"),
		Db_port:               utils.IntPointer(3306),
		Db_name:               utils.StringPointer("cgrates"),
		Db_user:               utils.StringPointer("cgrates"),
		Db_password:           utils.StringPointer(""),
		String_indexed_fields: &[]string{},
		Prefix_indexed_fields: &[]string{},
		Opts: map[string]interface{}{
			utils.MongoQueryTimeoutCfg:  "10s",
			utils.SQLMaxOpenConnsCfg:    100.,
			utils.SQLMaxIdleConnsCfg:    10.,
			utils.SQLConnMaxLifetimeCfg: 0.,
			utils.SSLModeCfg:            utils.PostgressSSLModeDisable,
			utils.MysqlLocation:         "Local",
		},
		Items: map[string]*ItemOptJson{
			utils.CacheTBLTPResources: {
				Replicate: utils.BoolPointer(false),
				Remote:    utils.BoolPointer(false),
			},
			utils.CacheTBLTPStats: {
				Replicate: utils.BoolPointer(false),
				Remote:    utils.BoolPointer(false),
			},
			utils.CacheTBLTPThresholds: {
				Replicate: utils.BoolPointer(false),
				Remote:    utils.BoolPointer(false),
			},
			utils.CacheTBLTPFilters: {
				Replicate: utils.BoolPointer(false),
				Remote:    utils.BoolPointer(false),
			},
			utils.CacheSessionCostsTBL: {
				Replicate: utils.BoolPointer(false),
				Remote:    utils.BoolPointer(false),
			},
			utils.CacheTBLTPRoutes: {
				Replicate: utils.BoolPointer(false),
				Remote:    utils.BoolPointer(false),
			},
			utils.CacheTBLTPAttributes: {
				Replicate: utils.BoolPointer(false),
				Remote:    utils.BoolPointer(false),
			},
			utils.CacheTBLTPChargers: {
				Replicate: utils.BoolPointer(false),
				Remote:    utils.BoolPointer(false),
			},
			utils.CacheTBLTPDispatchers: {
				Replicate: utils.BoolPointer(false),
				Remote:    utils.BoolPointer(false),
			},
			utils.CacheTBLTPRateProfiles: {
				Replicate: utils.BoolPointer(false),
				Remote:    utils.BoolPointer(false),
			},
			utils.CacheTBLTPDispatcherHosts: {
				Replicate: utils.BoolPointer(false),
				Remote:    utils.BoolPointer(false),
			},
			utils.CacheTBLTPActionProfiles: {
				Replicate: utils.BoolPointer(false),
				Remote:    utils.BoolPointer(false),
			},
			utils.CacheTBLTPAccounts: {
				Replicate: utils.BoolPointer(false),
				Remote:    utils.BoolPointer(false),
			},
			utils.CacheCDRsTBL: {
				Replicate: utils.BoolPointer(false),
				Remote:    utils.BoolPointer(false),
			},
			utils.CacheVersions: {
				Replicate: utils.BoolPointer(false),
				Remote:    utils.BoolPointer(false),
			},
		},
	}
	dfCgrJSONCfg, err := NewCgrJsonCfgFromBytes([]byte(CGRATES_CFG_JSON))
	if err != nil {
		t.Error(err)
	}
	if cfg, err := dfCgrJSONCfg.DbJsonCfg(StorDBJSON); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCfg, cfg) {
		t.Errorf("Expected : %+v,\n Received: %+v", utils.ToJSON(eCfg), utils.ToJSON(cfg))
	}
}

func TestDfCdrsJsonCfg(t *testing.T) {
	eCfg := &CdrsJsonCfg{
		Enabled:              utils.BoolPointer(false),
		Extra_fields:         &[]string{},
		Store_cdrs:           utils.BoolPointer(true),
		Session_cost_retries: utils.IntPointer(5),
		Chargers_conns:       &[]string{},
		Attributes_conns:     &[]string{},
		Thresholds_conns:     &[]string{},
		Stats_conns:          &[]string{},
		Online_cdr_exports:   &[]string{},
		Actions_conns:        &[]string{},
		Ees_conns:            &[]string{},
	}
	dfCgrJSONCfg, err := NewCgrJsonCfgFromBytes([]byte(CGRATES_CFG_JSON))
	if err != nil {
		t.Error(err)
	}
	if cfg, err := dfCgrJSONCfg.CdrsJsonCfg(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCfg, cfg) {
		t.Errorf("Received: %+v", cfg)
	}
}

func TestSmgJsonCfg(t *testing.T) {
	eCfg := &SessionSJsonCfg{
		Enabled:               utils.BoolPointer(false),
		Listen_bijson:         utils.StringPointer("127.0.0.1:2014"),
		Listen_bigob:          utils.StringPointer(""),
		Chargers_conns:        &[]string{},
		Cdrs_conns:            &[]string{},
		Resources_conns:       &[]string{},
		Thresholds_conns:      &[]string{},
		Stats_conns:           &[]string{},
		Routes_conns:          &[]string{},
		Attributes_conns:      &[]string{},
		Replication_conns:     &[]string{},
		Debit_interval:        utils.StringPointer("0s"),
		Store_session_costs:   utils.BoolPointer(false),
		Session_ttl:           utils.StringPointer("0s"),
		Session_indexes:       &[]string{},
		Client_protocol:       utils.Float64Pointer(1.0),
		Channel_sync_interval: utils.StringPointer("0"),
		Terminate_attempts:    utils.IntPointer(5),
		Alterable_fields:      &[]string{},
		Default_usage: map[string]string{
			utils.MetaAny:   "3h",
			utils.MetaVoice: "3h",
			utils.MetaData:  "1048576",
			utils.MetaSMS:   "1",
		},
		Stir: &STIRJsonCfg{
			Allowed_attest:      &[]string{utils.MetaAny},
			Payload_maxduration: utils.StringPointer("-1"),
			Default_attest:      utils.StringPointer("A"),
			Privatekey_path:     utils.StringPointer(""),
			Publickey_path:      utils.StringPointer(""),
		},
		Actions_conns: &[]string{},
	}
	dfCgrJSONCfg, err := NewCgrJsonCfgFromBytes([]byte(CGRATES_CFG_JSON))
	if err != nil {
		t.Error(err)
	}
	if cfg, err := dfCgrJSONCfg.SessionSJsonCfg(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCfg, cfg) {
		t.Error("Received: ", cfg)
	}
}

func TestFsAgentJsonCfg(t *testing.T) {
	eCfg := &FreeswitchAgentJsonCfg{
		Enabled:                utils.BoolPointer(false),
		Sessions_conns:         &[]string{rpcclient.BiRPCInternal},
		Subscribe_park:         utils.BoolPointer(true),
		Create_cdr:             utils.BoolPointer(false),
		Extra_fields:           &[]string{},
		Low_balance_ann_file:   utils.StringPointer(""),
		Empty_balance_context:  utils.StringPointer(""),
		Empty_balance_ann_file: utils.StringPointer(""),
		Max_wait_connection:    utils.StringPointer("2s"),
		Event_socket_conns: &[]*FsConnJsonCfg{
			{
				Address:    utils.StringPointer("127.0.0.1:8021"),
				Password:   utils.StringPointer("ClueCon"),
				Reconnects: utils.IntPointer(5),
				Alias:      utils.StringPointer(""),
			}},
	}
	dfCgrJSONCfg, err := NewCgrJsonCfgFromBytes([]byte(CGRATES_CFG_JSON))
	if err != nil {
		t.Error(err)
	}
	if cfg, err := dfCgrJSONCfg.FreeswitchAgentJsonCfg(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCfg, cfg) {
		t.Error("Received: ", cfg)
	}
}

func TestKamAgentJsonCfg(t *testing.T) {
	eCfg := &KamAgentJsonCfg{
		Enabled:        utils.BoolPointer(false),
		Sessions_conns: &[]string{rpcclient.BiRPCInternal},
		Create_cdr:     utils.BoolPointer(false),
		Evapi_conns: &[]*KamConnJsonCfg{
			{
				Address:    utils.StringPointer("127.0.0.1:8448"),
				Reconnects: utils.IntPointer(5),
			},
		},
		Timezone: utils.StringPointer(utils.EmptyString),
	}
	dfCgrJSONCfg, err := NewCgrJsonCfgFromBytes([]byte(CGRATES_CFG_JSON))
	if err != nil {
		t.Error(err)
	}
	if cfg, err := dfCgrJSONCfg.KamAgentJsonCfg(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCfg, cfg) {
		t.Errorf("Expecting: %s \n, received: %s: ",
			utils.ToJSON(eCfg), utils.ToJSON(cfg))
	}
}

func TestAsteriskAgentJsonCfg(t *testing.T) {
	eCfg := &AsteriskAgentJsonCfg{
		Enabled:        utils.BoolPointer(false),
		Sessions_conns: &[]string{rpcclient.BiRPCInternal},
		Create_cdr:     utils.BoolPointer(false),
		Asterisk_conns: &[]*AstConnJsonCfg{
			{
				Address:          utils.StringPointer("127.0.0.1:8088"),
				User:             utils.StringPointer("cgrates"),
				Password:         utils.StringPointer("CGRateS.org"),
				Connect_attempts: utils.IntPointer(3),
				Reconnects:       utils.IntPointer(5),
			},
		},
	}
	dfCgrJSONCfg, err := NewCgrJsonCfgFromBytes([]byte(CGRATES_CFG_JSON))
	if err != nil {
		t.Error(err)
	}
	if cfg, err := dfCgrJSONCfg.AsteriskAgentJsonCfg(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCfg, cfg) {
		t.Errorf("Expecting: %s, received: %s ", utils.ToJSON(eCfg), utils.ToJSON(cfg))
	}
}

func TestDiameterAgentJsonCfg(t *testing.T) {
	eCfg := &DiameterAgentJsonCfg{
		Enabled:              utils.BoolPointer(false),
		Listen:               utils.StringPointer("127.0.0.1:3868"),
		Listen_net:           utils.StringPointer(utils.TCP),
		Dictionaries_path:    utils.StringPointer("/usr/share/cgrates/diameter/dict/"),
		Sessions_conns:       &[]string{rpcclient.BiRPCInternal},
		Origin_host:          utils.StringPointer("CGR-DA"),
		Origin_realm:         utils.StringPointer("cgrates.org"),
		Vendor_id:            utils.IntPointer(0),
		Product_name:         utils.StringPointer("CGRateS"),
		Concurrent_requests:  utils.IntPointer(-1),
		Synced_conn_requests: utils.BoolPointer(false),
		Asr_template:         utils.StringPointer(""),
		Rar_template:         utils.StringPointer(""),
		Forced_disconnect:    utils.StringPointer(utils.MetaNone),
		Request_processors:   &[]*ReqProcessorJsnCfg{},
	}
	dfCgrJSONCfg, err := NewCgrJsonCfgFromBytes([]byte(CGRATES_CFG_JSON))
	if err != nil {
		t.Error(err)
	}
	if cfg, err := dfCgrJSONCfg.DiameterAgentJsonCfg(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCfg, cfg) {
		t.Errorf("expecting: %s, \n\nreceived: %s", utils.ToJSON(eCfg), utils.ToJSON(cfg))
	}
}

func TestRadiusAgentJsonCfg(t *testing.T) {
	eCfg := &RadiusAgentJsonCfg{
		Enabled:     utils.BoolPointer(false),
		Listen_net:  utils.StringPointer("udp"),
		Listen_auth: utils.StringPointer("127.0.0.1:1812"),
		Listen_acct: utils.StringPointer("127.0.0.1:1813"),
		Client_secrets: map[string]string{
			utils.MetaDefault: "CGRateS.org",
		},
		Client_dictionaries: map[string]string{
			utils.MetaDefault: "/usr/share/cgrates/radius/dict/",
		},
		Sessions_conns:     &[]string{utils.MetaInternal},
		Request_processors: &[]*ReqProcessorJsnCfg{},
	}
	dfCgrJSONCfg, err := NewCgrJsonCfgFromBytes([]byte(CGRATES_CFG_JSON))
	if err != nil {
		t.Error(err)
	}
	if cfg, err := dfCgrJSONCfg.RadiusAgentJsonCfg(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCfg, cfg) {
		rcv := *cfg.Request_processors
		t.Errorf("Received: %+v", rcv)
	}
}

func TestHttpAgentJsonCfg(t *testing.T) {
	eCfg := &[]*HttpAgentJsonCfg{}
	dfCgrJSONCfg, err := NewCgrJsonCfgFromBytes([]byte(CGRATES_CFG_JSON))
	if err != nil {
		t.Error(err)
	}
	if cfg, err := dfCgrJSONCfg.HttpAgentJsonCfg(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCfg, cfg) {
		t.Errorf("expecting: %+v, received: %+v", utils.ToJSON(eCfg), utils.ToJSON(cfg))
	}
}

func TestDNSAgentJsonCfg(t *testing.T) {
	eCfg := &DNSAgentJsonCfg{
		Enabled:            utils.BoolPointer(false),
		Listen_net:         utils.StringPointer("udp"),
		Listen:             utils.StringPointer("127.0.0.1:2053"),
		Sessions_conns:     &[]string{utils.ConcatenatedKey(utils.MetaInternal)},
		Timezone:           utils.StringPointer(""),
		Request_processors: &[]*ReqProcessorJsnCfg{},
	}
	dfCgrJSONCfg, err := NewCgrJsonCfgFromBytes([]byte(CGRATES_CFG_JSON))
	if err != nil {
		t.Error(err)
	}
	if cfg, err := dfCgrJSONCfg.DNSAgentJsonCfg(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCfg, cfg) {
		t.Errorf("expecting: %+v, received: %+v", utils.ToJSON(eCfg), utils.ToJSON(cfg))
	}
}

func TestDfAttributeServJsonCfg(t *testing.T) {
	eCfg := &AttributeSJsonCfg{
		Enabled:               utils.BoolPointer(false),
		Stats_conns:           &[]string{},
		Resources_conns:       &[]string{},
		Admins_conns:          &[]string{},
		Indexed_selects:       utils.BoolPointer(true),
		String_indexed_fields: nil,
		Prefix_indexed_fields: &[]string{},
		Suffix_indexed_fields: &[]string{},
		Nested_fields:         utils.BoolPointer(false),
		Process_runs:          utils.IntPointer(1),
	}
	dfCgrJSONCfg, err := NewCgrJsonCfgFromBytes([]byte(CGRATES_CFG_JSON))
	if err != nil {
		t.Error(err)
	}
	if cfg, err := dfCgrJSONCfg.AttributeServJsonCfg(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCfg, cfg) {
		t.Error("Received: ", utils.ToJSON(cfg))
	}
}

func TestDfChargerServJsonCfg(t *testing.T) {
	eCfg := &ChargerSJsonCfg{
		Enabled:               utils.BoolPointer(false),
		Indexed_selects:       utils.BoolPointer(true),
		Attributes_conns:      &[]string{},
		String_indexed_fields: nil,
		Prefix_indexed_fields: &[]string{},
		Suffix_indexed_fields: &[]string{},
		Nested_fields:         utils.BoolPointer(false),
	}
	dfCgrJSONCfg, err := NewCgrJsonCfgFromBytes([]byte(CGRATES_CFG_JSON))
	if err != nil {
		t.Error(err)
	}
	if cfg, err := dfCgrJSONCfg.ChargerServJsonCfg(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCfg, cfg) {
		t.Error("Received: ", utils.ToJSON(cfg))
	}
}

func TestDfFilterSJsonCfg(t *testing.T) {
	eCfg := &FilterSJsonCfg{
		Stats_conns:     &[]string{},
		Resources_conns: &[]string{},
		Admins_conns:    &[]string{},
	}
	dfCgrJSONCfg, err := NewCgrJsonCfgFromBytes([]byte(CGRATES_CFG_JSON))
	if err != nil {
		t.Error(err)
	}
	if cfg, err := dfCgrJSONCfg.FilterSJsonCfg(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCfg, cfg) {
		t.Errorf("Expected: %s, received: %s", utils.ToJSON(eCfg), utils.ToJSON(cfg))
	}
}

func TestDfResourceLimiterSJsonCfg(t *testing.T) {
	eCfg := &ResourceSJsonCfg{
		Enabled:               utils.BoolPointer(false),
		Indexed_selects:       utils.BoolPointer(true),
		Thresholds_conns:      &[]string{},
		Store_interval:        utils.StringPointer(""),
		String_indexed_fields: nil,
		Prefix_indexed_fields: &[]string{},
		Suffix_indexed_fields: &[]string{},
		Nested_fields:         utils.BoolPointer(false),
	}
	dfCgrJSONCfg, err := NewCgrJsonCfgFromBytes([]byte(CGRATES_CFG_JSON))
	if err != nil {
		t.Error(err)
	}
	if cfg, err := dfCgrJSONCfg.ResourceSJsonCfg(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCfg, cfg) {
		t.Errorf("Expected: %s, received: %s", utils.ToJSON(eCfg), utils.ToJSON(cfg))
	}
}

func TestDfStatServiceJsonCfg(t *testing.T) {
	eCfg := &StatServJsonCfg{
		Enabled:                  utils.BoolPointer(false),
		Indexed_selects:          utils.BoolPointer(true),
		Store_interval:           utils.StringPointer(""),
		Store_uncompressed_limit: utils.IntPointer(0),
		Thresholds_conns:         &[]string{},
		String_indexed_fields:    nil,
		Prefix_indexed_fields:    &[]string{},
		Suffix_indexed_fields:    &[]string{},
		Nested_fields:            utils.BoolPointer(false),
	}
	dfCgrJSONCfg, err := NewCgrJsonCfgFromBytes([]byte(CGRATES_CFG_JSON))
	if err != nil {
		t.Error(err)
	}
	if cfg, err := dfCgrJSONCfg.StatSJsonCfg(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCfg, cfg) {
		t.Error("Received: ", utils.ToJSON(cfg))
	}
}

func TestDfThresholdSJsonCfg(t *testing.T) {
	eCfg := &ThresholdSJsonCfg{
		Enabled:               utils.BoolPointer(false),
		Indexed_selects:       utils.BoolPointer(true),
		Store_interval:        utils.StringPointer(""),
		String_indexed_fields: nil,
		Prefix_indexed_fields: &[]string{},
		Suffix_indexed_fields: &[]string{},
		Nested_fields:         utils.BoolPointer(false),
		Actions_conns:         &[]string{},
	}
	dfCgrJSONCfg, err := NewCgrJsonCfgFromBytes([]byte(CGRATES_CFG_JSON))
	if err != nil {
		t.Error(err)
	}
	if cfg, err := dfCgrJSONCfg.ThresholdSJsonCfg(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCfg, cfg) {
		t.Errorf("expecting: %+v, received: %+v", utils.ToJSON(eCfg), utils.ToJSON(cfg))
	}
}

func TestDfRouteSJsonCfg(t *testing.T) {
	eCfg := &RouteSJsonCfg{
		Enabled:               utils.BoolPointer(false),
		Indexed_selects:       utils.BoolPointer(true),
		String_indexed_fields: nil,
		Prefix_indexed_fields: &[]string{},
		Suffix_indexed_fields: &[]string{},
		Attributes_conns:      &[]string{},
		Resources_conns:       &[]string{},
		Stats_conns:           &[]string{},
		Default_ratio:         utils.IntPointer(1),
		Nested_fields:         utils.BoolPointer(false),
	}
	dfCgrJSONCfg, err := NewCgrJsonCfgFromBytes([]byte(CGRATES_CFG_JSON))
	if err != nil {
		t.Error(err)
	}
	if cfg, err := dfCgrJSONCfg.RouteSJsonCfg(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCfg, cfg) {
		t.Errorf("expecting: %+v, received: %+v", utils.ToJSON(eCfg), utils.ToJSON(cfg))
	}
}

func TestDfLoaderJsonCfg(t *testing.T) {
	eCfg := []*LoaderJsonCfg{
		{
			ID:              utils.StringPointer(utils.MetaDefault),
			Enabled:         utils.BoolPointer(false),
			Tenant:          utils.StringPointer(""),
			Dry_run:         utils.BoolPointer(false),
			Run_delay:       utils.StringPointer("0"),
			Lock_filename:   utils.StringPointer(".cgr.lck"),
			Caches_conns:    &[]string{utils.MetaInternal},
			Field_separator: utils.StringPointer(","),
			Tp_in_dir:       utils.StringPointer("/var/spool/cgrates/loader/in"),
			Tp_out_dir:      utils.StringPointer("/var/spool/cgrates/loader/out"),
			Data: &[]*LoaderJsonDataType{
				{
					Type:      utils.StringPointer(utils.MetaFilters),
					File_name: utils.StringPointer(utils.FiltersCsv),
					Fields: &[]*FcTemplateJsonCfg{
						{Tag: utils.StringPointer(utils.Tenant),
							Path:      utils.StringPointer(utils.Tenant),
							Type:      utils.StringPointer(utils.MetaVariable),
							Value:     utils.StringPointer("~*req.0"),
							Mandatory: utils.BoolPointer(true)},
						{Tag: utils.StringPointer(utils.ID),
							Path:      utils.StringPointer(utils.ID),
							Type:      utils.StringPointer(utils.MetaVariable),
							Value:     utils.StringPointer("~*req.1"),
							Mandatory: utils.BoolPointer(true)},
						{Tag: utils.StringPointer("Type"),
							Path:  utils.StringPointer("Type"),
							Type:  utils.StringPointer(utils.MetaVariable),
							Value: utils.StringPointer("~*req.2")},
						{Tag: utils.StringPointer("Element"),
							Path:  utils.StringPointer("Element"),
							Type:  utils.StringPointer(utils.MetaVariable),
							Value: utils.StringPointer("~*req.3")},
						{Tag: utils.StringPointer("Values"),
							Path:  utils.StringPointer("Values"),
							Type:  utils.StringPointer(utils.MetaVariable),
							Value: utils.StringPointer("~*req.4")},
					},
				},
				{
					Type:      utils.StringPointer(utils.MetaAttributes),
					File_name: utils.StringPointer(utils.AttributesCsv),
					Fields: &[]*FcTemplateJsonCfg{
						{Tag: utils.StringPointer("TenantID"),
							Path:      utils.StringPointer(utils.Tenant),
							Type:      utils.StringPointer(utils.MetaVariable),
							Value:     utils.StringPointer("~*req.0"),
							Mandatory: utils.BoolPointer(true)},
						{Tag: utils.StringPointer("ProfileID"),
							Path:      utils.StringPointer(utils.ID),
							Type:      utils.StringPointer(utils.MetaVariable),
							Value:     utils.StringPointer("~*req.1"),
							Mandatory: utils.BoolPointer(true)},
						{Tag: utils.StringPointer("FilterIDs"),
							Path:  utils.StringPointer(utils.FilterIDs),
							Type:  utils.StringPointer(utils.MetaVariable),
							Value: utils.StringPointer("~*req.2")},
						{Tag: utils.StringPointer("Weight"),
							Path:  utils.StringPointer(utils.Weight),
							Type:  utils.StringPointer(utils.MetaVariable),
							Value: utils.StringPointer("~*req.3")},
						{Tag: utils.StringPointer("AttributeFilterIDs"),
							Path:  utils.StringPointer("AttributeFilterIDs"),
							Type:  utils.StringPointer(utils.MetaVariable),
							Value: utils.StringPointer("~*req.4")},
						{Tag: utils.StringPointer("Path"),
							Path:  utils.StringPointer(utils.Path),
							Type:  utils.StringPointer(utils.MetaVariable),
							Value: utils.StringPointer("~*req.5")},
						{Tag: utils.StringPointer("Type"),
							Path:  utils.StringPointer("Type"),
							Type:  utils.StringPointer(utils.MetaVariable),
							Value: utils.StringPointer("~*req.6")},
						{Tag: utils.StringPointer("Value"),
							Path:  utils.StringPointer("Value"),
							Type:  utils.StringPointer(utils.MetaVariable),
							Value: utils.StringPointer("~*req.7")},
						{Tag: utils.StringPointer("Blocker"),
							Path:  utils.StringPointer("Blocker"),
							Type:  utils.StringPointer(utils.MetaVariable),
							Value: utils.StringPointer("~*req.8")},
					},
				},
				{
					Type:      utils.StringPointer(utils.MetaResources),
					File_name: utils.StringPointer(utils.ResourcesCsv),
					Fields: &[]*FcTemplateJsonCfg{
						{Tag: utils.StringPointer(utils.Tenant),
							Path:      utils.StringPointer(utils.Tenant),
							Type:      utils.StringPointer(utils.MetaVariable),
							Value:     utils.StringPointer("~*req.0"),
							Mandatory: utils.BoolPointer(true)},
						{Tag: utils.StringPointer(utils.ID),
							Path:      utils.StringPointer(utils.ID),
							Type:      utils.StringPointer(utils.MetaVariable),
							Value:     utils.StringPointer("~*req.1"),
							Mandatory: utils.BoolPointer(true)},
						{Tag: utils.StringPointer("FilterIDs"),
							Path:  utils.StringPointer("FilterIDs"),
							Type:  utils.StringPointer(utils.MetaVariable),
							Value: utils.StringPointer("~*req.2")},
						{Tag: utils.StringPointer("Weight"),
							Path:  utils.StringPointer("Weight"),
							Type:  utils.StringPointer(utils.MetaVariable),
							Value: utils.StringPointer("~*req.3")},
						{Tag: utils.StringPointer("TTL"),
							Path:  utils.StringPointer("UsageTTL"),
							Type:  utils.StringPointer(utils.MetaVariable),
							Value: utils.StringPointer("~*req.4")},
						{Tag: utils.StringPointer("Limit"),
							Path:  utils.StringPointer("Limit"),
							Type:  utils.StringPointer(utils.MetaVariable),
							Value: utils.StringPointer("~*req.5")},
						{Tag: utils.StringPointer("AllocationMessage"),
							Path:  utils.StringPointer("AllocationMessage"),
							Type:  utils.StringPointer(utils.MetaVariable),
							Value: utils.StringPointer("~*req.6")},
						{Tag: utils.StringPointer("Blocker"),
							Path:  utils.StringPointer("Blocker"),
							Type:  utils.StringPointer(utils.MetaVariable),
							Value: utils.StringPointer("~*req.7")},
						{Tag: utils.StringPointer("Stored"),
							Path:  utils.StringPointer("Stored"),
							Type:  utils.StringPointer(utils.MetaVariable),
							Value: utils.StringPointer("~*req.8")},
						{Tag: utils.StringPointer("ThresholdIDs"),
							Path:  utils.StringPointer("ThresholdIDs"),
							Type:  utils.StringPointer(utils.MetaVariable),
							Value: utils.StringPointer("~*req.9")},
					},
				},
				{
					Type:      utils.StringPointer(utils.MetaStats),
					File_name: utils.StringPointer(utils.StatsCsv),
					Fields: &[]*FcTemplateJsonCfg{
						{Tag: utils.StringPointer(utils.Tenant),
							Path:      utils.StringPointer(utils.Tenant),
							Type:      utils.StringPointer(utils.MetaVariable),
							Value:     utils.StringPointer("~*req.0"),
							Mandatory: utils.BoolPointer(true)},
						{Tag: utils.StringPointer(utils.ID),
							Path:      utils.StringPointer(utils.ID),
							Type:      utils.StringPointer(utils.MetaVariable),
							Value:     utils.StringPointer("~*req.1"),
							Mandatory: utils.BoolPointer(true)},
						{Tag: utils.StringPointer("FilterIDs"),
							Path:  utils.StringPointer("FilterIDs"),
							Type:  utils.StringPointer(utils.MetaVariable),
							Value: utils.StringPointer("~*req.2")},
						{Tag: utils.StringPointer("Weight"),
							Path:  utils.StringPointer("Weight"),
							Type:  utils.StringPointer(utils.MetaVariable),
							Value: utils.StringPointer("~*req.3")},
						{Tag: utils.StringPointer("QueueLength"),
							Path:  utils.StringPointer("QueueLength"),
							Type:  utils.StringPointer(utils.MetaVariable),
							Value: utils.StringPointer("~*req.4")},
						{Tag: utils.StringPointer("TTL"),
							Path:  utils.StringPointer("TTL"),
							Type:  utils.StringPointer(utils.MetaVariable),
							Value: utils.StringPointer("~*req.5")},
						{Tag: utils.StringPointer("MinItems"),
							Path:  utils.StringPointer("MinItems"),
							Type:  utils.StringPointer(utils.MetaVariable),
							Value: utils.StringPointer("~*req.6")},
						{Tag: utils.StringPointer("MetricIDs"),
							Path:  utils.StringPointer("MetricIDs"),
							Type:  utils.StringPointer(utils.MetaVariable),
							Value: utils.StringPointer("~*req.7")},
						{Tag: utils.StringPointer("MetricFilterIDs"),
							Path:  utils.StringPointer("MetricFilterIDs"),
							Type:  utils.StringPointer(utils.MetaVariable),
							Value: utils.StringPointer("~*req.8")},
						{Tag: utils.StringPointer("Blocker"),
							Path:  utils.StringPointer("Blocker"),
							Type:  utils.StringPointer(utils.MetaVariable),
							Value: utils.StringPointer("~*req.9")},
						{Tag: utils.StringPointer("Stored"),
							Path:  utils.StringPointer("Stored"),
							Type:  utils.StringPointer(utils.MetaVariable),
							Value: utils.StringPointer("~*req.10")},
						{Tag: utils.StringPointer("ThresholdIDs"),
							Path:  utils.StringPointer("ThresholdIDs"),
							Type:  utils.StringPointer(utils.MetaVariable),
							Value: utils.StringPointer("~*req.11")},
					},
				},
				{
					Type:      utils.StringPointer(utils.MetaThresholds),
					File_name: utils.StringPointer(utils.ThresholdsCsv),
					Fields: &[]*FcTemplateJsonCfg{
						{Tag: utils.StringPointer(utils.Tenant),
							Path:      utils.StringPointer(utils.Tenant),
							Type:      utils.StringPointer(utils.MetaVariable),
							Value:     utils.StringPointer("~*req.0"),
							Mandatory: utils.BoolPointer(true)},
						{Tag: utils.StringPointer(utils.ID),
							Path:      utils.StringPointer(utils.ID),
							Type:      utils.StringPointer(utils.MetaVariable),
							Value:     utils.StringPointer("~*req.1"),
							Mandatory: utils.BoolPointer(true)},
						{Tag: utils.StringPointer("FilterIDs"),
							Path:  utils.StringPointer("FilterIDs"),
							Type:  utils.StringPointer(utils.MetaVariable),
							Value: utils.StringPointer("~*req.2")},
						{Tag: utils.StringPointer("Weight"),
							Path:  utils.StringPointer("Weight"),
							Type:  utils.StringPointer(utils.MetaVariable),
							Value: utils.StringPointer("~*req.3")},
						{Tag: utils.StringPointer("MaxHits"),
							Path:  utils.StringPointer("MaxHits"),
							Type:  utils.StringPointer(utils.MetaVariable),
							Value: utils.StringPointer("~*req.4")},
						{Tag: utils.StringPointer("MinHits"),
							Path:  utils.StringPointer("MinHits"),
							Type:  utils.StringPointer(utils.MetaVariable),
							Value: utils.StringPointer("~*req.5")},
						{Tag: utils.StringPointer("MinSleep"),
							Path:  utils.StringPointer("MinSleep"),
							Type:  utils.StringPointer(utils.MetaVariable),
							Value: utils.StringPointer("~*req.6")},
						{Tag: utils.StringPointer("Blocker"),
							Path:  utils.StringPointer("Blocker"),
							Type:  utils.StringPointer(utils.MetaVariable),
							Value: utils.StringPointer("~*req.7")},
						{Tag: utils.StringPointer("ActionProfileIDs"),
							Path:  utils.StringPointer("ActionProfileIDs"),
							Type:  utils.StringPointer(utils.MetaVariable),
							Value: utils.StringPointer("~*req.8")},
						{Tag: utils.StringPointer("Async"),
							Path:  utils.StringPointer("Async"),
							Type:  utils.StringPointer(utils.MetaVariable),
							Value: utils.StringPointer("~*req.9")},
					},
				},
				{
					Type:      utils.StringPointer(utils.MetaRoutes),
					File_name: utils.StringPointer(utils.RoutesCsv),
					Fields: &[]*FcTemplateJsonCfg{
						{Tag: utils.StringPointer(utils.Tenant),
							Path:      utils.StringPointer(utils.Tenant),
							Type:      utils.StringPointer(utils.MetaVariable),
							Value:     utils.StringPointer("~*req.0"),
							Mandatory: utils.BoolPointer(true)},
						{Tag: utils.StringPointer(utils.ID),
							Path:      utils.StringPointer(utils.ID),
							Type:      utils.StringPointer(utils.MetaVariable),
							Value:     utils.StringPointer("~*req.1"),
							Mandatory: utils.BoolPointer(true)},
						{Tag: utils.StringPointer("FilterIDs"),
							Path:  utils.StringPointer("FilterIDs"),
							Type:  utils.StringPointer(utils.MetaVariable),
							Value: utils.StringPointer("~*req.2")},
						{Tag: utils.StringPointer("Weight"),
							Path:  utils.StringPointer("Weight"),
							Type:  utils.StringPointer(utils.MetaVariable),
							Value: utils.StringPointer("~*req.3")},
						{Tag: utils.StringPointer("Sorting"),
							Path:  utils.StringPointer("Sorting"),
							Type:  utils.StringPointer(utils.MetaVariable),
							Value: utils.StringPointer("~*req.4")},
						{Tag: utils.StringPointer("SortingParameters"),
							Path:  utils.StringPointer("SortingParameters"),
							Type:  utils.StringPointer(utils.MetaVariable),
							Value: utils.StringPointer("~*req.5")},
						{Tag: utils.StringPointer("RouteID"),
							Path:  utils.StringPointer("RouteID"),
							Type:  utils.StringPointer(utils.MetaVariable),
							Value: utils.StringPointer("~*req.6")},
						{Tag: utils.StringPointer("RouteFilterIDs"),
							Path:  utils.StringPointer("RouteFilterIDs"),
							Type:  utils.StringPointer(utils.MetaVariable),
							Value: utils.StringPointer("~*req.7")},
						{Tag: utils.StringPointer("RouteAccountIDs"),
							Path:  utils.StringPointer("RouteAccountIDs"),
							Type:  utils.StringPointer(utils.MetaVariable),
							Value: utils.StringPointer("~*req.8")},
						{Tag: utils.StringPointer("RouteRatingPlanIDs"),
							Path:  utils.StringPointer("RouteRatingPlanIDs"),
							Type:  utils.StringPointer(utils.MetaVariable),
							Value: utils.StringPointer("~*req.9")},
						{Tag: utils.StringPointer("RouteResourceIDs"),
							Path:  utils.StringPointer("RouteResourceIDs"),
							Type:  utils.StringPointer(utils.MetaVariable),
							Value: utils.StringPointer("~*req.10")},
						{Tag: utils.StringPointer("RouteStatIDs"),
							Path:  utils.StringPointer("RouteStatIDs"),
							Type:  utils.StringPointer(utils.MetaVariable),
							Value: utils.StringPointer("~*req.11")},
						{Tag: utils.StringPointer("RouteWeight"),
							Path:  utils.StringPointer("RouteWeight"),
							Type:  utils.StringPointer(utils.MetaVariable),
							Value: utils.StringPointer("~*req.12")},
						{Tag: utils.StringPointer("RouteBlocker"),
							Path:  utils.StringPointer("RouteBlocker"),
							Type:  utils.StringPointer(utils.MetaVariable),
							Value: utils.StringPointer("~*req.13")},
						{Tag: utils.StringPointer("RouteParameters"),
							Path:  utils.StringPointer("RouteParameters"),
							Type:  utils.StringPointer(utils.MetaVariable),
							Value: utils.StringPointer("~*req.14")},
					},
				},
				{
					Type:      utils.StringPointer(utils.MetaChargers),
					File_name: utils.StringPointer(utils.ChargersCsv),
					Fields: &[]*FcTemplateJsonCfg{
						{Tag: utils.StringPointer(utils.Tenant),
							Path:      utils.StringPointer(utils.Tenant),
							Type:      utils.StringPointer(utils.MetaVariable),
							Value:     utils.StringPointer("~*req.0"),
							Mandatory: utils.BoolPointer(true)},
						{Tag: utils.StringPointer(utils.ID),
							Path:      utils.StringPointer(utils.ID),
							Type:      utils.StringPointer(utils.MetaVariable),
							Value:     utils.StringPointer("~*req.1"),
							Mandatory: utils.BoolPointer(true)},
						{Tag: utils.StringPointer("FilterIDs"),
							Path:  utils.StringPointer("FilterIDs"),
							Type:  utils.StringPointer(utils.MetaVariable),
							Value: utils.StringPointer("~*req.2")},
						{Tag: utils.StringPointer("Weight"),
							Path:  utils.StringPointer("Weight"),
							Type:  utils.StringPointer(utils.MetaVariable),
							Value: utils.StringPointer("~*req.3")},
						{Tag: utils.StringPointer("RunID"),
							Path:  utils.StringPointer("RunID"),
							Type:  utils.StringPointer(utils.MetaVariable),
							Value: utils.StringPointer("~*req.4")},
						{Tag: utils.StringPointer("AttributeIDs"),
							Path:  utils.StringPointer("AttributeIDs"),
							Type:  utils.StringPointer(utils.MetaVariable),
							Value: utils.StringPointer("~*req.5")},
					},
				},
				{
					Type:      utils.StringPointer(utils.MetaDispatchers),
					File_name: utils.StringPointer(utils.DispatcherProfilesCsv),
					Fields: &[]*FcTemplateJsonCfg{
						{Tag: utils.StringPointer(utils.Tenant),
							Path:      utils.StringPointer(utils.Tenant),
							Type:      utils.StringPointer(utils.MetaVariable),
							Value:     utils.StringPointer("~*req.0"),
							Mandatory: utils.BoolPointer(true)},
						{Tag: utils.StringPointer(utils.ID),
							Path:      utils.StringPointer(utils.ID),
							Type:      utils.StringPointer(utils.MetaVariable),
							Value:     utils.StringPointer("~*req.1"),
							Mandatory: utils.BoolPointer(true)},
						{Tag: utils.StringPointer("FilterIDs"),
							Path:  utils.StringPointer("FilterIDs"),
							Type:  utils.StringPointer(utils.MetaVariable),
							Value: utils.StringPointer("~*req.2")},
						{Tag: utils.StringPointer("Weight"),
							Path:  utils.StringPointer("Weight"),
							Type:  utils.StringPointer(utils.MetaVariable),
							Value: utils.StringPointer("~*req.3")},
						{Tag: utils.StringPointer("Strategy"),
							Path:  utils.StringPointer("Strategy"),
							Type:  utils.StringPointer(utils.MetaVariable),
							Value: utils.StringPointer("~*req.4")},
						{Tag: utils.StringPointer("StrategyParameters"),
							Path:  utils.StringPointer("StrategyParameters"),
							Type:  utils.StringPointer(utils.MetaVariable),
							Value: utils.StringPointer("~*req.5")},
						{Tag: utils.StringPointer("ConnID"),
							Path:  utils.StringPointer("ConnID"),
							Type:  utils.StringPointer(utils.MetaVariable),
							Value: utils.StringPointer("~*req.6")},
						{Tag: utils.StringPointer("ConnFilterIDs"),
							Path:  utils.StringPointer("ConnFilterIDs"),
							Type:  utils.StringPointer(utils.MetaVariable),
							Value: utils.StringPointer("~*req.7")},
						{Tag: utils.StringPointer("ConnWeight"),
							Path:  utils.StringPointer("ConnWeight"),
							Type:  utils.StringPointer(utils.MetaVariable),
							Value: utils.StringPointer("~*req.8")},
						{Tag: utils.StringPointer("ConnBlocker"),
							Path:  utils.StringPointer("ConnBlocker"),
							Type:  utils.StringPointer(utils.MetaVariable),
							Value: utils.StringPointer("~*req.9")},
						{Tag: utils.StringPointer("ConnParameters"),
							Path:  utils.StringPointer("ConnParameters"),
							Type:  utils.StringPointer(utils.MetaVariable),
							Value: utils.StringPointer("~*req.10")},
					},
				},
				{
					Type:      utils.StringPointer(utils.MetaDispatcherHosts),
					File_name: utils.StringPointer(utils.DispatcherHostsCsv),
					Fields: &[]*FcTemplateJsonCfg{
						{Tag: utils.StringPointer(utils.Tenant),
							Path:      utils.StringPointer(utils.Tenant),
							Type:      utils.StringPointer(utils.MetaVariable),
							Value:     utils.StringPointer("~*req.0"),
							Mandatory: utils.BoolPointer(true)},
						{Tag: utils.StringPointer(utils.ID),
							Path:      utils.StringPointer(utils.ID),
							Type:      utils.StringPointer(utils.MetaVariable),
							Value:     utils.StringPointer("~*req.1"),
							Mandatory: utils.BoolPointer(true)},
						{Tag: utils.StringPointer("Address"),
							Path:  utils.StringPointer("Address"),
							Type:  utils.StringPointer(utils.MetaVariable),
							Value: utils.StringPointer("~*req.2")},
						{Tag: utils.StringPointer("Transport"),
							Path:  utils.StringPointer("Transport"),
							Type:  utils.StringPointer(utils.MetaVariable),
							Value: utils.StringPointer("~*req.3")},
						{Tag: utils.StringPointer("TLS"),
							Path:  utils.StringPointer("TLS"),
							Type:  utils.StringPointer(utils.MetaVariable),
							Value: utils.StringPointer("~*req.4")},
					},
				},
				{
					Type:      utils.StringPointer(utils.MetaRateProfiles),
					File_name: utils.StringPointer(utils.RateProfilesCsv),
					Fields: &[]*FcTemplateJsonCfg{
						{Tag: utils.StringPointer(utils.Tenant),
							Path:      utils.StringPointer(utils.Tenant),
							Type:      utils.StringPointer(utils.MetaVariable),
							Value:     utils.StringPointer("~*req.0"),
							Mandatory: utils.BoolPointer(true)},
						{Tag: utils.StringPointer(utils.ID),
							Path:      utils.StringPointer(utils.ID),
							Type:      utils.StringPointer(utils.MetaVariable),
							Value:     utils.StringPointer("~*req.1"),
							Mandatory: utils.BoolPointer(true)},
						{Tag: utils.StringPointer("FilterIDs"),
							Path:  utils.StringPointer("FilterIDs"),
							Type:  utils.StringPointer(utils.MetaVariable),
							Value: utils.StringPointer("~*req.2")},
						{Tag: utils.StringPointer("Weight"),
							Path:  utils.StringPointer("Weight"),
							Type:  utils.StringPointer(utils.MetaVariable),
							Value: utils.StringPointer("~*req.3")},
						{Tag: utils.StringPointer("MinCost"),
							Path:  utils.StringPointer("MinCost"),
							Type:  utils.StringPointer(utils.MetaVariable),
							Value: utils.StringPointer("~*req.4")},
						{Tag: utils.StringPointer("MaxCost"),
							Path:  utils.StringPointer("MaxCost"),
							Type:  utils.StringPointer(utils.MetaVariable),
							Value: utils.StringPointer("~*req.5")},
						{Tag: utils.StringPointer("MaxCostStrategy"),
							Path:  utils.StringPointer("MaxCostStrategy"),
							Type:  utils.StringPointer(utils.MetaVariable),
							Value: utils.StringPointer("~*req.6")},
						{Tag: utils.StringPointer("RateID"),
							Path:  utils.StringPointer("RateID"),
							Type:  utils.StringPointer(utils.MetaVariable),
							Value: utils.StringPointer("~*req.7")},
						{Tag: utils.StringPointer("RateFilterIDs"),
							Path:  utils.StringPointer("RateFilterIDs"),
							Type:  utils.StringPointer(utils.MetaVariable),
							Value: utils.StringPointer("~*req.8")},
						{Tag: utils.StringPointer("RateActivationTimes"),
							Path:  utils.StringPointer("RateActivationTimes"),
							Type:  utils.StringPointer(utils.MetaVariable),
							Value: utils.StringPointer("~*req.9")},
						{Tag: utils.StringPointer("RateWeight"),
							Path:  utils.StringPointer("RateWeight"),
							Type:  utils.StringPointer(utils.MetaVariable),
							Value: utils.StringPointer("~*req.10")},
						{Tag: utils.StringPointer("RateBlocker"),
							Path:  utils.StringPointer("RateBlocker"),
							Type:  utils.StringPointer(utils.MetaVariable),
							Value: utils.StringPointer("~*req.11")},
						{Tag: utils.StringPointer("RateIntervalStart"),
							Path:  utils.StringPointer("RateIntervalStart"),
							Type:  utils.StringPointer(utils.MetaVariable),
							Value: utils.StringPointer("~*req.12")},
						{Tag: utils.StringPointer("RateFixedFee"),
							Path:  utils.StringPointer("RateFixedFee"),
							Type:  utils.StringPointer(utils.MetaVariable),
							Value: utils.StringPointer("~*req.13")},
						{Tag: utils.StringPointer("RateRecurrentFee"),
							Path:  utils.StringPointer("RateRecurrentFee"),
							Type:  utils.StringPointer(utils.MetaVariable),
							Value: utils.StringPointer("~*req.14")},
						{Tag: utils.StringPointer("RateUnit"),
							Path:  utils.StringPointer("RateUnit"),
							Type:  utils.StringPointer(utils.MetaVariable),
							Value: utils.StringPointer("~*req.15")},
						{Tag: utils.StringPointer("RateIncrement"),
							Path:  utils.StringPointer("RateIncrement"),
							Type:  utils.StringPointer(utils.MetaVariable),
							Value: utils.StringPointer("~*req.16")},
					},
				},
				{
					Type:      utils.StringPointer(utils.MetaActionProfiles),
					File_name: utils.StringPointer(utils.ActionProfilesCsv),
					Fields: &[]*FcTemplateJsonCfg{
						{Tag: utils.StringPointer(utils.Tenant),
							Path:      utils.StringPointer(utils.Tenant),
							Type:      utils.StringPointer(utils.MetaVariable),
							Value:     utils.StringPointer("~*req.0"),
							Mandatory: utils.BoolPointer(true)},
						{Tag: utils.StringPointer(utils.ID),
							Path:      utils.StringPointer(utils.ID),
							Type:      utils.StringPointer(utils.MetaVariable),
							Value:     utils.StringPointer("~*req.1"),
							Mandatory: utils.BoolPointer(true)},
						{Tag: utils.StringPointer("FilterIDs"),
							Path:  utils.StringPointer("FilterIDs"),
							Type:  utils.StringPointer(utils.MetaVariable),
							Value: utils.StringPointer("~*req.2")},
						{Tag: utils.StringPointer("Weight"),
							Path:  utils.StringPointer("Weight"),
							Type:  utils.StringPointer(utils.MetaVariable),
							Value: utils.StringPointer("~*req.3")},
						{Tag: utils.StringPointer("Schedule"),
							Path:  utils.StringPointer("Schedule"),
							Type:  utils.StringPointer(utils.MetaVariable),
							Value: utils.StringPointer("~*req.4")},
						{Tag: utils.StringPointer("TargetType"),
							Path:  utils.StringPointer("TargetType"),
							Type:  utils.StringPointer(utils.MetaVariable),
							Value: utils.StringPointer("~*req.5")},
						{Tag: utils.StringPointer("TargetIDs"),
							Path:  utils.StringPointer("TargetIDs"),
							Type:  utils.StringPointer(utils.MetaVariable),
							Value: utils.StringPointer("~*req.6")},
						{Tag: utils.StringPointer("ActionID"),
							Path:  utils.StringPointer("ActionID"),
							Type:  utils.StringPointer(utils.MetaVariable),
							Value: utils.StringPointer("~*req.7")},
						{Tag: utils.StringPointer("ActionFilterIDs"),
							Path:  utils.StringPointer("ActionFilterIDs"),
							Type:  utils.StringPointer(utils.MetaVariable),
							Value: utils.StringPointer("~*req.8")},
						{Tag: utils.StringPointer("ActionBlocker"),
							Path:  utils.StringPointer("ActionBlocker"),
							Type:  utils.StringPointer(utils.MetaVariable),
							Value: utils.StringPointer("~*req.9")},
						{Tag: utils.StringPointer("ActionTTL"),
							Path:  utils.StringPointer("ActionTTL"),
							Type:  utils.StringPointer(utils.MetaVariable),
							Value: utils.StringPointer("~*req.10")},
						{Tag: utils.StringPointer("ActionType"),
							Path:  utils.StringPointer("ActionType"),
							Type:  utils.StringPointer(utils.MetaVariable),
							Value: utils.StringPointer("~*req.11")},
						{Tag: utils.StringPointer("ActionOpts"),
							Path:  utils.StringPointer("ActionOpts"),
							Type:  utils.StringPointer(utils.MetaVariable),
							Value: utils.StringPointer("~*req.12")},
						{Tag: utils.StringPointer("ActionPath"),
							Path:  utils.StringPointer("ActionPath"),
							Type:  utils.StringPointer(utils.MetaVariable),
							Value: utils.StringPointer("~*req.13")},
						{Tag: utils.StringPointer("ActionValue"),
							Path:  utils.StringPointer("ActionValue"),
							Type:  utils.StringPointer(utils.MetaVariable),
							Value: utils.StringPointer("~*req.14")},
					},
				},
				{
					Type:      utils.StringPointer(utils.MetaAccounts),
					File_name: utils.StringPointer(utils.AccountsCsv),
					Fields: &[]*FcTemplateJsonCfg{
						{Tag: utils.StringPointer(utils.Tenant),
							Path:      utils.StringPointer(utils.Tenant),
							Type:      utils.StringPointer(utils.MetaVariable),
							Value:     utils.StringPointer("~*req.0"),
							Mandatory: utils.BoolPointer(true)},
						{Tag: utils.StringPointer(utils.ID),
							Path:      utils.StringPointer(utils.ID),
							Type:      utils.StringPointer(utils.MetaVariable),
							Value:     utils.StringPointer("~*req.1"),
							Mandatory: utils.BoolPointer(true)},
						{Tag: utils.StringPointer("FilterIDs"),
							Path:  utils.StringPointer("FilterIDs"),
							Type:  utils.StringPointer(utils.MetaVariable),
							Value: utils.StringPointer("~*req.2")},
						{Tag: utils.StringPointer("Weight"),
							Path:  utils.StringPointer("Weight"),
							Type:  utils.StringPointer(utils.MetaVariable),
							Value: utils.StringPointer("~*req.3")},
						{Tag: utils.StringPointer("BalanceID"),
							Path:  utils.StringPointer("BalanceID"),
							Type:  utils.StringPointer(utils.MetaVariable),
							Value: utils.StringPointer("~*req.4")},
						{Tag: utils.StringPointer("BalanceFilterIDs"),
							Path:  utils.StringPointer("BalanceFilterIDs"),
							Type:  utils.StringPointer(utils.MetaVariable),
							Value: utils.StringPointer("~*req.5")},
						{Tag: utils.StringPointer("BalanceWeight"),
							Path:  utils.StringPointer("BalanceWeight"),
							Type:  utils.StringPointer(utils.MetaVariable),
							Value: utils.StringPointer("~*req.6")},
						{Tag: utils.StringPointer("BalanceBlocker"),
							Path:  utils.StringPointer("BalanceBlocker"),
							Type:  utils.StringPointer(utils.MetaVariable),
							Value: utils.StringPointer("~*req.7")},
						{Tag: utils.StringPointer("BalanceType"),
							Path:  utils.StringPointer("BalanceType"),
							Type:  utils.StringPointer(utils.MetaVariable),
							Value: utils.StringPointer("~*req.8")},
						{Tag: utils.StringPointer("BalanceOpts"),
							Path:  utils.StringPointer("BalanceOpts"),
							Type:  utils.StringPointer(utils.MetaVariable),
							Value: utils.StringPointer("~*req.9")},
						{Tag: utils.StringPointer("BalanceCostIncrements"),
							Path:  utils.StringPointer("BalanceCostIncrements"),
							Type:  utils.StringPointer(utils.MetaVariable),
							Value: utils.StringPointer("~*req.10")},
						{Tag: utils.StringPointer("BalanceAttributeIDs"),
							Path:  utils.StringPointer("BalanceAttributeIDs"),
							Type:  utils.StringPointer(utils.MetaVariable),
							Value: utils.StringPointer("~*req.11")},
						{Tag: utils.StringPointer("BalanceRateProfileIDs"),
							Path:  utils.StringPointer("BalanceRateProfileIDs"),
							Type:  utils.StringPointer(utils.MetaVariable),
							Value: utils.StringPointer("~*req.12")},
						{Tag: utils.StringPointer("BalanceUnitFactors"),
							Path:  utils.StringPointer("BalanceUnitFactors"),
							Type:  utils.StringPointer(utils.MetaVariable),
							Value: utils.StringPointer("~*req.13")},
						{Tag: utils.StringPointer("BalanceUnits"),
							Path:  utils.StringPointer("BalanceUnits"),
							Type:  utils.StringPointer(utils.MetaVariable),
							Value: utils.StringPointer("~*req.14")},
						{Tag: utils.StringPointer("ThresholdIDs"),
							Path:  utils.StringPointer("ThresholdIDs"),
							Type:  utils.StringPointer(utils.MetaVariable),
							Value: utils.StringPointer("~*req.15")},
					},
				},
			},
		},
	}
	dfCgrJSONCfg, err := NewCgrJsonCfgFromBytes([]byte(CGRATES_CFG_JSON))
	if err != nil {
		t.Error(err)
	}
	if cfg, err := dfCgrJSONCfg.LoaderJsonCfg(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCfg, cfg) {
		t.Errorf("Expecting: %s,\nreceived: %s ",
			utils.ToJSON(eCfg), utils.ToJSON(cfg))
	}
}

func TestDfSureTaxJsonCfg(t *testing.T) {
	eCfg := &SureTaxJsonCfg{
		Url:                     utils.StringPointer(""),
		Client_number:           utils.StringPointer(""),
		Validation_key:          utils.StringPointer(""),
		Business_unit:           utils.StringPointer(""),
		Timezone:                utils.StringPointer("Local"),
		Include_local_cost:      utils.BoolPointer(false),
		Return_file_code:        utils.StringPointer("0"),
		Response_group:          utils.StringPointer("03"),
		Response_type:           utils.StringPointer("D4"),
		Regulatory_code:         utils.StringPointer("03"),
		Client_tracking:         utils.StringPointer(utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.CGRID),
		Customer_number:         utils.StringPointer("~*req.Subject"),
		Orig_number:             utils.StringPointer("~*req.Subject"),
		Term_number:             utils.StringPointer("~*req.Destination"),
		Bill_to_number:          utils.StringPointer(""),
		Zipcode:                 utils.StringPointer(""),
		Plus4:                   utils.StringPointer(""),
		P2PZipcode:              utils.StringPointer(""),
		P2PPlus4:                utils.StringPointer(""),
		Units:                   utils.StringPointer("1"),
		Unit_type:               utils.StringPointer("00"),
		Tax_included:            utils.StringPointer("0"),
		Tax_situs_rule:          utils.StringPointer("04"),
		Trans_type_code:         utils.StringPointer("010101"),
		Sales_type_code:         utils.StringPointer("R"),
		Tax_exemption_code_list: utils.StringPointer(""),
	}
	dfCgrJSONCfg, err := NewCgrJsonCfgFromBytes([]byte(CGRATES_CFG_JSON))
	if err != nil {
		t.Error(err)
	}
	if cfg, err := dfCgrJSONCfg.SureTaxJsonCfg(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCfg, cfg) {
		t.Error("Received: ", cfg)
	}
}

func TestDfHttpJsonCfg(t *testing.T) {
	eCfg := &HTTPJsonCfg{
		Json_rpc_url:        utils.StringPointer("/jsonrpc"),
		Registrars_url:      utils.StringPointer("/registrar"),
		Ws_url:              utils.StringPointer("/ws"),
		Freeswitch_cdrs_url: utils.StringPointer("/freeswitch_json"),
		Http_Cdrs:           utils.StringPointer("/cdr_http"),
		Use_basic_auth:      utils.BoolPointer(false),
		Auth_users:          utils.MapStringStringPointer(map[string]string{}),
		Client_opts: map[string]interface{}{
			utils.HTTPClientTLSClientConfigCfg:       false,
			utils.HTTPClientTLSHandshakeTimeoutCfg:   "10s",
			utils.HTTPClientDisableKeepAlivesCfg:     false,
			utils.HTTPClientDisableCompressionCfg:    false,
			utils.HTTPClientMaxIdleConnsCfg:          100.,
			utils.HTTPClientMaxIdleConnsPerHostCfg:   2.,
			utils.HTTPClientMaxConnsPerHostCfg:       0.,
			utils.HTTPClientIdleConnTimeoutCfg:       "90s",
			utils.HTTPClientResponseHeaderTimeoutCfg: "0",
			utils.HTTPClientExpectContinueTimeoutCfg: "0",
			utils.HTTPClientForceAttemptHTTP2Cfg:     true,
			utils.HTTPClientDialTimeoutCfg:           "30s",
			utils.HTTPClientDialFallbackDelayCfg:     "300ms",
			utils.HTTPClientDialKeepAliveCfg:         "30s",
		},
	}
	dfCgrJSONCfg, err := NewCgrJsonCfgFromBytes([]byte(CGRATES_CFG_JSON))
	if err != nil {
		t.Error(err)
	}
	if cfg, err := dfCgrJSONCfg.HttpJsonCfg(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCfg, cfg) {
		t.Errorf("Expected: %s ,received: %s", utils.ToJSON(eCfg), utils.ToJSON(cfg))
	}
}

func TestDfDispatcherSJsonCfg(t *testing.T) {
	eCfg := &DispatcherSJsonCfg{
		Enabled:               utils.BoolPointer(false),
		Indexed_selects:       utils.BoolPointer(true),
		String_indexed_fields: nil,
		Prefix_indexed_fields: &[]string{},
		Suffix_indexed_fields: &[]string{},
		Attributes_conns:      &[]string{},
		Nested_fields:         utils.BoolPointer(false),
	}
	dfCgrJSONCfg, err := NewCgrJsonCfgFromBytes([]byte(CGRATES_CFG_JSON))
	if err != nil {
		t.Error(err)
	}
	if cfg, err := dfCgrJSONCfg.DispatcherSJsonCfg(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCfg, cfg) {
		t.Errorf("expecting: %+v, received: %+v", utils.ToJSON(eCfg), utils.ToJSON(cfg))
	}
}

func TestDfLoaderCfg(t *testing.T) {
	cred := json.RawMessage(`".gapi/credentials.json"`)
	tok := json.RawMessage(`".gapi/token.json"`)
	eCfg := &LoaderCfgJson{
		Tpid:             utils.StringPointer(""),
		Data_path:        utils.StringPointer("./"),
		Disable_reverse:  utils.BoolPointer(false),
		Field_separator:  utils.StringPointer(","),
		Caches_conns:     &[]string{utils.MetaLocalHost},
		Actions_conns:    &[]string{utils.MetaLocalHost},
		Gapi_credentials: &cred,
		Gapi_token:       &tok,
	}
	dfCgrJSONCfg, err := NewCgrJsonCfgFromBytes([]byte(CGRATES_CFG_JSON))
	if err != nil {
		t.Error(err)
	}
	if cfg, err := dfCgrJSONCfg.LoaderCfgJson(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCfg, cfg) {
		t.Errorf("Expected: %s, received: %+v", utils.ToJSON(*eCfg), utils.ToJSON(cfg))
	}
}

func TestDfMigratorCfg(t *testing.T) {
	eCfg := &MigratorCfgJson{
		Out_dataDB_type:     utils.StringPointer("redis"),
		Out_dataDB_host:     utils.StringPointer("127.0.0.1"),
		Out_dataDB_port:     utils.StringPointer("6379"),
		Out_dataDB_name:     utils.StringPointer("10"),
		Out_dataDB_user:     utils.StringPointer("cgrates"),
		Out_dataDB_password: utils.StringPointer(""),
		Out_dataDB_encoding: utils.StringPointer("msgpack"),

		Out_storDB_type:     utils.StringPointer("mysql"),
		Out_storDB_host:     utils.StringPointer("127.0.0.1"),
		Out_storDB_port:     utils.StringPointer("3306"),
		Out_storDB_name:     utils.StringPointer("cgrates"),
		Out_storDB_user:     utils.StringPointer("cgrates"),
		Out_storDB_password: utils.StringPointer(""),
		Users_filters:       &[]string{},
		Out_storDB_opts:     make(map[string]interface{}),
		Out_dataDB_opts: map[string]interface{}{
			utils.RedisSentinelNameCfg:       "",
			utils.RedisClusterCfg:            false,
			utils.RedisClusterSyncCfg:        "5s",
			utils.RedisClusterOnDownDelayCfg: "0",
			utils.RedisTLS:                   false,
			utils.RedisClientCertificate:     "",
			utils.RedisClientKey:             "",
			utils.RedisCACertificate:         "",
		},
	}
	dfCgrJSONCfg, err := NewCgrJsonCfgFromBytes([]byte(CGRATES_CFG_JSON))
	if err != nil {
		t.Error(err)
	}
	if cfg, err := dfCgrJSONCfg.MigratorCfgJson(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCfg, cfg) {
		t.Errorf("Expected: %+v, received: %+v", utils.ToJSON(eCfg), utils.ToJSON(cfg))
	}
}

func TestDfTlsCfg(t *testing.T) {
	eCfg := &TlsJsonCfg{
		Server_certificate: utils.StringPointer(""),
		Server_key:         utils.StringPointer(""),
		Ca_certificate:     utils.StringPointer(""),
		Client_certificate: utils.StringPointer(""),
		Client_key:         utils.StringPointer(""),
		Server_name:        utils.StringPointer(""),
		Server_policy:      utils.IntPointer(4),
	}
	dfCgrJSONCfg, err := NewCgrJsonCfgFromBytes([]byte(CGRATES_CFG_JSON))
	if err != nil {
		t.Error(err)
	}
	if cfg, err := dfCgrJSONCfg.TlsCfgJson(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCfg, cfg) {
		t.Errorf("Expected: %+v, received: %+v", utils.ToJSON(eCfg), utils.ToJSON(cfg))
	}
}

func TestDfAnalyzerCfg(t *testing.T) {
	eCfg := &AnalyzerSJsonCfg{
		Enabled:          utils.BoolPointer(false),
		Cleanup_interval: utils.StringPointer("1h"),
		Db_path:          utils.StringPointer("/var/spool/cgrates/analyzers"),
		Index_type:       utils.StringPointer(utils.MetaScorch),
		Ttl:              utils.StringPointer("24h"),
	}
	dfCgrJSONCfg, err := NewCgrJsonCfgFromBytes([]byte(CGRATES_CFG_JSON))
	if err != nil {
		t.Error(err)
	}
	if cfg, err := dfCgrJSONCfg.AnalyzerCfgJson(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCfg, cfg) {
		t.Errorf("Expected: %+v, received: %+v", utils.ToJSON(eCfg), utils.ToJSON(cfg))
	}
}

func TestDfApierCfg(t *testing.T) {
	eCfg := &AdminSJsonCfg{
		Enabled:          utils.BoolPointer(false),
		Caches_conns:     &[]string{utils.MetaInternal},
		Actions_conns:    &[]string{},
		Attributes_conns: &[]string{},
		Ees_conns:        &[]string{},
	}
	dfCgrJSONCfg, err := NewCgrJsonCfgFromBytes([]byte(CGRATES_CFG_JSON))
	if err != nil {
		t.Error(err)
	}
	if cfg, err := dfCgrJSONCfg.AdminSCfgJson(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCfg, cfg) {
		t.Errorf("Expected: %+v, received: %+v", utils.ToJSON(eCfg), utils.ToJSON(cfg))
	}
}

func TestDfEventReaderCfg(t *testing.T) {
	cdrFields := []*FcTemplateJsonCfg{
		{Tag: utils.StringPointer(utils.ToR), Path: utils.StringPointer(utils.MetaCgreq + utils.NestingSep + utils.ToR), Type: utils.StringPointer(utils.MetaVariable),
			Value: utils.StringPointer("~*req.2"), Mandatory: utils.BoolPointer(true)},
		{Tag: utils.StringPointer(utils.OriginID), Path: utils.StringPointer(utils.MetaCgreq + utils.NestingSep + utils.OriginID), Type: utils.StringPointer(utils.MetaVariable),
			Value: utils.StringPointer("~*req.3"), Mandatory: utils.BoolPointer(true)},
		{Tag: utils.StringPointer(utils.RequestType), Path: utils.StringPointer(utils.MetaCgreq + utils.NestingSep + utils.RequestType), Type: utils.StringPointer(utils.MetaVariable),
			Value: utils.StringPointer("~*req.4"), Mandatory: utils.BoolPointer(true)},
		{Tag: utils.StringPointer(utils.Tenant), Path: utils.StringPointer(utils.MetaCgreq + utils.NestingSep + utils.Tenant), Type: utils.StringPointer(utils.MetaVariable),
			Value: utils.StringPointer("~*req.6"), Mandatory: utils.BoolPointer(true)},
		{Tag: utils.StringPointer(utils.Category), Path: utils.StringPointer(utils.MetaCgreq + utils.NestingSep + utils.Category), Type: utils.StringPointer(utils.MetaVariable),
			Value: utils.StringPointer("~*req.7"), Mandatory: utils.BoolPointer(true)},
		{Tag: utils.StringPointer(utils.AccountField), Path: utils.StringPointer(utils.MetaCgreq + utils.NestingSep + utils.AccountField), Type: utils.StringPointer(utils.MetaVariable),
			Value: utils.StringPointer("~*req.8"), Mandatory: utils.BoolPointer(true)},
		{Tag: utils.StringPointer(utils.Subject), Path: utils.StringPointer(utils.MetaCgreq + utils.NestingSep + utils.Subject), Type: utils.StringPointer(utils.MetaVariable),
			Value: utils.StringPointer("~*req.9"), Mandatory: utils.BoolPointer(true)},
		{Tag: utils.StringPointer(utils.Destination), Path: utils.StringPointer(utils.MetaCgreq + utils.NestingSep + utils.Destination), Type: utils.StringPointer(utils.MetaVariable),
			Value: utils.StringPointer("~*req.10"), Mandatory: utils.BoolPointer(true)},
		{Tag: utils.StringPointer(utils.SetupTime), Path: utils.StringPointer(utils.MetaCgreq + utils.NestingSep + utils.SetupTime), Type: utils.StringPointer(utils.MetaVariable),
			Value: utils.StringPointer("~*req.11"), Mandatory: utils.BoolPointer(true)},
		{Tag: utils.StringPointer(utils.AnswerTime), Path: utils.StringPointer(utils.MetaCgreq + utils.NestingSep + utils.AnswerTime), Type: utils.StringPointer(utils.MetaVariable),
			Value: utils.StringPointer("~*req.12"), Mandatory: utils.BoolPointer(true)},
		{Tag: utils.StringPointer(utils.Usage), Path: utils.StringPointer(utils.MetaCgreq + utils.NestingSep + utils.Usage), Type: utils.StringPointer(utils.MetaVariable),
			Value: utils.StringPointer("~*req.13"), Mandatory: utils.BoolPointer(true)},
	}
	eCfg := &ERsJsonCfg{
		Enabled:        utils.BoolPointer(false),
		Sessions_conns: &[]string{utils.MetaInternal},
		Readers: &[]*EventReaderJsonCfg{
			{
				Id:                    utils.StringPointer(utils.MetaDefault),
				Type:                  utils.StringPointer(utils.MetaNone),
				Run_delay:             utils.StringPointer("0"),
				Concurrent_requests:   utils.IntPointer(1024),
				Source_path:           utils.StringPointer("/var/spool/cgrates/ers/in"),
				Processed_path:        utils.StringPointer("/var/spool/cgrates/ers/out"),
				Tenant:                utils.StringPointer(utils.EmptyString),
				Timezone:              utils.StringPointer(utils.EmptyString),
				Filters:               &[]string{},
				Flags:                 &[]string{},
				Fields:                &cdrFields,
				Cache_dump_fields:     &[]*FcTemplateJsonCfg{},
				Partial_commit_fields: &[]*FcTemplateJsonCfg{},
				Opts: map[string]interface{}{
					"csvFieldSeparator":   ",",
					"csvHeaderDefineChar": ":",
					"csvRowLength":        0.,
					"xmlRootPath":         "",
					"partialOrderField":   "~*req.AnswerTime",
					"natsSubject":         "cgrates_cdrs",
				},
			},
		},
		Partial_cache_ttl:    utils.StringPointer("1s"),
		Partial_cache_action: utils.StringPointer(utils.MetaNone),
	}
	dfCgrJSONCfg, err := NewCgrJsonCfgFromBytes([]byte(CGRATES_CFG_JSON))
	if err != nil {
		t.Error(err)
	}
	if cfg, err := dfCgrJSONCfg.ERsJsonCfg(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCfg, cfg) {
		t.Errorf("Expected: %+v, \nreceived: %+v", utils.ToJSON(eCfg), utils.ToJSON(cfg))
	}
}

func TestDfEventExporterCfg(t *testing.T) {
	eCfg := &EEsJsonCfg{
		Enabled:          utils.BoolPointer(false),
		Attributes_conns: &[]string{},
		Cache: map[string]*CacheParamJsonCfg{
			utils.MetaFileCSV: {
				Limit:      utils.IntPointer(-1),
				Ttl:        utils.StringPointer("5s"),
				Static_ttl: utils.BoolPointer(false),
			},
		},
		Exporters: &[]*EventExporterJsonCfg{
			{
				Id:                utils.StringPointer(utils.MetaDefault),
				Type:              utils.StringPointer(utils.MetaNone),
				Export_path:       utils.StringPointer("/var/spool/cgrates/ees"),
				Attribute_context: utils.StringPointer(utils.EmptyString),
				Tenant:            utils.StringPointer(utils.EmptyString),
				Timezone:          utils.StringPointer(utils.EmptyString),
				Filters:           &[]string{},
				Attribute_ids:     &[]string{},
				Flags:             &[]string{},
				Synchronous:       utils.BoolPointer(false),
				Attempts:          utils.IntPointer(1),
				Fields:            &[]*FcTemplateJsonCfg{},
				Opts:              make(map[string]interface{}),
			},
		},
	}
	dfCgrJSONCfg, err := NewCgrJsonCfgFromBytes([]byte(CGRATES_CFG_JSON))
	if err != nil {
		t.Error(err)
	}
	if cfg, err := dfCgrJSONCfg.EEsJsonCfg(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCfg, cfg) {
		t.Errorf("Expected: %+v, \nreceived: %+v", utils.ToJSON(eCfg), utils.ToJSON(cfg))
	}
}

func TestDfRateSJsonCfg(t *testing.T) {
	eCfg := &RateSJsonCfg{
		Enabled:                    utils.BoolPointer(false),
		Indexed_selects:            utils.BoolPointer(true),
		String_indexed_fields:      nil,
		Prefix_indexed_fields:      &[]string{},
		Suffix_indexed_fields:      &[]string{},
		Nested_fields:              utils.BoolPointer(false),
		Rate_indexed_selects:       utils.BoolPointer(true),
		Rate_string_indexed_fields: nil,
		Rate_prefix_indexed_fields: &[]string{},
		Rate_suffix_indexed_fields: &[]string{},
		Rate_nested_fields:         utils.BoolPointer(false),
		Verbosity:                  utils.IntPointer(1000),
	}
	dfCgrJSONCfg, err := NewCgrJsonCfgFromBytes([]byte(CGRATES_CFG_JSON))
	if err != nil {
		t.Error(err)
	}
	if cfg, err := dfCgrJSONCfg.RateCfgJson(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCfg, cfg) {
		t.Error("Received: ", utils.ToJSON(cfg))
	}
}

func TestDfTemplateSJsonCfg(t *testing.T) {
	eCfg := FcTemplatesJsonCfg{
		"*errSip": {
			{
				Tag:       utils.StringPointer("Request"),
				Path:      utils.StringPointer(fmt.Sprintf("%s.Request", utils.MetaRep)),
				Type:      utils.StringPointer(utils.MetaConstant),
				Value:     utils.StringPointer("SIP/2.0 500 Internal Server Error"),
				Mandatory: utils.BoolPointer(true)},
		},
		utils.MetaErr: {
			{
				Tag:       utils.StringPointer("SessionId"),
				Path:      utils.StringPointer(fmt.Sprintf("%s.Session-Id", utils.MetaRep)),
				Type:      utils.StringPointer(utils.MetaVariable),
				Value:     utils.StringPointer("~*req.Session-Id"),
				Mandatory: utils.BoolPointer(true)},
			{
				Tag:       utils.StringPointer("OriginHost"),
				Path:      utils.StringPointer(fmt.Sprintf("%s.Origin-Host", utils.MetaRep)),
				Type:      utils.StringPointer(utils.MetaVariable),
				Value:     utils.StringPointer("~*vars.OriginHost"),
				Mandatory: utils.BoolPointer(true)},
			{
				Tag:       utils.StringPointer("OriginRealm"),
				Path:      utils.StringPointer(fmt.Sprintf("%s.Origin-Realm", utils.MetaRep)),
				Type:      utils.StringPointer(utils.MetaVariable),
				Value:     utils.StringPointer("~*vars.OriginRealm"),
				Mandatory: utils.BoolPointer(true)},
		},
		utils.MetaCCA: {
			{
				Tag:       utils.StringPointer("SessionId"),
				Path:      utils.StringPointer(fmt.Sprintf("%s.Session-Id", utils.MetaRep)),
				Type:      utils.StringPointer(utils.MetaVariable),
				Value:     utils.StringPointer("~*req.Session-Id"),
				Mandatory: utils.BoolPointer(true)},
			{
				Tag:   utils.StringPointer("ResultCode"),
				Path:  utils.StringPointer(fmt.Sprintf("%s.Result-Code", utils.MetaRep)),
				Type:  utils.StringPointer(utils.MetaConstant),
				Value: utils.StringPointer("2001")},
			{
				Tag:       utils.StringPointer("OriginHost"),
				Path:      utils.StringPointer(fmt.Sprintf("%s.Origin-Host", utils.MetaRep)),
				Type:      utils.StringPointer(utils.MetaVariable),
				Value:     utils.StringPointer("~*vars.OriginHost"),
				Mandatory: utils.BoolPointer(true)},
			{
				Tag:       utils.StringPointer("OriginRealm"),
				Path:      utils.StringPointer(fmt.Sprintf("%s.Origin-Realm", utils.MetaRep)),
				Type:      utils.StringPointer(utils.MetaVariable),
				Value:     utils.StringPointer("~*vars.OriginRealm"),
				Mandatory: utils.BoolPointer(true)},
			{
				Tag:       utils.StringPointer("AuthApplicationId"),
				Path:      utils.StringPointer(fmt.Sprintf("%s.Auth-Application-Id", utils.MetaRep)),
				Type:      utils.StringPointer(utils.MetaVariable),
				Value:     utils.StringPointer("~*vars.*appid"),
				Mandatory: utils.BoolPointer(true)},
			{
				Tag:       utils.StringPointer("CCRequestType"),
				Path:      utils.StringPointer(fmt.Sprintf("%s.CC-Request-Type", utils.MetaRep)),
				Type:      utils.StringPointer(utils.MetaVariable),
				Value:     utils.StringPointer("~*req.CC-Request-Type"),
				Mandatory: utils.BoolPointer(true)},
			{
				Tag:       utils.StringPointer("CCRequestNumber"),
				Path:      utils.StringPointer(fmt.Sprintf("%s.CC-Request-Number", utils.MetaRep)),
				Type:      utils.StringPointer(utils.MetaVariable),
				Value:     utils.StringPointer("~*req.CC-Request-Number"),
				Mandatory: utils.BoolPointer(true)},
		},
		utils.MetaASR: {
			{
				Tag:       utils.StringPointer("SessionId"),
				Path:      utils.StringPointer(fmt.Sprintf("%s.Session-Id", utils.MetaDiamreq)),
				Type:      utils.StringPointer(utils.MetaVariable),
				Value:     utils.StringPointer("~*req.Session-Id"),
				Mandatory: utils.BoolPointer(true)},
			{
				Tag:       utils.StringPointer("OriginHost"),
				Path:      utils.StringPointer(fmt.Sprintf("%s.Origin-Host", utils.MetaDiamreq)),
				Type:      utils.StringPointer(utils.MetaVariable),
				Value:     utils.StringPointer("~*req.Destination-Host"),
				Mandatory: utils.BoolPointer(true)},
			{
				Tag:       utils.StringPointer("OriginRealm"),
				Path:      utils.StringPointer(fmt.Sprintf("%s.Origin-Realm", utils.MetaDiamreq)),
				Type:      utils.StringPointer(utils.MetaVariable),
				Value:     utils.StringPointer("~*req.Destination-Realm"),
				Mandatory: utils.BoolPointer(true)},
			{
				Tag:       utils.StringPointer("DestinationRealm"),
				Path:      utils.StringPointer(fmt.Sprintf("%s.Destination-Realm", utils.MetaDiamreq)),
				Type:      utils.StringPointer(utils.MetaVariable),
				Value:     utils.StringPointer("~*req.Origin-Realm"),
				Mandatory: utils.BoolPointer(true)},
			{
				Tag:       utils.StringPointer("DestinationHost"),
				Path:      utils.StringPointer(fmt.Sprintf("%s.Destination-Host", utils.MetaDiamreq)),
				Type:      utils.StringPointer(utils.MetaVariable),
				Value:     utils.StringPointer("~*req.Origin-Host"),
				Mandatory: utils.BoolPointer(true)},
			{
				Tag:       utils.StringPointer("AuthApplicationId"),
				Path:      utils.StringPointer(fmt.Sprintf("%s.Auth-Application-Id", utils.MetaDiamreq)),
				Type:      utils.StringPointer(utils.MetaVariable),
				Value:     utils.StringPointer("~*vars.*appid"),
				Mandatory: utils.BoolPointer(true)},
		},
		utils.MetaRAR: {
			{
				Tag:       utils.StringPointer("SessionId"),
				Path:      utils.StringPointer(fmt.Sprintf("%s.Session-Id", utils.MetaDiamreq)),
				Type:      utils.StringPointer(utils.MetaVariable),
				Value:     utils.StringPointer("~*req.Session-Id"),
				Mandatory: utils.BoolPointer(true)},
			{
				Tag:       utils.StringPointer("OriginHost"),
				Path:      utils.StringPointer(fmt.Sprintf("%s.Origin-Host", utils.MetaDiamreq)),
				Type:      utils.StringPointer(utils.MetaVariable),
				Value:     utils.StringPointer("~*req.Destination-Host"),
				Mandatory: utils.BoolPointer(true)},
			{
				Tag:       utils.StringPointer("OriginRealm"),
				Path:      utils.StringPointer(fmt.Sprintf("%s.Origin-Realm", utils.MetaDiamreq)),
				Type:      utils.StringPointer(utils.MetaVariable),
				Value:     utils.StringPointer("~*req.Destination-Realm"),
				Mandatory: utils.BoolPointer(true)},
			{
				Tag:       utils.StringPointer("DestinationRealm"),
				Path:      utils.StringPointer(fmt.Sprintf("%s.Destination-Realm", utils.MetaDiamreq)),
				Type:      utils.StringPointer(utils.MetaVariable),
				Value:     utils.StringPointer("~*req.Origin-Realm"),
				Mandatory: utils.BoolPointer(true)},
			{
				Tag:       utils.StringPointer("DestinationHost"),
				Path:      utils.StringPointer(fmt.Sprintf("%s.Destination-Host", utils.MetaDiamreq)),
				Type:      utils.StringPointer(utils.MetaVariable),
				Value:     utils.StringPointer("~*req.Origin-Host"),
				Mandatory: utils.BoolPointer(true)},
			{
				Tag:       utils.StringPointer("AuthApplicationId"),
				Path:      utils.StringPointer(fmt.Sprintf("%s.Auth-Application-Id", utils.MetaDiamreq)),
				Type:      utils.StringPointer(utils.MetaVariable),
				Value:     utils.StringPointer("~*vars.*appid"),
				Mandatory: utils.BoolPointer(true)},
			{
				Tag:   utils.StringPointer("ReAuthRequestType"),
				Path:  utils.StringPointer(fmt.Sprintf("%s.Re-Auth-Request-Type", utils.MetaDiamreq)),
				Type:  utils.StringPointer(utils.MetaConstant),
				Value: utils.StringPointer("0"),
			},
		},
		utils.MetaCdrLog: {
			{
				Tag:       utils.StringPointer("ToR"),
				Path:      utils.StringPointer("*cdr.ToR"),
				Type:      utils.StringPointer(utils.MetaVariable),
				Value:     utils.StringPointer("~*req.BalanceType"),
				Mandatory: utils.BoolPointer(true),
			},
			{
				Tag:       utils.StringPointer("OriginHost"),
				Path:      utils.StringPointer("*cdr.OriginHost"),
				Type:      utils.StringPointer(utils.MetaConstant),
				Value:     utils.StringPointer("127.0.0.1"),
				Mandatory: utils.BoolPointer(true),
			},
			{
				Tag:       utils.StringPointer("RequestType"),
				Path:      utils.StringPointer("*cdr.RequestType"),
				Type:      utils.StringPointer(utils.MetaConstant),
				Value:     utils.StringPointer(utils.MetaNone),
				Mandatory: utils.BoolPointer(true),
			},
			{
				Tag:       utils.StringPointer("Tenant"),
				Path:      utils.StringPointer("*cdr.Tenant"),
				Type:      utils.StringPointer(utils.MetaVariable),
				Value:     utils.StringPointer("~*req.Tenant"),
				Mandatory: utils.BoolPointer(true),
			},
			{
				Tag:       utils.StringPointer("Account"),
				Path:      utils.StringPointer("*cdr.Account"),
				Type:      utils.StringPointer(utils.MetaVariable),
				Value:     utils.StringPointer("~*req.Account"),
				Mandatory: utils.BoolPointer(true),
			},
			{
				Tag:       utils.StringPointer("Subject"),
				Path:      utils.StringPointer("*cdr.Subject"),
				Type:      utils.StringPointer(utils.MetaVariable),
				Value:     utils.StringPointer("~*req.Account"),
				Mandatory: utils.BoolPointer(true),
			},
			{
				Tag:       utils.StringPointer("Cost"),
				Path:      utils.StringPointer("*cdr.Cost"),
				Type:      utils.StringPointer(utils.MetaVariable),
				Value:     utils.StringPointer("~*req.Cost"),
				Mandatory: utils.BoolPointer(true),
			},
			{
				Tag:       utils.StringPointer("Source"),
				Path:      utils.StringPointer("*cdr.Source"),
				Type:      utils.StringPointer(utils.MetaConstant),
				Value:     utils.StringPointer(utils.MetaCdrLog),
				Mandatory: utils.BoolPointer(true),
			},
			{
				Tag:       utils.StringPointer("Usage"),
				Path:      utils.StringPointer("*cdr.Usage"),
				Type:      utils.StringPointer(utils.MetaConstant),
				Value:     utils.StringPointer("1"),
				Mandatory: utils.BoolPointer(true),
			},
			{
				Tag:       utils.StringPointer("RunID"),
				Path:      utils.StringPointer("*cdr.RunID"),
				Type:      utils.StringPointer(utils.MetaVariable),
				Value:     utils.StringPointer("~*req.ActionType"),
				Mandatory: utils.BoolPointer(true),
			},
			{
				Tag:       utils.StringPointer("SetupTime"),
				Path:      utils.StringPointer("*cdr.SetupTime"),
				Type:      utils.StringPointer(utils.MetaConstant),
				Value:     utils.StringPointer("*now"),
				Mandatory: utils.BoolPointer(true),
			},
			{
				Tag:       utils.StringPointer("AnswerTime"),
				Path:      utils.StringPointer("*cdr.AnswerTime"),
				Type:      utils.StringPointer(utils.MetaConstant),
				Value:     utils.StringPointer("*now"),
				Mandatory: utils.BoolPointer(true),
			},
			{
				Tag:       utils.StringPointer("PreRated"),
				Path:      utils.StringPointer("*cdr.PreRated"),
				Type:      utils.StringPointer(utils.MetaConstant),
				Value:     utils.StringPointer("true"),
				Mandatory: utils.BoolPointer(true),
			},
		},
	}
	dfCgrJSONCfg, err := NewCgrJsonCfgFromBytes([]byte(CGRATES_CFG_JSON))
	if err != nil {
		t.Error(err)
	}
	if cfg, err := dfCgrJSONCfg.TemplateSJsonCfg(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCfg, cfg) {
		t.Errorf("Expected: %+v \n,received: %+v", utils.ToJSON(eCfg), utils.ToJSON(cfg))
	}
}

func TestDfActionSJsonCfg(t *testing.T) {
	eCfg := &ActionSJsonCfg{
		Enabled:                   utils.BoolPointer(false),
		Cdrs_conns:                &[]string{},
		Ees_conns:                 &[]string{},
		Thresholds_conns:          &[]string{},
		Stats_conns:               &[]string{},
		Accounts_conns:            &[]string{},
		Tenants:                   &[]string{},
		Indexed_selects:           utils.BoolPointer(true),
		String_indexed_fields:     nil,
		Prefix_indexed_fields:     &[]string{},
		Suffix_indexed_fields:     &[]string{},
		Nested_fields:             utils.BoolPointer(false),
		Dynaprepaid_actionprofile: &[]string{},
	}
	dfCgrJSONCfg, err := NewCgrJsonCfgFromBytes([]byte(CGRATES_CFG_JSON))
	if err != nil {
		t.Error(err)
	}
	if cfg, err := dfCgrJSONCfg.ActionSCfgJson(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCfg, cfg) {
		t.Errorf("\n Expected <%+v>,\nReceived:<%+v>", utils.ToJSON(eCfg), utils.ToJSON(cfg))
	}
}

func TestSetSection(t *testing.T) {
	jsn := `
		{
			"general": {
				"Node_id":2,
			},
		}
	`
	jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(jsn))
	if err != nil {
		t.Error(err)
	}
	payload := make(chan struct{})
	errExpect := "json: unsupported type: chan struct {}"
	if err := jsnCfg.SetSection(&context.Context{}, "general", payload); err == nil || err.Error() != errExpect {
		t.Error(err)
	}
}
