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

func (m *Migrator) migrateCurrentTPresources() (err error) {
	tpids, err := m.storDBIn.StorDB().GetTpIds(utils.TBLTPResources)
	if err != nil {
		return err
	}

	for _, tpid := range tpids {
		ids, err := m.storDBIn.StorDB().GetTpTableIds(tpid, utils.TBLTPResources,
			[]string{"id"}, map[string]string{}, nil)
		if err != nil {
			return err
		}
		for _, id := range ids {

			resources, err := m.storDBIn.StorDB().GetTPResources(tpid, "", id)
			if err != nil {
				return err
			}
			if resources == nil || m.dryRun {
				continue
			}
			if err := m.storDBOut.StorDB().SetTPResources(resources); err != nil {
				return err
			}
			for _, resource := range resources {
				if err := m.storDBIn.StorDB().RemTpData(utils.TBLTPResources, resource.TPid,
					map[string]string{"id": resource.ID}); err != nil {
					return err
				}
			}
			m.stats[utils.TpResources]++
		}
	}
	return
}

func (m *Migrator) migrateTPresources() (err error) {
	var vrs engine.Versions
	current := engine.CurrentStorDBVersions()
	if vrs, err = m.getVersions(utils.TpResources); err != nil {
		return
	}
	switch vrs[utils.TpResources] {
	case current[utils.TpResources]:
		if m.sameStorDB {
			break
		}
		if err := m.migrateCurrentTPresources(); err != nil {
			return err
		}
	}
	return m.ensureIndexesStorDB(utils.TBLTPResources)
}
