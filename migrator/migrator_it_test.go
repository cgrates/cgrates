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

/*
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

)

// subtests to be executed for each migrator
var sTestsITMigrator = []func(t *testing.T){
	testFlush,
	testMigratorAccounts, // Done
	testMigratorActionPlans,
	testMigratorActionTriggers,
	testMigratorActions,
	testMigratorSharedGroups,
	testMigratorStats,
	testMigratorSessionsCosts, // Done
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
	storDB, err := engine.ConfigureStorDB(cfg_in.StorDBType, cfg_in.StorDBHost, cfg_in.StorDBPort, cfg_in.StorDBName,
		cfg_in.StorDBUser, cfg_in.StorDBPass,
		config.CgrConfig().StorDBMaxOpenConns, config.CgrConfig().StorDBMaxIdleConns, config.CgrConfig().StorDBConnMaxLifetime, config.CgrConfig().StorDBCDRSIndexes)
	if err != nil {
		log.Fatal(err)
	}
	oldstorDB, err := engine.ConfigureStorDB(cfg_in.StorDBType, cfg_in.StorDBHost, cfg_in.StorDBPort, cfg_in.StorDBName,
		cfg_in.StorDBUser, cfg_in.StorDBPass,
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
	storDB, err := engine.ConfigureStorDB(cfg_in.StorDBType, cfg_in.StorDBHost, cfg_in.StorDBPort, cfg_in.StorDBName,
		cfg_in.StorDBUser, cfg_in.StorDBPass,
		config.CgrConfig().StorDBMaxOpenConns, config.CgrConfig().StorDBMaxIdleConns, config.CgrConfig().StorDBConnMaxLifetime, config.CgrConfig().StorDBCDRSIndexes)
	if err != nil {
		log.Fatal(err)
	}
	oldstorDB, err := engine.ConfigureStorDB(cfg_in.StorDBType, cfg_in.StorDBHost, cfg_in.StorDBPort, cfg_in.StorDBName,
		cfg_in.StorDBUser, cfg_in.StorDBPass,
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
	storDB, err := engine.ConfigureStorDB(cfg_in.StorDBType, cfg_in.StorDBHost, cfg_in.StorDBPort, cfg_in.StorDBName,
		cfg_in.StorDBUser, cfg_in.StorDBPass,
		config.CgrConfig().StorDBMaxOpenConns, config.CgrConfig().StorDBMaxIdleConns, config.CgrConfig().StorDBConnMaxLifetime, config.CgrConfig().StorDBCDRSIndexes)
	if err != nil {
		log.Fatal(err)
	}
	oldstorDB, err := engine.ConfigureStorDB(cfg_in.StorDBType, cfg_in.StorDBHost, cfg_in.StorDBPort, cfg_in.StorDBName,
		cfg_in.StorDBUser, cfg_in.StorDBPass,
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
	storDBIn, err := engine.ConfigureStorDB(cfg_in.StorDBType, cfg_in.StorDBHost, cfg_in.StorDBPort, cfg_in.StorDBName,
		cfg_in.StorDBUser, cfg_in.StorDBPass,
		config.CgrConfig().StorDBMaxOpenConns, config.CgrConfig().StorDBMaxIdleConns, config.CgrConfig().StorDBConnMaxLifetime, config.CgrConfig().StorDBCDRSIndexes)
	if err != nil {
		log.Fatal(err)
	}
	storDBOut, err := engine.ConfigureStorDB(cfg_out.StorDBType, cfg_out.StorDBHost, cfg_out.StorDBPort, cfg_out.StorDBName,
		cfg_out.StorDBUser, cfg_out.StorDBPass,
		config.CgrConfig().StorDBMaxOpenConns, config.CgrConfig().StorDBMaxIdleConns, config.CgrConfig().StorDBConnMaxLifetime, config.CgrConfig().StorDBCDRSIndexes)
	if err != nil {
		log.Fatal(err)
	}
	oldstorDB, err := ConfigureV1StorDB(cfg_out.StorDBType, cfg_out.StorDBHost, cfg_out.StorDBPort, cfg_out.StorDBName,
		cfg_out.StorDBUser, cfg_out.StorDBPass)
	if err != nil {
		log.Fatal(err)
	}
	mig, err = NewMigrator(dataDB2, dataDB, cfg_in.DataDbType, cfg_in.DBDataEncoding, storDBOut, storDBIn, cfg_in.StorDBType, oldDataDB,
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
		if err := mig.storDBIn.Flush(path.Join(cfg_in.DataFolderPath, "storage", cfg_in.StorDBType)); err != nil {
			t.Error(err)
		}
	}
	if path_out != "" {
		if err := mig.storDBOut.Flush(path.Join(cfg_out.DataFolderPath, "storage", cfg_out.StorDBType)); err != nil {
			t.Error(err)
		}
	}
}

func testMigratorActionPlans(t *testing.T) {
	v1ap := &v1ActionPlans{&v1ActionPlan{Id: "test", AccountIds: []string{"one"}, Timing: &engine.RateInterval{Timing: &engine.RITiming{Years: utils.Years{}, Months: utils.Months{}, MonthDays: utils.MonthDays{}, WeekDays: utils.WeekDays{}}}}}
	ap := &engine.ActionPlan{Id: "test", AccountIDs: utils.StringMap{"one": true}, ActionTimings: []*engine.ActionTiming{&engine.ActionTiming{Timing: &engine.RateInterval{Timing: &engine.RITiming{Years: utils.Years{}, Months: utils.Months{}, MonthDays: utils.MonthDays{}, WeekDays: utils.WeekDays{}}}}}}
	switch action {
	case utils.REDIS, utils.Mongo:
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
	case Move:
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
	switch action {
	case utils.REDIS, utils.MONGO:
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
	case Move:
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
	switch action {
	case utils.REDIS, utils.MONGO:
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
	case Move:
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
	switch action {
	case utils.REDIS, utils.MONGO:
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
	case Move:
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
	switch action {
	case utils.REDIS, utils.MONGO:
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
	case Move:
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
		currentVersion[utils.SessionSCosts] = 1
		err := mig.OutStorDB().SetVersions(currentVersion, false)
		if err != nil {
			t.Error("Error when setting version for SessionsCosts ", err.Error())
		}
		if vrs, err := mig.OutStorDB().GetVersions(utils.SessionSCosts); err != nil {
			t.Error(err)
		} else if vrs[utils.SessionSCosts] != 1 {
			t.Errorf("Expecting: 1, received: %+v", vrs[utils.SessionSCosts])
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
		if vrs, err := mig.OutStorDB().GetVersions(utils.SessionSCosts); err != nil {
			t.Error(err)
		} else if vrs[utils.SessionSCosts] != 3 {
			t.Errorf("Expecting: 3, received: %+v", vrs[utils.SessionSCosts])
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
*/
