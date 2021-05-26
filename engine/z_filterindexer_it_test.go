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
	testITTestAttributeProfileFilterIndexes2,
	testITTestThresholdInlineFilterIndexing,
	testITFlush,
	testITIsDBEmpty,
	testITTestStoreFilterIndexesWithTransID,
	testITFlush,
	testITIsDBEmpty,
	testITFlush,
	testITIsDBEmpty,
	testITResourceProfileIndexes,
	testITFlush,
	testITIsDBEmpty,
	testITStatQueueProfileIndexes,
	testITFlush,
	testITIsDBEmpty,
	testITChargerProfileIndexes,
	testITFlush,
	testITIsDBEmpty,
	testITDispatcherProfileIndexes,
	testITFlush,
	testITIsDBEmpty,
	testITFlush,
	testITIsDBEmpty,
	testITTestStoreFilterIndexesWithTransID2,
	testITFlush,
	testITIsDBEmpty,
	testITTestIndexingWithEmptyFltrID,
	testITTestIndexingWithEmptyFltrID2,
	testITFlush,
	testITIsDBEmpty,
	testITTestIndexingThresholds,
	testITFlush,
	testITIsDBEmpty,
	testITTestIndexingMetaNot,
	testITFlush,
	testITIsDBEmpty,
	testITFlush,
	testITIsDBEmpty,
	testITTestIndexingMetaSuffix,
}

func TestFilterIndexerIT(t *testing.T) {
	switch *dbType {
	case utils.MetaInternal:
		dataManager = NewDataManager(NewInternalDB(nil, nil, true),
			config.CgrConfig().CacheCfg(), nil)
	case utils.MetaMySQL:
		cfg := config.NewDefaultCGRConfig()
		redisDB, err := NewRedisStorage(
			fmt.Sprintf("%s:%s", cfg.DataDbCfg().Host, cfg.DataDbCfg().Port),
			4, cfg.DataDbCfg().User, cfg.DataDbCfg().Password, cfg.GeneralCfg().DBDataEncoding,
			utils.RedisMaxConns, "", false, 0, 0, false, utils.EmptyString, utils.EmptyString, utils.EmptyString)
		if err != nil {
			t.Fatal("Could not connect to Redis", err.Error())
		}
		cfgDBName = cfg.DataDbCfg().Name
		defer redisDB.Close()
		dataManager = NewDataManager(redisDB, config.CgrConfig().CacheCfg(), nil)
	case utils.MetaMongo:
		cdrsMongoCfgPath := path.Join(*dataDir, "conf", "samples", "tutmongo")
		mgoITCfg, err := config.NewCGRConfigFromPath(cdrsMongoCfgPath)
		if err != nil {
			t.Fatal(err)
		}
		mongoDB, err := NewMongoStorage(mgoITCfg.DataDbCfg().Host,
			mgoITCfg.DataDbCfg().Port, mgoITCfg.DataDbCfg().Name,
			mgoITCfg.DataDbCfg().User, mgoITCfg.DataDbCfg().Password,
			mgoITCfg.GeneralCfg().DBDataEncoding,
			utils.StorDB, nil, 10*time.Second)
		if err != nil {
			t.Fatal(err)
		}
		cfgDBName = mgoITCfg.DataDbCfg().Name
		defer mongoDB.Close()
		dataManager = NewDataManager(mongoDB, config.CgrConfig().CacheCfg(), nil)
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	for _, stest := range sTests {
		t.Run(*dbType, stest)
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
		t.Errorf("Expecting: true got :%+v", test)
	}
}

func testITSetFilterIndexes(t *testing.T) {
	idxes := map[string]utils.StringSet{
		"*string:Account:1001": {
			"RL1": struct{}{},
		},
		"*string:Account:1002": {
			"RL1": struct{}{},
			"RL2": struct{}{},
		},
		"*string:Account:dan": {
			"RL2": struct{}{},
		},
		"*string:Subject:dan": {
			"RL2": struct{}{},
			"RL3": struct{}{},
		},
		utils.ConcatenatedKey(utils.MetaNone, utils.MetaAny, utils.MetaAny): {
			"RL4": struct{}{},
			"RL5": struct{}{},
		},
	}
	if err := dataManager.SetIndexes(utils.CacheResourceFilterIndexes,
		"cgrates.org", idxes, false, utils.NonTransactional); err != nil {
		t.Error(err)
	}
}

func testITGetFilterIndexes(t *testing.T) {
	eIdxes := map[string]utils.StringSet{
		"*string:Account:1001": {
			"RL1": struct{}{},
		},
		"*string:Account:1002": {
			"RL1": struct{}{},
			"RL2": struct{}{},
		},
		"*string:Account:dan": {
			"RL2": struct{}{},
		},
		"*string:Subject:dan": {
			"RL2": struct{}{},
			"RL3": struct{}{},
		},
		utils.ConcatenatedKey(utils.MetaNone, utils.MetaAny, utils.MetaAny): {
			"RL4": struct{}{},
			"RL5": struct{}{},
		},
	}
	expectedsbjDan := map[string]utils.StringSet{
		"*string:Subject:dan": {
			"RL2": struct{}{},
			"RL3": struct{}{},
		},
	}

	if exsbjDan, err := dataManager.GetIndexes(
		utils.CacheResourceFilterIndexes, "cgrates.org",
		utils.ConcatenatedKey(utils.MetaString, "Subject", "dan"),
		false, false); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedsbjDan, exsbjDan) {
		t.Errorf("Expecting: %+v, received: %+v", expectedsbjDan, exsbjDan)
	}
	if rcv, err := dataManager.GetIndexes(
		utils.CacheResourceFilterIndexes,
		"cgrates.org", utils.EmptyString,
		false, false); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eIdxes, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", eIdxes, rcv)
	}
	//invalid tnt:context or index key
	if _, err := dataManager.GetIndexes(
		"unknown_key", "unkonwn_tenant",
		utils.EmptyString, false, false); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	}
}

