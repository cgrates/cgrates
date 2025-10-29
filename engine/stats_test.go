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
package engine

import (
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

var (
	cloneExpTimeStats time.Time
	expTimeStats      = time.Now().Add(time.Duration(20 * time.Minute))
	statService       *StatService
	dmSTS             *DataManager
	sqps              = []*StatQueueProfile{
		{
			Tenant:    "cgrates.org",
			ID:        "StatQueueProfile1",
			FilterIDs: []string{"FLTR_STATS_1"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			},
			QueueLength: 10,
			TTL:         time.Duration(10) * time.Second,
			Metrics: []*MetricWithFilters{
				{
					MetricID: "*sum#Usage",
				},
			},
			ThresholdIDs: []string{},
			Blocker:      false,
			Stored:       true,
			Weight:       20,
			MinItems:     1,
		},
		{
			Tenant:    "cgrates.org",
			ID:        "StatQueueProfile2",
			FilterIDs: []string{"FLTR_STATS_2"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			},
			QueueLength: 10,
			TTL:         time.Duration(10) * time.Second,
			Metrics: []*MetricWithFilters{
				{
					MetricID: "*sum#Usage",
				},
			},
			ThresholdIDs: []string{},
			Blocker:      false,
			Stored:       true,
			Weight:       20,
			MinItems:     1,
		},
		{
			Tenant:    "cgrates.org",
			ID:        "StatQueueProfilePrefix",
			FilterIDs: []string{"FLTR_STATS_3"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			},
			QueueLength: 10,
			TTL:         time.Duration(10) * time.Second,
			Metrics: []*MetricWithFilters{
				{
					MetricID: "*sum#Usage",
				},
			},
			ThresholdIDs: []string{},
			Blocker:      false,
			Stored:       true,
			Weight:       20,
			MinItems:     1,
		},
	}
	stqs = []*StatQueue{
		{Tenant: "cgrates.org", ID: "StatQueueProfile1", sqPrfl: sqps[0]},
		{Tenant: "cgrates.org", ID: "StatQueueProfile2", sqPrfl: sqps[1]},
		{Tenant: "cgrates.org", ID: "StatQueueProfilePrefix", sqPrfl: sqps[2]},
	}
	statsEvs = []*StatsArgsProcessEvent{
		{
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "event1",
				Event: map[string]any{
					"Stats":          "StatQueueProfile1",
					utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
					"UsageInterval":  "1s",
					"PddInterval":    "1s",
					"Weight":         "9.0",
					utils.Usage:      time.Duration(135 * time.Second),
					utils.COST:       123.0,
				},
			},
		},
		{
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "event2",
				Event: map[string]any{
					"Stats":          "StatQueueProfile2",
					utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
					"UsageInterval":  "1s",
					"PddInterval":    "1s",
					"Weight":         "15.0",
					utils.Usage:      time.Duration(45 * time.Second),
				},
			},
		},
		{
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "event3",
				Event: map[string]any{
					"Stats":     "StatQueueProfilePrefix",
					utils.Usage: time.Duration(30 * time.Second),
				},
			},
		},
	}
)

func TestStatQueuesPopulateService(t *testing.T) {
	defaultCfg, _ := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, true, defaultCfg.DataDbCfg().Items)
	dmSTS = NewDataManager(data, config.CgrConfig().CacheCfg(), nil)

	defaultCfg.StatSCfg().StoreInterval = 1
	defaultCfg.StatSCfg().StringIndexedFields = nil
	defaultCfg.StatSCfg().PrefixIndexedFields = nil
	statService, err = NewStatService(dmSTS, defaultCfg,
		&FilterS{dm: dmSTS, cfg: defaultCfg}, nil)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
}

func TestStatQueuesAddFilters(t *testing.T) {
	fltrSts1 := &Filter{
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
				Values:  []string{(1 * time.Second).String()},
			},
			{
				Type:    utils.MetaGreaterOrEqual,
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Usage,
				Values:  []string{(1 * time.Second).String()},
			},
			{
				Type:    utils.MetaGreaterOrEqual,
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Weight,
				Values:  []string{"9.0"},
			},
		},
	}
	dmSTS.SetFilter(fltrSts1)
	fltrSts2 := &Filter{
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
				Values:  []string{(1 * time.Second).String()},
			},
			{
				Type:    utils.MetaGreaterOrEqual,
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Usage,
				Values:  []string{(1 * time.Second).String()},
			},
			{
				Type:    utils.MetaGreaterOrEqual,
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Weight,
				Values:  []string{"15.0"},
			},
		},
	}
	dmSTS.SetFilter(fltrSts2)
	fltrSts3 := &Filter{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "FLTR_STATS_3",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaPrefix,
				Element: "~*req.Stats",
				Values:  []string{"StatQueueProfilePrefix"},
			},
		},
	}
	dmSTS.SetFilter(fltrSts3)
}

