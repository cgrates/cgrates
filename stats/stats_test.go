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
package stats

import (
	"bytes"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"sync"
	"testing"
	"testing/synctest"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/guardian"
	"github.com/cgrates/rpcclient"
)

func mustNewStatSum(minItems uint64, fieldName string, filterIDs []string) utils.StatMetric {
	sum, err := utils.NewStatSum(minItems, fieldName, filterIDs)
	if err != nil {
		panic(err)
	}
	return sum
}

var (
	testStatsPrfs = []*utils.StatQueueProfile{
		{
			Tenant:      "cgrates.org",
			ID:          "StatQueueProfile1",
			FilterIDs:   []string{"FLTR_STATS_1", "*ai:*now:2014-07-14T14:25:00Z"},
			QueueLength: 10,
			TTL:         10 * time.Second,
			Metrics: []*utils.MetricWithFilters{
				{
					MetricID: "*sum#~*req.Usage",
				},
			},
			ThresholdIDs: []string{},
			Blockers:     utils.DynamicBlockers{{Blocker: false}},
			Stored:       true,
			Weights: utils.DynamicWeights{
				{
					Weight: 20,
				},
			},
			MinItems: 1,
		},
		{
			Tenant:      "cgrates.org",
			ID:          "StatQueueProfile2",
			FilterIDs:   []string{"FLTR_STATS_2", "*ai:*now:2014-07-14T14:25:00Z"},
			QueueLength: 10,
			TTL:         10 * time.Second,
			Metrics: []*utils.MetricWithFilters{
				{
					MetricID: "*sum#~*req.Usage",
				},
			},
			ThresholdIDs: []string{},
			Blockers:     utils.DynamicBlockers{{Blocker: false}},
			Stored:       true,
			Weights: utils.DynamicWeights{
				{
					Weight: 20,
				},
			},
			MinItems: 1,
		},
		{
			Tenant:      "cgrates.org",
			ID:          "StatQueueProfilePrefix",
			FilterIDs:   []string{"FLTR_STATS_3", "*ai:*now:2014-07-14T14:25:00Z"},
			QueueLength: 10,
			TTL:         10 * time.Second,
			Metrics: []*utils.MetricWithFilters{
				{
					MetricID: "*sum#~*req.Usage",
				},
			},
			ThresholdIDs: []string{},
			Blockers:     utils.DynamicBlockers{{Blocker: false}},
			Stored:       true,
			Weights: utils.DynamicWeights{
				{
					Weight: 20,
				},
			},
			MinItems: 1,
		},
	}
	testStatsQ = []*utils.StatQueue{
		{
			Tenant: "cgrates.org",
			ID:     "StatQueueProfile1",
			SQMetrics: map[string]utils.StatMetric{
				utils.MetaSum: mustNewStatSum(1, "~*req.Usage", nil),
			},
		},
		{
			Tenant: "cgrates.org",
			ID:     "StatQueueProfile2",
			SQMetrics: map[string]utils.StatMetric{
				utils.MetaSum: mustNewStatSum(1, "~*req.Usage", nil),
			},
		},
		{
			Tenant: "cgrates.org",
			ID:     "StatQueueProfilePrefix",
			SQMetrics: map[string]utils.StatMetric{
				utils.MetaSum: mustNewStatSum(1, "~*req.Usage", nil),
			},
		},
	}
	testStatsArgs = []*utils.CGREvent{
		{
			Tenant: "cgrates.org",
			ID:     "event1",
			Event: map[string]any{
				"Stats":          "StatQueueProfile1",
				utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
				"UsageInterval":  "1s",
				"PddInterval":    "1s",
				"Weight":         "9.0",
				utils.Usage:      135 * time.Second,
				utils.Cost:       123.0,
			},
		},
		{
			Tenant: "cgrates.org",
			ID:     "event2",
			Event: map[string]any{
				"Stats":          "StatQueueProfile2",
				utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
				"UsageInterval":  "1s",
				"PddInterval":    "1s",
				"Weight":         "15.0",
				utils.Usage:      45 * time.Second,
			},
		},
		{
			Tenant: "cgrates.org",
			ID:     "event3",
			Event: map[string]any{
				"Stats":     "StatQueueProfilePrefix",
				utils.Usage: 30 * time.Second,
			},
		},
	}
)

func prepareStatsData(t *testing.T, dm *engine.DataManager) {
	cfg := config.NewDefaultCGRConfig()
	if err := dm.SetFilter(context.Background(), &engine.Filter{
		Tenant: cfg.GeneralCfg().DefaultTenant,
		ID:     "FLTR_STATS_1",
		Rules: []*engine.FilterRule{
			{
				Type:    utils.MetaString,
				Element: "~*req.Stats",
				Values:  []string{"StatQueueProfile1"},
			},
			{
				Type:    utils.MetaGreaterOrEqual,
				Element: "~*req.UsageInterval",
				Values:  []string{time.Second.String()},
			},
			{
				Type:    utils.MetaGreaterOrEqual,
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Usage,
				Values:  []string{time.Second.String()},
			},
			{
				Type:    utils.MetaGreaterOrEqual,
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Weight,
				Values:  []string{"9.0"},
			},
		},
	}, true); err != nil {
		t.Fatal(err)
	}
	if err := dm.SetFilter(context.Background(), &engine.Filter{
		Tenant: cfg.GeneralCfg().DefaultTenant,
		ID:     "FLTR_STATS_2",
		Rules: []*engine.FilterRule{
			{
				Type:    utils.MetaString,
				Element: "~*req.Stats",
				Values:  []string{"StatQueueProfile2"},
			},
			{
				Type:    utils.MetaGreaterOrEqual,
				Element: "~*req.PddInterval",
				Values:  []string{time.Second.String()},
			},
			{
				Type:    utils.MetaGreaterOrEqual,
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Usage,
				Values:  []string{time.Second.String()},
			},
			{
				Type:    utils.MetaGreaterOrEqual,
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Weight,
				Values:  []string{"15.0"},
			},
		},
	}, true); err != nil {
		t.Fatal(err)
	}
	if err := dm.SetFilter(context.Background(), &engine.Filter{
		Tenant: cfg.GeneralCfg().DefaultTenant,
		ID:     "FLTR_STATS_3",
		Rules: []*engine.FilterRule{
			{
				Type:    utils.MetaPrefix,
				Element: "~*req.Stats",
				Values:  []string{"StatQueueProfilePrefix"},
			},
		},
	}, true); err != nil {
		t.Fatal(err)
	}
	for _, statQueueProfile := range testStatsPrfs {
		dm.SetStatQueueProfile(context.Background(), statQueueProfile, true)
	}
	for i, statQueue := range testStatsQ {
		statSum, err := utils.NewStatMetric("*sum#~*req.Usage", uint64(testStatsPrfs[i].MinItems), []string{})
		if err != nil {
			t.Fatal(err)
		}
		statQueue.SQMetrics["*sum#~*req.Usage"] = statSum
		dm.SetStatQueue(context.Background(), statQueue)
	}
	//Test each statQueueProfile from cache
	for _, sqp := range testStatsPrfs {
		if tempStat, err := dm.GetStatQueueProfile(context.Background(), sqp.Tenant,
			sqp.ID, true, false, utils.NonTransactional); err != nil {
			t.Errorf("Error: %+v", err)
		} else if !reflect.DeepEqual(sqp, tempStat) {
			t.Errorf("Expecting: %+v, received: %+v", sqp, tempStat)
		}
	}
}

func TestNewStatService(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	cacheS := engine.NewCacheS(cfg, nil, nil, nil)
	dm.SetCache(cacheS)
	fltrS := engine.NewFilterS(cfg, nil, dm)
	sSrv := &StatS{
		dm:               dm,
		filters:          fltrS,
		cfg:              cfg,
		storedStatQueues: make(utils.StringSet),
	}
	result := NewStatService(cfg, dm, cacheS, fltrS, nil)
	if !reflect.DeepEqual(sSrv.dm, result.dm) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", sSrv.dm, result.dm)
	}
	if !reflect.DeepEqual(sSrv.filters, result.filters) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", sSrv.filters, result.filters)
	}
	if !reflect.DeepEqual(sSrv.cfg, result.cfg) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", sSrv.cfg, result.cfg)
	}
	if !reflect.DeepEqual(sSrv.storedStatQueues, sSrv.storedStatQueues) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", sSrv.storedStatQueues, sSrv.storedStatQueues)
	}
}

func TestStatQueuesMatchingStatQueuesForEvent(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dmSTS := engine.NewDataManager(dbCM, cfg, nil)
	cacheS := engine.NewCacheS(cfg, nil, nil, nil)
	dmSTS.SetCache(cacheS)

	cfg.StatSCfg().StoreInterval = 1
	cfg.StatSCfg().StringIndexedFields = nil
	cfg.StatSCfg().PrefixIndexedFields = nil
	statService := NewStatService(cfg, dmSTS, cacheS,
		engine.NewFilterS(cfg, nil, dmSTS), nil)
	prepareStatsData(t, dmSTS)
	msq, unlock, err := statService.matchingStatQueuesForEvent(context.TODO(), testStatsArgs[0].Tenant, testStatsArgs[0])
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	unlock()
	if !reflect.DeepEqual(testStatsQ[0].Tenant, msq[0].statQueue.Tenant) {
		t.Errorf("Expecting: %+v, received: %+v", testStatsQ[0].Tenant, msq[0].statQueue.Tenant)
	} else if !reflect.DeepEqual(testStatsQ[0].ID, msq[0].statQueue.ID) {
		t.Errorf("Expecting: %+v, received: %+v", testStatsQ[0].ID, msq[0].statQueue.ID)
	} else if !reflect.DeepEqual(testStatsPrfs[0], msq[0].profile) {
		t.Errorf("Expecting: %+v, received: %+v", testStatsPrfs[0], msq[0].profile)
	}
	msq, unlock, err = statService.matchingStatQueuesForEvent(context.TODO(), testStatsArgs[1].Tenant, testStatsArgs[1])
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	unlock()
	if !reflect.DeepEqual(testStatsQ[1].Tenant, msq[0].statQueue.Tenant) {
		t.Errorf("Expecting: %+v, received: %+v", testStatsQ[1].Tenant, msq[0].statQueue.Tenant)
	} else if !reflect.DeepEqual(testStatsQ[1].ID, msq[0].statQueue.ID) {
		t.Errorf("Expecting: %+v, received: %+v", testStatsQ[1].ID, msq[0].statQueue.ID)
	} else if !reflect.DeepEqual(testStatsPrfs[1], msq[0].profile) {
		t.Errorf("Expecting: %+v, received: %+v", testStatsPrfs[1], msq[0].profile)
	}
	msq, unlock, err = statService.matchingStatQueuesForEvent(context.TODO(), testStatsArgs[2].Tenant, testStatsArgs[2])
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	unlock()
	if !reflect.DeepEqual(testStatsQ[2].Tenant, msq[0].statQueue.Tenant) {
		t.Errorf("Expecting: %+v, received: %+v", testStatsQ[2].Tenant, msq[0].statQueue.Tenant)
	} else if !reflect.DeepEqual(testStatsQ[2].ID, msq[0].statQueue.ID) {
		t.Errorf("Expecting: %+v, received: %+v", testStatsQ[2].ID, msq[0].statQueue.ID)
	} else if !reflect.DeepEqual(testStatsPrfs[2], msq[0].profile) {
		t.Errorf("Expecting: %+v, received: %+v", testStatsPrfs[2], msq[0].profile)
	}
}

func TestStatQueuesProcessEvent(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dmSTS := engine.NewDataManager(dbCM, cfg, nil)
	cacheS := engine.NewCacheS(cfg, nil, nil, nil)
	dmSTS.SetCache(cacheS)

	cfg.StatSCfg().StoreInterval = 1
	cfg.StatSCfg().StringIndexedFields = nil
	cfg.StatSCfg().PrefixIndexedFields = nil
	statService := NewStatService(cfg, dmSTS, cacheS,
		engine.NewFilterS(cfg, nil, dmSTS), nil)

	prepareStatsData(t, dmSTS)

	stq := map[string]string{}
	reply := []string{}
	expected := []string{"StatQueueProfile1"}
	err := statService.V1ProcessEvent(context.TODO(), testStatsArgs[0], &reply)
	if err != nil {
		t.Errorf("Error: %+v", err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expecting: %+v, received: %+v", expected, reply)
	}
	err = statService.V1GetQueueStringMetrics(context.TODO(), &utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: testStatsQ[0].Tenant, ID: testStatsQ[0].ID}}, &stq)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}

	expected = []string{"StatQueueProfile2"}
	err = statService.V1ProcessEvent(context.TODO(), testStatsArgs[1], &reply)
	if err != nil {
		t.Errorf("Error: %+v", err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expecting: %+v, received: %+v", expected, reply)
	}
	err = statService.V1GetQueueStringMetrics(context.TODO(), &utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: testStatsQ[1].Tenant, ID: testStatsQ[1].ID}}, &stq)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}

	expected = []string{"StatQueueProfilePrefix"}
	err = statService.V1ProcessEvent(context.TODO(), testStatsArgs[2], &reply)
	if err != nil {
		t.Errorf("Error: %+v", err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expecting: %+v, received: %+v", expected, reply)
	}
	err = statService.V1GetQueueStringMetrics(context.TODO(), &utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: testStatsQ[2].Tenant, ID: testStatsQ[2].ID}}, &stq)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
}

func TestStatQueuesMatchWithIndexFalse(t *testing.T) {

	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dmSTS := engine.NewDataManager(dbCM, cfg, nil)
	cacheS := engine.NewCacheS(cfg, nil, nil, nil)
	dmSTS.SetCache(cacheS)

	cfg.StatSCfg().StoreInterval = 1
	cfg.StatSCfg().StringIndexedFields = nil
	cfg.StatSCfg().PrefixIndexedFields = nil
	statService := NewStatService(cfg, dmSTS, cacheS,
		engine.NewFilterS(cfg, nil, dmSTS), nil)
	prepareStatsData(t, dmSTS)

	statService.cfg.StatSCfg().IndexedSelects = false
	msq, unlock, err := statService.matchingStatQueuesForEvent(context.TODO(), testStatsArgs[0].Tenant, testStatsArgs[0])
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	unlock()
	if !reflect.DeepEqual(testStatsQ[0].Tenant, msq[0].statQueue.Tenant) {
		t.Errorf("Expecting: %+v, received: %+v", testStatsQ[0].Tenant, msq[0].statQueue.Tenant)
	} else if !reflect.DeepEqual(testStatsQ[0].ID, msq[0].statQueue.ID) {
		t.Errorf("Expecting: %+v, received: %+v", testStatsQ[0].ID, msq[0].statQueue.ID)
	} else if !reflect.DeepEqual(testStatsPrfs[0], msq[0].profile) {
		t.Errorf("Expecting: %+v, received: %+v", testStatsPrfs[0], msq[0].profile)
	}
	msq, unlock, err = statService.matchingStatQueuesForEvent(context.TODO(), testStatsArgs[1].Tenant, testStatsArgs[1])
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	unlock()
	if !reflect.DeepEqual(testStatsQ[1].Tenant, msq[0].statQueue.Tenant) {
		t.Errorf("Expecting: %+v, received: %+v", testStatsQ[1].Tenant, msq[0].statQueue.Tenant)
	} else if !reflect.DeepEqual(testStatsQ[1].ID, msq[0].statQueue.ID) {
		t.Errorf("Expecting: %+v, received: %+v", testStatsQ[1].ID, msq[0].statQueue.ID)
	} else if !reflect.DeepEqual(testStatsPrfs[1], msq[0].profile) {
		t.Errorf("Expecting: %+v, received: %+v", testStatsPrfs[1], msq[0].profile)
	}
	msq, unlock, err = statService.matchingStatQueuesForEvent(context.TODO(), testStatsArgs[2].Tenant, testStatsArgs[2])
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	unlock()
	if !reflect.DeepEqual(testStatsQ[2].Tenant, msq[0].statQueue.Tenant) {
		t.Errorf("Expecting: %+v, received: %+v", testStatsQ[2].Tenant, msq[0].statQueue.Tenant)
	} else if !reflect.DeepEqual(testStatsQ[2].ID, msq[0].statQueue.ID) {
		t.Errorf("Expecting: %+v, received: %+v", testStatsQ[2].ID, msq[0].statQueue.ID)
	} else if !reflect.DeepEqual(testStatsPrfs[2], msq[0].profile) {
		t.Errorf("Expecting: %+v, received: %+v", testStatsPrfs[2], msq[0].profile)
	}
}

