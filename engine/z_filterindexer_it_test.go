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
	testITIndexRateProfile,
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
		cfg, _ := config.NewDefaultCGRConfig()
		redisDB, err := NewRedisStorage(
			fmt.Sprintf("%s:%s", cfg.DataDbCfg().DataDbHost, cfg.DataDbCfg().DataDbPort),
			4, cfg.DataDbCfg().DataDbUser, cfg.DataDbCfg().DataDbPass, cfg.GeneralCfg().DBDataEncoding,
			utils.REDIS_MAX_CONNS, "", false, 0, 0)
		if err != nil {
			t.Fatal("Could not connect to Redis", err.Error())
		}
		cfgDBName = cfg.DataDbCfg().DataDbName
		dataManager = NewDataManager(redisDB, config.CgrConfig().CacheCfg(), nil)
	case utils.MetaMongo:
		cdrsMongoCfgPath := path.Join(*dataDir, "conf", "samples", "tutmongo")
		mgoITCfg, err := config.NewCGRConfigFromPath(cdrsMongoCfgPath)
		if err != nil {
			t.Fatal(err)
		}
		mongoDB, err := NewMongoStorage(mgoITCfg.StorDbCfg().Host,
			mgoITCfg.StorDbCfg().Port, mgoITCfg.StorDbCfg().Name,
			mgoITCfg.StorDbCfg().User, mgoITCfg.StorDbCfg().Password,
			mgoITCfg.GeneralCfg().DBDataEncoding,
			utils.StorDB, nil, false, 10*time.Second)
		if err != nil {
			t.Fatal(err)
		}
		cfgDBName = mgoITCfg.StorDbCfg().Name
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
		t.Errorf("\nExpecting: true got :%+v", test)
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
		utils.ConcatenatedKey(utils.META_NONE, utils.ANY, utils.ANY): {
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
		utils.ConcatenatedKey(utils.META_NONE, utils.ANY, utils.ANY): {
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
				Element: "~*req.EventType",
				Type:    utils.MetaString,
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
				Element: "~*req.Account",
				Type:    utils.MetaString,
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
				Element: "~*req.Destination",
				Type:    utils.MetaString,
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
		th.ID, utils.NonTransactional, true); err != nil {
		t.Error(err)
	}
	if err := dataManager.RemoveThresholdProfile(th2.Tenant,
		th2.ID, utils.NonTransactional, true); err != nil {
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
				Element: "~*req.EventType",
				Type:    utils.MetaString,
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
				Value: config.NewRSRParsersMustCompile("Val1", utils.INFIELD_SEP),
			},
		},
		Weight: 20,
	}
	//Set AttributeProfile with 2 contexts ( con1 , con2)
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
				Value: config.NewRSRParsersMustCompile("Val1", utils.INFIELD_SEP),
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
				Element: "~*req.EventType",
				Type:    utils.MetaString,
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
		attrProfile.ID, utils.NonTransactional, true); err != nil {
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
				Element: "EventType",
				Type:    utils.MetaString,
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
				Value: config.NewRSRParsersMustCompile("Val1", utils.INFIELD_SEP),
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
				Value: config.NewRSRParsersMustCompile("Val1", utils.INFIELD_SEP),
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
				Element: "~*req.EventType",
				Type:    utils.MetaString,
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
		attrProfile.ID, utils.NonTransactional, true); err != nil {
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
				Element: "~*req.EventType",
				Type:    utils.MetaString,
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
		MinSleep:           time.Duration(0 * time.Second),
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
		MinSleep:           time.Duration(0 * time.Second),
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
		th.ID, utils.NonTransactional, true); err != nil {
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
		utils.ConcatenatedKey(utils.META_NONE,
			utils.ANY, utils.ANY): {
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
		utils.ConcatenatedKey(utils.META_NONE,
			utils.ANY, utils.ANY): {
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
		utils.ConcatenatedKey(utils.META_NONE, utils.META_ANY, utils.META_ANY),
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
		utils.ConcatenatedKey(utils.META_NONE, utils.META_ANY, utils.META_ANY),
		true, true); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eMp, rcvMp) {
		t.Errorf("Expecting: %+v, received: %+v", eMp, rcvMp)
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
		utils.ConcatenatedKey(utils.MetaString, utils.MetaReq+utils.NestingSep+utils.Account, "1001"),
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
		utils.ConcatenatedKey(utils.MetaString, utils.MetaReq+utils.NestingSep+utils.Account, "1001"),
		true, true); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eMp, rcvMp) {
		t.Errorf("Expecting: %+v, received: %+v", eMp, rcvMp)
	}
}

