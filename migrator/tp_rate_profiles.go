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

func (m *Migrator) migrateCurrentTPRateProfiles() (err error) {
	tpids, err := m.storDBIn.StorDB().GetTpIds(utils.TBLTPRateProfiles)
	if err != nil {
		return err
	}

	for _, tpid := range tpids {
		ids, err := m.storDBIn.StorDB().GetTpTableIds(tpid, utils.TBLTPRateProfiles,
			[]string{"id"}, map[string]string{}, nil)
		if err != nil {
			return err
		}
		for _, id := range ids {
			rateProfiles, err := m.storDBIn.StorDB().GetTPRateProfiles(tpid, utils.EmptyString, id)
			if err != nil {
				return err
			}
			if rateProfiles == nil || m.dryRun {
				continue
			}
			if err := m.storDBOut.StorDB().SetTPRateProfiles(rateProfiles); err != nil {
				return err
			}
			for _, rateProfile := range rateProfiles {
				if err := m.storDBIn.StorDB().RemTpData(utils.TBLTPRateProfiles, rateProfile.TPid,
					map[string]string{"tenant": rateProfile.Tenant, "id": rateProfile.ID}); err != nil {
					return err
				}
			}
			m.stats[utils.TpRateProfiles]++
		}
	}
	return
}

func (m *Migrator) migrateTPRateProfiles() (err error) {
	var vrs engine.Versions
	current := engine.CurrentStorDBVersions()
	if vrs, err = m.getVersions(utils.TpRateProfiles); err != nil {
		return err
	}
	switch vrs[utils.TpRateProfiles] {
	case current[utils.TpRateProfiles]:
		if m.sameStorDB {
			break
		}
		if err := m.migrateCurrentTPRateProfiles(); err != nil {
			return err
		}
	}
	return m.ensureIndexesStorDB(utils.TpRateProfiles)
}
