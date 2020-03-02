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

func (m *Migrator) migrateCurrentDispatcher() (err error) {
	var ids []string
	ids, err = m.dmIN.DataManager().DataDB().GetKeysForPrefix(utils.DispatcherProfilePrefix)
	if err != nil {
		return err
	}
	for _, id := range ids {
		tntID := strings.SplitN(strings.TrimPrefix(id, utils.DispatcherProfilePrefix), utils.InInFieldSep, 2)
		if len(tntID) < 2 {
			return fmt.Errorf("Invalid key <%s> when migrating dispatcher profiles", id)
		}
		dpp, err := m.dmIN.DataManager().GetDispatcherProfile(tntID[0], tntID[1], false, false, utils.NonTransactional)
		if err != nil {
			return err
		}
		if dpp == nil || m.dryRun {
			continue
		}
		if err := m.dmOut.DataManager().SetDispatcherProfile(dpp, true); err != nil {
			return err
		}
		if err := m.dmIN.DataManager().RemoveDispatcherProfile(tntID[0],
			tntID[1], utils.NonTransactional, false); err != nil {
			return err
		}
		m.stats[utils.Dispatchers] += 1
	}
	return
}

func (m *Migrator) migrateCurrentDispatcherHost() (err error) {
	var ids []string
	ids, err = m.dmIN.DataManager().DataDB().GetKeysForPrefix(utils.DispatcherHostPrefix)
	if err != nil {
		return err
	}
	for _, id := range ids {
		tntID := strings.SplitN(strings.TrimPrefix(id, utils.DispatcherHostPrefix), utils.InInFieldSep, 2)
		if len(tntID) < 2 {
			return fmt.Errorf("Invalid key <%s> when migrating dispatcher hosts", id)
		}
		dpp, err := m.dmIN.DataManager().GetDispatcherHost(tntID[0], tntID[1], false, false, utils.NonTransactional)
		if err != nil {
			return err
		}
		if dpp == nil || m.dryRun {
			continue
		}
		if err := m.dmOut.DataManager().SetDispatcherHost(dpp); err != nil {
			return err
		}
		if err := m.dmIN.DataManager().RemoveDispatcherHost(tntID[0],
			tntID[1], utils.NonTransactional); err != nil {
			return err
		}
	}
	return
}

func (m *Migrator) migrateDispatchers() (err error) {
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
			"version number is not defined for DispatcherProfile model")
	}
	switch vrs[utils.Dispatchers] {
	case current[utils.Dispatchers]:
		if m.sameDataDB {
			break
		}
		if err = m.migrateCurrentDispatcher(); err != nil {
			return err
		}
		if err = m.migrateCurrentDispatcherHost(); err != nil {
			return err
		}
	}
	return m.ensureIndexesDataDB(engine.ColDpp, engine.ColDph)
}
