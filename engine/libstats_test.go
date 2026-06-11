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

func TestStatRemEventWithID(t *testing.T) {
	type step struct {
		rem        []string
		wantVal    *utils.Decimal
		wantEvents int
	}
	tests := []struct {
		name    string
		metric  *utils.StatASR
		initVal *utils.Decimal
		steps   []step
	}{
		{
			name: "compress factor 1",
			metric: &utils.StatASR{
				Metric: &utils.Metric{
					Value: utils.NewDecimal(1, 0),
					Count: 2,
					Events: map[string]*utils.DecimalWithCompress{
						"cgrates.org:TestRemEventWithID_1": {Stat: utils.NewDecimal(1, 0), CompressFactor: 1},
						"cgrates.org:TestRemEventWithID_2": {Stat: utils.NewDecimal(0, 0), CompressFactor: 1},
					},
				},
			},
			initVal: utils.NewDecimal(50, 0),
			steps: []step{
				{rem: []string{"cgrates.org:TestRemEventWithID_1"}, wantVal: utils.NewDecimal(0, 0), wantEvents: 1},
				{rem: []string{"cgrates.org:TestRemEventWithID_5"}, wantVal: utils.NewDecimal(0, 0), wantEvents: 1}, // non existent
				{rem: []string{"cgrates.org:TestRemEventWithID_2"}, wantVal: utils.DecimalNaN, wantEvents: 0},
				{rem: []string{"cgrates.org:TestRemEventWithID_2"}, wantVal: utils.DecimalNaN, wantEvents: 0},
			},
		},
		{
			name: "compress factor 2",
			metric: &utils.StatASR{
				Metric: &utils.Metric{
					Value: utils.NewDecimal(2, 0),
					Count: 4,
					Events: map[string]*utils.DecimalWithCompress{
						"cgrates.org:TestRemEventWithID_1": {Stat: utils.NewDecimal(1, 0), CompressFactor: 2},
						"cgrates.org:TestRemEventWithID_2": {Stat: utils.NewDecimal(0, 0), CompressFactor: 2},
					},
				},
			},
			initVal: utils.NewDecimal(50, 0),
			steps: []step{
				{rem: []string{"cgrates.org:TestRemEventWithID_1", "cgrates.org:TestRemEventWithID_2"}, wantVal: utils.NewDecimal(50, 0), wantEvents: 2},
				{rem: []string{"cgrates.org:TestRemEventWithID_5"}, wantVal: utils.NewDecimal(50, 0), wantEvents: 2}, // non existent
				{rem: []string{"cgrates.org:TestRemEventWithID_2", "cgrates.org:TestRemEventWithID_1"}, wantVal: utils.DecimalNaN, wantEvents: 0},
				{rem: []string{"cgrates.org:TestRemEventWithID_2"}, wantVal: utils.DecimalNaN, wantEvents: 0},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sq := &StatQueue{SQMetrics: map[string]utils.StatMetric{utils.MetaASR: tt.metric}}
			m := tt.metric
			if got := m.GetValue(); got.Compare(tt.initVal) != 0 {
				t.Errorf("initial GetValue: %v", got)
			}
			for i, s := range tt.steps {
				for _, id := range s.rem {
					sq.remEventWithID(id)
				}
				if got := m.GetValue(); got.Compare(s.wantVal) != 0 {
					t.Errorf("step %d GetValue: %v", i, got)
				} else if len(m.Events) != s.wantEvents {
					t.Errorf("step %d Events: %+v", i, m.Events)
				}
			}
		})
	}
}

