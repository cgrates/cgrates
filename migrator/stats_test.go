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
package migrator

import (
	"reflect"
	"testing"

	"github.com/cgrates/cgrates/engine"
)

/*
func TestV1StatsAsStats(t *testing.T) {
	var filters []*engine.FilterRule
	v1Sts := &v1Stat{
		Id:              "test",      // Config id, unique per config instance
		QueueLength:     10,          // Number of items in the stats buffer
		TimeWindow:      time.Second, // Will only keep the CDRs who's call setup time is not older than time.Now()-TimeWindow
		SaveInterval:    time.Second,
		Metrics:         []string{"ASR", "ACD", "ACC"},
		SetupInterval:   []time.Time{time.Now()},
		ToR:             []string{},
		CdrHost:         []string{},
		CdrSource:       []string{},
		ReqType:         []string{},
		Direction:       []string{},
		Tenant:          []string{},
		Category:        []string{},
		Account:         []string{},
		Subject:         []string{},
		DestinationIds:  []string{},
		UsageInterval:   []time.Duration{time.Second},
		PddInterval:     []time.Duration{time.Second},
		Supplier:        []string{},
		DisconnectCause: []string{},
		MediationRunIds: []string{},
		RatedAccount:    []string{},
		RatedSubject:    []string{},
		CostInterval:    []float64{},
	}

	x, _ := engine.NewFilterRule(utils.MetaGreaterOrEqual, "SetupInterval", []string{v1Sts.SetupInterval[0].String()})
	filters = append(filters, x)
	x, _ = engine.NewFilterRule(utils.MetaGreaterOrEqual, "UsageInterval", []string{v1Sts.UsageInterval[0].String()})
	filters = append(filters, x)
	x, _ = engine.NewFilterRule(utils.MetaGreaterOrEqual, "PddInterval", []string{v1Sts.PddInterval[0].String()})
	filters = append(filters, x)

	filter := &engine.Filter{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     v1Sts.Id,
		Rules:  filters}

	sqp := &engine.StatQueueProfile{
		Tenant:      "cgrates.org",
		ID:          "test",
		FilterIDs:   []string{v1Sts.Id},
		QueueLength: 10,
		TTL:         0,
		Metrics: []*engine.MetricWithFilters{
			{
				MetricID: "*asr",
			},
			{
				MetricID: utils.MetaACD,
			},
			{
				MetricID: "*acc",
			},
		},
		Blocker:      false,
		ThresholdIDs: []string{"TestB"},
		Stored:       true,
		Weight:       float64(0),
		MinItems:     0,
	}
	fltr, _, newsqp, err := v1Sts.AsStatQP()
	if err != nil {
		t.Errorf("err")
	}
	if !reflect.DeepEqual(sqp.Tenant, newsqp.Tenant) {
		t.Errorf("Expecting: %+v, received: %+v", sqp.Tenant, newsqp.Tenant)
	}
	if !reflect.DeepEqual(sqp.ID, newsqp.ID) {
		t.Errorf("Expecting: %+v, received: %+v", sqp.ID, newsqp.ID)
	}
	if !reflect.DeepEqual(sqp.FilterIDs, newsqp.FilterIDs) {
		t.Errorf("Expecting: %+v, received: %+v", sqp.FilterIDs, newsqp.FilterIDs)
	}
	if !reflect.DeepEqual(sqp.QueueLength, newsqp.QueueLength) {
		t.Errorf("Expecting: %+v, received: %+v", sqp.QueueLength, newsqp.QueueLength)
	}
	if !reflect.DeepEqual(sqp.TTL, newsqp.TTL) {
		t.Errorf("Expecting: %+v, received: %+v", sqp.TTL, newsqp.TTL)
	}
	if !reflect.DeepEqual(sqp.Metrics, newsqp.Metrics) {
		t.Errorf("Expecting: %+v, received: %+v", sqp.Metrics, newsqp.Metrics)
	}
	if !reflect.DeepEqual(sqp.ThresholdIDs, newsqp.ThresholdIDs) {
		t.Errorf("Expecting: %+v, received: %+v", sqp.ThresholdIDs, newsqp.ThresholdIDs)
	}
	if !reflect.DeepEqual(sqp.Blocker, newsqp.Blocker) {
		t.Errorf("Expecting: %+v, received: %+v", sqp.Blocker, newsqp.Blocker)
	}
	if !reflect.DeepEqual(sqp.Stored, newsqp.Stored) {
		t.Errorf("Expecting: %+v, received: %+v", sqp.Stored, newsqp.Stored)
	}
	if !reflect.DeepEqual(sqp.Weight, newsqp.Weight) {
		t.Errorf("Expecting: %+v, received: %+v", sqp.Weight, newsqp.Weight)
	}
	if !reflect.DeepEqual(sqp, newsqp) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(sqp), utils.ToJSON(newsqp))
	}
	if !reflect.DeepEqual(filter, fltr) {
		t.Errorf("Expecting: %+v, received: %+v", filter, fltr)
	}
}

*/

func TestRemakeQueue(t *testing.T) {
	sq := &engine.StatQueue{
		Tenant: "cgrates.org",
		ID:     "StatsID",
		SQItems: []engine.SQItem{{
			EventID: "ev1",
		}},
		SQMetrics: map[string]engine.StatMetric{
			"*tcc":                nil,
			"*sum#~*req.Usage":    nil,
			"*average#~*req.Cost": nil,
		},
	}
	expected := &engine.StatQueue{
		Tenant:  sq.Tenant,
		ID:      sq.ID,
		SQItems: sq.SQItems,
		SQMetrics: map[string]engine.StatMetric{
			"*tcc":                nil,
			"*sum#~*req.Usage":    nil,
			"*average#~*req.Cost": nil,
		},
	}

	if rply := remakeQueue(sq); !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting: %+v, received: %+v", expected, rply)
	}
	return
}
