// +build integration

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
	"flag"
	"log"
	"path"
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	Move            = "move"
	dbtype          string
	mig             *Migrator
	dataDir         = flag.String("data_dir", "/usr/share/cgrates", "CGR data dir path here")
	loadHistorySize = flag.Int("load_history_size", config.CgrConfig().LoadHistorySize, "Limit the number of records in the load history")
)

// subtests to be executed for each migrator
var sTestsITMigrator = []func(t *testing.T){
	testFlush,
	testMigratorAccounts,
	testMigratorActionPlans,
	testMigratorActionTriggers,
	testMigratorActions,
	testMigratorSharedGroups,
	testMigratorStats,
	testMigratorTPRatingProfile,
	testFlush,
}

func TestMigratorITPostgresConnect(t *testing.T) {
	cdrsPostgresCfgPath := path.Join(*dataDir, "conf", "samples", "tutpostgres")
	postgresITCfg, err := config.NewCGRConfigFromFolder(cdrsPostgresCfgPath)
	if err != nil {
		t.Fatal(err)
	}
	dataDB, err := engine.ConfigureDataStorage(postgresITCfg.DataDbType, postgresITCfg.DataDbHost, postgresITCfg.DataDbPort, postgresITCfg.DataDbName, postgresITCfg.DataDbUser, postgresITCfg.DataDbPass, postgresITCfg.DBDataEncoding, config.CgrConfig().CacheCfg(), *loadHistorySize)
	if err != nil {
		log.Fatal(err)
	}
	dataDB2, err := engine.ConfigureDataStorage(postgresITCfg.DataDbType, postgresITCfg.DataDbHost, postgresITCfg.DataDbPort, postgresITCfg.DataDbName, postgresITCfg.DataDbUser, postgresITCfg.DataDbPass, postgresITCfg.DBDataEncoding, config.CgrConfig().CacheCfg(), *loadHistorySize)
	if err != nil {
		log.Fatal(err)
	}
	oldDataDB, err := ConfigureV1DataStorage(postgresITCfg.DataDbType, postgresITCfg.DataDbHost, postgresITCfg.DataDbPort, postgresITCfg.DataDbName, postgresITCfg.DataDbUser, postgresITCfg.DataDbPass, postgresITCfg.DBDataEncoding)
	if err != nil {
		log.Fatal(err)
	}
	storDB, err := engine.ConfigureStorStorage(postgresITCfg.StorDBType, postgresITCfg.StorDBHost, postgresITCfg.StorDBPort, postgresITCfg.StorDBName, postgresITCfg.StorDBUser, postgresITCfg.StorDBPass, postgresITCfg.DBDataEncoding,
		config.CgrConfig().StorDBMaxOpenConns, config.CgrConfig().StorDBMaxIdleConns, config.CgrConfig().StorDBConnMaxLifetime, config.CgrConfig().StorDBCDRSIndexes)
	if err != nil {
		log.Fatal(err)
	}
	oldstorDB, err := engine.ConfigureStorStorage(postgresITCfg.StorDBType, postgresITCfg.StorDBHost, postgresITCfg.StorDBPort, postgresITCfg.StorDBName, postgresITCfg.StorDBUser, postgresITCfg.StorDBPass, postgresITCfg.DBDataEncoding,
		config.CgrConfig().StorDBMaxOpenConns, config.CgrConfig().StorDBMaxIdleConns, config.CgrConfig().StorDBConnMaxLifetime, config.CgrConfig().StorDBCDRSIndexes)
	if err != nil {
		log.Fatal(err)
	}
	mig, err = NewMigrator(dataDB, dataDB2, postgresITCfg.DataDbType, postgresITCfg.DBDataEncoding, storDB, postgresITCfg.StorDBType, oldDataDB, postgresITCfg.DataDbType, postgresITCfg.DBDataEncoding, oldstorDB, postgresITCfg.StorDBType, false, true)
	if err != nil {
		log.Fatal(err)
	}
}

func TestMigratorITPostgres(t *testing.T) {
	dbtype = utils.REDIS
	log.Print("REDIS+POSTGRES")
	for _, stest := range sTestsITMigrator {
		t.Run("TestITMigratorOnPostgres", stest)
	}
}

func TestMigratorITRedisConnect(t *testing.T) {
	cdrsMysqlCfgPath := path.Join(*dataDir, "conf", "samples", "tutmysql")
	mysqlITCfg, err := config.NewCGRConfigFromFolder(cdrsMysqlCfgPath)
	if err != nil {
		t.Fatal(err)
	}
	dataDB, err := engine.ConfigureDataStorage(mysqlITCfg.DataDbType, mysqlITCfg.DataDbHost, mysqlITCfg.DataDbPort, mysqlITCfg.DataDbName, mysqlITCfg.DataDbUser, mysqlITCfg.DataDbPass, mysqlITCfg.DBDataEncoding, config.CgrConfig().CacheCfg(), *loadHistorySize)
	if err != nil {
		log.Fatal(err)
	}
	dataDB2, err := engine.ConfigureDataStorage(mysqlITCfg.DataDbType, mysqlITCfg.DataDbHost, mysqlITCfg.DataDbPort, mysqlITCfg.DataDbName, mysqlITCfg.DataDbUser, mysqlITCfg.DataDbPass, mysqlITCfg.DBDataEncoding, config.CgrConfig().CacheCfg(), *loadHistorySize)
	if err != nil {
		log.Fatal(err)
	}
	oldDataDB, err := ConfigureV1DataStorage(mysqlITCfg.DataDbType, mysqlITCfg.DataDbHost, mysqlITCfg.DataDbPort, mysqlITCfg.DataDbName, mysqlITCfg.DataDbUser, mysqlITCfg.DataDbPass, mysqlITCfg.DBDataEncoding)
	if err != nil {
		log.Fatal(err)
	}
	storDB, err := engine.ConfigureStorStorage(mysqlITCfg.StorDBType, mysqlITCfg.StorDBHost, mysqlITCfg.StorDBPort, mysqlITCfg.StorDBName, mysqlITCfg.StorDBUser, mysqlITCfg.StorDBPass, mysqlITCfg.DBDataEncoding,
		config.CgrConfig().StorDBMaxOpenConns, config.CgrConfig().StorDBMaxIdleConns, config.CgrConfig().StorDBConnMaxLifetime, config.CgrConfig().StorDBCDRSIndexes)
	if err != nil {
		log.Fatal(err)
	}
	oldstorDB, err := engine.ConfigureStorStorage(mysqlITCfg.StorDBType, mysqlITCfg.StorDBHost, mysqlITCfg.StorDBPort, mysqlITCfg.StorDBName, mysqlITCfg.StorDBUser, mysqlITCfg.StorDBPass, mysqlITCfg.DBDataEncoding,
		config.CgrConfig().StorDBMaxOpenConns, config.CgrConfig().StorDBMaxIdleConns, config.CgrConfig().StorDBConnMaxLifetime, config.CgrConfig().StorDBCDRSIndexes)
	if err != nil {
		log.Fatal(err)
	}
	mig, err = NewMigrator(dataDB, dataDB2, mysqlITCfg.DataDbType, mysqlITCfg.DBDataEncoding, storDB, mysqlITCfg.StorDBType, oldDataDB, mysqlITCfg.DataDbType, mysqlITCfg.DBDataEncoding, oldstorDB, mysqlITCfg.StorDBType, false, true)
	if err != nil {
		log.Fatal(err)
	}
}

func TestMigratorITRedis(t *testing.T) {
	dbtype = utils.REDIS
	log.Print("REDIS+MYSQL")
	for _, stest := range sTestsITMigrator {
		t.Run("TestITMigratorOnRedis", stest)
	}
}

