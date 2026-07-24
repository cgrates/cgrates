// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package migrator

import (
	"fmt"
	"strings"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func (m *Migrator) migrateCurrentResource() (err error) {
	var ids []string
	ids, err = m.dmIN.DataManager().DataDB().GetKeysForPrefix(utils.ResourceProfilesPrefix)
	if err != nil {
		return err
	}
	for _, id := range ids {
		tntID := strings.SplitN(strings.TrimPrefix(id, utils.ResourceProfilesPrefix), utils.InInFieldSep, 2)
		if len(tntID) < 2 {
			return fmt.Errorf("Invalid key <%s> when migrating resource profiles", id)
		}
		res, err := m.dmIN.DataManager().GetResourceProfile(tntID[0], tntID[1], false, false, utils.NonTransactional)
		if err != nil {
			return err
		}
		if res == nil || m.dryRun {
			continue
		}
		if err := m.dmOut.DataManager().SetResourceProfile(res, true); err != nil {
			return err
		}
		if err := m.dmIN.DataManager().RemoveResourceProfile(tntID[0], tntID[1], utils.NonTransactional, false); err != nil {
			return err
		}
		m.stats[utils.Resource] += 1
	}
	return
}

func (m *Migrator) migrateResources() (err error) {
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
	switch vrs[utils.Resource] {
	case current[utils.Resource]:
		if m.sameDataDB {
			break
		}
		if err = m.migrateCurrentResource(); err != nil {
			return err
		}
	}
	return m.ensureIndexesDataDB(engine.ColRsP)
}
