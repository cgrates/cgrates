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

var sq *StatQueue

func TestStatQueuesSort(t *testing.T) {
	sInsts := StatQueues{
		&StatQueue{sqPrfl: &StatQueueProfile{ID: "FIRST", Weight: 30.0}},
		&StatQueue{sqPrfl: &StatQueueProfile{ID: "SECOND", Weight: 40.0}},
		&StatQueue{sqPrfl: &StatQueueProfile{ID: "THIRD", Weight: 30.0}},
		&StatQueue{sqPrfl: &StatQueueProfile{ID: "FOURTH", Weight: 35.0}},
	}
	sInsts.Sort()
	eSInst := StatQueues{
		&StatQueue{sqPrfl: &StatQueueProfile{ID: "SECOND", Weight: 40.0}},
		&StatQueue{sqPrfl: &StatQueueProfile{ID: "FOURTH", Weight: 35.0}},
		&StatQueue{sqPrfl: &StatQueueProfile{ID: "FIRST", Weight: 30.0}},
		&StatQueue{sqPrfl: &StatQueueProfile{ID: "THIRD", Weight: 30.0}},
	}
	if !reflect.DeepEqual(eSInst, sInsts) {
		t.Errorf("expecting: %+v, received: %+v", eSInst, sInsts)
	}
}

func TestStatRemEventWithID(t *testing.T) {
	sq = &StatQueue{
		SQMetrics: map[string]StatMetric{
			utils.MetaASR: &StatASR{
				Answered: 1,
				Count:    2,
				Events: map[string]*StatWithCompress{
					"cgrates.org:TestRemEventWithID_1": {Stat: 1, CompressFactor: 1},
					"cgrates.org:TestRemEventWithID_2": {Stat: 0, CompressFactor: 1},
				},
			},
		},
	}
	asrMetric := sq.SQMetrics[utils.MetaASR].(*StatASR)
	if asr := asrMetric.GetFloat64Value(config.CgrConfig().GeneralCfg().RoundingDecimals); asr != 50 {
		t.Errorf("received asrMetric: %v", asrMetric)
	}
	sq.remEventWithID("cgrates.org:TestRemEventWithID_1")
	if asr := asrMetric.GetFloat64Value(config.CgrConfig().GeneralCfg().RoundingDecimals); asr != 0 {
		t.Errorf("received asrMetric: %v", asrMetric)
	} else if len(asrMetric.Events) != 1 {
		t.Errorf("unexpected Events in asrMetric: %+v", asrMetric.Events)
	}
	sq.remEventWithID("cgrates.org:TestRemEventWithID_5") // non existent
	if asr := asrMetric.GetFloat64Value(config.CgrConfig().GeneralCfg().RoundingDecimals); asr != 0 {
		t.Errorf("received asrMetric: %v", asrMetric)
	} else if len(asrMetric.Events) != 1 {
		t.Errorf("unexpected Events in asrMetric: %+v", asrMetric.Events)
	}
	sq.remEventWithID("cgrates.org:TestRemEventWithID_2")
	if asr := asrMetric.GetFloat64Value(config.CgrConfig().GeneralCfg().RoundingDecimals); asr != -1 {
		t.Errorf("received asrMetric: %v", asrMetric)
	} else if len(asrMetric.Events) != 0 {
		t.Errorf("unexpected Events in asrMetric: %+v", asrMetric.Events)
	}
	sq.remEventWithID("cgrates.org:TestRemEventWithID_2")
	if asr := asrMetric.GetFloat64Value(config.CgrConfig().GeneralCfg().RoundingDecimals); asr != -1 {
		t.Errorf("received asrMetric: %v", asrMetric)
	} else if len(asrMetric.Events) != 0 {
		t.Errorf("unexpected Events in asrMetric: %+v", asrMetric.Events)
	}
}

func TestStatRemEventWithID2(t *testing.T) {
	sq = &StatQueue{
		SQMetrics: map[string]StatMetric{
			utils.MetaASR: &StatASR{
				Answered: 2,
				Count:    4,
				Events: map[string]*StatWithCompress{
					"cgrates.org:TestRemEventWithID_1": {Stat: 1, CompressFactor: 2},
					"cgrates.org:TestRemEventWithID_2": {Stat: 0, CompressFactor: 2},
				},
			},
		},
	}
	asrMetric := sq.SQMetrics[utils.MetaASR].(*StatASR)
	if asr := asrMetric.GetFloat64Value(config.CgrConfig().GeneralCfg().RoundingDecimals); asr != 50 {
		t.Errorf("received asrMetric: %v", asrMetric)
	}
	sq.remEventWithID("cgrates.org:TestRemEventWithID_1")
	sq.remEventWithID("cgrates.org:TestRemEventWithID_2")
	if asr := asrMetric.GetFloat64Value(config.CgrConfig().GeneralCfg().RoundingDecimals); asr != 50 {
		t.Errorf("received asrMetric: %v", asrMetric)
	} else if len(asrMetric.Events) != 2 {
		t.Errorf("unexpected Events in asrMetric: %+v", asrMetric.Events)
	}
	sq.remEventWithID("cgrates.org:TestRemEventWithID_5") // non existent
	if asr := asrMetric.GetFloat64Value(config.CgrConfig().GeneralCfg().RoundingDecimals); asr != 50 {
		t.Errorf("received asrMetric: %v", asrMetric)
	} else if len(asrMetric.Events) != 2 {
		t.Errorf("unexpected Events in asrMetric: %+v", asrMetric.Events)
	}
	sq.remEventWithID("cgrates.org:TestRemEventWithID_2")
	sq.remEventWithID("cgrates.org:TestRemEventWithID_1")
	if asr := asrMetric.GetFloat64Value(config.CgrConfig().GeneralCfg().RoundingDecimals); asr != -1 {
		t.Errorf("received asrMetric: %v", asrMetric)
	} else if len(asrMetric.Events) != 0 {
		t.Errorf("unexpected Events in asrMetric: %+v", asrMetric.Events)
	}
	sq.remEventWithID("cgrates.org:TestRemEventWithID_2")
	if asr := asrMetric.GetFloat64Value(config.CgrConfig().GeneralCfg().RoundingDecimals); asr != -1 {
		t.Errorf("received asrMetric: %v", asrMetric)
	} else if len(asrMetric.Events) != 0 {
		t.Errorf("unexpected Events in asrMetric: %+v", asrMetric.Events)
	}
}

