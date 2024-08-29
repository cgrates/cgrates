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
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

var sq *StatQueue

func TestStatQueuesSort(t *testing.T) {
	sInsts := StatQueues{
		&StatQueue{weight: 30.0, sqPrfl: &StatQueueProfile{ID: "FIRST"}},
		&StatQueue{weight: 40.0, sqPrfl: &StatQueueProfile{ID: "SECOND"}},
		&StatQueue{weight: 30.0, sqPrfl: &StatQueueProfile{ID: "THIRD"}},
		&StatQueue{weight: 35.0, sqPrfl: &StatQueueProfile{ID: "FOURTH"}},
	}
	sInsts.Sort()
	eSInst := StatQueues{
		&StatQueue{weight: 40.0, sqPrfl: &StatQueueProfile{ID: "SECOND"}},
		&StatQueue{weight: 35.0, sqPrfl: &StatQueueProfile{ID: "FOURTH"}},
		&StatQueue{weight: 30.0, sqPrfl: &StatQueueProfile{ID: "FIRST"}},
		&StatQueue{weight: 30.0, sqPrfl: &StatQueueProfile{ID: "THIRD"}},
	}

	if !reflect.DeepEqual(eSInst, sInsts) {
		t.Errorf("expecting: %+v, received: %+v", eSInst, sInsts)
	}
}

func TestStatRemEventWithID(t *testing.T) {
	sq = &StatQueue{
		SQMetrics: map[string]StatMetric{
			utils.MetaASR: &StatASR{
				Metric: &Metric{
					Value: utils.NewDecimal(1, 0),
					Count: 2,
					Events: map[string]*DecimalWithCompress{
						"cgrates.org:TestRemEventWithID_1": {Stat: utils.NewDecimal(1, 0), CompressFactor: 1},
						"cgrates.org:TestRemEventWithID_2": {Stat: utils.NewDecimal(0, 0), CompressFactor: 1},
					},
				},
			},
		},
	}
	asrMetric := sq.SQMetrics[utils.MetaASR].(*StatASR)
	if asr := asrMetric.GetValue(); asr.Compare(utils.NewDecimal(50, 0)) != 0 {
		t.Errorf("received asrMetric: %v", utils.ToJSON(asrMetric.GetValue()))
	}
	sq.remEventWithID("cgrates.org:TestRemEventWithID_1")
	if asr := asrMetric.GetValue(); asr.Compare(utils.NewDecimal(0, 0)) != 0 {
		t.Errorf("received asrMetric: %v", asrMetric)
	} else if len(asrMetric.Events) != 1 {
		t.Errorf("unexpected Events in asrMetric: %+v", asrMetric.Events)
	}
	sq.remEventWithID("cgrates.org:TestRemEventWithID_5") // non existent
	if asr := asrMetric.GetValue(); asr.Compare(utils.NewDecimal(0, 0)) != 0 {
		t.Errorf("received asrMetric: %v", asrMetric)
	} else if len(asrMetric.Events) != 1 {
		t.Errorf("unexpected Events in asrMetric: %+v", asrMetric.Events)
	}
	sq.remEventWithID("cgrates.org:TestRemEventWithID_2")
	if asr := asrMetric.GetValue(); asr.Compare(utils.DecimalNaN) != 0 {
		t.Errorf("received asrMetric: %v", asrMetric)
	} else if len(asrMetric.Events) != 0 {
		t.Errorf("unexpected Events in asrMetric: %+v", asrMetric.Events)
	}
	sq.remEventWithID("cgrates.org:TestRemEventWithID_2")
	if asr := asrMetric.GetValue(); asr.Compare(utils.DecimalNaN) != 0 {
		t.Errorf("received asrMetric: %v", asrMetric)
	} else if len(asrMetric.Events) != 0 {
		t.Errorf("unexpected Events in asrMetric: %+v", asrMetric.Events)
	}
}

func TestStatRemEventWithID2(t *testing.T) {
	sq = &StatQueue{
		SQMetrics: map[string]StatMetric{
			utils.MetaASR: &StatASR{
				Metric: &Metric{
					Value: utils.NewDecimal(2, 0),
					Count: 4,
					Events: map[string]*DecimalWithCompress{
						"cgrates.org:TestRemEventWithID_1": {Stat: utils.NewDecimal(1, 0), CompressFactor: 2},
						"cgrates.org:TestRemEventWithID_2": {Stat: utils.NewDecimal(0, 0), CompressFactor: 2},
					},
				},
			},
		},
	}
	asrMetric := sq.SQMetrics[utils.MetaASR].(*StatASR)
	if asr := asrMetric.GetValue(); asr.Compare(utils.NewDecimal(50, 0)) != 0 {
		t.Errorf("received asrMetric: %v", asrMetric)
	}
	sq.remEventWithID("cgrates.org:TestRemEventWithID_1")
	sq.remEventWithID("cgrates.org:TestRemEventWithID_2")
	if asr := asrMetric.GetValue(); asr.Compare(utils.NewDecimal(50, 0)) != 0 {
		t.Errorf("received asrMetric: %v", asrMetric)
	} else if len(asrMetric.Events) != 2 {
		t.Errorf("unexpected Events in asrMetric: %+v", asrMetric.Events)
	}
	sq.remEventWithID("cgrates.org:TestRemEventWithID_5") // non existent
	if asr := asrMetric.GetValue(); asr.Compare(utils.NewDecimal(50, 0)) != 0 {
		t.Errorf("received asrMetric: %v", asrMetric)
	} else if len(asrMetric.Events) != 2 {
		t.Errorf("unexpected Events in asrMetric: %+v", asrMetric.Events)
	}
	sq.remEventWithID("cgrates.org:TestRemEventWithID_2")
	sq.remEventWithID("cgrates.org:TestRemEventWithID_1")
	if asr := asrMetric.GetValue(); asr.Compare(utils.DecimalNaN) != 0 {
		t.Errorf("received asrMetric: %v", asrMetric)
	} else if len(asrMetric.Events) != 0 {
		t.Errorf("unexpected Events in asrMetric: %+v", asrMetric.Events)
	}
	sq.remEventWithID("cgrates.org:TestRemEventWithID_2")
	if asr := asrMetric.GetValue(); asr.Compare(utils.DecimalNaN) != 0 {
		t.Errorf("received asrMetric: %v", asrMetric)
	} else if len(asrMetric.Events) != 0 {
		t.Errorf("unexpected Events in asrMetric: %+v", asrMetric.Events)
	}
}

