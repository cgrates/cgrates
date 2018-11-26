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

	dataDBType = flag.String("datadb_type", dfltCfg.DataDbCfg().DataDbType,
		"The type of the DataDB database <*redis|*mongo>")
	dataDBHost = flag.String("datadb_host", dfltCfg.DataDbCfg().DataDbHost,
		"The DataDb host to connect to.")
	dataDBPort = flag.String("datadb_port", dfltCfg.DataDbCfg().DataDbPort,
		"The DataDb port to bind to.")
	dataDBName = flag.String("datadb_name", dfltCfg.DataDbCfg().DataDbName,
		"The name/number of the DataDb to connect to.")
	dataDBUser = flag.String("datadb_user", dfltCfg.DataDbCfg().DataDbUser,
		"The DataDb user to sign in as.")
	dataDBPasswd = flag.String("datadb_passwd", dfltCfg.DataDbCfg().DataDbPass,
		"The DataDb user's password.")
	dbDataEncoding = flag.String("dbdata_encoding", dfltCfg.GeneralCfg().DBDataEncoding,
		"The encoding used to store object data in strings")
	dbRedisSentinel = flag.String("redis_sentinel", dfltCfg.DataDbCfg().DataDbSentinelName,
		"The name of redis sentinel")

	storDBType = flag.String("stordb_type", dfltCfg.StorDbCfg().StorDBType,
		"The type of the storDb database <*mysql|*postgres|*mongo>")
	storDBHost = flag.String("stordb_host", dfltCfg.StorDbCfg().StorDBHost,
		"The storDb host to connect to.")
	storDBPort = flag.String("stordb_port", dfltCfg.StorDbCfg().StorDBPort,
		"The storDb port to bind to.")
	storDBName = flag.String("stordb_name", dfltCfg.StorDbCfg().StorDBName,
		"The name/number of the storDb to connect to.")
	storDBUser = flag.String("stordb_user", dfltCfg.StorDbCfg().StorDBUser,
		"The storDb user to sign in as.")
	storDBPasswd = flag.String("stordb_passwd", dfltCfg.StorDbCfg().StorDBPass,
		"The storDb user's password.")

	flush = flag.Bool("flushdb", false,
		"Flush the database before importing")
	tpid = flag.String("tpid", dfltCfg.LoaderCgrCfg().TpID,
		"The tariff plan ID from the database")
	dataPath = flag.String("path", dfltCfg.LoaderCgrCfg().DataPath,
		"The path to folder containing the data files")
	version = flag.Bool("version", false,
		"Prints the application version.")
	verbose = flag.Bool("verbose", false,
		"Enable detailed verbose logging output")
	dryRun = flag.Bool("dry_run", false,
		"When true will not save loaded data to dataDb but just parse it for consistency and errors.")
	fieldSep = flag.String("field_sep", ",",
		`Separator for csv file (by default "," is used)`)

	fromStorDB    = flag.Bool("from_stordb", false, "Load the tariff plan from storDb to dataDb")
	toStorDB      = flag.Bool("to_stordb", false, "Import the tariff plan from files to storDb")
	rpcEncoding   = flag.String("rpc_encoding", utils.MetaJSONrpc, "RPC encoding used <*gob|*json>")
	cacheSAddress = flag.String("caches_address", dfltCfg.LoaderCgrCfg().CachesConns[0].Address,
		"CacheS component to contact for cache reloads, empty to disable automatic cache reloads")
	schedulerAddress = flag.String("scheduler_address", dfltCfg.LoaderCgrCfg().SchedulerConns[0].Address, "")

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

	// Data for DataDB
	if *dataDBType != dfltCfg.DataDbCfg().DataDbType {
		ldrCfg.DataDbCfg().DataDbType = strings.TrimPrefix(*dataDBType, "*")
	}

	if *dataDBHost != dfltCfg.DataDbCfg().DataDbHost {
		ldrCfg.DataDbCfg().DataDbHost = *dataDBHost
	}

	if *dataDBPort != dfltCfg.DataDbCfg().DataDbPort {
		ldrCfg.DataDbCfg().DataDbPort = *dataDBPort
	}

	if *dataDBName != dfltCfg.DataDbCfg().DataDbName {
		ldrCfg.DataDbCfg().DataDbName = *dataDBName
	}

	if *dataDBUser != dfltCfg.DataDbCfg().DataDbUser {
		ldrCfg.DataDbCfg().DataDbUser = *dataDBUser
	}

	if *dataDBPasswd != dfltCfg.DataDbCfg().DataDbPass {
		ldrCfg.DataDbCfg().DataDbPass = *dataDBPasswd
	}

	if *dbRedisSentinel != dfltCfg.DataDbCfg().DataDbSentinelName {
		ldrCfg.DataDbCfg().DataDbSentinelName = *dbRedisSentinel
	}

	if *dbDataEncoding != dfltCfg.GeneralCfg().DBDataEncoding {
		ldrCfg.GeneralCfg().DBDataEncoding = *dbDataEncoding
	}

	// Data for StorDB
	if *storDBType != dfltCfg.StorDbCfg().StorDBType {
		ldrCfg.StorDbCfg().StorDBType = strings.TrimPrefix(*storDBType, "*")
	}

	if *storDBHost != dfltCfg.StorDbCfg().StorDBHost {
		ldrCfg.StorDbCfg().StorDBHost = *storDBHost
	}

	if *storDBPort != dfltCfg.StorDbCfg().StorDBPort {
		ldrCfg.StorDbCfg().StorDBPort = *storDBPort
	}

	if *storDBName != dfltCfg.StorDbCfg().StorDBName {
		ldrCfg.StorDbCfg().StorDBName = *storDBName
	}

	if *storDBUser != dfltCfg.StorDbCfg().StorDBUser {
		ldrCfg.StorDbCfg().StorDBUser = *storDBUser
	}

	if *storDBPasswd != dfltCfg.StorDbCfg().StorDBPass {
		ldrCfg.StorDbCfg().StorDBPass = *storDBPasswd
	}

	if *tpid != dfltCfg.LoaderCgrCfg().DataPath {
		ldrCfg.LoaderCgrCfg().TpID = *tpid
	}

	if *dataPath != dfltCfg.LoaderCgrCfg().DataPath {
		ldrCfg.LoaderCgrCfg().DataPath = *dataPath
	}

	if rune((*fieldSep)[0]) != dfltCfg.LoaderCgrCfg().FieldSeparator {
		ldrCfg.LoaderCgrCfg().FieldSeparator = rune((*fieldSep)[0])
	}

	if *cacheSAddress != dfltCfg.LoaderCgrCfg().CachesConns[0].Address {
		ldrCfg.LoaderCgrCfg().CachesConns = make([]*config.HaPoolConfig, 0)
		if *cacheSAddress != "" {
			ldrCfg.LoaderCgrCfg().CachesConns = append(ldrCfg.LoaderCgrCfg().CachesConns,
				&config.HaPoolConfig{
					Address:   *cacheSAddress,
					Transport: *rpcEncoding,
				})
		}
	}

	if *schedulerAddress != dfltCfg.LoaderCgrCfg().SchedulerConns[0].Address {
		ldrCfg.LoaderCgrCfg().SchedulerConns = make([]*config.HaPoolConfig, 0)
		if *schedulerAddress != "" {
			ldrCfg.LoaderCgrCfg().SchedulerConns = append(ldrCfg.LoaderCgrCfg().SchedulerConns,
				&config.HaPoolConfig{Address: *schedulerAddress})
		}
	}

	if *rpcEncoding != dfltCfg.LoaderCgrCfg().CachesConns[0].Transport &&
		len(ldrCfg.LoaderCgrCfg().CachesConns) != 0 {
		ldrCfg.LoaderCgrCfg().CachesConns[0].Transport = *rpcEncoding
	}

	if *importID == "" {
		*importID = utils.UUIDSha1Prefix()
	}

	if *timezone != dfltCfg.GeneralCfg().DefaultTimezone {
		ldrCfg.GeneralCfg().DefaultTimezone = *timezone
	}

	if *disableReverse != dfltCfg.LoaderCgrCfg().DisableReverse {
		ldrCfg.LoaderCgrCfg().DisableReverse = *disableReverse
	}

	if !*toStorDB {
		if dm, err = engine.ConfigureDataStorage(ldrCfg.DataDbCfg().DataDbType,
			ldrCfg.DataDbCfg().DataDbHost, ldrCfg.DataDbCfg().DataDbPort,
			ldrCfg.DataDbCfg().DataDbName, ldrCfg.DataDbCfg().DataDbUser,
			ldrCfg.DataDbCfg().DataDbPass, ldrCfg.GeneralCfg().DBDataEncoding,
			config.CgrConfig().CacheCfg(), ldrCfg.DataDbCfg().DataDbSentinelName); err != nil {
			log.Fatalf("Coud not open dataDB connection: %s", err.Error())
		}
		defer dm.DataDB().Close()
	}

	if *fromStorDB || *toStorDB {
		if storDb, err = engine.ConfigureLoadStorage(ldrCfg.StorDbCfg().StorDBType,
			ldrCfg.StorDbCfg().StorDBHost, ldrCfg.StorDbCfg().StorDBPort,
			ldrCfg.StorDbCfg().StorDBName, ldrCfg.StorDbCfg().StorDBUser,
			ldrCfg.StorDbCfg().StorDBPass, ldrCfg.GeneralCfg().DBDataEncoding,
			config.CgrConfig().StorDbCfg().StorDBMaxOpenConns,
			config.CgrConfig().StorDbCfg().StorDBMaxIdleConns,
			config.CgrConfig().StorDbCfg().StorDBConnMaxLifetime,
			config.CgrConfig().StorDbCfg().StorDBCDRSIndexes); err != nil {
			log.Fatalf("Coud not open storDB connection: %s", err.Error())
		}
		defer storDb.Close()
	}

	if !*dryRun {
		//tpid_remove
		if *toStorDB { // Import files from a directory into storDb
			if ldrCfg.LoaderCgrCfg().TpID == "" {
				log.Fatal("TPid required.")
			}
			if *flushStorDB {
				if err = storDb.RemTpData("", ldrCfg.LoaderCgrCfg().TpID, map[string]string{}); err != nil {
					log.Fatal(err)
				}
			}
			csvImporter := engine.TPCSVImporter{
				TPid:     ldrCfg.LoaderCgrCfg().TpID,
				StorDb:   storDb,
				DirPath:  *dataPath,
				Sep:      ldrCfg.LoaderCgrCfg().FieldSeparator,
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
		loader = engine.NewFileCSVStorage(ldrCfg.LoaderCgrCfg().FieldSeparator,
			path.Join(*dataPath, utils.DESTINATIONS_CSV),
			path.Join(*dataPath, utils.TIMINGS_CSV),
			path.Join(*dataPath, utils.RATES_CSV),
			path.Join(*dataPath, utils.DESTINATION_RATES_CSV),
			path.Join(*dataPath, utils.RATING_PLANS_CSV),
			path.Join(*dataPath, utils.RATING_PROFILES_CSV),
			path.Join(*dataPath, utils.SHARED_GROUPS_CSV),
			path.Join(*dataPath, utils.ACTIONS_CSV),
			path.Join(*dataPath, utils.ACTION_PLANS_CSV),
			path.Join(*dataPath, utils.ACTION_TRIGGERS_CSV),
			path.Join(*dataPath, utils.ACCOUNT_ACTIONS_CSV),
			path.Join(*dataPath, utils.DERIVED_CHARGERS_CSV),
			path.Join(*dataPath, utils.USERS_CSV),
			path.Join(*dataPath, utils.ALIASES_CSV),
			path.Join(*dataPath, utils.ResourcesCsv),
			path.Join(*dataPath, utils.StatsCsv),
			path.Join(*dataPath, utils.ThresholdsCsv),
			path.Join(*dataPath, utils.FiltersCsv),
			path.Join(*dataPath, utils.SuppliersCsv),
			path.Join(*dataPath, utils.AttributesCsv),
			path.Join(*dataPath, utils.ChargersCsv),
		)
	}

	tpReader := engine.NewTpReader(dm.DataDB(), loader,
		ldrCfg.LoaderCgrCfg().TpID, ldrCfg.GeneralCfg().DefaultTimezone)

	if err = tpReader.LoadAll(); err != nil {
		log.Fatal(err)
	}

	if *dryRun { // We were just asked to parse the data, not saving it
		return
	}
	if len(ldrCfg.LoaderCgrCfg().CachesConns) != 0 { // Init connection to CacheS so we can reload it's data
		if cacheS, err = rpcclient.NewRpcClient("tcp",
			ldrCfg.LoaderCgrCfg().CachesConns[0].Address,
			ldrCfg.LoaderCgrCfg().CachesConns[0].Tls, ldrCfg.TlsCfg().ClientKey,
			ldrCfg.TlsCfg().ClientCerificate, ldrCfg.TlsCfg().CaCertificate, 3, 3,
			time.Duration(1*time.Second), time.Duration(5*time.Minute),
			strings.TrimPrefix(ldrCfg.LoaderCgrCfg().CachesConns[0].Transport, utils.Meta),
			nil, false); err != nil {
			log.Fatalf("Could not connect to CacheS: %s", err.Error())
			return
		}
	} else {
		log.Print("WARNING: automatic cache reloading is disabled!")
	}

	// FixMe: remove users reloading as soon as not longer supported
	if *usersAddress != "" { // Init connection to rater so we can reload it's data
		if len(ldrCfg.LoaderCgrCfg().CachesConns) != 0 &&
			*usersAddress == ldrCfg.LoaderCgrCfg().CachesConns[0].Address {
			userS = cacheS
		} else {
			if userS, err = rpcclient.NewRpcClient("tcp", *usersAddress,
				ldrCfg.LoaderCgrCfg().CachesConns[0].Tls,
				ldrCfg.TlsCfg().ClientKey, ldrCfg.TlsCfg().ClientCerificate,
				ldrCfg.TlsCfg().CaCertificate, 3, 3,
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
		var dstIds, revDstIDs, rplIds, rpfIds, actIds, aapIDs, shgIds, alsIds, dcsIds, rspIDs, resIDs, aatIDs, ralsIDs, stqIDs, stqpIDs, trsIDs, trspfIDs, flrIDs, spfIDs, apfIDs, chargerIDs []string
		if cacheS != nil {
			dstIds, _ = tpReader.GetLoadedIds(utils.DESTINATION_PREFIX)
			revDstIDs, _ = tpReader.GetLoadedIds(utils.REVERSE_DESTINATION_PREFIX)
			rplIds, _ = tpReader.GetLoadedIds(utils.RATING_PLAN_PREFIX)
			rpfIds, _ = tpReader.GetLoadedIds(utils.RATING_PROFILE_PREFIX)
			actIds, _ = tpReader.GetLoadedIds(utils.ACTION_PREFIX)
			aapIDs, _ = tpReader.GetLoadedIds(utils.AccountActionPlansPrefix)
			shgIds, _ = tpReader.GetLoadedIds(utils.SHARED_GROUP_PREFIX)
			alsIds, _ = tpReader.GetLoadedIds(utils.ALIASES_PREFIX)
			dcsIds, _ = tpReader.GetLoadedIds(utils.DERIVEDCHARGERS_PREFIX)
			rspIDs, _ = tpReader.GetLoadedIds(utils.ResourceProfilesPrefix)
			resIDs, _ = tpReader.GetLoadedIds(utils.ResourcesPrefix)
			aatIDs, _ = tpReader.GetLoadedIds(utils.ACTION_TRIGGER_PREFIX)
			ralsIDs, _ = tpReader.GetLoadedIds(utils.REVERSE_ALIASES_PREFIX)
			stqIDs, _ = tpReader.GetLoadedIds(utils.StatQueuePrefix)
			stqpIDs, _ = tpReader.GetLoadedIds(utils.StatQueueProfilePrefix)
			trsIDs, _ = tpReader.GetLoadedIds(utils.ThresholdPrefix)
			trspfIDs, _ = tpReader.GetLoadedIds(utils.ThresholdProfilePrefix)
			flrIDs, _ = tpReader.GetLoadedIds(utils.FilterPrefix)
			spfIDs, _ = tpReader.GetLoadedIds(utils.SupplierProfilePrefix)
			apfIDs, _ = tpReader.GetLoadedIds(utils.AttributeProfilePrefix)
			chargerIDs, _ = tpReader.GetLoadedIds(utils.ChargerProfilePrefix)
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
			if err = cacheS.Call(utils.ApierV1ReloadCache,
				utils.AttrReloadCache{ArgsCache: utils.ArgsCache{
					DestinationIDs:        &dstIds,
					ReverseDestinationIDs: &revDstIDs,
					RatingPlanIDs:         &rplIds,
					RatingProfileIDs:      &rpfIds,
					ActionIDs:             &actIds,
					ActionPlanIDs:         &aps,
					AccountActionPlanIDs:  &aapIDs,
					SharedGroupIDs:        &shgIds,
					AliasIDs:              &alsIds,
					DerivedChargerIDs:     &dcsIds,
					ResourceProfileIDs:    &rspIDs,
					ResourceIDs:           &resIDs,
					ActionTriggerIDs:      &aatIDs,
					ReverseAliasIDs:       &ralsIDs,
					StatsQueueIDs:         &stqIDs,
					StatsQueueProfileIDs:  &stqpIDs,
					ThresholdIDs:          &trsIDs,
					ThresholdProfileIDs:   &trspfIDs,
					FilterIDs:             &flrIDs,
					SupplierProfileIDs:    &spfIDs,
					AttributeProfileIDs:   &apfIDs,
					ChargerProfileIDs:     &chargerIDs},
					FlushAll: *flush,
				}, &reply); err != nil {
				log.Printf("WARNING: Got error on cache reload: %s\n", err.Error())
			}
			if *verbose {
				log.Print("Clearing cached indexes")
			}
			var cacheIDs []string
			if len(apfIDs) != 0 {
				cacheIDs = append(cacheIDs, utils.CacheAttributeFilterIndexes)
			}
			if len(spfIDs) != 0 {
				cacheIDs = append(cacheIDs, utils.CacheSupplierFilterIndexes)
			}
			if len(trspfIDs) != 0 {
				cacheIDs = append(cacheIDs, utils.CacheThresholdFilterIndexes)
			}
			if len(stqpIDs) != 0 {
				cacheIDs = append(cacheIDs, utils.CacheStatFilterIndexes)
			}
			if len(rspIDs) != 0 {
				cacheIDs = append(cacheIDs, utils.CacheResourceFilterIndexes)
			}
			if len(chargerIDs) != 0 {
				cacheIDs = append(cacheIDs, utils.CacheChargerFilterIndexes)
			}
			if err = cacheS.Call(utils.CacheSv1Clear, cacheIDs, &reply); err != nil {
				log.Printf("WARNING: Got error on cache clear: %s\n", err.Error())
			}

			if len(aps) != 0 {
				if *verbose {
					log.Print("Reloading scheduler")
				}
				if err = cacheS.Call(utils.ApierV1ReloadScheduler, "", &reply); err != nil {
					log.Printf("WARNING: Got error on scheduler reload: %s\n", err.Error())
				}
			}

			if userS != nil && len(userIds) > 0 {
				if *verbose {
					log.Print("Reloading Users data")
				}
				var reply string
				if err := userS.Call(utils.UsersV1ReloadUsers, "", &reply); err != nil {
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