func TestStatRemExpired(t *testing.T) {
	sq = &StatQueue{
		SQMetrics: map[string]StatMetric{
			utils.MetaASR: &StatASR{
				Answered: 2,
				Count:    3,
				Events: map[string]*StatWithCompress{
					"cgrates.org:TestStatRemExpired_1": {Stat: 1, CompressFactor: 1},
					"cgrates.org:TestStatRemExpired_2": {Stat: 0, CompressFactor: 1},
					"cgrates.org:TestStatRemExpired_3": {Stat: 1, CompressFactor: 1},
				},
			},
		},
		SQItems: []SQItem{
			{"cgrates.org:TestStatRemExpired_1", utils.TimePointer(time.Now())},
			{"cgrates.org:TestStatRemExpired_2", utils.TimePointer(time.Now())},
			{"cgrates.org:TestStatRemExpired_3", utils.TimePointer(time.Now().Add(time.Minute))},
		},
	}
	asrMetric := sq.SQMetrics[utils.MetaASR].(*StatASR)
	if asr := asrMetric.GetFloat64Value(config.CgrConfig().GeneralCfg().RoundingDecimals); asr != 66.66667 {
		t.Errorf("received asrMetric: %v", asrMetric)
	}
	sq.remExpired()
	if asr := asrMetric.GetFloat64Value(config.CgrConfig().GeneralCfg().RoundingDecimals); asr != 100 {
		t.Errorf("received asrMetric: %v", asrMetric)
	} else if len(asrMetric.Events) != 1 {
		t.Errorf("unexpected Events in asrMetric: %+v", asrMetric.Events)
	}
	if len(sq.SQItems) != 1 {
		t.Errorf("Unexpected items: %+v", sq.SQItems)
	}
}

func TestStatRemOnQueueLength(t *testing.T) {
	sq = &StatQueue{
		sqPrfl: &StatQueueProfile{
			QueueLength: 2,
		},
		SQItems: []SQItem{
			{"cgrates.org:TestStatRemExpired_1", nil},
		},
	}
	sq.remOnQueueLength()
	if len(sq.SQItems) != 1 {
		t.Errorf("wrong items: %+v", sq.SQItems)
	}
	sq.SQItems = []SQItem{
		{"cgrates.org:TestStatRemExpired_1", nil},
		{"cgrates.org:TestStatRemExpired_2", nil},
	}
	sq.remOnQueueLength()
	if len(sq.SQItems) != 1 {
		t.Errorf("wrong items: %+v", sq.SQItems)
	} else if sq.SQItems[0].EventID != "cgrates.org:TestStatRemExpired_2" {
		t.Errorf("wrong item in SQItems: %+v", sq.SQItems[0])
	}
	sq.sqPrfl.QueueLength = -1
	sq.SQItems = []SQItem{
		{"cgrates.org:TestStatRemExpired_1", nil},
		{"cgrates.org:TestStatRemExpired_2", nil},
		{"cgrates.org:TestStatRemExpired_3", nil},
	}
	sq.remOnQueueLength()
	if len(sq.SQItems) != 3 {
		t.Errorf("wrong items: %+v", sq.SQItems)
	}
}

func TestStatAddStatEvent(t *testing.T) {
	sq = &StatQueue{
		SQMetrics: map[string]StatMetric{
			utils.MetaASR: &StatASR{
				Answered: 1,
				Count:    1,
				Events: map[string]*StatWithCompress{
					"cgrates.org:TestStatRemExpired_1": {Stat: 1, CompressFactor: 1},
				},
			},
		},
	}
	asrMetric := sq.SQMetrics[utils.MetaASR].(*StatASR)
	if asr := asrMetric.GetFloat64Value(config.CgrConfig().GeneralCfg().RoundingDecimals); asr != 100 {
		t.Errorf("received ASR: %v", asr)
	}
	ev1 := &utils.CGREvent{Tenant: "cgrates.org", ID: "TestStatAddStatEvent_1"}
	sq.addStatEvent(ev1.Tenant, ev1.ID, nil, utils.MapStorage{utils.MetaReq: ev1.Event})
	if asr := asrMetric.GetFloat64Value(config.CgrConfig().GeneralCfg().RoundingDecimals); asr != 50 {
		t.Errorf("received ASR: %v", asr)
	} else if asrMetric.Answered != 1 || asrMetric.Count != 2 {
		t.Errorf("ASR: %v", asrMetric)
	}
	ev1.Event = map[string]interface{}{
		utils.AnswerTime: time.Now()}
	sq.addStatEvent(ev1.Tenant, ev1.ID, nil, utils.MapStorage{utils.MetaReq: ev1.Event})
	if asr := asrMetric.GetFloat64Value(config.CgrConfig().GeneralCfg().RoundingDecimals); asr != 66.66667 {
		t.Errorf("received ASR: %v", asr)
	} else if asrMetric.Answered != 2 || asrMetric.Count != 3 {
		t.Errorf("ASR: %v", asrMetric)
	}
}