func TestStatRemExpired(t *testing.T) {
	sq = &StatQueue{
		SQMetrics: map[string]StatMetric{
			utils.MetaASR: &StatASR{
				Metric: &Metric{
					Value: utils.NewDecimal(2, 0),
					Count: 3,
					Events: map[string]*DecimalWithCompress{
						"cgrates.org:TestStatRemExpired_1": {Stat: utils.NewDecimal(1, 0), CompressFactor: 1},
						"cgrates.org:TestStatRemExpired_2": {Stat: utils.NewDecimal(0, 0), CompressFactor: 1},
						"cgrates.org:TestStatRemExpired_3": {Stat: utils.NewDecimal(1, 0), CompressFactor: 1},
					},
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
	if asr := asrMetric.GetValue(); asr.Compare(utils.NewDecimalFromFloat64(66.66666666666667)) != 0 {
		t.Errorf("received asrMetric: %v", asrMetric.GetValue())
	}
	sq.remExpired()
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
				Metric: &Metric{
					Value: utils.NewDecimal(1, 0),
					Count: 1,
					Events: map[string]*DecimalWithCompress{
						"cgrates.org:TestStatRemExpired_1": {Stat: utils.NewDecimalFromFloat64(1), CompressFactor: 1},
					},
				},
			},
		},
		sqPrfl: &StatQueueProfile{
			Metrics: []*MetricWithFilters{
				{
					MetricID: utils.MetaASR,
				},
			},
		},
	}
	asrMetric := sq.SQMetrics[utils.MetaASR].(*StatASR)
	if asr := asrMetric.GetValue(); asr.Compare(utils.NewDecimalFromFloat64(100)) != 0 {
		t.Errorf("received ASR: %v", asr)
	}
	ev1 := &utils.CGREvent{Tenant: "cgrates.org", ID: "TestStatAddStatEvent_1"}
	sq.addStatEvent(context.Background(), ev1.Tenant, ev1.ID, nil, utils.MapStorage{utils.MetaOpts: ev1.Event})
	if asr := asrMetric.GetValue(); asr.Compare(utils.NewDecimalFromFloat64(50)) != 0 {
		t.Errorf("received ASR: %v", asr)
	} else if asrMetric.Value.Compare(utils.NewDecimal(1, 0)) != 0 || asrMetric.Count != 2 {
		t.Errorf("ASR: %v", asrMetric)
	}
	/*
		ev1.Event = map[string]any{
			utils.AnswerTime: time.Now()}
	*/
	ev1.APIOpts = map[string]any{
		utils.MetaStartTime: time.Now()}
	sq.addStatEvent(context.Background(), ev1.Tenant, ev1.ID, nil, utils.MapStorage{utils.MetaOpts: ev1.APIOpts})
	if asr := asrMetric.GetValue(); asr.Compare(utils.NewDecimalFromFloat64(66.66666666666667)) != 0 {
		t.Errorf("received ASR: %v", asr)
	} else if asrMetric.Value.Compare(utils.NewDecimal(2, 0)) != 0 || asrMetric.Count != 3 {
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
				Metric: &Metric{
					FilterIDs: []string{"*string:~*req.Account:1002"},
					Value:     utils.NewDecimal(0, 0),
					Events: map[string]*DecimalWithCompress{
						"cgrates.org:TestStatRemExpired_2": {Stat: utils.NewDecimalFromFloat64(float64(time.Minute)), CompressFactor: 1},
					},
				},
			},
			utils.MetaASR: &StatASR{
				Metric: &Metric{
					FilterIDs: []string{"*string:~*req.Account:1001"},
					Value:     utils.NewDecimal(0, 0),
					Events: map[string]*DecimalWithCompress{
						"cgrates.org:TestStatRemExpired_1": {Stat: utils.NewDecimalFromFloat64(1), CompressFactor: 1},
					},
				},
			},
		}}
	sq.remOnQueueLength()
	if len(sq.SQItems) != 1 {
		t.Errorf("wrong items: %+v", utils.ToJSON(sq.SQItems))
	}
}

func TestStatCompress(t *testing.T) {
	asr := &StatASR{
		Metric: &Metric{
			Value: utils.NewDecimal(2, 0),
			Count: 4,
			Events: map[string]*DecimalWithCompress{
				"cgrates.org:TestStatRemExpired_1": {Stat: utils.NewDecimalFromFloat64(1), CompressFactor: 1},
				"cgrates.org:TestStatRemExpired_2": {Stat: utils.NewDecimalFromFloat64(0), CompressFactor: 1},
				"cgrates.org:TestStatRemExpired_3": {Stat: utils.NewDecimalFromFloat64(1), CompressFactor: 1},
				"cgrates.org:TestStatRemExpired_4": {Stat: utils.NewDecimalFromFloat64(0), CompressFactor: 1},
			},
		},
	}
	expectedASR := &StatASR{
		Metric: &Metric{
			Value: utils.NewDecimal(2, 0),
			Count: 4,
			Events: map[string]*DecimalWithCompress{
				"cgrates.org:TestStatRemExpired_4": {Stat: utils.NewDecimalFromFloat64(0.5), CompressFactor: 4},
			},
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
	if sq.Compress(100) {
		t.Errorf("StatQueue compressed: %s", utils.ToJSON(sq))
	}
	if !sq.Compress(2) {
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
		Metric: &Metric{
			Value: utils.NewDecimal(2, 0),
			Count: 4,
			Events: map[string]*DecimalWithCompress{
				"cgrates.org:TestStatRemExpired_1": {Stat: utils.NewDecimalFromFloat64(1), CompressFactor: 1},
				"cgrates.org:TestStatRemExpired_2": {Stat: utils.NewDecimalFromFloat64(0), CompressFactor: 1},
				"cgrates.org:TestStatRemExpired_3": {Stat: utils.NewDecimalFromFloat64(1), CompressFactor: 1},
				"cgrates.org:TestStatRemExpired_4": {Stat: utils.NewDecimalFromFloat64(0), CompressFactor: 1},
			},
		},
	}
	expectedASR := &StatASR{
		Metric: &Metric{
			Value: utils.NewDecimal(2, 0),
			Count: 4,
			Events: map[string]*DecimalWithCompress{
				"cgrates.org:TestStatRemExpired_4": {Stat: utils.NewDecimalFromFloat64(0.5), CompressFactor: 4},
			},
		},
	}
	tcd := &StatTCD{
		Metric: &Metric{
			Value: utils.NewDecimal(int64(3*time.Minute), 0),
			Count: 2,
			Events: map[string]*DecimalWithCompress{
				"cgrates.org:TestStatRemExpired_2": {Stat: utils.NewDecimal(int64(time.Minute), 0), CompressFactor: 1},
				"cgrates.org:TestStatRemExpired_3": {Stat: utils.NewDecimal(int64(2*time.Minute), 0), CompressFactor: 1},
			},
		},
	}
	expectedTCD := &StatTCD{
		Metric: &Metric{
			Value: utils.NewDecimal(int64(3*time.Minute), 0),
			Count: 2,
			Events: map[string]*DecimalWithCompress{
				"cgrates.org:TestStatRemExpired_4": {Stat: utils.NewDecimalFromFloat64(float64(time.Minute + 30*time.Second)), CompressFactor: 2},
			},
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
	if sq.Compress(100) {
		t.Errorf("StatQueue compressed: %s", utils.ToJSON(sq))
	}
	if !sq.Compress(2) {
		t.Errorf("StatQueue not compressed: %s", utils.ToJSON(sq))
	}
	if !reflect.DeepEqual(sq.SQItems, expectedSqItems) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expectedSqItems), utils.ToJSON(sq.SQItems))
	}
	if rply := sq.SQMetrics[utils.MetaASR].(*StatASR); !reflect.DeepEqual(*rply, *expectedASR) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expectedASR), utils.ToJSON(rply))
	}
	if rply := sq.SQMetrics[utils.MetaTCD].(*StatTCD); !rply.Equal(expectedTCD.Metric) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expectedTCD), utils.ToJSON(rply))
	}
}

