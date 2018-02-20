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
	"database/sql"
	"fmt"
	"log"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func (m *Migrator) migrateSessionsCosts() (err error) {
	var vrs engine.Versions
	vrs, err = m.OutStorDB().GetVersions(utils.TBLVersions)
	if err != nil {
		return utils.NewCGRError(utils.Migrator,
			utils.ServerErrorCaps,
			err.Error(),
			fmt.Sprintf("error: <%s> when querying OutStorDB for versions", err.Error()))
	} else if len(vrs) == 0 {
		return utils.NewCGRError(utils.Migrator,
			utils.MandatoryIEMissingCaps,
			utils.UndefinedVersion,
			"version number is not defined for SessionsCosts model")
	}
	if vrs[utils.SessionsCosts] < 2 {
		var isPostGres bool
		var storSQL *sql.DB
		switch m.storDBType {
		case utils.MYSQL:
			isPostGres = false
			storSQL = m.storDB.(*engine.SQLStorage).Db
		case utils.POSTGRES:
			isPostGres = true
			storSQL = m.storDB.(*engine.SQLStorage).Db
		default:
			return utils.NewCGRError(utils.Migrator,
				utils.MandatoryIEMissingCaps,
				utils.UnsupportedDB,
				fmt.Sprintf("unsupported database type: <%s>", m.storDBType))
		}
		qry := "RENAME TABLE sm_costs TO sessions_costs;"
		if isPostGres {
			qry = "ALTER TABLE sm_costs RENAME TO sessions_costs"
		}
		if _, err := storSQL.Exec(qry); err != nil {
			log.Print(err)
			return err
		}
		m.stats[utils.SessionsCosts] = 2
		vrs = engine.Versions{utils.SessionsCosts: engine.CurrentStorDBVersions()[utils.SessionsCosts]}
		if err := m.storDB.SetVersions(vrs, false); err != nil {
			return utils.NewCGRError(utils.Migrator,
				utils.ServerErrorCaps,
				err.Error(),
				fmt.Sprintf("error: <%s> when updating SessionsCosts version into StorDB", err.Error()))
		}
	}
	return nil
}
