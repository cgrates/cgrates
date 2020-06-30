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

func (m *Migrator) migrateCurrentTPdestinationrates() (err error) {
	tpids, err := m.storDBIn.StorDB().GetTpIds(utils.TBLTPDestinationRates)
	if err != nil {
		return err
	}
	for _, tpid := range tpids {
		ids, err := m.storDBIn.StorDB().GetTpTableIds(tpid, utils.TBLTPDestinationRates, utils.TPDistinctIds{"tag"}, map[string]string{}, nil)
		if err != nil {
			return err
		}
		for _, id := range ids {
			destRate, err := m.storDBIn.StorDB().GetTPDestinationRates(tpid, id, nil)
			if err != nil {
				return err
			}
			if destRate != nil {
				if m.dryRun != true {
					if err := m.storDBOut.StorDB().SetTPDestinationRates(destRate); err != nil {
						return err
					}
					for _, dest := range destRate {
						if err := m.storDBIn.StorDB().RemTpData(utils.TBLTPDestinationRates, dest.TPid, map[string]string{"tag": dest.ID}); err != nil {
							return err
						}
					}
					m.stats[utils.TpDestinationRates] += 1
				}
			}
		}
	}
	return
}

func (m *Migrator) migrateTPdestinationrates() (err error) {
	var vrs engine.Versions
	current := engine.CurrentStorDBVersions()
	if vrs, err = m.getVersions(utils.TpDestinationRates); err != nil {
		return
	}
	switch vrs[utils.TpDestinationRates] {
	case current[utils.TpDestinationRates]:
		if m.sameStorDB {
			break
		}
		if err := m.migrateCurrentTPdestinationrates(); err != nil {
			return err
		}
	}
	return m.ensureIndexesStorDB(utils.TBLTPDestinationRates)
}