func TestStatQueuesV1ProcessEvent(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dmSTS := engine.NewDataManager(dbCM, cfg, nil)
	cacheS := engine.NewCacheS(cfg, nil, nil, nil)
	dmSTS.SetCache(cacheS)

	cfg.StatSCfg().StoreInterval = 1
	cfg.StatSCfg().StringIndexedFields = nil
	cfg.StatSCfg().PrefixIndexedFields = nil
	statService := NewStatService(cfg, dmSTS, cacheS,
		engine.NewFilterS(cfg, nil, dmSTS), nil)
	prepareStatsData(t, dmSTS)

	sqPrf := &utils.StatQueueProfile{
		Tenant:      "cgrates.org",
		ID:          "StatQueueProfile3",
		FilterIDs:   []string{"FLTR_STATS_1", "*ai:~*req.AnswerTime:2014-07-14T14:25:00Z"},
		QueueLength: 10,
		TTL:         10 * time.Second,
		Metrics: []*utils.MetricWithFilters{
			{
				MetricID: "*sum#~*req.Usage",
			},
		},
		ThresholdIDs: []string{},
		Stored:       true,
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
		MinItems: 1,
	}
	statSum, err := utils.NewStatMetric("*sum#~*req.Usage", uint64(sqPrf.MinItems), []string{})
	if err != nil {
		t.Fatal(err)
	}
	sq := &utils.StatQueue{Tenant: "cgrates.org", ID: "StatQueueProfile3", SQMetrics: map[string]utils.StatMetric{"*sum#~*req.Usage": statSum}}
	if err := dmSTS.SetStatQueueProfile(context.TODO(), sqPrf, true); err != nil {
		t.Error(err)
	}
	if err := dmSTS.SetStatQueue(context.TODO(), sq); err != nil {
		t.Error(err)
	}
	if tempStat, err := dmSTS.GetStatQueueProfile(context.TODO(), sqPrf.Tenant,
		sqPrf.ID, true, false, utils.NonTransactional); err != nil {
		t.Errorf("Error: %+v", err)
	} else if !reflect.DeepEqual(sqPrf, tempStat) {
		t.Errorf("Expecting: %+v, received: %+v", sqPrf, tempStat)
	}
	ev := testStatsArgs[0].Clone()
	ev.APIOpts = make(map[string]any)
	ev.APIOpts[utils.OptsStatsProfileIDs] = []string{"StatQueueProfile1", "StatQueueProfile2", "StatQueueProfile3"}
	reply := []string{}
	expected := []string{"StatQueueProfile1", "StatQueueProfile3"}
	expectedRev := []string{"StatQueueProfile3", "StatQueueProfile1"}
	if err := statService.V1ProcessEvent(context.TODO(), ev, &reply); err != nil {
		t.Errorf("Error: %+v", err)
	} else if !reflect.DeepEqual(reply, expected) && !reflect.DeepEqual(reply, expectedRev) {
		t.Errorf("Expecting: %+v, received: %+v", expected, reply)
	}
}

func TestStatQueuesUpdateStatQueue(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	idb, err := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	if err != nil {
		t.Error(err)
	}
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: idb}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	cacheS := engine.NewCacheS(cfg, nil, nil, nil)
	dm.SetCache(cacheS)
	sqp := &utils.StatQueueProfile{
		Tenant:      "cgrates.org",
		ID:          "THUP1",
		Stored:      true,
		QueueLength: 1,
		Metrics:     []*utils.MetricWithFilters{{MetricID: utils.MetaTCC}},
	}
	sqm := utils.NewTCC(0, utils.EmptyString, nil)
	if err := sqm.AddEvent("ev1", utils.MapStorage{utils.MetaOpts: utils.MapStorage{utils.MetaCost: 10}}); err != nil {
		t.Fatal(err)
	}
	sq := &utils.StatQueue{
		Tenant:    sqp.Tenant,
		ID:        sqp.ID,
		SQItems:   []utils.SQItem{{EventID: "ev1"}},
		SQMetrics: map[string]utils.StatMetric{utils.MetaTCC: sqm, utils.MetaTCD: sqm},
	}
	sqm2 := utils.NewTCC(0, utils.EmptyString, nil)

	expTh := &utils.StatQueue{
		Tenant:    sqp.Tenant,
		ID:        sqp.ID,
		SQMetrics: map[string]utils.StatMetric{utils.MetaTCC: sqm2},
	}

	if err := dm.SetStatQueueProfile(context.Background(), sqp, true); err != nil {
		t.Fatal(err)
	}

	if th, err := dm.GetStatQueue(context.Background(), sqp.Tenant, sqp.ID, false, false, utils.NonTransactional); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(expTh, th) {
		t.Errorf("Expected: %s, received: %s", utils.ToJSON(expTh), utils.ToJSON(th))
	}

	if err := dm.RemoveStatQueue(context.Background(), sqp.Tenant, sqp.ID); err != nil {
		t.Fatal(err)
	}
	if err := dm.SetStatQueueProfile(context.Background(), sqp, true); err != nil {
		t.Fatal(err)
	}

	if th, err := dm.GetStatQueue(context.Background(), sqp.Tenant, sqp.ID, false, false, utils.NonTransactional); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(expTh, th) {
		t.Errorf("Expected: %s, received: %s", utils.ToJSON(expTh), utils.ToJSON(th))
	}
	if err := dm.SetStatQueue(context.Background(), sq); err != nil {
		t.Fatal(err)
	}

	sqp = &utils.StatQueueProfile{
		Tenant:      "cgrates.org",
		ID:          "THUP1",
		Stored:      true,
		QueueLength: 1,
		Metrics:     []*utils.MetricWithFilters{{MetricID: utils.MetaTCC, FilterIDs: []string{"*string:~*req.Account:1001"}}},
	}

	if err := dm.SetStatQueueProfile(context.Background(), sqp, true); err != nil {
		t.Fatal(err)
	}

	sqm3 := utils.NewTCC(0, utils.EmptyString, []string{"*string:~*req.Account:1001"})

	expTh = &utils.StatQueue{
		Tenant:    sqp.Tenant,
		ID:        sqp.ID,
		SQItems:   []utils.SQItem{{EventID: "ev1"}},
		SQMetrics: map[string]utils.StatMetric{utils.MetaTCC: sqm3},
	}
	if th, err := dm.GetStatQueue(context.Background(), sqp.Tenant, sqp.ID, false, false, utils.NonTransactional); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(expTh, th) {
		t.Errorf("Expected: %s, received: %s", utils.ToJSON(expTh), utils.ToJSON(th))
	}

	if err := dm.SetStatQueue(context.Background(), sq); err != nil {
		t.Fatal(err)
	}

	sqm2 = utils.NewTCC(5, utils.EmptyString, nil)

	expTh = &utils.StatQueue{
		Tenant:    sqp.Tenant,
		ID:        sqp.ID,
		SQMetrics: map[string]utils.StatMetric{utils.MetaTCC: sqm2},
	}
	delete(sq.SQMetrics, utils.MetaTCD)
	if err := dm.SetStatQueue(context.Background(), sq); err != nil {
		t.Fatal(err)
	}

	sqp = &utils.StatQueueProfile{
		Tenant:      "cgrates.org",
		ID:          "THUP1",
		Stored:      true,
		QueueLength: 1,
		Metrics:     []*utils.MetricWithFilters{{MetricID: utils.MetaTCC}},
		MinItems:    5,
	}

	if err := dm.SetStatQueueProfile(context.Background(), sqp, true); err != nil {
		t.Fatal(err)
	}

	if th, err := dm.GetStatQueue(context.Background(), sqp.Tenant, sqp.ID, false, false, utils.NonTransactional); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(expTh, th) {
		t.Errorf("Expected: %s, received: %s", utils.ToJSON(expTh), utils.ToJSON(th))
	}
	if err := dm.SetStatQueue(context.Background(), sq); err != nil {
		t.Fatal(err)
	}

	sqp = &utils.StatQueueProfile{
		Tenant:      "cgrates.org",
		ID:          "THUP1",
		Stored:      true,
		QueueLength: 1,
		Metrics:     []*utils.MetricWithFilters{{MetricID: utils.MetaTCC}},
		MinItems:    5,
		TTL:         10,
	}

	if err := dm.SetStatQueueProfile(context.Background(), sqp, true); err != nil {
		t.Fatal(err)
	}

	if th, err := dm.GetStatQueue(context.Background(), sqp.Tenant, sqp.ID, false, false, utils.NonTransactional); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(expTh, th) {
		t.Errorf("Expected: %s, received: %s", utils.ToJSON(expTh), utils.ToJSON(th))
	}

	if err := dm.SetStatQueue(context.Background(), sq); err != nil {
		t.Fatal(err)
	}

	sqp = &utils.StatQueueProfile{
		Tenant:      "cgrates.org",
		ID:          "THUP1",
		Stored:      true,
		QueueLength: 10,
		Metrics:     []*utils.MetricWithFilters{{MetricID: utils.MetaTCC}},
		TTL:         10,
		MinItems:    5,
	}

	if err := dm.SetStatQueueProfile(context.Background(), sqp, true); err != nil {
		t.Fatal(err)
	}

	if th, err := dm.GetStatQueue(context.Background(), sqp.Tenant, sqp.ID, false, false, utils.NonTransactional); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(expTh, th) {
		t.Errorf("Expected: %s, received: %s", utils.ToJSON(expTh), utils.ToJSON(th))
	}

	if err := dm.SetStatQueue(context.Background(), sq); err != nil {
		t.Fatal(err)
	}

	sqp = &utils.StatQueueProfile{
		Tenant:      "cgrates.org",
		ID:          "THUP1",
		Stored:      false,
		QueueLength: 10,
		Metrics:     []*utils.MetricWithFilters{{MetricID: utils.MetaTCC}},
		TTL:         10,
		MinItems:    5,
	}

	if err := dm.SetStatQueueProfile(context.Background(), sqp, true); err != nil {
		t.Fatal(err)
	}

	if th, err := dm.GetStatQueue(context.Background(), sqp.Tenant, sqp.ID, false, false, utils.NonTransactional); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(expTh, th) {
		t.Errorf("Expected: %s, received: %s", utils.ToJSON(expTh), utils.ToJSON(th))
	}

	if err := dm.RemoveStatQueueProfile(context.Background(), sqp.Tenant, sqp.ID, true); err != nil {
		t.Fatal(err)
	}
	if _, err := dm.GetStatQueue(context.Background(), sqp.Tenant, sqp.ID, false, false, utils.NonTransactional); err != utils.ErrNotFound {
		t.Fatal(err)
	}
}

func TestStatQueueMatchingStatQueuesForEventLocks(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()

	defer func() {
		guardian.Guardian = guardian.New()
	}()
	cacheS := engine.NewCacheS(cfg, nil, nil, nil)
	db, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: db}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	dm.SetCache(cacheS)
	cfg.StatSCfg().StoreInterval = 1
	cfg.StatSCfg().StringIndexedFields = nil
	cfg.StatSCfg().PrefixIndexedFields = nil
	sS := NewStatService(cfg, dm, cacheS,
		engine.NewFilterS(cfg, nil, dm), nil)

	ids := utils.StringSet{}
	for i := 0; i < 10; i++ {
		sqPrf := &utils.StatQueueProfile{
			Tenant: "cgrates.org",
			ID:     fmt.Sprintf("STS%d", i),
			Weights: utils.DynamicWeights{
				{
					Weight: 20.00,
				},
			},
			ThresholdIDs: []string{utils.MetaNone},
			QueueLength:  1,
			Stored:       true,
		}
		dm.SetStatQueueProfile(context.Background(), sqPrf, true)
		ids.Add(sqPrf.ID)
	}
	dm.RemoveStatQueue(context.Background(), "cgrates.org", "STS1")
	ev := &utils.CGREvent{
		APIOpts: map[string]any{
			utils.OptsStatsProfileIDs: ids.AsSlice(),
		},
	}
	_, _, err := sS.matchingStatQueuesForEvent(context.Background(), "cgrates.org", ev)
	if err != utils.ErrNotFound {
		t.Errorf("Error: %+v", err)
	}
}

func TestStatQueueMatchingStatQueuesForEventLocks2(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cacheS := engine.NewCacheS(cfg, nil, nil, nil)
	db, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: db}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	dm.SetCache(cacheS)
	cfg.StatSCfg().StoreInterval = 1
	cfg.StatSCfg().StringIndexedFields = nil
	cfg.StatSCfg().PrefixIndexedFields = nil
	sS := NewStatService(cfg, dm, cacheS,
		engine.NewFilterS(cfg, nil, dm), nil)

	ids := utils.StringSet{}
	for i := 0; i < 10; i++ {
		sqPrf := &utils.StatQueueProfile{
			Tenant:      "cgrates.org",
			ID:          fmt.Sprintf("STS%d", i),
			QueueLength: 1,
			Stored:      true,
			Weights: utils.DynamicWeights{
				{
					Weight: 20.00,
				},
			},
			ThresholdIDs: []string{utils.MetaNone},
		}
		dm.SetStatQueueProfile(context.Background(), sqPrf, true)
		ids.Add(sqPrf.ID)
	}
	sqPrf := &utils.StatQueueProfile{
		Tenant:      "cgrates.org",
		ID:          "STS20",
		FilterIDs:   []string{"FLTR_STATS_201"},
		QueueLength: 1,
		Stored:      true,
		Weights: utils.DynamicWeights{
			{
				Weight: 20.00,
			},
		},
		ThresholdIDs: []string{utils.MetaNone},
	}
	err := db.SetStatQueueProfileDrv(context.Background(), sqPrf)
	if err != nil {
		t.Fatal(err)
	}
	ids.Add(sqPrf.ID)
	ev := &utils.CGREvent{
		APIOpts: map[string]any{
			utils.OptsStatsProfileIDs: ids.AsSlice(),
		},
	}
	_, _, err = sS.matchingStatQueuesForEvent(context.Background(), "cgrates.org", ev)
	expErr := utils.ErrPrefixNotFound(sqPrf.FilterIDs[0])
	if err == nil || err.Error() != expErr.Error() {
		t.Errorf("Expected error: %s ,received: %+v", expErr, err)
	}
}

func TestStatQueueMatchingStatQueuesForEventLocks3(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cacheS := engine.NewCacheS(cfg, nil, nil, nil)
	db := &engine.DataDBMock{
		GetStatQueueProfileDrvF: func(ctx *context.Context, tnt, id string) (*utils.StatQueueProfile, error) {
			if id == "STS1" {
				return nil, utils.ErrNotImplemented
			}
			sqPrf := &utils.StatQueueProfile{
				Tenant:      "cgrates.org",
				ID:          id,
				QueueLength: 1,
				Stored:      true,
				Weights: utils.DynamicWeights{
					{
						Weight: 20.00,
					},
				},
				ThresholdIDs: []string{utils.MetaNone},
			}
			cacheS.Set(ctx, utils.CacheStatQueues, sqPrf.TenantID(), &utils.StatQueue{
				Tenant:    sqPrf.Tenant,
				ID:        sqPrf.ID,
				SQMetrics: make(map[string]utils.StatMetric),
			}, nil, true, utils.NonTransactional)
			return sqPrf, nil
		},
	}
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: db}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	dm.SetCache(cacheS)
	cfg.StatSCfg().StoreInterval = 1
	cfg.StatSCfg().StringIndexedFields = nil
	cfg.StatSCfg().PrefixIndexedFields = nil
	sS := NewStatService(cfg, dm, cacheS,
		engine.NewFilterS(cfg, nil, dm), nil)

	ids := utils.StringSet{}
	for i := 0; i < 10; i++ {
		ids.Add(fmt.Sprintf("STS%d", i))
	}
	ev := &utils.CGREvent{
		APIOpts: map[string]any{
			utils.OptsStatsProfileIDs: ids.AsSlice(),
		},
	}
	_, _, err := sS.matchingStatQueuesForEvent(context.Background(), "cgrates.org", ev)
	if err != utils.ErrNotImplemented {
		t.Fatalf("Error: %+v", err)
	}
}

func TestStatQueueReload(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.StatSCfg().StoreInterval = 5 * time.Millisecond
	cacheS := engine.NewCacheS(cfg, nil, nil, nil)
	synctest.Test(t, func(*testing.T) {
		sS := NewStatService(cfg, nil, cacheS, nil, nil)
		sS.StartLoop(context.Background())
		sS.Reload(context.Background())
		sS.Shutdown(context.Background())
		sS.Shutdown(context.Background())
		sS.Reload(context.Background())
	})
}