func TestStatQueuesPopulateStatsService(t *testing.T) {
	for _, statQueueProfile := range sqps {
		dmSTS.SetStatQueueProfile(statQueueProfile, true)
	}
	for _, statQueue := range stqs {
		dmSTS.SetStatQueue(statQueue)
	}
	//Test each statQueueProfile from cache
	for _, sqp := range sqps {
		if tempStat, err := dmSTS.GetStatQueueProfile(sqp.Tenant,
			sqp.ID, true, false, utils.NonTransactional); err != nil {
			t.Errorf("Error: %+v", err)
		} else if !reflect.DeepEqual(sqp, tempStat) {
			t.Errorf("Expecting: %+v, received: %+v", sqp, tempStat)
		}
	}
}

func TestStatQueuesMatchingStatQueuesForEvent(t *testing.T) {
	msq, err := statService.matchingStatQueuesForEvent(statsEvs[0])
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(stqs[0].Tenant, msq[0].Tenant) {
		t.Errorf("Expecting: %+v, received: %+v", stqs[0].Tenant, msq[0].Tenant)
	} else if !reflect.DeepEqual(stqs[0].ID, msq[0].ID) {
		t.Errorf("Expecting: %+v, received: %+v", stqs[0].ID, msq[0].ID)
	} else if !reflect.DeepEqual(stqs[0].sqPrfl, msq[0].sqPrfl) {
		t.Errorf("Expecting: %+v, received: %+v", stqs[0].sqPrfl, msq[0].sqPrfl)
	}
	msq, err = statService.matchingStatQueuesForEvent(statsEvs[1])
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(stqs[1].Tenant, msq[0].Tenant) {
		t.Errorf("Expecting: %+v, received: %+v", stqs[1].Tenant, msq[0].Tenant)
	} else if !reflect.DeepEqual(stqs[1].ID, msq[0].ID) {
		t.Errorf("Expecting: %+v, received: %+v", stqs[1].ID, msq[0].ID)
	} else if !reflect.DeepEqual(stqs[1].sqPrfl, msq[0].sqPrfl) {
		t.Errorf("Expecting: %+v, received: %+v", stqs[1].sqPrfl, msq[0].sqPrfl)
	}
	msq, err = statService.matchingStatQueuesForEvent(statsEvs[2])
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(stqs[2].Tenant, msq[0].Tenant) {
		t.Errorf("Expecting: %+v, received: %+v", stqs[2].Tenant, msq[0].Tenant)
	} else if !reflect.DeepEqual(stqs[2].ID, msq[0].ID) {
		t.Errorf("Expecting: %+v, received: %+v", stqs[2].ID, msq[0].ID)
	} else if !reflect.DeepEqual(stqs[2].sqPrfl, msq[0].sqPrfl) {
		t.Errorf("Expecting: %+v, received: %+v", stqs[2].sqPrfl, msq[0].sqPrfl)
	}
}

