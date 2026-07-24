// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package migrator

import (
	"fmt"
	"strings"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func (m *Migrator) migrateCurrentCharger() (err error) {
	var ids []string
	ids, err = m.dmIN.DataManager().DataDB().GetKeysForPrefix(utils.ChargerProfilePrefix)
	if err != nil {
		return err
	}
	for _, id := range ids {
		tntID := strings.SplitN(strings.TrimPrefix(id, utils.ChargerProfilePrefix), utils.InInFieldSep, 2)
		if len(tntID) < 2 {
			return fmt.Errorf("Invalid key <%s> when migrating chargers", id)
		}
		cpp, err := m.dmIN.DataManager().GetChargerProfile(tntID[0], tntID[1], false, false, utils.NonTransactional)
		if err != nil {
			return err
		}
		if cpp == nil || m.dryRun {
			continue
		}
		if err := m.dmOut.DataManager().SetChargerProfile(cpp, true); err != nil {
			return err
		}
		if err := m.dmIN.DataManager().RemoveChargerProfile(tntID[0],
			tntID[1], utils.NonTransactional, false); err != nil {
			return err
		}
		m.stats[utils.Chargers] += 1
	}
	return
}

func (m *Migrator) migrateChargers() (err error) {
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
			"version number is not defined for ChargerProfile model")
	}
	switch vrs[utils.Chargers] {
	case current[utils.Chargers]:
		if m.sameDataDB {
			break
		}
		if err = m.migrateCurrentCharger(); err != nil {
			return err
		}
	}
	return m.ensureIndexesDataDB(engine.ColCpp)
}
