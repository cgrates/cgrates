/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/
package thresholds

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"testing"
	"testing/synctest"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/guardian"
)

func TestThresholdsCache(t *testing.T) {
	var dmTH *engine.DataManager
	tPrfls := []*utils.ThresholdProfile{
		{
			Tenant:    "cgrates.org",
			ID:        "TH_1",
			FilterIDs: []string{"FLTR_TH_1", "*ai:~*req.AnswerTime:2014-07-14T14:35:00Z"},
			MaxHits:   12,
			Blocker:   false,
			Weights: utils.DynamicWeights{
				{
					Weight: 20.0,
				},
			},
			ActionProfileIDs: []string{"ACT_1", "ACT_2"},
			Async:            false,
		},
		{
			Tenant:    "cgrates.org",
			ID:        "TH_2",
			FilterIDs: []string{"FLTR_TH_2", "*ai:~*req.AnswerTime:2014-07-14T14:35:00Z"},
			MaxHits:   12,
			MinSleep:  5 * time.Minute,
			Blocker:   false,
			Weights: utils.DynamicWeights{
				{
					Weight: 20.0,
				},
			},
			ActionProfileIDs: []string{"ACT_1", "ACT_2"},
			Async:            false,
		},
		{
			Tenant:    "cgrates.org",
			ID:        "TH_3",
			FilterIDs: []string{"FLTR_TH_3", "*ai:~*req.AnswerTime:2014-07-14T14:35:00Z"},
			MaxHits:   12,
			MinSleep:  5 * time.Minute,
			Blocker:   false,
			Weights: utils.DynamicWeights{
				{
					Weight: 20.0,
				},
			},
			ActionProfileIDs: []string{"ACT_1", "ACT_2"},
			Async:            false,
		},
	}
	ths := []*utils.Threshold{
		{
			Tenant: "cgrates.org",
			ID:     "TH_1",
			Hits:   0,
		},
		{
			Tenant: "cgrates.org",
			ID:     "TH_2",
			Hits:   0,
		},
		{
			Tenant: "cgrates.org",
			ID:     "TH_3",
			Hits:   0,
		},
	}

	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dmTH = engine.NewDataManager(dbCM, cfg, nil)
	cacheS := engine.NewCacheS(cfg, nil, nil, nil)
	dmTH.SetCache(cacheS)
	cfg.ThresholdSCfg().StoreInterval = 0
	cfg.ThresholdSCfg().StringIndexedFields = nil
	cfg.ThresholdSCfg().PrefixIndexedFields = nil

	fltrTh1 := &engine.Filter{
		Tenant: cfg.GeneralCfg().DefaultTenant,
		ID:     "FLTR_TH_1",
		Rules: []*engine.FilterRule{
			{
				Type:    utils.MetaString,
				Element: "~*req.Threshold",
				Values:  []string{"TH_1"},
			},
			{
				Type:    utils.MetaGreaterOrEqual,
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Weight,
				Values:  []string{"9.0"},
			},
		},
	}
	dmTH.SetFilter(context.Background(), fltrTh1, true)
	fltrTh2 := &engine.Filter{
		Tenant: cfg.GeneralCfg().DefaultTenant,
		ID:     "FLTR_TH_2",
		Rules: []*engine.FilterRule{
			{
				Type:    utils.MetaString,
				Element: "~*req.Threshold",
				Values:  []string{"TH_2"},
			},
			{
				Type:    utils.MetaGreaterOrEqual,
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Weight,
				Values:  []string{"15.0"},
			},
		},
	}
	dmTH.SetFilter(context.Background(), fltrTh2, true)
	fltrTh3 := &engine.Filter{
		Tenant: cfg.GeneralCfg().DefaultTenant,
		ID:     "FLTR_TH_3",
		Rules: []*engine.FilterRule{
			{
				Type:    utils.MetaPrefix,
				Element: "~*req.Threshold",
				Values:  []string{"ThresholdPrefix"},
			},
		},
	}
	dmTH.SetFilter(context.Background(), fltrTh3, true)
	for _, th := range tPrfls {
		if err := dmTH.SetThresholdProfile(context.TODO(), th, true); err != nil {
			t.Errorf("Error: %+v", err)
		}
	}
	//Test each threshold profile from cache
	for _, th := range tPrfls {
		if temptTh, err := dmTH.GetThresholdProfile(context.TODO(), th.Tenant,
			th.ID, true, false, utils.NonTransactional); err != nil {
			t.Errorf("Error: %+v", err)
		} else if !reflect.DeepEqual(th, temptTh) {
			t.Errorf("Expecting: %+v, received: %+v", th, temptTh)
		}
	}
	for _, th := range ths {
		if err := dmTH.SetThreshold(context.TODO(), th); err != nil {
			t.Errorf("Error: %+v", err)
		}
	}
	//Test each threshold profile from cache
	for _, th := range ths {
		if temptTh, err := dmTH.GetThreshold(context.TODO(), th.Tenant,
			th.ID, true, false, utils.NonTransactional); err != nil {
			t.Errorf("Error: %+v", err)
		} else if !reflect.DeepEqual(th, temptTh) {
			t.Errorf("Expecting: %+v, received: %+v", th, temptTh)
		}
	}

}

