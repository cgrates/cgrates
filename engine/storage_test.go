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
	idb, err := NewInternalDB(nil, nil, true, nil, map[string]*config.ItemOpt{
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
	if err != nil {
		t.Fatal(err)
	}
	idb.db.Set("chID", "itmID", true, []string{utils.EmptyString}, true, "trID")
	idb.db.Set("chID2", "itmIDv", true, []string{"grpID"}, true, "trID")

	if err := idb.RemoveIndexesDrv("chID", utils.EmptyString); err != nil {
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
	idb, err := NewInternalDB(nil, nil, true, nil, map[string]*config.ItemOpt{
		utils.CacheDispatcherHosts: {
			Limit:  2,
			Remote: true,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
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
	idb, err := NewInternalDB(nil, nil, true, nil, map[string]*config.ItemOpt{
		utils.CacheDispatcherHosts: {
			Limit:  2,
			Remote: true,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
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
	idb, err := NewInternalDB(nil, nil, true, nil, map[string]*config.ItemOpt{
		utils.CacheStatQueues: {
			Limit:     4,
			StaticTTL: true,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
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
	db, dErr := NewInternalDB(nil, nil, true, nil, cfg.DataDbCfg().Items)
	if dErr != nil {
		t.Error(dErr)
	}
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
	db, dErr := NewInternalDB(nil, nil, true, nil, cfg.DataDbCfg().Items)
	if dErr != nil {
		t.Error(dErr)
	}
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
	storDB, err := NewInternalDB(nil, nil, false, nil, cfg.StorDbCfg().Items)
	if err != nil {
		t.Error(err)
	}
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
	}); !reflect.DeepEqual(snd, rcv) {
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
	storDB, err := NewInternalDB(nil, nil, false, nil, cfg.StorDbCfg().Items)
	if err != nil {
		t.Error(err)
	}

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

	storDB, err := NewInternalDB(nil, nil, false, nil, config.CgrConfig().StorDbCfg().Items)
	if err != nil {
		t.Error(err)
	}
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
	}); !reflect.DeepEqual(rcv, tpThresholds) {
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
	storDB, err := NewInternalDB(nil, nil, false, nil, config.CgrConfig().StorDbCfg().Items)
	if err != nil {
		t.Error(err)
	}
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
	}); !reflect.DeepEqual(rcv, tpFilters) {
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
	storDB, err := NewInternalDB(nil, nil, false, nil, config.CgrConfig().StorDbCfg().Items)
	if err != nil {
		t.Error(err)
	}
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
	}); !reflect.DeepEqual(rcv, tpRoutes) {
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
	storDB, err := NewInternalDB(nil, nil, false, nil, config.CgrConfig().StorDbCfg().Items)
	if err != nil {
		t.Error(err)
	}
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
	}); !reflect.DeepEqual(rcv, tpAProfile) {
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
	storDB, err := NewInternalDB(nil, nil, false, nil, config.CgrConfig().StorDbCfg().Items)
	if err != nil {
		t.Error(err)
	}
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
	storDB, err := NewInternalDB(nil, nil, false, nil, config.CgrConfig().StorDbCfg().Items)
	if err != nil {
		t.Error(err)
	}
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
	}); !reflect.DeepEqual(rcv, snd) {
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

func TestIDBGetTpIdsEmptyCol(t *testing.T) {
	storDB, err := NewInternalDB(nil, nil, false, nil, config.CgrConfig().StorDbCfg().Items)
	if err != nil {
		t.Error(err)
	}
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
					StatIDs:         []string{"Stat1"},
					Weight:          10,
					Blocker:         false,
					RouteParameters: "SortingParam1",
				},
			},
			Weight: 20,
		}}

	if err := storDB.SetTPRoutes(tpRoutes); err != nil {
		t.Errorf("Unable to set TPRoutes")
	}

	tpFilters := []*utils.TPFilterProfile{
		{
			TPid:   "TP2",
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
	}
	if err := storDB.SetTPFilters(tpFilters); err != nil {
		t.Errorf("Unable to set TPFilters")
	}
	etpIds := []string{"TP1", "TP2"}
	if tpIds, err := storDB.GetTpIds(utils.EmptyString); err != nil {
		t.Error(err)
	} else if sort.Slice(tpIds, func(i, j int) bool {
		return tpIds[i] < tpIds[j]
	}); !slices.Equal(tpIds, etpIds) {
		t.Errorf("Expected  %v,Received %v", tpIds, etpIds)
	}
}

func TestIDBGetTpTableIds(t *testing.T) {
	storDB, err := NewInternalDB(nil, nil, false, nil, config.CgrConfig().StorDbCfg().Items)
	if err != nil {
		t.Error(err)
	}
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
		{
			TPid:             "TP1",
			LoadId:           "TEST_LOADID",
			Tenant:           "cgrates.org",
			Account:          "1002",
			ActionPlanId:     "PACKAGE_10_SHARED_A_5",
			ActionTriggersId: "STANDARD_TRIGGERS",
			AllowNegative:    true,
			Disabled:         true,
		},
	}
	if err := storDB.SetTPAccountActions(tpAccActions); err != nil {
		t.Error(err)
	}
	expIds := []string{"ID", "TEST_LOADID"}
	if tpIds, err := storDB.GetTpTableIds("TP1", utils.TBLTPAccountActions, utils.TPDistinctIds{utils.IDCfg}, nil, &utils.PaginatorWithSearch{}); err != nil {
		t.Error(err)
	} else if sort.Slice(tpIds, func(i, j int) bool {
		return tpIds[i] < tpIds[j]
	}); !slices.Equal(tpIds, expIds) {
		t.Errorf("Expected  %v,Received %v", tpIds, expIds)
	}
}

func TestIDBGetTPDestinationRatesPaginator(t *testing.T) {
	storDB, err := NewInternalDB(nil, nil, true, nil, config.CgrConfig().StorDbCfg().Items)
	if err != nil {
		t.Error(err)
	}
	destRates := []*utils.TPDestinationRate{
		{
			TPid: "TEST_TPID",
			ID:   "TEST_DSTRATE",
			DestinationRates: []*utils.DestinationRate{
				{
					DestinationId:    "TEST_DEST1",
					RateId:           "TEST_RATE1",
					RoundingMethod:   "*up",
					RoundingDecimals: 4},
				{
					DestinationId:    "TEST_DEST2",
					RateId:           "TEST_RATE2",
					RoundingMethod:   "*up",
					RoundingDecimals: 4},
			},
		},
		{
			TPid: "TEST_TPID",
			ID:   "RT_STD_WEEKEND",
			DestinationRates: []*utils.DestinationRate{
				{
					DestinationId:    "GERMANY",
					RateId:           "R2",
					Rate:             csvr.rates["R2"],
					RoundingMethod:   utils.MetaRoundingMiddle,
					RoundingDecimals: 4,
				},
				{
					DestinationId:    "GERMANY_O2",
					RateId:           "R3",
					Rate:             csvr.rates["R3"],
					RoundingMethod:   utils.MetaRoundingMiddle,
					RoundingDecimals: 4,
				},
			},
		},
	}
	if err := storDB.SetTPDestinationRates(destRates); err != nil {
		t.Error(err)
	}
	if dstRates, err := storDB.GetTPDestinationRates(destRates[0].TPid, "TEST", &utils.Paginator{Limit: utils.IntPointer(1)}); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(destRates[0], dstRates[0]) {
		t.Errorf("Expected %v,\nReceived %v", utils.ToJSON(destRates[0]), utils.ToJSON(dstRates[0]))
	}
}

func TestIDBGetTPRatingPlans(t *testing.T) {
	storDB, err := NewInternalDB(nil, nil, true, nil, config.CgrConfig().StorDbCfg().Items)
	if err != nil {
		t.Error(err)
	}
	ratingPlans := []*utils.TPRatingPlan{
		{
			TPid: "TP1",
			ID:   "Plan1",
			RatingPlanBindings: []*utils.TPRatingPlanBinding{
				{
					DestinationRatesId: "DR_FREESWITCH_USERS",
					TimingId:           "ALWAYS",
					Weight:             10,
				},
			},
		},
		{
			TPid: "TP1",
			ID:   "Plan2",
			RatingPlanBindings: []*utils.TPRatingPlanBinding{
				{
					DestinationRatesId: "1",
					TimingId:           "ALWAYS",
					Weight:             0.0,
				},
			},
		},
		{
			TPid: "TP1",
			ID:   "Plan3",
			RatingPlanBindings: []*utils.TPRatingPlanBinding{
				{
					DestinationRatesId: "2",
					TimingId:           "ALWAYS",
					Weight:             2,
				},
			},
		},
	}
	if err := storDB.SetTPRatingPlans(ratingPlans); err != nil {
		t.Error(err)
	}

	if rPlans, err := storDB.GetTPRatingPlans("TP1", "Plan", &utils.Paginator{Limit: utils.IntPointer(1), Offset: utils.IntPointer(1)}); err != nil {
		t.Error(err)
	} else if len(rPlans) != 1 {
		t.Errorf("Expected slice length: 1 ,Received : %v", len(rPlans))
	}
}

func TestIDBRemoveSMCost(t *testing.T) {
	storDB, err := NewInternalDB(nil, nil, true, nil, config.CgrConfig().StorDbCfg().Items)
	if err != nil {
		t.Error(err)
	}
	// READ
	if _, err := storDB.GetSMCosts("", "", "", ""); err != utils.ErrNotFound {
		t.Error(err)
	}
	// WRITE
	var snd = []*SMCost{
		{
			CGRID:       "88ed9c38005f07576a1e1af293063833b60edcc6",
			RunID:       "1",
			OriginHost:  "host2",
			OriginID:    "2",
			CostDetails: NewBareEventCost(),
		},
		{
			CGRID:       "88ed9c38005f07576a1e1af293063833b60edcc2",
			RunID:       "2",
			OriginHost:  "host2",
			OriginID:    "2",
			CostDetails: NewBareEventCost(),
		},
	}
	for _, smc := range snd {
		if err := storDB.SetSMCost(smc); err != nil {
			t.Error(err)
		}
	}
	// READ
	if rcv, err := storDB.GetSMCosts("", "", "host2", ""); err != nil {
		t.Error(err)
	} else if sort.Slice(rcv, func(i, j int) bool {
		return rcv[i].CGRID < rcv[j].CGRID
	}); slices.Equal(snd, rcv) {
		t.Errorf("Expected %+v,Received %+v", utils.ToJSON(snd), utils.ToJSON(rcv))
	}
	// REMOVE
	for _, smc := range snd {
		if err := storDB.RemoveSMCost(smc); err != nil {
			t.Error(err)
		}
	}
	// READ
	if _, err := storDB.GetSMCosts("", "", "", ""); err != utils.ErrNotFound {
		t.Error(err)
	}
}

func TestIDBRemoveSMC(t *testing.T) {
	storDB, err := NewInternalDB(nil, nil, false, nil, config.CgrConfig().StorDbCfg().Items)
	if err != nil {
		t.Error(err)
	}
	// READ
	if _, err := storDB.GetSMCosts("", "", "", ""); err != utils.ErrNotFound {
		t.Error(err)
	}
	// WRITE
	var snd = []*SMCost{
		{
			CGRID:       "CGRID1",
			RunID:       "11",
			OriginHost:  "host22",
			OriginID:    "O1",
			CostDetails: NewBareEventCost(),
		},
		{
			CGRID:       "CGRID2",
			RunID:       "12",
			OriginHost:  "host22",
			OriginID:    "O2",
			CostDetails: NewBareEventCost(),
		},
		{
			CGRID:       "CGRID3",
			RunID:       "13",
			OriginHost:  "host23",
			OriginID:    "O3",
			CostDetails: NewBareEventCost(),
		},
	}
	for _, smc := range snd {
		if err := storDB.SetSMCost(smc); err != nil {
			t.Error(err)
		}
	}
	// READ
	if rcv, err := storDB.GetSMCosts("", "", "host22", ""); err != nil {
		t.Fatal(err)
	} else if len(rcv) != 2 {
		t.Errorf("Expected 2 results received %v ", len(rcv))
	}
	// REMOVE
	if err := storDB.RemoveSMCosts(&utils.SMCostFilter{
		RunIDs:         []string{"12", "13"},
		NotRunIDs:      []string{"11"},
		OriginHosts:    []string{"host22", "host23"},
		NotOriginHosts: []string{"host21"},
	}); err != nil {
		t.Error(err)
	}
	// READ
	if rcv, err := storDB.GetSMCosts("", "", "", ""); err != nil {
		t.Error(err)
	} else if len(rcv) != 1 {
		t.Errorf("Expected 1 result received %v ", len(rcv))
	}
	// REMOVE
	if err := storDB.RemoveSMCosts(&utils.SMCostFilter{}); err != nil {
		t.Error(err)
	}
	// READ
	if _, err := storDB.GetSMCosts("", "", "", ""); err != utils.ErrNotFound {
		t.Error(err)
	}
}

func TestIDBVersions(t *testing.T) {
	dataDB, dErr := NewInternalDB(nil, nil, true, nil, config.CgrConfig().DataDbCfg().Items)
	if dErr != nil {
		t.Error(dErr)
	}
	if _, err := dataDB.GetVersions(utils.Accounts); err != utils.ErrNotFound {
		t.Error(err)
	}
	vrs := Versions{
		utils.Accounts:       3,
		utils.Actions:        2,
		utils.ActionTriggers: 2,
		utils.ActionPlans:    2,
		utils.SharedGroups:   2,
		utils.CostDetails:    1,
	}
	if err := dataDB.SetVersions(vrs, false); err != nil {
		t.Error(err)
	}
	if rcv, err := dataDB.GetVersions(""); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(vrs, rcv) {
		t.Errorf("Expecting: %v, received: %v", vrs, rcv)
	}
	delete(vrs, utils.SharedGroups)
	if err := dataDB.SetVersions(vrs, true); err != nil { // overwrite
		t.Error(err)
	}
	if rcv, err := dataDB.GetVersions(""); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(vrs, rcv) {
		t.Errorf("Expecting: %v, received: %v", vrs, rcv)
	}
	eAcnts := Versions{utils.Accounts: vrs[utils.Accounts]}
	if rcv, err := dataDB.GetVersions(utils.Accounts); err != nil { //query one element
		t.Error(err)
	} else if !reflect.DeepEqual(eAcnts, rcv) {
		t.Errorf("Expecting: %v, received: %v", eAcnts, rcv)
	}
	if _, err := dataDB.GetVersions(utils.NotAvailable); err != utils.ErrNotFound { //query non-existent
		t.Error(err)
	}
	eAcnts[utils.Accounts] = 2
	vrs[utils.Accounts] = eAcnts[utils.Accounts]
	if err := dataDB.SetVersions(eAcnts, false); err != nil { // change one element
		t.Error(err)
	}
	if rcv, err := dataDB.GetVersions(""); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(vrs, rcv) {
		t.Errorf("Expecting: %v, received: %v", vrs, rcv)
	}
	if err = dataDB.RemoveVersions(eAcnts); err != nil { // remove one element
		t.Error(err)
	}
	delete(vrs, utils.Accounts)
	if rcv, err := dataDB.GetVersions(""); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(vrs, rcv) {
		t.Errorf("Expecting: %v, received: %v", vrs, rcv)
	}
	if err = dataDB.RemoveVersions(nil); err != nil { // remove one element
		t.Error(err)
	}
	if _, err := dataDB.GetVersions(""); err != utils.ErrNotFound { //query non-existent
		t.Error(err)
	}
}

func TestIDBGetCDR(t *testing.T) {
	storDB, err := NewInternalDB([]string{utils.AccountField, utils.CGRID, utils.OriginID, utils.RequestType, utils.Tenant, utils.Category, utils.RunID, utils.Source, utils.ToR, utils.Subject, utils.OriginHost, "ExtraHeader1", "ExtraHeader2"}, []string{"Destination", "Header2"}, false, nil, config.CgrConfig().StorDbCfg().Items)
	if err != nil {
		t.Fatal(err)
	}
	cdr := &CDR{
		CGRID:       "CGR1",
		RunID:       utils.MetaRaw,
		OrderID:     time.Now().UnixNano(),
		OriginHost:  "127.0.0.1",
		Source:      "testSetCDRs",
		OriginID:    "testevent1",
		ToR:         utils.MetaVoice,
		RequestType: utils.MetaPrepaid,
		Tenant:      "cgrates.org",
		Category:    "call",
		Account:     "1004",
		Subject:     "1004",
		Destination: "1007",
		SetupTime:   time.Date(2015, 12, 12, 14, 52, 0, 0, time.UTC),
		AnswerTime:  time.Date(2015, 12, 12, 14, 52, 20, 0, time.UTC),
		Usage:       35 * time.Second,
		ExtraFields: map[string]string{"ExtraHeader1": "ExtraVal1", "Header2": "Val2", "ExtraHeader2": "Val"},
		Cost:        -1,
	}
	if err := storDB.SetCDR(cdr, false); err != nil {
		t.Error(err)
	}
	if cdrs, _, err := storDB.GetCDRs(&utils.CDRsFilter{CGRIDs: []string{cdr.CGRID}, OriginIDs: []string{"testevent1"}, RequestTypes: []string{utils.MetaPrepaid}, Tenants: []string{"cgrates.org"}, Categories: []string{"call"}, Subjects: []string{"1004"}, Sources: []string{"testSetCDRs"}, ToRs: []string{utils.MetaVoice}, RunIDs: []string{utils.MetaRaw}, OriginHosts: []string{"127.0.0.1"}, DestinationPrefixes: []string{"100"}, ExtraFields: map[string]string{"ExtraHeader1": "ExtraVal1", "Header2": "Val2"}, NotExtraFields: map[string]string{"ExtraHeader2": "ExtraVal2", "Header2": "Hdr"}}, false); err != nil {
		t.Error(err)
	} else if len(cdrs) != 1 {
		t.Errorf("cdr %+v, Unexpected number of CDRs returned: %d", cdr, len(cdrs))
	}
}

func TestIDBGeTps(t *testing.T) {
	storDB, err := NewInternalDB(nil, nil, false, nil, config.CgrConfig().StorDbCfg().Items)
	if err != nil {
		t.Error(err)
	}
	resources := []*utils.TPResourceProfile{
		{
			TPid:              "TP1",
			Tenant:            "cgrates.org",
			ID:                "ResGroup21",
			FilterIDs:         []string{"*string:~*req.Account:1001"},
			UsageTTL:          "1s",
			AllocationMessage: "call",
			Weight:            10,
			Limit:             "2",
			Blocker:           true,
			Stored:            true,
		},
	}
	if err := storDB.SetTPResources(resources); err != nil {
		t.Error(err)
	}
	if _, err := storDB.GetTPResources("TP1", "cgrates.org", resources[0].ID); err != nil {
		t.Error(err)
	}
	stats := []*utils.TPStatProfile{
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
	if err := storDB.SetTPStats(stats); err != nil {
		t.Error(err)
	}
	if _, err := storDB.GetTPStats("TP1", "cgrates.org", stats[0].ID); err != nil {
		t.Error(err)
	}
	thresholds := []*utils.TPThresholdProfile{
		{TPid: "TP1",
			Tenant:    "cgrates.org",
			ID:        "TH_1",
			FilterIDs: []string{"FLTR_1", "FLTR_2"},
			MaxHits:   12,
			MinHits:   10,
			MinSleep:  "1s",
			Blocker:   false,
			Weight:    20.0,
			ActionIDs: []string{"WARN3"},
		}}
	if err := storDB.SetTPThresholds(thresholds); err != nil {
		t.Error(err)
	}
	if _, err := storDB.GetTPThresholds("TP1", "cgrates.org", thresholds[0].ID); err != nil {
		t.Error(err)
	}
	filters := []*utils.TPFilterProfile{
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
	if err := storDB.SetTPFilters(filters); err != nil {
		t.Error(err)
	}
	if _, err := storDB.GetTPFilters("TP1", "cgrates.org", filters[0].ID); err != nil {
		t.Error(err)
	}
	routes := []*utils.TPRouteProfile{
		{
			TPid:      "TP1",
			Tenant:    "cgrates.org",
			ID:        "SUPL_1",
			FilterIDs: []string{"*string:~*req.Accout:1007"},
			Sorting:   "*lowest_cost",
			Routes: []*utils.TPRoute{
				{
					ID:              "supplier1",
					StatIDs:         []string{"Stat1"},
					Weight:          10,
					Blocker:         false,
					RouteParameters: "SortingParam1",
				},
			},
			Weight: 20,
		},
	}
	if err := storDB.SetTPRoutes(routes); err != nil {
		t.Error(err)
	}
	if _, err := storDB.GetTPRoutes("TP1", "cgrates.org", routes[0].ID); err != nil {
		t.Error(err)
	}
	attributes := []*utils.TPAttributeProfile{
		{TPid: "TP1",
			Tenant:    "cgrates.org",
			ID:        "ALS1",
			Contexts:  []string{"con1"},
			FilterIDs: []string{"FLTR_ACNT_dan", "FLTR_DST_DE"},
			Attributes: []*utils.TPAttribute{
				{
					Path:  utils.MetaReq + utils.NestingSep + "FL1",
					Value: "Al1",
				},
			},
			Weight: 20,
		}}
	if err := storDB.SetTPAttributes(attributes); err != nil {
		t.Error(err)
	}
	if _, err := storDB.GetTPAttributes("TP1", "cgrates.org", attributes[0].ID); err != nil {
		t.Error(err)
	}
	chargers := []*utils.TPChargerProfile{
		{TPid: "TP1",
			Tenant:       "cgrates.org",
			ID:           "Chrgs",
			FilterIDs:    []string{"FLTR_ACNT_dan", "FLTR_DST_DE"},
			RunID:        utils.MetaDefault,
			AttributeIDs: []string{"Attr1", "Attr2"},
			Weight:       20},
	}
	if err := storDB.SetTPChargers(chargers); err != nil {
		t.Error(err)
	}
	if _, err := storDB.GetTPChargers("TP1", "cgrates.org", chargers[0].ID); err != nil {
		t.Error(err)
	}

	dispatcherProfiles := []*utils.TPDispatcherProfile{
		{
			TPid:      "TP1",
			Tenant:    "cgrates.org",
			ID:        "Dsp1",
			FilterIDs: []string{"*string:~*req.Account:1002"},
			ActivationInterval: &utils.TPActivationInterval{
				ActivationTime: "2014-07-29T15:00:00Z",
				ExpiryTime:     "",
			},
			Strategy: utils.MetaFirst,
			Weight:   10,
		},
	}
	if err := storDB.SetTPDispatcherProfiles(dispatcherProfiles); err != nil {
		t.Error(err)
	}
	if _, err := storDB.GetTPDispatcherProfiles("TP1", "cgrates.org", dispatcherProfiles[0].ID); err != nil {
		t.Error(err)
	}
	dispatcherHosts := []*utils.TPDispatcherHost{

		{
			TPid:   "TP1",
			Tenant: "cgrates.org",
			ID:     "ALL",
			Conn: &utils.TPDispatcherHostConn{
				Address:   "127.0.0.1:2012",
				Transport: utils.MetaJSON,
				TLS:       true,
			},
		},
	}
	if err := storDB.SetTPDispatcherHosts(dispatcherHosts); err != nil {
		t.Error(err)
	}
	if _, err := storDB.GetTPDispatcherHosts("TP1", "cgrates.org", dispatcherHosts[0].ID); err != nil {
		t.Error(err)
	}

}

func TestIDBGetAllActionPlanDrv(t *testing.T) {
	dataDB, dErr := NewInternalDB(nil, nil, true, nil, config.CgrConfig().DataDbCfg().Items)
	if dErr != nil {
		t.Error(dErr)
	}
	acPln := []struct {
		key string
		apl *ActionPlan
	}{
		{
			key: " MORE_MINUTES",
			apl: &ActionPlan{
				Id:         "MORE_MINUTES",
				AccountIDs: utils.StringMap{"1001": true},
			},
		},
		{
			key: "PACKAGE_10_SHARED_A_5",
			apl: &ActionPlan{
				Id: "PACKAGE_10_SHARED_A_5",
				AccountIDs: utils.StringMap{
					"cgrates.org:1001": true,
				},
			},
		},
	}
	for _, tt := range acPln {
		t.Run(tt.key, func(t *testing.T) {
			if err := dataDB.SetActionPlanDrv(tt.key, tt.apl); err != nil {
				t.Error(err)
			}
		})
	}

	if apl, err := dataDB.GetAllActionPlansDrv(); err != nil {
		t.Error(err)
	} else if len(apl) != 2 {
		t.Errorf("Expected : 2,Received %v", len(apl))
	}
}

func TestComposeURI(t *testing.T) {
	tests := []struct {
		name     string
		scheme   string
		host     string
		port     string
		db       string
		user     string
		pass     string
		expected string
		parseErr bool
	}{
		{
			name:     "multiple nodes",
			scheme:   "mongodb",
			host:     "clusternode1:1230,clusternode2:1231,clusternode3",
			port:     "1232",
			db:       "cgrates",
			user:     "user",
			pass:     "pass",
			expected: "mongodb://user:pass@clusternode1:1230,clusternode2:1231,clusternode3:1232/cgrates",
		},
		{
			name:     "no port",
			scheme:   "mongodb",
			host:     "localhost:1234",
			port:     "0",
			db:       "cgrates",
			user:     "user",
			pass:     "pass",
			expected: "mongodb://user:pass@localhost:1234/cgrates",
		},
		{
			name:     "with port",
			scheme:   "mongodb",
			host:     "localhost",
			port:     "1234",
			db:       "cgrates",
			user:     "user",
			pass:     "pass",
			expected: "mongodb://user:pass@localhost:1234/cgrates",
		},
		{
			name:     "no password",
			scheme:   "mongodb",
			host:     "localhost",
			port:     "1234",
			db:       "cgrates",
			user:     "user",
			pass:     "",
			expected: "mongodb://localhost:1234/cgrates",
		},
		{
			name:     "no db",
			scheme:   "mongodb",
			host:     "localhost",
			port:     "1234",
			db:       "",
			user:     "user",
			pass:     "pass",
			expected: "mongodb://user:pass@localhost:1234",
		},
		{
			name:     "different scheme",
			scheme:   "mongodb+srv",
			host:     "cgr.abcdef.mongodb.net",
			port:     "0",
			db:       "?retryWrites=true&w=majority",
			user:     "user",
			pass:     "pass",
			expected: "mongodb+srv://user:pass@cgr.abcdef.mongodb.net/?retryWrites=true&w=majority",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := composeMongoURI(tt.scheme, tt.host, tt.port, tt.db, tt.user, tt.pass)
			if url != tt.expected {
				t.Errorf("expected %v,\nreceived %v", tt.expected, url)
			}
		})
	}
}
