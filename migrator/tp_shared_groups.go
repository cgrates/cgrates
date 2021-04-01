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
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func (m *Migrator) migrateCurrentTPsharedgroups() (err error) {
	tpids, err := m.storDBIn.StorDB().GetTpIds(utils.TBLTPSharedGroups)
	if err != nil {
		return err
	}

	for _, tpid := range tpids {
		ids, err := m.storDBIn.StorDB().GetTpTableIds(tpid, utils.TBLTPSharedGroups, utils.TPDistinctIds{"tag"}, map[string]string{}, nil)
		if err != nil {
			return err
		}
		for _, id := range ids {

			sharedGroup, err := m.storDBIn.StorDB().GetTPSharedGroups(tpid, id)
			if err != nil {
				return err
			}
			if sharedGroup != nil {
				if !m.dryRun {
					if err := m.storDBOut.StorDB().SetTPSharedGroups(sharedGroup); err != nil {
						return err
					}
					for _, shrGr := range sharedGroup {
						if err := m.storDBIn.StorDB().RemTpData(utils.TBLTPSharedGroups, shrGr.TPid,
							map[string]string{"id": shrGr.ID}); err != nil {
							return err
						}
					}
					m.stats[utils.TpSharedGroups] += 1
				}
			}
		}
	}
	return
}

func (m *Migrator) migrateTPsharedgroups() (err error) {
	var vrs engine.Versions
	current := engine.CurrentStorDBVersions()
	if vrs, err = m.getVersions(utils.TpSharedGroups); err != nil {
		return
	}
	switch vrs[utils.TpSharedGroups] {
	case current[utils.TpSharedGroups]:
		if m.sameStorDB {
			break
		}
		if err := m.migrateCurrentTPsharedgroups(); err != nil {
			return err
		}
	}
	return m.ensureIndexesStorDB(utils.TBLTPResources)
}