func TestThresholdsmatchingThresholdsForEvent(t *testing.T) {
	var dmTH *engine.DataManager
	var thServ *ThresholdS
	var tPrfls = []*utils.ThresholdProfile{
		{
			Tenant:    "cgrates.org",
			ID:        "TH_1",
			FilterIDs: []string{"FLTR_TH_1", "*ai:*now:2014-07-14T14:35:00Z"},
			MaxHits:   12,
			Blocker:   false,
			Weights: utils.DynamicWeights{
				{
					Weight: 20.0,
				},
			},
			ActionProfileIDs: []string{"ACT_1", "ACT_2"},
			Async:            false,
		},
		{
			Tenant:    "cgrates.org",
			ID:        "TH_2",
			FilterIDs: []string{"FLTR_TH_2", "*ai:*now:2014-07-14T14:35:00Z"},
			MaxHits:   12,
			MinSleep:  5 * time.Minute,
			Blocker:   false,
			Weights: utils.DynamicWeights{
				{
					Weight: 20.0,
				},
			},
			ActionProfileIDs: []string{"ACT_1", "ACT_2"},
			Async:            false,
		},
		{
			Tenant:    "cgrates.org",
			ID:        "TH_3",
			FilterIDs: []string{"FLTR_TH_3", "*ai:*now:2014-07-14T14:35:00Z"},
			MaxHits:   12,
			MinSleep:  5 * time.Minute,
			Blocker:   false,
			Weights: utils.DynamicWeights{
				{
					Weight: 20.0,
				},
			},
			ActionProfileIDs: []string{"ACT_1", "ACT_2"},
			Async:            false,
		},
	}
	ths := []*utils.Threshold{
		{
			Tenant: "cgrates.org",
			ID:     "TH_1",
			Hits:   0,
		},
		{
			Tenant: "cgrates.org",
			ID:     "TH_2",
			Hits:   0,
		},
		{
			Tenant: "cgrates.org",
			ID:     "TH_3",
			Hits:   0,
		},
	}
	argsGetThresholds := []*utils.CGREvent{
		{
			Tenant: "cgrates.org",
			ID:     "Ev1",
			Event: map[string]any{
				"Threshold": "TH_1",
				"Weight":    "10.0",
			},
		},
		{
			Tenant: "cgrates.org",
			ID:     "Ev1",
			Event: map[string]any{
				"Threshold": "TH_2",
				"Weight":    "20.0",
			},
		},
		{
			Tenant: "cgrates.org",
			ID:     "Ev1",
			Event: map[string]any{
				"Threshold": "ThresholdPrefix123",
			},
		},
	}

	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dmTH = engine.NewDataManager(dbCM, cfg, nil)
	cacheS := engine.NewCacheS(cfg, nil, nil, nil)
	dmTH.SetCache(cacheS)
	cfg.ThresholdSCfg().StoreInterval = 0
	cfg.ThresholdSCfg().StringIndexedFields = nil
	cfg.ThresholdSCfg().PrefixIndexedFields = nil
	thServ = NewThresholdService(cfg, dmTH, cacheS, engine.NewFilterS(cfg, nil, dmTH), nil)

	fltrTh1 := &engine.Filter{
		Tenant: cfg.GeneralCfg().DefaultTenant,
		ID:     "FLTR_TH_1",
		Rules: []*engine.FilterRule{
			{
				Type:    utils.MetaString,
				Element: "~*req.Threshold",
				Values:  []string{"TH_1"},
			},
			{
				Type:    utils.MetaGreaterOrEqual,
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Weight,
				Values:  []string{"9.0"},
			},
		},
	}
	dmTH.SetFilter(context.Background(), fltrTh1, true)
	fltrTh2 := &engine.Filter{
		Tenant: cfg.GeneralCfg().DefaultTenant,
		ID:     "FLTR_TH_2",
		Rules: []*engine.FilterRule{
			{
				Type:    utils.MetaString,
				Element: "~*req.Threshold",
				Values:  []string{"TH_2"},
			},
			{
				Type:    utils.MetaGreaterOrEqual,
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Weight,
				Values:  []string{"15.0"},
			},
		},
	}
	fltrTh3 := &engine.Filter{
		Tenant: cfg.GeneralCfg().DefaultTenant,
		ID:     "FLTR_TH_3",
		Rules: []*engine.FilterRule{
			{
				Type:    utils.MetaPrefix,
				Element: "~*req.Threshold",
				Values:  []string{"ThresholdPrefix"},
			},
		},
	}
	dmTH.SetFilter(context.TODO(), fltrTh2, true)
	dmTH.SetFilter(context.TODO(), fltrTh3, true)
	for _, th := range tPrfls {
		if err := dmTH.SetThresholdProfile(context.TODO(), th, true); err != nil {
			t.Errorf("Error: %+v", err)
		}
	}
	//Test each threshold profile from cache
	for _, th := range tPrfls {
		if temptTh, err := dmTH.GetThresholdProfile(context.TODO(), th.Tenant,
			th.ID, true, false, utils.NonTransactional); err != nil {
			t.Errorf("Error: %+v", err)
		} else if !reflect.DeepEqual(th, temptTh) {
			t.Errorf("Expecting: %+v, received: %+v", th, temptTh)
		}
	}
	for _, th := range ths {
		if err := dmTH.SetThreshold(context.TODO(), th); err != nil {
			t.Errorf("Error: %+v", err)
		}
	}
	//Test each threshold profile from cache
	for _, th := range ths {
		if temptTh, err := dmTH.GetThreshold(context.TODO(), th.Tenant,
			th.ID, true, false, utils.NonTransactional); err != nil {
			t.Errorf("Error: %+v", err)
		} else if !reflect.DeepEqual(th, temptTh) {
			t.Errorf("Expecting: %+v, received: %+v", th, temptTh)
		}
	}
	thMatched, unlock, err := thServ.matchingThresholdsForEvent(context.TODO(), argsGetThresholds[0].Tenant, argsGetThresholds[0])
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	unlock()
	if !reflect.DeepEqual(ths[0].Tenant, thMatched[0].threshold.Tenant) {
		t.Errorf("Expecting: %+v, received: %+v", ths[0].Tenant, thMatched[0].threshold.Tenant)
	} else if !reflect.DeepEqual(ths[0].ID, thMatched[0].threshold.ID) {
		t.Errorf("Expecting: %+v, received: %+v", ths[0].ID, thMatched[0].threshold.ID)
	} else if !reflect.DeepEqual(ths[0].Hits, thMatched[0].threshold.Hits) {
		t.Errorf("Expecting: %+v, received: %+v", ths[0].Hits, thMatched[0].threshold.Hits)
	}
	thMatched, unlock, err = thServ.matchingThresholdsForEvent(context.TODO(), argsGetThresholds[1].Tenant, argsGetThresholds[1])
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	unlock()
	if !reflect.DeepEqual(ths[1].Tenant, thMatched[0].threshold.Tenant) {
		t.Errorf("Expecting: %+v, received: %+v", ths[1].Tenant, thMatched[0].threshold.Tenant)
	} else if !reflect.DeepEqual(ths[1].ID, thMatched[0].threshold.ID) {
		t.Errorf("Expecting: %+v, received: %+v", ths[1].ID, thMatched[0].threshold.ID)
	} else if !reflect.DeepEqual(ths[1].Hits, thMatched[0].threshold.Hits) {
		t.Errorf("Expecting: %+v, received: %+v", ths[1].Hits, thMatched[0].threshold.Hits)
	}
	thMatched, unlock, err = thServ.matchingThresholdsForEvent(context.TODO(), argsGetThresholds[2].Tenant, argsGetThresholds[2])
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	unlock()
	if !reflect.DeepEqual(ths[2].Tenant, thMatched[0].threshold.Tenant) {
		t.Errorf("Expecting: %+v, received: %+v", ths[2].Tenant, thMatched[0].threshold.Tenant)
	} else if !reflect.DeepEqual(ths[2].ID, thMatched[0].threshold.ID) {
		t.Errorf("Expecting: %+v, received: %+v", ths[2].ID, thMatched[0].threshold.ID)
	} else if !reflect.DeepEqual(ths[2].Hits, thMatched[0].threshold.Hits) {
		t.Errorf("Expecting: %+v, received: %+v", ths[2].Hits, thMatched[0].threshold.Hits)
	}
}

