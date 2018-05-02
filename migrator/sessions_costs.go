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

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func (m *Migrator) migrateSessionSCosts() (err error) {
	var vrs engine.Versions
	vrs, err = m.storDBOut.GetVersions("")
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
	switch vrs[utils.SessionSCosts] {
	case 0, 1:
		var isPostGres bool
		var storSQL *sql.DB
		switch m.storDBType {
		case utils.MYSQL:
			isPostGres = false
			storSQL = m.storDBOut.(*engine.SQLStorage).Db
		case utils.POSTGRES:
			isPostGres = true
			storSQL = m.storDBOut.(*engine.SQLStorage).Db
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
			return err
		}
		fallthrough // incremental updates
	case 2: // Simply removing them should be enough since if the system is offline they are most probably stale already
		if err := m.storDBOut.RemoveSMCost(nil); err != nil {
			return err
		}

	}
	m.stats[utils.SessionSCosts] = -1
	vrs = engine.Versions{utils.SessionSCosts: engine.CurrentStorDBVersions()[utils.SessionSCosts]}
	if err := m.storDBOut.SetVersions(vrs, false); err != nil {
		return utils.NewCGRError(utils.Migrator,
			utils.ServerErrorCaps,
			err.Error(),
			fmt.Sprintf("error: <%s> when updating SessionSCosts version into StorDB", err.Error()))
	}
	return nil
}
