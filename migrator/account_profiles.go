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

func (m *Migrator) migrateCurrentAccountProfiles() (err error) {
	var ids []string
	ids, err = m.dmIN.DataManager().DataDB().GetKeysForPrefix(utils.AccountProfilePrefix)
	if err != nil {
		return err
	}
	for _, id := range ids {
		tntID := strings.SplitN(strings.TrimPrefix(id, utils.AccountProfilePrefix), utils.InInFieldSep, 2)
		if len(tntID) < 2 {
			return fmt.Errorf("Invalid key <%s> when migrating from account profiles", id)
		}
		ap, err := m.dmIN.DataManager().GetAccountProfile(tntID[0], tntID[1])
		if err != nil {
			return err
		}
		if ap == nil || m.dryRun {
			continue
		}
		if err := m.dmOut.DataManager().SetAccount(ap, true); err != nil {
			return err
		}
		if err := m.dmIN.DataManager().RemoveAccountProfile(tntID[0], tntID[1], utils.NonTransactional, false); err != nil {
			return err
		}
		m.stats[utils.AccountProfilesString]++
	}
	return
}

func (m *Migrator) migrateAccountProfiles() (err error) {
	var vrs engine.Versions
	current := engine.CurrentDataDBVersions()
	if vrs, err = m.getVersions(utils.AccountProfilesString); err != nil {
		return
	}
	migrated := true
	for {
		version := vrs[utils.AccountProfilesString]
		for {
			switch version {
			default:
				return fmt.Errorf("Unsupported version %v", version)
			case current[utils.AccountProfilesString]:
				migrated = false
				if m.sameDataDB {
					break
				}
				if err = m.migrateCurrentAccountProfiles(); err != nil {
					return
				}
			}
			if version == current[utils.AccountProfilesString] || err == utils.ErrNoMoreData {
				break
			}
		}
		if err == utils.ErrNoMoreData || !migrated {
			break
		}
		m.stats[utils.AccountProfilesString]++
	}
	//All done, update version with current one
	if err = m.setVersions(utils.AccountProfilesString); err != nil {
		return
	}
	return m.ensureIndexesDataDB(engine.ColApp)
}
