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
	"fmt"
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
	isPostgres      bool
	path_in         string
	path_out        string
	cfg_in          *config.CGRConfig
	cfg_out         *config.CGRConfig
	Move            = "move"
	action          string
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
	testMigratorSessionsCosts,
	testFlush,
	testMigratorAlias,
	//FIXME testMigratorReverseAlias,
	testMigratorCdrStats,
	testMigratorDerivedChargers,
	testMigratorDestinations,
	testMigratorReverseDestinations,
	testMigratorLCR,
	testMigratorRatingPlan,
	testMigratorRatingProfile,
	testMigratorRQF,
	testMigratorResource,
	testMigratorSubscribers,
	testMigratorTimings,
	testMigratorThreshold,
	testMigratorAttributeProfile,
	//TPS
	testMigratorTPRatingProfile,
	testMigratorTPSuppliers,
	testMigratorTPActions,
	testMigratorTPAccountActions,
	testMigratorTpActionTriggers,
	testMigratorTpActionPlans,
	testMigratorTpUsers,
	testMigratorTpTimings,
	testMigratorTpThreshold,
	testMigratorTpStats,
	testMigratorTpSharedGroups,
	testMigratorTpResources,
	testMigratorTpRatingProfiles,
	testMigratorTpRatingPlans,
	testMigratorTpRates,
	testMigratorTpFilter,
	testMigratorTpDestination,
	testMigratorTpDestinationRate,
	testMigratorTpDerivedChargers,
	testMigratorTpCdrStats,
	testMigratorTpAliases,
	testFlush,
}

func TestMigratorITPostgresConnect(t *testing.T) {
	path_in := path.Join(*dataDir, "conf", "samples", "tutpostgres")
	cfg_in, err := config.NewCGRConfigFromFolder(path_in)
	if err != nil {
		t.Fatal(err)
	}
	dataDB, err := engine.ConfigureDataStorage(cfg_in.DataDbType, cfg_in.DataDbHost, cfg_in.DataDbPort, cfg_in.DataDbName, cfg_in.DataDbUser,
		cfg_in.DataDbPass, cfg_in.DBDataEncoding, config.CgrConfig().CacheCfg(), *loadHistorySize)
	if err != nil {
		log.Fatal(err)
	}
	dataDB2, err := engine.ConfigureDataStorage(cfg_in.DataDbType, cfg_in.DataDbHost, cfg_in.DataDbPort, cfg_in.DataDbName, cfg_in.DataDbUser,
		cfg_in.DataDbPass, cfg_in.DBDataEncoding, config.CgrConfig().CacheCfg(), *loadHistorySize)
	if err != nil {
		log.Fatal(err)
	}
	oldDataDB, err := ConfigureV1DataStorage(cfg_in.DataDbType, cfg_in.DataDbHost, cfg_in.DataDbPort, cfg_in.DataDbName, cfg_in.DataDbUser,
		cfg_in.DataDbPass, cfg_in.DBDataEncoding)
	if err != nil {
		log.Fatal(err)
	}
	storDB, err := engine.ConfigureStorStorage(cfg_in.StorDBType, cfg_in.StorDBHost, cfg_in.StorDBPort, cfg_in.StorDBName,
		cfg_in.StorDBUser, cfg_in.StorDBPass, cfg_in.DBDataEncoding,
		config.CgrConfig().StorDBMaxOpenConns, config.CgrConfig().StorDBMaxIdleConns, config.CgrConfig().StorDBConnMaxLifetime, config.CgrConfig().StorDBCDRSIndexes)
	if err != nil {
		log.Fatal(err)
	}
	oldstorDB, err := engine.ConfigureStorStorage(cfg_in.StorDBType, cfg_in.StorDBHost, cfg_in.StorDBPort, cfg_in.StorDBName,
		cfg_in.StorDBUser, cfg_in.StorDBPass, cfg_in.DBDataEncoding,
		config.CgrConfig().StorDBMaxOpenConns, config.CgrConfig().StorDBMaxIdleConns, config.CgrConfig().StorDBConnMaxLifetime, config.CgrConfig().StorDBCDRSIndexes)
	if err != nil {
		log.Fatal(err)
	}
	mig, err = NewMigrator(dataDB, dataDB2, cfg_in.DataDbType, cfg_in.DBDataEncoding, storDB, cfg_in.StorDBType, oldDataDB,
		cfg_in.DataDbType, cfg_in.DBDataEncoding, oldstorDB, cfg_in.StorDBType, false, true, true, false, false)
	if err != nil {
		log.Fatal(err)
	}
}

func TestMigratorITPostgres(t *testing.T) {
	action = utils.REDIS
	isPostgres = true
	for _, stest := range sTestsITMigrator {
		t.Run("TestITMigratorOnPostgres", stest)
	}
}

func TestMigratorITRedisConnect(t *testing.T) {
	path_in := path.Join(*dataDir, "conf", "samples", "tutmysql")
	cfg_in, err := config.NewCGRConfigFromFolder(path_in)
	if err != nil {
		t.Fatal(err)
	}
	dataDB, err := engine.ConfigureDataStorage(cfg_in.DataDbType, cfg_in.DataDbHost, cfg_in.DataDbPort, cfg_in.DataDbName,
		cfg_in.DataDbUser, cfg_in.DataDbPass, cfg_in.DBDataEncoding, config.CgrConfig().CacheCfg(), *loadHistorySize)
	if err != nil {
		log.Fatal(err)
	}
	dataDB2, err := engine.ConfigureDataStorage(cfg_in.DataDbType, cfg_in.DataDbHost, cfg_in.DataDbPort, cfg_in.DataDbName,
		cfg_in.DataDbUser, cfg_in.DataDbPass, cfg_in.DBDataEncoding, config.CgrConfig().CacheCfg(), *loadHistorySize)
	if err != nil {
		log.Fatal(err)
	}
	oldDataDB, err := ConfigureV1DataStorage(cfg_in.DataDbType, cfg_in.DataDbHost, cfg_in.DataDbPort, cfg_in.DataDbName,
		cfg_in.DataDbUser, cfg_in.DataDbPass, cfg_in.DBDataEncoding)
	if err != nil {
		log.Fatal(err)
	}
	storDB, err := engine.ConfigureStorStorage(cfg_in.StorDBType, cfg_in.StorDBHost, cfg_in.StorDBPort, cfg_in.StorDBName,
		cfg_in.StorDBUser, cfg_in.StorDBPass, cfg_in.DBDataEncoding,
		config.CgrConfig().StorDBMaxOpenConns, config.CgrConfig().StorDBMaxIdleConns, config.CgrConfig().StorDBConnMaxLifetime, config.CgrConfig().StorDBCDRSIndexes)
	if err != nil {
		log.Fatal(err)
	}
	oldstorDB, err := engine.ConfigureStorStorage(cfg_in.StorDBType, cfg_in.StorDBHost, cfg_in.StorDBPort, cfg_in.StorDBName,
		cfg_in.StorDBUser, cfg_in.StorDBPass, cfg_in.DBDataEncoding,
		config.CgrConfig().StorDBMaxOpenConns, config.CgrConfig().StorDBMaxIdleConns, config.CgrConfig().StorDBConnMaxLifetime, config.CgrConfig().StorDBCDRSIndexes)
	if err != nil {
		log.Fatal(err)
	}
	mig, err = NewMigrator(dataDB, dataDB2, cfg_in.DataDbType, cfg_in.DBDataEncoding, storDB, cfg_in.StorDBType, oldDataDB,
		cfg_in.DataDbType, cfg_in.DBDataEncoding, oldstorDB, cfg_in.StorDBType, false, true, true, false, false)
	if err != nil {
		log.Fatal(err)
	}
}

func TestMigratorITRedis(t *testing.T) {
	action = utils.REDIS
	isPostgres = false
	for _, stest := range sTestsITMigrator {
		t.Run("TestITMigratorOnRedis", stest)
	}
}

func TestMigratorITMongoConnect(t *testing.T) {
	path_in := path.Join(*dataDir, "conf", "samples", "tutmongo")
	cfg_in, err := config.NewCGRConfigFromFolder(path_in)
	if err != nil {
		t.Fatal(err)
	}
	dataDB, err := engine.ConfigureDataStorage(cfg_in.DataDbType, cfg_in.DataDbHost, cfg_in.DataDbPort, cfg_in.DataDbName,
		cfg_in.DataDbUser, cfg_in.DataDbPass, cfg_in.DBDataEncoding, config.CgrConfig().CacheCfg(), *loadHistorySize)
	if err != nil {
		log.Fatal(err)
	}
	dataDB2, err := engine.ConfigureDataStorage(cfg_in.DataDbType, cfg_in.DataDbHost, cfg_in.DataDbPort, cfg_in.DataDbName,
		cfg_in.DataDbUser, cfg_in.DataDbPass, cfg_in.DBDataEncoding, config.CgrConfig().CacheCfg(), *loadHistorySize)
	if err != nil {
		log.Fatal(err)
	}
	oldDataDB, err := ConfigureV1DataStorage(cfg_in.DataDbType, cfg_in.DataDbHost, cfg_in.DataDbPort, cfg_in.DataDbName,
		cfg_in.DataDbUser, cfg_in.DataDbPass, cfg_in.DBDataEncoding)
	if err != nil {
		log.Fatal(err)
	}
	storDB, err := engine.ConfigureStorStorage(cfg_in.StorDBType, cfg_in.StorDBHost, cfg_in.StorDBPort, cfg_in.StorDBName,
		cfg_in.StorDBUser, cfg_in.StorDBPass, cfg_in.DBDataEncoding,
		config.CgrConfig().StorDBMaxOpenConns, config.CgrConfig().StorDBMaxIdleConns, config.CgrConfig().StorDBConnMaxLifetime, config.CgrConfig().StorDBCDRSIndexes)
	if err != nil {
		log.Fatal(err)
	}
	oldstorDB, err := engine.ConfigureStorStorage(cfg_in.StorDBType, cfg_in.StorDBHost, cfg_in.StorDBPort, cfg_in.StorDBName,
		cfg_in.StorDBUser, cfg_in.StorDBPass, cfg_in.DBDataEncoding,
		config.CgrConfig().StorDBMaxOpenConns, config.CgrConfig().StorDBMaxIdleConns, config.CgrConfig().StorDBConnMaxLifetime, config.CgrConfig().StorDBCDRSIndexes)
	if err != nil {
		log.Fatal(err)
	}
	mig, err = NewMigrator(dataDB, dataDB2, cfg_in.DataDbType, cfg_in.DBDataEncoding, storDB, cfg_in.StorDBType, oldDataDB,
		cfg_in.DataDbType, cfg_in.DBDataEncoding, oldstorDB, cfg_in.StorDBType, false, true, true, false, false)
	if err != nil {
		log.Fatal(err)
	}
}

func TestMigratorITMongo(t *testing.T) {
	action = utils.MONGO
	for _, stest := range sTestsITMigrator {
		t.Run("TestITMigratorOnMongo", stest)
	}
}

func TestMigratorITMoveConnect(t *testing.T) {
	path_in := path.Join(*dataDir, "conf", "samples", "tutmongo")
	cfg_in, err := config.NewCGRConfigFromFolder(path_in)
	if err != nil {
		t.Fatal(err)
	}
	path_out := path.Join(*dataDir, "conf", "samples", "tutmysql")
	cfg_out, err := config.NewCGRConfigFromFolder(path_out)
	if err != nil {
		t.Fatal(err)
	}
	dataDB, err := engine.ConfigureDataStorage(cfg_in.DataDbType, cfg_in.DataDbHost, cfg_in.DataDbPort, cfg_in.DataDbName,
		cfg_in.DataDbUser, cfg_in.DataDbPass, cfg_in.DBDataEncoding, config.CgrConfig().CacheCfg(), *loadHistorySize)
	if err != nil {
		log.Fatal(err)
	}
	dataDB2, err := engine.ConfigureDataStorage(cfg_out.DataDbType, cfg_out.DataDbHost, cfg_out.DataDbPort, cfg_out.DataDbName,
		cfg_out.DataDbUser, cfg_out.DataDbPass, cfg_out.DBDataEncoding, config.CgrConfig().CacheCfg(), *loadHistorySize)
	if err != nil {
		log.Fatal(err)
	}
	oldDataDB, err := ConfigureV1DataStorage(cfg_out.DataDbType, cfg_out.DataDbHost, cfg_out.DataDbPort, cfg_out.DataDbName,
		cfg_out.DataDbUser, cfg_out.DataDbPass, cfg_out.DBDataEncoding)
	if err != nil {
		log.Fatal(err)
	}
	storDB, err := engine.ConfigureStorStorage(cfg_in.StorDBType, cfg_in.StorDBHost, cfg_in.StorDBPort, cfg_in.StorDBName,
		cfg_in.StorDBUser, cfg_in.StorDBPass, cfg_in.DBDataEncoding,
		config.CgrConfig().StorDBMaxOpenConns, config.CgrConfig().StorDBMaxIdleConns, config.CgrConfig().StorDBConnMaxLifetime, config.CgrConfig().StorDBCDRSIndexes)
	if err != nil {
		log.Fatal(err)
	}
	oldstorDB, err := engine.ConfigureStorStorage(cfg_out.StorDBType, cfg_out.StorDBHost, cfg_out.StorDBPort, cfg_out.StorDBName,
		cfg_out.StorDBUser, cfg_out.StorDBPass, cfg_out.DBDataEncoding,
		config.CgrConfig().StorDBMaxOpenConns, config.CgrConfig().StorDBMaxIdleConns, config.CgrConfig().StorDBConnMaxLifetime, config.CgrConfig().StorDBCDRSIndexes)
	if err != nil {
		log.Fatal(err)
	}
	mig, err = NewMigrator(dataDB2, dataDB, cfg_in.DataDbType, cfg_in.DBDataEncoding, storDB, cfg_in.StorDBType, oldDataDB,
		cfg_in.DataDbType, cfg_in.DBDataEncoding, oldstorDB, cfg_in.StorDBType, false, false, false, false, false)
	if err != nil {
		log.Fatal(err)
	}
}

func TestMigratorITMove(t *testing.T) {
	action = Move
	for _, stest := range sTestsITMigrator {
		t.Run("TestITMigratorOnMongo", stest)
	}
}

func testFlush(t *testing.T) {
	mig.dmOut.DataDB().Flush("")
	if err := engine.SetDBVersions(mig.dmOut.DataDB()); err != nil {
		t.Error("Error  ", err.Error())
	}
	if path_out != "" {
		if err := mig.InStorDB().Flush(path.Join(cfg_in.DataFolderPath, "storage", cfg_in.StorDBType)); err != nil {
			t.Error(err)
		}
	}
	if path_out != "" {
		if err := mig.OutStorDB().Flush(path.Join(cfg_out.DataFolderPath, "storage", cfg_out.StorDBType)); err != nil {
			t.Error(err)
		}
	}
}

