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

func (m *Migrator) migrateCurrentTPRoutes() (err error) {
	tpids, err := m.storDBIn.StorDB().GetTpIds(utils.TBLTPRoutes)
	if err != nil {
		return err
	}

	for _, tpid := range tpids {
		ids, err := m.storDBIn.StorDB().GetTpTableIds(tpid, utils.TBLTPRoutes,
			[]string{"id"}, map[string]string{}, nil)
		if err != nil {
			return err
		}
		for _, id := range ids {
			routes, err := m.storDBIn.StorDB().GetTPRoutes(tpid, "", id)
			if err != nil {
				return err
			}
			if routes == nil || m.dryRun {
				continue
			}
			if err := m.storDBOut.StorDB().SetTPRoutes(routes); err != nil {
				return err
			}
			for _, route := range routes {
				if err := m.storDBIn.StorDB().RemTpData(utils.TBLTPRoutes, route.TPid,
					map[string]string{"tenant": route.Tenant, "id": route.ID}); err != nil {
					return err
				}
			}

			m.stats[utils.TpRoutes]++
		}
	}
	return
}

func (m *Migrator) migrateTPRoutes() (err error) {
	var vrs engine.Versions
	current := engine.CurrentStorDBVersions()
	if vrs, err = m.getVersions(utils.TpRoutes); err != nil {
		return
	}
	switch vrs[utils.TpRoutes] {
	case current[utils.TpRoutes]:
		if m.sameStorDB {
			break
		}
		if err := m.migrateCurrentTPRoutes(); err != nil {
			return err
		}
	}
	return m.ensureIndexesStorDB(utils.TBLTPRoutes)
}