func TestStatQueueReloadShutdownConcurrent(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.StatSCfg().StoreInterval = 5 * time.Millisecond
	cacheS := engine.NewCacheS(cfg, nil, nil, nil)
	synctest.Test(t, func(*testing.T) {
		sS := NewStatService(cfg, nil, cacheS, nil, nil)
		sS.StartLoop(context.Background())
		var wg sync.WaitGroup
		wg.Go(func() { sS.Reload(context.Background()) })
		wg.Go(func() { sS.Shutdown(context.Background()) })
		wg.Wait()
	})
}

func TestStatQueueStartLoop(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cacheS := engine.NewCacheS(cfg, nil, nil, nil)
	synctest.Test(t, func(*testing.T) {
		sS := NewStatService(cfg, nil, cacheS, nil, nil)
		sS.StartLoop(context.Background())
		sS.backupLoop.Wait()
	})
}

func TestStatQueueStoreStatsOK(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	cacheS := engine.NewCacheS(cfg, nil, nil, nil)
	dm.SetCache(cacheS)
	sS := NewStatService(cfg, dm, cacheS, nil, nil)

	exp := &utils.StatQueue{
		Tenant:    "cgrates.org",
		ID:        "SQ1",
		SQMetrics: make(map[string]utils.StatMetric),
	}
	cacheS.SetWithoutReplicate(utils.CacheStatQueues, "cgrates.org:SQ1", exp, nil, true,
		utils.NonTransactional)
	sS.storedStatQueues.Add("cgrates.org:SQ1")
	sS.storeStats(context.Background())

	if rcv, err := sS.dm.GetStatQueue(context.Background(), "cgrates.org", "SQ1", true, false,
		utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>",
			utils.ToJSON(exp), utils.ToJSON(rcv))
	}

	cacheS.Remove(context.Background(), utils.CacheStatQueues, "cgrates.org:SQ1", true, utils.NonTransactional)
}

func TestStatQueueStoreStatsStoreSQErr(t *testing.T) {
	tmpLogger := utils.Logger
	defer func() {
		utils.Logger = tmpLogger
	}()

	var buf bytes.Buffer
	utils.Logger = utils.NewStdLoggerWithWriter(&buf, "", 4)

	cfg := config.NewDefaultCGRConfig()
	cacheS := engine.NewCacheS(cfg, nil, nil, nil)
	sS := NewStatService(cfg, nil, cacheS, nil, nil)

	value := &utils.StatQueue{
		Tenant:    "cgrates.org",
		ID:        "SQ1",
		SQMetrics: make(map[string]utils.StatMetric),
	}

	cacheS.SetWithoutReplicate(utils.CacheStatQueues, "SQ1", value, nil, true,
		utils.NonTransactional)
	sS.storedStatQueues.Add("SQ1")
	exp := utils.StringSet{
		"SQ1": struct{}{},
	}
	expLog := `[WARNING] <StatS> failed saving StatQueue with ID: cgrates.org:SQ1, error: NO_DATABASE_CONNECTION`
	sS.storeStats(context.Background())

	if !reflect.DeepEqual(sS.storedStatQueues, exp) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", exp, sS.storedStatQueues)
	}
	if rcvLog := buf.String(); !strings.Contains(rcvLog, expLog) {
		t.Errorf("expected log <%+v>\n to be in included in: <%+v>", expLog, rcvLog)
	}

	cacheS.Remove(context.Background(), utils.CacheStatQueues, "SQ1", true, utils.NonTransactional)
}

func TestStatQueueStoreStatsCacheGetErr(t *testing.T) {
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
	sS := NewStatService(cfg, dm, cacheS, nil, nil)

	value := &utils.StatQueue{
		Tenant:    "cgrates.org",
		ID:        "SQ1",
		SQMetrics: make(map[string]utils.StatMetric),
	}

	cacheS.SetWithoutReplicate(utils.CacheStatQueues, "SQ2", value, nil, true,
		utils.NonTransactional)
	sS.storedStatQueues.Add("SQ1")
	expLog := `[WARNING] <StatS> failed retrieving from cache stat queue with ID: SQ1`
	sS.storeStats(context.Background())

	if rcvLog := buf.String(); !strings.Contains(rcvLog, expLog) {
		t.Errorf("expected <%+v> \nto be included in: <%+v>", expLog, rcvLog)
	}

	cacheS.Remove(context.Background(), utils.CacheStatQueues, "SQ2", true, utils.NonTransactional)
}

func TestStatQueueStoreStatQueueCacheSetErr(t *testing.T) {
	tmpLogger := utils.Logger
	defer func() {
		utils.Logger = tmpLogger
	}()
	var buf bytes.Buffer
	utils.Logger = utils.NewStdLoggerWithWriter(&buf, "", 4)

	cfg := config.NewDefaultCGRConfig()
	cfg.CacheCfg().ReplicationConns = []string{"test"}
	cfg.CacheCfg().Partitions[utils.CacheStatQueues].Replicate = true
	cfg.RPCConns()["test"] = &config.RPCConn{Conns: []*config.RemoteHost{{}}}
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	cM := engine.NewConnManager(cfg)
	dm := engine.NewDataManager(dbCM, cfg, cM)
	cacheS := engine.NewCacheS(cfg, dm, cM, nil)
	cM.SetCache(cacheS)
	dm.SetCache(cacheS)
	filterS := engine.NewFilterS(cfg, nil, dm)
	sS := NewStatService(cfg, dm, cacheS, filterS, cM)

	sq := &utils.StatQueue{
		Tenant:    "cgrates.org",
		ID:        "SQ1",
		SQMetrics: make(map[string]utils.StatMetric),
	}
	cacheS.SetWithoutReplicate(utils.CacheStatQueues, sq.TenantID(), sq, nil, true, utils.NonTransactional)
	expLog := `[WARNING] <StatS> failed caching StatQueue with ID: cgrates.org:SQ1, error: DISCONNECTED`
	if err := sS.storeStatQueue(context.Background(), sq); err == nil ||
		err.Error() != utils.ErrDisconnected.Error() {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrDisconnected, err)
	} else if rcv := buf.String(); !strings.Contains(rcv, expLog) {
		t.Errorf("expected log <%+v> to be included in <%+v>", expLog, rcv)
	}
}

func TestStatQueueStoreStatQueueOK(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	cacheS := engine.NewCacheS(cfg, nil, nil, nil)
	dm.SetCache(cacheS)
	sS := NewStatService(cfg, dm, cacheS, nil, nil)

	sq := &utils.StatQueue{
		Tenant:    "cgrates.org",
		ID:        "SQ1",
		SQMetrics: make(map[string]utils.StatMetric),
	}

	if err := sS.storeStatQueue(context.Background(), sq); err != nil {
		t.Error(err)
	}
}

func TestStatQueueProcessEventOK(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	cacheS := engine.NewCacheS(cfg, nil, nil, nil)
	dm.SetCache(cacheS)
	filterS := engine.NewFilterS(cfg, nil, dm)
	sS := NewStatService(cfg, dm, cacheS, filterS, nil)

	sqPrf := &utils.StatQueueProfile{
		Tenant:    "cgrates.org",
		ID:        "SQ1",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
		Blockers:     utils.DynamicBlockers{{Blocker: true}},
		QueueLength:  10,
		ThresholdIDs: []string{"*none"},
		MinItems:     5,
		Metrics: []*utils.MetricWithFilters{
			{
				MetricID: utils.MetaTCD,
			},
		},
	}
	stat, err := utils.NewStatMetric("*tcd", uint64(sqPrf.MinItems), []string{})
	if err != nil {
		t.Fatal(err)
	}
	sq := &utils.StatQueue{
		Tenant: "cgrates.org",
		ID:     "SQ1",
		SQItems: []utils.SQItem{
			{
				EventID: "SqProcessEvent",
			},
		},
		SQMetrics: map[string]utils.StatMetric{
			utils.MetaTCD: stat,
		},
	}

	if err := dm.SetStatQueueProfile(context.Background(), sqPrf, true); err != nil {
		t.Error(err)
	}
	if err := dm.SetStatQueue(context.Background(), sq); err != nil {
		t.Error(err)
	}

	args := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "SqProcessEvent",
		Event: map[string]any{
			utils.AccountField: "1001",
		},
		APIOpts: map[string]any{
			utils.MetaUsage:           "10s",
			utils.OptsStatsProfileIDs: []string{"SQ1"},
		},
	}

	expIDs := []string{"SQ1"}
	if rcvIDs, err := sS.processEvent(context.Background(), args.Tenant, args); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcvIDs, expIDs) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", expIDs, rcvIDs)
	}
}

func TestStatQueueProcessEventProcessThPartExec(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	cacheS := engine.NewCacheS(cfg, nil, nil, nil)
	dm.SetCache(cacheS)
	filterS := engine.NewFilterS(cfg, nil, dm)
	sS := NewStatService(cfg, dm, cacheS, filterS, nil)

	sqPrf := &utils.StatQueueProfile{
		Tenant:    "cgrates.org",
		ID:        "SQ1",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
		Blockers:     utils.DynamicBlockers{{Blocker: true}},
		QueueLength:  10,
		ThresholdIDs: []string{"*none"},
		MinItems:     5,
	}
	sq := &utils.StatQueue{
		Tenant: "cgrates.org",
		ID:     "SQ1",
		SQItems: []utils.SQItem{
			{
				EventID: "SqProcessEvent",
			},
		},
		SQMetrics: make(map[string]utils.StatMetric),
	}

	if err := dm.SetStatQueueProfile(context.Background(), sqPrf, true); err != nil {
		t.Error(err)
	}
	if err := dm.SetStatQueue(context.Background(), sq); err != nil {
		t.Error(err)
	}

	args := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "SqProcessEvent",
		Event: map[string]any{
			utils.AccountField: "1002",
		},
	}

	if _, err := sS.processEvent(context.Background(), args.Tenant, args); err == nil ||
		err != utils.ErrNotFound {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}
}

func TestStatQueueV1ProcessEventMissingArgs(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	cacheS := engine.NewCacheS(cfg, dm, nil, nil)
	dm.SetCache(cacheS)
	filterS := engine.NewFilterS(cfg, nil, dm)
	sS := NewStatService(cfg, dm, cacheS, filterS, nil)

	sqPrf := &utils.StatQueueProfile{
		Tenant:    "cgrates.org",
		ID:        "SQ1",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
		Blockers:     utils.DynamicBlockers{{Blocker: true}},
		QueueLength:  10,
		ThresholdIDs: []string{"*none"},
		MinItems:     5,
		Metrics: []*utils.MetricWithFilters{
			{
				MetricID: utils.MetaTCD,
			},
		},
	}
	sq := &utils.StatQueue{
		Tenant: "cgrates.org",
		ID:     "SQ1",
		SQItems: []utils.SQItem{
			{
				EventID: "SqProcessEvent",
			},
		},
		SQMetrics: map[string]utils.StatMetric{
			utils.MetaTCD: &utils.StatTCD{},
		},
	}

	if err := dm.SetStatQueueProfile(context.Background(), sqPrf, true); err != nil {
		t.Error(err)
	}
	if err := dm.SetStatQueue(context.Background(), sq); err != nil {
		t.Error(err)
	}

	var reply []string
	experr := `MANDATORY_IE_MISSING: [CGREvent]`
	if err := sS.V1ProcessEvent(context.Background(), nil, &reply); err == nil ||
		err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	args := &utils.CGREvent{
		Tenant: "cgrates.org",
		Event: map[string]any{
			utils.AccountField: "1001",
		},
		APIOpts: map[string]any{
			utils.OptsStatsProfileIDs: []string{"SQ1"},
		},
	}

	experr = `MANDATORY_IE_MISSING: [ID]`
	if err := sS.V1ProcessEvent(context.Background(), args, &reply); err == nil ||
		err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	args = &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "SqProcessEvent",
		Event:  nil,
		APIOpts: map[string]any{
			utils.OptsStatsProfileIDs: []string{"SQ1"},
		},
	}

	experr = `MANDATORY_IE_MISSING: [Event]`
	if err := sS.V1ProcessEvent(context.Background(), args, &reply); err == nil ||
		err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}
}

func TestStatQueueV1GetQueueIDsOK(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	cacheS := engine.NewCacheS(cfg, dm, nil, nil)
	dm.SetCache(cacheS)
	filterS := engine.NewFilterS(cfg, nil, dm)
	sS := NewStatService(cfg, dm, cacheS, filterS, nil)

	sqPrf := &utils.StatQueueProfile{
		Tenant:    "cgrates.org",
		ID:        "SQ1",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
		Blockers:     utils.DynamicBlockers{{Blocker: true}},
		QueueLength:  10,
		ThresholdIDs: []string{"*none"},
		MinItems:     5,
		Metrics: []*utils.MetricWithFilters{
			{
				MetricID: utils.MetaTCD,
			},
		},
	}
	sq1 := &utils.StatQueue{
		Tenant: "cgrates.org",
		ID:     "SQ1",
		SQItems: []utils.SQItem{
			{
				EventID: "SqProcessEvent",
			},
		},
		SQMetrics: map[string]utils.StatMetric{
			utils.MetaTCD: &utils.StatTCD{},
		},
	}

	sq2 := &utils.StatQueue{
		Tenant:    "testTenant",
		ID:        "SQ2",
		SQMetrics: make(map[string]utils.StatMetric),
	}
	sq3 := &utils.StatQueue{
		Tenant:    "cgrates.org",
		ID:        "SQ3",
		SQMetrics: make(map[string]utils.StatMetric),
	}

	if err := dm.SetStatQueueProfile(context.Background(), sqPrf, true); err != nil {
		t.Error(err)
	}
	if err := dm.SetStatQueue(context.Background(), sq1); err != nil {
		t.Error(err)
	}
	if err := dm.SetStatQueue(context.Background(), sq2); err != nil {
		t.Error(err)
	}
	if err := dm.SetStatQueue(context.Background(), sq3); err != nil {
		t.Error(err)
	}

	expIDs := []string{"SQ1", "SQ3"}
	var qIDs []string
	if err := sS.V1GetQueueIDs(context.Background(), &utils.TenantWithAPIOpts{}, &qIDs); err != nil {
		t.Error(err)
	} else {
		sort.Strings(qIDs)
		if !reflect.DeepEqual(qIDs, expIDs) {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>", expIDs, qIDs)
		}
	}
}

func TestStatQueueV1GetQueueIDsGetKeysForPrefixErr(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data := &engine.DataDBMock{}
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	cacheS := engine.NewCacheS(cfg, dm, nil, nil)
	dm.SetCache(cacheS)
	filterS := engine.NewFilterS(cfg, nil, dm)
	sS := NewStatService(cfg, dm, cacheS, filterS, nil)

	var qIDs []string
	if err := sS.V1GetQueueIDs(context.Background(), &utils.TenantWithAPIOpts{}, &qIDs); err == nil ||
		err.Error() != utils.ErrNotImplemented.Error() {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotImplemented, err)
	}
}

func TestStatQueueV1GetStatQueueOK(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	cacheS := engine.NewCacheS(cfg, dm, nil, nil)
	dm.SetCache(cacheS)
	filterS := engine.NewFilterS(cfg, nil, dm)
	sS := NewStatService(cfg, dm, cacheS, filterS, nil)

	sqPrf := &utils.StatQueueProfile{
		Tenant:    "cgrates.org",
		ID:        "SQ1",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
		Blockers:     utils.DynamicBlockers{{Blocker: true}},
		QueueLength:  10,
		ThresholdIDs: []string{"*none"},
		MinItems:     5,
		Metrics: []*utils.MetricWithFilters{
			{
				MetricID: utils.MetaTCD,
			},
		},
	}
	sq := &utils.StatQueue{
		Tenant: "cgrates.org",
		ID:     "SQ1",
		SQItems: []utils.SQItem{
			{
				EventID: "SqProcessEvent",
			},
		},
		SQMetrics: map[string]utils.StatMetric{
			utils.MetaTCD: &utils.StatTCD{},
		},
	}

	if err := dm.SetStatQueueProfile(context.Background(), sqPrf, true); err != nil {
		t.Error(err)
	}
	if err := dm.SetStatQueue(context.Background(), sq); err != nil {
		t.Error(err)
	}

	var reply utils.StatQueue
	if err := sS.V1GetStatQueue(context.Background(), &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			ID: "SQ1",
		},
	}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, *sq) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>",
			utils.ToJSON(*sq), utils.ToJSON(reply))
	}
}