func TestMigratorITMongoConnect(t *testing.T) {
	cdrsMongoCfgPath := path.Join(*dataDir, "conf", "samples", "tutmongo")
	mgoITCfg, err := config.NewCGRConfigFromFolder(cdrsMongoCfgPath)
	if err != nil {
		t.Fatal(err)
	}
	dataDB, err := engine.ConfigureDataStorage(mgoITCfg.DataDbType, mgoITCfg.DataDbHost, mgoITCfg.DataDbPort, mgoITCfg.DataDbName, mgoITCfg.DataDbUser, mgoITCfg.DataDbPass, mgoITCfg.DBDataEncoding, config.CgrConfig().CacheCfg(), *loadHistorySize)
	if err != nil {
		log.Fatal(err)
	}
	dataDB2, err := engine.ConfigureDataStorage(mgoITCfg.DataDbType, mgoITCfg.DataDbHost, mgoITCfg.DataDbPort, mgoITCfg.DataDbName, mgoITCfg.DataDbUser, mgoITCfg.DataDbPass, mgoITCfg.DBDataEncoding, config.CgrConfig().CacheCfg(), *loadHistorySize)
	if err != nil {
		log.Fatal(err)
	}
	oldDataDB, err := ConfigureV1DataStorage(mgoITCfg.DataDbType, mgoITCfg.DataDbHost, mgoITCfg.DataDbPort, mgoITCfg.DataDbName, mgoITCfg.DataDbUser, mgoITCfg.DataDbPass, mgoITCfg.DBDataEncoding)
	if err != nil {
		log.Fatal(err)
	}
	storDB, err := engine.ConfigureStorStorage(mgoITCfg.StorDBType, mgoITCfg.StorDBHost, mgoITCfg.StorDBPort, mgoITCfg.StorDBName, mgoITCfg.StorDBUser, mgoITCfg.StorDBPass, mgoITCfg.DBDataEncoding,
		config.CgrConfig().StorDBMaxOpenConns, config.CgrConfig().StorDBMaxIdleConns, config.CgrConfig().StorDBConnMaxLifetime, config.CgrConfig().StorDBCDRSIndexes)
	if err != nil {
		log.Fatal(err)
	}
	oldstorDB, err := engine.ConfigureStorStorage(mgoITCfg.StorDBType, mgoITCfg.StorDBHost, mgoITCfg.StorDBPort, mgoITCfg.StorDBName, mgoITCfg.StorDBUser, mgoITCfg.StorDBPass, mgoITCfg.DBDataEncoding,
		config.CgrConfig().StorDBMaxOpenConns, config.CgrConfig().StorDBMaxIdleConns, config.CgrConfig().StorDBConnMaxLifetime, config.CgrConfig().StorDBCDRSIndexes)
	if err != nil {
		log.Fatal(err)
	}
	mig, err = NewMigrator(dataDB, dataDB2, mgoITCfg.DataDbType, mgoITCfg.DBDataEncoding, storDB, mgoITCfg.StorDBType, oldDataDB, mgoITCfg.DataDbType, mgoITCfg.DBDataEncoding, oldstorDB, mgoITCfg.StorDBType, false, true)
	if err != nil {
		log.Fatal(err)
	}
}

func TestMigratorITMongo(t *testing.T) {
	dbtype = utils.MONGO
	log.Print("MONGO")
	for _, stest := range sTestsITMigrator {
		t.Run("TestITMigratorOnMongo", stest)
	}
}

func TestMigratorITMoveConnect(t *testing.T) {
	cdrsMongoCfgPath := path.Join(*dataDir, "conf", "samples", "tutmongo")
	mgoITCfg, err := config.NewCGRConfigFromFolder(cdrsMongoCfgPath)
	if err != nil {
		t.Fatal(err)
	}
	cdrsMysqlCfgPath := path.Join(*dataDir, "conf", "samples", "tutmysql")
	mysqlITCfg, err := config.NewCGRConfigFromFolder(cdrsMysqlCfgPath)
	if err != nil {
		t.Fatal(err)
	}
	dataDB, err := engine.ConfigureDataStorage(mgoITCfg.DataDbType, mgoITCfg.DataDbHost, mgoITCfg.DataDbPort, mgoITCfg.DataDbName, mgoITCfg.DataDbUser, mgoITCfg.DataDbPass, mgoITCfg.DBDataEncoding, config.CgrConfig().CacheCfg(), *loadHistorySize)
	if err != nil {
		log.Fatal(err)
	}
	dataDB2, err := engine.ConfigureDataStorage(mysqlITCfg.DataDbType, mysqlITCfg.DataDbHost, mysqlITCfg.DataDbPort, mysqlITCfg.DataDbName, mysqlITCfg.DataDbUser, mysqlITCfg.DataDbPass, mysqlITCfg.DBDataEncoding, config.CgrConfig().CacheCfg(), *loadHistorySize)
	if err != nil {
		log.Fatal(err)
	}
	oldDataDB, err := ConfigureV1DataStorage(mysqlITCfg.DataDbType, mysqlITCfg.DataDbHost, mysqlITCfg.DataDbPort, mysqlITCfg.DataDbName, mysqlITCfg.DataDbUser, mysqlITCfg.DataDbPass, mysqlITCfg.DBDataEncoding)
	if err != nil {
		log.Fatal(err)
	}
	storDB, err := engine.ConfigureStorStorage(mgoITCfg.StorDBType, mgoITCfg.StorDBHost, mgoITCfg.StorDBPort, mgoITCfg.StorDBName, mgoITCfg.StorDBUser, mgoITCfg.StorDBPass, mgoITCfg.DBDataEncoding,
		config.CgrConfig().StorDBMaxOpenConns, config.CgrConfig().StorDBMaxIdleConns, config.CgrConfig().StorDBConnMaxLifetime, config.CgrConfig().StorDBCDRSIndexes)
	if err != nil {
		log.Fatal(err)
	}
	oldstorDB, err := engine.ConfigureStorStorage(mysqlITCfg.StorDBType, mysqlITCfg.StorDBHost, mysqlITCfg.StorDBPort, mysqlITCfg.StorDBName, mysqlITCfg.StorDBUser, mysqlITCfg.StorDBPass, mysqlITCfg.DBDataEncoding,
		config.CgrConfig().StorDBMaxOpenConns, config.CgrConfig().StorDBMaxIdleConns, config.CgrConfig().StorDBConnMaxLifetime, config.CgrConfig().StorDBCDRSIndexes)
	if err != nil {
		log.Fatal(err)
	}
	mig, err = NewMigrator(dataDB2, dataDB, mgoITCfg.DataDbType, mgoITCfg.DBDataEncoding, storDB, mgoITCfg.StorDBType, oldDataDB, mgoITCfg.DataDbType, mgoITCfg.DBDataEncoding, oldstorDB, mgoITCfg.StorDBType, false, false)
	if err != nil {
		log.Fatal(err)
	}
}

func TestMigratorITMove(t *testing.T) {
	dbtype = Move
	log.Print("Move")
	for _, stest := range sTestsITMigrator {
		t.Run("TestITMigratorOnMongo", stest)
	}
}

func testFlush(t *testing.T) {
	mig.dmOut.DataDB().Flush("")
	if err := engine.SetDBVersions(mig.dmOut.DataDB()); err != nil {
		t.Error("Error  ", err.Error())
	}

}

