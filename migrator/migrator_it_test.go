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
	"path"
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	mongo     *config.CGRConfig
	rdsITdb   *engine.RedisStorage
	mgoITdb   *engine.MongoStorage
	onStor    engine.DataDB
	onStorCfg string
	dbtype    string
	mig       *Migrator
	dataDir   = flag.String("data_dir", "/usr/share/cgrates", "CGR data dir path here")
	db_passwd = ""
)

// subtests to be executed for each migrator
var sTestsITMigrator = []func(t *testing.T){
	testOnStorITFlush,
	testMigratorAccounts,
	testMigratorActionPlans,
	testMigratorActionTriggers,
	testMigratorActions,
	testMigratorSharedGroups,
}

func TestOnStorITRedisConnect(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	rdsITdb, err := engine.NewRedisStorage(fmt.Sprintf("%s:%s", cfg.TpDbHost, cfg.TpDbPort), 4, cfg.TpDbPass, cfg.DBDataEncoding, utils.REDIS_MAX_CONNS, nil, 1)
	if err != nil {
		t.Fatal("Could not connect to Redis", err.Error())
	}
	onStorCfg = cfg.DataDbName
	mig = NewMigrator(rdsITdb, rdsITdb, utils.REDIS, utils.JSON, rdsITdb, utils.REDIS)
}

func TestOnStorITRedis(t *testing.T) {
	dbtype = utils.REDIS
	onStor = rdsITdb
	for _, stest := range sTestsITMigrator {
		t.Run("TestITMigratorOnRedis", stest)
	}
}

func TestOnStorITMongoConnect(t *testing.T) {
	cdrsMongoCfgPath := path.Join(*dataDir, "conf", "samples", "cdrsv2mongo")
	mgoITCfg, err := config.NewCGRConfigFromFolder(cdrsMongoCfgPath)
	if err != nil {
		t.Fatal(err)
	}
	if mgoITdb, err = engine.NewMongoStorage(mgoITCfg.StorDBHost, mgoITCfg.StorDBPort, mgoITCfg.StorDBName, mgoITCfg.StorDBUser, db_passwd,
		utils.StorDB, nil, mgoITCfg.CacheConfig, mgoITCfg.LoadHistorySize); err != nil {
		t.Fatal(err)
	}
	mongo = mgoITCfg
	onStorCfg = mgoITCfg.StorDBName
	mig = NewMigrator(mgoITdb, mgoITdb, utils.MONGO, utils.JSON, mgoITdb, utils.MONGO)
}

func TestOnStorITMongo(t *testing.T) {
	dbtype = utils.MONGO
	onStor = mgoITdb
	for _, stest := range sTestsITMigrator {
		t.Run("TestITMigratorOnMongo", stest)
	}
}

func testOnStorITFlush(t *testing.T) {
	switch {
	case dbtype == utils.REDIS:
		dataDB := mig.dataDB.(*engine.RedisStorage)
		err := dataDB.Cmd("FLUSHALL").Err
		if err != nil {
			t.Error("Error when flushing Redis ", err.Error())
		}
	case dbtype == utils.MONGO:
		err := engine.InitDataDb(mongo)
		if err != nil {
			t.Error("Error when flushing Mongo ", err.Error())
		}
	}
}

