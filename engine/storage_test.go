/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2012-2015 ITsysCOM

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

	"github.com/cgrates/cgrates/cache2go"
	"github.com/cgrates/cgrates/utils"
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

func TestStorageDestinationContainsPrefixShort(t *testing.T) {
	dest, err := ratingStorage.GetDestination("NAT")
	precision := dest.containsPrefix("0723")
	if err != nil || precision != 4 {
		t.Error("Error finding prefix: ", err, precision)
	}
}

func TestStorageDestinationContainsPrefixLong(t *testing.T) {
	dest, err := ratingStorage.GetDestination("NAT")
	precision := dest.containsPrefix("0723045326")
	if err != nil || precision != 4 {
		t.Error("Error finding prefix: ", err, precision)
	}
}

func TestStorageDestinationContainsPrefixNotExisting(t *testing.T) {
	dest, err := ratingStorage.GetDestination("NAT")
	precision := dest.containsPrefix("072")
	if err != nil || precision != 0 {
		t.Error("Error finding prefix: ", err, precision)
	}
}

func TestStorageCacheRefresh(t *testing.T) {
	ratingStorage.SetDestination(&Destination{"T11", []string{"0"}})
	ratingStorage.GetDestination("T11")
	ratingStorage.SetDestination(&Destination{"T11", []string{"1"}})
	t.Log("Test cache refresh")
	err := ratingStorage.CacheRatingAll()
	if err != nil {
		t.Error("Error cache rating: ", err)
	}
	d, err := ratingStorage.GetDestination("T11")
	p := d.containsPrefix("1")
	if err != nil || p == 0 {
		t.Error("Error refreshing cache:", d)
	}
}

func TestStorageGetAliases(t *testing.T) {
	ala := &Alias{
		Direction: "*out",
		Tenant:    "vdf",
		Category:  "0",
		Account:   "b1",
		Subject:   "b1",
		Context:   utils.ALIAS_CONTEXT_RATING,
		Values: AliasValues{
			&AliasValue{
				Pairs:         AliasPairs{"Subject": map[string]string{"b1": "aaa"}},
				Weight:        10,
				DestinationId: utils.ANY,
			},
		},
	}
	alb := &Alias{
		Direction: "*out",
		Tenant:    "vdf",
		Category:  "0",
		Account:   "b1",
		Subject:   "b1",
		Context:   "*other",
		Values: AliasValues{
			&AliasValue{
				Pairs:         AliasPairs{"Account": map[string]string{"b1": "aaa"}},
				Weight:        10,
				DestinationId: utils.ANY,
			},
		},
	}
	accountingStorage.SetAlias(ala)
	accountingStorage.SetAlias(alb)
	foundAlias, err := accountingStorage.GetAlias(ala.GetId(), true)
	if err != nil || len(foundAlias.Values) != 1 {
		t.Errorf("Alias get error %+v, %v: ", foundAlias, err)
	}
	foundAlias, err = accountingStorage.GetAlias(alb.GetId(), true)
	if err != nil || len(foundAlias.Values) != 1 {
		t.Errorf("Alias get error %+v, %v: ", foundAlias, err)
	}
	foundAlias, err = accountingStorage.GetAlias(ala.GetId(), false)
	if err != nil || len(foundAlias.Values) != 1 {
		t.Errorf("Alias get error %+v, %v: ", foundAlias, err)
	}
	foundAlias, err = accountingStorage.GetAlias(alb.GetId(), false)
	if err != nil || len(foundAlias.Values) != 1 {
		t.Errorf("Alias get error %+v, %v: ", foundAlias, err)
	}
}

func TestStorageCacheGetReverseAliases(t *testing.T) {
	ala := &Alias{
		Direction: "*out",
		Tenant:    "vdf",
		Category:  "0",
		Account:   "b1",
		Subject:   "b1",
		Context:   utils.ALIAS_CONTEXT_RATING,
	}
	alb := &Alias{
		Direction: "*out",
		Tenant:    "vdf",
		Category:  "0",
		Account:   "b1",
		Subject:   "b1",
		Context:   "*other",
	}
	if x, err := cache2go.Get(utils.REVERSE_ALIASES_PREFIX + "aaa" + "Subject" + utils.ALIAS_CONTEXT_RATING); err == nil {
		aliasKeys := x.(map[interface{}]struct{})
		_, found := aliasKeys[utils.ConcatenatedKey(ala.GetId(), utils.ANY)]
		if !found {
			t.Error("Error getting reverse alias: ", aliasKeys, ala.GetId()+utils.ANY)
		}
	} else {
		t.Error("Error getting reverse alias: ", err)
	}
	if x, err := cache2go.Get(utils.REVERSE_ALIASES_PREFIX + "aaa" + "Account" + "*other"); err == nil {
		aliasKeys := x.(map[interface{}]struct{})
		_, found := aliasKeys[utils.ConcatenatedKey(alb.GetId(), utils.ANY)]
		if !found {
			t.Error("Error getting reverse alias: ", aliasKeys)
		}
	} else {
		t.Error("Error getting reverse alias: ", err)
	}
}

