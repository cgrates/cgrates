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
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
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
	t1 := time.Date(2013, 8, 28, 22, 27, 30, 11, time.UTC)
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
	dest, err := dm.GetDestination("NAT", false, true, utils.NonTransactional)
	precision := dest.containsPrefix("0723")
	if err != nil || precision != 4 {
		t.Error("Error finding prefix: ", err, precision)
	}
}

func TestStorageDestinationContainsPrefixLong(t *testing.T) {
	dest, err := dm.GetDestination("NAT", false, true, utils.NonTransactional)
	precision := dest.containsPrefix("0723045326")
	if err != nil || precision != 4 {
		t.Error("Error finding prefix: ", err, precision)
	}
}

func TestStorageDestinationContainsPrefixNotExisting(t *testing.T) {
	dest, err := dm.GetDestination("NAT", false, true, utils.NonTransactional)
	precision := dest.containsPrefix("072")
	if err != nil || precision != 0 {
		t.Error("Error finding prefix: ", err, precision)
	}
}

/*
func TestStorageCacheRefresh(t *testing.T) {
	dm.SetDestination(&Destination{"T11", []string{"0"}}, utils.NonTransactional)
	dm.GetDestination("T11", false, utils.NonTransactional)
	dm.SetDestination(&Destination{"T11", []string{"1"}}, utils.NonTransactional)
	t.Log("Test cache refresh")
	err := LoadAllDataDBCache(dm)
	if err != nil {
		t.Error("Error cache rating: ", err)
	}
	d, err := dm.GetDestination("T11", false, utils.NonTransactional)
	p := d.containsPrefix("1")
	if err != nil || p == 0 {
		t.Error("Error refreshing cache:", d)
	}
}
*/

