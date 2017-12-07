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

func (m *Migrator) migrateCurrentTPaliases() (err error) {
	tpids, err := m.InStorDB().GetTpIds(utils.TBLTPAliases)
	if err != nil {
		return err
	}
	for _, tpid := range tpids {
		dests, err := m.InStorDB().GetTPAliases(&utils.TPAliases{TPid: tpid})
		if err != nil {
			return err
		}
		for _, dest := range dests {
			dest.TPid = tpid
		}
		if dests != nil {
			if m.dryRun != true {
				if err := m.OutStorDB().SetTPAliases(dests); err != nil {
					return err
				}
				m.stats[utils.TpAliases] += 1
			}
		}
	}

	return
}

func (m *Migrator) migrateTPaliases() (err error) {
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
	switch vrs[utils.TpAliases] {
	case current[utils.TpAliases]:
		if m.sameStorDB {
			return
		}
		if err := m.migrateCurrentTPaliases(); err != nil {
			return err
		}
		return
	}
	return
}
