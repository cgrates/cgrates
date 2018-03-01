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

	"github.com/cgrates/cgrates/cache"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

var (
	redisDB     *RedisStorage
	mongoDB     *MongoStorage
	dataManager *DataManager
	cfgDBName   string
)

// subtests to be executed for each confDIR
var sTests = []func(t *testing.T){
	testITFlush,
	testITIsDBEmpty,
	testITSetFilterIndexes,
	testITGetFilterIndexes,
	testITMatchFilterIndex,
	testITFlush,
	testITIsDBEmpty,
	testITTestThresholdFilterIndexes,
	testITTestAttributeProfileFilterIndexes,
	testITTestThresholdInlineFilterIndexing,
	testITFlush,
	testITIsDBEmpty,
	testITTestStoreFilterIndexesWithTransID,
	testITTestStoreFilterIndexesWithTransID2,
	testITFlush,
	testITIsDBEmpty,
	testITTestIndexingWithEmptyFltrID,
	testITTestIndexingWithEmptyFltrID2,
}

func TestITRedisConnect(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	redisDB, err = NewRedisStorage(fmt.Sprintf("%s:%s", cfg.DataDbHost, cfg.DataDbPort), 4,
		cfg.DataDbPass, cfg.DBDataEncoding, utils.REDIS_MAX_CONNS, nil, 1)
	if err != nil {
		t.Fatal("Could not connect to Redis", err.Error())
	}
	cfgDBName = cfg.DataDbName
}

func TestITRedis(t *testing.T) {
	dataManager = NewDataManager(redisDB)
	for _, stest := range sTests {
		t.Run("TestITRedis", stest)
	}
}

func TestITMongoConnect(t *testing.T) {
	cdrsMongoCfgPath := path.Join(*dataDir, "conf", "samples", "cdrsv2mongo")
	mgoITCfg, err := config.NewCGRConfigFromFolder(cdrsMongoCfgPath)
	if err != nil {
		t.Fatal(err)
	}
	if mongoDB, err = NewMongoStorage(mgoITCfg.StorDBHost, mgoITCfg.StorDBPort,
		mgoITCfg.StorDBName, mgoITCfg.StorDBUser, mgoITCfg.StorDBPass,
		utils.StorDB, nil, mgoITCfg.CacheCfg(), mgoITCfg.LoadHistorySize); err != nil {
		t.Fatal(err)
	}
	cfgDBName = mgoITCfg.StorDBName
}

func TestITMongo(t *testing.T) {
	dataManager = NewDataManager(mongoDB)
	for _, stest := range sTests {
		t.Run("TestITMongo", stest)
	}
}

func testITFlush(t *testing.T) {
	if err := dataManager.DataDB().Flush(""); err != nil {
		t.Error(err)
	}
	cache.Flush()
}

func testITIsDBEmpty(t *testing.T) {
	test, err := dataManager.DataDB().IsDBEmpty()
	if err != nil {
		t.Error(err)
	} else if test != true {
		t.Errorf("\nExpecting: true got :%+v", test)
	}
}

func testITSetFilterIndexes(t *testing.T) {
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
	if err := dataManager.SetFilterIndexes(
		GetDBIndexKey(utils.ResourceProfilesPrefix, "cgrates.org", false),
		idxes, false, utils.NonTransactional); err != nil {
		t.Error(err)
	}
}