func testMigratorAccounts(t *testing.T) {
	v1d := &v1Balance{Value: 100000, Weight: 10, DestinationIds: "NAT", ExpirationDate: time.Date(2012, 1, 1, 0, 0, 0, 0, time.UTC).Local(), Timings: []*engine.RITiming{&engine.RITiming{Years: utils.Years{}, Months: utils.Months{}, MonthDays: utils.MonthDays{}, WeekDays: utils.WeekDays{}}}}
	v1b := &v1Balance{Value: 100000, Weight: 10, DestinationIds: "NAT", ExpirationDate: time.Date(2012, 1, 1, 0, 0, 0, 0, time.UTC).Local(), Timings: []*engine.RITiming{&engine.RITiming{Years: utils.Years{}, Months: utils.Months{}, MonthDays: utils.MonthDays{}, WeekDays: utils.WeekDays{}}}}
	v1Acc := &v1Account{Id: "*OUT:CUSTOMER_1:rif", BalanceMap: map[string]v1BalanceChain{utils.DATA: v1BalanceChain{v1d}, utils.VOICE: v1BalanceChain{v1b}, utils.MONETARY: v1BalanceChain{&v1Balance{Value: 21, ExpirationDate: time.Date(2012, 1, 1, 0, 0, 0, 0, time.UTC).Local(), Timings: []*engine.RITiming{&engine.RITiming{Years: utils.Years{}, Months: utils.Months{}, MonthDays: utils.MonthDays{}, WeekDays: utils.WeekDays{}}}}}}}

	v2d := &engine.Balance{Uuid: "", ID: "", Value: 100000, Directions: utils.StringMap{"*OUT": true}, ExpirationDate: time.Date(2012, 1, 1, 0, 0, 0, 0, time.UTC).Local(), Weight: 10, DestinationIDs: utils.StringMap{"NAT": true},
		RatingSubject: "", Categories: utils.NewStringMap(), SharedGroups: utils.NewStringMap(), Timings: []*engine.RITiming{&engine.RITiming{Years: utils.Years{}, Months: utils.Months{}, MonthDays: utils.MonthDays{}, WeekDays: utils.WeekDays{}}}, TimingIDs: utils.NewStringMap(""), Factor: engine.ValueFactor{}}
	v2b := &engine.Balance{Uuid: "", ID: "", Value: 0.0001, Directions: utils.StringMap{"*OUT": true}, ExpirationDate: time.Date(2012, 1, 1, 0, 0, 0, 0, time.UTC).Local(), Weight: 10, DestinationIDs: utils.StringMap{"NAT": true},
		RatingSubject: "", Categories: utils.NewStringMap(), SharedGroups: utils.NewStringMap(), Timings: []*engine.RITiming{&engine.RITiming{Years: utils.Years{}, Months: utils.Months{}, MonthDays: utils.MonthDays{}, WeekDays: utils.WeekDays{}}}, TimingIDs: utils.NewStringMap(""), Factor: engine.ValueFactor{}}
	m2 := &engine.Balance{Uuid: "", ID: "", Value: 21, Directions: utils.StringMap{"*OUT": true}, ExpirationDate: time.Date(2012, 1, 1, 0, 0, 0, 0, time.UTC).Local(), DestinationIDs: utils.NewStringMap(""), RatingSubject: "",
		Categories: utils.NewStringMap(), SharedGroups: utils.NewStringMap(), Timings: []*engine.RITiming{&engine.RITiming{Years: utils.Years{}, Months: utils.Months{}, MonthDays: utils.MonthDays{}, WeekDays: utils.WeekDays{}}}, TimingIDs: utils.NewStringMap(""), Factor: engine.ValueFactor{}}
	testAccount := &engine.Account{ID: "CUSTOMER_1:rif", BalanceMap: map[string]engine.Balances{utils.DATA: engine.Balances{v2d}, utils.VOICE: engine.Balances{v2b}, utils.MONETARY: engine.Balances{m2}}, UnitCounters: engine.UnitCounters{}, ActionTriggers: engine.ActionTriggers{}}
	switch {
	case dbtype == utils.REDIS:
		err := mig.oldDataDB.setV1Account(v1Acc)
		if err != nil {
			t.Error("Error when setting v1 acc ", err.Error())
		}
		currentVersion := engine.Versions{utils.StatS: 2, utils.Thresholds: 2, utils.Accounts: 1, utils.Actions: 2, utils.ActionTriggers: 2, utils.ActionPlans: 2, utils.SharedGroups: 2}
		err = mig.dmOut.DataDB().SetVersions(currentVersion, false)
		if err != nil {
			t.Error("Error when setting version for stats ", err.Error())
		}
		err, _ = mig.Migrate([]string{utils.MetaAccounts})
		if err != nil {
			t.Error("Error when migrating accounts ", err.Error())
		}
		result, err := mig.dmOut.DataDB().GetAccount(testAccount.ID)
		if err != nil {
			t.Error("Error when getting account ", err.Error())
		}
		if !reflect.DeepEqual(testAccount.BalanceMap["*voice"][0], result.BalanceMap["*voice"][0]) {
			t.Errorf("Expecting: %+v, received: %+v", testAccount.BalanceMap["*voice"][0], result.BalanceMap["*voice"][0])
		} else if !reflect.DeepEqual(testAccount, result) {
			t.Errorf("Expecting: %+v, received: %+v", testAccount, result)
		}
	case dbtype == utils.MONGO:
		err := mig.oldDataDB.setV1Account(v1Acc)
		if err != nil {
			t.Error("Error when marshaling ", err.Error())
		}
		currentVersion := engine.Versions{utils.StatS: 2, utils.Thresholds: 2, utils.Accounts: 1, utils.Actions: 2, utils.ActionTriggers: 2, utils.ActionPlans: 2, utils.SharedGroups: 2}
		err = mig.dmOut.DataDB().SetVersions(currentVersion, false)
		if err != nil {
			t.Error("Error when setting version for stats ", err.Error())
		}
		err, _ = mig.Migrate([]string{utils.MetaAccounts})
		if err != nil {
			t.Error("Error when migrating accounts ", err.Error())
		}
		result, err := mig.dmOut.DataDB().GetAccount(testAccount.ID)
		if err != nil {
			t.Error("Error when getting account ", err.Error())
		}
		if !reflect.DeepEqual(testAccount, result) {
			t.Errorf("Expecting: %+v, received: %+v", testAccount, result)
		}
	case dbtype == Move:
		if err := mig.dmIN.DataDB().SetAccount(testAccount); err != nil {
			log.Print("GOT ERR DMIN", err)
		}
		currentVersion := engine.CurrentDataDBVersions()
		err := mig.dmOut.DataDB().SetVersions(currentVersion, false)
		if err != nil {
			t.Error("Error when setting version for stats ", err.Error())
		}
		err, _ = mig.Migrate([]string{utils.MetaAccounts})
		if err != nil {
			t.Error("Error when migrating accounts ", err.Error())
		}
		result, err := mig.dmOut.DataDB().GetAccount(testAccount.ID)
		if err != nil {
			log.Print("GOT ERR DMOUT", err)
		}
		if !reflect.DeepEqual(testAccount, result) {
			t.Errorf("Expecting: %+v, received: %+v", testAccount, result)
		}
	}
}

func testMigratorActionPlans(t *testing.T) {
	v1ap := &v1ActionPlans{&v1ActionPlan{Id: "test", AccountIds: []string{"one"}, Timing: &engine.RateInterval{Timing: &engine.RITiming{Years: utils.Years{}, Months: utils.Months{}, MonthDays: utils.MonthDays{}, WeekDays: utils.WeekDays{}}}}}
	ap := &engine.ActionPlan{Id: "test", AccountIDs: utils.StringMap{"one": true}, ActionTimings: []*engine.ActionTiming{&engine.ActionTiming{Timing: &engine.RateInterval{Timing: &engine.RITiming{Years: utils.Years{}, Months: utils.Months{}, MonthDays: utils.MonthDays{}, WeekDays: utils.WeekDays{}}}}}}
	switch {
	case dbtype == utils.REDIS:
		err := mig.oldDataDB.setV1ActionPlans(v1ap)
		if err != nil {
			t.Error("Error when setting v1 ActionPlan ", err.Error())
		}
		currentVersion := engine.Versions{utils.StatS: 2, utils.Thresholds: 2, utils.Accounts: 2, utils.Actions: 2, utils.ActionTriggers: 2, utils.ActionPlans: 1, utils.SharedGroups: 2}
		err = mig.dmOut.DataDB().SetVersions(currentVersion, false)
		if err != nil {
			t.Error("Error when setting version for stats ", err.Error())
		}
		err, _ = mig.Migrate([]string{utils.MetaActionPlans})
		if err != nil {
			t.Error("Error when migrating ActionPlans ", err.Error())
		}
		result, err := mig.dmOut.DataDB().GetActionPlan(ap.Id, true, utils.NonTransactional)
		if err != nil {
			t.Error("Error when getting ActionPlan ", err.Error())
		}
		if ap.Id != result.Id || !reflect.DeepEqual(ap.AccountIDs, result.AccountIDs) {
			t.Errorf("Expecting: %+v, received: %+v", *ap, result)
		} else if !reflect.DeepEqual(ap.ActionTimings[0].Timing, result.ActionTimings[0].Timing) {
			t.Errorf("Expecting: %+v, received: %+v", ap.ActionTimings[0].Timing, result.ActionTimings[0].Timing)
		} else if ap.ActionTimings[0].Weight != result.ActionTimings[0].Weight || ap.ActionTimings[0].ActionsID != result.ActionTimings[0].ActionsID {
			t.Errorf("Expecting: %+v, received: %+v", ap.ActionTimings[0].Weight, result.ActionTimings[0].Weight)
		}
	case dbtype == utils.MONGO:
		err := mig.oldDataDB.setV1ActionPlans(v1ap)
		if err != nil {
			t.Error("Error when setting v1 ActionPlans ", err.Error())
		}
		currentVersion := engine.Versions{utils.StatS: 2, utils.Thresholds: 2, utils.Accounts: 2, utils.Actions: 2, utils.ActionTriggers: 2, utils.ActionPlans: 1, utils.SharedGroups: 2}
		err = mig.dmOut.DataDB().SetVersions(currentVersion, false)
		if err != nil {
			t.Error("Error when setting version for stats ", err.Error())
		}
		err, _ = mig.Migrate([]string{utils.MetaActionPlans})
		if err != nil {
			t.Error("Error when migrating ActionPlans ", err.Error())
		}
		result, err := mig.dmOut.DataDB().GetActionPlan(ap.Id, true, utils.NonTransactional)
		if err != nil {
			t.Error("Error when getting ActionPlan ", err.Error())
		}
		if ap.Id != result.Id || !reflect.DeepEqual(ap.AccountIDs, result.AccountIDs) {
			t.Errorf("Expecting: %+v, received: %+v", *ap, result)
		} else if !reflect.DeepEqual(ap.ActionTimings[0].Timing, result.ActionTimings[0].Timing) {
			t.Errorf("Expecting: %+v, received: %+v", ap.ActionTimings[0].Timing, result.ActionTimings[0].Timing)
		} else if ap.ActionTimings[0].Weight != result.ActionTimings[0].Weight || ap.ActionTimings[0].ActionsID != result.ActionTimings[0].ActionsID {
			t.Errorf("Expecting: %+v, received: %+v", ap.ActionTimings[0].Weight, result.ActionTimings[0].Weight)
		}
	case dbtype == Move:
		if err := mig.dmIN.DataDB().SetActionPlan(ap.Id, ap, true, utils.NonTransactional); err != nil {
			t.Error("Error when setting ActionPlan ", err.Error())
		}
		currentVersion := engine.CurrentDataDBVersions()
		err := mig.dmOut.DataDB().SetVersions(currentVersion, false)
		if err != nil {
			t.Error("Error when setting version for stats ", err.Error())
		}
		err, _ = mig.Migrate([]string{utils.MetaActionPlans})
		if err != nil {
			t.Error("Error when migrating ActionPlans ", err.Error())
		}
		result, err := mig.dmOut.DataDB().GetActionPlan(ap.Id, true, utils.NonTransactional)
		if err != nil {
			t.Error("Error when getting ActionPlan ", err.Error())
		}
		if ap.Id != result.Id || !reflect.DeepEqual(ap.AccountIDs, result.AccountIDs) {
			t.Errorf("Expecting: %+v, received: %+v", *ap, result)
		} else if !reflect.DeepEqual(ap.ActionTimings[0].Timing, result.ActionTimings[0].Timing) {
			t.Errorf("Expecting: %+v, received: %+v", ap.ActionTimings[0].Timing, result.ActionTimings[0].Timing)
		} else if ap.ActionTimings[0].Weight != result.ActionTimings[0].Weight || ap.ActionTimings[0].ActionsID != result.ActionTimings[0].ActionsID {
			t.Errorf("Expecting: %+v, received: %+v", ap.ActionTimings[0].Weight, result.ActionTimings[0].Weight)
		}
	}
}

