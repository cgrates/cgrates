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

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

// Various helpers to deal with database

// NewDataDBConn creates a DataDB connection
func NewDataDBConn(dbType, host, port, name, user,
	pass, marshaler string, opts *config.DataDBOpts,
	itmsCfg map[string]*config.ItemOpts) (d DataDBDriver, err error) {
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
		d = NewInternalDB(nil, nil, itmsCfg)
	default:
		err = fmt.Errorf("unsupported db_type <%s>", dbType)
	}
	return
}

// NewStorDBConn returns a StorDB(implements Storage interface) based on dbType
func NewStorDBConn(dbType, host, port, name, user, pass, marshaler string,
	stringIndexedFields, prefixIndexedFields []string,
	opts *config.StorDBOpts, itmsCfg map[string]*config.ItemOpts) (db StorDB, err error) {
	switch dbType {
	case utils.MetaMongo:
		db, err = NewMongoStorage(opts.MongoConnScheme, host, port, name, user, pass, marshaler, utils.MetaStorDB, stringIndexedFields, opts.MongoQueryTimeout)
	case utils.MetaPostgres:
		db, err = NewPostgresStorage(host, port, name, user, pass, opts.PgSSLMode,
			opts.SQLMaxOpenConns, opts.SQLMaxIdleConns, opts.SQLConnMaxLifetime)
	case utils.MetaMySQL:
		db, err = NewMySQLStorage(host, port, name, user, pass, opts.SQLMaxOpenConns, opts.SQLMaxIdleConns,
			opts.SQLConnMaxLifetime, opts.MySQLLocation, opts.SQLDSNParams)
	case utils.MetaInternal:
		db = NewInternalDB(stringIndexedFields, prefixIndexedFields, itmsCfg)
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

// checkNestedFields checks if there are elements or values nested (e.g *opts.*rateSCost.Cost)
func checkNestedFields(elem string, values []string) bool {
	if len(strings.Split(elem, utils.NestingSep)) > 2 {
		return true
	}
	for _, val := range values {
		if len(strings.Split(val, utils.NestingSep)) > 2 {
			return true
		}
	}
	return false
}
