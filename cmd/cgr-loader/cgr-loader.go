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

package main

import (
	"flag"
	"fmt"
	"log"
	"path"
	"strings"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

var (
	dfltCfg = config.CgrConfig()
	cfgDir  = flag.String("config_dir", "",
		"Configuration directory path.")

	dataDBType = flag.String("datadb_type", dfltCfg.DataDbType,
		"The type of the DataDB database <*redis|*mongo>")
	dataDBHost = flag.String("datadb_host", dfltCfg.DataDbHost,
		"The DataDb host to connect to.")
	dataDBPort = flag.String("datadb_port", dfltCfg.DataDbPort,
		"The DataDb port to bind to.")
	dataDBName = flag.String("datadb_name", dfltCfg.DataDbName,
		"The name/number of the DataDb to connect to.")
	dataDBUser = flag.String("datadb_user", dfltCfg.DataDbUser,
		"The DataDb user to sign in as.")
	dataDBPasswd = flag.String("datadb_passwd", dfltCfg.DataDbPass,
		"The DataDb user's password.")

	storDBType = flag.String("stordb_type", dfltCfg.StorDBType,
		"The type of the storDb database <*mysql|*postgres|*mongo>")
	storDBHost = flag.String("stordb_host", dfltCfg.StorDBHost,
		"The storDb host to connect to.")
	storDBPort = flag.String("stordb_port", dfltCfg.StorDBPort,
		"The storDb port to bind to.")
	storDBName = flag.String("stordb_name", dfltCfg.StorDBName,
		"The name/number of the storDb to connect to.")
	storDBUser = flag.String("stordb_user", dfltCfg.StorDBUser,
		"The storDb user to sign in as.")
	storDBPasswd = flag.String("stordb_passwd", dfltCfg.StorDBPass,
		"The storDb user's password.")

	dbDataEncoding = flag.String("dbdata_encoding", dfltCfg.DBDataEncoding,
		"The encoding used to store object data in strings")

	flush = flag.Bool("flushdb", false,
		"Flush the database before importing")
	tpid = flag.String("tpid", dfltCfg.LoaderCgrConfig.TpID,
		"The tariff plan ID from the database")
	dataPath = flag.String("path", dfltCfg.LoaderCgrConfig.DataPath,
		"The path to folder containing the data files")
	version = flag.Bool("version", false,
		"Prints the application version.")
	verbose = flag.Bool("verbose", false,
		"Enable detailed verbose logging output")
	dryRun = flag.Bool("dry_run", false,
		"When true will not save loaded data to dataDb but just parse it for consistency and errors.")
	//validate = flag.Bool("validate", false,
	//	"When true will run various check on the loaded data to check for structural errors")

	fromStorDB    = flag.Bool("from_stordb", false, "Load the tariff plan from storDb to dataDb")
	toStorDB      = flag.Bool("to_stordb", false, "Import the tariff plan from files to storDb")
	rpcEncoding   = flag.String("rpc_encoding", utils.MetaJSONrpc, "RPC encoding used <gob|json>")
	cacheSAddress = flag.String("caches_address", dfltCfg.LoaderCgrConfig.CachesConns[0].Address,
		"CacheS component to contact for cache reloads, empty to disable automatic cache reloads")
	schedulerAddress = flag.String("scheduler_address", dfltCfg.LoaderCgrConfig.SchedulerConns[0].Address, "")

	importID       = flag.String("import_id", "", "Uniquely identify an import/load, postpended to some automatic fields")
	timezone       = flag.String("timezone", "", `Timezone for timestamps where not specified <""|UTC|Local|$IANA_TZ_DB>`)
	disableReverse = flag.Bool("disable_reverse_mappings", false, "Will disable reverse mappings rebuilding")
	flushStorDB    = flag.Bool("flush_stordb", false, "Remove tariff plan data for id from the database")
	remove         = flag.Bool("remove", false, "Will remove instead of adding data from DB")

	usersAddress = flag.String("users", "", "Users service to contact for data reloads, empty to disable automatic data reloads")

	err           error
	dm            *engine.DataManager
	storDb        engine.LoadStorage
	cacheS, userS rpcclient.RpcClientConnection
	loader        engine.LoadReader
)

func main() {
	flag.Parse()
	if *version {
		fmt.Println(utils.GetCGRVersion())
		return
	}

	ldrCfg := config.CgrConfig()
	if *cfgDir != "" {
		if ldrCfg, err = config.NewCGRConfigFromFolder(*cfgDir); err != nil {
			log.Fatalf("Error loading config file %s", err.Error())
		}
	}

	if *dataDBType != dfltCfg.DataDbType {
		ldrCfg.DataDbType = *dataDBType
	}

	if *dataDBHost != dfltCfg.DataDbHost {
		ldrCfg.DataDbHost = *dataDBHost
	}

	if *dataDBPort != dfltCfg.DataDbPort {
		ldrCfg.DataDbPort = *dataDBPort
	}

	if *dataDBName != dfltCfg.DataDbName {
		ldrCfg.DataDbName = *dataDBName
	}

	if *dataDBUser != dfltCfg.DataDbUser {
		ldrCfg.DataDbUser = *dataDBUser
	}

	if *dataDBPasswd != dfltCfg.DataDbPass {
		ldrCfg.DataDbPass = *dataDBPasswd
	}

	if *storDBType != dfltCfg.StorDBType {
		ldrCfg.StorDBType = *storDBType
	}

	if *storDBHost != dfltCfg.StorDBHost {
		ldrCfg.StorDBHost = *storDBHost
	}

	if *storDBPort != dfltCfg.StorDBPort {
		ldrCfg.StorDBPort = *storDBPort
	}

	if *storDBName != dfltCfg.StorDBName {
		ldrCfg.StorDBName = *storDBName
	}

	if *storDBUser != dfltCfg.StorDBUser {
		ldrCfg.StorDBUser = *storDBUser
	}

	if *storDBPasswd != "" {
		ldrCfg.StorDBPass = *storDBPasswd
	}

	if *dbDataEncoding != "" {
		ldrCfg.DBDataEncoding = *dbDataEncoding
	}

	if *tpid != "" {
		ldrCfg.LoaderCgrConfig.TpID = *tpid
	}

	if *dataPath != "" {
		ldrCfg.LoaderCgrConfig.DataPath = *dataPath
	}

	if *cacheSAddress != dfltCfg.LoaderCgrConfig.CachesConns[0].Address {
		ldrCfg.LoaderCgrConfig.CachesConns = make([]*config.HaPoolConfig, 0)
		if *cacheSAddress != "" {
			ldrCfg.LoaderCgrConfig.CachesConns = append(ldrCfg.LoaderCgrConfig.CachesConns,
				&config.HaPoolConfig{Address: *cacheSAddress})
		}
	}

	if *schedulerAddress != dfltCfg.LoaderCgrConfig.SchedulerConns[0].Address {
		ldrCfg.LoaderCgrConfig.SchedulerConns = make([]*config.HaPoolConfig, 0)
		if *schedulerAddress != "" {
			ldrCfg.LoaderCgrConfig.SchedulerConns = append(ldrCfg.LoaderCgrConfig.SchedulerConns,
				&config.HaPoolConfig{Address: *schedulerAddress})
		}
	}

	if *rpcEncoding != dfltCfg.LoaderCgrConfig.CachesConns[0].Transport &&
		len(ldrCfg.LoaderCgrConfig.CachesConns) != 0 {
		ldrCfg.LoaderCgrConfig.CachesConns[0].Transport = *rpcEncoding
	}

	if *importID == "" {
		*importID = utils.UUIDSha1Prefix()
	}

	if *timezone != dfltCfg.DefaultTimezone {
		ldrCfg.DefaultTimezone = *timezone
	}

	if *disableReverse != dfltCfg.LoaderCgrConfig.DisableReverse {
		ldrCfg.LoaderCgrConfig.DisableReverse = *disableReverse
	}

	if !*toStorDB {
		if dm, err = engine.ConfigureDataStorage(ldrCfg.DataDbType, ldrCfg.DataDbHost,
			ldrCfg.DataDbPort, ldrCfg.DataDbName,
			ldrCfg.DataDbUser, ldrCfg.DataDbPass, ldrCfg.DBDataEncoding,
			config.CgrConfig().CacheCfg(), ldrCfg.LoadHistorySize); err != nil {
			log.Fatalf("Coud not open dataDB connection: %s", err.Error())
		}
		defer dm.DataDB().Close()
	}

	if *fromStorDB || *toStorDB {
		if storDb, err = engine.ConfigureLoadStorage(ldrCfg.StorDBType, ldrCfg.StorDBHost, ldrCfg.StorDBPort,
			ldrCfg.StorDBName, ldrCfg.StorDBUser, ldrCfg.StorDBPass, ldrCfg.DBDataEncoding,
			config.CgrConfig().StorDBMaxOpenConns, config.CgrConfig().StorDBMaxIdleConns,
			config.CgrConfig().StorDBConnMaxLifetime, config.CgrConfig().StorDBCDRSIndexes); err != nil {
			log.Fatalf("Coud not open storDB connection: %s", err.Error())
		}
		defer storDb.Close()
	}

	if !*dryRun {
		//tpid_remove

		if *toStorDB { // Import files from a directory into storDb
			if ldrCfg.LoaderCgrConfig.TpID == "" {
				log.Fatal("TPid required.")
			}
			if *flushStorDB {
				if err = storDb.RemTpData("", ldrCfg.LoaderCgrConfig.TpID, map[string]string{}); err != nil {
					log.Fatal(err)
				}
			}
			csvImporter := engine.TPCSVImporter{
				TPid:     ldrCfg.LoaderCgrConfig.TpID,
				StorDb:   storDb,
				DirPath:  *dataPath,
				Sep:      ',',
				Verbose:  *verbose,
				ImportId: *importID,
			}
			if errImport := csvImporter.Run(); errImport != nil {
				log.Fatal(errImport)
			}
			return
		}
	}

	if *fromStorDB { // Load Tariff Plan from storDb into dataDb
		loader = storDb
	} else { // Default load from csv files to dataDb
		loader = engine.NewFileCSVStorage(',',
			path.Join(*dataPath, utils.DESTINATIONS_CSV),
			path.Join(*dataPath, utils.TIMINGS_CSV),
			path.Join(*dataPath, utils.RATES_CSV),
			path.Join(*dataPath, utils.DESTINATION_RATES_CSV),
			path.Join(*dataPath, utils.RATING_PLANS_CSV),
			path.Join(*dataPath, utils.RATING_PROFILES_CSV),
			path.Join(*dataPath, utils.SHARED_GROUPS_CSV),
			path.Join(*dataPath, utils.LCRS_CSV),
			path.Join(*dataPath, utils.ACTIONS_CSV),
			path.Join(*dataPath, utils.ACTION_PLANS_CSV),
			path.Join(*dataPath, utils.ACTION_TRIGGERS_CSV),
			path.Join(*dataPath, utils.ACCOUNT_ACTIONS_CSV),
			path.Join(*dataPath, utils.DERIVED_CHARGERS_CSV),
			path.Join(*dataPath, utils.CDR_STATS_CSV),
			path.Join(*dataPath, utils.USERS_CSV),
			path.Join(*dataPath, utils.ALIASES_CSV),
			path.Join(*dataPath, utils.ResourcesCsv),
			path.Join(*dataPath, utils.StatsCsv),
			path.Join(*dataPath, utils.ThresholdsCsv),
			path.Join(*dataPath, utils.FiltersCsv),
			path.Join(*dataPath, utils.SuppliersCsv),
			path.Join(*dataPath, utils.AttributesCsv),
		)
	}

	tpReader := engine.NewTpReader(dm.DataDB(), loader,
		ldrCfg.LoaderCgrConfig.TpID, ldrCfg.DefaultTimezone)

	if err = tpReader.LoadAll(); err != nil {
		log.Fatal(err)
	}
	if *verbose {
		tpReader.ShowStatistics()
	}
	if *dryRun { // We were just asked to parse the data, not saving it
		return
	}
	if len(ldrCfg.LoaderCgrConfig.CachesConns) != 0 { // Init connection to CacheS so we can reload it's data
		if cacheS, err = rpcclient.NewRpcClient("tcp",
			ldrCfg.LoaderCgrConfig.CachesConns[0].Address, 3, 3,
			time.Duration(1*time.Second), time.Duration(5*time.Minute),
			strings.TrimPrefix(ldrCfg.LoaderCgrConfig.CachesConns[0].Transport, utils.Meta),
			nil, false); err != nil {
			log.Fatalf("Could not connect to CacheS: %s", err.Error())
			return
		}
	} else {
		log.Print("WARNING: automatic cache reloading is disabled!")
	}

	// FixMe: remove users reloading as soon as not longer supported
	if *usersAddress != "" { // Init connection to rater so we can reload it's data
		if len(ldrCfg.LoaderCgrConfig.CachesConns) != 0 &&
			*usersAddress == ldrCfg.LoaderCgrConfig.CachesConns[0].Address {
			userS = cacheS
		} else {
			if userS, err = rpcclient.NewRpcClient("tcp", *usersAddress, 3, 3,
				time.Duration(1*time.Second), time.Duration(5*time.Minute),
				strings.TrimPrefix(*rpcEncoding, utils.Meta), nil, false); err != nil {
				log.Fatalf("Could not connect to UserS API: %s", err.Error())
				return
			}
		}
	} else {
		log.Print("WARNING: Users automatic data reload is disabled!")
	}

	if !*remove {
		// write maps to database
		if err := tpReader.WriteToDatabase(*flush, *verbose, *disableReverse); err != nil {
			log.Fatal("Could not write to database: ", err)
		}
		var dstIds, revDstIDs, rplIds, rpfIds, actIds, aapIDs, shgIds, alsIds, lcrIds, dcsIds, rspIDs, resIDs, aatIDs, ralsIDs []string
		if cacheS != nil {
			dstIds, _ = tpReader.GetLoadedIds(utils.DESTINATION_PREFIX)
			revDstIDs, _ = tpReader.GetLoadedIds(utils.REVERSE_DESTINATION_PREFIX)
			rplIds, _ = tpReader.GetLoadedIds(utils.RATING_PLAN_PREFIX)
			rpfIds, _ = tpReader.GetLoadedIds(utils.RATING_PROFILE_PREFIX)
			actIds, _ = tpReader.GetLoadedIds(utils.ACTION_PREFIX)
			aapIDs, _ = tpReader.GetLoadedIds(utils.AccountActionPlansPrefix)
			shgIds, _ = tpReader.GetLoadedIds(utils.SHARED_GROUP_PREFIX)
			alsIds, _ = tpReader.GetLoadedIds(utils.ALIASES_PREFIX)
			lcrIds, _ = tpReader.GetLoadedIds(utils.LCR_PREFIX)
			dcsIds, _ = tpReader.GetLoadedIds(utils.DERIVEDCHARGERS_PREFIX)
			rspIDs, _ = tpReader.GetLoadedIds(utils.ResourceProfilesPrefix)
			resIDs, _ = tpReader.GetLoadedIds(utils.ResourcesPrefix)
			aatIDs, _ = tpReader.GetLoadedIds(utils.ACTION_TRIGGER_PREFIX)
			ralsIDs, _ = tpReader.GetLoadedIds(utils.REVERSE_ALIASES_PREFIX)
		}
		aps, _ := tpReader.GetLoadedIds(utils.ACTION_PLAN_PREFIX)
		// for users reloading
		var userIds []string
		if userS != nil {
			userIds, _ = tpReader.GetLoadedIds(utils.USERS_PREFIX)
		}
		// release the reader with it's structures
		tpReader.Init()

		// Reload scheduler and cache
		if cacheS != nil {
			var reply string
			// Reload cache first since actions could be calling info from within
			if *verbose {
				log.Print("Reloading cache")
			}
			if err = cacheS.Call("ApierV1.ReloadCache",
				utils.AttrReloadCache{ArgsCache: utils.ArgsCache{
					DestinationIDs:        &dstIds,
					ReverseDestinationIDs: &revDstIDs,
					RatingPlanIDs:         &rplIds,
					RatingProfileIDs:      &rpfIds,
					ActionIDs:             &actIds,
					ActionPlanIDs:         &aps,
					AccountActionPlanIDs:  &aapIDs,
					ActionTriggerIDs:      &aatIDs,
					SharedGroupIDs:        &shgIds,
					LCRids:                &lcrIds,
					DerivedChargerIDs:     &dcsIds,
					AliasIDs:              &alsIds,
					ReverseAliasIDs:       &ralsIDs,
					ResourceProfileIDs:    &rspIDs,
					ResourceIDs:           &resIDs},
					FlushAll: *flush,
				}, &reply); err != nil {
				log.Printf("WARNING: Got error on cache reload: %s\n", err.Error())
			}

			if len(aps) != 0 {
				if *verbose {
					log.Print("Reloading scheduler")
				}
				if err = cacheS.Call("ApierV1.ReloadScheduler", "", &reply); err != nil {
					log.Printf("WARNING: Got error on scheduler reload: %s\n", err.Error())
				}
			}

			if userS != nil && len(userIds) > 0 {
				if *verbose {
					log.Print("Reloading Users data")
				}
				var reply string
				if err := userS.Call("UsersV1.ReloadUsers", "", &reply); err != nil {
					log.Printf("WARNING: Failed reloading users data, error: %s\n", err.Error())
				}
			}

		}

	} else {
		if err := tpReader.RemoveFromDatabase(*verbose, *disableReverse); err != nil {
			log.Fatal("Could not delete from database: ", err)
		}
	}
}
