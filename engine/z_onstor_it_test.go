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
	"testing"
	"time"

	"github.com/ericlagergren/decimal"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

var (
	rdsITdb   *RedisStorage
	mgoITdb   *MongoStorage
	onStor    *DataManager
	onStorCfg string

	// subtests to be executed for each confDIR
	sTestsOnStorIT = []func(t *testing.T){
		testOnStorITFlush,
		testOnStorITIsDBEmpty,
		testOnStorITCacheDestinations,
		testOnStorITCacheReverseDestinations,
		testOnStorITCacheActionPlan,
		testOnStorITCacheAccountActionPlans,

		// ToDo: test cache flush for a prefix
		// ToDo: testOnStorITLoadAccountingCache
		testOnStorITHasData,
		testOnStorITPushPop,
		testOnStorITRatingPlan,
		testOnStorITRatingProfile,
		testOnStorITCRUDDestinations,
		testOnStorITCRUDReverseDestinations,
		testOnStorITActions,
		testOnStorITSharedGroup,
		testOnStorITCRUDActionPlan,
		testOnStorITCRUDAccountActionPlans,
		testOnStorITCRUDAccount,
		testOnStorITResource,
		testOnStorITResourceProfile,
		testOnStorITTiming,
		//testOnStorITCRUDHistory,
		testOnStorITCRUDStructVersion,
		testOnStorITStatQueueProfile,
		testOnStorITStatQueue,
		testOnStorITThresholdProfile,
		testOnStorITThreshold,
		testOnStorITFilter,
		testOnStorITRouteProfile,
		testOnStorITAttributeProfile,
		testOnStorITFlush,
		testOnStorITIsDBEmpty,
		testOnStorITTestAttributeSubstituteIface,
		testOnStorITChargerProfile,
		testOnStorITDispatcherProfile,
		testOnStorITRateProfile,
		testOnStorITActionProfile,
		testOnStorITAccountProfile,
		//testOnStorITCacheActionTriggers,
		//testOnStorITCRUDActionTriggers,
	}
)

func TestOnStorIT(t *testing.T) {
	switch *dbType {
	case utils.MetaInternal:
		onStor = NewDataManager(NewInternalDB(nil, nil, true),
			config.CgrConfig().CacheCfg(), nil)
	case utils.MetaMySQL:
		cfg := config.NewDefaultCGRConfig()
		rdsITdb, err = NewRedisStorage(
			fmt.Sprintf("%s:%s", cfg.DataDbCfg().Host, cfg.DataDbCfg().Port),
			4, cfg.DataDbCfg().User, cfg.DataDbCfg().Password, cfg.GeneralCfg().DBDataEncoding,
			utils.RedisMaxConns, "", false, 0, 0, false, utils.EmptyString, utils.EmptyString, utils.EmptyString)
		if err != nil {
			t.Fatal("Could not connect to Redis", err.Error())
		}
		onStorCfg = cfg.DataDbCfg().Name
		onStor = NewDataManager(rdsITdb, config.CgrConfig().CacheCfg(), nil)
	case utils.MetaMongo:
		cdrsMongoCfgPath := path.Join(*dataDir, "conf", "samples", "cdrsv2mongo")
		mgoITCfg, err := config.NewCGRConfigFromPath(cdrsMongoCfgPath)
		if err != nil {
			t.Fatal(err)
		}
		if mgoITdb, err = NewMongoStorage(mgoITCfg.StorDbCfg().Host,
			mgoITCfg.StorDbCfg().Port, mgoITCfg.StorDbCfg().Name,
			mgoITCfg.StorDbCfg().User, mgoITCfg.StorDbCfg().Password,
			mgoITCfg.GeneralCfg().DBDataEncoding,
			utils.StorDB, nil, 10*time.Second); err != nil {
			t.Fatal(err)
		}
		onStorCfg = mgoITCfg.StorDbCfg().Name
		onStor = NewDataManager(mgoITdb, config.CgrConfig().CacheCfg(), nil)
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}

	for _, stest := range sTestsOnStorIT {
		t.Run(*dbType, stest)
	}
}

func testOnStorITFlush(t *testing.T) {
	if err := onStor.DataDB().Flush(""); err != nil {
		t.Error(err)
	}
	Cache.Clear(nil)
}

func testOnStorITIsDBEmpty(t *testing.T) {
	test, err := onStor.DataDB().IsDBEmpty()
	if err != nil {
		t.Error(err)
	} else if test != true {
		t.Errorf("Expecting: true got :%+v", test)
	}
}

func testOnStorITCacheDestinations(t *testing.T) {
	if onStor.dataDB.GetStorageType() == utils.INTERNAL {
		t.SkipNow()
	}

	if err := onStor.CacheDataFromDB("INVALID", nil, false); err == nil || err.Error() != utils.UnsupportedCachePrefix {
		t.Error(err)
	}
	dst := &Destination{Id: "TEST_CACHE", Prefixes: []string{"+491", "+492", "+493"}}
	if err := onStor.SetDestination(dst, utils.NonTransactional); err != nil {
		t.Error(err)
	}

	if _, hasIt := Cache.Get(utils.CacheDestinations, dst.Id); hasIt {
		t.Error("Already in cache")
	}
	if err := onStor.CacheDataFromDB(utils.DestinationPrefix, []string{dst.Id}, true); err != nil { // Should not cache due to mustBeCached
		t.Error(err)
	}
	if _, hasIt := Cache.Get(utils.CacheDestinations, dst.Id); hasIt {
		t.Error("Should not be in cache")
	}
	if err := onStor.CacheDataFromDB(utils.DestinationPrefix, []string{dst.Id}, false); err != nil {
		t.Error(err)
	}
	if itm, hasIt := Cache.Get(utils.CacheDestinations, dst.Id); !hasIt {
		t.Error("Did not cache")
	} else if !reflect.DeepEqual(dst, itm.(*Destination)) {
		t.Error("Wrong item in the cache")
	}
}