func testITMatchFilterIndex(t *testing.T) {
	eMp := map[string]utils.StringSet{
		"*string:Account:1002": {
			"RL1": struct{}{},
			"RL2": struct{}{},
		},
	}
	if rcvMp, err := dataManager.GetIndexes(
		utils.CacheResourceFilterIndexes, "cgrates.org",
		utils.ConcatenatedKey(utils.MetaString, "Account", "1002"),
		false, true); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eMp, rcvMp) {
		t.Errorf("Expecting: %+v, received: %+v", eMp, rcvMp)
	}

	//invalid tnt:context or index key
	if _, err := dataManager.GetIndexes(
		utils.CacheResourceFilterIndexes, "cgrates.org",
		utils.ConcatenatedKey(utils.MetaString, "NonexistentField", "1002"),
		true, true); err == nil ||
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
				Type:    utils.MetaString,
				Element: "~*req.EventType",
				Values:  []string{"Event1", "Event2"},
			},
		},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			ExpiryTime:     time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
	}
	if err := dataManager.SetFilter(fp, true); err != nil {
		t.Error(err)
	}
	th := &ThresholdProfile{
		Tenant:             "cgrates.org",
		ID:                 "THD_Test",
		ActivationInterval: &utils.ActivationInterval{},
		FilterIDs:          []string{"Filter1"},
		MaxHits:            12,
		MinSleep:           0,
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
		MinSleep:           0,
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
	eIdxes := map[string]utils.StringSet{
		"*string:*req.EventType:Event1": {
			"THD_Test":  struct{}{},
			"THD_Test2": struct{}{},
		},
		"*string:*req.EventType:Event2": {
			"THD_Test":  struct{}{},
			"THD_Test2": struct{}{},
		},
	}
	if rcvIdx, err := dataManager.GetIndexes(
		utils.CacheThresholdFilterIndexes, th.Tenant,
		utils.EmptyString, false, false); err != nil {
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
				Type:    utils.MetaString,
				Element: "~*req.Account",
				Values:  []string{"1001", "1002"},
			},
		},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			ExpiryTime:     time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
	}
	if err := dataManager.SetFilter(fp2, true); err != nil {
		t.Error(err)
	}
	cloneTh1 := new(ThresholdProfile)
	*cloneTh1 = *th
	cloneTh1.FilterIDs = []string{"Filter2"}
	if err := dataManager.SetThresholdProfile(cloneTh1, true); err != nil {
		t.Error(err)
	}
	eIdxes = map[string]utils.StringSet{
		"*string:*req.Account:1001": {
			"THD_Test": struct{}{},
		},
		"*string:*req.Account:1002": {
			"THD_Test": struct{}{},
		},
		"*string:*req.EventType:Event1": {
			"THD_Test2": struct{}{},
		},
		"*string:*req.EventType:Event2": {
			"THD_Test2": struct{}{},
		},
	}
	if rcvIdx, err := dataManager.GetIndexes(
		utils.CacheThresholdFilterIndexes, th.Tenant,
		utils.EmptyString, false, false); err != nil {
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
				Type:    utils.MetaString,
				Element: "~*req.Destination",
				Values:  []string{"10", "20"},
			},
		},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			ExpiryTime:     time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
	}
	if err := dataManager.SetFilter(fp3, true); err != nil {
		t.Error(err)
	}

	clone2Th1 := new(ThresholdProfile)
	*clone2Th1 = *th
	clone2Th1.FilterIDs = []string{"Filter1", "Filter3"}
	if err := dataManager.SetThresholdProfile(clone2Th1, true); err != nil {
		t.Error(err)
	}
	eIdxes = map[string]utils.StringSet{
		"*string:*req.Destination:10": {
			"THD_Test": struct{}{},
		},
		"*string:*req.Destination:20": {
			"THD_Test": struct{}{},
		},
		"*string:*req.EventType:Event1": {
			"THD_Test":  struct{}{},
			"THD_Test2": struct{}{},
		},
		"*string:*req.EventType:Event2": {
			"THD_Test":  struct{}{},
			"THD_Test2": struct{}{},
		},
	}
	if rcvIdx, err := dataManager.GetIndexes(
		utils.CacheThresholdFilterIndexes, th.Tenant,
		utils.EmptyString, false, false); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eIdxes, rcvIdx) {
		t.Errorf("Expecting %+v, received: %+v", utils.ToJSON(eIdxes), utils.ToJSON(rcvIdx))
	}

	//replace old filter with two different filters
	fp3 = &Filter{
		Tenant: "cgrates.org",
		ID:     "Filter3",
		Rules: []*FilterRule{
			{
				Element: "~*req.Destination",
				Type:    utils.MetaString,
				Values:  []string{"30", "50"},
			},
		},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			ExpiryTime:     time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
	}
	if err := dataManager.SetFilter(fp3, true); err != nil {
		t.Error(err)
	}

	eIdxes = map[string]utils.StringSet{
		utils.CacheThresholdFilterIndexes: {
			"THD_Test": struct{}{},
		},
	}
	if rcvIdx, err := dataManager.GetIndexes(
		utils.CacheReverseFilterIndexes, "cgrates.org:Filter3",
		utils.EmptyString, false, false); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eIdxes, rcvIdx) {
		t.Errorf("Expecting %+v, received: %+v", utils.ToJSON(eIdxes), utils.ToJSON(rcvIdx))
	}

	eIdxes = map[string]utils.StringSet{
		"*string:*req.Destination:30": {
			"THD_Test": struct{}{},
		},
		"*string:*req.Destination:50": {
			"THD_Test": struct{}{},
		},
		"*string:*req.EventType:Event1": {
			"THD_Test":  struct{}{},
			"THD_Test2": struct{}{},
		},
		"*string:*req.EventType:Event2": {
			"THD_Test":  struct{}{},
			"THD_Test2": struct{}{},
		},
	}
	if rcvIdx, err := dataManager.GetIndexes(
		utils.CacheThresholdFilterIndexes, th.Tenant,
		utils.EmptyString, false, false); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eIdxes, rcvIdx) {
		t.Errorf("Expecting %+v, received: %+v", utils.ToJSON(eIdxes), utils.ToJSON(rcvIdx))
	}

	//remove thresholds
	if err := dataManager.RemoveThresholdProfile(th.Tenant,
		th.ID, true); err != nil {
		t.Error(err)
	}
	if err := dataManager.RemoveThresholdProfile(th2.Tenant,
		th2.ID, true); err != nil {
		t.Error(err)
	}
	if _, err := dataManager.GetIndexes(
		utils.CacheThresholdFilterIndexes, th.Tenant,
		utils.EmptyString, false, false); err != utils.ErrNotFound {
		t.Error(err)
	}
	if _, err := dataManager.GetIndexes(
		utils.CacheReverseFilterIndexes, "cgrates.org:Filter3",
		utils.EmptyString, false, false); err != utils.ErrNotFound {
		t.Error(err)
	}
}

