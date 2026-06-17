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

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func NewMigrator(dmFrom, dmTo *engine.DataManager, dryRun, sameDataDB bool) *Migrator {
	return &Migrator{
		dmFrom:     dmFrom,
		dmTo:       dmTo,
		dryRun:     dryRun,
		sameDataDB: sameDataDB,
		stats:      make(map[string]int),
	}
}

type Migrator struct {
	dmFrom     *engine.DataManager
	dmTo       *engine.DataManager
	dryRun     bool
	sameDataDB bool
	stats      map[string]int
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
			dataDB, _, err := m.dmTo.DBConns().GetConn(utils.CacheVersions)
			if err != nil {
				return err, nil
			}
			if err = engine.OverwriteDBVersions(dataDB); err != nil {
				return utils.NewCGRError(utils.Migrator, utils.ServerErrorCaps, err.Error(),
					fmt.Sprintf("error: <%s> when seting versions for DataDB", err.Error())), nil
			}
		case utils.MetaEnsureIndexes:
			mongoDBFound := false // track if no mongo DBs were found for case *ensureIndexes
			for _, db := range m.dmTo.DB() {
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
	for _, db := range m.dmTo.DB() {
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
		for _, db := range m.dmFrom.DB() {
			db.Close()
		}
	}
	if m.dmTo != nil && !m.sameDataDB {
		for _, db := range m.dmTo.DB() {
			db.Close()
		}
	}
}