func TestStatQueuesProcessEvent(t *testing.T) {
	stq := map[string]string{}
	reply := []string{}
	expected := []string{"StatQueueProfile1"}
	err := statService.V1ProcessEvent(statsEvs[0], &reply)
	if err != nil {
		t.Errorf("Error: %+v", err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expecting: %+v, received: %+v", expected, reply)
	}
	err = statService.V1GetQueueStringMetrics(&utils.TenantID{Tenant: stqs[0].Tenant, ID: stqs[0].ID}, &stq)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}

	expected = []string{"StatQueueProfile2"}
	err = statService.V1ProcessEvent(statsEvs[1], &reply)
	if err != nil {
		t.Errorf("Error: %+v", err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expecting: %+v, received: %+v", expected, reply)
	}
	err = statService.V1GetQueueStringMetrics(&utils.TenantID{Tenant: stqs[1].Tenant, ID: stqs[1].ID}, &stq)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}

	expected = []string{"StatQueueProfilePrefix"}
	err = statService.V1ProcessEvent(statsEvs[2], &reply)
	if err != nil {
		t.Errorf("Error: %+v", err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expecting: %+v, received: %+v", expected, reply)
	}
	err = statService.V1GetQueueStringMetrics(&utils.TenantID{Tenant: stqs[2].Tenant, ID: stqs[2].ID}, &stq)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
}

func TestStatQueuesMatchWithIndexFalse(t *testing.T) {
	statService.cgrcfg.StatSCfg().IndexedSelects = false
	msq, err := statService.matchingStatQueuesForEvent(statsEvs[0])
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(stqs[0].Tenant, msq[0].Tenant) {
		t.Errorf("Expecting: %+v, received: %+v", stqs[0].Tenant, msq[0].Tenant)
	} else if !reflect.DeepEqual(stqs[0].ID, msq[0].ID) {
		t.Errorf("Expecting: %+v, received: %+v", stqs[0].ID, msq[0].ID)
	} else if !reflect.DeepEqual(stqs[0].sqPrfl, msq[0].sqPrfl) {
		t.Errorf("Expecting: %+v, received: %+v", stqs[0].sqPrfl, msq[0].sqPrfl)
	}
	msq, err = statService.matchingStatQueuesForEvent(statsEvs[1])
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(stqs[1].Tenant, msq[0].Tenant) {
		t.Errorf("Expecting: %+v, received: %+v", stqs[1].Tenant, msq[0].Tenant)
	} else if !reflect.DeepEqual(stqs[1].ID, msq[0].ID) {
		t.Errorf("Expecting: %+v, received: %+v", stqs[1].ID, msq[0].ID)
	} else if !reflect.DeepEqual(stqs[1].sqPrfl, msq[0].sqPrfl) {
		t.Errorf("Expecting: %+v, received: %+v", stqs[1].sqPrfl, msq[0].sqPrfl)
	}
	msq, err = statService.matchingStatQueuesForEvent(statsEvs[2])
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(stqs[2].Tenant, msq[0].Tenant) {
		t.Errorf("Expecting: %+v, received: %+v", stqs[2].Tenant, msq[0].Tenant)
	} else if !reflect.DeepEqual(stqs[2].ID, msq[0].ID) {
		t.Errorf("Expecting: %+v, received: %+v", stqs[2].ID, msq[0].ID)
	} else if !reflect.DeepEqual(stqs[2].sqPrfl, msq[0].sqPrfl) {
		t.Errorf("Expecting: %+v, received: %+v", stqs[2].sqPrfl, msq[0].sqPrfl)
	}
}
func TestStatQueuesV1ProcessEvent(t *testing.T) {
	sqPrf := &StatQueueProfile{
		Tenant:    "cgrates.org",
		ID:        "StatQueueProfile3",
		FilterIDs: []string{"FLTR_STATS_1"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
		QueueLength: 10,
		TTL:         time.Duration(10) * time.Second,
		Metrics: []*MetricWithFilters{
			{
				MetricID: "*sum#Usage",
			},
		},
		ThresholdIDs: []string{},
		Stored:       true,
		Weight:       20,
		MinItems:     1,
	}
	sq := &StatQueue{Tenant: "cgrates.org", ID: "StatQueueProfile3", sqPrfl: sqPrf}
	if err := dmSTS.SetStatQueueProfile(sqPrf, true); err != nil {
		t.Error(err)
	}
	if err := dmSTS.SetStatQueue(sq); err != nil {
		t.Error(err)
	}
	if tempStat, err := dmSTS.GetStatQueueProfile(sqPrf.Tenant,
		sqPrf.ID, true, false, utils.NonTransactional); err != nil {
		t.Errorf("Error: %+v", err)
	} else if !reflect.DeepEqual(sqPrf, tempStat) {
		t.Errorf("Expecting: %+v, received: %+v", sqPrf, tempStat)
	}
	ev := &StatsArgsProcessEvent{
		StatIDs:  []string{"StatQueueProfile1", "StatQueueProfile2", "StatQueueProfile3"},
		CGREvent: statsEvs[0].CGREvent,
	}
	reply := []string{}
	expected := []string{"StatQueueProfile1", "StatQueueProfile3"}
	expectedRev := []string{"StatQueueProfile3", "StatQueueProfile1"}
	if err := statService.V1ProcessEvent(ev, &reply); err != nil {
		t.Errorf("Error: %+v", err)
	} else if !reflect.DeepEqual(reply, expected) && !reflect.DeepEqual(reply, expectedRev) {
		t.Errorf("Expecting: %+v, received: %+v", expected, reply)
	}
}

func TestStatQueueFloatMetrics(t *testing.T) {
	cfg, err := config.NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
	}
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)

	sS, err := NewStatService(dm, cfg, nil, nil)
	if err != nil {
		t.Error(err)
	}
	sq := &StatQueue{
		Tenant: "cgrates.org",
		ID:     "TestStats",
		SQItems: []SQItem{
			{EventID: "cgrates.org:ev1"},
			{EventID: "cgrates.org:ev2"},
			{EventID: "cgrates.org:ev3"},
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
	dm.SetStatQueue(sq)
	args := &utils.TenantID{
		Tenant: "cgrates.org",
		ID:     "TestStats",
	}
	var reply map[string]float64
	exp := map[string]float64{
		utils.MetaASR: 66.66667,
	}
	if err := sS.V1GetQueueFloatMetrics(args, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, exp) {
		t.Errorf("Expected %v,Received %v", utils.ToJSON(exp), utils.ToJSON(reply))
	}
}

func TestStatStoreStatQueue(t *testing.T) {
	cfg, err := config.NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
	}
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)

	sS, err := NewStatService(dm, cfg, nil, nil)
	if err != nil {
		t.Error(err)
	}
	sq := &StatQueue{
		Tenant: "cgrates.org",
		ID:     "SQ_1",
		dirty:  utils.BoolPointer(true),
		SQMetrics: map[string]StatMetric{
			utils.MetaASR: &StatASR{
				Answered: 1,
				Count:    1,
				Events: map[string]*StatWithCompress{
					"cgrates.org:TestStatRemExpired_1": {Stat: 1, CompressFactor: 1},
				},
			},
		},
		sqPrfl: &StatQueueProfile{
			QueueLength: 2, //unlimited que
		},
	}
	if err := sS.StoreStatQueue(sq); err != nil {
		t.Error(err)
	} else if *sq.dirty {
		t.Error("Expected false")
	}
	args := &utils.TenantIDWithArgDispatcher{
		TenantID: &utils.TenantID{
			Tenant: "cgrates.org",
			ID:     "SQ_1",
		},
	}
	var reply StatQueue
	if err := sS.V1GetStatQueue(args, &reply); err != nil {
		t.Error(err)
	}
}