func TestThresholdsUpdateThreshold(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	idb, err := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	if err != nil {
		t.Error(err)
	}
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: idb}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	cacheS := engine.NewCacheS(cfg, nil, nil, nil)
	dm.SetCache(cacheS)
	thp := &utils.ThresholdProfile{
		Tenant: "cgrates.org",
		ID:     "THUP1",
	}
	th := &utils.Threshold{
		Tenant: thp.Tenant,
		ID:     thp.ID,
		Hits:   5,
		Snooze: time.Now(),
	}
	expTh := &utils.Threshold{
		Tenant: thp.Tenant,
		ID:     thp.ID,
	}

	if err := dm.SetThresholdProfile(context.Background(), thp, true); err != nil {
		t.Fatal(err)
	}

	if th, err := dm.GetThreshold(context.Background(), thp.Tenant, thp.ID, false, false, utils.NonTransactional); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(expTh, th) {
		t.Errorf("Expected: %s, received: %s", utils.ToJSON(expTh), utils.ToJSON(th))
	}

	if err := dm.RemoveThreshold(context.Background(), th.Tenant, th.ID); err != nil {
		t.Fatal(err)
	}
	if err := dm.SetThresholdProfile(context.Background(), thp, true); err != nil {
		t.Fatal(err)
	}

	if th, err := dm.GetThreshold(context.Background(), thp.Tenant, thp.ID, false, false, utils.NonTransactional); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(expTh, th) {
		t.Errorf("Expected: %s, received: %s", utils.ToJSON(expTh), utils.ToJSON(th))
	}

	if err := dm.SetThreshold(context.Background(), th); err != nil {
		t.Fatal(err)
	}
	thp = &utils.ThresholdProfile{
		Tenant:  "cgrates.org",
		ID:      "THUP1",
		MaxHits: 1,
	}

	if err := dm.SetThresholdProfile(context.Background(), thp, true); err != nil {
		t.Fatal(err)
	}

	if th, err := dm.GetThreshold(context.Background(), thp.Tenant, thp.ID, false, false, utils.NonTransactional); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(expTh, th) {
		t.Errorf("Expected: %s, received: %s", utils.ToJSON(expTh), utils.ToJSON(th))
	}

	if err := dm.SetThreshold(context.Background(), th); err != nil {
		t.Fatal(err)
	}
	thp = &utils.ThresholdProfile{
		Tenant:  "cgrates.org",
		ID:      "THUP1",
		MaxHits: 1,
		MinHits: 1,
	}

	if err := dm.SetThresholdProfile(context.Background(), thp, true); err != nil {
		t.Fatal(err)
	}

	if th, err := dm.GetThreshold(context.Background(), thp.Tenant, thp.ID, false, false, utils.NonTransactional); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(expTh, th) {
		t.Errorf("Expected: %s, received: %s", utils.ToJSON(expTh), utils.ToJSON(th))
	}

	if err := dm.SetThreshold(context.Background(), th); err != nil {
		t.Fatal(err)
	}
	thp = &utils.ThresholdProfile{
		Tenant:   "cgrates.org",
		ID:       "THUP1",
		MaxHits:  1,
		MinHits:  1,
		MinSleep: 1,
	}

	if err := dm.SetThresholdProfile(context.Background(), thp, true); err != nil {
		t.Fatal(err)
	}

	if th, err := dm.GetThreshold(context.Background(), thp.Tenant, thp.ID, false, false, utils.NonTransactional); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(expTh, th) {
		t.Errorf("Expected: %s, received: %s", utils.ToJSON(expTh), utils.ToJSON(th))
	}
	if err := dm.RemoveThresholdProfile(context.Background(), thp.Tenant, thp.ID, true); err != nil {
		t.Fatal(err)
	}
	if _, err := dm.GetThreshold(context.Background(), thp.Tenant, thp.ID, false, false, utils.NonTransactional); err != utils.ErrNotFound {
		t.Fatal(err)
	}
}

