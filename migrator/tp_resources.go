// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

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
			utils.TPDistinctIds{"id"}, map[string]string{}, nil)
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

func (m *Migrator) migrateTPips() (err error) {
	var vrs engine.Versions
	current := engine.CurrentStorDBVersions()
	if vrs, err = m.getVersions(utils.TpIPs); err != nil {
		return
	}
	switch vrs[utils.TpIPs] {
	case current[utils.TpIPs]:
		if m.sameStorDB {
			break
		}
		if err := m.migrateCurrentTPresources(); err != nil {
			return err
		}
	}
	return m.ensureIndexesStorDB(utils.TBLTPIPs)
}