func TestStatsGetStatQueuesForEvent(t *testing.T) {
	cfg, err := config.NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
	}
	Cache.Clear(nil)
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	sS, err := NewStatService(dm, cfg, NewFilterS(cfg, nil, dm), nil)
	if err != nil {
		t.Error(err)
	}
	args := &StatsArgsProcessEvent{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     utils.GenUUID(),
			Event: map[string]any{
				"Stat":           "Stat1_1",
				utils.AnswerTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
				utils.Usage:      time.Duration(11 * time.Second),
				utils.COST:       12.5,
				utils.PDD:        time.Duration(12 * time.Second),
			},
		},
	}
	fltr := &Filter{
		Tenant: "cgrates.org",
		ID:     "FLTR_STAT_1",
		Rules: []*FilterRule{{
			Type:    "*string",
			Element: "~*req.Stat",
			Values:  []string{"Stat1_1"},
		},
		},
	}
	dm.SetFilter(fltr)
	sqP := &StatQueueProfile{
		Tenant: "cgrates.org",
		ID:     "Stats1",
		FilterIDs: []string{
			"FLTR_STAT_1",
		},
		QueueLength: 10,
		TTL:         time.Duration(10) * time.Second,
		Metrics: []*MetricWithFilters{
			{
				MetricID: utils.MetaACD,
			},
			{
				MetricID: utils.MetaTCD,
			},
		},
		ThresholdIDs: []string{"*none"},
		Weight:       20,
		MinItems:     1,
	}
	sq := &StatQueue{
		Tenant: "cgrates.org",
		ID:     "Stats1",
		SQItems: []SQItem{
			{EventID: "cgrates.org:ev1", ExpiryTime: utils.TimePointer(time.Date(2023, 12, 24, 17, 0, 0, 0, time.UTC))},
		},
		SQMetrics: map[string]StatMetric{
			utils.MetaASR: &StatASR{
				Answered: 2,
				Count:    3,
				Events: map[string]*StatWithCompress{
					"cgrates.org:ev1": {Stat: 1},
				},
			},
		},
	}
	dm.SetStatQueue(sq)
	dm.SetStatQueueProfile(sqP, true)
	var reply []string
	expIds := []string{
		"Stats1",
	}
	if err := sS.V1GetStatQueuesForEvent(args, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expIds, reply) {
		t.Errorf("Expected %v,Received %v", expIds, reply)
	}
}

