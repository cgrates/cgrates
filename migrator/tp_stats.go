// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package migrator

import (
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func (m *Migrator) migrateCurrentTPstats() (err error) {
	tpids, err := m.storDBIn.StorDB().GetTpIds(utils.TBLTPStats)
	if err != nil {
		return err
	}

	for _, tpid := range tpids {
		ids, err := m.storDBIn.StorDB().GetTpTableIds(tpid, utils.TBLTPStats,
			utils.TPDistinctIds{"id"}, map[string]string{}, nil)
		if err != nil {
			return err
		}
		for _, id := range ids {
			stats, err := m.storDBIn.StorDB().GetTPStats(tpid, "", id)
			if err != nil {
				return err
			}
			if stats == nil || m.dryRun {
				continue
			}
			if err := m.storDBOut.StorDB().SetTPStats(stats); err != nil {
				return err
			}
			for _, stat := range stats {
				if err := m.storDBIn.StorDB().RemTpData(utils.TBLTPStats, stat.TPid,
					map[string]string{"id": stat.ID}); err != nil {
					return err
				}
			}
			m.stats[utils.TpStats]++
		}
	}
	return
}

func (m *Migrator) migrateTPstats() (err error) {
	var vrs engine.Versions
	current := engine.CurrentStorDBVersions()
	if vrs, err = m.getVersions(utils.TpStats); err != nil {
		return
	}
	switch vrs[utils.TpStats] {
	case current[utils.TpStats]:
		if m.sameStorDB {
			break
		}
		if err := m.migrateCurrentTPstats(); err != nil {
			return err
		}
	}
	return m.ensureIndexesStorDB(utils.TBLTPStats)
}
