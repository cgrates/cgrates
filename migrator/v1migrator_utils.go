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

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

func NewMigratorDataDB(db_type, host, port, name, user, pass, marshaler string,
	cacheCfg config.CacheConfig, loadHistorySize int) (db MigratorDataDB, err error) {
	dm, err := engine.ConfigureDataStorage(db_type,
		host, port, name, user, pass, marshaler,
		cacheCfg, loadHistorySize)
	if err != nil {
		return nil, err
	}
	switch db_type {
	case utils.REDIS:
		d := newRedisMigrator(dm)
		db = d.(MigratorDataDB)
	case utils.MONGO:
		d := newMongoMigrator(dm)
		db = d.(MigratorDataDB)
	default:
		err = errors.New(fmt.Sprintf("Unknown db '%s' valid options are '%s' or '%s'",
			db_type, utils.REDIS, utils.MONGO))
	}
}

/*

func NewMigratorStorDB(db_type, host, port, name, user, pass, marshaler string,
	cacheCfg config.CacheConfig, loadHistorySize int) (db MigratorDataDB, err error) {
	dm, err := engine.ConfigureStorStorage(db_type,
		host, port, name, user, pass, marshaler,
		cacheCfg, loadHistorySize)
	if err != nil {
		return nil, err
	}
	switch db_type {
	case utils.MONGO:
		d := newRedisMigrator(dm)
		db = d.(MigratorDataDB)
	case utils.MYSQL:
		d := newMongoMigrator(dm)
		db = d.(MigratorDataDB)
	default:
		err = errors.New(fmt.Sprintf("Unknown db '%s' valid options are '%s' or '%s'",
			db_type, utils.REDIS, utils.MONGO))
	}
}
*/

/*

func ConfigureV1StorDB(db_type, host, port, name, user, pass string) (db MigratorStorDB, err error) {
	var d MigratorStorDB
	switch db_type {
	case utils.MONGO:
		d, err = newv1MongoStorage(host, port, name, user, pass, utils.StorDB, nil)
		db = d.(MigratorStorDB)
	case utils.MYSQL:
		d, err = newSqlStorage(host, port, name, user, pass)
		db = d.(MigratorStorDB)
	default:
		err = errors.New(fmt.Sprintf("Unknown db '%s' valid options are '%s'",
			db_type, utils.MONGO))
	}
	if err != nil {
		return nil, err
	}
	return d, nil
}
*/