func TestStatQueueV1GetStatQueueNotFound(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	cacheS := engine.NewCacheS(cfg, dm, nil, nil)
	dm.SetCache(cacheS)
	filterS := engine.NewFilterS(cfg, nil, dm)
	sS := NewStatService(cfg, dm, cacheS, filterS, nil)

	var reply utils.StatQueue
	if err := sS.V1GetStatQueue(context.Background(), &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			ID: "SQ1",
		},
	}, &reply); err == nil || err != utils.ErrNotFound {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}
}

func TestStatQueueV1GetStatQueueMissingArgs(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	cacheS := engine.NewCacheS(cfg, dm, nil, nil)
	dm.SetCache(cacheS)
	filterS := engine.NewFilterS(cfg, nil, dm)
	sS := NewStatService(cfg, dm, cacheS, filterS, nil)

	sqPrf := &utils.StatQueueProfile{
		Tenant:    "cgrates.org",
		ID:        "SQ1",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
		Blockers:     utils.DynamicBlockers{{Blocker: true}},
		QueueLength:  10,
		ThresholdIDs: []string{"*none"},
		MinItems:     5,
		Metrics: []*utils.MetricWithFilters{
			{
				MetricID: utils.MetaTCD,
			},
		},
	}
	sq := &utils.StatQueue{
		Tenant: "cgrates.org",
		ID:     "SQ1",
		SQItems: []utils.SQItem{
			{
				EventID: "SqProcessEvent",
			},
		},
		SQMetrics: map[string]utils.StatMetric{
			utils.MetaTCD: &utils.StatTCD{},
		},
	}

	if err := dm.SetStatQueueProfile(context.Background(), sqPrf, true); err != nil {
		t.Error(err)
	}
	if err := dm.SetStatQueue(context.Background(), sq); err != nil {
		t.Error(err)
	}

	experr := `MANDATORY_IE_MISSING: [ID]`
	var reply utils.StatQueue
	if err := sS.V1GetStatQueue(context.Background(), &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{},
	}, &reply); err == nil || err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}
}

func TestStatQueueV1GetStatQueuesForEventOK(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	cacheS := engine.NewCacheS(cfg, dm, nil, nil)
	dm.SetCache(cacheS)
	filterS := engine.NewFilterS(cfg, nil, dm)
	sS := NewStatService(cfg, dm, cacheS, filterS, nil)

	sqPrf1 := &utils.StatQueueProfile{
		Tenant:    "cgrates.org",
		ID:        "SQ1",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
		Blockers:     utils.DynamicBlockers{{Blocker: true}},
		QueueLength:  10,
		ThresholdIDs: []string{"*none"},
		MinItems:     5,
		Metrics: []*utils.MetricWithFilters{
			{
				MetricID: utils.MetaTCD,
			},
		},
	}

	if err := dm.SetStatQueueProfile(context.Background(), sqPrf1, true); err != nil {
		t.Error(err)
	}

	sqPrf2 := &utils.StatQueueProfile{
		Tenant: "cgrates.org",
		ID:     "SQ2",
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
		Blockers:     utils.DynamicBlockers{{Blocker: false}},
		QueueLength:  10,
		ThresholdIDs: []string{"*none"},
		MinItems:     5,
		Metrics: []*utils.MetricWithFilters{
			{
				MetricID: utils.MetaACD,
			},
		},
	}

	if err := dm.SetStatQueueProfile(context.Background(), sqPrf2, true); err != nil {
		t.Error(err)
	}

	args := &utils.CGREvent{
		ID: "TestGetStatQueuesForEvent",
		Event: map[string]any{
			utils.AccountField: "1001",
		},
	}

	exp := []string{"SQ1", "SQ2"}
	var reply []string
	if err := sS.V1GetStatQueuesForEvent(context.Background(), args, &reply); err != nil {
		t.Error(err)
	} else {
		sort.Strings(reply)
		if !reflect.DeepEqual(reply, exp) {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>", exp, reply)
		}
	}
}

func TestStatQueueV1GetStatQueuesForEventNotFoundErr(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	cacheS := engine.NewCacheS(cfg, dm, nil, nil)
	dm.SetCache(cacheS)
	filterS := engine.NewFilterS(cfg, nil, dm)
	sS := NewStatService(cfg, dm, cacheS, filterS, nil)

	sqPrf := &utils.StatQueueProfile{
		Tenant:    "cgrates.org",
		ID:        "SQ1",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
		Blockers:     utils.DynamicBlockers{{Blocker: true}},
		QueueLength:  10,
		ThresholdIDs: []string{"*none"},
		MinItems:     5,
		Metrics: []*utils.MetricWithFilters{
			{
				MetricID: utils.MetaTCD,
			},
		},
	}

	if err := dm.SetStatQueueProfile(context.Background(), sqPrf, true); err != nil {
		t.Error(err)
	}

	args := &utils.CGREvent{
		ID: "TestGetStatQueuesForEvent",
		Event: map[string]any{
			utils.AccountField: "1002",
		},
	}

	var reply []string
	if err := sS.V1GetStatQueuesForEvent(context.Background(), args, &reply); err == nil ||
		err != utils.ErrNotFound {
		t.Errorf("expected: <%+v>, received: <%+v>", utils.ErrNotFound, err)
	}
}

func TestStatQueueV1GetStatQueuesForEventMissingArgs(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	cacheS := engine.NewCacheS(cfg, dm, nil, nil)
	dm.SetCache(cacheS)
	filterS := engine.NewFilterS(cfg, nil, dm)
	sS := NewStatService(cfg, dm, cacheS, filterS, nil)

	sqPrf := &utils.StatQueueProfile{
		Tenant:    "cgrates.org",
		ID:        "SQ1",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
		Blockers:     utils.DynamicBlockers{{Blocker: true}},
		QueueLength:  10,
		ThresholdIDs: []string{"*none"},
		MinItems:     5,
		Metrics: []*utils.MetricWithFilters{
			{
				MetricID: utils.MetaTCD,
			},
		},
	}

	if err := dm.SetStatQueueProfile(context.Background(), sqPrf, true); err != nil {
		t.Error(err)
	}

	experr := `MANDATORY_IE_MISSING: [CGREvent]`
	var reply []string
	if err := sS.V1GetStatQueuesForEvent(context.Background(), nil, &reply); err == nil ||
		err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	args := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     utils.EmptyString,
		Event: map[string]any{
			utils.AccountField: "1001",
		},
	}

	experr = `MANDATORY_IE_MISSING: [ID]`
	if err := sS.V1GetStatQueuesForEvent(context.Background(), args, &reply); err == nil ||
		err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	args = &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "TestGetStatQueuesForEvent",
		Event:  nil,
	}

	experr = `MANDATORY_IE_MISSING: [Event]`
	if err := sS.V1GetStatQueuesForEvent(context.Background(), args, &reply); err == nil ||
		err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}
}

func TestStatQueueV1ResetStatQueueOK(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.StatSCfg().StoreInterval = 1
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	cacheS := engine.NewCacheS(cfg, dm, nil, nil)
	dm.SetCache(cacheS)
	filterS := engine.NewFilterS(cfg, nil, dm)
	sS := NewStatService(cfg, dm, cacheS, filterS, nil)

	sq := &utils.StatQueue{
		Tenant: "cgrates.org",
		ID:     "SQ1",
		SQItems: []utils.SQItem{
			{
				EventID: "SqProcessEvent",
			},
		},
		SQMetrics: map[string]utils.StatMetric{
			utils.MetaTCD: &utils.StatTCD{
				Metric: &utils.Metric{
					Value:  utils.NewDecimal(int64(time.Hour), 0),
					Count:  1,
					Events: make(map[string]*utils.DecimalWithCompress),
				},
			},
		},
	}

	if err := dm.SetStatQueue(context.Background(), sq); err != nil {
		t.Error(err)
	}
	cacheS.Clear(nil)
	expStored := utils.StringSet{
		"cgrates.org:SQ1": {},
	}
	expSq := &utils.StatQueue{
		Tenant:  "cgrates.org",
		ID:      "SQ1",
		SQItems: []utils.SQItem{},
		SQMetrics: map[string]utils.StatMetric{
			utils.MetaTCD: &utils.StatTCD{
				Metric: &utils.Metric{
					Value:  utils.NewDecimal(0, 0),
					Events: make(map[string]*utils.DecimalWithCompress),
				},
			},
		},
	}
	var reply string

	if err := sS.V1ResetStatQueue(context.Background(), &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			ID: "SQ1",
		},
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Unexpected reply returned: <%q>", reply)
	}
	if nSq, err := sS.dm.GetStatQueue(context.Background(), "cgrates.org", "SQ1", true, true, utils.NonTransactional); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(nSq, expSq) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ToJSON(expSq), utils.ToJSON(nSq))
	} else if !reflect.DeepEqual(sS.storedStatQueues, expStored) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", expStored, sS.storedStatQueues)
	}
}

func TestStatQueueV1ResetStatQueueNotFoundErr(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.StatSCfg().StoreInterval = 1
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	cacheS := engine.NewCacheS(cfg, dm, nil, nil)
	dm.SetCache(cacheS)
	filterS := engine.NewFilterS(cfg, nil, dm)
	sS := NewStatService(cfg, dm, cacheS, filterS, nil)

	sq := &utils.StatQueue{
		Tenant: "cgrates.org",
		ID:     "SQ1",
		SQItems: []utils.SQItem{
			{
				EventID: "SqProcessEvent",
			},
		},
		SQMetrics: map[string]utils.StatMetric{
			utils.MetaTCD: &utils.StatTCD{
				Metric: &utils.Metric{
					Value:  utils.NewDecimal(int64(time.Hour), 0),
					Count:  1,
					Events: make(map[string]*utils.DecimalWithCompress),
				},
			},
		},
	}

	if err := dm.SetStatQueue(context.Background(), sq); err != nil {
		t.Error(err)
	}

	var reply string
	if err := sS.V1ResetStatQueue(context.Background(), &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			ID: "SQ2",
		},
	}, &reply); err == nil || err != utils.ErrNotFound {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}
}

func TestStatQueueV1ResetStatQueueMissingArgs(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.StatSCfg().StoreInterval = 1
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	cacheS := engine.NewCacheS(cfg, dm, nil, nil)
	dm.SetCache(cacheS)
	filterS := engine.NewFilterS(cfg, nil, dm)
	sS := NewStatService(cfg, dm, cacheS, filterS, nil)

	sq := &utils.StatQueue{
		Tenant: "cgrates.org",
		ID:     "SQ1",
		SQItems: []utils.SQItem{
			{
				EventID: "SqProcessEvent",
			},
		},
		SQMetrics: map[string]utils.StatMetric{
			utils.MetaTCD: &utils.StatTCD{
				Metric: &utils.Metric{
					Value:  utils.NewDecimal(int64(time.Hour), 0),
					Count:  1,
					Events: make(map[string]*utils.DecimalWithCompress),
				},
			},
		},
	}

	if err := dm.SetStatQueue(context.Background(), sq); err != nil {
		t.Error(err)
	}

	experr := `MANDATORY_IE_MISSING: [ID]`
	var reply string
	if err := sS.V1ResetStatQueue(context.Background(), &utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{}}, &reply); err == nil ||
		err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}
}

func TestStatQueueV1ResetStatQueueUnsupportedMetricType(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.StatSCfg().StoreInterval = 1
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	cacheS := engine.NewCacheS(cfg, dm, nil, nil)
	dm.SetCache(cacheS)
	filterS := engine.NewFilterS(cfg, nil, dm)
	sS := NewStatService(cfg, dm, cacheS, filterS, nil)

	sq := &utils.StatQueue{
		Tenant: "cgrates.org",
		ID:     "SQ1",
		SQItems: []utils.SQItem{
			{
				EventID: "SqProcessEvent",
			},
		},
		SQMetrics: map[string]utils.StatMetric{
			"testMetricType": &utils.StatTCD{
				Metric: &utils.Metric{
					Value:  utils.NewDecimal(int64(time.Hour), 0),
					Count:  1,
					Events: make(map[string]*utils.DecimalWithCompress),
				},
			},
		},
	}

	if err := dm.SetStatQueue(context.Background(), sq); err != nil {
		t.Error(err)
	}

	experr := `unsupported metric type <testMetricType>`
	var reply string

	if err := sS.V1ResetStatQueue(context.Background(), &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			ID: "SQ1",
		},
	}, &reply); err == nil || err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}
}

func TestStatQueueProcessThresholdsOKNoThIDs(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.StatSCfg().Conns[utils.MetaThresholds] = []*config.DynamicConns{{ConnIDs: []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds)}}}
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	cacheS := engine.NewCacheS(cfg, dm, nil, nil)
	dm.SetCache(cacheS)

	filterS := engine.NewFilterS(cfg, nil, dm)
	sS := NewStatService(cfg, dm, cacheS, filterS, nil)

	sqPrf := &utils.StatQueueProfile{
		Tenant:    "cgrates.org",
		ID:        "SQ1",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
		Blockers:     utils.DynamicBlockers{{Blocker: true}},
		QueueLength:  10,
		ThresholdIDs: []string{"*none"},
		MinItems:     5,
		Metrics: []*utils.MetricWithFilters{
			{
				MetricID: "testMetricType",
			},
		},
	}
	sq := &utils.StatQueue{
		Tenant: "cgrates.org",
		ID:     "SQ1",
		SQItems: []utils.SQItem{
			{
				EventID: "SqProcessEvent",
			},
		},
		SQMetrics: map[string]utils.StatMetric{
			"testMetricType": &utils.StatTCD{
				Metric: &utils.Metric{
					Value:  utils.NewDecimal(int64(time.Hour), 0),
					Count:  1,
					Events: make(map[string]*utils.DecimalWithCompress),
				},
			},
		},
	}

	if err := dm.SetStatQueue(context.Background(), sq); err != nil {
		t.Error(err)
	}

	sQs := []*matchedStatQueue{{statQueue: sq, profile: sqPrf}}

	if err := sS.processThresholds(context.Background(), sQs, nil, "cgrates.org", nil); err != nil {
		t.Error(err)
	}
}

func TestStatQueueProcessThresholdsOK(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.StatSCfg().Conns[utils.MetaThresholds] = []*config.DynamicConns{{ConnIDs: []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds)}}}
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	cacheS := engine.NewCacheS(cfg, dm, nil, nil)
	dm.SetCache(cacheS)

	ccM := &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.ThresholdSv1ProcessEvent: func(ctx *context.Context, args, reply any) error {
				exp := &utils.CGREvent{
					Tenant: "cgrates.org",
					ID:     args.(*utils.CGREvent).ID,
					Event: map[string]any{
						utils.EventType:  utils.StatUpdate,
						utils.StatID:     "SQ1",
						"testMetricType": utils.NewDecimal(int64(time.Hour), 0),
					},
					APIOpts: map[string]any{
						utils.MetaEventType:            utils.StatUpdate,
						utils.OptsThresholdsProfileIDs: []string{"TH1"},
					},
				}
				if !reflect.DeepEqual(exp, args) {
					t.Errorf("Expected: <%+v>, \nreceived: <%+v>",
						utils.ToJSON(exp), utils.ToJSON(args))
					return fmt.Errorf("expected: <%+v>, \nreceived: <%+v>",
						utils.ToJSON(exp), utils.ToJSON(args))
				}
				return nil
			},
		},
	}
	rpcInternal := make(chan birpc.ClientConnector, 1)
	rpcInternal <- ccM
	cM := engine.NewConnManager(cfg)
	cM.SetCache(cacheS)
	cM.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds), utils.ThresholdSv1, rpcInternal)

	filterS := engine.NewFilterS(cfg, nil, dm)
	sS := NewStatService(cfg, dm, cacheS, filterS, cM)

	sqPrf := &utils.StatQueueProfile{
		Tenant:    "cgrates.org",
		ID:        "SQ1",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
		Blockers:     utils.DynamicBlockers{{Blocker: true}},
		QueueLength:  10,
		ThresholdIDs: []string{"TH1"},
		MinItems:     5,
		Metrics: []*utils.MetricWithFilters{
			{
				MetricID: "testMetricType",
			},
		},
	}
	sq := &utils.StatQueue{
		Tenant: "cgrates.org",
		ID:     "SQ1",
		SQItems: []utils.SQItem{
			{
				EventID: "SqProcessEvent",
			},
		},
		SQMetrics: map[string]utils.StatMetric{
			"testMetricType": &utils.StatTCD{
				Metric: &utils.Metric{
					Value: utils.NewDecimal(int64(time.Hour), 0),
					Count: 1,
					Events: map[string]*utils.DecimalWithCompress{
						"SqProcessEvent": {},
					},
				},
			},
		},
	}

	if err := dm.SetStatQueue(context.Background(), sq); err != nil {
		t.Error(err)
	}

	if err := sS.processThresholds(context.Background(), []*matchedStatQueue{{statQueue: sq, profile: sqPrf}}, nil, "cgrates.org", nil); err != nil {
		t.Error(err)
	}
}

