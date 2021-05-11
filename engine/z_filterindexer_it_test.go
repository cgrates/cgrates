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

	"github.com/cgrates/birpc/context"
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
	testITAccountIndexes,
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
	testITActionProfileIndexes,
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
	testITIndexRateProfileRateIndexes,
	testITFlush,
	testITIsDBEmpty,
	testITIndexRateProfileIndexes,
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
	if err := dataManager.SetIndexes(context.Background(), utils.CacheResourceFilterIndexes,
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

	if exsbjDan, err := dataManager.GetIndexes(context.Background(),
		utils.CacheResourceFilterIndexes, "cgrates.org",
		utils.ConcatenatedKey(utils.MetaString, "Subject", "dan"),
		false, false); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedsbjDan, exsbjDan) {
		t.Errorf("Expecting: %+v, received: %+v", expectedsbjDan, exsbjDan)
	}
	if rcv, err := dataManager.GetIndexes(context.Background(),
		utils.CacheResourceFilterIndexes,
		"cgrates.org", utils.EmptyString,
		false, false); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eIdxes, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", eIdxes, rcv)
	}
	//invalid tnt:context or index key
	if _, err := dataManager.GetIndexes(context.Background(),
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
	if rcvMp, err := dataManager.GetIndexes(context.Background(),
		utils.CacheResourceFilterIndexes, "cgrates.org",
		utils.ConcatenatedKey(utils.MetaString, "Account", "1002"),
		false, true); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eMp, rcvMp) {
		t.Errorf("Expecting: %+v, received: %+v", eMp, rcvMp)
	}

	//invalid tnt:context or index key
	if _, err := dataManager.GetIndexes(context.Background(),
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
	if err := dataManager.SetFilter(context.Background(), fp, true); err != nil {
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
	if rcvIdx, err := dataManager.GetIndexes(context.Background(),
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
	if err := dataManager.SetFilter(context.Background(), fp2, true); err != nil {
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
	if rcvIdx, err := dataManager.GetIndexes(context.Background(),
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
	if err := dataManager.SetFilter(context.Background(), fp3, true); err != nil {
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
	if rcvIdx, err := dataManager.GetIndexes(context.Background(),
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
	if err := dataManager.SetFilter(context.Background(), fp3, true); err != nil {
		t.Error(err)
	}

	eIdxes = map[string]utils.StringSet{
		utils.CacheThresholdFilterIndexes: {
			"THD_Test": struct{}{},
		},
	}
	if rcvIdx, err := dataManager.GetIndexes(context.Background(),
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
	if rcvIdx, err := dataManager.GetIndexes(context.Background(),
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
	if _, err := dataManager.GetIndexes(context.Background(),
		utils.CacheThresholdFilterIndexes, th.Tenant,
		utils.EmptyString, false, false); err != utils.ErrNotFound {
		t.Error(err)
	}
	if _, err := dataManager.GetIndexes(context.Background(),
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
	if err := dataManager.SetFilter(context.Background(), fp, true); err != nil {
		t.Error(err)
	}
	attrProfile := &AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "AttrPrf",
		FilterIDs: []string{"AttrFilter", "*ai:~*req.AnswerTime:2014-07-14T14:25:00Z"},
		Contexts:  []string{"con1", "con2"},
		Attributes: []*Attribute{
			{
				Path:  "FN1",
				Value: config.NewRSRParsersMustCompile("Val1", utils.InfieldSep),
			},
		},
		Weight: 20,
	}
	//Set AttributeProfile with 2 contexts (con1 , con2)
	if err := dataManager.SetAttributeProfile(context.Background(), attrProfile, true); err != nil {
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
		if rcvIdx, err := dataManager.GetIndexes(context.Background(),
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
		FilterIDs: []string{"AttrFilter", "*ai:~*req.AnswerTime:2014-07-14T14:25:00Z"},
		Contexts:  []string{"con3"},
		Attributes: []*Attribute{
			{
				Path:  "FN1",
				Value: config.NewRSRParsersMustCompile("Val1", utils.InfieldSep),
			},
		},
		Weight: 20,
	}
	if err := dataManager.SetAttributeProfile(context.Background(), attrProfile, true); err != nil {
		t.Error(err)
	}
	//check indexes with the new context (con3)
	if rcvIdx, err := dataManager.GetIndexes(context.Background(),
		utils.CacheAttributeFilterIndexes,
		utils.ConcatenatedKey(attrProfile.Tenant, "con3"),
		utils.EmptyString, false, false); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eIdxes, rcvIdx) {
		t.Errorf("Expecting %+v, received: %+v", eIdxes, rcvIdx)
	}
	//check if old contexts was delete
	for _, ctx := range []string{"con1", "con2"} {
		if _, err = dataManager.GetIndexes(context.Background(),
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
	if err := dataManager.SetFilter(context.Background(), fp, true); err != nil {
		t.Error(err)
	}
	eIdxes = map[string]utils.StringSet{
		"*string:*req.EventType:Event3": {
			"AttrPrf": struct{}{},
		},
	}
	for _, ctx := range attrProfile.Contexts {
		if rcvIdx, err := dataManager.GetIndexes(context.Background(),
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
	if rcvIdx, err := dataManager.GetIndexes(context.Background(),
		utils.CacheReverseFilterIndexes,
		fp.TenantID(),
		utils.EmptyString, false, false); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eIdxes, rcvIdx) {
		t.Errorf("Expecting %+v, received: %+v", eIdxes, rcvIdx)
	}

	if err := dataManager.RemoveAttributeProfile(context.Background(), attrProfile.Tenant,
		attrProfile.ID, utils.NonTransactional, true); err != nil {
		t.Error(err)
	}
	//check if index is removed
	if _, err := dataManager.GetIndexes(context.Background(),
		utils.CacheAttributeFilterIndexes,
		utils.ConcatenatedKey("cgrates.org", "con3"),
		utils.MetaString, false, false); err != nil && err != utils.ErrNotFound {
		t.Error(err)
	}

	if _, err := dataManager.GetIndexes(context.Background(),
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
	if err := dataManager.SetFilter(context.Background(), fp, true); err != nil {
		t.Error(err)
	}
	attrProfile := &AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "AttrPrf",
		FilterIDs: []string{"AttrFilter", "*ai:~*req.AnswerTime:2014-07-14T14:25:00Z"},
		Contexts:  []string{"con1", "con2"},
		Attributes: []*Attribute{
			{
				Path:  "FN1",
				Value: config.NewRSRParsersMustCompile("Val1", utils.InfieldSep),
			},
		},
		Weight: 20,
	}
	//Set AttributeProfile with 2 contexts ( con1 , con2)
	if err := dataManager.SetAttributeProfile(context.Background(), attrProfile, true); err != nil {
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
		if rcvIdx, err := dataManager.GetIndexes(context.Background(),
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
		FilterIDs: []string{"AttrFilter", "*ai:~*req.AnswerTime:2014-07-14T14:25:00Z"},
		Contexts:  []string{"con3"},
		Attributes: []*Attribute{
			{
				Path:  "FN1",
				Value: config.NewRSRParsersMustCompile("Val1", utils.InfieldSep),
			},
		},
		Weight: 20,
	}
	if err := dataManager.SetAttributeProfile(context.Background(), attrProfile, true); err != nil {
		t.Error(err)
	}
	//check indexes with the new context (con3)
	if rcvIdx, err := dataManager.GetIndexes(context.Background(),
		utils.CacheAttributeFilterIndexes,
		utils.ConcatenatedKey(attrProfile.Tenant, "con3"),
		utils.EmptyString, false, false); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eIdxes, rcvIdx) {
		t.Errorf("Expecting %+v, received: %+v", eIdxes, rcvIdx)
	}
	//check if old contexts was delete
	for _, ctx := range []string{"con1", "con2"} {
		if _, err = dataManager.GetIndexes(context.Background(),
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
	if err := dataManager.SetFilter(context.Background(), fp, true); err != nil {
		t.Error(err)
	}
	eIdxes = map[string]utils.StringSet{
		"*string:*req.EventType:Event3": {
			"AttrPrf": struct{}{},
		},
	}
	for _, ctx := range attrProfile.Contexts {
		if rcvIdx, err := dataManager.GetIndexes(context.Background(),
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
	if rcvIdx, err := dataManager.GetIndexes(context.Background(),
		utils.CacheReverseFilterIndexes,
		fp.TenantID(),
		utils.EmptyString, false, false); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eIdxes, rcvIdx) {
		t.Errorf("Expecting %+v, received: %+v", eIdxes, rcvIdx)
	}

	if err := dataManager.RemoveAttributeProfile(context.Background(), attrProfile.Tenant,
		attrProfile.ID, utils.NonTransactional, true); err != nil {
		t.Error(err)
	}
	//check if index is removed
	if _, err := dataManager.GetIndexes(context.Background(),
		utils.CacheAttributeFilterIndexes,
		utils.ConcatenatedKey("cgrates.org", "con3"),
		utils.MetaString, false, false); err != nil && err != utils.ErrNotFound {
		t.Error(err)
	}

	if _, err := dataManager.GetIndexes(context.Background(),
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
	if err := dataManager.SetFilter(context.Background(), fp, true); err != nil {
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
	if rcvIdx, err := dataManager.GetIndexes(context.Background(),
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
	if rcvIdx, err := dataManager.GetIndexes(context.Background(),
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
	if _, err := dataManager.GetIndexes(context.Background(),
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
	if err := dataManager.SetIndexes(context.Background(), utils.CacheResourceFilterIndexes,
		"cgrates.org", idxes, false, "transaction1"); err != nil {
		t.Error(err)
	}

	//commit transaction
	if err := dataManager.SetIndexes(context.Background(), utils.CacheResourceFilterIndexes,
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
	if rcv, err := dataManager.GetIndexes(context.Background(),
		utils.CacheResourceFilterIndexes, "cgrates.org",
		utils.EmptyString, false, false); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eIdx, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", eIdx, rcv)
	}
}

func testITAccountIndexes(t *testing.T) {
	fltr1 := &Filter{
		Tenant: "cgrates.org",
		ID:     "FIRST",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaString,
				Element: "~*req.Destination",
				Values:  []string{"DEST1", "DEST2", "~DynamicValue"},
			},
		},
	}
	if err := dataManager.SetFilter(context.Background(), fltr1, true); err != nil {
		t.Error(err)
	}

	accPrf1 := &utils.Account{
		Tenant:    "cgrates.org",
		ID:        "test_ID1",
		FilterIDs: []string{"FIRST", "*string:~*req.Account:DAN"},
		Balances: map[string]*utils.Balance{
			"VoiceBalance": {
				ID:    "VoiceBalance",
				Type:  utils.MetaAbstract,
				Units: utils.NewDecimal(100, 0),
			},
		},
	}
	accPrf2 := &utils.Account{
		Tenant:    "cgrates.org",
		ID:        "test_ID2",
		FilterIDs: []string{"FIRST"},
		Balances: map[string]*utils.Balance{
			"ConcreteBalance": {
				ID:    "ConcreteBalance",
				Type:  utils.MetaConcrete,
				Units: utils.NewDecimal(200, 0),
			},
		},
	}

	if err := dataManager.SetAccount(accPrf1, true); err != nil {
		t.Error(err)
	} else if err := dataManager.SetAccount(accPrf2, true); err != nil {
		t.Error(err)
	}

	eIdxes := map[string]utils.StringSet{
		"*string:*req.Account:DAN": {
			"test_ID1": struct{}{},
		},
		"*string:*req.Destination:DEST1": {
			"test_ID1": struct{}{},
			"test_ID2": struct{}{},
		},
		"*string:*req.Destination:DEST2": {
			"test_ID1": struct{}{},
			"test_ID2": struct{}{},
		},
	}
	if rcvIDx, err := dataManager.GetIndexes(context.Background(), utils.CacheAccountsFilterIndexes,
		"cgrates.org", utils.EmptyString, false, false); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcvIDx, eIdxes) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(eIdxes), utils.ToJSON(rcvIDx))
	}

	//add another filter for matching
	fltr2 := &Filter{
		Tenant: "cgrates.org",
		ID:     "SECOND",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaString,
				Element: "dan",
				Values:  []string{"DEST3", "~*req.Accounts", "~*req.Owner"},
			},
		},
	}
	if err := dataManager.SetFilter(context.Background(), fltr2, true); err != nil {
		t.Error(err)
	}

	accPrf3 := &utils.Account{
		Tenant:    "cgrates.org",
		ID:        "test_ID3",
		FilterIDs: []string{"SECOND", "*string:~*req.Account:DAN"},
		Balances: map[string]*utils.Balance{
			"ConcreteBalance": {
				ID:    "ConcreteBalance",
				Type:  utils.MetaConcrete,
				Units: utils.NewDecimal(200, 0),
			},
		},
	}
	if err := dataManager.SetAccount(accPrf3, true); err != nil {
		t.Error(err)
	}

	eIdxes = map[string]utils.StringSet{
		"*string:*req.Accounts:dan": {
			"test_ID3": struct{}{},
		},
		"*string:*req.Owner:dan": {
			"test_ID3": struct{}{},
		},
		"*string:*req.Account:DAN": {
			"test_ID1": struct{}{},
			"test_ID3": struct{}{},
		},
		"*string:*req.Destination:DEST1": {
			"test_ID1": struct{}{},
			"test_ID2": struct{}{},
		},
		"*string:*req.Destination:DEST2": {
			"test_ID1": struct{}{},
			"test_ID2": struct{}{},
		},
	}
	if rcvIDx, err := dataManager.GetIndexes(context.Background(), utils.CacheAccountsFilterIndexes,
		"cgrates.org", utils.EmptyString, false, false); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcvIDx, eIdxes) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(eIdxes), utils.ToJSON(rcvIDx))
	}

	eIdxes = map[string]utils.StringSet{
		"*string:*req.Destination:DEST1": {
			"test_ID1": struct{}{},
			"test_ID2": struct{}{},
		},
	}
	if rcvIDx, err := dataManager.GetIndexes(context.Background(), utils.CacheAccountsFilterIndexes,
		"cgrates.org", "*string:*req.Destination:DEST1", false, false); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcvIDx, eIdxes) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(eIdxes), utils.ToJSON(rcvIDx))
	}

	//here we will update the filters
	fltr1 = &Filter{
		Tenant: "cgrates.org",
		ID:     "FIRST",
		Rules:  []*FilterRule{{Type: utils.MetaString, Element: "~*req.Destination", Values: []string{"DEST5"}}},
	}
	fltr2 = &Filter{
		Tenant: "cgrates.org",
		ID:     "SECOND",
		Rules:  []*FilterRule{{Type: utils.MetaString, Element: "DEST4", Values: []string{"~*req.CGRID"}}},
	}
	if err := dataManager.SetFilter(context.Background(), fltr1, true); err != nil {
		t.Error(err)
	} else if err := dataManager.SetFilter(context.Background(), fltr2, true); err != nil {
		t.Error(err)
	}

	eIdxes = map[string]utils.StringSet{
		"*string:*req.Destination:DEST5": {
			"test_ID1": struct{}{},
			"test_ID2": struct{}{},
		},
		"*string:*req.CGRID:DEST4": {
			"test_ID3": struct{}{},
		},
		"*string:*req.Account:DAN": {
			"test_ID1": struct{}{},
			"test_ID3": struct{}{},
		},
	}
	if rcvIDx, err := dataManager.GetIndexes(context.Background(), utils.CacheAccountsFilterIndexes,
		"cgrates.org", utils.EmptyString, false, false); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcvIDx, eIdxes) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(eIdxes), utils.ToJSON(rcvIDx))
	}

	// here we will get the reverse indexing
	eIdxes = map[string]utils.StringSet{
		utils.CacheAccountsFilterIndexes: {
			"test_ID1": struct{}{},
			"test_ID2": struct{}{},
		},
	}
	if rcvIDx, err := dataManager.GetIndexes(context.Background(), utils.CacheReverseFilterIndexes,
		"cgrates.org:FIRST", utils.EmptyString, false, false); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcvIDx, eIdxes) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(eIdxes), utils.ToJSON(rcvIDx))
	}

	eIdxes = map[string]utils.StringSet{
		utils.CacheAccountsFilterIndexes: {
			"test_ID3": struct{}{},
		},
	}
	if rcvIDx, err := dataManager.GetIndexes(context.Background(), utils.CacheReverseFilterIndexes,
		"cgrates.org:SECOND", utils.EmptyString, false, false); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcvIDx, eIdxes) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(eIdxes), utils.ToJSON(rcvIDx))
	}

	//invalid tnt:context or index key
	eIdxes = nil
	if rcvIDx, err := dataManager.GetIndexes(context.Background(), utils.CacheAccountsFilterIndexes,
		"cgrates.org", "*string:*req.Destination:DEST6", false, false); err == nil || err != utils.ErrNotFound {
		t.Errorf("Expected %+v, received %+v", utils.ErrNotFound, err)
	} else if !reflect.DeepEqual(rcvIDx, eIdxes) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(eIdxes), utils.ToJSON(rcvIDx))
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
				Values:  []string{"DEST_RES1", "~DynamicValue", "DEST_RES2"},
			},
		},
	}
	if err := dataManager.SetFilter(context.Background(), fltr1, true); err != nil {
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

	expected := "broken reference to filter: FIRST for item with ID: cgrates.org:RES_PRF1"
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
	if rcvIDx, err := dataManager.GetIndexes(context.Background(), utils.CacheResourceFilterIndexes,
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
	if err := dataManager.SetFilter(context.Background(), fltr1, true); err != nil {
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
	if rcvIDx, err := dataManager.GetIndexes(context.Background(), utils.CacheResourceFilterIndexes,
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
	if rcvIDx, err := dataManager.GetIndexes(context.Background(), utils.CacheReverseFilterIndexes,
		"cgrates.org:RES_FLTR1", utils.EmptyString, false, false); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcvIDx, eIdxes) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(eIdxes), utils.ToJSON(rcvIDx))
	}

	//as we updated our filter, the old one is deleted
	if _, err := dataManager.GetIndexes(context.Background(), utils.CacheResourceFilterIndexes,
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
	if err := dataManager.SetFilter(context.Background(), fltr1, true); err != nil {
		t.Error(err)
	}
	if err := dataManager.SetFilter(context.Background(), fltr2, true); err != nil {
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
	if rcvIDx, err := dataManager.GetIndexes(context.Background(), utils.CacheStatFilterIndexes,
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
	if rcvIDx, err := dataManager.GetIndexes(context.Background(), utils.CacheReverseFilterIndexes,
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
	if rcvIDx, err := dataManager.GetIndexes(context.Background(), utils.CacheReverseFilterIndexes,
		"cgrates.org:SQUEUE2", utils.EmptyString, false, false); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcvIDx, eIdxes) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(eIdxes), utils.ToJSON(rcvIDx))
	}

	//invalid tnt:context or index key
	eIdxes = nil
	if rcvIDx, err := dataManager.GetIndexes(context.Background(), utils.CacheStatFilterIndexes,
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
	if err := dataManager.SetFilter(context.Background(), fltr1, true); err != nil {
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
	if rcvIDx, err := dataManager.GetIndexes(context.Background(), utils.CacheChargerFilterIndexes,
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
	if err := dataManager.SetFilter(context.Background(), fltr1, true); err != nil {
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
	if rcvIDx, err := dataManager.GetIndexes(context.Background(), utils.CacheChargerFilterIndexes,
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
	if rcvIDx, err := dataManager.GetIndexes(context.Background(), utils.CacheReverseFilterIndexes,
		"cgrates.org:CHARGER_FLTR", utils.EmptyString, false, false); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcvIDx, expIdx) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(expIdx), utils.ToJSON(rcvIDx))
	}

	//the old filter is deleted
	expIdx = nil
	if rcvIDx, err := dataManager.GetIndexes(context.Background(), utils.CacheChargerFilterIndexes,
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
	if err := dataManager.SetFilter(context.Background(), fltr1, true); err != nil {
		t.Error(err)
	} else if err := dataManager.SetFilter(context.Background(), fltr2, true); err != nil {
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
	if rcvIDx, err := dataManager.GetIndexes(context.Background(), utils.CacheDispatcherFilterIndexes,
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
	if rcvIDx, err := dataManager.GetIndexes(context.Background(), utils.CacheReverseFilterIndexes,
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
	if rcvIDx, err := dataManager.GetIndexes(context.Background(), utils.CacheReverseFilterIndexes,
		"cgrates.org:DISPATCHER_FLTR2", utils.EmptyString, false, false); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcvIDx, expIdx) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(expIdx), utils.ToJSON(rcvIDx))
	}

	//invalid tnt:context or index key
	expIdx = nil
	if rcvIDx, err := dataManager.GetIndexes(context.Background(), utils.CacheDispatcherFilterIndexes,
		"cgrates.org:attributes", utils.EmptyString, false, false); err == nil || err != utils.ErrNotFound {
		t.Errorf("Expectedd %+v, received %+v", utils.ErrNotFound, err)
	} else if !reflect.DeepEqual(rcvIDx, expIdx) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(expIdx), utils.ToJSON(rcvIDx))
	}
}

func testITActionProfileIndexes(t *testing.T) {
	fltr1 := &Filter{
		Tenant: "itsyscom",
		ID:     "ACTPRF_FLTR1",
		Rules:  []*FilterRule{{Type: utils.MetaString, Element: "~*req.Destination", Values: []string{"ACC1", "ACC2", "~*req.Account"}}},
	}
	fltr2 := &Filter{
		Tenant: "itsyscom",
		ID:     "ACTPRF_FLTR2",
		Rules:  []*FilterRule{{Type: utils.MetaString, Element: "20m", Values: []string{"USAGE", "~*opts.Debited", "~*req.Usage"}}},
	}
	if err := dataManager.SetFilter(context.Background(), fltr1, true); err != nil {
		t.Error(err)
	} else if err := dataManager.SetFilter(context.Background(), fltr2, true); err != nil {
		t.Error(err)
	}

	actPrf1 := &ActionProfile{
		Tenant:    "itsyscom",
		ID:        "ACTPRF1",
		FilterIDs: []string{"ACTPRF_FLTR1", "*prefix:~*req.Destination:123"},
	}
	actPrf2 := &ActionProfile{
		Tenant:    "itsyscom",
		ID:        "ACTPRF2",
		FilterIDs: []string{"ACTPRF_FLTR2"},
	}
	if err := dataManager.SetActionProfile(actPrf1, true); err != nil {
		t.Error(err)
	} else if err := dataManager.SetActionProfile(actPrf2, true); err != nil {
		t.Error(err)
	}

	expIdx := map[string]utils.StringSet{
		"*string:*req.Destination:ACC1": {
			"ACTPRF1": struct{}{},
		},
		"*string:*req.Destination:ACC2": {
			"ACTPRF1": struct{}{},
		},
		"*prefix:*req.Destination:123": {
			"ACTPRF1": struct{}{},
		},
		"*string:*opts.Debited:20m": {
			"ACTPRF2": struct{}{},
		},
		"*string:*req.Usage:20m": {
			"ACTPRF2": struct{}{},
		},
	}
	if rcvIdx, err := dataManager.GetIndexes(context.Background(), utils.CacheActionProfilesFilterIndexes,
		"itsyscom", utils.EmptyString, false, false); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcvIdx, expIdx) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(expIdx), utils.ToJSON(rcvIdx))
	}

	expIdx = map[string]utils.StringSet{
		"*string:*req.Destination:ACC1": {
			"ACTPRF1": struct{}{},
		},
	}
	if rcvIdx, err := dataManager.GetIndexes(context.Background(), utils.CacheActionProfilesFilterIndexes,
		"itsyscom", "*string:*req.Destination:ACC1", false, false); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcvIdx, expIdx) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(expIdx), utils.ToJSON(rcvIdx))
	}

	// we will update the filter
	fltr1 = &Filter{
		Tenant: "itsyscom",
		ID:     "ACTPRF_FLTR1",
		Rules:  []*FilterRule{{Type: utils.MetaString, Element: "~*req.ToR", Values: []string{"*voice", "~*req.Account"}}},
	}
	fltr2 = &Filter{
		Tenant: "itsyscom",
		ID:     "ACTPRF_FLTR2",
		Rules:  []*FilterRule{{Type: utils.MetaString, Element: "~*req.CGRID", Values: []string{"CHANGED_ID_1", "~*req.Account"}}},
	}

	if err := dataManager.SetFilter(context.Background(), fltr1, true); err != nil {
		t.Error(err)
	} else if err := dataManager.SetFilter(context.Background(), fltr2, true); err != nil {
		t.Error(err)
	}

	actPrf1 = &ActionProfile{
		Tenant:    "itsyscom",
		ID:        "CHANGED_ACTPRF1",
		FilterIDs: []string{"ACTPRF_FLTR1", "*prefix:~*req.Destination:123"},
	}
	actPrf2 = &ActionProfile{
		Tenant:    "itsyscom",
		ID:        "CHANGED_ACTPRF2",
		FilterIDs: []string{"ACTPRF_FLTR2"},
	}
	if err := dataManager.SetActionProfile(actPrf1, true); err != nil {
		t.Error(err)
	} else if err := dataManager.SetActionProfile(actPrf2, true); err != nil {
		t.Error(err)
	}

	expIdx = map[string]utils.StringSet{
		"*string:*req.ToR:*voice": {
			"ACTPRF1":         struct{}{},
			"CHANGED_ACTPRF1": struct{}{},
		},
		"*string:*req.CGRID:CHANGED_ID_1": {
			"ACTPRF2":         struct{}{},
			"CHANGED_ACTPRF2": struct{}{},
		},
		"*prefix:*req.Destination:123": {
			"ACTPRF1":         struct{}{},
			"CHANGED_ACTPRF1": struct{}{},
		},
	}
	if rcvIdx, err := dataManager.GetIndexes(context.Background(), utils.CacheActionProfilesFilterIndexes,
		"itsyscom", utils.EmptyString, false, false); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcvIdx, expIdx) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(expIdx), utils.ToJSON(rcvIdx))
	}

	//here we will get the reverse indexes
	expIdx = map[string]utils.StringSet{
		utils.CacheActionProfilesFilterIndexes: {
			"ACTPRF1":         struct{}{},
			"CHANGED_ACTPRF1": struct{}{},
		},
	}
	if rcvIdx, err := dataManager.GetIndexes(context.Background(), utils.CacheReverseFilterIndexes,
		"itsyscom:ACTPRF_FLTR1", utils.EmptyString, false, false); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcvIdx, expIdx) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(expIdx), utils.ToJSON(rcvIdx))
	}

	expIdx = map[string]utils.StringSet{
		utils.CacheActionProfilesFilterIndexes: {
			"ACTPRF2":         struct{}{},
			"CHANGED_ACTPRF2": struct{}{},
		},
	}
	if rcvIdx, err := dataManager.GetIndexes(context.Background(), utils.CacheReverseFilterIndexes,
		"itsyscom:ACTPRF_FLTR2", utils.EmptyString, false, false); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcvIdx, expIdx) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(expIdx), utils.ToJSON(rcvIdx))
	}

	//invalid tnt:context or index key
	expIdx = nil
	if rcvIdx, err := dataManager.GetIndexes(context.Background(), utils.CacheActionProfilesFilterIndexes,
		"itsyscom", "*string:*req.Destination:ACC7", false, false); err == nil || err != utils.ErrNotFound {
		t.Errorf("Expected %+v, received %+v", utils.ErrNotFound, err)
	} else if !reflect.DeepEqual(rcvIdx, expIdx) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(expIdx), utils.ToJSON(rcvIdx))
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
	if err := dataManager.SetIndexes(context.Background(), utils.CacheResourceFilterIndexes,
		"cgrates.org", idxes, false, transID); err != nil {
		t.Error(err)
	}
	//commit transaction
	if err := dataManager.SetIndexes(context.Background(), utils.CacheResourceFilterIndexes,
		"cgrates.org", nil, true, transID); err != nil {
		t.Error(err)
	}
	//verify if old key was deleted
	if _, err := dataManager.GetIndexes(context.Background(),
		"tmp_"+utils.CacheResourceFilterIndexes,
		utils.ConcatenatedKey("cgrates.org", transID),
		utils.EmptyString, false, false); err != utils.ErrNotFound {
		t.Error(err)
	}
	//verify new key and check if data was moved
	if rcv, err := dataManager.GetIndexes(context.Background(),
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
	if rcvIdx, err := dataManager.GetIndexes(context.Background(),
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
	if rcvMp, err := dataManager.GetIndexes(context.Background(),
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
	if rcvIdx, err := dataManager.GetIndexes(context.Background(),
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
	if rcvMp, err := dataManager.GetIndexes(context.Background(),
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
	if err := dataManager.SetFilter(context.Background(), fltr1, true); err != nil {
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
	if rcvIdx, err := dataManager.GetIndexes(context.Background(), utils.CacheRouteFilterIndexes,
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
	if rcvIdx, err := dataManager.GetIndexes(context.Background(), utils.CacheRouteFilterIndexes,
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
	if rcvIdx, err := dataManager.GetIndexes(context.Background(), utils.CacheReverseFilterIndexes,
		"cgrates.org:FIRST", utils.EmptyString, false, false); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expIdx, rcvIdx) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(expIdx), utils.ToJSON(rcvIdx))
	}

	//invalid tnt:context or index key
	if _, err := dataManager.GetIndexes(context.Background(), utils.CacheRouteFilterIndexes,
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
	if rcvIdx, err := dataManager.GetIndexes(context.Background(),
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
	if rcvMp, err := dataManager.GetIndexes(context.Background(),
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
	if rcvIdx, err := dataManager.GetIndexes(context.Background(),
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
	if rcvMp, err := dataManager.GetIndexes(context.Background(),
		utils.CacheThresholdFilterIndexes, th.Tenant,
		utils.ConcatenatedKey(utils.MetaString, utils.MetaReq+utils.NestingSep+utils.AccountField, "1001"),
		true, true); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eMp, rcvMp) {
		t.Errorf("Expecting: %+v, received: %+v", eMp, rcvMp)
	}
}

func testITIndexRateProfileRateIndexes(t *testing.T) {
	rPrf := &utils.RateProfile{
		Tenant:    "cgrates.org",
		ID:        "RP1",
		FilterIDs: []string{"*string:~*req.Subject:1001", "*string:~*req.Subject:1002"},
		Weights: utils.DynamicWeights{
			{
				Weight: 0,
			},
		},
		MaxCostStrategy: "*free",
		Rates: map[string]*utils.Rate{
			"FIRST_GI": {
				ID:        "FIRST_GI",
				FilterIDs: []string{"*string:~*req.Category:call"},
				Weights: utils.DynamicWeights{
					{
						Weight: 0,
					},
				},
				Blocker: false,
			},
			"SECOND_GI": {
				ID:        "SECOND_GI",
				FilterIDs: []string{"*string:~*req.Category:voice"},
				Weights: utils.DynamicWeights{
					{
						Weight: 10,
					},
				},
				Blocker: false,
			},
		},
	}
	if err := dataManager.SetRateProfile(context.Background(), rPrf, true); err != nil {
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
	if rcvIdx, err := dataManager.GetIndexes(context.Background(),
		utils.CacheRateFilterIndexes, "cgrates.org:RP1",
		utils.EmptyString, true, true); err != nil {
		t.Error(err)
	} else {
		if !reflect.DeepEqual(eIdxes, rcvIdx) {
			t.Errorf("Expecting %+v, received: %+v", utils.ToJSON(eIdxes), utils.ToJSON(rcvIdx))
		}
	}

	// update the RateProfile by adding a new Rate
	rPrf = &utils.RateProfile{ // recreate the profile because if we test on internal
		Tenant:    "cgrates.org", // each update on the original item will update the item from DB
		ID:        "RP1",
		FilterIDs: []string{"*string:~*req.Subject:1001", "*string:~*req.Subject:1002"},
		Weights: utils.DynamicWeights{
			{
				Weight: 0,
			},
		},
		MaxCostStrategy: "*free",
		Rates: map[string]*utils.Rate{
			"FIRST_GI": {
				ID:        "FIRST_GI",
				FilterIDs: []string{"*string:~*req.Category:call"},
				Weights: utils.DynamicWeights{
					{
						Weight: 0,
					},
				},
				Blocker: false,
			},
			"SECOND_GI": {
				ID:        "SECOND_GI",
				FilterIDs: []string{"*string:~*req.Category:voice"},
				Weights: utils.DynamicWeights{
					{
						Weight: 10,
					},
				},
				Blocker: false,
			},
			"THIRD_GI": {
				ID:        "THIRD_GI",
				FilterIDs: []string{"*string:~*req.Category:custom"},
				Weights: utils.DynamicWeights{
					{
						Weight: 20,
					},
				},
				Blocker: false,
			},
		},
	}
	if err := dataManager.SetRateProfile(context.Background(), rPrf, true); err != nil {
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
	if rcvIdx, err := dataManager.GetIndexes(context.Background(),
		utils.CacheRateFilterIndexes, "cgrates.org:RP1",
		utils.EmptyString, true, true); err != nil {
		t.Error(err)
	} else {
		if !reflect.DeepEqual(eIdxes, rcvIdx) {
			t.Errorf("Expecting %+v, received: %+v", eIdxes, rcvIdx)
		}
	}

	//here we will set a filter
	fltr := &Filter{
		Tenant: "cgrates.org",
		ID:     "FLTR",
		Rules:  []*FilterRule{{Type: utils.MetaString, Element: "Dan", Values: []string{"~*req.Account", "~*req.Destination", "DAN2"}}},
	}
	if err := dataManager.SetFilter(context.Background(), fltr, true); err != nil {
		t.Error(err)
	}
	rPrf2 := &utils.RateProfile{
		Tenant: "cgrates.org",
		ID:     "RP2",
		Weights: utils.DynamicWeights{
			{
				Weight: 0,
			},
		},
		MaxCostStrategy: "*free",
		Rates: map[string]*utils.Rate{
			"CUSTOM_RATE1": {
				ID:        "CUSTOM_RATE1",
				FilterIDs: []string{"*string:~*req.Subject:1001", "FLTR"},
				Weights: utils.DynamicWeights{
					{
						Weight: 0,
					},
				},
				Blocker: false,
			},
			"CUSTOM_RATE2": {
				ID:        "CUSTOM_RATE2",
				FilterIDs: []string{"*string:~*req.Subject:1001", "*string:~*req.Category:call"},
				Weights: utils.DynamicWeights{
					{
						Weight: 10,
					},
				},
				Blocker: false,
			},
		},
	}
	if err := dataManager.SetRateProfile(context.Background(), rPrf2, true); err != nil {
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
		"*string:*req.Account:Dan": {
			"CUSTOM_RATE1": struct{}{},
		},
		"*string:*req.Destination:Dan": {
			"CUSTOM_RATE1": struct{}{},
		},
	}
	if rcvIdx, err := dataManager.GetIndexes(context.Background(),
		utils.CacheRateFilterIndexes, "cgrates.org:RP2",
		utils.EmptyString, true, true); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eIdxes, rcvIdx) {
		t.Errorf("Expecting %+v, received: %+v", eIdxes, rcvIdx)
	}

	//here we will check the reverse indexes
	eIdxes = map[string]utils.StringSet{
		utils.CacheRateFilterIndexes: {
			"CUSTOM_RATE1:RP2": struct{}{},
		},
	}
	if rcvIdx, err := dataManager.GetIndexes(context.Background(),
		utils.CacheReverseFilterIndexes, "cgrates.org:FLTR",
		utils.EmptyString, true, true); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eIdxes, rcvIdx) {
		t.Errorf("Expecting %+v, received: %+v", utils.ToJSON(eIdxes), utils.ToJSON(rcvIdx))
	}

	//now we will update the filter
	fltr = &Filter{
		Tenant: "cgrates.org",
		ID:     "FLTR",
		Rules:  []*FilterRule{{Type: utils.MetaString, Element: "10m", Values: []string{"~*req.Usage", "~*req.Debited", "DAN2"}}},
	}
	if err := dataManager.SetFilter(context.Background(), fltr, true); err != nil {
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
		"*string:*req.Usage:10m": {
			"CUSTOM_RATE1": struct{}{},
		},
		"*string:*req.Debited:10m": {
			"CUSTOM_RATE1": struct{}{},
		},
	}
	if rcvIdx, err := dataManager.GetIndexes(context.Background(),
		utils.CacheRateFilterIndexes, "cgrates.org:RP2",
		utils.EmptyString, true, true); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eIdxes, rcvIdx) {
		t.Errorf("Expecting %+v, received: %+v", utils.ToJSON(eIdxes), utils.ToJSON(rcvIdx))
	}

	//invalid or inexisting tenant:context or index key
	eIdxes = nil
	if rcvIdx, err := dataManager.GetIndexes(context.Background(),
		utils.CacheRateFilterIndexes, "cgrates.org:RP4",
		utils.EmptyString, true, true); err == nil || err != utils.ErrNotFound {
		t.Errorf("Expected %+v, received %+v", utils.ErrNotFound, err)
	} else if !reflect.DeepEqual(eIdxes, rcvIdx) {
		t.Errorf("Expecting %+v, received: %+v", utils.ToJSON(eIdxes), utils.ToJSON(rcvIdx))
	}

}

func testITIndexRateProfileIndexes(t *testing.T) {
	fltr1 := &Filter{
		Tenant: "cgrates.org",
		ID:     "FLTR",
		Rules:  []*FilterRule{{Type: utils.MetaString, Element: "~*req.Usage", Values: []string{"10m"}}},
	}
	if err := dataManager.SetFilter(context.Background(), fltr1, true); err != nil {
		t.Error(err)
	}
	rPrf1 := &utils.RateProfile{
		Tenant:          "cgrates.org",
		ID:              "RP1",
		FilterIDs:       []string{"*string:~*req.Subject:1004|1005", "FLTR"},
		MaxCostStrategy: "*free",
		Rates: map[string]*utils.Rate{
			"CUSTOM1_RATE1": {
				ID:        "CUSTOM1_RATE1",
				FilterIDs: []string{"*string:~*req.Subject:1001"},
				Blocker:   false,
			},
			"CUSTOM1_RATE2": {
				ID:        "CUSTOM1_RATE2",
				FilterIDs: []string{"*string:~*req.Subject:1001", "*string:~*req.Category:call"},
				Blocker:   false,
			},
		},
	}
	rPrf2 := &utils.RateProfile{
		Tenant:          "cgrates.org",
		ID:              "RP2",
		FilterIDs:       []string{"*string:~*req.ToR:*sms|*voice", "*string:~*req.Subject:1004"},
		MaxCostStrategy: "*free",
		Rates: map[string]*utils.Rate{
			"CUSTOM2_RATE1": {
				ID:        "CUSTOM2_RATE1",
				FilterIDs: []string{"*string:~*req.Subject:1009"},
				Blocker:   false,
			},
		},
	}
	if err := dataManager.SetRateProfile(context.Background(), rPrf1, true); err != nil {
		t.Error(err)
	} else if err := dataManager.SetRateProfile(context.Background(), rPrf2, true); err != nil {
		t.Error(err)
	}

	expIdx := map[string]utils.StringSet{
		"*string:*req.Subject:1004": {
			"RP1": struct{}{},
			"RP2": struct{}{},
		},
		"*string:*req.Subject:1005": {
			"RP1": struct{}{},
		},
		"*string:*req.ToR:*sms": {
			"RP2": struct{}{},
		},
		"*string:*req.ToR:*voice": {
			"RP2": struct{}{},
		},
		"*string:*req.Usage:10m": {
			"RP1": struct{}{},
		},
	}
	if rcv, err := dataManager.GetIndexes(context.Background(), utils.CacheRateProfilesFilterIndexes,
		"cgrates.org", utils.EmptyString, false, false); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, expIdx) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(expIdx), utils.ToJSON(rcv))
	}

	//we will update the filter
	fltr1 = &Filter{
		Tenant: "cgrates.org",
		ID:     "FLTR",
		Rules:  []*FilterRule{{Type: utils.MetaString, Element: "~*req.CustomField", Values: []string{"234", "567"}}},
	}
	if err := dataManager.SetFilter(context.Background(), fltr1, true); err != nil {
		t.Error(err)
	}
	expIdx = map[string]utils.StringSet{
		"*string:*req.Subject:1004": {
			"RP1": struct{}{},
			"RP2": struct{}{},
		},
		"*string:*req.Subject:1005": {
			"RP1": struct{}{},
		},
		"*string:*req.ToR:*sms": {
			"RP2": struct{}{},
		},
		"*string:*req.ToR:*voice": {
			"RP2": struct{}{},
		},
		"*string:*req.CustomField:234": {
			"RP1": struct{}{},
		},
		"*string:*req.CustomField:567": {
			"RP1": struct{}{},
		},
	}
	if rcv, err := dataManager.GetIndexes(context.Background(), utils.CacheRateProfilesFilterIndexes,
		"cgrates.org", utils.EmptyString, false, false); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, expIdx) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(expIdx), utils.ToJSON(rcv))
	}

	//here we will get the reverse indexes
	expIdx = map[string]utils.StringSet{
		utils.CacheRateProfilesFilterIndexes: {
			"RP1": struct{}{},
		},
	}
	if rcv, err := dataManager.GetIndexes(context.Background(), utils.CacheReverseFilterIndexes,
		"cgrates.org:FLTR", utils.EmptyString, false, false); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, expIdx) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(expIdx), utils.ToJSON(rcv))
	}

	//nothing to get with with an invalid indexKey
	expIdx = nil
	if rcv, err := dataManager.GetIndexes(context.Background(), utils.CacheRateProfilesFilterIndexes,
		"cgrates.org", "*string:*req.CustomField:2346", false, false); err == nil || err != utils.ErrNotFound {
		t.Errorf("Expected %+v, received %+v", utils.ErrNotFound, err)
	} else if !reflect.DeepEqual(rcv, expIdx) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(expIdx), utils.ToJSON(rcv))
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
	if rcvIdx, err := dataManager.GetIndexes(context.Background(),
		utils.CacheThresholdFilterIndexes, th.Tenant,
		utils.EmptyString, false, false); err != nil {
		t.Error(err)
	} else {
		if !reflect.DeepEqual(eIdxes, rcvIdx) {
			t.Errorf("Expecting %+v,\n received: %+v", utils.ToJSON(eIdxes), utils.ToJSON(rcvIdx))
		}
	}
}