func testITTestAttributeProfileFilterIndexes(t *testing.T) {
	fp := &Filter{
		Tenant: "cgrates.org",
		ID:     "AttrFilter",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaString,
				Element: "~*req.EventType",
				Values:  []string{"Event1", "Event2"},
			},
		},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			ExpiryTime:     time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
	}
	if err := dataManager.SetFilter(fp, true); err != nil {
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
				Path:  "FN1",
				Value: config.NewRSRParsersMustCompile("Val1", utils.InfieldSep),
			},
		},
		Weight: 20,
	}
	//Set AttributeProfile with 2 contexts (con1 , con2)
	if err := dataManager.SetAttributeProfile(attrProfile, true); err != nil {
		t.Error(err)
	}
	eIdxes := map[string]utils.StringSet{
		"*string:*req.EventType:Event1": {
			"AttrPrf": struct{}{},
		},
		"*string:*req.EventType:Event2": {
			"AttrPrf": struct{}{},
		},
	}
	for _, ctx := range attrProfile.Contexts {
		if rcvIdx, err := dataManager.GetIndexes(
			utils.CacheAttributeFilterIndexes,
			utils.ConcatenatedKey(attrProfile.Tenant, ctx),
			utils.EmptyString, false, false); err != nil {
			t.Error(err)
		} else if !reflect.DeepEqual(eIdxes, rcvIdx) {
			t.Errorf("Expecting %+v, received: %+v", eIdxes, rcvIdx)
		}
	}
	//Set AttributeProfile with 1 new context (con3)
	attrProfile = &AttributeProfile{ // recreate the profile because if we test on internal
		Tenant:    "cgrates.org", // each update on the original item will update the item from DB
		ID:        "AttrPrf",
		FilterIDs: []string{"AttrFilter"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
		Contexts: []string{"con3"},
		Attributes: []*Attribute{
			{
				Path:  "FN1",
				Value: config.NewRSRParsersMustCompile("Val1", utils.InfieldSep),
			},
		},
		Weight: 20,
	}
	if err := dataManager.SetAttributeProfile(attrProfile, true); err != nil {
		t.Error(err)
	}
	//check indexes with the new context (con3)
	if rcvIdx, err := dataManager.GetIndexes(
		utils.CacheAttributeFilterIndexes,
		utils.ConcatenatedKey(attrProfile.Tenant, "con3"),
		utils.EmptyString, false, false); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eIdxes, rcvIdx) {
		t.Errorf("Expecting %+v, received: %+v", eIdxes, rcvIdx)
	}
	//check if old contexts was delete
	for _, ctx := range []string{"con1", "con2"} {
		if _, err = dataManager.GetIndexes(
			utils.CacheAttributeFilterIndexes,
			utils.ConcatenatedKey(attrProfile.Tenant, ctx),
			utils.EmptyString, false, false); err == nil ||
			err != utils.ErrNotFound {
			t.Error(err)
		}
	}

	fp = &Filter{
		Tenant: "cgrates.org",
		ID:     "AttrFilter",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaString,
				Element: "~*req.EventType",
				Values:  []string{"Event3", "~*req.Event4"},
			},
		},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			ExpiryTime:     time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
	}
	if err := dataManager.SetFilter(fp, true); err != nil {
		t.Error(err)
	}
	eIdxes = map[string]utils.StringSet{
		"*string:*req.EventType:Event3": {
			"AttrPrf": struct{}{},
		},
	}
	for _, ctx := range attrProfile.Contexts {
		if rcvIdx, err := dataManager.GetIndexes(
			utils.CacheAttributeFilterIndexes,
			utils.ConcatenatedKey(attrProfile.Tenant, ctx),
			utils.EmptyString, false, false); err != nil {
			t.Error(err)
		} else if !reflect.DeepEqual(eIdxes, rcvIdx) {
			t.Errorf("Expecting %+v, received: %+v", utils.ToJSON(eIdxes), rcvIdx)
		}
	}

	eIdxes = map[string]utils.StringSet{
		"*attribute_filter_indexes": {
			"AttrPrf": struct{}{},
		},
	}
	if rcvIdx, err := dataManager.GetIndexes(
		utils.CacheReverseFilterIndexes,
		fp.TenantID(),
		utils.EmptyString, false, false); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eIdxes, rcvIdx) {
		t.Errorf("Expecting %+v, received: %+v", eIdxes, rcvIdx)
	}

	if err := dataManager.RemoveAttributeProfile(attrProfile.Tenant,
		attrProfile.ID, true); err != nil {
		t.Error(err)
	}
	//check if index is removed
	if _, err := dataManager.GetIndexes(
		utils.CacheAttributeFilterIndexes,
		utils.ConcatenatedKey("cgrates.org", "con3"),
		utils.MetaString, false, false); err != nil && err != utils.ErrNotFound {
		t.Error(err)
	}

	if _, err := dataManager.GetIndexes(
		utils.CacheReverseFilterIndexes,
		fp.TenantID(),
		utils.EmptyString, false, false); err != utils.ErrNotFound {
		t.Error(err)
	}
}

func testITTestAttributeProfileFilterIndexes2(t *testing.T) {
	fp := &Filter{
		Tenant: "cgrates.org",
		ID:     "AttrFilter",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaString,
				Element: "EventType",
				Values:  []string{"~*req.Event1", "~*req.Event2"},
			},
		},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			ExpiryTime:     time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
	}
	if err := dataManager.SetFilter(fp, true); err != nil {
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
				Path:  "FN1",
				Value: config.NewRSRParsersMustCompile("Val1", utils.InfieldSep),
			},
		},
		Weight: 20,
	}
	//Set AttributeProfile with 2 contexts ( con1 , con2)
	if err := dataManager.SetAttributeProfile(attrProfile, true); err != nil {
		t.Error(err)
	}
	eIdxes := map[string]utils.StringSet{
		"*string:*req.Event1:EventType": {
			"AttrPrf": struct{}{},
		},
		"*string:*req.Event2:EventType": {
			"AttrPrf": struct{}{},
		},
	}
	for _, ctx := range attrProfile.Contexts {
		if rcvIdx, err := dataManager.GetIndexes(
			utils.CacheAttributeFilterIndexes,
			utils.ConcatenatedKey(attrProfile.Tenant, ctx),
			utils.EmptyString, false, false); err != nil {
			t.Error(err)
		} else if !reflect.DeepEqual(eIdxes, rcvIdx) {
			t.Errorf("Expecting %+v, received: %+v", eIdxes, rcvIdx)
		}
	}
	//Set AttributeProfile with 1 new context (con3)
	attrProfile = &AttributeProfile{ // recreate the profile because if we test on internal
		Tenant:    "cgrates.org", // each update on the original item will update the item from DB
		ID:        "AttrPrf",
		FilterIDs: []string{"AttrFilter"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
		Contexts: []string{"con3"},
		Attributes: []*Attribute{
			{
				Path:  "FN1",
				Value: config.NewRSRParsersMustCompile("Val1", utils.InfieldSep),
			},
		},
		Weight: 20,
	}
	if err := dataManager.SetAttributeProfile(attrProfile, true); err != nil {
		t.Error(err)
	}
	//check indexes with the new context (con3)
	if rcvIdx, err := dataManager.GetIndexes(
		utils.CacheAttributeFilterIndexes,
		utils.ConcatenatedKey(attrProfile.Tenant, "con3"),
		utils.EmptyString, false, false); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eIdxes, rcvIdx) {
		t.Errorf("Expecting %+v, received: %+v", eIdxes, rcvIdx)
	}
	//check if old contexts was delete
	for _, ctx := range []string{"con1", "con2"} {
		if _, err = dataManager.GetIndexes(
			utils.CacheAttributeFilterIndexes,
			utils.ConcatenatedKey(attrProfile.Tenant, ctx),
			utils.EmptyString, false, false); err == nil ||
			err != utils.ErrNotFound {
			t.Error(err)
		}
	}

	fp = &Filter{
		Tenant: "cgrates.org",
		ID:     "AttrFilter",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaString,
				Element: "~*req.EventType",
				Values:  []string{"Event3", "~*req.Event4"},
			},
		},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			ExpiryTime:     time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
	}
	if err := dataManager.SetFilter(fp, true); err != nil {
		t.Error(err)
	}
	eIdxes = map[string]utils.StringSet{
		"*string:*req.EventType:Event3": {
			"AttrPrf": struct{}{},
		},
	}
	for _, ctx := range attrProfile.Contexts {
		if rcvIdx, err := dataManager.GetIndexes(
			utils.CacheAttributeFilterIndexes,
			utils.ConcatenatedKey(attrProfile.Tenant, ctx),
			utils.EmptyString, false, false); err != nil {
			t.Error(err)
		} else if !reflect.DeepEqual(eIdxes, rcvIdx) {
			t.Errorf("Expecting %+v, received: %+v", utils.ToJSON(eIdxes), rcvIdx)
		}
	}

	eIdxes = map[string]utils.StringSet{
		"*attribute_filter_indexes": {
			"AttrPrf": struct{}{},
		},
	}
	if rcvIdx, err := dataManager.GetIndexes(
		utils.CacheReverseFilterIndexes,
		fp.TenantID(),
		utils.EmptyString, false, false); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eIdxes, rcvIdx) {
		t.Errorf("Expecting %+v, received: %+v", eIdxes, rcvIdx)
	}

	if err := dataManager.RemoveAttributeProfile(attrProfile.Tenant,
		attrProfile.ID, true); err != nil {
		t.Error(err)
	}
	//check if index is removed
	if _, err := dataManager.GetIndexes(
		utils.CacheAttributeFilterIndexes,
		utils.ConcatenatedKey("cgrates.org", "con3"),
		utils.MetaString, false, false); err != nil && err != utils.ErrNotFound {
		t.Error(err)
	}

	if _, err := dataManager.GetIndexes(
		utils.CacheReverseFilterIndexes,
		fp.TenantID(),
		utils.EmptyString, false, false); err != utils.ErrNotFound {
		t.Error(err)
	}
}

