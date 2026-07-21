// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package migrator

import (
	"fmt"
	"strings"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func (m *Migrator) migrateCurrentCharger() error {
	dataDB, _, err := m.dmFrom.DBConns().GetConn(utils.MetaChargerProfiles)
	if err != nil {
		return err
	}
	ids, err := dataDB.GetKeysForPrefix(context.TODO(), utils.ChargerProfilePrefix, "")
	if err != nil {
		return err
	}
	for _, id := range ids {
		tntID := strings.SplitN(strings.TrimPrefix(id, utils.ChargerProfilePrefix), utils.InInFieldSep, 2)
		if len(tntID) < 2 {
			return fmt.Errorf("Invalid key <%s> when migrating chargers", id)
		}
		cpp, err := m.dmFrom.GetChargerProfile(context.TODO(), tntID[0], tntID[1], false, false, utils.NonTransactional)
		if err != nil {
			return err
		}
		if cpp == nil || m.dryRun {
			continue
		}
		if err := m.dmTo.SetChargerProfile(context.TODO(), cpp, true); err != nil {
			return err
		}
		if err := m.dmFrom.RemoveChargerProfile(context.TODO(), tntID[0], tntID[1], false); err != nil {
			return err
		}
		m.stats[utils.Chargers]++
	}
	return nil
}

func (m *Migrator) migrateChargers() error {
	vrs, err := m.getVersions(utils.Chargers)
	if err != nil {
		return err
	}
	if vrs[utils.Chargers] != engine.CurrentDataDBVersions()[utils.Chargers] {
		return fmt.Errorf("Unsupported version %v", vrs[utils.Chargers])
	}
	if !m.sameDataDB {
		if err = m.migrateCurrentCharger(); err != nil {
			return err
		}
	}
	if err = m.setVersions(utils.Chargers); err != nil {
		return err
	}
	return m.ensureIndexesDataDB(engine.ColCpp)
}
