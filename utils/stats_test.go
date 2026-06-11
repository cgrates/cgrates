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
package utils

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
	"time"
)

func TestStatCompress(t *testing.T) {
	asr := &StatASR{
		Metric: &Metric{
			Value: NewDecimal(2, 0),
			Count: 4,
			Events: map[string]*DecimalWithCompress{
				"cgrates.org:TestStatRemExpired_1": {Stat: NewDecimalFromFloat64(1), CompressFactor: 1},
				"cgrates.org:TestStatRemExpired_2": {Stat: NewDecimalFromFloat64(0), CompressFactor: 1},
				"cgrates.org:TestStatRemExpired_3": {Stat: NewDecimalFromFloat64(1), CompressFactor: 1},
				"cgrates.org:TestStatRemExpired_4": {Stat: NewDecimalFromFloat64(0), CompressFactor: 1},
			},
		},
	}
	expectedASR := &StatASR{
		Metric: &Metric{
			Value: NewDecimal(2, 0),
			Count: 4,
			Events: map[string]*DecimalWithCompress{
				"cgrates.org:TestStatRemExpired_4": {Stat: NewDecimalFromFloat64(0.5), CompressFactor: 4},
			},
		},
	}
	sqItems := []SQItem{
		{"cgrates.org:TestStatRemExpired_1", TimePointer(time.Now())},
		{"cgrates.org:TestStatRemExpired_2", TimePointer(time.Now().Add(time.Minute))},
		{"cgrates.org:TestStatRemExpired_3", TimePointer(time.Now().Add(2 * time.Minute))},
		{"cgrates.org:TestStatRemExpired_4", nil},
	}
	expectedSqItems := []SQItem{
		{"cgrates.org:TestStatRemExpired_4", nil},
	}
	sq := &StatQueue{
		SQItems: sqItems,
		SQMetrics: map[string]StatMetric{
			MetaASR: asr,
		},
	}
	if sq.Compress(100) {
		t.Errorf("StatQueue compressed: %s", ToJSON(sq))
	}
	if !sq.Compress(2) {
		t.Errorf("StatQueue not compressed: %s", ToJSON(sq))
	}
	if !reflect.DeepEqual(sq.SQItems, expectedSqItems) {
		t.Errorf("Expected: %s , received: %s", ToJSON(expectedSqItems), ToJSON(sq.SQItems))
	}
	if rply := sq.SQMetrics[MetaASR].(*StatASR); !reflect.DeepEqual(*rply, *expectedASR) {
		t.Errorf("Expected: %s , received: %s", ToJSON(expectedASR), ToJSON(rply))
	}
}

