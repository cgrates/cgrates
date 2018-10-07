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
	"testing"
	"time"

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
	dest, err := dm.DataDB().GetDestination("NAT", true, utils.NonTransactional)
	precision := dest.containsPrefix("0723")
	if err != nil || precision != 4 {
		t.Error("Error finding prefix: ", err, precision)
	}
}

func TestStorageDestinationContainsPrefixLong(t *testing.T) {
	dest, err := dm.DataDB().GetDestination("NAT", true, utils.NonTransactional)
	precision := dest.containsPrefix("0723045326")
	if err != nil || precision != 4 {
		t.Error("Error finding prefix: ", err, precision)
	}
}

func TestStorageDestinationContainsPrefixNotExisting(t *testing.T) {
	dest, err := dm.DataDB().GetDestination("NAT", true, utils.NonTransactional)
	precision := dest.containsPrefix("072")
	if err != nil || precision != 0 {
		t.Error("Error finding prefix: ", err, precision)
	}
}

/*
func TestStorageCacheRefresh(t *testing.T) {
	dm.DataDB().SetDestination(&Destination{"T11", []string{"0"}}, utils.NonTransactional)
	dm.DataDB().GetDestination("T11", false, utils.NonTransactional)
	dm.DataDB().SetDestination(&Destination{"T11", []string{"1"}}, utils.NonTransactional)
	t.Log("Test cache refresh")
	err := dm.LoadDataDBCache(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	if err != nil {
		t.Error("Error cache rating: ", err)
	}
	d, err := dm.DataDB().GetDestination("T11", false, utils.NonTransactional)
	p := d.containsPrefix("1")
	if err != nil || p == 0 {
		t.Error("Error refreshing cache:", d)
	}
}
*/

