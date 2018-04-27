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

func (m *Migrator) migrateCurrentCDRs() (err error) {
	if m.sameStorDB { // no move
		return
	}
	cdrs, _, err := m.storDB.GetCDRs(new(utils.CDRsFilter), false)
	if err != nil {
		return err
	}
	for _, cdr := range cdrs {
		if err := m.oldStorDB.SetCDR(cdr, true); err != nil {
			return err
		}
	}
	return
}

func (m *Migrator) migrateCDRs() (err error) {
	var vrs engine.Versions
	vrs, err = m.dmIN.DataDB().GetVersions("")
	if err != nil {
		return utils.NewCGRError(utils.Migrator,
			utils.ServerErrorCaps,
			err.Error(),
			fmt.Sprintf("error: <%s> when querying oldDataDB for versions", err.Error()))
	} else if len(vrs) == 0 {
		return utils.NewCGRError(utils.Migrator,
			utils.MandatoryIEMissingCaps,
			utils.UndefinedVersion,
			"version number is not defined for Actions")
	}
	current := engine.CurrentDataDBVersions()
	switch vrs[utils.CDRs] {
	case current[utils.CDRs]:
		return m.migrateCurrentCDRs()
	}
	return
}
