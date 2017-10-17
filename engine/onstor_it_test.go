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
	rdsITdb   *RedisStorage
	mgoITdb   *MongoStorage
	onStor    *DataManager
	onStorCfg string
)

// subtests to be executed for each confDIR
var sTestsOnStorIT = []func(t *testing.T){
	testOnStorITFlush,
	testOnStorITIsDBEmpty,
	testOnStorITSetGetDerivedCharges,
	testOnStorITSetReqFilterIndexes,
	testOnStorITGetReqFilterIndexes,
	testOnStorITMatchReqFilterIndex,
	testOnStorITCacheDestinations,
	testOnStorITCacheReverseDestinations,
	testOnStorITCacheRatingPlan,
	testOnStorITCacheRatingProfile,
	testOnStorITCacheActions,
	testOnStorITCacheActionPlan,
	testOnStorITCacheAccountActionPlans,
	testOnStorITCacheActionTriggers,
	testOnStorITCacheSharedGroup,
	testOnStorITCacheDerivedChargers,
	testOnStorITCacheLCR,
	testOnStorITCacheAlias,
	testOnStorITCacheReverseAlias,
	testOnStorITCacheResource,
	testOnStorITCacheResourceProfile,
	testOnStorITCacheTiming,
	// ToDo: test cache flush for a prefix
	// ToDo: testOnStorITLoadAccountingCache
	testOnStorITHasData,
	testOnStorITPushPop,
	testOnStorITCRUDRatingPlan,
	testOnStorITCRUDRatingProfile,
	testOnStorITCRUDDestinations,
	testOnStorITCRUDReverseDestinations,
	testOnStorITCRUDLCR,
	testOnStorITCRUDCdrStats,
	testOnStorITCRUDActions,
	testOnStorITCRUDSharedGroup,
	testOnStorITCRUDActionTriggers,
	testOnStorITCRUDActionPlan,
	testOnStorITCRUDAccountActionPlans,
	testOnStorITCRUDAccount,
	testOnStorITCRUDCdrStatsQueue,
	testOnStorITCRUDSubscribers,
	testOnStorITCRUDUser,
	testOnStorITCRUDAlias,
	testOnStorITCRUDReverseAlias,
	testOnStorITCRUDResource,
	testOnStorITCRUDResourceProfile,
	testOnStorITCRUDTiming,
	testOnStorITCRUDHistory,
	testOnStorITCRUDStructVersion,
	testOnStorITCRUDStatQueueProfile,
	testOnStorITCRUDStoredStatQueue,
	testOnStorITCRUDThresholdProfile,
	testOnStorITCRUDThreshold,
	testOnStorITCRUDFilter,
}

