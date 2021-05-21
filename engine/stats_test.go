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
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

func TestNewStatService(t *testing.T) {
	testIntDB := NewInternalDB(nil, nil, true)
	testDM := NewDataManager(testIntDB, config.CgrConfig().CacheCfg(), nil)
	testCgrCfg := config.NewDefaultCGRConfig()
	testFltrS := &FilterS{dm: testDM, cfg: testCgrCfg}
	expStruct := &StatService{
		dm:               testDM,
		filterS:          testFltrS,
		cgrcfg:           testCgrCfg,
		storedStatQueues: make(utils.StringSet),
	}
	result := NewStatService(testDM, testCgrCfg, testFltrS, nil)
	if !reflect.DeepEqual(expStruct.dm, result.dm) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", expStruct.dm, result.dm)
	}
	if !reflect.DeepEqual(expStruct.filterS, result.filterS) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", expStruct.filterS, result.filterS)
	}
	if !reflect.DeepEqual(expStruct.cgrcfg, result.cgrcfg) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", expStruct.cgrcfg, result.cgrcfg)
	}
	if !reflect.DeepEqual(expStruct.storedStatQueues, expStruct.storedStatQueues) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", expStruct.storedStatQueues, expStruct.storedStatQueues)
	}
}

func TestMatchingStatQueuesForEvent(t *testing.T) {
	var statService *StatService
	var dmSTS *DataManager
	sqps := []*StatQueueProfile{
		{
			Tenant:      "cgrates.org",
			ID:          "StatQueueProfile1",
			FilterIDs:   []string{"FLTR_STATS_1", "*ai:*now:2014-07-14T14:25:00Z"},
			QueueLength: 10,
			TTL:         10 * time.Second,
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
			Tenant:      "cgrates.org",
			ID:          "StatQueueProfile2",
			FilterIDs:   []string{"FLTR_STATS_2", "*ai:*now:2014-07-14T14:25:00Z"},
			QueueLength: 10,
			TTL:         10 * time.Second,
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
			Tenant:      "cgrates.org",
			ID:          "StatQueueProfilePrefix",
			FilterIDs:   []string{"FLTR_STATS_3", "*ai:*now:2014-07-14T14:25:00Z"},
			QueueLength: 10,
			TTL:         10 * time.Second,
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
	stqs := []*StatQueue{
		{Tenant: "cgrates.org", ID: "StatQueueProfile1", sqPrfl: sqps[0], SQMetrics: make(map[string]StatMetric)},
		{Tenant: "cgrates.org", ID: "StatQueueProfile2", sqPrfl: sqps[1], SQMetrics: make(map[string]StatMetric)},
		{Tenant: "cgrates.org", ID: "StatQueueProfilePrefix", sqPrfl: sqps[2], SQMetrics: make(map[string]StatMetric)},
	}
	statsEvs := []*StatsArgsProcessEvent{
		{
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "event1",
				Event: map[string]interface{}{
					"Stats":          "StatQueueProfile1",
					utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
					"UsageInterval":  "1s",
					"PddInterval":    "1s",
					"Weight":         "9.0",
					utils.Usage:      135 * time.Second,
					utils.Cost:       123.0,
				},
			},
		},
		{
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "event2",
				Event: map[string]interface{}{
					"Stats":          "StatQueueProfile2",
					utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
					"UsageInterval":  "1s",
					"PddInterval":    "1s",
					"Weight":         "15.0",
					utils.Usage:      45 * time.Second,
				},
			},
		},
		{
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "event3",
				Event: map[string]interface{}{
					"Stats":     "StatQueueProfilePrefix",
					utils.Usage: 30 * time.Second,
				},
			},
		},
	}
	defaultCfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, true)
	dmSTS = NewDataManager(data, config.CgrConfig().CacheCfg(), nil)

	defaultCfg.StatSCfg().StoreInterval = 1
	defaultCfg.StatSCfg().StringIndexedFields = nil
	defaultCfg.StatSCfg().PrefixIndexedFields = nil
	statService = NewStatService(dmSTS, defaultCfg,
		&FilterS{dm: dmSTS, cfg: defaultCfg}, nil)

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
				Values:  []string{(time.Second).String()},
			},
			{
				Type:    utils.MetaGreaterOrEqual,
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Usage,
				Values:  []string{(time.Second).String()},
			},
			{
				Type:    utils.MetaGreaterOrEqual,
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Weight,
				Values:  []string{"9.0"},
			},
		},
	}
	dmSTS.SetFilter(context.Background(), fltrSts1, true)
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
				Values:  []string{(time.Second).String()},
			},
			{
				Type:    utils.MetaGreaterOrEqual,
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Usage,
				Values:  []string{(time.Second).String()},
			},
			{
				Type:    utils.MetaGreaterOrEqual,
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Weight,
				Values:  []string{"15.0"},
			},
		},
	}
	dmSTS.SetFilter(context.Background(), fltrSts2, true)
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
	dmSTS.SetFilter(context.Background(), fltrSts3, true)
	for _, statQueueProfile := range sqps {
		dmSTS.SetStatQueueProfile(context.TODO(), statQueueProfile, true)
	}
	for _, statQueue := range stqs {
		dmSTS.SetStatQueue(context.TODO(),statQueue)
	}
	//Test each statQueueProfile from cache
	for _, sqp := range sqps {
		if tempStat, err := dmSTS.GetStatQueueProfile(context.TODO(), sqp.Tenant,
			sqp.ID, true, false, utils.NonTransactional); err != nil {
			t.Errorf("Error: %+v", err)
		} else if !reflect.DeepEqual(sqp, tempStat) {
			t.Errorf("Expecting: %+v, received: %+v", sqp, tempStat)
		}
	}
	msq, err := statService.matchingStatQueuesForEvent(context.TODO(), statsEvs[0].Tenant, statsEvs[0].StatIDs,
		utils.MapStorage{utils.MetaReq: statsEvs[0].Event, utils.MetaOpts: statsEvs[0].APIOpts})
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
	msq, err = statService.matchingStatQueuesForEvent(context.TODO(), statsEvs[1].Tenant, statsEvs[1].StatIDs,
		utils.MapStorage{utils.MetaReq: statsEvs[1].Event, utils.MetaOpts: statsEvs[1].APIOpts})
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
	msq, err = statService.matchingStatQueuesForEvent(context.TODO(), statsEvs[2].Tenant, statsEvs[2].StatIDs,
		utils.MapStorage{utils.MetaReq: statsEvs[2].Event, utils.MetaOpts: statsEvs[2].APIOpts})
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
	var statService *StatService
	var dmSTS *DataManager
	sqps := []*StatQueueProfile{
		{
			Tenant:      "cgrates.org",
			ID:          "StatQueueProfile1",
			FilterIDs:   []string{"FLTR_STATS_1", "*ai:*now:2014-07-14T14:25:00Z"},
			QueueLength: 10,
			TTL:         10 * time.Second,
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
			Tenant:      "cgrates.org",
			ID:          "StatQueueProfile2",
			FilterIDs:   []string{"FLTR_STATS_2", "*ai:*now:2014-07-14T14:25:00Z"},
			QueueLength: 10,
			TTL:         10 * time.Second,
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
			Tenant:      "cgrates.org",
			ID:          "StatQueueProfilePrefix",
			FilterIDs:   []string{"FLTR_STATS_3", "*ai:*now:2014-07-14T14:25:00Z"},
			QueueLength: 10,
			TTL:         10 * time.Second,
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
	stqs := []*StatQueue{
		{Tenant: "cgrates.org", ID: "StatQueueProfile1", sqPrfl: sqps[0], SQMetrics: make(map[string]StatMetric)},
		{Tenant: "cgrates.org", ID: "StatQueueProfile2", sqPrfl: sqps[1], SQMetrics: make(map[string]StatMetric)},
		{Tenant: "cgrates.org", ID: "StatQueueProfilePrefix", sqPrfl: sqps[2], SQMetrics: make(map[string]StatMetric)},
	}
	statsEvs := []*StatsArgsProcessEvent{
		{
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "event1",
				Event: map[string]interface{}{
					"Stats":          "StatQueueProfile1",
					utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
					"UsageInterval":  "1s",
					"PddInterval":    "1s",
					"Weight":         "9.0",
					utils.Usage:      135 * time.Second,
					utils.Cost:       123.0,
				},
			},
		},
		{
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "event2",
				Event: map[string]interface{}{
					"Stats":          "StatQueueProfile2",
					utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
					"UsageInterval":  "1s",
					"PddInterval":    "1s",
					"Weight":         "15.0",
					utils.Usage:      45 * time.Second,
				},
			},
		},
		{
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "event3",
				Event: map[string]interface{}{
					"Stats":     "StatQueueProfilePrefix",
					utils.Usage: 30 * time.Second,
				},
			},
		},
	}
	defaultCfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, true)
	dmSTS = NewDataManager(data, config.CgrConfig().CacheCfg(), nil)

	defaultCfg.StatSCfg().StoreInterval = 1
	defaultCfg.StatSCfg().StringIndexedFields = nil
	defaultCfg.StatSCfg().PrefixIndexedFields = nil
	statService = NewStatService(dmSTS, defaultCfg,
		&FilterS{dm: dmSTS, cfg: defaultCfg}, nil)

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
				Values:  []string{(time.Second).String()},
			},
			{
				Type:    utils.MetaGreaterOrEqual,
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Usage,
				Values:  []string{(time.Second).String()},
			},
			{
				Type:    utils.MetaGreaterOrEqual,
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Weight,
				Values:  []string{"9.0"},
			},
		},
	}
	dmSTS.SetFilter(context.Background(), fltrSts1, true)
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
				Values:  []string{(time.Second).String()},
			},
			{
				Type:    utils.MetaGreaterOrEqual,
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Usage,
				Values:  []string{(time.Second).String()},
			},
			{
				Type:    utils.MetaGreaterOrEqual,
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Weight,
				Values:  []string{"15.0"},
			},
		},
	}
	dmSTS.SetFilter(context.Background(), fltrSts2, true)
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
	dmSTS.SetFilter(context.Background(), fltrSts3, true)
	for _, statQueueProfile := range sqps {
		dmSTS.SetStatQueueProfile(context.TODO(), statQueueProfile, true)
	}
	for _, statQueue := range stqs {
		dmSTS.SetStatQueue(context.TODO(),statQueue)
	}
	//Test each statQueueProfile from cache
	for _, sqp := range sqps {
		if tempStat, err := dmSTS.GetStatQueueProfile(context.TODO(), sqp.Tenant,
			sqp.ID, true, false, utils.NonTransactional); err != nil {
			t.Errorf("Error: %+v", err)
		} else if !reflect.DeepEqual(sqp, tempStat) {
			t.Errorf("Expecting: %+v, received: %+v", sqp, tempStat)
		}
	}
	stq := map[string]string{}
	reply := []string{}
	expected := []string{"StatQueueProfile1"}
	err := statService.V1ProcessEvent(context.TODO(), statsEvs[0], &reply)
	if err != nil {
		t.Errorf("Error: %+v", err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expecting: %+v, received: %+v", expected, reply)
	}
	err = statService.V1GetQueueStringMetrics(context.TODO(), &utils.TenantID{Tenant: stqs[0].Tenant, ID: stqs[0].ID}, &stq)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}

	expected = []string{"StatQueueProfile2"}
	err = statService.V1ProcessEvent(context.TODO(), statsEvs[1], &reply)
	if err != nil {
		t.Errorf("Error: %+v", err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expecting: %+v, received: %+v", expected, reply)
	}
	err = statService.V1GetQueueStringMetrics(context.TODO(), &utils.TenantID{Tenant: stqs[1].Tenant, ID: stqs[1].ID}, &stq)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}

	expected = []string{"StatQueueProfilePrefix"}
	err = statService.V1ProcessEvent(context.TODO(), statsEvs[2], &reply)
	if err != nil {
		t.Errorf("Error: %+v", err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expecting: %+v, received: %+v", expected, reply)
	}
	err = statService.V1GetQueueStringMetrics(context.TODO(), &utils.TenantID{Tenant: stqs[2].Tenant, ID: stqs[2].ID}, &stq)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
}