func testMigratorActionTriggers(t *testing.T) {
	tim := time.Date(2012, time.February, 27, 23, 59, 59, 0, time.UTC).Local()
	v1atrs := &v1ActionTriggers{
		&v1ActionTrigger{
			Id:                    "Test",
			BalanceType:           "*monetary",
			BalanceDirection:      "*out",
			ThresholdType:         "*max_balance",
			ThresholdValue:        2,
			ActionsId:             "TEST_ACTIONS",
			Executed:              true,
			BalanceExpirationDate: tim,
		},
	}
	atrs := engine.ActionTriggers{
		&engine.ActionTrigger{
			ID: "Test",
			Balance: &engine.BalanceFilter{
				Timings:        []*engine.RITiming{},
				ExpirationDate: utils.TimePointer(tim),
				Type:           utils.StringPointer(utils.MONETARY),
				Directions:     utils.StringMapPointer(utils.NewStringMap(utils.OUT)),
			},
			ExpirationDate:    tim,
			LastExecutionTime: tim,
			ActivationDate:    tim,
			ThresholdType:     utils.TRIGGER_MAX_BALANCE,
			ThresholdValue:    2,
			ActionsID:         "TEST_ACTIONS",
			Executed:          true,
		},
	}
	switch {
	case dbtype == utils.REDIS:
		err := mig.oldDataDB.setV1ActionTriggers(v1atrs)
		if err != nil {
			t.Error("Error when setting v1 ActionTriggers ", err.Error())
		}
		currentVersion := engine.Versions{utils.StatS: 2, utils.Thresholds: 2, utils.Accounts: 2, utils.Actions: 2, utils.ActionTriggers: 1, utils.ActionPlans: 2, utils.SharedGroups: 2}
		err = mig.dmOut.DataDB().SetVersions(currentVersion, false)
		if err != nil {
			t.Error("Error when setting version for stats ", err.Error())
		}
		err, _ = mig.Migrate([]string{utils.MetaActionTriggers})
		if err != nil {
			t.Error("Error when migrating ActionTriggers ", err.Error())
		}
		result, err := mig.dmOut.GetActionTriggers((*v1atrs)[0].Id, true, utils.NonTransactional)
		if err != nil {
			t.Error("Error when getting ActionTriggers ", err.Error())
		}
		if !reflect.DeepEqual(atrs[0].ID, result[0].ID) {
			t.Errorf("Expecting: %+v, received: %+v", atrs[0].ID, result[0].ID)
		} else if !reflect.DeepEqual(atrs[0].UniqueID, result[0].UniqueID) {
			t.Errorf("Expecting: %+v, received: %+v", atrs[0].UniqueID, result[0].UniqueID)
		} else if !reflect.DeepEqual(atrs[0].ThresholdType, result[0].ThresholdType) {
			t.Errorf("Expecting: %+v, received: %+v", atrs[0].ThresholdType, result[0].ThresholdType)
		} else if !reflect.DeepEqual(atrs[0].ThresholdValue, result[0].ThresholdValue) {
			t.Errorf("Expecting: %+v, received: %+v", atrs[0].ThresholdValue, result[0].ThresholdValue)
		} else if !reflect.DeepEqual(atrs[0].Recurrent, result[0].Recurrent) {
			t.Errorf("Expecting: %+v, received: %+v", atrs[0].Recurrent, result[0].Recurrent)
		} else if !reflect.DeepEqual(atrs[0].MinSleep, result[0].MinSleep) {
			t.Errorf("Expecting: %+v, received: %+v", atrs[0].MinSleep, result[0].MinSleep)
		} else if !reflect.DeepEqual(atrs[0].ExpirationDate, result[0].ExpirationDate) {
			t.Errorf("Expecting: %+v, received: %+v", atrs[0].ExpirationDate, result[0].ExpirationDate)
		} else if !reflect.DeepEqual(atrs[0].ActivationDate, result[0].ActivationDate) {
			t.Errorf("Expecting: %+v, received: %+v", atrs[0].ActivationDate, result[0].ActivationDate)
		} else if !reflect.DeepEqual(atrs[0].Balance.Type, result[0].Balance.Type) {
			t.Errorf("Expecting: %+v, received: %+v", atrs[0].Balance.Type, result[0].Balance.Type)
		} else if !reflect.DeepEqual(atrs[0].Weight, result[0].Weight) {
			t.Errorf("Expecting: %+v, received: %+v", atrs[0].Weight, result[0].Weight)
		} else if !reflect.DeepEqual(atrs[0].ActionsID, result[0].ActionsID) {
			t.Errorf("Expecting: %+v, received: %+v", atrs[0].ActionsID, result[0].ActionsID)
		} else if !reflect.DeepEqual(atrs[0].MinQueuedItems, result[0].MinQueuedItems) {
			t.Errorf("Expecting: %+v, received: %+v", atrs[0].MinQueuedItems, result[0].MinQueuedItems)
		} else if !reflect.DeepEqual(atrs[0].Executed, result[0].Executed) {
			t.Errorf("Expecting: %+v, received: %+v", atrs[0].Executed, result[0].Executed)
		} else if !reflect.DeepEqual(atrs[0].LastExecutionTime, result[0].LastExecutionTime) {
			t.Errorf("Expecting: %+v, received: %+v", atrs[0].LastExecutionTime, result[0].LastExecutionTime)
		}
		//Testing each field of balance
		if !reflect.DeepEqual(atrs[0].Balance.Uuid, result[0].Balance.Uuid) {
			t.Errorf("Expecting: %+v, received: %+v", atrs[0].Balance.Uuid, result[0].Balance.Uuid)
		} else if !reflect.DeepEqual(atrs[0].Balance.ID, result[0].Balance.ID) {
			t.Errorf("Expecting: %+v, received: %+v", atrs[0].Balance.ID, result[0].Balance.ID)
		} else if !reflect.DeepEqual(atrs[0].Balance.Type, result[0].Balance.Type) {
			t.Errorf("Expecting: %+v, received: %+v", atrs[0].Balance.Type, result[0].Balance.Type)
		} else if !reflect.DeepEqual(atrs[0].Balance.Value, result[0].Balance.Value) {
			t.Errorf("Expecting: %+v, received: %+v", atrs[0].Balance.Value, result[0].Balance.Value)
		} else if !reflect.DeepEqual(atrs[0].Balance.Directions, result[0].Balance.Directions) {
			t.Errorf("Expecting: %+v, received: %+v", atrs[0].Balance.Directions, result[0].Balance.Directions)
		} else if !reflect.DeepEqual(atrs[0].Balance.ExpirationDate, result[0].Balance.ExpirationDate) {
			t.Errorf("Expecting: %+v, received: %+v", atrs[0].Balance.ExpirationDate, result[0].Balance.ExpirationDate)
		} else if !reflect.DeepEqual(atrs[0].Balance.Weight, result[0].Balance.Weight) {
			t.Errorf("Expecting: %+v, received: %+v", atrs[0].Balance.Weight, result[0].Balance.Weight)
		} else if !reflect.DeepEqual(atrs[0].Balance.DestinationIDs, result[0].Balance.DestinationIDs) {
			t.Errorf("Expecting: %+v, received: %+v", atrs[0].Balance.DestinationIDs, result[0].Balance.DestinationIDs)
		} else if !reflect.DeepEqual(atrs[0].Balance.RatingSubject, result[0].Balance.RatingSubject) {
			t.Errorf("Expecting: %+v, received: %+v", atrs[0].Balance.RatingSubject, result[0].Balance.RatingSubject)
		} else if !reflect.DeepEqual(atrs[0].Balance.Categories, result[0].Balance.Categories) {
			t.Errorf("Expecting: %+v, received: %+v", atrs[0].Balance.Categories, result[0].Balance.Categories)
		} else if !reflect.DeepEqual(atrs[0].Balance.SharedGroups, result[0].Balance.SharedGroups) {
			t.Errorf("Expecting: %+v, received: %+v", atrs[0].Balance.SharedGroups, result[0].Balance.SharedGroups)
		} else if !reflect.DeepEqual(atrs[0].Balance.TimingIDs, result[0].Balance.TimingIDs) {
			t.Errorf("Expecting: %+v, received: %+v", atrs[0].Balance.TimingIDs, result[0].Balance.TimingIDs)
		} else if !reflect.DeepEqual(atrs[0].Balance.TimingIDs, result[0].Balance.TimingIDs) {
			t.Errorf("Expecting: %+v, received: %+v", atrs[0].Balance.Timings, result[0].Balance.Timings)
		} else if !reflect.DeepEqual(atrs[0].Balance.Disabled, result[0].Balance.Disabled) {
			t.Errorf("Expecting: %+v, received: %+v", atrs[0].Balance.Disabled, result[0].Balance.Disabled)
		} else if !reflect.DeepEqual(atrs[0].Balance.Factor, result[0].Balance.Factor) {
			t.Errorf("Expecting: %+v, received: %+v", atrs[0].Balance.Factor, result[0].Balance.Factor)
		} else if !reflect.DeepEqual(atrs[0].Balance.Blocker, result[0].Balance.Blocker) {
			t.Errorf("Expecting: %+v, received: %+v", atrs[0].Balance.Blocker, result[0].Balance.Blocker)
		}
	case dbtype == utils.MONGO:
		err, _ := mig.Migrate([]string{utils.MetaActionTriggers})
		if err != nil && err != utils.ErrNotImplemented {
			t.Error("Error when migrating ActionTriggers ", err.Error())
		}

	case dbtype == Move:
		if err := mig.dmIN.SetActionTriggers(atrs[0].ID, atrs, utils.NonTransactional); err != nil {
			t.Error("Error when setting ActionPlan ", err.Error())
		}
		currentVersion := engine.CurrentDataDBVersions()
		err := mig.dmOut.DataDB().SetVersions(currentVersion, false)
		if err != nil {
			t.Error("Error when setting version for stats ", err.Error())
		}
		err, _ = mig.Migrate([]string{utils.MetaActionTriggers})
		if err != nil {
			t.Error("Error when migrating ActionPlans ", err.Error())
		}
		result, err := mig.dmOut.GetActionTriggers(atrs[0].ID, true, utils.NonTransactional)
		if err != nil {
			t.Error("Error when getting ActionTriggers ", err.Error())
		}
		if !reflect.DeepEqual(atrs[0].ID, result[0].ID) {
			t.Errorf("Expecting: %+v, received: %+v", atrs[0].ID, result[0].ID)
		} else if !reflect.DeepEqual(atrs[0].UniqueID, result[0].UniqueID) {
			t.Errorf("Expecting: %+v, received: %+v", atrs[0].UniqueID, result[0].UniqueID)
		} else if !reflect.DeepEqual(atrs[0].ThresholdType, result[0].ThresholdType) {
			t.Errorf("Expecting: %+v, received: %+v", atrs[0].ThresholdType, result[0].ThresholdType)
		} else if !reflect.DeepEqual(atrs[0].ThresholdValue, result[0].ThresholdValue) {
			t.Errorf("Expecting: %+v, received: %+v", atrs[0].ThresholdValue, result[0].ThresholdValue)
		} else if !reflect.DeepEqual(atrs[0].Recurrent, result[0].Recurrent) {
			t.Errorf("Expecting: %+v, received: %+v", atrs[0].Recurrent, result[0].Recurrent)
		} else if !reflect.DeepEqual(atrs[0].MinSleep, result[0].MinSleep) {
			t.Errorf("Expecting: %+v, received: %+v", atrs[0].MinSleep, result[0].MinSleep)
		} else if !reflect.DeepEqual(atrs[0].ExpirationDate, result[0].ExpirationDate) {
			t.Errorf("Expecting: %+v, received: %+v", atrs[0].ExpirationDate, result[0].ExpirationDate)
		} else if !reflect.DeepEqual(atrs[0].ActivationDate, result[0].ActivationDate) {
			t.Errorf("Expecting: %+v, received: %+v", atrs[0].ActivationDate, result[0].ActivationDate)
		} else if !reflect.DeepEqual(atrs[0].Balance.Type, result[0].Balance.Type) {
			t.Errorf("Expecting: %+v, received: %+v", atrs[0].Balance.Type, result[0].Balance.Type)
		} else if !reflect.DeepEqual(atrs[0].Weight, result[0].Weight) {
			t.Errorf("Expecting: %+v, received: %+v", atrs[0].Weight, result[0].Weight)
		} else if !reflect.DeepEqual(atrs[0].ActionsID, result[0].ActionsID) {
			t.Errorf("Expecting: %+v, received: %+v", atrs[0].ActionsID, result[0].ActionsID)
		} else if !reflect.DeepEqual(atrs[0].MinQueuedItems, result[0].MinQueuedItems) {
			t.Errorf("Expecting: %+v, received: %+v", atrs[0].MinQueuedItems, result[0].MinQueuedItems)
		} else if !reflect.DeepEqual(atrs[0].Executed, result[0].Executed) {
			t.Errorf("Expecting: %+v, received: %+v", atrs[0].Executed, result[0].Executed)
		} else if !reflect.DeepEqual(atrs[0].LastExecutionTime, result[0].LastExecutionTime) {
			t.Errorf("Expecting: %+v, received: %+v", atrs[0].LastExecutionTime, result[0].LastExecutionTime)
		}
		//Testing each field of balance
		if !reflect.DeepEqual(atrs[0].Balance.Uuid, result[0].Balance.Uuid) {
			t.Errorf("Expecting: %+v, received: %+v", atrs[0].Balance.Uuid, result[0].Balance.Uuid)
		} else if !reflect.DeepEqual(atrs[0].Balance.ID, result[0].Balance.ID) {
			t.Errorf("Expecting: %+v, received: %+v", atrs[0].Balance.ID, result[0].Balance.ID)
		} else if !reflect.DeepEqual(atrs[0].Balance.Type, result[0].Balance.Type) {
			t.Errorf("Expecting: %+v, received: %+v", atrs[0].Balance.Type, result[0].Balance.Type)
		} else if !reflect.DeepEqual(atrs[0].Balance.Value, result[0].Balance.Value) {
			t.Errorf("Expecting: %+v, received: %+v", atrs[0].Balance.Value, result[0].Balance.Value)
		} else if !reflect.DeepEqual(atrs[0].Balance.Directions, result[0].Balance.Directions) {
			t.Errorf("Expecting: %+v, received: %+v", atrs[0].Balance.Directions, result[0].Balance.Directions)
		} else if !reflect.DeepEqual(atrs[0].Balance.ExpirationDate, result[0].Balance.ExpirationDate) {
			t.Errorf("Expecting: %+v, received: %+v", atrs[0].Balance.ExpirationDate, result[0].Balance.ExpirationDate)
		} else if !reflect.DeepEqual(atrs[0].Balance.Weight, result[0].Balance.Weight) {
			t.Errorf("Expecting: %+v, received: %+v", atrs[0].Balance.Weight, result[0].Balance.Weight)
		} else if !reflect.DeepEqual(atrs[0].Balance.DestinationIDs, result[0].Balance.DestinationIDs) {
			t.Errorf("Expecting: %+v, received: %+v", atrs[0].Balance.DestinationIDs, result[0].Balance.DestinationIDs)
		} else if !reflect.DeepEqual(atrs[0].Balance.RatingSubject, result[0].Balance.RatingSubject) {
			t.Errorf("Expecting: %+v, received: %+v", atrs[0].Balance.RatingSubject, result[0].Balance.RatingSubject)
		} else if !reflect.DeepEqual(atrs[0].Balance.Categories, result[0].Balance.Categories) {
			t.Errorf("Expecting: %+v, received: %+v", atrs[0].Balance.Categories, result[0].Balance.Categories)
		} else if !reflect.DeepEqual(atrs[0].Balance.SharedGroups, result[0].Balance.SharedGroups) {
			t.Errorf("Expecting: %+v, received: %+v", atrs[0].Balance.SharedGroups, result[0].Balance.SharedGroups)
		} else if !reflect.DeepEqual(atrs[0].Balance.TimingIDs, result[0].Balance.TimingIDs) {
			t.Errorf("Expecting: %+v, received: %+v", atrs[0].Balance.TimingIDs, result[0].Balance.TimingIDs)
		} else if !reflect.DeepEqual(atrs[0].Balance.TimingIDs, result[0].Balance.TimingIDs) {
			t.Errorf("Expecting: %+v, received: %+v", atrs[0].Balance.Timings, result[0].Balance.Timings)
		} else if !reflect.DeepEqual(atrs[0].Balance.Disabled, result[0].Balance.Disabled) {
			t.Errorf("Expecting: %+v, received: %+v", atrs[0].Balance.Disabled, result[0].Balance.Disabled)
		} else if !reflect.DeepEqual(atrs[0].Balance.Factor, result[0].Balance.Factor) {
			t.Errorf("Expecting: %+v, received: %+v", atrs[0].Balance.Factor, result[0].Balance.Factor)
		} else if !reflect.DeepEqual(atrs[0].Balance.Blocker, result[0].Balance.Blocker) {
			t.Errorf("Expecting: %+v, received: %+v", atrs[0].Balance.Blocker, result[0].Balance.Blocker)
		}
	}
}

