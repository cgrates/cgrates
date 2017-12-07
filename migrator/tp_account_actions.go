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

func (m *Migrator) migrateCurrentTPaccountAcction() (err error) {
	tpids, err := m.InStorDB().GetTpIds(utils.TBLTPAccountActions)
	if err != nil {
		return err
	}
	for _, tpid := range tpids {
		dest, err := m.InStorDB().GetTPAccountActions(&utils.TPAccountActions{TPid: tpid})
		if err != nil {
			return err
		}
		if dest != nil {
			if m.dryRun != true {
				if err := m.OutStorDB().SetTPAccountActions(dest); err != nil {
					return err
				}
				m.stats[utils.TpAccountActionsV] += 1
			}
		}
	}
	return
}

func (m *Migrator) migrateTPaccountacction() (err error) {
	var vrs engine.Versions
	current := engine.CurrentStorDBVersions()
	vrs, err = m.OutStorDB().GetVersions(utils.TBLVersions)
	if err != nil {
		return utils.NewCGRError(utils.Migrator,
			utils.ServerErrorCaps,
			err.Error(),
			fmt.Sprintf("error: <%s> when querying oldDataDB for versions", err.Error()))
	} else if len(vrs) == 0 {
		return utils.NewCGRError(utils.Migrator,
			utils.MandatoryIEMissingCaps,
			utils.UndefinedVersion,
			"version number is not defined for ActionTriggers model")
	}
	switch vrs[utils.TpAccountActionsV] {
	case current[utils.TpAccountActionsV]:
		if m.sameStorDB {
			return
		}
		if err := m.migrateCurrentTPaccountAcction(); err != nil {
			return err
		}
		return
	}
	return
}
