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

func NewMigrator(dmIN, dmOut MigratorDataDB,
	storDBIn, storDBOut MigratorStorDB,
	dryRun, sameDataDB, sameStorDB, sameOutDB bool) (m *Migrator, err error) {
	stats := make(map[string]int)
	m = &Migrator{
		dmOut:      dmOut,
		dmIN:       dmIN,
		storDBIn:   storDBIn,
		storDBOut:  storDBOut,
		dryRun:     dryRun,
		sameDataDB: sameDataDB,
		sameStorDB: sameStorDB,
		sameOutDB:  sameOutDB,
		stats:      stats,
	}
	return m, err
}

type Migrator struct {
	dmIN       MigratorDataDB
	dmOut      MigratorDataDB
	storDBIn   MigratorStorDB
	storDBOut  MigratorStorDB
	dryRun     bool
	sameDataDB bool
	sameStorDB bool
	sameOutDB  bool // needed in case we set version and we use same DataDB as StorDB to store the versions without overwriting them
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
			err = engine.OverwriteDBVersions(m.dmOut.DataManager().DataDB())
			if err != nil {
				return utils.NewCGRError(utils.Migrator, utils.ServerErrorCaps, err.Error(),
					fmt.Sprintf("error: <%s> when seting versions for DataDB", err.Error())), nil
			}
			if m.sameOutDB {
				err = engine.SetDBVersions(m.storDBOut.StorDB())
			} else {
				err = engine.OverwriteDBVersions(m.storDBOut.StorDB())
			}
			if err != nil {
				return utils.NewCGRError(utils.Migrator, utils.ServerErrorCaps, err.Error(),
					fmt.Sprintf("error: <%s> when seting versions for StorDB", err.Error())), nil
			}
		case utils.MetaEnsureIndexes:
			if m.storDBOut.StorDB().GetStorageType() == utils.Mongo {
				mgo := m.storDBOut.StorDB().(*engine.MongoStorage)
				if err = mgo.EnsureIndexes(); err != nil {
					return
				}
			} else {
				log.Printf("The StorDB type has to be %s .\n ", utils.Mongo)
			}

			if m.dmOut.DataManager().DataDB().GetStorageType() == utils.Mongo {
				mgo := m.dmOut.DataManager().DataDB().(*engine.MongoStorage)
				if err = mgo.EnsureIndexes(); err != nil {
					return
				}
			} else {
				log.Printf("The DataDB type has to be %s .\n ", utils.Mongo)
			}
		case utils.MetaCDRs:
			err = m.migrateCDRs()
		case utils.MetaStats:
			err = m.migrateStats()
		case utils.MetaThresholds:
			err = m.migrateThresholds()
		case utils.MetaAttributes:
			err = m.migrateAttributeProfile()
		case utils.MetaFilters:
			err = m.migrateFilters()
		case utils.MetaRoutes:
			err = m.migrateRouteProfiles()
		case utils.MetaAccountProfiles:
			err = m.migrateAccountProfiles()
		//only Move
		case utils.MetaActionProfiles:
			err = m.migrateActionProfiles()
		case utils.MetaDestinations:
			err = m.migrateDestinations()
		case utils.MetaReverseDestinations:
			err = m.migrateReverseDestinations()
		case utils.MetaTimings:
			err = m.migrateTimings()
		case utils.MetaResources:
			err = m.migrateResources()
		case utils.MetaRateProfiles:
			err = m.migrateRateProfiles()
		case utils.MetaSubscribers:
			err = m.migrateSubscribers()
		case utils.MetaChargers:
			err = m.migrateChargers()
		case utils.MetaDispatchers:
			err = m.migrateDispatchers()
			//TPs
		case utils.MetaTpFilters:
			err = m.migrateTPfilters()
		case utils.MetaTpThresholds:
			err = m.migrateTPthresholds()
		case utils.MetaTpRoutes:
			err = m.migrateTPRoutes()
		case utils.MetaTpStats:
			err = m.migrateTPstats()
		case utils.MetaTpRateProfiles:
			err = m.migrateTPRateProfiles()
		case utils.MetaTpActionProfiles:
			err = m.migrateTPActionProfiles()
		case utils.MetaTpResources:
			err = m.migrateTPresources()
		case utils.MetaTpTimings:
			err = m.migrateTpTimings()
		case utils.MetaTpDestinations:
			err = m.migrateTPDestinations()
		case utils.MetaTpChargers:
			err = m.migrateTPChargers()
		case utils.MetaTpDispatchers:
			err = m.migrateTPDispatchers()
		case utils.MetaLoadIDs:
			err = m.migrateLoadIDs()
			//DATADB ALL
		case utils.MetaDataDB:
			if err := m.migrateStats(); err != nil {
				log.Print("ERROR: ", utils.MetaStats, " ", err)
			}
			if err := m.migrateThresholds(); err != nil {
				log.Print("ERROR: ", utils.MetaThresholds, " ", err)
			}
			if err := m.migrateRouteProfiles(); err != nil {
				log.Print("ERROR: ", utils.MetaRoutes, " ", err)
			}
			if err := m.migrateAttributeProfile(); err != nil {
				log.Print("ERROR: ", utils.MetaAttributes, " ", err)
			}
			if err := m.migrateDestinations(); err != nil {
				log.Print("ERROR: ", utils.MetaDestinations, " ", err)
			}
			if err := m.migrateReverseDestinations(); err != nil {
				log.Print("ERROR: ", utils.MetaReverseDestinations, " ", err)
			}
			if err := m.migrateTimings(); err != nil {
				log.Print("ERROR: ", utils.MetaTimings, " ", err)
			}
			if err := m.migrateFilters(); err != nil {
				log.Print("ERROR: ", utils.MetaFilters, " ", err)
			}
			if err := m.migrateResources(); err != nil {
				log.Print("ERROR: ", utils.MetaResources, " ", err)
			}
			if err := m.migrateSubscribers(); err != nil {
				log.Print("ERROR: ", utils.MetaSubscribers, " ", err)
			}
			if err := m.migrateDispatchers(); err != nil {
				log.Print("ERROR: ", utils.MetaDispatchers, " ", err)
			}
			if err = m.migrateLoadIDs(); err != nil {
				log.Print("ERROR: ", utils.MetaLoadIDs, " ", err)
			}
			err = nil
			//STORDB ALL
		case utils.MetaStorDB:
			if err := m.migrateTPfilters(); err != nil {
				log.Print("ERROR: ", utils.MetaTpFilters, " ", err)
			}
			if err := m.migrateTPthresholds(); err != nil {
				log.Print("ERROR: ", utils.MetaTpThresholds, " ", err)
			}
			if err := m.migrateTPRoutes(); err != nil {
				log.Print("ERROR: ", utils.MetaTpRoutes, " ", err)
			}
			if err := m.migrateTPstats(); err != nil {
				log.Print("ERROR: ", utils.MetaTpStats, " ", err)
			}
			if err := m.migrateTPresources(); err != nil {
				log.Print("ERROR: ", utils.MetaTpResources, " ", err)
			}
			if err := m.migrateTpTimings(); err != nil {
				log.Print("ERROR: ", utils.MetaTpTimings, " ", err)
			}
			if err := m.migrateTPDestinations(); err != nil {
				log.Print("ERROR: ", utils.MetaTpDestinations, " ", err)
			}
			if err := m.migrateTPChargers(); err != nil {
				log.Print("ERROR: ", utils.MetaTpChargers, " ", err)
			}
			if err := m.migrateTPDispatchers(); err != nil {
				log.Print("ERROR: ", utils.MetaTpDispatchers, " ", err)
			}
			if err := m.migrateCDRs(); err != nil {
				log.Print("ERROR: ", utils.MetaCDRs, " ", err)
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
	if m.dmOut.DataManager().DataDB().GetStorageType() != utils.Mongo {
		return nil
	}
	mgo := m.dmOut.DataManager().DataDB().(*engine.MongoStorage)
	return mgo.EnsureIndexes(cols...)
}

func (m *Migrator) ensureIndexesStorDB(cols ...string) error {
	if m.storDBOut.StorDB().GetStorageType() != utils.Mongo {
		return nil
	}
	mgo := m.storDBOut.StorDB().(*engine.MongoStorage)
	return mgo.EnsureIndexes(cols...)
}

// closes all opened DBs
func (m *Migrator) Close() {
	if m.dmIN != nil {
		m.dmIN.close()
	}
	if m.dmOut != nil {
		m.dmOut.close()
	}
	if m.storDBIn != nil {
		m.storDBIn.close()
	}
	if m.storDBOut != nil {
		m.storDBOut.close()
	}
}
