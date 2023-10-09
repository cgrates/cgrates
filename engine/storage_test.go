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
	"bytes"
	"reflect"
	"slices"
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

func TestStorageDecodeCodecMsgpackMarshaler(t *testing.T) {
	type stc struct {
		Name string
	}

	var s stc
	mp := make(map[string]any)
	var slc []string
	var slcB []byte
	var arr *[1]int
	var nm int
	var fl float64
	var str string
	var bl bool
	var td time.Duration

	tests := []struct {
		name     string
		expBytes []byte
		val      any
		decode   any
		rng      bool
	}{
		{
			name:     "map",
			expBytes: []byte{129, 164, 107, 101, 121, 49, 166, 118, 97, 108, 117, 101, 49},
			val:      map[string]any{"key1": "value1"},
			decode:   mp,
			rng:      true,
		},
		{
			name:     "int",
			expBytes: []byte{1},
			val:      1,
			decode:   nm,
			rng:      false,
		},
		{
			name:     "string",
			expBytes: []byte{164, 116, 101, 115, 116},
			val:      "test",
			decode:   str,
			rng:      false,
		},
		{
			name:     "float64",
			expBytes: []byte{203, 63, 248, 0, 0, 0, 0, 0, 0},
			val:      1.5,
			decode:   fl,
			rng:      false,
		},
		{
			name:     "boolean",
			expBytes: []byte{195},
			val:      true,
			decode:   bl,
			rng:      false,
		},
		{
			name:     "slice",
			expBytes: []byte{145, 164, 118, 97, 108, 49},
			val:      []string{"val1"},
			decode:   slc,
			rng:      true,
		},
		{
			name:     "array",
			expBytes: []byte{145, 1},
			val:      &[1]int{1},
			decode:   arr,
			rng:      true,
		},
		{
			name:     "struct",
			expBytes: []byte{129, 164, 78, 97, 109, 101, 164, 116, 101, 115, 116},
			val:      stc{"test"},
			decode:   s,
			rng:      true,
		},
		{
			name:     "time duration",
			expBytes: []byte{210, 59, 154, 202, 0},
			val:      1 * time.Second,
			decode:   td,
			rng:      false,
		},
		{
			name:     "slice of bytes",
			expBytes: []byte{162, 5, 8},
			val:      []byte{5, 8},
			decode:   slcB,
			rng:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := NewCodecMsgpackMarshaler()

			b, err := ms.Marshal(tt.val)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(b, tt.expBytes) {
				t.Fatalf("expected: %+v,\nreceived: %+v", tt.expBytes, b)
			}

			err = ms.Unmarshal(b, &tt.decode)
			if err != nil {
				t.Fatal(err)
			}

			if tt.rng {
				if !reflect.DeepEqual(tt.decode, tt.val) {
					t.Errorf("expected %v, received %v", tt.val, tt.decode)
				}
			} else {
				if tt.decode != tt.val {
					t.Errorf("expected %v, received %v", tt.val, tt.decode)
				}
			}
		})
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

func TestIDBTpResources(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	storDB := NewInternalDB(nil, nil, false, cfg.StorDbCfg().Items)
	// READ
	if _, err := storDB.GetTPResources("TP1", utils.EmptyString, utils.EmptyString); err != utils.ErrNotFound {
		t.Error(err)
	}
	//WRITE
	var snd = []*utils.TPResourceProfile{
		{
			TPid:         "TP1",
			ID:           "RP1",
			Weight:       10.8,
			FilterIDs:    []string{"FILTR_RES_1"},
			ThresholdIDs: []string{"TH1"},
			Stored:       true,
		},
		{
			TPid:         "TP1",
			ID:           "RP2",
			Weight:       20.6,
			ThresholdIDs: []string{"TH2"},
			FilterIDs:    []string{"FLTR_RES_2"},
			Blocker:      true,
			Stored:       false,
		},
	}
	if err := storDB.SetTPResources(snd); err != nil {
		t.Error(err)
	}
	// READ
	if rcv, err := storDB.GetTPResources("TP1", utils.EmptyString, utils.EmptyString); err != nil {
		t.Error(err)
	} else if sort.Slice(rcv, func(a, b int) bool {
		return rcv[a].ID < rcv[b].ID
	}); !slices.Equal(snd, rcv) {
		t.Errorf("Expecting: %+v, received: %+v ", utils.ToJSON(snd), utils.ToJSON(rcv))
	}
	// UPDATE
	snd[0].Weight = 2.1
	snd[1].Weight = 8.1
	if err := storDB.SetTPResources(snd); err != nil {
		t.Error(err)
	}
	// READ
	if rcv, err := storDB.GetTPResources("TP1", utils.EmptyString, utils.EmptyString); err != nil {
		t.Error(err)
	} else if sort.Slice(rcv, func(a, b int) bool {
		return rcv[a].ID < rcv[b].ID
	}); rcv[0].Weight != snd[0].Weight {
		t.Errorf("Expecting: %+v, received: %+v ", utils.ToJSON(snd[0]), utils.ToJSON(rcv[0]))
	}
	//tpIDs
	expIds := []string{":RP1", ":RP2"}
	if tpIds, err := storDB.GetTpTableIds("TP1", utils.TBLTPResources, utils.TPDistinctIds{utils.TenantCfg, utils.IDCfg}, nil, &utils.PaginatorWithSearch{}); err != nil {
		t.Error(err)
	} else if slices.Sort(tpIds); !slices.Equal(tpIds, expIds) {
		t.Errorf("Expected %v,Received %v", expIds, tpIds)
	}
	// REMOVE
	if err := storDB.RemTpData("", "TP1", nil); err != nil {
		t.Error(err)
	}
	// READ
	if _, err := storDB.GetTPResources("TP1", utils.EmptyString, utils.EmptyString); err != utils.ErrNotFound {
		t.Error(err)
	}
}

func TestIDBTpStats(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	storDB := NewInternalDB(nil, nil, false, cfg.StorDbCfg().Items)

	// READ
	if _, err := storDB.GetTPStats("TP1", utils.EmptyString, utils.EmptyString); err != utils.ErrNotFound {
		t.Error(err)
	}
	//WRITE
	eTPs := []*utils.TPStatProfile{
		{
			TPid:        "TP1",
			Tenant:      "cgrates.org",
			ID:          "Stats1",
			FilterIDs:   []string{"FLTR_1"},
			QueueLength: 100,
			TTL:         "1s",
			Metrics: []*utils.MetricWithFilters{
				{
					MetricID: utils.MetaASR,
				},
			},
			ThresholdIDs: []string{"*none"},
			Weight:       20.0,
			Stored:       true,
			MinItems:     1,
		},
	}

	if err := storDB.SetTPStats(eTPs); err != nil {
		t.Error(err)
	}
	// READ
	if rcv, err := storDB.GetTPStats("TP1", utils.EmptyString, utils.EmptyString); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eTPs[0], rcv[0]) {
		t.Errorf("Expecting: %+v,\n received: %+v", utils.ToJSON(eTPs[0]), utils.ToJSON(rcv[0]))
	}

	// UPDATE
	eTPs[0].Metrics = []*utils.MetricWithFilters{
		{
			MetricID: utils.MetaACD,
		},
	}
	if err := storDB.SetTPStats(eTPs); err != nil {
		t.Error(err)
	}
	// READ
	if rcv, err := storDB.GetTPStats("TP1", utils.EmptyString, utils.EmptyString); err != nil {
		t.Error(err)
	} else if eTPs[0].Metrics[0].MetricID != rcv[0].Metrics[0].MetricID {
		t.Errorf("Expecting: %+v,\n received:  %+v", utils.ToJSON(eTPs[0]), utils.ToJSON(rcv[0]))
	}

	// REMOVE
	if err := storDB.RemTpData(utils.TBLTPStats, "TP1", nil); err != nil {
		t.Error(err)
	}
	// READ
	if ids, err := storDB.GetTPStats("TP1", utils.EmptyString, utils.EmptyString); err != utils.ErrNotFound {
		t.Error(err)
		t.Error(utils.ToJSON(ids))
	}
}

