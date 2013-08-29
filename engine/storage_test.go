/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2013 ITsysCOM

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITH*out ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>
*/

package engine

import (
	"testing"
	"time"
)

func TestMsgpackStructsAdded(t *testing.T) {
	var a = struct{ First string }{"test"}
	var b = struct {
		First  string
		Second string
	}{}
	m := NewCodecMsgpackMarshaler()
	buf, err := m.Marshal(&a)
	if err != nil {
		t.Error("error marshaling structure: ", err)
	}
	err = m.Unmarshal(buf, &b)
	if err != nil || b.First != "test" || b.Second != "" {
		t.Error("error unmarshalling structure: ", b, err)
	}
}

func TestMsgpackStructsMissing(t *testing.T) {
	var a = struct {
		First  string
		Second string
	}{"test1", "test2"}
	var b = struct{ First string }{}
	m := NewCodecMsgpackMarshaler()
	buf, err := m.Marshal(&a)
	if err != nil {
		t.Error("error marshaling structure: ", err)
	}
	err = m.Unmarshal(buf, &b)
	if err != nil || b.First != "test1" {
		t.Error("error unmarshalling structure: ", b, err)
	}
}

func TestMsgpackTime(t *testing.T) {
	t1 := time.Date(2013, 8, 28, 22, 27, 0, 0, time.UTC)
	m := NewCodecMsgpackMarshaler()
	buf, err := m.Marshal(&t1)
	if err != nil {
		t.Error("error marshaling structure: ", err)
	}
	var t2 time.Time
	err = m.Unmarshal(buf, &t2)
	if err != nil || t1 != t2 || !t1.Equal(t2) {
		t.Errorf("error unmarshalling structure: %#v %#v %v", t1, t2, err)
	}
}

/************************** Benchmarks *****************************/

func GetUB() *UserBalance {
	uc := &UnitsCounter{
		Direction:     OUTBOUND,
		BalanceId:     SMS,
		Units:         100,
		MinuteBuckets: []*MinuteBucket{&MinuteBucket{Weight: 20, Price: 1, DestinationId: "NAT"}, &MinuteBucket{Weight: 10, Price: 10, PriceType: ABSOLUTE, DestinationId: "RET"}},
	}
	at := &ActionTrigger{
		Id:             "some_uuid",
		BalanceId:      CREDIT,
		Direction:      OUTBOUND,
		ThresholdValue: 100.0,
		DestinationId:  "NAT",
		Weight:         10.0,
		ActionsId:      "Commando",
	}
	var zeroTime time.Time
	zeroTime = zeroTime.UTC() // for deep equal to find location
	ub := &UserBalance{
		Id:             "rif",
		Type:           UB_TYPE_POSTPAID,
		BalanceMap:     map[string]BalanceChain{SMS + OUTBOUND: BalanceChain{&Balance{Value: 14, ExpirationDate: zeroTime}}, TRAFFIC + OUTBOUND: BalanceChain{&Balance{Value: 1024, ExpirationDate: zeroTime}}},
		MinuteBuckets:  []*MinuteBucket{&MinuteBucket{Weight: 20, Price: 1, DestinationId: "NAT"}, &MinuteBucket{Weight: 10, Price: 10, PriceType: ABSOLUTE, DestinationId: "RET"}},
		UnitCounters:   []*UnitsCounter{uc, uc},
		ActionTriggers: ActionTriggerPriotityList{at, at, at},
	}
	return ub
}

func BenchmarkMarshallerJSONStoreRestore(b *testing.B) {
	b.StopTimer()
	d := time.Date(2012, time.February, 1, 14, 30, 1, 0, time.UTC)
	i := &Interval{Months: []time.Month{time.February},
		MonthDays: []int{1},
		WeekDays:  []time.Weekday{time.Wednesday, time.Thursday},
		StartTime: "14:30:00",
		EndTime:   "15:00:00"}
	ap := &ActivationPeriod{ActivationTime: d}
	ap.AddInterval(i)
	ub := GetUB()

	ap1 := ActivationPeriod{}
	ub1 := &UserBalance{}
	b.StartTimer()
	ms := new(JSONMarshaler)
	for i := 0; i < b.N; i++ {
		result, _ := ms.Marshal(ap)
		ms.Unmarshal(result, ap1)
		result, _ = ms.Marshal(ub)
		ms.Unmarshal(result, ub1)
	}
}

func BenchmarkMarshallerBSONStoreRestore(b *testing.B) {
	b.StopTimer()
	d := time.Date(2012, time.February, 1, 14, 30, 1, 0, time.UTC)
	i := &Interval{Months: []time.Month{time.February},
		MonthDays: []int{1},
		WeekDays:  []time.Weekday{time.Wednesday, time.Thursday},
		StartTime: "14:30:00",
		EndTime:   "15:00:00"}
	ap := &ActivationPeriod{ActivationTime: d}
	ap.AddInterval(i)
	ub := GetUB()

	ap1 := ActivationPeriod{}
	ub1 := &UserBalance{}
	b.StartTimer()
	ms := new(BSONMarshaler)
	for i := 0; i < b.N; i++ {
		result, _ := ms.Marshal(ap)
		ms.Unmarshal(result, ap1)
		result, _ = ms.Marshal(ub)
		ms.Unmarshal(result, ub1)
	}
}

