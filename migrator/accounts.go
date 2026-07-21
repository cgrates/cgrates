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

func (m *Migrator) migrateCurrentAccounts() error {
	db, _, err := m.dmFrom.DBConns().GetConn(utils.MetaAccounts)
	if err != nil {
		return err
	}
	ids, err := db.GetKeysForPrefix(context.TODO(), utils.AccountPrefix, "")
	if err != nil {
		return err
	}
	for _, id := range ids {
		tntID := strings.SplitN(strings.TrimPrefix(id, utils.AccountPrefix), utils.InInFieldSep, 2)
		if len(tntID) < 2 {
			return fmt.Errorf("Invalid key <%s> when migrating from account ", id)
		}
		ap, err := m.dmFrom.GetAccount(context.TODO(), tntID[0], tntID[1])
		if err != nil {
			return err
		}
		if ap == nil || m.dryRun {
			continue
		}
		if err := m.dmTo.SetAccount(context.TODO(), ap, true); err != nil {
			return err
		}
		if err := m.dmFrom.RemoveAccount(context.TODO(), tntID[0], tntID[1], false); err != nil {
			return err
		}
		m.stats[utils.AccountsString]++
	}
	return nil
}

func (m *Migrator) migrateAccounts() error {
	vrs, err := m.getVersions(utils.AccountsString)
	if err != nil {
		return err
	}
	if vrs[utils.AccountsString] != engine.CurrentDataDBVersions()[utils.AccountsString] {
		return fmt.Errorf("Unsupported version %v", vrs[utils.AccountsString])
	}
	if !m.sameDataDB {
		if err = m.migrateCurrentAccounts(); err != nil {
			return err
		}
	}
	if err = m.setVersions(utils.AccountsString); err != nil {
		return err
	}
	return m.ensureIndexesDataDB(engine.ColApp)
}
