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
	"encoding/json"
	"reflect"
	"testing"
	"time"
)

func TestApRestoreFromStorage(t *testing.T) {
	cd := &CallDescriptor{
		Direction:   OUTBOUND,
		TOR:         "0",
		Tenant:      "CUSTOMER_1",
		Subject:     "rif:from:tm",
		Destination: "49"}
	cd.LoadActivationPeriods()
	if len(cd.ActivationPeriods) != 2 {
		t.Error("Error restoring activation periods: ", len(cd.ActivationPeriods))
	}
}

func TestApStoreRestoreJson(t *testing.T) {
	d := time.Date(2012, time.February, 1, 14, 30, 1, 0, time.UTC)
	i := &Interval{Months: []time.Month{time.February},
		MonthDays: []int{1},
		WeekDays:  []time.Weekday{time.Wednesday, time.Thursday},
		StartTime: "14:30:00",
		EndTime:   "15:00:00"}
	ap := &ActivationPeriod{ActivationTime: d}
	ap.AddInterval(i)
	result, _ := json.Marshal(ap)
	ap1 := &ActivationPeriod{}
	json.Unmarshal(result, ap1)
	if !reflect.DeepEqual(ap, ap1) {
		t.Errorf("Expected %v was %v", ap, ap1)
	}
}

func TestApStoreRestoreBlank(t *testing.T) {
	d := time.Date(2012, time.February, 1, 14, 30, 1, 0, time.UTC)
	i := &Interval{}
	ap := &ActivationPeriod{ActivationTime: d}
	ap.AddInterval(i)
	result, _ := json.Marshal(ap)
	ap1 := ActivationPeriod{}
	json.Unmarshal(result, &ap1)
	if reflect.DeepEqual(ap, ap1) {
		t.Errorf("Expected %v was %v", ap, ap1)
	}
}

func TestFallbackDirect(t *testing.T) {
	cd := &CallDescriptor{TOR: "0", Direction: OUTBOUND, Tenant: "CUSTOMER_2", Subject: "danb:87.139.12.167", Destination: "41"}
	cd.LoadActivationPeriods()
	if len(cd.ActivationPeriods) != 1 {
		t.Error("Error restoring activation periods: ", len(cd.ActivationPeriods))
	}
}

func TestFallbackMultiple(t *testing.T) {
	cd := &CallDescriptor{TOR: "0", Direction: OUTBOUND, Tenant: "vdf", Subject: "fall", Destination: "0723045"}
	cd.LoadActivationPeriods()
	if len(cd.ActivationPeriods) != 1 {
		t.Error("Error restoring activation periods: ", len(cd.ActivationPeriods))
	}
}

func TestFallbackWithBackTrace(t *testing.T) {
	cd := &CallDescriptor{TOR: "0", Direction: OUTBOUND, Tenant: "CUSTOMER_2", Subject: "danb:87.139.12.167", Destination: "4123"}
	cd.LoadActivationPeriods()
	if len(cd.ActivationPeriods) != 1 {
		t.Error("Error restoring activation periods: ", len(cd.ActivationPeriods))
	}
}

func TestFallbackDefault(t *testing.T) {
	cd := &CallDescriptor{TOR: "0", Direction: OUTBOUND, Tenant: "vdf", Subject: "one", Destination: "0723"}
	cd.LoadActivationPeriods()
	if len(cd.ActivationPeriods) != 1 {
		t.Error("Error restoring activation periods: ", len(cd.ActivationPeriods))
	}
}

func TestFallbackNoInfiniteLoop(t *testing.T) {
	cd := &CallDescriptor{TOR: "0", Direction: OUTBOUND, Tenant: "vdf", Subject: "rif", Destination: "0721"}
	cd.LoadActivationPeriods()
	if len(cd.ActivationPeriods) != 0 {
		t.Error("Error restoring activation periods: ", len(cd.ActivationPeriods))
	}
}

