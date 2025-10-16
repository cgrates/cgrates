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
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func (m *Migrator) migrateLoadIDs() (err error) {
	var vrs engine.Versions
	if vrs, err = m.getVersions(utils.LoadIDsVrs); err != nil {
		return
	}
	if vrs[utils.LoadIDs] != 1 {
		mOutDB, err := m.GetOUTConn(utils.MetaLoadIDs)
		if err != nil {
			return err
		}
		dataDB, _, err := mOutDB.DataManager().DBConns().GetConn(utils.MetaLoadIDs)
		if err != nil {
			return err
		}
		if err = dataDB.RemoveLoadIDsDrv(); err != nil {
			return err
		}
		if err = m.setVersions(utils.LoadIDsVrs); err != nil {
			return err
		}
	}

	return
}