func TestThresholdsStoreThresholdsOK(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	cacheS := engine.NewCacheS(cfg, nil, nil, nil)
	dm.SetCache(cacheS)
	tS := NewThresholdService(cfg, dm, cacheS, nil, nil)

	exp := &utils.Threshold{
		Tenant: "cgrates.org",
		ID:     "TH1",
		Hits:   2,
	}
	cacheS.SetWithoutReplicate(utils.CacheThresholds, "cgrates.org:TH1", exp, nil, true,
		utils.NonTransactional)
	tS.storedThresholds.Add("cgrates.org:TH1")
	tS.storeThresholds(context.Background())

	if rcv, err := tS.dm.GetThreshold(context.Background(), "cgrates.org", "TH1", true, false,
		utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>",
			utils.ToJSON(exp), utils.ToJSON(rcv))
	}

	cacheS.Remove(context.Background(), utils.CacheThresholds, "cgrates.org:TH1", true, utils.NonTransactional)
}

func TestThresholdsStoreThresholdsStoreThErr(t *testing.T) {
	tmpLogger := utils.Logger
	defer func() {
		utils.Logger = tmpLogger
	}()

	var buf bytes.Buffer
	utils.Logger = utils.NewStdLoggerWithWriter(&buf, "", 4)

	cfg := config.NewDefaultCGRConfig()
	cacheS := engine.NewCacheS(cfg, nil, nil, nil)
	tS := NewThresholdService(cfg, nil, cacheS, nil, nil)

	value := &utils.Threshold{
		Tenant: "cgrates.org",
		ID:     "TH1",
		Hits:   2,
	}

	cacheS.SetWithoutReplicate(utils.CacheThresholds, "TH1", value, nil, true,
		utils.NonTransactional)
	tS.storedThresholds.Add("TH1")
	exp := utils.StringSet{
		"TH1": struct{}{},
	}
	expLog := `[WARNING] <ThresholdS> failed saving Threshold with tenant: cgrates.org and ID: TH1, error: NO_DATABASE_CONNECTION`
	tS.storeThresholds(context.Background())

	if !reflect.DeepEqual(tS.storedThresholds, exp) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", exp, tS.storedThresholds)
	}
	if rcvLog := buf.String(); !strings.Contains(rcvLog, expLog) {
		t.Errorf("expected log <%+v>\n to be in included in: <%+v>", expLog, rcvLog)
	}

	cacheS.Remove(context.Background(), utils.CacheThresholds, "TH1", true, utils.NonTransactional)
}

func TestThresholdsStoreThresholdsCacheGetErr(t *testing.T) {
	tmpLogger := utils.Logger
	defer func() {
		utils.Logger = tmpLogger
	}()

	var buf bytes.Buffer
	utils.Logger = utils.NewStdLoggerWithWriter(&buf, "", 4)

	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	cacheS := engine.NewCacheS(cfg, nil, nil, nil)
	dm.SetCache(cacheS)
	tS := NewThresholdService(cfg, dm, cacheS, nil, nil)

	value := &utils.Threshold{
		Tenant: "cgrates.org",
		ID:     "TH1",
		Hits:   2,
	}

	cacheS.SetWithoutReplicate(utils.CacheThresholds, "TH2", value, nil, true,
		utils.NonTransactional)
	tS.storedThresholds.Add("TH1")
	expLog := `[WARNING] <ThresholdS> failed retrieving from cache threshold with ID: TH1`
	tS.storeThresholds(context.Background())

	if rcvLog := buf.String(); !strings.Contains(rcvLog, expLog) {
		t.Errorf("expected <%+v> \nto be included in: <%+v>", expLog, rcvLog)
	}

	cacheS.Remove(context.Background(), utils.CacheThresholds, "TH2", true, utils.NonTransactional)
}

func TestThresholdsProcessEventStoreThOK(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.ThresholdSCfg().StoreInterval = -1
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	cacheS := engine.NewCacheS(cfg, nil, nil, nil)
	dm.SetCache(cacheS)
	filterS := engine.NewFilterS(cfg, nil, dm)
	tS := NewThresholdService(cfg, dm, cacheS, filterS, nil)

	thPrf := &utils.ThresholdProfile{
		Tenant:    "cgrates.org",
		ID:        "TH2",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		MinHits:   2,
		MaxHits:   5,
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
		Blocker: true,
	}

	if err := dm.SetThresholdProfile(context.Background(), thPrf, true); err != nil {
		t.Error(err)
	}

	args := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "ThdProcessEvent",
		Event: map[string]any{
			utils.AccountField: "1001",
		},
		APIOpts: map[string]any{
			utils.OptsThresholdsProfileIDs: []string{"TH2"},
		},
	}
	exp := &utils.Threshold{
		Tenant: "cgrates.org",
		ID:     "TH2",
		Hits:   1,
	}
	expIDs := []string{"TH2"}
	if rcvIDs, err := tS.processEvent(context.Background(), args.Tenant, args); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcvIDs, expIDs) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", expIDs, rcvIDs)
	} else if rcv, err := tS.dm.GetThreshold(context.Background(), "cgrates.org", "TH2", true, false,
		utils.NonTransactional); err != nil {
		t.Error(err)
	} else {
		rcv.Snooze = time.Time{}
		if !reflect.DeepEqual(rcv, exp) {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>", exp, rcv)
		}
	}

}