func TestFallbackNoInfiniteLoopSelf(t *testing.T) {
	cd := &CallDescriptor{TOR: "0", Direction: OUTBOUND, Tenant: "vdf", Subject: "inf", Destination: "0721"}
	cd.LoadActivationPeriods()
	if len(cd.ActivationPeriods) != 0 {
		t.Error("Error restoring activation periods: ", len(cd.ActivationPeriods))
	}
}

func TestApAddIntervalIfNotPresent(t *testing.T) {
	i1 := &Interval{Months: Months{time.February},
		MonthDays: MonthDays{1},
		WeekDays:  []time.Weekday{time.Wednesday, time.Thursday},
		StartTime: "14:30:00",
		EndTime:   "15:00:00"}
	i2 := &Interval{Months: Months{time.February},
		MonthDays: MonthDays{1},
		WeekDays:  []time.Weekday{time.Wednesday, time.Thursday},
		StartTime: "14:30:00",
		EndTime:   "15:00:00"}
	i3 := &Interval{Months: Months{time.February},
		MonthDays: MonthDays{1},
		WeekDays:  []time.Weekday{time.Wednesday},
		StartTime: "14:30:00",
		EndTime:   "15:00:00"}
	ap := &ActivationPeriod{}
	ap.AddInterval(i1)
	ap.AddInterval(i2)
	if len(ap.Intervals) != 1 {
		t.Error("Wronfully appended interval ;)")
	}
	ap.AddInterval(i3)
	if len(ap.Intervals) != 2 {
		t.Error("Wronfully not appended interval ;)")
	}
}

func TestApAddIntervalGroups(t *testing.T) {
	i1 := &Interval{
		Prices: PriceGroups{&Price{0, 1, 1 * time.Second, 1 * time.Second}},
	}
	i2 := &Interval{
		Prices: PriceGroups{&Price{30 * time.Second, 2, 1 * time.Second, 1 * time.Second}},
	}
	i3 := &Interval{
		Prices: PriceGroups{&Price{30 * time.Second, 2, 1 * time.Second, 1 * time.Second}},
	}
	ap := &ActivationPeriod{}
	ap.AddInterval(i1)
	ap.AddInterval(i2)
	ap.AddInterval(i3)
	if len(ap.Intervals) != 1 {
		t.Error("Wronfully appended interval ;)")
	}
	if len(ap.Intervals[0].Prices) != 2 {
		t.Error("Group prices not formed: ", ap.Intervals[0].Prices)
	}
}

/**************************** Benchmarks *************************************/

func BenchmarkActivationPeriodStoreRestoreJson(b *testing.B) {
	b.StopTimer()
	d := time.Date(2012, time.February, 1, 14, 30, 1, 0, time.UTC)
	i := &Interval{Months: []time.Month{time.February},
		MonthDays: []int{1},
		WeekDays:  []time.Weekday{time.Wednesday, time.Thursday},
		StartTime: "14:30:00",
		EndTime:   "15:00:00"}
	ap := &ActivationPeriod{ActivationTime: d}
	ap.AddInterval(i)

	ap1 := ActivationPeriod{}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		result, _ := json.Marshal(ap)
		json.Unmarshal(result, &ap1)
	}
}

func BenchmarkActivationPeriodStoreRestore(b *testing.B) {
	b.StopTimer()
	d := time.Date(2012, time.February, 1, 14, 30, 1, 0, time.UTC)
	i := &Interval{Months: []time.Month{time.February},
		MonthDays: []int{1},
		WeekDays:  []time.Weekday{time.Wednesday, time.Thursday},
		StartTime: "14:30:00",
		EndTime:   "15:00:00"}
	ap := &ActivationPeriod{ActivationTime: d}
	ap.AddInterval(i)

	ap1 := &ActivationPeriod{}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		result, _ := marsh.Marshal(ap)
		marsh.Unmarshal(result, ap1)
	}
}

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