func TestStatRemExpired(t *testing.T) {
	sq = &StatQueue{
		SQMetrics: map[string]utils.StatMetric{
			utils.MetaASR: &utils.StatASR{
				Metric: &utils.Metric{
					Value: utils.NewDecimal(2, 0),
					Count: 3,
					Events: map[string]*utils.DecimalWithCompress{
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
	asrMetric := sq.SQMetrics[utils.MetaASR].(*utils.StatASR)
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
		SQMetrics: map[string]utils.StatMetric{
			utils.MetaASR: &utils.StatASR{
				Metric: &utils.Metric{
					Value: utils.NewDecimal(1, 0),
					Count: 1,
					Events: map[string]*utils.DecimalWithCompress{
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
	asrMetric := sq.SQMetrics[utils.MetaASR].(*utils.StatASR)
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
		SQMetrics: map[string]utils.StatMetric{
			utils.MetaTCD: &utils.StatTCD{
				Metric: &utils.Metric{
					FilterIDs: []string{"*string:~*req.Account:1002"},
					Value:     utils.NewDecimal(0, 0),
					Events: map[string]*utils.DecimalWithCompress{
						"cgrates.org:TestStatRemExpired_2": {Stat: utils.NewDecimalFromFloat64(float64(time.Minute)), CompressFactor: 1},
					},
				},
			},
			utils.MetaASR: &utils.StatASR{
				Metric: &utils.Metric{
					FilterIDs: []string{"*string:~*req.Account:1001"},
					Value:     utils.NewDecimal(0, 0),
					Events: map[string]*utils.DecimalWithCompress{
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

func TestStatRemoveExpiredTTL(t *testing.T) {
	sq = &StatQueue{
		ttl: utils.DurationPointer(100 * time.Millisecond),
		SQMetrics: map[string]utils.StatMetric{
			utils.MetaASR: &utils.StatASR{
				Metric: &utils.Metric{
					Value: utils.NewDecimal(1, 0),
					Count: 1,
					Events: map[string]*utils.DecimalWithCompress{
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
		SQMetrics: map[string]utils.StatMetric{
			utils.MetaASR: &utils.StatASR{
				Metric: &utils.Metric{
					Value: utils.NewDecimal(1, 0),
					Count: 1,
					Events: map[string]*utils.DecimalWithCompress{
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
func (sMM statMetricMock) Clone() utils.StatMetric {
	return sMM
}

type mockMarshal string

func (m mockMarshal) Marshal(v any) ([]byte, error)      { return nil, errors.New(string(m)) }
func (m mockMarshal) Unmarshal(data []byte, v any) error { return errors.New(string(m)) }
func TestStatQueueNewStoredStatQueue(t *testing.T) {
	sq := &StatQueue{
		SQMetrics: map[string]utils.StatMetric{
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
		SQMetrics: map[string]utils.StatMetric{},
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

	sm, err := utils.NewStatMetric(utils.MetaTCD, 0, []string{})
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
		SQMetrics: map[string]utils.StatMetric{
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
		SQMetrics: map[string]utils.StatMetric{
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
		SQMetrics: map[string]utils.StatMetric{
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
		SQMetrics: map[string]utils.StatMetric{
			utils.MetaTCD: &utils.StatTCD{Metric: &utils.Metric{}},
		},
	}

	experr := utils.ErrWrongPath
	err := sq.ProcessEvent(context.Background(), tnt, evID, filters, evNm)

	if err == nil || err != experr {
		t.Errorf("\nexpected: %q, \nreceived: %q", experr, err)
	}
}

func TestStatQueueaddStatEventNoPass(t *testing.T) {
	sm, err := utils.NewStatMetric(utils.MetaTCD, 0, []string{"*string:~*req.Account:1001"})
	if err != nil {
		t.Fatal(err)
	}

	sq := &StatQueue{
		SQMetrics: map[string]utils.StatMetric{
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
	idb, err := NewInternalDB(nil, nil, nil, config.CgrConfig().DbCfg().Items)
	if err != nil {
		t.Fatal(err)
	}
	filters := &FilterS{
		cfg: config.CgrConfig(),
		dm: &DataManager{
			dbConns: &DBConnManager{dbs: map[string]DataDB{
				utils.MetaDefault: idb,
			}},
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
		SQMetrics: map[string]utils.StatMetric{
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

func TestStatQAddStatEventFilterPassErr(t *testing.T) {

	cfg := config.NewDefaultCGRConfig()
	data, _ := NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	cM := NewConnManager(cfg)
	dbCM := NewDBConnManager(map[string]DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := NewDataManager(dbCM, cfg, cM)

	fltrS := NewFilterS(cfg, cM, dm)

	sq = &StatQueue{
		SQMetrics: map[string]utils.StatMetric{
			utils.MetaASR: &utils.StatASR{
				Metric: &utils.Metric{
					Value: utils.NewDecimal(1, 0),
					Count: 1,
					Events: map[string]*utils.DecimalWithCompress{
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
	asrMetric := sq.SQMetrics[utils.MetaASR].(*utils.StatASR)
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
	data, _ := NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	cM := NewConnManager(cfg)
	dbCM := NewDBConnManager(map[string]DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := NewDataManager(dbCM, cfg, cM)

	fltrS := NewFilterS(cfg, cM, dm)

	sq = &StatQueue{
		SQMetrics: map[string]utils.StatMetric{
			utils.MetaASR: &utils.StatASR{
				Metric: &utils.Metric{
					Value: utils.NewDecimal(1, 0),
					Count: 1,
					Events: map[string]*utils.DecimalWithCompress{
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
	asrMetric := sq.SQMetrics[utils.MetaASR].(*utils.StatASR)
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
	data, _ := NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	cM := NewConnManager(cfg)
	dbCM := NewDBConnManager(map[string]DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := NewDataManager(dbCM, cfg, cM)

	fltrS := NewFilterS(cfg, cM, dm)

	sq = &StatQueue{
		SQMetrics: map[string]utils.StatMetric{
			utils.MetaASR: &utils.StatASR{
				Metric: &utils.Metric{
					Value: utils.NewDecimal(1, 0),
					Count: 1,
					Events: map[string]*utils.DecimalWithCompress{
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
	asrMetric := sq.SQMetrics[utils.MetaASR].(*utils.StatASR)
	if asr := asrMetric.GetValue(); asr.Compare(utils.NewDecimalFromFloat64(100)) != 0 {
		t.Errorf("received ASR: %v", asr)
	}
	ev1 := &utils.CGREvent{Tenant: "cgrates.org", ID: "TestStatAddStatEvent_1"}

	exp := &StatQueue{
		SQMetrics: map[string]utils.StatMetric{
			utils.MetaASR: &utils.StatASR{
				Metric: &utils.Metric{
					Value: utils.NewDecimal(1, 0),
					Count: 1,
					Events: map[string]*utils.DecimalWithCompress{
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
