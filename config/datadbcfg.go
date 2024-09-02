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
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/utils"
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

type DataDBOpts struct {
	RedisMaxConns           int
	RedisConnectAttempts    int
	RedisSentinel           string
	RedisCluster            bool
	RedisClusterSync        time.Duration
	RedisClusterOndownDelay time.Duration
	RedisConnectTimeout     time.Duration
	RedisReadTimeout        time.Duration
	RedisWriteTimeout       time.Duration
	RedisTLS                bool
	RedisClientCertificate  string
	RedisClientKey          string
	RedisCACertificate      string
	MongoQueryTimeout       time.Duration
	MongoConnScheme         string
}

// DataDbCfg Database config
type DataDbCfg struct {
	Type        string
	Host        string   // The host to connect to. Values that start with / are for UNIX domain sockets.
	Port        string   // The port to bind to.
	Name        string   // The name of the database to connect to.
	User        string   // The user to sign in as.
	Password    string   // The user's password.
	RmtConns    []string // Remote DataDB  connIDs
	RmtConnID   string
	RplConns    []string // Replication connIDs
	RplFiltered bool
	RplCache    string
	Items       map[string]*ItemOpts
	Opts        *DataDBOpts
}

// loadDataDBCfg loads the DataDB section of the configuration
func (dbcfg *DataDbCfg) Load(ctx *context.Context, jsnCfg ConfigDB, cfg *CGRConfig) (err error) {
	jsnDataDbCfg := new(DbJsonCfg)
	if err = jsnCfg.GetSection(ctx, DataDBJSON, jsnDataDbCfg); err != nil {
		return
	}
	if err = dbcfg.loadFromJSONCfg(jsnDataDbCfg); err != nil {
		return
	}
	// in case of internalDB we need to disable the cache
	// so we enforce it here
	if cfg.dataDbCfg.Type == utils.MetaInternal {
		// overwrite only DataDBPartitions and leave other unmodified ( e.g. *diameter_messages, *closed_sessions, etc... )
		for key := range utils.DataDBPartitions {
			if _, has := cfg.cacheCfg.Partitions[key]; has {
				cfg.cacheCfg.Partitions[key] = &CacheParamCfg{}
			}
		}
	}
	return
}

