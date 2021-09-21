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
	"log"
	"os"
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

func TestStatQueuesMatchingStatQueuesForEvent(t *testing.T) {
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
		dmSTS.SetStatQueue(context.TODO(), statQueue)
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
	msq, err := statService.matchingStatQueuesForEvent(context.TODO(), statsEvs[0].Tenant, nil,
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
	msq, err = statService.matchingStatQueuesForEvent(context.TODO(), statsEvs[1].Tenant, nil,
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
	msq, err = statService.matchingStatQueuesForEvent(context.TODO(), statsEvs[2].Tenant, nil,
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
		dmSTS.SetStatQueue(context.TODO(), statQueue)
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
		dmSTS.SetStatQueue(context.TODO(), statQueue)
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
	msq, err := statService.matchingStatQueuesForEvent(context.TODO(), statsEvs[0].Tenant, nil,
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
	msq, err = statService.matchingStatQueuesForEvent(context.TODO(), statsEvs[1].Tenant, nil,
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
	msq, err = statService.matchingStatQueuesForEvent(context.TODO(), statsEvs[2].Tenant, nil,
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
		dmSTS.SetStatQueue(context.TODO(), statQueue)
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
	if err := dmSTS.SetStatQueue(context.TODO(), sq); err != nil {
		t.Error(err)
	}
	if tempStat, err := dmSTS.GetStatQueueProfile(context.TODO(), sqPrf.Tenant,
		sqPrf.ID, true, false, utils.NonTransactional); err != nil {
		t.Errorf("Error: %+v", err)
	} else if !reflect.DeepEqual(sqPrf, tempStat) {
		t.Errorf("Expecting: %+v, received: %+v", sqPrf, tempStat)
	}
	statsEvs[0].CGREvent.APIOpts = make(map[string]interface{})
	statsEvs[0].CGREvent.APIOpts[utils.OptsStatsStatIDs] = []string{"StatQueueProfile1", "StatQueueProfile2", "StatQueueProfile3"}
	ev := &StatsArgsProcessEvent{
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
	statsEvs[0].CGREvent.APIOpts = nil
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
	if th, err := dm.GetStatQueue(context.Background(), sqp.Tenant, sqp.ID, false, false, utils.NonTransactional); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(expTh, th) {
		t.Errorf("Expected: %s, received: %s", utils.ToJSON(expTh), utils.ToJSON(th))
	}

	if err := dm.SetStatQueue(context.Background(), sq); err != nil {
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
		dm.SetStatQueueProfile(context.Background(), rPrf, true)
		prfs = append(prfs, rPrf)
		ids.Add(rPrf.ID)
	}
	dm.RemoveStatQueue(context.Background(), "cgrates.org", "STS1")
	_, err := rS.matchingStatQueuesForEvent(context.Background(), "cgrates.org", ids.AsSlice(), utils.MapStorage{})
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
		dm.SetStatQueueProfile(context.Background(), rPrf, true)
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
	err = db.SetStatQueueProfileDrv(context.Background(), rPrf)
	if err != nil {
		t.Fatal(err)
	}
	prfs = append(prfs, rPrf)
	ids.Add(rPrf.ID)
	_, err := rS.matchingStatQueuesForEvent(context.Background(), "cgrates.org", ids.AsSlice(), utils.MapStorage{})
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
		dm.SetStatQueueProfile(context.Background(), rPrf, true)
		prfs = append(prfs, rPrf)
		ids.Add(rPrf.ID)
	}
	mres, err := rS.matchingStatQueuesForEvent(context.Background(), "cgrates.org", ids.AsSlice(), utils.MapStorage{})
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

func TestStatQueueMatchingStatQueuesForEventLocks3(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	prfs := make([]*StatQueueProfile, 0)
	tmp := Cache
	defer func() { Cache = tmp }()
	Cache = NewCacheS(cfg, nil, nil)
	db := &DataDBMock{
		GetStatQueueProfileDrvF: func(ctx *context.Context, tnt, id string) (*StatQueueProfile, error) {
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
	_, err := rS.matchingStatQueuesForEvent(context.Background(), "cgrates.org", ids.AsSlice(), utils.MapStorage{})
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
		dm.SetStatQueueProfile(context.Background(), rPrf, true)
		prfs = append(prfs, rPrf)
		ids.Add(rPrf.ID)
	}
	ids.Add("STS20")
	mres, err := rS.matchingStatQueuesForEvent(context.Background(), "cgrates.org", ids.AsSlice(), utils.MapStorage{})
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
	data := NewInternalDB(nil, nil, true)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	filterS := NewFilterS(cfg, nil, dm)
	sS := &StatService{
		dm:          dm,
		filterS:     filterS,
		stopBackup:  make(chan struct{}),
		loopStopped: make(chan struct{}, 1),
		cgrcfg:      cfg,
	}
	sS.loopStopped <- struct{}{}
	sS.Reload(context.Background())
	close(sS.stopBackup)
}

func TestStatQueueStartLoop(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.StatSCfg().StoreInterval = -1
	data := NewInternalDB(nil, nil, true)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	filterS := NewFilterS(cfg, nil, dm)
	sS := &StatService{
		dm:          dm,
		filterS:     filterS,
		stopBackup:  make(chan struct{}),
		loopStopped: make(chan struct{}, 1),
		cgrcfg:      cfg,
	}

	sS.StartLoop(context.Background())
	time.Sleep(10 * time.Millisecond)

	if len(sS.loopStopped) != 1 {
		t.Errorf("expected loopStopped field to have only one element, received: <%+v>", len(sS.loopStopped))
	}
}

func TestStatQueueShutdown(t *testing.T) {
	utils.Logger.SetLogLevel(6)
	utils.Logger.SetSyslog(nil)

	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()

	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, true)
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
	utils.Logger.SetLogLevel(0)
}

func TestStatQueueStoreStatsOK(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()

	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, true)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	sS := NewStatService(dm, cfg, nil, nil)

	exp := &StatQueue{
		dirty:  utils.BoolPointer(true),
		Tenant: "cgrates.org",
		ID:     "SQ1",
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
	defer func() {
		Cache = tmp
	}()

	utils.Logger.SetLogLevel(4)
	utils.Logger.SetSyslog(nil)

	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()

	cfg := config.NewDefaultCGRConfig()
	sS := NewStatService(nil, cfg, nil, nil)

	value := &StatQueue{
		dirty:  utils.BoolPointer(true),
		Tenant: "cgrates.org",
		ID:     "SQ1",
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

	utils.Logger.SetLogLevel(0)
	Cache.Remove(context.Background(), utils.CacheStatQueues, "SQ1", true, utils.NonTransactional)
}

func TestStatQueueStoreStatsCacheGetErr(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()

	utils.Logger.SetLogLevel(4)
	utils.Logger.SetSyslog(nil)

	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()

	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, true)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	sS := NewStatService(dm, cfg, nil, nil)

	value := &StatQueue{
		dirty:  utils.BoolPointer(true),
		Tenant: "cgrates.org",
		ID:     "SQ1",
	}

	Cache.SetWithoutReplicate(utils.CacheStatQueues, "SQ2", value, nil, true,
		utils.NonTransactional)
	sS.storedStatQueues.Add("SQ1")
	expLog := `[WARNING] <StatS> failed retrieving from cache stat queue with ID: SQ1`
	sS.storeStats(context.Background())

	if rcvLog := buf.String(); !strings.Contains(rcvLog, expLog) {
		t.Errorf("expected <%+v> \nto be included in: <%+v>", expLog, rcvLog)
	}

	utils.Logger.SetLogLevel(0)
	Cache.Remove(context.Background(), utils.CacheStatQueues, "SQ2", true, utils.NonTransactional)
}

func TestStatQueueStoreStatQueueCacheSetErr(t *testing.T) {
	utils.Logger.SetLogLevel(4)
	utils.Logger.SetSyslog(nil)

	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()

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
	data := NewInternalDB(nil, nil, true)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	connMgr = NewConnManager(cfg)
	Cache = NewCacheS(cfg, dm, nil)
	filterS := NewFilterS(cfg, nil, dm)
	sS := NewStatService(dm, cfg, filterS, connMgr)

	sq := &StatQueue{
		Tenant: "cgrates.org",
		ID:     "SQ1",
		dirty:  utils.BoolPointer(true),
	}

	expLog := `[WARNING] <StatS> failed caching StatQueue with ID: cgrates.org:SQ1, error: DISCONNECTED`
	if err := sS.StoreStatQueue(context.Background(), sq); err == nil ||
		err.Error() != utils.ErrDisconnected.Error() {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrDisconnected, err)
	} else if rcv := buf.String(); !strings.Contains(rcv, expLog) {
		t.Errorf("expected log <%+v> to be included in <%+v>", expLog, rcv)
	}

	utils.Logger.SetLogLevel(0)
}

func TestStatQueueStoreThresholdNilDirtyField(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, true)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	sS := NewStatService(dm, cfg, nil, nil)

	sq := &StatQueue{
		Tenant: "cgrates.org",
		ID:     "SQ1",
	}

	if err := sS.StoreStatQueue(context.Background(), sq); err != nil {
		t.Error(err)
	}
}

func TestStatQueueSetCloneable(t *testing.T) {
	args := &StatsArgsProcessEvent{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "EventTest",
			Event:  map[string]interface{}{},
			APIOpts: map[string]interface{}{
				utils.OptsStatsStatIDs: []string{"SQ_ID"},
			},
		},
		clnb: false,
	}

	exp := &StatsArgsProcessEvent{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "EventTest",
			Event:  map[string]interface{}{},
			APIOpts: map[string]interface{}{
				utils.OptsStatsStatIDs: []string{"SQ_ID"},
			},
		},
		clnb: true,
	}
	args.SetCloneable(true)

	if !reflect.DeepEqual(args, exp) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", exp, args)
	}
}

func TestStatQueueRPCCloneOK(t *testing.T) {
	args := &StatsArgsProcessEvent{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "EventTest",
			Event:  make(map[string]interface{}),
			APIOpts: map[string]interface{}{
				utils.OptsStatsStatIDs: []string{"SQ_ID"},
			},
		},
		clnb: true,
	}

	exp := &StatsArgsProcessEvent{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "EventTest",
			Event:  make(map[string]interface{}),
			APIOpts: map[string]interface{}{
				utils.OptsStatsStatIDs: []string{"SQ_ID"},
			},
		},
	}

	if out, err := args.RPCClone(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(out.(*StatsArgsProcessEvent), exp) {
		t.Errorf("expected: <%T>, \nreceived: <%T>",
			args, exp)
	}
}

func TestStatQueueRPCCloneNotCloneable(t *testing.T) {
	args := &StatsArgsProcessEvent{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "EventTest",
			Event:  make(map[string]interface{}),
			APIOpts: map[string]interface{}{
				utils.OptsStatsStatIDs: []string{"SQ_ID"},
			},
		},
		clnb: false,
	}

	exp := &StatsArgsProcessEvent{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "EventTest",
			Event:  make(map[string]interface{}),
			APIOpts: map[string]interface{}{
				utils.OptsStatsStatIDs: []string{"SQ_ID"},
			},
		},
	}

	if out, err := args.RPCClone(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(out.(*StatsArgsProcessEvent), exp) {
		t.Errorf("expected: <%T>, \nreceived: <%T>",
			args, exp)
	}
}

