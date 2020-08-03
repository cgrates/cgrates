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
	pass, marshaler, sentinelName string,
	itemsCacheCfg map[string]*config.ItemOpt) (d DataDB, err error) {
	switch dbType {
	case utils.REDIS:
		var dbNo int
		dbNo, err = strconv.Atoi(name)
		if err != nil {
			utils.Logger.Crit("Redis db name must be an integer!")
			return nil, err
		}
		if port != "" && strings.Index(host, ":") == -1 {
			host += ":" + port
		}
		d, err = NewRedisStorage(host, dbNo, user, pass, marshaler, utils.REDIS_MAX_CONNS, sentinelName)
	case utils.MONGO:
		d, err = NewMongoStorage(host, port, name, user, pass, marshaler, utils.DataDB, nil, true)
	case utils.INTERNAL:
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
	case utils.MONGO:
		db, err = NewMongoStorage(host, port, name, user, pass, marshaler, utils.StorDB, stringIndexedFields, false)
	case utils.POSTGRES:
		db, err = NewPostgresStorage(host, port, name, user, pass, sslmode, maxConn, maxIdleConn, connMaxLifetime)
	case utils.MYSQL:
		db, err = NewMySQLStorage(host, port, name, user, pass, maxConn, maxIdleConn, connMaxLifetime)
	case utils.INTERNAL:
		db = NewInternalDB(stringIndexedFields, prefixIndexedFields, false, itemsCacheCfg)
	default:
		err = fmt.Errorf("unknown db '%s' valid options are [%s, %s, %s, %s]",
			dbType, utils.MYSQL, utils.MONGO, utils.POSTGRES, utils.INTERNAL)
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
	Opts           map[string]interface{}
	*utils.TenantArg
}

type ArgsV2CDRSStoreSMCost struct {
	Cost           *V2SMCost
	CheckDuplicate bool
	Opts           map[string]interface{}
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
