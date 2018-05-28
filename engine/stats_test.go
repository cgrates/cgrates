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

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

var (
	cloneExpTimeStats time.Time
	expTimeStats      = time.Now().Add(time.Duration(20 * time.Minute))
	stsserv           StatService
	dmSTS             *DataManager
	sqps              = []*StatQueueProfile{
		&StatQueueProfile{
			Tenant:    "cgrates.org",
			ID:        "statsprofile1",
			FilterIDs: []string{"filter7"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			},
			QueueLength: 10,
			TTL:         time.Duration(10) * time.Second,
			Metrics: []*utils.MetricWithParams{
				&utils.MetricWithParams{
					MetricID:   utils.MetaSum,
					Parameters: utils.Usage,
				},
			},
			ThresholdIDs: []string{},
			Blocker:      true,
			Stored:       true,
			Weight:       20,
			MinItems:     1,
		},
		&StatQueueProfile{
			Tenant:    "cgrates.org",
			ID:        "statsprofile2",
			FilterIDs: []string{"filter8"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			},
			QueueLength: 10,
			TTL:         time.Duration(10) * time.Second,
			Metrics: []*utils.MetricWithParams{
				&utils.MetricWithParams{
					MetricID:   utils.MetaSum,
					Parameters: utils.Usage,
				},
			},
			ThresholdIDs: []string{},
			Blocker:      true,
			Stored:       true,
			Weight:       20,
			MinItems:     1,
		},
		&StatQueueProfile{
			Tenant:    "cgrates.org",
			ID:        "statsprofile3",
			FilterIDs: []string{"preffilter4"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			},
			QueueLength: 10,
			TTL:         time.Duration(10) * time.Second,
			Metrics: []*utils.MetricWithParams{
				&utils.MetricWithParams{
					MetricID:   utils.MetaSum,
					Parameters: utils.Usage,
				},
			},
			ThresholdIDs: []string{},
			Blocker:      true,
			Stored:       true,
			Weight:       20,
			MinItems:     1,
		},
		&StatQueueProfile{
			Tenant:    "cgrates.org",
			ID:        "statsprofile4",
			FilterIDs: []string{"defaultf4"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			},
			QueueLength: 10,
			TTL:         time.Duration(10) * time.Second,
			Metrics: []*utils.MetricWithParams{
				&utils.MetricWithParams{
					MetricID:   utils.MetaSum,
					Parameters: utils.Usage,
				},
			},
			ThresholdIDs: []string{},
			Blocker:      true,
			Stored:       true,
			Weight:       20,
			MinItems:     1,
		},
	}
	stqs = []*StatQueue{
		&StatQueue{Tenant: "cgrates.org", ID: "statsprofile1", sqPrfl: sqps[0]},
		&StatQueue{Tenant: "cgrates.org", ID: "statsprofile2", sqPrfl: sqps[1]},
		&StatQueue{Tenant: "cgrates.org", ID: "statsprofile3", sqPrfl: sqps[2]},
		&StatQueue{Tenant: "cgrates.org", ID: "statsprofile4", sqPrfl: sqps[3]},
	}
	statsEvs = []*utils.CGREvent{
		&utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "event1",
			Event: map[string]interface{}{
				"Stats":          "StatsProfile1",
				utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
				"UsageInterval":  "1s",
				"PddInterval":    "1s",
				"Weight":         "20.0",
				utils.Usage:      time.Duration(135 * time.Second),
				utils.COST:       123.0,
			}},
		&utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "event2",
			Event: map[string]interface{}{
				"Stats":          "StatsProfile2",
				utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
				"UsageInterval":  "1s",
				"PddInterval":    "1s",
				"Weight":         "21.0",
				utils.Usage:      time.Duration(45 * time.Second),
			}},
		&utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "event3",
			Event: map[string]interface{}{
				"Stats":     "StatsProfilePrefix",
				utils.Usage: time.Duration(30 * time.Second),
			}},
		&utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "event3",
			Event: map[string]interface{}{
				"Weight":    "200.0",
				utils.Usage: time.Duration(65 * time.Second),
			}},
	}
)