func TestStatQueueProcessThresholdsErrPartExec(t *testing.T) {
	tmpLogger := utils.Logger
	defer func() {
		utils.Logger = tmpLogger
	}()
	var buf bytes.Buffer
	utils.Logger = utils.NewStdLoggerWithWriter(&buf, "", 4)

	cfg := config.NewDefaultCGRConfig()
	cfg.StatSCfg().Conns[utils.MetaThresholds] = []*config.DynamicConns{{ConnIDs: []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds)}}}
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	cacheS := engine.NewCacheS(cfg, dm, nil, nil)
	dm.SetCache(cacheS)

	ccM := &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.ThresholdSv1ProcessEvent: func(ctx *context.Context, args, reply any) error {
				return utils.ErrExists
			},
		},
	}
	rpcInternal := make(chan birpc.ClientConnector, 1)
	rpcInternal <- ccM
	cM := engine.NewConnManager(cfg)
	cM.SetCache(cacheS)
	cM.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds), utils.ThresholdSv1, rpcInternal)

	filterS := engine.NewFilterS(cfg, nil, dm)
	sS := NewStatService(cfg, dm, cacheS, filterS, cM)

	sqPrf := &utils.StatQueueProfile{
		Tenant:    "cgrates.org",
		ID:        "SQ1",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
		Blockers:     utils.DynamicBlockers{{Blocker: true}},
		QueueLength:  10,
		ThresholdIDs: []string{"TH1"},
		MinItems:     5,
		Metrics: []*utils.MetricWithFilters{
			{
				MetricID: "testMetricType",
			},
		},
	}
	sq := &utils.StatQueue{
		Tenant: "cgrates.org",
		ID:     "SQ1",
		SQItems: []utils.SQItem{
			{
				EventID: "SqProcessEvent",
			},
		},
		SQMetrics: map[string]utils.StatMetric{
			"testMetricType": &utils.StatTCD{
				Metric: &utils.Metric{
					Value:  utils.NewDecimal(int64(time.Hour), 0),
					Count:  1,
					Events: make(map[string]*utils.DecimalWithCompress),
				},
			},
		},
	}

	if err := dm.SetStatQueue(context.Background(), sq); err != nil {
		t.Error(err)
	}

	sQs := []*matchedStatQueue{{statQueue: sq, profile: sqPrf}}

	expLog := `[WARNING] <StatS> error: EXISTS`
	if err := sS.processThresholds(context.Background(), sQs, nil, "cgrates.org", nil); err == nil ||
		err != utils.ErrPartiallyExecuted {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrPartiallyExecuted, err)
	} else if rcvLog := buf.String(); !strings.Contains(rcvLog, expLog) {
		t.Errorf("expected log <%+v> to be included in <%+v>", expLog, rcvLog)
	}
}

func TestStatQueueV1GetQueueFloatMetricsOK(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.StatSCfg().StoreInterval = 1
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	cacheS := engine.NewCacheS(cfg, dm, nil, nil)
	dm.SetCache(cacheS)
	filterS := engine.NewFilterS(cfg, nil, dm)
	sS := NewStatService(cfg, dm, cacheS, filterS, nil)

	sq := &utils.StatQueue{
		Tenant: "cgrates.org",
		ID:     "SQ1",
		SQItems: []utils.SQItem{
			{
				EventID: "SqProcessEvent",
			},
		},
		SQMetrics: map[string]utils.StatMetric{
			utils.MetaTCD: &utils.StatTCD{
				Metric: &utils.Metric{
					Value: utils.NewDecimal(int64(time.Hour), 0),
					Count: 1,
					Events: map[string]*utils.DecimalWithCompress{
						"SqProcessEvent": {},
					},
				},
			},
		},
	}

	if err := dm.SetStatQueue(context.Background(), sq); err != nil {
		t.Error(err)
	}

	expected := map[string]float64{
		utils.MetaTCD: 3600000000000,
	}
	reply := map[string]float64{}
	if err := sS.V1GetQueueFloatMetrics(context.Background(), &utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{ID: "SQ1"}}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", expected, reply)
	}
}

func TestStatQueueV1GetQueueFloatMetricsErrNotFound(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.StatSCfg().StoreInterval = 1
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	cacheS := engine.NewCacheS(cfg, dm, nil, nil)
	dm.SetCache(cacheS)
	filterS := engine.NewFilterS(cfg, nil, dm)
	sS := NewStatService(cfg, dm, cacheS, filterS, nil)

	sq := &utils.StatQueue{
		Tenant: "cgrates.org",
		ID:     "SQ1",
		SQItems: []utils.SQItem{
			{
				EventID: "SqProcessEvent",
			},
		},
		SQMetrics: map[string]utils.StatMetric{
			utils.MetaTCD: &utils.StatTCD{
				Metric: &utils.Metric{
					Value: utils.NewDecimal(int64(time.Hour), 0),
					Count: 1,
					Events: map[string]*utils.DecimalWithCompress{
						"SqProcessEvent": {},
					},
				},
			},
		},
	}

	if err := dm.SetStatQueue(context.Background(), sq); err != nil {
		t.Error(err)
	}

	reply := map[string]float64{}
	if err := sS.V1GetQueueFloatMetrics(context.Background(), &utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{ID: "SQ2"}}, &reply); err == nil || err != utils.ErrNotFound {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}
}

func TestStatQueueV1GetQueueFloatMetricsMissingArgs(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.StatSCfg().StoreInterval = 1
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	cacheS := engine.NewCacheS(cfg, dm, nil, nil)
	dm.SetCache(cacheS)
	filterS := engine.NewFilterS(cfg, nil, dm)
	sS := NewStatService(cfg, dm, cacheS, filterS, nil)

	sq := &utils.StatQueue{
		Tenant: "cgrates.org",
		ID:     "SQ1",
		SQItems: []utils.SQItem{
			{
				EventID: "SqProcessEvent",
			},
		},
		SQMetrics: map[string]utils.StatMetric{
			utils.MetaTCD: &utils.StatTCD{
				Metric: &utils.Metric{
					Value:  utils.NewDecimal(int64(time.Hour), 0),
					Count:  1,
					Events: make(map[string]*utils.DecimalWithCompress),
				},
			},
		},
	}

	if err := dm.SetStatQueue(context.Background(), sq); err != nil {
		t.Error(err)
	}

	experr := `MANDATORY_IE_MISSING: [ID]`
	reply := map[string]float64{}
	if err := sS.V1GetQueueFloatMetrics(context.Background(), &utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{}}, &reply); err == nil ||
		err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}
}

func TestStatQueueV1GetQueueFloatMetricsErrGetStats(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.StatSCfg().StoreInterval = 1
	cacheS := engine.NewCacheS(cfg, nil, nil, nil)
	filterS := engine.NewFilterS(cfg, nil, nil)
	sS := NewStatService(cfg, nil, cacheS, filterS, nil)

	experr := `SERVER_ERROR: NO_DATABASE_CONNECTION`
	reply := map[string]float64{}
	if err := sS.V1GetQueueFloatMetrics(context.Background(), &utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{ID: "SQ1"}}, &reply); err == nil || err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}
}

func TestStatQueueV1GetQueueStringMetricsOK(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.StatSCfg().StoreInterval = 1
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	cacheS := engine.NewCacheS(cfg, dm, nil, nil)
	dm.SetCache(cacheS)
	filterS := engine.NewFilterS(cfg, nil, dm)
	sS := NewStatService(cfg, dm, cacheS, filterS, nil)

	sq := &utils.StatQueue{
		Tenant: "cgrates.org",
		ID:     "SQ1",
		SQItems: []utils.SQItem{
			{
				EventID: "SqProcessEvent",
			},
		},
		SQMetrics: map[string]utils.StatMetric{
			utils.MetaTCD: &utils.StatTCD{
				Metric: &utils.Metric{
					Value: utils.NewDecimal(int64(time.Hour), 0),
					Count: 1,
					Events: map[string]*utils.DecimalWithCompress{
						"SqProcessEvent": {},
					},
				},
			},
		},
	}

	if err := dm.SetStatQueue(context.Background(), sq); err != nil {
		t.Error(err)
	}

	expected := map[string]string{
		utils.MetaTCD: "1h0m0s",
	}
	reply := map[string]string{}
	if err := sS.V1GetQueueStringMetrics(context.Background(), &utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{ID: "SQ1"}}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", expected, reply)
	}
}

func TestStatQueueV1GetQueueStringMetricsErrNotFound(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.StatSCfg().StoreInterval = 1
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	cacheS := engine.NewCacheS(cfg, dm, nil, nil)
	dm.SetCache(cacheS)
	filterS := engine.NewFilterS(cfg, nil, dm)
	sS := NewStatService(cfg, dm, cacheS, filterS, nil)

	sq := &utils.StatQueue{
		Tenant: "cgrates.org",
		ID:     "SQ1",
		SQItems: []utils.SQItem{
			{
				EventID: "SqProcessEvent",
			},
		},
		SQMetrics: map[string]utils.StatMetric{
			utils.MetaTCD: &utils.StatTCD{
				Metric: &utils.Metric{
					Value:  utils.NewDecimal(int64(time.Hour), 0),
					Count:  1,
					Events: make(map[string]*utils.DecimalWithCompress),
				},
			},
		},
	}

	if err := dm.SetStatQueue(context.Background(), sq); err != nil {
		t.Error(err)
	}

	reply := map[string]string{}
	if err := sS.V1GetQueueStringMetrics(context.Background(), &utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{ID: "SQ2"}}, &reply); err == nil || err != utils.ErrNotFound {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}
}

func TestStatQueueV1GetQueueStringMetricsMissingArgs(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.StatSCfg().StoreInterval = 1
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	cacheS := engine.NewCacheS(cfg, dm, nil, nil)
	dm.SetCache(cacheS)
	filterS := engine.NewFilterS(cfg, nil, dm)
	sS := NewStatService(cfg, dm, cacheS, filterS, nil)

	sq := &utils.StatQueue{
		Tenant: "cgrates.org",
		ID:     "SQ1",
		SQItems: []utils.SQItem{
			{
				EventID: "SqProcessEvent",
			},
		},
		SQMetrics: map[string]utils.StatMetric{
			utils.MetaTCD: &utils.StatTCD{
				Metric: &utils.Metric{
					Value:  utils.NewDecimal(int64(time.Hour), 0),
					Count:  1,
					Events: make(map[string]*utils.DecimalWithCompress),
				},
			},
		},
	}

	if err := dm.SetStatQueue(context.Background(), sq); err != nil {
		t.Error(err)
	}

	experr := `MANDATORY_IE_MISSING: [ID]`
	reply := map[string]string{}
	if err := sS.V1GetQueueStringMetrics(context.Background(), &utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{}}, &reply); err == nil ||
		err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}
}

func TestStatQueueV1GetQueueStringMetricsErrGetStats(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.StatSCfg().StoreInterval = 1
	cacheS := engine.NewCacheS(cfg, nil, nil, nil)
	filterS := engine.NewFilterS(cfg, nil, nil)
	sS := NewStatService(cfg, nil, cacheS, filterS, nil)

	experr := `SERVER_ERROR: NO_DATABASE_CONNECTION`
	reply := map[string]string{}
	if err := sS.V1GetQueueStringMetrics(context.Background(), &utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{ID: "SQ1"}}, &reply); err == nil || err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}
}

func TestStatQueueGetStatQueueOK(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.StatSCfg().StoreInterval = 1
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	cacheS := engine.NewCacheS(cfg, dm, nil, nil)
	dm.SetCache(cacheS)
	filterS := engine.NewFilterS(cfg, nil, dm)
	sS := NewStatService(cfg, dm, cacheS, filterS, nil)

	sq := &utils.StatQueue{
		Tenant: "cgrates.org",
		ID:     "SQ1",
		SQItems: []utils.SQItem{
			{
				EventID:    "SqProcessEvent",
				ExpiryTime: utils.TimePointer(time.Now()),
			},
		},
		SQMetrics: map[string]utils.StatMetric{
			utils.MetaTCD: &utils.StatTCD{
				Metric: &utils.Metric{
					Value:  utils.NewDecimal(int64(time.Hour), 0),
					Count:  1,
					Events: make(map[string]*utils.DecimalWithCompress),
				},
			},
		},
	}

	if err := dm.SetStatQueue(context.Background(), sq); err != nil {
		t.Error(err)
	}

	expSq := &utils.StatQueue{
		Tenant:  "cgrates.org",
		ID:      "SQ1",
		SQItems: []utils.SQItem{},
		SQMetrics: map[string]utils.StatMetric{
			utils.MetaTCD: &utils.StatTCD{
				Metric: &utils.Metric{
					Value:  utils.NewDecimal(int64(time.Hour), 0),
					Count:  1,
					Events: make(map[string]*utils.DecimalWithCompress),
				},
			},
		},
	}
	if rcv, err := sS.getStatQueue(context.Background(), "cgrates.org", "SQ1"); err != nil {
		t.Error(err)
	} else if utils.ToJSON(expSq) != utils.ToJSON(rcv) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>",
			utils.ToJSON(expSq), utils.ToJSON(rcv))
	}
}
func TestStatQueueProcessEventProfileIgnoreFilters(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	cacheS := engine.NewCacheS(cfg, nil, nil, nil)
	dm.SetCache(cacheS)
	filterS := engine.NewFilterS(cfg, nil, dm)
	sS := NewStatService(cfg, dm, cacheS, filterS, nil)
	cfg.StatSCfg().Opts.ProfileIgnoreFilters = []*config.DynamicBoolOpt{
		config.NewDynamicBoolOpt(nil, "", true, nil),
	}
	sqPrf := &utils.StatQueueProfile{
		Tenant:    "cgrates.org",
		ID:        "SQ1",
		FilterIDs: []string{"*string:~*req.Stat:testStatValue"},
	}
	sq := &utils.StatQueue{
		Tenant: "cgrates.org",
		ID:     "SQ1",
		SQItems: []utils.SQItem{
			{
				EventID: "SqProcessEvent",
			},
		},
		SQMetrics: make(map[string]utils.StatMetric),
	}

	if err := dm.SetStatQueueProfile(context.Background(), sqPrf, true); err != nil {
		t.Error(err)
	}
	if err := dm.SetStatQueue(context.Background(), sq); err != nil {
		t.Error(err)
	}
	//should match the stat queue for event because the option is false but the filter matches
	args2 := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "SqProcessEvent",
		Event: map[string]any{
			"Stat": "testStatValue",
		},
		APIOpts: map[string]any{
			utils.OptsStatsProfileIDs:      []string{"SQ1"},
			utils.MetaProfileIgnoreFilters: false,
		},
	}

	expIDs2 := []string{"SQ1"}
	if rcvIDs2, err := sS.processEvent(context.Background(), args2.Tenant, args2); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcvIDs2, expIDs2) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", expIDs2, rcvIDs2)
	}
	//should match the stat queue for event because the option is true even if the filter doesn't match
	args := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "SqProcessEvent",
		Event: map[string]any{
			"Stat": "testStatValue2",
		},
		APIOpts: map[string]any{
			utils.OptsStatsProfileIDs:      []string{"SQ1"},
			utils.MetaProfileIgnoreFilters: true,
		},
	}

	expIDs := []string{"SQ1"}
	if rcvIDs, err := sS.processEvent(context.Background(), args.Tenant, args); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcvIDs, expIDs) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", expIDs, rcvIDs)
	}
}