func TestThresholdsProcessEventMaxHitsDMErr(t *testing.T) {
	tmpLogger := utils.Logger
	var buf bytes.Buffer
	utils.Logger = utils.NewStdLoggerWithWriter(&buf, "", 4)
	defer func() {
		utils.Logger = tmpLogger
	}()

	cfg := config.NewDefaultCGRConfig()
	cfg.RPCConns()["test"] = &config.RPCConn{Conns: []*config.RemoteHost{{}}}
	cfg.CacheCfg().ReplicationConns = []string{"test"}
	cfg.CacheCfg().Partitions[utils.CacheThresholds].Replicate = true
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	cM := engine.NewConnManager(cfg)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, cM)
	filterS := engine.NewFilterS(cfg, nil, dm)
	cacheS := engine.NewCacheS(cfg, dm, cM, nil)
	dm.SetCache(cacheS)
	cM.SetCache(cacheS)
	tS := NewThresholdService(cfg, dm, cacheS, filterS, cM)
	thPrf := &utils.ThresholdProfile{
		Tenant:    "cgrates.org",
		ID:        "TH3",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		MinHits:   2,
		MaxHits:   5,
		Weights: utils.DynamicWeights{
			{
				Weight: 10.0,
			},
		},
		ActionProfileIDs: []string{utils.MetaNone},
		Blocker:          true,
	}
	th := &utils.Threshold{
		Tenant: "cgrates.org",
		ID:     "TH3",
		Hits:   4,
	}

	if err := dm.SetThresholdProfile(context.Background(), thPrf, true); err != nil {
		t.Error(err)
	}
	if err := dm.SetThreshold(context.Background(), th); err != nil {
		t.Error(err)
	}
	cacheS.SetWithoutReplicate(utils.CacheThresholdProfiles, thPrf.TenantID(), thPrf, nil, true, utils.NonTransactional)
	cacheS.SetWithoutReplicate(utils.CacheThresholds, th.TenantID(), th, nil, true, utils.NonTransactional)

	args := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "ThdProcessEvent",
		Event: map[string]any{
			utils.AccountField: "1001",
		},
		APIOpts: map[string]any{
			utils.OptsThresholdsProfileIDs: []string{"TH3"},
		},
	}

	if _, err := tS.processEvent(context.Background(), args.Tenant, args); err == nil ||
		err != utils.ErrPartiallyExecuted {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>",
			utils.ErrPartiallyExecuted, err)
	}

}

func TestThresholdsProcessEventNotFound(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	cacheS := engine.NewCacheS(cfg, nil, nil, nil)
	dm.SetCache(cacheS)
	filterS := engine.NewFilterS(cfg, nil, dm)
	tS := NewThresholdService(cfg, dm, cacheS, filterS, nil)

	thPrf := &utils.ThresholdProfile{
		Tenant:    "cgrates.org",
		ID:        "TH5",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		MinHits:   2,
		MaxHits:   5,
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
		Blocker: true,
	}
	th := &utils.Threshold{
		Tenant: "cgrates.org",
		ID:     "TH5",
		Hits:   2,
	}

	if err := dm.SetThresholdProfile(context.Background(), thPrf, true); err != nil {
		t.Error(err)
	}
	if err := dm.SetThreshold(context.Background(), th); err != nil {
		t.Error(err)
	}

	args := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "ThdProcessEvent",
		Event: map[string]any{
			utils.AccountField: "1001",
		},
		APIOpts: map[string]any{
			utils.OptsThresholdsProfileIDs: []string{"TH6"},
		},
	}

	if _, err := tS.processEvent(context.Background(), args.Tenant, args); err == nil ||
		err != utils.ErrNotFound {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}

}

func TestThresholdMatchingThresholdForEventLocks2(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()

	defer func() {
		guardian.Guardian = guardian.New()
	}()
	cacheS := engine.NewCacheS(cfg, nil, nil, nil)
	db, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: db}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	dm.SetCache(cacheS)
	cfg.ThresholdSCfg().StoreInterval = 1
	cfg.ThresholdSCfg().StringIndexedFields = nil
	cfg.ThresholdSCfg().PrefixIndexedFields = nil
	tS := NewThresholdService(cfg, dm, cacheS,
		engine.NewFilterS(cfg, nil, dm), nil)

	ids := utils.StringSet{}
	for i := 0; i < 10; i++ {
		thPrf := &utils.ThresholdProfile{
			Tenant: "cgrates.org",
			ID:     fmt.Sprintf("TH%d", i),
			Weights: utils.DynamicWeights{
				{
					Weight: 20.00,
				},
			},
			MaxHits: 5,
		}
		dm.SetThresholdProfile(context.Background(), thPrf, true)
		ids.Add(thPrf.ID)
	}
	thPrf := &utils.ThresholdProfile{
		Tenant:    "cgrates.org",
		ID:        "TH20",
		FilterIDs: []string{"FLTR_TH_201"},
		Weights: utils.DynamicWeights{
			{
				Weight: 20.00,
			},
		},
		MaxHits: 5,
	}
	err := db.SetThresholdProfileDrv(context.Background(), thPrf)
	if err != nil {
		t.Fatal(err)
	}
	ids.Add(thPrf.ID)
	ev := &utils.CGREvent{
		APIOpts: map[string]any{
			utils.OptsThresholdsProfileIDs: ids.AsSlice(),
		},
	}
	_, _, err = tS.matchingThresholdsForEvent(context.Background(), "cgrates.org", ev)
	expErr := utils.ErrPrefixNotFound(thPrf.FilterIDs[0])
	if err == nil || err.Error() != expErr.Error() {
		t.Errorf("Expected error: %s ,received: %+v", expErr, err)
	}
}

