// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package migrator

import (
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func (m *Migrator) migrateCurrentTPfilters() (err error) {
	tpids, err := m.storDBIn.StorDB().GetTpIds(utils.TBLTPFilters)
	if err != nil {
		return err
	}
	for _, tpid := range tpids {
		ids, err := m.storDBIn.StorDB().GetTpTableIds(tpid, utils.TBLTPFilters,
			utils.TPDistinctIds{"id"}, map[string]string{}, nil)
		if err != nil {
			return err
		}
		for _, id := range ids {
			fltrs, err := m.storDBIn.StorDB().GetTPFilters(tpid, "", id)
			if err != nil {
				return err
			}
			if fltrs == nil || m.dryRun {
				continue
			}
			if err := m.storDBOut.StorDB().SetTPFilters(fltrs); err != nil {
				return err
			}
			for _, fltr := range fltrs {
				if err := m.storDBIn.StorDB().RemTpData(utils.TBLTPFilters,
					fltr.TPid, map[string]string{"tenant": fltr.Tenant, "id": fltr.ID}); err != nil {
					return err
				}
			}
			m.stats[utils.TpFilters]++
		}
	}
	return
}

func (m *Migrator) migrateTPfilters() (err error) {
	var vrs engine.Versions
	current := engine.CurrentStorDBVersions()
	if vrs, err = m.getVersions(utils.TpFilters); err != nil {
		return
	}
	switch vrs[utils.TpFilters] {
	case current[utils.TpFilters]:
		if m.sameStorDB {
			break
		}
		if err := m.migrateCurrentTPfilters(); err != nil {
			return err
		}
	}
	return m.ensureIndexesStorDB(utils.TBLTPFilters)
}
