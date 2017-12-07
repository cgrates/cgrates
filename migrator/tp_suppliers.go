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

func (m *Migrator) migrateCurrentTPSuppliers() (err error) {
	tpids, err := m.InStorDB().GetTpIds(utils.TBLTPSuppliers)
	if err != nil {
		return err
	}

	for _, tpid := range tpids {
		ids, err := m.InStorDB().GetTpTableIds(tpid, utils.TBLTPSuppliers, utils.TPDistinctIds{"id"}, map[string]string{}, nil)
		if err != nil {
			return err
		}
		for _, id := range ids {

			dest, err := m.InStorDB().GetTPSuppliers(tpid, id)
			if err != nil {
				return err
			}
			if dest != nil {
				if m.dryRun != true {
					if err := m.OutStorDB().SetTPSuppliers(dest); err != nil {
						return err
					}
					m.stats[utils.TpSuppliers] += 1
				}
			}
		}
	}
	return
}

func (m *Migrator) migrateTPSuppliers() (err error) {
	var vrs engine.Versions
	current := engine.CurrentStorDBVersions()
	vrs, err = m.OutStorDB().GetVersions(utils.TBLTPSuppliers)
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
	switch vrs[utils.TpSuppliers] {
	case current[utils.TpSuppliers]:
		if m.sameStorDB {
			return
		}
		if err := m.migrateCurrentTPSuppliers(); err != nil {
			return err
		}
		return
	}
	return
}