func TestIDBTPThresholds(t *testing.T) {

	storDB := NewInternalDB(nil, nil, false, config.CgrConfig().StorDbCfg().Items)
	//READ
	if _, err := storDB.GetTPThresholds("TH1", utils.EmptyString, utils.EmptyString); err != utils.ErrNotFound {
		t.Error(err)
	}

	//WRITE
	tpThresholds := []*utils.TPThresholdProfile{
		{
			TPid:      "TP1",
			Tenant:    "cgrates.org",
			ID:        "Th1",
			FilterIDs: []string{"*string:~*req.Account:1002", "*string:~*req.DryRun:*default"},
			MaxHits:   -1,
			MinSleep:  "1s",
			Blocker:   true,
			Weight:    10,
			ActionIDs: []string{"ACT_TOPUP_RST"},
			Async:     true,
		},
		{
			TPid:      "TP1",
			Tenant:    "cgrates.org",
			ID:        "Th2",
			FilterIDs: []string{"*string:~*req.Destination:10"},
			MaxHits:   -1,
			MinSleep:  "1s",
			Blocker:   true,
			Weight:    20,
			ActionIDs: []string{"ACT_LOG_WARNING"},
			Async:     true,
		},
	}
	if err := storDB.SetTPThresholds(tpThresholds); err != nil {
		t.Errorf("Unable to set TPThresholds")
	}

	//READ
	if rcv, err := storDB.GetTPThresholds(tpThresholds[0].TPid, utils.EmptyString, utils.EmptyString); err != nil {
		t.Error(err)
	} else if sort.Slice(rcv, func(i, j int) bool {
		return rcv[i].ID < rcv[j].ID
	}); !slices.Equal(rcv, tpThresholds) {
		t.Errorf("Expecting: %+v , Received: %+v", utils.ToJSON(tpThresholds), utils.ToJSON(rcv))
	}

	//UPDATE
	tpThresholds[0].FilterIDs = []string{"*string:~*req.Destination:101"}
	tpThresholds[1].FilterIDs = []string{"*string:~*req.Destination:101"}
	if err := storDB.SetTPThresholds(tpThresholds); err != nil {
		t.Errorf("Unable to set TPThresholds")
	}

	//READ
	if rcv, err := storDB.GetTPThresholds(tpThresholds[0].TPid, utils.EmptyString, utils.EmptyString); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(tpThresholds[0], rcv[0]) &&
		!reflect.DeepEqual(tpThresholds[0], rcv[1]) {
		t.Errorf("Expecting:\n%+v\nReceived:\n%+v\n||\n%+v",
			utils.ToJSON(tpThresholds[0]), utils.ToJSON(rcv[0]), utils.ToJSON(rcv[1]))
	}

	//REMOVE and READ
	if err := storDB.RemTpData(utils.EmptyString, tpThresholds[0].TPid, nil); err != nil {
		t.Error(err)
	} else if _, err := storDB.GetTPRoutes(tpThresholds[0].TPid, utils.EmptyString, utils.EmptyString); err != utils.ErrNotFound {
		t.Error(err)
	}
}

