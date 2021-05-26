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

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func (m *Migrator) migrateCurrentResource() (err error) {
	var ids []string
	ids, err = m.dmIN.DataManager().DataDB().GetKeysForPrefix(context.TODO(), utils.ResourceProfilesPrefix)
	if err != nil {
		return err
	}
	for _, id := range ids {
		tntID := strings.SplitN(strings.TrimPrefix(id, utils.ResourceProfilesPrefix), utils.InInFieldSep, 2)
		if len(tntID) < 2 {
			return fmt.Errorf("Invalid key <%s> when migrating resource profiles", id)
		}
		res, err := m.dmIN.DataManager().GetResourceProfile(context.TODO(), tntID[0], tntID[1], false, false, utils.NonTransactional)
		if err != nil {
			return err
		}
		if res == nil || m.dryRun {
			continue
		}
		if err := m.dmOut.DataManager().SetResourceProfile(context.TODO(), res, true); err != nil {
			return err
		}
		if err := m.dmIN.DataManager().RemoveResourceProfile(context.TODO(), tntID[0], tntID[1], false); err != nil {
			return err
		}
		m.stats[utils.Resource]++
	}
	return
}

func (m *Migrator) migrateResources() (err error) {
	var vrs engine.Versions
	current := engine.CurrentDataDBVersions()
	if vrs, err = m.getVersions(utils.Resource); err != nil {
		return
	}

	migrated := true
	for {
		version := vrs[utils.Resource]
		for {
			switch version {
			default:
				return fmt.Errorf("Unsupported version %v", version)
			case current[utils.Resource]:
				migrated = false
				if m.sameDataDB {
					break
				}
				if err = m.migrateCurrentResource(); err != nil {
					return
				}
			}
			if version == current[utils.Resource] || err == utils.ErrNoMoreData {
				break
			}
		}
		if err == utils.ErrNoMoreData || !migrated {
			break
		}
		// if !m.dryRun {
		// if err = m.dmIN.DataManager().SetResourceProfile(v2, true); err != nil {
		// return
		// }
		// }
		m.stats[utils.Resource]++
	}
	// All done, update version wtih current one
	if err = m.setVersions(utils.Resource); err != nil {
		return
	}
	return m.ensureIndexesDataDB(engine.ColRsP)
}
