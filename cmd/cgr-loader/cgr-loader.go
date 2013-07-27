/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2013 ITsysCOM

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
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/cgrates/config"
	"log"
	"path"
)

var (
	//separator = flag.String("separator", ",", "Default field separator")
	cgrConfig,_ = config.NewDefaultCGRConfig()
	data_db_type = flag.String("datadb_type", cgrConfig.DataDBType, "The type of the dataDb database (redis|mongo|postgres|mysql)")
	data_db_host = flag.String("datadb_host", cgrConfig.DataDBHost, "The dataDb host to connect to.")
	data_db_port = flag.String("datadb_port", cgrConfig.DataDBPort, "The dataDb port to bind to.")
	data_db_name = flag.String("datadb_name", cgrConfig.DataDBName, "The name/number of the dataDb to connect to.")
	data_db_user = flag.String("datadb_user", cgrConfig.DataDBUser, "The dataDb user to sign in as.")
	data_db_pass = flag.String("datadb_passwd", cgrConfig.DataDBPass,  "The dataDb user's password.")

	stor_db_type = flag.String("stordb_type", cgrConfig.StorDBType, "The type of the storDb database (redis|mongo|postgres|mysql)")
	stor_db_host = flag.String("stordb_host", cgrConfig.StorDBHost, "The storDb host to connect to.")
	stor_db_port = flag.String("stordb_port", cgrConfig.StorDBPort, "The storDb port to bind to.")
	stor_db_name = flag.String("stordb_name", cgrConfig.StorDBName, "The name/number of the storDb to connect to.")
	stor_db_user = flag.String("stordb_user", cgrConfig.StorDBUser, "The storDb user to sign in as.")
	stor_db_pass = flag.String("stordb_passwd", cgrConfig.StorDBPass, "The storDb user's password.")

	flush    = flag.Bool("flush", false, "Flush the database before importing")
	tpid = flag.String("tpid", "", "The tariff plan id from the database")
	dataPath = flag.String("path", ".", "The path containing the data files")
	version  = flag.Bool("version", false, "Prints the application version.")
	verbose  = flag.Bool("verbose", false, "Enable detailed verbose logging output")
	fromStorDb = flag.Bool("from_stordb", false, "Load the tariff plan from storDb to dataDb")
	toStorDb = flag.Bool("to_stordb", false, "Import the tariff plan from files to storDb")

)

func main() {
	flag.Parse()
	if *version {
		fmt.Println("CGRateS " + utils.VERSION)
		return
	}
	var errDataDb, errStorDb, err error
	var dataDb, storDb engine.DataStorage
	// Init necessary db connections
	if *fromStorDb {
		dataDb, errDataDb = engine.ConfigureDatabase(*data_db_type, *data_db_host, *data_db_port, *data_db_name, *data_db_user, *data_db_pass)
		storDb, errStorDb = engine.ConfigureDatabase(*stor_db_type, *stor_db_host, *stor_db_port, *stor_db_name, *stor_db_user, *stor_db_pass)
	} else if *toStorDb { // Import from csv files to storDb
		storDb, errStorDb = engine.ConfigureDatabase(*stor_db_type, *stor_db_host, *stor_db_port, *stor_db_name, *stor_db_user, *stor_db_pass)
	} else { // Default load from csv files to dataDb
		dataDb, errDataDb = engine.ConfigureDatabase(*data_db_type, *data_db_host, *data_db_port, *data_db_name, *data_db_user, *data_db_pass)
	}
	// Defer databases opened to be closed when we are done
	for _,db := range []engine.DataStorage{ dataDb, storDb } {
		if db != nil { defer db.Close() }
	}
	// Stop on db errors
	for _,err = range []error{errDataDb, errStorDb} {
		if err != nil {
			log.Fatalf("Could not open database connection: %v", err)
		}
	}
	var loader engine.TPLoader
	if *fromStorDb { // Load Tariff Plan from storDb into dataDb
		loader = engine.NewDbReader(storDb, dataDb, *tpid)
	} else if *toStorDb { // Import files from a directory into storDb
		if *tpid == "" {
			log.Fatal("TPid required, please define it via *-tpid* command argument.")
		}
		csvImporter := engine.TPCSVImporter{ *tpid, storDb, *dataPath, ',', *verbose }
		if errImport := csvImporter.Run(); errImport != nil {
			log.Fatal(errImport)
		}
		return
	} else { // Default load from csv files to dataDb
		for fn, v := range engine.FileValidators {
			err := engine.ValidateCSVData(path.Join(*dataPath, fn), v.Rule)
			if err != nil {
				log.Fatal(err, "\n\t", v.Message)
			}
		}
		loader = engine.NewFileCSVReader(dataDb, ',', utils.DESTINATIONS_CSV, utils.TIMINGS_CSV, utils.RATES_CSV, utils.DESTINATION_RATES_CSV, utils.DESTRATE_TIMINGS_CSV, utils.RATE_PROFILES_CSV, utils.ACTIONS_CSV, utils.ACTION_TIMINGS_CSV, utils.ACTION_TRIGGERS_CSV, utils.ACCOUNT_ACTIONS_CSV)
	}

	err = loader.LoadDestinations()
	if err != nil {
		log.Fatal(err)
	}
	err = loader.LoadTimings()
	if err != nil {
		log.Fatal(err)
	}
	err = loader.LoadRates()
	if err != nil {
		log.Fatal(err)
	}
	err = loader.LoadDestinationRates()
	if err != nil {
		log.Fatal(err)
	}
	err = loader.LoadDestinationRateTimings()
	if err != nil {
		log.Fatal(err)
	}
	err = loader.LoadRatingProfiles()
	if err != nil {
		log.Fatal(err)
	}
	err = loader.LoadActions()
	if err != nil {
		log.Fatal(err)
	}
	err = loader.LoadActionTimings()
	if err != nil {
		log.Fatal(err)
	}
	err = loader.LoadActionTriggers()
	if err != nil {
		log.Fatal(err)
	}
	err = loader.LoadAccountActions()
	if err != nil {
		log.Fatal(err)
	}

	// write maps to database
	if err := loader.WriteToDatabase(*flush, true); err != nil {
		log.Fatal("Could not write to database: ", err)
	}
}
