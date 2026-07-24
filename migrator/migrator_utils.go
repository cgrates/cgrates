// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

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
