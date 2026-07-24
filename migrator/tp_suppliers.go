// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package migrator

import (
	"fmt"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func (m *Migrator) migrateCurrentTPSuppliers() (err error) {
	tpids, err := m.storDBIn.StorDB().GetTpIds(utils.TBLTPSuppliers)
	if err != nil {
		return err
	}

	for _, tpid := range tpids {
		ids, err := m.storDBIn.StorDB().GetTpTableIds(tpid, utils.TBLTPSuppliers,
			utils.TPDistinctIds{"id"}, map[string]string{}, nil)
		if err != nil {
			return err
		}
		for _, id := range ids {

			suppliers, err := m.storDBIn.StorDB().GetTPSuppliers(tpid, "", id)
			if err != nil {
				return err
			}
			if suppliers != nil {
				if m.dryRun != true {
					if err := m.storDBOut.StorDB().SetTPSuppliers(suppliers); err != nil {
						return err
					}
					for _, supplier := range suppliers {
						if err := m.storDBIn.StorDB().RemTpData(utils.TBLTPSuppliers, supplier.TPid,
							map[string]string{"tenant": supplier.Tenant, "id": supplier.ID}); err != nil {
							return err
						}
					}

					m.stats[utils.TpSuppliers] += 1
				}
			}
		}
	}
	return
}

func (m *Migrator) migrateTPSuppliers() (err error) {
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
	switch vrs[utils.TpSuppliers] {
	case current[utils.TpSuppliers]:
		if m.sameStorDB {
			break
		}
		if err := m.migrateCurrentTPSuppliers(); err != nil {
			return err
		}
	}
	return m.ensureIndexesStorDB(utils.TBLTPSuppliers)
}