func TestStorageDisabledAccount(t *testing.T) {
	acc, err := dm.GetAccount("cgrates.org:alodis")
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
	a1, err := dm.GetAccount("cgrates.org:12345")
	if err != nil {
		t.Error("Error getting account: ", err)
	}
	a2, err := dm.GetAccount("cgrates.org:123456")
	if err != nil {
		t.Error("Error getting account: ", err)
	}
	if a1.BalanceMap[utils.MetaVoice][0].Uuid == a2.BalanceMap[utils.MetaVoice][0].Uuid ||
		a1.BalanceMap[utils.MetaMonetary][0].Uuid == a2.BalanceMap[utils.MetaMonetary][0].Uuid {
		t.Errorf("Identical uuids in different accounts: %+v <-> %+v", a1.BalanceMap[utils.MetaVoice][0], a1.BalanceMap[utils.MetaMonetary][0])
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
			Type:           utils.StringPointer(utils.MetaMonetary),
			DestinationIDs: utils.StringMapPointer(utils.NewStringMap("NAT")),
		},
		Weight:    10.0,
		ActionsID: "Commando",
	}
	var zeroTime time.Time
	zeroTime = zeroTime.UTC() // for deep equal to find location
	ub := &Account{
		ID:            "rif",
		AllowNegative: true,
		BalanceMap: map[string]Balances{
			utils.MetaSMS:  {&Balance{Value: 14, ExpirationDate: zeroTime}},
			utils.MetaData: {&Balance{Value: 1024, ExpirationDate: zeroTime}},
			utils.MetaVoice: {&Balance{Weight: 20, DestinationIDs: utils.NewStringMap("NAT")},
				&Balance{Weight: 10, DestinationIDs: utils.NewStringMap("RET")}}},
		UnitCounters:   UnitCounters{utils.MetaSMS: []*UnitCounter{uc, uc}},
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

func TestIDBRemoveIndexesDrv(t *testing.T) {
	idb := NewInternalDB(nil, nil, true, map[string]*config.ItemOpt{
		"chID": {
			Limit:     3,
			TTL:       4 * time.Minute,
			StaticTTL: false,
			Remote:    true,
			Replicate: true,

			RouteID: "route",
			APIKey:  "api",
		},
		"chID2": {

			Limit:     3,
			TTL:       4 * time.Minute,
			StaticTTL: false,
			Remote:    true,
			Replicate: true,

			RouteID: "route",
			APIKey:  "api",
		},
	},
	)
	idb.db.Set("chID", "itmID", true, []string{utils.EmptyString}, true, "trID")
	idb.db.Set("chID2", "itmIDv", true, []string{"grpID"}, true, "trID")

	if err := idb.RemoveIndexesDrv("chID", utils.EmptyString, utils.EmptyString); err != nil {
		t.Error(err)
	}
	if err := idb.RemoveIndexesDrv("chID2", "itmID", "v"); err != nil {
		t.Error(err)
	}
	if has := idb.db.HasGroup("chID", utils.EmptyString); has {
		t.Error("group should be removed")
	}
}

func TestIDBGetDispatcherHostDrv(t *testing.T) {
	idb := NewInternalDB(nil, nil, true, map[string]*config.ItemOpt{
		utils.CacheDispatcherHosts: {
			Limit:  2,
			Remote: true,
		},
	})
	dsp := &DispatcherHost{
		Tenant: "cgrates.org",
	}
	tenant, acc := "cgrates", "acc1"
	idb.db.Set(utils.CacheDispatcherHosts, utils.ConcatenatedKey(tenant, acc), dsp, []string{"id", "id3"}, true, utils.NonTransactional)

	if val, err := idb.GetDispatcherHostDrv(tenant, acc); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(dsp, val) {
		t.Errorf("expected %+v,received %+v", utils.ToJSON(dsp), utils.ToJSON(val))
	}
}

func TestIDBRemoveDispatcherHostDrv(t *testing.T) {
	idb := NewInternalDB(nil, nil, true, map[string]*config.ItemOpt{
		utils.CacheDispatcherHosts: {
			Limit:  2,
			Remote: true,
		},
	})
	dsp := &DispatcherHost{
		Tenant: "cgrates.org",
	}
	tenant, acc := "cgrates", "acc1"
	idb.db.Set(utils.CacheDispatcherHosts, utils.ConcatenatedKey(tenant, acc), dsp, []string{"id", "id3"}, true, utils.NonTransactional)

	if err := idb.RemoveDispatcherHostDrv(tenant, acc); err != nil {
		t.Error(err)
	}
	if _, has := idb.db.Get(utils.CacheDispatcherHosts, utils.ConcatenatedKey(tenant, acc)); has {
		t.Error("should been removed")
	}
}

func TestIDBSetStatQueueDrvNil(t *testing.T) {
	idb := NewInternalDB(nil, nil, true, map[string]*config.ItemOpt{
		utils.CacheStatQueues: {
			Limit:     4,
			StaticTTL: true,
		},
	})
	ssq := &StoredStatQueue{
		Tenant: "cgrates",
		ID:     "id",
		SQItems: []SQItem{
			{
				EventID: "event1",
			},
			{
				EventID: "event2",
			},
		},
		Compressed: false,
		SQMetrics: map[string][]byte{
			strings.Join([]string{utils.MetaASR, "test"}, "#"): []byte("val"),
		},
	}
	if err := idb.SetStatQueueDrv(ssq, nil); err == nil {
		t.Error(err)
	}
}

func TestGetTpTableIds(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.DataDbCfg().Items = map[string]*config.ItemOpt{
		utils.CacheTBLTPRates: {
			Limit:     3,
			StaticTTL: true,
		},
	}
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	tpid := "*prf"
	paginator := &utils.PaginatorWithSearch{
		Paginator: &utils.Paginator{},
		Search:    "",
	}
	expIds := []string{"Item1", "Item2"}
	filters := map[string]string{}
	db.db.Set(utils.CacheTBLTPRates, "*prf:Item1", "val", []string{"grpId"}, true, utils.NonTransactional)
	db.db.Set(utils.CacheTBLTPRates, "*prf:Item2", "val", []string{"grpId"}, true, utils.NonTransactional)
	if val, err := db.GetTpTableIds(tpid, utils.TBLTPRates, []string{}, filters, paginator); err != nil {
		t.Error(err)
	} else {
		sort.Slice(val, func(i, j int) bool {
			return val[i] < val[j]
		})
		if !reflect.DeepEqual(val, expIds) {
			t.Errorf("expected %v,received %v", utils.ToJSON(val), utils.ToJSON(expIds))
		}
	}
}
func TestIDBGetTpIds(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.DataDbCfg().Items = map[string]*config.ItemOpt{
		utils.CacheTBLTPRates: {
			Limit:     3,
			StaticTTL: true,
		},
	}
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	db.db.Set(utils.CacheTBLTPRates, "item_ID1", "value", []string{"grpID"}, true, utils.NonTransactional)
	db.db.Set(utils.CacheTBLTPRates, "item_ID2", "value", []string{"grpID"}, true, utils.NonTransactional)
	exp := []string{"item_ID1", "item_ID2"}
	val, err := db.GetTpIds(utils.TBLTPRates)
	if err != nil {
		t.Error(err)
	}
	sort.Slice(val, func(i, j int) bool {
		return val[i] < val[j]
	})
	if !reflect.DeepEqual(val, exp) {
		t.Errorf("expected %v,received %v", utils.ToJSON(val), utils.ToJSON(exp))
	}
}
