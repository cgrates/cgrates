/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2012-2015 ITsysCOM

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
	"net/rpc"
	"path"
	"strconv"
	"strings"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

var (
	//separator = flag.String("separator", ",", "Default field separator")
	cgrConfig, _ = config.NewDefaultCGRConfig()
	migrateRC8   = flag.String("migrate_rc8", "", "Migrate Accounts, Actions, ActionTriggers, DerivedChargers, ActionPlans and SharedGroups to RC8 structures, possible values: *all,acc,atr,act,dcs,apl,shg")
	tpdb_type    = flag.String("tpdb_type", cgrConfig.TpDbType, "The type of the TariffPlan database <redis>")
	tpdb_host    = flag.String("tpdb_host", cgrConfig.TpDbHost, "The TariffPlan host to connect to.")
	tpdb_port    = flag.String("tpdb_port", cgrConfig.TpDbPort, "The TariffPlan port to bind to.")
	tpdb_name    = flag.String("tpdb_name", cgrConfig.TpDbName, "The name/number of the TariffPlan to connect to.")
	tpdb_user    = flag.String("tpdb_user", cgrConfig.TpDbUser, "The TariffPlan user to sign in as.")
	tpdb_pass    = flag.String("tpdb_passwd", cgrConfig.TpDbPass, "The TariffPlan user's password.")

	datadb_type = flag.String("datadb_type", cgrConfig.DataDbType, "The type of the DataDb database <redis>")
	datadb_host = flag.String("datadb_host", cgrConfig.DataDbHost, "The DataDb host to connect to.")
	datadb_port = flag.String("datadb_port", cgrConfig.DataDbPort, "The DataDb port to bind to.")
	datadb_name = flag.String("datadb_name", cgrConfig.DataDbName, "The name/number of the DataDb to connect to.")
	datadb_user = flag.String("datadb_user", cgrConfig.DataDbUser, "The DataDb user to sign in as.")
	datadb_pass = flag.String("datadb_passwd", cgrConfig.DataDbPass, "The DataDb user's password.")

	stor_db_type = flag.String("stordb_type", cgrConfig.StorDBType, "The type of the storDb database <mysql>")
	stor_db_host = flag.String("stordb_host", cgrConfig.StorDBHost, "The storDb host to connect to.")
	stor_db_port = flag.String("stordb_port", cgrConfig.StorDBPort, "The storDb port to bind to.")
	stor_db_name = flag.String("stordb_name", cgrConfig.StorDBName, "The name/number of the storDb to connect to.")
	stor_db_user = flag.String("stordb_user", cgrConfig.StorDBUser, "The storDb user to sign in as.")
	stor_db_pass = flag.String("stordb_passwd", cgrConfig.StorDBPass, "The storDb user's password.")

	dbdata_encoding = flag.String("dbdata_encoding", cgrConfig.DBDataEncoding, "The encoding used to store object data in strings")

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
	historyServer   = flag.String("history_server", cgrConfig.RPCGOBListen, "The history server address:port, empty to disable automaticautomatic  history archiving")
	raterAddress    = flag.String("rater_address", cgrConfig.RPCGOBListen, "Rater service to contact for cache reloads, empty to disable automatic cache reloads")
	cdrstatsAddress = flag.String("cdrstats_address", cgrConfig.RPCGOBListen, "CDRStats service to contact for data reloads, empty to disable automatic data reloads")
	usersAddress    = flag.String("users_address", cgrConfig.RPCGOBListen, "Users service to contact for data reloads, empty to disable automatic data reloads")
	runId           = flag.String("runid", "", "Uniquely identify an import/load, postpended to some automatic fields")
	loadHistorySize = flag.Int("load_history_size", cgrConfig.LoadHistorySize, "Limit the number of records in the load history")
	timezone        = flag.String("timezone", cgrConfig.DefaultTimezone, `Timezone for timestamps where not specified <""|UTC|Local|$IANA_TZ_DB>`)
)