func testITTestThresholdInlineFilterIndexing(t *testing.T) {
	fp := &Filter{
		Tenant: "cgrates.org",
		ID:     "Filter1",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaString,
				Element: "~*req.EventType",
				Values:  []string{"Event1", "Event2"},
			},
		},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			ExpiryTime:     time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
	}
	if err := dataManager.SetFilter(fp, true); err != nil {
		t.Error(err)
	}
	th := &ThresholdProfile{
		Tenant:             "cgrates.org",
		ID:                 "THD_Test",
		ActivationInterval: &utils.ActivationInterval{},
		FilterIDs:          []string{"Filter1"},
		MaxHits:            12,
		MinSleep:           0,
		Blocker:            true,
		Weight:             1.4,
		ActionIDs:          []string{},
	}

	if err := dataManager.SetThresholdProfile(th, true); err != nil {
		t.Error(err)
	}
	eIdxes := map[string]utils.StringSet{
		"*string:*req.EventType:Event1": {
			"THD_Test": struct{}{},
		},
		"*string:*req.EventType:Event2": {
			"THD_Test": struct{}{},
		},
	}
	if rcvIdx, err := dataManager.GetIndexes(
		utils.CacheThresholdFilterIndexes, th.Tenant,
		utils.EmptyString, false, false); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eIdxes, rcvIdx) {
		t.Errorf("Expecting %+v, received: %+v", eIdxes, rcvIdx)
	}
	//Add an InlineFilter
	th = &ThresholdProfile{ // recreate the profile because if we test on internal
		Tenant:             "cgrates.org", // each update on the original item will update the item from DB
		ID:                 "THD_Test",
		ActivationInterval: &utils.ActivationInterval{},
		FilterIDs:          []string{"Filter1", "*string:~*req.Account:1001"},
		MaxHits:            12,
		MinSleep:           0,
		Blocker:            true,
		Weight:             1.4,
		ActionIDs:          []string{},
	}
	if err := dataManager.SetThresholdProfile(th, true); err != nil {
		t.Error(err)
	}
	eIdxes = map[string]utils.StringSet{
		"*string:*req.Account:1001": {
			"THD_Test": struct{}{},
		},
		"*string:*req.EventType:Event1": {
			"THD_Test": struct{}{},
		},
		"*string:*req.EventType:Event2": {
			"THD_Test": struct{}{},
		},
	}
	if rcvIdx, err := dataManager.GetIndexes(
		utils.CacheThresholdFilterIndexes, th.Tenant,
		utils.EmptyString, false, false); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eIdxes, rcvIdx) {
		t.Errorf("Expecting %+v, received: %+v", utils.ToJSON(eIdxes), utils.ToJSON(rcvIdx))
	}
	//remove threshold
	if err := dataManager.RemoveThresholdProfile(th.Tenant,
		th.ID, true); err != nil {
		t.Error(err)
	}
	if _, err := dataManager.GetIndexes(
		utils.CacheThresholdFilterIndexes, th.Tenant,
		utils.EmptyString, false, false); err != utils.ErrNotFound {
		t.Error(err)
	}
}

func testITTestStoreFilterIndexesWithTransID(t *testing.T) {
	idxes := map[string]utils.StringSet{
		"*string:Account:1001": {
			"RL1": struct{}{},
		},
		"*string:Account:1002": {
			"RL1": struct{}{},
			"RL2": struct{}{},
		},
		"*string:Account:dan": {
			"RL2": struct{}{},
		},
		"*string:Subject:dan": {
			"RL2": struct{}{},
			"RL3": struct{}{},
		},
		utils.ConcatenatedKey(utils.MetaNone,
			utils.MetaAny, utils.MetaAny): {
			"RL4": struct{}{},
			"RL5": struct{}{},
		},
	}
	if err := dataManager.SetIndexes(utils.CacheResourceFilterIndexes,
		"cgrates.org", idxes, false, "transaction1"); err != nil {
		t.Error(err)
	}

	//commit transaction
	if err := dataManager.SetIndexes(utils.CacheResourceFilterIndexes,
		"cgrates.org", idxes, true, "transaction1"); err != nil {
		t.Error(err)
	}
	eIdx := map[string]utils.StringSet{
		"*string:Account:1001": {
			"RL1": struct{}{},
		},
		"*string:Account:1002": {
			"RL1": struct{}{},
			"RL2": struct{}{},
		},
		"*string:Account:dan": {
			"RL2": struct{}{},
		},
		"*string:Subject:dan": {
			"RL2": struct{}{},
			"RL3": struct{}{},
		},
		utils.ConcatenatedKey(utils.MetaNone,
			utils.MetaAny, utils.MetaAny): {
			"RL4": struct{}{},
			"RL5": struct{}{},
		},
	}

	//verify new key and check if data was moved
	if rcv, err := dataManager.GetIndexes(
		utils.CacheResourceFilterIndexes, "cgrates.org",
		utils.EmptyString, false, false); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eIdx, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", eIdx, rcv)
	}
}

func testITResourceProfileIndexes(t *testing.T) {
	fltr1 := &Filter{
		Tenant: "cgrates.org",
		ID:     "RES_FLTR1",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaString,
				Element: "~*req.Destinations",
				Values:  []string{"DEST_RES1", "~*opts.DynamicValue", "DEST_RES2"},
			},
		},
	}
	if err := dataManager.SetFilter(fltr1, true); err != nil {
		t.Error(err)
	}

	resPref1 := &ResourceProfile{
		Tenant:    "cgrates.org",
		ID:        "RES_PRF1",
		FilterIDs: []string{"FIRST", "RES_FLTR1", "*string:~*req.Account:DAN"},
		Limit:     23,
		Stored:    true,
	}
	resPref2 := &ResourceProfile{
		Tenant:    "cgrates.org",
		ID:        "RES_PRF2",
		FilterIDs: []string{"RES_FLTR1"},
		Limit:     23,
		Stored:    true,
	}

	expected := "broken reference to filter: <FIRST> for item with ID: cgrates.org:RES_PRF1"
	if err := dataManager.SetResourceProfile(resPref1, true); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}

	resPref1.FilterIDs = []string{"RES_FLTR1", "*string:~*req.Account:DAN"}
	if err := dataManager.SetResourceProfile(resPref1, true); err != nil {
		t.Error(err)
	}
	if err := dataManager.SetResourceProfile(resPref2, true); err != nil {
		t.Error(err)
	}

	eIdxes := map[string]utils.StringSet{
		"*string:*req.Account:DAN": {
			"RES_PRF1": struct{}{},
		},
		"*string:*req.Destinations:DEST_RES1": {
			"RES_PRF1": struct{}{},
			"RES_PRF2": struct{}{},
		},
		"*string:*req.Destinations:DEST_RES2": {
			"RES_PRF1": struct{}{},
			"RES_PRF2": struct{}{},
		},
	}
	if rcvIDx, err := dataManager.GetIndexes(utils.CacheResourceFilterIndexes,
		"cgrates.org", utils.EmptyString, false, false); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcvIDx, eIdxes) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(eIdxes), utils.ToJSON(rcvIDx))
	}

	//we will change the rules of our filter, to be updated
	fltr1 = &Filter{
		Tenant: "cgrates.org",
		ID:     "RES_FLTR1",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaString,
				Element: "~*req.Usage",
				Values:  []string{"10m"},
			},
		},
	}
	if err := dataManager.SetFilter(fltr1, true); err != nil {
		t.Error(err)
	}
	resPref1.ID = "RES_PRF_CHANGED1"
	resPref2.ID = "RES_PRF_CHANGED2"
	if err := dataManager.SetResourceProfile(resPref1, true); err != nil {
		t.Error(err)
	}
	if err := dataManager.SetResourceProfile(resPref2, true); err != nil {
		t.Error(err)
	}

	eIdxes = map[string]utils.StringSet{
		"*string:*req.Account:DAN": {
			"RES_PRF1":         struct{}{},
			"RES_PRF_CHANGED1": struct{}{},
		},
		"*string:*req.Usage:10m": {
			"RES_PRF1":         struct{}{},
			"RES_PRF2":         struct{}{},
			"RES_PRF_CHANGED1": struct{}{},
			"RES_PRF_CHANGED2": struct{}{},
		},
	}
	if rcvIDx, err := dataManager.GetIndexes(utils.CacheResourceFilterIndexes,
		"cgrates.org", utils.EmptyString, false, false); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcvIDx, eIdxes) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(eIdxes), utils.ToJSON(rcvIDx))
	}

	//here we will check the reverse indexing
	eIdxes = map[string]utils.StringSet{
		utils.CacheResourceFilterIndexes: {
			"RES_PRF1":         struct{}{},
			"RES_PRF2":         struct{}{},
			"RES_PRF_CHANGED1": struct{}{},
			"RES_PRF_CHANGED2": struct{}{},
		},
	}
	if rcvIDx, err := dataManager.GetIndexes(utils.CacheReverseFilterIndexes,
		"cgrates.org:RES_FLTR1", utils.EmptyString, false, false); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcvIDx, eIdxes) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(eIdxes), utils.ToJSON(rcvIDx))
	}

	//as we updated our filter, the old one is deleted
	if _, err := dataManager.GetIndexes(utils.CacheResourceFilterIndexes,
		"cgrates.org", "*string:*req.Destinations:DEST_RES1", false, false); err == nil || err != utils.ErrNotFound {
		t.Errorf("Expected %+v, receive %+v", utils.ErrNotFound, err)
	}
}

