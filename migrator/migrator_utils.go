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

func NewMigratorDataDBs(dbConnIDList []string, marshaler string,
	cfg *config.CGRConfig) (db map[string]MigratorDataDB, err error) {
	dataDBs := make(map[string]engine.DataDB, len(dbConnIDList))
	for _, dbConnID := range dbConnIDList {
		dbCon, err := engine.NewDataDBConn(cfg.DbCfg().DBConns[dbConnID].Type,
			cfg.DbCfg().DBConns[dbConnID].Host, cfg.DbCfg().DBConns[dbConnID].Port,
			cfg.DbCfg().DBConns[dbConnID].Name, cfg.DbCfg().DBConns[dbConnID].User,
			cfg.DbCfg().DBConns[dbConnID].Password, marshaler,
			cfg.DbCfg().DBConns[dbConnID].StringIndexedFields,
			cfg.DbCfg().DBConns[dbConnID].PrefixIndexedFields,
			cfg.MigratorCgrCfg().OutDBOpts, cfg.DbCfg().Items)
		if err != nil {
			return nil, err
		}
		dataDBs[dbConnID] = dbCon
	}
	dbcManager := engine.NewDBConnManager(dataDBs, cfg.DbCfg())
	dm := engine.NewDataManager(dbcManager, cfg, nil)
	d := make(map[string]MigratorDataDB, len(dbConnIDList))
	for _, dbConnID := range dbConnIDList {
		switch cfg.DbCfg().DBConns[dbConnID].Type {
		case utils.MetaRedis:
			d[dbConnID] = newRedisMigrator(dm)
		case utils.MetaMongo:
			d[dbConnID] = newMongoMigrator(dm)
		case utils.MetaInternal:
			d[dbConnID] = newInternalMigrator(dm)
		default:
			err = fmt.Errorf("unknown db '%s' valid options are '%s' or '%s or '%s'",
				cfg.DbCfg().DBConns[dbConnID].Type, utils.MetaRedis, utils.MetaMongo, utils.MetaInternal)
		}
	}
	return d, nil
}

func (m *Migrator) getVersions(str string) (vrs engine.Versions, err error) {
	mInDB, err := m.GetINConn(utils.CacheVersions)
	if err != nil {
		return nil, err
	}
	dataDB, _, err := mInDB.DataManager().DBConns().GetConn(utils.CacheVersions)
	if err != nil {
		return nil, err
	}
	vrs, err = dataDB.GetVersions(utils.EmptyString)
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
	mOutDB, err := m.GetOUTConn(utils.CacheVersions)
	if err != nil {
		return err
	}
	dataDB, _, err := mOutDB.DataManager().DBConns().GetConn(utils.CacheVersions)
	if err != nil {
		return err
	}
	vrs := engine.Versions{str: engine.CurrentDataDBVersions()[str]}
	err = dataDB.SetVersions(vrs, false)
	if err != nil {
		err = utils.NewCGRError(utils.Migrator,
			utils.ServerErrorCaps,
			err.Error(),
			fmt.Sprintf("error: <%s> when updating %s version into DataDB", err.Error(), str))
	}
	return
}
