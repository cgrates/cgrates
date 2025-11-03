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
	"fmt"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/ltcache"
)

func defaultDBPort(dbType, port string) string {
	if port == utils.MetaDynamic {
		switch dbType {
		case utils.MetaMySQL:
			port = "3306"
		case utils.MetaPostgres:
			port = "5432"
		case utils.MetaMongo:
			port = "27017"
		case utils.MetaRedis:
			port = "6379"
		case utils.MetaInternal:
			port = "internal"
		}
	}
	return port
}

type DBOpts struct {
	InternalDBDumpPath        string        // Path to the dump file
	InternalDBBackupPath      string        // Path where db dump will backup
	InternalDBStartTimeout    time.Duration // Transcache recover from dump files timeout duration
	InternalDBDumpInterval    time.Duration // Regurarly dump database to file
	InternalDBRewriteInterval time.Duration // Regurarly rewrite dump files
	InternalDBFileSizeLimit   int64         // maximum size that can be written in a singular dump file
	RedisMaxConns             int
	RedisConnectAttempts      int
	RedisSentinel             string
	RedisCluster              bool
	RedisClusterSync          time.Duration
	RedisClusterOndownDelay   time.Duration
	RedisConnectTimeout       time.Duration
	RedisReadTimeout          time.Duration
	RedisWriteTimeout         time.Duration
	RedisPoolPipelineWindow   time.Duration
	RedisPoolPipelineLimit    int
	RedisTLS                  bool
	RedisClientCertificate    string
	RedisClientKey            string
	RedisCACertificate        string
	MongoQueryTimeout         time.Duration
	MongoConnScheme           string
	SQLMaxOpenConns           int
	SQLMaxIdleConns           int
	SQLLogLevel               int
	SQLConnMaxLifetime        time.Duration
	SQLDSNParams              map[string]string
	PgSSLMode                 string
	PgSSLCert                 string
	PgSSLKey                  string
	PgSSLPassword             string
	PgSSLCertMode             string
	PgSSLRootCert             string
	MySQLLocation             string
}

// DBConn the config to establish connection to DataDB
type DBConn struct {
	Type                string
	Host                string // The host to connect to. Values that start with / are for UNIX domain sockets.
	Port                string // The port to bind to.
	Name                string // The name of the database to connect to.
	User                string // The user to sign in as.
	Password            string // The user's password.
	StringIndexedFields []string
	PrefixIndexedFields []string
	RmtConns            []string // Remote DataDB  connIDs
	RmtConnID           string
	RplConns            []string // Replication connIDs
	RplFiltered         bool
	RplCache            string
	Opts                *DBOpts
}

// loadFromJSONCfg load the DBConn section of the DataDBCfg
func (dbC *DBConn) loadFromJSONCfg(jsnDbConnCfg *DbConnJson) (err error) {
	if dbC == nil {
		return
	}
	if jsnDbConnCfg.Db_type != nil {
		if !strings.HasPrefix(*jsnDbConnCfg.Db_type, "*") {
			dbC.Type = fmt.Sprintf("*%v", *jsnDbConnCfg.Db_type)
		} else {
			dbC.Type = *jsnDbConnCfg.Db_type
		}
	}
	if jsnDbConnCfg.Db_host != nil {
		dbC.Host = *jsnDbConnCfg.Db_host
	}
	if jsnDbConnCfg.Db_port != nil {
		port := strconv.Itoa(*jsnDbConnCfg.Db_port)
		if port == "-1" {
			port = utils.MetaDynamic
		}
		dbC.Port = defaultDBPort(dbC.Type, port)
	}
	if jsnDbConnCfg.Db_name != nil {
		dbC.Name = *jsnDbConnCfg.Db_name
	}
	if jsnDbConnCfg.Db_user != nil {
		dbC.User = *jsnDbConnCfg.Db_user
	}
	if jsnDbConnCfg.Db_password != nil {
		dbC.Password = *jsnDbConnCfg.Db_password
	}
	if jsnDbConnCfg.String_indexed_fields != nil {
		dbC.StringIndexedFields = *jsnDbConnCfg.String_indexed_fields
	}
	if jsnDbConnCfg.Prefix_indexed_fields != nil {
		dbC.PrefixIndexedFields = *jsnDbConnCfg.Prefix_indexed_fields
	}
	if jsnDbConnCfg.Remote_conns != nil {
		dbC.RmtConns = make([]string, len(*jsnDbConnCfg.Remote_conns))
		for idx, rmtConn := range *jsnDbConnCfg.Remote_conns {
			if rmtConn == utils.MetaInternal {
				return fmt.Errorf("remote connection ID needs to be different than <%s> ", utils.MetaInternal)
			}
			dbC.RmtConns[idx] = rmtConn
		}
	}
	if jsnDbConnCfg.Replication_conns != nil {
		dbC.RplConns = make([]string, len(*jsnDbConnCfg.Replication_conns))
		for idx, rplConn := range *jsnDbConnCfg.Replication_conns {
			if rplConn == utils.MetaInternal {
				return fmt.Errorf("remote connection ID needs to be different than <%s> ", utils.MetaInternal)
			}
			dbC.RplConns[idx] = rplConn
		}
	}
	if jsnDbConnCfg.Replication_filtered != nil {
		dbC.RplFiltered = *jsnDbConnCfg.Replication_filtered
	}
	if jsnDbConnCfg.Remote_conn_id != nil {
		dbC.RmtConnID = *jsnDbConnCfg.Remote_conn_id
	}
	if jsnDbConnCfg.Replication_cache != nil {
		dbC.RplCache = *jsnDbConnCfg.Replication_cache
	}
	if jsnDbConnCfg.Opts != nil {
		err = dbC.Opts.loadFromJSONCfg(jsnDbConnCfg.Opts)
	}
	return
}

