// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package migrator

import (
	"fmt"
	"strings"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func (m *Migrator) migrateCurrentSupplierProfile() (err error) {
	var ids []string
	ids, err = m.dmIN.DataManager().DataDB().GetKeysForPrefix(utils.SupplierProfilePrefix)
	if err != nil {
		return err
	}
	for _, id := range ids {
		tntID := strings.SplitN(strings.TrimPrefix(id, utils.SupplierProfilePrefix), utils.InInFieldSep, 2)
		if len(tntID) < 2 {
			return fmt.Errorf("Invalid key <%s> when migrating supplier profiles", id)
		}
		splp, err := m.dmIN.DataManager().GetSupplierProfile(tntID[0], tntID[1], false, false, utils.NonTransactional)
		if err != nil {
			return err
		}
		if splp == nil || m.dryRun {
			continue
		}
		if err := m.dmOut.DataManager().SetSupplierProfile(splp, true); err != nil {
			return err
		}
		if err := m.dmIN.DataManager().RemoveSupplierProfile(tntID[0], tntID[1], utils.NonTransactional, true); err != nil {
			return err
		}
		m.stats[utils.Suppliers] += 1
	}
	return
}

func (m *Migrator) migrateSupplierProfiles() (err error) {
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
			"version number is not defined for SupplierProfiles model")
	}
	switch vrs[utils.Suppliers] {
	case current[utils.Suppliers]:
		if m.sameDataDB {
			break
		}
		if err = m.migrateCurrentSupplierProfile(); err != nil {
			return err
		}
	}
	return m.ensureIndexesDataDB(engine.ColSpp)
}
