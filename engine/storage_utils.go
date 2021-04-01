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

	"github.com/cgrates/cgrates/utils"
)

// Various helpers to deal with database

// NewDataDBConn creates a DataDB connection
func NewDataDBConn(dbType, host, port, name, user,
	pass, marshaler string, opts map[string]interface{}) (d DataDB, err error) {
	switch dbType {
	case utils.Redis:
		var dbNo int
		dbNo, err = strconv.Atoi(name)
		if err != nil {
			utils.Logger.Crit("Redis db name must be an integer!")
			return
		}
		if port != "" && !strings.Contains(host, ":") {
			host += ":" + port
		}
		var isCluster bool
		if isCluster, err = utils.IfaceAsBool(opts[utils.RedisClusterCfg]); err != nil {
			return
		}
		var clusterSync, clusterOnDownDelay time.Duration
		if clusterSync, err = utils.IfaceAsDuration(opts[utils.RedisClusterSyncCfg]); err != nil {
			return
		}
		if clusterOnDownDelay, err = utils.IfaceAsDuration(opts[utils.RedisClusterOnDownDelayCfg]); err != nil {
			return
		}
		var hasTlsConn bool
		if hasTlsConn, err = utils.IfaceAsBool(opts[utils.RedisTLS]); err != nil {
			return
		}
		d, err = NewRedisStorage(host, dbNo, user, pass, marshaler,
			utils.RedisMaxConns, utils.IfaceAsString(opts[utils.RedisSentinelNameCfg]),
			isCluster, clusterSync, clusterOnDownDelay, hasTlsConn,
			utils.IfaceAsString(opts[utils.RedisClientCertificate]),
			utils.IfaceAsString(opts[utils.RedisClientKey]),
			utils.IfaceAsString(opts[utils.RedisCACertificate]))
	case utils.Mongo:
		var ttl time.Duration
		if ttl, err = utils.IfaceAsDuration(opts[utils.QueryTimeoutCfg]); err != nil {
			return
		}
		d, err = NewMongoStorage(host, port, name, user, pass, marshaler, utils.DataDB, nil, ttl)
	case utils.Internal:
		d = NewInternalDB(nil, nil, true)
	default:
		err = fmt.Errorf("unsupported db_type <%s>", dbType)
	}
	return
}

// NewStorDBConn returns a StorDB(implements Storage interface) based on dbType
func NewStorDBConn(dbType, host, port, name, user, pass, marshaler string,
	stringIndexedFields, prefixIndexedFields []string,
	opts map[string]interface{}) (db StorDB, err error) {
	switch dbType {
	case utils.Mongo:
		var ttl time.Duration
		if ttl, err = utils.IfaceAsDuration(opts[utils.QueryTimeoutCfg]); err != nil {
			return nil, err
		}
		db, err = NewMongoStorage(host, port, name, user, pass, marshaler, utils.StorDB, stringIndexedFields, ttl)
	case utils.Postgres:
		var maxConn, maxIdleConn, connMaxLifetime int64
		if maxConn, err = utils.IfaceAsTInt64(opts[utils.MaxOpenConnsCfg]); err != nil {
			return
		}
		if maxIdleConn, err = utils.IfaceAsTInt64(opts[utils.MaxIdleConnsCfg]); err != nil {
			return
		}
		if connMaxLifetime, err = utils.IfaceAsTInt64(opts[utils.ConnMaxLifetimeCfg]); err != nil {
			return
		}
		db, err = NewPostgresStorage(host, port, name, user, pass, utils.IfaceAsString(opts[utils.SSLModeCfg]),
			int(maxConn), int(maxIdleConn), int(connMaxLifetime))
	case utils.MySQL:
		var maxConn, maxIdleConn, connMaxLifetime int64
		if maxConn, err = utils.IfaceAsTInt64(opts[utils.MaxOpenConnsCfg]); err != nil {
			return
		}
		if maxIdleConn, err = utils.IfaceAsTInt64(opts[utils.MaxIdleConnsCfg]); err != nil {
			return
		}
		if connMaxLifetime, err = utils.IfaceAsTInt64(opts[utils.ConnMaxLifetimeCfg]); err != nil {
			return
		}
		db, err = NewMySQLStorage(host, port, name, user, pass, int(maxConn), int(maxIdleConn),
			int(connMaxLifetime), utils.IfaceAsString(opts[utils.MysqlLocation]))
	case utils.Internal:
		db = NewInternalDB(stringIndexedFields, prefixIndexedFields, false)
	default:
		err = fmt.Errorf("unknown db '%s' valid options are [%s, %s, %s, %s]",
			dbType, utils.MySQL, utils.Mongo, utils.Postgres, utils.Internal)
	}
	return
}
