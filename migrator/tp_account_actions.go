// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package migrator

import (
	"fmt"

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
			if m.dryRun != true {
				if err := m.storDBOut.StorDB().SetTPAccountActions(accAct); err != nil {
					return err
				}
				for _, acc := range accAct {
					if err := m.storDBIn.StorDB().RemTpData(utils.TBLTPAccountActions, acc.TPid,
						map[string]string{"loadid": acc.LoadId, "tenant": acc.Tenant, "account": acc.Account}); err != nil {
						return err
					}
				}
				m.stats[utils.TpAccountActionsV] += 1
			}
		}
	}
	return
}

func (m *Migrator) migrateTPaccountacction() (err error) {
	var vrs engine.Versions
	current := engine.CurrentStorDBVersions()
	vrs, err = m.storDBOut.StorDB().GetVersions("")
	if err != nil {
		return utils.NewCGRError(utils.Migrator,
			utils.ServerErrorCaps,
			err.Error(),
			fmt.Sprintf("error: <%s> when querying oldDataDB for versions", err.Error()))
	} else if len(vrs) == 0 {
		return utils.NewCGRError(utils.Migrator,
			utils.MandatoryIEMissingCaps,
			utils.UndefinedVersion,
			"version number is not defined for ActionTriggers model")
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
