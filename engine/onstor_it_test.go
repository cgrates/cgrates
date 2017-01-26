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
	onStor    DataDB
	onStorCfg string
)

// subtests to be executed for each confDIR
var sTestsOnStorIT = []func(t *testing.T){
	testOnStorITFlush,
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
	testOnStorITCacheResourceLimit,
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
	testOnStorITCRUDResourceLimit,
	testOnStorITCRUDHistory,
	testOnStorITCRUDStructVersion,
}

func TestOnStorITRedisConnect(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	rdsITdb, err = NewRedisStorage(fmt.Sprintf("%s:%s", cfg.TpDbHost, cfg.TpDbPort), 4, cfg.TpDbPass, cfg.DBDataEncoding, utils.REDIS_MAX_CONNS, nil, 1)
	if err != nil {
		t.Fatal("Could not connect to Redis", err.Error())
	}
	onStorCfg = cfg.DataDbName
}

func TestOnStorITRedis(t *testing.T) {
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
	if mgoITdb, err = NewMongoStorage(mgoITCfg.StorDBHost, mgoITCfg.StorDBPort, mgoITCfg.StorDBName, mgoITCfg.StorDBUser, mgoITCfg.StorDBPass,
		utils.StorDB, nil, mgoITCfg.CacheConfig, mgoITCfg.LoadHistorySize); err != nil {
		t.Fatal(err)
	}
	onStorCfg = mgoITCfg.StorDBName
}
func TestOnStorITMongo(t *testing.T) {
	onStor = mgoITdb
	for _, stest := range sTestsOnStorIT {
		t.Run("TestOnStorITMongo", stest)
	}
}

func testOnStorITFlush(t *testing.T) {
	if err := onStor.Flush(""); err != nil {
		t.Error(err)
	}
	cache.Flush()
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
	if err := onStor.SetDerivedChargers(keyCharger1, charger1, utils.NonTransactional); err != nil {
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
	if err := onStor.SetReqFilterIndexes(utils.ResourceLimitsIndex, idxes); err != nil {
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
	if idxes, err := onStor.GetReqFilterIndexes(utils.ResourceLimitsIndex); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eIdxes, idxes) {
		t.Errorf("Expecting: %+v, received: %+v", eIdxes, idxes)
	}
	if _, err := onStor.GetReqFilterIndexes("unknown_key"); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	}
}

func testOnStorITMatchReqFilterIndex(t *testing.T) {
	eMp := utils.StringMap{
		"RL1": true,
		"RL2": true,
	}
	if rcvMp, err := onStor.MatchReqFilterIndex(utils.ResourceLimitsIndex,
		utils.ConcatenatedKey("Account", "1002")); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eMp, rcvMp) {
		t.Errorf("Expecting: %+v, received: %+v", eMp, rcvMp)
	}
	if _, err := onStor.MatchReqFilterIndex(utils.ResourceLimitsIndex,
		utils.ConcatenatedKey("NonexistentField", "1002")); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	}
}