func TestStatCompress3(t *testing.T) {
	tmNow := time.Now()
	asr := &StatASR{
		Metric: &Metric{
			Value: utils.NewDecimal(2, 0),
			Count: 4,
			Events: map[string]*DecimalWithCompress{
				"cgrates.org:TestStatRemExpired_1": {Stat: utils.NewDecimalFromFloat64(1), CompressFactor: 1},
				"cgrates.org:TestStatRemExpired_2": {Stat: utils.NewDecimalFromFloat64(0), CompressFactor: 1},
				"cgrates.org:TestStatRemExpired_3": {Stat: utils.NewDecimalFromFloat64(1), CompressFactor: 1},
				"cgrates.org:TestStatRemExpired_4": {Stat: utils.NewDecimalFromFloat64(0), CompressFactor: 1},
			},
		},
	}
	expectedASR := &StatASR{
		Metric: &Metric{
			Value: utils.NewDecimal(2, 0),
			Count: 4,
			Events: map[string]*DecimalWithCompress{
				"cgrates.org:TestStatRemExpired_4": {Stat: utils.NewDecimalFromFloat64(0.5), CompressFactor: 4},
			},
		},
	}
	tcd := &StatTCD{
		Metric: &Metric{
			Value: utils.NewDecimal(int64(3*time.Minute), 0),
			Count: 2,
			Events: map[string]*DecimalWithCompress{
				"cgrates.org:TestStatRemExpired_2": {Stat: utils.NewDecimal(int64(time.Minute), 0), CompressFactor: 1},
				"cgrates.org:TestStatRemExpired_3": {Stat: utils.NewDecimal(int64(2*time.Minute), 0), CompressFactor: 1},
			},
		},
	}
	expectedTCD := &StatTCD{
		Metric: &Metric{
			Value: utils.NewDecimal(int64(3*time.Minute), 0),
			Count: 2,
			Events: map[string]*DecimalWithCompress{
				"cgrates.org:TestStatRemExpired_2": {Stat: utils.NewDecimal(int64(time.Minute), 0), CompressFactor: 1},
				"cgrates.org:TestStatRemExpired_3": {Stat: utils.NewDecimal(int64(2*time.Minute), 0), CompressFactor: 1},
			},
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
	if sq.Compress(100) {
		t.Errorf("StatQueue compressed: %s", utils.ToJSON(sq))
	}
	if !sq.Compress(3) {
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
		Metric: &Metric{
			Value: utils.NewDecimal(2, 0),
			Count: 4,
			Events: map[string]*DecimalWithCompress{
				"cgrates.org:TestStatRemExpired_4": {Stat: utils.NewDecimalFromFloat64(0.5), CompressFactor: 4},
			},
		},
	}
	asr := &StatASR{
		Metric: &Metric{
			Value: utils.NewDecimal(2, 0),
			Count: 4,
			Events: map[string]*DecimalWithCompress{
				"cgrates.org:TestStatRemExpired_4": {Stat: utils.NewDecimalFromFloat64(0.5), CompressFactor: 4},
			},
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
		Metric: &Metric{
			Value: utils.NewDecimal(2, 0),
			Count: 4,
			Events: map[string]*DecimalWithCompress{
				"cgrates.org:TestStatRemExpired_4": {Stat: utils.NewDecimalFromFloat64(0.5), CompressFactor: 4},
			},
		},
	}
	asr := &StatASR{
		Metric: &Metric{
			Value: utils.NewDecimal(2, 0),
			Count: 4,
			Events: map[string]*DecimalWithCompress{
				"cgrates.org:TestStatRemExpired_4": {Stat: utils.NewDecimalFromFloat64(0.5), CompressFactor: 4},
			},
		},
	}
	expectedTCD := &StatTCD{
		Metric: &Metric{
			Value: utils.NewDecimal(int64(3*time.Minute), 0),
			Count: 2,
			Events: map[string]*DecimalWithCompress{
				"cgrates.org:TestStatRemExpired_4": {Stat: utils.NewDecimal(int64(time.Minute+30*time.Second), 0), CompressFactor: 2},
			},
		},
	}
	tcd := &StatTCD{
		Metric: &Metric{
			Value: utils.NewDecimal(int64(3*time.Minute), 0),
			Count: 2,
			Events: map[string]*DecimalWithCompress{
				"cgrates.org:TestStatRemExpired_4": {Stat: utils.NewDecimal(int64(time.Minute+30*time.Second), 0), CompressFactor: 2},
			},
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
		Metric: &Metric{
			Value: utils.NewDecimal(2, 0),
			Count: 4,
			Events: map[string]*DecimalWithCompress{
				"cgrates.org:TestStatRemExpired_3": {Stat: utils.NewDecimalFromFloat64(0.5), CompressFactor: 4},
			},
		},
	}
	asr := &StatASR{
		Metric: &Metric{
			Value: utils.NewDecimal(2, 0),
			Count: 4,
			Events: map[string]*DecimalWithCompress{
				"cgrates.org:TestStatRemExpired_3": {Stat: utils.NewDecimalFromFloat64(0.5), CompressFactor: 4},
			},
		},
	}
	expectedTCD := &StatTCD{
		Metric: &Metric{
			Value: utils.NewDecimal(int64(3*time.Minute), 0),
			Count: 2,
			Events: map[string]*DecimalWithCompress{
				"cgrates.org:TestStatRemExpired_2": {Stat: utils.NewDecimal(int64(time.Minute), 0), CompressFactor: 1},
				"cgrates.org:TestStatRemExpired_4": {Stat: utils.NewDecimal(int64(2*time.Minute), 0), CompressFactor: 1},
			},
		},
	}
	tcd := &StatTCD{
		Metric: &Metric{
			Value: utils.NewDecimal(int64(3*time.Minute), 0),
			Count: 2,
			Events: map[string]*DecimalWithCompress{
				"cgrates.org:TestStatRemExpired_2": {Stat: utils.NewDecimal(int64(time.Minute), 0), CompressFactor: 1},
				"cgrates.org:TestStatRemExpired_4": {Stat: utils.NewDecimal(int64(2*time.Minute), 0), CompressFactor: 1},
			},
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
				Metric: &Metric{
					Value: utils.NewDecimal(1, 0),
					Count: 1,
					Events: map[string]*DecimalWithCompress{
						"grates.org:TestStatRemExpired_1": {Stat: utils.NewDecimal(1, 0), CompressFactor: 1},
					},
				},
			},
		},
		sqPrfl: &StatQueueProfile{
			QueueLength: 0, //unlimited que
		},
	}

	//add ev1 with ttl 100ms (after 100ms the event should be removed)
	ev1 := &utils.CGREvent{Tenant: "cgrates.org", ID: "TestStatAddStatEvent_1"}
	sq.ProcessEvent(context.Background(), ev1.Tenant, ev1.ID, nil, utils.MapStorage{utils.MetaReq: ev1.Event})

	if len(sq.SQItems) != 1 && sq.SQItems[0].EventID != "TestStatAddStatEvent_1" {
		t.Errorf("Expecting: 1, received: %+v", len(sq.SQItems))
	}
	//after 150ms the event expired
	time.Sleep(150 * time.Millisecond)

	//processing a new event should clean the expired events and add the new one
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "TestStatAddStatEvent_2"}
	sq.ProcessEvent(context.Background(), ev2.Tenant, ev2.ID, nil, utils.MapStorage{utils.MetaReq: ev2.Event})
	if len(sq.SQItems) != 1 && sq.SQItems[0].EventID != "TestStatAddStatEvent_2" {
		t.Errorf("Expecting: 1, received: %+v", len(sq.SQItems))
	}
}

func TestStatRemoveExpiredQueue(t *testing.T) {
	sq = &StatQueue{
		SQMetrics: map[string]StatMetric{
			utils.MetaASR: &StatASR{
				Metric: &Metric{
					Value: utils.NewDecimal(1, 0),
					Count: 1,
					Events: map[string]*DecimalWithCompress{
						"grates.org:TestStatRemExpired_1": {Stat: utils.NewDecimal(1, 0), CompressFactor: 1},
					},
				},
			},
		},
		sqPrfl: &StatQueueProfile{
			QueueLength: 2, //unlimited que
		},
	}

	//add ev1 with ttl 100ms (after 100ms the event should be removed)
	ev1 := &utils.CGREvent{Tenant: "cgrates.org", ID: "TestStatAddStatEvent_1"}
	sq.ProcessEvent(context.Background(), ev1.Tenant, ev1.ID, nil, utils.MapStorage{utils.MetaReq: ev1.Event})

	if len(sq.SQItems) != 1 && sq.SQItems[0].EventID != "TestStatAddStatEvent_1" {
		t.Errorf("Expecting: 1, received: %+v", len(sq.SQItems))
	}
	//after 150ms the event expired
	time.Sleep(150 * time.Millisecond)

	//processing a new event should clean the expired events and add the new one
	ev2 := &utils.CGREvent{Tenant: "cgrates.org", ID: "TestStatAddStatEvent_2"}
	sq.ProcessEvent(context.Background(), ev2.Tenant, ev2.ID, nil, utils.MapStorage{utils.MetaReq: ev2.Event})
	if len(sq.SQItems) != 2 && sq.SQItems[0].EventID != "TestStatAddStatEvent_1" &&
		sq.SQItems[1].EventID != "TestStatAddStatEvent_2" {
		t.Errorf("Expecting: 2, received: %+v", len(sq.SQItems))
	}

	//processing a new event should clean the expired events and add the new one
	ev3 := &utils.CGREvent{Tenant: "cgrates.org", ID: "TestStatAddStatEvent_3"}
	sq.ProcessEvent(context.Background(), ev3.Tenant, ev3.ID, nil, utils.MapStorage{utils.MetaReq: ev3.Event})
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
func (sMM statMetricMock) Clone() StatMetric {
	return sMM
}

type mockMarshal string

func (m mockMarshal) Marshal(v any) ([]byte, error)      { return nil, errors.New(string(m)) }
func (m mockMarshal) Unmarshal(data []byte, v any) error { return errors.New(string(m)) }
func TestStatQueueNewStoredStatQueue(t *testing.T) {
	sq := &StatQueue{
		SQMetrics: map[string]StatMetric{
			"key": statMetricMock(""),
		},
	}
	experr := "marshal mock error"
	var ms utils.Marshaler = mockMarshal(experr)

	rcv, err := NewStoredStatQueue(sq, ms)

	if err == nil || err.Error() != experr {
		t.Fatalf("\nreceived: %q, \nexpected: %q", experr, err)
	}

	if rcv != nil {
		t.Errorf("\nreceived: <%+v>, \nexpected: <%+v>", nil, rcv)
	}
}

func TestStatQueueAsStatQueueNilStoredSq(t *testing.T) {
	var ssq *StoredStatQueue
	var ms utils.Marshaler

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
	var ms utils.Marshaler

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
	var ms utils.Marshaler

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
	ms, err := utils.NewMarshaler(utils.JSON)
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
	ms, err := utils.NewMarshaler(utils.JSON)
	if err != nil {
		t.Fatal(err)
	}

	sm, err := NewStatMetric(utils.MetaTCD, 0, []string{})
	if err != nil {
		t.Fatal(err)
	}

	msm, err := ms.Marshal(sm)
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
	var minItems uint64

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

func TestStatQueueProcessEventremExpiredErr(t *testing.T) {
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
			"key": statMetricMock("remExpired error"),
		},
	}

	experr := "remExpired mock error"
	err := sq.ProcessEvent(context.Background(), tnt, evID, filters, evNm)

	if err == nil || err.Error() != experr {
		t.Errorf("\nexpected: %q, \nreceived: %q", experr, err)
	}
}

func TestStatQueueProcessEventremOnQueueLengthErr(t *testing.T) {
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
			"key": statMetricMock("remExpired error"),
		},
	}

	experr := "remExpired mock error"
	err := sq.ProcessEvent(context.Background(), tnt, evID, filters, evNm)

	if err == nil || err.Error() != experr {
		t.Errorf("\nexpected: %q, \nreceived: %q", experr, err)
	}
}

