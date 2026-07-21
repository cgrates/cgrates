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

func (m *Migrator) migrateCurrentRequestFilter() error {
	dataDB, _, err := m.dmFrom.DBConns().GetConn(utils.MetaFilters)
	if err != nil {
		return err
	}
	ids, err := dataDB.GetKeysForPrefix(context.TODO(), utils.FilterPrefix, "")
	if err != nil {
		return err
	}
	for _, id := range ids {
		tntID := strings.SplitN(strings.TrimPrefix(id, utils.FilterPrefix), utils.InInFieldSep, 2)
		if len(tntID) < 2 {
			return fmt.Errorf("Invalid key <%s> when migrating filters", id)
		}
		fl, err := m.dmFrom.GetFilter(context.TODO(), tntID[0], tntID[1], false, false, utils.NonTransactional)
		if err != nil {
			return err
		}
		if m.dryRun || fl == nil {
			continue
		}
		if err := m.dmTo.SetFilter(context.TODO(), fl, true); err != nil {
			return err
		}
		if err := m.dmFrom.RemoveFilter(context.TODO(), tntID[0], tntID[1], true); err != nil {
			return err
		}
		m.stats[utils.RQF]++
	}
	return nil
}

func (m *Migrator) migrateFilters() error {
	vrs, err := m.getVersions(utils.RQF)
	if err != nil {
		return err
	}
	if vrs[utils.RQF] != engine.CurrentDataDBVersions()[utils.RQF] {
		return fmt.Errorf("Unsupported version %v", vrs[utils.RQF])
	}
	if m.sameDataDB {
		return nil
	}
	return m.migrateCurrentRequestFilter()
}