func testMigratorActions(t *testing.T) {
	v1act := &v1Actions{&v1Action{Id: "test", ActionType: "", BalanceType: "", Direction: "INBOUND", ExtraParameters: "", ExpirationString: "", Balance: &v1Balance{Timings: []*engine.RITiming{&engine.RITiming{Years: utils.Years{}, Months: utils.Months{}, MonthDays: utils.MonthDays{}, WeekDays: utils.WeekDays{}}}}}}
	act := &engine.Actions{&engine.Action{Id: "test", ActionType: "", ExtraParameters: "", ExpirationString: "", Weight: 0.00, Balance: &engine.BalanceFilter{Timings: []*engine.RITiming{&engine.RITiming{Years: utils.Years{}, Months: utils.Months{}, MonthDays: utils.MonthDays{}, WeekDays: utils.WeekDays{}}}}}}
	switch {
	case dbtype == utils.REDIS:
		err := mig.oldDataDB.setV1Actions(v1act)
		if err != nil {
			t.Error("Error when setting v1 Actions ", err.Error())
		}
		currentVersion := engine.Versions{utils.StatS: 2, utils.Thresholds: 2, utils.Accounts: 2, utils.Actions: 1, utils.ActionTriggers: 2, utils.ActionPlans: 2, utils.SharedGroups: 2}
		err = mig.dmOut.DataDB().SetVersions(currentVersion, false)
		if err != nil {
			t.Error("Error when setting version for stats ", err.Error())
		}
		err, _ = mig.Migrate([]string{utils.MetaActions})
		if err != nil {
			t.Error("Error when migrating Actions ", err.Error())
		}
		result, err := mig.dmOut.GetActions((*v1act)[0].Id, true, utils.NonTransactional)
		if err != nil {
			t.Error("Error when getting Actions ", err.Error())
		}
		if !reflect.DeepEqual(*act, result) {
			t.Errorf("Expecting: %+v, received: %+v", *act, result)
		}

	case dbtype == utils.MONGO:
		err := mig.oldDataDB.setV1Actions(v1act)
		if err != nil {
			t.Error("Error when setting v1 Actions ", err.Error())
		}
		currentVersion := engine.Versions{utils.StatS: 2, utils.Thresholds: 2, utils.Accounts: 2, utils.Actions: 1, utils.ActionTriggers: 2, utils.ActionPlans: 2, utils.SharedGroups: 2}
		err = mig.dmOut.DataDB().SetVersions(currentVersion, false)
		if err != nil {
			t.Error("Error when setting version for stats ", err.Error())
		}
		err, _ = mig.Migrate([]string{utils.MetaActions})
		if err != nil {
			t.Error("Error when migrating Actions ", err.Error())
		}
		result, err := mig.dmOut.GetActions((*v1act)[0].Id, true, utils.NonTransactional)
		if err != nil {
			t.Error("Error when getting Actions ", err.Error())
		}
		if !reflect.DeepEqual(*act, result) {
			t.Errorf("Expecting: %+v, received: %+v", *act, result)
		}
	case dbtype == Move:
		if err := mig.dmIN.SetActions((*v1act)[0].Id, *act, utils.NonTransactional); err != nil {
			t.Error("Error when setting ActionPlan ", err.Error())
		}
		currentVersion := engine.CurrentDataDBVersions()
		err := mig.dmOut.DataDB().SetVersions(currentVersion, false)
		if err != nil {
			t.Error("Error when setting version for stats ", err.Error())
		}
		err, _ = mig.Migrate([]string{utils.MetaActions})
		if err != nil {
			t.Error("Error when migrating ActionPlans ", err.Error())
		}
		result, err := mig.dmOut.GetActions((*v1act)[0].Id, true, utils.NonTransactional)
		if err != nil {
			t.Error("Error when getting ActionPlan ", err.Error())
		}
		if !reflect.DeepEqual(*act, result) {
			t.Errorf("Expecting: %+v, received: %+v", *act, result)
		}
	}
}