// DBConns the config for all DataDB connections
type DBConns map[string]*DBConn

// DbCfg Database config
type DbCfg struct {
	DBConns DBConns
	Items   map[string]*ItemOpts
}

// loadDataDBCfg loads the DataDB section of the configuration
func (dbcfg *DbCfg) Load(ctx *context.Context, jsnCfg ConfigDB, cfg *CGRConfig) (err error) {
	jsnDbCfg := new(DbJsonCfg)
	if err = jsnCfg.GetSection(ctx, DBJSON, jsnDbCfg); err != nil {
		return
	}
	if err = dbcfg.loadFromJSONCfg(jsnDbCfg); err != nil {
		return
	}
	// in case of internalDB we need to disable the cache
	// so we enforce it here
	for _, dbCfg := range cfg.dbCfg.DBConns {
		if dbCfg.Type != utils.MetaInternal {
			continue
		}
		// overwrite only StatelessDataDBPartitions and leave other unmodified ( e.g. *diameter_messages, *closed_sessions, etc... )
		for key := range utils.StatelessDataDBPartitions {
			if _, has := cfg.cacheCfg.Partitions[key]; has {
				cfg.cacheCfg.Partitions[key] = &CacheParamCfg{}
			}
		}
		return // there is only 1 internalDB, stop searching for more
	}
	return
}

func (dbOpts *DBOpts) loadFromJSONCfg(jsnCfg *DBOptsJson) (err error) {
	if jsnCfg == nil {
		return
	}
	if jsnCfg.InternalDBDumpPath != nil {
		dbOpts.InternalDBDumpPath = *jsnCfg.InternalDBDumpPath
	}
	if jsnCfg.InternalDBBackupPath != nil {
		dbOpts.InternalDBBackupPath = *jsnCfg.InternalDBBackupPath
	}
	if jsnCfg.InternalDBStartTimeout != nil {
		if dbOpts.InternalDBStartTimeout, err = utils.ParseDurationWithNanosecs(*jsnCfg.InternalDBStartTimeout); err != nil {
			return err
		}
	}
	if jsnCfg.InternalDBDumpInterval != nil {
		if dbOpts.InternalDBDumpInterval, err = utils.ParseDurationWithNanosecs(*jsnCfg.InternalDBDumpInterval); err != nil {
			return err
		}
	}
	if jsnCfg.InternalDBRewriteInterval != nil {
		if dbOpts.InternalDBRewriteInterval, err = utils.ParseDurationWithNanosecs(*jsnCfg.InternalDBRewriteInterval); err != nil {
			return err
		}
	}
	if jsnCfg.InternalDBFileSizeLimit != nil {
		if dbOpts.InternalDBFileSizeLimit, err = utils.ParseBinarySize(*jsnCfg.InternalDBFileSizeLimit); err != nil {
			return err
		}
	}
	if jsnCfg.RedisMaxConns != nil {
		dbOpts.RedisMaxConns = *jsnCfg.RedisMaxConns
	}
	if jsnCfg.RedisConnectAttempts != nil {
		dbOpts.RedisConnectAttempts = *jsnCfg.RedisConnectAttempts
	}
	if jsnCfg.RedisSentinel != nil {
		dbOpts.RedisSentinel = *jsnCfg.RedisSentinel
	}
	if jsnCfg.RedisCluster != nil {
		dbOpts.RedisCluster = *jsnCfg.RedisCluster
	}
	if jsnCfg.RedisClusterSync != nil {
		if dbOpts.RedisClusterSync, err = utils.ParseDurationWithNanosecs(*jsnCfg.RedisClusterSync); err != nil {
			return
		}
	}
	if jsnCfg.RedisClusterOndownDelay != nil {
		if dbOpts.RedisClusterOndownDelay, err = utils.ParseDurationWithNanosecs(*jsnCfg.RedisClusterOndownDelay); err != nil {
			return
		}
	}
	if jsnCfg.RedisConnectTimeout != nil {
		if dbOpts.RedisConnectTimeout, err = utils.ParseDurationWithNanosecs(*jsnCfg.RedisConnectTimeout); err != nil {
			return
		}
	}
	if jsnCfg.RedisReadTimeout != nil {
		if dbOpts.RedisReadTimeout, err = utils.ParseDurationWithNanosecs(*jsnCfg.RedisReadTimeout); err != nil {
			return
		}
	}
	if jsnCfg.RedisWriteTimeout != nil {
		if dbOpts.RedisWriteTimeout, err = utils.ParseDurationWithNanosecs(*jsnCfg.RedisWriteTimeout); err != nil {
			return
		}
	}
	if jsnCfg.RedisPoolPipelineWindow != nil {
		if dbOpts.RedisPoolPipelineWindow, err = utils.ParseDurationWithNanosecs(*jsnCfg.RedisPoolPipelineWindow); err != nil {
			return
		}
	}
	if jsnCfg.RedisPoolPipelineLimit != nil {
		dbOpts.RedisPoolPipelineLimit = *jsnCfg.RedisPoolPipelineLimit
	}
	if jsnCfg.RedisTLS != nil {
		dbOpts.RedisTLS = *jsnCfg.RedisTLS
	}
	if jsnCfg.RedisClientCertificate != nil {
		dbOpts.RedisClientCertificate = *jsnCfg.RedisClientCertificate
	}
	if jsnCfg.RedisClientKey != nil {
		dbOpts.RedisClientKey = *jsnCfg.RedisClientKey
	}
	if jsnCfg.RedisCACertificate != nil {
		dbOpts.RedisCACertificate = *jsnCfg.RedisCACertificate
	}
	if jsnCfg.MongoQueryTimeout != nil {
		if dbOpts.MongoQueryTimeout, err = utils.ParseDurationWithNanosecs(*jsnCfg.MongoQueryTimeout); err != nil {
			return
		}
	}
	if jsnCfg.MongoConnScheme != nil {
		dbOpts.MongoConnScheme = *jsnCfg.MongoConnScheme
	}
	if jsnCfg.SQLMaxOpenConns != nil {
		dbOpts.SQLMaxOpenConns = *jsnCfg.SQLMaxOpenConns
	}
	if jsnCfg.SQLMaxIdleConns != nil {
		dbOpts.SQLMaxIdleConns = *jsnCfg.SQLMaxIdleConns
	}
	if jsnCfg.SQLLogLevel != nil {
		dbOpts.SQLLogLevel = *jsnCfg.SQLLogLevel
	}
	if jsnCfg.SQLConnMaxLifetime != nil {
		if dbOpts.SQLConnMaxLifetime, err = utils.ParseDurationWithNanosecs(*jsnCfg.SQLConnMaxLifetime); err != nil {
			return
		}
	}
	if jsnCfg.MYSQLDSNParams != nil {
		dbOpts.SQLDSNParams = make(map[string]string)
		dbOpts.SQLDSNParams = jsnCfg.MYSQLDSNParams
	}
	if jsnCfg.PgSSLMode != nil {
		dbOpts.PgSSLMode = *jsnCfg.PgSSLMode
	}
	if jsnCfg.PgSSLCert != nil {
		dbOpts.PgSSLCert = *jsnCfg.PgSSLCert
	}
	if jsnCfg.PgSSLKey != nil {
		dbOpts.PgSSLKey = *jsnCfg.PgSSLKey
	}
	if jsnCfg.PgSSLPassword != nil {
		dbOpts.PgSSLPassword = *jsnCfg.PgSSLPassword
	}
	if jsnCfg.PgSSLCertMode != nil {
		dbOpts.PgSSLCertMode = *jsnCfg.PgSSLCertMode
	}
	if jsnCfg.PgSSLRootCert != nil {
		dbOpts.PgSSLRootCert = *jsnCfg.PgSSLRootCert
	}
	if jsnCfg.MySQLLocation != nil {
		dbOpts.MySQLLocation = *jsnCfg.MySQLLocation
	}
	return
}