func TestStatRemOnQueueLength2(t *testing.T) {
	sq = &StatQueue{
		sqPrfl: &StatQueueProfile{
			QueueLength: 2,
			FilterIDs:   []string{"*string:~Account:1001|1002"},
		},
		SQItems: []SQItem{
			{"cgrates.org:TestStatRemExpired_1", nil},
			{"cgrates.org:TestStatRemExpired_2", nil},
		},
		SQMetrics: map[string]StatMetric{
			utils.MetaTCD: &StatTCD{
				FilterIDs: []string{"*string:~*req.Account:1002"},
				Events: map[string]*DurationWithCompress{
					"cgrates.org:TestStatRemExpired_2": {Duration: time.Minute, CompressFactor: 1},
				},
			},
			utils.MetaASR: &StatASR{
				FilterIDs: []string{"*string:~*req.Account:1001"},
				Events: map[string]*StatWithCompress{
					"cgrates.org:TestStatRemExpired_1": {Stat: 1, CompressFactor: 1},
				},
			},
		},
	}
	sq.remOnQueueLength()
	if len(sq.SQItems) != 1 {
		t.Errorf("wrong items: %+v", utils.ToJSON(sq.SQItems))
	}
}

func TestStatCompress(t *testing.T) {
	asr := &StatASR{
		Answered: 2,
		Count:    4,
		Events: map[string]*StatWithCompress{
			"cgrates.org:TestStatRemExpired_1": {Stat: 1, CompressFactor: 1},
			"cgrates.org:TestStatRemExpired_2": {Stat: 0, CompressFactor: 1},
			"cgrates.org:TestStatRemExpired_3": {Stat: 1, CompressFactor: 1},
			"cgrates.org:TestStatRemExpired_4": {Stat: 0, CompressFactor: 1},
		},
	}
	expectedASR := &StatASR{
		Answered: 2,
		Count:    4,
		Events: map[string]*StatWithCompress{
			"cgrates.org:TestStatRemExpired_4": {Stat: 0.5, CompressFactor: 4},
		},
	}
	sqItems := []SQItem{
		{"cgrates.org:TestStatRemExpired_1", utils.TimePointer(time.Now())},
		{"cgrates.org:TestStatRemExpired_2", utils.TimePointer(time.Now().Add(time.Minute))},
		{"cgrates.org:TestStatRemExpired_3", utils.TimePointer(time.Now().Add(2 * time.Minute))},
		{"cgrates.org:TestStatRemExpired_4", nil},
	}
	expectedSqItems := []SQItem{
		{"cgrates.org:TestStatRemExpired_4", nil},
	}
	sq = &StatQueue{
		SQItems: sqItems,
		SQMetrics: map[string]StatMetric{
			utils.MetaASR: asr,
		},
	}
	if sq.Compress(int64(100), config.CgrConfig().GeneralCfg().RoundingDecimals) {
		t.Errorf("StatQueue compressed: %s", utils.ToJSON(sq))
	}
	if !sq.Compress(int64(2), config.CgrConfig().GeneralCfg().RoundingDecimals) {
		t.Errorf("StatQueue not compressed: %s", utils.ToJSON(sq))
	}
	if !reflect.DeepEqual(sq.SQItems, expectedSqItems) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expectedSqItems), utils.ToJSON(sq.SQItems))
	}
	if rply := sq.SQMetrics[utils.MetaASR].(*StatASR); !reflect.DeepEqual(*rply, *expectedASR) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expectedASR), utils.ToJSON(rply))
	}
}

func TestStatCompress2(t *testing.T) {
	asr := &StatASR{
		Answered: 2,
		Count:    4,
		Events: map[string]*StatWithCompress{
			"cgrates.org:TestStatRemExpired_1": {Stat: 1, CompressFactor: 1},
			"cgrates.org:TestStatRemExpired_2": {Stat: 0, CompressFactor: 1},
			"cgrates.org:TestStatRemExpired_3": {Stat: 1, CompressFactor: 1},
			"cgrates.org:TestStatRemExpired_4": {Stat: 0, CompressFactor: 1},
		},
	}
	expectedASR := &StatASR{
		Answered: 2,
		Count:    4,
		Events: map[string]*StatWithCompress{
			"cgrates.org:TestStatRemExpired_4": {Stat: 0.5, CompressFactor: 4},
		},
	}
	tcd := &StatTCD{
		Sum:   3 * time.Minute,
		Count: 2,
		Events: map[string]*DurationWithCompress{
			"cgrates.org:TestStatRemExpired_2": {Duration: time.Minute, CompressFactor: 1},
			"cgrates.org:TestStatRemExpired_3": {Duration: 2 * time.Minute, CompressFactor: 1},
		},
	}
	expectedTCD := &StatTCD{
		Sum:   3 * time.Minute,
		Count: 2,
		Events: map[string]*DurationWithCompress{
			"cgrates.org:TestStatRemExpired_4": {Duration: time.Minute + 30*time.Second, CompressFactor: 2},
		},
	}
	sqItems := []SQItem{
		{"cgrates.org:TestStatRemExpired_1", utils.TimePointer(time.Now())},
		{"cgrates.org:TestStatRemExpired_2", utils.TimePointer(time.Now().Add(time.Minute))},
		{"cgrates.org:TestStatRemExpired_3", utils.TimePointer(time.Now().Add(2 * time.Minute))},
		{"cgrates.org:TestStatRemExpired_4", nil},
	}
	expectedSqItems := []SQItem{
		{"cgrates.org:TestStatRemExpired_4", nil},
	}
	sq = &StatQueue{
		SQItems: sqItems,
		ttl:     utils.DurationPointer(time.Second),
		SQMetrics: map[string]StatMetric{
			utils.MetaASR: asr,
			utils.MetaTCD: tcd,
		},
	}
	if sq.Compress(int64(100), config.CgrConfig().GeneralCfg().RoundingDecimals) {
		t.Errorf("StatQueue compressed: %s", utils.ToJSON(sq))
	}
	if !sq.Compress(int64(2), config.CgrConfig().GeneralCfg().RoundingDecimals) {
		t.Errorf("StatQueue not compressed: %s", utils.ToJSON(sq))
	}
	if !reflect.DeepEqual(sq.SQItems, expectedSqItems) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expectedSqItems), utils.ToJSON(sq.SQItems))
	}
	if rply := sq.SQMetrics[utils.MetaASR].(*StatASR); !reflect.DeepEqual(*rply, *expectedASR) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expectedASR), utils.ToJSON(rply))
	}
	if rply := sq.SQMetrics[utils.MetaTCD].(*StatTCD); !reflect.DeepEqual(*rply, *expectedTCD) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expectedTCD), utils.ToJSON(rply))
	}
}