func TestIDBTPFilters(t *testing.T) {
	storDB := NewInternalDB(nil, nil, false, config.CgrConfig().StorDbCfg().Items)
	//READ
	if _, err := storDB.GetTPFilters("TP1", utils.EmptyString, utils.EmptyString); err != utils.ErrNotFound {
		t.Error(err)
	}

	//WRITE
	tpFilters := []*utils.TPFilterProfile{
		{
			TPid:   "TP1",
			Tenant: "cgrates.org",
			ID:     "Filter1",
			Filters: []*utils.TPFilter{
				{
					Type:    utils.MetaString,
					Element: "Account",
					Values:  []string{"1001", "1002"},
				},
			},
		},
		{
			TPid:   "TP1",
			Tenant: "cgrates.org",
			ID:     "Filter2",
			Filters: []*utils.TPFilter{
				{
					Type:    utils.MetaPrefix,
					Element: "Destination",
					Values:  []string{"10"},
				},
			},
		},
	}
	if err := storDB.SetTPFilters(tpFilters); err != nil {
		t.Errorf("Unable to set TPFilters")
	}

	//READ
	if rcv, err := storDB.GetTPFilters(tpFilters[0].TPid, utils.EmptyString, utils.EmptyString); err != nil {
		t.Error(err)
	} else if sort.Slice(rcv, func(i, j int) bool {
		return rcv[i].ID < rcv[j].ID
	}); !slices.Equal(rcv, tpFilters) {
		t.Errorf("Expecting: %+v , Received: %+v", utils.ToJSON(tpFilters), utils.ToJSON(rcv))
	}

	//UPDATE and WRITE
	tpFilters[1].Filters[0].Element = "Account"
	if err := storDB.SetTPFilters(tpFilters); err != nil {
		t.Errorf("Unable to set TPFilters")
	}

	//READ
	if rcv, err := storDB.GetTPFilters(tpFilters[0].TPid, utils.EmptyString, utils.EmptyString); err != nil {
		t.Error(err)
	} else if sort.Slice(rcv, func(i, j int) bool {
		return rcv[i].ID < rcv[j].ID
	}); rcv[1].Filters[0].Element != tpFilters[1].Filters[0].Element {
		t.Errorf("Expecting: %+v , Received: %+v", utils.ToJSON(tpFilters[1]), utils.ToJSON(rcv[1]))
	}

	//REMOVE and READ
	if err := storDB.RemTpData(utils.EmptyString, tpFilters[0].TPid, nil); err != nil {
		t.Error(err)
	} else if _, err := storDB.GetTPFilters(tpFilters[0].TPid, utils.EmptyString, utils.EmptyString); err != utils.ErrNotFound {
		t.Error(err)
	}
}