func TestStatQueueProcessEventProfileIgnoreFiltersError(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	cacheS := engine.NewCacheS(cfg, nil, nil, nil)
	dm.SetCache(cacheS)
	filterS := engine.NewFilterS(cfg, nil, dm)
	sS := NewStatService(cfg, dm, cacheS, filterS, nil)
	cfg.StatSCfg().Opts.ProfileIgnoreFilters = []*config.DynamicBoolOpt{
		config.NewDynamicBoolOpt(nil, "", true, nil),
	}
	sqPrf := &utils.StatQueueProfile{
		Tenant:    "cgrates.org",
		ID:        "SQ1",
		FilterIDs: []string{"*string:~*req.Stat:testStatValue"},
	}
	sq := &utils.StatQueue{
		Tenant: "cgrates.org",
		ID:     "SQ1",
		SQItems: []utils.SQItem{
			{
				EventID: "SqProcessEvent",
			},
		},
		SQMetrics: make(map[string]utils.StatMetric),
	}

	if err := dm.SetStatQueueProfile(context.Background(), sqPrf, true); err != nil {
		t.Error(err)
	}
	if err := dm.SetStatQueue(context.Background(), sq); err != nil {
		t.Error(err)
	}

	args2 := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "SqProcessEvent",
		Event: map[string]any{
			"Stat": "testStatValue",
		},
		APIOpts: map[string]any{
			utils.OptsStatsProfileIDs:      []string{"SQ1"},
			utils.MetaProfileIgnoreFilters: time.Second,
		},
	}

	if _, err := sS.processEvent(context.Background(), args2.Tenant, args2); err == nil || err.Error() != "cannot convert field: 1s to bool" {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", "cannot convert field: 1s to bool", err)
	}

}

func TestStatQueueV1GetStatQueuesForEventProfileIgnoreFilters(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.StatSCfg().Opts.ProfileIgnoreFilters = []*config.DynamicBoolOpt{
		config.NewDynamicBoolOpt(nil, "", true, nil),
	}
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	cacheS := engine.NewCacheS(cfg, dm, nil, nil)
	dm.SetCache(cacheS)
	filterS := engine.NewFilterS(cfg, nil, dm)
	sS := NewStatService(cfg, dm, cacheS, filterS, nil)

	sqPrf1 := &utils.StatQueueProfile{
		Tenant:    "cgrates.org",
		ID:        "SQ1",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
		Blockers:     utils.DynamicBlockers{{Blocker: true}},
		QueueLength:  10,
		ThresholdIDs: []string{"*none"},
		MinItems:     5,
		Metrics: []*utils.MetricWithFilters{
			{
				MetricID: utils.MetaTCD,
			},
		},
	}

	if err := dm.SetStatQueueProfile(context.Background(), sqPrf1, true); err != nil {
		t.Error(err)
	}

	args := &utils.CGREvent{
		ID: "TestGetStatQueuesForEvent",
		Event: map[string]any{
			utils.AccountField: "1002",
		},
		APIOpts: map[string]any{
			utils.OptsStatsProfileIDs:      []string{"SQ1"},
			utils.MetaProfileIgnoreFilters: true,
		},
	}

	exp := []string{"SQ1"}
	var reply []string
	if err := sS.V1GetStatQueuesForEvent(context.Background(), args, &reply); err != nil {
		t.Error(err)
	} else {
		sort.Strings(reply)
		if !reflect.DeepEqual(reply, exp) {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>", exp, reply)
		}
	}
}

func TestStatSV1GetQueueDecimalMetricsOK(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.StatSCfg().StoreInterval = 1
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	cacheS := engine.NewCacheS(cfg, dm, nil, nil)
	dm.SetCache(cacheS)
	filterS := engine.NewFilterS(cfg, nil, dm)
	sS := NewStatService(cfg, dm, cacheS, filterS, nil)

	sq := &utils.StatQueue{
		Tenant: "cgrates.org",
		ID:     "SQ1",
		SQItems: []utils.SQItem{
			{
				EventID: "SqProcessEvent",
			},
		},
		SQMetrics: map[string]utils.StatMetric{
			utils.MetaTCD: &utils.StatTCD{
				Metric: &utils.Metric{
					Value: utils.NewDecimal(int64(time.Hour), 0),
					Count: 1,
					Events: map[string]*utils.DecimalWithCompress{
						"SqProcessEvent": {},
					},
				},
			},
		},
	}

	if err := dm.SetStatQueue(context.Background(), sq); err != nil {
		t.Error(err)
	}

	expected := map[string]float64{
		utils.MetaTCD: 3600000000000,
	}
	var reply map[string]*utils.Decimal
	if err := sS.V1GetQueueDecimalMetrics(context.Background(), &utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{ID: "SQ1"}}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(utils.ToJSON(expected), utils.ToJSON(reply)) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ToJSON(expected), utils.ToJSON(reply))
	}
}

func TestStatSV1GetQueueDecimalMetricsErrNotFound(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.StatSCfg().StoreInterval = 1
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	cacheS := engine.NewCacheS(cfg, dm, nil, nil)
	dm.SetCache(cacheS)
	filterS := engine.NewFilterS(cfg, nil, dm)
	sS := NewStatService(cfg, dm, cacheS, filterS, nil)

	sq := &utils.StatQueue{
		Tenant: "cgrates.org",
		ID:     "SQ1",
		SQItems: []utils.SQItem{
			{
				EventID: "SqProcessEvent",
			},
		},
		SQMetrics: map[string]utils.StatMetric{
			utils.MetaTCD: &utils.StatTCD{
				Metric: &utils.Metric{
					Value: utils.NewDecimal(int64(time.Hour), 0),
					Count: 1,
					Events: map[string]*utils.DecimalWithCompress{
						"SqProcessEvent": {},
					},
				},
			},
		},
	}

	if err := dm.SetStatQueue(context.Background(), sq); err != nil {
		t.Error(err)
	}

	var reply map[string]*utils.Decimal
	if err := sS.V1GetQueueDecimalMetrics(context.Background(), &utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{ID: "SQ2"}}, &reply); err == nil || err != utils.ErrNotFound {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}
}

func TestStatV1GetQueueDecimalMetricsMissingArgs(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.StatSCfg().StoreInterval = 1
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	cacheS := engine.NewCacheS(cfg, dm, nil, nil)
	dm.SetCache(cacheS)
	filterS := engine.NewFilterS(cfg, nil, dm)
	sS := NewStatService(cfg, dm, cacheS, filterS, nil)

	sq := &utils.StatQueue{
		Tenant: "cgrates.org",
		ID:     "SQ1",
		SQItems: []utils.SQItem{
			{
				EventID: "SqProcessEvent",
			},
		},
		SQMetrics: map[string]utils.StatMetric{
			utils.MetaTCD: &utils.StatTCD{
				Metric: &utils.Metric{
					Value:  utils.NewDecimal(int64(time.Hour), 0),
					Count:  1,
					Events: make(map[string]*utils.DecimalWithCompress),
				},
			},
		},
	}

	if err := dm.SetStatQueue(context.Background(), sq); err != nil {
		t.Error(err)
	}

	experr := `MANDATORY_IE_MISSING: [ID]`
	var reply map[string]*utils.Decimal
	if err := sS.V1GetQueueDecimalMetrics(context.Background(), &utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{}}, &reply); err == nil ||
		err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}
}

func TestStatV1GetQueueDecimalMetricsErrGetStats(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.StatSCfg().StoreInterval = 1
	cacheS := engine.NewCacheS(cfg, nil, nil, nil)
	filterS := engine.NewFilterS(cfg, nil, nil)
	sS := NewStatService(cfg, nil, cacheS, filterS, nil)

	experr := `SERVER_ERROR: NO_DATABASE_CONNECTION`
	var reply map[string]*utils.Decimal
	if err := sS.V1GetQueueDecimalMetrics(context.Background(), &utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{ID: "SQ1"}}, &reply); err == nil || err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}
}

func TestStatSV1GetQueueStringMetricsIntOptsErr(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dmSTS := engine.NewDataManager(dbCM, cfg, nil)
	cacheS := engine.NewCacheS(cfg, nil, nil, nil)
	dmSTS.SetCache(cacheS)

	cfg.StatSCfg().StoreInterval = 1
	cfg.StatSCfg().StringIndexedFields = nil
	cfg.StatSCfg().PrefixIndexedFields = nil
	cfg.StatSCfg().Opts.RoundingDecimals = []*config.DynamicIntOpt{
		// function will return error after trying to parse the filter
		config.NewDynamicIntOpt([]string{"*string.invalid:filter"}, "cgrates.org", 4, nil),
	}
	statService := NewStatService(cfg, dmSTS, cacheS,
		engine.NewFilterS(cfg, nil, dmSTS), nil)

	prepareStatsData(t, dmSTS)

	stq := map[string]string{}

	experr := `inline parse error for string: <*string.invalid:filter>`

	err := statService.V1GetQueueStringMetrics(context.TODO(), &utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: testStatsQ[0].Tenant, ID: testStatsQ[0].ID}}, &stq)
	if err.Error() != experr {
		t.Errorf("Expected error <%v>, Received <%v>", experr, err)
	}

}

func TestStatSV1GetStatQueuesForEventsqIDsErr(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.StatSCfg().Opts.ProfileIDs = []*config.DynamicStringSliceOpt{
		// function will return error after trying to parse the filter
		{
			FilterIDs: []string{"*string.invalid:filter"},
			Tenant:    "cgrates.org",
			Values:    []string{"value2"},
		},
	}
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	cacheS := engine.NewCacheS(cfg, dm, nil, nil)
	dm.SetCache(cacheS)
	filterS := engine.NewFilterS(cfg, nil, dm)
	sS := NewStatService(cfg, dm, cacheS, filterS, nil)

	sqPrf1 := &utils.StatQueueProfile{
		Tenant:    "cgrates.org",
		ID:        "SQ1",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
		Blockers:     utils.DynamicBlockers{{Blocker: true}},
		QueueLength:  10,
		ThresholdIDs: []string{"*none"},
		MinItems:     5,
		Metrics: []*utils.MetricWithFilters{
			{
				MetricID: utils.MetaTCD,
			},
		},
	}

	if err := dm.SetStatQueueProfile(context.Background(), sqPrf1, true); err != nil {
		t.Error(err)
	}

	sqPrf2 := &utils.StatQueueProfile{
		Tenant: "cgrates.org",
		ID:     "SQ2",
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
		Blockers:     utils.DynamicBlockers{{Blocker: false}},
		QueueLength:  10,
		ThresholdIDs: []string{"*none"},
		MinItems:     5,
		Metrics: []*utils.MetricWithFilters{
			{
				MetricID: utils.MetaACD,
			},
		},
	}

	if err := dm.SetStatQueueProfile(context.Background(), sqPrf2, true); err != nil {
		t.Error(err)
	}

	args := &utils.CGREvent{
		ID: "TestGetStatQueuesForEvent",
		Event: map[string]any{
			utils.AccountField: "1001",
		},
	}

	experr := `inline parse error for string: <*string.invalid:filter>`
	var reply []string
	if err := sS.V1GetStatQueuesForEvent(context.Background(), args, &reply); err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}
}

func TestStatSV1GetStatQueuesForEventignFiltersErr(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.StatSCfg().Opts.ProfileIgnoreFilters = []*config.DynamicBoolOpt{
		// function will return error after trying to parse the filter
		config.NewDynamicBoolOpt([]string{"*string.invalid:filter"}, "cgrates.org", false, nil),
	}
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	cacheS := engine.NewCacheS(cfg, dm, nil, nil)
	dm.SetCache(cacheS)
	filterS := engine.NewFilterS(cfg, nil, dm)
	sS := NewStatService(cfg, dm, cacheS, filterS, nil)

	sqPrf1 := &utils.StatQueueProfile{
		Tenant:    "cgrates.org",
		ID:        "SQ1",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
		Blockers:     utils.DynamicBlockers{{Blocker: true}},
		QueueLength:  10,
		ThresholdIDs: []string{"*none"},
		MinItems:     5,
		Metrics: []*utils.MetricWithFilters{
			{
				MetricID: utils.MetaTCD,
			},
		},
	}

	if err := dm.SetStatQueueProfile(context.Background(), sqPrf1, true); err != nil {
		t.Error(err)
	}

	sqPrf2 := &utils.StatQueueProfile{
		Tenant: "cgrates.org",
		ID:     "SQ2",
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
		Blockers:     utils.DynamicBlockers{{Blocker: false}},
		QueueLength:  10,
		ThresholdIDs: []string{"*none"},
		MinItems:     5,
		Metrics: []*utils.MetricWithFilters{
			{
				MetricID: utils.MetaACD,
			},
		},
	}

	if err := dm.SetStatQueueProfile(context.Background(), sqPrf2, true); err != nil {
		t.Error(err)
	}

	args := &utils.CGREvent{
		ID: "TestGetStatQueuesForEvent",
		Event: map[string]any{
			utils.AccountField: "1001",
		},
	}

	experr := `inline parse error for string: <*string.invalid:filter>`
	var reply []string
	if err := sS.V1GetStatQueuesForEvent(context.Background(), args, &reply); err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}
}

func TestStatQueuesProcessEventidsErr(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dmSTS := engine.NewDataManager(dbCM, cfg, nil)
	cacheS := engine.NewCacheS(cfg, nil, nil, nil)
	dmSTS.SetCache(cacheS)

	cfg.StatSCfg().StoreInterval = 1
	cfg.StatSCfg().StringIndexedFields = nil
	cfg.StatSCfg().PrefixIndexedFields = nil
	statService := NewStatService(cfg, dmSTS, cacheS,
		engine.NewFilterS(cfg, nil, dmSTS), nil)

	prepareStatsData(t, dmSTS)
	args := &utils.CGREvent{
		Tenant: utils.EmptyString,
		ID:     "SqProcessEvent",
		Event: map[string]any{
			utils.AccountField: "1002",
		},
	}

	reply := []string{}
	err := statService.V1ProcessEvent(context.TODO(), args, &reply)
	if err != utils.ErrNotFound {
		t.Errorf("Expecting error: %+v, received error: %+v", utils.ErrNotFound, err)
	}
}

func TestStatSMatchingStatQueuesForEventNoSqs(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dmSTS := engine.NewDataManager(dbCM, cfg, nil)
	cacheS := engine.NewCacheS(cfg, nil, nil, nil)
	dmSTS.SetCache(cacheS)

	cfg.StatSCfg().StoreInterval = 1
	cfg.StatSCfg().StringIndexedFields = nil
	cfg.StatSCfg().PrefixIndexedFields = nil

	statService := NewStatService(cfg, dmSTS, cacheS,
		engine.NewFilterS(cfg, nil, dmSTS), nil)
	prepareStatsData(t, dmSTS)
	ev := testStatsArgs[0].Clone()
	ev.APIOpts = map[string]any{utils.OptsStatsProfileIDs: []string{"statsIds"}}
	_, _, err := statService.matchingStatQueuesForEvent(context.TODO(), ev.Tenant, ev)
	if err == nil || err != utils.ErrNotFound {
		t.Errorf("Expected error <%v>, Received error <%v>", utils.ErrNotFound, err)
	}
}

func TestStatQueuesMatchingStatQueuesForEventWeightErr(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dmSTS := engine.NewDataManager(dbCM, cfg, nil)
	cacheS := engine.NewCacheS(cfg, nil, nil, nil)
	dmSTS.SetCache(cacheS)

	cfg.StatSCfg().StoreInterval = 1
	cfg.StatSCfg().StringIndexedFields = nil
	cfg.StatSCfg().PrefixIndexedFields = nil
	statService := NewStatService(cfg, dmSTS, cacheS,
		engine.NewFilterS(cfg, nil, dmSTS), nil)
	prepareStatsData(t, dmSTS)

	sqp := &utils.StatQueueProfile{
		Tenant:      "cgrates.org",
		ID:          "StatQueueProfile1",
		FilterIDs:   []string{"FLTR_STATS_1"},
		QueueLength: 10,
		TTL:         10 * time.Second,
		Metrics: []*utils.MetricWithFilters{
			{
				MetricID: "*sum#~*req.Usage",
			},
		},
		ThresholdIDs: []string{},
		Stored:       true,
		Weights: utils.DynamicWeights{
			{
				FilterIDs: []string{"*stirng:~*req.Account:1001"},
				Weight:    64,
			},
		},
		MinItems: 1,
	}

	if err := statService.dm.SetStatQueueProfile(context.Background(), sqp, true); err != nil {
		t.Error(err)
	}

	expErr := "Unsupported filter Type: *stirng"
	_, _, err := statService.matchingStatQueuesForEvent(context.TODO(), testStatsArgs[0].Tenant, testStatsArgs[0])

	if err == nil || err.Error() != expErr {
		t.Errorf("Expected error <%v>, Received error <%v>", expErr, err)
	}
}