func TestStatCompress3(t *testing.T) {
	tmNow := time.Now()
	asr := &StatASR{
		Answered: 2,
		Count:    4,
		Events: map[string]*StatWithCompress{
			"cgrates.org:TestStatRemExpired_1": {Stat: 1, CompressFactor: 1},
			"cgrates.org:TestStatRemExpired_2": {Stat: 0, CompressFactor: 1},
			"cgrates.org:TestStatRemExpired_3": {Stat: 1, CompressFactor: 1},
			"cgrates.org:TestStatRemExpired_4": {Stat: 0, CompressFactor: 1},
		},
	}
	expectedASR := &StatASR{
		Answered: 2,
		Count:    4,
		Events: map[string]*StatWithCompress{
			"cgrates.org:TestStatRemExpired_4": {Stat: 0.5, CompressFactor: 4},
		},
	}
	tcd := &StatTCD{
		Sum:   3 * time.Minute,
		Count: 2,
		Events: map[string]*DurationWithCompress{
			"cgrates.org:TestStatRemExpired_2": {Duration: time.Minute, CompressFactor: 1},
			"cgrates.org:TestStatRemExpired_3": {Duration: 2 * time.Minute, CompressFactor: 1},
		},
	}
	expectedTCD := &StatTCD{
		Sum:   3 * time.Minute,
		Count: 2,
		Events: map[string]*DurationWithCompress{
			"cgrates.org:TestStatRemExpired_2": {Duration: time.Minute, CompressFactor: 1},
			"cgrates.org:TestStatRemExpired_3": {Duration: 2 * time.Minute, CompressFactor: 1},
		},
	}
	sqItems := []SQItem{
		{"cgrates.org:TestStatRemExpired_1", utils.TimePointer(tmNow)},
		{"cgrates.org:TestStatRemExpired_2", utils.TimePointer(tmNow.Add(time.Minute))},
		{"cgrates.org:TestStatRemExpired_3", utils.TimePointer(tmNow.Add(2 * time.Minute))},
		{"cgrates.org:TestStatRemExpired_4", nil},
	}
	expectedSqItems := []SQItem{
		{"cgrates.org:TestStatRemExpired_2", utils.TimePointer(tmNow.Add(time.Minute))},
		{"cgrates.org:TestStatRemExpired_3", utils.TimePointer(tmNow.Add(2 * time.Minute))},
		{"cgrates.org:TestStatRemExpired_4", nil},
	}
	sq = &StatQueue{
		SQItems: sqItems,
		ttl:     utils.DurationPointer(time.Second),
		SQMetrics: map[string]StatMetric{
			utils.MetaASR: asr,
			utils.MetaTCD: tcd,
		},
	}
	if sq.Compress(int64(100), config.CgrConfig().GeneralCfg().RoundingDecimals) {
		t.Errorf("StatQueue compressed: %s", utils.ToJSON(sq))
	}
	if !sq.Compress(int64(3), config.CgrConfig().GeneralCfg().RoundingDecimals) {
		t.Errorf("StatQueue not compressed: %s", utils.ToJSON(sq))
	}
	if !reflect.DeepEqual(sq.SQItems, expectedSqItems) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expectedSqItems), utils.ToJSON(sq.SQItems))
	}
	if rply := sq.SQMetrics[utils.MetaASR].(*StatASR); !reflect.DeepEqual(*rply, *expectedASR) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expectedASR), utils.ToJSON(rply))
	}
	if rply := sq.SQMetrics[utils.MetaTCD].(*StatTCD); !reflect.DeepEqual(*rply, *expectedTCD) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expectedTCD), utils.ToJSON(rply))
	}
}

func TestStatExpand(t *testing.T) {
	expectedASR := &StatASR{
		Answered: 2,
		Count:    4,
		Events: map[string]*StatWithCompress{
			"cgrates.org:TestStatRemExpired_4": {Stat: 0.5, CompressFactor: 4},
		},
	}
	asr := &StatASR{
		Answered: 2,
		Count:    4,
		Events: map[string]*StatWithCompress{
			"cgrates.org:TestStatRemExpired_4": {Stat: 0.5, CompressFactor: 4},
		},
	}
	expectedSqItems := []SQItem{
		{"cgrates.org:TestStatRemExpired_4", nil},
		{"cgrates.org:TestStatRemExpired_4", nil},
		{"cgrates.org:TestStatRemExpired_4", nil},
		{"cgrates.org:TestStatRemExpired_4", nil},
	}
	sqItems := []SQItem{
		{"cgrates.org:TestStatRemExpired_4", nil},
	}
	sq = &StatQueue{
		SQItems: sqItems,
		SQMetrics: map[string]StatMetric{
			utils.MetaASR: asr,
		},
	}
	sq.Expand()
	if !reflect.DeepEqual(sq.SQItems, expectedSqItems) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expectedSqItems), utils.ToJSON(sq.SQItems))
	}
	if rply := sq.SQMetrics[utils.MetaASR].(*StatASR); !reflect.DeepEqual(*rply, *expectedASR) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expectedASR), utils.ToJSON(rply))
	}
}

