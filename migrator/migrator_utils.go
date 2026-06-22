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

func NewMigratorDataDBs(dbConnIDList []string, marshaler string,
	cfg *config.CGRConfig, cache *engine.CacheS) (db *engine.DataManager, err error) {
	dataDBs := make(map[string]engine.DataDB, len(dbConnIDList))
	for _, dbConnID := range dbConnIDList {
		dbCon, err := engine.NewDBConn(cfg.DbCfg().DBConns[dbConnID].Type,
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
	dm.SetCache(cache)
	return dm, nil
}

func (m *Migrator) getVersions(str string) (engine.Versions, error) {
	dataDB, _, err := m.dmFrom.DBConns().GetConn(utils.CacheVersions)
	if err != nil {
		return nil, err
	}
	vrs, err := dataDB.GetVersions("")
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
	return vrs, nil
}

func (m *Migrator) setVersions(str string) error {
	dataDB, _, err := m.dmTo.DBConns().GetConn(utils.CacheVersions)
	if err != nil {
		return err
	}
	vrs := engine.Versions{str: engine.CurrentDataDBVersions()[str]}
	if err = dataDB.SetVersions(vrs, false); err != nil {
		return utils.NewCGRError(utils.Migrator,
			utils.ServerErrorCaps,
			err.Error(),
			fmt.Sprintf("error: <%s> when updating %s version into DataDB", err.Error(), str))
	}
	return nil
}