func testMigratorAccounts(t *testing.T) {
	v1b := &v1Balance{Value: 10, Weight: 10, DestinationIds: "NAT", ExpirationDate: time.Date(2012, 1, 1, 0, 0, 0, 0, time.UTC).Local(), Timings: []*engine.RITiming{&engine.RITiming{Years: utils.Years{}, Months: utils.Months{}, MonthDays: utils.MonthDays{}, WeekDays: utils.WeekDays{}}}}
	v1Acc := &v1Account{Id: "OUT:CUSTOMER_1:rif", BalanceMap: map[string]v1BalanceChain{utils.VOICE: v1BalanceChain{v1b}, utils.MONETARY: v1BalanceChain{&v1Balance{Value: 21, ExpirationDate: time.Date(2012, 1, 1, 0, 0, 0, 0, time.UTC).Local(), Timings: []*engine.RITiming{&engine.RITiming{Years: utils.Years{}, Months: utils.Months{}, MonthDays: utils.MonthDays{}, WeekDays: utils.WeekDays{}}}}}}}
	v2 := &engine.Balance{Uuid: "", ID: "", Value: 10, Directions: utils.StringMap{"*OUT": true}, ExpirationDate: time.Date(2012, 1, 1, 0, 0, 0, 0, time.UTC).Local(), Weight: 10, DestinationIDs: utils.StringMap{"NAT": true},
		RatingSubject: "", Categories: utils.NewStringMap(), SharedGroups: utils.NewStringMap(), Timings: []*engine.RITiming{&engine.RITiming{Years: utils.Years{}, Months: utils.Months{}, MonthDays: utils.MonthDays{}, WeekDays: utils.WeekDays{}}}, TimingIDs: utils.NewStringMap(""), Factor: engine.ValueFactor{}}
	m2 := &engine.Balance{Uuid: "", ID: "", Value: 21, Directions: utils.StringMap{"*OUT": true}, ExpirationDate: time.Date(2012, 1, 1, 0, 0, 0, 0, time.UTC).Local(), DestinationIDs: utils.NewStringMap(""), RatingSubject: "",
		Categories: utils.NewStringMap(), SharedGroups: utils.NewStringMap(), Timings: []*engine.RITiming{&engine.RITiming{Years: utils.Years{}, Months: utils.Months{}, MonthDays: utils.MonthDays{}, WeekDays: utils.WeekDays{}}}, TimingIDs: utils.NewStringMap(""), Factor: engine.ValueFactor{}}
	testAccount := &engine.Account{ID: "CUSTOMER_1:rif", BalanceMap: map[string]engine.Balances{utils.VOICE: engine.Balances{v2}, utils.MONETARY: engine.Balances{m2}}, UnitCounters: engine.UnitCounters{}, ActionTriggers: engine.ActionTriggers{}}
	switch {
	case dbtype == utils.REDIS:
		bit, err := mig.mrshlr.Marshal(v1Acc)
		if err != nil {
			t.Error("Error when marshaling ", err.Error())
		}
		err = mig.SetV1onRedis(v1AccountDBPrefix+v1Acc.Id, bit)
		if err != nil {
			t.Error("Error when setting v1 acc ", err.Error())
		}
		err = mig.Migrate(utils.MetaAccounts)
		if err != nil {
			t.Error("Error when migrating accounts ", err.Error())
		}
		result, err := mig.dataDB.GetAccount(testAccount.ID)
		if err != nil {
			t.Error("Error when getting account ", err.Error())
		}
		if !reflect.DeepEqual(testAccount, result) {
			t.Errorf("Expecting: %+v, received: %+v", testAccount, result)
		}
	case dbtype == utils.MONGO:
		err := mig.SetV1onMongoAccount(v1AccountDBPrefix, v1Acc)
		if err != nil {
			t.Error("Error when marshaling ", err.Error())
		}
		err = mig.Migrate(utils.MetaAccounts)
		if err != nil {
			t.Error("Error when migrating accounts ", err.Error())
		}
		result, err := mig.dataDB.GetAccount(testAccount.ID)
		if err != nil {
			t.Error("Error when getting account ", err.Error())
		}
		if !reflect.DeepEqual(testAccount, result) {
			t.Errorf("Expecting: %+v, received: %+v", testAccount, result)
		}
	}
}