// loadFromJSONCfg loads Database config from JsonCfg
func (dbcfg *DbCfg) loadFromJSONCfg(jsnDbCfg *DbJsonCfg) (err error) {
	if jsnDbCfg == nil {
		return nil
	}
	if jsnDbCfg.Db_conns != nil {
		// hardcoded *default connection to internal, can be overwritten
		if _, exists := dbcfg.DBConns[utils.MetaDefault]; !exists {
			dbcfg.DBConns[utils.MetaDefault] = &DBConn{
				Type: utils.MetaInternal,
				Opts: &DBOpts{},
			}
		}
		for kJsn, vJsn := range jsnDbCfg.Db_conns {
			if _, exists := dbcfg.DBConns[kJsn]; !exists {
				dbcfg.DBConns[kJsn] = &DBConn{
					Opts: &DBOpts{},
				}
			}
			if err = dbcfg.DBConns[kJsn].loadFromJSONCfg(vJsn); err != nil {
				return err
			}
		}
	}
	if jsnDbCfg.Items != nil {
		for kJsn, vJsn := range jsnDbCfg.Items {
			val, has := dbcfg.Items[kJsn]
			if val == nil || !has {
				val = &ItemOpts{Limit: -1}
			}
			if err = val.loadFromJSONCfg(vJsn); err != nil {
				return
			}
			dbcfg.Items[kJsn] = val
		}
	}

	return
}

func (DbCfg) SName() string               { return DBJSON }
func (dbcfg DbCfg) CloneSection() Section { return dbcfg.Clone() }