func TestIDTPRoutes(t *testing.T) {
	storDB := NewInternalDB(nil, nil, false, config.CgrConfig().StorDbCfg().Items)
	//READ
	if _, err := storDB.GetTPRoutes("TP1", utils.EmptyString, utils.EmptyString); err != utils.ErrNotFound {
		t.Error(err)
	}
	//WRITE
	tpRoutes := []*utils.TPRouteProfile{
		{
			TPid:      "TP1",
			Tenant:    "cgrates.org",
			ID:        "SUPL_1",
			FilterIDs: []string{"*string:~*req.Accout:1007"},
			Sorting:   "*lowest_cost",
			Routes: []*utils.TPRoute{
				{
					ID:              "supplier1",
					FilterIDs:       []string{"FLTR_1"},
					AccountIDs:      []string{"Acc1", "Acc2"},
					RatingPlanIDs:   []string{"RPL_1"},
					ResourceIDs:     []string{"ResGroup1"},
					StatIDs:         []string{"Stat1"},
					Weight:          10,
					Blocker:         false,
					RouteParameters: "SortingParam1",
				},
			},
			Weight: 20,
		},
		{
			TPid:      "TP1",
			Tenant:    "cgrates.org",
			ID:        "SUPL_2",
			FilterIDs: []string{"*string:~*req.Destination:100"},
			Sorting:   "*lowest_cost",
			Routes: []*utils.TPRoute{
				{
					ID:              "supplier1",
					FilterIDs:       []string{"FLTR_1"},
					AccountIDs:      []string{"Acc1", "Acc2"},
					RatingPlanIDs:   []string{"RPL_1"},
					ResourceIDs:     []string{"ResGroup1"},
					StatIDs:         []string{"Stat1"},
					Weight:          10,
					Blocker:         false,
					RouteParameters: "SortingParam2",
				},
			},
			Weight: 10,
		},
	}
	if err := storDB.SetTPRoutes(tpRoutes); err != nil {
		t.Errorf("Unable to set TPRoutes")
	}

	//READ
	if rcv, err := storDB.GetTPRoutes(tpRoutes[0].TPid, utils.EmptyString, utils.EmptyString); err != nil {
		t.Error(err)
	} else if sort.Slice(rcv, func(i, j int) bool {
		return rcv[i].ID < rcv[j].ID
	}); !slices.Equal(rcv, tpRoutes) {
		t.Errorf("Expecting: %v Received: %+v", utils.ToJSON(tpRoutes), utils.ToJSON(rcv))
	}

	//UPDATE
	tpRoutes[0].Sorting = "*higher_cost"
	if err := storDB.SetTPRoutes(tpRoutes); err != nil {
		t.Errorf("Unable to set TPRoutes")
	}
	//READ
	if rcv, err := storDB.GetTPRoutes(tpRoutes[0].TPid, utils.EmptyString, utils.EmptyString); err != nil {
		t.Error(err)
	} else if sort.Slice(rcv, func(i, j int) bool {
		return rcv[i].ID < rcv[j].ID
	}); tpRoutes[0].Sorting != rcv[0].Sorting {
		t.Errorf("Expecting: %v Received: %+v", utils.ToJSON(tpRoutes[0]), utils.ToJSON(rcv[0]))
	}

	//REMOVE and READ
	if err := storDB.RemTpData(utils.EmptyString, tpRoutes[0].TPid, nil); err != nil {
		t.Error(err)
	} else if _, err := storDB.GetTPRoutes(tpRoutes[0].TPid, utils.EmptyString, utils.EmptyString); err != utils.ErrNotFound {
		t.Error(err)
	}
}