func TestStatsPopulateStatsService(t *testing.T) {
	data, _ := NewMapStorage()
	dmSTS = NewDataManager(data)
	var filters1 []*FilterRule
	var filters2 []*FilterRule
	var preffilter []*FilterRule
	var defaultf []*FilterRule
	second := 1 * time.Second
	stsserv = StatService{
		dm:      dmSTS,
		filterS: &FilterS{dm: dmSTS},
	}
	ref := NewFilterIndexer(dmSTS, utils.StatQueueProfilePrefix, "cgrates.org")
	//filter1
	x, err := NewFilterRule(MetaString, "Stats", []string{"StatsProfile1"})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	filters1 = append(filters1, x)
	x, err = NewFilterRule(MetaGreaterOrEqual, "UsageInterval", []string{second.String()})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	filters1 = append(filters1, x)
	x, err = NewFilterRule(MetaGreaterOrEqual, utils.Usage, []string{second.String()})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	filters1 = append(filters1, x)
	x, err = NewFilterRule(MetaGreaterOrEqual, "Weight", []string{"9.0"})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	filters1 = append(filters1, x)
	filter7 := &Filter{Tenant: config.CgrConfig().DefaultTenant, ID: "filter7", Rules: filters1}
	dmSTS.SetFilter(filter7)
	ref.IndexTPFilter(FilterToTPFilter(filter7), "statsprofile1")
	//filter2
	x, err = NewFilterRule(MetaString, "Stats", []string{"StatsProfile2"})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	filters2 = append(filters2, x)
	x, err = NewFilterRule(MetaGreaterOrEqual, "PddInterval", []string{second.String()})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	filters2 = append(filters2, x)
	x, err = NewFilterRule(MetaGreaterOrEqual, utils.Usage, []string{second.String()})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	filters2 = append(filters2, x)
	x, err = NewFilterRule(MetaGreaterOrEqual, "Weight", []string{"15.0"})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	filters2 = append(filters2, x)
	filter8 := &Filter{Tenant: config.CgrConfig().DefaultTenant, ID: "filter8", Rules: filters2}
	dmSTS.SetFilter(filter8)
	ref.IndexTPFilter(FilterToTPFilter(filter8), "statsprofile2")
	//prefix filter
	x, err = NewFilterRule(MetaPrefix, "Stats", []string{"StatsProfilePrefix"})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	preffilter = append(preffilter, x)
	x, err = NewFilterRule(MetaGreaterOrEqual, utils.Usage, []string{second.String()})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	preffilter = append(preffilter, x)
	preffilter4 := &Filter{Tenant: config.CgrConfig().DefaultTenant, ID: "preffilter4", Rules: preffilter}
	dmSTS.SetFilter(preffilter4)
	ref.IndexTPFilter(FilterToTPFilter(preffilter4), "statsprofile3")
	//default filter
	x, err = NewFilterRule(MetaGreaterOrEqual, "Weight", []string{"200.00"})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	defaultf = append(defaultf, x)
	x, err = NewFilterRule(MetaGreaterOrEqual, utils.Usage, []string{second.String()})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	defaultf = append(defaultf, x)
	defaultf4 := &Filter{Tenant: config.CgrConfig().DefaultTenant, ID: "defaultf4", Rules: defaultf}
	dmSTS.SetFilter(defaultf4)
	ref.IndexTPFilter(FilterToTPFilter(defaultf4), "statsprofile4")
	for _, stq := range stqs {
		dmSTS.SetStatQueue(stq)
	}
	for _, sqp := range sqps {
		dmSTS.SetStatQueueProfile(sqp, false)
	}
	err = ref.StoreIndexes(true, utils.NonTransactional)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
}

func TestStatsmatchingStatQueuesForEvent(t *testing.T) {
	msq, err := stsserv.matchingStatQueuesForEvent(statsEvs[0])
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
	msq, err = stsserv.matchingStatQueuesForEvent(statsEvs[1])
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
	msq, err = stsserv.matchingStatQueuesForEvent(statsEvs[2])
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
	msq, err = stsserv.matchingStatQueuesForEvent(statsEvs[3])
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(stqs[3].Tenant, msq[0].Tenant) {
		t.Errorf("Expecting: %+v, received: %+v", stqs[3].Tenant, msq[0].Tenant)
	} else if !reflect.DeepEqual(stqs[3].ID, msq[0].ID) {
		t.Errorf("Expecting: %+v, received: %+v", stqs[3].ID, msq[0].ID)
	} else if !reflect.DeepEqual(stqs[3].sqPrfl, msq[0].sqPrfl) {
		t.Errorf("Expecting: %+v, received: %+v", stqs[3].sqPrfl, msq[0].sqPrfl)
	}
}

func TestStatSprocessEvent(t *testing.T) {
	stq := map[string]string{}
	reply := []string{}
	expected := []string{"statsprofile1"}
	err := stsserv.V1ProcessEvent(statsEvs[0], &reply)
	if err != nil {
		t.Errorf("Error: %+v", err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expecting: %+v, received: %+v", expected, reply)
	}
	err = stsserv.V1GetQueueStringMetrics(&utils.TenantID{Tenant: stqs[0].Tenant, ID: stqs[0].ID}, &stq)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	expected = []string{"statsprofile2"}
	err = stsserv.V1ProcessEvent(statsEvs[1], &reply)
	if err != nil {
		t.Errorf("Error: %+v", err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expecting: %+v, received: %+v", expected, reply)
	}
	err = stsserv.V1GetQueueStringMetrics(&utils.TenantID{Tenant: stqs[1].Tenant, ID: stqs[1].ID}, &stq)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	expected = []string{"statsprofile3"}
	err = stsserv.V1ProcessEvent(statsEvs[2], &reply)
	if err != nil {
		t.Errorf("Error: %+v", err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expecting: %+v, received: %+v", expected, reply)
	}
	err = stsserv.V1GetQueueStringMetrics(&utils.TenantID{Tenant: stqs[2].Tenant, ID: stqs[2].ID}, &stq)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	expected = []string{"statsprofile4"}
	err = stsserv.V1ProcessEvent(statsEvs[3], &reply)
	if err != nil {
		t.Errorf("Error: %+v", err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expecting: %+v, received: %+v", expected, reply)
	}
	err = stsserv.V1GetQueueStringMetrics(&utils.TenantID{Tenant: stqs[3].Tenant, ID: stqs[3].ID}, &stq)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
}
