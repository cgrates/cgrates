// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package migrator

import (
	"fmt"
	"strings"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func (m *Migrator) migrateCurrentTiming() (err error) {
	var ids []string
	ids, err = m.dmIN.DataManager().DataDB().GetKeysForPrefix(utils.TimingsPrefix)
	if err != nil {
		return err
	}
	for _, id := range ids {
		idg := strings.TrimPrefix(id, utils.TimingsPrefix)
		tm, err := m.dmIN.DataManager().GetTiming(idg, true, utils.NonTransactional)
		if err != nil {
			return err
		}
		if tm == nil || m.dryRun {
			continue
		}
		if err := m.dmOut.DataManager().SetTiming(tm); err != nil {
			return err
		}
		m.stats[utils.Timing] += 1
	}
	return
}

func (m *Migrator) migrateTimings() (err error) {
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
	switch vrs[utils.Timing] {
	case current[utils.Timing]:
		if m.sameDataDB {
			break
		}
		if err = m.migrateCurrentTiming(); err != nil {
			return err
		}
	}
	return m.ensureIndexesDataDB(engine.ColTmg)
}