func TestStatQueueProcessEventaddStatEvent(t *testing.T) {
	tnt, evID := "tenant", "eventID"
	filters := &FilterS{}
	evNm := utils.MapStorage{
		"key": nil,
	}

	sq := &StatQueue{
		sqPrfl: &StatQueueProfile{
			QueueLength: 1,
			Metrics: []*MetricWithFilters{
				{
					MetricID: utils.MetaTCD,
				},
			},
		},
		SQItems: []SQItem{
			{
				EventID: evID,
			},
		},
		SQMetrics: map[string]StatMetric{
			utils.MetaTCD: &StatTCD{Metric: &Metric{}},
		},
	}

	experr := utils.ErrWrongPath
	err := sq.ProcessEvent(context.Background(), tnt, evID, filters, evNm)

	if err == nil || err != experr {
		t.Errorf("\nexpected: %q, \nreceived: %q", experr, err)
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
			utils.MetaTCD: statMetricMock("populate idMap"),
			utils.MetaReq: sm,
		},
		ttl: &ttl,
	}

	maxQL := uint64(1)

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
	rcv := sq.Compress(maxQL)
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

func TestStatQueueaddStatEventNoPass(t *testing.T) {
	sm, err := NewStatMetric(utils.MetaTCD, 0, []string{"*string:~*req.Account:1001"})
	if err != nil {
		t.Fatal(err)
	}

	sq := &StatQueue{
		SQMetrics: map[string]StatMetric{
			utils.MetaTCD: sm,
		},
		sqPrfl: &StatQueueProfile{
			Metrics: []*MetricWithFilters{
				{
					FilterIDs: []string{"*string:~*req.Account:1001"},
					MetricID:  utils.MetaTCD,
				},
			},
		},
	}
	sq.lock(utils.EmptyString)

	tnt, evID := "cgrates.org", "eventID"
	filters := &FilterS{
		cfg: config.CgrConfig(),
		dm: &DataManager{
			dataDB: NewInternalDB(nil, nil, config.CgrConfig().DataDbCfg().Items),
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
		sqPrfl: &StatQueueProfile{
			Metrics: []*MetricWithFilters{
				{
					MetricID:  utils.MetaTCD,
					FilterIDs: []string{"*string:~*req.Account:1001"},
				},
			},
		},
	}
	err = sq.addStatEvent(context.Background(), tnt, evID, filters, evNm)
	sq.unlock()

	if err != nil {
		t.Fatalf("\nexpected: <%+v>, \nreceived: <%+v>", nil, err)
	}

	if !reflect.DeepEqual(sq, exp) {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", exp.sqPrfl, sq.sqPrfl)
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
	rply := &StatQueueWithAPIOpts{}
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
		t.Errorf("Expected: %+v , received: %+v", exp2, rply)
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(exp2), utils.ToJSON(rply))
	}

}

