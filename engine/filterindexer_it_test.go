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

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

var (
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
	testITFlush,
	testITIsDBEmpty,
	testITTestIndexingThresholds,
}

func TestFilterIndexerITRedis(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	redisDB, err := NewRedisStorage(
		fmt.Sprintf("%s:%s", cfg.DataDbCfg().DataDbHost, cfg.DataDbCfg().DataDbPort),
		4, cfg.DataDbCfg().DataDbPass, cfg.GeneralCfg().DBDataEncoding,
		utils.REDIS_MAX_CONNS, nil, "")
	if err != nil {
		t.Fatal("Could not connect to Redis", err.Error())
	}
	cfgDBName = cfg.DataDbCfg().DataDbName
	dataManager = NewDataManager(redisDB)
	for _, stest := range sTests {
		t.Run("TestITRedis", stest)
	}
}

func TestFilterIndexerITMongo(t *testing.T) {
	cdrsMongoCfgPath := path.Join(*dataDir, "conf", "samples", "tutmongo")
	mgoITCfg, err := config.NewCGRConfigFromFolder(cdrsMongoCfgPath)
	if err != nil {
		t.Fatal(err)
	}
	mongoDB, err := NewMongoStorage(mgoITCfg.StorDbCfg().StorDBHost,
		mgoITCfg.StorDbCfg().StorDBPort, mgoITCfg.StorDbCfg().StorDBName,
		mgoITCfg.StorDbCfg().StorDBUser, mgoITCfg.StorDbCfg().StorDBPass,
		utils.StorDB, nil, mgoITCfg.CacheCfg(), false)
	if err != nil {
		t.Fatal(err)
	}
	cfgDBName = mgoITCfg.StorDbCfg().StorDBName
	dataManager = NewDataManager(mongoDB)
	for _, stest := range sTests {
		t.Run("TestITMongo", stest)
	}
}

func TestFilterIndexerITInternal(t *testing.T) {
	mapDataDB, err := NewMapStorage()
	if err != nil {
		t.Fatal(err)
	}
	dataManager = NewDataManager(mapDataDB)
	for _, stest := range sTests {
		t.Run("TestITInternal", stest)
	}
}

func testITFlush(t *testing.T) {
	if err := dataManager.DataDB().Flush(""); err != nil {
		t.Error(err)
	}
	Cache.Clear(nil)
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
		"*string:Account:1001": {
			"RL1": true,
		},
		"*string:Account:1002": {
			"RL1": true,
			"RL2": true,
		},
		"*string:Account:dan": {
			"RL2": true,
		},
		"*string:Subject:dan": {
			"RL2": true,
			"RL3": true,
		},
		utils.ConcatenatedKey(utils.META_NONE, utils.ANY, utils.ANY): {
			"RL4": true,
			"RL5": true,
		},
	}
	if err := dataManager.SetFilterIndexes(
		utils.PrefixToIndexCache[utils.ResourceProfilesPrefix], "cgrates.org",
		idxes, false, utils.NonTransactional); err != nil {
		t.Error(err)
	}
}

func testITGetFilterIndexes(t *testing.T) {
	eIdxes := map[string]utils.StringMap{
		"*string:Account:1001": {
			"RL1": true,
		},
		"*string:Account:1002": {
			"RL1": true,
			"RL2": true,
		},
		"*string:Account:dan": {
			"RL2": true,
		},
		"*string:Subject:dan": {
			"RL2": true,
			"RL3": true,
		},
		utils.ConcatenatedKey(utils.META_NONE, utils.ANY, utils.ANY): {
			"RL4": true,
			"RL5": true,
		},
	}
	sbjDan := map[string]string{
		"Subject": "dan",
	}
	expectedsbjDan := map[string]utils.StringMap{
		"*string:Subject:dan": {
			"RL2": true,
			"RL3": true,
		},
	}

	if exsbjDan, err := dataManager.GetFilterIndexes(
		utils.PrefixToIndexCache[utils.ResourceProfilesPrefix],
		"cgrates.org", MetaString, sbjDan); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedsbjDan, exsbjDan) {
		t.Errorf("Expecting: %+v, received: %+v", expectedsbjDan, exsbjDan)
	}
	if rcv, err := dataManager.GetFilterIndexes(
		utils.PrefixToIndexCache[utils.ResourceProfilesPrefix],
		"cgrates.org", utils.EmptyString, nil); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eIdxes, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", eIdxes, rcv)
	}
	if _, err := dataManager.GetFilterIndexes("unknown_key", "unkonwn_tenant",
		utils.EmptyString, nil); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	}
}

