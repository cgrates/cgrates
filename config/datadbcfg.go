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
	"strconv"
	"strings"
	"time"

	"github.com/cgrates/cgrates/utils"
)

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
	MongoQueryTimeout       time.Duration
	RedisTLS                bool
	RedisClientCertificate  string
	RedisClientKey          string
	RedisCACertificate      string
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
	Items       map[string]*ItemOpt
	Opts        *DataDBOpts
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
	if jsnCfg.MongoQueryTimeout != nil {
		if dbOpts.MongoQueryTimeout, err = utils.ParseDurationWithNanosecs(*jsnCfg.MongoQueryTimeout); err != nil {
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
	return
}

// loadFromJSONCfg loads Database config from JsonCfg
func (dbcfg *DataDbCfg) loadFromJSONCfg(jsnDbCfg *DbJsonCfg) (err error) {
	if jsnDbCfg == nil {
		return
	}
	if jsnDbCfg.Db_type != nil {
		dbcfg.Type = strings.TrimPrefix(*jsnDbCfg.Db_type, "*")
	}
	if jsnDbCfg.Db_host != nil {
		dbcfg.Host = *jsnDbCfg.Db_host
	}
	if jsnDbCfg.Db_port != nil {
		port := strconv.Itoa(*jsnDbCfg.Db_port)
		if port == "-1" {
			port = utils.MetaDynamic
		}
		dbcfg.Port = dbDefaultsCfg.dbPort(dbcfg.Type, port)
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
			// if we have the connection internal we change the name so we can have internal rpc for each subsystem
			if rmtConn == utils.MetaInternal {
				return fmt.Errorf("Remote connection ID needs to be different than *internal")
			}
			dbcfg.RmtConns[idx] = rmtConn
		}
	}
	if jsnDbCfg.Replication_conns != nil {
		dbcfg.RplConns = make([]string, len(*jsnDbCfg.Replication_conns))
		for idx, rplConn := range *jsnDbCfg.Replication_conns {
			// if we have the connection internal we change the name so we can have internal rpc for each subsystem
			if rplConn == utils.MetaInternal {
				return fmt.Errorf("Replication connection ID needs to be different than *internal")
			}
			dbcfg.RplConns[idx] = rplConn
		}
	}
	if jsnDbCfg.Items != nil {
		for kJsn, vJsn := range *jsnDbCfg.Items {
			val, has := dbcfg.Items[kJsn]
			if val == nil || !has {
				val = &ItemOpt{Limit: -1}
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
		MongoQueryTimeout:       dbOpts.MongoQueryTimeout,
		RedisTLS:                dbOpts.RedisTLS,
		RedisClientCertificate:  dbOpts.RedisClientCertificate,
		RedisClientKey:          dbOpts.RedisClientKey,
		RedisCACertificate:      dbOpts.RedisCACertificate,
	}
}

// Clone returns the cloned object
func (dbcfg *DataDbCfg) Clone() (cln *DataDbCfg) {
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
		Items:       make(map[string]*ItemOpt),
		Opts:        dbcfg.Opts.Clone(),
	}
	for k, itm := range dbcfg.Items {
		cln.Items[k] = itm.Clone()
	}
	if dbcfg.RmtConns != nil {
		cln.RmtConns = make([]string, len(dbcfg.RmtConns))
		for i, conn := range dbcfg.RmtConns {
			cln.RmtConns[i] = conn
		}
	}
	if dbcfg.RplConns != nil {
		cln.RplConns = make([]string, len(dbcfg.RplConns))
		for i, conn := range dbcfg.RplConns {
			cln.RplConns[i] = conn
		}
	}
	return
}

// AsMapInterface returns the config as a map[string]interface{}
func (dbcfg *DataDbCfg) AsMapInterface() (mp map[string]interface{}) {
	opts := map[string]interface{}{
		utils.RedisMaxConnsCfg:           dbcfg.Opts.RedisMaxConns,
		utils.RedisConnectAttemptsCfg:    dbcfg.Opts.RedisConnectAttempts,
		utils.RedisSentinelNameCfg:       dbcfg.Opts.RedisSentinel,
		utils.RedisClusterCfg:            dbcfg.Opts.RedisCluster,
		utils.RedisClusterSyncCfg:        dbcfg.Opts.RedisClusterSync.String(),
		utils.RedisClusterOnDownDelayCfg: dbcfg.Opts.RedisClusterOndownDelay.String(),
		utils.RedisConnectTimeoutCfg:     dbcfg.Opts.RedisConnectTimeout.String(),
		utils.RedisReadTimeoutCfg:        dbcfg.Opts.RedisReadTimeout.String(),
		utils.RedisWriteTimeoutCfg:       dbcfg.Opts.RedisWriteTimeout.String(),
		utils.MongoQueryTimeoutCfg:       dbcfg.Opts.MongoQueryTimeout.String(),
		utils.RedisTLS:                   dbcfg.Opts.RedisTLS,
		utils.RedisClientCertificate:     dbcfg.Opts.RedisClientCertificate,
		utils.RedisClientKey:             dbcfg.Opts.RedisClientKey,
		utils.RedisCACertificate:         dbcfg.Opts.RedisCACertificate,
	}
	mp = map[string]interface{}{
		utils.DataDbTypeCfg:          utils.Meta + dbcfg.Type,
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
		items := make(map[string]interface{})
		for key, item := range dbcfg.Items {
			items[key] = item.AsMapInterface()
		}
		mp[utils.ItemsCfg] = items
	}
	if dbcfg.Port != "" {
		mp[utils.DataDbPortCfg], _ = strconv.Atoi(dbcfg.Port)
	}
	return
}

// ItemOpt the options for the stored items
type ItemOpt struct {
	Limit     int
	TTL       time.Duration
	StaticTTL bool
	Remote    bool
	Replicate bool
	// used for ArgDispatcher in case we send this to a dispatcher engine
	RouteID string
	APIKey  string
}

// AsMapInterface returns the config as a map[string]interface{}
func (itm *ItemOpt) AsMapInterface() (initialMP map[string]interface{}) {
	initialMP = map[string]interface{}{
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

func (itm *ItemOpt) loadFromJSONCfg(jsonItm *ItemOptJson) (err error) {
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
func (itm *ItemOpt) Clone() *ItemOpt {
	return &ItemOpt{
		Limit:     itm.Limit,
		TTL:       itm.TTL,
		StaticTTL: itm.StaticTTL,
		Remote:    itm.Remote,
		Replicate: itm.Replicate,
		APIKey:    itm.APIKey,
		RouteID:   itm.RouteID,
	}
}