func TestIDBTPAttributes(t *testing.T) {
	storDB := NewInternalDB(nil, nil, false, config.CgrConfig().StorDbCfg().Items)
	//READ
	if _, err := storDB.GetTPAttributes("TP_ID", utils.EmptyString, utils.EmptyString); err != utils.ErrNotFound {
		t.Error(err)
	}

	//WRITE
	tpAProfile := []*utils.TPAttributeProfile{
		{
			TPid:   "TP_ID",
			Tenant: "cgrates.org",
			ID:     "APROFILE_ID1",
			Attributes: []*utils.TPAttribute{
				{
					Type:      utils.MetaString,
					Path:      utils.MetaReq + utils.NestingSep + utils.AccountField + utils.InInFieldSep,
					Value:     "101",
					FilterIDs: []string{"*string:~*req.Account:101"},
				},
				{
					Type:      utils.MetaString,
					Path:      utils.MetaReq + utils.NestingSep + utils.AccountField + utils.InInFieldSep,
					Value:     "108",
					FilterIDs: []string{"*string:~*req.Account:102"},
				},
			},
		},
		{
			TPid:   "TP_ID",
			Tenant: "cgrates.org",
			ID:     "APROFILE_ID2",
			Attributes: []*utils.TPAttribute{
				{
					Type:      utils.MetaString,
					Path:      utils.MetaReq + utils.NestingSep + utils.Destination + utils.InInFieldSep,
					Value:     "12",
					FilterIDs: []string{"*string:~*req.Destination:11"},
				},
				{
					Type:      utils.MetaString,
					Path:      utils.MetaReq + utils.NestingSep + utils.Destination + utils.InInFieldSep,
					Value:     "13",
					FilterIDs: []string{"*string:~*req.Destination:10"},
				},
			},
		},
	}
	if err := storDB.SetTPAttributes(tpAProfile); err != nil {
		t.Errorf("Unable to set TPAttributeProfile:%s", err)
	}

	//READ
	if rcv, err := storDB.GetTPAttributes(tpAProfile[0].TPid, utils.EmptyString, utils.EmptyString); err != nil {
		t.Error(err)
	} else if sort.Slice(rcv, func(i, j int) bool {
		return rcv[i].ID < rcv[j].ID
	}); !slices.Equal(rcv, tpAProfile) {
		t.Errorf("Expected %v, Received %v", utils.ToJSON(rcv), utils.ToJSON(tpAProfile))
	}

	//UPDATE
	tpAProfile[0].Attributes[0].Value = "107"
	if err := storDB.SetTPAttributes(tpAProfile); err != nil {
		t.Error(err)
	}

	//READ
	if rcv, err := storDB.GetTPAttributes(tpAProfile[0].TPid, utils.EmptyString, utils.EmptyString); err != nil {
		t.Error(err)
	} else if sort.Slice(rcv, func(i, j int) bool {
		return rcv[i].ID < rcv[j].ID
	}); tpAProfile[0].Attributes[0].Value != rcv[0].Attributes[0].Value {
		t.Errorf("Expected %v, Received %v", utils.ToJSON(rcv[0]), utils.ToJSON(tpAProfile[0]))
	}

	//REMOVE and READ
	if err := storDB.RemTpData(utils.EmptyString, tpAProfile[0].TPid, nil); err != nil {
		t.Error(err)
	} else if _, err := storDB.GetTPAttributes(tpAProfile[0].TPid, utils.EmptyString, utils.EmptyString); err != utils.ErrNotFound {
		t.Error(err)
	}
}

