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

func (m *Migrator) migrateCurrentTPActionProfiles() (err error) {
	tpIds, err := m.storDBIn.StorDB().GetTpIds(utils.TBLTPActionProfiles)
	if err != nil {
		return err
	}

	for _, tpid := range tpIds {
		ids, err := m.storDBIn.StorDB().GetTpTableIds(tpid, utils.TBLTPActionProfiles,
			[]string{"id"}, map[string]string{}, nil)
		if err != nil {
			return err
		}

		for _, id := range ids {
			actionProfiles, err := m.storDBIn.StorDB().GetTPActionProfiles(tpid, utils.EmptyString, id)
			if err != nil {
				return err
			}
			if actionProfiles == nil || m.dryRun {
				continue
			}
			if err := m.storDBOut.StorDB().SetTPActionProfiles(actionProfiles); err != nil {
				return err
			}
			for _, actionProfile := range actionProfiles {
				if err := m.storDBIn.StorDB().RemTpData(utils.TBLTPActionProfiles, actionProfile.TPid,
					map[string]string{"id": actionProfile.ID}); err != nil {
					return err
				}
			}
			m.stats[utils.TpActionProfiles]++
		}
	}
	return
}

func (m *Migrator) migrateTPActionProfiles() (err error) {
	var vrs engine.Versions
	current := engine.CurrentStorDBVersions()
	if vrs, err = m.getVersions(utils.TpActionProfiles); err != nil {
		return
	}
	switch vrs[utils.TpActionProfiles] {
	case current[utils.TpActionProfiles]:
		if m.sameStorDB {
			break
		}
		if err := m.migrateCurrentTPActionProfiles(); err != nil {
			return err
		}
	}
	return m.ensureIndexesStorDB(utils.TBLTPActionProfiles)
}
