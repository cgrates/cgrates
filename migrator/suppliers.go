/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>
*/

package migrator

import (
	"fmt"
	"strings"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func (m *Migrator) migrateCurrentSupplierProfile() (err error) {
	var ids []string
	tenant := config.CgrConfig().GeneralCfg().DefaultTenant
	ids, err = m.dmIN.DataManager().DataDB().GetKeysForPrefix(utils.SupplierProfilePrefix)
	if err != nil {
		return err
	}
	for _, id := range ids {
		idg := strings.TrimPrefix(id, utils.SupplierProfilePrefix)
		splp, err := m.dmIN.DataManager().GetSupplierProfile(tenant, idg, false, false, utils.NonTransactional)
		if err != nil {
			return err
		}
		if splp != nil {
			if m.dryRun != true {
				if err := m.dmOut.DataManager().SetSupplierProfile(splp, true); err != nil {
					return err
				}
				m.stats[utils.Suppliers] += 1
			}
		}
	}
	return
}

func (m *Migrator) migrateSupplierProfiles() (err error) {
	var vrs engine.Versions
	current := engine.CurrentDataDBVersions()
	vrs, err = m.dmOut.DataManager().DataDB().GetVersions("")
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
			return
		}
		if err := m.migrateCurrentSupplierProfile(); err != nil {
			return err
		}
		return
	}
	return
}