func TestStatQueueLockUnlockStatQueueProfiles(t *testing.T) {
	sqPrf := &StatQueueProfile{
		Tenant: "cgrates.org",
		ID:     "SQ1",
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
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

func TestStatQueueProfileSet(t *testing.T) {
	sq := StatQueueProfile{}
	exp := StatQueueProfile{
		Tenant:    "cgrates.org",
		ID:        "ID",
		FilterIDs: []string{"fltr1", "*string:~*req.Account:1001"},
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
		QueueLength:  10,
		TTL:          10,
		MinItems:     10,
		Stored:       true,
		Blockers:     utils.DynamicBlockers{{Blocker: true}},
		ThresholdIDs: []string{"TH1"},
		Metrics: []*MetricWithFilters{{
			MetricID: utils.MetaTCD,
		}, {
			MetricID:  utils.MetaACD,
			FilterIDs: []string{"fltr1"},
		}},
	}
	if err := sq.Set([]string{}, "", false, utils.EmptyString); err != utils.ErrWrongPath {
		t.Error(err)
	}
	if err := sq.Set([]string{""}, "", false, utils.EmptyString); err != nil {
		t.Error(err)
	}
	if err := sq.Set([]string{"NotAField"}, ";", false, utils.EmptyString); err != utils.ErrWrongPath {
		t.Error(err)
	}
	if err := sq.Set([]string{"NotAField", "1"}, ";", false, utils.EmptyString); err != utils.ErrWrongPath {
		t.Error(err)
	}

	if err := sq.Set([]string{utils.Tenant}, "cgrates.org", false, utils.EmptyString); err != nil {
		t.Error(err)
	}
	if err := sq.Set([]string{utils.ID}, "ID", false, utils.EmptyString); err != nil {
		t.Error(err)
	}
	if err := sq.Set([]string{utils.FilterIDs}, "fltr1;*string:~*req.Account:1001", false, utils.EmptyString); err != nil {
		t.Error(err)
	}
	if err := sq.Set([]string{utils.Weights}, ";10", false, utils.EmptyString); err != nil {
		t.Error(err)
	}
	if err := sq.Set([]string{utils.QueueLength}, 10, false, utils.EmptyString); err != nil {
		t.Error(err)
	}
	if err := sq.Set([]string{utils.TTL}, 10, false, utils.EmptyString); err != nil {
		t.Error(err)
	}
	if err := sq.Set([]string{utils.MinItems}, 10, false, utils.EmptyString); err != nil {
		t.Error(err)
	}
	if err := sq.Set([]string{utils.Stored}, true, false, utils.EmptyString); err != nil {
		t.Error(err)
	}
	if err := sq.Set([]string{utils.Blockers}, ";true", false, utils.EmptyString); err != nil {
		t.Error(err)
	}
	if err := sq.Set([]string{utils.ThresholdIDs}, "TH1", false, utils.EmptyString); err != nil {
		t.Error(err)
	}

	if err := sq.Set([]string{utils.Metrics, utils.MetricID}, "*tcd;*acd", false, utils.EmptyString); err != nil {
		t.Error(err)
	}
	if err := sq.Set([]string{utils.Metrics, utils.FilterIDs}, "fltr1", false, utils.EmptyString); err != nil {
		t.Error(err)
	}

	if err := sq.Set([]string{utils.Metrics, "wrong"}, "fltr1", false, utils.EmptyString); err != utils.ErrWrongPath {
		t.Error(err)
	}
	if !reflect.DeepEqual(exp, sq) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(sq))
	}
}