func TestStatExpand2(t *testing.T) {
	tmNow := time.Now()
	expectedASR := &StatASR{
		Answered: 2,
		Count:    4,
		Events: map[string]*StatWithCompress{
			"cgrates.org:TestStatRemExpired_4": {Stat: 0.5, CompressFactor: 4},
		},
	}
	asr := &StatASR{
		Answered: 2,
		Count:    4,
		Events: map[string]*StatWithCompress{
			"cgrates.org:TestStatRemExpired_4": {Stat: 0.5, CompressFactor: 4},
		},
	}
	expectedTCD := &StatTCD{
		Sum:   3 * time.Minute,
		Count: 2,
		Events: map[string]*DurationWithCompress{
			"cgrates.org:TestStatRemExpired_4": {Duration: time.Minute + 30*time.Second, CompressFactor: 2},
		},
	}
	tcd := &StatTCD{
		Sum:   3 * time.Minute,
		Count: 2,
		Events: map[string]*DurationWithCompress{
			"cgrates.org:TestStatRemExpired_4": {Duration: time.Minute + 30*time.Second, CompressFactor: 2},
		},
	}
	expectedSqItems := []SQItem{
		{"cgrates.org:TestStatRemExpired_4", nil},
		{"cgrates.org:TestStatRemExpired_4", nil},
		{"cgrates.org:TestStatRemExpired_4", nil},
		{"cgrates.org:TestStatRemExpired_4", nil},
	}
	sqItems := []SQItem{
		{"cgrates.org:TestStatRemExpired_3", utils.TimePointer(tmNow.Add(2 * time.Minute))},
		{"cgrates.org:TestStatRemExpired_4", nil},
	}
	sq = &StatQueue{
		SQItems: sqItems,
		ttl:     utils.DurationPointer(time.Second),
		SQMetrics: map[string]StatMetric{
			utils.MetaASR: asr,
			utils.MetaTCD: tcd,
		},
	}
	sq.Expand()
	if !reflect.DeepEqual(sq.SQItems, expectedSqItems) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expectedSqItems), utils.ToJSON(sq.SQItems))
	}
	if rply := sq.SQMetrics[utils.MetaASR].(*StatASR); !reflect.DeepEqual(*rply, *expectedASR) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expectedASR), utils.ToJSON(rply))
	}
	if rply := sq.SQMetrics[utils.MetaTCD].(*StatTCD); !reflect.DeepEqual(*rply, *expectedTCD) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expectedTCD), utils.ToJSON(rply))
	}
}
func TestStatExpand3(t *testing.T) {
	tmNow := time.Now()
	expectedASR := &StatASR{
		Answered: 2,
		Count:    4,
		Events: map[string]*StatWithCompress{
			"cgrates.org:TestStatRemExpired_3": {Stat: 0.5, CompressFactor: 4},
		},
	}
	asr := &StatASR{
		Answered: 2,
		Count:    4,
		Events: map[string]*StatWithCompress{
			"cgrates.org:TestStatRemExpired_3": {Stat: 0.5, CompressFactor: 4},
		},
	}
	expectedTCD := &StatTCD{
		Sum:   3 * time.Minute,
		Count: 2,
		Events: map[string]*DurationWithCompress{
			"cgrates.org:TestStatRemExpired_2": {Duration: time.Minute, CompressFactor: 1},
			"cgrates.org:TestStatRemExpired_4": {Duration: 2 * time.Minute, CompressFactor: 1},
		},
	}
	tcd := &StatTCD{
		Sum:   3 * time.Minute,
		Count: 2,
		Events: map[string]*DurationWithCompress{
			"cgrates.org:TestStatRemExpired_2": {Duration: time.Minute, CompressFactor: 1},
			"cgrates.org:TestStatRemExpired_4": {Duration: 2 * time.Minute, CompressFactor: 1},
		},
	}
	expectedSqItems := []SQItem{
		{"cgrates.org:TestStatRemExpired_2", utils.TimePointer(tmNow.Add(time.Minute))},
		{"cgrates.org:TestStatRemExpired_3", utils.TimePointer(tmNow.Add(2 * time.Minute))},
		{"cgrates.org:TestStatRemExpired_3", utils.TimePointer(tmNow.Add(2 * time.Minute))},
		{"cgrates.org:TestStatRemExpired_3", utils.TimePointer(tmNow.Add(2 * time.Minute))},
		{"cgrates.org:TestStatRemExpired_3", utils.TimePointer(tmNow.Add(2 * time.Minute))},
		{"cgrates.org:TestStatRemExpired_4", nil},
	}
	sqItems := []SQItem{
		{"cgrates.org:TestStatRemExpired_2", utils.TimePointer(tmNow.Add(time.Minute))},
		{"cgrates.org:TestStatRemExpired_3", utils.TimePointer(tmNow.Add(2 * time.Minute))},
		{"cgrates.org:TestStatRemExpired_4", nil},
	}
	sq = &StatQueue{
		SQItems: sqItems,
		ttl:     utils.DurationPointer(time.Second),
		SQMetrics: map[string]StatMetric{
			utils.MetaASR: asr,
			utils.MetaTCD: tcd,
		},
	}
	sq.Expand()
	if !reflect.DeepEqual(sq.SQItems, expectedSqItems) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expectedSqItems), utils.ToJSON(sq.SQItems))
	}
	if rply := sq.SQMetrics[utils.MetaASR].(*StatASR); !reflect.DeepEqual(*rply, *expectedASR) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expectedASR), utils.ToJSON(rply))
	}
	if rply := sq.SQMetrics[utils.MetaTCD].(*StatTCD); !reflect.DeepEqual(*rply, *expectedTCD) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expectedTCD), utils.ToJSON(rply))
	}
}

