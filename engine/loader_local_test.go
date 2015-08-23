/*
Real-Time Charging System for Telecom Environments
Copyright (C) 2012-2015 ITsysCOM GmbH

This program is free software: you can Storagetribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITH*out ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>
*/

package engine

import (
	"flag"
	"path"
	"testing"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
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
var ratingDbCsv, ratingDbStor, ratingDbApier RatingStorage        // Each ratingDb will have it's own sources to collect data
var accountDbCsv, accountDbStor, accountDbApier AccountingStorage // Each ratingDb will have it's own sources to collect data
var storDb LoadStorage
var lCfg *config.CGRConfig

var tpCsvScenario = flag.String("tp_scenario", "prepaid1centpsec", "Use this scenario folder to import tp csv data from")

// Create connection to ratingDb
// Will use 3 different datadbs in order to be able to see differences in data loaded
func TestConnDataDbs(t *testing.T) {
	if !*testLocal {
		return
	}
	lCfg, _ = config.NewDefaultCGRConfig()
	var err error
	if ratingDbCsv, err = ConfigureRatingStorage(lCfg.TpDbType, lCfg.TpDbHost, lCfg.TpDbPort, "4", lCfg.TpDbUser, lCfg.TpDbPass, lCfg.DBDataEncoding); err != nil {
		t.Fatal("Error on ratingDb connection: ", err.Error())
	}
	if ratingDbStor, err = ConfigureRatingStorage(lCfg.TpDbType, lCfg.TpDbHost, lCfg.TpDbPort, "5", lCfg.TpDbUser, lCfg.TpDbPass, lCfg.DBDataEncoding); err != nil {
		t.Fatal("Error on ratingDb connection: ", err.Error())
	}
	if ratingDbApier, err = ConfigureRatingStorage(lCfg.TpDbType, lCfg.TpDbHost, lCfg.TpDbPort, "6", lCfg.TpDbUser, lCfg.TpDbPass, lCfg.DBDataEncoding); err != nil {
		t.Fatal("Error on ratingDb connection: ", err.Error())
	}
	if accountDbCsv, err = ConfigureAccountingStorage(lCfg.DataDbType, lCfg.DataDbHost, lCfg.DataDbPort, "7",
		lCfg.DataDbUser, lCfg.DataDbPass, lCfg.DBDataEncoding); err != nil {
		t.Fatal("Error on ratingDb connection: ", err.Error())
	}
	if accountDbStor, err = ConfigureAccountingStorage(lCfg.DataDbType, lCfg.DataDbHost, lCfg.DataDbPort, "8",
		lCfg.DataDbUser, lCfg.DataDbPass, lCfg.DBDataEncoding); err != nil {
		t.Fatal("Error on ratingDb connection: ", err.Error())
	}
	if accountDbApier, err = ConfigureAccountingStorage(lCfg.DataDbType, lCfg.DataDbHost, lCfg.DataDbPort, "9",
		lCfg.DataDbUser, lCfg.DataDbPass, lCfg.DBDataEncoding); err != nil {
		t.Fatal("Error on ratingDb connection: ", err.Error())
	}
	for _, db := range []Storage{ratingDbCsv, ratingDbStor, ratingDbApier, accountDbCsv, accountDbStor, accountDbApier} {
		if err = db.Flush(""); err != nil {
			t.Fatal("Error when flushing datadb")
		}
	}

}

// Create/reset storage tariff plan tables, used as database connectin establishment also
func TestCreateStorTpTables(t *testing.T) {
	if !*testLocal {
		return
	}
	var db *MySQLStorage
	if d, err := NewMySQLStorage(lCfg.StorDBHost, lCfg.StorDBPort, lCfg.StorDBName, lCfg.StorDBUser, lCfg.StorDBPass, lCfg.StorDBMaxOpenConns, lCfg.StorDBMaxIdleConns); err != nil {
		t.Error("Error on opening database connection: ", err)
		return
	} else {
		db = d.(*MySQLStorage)
		storDb = d.(LoadStorage)
	}
	// Creating the table serves also as reset since there is a drop prior to create
	if err := db.CreateTablesFromScript(path.Join(*dataDir, "storage", "mysql", utils.CREATE_TARIFFPLAN_TABLES_SQL)); err != nil {
		t.Error("Error on db creation: ", err.Error())
		return // No point in going further
	}
}

// Loads data from csv files in tp scenario to ratingDbCsv
func TestLoadFromCSV(t *testing.T) {
	if !*testLocal {
		return
	}
	/*var err error
	for fn, v := range FileValidators {
		if err = ValidateCSVData(path.Join(*dataDir, "tariffplans", *tpCsvScenario, fn), v.Rule); err != nil {
			t.Error("Failed validating data: ", err.Error())
		}
	}*/
	loader := NewTpReader(ratingDbCsv, accountDbCsv, NewFileCSVStorage(utils.CSV_SEP,
		path.Join(*dataDir, "tariffplans", *tpCsvScenario, utils.DESTINATIONS_CSV),
		path.Join(*dataDir, "tariffplans", *tpCsvScenario, utils.TIMINGS_CSV),
		path.Join(*dataDir, "tariffplans", *tpCsvScenario, utils.RATES_CSV),
		path.Join(*dataDir, "tariffplans", *tpCsvScenario, utils.DESTINATION_RATES_CSV),
		path.Join(*dataDir, "tariffplans", *tpCsvScenario, utils.RATING_PLANS_CSV),
		path.Join(*dataDir, "tariffplans", *tpCsvScenario, utils.RATING_PROFILES_CSV),
		path.Join(*dataDir, "tariffplans", *tpCsvScenario, utils.SHARED_GROUPS_CSV),
		path.Join(*dataDir, "tariffplans", *tpCsvScenario, utils.LCRS_CSV),
		path.Join(*dataDir, "tariffplans", *tpCsvScenario, utils.ACTIONS_CSV),
		path.Join(*dataDir, "tariffplans", *tpCsvScenario, utils.ACTION_PLANS_CSV),
		path.Join(*dataDir, "tariffplans", *tpCsvScenario, utils.ACTION_TRIGGERS_CSV),
		path.Join(*dataDir, "tariffplans", *tpCsvScenario, utils.ACCOUNT_ACTIONS_CSV),
		path.Join(*dataDir, "tariffplans", *tpCsvScenario, utils.DERIVED_CHARGERS_CSV),
		path.Join(*dataDir, "tariffplans", *tpCsvScenario, utils.CDR_STATS_CSV),

		path.Join(*dataDir, "tariffplans", *tpCsvScenario, utils.USERS_CSV),
		path.Join(*dataDir, "tariffplans", *tpCsvScenario, utils.ALIASES_CSV),
	), "", "", lCfg.LoadHistorySize)

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
	if err = loader.LoadActionPlans(); err != nil {
		t.Error("Failed loading action timings: ", err.Error())
	}
	if err = loader.LoadActionTriggers(); err != nil {
		t.Error("Failed loading action triggers: ", err.Error())
	}
	if err = loader.LoadAccountActions(); err != nil {
		t.Error("Failed loading account actions: ", err.Error())
	}
	if err = loader.LoadDerivedChargers(); err != nil {
		t.Error("Failed loading derived chargers: ", err.Error())
	}
	if err = loader.LoadUsers(); err != nil {
		t.Error("Failed loading users: ", err.Error())
	}
	if err = loader.LoadAliases(); err != nil {
		t.Error("Failed loading aliases: ", err.Error())
	}
	if err := loader.WriteToDatabase(true, false); err != nil {
		t.Error("Could not write data into ratingDb: ", err.Error())
	}
}

// Imports data from csv files in tpScenario to storDb
func TestImportToStorDb(t *testing.T) {
	if !*testLocal {
		return
	}
	csvImporter := TPCSVImporter{
		TPid:     utils.TEST_SQL,
		StorDb:   storDb,
		DirPath:  path.Join(*dataDir, "tariffplans", *tpCsvScenario),
		Sep:      utils.CSV_SEP,
		Verbose:  false,
		ImportId: utils.TEST_SQL}
	if err := csvImporter.Run(); err != nil {
		t.Error("Error when importing tpdata to storDb: ", err)
	}
	if tpids, err := storDb.GetTpIds(); err != nil {
		t.Error("Error when querying storDb for imported data: ", err)
	} else if len(tpids) != 1 || tpids[0] != utils.TEST_SQL {
		t.Errorf("Data in storDb is different than expected %v", tpids)
	}
}

// Loads data from storDb into ratingDb
func TestLoadFromStorDb(t *testing.T) {
	if !*testLocal {
		return
	}
	loader := NewTpReader(ratingDbStor, accountDbStor, storDb, utils.TEST_SQL, "", lCfg.LoadHistorySize)
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
	if err := loader.LoadActionPlans(); err != nil {
		t.Error("Failed loading action timings: ", err.Error())
	}
	if err := loader.LoadActionTriggers(); err != nil {
		t.Error("Failed loading action triggers: ", err.Error())
	}
	if err := loader.LoadAccountActions(); err != nil {
		t.Error("Failed loading account actions: ", err.Error())
	}
	if err := loader.LoadDerivedChargers(); err != nil {
		t.Error("Failed loading derived chargers: ", err.Error())
	}
	if err := loader.WriteToDatabase(true, false); err != nil {
		t.Error("Could not write data into ratingDb: ", err.Error())
	}
}

func TestLoadIndividualProfiles(t *testing.T) {
	if !*testLocal {
		return
	}
	loader := NewTpReader(ratingDbApier, accountDbApier, storDb, utils.TEST_SQL, "", lCfg.LoadHistorySize)
	// Load ratingPlans. This will also set destination keys
	if ratingPlans, err := storDb.GetTpRatingPlans(utils.TEST_SQL, "", nil); err != nil {
		t.Fatal("Could not retrieve rating plans")
	} else {
		rpls, err := TpRatingPlans(ratingPlans).GetRatingPlans()
		if err != nil {
			t.Fatal("Could not convert rating plans")
		}
		for tag := range rpls {
			if loaded, err := loader.LoadRatingPlansFiltered(tag); err != nil {
				t.Fatalf("Could not load ratingPlan for tag: %s, error: %s", tag, err.Error())
			} else if !loaded {
				t.Fatal("Cound not find ratingPLan with id:", tag)
			}
		}
	}
	// Load rating profiles
	loadId := utils.CSV_LOAD + "_" + utils.TEST_SQL
	if ratingProfiles, err := storDb.GetTpRatingProfiles(&TpRatingProfile{Tpid: utils.TEST_SQL, Loadid: loadId}); err != nil {
		t.Fatal("Could not retrieve rating profiles, error: ", err.Error())
	} else if len(ratingProfiles) == 0 {
		t.Fatal("Could not retrieve rating profiles")
	} else {
		rpfs, err := TpRatingProfiles(ratingProfiles).GetRatingProfiles()
		if err != nil {
			t.Fatal("Could not convert rating profiles")
		}
		for rpId := range rpfs {
			rp, _ := utils.NewTPRatingProfileFromKeyId(utils.TEST_SQL, loadId, rpId)
			mrp := APItoModelRatingProfile(rp)
			if err := loader.LoadRatingProfilesFiltered(&mrp[0]); err != nil {
				t.Fatalf("Could not load ratingProfile with id: %s, error: %s", rpId, err.Error())
			}
		}
	}
	// Load derived chargers
	loadId = utils.CSV_LOAD + "_" + utils.TEST_SQL
	if derivedChargers, err := storDb.GetTpDerivedChargers(&TpDerivedCharger{Tpid: utils.TEST_SQL, Loadid: loadId}); err != nil {
		t.Fatal("Could not retrieve derived chargers, error: ", err.Error())
	} else if len(derivedChargers) == 0 {
		t.Fatal("Could not retrieve derived chargers")
	} else {
		dcs, err := TpDerivedChargers(derivedChargers).GetDerivedChargers()
		if err != nil {
			t.Fatal("Could not convert derived chargers")
		}
		for dcId := range dcs {
			mdc := &TpDerivedCharger{Tpid: utils.TEST_SQL, Loadid: loadId}
			mdc.SetDerivedChargersId(dcId)
			if err := loader.LoadDerivedChargersFiltered(mdc, true); err != nil {
				t.Fatalf("Could not load derived charger with id: %s, error: %s", dcId, err.Error())
			}
		}
	}
	// Load cdr stats
	//loadId = utils.CSV_LOAD + "_" + utils.TEST_SQL
	if cdrStats, err := storDb.GetTpCdrStats(utils.TEST_SQL, ""); err != nil {
		t.Fatal("Could not retrieve cdr stats, error: ", err.Error())
	} else if len(cdrStats) == 0 {
		t.Fatal("Could not retrieve cdr stats")
	} else {
		cds, err := TpCdrStats(cdrStats).GetCdrStats()
		if err != nil {
			t.Fatal("Could not convert cdr stats")
		}
		for id := range cds {
			if err := loader.LoadCdrStatsFiltered(id, true); err != nil {
				t.Fatalf("Could not load cdr stats with id: %s, error: %s", id, err.Error())
			}
		}
	}
	// Load account actions
	if accountActions, err := storDb.GetTpAccountActions(&TpAccountAction{Tpid: utils.TEST_SQL, Loadid: loadId}); err != nil {
		t.Fatal("Could not retrieve account action profiles, error: ", err.Error())
	} else if len(accountActions) == 0 {
		t.Error("No account actions")
	} else {
		aas, err := TpAccountActions(accountActions).GetAccountActions()
		if err != nil {
			t.Fatal("Could not convert account actions")
		}
		for aaId := range aas {
			aa, _ := utils.NewTPAccountActionsFromKeyId(utils.TEST_SQL, loadId, aaId)
			maa := APItoModelAccountAction(aa)

			if err := loader.LoadAccountActionsFiltered(maa); err != nil {
				t.Fatalf("Could not load account actions with id: %s, error: %s", aaId, err.Error())
			}
		}
	}
}

// Compares previously loaded data from csv and stor to be identical, redis specific tests
func TestMatchLoadCsvWithStor(t *testing.T) {
	if !*testLocal {
		return
	}
	rsCsv, redisDb := ratingDbCsv.(*RedisStorage)
	if !redisDb {
		return // We only support these tests for redis
	}
	rsStor := ratingDbStor.(*RedisStorage)
	rsApier := ratingDbApier.(*RedisStorage)
	keysCsv, err := rsCsv.db.Keys("*")
	if err != nil {
		t.Fatal("Failed querying redis keys for csv data")
	}
	for _, key := range keysCsv {
		var refVal []byte
		for idx, rs := range []*RedisStorage{rsCsv, rsStor, rsApier} {
			qVal, err := rs.db.Get(key)
			if err != nil {
				t.Fatalf("Run: %d, could not retrieve key %s, error: %s", idx, key, err.Error())
			}
			if idx == 0 { // Only compare at second iteration, first one is to set reference value
				refVal = qVal
				continue
			}
			if len(refVal) != len(qVal) {
				t.Errorf("Missmatched data for key: %s\n\t, reference val: %s \n\t retrieved value: %s\n on iteration: %d", key, refVal, qVal, idx)
			}
		}
	}
}
