// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

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
	opts *config.DataDBOpts, itmsCfg map[string]*config.ItemOpt) (db MigratorDataDB, err error) {
	var dbCon engine.DataDB
	if dbCon, err = engine.NewDataDBConn(db_type, host,
		port, name, user, pass, marshaler, opts, nil); err != nil {
		return
	}
	dm := engine.NewDataManager(dbCon, cacheCfg, nil)
	switch db_type {
	case utils.MetaRedis:
		db = newRedisMigrator(dm)
	case utils.MetaMongo:
		db = newMongoMigrator(dm)
	case utils.MetaInternal:
		db = newInternalMigrator(dm)
	default:
		err = fmt.Errorf("unknown db '%s' valid options are '%s' or '%s or '%s'",
			db_type, utils.MetaRedis, utils.MetaMongo, utils.MetaInternal)
	}
	return
}

func NewMigratorStorDB(db_type, host, port, name, user, pass, marshaler string,
	stringIndexedFields, prefixIndexedFields []string,
	opts *config.StorDBOpts, itmsCfg map[string]*config.ItemOpt) (db MigratorStorDB, err error) {
	var storDb engine.StorDB
	if storDb, err = engine.NewStorDBConn(db_type, host, port, name, user,
		pass, marshaler, stringIndexedFields, prefixIndexedFields, opts, itmsCfg); err != nil {
		return
	}
	switch db_type {
	case utils.MetaMongo:
		db = newMongoStorDBMigrator(storDb)
	case utils.MetaMySQL:
		db = newMigratorSQL(storDb)
	case utils.MetaPostgres:
		db = newMigratorSQL(storDb)
	case utils.MetaInternal:
		db = newInternalStorDBMigrator(storDb)
	default:
		err = fmt.Errorf("Unknown db '%s' valid options are [%s, %s, %s, %s]",
			db_type, utils.MetaMySQL, utils.MetaMongo, utils.MetaPostgres, utils.MetaInternal)
	}
	return
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