func TestThresholdMatchingThresholdForEventLocks3(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()

	defer func() {
		guardian.Guardian = guardian.New()
	}()
	cacheS := engine.NewCacheS(cfg, nil, nil, nil)
	db := &engine.DataDBMock{
		GetThresholdProfileDrvF: func(ctx *context.Context, tnt, id string) (*utils.ThresholdProfile, error) {
			if id == "TH1" {
				return nil, utils.ErrNotImplemented
			}
			thPrf := &utils.ThresholdProfile{
				Tenant:  "cgrates.org",
				ID:      id,
				MaxHits: 5,
				Weights: utils.DynamicWeights{
					{
						Weight: 20.00,
					},
				},
			}
			cacheS.Set(ctx, utils.CacheThresholds, thPrf.TenantID(), &utils.Threshold{
				Tenant: thPrf.Tenant,
				ID:     thPrf.ID,
			}, nil, true, utils.NonTransactional)
			return thPrf, nil
		},
	}
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: db}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	dm.SetCache(cacheS)
	cfg.ThresholdSCfg().StoreInterval = 1
	cfg.ThresholdSCfg().StringIndexedFields = nil
	cfg.ThresholdSCfg().PrefixIndexedFields = nil
	tS := NewThresholdService(cfg, dm, cacheS,
		engine.NewFilterS(cfg, nil, dm), nil)

	ids := utils.StringSet{}
	for i := 0; i < 10; i++ {
		ids.Add(fmt.Sprintf("TH%d", i))
	}
	ev := &utils.CGREvent{
		APIOpts: map[string]any{
			utils.OptsThresholdsProfileIDs: ids.AsSlice(),
		},
	}
	_, _, err := tS.matchingThresholdsForEvent(context.Background(), "cgrates.org", ev)
	if err != utils.ErrNotImplemented {
		t.Fatalf("Error: %+v", err)
	}
}

func TestThresholdMatchingThresholdForEventLocks5(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()

	defer func() {
		guardian.Guardian = guardian.New()
	}()
	cacheS := engine.NewCacheS(cfg, nil, nil, nil)
	db, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: db}, cfg.DbCfg())
	connMgr := engine.NewConnManager(cfg)
	connMgr.SetCache(cacheS)
	dm := engine.NewDataManager(dbCM, cfg, connMgr)
	dm.SetCache(cacheS)
	cfg.ThresholdSCfg().StoreInterval = 1
	cfg.ThresholdSCfg().StringIndexedFields = nil
	cfg.ThresholdSCfg().PrefixIndexedFields = nil
	cfg.RPCConns()["test"] = &config.RPCConn{Conns: []*config.RemoteHost{{}}}
	cfg.DbCfg().DBConns[utils.MetaDefault].RmtConns = []string{"test"}
	cfg.DbCfg().Items[utils.CacheThresholds].Remote = true
	tS := NewThresholdService(cfg, dm, cacheS,
		engine.NewFilterS(cfg, nil, dm), nil)

	ids := utils.StringSet{}
	for i := 0; i < 10; i++ {
		thPrf := &utils.ThresholdProfile{
			Tenant: "cgrates.org",
			ID:     fmt.Sprintf("TH%d", i),
			Weights: utils.DynamicWeights{
				{
					Weight: 20.00,
				},
			},
			MaxHits: 5,
		}
		dm.SetThresholdProfile(context.Background(), thPrf, true)
		ids.Add(thPrf.ID)
	}
	ev := &utils.CGREvent{
		APIOpts: map[string]any{
			utils.OptsThresholdsProfileIDs: ids.AsSlice(),
		},
	}
	dm.RemoveThreshold(context.Background(), "cgrates.org", "TH1")
	_, _, err := tS.matchingThresholdsForEvent(context.Background(), "cgrates.org", ev)
	if err != utils.ErrDisconnected {
		t.Errorf("Error: %+v", err)
	}
}

func TestThresholdsRunBackupStop(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.ThresholdSCfg().StoreInterval = 5 * time.Millisecond
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	cacheS := engine.NewCacheS(cfg, nil, nil, nil)
	dm.SetCache(cacheS)
	tnt := "cgrates.org"
	thID := "Th1"
	tS := &ThresholdS{
		dm:    dm,
		cache: cacheS,
		storedThresholds: utils.StringSet{
			thID: struct{}{},
		},
		cfg:        cfg,
		stopBackup: make(chan struct{}),
	}
	value := &utils.Threshold{
		Tenant: tnt,
		ID:     thID,
	}
	cacheS.SetWithoutReplicate(utils.CacheThresholds, thID, value, nil, true, "")

	// Backup loop checks for the state of the stopBackup
	// channel after storing the threshold. Channel can be
	// safely closed beforehand.
	close(tS.stopBackup)
	tS.StartLoop(context.Background())
	tS.backupLoop.Wait()

	want := &utils.Threshold{
		Tenant: tnt,
		ID:     thID,
	}
	if got, err := tS.dm.GetThreshold(context.Background(), tnt, thID, true, false, utils.NonTransactional); err != nil {
		t.Errorf("dm.GetThreshold(%q,%q): got unexpected err=%v", tnt, thID, err)
	} else if !reflect.DeepEqual(got, want) {
		t.Errorf("dm.GetThreshold(%q,%q) = %v, want %v", tnt, thID, got, want)
	}
}

func TestThresholdsReload(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.ThresholdSCfg().StoreInterval = 5 * time.Millisecond
	cacheS := engine.NewCacheS(cfg, nil, nil, nil)
	synctest.Test(t, func(*testing.T) {
		tS := NewThresholdService(cfg, nil, cacheS, nil, nil)
		tS.StartLoop(context.Background())
		tS.Reload(context.Background())
		tS.Shutdown(context.Background())
		tS.Shutdown(context.Background())
		tS.Reload(context.Background())
	})
}

func TestThresholdsReloadShutdownConcurrent(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.ThresholdSCfg().StoreInterval = 5 * time.Millisecond
	cacheS := engine.NewCacheS(cfg, nil, nil, nil)
	synctest.Test(t, func(*testing.T) {
		tS := NewThresholdService(cfg, nil, cacheS, nil, nil)
		tS.StartLoop(context.Background())
		var wg sync.WaitGroup
		wg.Go(func() { tS.Reload(context.Background()) })
		wg.Go(func() { tS.Shutdown(context.Background()) })
		wg.Wait()
	})
}

