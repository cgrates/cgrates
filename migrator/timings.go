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
	"strings"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func (m *Migrator) migrateCurrentTiming() (err error) {
	var ids []string
	ids, err = m.dmIN.DataManager().DataDB().GetKeysForPrefix(utils.TimingsPrefix)
	if err != nil {
		return err
	}
	for _, id := range ids {
		idg := strings.TrimPrefix(id, utils.TimingsPrefix)
		tm, err := m.dmIN.DataManager().GetTiming(idg, true, utils.NonTransactional)
		if err != nil {
			return err
		}
		if tm == nil || m.dryRun {
			continue
		}
		if err := m.dmOut.DataManager().SetTiming(tm); err != nil {
			return err
		}
		m.stats[utils.Timing] += 1
	}
	return
}

func (m *Migrator) migrateTimings() (err error) {
	var vrs engine.Versions
	current := engine.CurrentDataDBVersions()
	if vrs, err = m.getVersions(utils.Timing); err != nil {
		return
	}
	switch vrs[utils.Timing] {
	case current[utils.Timing]:
		if m.sameDataDB {
			break
		}
		if err = m.migrateCurrentTiming(); err != nil {
			return err
		}
	}
	return m.ensureIndexesDataDB(engine.ColTmg)
}