func TestStatQueueProfileAsInterface(t *testing.T) {
	sqp := StatQueueProfile{
		Tenant:    "cgrates.org",
		ID:        "ID",
		FilterIDs: []string{"fltr1", "*string:~*req.Account:1001"},
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
		QueueLength:  10,
		TTL:          10,
		MinItems:     10,
		Stored:       true,
		Blockers:     utils.DynamicBlockers{{Blocker: true}},
		ThresholdIDs: []string{"TH1"},
		Metrics: []*MetricWithFilters{{
			MetricID: utils.MetaTCD,
		}, {
			MetricID:  utils.MetaACD,
			FilterIDs: []string{"fltr1"},
		}, {

			Blockers: utils.DynamicBlockers{{Blocker: true}},
		}},
	}
	if _, err := sqp.FieldAsInterface(nil); err != utils.ErrNotFound {
		t.Fatal(err)
	}
	if _, err := sqp.FieldAsInterface([]string{"field"}); err != utils.ErrNotFound {
		t.Fatal(err)
	}
	if _, err := sqp.FieldAsInterface([]string{"field", ""}); err != utils.ErrNotFound {
		t.Fatal(err)
	}
	if val, err := sqp.FieldAsInterface([]string{utils.Tenant}); err != nil {
		t.Fatal(err)
	} else if exp := "cgrates.org"; exp != val {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(val))
	}
	if val, err := sqp.FieldAsInterface([]string{utils.ID}); err != nil {
		t.Fatal(err)
	} else if exp := utils.ID; exp != val {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(val))
	}
	if val, err := sqp.FieldAsInterface([]string{utils.FilterIDs}); err != nil {
		t.Fatal(err)
	} else if exp := sqp.FilterIDs; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(val))
	}
	if val, err := sqp.FieldAsInterface([]string{utils.FilterIDs + "[0]"}); err != nil {
		t.Fatal(err)
	} else if exp := sqp.FilterIDs[0]; exp != val {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(val))
	}
	if val, err := sqp.FieldAsInterface([]string{utils.Weights}); err != nil {
		t.Fatal(err)
	} else if exp := sqp.Weights; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(val))
	}
	if val, err := sqp.FieldAsInterface([]string{utils.ThresholdIDs}); err != nil {
		t.Fatal(err)
	} else if exp := sqp.ThresholdIDs; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(val))
	}
	if val, err := sqp.FieldAsInterface([]string{utils.ThresholdIDs + "[0]"}); err != nil {
		t.Fatal(err)
	} else if exp := sqp.ThresholdIDs[0]; exp != val {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(val))
	}
	if val, err := sqp.FieldAsInterface([]string{utils.Metrics}); err != nil {
		t.Fatal(err)
	} else if exp := sqp.Metrics; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(val))
	}
	if val, err := sqp.FieldAsInterface([]string{utils.Metrics + "[0]"}); err != nil {
		t.Fatal(err)
	} else if exp := sqp.Metrics[0]; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(val))
	}

	if val, err := sqp.FieldAsInterface([]string{utils.QueueLength}); err != nil {
		t.Fatal(err)
	} else if exp := sqp.QueueLength; exp != val {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(val))
	}
	if val, err := sqp.FieldAsInterface([]string{utils.TTL}); err != nil {
		t.Fatal(err)
	} else if exp := sqp.TTL; exp != val {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(val))
	}
	if val, err := sqp.FieldAsInterface([]string{utils.MinItems}); err != nil {
		t.Fatal(err)
	} else if exp := sqp.MinItems; exp != val {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(val))
	}
	if val, err := sqp.FieldAsInterface([]string{utils.Stored}); err != nil {
		t.Fatal(err)
	} else if exp := sqp.Stored; exp != val {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(val))
	}
	if val, err := sqp.FieldAsInterface([]string{utils.Blockers}); err != nil {
		t.Fatal(err)
	} else if exp := sqp.Blockers; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(val))
	}
	if _, err := sqp.FieldAsInterface([]string{utils.Metrics + "[4]"}); err != utils.ErrNotFound {
		t.Fatal(err)
	}
	if _, err := sqp.FieldAsInterface([]string{utils.Metrics + "4]"}); err != utils.ErrNotFound {
		t.Fatal(err)
	}
	if _, err := sqp.FieldAsInterface([]string{utils.Metrics + "[4]", ""}); err != utils.ErrNotFound {
		t.Fatal(err)
	}
	if _, err := sqp.FieldAsInterface([]string{utils.Metrics + "[0]", ""}); err != utils.ErrNotFound {
		t.Fatal(err)
	}
	if _, err := sqp.FieldAsInterface([]string{utils.Metrics + "[0]", "", ""}); err != utils.ErrNotFound {
		t.Fatal(err)
	}

	if val, err := sqp.FieldAsInterface([]string{utils.Metrics + "[0]", utils.MetricID}); err != nil {
		t.Fatal(err)
	} else if exp := sqp.Metrics[0].MetricID; exp != val {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(val))
	}

	if val, err := sqp.FieldAsInterface([]string{utils.Metrics + "[0]", utils.FilterIDs}); err != nil {
		t.Fatal(err)
	} else if exp := sqp.Metrics[0].FilterIDs; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(val))
	}
	if val, err := sqp.FieldAsInterface([]string{utils.Metrics + "[1]", utils.FilterIDs + "[0]"}); err != nil {
		t.Fatal(err)
	} else if exp := sqp.Metrics[1].FilterIDs[0]; exp != val {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(val))
	}

	if _, err := sqp.FieldAsString([]string{""}); err != utils.ErrNotFound {
		t.Fatal(err)
	}
	if val, err := sqp.FieldAsString([]string{utils.ID}); err != nil {
		t.Fatal(err)
	} else if exp := "ID"; exp != val {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(val))
	}
	if val, exp := sqp.String(), utils.ToJSON(sqp); exp != val {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(val))
	}

	if _, err := sqp.Metrics[0].FieldAsString([]string{""}); err != utils.ErrNotFound {
		t.Fatal(err)
	}
	if val, err := sqp.Metrics[0].FieldAsString([]string{utils.MetricID}); err != nil {
		t.Fatal(err)
	} else if exp := utils.MetaTCD; exp != val {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(val))
	}
	if val, exp := sqp.Metrics[0].String(), utils.ToJSON(sqp.Metrics[0]); exp != val {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(val))
	}

	if val, err := sqp.Metrics[2].FieldAsInterface([]string{utils.Blockers}); err != nil {
		t.Fatal(err)
	} else if exp := sqp.Metrics[2].Blockers; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(val))
	}
}

func TestStatQueueProfileMerge(t *testing.T) {
	sqp := &StatQueueProfile{}
	exp := &StatQueueProfile{
		Tenant:    "cgrates.org",
		ID:        "ID",
		FilterIDs: []string{"fltr1", "*string:~*req.Account:1001"},
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
		QueueLength:  10,
		TTL:          10,
		MinItems:     10,
		Stored:       true,
		Blockers:     utils.DynamicBlockers{{Blocker: true}},
		ThresholdIDs: []string{"TH1"},
		Metrics: []*MetricWithFilters{{
			MetricID: utils.MetaTCD,
		}, {
			MetricID:  utils.MetaACD,
			FilterIDs: []string{"fltr1"},
		}},
	}
	if sqp.Merge(&StatQueueProfile{
		Tenant:    "cgrates.org",
		ID:        "ID",
		FilterIDs: []string{"fltr1", "*string:~*req.Account:1001"},
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
		QueueLength:  10,
		TTL:          10,
		MinItems:     10,
		Stored:       true,
		Blockers:     utils.DynamicBlockers{{Blocker: true}},
		ThresholdIDs: []string{"TH1"},
		Metrics: []*MetricWithFilters{{
			MetricID: utils.MetaTCD,
		}, {
			MetricID:  utils.MetaACD,
			FilterIDs: []string{"fltr1"},
		}},
	}); !reflect.DeepEqual(exp, sqp) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(sqp))
	}
}

