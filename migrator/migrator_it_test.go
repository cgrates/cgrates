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
	rdsITdb   *engine.RedisStorage
	mgoITdb   *engine.MongoStorage
	onStor    engine.DataDB
	onStorCfg string
	dbtype    string
	mig       *Migrator
	dataDir   = flag.String("data_dir", "/usr/share/cgrates", "CGR data dir path here")
)

// subtests to be executed for each migrator
var sTestsOnStorIT = []func(t *testing.T){
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
	for _, stest := range sTestsOnStorIT {
		t.Run("TestOnStorITRedis", stest)
	}
}

func TestOnStorITMongoConnect(t *testing.T) {
	cdrsMongoCfgPath := path.Join(*dataDir, "conf", "samples", "cdrsv2mongo")
	mgoITCfg, err := config.NewCGRConfigFromFolder(cdrsMongoCfgPath)
	if err != nil {
		t.Fatal(err)
	}
	if mgoITdb, err = engine.NewMongoStorage(mgoITCfg.StorDBHost, mgoITCfg.StorDBPort, mgoITCfg.StorDBName, mgoITCfg.StorDBUser, mgoITCfg.StorDBPass,
		utils.StorDB, nil, mgoITCfg.CacheConfig, mgoITCfg.LoadHistorySize); err != nil {
		t.Fatal(err)
	}
	onStorCfg = mgoITCfg.StorDBName
	//mig = NewMigrator(mgoITdb, mgoITdb, utils.MONGO, utils.JSON, mgoITdb, utils.MONGO)
}

func TestOnStorITMongo(t *testing.T) {
	dbtype = utils.MONGO
	onStor = mgoITdb
	for _, stest := range sTestsOnStorIT {
		t.Run("TestOnStorITMongo", stest)
	}
}

func testOnStorITFlush(t *testing.T) {
	switch {
	case dbtype == utils.REDIS:
		err := mig.FlushRedis()
		if err != nil {
			t.Error("Error when flushing redis ", err.Error())
		}
	case dbtype == utils.MONGO:
		err := mig.FlushMongo()
		if err != nil {
			t.Error("Error when flushing redis ", err.Error())
		}
	}
}

func testMigratorAccounts(t *testing.T) {
	v1b := &v1Balance{Value: 10, Weight: 10, DestinationIds: "NAT", ExpirationDate: time.Date(2012, 1, 1, 0, 0, 0, 0, time.UTC).Local()}
	v1Acc := &v1Account{Id: "OUT:CUSTOMER_1:rif", BalanceMap: map[string]v1BalanceChain{utils.VOICE: v1BalanceChain{v1b}, utils.MONETARY: v1BalanceChain{&v1Balance{Value: 21, ExpirationDate: time.Date(2012, 1, 1, 0, 0, 0, 0, time.UTC).Local()}}}}
	v2 := &engine.Balance{Uuid: "", ID: "", Value: 10, Directions: utils.StringMap{"*OUT": true}, ExpirationDate: time.Date(2012, 1, 1, 0, 0, 0, 0, time.UTC).Local(), Weight: 10, DestinationIDs: utils.StringMap{"NAT": true}, RatingSubject: "", Categories: utils.NewStringMap(""), SharedGroups: utils.NewStringMap(""), TimingIDs: utils.NewStringMap("")}
	m2 := &engine.Balance{Uuid: "", ID: "", Value: 21, Directions: utils.StringMap{"*OUT": true}, ExpirationDate: time.Date(2012, 1, 1, 0, 0, 0, 0, time.UTC).Local(), DestinationIDs: utils.NewStringMap(""), RatingSubject: "", Categories: utils.NewStringMap(""), SharedGroups: utils.NewStringMap(""), TimingIDs: utils.NewStringMap("")}
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
		t.Errorf("not yet")
	}
}

func testMigratorActionPlans(t *testing.T) {
	v1ap := &v1ActionPlan{Id: "test", AccountIds: []string{"one"}, Timing: &engine.RateInterval{Timing: new(engine.RITiming)}}
	ap := &engine.ActionPlan{Id: "test", AccountIDs: utils.StringMap{"one": true}, ActionTimings: []*engine.ActionTiming{&engine.ActionTiming{Timing: &engine.RateInterval{Timing: new(engine.RITiming)}}}}
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
		t.Errorf("not yet")
	}
}

func testMigratorActionTriggers(t *testing.T) {
	tim := time.Date(0001, time.January, 1, 0, 0, 0, 0, time.UTC)
	v1atrs := &v1ActionTrigger{
		Id:               "Test",
		BalanceType:      "*monetary",
		BalanceDirection: "*out",
		ThresholdType:    "*max_balance",
		ThresholdValue:   2,
		ActionsId:        "TEST_ACTIONS",
		Executed:         true,
	}

	atrs := engine.ActionTriggers{
		&engine.ActionTrigger{
			ID: "Test",
			Balance: &engine.BalanceFilter{
				Type:       utils.StringPointer(utils.MONETARY),
				Directions: utils.StringMapPointer(utils.NewStringMap(utils.OUT)),
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
		if !reflect.DeepEqual(atrs, result) {
			t.Errorf("Expecting: %+v, received: %+v", atrs, result)
		}
	case dbtype == utils.MONGO:
		t.Errorf("not yet")
	}
}

func testMigratorActions(t *testing.T) {
	v1act := &v1Action{Id: "test", ActionType: "", BalanceType: "", Direction: "INBOUND", ExtraParameters: "", ExpirationString: "", Balance: &v1Balance{}}
	act := engine.Actions{&engine.Action{Id: "test", ActionType: "", ExtraParameters: "", ExpirationString: "", Weight: 0.00, Balance: &engine.BalanceFilter{}}}
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
		t.Errorf("not yet")
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
			t.Error("Error when setting v1 Actions ", err.Error())
		}

		err = mig.Migrate("migrateSharedGroups")
		if err != nil {
			t.Error("Error when migrating Actions ", err.Error())
		}
		result, err := mig.tpDB.GetSharedGroup(v1sg.Id, true, utils.NonTransactional)
		if err != nil {
			t.Error("Error when getting Actions ", err.Error())
		}
		if !reflect.DeepEqual(sg, result) {
			t.Errorf("Expecting: %+v, received: %+v", sg, result)
		}
	case dbtype == utils.MONGO:
		t.Errorf("not yet")
	}
}