func TestStatQueuesMatchWithIndexFalse(t *testing.T) {
	var statService *StatService
	var dmSTS *DataManager
	sqps := []*StatQueueProfile{
		{
			Tenant:      "cgrates.org",
			ID:          "StatQueueProfile1",
			FilterIDs:   []string{"FLTR_STATS_1", "*ai:*now:2014-07-14T14:25:00Z"},
			QueueLength: 10,
			TTL:         10 * time.Second,
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
			Tenant:      "cgrates.org",
			ID:          "StatQueueProfile2",
			FilterIDs:   []string{"FLTR_STATS_2", "*ai:*now:2014-07-14T14:25:00Z"},
			QueueLength: 10,
			TTL:         10 * time.Second,
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
			Tenant:      "cgrates.org",
			ID:          "StatQueueProfilePrefix",
			FilterIDs:   []string{"FLTR_STATS_3", "*ai:*now:2014-07-14T14:25:00Z"},
			QueueLength: 10,
			TTL:         10 * time.Second,
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
	stqs := []*StatQueue{
		{Tenant: "cgrates.org", ID: "StatQueueProfile1", sqPrfl: sqps[0], SQMetrics: make(map[string]StatMetric)},
		{Tenant: "cgrates.org", ID: "StatQueueProfile2", sqPrfl: sqps[1], SQMetrics: make(map[string]StatMetric)},
		{Tenant: "cgrates.org", ID: "StatQueueProfilePrefix", sqPrfl: sqps[2], SQMetrics: make(map[string]StatMetric)},
	}
	statsEvs := []*StatsArgsProcessEvent{
		{
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "event1",
				Event: map[string]interface{}{
					"Stats":          "StatQueueProfile1",
					utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
					"UsageInterval":  "1s",
					"PddInterval":    "1s",
					"Weight":         "9.0",
					utils.Usage:      135 * time.Second,
					utils.Cost:       123.0,
				},
			},
		},
		{
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "event2",
				Event: map[string]interface{}{
					"Stats":          "StatQueueProfile2",
					utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
					"UsageInterval":  "1s",
					"PddInterval":    "1s",
					"Weight":         "15.0",
					utils.Usage:      45 * time.Second,
				},
			},
		},
		{
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "event3",
				Event: map[string]interface{}{
					"Stats":     "StatQueueProfilePrefix",
					utils.Usage: 30 * time.Second,
				},
			},
		},
	}
	defaultCfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, true)
	dmSTS = NewDataManager(data, config.CgrConfig().CacheCfg(), nil)

	defaultCfg.StatSCfg().StoreInterval = 1
	defaultCfg.StatSCfg().StringIndexedFields = nil
	defaultCfg.StatSCfg().PrefixIndexedFields = nil
	statService = NewStatService(dmSTS, defaultCfg,
		&FilterS{dm: dmSTS, cfg: defaultCfg}, nil)

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
				Values:  []string{(time.Second).String()},
			},
			{
				Type:    utils.MetaGreaterOrEqual,
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Usage,
				Values:  []string{(time.Second).String()},
			},
			{
				Type:    utils.MetaGreaterOrEqual,
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Weight,
				Values:  []string{"9.0"},
			},
		},
	}
	dmSTS.SetFilter(context.Background(), fltrSts1, true)
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
				Values:  []string{(time.Second).String()},
			},
			{
				Type:    utils.MetaGreaterOrEqual,
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Usage,
				Values:  []string{(time.Second).String()},
			},
			{
				Type:    utils.MetaGreaterOrEqual,
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Weight,
				Values:  []string{"15.0"},
			},
		},
	}
	dmSTS.SetFilter(context.Background(), fltrSts2, true)
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
	dmSTS.SetFilter(context.Background(), fltrSts3, true)
	for _, statQueueProfile := range sqps {
		dmSTS.SetStatQueueProfile(context.TODO(), statQueueProfile, true)
	}
	for _, statQueue := range stqs {
		dmSTS.SetStatQueue(context.TODO(),statQueue)
	}
	//Test each statQueueProfile from cache
	for _, sqp := range sqps {
		if tempStat, err := dmSTS.GetStatQueueProfile(context.TODO(), sqp.Tenant,
			sqp.ID, true, false, utils.NonTransactional); err != nil {
			t.Errorf("Error: %+v", err)
		} else if !reflect.DeepEqual(sqp, tempStat) {
			t.Errorf("Expecting: %+v, received: %+v", sqp, tempStat)
		}
	}
	statService.cgrcfg.StatSCfg().IndexedSelects = false
	msq, err := statService.matchingStatQueuesForEvent(context.TODO(), statsEvs[0].Tenant, statsEvs[0].StatIDs,
		utils.MapStorage{utils.MetaReq: statsEvs[0].Event, utils.MetaOpts: statsEvs[0].APIOpts})
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
	msq, err = statService.matchingStatQueuesForEvent(context.TODO(), statsEvs[1].Tenant, statsEvs[1].StatIDs,
		utils.MapStorage{utils.MetaReq: statsEvs[1].Event, utils.MetaOpts: statsEvs[1].APIOpts})
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
	msq, err = statService.matchingStatQueuesForEvent(context.TODO(), statsEvs[2].Tenant, statsEvs[2].StatIDs,
		utils.MapStorage{utils.MetaReq: statsEvs[2].Event, utils.MetaOpts: statsEvs[2].APIOpts})
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
	var statService *StatService
	var dmSTS *DataManager
	sqps := []*StatQueueProfile{
		{
			Tenant:      "cgrates.org",
			ID:          "StatQueueProfile1",
			FilterIDs:   []string{"FLTR_STATS_1", "*ai:~*req.AnswerTime:2014-07-14T14:25:00Z"},
			QueueLength: 10,
			TTL:         10 * time.Second,
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
			Tenant:      "cgrates.org",
			ID:          "StatQueueProfile2",
			FilterIDs:   []string{"FLTR_STATS_2", "*ai:~*req.AnswerTime:2014-07-14T14:25:00Z"},
			QueueLength: 10,
			TTL:         10 * time.Second,
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
			Tenant:      "cgrates.org",
			ID:          "StatQueueProfilePrefix",
			FilterIDs:   []string{"FLTR_STATS_3", "*ai:~*req.AnswerTime:2014-07-14T14:25:00Z"},
			QueueLength: 10,
			TTL:         10 * time.Second,
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
	stqs := []*StatQueue{
		{Tenant: "cgrates.org", ID: "StatQueueProfile1", sqPrfl: sqps[0], SQMetrics: make(map[string]StatMetric)},
		{Tenant: "cgrates.org", ID: "StatQueueProfile2", sqPrfl: sqps[1], SQMetrics: make(map[string]StatMetric)},
		{Tenant: "cgrates.org", ID: "StatQueueProfilePrefix", sqPrfl: sqps[2], SQMetrics: make(map[string]StatMetric)},
	}
	statsEvs := []*StatsArgsProcessEvent{
		{
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "event1",
				Event: map[string]interface{}{
					"Stats":          "StatQueueProfile1",
					utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
					"UsageInterval":  "1s",
					"PddInterval":    "1s",
					"Weight":         "9.0",
					utils.Usage:      135 * time.Second,
					utils.Cost:       123.0,
				},
			},
		},
		{
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "event2",
				Event: map[string]interface{}{
					"Stats":          "StatQueueProfile2",
					utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
					"UsageInterval":  "1s",
					"PddInterval":    "1s",
					"Weight":         "15.0",
					utils.Usage:      45 * time.Second,
				},
			},
		},
		{
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "event3",
				Event: map[string]interface{}{
					"Stats":     "StatQueueProfilePrefix",
					utils.Usage: 30 * time.Second,
				},
			},
		},
	}
	defaultCfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, true)
	dmSTS = NewDataManager(data, config.CgrConfig().CacheCfg(), nil)

	defaultCfg.StatSCfg().StoreInterval = 1
	defaultCfg.StatSCfg().StringIndexedFields = nil
	defaultCfg.StatSCfg().PrefixIndexedFields = nil
	statService = NewStatService(dmSTS, defaultCfg,
		&FilterS{dm: dmSTS, cfg: defaultCfg}, nil)

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
				Values:  []string{(time.Second).String()},
			},
			{
				Type:    utils.MetaGreaterOrEqual,
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Usage,
				Values:  []string{(time.Second).String()},
			},
			{
				Type:    utils.MetaGreaterOrEqual,
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Weight,
				Values:  []string{"9.0"},
			},
		},
	}
	dmSTS.SetFilter(context.Background(), fltrSts1, true)
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
				Values:  []string{(time.Second).String()},
			},
			{
				Type:    utils.MetaGreaterOrEqual,
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Usage,
				Values:  []string{(time.Second).String()},
			},
			{
				Type:    utils.MetaGreaterOrEqual,
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Weight,
				Values:  []string{"15.0"},
			},
		},
	}
	dmSTS.SetFilter(context.Background(), fltrSts2, true)
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
	dmSTS.SetFilter(context.Background(), fltrSts3, true)
	for _, statQueueProfile := range sqps {
		dmSTS.SetStatQueueProfile(context.TODO(), statQueueProfile, true)
	}
	for _, statQueue := range stqs {
		dmSTS.SetStatQueue(context.TODO(),statQueue)
	}
	//Test each statQueueProfile from cache
	for _, sqp := range sqps {
		if tempStat, err := dmSTS.GetStatQueueProfile(context.TODO(), sqp.Tenant,
			sqp.ID, true, false, utils.NonTransactional); err != nil {
			t.Errorf("Error: %+v", err)
		} else if !reflect.DeepEqual(sqp, tempStat) {
			t.Errorf("Expecting: %+v, received: %+v", sqp, tempStat)
		}
	}
	sqPrf := &StatQueueProfile{
		Tenant:      "cgrates.org",
		ID:          "StatQueueProfile3",
		FilterIDs:   []string{"FLTR_STATS_1", "*ai:~*req.AnswerTime:2014-07-14T14:25:00Z"},
		QueueLength: 10,
		TTL:         10 * time.Second,
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
	if err := dmSTS.SetStatQueueProfile(context.TODO(), sqPrf, true); err != nil {
		t.Error(err)
	}
	if err := dmSTS.SetStatQueue(context.TODO(),sq); err != nil {
		t.Error(err)
	}
	if tempStat, err := dmSTS.GetStatQueueProfile(context.TODO(), sqPrf.Tenant,
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
	if err := statService.V1ProcessEvent(context.TODO(), ev, &reply); err != nil {
		t.Errorf("Error: %+v", err)
	} else if !reflect.DeepEqual(reply, expected) && !reflect.DeepEqual(reply, expectedRev) {
		t.Errorf("Expecting: %+v, received: %+v", expected, reply)
	}
}
