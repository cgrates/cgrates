// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package migrator

import (
	"fmt"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func (m *Migrator) migrateCurrentTPactionplans() (err error) {
	tpids, err := m.storDBIn.StorDB().GetTpIds(utils.TBLTPActionPlans)
	if err != nil {
		return err
	}
	for _, tpid := range tpids {
		ids, err := m.storDBIn.StorDB().GetTpTableIds(tpid, utils.TBLTPActionPlans,
			utils.TPDistinctIds{"tag"}, map[string]string{}, nil)
		if err != nil {
			return err
		}
		for _, id := range ids {
			actPln, err := m.storDBIn.StorDB().GetTPActionPlans(tpid, id)
			if err != nil {
				return err
			}
			if actPln != nil {
				if m.dryRun != true {
					if err := m.storDBOut.StorDB().SetTPActionPlans(actPln); err != nil {
						return err
					}
					for _, act := range actPln {
						if err := m.storDBIn.StorDB().RemTpData(utils.TBLTPActionPlans,
							act.TPid, map[string]string{"tag": act.ID}); err != nil {
							return err
						}
					}
					m.stats[utils.TpActionPlans] += 1
				}
			}
		}
	}
	return
}

func (m *Migrator) migrateTPactionplans() (err error) {
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
	switch vrs[utils.TpActionPlans] {
	case current[utils.TpActionPlans]:
		if m.sameStorDB {
			break
		}
		if err := m.migrateCurrentTPactionplans(); err != nil {
			return err
		}
	}
	return m.ensureIndexesStorDB(utils.TBLTPActionPlans)
}