func testOnStorITCacheDestinations(t *testing.T) {
	if err := onStor.CacheDataFromDB("INVALID", nil, false); err == nil || err.Error() != utils.UnsupportedCachePrefix {
		t.Error(err)
	}
	dst := &Destination{Id: "TEST_CACHE", Prefixes: []string{"+491", "+492", "+493"}}
	if err := onStor.SetDestination(dst, utils.NonTransactional); err != nil {
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
	if err := onStor.SetReverseDestination(dst, utils.NonTransactional); err != nil {
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
	if err := onStor.SetRatingPlan(rp, utils.NonTransactional); err != nil {
		t.Error(err)
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
	if err := onStor.SetRatingProfile(rpf, utils.NonTransactional); err != nil {
		t.Error(err)
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
	if err := onStor.SetActions(acts[0].Id, acts, utils.NonTransactional); err != nil {
		t.Error(err)
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
	if err := onStor.SetActionPlan(ap.Id, ap, true, utils.NonTransactional); err != nil {
		t.Error(err)
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
	if err := onStor.SetAccountActionPlans(acntID, aAPs, true); err != nil {
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
	if err := onStor.SetActionTriggers(atsID, ats, utils.NonTransactional); err != nil {
		t.Error(err)
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
	if err := onStor.SetSharedGroup(sg, utils.NonTransactional); err != nil {
		t.Error(err)
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
	if err := onStor.SetDerivedChargers(keyDCS, dcs, utils.NonTransactional); err != nil {
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
	if err := onStor.SetLCR(lcr, utils.NonTransactional); err != nil {
		t.Error(err)
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
	if err := onStor.SetAlias(als, utils.NonTransactional); err != nil {
		t.Error(err)
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
	if err := onStor.SetReverseAlias(als, utils.NonTransactional); err != nil {
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

func testOnStorITCacheResourceLimit(t *testing.T) {
	rL := &ResourceLimit{
		ID:     "RL_TEST",
		Weight: 10,
		Filters: []*RequestFilter{
			&RequestFilter{Type: MetaString, FieldName: "Account", Values: []string{"dan", "1002"}},
			&RequestFilter{Type: MetaRSRFields, Values: []string{"Subject(~^1.*1$)", "Destination(1002)"},
				rsrFields: utils.ParseRSRFieldsMustCompile("Subject(~^1.*1$);Destination(1002)", utils.INFIELD_SEP),
			}},
		ActivationTime: time.Date(2014, 7, 3, 13, 43, 0, 0, time.UTC).Local(),
		ExpiryTime:     time.Date(2015, 7, 3, 13, 43, 0, 0, time.UTC).Local(),
		Limit:          1,
		ActionTriggers: make(ActionTriggers, 0),
		UsageTTL:       time.Duration(1 * time.Millisecond),
		Usage:          make(map[string]*ResourceUsage),
	}
	if err := onStor.SetResourceLimit(rL, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	if _, hasIt := cache.Get(utils.ResourceLimitsPrefix + rL.ID); hasIt {
		t.Error("Already in cache")
	}
	if err := onStor.CacheDataFromDB(utils.ResourceLimitsPrefix, []string{rL.ID}, false); err != nil {
		t.Error(err)
	}
	if itm, hasIt := cache.Get(utils.ResourceLimitsPrefix + rL.ID); !hasIt {
		t.Error("Did not cache")
	} else if rcv := itm.(*ResourceLimit); !reflect.DeepEqual(rL, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", rL.ActivationTime, rcv.ActivationTime)
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
	if rcv, err := onStor.HasData(utils.RATING_PLAN_PREFIX, rp.Id); err != nil {
		t.Error(err)
	} else if rcv != true {
		t.Errorf("Expecting: true, received: %v", rcv)
	}
}

func testOnStorITPushPop(t *testing.T) {
	if err := onStor.PushTask(&Task{Uuid: "1"}); err != nil {
		t.Error("Error pushing task: ", err)
	}
	if err := onStor.PushTask(&Task{Uuid: "2"}); err != nil {
		t.Error("Error pushing task: ", err)
	}
	if err := onStor.PushTask(&Task{Uuid: "3"}); err != nil {
		t.Error("Error pushing task: ", err)
	}
	if err := onStor.PushTask(&Task{Uuid: "4"}); err != nil {
		t.Error("Error pushing task: ", err)
	}
	if task, err := onStor.PopTask(); err != nil && task.Uuid != "1" {
		t.Error("Error poping task: ", task, err)
	}
	if task, err := onStor.PopTask(); err != nil && task.Uuid != "2" {
		t.Error("Error poping task: ", task, err)
	}
	if task, err := onStor.PopTask(); err != nil && task.Uuid != "3" {
		t.Error("Error poping task: ", task, err)
	}
	if task, err := onStor.PopTask(); err != nil && task.Uuid != "4" {
		t.Error("Error poping task: ", task, err)
	}
	if task, err := onStor.PopTask(); err == nil && task != nil {
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
	if _, rcvErr := onStor.GetRatingPlan(rp.Id, true, utils.NonTransactional); rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
	if err := onStor.SetRatingPlan(rp, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	if rcv, err := onStor.GetRatingPlan(rp.Id, true, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rp, rcv) {
		t.Errorf("Expecting: %v, received: %v", rp, rcv)
	}
	// FixMe
	// if err = onStor.SelectDatabase("13"); err != nil {
	// 	t.Error(err)
	// }
	// if _, rcvErr := onStor.GetRatingPlan(rp.Id, false, utils.NonTransactional); rcvErr != utils.ErrNotFound {
	// 	t.Error(rcvErr)
	// }

	if rcv, err := onStor.GetRatingPlan(rp.Id, false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rp, rcv) {
		t.Errorf("Expecting: %v, received: %v", rp, rcv)
	}
	// if err = onStor.SelectDatabase(onStorCfg); err != nil {
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
	if _, rcvErr := onStor.GetRatingProfile(rpf.Id, true, utils.NonTransactional); rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
	if err := onStor.SetRatingProfile(rpf, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	if rcv, err := onStor.GetRatingProfile(rpf.Id, true, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rpf, rcv) {
		t.Errorf("Expecting: %v, received: %v", rpf, rcv)
	}
	// FixMe
	// if err = onStor.SelectDatabase("13"); err != nil {
	// 	t.Error(err)
	// }
	// if _, rcvErr := onStor.GetRatingProfile(rpf.Id, false, utils.NonTransactional); rcvErr != utils.ErrNotFound {
	// 	t.Error(rcvErr)
	// }
	//
	if rcv, err := onStor.GetRatingProfile(rpf.Id, false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rpf, rcv) {
		t.Errorf("Expecting: %v, received: %v", rpf, rcv)
	}
	// if err = onStor.SelectDatabase(onStorCfg); err != nil {
	// 	t.Error(err)
	// }
	if err = onStor.RemoveRatingProfile(rpf.Id, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	if _, rcvErr := onStor.GetRatingProfile(rpf.Id, true, utils.NonTransactional); rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
}

func testOnStorITCRUDDestinations(t *testing.T) {
	dst := &Destination{Id: "CRUDDestination2", Prefixes: []string{"+491", "+492", "+493"}}
	if _, rcvErr := onStor.GetDestination(dst.Id, true, utils.NonTransactional); rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
	if err := onStor.SetDestination(dst, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	if rcv, err := onStor.GetDestination(dst.Id, true, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(dst, rcv) {
		t.Errorf("Expecting: %v, received: %v", dst, rcv)
	}
	// FixMe
	// if err = onStor.SelectDatabase("13"); err != nil {
	// 	t.Error(err)
	// }
	// if _, rcvErr := onStor.GetDestination(dst.Id, false, utils.NonTransactional); rcvErr != utils.ErrNotFound {
	// 	t.Error(rcvErr)
	// }
	//
	// if rcv, err := onStor.GetDestination(dst.Id, false, utils.NonTransactional); err != nil {
	// 	t.Error(err)
	// } else if !reflect.DeepEqual(dst, rcv) {
	// 	t.Errorf("Expecting: %v, received: %v", dst, rcv)
	// }
	// if err = onStor.SelectDatabase(onStorCfg); err != nil {
	// 	t.Error(err)
	// }

	if err = onStor.RemoveDestination(dst.Id, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	if _, rcvErr := onStor.GetDestination(dst.Id, true, utils.NonTransactional); rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
}

func testOnStorITCRUDReverseDestinations(t *testing.T) {
	dst := &Destination{Id: "CRUDReverseDestination", Prefixes: []string{"+494", "+495", "+496"}}
	dst2 := &Destination{Id: "CRUDReverseDestination2", Prefixes: []string{"+497", "+498", "+499"}}
	if _, rcvErr := onStor.GetReverseDestination(dst.Id, true, utils.NonTransactional); rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
	if err := onStor.SetReverseDestination(dst, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	for i, _ := range dst.Prefixes {
		if rcv, err := onStor.GetReverseDestination(dst.Prefixes[i], true, utils.NonTransactional); err != nil {
			t.Error(err)
		} else if !reflect.DeepEqual([]string{dst.Id}, rcv) {
			t.Errorf("Expecting: %v, received: %v", []string{dst.Id}, rcv)
		}
	}
	if err := onStor.UpdateReverseDestination(dst, dst2, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	for i, _ := range dst.Prefixes {
		if rcv, err := onStor.GetReverseDestination(dst2.Prefixes[i], true, utils.NonTransactional); err != nil {
			t.Error(err)
		} else if !reflect.DeepEqual([]string{dst2.Id}, rcv) {
			t.Errorf("Expecting: %v, received: %v", []string{dst.Id}, rcv)
		}
	}
	// FixMe
	// if err = onStor.SelectDatabase("13"); err != nil {
	// 	t.Error(err)
	// }
	// if _, rcvErr := onStor.GetReverseDestination(dst2.Id, false, utils.NonTransactional); rcvErr != utils.ErrNotFound {
	// 	t.Error(rcvErr)
	// }
	//
	for i, _ := range dst.Prefixes {
		if rcv, err := onStor.GetReverseDestination(dst2.Prefixes[i], false, utils.NonTransactional); err != nil {
			t.Error(err)
		} else if !reflect.DeepEqual([]string{dst2.Id}, rcv) {
			t.Errorf("Expecting: %v, received: %v", []string{dst.Id}, rcv)
		}
	}
	// if err = onStor.SelectDatabase(onStorCfg); err != nil {
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

	if _, rcvErr := onStor.GetLCR(lcr.GetId(), true, utils.NonTransactional); rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
	if err := onStor.SetLCR(lcr, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	if rcv, err := onStor.GetLCR(lcr.GetId(), true, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(lcr, rcv) {
		t.Errorf("Expecting: %v, received: %v", lcr, rcv)
	}
	// FixMe
	// if err = onStor.SelectDatabase("13"); err != nil {
	// 	t.Error(err)
	// }
	// if _, rcvErr := onStor.GetLCR(lcr.GetId(), false, utils.NonTransactional); rcvErr != utils.ErrNotFound {
	// 	t.Error(rcvErr)
	// }
	//
	if rcv, err := onStor.GetLCR(lcr.GetId(), false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(lcr, rcv) {
		t.Errorf("Expecting: %v, received: %v", lcr, rcv)
	}
	// if err = onStor.SelectDatabase(onStorCfg); err != nil {
	// 	t.Error(err)
	// }
}

func testOnStorITCRUDCdrStats(t *testing.T) {
	cdrs := &CdrStats{Metrics: []string{ASR, PDD, ACD, TCD, ACC, TCC, DDC}}

	if _, rcvErr := onStor.GetCdrStats(utils.NonTransactional); rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
	if err := onStor.SetCdrStats(cdrs); err != nil {
		t.Error(err)
	}
	if rcv, err := onStor.GetCdrStats(utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(cdrs.Metrics, rcv.Metrics) {
		t.Errorf("Expecting: %v, received: %v", cdrs.Metrics, rcv.Metrics)
	}
	if rcv, err := onStor.GetAllCdrStats(); err != nil {
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
	if _, rcvErr := onStor.GetActions(acts[0].Id, true, utils.NonTransactional); rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
	if err := onStor.SetActions(acts[0].Id, acts, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	if rcv, err := onStor.GetActions(acts[0].Id, true, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(acts[0], rcv[0]) {
		t.Errorf("Expecting: %v, received: %v", acts[0], rcv[0])
	}
	// FixMe
	// if err = onStor.SelectDatabase("13"); err != nil {
	// 	t.Error(err)
	// }
	// if _, rcvErr := onStor.GetActions(acts[0].Id, false, utils.NonTransactional); rcvErr != utils.ErrNotFound {
	// 	t.Error(rcvErr)
	// }
	//
	// if rcv, err := onStor.GetActions(acts[0].Id, false, utils.NonTransactional); err != nil {
	// 	t.Error(err)
	// } else if !reflect.DeepEqual(acts[0], rcv[0]) {
	// 	t.Errorf("Expecting: %v, received: %v", acts[0], rcv[0])
	// }
	// if err = onStor.SelectDatabase(onStorCfg); err != nil {
	// 	t.Error(err)
	// }

	if err := onStor.RemoveActions(acts[0].Id, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	if _, rcvErr := onStor.GetActions(acts[0].Id, true, utils.NonTransactional); rcvErr != utils.ErrNotFound {
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
	if _, rcvErr := onStor.GetSharedGroup(sg.Id, true, utils.NonTransactional); rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
	if err := onStor.SetSharedGroup(sg, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	if rcv, err := onStor.GetSharedGroup(sg.Id, true, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(sg, rcv) {
		t.Errorf("Expecting: %v, received: %v", sg, rcv)
	}
	// FixMe
	// if err = onStor.SelectDatabase("13"); err != nil {
	// 	t.Error(err)
	// }
	// if _, rcvErr := onStor.GetSharedGroup(sg.Id, false, utils.NonTransactional); rcvErr != utils.ErrNotFound {
	// 	t.Error(rcvErr)
	// }

	if rcv, err := onStor.GetSharedGroup(sg.Id, false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(sg, rcv) {
		t.Errorf("Expecting: %v, received: %v", sg, rcv)
	}
	// if err = onStor.SelectDatabase(onStorCfg); err != nil {
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
	// if err = onStor.SelectDatabase("13"); err != nil {
	// 	t.Error(err)
	// }
	// if _, rcvErr := onStor.GetActionTriggers(sg.Id, false, utils.NonTransactional); rcvErr != utils.ErrNotFound {
	// 	t.Error(rcvErr)
	// }
	//
	if rcv, err := onStor.GetActionTriggers(atsID, false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(ats[0], rcv[0]) {
		t.Errorf("Expecting: %v, received: %v", ats[0], rcv[0])
	}
	// if err = onStor.SelectDatabase(onStorCfg); err != nil {
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
	if _, rcvErr := onStor.GetActionPlan(ap.Id, true, utils.NonTransactional); rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
	if err := onStor.SetActionPlan(ap.Id, ap, true, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	if rcv, err := onStor.GetActionPlan(ap.Id, true, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(ap, rcv) {
		t.Errorf("Expecting: %v, received: %v", ap, rcv)
	}
	// FixMe
	// if err = onStor.SelectDatabase("13"); err != nil {
	// 	t.Error(err)
	// }
	// if _, rcvErr := onStor.GetActionPlan(ap.Id, false, utils.NonTransactional); rcvErr != utils.ErrNotFound {
	// 	t.Error(rcvErr)
	// }
	//
	// if rcv, err := onStor.GetActionPlan(ap.Id, false, utils.NonTransactional); err != nil {
	// 	t.Error(err)
	// } else if !reflect.DeepEqual(ap, rcv) {
	// 	t.Errorf("Expecting: %v, received: %v", ap, rcv)
	// }
	// if err = onStor.SelectDatabase(onStorCfg); err != nil {
	// 	t.Error(err)
	// }
	if rcv, err := onStor.GetAllActionPlans(); err != nil {
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
	if _, rcvErr := onStor.GetAccountActionPlans(acntID, true, utils.NonTransactional); rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
	if err := onStor.SetAccountActionPlans(acntID, aAPs, true); err != nil {
		t.Error(err)
	}
	if rcv, err := onStor.GetAccountActionPlans(acntID, true, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(aAPs, rcv) {
		t.Errorf("Expecting: %v, received: %v", aAPs, rcv)
	}
	if err := onStor.SetAccountActionPlans(acntID, aAPs2, false); err != nil {
		t.Error(err)
	}
	if rcv, err := onStor.GetAccountActionPlans(acntID, true, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expect, rcv) {
		t.Errorf("Expecting: %v, received: %v", expect, rcv)
	}
	// FixMe
	// if err = onStor.SelectDatabase("13"); err != nil {
	// 	t.Error(err)
	// }
	// if _, rcvErr := onStor.GetAccountActionPlans(acntID, false, utils.NonTransactional); rcvErr != utils.ErrNotFound {
	// 	t.Error(rcvErr)
	// }
	//
	if rcv, err := onStor.GetAccountActionPlans(acntID, false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expect, rcv) {
		t.Errorf("Expecting: %v, received: %v", expect, rcv)
	}
	// if err = onStor.SelectDatabase(onStorCfg); err != nil {
	// 	t.Error(err)
	// }
	if err := onStor.RemAccountActionPlans(acntID, aAPs2); err != nil {
		t.Error(err)
	}
	if rcv, err := onStor.GetAccountActionPlans(acntID, true, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(aAPs, rcv) {
		t.Errorf("Expecting: %v, received: %v", aAPs, rcv)
	}
	if err := onStor.RemAccountActionPlans(acntID, aAPs); err != nil {
		t.Error(err)
	}
	if _, rcvErr := onStor.GetAccountActionPlans(acntID, true, utils.NonTransactional); rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
}

func testOnStorITCRUDAccount(t *testing.T) {
	acc := &Account{
		ID:         utils.ConcatenatedKey("cgrates.org", "account2"),
		BalanceMap: map[string]Balances{utils.MONETARY: Balances{&Balance{Value: 10, Weight: 10}}},
	}
	if _, rcvErr := onStor.GetAccount(acc.ID); rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
	if err := onStor.SetAccount(acc); err != nil {
		t.Error(err)
	}
	if rcv, err := onStor.GetAccount(acc.ID); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(acc.ID, rcv.ID) {
		t.Errorf("Expecting: %v, received: %v", acc.ID, rcv.ID)
	} else if !reflect.DeepEqual(acc.BalanceMap[utils.MONETARY][0].Value, rcv.BalanceMap[utils.MONETARY][0].Value) {
		t.Errorf("Expecting: %v, received: %v", acc.BalanceMap[utils.MONETARY][0].Value, rcv.BalanceMap[utils.MONETARY][0].Value)
	} else if !reflect.DeepEqual(acc.BalanceMap[utils.MONETARY][0].Weight, rcv.BalanceMap[utils.MONETARY][0].Weight) {
		t.Errorf("Expecting: %v, received: %v", acc.BalanceMap[utils.MONETARY][0].Weight, rcv.BalanceMap[utils.MONETARY][0].Weight)
	}
	if err := onStor.RemoveAccount(acc.ID); err != nil {
		t.Error(err)
	}
	if _, rcvErr := onStor.GetAccount(acc.ID); rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
}

func testOnStorITCRUDCdrStatsQueue(t *testing.T) {
	sq := &StatsQueue{
		conf: &CdrStats{Id: "TTT"},
		Cdrs: []*QCdr{
			&QCdr{Cost: 9.0,
				SetupTime:  time.Date(2012, 1, 1, 0, 0, 0, 0, time.UTC).Local(),
				AnswerTime: time.Date(2012, 1, 1, 0, 0, 10, 0, time.UTC).Local(),
				EventTime:  time.Date(2012, 1, 1, 0, 0, 0, 0, time.UTC).Local(),
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
}

func testOnStorITCRUDSubscribers(t *testing.T) {
	if sbs, err := onStor.GetSubscribers(); err != nil {
		t.Error(err)
	} else if len(sbs) != 0 {
		t.Errorf("Received subscribers: %+v", sbs)
	}
	sbsc := &SubscriberData{
		ExpTime: time.Date(2012, 1, 1, 0, 0, 0, 0, time.UTC).Local(),
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

	if _, rcvErr := onStor.GetAlias(als.GetId(), true, utils.NonTransactional); rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
	if err := onStor.SetAlias(als, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	if rcv, err := onStor.GetAlias(als.GetId(), true, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(als, rcv) {
		t.Errorf("Expecting: %v, received: %v", als, rcv)
	}
	// FixMe
	// if err = onStor.SelectDatabase("13"); err != nil {
	// 	t.Error(err)
	// }
	// if _, rcvErr := onStor.GetAlias(als.GetId(), false, utils.NonTransactional); rcvErr != utils.ErrNotFound {
	// 	t.Error(rcvErr)
	// }
	//
	if rcv, err := onStor.GetAlias(als.GetId(), false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(als, rcv) {
		t.Errorf("Expecting: %v, received: %v", als, rcv)
	}
	// if err = onStor.SelectDatabase(onStorCfg); err != nil {
	// 	t.Error(err)
	// }
	if err := onStor.RemoveAlias(als.GetId(), utils.NonTransactional); err != nil {
		t.Error(err)
	}
	if _, rcvErr := onStor.GetAlias(als.GetId(), true, utils.NonTransactional); rcvErr != utils.ErrNotFound {
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
	if _, rcvErr := onStor.GetReverseAlias(rvAlsID, true, utils.NonTransactional); rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
	if err := onStor.SetReverseAlias(als, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	if rcv, err := onStor.GetReverseAlias(rvAlsID, true, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(exp, rcv[0]) {
		t.Errorf("Expecting: %v, received: %v", exp, rcv[0])
	}
	// FixMe
	// if err = onStor.SelectDatabase("13"); err != nil {
	// 	t.Error(err)
	// }
	// if _, rcvErr := onStor.GetReverseAlias(rvAlsID, false, utils.NonTransactional); rcvErr != utils.ErrNotFound {
	// 	t.Error(rcvErr)
	// }
	//
	if rcv, err := onStor.GetReverseAlias(rvAlsID, false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(exp, rcv[0]) {
		t.Errorf("Expecting: %v, received: %v", exp, rcv[0])
	}
	// if err = onStor.SelectDatabase(onStorCfg); err != nil {
	// 	t.Error(err)
	// }
}

func testOnStorITCRUDResourceLimit(t *testing.T) {
	rL := &ResourceLimit{
		ID:     "RL_TEST2",
		Weight: 10,
		Filters: []*RequestFilter{
			&RequestFilter{Type: MetaString, FieldName: "Account", Values: []string{"dan", "1002"}},
			&RequestFilter{Type: MetaRSRFields, Values: []string{"Subject(~^1.*1$)", "Destination(1002)"},
				rsrFields: utils.ParseRSRFieldsMustCompile("Subject(~^1.*1$);Destination(1002)", utils.INFIELD_SEP),
			}},
		ActivationTime: time.Date(2014, 7, 3, 13, 43, 0, 0, time.UTC).Local(),
		ExpiryTime:     time.Date(2015, 7, 3, 13, 43, 0, 0, time.UTC).Local(),
		Limit:          1,
		ActionTriggers: make(ActionTriggers, 0),
		UsageTTL:       time.Duration(1 * time.Millisecond),
		Usage:          make(map[string]*ResourceUsage),
	}
	if _, rcvErr := onStor.GetResourceLimit(rL.ID, true, utils.NonTransactional); rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
	if err := onStor.SetResourceLimit(rL, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	if rcv, err := onStor.GetResourceLimit(rL.ID, true, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rL, rcv) {
		t.Errorf("Expecting: %v, received: %v", rL, rcv)
	}
	// FixMe
	// if err = onStor.SelectDatabase("13"); err != nil {
	// 	t.Error(err)
	// }
	// if _, rcvErr := onStor.GetResourceLimit(rL.ID, false, utils.NonTransactional); rcvErr != utils.ErrNotFound {
	// 	t.Error(rcvErr)
	// }
	//
	if rcv, err := onStor.GetResourceLimit(rL.ID, false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rL, rcv) {
		t.Errorf("Expecting: %v, received: %v", rL, rcv)
	}
	// if err = onStor.SelectDatabase(onStorCfg); err != nil {
	// 	t.Error(err)
	// }
	if err := onStor.RemoveResourceLimit(rL.ID, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	if _, rcvErr := onStor.GetResourceLimit(rL.ID, true, utils.NonTransactional); rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
}

func testOnStorITCRUDHistory(t *testing.T) {
	time := time.Date(2012, 1, 1, 0, 0, 0, 0, time.UTC).Local()
	ist := &utils.LoadInstance{"Load", "RatingLoad", "Account", time}
	if err := onStor.AddLoadHistory(ist, 1, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	if rcv, err := onStor.GetLoadHistory(1, true, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(ist, rcv[0]) {
		t.Errorf("Expecting: %v, received: %v", ist, rcv[0])
	}
}

func testOnStorITCRUDStructVersion(t *testing.T) {
	cv := &StructVersion{
		Destinations:    "1",
		RatingPlans:     "1",
		RatingProfiles:  "1",
		Lcrs:            "1",
		DerivedChargers: "1",
		Actions:         "1",
		ActionPlans:     "1",
		ActionTriggers:  "1",
		SharedGroups:    "1",
		Accounts:        "1",
		CdrStats:        "1",
		Users:           "1",
		Alias:           "1",
		PubSubs:         "1",
		LoadHistory:     "1",
		Cdrs:            "1",
		SMCosts:         "1",
		ResourceLimits:  "1",
	}
	if _, rcvErr := onStor.GetStructVersion(); rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
	if err := onStor.SetStructVersion(CurrentVersion); err != nil {
		t.Error(err)
	}
	if rcv, err := onStor.GetStructVersion(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(cv, rcv) {
		t.Errorf("Expecting: %v, received: %v", cv, rcv)
	}
}