func TestIDBRemTpData(t *testing.T) {
	storDB := NewInternalDB(nil, nil, false, config.CgrConfig().StorDbCfg().Items)
	tpAccActions := []*utils.TPAccountActions{
		{
			TPid:          "TP1",
			LoadId:        "ID",
			Tenant:        "cgrates.org",
			Account:       "1001",
			ActionPlanId:  "PREPAID_10",
			AllowNegative: true,
			Disabled:      false,
		},
	}
	if err := storDB.SetTPAccountActions(tpAccActions); err != nil {
		t.Error(err)
	}
	tpRatingProfiles := []*utils.TPRatingProfile{
		{
			TPid:     "TP1",
			LoadId:   "TEST_LOADID",
			Tenant:   "cgrates.org",
			Category: "call",
			Subject:  "*any",
		},
	}
	if err := storDB.SetTPRatingProfiles(tpRatingProfiles); err != nil {
		t.Error(err)
	}

	if err := storDB.RemTpData(utils.TBLTPAccountActions, tpAccActions[0].TPid, map[string]string{"tenant": "cgrates.org"}); err != nil {
		t.Error(err)
	}

	if err := storDB.RemTpData(utils.TBLTPRatingProfiles, tpRatingProfiles[0].TPid, map[string]string{"category": "call"}); err != nil {
		t.Error(err)
	}
}

func TestIDBTpSharedGroups(t *testing.T) {
	storDB := NewInternalDB(nil, nil, false, config.CgrConfig().StorDbCfg().Items)
	// READ
	if _, err := storDB.GetTPSharedGroups("TP1", ""); err != utils.ErrNotFound {
		t.Error(err)
	}
	// WRITE
	var snd = []*utils.TPSharedGroups{
		{
			TPid: "TP1",
			ID:   "1",
			SharedGroups: []*utils.TPSharedGroup{
				{
					Account:       "test",
					Strategy:      "*lowest_cost",
					RatingSubject: "test",
				},
			},
		},
		{
			TPid: "TP1",
			ID:   "2",
			SharedGroups: []*utils.TPSharedGroup{
				{
					Account:       "test",
					Strategy:      "*lowest_cost",
					RatingSubject: "test",
				},
			},
		},
	}
	if err := storDB.SetTPSharedGroups(snd); err != nil {
		t.Error(err)
	}
	// READ
	if rcv, err := storDB.GetTPSharedGroups("TP1", ""); err != nil {
		t.Error(err)
	} else if sort.Slice(rcv, func(i, j int) bool {
		return rcv[i].ID < rcv[j].ID
	}); !slices.Equal(rcv, snd) {
		t.Errorf("Expected %v, Received %v", utils.ToJSON(rcv), utils.ToJSON(snd))
	}
	// UPDATE
	snd[0].SharedGroups[0].Strategy = "*highest_cost"

	if err := storDB.SetTPSharedGroups(snd); err != nil {
		t.Error(err)
	}
	// READ
	if rcv, err := storDB.GetTPSharedGroups("TP1", ""); err != nil {
		t.Error(err)
	} else if sort.Slice(rcv, func(i, j int) bool {
		return rcv[i].ID < rcv[j].ID
	}); snd[0].SharedGroups[0].Strategy != rcv[0].SharedGroups[0].Strategy {
		t.Errorf("Expected %v, Received %v", utils.ToJSON(rcv[0]), utils.ToJSON(snd[0]))
	}
	// REMOVE
	if err := storDB.RemTpData("", "TP1", nil); err != nil {
		t.Error(err)
	}
	// READ
	if _, err := storDB.GetTPSharedGroups("TP1", ""); err != utils.ErrNotFound {
		t.Error(err)
	}
}
