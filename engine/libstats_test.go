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
	"encoding/json"
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
	ev1.Event = map[string]any{
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

func TestStatQueueSqID(t *testing.T) {
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

func (sMM *statMetricMock) Clone() StatMetric {
	return nil
}

func (sMM *statMetricMock) GetValue(roundingDecimal int) any {
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

func (sMM *statMetricMock) AddOneEvent(ev utils.DataProvider) error {
	return nil
}

func (sMM *statMetricMock) RemEvent(evTenantID string) {
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

func TestStatQueueAsStatQueueNilStoredSq(t *testing.T) {
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

func TestStatQueueAsStatQueueSuccess(t *testing.T) {
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

func TestStatQueueAsStatQueueUnsupportedMetric(t *testing.T) {
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

func TestStatQueueAsStatQueueErrLoadMarshaled(t *testing.T) {
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

func TestStatQueueAsStatQueueOK(t *testing.T) {
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

func TestStatQueueNewStatQueue(t *testing.T) {
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

func TestStatQueueCompress(t *testing.T) {
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
}

func TestStatQueueaddStatEventPassErr(t *testing.T) {
	sq := &StatQueue{
		SQMetrics: map[string]StatMetric{
			utils.MetaTCD: &statMetricMock{
				testcase: "pass error",
			},
		},
	}
	tnt, evID := "tenant", "eventID"
	idb, err := NewInternalDB(nil, nil, true, nil, config.CgrConfig().DataDbCfg().Items)
	if err != nil {
		t.Fatal(err)
	}
	filters := &FilterS{
		cfg: config.CgrConfig(),
		dm: &DataManager{
			dataDB: idb,
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
	err = sq.addStatEvent(tnt, evID, filters, evNm)

	if err == nil || err.Error() != experr {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	}
}

func TestStatQueueJSONMarshall(t *testing.T) {
	var rply *StatQueue
	exp, err := NewStatQueue("cgrates.org", "STS", []*MetricWithFilters{
		{MetricID: utils.MetaASR},
		{MetricID: utils.MetaTCD},
	}, 1)
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal([]byte(utils.ToJSON(exp)), &rply); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(rply, exp) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(exp), utils.ToJSON(rply))
	}

}

func TestStatQueueWithAPIOptsJSONMarshall(t *testing.T) {
	rply := &StatQueueWithAPIOpts{ /*StatQueue: &StatQueue{}*/ }
	exp, err := NewStatQueue("cgrates.org", "STS", []*MetricWithFilters{
		{MetricID: utils.MetaASR},
		{MetricID: utils.MetaTCD},
	}, 1)
	exp2 := &StatQueueWithAPIOpts{
		StatQueue: exp,
		APIOpts:   map[string]any{"a": "a"},
	}
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal([]byte(utils.ToJSON(exp2)), rply); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(rply, exp2) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(exp2), utils.ToJSON(rply))
	}

}

func TestStatQueueLockUnlockStatQueueProfiles(t *testing.T) {
	sqPrf := &StatQueueProfile{
		Tenant:      "cgrates.org",
		ID:          "SQ1",
		Weight:      10,
		QueueLength: 10,
	}

	//lock profile with empty lkID parameter
	sqPrf.lock(utils.EmptyString)

	if !sqPrf.isLocked() {
		t.Fatal("expected profile to be locked")
	} else if sqPrf.lkID == utils.EmptyString {
		t.Fatal("expected struct field \"lkID\" to be non-empty")
	}

	//unlock previously locked profile
	sqPrf.unlock()

	if sqPrf.isLocked() {
		t.Fatal("expected profile to be unlocked")
	} else if sqPrf.lkID != utils.EmptyString {
		t.Fatal("expected struct field \"lkID\" to be empty")
	}

	//unlock an already unlocked profile - nothing happens
	sqPrf.unlock()

	if sqPrf.isLocked() {
		t.Fatal("expected profile to be unlocked")
	} else if sqPrf.lkID != utils.EmptyString {
		t.Fatal("expected struct field \"lkID\" to be empty")
	}
}

func TestStatQueueLockUnlockStatQueues(t *testing.T) {
	sq := &StatQueue{
		Tenant: "cgrates.org",
		ID:     "SQ1",
	}

	//lock resource with empty lkID parameter
	sq.lock(utils.EmptyString)

	if !sq.isLocked() {
		t.Fatal("expected resource to be locked")
	} else if sq.lkID == utils.EmptyString {
		t.Fatal("expected struct field \"lkID\" to be non-empty")
	}

	//unlock previously locked resource
	sq.unlock()

	if sq.isLocked() {
		t.Fatal("expected resource to be unlocked")
	} else if sq.lkID != utils.EmptyString {
		t.Fatal("expected struct field \"lkID\" to be empty")
	}

	//unlock an already unlocked resource - nothing happens
	sq.unlock()

	if sq.isLocked() {
		t.Fatal("expected resource to be unlocked")
	} else if sq.lkID != utils.EmptyString {
		t.Fatal("expected struct field \"lkID\" to be empty")
	}
}

func TestStatQueueUnmarshalJSON(t *testing.T) {
	sq := &StatQueue{}
	if err := sq.UnmarshalJSON(nil); err == nil {
		t.Error(err)
	}
	ssq := &StatQueueWithAPIOpts{}
	if err := ssq.UnmarshalJSON(nil); err == nil {
		t.Error(err)
	} else if err := ssq.UnmarshalJSON([]byte("value:key")); err == nil {
		t.Error(err)
	}

}

func TestLibRoutesRouteIDs(t *testing.T) {
	sortedRoutesList := SortedRoutesList{
		{
			Routes: []*SortedRoute{
				{RouteID: "1"},
				{RouteID: "2"},
			},
		},
		{
			Routes: []*SortedRoute{
				{RouteID: "3"},
				{RouteID: "4"},
			},
		},
	}
	expectedIDs := []string{"1", "2", "3", "4"}
	actualIDs := sortedRoutesList.RouteIDs()
	if len(actualIDs) != len(expectedIDs) {
		t.Errorf("expected length %d, got %d", len(expectedIDs), len(actualIDs))
		return
	}
	for i := range expectedIDs {
		if actualIDs[i] != expectedIDs[i] {
			t.Errorf("at index %d, expected %s, got %s", i, expectedIDs[i], actualIDs[i])
		}
	}
}

func TestSortLeastCost(t *testing.T) {
	sRoutes := &SortedRoutes{
		ProfileID: "test-profile",
		Sorting:   "LeastCost",
		Routes: []*SortedRoute{
			{
				RouteID:        "route1",
				SortingData:    map[string]any{"Cost": 10.0, "Weight": 3.0},
				sortingDataF64: map[string]float64{"Cost": 10.0, "Weight": 3.0},
			},
			{
				RouteID:        "route2",
				SortingData:    map[string]any{"Cost": 5.0, "Weight": 2.0},
				sortingDataF64: map[string]float64{"Cost": 5.0, "Weight": 2.0},
			},
		},
	}
	sRoutes.SortLeastCost()

	expectedOrder := []string{"route2", "route1"}

	for i, route := range sRoutes.Routes {
		if route.RouteID != expectedOrder[i] {
			t.Errorf("Expected route ID %s at position %d, but got %s", expectedOrder[i], i, route.RouteID)
		}
	}
}

func TestSortedRoutessAsNavigableMap(t *testing.T) {
	route := &SortedRoute{
		RouteID:         "route1",
		RouteParameters: "param1",
		SortingData: map[string]any{
			"Cost": 10.5,
			"PDD":  20.0,
		},
	}

	navigableMap := route.AsNavigableMap()

	if navigableMap.Type != utils.NMMapType {
		t.Errorf("Expected Type %v, got %v", utils.NMMapType, navigableMap.Type)
	}

	routeIDNode, exists := navigableMap.Map[utils.RouteID]
	if !exists {
		t.Fatalf("RouteID node does not exist in navigable map")
	}
	if routeIDNode.Value == nil || routeIDNode.Value.Data != "route1" {
		t.Errorf("Expected RouteID value 'route1', got %v", routeIDNode.Value)
	}

	routeParamsNode, exists := navigableMap.Map[utils.RouteParameters]
	if !exists {
		t.Fatalf("RouteParameters node does not exist in navigable map")
	}
	if routeParamsNode.Value == nil || routeParamsNode.Value.Data != "param1" {
		t.Errorf("Expected RouteParameters value 'param1', got %v", routeParamsNode.Value)
	}

	sortingDataNode, exists := navigableMap.Map[utils.SortingData]
	if !exists {
		t.Fatalf("SortingData node does not exist in navigable map")
	}
	if sortingDataNode.Type != utils.NMMapType {
		t.Errorf("Expected SortingData node Type %v, got %v", utils.NMMapType, sortingDataNode.Type)
	}

	expectedSortingData := map[string]any{
		"Cost": 10.5,
		"PDD":  20.0,
	}

	for key, expectedValue := range expectedSortingData {
		actualNode, exists := sortingDataNode.Map[key]
		if !exists {
			t.Errorf("Expected SortingData key %s to exist", key)
			continue
		}
		if actualNode.Value == nil || !reflect.DeepEqual(actualNode.Value.Data, expectedValue) {
			t.Errorf("For key %s, expected value %v, got %v", key, expectedValue, actualNode.Value)
		}
	}
}

func TestRoutesSortedRoutesAsNavigableMap(t *testing.T) {
	route1 := &SortedRoute{
		RouteID:         "route1",
		RouteParameters: "param1",
		SortingData: map[string]any{
			"Cost": 10.5,
			"PDD":  20.0,
		},
	}

	route2 := &SortedRoute{
		RouteID:         "route2",
		RouteParameters: "param2",
		SortingData: map[string]any{
			"Cost": 5.5,
			"PDD":  15.0,
		},
	}

	sRoutes := &SortedRoutes{
		ProfileID: "profile1",
		Sorting:   "leastCost",
		Routes:    []*SortedRoute{route1, route2},
	}

	navigableMap := sRoutes.AsNavigableMap()

	if navigableMap.Type != utils.NMMapType {
		t.Errorf("Expected Type %v, got %v", utils.NMMapType, navigableMap.Type)
	}

	profileIDNode, exists := navigableMap.Map[utils.ProfileID]
	if !exists {
		t.Fatalf("ProfileID node does not exist in navigable map")
	}
	if profileIDNode.Value == nil || profileIDNode.Value.Data != "profile1" {
		t.Errorf("Expected ProfileID value 'profile1', got %v", profileIDNode.Value)
	}

	sortingNode, exists := navigableMap.Map[utils.Sorting]
	if !exists {
		t.Fatalf("Sorting node does not exist in navigable map")
	}
	if sortingNode.Value == nil || sortingNode.Value.Data != "leastCost" {
		t.Errorf("Expected Sorting value 'leastCost', got %v", sortingNode.Value)
	}

	capRoutesNode, exists := navigableMap.Map[utils.CapRoutes]
	if !exists {
		t.Fatalf("CapRoutes node does not exist in navigable map")
	}
	if capRoutesNode.Type != utils.NMSliceType {
		t.Errorf("Expected CapRoutes node Type %v, got %v", utils.NMSliceType, capRoutesNode.Type)
	}

	if len(capRoutesNode.Slice) != 2 {
		t.Errorf("Expected 2 routes in CapRoutes, got %d", len(capRoutesNode.Slice))
	}

	for i, routeNode := range capRoutesNode.Slice {
		if routeNode.Type != utils.NMMapType {
			t.Errorf("Expected route node Type %v, got %v", utils.NMMapType, routeNode.Type)
		}
		expectedRouteID := "route1"
		expectedRouteParameters := "param1"
		if i == 1 {
			expectedRouteID = "route2"
			expectedRouteParameters = "param2"
		}
		routeIDNode, exists := routeNode.Map[utils.RouteID]
		if !exists || routeIDNode.Value == nil || routeIDNode.Value.Data != expectedRouteID {
			t.Errorf("For route %d, expected RouteID value '%s', got %v", i+1, expectedRouteID, routeIDNode.Value)
		}
		routeParamsNode, exists := routeNode.Map[utils.RouteParameters]
		if !exists || routeParamsNode.Value == nil || routeParamsNode.Value.Data != expectedRouteParameters {
			t.Errorf("For route %d, expected RouteParameters value '%s', got %v", i+1, expectedRouteParameters, routeParamsNode.Value)
		}
	}
}

func TestRoutesSortedRoutesListAsNavigableMap(t *testing.T) {
	route1 := &SortedRoute{
		RouteID:         "route1",
		RouteParameters: "param1",
		SortingData: map[string]any{
			"Cost": 10.5,
			"PDD":  20.0,
		},
	}

	route2 := &SortedRoute{
		RouteID:         "route2",
		RouteParameters: "param2",
		SortingData: map[string]any{
			"Cost": 5.5,
			"PDD":  15.0,
		},
	}

	sRoutesList := SortedRoutesList{
		{ProfileID: "profile1", Sorting: "leastCost", Routes: []*SortedRoute{route1}},
		{ProfileID: "profile2", Sorting: "maxCost", Routes: []*SortedRoute{route2}},
	}

	navigableMap := sRoutesList.AsNavigableMap()

	if navigableMap.Type != utils.NMSliceType {
		t.Errorf("Expected Type %v, got %v", utils.NMSliceType, navigableMap.Type)
	}

	if len(navigableMap.Slice) != 2 {
		t.Errorf("Expected slice length 2, got %d", len(navigableMap.Slice))
	}

	firstRouteNode := navigableMap.Slice[0]
	if firstRouteNode.Type != utils.NMMapType {
		t.Errorf("Expected first route node Type %v, got %v", utils.NMMapType, firstRouteNode.Type)
	}
	profileIDNode, exists := firstRouteNode.Map[utils.ProfileID]
	if !exists || profileIDNode.Value == nil || profileIDNode.Value.Data != "profile1" {
		t.Errorf("Expected ProfileID value 'profile1', got %v", profileIDNode.Value)
	}

	secondRouteNode := navigableMap.Slice[1]
	if secondRouteNode.Type != utils.NMMapType {
		t.Errorf("Expected second route node Type %v, got %v", utils.NMMapType, secondRouteNode.Type)
	}
	profileIDNode2, exists := secondRouteNode.Map[utils.ProfileID]
	if !exists || profileIDNode2.Value == nil || profileIDNode2.Value.Data != "profile2" {
		t.Errorf("Expected ProfileID value 'profile2', got %v", profileIDNode2.Value)
	}

	for idx, routeNode := range navigableMap.Slice {
		if routeNode.Type != utils.NMMapType {
			t.Errorf("Expected route node %d Type %v, got %v", idx, utils.NMMapType, routeNode.Type)
		}
		profileIDNode, exists := routeNode.Map[utils.ProfileID]
		if !exists || profileIDNode.Value == nil {
			t.Errorf("Expected ProfileID value, got nil at route %d", idx)
		}
	}
}

func TestSortWeight(t *testing.T) {
	routes := []*SortedRoute{
		{
			RouteID:        "route1",
			SortingData:    map[string]any{"Weight": 10.5},
			sortingDataF64: map[string]float64{"Weight": 10.5},
		},
		{
			RouteID:        "route2",
			SortingData:    map[string]any{"Weight": 20.3},
			sortingDataF64: map[string]float64{"Weight": 20.3},
		},
		{
			RouteID:        "route3",
			SortingData:    map[string]any{"Weight": 15.7},
			sortingDataF64: map[string]float64{"Weight": 15.7},
		},
	}

	sortedRoutes := &SortedRoutes{
		Routes: routes,
	}

	sortedRoutes.SortWeight()

	if sortedRoutes.Routes[0].RouteID != "route2" || sortedRoutes.Routes[1].RouteID != "route3" || sortedRoutes.Routes[2].RouteID != "route1" {
		t.Errorf("Expected sorted routes to be route2, route3, route1, but got %v", sortedRoutes.Routes)
	}
}

func TestSortQOS(t *testing.T) {
	routes := []*SortedRoute{
		{
			RouteID:         "route1",
			RouteParameters: "param1",
			sortingDataF64: map[string]float64{
				utils.MetaPDD: 10,
				utils.Weight:  5,
			},
		},
		{
			RouteID:         "route2",
			RouteParameters: "param2",
			sortingDataF64: map[string]float64{
				utils.MetaPDD: 20,
				utils.Weight:  5,
			},
		},
		{
			RouteID:         "route3",
			RouteParameters: "param3",
			sortingDataF64: map[string]float64{
				utils.MetaPDD: 15,
				utils.Weight:  3,
			},
		},
	}

	sortedRoutes := &SortedRoutes{
		ProfileID: "testProfile",
		Sorting:   "QOS",
		Routes:    routes,
	}

	sortedRoutes.SortQOS([]string{utils.MetaPDD})
	if sortedRoutes.Routes[0].RouteID != "route1" {
		t.Errorf("Expected route1 at position 0, got %s", sortedRoutes.Routes[0].RouteID)
	}
	if sortedRoutes.Routes[1].RouteID != "route3" {
		t.Errorf("Expected route3 at position 1, got %s", sortedRoutes.Routes[1].RouteID)
	}
	if sortedRoutes.Routes[2].RouteID != "route2" {
		t.Errorf("Expected route2 at position 2, got %s", sortedRoutes.Routes[2].RouteID)
	}

	sortedRoutes.SortQOS([]string{utils.MetaPDD, utils.Weight})
	if sortedRoutes.Routes[0].RouteID != "route1" {
		t.Errorf("Expected route1 at position 0, got %s", sortedRoutes.Routes[0].RouteID)
	}
	if sortedRoutes.Routes[1].RouteID != "route3" {
		t.Errorf("Expected route3 at position 1, got %s", sortedRoutes.Routes[1].RouteID)
	}
	if sortedRoutes.Routes[2].RouteID != "route2" {
		t.Errorf("Expected route2 at position 2, got %s", sortedRoutes.Routes[2].RouteID)
	}
}

func TestSortLeastCosts(t *testing.T) {
	sRoutes := &SortedRoutes{
		ProfileID: "TestProfile",
		Sorting:   "LeastCost",
		Routes: []*SortedRoute{
			{
				RouteID: "route1",
				sortingDataF64: map[string]float64{
					utils.Cost:   0.1,
					utils.Weight: 10.0,
				},
				SortingData: map[string]any{
					utils.Cost:   0.1,
					utils.Weight: 10.0,
				},
				RouteParameters: "param1",
			},
			{
				RouteID: "route2",
				sortingDataF64: map[string]float64{
					utils.Cost:   0.1,
					utils.Weight: 20.0,
				},
				SortingData: map[string]any{
					utils.Cost:   0.1,
					utils.Weight: 20.0,
				},
				RouteParameters: "param2",
			},
			{
				RouteID: "route3",
				sortingDataF64: map[string]float64{
					utils.Cost:   0.05,
					utils.Weight: 10.0,
				},
				SortingData: map[string]any{
					utils.Cost:   0.05,
					utils.Weight: 10.0,
				},
				RouteParameters: "param3",
			},
		},
	}

	sRoutes.SortLeastCost()

	expectedRoutes := &SortedRoutes{
		ProfileID: "TestProfile",
		Sorting:   "LeastCost",
		Routes: []*SortedRoute{
			{
				RouteID: "route3",
				sortingDataF64: map[string]float64{
					utils.Cost:   0.05,
					utils.Weight: 10.0,
				},
				SortingData: map[string]any{
					utils.Cost:   0.05,
					utils.Weight: 10.0,
				},
				RouteParameters: "param3",
			},
			{
				RouteID: "route2",
				sortingDataF64: map[string]float64{
					utils.Cost:   0.1,
					utils.Weight: 20.0,
				},
				SortingData: map[string]any{
					utils.Cost:   0.1,
					utils.Weight: 20.0,
				},
				RouteParameters: "param2",
			},
			{
				RouteID: "route1",
				sortingDataF64: map[string]float64{
					utils.Cost:   0.1,
					utils.Weight: 10.0,
				},
				SortingData: map[string]any{
					utils.Cost:   0.1,
					utils.Weight: 10.0,
				},
				RouteParameters: "param1",
			},
		},
	}

	if !reflect.DeepEqual(expectedRoutes, sRoutes) {
		t.Errorf("Expected: %+v, received: %+v", expectedRoutes, sRoutes)
	}

}

func TestSortResourceDescendent(t *testing.T) {
	sRoutes := &SortedRoutes{
		ProfileID: "TestProfile",
		Sorting:   "ResourceDescendent",
		Routes: []*SortedRoute{
			{
				RouteID: "route1",
				sortingDataF64: map[string]float64{
					utils.ResourceUsage: 0.5,
					utils.Weight:        10.0,
				},
				SortingData: map[string]any{
					utils.ResourceUsage: 0.5,
					utils.Weight:        10.0,
				},
				RouteParameters: "param1",
			},
			{
				RouteID: "route2",
				sortingDataF64: map[string]float64{
					utils.ResourceUsage: 0.5,
					utils.Weight:        20.0,
				},
				SortingData: map[string]any{
					utils.ResourceUsage: 0.5,
					utils.Weight:        20.0,
				},
				RouteParameters: "param2",
			},
			{
				RouteID: "route3",
				sortingDataF64: map[string]float64{
					utils.ResourceUsage: 0.8,
					utils.Weight:        10.0,
				},
				SortingData: map[string]any{
					utils.ResourceUsage: 0.8,
					utils.Weight:        10.0,
				},
				RouteParameters: "param3",
			},
		},
	}

	sRoutes.SortResourceDescendent()

	expectedRoutes := &SortedRoutes{
		ProfileID: "TestProfile",
		Sorting:   "ResourceDescendent",
		Routes: []*SortedRoute{
			{
				RouteID: "route3",
				sortingDataF64: map[string]float64{
					utils.ResourceUsage: 0.8,
					utils.Weight:        10.0,
				},
				SortingData: map[string]any{
					utils.ResourceUsage: 0.8,
					utils.Weight:        10.0,
				},
				RouteParameters: "param3",
			},
			{
				RouteID: "route2",
				sortingDataF64: map[string]float64{
					utils.ResourceUsage: 0.5,
					utils.Weight:        20.0,
				},
				SortingData: map[string]any{
					utils.ResourceUsage: 0.5,
					utils.Weight:        20.0,
				},
				RouteParameters: "param2",
			},
			{
				RouteID: "route1",
				sortingDataF64: map[string]float64{
					utils.ResourceUsage: 0.5,
					utils.Weight:        10.0,
				},
				SortingData: map[string]any{
					utils.ResourceUsage: 0.5,
					utils.Weight:        10.0,
				},
				RouteParameters: "param1",
			},
		},
	}

	if !reflect.DeepEqual(expectedRoutes, sRoutes) {
		t.Errorf("Expected: %+v, received: %+v", expectedRoutes, sRoutes)
	}

}