func testITStatQueueProfileIndexes(t *testing.T) {
	fltr1 := &Filter{
		Tenant: "cgrates.org",
		ID:     "SQUEUE1",
		Rules:  []*FilterRule{{Type: utils.MetaString, Element: "~*req.Usage", Values: []string{"10m"}}},
	}
	fltr2 := &Filter{
		Tenant: "cgrates.org",
		ID:     "SQUEUE2",
		Rules:  []*FilterRule{{Type: utils.MetaString, Element: "~*req.Destination", Values: []string{"~*req.Owner", "Dan1", "Dan2"}}},
	}
	if err := dataManager.SetFilter(fltr1, true); err != nil {
		t.Error(err)
	}
	if err := dataManager.SetFilter(fltr2, true); err != nil {
		t.Error(err)
	}

	statQueue1 := &StatQueueProfile{
		Tenant:    "cgrates.org",
		ID:        "SQUEUE_PRF1",
		FilterIDs: []string{"SQUEUE1", "*string:~*opts.ToR:*data"},
		TTL:       time.Minute,
	}
	statQueue2 := &StatQueueProfile{
		Tenant:    "cgrates.org",
		ID:        "SQUEUE_PRF2",
		FilterIDs: []string{"SQUEUE2", "*string:~*opts.ToR:*voice"},
		TTL:       time.Minute,
	}
	statQueue3 := &StatQueueProfile{
		Tenant:    "cgrates.org",
		ID:        "SQUEUE_PRF3",
		FilterIDs: []string{"SQUEUE2", "SQUEUE1", "*string:~*opts.ToR:~*req.Usage"},
		TTL:       time.Minute,
	}
	if err := dataManager.SetStatQueueProfile(statQueue1, true); err != nil {
		t.Error(err)
	} else if err := dataManager.SetStatQueueProfile(statQueue2, true); err != nil {
		t.Error(err)
	} else if err := dataManager.SetStatQueueProfile(statQueue3, true); err != nil {
		t.Error(err)
	}
	eIdxes := map[string]utils.StringSet{
		"*string:*req.Usage:10m": {
			"SQUEUE_PRF1": struct{}{},
			"SQUEUE_PRF3": struct{}{},
		},
		"*string:*req.Destination:Dan1": {
			"SQUEUE_PRF2": struct{}{},
			"SQUEUE_PRF3": struct{}{},
		},
		"*string:*req.Destination:Dan2": {
			"SQUEUE_PRF2": struct{}{},
			"SQUEUE_PRF3": struct{}{},
		},
		"*string:*opts.ToR:*voice": {
			"SQUEUE_PRF2": struct{}{},
		},
		"*string:*opts.ToR:*data": {
			"SQUEUE_PRF1": struct{}{},
		},
	}
	if rcvIDx, err := dataManager.GetIndexes(utils.CacheStatFilterIndexes,
		"cgrates.org", utils.EmptyString, false, false); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcvIDx, eIdxes) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(eIdxes), utils.ToJSON(rcvIDx))
	}

	//here we will check the reverse indexing
	eIdxes = map[string]utils.StringSet{
		utils.CacheStatFilterIndexes: {
			"SQUEUE_PRF1": struct{}{},
			"SQUEUE_PRF3": struct{}{},
		},
	}
	if rcvIDx, err := dataManager.GetIndexes(utils.CacheReverseFilterIndexes,
		"cgrates.org:SQUEUE1", utils.EmptyString, false, false); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcvIDx, eIdxes) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(eIdxes), utils.ToJSON(rcvIDx))
	}

	eIdxes = map[string]utils.StringSet{
		utils.CacheStatFilterIndexes: {
			"SQUEUE_PRF2": struct{}{},
			"SQUEUE_PRF3": struct{}{},
		},
	}
	if rcvIDx, err := dataManager.GetIndexes(utils.CacheReverseFilterIndexes,
		"cgrates.org:SQUEUE2", utils.EmptyString, false, false); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcvIDx, eIdxes) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(eIdxes), utils.ToJSON(rcvIDx))
	}

	//invalid tnt:context or index key
	eIdxes = nil
	if rcvIDx, err := dataManager.GetIndexes(utils.CacheStatFilterIndexes,
		"cgrates.org", "*string:~*opts.ToR:~*req.Usage", false, false); err == nil || err != utils.ErrNotFound {
		t.Errorf("Expected %+v, receive %+v", utils.ErrNotFound, err)
	} else if !reflect.DeepEqual(rcvIDx, eIdxes) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(eIdxes), utils.ToJSON(rcvIDx))
	}
}

