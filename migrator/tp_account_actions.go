// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package migrator

import (
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func (m *Migrator) migrateCurrentTPaccountAcction() (err error) {
	tpids, err := m.storDBIn.StorDB().GetTpIds(utils.TBLTPAccountActions)
	if err != nil {
		return err
	}
	for _, tpid := range tpids {
		accAct, err := m.storDBIn.StorDB().GetTPAccountActions(&utils.TPAccountActions{TPid: tpid})
		if err != nil {
			return err
		}
		if accAct != nil {
			if !m.dryRun {
				if err := m.storDBOut.StorDB().SetTPAccountActions(accAct); err != nil {
					return err
				}
				for _, acc := range accAct {
					if err := m.storDBIn.StorDB().RemTpData(utils.TBLTPAccountActions, acc.TPid,
						map[string]string{"loadid": acc.LoadId, "tenant": acc.Tenant, "account": acc.Account}); err != nil {
						return err
					}
				}
				m.stats[utils.TpAccountActionsV]++
			}
		}
	}
	return
}

func (m *Migrator) migrateTPaccountacction() (err error) {
	var vrs engine.Versions
	current := engine.CurrentStorDBVersions()
	if vrs, err = m.getVersions(utils.TpAccountActionsV); err != nil {
		return
	}
	switch vrs[utils.TpAccountActionsV] {
	case current[utils.TpAccountActionsV]:
		if m.sameStorDB {
			break
		}
		if err := m.migrateCurrentTPaccountAcction(); err != nil {
			return err
		}
	}
	return m.ensureIndexesStorDB(utils.TBLTPAccountActions)
}