func TestStatRemoveExpiredTTL(t *testing.T) {
	sq = &StatQueue{
		ttl: utils.DurationPointer(100 * time.Millisecond),
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
			QueueLength: 0, //unlimited que
		},
	}

	//add ev1 with ttl 100ms (after 100ms the event should be removed)
	ev1 := &utils.CGREvent{Tenant: "cgrates.org", ID: "TestStatAddStatEvent_1"}
	sq.ProcessEvent(ev1.Tenant, ev1.ID, nil, utils.MapStorage{utils.MetaReq: ev1.Event})

	if len(sq.SQItems) != 1 && sq.SQItems[0].EventID != "TestStatAddStatEvent_1" {
		t.Errorf("Expecting: 1, received: %+v", len(sq.SQItems))
	}
	//after 150ms the event expired
	time.Sleep(150 * time.Millisecond)

	//processing a new event should clean the expired events and add the new one
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "TestStatAddStatEvent_2"}
	sq.ProcessEvent(ev2.Tenant, ev2.ID, nil, utils.MapStorage{utils.MetaReq: ev2.Event})
	if len(sq.SQItems) != 1 && sq.SQItems[0].EventID != "TestStatAddStatEvent_2" {
		t.Errorf("Expecting: 1, received: %+v", len(sq.SQItems))
	}
}

func TestStatRemoveExpiredQueue(t *testing.T) {
	sq = &StatQueue{
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

	//add ev1 with ttl 100ms (after 100ms the event should be removed)
	ev1 := &utils.CGREvent{Tenant: "cgrates.org", ID: "TestStatAddStatEvent_1"}
	sq.ProcessEvent(ev1.Tenant, ev1.ID, nil, utils.MapStorage{utils.MetaReq: ev1.Event})

	if len(sq.SQItems) != 1 && sq.SQItems[0].EventID != "TestStatAddStatEvent_1" {
		t.Errorf("Expecting: 1, received: %+v", len(sq.SQItems))
	}
	//after 150ms the event expired
	time.Sleep(150 * time.Millisecond)

	//processing a new event should clean the expired events and add the new one
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "TestStatAddStatEvent_2"}
	sq.ProcessEvent(ev2.Tenant, ev2.ID, nil, utils.MapStorage{utils.MetaReq: ev2.Event})
	if len(sq.SQItems) != 2 && sq.SQItems[0].EventID != "TestStatAddStatEvent_1" &&
		sq.SQItems[1].EventID != "TestStatAddStatEvent_2" {
		t.Errorf("Expecting: 2, received: %+v", len(sq.SQItems))
	}

	//processing a new event should clean the expired events and add the new one
	ev3 := &utils.CGREvent{Tenant: "cgrates.org", ID: "TestStatAddStatEvent_3"}
	sq.ProcessEvent(ev3.Tenant, ev3.ID, nil, utils.MapStorage{utils.MetaReq: ev3.Event})
	if len(sq.SQItems) != 2 && sq.SQItems[0].EventID != "TestStatAddStatEvent_2" &&
		sq.SQItems[1].EventID != "TestStatAddStatEvent_3" {
		t.Errorf("Expecting: 2, received: %+v", len(sq.SQItems))
	}
}

func TestLibstatsSqID(t *testing.T) {
	ssq := &StoredStatQueue{
		ID:     "testID",
		Tenant: "testTenant",
	}

	exp := "testTenant:testID"
	rcv := ssq.SqID()

	if rcv != exp {
		t.Errorf("\nexpected: %q, \nreceived: %q", exp, rcv)
	}
}

type statMetricMock struct {
	testcase string
}

func (sMM *statMetricMock) GetValue(roundingDecimal int) interface{} {
	return nil
}

func (sMM *statMetricMock) GetStringValue(roundingDecimal int) (val string) {
	return
}

func (sMM *statMetricMock) GetFloat64Value(roundingDecimal int) (val float64) {
	return
}

func (sMM *statMetricMock) AddEvent(evID string, ev utils.DataProvider) error {
	return nil
}

func (sMM *statMetricMock) RemEvent(evTenantID string) error {
	switch sMM.testcase {
	case "remExpired error":
		return fmt.Errorf("remExpired mock error")
	}
	return nil
}

func (sMM *statMetricMock) Marshal(ms Marshaler) (marshaled []byte, err error) {
	err = fmt.Errorf("marshal mock error")
	return
}

func (sMM *statMetricMock) LoadMarshaled(ms Marshaler, marshaled []byte) (err error) {
	return nil
}

func (sMM *statMetricMock) GetFilterIDs() (filterIDs []string) {
	switch sMM.testcase {
	case "pass error":
		filterIDs = []string{"filter1", "filter2"}
		return
	}
	return
}

func (sMM *statMetricMock) GetMinItems() (minIts int) {
	return 0
}

func (sMM *statMetricMock) Compress(queueLen int64, defaultID string, roundingDec int) (eventIDs []string) {
	switch sMM.testcase {
	case "populate idMap":
		eventIDs = []string{"id1", "id2", "id3", "id4", "id5", "id6"}
		return
	}
	return
}

func (sMM *statMetricMock) GetCompressFactor(events map[string]int) map[string]int {
	return nil
}

func TestLibstatsNewStoredStatQueue(t *testing.T) {
	sq := &StatQueue{
		SQMetrics: map[string]StatMetric{
			"key": &statMetricMock{},
		},
	}
	var ms Marshaler

	experr := "marshal mock error"
	rcv, err := NewStoredStatQueue(sq, ms)

	if err == nil || err.Error() != experr {
		t.Fatalf("\nreceived: %q, \nexpected: %q", experr, err)
	}

	if rcv != nil {
		t.Errorf("\nreceived: <%+v>, \nexpected: <%+v>", nil, rcv)
	}
}

