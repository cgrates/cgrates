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

package engine

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

// Various helpers to deal with database

// NewDataDBConn creates a DataDB connection
func NewDataDBConn(dbType, host, port, name, user,
	pass, marshaler string, opts *config.DataDBOpts,
	itmsCfg map[string]*config.ItemOpt) (d DataDB, err error) {
	switch dbType {
	case utils.MetaRedis:
		var dbNo int
		dbNo, err = strconv.Atoi(name)
		if err != nil {
			utils.Logger.Crit("Redis db name must be an integer!")
			return
		}
		if port != "" && !strings.Contains(host, ":") {
			host += ":" + port
		}
		d, err = NewRedisStorage(host, dbNo, user, pass, marshaler,
			opts.RedisMaxConns, opts.RedisConnectAttempts,
			opts.RedisSentinel,
			opts.RedisCluster, opts.RedisClusterSync, opts.RedisClusterOndownDelay,
			opts.RedisConnectTimeout, opts.RedisReadTimeout, opts.RedisWriteTimeout,
			opts.RedisPoolPipelineWindow, opts.RedisPoolPipelineLimit,
			opts.RedisTLS, opts.RedisClientCertificate, opts.RedisClientKey, opts.RedisCACertificate)
	case utils.MetaMongo:
		d, err = NewMongoStorage(opts.MongoConnScheme, host, port, name, user, pass, marshaler, utils.DataDB, nil, opts.MongoQueryTimeout)
	case utils.MetaInternal:
		d, err = NewInternalDB(nil, nil, true, opts.ToTransCacheOpts(), itmsCfg)
	default:
		err = fmt.Errorf("unsupported db_type <%s>", dbType)
	}
	return
}

// NewStorDBConn returns a StorDB(implements Storage interface) based on dbType
func NewStorDBConn(dbType, host, port, name, user, pass, marshaler string,
	stringIndexedFields, prefixIndexedFields []string,
	opts *config.StorDBOpts, itmsCfg map[string]*config.ItemOpt) (db StorDB, err error) {
	switch dbType {
	case utils.MetaMongo:
		db, err = NewMongoStorage(opts.MongoConnScheme, host, port, name, user, pass, marshaler, utils.StorDB, stringIndexedFields, opts.MongoQueryTimeout)
	case utils.MetaPostgres:
		db, err = NewPostgresStorage(host, port, name, user, pass, opts.PgSchema, opts.PgSSLMode,
			opts.PgSSLCert, opts.PgSSLKey, opts.PgSSLPassword, opts.PgSSLCertMode, opts.PgSSLRootCert,
			opts.SQLMaxOpenConns, opts.SQLMaxIdleConns, opts.SQLLogLevel, opts.SQLConnMaxLifetime)
	case utils.MetaMySQL:
		db, err = NewMySQLStorage(host, port, name, user, pass, marshaler, opts.SQLMaxOpenConns, opts.SQLMaxIdleConns, opts.SQLLogLevel,
			opts.SQLConnMaxLifetime, opts.MySQLLocation, opts.MySQLDSNParams)
	case utils.MetaInternal:
		db, err = NewInternalDB(stringIndexedFields, prefixIndexedFields, false, opts.ToTransCacheOpts(), itmsCfg)
	default:
		err = fmt.Errorf("unknown db '%s' valid options are [%s, %s, %s, %s]",
			dbType, utils.MetaMySQL, utils.MetaMongo, utils.MetaPostgres, utils.MetaInternal)
	}
	return
}

// composeMongoURI constructs a MongoDB URI from the given parameters:
//   - scheme: "mongodb" or "mongodb+srv"
//   - host: MongoDB server host (e.g., "localhost").
//   - port: MongoDB server port, excluded if "0".
//   - db: Database name, may include additional parameters (e.g., "db?retryWrites=true").
//   - user: Username for auth, omitted if empty.
//   - pass: Password for auth, only if username is set.
func composeMongoURI(scheme, host, port, db, user, pass string) string {
	uri := scheme + "://"
	if user != "" && pass != "" {
		uri += user + ":" + pass + "@"
	}
	uri += host
	if port != "0" {
		uri += ":" + port
	}
	if db != "" {
		uri += "/" + db
	}
	return uri
}

// SMCost stores one Cost coming from SM
type SMCost struct {
	CGRID       string
	RunID       string
	OriginHost  string
	OriginID    string
	CostSource  string
	Usage       time.Duration
	CostDetails *EventCost
}

// Clone clones SMCost
func (s *SMCost) Clone() *SMCost {
	if s == nil {
		return nil
	}
	clone := &SMCost{
		CGRID:       s.CGRID,
		RunID:       s.RunID,
		OriginHost:  s.OriginHost,
		OriginID:    s.OriginID,
		CostSource:  s.CostSource,
		Usage:       s.Usage,
		CostDetails: s.CostDetails.Clone(),
	}
	return clone
}

// CacheClone returns a clone of SMCost used by ltcache CacheCloner
func (s *SMCost) CacheClone() any {
	return s.Clone()
}

type AttrCDRSStoreSMCost struct {
	Cost           *SMCost
	CheckDuplicate bool
	APIOpts        map[string]any
	Tenant         string
}

type ArgsV2CDRSStoreSMCost struct {
	Cost           *V2SMCost
	CheckDuplicate bool
	APIOpts        map[string]any
	Tenant         string
}

type V2SMCost struct {
	CGRID       string
	RunID       string
	OriginHost  string
	OriginID    string
	CostSource  string
	Usage       time.Duration
	CostDetails *EventCost
}