func TestStatQueueProcessEventProfileIDsErr(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	cacheS := engine.NewCacheS(cfg, nil, nil, nil)
	dm.SetCache(cacheS)
	filterS := engine.NewFilterS(cfg, nil, dm)
	sS := NewStatService(cfg, dm, cacheS, filterS, nil)

	args := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "SqProcessEvent",
		Event: map[string]any{
			utils.AccountField: "1001",
		},
		APIOpts: map[string]any{},
	}
	sS.cfg.StatSCfg().Opts.ProfileIDs = []*config.DynamicStringSliceOpt{

		{
			FilterIDs: []string{"*string.invalid:filter"},
			Tenant:    "cgrates.org",
			Values:    []string{"value2"},
		},
	}

	experr := `inline parse error for string: <*string.invalid:filter>`
	if _, err := sS.processEvent(context.Background(), args.Tenant, args); err == nil ||
		err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}

}

func TestStatQueueProcessEventExpiredErr(t *testing.T) {

	tmpl := utils.Logger
	defer func() {
		utils.Logger = tmpl
	}()

	buf := new(bytes.Buffer)
	utils.Logger = utils.NewStdLoggerWithWriter(buf, "", 7)

	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	cacheS := engine.NewCacheS(cfg, nil, nil, nil)
	dm.SetCache(cacheS)
	filterS := engine.NewFilterS(cfg, nil, dm)
	sS := NewStatService(cfg, dm, cacheS, filterS, nil)

	sqPrf := &utils.StatQueueProfile{
		Tenant:    "cgrates.org",
		ID:        "SQ1",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
		Blockers:     utils.DynamicBlockers{{Blocker: true}},
		QueueLength:  -1,
		ThresholdIDs: []string{"*none"},
		MinItems:     5,
		Metrics: []*utils.MetricWithFilters{
			{
				MetricID: utils.MetaTCD,
			},
		},
	}
	expiry := time.Date(2021, 1, 1, 23, 59, 59, 10, time.UTC)

	stq := &utils.StatQueue{
		Tenant: "cgrates.org",
		ID:     "SQ1",
		SQItems: []utils.SQItem{
			{
				EventID:    "SqProcessEvent",
				ExpiryTime: &expiry,
			},
		},
		SQMetrics: map[string]utils.StatMetric{
			"key": statMetricMock("remExpired error"),
		},
	}

	if err := dm.SetStatQueueProfile(context.Background(), sqPrf, true); err != nil {
		t.Error(err)
	}
	if err := dm.SetStatQueue(context.Background(), stq); err != nil {
		t.Error(err)
	}

	args := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "SqProcessEvent",
		Event: map[string]any{
			utils.AccountField: "1001",
		},
		APIOpts: map[string]any{
			utils.MetaUsage:           "10s",
			utils.OptsStatsProfileIDs: []string{"SQ1"},
		},
	}

	if rcv, err := sS.processEvent(context.Background(), args.Tenant, args); err == nil || err.Error() != utils.ErrPartiallyExecuted.Error() {
		t.Errorf("*StatS.processEvent err=%v, want %v", err, utils.ErrPartiallyExecuted)
	} else if !reflect.DeepEqual([]string{"SQ1"}, rcv) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", nil, rcv)

	}
	expErr := "cgrates.org:SQ1, ignoring event: cgrates.org:SqProcessEvent, error: remExpired mock error"
	if rcvTxt := buf.String(); !strings.Contains(rcvTxt, expErr) {
		t.Errorf("Expected <%v>, Received <%v>", expErr, rcvTxt)
	}

	buf.Reset()

}

func TestStatQueueProcessEventBlockerErr(t *testing.T) {

	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	cacheS := engine.NewCacheS(cfg, nil, nil, nil)
	dm.SetCache(cacheS)
	filterS := engine.NewFilterS(cfg, nil, dm)
	sS := NewStatService(cfg, dm, cacheS, filterS, nil)

	sqPrf := &utils.StatQueueProfile{
		Tenant:    "cgrates.org",
		ID:        "SQ1",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
		Blockers: []*utils.DynamicBlocker{
			{
				FilterIDs: []string{"*stirng:~*req.Account:1001"},
				Blocker:   true,
			},
		},
		QueueLength:  10,
		ThresholdIDs: []string{"*none"},
		MinItems:     5,
		Metrics: []*utils.MetricWithFilters{
			{
				MetricID: utils.MetaTCD,
			},
		},
	}
	stat, err := utils.NewStatMetric("*tcd", uint64(sqPrf.MinItems), []string{})
	if err != nil {
		t.Fatal(err)
	}
	stq := &utils.StatQueue{
		Tenant: "cgrates.org",
		ID:     "SQ1",
		SQItems: []utils.SQItem{
			{
				EventID: "SqProcessEvent",
			},
		},
		SQMetrics: map[string]utils.StatMetric{
			utils.MetaTCD: stat,
		},
	}

	if err := dm.SetStatQueueProfile(context.Background(), sqPrf, true); err != nil {
		t.Error(err)
	}
	if err := dm.SetStatQueue(context.Background(), stq); err != nil {
		t.Error(err)
	}

	args := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "SqProcessEvent",
		Event: map[string]any{
			utils.AccountField: "1001",
		},
		APIOpts: map[string]any{
			utils.MetaUsage:           "10s",
			utils.OptsStatsProfileIDs: []string{"SQ1"},
		},
	}

	expErr := "Unsupported filter Type: *stirng"
	if _, err := sS.processEvent(context.Background(), args.Tenant, args); err == nil || err.Error() != expErr {
		t.Errorf("Expected error <%v>, Received error <%v>", expErr, err)
	}

}

func TestStatRemEventWithID(t *testing.T) {
	type step struct {
		rem        []string
		wantVal    *utils.Decimal
		wantEvents int
	}
	tests := []struct {
		name    string
		metric  *utils.StatASR
		initVal *utils.Decimal
		steps   []step
	}{
		{
			name: "compress factor 1",
			metric: &utils.StatASR{
				Metric: &utils.Metric{
					Value: utils.NewDecimal(1, 0),
					Count: 2,
					Events: map[string]*utils.DecimalWithCompress{
						"cgrates.org:TestRemEventWithID_1": {Stat: utils.NewDecimal(1, 0), CompressFactor: 1},
						"cgrates.org:TestRemEventWithID_2": {Stat: utils.NewDecimal(0, 0), CompressFactor: 1},
					},
				},
			},
			initVal: utils.NewDecimal(50, 0),
			steps: []step{
				{rem: []string{"cgrates.org:TestRemEventWithID_1"}, wantVal: utils.NewDecimal(0, 0), wantEvents: 1},
				{rem: []string{"cgrates.org:TestRemEventWithID_5"}, wantVal: utils.NewDecimal(0, 0), wantEvents: 1}, // non existent
				{rem: []string{"cgrates.org:TestRemEventWithID_2"}, wantVal: utils.DecimalNaN, wantEvents: 0},
				{rem: []string{"cgrates.org:TestRemEventWithID_2"}, wantVal: utils.DecimalNaN, wantEvents: 0},
			},
		},
		{
			name: "compress factor 2",
			metric: &utils.StatASR{
				Metric: &utils.Metric{
					Value: utils.NewDecimal(2, 0),
					Count: 4,
					Events: map[string]*utils.DecimalWithCompress{
						"cgrates.org:TestRemEventWithID_1": {Stat: utils.NewDecimal(1, 0), CompressFactor: 2},
						"cgrates.org:TestRemEventWithID_2": {Stat: utils.NewDecimal(0, 0), CompressFactor: 2},
					},
				},
			},
			initVal: utils.NewDecimal(50, 0),
			steps: []step{
				{rem: []string{"cgrates.org:TestRemEventWithID_1", "cgrates.org:TestRemEventWithID_2"}, wantVal: utils.NewDecimal(50, 0), wantEvents: 2},
				{rem: []string{"cgrates.org:TestRemEventWithID_5"}, wantVal: utils.NewDecimal(50, 0), wantEvents: 2}, // non existent
				{rem: []string{"cgrates.org:TestRemEventWithID_2", "cgrates.org:TestRemEventWithID_1"}, wantVal: utils.DecimalNaN, wantEvents: 0},
				{rem: []string{"cgrates.org:TestRemEventWithID_2"}, wantVal: utils.DecimalNaN, wantEvents: 0},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sq := &utils.StatQueue{SQMetrics: map[string]utils.StatMetric{utils.MetaASR: tt.metric}}
			m := tt.metric
			if got := m.GetValue(); got.Compare(tt.initVal) != 0 {
				t.Errorf("initial GetValue: %v", got)
			}
			for i, s := range tt.steps {
				for _, id := range s.rem {
					remEventWithID(sq, id)
				}
				if got := m.GetValue(); got.Compare(s.wantVal) != 0 {
					t.Errorf("step %d GetValue: %v", i, got)
				} else if len(m.Events) != s.wantEvents {
					t.Errorf("step %d Events: %+v", i, m.Events)
				}
			}
		})
	}
}

func TestStatRemExpired(t *testing.T) {
	sq := &utils.StatQueue{
		SQMetrics: map[string]utils.StatMetric{
			utils.MetaASR: &utils.StatASR{
				Metric: &utils.Metric{
					Value: utils.NewDecimal(2, 0),
					Count: 3,
					Events: map[string]*utils.DecimalWithCompress{
						"cgrates.org:TestStatRemExpired_1": {Stat: utils.NewDecimal(1, 0), CompressFactor: 1},
						"cgrates.org:TestStatRemExpired_2": {Stat: utils.NewDecimal(0, 0), CompressFactor: 1},
						"cgrates.org:TestStatRemExpired_3": {Stat: utils.NewDecimal(1, 0), CompressFactor: 1},
					},
				},
			},
		},
		SQItems: []utils.SQItem{
			{EventID: "cgrates.org:TestStatRemExpired_1", ExpiryTime: utils.TimePointer(time.Now())},
			{EventID: "cgrates.org:TestStatRemExpired_2", ExpiryTime: utils.TimePointer(time.Now())},
			{EventID: "cgrates.org:TestStatRemExpired_3", ExpiryTime: utils.TimePointer(time.Now().Add(time.Minute))},
		},
	}
	asrMetric := sq.SQMetrics[utils.MetaASR].(*utils.StatASR)
	if asr := asrMetric.GetValue(); asr.Compare(utils.NewDecimalFromFloat64(66.66666666666667)) != 0 {
		t.Errorf("received asrMetric: %v", asrMetric.GetValue())
	}
	remExpired(sq)
	if asr := asrMetric.GetValue(); asr.Compare(utils.NewDecimalFromFloat64(100)) != 0 {
		t.Errorf("received asrMetric: %v", asrMetric.GetValue())
	} else if len(asrMetric.Events) != 1 {
		t.Errorf("unexpected Events in asrMetric: %+v", asrMetric.Events)
	}
	if len(sq.SQItems) != 1 {
		t.Errorf("Unexpected items: %+v", sq.SQItems)
	}
}

func TestStatRemOnQueueLength(t *testing.T) {
	m := &matchedStatQueue{
		statQueue: &utils.StatQueue{
			SQItems: []utils.SQItem{
				{EventID: "cgrates.org:TestStatRemExpired_1"},
			},
		},
		profile: &utils.StatQueueProfile{
			QueueLength: 2,
		},
	}
	m.remOnQueueLength()
	if len(m.statQueue.SQItems) != 1 {
		t.Errorf("wrong items: %+v", m.statQueue.SQItems)
	}
	m.statQueue.SQItems = []utils.SQItem{
		{EventID: "cgrates.org:TestStatRemExpired_1"},
		{EventID: "cgrates.org:TestStatRemExpired_2"},
	}
	m.remOnQueueLength()
	if len(m.statQueue.SQItems) != 1 {
		t.Errorf("wrong items: %+v", m.statQueue.SQItems)
	} else if m.statQueue.SQItems[0].EventID != "cgrates.org:TestStatRemExpired_2" {
		t.Errorf("wrong item in SQItems: %+v", m.statQueue.SQItems[0])
	}
	m.profile.QueueLength = -1
	m.statQueue.SQItems = []utils.SQItem{
		{EventID: "cgrates.org:TestStatRemExpired_1"},
		{EventID: "cgrates.org:TestStatRemExpired_2"},
		{EventID: "cgrates.org:TestStatRemExpired_3"},
	}
	m.remOnQueueLength()
	if len(m.statQueue.SQItems) != 3 {
		t.Errorf("wrong items: %+v", m.statQueue.SQItems)
	}
}

func TestStatAddStatEvent(t *testing.T) {
	m := &matchedStatQueue{
		cfg: config.NewDefaultCGRConfig(),
		statQueue: &utils.StatQueue{
			SQMetrics: map[string]utils.StatMetric{
				utils.MetaASR: &utils.StatASR{
					Metric: &utils.Metric{
						Value: utils.NewDecimal(1, 0),
						Count: 1,
						Events: map[string]*utils.DecimalWithCompress{
							"cgrates.org:TestStatRemExpired_1": {Stat: utils.NewDecimalFromFloat64(1), CompressFactor: 1},
						},
					},
				},
			},
		},
		profile: &utils.StatQueueProfile{
			Metrics: []*utils.MetricWithFilters{
				{
					MetricID: utils.MetaASR,
				},
			},
		},
	}
	asrMetric := m.statQueue.SQMetrics[utils.MetaASR].(*utils.StatASR)
	if asr := asrMetric.GetValue(); asr.Compare(utils.NewDecimalFromFloat64(100)) != 0 {
		t.Errorf("received ASR: %v", asr)
	}
	ev1 := &utils.CGREvent{Tenant: "cgrates.org", ID: "TestStatAddStatEvent_1"}
	m.addStatEvent(context.Background(), ev1.Tenant, ev1.ID, utils.MapStorage{utils.MetaOpts: ev1.Event})
	if asr := asrMetric.GetValue(); asr.Compare(utils.NewDecimalFromFloat64(50)) != 0 {
		t.Errorf("received ASR: %v", asr)
	} else if asrMetric.Value.Compare(utils.NewDecimal(1, 0)) != 0 || asrMetric.Count != 2 {
		t.Errorf("ASR: %v", asrMetric)
	}
	ev1.APIOpts = map[string]any{
		utils.MetaStartTime: time.Now()}
	m.addStatEvent(context.Background(), ev1.Tenant, ev1.ID, utils.MapStorage{utils.MetaOpts: ev1.APIOpts})
	if asr := asrMetric.GetValue(); asr.Compare(utils.NewDecimalFromFloat64(66.66666666666667)) != 0 {
		t.Errorf("received ASR: %v", asr)
	} else if asrMetric.Value.Compare(utils.NewDecimal(2, 0)) != 0 || asrMetric.Count != 3 {
		t.Errorf("ASR: %v", asrMetric)
	}
}

func TestStatRemOnQueueLength2(t *testing.T) {
	m := &matchedStatQueue{
		statQueue: &utils.StatQueue{
			SQItems: []utils.SQItem{
				{EventID: "cgrates.org:TestStatRemExpired_1"},
				{EventID: "cgrates.org:TestStatRemExpired_2"},
			},
			SQMetrics: map[string]utils.StatMetric{
				utils.MetaTCD: &utils.StatTCD{
					Metric: &utils.Metric{
						FilterIDs: []string{"*string:~*req.Account:1002"},
						Value:     utils.NewDecimal(0, 0),
						Events: map[string]*utils.DecimalWithCompress{
							"cgrates.org:TestStatRemExpired_2": {Stat: utils.NewDecimalFromFloat64(float64(time.Minute)), CompressFactor: 1},
						},
					},
				},
				utils.MetaASR: &utils.StatASR{
					Metric: &utils.Metric{
						FilterIDs: []string{"*string:~*req.Account:1001"},
						Value:     utils.NewDecimal(0, 0),
						Events: map[string]*utils.DecimalWithCompress{
							"cgrates.org:TestStatRemExpired_1": {Stat: utils.NewDecimalFromFloat64(1), CompressFactor: 1},
						},
					},
				},
			},
		},
		profile: &utils.StatQueueProfile{
			QueueLength: 2,
			FilterIDs:   []string{"*string:~Account:1001|1002"},
		},
	}
	m.remOnQueueLength()
	if len(m.statQueue.SQItems) != 1 {
		t.Errorf("wrong items: %+v", utils.ToJSON(m.statQueue.SQItems))
	}
}

