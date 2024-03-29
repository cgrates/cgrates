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

func (m *Migrator) migrateCurrentTPactions() (err error) {
	tpids, err := m.storDBIn.StorDB().GetTpIds(utils.TBLTPActions)
	if err != nil {
		return err
	}
	for _, tpid := range tpids {
		ids, err := m.storDBIn.StorDB().GetTpTableIds(tpid, utils.TBLTPActions, utils.TPDistinctIds{"tag"}, map[string]string{}, nil)
		if err != nil {
			return err
		}
		for _, id := range ids {
			action, err := m.storDBIn.StorDB().GetTPActions(tpid, id)
			if err != nil {
				return err
			}
			if action != nil {
				if !m.dryRun {
					if err := m.storDBOut.StorDB().SetTPActions(action); err != nil {
						return err
					}
					for _, act := range action {
						if err := m.storDBIn.StorDB().RemTpData(utils.TBLTPActions, act.TPid, map[string]string{"tag": act.ID}); err != nil {
							return err
						}
					}
					m.stats[utils.TpActions] += 1
				}
			}
		}
	}
	return
}

func (m *Migrator) migrateTPactions() (err error) {
	var vrs engine.Versions
	current := engine.CurrentStorDBVersions()
	if vrs, err = m.getVersions(utils.TpActions); err != nil {
		return
	}
	switch vrs[utils.TpActions] {
	case current[utils.TpActions]:
		if m.sameStorDB {
			break
		}
		if err := m.migrateCurrentTPactions(); err != nil {
			return err
		}
	}
	return m.ensureIndexesStorDB(utils.TBLTPActions)
}
