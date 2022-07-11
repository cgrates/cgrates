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
		d, err = NewRedisStorage(host, dbNo, user, pass, marshaler, opts.RedisMaxConns, opts.RedisConnectAttempts,
			opts.RedisSentinel, opts.RedisCluster, opts.RedisClusterSync, opts.RedisClusterOndownDelay,
			opts.RedisConnectTimeout, opts.RedisReadTimeout, opts.RedisWriteTimeout, opts.RedisTLS,
			opts.RedisClientCertificate, opts.RedisClientKey, opts.RedisCACertificate)
	case utils.Mongo:
		d, err = NewMongoStorage(host, port, name, user, pass, marshaler, utils.DataDB, nil, opts.MongoQueryTimeout)
	case utils.Internal:
		d = NewInternalDB(nil, nil, true, itmsCfg)
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
	case utils.Mongo:
		db, err = NewMongoStorage(host, port, name, user, pass, marshaler, utils.StorDB, stringIndexedFields, opts.MongoQueryTimeout)
	case utils.Postgres:
		db, err = NewPostgresStorage(host, port, name, user, pass, opts.PgSSLMode,
			opts.SQLMaxOpenConns, opts.SQLMaxIdleConns, opts.SQLConnMaxLifetime)
	case utils.MySQL:
		db, err = NewMySQLStorage(host, port, name, user, pass, opts.SQLMaxOpenConns, opts.SQLMaxIdleConns,
			opts.SQLConnMaxLifetime, opts.MySQLLocation, opts.MySQLDSNParams)
	case utils.Internal:
		db = NewInternalDB(stringIndexedFields, prefixIndexedFields, false, itmsCfg)
	default:
		err = fmt.Errorf("unknown db '%s' valid options are [%s, %s, %s, %s]",
			dbType, utils.MySQL, utils.Mongo, utils.Postgres, utils.Internal)
	}
	return
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

type AttrCDRSStoreSMCost struct {
	Cost           *SMCost
	CheckDuplicate bool
	APIOpts        map[string]interface{}
	Tenant         string
}

type ArgsV2CDRSStoreSMCost struct {
	Cost           *V2SMCost
	CheckDuplicate bool
	APIOpts        map[string]interface{}
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
