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

	"github.com/cgrates/cgrates/utils"
)

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
	sq := &utils.StatQueue{
		SQMetrics: map[string]utils.StatMetric{
			"key": statMetricMock(""),
		},
	}
	experr := "marshal mock error"
	var ms utils.Marshaler = mockMarshal(experr)

	rcv, err := NewStoredStatQueue(sq, ms, 0)

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
		SQItems: []utils.SQItem{
			{
				EventID: "testEventID",
			},
		},
	}
	var ms utils.Marshaler

	exp := &utils.StatQueue{
		SQItems: []utils.SQItem{
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
		SQItems: []utils.SQItem{
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
		SQItems: []utils.SQItem{
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
		SQItems: []utils.SQItem{
			{
				EventID: "testEventID",
			},
		},
		SQMetrics: map[string][]byte{
			utils.MetaTCD: msm,
		},
		Compressed: true,
	}

	exp := &utils.StatQueue{
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

func TestStoredStatQueueAsMapStringInterface(t *testing.T) {
	tests := []struct {
		name string
		ssq  *StoredStatQueue
		want map[string]any
	}{
		{
			ssq: &StoredStatQueue{
				Tenant: "cgrates.org",
				ID:     "ssq02",
				SQItems: []utils.SQItem{
					{
						EventID: "testID",
					},
				},
				SQMetrics: map[string][]byte{
					utils.MetaTCD: []byte(""),
				},
				Compressed: true,
			},
			want: map[string]any{
				utils.Tenant: "cgrates.org",
				utils.ID:     "ssq02",
				utils.SQItems: []utils.SQItem{
					{
						EventID: "testID",
					},
				},
				utils.SQMetrics: map[string][]byte{
					utils.MetaTCD: []byte(""),
				},
				utils.Compressed: true,
			},
		},
		{
			name: "Nil case",
			ssq:  nil,
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.ssq.AsMapStringInterface()
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Expected %v, recieved %v", tt.want, got)
			}
		})
	}
}

func TestMapStringInterfaceToStoredStatQueue(t *testing.T) {
	date := time.Date(2026, 7, 17, 15, 0, 0, 0, time.UTC)
	tests := []struct {
		name    string
		m       map[string]any
		ssq     *StoredStatQueue
		wantErr string
	}{
		{
			m: map[string]any{
				utils.Tenant: "cgrates.org",
				utils.ID:     "ssq01",
				utils.SQItems: []utils.SQItem{
					{
						EventID: "testID",
					},
				},
				utils.SQMetrics: map[string][]byte{
					utils.MetaTCD: []byte(""),
				},
				utils.Compressed: true,
			},
			ssq: &StoredStatQueue{
				Tenant:     "cgrates.org",
				ID:         "ssq01",
				SQItems:    nil,
				SQMetrics:  nil,
				Compressed: true,
			},
		},
		{
			name: "SQItems as any",
			m: map[string]any{
				utils.Tenant: "cgrates.org",
				utils.ID:     "ssq02",
				utils.SQItems: []any{
					map[string]any{
						utils.EventID: "testID",
					},
				},
				utils.SQMetrics: map[string]any{
					utils.MetaTCD: []byte(""),
				},
				utils.Compressed: true,
			},
			ssq: &StoredStatQueue{
				Tenant: "cgrates.org",
				ID:     "ssq02",
				SQItems: []utils.SQItem{
					{
						EventID: "testID",
					},
				},
				SQMetrics:  map[string][]byte{},
				Compressed: true,
			},
		},
		{
			name: "ExpiryTime as time.Time",
			m: map[string]any{
				utils.SQItems: []any{
					map[string]any{
						utils.EventID:    "testID",
						utils.ExpiryTime: &date,
					},
				},
			},
			ssq: &StoredStatQueue{
				SQItems: []utils.SQItem{
					{
						EventID:    "testID",
						ExpiryTime: &date,
					},
				},
			},
		},
		{
			name: "ExpiryTime as string",
			m: map[string]any{
				utils.SQItems: []any{
					map[string]any{
						utils.EventID:    "testID",
						utils.ExpiryTime: "2026-07-17T15:00:00Z",
					},
				},
			},
			ssq: &StoredStatQueue{
				SQItems: []utils.SQItem{
					{
						EventID:    "testID",
						ExpiryTime: &date,
					},
				},
			},
		},
		{
			name: "MetaTCD as string",
			m: map[string]any{
				utils.SQMetrics: map[string]any{
					utils.MetaTCD: "",
				},
			},
			ssq: &StoredStatQueue{
				SQMetrics: map[string][]byte{
					utils.MetaTCD: []byte(""),
				},
			},
		},
		{
			name: "Error case: ExpiryTime",
			m: map[string]any{
				utils.SQItems: []any{
					map[string]any{
						utils.EventID:    "testID",
						utils.ExpiryTime: `"2026-07-17T15:00:00Z`,
					},
				},
			},
			ssq:     nil,
			wantErr: `parsing time "\"2026-07-17T15:00:00Z" as "2006-01-02T15:04:05Z07:00": cannot parse "\"2026-07-17T15:00:00Z" as "2006"`,
		},
		{
			name: "Error case: MetaTCD",
			m: map[string]any{
				utils.SQMetrics: map[string]any{
					utils.MetaTCD: "1h0m0s",
				},
				utils.Compressed: true,
			},
			ssq:     nil,
			wantErr: `failed to decode base64 string: illegal base64 data at input byte 4`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := MapStringInterfaceToStoredStatQueue(tt.m)
			if gotErr != nil && gotErr.Error() != tt.wantErr {
				t.Errorf("Expected %v, recieved %v", tt.wantErr, gotErr)
			}

			if !reflect.DeepEqual(utils.ToJSON(got), utils.ToJSON(tt.ssq)) {
				t.Errorf("Expected %+v, recieved %+v", utils.ToJSON(tt.ssq), utils.ToJSON(got))
			}
		})
	}
}