func TestStatQueueProfile_Set(t *testing.T) {
	type fields struct {
		Tenant       string
		ID           string
		FilterIDs    []string
		QueueLength  int
		TTL          time.Duration
		MinItems     int
		Metrics      []*MetricWithFilters
		Stored       bool
		Blocker      bool
		Weights      utils.DynamicWeights
		ThresholdIDs []string
		lkID         string
	}
	type args struct {
		path      []string
		val       any
		newBranch bool
		in3       string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sqp := &StatQueueProfile{
				Tenant:       tt.fields.Tenant,
				ID:           tt.fields.ID,
				FilterIDs:    tt.fields.FilterIDs,
				QueueLength:  tt.fields.QueueLength,
				TTL:          tt.fields.TTL,
				MinItems:     tt.fields.MinItems,
				Metrics:      tt.fields.Metrics,
				Stored:       tt.fields.Stored,
				Blockers:     utils.DynamicBlockers{{Blocker: true}},
				Weights:      tt.fields.Weights,
				ThresholdIDs: tt.fields.ThresholdIDs,
				lkID:         tt.fields.lkID,
			}
			if err := sqp.Set(tt.args.path, tt.args.val, tt.args.newBranch, tt.args.in3); err != nil != tt.wantErr {
				t.Errorf("StatQueueProfile.Set() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestStatQueueGobEncode(t *testing.T) {
	exp := []byte{67, 127, 3, 1, 1, 8, 115, 113, 69, 110, 99, 111, 100, 101, 1, 255, 128, 0, 1, 4, 1, 6, 84, 101, 110, 97, 110, 116, 1, 12, 0, 1, 2, 73, 68, 1, 12, 0, 1, 7, 83, 81, 73, 116, 101, 109, 115, 1, 255, 134, 0, 1, 9, 83, 81, 77, 101, 116, 114, 105, 99, 115, 1, 255, 136, 0, 0, 0, 30, 255, 133, 2, 1, 1, 15, 91, 93, 101, 110, 103, 105, 110, 101, 46, 83, 81, 73, 116, 101, 109, 1, 255, 134, 0, 1, 255, 130, 0, 0, 48, 255, 129, 3, 1, 1, 6, 83, 81, 73, 116, 101, 109, 1, 255, 130, 0, 1, 2, 1, 7, 69, 118, 101, 110, 116, 73, 68, 1, 12, 0, 1, 10, 69, 120, 112, 105, 114, 121, 84, 105, 109, 101, 1, 255, 132, 0, 0, 0, 10, 255, 131, 5, 1, 2, 255, 138, 0, 0, 0, 44, 255, 135, 4, 1, 1, 28, 109, 97, 112, 91, 115, 116, 114, 105, 110, 103, 93, 101, 110, 103, 105, 110, 101, 46, 83, 116, 97, 116, 77, 101, 116, 114, 105, 99, 1, 255, 136, 0, 1, 12, 1, 16, 0, 0, 3, 255, 128, 0}
	sq := &StatQueue{}
	if rcv, err := sq.GobEncode(); err != nil {
		t.Error(err)
	} else if string(rcv) != string(exp) {
		t.Errorf("Expected <%v>, \nReceived <%v>", string(exp), string(rcv))
	}
}
func TestStatQueueGobDecode(t *testing.T) {
	rply := []byte{77, 1}
	expErr := "unexpected EOF"
	sq := &StatQueue{}
	if err := sq.GobDecode(rply); err == nil || err.Error() != expErr {
		t.Errorf("Expected error <%v>, Received error <%v>", expErr, err.Error())
	}
}

func TestStatQueueClone(t *testing.T) {
	exTime := time.Date(2021, 1, 1, 23, 59, 59, 0, time.UTC)
	sq := &StatQueue{
		Tenant: "testTnt",
		ID:     "testId",
		SQItems: []SQItem{
			{
				EventID:    "testEventId",
				ExpiryTime: &exTime,
			},
		},
		SQMetrics: map[string]StatMetric{
			"key": statMetricMock("remExpired error"),
		},
		lkID:   "testLkId",
		dirty:  utils.BoolPointer(false),
		ttl:    utils.DurationPointer(time.Duration(3)),
		weight: 2,
	}
	exp := &StatQueue{
		Tenant: "testTnt",
		ID:     "testId",
		SQItems: []SQItem{
			{
				EventID:    "testEventId",
				ExpiryTime: &exTime,
			},
		},
		SQMetrics: map[string]StatMetric{
			"key": statMetricMock("remExpired error"),
		},
		lkID:   "testLkId",
		dirty:  utils.BoolPointer(false),
		ttl:    utils.DurationPointer(time.Duration(3)),
		weight: 2,
	}
	if rcv := sq.Clone(); !reflect.DeepEqual(utils.ToJSON(rcv), utils.ToJSON(exp)) {
		t.Errorf("Expected <%v>, \nReceived <%v>", utils.ToJSON(exp), utils.ToJSON(rcv))
	}
}

func TestStatQueueWithAPIOptsMarshalJSONNil(t *testing.T) {
	var ssq *StatQueueWithAPIOpts
	if _, err := ssq.MarshalJSON(); err != nil {
		t.Errorf("Expected error <nil>, Received error <%v>", err)
	}

}

func TestStatQueueUnmarshalJSONErrUnmarsheling(t *testing.T) {
	sq := &StatQueue{}
	expErr := "invalid character 'Ô' looking for beginning of value"
	if err := sq.UnmarshalJSON([]byte{212}); err == nil || err.Error() != expErr {
		t.Errorf("Expected error <%v>, Received error <%v>", expErr, err)
	}

}

func TestStatQueueWithAPIOptsUnmarshalJSONErrWithSSQ(t *testing.T) {
	ssq := &StatQueueWithAPIOpts{}
	expErr := "invalid character 'Ô' looking for beginning of value"
	if err := ssq.UnmarshalJSON([]byte{212}); err == nil || err.Error() != expErr {
		t.Errorf("Expected error <%v>, Received error <%v>", expErr, err)
	}

}

func TestStatQueueProfileSetBlockersErr(t *testing.T) {
	sq := StatQueueProfile{}

	expErr := "invalid DynamicBlocker format for string <incorrect input>"
	if err := sq.Set([]string{utils.Metrics, utils.Blockers}, "incorrect input", false, utils.EmptyString); err == nil || err.Error() != expErr {
		t.Errorf("Expected error <%v>, Received error <%v>", expErr, err)
	}

}

func TestStatQueueProfileSetBlockersOK(t *testing.T) {
	sq := StatQueueProfile{}

	exp := StatQueueProfile{
		Metrics: []*MetricWithFilters{
			{
				Blockers: utils.DynamicBlockers{&utils.DynamicBlocker{
					FilterIDs: []string{"*string:~*opts.*cost:0"},
					Blocker:   false,
				},
					&utils.DynamicBlocker{FilterIDs: []string{"*suffix:~*req.Destination:+4432", "eq:~*opts.*usage:10s"},
						Blocker: false},
					&utils.DynamicBlocker{FilterIDs: []string{"*notstring:~*req.RequestType:*prepaid"},
						Blocker: true},
					&utils.DynamicBlocker{FilterIDs: nil,
						Blocker: false},
				},
			},
		},
	}

	if err := sq.Set([]string{utils.Metrics, utils.Blockers}, "*string:~*opts.*cost:0;false;*suffix:~*req.Destination:+4432&eq:~*opts.*usage:10s;false;*notstring:~*req.RequestType:*prepaid;true;;false", false, utils.EmptyString); err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(exp, sq) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(sq))
	}
}

func TestStatQueueUnmarshalJSONOK(t *testing.T) {

	sqData := &StatQueue{
		Tenant: "cgrates.org",
		ID:     "STS",
		SQMetrics: map[string]StatMetric{
			utils.MetaASR:      new(StatASR),
			utils.MetaACD:      new(StatACD),
			utils.MetaTCD:      new(StatTCD),
			utils.MetaACC:      new(StatACC),
			utils.MetaTCC:      new(StatTCC),
			utils.MetaPDD:      new(StatPDD),
			utils.MetaDDC:      new(StatDDC),
			utils.MetaSum:      new(StatSum),
			utils.MetaAverage:  new(StatAverage),
			utils.MetaDistinct: new(StatDistinct),
		},
	}

	sq := &StatQueue{}

	data := []byte(utils.ToJSON(sqData))

	if err := sq.UnmarshalJSON(data); err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(utils.ToJSON(sqData), utils.ToJSON(sq)) {
		t.Errorf("Expected <%v>, \nReceived <%v>", utils.ToJSON(sqData), utils.ToJSON(sq))
	}

}

func TestStatQueueUnmarshalJSONBadMetric(t *testing.T) {

	sqData := &StatQueue{
		Tenant: "cgrates.org",
		ID:     "STS",
		SQMetrics: map[string]StatMetric{
			"Inexistent": new(StatASR),
		},
	}

	sq := &StatQueue{}

	data := []byte(utils.ToJSON(sqData))

	expErr := "unsupported metric type <Inexistent>"
	if err := sq.UnmarshalJSON(data); err == nil || err.Error() != expErr {
		t.Errorf("Expected error <%v>, Received error <%v>", expErr, err)
	}

}

func TestStatQueueUnmarshalJSONValUnmarshalErr(t *testing.T) {

	sqData := &StatQueue{
		Tenant: "cgrates.org",
		ID:     "STS",
		SQMetrics: map[string]StatMetric{
			utils.MetaASR: statMetricMock("bad value error"),
		},
	}

	sq := &StatQueue{}

	data := []byte(utils.ToJSON(sqData))

	expErr := "json: cannot unmarshal string into Go value of type engine.StatASR"
	if err := sq.UnmarshalJSON(data); err == nil || err.Error() != expErr {
		t.Errorf("Expected error <%v>, Received error <%v>", expErr, err)
	}

}

func TestStatQueueCompressTTLTrue(t *testing.T) {

	ttl := time.Millisecond
	expiryTime1 := time.Date(2021, 1, 1, 23, 59, 59, 0, time.UTC)

	sq := &StatQueue{
		SQItems: []SQItem{
			{
				EventID:    "id1",
				ExpiryTime: nil,
			},
			{
				EventID:    "id2",
				ExpiryTime: &expiryTime1,
			},
		},
		SQMetrics: map[string]StatMetric{
			utils.MetaTCD: &StatTCD{
				Metric: &Metric{

					Events: map[string]*DecimalWithCompress{
						"id1": {},
						"id2": {},
					},
				},
			},
		},
		ttl: &ttl,
	}

	maxQL := uint64(1)

	exp := []SQItem{
		{
			EventID:    "id2",
			ExpiryTime: &expiryTime1,
		},
		{
			EventID:    "id1",
			ExpiryTime: nil,
		},
	}
	rcv := sq.Compress(maxQL)
	if rcv != true {
		t.Fatalf("\nexpected: <%+v>, \nreceived: <%+v>", true, rcv)
	}

	if !reflect.DeepEqual(exp, sq.SQItems) {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", exp, sq.SQItems)
	}
}

func TestStatQAddStatEventFilterPassErr(t *testing.T) {

	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg.CacheCfg(), cM)

	fltrS := NewFilterS(cfg, cM, dm)

	sq = &StatQueue{
		SQMetrics: map[string]StatMetric{
			utils.MetaASR: &StatASR{
				Metric: &Metric{
					Value: utils.NewDecimal(1, 0),
					Count: 1,
					Events: map[string]*DecimalWithCompress{
						"cgrates.org:TestStatRemExpired_1": {Stat: utils.NewDecimalFromFloat64(1), CompressFactor: 1},
					},
				},
			},
		},
		sqPrfl: &StatQueueProfile{
			Metrics: []*MetricWithFilters{
				{
					FilterIDs: []string{"*"},
					MetricID:  utils.MetaASR,
				},
			},
		},
	}
	asrMetric := sq.SQMetrics[utils.MetaASR].(*StatASR)
	if asr := asrMetric.GetValue(); asr.Compare(utils.NewDecimalFromFloat64(100)) != 0 {
		t.Errorf("received ASR: %v", asr)
	}
	ev1 := &utils.CGREvent{Tenant: "cgrates.org", ID: "TestStatAddStatEvent_1"}

	expErr := `inline parse error for string: <*>`
	if err := sq.addStatEvent(context.Background(), ev1.Tenant, ev1.ID, fltrS, utils.MapStorage{utils.MetaOpts: ev1.Event}); err == nil || err.Error() != expErr {
		t.Errorf("Expected error %s received: %v", expErr, err)
	}

}

