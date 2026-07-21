//go:build integration
// +build integration

// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"path"
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

func TestLoadConfig(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaRedis:
	case utils.MetaInternal, utils.MetaMongo, utils.MetaMySQL, utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	// DataDb
	*cfgPath = path.Join(*utils.DataDir, "conf", "samples", "tutredis")
	*dataDBType = utils.MetaRedis
	*dataDBHost = "localhost"
	*dataDBPort = "2012"
	*dataDBName = "100"
	*dataDBUser = "cgrates2"
	*dataDBPasswd = "toor"
	*dbRedisSentinel = "sentinel1"
	expDBcfg := &config.DbCfg{
		DBConns: config.DBConns{
			utils.StorDB: &config.DBConn{
				Type:                utils.MetaMySQL,
				Host:                "127.0.0.1",
				Port:                "3306",
				Name:                utils.CGRateSLwr,
				User:                utils.CGRateSLwr,
				Password:            "CGRateS.org",
				StringIndexedFields: []string{},
				PrefixIndexedFields: []string{},
				RmtConns:            []string{},
				RplConns:            []string{},
				Opts: &config.DBOpts{
					InternalDBDumpPath:        "/var/lib/cgrates/internal_db/db",
					InternalDBBackupPath:      "/var/lib/cgrates/internal_db/backup/db",
					InternalDBStartTimeout:    300000000000,
					InternalDBDumpInterval:    time.Minute,
					InternalDBRewriteInterval: time.Hour,
					InternalDBFileSizeLimit:   1073741824,
					RedisBatchSize:            1000,
					RedisMaxConns:             10,
					RedisConnectAttempts:      20,
					MongoQueryTimeout:         10 * time.Second,
					MongoConnScheme:           "mongodb",
					RedisClusterSync:          5 * time.Second,
					RedisClusterOndownDelay:   0,
					RedisCluster:              false,
					RedisPoolPipelineWindow:   150 * time.Microsecond,
					RedisTLS:                  false,
					SQLMaxOpenConns:           100,
					SQLMaxIdleConns:           10,
					SQLLogLevel:               3,
					SQLConnMaxLifetime:        0,
					SQLDSNParams:              map[string]string{},
					PgSSLMode:                 "disable",
					MySQLLocation:             "Local",
				},
			},
			utils.MetaDefault: &config.DBConn{
				Type:                utils.MetaRedis,
				Host:                "localhost",
				Port:                "2012",
				Name:                "100",
				User:                "cgrates2",
				StringIndexedFields: []string{},
				PrefixIndexedFields: []string{},
				RmtConns:            []string{},
				RplConns:            []string{},
				Password:            "toor",
				Opts: &config.DBOpts{
					InternalDBDumpPath:        "/var/lib/cgrates/internal_db/db",
					InternalDBBackupPath:      "/var/lib/cgrates/internal_db/backup/db",
					InternalDBStartTimeout:    300000000000,
					InternalDBDumpInterval:    time.Minute,
					InternalDBRewriteInterval: time.Hour,
					InternalDBFileSizeLimit:   1073741824,
					RedisBatchSize:            1000,
					RedisMaxConns:             10,
					RedisConnectAttempts:      20,
					RedisSentinel:             "sentinel1",
					MongoQueryTimeout:         10 * time.Second,
					MongoConnScheme:           "mongodb",
					RedisClusterSync:          5 * time.Second,
					RedisClusterOndownDelay:   0,
					RedisCluster:              false,
					RedisPoolPipelineWindow:   150 * time.Microsecond,
					RedisTLS:                  false,
					SQLMaxOpenConns:           100,
					SQLMaxIdleConns:           10,
					SQLLogLevel:               3,
					SQLConnMaxLifetime:        0,
					SQLDSNParams:              map[string]string{},
					PgSSLMode:                 "disable",
					MySQLLocation:             "Local",
				},
			},
		},
	}
	// Loader
	*tpid = "1"
	*disableReverse = true
	*dataPath = "./path"
	*fieldSep = "$"
	*cacheSAddress = ""
	*schedulerAddress = ""
	// General
	*cachingArg = utils.MetaLoad
	*cachingDlay = 5 * time.Second
	*dbDataEncoding = utils.MetaJSON
	*timezone = utils.Local
	ldrCfg := loadConfig()
	ldrCfg.DbCfg().Items = nil
	if !reflect.DeepEqual(ldrCfg.DbCfg(), expDBcfg) {
		t.Errorf("Expected %s \nreceived %s", utils.ToJSON(expDBcfg), utils.ToJSON(ldrCfg.DbCfg()))
	}
	if ldrCfg.GeneralCfg().DBDataEncoding != utils.MetaJSON {
		t.Errorf("Expected %s received %s", utils.MetaJSON, ldrCfg.GeneralCfg().DBDataEncoding)
	}
	if ldrCfg.GeneralCfg().DefaultTimezone != utils.Local {
		t.Errorf("Expected %s received %s", utils.Local, ldrCfg.GeneralCfg().DefaultTimezone)
	}
	if !ldrCfg.LoaderCgrCfg().DisableReverse {
		t.Errorf("Expected %v received %v", true, ldrCfg.LoaderCgrCfg().DisableReverse)
	}
	if ldrCfg.GeneralCfg().DefaultCaching != utils.MetaLoad {
		t.Errorf("Expected %s received %s", utils.MetaLoad, ldrCfg.GeneralCfg().DefaultCaching)
	}
	if ldrCfg.GeneralCfg().CachingDelay != 5*time.Second {
		t.Errorf("Expected %s received %s", 5*time.Second, ldrCfg.GeneralCfg().CachingDelay)
	}
	if *importID == utils.EmptyString {
		t.Errorf("Expected importID to be populated")
	}
	if ldrCfg.LoaderCgrCfg().TpID != "1" {
		t.Errorf("Expected %s received %s", "1", ldrCfg.LoaderCgrCfg().TpID)
	}
	if ldrCfg.LoaderCgrCfg().DataPath != "./path" {
		t.Errorf("Expected %s received %s", "./path", ldrCfg.LoaderCgrCfg().DataPath)
	}
	if ldrCfg.LoaderCgrCfg().FieldSeparator != '$' {
		t.Errorf("Expected %v received %v", '$', ldrCfg.LoaderCgrCfg().FieldSeparator)
	}
	if !reflect.DeepEqual(ldrCfg.LoaderCgrCfg().CachesConns, []string{}) {
		t.Errorf("Expected %v received %v", []string{}, ldrCfg.LoaderCgrCfg().CachesConns)
	}
	if !reflect.DeepEqual(ldrCfg.LoaderCgrCfg().ActionSConns, []string{}) {
		t.Errorf("Expected %v received %v", []string{}, ldrCfg.LoaderCgrCfg().ActionSConns)
	}
	*cacheSAddress = "127.0.0.1"
	*schedulerAddress = "127.0.0.2"
	*rpcEncoding = utils.MetaJSON
	ldrCfg = loadConfig()
	expAddrs := []string{"127.0.0.1"}
	if !reflect.DeepEqual(ldrCfg.LoaderCgrCfg().CachesConns, expAddrs) {
		t.Errorf("Expected %v received %v", expAddrs, ldrCfg.LoaderCgrCfg().CachesConns)
	}
	expAddrs = []string{"127.0.0.2"}
	if !reflect.DeepEqual(ldrCfg.LoaderCgrCfg().ActionSConns, expAddrs) {
		t.Errorf("Expected %v received %v", expAddrs, ldrCfg.LoaderCgrCfg().ActionSConns)
	}
	expaddr := config.RPCConns{
		utils.MetaBiJSONLocalHost: {
			Strategy: rpcclient.PoolFirst,
			PoolSize: 0,
			Conns: []*config.RemoteHost{{
				Address:   "127.0.0.1:2014",
				Transport: rpcclient.BiRPCJSON,
			}},
		},
		utils.MetaInternal: {
			Strategy: rpcclient.PoolFirst,
			PoolSize: 0,
			Conns: []*config.RemoteHost{{
				Address: utils.MetaInternal,
			}},
		},
		rpcclient.BiRPCInternal: {
			Strategy: rpcclient.PoolFirst,
			PoolSize: 0,
			Conns: []*config.RemoteHost{{
				Address: rpcclient.BiRPCInternal,
			}},
		},
		"*localhost": {
			Strategy: rpcclient.PoolFirst,
			Conns:    []*config.RemoteHost{{Address: "127.0.0.1:2012", Transport: utils.MetaJSON}},
		},
		"127.0.0.1": {
			Strategy: rpcclient.PoolFirst,
			Conns:    []*config.RemoteHost{{Address: "127.0.0.1", Transport: utils.MetaJSON}},
		},
		"127.0.0.2": {
			Strategy: rpcclient.PoolFirst,
			Conns:    []*config.RemoteHost{{Address: "127.0.0.2", Transport: utils.MetaJSON}},
		},
	}
	if !reflect.DeepEqual(ldrCfg.RPCConns(), expaddr) {
		t.Errorf("Expected %v received %v", utils.ToJSON(expaddr), utils.ToJSON(ldrCfg.RPCConns()))
	}
}
