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
	ids, err = m.dmIN.DataManager().DataDB().GetKeysForPrefix(utils.TimingsPrefix, utils.EmptyString)
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
		m.stats[utils.Timing]++
	}
	return
}

func (m *Migrator) migrateTimings() (err error) {
	var vrs engine.Versions
	current := engine.CurrentDataDBVersions()
	if vrs, err = m.getVersions(utils.Timing); err != nil {
		return
	}
	switch version := vrs[utils.Timing]; version {
	default:
		return fmt.Errorf("Unsupported version %v", version)
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