func testITGetFilterIndexes(t *testing.T) {
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
	if exsbjDan, err := dataManager.GetFilterIndexes(
		GetDBIndexKey(utils.ResourceProfilesPrefix, "cgrates.org", false), MetaString,
		sbjDan); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedsbjDan, exsbjDan) {
		t.Errorf("Expecting: %+v, received: %+v", expectedsbjDan, exsbjDan)
	}
	if rcv, err := dataManager.GetFilterIndexes(
		GetDBIndexKey(utils.ResourceProfilesPrefix, "cgrates.org", false), MetaString,
		nil); err != nil {
		t.Error(err)
	} else {
		if !reflect.DeepEqual(eIdxes, rcv) {
			t.Errorf("Expecting: %+v, received: %+v", eIdxes, rcv)
		}
	}
	if _, err := dataManager.GetFilterIndexes("unknown_key", MetaString, nil); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	}
	if err := dataManager.RemoveFilterIndexes(
		GetDBIndexKey(utils.ResourceProfilesPrefix, "cgrates.org", false)); err != nil {
		t.Error(err)
	}
	_, err := dataManager.GetFilterIndexes(
		GetDBIndexKey(utils.ResourceProfilesPrefix, "cgrates.org", false), MetaString, nil)
	if err != utils.ErrNotFound {
		t.Error(err)
	}
	if err := dataManager.SetFilterIndexes(
		GetDBIndexKey(utils.ResourceProfilesPrefix, "cgrates.org", false),
		eIdxes, false, utils.NonTransactional); err != nil {
		t.Error(err)
	}
}

func testITMatchFilterIndex(t *testing.T) {
	eMp := utils.StringMap{
		"RL1": true,
		"RL2": true,
	}
	if rcvMp, err := dataManager.MatchFilterIndex(
		GetDBIndexKey(utils.ResourceProfilesPrefix, "cgrates.org", false),
		MetaString, "Account", "1002"); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eMp, rcvMp) {
		t.Errorf("Expecting: %+v, received: %+v", eMp, rcvMp)
	}
	if _, err := dataManager.MatchFilterIndex(
		GetDBIndexKey(utils.ResourceProfilesPrefix, "cgrates.org", false),
		MetaString, "NonexistentField", "1002"); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	}
}