func (dbOpts *DataDBOpts) loadFromJSONCfg(jsnCfg *DBOptsJson) (err error) {
	if jsnCfg == nil {
		return
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
	return
}

// loadFromJSONCfg loads Database config from JsonCfg
func (dbcfg *DataDbCfg) loadFromJSONCfg(jsnDbCfg *DbJsonCfg) (err error) {
	if jsnDbCfg == nil {
		return nil
	}
	if jsnDbCfg.Db_type != nil {
		if !strings.HasPrefix(*jsnDbCfg.Db_type, "*") {
			dbcfg.Type = fmt.Sprintf("*%v", *jsnDbCfg.Db_type)
		} else {
			dbcfg.Type = *jsnDbCfg.Db_type
		}
	}
	if jsnDbCfg.Db_host != nil {
		dbcfg.Host = *jsnDbCfg.Db_host
	}
	if jsnDbCfg.Db_port != nil {
		port := strconv.Itoa(*jsnDbCfg.Db_port)
		if port == "-1" {
			port = utils.MetaDynamic
		}
		dbcfg.Port = defaultDBPort(dbcfg.Type, port)
	}
	if jsnDbCfg.Db_name != nil {
		dbcfg.Name = *jsnDbCfg.Db_name
	}
	if jsnDbCfg.Db_user != nil {
		dbcfg.User = *jsnDbCfg.Db_user
	}
	if jsnDbCfg.Db_password != nil {
		dbcfg.Password = *jsnDbCfg.Db_password
	}
	if jsnDbCfg.Remote_conns != nil {
		dbcfg.RmtConns = make([]string, len(*jsnDbCfg.Remote_conns))
		for idx, rmtConn := range *jsnDbCfg.Remote_conns {
			if rmtConn == utils.MetaInternal {
				return fmt.Errorf("Remote connection ID needs to be different than <%s> ", utils.MetaInternal)
			}
			dbcfg.RmtConns[idx] = rmtConn
		}
	}
	if jsnDbCfg.Replication_conns != nil {
		dbcfg.RplConns = make([]string, len(*jsnDbCfg.Replication_conns))
		for idx, rplConn := range *jsnDbCfg.Replication_conns {
			if rplConn == utils.MetaInternal {
				return fmt.Errorf("Remote connection ID needs to be different than <%s> ", utils.MetaInternal)
			}
			dbcfg.RplConns[idx] = rplConn
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
	if jsnDbCfg.Replication_filtered != nil {
		dbcfg.RplFiltered = *jsnDbCfg.Replication_filtered
	}
	if jsnDbCfg.Remote_conn_id != nil {
		dbcfg.RmtConnID = *jsnDbCfg.Remote_conn_id
	}
	if jsnDbCfg.Replication_cache != nil {
		dbcfg.RplCache = *jsnDbCfg.Replication_cache
	}
	if jsnDbCfg.Opts != nil {
		err = dbcfg.Opts.loadFromJSONCfg(jsnDbCfg.Opts)
	}
	return
}

func (DataDbCfg) SName() string               { return DataDBJSON }
func (dbcfg DataDbCfg) CloneSection() Section { return dbcfg.Clone() }

func (dbOpts *DataDBOpts) Clone() *DataDBOpts {
	return &DataDBOpts{
		RedisMaxConns:           dbOpts.RedisMaxConns,
		RedisConnectAttempts:    dbOpts.RedisConnectAttempts,
		RedisSentinel:           dbOpts.RedisSentinel,
		RedisCluster:            dbOpts.RedisCluster,
		RedisClusterSync:        dbOpts.RedisClusterSync,
		RedisClusterOndownDelay: dbOpts.RedisClusterOndownDelay,
		RedisConnectTimeout:     dbOpts.RedisConnectTimeout,
		RedisReadTimeout:        dbOpts.RedisReadTimeout,
		RedisWriteTimeout:       dbOpts.RedisWriteTimeout,
		RedisTLS:                dbOpts.RedisTLS,
		RedisClientCertificate:  dbOpts.RedisClientCertificate,
		RedisClientKey:          dbOpts.RedisClientKey,
		RedisCACertificate:      dbOpts.RedisCACertificate,
		MongoQueryTimeout:       dbOpts.MongoQueryTimeout,
		MongoConnScheme:         dbOpts.MongoConnScheme,
	}
}

// Clone returns the cloned object
func (dbcfg DataDbCfg) Clone() (cln *DataDbCfg) {
	cln = &DataDbCfg{
		Type:        dbcfg.Type,
		Host:        dbcfg.Host,
		Port:        dbcfg.Port,
		Name:        dbcfg.Name,
		User:        dbcfg.User,
		Password:    dbcfg.Password,
		RplFiltered: dbcfg.RplFiltered,
		RplCache:    dbcfg.RplCache,
		RmtConnID:   dbcfg.RmtConnID,
		Items:       make(map[string]*ItemOpts),
		Opts:        dbcfg.Opts.Clone(),
	}
	for k, itm := range dbcfg.Items {
		cln.Items[k] = itm.Clone()
	}
	if dbcfg.RmtConns != nil {
		cln.RmtConns = slices.Clone(dbcfg.RmtConns)
	}
	if dbcfg.RplConns != nil {
		cln.RplConns = slices.Clone(dbcfg.RplConns)
	}
	return
}

// AsMapInterface returns the config as a map[string]any
func (dbcfg DataDbCfg) AsMapInterface(string) any {
	opts := map[string]any{
		utils.RedisMaxConnsCfg:           dbcfg.Opts.RedisMaxConns,
		utils.RedisConnectAttemptsCfg:    dbcfg.Opts.RedisConnectAttempts,
		utils.RedisSentinelNameCfg:       dbcfg.Opts.RedisSentinel,
		utils.RedisClusterCfg:            dbcfg.Opts.RedisCluster,
		utils.RedisClusterSyncCfg:        dbcfg.Opts.RedisClusterSync.String(),
		utils.RedisClusterOnDownDelayCfg: dbcfg.Opts.RedisClusterOndownDelay.String(),
		utils.RedisConnectTimeoutCfg:     dbcfg.Opts.RedisConnectTimeout.String(),
		utils.RedisReadTimeoutCfg:        dbcfg.Opts.RedisReadTimeout.String(),
		utils.RedisWriteTimeoutCfg:       dbcfg.Opts.RedisWriteTimeout.String(),
		utils.RedisTLSCfg:                dbcfg.Opts.RedisTLS,
		utils.RedisClientCertificateCfg:  dbcfg.Opts.RedisClientCertificate,
		utils.RedisClientKeyCfg:          dbcfg.Opts.RedisClientKey,
		utils.RedisCACertificateCfg:      dbcfg.Opts.RedisCACertificate,
		utils.MongoQueryTimeoutCfg:       dbcfg.Opts.MongoQueryTimeout.String(),
		utils.MongoConnSchemeCfg:         dbcfg.Opts.MongoConnScheme,
	}
	mp := map[string]any{
		utils.DataDbTypeCfg:          dbcfg.Type,
		utils.DataDbHostCfg:          dbcfg.Host,
		utils.DataDbNameCfg:          dbcfg.Name,
		utils.DataDbUserCfg:          dbcfg.User,
		utils.DataDbPassCfg:          dbcfg.Password,
		utils.RemoteConnsCfg:         dbcfg.RmtConns,
		utils.RemoteConnIDCfg:        dbcfg.RmtConnID,
		utils.ReplicationConnsCfg:    dbcfg.RplConns,
		utils.ReplicationFilteredCfg: dbcfg.RplFiltered,
		utils.ReplicationCache:       dbcfg.RplCache,
		utils.OptsCfg:                opts,
	}
	if dbcfg.Items != nil {
		items := make(map[string]any)
		for key, item := range dbcfg.Items {
			items[key] = item.AsMapInterface()
		}
		mp[utils.ItemsCfg] = items
	}
	if dbcfg.Port != "" {
		mp[utils.DataDbPortCfg], _ = strconv.Atoi(dbcfg.Port)
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
	// used for ArgDispatcher in case we send this to a dispatcher engine
	RouteID string
	APIKey  string
}

// AsMapInterface returns the config as a map[string]any
func (itm *ItemOpts) AsMapInterface() (initialMP map[string]any) {
	initialMP = map[string]any{
		utils.RemoteCfg:    itm.Remote,
		utils.ReplicateCfg: itm.Replicate,
		utils.LimitCfg:     itm.Limit,
		utils.StaticTTLCfg: itm.StaticTTL,
	}
	if itm.APIKey != utils.EmptyString {
		initialMP[utils.APIKeyCfg] = itm.APIKey
	}
	if itm.RouteID != utils.EmptyString {
		initialMP[utils.RouteIDCfg] = itm.RouteID
	}
	if itm.TTL != 0 {
		initialMP[utils.TTLCfg] = itm.TTL.String()
	}
	return
}

func (itm *ItemOpts) loadFromJSONCfg(jsonItm *ItemOptsJson) (err error) {
	if jsonItm == nil {
		return
	}
	if jsonItm.Limit != nil {
		itm.Limit = *jsonItm.Limit
	}
	if jsonItm.Static_ttl != nil {
		itm.StaticTTL = *jsonItm.Static_ttl
	}
	if jsonItm.Remote != nil {
		itm.Remote = *jsonItm.Remote
	}
	if jsonItm.Replicate != nil {
		itm.Replicate = *jsonItm.Replicate
	}
	if jsonItm.Route_id != nil {
		itm.RouteID = *jsonItm.Route_id
	}
	if jsonItm.Api_key != nil {
		itm.APIKey = *jsonItm.Api_key
	}
	if jsonItm.Ttl != nil {
		itm.TTL, err = utils.ParseDurationWithNanosecs(*jsonItm.Ttl)
	}
	return
}

// Clone returns a deep copy of ItemOpt
func (itm *ItemOpts) Clone() *ItemOpts {
	return &ItemOpts{
		Limit:     itm.Limit,
		TTL:       itm.TTL,
		StaticTTL: itm.StaticTTL,
		Remote:    itm.Remote,
		Replicate: itm.Replicate,
		APIKey:    itm.APIKey,
		RouteID:   itm.RouteID,
	}
}

func (itm *ItemOpts) Equals(itm2 *ItemOpts) bool {
	return itm == nil && itm2 == nil ||
		itm != nil && itm2 != nil &&
			itm.Remote == itm2.Remote &&
			itm.Replicate == itm2.Replicate &&
			itm.RouteID == itm2.RouteID &&
			itm.APIKey == itm2.APIKey &&
			itm.Limit == itm2.Limit &&
			itm.TTL == itm2.TTL &&
			itm.StaticTTL == itm2.StaticTTL
}

type ItemOptsJson struct {
	Limit      *int
	Ttl        *string
	Static_ttl *bool
	Remote     *bool
	Replicate  *bool
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
	RedisMaxConns           *int              `json:"redisMaxConns"`
	RedisConnectAttempts    *int              `json:"redisConnectAttempts"`
	RedisSentinel           *string           `json:"redisSentinel"`
	RedisCluster            *bool             `json:"redisCluster"`
	RedisClusterSync        *string           `json:"redisClusterSync"`
	RedisClusterOndownDelay *string           `json:"redisClusterOndownDelay"`
	RedisConnectTimeout     *string           `json:"redisConnectTimeout"`
	RedisReadTimeout        *string           `json:"redisReadTimeout"`
	RedisWriteTimeout       *string           `json:"redisWriteTimeout"`
	RedisTLS                *bool             `json:"redisTLS"`
	RedisClientCertificate  *string           `json:"redisClientCertificate"`
	RedisClientKey          *string           `json:"redisClientKey"`
	RedisCACertificate      *string           `json:"redisCACertificate"`
	MongoQueryTimeout       *string           `json:"mongoQueryTimeout"`
	MongoConnScheme         *string           `json:"mongoConnScheme"`
	SQLMaxOpenConns         *int              `json:"sqlMaxOpenConns"`
	SQLMaxIdleConns         *int              `json:"sqlMaxIdleConns"`
	SQLConnMaxLifetime      *string           `json:"sqlConnMaxLifetime"`
	MYSQLDSNParams          map[string]string `json:"mysqlDSNParams"`
	PgSSLMode               *string           `json:"pgSSLMode"`
	MySQLLocation           *string           `json:"mysqlLocation"`
}

// Database config
type DbJsonCfg struct {
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
	Items                 map[string]*ItemOptsJson
	Opts                  *DBOptsJson
}

func diffDataDBOptsJsonCfg(d *DBOptsJson, v1, v2 *DataDBOpts) *DBOptsJson {
	if d == nil {
		d = new(DBOptsJson)
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
	return d
}

func diffDataDBJsonCfg(d *DbJsonCfg, v1, v2 *DataDbCfg) *DbJsonCfg {
	if d == nil {
		d = new(DbJsonCfg)
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
	d.Items = diffMapItemOptJson(d.Items, v1.Items, v2.Items)
	d.Opts = diffDataDBOptsJsonCfg(d.Opts, v1.Opts, v2.Opts)
	return d
}
