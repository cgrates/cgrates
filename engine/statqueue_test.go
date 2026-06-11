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

	"github.com/cgrates/cgrates/utils"
)

func TestStatCompress(t *testing.T) {
	asr := &utils.StatASR{
		Metric: &utils.Metric{
			Value: utils.NewDecimal(2, 0),
			Count: 4,
			Events: map[string]*utils.DecimalWithCompress{
				"cgrates.org:TestStatRemExpired_1": {Stat: utils.NewDecimalFromFloat64(1), CompressFactor: 1},
				"cgrates.org:TestStatRemExpired_2": {Stat: utils.NewDecimalFromFloat64(0), CompressFactor: 1},
				"cgrates.org:TestStatRemExpired_3": {Stat: utils.NewDecimalFromFloat64(1), CompressFactor: 1},
				"cgrates.org:TestStatRemExpired_4": {Stat: utils.NewDecimalFromFloat64(0), CompressFactor: 1},
			},
		},
	}
	expectedASR := &utils.StatASR{
		Metric: &utils.Metric{
			Value: utils.NewDecimal(2, 0),
			Count: 4,
			Events: map[string]*utils.DecimalWithCompress{
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
		SQMetrics: map[string]utils.StatMetric{
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
	if rply := sq.SQMetrics[utils.MetaASR].(*utils.StatASR); !reflect.DeepEqual(*rply, *expectedASR) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expectedASR), utils.ToJSON(rply))
	}
}

func TestStatCompress2(t *testing.T) {
	asr := &utils.StatASR{
		Metric: &utils.Metric{
			Value: utils.NewDecimal(2, 0),
			Count: 4,
			Events: map[string]*utils.DecimalWithCompress{
				"cgrates.org:TestStatRemExpired_1": {Stat: utils.NewDecimalFromFloat64(1), CompressFactor: 1},
				"cgrates.org:TestStatRemExpired_2": {Stat: utils.NewDecimalFromFloat64(0), CompressFactor: 1},
				"cgrates.org:TestStatRemExpired_3": {Stat: utils.NewDecimalFromFloat64(1), CompressFactor: 1},
				"cgrates.org:TestStatRemExpired_4": {Stat: utils.NewDecimalFromFloat64(0), CompressFactor: 1},
			},
		},
	}
	expectedASR := &utils.StatASR{
		Metric: &utils.Metric{
			Value: utils.NewDecimal(2, 0),
			Count: 4,
			Events: map[string]*utils.DecimalWithCompress{
				"cgrates.org:TestStatRemExpired_4": {Stat: utils.NewDecimalFromFloat64(0.5), CompressFactor: 4},
			},
		},
	}
	tcd := &utils.StatTCD{
		Metric: &utils.Metric{
			Value: utils.NewDecimal(int64(3*time.Minute), 0),
			Count: 2,
			Events: map[string]*utils.DecimalWithCompress{
				"cgrates.org:TestStatRemExpired_2": {Stat: utils.NewDecimal(int64(time.Minute), 0), CompressFactor: 1},
				"cgrates.org:TestStatRemExpired_3": {Stat: utils.NewDecimal(int64(2*time.Minute), 0), CompressFactor: 1},
			},
		},
	}
	expectedTCD := &utils.StatTCD{
		Metric: &utils.Metric{
			Value: utils.NewDecimal(int64(3*time.Minute), 0),
			Count: 2,
			Events: map[string]*utils.DecimalWithCompress{
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
		SQMetrics: map[string]utils.StatMetric{
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
	if rply := sq.SQMetrics[utils.MetaASR].(*utils.StatASR); !reflect.DeepEqual(*rply, *expectedASR) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expectedASR), utils.ToJSON(rply))
	}
	if rply := sq.SQMetrics[utils.MetaTCD].(*utils.StatTCD); !rply.Equal(expectedTCD.Metric) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expectedTCD), utils.ToJSON(rply))
	}
}