func testITIndexRateProfile(t *testing.T) {
	rPrf := &RateProfile{
		Tenant:           "cgrates.org",
		ID:               "RP1",
		FilterIDs:        []string{"*string:~*req.Subject:1001", "*string:~*req.Subject:1002"},
		Weight:           0,
		ConnectFee:       0.1,
		RoundingMethod:   "*up",
		RoundingDecimals: 4,
		MinCost:          0.1,
		MaxCost:          0.6,
		MaxCostStrategy:  "*free",
		Rates: map[string]*Rate{
			"FIRST_GI": {
				ID:        "FIRST_GI",
				FilterIDs: []string{"*string:~*req.Category:call"},
				Weight:    0,
				IntervalRates: []*IntervalRate{
					{
						Value:     0.12,
						Unit:      time.Duration(1 * time.Minute),
						Increment: time.Duration(1 * time.Minute),
					},
				},
				Blocker: false,
			},
			"SECOND_GI": {
				ID:        "SECOND_GI",
				FilterIDs: []string{"*string:~*req.Category:voice"},
				Weight:    10,
				IntervalRates: []*IntervalRate{
					{
						Value:     0.06,
						Unit:      time.Duration(1 * time.Minute),
						Increment: time.Duration(1 * time.Second),
					},
				},
				Blocker: false,
			},
		},
	}
	if err := dataManager.SetRateProfile(rPrf, true); err != nil {
		t.Error(err)
	}
	eIdxes := map[string]utils.StringSet{
		"*string:*req.Category:call": {
			"FIRST_GI": struct{}{},
		},
		"*string:*req.Category:voice": {
			"SECOND_GI": struct{}{},
		},
	}
	if rcvIdx, err := dataManager.GetIndexes(
		utils.CacheRateFilterIndexes, "cgrates.org:RP1",
		utils.EmptyString, true, true); err != nil {
		t.Error(err)
	} else {
		if !reflect.DeepEqual(eIdxes, rcvIdx) {
			t.Errorf("Expecting %+v, received: %+v", utils.ToJSON(eIdxes), utils.ToJSON(rcvIdx))
		}
	}
	// update the RateProfile by adding a new Rate
	rPrf = &RateProfile{ // recreate the profile because if we test on internal
		Tenant:           "cgrates.org", // each update on the original item will update the item from DB
		ID:               "RP1",
		FilterIDs:        []string{"*string:~*req.Subject:1001", "*string:~*req.Subject:1002"},
		Weight:           0,
		ConnectFee:       0.1,
		RoundingMethod:   "*up",
		RoundingDecimals: 4,
		MinCost:          0.1,
		MaxCost:          0.6,
		MaxCostStrategy:  "*free",
		Rates: map[string]*Rate{
			"FIRST_GI": {
				ID:        "FIRST_GI",
				FilterIDs: []string{"*string:~*req.Category:call"},
				Weight:    0,
				IntervalRates: []*IntervalRate{
					{
						Value:     0.12,
						Unit:      time.Duration(1 * time.Minute),
						Increment: time.Duration(1 * time.Minute),
					},
				},
				Blocker: false,
			},
			"SECOND_GI": {
				ID:        "SECOND_GI",
				FilterIDs: []string{"*string:~*req.Category:voice"},
				Weight:    10,
				IntervalRates: []*IntervalRate{
					{
						Value:     0.06,
						Unit:      time.Duration(1 * time.Minute),
						Increment: time.Duration(1 * time.Second),
					},
				},
				Blocker: false,
			},
			"THIRD_GI": {
				ID:        "THIRD_GI",
				FilterIDs: []string{"*string:~*req.Category:custom"},
				Weight:    20,
				IntervalRates: []*IntervalRate{
					{
						Value:     0.06,
						Unit:      time.Duration(1 * time.Minute),
						Increment: time.Duration(1 * time.Second),
					},
				},
				Blocker: false,
			},
		},
	}
	if err := dataManager.SetRateProfile(rPrf, true); err != nil {
		t.Error(err)
	}
	eIdxes = map[string]utils.StringSet{
		"*string:*req.Category:call": {
			"FIRST_GI": struct{}{},
		},
		"*string:*req.Category:voice": {
			"SECOND_GI": struct{}{},
		},
		"*string:*req.Category:custom": {
			"THIRD_GI": struct{}{},
		},
	}
	if rcvIdx, err := dataManager.GetIndexes(
		utils.CacheRateFilterIndexes, "cgrates.org:RP1",
		utils.EmptyString, true, true); err != nil {
		t.Error(err)
	} else {
		if !reflect.DeepEqual(eIdxes, rcvIdx) {
			t.Errorf("Expecting %+v, received: %+v", eIdxes, rcvIdx)
		}
	}
	rPrf2 := &RateProfile{
		Tenant:           "cgrates.org",
		ID:               "RP2",
		Weight:           0,
		ConnectFee:       0.1,
		RoundingMethod:   "*up",
		RoundingDecimals: 4,
		MinCost:          0.1,
		MaxCost:          0.6,
		MaxCostStrategy:  "*free",
		Rates: map[string]*Rate{
			"CUSTOM_RATE1": {
				ID:        "CUSTOM_RATE1",
				FilterIDs: []string{"*string:~*req.Subject:1001"},
				Weight:    0,
				IntervalRates: []*IntervalRate{
					{
						Value:     0.12,
						Unit:      time.Duration(1 * time.Minute),
						Increment: time.Duration(1 * time.Minute),
					},
				},
				Blocker: false,
			},
			"CUSTOM_RATE2": {
				ID:        "CUSTOM_RATE2",
				FilterIDs: []string{"*string:~*req.Subject:1001", "*string:~*req.Category:call"},
				Weight:    10,
				IntervalRates: []*IntervalRate{
					{
						Value:     0.6,
						Unit:      time.Duration(1 * time.Minute),
						Increment: time.Duration(1 * time.Second),
					},
				},
				Blocker: false,
			},
		},
	}
	if err := dataManager.SetRateProfile(rPrf2, true); err != nil {
		t.Error(err)
	}
	eIdxes = map[string]utils.StringSet{
		"*string:*req.Subject:1001": {
			"CUSTOM_RATE1": struct{}{},
			"CUSTOM_RATE2": struct{}{},
		},
		"*string:*req.Category:call": {
			"CUSTOM_RATE2": struct{}{},
		},
	}
	if rcvIdx, err := dataManager.GetIndexes(
		utils.CacheRateFilterIndexes, "cgrates.org:RP2",
		utils.EmptyString, true, true); err != nil {
		t.Error(err)
	} else {
		if !reflect.DeepEqual(eIdxes, rcvIdx) {
			t.Errorf("Expecting %+v, received: %+v", eIdxes, rcvIdx)
		}
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
