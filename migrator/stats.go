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

func (m *Migrator) migrateCurrentStats() error {
	dataDB, _, err := m.dmFrom.DBConns().GetConn(utils.MetaStatQueueProfiles)
	if err != nil {
		return err
	}
	ids, err := dataDB.GetKeysForPrefix(context.Background(), utils.StatQueueProfilePrefix, "")
	if err != nil {
		return err
	}
	for _, id := range ids {
		tntID := strings.SplitN(strings.TrimPrefix(id, utils.StatQueueProfilePrefix), utils.InInFieldSep, 2)
		if len(tntID) < 2 {
			return fmt.Errorf("Invalid key <%s> when migrating stat queue profiles", id)
		}
		sqp, err := m.dmFrom.GetStatQueueProfile(context.TODO(), tntID[0], tntID[1], false, false, utils.NonTransactional)
		if err != nil {
			return err
		}
		sgs, err := m.dmFrom.GetStatQueue(context.TODO(), tntID[0], tntID[1], false, false, utils.NonTransactional)
		if err != nil {
			return err
		}
		if sqp == nil || m.dryRun {
			continue
		}
		if err := m.dmTo.SetStatQueueProfile(context.TODO(), sqp, true); err != nil {
			return err
		}
		if sgs != nil {
			if err := m.dmTo.SetStatQueue(context.TODO(), sgs); err != nil {
				return err
			}
		}
		if err := m.dmFrom.RemoveStatQueueProfile(context.TODO(), tntID[0], tntID[1], false); err != nil {
			return err
		}
		m.stats[utils.Stats]++
	}
	return nil
}

func (m *Migrator) migrateStats() error {
	vrs, err := m.getVersions(utils.Stats)
	if err != nil {
		return err
	}
	if vrs[utils.Stats] != engine.CurrentDataDBVersions()[utils.Stats] {
		return fmt.Errorf("Unsupported version %v", vrs[utils.Stats])
	}
	if m.sameDataDB {
		return nil
	}
	return m.migrateCurrentStats()
}
