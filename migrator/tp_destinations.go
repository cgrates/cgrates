// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package migrator

import (
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func (m *Migrator) migrateCurrentTPDestinations() (err error) {
	tpids, err := m.storDBIn.StorDB().GetTpIds(utils.TBLTPDestinations)
	if err != nil {
		return err
	}
	for _, tpid := range tpids {
		ids, err := m.storDBIn.StorDB().GetTpTableIds(tpid, utils.TBLTPDestinations, utils.TPDistinctIds{"tag"}, map[string]string{}, nil)
		if err != nil {
			return err
		}
		for _, id := range ids {
			destinations, err := m.storDBIn.StorDB().GetTPDestinations(tpid, id)
			if err != nil {
				return err
			}
			if destinations != nil {
				if !m.dryRun {
					if err := m.storDBOut.StorDB().SetTPDestinations(destinations); err != nil {
						return err
					}
					for _, dest := range destinations {
						if err := m.storDBIn.StorDB().RemTpData(utils.TBLTPDestinations, dest.TPid, map[string]string{"tag": dest.ID}); err != nil {
							return err
						}
					}
					m.stats[utils.TpDestinations] += 1
				}
			}
		}
	}
	return
}

func (m *Migrator) migrateTPDestinations() (err error) {
	var vrs engine.Versions
	current := engine.CurrentStorDBVersions()
	if vrs, err = m.getVersions(utils.TpDestinations); err != nil {
		return
	}
	switch vrs[utils.TpDestinations] {
	case current[utils.TpDestinations]:
		if m.sameStorDB {
			break
		}
		if err := m.migrateCurrentTPDestinations(); err != nil {
			return err
		}
	}
	return m.ensureIndexesStorDB(utils.TBLTPDestinations)
}
