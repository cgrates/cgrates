// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package migrator

import (
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func (m *Migrator) migrateCurrentTPthresholds() (err error) {
	tpids, err := m.storDBIn.StorDB().GetTpIds(utils.TBLTPThresholds)
	if err != nil {
		return err
	}

	for _, tpid := range tpids {
		ids, err := m.storDBIn.StorDB().GetTpTableIds(tpid, utils.TBLTPThresholds,
			utils.TPDistinctIds{"id"}, map[string]string{}, nil)
		if err != nil {
			return err
		}
		for _, id := range ids {
			thresholds, err := m.storDBIn.StorDB().GetTPThresholds(tpid, "", id)
			if err != nil {
				return err
			}
			if thresholds == nil || m.dryRun {
				continue
			}
			if err := m.storDBOut.StorDB().SetTPThresholds(thresholds); err != nil {
				return err
			}
			for _, threshold := range thresholds {
				if err := m.storDBIn.StorDB().RemTpData(utils.TBLTPThresholds, threshold.TPid,
					map[string]string{"tenant": threshold.Tenant, "id": threshold.ID}); err != nil {
					return err
				}
			}
			m.stats[utils.TpThresholds]++
		}
	}
	return
}

func (m *Migrator) migrateTPthresholds() (err error) {
	var vrs engine.Versions
	current := engine.CurrentStorDBVersions()
	if vrs, err = m.getVersions(utils.TpThresholds); err != nil {
		return
	}
	switch vrs[utils.TpThresholds] {
	case current[utils.TpThresholds]:
		if m.sameStorDB {
			break
		}
		if err := m.migrateCurrentTPthresholds(); err != nil {
			return err
		}
	}
	return m.ensureIndexesStorDB(utils.TBLTPThresholds)
}