func TestThresholdsStartLoop(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cacheS := engine.NewCacheS(cfg, nil, nil, nil)
	synctest.Test(t, func(*testing.T) {
		tS := NewThresholdService(cfg, nil, cacheS, nil, nil)
		tS.StartLoop(context.Background())
		tS.backupLoop.Wait()
	})
}

func TestThresholdsMatchingThresholdsForEventNotFoundErr(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	cacheS := engine.NewCacheS(cfg, dm, nil, nil)
	dm.SetCache(cacheS)
	filterS := engine.NewFilterS(cfg, nil, dm)
	tS := NewThresholdService(cfg, dm, cacheS, filterS, nil)

	thPrf1 := &utils.ThresholdProfile{
		Tenant:    "cgrates.org",
		ID:        "TH1",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		MinHits:   2,
		MaxHits:   5,
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
		Blocker: true,
	}
	if err := dm.SetThresholdProfile(context.Background(), thPrf1, true); err != nil {
		t.Error(err)
	}

	args := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "TestEvent",
		Event:  map[string]any{},
	}

	if _, _, err := tS.matchingThresholdsForEvent(context.Background(), "cgrates.org", args); err == nil ||
		err != utils.ErrNotFound {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}
}

func TestThresholdsStoreThresholdCacheSetErr(t *testing.T) {
	tmpLogger := utils.Logger
	var buf bytes.Buffer
	utils.Logger = utils.NewStdLoggerWithWriter(&buf, "", 4)

	defer func() {
		utils.Logger = tmpLogger
	}()

	cfg := config.NewDefaultCGRConfig()
	cfg.CacheCfg().ReplicationConns = []string{"test"}
	cfg.CacheCfg().Partitions[utils.CacheThresholds].Replicate = true
	cfg.RPCConns()["test"] = &config.RPCConn{Conns: []*config.RemoteHost{{}}}
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	cM := engine.NewConnManager(cfg)
	dm := engine.NewDataManager(dbCM, cfg, cM)
	cacheS := engine.NewCacheS(cfg, dm, cM, nil)
	dm.SetCache(cacheS)
	cM.SetCache(cacheS)
	filterS := engine.NewFilterS(cfg, nil, dm)
	tS := NewThresholdService(cfg, dm, cacheS, filterS, cM)

	th := &utils.Threshold{
		Tenant: "cgrates.org",
		ID:     "TH1",
	}
	cacheS.SetWithoutReplicate(utils.CacheThresholds, th.TenantID(), th, nil, true, utils.NonTransactional)
	expLog := `[WARNING] <ThresholdS> failed caching Threshold with ID: cgrates.org:TH1, error: DISCONNECTED`
	if err := tS.storeThreshold(context.Background(), th); err == nil ||
		err.Error() != utils.ErrDisconnected.Error() {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrDisconnected, err)
	} else if rcv := buf.String(); !strings.Contains(rcv, expLog) {
		t.Errorf("expected log <%+v> to be included in <%+v>", expLog, rcv)
	}
}

func TestThreholdsMatchingThresholdsForEventDoesNotPass(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	cacheS := engine.NewCacheS(cfg, dm, nil, nil)
	dm.SetCache(cacheS)
	filterS := engine.NewFilterS(cfg, nil, dm)
	tS := NewThresholdService(cfg, dm, cacheS, filterS, nil)

	thPrf1 := &utils.ThresholdProfile{
		Tenant:    "cgrates.org",
		ID:        "TH1",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		MinHits:   2,
		MaxHits:   5,
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
		Blocker: true,
	}
	if err := dm.SetThresholdProfile(context.Background(), thPrf1, true); err != nil {
		t.Error(err)
	}

	args := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "TestEvent",
		Event: map[string]any{
			utils.AccountField: "1002",
		},
		APIOpts: map[string]any{
			utils.OptsThresholdsProfileIDs: []string{"TH1"},
		},
	}
	if _, _, err := tS.matchingThresholdsForEvent(context.Background(), "cgrates.org",
		args); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("expected: <%+v>, received: <%+v>", utils.ErrNotFound, err)
	}
}

func TestThresholdsProcessEventIgnoreFilters(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	cacheS := engine.NewCacheS(cfg, nil, nil, nil)
	dm.SetCache(cacheS)
	filterS := engine.NewFilterS(cfg, nil, dm)
	tS := NewThresholdService(cfg, dm, cacheS, filterS, nil)
	cfg.ThresholdSCfg().Opts.ProfileIgnoreFilters = []*config.DynamicBoolOpt{
		config.NewDynamicBoolOpt(nil, "", true, nil),
	}
	thPrf := &utils.ThresholdProfile{
		Tenant:    "cgrates.org",
		ID:        "TH",
		FilterIDs: []string{"*string:~*req.Threshold:testThresholdValue"},
	}
	th := &utils.Threshold{
		Tenant: "cgrates.org",
		ID:     "TH",
	}

	if err := dm.SetThresholdProfile(context.Background(), thPrf, true); err != nil {
		t.Error(err)
	}
	if err := dm.SetThreshold(context.Background(), th); err != nil {
		t.Error(err)
	}
	// testing if the profile matches with profile ignore filters on false
	args := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "ThdProcessEvent",
		Event: map[string]any{
			"Threshold": "testThresholdValue",
		},
		APIOpts: map[string]any{
			utils.OptsThresholdsProfileIDs: []string{"TH"},
			utils.MetaProfileIgnoreFilters: false,
		},
	}
	exp := []string{"TH"}
	if rcv, err := tS.processEvent(context.Background(), args.Tenant, args); err != nil {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", nil, err)
	} else if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", exp, rcv)
	}
	// testing if the profile matches with profile ignore filters on true
	args2 := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "ThdProcessEvent",
		Event: map[string]any{
			"Threshold": "testThresholdValue2",
		},
		APIOpts: map[string]any{
			utils.OptsThresholdsProfileIDs: []string{"TH"},
			utils.MetaProfileIgnoreFilters: true,
		},
	}
	exp2 := []string{"TH"}
	if rcv2, err := tS.processEvent(context.Background(), args2.Tenant, args2); err != nil {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", nil, err)
	} else if !reflect.DeepEqual(rcv2, exp) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", exp2, rcv2)
	}
}