func TestStatSGetQueueIDs(t *testing.T) {
	cfg, err := config.NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
	}
	Cache.Clear(nil)
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	sS, err := NewStatService(dm, cfg, NewFilterS(cfg, nil, dm), nil)
	if err != nil {
		t.Error(err)
	}
	sqs := StatQueues{
		&StatQueue{
			Tenant: "cgrates.org",
			ID:     "Stats1",
			SQMetrics: map[string]StatMetric{
				"*tcc":           nil,
				"*sum:~Usage":    nil,
				"*avreage:~Cost": nil,
			},
		},
		&StatQueue{
			Tenant: "cgrates.org",
			ID:     "Stats2",
			SQMetrics: map[string]StatMetric{
				utils.MetaASR: &StatASR{
					Answered: 2,
					Count:    3,
				},
			},
		},
	}
	for _, sq := range sqs {
		dm.SetStatQueue(sq)
	}
	var qIDs []string
	expqIds := []string{
		"Stats1", "Stats2",
	}
	if err := sS.V1GetQueueIDs("cgrates.org", &qIDs); err != nil {
		t.Error(err)
	}
	sort.Slice(qIDs, func(i, j int) bool {
		return qIDs[i] < qIDs[j]
	})
	if !reflect.DeepEqual(expqIds, qIDs) {
		t.Errorf("Expected %v,Received %v", expqIds, qIDs)
	}
}

func TestStatSReloadRunBackUp(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	cfg.StatSCfg().StoreInterval = 1
	Cache.Clear(nil)
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	sts, err := NewStatService(dm, cfg, nil, nil)
	sts.storedStatQueues["SQ1"] = true
	if err != nil {
		t.Error(err)
	}
	go func() {
		time.Sleep(2 * time.Millisecond)
		sts.loopStoped <- struct{}{}
	}()
	Cache.Set(utils.CacheStatQueues, "SQ1", &StatQueue{
		Tenant: "cgrates.org",
		ID:     "SQ1",
		SQMetrics: map[string]StatMetric{
			utils.MetaASR: &StatASR{
				Answered: 1,
				Count:    2,
				Events: map[string]*StatWithCompress{
					"cgrates.org:TestRemEventWithID_1": {Stat: 1, CompressFactor: 1},
				},
			},
		}}, []string{}, true, utils.NonTransactional)
	sts.Reload()
}

func TestStatProcessEvent2(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	tpDm := dm
	defer func() {
		dm = tpDm
	}()
	cfg.StatSCfg().ThresholdSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds)}
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	clientConn := make(chan birpc.ClientConnector, 1)
	clientConn <- clMock(func(ctx *context.Context, serviceMethod string, _, _ any) error {
		if serviceMethod == utils.ThresholdSv1ProcessEvent {

			return nil
		}
		return utils.ErrNotImplemented
	})
	connMgr := NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds): clientConn,
	})
	sts, err := NewStatService(dm, cfg, NewFilterS(cfg, nil, dm), connMgr)
	if err != nil {
		t.Error(err)
	}

	args := &StatsArgsProcessEvent{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "event1",
			Event: map[string]any{
				utils.Account:    "1001",
				utils.AnswerTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
				utils.Usage:      time.Duration(135 * time.Second),
				utils.COST:       123.0,
				utils.PDD:        time.Duration(12 * time.Second),
			},
		},
		ArgDispatcher: &utils.ArgDispatcher{
			APIKey: utils.StringPointer("12345"),
		},
	}
	fltr := &Filter{
		Tenant: "cgrates.org",
		ID:     "FLTR_TH_Stats1",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaString,
				Element: "~*req.Account",
				Values:  []string{"1001"},
			},
		},
	}
	dm.SetFilter(fltr)
	dm.SetStatQueueProfile(&StatQueueProfile{
		Tenant:    "cgrates.org",
		ID:        "STS_PoccessCDR",
		FilterIDs: []string{"FLTR_TH_Stats1"},
		Metrics: []*MetricWithFilters{{
			MetricID: "*sum:~*req.Usage",
		}},
		ThresholdIDs: []string{"Th1"},
		Blocker:      true,
		Stored:       true,
		Weight:       20,
		MinItems:     0,
	}, true)

	dm.SetStatQueue(&StatQueue{
		Tenant: "cgrates.org",
		ID:     "STS_PoccessCDR",
		SQMetrics: map[string]StatMetric{
			utils.MetaASR: &StatASR{
				Answered: 2,
				Count:    3,
				Events: map[string]*StatWithCompress{
					"cgrates.org:ev1": {Stat: 1},
				},
			},
		},
	})
	SetDataStorage(dm)
	if _, err := sts.processEvent(args); err != nil {
		t.Error(err)
	}

}