func testMigratorSharedGroups(t *testing.T) {
	v1sqp := &v1SharedGroup{
		Id: "Test",
		AccountParameters: map[string]*engine.SharingParameters{
			"test": &engine.SharingParameters{Strategy: "*highest"},
		},
		MemberIds: []string{"1", "2", "3"},
	}
	sqp := &engine.SharedGroup{
		Id: "Test",
		AccountParameters: map[string]*engine.SharingParameters{
			"test": &engine.SharingParameters{Strategy: "*highest"},
		},
		MemberIds: utils.NewStringMap("1", "2", "3"),
	}
	switch {
	case dbtype == utils.REDIS:
		err := mig.oldDataDB.setV1SharedGroup(v1sqp)
		if err != nil {
			t.Error("Error when setting v1 SharedGroup ", err.Error())
		}
		currentVersion := engine.Versions{utils.StatS: 2, utils.Thresholds: 2, utils.Accounts: 2, utils.Actions: 2, utils.ActionTriggers: 2, utils.ActionPlans: 2, utils.SharedGroups: 1}
		err = mig.dmOut.DataDB().SetVersions(currentVersion, false)
		if err != nil {
			t.Error("Error when setting version for stats ", err.Error())
		}
		err, _ = mig.Migrate([]string{utils.MetaSharedGroups})
		if err != nil {
			t.Error("Error when migrating SharedGroup ", err.Error())
		}
		result, err := mig.dmOut.GetSharedGroup(v1sqp.Id, true, utils.NonTransactional)
		if err != nil {
			t.Error("Error when getting SharedGroup ", err.Error())
		}
		if !reflect.DeepEqual(sqp, result) {
			t.Errorf("Expecting: %+v, received: %+v", sqp, result)
		}
	case dbtype == utils.MONGO:
		err := mig.oldDataDB.setV1SharedGroup(v1sqp)
		if err != nil {
			t.Error("Error when setting v1 SharedGroup ", err.Error())
		}
		currentVersion := engine.Versions{utils.StatS: 2, utils.Thresholds: 2, utils.Accounts: 2, utils.Actions: 2, utils.ActionTriggers: 2, utils.ActionPlans: 2, utils.SharedGroups: 1}
		err = mig.dmOut.DataDB().SetVersions(currentVersion, false)
		if err != nil {
			t.Error("Error when setting version for stats ", err.Error())
		}
		err, _ = mig.Migrate([]string{utils.MetaSharedGroups})
		if err != nil {
			t.Error("Error when migrating SharedGroup ", err.Error())
		}
		result, err := mig.dmOut.GetSharedGroup(v1sqp.Id, true, utils.NonTransactional)
		if err != nil {
			t.Error("Error when getting SharedGroup ", err.Error())
		}
		if !reflect.DeepEqual(sqp, result) {
			t.Errorf("Expecting: %+v, received: %+v", sqp, result)
		}
	case dbtype == Move:
		if err := mig.dmIN.SetSharedGroup(sqp, utils.NonTransactional); err != nil {
			t.Error("Error when setting SharedGroup ", err.Error())
		}
		currentVersion := engine.CurrentDataDBVersions()
		err := mig.dmOut.DataDB().SetVersions(currentVersion, false)
		if err != nil {
			t.Error("Error when setting version for stats ", err.Error())
		}
		err, _ = mig.Migrate([]string{utils.MetaSharedGroups})
		if err != nil {
			t.Error("Error when migrating SharedGroup ", err.Error())
		}
		result, err := mig.dmOut.GetSharedGroup(sqp.Id, true, utils.NonTransactional)
		if err != nil {
			t.Error("Error when getting SharedGroup ", err.Error())
		}
		if !reflect.DeepEqual(sqp, result) {
			t.Errorf("Expecting: %+v, received: %+v", sqp, result)
		}
	}
}