func TestLibstatsAsStatQueueNilStoredSq(t *testing.T) {
	var ssq *StoredStatQueue
	var ms Marshaler

	rcv, err := ssq.AsStatQueue(ms)

	if err != nil {
		t.Fatalf("\nexpected: <%+v>, \nreceived: <%+v>", nil, err)
	}

	if rcv != nil {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", nil, rcv)
	}
}

func TestLibstatsAsStatQueueSuccess(t *testing.T) {
	ssq := &StoredStatQueue{
		SQItems: []SQItem{
			{
				EventID: "testEventID",
			},
		},
	}
	var ms Marshaler

	exp := &StatQueue{
		SQItems: []SQItem{
			{
				EventID: "testEventID",
			},
		},
		SQMetrics: map[string]StatMetric{},
	}
	rcv, err := ssq.AsStatQueue(ms)

	if err != nil {
		t.Fatalf("\nexpected: <%+v>, \nreceived: <%+v>", nil, err)
	}

	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", exp, rcv)
	}
}

func TestLibstatsAsStatQueueUnsupportedMetric(t *testing.T) {
	ssq := &StoredStatQueue{
		SQItems: []SQItem{
			{
				EventID: "testEventID",
			},
		},
		SQMetrics: map[string][]byte{
			"key": []byte("sqmetric"),
		},
	}
	var ms Marshaler

	experr := fmt.Sprintf("unsupported metric type <%s>", "key")
	rcv, err := ssq.AsStatQueue(ms)

	if err == nil || err.Error() != experr {
		t.Fatalf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	if rcv != nil {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", nil, rcv)
	}
}

func TestLibstatsAsStatQueueErrLoadMarshaled(t *testing.T) {
	ssq := &StoredStatQueue{
		SQItems: []SQItem{
			{
				EventID: "testEventID",
			},
		},
		SQMetrics: map[string][]byte{
			utils.MetaTCD: []byte(""),
		},
		Compressed: true,
	}
	ms, err := NewMarshaler(utils.JSON)
	if err != nil {
		t.Fatal(err)
	}

	experr := "unexpected end of JSON input"
	rcv, err := ssq.AsStatQueue(ms)

	if err == nil || err.Error() != experr {
		t.Fatalf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	if rcv != nil {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", nil, rcv)
	}
}

func TestLibstatsAsStatQueueOK(t *testing.T) {
	ms, err := NewMarshaler(utils.JSON)
	if err != nil {
		t.Fatal(err)
	}

	sm, err := NewStatMetric(utils.MetaTCD, 0, []string{})
	if err != nil {
		t.Fatal(err)
	}

	msm, err := sm.Marshal(ms)
	if err != nil {
		t.Fatal(err)
	}

	ssq := &StoredStatQueue{
		SQItems: []SQItem{
			{
				EventID: "testEventID",
			},
		},
		SQMetrics: map[string][]byte{
			utils.MetaTCD: msm,
		},
		Compressed: true,
	}

	exp := &StatQueue{
		SQMetrics: map[string]StatMetric{
			utils.MetaTCD: sm,
		},
	}
	rcv, err := ssq.AsStatQueue(ms)

	if err != nil {
		t.Fatalf("\nexpected: <%+v>, \nreceived: <%+v>", nil, err)
	}

	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", exp, rcv)
	}
}

func TestLibstatsNewStatQueue(t *testing.T) {
	tnt := "tenant"
	id := "id"
	metrics := []*MetricWithFilters{
		{
			MetricID: "invalid",
		},
	}
	minItems := 0

	experr := fmt.Sprintf("unsupported metric type <%s>", metrics[0].MetricID)
	exp := &StatQueue{
		Tenant: tnt,
		ID:     id,
		SQMetrics: map[string]StatMetric{
			"invalid": nil,
		},
	}
	rcv, err := NewStatQueue(tnt, id, metrics, minItems)

	if err == nil || err.Error() != experr {
		t.Fatalf("\nexpected: %q, \nreceived: %q", experr, err)
	}

	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", exp, rcv)
	}
}

func TestLibstatsProcessEventremExpiredErr(t *testing.T) {
	tnt, evID := "tenant", "eventID"
	filters := &FilterS{}
	expiry := time.Date(2021, 1, 1, 23, 59, 59, 10, time.UTC)
	evNm := utils.MapStorage{
		"key": nil,
	}

	sq := &StatQueue{
		sqPrfl: &StatQueueProfile{
			QueueLength: -1,
		},
		SQItems: []SQItem{
			{
				EventID:    evID,
				ExpiryTime: &expiry,
			},
		},
		SQMetrics: map[string]StatMetric{
			"key": &statMetricMock{
				testcase: "remExpired error",
			},
		},
	}

	experr := "remExpired mock error"
	err := sq.ProcessEvent(tnt, evID, filters, evNm)

	if err == nil || err.Error() != experr {
		t.Errorf("\nexpected: %q, \nreceived: %q", experr, err)
	}
}

func TestLibstatsProcessEventremOnQueueLengthErr(t *testing.T) {
	tnt, evID := "tenant", "eventID"
	filters := &FilterS{}
	evNm := utils.MapStorage{
		"key": nil,
	}

	sq := &StatQueue{
		sqPrfl: &StatQueueProfile{
			QueueLength: 1,
		},
		SQItems: []SQItem{
			{
				EventID: evID,
			},
		},
		SQMetrics: map[string]StatMetric{
			"key": &statMetricMock{
				testcase: "remExpired error",
			},
		},
	}

	experr := "remExpired mock error"
	err := sq.ProcessEvent(tnt, evID, filters, evNm)

	if err == nil || err.Error() != experr {
		t.Errorf("\nexpected: %q, \nreceived: %q", experr, err)
	}
}