func testITTestThresholdFilterIndexes(t *testing.T) {
	fp := &Filter{
		Tenant: "cgrates.org",
		ID:     "Filter1",
		Rules: []*FilterRule{
			&FilterRule{
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
	if err := dataManager.SetFilter(fp); err != nil {
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

	if err := dataManager.SetThresholdProfile(th, true); err != nil {
		t.Error(err)
	}
	if err := dataManager.SetThresholdProfile(th2, true); err != nil {
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
	rfi := NewFilterIndexer(onStor, utils.ThresholdProfilePrefix, th.Tenant)
	if rcvIdx, err := dataManager.GetFilterIndexes(
		GetDBIndexKey(rfi.itemType, rfi.dbKeySuffix, false), MetaString,
		nil); err != nil {
		t.Error(err)
	} else {
		if !reflect.DeepEqual(eIdxes, rcvIdx) {
			t.Errorf("Expecting %+v, received: %+v", eIdxes, rcvIdx)
		}
	}
	if reverseRcvIdx, err := dataManager.GetFilterReverseIndexes(
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
		Rules: []*FilterRule{
			&FilterRule{
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
	if err := dataManager.SetFilter(fp2); err != nil {
		t.Error(err)
	}

	th.FilterIDs = []string{"Filter2"}
	if err := dataManager.SetThresholdProfile(th, true); err != nil {
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
	if rcvIdx, err := dataManager.GetFilterIndexes(
		GetDBIndexKey(rfi.itemType, rfi.dbKeySuffix, false), MetaString,
		nil); err != nil {
		t.Error(err)
	} else {
		if !reflect.DeepEqual(eIdxes, rcvIdx) {
			t.Errorf("Expecting %+v, received: %+v", eIdxes, rcvIdx)
		}
	}
	if reverseRcvIdx, err := dataManager.GetFilterReverseIndexes(
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
		Rules: []*FilterRule{
			&FilterRule{
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
	if err := dataManager.SetFilter(fp3); err != nil {
		t.Error(err)
	}

	th.FilterIDs = []string{"Filter1", "Filter3"}
	if err := dataManager.SetThresholdProfile(th, true); err != nil {
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
	if rcvIdx, err := dataManager.GetFilterIndexes(
		GetDBIndexKey(rfi.itemType, rfi.dbKeySuffix, false), MetaString,
		nil); err != nil {
		t.Error(err)
	} else {
		if !reflect.DeepEqual(eIdxes, rcvIdx) {
			t.Errorf("Expecting %+v, received: %+v", eIdxes, rcvIdx)
		}
	}
	if reverseRcvIdx, err := dataManager.GetFilterReverseIndexes(
		GetDBIndexKey(rfi.itemType, rfi.dbKeySuffix, true),
		nil); err != nil {
		t.Error(err)
	} else {
		if !reflect.DeepEqual(reverseIdxes, reverseRcvIdx) {
			t.Errorf("Expecting %+v, received: %+v", reverseIdxes, reverseRcvIdx)
		}
	}
	//remove thresholds
	if err := dataManager.RemoveThresholdProfile(th.Tenant,
		th.ID, utils.NonTransactional, true); err != nil {
		t.Error(err)
	}
	if err := dataManager.RemoveThresholdProfile(th2.Tenant,
		th2.ID, utils.NonTransactional, true); err != nil {
		t.Error(err)
	}
	if _, err := dataManager.GetFilterIndexes(
		GetDBIndexKey(rfi.itemType, rfi.dbKeySuffix, false), MetaString,
		nil); err != utils.ErrNotFound {
		t.Error(err)
	}
	if _, err := dataManager.GetFilterReverseIndexes(
		GetDBIndexKey(rfi.itemType, rfi.dbKeySuffix, true),
		nil); err != utils.ErrNotFound {
		t.Error(err)
	}
}

func testITTestAttributeProfileFilterIndexes(t *testing.T) {
	fp := &Filter{
		Tenant: "cgrates.org",
		ID:     "Filter1",
		Rules: []*FilterRule{
			&FilterRule{
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
	if err := dataManager.SetFilter(fp); err != nil {
		t.Error(err)
	}
	attrProfile := &AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "AttrPrf",
		FilterIDs: []string{"Filter1"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
		Contexts: []string{"con1", "con2"},
		Attributes: []*Attribute{
			&Attribute{
				FieldName:  "FN1",
				Initial:    "Init1",
				Substitute: "Val1",
				Append:     true,
			},
		},
		Weight: 20,
	}
	//Set AttributeProfile with 2 contexts ( con1 , con2)
	if err := dataManager.SetAttributeProfile(attrProfile, true); err != nil {
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
		rfi := NewFilterIndexer(onStor, utils.AttributeProfilePrefix,
			utils.ConcatenatedKey(attrProfile.Tenant, ctx))
		if rcvIdx, err := dataManager.GetFilterIndexes(
			GetDBIndexKey(rfi.itemType, rfi.dbKeySuffix, false), MetaString,
			nil); err != nil {
			t.Error(err)
		} else {
			if !reflect.DeepEqual(eIdxes, rcvIdx) {
				t.Errorf("Expecting %+v, received: %+v", eIdxes, rcvIdx)
			}
		}
		if reverseRcvIdx, err := dataManager.GetFilterReverseIndexes(
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
	if err := dataManager.SetAttributeProfile(attrProfile, true); err != nil {
		t.Error(err)
	}
	//check indexes with the new context (con3)
	rfi := NewFilterIndexer(onStor, utils.AttributeProfilePrefix,
		utils.ConcatenatedKey(attrProfile.Tenant, "con3"))
	if rcvIdx, err := dataManager.GetFilterIndexes(
		GetDBIndexKey(rfi.itemType, rfi.dbKeySuffix, false), MetaString,
		nil); err != nil {
		t.Error(err)
	} else {
		if !reflect.DeepEqual(eIdxes, rcvIdx) {
			t.Errorf("Expecting %+v, received: %+v", eIdxes, rcvIdx)
		}
	}
	if reverseRcvIdx, err := dataManager.GetFilterReverseIndexes(
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
		rfi := NewFilterIndexer(onStor, utils.AttributeProfilePrefix,
			utils.ConcatenatedKey(attrProfile.Tenant, ctx))
		if _, err := dataManager.GetFilterIndexes(
			GetDBIndexKey(rfi.itemType, rfi.dbKeySuffix, false), MetaString,
			nil); err != nil && err != utils.ErrNotFound {
			t.Error(err)
		}
		if _, err := dataManager.GetFilterReverseIndexes(
			GetDBIndexKey(rfi.itemType, rfi.dbKeySuffix, true),
			nil); err != nil && err != utils.ErrNotFound {
			t.Error(err)
		}
	}

	if err := dataManager.RemoveAttributeProfile(attrProfile.Tenant,
		attrProfile.ID, attrProfile.Contexts, utils.NonTransactional, true); err != nil {
		t.Error(err)
	}
	//check if index is removed
	rfi = NewFilterIndexer(onStor, utils.AttributeProfilePrefix, utils.ConcatenatedKey("cgrates.org", "con3"))
	if _, err := dataManager.GetFilterIndexes(
		GetDBIndexKey(rfi.itemType, rfi.dbKeySuffix, false), MetaString,
		nil); err != nil && err != utils.ErrNotFound {
		t.Error(err)
	}
	if _, err := dataManager.GetFilterReverseIndexes(
		GetDBIndexKey(rfi.itemType, rfi.dbKeySuffix, true),
		nil); err != nil && err != utils.ErrNotFound {
		t.Error(err)
	}
}

func testITTestThresholdInlineFilterIndexing(t *testing.T) {
	fp := &Filter{
		Tenant: "cgrates.org",
		ID:     "Filter1",
		Rules: []*FilterRule{
			&FilterRule{
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
	if err := dataManager.SetFilter(fp); err != nil {
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

	if err := dataManager.SetThresholdProfile(th, true); err != nil {
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
	rfi := NewFilterIndexer(onStor, utils.ThresholdProfilePrefix, th.Tenant)
	if rcvIdx, err := dataManager.GetFilterIndexes(
		GetDBIndexKey(rfi.itemType, rfi.dbKeySuffix, false), MetaString,
		nil); err != nil {
		t.Error(err)
	} else {
		if !reflect.DeepEqual(eIdxes, rcvIdx) {
			t.Errorf("Expecting %+v, received: %+v", eIdxes, rcvIdx)
		}
	}
	if reverseRcvIdx, err := dataManager.GetFilterReverseIndexes(
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
	if err := dataManager.SetThresholdProfile(th, true); err != nil {
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
	if rcvIdx, err := dataManager.GetFilterIndexes(
		GetDBIndexKey(rfi.itemType, rfi.dbKeySuffix, false), MetaString,
		nil); err != nil {
		t.Error(err)
	} else {
		if !reflect.DeepEqual(eIdxes, rcvIdx) {
			t.Errorf("Expecting %+v, received: %+v", eIdxes, rcvIdx)
		}
	}
	if reverseRcvIdx, err := dataManager.GetFilterReverseIndexes(
		GetDBIndexKey(rfi.itemType, rfi.dbKeySuffix, true),
		nil); err != nil {
		t.Error(err)
	} else {
		if !reflect.DeepEqual(reverseIdxes, reverseRcvIdx) {
			t.Errorf("Expecting %+v, received: %+v", reverseIdxes, reverseRcvIdx)
		}
	}
	//remove threshold
	if err := dataManager.RemoveThresholdProfile(th.Tenant,
		th.ID, utils.NonTransactional, true); err != nil {
		t.Error(err)
	}
	if _, err := dataManager.GetFilterIndexes(
		GetDBIndexKey(rfi.itemType, rfi.dbKeySuffix, false), MetaString,
		nil); err != utils.ErrNotFound {
		t.Error(err)
	}
	if _, err := dataManager.GetFilterReverseIndexes(
		GetDBIndexKey(rfi.itemType, rfi.dbKeySuffix, true),
		nil); err != utils.ErrNotFound {
		t.Error(err)
	}
}

func testITTestStoreFilterIndexesWithTransID(t *testing.T) {
	tmpKey := utils.TEMP_DESTINATION_PREFIX + utils.ConcatenatedKey(GetDBIndexKey(utils.ResourceProfilesPrefix, "cgrates.org", false), "transaction1")
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
	if err := dataManager.SetFilterIndexes(
		GetDBIndexKey(utils.ResourceProfilesPrefix, "cgrates.org", false),
		idxes, false, "transaction1"); err != nil {
		t.Error(err)
	}
	if rcv, err := dataManager.GetFilterIndexes(tmpKey, MetaString, nil); err != nil {
		t.Error(err)
	} else {
		if !reflect.DeepEqual(idxes, rcv) {
			t.Errorf("Expecting: %+v, received: %+v", idxes, rcv)
		}
	}
	//commit transaction
	if err := dataManager.SetFilterIndexes(
		GetDBIndexKey(utils.ResourceProfilesPrefix, "cgrates.org", false),
		idxes, true, "transaction1"); err != nil {
		t.Error(err)
	}
	//verify if old key was deleted
	if _, err := dataManager.GetFilterIndexes(tmpKey, MetaString, nil); err != utils.ErrNotFound {
		t.Error(err)
	}
	//verify new key and check if data was moved
	if rcv, err := dataManager.GetFilterIndexes(
		GetDBIndexKey(utils.ResourceProfilesPrefix, "cgrates.org", false),
		MetaString, nil); err != nil {
		t.Error(err)
	} else {
		if !reflect.DeepEqual(idxes, rcv) {
			t.Errorf("Expecting: %+v, received: %+v", idxes, rcv)
		}
	}
}

func testITTestStoreFilterIndexesWithTransID2(t *testing.T) {
	tmpKey := utils.TEMP_DESTINATION_PREFIX + utils.ConcatenatedKey(GetDBIndexKey(utils.ResourceProfilesPrefix, "cgrates.org", false), "transaction1")
	idxes := map[string]utils.StringMap{
		"*string:Event:Event1": utils.StringMap{
			"RL1": true,
		},
		"*string:Event:Event2": utils.StringMap{
			"RL1": true,
			"RL2": true,
		},
	}
	if err := dataManager.SetFilterIndexes(
		GetDBIndexKey(utils.ResourceProfilesPrefix, "cgrates.org", false),
		idxes, false, "transaction1"); err != nil {
		t.Error(err)
	}
	if rcv, err := dataManager.GetFilterIndexes(tmpKey, MetaString, nil); err != nil {
		t.Error(err)
	} else {
		if !reflect.DeepEqual(idxes, rcv) {
			t.Errorf("Expecting: %+v, received: %+v", idxes, rcv)
		}
	}
	//commit transaction
	if err := dataManager.SetFilterIndexes(
		GetDBIndexKey(utils.ResourceProfilesPrefix, "cgrates.org", false),
		idxes, true, "transaction1"); err != nil {
		t.Error(err)
	}
	//verify if old key was deleted
	if _, err := dataManager.GetFilterIndexes(tmpKey, MetaString, nil); err != utils.ErrNotFound {
		t.Error(err)
	}
	//verify new key and check if data was moved
	if rcv, err := dataManager.GetFilterIndexes(
		GetDBIndexKey(utils.ResourceProfilesPrefix, "cgrates.org", false),
		MetaString, nil); err != nil {
		t.Error(err)
	} else {
		if !reflect.DeepEqual(idxes, rcv) {
			t.Errorf("Expecting: %+v, received: %+v", idxes, rcv)
		}
	}
}

func testITTestIndexingWithEmptyFltrID(t *testing.T) {
	th := &ThresholdProfile{
		Tenant:             "cgrates.org",
		ID:                 "THD_Test",
		ActivationInterval: &utils.ActivationInterval{},
		FilterIDs:          []string{},
		Recurrent:          true,
		MinSleep:           time.Duration(0 * time.Second),
		Blocker:            true,
		Weight:             1.4,
		ActionIDs:          []string{},
	}

	if err := dataManager.SetThresholdProfile(th, true); err != nil {
		t.Error(err)
	}
	eIdxes := map[string]utils.StringMap{
		"*default:*any:*any": utils.StringMap{
			"THD_Test": true,
		},
	}
	reverseIdxes := map[string]utils.StringMap{
		"THD_Test": utils.StringMap{
			"*default:*any:*any": true,
		},
	}
	rfi := NewFilterIndexer(onStor, utils.ThresholdProfilePrefix, th.Tenant)
	if rcvIdx, err := dataManager.GetFilterIndexes(
		GetDBIndexKey(rfi.itemType, rfi.dbKeySuffix, false), MetaString,
		nil); err != nil {
		t.Error(err)
	} else {
		if !reflect.DeepEqual(eIdxes, rcvIdx) {
			t.Errorf("Expecting %+v, received: %+v", eIdxes, rcvIdx)
		}
	}
	if reverseRcvIdx, err := dataManager.GetFilterReverseIndexes(
		GetDBIndexKey(rfi.itemType, rfi.dbKeySuffix, true),
		nil); err != nil {
		t.Error(err)
	} else {
		if !reflect.DeepEqual(reverseIdxes, reverseRcvIdx) {
			t.Errorf("Expecting %+v, received: %+v", reverseIdxes, reverseRcvIdx)
		}
	}
	eMp := utils.StringMap{
		"THD_Test": true,
	}
	if rcvMp, err := dataManager.MatchFilterIndex(
		GetDBIndexKey(rfi.itemType, rfi.dbKeySuffix, false),
		utils.MetaDefault, utils.META_ANY, utils.META_ANY); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eMp, rcvMp) {
		t.Errorf("Expecting: %+v, received: %+v", eMp, rcvMp)
	}
}

func testITTestIndexingWithEmptyFltrID2(t *testing.T) {
	splProfile := &SupplierProfile{
		Tenant:    "cgrates.org",
		ID:        "SPL_Weight",
		FilterIDs: []string{},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
		Sorting:           "*weight",
		SortingParameters: []string{},
		Suppliers: []*Supplier{
			&Supplier{
				ID:                 "supplier1",
				FilterIDs:          []string{""},
				AccountIDs:         []string{""},
				RatingPlanIDs:      []string{""},
				ResourceIDs:        []string{""},
				StatIDs:            []string{""},
				Weight:             10,
				SupplierParameters: "",
			},
		},
		Weight: 20,
	}

	if err := dataManager.SetSupplierProfile(splProfile, true); err != nil {
		t.Error(err)
	}
	eIdxes := map[string]utils.StringMap{
		"*default:*any:*any": utils.StringMap{
			"SPL_Weight": true,
		},
	}
	reverseIdxes := map[string]utils.StringMap{
		"SPL_Weight": utils.StringMap{
			"*default:*any:*any": true,
		},
	}
	rfi := NewFilterIndexer(onStor, utils.SupplierProfilePrefix, splProfile.Tenant)
	if rcvIdx, err := dataManager.GetFilterIndexes(
		GetDBIndexKey(rfi.itemType, rfi.dbKeySuffix, false), MetaString,
		nil); err != nil {
		t.Error(err)
	} else {
		if !reflect.DeepEqual(eIdxes, rcvIdx) {
			t.Errorf("Expecting %+v, received: %+v", eIdxes, rcvIdx)
		}
	}
	if reverseRcvIdx, err := dataManager.GetFilterReverseIndexes(
		GetDBIndexKey(rfi.itemType, rfi.dbKeySuffix, true),
		nil); err != nil {
		t.Error(err)
	} else {
		if !reflect.DeepEqual(reverseIdxes, reverseRcvIdx) {
			t.Errorf("Expecting %+v, received: %+v", reverseIdxes, reverseRcvIdx)
		}
	}
	eMp := utils.StringMap{
		"SPL_Weight": true,
	}
	if rcvMp, err := dataManager.MatchFilterIndex(
		GetDBIndexKey(rfi.itemType, rfi.dbKeySuffix, false),
		utils.MetaDefault, utils.META_ANY, utils.META_ANY); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eMp, rcvMp) {
		t.Errorf("Expecting: %+v, received: %+v", eMp, rcvMp)
	}
}
