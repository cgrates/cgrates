// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package migrator

import (
	"fmt"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func (m *Migrator) migrateCurrentTPactions() (err error) {
	tpids, err := m.storDBIn.StorDB().GetTpIds(utils.TBLTPActions)
	if err != nil {
		return err
	}
	for _, tpid := range tpids {
		ids, err := m.storDBIn.StorDB().GetTpTableIds(tpid, utils.TBLTPActions, utils.TPDistinctIds{"tag"}, map[string]string{}, nil)
		if err != nil {
			return err
		}
		for _, id := range ids {
			action, err := m.storDBIn.StorDB().GetTPActions(tpid, id)
			if err != nil {
				return err
			}
			if action != nil {
				if m.dryRun != true {
					if err := m.storDBOut.StorDB().SetTPActions(action); err != nil {
						return err
					}
					for _, act := range action {
						if err := m.storDBIn.StorDB().RemTpData(utils.TBLTPActions, act.TPid, map[string]string{"tag": act.ID}); err != nil {
							return err
						}
					}
					m.stats[utils.TpActions] += 1
				}
			}
		}
	}
	return
}

func (m *Migrator) migrateTPactions() (err error) {
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
	switch vrs[utils.TpActions] {
	case current[utils.TpActions]:
		if m.sameStorDB {
			break
		}
		if err := m.migrateCurrentTPactions(); err != nil {
			return err
		}
	}
	return m.ensureIndexesStorDB(utils.TBLTPActions)
}
