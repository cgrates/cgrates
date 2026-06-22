/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/

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
