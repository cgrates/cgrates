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
	"log"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func NewMigrator(dm *engine.DataManager, dataDBType, dataDBEncoding string, storDB engine.Storage, storDBType string, oldDataDB V1DataDB, oldDataDBType, oldDataDBEncoding string, oldStorDB engine.Storage, oldStorDBType string, dryRun bool) (m *Migrator, err error) {
	var mrshlr engine.Marshaler
	var oldmrshlr engine.Marshaler
	if dataDBEncoding == utils.MSGPACK {
		mrshlr = engine.NewCodecMsgpackMarshaler()
	} else if dataDBEncoding == utils.JSON {
		mrshlr = new(engine.JSONMarshaler)
	} else if oldDataDBEncoding == utils.MSGPACK {
		oldmrshlr = engine.NewCodecMsgpackMarshaler()
	} else if oldDataDBEncoding == utils.JSON {
		oldmrshlr = new(engine.JSONMarshaler)
	}
	stats := make(map[string]int)

	m = &Migrator{
		dm: dm, dataDBType: dataDBType,
		storDB: storDB, storDBType: storDBType, mrshlr: mrshlr,
		oldDataDB: oldDataDB, oldDataDBType: oldDataDBType,
		oldStorDB: oldStorDB, oldStorDBType: oldStorDBType,
		oldmrshlr: oldmrshlr, dryRun: dryRun, stats: stats,
	}
	return m, err
}

type Migrator struct {
	dm            *engine.DataManager
	dataDBType    string
	storDB        engine.Storage
	storDBType    string
	mrshlr        engine.Marshaler
	oldDataDB     V1DataDB
	oldDataDBType string
	oldStorDB     engine.Storage
	oldStorDBType string
	oldmrshlr     engine.Marshaler
	dryRun        bool
	stats         map[string]int
}

// Migrate implements the tasks to migrate, used as a dispatcher to the individual methods
func (m *Migrator) Migrate(taskIDs []string) (err error, stats map[string]int) {
	stats = make(map[string]int)
	for _, taskID := range taskIDs {
		log.Print("migrating", taskID)
		switch taskID {
		default: // unsupported taskID
			err = utils.NewCGRError(utils.Migrator,
				utils.MandatoryIEMissingCaps,
				utils.UnsupportedMigrationTask,
				fmt.Sprintf("task <%s> is not a supported migration task", taskID))
		case utils.MetaSetVersions:
			if m.dryRun != true {
				if err := m.storDB.SetVersions(engine.CurrentDBVersions(m.storDBType), true); err != nil {
					return utils.NewCGRError(utils.Migrator,
						utils.ServerErrorCaps,
						err.Error(),
						fmt.Sprintf("error: <%s> when updating CostDetails version into StorDB", err.Error())), nil
				}
				if err := m.dm.DataDB().SetVersions(engine.CurrentDBVersions(m.dataDBType), true); err != nil {
					return utils.NewCGRError(utils.Migrator,
						utils.ServerErrorCaps,
						err.Error(),
						fmt.Sprintf("error: <%s> when updating CostDetails version into StorDB", err.Error())), nil
				}
			} else {
				log.Print("Cannot dryRun SetVersions!")
			}
		case utils.MetaCostDetails:
			err = m.migrateCostDetails()
		case utils.MetaAccounts:
			err = m.migrateAccounts()
		case utils.MetaActionPlans:
			err = m.migrateActionPlans()
		case utils.MetaActionTriggers:
			err = m.migrateActionTriggers()
		case utils.MetaActions:
			err = m.migrateActions()
		case utils.MetaSharedGroups:
			err = m.migrateSharedGroups()
		case utils.MetaStats:
			err = m.migrateStats()
		case utils.MetaThresholds:
			err = m.migrateStats()
		}
	}
	for k, v := range m.stats {
		stats[k] = v
	}
	return
}