func testOnStorITCacheReverseDestinations(t *testing.T) {
	dst := &Destination{Id: "TEST_CACHE", Prefixes: []string{"+491", "+492", "+493"}}
	if err := onStor.SetReverseDestination(dst.Id, dst.Prefixes, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	for _, prfx := range dst.Prefixes {
		if _, hasIt := Cache.Get(utils.CacheReverseDestinations, dst.Id); hasIt &&
			onStor.dataDB.GetStorageType() != utils.INTERNAL {
			t.Errorf("Prefix: %s already in cache", prfx)
		}
	}
	if err := onStor.CacheDataFromDB(utils.ReverseDestinationPrefix, dst.Prefixes, false); err != nil {
		t.Error(err)
	}
	if onStor.dataDB.GetStorageType() != utils.INTERNAL {
		for _, prfx := range dst.Prefixes {
			if itm, hasIt := Cache.Get(utils.CacheReverseDestinations, prfx); !hasIt {
				t.Error("Did not cache")
			} else if !reflect.DeepEqual([]string{dst.Id}, itm.([]string)) {
				t.Error("Wrong item in the cache")
			}
		}
	}
}

func testOnStorITCacheActionPlan(t *testing.T) {
	ap := &ActionPlan{
		Id:         "MORE_MINUTES",
		AccountIDs: utils.StringMap{"vdf:minitsboy": true},
		ActionTimings: []*ActionTiming{
			{
				Uuid: utils.GenUUID(),
				Timing: &RateInterval{
					Timing: &RITiming{
						Years:     utils.Years{2012},
						Months:    utils.Months{},
						MonthDays: utils.MonthDays{},
						WeekDays:  utils.WeekDays{},
						StartTime: utils.MetaASAP,
					},
				},
				Weight:    10,
				ActionsID: "MINI",
			},
			{
				Uuid: utils.GenUUID(),
				Timing: &RateInterval{
					Timing: &RITiming{
						Years:     utils.Years{2012},
						Months:    utils.Months{},
						MonthDays: utils.MonthDays{},
						WeekDays:  utils.WeekDays{},
						StartTime: utils.MetaASAP,
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
	expectedCAp := []string{"apl_MORE_MINUTES"}
	if itm, err := onStor.DataDB().GetKeysForPrefix(utils.ActionPlanPrefix); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedCAp, itm) {
		t.Errorf("Expected : %+v, but received %+v", expectedCAp, itm)
	}
	if _, hasIt := Cache.Get(utils.CacheActionPlans, ap.Id); hasIt &&
		onStor.dataDB.GetStorageType() != utils.INTERNAL {
		t.Error("Already in cache")
	}
	if err := onStor.CacheDataFromDB(utils.ActionPlanPrefix, []string{ap.Id}, false); err != nil {
		t.Error(err)
	}
	if onStor.dataDB.GetStorageType() != utils.INTERNAL {
		if itm, hasIt := Cache.Get(utils.CacheActionPlans, ap.Id); !hasIt {
			t.Error("Did not cache")
		} else if rcv := itm.(*ActionPlan); !reflect.DeepEqual(ap, rcv) {
			t.Errorf("Expecting: %+v, received: %+v", ap, rcv)
		}
	}
	if err := onStor.RemoveActionPlan(ap.Id, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	if err := onStor.SetActionPlan(ap.Id, ap, true, utils.NonTransactional); err != nil {
		t.Error(err)
	}
}

func testOnStorITCacheAccountActionPlans(t *testing.T) {
	acntID := utils.ConcatenatedKey("cgrates.org", "1001")
	aAPs := []string{"PACKAGE_10_SHARED_A_5", "USE_SHARED_A", "apl_PACKAGE_1001"}
	if err := onStor.SetAccountActionPlans(acntID, aAPs, true); err != nil {
		t.Error(err)
	}
	if _, hasIt := Cache.Get(utils.CacheAccountActionPlans, acntID); hasIt &&
		onStor.dataDB.GetStorageType() != utils.INTERNAL {
		t.Error("Already in cache")
	}
	if err := onStor.CacheDataFromDB(utils.AccountActionPlansPrefix, []string{acntID}, false); err != nil {
		t.Error(err)
	}
	if onStor.dataDB.GetStorageType() != utils.INTERNAL {
		if itm, hasIt := Cache.Get(utils.CacheAccountActionPlans, acntID); !hasIt {
			t.Error("Did not cache")
		} else if rcv := itm.([]string); !reflect.DeepEqual(aAPs, rcv) {
			t.Errorf("Expecting: %+v, received: %+v", aAPs, rcv)
		}
	}
}

func testOnStorITCacheActionTriggers(t *testing.T) {
	ats := ActionTriggers{
		&ActionTrigger{
			ID: "testOnStorITCacheActionTrigger",
			Balance: &BalanceFilter{Type: utils.StringPointer(utils.MetaMonetary),
				Timings: make([]*RITiming, 0)},
			ThresholdValue:    2,
			ThresholdType:     utils.TriggerMaxEventCounter,
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
	if itm, err := onStor.DataDB().GetKeysForPrefix(utils.ActionTriggerPrefix); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedCAt, itm) {
		t.Errorf("Expected : %+v, but received %+v", expectedCAt, itm)
	}
	if _, hasIt := Cache.Get(utils.CacheActionTriggers, atsID); hasIt &&
		onStor.dataDB.GetStorageType() != utils.INTERNAL {
		t.Error("Already in cache")
	}
	if err := onStor.CacheDataFromDB(utils.ActionTriggerPrefix, []string{atsID}, false); err != nil {
		t.Error(err)
	}
	if itm, hasIt := Cache.Get(utils.CacheActionTriggers, atsID); !hasIt {
		t.Error("Did not cache")
	} else if rcv := itm.(ActionTriggers); !reflect.DeepEqual(ats, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", ats, rcv)
	}
}

func testOnStorITHasData(t *testing.T) {
	rp := &RatingPlan{
		Id: "HasData",
		Timings: map[string]*RITiming{
			"59a981b9": {
				Years:     utils.Years{},
				Months:    utils.Months{},
				MonthDays: utils.MonthDays{},
				WeekDays:  utils.WeekDays{1, 2, 3, 4, 5},
				StartTime: "00:00:00",
			},
		},
		Ratings: map[string]*RIRate{
			"ebefae11": {
				ConnectFee: 0,
				Rates: []*RGRate{
					{
						GroupIntervalStart: 0,
						Value:              0.2,
						RateIncrement:      time.Second,
						RateUnit:           time.Minute,
					},
				},
				RoundingMethod:   utils.MetaRoundingMiddle,
				RoundingDecimals: 4,
			},
		},
		DestinationRates: map[string]RPRateList{
			"GERMANY": []*RPRate{
				{
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
	if itm, err := onStor.DataDB().GetKeysForPrefix(utils.RatingPlanPrefix); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(len(expectedRP), len(itm)) {
		t.Errorf("Expected : %+v, but received %+v", len(expectedRP), len(itm))
	}
	if rcv, err := onStor.HasData(utils.RatingPlanPrefix, rp.Id, ""); err != nil {
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
			"59a981b9": {
				Years:     utils.Years{},
				Months:    utils.Months{},
				MonthDays: utils.MonthDays{},
				WeekDays:  utils.WeekDays{1, 2, 3, 4, 5},
				StartTime: "00:00:00",
			},
		},
		Ratings: map[string]*RIRate{
			"ebefae11": {
				ConnectFee: 0,
				Rates: []*RGRate{
					{
						GroupIntervalStart: 0,
						Value:              0.2,
						RateIncrement:      time.Second,
						RateUnit:           time.Minute,
					},
				},
				RoundingMethod:   utils.MetaRoundingMiddle,
				RoundingDecimals: 4,
			},
		},
		DestinationRates: map[string]RPRateList{
			"GERMANY": []*RPRate{
				{
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
	if onStor.dataDB.GetStorageType() != utils.INTERNAL {
		//get from cache
		if rcv, err := onStor.GetRatingPlan(rp.Id, false, utils.NonTransactional); err != nil {
			t.Error(err)
		} else if !reflect.DeepEqual(rp, rcv) {
			t.Errorf("Expecting: %v, received: %v", rp, rcv)
		}
	}
	//get from database
	if rcv, err := onStor.GetRatingPlan(rp.Id, true, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rp, rcv) {
		t.Errorf("Expecting: %v, received: %v", rp, rcv)
	}
	expectedRP := []string{"rpl_HasData", "rpl_CRUDRatingPlan"}
	if itm, err := onStor.DataDB().GetKeysForPrefix(utils.RatingPlanPrefix); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(len(expectedRP), len(itm)) {
		t.Errorf("Expected : %+v, but received %+v", len(expectedRP), len(itm))
	}
	//update
	rp.Timings = map[string]*RITiming{
		"59a981b9": {
			Years:     utils.Years{},
			Months:    utils.Months{},
			MonthDays: utils.MonthDays{},
			WeekDays:  utils.WeekDays{1, 2, 3, 4, 5},
			StartTime: "00:00:00",
		},
		"59a981a1": {
			Years:     utils.Years{},
			Months:    utils.Months{},
			MonthDays: utils.MonthDays{},
			WeekDays:  utils.WeekDays{6, 7},
			StartTime: "00:00:00",
		},
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
				ActivationTime: time.Date(2013, 10, 1, 0, 0, 0, 0, time.UTC),
				RatingPlanId:   "TDRT",
				FallbackKeys:   []string{"*out:test:1:danb", "*out:test:1:rif"},
			}},
	}
	if _, rcvErr := onStor.GetRatingProfile(rpf.Id, false,
		utils.NonTransactional); rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
	if err := onStor.SetRatingProfile(rpf, utils.NonTransactional); err != nil {
		t.Error(err)
	}

	//get from database
	if rcv, err := onStor.GetRatingProfile(rpf.Id, true,
		utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rpf, rcv) {
		t.Errorf("Expecting: %v, received: %v", utils.ToJSON(rpf), utils.ToJSON(rcv))
	}

	if onStor.dataDB.GetStorageType() != utils.INTERNAL {
		//get from cache
		if rcv, err := onStor.GetRatingProfile(rpf.Id, false,
			utils.NonTransactional); err != nil {
			t.Error(err)
		} else if !reflect.DeepEqual(rpf, rcv) {
			t.Errorf("Expecting: %v, received: %v", utils.ToJSON(rpf), utils.ToJSON(rcv))
		}
	}

	expectedCRPl := []string{"rpf_*out:test:1:trp"}
	if itm, err := onStor.DataDB().GetKeysForPrefix(utils.RatingProfilePrefix); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedCRPl, itm) {
		t.Errorf("Expected : %+v, but received %+v", expectedCRPl, itm)
	}
	//update
	rpf.RatingPlanActivations = RatingPlanActivations{
		&RatingPlanActivation{
			ActivationTime: time.Date(2013, 10, 1, 0, 0, 0, 0, time.UTC),
			RatingPlanId:   "TDRT",
			FallbackKeys:   []string{"*out:test:1:danb", "*out:test:1:teo"},
		},
	}
	if err := onStor.SetRatingProfile(rpf, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	//get from database
	if rcv, err := onStor.GetRatingProfile(rpf.Id, true,
		utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rpf, rcv) {
		t.Errorf("Expecting: %v, received: %v", utils.ToJSON(rpf), utils.ToJSON(rcv))
	}
	if onStor.dataDB.GetStorageType() != utils.INTERNAL {
		//get from cache
		if rcv, err := onStor.GetRatingProfile(rpf.Id, false,
			utils.NonTransactional); err != nil {
			t.Error(err)
		} else if !reflect.DeepEqual(rpf, rcv) {
			t.Errorf("Expecting: %v, received: %v", utils.ToJSON(rpf), utils.ToJSON(rcv))
		}
	}

	if err = onStor.RemoveRatingProfile(rpf.Id, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	//check database if removed
	if _, rcvErr := onStor.GetRatingProfile(rpf.Id, true,
		utils.NonTransactional); rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
	//check cache if removed
	if _, rcvErr := onStor.GetRatingProfile(rpf.Id, false,
		utils.NonTransactional); rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
}

func testOnStorITCRUDDestinations(t *testing.T) {
	dst := &Destination{Id: "CRUDDestination2", Prefixes: []string{"+491", "+492", "+493"}}
	if _, rcvErr := onStor.GetDestination(dst.Id, false, true, utils.NonTransactional); rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
	if err := onStor.SetDestination(dst, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	if rcv, err := onStor.GetDestination(dst.Id, false, true, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(dst, rcv) {
		t.Errorf("Expecting: %v, received: %v", dst, rcv)
	}
	// FixMe
	// if err = onStor.DataDB().SelectDatabase("13"); err != nil {
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
	// if err = onStor.DataDB().SelectDatabase(onStorCfg); err != nil {
	// 	t.Error(err)
	// }

	if err = onStor.RemoveDestination(dst.Id, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	if _, rcvErr := onStor.GetDestination(dst.Id, false, true, utils.NonTransactional); rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
}

func testOnStorITCRUDReverseDestinations(t *testing.T) {
	dst := &Destination{Id: "CRUDReverseDestination", Prefixes: []string{"+494", "+495", "+496"}}
	dst2 := &Destination{Id: "CRUDReverseDestination", Prefixes: []string{"+497", "+498", "+499"}}
	if _, rcvErr := onStor.GetReverseDestination(dst.Id, false, true, utils.NonTransactional); rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
	if err := onStor.SetReverseDestination(dst.Id, dst.Prefixes, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	for i := range dst.Prefixes {
		if rcv, err := onStor.GetReverseDestination(dst.Prefixes[i], false, true, utils.NonTransactional); err != nil {
			t.Error(err)
		} else if !reflect.DeepEqual([]string{dst.Id}, rcv) {
			t.Errorf("Expecting: %v, received: %v", []string{dst.Id}, rcv)
		}
	}
	if err := onStor.UpdateReverseDestination(dst, dst2, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	for i := range dst.Prefixes {
		if rcv, err := onStor.GetReverseDestination(dst2.Prefixes[i], false, true, utils.NonTransactional); err != nil {
			t.Error(err)
		} else if !reflect.DeepEqual([]string{dst2.Id}, rcv) {
			t.Errorf("Expecting: %v, received: %v", []string{dst.Id}, rcv)
		}
	}
	// FixMe
	// if err = onStor.DataDB().SelectDatabase("13"); err != nil {
	// 	t.Error(err)
	// }
	// if _, rcvErr := onStor.GetReverseDestination(dst2.Id, false, utils.NonTransactional); rcvErr != utils.ErrNotFound {
	// 	t.Error(rcvErr)
	// }
	//
	for i := range dst.Prefixes {
		if rcv, err := onStor.GetReverseDestination(dst2.Prefixes[i], true, true, utils.NonTransactional); err != nil {
			t.Error(err)
		} else if !reflect.DeepEqual([]string{dst2.Id}, rcv) {
			t.Errorf("Expecting: %v, received: %v", []string{dst.Id}, rcv)
		}
	}
	// if err = onStor.DataDB().SelectDatabase(onStorCfg); err != nil {
	// 	t.Error(err)
	// }
}

func testOnStorITActions(t *testing.T) {
	acts := Actions{
		&Action{
			Id:               "MINI",
			ActionType:       utils.MetaTopUpReset,
			ExpirationString: utils.MetaUnlimited,
			Weight:           10,
			Balance: &BalanceFilter{
				Type: utils.StringPointer(utils.MetaMonetary),
				Uuid: utils.StringPointer(utils.GenUUID()),
				Value: &utils.ValueFormula{Static: 10,
					Params: make(map[string]interface{})},
				Weight:   utils.Float64Pointer(10),
				Disabled: utils.BoolPointer(false),
				Timings: []*RITiming{
					{
						Years:     utils.Years{2016, 2017},
						Months:    utils.Months{time.January, time.February, time.March},
						MonthDays: utils.MonthDays{1, 2, 3, 4},
						WeekDays:  utils.WeekDays{1, 2, 3},
						StartTime: utils.MetaASAP,
					},
				},
				Blocker: utils.BoolPointer(false),
			},
		},
		&Action{
			Id:               "MINI",
			ActionType:       utils.MetaTopUp,
			ExpirationString: utils.MetaUnlimited,
			Weight:           10,
			Balance: &BalanceFilter{
				Type: utils.StringPointer(utils.MetaVoice),
				Uuid: utils.StringPointer(utils.GenUUID()),
				Value: &utils.ValueFormula{Static: 100,
					Params: make(map[string]interface{})},
				Weight:         utils.Float64Pointer(10),
				RatingSubject:  utils.StringPointer("test"),
				DestinationIDs: utils.StringMapPointer(utils.NewStringMap("NAT")),
				Disabled:       utils.BoolPointer(false),
				Timings: []*RITiming{
					{
						Years:     utils.Years{2016, 2017},
						Months:    utils.Months{time.January, time.February, time.March},
						MonthDays: utils.MonthDays{1, 2, 3, 4},
						WeekDays:  utils.WeekDays{1, 2, 3},
						StartTime: utils.MetaASAP,
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
	if rcv, err := onStor.HasData(utils.ActionPrefix, acts[0].Id, ""); err != nil {
		t.Error(err)
	} else if rcv != true {
		t.Errorf("Expecting: true, received: %v", rcv)
	}
	//get from database
	if rcv, err := onStor.GetActions(acts[0].Id,
		true, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(acts[0], rcv[0]) {
		t.Errorf("Expecting: %v, received: %v", utils.ToJSON(acts[0]), utils.ToJSON(rcv[0]))
	}
	if onStor.dataDB.GetStorageType() != utils.INTERNAL {
		//get from cache
		if rcv, err := onStor.GetActions(acts[0].Id,
			false, utils.NonTransactional); err != nil {
			t.Error(err)
		} else if !reflect.DeepEqual(acts[0], rcv[0]) {
			t.Errorf("Expecting: %v, received: %v", utils.ToJSON(acts[0]), utils.ToJSON(rcv[0]))
		}
	}
	expectedCA := []string{"act_MINI"}
	if itm, err := onStor.DataDB().GetKeysForPrefix(utils.ActionPrefix); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedCA, itm) {
		t.Errorf("Expected : %+v, but received %+v", expectedCA, itm)
	}
	//update
	acts = Actions{
		&Action{
			Id:               "MINI",
			ActionType:       utils.MetaTopUpReset,
			ExpirationString: utils.MetaUnlimited,
			Weight:           10,
			Balance: &BalanceFilter{
				Type: utils.StringPointer(utils.MetaMonetary),
				Uuid: utils.StringPointer(utils.GenUUID()),
				Value: &utils.ValueFormula{Static: 10,
					Params: make(map[string]interface{})},
				Weight:   utils.Float64Pointer(10),
				Disabled: utils.BoolPointer(false),
				Timings: []*RITiming{
					{
						Years:     utils.Years{2016, 2017},
						Months:    utils.Months{time.January, time.February, time.March},
						MonthDays: utils.MonthDays{1, 2, 3, 4},
						WeekDays:  utils.WeekDays{1, 2, 3},
						StartTime: utils.MetaASAP,
					},
				},
				Blocker: utils.BoolPointer(false),
			},
		},
		&Action{
			Id:               "MINI",
			ActionType:       utils.MetaTopUp,
			ExpirationString: utils.MetaUnlimited,
			Weight:           10,
			Balance: &BalanceFilter{
				Type: utils.StringPointer(utils.MetaVoice),
				Uuid: utils.StringPointer(utils.GenUUID()),
				Value: &utils.ValueFormula{Static: 100,
					Params: make(map[string]interface{})},
				Weight:         utils.Float64Pointer(10),
				RatingSubject:  utils.StringPointer("test"),
				DestinationIDs: utils.StringMapPointer(utils.NewStringMap("NAT")),
				Disabled:       utils.BoolPointer(false),
				Timings: []*RITiming{
					{
						Years:     utils.Years{2016, 2017},
						Months:    utils.Months{time.January, time.February, time.March},
						MonthDays: utils.MonthDays{1, 2, 3, 4},
						WeekDays:  utils.WeekDays{1, 2, 3},
						StartTime: utils.MetaASAP,
					},
				},
				Blocker: utils.BoolPointer(false),
			},
		},
		&Action{
			Id:               "MINI",
			ActionType:       utils.MetaDebit,
			ExpirationString: utils.MetaUnlimited,
			Weight:           20,
			Balance: &BalanceFilter{
				Type: utils.StringPointer(utils.MetaVoice),
				Uuid: utils.StringPointer(utils.GenUUID()),
				Value: &utils.ValueFormula{Static: 200,
					Params: make(map[string]interface{})},
				Weight:         utils.Float64Pointer(20),
				RatingSubject:  utils.StringPointer("test"),
				DestinationIDs: utils.StringMapPointer(utils.NewStringMap("NAT")),
				Disabled:       utils.BoolPointer(false),
				Timings: []*RITiming{
					{
						Years:     utils.Years{2016, 2017},
						Months:    utils.Months{time.January, time.February, time.March},
						MonthDays: utils.MonthDays{1, 2, 3, 4},
						WeekDays:  utils.WeekDays{1, 2, 3},
						StartTime: utils.MetaASAP,
					},
				},
				Blocker: utils.BoolPointer(false),
			},
		},
	}
	if err := onStor.SetActions(acts[0].Id,
		acts, utils.NonTransactional); err != nil {
		t.Error(err)
	}

	//get from database
	if rcv, err := onStor.GetActions(acts[0].Id,
		true, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(acts[0], rcv[0]) {
		t.Errorf("Expecting: %v, received: %v", utils.ToJSON(acts[0]), utils.ToJSON(rcv[0]))
	}
	if onStor.dataDB.GetStorageType() != utils.INTERNAL {
		//get from cache
		if rcv, err := onStor.GetActions(acts[0].Id,
			false, utils.NonTransactional); err != nil {
			t.Error(err)
		} else if !reflect.DeepEqual(acts[0], rcv[0]) {
			t.Errorf("Expecting: %v, received: %v", utils.ToJSON(acts[0]), utils.ToJSON(rcv[0]))
		}
	}
	if err := onStor.RemoveActions(acts[0].Id, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	//check database if removed
	if _, rcvErr := onStor.GetActions(acts[0].Id,
		true, utils.NonTransactional); rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
	//check cache if removed
	if _, rcvErr := onStor.GetActions(acts[0].Id,
		false, utils.NonTransactional); rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
}

func testOnStorITSharedGroup(t *testing.T) {
	sg := &SharedGroup{
		Id: "SG2",
		AccountParameters: map[string]*SharingParameters{
			"*any": {
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
	if onStor.dataDB.GetStorageType() != utils.INTERNAL {
		//get from cache
		if rcv, err := onStor.GetSharedGroup(sg.Id, false,
			utils.NonTransactional); err != nil {
			t.Error(err)
		} else if !reflect.DeepEqual(sg, rcv) {
			t.Errorf("Expecting: %v, received: %v", sg, rcv)
		}
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
		"*any": {
			Strategy:      "*lowest",
			RatingSubject: "",
		},
		"*any2": {
			Strategy:      "*lowest2",
			RatingSubject: "",
		},
	}
	if err := onStor.SetSharedGroup(sg, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	if onStor.dataDB.GetStorageType() != utils.INTERNAL {
		//get from cache
		if rcv, err := onStor.GetSharedGroup(sg.Id, false,
			utils.NonTransactional); err != nil {
			t.Error(err)
		} else if !reflect.DeepEqual(sg, rcv) {
			t.Errorf("Expecting: %v, received: %v", sg, rcv)
		}
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
			ID: "testOnStorITCRUDActionTriggers",
			Balance: &BalanceFilter{Type: utils.StringPointer(utils.MetaMonetary),
				Timings: make([]*RITiming, 0)},
			ThresholdValue:    2,
			ThresholdType:     utils.TriggerMaxEventCounter,
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
			{
				Uuid: utils.GenUUID(),
				Timing: &RateInterval{
					Timing: &RITiming{
						Years:     utils.Years{2012},
						Months:    utils.Months{},
						MonthDays: utils.MonthDays{},
						WeekDays:  utils.WeekDays{},
						StartTime: utils.MetaASAP,
					},
				},
				Weight:    10,
				ActionsID: "MINI",
			},
			{
				Uuid: utils.GenUUID(),
				Timing: &RateInterval{
					Timing: &RITiming{
						Years:     utils.Years{2012},
						Months:    utils.Months{},
						MonthDays: utils.MonthDays{},
						WeekDays:  utils.WeekDays{},
						StartTime: utils.MetaASAP,
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
	// if err = onStor.DataDB().SelectDatabase("13"); err != nil {
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
	// if err = onStor.DataDB().SelectDatabase(onStorCfg); err != nil {
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
	// if err = onStor.DataDB().SelectDatabase("13"); err != nil {
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
	// if err = onStor.DataDB().SelectDatabase(onStorCfg); err != nil {
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
		BalanceMap: map[string]Balances{utils.MetaMonetary: {&Balance{Value: 10, Weight: 10}}},
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
	} else if !reflect.DeepEqual(acc.BalanceMap[utils.MetaMonetary][0].Value, rcv.BalanceMap[utils.MetaMonetary][0].Value) {
		t.Errorf("Expecting: %v, received: %v", acc.BalanceMap[utils.MetaMonetary][0].Value, rcv.BalanceMap[utils.MetaMonetary][0].Value)
	} else if !reflect.DeepEqual(acc.BalanceMap[utils.MetaMonetary][0].Weight, rcv.BalanceMap[utils.MetaMonetary][0].Weight) {
		t.Errorf("Expecting: %v, received: %v", acc.BalanceMap[utils.MetaMonetary][0].Weight, rcv.BalanceMap[utils.MetaMonetary][0].Weight)
	}
	if err := onStor.RemoveAccount(acc.ID); err != nil {
		t.Error(err)
	}
	if _, rcvErr := onStor.GetAccount(acc.ID); rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
}

func testOnStorITResourceProfile(t *testing.T) {
	rL := &ResourceProfile{
		Tenant:    "cgrates.org",
		ID:        "RL_TEST2",
		Weight:    10,
		FilterIDs: []string{"*string:~*req.Account:1001"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 3, 13, 43, 0, 0, time.UTC),
			ExpiryTime:     time.Date(2015, 7, 3, 13, 43, 0, 0, time.UTC)},
		Limit:        1,
		ThresholdIDs: []string{"TEST_ACTIONS"},
		UsageTTL:     3 * time.Nanosecond,
	}
	if _, rcvErr := onStor.GetResourceProfile(rL.Tenant, rL.ID,
		true, false, utils.NonTransactional); rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
	if err := onStor.SetResourceProfile(rL, false); err != nil {
		t.Error(err)
	}
	//get from database
	if rcv, err := onStor.GetResourceProfile(rL.Tenant, rL.ID,
		false, true, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rL, rcv) {
		t.Errorf("Expecting: %v, received: %v", utils.ToJSON(rL), utils.ToJSON(rcv))
	}
	expectedR := []string{"rsp_cgrates.org:RL_TEST2"}
	if itm, err := onStor.DataDB().GetKeysForPrefix(utils.ResourceProfilesPrefix); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedR, itm) {
		t.Errorf("Expected : %+v, but received %+v", expectedR, itm)
	}
	//update
	rL.ThresholdIDs = []string{"TH1", "TH2"}
	if err := onStor.SetResourceProfile(rL, false); err != nil {
		t.Error(err)
	}

	//get from database
	if rcv, err := onStor.GetResourceProfile(rL.Tenant, rL.ID,
		false, false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rL, rcv) {
		t.Errorf("Expecting: %v, received: %v", utils.ToJSON(rL), utils.ToJSON(rcv))
	}

	if err := onStor.RemoveResourceProfile(rL.Tenant, rL.ID,
		utils.NonTransactional, false); err != nil {
		t.Error(err)
	}
	//check database if removed
	if _, rcvErr := onStor.GetResourceProfile(rL.Tenant, rL.ID,
		false, false, utils.NonTransactional); rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
}

func testOnStorITResource(t *testing.T) {
	res := &Resource{
		Tenant: "cgrates.org",
		ID:     "RL1",
		Usages: map[string]*ResourceUsage{
			"RU1": {
				ID:         "RU1",
				ExpiryTime: time.Date(2014, 7, 3, 13, 43, 0, 0, time.UTC),
				Units:      2,
			},
		},
		TTLIdx: []string{"RU1"},
	}
	if _, rcvErr := onStor.GetResource(res.Tenant, res.ID,
		true, false, utils.NonTransactional); rcvErr != nil && rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
	if err := onStor.SetResource(res, nil, 0, true); err != nil {
		t.Error(err)
	}
	//get from database
	if rcv, err := onStor.GetResource("cgrates.org", "RL1",
		false, false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !(reflect.DeepEqual(res, rcv)) {
		t.Errorf("Expecting: %v, received: %v", utils.ToJSON(res), utils.ToJSON(rcv))
	}
	expectedT := []string{"res_cgrates.org:RL1"}
	if itm, err := onStor.DataDB().GetKeysForPrefix(utils.ResourcesPrefix); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedT, itm) {
		t.Errorf("Expected : %+v, but received %+v", expectedT, itm)
	}
	//update
	res.TTLIdx = []string{"RU1", "RU2"}
	if err := onStor.SetResource(res, nil, 0, true); err != nil {
		t.Error(err)
	}

	//get from database
	if rcv, err := onStor.GetResource("cgrates.org", "RL1",
		false, false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !(reflect.DeepEqual(res, rcv)) {
		t.Errorf("Expecting: %v, received: %v", utils.ToJSON(res), utils.ToJSON(rcv))
	}

	if err := onStor.RemoveResource(res.Tenant, res.ID, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	//check database if removed
	if _, rcvErr := onStor.GetResource(res.Tenant, res.ID,
		false, false, utils.NonTransactional); rcvErr != utils.ErrNotFound {
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
	if onStor.dataDB.GetStorageType() != utils.INTERNAL {
		//get from cache
		if rcv, err := onStor.GetTiming(tmg.ID, false, utils.NonTransactional); err != nil {
			t.Error(err)
		} else if !reflect.DeepEqual(tmg, rcv) {
			t.Errorf("Expecting: %v, received: %v", tmg, rcv)
		}
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
	ist := &utils.LoadInstance{
		LoadID:           "Load",
		RatingLoadID:     "RatingLoad",
		AccountingLoadID: "Account",
		LoadTime:         time,
	}
	if err := onStor.DataDB().AddLoadHistory(ist, 1, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	if rcv, err := onStor.DataDB().GetLoadHistory(1, true, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(ist, rcv[0]) {
		t.Errorf("Expecting: %v, received: %v", utils.ToJSON(ist), utils.ToJSON(rcv[0]))
	}
}

func testOnStorITCRUDStructVersion(t *testing.T) {
	if _, err := onStor.DataDB().GetVersions(utils.Accounts); err != utils.ErrNotFound {
		t.Error(err)
	}
	vrs := Versions{
		utils.Accounts:       3,
		utils.Actions:        2,
		utils.ActionTriggers: 2,
		utils.ActionPlans:    2,
		utils.SharedGroups:   2,
		utils.CostDetails:    1,
	}
	if err := onStor.DataDB().SetVersions(vrs, false); err != nil {
		t.Error(err)
	}
	if rcv, err := onStor.DataDB().GetVersions(""); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(vrs, rcv) {
		t.Errorf("Expecting: %v, received: %v", vrs, rcv)
	}
	delete(vrs, utils.SharedGroups)
	if err := onStor.DataDB().SetVersions(vrs, true); err != nil { // overwrite
		t.Error(err)
	}
	if rcv, err := onStor.DataDB().GetVersions(""); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(vrs, rcv) {
		t.Errorf("Expecting: %v, received: %v", vrs, rcv)
	}
	eAcnts := Versions{utils.Accounts: vrs[utils.Accounts]}
	if rcv, err := onStor.DataDB().GetVersions(utils.Accounts); err != nil { //query one element
		t.Error(err)
	} else if !reflect.DeepEqual(eAcnts, rcv) {
		t.Errorf("Expecting: %v, received: %v", eAcnts, rcv)
	}
	if _, err := onStor.DataDB().GetVersions(utils.NotAvailable); err != utils.ErrNotFound { //query non-existent
		t.Error(err)
	}
	eAcnts[utils.Accounts] = 2
	vrs[utils.Accounts] = eAcnts[utils.Accounts]
	if err := onStor.DataDB().SetVersions(eAcnts, false); err != nil { // change one element
		t.Error(err)
	}
	if rcv, err := onStor.DataDB().GetVersions(""); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(vrs, rcv) {
		t.Errorf("Expecting: %v, received: %v", vrs, rcv)
	}
	if err = onStor.DataDB().RemoveVersions(eAcnts); err != nil { // remove one element
		t.Error(err)
	}
	delete(vrs, utils.Accounts)
	if rcv, err := onStor.DataDB().GetVersions(""); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(vrs, rcv) {
		t.Errorf("Expecting: %v, received: %v", vrs, rcv)
	}
	if err = onStor.DataDB().RemoveVersions(nil); err != nil { // remove one element
		t.Error(err)
	}
	if _, err := onStor.DataDB().GetVersions(""); err != utils.ErrNotFound { //query non-existent
		t.Error(err)
	}
}

func testOnStorITStatQueueProfile(t *testing.T) {
	sq := &StatQueueProfile{
		Tenant:             "cgrates.org",
		ID:                 "test",
		ActivationInterval: &utils.ActivationInterval{},
		FilterIDs:          []string{"*string:~*req.Account:1001"},
		QueueLength:        2,
		TTL:                0,
		Stored:             true,
		ThresholdIDs:       []string{"Thresh1"},
	}
	if _, rcvErr := onStor.GetStatQueueProfile(sq.Tenant, sq.ID,
		true, false, utils.NonTransactional); rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
	if err := onStor.SetStatQueueProfile(sq, false); err != nil {
		t.Error(err)
	}
	//get from database
	if rcv, err := onStor.GetStatQueueProfile(sq.Tenant,
		sq.ID, false, false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(sq, rcv) {
		t.Errorf("Expecting: %v, received: %v", utils.ToJSON(sq), utils.ToJSON(rcv))
	}
	expectedR := []string{"sqp_cgrates.org:test"}
	if itm, err := onStor.DataDB().GetKeysForPrefix(utils.StatQueueProfilePrefix); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedR, itm) {
		t.Errorf("Expected : %+v, but received %+v", expectedR, itm)
	}
	//update
	sq.ThresholdIDs = []string{"TH1", "TH2"}
	if err := onStor.SetStatQueueProfile(sq, false); err != nil {
		t.Error(err)
	}

	//get from database
	if rcv, err := onStor.GetStatQueueProfile(sq.Tenant,
		sq.ID, false, false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(sq, rcv) {
		t.Errorf("Expecting: %v, received: %v", utils.ToJSON(sq), utils.ToJSON(rcv))
	}
	if err := onStor.RemoveStatQueueProfile(sq.Tenant, sq.ID,
		utils.NonTransactional, false); err != nil {
		t.Error(err)
	}
	//check database if removed
	if _, rcvErr := onStor.GetStatQueueProfile(sq.Tenant,
		sq.ID, true, false, utils.NonTransactional); rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
}

func testOnStorITStatQueue(t *testing.T) {
	eTime := utils.TimePointer(time.Date(2013, 10, 1, 0, 0, 0, 0, time.UTC))
	sq := &StatQueue{
		Tenant: "cgrates.org",
		ID:     "Test_StatQueue",
		SQItems: []SQItem{
			{EventID: "cgrates.org:ev1", ExpiryTime: eTime},
			{EventID: "cgrates.org:ev2", ExpiryTime: eTime},
			{EventID: "cgrates.org:ev3", ExpiryTime: eTime},
		},
		SQMetrics: map[string]StatMetric{
			utils.MetaASR: &StatASR{
				Answered: 2,
				Count:    3,
				Events: map[string]*StatWithCompress{
					"cgrates.org:ev1": {Stat: 1},
					"cgrates.org:ev2": {Stat: 1},
					"cgrates.org:ev3": {Stat: 0},
				},
			},
		},
	}
	if _, rcvErr := onStor.GetStatQueue(sq.Tenant, sq.ID,
		true, false, utils.NonTransactional); rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
	if err := onStor.SetStatQueue(sq, nil, 0, nil, 0, true); err != nil {
		t.Error(err)
	}
	//get from database
	if rcv, err := onStor.GetStatQueue(sq.Tenant,
		sq.ID, false, false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(sq, rcv) {
		t.Errorf("Expecting: %v, received: %v", utils.ToJSON(sq), utils.ToJSON(rcv))
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
			Events: map[string]*StatWithCompress{
				"cgrates.org:ev1": {Stat: 1},
				"cgrates.org:ev2": {Stat: 1},
				"cgrates.org:ev3": {Stat: 1},
			},
		},
	}
	if err := onStor.SetStatQueue(sq, nil, 0, nil, 0, true); err != nil {
		t.Error(err)
	}

	//get from database
	if rcv, err := onStor.GetStatQueue(sq.Tenant,
		sq.ID, false, false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(sq, rcv) {
		t.Errorf("Expecting: %v, received: %v", utils.ToJSON(sq), utils.ToJSON(rcv))
	}
	if err := onStor.RemoveStatQueue(sq.Tenant, sq.ID,
		utils.NonTransactional); err != nil {
		t.Error(err)
	}
	//check database if removed
	if _, rcvErr := onStor.GetStatQueue(sq.Tenant,
		sq.ID, false, false, utils.NonTransactional); rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
}

func testOnStorITThresholdProfile(t *testing.T) {
	fp := &Filter{
		Tenant: "cgrates.org",
		ID:     "TestFilter2",
		Rules: []*FilterRule{
			{
				Element: "Account",
				Type:    utils.MetaString,
				Values:  []string{"1001", "1002"},
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
		MaxHits:            12,
		MinSleep:           0,
		Blocker:            true,
		Weight:             1.4,
		ActionIDs:          []string{"Action1"},
	}
	if err := onStor.SetFilter(fp, true); err != nil {
		t.Error(err)
	}
	if _, rcvErr := onStor.GetThresholdProfile(th.Tenant, th.ID,
		true, false, utils.NonTransactional); rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
	if err := onStor.SetThresholdProfile(th, true); err != nil {
		t.Error(err)
	}
	//get from database
	if rcv, err := onStor.GetThresholdProfile(th.Tenant, th.ID,
		false, false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(th, rcv) {
		t.Errorf("Expecting: %v, received: %v", utils.ToJSON(th), utils.ToJSON(rcv))
	}
	expectedR := []string{"thp_cgrates.org:test"}
	if itm, err := onStor.DataDB().GetKeysForPrefix(utils.ThresholdProfilePrefix); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedR, itm) {
		t.Errorf("Expected : %+v, but received %+v", expectedR, itm)
	}
	//update
	th.ActionIDs = []string{"Action1", "Action2"}
	if err := onStor.SetThresholdProfile(th, true); err != nil {
		t.Error(err)
	}

	//get from database
	if rcv, err := onStor.GetThresholdProfile(th.Tenant, th.ID,
		false, false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(th, rcv) {
		t.Errorf("Expecting: %v, received: %v", utils.ToJSON(th), utils.ToJSON(rcv))
	}
	if err := onStor.RemoveThresholdProfile(th.Tenant,
		th.ID, utils.NonTransactional, false); err != nil {
		t.Error(err)
	}
	//check database if removed
	if _, rcvErr := onStor.GetThresholdProfile(th.Tenant,
		th.ID, false, false, utils.NonTransactional); rcvErr != utils.ErrNotFound {
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
		true, false, utils.NonTransactional); rcvErr != nil && rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
	if err := onStor.SetThreshold(th, 0, true); err != nil {
		t.Error(err)
	}
	//get from database
	if rcv, err := onStor.GetThreshold("cgrates.org", "TH1",
		false, false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !(reflect.DeepEqual(th, rcv)) {
		t.Errorf("Expecting: %v, received: %v", utils.ToJSON(th), utils.ToJSON(rcv))
	}
	expectedT := []string{"thd_cgrates.org:TH1"}
	if itm, err := onStor.DataDB().GetKeysForPrefix(utils.ThresholdPrefix); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedT, itm) {
		t.Errorf("Expected : %+v, but received %+v", expectedT, itm)
	}
	//update
	th.Hits = 20
	if err := onStor.SetThreshold(th, 0, true); err != nil {
		t.Error(err)
	}

	//get from database
	if rcv, err := onStor.GetThreshold("cgrates.org", "TH1",
		false, false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !(reflect.DeepEqual(th, rcv)) {
		t.Errorf("Expecting: %v, received: %v", utils.ToJSON(th), utils.ToJSON(rcv))
	}
	if err := onStor.RemoveThreshold(th.Tenant, th.ID, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	//check database if removed
	if _, rcvErr := onStor.GetThreshold(th.Tenant, th.ID,
		false, false, utils.NonTransactional); rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
}

func testOnStorITFilter(t *testing.T) {
	fp := &Filter{
		Tenant: "cgrates.org",
		ID:     "Filter1",
		Rules: []*FilterRule{
			{
				Element: "Account",
				Type:    utils.MetaString,
				Values:  []string{"1001", "1002"},
			},
		},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			ExpiryTime:     time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
	}
	if err := fp.Compile(); err != nil {
		t.Fatal(err)
	}
	if _, rcvErr := onStor.GetFilter("cgrates.org", "Filter1",
		true, false, utils.NonTransactional); rcvErr != nil && rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
	if err := onStor.SetFilter(fp, true); err != nil {
		t.Error(err)
	}
	//get from cache
	if rcv, err := onStor.GetFilter("cgrates.org", "Filter1",
		true, false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !(reflect.DeepEqual(fp, rcv)) {
		t.Errorf("Expecting: %v, received: %v", fp, rcv)
	}
	//get from database
	if rcv, err := onStor.GetFilter("cgrates.org", "Filter1",
		false, false, utils.NonTransactional); err != nil {
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
	fp.Rules = []*FilterRule{
		{
			Element: "Account",
			Type:    utils.MetaString,
			Values:  []string{"1001", "1002"},
		},
		{
			Element: "Destination",
			Type:    utils.MetaString,
			Values:  []string{"10", "20"},
		},
	}
	if err := fp.Compile(); err != nil {
		t.Fatal(err)
	}
	if err := onStor.SetFilter(fp, true); err != nil {
		t.Error(err)
	}

	//get from cache
	if rcv, err := onStor.GetFilter("cgrates.org", "Filter1",
		true, false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !(reflect.DeepEqual(fp, rcv)) {
		t.Errorf("Expecting: %v, received: %v", fp, rcv)
	}
	//get from database
	if rcv, err := onStor.GetFilter("cgrates.org", "Filter1",
		false, false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !(reflect.DeepEqual(fp, rcv)) {
		t.Errorf("Expecting: %v, received: %v", fp, rcv)
	}
	if err := onStor.RemoveFilter(fp.Tenant, fp.ID, utils.NonTransactional, true); err != nil {
		t.Error(err)
	}
	//check cache if removed
	if _, rcvErr := onStor.GetFilter("cgrates.org", "Filter1",
		true, false, utils.NonTransactional); rcvErr != nil && rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
	//check database if removed
	if _, rcvErr := onStor.GetFilter("cgrates.org", "Filter1",
		false, false, utils.NonTransactional); rcvErr != nil && rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
}

func testOnStorITRouteProfile(t *testing.T) {
	splProfile := &RouteProfile{
		Tenant:    "cgrates.org",
		ID:        "SPRF_1",
		FilterIDs: []string{"*string:~*reg.Accout:1002", "*string:~*reg.Destination:11"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
		Sorting:           "*lowest_cost",
		SortingParameters: []string{},
		Routes: []*Route{
			{
				ID:              "supplier1",
				FilterIDs:       []string{"FLTR_DST_DE"},
				AccountIDs:      []string{"Account1"},
				RatingPlanIDs:   []string{"RPL_1"},
				ResourceIDs:     []string{"ResGR1"},
				StatIDs:         []string{"Stat1"},
				Weight:          10,
				RouteParameters: "param1",
			},
		},
		Weight: 20,
	}
	if _, rcvErr := onStor.GetRouteProfile("cgrates.org", "SPRF_1",
		true, false, utils.NonTransactional); rcvErr != nil && rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
	if err := onStor.SetRouteProfile(splProfile, false); err != nil {
		t.Error(err)
	}
	//get from cache
	if rcv, err := onStor.GetRouteProfile("cgrates.org", "SPRF_1",
		true, false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !(reflect.DeepEqual(splProfile, rcv)) {
		t.Errorf("Expecting: %v, received: %v", splProfile, rcv)
	}
	//get from database
	if rcv, err := onStor.GetRouteProfile("cgrates.org", "SPRF_1",
		false, false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !(reflect.DeepEqual(splProfile, rcv)) {
		t.Errorf("Expecting: %v, received: %v", splProfile, rcv)
	}
	expectedT := []string{"rpp_cgrates.org:SPRF_1"}
	if itm, err := onStor.DataDB().GetKeysForPrefix(utils.RouteProfilePrefix); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedT, itm) {
		t.Errorf("Expected : %+v, but received %+v", expectedT, itm)
	}
	//update
	splProfile.Routes = []*Route{
		{
			ID:              "supplier1",
			FilterIDs:       []string{"FLTR_DST_DE"},
			AccountIDs:      []string{"Account1"},
			RatingPlanIDs:   []string{"RPL_1"},
			ResourceIDs:     []string{"ResGR1"},
			StatIDs:         []string{"Stat1"},
			Weight:          10,
			RouteParameters: "param1",
		},
		{
			ID:              "supplier2",
			FilterIDs:       []string{"FLTR_DST_DE"},
			AccountIDs:      []string{"Account2"},
			RatingPlanIDs:   []string{"RPL_2"},
			ResourceIDs:     []string{"ResGR2"},
			StatIDs:         []string{"Stat2"},
			Weight:          20,
			RouteParameters: "param2",
		},
	}
	if err := onStor.SetRouteProfile(splProfile, false); err != nil {
		t.Error(err)
	}

	//get from cache
	if rcv, err := onStor.GetRouteProfile("cgrates.org", "SPRF_1",
		true, false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !(reflect.DeepEqual(splProfile, rcv)) {
		t.Errorf("Expecting: %v, received: %v", splProfile, rcv)
	}
	//get from database
	if rcv, err := onStor.GetRouteProfile("cgrates.org", "SPRF_1",
		false, false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !(reflect.DeepEqual(splProfile, rcv)) {
		t.Errorf("Expecting: %v, received: %v", splProfile, rcv)
	}
	if err := onStor.RemoveRouteProfile(splProfile.Tenant, splProfile.ID,
		utils.NonTransactional, false); err != nil {
		t.Error(err)
	}
	//check cache if removed
	if _, rcvErr := onStor.GetRouteProfile("cgrates.org", "SPRF_1",
		true, false, utils.NonTransactional); rcvErr != nil && rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
	//check database if removed
	if _, rcvErr := onStor.GetRouteProfile("cgrates.org", "SPRF_1",
		false, false, utils.NonTransactional); rcvErr != nil && rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
}

func testOnStorITAttributeProfile(t *testing.T) {
	attrProfile := &AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "AttrPrf1",
		FilterIDs: []string{"*string:~*reg.Accout:1002", "*string:~*reg.Destination:11"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
		Contexts: []string{"con1"},
		Attributes: []*Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "FN1",
				Value: config.NewRSRParsersMustCompile("Al1", utils.InfieldSep),
			},
		},
		Weight: 20,
	}
	if _, rcvErr := onStor.GetAttributeProfile("cgrates.org", "AttrPrf1",
		true, false, utils.NonTransactional); rcvErr != nil && rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
	if err := onStor.SetAttributeProfile(attrProfile, false); err != nil {
		t.Error(err)
	}
	//get from cache
	if rcv, err := onStor.GetAttributeProfile("cgrates.org", "AttrPrf1",
		true, false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !(reflect.DeepEqual(attrProfile, rcv)) {
		t.Errorf("Expecting: %v, received: %v", attrProfile, rcv)
	}
	//get from database
	if rcv, err := onStor.GetAttributeProfile("cgrates.org", "AttrPrf1",
		false, false, utils.NonTransactional); err != nil {
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
	attrProfile.Contexts = []string{"con1", "con2", "con3"}
	if err := onStor.SetAttributeProfile(attrProfile, false); err != nil {
		t.Error(err)
	}

	//get from cache
	if rcv, err := onStor.GetAttributeProfile("cgrates.org", "AttrPrf1",
		true, false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !(reflect.DeepEqual(attrProfile, rcv)) {
		t.Errorf("Expecting: %v, received: %v", attrProfile, rcv)
	}
	//get from database
	if rcv, err := onStor.GetAttributeProfile("cgrates.org", "AttrPrf1",
		false, false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !(reflect.DeepEqual(attrProfile, rcv)) {
		t.Errorf("Expecting: %v, received: %v", attrProfile, rcv)
	}
	if err := onStor.RemoveAttributeProfile(attrProfile.Tenant,
		attrProfile.ID, utils.NonTransactional, false); err != nil {
		t.Error(err)
	}
	//check cache if removed
	if _, rcvErr := onStor.GetAttributeProfile("cgrates.org", "AttrPrf1",
		true, false, utils.NonTransactional); rcvErr != nil && rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
	//check database if removed
	if _, rcvErr := onStor.GetAttributeProfile("cgrates.org", "AttrPrf1",
		false, true, utils.NonTransactional); rcvErr != nil && rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
}

func testOnStorITTestAttributeSubstituteIface(t *testing.T) {
	attrProfile := &AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "AttrPrf1",
		FilterIDs: []string{"*string:~*reg.Accout:1002", "*string:~*reg.Destination:11"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
		Contexts: []string{"con1"},
		Attributes: []*Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "FN1",
				Value: config.NewRSRParsersMustCompile("Val1", utils.InfieldSep),
			},
		},
		Weight: 20,
	}
	if _, rcvErr := onStor.GetAttributeProfile("cgrates.org", "AttrPrf1",
		true, false, utils.NonTransactional); rcvErr != nil && rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
	if err := onStor.SetAttributeProfile(attrProfile, false); err != nil {
		t.Error(err)
	}
	//check database
	if rcv, err := onStor.GetAttributeProfile("cgrates.org", "AttrPrf1",
		false, false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !(reflect.DeepEqual(attrProfile, rcv)) {
		t.Errorf("Expecting: %v, received: %v", attrProfile, rcv)
	}
	attrProfile.Attributes = []*Attribute{
		{
			Path:  utils.MetaReq + utils.NestingSep + "FN1",
			Value: config.NewRSRParsersMustCompile("123.123", utils.InfieldSep),
		},
	}
	if err := onStor.SetAttributeProfile(attrProfile, false); err != nil {
		t.Error(err)
	}
	//check database
	if rcv, err := onStor.GetAttributeProfile("cgrates.org", "AttrPrf1",
		false, false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !(reflect.DeepEqual(attrProfile, rcv)) {
		t.Errorf("Expecting: %v, received: %v", utils.ToJSON(attrProfile), utils.ToJSON(rcv))
	}
	attrProfile.Attributes = []*Attribute{
		{
			Path:  utils.MetaReq + utils.NestingSep + "FN1",
			Value: config.NewRSRParsersMustCompile("true", utils.InfieldSep),
		},
	}
	if err := onStor.SetAttributeProfile(attrProfile, false); err != nil {
		t.Error(err)
	}
	//check database
	if rcv, err := onStor.GetAttributeProfile("cgrates.org", "AttrPrf1",
		false, false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !(reflect.DeepEqual(attrProfile, rcv)) {
		t.Errorf("Expecting: %v, received: %v", utils.ToJSON(attrProfile), utils.ToJSON(rcv))
	}
}

func testOnStorITChargerProfile(t *testing.T) {
	cpp := &ChargerProfile{
		Tenant:    "cgrates.org",
		ID:        "CPP_1",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
		RunID:        "*rated",
		AttributeIDs: []string{"ATTR_1"},
		Weight:       20,
	}
	if _, rcvErr := onStor.GetChargerProfile("cgrates.org", "CPP_1",
		true, false, utils.NonTransactional); rcvErr != nil && rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
	if err := onStor.SetChargerProfile(cpp, false); err != nil {
		t.Error(err)
	}
	//get from database
	if rcv, err := onStor.GetChargerProfile("cgrates.org", "CPP_1",
		false, false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !(reflect.DeepEqual(cpp, rcv)) {
		t.Errorf("Expecting: %v, received: %v", cpp, rcv)
	}
	expectedT := []string{"cpp_cgrates.org:CPP_1"}
	if itm, err := onStor.DataDB().GetKeysForPrefix(utils.ChargerProfilePrefix); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedT, itm) {
		t.Errorf("Expected : %+v, but received %+v", expectedT, itm)
	}
	//update
	cpp.FilterIDs = []string{"*string:~*req.Accout:1001", "*prefix:~*req.Destination:10"}
	if err := onStor.SetChargerProfile(cpp, false); err != nil {
		t.Error(err)
	}

	//get from database
	if rcv, err := onStor.GetChargerProfile("cgrates.org", "CPP_1",
		false, false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !(reflect.DeepEqual(cpp, rcv)) {
		t.Errorf("Expecting: %v, received: %v", cpp, rcv)
	}
	if err := onStor.RemoveChargerProfile(cpp.Tenant, cpp.ID,
		utils.NonTransactional, false); err != nil {
		t.Error(err)
	}
	//check database if removed
	if _, rcvErr := onStor.GetChargerProfile("cgrates.org", "CPP_1",
		false, false, utils.NonTransactional); rcvErr != nil && rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
}

func testOnStorITDispatcherProfile(t *testing.T) {
	dpp := &DispatcherProfile{
		Tenant:    "cgrates.org",
		ID:        "Dsp1",
		FilterIDs: []string{"*string:Account:1001"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
		Strategy: utils.MetaFirst,
		// Hosts:    []string{"192.168.56.203"},
		Weight: 20,
	}
	if _, rcvErr := onStor.GetDispatcherProfile("cgrates.org", "Dsp1",
		true, false, utils.NonTransactional); rcvErr != nil && rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
	if err := onStor.SetDispatcherProfile(dpp, false); err != nil {
		t.Error(err)
	}
	//get from database
	if rcv, err := onStor.GetDispatcherProfile("cgrates.org", "Dsp1",
		false, false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !(reflect.DeepEqual(dpp, rcv)) {
		t.Errorf("Expecting: %v, received: %v", dpp, rcv)
	}
	expectedT := []string{"dpp_cgrates.org:Dsp1"}
	if itm, err := onStor.DataDB().GetKeysForPrefix(utils.DispatcherProfilePrefix); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedT, itm) {
		t.Errorf("Expected : %+v, but received %+v", expectedT, itm)
	}
	//update
	dpp.FilterIDs = []string{"*string:~*req.Accout:1001", "*prefix:~*req.Destination:10"}
	if err := onStor.SetDispatcherProfile(dpp, false); err != nil {
		t.Error(err)
	}

	//get from database
	if rcv, err := onStor.GetDispatcherProfile("cgrates.org", "Dsp1",
		false, false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !(reflect.DeepEqual(dpp, rcv)) {
		t.Errorf("Expecting: %v, received: %v", dpp, rcv)
	}
	if err := onStor.RemoveDispatcherProfile(dpp.Tenant, dpp.ID,
		utils.NonTransactional, false); err != nil {
		t.Error(err)
	}
	//check database if removed
	if _, rcvErr := onStor.GetDispatcherProfile("cgrates.org", "Dsp1",
		false, false, utils.NonTransactional); rcvErr != nil && rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
}

func testOnStorITRateProfile(t *testing.T) {
	rPrf := &RateProfile{
		Tenant:    "cgrates.org",
		ID:        "RP1",
		FilterIDs: []string{"*string:~*req.Subject:1001", "*string:~*req.Subject:1002"},
		Weights: utils.DynamicWeights{
			{
				Weight: 0,
			},
		},
		MaxCostStrategy: "*free",
		Rates: map[string]*Rate{
			"FIRST_GI": {
				ID:        "FIRST_GI",
				FilterIDs: []string{"*gi:~*req.Usage:0"},
				Weights: utils.DynamicWeights{
					{
						Weight: 0,
					},
				},
				Blocker: false,
			},
			"SECOND_GI": {
				ID:        "SECOND_GI",
				FilterIDs: []string{"*gi:~*req.Usage:1m"},
				Weights: utils.DynamicWeights{
					{
						Weight: 10,
					},
				},
				Blocker: false,
			},
		},
	}
	if _, rcvErr := onStor.GetRateProfile("cgrates.org", "RP1",
		true, false, utils.NonTransactional); rcvErr != nil && rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
	if err := onStor.SetRateProfile(rPrf, false); err != nil {
		t.Error(err)
	}
	if err = rPrf.Compile(); err != nil {
		t.Fatal(err)
	}
	//get from database
	if rcv, err := onStor.GetRateProfile("cgrates.org", "RP1",
		false, false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rPrf, rcv) {
		t.Errorf("Expecting: %v, received: %v", rPrf, rcv)
	}
	expectedT := []string{"rtp_cgrates.org:RP1"}
	if itm, err := onStor.DataDB().GetKeysForPrefix(utils.RateProfilePrefix); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedT, itm) {
		t.Errorf("Expected : %+v, but received %+v", expectedT, itm)
	}
	//update
	rPrf.FilterIDs = []string{"*string:~*req.Accout:1001", "*prefix:~*req.Destination:10"}
	if err := onStor.SetRateProfile(rPrf, false); err != nil {
		t.Error(err)
	}

	//get from database
	if rcv, err := onStor.GetRateProfile("cgrates.org", "RP1",
		false, false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !(reflect.DeepEqual(rPrf, rcv)) {
		t.Errorf("Expecting: %v, received: %v", rPrf, rcv)
	}
	if err := onStor.RemoveRateProfile(rPrf.Tenant, rPrf.ID,
		utils.NonTransactional, false); err != nil {
		t.Error(err)
	}
	//check database if removed
	if _, rcvErr := onStor.GetRateProfile("cgrates.org", "RP1",
		false, false, utils.NonTransactional); rcvErr != nil && rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
}

func testOnStorITActionProfile(t *testing.T) {
	actPrf := &ActionProfile{
		Tenant:    "cgrates.org",
		ID:        "TEST_ID1",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		Weight:    20,
		Schedule:  utils.MetaASAP,
		Targets: map[string]utils.StringSet{
			utils.MetaAccounts: utils.NewStringSet([]string{"acc1", "acc2", "acc3"}),
		},
		Actions: []*APAction{
			{
				ID:        "TOPUP",
				FilterIDs: []string{},
				Type:      "*topup",
				Diktats: []*APDiktat{{
					Path: "~*balance.TestBalance.Value",
				}},
			},
			{
				ID:        "TOPUP_TEST_VOICE",
				FilterIDs: []string{},
				Type:      "*topup",
				Diktats: []*APDiktat{{
					Path: "~*balance.TestVoiceBalance.Value",
				}},
			},
		},
	}

	//empty in database
	if _, err := onStor.GetActionProfile("cgrates.org", "TEST_ID1",
		true, false, utils.NonTransactional); err != utils.ErrNotFound {
		t.Error(err)
	}

	//get from database
	if err := onStor.SetActionProfile(actPrf, false); err != nil {
		t.Error(err)
	}
	if rcv, err := onStor.GetActionProfile("cgrates.org", "TEST_ID1",
		true, false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, actPrf) {
		t.Errorf("Expecting: %v, received: %v", actPrf, rcv)
	}

	//craft akeysFromPrefix
	expectedKey := []string{"acp_cgrates.org:TEST_ID1"}
	if rcv, err := onStor.DataDB().GetKeysForPrefix(utils.ActionProfilePrefix); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedKey, rcv) {
		t.Errorf("Expecting: %v, received: %v", expectedKey, rcv)
	}

	//updateFilters
	actPrf.FilterIDs = []string{"*prefix:~*req.Destination:10"}
	if err := onStor.SetActionProfile(actPrf, false); err != nil {
		t.Error(err)
	} else if rcv, err := onStor.GetActionProfile("cgrates.org", "TEST_ID1",
		false, false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(actPrf, rcv) {
		t.Errorf("Expecting: %v, received: %v", actPrf, rcv)
	}

	//remove from database
	if err := onStor.RemoveActionProfile("cgrates.org", "TEST_ID1",
		utils.NonTransactional, false); err != nil {
		t.Error(err)
	} else if _, err := onStor.GetActionProfile("cgrates.org", "TEST_ID1",
		false, false, utils.NonTransactional); err != utils.ErrNotFound {
		t.Error(err)
	}
}

func testOnStorITAccountProfile(t *testing.T) {
	acctPrf := &utils.AccountProfile{
		Tenant:    "cgrates.org",
		ID:        "RP1",
		FilterIDs: []string{"test_filterId"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			ExpiryTime:     time.Date(2014, 7, 15, 14, 25, 0, 0, time.UTC),
		},
		Weights: utils.DynamicWeights{
			{
				Weight: 2,
			},
		},
		Balances: map[string]*utils.Balance{
			"VoiceBalance": {
				ID:        "VoiceBalance",
				FilterIDs: []string{"FLTR_RES_GR2"},
				Weights: utils.DynamicWeights{
					{
						Weight: 10,
					},
				},
				Type: utils.MetaVoice,
				Units: &utils.Decimal{
					Big: new(decimal.Big).SetUint64(10),
				},
				Opts: map[string]interface{}{
					"key1": "val1",
				},
			}},
		ThresholdIDs: []string{"test_thrs"},
	}

	//empty in database
	if _, err := onStor.GetAccountProfile("cgrates.org", "RP1"); err != utils.ErrNotFound {
		t.Error(err)
	}

	//get from database
	if err := onStor.SetAccountProfile(acctPrf, false); err != nil {
		t.Error(err)
	}
	if rcv, err := onStor.GetAccountProfile("cgrates.org", "RP1"); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, acctPrf) {
		t.Errorf("Expecting: %v, received: %v", acctPrf, rcv)
	}

	//craft akeysFromPrefix
	expectedKey := []string{"anp_cgrates.org:RP1"}
	if rcv, err := onStor.DataDB().GetKeysForPrefix(utils.AccountProfilePrefix); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedKey, rcv) {
		t.Errorf("Expecting: %v, received: %v", expectedKey, rcv)
	}

	//updateFilters
	acctPrf.FilterIDs = []string{"*prefix:~*req.Destination:10"}
	if err := onStor.SetAccountProfile(acctPrf, false); err != nil {
		t.Error(err)
	} else if rcv, err := onStor.GetAccountProfile("cgrates.org", "RP1"); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(acctPrf, rcv) {
		t.Errorf("Expecting: %v, received: %v", acctPrf, rcv)
	}

	//remove from database
	if err := onStor.RemoveAccountProfile("cgrates.org", "RP1",
		utils.NonTransactional, false); err != nil {
		t.Error(err)
	} else if _, err := onStor.GetAccountProfile("cgrates.org", "RP1"); err != utils.ErrNotFound {
		t.Error(err)
	}
}
