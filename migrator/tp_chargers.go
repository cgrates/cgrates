// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package migrator

import (
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func (m *Migrator) migrateCurrentTPChargers() (err error) {
	tpids, err := m.storDBIn.StorDB().GetTpIds(utils.TBLTPChargers)
	if err != nil {
		return err
	}

	for _, tpid := range tpids {
		ids, err := m.storDBIn.StorDB().GetTpTableIds(tpid, utils.TBLTPChargers,
			utils.TPDistinctIds{"id"}, map[string]string{}, nil)
		if err != nil {
			return err
		}
		for _, id := range ids {
			chargers, err := m.storDBIn.StorDB().GetTPChargers(tpid, "", id)
			if err != nil {
				return err
			}
			if chargers == nil || m.dryRun {
				continue
			}
			if err := m.storDBOut.StorDB().SetTPChargers(chargers); err != nil {
				return err
			}
			for _, charger := range chargers {
				if err := m.storDBIn.StorDB().RemTpData(utils.TBLTPChargers, charger.TPid,
					map[string]string{"id": charger.ID}); err != nil {
					return err
				}
			}
			m.stats[utils.TpChargers]++
		}
	}
	return
}

func (m *Migrator) migrateTPChargers() (err error) {
	var vrs engine.Versions
	current := engine.CurrentStorDBVersions()
	if vrs, err = m.getVersions(utils.TpChargers); err != nil {
		return
	}
	switch vrs[utils.TpChargers] {
	case current[utils.TpChargers]:
		if m.sameStorDB {
			break
		}
		if err := m.migrateCurrentTPChargers(); err != nil {
			return err
		}
	}
	return m.ensureIndexesStorDB(utils.TBLTPChargers)
}
