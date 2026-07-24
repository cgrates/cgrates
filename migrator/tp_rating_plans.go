// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package migrator

import (
	"fmt"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func (m *Migrator) migrateCurrentTPratingplans() (err error) {
	tpids, err := m.storDBIn.StorDB().GetTpIds(utils.TBLTPRatingPlans)
	if err != nil {
		return err
	}
	for _, tpid := range tpids {
		ids, err := m.storDBIn.StorDB().GetTpTableIds(tpid, utils.TBLTPRatingPlans, utils.TPDistinctIds{"tag"}, map[string]string{}, nil)
		if err != nil {
			return err
		}
		if len(ids) != 0 {
			for _, id := range ids {
				ratingPlan, err := m.storDBIn.StorDB().GetTPRatingPlans(tpid, id, nil)
				if err != nil {
					return err
				}
				if ratingPlan != nil {
					if m.dryRun != true {
						if err := m.storDBOut.StorDB().SetTPRatingPlans(ratingPlan); err != nil {
							return err
						}
						for _, ratPln := range ratingPlan {
							if err := m.storDBIn.StorDB().RemTpData(utils.TBLTPRatingPlans, ratPln.TPid, map[string]string{"tag": ratPln.ID}); err != nil {
								return err
							}
						}
						m.stats[utils.TpRatingPlans] += 1
					}
				}
			}
		}
	}
	return
}

func (m *Migrator) migrateTPratingplans() (err error) {
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
	switch vrs[utils.TpRatingPlans] {
	case current[utils.TpRatingPlans]:
		if m.sameStorDB {
			break
		}
		if err := m.migrateCurrentTPratingplans(); err != nil {
			return err
		}
	}
	return m.ensureIndexesStorDB(utils.TBLTPRatingPlans)
}
