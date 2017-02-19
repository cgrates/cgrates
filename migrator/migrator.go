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

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func NewMigrator(tpDB engine.RatingStorage, dataDB engine.AccountingStorage, dataDBType string, storDB engine.Storage, storDBType string) *Migrator {
	return &Migrator{tpDB: tpDB, dataDB: dataDB, dataDBType: dataDBType, storDB: storDB, storDBType: storDBType}
}

type Migrator struct {
	tpDB       engine.RatingStorage // ToDo: unify the databases when ready
	dataDB     engine.AccountingStorage
	dataDBType string
	storDB     engine.Storage
	storDBType string
}

// Migrate implements the tasks to migrate, used as a dispatcher to the individual methods
func (m *Migrator) Migrate(taskID string) (err error) {
	switch taskID {
	default: // unsupported taskID
		err = utils.NewCGRError(utils.Migrator,
			utils.MandatoryIEMissingCaps,
			utils.UnsupportedMigrationTask,
			fmt.Sprintf("task <%s> is not a supported migration task", taskID))
	case utils.MetaSetVersions:
		if err := m.storDB.SetVersions(engine.CurrentStorDBVersions()); err != nil {
			return utils.NewCGRError(utils.Migrator,
				utils.ServerErrorCaps,
				err.Error(),
				fmt.Sprintf("error: <%s> when updating CostDetails version into StorDB", err.Error()))
		}
	case utils.MetaCostDetails:
		err = m.migrateCostDetails()
	case utils.MetaAccounts:
		err = m.migrateAccounts()
	}
	return
}