func TestStorageCacheRemoveCachedAliases(t *testing.T) {
	ala := &Alias{
		Direction: "*out",
		Tenant:    "vdf",
		Category:  "0",
		Account:   "b1",
		Subject:   "b1",
		Context:   utils.ALIAS_CONTEXT_RATING,
	}
	alb := &Alias{
		Direction: "*out",
		Tenant:    "vdf",
		Category:  "0",
		Account:   "b1",
		Subject:   "b1",
		Context:   utils.ALIAS_CONTEXT_RATING,
	}
	accountingStorage.RemoveAlias(ala.GetId())
	accountingStorage.RemoveAlias(alb.GetId())

	if _, err := cache2go.Get(utils.ALIASES_PREFIX + ala.GetId()); err == nil {
		t.Error("Error removing cached alias: ", err)
	}
	if _, err := cache2go.Get(utils.ALIASES_PREFIX + alb.GetId()); err == nil {
		t.Error("Error removing cached alias: ", err)
	}

	if _, err := cache2go.Get(utils.REVERSE_ALIASES_PREFIX + "aaa" + utils.ALIAS_CONTEXT_RATING); err == nil {
		t.Error("Error removing cached reverse alias: ", err)
	}
	if _, err := cache2go.Get(utils.REVERSE_ALIASES_PREFIX + "aaa" + utils.ALIAS_CONTEXT_RATING); err == nil {
		t.Error("Error removing cached reverse alias: ", err)
	}
}

func TestStorageDisabledAccount(t *testing.T) {
	acc, err := accountingStorage.GetAccount("cgrates.org:alodis")
	if err != nil || acc == nil {
		t.Error("Error loading disabled user account: ", err, acc)
	}
	if acc.Disabled != true || acc.AllowNegative != true {
		t.Errorf("Error loading user account properties: %+v", acc)
	}
}

// Install fails to detect them and starting server will panic, these tests will fix this
func TestStoreInterfaces(t *testing.T) {
	rds := new(RedisStorage)
	var _ RatingStorage = rds
	var _ AccountingStorage = rds
	sql := new(SQLStorage)
	var _ CdrStorage = sql
	var _ LogStorage = sql
}

func TestDifferentUuid(t *testing.T) {
	a1, err := accountingStorage.GetAccount("cgrates.org:12345")
	if err != nil {
		t.Error("Error getting account: ", err)
	}
	a2, err := accountingStorage.GetAccount("cgrates.org:123456")
	if err != nil {
		t.Error("Error getting account: ", err)
	}
	if a1.BalanceMap[utils.VOICE][0].Uuid == a2.BalanceMap[utils.VOICE][0].Uuid ||
		a1.BalanceMap[utils.MONETARY][0].Uuid == a2.BalanceMap[utils.MONETARY][0].Uuid {
		t.Errorf("Identical uuids in different accounts: %+v <-> %+v", a1.BalanceMap[utils.VOICE][0], a1.BalanceMap[utils.MONETARY][0])
	}
}

func TestStorageTask(t *testing.T) {
	// clean previous unused tasks
	for i := 0; i < 18; i++ {
		ratingStorage.PopTask()
	}

	if err := ratingStorage.PushTask(&Task{Uuid: "1"}); err != nil {
		t.Error("Error pushing task: ", err)
	}
	if err := ratingStorage.PushTask(&Task{Uuid: "2"}); err != nil {
		t.Error("Error pushing task: ", err)
	}
	if err := ratingStorage.PushTask(&Task{Uuid: "3"}); err != nil {
		t.Error("Error pushing task: ", err)
	}
	if err := ratingStorage.PushTask(&Task{Uuid: "4"}); err != nil {
		t.Error("Error pushing task: ", err)
	}
	if task, err := ratingStorage.PopTask(); err != nil && task.Uuid != "1" {
		t.Error("Error poping task: ", task, err)
	}
	if task, err := ratingStorage.PopTask(); err != nil && task.Uuid != "2" {
		t.Error("Error poping task: ", task, err)
	}
	if task, err := ratingStorage.PopTask(); err != nil && task.Uuid != "3" {
		t.Error("Error poping task: ", task, err)
	}
	if task, err := ratingStorage.PopTask(); err != nil && task.Uuid != "4" {
		t.Error("Error poping task: ", task, err)
	}
	if task, err := ratingStorage.PopTask(); err == nil && task != nil {
		t.Errorf("Error poping task %+v, %v: ", task, err)
	}
}

/************************** Benchmarks *****************************/