func (dbOpts *DBOpts) Clone() *DBOpts {
	if dbOpts == nil {
		dbOpts = &DBOpts{}
	}
	return &DBOpts{
		InternalDBDumpPath:        dbOpts.InternalDBDumpPath,
		InternalDBBackupPath:      dbOpts.InternalDBBackupPath,
		InternalDBStartTimeout:    dbOpts.InternalDBStartTimeout,
		InternalDBDumpInterval:    dbOpts.InternalDBDumpInterval,
		InternalDBRewriteInterval: dbOpts.InternalDBRewriteInterval,
		InternalDBFileSizeLimit:   dbOpts.InternalDBFileSizeLimit,
		RedisMaxConns:             dbOpts.RedisMaxConns,
		RedisConnectAttempts:      dbOpts.RedisConnectAttempts,
		RedisSentinel:             dbOpts.RedisSentinel,
		RedisCluster:              dbOpts.RedisCluster,
		RedisClusterSync:          dbOpts.RedisClusterSync,
		RedisClusterOndownDelay:   dbOpts.RedisClusterOndownDelay,
		RedisConnectTimeout:       dbOpts.RedisConnectTimeout,
		RedisReadTimeout:          dbOpts.RedisReadTimeout,
		RedisWriteTimeout:         dbOpts.RedisWriteTimeout,
		RedisPoolPipelineWindow:   dbOpts.RedisPoolPipelineWindow,
		RedisPoolPipelineLimit:    dbOpts.RedisPoolPipelineLimit,
		RedisTLS:                  dbOpts.RedisTLS,
		RedisClientCertificate:    dbOpts.RedisClientCertificate,
		RedisClientKey:            dbOpts.RedisClientKey,
		RedisCACertificate:        dbOpts.RedisCACertificate,
		MongoQueryTimeout:         dbOpts.MongoQueryTimeout,
		MongoConnScheme:           dbOpts.MongoConnScheme,
		SQLMaxOpenConns:           dbOpts.SQLMaxOpenConns,
		SQLMaxIdleConns:           dbOpts.SQLMaxIdleConns,
		SQLLogLevel:               dbOpts.SQLLogLevel,
		SQLConnMaxLifetime:        dbOpts.SQLConnMaxLifetime,
		SQLDSNParams:              dbOpts.SQLDSNParams,
		PgSSLMode:                 dbOpts.PgSSLMode,
		PgSSLCert:                 dbOpts.PgSSLCert,
		PgSSLKey:                  dbOpts.PgSSLKey,
		PgSSLPassword:             dbOpts.PgSSLPassword,
		PgSSLCertMode:             dbOpts.PgSSLCertMode,
		PgSSLRootCert:             dbOpts.PgSSLRootCert,
		MySQLLocation:             dbOpts.MySQLLocation,
	}
}

// Clone returns the cloned object
func (dbcfg DbCfg) Clone() (cln *DbCfg) {
	cln = &DbCfg{
		DBConns: make(DBConns),
		Items:   make(map[string]*ItemOpts),
	}
	for k, v := range dbcfg.DBConns {
		cln.DBConns[k] = v.Clone()
	}
	for k, itm := range dbcfg.Items {
		cln.Items[k] = itm.Clone()
	}
	return
}

// Clone returns the cloned object
func (dbC *DBConn) Clone() (cln *DBConn) {
	cln = &DBConn{
		Type:        dbC.Type,
		Host:        dbC.Host,
		Port:        dbC.Port,
		Name:        dbC.Name,
		User:        dbC.User,
		Password:    dbC.Password,
		RplFiltered: dbC.RplFiltered,
		RplCache:    dbC.RplCache,
		RmtConnID:   dbC.RmtConnID,
		Opts:        dbC.Opts.Clone(),
	}
	if dbC.StringIndexedFields != nil {
		cln.StringIndexedFields = slices.Clone(dbC.StringIndexedFields)
	}
	if dbC.PrefixIndexedFields != nil {
		cln.PrefixIndexedFields = slices.Clone(dbC.PrefixIndexedFields)
	}
	if dbC.RmtConns != nil {
		cln.RmtConns = slices.Clone(dbC.RmtConns)
	}
	if dbC.RplConns != nil {
		cln.RplConns = slices.Clone(dbC.RplConns)
	}
	return
}

