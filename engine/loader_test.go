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

package engine

import (
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/cgrates/config"
	"testing"
	"path"
	//"fmt"
	"flag"
)

/*
README:

 Enable local tests by passing '-local' to the go test command
 Tests in this file combine end2end tests using both redis and MySQL.
 It is expected that the data folder of CGRateS exists at path /usr/share/cgrates/data or passed via command arguments.
 Prior running the tests, create database and users by running:
  mysql -pyourrootpwd < /usr/share/cgrates/data/storage/mysql/create_db_with_users.sql
 What these tests do:
  * Connect to redis using 2 handles, one where we store CSV reference data and one where we store data out of storDb, each with it's own db number
  * Flush data in each handle to start clean
*/

// Globals used
var dataDbCsv, dataDbStor DataStorage // Test data coming from csv files getting inthere
var storDb LoadStorage
var cfg *config.CGRConfig

// Arguments received via test command
var testLocal = flag.Bool("local", false, "Perform the tests only on local test environment, not by default.") // This flag will be passed here via "go test -local" args
var dataDir = flag.String("data_dir", "/usr/share/cgrates/data", "CGR data dir path here")
var tpCsvScenario = flag.String("tp_scenario", "prepaid1centpsec", "Use this scenario folder to import tp csv data from")


// Create connection to dataDb
// Will use 3 different datadbs in order to be able to see differences in data loaded
func TestConnDataDbs(t *testing.T) {
	if !*testLocal { 
		return
	}
	cfg,_ = config.NewDefaultCGRConfig()
	var err error
	if dataDbCsv, err = ConfigureDataStorage(cfg.DataDBType, cfg.DataDBHost, cfg.DataDBPort, "13", cfg.DataDBUser, cfg.DataDBPass, cfg.DBDataEncoding); err != nil {
		t.Fatal("Error on dataDb connection: ", err.Error())
	}
	if dataDbStor, err = ConfigureDataStorage(cfg.DataDBType, cfg.DataDBHost, cfg.DataDBPort, "14", cfg.DataDBUser, cfg.DataDBPass, cfg.DBDataEncoding); err != nil {
		t.Fatal("Error on dataDb connection: ", err.Error())
	}
}

// Create/reset storage tariff plan tables, used as database connectin establishment also
func TestCreateStorTpTables(t *testing.T) {
	if !*testLocal { 
		return
	}
	var db *MySQLStorage
	if d, err := NewMySQLStorage(cfg.StorDBHost, cfg.StorDBPort, cfg.StorDBName, cfg.StorDBUser, cfg.StorDBPass); err != nil {
		t.Error("Error on opening database connection: ",err)
		return
	} else {
		db = d.(*MySQLStorage)
		storDb = d.(LoadStorage)
	}
	// Creating the table serves also as reset since there is a drop prior to create
	if err := db.CreateTablesFromScript(path.Join(*dataDir, "storage", "mysql", CREATE_TARIFFPLAN_TABLES_SQL)); err != nil {
		t.Error("Error on db creation: ", err.Error())
		return // No point in going further
	}
}