func testMigratorStats(t *testing.T) {
	tim := time.Date(2012, time.February, 27, 23, 59, 59, 0, time.UTC).Local()
	var filters []*engine.RequestFilter
	v1Sts := &v1Stat{
		Id:              "test",                         // Config id, unique per config instance
		QueueLength:     10,                             // Number of items in the stats buffer
		TimeWindow:      time.Duration(1) * time.Second, // Will only keep the CDRs who's call setup time is not older than time.Now()-TimeWindow
		SaveInterval:    time.Duration(1) * time.Second,
		Metrics:         []string{"ASR", "ACD", "ACC"},
		SetupInterval:   []time.Time{time.Now()},
		TOR:             []string{},
		CdrHost:         []string{},
		CdrSource:       []string{},
		ReqType:         []string{},
		Direction:       []string{},
		Tenant:          []string{},
		Category:        []string{},
		Account:         []string{},
		Subject:         []string{},
		DestinationIds:  []string{},
		UsageInterval:   []time.Duration{1 * time.Second},
		PddInterval:     []time.Duration{1 * time.Second},
		Supplier:        []string{},
		DisconnectCause: []string{},
		MediationRunIds: []string{},
		RatedAccount:    []string{},
		RatedSubject:    []string{},
		CostInterval:    []float64{},
		Triggers: engine.ActionTriggers{
			&engine.ActionTrigger{
				ID: "Test",
				Balance: &engine.BalanceFilter{
					ID:             utils.StringPointer("TESTB"),
					Timings:        []*engine.RITiming{},
					ExpirationDate: utils.TimePointer(tim),
					Type:           utils.StringPointer(utils.MONETARY),
					Directions:     utils.StringMapPointer(utils.NewStringMap(utils.OUT)),
				},
				ExpirationDate:    tim,
				LastExecutionTime: tim,
				ActivationDate:    tim,
				ThresholdType:     utils.TRIGGER_MAX_BALANCE,
				ThresholdValue:    2,
				ActionsID:         "TEST_ACTIONS",
				Executed:          true,
			},
		},
	}

	x, _ := engine.NewRequestFilter(engine.MetaGreaterOrEqual, "SetupInterval", []string{v1Sts.SetupInterval[0].String()})
	filters = append(filters, x)
	x, _ = engine.NewRequestFilter(engine.MetaGreaterOrEqual, "UsageInterval", []string{v1Sts.UsageInterval[0].String()})
	filters = append(filters, x)
	x, _ = engine.NewRequestFilter(engine.MetaGreaterOrEqual, "PddInterval", []string{v1Sts.PddInterval[0].String()})
	filters = append(filters, x)

	filter := &engine.Filter{Tenant: config.CgrConfig().DefaultTenant, ID: v1Sts.Id, RequestFilters: filters}

	sqp := &engine.StatQueueProfile{
		Tenant:      "cgrates.org",
		ID:          "test",
		FilterIDs:   []string{v1Sts.Id},
		QueueLength: 10,
		TTL:         time.Duration(0) * time.Second,
		Metrics:     []string{"*asr", "*acd", "*acc"},
		Thresholds:  []string{"Test"},
		Blocker:     false,
		Stored:      true,
		Weight:      float64(0),
		MinItems:    0,
	}
	sq := &engine.StatQueue{Tenant: config.CgrConfig().DefaultTenant,
		ID:        v1Sts.Id,
		SQMetrics: make(map[string]engine.StatMetric),
	}
	for _, metricID := range sqp.Metrics {
		if metric, err := engine.NewStatMetric(metricID, 0); err != nil {
			t.Error("Error when creating newstatMETRIc ", err.Error())
		} else {
			sq.SQMetrics[metricID] = metric
		}
	}
	switch {
	case dbtype == utils.REDIS:
		err := mig.oldDataDB.setV1Stats(v1Sts)
		if err != nil {
			t.Error("Error when setting v1Stat ", err.Error())
		}
		currentVersion := engine.Versions{utils.StatS: 1, utils.Thresholds: 2, utils.Accounts: 2, utils.Actions: 2, utils.ActionTriggers: 2, utils.ActionPlans: 2, utils.SharedGroups: 2}
		err = mig.dmOut.DataDB().SetVersions(currentVersion, false)
		if err != nil {
			t.Error("Error when setting version for stats ", err.Error())
		}
		err, _ = mig.Migrate([]string{utils.MetaStats})
		if err != nil {
			t.Error("Error when migrating Stats ", err.Error())
		}
		result, err := mig.dmOut.GetStatQueueProfile("cgrates.org", v1Sts.Id, true, utils.NonTransactional)
		if err != nil {
			t.Error("Error when getting Stats ", err.Error())
		}
		if !reflect.DeepEqual(sqp.Tenant, result.Tenant) {
			t.Errorf("Expecting: %+v, received: %+v", sqp.Tenant, result.Tenant)
		}
		if !reflect.DeepEqual(sqp.ID, result.ID) {
			t.Errorf("Expecting: %+v, received: %+v", sqp.ID, result.ID)
		}
		if !reflect.DeepEqual(sqp.FilterIDs, result.FilterIDs) {
			t.Errorf("Expecting: %+v, received: %+v", sqp.FilterIDs, result.FilterIDs)
		}
		if !reflect.DeepEqual(sqp.QueueLength, result.QueueLength) {
			t.Errorf("Expecting: %+v, received: %+v", sqp.QueueLength, result.QueueLength)
		}
		if !reflect.DeepEqual(sqp.TTL, result.TTL) {
			t.Errorf("Expecting: %+v, received: %+v", sqp.TTL, result.TTL)
		}
		if !reflect.DeepEqual(sqp.Metrics, result.Metrics) {
			t.Errorf("Expecting: %+v, received: %+v", sqp.Metrics, result.Metrics)
		}
		if !reflect.DeepEqual(sqp.Thresholds, result.Thresholds) {
			t.Errorf("Expecting: %+v, received: %+v", sqp.Thresholds, result.Thresholds)
		}
		if !reflect.DeepEqual(sqp.Blocker, result.Blocker) {
			t.Errorf("Expecting: %+v, received: %+v", sqp.Blocker, result.Blocker)
		}
		if !reflect.DeepEqual(sqp.Stored, result.Stored) {
			t.Errorf("Expecting: %+v, received: %+v", sqp.Stored, result.Stored)
		}
		if !reflect.DeepEqual(sqp.Weight, result.Weight) {
			t.Errorf("Expecting: %+v, received: %+v", sqp.Weight, result.Weight)
		}
		if !reflect.DeepEqual(sqp, result) {
			t.Errorf("Expecting: %+v, received: %+v", sqp, result)
		}
		result1, err := mig.dmOut.GetFilter("cgrates.org", v1Sts.Id, true, utils.NonTransactional)
		if err != nil {
			t.Error("Error when getting Stats ", err.Error())
		}
		if !reflect.DeepEqual(filter.ActivationInterval, result1.ActivationInterval) {
			t.Errorf("Expecting: %+v, received: %+v", filter.ActivationInterval, result1.ActivationInterval)
		}
		if !reflect.DeepEqual(filter.Tenant, result1.Tenant) {
			t.Errorf("Expecting: %+v, received: %+v", filter.Tenant, result1.Tenant)
		}

		result2, err := mig.dmOut.GetStatQueue("cgrates.org", sq.ID, true, utils.NonTransactional)
		if err != nil {
			t.Error("Error when getting Stats ", err.Error())
		}
		if !reflect.DeepEqual(sq.ID, result2.ID) {
			t.Errorf("Expecting: %+v, received: %+v", sq.ID, result2.ID)
		}
	case dbtype == utils.MONGO:
		err := mig.oldDataDB.setV1Stats(v1Sts)
		if err != nil {
			t.Error("Error when setting v1Stat ", err.Error())
		}
		currentVersion := engine.Versions{utils.StatS: 1, utils.Thresholds: 2, utils.Accounts: 2, utils.Actions: 2, utils.ActionTriggers: 2, utils.ActionPlans: 2, utils.SharedGroups: 2}
		err = mig.dmOut.DataDB().SetVersions(currentVersion, false)
		if err != nil {
			t.Error("Error when setting version for stats ", err.Error())
		}
		err, _ = mig.Migrate([]string{utils.MetaStats})
		if err != nil {
			t.Error("Error when migrating Stats ", err.Error())
		}
		result, err := mig.dmOut.GetStatQueueProfile("cgrates.org", v1Sts.Id, true, utils.NonTransactional)
		if err != nil {
			t.Error("Error when getting Stats ", err.Error())
		}
		if !reflect.DeepEqual(sqp.Tenant, result.Tenant) {
			t.Errorf("Expecting: %+v, received: %+v", sqp.Tenant, result.Tenant)
		}
		if !reflect.DeepEqual(sqp.ID, result.ID) {
			t.Errorf("Expecting: %+v, received: %+v", sqp.ID, result.ID)
		}
		if !reflect.DeepEqual(sqp.FilterIDs, result.FilterIDs) {
			t.Errorf("Expecting: %+v, received: %+v", sqp.FilterIDs, result.FilterIDs)
		}
		if !reflect.DeepEqual(sqp.QueueLength, result.QueueLength) {
			t.Errorf("Expecting: %+v, received: %+v", sqp.QueueLength, result.QueueLength)
		}
		if !reflect.DeepEqual(sqp.TTL, result.TTL) {
			t.Errorf("Expecting: %+v, received: %+v", sqp.TTL, result.TTL)
		}
		if !reflect.DeepEqual(sqp.Metrics, result.Metrics) {
			t.Errorf("Expecting: %+v, received: %+v", sqp.Metrics, result.Metrics)
		}
		if !reflect.DeepEqual(sqp.Thresholds, result.Thresholds) {
			t.Errorf("Expecting: %+v, received: %+v", sqp.Thresholds, result.Thresholds)
		}
		if !reflect.DeepEqual(sqp.Blocker, result.Blocker) {
			t.Errorf("Expecting: %+v, received: %+v", sqp.Blocker, result.Blocker)
		}
		if !reflect.DeepEqual(sqp.Stored, result.Stored) {
			t.Errorf("Expecting: %+v, received: %+v", sqp.Stored, result.Stored)
		}
		if !reflect.DeepEqual(sqp.Weight, result.Weight) {
			t.Errorf("Expecting: %+v, received: %+v", sqp.Weight, result.Weight)
		}
		if !reflect.DeepEqual(sqp, result) {
			t.Errorf("Expecting: %+v, received: %+v", sqp, result)
		}
		result1, err := mig.dmOut.GetFilter("cgrates.org", v1Sts.Id, true, utils.NonTransactional)
		if err != nil {
			t.Error("Error when getting Stats ", err.Error())
		}
		if !reflect.DeepEqual(filter.ActivationInterval, result1.ActivationInterval) {
			t.Errorf("Expecting: %+v, received: %+v", filter.ActivationInterval, result1.ActivationInterval)
		}
		if !reflect.DeepEqual(filter.Tenant, result1.Tenant) {
			t.Errorf("Expecting: %+v, received: %+v", filter.Tenant, result1.Tenant)
		}
		result2, err := mig.dmOut.GetStatQueue("cgrates.org", sq.ID, true, utils.NonTransactional)
		if err != nil {
			t.Error("Error when getting Stats ", err.Error())
		}
		if !reflect.DeepEqual(sq.ID, result2.ID) {
			t.Errorf("Expecting: %+v, received: %+v", sq.ID, result2.ID)
		}
	case dbtype == Move:
		if err := mig.dmIN.SetStatQueueProfile(sqp); err != nil {
			t.Error("Error when setting Stats ", err.Error())
		}
		if err := mig.dmIN.SetStatQueue(sq); err != nil {
			t.Error("Error when setting Stats ", err.Error())
		}
		currentVersion := engine.CurrentDataDBVersions()
		err := mig.dmOut.DataDB().SetVersions(currentVersion, false)
		if err != nil {
			t.Error("Error when setting version for stats ", err.Error())
		}
		err, _ = mig.Migrate([]string{utils.MetaStats})
		if err != nil {
			t.Error("Error when migrating Stats ", err.Error())
		}
		result, err := mig.dmOut.GetStatQueueProfile(sqp.Tenant, sqp.ID, true, utils.NonTransactional)
		if err != nil {
			t.Error("Error when getting Stats ", err.Error())
		}
		result1, err := mig.dmOut.GetStatQueue(sq.Tenant, sq.ID, true, utils.NonTransactional)
		if err != nil {
			t.Error("Error when getting Stats ", err.Error())
		}
		if !reflect.DeepEqual(sqp, result) {
			t.Errorf("Expecting: %+v, received: %+v", sqp, result)
		}
		if !reflect.DeepEqual(sq.ID, result1.ID) {
			t.Errorf("Expecting: %+v, received: %+v", sq.ID, result1.ID)
		}
	}
}

