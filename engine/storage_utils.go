// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package engine

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

// Various helpers to deal with database

// NewDBConn creates a DB connection
func NewDBConn(dbType, host, port, name, user,
	pass, marshaler string, stringIndexedFields, prefixIndexedFields []string,
	opts *config.DBOpts, itmsCfg map[string]*config.ItemOpts) (d DBDriver, err error) {
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
			opts.RedisCACertificate, opts.RedisBatchSize, stringIndexedFields, prefixIndexedFields)
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
		err = fmt.Errorf("unsupported dbType <%s>", dbType)
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

// checkNestedFields checks if there are elements or values nested (e.g *opts.*ratesCost.Cost)
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