func GetUB() *Account {
	uc := &UnitCounter{
		Counters: CounterFilters{&CounterFilter{Value: 1}, &CounterFilter{Filter: &BalanceFilter{Weight: utils.Float64Pointer(20), DestinationIDs: utils.StringMapPointer(utils.NewStringMap("NAT"))}}, &CounterFilter{Filter: &BalanceFilter{Weight: utils.Float64Pointer(10), DestinationIDs: utils.StringMapPointer(utils.NewStringMap("RET"))}}},
	}
	at := &ActionTrigger{
		ID:             "some_uuid",
		ThresholdValue: 100.0,
		Balance: &BalanceFilter{
			Type:           utils.StringPointer(utils.MONETARY),
			Directions:     utils.StringMapPointer(utils.NewStringMap(utils.OUT)),
			DestinationIDs: utils.StringMapPointer(utils.NewStringMap("NAT")),
		},
		Weight:    10.0,
		ActionsId: "Commando",
	}
	var zeroTime time.Time
	zeroTime = zeroTime.UTC() // for deep equal to find location
	ub := &Account{
		Id:             "rif",
		AllowNegative:  true,
		BalanceMap:     map[string]BalanceChain{utils.SMS: BalanceChain{&Balance{Value: 14, ExpirationDate: zeroTime}}, utils.DATA: BalanceChain{&Balance{Value: 1024, ExpirationDate: zeroTime}}, utils.VOICE: BalanceChain{&Balance{Weight: 20, DestinationIds: utils.NewStringMap("NAT")}, &Balance{Weight: 10, DestinationIds: utils.NewStringMap("RET")}}},
		UnitCounters:   UnitCounters{utils.SMS: []*UnitCounter{uc, uc}},
		ActionTriggers: ActionTriggers{at, at, at},
	}
	return ub
}

func BenchmarkMarshallerJSONStoreRestore(b *testing.B) {
	b.StopTimer()
	i := &RateInterval{
		Timing: &RITiming{
			Months:    []time.Month{time.February},
			MonthDays: []int{1},
			WeekDays:  []time.Weekday{time.Wednesday, time.Thursday},
			StartTime: "14:30:00",
			EndTime:   "15:00:00"}}
	ap := &RatingPlan{Id: "test"}
	ap.AddRateInterval("NAT", i)
	ub := GetUB()

	ap1 := RatingPlan{}
	ub1 := &Account{}
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
	i := &RateInterval{
		Timing: &RITiming{
			Months:    []time.Month{time.February},
			MonthDays: []int{1},
			WeekDays:  []time.Weekday{time.Wednesday, time.Thursday},
			StartTime: "14:30:00",
			EndTime:   "15:00:00"}}
	ap := &RatingPlan{Id: "test"}
	ap.AddRateInterval("NAT", i)
	ub := GetUB()

	ap1 := RatingPlan{}
	ub1 := &Account{}
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
	i := &RateInterval{
		Timing: &RITiming{Months: []time.Month{time.February},
			MonthDays: []int{1},
			WeekDays:  []time.Weekday{time.Wednesday, time.Thursday},
			StartTime: "14:30:00",
			EndTime:   "15:00:00"}}
	ap := &RatingPlan{Id: "test"}
	ap.AddRateInterval("NAT", i)
	ub := GetUB()

	ap1 := RatingPlan{}
	ub1 := &Account{}
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
	i := &RateInterval{
		Timing: &RITiming{Months: []time.Month{time.February},
			MonthDays: []int{1},
			WeekDays:  []time.Weekday{time.Wednesday, time.Thursday},
			StartTime: "14:30:00",
			EndTime:   "15:00:00"}}
	ap := &RatingPlan{Id: "test"}
	ap.AddRateInterval("NAT", i)
	ub := GetUB()

	ap1 := RatingPlan{}
	ub1 := &Account{}
	b.StartTimer()
	ms := new(GOBMarshaler)
	for i := 0; i < b.N; i++ {
		result, _ := ms.Marshal(ap)
		ms.Unmarshal(result, ap1)
		result, _ = ms.Marshal(ub)
		ms.Unmarshal(result, ub1)
	}
}

func BenchmarkMarshallerCodecMsgpackStoreRestore(b *testing.B) {
	b.StopTimer()
	i := &RateInterval{
		Timing: &RITiming{
			Months:    []time.Month{time.February},
			MonthDays: []int{1},
			WeekDays:  []time.Weekday{time.Wednesday, time.Thursday},
			StartTime: "14:30:00",
			EndTime:   "15:00:00"}}
	ap := &RatingPlan{Id: "test"}
	ap.AddRateInterval("NAT", i)
	ub := GetUB()

	ap1 := RatingPlan{}
	ub1 := &Account{}
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
	i := &RateInterval{
		Timing: &RITiming{
			Months:    []time.Month{time.February},
			MonthDays: []int{1},
			WeekDays:  []time.Weekday{time.Wednesday, time.Thursday},
			StartTime: "14:30:00",
			EndTime:   "15:00:00"}}
	ap := &RatingPlan{Id: "test"}
	ap.AddRateInterval("NAT", i)
	ub := GetUB()

	ap1 := RatingPlan{}
	ub1 := &Account{}
	b.StartTimer()
	ms := NewBincMarshaler()
	for i := 0; i < b.N; i++ {
		result, _ := ms.Marshal(ap)
		ms.Unmarshal(result, ap1)
		result, _ = ms.Marshal(ub)
		ms.Unmarshal(result, ub1)
	}
}
