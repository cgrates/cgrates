// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package migrator

import (
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func (m *Migrator) migrateCurrentTPrates() (err error) {
	tpids, err := m.storDBIn.StorDB().GetTpIds(utils.TBLTPRates)
	if err != nil {
		return err
	}
	for _, tpid := range tpids {
		ids, err := m.storDBIn.StorDB().GetTpTableIds(tpid, utils.TBLTPRates, utils.TPDistinctIds{"tag"}, map[string]string{}, nil)
		if err != nil {
			return err
		}
		for _, id := range ids {

			rates, err := m.storDBIn.StorDB().GetTPRates(tpid, id)
			if err != nil {
				return err
			}
			if rates != nil {
				if !m.dryRun {
					if err := m.storDBOut.StorDB().SetTPRates(rates); err != nil {
						return err
					}
					for _, rate := range rates {
						if err := m.storDBIn.StorDB().RemTpData(utils.TBLTPRates, rate.TPid, map[string]string{"tag": rate.ID}); err != nil {
							return err
						}
					}
					m.stats[utils.TpRates] += 1
				}
			}
		}
	}
	return
}

func (m *Migrator) migrateTPrates() (err error) {
	var vrs engine.Versions
	current := engine.CurrentStorDBVersions()
	if vrs, err = m.getVersions(utils.TpRates); err != nil {
		return
	}
	switch vrs[utils.TpRates] {
	case current[utils.TpRates]:
		if m.sameStorDB {
			break
		}
		if err := m.migrateCurrentTPrates(); err != nil {
			return err
		}
	}
	return m.ensureIndexesStorDB(utils.TBLTPRates)
}
