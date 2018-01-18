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
	datadb_type = flag.String("datadb_type", config.CgrConfig().DataDbType, "The type of the DataDb database <*redis|*mongo>")
	datadb_host = flag.String("datadb_host", utils.MetaDynamic, "The DataDb host to connect to.")
	datadb_port = flag.String("datadb_port", utils.MetaDynamic, "The DataDb port to bind to.")
	datadb_name = flag.String("datadb_name", utils.MetaDynamic, "The name/number of the DataDb to connect to.")
	datadb_user = flag.String("datadb_user", utils.MetaDynamic, "The DataDb user to sign in as.")
	datadb_pass = flag.String("datadb_passwd", utils.MetaDynamic, "The DataDb user's password.")

	stor_db_type = flag.String("stordb_type", config.CgrConfig().StorDBType, "The type of the storDb database <*mysql|*postgres|*mongo>")
	stor_db_host = flag.String("stordb_host", utils.MetaDynamic, "The storDb host to connect to.")
	stor_db_port = flag.String("stordb_port", utils.MetaDynamic, "The storDb port to bind to.")
	stor_db_name = flag.String("stordb_name", utils.MetaDynamic, "The name/number of the storDb to connect to.")
	stor_db_user = flag.String("stordb_user", utils.MetaDynamic, "The storDb user to sign in as.")
	stor_db_pass = flag.String("stordb_passwd", utils.MetaDynamic, "The storDb user's password.")

	dbdata_encoding = flag.String("dbdata_encoding", config.CgrConfig().DBDataEncoding, "The encoding used to store object data in strings")

	flush           = flag.Bool("flushdb", false, "Flush the database before importing")
	tpid            = flag.String("tpid", "", "The tariff plan id from the database")
	dataPath        = flag.String("path", "./", "The path to folder containing the data files")
	version         = flag.Bool("version", false, "Prints the application version.")
	verbose         = flag.Bool("verbose", false, "Enable detailed verbose logging output")
	dryRun          = flag.Bool("dry_run", false, "When true will not save loaded data to dataDb but just parse it for consistency and errors.")
	validate        = flag.Bool("validate", false, "When true will run various check on the loaded data to check for structural errors")
	stats           = flag.Bool("stats", false, "Generates statsistics about given data.")
	fromStorDb      = flag.Bool("from_stordb", false, "Load the tariff plan from storDb to dataDb")
	toStorDb        = flag.Bool("to_stordb", false, "Import the tariff plan from files to storDb")
	rpcEncoding     = flag.String("rpc_encoding", "json", "RPC encoding used <gob|json>")
	ralsAddress     = flag.String("rals", config.CgrConfig().RPCJSONListen, "Rater service to contact for cache reloads, empty to disable automatic cache reloads")
	cdrstatsAddress = flag.String("cdrstats", config.CgrConfig().RPCJSONListen, "CDRStats service to contact for data reloads, empty to disable automatic data reloads")
	usersAddress    = flag.String("users", config.CgrConfig().RPCJSONListen, "Users service to contact for data reloads, empty to disable automatic data reloads")
	runId           = flag.String("runid", "", "Uniquely identify an import/load, postpended to some automatic fields")
	loadHistorySize = flag.Int("load_history_size", config.CgrConfig().LoadHistorySize, "Limit the number of records in the load history")
	timezone        = flag.String("timezone", config.CgrConfig().DefaultTimezone, `Timezone for timestamps where not specified <""|UTC|Local|$IANA_TZ_DB>`)
	disable_reverse = flag.Bool("disable_reverse_mappings", false, "Will disable reverse mappings rebuilding")
	remove          = flag.Bool("remove", false, "Will remove any data from db that matches data files")
)

