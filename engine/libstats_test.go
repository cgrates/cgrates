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