func TestStatQAddStatEventBlockerFromDynamicsErr(t *testing.T) {

	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg.CacheCfg(), cM)

	fltrS := NewFilterS(cfg, cM, dm)

	sq = &StatQueue{
		SQMetrics: map[string]StatMetric{
			utils.MetaASR: &StatASR{
				Metric: &Metric{
					Value: utils.NewDecimal(1, 0),
					Count: 1,
					Events: map[string]*DecimalWithCompress{
						"cgrates.org:TestStatRemExpired_1": {Stat: utils.NewDecimalFromFloat64(1), CompressFactor: 1},
					},
				},
			},
		},
		sqPrfl: &StatQueueProfile{
			Metrics: []*MetricWithFilters{
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
	asrMetric := sq.SQMetrics[utils.MetaASR].(*StatASR)
	if asr := asrMetric.GetValue(); asr.Compare(utils.NewDecimalFromFloat64(100)) != 0 {
		t.Errorf("received ASR: %v", asr)
	}
	ev1 := &utils.CGREvent{Tenant: "cgrates.org", ID: "TestStatAddStatEvent_1"}

	expErr := `NOT_IMPLEMENTED:*stirng`
	if err := sq.addStatEvent(context.Background(), ev1.Tenant, ev1.ID, fltrS, utils.MapStorage{utils.MetaOpts: ev1.Event}); err == nil || err.Error() != expErr {
		t.Errorf("Expected error %s received: %v", expErr, err)
	}

}

func TestStatQAddStatEventBlockNotLast(t *testing.T) {

	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg.CacheCfg(), cM)

	fltrS := NewFilterS(cfg, cM, dm)

	sq = &StatQueue{
		SQMetrics: map[string]StatMetric{
			utils.MetaASR: &StatASR{
				Metric: &Metric{
					Value: utils.NewDecimal(1, 0),
					Count: 1,
					Events: map[string]*DecimalWithCompress{
						"cgrates.org:TestStatRemExpired_1": {Stat: utils.NewDecimalFromFloat64(1), CompressFactor: 1},
					},
				},
			},
		},
		sqPrfl: &StatQueueProfile{
			Metrics: []*MetricWithFilters{
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
	asrMetric := sq.SQMetrics[utils.MetaASR].(*StatASR)
	if asr := asrMetric.GetValue(); asr.Compare(utils.NewDecimalFromFloat64(100)) != 0 {
		t.Errorf("received ASR: %v", asr)
	}
	ev1 := &utils.CGREvent{Tenant: "cgrates.org", ID: "TestStatAddStatEvent_1"}

	exp := &StatQueue{
		SQMetrics: map[string]StatMetric{
			utils.MetaASR: &StatASR{
				Metric: &Metric{
					Value: utils.NewDecimal(1, 0),
					Count: 1,
					Events: map[string]*DecimalWithCompress{
						"cgrates.org:TestStatRemExpired_1": {Stat: utils.NewDecimalFromFloat64(1), CompressFactor: 1},
					},
				},
			},
		},
		SQItems: []SQItem{
			{
				EventID: "eventID",
			},
		},
		sqPrfl: &StatQueueProfile{
			Metrics: []*MetricWithFilters{
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

	if err := sq.addStatEvent(context.Background(), ev1.Tenant, ev1.ID, fltrS, utils.MapStorage{utils.MetaOpts: ev1.Event}); err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(exp.sqPrfl, sq.sqPrfl) {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", exp.sqPrfl, sq.sqPrfl)
	}

}
