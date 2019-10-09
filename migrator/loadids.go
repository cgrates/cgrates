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

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

//
func (m *Migrator) migrateLoadIDs() (err error) {
	var vrs engine.Versions
	if vrs, err = m.dmIN.DataManager().DataDB().GetVersions(""); err != nil {
		return utils.NewCGRError(utils.Migrator,
			utils.ServerErrorCaps,
			err.Error(),
			fmt.Sprintf("error: <%s> when querying oldDataDB for versions", err.Error()))
	}
	if vrs[utils.LoadIDs] != 1 {
		if err = m.dmOut.DataManager().DataDB().RemoveLoadIDsDrv(); err != nil {
			return
		}
		// All done, update version wtih current one
		vrs := engine.Versions{utils.LoadIDsVrs: engine.CurrentDataDBVersions()[utils.LoadIDsVrs]}
		if err = m.dmOut.DataManager().DataDB().SetVersions(vrs, false); err != nil {
			return utils.NewCGRError(utils.Migrator,
				utils.ServerErrorCaps,
				err.Error(),
				fmt.Sprintf("error: <%s> when updating LoadIDs version into dataDB", err))
		}
	}

	return
}