func TestStatQueueProcessEventOK(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, true)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	filterS := NewFilterS(cfg, nil, dm)
	sS := NewStatService(dm, cfg, filterS, nil)

	sqPrf := &StatQueueProfile{
		Tenant:       "cgrates.org",
		ID:           "SQ1",
		FilterIDs:    []string{"*string:~*req.Account:1001"},
		Weight:       10,
		Blocker:      true,
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
	}

	if err := dm.SetStatQueueProfile(context.Background(), sqPrf, true); err != nil {
		t.Error(err)
	}
	if err := dm.SetStatQueue(context.Background(), sq); err != nil {
		t.Error(err)
	}

	args := &StatsArgsProcessEvent{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "SqProcessEvent",
			Event: map[string]interface{}{
				utils.AccountField: "1001",
			},
			APIOpts: map[string]interface{}{
				utils.OptsStatsStatIDs: []string{"SQ1"},
			},
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
	data := NewInternalDB(nil, nil, true)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	filterS := NewFilterS(cfg, nil, dm)
	sS := NewStatService(dm, cfg, filterS, nil)

	sqPrf := &StatQueueProfile{
		Tenant:       "cgrates.org",
		ID:           "SQ1",
		FilterIDs:    []string{"*string:~*req.Account:1001"},
		Weight:       10,
		Blocker:      true,
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

	args := &StatsArgsProcessEvent{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "SqProcessEvent",
			Event: map[string]interface{}{
				utils.AccountField: "1002",
			},
		},
	}

	if _, err := sS.processEvent(context.Background(), args.Tenant, args); err == nil ||
		err != utils.ErrNotFound {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}
}

