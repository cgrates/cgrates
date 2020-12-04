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
	"fmt"
	"strings"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func (m *Migrator) migrateCurrentDestinations() (err error) {
	var ids []string
	ids, err = m.dmIN.DataManager().DataDB().GetKeysForPrefix(utils.DESTINATION_PREFIX)
	if err != nil {
		return err
	}
	for _, id := range ids {
		idg := strings.TrimPrefix(id, utils.DESTINATION_PREFIX)
		dst, err := m.dmIN.DataManager().GetDestination(idg, false, true, utils.NonTransactional)
		if err != nil {
			return err
		}
		if dst == nil || m.dryRun {
			continue
		}
		if err := m.dmOut.DataManager().SetDestination(dst, utils.NonTransactional); err != nil {
			return err
		}
		m.stats[utils.Destinations]++
	}
	return
}

func (m *Migrator) migrateDestinations() (err error) {
	var vrs engine.Versions
	current := engine.CurrentDataDBVersions()
	if vrs, err = m.getVersions(utils.Destinations); err != nil {
		return
	}
	migrated := true
	for {
		version := vrs[utils.Destinations]
		for {
			switch version {
			default:
				return fmt.Errorf("Unsupported version %v", version)
			case current[utils.Destinations]:
				migrated = false
				if m.sameDataDB {
					break
				}
				if err = m.migrateCurrentDestinations(); err != nil {
					return
				}
			}
			if version == current[utils.Destinations] || err == utils.ErrNoMoreData {
				break
			}
		}
		if err == utils.ErrNoMoreData || !migrated {
			break
		}

		// if !m.dryRun  {
		// 		if err = m.dmIN.DataManager().SetDestination(v2, true); err != nil {
		// 	return
		// }
		// }
		m.stats[utils.Destinations]++
	}
	// All done, update version wtih current one
	if err = m.setVersions(utils.Destinations); err != nil {
		return
	}
	return
}

func (m *Migrator) migrateCurrentReverseDestinations() (err error) {
	var ids []string
	ids, err = m.dmIN.DataManager().DataDB().GetKeysForPrefix(utils.REVERSE_DESTINATION_PREFIX)
	if err != nil {
		return err
	}
	for _, id := range ids {
		id := strings.TrimPrefix(id, utils.REVERSE_DESTINATION_PREFIX)
		rdst, err := m.dmIN.DataManager().GetReverseDestination(id, false, true, utils.NonTransactional)
		if err != nil {
			return err
		}
		if rdst == nil {
			continue
		}
		for _, rdid := range rdst {
			rdstn, err := m.dmIN.DataManager().GetDestination(rdid, false, true, utils.NonTransactional)
			if err != nil {
				return err
			}
			if rdstn == nil || m.dryRun {
				continue
			}
			if err := m.dmOut.DataManager().SetDestination(rdstn, utils.NonTransactional); err != nil {
				return err
			}
			if err := m.dmOut.DataManager().SetReverseDestination(rdstn, utils.NonTransactional); err != nil {
				return err
			}
			m.stats[utils.ReverseDestinations]++
		}
	}
	return
}

func (m *Migrator) migrateReverseDestinations() (err error) {
	var vrs engine.Versions
	current := engine.CurrentDataDBVersions()
	vrs, err = m.dmIN.DataManager().DataDB().GetVersions("")
	if err != nil {
		return utils.NewCGRError(utils.Migrator,
			utils.ServerErrorCaps,
			err.Error(),
			fmt.Sprintf("error: <%s> when querying oldDataDB for versions", err.Error()))
	} else if len(vrs) == 0 {
		return utils.NewCGRError(utils.Migrator,
			utils.MandatoryIEMissingCaps,
			utils.UndefinedVersion,
			"version number is not defined for ActionTriggers model")
	}
	switch vrs[utils.ReverseDestinations] {
	case current[utils.ReverseDestinations]:
		if m.sameDataDB {
			break
		}
		if err = m.migrateCurrentReverseDestinations(); err != nil {
			return err
		}
	}
	return m.ensureIndexesDataDB(engine.ColDst, engine.ColRds)
}
