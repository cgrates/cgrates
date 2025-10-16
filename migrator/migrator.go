/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/

package migrator

import (
	"fmt"
	"log"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func NewMigrator(dataDBCfg *config.DbCfg, dmFrom, dmTo map[string]MigratorDataDB,
	dryRun, sameDataDB bool) (m *Migrator, err error) {
	stats := make(map[string]int)

	m = &Migrator{
		dataDBCfg:  dataDBCfg,
		dmTo:       dmTo,
		dmFrom:     dmFrom,
		dryRun:     dryRun,
		sameDataDB: sameDataDB,
		stats:      stats,
	}
	return m, err
}

type Migrator struct {
	dataDBCfg  *config.DbCfg
	dmFrom     map[string]MigratorDataDB
	dmTo       map[string]MigratorDataDB
	dryRun     bool
	sameDataDB bool
	stats      map[string]int
}

// GetINConn returns the IN MigratorDataDB where the provided itemID is stored.
// Returns the *default MigratorDataDB conn if it can't find the item's corresponding DB
func (m *Migrator) GetINConn(itemID string) (db MigratorDataDB, err error) {
	var ok bool
	dbConnID, ok := m.dataDBCfg.Items[itemID]
	if !ok {
		return nil, fmt.Errorf("couldn't find item with ID: <%v>", itemID)
	}
	if db, ok = m.dmFrom[dbConnID.DBConn]; !ok {
		// return *default db if DBConn is not found
		return m.dmFrom[utils.MetaDefault], nil
	}
	return
}

// GetOUTConn returns the OUT MigratorDataDB where the provided itemID is stored.
// Returns the *default MigratorDataDB conn if it can't find the item's corresponding DB
func (m *Migrator) GetOUTConn(itemID string) (db MigratorDataDB, err error) {
	var ok bool
	dbConnID, ok := m.dataDBCfg.Items[itemID]
	if !ok {
		return nil, fmt.Errorf("couldn't find item with ID: <%v>", itemID)
	}
	if db, ok = m.dmTo[dbConnID.DBConn]; !ok {
		// return *default db if DBConn is not found
		return m.dmTo[utils.MetaDefault], nil
	}
	return
}

// Migrate implements the tasks to migrate, used as a dispatcher to the individual methods
func (m *Migrator) Migrate(taskIDs []string) (err error, stats map[string]int) {
	stats = make(map[string]int)
	for _, taskID := range taskIDs {
		switch taskID {
		default: // unsupported taskID
			err = utils.NewCGRError(utils.Migrator,
				utils.MandatoryIEMissingCaps,
				utils.UnsupportedMigrationTask,
				fmt.Sprintf("task <%s> is not a supported migration task", taskID))
		case utils.MetaSetVersions:
			if m.dryRun {
				log.Print("Cannot dryRun SetVersions!")
				return
			}
			if dmTo, err := m.GetOUTConn(utils.CacheVersions); err != nil {
				return err, nil
			} else {
				if dataDB, _, err := dmTo.DataManager().DBConns().GetConn(utils.CacheVersions); err != nil {
					return err, nil
				} else if err = engine.OverwriteDBVersions(dataDB); err != nil {
					return utils.NewCGRError(utils.Migrator, utils.ServerErrorCaps, err.Error(),
						fmt.Sprintf("error: <%s> when seting versions for DataDB", err.Error())), nil
				}
			}
		case utils.MetaEnsureIndexes:
			mongoDBFound := false // track if no mongo DBs were found for case *ensure_indexes
			// since DataManager is the same in all OUT migratorDataDBs, taking the default
			// migratorDataDB will suffice
			for _, db := range m.dmTo[utils.MetaDefault].DataManager().DataDB() {
				if db.GetStorageType() == utils.MetaMongo {
					mgo := db.(*engine.MongoStorage)
					if err = mgo.EnsureIndexes(); err != nil {
						return
					}
					mongoDBFound = true
				}
			}
			if !mongoDBFound {
				log.Printf("The DataDB type has to be %s .\n ", utils.MetaMongo)
			}

		case utils.MetaStats:
			err = m.migrateStats()
		case utils.MetaFilters:
			err = m.migrateFilters()
		case utils.MetaAccounts:
			err = m.migrateAccounts()
		//only Move
		case utils.MetaChargers:
			err = m.migrateChargers()
			//TPs
		case utils.MetaLoadIDs:
			err = m.migrateLoadIDs()
			//DATADB ALL
		case utils.MetaDataDB:
			if err := m.migrateStats(); err != nil {
				log.Print("ERROR: ", utils.MetaStats, " ", err)
			}
			if err := m.migrateFilters(); err != nil {
				log.Print("ERROR: ", utils.MetaFilters, " ", err)
			}
			if err = m.migrateLoadIDs(); err != nil {
				log.Print("ERROR: ", utils.MetaLoadIDs, " ", err)
			}
			err = nil

		}
	}
	for k, v := range m.stats {
		stats[k] = v
	}
	return
}

func (m *Migrator) ensureIndexesDataDB(cols ...string) error {
	// since DataManager is the same in all OUT migratorDataDBs, taking the default
	// migratorDataDB will suffice
	for _, db := range m.dmTo[utils.MetaDefault].DataManager().DataDB() {
		if db.GetStorageType() == utils.MetaMongo {
			mgo := db.(*engine.MongoStorage)
			if err := mgo.EnsureIndexes(cols...); err != nil {
				return err
			}
		}
	}
	return nil
}

// closes all opened DBs
func (m *Migrator) Close() {
	if m.dmFrom != nil {
		for _, dm := range m.dmFrom {
			dm.close()
		}
	}
	if m.dmTo != nil {
		for _, dm := range m.dmTo {
			dm.close()
		}
	}
}
