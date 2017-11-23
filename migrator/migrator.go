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

func NewMigrator(dmIN *engine.DataManager, dmOut *engine.DataManager, dataDBType, dataDBEncoding string, storDB engine.Storage, storDBType string, oldDataDB MigratorDataDB, oldDataDBType, oldDataDBEncoding string, oldStorDB engine.Storage, oldStorDBType string, dryRun bool, sameDBname bool) (m *Migrator, err error) {
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
		dmOut: dmOut, dataDBType: dataDBType,
		storDB: storDB, storDBType: storDBType,
		mrshlr: mrshlr, dmIN: dmIN,
		oldDataDB: oldDataDB, oldDataDBType: oldDataDBType,
		oldStorDB: oldStorDB, oldStorDBType: oldStorDBType,
		oldmrshlr: oldmrshlr, dryRun: dryRun, sameDBname: sameDBname, stats: stats,
	}
	return m, err
}

type Migrator struct {
	dmIN          *engine.DataManager //oldatadb
	dmOut         *engine.DataManager
	dataDBType    string
	storDB        engine.Storage
	storDBType    string
	mrshlr        engine.Marshaler
	oldDataDB     MigratorDataDB
	oldDataDBType string
	oldStorDB     engine.Storage
	oldStorDBType string
	oldmrshlr     engine.Marshaler
	dryRun        bool
	sameDBname    bool
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
				if err := m.dmOut.DataDB().SetVersions(engine.CurrentDBVersions(m.dataDBType), true); err != nil {
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
			err = m.migrateThresholds()
		//only Move
		case utils.MetaRatingPlans:
			err = m.migrateRatingPlans()
		case utils.MetaRatingProfile:
			err = m.migrateRatingProfiles()
		case utils.MetaDestinations:
			err = m.migrateDestinations()
		case utils.MetaReverseDestinations:
			err = m.migrateReverseDestinations()
		case utils.MetaLCR:
			err = m.migrateLCR()
		case utils.MetaCdrStats:
			err = m.migrateCdrStats()
		case utils.MetaTiming:
			err = m.migrateTimings()
		case utils.MetaRQF:
			err = m.migrateRequestFilter()
		case utils.MetaResource:
			err = m.migrateResources()
		case utils.MetaReverseAlias:
			err = m.migrateReverseAlias()
		case utils.MetaAlias:
			err = m.migrateAlias()
		case utils.MetaUser:
			err = m.migrateUser()
		case utils.MetaSubscribers:
			err = m.migrateSubscribers()
		case utils.MetaDerivedChargersV:
			err = m.migrateDerivedChargers()
			//TPS
		case utils.MetaTpRatingPlans:
			err = m.migrateTPratingplans()
		case utils.MetaTpLcrs:
			err = m.migrateTPlcrs()
		case utils.MetaTpFilters:
			err = m.migrateTPfilters()
		case utils.MetaTpDestinationRates:
			err = m.migrateTPdestinationrates()
		case utils.MetaTpActionTriggers:
			err = m.migrateTPactiontriggers()
		case utils.MetaTpAccountActions:
			err = m.migrateTPaccountacction()
		case utils.MetaTpActionPlans:
			err = m.migrateTPactionplans()
		case utils.MetaTpActions:
			err = m.migrateTPactions()
		case utils.MetaTpDerivedCharges:
			err = m.migrateTPderivedchargers()
		case utils.MetaTpThresholds:
			err = m.migrateTPthresholds()
		case utils.MetaTpStats:
			err = m.migrateTPstats()
		case utils.MetaTpSharedGroups:
			err = m.migrateTPsharedgroups()
		case utils.MetaTpRatingProfiles:
			err = m.migrateTPratingprofiles()
		case utils.MetaTpResources:
			err = m.migrateTPresources()
		case utils.MetaTpRates:
			err = m.migrateTPrates()
		case utils.MetaTpTiming:
			err = m.migrateTpTimings()
		case utils.MetaTpAliases:
			err = m.migrateTPaliases()
		case utils.MetaTpUsers:
			err = m.migrateTPusers()
		case utils.MetaTpDerivedChargersV:
			err = m.migrateTPderivedchargers()
		case utils.MetaTpCdrStats:
			err = m.migrateTPcdrstats()
		case utils.MetaTpDestinations:
			err = m.migrateTPDestinations()
			//DATADB ALL
		case utils.MetaDataDB:
			if err := m.migrateAccounts(); err != nil {
				log.Print("GOT ", utils.MetaAccounts, " ", err)
			}
			if err := m.migrateActionPlans(); err != nil {
				log.Print("GOT ", utils.MetaActionPlans, " ", err)
			}
			if err := m.migrateActionTriggers(); err != nil {
				log.Print("GOT ", utils.MetaActionTriggers, " ", err)
			}
			if err := m.migrateActions(); err != nil {
				log.Print("GOT ", utils.MetaActions, " ", err)
			}
			if err := m.migrateSharedGroups(); err != nil {
				log.Print("GOT ", utils.MetaSharedGroups, " ", err)
			}
			if err := m.migrateStats(); err != nil {
				log.Print("GOT ", utils.MetaStats, " ", err)
			}
			if err := m.migrateThresholds(); err != nil {
				log.Print("GOT ", utils.MetaThresholds, " ", err)
			}
			if err := m.migrateRatingPlans(); err != nil {
				log.Print("GOT ", utils.MetaRatingPlans, " ", err)
			}
			if err := m.migrateRatingProfiles(); err != nil {
				log.Print("GOT ", utils.MetaRatingProfile, " ", err)
			}
			if err := m.migrateDestinations(); err != nil {
				log.Print("GOT ", utils.MetaDestinations, " ", err)
			}
			if err := m.migrateReverseDestinations(); err != nil {
				log.Print("GOT ", utils.MetaReverseDestinations, " ", err)
			}
			if err := m.migrateLCR(); err != nil {
				log.Print("GOT ", utils.MetaLCR, " ", err)
			}
			if err := m.migrateCdrStats(); err != nil {
				log.Print("GOT ", utils.MetaCdrStats, " ", err)
			}
			if err := m.migrateTimings(); err != nil {
				log.Print("GOT ", utils.MetaTiming, " ", err)
			}
			if err := m.migrateRequestFilter(); err != nil {
				log.Print("GOT ", utils.MetaRQF, " ", err)
			}
			if err := m.migrateResources(); err != nil {
				log.Print("GOT ", utils.MetaResource, " ", err)
			}
			if err := m.migrateReverseAlias(); err != nil {
				log.Print("GOT ", utils.MetaReverseAlias, " ", err)
			}
			if err := m.migrateAlias(); err != nil {
				log.Print("GOT ", utils.MetaAlias, " ", err)
			}
			if err := m.migrateUser(); err != nil {
				log.Print("GOT ", utils.MetaUser, " ", err)
			}
			if err := m.migrateSubscribers(); err != nil {
				log.Print("GOT ", utils.MetaSubscribers, " ", err)
			}
			if err := m.migrateDerivedChargers(); err != nil {
				log.Print("GOT ", utils.MetaDerivedChargersV, " ", err)
			}
			err = nil
			//STORDB ALL
		case utils.MetaStorDB:
			if err := m.migrateTPratingplans(); err != nil {
				log.Print("GOT ", utils.MetaTpRatingPlans, " ", err)
			}
			if err := m.migrateTPlcrs(); err != nil {
				log.Print("GOT ", utils.MetaTpLcrs, " ", err)
			}
			if err := m.migrateTPfilters(); err != nil {
				log.Print("GOT ", utils.MetaTpFilters, " ", err)
			}
			if err := m.migrateTPdestinationrates(); err != nil {
				log.Print("GOT ", utils.MetaTpDestinationRates, " ", err)
			}
			if err := m.migrateTPactiontriggers(); err != nil {
				log.Print("GOT ", utils.MetaTpActionTriggers, " ", err)
			}
			if err := m.migrateTPaccountacction(); err != nil {
				log.Print("GOT ", utils.MetaTpAccountActions, " ", err)
			}
			if err := m.migrateTPactionplans(); err != nil {
				log.Print("GOT ", utils.MetaTpActionPlans, " ", err)
			}
			if err := m.migrateTPactions(); err != nil {
				log.Print("GOT ", utils.MetaTpActions, " ", err)
			}
			if err := m.migrateTPderivedchargers(); err != nil {
				log.Print("GOT ", utils.MetaTpDerivedCharges, " ", err)
			}
			if err := m.migrateTPthresholds(); err != nil {
				log.Print("GOT ", utils.MetaTpThresholds, " ", err)
			}
			if err := m.migrateTPstats(); err != nil {
				log.Print("GOT ", utils.MetaTpStats, " ", err)
			}
			if err := m.migrateTPsharedgroups(); err != nil {
				log.Print("GOT ", utils.MetaTpSharedGroups, " ", err)
			}
			if err := m.migrateTPratingprofiles(); err != nil {
				log.Print("GOT ", utils.MetaTpRatingProfiles, " ", err)
			}
			if err := m.migrateTPresources(); err != nil {
				log.Print("GOT ", utils.MetaTpResources, " ", err)
			}
			if err := m.migrateTPrates(); err != nil {
				log.Print("GOT ", utils.MetaTpRates, " ", err)
			}
			if err := m.migrateTpTimings(); err != nil {
				log.Print("GOT ", utils.MetaTpTiming, " ", err)
			}
			if err := m.migrateTPaliases(); err != nil {
				log.Print("GOT ", utils.MetaTpAliases, " ", err)
			}
			if err := m.migrateTPusers(); err != nil {
				log.Print("GOT ", utils.MetaTpUsers, " ", err)
			}
			if err := m.migrateTPderivedchargers(); err != nil {
				log.Print("GOT ", utils.MetaTpDerivedChargersV, " ", err)
			}
			if err := m.migrateTPcdrstats(); err != nil {
				log.Print("GOT ", utils.MetaTpCdrStats, " ", err)
			}
			if err := m.migrateTPDestinations(); err != nil {
				log.Print("GOT ", utils.MetaTpDestinations, " ", err)
			}
			err = nil
		}
	}
	for k, v := range m.stats {
		stats[k] = v
	}
	return
}