func BenchmarkMarshallerJSONBufStoreRestore(b *testing.B) {
	b.StopTimer()
	d := time.Date(2012, time.February, 1, 14, 30, 1, 0, time.UTC)
	i := &Interval{Months: []time.Month{time.February},
		MonthDays: []int{1},
		WeekDays:  []time.Weekday{time.Wednesday, time.Thursday},
		StartTime: "14:30:00",
		EndTime:   "15:00:00"}
	ap := &ActivationPeriod{ActivationTime: d}
	ap.AddInterval(i)
	ub := GetUB()

	ap1 := ActivationPeriod{}
	ub1 := &UserBalance{}
	b.StartTimer()
	ms := new(JSONBufMarshaler)
	for i := 0; i < b.N; i++ {
		result, _ := ms.Marshal(ap)
		ms.Unmarshal(result, ap1)
		result, _ = ms.Marshal(ub)
		ms.Unmarshal(result, ub1)
	}
}

func BenchmarkMarshallerGOBStoreRestore(b *testing.B) {
	b.StopTimer()
	d := time.Date(2012, time.February, 1, 14, 30, 1, 0, time.UTC)
	i := &Interval{Months: []time.Month{time.February},
		MonthDays: []int{1},
		WeekDays:  []time.Weekday{time.Wednesday, time.Thursday},
		StartTime: "14:30:00",
		EndTime:   "15:00:00"}
	ap := &ActivationPeriod{ActivationTime: d}
	ap.AddInterval(i)
	ub := GetUB()

	ap1 := ActivationPeriod{}
	ub1 := &UserBalance{}
	b.StartTimer()
	ms := new(GOBMarshaler)
	for i := 0; i < b.N; i++ {
		result, _ := ms.Marshal(ap)
		ms.Unmarshal(result, ap1)
		result, _ = ms.Marshal(ub)
		ms.Unmarshal(result, ub1)
	}
}

func BenchmarkMarshallerMsgpackStoreRestore(b *testing.B) {
	b.StopTimer()
	d := time.Date(2012, time.February, 1, 14, 30, 1, 0, time.UTC)
	i := &Interval{Months: []time.Month{time.February},
		MonthDays: []int{1},
		WeekDays:  []time.Weekday{time.Wednesday, time.Thursday},
		StartTime: "14:30:00",
		EndTime:   "15:00:00"}
	ap := &ActivationPeriod{ActivationTime: d}
	ap.AddInterval(i)
	ub := GetUB()

	ap1 := ActivationPeriod{}
	ub1 := &UserBalance{}
	b.StartTimer()
	ms := new(MsgpackMarshaler)
	for i := 0; i < b.N; i++ {
		result, _ := ms.Marshal(ap)
		ms.Unmarshal(result, ap1)

		result, _ = ms.Marshal(ub)
		ms.Unmarshal(result, ub1)
	}
}

func BenchmarkMarshallerCodecMsgpackStoreRestore(b *testing.B) {
	b.StopTimer()
	d := time.Date(2012, time.February, 1, 14, 30, 1, 0, time.UTC)
	i := &Interval{Months: []time.Month{time.February},
		MonthDays: []int{1},
		WeekDays:  []time.Weekday{time.Wednesday, time.Thursday},
		StartTime: "14:30:00",
		EndTime:   "15:00:00"}
	ap := &ActivationPeriod{ActivationTime: d}
	ap.AddInterval(i)
	ub := GetUB()

	ap1 := ActivationPeriod{}
	ub1 := &UserBalance{}
	b.StartTimer()
	ms := NewCodecMsgpackMarshaler()
	for i := 0; i < b.N; i++ {
		result, _ := ms.Marshal(ap)
		ms.Unmarshal(result, ap1)
		result, _ = ms.Marshal(ub)
		ms.Unmarshal(result, ub1)
	}
}

func BenchmarkMarshallerBincStoreRestore(b *testing.B) {
	b.StopTimer()
	d := time.Date(2012, time.February, 1, 14, 30, 1, 0, time.UTC)
	i := &Interval{Months: []time.Month{time.February},
		MonthDays: []int{1},
		WeekDays:  []time.Weekday{time.Wednesday, time.Thursday},
		StartTime: "14:30:00",
		EndTime:   "15:00:00"}
	ap := &ActivationPeriod{ActivationTime: d}
	ap.AddInterval(i)
	ub := GetUB()

	ap1 := ActivationPeriod{}
	ub1 := &UserBalance{}
	b.StartTimer()
	ms := NewBincMarshaler()
	for i := 0; i < b.N; i++ {
		result, _ := ms.Marshal(ap)
		ms.Unmarshal(result, ap1)
		result, _ = ms.Marshal(ub)
		ms.Unmarshal(result, ub1)
	}
}
