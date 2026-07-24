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
	ids, err = m.dmIN.DataManager().DataDB().GetKeysForPrefix(utils.RATING_PLAN_PREFIX)
	if err != nil {
		return err
	}
	for _, id := range ids {
		idg := strings.TrimPrefix(id, utils.RATING_PLAN_PREFIX)
		rp, err := m.dmIN.DataManager().GetRatingPlan(idg, true, utils.NonTransactional)
		if err != nil {
			return err
		}
		if rp == nil || m.dryRun {
			continue
		}
		if err := m.dmOut.DataManager().SetRatingPlan(rp, utils.NonTransactional); err != nil {
			return err
		}
		if err := m.dmIN.DataManager().RemoveRatingPlan(idg, utils.NonTransactional); err != nil {
			return err
		}
		m.stats[utils.RatingPlan] += 1
	}
	return
}

func (m *Migrator) migrateRatingPlans() (err error) {
	var vrs engine.Versions
	current := engine.CurrentDataDBVersions()
	vrs, err = m.dmIN.DataManager().DataDB().GetVersions("")
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
	switch vrs[utils.RatingPlan] {
	case current[utils.RatingPlan]:
		if m.sameDataDB {
			break
		}
		if err = m.migrateCurrentRatingPlans(); err != nil {
			return err
		}
	}
	return m.ensureIndexesDataDB(engine.ColRpl)
}