func TestStatQueueProcessEventProcessEventErr(t *testing.T) {
	utils.Logger.SetLogLevel(4)
	utils.Logger.SetSyslog(nil)

	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()

	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, true)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	filterS := NewFilterS(cfg, nil, dm)
	sS := NewStatService(dm, cfg, filterS, nil)

	sqPrf := &StatQueueProfile{
		Tenant:       "cgrates.org",
		ID:           "SQ1",
		FilterIDs:    []string{"*string:~*req.Account:1001"},
		Weight:       10,
		Blocker:      true,
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

	args := &StatsArgsProcessEvent{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "SqProcessEvent",
			Event: map[string]interface{}{
				utils.AccountField: "1001",
			},
			APIOpts: map[string]interface{}{
				utils.OptsStatsStatIDs: []string{"SQ1"},
			},
		},
	}

	expLog := `[WARNING] <StatS> Queue: cgrates.org:SQ1, ignoring event: cgrates.org:SqProcessEvent, error: NOT_FOUND:Usage`
	expIDs := []string{"SQ1"}
	if rcvIDs, err := sS.processEvent(context.Background(), args.Tenant, args); err == nil ||
		err.Error() != utils.ErrPartiallyExecuted.Error() {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrPartiallyExecuted, err)
	} else if !reflect.DeepEqual(rcvIDs, expIDs) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", expIDs, rcvIDs)
	} else if rcvLog := buf.String(); !strings.Contains(rcvLog, expLog) {
		t.Errorf("expected log <%+v> to be included in: <%+v>",
			expLog, rcvLog)
	}

	utils.Logger.SetLogLevel(0)
}

