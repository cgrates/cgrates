// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

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
			utils.TPDistinctIds{"id"}, map[string]string{}, nil)
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
