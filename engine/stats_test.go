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
	"bytes"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

var (
	testStatsPrfs = []*StatQueueProfile{
		{
			Tenant:      "cgrates.org",
			ID:          "StatQueueProfile1",
			FilterIDs:   []string{"FLTR_STATS_1", "*ai:*now:2014-07-14T14:25:00Z"},
			QueueLength: 10,
			TTL:         10 * time.Second,
			Metrics: []*MetricWithFilters{
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
			Metrics: []*MetricWithFilters{
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
			Metrics: []*MetricWithFilters{
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
	testStatsQ = []*StatQueue{
		{
			Tenant: "cgrates.org",
			ID:     "StatQueueProfile1",
			sqPrfl: testStatsPrfs[0],
			SQMetrics: map[string]StatMetric{
				utils.MetaSum: NewStatSum(1, "~*req.Usage", nil),
			},
		},
		{
			Tenant: "cgrates.org",
			ID:     "StatQueueProfile2",
			sqPrfl: testStatsPrfs[1],
			SQMetrics: map[string]StatMetric{
				utils.MetaSum: NewStatSum(1, "~*req.Usage", nil),
			},
		},
		{
			Tenant: "cgrates.org",
			ID:     "StatQueueProfilePrefix",
			sqPrfl: testStatsPrfs[2],
			SQMetrics: map[string]StatMetric{
				utils.MetaSum: NewStatSum(1, "~*req.Usage", nil),
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

func prepareStatsData(t *testing.T, dm *DataManager) {
	if err := dm.SetFilter(context.Background(), &Filter{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "FLTR_STATS_1",
		Rules: []*FilterRule{
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
	if err := dm.SetFilter(context.Background(), &Filter{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "FLTR_STATS_2",
		Rules: []*FilterRule{
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
	if err := dm.SetFilter(context.Background(), &Filter{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "FLTR_STATS_3",
		Rules: []*FilterRule{
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
	for _, statQueue := range testStatsQ {
		statSum, err := NewStatMetric("*sum#~*req.Usage", uint64(statQueue.sqPrfl.MinItems), []string{})
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
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	fltrS := &FilterS{dm: dm, cfg: cfg}
	sSrv := &StatS{
		dm:               dm,
		fltrS:            fltrS,
		cfg:              cfg,
		storedStatQueues: make(utils.StringSet),
	}
	result := NewStatService(dm, cfg, fltrS, nil)
	if !reflect.DeepEqual(sSrv.dm, result.dm) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", sSrv.dm, result.dm)
	}
	if !reflect.DeepEqual(sSrv.fltrS, result.fltrS) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", sSrv.fltrS, result.fltrS)
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
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dmSTS := NewDataManager(data, config.CgrConfig().CacheCfg(), nil)

	cfg.StatSCfg().StoreInterval = 1
	cfg.StatSCfg().StringIndexedFields = nil
	cfg.StatSCfg().PrefixIndexedFields = nil
	statService := NewStatService(dmSTS, cfg,
		&FilterS{dm: dmSTS, cfg: cfg}, nil)
	prepareStatsData(t, dmSTS)
	msq, err := statService.matchingStatQueuesForEvent(context.TODO(), testStatsArgs[0].Tenant, nil,
		testStatsArgs[0].AsDataProvider(), false)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	msq.unlock()
	if !reflect.DeepEqual(testStatsQ[0].Tenant, msq[0].Tenant) {
		t.Errorf("Expecting: %+v, received: %+v", testStatsQ[0].Tenant, msq[0].Tenant)
	} else if !reflect.DeepEqual(testStatsQ[0].ID, msq[0].ID) {
		t.Errorf("Expecting: %+v, received: %+v", testStatsQ[0].ID, msq[0].ID)
	} else if !reflect.DeepEqual(testStatsQ[0].sqPrfl, msq[0].sqPrfl) {
		t.Errorf("Expecting: %+v, received: %+v", testStatsQ[0].sqPrfl, msq[0].sqPrfl)
	}
	msq, err = statService.matchingStatQueuesForEvent(context.TODO(), testStatsArgs[1].Tenant, nil,
		testStatsArgs[1].AsDataProvider(), false)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	msq.unlock()
	if !reflect.DeepEqual(testStatsQ[1].Tenant, msq[0].Tenant) {
		t.Errorf("Expecting: %+v, received: %+v", testStatsQ[1].Tenant, msq[0].Tenant)
	} else if !reflect.DeepEqual(testStatsQ[1].ID, msq[0].ID) {
		t.Errorf("Expecting: %+v, received: %+v", testStatsQ[1].ID, msq[0].ID)
	} else if !reflect.DeepEqual(testStatsQ[1].sqPrfl, msq[0].sqPrfl) {
		t.Errorf("Expecting: %+v, received: %+v", testStatsQ[1].sqPrfl, msq[0].sqPrfl)
	}
	msq, err = statService.matchingStatQueuesForEvent(context.TODO(), testStatsArgs[2].Tenant, nil,
		testStatsArgs[2].AsDataProvider(), false)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	msq.unlock()
	if !reflect.DeepEqual(testStatsQ[2].Tenant, msq[0].Tenant) {
		t.Errorf("Expecting: %+v, received: %+v", testStatsQ[2].Tenant, msq[0].Tenant)
	} else if !reflect.DeepEqual(testStatsQ[2].ID, msq[0].ID) {
		t.Errorf("Expecting: %+v, received: %+v", testStatsQ[2].ID, msq[0].ID)
	} else if !reflect.DeepEqual(testStatsQ[2].sqPrfl, msq[0].sqPrfl) {
		t.Errorf("Expecting: %+v, received: %+v", testStatsQ[2].sqPrfl, msq[0].sqPrfl)
	}
}

func TestStatQueuesProcessEvent(t *testing.T) {
	Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dmSTS := NewDataManager(data, config.CgrConfig().CacheCfg(), nil)

	cfg.StatSCfg().StoreInterval = 1
	cfg.StatSCfg().StringIndexedFields = nil
	cfg.StatSCfg().PrefixIndexedFields = nil
	statService := NewStatService(dmSTS, cfg,
		&FilterS{dm: dmSTS, cfg: cfg}, nil)

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
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dmSTS := NewDataManager(data, config.CgrConfig().CacheCfg(), nil)

	cfg.StatSCfg().StoreInterval = 1
	cfg.StatSCfg().StringIndexedFields = nil
	cfg.StatSCfg().PrefixIndexedFields = nil
	statService := NewStatService(dmSTS, cfg,
		&FilterS{dm: dmSTS, cfg: cfg}, nil)
	prepareStatsData(t, dmSTS)

	statService.cfg.StatSCfg().IndexedSelects = false
	msq, err := statService.matchingStatQueuesForEvent(context.TODO(), testStatsArgs[0].Tenant, nil,
		testStatsArgs[0].AsDataProvider(), false)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	msq.unlock()
	if !reflect.DeepEqual(testStatsQ[0].Tenant, msq[0].Tenant) {
		t.Errorf("Expecting: %+v, received: %+v", testStatsQ[0].Tenant, msq[0].Tenant)
	} else if !reflect.DeepEqual(testStatsQ[0].ID, msq[0].ID) {
		t.Errorf("Expecting: %+v, received: %+v", testStatsQ[0].ID, msq[0].ID)
	} else if !reflect.DeepEqual(testStatsQ[0].sqPrfl, msq[0].sqPrfl) {
		t.Errorf("Expecting: %+v, received: %+v", testStatsQ[0].sqPrfl, msq[0].sqPrfl)
	}
	msq, err = statService.matchingStatQueuesForEvent(context.TODO(), testStatsArgs[1].Tenant, nil,
		testStatsArgs[1].AsDataProvider(), false)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	msq.unlock()
	if !reflect.DeepEqual(testStatsQ[1].Tenant, msq[0].Tenant) {
		t.Errorf("Expecting: %+v, received: %+v", testStatsQ[1].Tenant, msq[0].Tenant)
	} else if !reflect.DeepEqual(testStatsQ[1].ID, msq[0].ID) {
		t.Errorf("Expecting: %+v, received: %+v", testStatsQ[1].ID, msq[0].ID)
	} else if !reflect.DeepEqual(testStatsQ[1].sqPrfl, msq[0].sqPrfl) {
		t.Errorf("Expecting: %+v, received: %+v", testStatsQ[1].sqPrfl, msq[0].sqPrfl)
	}
	msq, err = statService.matchingStatQueuesForEvent(context.TODO(), testStatsArgs[2].Tenant, nil,
		testStatsArgs[2].AsDataProvider(), false)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	msq.unlock()
	if !reflect.DeepEqual(testStatsQ[2].Tenant, msq[0].Tenant) {
		t.Errorf("Expecting: %+v, received: %+v", testStatsQ[2].Tenant, msq[0].Tenant)
	} else if !reflect.DeepEqual(testStatsQ[2].ID, msq[0].ID) {
		t.Errorf("Expecting: %+v, received: %+v", testStatsQ[2].ID, msq[0].ID)
	} else if !reflect.DeepEqual(testStatsQ[2].sqPrfl, msq[0].sqPrfl) {
		t.Errorf("Expecting: %+v, received: %+v", testStatsQ[2].sqPrfl, msq[0].sqPrfl)
	}
}

func TestStatQueuesV1ProcessEvent(t *testing.T) {
	Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dmSTS := NewDataManager(data, config.CgrConfig().CacheCfg(), nil)

	cfg.StatSCfg().StoreInterval = 1
	cfg.StatSCfg().StringIndexedFields = nil
	cfg.StatSCfg().PrefixIndexedFields = nil
	statService := NewStatService(dmSTS, cfg,
		&FilterS{dm: dmSTS, cfg: cfg}, nil)
	prepareStatsData(t, dmSTS)

	sqPrf := &StatQueueProfile{
		Tenant:      "cgrates.org",
		ID:          "StatQueueProfile3",
		FilterIDs:   []string{"FLTR_STATS_1", "*ai:~*req.AnswerTime:2014-07-14T14:25:00Z"},
		QueueLength: 10,
		TTL:         10 * time.Second,
		Metrics: []*MetricWithFilters{
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
	statSum, err := NewStatMetric("*sum#~*req.Usage", uint64(sqPrf.MinItems), []string{})
	if err != nil {
		t.Fatal(err)
	}
	sq := &StatQueue{Tenant: "cgrates.org", ID: "StatQueueProfile3", sqPrfl: sqPrf, SQMetrics: map[string]StatMetric{"*sum#~*req.Usage": statSum}}
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
	dm := NewDataManager(NewInternalDB(nil, nil, cfg.DataDbCfg().Items), cfg.CacheCfg(), nil)
	sqp := &StatQueueProfile{
		Tenant:      "cgrates.org",
		ID:          "THUP1",
		Stored:      true,
		QueueLength: 1,
		Metrics:     []*MetricWithFilters{{MetricID: utils.MetaTCC}},
	}
	sqm := NewTCC(0, utils.EmptyString, nil)
	if err := sqm.AddEvent("ev1", utils.MapStorage{utils.MetaOpts: utils.MapStorage{utils.MetaCost: 10}}); err != nil {
		t.Fatal(err)
	}
	sq := &StatQueue{
		Tenant:    sqp.Tenant,
		ID:        sqp.ID,
		SQItems:   []SQItem{{EventID: "ev1"}},
		SQMetrics: map[string]StatMetric{utils.MetaTCC: sqm, utils.MetaTCD: sqm},
	}
	sqm2 := NewTCC(0, utils.EmptyString, nil)

	expTh := &StatQueue{
		Tenant:    sqp.Tenant,
		ID:        sqp.ID,
		SQMetrics: map[string]StatMetric{utils.MetaTCC: sqm2},
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

	sqp = &StatQueueProfile{
		Tenant:      "cgrates.org",
		ID:          "THUP1",
		Stored:      true,
		QueueLength: 1,
		Metrics:     []*MetricWithFilters{{MetricID: utils.MetaTCC, FilterIDs: []string{"*string:~*req.Account:1001"}}},
	}

	if err := dm.SetStatQueueProfile(context.Background(), sqp, true); err != nil {
		t.Fatal(err)
	}

	sqm3 := NewTCC(0, utils.EmptyString, []string{"*string:~*req.Account:1001"})

	expTh = &StatQueue{
		Tenant:    sqp.Tenant,
		ID:        sqp.ID,
		SQItems:   []SQItem{{EventID: "ev1"}},
		SQMetrics: map[string]StatMetric{utils.MetaTCC: sqm3},
	}
	if th, err := dm.GetStatQueue(context.Background(), sqp.Tenant, sqp.ID, false, false, utils.NonTransactional); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(expTh, th) {
		t.Errorf("Expected: %s, received: %s", utils.ToJSON(expTh), utils.ToJSON(th))
	}

	if err := dm.SetStatQueue(context.Background(), sq); err != nil {
		t.Fatal(err)
	}

	sqm2 = NewTCC(5, utils.EmptyString, nil)

	expTh = &StatQueue{
		Tenant:    sqp.Tenant,
		ID:        sqp.ID,
		SQMetrics: map[string]StatMetric{utils.MetaTCC: sqm2},
	}
	delete(sq.SQMetrics, utils.MetaTCD)
	if err := dm.SetStatQueue(context.Background(), sq); err != nil {
		t.Fatal(err)
	}

	sqp = &StatQueueProfile{
		Tenant:      "cgrates.org",
		ID:          "THUP1",
		Stored:      true,
		QueueLength: 1,
		Metrics:     []*MetricWithFilters{{MetricID: utils.MetaTCC}},
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

	sqp = &StatQueueProfile{
		Tenant:      "cgrates.org",
		ID:          "THUP1",
		Stored:      true,
		QueueLength: 1,
		Metrics:     []*MetricWithFilters{{MetricID: utils.MetaTCC}},
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

	sqp = &StatQueueProfile{
		Tenant:      "cgrates.org",
		ID:          "THUP1",
		Stored:      true,
		QueueLength: 10,
		Metrics:     []*MetricWithFilters{{MetricID: utils.MetaTCC}},
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

	sqp = &StatQueueProfile{
		Tenant:      "cgrates.org",
		ID:          "THUP1",
		Stored:      false,
		QueueLength: 10,
		Metrics:     []*MetricWithFilters{{MetricID: utils.MetaTCC}},
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
	tmp := Cache
	defer func() { Cache = tmp }()
	Cache = NewCacheS(cfg, nil, nil, nil)
	db := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, config.CgrConfig().CacheCfg(), nil)
	cfg.StatSCfg().StoreInterval = 1
	cfg.StatSCfg().StringIndexedFields = nil
	cfg.StatSCfg().PrefixIndexedFields = nil
	rS := NewStatService(dm, cfg,
		&FilterS{dm: dm, cfg: cfg}, nil)

	prfs := make([]*StatQueueProfile, 0)
	ids := utils.StringSet{}
	for i := 0; i < 10; i++ {
		rPrf := &StatQueueProfile{
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
		dm.SetStatQueueProfile(context.Background(), rPrf, true)
		prfs = append(prfs, rPrf)
		ids.Add(rPrf.ID)
	}
	dm.RemoveStatQueue(context.Background(), "cgrates.org", "STS1")
	_, err := rS.matchingStatQueuesForEvent(context.Background(), "cgrates.org", ids.AsSlice(), utils.MapStorage{}, false)
	if err != utils.ErrNotFound {
		t.Errorf("Error: %+v", err)
	}
	for _, rPrf := range prfs {
		if rPrf.isLocked() {
			t.Fatalf("Expected profile to not be locked %q", rPrf.ID)
		}
		if rPrf.ID == "STS1" {
			continue
		}
		if r, err := dm.GetStatQueue(context.Background(), rPrf.Tenant, rPrf.ID, true, false, utils.NonTransactional); err != nil {
			t.Errorf("error %s for <%s>", err, rPrf.ID)
		} else if r.isLocked() {
			t.Fatalf("Expected StatQueue to not be locked %q", rPrf.ID)
		}
	}

}

func TestStatQueueMatchingStatQueuesForEventLocks2(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	tmp := Cache
	defer func() { Cache = tmp }()
	Cache = NewCacheS(cfg, nil, nil, nil)
	db := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, config.CgrConfig().CacheCfg(), nil)
	cfg.StatSCfg().StoreInterval = 1
	cfg.StatSCfg().StringIndexedFields = nil
	cfg.StatSCfg().PrefixIndexedFields = nil
	rS := NewStatService(dm, cfg,
		&FilterS{dm: dm, cfg: cfg}, nil)

	prfs := make([]*StatQueueProfile, 0)
	ids := utils.StringSet{}
	for i := 0; i < 10; i++ {
		rPrf := &StatQueueProfile{
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
		dm.SetStatQueueProfile(context.Background(), rPrf, true)
		prfs = append(prfs, rPrf)
		ids.Add(rPrf.ID)
	}
	rPrf := &StatQueueProfile{
		Tenant:      "cgrates.org",
		ID:          "STS20",
		FilterIDs:   []string{"FLTR_RES_201"},
		QueueLength: 1,
		Stored:      true,
		Weights: utils.DynamicWeights{
			{
				Weight: 20.00,
			},
		},
		ThresholdIDs: []string{utils.MetaNone},
	}
	err := db.SetStatQueueProfileDrv(context.Background(), rPrf)
	if err != nil {
		t.Fatal(err)
	}
	prfs = append(prfs, rPrf)
	ids.Add(rPrf.ID)
	_, err = rS.matchingStatQueuesForEvent(context.Background(), "cgrates.org", ids.AsSlice(), utils.MapStorage{}, false)
	expErr := utils.ErrPrefixNotFound(rPrf.FilterIDs[0])
	if err == nil || err.Error() != expErr.Error() {
		t.Errorf("Expected error: %s ,received: %+v", expErr, err)
	}
	for _, rPrf := range prfs {
		if rPrf.isLocked() {
			t.Fatalf("Expected profile to not be locked %q", rPrf.ID)
		}
		if rPrf.ID == "STS20" {
			continue
		}
		if r, err := dm.GetStatQueue(context.Background(), rPrf.Tenant, rPrf.ID, true, false, utils.NonTransactional); err != nil {
			t.Errorf("error %s for <%s>", err, rPrf.ID)
		} else if r.isLocked() {
			t.Fatalf("Expected StatQueue to not be locked %q", rPrf.ID)
		}
	}
}

/*
func TestStatQueueMatchingStatQueuesForEventLocksBlocker(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	tmp := Cache
	defer func() { Cache = tmp }()
	Cache = NewCacheS(cfg, nil, nil,nil)
	db := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, config.CgrConfig().CacheCfg(), nil)
	cfg.StatSCfg().StoreInterval = 1
	cfg.StatSCfg().StringIndexedFields = nil
	cfg.StatSCfg().PrefixIndexedFields = nil
	rS := NewStatService(dm, cfg,
		&FilterS{dm: dm, cfg: cfg}, nil)

	prfs := make([]*StatQueueProfile, 0)
	ids := utils.StringSet{}
	for i := 0; i < 10; i++ {
		rPrf := &StatQueueProfile{
			Tenant:      "cgrates.org",
			ID:          fmt.Sprintf("STS%d", i),
			QueueLength: 1,
			Stored:      true,
			Weights: utils.DynamicWeights{
				{
					Weight: float64(10 - i),
				},
			},
			Blockers:     utils.Blockers{{Blocker: i == 4}},
			ThresholdIDs: []string{utils.MetaNone},
		}
		dm.SetStatQueueProfile(context.Background(), rPrf, true)
		prfs = append(prfs, rPrf)
		ids.Add(rPrf.ID)
	}
	mres, err := rS.matchingStatQueuesForEvent(context.Background(), "cgrates.org", ids.AsSlice(), utils.MapStorage{}, false)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	defer mres.unlock()
	if len(mres) != 10 {
		t.Fatalf("Expected 10 StatQueues, but received %v", len(mres))
	}
	for _, rPrf := range prfs[5:] {
		if rPrf.isLocked() {
			t.Errorf("Expected profile to not be locked %q", rPrf.ID)
		}
		if r, err := dm.GetStatQueue(context.Background(), rPrf.Tenant, rPrf.ID, true, false, utils.NonTransactional); err != nil {
			t.Errorf("error %s for <%s>", err, rPrf.ID)
		} else if r.isLocked() {
			t.Fatalf("Expected StatQueue to not be locked %q", rPrf.ID)
		}
	}
	for _, rPrf := range prfs[:5] {
		if !rPrf.isLocked() {
			t.Errorf("Expected profile to be locked %q", rPrf.ID)
		}
		if r, err := dm.GetStatQueue(context.Background(), rPrf.Tenant, rPrf.ID, true, false, utils.NonTransactional); err != nil {
			t.Errorf("error %s for <%s>", err, rPrf.ID)
		} else if !r.isLocked() {
			t.Fatalf("Expected StatQueue to be locked %q", rPrf.ID)
		}
	}
}
*/

func TestStatQueueMatchingStatQueuesForEventLocks3(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	prfs := make([]*StatQueueProfile, 0)
	tmp := Cache
	defer func() { Cache = tmp }()
	Cache = NewCacheS(cfg, nil, nil, nil)
	db := &DataDBMock{
		GetStatQueueProfileDrvF: func(ctx *context.Context, tnt, id string) (*StatQueueProfile, error) {
			if id == "STS1" {
				return nil, utils.ErrNotImplemented
			}
			rPrf := &StatQueueProfile{
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
			Cache.Set(ctx, utils.CacheStatQueues, rPrf.TenantID(), &StatQueue{
				Tenant:    rPrf.Tenant,
				ID:        rPrf.ID,
				SQMetrics: make(map[string]StatMetric),
			}, nil, true, utils.NonTransactional)
			prfs = append(prfs, rPrf)
			return rPrf, nil
		},
	}
	dm := NewDataManager(db, config.CgrConfig().CacheCfg(), nil)
	cfg.StatSCfg().StoreInterval = 1
	cfg.StatSCfg().StringIndexedFields = nil
	cfg.StatSCfg().PrefixIndexedFields = nil
	rS := NewStatService(dm, cfg,
		&FilterS{dm: dm, cfg: cfg}, nil)

	ids := utils.StringSet{}
	for i := 0; i < 10; i++ {
		ids.Add(fmt.Sprintf("STS%d", i))
	}
	_, err := rS.matchingStatQueuesForEvent(context.Background(), "cgrates.org", ids.AsSlice(), utils.MapStorage{}, false)
	if err != utils.ErrNotImplemented {
		t.Fatalf("Error: %+v", err)
	}
	for _, rPrf := range prfs {
		if rPrf.isLocked() {
			t.Fatalf("Expected profile to not be locked %q", rPrf.ID)
		}
	}

}

func TestStatQueueMatchingStatQueuesForEventLocks4(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	tmp := Cache
	defer func() { Cache = tmp }()
	Cache = NewCacheS(cfg, nil, nil, nil)
	db := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, config.CgrConfig().CacheCfg(), nil)
	cfg.StatSCfg().StoreInterval = 1
	cfg.StatSCfg().StringIndexedFields = nil
	cfg.StatSCfg().PrefixIndexedFields = nil
	rS := NewStatService(dm, cfg,
		&FilterS{dm: dm, cfg: cfg}, nil)

	prfs := make([]*StatQueueProfile, 0)
	ids := utils.StringSet{}
	for i := 0; i < 10; i++ {
		rPrf := &StatQueueProfile{
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
		dm.SetStatQueueProfile(context.Background(), rPrf, true)
		prfs = append(prfs, rPrf)
		ids.Add(rPrf.ID)
	}
	ids.Add("STS20")
	mres, err := rS.matchingStatQueuesForEvent(context.Background(), "cgrates.org", ids.AsSlice(), utils.MapStorage{}, false)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	defer mres.unlock()
	for _, rPrf := range prfs {
		if !rPrf.isLocked() {
			t.Errorf("Expected profile to be locked %q", rPrf.ID)
		}
		if r, err := dm.GetStatQueue(context.Background(), rPrf.Tenant, rPrf.ID, true, false, utils.NonTransactional); err != nil {
			t.Errorf("error %s for <%s>", err, rPrf.ID)
		} else if !r.isLocked() {
			t.Fatalf("Expected StatQueue to be locked %q", rPrf.ID)
		}
	}

}

func TestStatQueueReload(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.StatSCfg().StoreInterval = 5 * time.Millisecond
	sS := &StatS{
		stopBackup:  make(chan struct{}),
		loopStopped: make(chan struct{}, 1),
		cfg:         cfg,
	}
	sS.loopStopped <- struct{}{}
	sS.Reload(context.Background())
	close(sS.stopBackup)
	select {
	case <-sS.loopStopped:
	case <-time.After(time.Second):
		t.Error("timed out waiting for loop to stop")
	}
}

func TestStatQueueStartLoop(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.StatSCfg().StoreInterval = -1
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	filterS := NewFilterS(cfg, nil, dm)
	sS := &StatS{
		dm:          dm,
		fltrS:       filterS,
		stopBackup:  make(chan struct{}),
		loopStopped: make(chan struct{}, 1),
		cfg:         cfg,
	}

	sS.StartLoop(context.Background())
	time.Sleep(10 * time.Millisecond)

	if len(sS.loopStopped) != 1 {
		t.Errorf("expected loopStopped field to have only one element, received: <%+v>", len(sS.loopStopped))
	}
}

func TestStatQueueShutdown(t *testing.T) {
	tmpLogger := utils.Logger
	defer func() {
		utils.Logger = tmpLogger
	}()
	var buf bytes.Buffer
	utils.Logger = utils.NewStdLoggerWithWriter(&buf, "", 6)

	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	sS := NewStatService(dm, cfg, nil, nil)

	expLog1 := `[INFO] <StatS> service shutdown initialized`
	expLog2 := `[INFO] <StatS> service shutdown complete`
	sS.Shutdown(context.Background())

	if rcvLog := buf.String(); !strings.Contains(rcvLog, expLog1) ||
		!strings.Contains(rcvLog, expLog2) {
		t.Errorf("expected logs <%+v> and <%+v> \n to be included in <%+v>",
			expLog1, expLog2, rcvLog)
	}
}

func TestStatQueueStoreStatsOK(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()

	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	sS := NewStatService(dm, cfg, nil, nil)

	exp := &StatQueue{
		dirty:     utils.BoolPointer(true),
		Tenant:    "cgrates.org",
		ID:        "SQ1",
		SQMetrics: make(map[string]StatMetric),
	}
	Cache.SetWithoutReplicate(utils.CacheStatQueues, "cgrates.org:SQ1", exp, nil, true,
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

	Cache.Remove(context.Background(), utils.CacheStatQueues, "cgrates.org:SQ1", true, utils.NonTransactional)
}

func TestStatQueueStoreStatsStoreSQErr(t *testing.T) {
	tmp := Cache
	tmpLogger := utils.Logger
	defer func() {
		Cache = tmp
		utils.Logger = tmpLogger
	}()

	var buf bytes.Buffer
	utils.Logger = utils.NewStdLoggerWithWriter(&buf, "", 4)

	cfg := config.NewDefaultCGRConfig()
	sS := NewStatService(nil, cfg, nil, nil)

	value := &StatQueue{
		dirty:     utils.BoolPointer(true),
		Tenant:    "cgrates.org",
		ID:        "SQ1",
		SQMetrics: make(map[string]StatMetric),
	}

	Cache.SetWithoutReplicate(utils.CacheStatQueues, "SQ1", value, nil, true,
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

	Cache.Remove(context.Background(), utils.CacheStatQueues, "SQ1", true, utils.NonTransactional)
}

func TestStatQueueStoreStatsCacheGetErr(t *testing.T) {
	tmp := Cache
	tmpLogger := utils.Logger
	defer func() {
		utils.Logger = tmpLogger
		Cache = tmp
	}()
	var buf bytes.Buffer
	utils.Logger = utils.NewStdLoggerWithWriter(&buf, "", 4)

	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	sS := NewStatService(dm, cfg, nil, nil)

	value := &StatQueue{
		dirty:     utils.BoolPointer(true),
		Tenant:    "cgrates.org",
		ID:        "SQ1",
		SQMetrics: make(map[string]StatMetric),
	}

	Cache.SetWithoutReplicate(utils.CacheStatQueues, "SQ2", value, nil, true,
		utils.NonTransactional)
	sS.storedStatQueues.Add("SQ1")
	expLog := `[WARNING] <StatS> failed retrieving from cache stat queue with ID: SQ1`
	sS.storeStats(context.Background())

	if rcvLog := buf.String(); !strings.Contains(rcvLog, expLog) {
		t.Errorf("expected <%+v> \nto be included in: <%+v>", expLog, rcvLog)
	}

	Cache.Remove(context.Background(), utils.CacheStatQueues, "SQ2", true, utils.NonTransactional)
}

func TestStatQueueStoreStatQueueCacheSetErr(t *testing.T) {
	tmpLogger := utils.Logger
	defer func() {
		utils.Logger = tmpLogger
	}()
	var buf bytes.Buffer
	utils.Logger = utils.NewStdLoggerWithWriter(&buf, "", 4)

	tmp := Cache
	tmpC := config.CgrConfig()
	tmpCM := connMgr
	defer func() {
		Cache = tmp
		config.SetCgrConfig(tmpC)
		connMgr = tmpCM
	}()

	cfg := config.NewDefaultCGRConfig()
	cfg.CacheCfg().ReplicationConns = []string{"test"}
	cfg.CacheCfg().Partitions[utils.CacheStatQueues].Replicate = true
	cfg.RPCConns()["test"] = &config.RPCConn{Conns: []*config.RemoteHost{{}}}
	config.SetCgrConfig(cfg)
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	connMgr = NewConnManager(cfg)
	Cache = NewCacheS(cfg, dm, nil, nil)
	filterS := NewFilterS(cfg, nil, dm)
	sS := NewStatService(dm, cfg, filterS, connMgr)

	sq := &StatQueue{
		Tenant:    "cgrates.org",
		ID:        "SQ1",
		SQMetrics: make(map[string]StatMetric),
		dirty:     utils.BoolPointer(true),
	}
	Cache.SetWithoutReplicate(utils.CacheStatQueues, sq.TenantID(), sq, nil, true, utils.NonTransactional)
	expLog := `[WARNING] <StatS> failed caching StatQueue with ID: cgrates.org:SQ1, error: DISCONNECTED`
	if err := sS.StoreStatQueue(context.Background(), sq); err == nil ||
		err.Error() != utils.ErrDisconnected.Error() {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrDisconnected, err)
	} else if rcv := buf.String(); !strings.Contains(rcv, expLog) {
		t.Errorf("expected log <%+v> to be included in <%+v>", expLog, rcv)
	}
}

func TestStatQueueStoreThresholdNilDirtyField(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	sS := NewStatService(dm, cfg, nil, nil)

	sq := &StatQueue{
		Tenant:    "cgrates.org",
		ID:        "SQ1",
		SQMetrics: make(map[string]StatMetric),
	}

	if err := sS.StoreStatQueue(context.Background(), sq); err != nil {
		t.Error(err)
	}
}

func TestStatQueueProcessEventOK(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	filterS := NewFilterS(cfg, nil, dm)
	sS := NewStatService(dm, cfg, filterS, nil)

	sqPrf := &StatQueueProfile{
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
		Metrics: []*MetricWithFilters{
			{
				MetricID: utils.MetaTCD,
			},
		},
	}
	stat, err := NewStatMetric("*tcd", uint64(sqPrf.MinItems), []string{})
	if err != nil {
		t.Fatal(err)
	}
	sq := &StatQueue{
		sqPrfl: sqPrf,
		Tenant: "cgrates.org",
		ID:     "SQ1",
		SQItems: []SQItem{
			{
				EventID: "SqProcessEvent",
			},
		},
		SQMetrics: map[string]StatMetric{
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
	Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	filterS := NewFilterS(cfg, nil, dm)
	sS := NewStatService(dm, cfg, filterS, nil)

	sqPrf := &StatQueueProfile{
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
	sq := &StatQueue{
		sqPrfl: sqPrf,
		Tenant: "cgrates.org",
		ID:     "SQ1",
		SQItems: []SQItem{
			{
				EventID: "SqProcessEvent",
			},
		},
		SQMetrics: make(map[string]StatMetric),
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
	tmp := Cache
	tmpC := config.CgrConfig()
	defer func() {
		Cache = tmp
		config.SetCgrConfig(tmpC)
	}()

	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	Cache = NewCacheS(cfg, dm, nil, nil)
	filterS := NewFilterS(cfg, nil, dm)
	sS := NewStatService(dm, cfg, filterS, nil)

	sqPrf := &StatQueueProfile{
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
		Metrics: []*MetricWithFilters{
			{
				MetricID: utils.MetaTCD,
			},
		},
	}
	sq := &StatQueue{
		sqPrfl: sqPrf,
		Tenant: "cgrates.org",
		ID:     "SQ1",
		SQItems: []SQItem{
			{
				EventID: "SqProcessEvent",
			},
		},
		SQMetrics: map[string]StatMetric{
			utils.MetaTCD: &StatTCD{},
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
	tmp := Cache
	tmpC := config.CgrConfig()
	defer func() {
		Cache = tmp
		config.SetCgrConfig(tmpC)
	}()

	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	Cache = NewCacheS(cfg, dm, nil, nil)
	filterS := NewFilterS(cfg, nil, dm)
	sS := NewStatService(dm, cfg, filterS, nil)

	sqPrf := &StatQueueProfile{
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
		Metrics: []*MetricWithFilters{
			{
				MetricID: utils.MetaTCD,
			},
		},
	}
	sq1 := &StatQueue{
		sqPrfl: sqPrf,
		Tenant: "cgrates.org",
		ID:     "SQ1",
		SQItems: []SQItem{
			{
				EventID: "SqProcessEvent",
			},
		},
		SQMetrics: map[string]StatMetric{
			utils.MetaTCD: &StatTCD{},
		},
	}

	sq2 := &StatQueue{
		sqPrfl:    nil,
		Tenant:    "testTenant",
		ID:        "SQ2",
		SQMetrics: make(map[string]StatMetric),
	}
	sq3 := &StatQueue{
		sqPrfl:    nil,
		Tenant:    "cgrates.org",
		ID:        "SQ3",
		SQMetrics: make(map[string]StatMetric),
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
	tmp := Cache
	tmpC := config.CgrConfig()
	defer func() {
		Cache = tmp
		config.SetCgrConfig(tmpC)
	}()

	cfg := config.NewDefaultCGRConfig()
	data := &DataDBMock{}
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	Cache = NewCacheS(cfg, dm, nil, nil)
	filterS := NewFilterS(cfg, nil, dm)
	sS := NewStatService(dm, cfg, filterS, nil)

	var qIDs []string
	if err := sS.V1GetQueueIDs(context.Background(), &utils.TenantWithAPIOpts{}, &qIDs); err == nil ||
		err.Error() != utils.ErrNotImplemented.Error() {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotImplemented, err)
	}
}

func TestStatQueueV1GetStatQueueOK(t *testing.T) {
	tmp := Cache
	tmpC := config.CgrConfig()
	defer func() {
		Cache = tmp
		config.SetCgrConfig(tmpC)
	}()

	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	Cache = NewCacheS(cfg, dm, nil, nil)
	filterS := NewFilterS(cfg, nil, dm)
	sS := NewStatService(dm, cfg, filterS, nil)

	sqPrf := &StatQueueProfile{
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
		Metrics: []*MetricWithFilters{
			{
				MetricID: utils.MetaTCD,
			},
		},
	}
	sq := &StatQueue{
		sqPrfl: sqPrf,
		Tenant: "cgrates.org",
		ID:     "SQ1",
		SQItems: []SQItem{
			{
				EventID: "SqProcessEvent",
			},
		},
		SQMetrics: map[string]StatMetric{
			utils.MetaTCD: &StatTCD{},
		},
	}

	if err := dm.SetStatQueueProfile(context.Background(), sqPrf, true); err != nil {
		t.Error(err)
	}
	if err := dm.SetStatQueue(context.Background(), sq); err != nil {
		t.Error(err)
	}

	var reply StatQueue
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
	tmp := Cache
	tmpC := config.CgrConfig()
	defer func() {
		Cache = tmp
		config.SetCgrConfig(tmpC)
	}()

	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	Cache = NewCacheS(cfg, dm, nil, nil)
	filterS := NewFilterS(cfg, nil, dm)
	sS := NewStatService(dm, cfg, filterS, nil)

	var reply StatQueue
	if err := sS.V1GetStatQueue(context.Background(), &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			ID: "SQ1",
		},
	}, &reply); err == nil || err != utils.ErrNotFound {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}
}

func TestStatQueueV1GetStatQueueMissingArgs(t *testing.T) {
	tmp := Cache
	tmpC := config.CgrConfig()
	defer func() {
		Cache = tmp
		config.SetCgrConfig(tmpC)
	}()

	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	Cache = NewCacheS(cfg, dm, nil, nil)
	filterS := NewFilterS(cfg, nil, dm)
	sS := NewStatService(dm, cfg, filterS, nil)

	sqPrf := &StatQueueProfile{
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
		Metrics: []*MetricWithFilters{
			{
				MetricID: utils.MetaTCD,
			},
		},
	}
	sq := &StatQueue{
		sqPrfl: sqPrf,
		Tenant: "cgrates.org",
		ID:     "SQ1",
		SQItems: []SQItem{
			{
				EventID: "SqProcessEvent",
			},
		},
		SQMetrics: map[string]StatMetric{
			utils.MetaTCD: &StatTCD{},
		},
	}

	if err := dm.SetStatQueueProfile(context.Background(), sqPrf, true); err != nil {
		t.Error(err)
	}
	if err := dm.SetStatQueue(context.Background(), sq); err != nil {
		t.Error(err)
	}

	experr := `MANDATORY_IE_MISSING: [ID]`
	var reply StatQueue
	if err := sS.V1GetStatQueue(context.Background(), &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{},
	}, &reply); err == nil || err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}
}

func TestStatQueueV1GetStatQueuesForEventOK(t *testing.T) {
	tmp := Cache
	tmpC := config.CgrConfig()
	defer func() {
		Cache = tmp
		config.SetCgrConfig(tmpC)
	}()

	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	Cache = NewCacheS(cfg, dm, nil, nil)
	filterS := NewFilterS(cfg, nil, dm)
	sS := NewStatService(dm, cfg, filterS, nil)

	sqPrf1 := &StatQueueProfile{
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
		Metrics: []*MetricWithFilters{
			{
				MetricID: utils.MetaTCD,
			},
		},
	}

	if err := dm.SetStatQueueProfile(context.Background(), sqPrf1, true); err != nil {
		t.Error(err)
	}

	sqPrf2 := &StatQueueProfile{
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
		Metrics: []*MetricWithFilters{
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
	tmp := Cache
	tmpC := config.CgrConfig()
	defer func() {
		Cache = tmp
		config.SetCgrConfig(tmpC)
	}()

	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	Cache = NewCacheS(cfg, dm, nil, nil)
	filterS := NewFilterS(cfg, nil, dm)
	sS := NewStatService(dm, cfg, filterS, nil)

	sqPrf := &StatQueueProfile{
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
		Metrics: []*MetricWithFilters{
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
	tmp := Cache
	tmpC := config.CgrConfig()
	defer func() {
		Cache = tmp
		config.SetCgrConfig(tmpC)
	}()

	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	Cache = NewCacheS(cfg, dm, nil, nil)
	filterS := NewFilterS(cfg, nil, dm)
	sS := NewStatService(dm, cfg, filterS, nil)

	sqPrf := &StatQueueProfile{
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
		Metrics: []*MetricWithFilters{
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
	tmp := Cache
	tmpC := config.CgrConfig()
	defer func() {
		Cache = tmp
		config.SetCgrConfig(tmpC)
	}()

	cfg := config.NewDefaultCGRConfig()
	cfg.StatSCfg().StoreInterval = 1
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	Cache = NewCacheS(cfg, dm, nil, nil)
	filterS := NewFilterS(cfg, nil, dm)
	sS := NewStatService(dm, cfg, filterS, nil)

	sqPrf := &StatQueueProfile{
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
		Metrics: []*MetricWithFilters{
			{
				MetricID: utils.MetaTCD,
			},
		},
	}
	sq := &StatQueue{
		sqPrfl: sqPrf,
		dirty:  utils.BoolPointer(false),
		Tenant: "cgrates.org",
		ID:     "SQ1",
		SQItems: []SQItem{
			{
				EventID: "SqProcessEvent",
			},
		},
		SQMetrics: map[string]StatMetric{
			utils.MetaTCD: &StatTCD{
				Metric: &Metric{
					Value:  utils.NewDecimal(int64(time.Hour), 0),
					Count:  1,
					Events: make(map[string]*DecimalWithCompress),
				},
			},
		},
	}

	if err := dm.SetStatQueue(context.Background(), sq); err != nil {
		t.Error(err)
	}

	expStored := utils.StringSet{
		"cgrates.org:SQ1": {},
	}
	expSq := &StatQueue{
		sqPrfl:  sqPrf,
		dirty:   utils.BoolPointer(true),
		Tenant:  "cgrates.org",
		ID:      "SQ1",
		SQItems: []SQItem{},
		SQMetrics: map[string]StatMetric{
			utils.MetaTCD: &StatTCD{
				Metric: &Metric{
					Value:  utils.NewDecimal(0, 0),
					Events: make(map[string]*DecimalWithCompress),
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
	} else if !reflect.DeepEqual(sq, expSq) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ToJSON(expSq), utils.ToJSON(sq))
	} else if !reflect.DeepEqual(sS.storedStatQueues, expStored) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", expStored, sS.storedStatQueues)
	}
}

func TestStatQueueV1ResetStatQueueNotFoundErr(t *testing.T) {
	tmp := Cache
	tmpC := config.CgrConfig()
	defer func() {
		Cache = tmp
		config.SetCgrConfig(tmpC)
	}()

	cfg := config.NewDefaultCGRConfig()
	cfg.StatSCfg().StoreInterval = 1
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	Cache = NewCacheS(cfg, dm, nil, nil)
	filterS := NewFilterS(cfg, nil, dm)
	sS := NewStatService(dm, cfg, filterS, nil)

	sqPrf := &StatQueueProfile{
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
		Metrics: []*MetricWithFilters{
			{
				MetricID: utils.MetaTCD,
			},
		},
	}
	sq := &StatQueue{
		sqPrfl: sqPrf,
		dirty:  utils.BoolPointer(false),
		Tenant: "cgrates.org",
		ID:     "SQ1",
		SQItems: []SQItem{
			{
				EventID: "SqProcessEvent",
			},
		},
		SQMetrics: map[string]StatMetric{
			utils.MetaTCD: &StatTCD{
				Metric: &Metric{
					Value:  utils.NewDecimal(int64(time.Hour), 0),
					Count:  1,
					Events: make(map[string]*DecimalWithCompress),
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
	tmp := Cache
	tmpC := config.CgrConfig()
	defer func() {
		Cache = tmp
		config.SetCgrConfig(tmpC)
	}()

	cfg := config.NewDefaultCGRConfig()
	cfg.StatSCfg().StoreInterval = 1
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	Cache = NewCacheS(cfg, dm, nil, nil)
	filterS := NewFilterS(cfg, nil, dm)
	sS := NewStatService(dm, cfg, filterS, nil)

	sqPrf := &StatQueueProfile{
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
		Metrics: []*MetricWithFilters{
			{
				MetricID: utils.MetaTCD,
			},
		},
	}
	sq := &StatQueue{
		sqPrfl: sqPrf,
		dirty:  utils.BoolPointer(false),
		Tenant: "cgrates.org",
		ID:     "SQ1",
		SQItems: []SQItem{
			{
				EventID: "SqProcessEvent",
			},
		},
		SQMetrics: map[string]StatMetric{
			utils.MetaTCD: &StatTCD{
				Metric: &Metric{
					Value:  utils.NewDecimal(int64(time.Hour), 0),
					Count:  1,
					Events: make(map[string]*DecimalWithCompress),
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
	tmp := Cache
	tmpC := config.CgrConfig()
	defer func() {
		Cache = tmp
		config.SetCgrConfig(tmpC)
	}()

	cfg := config.NewDefaultCGRConfig()
	cfg.StatSCfg().StoreInterval = 1
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	Cache = NewCacheS(cfg, dm, nil, nil)
	filterS := NewFilterS(cfg, nil, dm)
	sS := NewStatService(dm, cfg, filterS, nil)

	sqPrf := &StatQueueProfile{
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
		Metrics: []*MetricWithFilters{
			{
				MetricID: "testMetricType",
			},
		},
	}
	sq := &StatQueue{
		sqPrfl: sqPrf,
		dirty:  utils.BoolPointer(false),
		Tenant: "cgrates.org",
		ID:     "SQ1",
		SQItems: []SQItem{
			{
				EventID: "SqProcessEvent",
			},
		},
		SQMetrics: map[string]StatMetric{
			"testMetricType": &StatTCD{
				Metric: &Metric{
					Value:  utils.NewDecimal(int64(time.Hour), 0),
					Count:  1,
					Events: make(map[string]*DecimalWithCompress),
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
	tmp := Cache
	tmpC := config.CgrConfig()
	tmpCM := connMgr
	defer func() {
		Cache = tmp
		config.SetCgrConfig(tmpC)
		connMgr = tmpCM
	}()

	cfg := config.NewDefaultCGRConfig()
	cfg.StatSCfg().ThresholdSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds)}
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	Cache = NewCacheS(cfg, dm, nil, nil)

	filterS := NewFilterS(cfg, nil, dm)
	sS := NewStatService(dm, cfg, filterS, nil)

	sqPrf := &StatQueueProfile{
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
		Metrics: []*MetricWithFilters{
			{
				MetricID: "testMetricType",
			},
		},
	}
	sq := &StatQueue{
		sqPrfl: sqPrf,
		dirty:  utils.BoolPointer(false),
		Tenant: "cgrates.org",
		ID:     "SQ1",
		SQItems: []SQItem{
			{
				EventID: "SqProcessEvent",
			},
		},
		SQMetrics: map[string]StatMetric{
			"testMetricType": &StatTCD{
				Metric: &Metric{
					Value:  utils.NewDecimal(int64(time.Hour), 0),
					Count:  1,
					Events: make(map[string]*DecimalWithCompress),
				},
			},
		},
	}

	if err := dm.SetStatQueue(context.Background(), sq); err != nil {
		t.Error(err)
	}

	sQs := StatQueues{
		sq,
	}

	if err := sS.processThresholds(context.Background(), sQs, nil); err != nil {
		t.Error(err)
	}
}

func TestStatQueueProcessThresholdsOK(t *testing.T) {
	tmp := Cache
	tmpC := config.CgrConfig()
	tmpCM := connMgr
	defer func() {
		Cache = tmp
		config.SetCgrConfig(tmpC)
		connMgr = tmpCM
	}()

	cfg := config.NewDefaultCGRConfig()
	cfg.StatSCfg().ThresholdSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds)}
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	Cache = NewCacheS(cfg, dm, nil, nil)

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
	connMgr = NewConnManager(cfg)
	connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds), utils.ThresholdSv1, rpcInternal)

	filterS := NewFilterS(cfg, nil, dm)
	sS := NewStatService(dm, cfg, filterS, connMgr)

	sqPrf := &StatQueueProfile{
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
		Metrics: []*MetricWithFilters{
			{
				MetricID: "testMetricType",
			},
		},
	}
	sq := &StatQueue{
		sqPrfl: sqPrf,
		dirty:  utils.BoolPointer(false),
		Tenant: "cgrates.org",
		ID:     "SQ1",
		SQItems: []SQItem{
			{
				EventID: "SqProcessEvent",
			},
		},
		SQMetrics: map[string]StatMetric{
			"testMetricType": &StatTCD{
				Metric: &Metric{
					Value: utils.NewDecimal(int64(time.Hour), 0),
					Count: 1,
					Events: map[string]*DecimalWithCompress{
						"SqProcessEvent": {},
					},
				},
			},
		},
	}

	if err := dm.SetStatQueue(context.Background(), sq); err != nil {
		t.Error(err)
	}

	if err := sS.processThresholds(context.Background(), StatQueues{sq}, nil); err != nil {
		t.Error(err)
	}
}

func TestStatQueueProcessThresholdsErrPartExec(t *testing.T) {
	tmp := Cache
	tmpC := config.CgrConfig()
	tmpCM := connMgr
	tmpLogger := utils.Logger
	defer func() {
		Cache = tmp
		config.SetCgrConfig(tmpC)
		connMgr = tmpCM
		utils.Logger = tmpLogger
	}()
	var buf bytes.Buffer
	utils.Logger = utils.NewStdLoggerWithWriter(&buf, "", 4)

	cfg := config.NewDefaultCGRConfig()
	cfg.StatSCfg().ThresholdSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds)}
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	Cache = NewCacheS(cfg, dm, nil, nil)

	ccM := &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.ThresholdSv1ProcessEvent: func(ctx *context.Context, args, reply any) error {
				return utils.ErrExists
			},
		},
	}
	rpcInternal := make(chan birpc.ClientConnector, 1)
	rpcInternal <- ccM
	connMgr = NewConnManager(cfg)
	connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds), utils.ThresholdSv1, rpcInternal)

	filterS := NewFilterS(cfg, nil, dm)
	sS := NewStatService(dm, cfg, filterS, connMgr)

	sqPrf := &StatQueueProfile{
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
		Metrics: []*MetricWithFilters{
			{
				MetricID: "testMetricType",
			},
		},
	}
	sq := &StatQueue{
		sqPrfl: sqPrf,
		dirty:  utils.BoolPointer(false),
		Tenant: "cgrates.org",
		ID:     "SQ1",
		SQItems: []SQItem{
			{
				EventID: "SqProcessEvent",
			},
		},
		SQMetrics: map[string]StatMetric{
			"testMetricType": &StatTCD{
				Metric: &Metric{
					Value:  utils.NewDecimal(int64(time.Hour), 0),
					Count:  1,
					Events: make(map[string]*DecimalWithCompress),
				},
			},
		},
	}

	if err := dm.SetStatQueue(context.Background(), sq); err != nil {
		t.Error(err)
	}

	sQs := StatQueues{
		sq,
	}

	expLog := `[WARNING] <StatS> error: EXISTS`
	if err := sS.processThresholds(context.Background(), sQs, nil); err == nil ||
		err != utils.ErrPartiallyExecuted {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrPartiallyExecuted, err)
	} else if rcvLog := buf.String(); !strings.Contains(rcvLog, expLog) {
		t.Errorf("expected log <%+v> to be included in <%+v>", expLog, rcvLog)
	}
}

func TestStatQueueV1GetQueueFloatMetricsOK(t *testing.T) {
	tmp := Cache
	tmpC := config.CgrConfig()
	defer func() {
		Cache = tmp
		config.SetCgrConfig(tmpC)
	}()

	cfg := config.NewDefaultCGRConfig()
	cfg.StatSCfg().StoreInterval = 1
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	Cache = NewCacheS(cfg, dm, nil, nil)
	filterS := NewFilterS(cfg, nil, dm)
	sS := NewStatService(dm, cfg, filterS, nil)

	sqPrf := &StatQueueProfile{
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
		Metrics: []*MetricWithFilters{
			{
				MetricID: utils.MetaTCD,
			},
		},
	}
	sq := &StatQueue{
		sqPrfl: sqPrf,
		dirty:  utils.BoolPointer(false),
		Tenant: "cgrates.org",
		ID:     "SQ1",
		SQItems: []SQItem{
			{
				EventID: "SqProcessEvent",
			},
		},
		SQMetrics: map[string]StatMetric{
			utils.MetaTCD: &StatTCD{
				Metric: &Metric{
					Value: utils.NewDecimal(int64(time.Hour), 0),
					Count: 1,
					Events: map[string]*DecimalWithCompress{
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
	tmp := Cache
	tmpC := config.CgrConfig()
	defer func() {
		Cache = tmp
		config.SetCgrConfig(tmpC)
	}()

	cfg := config.NewDefaultCGRConfig()
	cfg.StatSCfg().StoreInterval = 1
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	Cache = NewCacheS(cfg, dm, nil, nil)
	filterS := NewFilterS(cfg, nil, dm)
	sS := NewStatService(dm, cfg, filterS, nil)

	sqPrf := &StatQueueProfile{
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
		Metrics: []*MetricWithFilters{
			{
				MetricID: utils.MetaTCD,
			},
		},
	}
	sq := &StatQueue{
		sqPrfl: sqPrf,
		dirty:  utils.BoolPointer(false),
		Tenant: "cgrates.org",
		ID:     "SQ1",
		SQItems: []SQItem{
			{
				EventID: "SqProcessEvent",
			},
		},
		SQMetrics: map[string]StatMetric{
			utils.MetaTCD: &StatTCD{
				Metric: &Metric{
					Value: utils.NewDecimal(int64(time.Hour), 0),
					Count: 1,
					Events: map[string]*DecimalWithCompress{
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
	tmp := Cache
	tmpC := config.CgrConfig()
	defer func() {
		Cache = tmp
		config.SetCgrConfig(tmpC)
	}()

	cfg := config.NewDefaultCGRConfig()
	cfg.StatSCfg().StoreInterval = 1
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	Cache = NewCacheS(cfg, dm, nil, nil)
	filterS := NewFilterS(cfg, nil, dm)
	sS := NewStatService(dm, cfg, filterS, nil)

	sqPrf := &StatQueueProfile{
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
		Metrics: []*MetricWithFilters{
			{
				MetricID: utils.MetaTCD,
			},
		},
	}
	sq := &StatQueue{
		sqPrfl: sqPrf,
		dirty:  utils.BoolPointer(false),
		Tenant: "cgrates.org",
		ID:     "SQ1",
		SQItems: []SQItem{
			{
				EventID: "SqProcessEvent",
			},
		},
		SQMetrics: map[string]StatMetric{
			utils.MetaTCD: &StatTCD{
				Metric: &Metric{
					Value:  utils.NewDecimal(int64(time.Hour), 0),
					Count:  1,
					Events: make(map[string]*DecimalWithCompress),
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
	tmp := Cache
	tmpC := config.CgrConfig()
	defer func() {
		Cache = tmp
		config.SetCgrConfig(tmpC)
	}()

	cfg := config.NewDefaultCGRConfig()
	cfg.StatSCfg().StoreInterval = 1
	Cache = NewCacheS(cfg, nil, nil, nil)
	filterS := NewFilterS(cfg, nil, nil)
	sS := NewStatService(nil, cfg, filterS, nil)

	experr := `SERVER_ERROR: NO_DATABASE_CONNECTION`
	reply := map[string]float64{}
	if err := sS.V1GetQueueFloatMetrics(context.Background(), &utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{ID: "SQ1"}}, &reply); err == nil || err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}
}

func TestStatQueueV1GetQueueStringMetricsOK(t *testing.T) {
	tmp := Cache
	tmpC := config.CgrConfig()
	defer func() {
		Cache = tmp
		config.SetCgrConfig(tmpC)
	}()

	cfg := config.NewDefaultCGRConfig()
	cfg.StatSCfg().StoreInterval = 1
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	Cache = NewCacheS(cfg, dm, nil, nil)
	filterS := NewFilterS(cfg, nil, dm)
	sS := NewStatService(dm, cfg, filterS, nil)

	sqPrf := &StatQueueProfile{
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
		Metrics: []*MetricWithFilters{
			{
				MetricID: utils.MetaTCD,
			},
		},
	}
	sq := &StatQueue{
		sqPrfl: sqPrf,
		dirty:  utils.BoolPointer(false),
		Tenant: "cgrates.org",
		ID:     "SQ1",
		SQItems: []SQItem{
			{
				EventID: "SqProcessEvent",
			},
		},
		SQMetrics: map[string]StatMetric{
			utils.MetaTCD: &StatTCD{
				Metric: &Metric{
					Value: utils.NewDecimal(int64(time.Hour), 0),
					Count: 1,
					Events: map[string]*DecimalWithCompress{
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
	tmp := Cache
	tmpC := config.CgrConfig()
	defer func() {
		Cache = tmp
		config.SetCgrConfig(tmpC)
	}()

	cfg := config.NewDefaultCGRConfig()
	cfg.StatSCfg().StoreInterval = 1
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	Cache = NewCacheS(cfg, dm, nil, nil)
	filterS := NewFilterS(cfg, nil, dm)
	sS := NewStatService(dm, cfg, filterS, nil)

	sqPrf := &StatQueueProfile{
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
		Metrics: []*MetricWithFilters{
			{
				MetricID: utils.MetaTCD,
			},
		},
	}
	sq := &StatQueue{
		sqPrfl: sqPrf,
		dirty:  utils.BoolPointer(false),
		Tenant: "cgrates.org",
		ID:     "SQ1",
		SQItems: []SQItem{
			{
				EventID: "SqProcessEvent",
			},
		},
		SQMetrics: map[string]StatMetric{
			utils.MetaTCD: &StatTCD{
				Metric: &Metric{
					Value:  utils.NewDecimal(int64(time.Hour), 0),
					Count:  1,
					Events: make(map[string]*DecimalWithCompress),
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
	tmp := Cache
	tmpC := config.CgrConfig()
	defer func() {
		Cache = tmp
		config.SetCgrConfig(tmpC)
	}()

	cfg := config.NewDefaultCGRConfig()
	cfg.StatSCfg().StoreInterval = 1
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	Cache = NewCacheS(cfg, dm, nil, nil)
	filterS := NewFilterS(cfg, nil, dm)
	sS := NewStatService(dm, cfg, filterS, nil)

	sqPrf := &StatQueueProfile{
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
		Metrics: []*MetricWithFilters{
			{
				MetricID: utils.MetaTCD,
			},
		},
	}
	sq := &StatQueue{
		sqPrfl: sqPrf,
		dirty:  utils.BoolPointer(false),
		Tenant: "cgrates.org",
		ID:     "SQ1",
		SQItems: []SQItem{
			{
				EventID: "SqProcessEvent",
			},
		},
		SQMetrics: map[string]StatMetric{
			utils.MetaTCD: &StatTCD{
				Metric: &Metric{
					Value:  utils.NewDecimal(int64(time.Hour), 0),
					Count:  1,
					Events: make(map[string]*DecimalWithCompress),
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
	tmp := Cache
	tmpC := config.CgrConfig()
	defer func() {
		Cache = tmp
		config.SetCgrConfig(tmpC)
	}()

	cfg := config.NewDefaultCGRConfig()
	cfg.StatSCfg().StoreInterval = 1
	Cache = NewCacheS(cfg, nil, nil, nil)
	filterS := NewFilterS(cfg, nil, nil)
	sS := NewStatService(nil, cfg, filterS, nil)

	experr := `SERVER_ERROR: NO_DATABASE_CONNECTION`
	reply := map[string]string{}
	if err := sS.V1GetQueueStringMetrics(context.Background(), &utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{ID: "SQ1"}}, &reply); err == nil || err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}
}

func TestStatQueueStoreStatQueueStoreIntervalDisabled(t *testing.T) {
	tmp := Cache
	tmpC := config.CgrConfig()
	tmpCM := connMgr
	defer func() {
		Cache = tmp
		config.SetCgrConfig(tmpC)
		connMgr = tmpCM
	}()

	cfg := config.NewDefaultCGRConfig()
	cfg.StatSCfg().StoreInterval = -1
	config.SetCgrConfig(cfg)
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	connMgr = NewConnManager(cfg)
	Cache = NewCacheS(cfg, dm, nil, nil)
	filterS := NewFilterS(cfg, nil, dm)
	sS := NewStatService(dm, cfg, filterS, connMgr)

	sq := &StatQueue{
		Tenant:    "cgrates.org",
		ID:        "SQ1",
		SQMetrics: make(map[string]StatMetric),
		dirty:     utils.BoolPointer(true),
	}

	sS.storeStatQueue(context.Background(), sq)

	if *sq.dirty != false {
		t.Error("expected dirty to be false")
	}
}

func TestStatQueueGetStatQueueOK(t *testing.T) {
	tmp := Cache
	tmpC := config.CgrConfig()
	defer func() {
		Cache = tmp
		config.SetCgrConfig(tmpC)
	}()

	cfg := config.NewDefaultCGRConfig()
	cfg.StatSCfg().StoreInterval = 1
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	Cache = NewCacheS(cfg, dm, nil, nil)
	filterS := NewFilterS(cfg, nil, dm)
	sS := NewStatService(dm, cfg, filterS, nil)

	sqPrf := &StatQueueProfile{
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
		Metrics: []*MetricWithFilters{
			{
				MetricID: utils.MetaTCD,
			},
		},
	}
	sq := &StatQueue{
		sqPrfl: sqPrf,
		dirty:  utils.BoolPointer(false),
		Tenant: "cgrates.org",
		ID:     "SQ1",
		SQItems: []SQItem{
			{
				EventID:    "SqProcessEvent",
				ExpiryTime: utils.TimePointer(time.Now()),
			},
		},
		SQMetrics: map[string]StatMetric{
			utils.MetaTCD: &StatTCD{
				Metric: &Metric{
					Value:  utils.NewDecimal(int64(time.Hour), 0),
					Count:  1,
					Events: make(map[string]*DecimalWithCompress),
				},
			},
		},
	}

	if err := dm.SetStatQueue(context.Background(), sq); err != nil {
		t.Error(err)
	}

	expected := utils.StringSet{
		utils.ConcatenatedKey(sq.Tenant, sq.ID): struct{}{},
	}
	if rcv, err := sS.getStatQueue(context.Background(), "cgrates.org", "SQ1"); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, sq) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>",
			utils.ToJSON(sq), utils.ToJSON(rcv))
	} else if !reflect.DeepEqual(sS.storedStatQueues, expected) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", expected, sS.storedStatQueues)
	}
}
func TestStatQueueProcessEventProfileIgnoreFilters(t *testing.T) {
	Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	filterS := NewFilterS(cfg, nil, dm)
	sS := NewStatService(dm, cfg, filterS, nil)
	cfg.StatSCfg().Opts.ProfileIgnoreFilters = []*utils.DynamicBoolOpt{
		{
			Value: true,
		},
	}
	sqPrf := &StatQueueProfile{
		Tenant:    "cgrates.org",
		ID:        "SQ1",
		FilterIDs: []string{"*string:~*req.Stat:testStatValue"},
	}
	sq := &StatQueue{
		sqPrfl: sqPrf,
		Tenant: "cgrates.org",
		ID:     "SQ1",
		SQItems: []SQItem{
			{
				EventID: "SqProcessEvent",
			},
		},
		SQMetrics: make(map[string]StatMetric),
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
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	filterS := NewFilterS(cfg, nil, dm)
	sS := NewStatService(dm, cfg, filterS, nil)
	cfg.StatSCfg().Opts.ProfileIgnoreFilters = []*utils.DynamicBoolOpt{
		{
			Value: true,
		},
	}
	sqPrf := &StatQueueProfile{
		Tenant:    "cgrates.org",
		ID:        "SQ1",
		FilterIDs: []string{"*string:~*req.Stat:testStatValue"},
	}
	sq := &StatQueue{
		sqPrfl: sqPrf,
		Tenant: "cgrates.org",
		ID:     "SQ1",
		SQItems: []SQItem{
			{
				EventID: "SqProcessEvent",
			},
		},
		SQMetrics: make(map[string]StatMetric),
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
	tmp := Cache
	tmpC := config.CgrConfig()
	defer func() {
		Cache = tmp
		config.SetCgrConfig(tmpC)
	}()

	cfg := config.NewDefaultCGRConfig()
	cfg.StatSCfg().Opts.ProfileIgnoreFilters = []*utils.DynamicBoolOpt{
		{
			Value: true,
		},
	}
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	Cache = NewCacheS(cfg, dm, nil, nil)
	filterS := NewFilterS(cfg, nil, dm)
	sS := NewStatService(dm, cfg, filterS, nil)

	sqPrf1 := &StatQueueProfile{
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
		Metrics: []*MetricWithFilters{
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
	tmp := Cache
	tmpC := config.CgrConfig()
	defer func() {
		Cache = tmp
		config.SetCgrConfig(tmpC)
	}()

	cfg := config.NewDefaultCGRConfig()
	cfg.StatSCfg().StoreInterval = 1
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	Cache = NewCacheS(cfg, dm, nil, nil)
	filterS := NewFilterS(cfg, nil, dm)
	sS := NewStatService(dm, cfg, filterS, nil)

	sqPrf := &StatQueueProfile{
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
		Metrics: []*MetricWithFilters{
			{
				MetricID: utils.MetaTCD,
			},
		},
	}
	sq := &StatQueue{
		sqPrfl: sqPrf,
		dirty:  utils.BoolPointer(false),
		Tenant: "cgrates.org",
		ID:     "SQ1",
		SQItems: []SQItem{
			{
				EventID: "SqProcessEvent",
			},
		},
		SQMetrics: map[string]StatMetric{
			utils.MetaTCD: &StatTCD{
				Metric: &Metric{
					Value: utils.NewDecimal(int64(time.Hour), 0),
					Count: 1,
					Events: map[string]*DecimalWithCompress{
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
	tmp := Cache
	tmpC := config.CgrConfig()
	defer func() {
		Cache = tmp
		config.SetCgrConfig(tmpC)
	}()

	cfg := config.NewDefaultCGRConfig()
	cfg.StatSCfg().StoreInterval = 1
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	Cache = NewCacheS(cfg, dm, nil, nil)
	filterS := NewFilterS(cfg, nil, dm)
	sS := NewStatService(dm, cfg, filterS, nil)

	sqPrf := &StatQueueProfile{
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
		Metrics: []*MetricWithFilters{
			{
				MetricID: utils.MetaTCD,
			},
		},
	}
	sq := &StatQueue{
		sqPrfl: sqPrf,
		dirty:  utils.BoolPointer(false),
		Tenant: "cgrates.org",
		ID:     "SQ1",
		SQItems: []SQItem{
			{
				EventID: "SqProcessEvent",
			},
		},
		SQMetrics: map[string]StatMetric{
			utils.MetaTCD: &StatTCD{
				Metric: &Metric{
					Value: utils.NewDecimal(int64(time.Hour), 0),
					Count: 1,
					Events: map[string]*DecimalWithCompress{
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
	tmp := Cache
	tmpC := config.CgrConfig()
	defer func() {
		Cache = tmp
		config.SetCgrConfig(tmpC)
	}()

	cfg := config.NewDefaultCGRConfig()
	cfg.StatSCfg().StoreInterval = 1
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	Cache = NewCacheS(cfg, dm, nil, nil)
	filterS := NewFilterS(cfg, nil, dm)
	sS := NewStatService(dm, cfg, filterS, nil)

	sqPrf := &StatQueueProfile{
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
		Metrics: []*MetricWithFilters{
			{
				MetricID: utils.MetaTCD,
			},
		},
	}
	sq := &StatQueue{
		sqPrfl: sqPrf,
		dirty:  utils.BoolPointer(false),
		Tenant: "cgrates.org",
		ID:     "SQ1",
		SQItems: []SQItem{
			{
				EventID: "SqProcessEvent",
			},
		},
		SQMetrics: map[string]StatMetric{
			utils.MetaTCD: &StatTCD{
				Metric: &Metric{
					Value:  utils.NewDecimal(int64(time.Hour), 0),
					Count:  1,
					Events: make(map[string]*DecimalWithCompress),
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
	tmp := Cache
	tmpC := config.CgrConfig()
	defer func() {
		Cache = tmp
		config.SetCgrConfig(tmpC)
	}()

	cfg := config.NewDefaultCGRConfig()
	cfg.StatSCfg().StoreInterval = 1
	Cache = NewCacheS(cfg, nil, nil, nil)
	filterS := NewFilterS(cfg, nil, nil)
	sS := NewStatService(nil, cfg, filterS, nil)

	experr := `SERVER_ERROR: NO_DATABASE_CONNECTION`
	var reply map[string]*utils.Decimal
	if err := sS.V1GetQueueDecimalMetrics(context.Background(), &utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{ID: "SQ1"}}, &reply); err == nil || err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}
}

func TestStatSV1GetQueueStringMetricsIntOptsErr(t *testing.T) {
	Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dmSTS := NewDataManager(data, config.CgrConfig().CacheCfg(), nil)

	cfg.StatSCfg().StoreInterval = 1
	cfg.StatSCfg().StringIndexedFields = nil
	cfg.StatSCfg().PrefixIndexedFields = nil
	cfg.StatSCfg().Opts.RoundingDecimals = []*utils.DynamicIntOpt{
		// function will return error after trying to parse the filter
		{
			FilterIDs: []string{"*string.invalid:filter"},
			Tenant:    "cgrates.org",
			Value:     4,
		},
	}
	statService := NewStatService(dmSTS, cfg,
		&FilterS{dm: dmSTS, cfg: cfg}, nil)

	prepareStatsData(t, dmSTS)

	stq := map[string]string{}

	experr := `inline parse error for string: <*string.invalid:filter>`

	err := statService.V1GetQueueStringMetrics(context.TODO(), &utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: testStatsQ[0].Tenant, ID: testStatsQ[0].ID}}, &stq)
	if err.Error() != experr {
		t.Errorf("Expected error <%v>, Received <%v>", experr, err)
	}

}

func TestStatSV1GetStatQueuesForEventsqIDsErr(t *testing.T) {
	tmp := Cache
	tmpC := config.CgrConfig()
	defer func() {
		Cache = tmp
		config.SetCgrConfig(tmpC)
	}()

	cfg := config.NewDefaultCGRConfig()
	cfg.StatSCfg().Opts.ProfileIDs = []*utils.DynamicStringSliceOpt{
		// function will return error after trying to parse the filter
		{
			FilterIDs: []string{"*string.invalid:filter"},
			Tenant:    "cgrates.org",
			Values:    []string{"value2"},
		},
	}
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	Cache = NewCacheS(cfg, dm, nil, nil)
	filterS := NewFilterS(cfg, nil, dm)
	sS := NewStatService(dm, cfg, filterS, nil)

	sqPrf1 := &StatQueueProfile{
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
		Metrics: []*MetricWithFilters{
			{
				MetricID: utils.MetaTCD,
			},
		},
	}

	if err := dm.SetStatQueueProfile(context.Background(), sqPrf1, true); err != nil {
		t.Error(err)
	}

	sqPrf2 := &StatQueueProfile{
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
		Metrics: []*MetricWithFilters{
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
	tmp := Cache
	tmpC := config.CgrConfig()
	defer func() {
		Cache = tmp
		config.SetCgrConfig(tmpC)
	}()

	cfg := config.NewDefaultCGRConfig()
	cfg.StatSCfg().Opts.ProfileIgnoreFilters = []*utils.DynamicBoolOpt{
		// function will return error after trying to parse the filter
		{
			FilterIDs: []string{"*string.invalid:filter"},
			Tenant:    "cgrates.org",
			Value:     false,
		},
	}
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	Cache = NewCacheS(cfg, dm, nil, nil)
	filterS := NewFilterS(cfg, nil, dm)
	sS := NewStatService(dm, cfg, filterS, nil)

	sqPrf1 := &StatQueueProfile{
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
		Metrics: []*MetricWithFilters{
			{
				MetricID: utils.MetaTCD,
			},
		},
	}

	if err := dm.SetStatQueueProfile(context.Background(), sqPrf1, true); err != nil {
		t.Error(err)
	}

	sqPrf2 := &StatQueueProfile{
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
		Metrics: []*MetricWithFilters{
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
	Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dmSTS := NewDataManager(data, config.CgrConfig().CacheCfg(), nil)

	cfg.StatSCfg().StoreInterval = 1
	cfg.StatSCfg().StringIndexedFields = nil
	cfg.StatSCfg().PrefixIndexedFields = nil
	statService := NewStatService(dmSTS, cfg,
		&FilterS{dm: dmSTS, cfg: cfg}, nil)

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
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dmSTS := NewDataManager(data, config.CgrConfig().CacheCfg(), nil)

	cfg.StatSCfg().StoreInterval = 1
	cfg.StatSCfg().StringIndexedFields = nil
	cfg.StatSCfg().PrefixIndexedFields = nil

	statService := NewStatService(dmSTS, cfg,
		&FilterS{dm: dmSTS, cfg: cfg}, nil)
	prepareStatsData(t, dmSTS)
	_, err := statService.matchingStatQueuesForEvent(context.TODO(), testStatsArgs[0].Tenant, []string{"statsIds"},
		testStatsArgs[0].AsDataProvider(), false)
	if err == nil || err != utils.ErrNotFound {
		t.Errorf("Expected error <%v>, Received error <%v>", utils.ErrNotFound, err)
	}
}

func TestStatQueuesMatchingStatQueuesForEventWeightErr(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dmSTS := NewDataManager(data, config.CgrConfig().CacheCfg(), nil)

	cfg.StatSCfg().StoreInterval = 1
	cfg.StatSCfg().StringIndexedFields = nil
	cfg.StatSCfg().PrefixIndexedFields = nil
	statService := NewStatService(dmSTS, cfg,
		&FilterS{dm: dmSTS, cfg: cfg}, nil)
	prepareStatsData(t, dmSTS)

	sqp := &StatQueueProfile{
		Tenant:      "cgrates.org",
		ID:          "StatQueueProfile1",
		FilterIDs:   []string{"FLTR_STATS_1"},
		QueueLength: 10,
		TTL:         10 * time.Second,
		Metrics: []*MetricWithFilters{
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

	expErr := "NOT_IMPLEMENTED:*stirng"
	_, err := statService.matchingStatQueuesForEvent(context.TODO(), testStatsArgs[0].Tenant, nil,
		testStatsArgs[0].AsDataProvider(), false)

	if err == nil || err.Error() != expErr {
		t.Errorf("Expected error <%v>, Received error <%v>", expErr, err)
	}
}

func TestStatQueueProcessEventProfileIDsErr(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	filterS := NewFilterS(cfg, nil, dm)
	sS := NewStatService(dm, cfg, filterS, nil)

	args := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "SqProcessEvent",
		Event: map[string]any{
			utils.AccountField: "1001",
		},
		APIOpts: map[string]any{},
	}
	sS.cfg.StatSCfg().Opts.ProfileIDs = []*utils.DynamicStringSliceOpt{

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

func TestStatQueueProcessEventPrometheusStatIDsErr(t *testing.T) {

	defer func() {
		Cache = NewCacheS(config.NewDefaultCGRConfig(), nil, nil, nil)
	}()

	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	filterS := NewFilterS(cfg, nil, dm)
	sS := NewStatService(dm, cfg, filterS, nil)

	sqPrf := &StatQueueProfile{
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
		Metrics: []*MetricWithFilters{
			{
				MetricID: utils.MetaTCD,
			},
		},
	}
	stat, err := NewStatMetric("*tcd", uint64(sqPrf.MinItems), []string{})
	if err != nil {
		t.Fatal(err)
	}
	stq := &StatQueue{
		sqPrfl: sqPrf,
		Tenant: "cgrates.org",
		ID:     "SQ1",
		SQItems: []SQItem{
			{
				EventID: "SqProcessEvent",
			},
		},
		SQMetrics: map[string]StatMetric{
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

	sS.cfg.StatSCfg().Opts.PrometheusStatIDs = []*utils.DynamicStringSliceOpt{

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
		Cache = NewCacheS(config.NewDefaultCGRConfig(), nil, nil, nil)
	}()

	buf := new(bytes.Buffer)
	utils.Logger = utils.NewStdLoggerWithWriter(buf, "", 7)

	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	filterS := NewFilterS(cfg, nil, dm)
	sS := NewStatService(dm, cfg, filterS, nil)

	sqPrf := &StatQueueProfile{
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
		Metrics: []*MetricWithFilters{
			{
				MetricID: utils.MetaTCD,
			},
		},
	}
	expiry := time.Date(2021, 1, 1, 23, 59, 59, 10, time.UTC)

	stq := &StatQueue{
		sqPrfl: sqPrf,
		Tenant: "cgrates.org",
		ID:     "SQ1",
		SQItems: []SQItem{
			{
				EventID:    "SqProcessEvent",
				ExpiryTime: &expiry,
			},
		},
		SQMetrics: map[string]StatMetric{
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

	if rcv, err := sS.processEvent(context.Background(), args.Tenant, args); err != nil {
		t.Error(err)
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

	defer func() {
		Cache = NewCacheS(config.NewDefaultCGRConfig(), nil, nil, nil)
	}()

	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	filterS := NewFilterS(cfg, nil, dm)
	sS := NewStatService(dm, cfg, filterS, nil)

	sqPrf := &StatQueueProfile{
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
		Metrics: []*MetricWithFilters{
			{
				MetricID: utils.MetaTCD,
			},
		},
	}
	stat, err := NewStatMetric("*tcd", uint64(sqPrf.MinItems), []string{})
	if err != nil {
		t.Fatal(err)
	}
	stq := &StatQueue{
		sqPrfl: sqPrf,
		Tenant: "cgrates.org",
		ID:     "SQ1",
		SQItems: []SQItem{
			{
				EventID: "SqProcessEvent",
			},
		},
		SQMetrics: map[string]StatMetric{
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

	expErr := "NOT_IMPLEMENTED:*stirng"
	if _, err := sS.processEvent(context.Background(), args.Tenant, args); err == nil || err.Error() != expErr {
		t.Errorf("Expected error <%v>, Received error <%v>", expErr, err)
	}

}
