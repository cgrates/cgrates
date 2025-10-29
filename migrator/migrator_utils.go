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

package migrator

import (
	"errors"
	"fmt"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func NewMigratorDataDB(db_type, host, port, name, user, pass,
	marshaler string, cacheCfg config.CacheCfg, sentinelName string,
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
	case utils.MetaRedis:
		d = newRedisMigrator(dm)
	case utils.MetaMongo:
		d = newMongoMigrator(dm)
		db = d.(MigratorDataDB)
	case utils.MetaInternal:
		d = newInternalMigrator(dm)
		db = d.(MigratorDataDB)
	default:
		err = errors.New(fmt.Sprintf("Unknown db '%s' valid options are '%s' or '%s or '%s'",
			db_type, utils.MetaRedis, utils.MetaMongo, utils.MetaInternal))
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
	case utils.MetaMongo:
		d = newMongoStorDBMigrator(storDb)
		db = d.(MigratorStorDB)
	case utils.MetaMySQL:
		d = newMigratorSQL(storDb)
		db = d.(MigratorStorDB)
	case utils.MetaPostgres:
		d = newMigratorSQL(storDb)
		db = d.(MigratorStorDB)
	case utils.MetaInternal:
		d = newInternalStorDBMigrator(storDb)
		db = d.(MigratorStorDB)
	default:
		err = errors.New(fmt.Sprintf("Unknown db '%s' valid options are [%s, %s, %s, %s]",
			db_type, utils.MetaMySQL, utils.MetaMongo, utils.MetaPostgres, utils.MetaInternal))
	}
	return d, nil
}