func TestStatCompress3(t *testing.T) {
	tmNow := time.Now()
	asr := &utils.StatASR{
		Metric: &utils.Metric{
			Value: utils.NewDecimal(2, 0),
			Count: 4,
			Events: map[string]*utils.DecimalWithCompress{
				"cgrates.org:TestStatRemExpired_1": {Stat: utils.NewDecimalFromFloat64(1), CompressFactor: 1},
				"cgrates.org:TestStatRemExpired_2": {Stat: utils.NewDecimalFromFloat64(0), CompressFactor: 1},
				"cgrates.org:TestStatRemExpired_3": {Stat: utils.NewDecimalFromFloat64(1), CompressFactor: 1},
				"cgrates.org:TestStatRemExpired_4": {Stat: utils.NewDecimalFromFloat64(0), CompressFactor: 1},
			},
		},
	}
	expectedASR := &utils.StatASR{
		Metric: &utils.Metric{
			Value: utils.NewDecimal(2, 0),
			Count: 4,
			Events: map[string]*utils.DecimalWithCompress{
				"cgrates.org:TestStatRemExpired_4": {Stat: utils.NewDecimalFromFloat64(0.5), CompressFactor: 4},
			},
		},
	}
	tcd := &utils.StatTCD{
		Metric: &utils.Metric{
			Value: utils.NewDecimal(int64(3*time.Minute), 0),
			Count: 2,
			Events: map[string]*utils.DecimalWithCompress{
				"cgrates.org:TestStatRemExpired_2": {Stat: utils.NewDecimal(int64(time.Minute), 0), CompressFactor: 1},
				"cgrates.org:TestStatRemExpired_3": {Stat: utils.NewDecimal(int64(2*time.Minute), 0), CompressFactor: 1},
			},
		},
	}
	expectedTCD := &utils.StatTCD{
		Metric: &utils.Metric{
			Value: utils.NewDecimal(int64(3*time.Minute), 0),
			Count: 2,
			Events: map[string]*utils.DecimalWithCompress{
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
		SQMetrics: map[string]utils.StatMetric{
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
	if rply := sq.SQMetrics[utils.MetaASR].(*utils.StatASR); !reflect.DeepEqual(*rply, *expectedASR) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expectedASR), utils.ToJSON(rply))
	}
	if rply := sq.SQMetrics[utils.MetaTCD].(*utils.StatTCD); !reflect.DeepEqual(*rply, *expectedTCD) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(expectedTCD), utils.ToJSON(rply))
	}
}

