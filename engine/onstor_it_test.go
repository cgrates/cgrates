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
package engine

import (
	"fmt"
	"path"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/cgrates/cgrates/cache"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

var (
	rdsITdb    *RedisStorage
	mgoITdb    *MongoStorage
	onStor     *DataManager
	onStorCfg  string
	sleepDelay time.Duration
)

// subtests to be executed for each confDIR
var sTestsOnStorIT = []func(t *testing.T){
	testOnStorITFlush,
	testOnStorITIsDBEmpty,
	testOnStorITSetGetDerivedCharges,
	testOnStorITSetFilterIndexes,
	testOnStorITGetFilterIndexes,
	testOnStorITMatchFilterIndex,
	testOnStorITCacheDestinations,
	testOnStorITCacheReverseDestinations,
	testOnStorITCacheActionPlan,
	testOnStorITCacheAccountActionPlans,
	testOnStorITCacheDerivedChargers,
	// ToDo: test cache flush for a prefix
	// ToDo: testOnStorITLoadAccountingCache
	testOnStorITHasData,
	testOnStorITPushPop,
	testOnStorITRatingPlan,
	testOnStorITRatingProfile,
	testOnStorITCRUDDestinations,
	testOnStorITCRUDReverseDestinations,
	testOnStorITLCR,
	testOnStorITActions,
	testOnStorITSharedGroup,
	testOnStorITCRUDActionPlan,
	testOnStorITCRUDAccountActionPlans,
	testOnStorITCRUDAccount,
	testOnStorITCRUDCdrStatsQueue,
	testOnStorITCRUDSubscribers,
	testOnStorITResource,
	testOnStorITResourceProfile,
	testOnStorITTiming,
	testOnStorITCRUDHistory,
	testOnStorITCRUDStructVersion,
	testOnStorITStatQueueProfile,
	testOnStorITStatQueue,
	testOnStorITThresholdProfile,
	testOnStorITThreshold,
	testOnStorITFilter,
	testOnStorITSupplierProfile,
	testOnStorITAttributeProfile,
	testOnStorITFlush,
	testOnStorITIsDBEmpty,
	testOnStorITTestThresholdFilterIndexes,
	testOnStorITTestAttributeProfileFilterIndexes,
	testOnStorITTestThresholdInlineFilterIndexing,
	//testOnStorITCacheActionTriggers,
	//testOnStorITCacheAlias,
	//testOnStorITCacheReverseAlias,
	//testOnStorITCRUDActionTriggers,
	//testOnStorITCRUDUser,
	//testOnStorITCRUDAlias,
	//testOnStorITCRUDReverseAlias,
}

func TestOnStorITRedisConnect(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	rdsITdb, err = NewRedisStorage(fmt.Sprintf("%s:%s", cfg.DataDbHost, cfg.DataDbPort), 4,
		cfg.DataDbPass, cfg.DBDataEncoding, utils.REDIS_MAX_CONNS, nil, 1)
	if err != nil {
		t.Fatal("Could not connect to Redis", err.Error())
	}
	onStorCfg = cfg.DataDbName
}

func TestOnStorITRedis(t *testing.T) {
	onStor = NewDataManager(rdsITdb)
	for _, stest := range sTestsOnStorIT {
		t.Run("TestOnStorITRedis", stest)
	}
}

func TestOnStorITMongoConnect(t *testing.T) {
	sleepDelay = 500 * time.Millisecond
	cdrsMongoCfgPath := path.Join(*dataDir, "conf", "samples", "cdrsv2mongo")
	mgoITCfg, err := config.NewCGRConfigFromFolder(cdrsMongoCfgPath)
	if err != nil {
		t.Fatal(err)
	}
	if mgoITdb, err = NewMongoStorage(mgoITCfg.StorDBHost, mgoITCfg.StorDBPort,
		mgoITCfg.StorDBName, mgoITCfg.StorDBUser, mgoITCfg.StorDBPass,
		utils.StorDB, nil, mgoITCfg.CacheCfg(), mgoITCfg.LoadHistorySize); err != nil {
		t.Fatal(err)
	}
	onStorCfg = mgoITCfg.StorDBName
}

func TestOnStorITMongo(t *testing.T) {
	onStor = NewDataManager(mgoITdb)
	for _, stest := range sTestsOnStorIT {
		t.Run("TestOnStorITMongo", stest)
	}
}

func testOnStorITFlush(t *testing.T) {
	if err := onStor.DataDB().Flush(""); err != nil {
		t.Error(err)
	}
	cache.Flush()
}

func testOnStorITIsDBEmpty(t *testing.T) {
	test, err := onStor.DataDB().IsDBEmpty()
	if err != nil {
		t.Error(err)
	} else if test != true {
		t.Errorf("\nExpecting: true got :%+v", test)
	}
}

func testOnStorITSetGetDerivedCharges(t *testing.T) {
	keyCharger1 := utils.ConcatenatedKey("*out", "cgrates.org", "call", "dan", "dan")
	if _, err := onStor.GetDerivedChargers(keyCharger1, true, utils.NonTransactional); err == nil {
		t.Error("DC exists")
	}
	charger1 := &utils.DerivedChargers{DestinationIDs: make(utils.StringMap),
		Chargers: []*utils.DerivedCharger{
			&utils.DerivedCharger{RunID: "extra1", RequestTypeField: "^prepaid", DirectionField: "*default", TenantField: "*default", CategoryField: "*default",
				AccountField: "rif", SubjectField: "rif", DestinationField: "*default", SetupTimeField: "*default", AnswerTimeField: "*default", UsageField: "*default"},
			&utils.DerivedCharger{RunID: "extra2", RequestTypeField: "*default", DirectionField: "*default", TenantField: "*default", CategoryField: "*default",
				AccountField: "ivo", SubjectField: "ivo", DestinationField: "*default", SetupTimeField: "*default", AnswerTimeField: "*default", UsageField: "*default"},
		}}
	if err := onStor.DataDB().SetDerivedChargers(keyCharger1, charger1, utils.NonTransactional); err != nil {
		t.Error("Error on setting DerivedChargers", err.Error())
	}
	// Retrieve from db
	if rcvCharger, err := onStor.GetDerivedChargers(keyCharger1, true, utils.NonTransactional); err != nil {
		t.Error("Error when retrieving DerivedCHarger", err.Error())
	} else if !reflect.DeepEqual(rcvCharger, charger1) {
		for i, eChrg := range charger1.Chargers {
			if !reflect.DeepEqual(eChrg, rcvCharger.Chargers[i]) {
				t.Logf("Expecting: %+v, received: %+v", eChrg, rcvCharger.Chargers[i])
			}
		}
		t.Errorf("Expecting %v, received: %v", charger1, rcvCharger)
	}
	// Retrieve from cache
	if rcvCharger, err := onStor.GetDerivedChargers(keyCharger1, false, utils.NonTransactional); err != nil {
		t.Error("Error when retrieving DerivedCHarger", err.Error())
	} else if !reflect.DeepEqual(rcvCharger, charger1) {
		t.Errorf("Expecting %v, received: %v", charger1, rcvCharger)
	}
	if err := onStor.RemoveDerivedChargers(keyCharger1, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	if _, rcvErr := onStor.GetDerivedChargers(keyCharger1, false, utils.NonTransactional); rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
}

func testOnStorITSetFilterIndexes(t *testing.T) {
	idxes := map[string]utils.StringMap{
		"*string:Account:1001": utils.StringMap{
			"RL1": true,
		},
		"*string:Account:1002": utils.StringMap{
			"RL1": true,
			"RL2": true,
		},
		"*string:Account:dan": utils.StringMap{
			"RL2": true,
		},
		"*string:Subject:dan": utils.StringMap{
			"RL2": true,
			"RL3": true,
		},
		utils.ConcatenatedKey(utils.MetaDefault, utils.ANY, utils.ANY): utils.StringMap{
			"RL4": true,
			"RL5": true,
		},
	}
	if err := onStor.SetFilterIndexes(
		GetDBIndexKey(utils.ResourceProfilesPrefix, "cgrates.org", false),
		idxes); err != nil {
		t.Error(err)
	}
}

func testOnStorITGetFilterIndexes(t *testing.T) {
	eIdxes := map[string]utils.StringMap{
		"*string:Account:1001": utils.StringMap{
			"RL1": true,
		},
		"*string:Account:1002": utils.StringMap{
			"RL1": true,
			"RL2": true,
		},
		"*string:Account:dan": utils.StringMap{
			"RL2": true,
		},
		"*string:Subject:dan": utils.StringMap{
			"RL2": true,
			"RL3": true,
		},
		utils.ConcatenatedKey(utils.MetaDefault, utils.ANY, utils.ANY): utils.StringMap{
			"RL4": true,
			"RL5": true,
		},
	}
	sbjDan := map[string]string{
		"Subject": "dan",
	}
	expectedsbjDan := map[string]utils.StringMap{
		"*string:Subject:dan": utils.StringMap{
			"RL2": true,
			"RL3": true,
		},
	}
	if exsbjDan, err := onStor.GetFilterIndexes(
		GetDBIndexKey(utils.ResourceProfilesPrefix, "cgrates.org", false), MetaString,
		sbjDan); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedsbjDan, exsbjDan) {
		t.Errorf("Expecting: %+v, received: %+v", expectedsbjDan, exsbjDan)
	}
	if rcv, err := onStor.GetFilterIndexes(
		GetDBIndexKey(utils.ResourceProfilesPrefix, "cgrates.org", false), MetaString,
		nil); err != nil {
		t.Error(err)
	} else {
		if !reflect.DeepEqual(eIdxes, rcv) {
			t.Errorf("Expecting: %+v, received: %+v", eIdxes, rcv)
		}
	}
	if _, err := onStor.GetFilterIndexes("unknown_key", MetaString, nil); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	}
	if err := onStor.RemoveFilterIndexes(
		GetDBIndexKey(utils.ResourceProfilesPrefix, "cgrates.org", false)); err != nil {
		t.Error(err)
	}
	_, err := onStor.GetFilterIndexes(
		GetDBIndexKey(utils.ResourceProfilesPrefix, "cgrates.org", false), MetaString, nil)
	if err != utils.ErrNotFound {
		t.Error(err)
	}
	if err := onStor.SetFilterIndexes(
		GetDBIndexKey(utils.ResourceProfilesPrefix, "cgrates.org", false),
		eIdxes); err != nil {
		t.Error(err)
	}
}

func testOnStorITMatchFilterIndex(t *testing.T) {
	eMp := utils.StringMap{
		"RL1": true,
		"RL2": true,
	}
	if rcvMp, err := onStor.MatchFilterIndex(
		GetDBIndexKey(utils.ResourceProfilesPrefix, "cgrates.org", false),
		MetaString, "Account", "1002"); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eMp, rcvMp) {
		t.Errorf("Expecting: %+v, received: %+v", eMp, rcvMp)
	}
	if _, err := onStor.MatchFilterIndex(
		GetDBIndexKey(utils.ResourceProfilesPrefix, "cgrates.org", false),
		MetaString, "NonexistentField", "1002"); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	}
}

func testOnStorITCacheDestinations(t *testing.T) {
	if err := onStor.CacheDataFromDB("INVALID", nil, false); err == nil || err.Error() != utils.UnsupportedCachePrefix {
		t.Error(err)
	}
	dst := &Destination{Id: "TEST_CACHE", Prefixes: []string{"+491", "+492", "+493"}}
	if err := onStor.DataDB().SetDestination(dst, utils.NonTransactional); err != nil {
		t.Error(err)
	}

	if _, hasIt := cache.Get(utils.DESTINATION_PREFIX + dst.Id); hasIt {
		t.Error("Already in cache")
	}
	if err := onStor.CacheDataFromDB(utils.DESTINATION_PREFIX, []string{dst.Id}, true); err != nil { // Should not cache due to mustBeCached
		t.Error(err)
	}
	if _, hasIt := cache.Get(utils.DESTINATION_PREFIX + dst.Id); hasIt {
		t.Error("Should not be in cache")
	}
	if err := onStor.CacheDataFromDB(utils.DESTINATION_PREFIX, []string{dst.Id}, false); err != nil {
		t.Error(err)
	}
	if itm, hasIt := cache.Get(utils.DESTINATION_PREFIX + dst.Id); !hasIt {
		t.Error("Did not cache")
	} else if !reflect.DeepEqual(dst, itm.(*Destination)) {
		t.Error("Wrong item in the cache")
	}
}

