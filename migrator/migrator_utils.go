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
	"fmt"
	"strings"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	SUPPLIER = "Supplier"
)

func NewMigratorDataDB(db_type, host, port, name, user, pass,
	marshaler string, cacheCfg *config.CacheCfg,
	opts map[string]interface{}) (db MigratorDataDB, err error) {
	dbCon, err := engine.NewDataDBConn(db_type, host,
		port, name, user, pass, marshaler, opts)
	if err != nil {
		return nil, err
	}
	dm := engine.NewDataManager(dbCon, cacheCfg, nil)
	var d MigratorDataDB
	switch db_type {
	case utils.Redis:
		d = newRedisMigrator(dm)
	case utils.Mongo:
		d = newMongoMigrator(dm)
		db = d.(MigratorDataDB)
	case utils.Internal:
		d = newInternalMigrator(dm)
		db = d.(MigratorDataDB)
	default:
		err = fmt.Errorf("unknown db '%s' valid options are '%s' or '%s or '%s'",
			db_type, utils.Redis, utils.Mongo, utils.Internal)
	}
	return d, nil
}

func NewMigratorStorDB(db_type, host, port, name, user, pass, marshaler string,
	stringIndexedFields, prefixIndexedFields []string,
	opts map[string]interface{}) (db MigratorStorDB, err error) {
	var d MigratorStorDB
	storDb, err := engine.NewStorDBConn(db_type, host, port, name, user,
		pass, marshaler, stringIndexedFields, prefixIndexedFields, opts)
	if err != nil {
		return nil, err
	}
	switch db_type {
	case utils.Mongo:
		d = newMongoStorDBMigrator(storDb)
		db = d.(MigratorStorDB)
	case utils.MySQL:
		d = newMigratorSQL(storDb)
		db = d.(MigratorStorDB)
	case utils.Postgres:
		d = newMigratorSQL(storDb)
		db = d.(MigratorStorDB)
	case utils.Internal:
		d = newInternalStorDBMigrator(storDb)
		db = d.(MigratorStorDB)
	default:
		err = fmt.Errorf("Unknown db '%s' valid options are [%s, %s, %s, %s]",
			db_type, utils.MySQL, utils.Mongo, utils.Postgres, utils.Internal)
	}
	return d, nil
}
func (m *Migrator) getVersions(str string) (vrs engine.Versions, err error) {
	if str == utils.CDRs || str == utils.SessionSCosts || strings.HasPrefix(str, "Tp") {
		vrs, err = m.storDBIn.StorDB().GetVersions(utils.EmptyString)
	} else {
		vrs, err = m.dmIN.DataManager().DataDB().GetVersions(utils.EmptyString)
	}
	if err != nil {
		return nil, utils.NewCGRError(utils.Migrator,
			utils.ServerErrorCaps,
			err.Error(),
			fmt.Sprintf("error: <%s> when querying oldDataDB for versions", err.Error()))
	} else if len(vrs) == 0 {
		return nil, utils.NewCGRError(utils.Migrator,
			utils.MandatoryIEMissingCaps,
			utils.UndefinedVersion,
			"version number is not defined for "+str)
	}
	return
}

func (m *Migrator) setVersions(str string) (err error) {
	if str == utils.CDRs || str == utils.SessionSCosts || strings.HasPrefix(str, "Tp") {
		vrs := engine.Versions{str: engine.CurrentStorDBVersions()[str]}
		err = m.storDBOut.StorDB().SetVersions(vrs, false)
	} else {
		vrs := engine.Versions{str: engine.CurrentDataDBVersions()[str]}
		err = m.dmOut.DataManager().DataDB().SetVersions(vrs, false)
	}
	if err != nil {
		err = utils.NewCGRError(utils.Migrator,
			utils.ServerErrorCaps,
			err.Error(),
			fmt.Sprintf("error: <%s> when updating %s version into StorDB", err.Error(), str))
	}
	return
}
