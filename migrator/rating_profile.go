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

func (m *Migrator) migrateCurrentRatingProfiles() (err error) {
	var ids []string
	ids, err = m.dmIN.DataManager().DataDB().GetKeysForPrefix(utils.RatingProfilePrefix)
	if err != nil {
		return err
	}
	for _, id := range ids {
		idg := strings.TrimPrefix(id, utils.RatingProfilePrefix)
		rp, err := m.dmIN.DataManager().GetRatingProfile(idg, true, utils.NonTransactional)
		if err != nil {
			return err
		}
		if rp == nil || m.dryRun {
			continue
		}
		if err := m.dmOut.DataManager().SetRatingProfile(rp, utils.NonTransactional); err != nil {
			return err
		}
		if err := m.dmIN.DataManager().RemoveRatingProfile(idg, utils.NonTransactional); err != nil {
			return err
		}
		m.stats[utils.RatingProfile]++
	}
	return
}

func (m *Migrator) migrateRatingProfiles() (err error) {
	var vrs engine.Versions
	current := engine.CurrentDataDBVersions()
	if vrs, err = m.getVersions(utils.RatingProfile); err != nil {
		return
	}

	migrated := true
	for {
		version := vrs[utils.RatingProfile]
		for {
			switch version {
			default:
				return fmt.Errorf("Unsupported version %v", version)
			case current[utils.RatingProfile]:
				migrated = false
				if m.sameDataDB {
					break
				}
				if err = m.migrateCurrentRatingProfiles(); err != nil {
					return err
				}
			}
			if version == current[utils.RatingProfile] || err == utils.ErrNoMoreData {
				break
			}
		}
		if err == utils.ErrNoMoreData || !migrated {
			break
		}
		// if !m.dryRun {
		// if err = m.dmIN.DataManager().SetRatingProfile(v2, true); err != nil {
		// return
		// }
		// }
		m.stats[utils.RatingProfile]++
	}
	// All done, update version wtih current one
	if err = m.setVersions(utils.RatingProfile); err != nil {
		return
	}
	return m.ensureIndexesDataDB(engine.ColRpf)
}