func TestStatQueueV1ProcessEventProcessEventErr(t *testing.T) {
	tmp := Cache
	tmpC := config.CgrConfig()
	defer func() {
		Cache = tmp
		config.SetCgrConfig(tmpC)
	}()

	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, true)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	Cache = NewCacheS(cfg, dm, nil)
	filterS := NewFilterS(cfg, nil, dm)
	sS := NewStatService(dm, cfg, filterS, nil)

	sqPrf := &StatQueueProfile{
		Tenant:       "cgrates.org",
		ID:           "SQ1",
		FilterIDs:    []string{"*string:~*req.Account:1001"},
		Weight:       10,
		Blocker:      true,
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

	args := &StatsArgsProcessEvent{
		CGREvent: &utils.CGREvent{
			ID: "SqProcessEvent",
			Event: map[string]interface{}{
				utils.AccountField: "1001",
			},
			APIOpts: map[string]interface{}{
				utils.OptsStatsStatIDs: []string{"SQ1"},
			},
		},
	}

	var reply []string
	if err := sS.V1ProcessEvent(context.Background(), args, &reply); err == nil ||
		err.Error() != utils.ErrPartiallyExecuted.Error() {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrPartiallyExecuted, err)
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
	data := NewInternalDB(nil, nil, true)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	Cache = NewCacheS(cfg, dm, nil)
	filterS := NewFilterS(cfg, nil, dm)
	sS := NewStatService(dm, cfg, filterS, nil)

	sqPrf := &StatQueueProfile{
		Tenant:       "cgrates.org",
		ID:           "SQ1",
		FilterIDs:    []string{"*string:~*req.Account:1001"},
		Weight:       10,
		Blocker:      true,
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

	args := &StatsArgsProcessEvent{
		CGREvent: nil,
	}

	var reply []string
	experr := `MANDATORY_IE_MISSING: [CGREvent]`
	if err := sS.V1ProcessEvent(context.Background(), args, &reply); err == nil ||
		err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	args = &StatsArgsProcessEvent{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			Event: map[string]interface{}{
				utils.AccountField: "1001",
			},
			APIOpts: map[string]interface{}{
				utils.OptsStatsStatIDs: []string{"SQ1"},
			},
		},
	}

	experr = `MANDATORY_IE_MISSING: [ID]`
	if err := sS.V1ProcessEvent(context.Background(), args, &reply); err == nil ||
		err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	args = &StatsArgsProcessEvent{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "SqProcessEvent",
			Event:  nil,
			APIOpts: map[string]interface{}{
				utils.OptsStatsStatIDs: []string{"SQ1"},
			},
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
	data := NewInternalDB(nil, nil, true)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	Cache = NewCacheS(cfg, dm, nil)
	filterS := NewFilterS(cfg, nil, dm)
	sS := NewStatService(dm, cfg, filterS, nil)

	sqPrf := &StatQueueProfile{
		Tenant:       "cgrates.org",
		ID:           "SQ1",
		FilterIDs:    []string{"*string:~*req.Account:1001"},
		Weight:       10,
		Blocker:      true,
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
		sqPrfl: nil,
		Tenant: "testTenant",
		ID:     "SQ2",
	}
	sq3 := &StatQueue{
		sqPrfl: nil,
		Tenant: "cgrates.org",
		ID:     "SQ3",
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
	if err := sS.V1GetQueueIDs(context.Background(), utils.EmptyString, &qIDs); err != nil {
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
	Cache = NewCacheS(cfg, dm, nil)
	filterS := NewFilterS(cfg, nil, dm)
	sS := NewStatService(dm, cfg, filterS, nil)

	var qIDs []string
	if err := sS.V1GetQueueIDs(context.Background(), utils.EmptyString, &qIDs); err == nil ||
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
	data := NewInternalDB(nil, nil, true)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	Cache = NewCacheS(cfg, dm, nil)
	filterS := NewFilterS(cfg, nil, dm)
	sS := NewStatService(dm, cfg, filterS, nil)

	sqPrf := &StatQueueProfile{
		Tenant:       "cgrates.org",
		ID:           "SQ1",
		FilterIDs:    []string{"*string:~*req.Account:1001"},
		Weight:       10,
		Blocker:      true,
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
	data := NewInternalDB(nil, nil, true)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	Cache = NewCacheS(cfg, dm, nil)
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
	data := NewInternalDB(nil, nil, true)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	Cache = NewCacheS(cfg, dm, nil)
	filterS := NewFilterS(cfg, nil, dm)
	sS := NewStatService(dm, cfg, filterS, nil)

	sqPrf := &StatQueueProfile{
		Tenant:       "cgrates.org",
		ID:           "SQ1",
		FilterIDs:    []string{"*string:~*req.Account:1001"},
		Weight:       10,
		Blocker:      true,
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
	data := NewInternalDB(nil, nil, true)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	Cache = NewCacheS(cfg, dm, nil)
	filterS := NewFilterS(cfg, nil, dm)
	sS := NewStatService(dm, cfg, filterS, nil)

	sqPrf1 := &StatQueueProfile{
		Tenant:       "cgrates.org",
		ID:           "SQ1",
		FilterIDs:    []string{"*string:~*req.Account:1001"},
		Weight:       10,
		Blocker:      true,
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
		Tenant:       "cgrates.org",
		ID:           "SQ2",
		Weight:       20,
		Blocker:      false,
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

	args := &StatsArgsProcessEvent{
		CGREvent: &utils.CGREvent{
			ID: "TestGetStatQueuesForEvent",
			Event: map[string]interface{}{
				utils.AccountField: "1001",
			},
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
	data := NewInternalDB(nil, nil, true)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	Cache = NewCacheS(cfg, dm, nil)
	filterS := NewFilterS(cfg, nil, dm)
	sS := NewStatService(dm, cfg, filterS, nil)

	sqPrf := &StatQueueProfile{
		Tenant:       "cgrates.org",
		ID:           "SQ1",
		FilterIDs:    []string{"*string:~*req.Account:1001"},
		Weight:       10,
		Blocker:      true,
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

	args := &StatsArgsProcessEvent{
		CGREvent: &utils.CGREvent{
			ID: "TestGetStatQueuesForEvent",
			Event: map[string]interface{}{
				utils.AccountField: "1002",
			},
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
	data := NewInternalDB(nil, nil, true)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	Cache = NewCacheS(cfg, dm, nil)
	filterS := NewFilterS(cfg, nil, dm)
	sS := NewStatService(dm, cfg, filterS, nil)

	sqPrf := &StatQueueProfile{
		Tenant:       "cgrates.org",
		ID:           "SQ1",
		FilterIDs:    []string{"*string:~*req.Account:1001"},
		Weight:       10,
		Blocker:      true,
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

	args := &StatsArgsProcessEvent{
		CGREvent: nil,
	}

	experr := `MANDATORY_IE_MISSING: [CGREvent]`
	var reply []string
	if err := sS.V1GetStatQueuesForEvent(context.Background(), args, &reply); err == nil ||
		err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	args = &StatsArgsProcessEvent{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     utils.EmptyString,
			Event: map[string]interface{}{
				utils.AccountField: "1001",
			},
		},
	}

	experr = `MANDATORY_IE_MISSING: [ID]`
	if err := sS.V1GetStatQueuesForEvent(context.Background(), args, &reply); err == nil ||
		err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	args = &StatsArgsProcessEvent{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "TestGetStatQueuesForEvent",
			Event:  nil,
		},
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
	data := NewInternalDB(nil, nil, true)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	Cache = NewCacheS(cfg, dm, nil)
	filterS := NewFilterS(cfg, nil, dm)
	sS := NewStatService(dm, cfg, filterS, nil)

	sqPrf := &StatQueueProfile{
		Tenant:       "cgrates.org",
		ID:           "SQ1",
		FilterIDs:    []string{"*string:~*req.Account:1001"},
		Weight:       10,
		Blocker:      true,
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
				Sum: time.Minute,
				val: utils.DurationPointer(time.Hour),
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
				Events: make(map[string]*DurationWithCompress),
			},
		},
	}
	var reply string

	if err := sS.V1ResetStatQueue(context.Background(), &utils.TenantID{
		ID: "SQ1",
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
	data := NewInternalDB(nil, nil, true)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	Cache = NewCacheS(cfg, dm, nil)
	filterS := NewFilterS(cfg, nil, dm)
	sS := NewStatService(dm, cfg, filterS, nil)

	sqPrf := &StatQueueProfile{
		Tenant:       "cgrates.org",
		ID:           "SQ1",
		FilterIDs:    []string{"*string:~*req.Account:1001"},
		Weight:       10,
		Blocker:      true,
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
				Sum: time.Minute,
				val: utils.DurationPointer(time.Hour),
			},
		},
	}

	if err := dm.SetStatQueue(context.Background(), sq); err != nil {
		t.Error(err)
	}

	var reply string
	if err := sS.V1ResetStatQueue(context.Background(), &utils.TenantID{
		ID: "SQ2",
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
	data := NewInternalDB(nil, nil, true)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	Cache = NewCacheS(cfg, dm, nil)
	filterS := NewFilterS(cfg, nil, dm)
	sS := NewStatService(dm, cfg, filterS, nil)

	sqPrf := &StatQueueProfile{
		Tenant:       "cgrates.org",
		ID:           "SQ1",
		FilterIDs:    []string{"*string:~*req.Account:1001"},
		Weight:       10,
		Blocker:      true,
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
				Sum: time.Minute,
				val: utils.DurationPointer(time.Hour),
			},
		},
	}

	if err := dm.SetStatQueue(context.Background(), sq); err != nil {
		t.Error(err)
	}

	experr := `MANDATORY_IE_MISSING: [ID]`
	var reply string
	if err := sS.V1ResetStatQueue(context.Background(), &utils.TenantID{}, &reply); err == nil ||
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
	data := NewInternalDB(nil, nil, true)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	Cache = NewCacheS(cfg, dm, nil)
	filterS := NewFilterS(cfg, nil, dm)
	sS := NewStatService(dm, cfg, filterS, nil)

	sqPrf := &StatQueueProfile{
		Tenant:       "cgrates.org",
		ID:           "SQ1",
		FilterIDs:    []string{"*string:~*req.Account:1001"},
		Weight:       10,
		Blocker:      true,
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
				Sum: time.Minute,
				val: utils.DurationPointer(time.Hour),
			},
		},
	}

	if err := dm.SetStatQueue(context.Background(), sq); err != nil {
		t.Error(err)
	}

	experr := `unsupported metric type <testMetricType>`
	var reply string

	if err := sS.V1ResetStatQueue(context.Background(), &utils.TenantID{
		ID: "SQ1",
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
	data := NewInternalDB(nil, nil, true)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	Cache = NewCacheS(cfg, dm, nil)

	filterS := NewFilterS(cfg, nil, dm)
	sS := NewStatService(dm, cfg, filterS, nil)

	sqPrf := &StatQueueProfile{
		Tenant:       "cgrates.org",
		ID:           "SQ1",
		FilterIDs:    []string{"*string:~*req.Account:1001"},
		Weight:       10,
		Blocker:      true,
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
				Sum: time.Minute,
				val: utils.DurationPointer(time.Hour),
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
	data := NewInternalDB(nil, nil, true)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	Cache = NewCacheS(cfg, dm, nil)

	ccM := &ccMock{
		calls: map[string]func(ctx *context.Context, args interface{}, reply interface{}) error{
			utils.ThresholdSv1ProcessEvent: func(ctx *context.Context, args, reply interface{}) error {
				exp := &ThresholdsArgsProcessEvent{
					CGREvent: &utils.CGREvent{
						Tenant: "cgrates.org",
						ID:     args.(*ThresholdsArgsProcessEvent).CGREvent.ID,
						Event: map[string]interface{}{
							utils.EventType:  utils.StatUpdate,
							utils.StatID:     "SQ1",
							"testMetricType": time.Duration(time.Hour),
						},
						APIOpts: map[string]interface{}{
							utils.MetaEventType:              utils.StatUpdate,
							utils.OptsThresholdsThresholdIDs: []string{"TH1"},
						},
					},
				}
				if !reflect.DeepEqual(exp, args) {
					return fmt.Errorf("\nexpected: <%+v>, \nreceived: <%+v>",
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
		Tenant:       "cgrates.org",
		ID:           "SQ1",
		FilterIDs:    []string{"*string:~*req.Account:1001"},
		Weight:       10,
		Blocker:      true,
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
				Sum: time.Minute,
				val: utils.DurationPointer(time.Hour),
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

func TestStatQueueProcessThresholdsErrPartExec(t *testing.T) {
	tmp := Cache
	tmpC := config.CgrConfig()
	tmpCM := connMgr
	defer func() {
		Cache = tmp
		config.SetCgrConfig(tmpC)
		connMgr = tmpCM
	}()

	utils.Logger.SetLogLevel(4)
	utils.Logger.SetSyslog(nil)
	defer func() {
		utils.Logger.SetLogLevel(0)
	}()

	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()

	cfg := config.NewDefaultCGRConfig()
	cfg.StatSCfg().ThresholdSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds)}
	data := NewInternalDB(nil, nil, true)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	Cache = NewCacheS(cfg, dm, nil)

	ccM := &ccMock{
		calls: map[string]func(ctx *context.Context, args interface{}, reply interface{}) error{
			utils.ThresholdSv1ProcessEvent: func(ctx *context.Context, args, reply interface{}) error {
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
		Tenant:       "cgrates.org",
		ID:           "SQ1",
		FilterIDs:    []string{"*string:~*req.Account:1001"},
		Weight:       10,
		Blocker:      true,
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
				Sum: time.Minute,
				val: utils.DurationPointer(time.Hour),
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
	data := NewInternalDB(nil, nil, true)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	Cache = NewCacheS(cfg, dm, nil)
	filterS := NewFilterS(cfg, nil, dm)
	sS := NewStatService(dm, cfg, filterS, nil)

	sqPrf := &StatQueueProfile{
		Tenant:       "cgrates.org",
		ID:           "SQ1",
		FilterIDs:    []string{"*string:~*req.Account:1001"},
		Weight:       10,
		Blocker:      true,
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
				Sum: time.Minute,
				val: utils.DurationPointer(time.Hour),
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
	if err := sS.V1GetQueueFloatMetrics(context.Background(), &utils.TenantID{
		ID: "SQ1",
	}, &reply); err != nil {
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
	data := NewInternalDB(nil, nil, true)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	Cache = NewCacheS(cfg, dm, nil)
	filterS := NewFilterS(cfg, nil, dm)
	sS := NewStatService(dm, cfg, filterS, nil)

	sqPrf := &StatQueueProfile{
		Tenant:       "cgrates.org",
		ID:           "SQ1",
		FilterIDs:    []string{"*string:~*req.Account:1001"},
		Weight:       10,
		Blocker:      true,
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
				Sum: time.Minute,
				val: utils.DurationPointer(time.Hour),
			},
		},
	}

	if err := dm.SetStatQueue(context.Background(), sq); err != nil {
		t.Error(err)
	}

	reply := map[string]float64{}
	if err := sS.V1GetQueueFloatMetrics(context.Background(), &utils.TenantID{
		ID: "SQ2",
	}, &reply); err == nil || err != utils.ErrNotFound {
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
	data := NewInternalDB(nil, nil, true)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	Cache = NewCacheS(cfg, dm, nil)
	filterS := NewFilterS(cfg, nil, dm)
	sS := NewStatService(dm, cfg, filterS, nil)

	sqPrf := &StatQueueProfile{
		Tenant:       "cgrates.org",
		ID:           "SQ1",
		FilterIDs:    []string{"*string:~*req.Account:1001"},
		Weight:       10,
		Blocker:      true,
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
				Sum: time.Minute,
				val: utils.DurationPointer(time.Hour),
			},
		},
	}

	if err := dm.SetStatQueue(context.Background(), sq); err != nil {
		t.Error(err)
	}

	experr := `MANDATORY_IE_MISSING: [ID]`
	reply := map[string]float64{}
	if err := sS.V1GetQueueFloatMetrics(context.Background(), &utils.TenantID{}, &reply); err == nil ||
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
	Cache = NewCacheS(cfg, nil, nil)
	filterS := NewFilterS(cfg, nil, nil)
	sS := NewStatService(nil, cfg, filterS, nil)

	experr := `SERVER_ERROR: NO_DATABASE_CONNECTION`
	reply := map[string]float64{}
	if err := sS.V1GetQueueFloatMetrics(context.Background(), &utils.TenantID{
		ID: "SQ1",
	}, &reply); err == nil || err.Error() != experr {
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
	data := NewInternalDB(nil, nil, true)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	Cache = NewCacheS(cfg, dm, nil)
	filterS := NewFilterS(cfg, nil, dm)
	sS := NewStatService(dm, cfg, filterS, nil)

	sqPrf := &StatQueueProfile{
		Tenant:       "cgrates.org",
		ID:           "SQ1",
		FilterIDs:    []string{"*string:~*req.Account:1001"},
		Weight:       10,
		Blocker:      true,
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
				Sum: time.Minute,
				val: utils.DurationPointer(time.Hour),
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
	if err := sS.V1GetQueueStringMetrics(context.Background(), &utils.TenantID{
		ID: "SQ1",
	}, &reply); err != nil {
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
	data := NewInternalDB(nil, nil, true)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	Cache = NewCacheS(cfg, dm, nil)
	filterS := NewFilterS(cfg, nil, dm)
	sS := NewStatService(dm, cfg, filterS, nil)

	sqPrf := &StatQueueProfile{
		Tenant:       "cgrates.org",
		ID:           "SQ1",
		FilterIDs:    []string{"*string:~*req.Account:1001"},
		Weight:       10,
		Blocker:      true,
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
				Sum: time.Minute,
				val: utils.DurationPointer(time.Hour),
			},
		},
	}

	if err := dm.SetStatQueue(context.Background(), sq); err != nil {
		t.Error(err)
	}

	reply := map[string]string{}
	if err := sS.V1GetQueueStringMetrics(context.Background(), &utils.TenantID{
		ID: "SQ2",
	}, &reply); err == nil || err != utils.ErrNotFound {
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
	data := NewInternalDB(nil, nil, true)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	Cache = NewCacheS(cfg, dm, nil)
	filterS := NewFilterS(cfg, nil, dm)
	sS := NewStatService(dm, cfg, filterS, nil)

	sqPrf := &StatQueueProfile{
		Tenant:       "cgrates.org",
		ID:           "SQ1",
		FilterIDs:    []string{"*string:~*req.Account:1001"},
		Weight:       10,
		Blocker:      true,
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
				Sum: time.Minute,
				val: utils.DurationPointer(time.Hour),
			},
		},
	}

	if err := dm.SetStatQueue(context.Background(), sq); err != nil {
		t.Error(err)
	}

	experr := `MANDATORY_IE_MISSING: [ID]`
	reply := map[string]string{}
	if err := sS.V1GetQueueStringMetrics(context.Background(), &utils.TenantID{}, &reply); err == nil ||
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
	Cache = NewCacheS(cfg, nil, nil)
	filterS := NewFilterS(cfg, nil, nil)
	sS := NewStatService(nil, cfg, filterS, nil)

	experr := `SERVER_ERROR: NO_DATABASE_CONNECTION`
	reply := map[string]string{}
	if err := sS.V1GetQueueStringMetrics(context.Background(), &utils.TenantID{
		ID: "SQ1",
	}, &reply); err == nil || err.Error() != experr {
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
	data := NewInternalDB(nil, nil, true)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	connMgr = NewConnManager(cfg)
	Cache = NewCacheS(cfg, dm, nil)
	filterS := NewFilterS(cfg, nil, dm)
	sS := NewStatService(dm, cfg, filterS, connMgr)

	sq := &StatQueue{
		Tenant: "cgrates.org",
		ID:     "SQ1",
		dirty:  utils.BoolPointer(true),
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
	data := NewInternalDB(nil, nil, true)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	Cache = NewCacheS(cfg, dm, nil)
	filterS := NewFilterS(cfg, nil, dm)
	sS := NewStatService(dm, cfg, filterS, nil)

	sqPrf := &StatQueueProfile{
		Tenant:       "cgrates.org",
		ID:           "SQ1",
		FilterIDs:    []string{"*string:~*req.Account:1001"},
		Weight:       10,
		Blocker:      true,
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
				Sum: time.Minute,
				val: utils.DurationPointer(time.Hour),
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
