// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

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
