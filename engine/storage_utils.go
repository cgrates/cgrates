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

// ConfigureDataStorage returns the DataManager using the given config
func ConfigureDataStorage(dbType, host, port, name, user, pass, marshaler string,
	cacheCfg config.CacheCfg, sentinelName string) (dm *DataManager, err error) {
	var d DataDB
	switch dbType {
	case utils.REDIS:
		var dbNb int
		dbNb, err = strconv.Atoi(name)
		if err != nil {
			utils.Logger.Crit("Redis db name must be an integer!")
			return nil, err
		}
		if port != "" && strings.Index(host, ":") == -1 {
			host += ":" + port
		}
		d, err = NewRedisStorage(host, dbNb, pass, marshaler, utils.REDIS_MAX_CONNS, sentinelName)
	case utils.MONGO:
		d, err = NewMongoStorage(host, port, name, user, pass, utils.DataDB, nil, true)
	case utils.INTERNAL:
		if marshaler == utils.JSON {
			d = NewInternalDBJson(nil, nil)
		} else {
			d = NewInternalDB(nil, nil)
		}
	default:
		err = fmt.Errorf("unknown db '%s' valid options are '%s' or '%s or '%s'",
			dbType, utils.REDIS, utils.MONGO, utils.INTERNAL)
	}
	if err != nil {
		return nil, err
	}
	return NewDataManager(d, cacheCfg), nil
}

func ConfigureStorStorage(db_type, host, port, name, user, pass, marshaler string,
	maxConn, maxIdleConn, connMaxLifetime int,
	stringIndexedFields, prefixIndexedFields []string) (db Storage, err error) {
	var d Storage
	switch db_type {
	case utils.MONGO:
		d, err = NewMongoStorage(host, port, name, user, pass, utils.StorDB, stringIndexedFields, false)
	case utils.POSTGRES:
		d, err = NewPostgresStorage(host, port, name, user, pass, maxConn, maxIdleConn, connMaxLifetime)
	case utils.MYSQL:
		d, err = NewMySQLStorage(host, port, name, user, pass, maxConn, maxIdleConn, connMaxLifetime)
	case utils.INTERNAL:
		d = NewInternalDB(stringIndexedFields, prefixIndexedFields)
	default:
		err = fmt.Errorf("unknown db '%s' valid options are [%s, %s, %s, %s]",
			db_type, utils.MYSQL, utils.MONGO, utils.POSTGRES, utils.INTERNAL)
	}
	if err != nil {
		return nil, err
	}
	return d, nil
}

func ConfigureLoadStorage(db_type, host, port, name, user, pass, marshaler string,
	maxConn, maxIdleConn, connMaxLifetime int,
	stringIndexedFields, prefixIndexedFields []string) (db LoadStorage, err error) {
	var d LoadStorage
	switch db_type {
	case utils.POSTGRES:
		d, err = NewPostgresStorage(host, port, name, user, pass, maxConn, maxIdleConn, connMaxLifetime)
	case utils.MYSQL:
		d, err = NewMySQLStorage(host, port, name, user, pass, maxConn, maxIdleConn, connMaxLifetime)
	case utils.MONGO:
		d, err = NewMongoStorage(host, port, name, user, pass, utils.StorDB, stringIndexedFields, false)
	case utils.INTERNAL:
		d = NewInternalDB(stringIndexedFields, prefixIndexedFields)
	default:
		err = fmt.Errorf("unknown db '%s' valid options are [%s, %s, %s, %s]",
			db_type, utils.MYSQL, utils.MONGO, utils.POSTGRES, utils.INTERNAL)
	}
	if err != nil {
		return nil, err
	}
	return d, nil
}

func ConfigureCdrStorage(db_type, host, port, name, user, pass string,
	maxConn, maxIdleConn, connMaxLifetime int,
	stringIndexedFields, prefixIndexedFields []string) (db CdrStorage, err error) {
	var d CdrStorage
	switch db_type {
	case utils.POSTGRES:
		d, err = NewPostgresStorage(host, port, name, user, pass, maxConn, maxIdleConn, connMaxLifetime)
	case utils.MYSQL:
		d, err = NewMySQLStorage(host, port, name, user, pass, maxConn, maxIdleConn, connMaxLifetime)
	case utils.MONGO:
		d, err = NewMongoStorage(host, port, name, user, pass, utils.StorDB, stringIndexedFields, false)
	case utils.INTERNAL:
		d = NewInternalDB(stringIndexedFields, prefixIndexedFields)
	default:
		err = fmt.Errorf("unknown db '%s' valid options are [%s, %s, %s, %s]",
			db_type, utils.MYSQL, utils.MONGO, utils.POSTGRES, utils.INTERNAL)
	}
	if err != nil {
		return nil, err
	}
	return d, nil
}

func ConfigureStorDB(db_type, host, port, name, user, pass string,
	maxConn, maxIdleConn, connMaxLifetime int,
	stringIndexedFields, prefixIndexedFields []string) (db StorDB, err error) {
	var d StorDB
	switch db_type {
	case utils.POSTGRES:
		d, err = NewPostgresStorage(host, port, name, user, pass, maxConn, maxIdleConn, connMaxLifetime)
	case utils.MYSQL:
		d, err = NewMySQLStorage(host, port, name, user, pass, maxConn, maxIdleConn, connMaxLifetime)
	case utils.MONGO:
		d, err = NewMongoStorage(host, port, name, user, pass, utils.StorDB, stringIndexedFields, false)
	case utils.INTERNAL:
		d = NewInternalDB(stringIndexedFields, prefixIndexedFields)
	default:
		err = fmt.Errorf("unknown db '%s' valid options are [%s, %s, %s, %s]",
			db_type, utils.MYSQL, utils.MONGO, utils.POSTGRES, utils.INTERNAL)
	}
	if err != nil {
		return nil, err
	}
	return d, nil
}

// Stores one Cost coming from SM
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