func testMigratorAccounts(t *testing.T) {
	v1d := &v1Balance{Value: 100000, Weight: 10, DestinationIds: "NAT", ExpirationDate: time.Date(2012, 1, 1, 0, 0, 0, 0, time.UTC), Timings: []*engine.RITiming{&engine.RITiming{Years: utils.Years{}, Months: utils.Months{}, MonthDays: utils.MonthDays{}, WeekDays: utils.WeekDays{}}}}
	v1b := &v1Balance{Value: 100000, Weight: 10, DestinationIds: "NAT", ExpirationDate: time.Date(2012, 1, 1, 0, 0, 0, 0, time.UTC), Timings: []*engine.RITiming{&engine.RITiming{Years: utils.Years{}, Months: utils.Months{}, MonthDays: utils.MonthDays{}, WeekDays: utils.WeekDays{}}}}
	v1Acc := &v1Account{Id: "*OUT:CUSTOMER_1:rif", BalanceMap: map[string]v1BalanceChain{utils.DATA: v1BalanceChain{v1d}, utils.VOICE: v1BalanceChain{v1b}, utils.MONETARY: v1BalanceChain{&v1Balance{Value: 21, ExpirationDate: time.Date(2012, 1, 1, 0, 0, 0, 0, time.UTC), Timings: []*engine.RITiming{&engine.RITiming{Years: utils.Years{}, Months: utils.Months{}, MonthDays: utils.MonthDays{}, WeekDays: utils.WeekDays{}}}}}}}

	v2d := &engine.Balance{Uuid: "", ID: "", Value: 100000, Directions: utils.StringMap{"*OUT": true}, ExpirationDate: time.Date(2012, 1, 1, 0, 0, 0, 0, time.UTC), Weight: 10, DestinationIDs: utils.StringMap{"NAT": true},
		RatingSubject: "", Categories: utils.NewStringMap(), SharedGroups: utils.NewStringMap(), Timings: []*engine.RITiming{&engine.RITiming{Years: utils.Years{}, Months: utils.Months{}, MonthDays: utils.MonthDays{}, WeekDays: utils.WeekDays{}}}, TimingIDs: utils.NewStringMap(""), Factor: engine.ValueFactor{}}
	v2b := &engine.Balance{Uuid: "", ID: "", Value: 0.0001, Directions: utils.StringMap{"*OUT": true}, ExpirationDate: time.Date(2012, 1, 1, 0, 0, 0, 0, time.UTC), Weight: 10, DestinationIDs: utils.StringMap{"NAT": true},
		RatingSubject: "", Categories: utils.NewStringMap(), SharedGroups: utils.NewStringMap(), Timings: []*engine.RITiming{&engine.RITiming{Years: utils.Years{}, Months: utils.Months{}, MonthDays: utils.MonthDays{}, WeekDays: utils.WeekDays{}}}, TimingIDs: utils.NewStringMap(""), Factor: engine.ValueFactor{}}
	m2 := &engine.Balance{Uuid: "", ID: "", Value: 21, Directions: utils.StringMap{"*OUT": true}, ExpirationDate: time.Date(2012, 1, 1, 0, 0, 0, 0, time.UTC), DestinationIDs: utils.NewStringMap(""), RatingSubject: "",
		Categories: utils.NewStringMap(), SharedGroups: utils.NewStringMap(), Timings: []*engine.RITiming{&engine.RITiming{Years: utils.Years{}, Months: utils.Months{}, MonthDays: utils.MonthDays{}, WeekDays: utils.WeekDays{}}}, TimingIDs: utils.NewStringMap(""), Factor: engine.ValueFactor{}}
	testAccount := &engine.Account{ID: "CUSTOMER_1:rif", BalanceMap: map[string]engine.Balances{utils.DATA: engine.Balances{v2d}, utils.VOICE: engine.Balances{v2b}, utils.MONETARY: engine.Balances{m2}}, UnitCounters: engine.UnitCounters{}, ActionTriggers: engine.ActionTriggers{}}
	switch {
	case action == utils.REDIS:
		err := mig.oldDataDB.setV1Account(v1Acc)
		if err != nil {
			t.Error("Error when setting v1 Accounts ", err.Error())
		}
		currentVersion := engine.Versions{utils.StatS: 2, utils.Thresholds: 2, utils.Accounts: 1, utils.Actions: 2, utils.ActionTriggers: 2, utils.ActionPlans: 2, utils.SharedGroups: 2}
		err = mig.dmOut.DataDB().SetVersions(currentVersion, false)
		if err != nil {
			t.Error("Error when setting version for Accounts ", err.Error())
		}
		err, _ = mig.Migrate([]string{utils.MetaAccounts})
		if err != nil {
			t.Error("Error when migrating Accounts ", err.Error())
		}
		result, err := mig.dmOut.DataDB().GetAccount(testAccount.ID)
		if err != nil {
			t.Error("Error when getting Accounts ", err.Error())
		}
		if !reflect.DeepEqual(testAccount.BalanceMap["*voice"][0], result.BalanceMap["*voice"][0]) {
			t.Errorf("Expecting: %+v, received: %+v", testAccount.BalanceMap["*voice"][0], result.BalanceMap["*voice"][0])
		} else if !reflect.DeepEqual(testAccount, result) {
			t.Errorf("Expecting: %+v, received: %+v", testAccount, result)
		}
	case action == utils.MONGO:
		err := mig.oldDataDB.setV1Account(v1Acc)
		if err != nil {
			t.Error("Error when marshaling ", err.Error())
		}
		currentVersion := engine.Versions{utils.StatS: 2, utils.Thresholds: 2, utils.Accounts: 1, utils.Actions: 2, utils.ActionTriggers: 2, utils.ActionPlans: 2, utils.SharedGroups: 2}
		err = mig.dmOut.DataDB().SetVersions(currentVersion, false)
		if err != nil {
			t.Error("Error when setting version for Accounts ", err.Error())
		}
		err, _ = mig.Migrate([]string{utils.MetaAccounts})
		if err != nil {
			t.Error("Error when migrating Accounts ", err.Error())
		}
		result, err := mig.dmOut.DataDB().GetAccount(testAccount.ID)
		if err != nil {
			t.Error("Error when getting Accounts ", err.Error())
		}
		if !reflect.DeepEqual(testAccount, result) {
			t.Errorf("Expecting: %+v, received: %+v", testAccount, result)
		}
	case action == Move:
		if err := mig.dmIN.DataDB().SetAccount(testAccount); err != nil {
			log.Print("GOT ERR DMIN", err)
		}
		currentVersion := engine.CurrentDataDBVersions()
		err := mig.dmOut.DataDB().SetVersions(currentVersion, false)
		if err != nil {
			t.Error("Error when setting version for Accounts ", err.Error())
		}
		err, _ = mig.Migrate([]string{utils.MetaAccounts})
		if err != nil {
			t.Error("Error when migrating Accounts ", err.Error())
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
	case action == utils.REDIS:
		err := mig.oldDataDB.setV1ActionPlans(v1ap)
		if err != nil {
			t.Error("Error when setting v1 ActionPlan ", err.Error())
		}
		currentVersion := engine.Versions{utils.StatS: 2, utils.Thresholds: 2, utils.Accounts: 2, utils.Actions: 2, utils.ActionTriggers: 2, utils.ActionPlans: 1, utils.SharedGroups: 2}
		err = mig.dmOut.DataDB().SetVersions(currentVersion, false)
		if err != nil {
			t.Error("Error when setting version for ActionPlan ", err.Error())
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
	case action == utils.MONGO:
		err := mig.oldDataDB.setV1ActionPlans(v1ap)
		if err != nil {
			t.Error("Error when setting v1 ActionPlans ", err.Error())
		}
		currentVersion := engine.Versions{utils.StatS: 2, utils.Thresholds: 2, utils.Accounts: 2, utils.Actions: 2, utils.ActionTriggers: 2, utils.ActionPlans: 1, utils.SharedGroups: 2}
		err = mig.dmOut.DataDB().SetVersions(currentVersion, false)
		if err != nil {
			t.Error("Error when setting version for ActionPlan ", err.Error())
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
	case action == Move:
		if err := mig.dmIN.DataDB().SetActionPlan(ap.Id, ap, true, utils.NonTransactional); err != nil {
			t.Error("Error when setting ActionPlan ", err.Error())
		}
		currentVersion := engine.CurrentDataDBVersions()
		err := mig.dmOut.DataDB().SetVersions(currentVersion, false)
		if err != nil {
			t.Error("Error when setting version for ActionPlan ", err.Error())
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
	tim := time.Date(2012, time.February, 27, 23, 59, 59, 0, time.UTC)
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
	case action == utils.REDIS:
		err := mig.oldDataDB.setV1ActionTriggers(v1atrs)
		if err != nil {
			t.Error("Error when setting v1 ActionTriggers ", err.Error())
		}
		currentVersion := engine.Versions{utils.StatS: 2, utils.Thresholds: 2, utils.Accounts: 2, utils.Actions: 2, utils.ActionTriggers: 1, utils.ActionPlans: 2, utils.SharedGroups: 2}
		err = mig.dmOut.DataDB().SetVersions(currentVersion, false)
		if err != nil {
			t.Error("Error when setting version for ActionTriggers ", err.Error())
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
	case action == utils.MONGO:
		err, _ := mig.Migrate([]string{utils.MetaActionTriggers})
		if err != nil && err != utils.ErrNotImplemented {
			t.Error("Error when migrating ActionTriggers ", err.Error())
		}

	case action == Move:
		if err := mig.dmIN.SetActionTriggers(atrs[0].ID, atrs, utils.NonTransactional); err != nil {
			t.Error("Error when setting ActionPlan ", err.Error())
		}
		currentVersion := engine.CurrentDataDBVersions()
		err := mig.dmOut.DataDB().SetVersions(currentVersion, false)
		if err != nil {
			t.Error("Error when setting version for ActionTriggers ", err.Error())
		}
		err, _ = mig.Migrate([]string{utils.MetaActionTriggers})
		if err != nil {
			t.Error("Error when migrating ActionTriggers ", err.Error())
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
	v1act := &v1Actions{
		&v1Action{
			Id:               "test",
			ActionType:       "",
			BalanceType:      "",
			Direction:        "INBOUND",
			ExtraParameters:  "",
			ExpirationString: "",
			Balance: &v1Balance{
				Timings: []*engine.RITiming{
					&engine.RITiming{
						Years:     utils.Years{},
						Months:    utils.Months{},
						MonthDays: utils.MonthDays{},
						WeekDays:  utils.WeekDays{},
					},
				},
			},
		},
	}
	act := &engine.Actions{
		&engine.Action{
			Id:               "test",
			ActionType:       "",
			ExtraParameters:  "",
			ExpirationString: "",
			Weight:           0.00,
			Balance: &engine.BalanceFilter{
				Timings: []*engine.RITiming{
					&engine.RITiming{
						Years:     utils.Years{},
						Months:    utils.Months{},
						MonthDays: utils.MonthDays{},
						WeekDays:  utils.WeekDays{},
					},
				},
			},
		},
	}
	switch {
	case action == utils.REDIS:
		err := mig.oldDataDB.setV1Actions(v1act)
		if err != nil {
			t.Error("Error when setting v1 Actions ", err.Error())
		}
		currentVersion := engine.Versions{utils.StatS: 2, utils.Thresholds: 2, utils.Accounts: 2, utils.Actions: 1, utils.ActionTriggers: 2, utils.ActionPlans: 2, utils.SharedGroups: 2}
		err = mig.dmOut.DataDB().SetVersions(currentVersion, false)
		if err != nil {
			t.Error("Error when setting version for Actions ", err.Error())
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

	case action == utils.MONGO:
		err := mig.oldDataDB.setV1Actions(v1act)
		if err != nil {
			t.Error("Error when setting v1 Actions ", err.Error())
		}
		currentVersion := engine.Versions{utils.StatS: 2, utils.Thresholds: 2, utils.Accounts: 2, utils.Actions: 1, utils.ActionTriggers: 2, utils.ActionPlans: 2, utils.SharedGroups: 2}
		err = mig.dmOut.DataDB().SetVersions(currentVersion, false)
		if err != nil {
			t.Error("Error when setting version for Actions ", err.Error())
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
	case action == Move:
		if err := mig.dmIN.SetActions((*v1act)[0].Id, *act, utils.NonTransactional); err != nil {
			t.Error("Error when setting ActionPlan ", err.Error())
		}
		currentVersion := engine.CurrentDataDBVersions()
		err := mig.dmOut.DataDB().SetVersions(currentVersion, false)
		if err != nil {
			t.Error("Error when setting version for Actions ", err.Error())
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
	case action == utils.REDIS:
		err := mig.oldDataDB.setV1SharedGroup(v1sqp)
		if err != nil {
			t.Error("Error when setting v1 SharedGroup ", err.Error())
		}
		currentVersion := engine.Versions{utils.StatS: 2, utils.Thresholds: 2, utils.Accounts: 2, utils.Actions: 2, utils.ActionTriggers: 2, utils.ActionPlans: 2, utils.SharedGroups: 1}
		err = mig.dmOut.DataDB().SetVersions(currentVersion, false)
		if err != nil {
			t.Error("Error when setting version for SharedGroup ", err.Error())
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
	case action == utils.MONGO:
		err := mig.oldDataDB.setV1SharedGroup(v1sqp)
		if err != nil {
			t.Error("Error when setting v1 SharedGroup ", err.Error())
		}
		currentVersion := engine.Versions{utils.StatS: 2, utils.Thresholds: 2, utils.Accounts: 2, utils.Actions: 2, utils.ActionTriggers: 2, utils.ActionPlans: 2, utils.SharedGroups: 1}
		err = mig.dmOut.DataDB().SetVersions(currentVersion, false)
		if err != nil {
			t.Error("Error when setting version for SharedGroup ", err.Error())
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
	case action == Move:
		if err := mig.dmIN.SetSharedGroup(sqp, utils.NonTransactional); err != nil {
			t.Error("Error when setting SharedGroup ", err.Error())
		}
		currentVersion := engine.CurrentDataDBVersions()
		err := mig.dmOut.DataDB().SetVersions(currentVersion, false)
		if err != nil {
			t.Error("Error when setting version for SharedGroup ", err.Error())
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
	tim := time.Date(2012, time.February, 27, 23, 59, 59, 0, time.UTC)
	var filters []*engine.FilterRule
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

	x, _ := engine.NewFilterRule(engine.MetaGreaterOrEqual, "SetupInterval", []string{v1Sts.SetupInterval[0].String()})
	filters = append(filters, x)
	x, _ = engine.NewFilterRule(engine.MetaGreaterOrEqual, "UsageInterval", []string{v1Sts.UsageInterval[0].String()})
	filters = append(filters, x)
	x, _ = engine.NewFilterRule(engine.MetaGreaterOrEqual, "PddInterval", []string{v1Sts.PddInterval[0].String()})
	filters = append(filters, x)

	filter := &engine.Filter{Tenant: config.CgrConfig().DefaultTenant, ID: v1Sts.Id, Rules: filters}

	sqp := &engine.StatQueueProfile{
		Tenant:      "cgrates.org",
		ID:          "test",
		FilterIDs:   []string{v1Sts.Id},
		QueueLength: 10,
		TTL:         time.Duration(0) * time.Second,
		Metrics: []*utils.MetricWithParams{
			&utils.MetricWithParams{MetricID: "*asr", Parameters: ""},
			&utils.MetricWithParams{MetricID: "*acd", Parameters: ""},
			&utils.MetricWithParams{MetricID: "*acc", Parameters: ""},
		},
		ThresholdIDs: []string{"Test"},
		Blocker:      false,
		Stored:       true,
		Weight:       float64(0),
		MinItems:     0,
	}
	sq := &engine.StatQueue{Tenant: config.CgrConfig().DefaultTenant,
		ID:        v1Sts.Id,
		SQMetrics: make(map[string]engine.StatMetric),
	}
	for _, metricwparam := range sqp.Metrics {
		if metric, err := engine.NewStatMetric(metricwparam.MetricID, 0, metricwparam.Parameters); err != nil {
			t.Error("Error when creating newstatMETRIc ", err.Error())
		} else {
			if _, has := sq.SQMetrics[metricwparam.MetricID]; !has {
				sq.SQMetrics[metricwparam.MetricID] = metric
			}
		}
	}
	switch {
	case action == utils.REDIS:
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
		if !reflect.DeepEqual(sqp.ThresholdIDs, result.ThresholdIDs) {
			t.Errorf("Expecting: %+v, received: %+v", sqp.ThresholdIDs, result.ThresholdIDs)
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
	case action == utils.MONGO:
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
		if !reflect.DeepEqual(sqp.ThresholdIDs, result.ThresholdIDs) {
			t.Errorf("Expecting: %+v, received: %+v", sqp.ThresholdIDs, result.ThresholdIDs)
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
	case action == Move:
		if err := mig.dmIN.SetStatQueueProfile(sqp, true); err != nil {
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

func testMigratorSessionsCosts(t *testing.T) {
	switch action {
	case utils.REDIS:
		currentVersion := engine.CurrentStorDBVersions()
		currentVersion[utils.SessionsCosts] = 1
		err := mig.OutStorDB().SetVersions(currentVersion, false)
		if err != nil {
			t.Error("Error when setting version for SessionsCosts ", err.Error())
		}
		if vrs, err := mig.OutStorDB().GetVersions(utils.SessionsCosts); err != nil {
			t.Error(err)
		} else if vrs[utils.SessionsCosts] != 1 {
			t.Errorf("Expecting: 1, received: %+v", vrs[utils.SessionsCosts])
		}
		var qry string
		if isPostgres {
			qry = `
	CREATE TABLE sm_costs (
	  id SERIAL PRIMARY KEY,
	  cgrid VARCHAR(40) NOT NULL,
	  run_id  VARCHAR(64) NOT NULL,
	  origin_host VARCHAR(64) NOT NULL,
	  origin_id VARCHAR(128) NOT NULL,
	  cost_source VARCHAR(64) NOT NULL,
	  usage BIGINT NOT NULL,
	  cost_details jsonb,
	  created_at TIMESTAMP WITH TIME ZONE,
	  deleted_at TIMESTAMP WITH TIME ZONE NULL,
	  UNIQUE (cgrid, run_id)
	);
		`
		} else {
			qry = fmt.Sprint("CREATE TABLE sm_costs (  id int(11) NOT NULL AUTO_INCREMENT,  cgrid varchar(40) NOT NULL,  run_id  varchar(64) NOT NULL,  origin_host varchar(64) NOT NULL,  origin_id varchar(128) NOT NULL,  cost_source varchar(64) NOT NULL,  `usage` BIGINT NOT NULL,  cost_details MEDIUMTEXT,  created_at TIMESTAMP NULL,deleted_at TIMESTAMP NULL,  PRIMARY KEY (`id`),UNIQUE KEY costid (cgrid, run_id),KEY origin_idx (origin_host, origin_id),KEY run_origin_idx (run_id, origin_id),KEY deleted_at_idx (deleted_at));")
		}
		if _, err := mig.OutStorDB().(*engine.SQLStorage).Db.Exec("DROP TABLE IF EXISTS sessions_costs;"); err != nil {
			t.Error(err)
		}
		if _, err := mig.OutStorDB().(*engine.SQLStorage).Db.Exec("DROP TABLE IF EXISTS sm_costs;"); err != nil {
			t.Error(err)
		}
		if _, err := mig.OutStorDB().(*engine.SQLStorage).Db.Exec(qry); err != nil {
			t.Error(err)
		}
		err, _ = mig.Migrate([]string{utils.MetaSessionsCosts})
		if vrs, err := mig.OutStorDB().GetVersions(utils.SessionsCosts); err != nil {
			t.Error(err)
		} else if vrs[utils.SessionsCosts] != 2 {
			t.Errorf("Expecting: 2, received: %+v", vrs[utils.SessionsCosts])
		}
	}
}

func testMigratorThreshold(t *testing.T) {
	tim := time.Date(2012, time.February, 27, 23, 59, 59, 0, time.UTC)
	tenant := config.CgrConfig().DefaultTenant
	var filters []*engine.FilterRule
	threshold := &v2ActionTrigger{
		ID:             "test2",              // original csv tag
		UniqueID:       "testUUID",           // individual id
		ThresholdType:  "*min_event_counter", //*min_event_counter, *max_event_counter, *min_balance_counter, *max_balance_counter, *min_balance, *max_balance, *balance_expired
		ThresholdValue: 5.32,
		Recurrent:      false,                          // reset excuted flag each run
		MinSleep:       time.Duration(5) * time.Second, // Minimum duration between two executions in case of recurrent triggers
		ExpirationDate: time.Now(),
		ActivationDate: time.Now(),
		Balance: &engine.BalanceFilter{
			ID:             utils.StringPointer("TESTZ"),
			Timings:        []*engine.RITiming{},
			ExpirationDate: utils.TimePointer(tim),
			Type:           utils.StringPointer(utils.MONETARY),
			Directions:     utils.StringMapPointer(utils.NewStringMap(utils.OUT)),
		},
		Weight:            0,
		ActionsID:         "Action1",
		MinQueuedItems:    10, // Trigger actions only if this number is hit (stats only)
		Executed:          false,
		LastExecutionTime: time.Now(),
	}
	x, err := engine.NewFilterRule(engine.MetaRSR, "Directions", threshold.Balance.Directions.Slice())
	if err != nil {
		t.Error("Error when creating new NewFilterRule", err.Error())
	}
	filters = append(filters, x)

	filter := &engine.Filter{Tenant: config.CgrConfig().DefaultTenant, ID: *threshold.Balance.ID, Rules: filters}

	thp := &engine.ThresholdProfile{
		ID:                 threshold.ID,
		Tenant:             config.CgrConfig().DefaultTenant,
		FilterIDs:          []string{filter.ID},
		Blocker:            false,
		Weight:             threshold.Weight,
		ActivationInterval: &utils.ActivationInterval{threshold.ExpirationDate, threshold.ActivationDate},
		MinSleep:           threshold.MinSleep,
	}
	th := &engine.Threshold{
		Tenant: config.CgrConfig().DefaultTenant,
		ID:     threshold.ID,
	}

	switch {
	case action == utils.REDIS:
		if err := mig.dmIN.SetFilter(filter); err != nil {
			t.Error("Error when setting Filter ", err.Error())
		}
		if err := mig.dmIN.SetThresholdProfile(thp, true); err != nil {
			t.Error("Error when setting threshold ", err.Error())
		}
		if err := mig.dmIN.SetThreshold(th); err != nil {
			t.Error("Error when setting threshold ", err.Error())
		}
		err := mig.oldDataDB.setV2ActionTrigger(threshold)
		if err != nil {
			t.Error("Error when setting threshold ", err.Error())
		}
		currentVersion := engine.CurrentDataDBVersions()
		currentVersion[utils.Thresholds] = 1
		err = mig.dmOut.DataDB().SetVersions(currentVersion, false)
		if err != nil {
			t.Error("Error when setting version for threshold ", err.Error())
		}
		err, _ = mig.Migrate([]string{utils.MetaThresholds})
		if err != nil {
			t.Error("Error when migrating threshold ", err.Error())
		}
		result, err := mig.dmOut.GetThreshold("cgrates.org", threshold.ID, true, utils.NonTransactional)
		if err != nil {
			t.Error("Error when getting threshold ", err.Error())
		}
		if !reflect.DeepEqual(th, result) {
			t.Errorf("Expecting: %+v, received: %+v", th, result)
		}
		thpr, err := mig.dmOut.GetThresholdProfile(thp.Tenant, thp.ID, true, utils.NonTransactional)
		if err != nil {
			t.Error("Error when getting thresholdProfile ", err.Error())
		}
		if !reflect.DeepEqual(thp.ID, thpr.ID) {
			t.Errorf("Expecting: %+v, received: %+v", thp.ID, thpr.ID)
		}
	case action == utils.MONGO:
		if err := mig.dmIN.SetFilter(filter); err != nil {
			t.Error("Error when setting Filter ", err.Error())
		}
		if err := mig.dmIN.SetThresholdProfile(thp, true); err != nil {
			t.Error("Error when setting Threshold ", err.Error())
		}
		if err := mig.dmIN.SetThreshold(th); err != nil {
			t.Error("Error when setting Threshold ", err.Error())
		}
		err := mig.oldDataDB.setV2ActionTrigger(threshold)
		if err != nil {
			t.Error("Error when setting Threshold ", err.Error())
		}
		currentVersion := engine.CurrentDataDBVersions()
		currentVersion[utils.Thresholds] = 1
		err = mig.dmOut.DataDB().SetVersions(currentVersion, false)
		if err != nil {
			t.Error("Error when setting version for Threshold ", err.Error())
		}
		err, _ = mig.Migrate([]string{utils.MetaThresholds})
		if err != nil {
			t.Error("Error when migrating Threshold ", err.Error())
		}
		result, err := mig.dmOut.GetThreshold("cgrates.org", threshold.ID, true, utils.NonTransactional)
		if err != nil {
			t.Error("Error when getting Threshold ", err.Error())
		}
		if !reflect.DeepEqual(th, result) {
			t.Errorf("Expecting: %+v, received: %+v", th, result)
		}
		thpr, err := mig.dmOut.GetThresholdProfile(thp.Tenant, thp.ID, true, utils.NonTransactional)
		if err != nil {
			t.Error("Error when getting thresholdProfile ", err.Error())
		}
		if !reflect.DeepEqual(thp.ID, thpr.ID) {
			t.Errorf("Expecting: %+v, received: %+v", thp.ID, thpr.ID)
		}
	case action == Move:
		if err := mig.dmIN.SetFilter(filter); err != nil {
			t.Error("Error when setting Filter ", err.Error())
		}
		if err := mig.dmIN.SetThresholdProfile(thp, true); err != nil {
			t.Error("Error when setting Threshold ", err.Error())
		}
		if err := mig.dmIN.SetThreshold(th); err != nil {
			t.Error("Error when setting Threshold ", err.Error())
		}
		err := mig.oldDataDB.setV2ActionTrigger(threshold)
		if err != nil {
			t.Error("Error when setting Threshold ", err.Error())
		}
		currentVersion := engine.CurrentDataDBVersions()
		err = mig.dmOut.DataDB().SetVersions(currentVersion, false)
		if err != nil {
			t.Error("Error when setting version for Threshold ", err.Error())
		}
		err, _ = mig.Migrate([]string{utils.MetaThresholds})
		if err != nil {
			t.Error("Error when migrating Threshold ", err.Error())
		}
		result, err := mig.dmOut.GetThreshold(tenant, threshold.ID, true, utils.NonTransactional)
		if err != nil {
			t.Error("Error when getting Threshold ", err.Error())
		}
		if !reflect.DeepEqual(th, result) {
			t.Errorf("Expecting: %+v, received: %+v", th, result)
		}
		thpr, err := mig.dmOut.GetThresholdProfile(thp.Tenant, thp.ID, true, utils.NonTransactional)
		if err != nil {
			t.Error("Error when getting ThresholdProfile ", err.Error())
		}
		if !reflect.DeepEqual(thp.ID, thpr.ID) {
			t.Errorf("Expecting: %+v, received: %+v", thp.ID, thpr.ID)
		}
	}
}

func testMigratorAlias(t *testing.T) {
	alias := &engine.Alias{
		Direction: "*out",
		Tenant:    "cgrates.org",
		Category:  "call",
		Account:   "dan",
		Subject:   "dan",
		Context:   "*rating",
		Values: engine.AliasValues{
			&engine.AliasValue{
				DestinationId: "EU_LANDLINE",
				Pairs: engine.AliasPairs{
					"Subject": map[string]string{
						"dan": "dan1",
						"rif": "rif1",
					},
					"Cli": map[string]string{
						"0723": "0724",
					},
				},
				Weight: 10,
			},

			&engine.AliasValue{
				DestinationId: "GLOBAL1",
				Pairs:         engine.AliasPairs{"Subject": map[string]string{"dan": "dan2"}},
				Weight:        20,
			},
		},
	}
	switch action {
	case Move:
		if err := mig.dmIN.DataDB().SetAlias(alias, utils.NonTransactional); err != nil {
			t.Error("Error when setting Alias ", err.Error())
		}
		currentVersion := engine.CurrentDataDBVersions()
		err := mig.dmOut.DataDB().SetVersions(currentVersion, false)
		if err != nil {
			t.Error("Error when setting version for Alias ", err.Error())
		}
		err, _ = mig.Migrate([]string{utils.MetaAlias})
		if err != nil {
			t.Error("Error when migrating Alias ", err.Error())
		}
		result, err := mig.dmOut.DataDB().GetAlias(alias.GetId(), true, utils.NonTransactional)
		if err != nil {
			t.Error("Error when getting Alias ", err.Error())
		}
		if !reflect.DeepEqual(alias, result) {
			t.Errorf("Expecting: %+v, received: %+v", alias, result)
		}
	}
}

func testMigratorCdrStats(t *testing.T) {
	cdrs := &engine.CdrStats{
		Id:              "",
		QueueLength:     10,                             // Number of items in the stats buffer
		TimeWindow:      time.Duration(1) * time.Second, // Will only keep the CDRs who's call setup time is not older than time.Now()-TimeWindow
		SaveInterval:    time.Duration(1) * time.Second,
		Metrics:         []string{engine.ASR, engine.PDD, engine.ACD, engine.TCD, engine.ACC, engine.TCC, engine.DDC},
		SetupInterval:   []time.Time{time.Date(2012, time.February, 27, 23, 59, 59, 0, time.UTC)}, // CDRFieldFilter on SetupInterval, 2 or less items (>= start interval,< stop_interval)
		TOR:             []string{""},                                                             // CDRFieldFilter on TORs
		CdrHost:         []string{""},                                                             // CDRFieldFilter on CdrHosts
		CdrSource:       []string{""},                                                             // CDRFieldFilter on CdrSources
		ReqType:         []string{""},                                                             // CDRFieldFilter on RequestTypes
		Direction:       []string{""},                                                             // CDRFieldFilter on Directions
		Tenant:          []string{""},                                                             // CDRFieldFilter on Tenants
		Category:        []string{""},                                                             // CDRFieldFilter on Categories
		Account:         []string{""},                                                             // CDRFieldFilter on Accounts
		Subject:         []string{""},                                                             // CDRFieldFilter on Subjects
		DestinationIds:  []string{""},                                                             // CDRFieldFilter on DestinationPrefixes
		UsageInterval:   []time.Duration{time.Duration(1) * time.Second},                          // CDRFieldFilter on UsageInterval, 2 or less items (>= Usage, <Usage)
		PddInterval:     []time.Duration{time.Duration(1) * time.Second},                          // CDRFieldFilter on PddInterval, 2 or less items (>= Pdd, <Pdd)
		Supplier:        []string{},                                                               // CDRFieldFilter on Suppliers
		DisconnectCause: []string{},                                                               // Filter on DisconnectCause
		MediationRunIds: []string{},                                                               // CDRFieldFilter on MediationRunIds
		RatedAccount:    []string{},                                                               // CDRFieldFilter on RatedAccounts
		RatedSubject:    []string{},                                                               // CDRFieldFilter on RatedSubjects
		CostInterval:    []float64{},                                                              // CDRFieldFilter on CostInterval, 2 or less items, (>=Cost, <Cost)
	}
	switch action {
	case Move:
		if err := mig.dmIN.SetCdrStats(cdrs); err != nil {
			t.Error("Error when setting CdrStats ", err.Error())
		}
		currentVersion := engine.CurrentDataDBVersions()
		err := mig.dmOut.DataDB().SetVersions(currentVersion, false)
		if err != nil {
			t.Error("Error when setting version for CdrStats ", err.Error())
		}
		err, _ = mig.Migrate([]string{utils.MetaCdrStats})
		if err != nil {
			t.Error("Error when migrating CdrStats ", err.Error())
		}
		result, err := mig.dmOut.GetCdrStats("")
		if err != nil {
			t.Error("Error when getting CdrStats ", err.Error())
		}
		if !reflect.DeepEqual(cdrs.Metrics, result.Metrics) {
			t.Errorf("Expecting: %v, received: %v", cdrs.Metrics, result.Metrics)
		} else if !reflect.DeepEqual(cdrs.SetupInterval, result.SetupInterval) {
			t.Errorf("Expecting: %+v, received: %+v", cdrs.SetupInterval, result.SetupInterval)
		} else if !reflect.DeepEqual(cdrs.PddInterval, result.PddInterval) {
			t.Errorf("Expecting: %+v, received: %+v", cdrs.PddInterval, result.PddInterval)
		} else if !reflect.DeepEqual(cdrs.SaveInterval, result.SaveInterval) {
			t.Errorf("Expecting: %+v, received: %+v", cdrs.SaveInterval, result.SaveInterval)
		}
	}
}

func testMigratorDerivedChargers(t *testing.T) {
	dcs := &utils.DerivedChargers{
		DestinationIDs: make(utils.StringMap),
		Chargers: []*utils.DerivedCharger{
			&utils.DerivedCharger{
				RunID: "extra1", RunFilters: "^filteredHeader1/filterValue1/", RequestTypeField: "^prepaid", DirectionField: utils.META_DEFAULT,
				TenantField: utils.META_DEFAULT, CategoryField: utils.META_DEFAULT, AccountField: "rif", SubjectField: "rif", DestinationField: utils.META_DEFAULT,
				SetupTimeField: utils.META_DEFAULT, PDDField: utils.META_DEFAULT, AnswerTimeField: utils.META_DEFAULT, UsageField: utils.META_DEFAULT,
				SupplierField: utils.META_DEFAULT, DisconnectCauseField: utils.META_DEFAULT, CostField: utils.META_DEFAULT, RatedField: utils.META_DEFAULT,
			},
		},
	}
	keyDCS := utils.ConcatenatedKey("*out", "itsyscom.com", "call", "dan", "dan")
	switch action {
	case Move:
		if err := mig.dmIN.DataDB().SetDerivedChargers(keyDCS, dcs, utils.NonTransactional); err != nil {
			t.Error("Error when setting DerivedChargers ", err.Error())
		}
		currentVersion := engine.CurrentDataDBVersions()
		err := mig.dmOut.DataDB().SetVersions(currentVersion, false)
		if err != nil {
			t.Error("Error when setting version for DerivedChargers ", err.Error())
		}
		err, _ = mig.Migrate([]string{utils.MetaDerivedChargersV})
		if err != nil {
			t.Error("Error when migrating DerivedChargers ", err.Error())
		}
		result, err := mig.dmOut.GetDerivedChargers(keyDCS, true, utils.NonTransactional)
		if err != nil {
			t.Error("Error when getting DerivedChargers ", err.Error())
		}
		if !reflect.DeepEqual(dcs, result) {
			t.Errorf("Expecting: %v, received: %v", dcs, result)
		}
	}
}

func testMigratorDestinations(t *testing.T) {
	dst := &engine.Destination{Id: "CRUDDestination2", Prefixes: []string{"+491", "+492", "+493"}}

	switch action {
	case Move:
		if err := mig.dmIN.DataDB().SetDestination(dst, utils.NonTransactional); err != nil {
			t.Error("Error when setting Destinations ", err.Error())
		}
		currentVersion := engine.CurrentDataDBVersions()
		err := mig.dmOut.DataDB().SetVersions(currentVersion, false)
		if err != nil {
			t.Error("Error when setting version for Destinations ", err.Error())
		}
		err, _ = mig.Migrate([]string{utils.MetaDestinations})
		if err != nil {
			t.Error("Error when migrating Destinations ", err.Error())
		}
		result, err := mig.dmOut.DataDB().GetDestination(dst.Id, true, utils.NonTransactional)
		if err != nil {
			t.Error("Error when getting Destinations ", err.Error())
		}
		if !reflect.DeepEqual(dst, result) {
			t.Errorf("Expecting: %v, received: %v", dst, result)
		}
	}
}

func testMigratorReverseDestinations(t *testing.T) {
	dst := &engine.Destination{Id: "CRUDReverseDestination", Prefixes: []string{"+494", "+495", "+496"}}
	switch action {
	case Move:
		if err := mig.dmIN.DataDB().SetDestination(dst, utils.NonTransactional); err != nil {
			t.Error("Error when setting Destinations ", err.Error())
		}
		if err := mig.dmIN.DataDB().SetReverseDestination(dst, utils.NonTransactional); err != nil {
			t.Error("Error when setting ReverseDestinations ", err.Error())
		}
		currentVersion := engine.CurrentDataDBVersions()
		err := mig.dmOut.DataDB().SetVersions(currentVersion, false)
		if err != nil {
			t.Error("Error when setting version for ReverseDestinations ", err.Error())
		}
		err, _ = mig.Migrate([]string{utils.MetaReverseDestinations})
		if err != nil {
			t.Error("Error when migrating ReverseDestinations ", err.Error())
		}
		for i, _ := range dst.Prefixes {
			result, err := mig.dmOut.DataDB().GetReverseDestination(dst.Prefixes[i], true, utils.NonTransactional)
			if err != nil {
				t.Error("Error when getting ReverseDestinations ", err.Error())
			} else if !reflect.DeepEqual([]string{dst.Id}, result) {
				t.Errorf("Expecting: %v, received: %v", []string{dst.Id}, result)
			}
		}
	}
}

func testMigratorLCR(t *testing.T) {
	lcr := &engine.LCR{
		Tenant:    "cgrates.org",
		Category:  "call",
		Direction: "*out",
		Account:   "testOnStorITCRUDLCR",
		Subject:   "testOnStorITCRUDLCR",
		Activations: []*engine.LCRActivation{
			&engine.LCRActivation{
				ActivationTime: time.Date(2012, 1, 1, 0, 0, 0, 0, time.UTC),
				Entries: []*engine.LCREntry{
					&engine.LCREntry{
						DestinationId:  "EU_LANDLINE",
						RPCategory:     "LCR_STANDARD",
						Strategy:       "*static",
						StrategyParams: "ivo;dan;rif",
						Weight:         10,
					},
					&engine.LCREntry{
						DestinationId:  "*any",
						RPCategory:     "LCR_STANDARD",
						Strategy:       "*lowest_cost",
						StrategyParams: "",
						Weight:         20,
					},
				},
			},
		},
	}
	switch action {
	case Move:
		if err := mig.dmIN.SetLCR(lcr, utils.NonTransactional); err != nil {
			t.Error("Error when setting Lcr ", err.Error())
		}
		currentVersion := engine.CurrentDataDBVersions()
		err := mig.dmOut.DataDB().SetVersions(currentVersion, false)
		if err != nil {
			t.Error("Error when setting version for Lcr ", err.Error())
		}
		err, _ = mig.Migrate([]string{utils.MetaLCR})
		if err != nil {
			t.Error("Error when migrating Lcr ", err.Error())
		}
		result, err := mig.dmOut.GetLCR(lcr.GetId(), true, utils.NonTransactional)
		if err != nil {
			t.Error("Error when getting Lcr ", err.Error())
		}
		if !reflect.DeepEqual(lcr, result) {
			t.Errorf("Expecting: %v, received: %v", lcr, result)
		}
	}
}

func testMigratorRatingPlan(t *testing.T) {
	rp := &engine.RatingPlan{
		Id: "CRUDRatingPlan",
		Timings: map[string]*engine.RITiming{
			"59a981b9": &engine.RITiming{
				Years:     utils.Years{},
				Months:    utils.Months{},
				MonthDays: utils.MonthDays{},
				WeekDays:  utils.WeekDays{1, 2, 3, 4, 5},
				StartTime: "00:00:00",
			},
		},
		Ratings: map[string]*engine.RIRate{
			"ebefae11": &engine.RIRate{
				ConnectFee: 0,
				Rates: []*engine.Rate{
					&engine.Rate{
						GroupIntervalStart: 0,
						Value:              0.2,
						RateIncrement:      time.Second,
						RateUnit:           time.Minute,
					},
				},
				RoundingMethod:   utils.ROUNDING_MIDDLE,
				RoundingDecimals: 4,
			},
		},
		DestinationRates: map[string]engine.RPRateList{
			"GERMANY": []*engine.RPRate{
				&engine.RPRate{
					Timing: "59a981b9",
					Rating: "ebefae11",
					Weight: 10,
				},
			},
		},
	}
	switch action {
	case Move:
		if err := mig.dmIN.SetRatingPlan(rp, utils.NonTransactional); err != nil {
			t.Error("Error when setting RatingPlan ", err.Error())
		}
		currentVersion := engine.CurrentDataDBVersions()
		err := mig.dmOut.DataDB().SetVersions(currentVersion, false)
		if err != nil {
			t.Error("Error when setting version for RatingPlan ", err.Error())
		}
		err, _ = mig.Migrate([]string{utils.MetaRatingPlans})
		if err != nil {
			t.Error("Error when migrating RatingPlan ", err.Error())
		}
		result, err := mig.dmOut.GetRatingPlan(rp.Id, true, utils.NonTransactional)
		if err != nil {
			t.Error("Error when getting RatingPlan ", err.Error())
		}
		if !reflect.DeepEqual(rp, result) {
			t.Errorf("Expecting: %v, received: %v", rp, result)
		}
	}
}

func testMigratorRatingProfile(t *testing.T) {
	rpf := &engine.RatingProfile{
		Id: "*out:test:1:trp",
		RatingPlanActivations: engine.RatingPlanActivations{
			&engine.RatingPlanActivation{
				ActivationTime:  time.Date(2013, 10, 1, 0, 0, 0, 0, time.UTC),
				RatingPlanId:    "TDRT",
				FallbackKeys:    []string{"*out:test:1:danb", "*out:test:1:rif"},
				CdrStatQueueIds: []string{},
			}},
	}
	switch action {
	case Move:
		if err := mig.dmIN.SetRatingProfile(rpf, utils.NonTransactional); err != nil {
			t.Error("Error when setting RatingProfile ", err.Error())
		}
		currentVersion := engine.CurrentDataDBVersions()
		err := mig.dmOut.DataDB().SetVersions(currentVersion, false)
		if err != nil {
			t.Error("Error when setting version for RatingProfile ", err.Error())
		}
		err, _ = mig.Migrate([]string{utils.MetaRatingProfile})
		if err != nil {
			t.Error("Error when migrating RatingProfile ", err.Error())
		}
		result, err := mig.dmOut.GetRatingProfile(rpf.Id, true, utils.NonTransactional)
		if err != nil {
			t.Error("Error when getting RatingProfile ", err.Error())
		}
		if !reflect.DeepEqual(rpf, result) {
			t.Errorf("Expecting: %v, received: %v", rpf, result)
		}
	}
}
func testMigratorRQF(t *testing.T) {
	fp := &engine.Filter{
		Tenant: "cgrates.org",
		ID:     "Filter1",
		Rules: []*engine.FilterRule{
			&engine.FilterRule{
				FieldName: "Account",
				Type:      "*string",
				Values:    []string{"1001", "1002"},
			},
		},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			ExpiryTime:     time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
	}
	switch action {
	case Move:
		if err := mig.dmIN.SetFilter(fp); err != nil {
			t.Error("Error when setting FilterRule ", err.Error())
		}
		currentVersion := engine.CurrentDataDBVersions()
		err := mig.dmOut.DataDB().SetVersions(currentVersion, false)
		if err != nil {
			t.Error("Error when setting version for FilterRule ", err.Error())
		}
		err, _ = mig.Migrate([]string{utils.MetaRQF})
		if err != nil {
			t.Error("Error when migrating FilterRule ", err.Error())
		}
		result, err := mig.dmOut.GetFilter(fp.Tenant, fp.ID, true, utils.NonTransactional)
		if err != nil {
			t.Error("Error when getting FilterRule ", err.Error())
		}
		if !reflect.DeepEqual(fp, result) {
			t.Errorf("Expecting: %v, received: %v", fp, result)
		}
	}
}

func testMigratorResource(t *testing.T) {
	var filters []*engine.FilterRule
	rL := &engine.ResourceProfile{
		Tenant:    "cgrates.org",
		ID:        "RL_TEST2",
		Weight:    10,
		FilterIDs: []string{"FLTR_RES_RL_TEST2"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 3, 13, 43, 0, 0, time.UTC),
			ExpiryTime:     time.Date(2015, 7, 3, 13, 43, 0, 0, time.UTC)},
		Limit:        1,
		ThresholdIDs: []string{"TEST_ACTIONS"},
		UsageTTL:     time.Duration(1 * time.Millisecond),
	}
	switch action {
	case Move:
		x, _ := engine.NewFilterRule(engine.MetaGreaterOrEqual, "PddInterval", []string{rL.UsageTTL.String()})
		filters = append(filters, x)

		filter := &engine.Filter{Tenant: "cgrates.org", ID: "FLTR_RES_RL_TEST2", Rules: filters}

		if err := mig.dmIN.SetFilter(filter); err != nil {
			t.Error("Error when setting filter ", err.Error())
		}
		if err := mig.dmIN.SetResourceProfile(rL, true); err != nil {
			t.Error("Error when setting ResourceProfile ", err.Error())
		}
		currentVersion := engine.CurrentDataDBVersions()
		err := mig.dmOut.DataDB().SetVersions(currentVersion, false)
		if err != nil {
			t.Error("Error when setting version for ResourceProfile ", err.Error())
		}
		err, _ = mig.Migrate([]string{utils.MetaResource})
		if err != nil {
			t.Error("Error when migrating ResourceProfile ", err.Error())
		}
		result, err := mig.dmOut.GetResourceProfile(rL.Tenant, rL.ID, true, utils.NonTransactional)
		if err != nil {
			t.Error("Error when getting ResourceProfile ", err.Error())
		}
		if !reflect.DeepEqual(rL, result) {
			t.Errorf("Expecting: %v, received: %v", rL, result)
		}
	}
}

func testMigratorSubscribers(t *testing.T) {
	sbsc := &engine.SubscriberData{
		ExpTime: time.Date(2012, 1, 1, 0, 0, 0, 0, time.UTC),
		Filters: utils.ParseRSRFieldsMustCompile("^*default", utils.INFIELD_SEP)}
	sbscID := "testOnStorITCRUDSubscribers"

	switch action {
	case Move:
		if err := mig.dmIN.SetSubscriber(sbscID, sbsc); err != nil {
			t.Error("Error when setting RatingProfile ", err.Error())
		}
		currentVersion := engine.CurrentDataDBVersions()
		err := mig.dmOut.DataDB().SetVersions(currentVersion, false)
		if err != nil {
			t.Error("Error when setting version for RatingProfile ", err.Error())
		}
		err, _ = mig.Migrate([]string{utils.MetaSubscribers})
		if err != nil {
			t.Error("Error when migrating RatingProfile ", err.Error())
		}
		result, err := mig.dmOut.GetSubscribers()
		if err != nil {
			t.Error("Error when getting RatingProfile ", err.Error())
		}
		if !reflect.DeepEqual(sbsc.ExpTime, result["testOnStorITCRUDSubscribers"].ExpTime) {
			t.Errorf("Expecting: %v, received: %v", sbsc.ExpTime, result["testOnStorITCRUDSubscribers"].ExpTime)
		} else if !reflect.DeepEqual(sbsc.Filters[0].Id, result["testOnStorITCRUDSubscribers"].Filters[0].Id) {
			t.Errorf("Expecting: %v, received: %v", sbsc.Filters[0].Id, result["testOnStorITCRUDSubscribers"].Filters[0].Id)
		}
	}
}

func testMigratorTimings(t *testing.T) {
	tmg := &utils.TPTiming{
		ID:        "TEST_TMG",
		Years:     utils.Years{2016, 2017},
		Months:    utils.Months{time.January, time.February, time.March},
		MonthDays: utils.MonthDays{1, 2, 3, 4},
		WeekDays:  utils.WeekDays{},
		StartTime: "00:00:00",
		EndTime:   "",
	}
	switch action {
	case Move:
		if err := mig.dmIN.SetTiming(tmg); err != nil {
			t.Error("Error when setting Timings ", err.Error())
		}
		currentVersion := engine.CurrentDataDBVersions()
		err := mig.dmOut.DataDB().SetVersions(currentVersion, false)
		if err != nil {
			t.Error("Error when setting version for Timings ", err.Error())
		}
		err, _ = mig.Migrate([]string{utils.MetaTiming})
		if err != nil {
			t.Error("Error when migrating Timings ", err.Error())
		}
		result, err := mig.dmOut.GetTiming(tmg.ID, true, utils.NonTransactional)
		if err != nil {
			t.Error("Error when getting Timings ", err.Error())
		}
		if !reflect.DeepEqual(tmg, result) {
			t.Errorf("Expecting: %v, received: %v", tmg, result)
		}
	}
}

func testMigratorAttributeProfile(t *testing.T) {
	mapSubstitutes := make(map[string]map[string]*v1Attribute)
	mapSubstitutes["FL1"] = make(map[string]*v1Attribute)
	mapSubstitutes["FL1"]["In1"] = &v1Attribute{
		FieldName:  "FL1",
		Initial:    "In1",
		Substitute: "Al1",
		Append:     true,
	}
	v1Attribute := &v1AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "attributeprofile1",
		Contexts:  []string{utils.MetaSessionS},
		FilterIDs: []string{"filter1"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			ExpiryTime:     time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
		Attributes: mapSubstitutes,
		Weight:     20,
	}
	attrPrf := &engine.AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "attributeprofile1",
		Contexts:  []string{utils.MetaSessionS},
		FilterIDs: []string{"filter1"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			ExpiryTime:     time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
		Attributes: []*engine.Attribute{
			&engine.Attribute{
				FieldName:  "FL1",
				Initial:    "In1",
				Substitute: "Al1",
				Append:     true,
			},
		},
		Weight: 20,
	}
	filterAttr := &engine.Filter{
		Tenant: attrPrf.Tenant,
		ID:     attrPrf.FilterIDs[0],
		Rules: []*engine.FilterRule{
			&engine.FilterRule{
				FieldName: "Name",
				Type:      "Type",
				Values:    []string{"Val1"},
			},
		},
	}
	switch {
	case action == utils.REDIS:
		if err := mig.dmIN.SetFilter(filterAttr); err != nil {
			t.Error("Error when setting Filter ", err.Error())
		}
		if err := mig.dmIN.SetAttributeProfile(attrPrf, true); err != nil {
			t.Error("Error when setting attributeProfile ", err.Error())
		}
		err := mig.oldDataDB.setV1AttributeProfile(v1Attribute)
		if err != nil {
			t.Error("Error when setting V1AttributeProfile ", err.Error())
		}
		currentVersion := engine.CurrentDataDBVersions()
		currentVersion[utils.Attributes] = 1
		err = mig.dmOut.DataDB().SetVersions(currentVersion, false)
		if err != nil {
			t.Error("Error when setting version for attributeProfile ", err.Error())
		}
		err, _ = mig.Migrate([]string{utils.MetaAttributes})
		if err != nil {
			t.Error("Error when migrating AttributeProfile ", err.Error())
		}
		result, err := mig.dmOut.GetAttributeProfile(attrPrf.Tenant, attrPrf.ID, true, utils.NonTransactional)
		if err != nil {
			t.Error("Error when getting AttributeProfile ", err.Error())
		}
		if !reflect.DeepEqual(attrPrf.Tenant, result.Tenant) {
			t.Errorf("Expecting: %+v, received: %+v", attrPrf.Tenant, result.Tenant)
		} else if !reflect.DeepEqual(attrPrf.ID, result.ID) {
			t.Errorf("Expecting: %+v, received: %+v", attrPrf.ID, result.ID)
		} else if !reflect.DeepEqual(attrPrf.FilterIDs, result.FilterIDs) {
			t.Errorf("Expecting: %+v, received: %+v", attrPrf.FilterIDs, result.FilterIDs)
		} else if !reflect.DeepEqual(attrPrf.Contexts, result.Contexts) {
			t.Errorf("Expecting: %+v, received: %+v", attrPrf.Contexts, result.Contexts)
		} else if !reflect.DeepEqual(attrPrf.ActivationInterval, result.ActivationInterval) {
			t.Errorf("Expecting: %+v, received: %+v", attrPrf.ActivationInterval, result.ActivationInterval)
		} else if !reflect.DeepEqual(attrPrf.Attributes, result.Attributes) {
			t.Errorf("Expecting: %+v, received: %+v", attrPrf.Attributes, result.Attributes)
		}
	case action == utils.MONGO:
		if err := mig.dmIN.SetAttributeProfile(attrPrf, true); err != nil {
			t.Error("Error when setting attributeProfile ", err.Error())
		}
		err := mig.oldDataDB.setV1AttributeProfile(v1Attribute)
		if err != nil {
			t.Error("Error when setting V1AttributeProfile ", err.Error())
		}
		currentVersion := engine.CurrentDataDBVersions()
		currentVersion[utils.Attributes] = 1
		err = mig.dmOut.DataDB().SetVersions(currentVersion, false)
		if err != nil {
			t.Error("Error when setting version for attributeProfile ", err.Error())
		}
		err, _ = mig.Migrate([]string{utils.MetaAttributes})
		if err != nil {
			t.Error("Error when migrating attributeProfile ", err.Error())
		}
		result, err := mig.dmOut.GetAttributeProfile("cgrates.org", attrPrf.ID, true, utils.NonTransactional)
		if err != nil {
			t.Error("Error when getting attributeProfile ", err.Error())
		}
		if !reflect.DeepEqual(attrPrf.Tenant, result.Tenant) {
			t.Errorf("Expecting: %+v, received: %+v", attrPrf.Tenant, result.Tenant)
		} else if !reflect.DeepEqual(attrPrf.ID, result.ID) {
			t.Errorf("Expecting: %+v, received: %+v", attrPrf.ID, result.ID)
		} else if !reflect.DeepEqual(attrPrf.FilterIDs, result.FilterIDs) {
			t.Errorf("Expecting: %+v, received: %+v", attrPrf.FilterIDs, result.FilterIDs)
		} else if !reflect.DeepEqual(attrPrf.Contexts, result.Contexts) {
			t.Errorf("Expecting: %+v, received: %+v", attrPrf.Contexts, result.Contexts)
		} else if !reflect.DeepEqual(attrPrf.ActivationInterval, result.ActivationInterval) {
			t.Errorf("Expecting: %+v, received: %+v", attrPrf.ActivationInterval, result.ActivationInterval)
		} else if !reflect.DeepEqual(attrPrf.Attributes, result.Attributes) {
			t.Errorf("Expecting: %+v, received: %+v", attrPrf.Attributes, result.Attributes)
		}
	case action == Move:
		if err := mig.dmIN.SetAttributeProfile(attrPrf, true); err != nil {
			t.Error("Error when setting AttributeProfile ", err.Error())
		}
		currentVersion := engine.CurrentDataDBVersions()
		err := mig.dmOut.DataDB().SetVersions(currentVersion, false)
		if err != nil {
			t.Error("Error when setting version for stats ", err.Error())
		}
		err, _ = mig.Migrate([]string{utils.MetaAttributes})
		if err != nil {
			t.Error("Error when migrating AttributeProfile ", err.Error())
		}
		result, err := mig.dmOut.GetAttributeProfile(attrPrf.Tenant, attrPrf.ID, true, utils.NonTransactional)
		if err != nil {
			t.Error("Error when getting Stats ", err.Error())
		}
		if !reflect.DeepEqual(attrPrf.Tenant, result.Tenant) {
			t.Errorf("Expecting: %+v, received: %+v", attrPrf.Tenant, result.Tenant)
		} else if !reflect.DeepEqual(attrPrf.ID, result.ID) {
			t.Errorf("Expecting: %+v, received: %+v", attrPrf.ID, result.ID)
		} else if !reflect.DeepEqual(attrPrf.FilterIDs, result.FilterIDs) {
			t.Errorf("Expecting: %+v, received: %+v", attrPrf.FilterIDs, result.FilterIDs)
		} else if !reflect.DeepEqual(attrPrf.Contexts, result.Contexts) {
			t.Errorf("Expecting: %+v, received: %+v", attrPrf.Contexts, result.Contexts)
		} else if !reflect.DeepEqual(attrPrf.ActivationInterval, result.ActivationInterval) {
			t.Errorf("Expecting: %+v, received: %+v", attrPrf.ActivationInterval, result.ActivationInterval)
		} else if !reflect.DeepEqual(attrPrf.Attributes, result.Attributes) {
			t.Errorf("Expecting: %+v, received: %+v", attrPrf.Attributes, result.Attributes)
		}
	}
}

//TP TESTS
func testMigratorTPRatingProfile(t *testing.T) {
	tpRatingProfile := []*utils.TPRatingProfile{
		&utils.TPRatingProfile{
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
		},
	}
	switch action {
	case Move:
		if err := mig.InStorDB().SetTPRatingProfiles(tpRatingProfile); err != nil {
			t.Error("Error when setting Stats ", err.Error())
		}
		currentVersion := engine.CurrentStorDBVersions()
		err := mig.OutStorDB().SetVersions(currentVersion, false)
		if err != nil {
			t.Error("Error when setting version for stats ", err.Error())
		}
		err, _ = mig.Migrate([]string{utils.MetaTpRatingProfiles})
		if err != nil {
			t.Error("Error when migrating Stats ", err.Error())
		}
		result, err := mig.OutStorDB().GetTPRatingProfiles(tpRatingProfile[0])
		if err != nil {
			t.Error("Error when getting Stats ", err.Error())
		}
		if !reflect.DeepEqual(tpRatingProfile[0], result[0]) {
			t.Errorf("Expecting: %+v, received: %+v", tpRatingProfile[0], result[0])
		}
	}
}

func testMigratorTPSuppliers(t *testing.T) {
	tpSplPr := []*utils.TPSupplierProfile{
		&utils.TPSupplierProfile{
			TPid:      "SupplierTPID12",
			Tenant:    "cgrates.org",
			ID:        "SUPL_1",
			FilterIDs: []string{"FLTR_ACNT_dan"},
			ActivationInterval: &utils.TPActivationInterval{
				ActivationTime: "2014-07-29T15:00:00Z",
				ExpiryTime:     "",
			},
			Sorting:           "*lowest_cost",
			SortingParameters: []string{"Parameter1"},
			Suppliers: []*utils.TPSupplier{
				&utils.TPSupplier{
					ID:            "supplier1",
					AccountIDs:    []string{"Account1"},
					FilterIDs:     []string{"FLTR_1"},
					RatingPlanIDs: []string{"RPL_1"},
					ResourceIDs:   []string{"ResGroup1"},
					StatIDs:       []string{"Stat1"},
					Weight:        10,
					Blocker:       false,
				},
			},
			Weight: 20,
		},
	}
	switch action {
	case Move:
		if err := mig.InStorDB().SetTPSuppliers(tpSplPr); err != nil {
			t.Error("Error when setting TpSupplier ", err.Error())
		}
		currentVersion := engine.CurrentStorDBVersions()
		err := mig.OutStorDB().SetVersions(currentVersion, false)
		if err != nil {
			t.Error("Error when setting version for TpSupplier ", err.Error())
		}
		err, _ = mig.Migrate([]string{utils.MetaTpSuppliers})
		if err != nil {
			t.Error("Error when migrating TpSupplier ", err.Error())
		}
		result, err := mig.OutStorDB().GetTPSuppliers(tpSplPr[0].TPid, tpSplPr[0].ID)
		if err != nil {
			t.Error("Error when getting TPSupplier ", err.Error())
		}
		if !reflect.DeepEqual(tpSplPr, result) {
			t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(tpSplPr), utils.ToJSON(result))
		}
	}
}

func testMigratorTPActions(t *testing.T) {
	tpActions := []*utils.TPActions{
		&utils.TPActions{
			TPid: "TPAcc",
			ID:   "ID",
			Actions: []*utils.TPAction{
				&utils.TPAction{
					Identifier:      "*topup_reset",
					BalanceId:       "BalID",
					BalanceType:     "*data",
					Directions:      "*out",
					Units:           "10",
					ExpiryTime:      "*unlimited",
					Filter:          "",
					TimingTags:      "2014-01-14T00:00:00Z",
					DestinationIds:  "DST_1002",
					RatingSubject:   "SPECIAL_1002",
					Categories:      "",
					SharedGroups:    "SHARED_A",
					BalanceWeight:   "10",
					ExtraParameters: "",
					BalanceBlocker:  "false",
					BalanceDisabled: "false",
					Weight:          10,
				},
				&utils.TPAction{
					Identifier:      "*log",
					BalanceId:       "BalID",
					BalanceType:     "*monetary",
					Directions:      "*out",
					Units:           "120",
					ExpiryTime:      "*unlimited",
					Filter:          "",
					TimingTags:      "2014-01-14T00:00:00Z",
					DestinationIds:  "*any",
					RatingSubject:   "SPECIAL_1002",
					Categories:      "",
					SharedGroups:    "SHARED_A",
					BalanceWeight:   "11",
					ExtraParameters: "",
					BalanceBlocker:  "false",
					BalanceDisabled: "false",
					Weight:          11,
				},
			},
		},
	}
	switch action {
	case Move:
		if err := mig.InStorDB().SetTPActions(tpActions); err != nil {
			t.Error("Error when setting TpActions ", err.Error())
		}
		currentVersion := engine.CurrentStorDBVersions()
		err := mig.OutStorDB().SetVersions(currentVersion, false)
		if err != nil {
			t.Error("Error when setting version for TpActions ", err.Error())
		}
		err, _ = mig.Migrate([]string{utils.MetaTpActions})
		if err != nil {
			t.Error("Error when migrating TpActions ", err.Error())
		}
		result, err := mig.OutStorDB().GetTPActions(tpActions[0].TPid, tpActions[0].ID)
		if err != nil {
			t.Error("Error when getting TpActions ", err.Error())
		}
		if !reflect.DeepEqual(tpActions[0].TPid, result[0].TPid) {
			t.Errorf("Expecting: %+v, received: %+v", tpActions[0].TPid, result[0].TPid)
		} else if !reflect.DeepEqual(tpActions[0].ID, result[0].ID) {
			t.Errorf("Expecting: %+v, received: %+v", tpActions[0].ID, result[0].ID)
		} else if !reflect.DeepEqual(tpActions[0].Actions[0], result[0].Actions[0]) &&
			!reflect.DeepEqual(tpActions[0].Actions[0], result[0].Actions[1]) {
			t.Errorf("Expecting: %+v, received: %+v", tpActions[0].Actions[0], result[0].Actions[0])
		} else if !reflect.DeepEqual(tpActions[0].Actions[1], result[0].Actions[1]) &&
			!reflect.DeepEqual(tpActions[0].Actions[1], result[0].Actions[0]) {
			t.Errorf("Expecting: %+v, received: %+v", tpActions[0].Actions[1], result[0].Actions[1])
		}
	}
}

func testMigratorTPAccountActions(t *testing.T) {
	tpAccActions := []*utils.TPAccountActions{
		&utils.TPAccountActions{
			TPid:             "TPAcc",
			LoadId:           "ID",
			Tenant:           "cgrates.org",
			Account:          "1001",
			ActionPlanId:     "PREPAID_10",
			ActionTriggersId: "STANDARD_TRIGGERS",
			AllowNegative:    true,
			Disabled:         false,
		},
	}
	switch action {
	case Move:
		if err := mig.InStorDB().SetTPAccountActions(tpAccActions); err != nil {
			t.Error("Error when setting TpAccountActions ", err.Error())
		}
		currentVersion := engine.CurrentStorDBVersions()
		err := mig.OutStorDB().SetVersions(currentVersion, false)
		if err != nil {
			t.Error("Error when setting version for TpAccountActions ", err.Error())
		}
		err, _ = mig.Migrate([]string{utils.MetaTpAccountActions})
		if err != nil {
			t.Error("Error when migrating TpAccountActions ", err.Error())
		}
		result, err := mig.OutStorDB().GetTPAccountActions(&utils.TPAccountActions{TPid: "TPAcc"})
		if err != nil {
			t.Error("Error when getting TpAccountActions ", err.Error())
		}
		if !reflect.DeepEqual(tpAccActions[0], result[0]) {
			t.Errorf("Expecting: %+v, received: %+v", tpAccActions[0], result[0])
		}
	}
}

func testMigratorTpActionTriggers(t *testing.T) {
	tpActionTriggers := []*utils.TPActionTriggers{
		&utils.TPActionTriggers{
			TPid: "TPAct",
			ID:   "STANDARD_TRIGGERS",
			ActionTriggers: []*utils.TPActionTrigger{
				&utils.TPActionTrigger{
					Id:                    "STANDARD_TRIGGERS",
					UniqueID:              "",
					ThresholdType:         "*min_balance",
					ThresholdValue:        2,
					Recurrent:             false,
					MinSleep:              "0",
					ExpirationDate:        "",
					ActivationDate:        "",
					BalanceId:             "",
					BalanceType:           "*monetary",
					BalanceDirections:     "*out",
					BalanceDestinationIds: "FS_USERS",
					BalanceWeight:         "",
					BalanceExpirationDate: "",
					BalanceTimingTags:     "",
					BalanceRatingSubject:  "",
					BalanceCategories:     "",
					BalanceSharedGroups:   "",
					BalanceBlocker:        "",
					BalanceDisabled:       "",
					MinQueuedItems:        3,
					ActionsId:             "LOG_WARNING",
					Weight:                10,
				},
				&utils.TPActionTrigger{
					Id:                    "STANDARD_TRIGGERS",
					UniqueID:              "",
					ThresholdType:         "*max_event_counter",
					ThresholdValue:        5,
					Recurrent:             false,
					MinSleep:              "0",
					ExpirationDate:        "",
					ActivationDate:        "",
					BalanceId:             "",
					BalanceType:           "*monetary",
					BalanceDirections:     "*out",
					BalanceDestinationIds: "FS_USERS",
					BalanceWeight:         "",
					BalanceExpirationDate: "",
					BalanceTimingTags:     "",
					BalanceRatingSubject:  "",
					BalanceCategories:     "",
					BalanceSharedGroups:   "",
					BalanceBlocker:        "",
					BalanceDisabled:       "",
					MinQueuedItems:        3,
					ActionsId:             "LOG_WARNING",
					Weight:                10,
				},
			},
		},
	}
	switch action {
	case Move:
		if err := mig.InStorDB().SetTPActionTriggers(tpActionTriggers); err != nil {
			t.Error("Error when setting TpActionTriggers ", err.Error())
		}
		currentVersion := engine.CurrentStorDBVersions()
		err := mig.OutStorDB().SetVersions(currentVersion, false)
		if err != nil {
			t.Error("Error when setting version for TpActionTriggers ", err.Error())
		}
		err, _ = mig.Migrate([]string{utils.MetaTpActionTriggers})
		if err != nil {
			t.Error("Error when migrating TpActionTriggers ", err.Error())
		}
		result, err := mig.OutStorDB().GetTPActionTriggers(tpActionTriggers[0].TPid, tpActionTriggers[0].ID)
		if err != nil {
			t.Error("Error when getting TpAccountActions ", err.Error())
		}
		if !reflect.DeepEqual(tpActionTriggers[0].TPid, result[0].TPid) {
			t.Errorf("Expecting: %+v, received: %+v", tpActionTriggers[0].TPid, result[0].TPid)
		} else if !reflect.DeepEqual(tpActionTriggers[0].ID, result[0].ID) {
			t.Errorf("Expecting: %+v, received: %+v", tpActionTriggers[0].ID, result[0].ID)
		} else if !reflect.DeepEqual(tpActionTriggers[0].ActionTriggers[0], result[0].ActionTriggers[0]) &&
			!reflect.DeepEqual(tpActionTriggers[0].ActionTriggers[0], result[0].ActionTriggers[1]) {
			t.Errorf("Expecting: %+v, received: %+v", tpActionTriggers[0].ActionTriggers[0], result[0].ActionTriggers[0])
		} else if !reflect.DeepEqual(tpActionTriggers[0].ActionTriggers[1], result[0].ActionTriggers[1]) &&
			!reflect.DeepEqual(tpActionTriggers[0].ActionTriggers[1], result[0].ActionTriggers[0]) {
			t.Errorf("Expecting: %+v, received: %+v", tpActionTriggers[0].ActionTriggers[1], result[0].ActionTriggers[1])
		}
	}
}

func testMigratorTpActionPlans(t *testing.T) {
	tpAccPlan := []*utils.TPActionPlan{
		&utils.TPActionPlan{
			TPid: "TPAcc",
			ID:   "ID",
			ActionPlan: []*utils.TPActionTiming{
				&utils.TPActionTiming{
					ActionsId: "AccId",
					TimingId:  "TimingID",
					Weight:    10,
				},
				&utils.TPActionTiming{
					ActionsId: "AccId2",
					TimingId:  "TimingID2",
					Weight:    11,
				},
			},
		},
	}
	switch action {
	case Move:
		if err := mig.InStorDB().SetTPActionPlans(tpAccPlan); err != nil {
			t.Error("Error when setting TpActionPlans ", err.Error())
		}
		currentVersion := engine.CurrentStorDBVersions()
		err := mig.OutStorDB().SetVersions(currentVersion, false)
		if err != nil {
			t.Error("Error when setting version for TpActionPlans ", err.Error())
		}
		err, _ = mig.Migrate([]string{utils.MetaTpActionPlans})
		if err != nil {
			t.Error("Error when migrating TpActionPlans ", err.Error())
		}
		result, err := mig.OutStorDB().GetTPActionPlans(tpAccPlan[0].TPid, tpAccPlan[0].ID)
		if err != nil {
			t.Error("Error when getting TpActionPlans ", err.Error())
		}
		if !reflect.DeepEqual(tpAccPlan[0], result[0]) {
			t.Errorf("Expecting: %+v, received: %+v", tpAccPlan[0], result[0])
		}
	}
}

func testMigratorTpUsers(t *testing.T) {
	tpUser := []*utils.TPUsers{
		&utils.TPUsers{
			TPid:     "TPU1",
			UserName: "User1",
			Tenant:   "Tenant1",
			Masked:   true,
			Weight:   20,
			Profile: []*utils.TPUserProfile{
				&utils.TPUserProfile{
					AttrName:  "UserProfile1",
					AttrValue: "ValUP1",
				},
				&utils.TPUserProfile{
					AttrName:  "UserProfile2",
					AttrValue: "ValUP2",
				},
			},
		},
	}
	switch action {
	case Move:
		if err := mig.InStorDB().SetTPUsers(tpUser); err != nil {
			t.Error("Error when setting TpUsers ", err.Error())
		}
		currentVersion := engine.CurrentStorDBVersions()
		err := mig.OutStorDB().SetVersions(currentVersion, false)
		if err != nil {
			t.Error("Error when setting version for TpUsers ", err.Error())
		}
		err, _ = mig.Migrate([]string{utils.MetaTpUsers})
		if err != nil {
			t.Error("Error when migrating TpUsers ", err.Error())
		}
		result, err := mig.OutStorDB().GetTPUsers(&utils.TPUsers{TPid: tpUser[0].TPid})
		if err != nil {
			t.Error("Error when getting TpUsers ", err.Error())
		}
		if !reflect.DeepEqual(tpUser[0], result[0]) {
			t.Errorf("Expecting: %+v, received: %+v", tpUser[0], result[0])
		}
	}
}

func testMigratorTpTimings(t *testing.T) {
	tpTiming := []*utils.ApierTPTiming{&utils.ApierTPTiming{
		TPid:      "TPT1",
		ID:        "Timing",
		Years:     "2017",
		Months:    "05",
		MonthDays: "01",
		WeekDays:  "1",
		Time:      "15:00:00Z",
	},
	}
	switch action {
	case Move:
		if err := mig.InStorDB().SetTPTimings(tpTiming); err != nil {
			t.Error("Error when setting TpTiming ", err.Error())
		}
		currentVersion := engine.CurrentStorDBVersions()
		err := mig.OutStorDB().SetVersions(currentVersion, false)
		if err != nil {
			t.Error("Error when setting version for TpTiming ", err.Error())
		}
		err, _ = mig.Migrate([]string{utils.MetaTpTiming})
		if err != nil {
			t.Error("Error when migrating TpTiming ", err.Error())
		}
		result, err := mig.OutStorDB().GetTPTimings(tpTiming[0].TPid, tpTiming[0].ID)
		if err != nil {
			t.Error("Error when getting TpTiming ", err.Error())
		}
		if !reflect.DeepEqual(tpTiming[0], result[0]) {
			t.Errorf("Expecting: %+v, received: %+v", tpTiming[0], result[0])
		}
	}
}

func testMigratorTpThreshold(t *testing.T) {
	tpThreshold := []*utils.TPThreshold{
		&utils.TPThreshold{
			TPid:      "TH1",
			Tenant:    "cgrates.org",
			ID:        "Threhold",
			FilterIDs: []string{"FLTR_1", "FLTR_2"},
			ActivationInterval: &utils.TPActivationInterval{
				ActivationTime: "2014-07-29T15:00:00Z",
				ExpiryTime:     "",
			},
			Recurrent: true,
			MinSleep:  "1s",
			Blocker:   true,
			Weight:    10,
			ActionIDs: []string{"Thresh1", "Thresh2"},
			Async:     true,
		},
	}
	switch action {
	case Move:
		if err := mig.InStorDB().SetTPThresholds(tpThreshold); err != nil {
			t.Error("Error when setting TpThreshold ", err.Error())
		}
		currentVersion := engine.CurrentStorDBVersions()
		err := mig.OutStorDB().SetVersions(currentVersion, false)
		if err != nil {
			t.Error("Error when setting version for TpThreshold ", err.Error())
		}
		err, _ = mig.Migrate([]string{utils.MetaTpThresholds})
		if err != nil {
			t.Error("Error when migrating TpThreshold ", err.Error())
		}
		result, err := mig.OutStorDB().GetTPThresholds(tpThreshold[0].TPid, tpThreshold[0].ID)
		if err != nil {
			t.Error("Error when getting TpThreshold ", err.Error())
		}
		if !reflect.DeepEqual(tpThreshold[0], result[0]) {
			t.Errorf("Expecting: %+v, received: %+v", tpThreshold[0], result[0])
		}
	}
}

func testMigratorTpStats(t *testing.T) {
	tpStat := []*utils.TPStats{
		&utils.TPStats{
			Tenant:    "cgrates.org",
			TPid:      "TPS1",
			ID:        "Stat1",
			FilterIDs: []string{"FLTR_1"},
			ActivationInterval: &utils.TPActivationInterval{
				ActivationTime: "2014-07-29T15:00:00Z",
				ExpiryTime:     "",
			},
			TTL: "1",
			Metrics: []*utils.MetricWithParams{
				&utils.MetricWithParams{MetricID: "MetricValue", Parameters: ""},
				&utils.MetricWithParams{MetricID: "MetricValueTwo", Parameters: ""},
			},
			Blocker:      false,
			Stored:       false,
			Weight:       20,
			MinItems:     1,
			ThresholdIDs: []string{"ThreshValue", "ThreshValueTwo"},
		},
	}
	switch action {
	case Move:
		if err := mig.InStorDB().SetTPStats(tpStat); err != nil {
			t.Error("Error when setting TpStats ", err.Error())
		}
		currentVersion := engine.CurrentStorDBVersions()
		err := mig.OutStorDB().SetVersions(currentVersion, false)
		if err != nil {
			t.Error("Error when setting version for TpStats ", err.Error())
		}
		err, _ = mig.Migrate([]string{utils.MetaTpStats})
		if err != nil {
			t.Error("Error when migrating TpStats ", err.Error())
		}
		result, err := mig.OutStorDB().GetTPStats(tpStat[0].TPid, tpStat[0].ID)
		if err != nil {
			t.Error("Error when getting TpStats ", err.Error())
		}
		if !reflect.DeepEqual(tpStat[0], result[0]) {
			t.Errorf("Expecting: %+v, received: %+v", tpStat[0], result[0])
		}
	}
}

func testMigratorTpSharedGroups(t *testing.T) {
	tpSharedGroups := []*utils.TPSharedGroups{
		&utils.TPSharedGroups{
			TPid: "Tpi",
			ID:   "TpSg",
			SharedGroups: []*utils.TPSharedGroup{
				&utils.TPSharedGroup{
					Account:       "AccOne",
					Strategy:      "StrategyOne",
					RatingSubject: "SubOne",
				},
				&utils.TPSharedGroup{
					Account:       "AccTow",
					Strategy:      "StrategyTwo",
					RatingSubject: "SubTwo",
				},
				&utils.TPSharedGroup{
					Account:       "AccPlus",
					Strategy:      "StrategyPlus",
					RatingSubject: "SubPlus",
				},
			},
		},
	}
	switch action {
	case Move:
		if err := mig.InStorDB().SetTPSharedGroups(tpSharedGroups); err != nil {
			t.Error("Error when setting TpSharedGroups ", err.Error())
		}
		currentVersion := engine.CurrentStorDBVersions()
		err := mig.OutStorDB().SetVersions(currentVersion, false)
		if err != nil {
			t.Error("Error when setting version for TpSharedGroups ", err.Error())
		}
		err, _ = mig.Migrate([]string{utils.MetaTpSharedGroups})
		if err != nil {
			t.Error("Error when migrating TpSharedGroups ", err.Error())
		}
		result, err := mig.OutStorDB().GetTPSharedGroups(tpSharedGroups[0].TPid, tpSharedGroups[0].ID)
		if err != nil {
			t.Error("Error when getting TpSharedGroups ", err.Error())
		}
		if !reflect.DeepEqual(tpSharedGroups[0], result[0]) {
			t.Errorf("Expecting: %+v, received: %+v", tpSharedGroups[0], result[0])
		}
	}
}

func testMigratorTpResources(t *testing.T) {
	tpRes := []*utils.TPResource{
		&utils.TPResource{
			Tenant:    "cgrates.org",
			TPid:      "TPR1",
			ID:        "ResGroup1",
			FilterIDs: []string{"FLTR_1"},
			ActivationInterval: &utils.TPActivationInterval{
				ActivationTime: "2014-07-29T15:00:00Z",
				ExpiryTime:     "",
			},
			UsageTTL:          "1s",
			Limit:             "7",
			AllocationMessage: "",
			Blocker:           true,
			Stored:            true,
			Weight:            20,
			ThresholdIDs:      []string{"ValOne", "ValTwo"},
		},
	}
	switch action {
	case Move:
		if err := mig.InStorDB().SetTPResources(tpRes); err != nil {
			t.Error("Error when setting TpResources ", err.Error())
		}
		currentVersion := engine.CurrentStorDBVersions()
		err := mig.OutStorDB().SetVersions(currentVersion, false)
		if err != nil {
			t.Error("Error when setting version for TpResources ", err.Error())
		}
		err, _ = mig.Migrate([]string{utils.MetaTpResources})
		if err != nil {
			t.Error("Error when migrating TpResources ", err.Error())
		}
		result, err := mig.OutStorDB().GetTPResources(tpRes[0].TPid, tpRes[0].ID)
		if err != nil {
			t.Error("Error when getting TpResources ", err.Error())
		}
		if !reflect.DeepEqual(tpRes[0], result[0]) {
			t.Errorf("Expecting: %+v, received: %+v", tpRes[0], result[0])
		}
	}
}

func testMigratorTpRatingProfiles(t *testing.T) {
	tpRatingProfile := []*utils.TPRatingProfile{
		&utils.TPRatingProfile{
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
		},
	}
	switch action {
	case Move:
		if err := mig.InStorDB().SetTPRatingProfiles(tpRatingProfile); err != nil {
			t.Error("Error when setting TpRatingProfiles ", err.Error())
		}
		currentVersion := engine.CurrentStorDBVersions()
		err := mig.OutStorDB().SetVersions(currentVersion, false)
		if err != nil {
			t.Error("Error when setting version for TpRatingProfiles ", err.Error())
		}
		err, _ = mig.Migrate([]string{utils.MetaTpRatingProfiles})
		if err != nil {
			t.Error("Error when migrating TpRatingProfiles ", err.Error())
		}
		result, err := mig.OutStorDB().GetTPRatingProfiles(&utils.TPRatingProfile{TPid: tpRatingProfile[0].TPid})
		if err != nil {
			t.Error("Error when getting TpRatingProfiles ", err.Error())
		}
		if !reflect.DeepEqual(tpRatingProfile[0], result[0]) {
			t.Errorf("Expecting: %+v, received: %+v", tpRatingProfile[0], result[0])
		}
	}
}

func testMigratorTpRatingPlans(t *testing.T) {
	tpRatingPlan := []*utils.TPRatingPlan{
		&utils.TPRatingPlan{
			TPid: "TPRP1",
			ID:   "Plan1",
			RatingPlanBindings: []*utils.TPRatingPlanBinding{
				&utils.TPRatingPlanBinding{
					DestinationRatesId: "RateId",
					TimingId:           "TimingID",
					Weight:             12,
				},
				&utils.TPRatingPlanBinding{
					DestinationRatesId: "DR_FREESWITCH_USERS",
					TimingId:           "ALWAYS",
					Weight:             10,
				},
			},
		},
	}
	switch action {
	case Move:
		if err := mig.InStorDB().SetTPRatingPlans(tpRatingPlan); err != nil {
			t.Error("Error when setting TpRatingPlans ", err.Error())
		}
		currentVersion := engine.CurrentStorDBVersions()
		err := mig.OutStorDB().SetVersions(currentVersion, false)
		if err != nil {
			t.Error("Error when setting version for TpRatingPlans ", err.Error())
		}
		err, _ = mig.Migrate([]string{utils.MetaTpRatingPlans})
		if err != nil {
			t.Error("Error when migrating TpRatingPlans ", err.Error())
		}
		result, err := mig.OutStorDB().GetTPRatingPlans("TPRP1", "Plan1", nil)
		if err != nil {
			t.Error("Error when getting TpRatingPlans ", err.Error())
		}
		if !reflect.DeepEqual(tpRatingPlan[0].TPid, result[0].TPid) {
			t.Errorf("Expecting: %+v, received: %+v", tpRatingPlan[0].TPid, result[0].TPid)
		} else if !reflect.DeepEqual(tpRatingPlan[0].ID, result[0].ID) {
			t.Errorf("Expecting: %+v, received: %+v", tpRatingPlan[0].ID, result[0].ID)
		} else if !reflect.DeepEqual(tpRatingPlan[0].RatingPlanBindings[0], result[0].RatingPlanBindings[0]) &&
			!reflect.DeepEqual(tpRatingPlan[0].RatingPlanBindings[1], result[0].RatingPlanBindings[1]) &&
			!reflect.DeepEqual(tpRatingPlan[0].RatingPlanBindings[1], result[0].RatingPlanBindings[0]) &&
			!reflect.DeepEqual(tpRatingPlan[0].RatingPlanBindings[0], result[0].RatingPlanBindings[1]) {
			t.Errorf("Expecting: %+v, received: %+v", tpRatingPlan[0].RatingPlanBindings[0], result[0].RatingPlanBindings[0])
		} else if !reflect.DeepEqual(tpRatingPlan[0].RatingPlanBindings[0], result[0].RatingPlanBindings[0]) &&
			!reflect.DeepEqual(tpRatingPlan[0].RatingPlanBindings[1], result[0].RatingPlanBindings[1]) &&
			!reflect.DeepEqual(tpRatingPlan[0].RatingPlanBindings[1], result[0].RatingPlanBindings[0]) &&
			!reflect.DeepEqual(tpRatingPlan[0].RatingPlanBindings[0], result[0].RatingPlanBindings[1]) {
			t.Errorf("Expecting: %+v, received: %+v", tpRatingPlan[0].RatingPlanBindings[1], result[0].RatingPlanBindings[1])
		}

	}
}

func testMigratorTpRates(t *testing.T) {
	tpRate := []*utils.TPRate{
		&utils.TPRate{
			TPid: "TPidTpRate",
			ID:   "RT_FS_USERS",
			RateSlots: []*utils.RateSlot{
				&utils.RateSlot{
					ConnectFee:         12,
					Rate:               3,
					RateUnit:           "6s",
					RateIncrement:      "6s",
					GroupIntervalStart: "0s",
				},
				&utils.RateSlot{
					ConnectFee:         12,
					Rate:               3,
					RateUnit:           "4s",
					RateIncrement:      "6s",
					GroupIntervalStart: "1s",
				},
			},
		},
	}
	switch action {
	case Move:
		if err := mig.InStorDB().SetTPRates(tpRate); err != nil {
			t.Error("Error when setting TpRates ", err.Error())
		}
		currentVersion := engine.CurrentStorDBVersions()
		err := mig.OutStorDB().SetVersions(currentVersion, false)
		if err != nil {
			t.Error("Error when setting version for TpRates ", err.Error())
		}
		err, _ = mig.Migrate([]string{utils.MetaTpRates})
		if err != nil {
			t.Error("Error when migrating TpRates ", err.Error())
		}
		result, err := mig.OutStorDB().GetTPRates(tpRate[0].TPid, tpRate[0].ID)
		if err != nil {
			t.Error("Error when getting TpRates ", err.Error())
		}
		if !reflect.DeepEqual(tpRate[0].TPid, result[0].TPid) {
			t.Errorf("Expecting: %+v, received: %+v", tpRate[0].TPid, result[0].TPid)
		} else if !reflect.DeepEqual(tpRate[0].ID, result[0].ID) {
			t.Errorf("Expecting: %+v, received: %+v", tpRate[0].ID, result[0].ID)
		}
		if !reflect.DeepEqual(tpRate[0].RateSlots[0].ConnectFee, result[0].RateSlots[0].ConnectFee) {
			t.Errorf("Expecting: %+v, received: %+v", tpRate[0].RateSlots[0].ConnectFee, result[0].RateSlots[0].ConnectFee)
		} else if !reflect.DeepEqual(tpRate[0].RateSlots[0].Rate, result[0].RateSlots[0].Rate) {
			t.Errorf("Expecting: %+v, received: %+v", tpRate[0].RateSlots[0].Rate, result[0].RateSlots[0].Rate)
		} else if !reflect.DeepEqual(tpRate[0].RateSlots[0].RateUnit, result[0].RateSlots[0].RateUnit) {
			t.Errorf("Expecting: %+v, received: %+v", tpRate[0].RateSlots[0].RateUnit, result[0].RateSlots[0].RateUnit)
		} else if !reflect.DeepEqual(tpRate[0].RateSlots[0].RateIncrement, result[0].RateSlots[0].RateIncrement) {
			t.Errorf("Expecting: %+v, received: %+v", tpRate[0].RateSlots[0].RateIncrement, result[0].RateSlots[0].RateIncrement)
		} else if !reflect.DeepEqual(tpRate[0].RateSlots[0].GroupIntervalStart, result[0].RateSlots[0].GroupIntervalStart) {
			t.Errorf("Expecting: %+v, received: %+v", tpRate[0].RateSlots[0].GroupIntervalStart, result[0].RateSlots[0].GroupIntervalStart)
		}
		if !reflect.DeepEqual(tpRate[0].RateSlots[1].ConnectFee, result[0].RateSlots[1].ConnectFee) {
			t.Errorf("Expecting: %+v, received: %+v", tpRate[0].RateSlots[1].ConnectFee, result[0].RateSlots[1].ConnectFee)
		} else if !reflect.DeepEqual(tpRate[0].RateSlots[1].Rate, result[0].RateSlots[1].Rate) {
			t.Errorf("Expecting: %+v, received: %+v", tpRate[0].RateSlots[1].Rate, result[0].RateSlots[1].Rate)
		} else if !reflect.DeepEqual(tpRate[0].RateSlots[1].RateUnit, result[0].RateSlots[1].RateUnit) {
			t.Errorf("Expecting: %+v, received: %+v", tpRate[0].RateSlots[1].RateUnit, result[0].RateSlots[1].RateUnit)
		} else if !reflect.DeepEqual(tpRate[0].RateSlots[1].RateIncrement, result[0].RateSlots[1].RateIncrement) {
			t.Errorf("Expecting: %+v, received: %+v", tpRate[0].RateSlots[1].RateIncrement, result[0].RateSlots[1].RateIncrement)
		} else if !reflect.DeepEqual(tpRate[0].RateSlots[1].GroupIntervalStart, result[0].RateSlots[1].GroupIntervalStart) {
			t.Errorf("Expecting: %+v, received: %+v", tpRate[0].RateSlots[1].GroupIntervalStart, result[0].RateSlots[1].GroupIntervalStart)
		}
	}
}

func testMigratorTpFilter(t *testing.T) {
	tpFilter := []*utils.TPFilterProfile{
		&utils.TPFilterProfile{
			TPid:   "TP1",
			Tenant: "cgrates.org",
			ID:     "Filter",
			Filters: []*utils.TPFilter{
				&utils.TPFilter{
					Type:      "*string",
					FieldName: "Account",
					Values:    []string{"1001", "1002"},
				},
			},
			ActivationInterval: &utils.TPActivationInterval{
				ActivationTime: "2014-07-29T15:00:00Z",
				ExpiryTime:     "",
			},
		},
	}
	switch action {
	case Move:
		if err := mig.InStorDB().SetTPFilters(tpFilter); err != nil {
			t.Error("Error when setting TpFilter ", err.Error())
		}
		currentVersion := engine.CurrentStorDBVersions()
		err := mig.OutStorDB().SetVersions(currentVersion, false)
		if err != nil {
			t.Error("Error when setting version for TpFilter ", err.Error())
		}
		err, _ = mig.Migrate([]string{utils.MetaTpFilters})
		if err != nil {
			t.Error("Error when migrating TpFilter ", err.Error())
		}
		result, err := mig.OutStorDB().GetTPFilters(tpFilter[0].TPid, tpFilter[0].ID)
		if err != nil {
			t.Error("Error when getting TpFilter ", err.Error())
		}
		if !reflect.DeepEqual(tpFilter[0].TPid, result[0].TPid) {
			t.Errorf("Expecting: %+v, received: %+v", tpFilter[0].TPid, result[0].TPid)
		} else if !reflect.DeepEqual(tpFilter[0].ID, result[0].ID) {
			t.Errorf("Expecting: %+v, received: %+v", tpFilter[0].ID, result[0].ID)
		} else if !reflect.DeepEqual(tpFilter[0].Filters, result[0].Filters) {
			t.Errorf("Expecting: %+v, received: %+v", tpFilter[0].Filters, result[0].Filters)
		} else if !reflect.DeepEqual(tpFilter[0].ActivationInterval, result[0].ActivationInterval) {
			t.Errorf("Expecting: %+v, received: %+v", tpFilter[0].ActivationInterval, result[0].ActivationInterval)
		}
	}
}

func testMigratorTpDestination(t *testing.T) {
	tpDestination := []*utils.TPDestination{
		&utils.TPDestination{
			TPid:     "TPD",
			ID:       "GERMANY",
			Prefixes: []string{"+49", "+4915"},
		},
	}
	switch action {
	case Move:
		if err := mig.InStorDB().SetTPDestinations(tpDestination); err != nil {
			t.Error("Error when setting TpDestination ", err.Error())
		}
		currentVersion := engine.CurrentStorDBVersions()
		err := mig.OutStorDB().SetVersions(currentVersion, false)
		if err != nil {
			t.Error("Error when setting version for TpDestination ", err.Error())
		}
		err, _ = mig.Migrate([]string{utils.MetaTpDestinations})
		if err != nil {
			t.Error("Error when migrating TpDestination ", err.Error())
		}
		result, err := mig.OutStorDB().GetTPDestinations(tpDestination[0].TPid, tpDestination[0].ID)
		if err != nil {
			t.Error("Error when getting TpDestination ", err.Error())
		}
		if !reflect.DeepEqual(tpDestination[0], result[0]) {
			t.Errorf("Expecting: %+v, received: %+v", tpDestination[0], result[0])
		}
	}
}

func testMigratorTpDestinationRate(t *testing.T) {
	tpDestRate := []*utils.TPDestinationRate{
		&utils.TPDestinationRate{
			TPid: "testTPid",
			ID:   "1",
			DestinationRates: []*utils.DestinationRate{
				&utils.DestinationRate{
					DestinationId:    "GERMANY",
					RateId:           "RT_1CENT",
					RoundingMethod:   "*up",
					RoundingDecimals: 0,
					MaxCost:          0.0,
					MaxCostStrategy:  "",
				},
			},
		},
	}

	switch action {
	case Move:
		if err := mig.InStorDB().SetTPDestinationRates(tpDestRate); err != nil {
			t.Error("Error when setting TpDestinationRate ", err.Error())
		}
		currentVersion := engine.CurrentStorDBVersions()
		err := mig.OutStorDB().SetVersions(currentVersion, false)
		if err != nil {
			t.Error("Error when setting version for TpDestinationRate ", err.Error())
		}
		err, _ = mig.Migrate([]string{utils.MetaTpDestinationRates})
		if err != nil {
			t.Error("Error when migrating TpDestinationRate ", err.Error())
		} //OutStorDB
		result, err := mig.InStorDB().GetTPDestinationRates("testTPid", "", nil)
		if err != nil {
			t.Error("Error when getting TpDestinationRate ", err.Error())
		}
		if !reflect.DeepEqual(tpDestRate[0].TPid, result[0].TPid) {
			t.Errorf("Expecting: %+v, received: %+v", tpDestRate[0].TPid, result[0].TPid)
		} else if !reflect.DeepEqual(tpDestRate[0].ID, result[0].ID) {
			t.Errorf("Expecting: %+v, received: %+v", tpDestRate[0].ID, result[0].ID)
		}
		if !reflect.DeepEqual(tpDestRate[0].DestinationRates[0].DestinationId, result[0].DestinationRates[0].DestinationId) {
			t.Errorf("Expecting: %+v, received: %+v", tpDestRate[0].DestinationRates[0].DestinationId, result[0].DestinationRates[0].DestinationId)
		} else if !reflect.DeepEqual(tpDestRate[0].DestinationRates[0].RateId, result[0].DestinationRates[0].RateId) {
			t.Errorf("Expecting: %+v, received: %+v", tpDestRate[0].DestinationRates[0].RateId, result[0].DestinationRates[0].RateId)
		} else if !reflect.DeepEqual(tpDestRate[0].DestinationRates[0], result[0].DestinationRates[0]) {
			t.Errorf("Expecting: %+v, received: %+v", tpDestRate[0].DestinationRates[0], result[0].DestinationRates[0])
		}
	}
}

func testMigratorTpDerivedChargers(t *testing.T) {
	tpDerivedChargers := []*utils.TPDerivedChargers{
		&utils.TPDerivedChargers{
			TPid:           "TPD",
			LoadId:         "LoadID",
			Direction:      "*out",
			Tenant:         "cgrates.org",
			Category:       "call",
			Account:        "1001",
			Subject:        "1001",
			DestinationIds: "",
			DerivedChargers: []*utils.TPDerivedCharger{
				&utils.TPDerivedCharger{
					RunId:                "derived_run1",
					RunFilters:           "",
					ReqTypeField:         "^*rated",
					DirectionField:       "*default",
					TenantField:          "*default",
					CategoryField:        "*default",
					AccountField:         "*default",
					SubjectField:         "^1002",
					DestinationField:     "*default",
					SetupTimeField:       "*default",
					PddField:             "*default",
					AnswerTimeField:      "*default",
					UsageField:           "*default",
					SupplierField:        "*default",
					DisconnectCauseField: "*default",
					CostField:            "*default",
					RatedField:           "*default",
				},
			},
		},
	}
	switch action {
	case Move:
		if err := mig.InStorDB().SetTPDerivedChargers(tpDerivedChargers); err != nil {
			t.Error("Error when setting TpDerivedChargers ", err.Error())
		}
		currentVersion := engine.CurrentStorDBVersions()
		err := mig.OutStorDB().SetVersions(currentVersion, false)
		if err != nil {
			t.Error("Error when setting version for TpDerivedChargers ", err.Error())
		}
		err, _ = mig.Migrate([]string{utils.MetaTpDerivedChargers})
		if err != nil {
			t.Error("Error when migrating TpDerivedChargers ", err.Error())
		}
		result, err := mig.OutStorDB().GetTPDerivedChargers(&utils.TPDerivedChargers{TPid: tpDerivedChargers[0].TPid})
		if err != nil {
			t.Error("Error when getting TpDerivedChargers ", err.Error())
		}
		if !reflect.DeepEqual(tpDerivedChargers[0], result[0]) {
			t.Errorf("Expecting: %+v, received: %+v", tpDerivedChargers[0], result[0])
		}
	}
}

func testMigratorTpCdrStats(t *testing.T) {
	tpCdrStats := []*utils.TPCdrStats{
		&utils.TPCdrStats{
			TPid: "TPCdr",
			ID:   "ID",
			CdrStats: []*utils.TPCdrStat{
				&utils.TPCdrStat{
					QueueLength:      "10",
					TimeWindow:       "0",
					SaveInterval:     "10s",
					Metrics:          "ASR",
					SetupInterval:    "",
					TORs:             "",
					CdrHosts:         "",
					CdrSources:       "",
					ReqTypes:         "",
					Directions:       "",
					Tenants:          "cgrates.org",
					Categories:       "",
					Accounts:         "",
					Subjects:         "1001",
					DestinationIds:   "1003",
					PddInterval:      "",
					UsageInterval:    "",
					Suppliers:        "suppl1",
					DisconnectCauses: "",
					MediationRunIds:  "*default",
					RatedAccounts:    "",
					RatedSubjects:    "",
					CostInterval:     "",
					ActionTriggers:   "CDRST1_WARN",
				},
				&utils.TPCdrStat{
					QueueLength:      "10",
					TimeWindow:       "0",
					SaveInterval:     "10s",
					Metrics:          "ACC",
					SetupInterval:    "",
					TORs:             "",
					CdrHosts:         "",
					CdrSources:       "",
					ReqTypes:         "",
					Directions:       "",
					Tenants:          "cgrates.org",
					Categories:       "",
					Accounts:         "",
					Subjects:         "1002",
					DestinationIds:   "1003",
					PddInterval:      "",
					UsageInterval:    "",
					Suppliers:        "suppl1",
					DisconnectCauses: "",
					MediationRunIds:  "*default",
					RatedAccounts:    "",
					RatedSubjects:    "",
					CostInterval:     "",
					ActionTriggers:   "CDRST1_WARN",
				},
			},
		},
	}
	switch action {
	case Move:
		if err := mig.InStorDB().SetTPCdrStats(tpCdrStats); err != nil {
			t.Error("Error when setting TpCdrStats ", err.Error())
		}
		currentVersion := engine.CurrentStorDBVersions()
		err := mig.OutStorDB().SetVersions(currentVersion, false)
		if err != nil {
			t.Error("Error when setting version for TpCdrStats ", err.Error())
		}
		err, _ = mig.Migrate([]string{utils.MetaTpCdrStats})
		if err != nil {
			t.Error("Error when migrating TpCdrStats ", err.Error())
		}
		result, err := mig.OutStorDB().GetTPCdrStats(tpCdrStats[0].TPid, tpCdrStats[0].ID)
		if err != nil {
			t.Error("Error when getting TpCdrStats ", err.Error())
		}
		if !reflect.DeepEqual(tpCdrStats[0], result[0]) {
			t.Errorf("Expecting: %+v, received: %+v", tpCdrStats[0], result[0])
		}
	}
}

func testMigratorTpAliases(t *testing.T) {
	tpAliases := []*utils.TPAliases{
		&utils.TPAliases{
			TPid:      "tpID",
			Direction: "*out",
			Tenant:    "cgrates.org",
			Category:  "call",
			Account:   "1001",
			Subject:   "1002",
			Context:   "",
			Values: []*utils.TPAliasValue{
				&utils.TPAliasValue{
					DestinationId: "1002",
					Target:        "1002",
					Original:      "1002",
					Alias:         "1002",
					Weight:        20.0,
				},
			},
		},
	}

	switch action {
	case Move:
		if err := mig.InStorDB().SetTPAliases(tpAliases); err != nil {
			t.Error("Error when setting TpAliases ", err.Error())
		}
		currentVersion := engine.CurrentStorDBVersions()
		err := mig.OutStorDB().SetVersions(currentVersion, false)
		if err != nil {
			t.Error("Error when setting version for TpAliases ", err.Error())
		}
		err, _ = mig.Migrate([]string{utils.MetaTpAliases})
		if err != nil {
			t.Error("Error when migrating TpAliases ", err.Error())
		}
		result, err := mig.OutStorDB().GetTPAliases(&utils.TPAliases{TPid: tpAliases[0].TPid})
		if err != nil {
			t.Error("Error when getting TpAliases ", err.Error())
		}
		if !reflect.DeepEqual(tpAliases[0], result[0]) {
			t.Errorf("Expecting: %+v, received: %+v", tpAliases[0], result[0])
		}
	}
}