func TestOnStorITRedisConnect(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	rdsITdb, err = NewRedisStorage(fmt.Sprintf("%s:%s", cfg.DataDbHost, cfg.DataDbPort), 4, cfg.DataDbPass, cfg.DBDataEncoding, utils.REDIS_MAX_CONNS, nil, 1)
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
	cdrsMongoCfgPath := path.Join(*dataDir, "conf", "samples", "cdrsv2mongo")
	mgoITCfg, err := config.NewCGRConfigFromFolder(cdrsMongoCfgPath)
	if err != nil {
		t.Fatal(err)
	}
	if mgoITdb, err = NewMongoStorage(mgoITCfg.StorDBHost, mgoITCfg.StorDBPort, mgoITCfg.StorDBName, mgoITCfg.StorDBUser, mgoITCfg.StorDBPass,
		utils.StorDB, nil, mgoITCfg.CacheConfig, mgoITCfg.LoadHistorySize); err != nil {
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
	if _, err := onStor.DataDB().GetDerivedChargers(keyCharger1, true, utils.NonTransactional); err == nil {
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
	if rcvCharger, err := onStor.DataDB().GetDerivedChargers(keyCharger1, true, utils.NonTransactional); err != nil {
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
	if rcvCharger, err := onStor.DataDB().GetDerivedChargers(keyCharger1, false, utils.NonTransactional); err != nil {
		t.Error("Error when retrieving DerivedCHarger", err.Error())
	} else if !reflect.DeepEqual(rcvCharger, charger1) {
		t.Errorf("Expecting %v, received: %v", charger1, rcvCharger)
	}
}

func testOnStorITSetReqFilterIndexes(t *testing.T) {
	idxes := map[string]map[string]utils.StringMap{
		"Account": map[string]utils.StringMap{
			"1001": utils.StringMap{
				"RL1": true,
			},
			"1002": utils.StringMap{
				"RL1": true,
				"RL2": true,
			},
			"dan": utils.StringMap{
				"RL2": true,
			},
		},
		"Subject": map[string]utils.StringMap{
			"dan": utils.StringMap{
				"RL2": true,
			},
		},
		utils.NOT_AVAILABLE: map[string]utils.StringMap{
			utils.NOT_AVAILABLE: utils.StringMap{
				"RL4": true,
				"RL5": true,
			},
		},
	}
	if err := onStor.DataDB().SetReqFilterIndexes(utils.ResourceProfilesStringIndex, idxes); err != nil {
		t.Error(err)
	}
}

func testOnStorITGetReqFilterIndexes(t *testing.T) {
	eIdxes := map[string]map[string]utils.StringMap{
		"Account": map[string]utils.StringMap{
			"1001": utils.StringMap{
				"RL1": true,
			},
			"1002": utils.StringMap{
				"RL1": true,
				"RL2": true,
			},
			"dan": utils.StringMap{
				"RL2": true,
			},
		},
		"Subject": map[string]utils.StringMap{
			"dan": utils.StringMap{
				"RL2": true,
			},
		},
		utils.NOT_AVAILABLE: map[string]utils.StringMap{
			utils.NOT_AVAILABLE: utils.StringMap{
				"RL4": true,
				"RL5": true,
			},
		},
	}
	if idxes, err := onStor.DataDB().GetReqFilterIndexes(utils.ResourceProfilesStringIndex); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eIdxes, idxes) {
		t.Errorf("Expecting: %+v, received: %+v", eIdxes, idxes)
	}
	if _, err := onStor.DataDB().GetReqFilterIndexes("unknown_key"); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	}
}

func testOnStorITMatchReqFilterIndex(t *testing.T) {
	eMp := utils.StringMap{
		"RL1": true,
		"RL2": true,
	}
	if rcvMp, err := onStor.DataDB().MatchReqFilterIndex(utils.ResourceProfilesStringIndex, "Account", "1002"); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eMp, rcvMp) {
		t.Errorf("Expecting: %+v, received: %+v", eMp, rcvMp)
	}
	if _, err := onStor.DataDB().MatchReqFilterIndex(utils.ResourceProfilesStringIndex, "NonexistentField", "1002"); err == nil || err != utils.ErrNotFound {
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

func testOnStorITCacheRatingPlan(t *testing.T) {
	rp := &RatingPlan{
		Id: "TEST_RP_CACHE",
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
	if err := onStor.DataDB().SetRatingPlan(rp, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	expectedCRPl := []string{"rpl_TEST_RP_CACHE"}
	if itm, err := onStor.DataDB().GetKeysForPrefix(utils.RATING_PLAN_PREFIX); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedCRPl, itm) {
		t.Errorf("Expected : %+v, but received %+v", expectedCRPl, itm)
	}
	if _, hasIt := cache.Get(utils.RATING_PLAN_PREFIX + rp.Id); hasIt {
		t.Error("Already in cache")
	}
	if err := onStor.CacheDataFromDB(utils.RATING_PLAN_PREFIX, []string{rp.Id}, false); err != nil {
		t.Error(err)
	}
	if itm, hasIt := cache.Get(utils.RATING_PLAN_PREFIX + rp.Id); !hasIt {
		t.Error("Did not cache")
	} else if rcvRp := itm.(*RatingPlan); !reflect.DeepEqual(rp, rcvRp) {
		t.Error("Wrong item in the cache")
	}
}

func testOnStorITCacheRatingProfile(t *testing.T) {
	rpf := &RatingProfile{
		Id: "*out:test:0:trp",
		RatingPlanActivations: RatingPlanActivations{
			&RatingPlanActivation{
				ActivationTime:  time.Date(2013, 10, 1, 0, 0, 0, 0, time.UTC).Local(),
				RatingPlanId:    "TDRT",
				FallbackKeys:    []string{"*out:test:0:danb", "*out:test:0:rif"},
				CdrStatQueueIds: []string{},
			}},
	}
	if err := onStor.DataDB().SetRatingProfile(rpf, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	expectedCRR := []string{"rpf_*out:test:0:trp"}
	if itm, err := onStor.DataDB().GetKeysForPrefix(utils.RATING_PROFILE_PREFIX); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedCRR, itm) {
		t.Errorf("Expected : %+v, but received %+v", expectedCRR, itm)
	}
	if _, hasIt := cache.Get(utils.RATING_PROFILE_PREFIX + rpf.Id); hasIt {
		t.Error("Already in cache")
	}
	if err := onStor.CacheDataFromDB(utils.RATING_PROFILE_PREFIX, []string{rpf.Id}, false); err != nil {
		t.Error(err)
	}
	if itm, hasIt := cache.Get(utils.RATING_PROFILE_PREFIX + rpf.Id); !hasIt {
		t.Error("Did not cache")
	} else if rcvRp := itm.(*RatingProfile); !reflect.DeepEqual(rpf, rcvRp) {
		t.Errorf("Expecting: %+v, received: %+v", rpf, rcvRp)
	}
}

func testOnStorITCacheActions(t *testing.T) {
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
				Timings:  make([]*RITiming, 0),
				Blocker:  utils.BoolPointer(false),
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
				Timings:        make([]*RITiming, 0),
				Blocker:        utils.BoolPointer(false),
			},
		},
	}
	if err := onStor.DataDB().SetActions(acts[0].Id, acts, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	expectedCA := []string{"act_MINI"}
	if itm, err := onStor.DataDB().GetKeysForPrefix(utils.ACTION_PREFIX); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedCA, itm) {
		t.Errorf("Expected : %+v, but received %+v", expectedCA, itm)
	}
	if _, hasIt := cache.Get(utils.ACTION_PREFIX + acts[0].Id); hasIt {
		t.Error("Already in cache")
	}
	if err := onStor.CacheDataFromDB(utils.ACTION_PREFIX, []string{acts[0].Id}, false); err != nil {
		t.Error(err)
	}
	if itm, hasIt := cache.Get(utils.ACTION_PREFIX + acts[0].Id); !hasIt {
		t.Error("Did not cache")
	} else if rcv := itm.(Actions); !reflect.DeepEqual(acts, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", acts, rcv)
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
			LastExecutionTime: time.Date(2012, 1, 1, 0, 0, 0, 0, time.UTC).Local(),
			ExpirationDate:    time.Date(2012, 1, 1, 0, 0, 0, 0, time.UTC).Local(),
			ActivationDate:    time.Date(2012, 1, 1, 0, 0, 0, 0, time.UTC).Local()},
	}
	atsID := ats[0].ID
	if err := onStor.DataDB().SetActionTriggers(atsID, ats, utils.NonTransactional); err != nil {
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

func testOnStorITCacheSharedGroup(t *testing.T) {
	sg := &SharedGroup{
		Id: "SG1",
		AccountParameters: map[string]*SharingParameters{
			"*any": &SharingParameters{
				Strategy:      "*lowest",
				RatingSubject: "",
			},
		},
		MemberIds: make(utils.StringMap),
	}
	if err := onStor.DataDB().SetSharedGroup(sg, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	expectedCSh := []string{"shg_SG1"}
	if itm, err := onStor.DataDB().GetKeysForPrefix(utils.SHARED_GROUP_PREFIX); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedCSh, itm) {
		t.Errorf("Expected : %+v, but received %+v", expectedCSh, itm)
	}
	if _, hasIt := cache.Get(utils.SHARED_GROUP_PREFIX + sg.Id); hasIt {
		t.Error("Already in cache")
	}
	if err := onStor.CacheDataFromDB(utils.SHARED_GROUP_PREFIX, []string{sg.Id}, false); err != nil {
		t.Error(err)
	}
	if itm, hasIt := cache.Get(utils.SHARED_GROUP_PREFIX + sg.Id); !hasIt {
		t.Error("Did not cache")
	} else if rcv := itm.(*SharedGroup); !reflect.DeepEqual(sg, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", sg, rcv)
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

func testOnStorITCacheLCR(t *testing.T) {
	lcr := &LCR{
		Tenant:    "cgrates.org",
		Category:  "call",
		Direction: "*out",
		Account:   "*any",
		Subject:   "*any",
		Activations: []*LCRActivation{
			&LCRActivation{
				ActivationTime: time.Date(2012, 1, 1, 0, 0, 0, 0, time.UTC).Local(),
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
	if err := onStor.DataDB().SetLCR(lcr, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	expectedCLCR := []string{"lcr_*out:cgrates.org:call:*any:*any"}
	if itm, err := onStor.DataDB().GetKeysForPrefix(utils.LCR_PREFIX); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedCLCR, itm) {
		t.Errorf("Expected : %+v, but received %+v", expectedCLCR, itm)
	}
	if _, hasIt := cache.Get(utils.LCR_PREFIX + lcr.GetId()); hasIt {
		t.Error("Already in cache")
	}
	if err := onStor.CacheDataFromDB(utils.LCR_PREFIX, []string{lcr.GetId()}, false); err != nil {
		t.Error(err)
	}
	if itm, hasIt := cache.Get(utils.LCR_PREFIX + lcr.GetId()); !hasIt {
		t.Error("Did not cache")
	} else if rcv := itm.(*LCR); !reflect.DeepEqual(lcr, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", lcr, rcv)
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

func testOnStorITCacheResourceProfile(t *testing.T) {
	rCfg := &ResourceProfile{
		Tenant: "cgrates.org",
		ID:     "RL_TEST",
		Weight: 10,
		Filters: []*RequestFilter{
			&RequestFilter{Type: MetaString, FieldName: "Account", Values: []string{"dan", "1002"}},
			&RequestFilter{Type: MetaRSRFields, Values: []string{"Subject(~^1.*1$)", "Destination(1002)"},
				rsrFields: utils.ParseRSRFieldsMustCompile("Subject(~^1.*1$);Destination(1002)", utils.INFIELD_SEP),
			}},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 3, 13, 43, 0, 0, time.UTC).Local(),
			ExpiryTime:     time.Date(2015, 7, 3, 13, 43, 0, 0, time.UTC).Local()},
		Limit:      1,
		Thresholds: []string{"TEST_ACTIONS"},
		UsageTTL:   time.Duration(1 * time.Millisecond),
	}
	if err := onStor.DataDB().SetResourceProfile(rCfg); err != nil {
		t.Error(err)
	}
	expectedR := []string{"rsp_cgrates.org:RL_TEST"}
	if itm, err := onStor.DataDB().GetKeysForPrefix(utils.ResourceProfilesPrefix); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedR, itm) {
		t.Errorf("Expected : %+v, but received %+v", expectedR, itm)
	}
	if _, hasIt := cache.Get(utils.ResourceProfilesPrefix + rCfg.TenantID()); hasIt {
		t.Error("Already in cache")
	}
	if err := onStor.CacheDataFromDB(utils.ResourceProfilesPrefix, []string{rCfg.TenantID()}, false); err != nil {
		t.Error(err)
	}
	if itm, hasIt := cache.Get(utils.ResourceProfilesPrefix + rCfg.TenantID()); !hasIt {
		t.Error("Did not cache")
	} else if rcv := itm.(*ResourceProfile); !reflect.DeepEqual(rCfg, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", rCfg, rcv)
	}
}

func testOnStorITCacheTiming(t *testing.T) {
	tmg := &utils.TPTiming{
		ID:        "TEST_TMG",
		Years:     utils.Years{2016, 2017},
		Months:    utils.Months{time.January, time.February, time.March},
		MonthDays: utils.MonthDays{1, 2, 3, 4},
		WeekDays:  utils.WeekDays{},
		StartTime: "00:00:00",
		EndTime:   "",
	}

	if err := onStor.SetTiming(tmg); err != nil {
		t.Error(err)
	}
	expectedT := []string{"tmg_TEST_TMG"}
	if itm, err := onStor.DataDB().GetKeysForPrefix(utils.TimingsPrefix); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedT, itm) {
		t.Errorf("Expected : %+v, but received %+v", expectedT, itm)
	}
	if _, hasIt := cache.Get(utils.TimingsPrefix + tmg.ID); hasIt {
		t.Error("Already in cache")
	}
	if err := onStor.CacheDataFromDB(utils.TimingsPrefix, []string{tmg.ID}, false); err != nil {
		t.Error(err)
	}
	if itm, hasIt := cache.Get(utils.TimingsPrefix + tmg.ID); !hasIt {
		t.Error("Did not cache")
	} else if rcv := itm.(*utils.TPTiming); !reflect.DeepEqual(tmg, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", tmg, rcv)
	}
}

func testOnStorITCacheResource(t *testing.T) {
	res := &Resource{
		Tenant: "cgrates.org",
		ID:     "RL1",
		Usages: map[string]*ResourceUsage{
			"RU1": &ResourceUsage{
				ID:         "RU1",
				ExpiryTime: time.Date(2014, 7, 3, 13, 43, 0, 0, time.UTC).Local(),
				Units:      2,
			},
		},
		TTLIdx: []string{"RU1"},
	}
	if err := onStor.DataDB().SetResource(res); err != nil {
		t.Error(err)
	}
	expectedT := []string{"res_cgrates.org:RL1"}
	if itm, err := onStor.DataDB().GetKeysForPrefix(utils.ResourcesPrefix); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedT, itm) {
		t.Errorf("Expected : %+v, but received %+v", expectedT, itm)
	}

	if _, hasIt := cache.Get(utils.ResourcesPrefix + res.TenantID()); hasIt {
		t.Error("Already in cache")
	}
	if err := onStor.CacheDataFromDB(utils.ResourcesPrefix, []string{res.TenantID()}, false); err != nil {
		t.Error(err)
	}
	if itm, hasIt := cache.Get(utils.ResourcesPrefix + res.TenantID()); !hasIt {
		t.Error("Did not cache")
	} else if rcv := itm.(*Resource); !reflect.DeepEqual(res, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", res, rcv)
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
	if err := onStor.DataDB().SetRatingPlan(rp, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	expectedRP := []string{"rpl_HasData", "rpl_TEST_RP_CACHE"}
	if itm, err := onStor.DataDB().GetKeysForPrefix(utils.RATING_PLAN_PREFIX); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(len(expectedRP), len(itm)) {
		t.Errorf("Expected : %+v, but received %+v", len(expectedRP), len(itm))
	}
	if rcv, err := onStor.DataDB().HasData(utils.RATING_PLAN_PREFIX, rp.Id); err != nil {
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

func testOnStorITCRUDRatingPlan(t *testing.T) {
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
	if _, rcvErr := onStor.DataDB().GetRatingPlan(rp.Id, true, utils.NonTransactional); rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
	if err := onStor.DataDB().SetRatingPlan(rp, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	expectedRP := []string{"rpl_TEST_RP_CACHE", "rpl_HasData", "rpl_CRUDRatingPlan"}

	if itm, err := onStor.DataDB().GetKeysForPrefix(utils.RATING_PLAN_PREFIX); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(len(expectedRP), len(itm)) {
		t.Errorf("Expected : %+v, but received %+v", len(expectedRP), len(itm))
	}

	if rcv, err := onStor.DataDB().GetRatingPlan(rp.Id, true, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rp, rcv) {
		t.Errorf("Expecting: %v, received: %v", rp, rcv)
	}
	// FixMe
	// if err = onStor.DataDB().SelectDatabase("13"); err != nil {
	// 	t.Error(err)
	// }
	// if _, rcvErr := onStor.DataDB().GetRatingPlan(rp.Id, false, utils.NonTransactional); rcvErr != utils.ErrNotFound {
	// 	t.Error(rcvErr)
	// }

	if rcv, err := onStor.DataDB().GetRatingPlan(rp.Id, false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rp, rcv) {
		t.Errorf("Expecting: %v, received: %v", rp, rcv)
	}
	// if err = onStor.DataDB().SelectDatabase(onStorCfg); err != nil {
	// 	t.Error(err)
	// }
}

func testOnStorITCRUDRatingProfile(t *testing.T) {
	rpf := &RatingProfile{
		Id: "*out:test:1:trp",
		RatingPlanActivations: RatingPlanActivations{
			&RatingPlanActivation{
				ActivationTime:  time.Date(2013, 10, 1, 0, 0, 0, 0, time.UTC).Local(),
				RatingPlanId:    "TDRT",
				FallbackKeys:    []string{"*out:test:1:danb", "*out:test:1:rif"},
				CdrStatQueueIds: []string{},
			}},
	}
	if _, rcvErr := onStor.DataDB().GetRatingProfile(rpf.Id, true, utils.NonTransactional); rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
	if err := onStor.DataDB().SetRatingProfile(rpf, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	if rcv, err := onStor.DataDB().GetRatingProfile(rpf.Id, true, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rpf, rcv) {
		t.Errorf("Expecting: %v, received: %v", rpf, rcv)
	}
	// FixMe
	// if err = onStor.DataDB().SelectDatabase("13"); err != nil {
	// 	t.Error(err)
	// }
	// if _, rcvErr := onStor.DataDB().GetRatingProfile(rpf.Id, false, utils.NonTransactional); rcvErr != utils.ErrNotFound {
	// 	t.Error(rcvErr)
	// }
	//
	if rcv, err := onStor.DataDB().GetRatingProfile(rpf.Id, false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rpf, rcv) {
		t.Errorf("Expecting: %v, received: %v", rpf, rcv)
	}
	// if err = onStor.DataDB().SelectDatabase(onStorCfg); err != nil {
	// 	t.Error(err)
	// }
	if err = onStor.DataDB().RemoveRatingProfile(rpf.Id, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	if _, rcvErr := onStor.DataDB().GetRatingProfile(rpf.Id, true, utils.NonTransactional); rcvErr != utils.ErrNotFound {
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

func testOnStorITCRUDLCR(t *testing.T) {
	lcr := &LCR{
		Tenant:    "cgrates.org",
		Category:  "call",
		Direction: "*out",
		Account:   "testOnStorITCRUDLCR",
		Subject:   "testOnStorITCRUDLCR",
		Activations: []*LCRActivation{
			&LCRActivation{
				ActivationTime: time.Date(2012, 1, 1, 0, 0, 0, 0, time.UTC).Local(),
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

	if _, rcvErr := onStor.DataDB().GetLCR(lcr.GetId(), true, utils.NonTransactional); rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
	if err := onStor.DataDB().SetLCR(lcr, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	if rcv, err := onStor.DataDB().GetLCR(lcr.GetId(), true, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(lcr, rcv) {
		t.Errorf("Expecting: %v, received: %v", lcr, rcv)
	}
	// FixMe
	// if err = onStor.DataDB().SelectDatabase("13"); err != nil {
	// 	t.Error(err)
	// }
	// if _, rcvErr := onStor.DataDB().GetLCR(lcr.GetId(), false, utils.NonTransactional); rcvErr != utils.ErrNotFound {
	// 	t.Error(rcvErr)
	// }
	//
	if rcv, err := onStor.DataDB().GetLCR(lcr.GetId(), false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(lcr, rcv) {
		t.Errorf("Expecting: %v, received: %v", lcr, rcv)
	}
	// if err = onStor.DataDB().SelectDatabase(onStorCfg); err != nil {
	// 	t.Error(err)
	// }
}

func testOnStorITCRUDCdrStats(t *testing.T) {
	cdrs := &CdrStats{Metrics: []string{ASR, PDD, ACD, TCD, ACC, TCC, DDC}}

	if _, rcvErr := onStor.DataDB().GetCdrStats(""); rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
	if err := onStor.DataDB().SetCdrStats(cdrs); err != nil {
		t.Error(err)
	}
	if rcv, err := onStor.DataDB().GetCdrStats(""); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(cdrs.Metrics, rcv.Metrics) {
		t.Errorf("Expecting: %v, received: %v", cdrs.Metrics, rcv.Metrics)
	}
	if rcv, err := onStor.DataDB().GetAllCdrStats(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual([]*CdrStats{cdrs}[0].Metrics, rcv[0].Metrics) {
		t.Errorf("Expecting: %v, received: %v", []*CdrStats{cdrs}[0].Metrics, rcv[0].Metrics)
	}
}

func testOnStorITCRUDActions(t *testing.T) {
	acts := Actions{
		&Action{
			Id:               "CRUDActions",
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
				Timings:  make([]*RITiming, 0),
				Blocker:  utils.BoolPointer(false),
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
				Timings:        make([]*RITiming, 0),
				Blocker:        utils.BoolPointer(false),
			},
		},
	}
	if _, rcvErr := onStor.DataDB().GetActions(acts[0].Id, true, utils.NonTransactional); rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
	if err := onStor.DataDB().SetActions(acts[0].Id, acts, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	if rcv, err := onStor.DataDB().GetActions(acts[0].Id, true, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(acts[0], rcv[0]) {
		t.Errorf("Expecting: %v, received: %v", acts[0], rcv[0])
	}
	// FixMe
	// if err = onStor.DataDB().SelectDatabase("13"); err != nil {
	// 	t.Error(err)
	// }
	// if _, rcvErr := onStor.DataDB().GetActions(acts[0].Id, false, utils.NonTransactional); rcvErr != utils.ErrNotFound {
	// 	t.Error(rcvErr)
	// }
	//
	// if rcv, err := onStor.DataDB().GetActions(acts[0].Id, false, utils.NonTransactional); err != nil {
	// 	t.Error(err)
	// } else if !reflect.DeepEqual(acts[0], rcv[0]) {
	// 	t.Errorf("Expecting: %v, received: %v", acts[0], rcv[0])
	// }
	// if err = onStor.DataDB().SelectDatabase(onStorCfg); err != nil {
	// 	t.Error(err)
	// }

	if err := onStor.DataDB().RemoveActions(acts[0].Id, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	if _, rcvErr := onStor.DataDB().GetActions(acts[0].Id, true, utils.NonTransactional); rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}

}

func testOnStorITCRUDSharedGroup(t *testing.T) {
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
	if _, rcvErr := onStor.DataDB().GetSharedGroup(sg.Id, true, utils.NonTransactional); rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
	if err := onStor.DataDB().SetSharedGroup(sg, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	if rcv, err := onStor.DataDB().GetSharedGroup(sg.Id, true, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(sg, rcv) {
		t.Errorf("Expecting: %v, received: %v", sg, rcv)
	}
	// FixMe
	// if err = onStor.DataDB().SelectDatabase("13"); err != nil {
	// 	t.Error(err)
	// }
	// if _, rcvErr := onStor.DataDB().GetSharedGroup(sg.Id, false, utils.NonTransactional); rcvErr != utils.ErrNotFound {
	// 	t.Error(rcvErr)
	// }

	if rcv, err := onStor.DataDB().GetSharedGroup(sg.Id, false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(sg, rcv) {
		t.Errorf("Expecting: %v, received: %v", sg, rcv)
	}
	// if err = onStor.DataDB().SelectDatabase(onStorCfg); err != nil {
	// 	t.Error(err)
	// }
}

func testOnStorITCRUDActionTriggers(t *testing.T) {
	ats := ActionTriggers{
		&ActionTrigger{
			ID:                "testOnStorITCRUDActionTriggers",
			Balance:           &BalanceFilter{Type: utils.StringPointer(utils.MONETARY), Directions: utils.StringMapPointer(utils.NewStringMap(utils.OUT)), Timings: make([]*RITiming, 0)},
			ThresholdValue:    2,
			ThresholdType:     utils.TRIGGER_MAX_EVENT_COUNTER,
			ActionsID:         "TEST_ACTIONS",
			LastExecutionTime: time.Date(2012, 1, 1, 0, 0, 0, 0, time.UTC).Local(),
			ExpirationDate:    time.Date(2012, 1, 1, 0, 0, 0, 0, time.UTC).Local(),
			ActivationDate:    time.Date(2012, 1, 1, 0, 0, 0, 0, time.UTC).Local()},
	}
	atsID := ats[0].ID
	if _, rcvErr := onStor.DataDB().GetActionTriggers(atsID, true, utils.NonTransactional); rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
	if err := onStor.DataDB().SetActionTriggers(atsID, ats, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	if rcv, err := onStor.DataDB().GetActionTriggers(atsID, true, utils.NonTransactional); err != nil {
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
	if rcv, err := onStor.DataDB().GetActionTriggers(atsID, false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(ats[0], rcv[0]) {
		t.Errorf("Expecting: %v, received: %v", ats[0], rcv[0])
	}
	// if err = onStor.DataDB().SelectDatabase(onStorCfg); err != nil {
	// 	t.Error(err)
	// }

	if err := onStor.DataDB().RemoveActionTriggers(atsID, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	if _, rcvErr := onStor.DataDB().GetActionTriggers(atsID, true, utils.NonTransactional); rcvErr != utils.ErrNotFound {
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
				SetupTime:  time.Date(2012, 1, 1, 0, 0, 0, 0, time.UTC).Local(),
				AnswerTime: time.Date(2012, 1, 1, 0, 0, 10, 0, time.UTC).Local(),
				EventTime:  time.Date(2012, 1, 1, 0, 0, 0, 0, time.UTC).Local(),
			}},
	}
	if _, rcvErr := onStor.DataDB().GetCdrStatsQueue(sq.GetId()); rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
	if err := onStor.DataDB().SetCdrStatsQueue(sq); err != nil {
		t.Error(err)
	}
	if rcv, err := onStor.DataDB().GetCdrStatsQueue(sq.GetId()); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(sq.Cdrs, rcv.Cdrs) {
		t.Errorf("Expecting: %v, received: %v", sq.Cdrs, rcv.Cdrs)
	}
}

func testOnStorITCRUDSubscribers(t *testing.T) {
	if sbs, err := onStor.DataDB().GetSubscribers(); err != nil {
		t.Error(err)
	} else if len(sbs) != 0 {
		t.Errorf("Received subscribers: %+v", sbs)
	}
	sbsc := &SubscriberData{
		ExpTime: time.Date(2012, 1, 1, 0, 0, 0, 0, time.UTC).Local(),
		Filters: utils.ParseRSRFieldsMustCompile("^*default", utils.INFIELD_SEP)}
	sbscID := "testOnStorITCRUDSubscribers"
	if err := onStor.DataDB().SetSubscriber(sbscID, sbsc); err != nil {
		t.Error(err)
	}
	if rcv, err := onStor.DataDB().GetSubscribers(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(sbsc.ExpTime, rcv[sbscID].ExpTime) { // Test just ExpTime since RSRField is more complex behind
		t.Errorf("Expecting: %v, received: %v", sbsc, rcv[sbscID])
	}
	if err := onStor.DataDB().RemoveSubscriber(sbscID); err != nil {
		t.Error(err)
	}
	if sbs, err := onStor.DataDB().GetSubscribers(); err != nil {
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
	if _, rcvErr := onStor.DataDB().GetUser(usr.GetId()); rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
	if err := onStor.DataDB().SetUser(usr); err != nil {
		t.Error(err)
	}
	if rcv, err := onStor.DataDB().GetUser(usr.GetId()); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(usr, rcv) {
		t.Errorf("Expecting: %v, received: %v", usr, rcv)
	}
	if rcv, err := onStor.DataDB().GetUsers(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(usr, rcv[0]) {
		t.Errorf("Expecting: %v, received: %v", usr, rcv[0])
	}
	if err := onStor.DataDB().RemoveUser(usr.GetId()); err != nil {
		t.Error(err)
	}
	if _, rcvErr := onStor.DataDB().GetUser(usr.GetId()); rcvErr != utils.ErrNotFound {
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

func testOnStorITCRUDResourceProfile(t *testing.T) {
	rL := &ResourceProfile{
		ID:     "RL_TEST2",
		Weight: 10,
		Filters: []*RequestFilter{
			&RequestFilter{Type: MetaString, FieldName: "Account", Values: []string{"dan", "1002"}},
			&RequestFilter{Type: MetaRSRFields, Values: []string{"Subject(~^1.*1$)", "Destination(1002)"},
				rsrFields: utils.ParseRSRFieldsMustCompile("Subject(~^1.*1$);Destination(1002)", utils.INFIELD_SEP),
			}},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 3, 13, 43, 0, 0, time.UTC).Local(),
			ExpiryTime:     time.Date(2015, 7, 3, 13, 43, 0, 0, time.UTC).Local()},
		Limit:      1,
		Thresholds: []string{"TEST_ACTIONS"},
		UsageTTL:   time.Duration(1 * time.Millisecond),
	}
	if _, rcvErr := onStor.DataDB().GetResourceProfile(rL.Tenant, rL.ID, true, utils.NonTransactional); rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
	if err := onStor.DataDB().SetResourceProfile(rL); err != nil {
		t.Error(err)
	}
	if rcv, err := onStor.DataDB().GetResourceProfile(rL.Tenant, rL.ID, true, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rL, rcv) {
		t.Errorf("Expecting: %v, received: %v", rL, rcv)
	}
	// FixMe
	// if err = onStor.DataDB().SelectDatabase("13"); err != nil {
	// 	t.Error(err)
	// }
	// if _, rcvErr := onStor.DataDB().GetResourceLimit(rL.ID, false, utils.NonTransactional); rcvErr != utils.ErrNotFound {
	// 	t.Error(rcvErr)
	// }
	//
	if rcv, err := onStor.DataDB().GetResourceProfile(rL.Tenant, rL.ID, false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rL, rcv) {
		t.Errorf("Expecting: %v, received: %v", rL, rcv)
	}
	// if err = onStor.DataDB().SelectDatabase(onStorCfg); err != nil {
	// 	t.Error(err)
	// }
	if err := onStor.DataDB().RemoveResourceProfile(rL.Tenant, rL.ID, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	if _, rcvErr := onStor.DataDB().GetResourceProfile(rL.Tenant, rL.ID, true, utils.NonTransactional); rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
}

func testOnStorITCRUDResource(t *testing.T) {
	res := &Resource{
		Tenant: "cgrates.org",
		ID:     "RL1",
		Usages: map[string]*ResourceUsage{
			"RU1": &ResourceUsage{
				ID:         "RU1",
				ExpiryTime: time.Date(2014, 7, 3, 13, 43, 0, 0, time.UTC).Local(),
				Units:      2,
			},
		},
		TTLIdx: []string{"RU1"},
	}
	if _, rcvErr := onStor.DataDB().GetResource("cgrates.org", "RL1", true, utils.NonTransactional); rcvErr != nil && rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
	if err := onStor.DataDB().SetResource(res); err != nil {
		t.Error(err)
	}
	if rcv, err := onStor.DataDB().GetResource("cgrates.org", "RL1", true, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !(reflect.DeepEqual(res, rcv)) {
		t.Errorf("Expecting: %v, received: %v", res, rcv)
	}
	if rcv, err := onStor.DataDB().GetResource("cgrates.org", "RL1", false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(res, rcv) {
		t.Errorf("Expecting: %v, received: %v", res, rcv)
	}
	if err := onStor.DataDB().RemoveResource(res.Tenant, res.ID, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	if _, rcvErr := onStor.DataDB().GetResource(res.Tenant, res.ID, true, utils.NonTransactional); rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
}

func testOnStorITCRUDTiming(t *testing.T) {
	tmg := &utils.TPTiming{
		ID:        "TEST",
		Years:     utils.Years{2016, 2017},
		Months:    utils.Months{time.January, time.February, time.March},
		MonthDays: utils.MonthDays{1, 2, 3, 4},
		WeekDays:  utils.WeekDays{},
		StartTime: "00:00:00",
		EndTime:   "",
	}
	if _, rcvErr := onStor.GetTiming(tmg.ID, true, utils.NonTransactional); rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
	if err := onStor.SetTiming(tmg); err != nil {
		t.Error(err)
	}
	if rcv, err := onStor.GetTiming(tmg.ID, true, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(tmg, rcv) {
		t.Errorf("Expecting: %v, received: %v", tmg, rcv)
	}
	// FixMe
	// if err = onStor.DataDB().SelectDatabase("13"); err != nil {
	// 	t.Error(err)
	// }
	// if _, rcvErr := onStor.DataDB().GetTiming(tmg.ID, false, utils.NonTransactional); rcvErr != utils.ErrNotFound {
	// 	t.Error(rcvErr)
	// }
	//
	if rcv, err := onStor.GetTiming(tmg.ID, false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(tmg, rcv) {
		t.Errorf("Expecting: %v, received: %v", tmg, rcv)
	}
	// if err = onStor.DataDB().SelectDatabase(onStorCfg); err != nil {
	// 	t.Error(err)
	// }
	if err := onStor.RemoveTiming(tmg.ID, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	if _, rcvErr := onStor.GetTiming(tmg.ID, true, utils.NonTransactional); rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
}

func testOnStorITCRUDHistory(t *testing.T) {
	time := time.Date(2012, 1, 1, 0, 0, 0, 0, time.UTC).Local()
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

func testOnStorITCRUDStatQueueProfile(t *testing.T) {
	timeTTL := time.Duration(0 * time.Second)
	sq := &StatQueueProfile{
		ID:                 "test",
		ActivationInterval: &utils.ActivationInterval{},
		Filters:            []*RequestFilter{},
		QueueLength:        2,
		TTL:                timeTTL,
		Metrics:            []string{},
		Stored:             true,
		Thresholds:         []string{},
	}
	if _, rcvErr := onStor.GetStatQueueProfile(sq.Tenant, sq.ID, true, utils.NonTransactional); rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
	if _, ok := cache.Get(utils.StatQueueProfilePrefix + sq.ID); ok != false {
		t.Error("Should not be in cache")
	}
	if err := onStor.SetStatQueueProfile(sq); err != nil {
		t.Error(err)
	}
	if _, ok := cache.Get(utils.StatQueueProfilePrefix + sq.ID); ok != false {
		t.Error("Should not be in cache")
	}
	if rcv, err := onStor.GetStatQueueProfile(sq.Tenant, sq.ID, true, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(sq, rcv) {
		t.Errorf("Expecting: %v, received: %v", sq, rcv)
	}
	if _, ok := cache.Get(utils.StatQueueProfilePrefix + sq.ID); ok != false {
		t.Error("Should not be in cache")
	}
	if err := onStor.RemoveStatQueueProfile(sq.Tenant, sq.ID, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	if _, ok := cache.Get(utils.StatQueueProfilePrefix + sq.ID); ok != false {
		t.Error("Should not be in cache")
	}
	if _, rcvErr := onStor.GetStatQueueProfile(sq.Tenant, sq.ID, true, utils.NonTransactional); rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
}

func testOnStorITCRUDStoredStatQueue(t *testing.T) {
	eTime := utils.TimePointer(time.Date(2013, 10, 1, 0, 0, 0, 0, time.UTC).Local())
	asr := &StatASR{
		Answered: 2,
		Count:    3,
		Events: map[string]bool{
			"cgrates.org:ev1": true,
			"cgrates.org:ev2": true,
			"cgrates.org:ev3": false,
		},
	}
	msrshled, err := asr.Marshal(onStor.DataDB().Marshaler())
	if err != nil {
		t.Error(err)
	}
	sq := &StoredStatQueue{
		Tenant: "cgrates.org",
		ID:     "testOnStorITCRUDStatQueue",
		SQItems: []struct {
			EventID    string     // Bounded to the original StatEvent
			ExpiryTime *time.Time // Used to auto-expire events
		}{{EventID: "cgrates.org:ev1", ExpiryTime: eTime},
			{EventID: "cgrates.org:ev2", ExpiryTime: eTime},
			{EventID: "cgrates.org:ev3", ExpiryTime: eTime}},
		SQMetrics: map[string][]byte{
			utils.MetaASR: msrshled,
		},
	}
	if err := onStor.DataDB().SetStoredStatQueue(sq); err != nil {
		t.Error(err)
	}
	if rcv, err := onStor.DataDB().GetStoredStatQueue(sq.Tenant, sq.ID); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(sq, rcv) {
		t.Errorf("Expecting: %v, received: %v", sq, rcv)
	}
	if err := onStor.DataDB().RemStoredStatQueue(sq.Tenant, sq.ID); err != nil {
		t.Error(err)
	}
	if _, rcvErr := onStor.DataDB().GetStoredStatQueue(sq.Tenant, sq.ID); rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
}

func testOnStorITCRUDThresholdProfile(t *testing.T) {
	timeMinSleep := time.Duration(0 * time.Second)
	th := &ThresholdProfile{
		Tenant:             "cgrates.org",
		ID:                 "test",
		ActivationInterval: &utils.ActivationInterval{},
		Filters:            []*RequestFilter{},
		Recurrent:          true,
		MinSleep:           timeMinSleep,
		Blocker:            true,
		Weight:             1.4,
		ActionIDs:          []string{},
	}
	if _, rcvErr := onStor.GetThresholdProfile(th.Tenant, th.ID,
		false, utils.NonTransactional); rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
	if err := onStor.SetThresholdProfile(th); err != nil {
		t.Error(err)
	}
	if rcv, err := onStor.GetThresholdProfile(th.Tenant, th.ID, true, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(th, rcv) {
		t.Errorf("Expecting: %v, received: %v", th, rcv)
	}
	if rcv, err := onStor.GetThresholdProfile(th.Tenant, th.ID, false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(th, rcv) {
		t.Errorf("Expecting: %v, received: %v", th, rcv)
	}
	if err := onStor.RemoveThresholdProfile(th.Tenant, th.ID, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	if _, rcvErr := onStor.GetThresholdProfile(th.Tenant, th.ID, true, utils.NonTransactional); rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
	if _, rcvErr := onStor.GetThresholdProfile(th.Tenant, th.ID, false, utils.NonTransactional); rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
}

func testOnStorITCRUDThreshold(t *testing.T) {
	res := &Threshold{
		Tenant: "cgrates.org",
		ID:     "TH1",
		Snooze: time.Date(2016, 10, 1, 0, 0, 0, 0, time.UTC).Local(),
	}
	if _, rcvErr := onStor.GetThreshold("cgrates.org", "TH1", true, utils.NonTransactional); rcvErr != nil && rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
	if err := onStor.SetThreshold(res); err != nil {
		t.Error(err)
	}
	if rcv, err := onStor.GetThreshold("cgrates.org", "TH1", true, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !(reflect.DeepEqual(res, rcv)) {
		t.Errorf("Expecting: %v, received: %v", res, rcv)
	}
	if rcv, err := onStor.GetThreshold("cgrates.org", "TH1", false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(res, rcv) {
		t.Errorf("Expecting: %v, received: %v", res, rcv)
	}
	if err := onStor.RemoveThreshold(res.Tenant, res.ID, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	if _, rcvErr := onStor.GetThreshold(res.Tenant, res.ID, true, utils.NonTransactional); rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
}

func testOnStorITCRUDFilter(t *testing.T) {
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
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC).Local(),
			ExpiryTime:     time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC).Local(),
		},
	}
	if _, rcvErr := onStor.GetFilter("cgrates.org", "Filter1", true, utils.NonTransactional); rcvErr != nil && rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
	if err := onStor.SetFilter(fp); err != nil {
		t.Error(err)
	}
	if rcv, err := onStor.GetFilter("cgrates.org", "Filter1", true, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !(reflect.DeepEqual(fp, rcv)) {
		t.Errorf("Expecting: %v, received: %v", fp, rcv)
	}
	if rcv, err := onStor.GetFilter("cgrates.org", "Filter1", false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(fp, rcv) {
		t.Errorf("Expecting: %v, received: %v", fp, rcv)
	}
	if err := onStor.RemoveFilter(fp.Tenant, fp.ID, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	if _, rcvErr := onStor.GetFilter("cgrates.org", "Filter1", true, utils.NonTransactional); rcvErr != nil && rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
}
