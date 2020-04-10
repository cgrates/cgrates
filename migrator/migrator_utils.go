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

package migrator

import (
	"errors"
	"fmt"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func NewMigratorDataDB(db_type, host, port, name, user, pass,
	marshaler string, cacheCfg *config.CacheCfg, sentinelName string,
	itemsCacheCfg map[string]*config.ItemOpt) (db MigratorDataDB, err error) {
	dbCon, err := engine.NewDataDBConn(db_type,
		host, port, name, user, pass, marshaler,
		sentinelName, itemsCacheCfg)
	if err != nil {
		return nil, err
	}
	dm := engine.NewDataManager(dbCon, cacheCfg, nil)
	var d MigratorDataDB
	switch db_type {
	case utils.REDIS:
		d = newRedisMigrator(dm)
	case utils.MONGO:
		d = newMongoMigrator(dm)
		db = d.(MigratorDataDB)
	case utils.INTERNAL:
		d = newInternalMigrator(dm)
		db = d.(MigratorDataDB)
	default:
		err = errors.New(fmt.Sprintf("Unknown db '%s' valid options are '%s' or '%s or '%s'",
			db_type, utils.REDIS, utils.MONGO, utils.INTERNAL))
	}
	return d, nil
}

func NewMigratorStorDB(db_type, host, port, name, user, pass, marshaler, sslmode string,
	maxConn, maxIdleConn, connMaxLifetime int, stringIndexedFields, prefixIndexedFields []string,
	itemsCacheCfg map[string]*config.ItemOpt) (db MigratorStorDB, err error) {
	var d MigratorStorDB
	storDb, err := engine.NewStorDBConn(db_type, host, port, name, user,
		pass, marshaler, sslmode, maxConn, maxIdleConn, connMaxLifetime,
		stringIndexedFields, prefixIndexedFields, itemsCacheCfg)
	if err != nil {
		return nil, err
	}
	switch db_type {
	case utils.MONGO:
		d = newMongoStorDBMigrator(storDb)
		db = d.(MigratorStorDB)
	case utils.MYSQL:
		d = newMigratorSQL(storDb)
		db = d.(MigratorStorDB)
	case utils.POSTGRES:
		d = newMigratorSQL(storDb)
		db = d.(MigratorStorDB)
	case utils.INTERNAL:
		d = newInternalStorDBMigrator(storDb)
		db = d.(MigratorStorDB)
	default:
		err = errors.New(fmt.Sprintf("Unknown db '%s' valid options are [%s, %s, %s, %s]",
			db_type, utils.MYSQL, utils.MONGO, utils.POSTGRES, utils.INTERNAL))
	}
	return d, nil
}