func TestStatCompress2(t *testing.T) {
	asr := &StatASR{
		Metric: &Metric{
			Value: NewDecimal(2, 0),
			Count: 4,
			Events: map[string]*DecimalWithCompress{
				"cgrates.org:TestStatRemExpired_1": {Stat: NewDecimalFromFloat64(1), CompressFactor: 1},
				"cgrates.org:TestStatRemExpired_2": {Stat: NewDecimalFromFloat64(0), CompressFactor: 1},
				"cgrates.org:TestStatRemExpired_3": {Stat: NewDecimalFromFloat64(1), CompressFactor: 1},
				"cgrates.org:TestStatRemExpired_4": {Stat: NewDecimalFromFloat64(0), CompressFactor: 1},
			},
		},
	}
	expectedASR := &StatASR{
		Metric: &Metric{
			Value: NewDecimal(2, 0),
			Count: 4,
			Events: map[string]*DecimalWithCompress{
				"cgrates.org:TestStatRemExpired_4": {Stat: NewDecimalFromFloat64(0.5), CompressFactor: 4},
			},
		},
	}
	tcd := &StatTCD{
		Metric: &Metric{
			Value: NewDecimal(int64(3*time.Minute), 0),
			Count: 2,
			Events: map[string]*DecimalWithCompress{
				"cgrates.org:TestStatRemExpired_2": {Stat: NewDecimal(int64(time.Minute), 0), CompressFactor: 1},
				"cgrates.org:TestStatRemExpired_3": {Stat: NewDecimal(int64(2*time.Minute), 0), CompressFactor: 1},
			},
		},
	}
	expectedTCD := &StatTCD{
		Metric: &Metric{
			Value: NewDecimal(int64(3*time.Minute), 0),
			Count: 2,
			Events: map[string]*DecimalWithCompress{
				"cgrates.org:TestStatRemExpired_4": {Stat: NewDecimalFromFloat64(float64(time.Minute + 30*time.Second)), CompressFactor: 2},
			},
		},
	}
	sqItems := []SQItem{
		{"cgrates.org:TestStatRemExpired_1", TimePointer(time.Now())},
		{"cgrates.org:TestStatRemExpired_2", TimePointer(time.Now().Add(time.Minute))},
		{"cgrates.org:TestStatRemExpired_3", TimePointer(time.Now().Add(2 * time.Minute))},
		{"cgrates.org:TestStatRemExpired_4", nil},
	}
	expectedSqItems := []SQItem{
		{"cgrates.org:TestStatRemExpired_4", nil},
	}
	sq := &StatQueue{
		SQItems: sqItems,
		SQMetrics: map[string]StatMetric{
			MetaASR: asr,
			MetaTCD: tcd,
		},
	}
	if sq.Compress(100) {
		t.Errorf("StatQueue compressed: %s", ToJSON(sq))
	}
	if !sq.Compress(2) {
		t.Errorf("StatQueue not compressed: %s", ToJSON(sq))
	}
	if !reflect.DeepEqual(sq.SQItems, expectedSqItems) {
		t.Errorf("Expected: %s , received: %s", ToJSON(expectedSqItems), ToJSON(sq.SQItems))
	}
	if rply := sq.SQMetrics[MetaASR].(*StatASR); !reflect.DeepEqual(*rply, *expectedASR) {
		t.Errorf("Expected: %s , received: %s", ToJSON(expectedASR), ToJSON(rply))
	}
	if rply := sq.SQMetrics[MetaTCD].(*StatTCD); !rply.Equal(expectedTCD.Metric) {
		t.Errorf("Expected: %s , received: %s", ToJSON(expectedTCD), ToJSON(rply))
	}
}

func TestStatCompress3(t *testing.T) {
	tmNow := time.Now()
	asr := &StatASR{
		Metric: &Metric{
			Value: NewDecimal(2, 0),
			Count: 4,
			Events: map[string]*DecimalWithCompress{
				"cgrates.org:TestStatRemExpired_1": {Stat: NewDecimalFromFloat64(1), CompressFactor: 1},
				"cgrates.org:TestStatRemExpired_2": {Stat: NewDecimalFromFloat64(0), CompressFactor: 1},
				"cgrates.org:TestStatRemExpired_3": {Stat: NewDecimalFromFloat64(1), CompressFactor: 1},
				"cgrates.org:TestStatRemExpired_4": {Stat: NewDecimalFromFloat64(0), CompressFactor: 1},
			},
		},
	}
	expectedASR := &StatASR{
		Metric: &Metric{
			Value: NewDecimal(2, 0),
			Count: 4,
			Events: map[string]*DecimalWithCompress{
				"cgrates.org:TestStatRemExpired_4": {Stat: NewDecimalFromFloat64(0.5), CompressFactor: 4},
			},
		},
	}
	tcd := &StatTCD{
		Metric: &Metric{
			Value: NewDecimal(int64(3*time.Minute), 0),
			Count: 2,
			Events: map[string]*DecimalWithCompress{
				"cgrates.org:TestStatRemExpired_2": {Stat: NewDecimal(int64(time.Minute), 0), CompressFactor: 1},
				"cgrates.org:TestStatRemExpired_3": {Stat: NewDecimal(int64(2*time.Minute), 0), CompressFactor: 1},
			},
		},
	}
	expectedTCD := &StatTCD{
		Metric: &Metric{
			Value: NewDecimal(int64(3*time.Minute), 0),
			Count: 2,
			Events: map[string]*DecimalWithCompress{
				"cgrates.org:TestStatRemExpired_2": {Stat: NewDecimal(int64(time.Minute), 0), CompressFactor: 1},
				"cgrates.org:TestStatRemExpired_3": {Stat: NewDecimal(int64(2*time.Minute), 0), CompressFactor: 1},
			},
		},
	}
	sqItems := []SQItem{
		{"cgrates.org:TestStatRemExpired_1", TimePointer(tmNow)},
		{"cgrates.org:TestStatRemExpired_2", TimePointer(tmNow.Add(time.Minute))},
		{"cgrates.org:TestStatRemExpired_3", TimePointer(tmNow.Add(2 * time.Minute))},
		{"cgrates.org:TestStatRemExpired_4", nil},
	}
	expectedSqItems := []SQItem{
		{"cgrates.org:TestStatRemExpired_2", TimePointer(tmNow.Add(time.Minute))},
		{"cgrates.org:TestStatRemExpired_3", TimePointer(tmNow.Add(2 * time.Minute))},
		{"cgrates.org:TestStatRemExpired_4", nil},
	}
	sq := &StatQueue{
		SQItems: sqItems,
		SQMetrics: map[string]StatMetric{
			MetaASR: asr,
			MetaTCD: tcd,
		},
	}
	if sq.Compress(100) {
		t.Errorf("StatQueue compressed: %s", ToJSON(sq))
	}
	if !sq.Compress(3) {
		t.Errorf("StatQueue not compressed: %s", ToJSON(sq))
	}
	if !reflect.DeepEqual(sq.SQItems, expectedSqItems) {
		t.Errorf("Expected: %s , received: %s", ToJSON(expectedSqItems), ToJSON(sq.SQItems))
	}
	if rply := sq.SQMetrics[MetaASR].(*StatASR); !reflect.DeepEqual(*rply, *expectedASR) {
		t.Errorf("Expected: %s , received: %s", ToJSON(expectedASR), ToJSON(rply))
	}
	if rply := sq.SQMetrics[MetaTCD].(*StatTCD); !reflect.DeepEqual(*rply, *expectedTCD) {
		t.Errorf("Expected: %s , received: %s", ToJSON(expectedTCD), ToJSON(rply))
	}
}