func main() {
	flag.Parse()
	if *version {
		fmt.Println(utils.GetCGRVersion())
		return
	}
	var errDataDB, errStorDb, err error
	var dm *engine.DataManager
	var storDb engine.LoadStorage
	var rater, cdrstats, users rpcclient.RpcClientConnection
	var loader engine.LoadReader

	*datadb_type = strings.TrimPrefix(*datadb_type, "*")
	*datadb_host = config.DBDefaults.DBHost(*datadb_type, *datadb_host)
	*datadb_port = config.DBDefaults.DBPort(*datadb_type, *datadb_port)
	*datadb_name = config.DBDefaults.DBName(*datadb_type, *datadb_name)
	*datadb_user = config.DBDefaults.DBUser(*datadb_type, *datadb_user)
	*datadb_pass = config.DBDefaults.DBPass(*datadb_type, *datadb_pass)

	*stor_db_type = strings.TrimPrefix(*stor_db_type, "*")
	*stor_db_host = config.DBDefaults.DBHost(*stor_db_type, *stor_db_host)
	*stor_db_port = config.DBDefaults.DBPort(*stor_db_type, *stor_db_port)
	*stor_db_name = config.DBDefaults.DBName(*stor_db_type, *stor_db_name)
	*stor_db_user = config.DBDefaults.DBUser(*stor_db_type, *stor_db_user)
	*stor_db_pass = config.DBDefaults.DBPass(*stor_db_type, *stor_db_pass)

	if !*toStorDb {
		dm, errDataDB = engine.ConfigureDataStorage(*datadb_type, *datadb_host, *datadb_port, *datadb_name,
			*datadb_user, *datadb_pass, *dbdata_encoding, config.CgrConfig().CacheCfg(), *loadHistorySize)
	}
	if *fromStorDb || *toStorDb {
		storDb, errStorDb = engine.ConfigureLoadStorage(*stor_db_type, *stor_db_host, *stor_db_port, *stor_db_name, *stor_db_user, *stor_db_pass, *dbdata_encoding,
			config.CgrConfig().StorDBMaxOpenConns, config.CgrConfig().StorDBMaxIdleConns, config.CgrConfig().StorDBConnMaxLifetime, config.CgrConfig().StorDBCDRSIndexes)
	}
	// Stop on db errors
	for _, err = range []error{errDataDB, errDataDB, errStorDb} {
		if err != nil {
			log.Fatalf("Could not open database connection: %v", err)
		}
	}
	// Defer databases opened to be closed when we are done
	for _, db := range []engine.Storage{dm.DataDB(), storDb} {
		if db != nil {
			defer db.Close()
		}
	}
	// Init necessary db connections, only if not already
	if !*dryRun { // make sure we do not need db connections on dry run, also not importing into any stordb
		if *toStorDb { // Import files from a directory into storDb
			if *tpid == "" {
				log.Fatal("TPid required, please define it via *-tpid* command argument.")
			}
			csvImporter := engine.TPCSVImporter{
				TPid:     *tpid,
				StorDb:   storDb,
				DirPath:  *dataPath,
				Sep:      ',',
				Verbose:  *verbose,
				ImportId: *runId,
			}
			if errImport := csvImporter.Run(); errImport != nil {
				log.Fatal(errImport)
			}
			return
		}
	}
	if *fromStorDb { // Load Tariff Plan from storDb into dataDb
		loader = storDb
	} else { // Default load from csv files to dataDb
		/*for fn, v := range engine.FileValidators {
			err := engine.ValidateCSVData(path.Join(*dataPath, fn), v.Rule)
			if err != nil {
				log.Fatal(err, "\n\t", v.Message)
			}
		}*/
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

	tpReader := engine.NewTpReader(dm.DataDB(), loader, *tpid, *timezone)
	err = tpReader.LoadAll()
	if err != nil {
		log.Fatal(err)
	}
	if *stats {
		tpReader.ShowStatistics()
	}
	if *validate {
		if !tpReader.IsValid() {
			return
		}
	}
	if *dryRun { // We were just asked to parse the data, not saving it
		return
	}
	if *ralsAddress != "" { // Init connection to rater so we can reload it's data
		if rater, err = rpcclient.NewRpcClient("tcp", *ralsAddress, 3, 3,
			time.Duration(1*time.Second), time.Duration(5*time.Minute), *rpcEncoding, nil, false); err != nil {
			log.Fatalf("Could not connect to RALs: %s", err.Error())
			return
		}
	} else {
		log.Print("WARNING: Rates automatic cache reloading is disabled!")
	}
	if *cdrstatsAddress != "" { // Init connection to rater so we can reload it's data
		if *cdrstatsAddress == *ralsAddress {
			cdrstats = rater
		} else {
			if cdrstats, err = rpcclient.NewRpcClient("tcp", *cdrstatsAddress, 3, 3,
				time.Duration(1*time.Second), time.Duration(5*time.Minute), *rpcEncoding, nil, false); err != nil {
				log.Fatalf("Could not connect to CDRStatS API: %s", err.Error())
				return
			}
		}
	} else {
		log.Print("WARNING: CDRStats automatic data reload is disabled!")
	}
	if *usersAddress != "" { // Init connection to rater so we can reload it's data
		if *usersAddress == *ralsAddress {
			users = rater
		} else {
			if users, err = rpcclient.NewRpcClient("tcp", *usersAddress, 3, 3,
				time.Duration(1*time.Second), time.Duration(5*time.Minute), *rpcEncoding, nil, false); err != nil {
				log.Fatalf("Could not connect to UserS API: %s", err.Error())
				return
			}
		}
	} else {
		log.Print("WARNING: Users automatic data reload is disabled!")
	}
	if !*remove {
		// write maps to database
		if err := tpReader.WriteToDatabase(*flush, *verbose, *disable_reverse); err != nil {
			log.Fatal("Could not write to database: ", err)
		}
		var dstIds, revDstIDs, rplIds, rpfIds, actIds, aapIDs, shgIds, alsIds, lcrIds, dcsIds, rspIDs, resIDs, aatIDs, ralsIDs []string
		if rater != nil {
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
		var statsQueueIds []string
		if cdrstats != nil {
			statsQueueIds, _ = tpReader.GetLoadedIds(utils.CDR_STATS_PREFIX)
		}
		var userIds []string
		if users != nil {
			userIds, _ = tpReader.GetLoadedIds(utils.USERS_PREFIX)
		}
		// release the reader with it's structures
		tpReader.Init()

		// Reload scheduler and cache
		if rater != nil {
			reply := ""

			// Reload cache first since actions could be calling info from within
			if *verbose {
				log.Print("Reloading cache")
			}
			if err = rater.Call("ApierV1.ReloadCache", utils.AttrReloadCache{ArgsCache: utils.ArgsCache{
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
				if err = rater.Call("ApierV1.ReloadScheduler", "", &reply); err != nil {
					log.Printf("WARNING: Got error on scheduler reload: %s\n", err.Error())
				}
			}

		}
		if cdrstats != nil {
			if *flush {
				statsQueueIds = []string{} // Force reload all
			}
			if len(statsQueueIds) != 0 {
				if *verbose {
					log.Print("Reloading CDRStats data")
				}
				var reply string
				if err := cdrstats.Call("CDRStatsV1.ReloadQueues", utils.AttrCDRStatsReloadQueues{StatsQueueIds: statsQueueIds}, &reply); err != nil {
					log.Printf("WARNING: Failed reloading stat queues, error: %s\n", err.Error())
				}
			}
		}

		if users != nil {
			if len(userIds) > 0 {
				if *verbose {
					log.Print("Reloading Users data")
				}
				var reply string
				if err := cdrstats.Call("UsersV1.ReloadUsers", "", &reply); err != nil {
					log.Printf("WARNING: Failed reloading users data, error: %s\n", err.Error())
				}

			}
		}
	} else {
		if err := tpReader.RemoveFromDatabase(*verbose, *disable_reverse); err != nil {
			log.Fatal("Could not delete from database: ", err)
		}
	}
}