func testMigratorActionPlans(t *testing.T) {
	v1ap := &v1ActionPlan{Id: "test", AccountIds: []string{"one"}, Timing: &engine.RateInterval{Timing: &engine.RITiming{Years: utils.Years{}, Months: utils.Months{}, MonthDays: utils.MonthDays{}, WeekDays: utils.WeekDays{}}}}
	ap := &engine.ActionPlan{Id: "test", AccountIDs: utils.StringMap{"one": true}, ActionTimings: []*engine.ActionTiming{&engine.ActionTiming{Timing: &engine.RateInterval{Timing: &engine.RITiming{Years: utils.Years{}, Months: utils.Months{}, MonthDays: utils.MonthDays{}, WeekDays: utils.WeekDays{}}}}}}
	switch {
	case dbtype == utils.REDIS:
		bit, err := mig.mrshlr.Marshal(v1ap)
		if err != nil {
			t.Error("Error when marshaling ", err.Error())
		}
		setv1id := utils.ACTION_PLAN_PREFIX + v1ap.Id
		err = mig.SetV1onRedis(setv1id, bit)
		if err != nil {
			t.Error("Error when setting v1 ActionPlan ", err.Error())
		}
		err = mig.Migrate("migrateActionPlans")
		if err != nil {
			t.Error("Error when migrating ActionPlans ", err.Error())
		}
		result, err := mig.tpDB.GetActionPlan(ap.Id, true, utils.NonTransactional)
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
		err := mig.SetV1onMongoActionPlan(utils.ACTION_PLAN_PREFIX, v1ap)
		if err != nil {
			t.Error("Error when setting v1 ActionPlans ", err.Error())
		}
		err = mig.Migrate("migrateActionPlans")
		if err != nil {
			t.Error("Error when migrating ActionPlans ", err.Error())
		}
		result, err := mig.tpDB.GetActionPlan(ap.Id, true, utils.NonTransactional)
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
	var v1Atr v1ActionTrigger
	v1atrs := &v1ActionTrigger{
		Id:                    "Test",
		BalanceType:           "*monetary",
		BalanceDirection:      "*out",
		ThresholdType:         "*max_balance",
		ThresholdValue:        2,
		ActionsId:             "TEST_ACTIONS",
		Executed:              true,
		BalanceExpirationDate: tim,
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
		bit, err := mig.mrshlr.Marshal(v1atrs)
		if err != nil {
			t.Error("Error when marshaling ", err.Error())
		}
		if err := mig.mrshlr.Unmarshal(bit, &v1Atr); err != nil {
			t.Error("Error when setting v1 ActionTriggers ", err.Error())
		}
		setv1id := utils.ACTION_TRIGGER_PREFIX + v1atrs.Id
		err = mig.SetV1onRedis(setv1id, bit)
		if err != nil {
			t.Error("Error when setting v1 ActionTriggers ", err.Error())
		}
		err = mig.Migrate("migrateActionTriggers")
		if err != nil {
			t.Error("Error when migrating ActionTriggers ", err.Error())
		}
		result, err := mig.tpDB.GetActionTriggers(v1atrs.Id, true, utils.NonTransactional)
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
		} else if !reflect.DeepEqual(atrs[0].Balance, result[0].Balance) {
			//	t.Errorf("Expecting: %+v, received: %+v", atrs[0].Balance, result[0].Balance)
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
		err := mig.SetV1onMongoActionTrigger(utils.ACTION_TRIGGER_PREFIX, v1atrs)
		if err != nil {
			t.Error("Error when setting v1 ActionTriggers ", err.Error())
		}
		err = mig.Migrate("migrateActionTriggers")
		if err != nil {
			t.Error("Error when migrating ActionTriggers ", err.Error())
		}
		result, err := mig.tpDB.GetActionTriggers(v1atrs.Id, true, utils.NonTransactional)
		if err != nil {
			t.Error("Error when getting ActionTriggers ", err.Error())
		}
		if !reflect.DeepEqual(atrs[0], result[0]) {
			t.Errorf("Expecting: %+v, received: %+v", atrs[0], result[0])
		}
		err = mig.DropV1Colection(utils.ACTION_TRIGGER_PREFIX)
		if err != nil {
			t.Error("Error when flushing v1 ActionTriggers ", err.Error())
		}
	}
}

func testMigratorActions(t *testing.T) {
	v1act := &v1Action{Id: "test", ActionType: "", BalanceType: "", Direction: "INBOUND", ExtraParameters: "", ExpirationString: "", Balance: &v1Balance{Timings: []*engine.RITiming{&engine.RITiming{Years: utils.Years{}, Months: utils.Months{}, MonthDays: utils.MonthDays{}, WeekDays: utils.WeekDays{}}}}}
	act := engine.Actions{&engine.Action{Id: "test", ActionType: "", ExtraParameters: "", ExpirationString: "", Weight: 0.00, Balance: &engine.BalanceFilter{Timings: []*engine.RITiming{&engine.RITiming{Years: utils.Years{}, Months: utils.Months{}, MonthDays: utils.MonthDays{}, WeekDays: utils.WeekDays{}}}}}}
	switch {
	case dbtype == utils.REDIS:
		bit, err := mig.mrshlr.Marshal(v1act)
		if err != nil {
			t.Error("Error when marshaling ", err.Error())
		}
		setv1id := utils.ACTION_PREFIX + v1act.Id
		err = mig.SetV1onRedis(setv1id, bit)
		if err != nil {
			t.Error("Error when setting v1 Actions ", err.Error())
		}
		err = mig.Migrate("migrateActions")
		if err != nil {
			t.Error("Error when migrating Actions ", err.Error())
		}
		result, err := mig.tpDB.GetActions(v1act.Id, true, utils.NonTransactional)
		if err != nil {
			t.Error("Error when getting Actions ", err.Error())
		}
		if !reflect.DeepEqual(act, result) {
			t.Errorf("Expecting: %+v, received: %+v", act, result)
		}
	case dbtype == utils.MONGO:
		err := mig.SetV1onMongoAction(utils.ACTION_PREFIX, v1act)
		if err != nil {
			t.Error("Error when setting v1 Actions ", err.Error())
		}
		err = mig.Migrate("migrateActions")
		if err != nil {
			t.Error("Error when migrating Actions ", err.Error())
		}
		result, err := mig.tpDB.GetActions(v1act.Id, true, utils.NonTransactional)
		if err != nil {
			t.Error("Error when getting Actions ", err.Error())
		}
		if !reflect.DeepEqual(act[0].Balance.Timings, result[0].Balance.Timings) {
			t.Errorf("Expecting: %+v, received: %+v", act[0].Balance.Timings, result[0].Balance.Timings)
		}
		err = mig.DropV1Colection(utils.ACTION_PREFIX)
		if err != nil {
			t.Error("Error when flushing v1 Actions ", err.Error())
		}
	}
}

func testMigratorSharedGroups(t *testing.T) {
	v1sg := &v1SharedGroup{
		Id: "Test",
		AccountParameters: map[string]*engine.SharingParameters{
			"test": &engine.SharingParameters{Strategy: "*highest"},
		},
		MemberIds: []string{"1", "2", "3"},
	}
	sg := &engine.SharedGroup{
		Id: "Test",
		AccountParameters: map[string]*engine.SharingParameters{
			"test": &engine.SharingParameters{Strategy: "*highest"},
		},
		MemberIds: utils.NewStringMap("1", "2", "3"),
	}
	switch {
	case dbtype == utils.REDIS:
		bit, err := mig.mrshlr.Marshal(v1sg)
		if err != nil {
			t.Error("Error when marshaling ", err.Error())
		}
		setv1id := utils.SHARED_GROUP_PREFIX + v1sg.Id
		err = mig.SetV1onRedis(setv1id, bit)
		if err != nil {
			t.Error("Error when setting v1 SharedGroup ", err.Error())
		}
		err = mig.Migrate("migrateSharedGroups")
		if err != nil {
			t.Error("Error when migrating SharedGroup ", err.Error())
		}
		result, err := mig.tpDB.GetSharedGroup(v1sg.Id, true, utils.NonTransactional)
		if err != nil {
			t.Error("Error when getting SharedGroup ", err.Error())
		}
		if !reflect.DeepEqual(sg, result) {
			t.Errorf("Expecting: %+v, received: %+v", sg, result)
		}
	case dbtype == utils.MONGO:
		err := mig.SetV1onMongoSharedGroup(utils.SHARED_GROUP_PREFIX, v1sg)
		if err != nil {
			t.Error("Error when setting v1 SharedGroup ", err.Error())
		}
		err = mig.Migrate("migrateSharedGroups")
		if err != nil {
			t.Error("Error when migrating SharedGroup ", err.Error())
		}
		result, err := mig.tpDB.GetSharedGroup(v1sg.Id, true, utils.NonTransactional)
		if err != nil {
			t.Error("Error when getting SharedGroup ", err.Error())
		}
		if !reflect.DeepEqual(sg, result) {
			t.Errorf("Expecting: %+v, received: %+v", sg, result)
		}
	}
}