func testOnStorITCacheReverseDestinations(t *testing.T) {
	dst := &Destination{Id: "TEST_CACHE", Prefixes: []string{"+491", "+492", "+493"}}
	if err := onStor.DataDB().SetReverseDestination(dst, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	for _, prfx := range dst.Prefixes {
		if _, hasIt := cache.Get(utils.REVERSE_DESTINATION_PREFIX + dst.Id); hasIt {
			t.Errorf("Prefix: %s already in cache", prfx)
		}
	}
	if err := onStor.CacheDataFromDB(utils.REVERSE_DESTINATION_PREFIX, dst.Prefixes, false); err != nil {
		t.Error(err)
	}
	for _, prfx := range dst.Prefixes {
		if itm, hasIt := cache.Get(utils.REVERSE_DESTINATION_PREFIX + prfx); !hasIt {
			t.Error("Did not cache")
		} else if !reflect.DeepEqual([]string{dst.Id}, itm.([]string)) {
			t.Error("Wrong item in the cache")
		}
	}
}

func testOnStorITCacheActionPlan(t *testing.T) {
	ap := &ActionPlan{
		Id:         "MORE_MINUTES",
		AccountIDs: utils.StringMap{"vdf:minitsboy": true},
		ActionTimings: []*ActionTiming{
			&ActionTiming{
				Uuid: utils.GenUUID(),
				Timing: &RateInterval{
					Timing: &RITiming{
						Years:     utils.Years{2012},
						Months:    utils.Months{},
						MonthDays: utils.MonthDays{},
						WeekDays:  utils.WeekDays{},
						StartTime: utils.ASAP,
					},
				},
				Weight:    10,
				ActionsID: "MINI",
			},
			&ActionTiming{
				Uuid: utils.GenUUID(),
				Timing: &RateInterval{
					Timing: &RITiming{
						Years:     utils.Years{2012},
						Months:    utils.Months{},
						MonthDays: utils.MonthDays{},
						WeekDays:  utils.WeekDays{},
						StartTime: utils.ASAP,
					},
				},
				Weight:    10,
				ActionsID: "SHARED",
			},
		},
	}
	if err := onStor.DataDB().SetActionPlan(ap.Id, ap, true, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	expectedCAp := []string{"apl_MORE_MINUTES"}
	if itm, err := onStor.DataDB().GetKeysForPrefix(utils.ACTION_PLAN_PREFIX); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedCAp, itm) {
		t.Errorf("Expected : %+v, but received %+v", expectedCAp, itm)
	}
	if _, hasIt := cache.Get(utils.ACTION_PLAN_PREFIX + ap.Id); hasIt {
		t.Error("Already in cache")
	}
	if err := onStor.CacheDataFromDB(utils.ACTION_PLAN_PREFIX, []string{ap.Id}, false); err != nil {
		t.Error(err)
	}
	if itm, hasIt := cache.Get(utils.ACTION_PLAN_PREFIX + ap.Id); !hasIt {
		t.Error("Did not cache")
	} else if rcv := itm.(*ActionPlan); !reflect.DeepEqual(ap, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", ap, rcv)
	}
	if err := onStor.DataDB().RemoveActionPlan(ap.Id, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	if err := onStor.DataDB().SetActionPlan(ap.Id, ap, true, utils.NonTransactional); err != nil {
		t.Error(err)
	}
}

func testOnStorITCacheAccountActionPlans(t *testing.T) {
	acntID := utils.ConcatenatedKey("cgrates.org", "1001")
	aAPs := []string{"PACKAGE_10_SHARED_A_5", "USE_SHARED_A", "apl_PACKAGE_1001"}
	if err := onStor.DataDB().SetAccountActionPlans(acntID, aAPs, true); err != nil {
		t.Error(err)
	}
	if _, hasIt := cache.Get(utils.AccountActionPlansPrefix + acntID); hasIt {
		t.Error("Already in cache")
	}
	if err := onStor.CacheDataFromDB(utils.AccountActionPlansPrefix, []string{acntID}, false); err != nil {
		t.Error(err)
	}
	if itm, hasIt := cache.Get(utils.AccountActionPlansPrefix + acntID); !hasIt {
		t.Error("Did not cache")
	} else if rcv := itm.([]string); !reflect.DeepEqual(aAPs, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", aAPs, rcv)
	}
}

func testOnStorITCacheActionTriggers(t *testing.T) {
	ats := ActionTriggers{
		&ActionTrigger{
			ID:                "testOnStorITCacheActionTrigger",
			Balance:           &BalanceFilter{Type: utils.StringPointer(utils.MONETARY), Directions: utils.StringMapPointer(utils.NewStringMap(utils.OUT)), Timings: make([]*RITiming, 0)},
			ThresholdValue:    2,
			ThresholdType:     utils.TRIGGER_MAX_EVENT_COUNTER,
			ActionsID:         "TEST_ACTIONS",
			LastExecutionTime: time.Date(2012, 1, 1, 0, 0, 0, 0, time.UTC),
			ExpirationDate:    time.Date(2012, 1, 1, 0, 0, 0, 0, time.UTC),
			ActivationDate:    time.Date(2012, 1, 1, 0, 0, 0, 0, time.UTC)},
	}
	atsID := ats[0].ID
	if err := onStor.SetActionTriggers(atsID, ats, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	expectedCAt := []string{"atr_testOnStorITCacheActionTrigger"}
	if itm, err := onStor.DataDB().GetKeysForPrefix(utils.ACTION_TRIGGER_PREFIX); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedCAt, itm) {
		t.Errorf("Expected : %+v, but received %+v", expectedCAt, itm)
	}
	if _, hasIt := cache.Get(utils.ACTION_TRIGGER_PREFIX + atsID); hasIt {
		t.Error("Already in cache")
	}
	if err := onStor.CacheDataFromDB(utils.ACTION_TRIGGER_PREFIX, []string{atsID}, false); err != nil {
		t.Error(err)
	}
	if itm, hasIt := cache.Get(utils.ACTION_TRIGGER_PREFIX + atsID); !hasIt {
		t.Error("Did not cache")
	} else if rcv := itm.(ActionTriggers); !reflect.DeepEqual(ats, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", ats, rcv)
	}
}

func testOnStorITCacheDerivedChargers(t *testing.T) {
	dcs := &utils.DerivedChargers{
		DestinationIDs: make(utils.StringMap),
		Chargers: []*utils.DerivedCharger{
			&utils.DerivedCharger{RunID: "extra1", RunFilters: "^filteredHeader1/filterValue1/", RequestTypeField: "^prepaid", DirectionField: utils.META_DEFAULT,
				TenantField: utils.META_DEFAULT, CategoryField: utils.META_DEFAULT, AccountField: "rif", SubjectField: "rif", DestinationField: utils.META_DEFAULT,
				SetupTimeField: utils.META_DEFAULT, PDDField: utils.META_DEFAULT, AnswerTimeField: utils.META_DEFAULT, UsageField: utils.META_DEFAULT,
				SupplierField: utils.META_DEFAULT, DisconnectCauseField: utils.META_DEFAULT, CostField: utils.META_DEFAULT, RatedField: utils.META_DEFAULT},
			&utils.DerivedCharger{RunID: "extra2", RequestTypeField: utils.META_DEFAULT, DirectionField: utils.META_DEFAULT, TenantField: utils.META_DEFAULT,
				CategoryField: utils.META_DEFAULT, AccountField: "ivo", SubjectField: "ivo", DestinationField: utils.META_DEFAULT,
				SetupTimeField: utils.META_DEFAULT, PDDField: utils.META_DEFAULT, AnswerTimeField: utils.META_DEFAULT, UsageField: utils.META_DEFAULT,
				SupplierField: utils.META_DEFAULT, DisconnectCauseField: utils.META_DEFAULT, CostField: utils.META_DEFAULT, RatedField: utils.META_DEFAULT},
		}}
	keyDCS := utils.ConcatenatedKey("*out", "itsyscom.com", "call", "dan", "dan")
	if err := onStor.DataDB().SetDerivedChargers(keyDCS, dcs, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	if _, hasIt := cache.Get(utils.DERIVEDCHARGERS_PREFIX + keyDCS); hasIt {
		t.Error("Already in cache")
	}
	if err := onStor.CacheDataFromDB(utils.DERIVEDCHARGERS_PREFIX, []string{keyDCS}, false); err != nil {
		t.Error(err)
	}
	if itm, hasIt := cache.Get(utils.DERIVEDCHARGERS_PREFIX + keyDCS); !hasIt {
		t.Error("Did not cache")
	} else if rcv := itm.(*utils.DerivedChargers); !reflect.DeepEqual(dcs, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", dcs, rcv)
	}
}

func testOnStorITCacheAlias(t *testing.T) {
	als := &Alias{
		Direction: "*out",
		Tenant:    "cgrates.org",
		Category:  "call",
		Account:   "dan",
		Subject:   "dan",
		Context:   "*rating",
		Values: AliasValues{
			&AliasValue{
				DestinationId: "EU_LANDLINE",
				Pairs: AliasPairs{
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

			&AliasValue{
				DestinationId: "GLOBAL1",
				Pairs:         AliasPairs{"Subject": map[string]string{"dan": "dan2"}},
				Weight:        20,
			},
		},
	}
	if err := onStor.DataDB().SetAlias(als, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	expectedCA := []string{"als_*out:cgrates.org:call:dan:dan:*rating"}
	if itm, err := onStor.DataDB().GetKeysForPrefix(utils.ALIASES_PREFIX); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedCA, itm) {
		t.Errorf("Expected : %+v, but received %+v", expectedCA, itm)
	}
	if _, hasIt := cache.Get(utils.ALIASES_PREFIX + als.GetId()); hasIt {
		t.Error("Already in cache")
	}
	if err := onStor.CacheDataFromDB(utils.ALIASES_PREFIX, []string{als.GetId()}, false); err != nil {
		t.Error(err)
	}
	if itm, hasIt := cache.Get(utils.ALIASES_PREFIX + als.GetId()); !hasIt {
		t.Error("Did not cache")
	} else if rcv := itm.(*Alias); !reflect.DeepEqual(als, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", als, rcv)
	}
}

func testOnStorITCacheReverseAlias(t *testing.T) {
	als := &Alias{
		Direction: "*out",
		Tenant:    "itsyscom.com",
		Category:  "call",
		Account:   "dan",
		Subject:   "dan",
		Context:   "*rating",
		Values: AliasValues{
			&AliasValue{
				DestinationId: "EU",
				Pairs: AliasPairs{
					"Account": map[string]string{
						"dan": "dan1",
						"rif": "rif1",
					},
					"Calling": map[string]string{
						"11234": "2234",
					},
				},
				Weight: 10,
			},

			&AliasValue{
				DestinationId: "US",
				Pairs:         AliasPairs{"Account": map[string]string{"dan": "dan2"}},
				Weight:        20,
			},
		},
	}
	if err := onStor.DataDB().SetReverseAlias(als, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	rvAlsID := strings.Join([]string{als.Values[1].Pairs["Account"]["dan"], "Account", als.Context}, "")
	if _, hasIt := cache.Get(utils.REVERSE_ALIASES_PREFIX + rvAlsID); hasIt {
		t.Error("Already in cache")
	}
	if err := onStor.CacheDataFromDB(utils.REVERSE_ALIASES_PREFIX, []string{rvAlsID}, false); err != nil {
		t.Error(err)
	}
	eRvrsAls := []string{utils.ConcatenatedKey(als.GetId(), als.Values[1].DestinationId)}
	if itm, hasIt := cache.Get(utils.REVERSE_ALIASES_PREFIX + rvAlsID); !hasIt {
		t.Error("Did not cache")
	} else if rcv := itm.([]string); !reflect.DeepEqual(eRvrsAls, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", eRvrsAls, rcv)
	}
}

func testOnStorITHasData(t *testing.T) {
	rp := &RatingPlan{
		Id: "HasData",
		Timings: map[string]*RITiming{
			"59a981b9": &RITiming{
				Years:     utils.Years{},
				Months:    utils.Months{},
				MonthDays: utils.MonthDays{},
				WeekDays:  utils.WeekDays{1, 2, 3, 4, 5},
				StartTime: "00:00:00",
			},
		},
		Ratings: map[string]*RIRate{
			"ebefae11": &RIRate{
				ConnectFee: 0,
				Rates: []*Rate{
					&Rate{
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
		DestinationRates: map[string]RPRateList{
			"GERMANY": []*RPRate{
				&RPRate{
					Timing: "59a981b9",
					Rating: "ebefae11",
					Weight: 10,
				},
			},
		},
	}
	if err := onStor.SetRatingPlan(rp, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	expectedRP := []string{"rpl_HasData"}
	if itm, err := onStor.DataDB().GetKeysForPrefix(utils.RATING_PLAN_PREFIX); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(len(expectedRP), len(itm)) {
		t.Errorf("Expected : %+v, but received %+v", len(expectedRP), len(itm))
	}
	if rcv, err := onStor.HasData(utils.RATING_PLAN_PREFIX, rp.Id); err != nil {
		t.Error(err)
	} else if rcv != true {
		t.Errorf("Expecting: true, received: %v", rcv)
	}
}

func testOnStorITPushPop(t *testing.T) {
	if err := onStor.DataDB().PushTask(&Task{Uuid: "1"}); err != nil {
		t.Error("Error pushing task: ", err)
	}
	if err := onStor.DataDB().PushTask(&Task{Uuid: "2"}); err != nil {
		t.Error("Error pushing task: ", err)
	}
	if err := onStor.DataDB().PushTask(&Task{Uuid: "3"}); err != nil {
		t.Error("Error pushing task: ", err)
	}
	if err := onStor.DataDB().PushTask(&Task{Uuid: "4"}); err != nil {
		t.Error("Error pushing task: ", err)
	}
	if task, err := onStor.DataDB().PopTask(); err != nil && task.Uuid != "1" {
		t.Error("Error poping task: ", task, err)
	}
	if task, err := onStor.DataDB().PopTask(); err != nil && task.Uuid != "2" {
		t.Error("Error poping task: ", task, err)
	}
	if task, err := onStor.DataDB().PopTask(); err != nil && task.Uuid != "3" {
		t.Error("Error poping task: ", task, err)
	}
	if task, err := onStor.DataDB().PopTask(); err != nil && task.Uuid != "4" {
		t.Error("Error poping task: ", task, err)
	}
	if task, err := onStor.DataDB().PopTask(); err == nil && task != nil {
		t.Errorf("Error poping task %+v, %v ", task, err)
	}
}

func testOnStorITRatingPlan(t *testing.T) {
	rp := &RatingPlan{
		Id: "CRUDRatingPlan",
		Timings: map[string]*RITiming{
			"59a981b9": &RITiming{
				Years:     utils.Years{},
				Months:    utils.Months{},
				MonthDays: utils.MonthDays{},
				WeekDays:  utils.WeekDays{1, 2, 3, 4, 5},
				StartTime: "00:00:00",
			},
		},
		Ratings: map[string]*RIRate{
			"ebefae11": &RIRate{
				ConnectFee: 0,
				Rates: []*Rate{
					&Rate{
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
		DestinationRates: map[string]RPRateList{
			"GERMANY": []*RPRate{
				&RPRate{
					Timing: "59a981b9",
					Rating: "ebefae11",
					Weight: 10,
				},
			},
		},
	}
	if _, rcvErr := onStor.GetRatingPlan(rp.Id, false,
		utils.NonTransactional); rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
	if err := onStor.SetRatingPlan(rp, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	//get from cache
	if rcv, err := onStor.GetRatingPlan(rp.Id, false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rp, rcv) {
		t.Errorf("Expecting: %v, received: %v", rp, rcv)
	}
	//get from database
	if rcv, err := onStor.GetRatingPlan(rp.Id, true, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rp, rcv) {
		t.Errorf("Expecting: %v, received: %v", rp, rcv)
	}
	expectedRP := []string{"rpl_HasData", "rpl_CRUDRatingPlan"}
	if itm, err := onStor.DataDB().GetKeysForPrefix(utils.RATING_PLAN_PREFIX); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(len(expectedRP), len(itm)) {
		t.Errorf("Expected : %+v, but received %+v", len(expectedRP), len(itm))
	}
	//update
	rp.Timings = map[string]*RITiming{
		"59a981b9": &RITiming{
			Years:     utils.Years{},
			Months:    utils.Months{},
			MonthDays: utils.MonthDays{},
			WeekDays:  utils.WeekDays{1, 2, 3, 4, 5},
			StartTime: "00:00:00",
		},
		"59a981a1": &RITiming{
			Years:     utils.Years{},
			Months:    utils.Months{},
			MonthDays: utils.MonthDays{},
			WeekDays:  utils.WeekDays{6, 7},
			StartTime: "00:00:00",
		},
	}
	time.Sleep(sleepDelay)
	if err := onStor.SetRatingPlan(rp, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	//get from cache
	if rcv, err := onStor.GetRatingPlan(rp.Id, false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rp, rcv) {
		t.Errorf("Expecting: %v, received: %v", rp, rcv)
	}
	//get from database
	if rcv, err := onStor.GetRatingPlan(rp.Id, true, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rp, rcv) {
		t.Errorf("Expecting: %v, received: %v", rp, rcv)
	}
	if err = onStor.RemoveRatingPlan(rp.Id, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	//check cache if removed
	if _, rcvErr := onStor.GetRatingPlan(rp.Id, false,
		utils.NonTransactional); rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
	//check database if removed
	if _, rcvErr := onStor.GetRatingPlan(rp.Id, true,
		utils.NonTransactional); rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
}

func testOnStorITRatingProfile(t *testing.T) {
	rpf := &RatingProfile{
		Id: "*out:test:1:trp",
		RatingPlanActivations: RatingPlanActivations{
			&RatingPlanActivation{
				ActivationTime:  time.Date(2013, 10, 1, 0, 0, 0, 0, time.UTC),
				RatingPlanId:    "TDRT",
				FallbackKeys:    []string{"*out:test:1:danb", "*out:test:1:rif"},
				CdrStatQueueIds: []string{},
			}},
	}
	if _, rcvErr := onStor.GetRatingProfile(rpf.Id, false,
		utils.NonTransactional); rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
	if err := onStor.SetRatingProfile(rpf, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	//get from cache
	if rcv, err := onStor.GetRatingProfile(rpf.Id, false,
		utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rpf, rcv) {
		t.Errorf("Expecting: %v, received: %v", utils.ToJSON(rpf), utils.ToJSON(rcv))
	}
	//get from database
	if rcv, err := onStor.GetRatingProfile(rpf.Id, true,
		utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rpf, rcv) {
		t.Errorf("Expecting: %v, received: %v", rpf, rcv)
	}
	expectedCRPl := []string{"rpf_*out:test:1:trp"}
	if itm, err := onStor.DataDB().GetKeysForPrefix(utils.RATING_PROFILE_PREFIX); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedCRPl, itm) {
		t.Errorf("Expected : %+v, but received %+v", expectedCRPl, itm)
	}
	//update
	rpf.RatingPlanActivations = RatingPlanActivations{
		&RatingPlanActivation{
			ActivationTime:  time.Date(2013, 10, 1, 0, 0, 0, 0, time.UTC),
			RatingPlanId:    "TDRT",
			FallbackKeys:    []string{"*out:test:1:danb", "*out:test:1:teo"},
			CdrStatQueueIds: []string{},
		},
	}
	time.Sleep(sleepDelay)
	if err := onStor.SetRatingProfile(rpf, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	//get from cache
	if rcv, err := onStor.GetRatingProfile(rpf.Id, false,
		utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rpf, rcv) {
		t.Errorf("Expecting: %v, received: %v", rpf, rcv)
	}
	//get from database
	if rcv, err := onStor.GetRatingProfile(rpf.Id, true,
		utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rpf, rcv) {
		t.Errorf("Expecting: %v, received: %v", rpf, rcv)
	}
	if err = onStor.RemoveRatingProfile(rpf.Id, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	//check cache if removed
	if _, rcvErr := onStor.GetRatingProfile(rpf.Id, false,
		utils.NonTransactional); rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
	//check database if removed
	if _, rcvErr := onStor.GetRatingProfile(rpf.Id, true,
		utils.NonTransactional); rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
}

func testOnStorITCRUDDestinations(t *testing.T) {
	dst := &Destination{Id: "CRUDDestination2", Prefixes: []string{"+491", "+492", "+493"}}
	if _, rcvErr := onStor.DataDB().GetDestination(dst.Id, true, utils.NonTransactional); rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
	if err := onStor.DataDB().SetDestination(dst, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	if rcv, err := onStor.DataDB().GetDestination(dst.Id, true, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(dst, rcv) {
		t.Errorf("Expecting: %v, received: %v", dst, rcv)
	}
	// FixMe
	// if err = onStor.DataDB().SelectDatabase("13"); err != nil {
	// 	t.Error(err)
	// }
	// if _, rcvErr := onStor.DataDB().GetDestination(dst.Id, false, utils.NonTransactional); rcvErr != utils.ErrNotFound {
	// 	t.Error(rcvErr)
	// }
	//
	// if rcv, err := onStor.DataDB().GetDestination(dst.Id, false, utils.NonTransactional); err != nil {
	// 	t.Error(err)
	// } else if !reflect.DeepEqual(dst, rcv) {
	// 	t.Errorf("Expecting: %v, received: %v", dst, rcv)
	// }
	// if err = onStor.DataDB().SelectDatabase(onStorCfg); err != nil {
	// 	t.Error(err)
	// }

	if err = onStor.DataDB().RemoveDestination(dst.Id, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	if _, rcvErr := onStor.DataDB().GetDestination(dst.Id, true, utils.NonTransactional); rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
}

func testOnStorITCRUDReverseDestinations(t *testing.T) {
	dst := &Destination{Id: "CRUDReverseDestination", Prefixes: []string{"+494", "+495", "+496"}}
	dst2 := &Destination{Id: "CRUDReverseDestination2", Prefixes: []string{"+497", "+498", "+499"}}
	if _, rcvErr := onStor.DataDB().GetReverseDestination(dst.Id, true, utils.NonTransactional); rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
	if err := onStor.DataDB().SetReverseDestination(dst, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	for i, _ := range dst.Prefixes {
		if rcv, err := onStor.DataDB().GetReverseDestination(dst.Prefixes[i], true, utils.NonTransactional); err != nil {
			t.Error(err)
		} else if !reflect.DeepEqual([]string{dst.Id}, rcv) {
			t.Errorf("Expecting: %v, received: %v", []string{dst.Id}, rcv)
		}
	}
	if err := onStor.DataDB().UpdateReverseDestination(dst, dst2, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	for i, _ := range dst.Prefixes {
		if rcv, err := onStor.DataDB().GetReverseDestination(dst2.Prefixes[i], true, utils.NonTransactional); err != nil {
			t.Error(err)
		} else if !reflect.DeepEqual([]string{dst2.Id}, rcv) {
			t.Errorf("Expecting: %v, received: %v", []string{dst.Id}, rcv)
		}
	}
	// FixMe
	// if err = onStor.DataDB().SelectDatabase("13"); err != nil {
	// 	t.Error(err)
	// }
	// if _, rcvErr := onStor.DataDB().GetReverseDestination(dst2.Id, false, utils.NonTransactional); rcvErr != utils.ErrNotFound {
	// 	t.Error(rcvErr)
	// }
	//
	for i, _ := range dst.Prefixes {
		if rcv, err := onStor.DataDB().GetReverseDestination(dst2.Prefixes[i], false, utils.NonTransactional); err != nil {
			t.Error(err)
		} else if !reflect.DeepEqual([]string{dst2.Id}, rcv) {
			t.Errorf("Expecting: %v, received: %v", []string{dst.Id}, rcv)
		}
	}
	// if err = onStor.DataDB().SelectDatabase(onStorCfg); err != nil {
	// 	t.Error(err)
	// }
}

func testOnStorITLCR(t *testing.T) {
	lcr := &LCR{
		Tenant:    "cgrates.org",
		Category:  "call",
		Direction: "*out",
		Account:   "*any",
		Subject:   "*any",
		Activations: []*LCRActivation{
			&LCRActivation{
				ActivationTime: time.Date(2012, 1, 1, 0, 0, 0, 0, time.UTC),
				Entries: []*LCREntry{
					&LCREntry{
						DestinationId:  "EU_LANDLINE",
						RPCategory:     "LCR_STANDARD",
						Strategy:       "*static",
						StrategyParams: "ivo;dan;rif",
						Weight:         10,
					},
					&LCREntry{
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

	if _, rcvErr := onStor.GetLCR(lcr.GetId(), false,
		utils.NonTransactional); rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
	if err := onStor.SetLCR(lcr, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	//get from cache
	if rcv, err := onStor.GetLCR(lcr.GetId(), false,
		utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(lcr, rcv) {
		t.Errorf("Expecting: %v, received: %v", lcr, rcv)
	}
	//get from database
	if rcv, err := onStor.GetLCR(lcr.GetId(), true,
		utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(lcr, rcv) {
		t.Errorf("Expecting: %v, received: %v", lcr, rcv)
	}
	expectedCLCR := []string{"lcr_*out:cgrates.org:call:*any:*any"}
	if itm, err := onStor.DataDB().GetKeysForPrefix(utils.LCR_PREFIX); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedCLCR, itm) {
		t.Errorf("Expected : %+v, but received %+v", expectedCLCR, itm)
	}
	//update
	lcr.Activations = []*LCRActivation{
		&LCRActivation{
			ActivationTime: time.Date(2012, 1, 1, 0, 0, 0, 0, time.UTC),
			Entries: []*LCREntry{
				&LCREntry{
					DestinationId:  "EU_LANDLINE",
					RPCategory:     "LCR_STANDARD",
					Strategy:       "*static",
					StrategyParams: "ivo;dan;teo",
					Weight:         10,
				},
				&LCREntry{
					DestinationId:  "*any",
					RPCategory:     "LCR_STANDARD",
					Strategy:       "*lowest_cost",
					StrategyParams: "",
					Weight:         20,
				},
			},
		},
	}
	if err := onStor.SetLCR(lcr, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	//get from cache
	if rcv, err := onStor.GetLCR(lcr.GetId(), false,
		utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(lcr, rcv) {
		t.Errorf("Expecting: %v, received: %v", lcr, rcv)
	}
	//get from database
	if rcv, err := onStor.GetLCR(lcr.GetId(), true,
		utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(lcr, rcv) {
		t.Errorf("Expecting: %v, received: %v", lcr, rcv)
	}
	if err := onStor.RemoveLCR(lcr.GetId(), utils.NonTransactional); err != nil {
		t.Error(err)
	}
	//check cache if removed
	if _, rcvErr := onStor.GetLCR(lcr.GetId(), false,
		utils.NonTransactional); rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
	//check database if removed
	if _, rcvErr := onStor.GetLCR(lcr.GetId(), true,
		utils.NonTransactional); rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
}

func testOnStorITActions(t *testing.T) {
	acts := Actions{
		&Action{
			Id:               "MINI",
			ActionType:       TOPUP_RESET,
			ExpirationString: UNLIMITED,
			Weight:           10,
			Balance: &BalanceFilter{
				Type:       utils.StringPointer(utils.MONETARY),
				Uuid:       utils.StringPointer(utils.GenUUID()),
				Directions: utils.StringMapPointer(utils.NewStringMap(utils.OUT)),
				Value: &utils.ValueFormula{Static: 10,
					Params: make(map[string]interface{})},
				Weight:   utils.Float64Pointer(10),
				Disabled: utils.BoolPointer(false),
				Timings: []*RITiming{
					&RITiming{
						Years:     utils.Years{2016, 2017},
						Months:    utils.Months{time.January, time.February, time.March},
						MonthDays: utils.MonthDays{1, 2, 3, 4},
						WeekDays:  utils.WeekDays{1, 2, 3},
						StartTime: utils.ASAP,
					},
				},
				Blocker: utils.BoolPointer(false),
			},
		},
		&Action{
			Id:               "MINI",
			ActionType:       TOPUP,
			ExpirationString: UNLIMITED,
			Weight:           10,
			Balance: &BalanceFilter{
				Type:       utils.StringPointer(utils.VOICE),
				Uuid:       utils.StringPointer(utils.GenUUID()),
				Directions: utils.StringMapPointer(utils.NewStringMap(utils.OUT)),
				Value: &utils.ValueFormula{Static: 100,
					Params: make(map[string]interface{})},
				Weight:         utils.Float64Pointer(10),
				RatingSubject:  utils.StringPointer("test"),
				DestinationIDs: utils.StringMapPointer(utils.NewStringMap("NAT")),
				Disabled:       utils.BoolPointer(false),
				Timings: []*RITiming{
					&RITiming{
						Years:     utils.Years{2016, 2017},
						Months:    utils.Months{time.January, time.February, time.March},
						MonthDays: utils.MonthDays{1, 2, 3, 4},
						WeekDays:  utils.WeekDays{1, 2, 3},
						StartTime: utils.ASAP,
					},
				},
				Blocker: utils.BoolPointer(false),
			},
		},
	}
	if _, rcvErr := onStor.GetActions(acts[0].Id,
		false, utils.NonTransactional); rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
	if err := onStor.SetActions(acts[0].Id,
		acts, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	//get from cache
	if rcv, err := onStor.GetActions(acts[0].Id,
		false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(acts[0], rcv[0]) {
		t.Errorf("Expecting: %v, received: %v", utils.ToJSON(acts[0]), utils.ToJSON(rcv[0]))
	}
	//get from database
	if rcv, err := onStor.GetActions(acts[0].Id,
		true, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(acts[0], rcv[0]) {
		t.Errorf("Expecting: %v, received: %v", utils.ToJSON(acts[0]), utils.ToJSON(rcv[0]))
	}
	expectedCA := []string{"act_MINI"}
	if itm, err := onStor.DataDB().GetKeysForPrefix(utils.ACTION_PREFIX); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedCA, itm) {
		t.Errorf("Expected : %+v, but received %+v", expectedCA, itm)
	}
	//update
	acts = Actions{
		&Action{
			Id:               "MINI",
			ActionType:       TOPUP_RESET,
			ExpirationString: UNLIMITED,
			Weight:           10,
			Balance: &BalanceFilter{
				Type:       utils.StringPointer(utils.MONETARY),
				Uuid:       utils.StringPointer(utils.GenUUID()),
				Directions: utils.StringMapPointer(utils.NewStringMap(utils.OUT)),
				Value: &utils.ValueFormula{Static: 10,
					Params: make(map[string]interface{})},
				Weight:   utils.Float64Pointer(10),
				Disabled: utils.BoolPointer(false),
				Timings: []*RITiming{
					&RITiming{
						Years:     utils.Years{2016, 2017},
						Months:    utils.Months{time.January, time.February, time.March},
						MonthDays: utils.MonthDays{1, 2, 3, 4},
						WeekDays:  utils.WeekDays{1, 2, 3},
						StartTime: utils.ASAP,
					},
				},
				Blocker: utils.BoolPointer(false),
			},
		},
		&Action{
			Id:               "MINI",
			ActionType:       TOPUP,
			ExpirationString: UNLIMITED,
			Weight:           10,
			Balance: &BalanceFilter{
				Type:       utils.StringPointer(utils.VOICE),
				Uuid:       utils.StringPointer(utils.GenUUID()),
				Directions: utils.StringMapPointer(utils.NewStringMap(utils.OUT)),
				Value: &utils.ValueFormula{Static: 100,
					Params: make(map[string]interface{})},
				Weight:         utils.Float64Pointer(10),
				RatingSubject:  utils.StringPointer("test"),
				DestinationIDs: utils.StringMapPointer(utils.NewStringMap("NAT")),
				Disabled:       utils.BoolPointer(false),
				Timings: []*RITiming{
					&RITiming{
						Years:     utils.Years{2016, 2017},
						Months:    utils.Months{time.January, time.February, time.March},
						MonthDays: utils.MonthDays{1, 2, 3, 4},
						WeekDays:  utils.WeekDays{1, 2, 3},
						StartTime: utils.ASAP,
					},
				},
				Blocker: utils.BoolPointer(false),
			},
		},
		&Action{
			Id:               "MINI",
			ActionType:       DEBIT,
			ExpirationString: UNLIMITED,
			Weight:           20,
			Balance: &BalanceFilter{
				Type:       utils.StringPointer(utils.VOICE),
				Uuid:       utils.StringPointer(utils.GenUUID()),
				Directions: utils.StringMapPointer(utils.NewStringMap(utils.OUT)),
				Value: &utils.ValueFormula{Static: 200,
					Params: make(map[string]interface{})},
				Weight:         utils.Float64Pointer(20),
				RatingSubject:  utils.StringPointer("test"),
				DestinationIDs: utils.StringMapPointer(utils.NewStringMap("NAT")),
				Disabled:       utils.BoolPointer(false),
				Timings: []*RITiming{
					&RITiming{
						Years:     utils.Years{2016, 2017},
						Months:    utils.Months{time.January, time.February, time.March},
						MonthDays: utils.MonthDays{1, 2, 3, 4},
						WeekDays:  utils.WeekDays{1, 2, 3},
						StartTime: utils.ASAP,
					},
				},
				Blocker: utils.BoolPointer(false),
			},
		},
	}
	time.Sleep(sleepDelay)
	if err := onStor.SetActions(acts[0].Id,
		acts, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	//get from cache
	if rcv, err := onStor.GetActions(acts[0].Id,
		false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(acts[0], rcv[0]) {
		t.Errorf("Expecting: %v, received: %v", utils.ToJSON(acts[0]), utils.ToJSON(rcv[0]))
	}
	//get from database
	if rcv, err := onStor.GetActions(acts[0].Id,
		true, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(acts[0], rcv[0]) {
		t.Errorf("Expecting: %v, received: %v", utils.ToJSON(acts[0]), utils.ToJSON(rcv[0]))
	}
	if err := onStor.RemoveActions(acts[0].Id, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	//check cache if removed
	if _, rcvErr := onStor.GetActions(acts[0].Id,
		false, utils.NonTransactional); rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
	//check database if removed
	if _, rcvErr := onStor.GetActions(acts[0].Id,
		true, utils.NonTransactional); rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
}

func testOnStorITSharedGroup(t *testing.T) {
	sg := &SharedGroup{
		Id: "SG2",
		AccountParameters: map[string]*SharingParameters{
			"*any": &SharingParameters{
				Strategy:      "*lowest",
				RatingSubject: "",
			},
		},
		MemberIds: make(utils.StringMap),
	}
	if _, rcvErr := onStor.GetSharedGroup(sg.Id, false,
		utils.NonTransactional); rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
	if err := onStor.SetSharedGroup(sg, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	//get from cache
	if rcv, err := onStor.GetSharedGroup(sg.Id, false,
		utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(sg, rcv) {
		t.Errorf("Expecting: %v, received: %v", sg, rcv)
	}
	//get from database
	if rcv, err := onStor.GetSharedGroup(sg.Id, true,
		utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(sg, rcv) {
		t.Errorf("Expecting: %v, received: %v", sg, rcv)
	}
	//update
	sg.AccountParameters = map[string]*SharingParameters{
		"*any": &SharingParameters{
			Strategy:      "*lowest",
			RatingSubject: "",
		},
		"*any2": &SharingParameters{
			Strategy:      "*lowest2",
			RatingSubject: "",
		},
	}
	if err := onStor.SetSharedGroup(sg, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	//get from cache
	if rcv, err := onStor.GetSharedGroup(sg.Id, false,
		utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(sg, rcv) {
		t.Errorf("Expecting: %v, received: %v", sg, rcv)
	}
	//get from database
	if rcv, err := onStor.GetSharedGroup(sg.Id, true,
		utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(sg, rcv) {
		t.Errorf("Expecting: %v, received: %v", sg, rcv)
	}
	if err := onStor.RemoveSharedGroup(sg.Id, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	//check cache if removed
	if _, rcvErr := onStor.GetSharedGroup(sg.Id, false,
		utils.NonTransactional); rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
	//check database if removed
	if _, rcvErr := onStor.GetSharedGroup(sg.Id, true,
		utils.NonTransactional); rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
}

func testOnStorITCRUDActionTriggers(t *testing.T) {
	ats := ActionTriggers{
		&ActionTrigger{
			ID:                "testOnStorITCRUDActionTriggers",
			Balance:           &BalanceFilter{Type: utils.StringPointer(utils.MONETARY), Directions: utils.StringMapPointer(utils.NewStringMap(utils.OUT)), Timings: make([]*RITiming, 0)},
			ThresholdValue:    2,
			ThresholdType:     utils.TRIGGER_MAX_EVENT_COUNTER,
			ActionsID:         "TEST_ACTIONS",
			LastExecutionTime: time.Date(2012, 1, 1, 0, 0, 0, 0, time.UTC),
			ExpirationDate:    time.Date(2012, 1, 1, 0, 0, 0, 0, time.UTC),
			ActivationDate:    time.Date(2012, 1, 1, 0, 0, 0, 0, time.UTC)},
	}
	atsID := ats[0].ID
	if _, rcvErr := onStor.GetActionTriggers(atsID, true, utils.NonTransactional); rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
	if err := onStor.SetActionTriggers(atsID, ats, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	if rcv, err := onStor.GetActionTriggers(atsID, true, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(ats[0], rcv[0]) {
		t.Errorf("Expecting: %v, received: %v", ats[0], rcv[0])
	}
	// FixMe
	// if err = onStor.DataDB().SelectDatabase("13"); err != nil {
	// 	t.Error(err)
	// }
	// if _, rcvErr := onStor.DataDB().GetActionTriggers(sg.Id, false, utils.NonTransactional); rcvErr != utils.ErrNotFound {
	// 	t.Error(rcvErr)
	// }
	//
	if rcv, err := onStor.GetActionTriggers(atsID, false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(ats[0], rcv[0]) {
		t.Errorf("Expecting: %v, received: %v", ats[0], rcv[0])
	}
	// if err = onStor.DataDB().SelectDatabase(onStorCfg); err != nil {
	// 	t.Error(err)
	// }

	if err := onStor.RemoveActionTriggers(atsID, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	if _, rcvErr := onStor.GetActionTriggers(atsID, true, utils.NonTransactional); rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
}

func testOnStorITCRUDActionPlan(t *testing.T) {
	ap := &ActionPlan{
		Id:         "MORE_MINUTES2",
		AccountIDs: utils.StringMap{"vdf:minitsboy": true},
		ActionTimings: []*ActionTiming{
			&ActionTiming{
				Uuid: utils.GenUUID(),
				Timing: &RateInterval{
					Timing: &RITiming{
						Years:     utils.Years{2012},
						Months:    utils.Months{},
						MonthDays: utils.MonthDays{},
						WeekDays:  utils.WeekDays{},
						StartTime: utils.ASAP,
					},
				},
				Weight:    10,
				ActionsID: "MINI",
			},
			&ActionTiming{
				Uuid: utils.GenUUID(),
				Timing: &RateInterval{
					Timing: &RITiming{
						Years:     utils.Years{2012},
						Months:    utils.Months{},
						MonthDays: utils.MonthDays{},
						WeekDays:  utils.WeekDays{},
						StartTime: utils.ASAP,
					},
				},
				Weight:    10,
				ActionsID: "SHARED",
			},
		},
	}
	if _, rcvErr := onStor.DataDB().GetActionPlan(ap.Id, true, utils.NonTransactional); rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
	if err := onStor.DataDB().SetActionPlan(ap.Id, ap, true, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	if rcv, err := onStor.DataDB().GetActionPlan(ap.Id, true, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(ap, rcv) {
		t.Errorf("Expecting: %v, received: %v", ap, rcv)
	}
	// FixMe
	// if err = onStor.DataDB().SelectDatabase("13"); err != nil {
	// 	t.Error(err)
	// }
	// if _, rcvErr := onStor.DataDB().GetActionPlan(ap.Id, false, utils.NonTransactional); rcvErr != utils.ErrNotFound {
	// 	t.Error(rcvErr)
	// }
	//
	// if rcv, err := onStor.DataDB().GetActionPlan(ap.Id, false, utils.NonTransactional); err != nil {
	// 	t.Error(err)
	// } else if !reflect.DeepEqual(ap, rcv) {
	// 	t.Errorf("Expecting: %v, received: %v", ap, rcv)
	// }
	// if err = onStor.DataDB().SelectDatabase(onStorCfg); err != nil {
	// 	t.Error(err)
	// }
	if rcv, err := onStor.DataDB().GetAllActionPlans(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(ap.Id, rcv[ap.Id].Id) {
		t.Errorf("Expecting: %v, received: %v", ap.Id, rcv[ap.Id].Id)
	}
}

func testOnStorITCRUDAccountActionPlans(t *testing.T) {
	acntID := utils.ConcatenatedKey("cgrates.org2", "1001")
	expect := []string{"PACKAGE_10_SHARED_A_5", "USE_SHARED_A", "apl_PACKAGE_1001"}
	aAPs := []string{"PACKAGE_10_SHARED_A_5", "apl_PACKAGE_1001"}
	aAPs2 := []string{"USE_SHARED_A"}
	if _, rcvErr := onStor.DataDB().GetAccountActionPlans(acntID, true, utils.NonTransactional); rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
	if err := onStor.DataDB().SetAccountActionPlans(acntID, aAPs, true); err != nil {
		t.Error(err)
	}
	if rcv, err := onStor.DataDB().GetAccountActionPlans(acntID, true, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(aAPs, rcv) {
		t.Errorf("Expecting: %v, received: %v", aAPs, rcv)
	}
	if err := onStor.DataDB().SetAccountActionPlans(acntID, aAPs2, false); err != nil {
		t.Error(err)
	}
	if rcv, err := onStor.DataDB().GetAccountActionPlans(acntID, true, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expect, rcv) {
		t.Errorf("Expecting: %v, received: %v", expect, rcv)
	}
	// FixMe
	// if err = onStor.DataDB().SelectDatabase("13"); err != nil {
	// 	t.Error(err)
	// }
	// if _, rcvErr := onStor.DataDB().GetAccountActionPlans(acntID, false, utils.NonTransactional); rcvErr != utils.ErrNotFound {
	// 	t.Error(rcvErr)
	// }
	//
	if rcv, err := onStor.DataDB().GetAccountActionPlans(acntID, false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expect, rcv) {
		t.Errorf("Expecting: %v, received: %v", expect, rcv)
	}
	// if err = onStor.DataDB().SelectDatabase(onStorCfg); err != nil {
	// 	t.Error(err)
	// }
	if err := onStor.DataDB().RemAccountActionPlans(acntID, aAPs2); err != nil {
		t.Error(err)
	}
	if rcv, err := onStor.DataDB().GetAccountActionPlans(acntID, true, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(aAPs, rcv) {
		t.Errorf("Expecting: %v, received: %v", aAPs, rcv)
	}
	if err := onStor.DataDB().RemAccountActionPlans(acntID, aAPs); err != nil {
		t.Error(err)
	}
	if _, rcvErr := onStor.DataDB().GetAccountActionPlans(acntID, true, utils.NonTransactional); rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
}

func testOnStorITCRUDAccount(t *testing.T) {
	acc := &Account{
		ID:         utils.ConcatenatedKey("cgrates.org", "account2"),
		BalanceMap: map[string]Balances{utils.MONETARY: Balances{&Balance{Value: 10, Weight: 10}}},
	}
	if _, rcvErr := onStor.DataDB().GetAccount(acc.ID); rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
	if err := onStor.DataDB().SetAccount(acc); err != nil {
		t.Error(err)
	}
	if rcv, err := onStor.DataDB().GetAccount(acc.ID); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(acc.ID, rcv.ID) {
		t.Errorf("Expecting: %v, received: %v", acc.ID, rcv.ID)
	} else if !reflect.DeepEqual(acc.BalanceMap[utils.MONETARY][0].Value, rcv.BalanceMap[utils.MONETARY][0].Value) {
		t.Errorf("Expecting: %v, received: %v", acc.BalanceMap[utils.MONETARY][0].Value, rcv.BalanceMap[utils.MONETARY][0].Value)
	} else if !reflect.DeepEqual(acc.BalanceMap[utils.MONETARY][0].Weight, rcv.BalanceMap[utils.MONETARY][0].Weight) {
		t.Errorf("Expecting: %v, received: %v", acc.BalanceMap[utils.MONETARY][0].Weight, rcv.BalanceMap[utils.MONETARY][0].Weight)
	}
	if err := onStor.DataDB().RemoveAccount(acc.ID); err != nil {
		t.Error(err)
	}
	if _, rcvErr := onStor.DataDB().GetAccount(acc.ID); rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
}

func testOnStorITCRUDCdrStatsQueue(t *testing.T) {
	sq := &CDRStatsQueue{
		conf: &CdrStats{Id: "TTT"},
		Cdrs: []*QCdr{
			&QCdr{Cost: 9.0,
				SetupTime:  time.Date(2012, 1, 1, 0, 0, 0, 0, time.UTC),
				AnswerTime: time.Date(2012, 1, 1, 0, 0, 10, 0, time.UTC),
				EventTime:  time.Date(2012, 1, 1, 0, 0, 0, 0, time.UTC),
			}},
	}
	if _, rcvErr := onStor.GetCdrStatsQueue(sq.GetId()); rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
	if err := onStor.SetCdrStatsQueue(sq); err != nil {
		t.Error(err)
	}
	if rcv, err := onStor.GetCdrStatsQueue(sq.GetId()); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(sq.Cdrs, rcv.Cdrs) {
		t.Errorf("Expecting: %v, received: %v", sq.Cdrs, rcv.Cdrs)
	}
	if err := onStor.RemoveCdrStatsQueue(sq.GetId()); err != nil {
		t.Error(err)
	}
	if _, rcvErr := onStor.GetCdrStatsQueue(sq.GetId()); rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
}

func testOnStorITCRUDSubscribers(t *testing.T) {
	if sbs, err := onStor.GetSubscribers(); err != nil {
		t.Error(err)
	} else if len(sbs) != 0 {
		t.Errorf("Received subscribers: %+v", sbs)
	}
	sbsc := &SubscriberData{
		ExpTime: time.Date(2012, 1, 1, 0, 0, 0, 0, time.UTC),
		Filters: utils.ParseRSRFieldsMustCompile("^*default", utils.INFIELD_SEP)}
	sbscID := "testOnStorITCRUDSubscribers"
	if err := onStor.SetSubscriber(sbscID, sbsc); err != nil {
		t.Error(err)
	}
	if rcv, err := onStor.GetSubscribers(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(sbsc.ExpTime, rcv[sbscID].ExpTime) { // Test just ExpTime since RSRField is more complex behind
		t.Errorf("Expecting: %v, received: %v", sbsc, rcv[sbscID])
	}
	if err := onStor.RemoveSubscriber(sbscID); err != nil {
		t.Error(err)
	}
	if sbs, err := onStor.GetSubscribers(); err != nil {
		t.Error(err)
	} else if len(sbs) != 0 {
		t.Errorf("Received subscribers: %+v", sbs)
	}
}

func testOnStorITCRUDUser(t *testing.T) {
	usr := &UserProfile{
		Tenant:   "test",
		UserName: "user",
		Profile: map[string]string{
			"t": "v",
		},
	}
	if _, rcvErr := onStor.GetUser(usr.GetId()); rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
	if err := onStor.SetUser(usr); err != nil {
		t.Error(err)
	}
	if rcv, err := onStor.GetUser(usr.GetId()); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(usr, rcv) {
		t.Errorf("Expecting: %v, received: %v", usr, rcv)
	}
	if rcv, err := onStor.GetUsers(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(usr, rcv[0]) {
		t.Errorf("Expecting: %v, received: %v", usr, rcv[0])
	}
	if err := onStor.RemoveUser(usr.GetId()); err != nil {
		t.Error(err)
	}
	if _, rcvErr := onStor.GetUser(usr.GetId()); rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
}

func testOnStorITCRUDAlias(t *testing.T) {
	als := &Alias{
		Direction: "*out",
		Tenant:    "cgrates.org2",
		Category:  "call",
		Account:   "dan",
		Subject:   "dan",
		Context:   "*rating",
		Values: AliasValues{
			&AliasValue{
				DestinationId: "EU_LANDLINE",
				Pairs: AliasPairs{
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

			&AliasValue{
				DestinationId: "GLOBAL2",
				Pairs:         AliasPairs{"Subject": map[string]string{"dan": "dan2"}},
				Weight:        20,
			},
		},
	}

	if _, rcvErr := onStor.DataDB().GetAlias(als.GetId(), true, utils.NonTransactional); rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
	if err := onStor.DataDB().SetAlias(als, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	if rcv, err := onStor.DataDB().GetAlias(als.GetId(), true, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(als, rcv) {
		t.Errorf("Expecting: %v, received: %v", als, rcv)
	}
	// FixMe
	// if err = onStor.DataDB().SelectDatabase("13"); err != nil {
	// 	t.Error(err)
	// }
	// if _, rcvErr := onStor.DataDB().GetAlias(als.GetId(), false, utils.NonTransactional); rcvErr != utils.ErrNotFound {
	// 	t.Error(rcvErr)
	// }
	//
	if rcv, err := onStor.DataDB().GetAlias(als.GetId(), false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(als, rcv) {
		t.Errorf("Expecting: %v, received: %v", als, rcv)
	}
	// if err = onStor.DataDB().SelectDatabase(onStorCfg); err != nil {
	// 	t.Error(err)
	// }
	if err := onStor.DataDB().RemoveAlias(als.GetId(), utils.NonTransactional); err != nil {
		t.Error(err)
	}
	if _, rcvErr := onStor.DataDB().GetAlias(als.GetId(), true, utils.NonTransactional); rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
}

func testOnStorITCRUDReverseAlias(t *testing.T) {
	als := &Alias{
		Direction: "*out",
		Tenant:    "itsyscom.com",
		Category:  "call",
		Account:   "testOnStorITCRUDReverseAlias",
		Subject:   "testOnStorITCRUDReverseAlias",
		Context:   "*rating",
		Values: AliasValues{
			&AliasValue{
				DestinationId: "EU",
				Pairs: AliasPairs{
					"Account": map[string]string{
						"dan": "testOnStorITCRUDReverseAlias1",
						"rif": "testOnStorITCRUDReverseAlias2",
					},
					"Calling": map[string]string{
						"11234": "2234",
					},
				},
				Weight: 10,
			},

			&AliasValue{
				DestinationId: "US",
				Pairs:         AliasPairs{"Account": map[string]string{"dan": "testOnStorITCRUDReverseAlias3"}},
				Weight:        20,
			},
		},
	}
	rvAlsID := strings.Join([]string{als.Values[1].Pairs["Account"]["dan"], "Account", als.Context}, "")
	exp := strings.Join([]string{als.Direction, ":", als.Tenant, ":", als.Category, ":", als.Account, ":", als.Subject, ":", als.Context, ":", als.Values[1].DestinationId}, "")
	if _, rcvErr := onStor.DataDB().GetReverseAlias(rvAlsID, true, utils.NonTransactional); rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
	if err := onStor.DataDB().SetReverseAlias(als, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	if rcv, err := onStor.DataDB().GetReverseAlias(rvAlsID, true, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(exp, rcv[0]) {
		t.Errorf("Expecting: %v, received: %v", exp, rcv[0])
	}
	// FixMe
	// if err = onStor.DataDB().SelectDatabase("13"); err != nil {
	// 	t.Error(err)
	// }
	// if _, rcvErr := onStor.DataDB().GetReverseAlias(rvAlsID, false, utils.NonTransactional); rcvErr != utils.ErrNotFound {
	// 	t.Error(rcvErr)
	// }
	//
	if rcv, err := onStor.DataDB().GetReverseAlias(rvAlsID, false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(exp, rcv[0]) {
		t.Errorf("Expecting: %v, received: %v", exp, rcv[0])
	}
	// if err = onStor.DataDB().SelectDatabase(onStorCfg); err != nil {
	// 	t.Error(err)
	// }
}

func testOnStorITResourceProfile(t *testing.T) {
	rL := &ResourceProfile{
		Tenant:    "cgrates.org",
		ID:        "RL_TEST2",
		Weight:    10,
		FilterIDs: []string{"FLTR_RES_RL_TEST2"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 3, 13, 43, 0, 0, time.UTC),
			ExpiryTime:     time.Date(2015, 7, 3, 13, 43, 0, 0, time.UTC)},
		Limit:      1,
		Thresholds: []string{"TEST_ACTIONS"},
		UsageTTL:   time.Duration(3 * time.Nanosecond),
	}
	if _, rcvErr := onStor.GetResourceProfile(rL.Tenant, rL.ID,
		true, utils.NonTransactional); rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
	if err := onStor.SetResourceProfile(rL, false); err != nil {
		t.Error(err)
	}
	//get from cache
	if rcv, err := onStor.GetResourceProfile(rL.Tenant, rL.ID,
		false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rL, rcv) {
		t.Errorf("Expecting: %v, received: %v", rL, rcv)
	}
	//get from database
	if rcv, err := onStor.GetResourceProfile(rL.Tenant, rL.ID,
		true, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rL, rcv) {
		t.Errorf("Expecting: %v, received: %v", rL, rcv)
	}
	expectedR := []string{"rsp_cgrates.org:RL_TEST2"}
	if itm, err := onStor.DataDB().GetKeysForPrefix(utils.ResourceProfilesPrefix); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedR, itm) {
		t.Errorf("Expected : %+v, but received %+v", expectedR, itm)
	}
	//update
	rL.Thresholds = []string{"TH1", "TH2"}
	time.Sleep(sleepDelay)
	if err := onStor.SetResourceProfile(rL, false); err != nil {
		t.Error(err)
	}
	//get from cache
	if rcv, err := onStor.GetResourceProfile(rL.Tenant, rL.ID,
		false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rL, rcv) {
		t.Errorf("Expecting: %v, received: %v", rL, rcv)
	}
	//get from database
	if rcv, err := onStor.GetResourceProfile(rL.Tenant, rL.ID,
		true, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rL, rcv) {
		t.Errorf("Expecting: %v, received: %v", rL, rcv)
	}
	if err := onStor.RemoveResourceProfile(rL.Tenant, rL.ID,
		utils.NonTransactional, false); err != nil {
		t.Error(err)
	}
	//check cache if removed
	if _, rcvErr := onStor.GetResourceProfile(rL.Tenant, rL.ID,
		false, utils.NonTransactional); rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
	//check database if removed
	if _, rcvErr := onStor.GetResourceProfile(rL.Tenant, rL.ID,
		true, utils.NonTransactional); rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
}

func testOnStorITResource(t *testing.T) {
	res := &Resource{
		Tenant: "cgrates.org",
		ID:     "RL1",
		Usages: map[string]*ResourceUsage{
			"RU1": &ResourceUsage{
				ID:         "RU1",
				ExpiryTime: time.Date(2014, 7, 3, 13, 43, 0, 0, time.UTC),
				Units:      2,
			},
		},
		TTLIdx: []string{"RU1"},
	}
	if _, rcvErr := onStor.GetResource(res.Tenant, res.ID,
		false, utils.NonTransactional); rcvErr != nil && rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
	if err := onStor.SetResource(res); err != nil {
		t.Error(err)
	}
	//get from cache
	if rcv, err := onStor.GetResource("cgrates.org", "RL1",
		false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !(reflect.DeepEqual(res, rcv)) {
		t.Errorf("Expecting: %v, received: %v", res, rcv)
	}
	//get from database
	if rcv, err := onStor.GetResource("cgrates.org", "RL1",
		true, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !(reflect.DeepEqual(res, rcv)) {
		t.Errorf("Expecting: %v, received: %v", res, rcv)
	}
	expectedT := []string{"res_cgrates.org:RL1"}
	if itm, err := onStor.DataDB().GetKeysForPrefix(utils.ResourcesPrefix); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedT, itm) {
		t.Errorf("Expected : %+v, but received %+v", expectedT, itm)
	}
	//update
	res.TTLIdx = []string{"RU1", "RU2"}
	time.Sleep(sleepDelay)
	if err := onStor.SetResource(res); err != nil {
		t.Error(err)
	}
	//get from cache
	if rcv, err := onStor.GetResource("cgrates.org", "RL1",
		false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !(reflect.DeepEqual(res, rcv)) {
		t.Errorf("Expecting: %v, received: %v", res, rcv)
	}
	//get from database
	if rcv, err := onStor.GetResource("cgrates.org", "RL1",
		true, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !(reflect.DeepEqual(res, rcv)) {
		t.Errorf("Expecting: %v, received: %v", res, rcv)
	}
	if err := onStor.RemoveResource(res.Tenant, res.ID, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	//check cache if removed
	if _, rcvErr := onStor.GetResource(res.Tenant, res.ID,
		false, utils.NonTransactional); rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
	//check database if removed
	if _, rcvErr := onStor.GetResource(res.Tenant, res.ID,
		true, utils.NonTransactional); rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
}

func testOnStorITTiming(t *testing.T) {
	tmg := &utils.TPTiming{
		ID:        "TEST",
		Years:     utils.Years{2016, 2017},
		Months:    utils.Months{time.January, time.February, time.March},
		MonthDays: utils.MonthDays{1, 2, 3, 4},
		WeekDays:  utils.WeekDays{1, 2, 3},
		StartTime: "00:00:00",
		EndTime:   "",
	}
	if _, rcvErr := onStor.GetTiming(tmg.ID, false,
		utils.NonTransactional); rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
	if err := onStor.SetTiming(tmg); err != nil {
		t.Error(err)
	}
	//get from cache
	if rcv, err := onStor.GetTiming(tmg.ID, false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(tmg, rcv) {
		t.Errorf("Expecting: %v, received: %v", tmg, rcv)
	}
	//get from database
	if rcv, err := onStor.GetTiming(tmg.ID, true, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(tmg, rcv) {
		t.Errorf("Expecting: %v, received: %v", tmg, rcv)
	}
	expectedT := []string{"tmg_TEST"}
	if itm, err := onStor.DataDB().GetKeysForPrefix(utils.TimingsPrefix); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedT, itm) {
		t.Errorf("Expected : %+v, but received %+v", expectedT, itm)
	}
	//update
	tmg.MonthDays = utils.MonthDays{1, 2, 3, 4, 5, 6, 7}
	time.Sleep(sleepDelay)
	if err := onStor.SetTiming(tmg); err != nil {
		t.Error(err)
	}
	//get from cache
	if rcv, err := onStor.GetTiming(tmg.ID, false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(tmg, rcv) {
		t.Errorf("Expecting: %v, received: %v", tmg, rcv)
	}
	//get from database
	if rcv, err := onStor.GetTiming(tmg.ID, true, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(tmg, rcv) {
		t.Errorf("Expecting: %v, received: %v", tmg, rcv)
	}
	if err := onStor.RemoveTiming(tmg.ID, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	//check cache if removed
	if _, rcvErr := onStor.GetTiming(tmg.ID, false,
		utils.NonTransactional); rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
	//check database if removed
	if _, rcvErr := onStor.GetTiming(tmg.ID, true,
		utils.NonTransactional); rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
}

func testOnStorITCRUDHistory(t *testing.T) {
	time := time.Date(2012, 1, 1, 0, 0, 0, 0, time.UTC)
	ist := &utils.LoadInstance{"Load", "RatingLoad", "Account", time}
	if err := onStor.DataDB().AddLoadHistory(ist, 1, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	if rcv, err := onStor.DataDB().GetLoadHistory(1, true, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(ist, rcv[0]) {
		t.Errorf("Expecting: %v, received: %v", ist, rcv[0])
	}
}

func testOnStorITCRUDStructVersion(t *testing.T) {
	CurrentVersion := Versions{utils.Accounts: 2, utils.Actions: 2, utils.ActionTriggers: 2, utils.ActionPlans: 2, utils.SharedGroups: 2, utils.COST_DETAILS: 2}
	if _, rcvErr := onStor.DataDB().GetVersions(utils.TBLVersions); rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
	if err := onStor.DataDB().SetVersions(CurrentVersion, false); err != nil {
		t.Error(err)
	}
	if rcv, err := onStor.DataDB().GetVersions(utils.TBLVersions); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(CurrentVersion, rcv) {
		t.Errorf("Expecting: %v, received: %v", CurrentVersion, rcv)
	} else if err = onStor.DataDB().RemoveVersions(rcv); err != nil {
		t.Error(err)
	}
	if _, rcvErr := onStor.DataDB().GetVersions(utils.TBLVersions); rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
}

func testOnStorITStatQueueProfile(t *testing.T) {
	sq := &StatQueueProfile{
		Tenant:             "cgrates.org",
		ID:                 "test",
		ActivationInterval: &utils.ActivationInterval{},
		FilterIDs:          []string{"FLTR_1"},
		QueueLength:        2,
		TTL:                time.Duration(0 * time.Second),
		Metrics: []*utils.MetricWithParams{
			&utils.MetricWithParams{},
		},
		Stored:     true,
		Thresholds: []string{"Thresh1"},
	}
	if _, rcvErr := onStor.GetStatQueueProfile(sq.Tenant, sq.ID,
		false, utils.NonTransactional); rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
	if err := onStor.SetStatQueueProfile(sq, false); err != nil {
		t.Error(err)
	}
	//get from cache
	if rcv, err := onStor.GetStatQueueProfile(sq.Tenant,
		sq.ID, false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(sq, rcv) {
		t.Errorf("Expecting: %v, received: %v", sq, rcv)
	}
	//get from database
	if rcv, err := onStor.GetStatQueueProfile(sq.Tenant,
		sq.ID, true, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(sq, rcv) {
		t.Errorf("Expecting: %v, received: %v", sq, rcv)
	}
	expectedR := []string{"sqp_cgrates.org:test"}
	if itm, err := onStor.DataDB().GetKeysForPrefix(utils.StatQueueProfilePrefix); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedR, itm) {
		t.Errorf("Expected : %+v, but received %+v", expectedR, itm)
	}
	//update
	sq.Thresholds = []string{"TH1", "TH2"}
	time.Sleep(sleepDelay)
	if err := onStor.SetStatQueueProfile(sq, false); err != nil {
		t.Error(err)
	}
	//get from cache
	if rcv, err := onStor.GetStatQueueProfile(sq.Tenant,
		sq.ID, false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(sq, rcv) {
		t.Errorf("Expecting: %v, received: %v", utils.ToJSON(sq), utils.ToJSON(rcv))
	}
	//get from database
	if rcv, err := onStor.GetStatQueueProfile(sq.Tenant,
		sq.ID, true, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(sq, rcv) {
		t.Errorf("Expecting: %v, received: %v", utils.ToJSON(sq), utils.ToJSON(rcv))
	}
	if err := onStor.RemoveStatQueueProfile(sq.Tenant, sq.ID,
		utils.NonTransactional, false); err != nil {
		t.Error(err)
	}
	//check cache if removed
	if _, rcvErr := onStor.GetStatQueueProfile(sq.Tenant,
		sq.ID, false, utils.NonTransactional); rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
	//check database if removed
	if _, rcvErr := onStor.GetStatQueueProfile(sq.Tenant,
		sq.ID, true, utils.NonTransactional); rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
}

func testOnStorITStatQueue(t *testing.T) {
	eTime := utils.TimePointer(time.Date(2013, 10, 1, 0, 0, 0, 0, time.UTC))
	sq := &StatQueue{
		Tenant: "cgrates.org",
		ID:     "Test_StatQueue",
		SQItems: []struct {
			EventID    string     // Bounded to the original StatEvent
			ExpiryTime *time.Time // Used to auto-expire events
		}{{EventID: "cgrates.org:ev1", ExpiryTime: eTime},
			{EventID: "cgrates.org:ev2", ExpiryTime: eTime},
			{EventID: "cgrates.org:ev3", ExpiryTime: eTime}},
		SQMetrics: map[string]StatMetric{
			utils.MetaASR: &StatASR{
				Answered: 2,
				Count:    3,
				Events: map[string]bool{
					"cgrates.org:ev1": true,
					"cgrates.org:ev2": true,
					"cgrates.org:ev3": false,
				},
			},
		},
	}
	if _, rcvErr := onStor.GetStatQueue(sq.Tenant, sq.ID,
		false, utils.NonTransactional); rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
	if err := onStor.SetStatQueue(sq); err != nil {
		t.Error(err)
	}
	//get from cache
	if rcv, err := onStor.GetStatQueue(sq.Tenant,
		sq.ID, false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(sq, rcv) {
		t.Errorf("Expecting: %v, received: %v", sq, rcv)
	}
	//get from database
	if rcv, err := onStor.GetStatQueue(sq.Tenant,
		sq.ID, true, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(sq, rcv) {
		t.Errorf("Expecting: %v, received: %v", sq, rcv)
	}
	expectedT := []string{"stq_cgrates.org:Test_StatQueue"}
	if itm, err := onStor.DataDB().GetKeysForPrefix(utils.StatQueuePrefix); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedT, itm) {
		t.Errorf("Expected : %+v, but received %+v", expectedT, itm)
	}
	//update
	sq.SQMetrics = map[string]StatMetric{
		utils.MetaASR: &StatASR{
			Answered: 3,
			Count:    3,
			Events: map[string]bool{
				"cgrates.org:ev1": true,
				"cgrates.org:ev2": true,
				"cgrates.org:ev3": true,
			},
		},
	}
	time.Sleep(sleepDelay)
	if err := onStor.SetStatQueue(sq); err != nil {
		t.Error(err)
	}
	//get from cache
	if rcv, err := onStor.GetStatQueue(sq.Tenant,
		sq.ID, false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(sq, rcv) {
		t.Errorf("Expecting: %v, received: %v", sq, rcv)
	}
	//get from database
	if rcv, err := onStor.GetStatQueue(sq.Tenant,
		sq.ID, true, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(sq, rcv) {
		t.Errorf("Expecting: %v, received: %v", sq, rcv)
	}
	if err := onStor.RemStatQueue(sq.Tenant, sq.ID,
		utils.NonTransactional); err != nil {
		t.Error(err)
	}
	//check cache if removed
	if _, rcvErr := onStor.GetStatQueue(sq.Tenant,
		sq.ID, false, utils.NonTransactional); rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
	//check database if removed
	if _, rcvErr := onStor.GetStatQueue(sq.Tenant,
		sq.ID, true, utils.NonTransactional); rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
}

func testOnStorITThresholdProfile(t *testing.T) {
	fp := &Filter{
		Tenant: "cgrates.org",
		ID:     "TestFilter2",
		RequestFilters: []*RequestFilter{
			&RequestFilter{
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
	th := &ThresholdProfile{
		Tenant:             "cgrates.org",
		ID:                 "test",
		ActivationInterval: &utils.ActivationInterval{},
		FilterIDs:          []string{"TestFilter2"},
		Recurrent:          true,
		MinSleep:           time.Duration(0 * time.Second),
		Blocker:            true,
		Weight:             1.4,
		ActionIDs:          []string{"Action1"},
	}
	if err := onStor.SetFilter(fp); err != nil {
		t.Error(err)
	}
	if _, rcvErr := onStor.GetThresholdProfile(th.Tenant, th.ID,
		false, utils.NonTransactional); rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
	if err := onStor.SetThresholdProfile(th, true); err != nil {
		t.Error(err)
	}
	//get from cache
	if rcv, err := onStor.GetThresholdProfile(th.Tenant, th.ID,
		false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(th, rcv) {
		t.Errorf("Expecting: %v, received: %v", th, rcv)
	}
	//get from database
	if rcv, err := onStor.GetThresholdProfile(th.Tenant, th.ID,
		true, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(th, rcv) {
		t.Errorf("Expecting: %v, received: %v", th, rcv)
	}
	expectedR := []string{"thp_cgrates.org:test"}
	if itm, err := onStor.DataDB().GetKeysForPrefix(utils.ThresholdProfilePrefix); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedR, itm) {
		t.Errorf("Expected : %+v, but received %+v", expectedR, itm)
	}
	//update
	th.ActionIDs = []string{"Action1", "Action2"}
	time.Sleep(sleepDelay)
	if err := onStor.SetThresholdProfile(th, true); err != nil {
		t.Error(err)
	}
	//get from cache
	if rcv, err := onStor.GetThresholdProfile(th.Tenant, th.ID,
		false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(th, rcv) {
		t.Errorf("Expecting: %v, received: %v", utils.ToJSON(th), utils.ToJSON(rcv))
	}
	//get from database
	if rcv, err := onStor.GetThresholdProfile(th.Tenant, th.ID,
		true, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(th, rcv) {
		t.Errorf("Expecting: %v, received: %v", utils.ToJSON(th), utils.ToJSON(rcv))
	}
	if err := onStor.RemoveThresholdProfile(th.Tenant,
		th.ID, utils.NonTransactional, false); err != nil {
		t.Error(err)
	}
	//check cache if removed
	if _, rcvErr := onStor.GetThresholdProfile(th.Tenant,
		th.ID, false, utils.NonTransactional); rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
	//check database if removed
	if _, rcvErr := onStor.GetThresholdProfile(th.Tenant,
		th.ID, true, utils.NonTransactional); rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
}

func testOnStorITThreshold(t *testing.T) {
	th := &Threshold{
		Tenant: "cgrates.org",
		ID:     "TH1",
		Snooze: time.Date(2016, 10, 1, 0, 0, 0, 0, time.UTC),
		Hits:   10,
	}
	if _, rcvErr := onStor.GetThreshold("cgrates.org", "TH1",
		false, utils.NonTransactional); rcvErr != nil && rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
	if err := onStor.SetThreshold(th); err != nil {
		t.Error(err)
	}
	//get from cache
	if rcv, err := onStor.GetThreshold("cgrates.org", "TH1",
		false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !(reflect.DeepEqual(th, rcv)) {
		t.Errorf("Expecting: %v, received: %v", th, rcv)
	}
	//get from database
	if rcv, err := onStor.GetThreshold("cgrates.org", "TH1",
		true, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !(reflect.DeepEqual(th, rcv)) {
		t.Errorf("Expecting: %v, received: %v", th, rcv)
	}
	expectedT := []string{"thd_cgrates.org:TH1"}
	if itm, err := onStor.DataDB().GetKeysForPrefix(utils.ThresholdPrefix); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedT, itm) {
		t.Errorf("Expected : %+v, but received %+v", expectedT, itm)
	}
	//update
	th.Hits = 20
	time.Sleep(sleepDelay)
	if err := onStor.SetThreshold(th); err != nil {
		t.Error(err)
	}
	//get from cache
	if rcv, err := onStor.GetThreshold("cgrates.org", "TH1",
		false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !(reflect.DeepEqual(th, rcv)) {
		t.Errorf("Expecting: %v, received: %v", th, rcv)
	}
	//get from database
	if rcv, err := onStor.GetThreshold("cgrates.org", "TH1",
		true, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !(reflect.DeepEqual(th, rcv)) {
		t.Errorf("Expecting: %v, received: %v", th, rcv)
	}
	if err := onStor.RemoveThreshold(th.Tenant, th.ID, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	//check cache if removed
	if _, rcvErr := onStor.GetThreshold(th.Tenant, th.ID,
		false, utils.NonTransactional); rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
	//check database if removed
	if _, rcvErr := onStor.GetThreshold(th.Tenant, th.ID,
		true, utils.NonTransactional); rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
}

func testOnStorITFilter(t *testing.T) {
	fp := &Filter{
		Tenant: "cgrates.org",
		ID:     "Filter1",
		RequestFilters: []*RequestFilter{
			&RequestFilter{
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
	if _, rcvErr := onStor.GetFilter("cgrates.org", "Filter1",
		false, utils.NonTransactional); rcvErr != nil && rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
	if err := onStor.SetFilter(fp); err != nil {
		t.Error(err)
	}
	//get from cache
	if rcv, err := onStor.GetFilter("cgrates.org", "Filter1",
		false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !(reflect.DeepEqual(fp, rcv)) {
		t.Errorf("Expecting: %v, received: %v", fp, rcv)
	}
	//get from database
	if rcv, err := onStor.GetFilter("cgrates.org", "Filter1",
		true, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !(reflect.DeepEqual(fp, rcv)) {
		t.Errorf("Expecting: %v, received: %v", fp, rcv)
	}
	expectedT := []string{"ftr_cgrates.org:Filter1", "ftr_cgrates.org:TestFilter2"}
	if itm, err := onStor.DataDB().GetKeysForPrefix(utils.FilterPrefix); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(len(expectedT), len(itm)) {
		t.Errorf("Expected : %+v, but received %+v", len(expectedT), len(itm))
	}
	//update
	fp.RequestFilters = []*RequestFilter{
		&RequestFilter{
			FieldName: "Account",
			Type:      "*string",
			Values:    []string{"1001", "1002"},
		},
		&RequestFilter{
			FieldName: "Destination",
			Type:      "*string",
			Values:    []string{"10", "20"},
		},
	}
	time.Sleep(sleepDelay)
	if err := onStor.SetFilter(fp); err != nil {
		t.Error(err)
	}
	//get from cache
	if rcv, err := onStor.GetFilter("cgrates.org", "Filter1",
		false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !(reflect.DeepEqual(fp, rcv)) {
		t.Errorf("Expecting: %v, received: %v", fp, rcv)
	}
	//get from database
	if rcv, err := onStor.GetFilter("cgrates.org", "Filter1",
		true, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !(reflect.DeepEqual(fp, rcv)) {
		t.Errorf("Expecting: %v, received: %v", fp, rcv)
	}
	if err := onStor.RemoveFilter(fp.Tenant, fp.ID, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	//check cache if removed
	if _, rcvErr := onStor.GetFilter("cgrates.org", "Filter1",
		false, utils.NonTransactional); rcvErr != nil && rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
	//check database if removed
	if _, rcvErr := onStor.GetFilter("cgrates.org", "Filter1",
		true, utils.NonTransactional); rcvErr != nil && rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
}

func testOnStorITSupplierProfile(t *testing.T) {
	splProfile := &SupplierProfile{
		Tenant:    "cgrates.org",
		ID:        "SPRF_1",
		FilterIDs: []string{"FLTR_ACNT_dan", "FLTR_DST_DE"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
		Sorting:       "*lowest_cost",
		SortingParams: []string{},
		Suppliers: []*Supplier{
			&Supplier{
				ID:                 "supplier1",
				FilterIDs:          []string{"FLTR_DST_DE"},
				AccountIDs:         []string{"Account1"},
				RatingPlanIDs:      []string{"RPL_1"},
				ResourceIDs:        []string{"ResGR1"},
				StatIDs:            []string{"Stat1"},
				Weight:             10,
				SupplierParameters: "param1",
			},
		},
		Weight: 20,
	}
	if _, rcvErr := onStor.GetSupplierProfile("cgrates.org", "SPRF_1",
		false, utils.NonTransactional); rcvErr != nil && rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
	if err := onStor.SetSupplierProfile(splProfile, false); err != nil {
		t.Error(err)
	}
	//get from cache
	if rcv, err := onStor.GetSupplierProfile("cgrates.org", "SPRF_1",
		false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !(reflect.DeepEqual(splProfile, rcv)) {
		t.Errorf("Expecting: %v, received: %v", splProfile, rcv)
	}
	//get from database
	if rcv, err := onStor.GetSupplierProfile("cgrates.org", "SPRF_1",
		true, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !(reflect.DeepEqual(splProfile, rcv)) {
		t.Errorf("Expecting: %v, received: %v", splProfile, rcv)
	}
	expectedT := []string{"spp_cgrates.org:SPRF_1"}
	if itm, err := onStor.DataDB().GetKeysForPrefix(utils.SupplierProfilePrefix); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedT, itm) {
		t.Errorf("Expected : %+v, but received %+v", expectedT, itm)
	}
	//update
	splProfile.Suppliers = []*Supplier{
		&Supplier{
			ID:                 "supplier1",
			FilterIDs:          []string{"FLTR_DST_DE"},
			AccountIDs:         []string{"Account1"},
			RatingPlanIDs:      []string{"RPL_1"},
			ResourceIDs:        []string{"ResGR1"},
			StatIDs:            []string{"Stat1"},
			Weight:             10,
			SupplierParameters: "param1",
		},
		&Supplier{
			ID:                 "supplier2",
			FilterIDs:          []string{"FLTR_DST_DE"},
			AccountIDs:         []string{"Account2"},
			RatingPlanIDs:      []string{"RPL_2"},
			ResourceIDs:        []string{"ResGR2"},
			StatIDs:            []string{"Stat2"},
			Weight:             20,
			SupplierParameters: "param2",
		},
	}
	time.Sleep(sleepDelay)
	if err := onStor.SetSupplierProfile(splProfile, false); err != nil {
		t.Error(err)
	}
	//get from cache
	if rcv, err := onStor.GetSupplierProfile("cgrates.org", "SPRF_1",
		false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !(reflect.DeepEqual(splProfile, rcv)) {
		t.Errorf("Expecting: %v, received: %v", splProfile, rcv)
	}
	//get from database
	if rcv, err := onStor.GetSupplierProfile("cgrates.org", "SPRF_1",
		true, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !(reflect.DeepEqual(splProfile, rcv)) {
		t.Errorf("Expecting: %v, received: %v", splProfile, rcv)
	}
	if err := onStor.RemoveSupplierProfile(splProfile.Tenant, splProfile.ID,
		utils.NonTransactional, false); err != nil {
		t.Error(err)
	}
	//check cache if removed
	if _, rcvErr := onStor.GetSupplierProfile("cgrates.org", "SPRF_1",
		false, utils.NonTransactional); rcvErr != nil && rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
	//check database if removed
	if _, rcvErr := onStor.GetSupplierProfile("cgrates.org", "SPRF_1",
		true, utils.NonTransactional); rcvErr != nil && rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
}

func testOnStorITAttributeProfile(t *testing.T) {
	mapSubstitutes := make(map[string]map[string]*Attribute)
	mapSubstitutes["FN1"] = make(map[string]*Attribute)
	mapSubstitutes["FN1"]["Init1"] = &Attribute{
		FieldName:  "FN1",
		Initial:    "Init1",
		Substitute: "Val1",
		Append:     true,
	}
	attrProfile := &AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "AttrPrf1",
		FilterIDs: []string{"FLTR_ACNT_dan", "FLTR_DST_DE"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
		Contexts:   []string{"con1"},
		Attributes: mapSubstitutes,
		Weight:     20,
	}
	if _, rcvErr := onStor.GetAttributeProfile("cgrates.org", "AttrPrf1",
		false, utils.NonTransactional); rcvErr != nil && rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
	if err := onStor.SetAttributeProfile(attrProfile, false); err != nil {
		t.Error(err)
	}
	//get from cache
	if rcv, err := onStor.GetAttributeProfile("cgrates.org", "AttrPrf1",
		false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !(reflect.DeepEqual(attrProfile, rcv)) {
		t.Errorf("Expecting: %v, received: %v", attrProfile, rcv)
	}
	//get from database
	if rcv, err := onStor.GetAttributeProfile("cgrates.org", "AttrPrf1",
		true, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !(reflect.DeepEqual(attrProfile, rcv)) {
		t.Errorf("Expecting: %v, received: %v", attrProfile, rcv)
	}
	expectedT := []string{"alp_cgrates.org:AttrPrf1"}
	if itm, err := onStor.DataDB().GetKeysForPrefix(utils.AttributeProfilePrefix); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedT, itm) {
		t.Errorf("Expected : %+v, but received %+v", expectedT, itm)
	}
	//update
	time.Sleep(sleepDelay)
	attrProfile.Contexts = []string{"con1", "con2", "con3"}
	if err := onStor.SetAttributeProfile(attrProfile, false); err != nil {
		t.Error(err)
	}
	//get from cache
	if rcv, err := onStor.GetAttributeProfile("cgrates.org", "AttrPrf1",
		false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !(reflect.DeepEqual(attrProfile, rcv)) {
		t.Errorf("Expecting: %v, received: %v", attrProfile, rcv)
	}
	//get from database
	if rcv, err := onStor.GetAttributeProfile("cgrates.org", "AttrPrf1",
		true, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !(reflect.DeepEqual(attrProfile, rcv)) {
		t.Errorf("Expecting: %v, received: %v", attrProfile, rcv)
	}
	if err := onStor.RemoveAttributeProfile(attrProfile.Tenant,
		attrProfile.ID, attrProfile.Contexts, utils.NonTransactional, false); err != nil {
		t.Error(err)
	}
	//check cache if removed
	if _, rcvErr := onStor.GetAttributeProfile("cgrates.org", "AttrPrf1",
		false, utils.NonTransactional); rcvErr != nil && rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
	//check database if removed
	if _, rcvErr := onStor.GetAttributeProfile("cgrates.org", "AttrPrf1",
		true, utils.NonTransactional); rcvErr != nil && rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
}

func testOnStorITTestThresholdFilterIndexes(t *testing.T) {
	fp := &Filter{
		Tenant: "cgrates.org",
		ID:     "Filter1",
		RequestFilters: []*RequestFilter{
			&RequestFilter{
				FieldName: "EventType",
				Type:      "*string",
				Values:    []string{"Event1", "Event2"},
			},
		},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			ExpiryTime:     time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
	}
	if err := onStor.SetFilter(fp); err != nil {
		t.Error(err)
	}
	timeMinSleep := time.Duration(0 * time.Second)
	th := &ThresholdProfile{
		Tenant:             "cgrates.org",
		ID:                 "THD_Test",
		ActivationInterval: &utils.ActivationInterval{},
		FilterIDs:          []string{"Filter1"},
		Recurrent:          true,
		MinSleep:           timeMinSleep,
		Blocker:            true,
		Weight:             1.4,
		ActionIDs:          []string{},
	}
	th2 := &ThresholdProfile{
		Tenant:             "cgrates.org",
		ID:                 "THD_Test2",
		ActivationInterval: &utils.ActivationInterval{},
		FilterIDs:          []string{"Filter1"},
		Recurrent:          true,
		MinSleep:           timeMinSleep,
		Blocker:            true,
		Weight:             1.4,
		ActionIDs:          []string{},
	}

	if err := onStor.SetThresholdProfile(th, true); err != nil {
		t.Error(err)
	}
	if err := onStor.SetThresholdProfile(th2, true); err != nil {
		t.Error(err)
	}
	eIdxes := map[string]utils.StringMap{
		"*string:EventType:Event1": utils.StringMap{
			"THD_Test":  true,
			"THD_Test2": true,
		},
		"*string:EventType:Event2": utils.StringMap{
			"THD_Test":  true,
			"THD_Test2": true,
		},
	}
	reverseIdxes := map[string]utils.StringMap{
		"THD_Test": utils.StringMap{
			"*string:EventType:Event1": true,
			"*string:EventType:Event2": true,
		},
		"THD_Test2": utils.StringMap{
			"*string:EventType:Event1": true,
			"*string:EventType:Event2": true,
		},
	}
	rfi := NewReqFilterIndexer(onStor, utils.ThresholdProfilePrefix, th.Tenant)
	if rcvIdx, err := onStor.GetFilterIndexes(
		GetDBIndexKey(rfi.itemType, rfi.dbKeySuffix, false), MetaString,
		nil); err != nil {
		t.Error(err)
	} else {
		if !reflect.DeepEqual(eIdxes, rcvIdx) {
			t.Errorf("Expecting %+v, received: %+v", eIdxes, rcvIdx)
		}
	}
	if reverseRcvIdx, err := onStor.GetFilterReverseIndexes(
		GetDBIndexKey(rfi.itemType, rfi.dbKeySuffix, true),
		nil); err != nil {
		t.Error(err)
	} else {
		if !reflect.DeepEqual(reverseIdxes, reverseRcvIdx) {
			t.Errorf("Expecting %+v, received: %+v", reverseIdxes, reverseRcvIdx)
		}
	}
	//Replace existing filter (Filter1 -> Filter2)
	fp2 := &Filter{
		Tenant: "cgrates.org",
		ID:     "Filter2",
		RequestFilters: []*RequestFilter{
			&RequestFilter{
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
	if err := onStor.SetFilter(fp2); err != nil {
		t.Error(err)
	}

	th.FilterIDs = []string{"Filter2"}
	if err := onStor.SetThresholdProfile(th, true); err != nil {
		t.Error(err)
	}
	eIdxes = map[string]utils.StringMap{
		"*string:Account:1001": utils.StringMap{
			"THD_Test": true,
		},
		"*string:Account:1002": utils.StringMap{
			"THD_Test": true,
		},
		"*string:EventType:Event1": utils.StringMap{
			"THD_Test2": true,
		},
		"*string:EventType:Event2": utils.StringMap{
			"THD_Test2": true,
		},
	}

	reverseIdxes = map[string]utils.StringMap{
		"THD_Test": utils.StringMap{
			"*string:Account:1001": true,
			"*string:Account:1002": true,
		},
		"THD_Test2": utils.StringMap{
			"*string:EventType:Event1": true,
			"*string:EventType:Event2": true,
		},
	}
	if rcvIdx, err := onStor.GetFilterIndexes(
		GetDBIndexKey(rfi.itemType, rfi.dbKeySuffix, false), MetaString,
		nil); err != nil {
		t.Error(err)
	} else {
		if !reflect.DeepEqual(eIdxes, rcvIdx) {
			t.Errorf("Expecting %+v, received: %+v", eIdxes, rcvIdx)
		}
	}
	if reverseRcvIdx, err := onStor.GetFilterReverseIndexes(
		GetDBIndexKey(rfi.itemType, rfi.dbKeySuffix, true),
		nil); err != nil {
		t.Error(err)
	} else {
		if !reflect.DeepEqual(reverseIdxes, reverseRcvIdx) {
			t.Errorf("Expecting %+v, received: %+v", reverseIdxes, reverseRcvIdx)
		}
	}
	//replace old filter with two different filters
	fp3 := &Filter{
		Tenant: "cgrates.org",
		ID:     "Filter3",
		RequestFilters: []*RequestFilter{
			&RequestFilter{
				FieldName: "Destination",
				Type:      "*string",
				Values:    []string{"10", "20"},
			},
		},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			ExpiryTime:     time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
	}
	if err := onStor.SetFilter(fp3); err != nil {
		t.Error(err)
	}

	th.FilterIDs = []string{"Filter1", "Filter3"}
	if err := onStor.SetThresholdProfile(th, true); err != nil {
		t.Error(err)
	}
	eIdxes = map[string]utils.StringMap{
		"*string:Destination:10": utils.StringMap{
			"THD_Test": true,
		},
		"*string:Destination:20": utils.StringMap{
			"THD_Test": true,
		},
		"*string:EventType:Event1": utils.StringMap{
			"THD_Test":  true,
			"THD_Test2": true,
		},
		"*string:EventType:Event2": utils.StringMap{
			"THD_Test":  true,
			"THD_Test2": true,
		},
	}
	reverseIdxes = map[string]utils.StringMap{
		"THD_Test": utils.StringMap{
			"*string:Destination:10":   true,
			"*string:Destination:20":   true,
			"*string:EventType:Event1": true,
			"*string:EventType:Event2": true,
		},
		"THD_Test2": utils.StringMap{
			"*string:EventType:Event1": true,
			"*string:EventType:Event2": true,
		},
	}
	if rcvIdx, err := onStor.GetFilterIndexes(
		GetDBIndexKey(rfi.itemType, rfi.dbKeySuffix, false), MetaString,
		nil); err != nil {
		t.Error(err)
	} else {
		if !reflect.DeepEqual(eIdxes, rcvIdx) {
			t.Errorf("Expecting %+v, received: %+v", eIdxes, rcvIdx)
		}
	}
	if reverseRcvIdx, err := onStor.GetFilterReverseIndexes(
		GetDBIndexKey(rfi.itemType, rfi.dbKeySuffix, true),
		nil); err != nil {
		t.Error(err)
	} else {
		if !reflect.DeepEqual(reverseIdxes, reverseRcvIdx) {
			t.Errorf("Expecting %+v, received: %+v", reverseIdxes, reverseRcvIdx)
		}
	}
	//remove thresholds
	if err := onStor.RemoveThresholdProfile(th.Tenant,
		th.ID, utils.NonTransactional, true); err != nil {
		t.Error(err)
	}
	if err := onStor.RemoveThresholdProfile(th2.Tenant,
		th2.ID, utils.NonTransactional, true); err != nil {
		t.Error(err)
	}
	if _, err := onStor.GetFilterIndexes(
		GetDBIndexKey(rfi.itemType, rfi.dbKeySuffix, false), MetaString,
		nil); err != utils.ErrNotFound {
		t.Error(err)
	}
	if _, err := onStor.GetFilterReverseIndexes(
		GetDBIndexKey(rfi.itemType, rfi.dbKeySuffix, true),
		nil); err != utils.ErrNotFound {
		t.Error(err)
	}

}

func testOnStorITTestAttributeProfileFilterIndexes(t *testing.T) {
	fp := &Filter{
		Tenant: "cgrates.org",
		ID:     "Filter1",
		RequestFilters: []*RequestFilter{
			&RequestFilter{
				FieldName: "EventType",
				Type:      "*string",
				Values:    []string{"Event1", "Event2"},
			},
		},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			ExpiryTime:     time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
	}
	if err := onStor.SetFilter(fp); err != nil {
		t.Error(err)
	}
	mapSubstitutes := make(map[string]map[string]*Attribute)
	mapSubstitutes["FN1"] = make(map[string]*Attribute)
	mapSubstitutes["FN1"]["Init1"] = &Attribute{
		FieldName:  "FN1",
		Initial:    "Init1",
		Substitute: "Val1",
		Append:     true,
	}
	attrProfile := &AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "AttrPrf",
		FilterIDs: []string{"Filter1"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
		Contexts:   []string{"con1", "con2"},
		Attributes: mapSubstitutes,
		Weight:     20,
	}
	//Set AttributeProfile with 2 contexts ( con1 , con2)
	if err := onStor.SetAttributeProfile(attrProfile, true); err != nil {
		t.Error(err)
	}
	eIdxes := map[string]utils.StringMap{
		"*string:EventType:Event1": utils.StringMap{
			"AttrPrf": true,
		},
		"*string:EventType:Event2": utils.StringMap{
			"AttrPrf": true,
		},
	}
	reverseIdxes := map[string]utils.StringMap{
		"AttrPrf": utils.StringMap{
			"*string:EventType:Event1": true,
			"*string:EventType:Event2": true,
		},
	}
	for _, ctx := range attrProfile.Contexts {
		rfi := NewReqFilterIndexer(onStor, utils.AttributeProfilePrefix,
			utils.ConcatenatedKey(attrProfile.Tenant, ctx))
		if rcvIdx, err := onStor.GetFilterIndexes(
			GetDBIndexKey(rfi.itemType, rfi.dbKeySuffix, false), MetaString,
			nil); err != nil {
			t.Error(err)
		} else {
			if !reflect.DeepEqual(eIdxes, rcvIdx) {
				t.Errorf("Expecting %+v, received: %+v", eIdxes, rcvIdx)
			}
		}
		if reverseRcvIdx, err := onStor.GetFilterReverseIndexes(
			GetDBIndexKey(rfi.itemType, rfi.dbKeySuffix, true),
			nil); err != nil {
			t.Error(err)
		} else {
			if !reflect.DeepEqual(reverseIdxes, reverseRcvIdx) {
				t.Errorf("Expecting %+v, received: %+v", reverseIdxes, reverseRcvIdx)
			}
		}
	}
	//Set AttributeProfile with 1 new context (con3)
	attrProfile.Contexts = []string{"con3"}
	if err := onStor.SetAttributeProfile(attrProfile, true); err != nil {
		t.Error(err)
	}
	//check indexes with the new context (con3)
	rfi := NewReqFilterIndexer(onStor, utils.AttributeProfilePrefix,
		utils.ConcatenatedKey(attrProfile.Tenant, "con3"))
	if rcvIdx, err := onStor.GetFilterIndexes(
		GetDBIndexKey(rfi.itemType, rfi.dbKeySuffix, false), MetaString,
		nil); err != nil {
		t.Error(err)
	} else {
		if !reflect.DeepEqual(eIdxes, rcvIdx) {
			t.Errorf("Expecting %+v, received: %+v", eIdxes, rcvIdx)
		}
	}
	if reverseRcvIdx, err := onStor.GetFilterReverseIndexes(
		GetDBIndexKey(rfi.itemType, rfi.dbKeySuffix, true),
		nil); err != nil {
		t.Error(err)
	} else {
		if !reflect.DeepEqual(reverseIdxes, reverseRcvIdx) {
			t.Errorf("Expecting %+v, received: %+v", reverseIdxes, reverseRcvIdx)
		}
	}

	//check if old contexts was delete
	for _, ctx := range []string{"con1", "con2"} {
		rfi := NewReqFilterIndexer(onStor, utils.AttributeProfilePrefix,
			utils.ConcatenatedKey(attrProfile.Tenant, ctx))
		if _, err := onStor.GetFilterIndexes(
			GetDBIndexKey(rfi.itemType, rfi.dbKeySuffix, false), MetaString,
			nil); err != nil && err != utils.ErrNotFound {
			t.Error(err)
		}
		if _, err := onStor.GetFilterReverseIndexes(
			GetDBIndexKey(rfi.itemType, rfi.dbKeySuffix, true),
			nil); err != nil && err != utils.ErrNotFound {
			t.Error(err)
		}
	}

	if err := onStor.RemoveAttributeProfile(attrProfile.Tenant,
		attrProfile.ID, attrProfile.Contexts, utils.NonTransactional, true); err != nil {
		t.Error(err)
	}
	//check if index is removed
	rfi = NewReqFilterIndexer(onStor, utils.AttributeProfilePrefix, utils.ConcatenatedKey("cgrates.org", "con3"))
	if _, err := onStor.GetFilterIndexes(
		GetDBIndexKey(rfi.itemType, rfi.dbKeySuffix, false), MetaString,
		nil); err != nil && err != utils.ErrNotFound {
		t.Error(err)
	}
	if _, err := onStor.GetFilterReverseIndexes(
		GetDBIndexKey(rfi.itemType, rfi.dbKeySuffix, true),
		nil); err != nil && err != utils.ErrNotFound {
		t.Error(err)
	}
}

func testOnStorITTestThresholdInlineFilterIndexing(t *testing.T) {
	fp := &Filter{
		Tenant: "cgrates.org",
		ID:     "Filter1",
		RequestFilters: []*RequestFilter{
			&RequestFilter{
				FieldName: "EventType",
				Type:      "*string",
				Values:    []string{"Event1", "Event2"},
			},
		},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			ExpiryTime:     time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
	}
	if err := onStor.SetFilter(fp); err != nil {
		t.Error(err)
	}
	th := &ThresholdProfile{
		Tenant:             "cgrates.org",
		ID:                 "THD_Test",
		ActivationInterval: &utils.ActivationInterval{},
		FilterIDs:          []string{"Filter1"},
		Recurrent:          true,
		MinSleep:           time.Duration(0 * time.Second),
		Blocker:            true,
		Weight:             1.4,
		ActionIDs:          []string{},
	}

	if err := onStor.SetThresholdProfile(th, true); err != nil {
		t.Error(err)
	}
	eIdxes := map[string]utils.StringMap{
		"*string:EventType:Event1": utils.StringMap{
			"THD_Test": true,
		},
		"*string:EventType:Event2": utils.StringMap{
			"THD_Test": true,
		},
	}
	reverseIdxes := map[string]utils.StringMap{
		"THD_Test": utils.StringMap{
			"*string:EventType:Event1": true,
			"*string:EventType:Event2": true,
		},
	}
	rfi := NewReqFilterIndexer(onStor, utils.ThresholdProfilePrefix, th.Tenant)
	if rcvIdx, err := onStor.GetFilterIndexes(
		GetDBIndexKey(rfi.itemType, rfi.dbKeySuffix, false), MetaString,
		nil); err != nil {
		t.Error(err)
	} else {
		if !reflect.DeepEqual(eIdxes, rcvIdx) {
			t.Errorf("Expecting %+v, received: %+v", eIdxes, rcvIdx)
		}
	}
	if reverseRcvIdx, err := onStor.GetFilterReverseIndexes(
		GetDBIndexKey(rfi.itemType, rfi.dbKeySuffix, true),
		nil); err != nil {
		t.Error(err)
	} else {
		if !reflect.DeepEqual(reverseIdxes, reverseRcvIdx) {
			t.Errorf("Expecting %+v, received: %+v", reverseIdxes, reverseRcvIdx)
		}
	}
	//Add an InlineFilter
	th.FilterIDs = []string{"Filter1", "*string:Account:1001"}
	if err := onStor.SetThresholdProfile(th, true); err != nil {
		t.Error(err)
	}
	eIdxes = map[string]utils.StringMap{
		"*string:Account:1001": utils.StringMap{
			"THD_Test": true,
		},
		"*string:EventType:Event1": utils.StringMap{
			"THD_Test": true,
		},
		"*string:EventType:Event2": utils.StringMap{
			"THD_Test": true,
		},
	}

	reverseIdxes = map[string]utils.StringMap{
		"THD_Test": utils.StringMap{
			"*string:Account:1001":     true,
			"*string:EventType:Event1": true,
			"*string:EventType:Event2": true,
		},
	}
	if rcvIdx, err := onStor.GetFilterIndexes(
		GetDBIndexKey(rfi.itemType, rfi.dbKeySuffix, false), MetaString,
		nil); err != nil {
		t.Error(err)
	} else {
		if !reflect.DeepEqual(eIdxes, rcvIdx) {
			t.Errorf("Expecting %+v, received: %+v", eIdxes, rcvIdx)
		}
	}
	if reverseRcvIdx, err := onStor.GetFilterReverseIndexes(
		GetDBIndexKey(rfi.itemType, rfi.dbKeySuffix, true),
		nil); err != nil {
		t.Error(err)
	} else {
		if !reflect.DeepEqual(reverseIdxes, reverseRcvIdx) {
			t.Errorf("Expecting %+v, received: %+v", reverseIdxes, reverseRcvIdx)
		}
	}
	//remove threshold
	if err := onStor.RemoveThresholdProfile(th.Tenant,
		th.ID, utils.NonTransactional, true); err != nil {
		t.Error(err)
	}
	if _, err := onStor.GetFilterIndexes(
		GetDBIndexKey(rfi.itemType, rfi.dbKeySuffix, false), MetaString,
		nil); err != utils.ErrNotFound {
		t.Error(err)
	}
	if _, err := onStor.GetFilterReverseIndexes(
		GetDBIndexKey(rfi.itemType, rfi.dbKeySuffix, true),
		nil); err != utils.ErrNotFound {
		t.Error(err)
	}
}