func testITChargerProfileIndexes(t *testing.T) {
	fltr1 := &Filter{
		Tenant: "cgrates.org",
		ID:     "CHARGER_FLTR",
		Rules:  []*FilterRule{{Type: utils.MetaString, Element: "~*req.Usage", Values: []string{"10m", "20m", "~*req.Usage"}}},
	}
	if err := dataManager.SetFilter(fltr1, true); err != nil {
		t.Error(err)
	}

	chrgr1 := &ChargerProfile{
		Tenant:    "cgrates.org",
		ID:        "CHARGER_PRF1",
		FilterIDs: []string{"CHARGER_FLTR"},
		Weight:    10,
	}
	chrgr2 := &ChargerProfile{
		Tenant:    "cgrates.org",
		ID:        "CHARGER_PRF2",
		FilterIDs: []string{"CHARGER_FLTR", "*string:~*req.Usage:~*req.Debited"},
		Weight:    10,
	}
	if err := dataManager.SetChargerProfile(chrgr1, true); err != nil {
		t.Error(err)
	} else if err := dataManager.SetChargerProfile(chrgr2, true); err != nil {
		t.Error(err)
	}

	expIdx := map[string]utils.StringSet{
		"*string:*req.Usage:10m": {
			"CHARGER_PRF1": struct{}{},
			"CHARGER_PRF2": struct{}{},
		},
		"*string:*req.Usage:20m": {
			"CHARGER_PRF1": struct{}{},
			"CHARGER_PRF2": struct{}{},
		},
	}
	if rcvIDx, err := dataManager.GetIndexes(utils.CacheChargerFilterIndexes,
		"cgrates.org", utils.EmptyString, false, false); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcvIDx, expIdx) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(expIdx), utils.ToJSON(rcvIDx))
	}

	//update the filter and the chargerProfiles for matching
	fltr1 = &Filter{
		Tenant: "cgrates.org",
		ID:     "CHARGER_FLTR",
		Rules:  []*FilterRule{{Type: utils.MetaString, Element: "~*req.CGRID", Values: []string{"~*req.Usage", "DAN1"}}},
	}
	if err := dataManager.SetFilter(fltr1, true); err != nil {
		t.Error(err)
	}
	chrgr1.ID = "CHANGED_CHARGER_PRF1"
	chrgr2.ID = "CHANGED_CHARGER_PRF2"
	if err := dataManager.SetChargerProfile(chrgr1, true); err != nil {
		t.Error(err)
	} else if err := dataManager.SetChargerProfile(chrgr2, true); err != nil {
		t.Error(err)
	}

	expIdx = map[string]utils.StringSet{
		"*string:*req.CGRID:DAN1": {
			"CHARGER_PRF1":         struct{}{},
			"CHARGER_PRF2":         struct{}{},
			"CHANGED_CHARGER_PRF1": struct{}{},
			"CHANGED_CHARGER_PRF2": struct{}{},
		},
	}
	if rcvIDx, err := dataManager.GetIndexes(utils.CacheChargerFilterIndexes,
		"cgrates.org", utils.EmptyString, false, false); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcvIDx, expIdx) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(expIdx), utils.ToJSON(rcvIDx))
	}

	//here we will check the reverse indexing
	expIdx = map[string]utils.StringSet{
		utils.CacheChargerFilterIndexes: {
			"CHARGER_PRF1":         struct{}{},
			"CHARGER_PRF2":         struct{}{},
			"CHANGED_CHARGER_PRF1": struct{}{},
			"CHANGED_CHARGER_PRF2": struct{}{},
		},
	}
	if rcvIDx, err := dataManager.GetIndexes(utils.CacheReverseFilterIndexes,
		"cgrates.org:CHARGER_FLTR", utils.EmptyString, false, false); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcvIDx, expIdx) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(expIdx), utils.ToJSON(rcvIDx))
	}

	//the old filter is deleted
	expIdx = nil
	if rcvIDx, err := dataManager.GetIndexes(utils.CacheChargerFilterIndexes,
		"cgrates.org", "*string:*req.Usage", false, false); err == nil || err != utils.ErrNotFound {
		t.Errorf("Expected %+v, received %+v", utils.Error, err)
	} else if !reflect.DeepEqual(rcvIDx, expIdx) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(expIdx), utils.ToJSON(rcvIDx))
	}
}

func testITDispatcherProfileIndexes(t *testing.T) {
	fltr1 := &Filter{
		Tenant: "cgrates.org",
		ID:     "DISPATCHER_FLTR1",
		Rules:  []*FilterRule{{Type: utils.MetaString, Element: "~*req.Destination", Values: []string{"ACC1", "ACC2", "~*req.Account"}}},
	}
	fltr2 := &Filter{
		Tenant: "cgrates.org",
		ID:     "DISPATCHER_FLTR2",
		Rules:  []*FilterRule{{Type: utils.MetaString, Element: "10m", Values: []string{"USAGE", "~*opts.Debited", "~*req.Usage", "~*opts.Usage"}}},
	}
	if err := dataManager.SetFilter(fltr1, true); err != nil {
		t.Error(err)
	} else if err := dataManager.SetFilter(fltr2, true); err != nil {
		t.Error(err)
	}

	dspPrf1 := &DispatcherProfile{
		Tenant:     "cgrates.org",
		ID:         "DISPATCHER_PRF1",
		Subsystems: []string{"thresholds"},
		FilterIDs:  []string{"DISPATCHER_FLTR1"},
	}
	dspPrf2 := &DispatcherProfile{
		Tenant:     "cgrates.org",
		ID:         "DISPATCHER_PRF2",
		Subsystems: []string{"thresholds"},
		FilterIDs:  []string{"DISPATCHER_FLTR2", "*prefix:23:~*req.Destination"},
	}
	if err := dataManager.SetDispatcherProfile(dspPrf1, true); err != nil {
		t.Error(err)
	}
	if err := dataManager.SetDispatcherProfile(dspPrf2, true); err != nil {
		t.Error(err)
	}

	expIdx := map[string]utils.StringSet{
		"*string:*req.Destination:ACC1": {
			"DISPATCHER_PRF1": struct{}{},
		},
		"*string:*req.Destination:ACC2": {
			"DISPATCHER_PRF1": struct{}{},
		},
		"*string:*opts.Debited:10m": {
			"DISPATCHER_PRF2": struct{}{},
		},
		"*string:*req.Usage:10m": {
			"DISPATCHER_PRF2": struct{}{},
		},
		"*string:*opts.Usage:10m": {
			"DISPATCHER_PRF2": struct{}{},
		},
		"*prefix:*req.Destination:23": {
			"DISPATCHER_PRF2": struct{}{},
		},
	}
	if rcvIDx, err := dataManager.GetIndexes(utils.CacheDispatcherFilterIndexes,
		"cgrates.org:thresholds", utils.EmptyString, false, false); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcvIDx, expIdx) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(expIdx), utils.ToJSON(rcvIDx))
	}

	//here we will get the reverse indexes
	expIdx = map[string]utils.StringSet{
		utils.CacheDispatcherFilterIndexes: {
			"DISPATCHER_PRF1": struct{}{},
		},
	}
	if rcvIDx, err := dataManager.GetIndexes(utils.CacheReverseFilterIndexes,
		"cgrates.org:DISPATCHER_FLTR1", utils.EmptyString, false, false); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcvIDx, expIdx) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(expIdx), utils.ToJSON(rcvIDx))
	}

	expIdx = map[string]utils.StringSet{
		utils.CacheDispatcherFilterIndexes: {
			"DISPATCHER_PRF2": struct{}{},
		},
	}
	if rcvIDx, err := dataManager.GetIndexes(utils.CacheReverseFilterIndexes,
		"cgrates.org:DISPATCHER_FLTR2", utils.EmptyString, false, false); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcvIDx, expIdx) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(expIdx), utils.ToJSON(rcvIDx))
	}

	//invalid tnt:context or index key
	expIdx = nil
	if rcvIDx, err := dataManager.GetIndexes(utils.CacheDispatcherFilterIndexes,
		"cgrates.org:attributes", utils.EmptyString, false, false); err == nil || err != utils.ErrNotFound {
		t.Errorf("Expectedd %+v, received %+v", utils.ErrNotFound, err)
	} else if !reflect.DeepEqual(rcvIDx, expIdx) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(expIdx), utils.ToJSON(rcvIDx))
	}
}