func TestThresholdsProcessEventIgnoreFiltersErr(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	cacheS := engine.NewCacheS(cfg, nil, nil, nil)
	dm.SetCache(cacheS)
	filterS := engine.NewFilterS(cfg, nil, dm)
	tS := NewThresholdService(cfg, dm, cacheS, filterS, nil)
	cfg.ThresholdSCfg().Opts.ProfileIgnoreFilters = []*config.DynamicBoolOpt{
		config.NewDynamicBoolOpt(nil, "", true, nil),
	}
	thPrf := &utils.ThresholdProfile{
		Tenant:    "cgrates.org",
		ID:        "TH",
		FilterIDs: []string{"*string:~*req.Threshold:testThresholdValue"},
	}
	th := &utils.Threshold{
		Tenant: "cgrates.org",
		ID:     "TH",
	}

	if err := dm.SetThresholdProfile(context.Background(), thPrf, true); err != nil {
		t.Error(err)
	}
	if err := dm.SetThreshold(context.Background(), th); err != nil {
		t.Error(err)
	}
	args := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "ThdProcessEvent",
		Event: map[string]any{
			"Threshold": "testThresholdValue",
		},
		APIOpts: map[string]any{
			utils.OptsThresholdsProfileIDs: []string{"TH"},
			utils.MetaProfileIgnoreFilters: time.Second,
		},
	}

	if _, err := tS.processEvent(context.Background(), args.Tenant, args); err == nil || err.Error() != "cannot convert field: 1s to bool" {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", nil, err)
	}

}

func TestThresholdSmatchingThresholdsForEventGetOptsErr(t *testing.T) {

	cfg := config.NewDefaultCGRConfig()
	cfg.ThresholdSCfg().Opts.ProfileIDs = []*config.DynamicStringSliceOpt{
		{
			FilterIDs: []string{"*string"},
			Tenant:    "cgrates.org",
			Values:    []string{"ProfIdVal"},
		},
	}

	dataDB, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	cM := engine.NewConnManager(cfg)
	cacheS := engine.NewCacheS(cfg, nil, nil, nil)
	cM.SetCache(cacheS)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: dataDB}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, cM)
	dm.SetCache(cacheS)
	filterS := engine.NewFilterS(cfg, cM, dm)
	tS := &ThresholdS{
		dm: dm,
		storedThresholds: utils.StringSet{
			"Th1": struct{}{},
		},
		cfg:        cfg,
		filters:    filterS,
		stopBackup: make(chan struct{}),
	}

	args := &utils.CGREvent{
		ID:     "TestV1DebitID",
		Tenant: "cgrates.org",
		Event: map[string]any{
			utils.AccountField: "1004",
		},
		APIOpts: map[string]any{
			utils.MetaUsage: "3m",
		},
	}

	expErr := "inline parse error for string: <*string>"
	if _, _, err := tS.matchingThresholdsForEvent(context.Background(), args.Tenant, args); err.Error() != expErr || err == nil {
		t.Errorf("Expected error <%v>, Received <%v>", expErr, err)
	}

}

func TestThresholdSmatchingThresholdsForEventWeightErr(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.ThresholdSCfg().Opts.ProfileIDs = []*config.DynamicStringSliceOpt{
		{
			Tenant: "cgrates.org",
			Values: []string{"ACC1"},
		},
	}
	cfg.ThresholdSCfg().Opts.ProfileIgnoreFilters = []*config.DynamicBoolOpt{
		config.NewDynamicBoolOpt(nil, "cgrates.org", true, nil),
	}
	cM := engine.NewConnManager(cfg)
	cacheS := engine.NewCacheS(cfg, nil, nil, nil)
	cM.SetCache(cacheS)
	db := &engine.DataDBMock{
		GetThresholdProfileDrvF: func(ctx *context.Context, tenant, id string) (tp *utils.ThresholdProfile, err error) {
			return &utils.ThresholdProfile{
				Tenant:           "cgrates.org",
				ID:               "THD_2",
				FilterIDs:        []string{"*string:~*req.Account:1001"},
				ActionProfileIDs: []string{"actPrfID"},
				MaxHits:          7,
				MinHits:          0,
				Weights: []*utils.DynamicWeight{
					{
						FilterIDs: []string{"*stirng:~*req.Account:1001"},
						Weight:    64,
					},
				},
				Async: true,
			}, nil
		},
		GetThresholdDrvF: func(ctx *context.Context, tenant, id string) (*utils.Threshold, error) {
			return &utils.Threshold{}, nil
		},
	}
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: db}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, cM)
	dm.SetCache(cacheS)
	filterS := engine.NewFilterS(cfg, cM, dm)
	tS := &ThresholdS{
		dm: dm,
		storedThresholds: utils.StringSet{
			"Th1": struct{}{},
		},
		cfg:        cfg,
		filters:    filterS,
		stopBackup: make(chan struct{}),
	}

	args := &utils.CGREvent{
		ID:     "EvID",
		Tenant: "cgrates.org",
		Event: map[string]any{
			utils.AccountField: "1004",
		},
		APIOpts: map[string]any{
			utils.MetaUsage: "3m",
		},
	}

	expErr := "Unsupported filter Type: *stirng"
	if _, _, err := tS.matchingThresholdsForEvent(context.Background(), args.Tenant, args); err.Error() != expErr || err == nil {
		t.Errorf("Expected error <%v>, Received <%v>", expErr, err)
	}
}
