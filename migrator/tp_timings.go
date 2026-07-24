// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package migrator

import (
	"fmt"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func (m *Migrator) migrateCurrentTPTiming() (err error) {
	tpids, err := m.storDBIn.StorDB().GetTpIds(utils.TBLTPTimings)
	if err != nil {
		return err
	}

	for _, tpid := range tpids {
		ids, err := m.storDBIn.StorDB().GetTpTableIds(tpid, utils.TBLTPTimings,
			utils.TPDistinctIds{"tag"}, map[string]string{}, nil)
		if err != nil {
			return err
		}
		for _, id := range ids {
			tm, err := m.storDBIn.StorDB().GetTPTimings(tpid, id)
			if err != nil {
				return err
			}
			if tm != nil {
				if m.dryRun != true {
					if err := m.storDBOut.StorDB().SetTPTimings(tm); err != nil {
						return err
					}
					for _, timing := range tm {
						if err := m.storDBIn.StorDB().RemTpData(utils.TBLTPTimings,
							timing.TPid, map[string]string{"tag": timing.ID}); err != nil {
							return err
						}
					}
					m.stats[utils.TpTiming] += 1
				}
			}
		}
	}
	return
}

func (m *Migrator) migrateTpTimings() (err error) {
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
	switch vrs[utils.TpTiming] {
	case current[utils.TpTiming]:
		if m.sameStorDB {
			break
		}
		if err := m.migrateCurrentTPTiming(); err != nil {
			return err
		}
	}
	return m.ensureIndexesStorDB(utils.TBLTPTimings)
}
