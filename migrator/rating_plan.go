// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package migrator

import (
	"fmt"
	"strings"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func (m *Migrator) migrateCurrentRatingPlans() (err error) {
	var ids []string
	ids, err = m.dmIN.DataManager().DataDB().GetKeysForPrefix(utils.RatingPlanPrefix, utils.EmptyString)
	if err != nil {
		return err
	}
	for _, id := range ids {
		idg := strings.TrimPrefix(id, utils.RatingPlanPrefix)
		rp, err := m.dmIN.DataManager().GetRatingPlan(idg, true, utils.NonTransactional)
		if err != nil {
			return err
		}
		if rp == nil || m.dryRun {
			continue
		}
		if err := m.dmOut.DataManager().SetRatingPlan(rp); err != nil {
			return err
		}
		if err := m.dmIN.DataManager().RemoveRatingPlan(idg, utils.NonTransactional); err != nil {
			return err
		}
		m.stats[utils.RatingPlan]++
	}
	return
}

func (m *Migrator) migrateRatingPlans() (err error) {
	var vrs engine.Versions
	current := engine.CurrentDataDBVersions()
	if vrs, err = m.getVersions(utils.RatingPlan); err != nil {
		return
	}

	migrated := true
	for {
		version := vrs[utils.RatingPlan]
		for {
			switch version {
			default:
				return fmt.Errorf("Unsupported version %v", version)
			case current[utils.RatingPlan]:
				migrated = false
				if m.sameDataDB {
					break
				}
				if err = m.migrateCurrentRatingPlans(); err != nil {
					return
				}
			}
			if version == current[utils.RatingPlan] || err == utils.ErrNoMoreData {
				break
			}
		}
		if err == utils.ErrNoMoreData || !migrated {
			break
		}
		// if !m.dryRun {
		// if err = m.dmIN.DataManager().SetRatingPlan(v2, true); err != nil {
		// return
		// }
		// }
		m.stats[utils.RatingPlan]++
	}
	// All done, update version wtih current one
	if err = m.setVersions(utils.RatingPlan); err != nil {
		return
	}
	return m.ensureIndexesDataDB(engine.ColRpl)
}