func testITTestStoreFilterIndexesWithTransID2(t *testing.T) {
	idxes := map[string]utils.StringSet{
		"*string:Event:Event1": {
			"RL1": struct{}{},
		},
		"*string:Event:Event2": {
			"RL1": struct{}{},
			"RL2": struct{}{},
		},
	}
	transID := "transaction1"
	if err := dataManager.SetIndexes(utils.CacheResourceFilterIndexes,
		"cgrates.org", idxes, false, transID); err != nil {
		t.Error(err)
	}
	//commit transaction
	if err := dataManager.SetIndexes(utils.CacheResourceFilterIndexes,
		"cgrates.org", nil, true, transID); err != nil {
		t.Error(err)
	}
	//verify if old key was deleted
	if _, err := dataManager.GetIndexes(
		"tmp_"+utils.CacheResourceFilterIndexes,
		utils.ConcatenatedKey("cgrates.org", transID),
		utils.EmptyString, false, false); err != utils.ErrNotFound {
		t.Error(err)
	}
	//verify new key and check if data was moved
	if rcv, err := dataManager.GetIndexes(
		utils.CacheResourceFilterIndexes,
		"cgrates.org", utils.EmptyString, false, false); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(idxes, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(idxes), utils.ToJSON(rcv))
	}
}

func testITTestIndexingWithEmptyFltrID(t *testing.T) {
	th := &ThresholdProfile{
		Tenant:             "cgrates.org",
		ID:                 "THD_Test",
		ActivationInterval: &utils.ActivationInterval{},
		FilterIDs:          []string{},
		MaxHits:            12,
		MinSleep:           0,
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
		MinSleep:           0,
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
	eIdxes := map[string]utils.StringSet{
		"*none:*any:*any": {
			"THD_Test":  struct{}{},
			"THD_Test2": struct{}{},
		},
	}
	if rcvIdx, err := dataManager.GetIndexes(
		utils.CacheThresholdFilterIndexes, th.Tenant,
		utils.EmptyString, false, false); err != nil {
		t.Error(err)
	} else {
		if !reflect.DeepEqual(eIdxes, rcvIdx) {
			t.Errorf("Expecting %+v, received: %+v", eIdxes, rcvIdx)
		}
	}
	eMp := map[string]utils.StringSet{
		"*none:*any:*any": {
			"THD_Test":  struct{}{},
			"THD_Test2": struct{}{},
		},
	}
	if rcvMp, err := dataManager.GetIndexes(
		utils.CacheThresholdFilterIndexes, th.Tenant,
		utils.ConcatenatedKey(utils.MetaNone, utils.MetaAny, utils.MetaAny),
		true, true); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eMp, rcvMp) {
		t.Errorf("Expecting: %+v, received: %+v", eMp, rcvMp)
	}
}

func testITTestIndexingWithEmptyFltrID2(t *testing.T) {
	splProfile := &RouteProfile{
		Tenant:    "cgrates.org",
		ID:        "SPL_Weight",
		FilterIDs: []string{},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
		Sorting:           "*weight",
		SortingParameters: []string{},
		Routes: []*Route{
			{
				ID:              "supplier1",
				FilterIDs:       []string{""},
				AccountIDs:      []string{""},
				RatingPlanIDs:   []string{""},
				ResourceIDs:     []string{""},
				StatIDs:         []string{""},
				Weight:          10,
				RouteParameters: "",
			},
		},
		Weight: 20,
	}
	splProfile2 := &RouteProfile{
		Tenant:    "cgrates.org",
		ID:        "SPL_Weight2",
		FilterIDs: []string{},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
		Sorting:           "*weight",
		SortingParameters: []string{},
		Routes: []*Route{
			{
				ID:              "supplier1",
				FilterIDs:       []string{""},
				AccountIDs:      []string{""},
				RatingPlanIDs:   []string{""},
				ResourceIDs:     []string{""},
				StatIDs:         []string{""},
				Weight:          10,
				RouteParameters: "",
			},
		},
		Weight: 20,
	}

	if err := dataManager.SetRouteProfile(splProfile, true); err != nil {
		t.Error(err)
	}
	if err := dataManager.SetRouteProfile(splProfile2, true); err != nil {
		t.Error(err)
	}
	eIdxes := map[string]utils.StringSet{
		"*none:*any:*any": {
			"SPL_Weight":  struct{}{},
			"SPL_Weight2": struct{}{},
		},
	}
	if rcvIdx, err := dataManager.GetIndexes(
		utils.CacheRouteFilterIndexes, splProfile.Tenant,
		utils.EmptyString, false, false); err != nil {
		t.Error(err)
	} else {
		if !reflect.DeepEqual(eIdxes, rcvIdx) {
			t.Errorf("Expecting %+v, received: %+v", eIdxes, rcvIdx)
		}
	}
	eMp := map[string]utils.StringSet{
		"*none:*any:*any": {
			"SPL_Weight":  struct{}{},
			"SPL_Weight2": struct{}{},
		},
	}
	if rcvMp, err := dataManager.GetIndexes(
		utils.CacheRouteFilterIndexes, splProfile.Tenant,
		utils.ConcatenatedKey(utils.MetaNone, utils.MetaAny, utils.MetaAny),
		true, true); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eMp, rcvMp) {
		t.Errorf("Expecting: %+v, received: %+v", eMp, rcvMp)
	}

	//make a filter for getting the indexes
	fltr1 := &Filter{
		Tenant: "cgrates.org",
		ID:     "FIRST",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaString,
				Element: "ORG_ID",
				Values:  []string{"~*req.OriginID", "~*opts.CGRID", "DAN"},
			},
		},
	}
	if err := dataManager.SetFilter(fltr1, true); err != nil {
		t.Error(err)
	}

	splProfile.ID = "SPL_WITH_FILTER1"
	splProfile.FilterIDs = []string{"FIRST", "*prefix:~*req.Account:123"}
	splProfile2.ID = "SPL_WITH_FILTER2"
	splProfile2.FilterIDs = []string{"FIRST"}
	if err := dataManager.SetRouteProfile(splProfile, true); err != nil {
		t.Error(err)
	}
	if err := dataManager.SetRouteProfile(splProfile2, true); err != nil {
		t.Error(err)
	}
	expIdx := map[string]utils.StringSet{
		"*none:*any:*any": {
			"SPL_Weight":  struct{}{},
			"SPL_Weight2": struct{}{},
		},
		"*string:*req.OriginID:ORG_ID": {
			"SPL_WITH_FILTER1": struct{}{},
			"SPL_WITH_FILTER2": struct{}{},
		},
		"*string:*opts.CGRID:ORG_ID": {
			"SPL_WITH_FILTER1": struct{}{},
			"SPL_WITH_FILTER2": struct{}{},
		},
		"*prefix:*req.Account:123": {
			"SPL_WITH_FILTER1": struct{}{},
		},
	}
	if rcvIdx, err := dataManager.GetIndexes(utils.CacheRouteFilterIndexes,
		"cgrates.org", utils.EmptyString, false, false); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expIdx, rcvIdx) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(expIdx), utils.ToJSON(rcvIdx))
	}

	expIdx = map[string]utils.StringSet{
		"*string:*opts.CGRID:ORG_ID": {
			"SPL_WITH_FILTER1": struct{}{},
			"SPL_WITH_FILTER2": struct{}{},
		},
	}
	if rcvIdx, err := dataManager.GetIndexes(utils.CacheRouteFilterIndexes,
		"cgrates.org", "*string:*opts.CGRID:ORG_ID", false, false); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expIdx, rcvIdx) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(expIdx), utils.ToJSON(rcvIdx))
	}

	//here we will check the reverse indexing
	expIdx = map[string]utils.StringSet{
		utils.CacheRouteFilterIndexes: {
			"SPL_WITH_FILTER1": struct{}{},
			"SPL_WITH_FILTER2": struct{}{},
		},
	}
	if rcvIdx, err := dataManager.GetIndexes(utils.CacheReverseFilterIndexes,
		"cgrates.org:FIRST", utils.EmptyString, false, false); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expIdx, rcvIdx) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(expIdx), utils.ToJSON(rcvIdx))
	}

	//invalid tnt:context or index key
	if _, err := dataManager.GetIndexes(utils.CacheRouteFilterIndexes,
		"cgrates.org", "*string:DAN:ORG_ID", false, false); err == nil || err != utils.ErrNotFound {
		t.Errorf("Expected %+v, received %+v", utils.ErrNotFound, err)
	}
}