func testMigratorTPRatingProfile(t *testing.T) {
	tpRatingProfile := &utils.TPRatingProfile{
		TPid:      "TPRProf1",
		LoadId:    "RPrf",
		Direction: "*out",
		Tenant:    "Tenant1",
		Category:  "Category",
		Subject:   "Subject",
		RatingPlanActivations: []*utils.TPRatingActivation{
			&utils.TPRatingActivation{
				ActivationTime:   "2014-07-29T15:00:00Z",
				RatingPlanId:     "PlanOne",
				FallbackSubjects: "FallBack",
				CdrStatQueueIds:  "RandomId",
			},
			&utils.TPRatingActivation{
				ActivationTime:   "2015-07-29T10:00:00Z",
				RatingPlanId:     "PlanTwo",
				FallbackSubjects: "FallOut",
				CdrStatQueueIds:  "RandomIdTwo",
			},
		},
	}
	switch dbtype {
	case Move:
		if err := mig.InStorDB().SetTPRatingProfiles([]*utils.TPRatingProfile{tpRatingProfile}); err != nil {
			t.Error("Error when setting Stats ", err.Error())
		}
		currentVersion := engine.CurrentDataDBVersions()
		err := mig.dmOut.DataDB().SetVersions(currentVersion, false)
		if err != nil {
			t.Error("Error when setting version for stats ", err.Error())
		}
		err, _ = mig.Migrate([]string{utils.MetaTpRatingProfiles})
		if err != nil {
			t.Error("Error when migrating Stats ", err.Error())
		}
		result, err := mig.OutStorDB().GetTPRatingProfiles(tpRatingProfile)
		if err != nil {
			t.Error("Error when getting Stats ", err.Error())
		}
		if !reflect.DeepEqual(tpRatingProfile, result[0]) {
			t.Errorf("Expecting: %+v, received: %+v", tpRatingProfile, result[0])
		}
	}
}