func TestStatExpand(t *testing.T) {
	tmNow := time.Now()
	tests := []struct {
		name      string
		sq        *StatQueue
		wantItems []SQItem
		wantASR   *utils.StatASR
		wantTCD   *utils.StatTCD
	}{
		{
			name: "single metric",
			sq: &StatQueue{
				SQItems: []SQItem{
					{"cgrates.org:TestStatRemExpired_4", nil},
				},
				SQMetrics: map[string]utils.StatMetric{
					utils.MetaASR: &utils.StatASR{
						Metric: &utils.Metric{
							Value: utils.NewDecimal(2, 0),
							Count: 4,
							Events: map[string]*utils.DecimalWithCompress{
								"cgrates.org:TestStatRemExpired_4": {Stat: utils.NewDecimalFromFloat64(0.5), CompressFactor: 4},
							},
						},
					},
				},
			},
			wantItems: []SQItem{
				{"cgrates.org:TestStatRemExpired_4", nil},
				{"cgrates.org:TestStatRemExpired_4", nil},
				{"cgrates.org:TestStatRemExpired_4", nil},
				{"cgrates.org:TestStatRemExpired_4", nil},
			},
			wantASR: &utils.StatASR{
				Metric: &utils.Metric{
					Value: utils.NewDecimal(2, 0),
					Count: 4,
					Events: map[string]*utils.DecimalWithCompress{
						"cgrates.org:TestStatRemExpired_4": {Stat: utils.NewDecimalFromFloat64(0.5), CompressFactor: 4},
					},
				},
			},
		},
		{
			name: "two single-event metrics",
			sq: &StatQueue{
				SQItems: []SQItem{
					{"cgrates.org:TestStatRemExpired_3", utils.TimePointer(tmNow.Add(2 * time.Minute))},
					{"cgrates.org:TestStatRemExpired_4", nil},
				},
				ttl: utils.DurationPointer(time.Second),
				SQMetrics: map[string]utils.StatMetric{
					utils.MetaASR: &utils.StatASR{
						Metric: &utils.Metric{
							Value: utils.NewDecimal(2, 0),
							Count: 4,
							Events: map[string]*utils.DecimalWithCompress{
								"cgrates.org:TestStatRemExpired_4": {Stat: utils.NewDecimalFromFloat64(0.5), CompressFactor: 4},
							},
						},
					},
					utils.MetaTCD: &utils.StatTCD{
						Metric: &utils.Metric{
							Value: utils.NewDecimal(int64(3*time.Minute), 0),
							Count: 2,
							Events: map[string]*utils.DecimalWithCompress{
								"cgrates.org:TestStatRemExpired_4": {Stat: utils.NewDecimal(int64(time.Minute+30*time.Second), 0), CompressFactor: 2},
							},
						},
					},
				},
			},
			wantItems: []SQItem{
				{"cgrates.org:TestStatRemExpired_4", nil},
				{"cgrates.org:TestStatRemExpired_4", nil},
				{"cgrates.org:TestStatRemExpired_4", nil},
				{"cgrates.org:TestStatRemExpired_4", nil},
			},
			wantASR: &utils.StatASR{
				Metric: &utils.Metric{
					Value: utils.NewDecimal(2, 0),
					Count: 4,
					Events: map[string]*utils.DecimalWithCompress{
						"cgrates.org:TestStatRemExpired_4": {Stat: utils.NewDecimalFromFloat64(0.5), CompressFactor: 4},
					},
				},
			},
			wantTCD: &utils.StatTCD{
				Metric: &utils.Metric{
					Value: utils.NewDecimal(int64(3*time.Minute), 0),
					Count: 2,
					Events: map[string]*utils.DecimalWithCompress{
						"cgrates.org:TestStatRemExpired_4": {Stat: utils.NewDecimal(int64(time.Minute+30*time.Second), 0), CompressFactor: 2},
					},
				},
			},
		},
		{
			name: "two multi-event metrics",
			sq: &StatQueue{
				SQItems: []SQItem{
					{"cgrates.org:TestStatRemExpired_2", utils.TimePointer(tmNow.Add(time.Minute))},
					{"cgrates.org:TestStatRemExpired_3", utils.TimePointer(tmNow.Add(2 * time.Minute))},
					{"cgrates.org:TestStatRemExpired_4", nil},
				},
				ttl: utils.DurationPointer(time.Second),
				SQMetrics: map[string]utils.StatMetric{
					utils.MetaASR: &utils.StatASR{
						Metric: &utils.Metric{
							Value: utils.NewDecimal(2, 0),
							Count: 4,
							Events: map[string]*utils.DecimalWithCompress{
								"cgrates.org:TestStatRemExpired_3": {Stat: utils.NewDecimalFromFloat64(0.5), CompressFactor: 4},
							},
						},
					},
					utils.MetaTCD: &utils.StatTCD{
						Metric: &utils.Metric{
							Value: utils.NewDecimal(int64(3*time.Minute), 0),
							Count: 2,
							Events: map[string]*utils.DecimalWithCompress{
								"cgrates.org:TestStatRemExpired_2": {Stat: utils.NewDecimal(int64(time.Minute), 0), CompressFactor: 1},
								"cgrates.org:TestStatRemExpired_4": {Stat: utils.NewDecimal(int64(2*time.Minute), 0), CompressFactor: 1},
							},
						},
					},
				},
			},
			wantItems: []SQItem{
				{"cgrates.org:TestStatRemExpired_2", utils.TimePointer(tmNow.Add(time.Minute))},
				{"cgrates.org:TestStatRemExpired_3", utils.TimePointer(tmNow.Add(2 * time.Minute))},
				{"cgrates.org:TestStatRemExpired_3", utils.TimePointer(tmNow.Add(2 * time.Minute))},
				{"cgrates.org:TestStatRemExpired_3", utils.TimePointer(tmNow.Add(2 * time.Minute))},
				{"cgrates.org:TestStatRemExpired_3", utils.TimePointer(tmNow.Add(2 * time.Minute))},
				{"cgrates.org:TestStatRemExpired_4", nil},
			},
			wantASR: &utils.StatASR{
				Metric: &utils.Metric{
					Value: utils.NewDecimal(2, 0),
					Count: 4,
					Events: map[string]*utils.DecimalWithCompress{
						"cgrates.org:TestStatRemExpired_3": {Stat: utils.NewDecimalFromFloat64(0.5), CompressFactor: 4},
					},
				},
			},
			wantTCD: &utils.StatTCD{
				Metric: &utils.Metric{
					Value: utils.NewDecimal(int64(3*time.Minute), 0),
					Count: 2,
					Events: map[string]*utils.DecimalWithCompress{
						"cgrates.org:TestStatRemExpired_2": {Stat: utils.NewDecimal(int64(time.Minute), 0), CompressFactor: 1},
						"cgrates.org:TestStatRemExpired_4": {Stat: utils.NewDecimal(int64(2*time.Minute), 0), CompressFactor: 1},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.sq.Expand()
			if !reflect.DeepEqual(tt.sq.SQItems, tt.wantItems) {
				t.Errorf("Expected: %s , received: %s", utils.ToJSON(tt.wantItems), utils.ToJSON(tt.sq.SQItems))
			}
			if rply := tt.sq.SQMetrics[utils.MetaASR].(*utils.StatASR); !reflect.DeepEqual(*rply, *tt.wantASR) {
				t.Errorf("Expected: %s , received: %s", utils.ToJSON(tt.wantASR), utils.ToJSON(rply))
			}
			if tt.wantTCD != nil {
				if rply := tt.sq.SQMetrics[utils.MetaTCD].(*utils.StatTCD); !reflect.DeepEqual(*rply, *tt.wantTCD) {
					t.Errorf("Expected: %s , received: %s", utils.ToJSON(tt.wantTCD), utils.ToJSON(rply))
				}
			}
		})
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
		SQMetrics: map[string]utils.StatMetric{
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
	sm, err := utils.NewStatMetric(utils.MetaTCD, 0, []string{"*string:~*req.Account:1001"})
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
		SQMetrics: map[string]utils.StatMetric{
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
	if err := sq.Set([]string{}, "", false); err != utils.ErrWrongPath {
		t.Error(err)
	}
	if err := sq.Set([]string{""}, "", false); err != nil {
		t.Error(err)
	}
	if err := sq.Set([]string{"NotAField"}, ";", false); err != utils.ErrWrongPath {
		t.Error(err)
	}
	if err := sq.Set([]string{"NotAField", "1"}, ";", false); err != utils.ErrWrongPath {
		t.Error(err)
	}

	if err := sq.Set([]string{utils.Tenant}, "cgrates.org", false); err != nil {
		t.Error(err)
	}
	if err := sq.Set([]string{utils.ID}, "ID", false); err != nil {
		t.Error(err)
	}
	if err := sq.Set([]string{utils.FilterIDs}, "fltr1;*string:~*req.Account:1001", false); err != nil {
		t.Error(err)
	}
	if err := sq.Set([]string{utils.Weights}, ";10", false); err != nil {
		t.Error(err)
	}
	if err := sq.Set([]string{utils.QueueLength}, 10, false); err != nil {
		t.Error(err)
	}
	if err := sq.Set([]string{utils.TTL}, 10, false); err != nil {
		t.Error(err)
	}
	if err := sq.Set([]string{utils.MinItems}, 10, false); err != nil {
		t.Error(err)
	}
	if err := sq.Set([]string{utils.Stored}, true, false); err != nil {
		t.Error(err)
	}
	if err := sq.Set([]string{utils.Blockers}, ";true", false); err != nil {
		t.Error(err)
	}
	if err := sq.Set([]string{utils.ThresholdIDs}, "TH1", false); err != nil {
		t.Error(err)
	}

	if err := sq.Set([]string{utils.Metrics, utils.MetricID}, "*tcd;*acd", false); err != nil {
		t.Error(err)
	}
	if err := sq.Set([]string{utils.Metrics, utils.FilterIDs}, "fltr1", false); err != nil {
		t.Error(err)
	}

	if err := sq.Set([]string{utils.Metrics, "wrong"}, "fltr1", false); err != utils.ErrWrongPath {
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
			if err := sqp.Set(tt.args.path, tt.args.val, tt.args.newBranch); err != nil != tt.wantErr {
				t.Errorf("StatQueueProfile.Set() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestStatQueueGobEncode(t *testing.T) {
	expTime := time.Date(2021, 1, 1, 23, 59, 59, 0, time.UTC)
	sq := &StatQueue{
		Tenant: "cgrates.org",
		ID:     "SQ1",
		SQItems: []SQItem{
			{EventID: "ev1", ExpiryTime: &expTime},
			{EventID: "ev2"},
		},
		SQMetrics: map[string]utils.StatMetric{
			utils.MetaASR: &utils.StatASR{
				Metric: &utils.Metric{
					Value: utils.NewDecimal(1, 0),
					Count: 2,
					Events: map[string]*utils.DecimalWithCompress{
						"ev1": {Stat: utils.NewDecimalFromFloat64(1), CompressFactor: 1},
						"ev2": {Stat: utils.NewDecimalFromFloat64(0), CompressFactor: 1},
					},
				},
			},
		},
	}
	b, err := sq.GobEncode()
	if err != nil {
		t.Fatal(err)
	}
	rcv := new(StatQueue)
	if err := rcv.GobDecode(b); err != nil {
		t.Fatal(err)
	}
	if utils.ToJSON(rcv) != utils.ToJSON(sq) {
		t.Errorf("Expected <%s>, \nReceived <%s>", utils.ToJSON(sq), utils.ToJSON(rcv))
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
		SQMetrics: map[string]utils.StatMetric{
			"key": statMetricMock("remExpired error"),
		},
		lkID:  "testLkId",
		dirty: utils.BoolPointer(false),
		ttl:   utils.DurationPointer(time.Duration(3)),
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
		SQMetrics: map[string]utils.StatMetric{
			"key": statMetricMock("remExpired error"),
		},
		lkID:  "testLkId",
		dirty: utils.BoolPointer(false),
		ttl:   utils.DurationPointer(time.Duration(3)),
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

func TestStatQueueUnmarshalJSONErr(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		wantErr string
	}{
		{
			name:    "invalid json",
			data:    []byte{212},
			wantErr: "invalid character 'Ô' looking for beginning of value",
		},
		{
			name: "unsupported metric",
			data: []byte(utils.ToJSON(&StatQueue{
				Tenant: "cgrates.org",
				ID:     "STS",
				SQMetrics: map[string]utils.StatMetric{
					"Inexistent": new(utils.StatASR),
				},
			})),
			wantErr: "unsupported metric type <Inexistent>",
		},
		{
			name: "metric value unmarshal error",
			data: []byte(utils.ToJSON(&StatQueue{
				Tenant: "cgrates.org",
				ID:     "STS",
				SQMetrics: map[string]utils.StatMetric{
					utils.MetaASR: statMetricMock("bad value error"),
				},
			})),
			wantErr: "json: cannot unmarshal string into Go value of type utils.StatASR",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sq := &StatQueue{}
			if err := sq.UnmarshalJSON(tt.data); err == nil || err.Error() != tt.wantErr {
				t.Errorf("Expected error <%v>, Received error <%v>", tt.wantErr, err)
			}
		})
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
	if err := sq.Set([]string{utils.Metrics, utils.Blockers}, "incorrect input", false); err == nil || err.Error() != expErr {
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

	if err := sq.Set([]string{utils.Metrics, utils.Blockers}, "*string:~*opts.*cost:0;false;*suffix:~*req.Destination:+4432&eq:~*opts.*usage:10s;false;*notstring:~*req.RequestType:*prepaid;true;;false", false); err != nil {
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
		SQMetrics: map[string]utils.StatMetric{
			utils.MetaASR:      new(utils.StatASR),
			utils.MetaACD:      new(utils.StatACD),
			utils.MetaTCD:      new(utils.StatTCD),
			utils.MetaACC:      new(utils.StatACC),
			utils.MetaTCC:      new(utils.StatTCC),
			utils.MetaPDD:      new(utils.StatPDD),
			utils.MetaDDC:      new(utils.StatDDC),
			utils.MetaSum:      new(utils.StatSum),
			utils.MetaAverage:  new(utils.StatAverage),
			utils.MetaDistinct: new(utils.StatDistinct),
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
		SQMetrics: map[string]utils.StatMetric{
			utils.MetaTCD: &utils.StatTCD{
				Metric: &utils.Metric{

					Events: map[string]*utils.DecimalWithCompress{
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

func TestMetricWithFiltersClone(t *testing.T) {
	tests := []struct {
		name     string
		original *MetricWithFilters
	}{
		{
			name: "Full clone",
			original: &MetricWithFilters{
				MetricID:  "metric_1",
				FilterIDs: []string{"f1", "f2"},
				Blockers: utils.DynamicBlockers{
					&utils.DynamicBlocker{FilterIDs: []string{"f1"}, Blocker: true},
				},
			},
		},
		{
			name:     "Nil clone",
			original: nil,
		},
		{
			name: "No filters or blockers",
			original: &MetricWithFilters{
				MetricID: "metric_2",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cloned := tt.original.Clone()

			if tt.original == nil {
				if cloned != nil {
					t.Errorf("Clone() = %v, want nil", cloned)
				}
				return
			}

			if cloned == tt.original {
				t.Errorf("Clone() returned the same reference")
			}

			if cloned.MetricID != tt.original.MetricID {
				t.Errorf("MetricID = %s, want %s", cloned.MetricID, tt.original.MetricID)
			}

			if !reflect.DeepEqual(cloned.FilterIDs, tt.original.FilterIDs) {
				t.Errorf("FilterIDs = %v, want %v", cloned.FilterIDs, tt.original.FilterIDs)
			}

			if !reflect.DeepEqual(cloned.Blockers, tt.original.Blockers) {
				t.Errorf("Blockers = %v, want %v", cloned.Blockers, tt.original.Blockers)
			}

			if len(tt.original.FilterIDs) > 0 && &cloned.FilterIDs[0] == &tt.original.FilterIDs[0] {
				t.Errorf("FilterIDs slice not deeply cloned")
			}
			if len(tt.original.Blockers) > 0 && cloned.Blockers[0] == tt.original.Blockers[0] {
				t.Errorf("Blockers slice not deeply cloned")
			}
		})
	}
}