// AsMapInterface returns the config as a map[string]any
func (dbcfg DbCfg) AsMapInterface() any {

	dbConns := make(map[string]map[string]any)
	for k, dbc := range dbcfg.DBConns {
		opts := map[string]any{
			utils.InternalDBDumpPathCfg:        dbc.Opts.InternalDBDumpPath,
			utils.InternalDBBackupPathCfg:      dbc.Opts.InternalDBBackupPath,
			utils.InternalDBStartTimeoutCfg:    dbc.Opts.InternalDBStartTimeout.String(),
			utils.InternalDBDumpIntervalCfg:    dbc.Opts.InternalDBDumpInterval.String(),
			utils.InternalDBRewriteIntervalCfg: dbc.Opts.InternalDBRewriteInterval.String(),
			utils.InternalDBFileSizeLimitCfg:   dbc.Opts.InternalDBFileSizeLimit,
			utils.RedisMaxConnsCfg:             dbc.Opts.RedisMaxConns,
			utils.RedisConnectAttemptsCfg:      dbc.Opts.RedisConnectAttempts,
			utils.RedisSentinelNameCfg:         dbc.Opts.RedisSentinel,
			utils.RedisClusterCfg:              dbc.Opts.RedisCluster,
			utils.RedisClusterSyncCfg:          dbc.Opts.RedisClusterSync.String(),
			utils.RedisClusterOnDownDelayCfg:   dbc.Opts.RedisClusterOndownDelay.String(),
			utils.RedisConnectTimeoutCfg:       dbc.Opts.RedisConnectTimeout.String(),
			utils.RedisReadTimeoutCfg:          dbc.Opts.RedisReadTimeout.String(),
			utils.RedisWriteTimeoutCfg:         dbc.Opts.RedisWriteTimeout.String(),
			utils.RedisPoolPipelineWindowCfg:   dbc.Opts.RedisPoolPipelineWindow.String(),
			utils.RedisPoolPipelineLimitCfg:    dbc.Opts.RedisPoolPipelineLimit,
			utils.RedisTLSCfg:                  dbc.Opts.RedisTLS,
			utils.RedisClientCertificateCfg:    dbc.Opts.RedisClientCertificate,
			utils.RedisClientKeyCfg:            dbc.Opts.RedisClientKey,
			utils.RedisCACertificateCfg:        dbc.Opts.RedisCACertificate,
			utils.MongoQueryTimeoutCfg:         dbc.Opts.MongoQueryTimeout.String(),
			utils.MongoConnSchemeCfg:           dbc.Opts.MongoConnScheme,
			utils.SQLMaxOpenConnsCfg:           dbc.Opts.SQLMaxOpenConns,
			utils.SQLMaxIdleConnsCfg:           dbc.Opts.SQLMaxIdleConns,
			utils.SQLLogLevelCfg:               dbc.Opts.SQLLogLevel,
			utils.SQLConnMaxLifetime:           dbc.Opts.SQLConnMaxLifetime.String(),
			utils.MYSQLDSNParams:               dbc.Opts.SQLDSNParams,
			utils.PgSSLModeCfg:                 dbc.Opts.PgSSLMode,
			utils.MysqlLocation:                dbc.Opts.MySQLLocation,
		}
		if dbc.Opts.PgSSLCert != "" {
			opts[utils.PgSSLCertCfg] = dbc.Opts.PgSSLCert
		}
		if dbc.Opts.PgSSLKey != "" {
			opts[utils.PgSSLKeyCfg] = dbc.Opts.PgSSLKey
		}
		if dbc.Opts.PgSSLPassword != "" {
			opts[utils.PgSSLPasswordCfg] = dbc.Opts.PgSSLPassword
		}
		if dbc.Opts.PgSSLCertMode != "" {
			opts[utils.PgSSLCertModeCfg] = dbc.Opts.PgSSLCertMode
		}
		if dbc.Opts.PgSSLRootCert != "" {
			opts[utils.PgSSLRootCertCfg] = dbc.Opts.PgSSLRootCert
		}
		dbConns[k] = map[string]any{
			utils.DataDbTypeCfg:          dbc.Type,
			utils.DataDbHostCfg:          dbc.Host,
			utils.DataDbNameCfg:          dbc.Name,
			utils.DataDbUserCfg:          dbc.User,
			utils.DataDbPassCfg:          dbc.Password,
			utils.StringIndexedFieldsCfg: dbc.StringIndexedFields,
			utils.PrefixIndexedFieldsCfg: dbc.PrefixIndexedFields,
			utils.RemoteConnsCfg:         dbc.RmtConns,
			utils.RemoteConnIDCfg:        dbc.RmtConnID,
			utils.ReplicationConnsCfg:    dbc.RplConns,
			utils.ReplicationFilteredCfg: dbc.RplFiltered,
			utils.ReplicationCache:       dbc.RplCache,
			utils.OptsCfg:                opts,
		}
		if dbc.Port != "" {
			dbConns[k][utils.DataDbPortCfg], _ = strconv.Atoi(dbc.Port)
		}
	}
	mp := map[string]any{
		utils.DataDbConnsCfg: dbConns,
	}
	if dbcfg.Items != nil {
		items := make(map[string]any)
		for key, item := range dbcfg.Items {
			items[key] = item.AsMapInterface()
		}
		mp[utils.ItemsCfg] = items
	}
	return mp
}

// ItemOpts the options for the stored items
type ItemOpts struct {
	Limit     int
	TTL       time.Duration
	StaticTTL bool
	Remote    bool
	Replicate bool
	DBConn    string // ID of the DB connection that this item belongs to

	// used for ArgDispatcher in case we send this to a dispatcher engine
	RouteID string
	APIKey  string
}

// AsMapInterface returns the config as a map[string]any
func (iI *ItemOpts) AsMapInterface() (initialMP map[string]any) {
	initialMP = map[string]any{
		utils.RemoteCfg:    iI.Remote,
		utils.ReplicateCfg: iI.Replicate,
		utils.LimitCfg:     iI.Limit,
		utils.StaticTTLCfg: iI.StaticTTL,
		utils.DBConnCfg:    iI.DBConn,
	}
	if iI.APIKey != utils.EmptyString {
		initialMP[utils.APIKeyCfg] = iI.APIKey
	}
	if iI.RouteID != utils.EmptyString {
		initialMP[utils.RouteIDCfg] = iI.RouteID
	}
	if iI.TTL != 0 {
		initialMP[utils.TTLCfg] = iI.TTL.String()
	}
	return
}

func (iI *ItemOpts) loadFromJSONCfg(jsonItm *ItemOptsJson) (err error) {
	if jsonItm == nil {
		return
	}
	if jsonItm.Limit != nil {
		iI.Limit = *jsonItm.Limit
	}
	if jsonItm.Static_ttl != nil {
		iI.StaticTTL = *jsonItm.Static_ttl
	}
	if jsonItm.Remote != nil {
		iI.Remote = *jsonItm.Remote
	}
	if jsonItm.Replicate != nil {
		iI.Replicate = *jsonItm.Replicate
	}
	if jsonItm.DbConn != nil {
		iI.DBConn = *jsonItm.DbConn
	}
	if jsonItm.Route_id != nil {
		iI.RouteID = *jsonItm.Route_id
	}
	if jsonItm.Api_key != nil {
		iI.APIKey = *jsonItm.Api_key
	}
	if jsonItm.Ttl != nil {
		iI.TTL, err = utils.ParseDurationWithNanosecs(*jsonItm.Ttl)
	}
	return
}

// Clone returns a deep copy of ItemOpt
func (iI *ItemOpts) Clone() *ItemOpts {
	return &ItemOpts{
		Limit:     iI.Limit,
		TTL:       iI.TTL,
		StaticTTL: iI.StaticTTL,
		Remote:    iI.Remote,
		Replicate: iI.Replicate,
		DBConn:    iI.DBConn,
		APIKey:    iI.APIKey,
		RouteID:   iI.RouteID,
	}
}

