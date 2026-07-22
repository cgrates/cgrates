// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package migrator

import (
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func (m *Migrator) migrateCurrentTPactiontriggers() (err error) {
	tpids, err := m.storDBIn.StorDB().GetTpIds(utils.TBLTPActionTriggers)
	if err != nil {
		return err
	}
	for _, tpid := range tpids {
		ids, err := m.storDBIn.StorDB().GetTpTableIds(tpid, utils.TBLTPActionTriggers, utils.TPDistinctIds{"tag"}, map[string]string{}, nil)
		if err != nil {
			return err
		}
		for _, id := range ids {
			actTrg, err := m.storDBIn.StorDB().GetTPActionTriggers(tpid, id)
			if err != nil {
				return err
			}
			if actTrg != nil {
				if !m.dryRun {
					if err := m.storDBOut.StorDB().SetTPActionTriggers(actTrg); err != nil {
						return err
					}
					for _, act := range actTrg {
						if err := m.storDBIn.StorDB().RemTpData(
							utils.TBLTPActionTriggers, act.TPid, map[string]string{"tag": act.ID}); err != nil {
							return err
						}
					}
					m.stats[utils.TpActionTriggers] += 1
				}
			}
		}
	}
	return
}

func (m *Migrator) migrateTPactiontriggers() (err error) {
	var vrs engine.Versions
	current := engine.CurrentStorDBVersions()
	if vrs, err = m.getVersions(utils.TpActionTriggers); err != nil {
		return
	}
	switch vrs[utils.TpActionTriggers] {
	case current[utils.TpActionTriggers]:
		if m.sameStorDB {
			break
		}
		if err := m.migrateCurrentTPactiontriggers(); err != nil {
			return err
		}
	}
	return m.ensureIndexesStorDB(utils.TBLTPActionTriggers)
}