func testITTestIndexingThresholds(t *testing.T) {
	th := &ThresholdProfile{
		Tenant:    "cgrates.org",
		ID:        "TH1",
		FilterIDs: []string{"*string:~*req.Account:1001", "*gt:~*req.Balance:1000"},
		ActionIDs: []string{},
	}
	th2 := &ThresholdProfile{
		Tenant:    "cgrates.org",
		ID:        "TH2",
		FilterIDs: []string{"*string:~*req.Account:1001", "*gt:~*req.Balance:1000"},
		ActionIDs: []string{},
	}
	th3 := &ThresholdProfile{
		Tenant:    "cgrates.org",
		ID:        "TH3",
		FilterIDs: []string{"*string:~*req.Account:1002", "*lt:~*req.Balance:1000"},
		ActionIDs: []string{},
	}
	if err := dataManager.SetThresholdProfile(th, true); err != nil {
		t.Error(err)
	}
	if err := dataManager.SetThresholdProfile(th2, true); err != nil {
		t.Error(err)
	}
	if err := dataManager.SetThresholdProfile(th3, true); err != nil {
		t.Error(err)
	}
	eIdxes := map[string]utils.StringSet{
		"*string:*req.Account:1001": {
			"TH1": struct{}{},
			"TH2": struct{}{},
		},
		"*string:*req.Account:1002": {
			"TH3": struct{}{},
		},
	}
	if rcvIdx, err := dataManager.GetIndexes(
		utils.CacheThresholdFilterIndexes, th.Tenant,
		utils.EmptyString, false, false); err != nil {
		t.Error(err)
	} else {
		if !reflect.DeepEqual(eIdxes, rcvIdx) {
			t.Errorf("Expecting %+v, received: %+v", eIdxes, rcvIdx)
		}
	}
	eMp := map[string]utils.StringSet{
		"*string:*req.Account:1001": {
			"TH1": struct{}{},
			"TH2": struct{}{},
		},
	}
	if rcvMp, err := dataManager.GetIndexes(
		utils.CacheThresholdFilterIndexes, th.Tenant,
		utils.ConcatenatedKey(utils.MetaString, utils.MetaReq+utils.NestingSep+utils.AccountField, "1001"),
		true, true); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eMp, rcvMp) {
		t.Errorf("Expecting: %+v, received: %+v", eMp, rcvMp)
	}
}

func testITTestIndexingMetaNot(t *testing.T) {
	th := &ThresholdProfile{
		Tenant:    "cgrates.org",
		ID:        "TH1",
		FilterIDs: []string{"*string:~*req.Account:1001", "*notstring:~*req.Destination:+49123"},
		ActionIDs: []string{},
	}
	th2 := &ThresholdProfile{
		Tenant:    "cgrates.org",
		ID:        "TH2",
		FilterIDs: []string{"*prefix:~*req.EventName:Name", "*notprefix:~*req.Destination:10"},
		ActionIDs: []string{},
	}
	th3 := &ThresholdProfile{
		Tenant:    "cgrates.org",
		ID:        "TH3",
		FilterIDs: []string{"*notstring:~*req.Account:1002", "*notstring:~*req.Balance:1000"},
		ActionIDs: []string{},
	}
	if err := dataManager.SetThresholdProfile(th, true); err != nil {
		t.Error(err)
	}
	if err := dataManager.SetThresholdProfile(th2, true); err != nil {
		t.Error(err)
	}
	if err := dataManager.SetThresholdProfile(th3, true); err != nil {
		t.Error(err)
	}
	eIdxes := map[string]utils.StringSet{
		"*string:*req.Account:1001": {
			"TH1": struct{}{},
		},
		"*prefix:*req.EventName:Name": {
			"TH2": struct{}{},
		},
	}
	if rcvIdx, err := dataManager.GetIndexes(
		utils.CacheThresholdFilterIndexes, th.Tenant,
		utils.EmptyString, false, false); err != nil {
		t.Error(err)
	} else {
		if !reflect.DeepEqual(eIdxes, rcvIdx) {
			t.Errorf("Expecting %+v, received: %+v", eIdxes, rcvIdx)
		}
	}
	eMp := map[string]utils.StringSet{
		"*string:*req.Account:1001": {
			"TH1": struct{}{},
		},
	}
	if rcvMp, err := dataManager.GetIndexes(
		utils.CacheThresholdFilterIndexes, th.Tenant,
		utils.ConcatenatedKey(utils.MetaString, utils.MetaReq+utils.NestingSep+utils.AccountField, "1001"),
		true, true); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eMp, rcvMp) {
		t.Errorf("Expecting: %+v, received: %+v", eMp, rcvMp)
	}
}

func testITTestIndexingMetaSuffix(t *testing.T) {
	th := &ThresholdProfile{
		Tenant:    "cgrates.org",
		ID:        "TH1",
		FilterIDs: []string{"*string:~*req.Account:1001", "*suffix:~*req.Subject:10"},
		ActionIDs: []string{},
	}
	th2 := &ThresholdProfile{
		Tenant:    "cgrates.org",
		ID:        "TH2",
		FilterIDs: []string{"*string:~*req.Destination:1002", "*suffix:~*req.Subject:101"},
		ActionIDs: []string{},
	}
	th3 := &ThresholdProfile{
		Tenant:    "cgrates.org",
		ID:        "TH3",
		FilterIDs: []string{"*string:~*req.Destination:1002", "*prefix:~*req.Account:100", "*suffix:~*req.Random:Prfx"},
		ActionIDs: []string{},
	}
	if err := dataManager.SetThresholdProfile(th, true); err != nil {
		t.Error(err)
	}
	if err := dataManager.SetThresholdProfile(th2, true); err != nil {
		t.Error(err)
	}
	if err := dataManager.SetThresholdProfile(th3, true); err != nil {
		t.Error(err)
	}
	eIdxes := map[string]utils.StringSet{
		"*prefix:*req.Account:100": {
			"TH3": struct{}{},
		},
		"*string:*req.Account:1001": {
			"TH1": struct{}{},
		},
		"*string:*req.Destination:1002": {
			"TH2": struct{}{},
			"TH3": struct{}{},
		},
		"*suffix:*req.Random:Prfx": {
			"TH3": struct{}{},
		},
		"*suffix:*req.Subject:10": {
			"TH1": struct{}{},
		},
		"*suffix:*req.Subject:101": {
			"TH2": struct{}{},
		},
	}
	if rcvIdx, err := dataManager.GetIndexes(
		utils.CacheThresholdFilterIndexes, th.Tenant,
		utils.EmptyString, false, false); err != nil {
		t.Error(err)
	} else {
		if !reflect.DeepEqual(eIdxes, rcvIdx) {
			t.Errorf("Expecting %+v,\n received: %+v", utils.ToJSON(eIdxes), utils.ToJSON(rcvIdx))
		}
	}
}