func (iI *ItemOpts) Equals(itm2 *ItemOpts) bool {
	return iI == nil && itm2 == nil ||
		iI != nil && itm2 != nil &&
			iI.Remote == itm2.Remote &&
			iI.Replicate == itm2.Replicate &&
			iI.RouteID == itm2.RouteID &&
			iI.APIKey == itm2.APIKey &&
			iI.Limit == itm2.Limit &&
			iI.TTL == itm2.TTL &&
			iI.DBConn == itm2.DBConn &&
			iI.StaticTTL == itm2.StaticTTL
}

type ItemOptsJson struct {
	Limit      *int
	Ttl        *string
	Static_ttl *bool
	Remote     *bool
	Replicate  *bool
	DbConn     *string
	// used for ArgDispatcher in case we send this to a dispatcher engine
	Route_id *string
	Api_key  *string
}

func diffItemOptJson(d *ItemOptsJson, v1, v2 *ItemOpts) *ItemOptsJson {
	if d == nil {
		d = new(ItemOptsJson)
	}
	if v2.Remote != v1.Remote {
		d.Remote = utils.BoolPointer(v2.Remote)
	}
	if v2.Replicate != v1.Replicate {
		d.Replicate = utils.BoolPointer(v2.Replicate)
	}
	if v2.Limit != v1.Limit {
		d.Limit = utils.IntPointer(v2.Limit)
	}
	if v2.StaticTTL != v1.StaticTTL {
		d.Static_ttl = utils.BoolPointer(v2.StaticTTL)
	}
	if v2.TTL != v1.TTL {
		d.Ttl = utils.StringPointer(v2.TTL.String())
	}
	if v2.DBConn != v1.DBConn {
		d.DbConn = utils.StringPointer(v2.DBConn)
	}
	if v2.RouteID != v1.RouteID {
		d.Route_id = utils.StringPointer(v2.RouteID)
	}
	if v2.APIKey != v1.APIKey {
		d.Api_key = utils.StringPointer(v2.APIKey)
	}
	return d
}

func diffMapItemOptJson(d map[string]*ItemOptsJson, v1, v2 map[string]*ItemOpts) map[string]*ItemOptsJson {
	if d == nil {
		d = make(map[string]*ItemOptsJson)
	}
	for k, val2 := range v2 {
		if val1, has := v1[k]; !has {
			d[k] = diffItemOptJson(d[k], new(ItemOpts), val2)
		} else if !val1.Equals(val2) {
			d[k] = diffItemOptJson(d[k], val1, val2)
		}
	}
	return d
}

type DBOptsJson struct {
	InternalDBDumpPath        *string           `json:"internalDBDumpPath"`
	InternalDBBackupPath      *string           `json:"internalDBBackupPath"`
	InternalDBStartTimeout    *string           `json:"internalDBStartTimeout"`
	InternalDBDumpInterval    *string           `json:"internalDBDumpInterval"`
	InternalDBRewriteInterval *string           `json:"internalDBRewriteInterval"`
	InternalDBFileSizeLimit   *string           `json:"internalDBFileSizeLimit"`
	RedisMaxConns             *int              `json:"redisMaxConns"`
	RedisConnectAttempts      *int              `json:"redisConnectAttempts"`
	RedisSentinel             *string           `json:"redisSentinel"`
	RedisCluster              *bool             `json:"redisCluster"`
	RedisClusterSync          *string           `json:"redisClusterSync"`
	RedisClusterOndownDelay   *string           `json:"redisClusterOndownDelay"`
	RedisConnectTimeout       *string           `json:"redisConnectTimeout"`
	RedisReadTimeout          *string           `json:"redisReadTimeout"`
	RedisWriteTimeout         *string           `json:"redisWriteTimeout"`
	RedisPoolPipelineWindow   *string           `json:"redisPoolPipelineWindow"`
	RedisPoolPipelineLimit    *int              `json:"redisPoolPipelineLimit"`
	RedisTLS                  *bool             `json:"redisTLS"`
	RedisClientCertificate    *string           `json:"redisClientCertificate"`
	RedisClientKey            *string           `json:"redisClientKey"`
	RedisCACertificate        *string           `json:"redisCACertificate"`
	MongoQueryTimeout         *string           `json:"mongoQueryTimeout"`
	MongoConnScheme           *string           `json:"mongoConnScheme"`
	SQLMaxOpenConns           *int              `json:"sqlMaxOpenConns"`
	SQLMaxIdleConns           *int              `json:"sqlMaxIdleConns"`
	SQLLogLevel               *int              `json:"sqlLogLevel"`
	SQLConnMaxLifetime        *string           `json:"sqlConnMaxLifetime"`
	MYSQLDSNParams            map[string]string `json:"mysqlDSNParams"`
	PgSSLMode                 *string           `json:"pgSSLMode"`
	PgSSLCert                 *string           `json:"pgSSLCert"`
	PgSSLKey                  *string           `json:"pgSSLKey"`
	PgSSLPassword             *string           `json:"pgSSLPassword"`
	PgSSLCertMode             *string           `json:"pgSSLCertMode"`
	PgSSLRootCert             *string           `json:"pgSSLRootCert"`
	MySQLLocation             *string           `json:"mysqlLocation"`
}

type DbConnJson struct {
	Db_type               *string
	Db_host               *string
	Db_port               *int
	Db_name               *string
	Db_user               *string
	Db_password           *string
	String_indexed_fields *[]string
	Prefix_indexed_fields *[]string
	Remote_conns          *[]string
	Remote_conn_id        *string
	Replication_conns     *[]string
	Replication_filtered  *bool
	Replication_cache     *string
	Opts                  *DBOptsJson
}

type DbConnsJson map[string]*DbConnJson