func testITMatchFilterIndex(t *testing.T) {
	eMp := utils.StringMap{
		"RL1": true,
		"RL2": true,
	}
	if rcvMp, err := dataManager.MatchFilterIndex(
		utils.CacheResourceFilterIndexes, "cgrates.org",
		MetaString, "Account", "1002"); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eMp, rcvMp) {
		t.Errorf("Expecting: %+v, received: %+v", eMp, rcvMp)
	}
	if _, err := dataManager.MatchFilterIndex(
		utils.CacheResourceFilterIndexes, "cgrates.org",
		MetaString, "NonexistentField", "1002"); err == nil ||
		err != utils.ErrNotFound {
		t.Error(err)
	}
}

func testITTestThresholdFilterIndexes(t *testing.T) {
	fp := &Filter{
		Tenant: "cgrates.org",
		ID:     "Filter1",
		Rules: []*FilterRule{
			{
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
		MaxHits:            12,
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
		MaxHits:            12,
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
		"*string:EventType:Event1": {
			"THD_Test":  true,
			"THD_Test2": true,
		},
		"*string:EventType:Event2": {
			"THD_Test":  true,
			"THD_Test2": true,
		},
	}
	rfi := NewFilterIndexer(onStor, utils.ThresholdProfilePrefix, th.Tenant)
	if rcvIdx, err := dataManager.GetFilterIndexes(
		utils.PrefixToIndexCache[rfi.itemType], rfi.dbKeySuffix,
		utils.EmptyString, nil); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eIdxes, rcvIdx) {
		t.Errorf("Expecting %+v, received: %+v", eIdxes, rcvIdx)
	}

	//Replace existing filter (Filter1 -> Filter2)
	fp2 := &Filter{
		Tenant: "cgrates.org",
		ID:     "Filter2",
		Rules: []*FilterRule{
			{
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
	time.Sleep(50 * time.Millisecond)
	if err := dataManager.SetThresholdProfile(th, true); err != nil {
		t.Error(err)
	}
	eIdxes = map[string]utils.StringMap{
		"*string:Account:1001": {
			"THD_Test": true,
		},
		"*string:Account:1002": {
			"THD_Test": true,
		},
		"*string:EventType:Event1": {
			"THD_Test2": true,
		},
		"*string:EventType:Event2": {
			"THD_Test2": true,
		},
	}
	if rcvIdx, err := dataManager.GetFilterIndexes(
		utils.PrefixToIndexCache[rfi.itemType], rfi.dbKeySuffix,
		utils.EmptyString, nil); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eIdxes, rcvIdx) {
		t.Errorf("Expecting %+v, received: %+v", eIdxes, rcvIdx)
	}
	//replace old filter with two different filters
	fp3 := &Filter{
		Tenant: "cgrates.org",
		ID:     "Filter3",
		Rules: []*FilterRule{
			{
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
	time.Sleep(50 * time.Millisecond)
	if err := dataManager.SetThresholdProfile(th, true); err != nil {
		t.Error(err)
	}
	eIdxes = map[string]utils.StringMap{
		"*string:Destination:10": {
			"THD_Test": true,
		},
		"*string:Destination:20": {
			"THD_Test": true,
		},
		"*string:EventType:Event1": {
			"THD_Test":  true,
			"THD_Test2": true,
		},
		"*string:EventType:Event2": {
			"THD_Test":  true,
			"THD_Test2": true,
		},
	}
	if rcvIdx, err := dataManager.GetFilterIndexes(
		utils.PrefixToIndexCache[rfi.itemType], rfi.dbKeySuffix,
		utils.EmptyString, nil); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eIdxes, rcvIdx) {
		t.Errorf("Expecting %+v, received: %+v", eIdxes, rcvIdx)
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
		utils.PrefixToIndexCache[rfi.itemType], rfi.dbKeySuffix,
		utils.EmptyString, nil); err != utils.ErrNotFound {
		t.Error(err)
	}
}

func testITTestAttributeProfileFilterIndexes(t *testing.T) {
	fp := &Filter{
		Tenant: "cgrates.org",
		ID:     "AttrFilter",
		Rules: []*FilterRule{
			{
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
		FilterIDs: []string{"AttrFilter"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
		Contexts: []string{"con1", "con2"},
		Attributes: []*Attribute{
			{
				FieldName:  "FN1",
				Initial:    "Init1",
				Substitute: config.NewRSRParsersMustCompile("Val1", true, utils.INFIELD_SEP),
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
		"*string:EventType:Event1": {
			"AttrPrf": true,
		},
		"*string:EventType:Event2": {
			"AttrPrf": true,
		},
	}
	for _, ctx := range attrProfile.Contexts {
		rfi := NewFilterIndexer(onStor, utils.AttributeProfilePrefix,
			utils.ConcatenatedKey(attrProfile.Tenant, ctx))
		if rcvIdx, err := dataManager.GetFilterIndexes(
			utils.PrefixToIndexCache[rfi.itemType], rfi.dbKeySuffix,
			utils.EmptyString, nil); err != nil {
			t.Error(err)
		} else if !reflect.DeepEqual(eIdxes, rcvIdx) {
			t.Errorf("Expecting %+v, received: %+v", eIdxes, rcvIdx)
		}
	}
	//Set AttributeProfile with 1 new context (con3)
	attrProfile.Contexts = []string{"con3"}
	time.Sleep(50 * time.Millisecond)
	if err := dataManager.SetAttributeProfile(attrProfile, true); err != nil {
		t.Error(err)
	}
	//check indexes with the new context (con3)
	rfi := NewFilterIndexer(onStor, utils.AttributeProfilePrefix,
		utils.ConcatenatedKey(attrProfile.Tenant, "con3"))
	if rcvIdx, err := dataManager.GetFilterIndexes(
		utils.PrefixToIndexCache[rfi.itemType], rfi.dbKeySuffix,
		utils.EmptyString, nil); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eIdxes, rcvIdx) {
		t.Errorf("Expecting %+v, received: %+v", eIdxes, rcvIdx)
	}
	//check if old contexts was delete
	for _, ctx := range []string{"con1", "con2"} {
		rfi := NewFilterIndexer(onStor, utils.AttributeProfilePrefix,
			utils.ConcatenatedKey(attrProfile.Tenant, ctx))
		if _, err := dataManager.GetFilterIndexes(
			utils.PrefixToIndexCache[rfi.itemType], rfi.dbKeySuffix,
			utils.EmptyString, nil); err != nil && err != utils.ErrNotFound {
			t.Error(err)
		}
	}

	if err := dataManager.RemoveAttributeProfile(attrProfile.Tenant,
		attrProfile.ID, utils.NonTransactional, true); err != nil {
		t.Error(err)
	}
	//check if index is removed
	rfi = NewFilterIndexer(onStor, utils.AttributeProfilePrefix,
		utils.ConcatenatedKey("cgrates.org", "con3"))
	if _, err := dataManager.GetFilterIndexes(
		utils.PrefixToIndexCache[rfi.itemType], rfi.dbKeySuffix,
		MetaString, nil); err != nil && err != utils.ErrNotFound {
		t.Error(err)
	}

}

func testITTestThresholdInlineFilterIndexing(t *testing.T) {
	fp := &Filter{
		Tenant: "cgrates.org",
		ID:     "Filter1",
		Rules: []*FilterRule{
			{
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
		MaxHits:            12,
		MinSleep:           time.Duration(0 * time.Second),
		Blocker:            true,
		Weight:             1.4,
		ActionIDs:          []string{},
	}

	if err := dataManager.SetThresholdProfile(th, true); err != nil {
		t.Error(err)
	}
	eIdxes := map[string]utils.StringMap{
		"*string:EventType:Event1": {
			"THD_Test": true,
		},
		"*string:EventType:Event2": {
			"THD_Test": true,
		},
	}
	rfi := NewFilterIndexer(onStor, utils.ThresholdProfilePrefix, th.Tenant)
	if rcvIdx, err := dataManager.GetFilterIndexes(
		utils.PrefixToIndexCache[rfi.itemType], rfi.dbKeySuffix,
		utils.EmptyString, nil); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eIdxes, rcvIdx) {
		t.Errorf("Expecting %+v, received: %+v", eIdxes, rcvIdx)
	}
	//Add an InlineFilter
	th.FilterIDs = []string{"Filter1", "*string:Account:1001"}
	time.Sleep(50 * time.Millisecond)
	if err := dataManager.SetThresholdProfile(th, true); err != nil {
		t.Error(err)
	}
	eIdxes = map[string]utils.StringMap{
		"*string:Account:1001": {
			"THD_Test": true,
		},
		"*string:EventType:Event1": {
			"THD_Test": true,
		},
		"*string:EventType:Event2": {
			"THD_Test": true,
		},
	}
	if rcvIdx, err := dataManager.GetFilterIndexes(
		utils.PrefixToIndexCache[rfi.itemType], rfi.dbKeySuffix,
		utils.EmptyString, nil); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eIdxes, rcvIdx) {
		t.Errorf("Expecting %+v, received: %+v", eIdxes, rcvIdx)
	}
	//remove threshold
	if err := dataManager.RemoveThresholdProfile(th.Tenant,
		th.ID, utils.NonTransactional, true); err != nil {
		t.Error(err)
	}
	if _, err := dataManager.GetFilterIndexes(
		utils.PrefixToIndexCache[rfi.itemType], rfi.dbKeySuffix,
		utils.EmptyString, nil); err != utils.ErrNotFound {
		t.Error(err)
	}
}

func testITTestStoreFilterIndexesWithTransID(t *testing.T) {
	idxes := map[string]utils.StringMap{
		"*string:Account:1001": {
			"RL1": true,
		},
		"*string:Account:1002": {
			"RL1": true,
			"RL2": true,
		},
		"*string:Account:dan": {
			"RL2": true,
		},
		"*string:Subject:dan": {
			"RL2": true,
			"RL3": true,
		},
		utils.ConcatenatedKey(utils.META_NONE,
			utils.ANY, utils.ANY): {
			"RL4": true,
			"RL5": true,
		},
	}
	if err := dataManager.SetFilterIndexes(
		utils.PrefixToIndexCache[utils.ResourceProfilesPrefix], "cgrates.org",
		idxes, false, "transaction1"); err != nil {
		t.Error(err)
	}

	//commit transaction
	if err := dataManager.SetFilterIndexes(
		utils.PrefixToIndexCache[utils.ResourceProfilesPrefix],
		"cgrates.org", idxes, true, "transaction1"); err != nil {
		t.Error(err)
	}
	eIdx := map[string]utils.StringMap{
		"*string:Account:1001": {
			"RL1": true,
		},
		"*string:Account:1002": {
			"RL1": true,
			"RL2": true,
		},
		"*string:Account:dan": {
			"RL2": true,
		},
		"*string:Subject:dan": {
			"RL2": true,
			"RL3": true,
		},
		utils.ConcatenatedKey(utils.META_NONE,
			utils.ANY, utils.ANY): {
			"RL4": true,
			"RL5": true,
		},
	}

	//verify new key and check if data was moved
	if rcv, err := dataManager.GetFilterIndexes(
		utils.PrefixToIndexCache[utils.ResourceProfilesPrefix], "cgrates.org",
		utils.EmptyString, nil); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eIdx, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", eIdx, rcv)
	}
}

func testITTestStoreFilterIndexesWithTransID2(t *testing.T) {
	idxes := map[string]utils.StringMap{
		"*string:Event:Event1": {
			"RL1": true,
		},
		"*string:Event:Event2": {
			"RL1": true,
			"RL2": true,
		},
	}
	transID := "transaction1"
	if err := dataManager.SetFilterIndexes(
		utils.PrefixToIndexCache[utils.ResourceProfilesPrefix], "cgrates.org",
		idxes, false, transID); err != nil {
		t.Error(err)
	}
	//commit transaction
	if err := dataManager.SetFilterIndexes(
		utils.PrefixToIndexCache[utils.ResourceProfilesPrefix], "cgrates.org",
		idxes, true, transID); err != nil {
		t.Error(err)
	}
	//verify if old key was deleted
	if _, err := dataManager.GetFilterIndexes(
		"tmp_"+utils.PrefixToIndexCache[utils.ResourceProfilesPrefix],
		utils.ConcatenatedKey("cgrates.org", transID),
		utils.EmptyString, nil); err != utils.ErrNotFound {
		t.Error(err)
	}
	//verify new key and check if data was moved
	if rcv, err := dataManager.GetFilterIndexes(
		utils.PrefixToIndexCache[utils.ResourceProfilesPrefix],
		"cgrates.org", utils.EmptyString, nil); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(idxes, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", idxes, rcv)
	}
}

func testITTestIndexingWithEmptyFltrID(t *testing.T) {
	th := &ThresholdProfile{
		Tenant:             "cgrates.org",
		ID:                 "THD_Test",
		ActivationInterval: &utils.ActivationInterval{},
		FilterIDs:          []string{},
		MaxHits:            12,
		MinSleep:           time.Duration(0 * time.Second),
		Blocker:            true,
		Weight:             1.4,
		ActionIDs:          []string{},
	}
	th2 := &ThresholdProfile{
		Tenant:             "cgrates.org",
		ID:                 "THD_Test2",
		ActivationInterval: &utils.ActivationInterval{},
		FilterIDs:          []string{},
		MaxHits:            12,
		MinSleep:           time.Duration(0 * time.Second),
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
		"*none:*any:*any": {
			"THD_Test":  true,
			"THD_Test2": true,
		},
	}
	rfi := NewFilterIndexer(onStor, utils.ThresholdProfilePrefix, th.Tenant)
	if rcvIdx, err := dataManager.GetFilterIndexes(
		utils.PrefixToIndexCache[rfi.itemType], rfi.dbKeySuffix,
		utils.META_NONE, nil); err != nil {
		t.Error(err)
	} else {
		if !reflect.DeepEqual(eIdxes, rcvIdx) {
			t.Errorf("Expecting %+v, received: %+v", eIdxes, rcvIdx)
		}
	}
	eMp := utils.StringMap{
		"THD_Test":  true,
		"THD_Test2": true,
	}
	if rcvMp, err := dataManager.MatchFilterIndex(utils.CacheThresholdFilterIndexes, th.Tenant,
		utils.META_NONE, utils.META_ANY, utils.META_ANY); err != nil {
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
			{
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
	splProfile2 := &SupplierProfile{
		Tenant:    "cgrates.org",
		ID:        "SPL_Weight2",
		FilterIDs: []string{},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
		Sorting:           "*weight",
		SortingParameters: []string{},
		Suppliers: []*Supplier{
			{
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
	if err := dataManager.SetSupplierProfile(splProfile2, true); err != nil {
		t.Error(err)
	}
	eIdxes := map[string]utils.StringMap{
		"*none:*any:*any": {
			"SPL_Weight":  true,
			"SPL_Weight2": true,
		},
	}
	rfi := NewFilterIndexer(onStor, utils.SupplierProfilePrefix, splProfile.Tenant)
	if rcvIdx, err := dataManager.GetFilterIndexes(
		utils.PrefixToIndexCache[rfi.itemType], rfi.dbKeySuffix,
		utils.EmptyString, nil); err != nil {
		t.Error(err)
	} else {
		if !reflect.DeepEqual(eIdxes, rcvIdx) {
			t.Errorf("Expecting %+v, received: %+v", eIdxes, rcvIdx)
		}
	}
	eMp := utils.StringMap{
		"SPL_Weight":  true,
		"SPL_Weight2": true,
	}
	if rcvMp, err := dataManager.MatchFilterIndex(utils.CacheSupplierFilterIndexes,
		splProfile.Tenant, utils.META_NONE, utils.META_ANY, utils.META_ANY); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eMp, rcvMp) {
		t.Errorf("Expecting: %+v, received: %+v", eMp, rcvMp)
	}
}

func testITTestIndexingThresholds(t *testing.T) {
	th := &ThresholdProfile{
		Tenant:    "cgrates.org",
		ID:        "TH1",
		FilterIDs: []string{"*string:Account:1001", "*gt:Balance:1000"},
		ActionIDs: []string{},
	}
	th2 := &ThresholdProfile{
		Tenant:    "cgrates.org",
		ID:        "TH2",
		FilterIDs: []string{"*string:Account:1001", "*gt:Balance:1000"},
		ActionIDs: []string{},
	}
	th3 := &ThresholdProfile{
		Tenant:    "cgrates.org",
		ID:        "TH3",
		FilterIDs: []string{"*string:Account:1002", "*lt:Balance:1000"},
		ActionIDs: []string{},
	}
	rfi := NewFilterIndexer(onStor, utils.ThresholdProfilePrefix, th.Tenant)
	if err := dataManager.SetThresholdProfile(th, true); err != nil {
		t.Error(err)
	}
	if err := dataManager.SetThresholdProfile(th2, true); err != nil {
		t.Error(err)
	}
	if err := dataManager.SetThresholdProfile(th3, true); err != nil {
		t.Error(err)
	}
	eIdxes := map[string]utils.StringMap{
		"*string:Account:1001": {
			"TH1": true,
			"TH2": true,
		},
		"*string:Account:1002": {
			"TH3": true,
		},
	}
	if rcvIdx, err := dataManager.GetFilterIndexes(
		utils.PrefixToIndexCache[rfi.itemType], rfi.dbKeySuffix,
		utils.EmptyString, nil); err != nil {
		t.Error(err)
	} else {
		if !reflect.DeepEqual(eIdxes, rcvIdx) {
			t.Errorf("Expecting %+v, received: %+v", eIdxes, rcvIdx)
		}
	}
	eMp := utils.StringMap{
		"TH1": true,
		"TH2": true,
	}
	if rcvMp, err := dataManager.MatchFilterIndex(utils.CacheThresholdFilterIndexes, th.Tenant,
		utils.MetaString, utils.Account, "1001"); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eMp, rcvMp) {
		t.Errorf("Expecting: %+v, received: %+v", eMp, rcvMp)
	}
}
