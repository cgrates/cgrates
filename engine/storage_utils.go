// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

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
	pass, marshaler, sentinelName string,
	itemsCacheCfg map[string]*config.ItemOpt) (d DataDB, err error) {
	switch dbType {
	case utils.MetaRedis:
		var dbNo int
		dbNo, err = strconv.Atoi(name)
		if err != nil {
			utils.Logger.Crit("Redis db name must be an integer!")
			return nil, err
		}
		if port != "" && !strings.Contains(host, ":") {
			host += ":" + port
		}
		d, err = NewRedisStorage(host, dbNo, pass, marshaler, utils.REDIS_MAX_CONNS, sentinelName)
	case utils.MetaMongo:
		d, err = NewMongoStorage(host, port, name, user, pass, marshaler, nil, true)
	case utils.MetaInternal:
		d = NewInternalDB(nil, nil, true, itemsCacheCfg)
	default:
		err = fmt.Errorf("unsupported db_type <%s>", dbType)
	}
	return
}

// NewStorDBConn returns a StorDB(implements Storage interface) based on dbType
func NewStorDBConn(dbType, host, port, name, user, pass, marshaler, sslmode string,
	maxConn, maxIdleConn, connMaxLifetime int,
	stringIndexedFields, prefixIndexedFields []string,
	itemsCacheCfg map[string]*config.ItemOpt) (db StorDB, err error) {
	switch dbType {
	case utils.MetaMongo:
		db, err = NewMongoStorage(host, port, name, user, pass, marshaler, stringIndexedFields, false)
	case utils.MetaPostgres:
		db, err = NewPostgresStorage(host, port, name, user, pass, sslmode, maxConn, maxIdleConn, connMaxLifetime)
	case utils.MetaMySQL:
		db, err = NewMySQLStorage(host, port, name, user, pass, maxConn, maxIdleConn, connMaxLifetime)
	case utils.MetaInternal:
		db = NewInternalDB(stringIndexedFields, prefixIndexedFields, false, itemsCacheCfg)
	default:
		err = fmt.Errorf("unknown db '%s' valid options are [%s, %s, %s, %s]",
			dbType, utils.MetaMySQL, utils.MetaMongo, utils.MetaPostgres, utils.MetaInternal)
	}
	return
}

// composeMongoURI constructs a MongoDB URI from the given parameters:
//   - scheme: only "mongodb" for now.
//   - host: MongoDB server host (e.g., "localhost").
//   - port: MongoDB server port, excluded if "0".
//   - db: Database name, may include additional parameters (e.g., "db?retryWrites=true").
//   - user: Username for auth, omitted if empty.
//   - pass: Password for auth, only if username is set.
//
// TODO: Should probably remove scheme parameter, as only "mongodb" is supported.
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

type AttrCDRSStoreSMCost struct {
	Cost           *SMCost
	CheckDuplicate bool
	*utils.ArgDispatcher
	*utils.TenantArg
}

type ArgsV2CDRSStoreSMCost struct {
	Cost           *V2SMCost
	CheckDuplicate bool
	*utils.ArgDispatcher
	*utils.TenantArg
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
