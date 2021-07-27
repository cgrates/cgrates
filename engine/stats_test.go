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
	"reflect"
	"testing"
	"time"

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
			Tenant:    "cgrates.org",
			ID:        "StatQueueProfile1",
			FilterIDs: []string{"FLTR_STATS_1"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			},
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
			Tenant:    "cgrates.org",
			ID:        "StatQueueProfile2",
			FilterIDs: []string{"FLTR_STATS_2"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			},
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
			Tenant:    "cgrates.org",
			ID:        "StatQueueProfilePrefix",
			FilterIDs: []string{"FLTR_STATS_3"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			},
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
	dmSTS.SetFilter(fltrSts1, true)
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
	dmSTS.SetFilter(fltrSts2, true)
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
	dmSTS.SetFilter(fltrSts3, true)
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
	msq, err := statService.matchingStatQueuesForEvent(statsEvs[0].Tenant, statsEvs[0].StatIDs, statsEvs[0].Time,
		utils.MapStorage{utils.MetaReq: statsEvs[0].Event, utils.MetaOpts: statsEvs[0].APIOpts})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	msq.unlock()
	if !reflect.DeepEqual(stqs[0].Tenant, msq[0].Tenant) {
		t.Errorf("Expecting: %+v, received: %+v", stqs[0].Tenant, msq[0].Tenant)
	} else if !reflect.DeepEqual(stqs[0].ID, msq[0].ID) {
		t.Errorf("Expecting: %+v, received: %+v", stqs[0].ID, msq[0].ID)
	} else if !reflect.DeepEqual(stqs[0].sqPrfl, msq[0].sqPrfl) {
		t.Errorf("Expecting: %+v, received: %+v", stqs[0].sqPrfl, msq[0].sqPrfl)
	}
	msq, err = statService.matchingStatQueuesForEvent(statsEvs[1].Tenant, statsEvs[1].StatIDs, statsEvs[1].Time,
		utils.MapStorage{utils.MetaReq: statsEvs[1].Event, utils.MetaOpts: statsEvs[1].APIOpts})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	msq.unlock()
	if !reflect.DeepEqual(stqs[1].Tenant, msq[0].Tenant) {
		t.Errorf("Expecting: %+v, received: %+v", stqs[1].Tenant, msq[0].Tenant)
	} else if !reflect.DeepEqual(stqs[1].ID, msq[0].ID) {
		t.Errorf("Expecting: %+v, received: %+v", stqs[1].ID, msq[0].ID)
	} else if !reflect.DeepEqual(stqs[1].sqPrfl, msq[0].sqPrfl) {
		t.Errorf("Expecting: %+v, received: %+v", stqs[1].sqPrfl, msq[0].sqPrfl)
	}
	msq, err = statService.matchingStatQueuesForEvent(statsEvs[2].Tenant, statsEvs[2].StatIDs, statsEvs[2].Time,
		utils.MapStorage{utils.MetaReq: statsEvs[2].Event, utils.MetaOpts: statsEvs[2].APIOpts})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	msq.unlock()
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
			Tenant:    "cgrates.org",
			ID:        "StatQueueProfile1",
			FilterIDs: []string{"FLTR_STATS_1"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			},
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
			Tenant:    "cgrates.org",
			ID:        "StatQueueProfile2",
			FilterIDs: []string{"FLTR_STATS_2"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			},
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
			Tenant:    "cgrates.org",
			ID:        "StatQueueProfilePrefix",
			FilterIDs: []string{"FLTR_STATS_3"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			},
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
	dmSTS.SetFilter(fltrSts1, true)
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
	dmSTS.SetFilter(fltrSts2, true)
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
	dmSTS.SetFilter(fltrSts3, true)
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
	var statService *StatService
	var dmSTS *DataManager
	sqps := []*StatQueueProfile{
		{
			Tenant:    "cgrates.org",
			ID:        "StatQueueProfile1",
			FilterIDs: []string{"FLTR_STATS_1"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			},
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
			Tenant:    "cgrates.org",
			ID:        "StatQueueProfile2",
			FilterIDs: []string{"FLTR_STATS_2"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			},
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
			Tenant:    "cgrates.org",
			ID:        "StatQueueProfilePrefix",
			FilterIDs: []string{"FLTR_STATS_3"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			},
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
	dmSTS.SetFilter(fltrSts1, true)
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
	dmSTS.SetFilter(fltrSts2, true)
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
	dmSTS.SetFilter(fltrSts3, true)
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
	statService.cgrcfg.StatSCfg().IndexedSelects = false
	msq, err := statService.matchingStatQueuesForEvent(statsEvs[0].Tenant, statsEvs[0].StatIDs, statsEvs[0].Time,
		utils.MapStorage{utils.MetaReq: statsEvs[0].Event, utils.MetaOpts: statsEvs[0].APIOpts})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	msq.unlock()
	if !reflect.DeepEqual(stqs[0].Tenant, msq[0].Tenant) {
		t.Errorf("Expecting: %+v, received: %+v", stqs[0].Tenant, msq[0].Tenant)
	} else if !reflect.DeepEqual(stqs[0].ID, msq[0].ID) {
		t.Errorf("Expecting: %+v, received: %+v", stqs[0].ID, msq[0].ID)
	} else if !reflect.DeepEqual(stqs[0].sqPrfl, msq[0].sqPrfl) {
		t.Errorf("Expecting: %+v, received: %+v", stqs[0].sqPrfl, msq[0].sqPrfl)
	}
	msq, err = statService.matchingStatQueuesForEvent(statsEvs[1].Tenant, statsEvs[1].StatIDs, statsEvs[1].Time,
		utils.MapStorage{utils.MetaReq: statsEvs[1].Event, utils.MetaOpts: statsEvs[1].APIOpts})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	msq.unlock()
	if !reflect.DeepEqual(stqs[1].Tenant, msq[0].Tenant) {
		t.Errorf("Expecting: %+v, received: %+v", stqs[1].Tenant, msq[0].Tenant)
	} else if !reflect.DeepEqual(stqs[1].ID, msq[0].ID) {
		t.Errorf("Expecting: %+v, received: %+v", stqs[1].ID, msq[0].ID)
	} else if !reflect.DeepEqual(stqs[1].sqPrfl, msq[0].sqPrfl) {
		t.Errorf("Expecting: %+v, received: %+v", stqs[1].sqPrfl, msq[0].sqPrfl)
	}
	msq, err = statService.matchingStatQueuesForEvent(statsEvs[2].Tenant, statsEvs[2].StatIDs, statsEvs[2].Time,
		utils.MapStorage{utils.MetaReq: statsEvs[2].Event, utils.MetaOpts: statsEvs[2].APIOpts})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	msq.unlock()
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
			Tenant:    "cgrates.org",
			ID:        "StatQueueProfile1",
			FilterIDs: []string{"FLTR_STATS_1"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			},
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
			Tenant:    "cgrates.org",
			ID:        "StatQueueProfile2",
			FilterIDs: []string{"FLTR_STATS_2"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			},
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
			Tenant:    "cgrates.org",
			ID:        "StatQueueProfilePrefix",
			FilterIDs: []string{"FLTR_STATS_3"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			},
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
	dmSTS.SetFilter(fltrSts1, true)
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
	dmSTS.SetFilter(fltrSts2, true)
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
	dmSTS.SetFilter(fltrSts3, true)
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
	sqPrf := &StatQueueProfile{
		Tenant:    "cgrates.org",
		ID:        "StatQueueProfile3",
		FilterIDs: []string{"FLTR_STATS_1"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
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

func TestStatQueuesUpdateStatQueue(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dm := NewDataManager(NewInternalDB(nil, nil, true), cfg.CacheCfg(), nil)
	sqp := &StatQueueProfile{
		Tenant:      "cgrates.org",
		ID:          "THUP1",
		Stored:      true,
		QueueLength: 1,
		Metrics:     []*MetricWithFilters{{MetricID: utils.MetaTCC}},
	}
	sqm, err := NewTCC(0, utils.EmptyString, []string{})
	if err != nil {
		t.Fatal(err)
	}
	if err = sqm.AddEvent("ev1", utils.MapStorage{utils.MetaReq: utils.MapStorage{utils.Cost: 10}}); err != nil {
		t.Fatal(err)
	}
	sq := &StatQueue{
		Tenant:    sqp.Tenant,
		ID:        sqp.ID,
		SQItems:   []SQItem{{EventID: "ev1"}},
		SQMetrics: map[string]StatMetric{utils.MetaTCC: sqm, utils.MetaTCD: sqm},
	}
	sqm2, err := NewTCC(0, utils.EmptyString, nil)
	if err != nil {
		t.Fatal(err)
	}
	expTh := &StatQueue{
		Tenant:    sqp.Tenant,
		ID:        sqp.ID,
		SQMetrics: map[string]StatMetric{utils.MetaTCC: sqm2},
	}

	if err := dm.SetStatQueueProfile(sqp, true); err != nil {
		t.Fatal(err)
	}

	if th, err := dm.GetStatQueue(sqp.Tenant, sqp.ID, false, false, utils.NonTransactional); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(expTh, th) {
		t.Errorf("Expected: %s, received: %s", utils.ToJSON(expTh), utils.ToJSON(th))
	}

	if err := dm.RemoveStatQueue(sqp.Tenant, sqp.ID); err != nil {
		t.Fatal(err)
	}
	if err := dm.SetStatQueueProfile(sqp, true); err != nil {
		t.Fatal(err)
	}

	if th, err := dm.GetStatQueue(sqp.Tenant, sqp.ID, false, false, utils.NonTransactional); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(expTh, th) {
		t.Errorf("Expected: %s, received: %s", utils.ToJSON(expTh), utils.ToJSON(th))
	}
	if err := dm.SetStatQueue(sq); err != nil {
		t.Fatal(err)
	}

	sqp = &StatQueueProfile{
		Tenant:      "cgrates.org",
		ID:          "THUP1",
		Stored:      true,
		QueueLength: 1,
		Metrics:     []*MetricWithFilters{{MetricID: utils.MetaTCC, FilterIDs: []string{"*string:~*req.Account:1001"}}},
	}

	if err := dm.SetStatQueueProfile(sqp, true); err != nil {
		t.Fatal(err)
	}

	sqm3, err := NewTCC(0, utils.EmptyString, []string{"*string:~*req.Account:1001"})
	if err != nil {
		t.Fatal(err)
	}
	expTh = &StatQueue{
		Tenant:    sqp.Tenant,
		ID:        sqp.ID,
		SQItems:   []SQItem{{EventID: "ev1"}},
		SQMetrics: map[string]StatMetric{utils.MetaTCC: sqm3},
	}
	if th, err := dm.GetStatQueue(sqp.Tenant, sqp.ID, false, false, utils.NonTransactional); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(expTh, th) {
		t.Errorf("Expected: %s, received: %s", utils.ToJSON(expTh), utils.ToJSON(th))
	}

	if err := dm.SetStatQueue(sq); err != nil {
		t.Fatal(err)
	}

	sqm2, err = NewTCC(5, utils.EmptyString, nil)
	if err != nil {
		t.Fatal(err)
	}
	expTh = &StatQueue{
		Tenant:    sqp.Tenant,
		ID:        sqp.ID,
		SQMetrics: map[string]StatMetric{utils.MetaTCC: sqm2},
	}
	delete(sq.SQMetrics, utils.MetaTCD)
	if err := dm.SetStatQueue(sq); err != nil {
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

	if err := dm.SetStatQueueProfile(sqp, true); err != nil {
		t.Fatal(err)
	}

	if th, err := dm.GetStatQueue(sqp.Tenant, sqp.ID, false, false, utils.NonTransactional); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(expTh, th) {
		t.Errorf("Expected: %s, received: %s", utils.ToJSON(expTh), utils.ToJSON(th))
	}
	if err := dm.SetStatQueue(sq); err != nil {
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

	if err := dm.SetStatQueueProfile(sqp, true); err != nil {
		t.Fatal(err)
	}

	if th, err := dm.GetStatQueue(sqp.Tenant, sqp.ID, false, false, utils.NonTransactional); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(expTh, th) {
		t.Errorf("Expected: %s, received: %s", utils.ToJSON(expTh), utils.ToJSON(th))
	}

	if err := dm.SetStatQueue(sq); err != nil {
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

	if err := dm.SetStatQueueProfile(sqp, true); err != nil {
		t.Fatal(err)
	}

	if th, err := dm.GetStatQueue(sqp.Tenant, sqp.ID, false, false, utils.NonTransactional); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(expTh, th) {
		t.Errorf("Expected: %s, received: %s", utils.ToJSON(expTh), utils.ToJSON(th))
	}

	if err := dm.SetStatQueue(sq); err != nil {
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

	if err := dm.SetStatQueueProfile(sqp, true); err != nil {
		t.Fatal(err)
	}

	if th, err := dm.GetStatQueue(sqp.Tenant, sqp.ID, false, false, utils.NonTransactional); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(expTh, th) {
		t.Errorf("Expected: %s, received: %s", utils.ToJSON(expTh), utils.ToJSON(th))
	}

	if err := dm.RemoveStatQueueProfile(sqp.Tenant, sqp.ID, true); err != nil {
		t.Fatal(err)
	}
	if _, err := dm.GetStatQueue(sqp.Tenant, sqp.ID, false, false, utils.NonTransactional); err != utils.ErrNotFound {
		t.Fatal(err)
	}
}

func TestStatQueueMatchingStatQueuesForEventLocks(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	tmp := Cache
	defer func() { Cache = tmp }()
	Cache = NewCacheS(cfg, nil, nil)
	db := NewInternalDB(nil, nil, true)
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
			Tenant:       "cgrates.org",
			ID:           fmt.Sprintf("STS%d", i),
			Weight:       20.00,
			ThresholdIDs: []string{utils.MetaNone},
			QueueLength:  1,
			Stored:       true,
		}
		dm.SetStatQueueProfile(rPrf, true)
		prfs = append(prfs, rPrf)
		ids.Add(rPrf.ID)
	}
	dm.RemoveStatQueue("cgrates.org", "STS1")
	_, err := rS.matchingStatQueuesForEvent("cgrates.org", ids.AsSlice(), nil, utils.MapStorage{})
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
		if r, err := dm.GetStatQueue(rPrf.Tenant, rPrf.ID, true, false, utils.NonTransactional); err != nil {
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
	Cache = NewCacheS(cfg, nil, nil)
	db := NewInternalDB(nil, nil, true)
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
			Tenant:       "cgrates.org",
			ID:           fmt.Sprintf("STS%d", i),
			QueueLength:  1,
			Stored:       true,
			Weight:       20.00,
			ThresholdIDs: []string{utils.MetaNone},
		}
		dm.SetStatQueueProfile(rPrf, true)
		prfs = append(prfs, rPrf)
		ids.Add(rPrf.ID)
	}
	rPrf := &StatQueueProfile{
		Tenant:       "cgrates.org",
		ID:           "STS20",
		FilterIDs:    []string{"FLTR_RES_201"},
		QueueLength:  1,
		Stored:       true,
		Weight:       20.00,
		ThresholdIDs: []string{utils.MetaNone},
	}
	err = db.SetStatQueueProfileDrv(rPrf)
	if err != nil {
		t.Fatal(err)
	}
	prfs = append(prfs, rPrf)
	ids.Add(rPrf.ID)
	_, err := rS.matchingStatQueuesForEvent("cgrates.org", ids.AsSlice(), nil, utils.MapStorage{})
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
		if r, err := dm.GetStatQueue(rPrf.Tenant, rPrf.ID, true, false, utils.NonTransactional); err != nil {
			t.Errorf("error %s for <%s>", err, rPrf.ID)
		} else if r.isLocked() {
			t.Fatalf("Expected StatQueue to not be locked %q", rPrf.ID)
		}
	}
}

func TestStatQueueMatchingStatQueuesForEventLocksBlocker(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	tmp := Cache
	defer func() { Cache = tmp }()
	Cache = NewCacheS(cfg, nil, nil)
	db := NewInternalDB(nil, nil, true)
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
			Tenant:       "cgrates.org",
			ID:           fmt.Sprintf("STS%d", i),
			QueueLength:  1,
			Stored:       true,
			Weight:       float64(10 - i),
			Blocker:      i == 4,
			ThresholdIDs: []string{utils.MetaNone},
		}
		dm.SetStatQueueProfile(rPrf, true)
		prfs = append(prfs, rPrf)
		ids.Add(rPrf.ID)
	}
	mres, err := rS.matchingStatQueuesForEvent("cgrates.org", ids.AsSlice(), nil, utils.MapStorage{})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	defer mres.unlock()
	if len(mres) != 5 {
		t.Fatal("Expected 6 StatQueues")
	}
	for _, rPrf := range prfs[5:] {
		if rPrf.isLocked() {
			t.Errorf("Expected profile to not be locked %q", rPrf.ID)
		}
		if r, err := dm.GetStatQueue(rPrf.Tenant, rPrf.ID, true, false, utils.NonTransactional); err != nil {
			t.Errorf("error %s for <%s>", err, rPrf.ID)
		} else if r.isLocked() {
			t.Fatalf("Expected StatQueue to not be locked %q", rPrf.ID)
		}
	}
	for _, rPrf := range prfs[:5] {
		if !rPrf.isLocked() {
			t.Errorf("Expected profile to be locked %q", rPrf.ID)
		}
		if r, err := dm.GetStatQueue(rPrf.Tenant, rPrf.ID, true, false, utils.NonTransactional); err != nil {
			t.Errorf("error %s for <%s>", err, rPrf.ID)
		} else if !r.isLocked() {
			t.Fatalf("Expected StatQueue to be locked %q", rPrf.ID)
		}
	}
}

func TestStatQueueMatchingStatQueuesForEventLocksActivationInterval(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	tmp := Cache
	defer func() { Cache = tmp }()
	Cache = NewCacheS(cfg, nil, nil)
	db := NewInternalDB(nil, nil, true)
	dm := NewDataManager(db, config.CgrConfig().CacheCfg(), nil)
	cfg.StatSCfg().StoreInterval = 1
	cfg.StatSCfg().StringIndexedFields = nil
	cfg.StatSCfg().PrefixIndexedFields = nil
	rS := NewStatService(dm, cfg,
		&FilterS{dm: dm, cfg: cfg}, nil)

	ids := utils.StringSet{}
	for i := 0; i < 10; i++ {
		rPrf := &StatQueueProfile{
			Tenant:       "cgrates.org",
			ID:           fmt.Sprintf("STS%d", i),
			QueueLength:  1,
			Stored:       true,
			Weight:       20.00,
			ThresholdIDs: []string{utils.MetaNone},
		}
		dm.SetStatQueueProfile(rPrf, true)
		ids.Add(rPrf.ID)
	}
	rPrf := &StatQueueProfile{
		Tenant:       "cgrates.org",
		ID:           "STS21",
		QueueLength:  1,
		Stored:       true,
		Weight:       20.00,
		ThresholdIDs: []string{utils.MetaNone},
		ActivationInterval: &utils.ActivationInterval{
			ExpiryTime: time.Now().Add(-5 * time.Second),
		},
	}
	dm.SetStatQueueProfile(rPrf, true)
	ids.Add(rPrf.ID)
	mres, err := rS.matchingStatQueuesForEvent("cgrates.org", ids.AsSlice(), utils.TimePointer(time.Now()), utils.MapStorage{})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	defer mres.unlock()
	if rPrf.isLocked() {
		t.Fatalf("Expected profile to not be locked %q", rPrf.ID)
	}
	if r, err := dm.GetStatQueue(rPrf.Tenant, rPrf.ID, true, false, utils.NonTransactional); err != nil {
		t.Errorf("error %s for <%s>", err, rPrf.ID)
	} else if r.isLocked() {
		t.Fatalf("Expected StatQueue to not be locked %q", rPrf.ID)
	}
}

func TestStatQueueMatchingStatQueuesForEventLocks3(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	prfs := make([]*StatQueueProfile, 0)
	tmp := Cache
	defer func() { Cache = tmp }()
	Cache = NewCacheS(cfg, nil, nil)
	db := &DataDBMock{
		GetStatQueueProfileDrvF: func(tnt, id string) (*StatQueueProfile, error) {
			if id == "STS1" {
				return nil, utils.ErrNotImplemented
			}
			rPrf := &StatQueueProfile{
				Tenant:       "cgrates.org",
				ID:           id,
				QueueLength:  1,
				Stored:       true,
				Weight:       20.00,
				ThresholdIDs: []string{utils.MetaNone},
			}
			Cache.Set(utils.CacheStatQueues, rPrf.TenantID(), &StatQueue{
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
	_, err := rS.matchingStatQueuesForEvent("cgrates.org", ids.AsSlice(), nil, utils.MapStorage{})
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
	Cache = NewCacheS(cfg, nil, nil)
	db := NewInternalDB(nil, nil, true)
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
			Tenant:       "cgrates.org",
			ID:           fmt.Sprintf("STS%d", i),
			Weight:       20.00,
			ThresholdIDs: []string{utils.MetaNone},
			QueueLength:  1,
			Stored:       true,
		}
		dm.SetStatQueueProfile(rPrf, true)
		prfs = append(prfs, rPrf)
		ids.Add(rPrf.ID)
	}
	ids.Add("STS20")
	mres, err := rS.matchingStatQueuesForEvent("cgrates.org", ids.AsSlice(), nil, utils.MapStorage{})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	defer mres.unlock()
	for _, rPrf := range prfs {
		if !rPrf.isLocked() {
			t.Errorf("Expected profile to be locked %q", rPrf.ID)
		}
		if r, err := dm.GetStatQueue(rPrf.Tenant, rPrf.ID, true, false, utils.NonTransactional); err != nil {
			t.Errorf("error %s for <%s>", err, rPrf.ID)
		} else if !r.isLocked() {
			t.Fatalf("Expected StatQueue to be locked %q", rPrf.ID)
		}
	}

}
