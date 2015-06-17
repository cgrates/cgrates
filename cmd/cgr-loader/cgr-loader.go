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

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/history"
	"github.com/cgrates/cgrates/utils"
)

var (
	//separator = flag.String("separator", ",", "Default field separator")
	cgrConfig, _  = config.NewDefaultCGRConfig()
	ratingdb_type = flag.String("ratingdb_type", cgrConfig.RatingDBType, "The type of the RatingDb database <redis>")
	ratingdb_host = flag.String("ratingdb_host", cgrConfig.RatingDBHost, "The RatingDb host to connect to.")
	ratingdb_port = flag.String("ratingdb_port", cgrConfig.RatingDBPort, "The RatingDb port to bind to.")
	ratingdb_name = flag.String("ratingdb_name", cgrConfig.RatingDBName, "The name/number of the RatingDb to connect to.")
	ratingdb_user = flag.String("ratingdb_user", cgrConfig.RatingDBUser, "The RatingDb user to sign in as.")
	ratingdb_pass = flag.String("ratingdb_passwd", cgrConfig.RatingDBPass, "The RatingDb user's password.")

	accountdb_type = flag.String("accountdb_type", cgrConfig.AccountDBType, "The type of the AccountingDb database <redis>")
	accountdb_host = flag.String("accountdb_host", cgrConfig.AccountDBHost, "The AccountingDb host to connect to.")
	accountdb_port = flag.String("accountdb_port", cgrConfig.AccountDBPort, "The AccountingDb port to bind to.")
	accountdb_name = flag.String("accountdb_name", cgrConfig.AccountDBName, "The name/number of the AccountingDb to connect to.")
	accountdb_user = flag.String("accountdb_user", cgrConfig.AccountDBUser, "The AccountingDb user to sign in as.")
	accountdb_pass = flag.String("accountdb_passwd", cgrConfig.AccountDBPass, "The AccountingDb user's password.")

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
	runId           = flag.String("runid", "", "Uniquely identify an import/load, postpended to some automatic fields")
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
	var rater, cdrstats *rpc.Client
	var loader engine.LoadReader
	// Init necessary db connections, only if not already
	if !*dryRun { // make sure we do not need db connections on dry run, also not importing into any stordb
		if *fromStorDb {
			ratingDb, errRatingDb = engine.ConfigureRatingStorage(*ratingdb_type, *ratingdb_host, *ratingdb_port, *ratingdb_name,
				*ratingdb_user, *ratingdb_pass, *dbdata_encoding)
			accountDb, errAccDb = engine.ConfigureAccountingStorage(*accountdb_type, *accountdb_host, *accountdb_port, *accountdb_name, *accountdb_user, *accountdb_pass, *dbdata_encoding)
			storDb, errStorDb = engine.ConfigureLoadStorage(*stor_db_type, *stor_db_host, *stor_db_port, *stor_db_name, *stor_db_user, *stor_db_pass, *dbdata_encoding,
				cgrConfig.StorDBMaxOpenConns, cgrConfig.StorDBMaxIdleConns)
		} else if *toStorDb { // Import from csv files to storDb
			storDb, errStorDb = engine.ConfigureLoadStorage(*stor_db_type, *stor_db_host, *stor_db_port, *stor_db_name, *stor_db_user, *stor_db_pass, *dbdata_encoding,
				cgrConfig.StorDBMaxOpenConns, cgrConfig.StorDBMaxIdleConns)
		} else { // Default load from csv files to dataDb
			ratingDb, errRatingDb = engine.ConfigureRatingStorage(*ratingdb_type, *ratingdb_host, *ratingdb_port, *ratingdb_name,
				*ratingdb_user, *ratingdb_pass, *dbdata_encoding)
			accountDb, errAccDb = engine.ConfigureAccountingStorage(*accountdb_type, *accountdb_host, *accountdb_port, *accountdb_name, *accountdb_user, *accountdb_pass, *dbdata_encoding)
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
			path.Join(*dataPath, utils.CDR_STATS_CSV))
	}
	tpReader := engine.NewTpReader(ratingDb, accountDb, loader, *tpid)
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
		if scribeAgent, err := history.NewProxyScribe(*historyServer); err != nil {
			log.Fatalf("Could not connect to history server, error: %s. Make sure you have properly configured it via -history_server flag.", err.Error())
			return
		} else {
			engine.SetHistoryScribe(scribeAgent)
			defer scribeAgent.Client.Close()
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

	// write maps to database
	if err := tpReader.WriteToDatabase(*flush, *verbose); err != nil {
		log.Fatal("Could not write to database: ", err)
	}
	if len(*historyServer) != 0 && *verbose {
		log.Print("Wrote history.")
	}
	// Reload scheduler and cache
	if rater != nil {
		reply := ""
		dstIds, _ := tpReader.GetLoadedIds(utils.DESTINATION_PREFIX)
		rplIds, _ := tpReader.GetLoadedIds(utils.RATING_PLAN_PREFIX)
		rpfIds, _ := tpReader.GetLoadedIds(utils.RATING_PROFILE_PREFIX)
		actIds, _ := tpReader.GetLoadedIds(utils.ACTION_PREFIX)
		shgIds, _ := tpReader.GetLoadedIds(utils.SHARED_GROUP_PREFIX)
		rpAliases, _ := tpReader.GetLoadedIds(utils.RP_ALIAS_PREFIX)
		accAliases, _ := tpReader.GetLoadedIds(utils.ACC_ALIAS_PREFIX)
		lcrIds, _ := tpReader.GetLoadedIds(utils.LCR_PREFIX)
		dcs, _ := tpReader.GetLoadedIds(utils.DERIVEDCHARGERS_PREFIX)
		// Reload cache first since actions could be calling info from within
		if *verbose {
			log.Print("Reloading cache")
		}
		if *flush {
			dstIds, rplIds, rpfIds, rpAliases, lcrIds = nil, nil, nil, nil, nil // Should reload all these on flush
		}
		if err = rater.Call("ApierV1.ReloadCache", utils.ApiReloadCache{
			DestinationIds:   dstIds,
			RatingPlanIds:    rplIds,
			RatingProfileIds: rpfIds,
			ActionIds:        actIds,
			SharedGroupIds:   shgIds,
			RpAliases:        rpAliases,
			AccAliases:       accAliases,
			LCRIds:           lcrIds,
			DerivedChargers:  dcs,
		}, &reply); err != nil {
			log.Printf("WARNING: Got error on cache reload: %s\n", err.Error())
		}
		actTmgIds, _ := tpReader.GetLoadedIds(utils.ACTION_TIMING_PREFIX)
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
		statsQueueIds, _ := tpReader.GetLoadedIds(utils.CDR_STATS_PREFIX)
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
}