func TestStorageGetAliases(t *testing.T) {
	ala := &Alias{
		Direction: "*out",
		Tenant:    "vdf",
		Category:  "0",
		Account:   "b1",
		Subject:   "b1",
		Context:   utils.MetaRating,
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
	dm.DataDB().SetAlias(ala, utils.NonTransactional)
	dm.DataDB().SetReverseAlias(ala, utils.NonTransactional)
	dm.DataDB().SetAlias(alb, utils.NonTransactional)
	dm.DataDB().SetReverseAlias(alb, utils.NonTransactional)
	foundAlias, err := dm.DataDB().GetAlias(ala.GetId(), true, utils.NonTransactional)
	if err != nil || len(foundAlias.Values) != 1 {
		t.Errorf("Alias get error %+v, %v: ", foundAlias, err)
	}
	foundAlias, err = dm.DataDB().GetAlias(alb.GetId(), true, utils.NonTransactional)
	if err != nil || len(foundAlias.Values) != 1 {
		t.Errorf("Alias get error %+v, %v: ", foundAlias, err)
	}
	foundAlias, err = dm.DataDB().GetAlias(ala.GetId(), false, utils.NonTransactional)
	if err != nil || len(foundAlias.Values) != 1 {
		t.Errorf("Alias get error %+v, %v: ", foundAlias, err)
	}
	foundAlias, err = dm.DataDB().GetAlias(alb.GetId(), false, utils.NonTransactional)
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
		Context:   utils.MetaRating,
	}
	alb := &Alias{
		Direction: "*out",
		Tenant:    "vdf",
		Category:  "0",
		Account:   "b1",
		Subject:   "b1",
		Context:   "*other",
	}
	dm.DataDB().GetReverseAlias("aaa"+"Subject"+utils.MetaRating, false, utils.NonTransactional)
	if x, ok := Cache.Get(utils.CacheReverseAliases, "aaa"+"Subject"+utils.MetaRating); ok {
		aliasKeys := x.([]string)
		if len(aliasKeys) != 1 {
			t.Error("Error getting reverse alias: ", aliasKeys, ala.GetId()+utils.ANY)
		}
	} else {
		t.Error("Error getting reverse alias: ", err)
	}
	dm.DataDB().GetReverseAlias("aaa"+"Account"+"*other", false, utils.NonTransactional)
	if x, ok := Cache.Get(utils.CacheReverseAliases, "aaa"+"Account"+"*other"); ok {
		aliasKeys := x.([]string)
		if len(aliasKeys) != 1 {
			t.Error("Error getting reverse alias: ", aliasKeys, alb.GetId()+utils.ANY)
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
		Context:   utils.MetaRating,
	}
	alb := &Alias{
		Direction: "*out",
		Tenant:    "vdf",
		Category:  "0",
		Account:   "b1",
		Subject:   "b1",
		Context:   "*other",
	}
	dm.DataDB().RemoveAlias(ala.GetId(), utils.NonTransactional)
	dm.DataDB().RemoveAlias(alb.GetId(), utils.NonTransactional)

	if _, ok := Cache.Get(utils.CacheAliases, ala.GetId()); ok {
		t.Error("Error removing cached alias: ", ok)
	}
	if _, ok := Cache.Get(utils.CacheAliases, alb.GetId()); ok {
		t.Error("Error removing cached alias: ", ok)
	}

	if _, ok := Cache.Get(utils.CacheReverseAliases, "aaa"+utils.MetaRating); ok {
		t.Error("Error removing cached reverse alias: ", ok)
	}
	if _, ok := Cache.Get(utils.CacheReverseAliases, "aaa"+utils.MetaRating); ok {
		t.Error("Error removing cached reverse alias: ", ok)
	}
}

func TestStorageDisabledAccount(t *testing.T) {
	acc, err := dm.DataDB().GetAccount("cgrates.org:alodis")
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
	var _ DataDB = rds
	sql := new(SQLStorage)
	var _ CdrStorage = sql
}

func TestDifferentUuid(t *testing.T) {
	a1, err := dm.DataDB().GetAccount("cgrates.org:12345")
	if err != nil {
		t.Error("Error getting account: ", err)
	}
	a2, err := dm.DataDB().GetAccount("cgrates.org:123456")
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
	for i := 0; i < 21; i++ {
		dm.DataDB().PopTask()
	}

	if err := dm.DataDB().PushTask(&Task{Uuid: "1"}); err != nil {
		t.Error("Error pushing task: ", err)
	}
	if err := dm.DataDB().PushTask(&Task{Uuid: "2"}); err != nil {
		t.Error("Error pushing task: ", err)
	}
	if err := dm.DataDB().PushTask(&Task{Uuid: "3"}); err != nil {
		t.Error("Error pushing task: ", err)
	}
	if err := dm.DataDB().PushTask(&Task{Uuid: "4"}); err != nil {
		t.Error("Error pushing task: ", err)
	}
	if task, err := dm.DataDB().PopTask(); err != nil && task.Uuid != "1" {
		t.Error("Error poping task: ", task, err)
	}
	if task, err := dm.DataDB().PopTask(); err != nil && task.Uuid != "2" {
		t.Error("Error poping task: ", task, err)
	}
	if task, err := dm.DataDB().PopTask(); err != nil && task.Uuid != "3" {
		t.Error("Error poping task: ", task, err)
	}
	if task, err := dm.DataDB().PopTask(); err != nil && task.Uuid != "4" {
		t.Error("Error poping task: ", task, err)
	}
	if task, err := dm.DataDB().PopTask(); err == nil && task != nil {
		t.Errorf("Error poping task %+v, %v ", task, err)
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
		ActionsID: "Commando",
	}
	var zeroTime time.Time
	zeroTime = zeroTime.UTC() // for deep equal to find location
	ub := &Account{
		ID:             "rif",
		AllowNegative:  true,
		BalanceMap:     map[string]Balances{utils.SMS: Balances{&Balance{Value: 14, ExpirationDate: zeroTime}}, utils.DATA: Balances{&Balance{Value: 1024, ExpirationDate: zeroTime}}, utils.VOICE: Balances{&Balance{Weight: 20, DestinationIDs: utils.NewStringMap("NAT")}, &Balance{Weight: 10, DestinationIDs: utils.NewStringMap("RET")}}},
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