func TestStatRemoveExpiredTTL(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		m := &matchedStatQueue{
			cfg: config.NewDefaultCGRConfig(),
			statQueue: &utils.StatQueue{
				SQMetrics: map[string]utils.StatMetric{
					utils.MetaASR: &utils.StatASR{
						Metric: &utils.Metric{
							Value: utils.NewDecimal(1, 0),
							Count: 1,
							Events: map[string]*utils.DecimalWithCompress{
								"grates.org:TestStatRemExpired_1": {Stat: utils.NewDecimal(1, 0), CompressFactor: 1},
							},
						},
					},
				},
			},
			profile: &utils.StatQueueProfile{
				QueueLength: 0, //unlimited que
			},
			ttl: utils.DurationPointer(100 * time.Millisecond),
		}

		//add ev1 with ttl 100ms (after 100ms the event should be removed)
		ev1 := &utils.CGREvent{Tenant: "cgrates.org", ID: "TestStatAddStatEvent_1"}
		m.processEvent(context.Background(), ev1.Tenant, ev1.ID, utils.MapStorage{utils.MetaReq: ev1.Event})

		if len(m.statQueue.SQItems) != 1 && m.statQueue.SQItems[0].EventID != "TestStatAddStatEvent_1" {
			t.Errorf("Expecting: 1, received: %+v", len(m.statQueue.SQItems))
		}
		//after 150ms the event expired
		time.Sleep(150 * time.Millisecond)

		//processing a new event should clean the expired events and add the new one
		ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "TestStatAddStatEvent_2"}
		m.processEvent(context.Background(), ev2.Tenant, ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
		if len(m.statQueue.SQItems) != 1 && m.statQueue.SQItems[0].EventID != "TestStatAddStatEvent_2" {
			t.Errorf("Expecting: 1, received: %+v", len(m.statQueue.SQItems))
		}
	})
}

func TestStatRemoveExpiredQueue(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		m := &matchedStatQueue{
			cfg: config.NewDefaultCGRConfig(),
			statQueue: &utils.StatQueue{
				SQMetrics: map[string]utils.StatMetric{
					utils.MetaASR: &utils.StatASR{
						Metric: &utils.Metric{
							Value: utils.NewDecimal(1, 0),
							Count: 1,
							Events: map[string]*utils.DecimalWithCompress{
								"grates.org:TestStatRemExpired_1": {Stat: utils.NewDecimal(1, 0), CompressFactor: 1},
							},
						},
					},
				},
			},
			profile: &utils.StatQueueProfile{
				QueueLength: 2, //unlimited que
			},
		}

		//add ev1 with ttl 100ms (after 100ms the event should be removed)
		ev1 := &utils.CGREvent{Tenant: "cgrates.org", ID: "TestStatAddStatEvent_1"}
		m.processEvent(context.Background(), ev1.Tenant, ev1.ID, utils.MapStorage{utils.MetaReq: ev1.Event})

		if len(m.statQueue.SQItems) != 1 && m.statQueue.SQItems[0].EventID != "TestStatAddStatEvent_1" {
			t.Errorf("Expecting: 1, received: %+v", len(m.statQueue.SQItems))
		}
		//after 150ms the event expired
		time.Sleep(150 * time.Millisecond)

		//processing a new event should clean the expired events and add the new one
		ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "TestStatAddStatEvent_2"}
		m.processEvent(context.Background(), ev2.Tenant, ev2.ID, utils.MapStorage{utils.MetaReq: ev2.Event})
		if len(m.statQueue.SQItems) != 2 && m.statQueue.SQItems[0].EventID != "TestStatAddStatEvent_1" &&
			m.statQueue.SQItems[1].EventID != "TestStatAddStatEvent_2" {
			t.Errorf("Expecting: 2, received: %+v", len(m.statQueue.SQItems))
		}

		//processing a new event should clean the expired events and add the new one
		ev3 := &utils.CGREvent{Tenant: "cgrates.org", ID: "TestStatAddStatEvent_3"}
		m.processEvent(context.Background(), ev3.Tenant, ev3.ID, utils.MapStorage{utils.MetaReq: ev3.Event})
		if len(m.statQueue.SQItems) != 2 && m.statQueue.SQItems[0].EventID != "TestStatAddStatEvent_2" &&
			m.statQueue.SQItems[1].EventID != "TestStatAddStatEvent_3" {
			t.Errorf("Expecting: 2, received: %+v", len(m.statQueue.SQItems))
		}
	})
}

func TestStatQueueProcessEventremExpiredErr(t *testing.T) {
	tnt, evID := "tenant", "eventID"
	expiry := time.Date(2021, 1, 1, 23, 59, 59, 10, time.UTC)
	evNm := utils.MapStorage{
		"key": nil,
	}

	m := &matchedStatQueue{
		statQueue: &utils.StatQueue{
			SQItems: []utils.SQItem{
				{
					EventID:    evID,
					ExpiryTime: &expiry,
				},
			},
			SQMetrics: map[string]utils.StatMetric{
				"key": statMetricMock("remExpired error"),
			},
		},
		profile: &utils.StatQueueProfile{
			QueueLength: -1,
		},
	}

	experr := "remExpired mock error"
	err := m.processEvent(context.Background(), tnt, evID, evNm)

	if err == nil || err.Error() != experr {
		t.Errorf("\nexpected: %q, \nreceived: %q", experr, err)
	}
}

func TestStatQueueProcessEventremOnQueueLengthErr(t *testing.T) {
	tnt, evID := "tenant", "eventID"
	evNm := utils.MapStorage{
		"key": nil,
	}

	m := &matchedStatQueue{
		statQueue: &utils.StatQueue{
			SQItems: []utils.SQItem{
				{
					EventID: evID,
				},
			},
			SQMetrics: map[string]utils.StatMetric{
				"key": statMetricMock("remExpired error"),
			},
		},
		profile: &utils.StatQueueProfile{
			QueueLength: 1,
		},
	}

	experr := "remExpired mock error"
	err := m.processEvent(context.Background(), tnt, evID, evNm)

	if err == nil || err.Error() != experr {
		t.Errorf("\nexpected: %q, \nreceived: %q", experr, err)
	}
}

func TestStatQueueProcessEventaddStatEvent(t *testing.T) {
	tnt, evID := "tenant", "eventID"
	filters := &engine.FilterS{}
	evNm := utils.MapStorage{
		"key": nil,
	}

	m := &matchedStatQueue{
		statQueue: &utils.StatQueue{
			SQItems: []utils.SQItem{
				{
					EventID: evID,
				},
			},
			SQMetrics: map[string]utils.StatMetric{
				utils.MetaTCD: &utils.StatTCD{Metric: &utils.Metric{}},
			},
		},
		profile: &utils.StatQueueProfile{
			QueueLength: 1,
			Metrics: []*utils.MetricWithFilters{
				{
					MetricID: utils.MetaTCD,
				},
			},
		},
	}

	m.cfg = config.NewDefaultCGRConfig()
	m.filters = filters
	experr := utils.ErrWrongPath
	err := m.processEvent(context.Background(), tnt, evID, evNm)

	if err == nil || err != experr {
		t.Errorf("\nexpected: %q, \nreceived: %q", experr, err)
	}
}

func TestStatQueueaddStatEventNoPass(t *testing.T) {
	sm, err := utils.NewStatMetric(utils.MetaTCD, 0, []string{"*string:~*req.Account:1001"})
	if err != nil {
		t.Fatal(err)
	}

	m := &matchedStatQueue{
		statQueue: &utils.StatQueue{
			SQMetrics: map[string]utils.StatMetric{
				utils.MetaTCD: sm,
			},
		},
		profile: &utils.StatQueueProfile{
			Metrics: []*utils.MetricWithFilters{
				{
					FilterIDs: []string{"*string:~*req.Account:1001"},
					MetricID:  utils.MetaTCD,
				},
			},
		},
	}

	tnt, evID := "cgrates.org", "eventID"
	cfg := config.NewDefaultCGRConfig()
	data, err := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	if err != nil {
		t.Fatal(err)
	}
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	cacheS := engine.NewCacheS(cfg, nil, nil, nil)
	dm.SetCache(cacheS)
	filters := engine.NewFilterS(cfg, nil, dm)
	evNm := utils.MapStorage{
		utils.MetaReq: utils.MapStorage{
			utils.MetaReq: nil,
		},
		utils.MetaOpts: nil,
		utils.MetaVars: utils.MapStorage{
			utils.OptsAttributesProcessRuns: 0,
		},
	}

	exp := &utils.StatQueue{
		SQMetrics: map[string]utils.StatMetric{
			utils.MetaTCD: sm,
		},
		SQItems: []utils.SQItem{
			{
				EventID: "eventID",
			},
		},
	}
	expProfile := &utils.StatQueueProfile{
		Metrics: []*utils.MetricWithFilters{
			{
				MetricID:  utils.MetaTCD,
				FilterIDs: []string{"*string:~*req.Account:1001"},
			},
		},
	}
	m.cfg = cfg
	m.filters = filters
	if err := m.addStatEvent(context.Background(), tnt, evID, evNm); err != nil {
		t.Fatalf("\nexpected: <%+v>, \nreceived: <%+v>", nil, err)
	}

	if !reflect.DeepEqual(m.statQueue, exp) {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", exp, m.statQueue)
	}
	if !reflect.DeepEqual(m.profile, expProfile) {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", expProfile, m.profile)
	}
}

func TestStatQAddStatEventFilterPassErr(t *testing.T) {

	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	cM := engine.NewConnManager(cfg)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, cM)
	cacheS := engine.NewCacheS(cfg, nil, nil, nil)
	cM.SetCache(cacheS)
	dm.SetCache(cacheS)

	fltrS := engine.NewFilterS(cfg, cM, dm)

	m := &matchedStatQueue{
		statQueue: &utils.StatQueue{
			SQMetrics: map[string]utils.StatMetric{
				utils.MetaASR: &utils.StatASR{
					Metric: &utils.Metric{
						Value: utils.NewDecimal(1, 0),
						Count: 1,
						Events: map[string]*utils.DecimalWithCompress{
							"cgrates.org:TestStatRemExpired_1": {Stat: utils.NewDecimalFromFloat64(1), CompressFactor: 1},
						},
					},
				},
			},
		},
		profile: &utils.StatQueueProfile{
			Metrics: []*utils.MetricWithFilters{
				{
					FilterIDs: []string{"*"},
					MetricID:  utils.MetaASR,
				},
			},
		},
	}
	asrMetric := m.statQueue.SQMetrics[utils.MetaASR].(*utils.StatASR)
	if asr := asrMetric.GetValue(); asr.Compare(utils.NewDecimalFromFloat64(100)) != 0 {
		t.Errorf("received ASR: %v", asr)
	}
	ev1 := &utils.CGREvent{Tenant: "cgrates.org", ID: "TestStatAddStatEvent_1"}

	m.cfg = cfg
	m.filters = fltrS
	expErr := `inline parse error for string: <*>`
	if err := m.addStatEvent(context.Background(), ev1.Tenant, ev1.ID, utils.MapStorage{utils.MetaOpts: ev1.Event}); err == nil || err.Error() != expErr {
		t.Errorf("Expected error %s received: %v", expErr, err)
	}

}

func TestStatQAddStatEventBlockerFromDynamicsErr(t *testing.T) {

	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	cM := engine.NewConnManager(cfg)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, cM)
	cacheS := engine.NewCacheS(cfg, nil, nil, nil)
	cM.SetCache(cacheS)
	dm.SetCache(cacheS)

	fltrS := engine.NewFilterS(cfg, cM, dm)

	m := &matchedStatQueue{
		statQueue: &utils.StatQueue{
			SQMetrics: map[string]utils.StatMetric{
				utils.MetaASR: &utils.StatASR{
					Metric: &utils.Metric{
						Value: utils.NewDecimal(1, 0),
						Count: 1,
						Events: map[string]*utils.DecimalWithCompress{
							"cgrates.org:TestStatRemExpired_1": {Stat: utils.NewDecimalFromFloat64(1), CompressFactor: 1},
						},
					},
				},
			},
		},
		profile: &utils.StatQueueProfile{
			Metrics: []*utils.MetricWithFilters{
				{
					MetricID: utils.MetaASR,
					Blockers: utils.DynamicBlockers{
						{
							FilterIDs: []string{"*stirng:~*req.Account:1001"},
							Blocker:   true,
						},
					},
				},
			},
		},
	}
	asrMetric := m.statQueue.SQMetrics[utils.MetaASR].(*utils.StatASR)
	if asr := asrMetric.GetValue(); asr.Compare(utils.NewDecimalFromFloat64(100)) != 0 {
		t.Errorf("received ASR: %v", asr)
	}
	ev1 := &utils.CGREvent{Tenant: "cgrates.org", ID: "TestStatAddStatEvent_1"}

	m.cfg = cfg
	m.filters = fltrS
	expErr := `Unsupported filter Type: *stirng`
	if err := m.addStatEvent(context.Background(), ev1.Tenant, ev1.ID, utils.MapStorage{utils.MetaOpts: ev1.Event}); err == nil || err.Error() != expErr {
		t.Errorf("Expected error %s received: %v", expErr, err)
	}

}

func TestStatQAddStatEventBlockNotLast(t *testing.T) {

	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	cM := engine.NewConnManager(cfg)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, cM)
	cacheS := engine.NewCacheS(cfg, nil, nil, nil)
	cM.SetCache(cacheS)
	dm.SetCache(cacheS)

	fltrS := engine.NewFilterS(cfg, cM, dm)

	m := &matchedStatQueue{
		statQueue: &utils.StatQueue{
			SQMetrics: map[string]utils.StatMetric{
				utils.MetaASR: &utils.StatASR{
					Metric: &utils.Metric{
						Value: utils.NewDecimal(1, 0),
						Count: 1,
						Events: map[string]*utils.DecimalWithCompress{
							"cgrates.org:TestStatRemExpired_1": {Stat: utils.NewDecimalFromFloat64(1), CompressFactor: 1},
						},
					},
				},
			},
		},
		profile: &utils.StatQueueProfile{
			Metrics: []*utils.MetricWithFilters{
				{
					MetricID: utils.MetaASR,
					Blockers: utils.DynamicBlockers{
						{
							Blocker: true,
						},
					},
				},
				{
					MetricID: utils.MetaTCD,
					Blockers: utils.DynamicBlockers{
						{
							Blocker: true,
						},
					},
				},
			},
		},
	}
	asrMetric := m.statQueue.SQMetrics[utils.MetaASR].(*utils.StatASR)
	if asr := asrMetric.GetValue(); asr.Compare(utils.NewDecimalFromFloat64(100)) != 0 {
		t.Errorf("received ASR: %v", asr)
	}
	ev1 := &utils.CGREvent{Tenant: "cgrates.org", ID: "TestStatAddStatEvent_1"}

	exp := &utils.StatQueueProfile{
		Metrics: []*utils.MetricWithFilters{
			{
				MetricID: utils.MetaASR,
				Blockers: utils.DynamicBlockers{
					{
						Blocker: true,
					},
				},
			},
			{
				MetricID: utils.MetaTCD,
				Blockers: utils.DynamicBlockers{
					{
						Blocker: true,
					},
				},
			},
		},
	}

	m.cfg = cfg
	m.filters = fltrS
	if err := m.addStatEvent(context.Background(), ev1.Tenant, ev1.ID, utils.MapStorage{utils.MetaOpts: ev1.Event}); err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(exp, m.profile) {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", exp, m.profile)
	}

}

type ccMock struct {
	calls map[string]func(ctx *context.Context, args any, reply any) error
}

func (ccM *ccMock) Call(ctx *context.Context, serviceMethod string, args any, reply any) (err error) {
	if call, has := ccM.calls[serviceMethod]; !has {
		return rpcclient.ErrUnsupporteServiceMethod
	} else {
		return call(ctx, args, reply)
	}
}

type statMetricMock string

func (statMetricMock) GetValue() *utils.Decimal {
	return nil
}

func (statMetricMock) GetStringValue(int) (val string) {
	return
}

func (statMetricMock) AddEvent(string, utils.DataProvider) error {
	return nil
}

func (statMetricMock) AddOneEvent(utils.DataProvider) error {
	return nil
}

func (sMM statMetricMock) RemEvent(string) error {
	if sMM == "remExpired error" {
		return fmt.Errorf("remExpired mock error")
	}
	return nil
}

func (sMM statMetricMock) GetMinItems() uint64 {
	return 0
}

func (sMM statMetricMock) Compress(uint64, string) []string {
	if sMM == "populate idMap" {
		return []string{"id1", "id2", "id3", "id4", "id5", "id6"}
	}
	return nil
}

func (sMM statMetricMock) GetFilterIDs() []string {
	if sMM == "pass error" {
		return []string{"filter1", "filter2"}
	}
	return nil
}
func (sMM statMetricMock) GetCompressFactor(map[string]uint64) map[string]uint64 {
	return nil
}
func (sMM statMetricMock) Clone() utils.StatMetric {
	return sMM
}
