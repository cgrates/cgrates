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
	"fmt"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	SUPPLIER = "Supplier"
)

func NewMigratorDataDB(db_type, host, port, name, user, pass,
	marshaler string, cfg *config.CGRConfig,
	opts *config.DataDBOpts, itmsCfg map[string]*config.ItemOpts) (db MigratorDataDB, err error) {
	dbCon, err := engine.NewDataDBConn(db_type, host,
		port, name, user, pass, marshaler, opts, itmsCfg)
	if err != nil {
		return nil, err
	}
	dm := engine.NewDataManager(dbCon, cfg, nil)
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
		err = fmt.Errorf("unknown db '%s' valid options are '%s' or '%s or '%s'",
			db_type, utils.MetaRedis, utils.MetaMongo, utils.MetaInternal)
	}
	return d, nil
}

func (m *Migrator) getVersions(str string) (vrs engine.Versions, err error) {
	vrs, err = m.dmIN.DataManager().DataDB().GetVersions(utils.EmptyString)
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
	vrs := engine.Versions{str: engine.CurrentDataDBVersions()[str]}
	err = m.dmOut.DataManager().DataDB().SetVersions(vrs, false)
	if err != nil {
		err = utils.NewCGRError(utils.Migrator,
			utils.ServerErrorCaps,
			err.Error(),
			fmt.Sprintf("error: <%s> when updating %s version into DataDB", err.Error(), str))
	}
	return
}