// Import data from csv files into dataDbCsv
func TestLoadFromCSV(t *testing.T) {
	if !*testLocal { 
		return
	}
	var err error
	for fn, v := range FileValidators {
		if err = ValidateCSVData(path.Join(*dataDir, "tariffplans", *tpCsvScenario, fn), v.Rule); err != nil {
			t.Error("Failed validating data: ", err.Error())
		}
	}
	loader := NewFileCSVReader(dataDbCsv, utils.CSV_SEP, 
			path.Join(*dataDir, "tariffplans", *tpCsvScenario, utils.DESTINATIONS_CSV),
			path.Join(*dataDir, "tariffplans", *tpCsvScenario, utils.TIMINGS_CSV),
			path.Join(*dataDir, "tariffplans", *tpCsvScenario, utils.RATES_CSV),
			path.Join(*dataDir, "tariffplans", *tpCsvScenario, utils.DESTINATION_RATES_CSV),
			path.Join(*dataDir, "tariffplans", *tpCsvScenario, utils.RATING_PLANS_CSV),
			path.Join(*dataDir, "tariffplans", *tpCsvScenario, utils.RATING_PROFILES_CSV),
			path.Join(*dataDir, "tariffplans", *tpCsvScenario, utils.ACTIONS_CSV),
			path.Join(*dataDir, "tariffplans", *tpCsvScenario, utils.ACTION_TIMINGS_CSV),
			path.Join(*dataDir, "tariffplans", *tpCsvScenario, utils.ACTION_TRIGGERS_CSV),
			path.Join(*dataDir, "tariffplans", *tpCsvScenario, utils.ACCOUNT_ACTIONS_CSV),
			)

	if err = loader.LoadDestinations(); err != nil {
		t.Error("Failed loading destinations: ", err.Error())
	}
	if err = loader.LoadTimings(); err != nil {
		t.Error("Failed loading timings: ", err.Error())
	}
	if err = loader.LoadRates(); err != nil {
		t.Error("Failed loading rates: ", err.Error())
	}
	if err = loader.LoadDestinationRates(); err != nil {
		t.Error("Failed loading destination rates: ", err.Error())
	}
	if err = loader.LoadRatingPlans(); err != nil {
		t.Error("Failed loading rating plans: ", err.Error())
	}
	if err = loader.LoadRatingProfiles(); err != nil {
		t.Error("Failed loading rating profiles: ", err.Error())
	}
	if err = loader.LoadActions(); err != nil {
		t.Error("Failed loading actions: ", err.Error())
	}
	if err = loader.LoadActionTimings(); err != nil {
		t.Error("Failed loading action timings: ", err.Error())
	}
	if err = loader.LoadActionTriggers(); err != nil {
		t.Error("Failed loading action triggers: ", err.Error())
	}
	if err = loader.LoadAccountActions(); err != nil {
		t.Error("Failed loading account actions: ", err.Error())
	}
	if err := loader.WriteToDatabase(true, false); err != nil {
		t.Error("Could not write data into dataDb: ", err.Error())
	}
}


func TestImportToStorDb(t *testing.T) {
	if !*testLocal {
		return
	}
	csvImporter := TPCSVImporter{TEST_SQL, storDb, path.Join(*dataDir, "tariffplans", *tpCsvScenario), utils.CSV_SEP, false, TEST_SQL}
	if err := csvImporter.Run(); err != nil {
		t.Error("Error when importing tpdata to storDb: ", err)
	}
	if tpids, err := storDb.GetTPIds(); err != nil {
		t.Error("Error when querying storDb for imported data: ", err)
	} else if len(tpids) != 1 || tpids[0] != TEST_SQL {
		t.Errorf("Data in storDb is different than expected %v", tpids)
	}
}

func TestLoadFromStorDb(t *testing.T) {
	if !*testLocal {
		return
	}
	loader := NewDbReader(storDb, dataDbStor, TEST_SQL)
	if err := loader.LoadDestinations(); err != nil {
		t.Error("Failed loading destinations: ", err.Error())
	}
	if err := loader.LoadTimings(); err != nil {
		t.Error("Failed loading timings: ", err.Error())
	}
	if err := loader.LoadRates(); err != nil {
		t.Error("Failed loading rates: ", err.Error())
	}
	if err := loader.LoadDestinationRates(); err != nil {
		t.Error("Failed loading destination rates: ", err.Error())
	}
	if err := loader.LoadRatingPlans(); err != nil {
		t.Error("Failed loading rating plans: ", err.Error())
	}
	if err := loader.LoadRatingProfiles(); err != nil {
		t.Error("Failed loading rating profiles: ", err.Error())
	}
	if err := loader.LoadActions(); err != nil {
		t.Error("Failed loading actions: ", err.Error())
	}
	if err := loader.LoadActionTimings(); err != nil {
		t.Error("Failed loading action timings: ", err.Error())
	}
	if err := loader.LoadActionTriggers(); err != nil {
		t.Error("Failed loading action triggers: ", err.Error())
	}
	if err := loader.LoadAccountActions(); err != nil {
		t.Error("Failed loading account actions: ", err.Error())
	}
	if err := loader.WriteToDatabase(true, false); err != nil {
		t.Error("Could not write data into dataDb: ", err.Error())
	}
}