func TestLibstatsProcessEventaddStatEvent(t *testing.T) {
	tnt, evID := "tenant", "eventID"
	filters := &FilterS{}
	evNm := utils.MapStorage{
		"key": nil,
	}

	sq := &StatQueue{
		sqPrfl: &StatQueueProfile{
			QueueLength: 1,
		},
		SQItems: []SQItem{
			{
				EventID: evID,
			},
		},
		SQMetrics: map[string]StatMetric{
			utils.MetaTCD: &StatTCD{},
		},
	}

	experr := utils.ErrWrongPath
	err := sq.ProcessEvent(tnt, evID, filters, evNm)

	if err == nil || err != experr {
		t.Errorf("\nexpected: %q, \nreceived: %q", experr, err)
	}
}

func TestLibstatsCompress(t *testing.T) {
	sm, err := NewStatMetric(utils.MetaTCD, 0, []string{"*string:~*req.Account:1001"})
	if err != nil {
		t.Fatal(err)
	}

	ttl := time.Millisecond
	expiryTime1 := time.Date(2021, 1, 1, 23, 59, 59, 0, time.UTC)
	expiryTime2 := time.Date(2021, 1, 2, 23, 59, 59, 0, time.UTC)
	expiryTime3 := time.Date(2021, 1, 3, 23, 59, 59, 0, time.UTC)
	expiryTime4 := time.Date(2021, 1, 4, 23, 59, 59, 0, time.UTC)
	sq := &StatQueue{
		SQItems: []SQItem{
			{
				EventID:    "id1",
				ExpiryTime: &expiryTime1,
			},
			{
				EventID:    "id2",
				ExpiryTime: &expiryTime2,
			},
			{
				EventID:    "id3",
				ExpiryTime: &expiryTime3,
			},
			{
				EventID:    "id4",
				ExpiryTime: &expiryTime4,
			},
			{
				EventID: "id5",
			},
		},
		SQMetrics: map[string]StatMetric{
			utils.MetaTCD: &statMetricMock{
				testcase: "populate idMap",
			},
			utils.MetaReq: sm,
		},
		ttl: &ttl,
	}

	maxQL := int64(1)
	roundDec := 1

	exp := []SQItem{
		{
			EventID:    "id1",
			ExpiryTime: &expiryTime1,
		},
		{
			EventID:    "id2",
			ExpiryTime: &expiryTime2,
		},
		{
			EventID:    "id3",
			ExpiryTime: &expiryTime3,
		},
		{
			EventID:    "id4",
			ExpiryTime: &expiryTime4,
		},
		{
			EventID: "id5",
		},
		{
			EventID: "id6",
		},
	}
	rcv := sq.Compress(maxQL, roundDec)
	if rcv != true {
		t.Fatalf("\nexpected: <%+v>, \nreceived: <%+v>", true, rcv)
	}

	if len(sq.SQItems) != len(exp) {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", exp, sq.SQItems)
	}
	// if !reflect.DeepEqual(sq.SQItems, exp) {
	// 	t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", exp, sq.SQItems)
	// }
}

func TestLibstatsaddStatEventPassErr(t *testing.T) {
	sq := &StatQueue{
		SQMetrics: map[string]StatMetric{
			utils.MetaTCD: &statMetricMock{
				testcase: "pass error",
			},
		},
	}
	tnt, evID := "tenant", "eventID"
	filters := &FilterS{
		cfg: config.CgrConfig(),
		dm: &DataManager{
			dataDB: NewInternalDB(nil, nil, true),
		},
		connMgr: &ConnManager{},
	}
	evNm := utils.MapStorage{
		utils.MetaReq: utils.MapStorage{
			utils.MetaReq: nil,
		},
		utils.MetaOpts: nil,
		utils.MetaVars: utils.MapStorage{
			utils.OptsAttributesProcessRuns: 0,
		},
	}

	experr := "NOT_FOUND:filter1"
	err := sq.addStatEvent(tnt, evID, filters, evNm)

	if err == nil || err.Error() != experr {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	}
}

func TestLibstatsaddStatEventNoPass(t *testing.T) {
	sm, err := NewStatMetric(utils.MetaTCD, 0, []string{"*string:~*req.Account:1001"})
	if err != nil {
		t.Fatal(err)
	}

	sq := &StatQueue{
		SQMetrics: map[string]StatMetric{
			utils.MetaTCD: sm,
		},
	}
	sq.Lock()

	tnt, evID := "cgrates.org", "eventID"
	filters := &FilterS{
		cfg: config.CgrConfig(),
		dm: &DataManager{
			dataDB: NewInternalDB(nil, nil, true),
		},
		connMgr: &ConnManager{},
	}
	evNm := utils.MapStorage{
		utils.MetaReq: utils.MapStorage{
			utils.MetaReq: nil,
		},
		utils.MetaOpts: nil,
		utils.MetaVars: utils.MapStorage{
			utils.OptsAttributesProcessRuns: 0,
		},
	}

	exp := &StatQueue{
		SQMetrics: map[string]StatMetric{
			utils.MetaTCD: sm,
		},
		SQItems: []SQItem{
			{
				EventID: "eventID",
			},
		},
	}
	err = sq.addStatEvent(tnt, evID, filters, evNm)
	sq.Unlock()

	if err != nil {
		t.Fatalf("\nexpected: <%+v>, \nreceived: <%+v>", nil, err)
	}

	if !reflect.DeepEqual(sq, exp) {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", exp, sq)
	}
}