func TestStatExpand(t *testing.T) {
	tmNow := time.Now()
	tests := []struct {
		name      string
		sq        *StatQueue
		wantItems []SQItem
		wantASR   *StatASR
		wantTCD   *StatTCD
	}{
		{
			name: "single metric",
			sq: &StatQueue{
				SQItems: []SQItem{
					{"cgrates.org:TestStatRemExpired_4", nil},
				},
				SQMetrics: map[string]StatMetric{
					MetaASR: &StatASR{
						Metric: &Metric{
							Value: NewDecimal(2, 0),
							Count: 4,
							Events: map[string]*DecimalWithCompress{
								"cgrates.org:TestStatRemExpired_4": {Stat: NewDecimalFromFloat64(0.5), CompressFactor: 4},
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
			wantASR: &StatASR{
				Metric: &Metric{
					Value: NewDecimal(2, 0),
					Count: 4,
					Events: map[string]*DecimalWithCompress{
						"cgrates.org:TestStatRemExpired_4": {Stat: NewDecimalFromFloat64(0.5), CompressFactor: 4},
					},
				},
			},
		},
		{
			name: "two single-event metrics",
			sq: &StatQueue{
				SQItems: []SQItem{
					{"cgrates.org:TestStatRemExpired_3", TimePointer(tmNow.Add(2 * time.Minute))},
					{"cgrates.org:TestStatRemExpired_4", nil},
				},
				SQMetrics: map[string]StatMetric{
					MetaASR: &StatASR{
						Metric: &Metric{
							Value: NewDecimal(2, 0),
							Count: 4,
							Events: map[string]*DecimalWithCompress{
								"cgrates.org:TestStatRemExpired_4": {Stat: NewDecimalFromFloat64(0.5), CompressFactor: 4},
							},
						},
					},
					MetaTCD: &StatTCD{
						Metric: &Metric{
							Value: NewDecimal(int64(3*time.Minute), 0),
							Count: 2,
							Events: map[string]*DecimalWithCompress{
								"cgrates.org:TestStatRemExpired_4": {Stat: NewDecimal(int64(time.Minute+30*time.Second), 0), CompressFactor: 2},
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
			wantASR: &StatASR{
				Metric: &Metric{
					Value: NewDecimal(2, 0),
					Count: 4,
					Events: map[string]*DecimalWithCompress{
						"cgrates.org:TestStatRemExpired_4": {Stat: NewDecimalFromFloat64(0.5), CompressFactor: 4},
					},
				},
			},
			wantTCD: &StatTCD{
				Metric: &Metric{
					Value: NewDecimal(int64(3*time.Minute), 0),
					Count: 2,
					Events: map[string]*DecimalWithCompress{
						"cgrates.org:TestStatRemExpired_4": {Stat: NewDecimal(int64(time.Minute+30*time.Second), 0), CompressFactor: 2},
					},
				},
			},
		},
		{
			name: "two multi-event metrics",
			sq: &StatQueue{
				SQItems: []SQItem{
					{"cgrates.org:TestStatRemExpired_2", TimePointer(tmNow.Add(time.Minute))},
					{"cgrates.org:TestStatRemExpired_3", TimePointer(tmNow.Add(2 * time.Minute))},
					{"cgrates.org:TestStatRemExpired_4", nil},
				},
				SQMetrics: map[string]StatMetric{
					MetaASR: &StatASR{
						Metric: &Metric{
							Value: NewDecimal(2, 0),
							Count: 4,
							Events: map[string]*DecimalWithCompress{
								"cgrates.org:TestStatRemExpired_3": {Stat: NewDecimalFromFloat64(0.5), CompressFactor: 4},
							},
						},
					},
					MetaTCD: &StatTCD{
						Metric: &Metric{
							Value: NewDecimal(int64(3*time.Minute), 0),
							Count: 2,
							Events: map[string]*DecimalWithCompress{
								"cgrates.org:TestStatRemExpired_2": {Stat: NewDecimal(int64(time.Minute), 0), CompressFactor: 1},
								"cgrates.org:TestStatRemExpired_4": {Stat: NewDecimal(int64(2*time.Minute), 0), CompressFactor: 1},
							},
						},
					},
				},
			},
			wantItems: []SQItem{
				{"cgrates.org:TestStatRemExpired_2", TimePointer(tmNow.Add(time.Minute))},
				{"cgrates.org:TestStatRemExpired_3", TimePointer(tmNow.Add(2 * time.Minute))},
				{"cgrates.org:TestStatRemExpired_3", TimePointer(tmNow.Add(2 * time.Minute))},
				{"cgrates.org:TestStatRemExpired_3", TimePointer(tmNow.Add(2 * time.Minute))},
				{"cgrates.org:TestStatRemExpired_3", TimePointer(tmNow.Add(2 * time.Minute))},
				{"cgrates.org:TestStatRemExpired_4", nil},
			},
			wantASR: &StatASR{
				Metric: &Metric{
					Value: NewDecimal(2, 0),
					Count: 4,
					Events: map[string]*DecimalWithCompress{
						"cgrates.org:TestStatRemExpired_3": {Stat: NewDecimalFromFloat64(0.5), CompressFactor: 4},
					},
				},
			},
			wantTCD: &StatTCD{
				Metric: &Metric{
					Value: NewDecimal(int64(3*time.Minute), 0),
					Count: 2,
					Events: map[string]*DecimalWithCompress{
						"cgrates.org:TestStatRemExpired_2": {Stat: NewDecimal(int64(time.Minute), 0), CompressFactor: 1},
						"cgrates.org:TestStatRemExpired_4": {Stat: NewDecimal(int64(2*time.Minute), 0), CompressFactor: 1},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.sq.Expand()
			if !reflect.DeepEqual(tt.sq.SQItems, tt.wantItems) {
				t.Errorf("Expected: %s , received: %s", ToJSON(tt.wantItems), ToJSON(tt.sq.SQItems))
			}
			if rply := tt.sq.SQMetrics[MetaASR].(*StatASR); !reflect.DeepEqual(*rply, *tt.wantASR) {
				t.Errorf("Expected: %s , received: %s", ToJSON(tt.wantASR), ToJSON(rply))
			}
			if tt.wantTCD != nil {
				if rply := tt.sq.SQMetrics[MetaTCD].(*StatTCD); !reflect.DeepEqual(*rply, *tt.wantTCD) {
					t.Errorf("Expected: %s , received: %s", ToJSON(tt.wantTCD), ToJSON(rply))
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
	sm, err := NewStatMetric(MetaTCD, 0, []string{"*string:~*req.Account:1001"})
	if err != nil {
		t.Fatal(err)
	}

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
			MetaTCD: statMetricMock("populate idMap"),
			MetaReq: sm,
		},
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
		{MetricID: MetaASR},
		{MetricID: MetaTCD},
	}, 1)
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal([]byte(ToJSON(exp)), &rply); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(rply, exp) {
		t.Errorf("Expected: %s , received: %s", ToJSON(exp), ToJSON(rply))
	}

}

func TestStatQueueWithAPIOptsJSONMarshall(t *testing.T) {
	rply := &StatQueueWithAPIOpts{}
	exp, err := NewStatQueue("cgrates.org", "STS", []*MetricWithFilters{
		{MetricID: MetaASR},
		{MetricID: MetaTCD},
	}, 1)
	exp2 := &StatQueueWithAPIOpts{
		StatQueue: exp,
		APIOpts:   map[string]any{"a": "a"},
	}
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal([]byte(ToJSON(exp2)), rply); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(rply, exp2) {
		t.Errorf("Expected: %+v , received: %+v", exp2, rply)
		t.Errorf("Expected: %s , received: %s", ToJSON(exp2), ToJSON(rply))
	}

}

func TestStatQueueProfileSet(t *testing.T) {
	sq := StatQueueProfile{}
	exp := StatQueueProfile{
		Tenant:    "cgrates.org",
		ID:        "ID",
		FilterIDs: []string{"fltr1", "*string:~*req.Account:1001"},
		Weights: DynamicWeights{
			{
				Weight: 10,
			},
		},
		QueueLength:  10,
		TTL:          10,
		MinItems:     10,
		Stored:       true,
		Blockers:     DynamicBlockers{{Blocker: true}},
		ThresholdIDs: []string{"TH1"},
		Metrics: []*MetricWithFilters{{
			MetricID: MetaTCD,
		}, {
			MetricID:  MetaACD,
			FilterIDs: []string{"fltr1"},
		}},
	}
	if err := sq.Set([]string{}, "", false); err != ErrWrongPath {
		t.Error(err)
	}
	if err := sq.Set([]string{""}, "", false); err != nil {
		t.Error(err)
	}
	if err := sq.Set([]string{"NotAField"}, ";", false); err != ErrWrongPath {
		t.Error(err)
	}
	if err := sq.Set([]string{"NotAField", "1"}, ";", false); err != ErrWrongPath {
		t.Error(err)
	}

	if err := sq.Set([]string{Tenant}, "cgrates.org", false); err != nil {
		t.Error(err)
	}
	if err := sq.Set([]string{ID}, "ID", false); err != nil {
		t.Error(err)
	}
	if err := sq.Set([]string{FilterIDs}, "fltr1;*string:~*req.Account:1001", false); err != nil {
		t.Error(err)
	}
	if err := sq.Set([]string{Weights}, ";10", false); err != nil {
		t.Error(err)
	}
	if err := sq.Set([]string{QueueLength}, 10, false); err != nil {
		t.Error(err)
	}
	if err := sq.Set([]string{TTL}, 10, false); err != nil {
		t.Error(err)
	}
	if err := sq.Set([]string{MinItems}, 10, false); err != nil {
		t.Error(err)
	}
	if err := sq.Set([]string{Stored}, true, false); err != nil {
		t.Error(err)
	}
	if err := sq.Set([]string{Blockers}, ";true", false); err != nil {
		t.Error(err)
	}
	if err := sq.Set([]string{ThresholdIDs}, "TH1", false); err != nil {
		t.Error(err)
	}

	if err := sq.Set([]string{Metrics, MetricID}, "*tcd;*acd", false); err != nil {
		t.Error(err)
	}
	if err := sq.Set([]string{Metrics, FilterIDs}, "fltr1", false); err != nil {
		t.Error(err)
	}

	if err := sq.Set([]string{Metrics, "wrong"}, "fltr1", false); err != ErrWrongPath {
		t.Error(err)
	}
	if !reflect.DeepEqual(exp, sq) {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(sq))
	}
}

func TestStatQueueProfileAsInterface(t *testing.T) {
	sqp := StatQueueProfile{
		Tenant:    "cgrates.org",
		ID:        "ID",
		FilterIDs: []string{"fltr1", "*string:~*req.Account:1001"},
		Weights: DynamicWeights{
			{
				Weight: 10,
			},
		},
		QueueLength:  10,
		TTL:          10,
		MinItems:     10,
		Stored:       true,
		Blockers:     DynamicBlockers{{Blocker: true}},
		ThresholdIDs: []string{"TH1"},
		Metrics: []*MetricWithFilters{{
			MetricID: MetaTCD,
		}, {
			MetricID:  MetaACD,
			FilterIDs: []string{"fltr1"},
		}, {

			Blockers: DynamicBlockers{{Blocker: true}},
		}},
	}
	if _, err := sqp.FieldAsInterface(nil); err != ErrNotFound {
		t.Fatal(err)
	}
	if _, err := sqp.FieldAsInterface([]string{"field"}); err != ErrNotFound {
		t.Fatal(err)
	}
	if _, err := sqp.FieldAsInterface([]string{"field", ""}); err != ErrNotFound {
		t.Fatal(err)
	}
	if val, err := sqp.FieldAsInterface([]string{Tenant}); err != nil {
		t.Fatal(err)
	} else if exp := "cgrates.org"; exp != val {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, err := sqp.FieldAsInterface([]string{ID}); err != nil {
		t.Fatal(err)
	} else if exp := ID; exp != val {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, err := sqp.FieldAsInterface([]string{FilterIDs}); err != nil {
		t.Fatal(err)
	} else if exp := sqp.FilterIDs; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, err := sqp.FieldAsInterface([]string{FilterIDs + "[0]"}); err != nil {
		t.Fatal(err)
	} else if exp := sqp.FilterIDs[0]; exp != val {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, err := sqp.FieldAsInterface([]string{Weights}); err != nil {
		t.Fatal(err)
	} else if exp := sqp.Weights; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, err := sqp.FieldAsInterface([]string{ThresholdIDs}); err != nil {
		t.Fatal(err)
	} else if exp := sqp.ThresholdIDs; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, err := sqp.FieldAsInterface([]string{ThresholdIDs + "[0]"}); err != nil {
		t.Fatal(err)
	} else if exp := sqp.ThresholdIDs[0]; exp != val {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, err := sqp.FieldAsInterface([]string{Metrics}); err != nil {
		t.Fatal(err)
	} else if exp := sqp.Metrics; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, err := sqp.FieldAsInterface([]string{Metrics + "[0]"}); err != nil {
		t.Fatal(err)
	} else if exp := sqp.Metrics[0]; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}

	if val, err := sqp.FieldAsInterface([]string{QueueLength}); err != nil {
		t.Fatal(err)
	} else if exp := sqp.QueueLength; exp != val {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, err := sqp.FieldAsInterface([]string{TTL}); err != nil {
		t.Fatal(err)
	} else if exp := sqp.TTL; exp != val {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, err := sqp.FieldAsInterface([]string{MinItems}); err != nil {
		t.Fatal(err)
	} else if exp := sqp.MinItems; exp != val {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, err := sqp.FieldAsInterface([]string{Stored}); err != nil {
		t.Fatal(err)
	} else if exp := sqp.Stored; exp != val {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, err := sqp.FieldAsInterface([]string{Blockers}); err != nil {
		t.Fatal(err)
	} else if exp := sqp.Blockers; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if _, err := sqp.FieldAsInterface([]string{Metrics + "[4]"}); err != ErrNotFound {
		t.Fatal(err)
	}
	if _, err := sqp.FieldAsInterface([]string{Metrics + "4]"}); err != ErrNotFound {
		t.Fatal(err)
	}
	if _, err := sqp.FieldAsInterface([]string{Metrics + "[4]", ""}); err != ErrNotFound {
		t.Fatal(err)
	}
	if _, err := sqp.FieldAsInterface([]string{Metrics + "[0]", ""}); err != ErrNotFound {
		t.Fatal(err)
	}
	if _, err := sqp.FieldAsInterface([]string{Metrics + "[0]", "", ""}); err != ErrNotFound {
		t.Fatal(err)
	}

	if val, err := sqp.FieldAsInterface([]string{Metrics + "[0]", MetricID}); err != nil {
		t.Fatal(err)
	} else if exp := sqp.Metrics[0].MetricID; exp != val {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}

	if val, err := sqp.FieldAsInterface([]string{Metrics + "[0]", FilterIDs}); err != nil {
		t.Fatal(err)
	} else if exp := sqp.Metrics[0].FilterIDs; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, err := sqp.FieldAsInterface([]string{Metrics + "[1]", FilterIDs + "[0]"}); err != nil {
		t.Fatal(err)
	} else if exp := sqp.Metrics[1].FilterIDs[0]; exp != val {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}

	if _, err := sqp.FieldAsString([]string{""}); err != ErrNotFound {
		t.Fatal(err)
	}
	if val, err := sqp.FieldAsString([]string{ID}); err != nil {
		t.Fatal(err)
	} else if exp := "ID"; exp != val {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, exp := sqp.String(), ToJSON(sqp); exp != val {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}

	if _, err := sqp.Metrics[0].FieldAsString([]string{""}); err != ErrNotFound {
		t.Fatal(err)
	}
	if val, err := sqp.Metrics[0].FieldAsString([]string{MetricID}); err != nil {
		t.Fatal(err)
	} else if exp := MetaTCD; exp != val {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, exp := sqp.Metrics[0].String(), ToJSON(sqp.Metrics[0]); exp != val {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}

	if val, err := sqp.Metrics[2].FieldAsInterface([]string{Blockers}); err != nil {
		t.Fatal(err)
	} else if exp := sqp.Metrics[2].Blockers; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
}

func TestStatQueueProfileMerge(t *testing.T) {
	sqp := &StatQueueProfile{}
	exp := &StatQueueProfile{
		Tenant:    "cgrates.org",
		ID:        "ID",
		FilterIDs: []string{"fltr1", "*string:~*req.Account:1001"},
		Weights: DynamicWeights{
			{
				Weight: 10,
			},
		},
		QueueLength:  10,
		TTL:          10,
		MinItems:     10,
		Stored:       true,
		Blockers:     DynamicBlockers{{Blocker: true}},
		ThresholdIDs: []string{"TH1"},
		Metrics: []*MetricWithFilters{{
			MetricID: MetaTCD,
		}, {
			MetricID:  MetaACD,
			FilterIDs: []string{"fltr1"},
		}},
	}
	if sqp.Merge(&StatQueueProfile{
		Tenant:    "cgrates.org",
		ID:        "ID",
		FilterIDs: []string{"fltr1", "*string:~*req.Account:1001"},
		Weights: DynamicWeights{
			{
				Weight: 10,
			},
		},
		QueueLength:  10,
		TTL:          10,
		MinItems:     10,
		Stored:       true,
		Blockers:     DynamicBlockers{{Blocker: true}},
		ThresholdIDs: []string{"TH1"},
		Metrics: []*MetricWithFilters{{
			MetricID: MetaTCD,
		}, {
			MetricID:  MetaACD,
			FilterIDs: []string{"fltr1"},
		}},
	}); !reflect.DeepEqual(exp, sqp) {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(sqp))
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
		Weights      DynamicWeights
		ThresholdIDs []string
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
				Blockers:     DynamicBlockers{{Blocker: true}},
				Weights:      tt.fields.Weights,
				ThresholdIDs: tt.fields.ThresholdIDs,
			}
			if err := sqp.Set(tt.args.path, tt.args.val, tt.args.newBranch); err != nil != tt.wantErr {
				t.Errorf("StatQueueProfile.Set() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestStatQueueGobEncode(t *testing.T) {
	gob.Register(new(StatASR)) // done by engine/caches.go in production
	expTime := time.Date(2021, 1, 1, 23, 59, 59, 0, time.UTC)
	sq := &StatQueue{
		Tenant: "cgrates.org",
		ID:     "SQ1",
		SQItems: []SQItem{
			{EventID: "ev1", ExpiryTime: &expTime},
			{EventID: "ev2"},
		},
		SQMetrics: map[string]StatMetric{
			MetaASR: &StatASR{
				Metric: &Metric{
					Value: NewDecimal(1, 0),
					Count: 2,
					Events: map[string]*DecimalWithCompress{
						"ev1": {Stat: NewDecimalFromFloat64(1), CompressFactor: 1},
						"ev2": {Stat: NewDecimalFromFloat64(0), CompressFactor: 1},
					},
				},
			},
		},
	}
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(sq); err != nil {
		t.Fatal(err)
	}
	rcv := new(StatQueue)
	if err := gob.NewDecoder(&buf).Decode(rcv); err != nil {
		t.Fatal(err)
	}
	if ToJSON(rcv) != ToJSON(sq) {
		t.Errorf("Expected <%s>, \nReceived <%s>", ToJSON(sq), ToJSON(rcv))
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
	}
	if rcv := sq.Clone(); !reflect.DeepEqual(ToJSON(rcv), ToJSON(exp)) {
		t.Errorf("Expected <%v>, \nReceived <%v>", ToJSON(exp), ToJSON(rcv))
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
			data: []byte(ToJSON(&StatQueue{
				Tenant: "cgrates.org",
				ID:     "STS",
				SQMetrics: map[string]StatMetric{
					"Inexistent": new(StatASR),
				},
			})),
			wantErr: "unsupported metric type <Inexistent>",
		},
		{
			name: "metric value unmarshal error",
			data: []byte(ToJSON(&StatQueue{
				Tenant: "cgrates.org",
				ID:     "STS",
				SQMetrics: map[string]StatMetric{
					MetaASR: statMetricMock("bad value error"),
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
	if err := sq.Set([]string{Metrics, Blockers}, "incorrect input", false); err == nil || err.Error() != expErr {
		t.Errorf("Expected error <%v>, Received error <%v>", expErr, err)
	}

}

func TestStatQueueProfileSetBlockersOK(t *testing.T) {
	sq := StatQueueProfile{}

	exp := StatQueueProfile{
		Metrics: []*MetricWithFilters{
			{
				Blockers: DynamicBlockers{&DynamicBlocker{
					FilterIDs: []string{"*string:~*opts.*cost:0"},
					Blocker:   false,
				},
					&DynamicBlocker{FilterIDs: []string{"*suffix:~*req.Destination:+4432", "eq:~*opts.*usage:10s"},
						Blocker: false},
					&DynamicBlocker{FilterIDs: []string{"*notstring:~*req.RequestType:*prepaid"},
						Blocker: true},
					&DynamicBlocker{FilterIDs: nil,
						Blocker: false},
				},
			},
		},
	}

	if err := sq.Set([]string{Metrics, Blockers}, "*string:~*opts.*cost:0;false;*suffix:~*req.Destination:+4432&eq:~*opts.*usage:10s;false;*notstring:~*req.RequestType:*prepaid;true;;false", false); err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(exp, sq) {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(sq))
	}
}

func TestStatQueueUnmarshalJSONOK(t *testing.T) {

	sqData := &StatQueue{
		Tenant: "cgrates.org",
		ID:     "STS",
		SQMetrics: map[string]StatMetric{
			MetaASR:      new(StatASR),
			MetaACD:      new(StatACD),
			MetaTCD:      new(StatTCD),
			MetaACC:      new(StatACC),
			MetaTCC:      new(StatTCC),
			MetaPDD:      new(StatPDD),
			MetaDDC:      new(StatDDC),
			MetaSum:      new(StatSum),
			MetaAverage:  new(StatAverage),
			MetaDistinct: new(StatDistinct),
		},
	}

	sq := &StatQueue{}

	data := []byte(ToJSON(sqData))

	if err := sq.UnmarshalJSON(data); err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(ToJSON(sqData), ToJSON(sq)) {
		t.Errorf("Expected <%v>, \nReceived <%v>", ToJSON(sqData), ToJSON(sq))
	}

}

func TestStatQueueCompressTTLTrue(t *testing.T) {

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
			MetaTCD: &StatTCD{
				Metric: &Metric{

					Events: map[string]*DecimalWithCompress{
						"id1": {},
						"id2": {},
					},
				},
			},
		},
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
				Blockers: DynamicBlockers{
					&DynamicBlocker{FilterIDs: []string{"f1"}, Blocker: true},
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

type statMetricMock string

func (statMetricMock) GetValue() *Decimal {
	return nil
}

func (statMetricMock) GetStringValue(int) (val string) {
	return
}

func (statMetricMock) AddEvent(string, DataProvider) error {
	return nil
}

func (statMetricMock) AddOneEvent(DataProvider) error {
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