func main() {
	flag.Parse()
	if *version {
		fmt.Println("CGRateS " + utils.VERSION)
		return
	}
	var errRatingDb, errAccDb, errStorDb, err error
	var ratingDb engine.RatingStorage
	var accountDb engine.AccountingStorage
	var storDb engine.LoadStorage
	var rater, cdrstats, users *rpc.Client
	var loader engine.LoadReader
	if *migrateRC8 != "" {
		if *datadb_type == "redis" && *tpdb_type == "redis" {
			var db_nb int
			db_nb, err = strconv.Atoi(*datadb_name)
			if err != nil {
				log.Print("Redis db name must be an integer!")
				return
			}
			host := *datadb_host
			if *datadb_port != "" {
				host += ":" + *datadb_port
			}
			migratorRC8acc, err := NewMigratorRC8(host, db_nb, *datadb_pass, *dbdata_encoding)
			if err != nil {
				log.Print(err.Error())
				return
			}
			if strings.Contains(*migrateRC8, "acc") || strings.Contains(*migrateRC8, "*all") {
				if err := migratorRC8acc.migrateAccounts(); err != nil {
					log.Print(err.Error())
				}
			}

			db_nb, err = strconv.Atoi(*tpdb_name)
			if err != nil {
				log.Print("Redis db name must be an integer!")
				return
			}
			host = *tpdb_host
			if *tpdb_port != "" {
				host += ":" + *tpdb_port
			}
			migratorRC8rat, err := NewMigratorRC8(host, db_nb, *tpdb_pass, *dbdata_encoding)
			if err != nil {
				log.Print(err.Error())
				return
			}
			if strings.Contains(*migrateRC8, "atr") || strings.Contains(*migrateRC8, "*all") {
				if err := migratorRC8rat.migrateActionTriggers(); err != nil {
					log.Print(err.Error())
				}
			}
			if strings.Contains(*migrateRC8, "act") || strings.Contains(*migrateRC8, "*all") {
				if err := migratorRC8rat.migrateActions(); err != nil {
					log.Print(err.Error())
				}
			}
			if strings.Contains(*migrateRC8, "dcs") || strings.Contains(*migrateRC8, "*all") {
				if err := migratorRC8rat.migrateDerivedChargers(); err != nil {
					log.Print(err.Error())
				}
			}
			if strings.Contains(*migrateRC8, "apl") || strings.Contains(*migrateRC8, "*all") {
				if err := migratorRC8rat.migrateActionPlans(); err != nil {
					log.Print(err.Error())
				}
			}
			if strings.Contains(*migrateRC8, "shg") || strings.Contains(*migrateRC8, "*all") {
				if err := migratorRC8rat.migrateSharedGroups(); err != nil {
					log.Print(err.Error())
				}
			}
			if strings.Contains(*migrateRC8, "int") {
				if err := migratorRC8acc.migrateAccountsInt(); err != nil {
					log.Print(err.Error())
				}
				if err := migratorRC8rat.migrateActionTriggersInt(); err != nil {
					log.Print(err.Error())
				}
				if err := migratorRC8rat.migrateActionsInt(); err != nil {
					log.Print(err.Error())
				}
			}
			if strings.Contains(*migrateRC8, "vf") {
				if err := migratorRC8rat.migrateActionsInt2(); err != nil {
					log.Print(err.Error())
				}
				if err := migratorRC8acc.writeVersion(); err != nil {
					log.Print(err.Error())
				}
			}
		} else if *datadb_type == "mongo" && *tpdb_type == "mongo" {
			mongoMigrator, err := NewMongoMigrator(*datadb_host, *datadb_port, *datadb_name, *datadb_user, *datadb_pass)
			if err != nil {
				log.Print(err.Error())
				return
			}
			if strings.Contains(*migrateRC8, "vf") {
				if err := mongoMigrator.migrateActions(); err != nil {
					log.Print(err.Error())
				}
				if err := mongoMigrator.writeVersion(); err != nil {
					log.Print(err.Error())
				}
			}
		}

		log.Print("Done!")
		return
	}
	// Init necessary db connections, only if not already
	if !*dryRun { // make sure we do not need db connections on dry run, also not importing into any stordb
		if *fromStorDb {
			ratingDb, errRatingDb = engine.ConfigureRatingStorage(*tpdb_type, *tpdb_host, *tpdb_port, *tpdb_name,
				*tpdb_user, *tpdb_pass, *dbdata_encoding)
			accountDb, errAccDb = engine.ConfigureAccountingStorage(*datadb_type, *datadb_host, *datadb_port, *datadb_name, *datadb_user, *datadb_pass, *dbdata_encoding)
			storDb, errStorDb = engine.ConfigureLoadStorage(*stor_db_type, *stor_db_host, *stor_db_port, *stor_db_name, *stor_db_user, *stor_db_pass, *dbdata_encoding,
				cgrConfig.StorDBMaxOpenConns, cgrConfig.StorDBMaxIdleConns, cgrConfig.StorDBCDRSIndexes)
		} else if *toStorDb { // Import from csv files to storDb
			storDb, errStorDb = engine.ConfigureLoadStorage(*stor_db_type, *stor_db_host, *stor_db_port, *stor_db_name, *stor_db_user, *stor_db_pass, *dbdata_encoding,
				cgrConfig.StorDBMaxOpenConns, cgrConfig.StorDBMaxIdleConns, cgrConfig.StorDBCDRSIndexes)
		} else { // Default load from csv files to dataDb
			ratingDb, errRatingDb = engine.ConfigureRatingStorage(*tpdb_type, *tpdb_host, *tpdb_port, *tpdb_name,
				*tpdb_user, *tpdb_pass, *dbdata_encoding)
			accountDb, errAccDb = engine.ConfigureAccountingStorage(*datadb_type, *datadb_host, *datadb_port, *datadb_name, *datadb_user, *datadb_pass, *dbdata_encoding)
		}
		// Defer databases opened to be closed when we are done
		for _, db := range []engine.Storage{ratingDb, accountDb, storDb} {
			if db != nil {
				defer db.Close()
			}
		}
		// Stop on db errors
		for _, err = range []error{errRatingDb, errAccDb, errStorDb} {
			if err != nil {
				log.Fatalf("Could not open database connection: %v", err)
			}
		}
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
		)
	}
	tpReader := engine.NewTpReader(ratingDb, accountDb, loader, *tpid, *timezone, *loadHistorySize)
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
	if *historyServer != "" { // Init scribeAgent so we can store the differences
		if scribeAgent, err := rpcclient.NewRpcClient("tcp", *historyServer, 3, 3, utils.GOB, nil); err != nil {
			log.Fatalf("Could not connect to history server, error: %s. Make sure you have properly configured it via -history_server flag.", err.Error())
			return
		} else {
			engine.SetHistoryScribe(scribeAgent)
			//defer scribeAgent.Client.Close()
		}
	} else {
		log.Print("WARNING: Rates history archiving is disabled!")
	}
	if *raterAddress != "" { // Init connection to rater so we can reload it's data
		rater, err = rpc.Dial("tcp", *raterAddress)
		if err != nil {
			log.Fatalf("Could not connect to rater: %s", err.Error())
			return
		}
	} else {
		log.Print("WARNING: Rates automatic cache reloading is disabled!")
	}
	if *cdrstatsAddress != "" { // Init connection to rater so we can reload it's data
		if *cdrstatsAddress == *raterAddress {
			cdrstats = rater
		} else {
			cdrstats, err = rpc.Dial("tcp", *cdrstatsAddress)
			if err != nil {
				log.Fatalf("Could not connect to CDRStats API: %s", err.Error())
				return
			}
		}
	} else {
		log.Print("WARNING: CDRStats automatic data reload is disabled!")
	}
	if *usersAddress != "" { // Init connection to rater so we can reload it's data
		if *usersAddress == *raterAddress {
			users = rater
		} else {
			users, err = rpc.Dial("tcp", *usersAddress)
			if err != nil {
				log.Fatalf("Could not connect to Users API: %s", err.Error())
				return
			}
		}
	} else {
		log.Print("WARNING: Users automatic data reload is disabled!")
	}

	// write maps to database
	if err := tpReader.WriteToDatabase(*flush, *verbose); err != nil {
		log.Fatal("Could not write to database: ", err)
	}
	if len(*historyServer) != 0 && *verbose {
		log.Print("Wrote history.")
	}
	var dstIds, rplIds, rpfIds, actIds, shgIds, alsIds, lcrIds, dcsIds []string
	if rater != nil {
		dstIds, _ = tpReader.GetLoadedIds(utils.DESTINATION_PREFIX)
		rplIds, _ = tpReader.GetLoadedIds(utils.RATING_PLAN_PREFIX)
		rpfIds, _ = tpReader.GetLoadedIds(utils.RATING_PROFILE_PREFIX)
		actIds, _ = tpReader.GetLoadedIds(utils.ACTION_PREFIX)
		shgIds, _ = tpReader.GetLoadedIds(utils.SHARED_GROUP_PREFIX)
		alsIds, _ = tpReader.GetLoadedIds(utils.ALIASES_PREFIX)
		lcrIds, _ = tpReader.GetLoadedIds(utils.LCR_PREFIX)
		dcsIds, _ = tpReader.GetLoadedIds(utils.DERIVEDCHARGERS_PREFIX)
	}
	actTmgIds, _ := tpReader.GetLoadedIds(utils.ACTION_PLAN_PREFIX)
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
		if *flush {
			dstIds, rplIds, rpfIds, lcrIds = nil, nil, nil, nil // Should reload all these on flush
		}
		if err = rater.Call("ApierV1.ReloadCache", utils.ApiReloadCache{
			DestinationIds:   dstIds,
			RatingPlanIds:    rplIds,
			RatingProfileIds: rpfIds,
			ActionIds:        actIds,
			SharedGroupIds:   shgIds,
			Aliases:          alsIds,
			LCRIds:           lcrIds,
			DerivedChargers:  dcsIds,
		}, &reply); err != nil {
			log.Printf("WARNING: Got error on cache reload: %s\n", err.Error())
		}

		if len(actTmgIds) != 0 {
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
}