// Database config
type DbJsonCfg struct {
	Db_conns DbConnsJson
	Items    map[string]*ItemOptsJson
}

func diffDataDBOptsJsonCfg(d *DBOptsJson, v1, v2 *DBOpts) *DBOptsJson {
	if d == nil {
		d = new(DBOptsJson)
	}
	if v1.InternalDBDumpPath != v2.InternalDBDumpPath {
		d.InternalDBDumpPath = utils.StringPointer(v2.InternalDBDumpPath)
	}
	if v1.InternalDBBackupPath != v2.InternalDBBackupPath {
		d.InternalDBBackupPath = utils.StringPointer(v2.InternalDBBackupPath)
	}
	if v1.InternalDBStartTimeout != v2.InternalDBStartTimeout {
		d.InternalDBStartTimeout = utils.StringPointer(v2.InternalDBStartTimeout.String())
	}
	if v1.InternalDBDumpInterval != v2.InternalDBDumpInterval {
		d.InternalDBDumpInterval = utils.StringPointer(v2.InternalDBDumpInterval.String())
	}
	if v1.InternalDBRewriteInterval != v2.InternalDBRewriteInterval {
		d.InternalDBRewriteInterval = utils.StringPointer(v2.InternalDBRewriteInterval.String())
	}
	if v1.InternalDBFileSizeLimit != v2.InternalDBFileSizeLimit {
		d.InternalDBFileSizeLimit = utils.StringPointer(fmt.Sprint(v2.InternalDBFileSizeLimit))
	}
	if v1.RedisMaxConns != v2.RedisMaxConns {
		d.RedisMaxConns = utils.IntPointer(v2.RedisMaxConns)
	}
	if v1.RedisConnectAttempts != v2.RedisConnectAttempts {
		d.RedisConnectAttempts = utils.IntPointer(v2.RedisConnectAttempts)
	}
	if v1.RedisSentinel != v2.RedisSentinel {
		d.RedisSentinel = utils.StringPointer(v2.RedisSentinel)
	}
	if v1.RedisCluster != v2.RedisCluster {
		d.RedisCluster = utils.BoolPointer(v2.RedisCluster)
	}
	if v1.RedisClusterSync != v2.RedisClusterSync {
		d.RedisClusterSync = utils.StringPointer(v2.RedisClusterSync.String())
	}
	if v1.RedisClusterOndownDelay != v2.RedisClusterOndownDelay {
		d.RedisClusterOndownDelay = utils.StringPointer(v2.RedisClusterOndownDelay.String())
	}
	if v1.RedisConnectTimeout != v2.RedisConnectTimeout {
		d.RedisConnectTimeout = utils.StringPointer(v2.RedisConnectTimeout.String())
	}
	if v1.RedisReadTimeout != v2.RedisReadTimeout {
		d.RedisReadTimeout = utils.StringPointer(v2.RedisReadTimeout.String())
	}
	if v1.RedisWriteTimeout != v2.RedisWriteTimeout {
		d.RedisWriteTimeout = utils.StringPointer(v2.RedisWriteTimeout.String())
	}
	if v1.RedisPoolPipelineWindow != v2.RedisPoolPipelineWindow {
		d.RedisPoolPipelineWindow = utils.StringPointer(v2.RedisPoolPipelineWindow.String())
	}
	if v1.RedisPoolPipelineLimit != v2.RedisPoolPipelineLimit {
		d.RedisPoolPipelineLimit = utils.IntPointer(v2.RedisPoolPipelineLimit)
	}
	if v1.RedisTLS != v2.RedisTLS {
		d.RedisTLS = utils.BoolPointer(v2.RedisTLS)
	}
	if v1.RedisClientCertificate != v2.RedisClientCertificate {
		d.RedisClientCertificate = utils.StringPointer(v2.RedisClientCertificate)
	}
	if v1.RedisClientKey != v2.RedisClientKey {
		d.RedisClientKey = utils.StringPointer(v2.RedisClientKey)
	}
	if v1.RedisCACertificate != v2.RedisCACertificate {
		d.RedisCACertificate = utils.StringPointer(v2.RedisCACertificate)
	}
	if v1.MongoQueryTimeout != v2.MongoQueryTimeout {
		d.MongoQueryTimeout = utils.StringPointer(v2.MongoQueryTimeout.String())
	}
	if v1.MongoConnScheme != v2.MongoConnScheme {
		d.MongoConnScheme = utils.StringPointer(v2.MongoConnScheme)
	}
	if v1.SQLMaxOpenConns != v2.SQLMaxOpenConns {
		d.SQLMaxOpenConns = utils.IntPointer(v2.SQLMaxOpenConns)
	}
	if v1.SQLMaxIdleConns != v2.SQLMaxIdleConns {
		d.SQLMaxIdleConns = utils.IntPointer(v2.SQLMaxIdleConns)
	}
	if v1.SQLLogLevel != v2.SQLLogLevel {
		d.SQLLogLevel = utils.IntPointer(v2.SQLLogLevel)
	}
	if v1.SQLConnMaxLifetime != v2.SQLConnMaxLifetime {
		d.SQLConnMaxLifetime = utils.StringPointer(v2.SQLConnMaxLifetime.String())
	}
	if !reflect.DeepEqual(v1.SQLDSNParams, v2.SQLDSNParams) {
		d.MYSQLDSNParams = v2.SQLDSNParams
	}
	if v1.PgSSLMode != v2.PgSSLMode {
		d.PgSSLMode = utils.StringPointer(v2.PgSSLMode)
	}
	if v1.PgSSLCert != v2.PgSSLCert {
		d.PgSSLCert = utils.StringPointer(v2.PgSSLCert)
	}
	if v1.PgSSLKey != v2.PgSSLKey {
		d.PgSSLKey = utils.StringPointer(v2.PgSSLKey)
	}
	if v1.PgSSLPassword != v2.PgSSLPassword {
		d.PgSSLPassword = utils.StringPointer(v2.PgSSLPassword)
	}
	if v1.PgSSLCertMode != v2.PgSSLCertMode {
		d.PgSSLCertMode = utils.StringPointer(v2.PgSSLCertMode)
	}
	if v1.PgSSLRootCert != v2.PgSSLRootCert {
		d.PgSSLRootCert = utils.StringPointer(v2.PgSSLRootCert)
	}
	if v1.MySQLLocation != v2.MySQLLocation {
		d.MySQLLocation = utils.StringPointer(v2.MySQLLocation)
	}
	return d
}

