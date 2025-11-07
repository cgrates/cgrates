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
	pass, marshaler string, stringIndexedFields, prefixIndexedFields []string,
	opts *config.DBOpts, itmsCfg map[string]*config.ItemOpts) (d DataDBDriver, err error) {
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
			opts.RedisMaxConns, opts.RedisConnectAttempts, opts.RedisSentinel,
			opts.RedisCluster, opts.RedisClusterSync, opts.RedisClusterOndownDelay,
			opts.RedisConnectTimeout, opts.RedisReadTimeout, opts.RedisWriteTimeout,
			opts.RedisPoolPipelineWindow, opts.RedisPoolPipelineLimit,
			opts.RedisTLS, opts.RedisClientCertificate, opts.RedisClientKey,
			opts.RedisCACertificate)
	case utils.MetaMongo:
		d, err = NewMongoStorage(opts.MongoConnScheme, host, port, name, user, pass,
			marshaler, stringIndexedFields, opts.MongoQueryTimeout)
	case utils.MetaPostgres:
		d, err = NewPostgresStorage(host, port, name, user, pass, marshaler, opts.PgSSLMode,
			opts.PgSSLCert, opts.PgSSLKey, opts.PgSSLPassword, opts.PgSSLCertMode,
			opts.PgSSLRootCert, opts.SQLMaxOpenConns, opts.SQLMaxIdleConns,
			opts.SQLLogLevel, opts.SQLConnMaxLifetime)
	case utils.MetaMySQL:
		d, err = NewMySQLStorage(host, port, name, user, pass, marshaler, opts.SQLMaxOpenConns,
			opts.SQLMaxIdleConns, opts.SQLLogLevel, opts.SQLConnMaxLifetime,
			opts.MySQLLocation, opts.SQLDSNParams)
	case utils.MetaInternal:
		d, err = NewInternalDB(stringIndexedFields, prefixIndexedFields,
			opts.ToTransCacheOpts(), itmsCfg)
	default:
		err = fmt.Errorf("unsupported db_type <%s>", dbType)
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