func diffDataDBConnJsonCfg(d *DbConnJson, v1, v2 *DBConn) *DbConnJson {
	if d == nil {
		d = new(DbConnJson)
	}
	if v1.Type != v2.Type {
		d.Db_type = utils.StringPointer(v2.Type)
	}
	if v1.Host != v2.Host {
		d.Db_host = utils.StringPointer(v2.Host)
	}
	if v1.Port != v2.Port {
		port, _ := strconv.Atoi(v2.Port)
		d.Db_port = utils.IntPointer(port)
	}
	if v1.Name != v2.Name {
		d.Db_name = utils.StringPointer(v2.Name)
	}
	if v1.User != v2.User {
		d.Db_user = utils.StringPointer(v2.User)
	}
	if v1.Password != v2.Password {
		d.Db_password = utils.StringPointer(v2.Password)
	}
	if !slices.Equal(v1.StringIndexedFields, v2.StringIndexedFields) {
		d.String_indexed_fields = &v2.StringIndexedFields
	}
	if !slices.Equal(v1.PrefixIndexedFields, v2.PrefixIndexedFields) {
		d.Prefix_indexed_fields = &v2.PrefixIndexedFields
	}
	if !slices.Equal(v1.RmtConns, v2.RmtConns) {
		d.Remote_conns = &v2.RmtConns
	}
	if v1.RmtConnID != v2.RmtConnID {
		d.Remote_conn_id = utils.StringPointer(v2.RmtConnID)
	}
	if !slices.Equal(v1.RplConns, v2.RplConns) {
		d.Replication_conns = &v2.RplConns
	}
	if v1.RplFiltered != v2.RplFiltered {
		d.Replication_filtered = utils.BoolPointer(v2.RplFiltered)
	}
	if v1.RplCache != v2.RplCache {
		d.Replication_cache = utils.StringPointer(v2.RplCache)
	}
	d.Opts = diffDataDBOptsJsonCfg(d.Opts, v1.Opts, v2.Opts)
	return d
}

func diffDataDBConnsJsonCfg(d DbConnsJson, v1, v2 DBConns) DbConnsJson {
	if d == nil {
		d = make(DbConnsJson)
	}
	for key, val2 := range v2 {
		if val1, has := v1[key]; !has {
			d[key] = diffDataDBConnJsonCfg(d[key], new(DBConn), val2)
		} else if !val1.Equals(val2) {
			d[key] = diffDataDBConnJsonCfg(d[key], val1, val2)
		}
	}
	return d
}

func (dbC *DBConn) Equals(dbC2 *DBConn) bool {
	if dbC2 == nil {
		return false
	}
	if dbC.Type != dbC2.Type ||
		dbC.Host != dbC2.Host ||
		dbC.Port != dbC2.Port ||
		dbC.Name != dbC2.Name ||
		dbC.User != dbC2.User ||
		dbC.Password != dbC2.Password ||
		dbC.RmtConnID != dbC2.RmtConnID ||
		dbC.RplFiltered != dbC2.RplFiltered ||
		dbC.RplCache != dbC2.RplCache {
		return false
	}
	if len(dbC.RmtConns) != len(dbC2.RmtConns) {
		return false
	}
	for i := range dbC.RmtConns {
		if dbC.RmtConns[i] != dbC2.RmtConns[i] {
			return false
		}
	}
	if len(dbC.RplConns) != len(dbC2.RplConns) {
		return false
	}
	for i := range dbC.RplConns {
		if dbC.RplConns[i] != dbC2.RplConns[i] {
			return false
		}
	}
	return true
}

func diffDataDBJsonCfg(d *DbJsonCfg, v1, v2 *DbCfg) *DbJsonCfg {
	if d == nil {
		d = new(DbJsonCfg)
	}
	d.Db_conns = diffDataDBConnsJsonCfg(d.Db_conns, v1.DBConns, v2.DBConns)
	d.Items = diffMapItemOptJson(d.Items, v1.Items, v2.Items)
	return d
}

// ToTransCacheOpts returns to ltcache.TransCacheOpts from DataDBOpts
func (d *DBOpts) ToTransCacheOpts() *ltcache.TransCacheOpts {
	if d == nil {
		return nil
	}
	return &ltcache.TransCacheOpts{
		DumpPath:        d.InternalDBDumpPath,
		BackupPath:      d.InternalDBBackupPath,
		StartTimeout:    d.InternalDBStartTimeout,
		DumpInterval:    d.InternalDBDumpInterval,
		RewriteInterval: d.InternalDBRewriteInterval,
		FileSizeLimit:   d.InternalDBFileSizeLimit,
	}
}
