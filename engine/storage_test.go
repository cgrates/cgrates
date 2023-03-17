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
	"fmt"
	"reflect"
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
	dest, err := dm.GetDestination("NAT", true, utils.NonTransactional)
	precision := dest.containsPrefix("0723")
	if err != nil || precision != 4 {
		t.Error("Error finding prefix: ", err, precision)
	}
}

func TestStorageDestinationContainsPrefixLong(t *testing.T) {
	dest, err := dm.GetDestination("NAT", true, utils.NonTransactional)
	precision := dest.containsPrefix("0723045326")
	if err != nil || precision != 4 {
		t.Error("Error finding prefix: ", err, precision)
	}
}

func TestStorageDestinationContainsPrefixNotExisting(t *testing.T) {
	dest, err := dm.GetDestination("NAT", true, utils.NonTransactional)
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
	err := dm.LoadDataDBCache(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
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
			utils.SMS:  {&Balance{Value: 14, ExpirationDate: zeroTime}},
			utils.DATA: {&Balance{Value: 1024, ExpirationDate: zeroTime}},
			utils.VOICE: {&Balance{Weight: 20, DestinationIDs: utils.NewStringMap("NAT")},
				&Balance{Weight: 10, DestinationIDs: utils.NewStringMap("RET")}}},
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

func TestIDBGetTpSuppliers(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	defer func() {
		cfg2, _ := config.NewDefaultCGRConfig()
		config.SetCgrConfig(cfg2)
	}()
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	suppliers := []*utils.TPSupplierProfile{
		{
			TPid:      "TP1",
			Tenant:    "cgrates.org",
			ID:        "SUPL_1",
			FilterIDs: []string{"FLTR_ACNT_dan", "FLTR_DST_DE"},
			ActivationInterval: &utils.TPActivationInterval{
				ActivationTime: "2014-07-29T15:00:00Z",
				ExpiryTime:     "",
			},
			Sorting:           "*lowest_cost",
			SortingParameters: []string{},
			Suppliers: []*utils.TPSupplier{
				{
					ID:                 "supplier1",
					FilterIDs:          []string{"FLTR_1"},
					AccountIDs:         []string{"Acc1", "Acc2"},
					RatingPlanIDs:      []string{"RPL_1"},
					ResourceIDs:        []string{"ResGroup1"},
					StatIDs:            []string{"Stat1"},
					Weight:             10,
					Blocker:            false,
					SupplierParameters: "SortingParam1",
				},
			},
			Weight: 20,
		},
	}
	if err := db.SetTPSuppliers(suppliers); err != nil {
		t.Error(err)
	}
	if _, err := db.GetTPSuppliers("TP1", "cgrates.org", "SUPL_1"); err != nil {
		t.Error(err)
	}
}

func TestTPDispatcherHosts(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, false, cfg.StorDbCfg().Items)
	dpp := []*utils.TPDispatcherHost{
		{
			TPid:   "TP1",
			Tenant: "cgrates.org",
			ID:     "ALL1",
			Conns: []*utils.TPDispatcherHostConn{
				{
					Address:   "127.0.0.1:2012",
					Transport: utils.MetaJSON,
					TLS:       true,
				},
				{
					Address:   "127.0.0.1:3012",
					Transport: utils.MetaJSON,
					TLS:       false,
				},
			}},
	}
	if err := db.SetTPDispatcherHosts(dpp); err != nil {
		t.Error(err)
	}
	if _, err := db.GetTPDispatcherHosts("TP1", "cgrates.org", "ALL1"); err != nil {
		t.Error(err)
	}
}

func TestTPThresholds(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, false, cfg.StorDbCfg().Items)
	thresholds := []*utils.TPThresholdProfile{
		{
			TPid:      "TH1",
			Tenant:    "cgrates.org",
			ID:        "Threshold",
			FilterIDs: []string{"FLTR_1", "FLTR_2"},
			ActivationInterval: &utils.TPActivationInterval{
				ActivationTime: "2014-07-29T15:00:00Z",
				ExpiryTime:     "",
			},
			MaxHits:   -1,
			MinSleep:  "1s",
			Blocker:   true,
			Weight:    10,
			ActionIDs: []string{"Thresh1"},
			Async:     true,
		},
	}
	if err := db.SetTPThresholds(thresholds); err != nil {
		t.Error(err)
	}
	if thds, err := db.GetTPThresholds("TH1", "cgrates.org", "Threshold"); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(thresholds, thds) {
		t.Errorf("Expected %v,Received %v", utils.ToJSON(thresholds), utils.ToJSON(thds))
	}
}

func TestTPFilters(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, false, cfg.StorDbCfg().Items)
	tpFltr := []*utils.TPFilterProfile{
		{
			TPid:   "TP1",
			Tenant: "cgrates.org",
			ID:     "FLT_1",
			Filters: []*utils.TPFilter{
				{
					Element: "Account",
					Type:    utils.MetaString,
					Values:  []string{"1001", "1002"},
				},
				{
					Type:    utils.MetaGreaterOrEqual,
					Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Weight,
					Values:  []string{"15.0"},
				},
			},
		},
		{
			TPid:   "TP1",
			Tenant: "cgrates.org",
			ID:     "FLT_2",
			Filters: []*utils.TPFilter{
				{
					Type:    utils.MetaPrefix,
					Element: "~*req.Cost",
					Values:  []string{"10", "15", "210"},
				},
			},
		},
	}
	if err := db.SetTPFilters(tpFltr); err != nil {
		t.Error(err)
	}
	for i := range tpFltr {
		if fltr, err := db.GetTPFilters("TP1", "cgrates.org", fmt.Sprintf("FLT_%d", i+1)); err != nil {
			t.Error(err)
		} else if !reflect.DeepEqual(fltr, tpFltr[i:i+1]) {
			t.Errorf("Expected %v,Received %v", utils.ToJSON(tpFltr[i:i+1]), utils.ToJSON(fltr))
		}
	}
}

func TestTPAttributes(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, false, cfg.StorDbCfg().Items)
	tpAttr := []*utils.TPAttributeProfile{
		{
			TPid:   "TP1",
			Tenant: "cgrates.org",
			ID:     "Attr1",

			ActivationInterval: &utils.TPActivationInterval{
				ActivationTime: "2019-07-29T15:00:00Z",
				ExpiryTime:     "",
			},
			Contexts: []string{"con1"},
			Attributes: []*utils.TPAttribute{
				{
					Path:      utils.MetaReq + utils.NestingSep + "FL1",
					Value:     "Al1",
					FilterIDs: []string{},
				},
			},
			Weight: 20,
		},
		{
			TPid:     "TP1",
			Tenant:   "cgrates.org",
			ID:       "Attr2",
			Contexts: []string{"con1"},
			ActivationInterval: &utils.TPActivationInterval{
				ActivationTime: "2019-07-14T14:35:00Z",
				ExpiryTime:     "",
			},
			Attributes: []*utils.TPAttribute{
				{
					Path:  utils.MetaReq + utils.NestingSep + "FL1",
					Value: "Al1",
				},
			},
			Weight: 20},
	}
	if err := db.SetTPAttributes(tpAttr); err != nil {
		t.Error(err)
	}
	for i := range tpAttr {
		if attr, err := db.GetTPAttributes("TP1", "cgrates.org", fmt.Sprintf("Attr%d", i+1)); err != nil {
			t.Error(err)
		} else if !reflect.DeepEqual(attr, tpAttr[i:i+1]) {
			t.Errorf("Expected %v,Received %v", utils.ToJSON(tpAttr[i:i+1]), utils.ToJSON(attr))
		}
	}
}

func TestTPChargers(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, false, cfg.StorDbCfg().Items)
	tpChrg := []*utils.TPChargerProfile{
		{
			TPid:   "TP1",
			Tenant: "cgrates.org",
			ID:     "Charger1",
			RunID:  "*rated",
			ActivationInterval: &utils.TPActivationInterval{
				ActivationTime: "2022-07-14T14:35:00Z",
				ExpiryTime:     "",
			},
			Weight: 20},
		{
			TPid:   "TP1",
			Tenant: "cgrates.org",
			ID:     "Charger2",
			RunID:  "*prepaid",
			ActivationInterval: &utils.TPActivationInterval{
				ActivationTime: "2022-08-14T14:35:00Z",
				ExpiryTime:     "",
			},
			Weight: 20,
		},
	}
	if err := db.SetTPChargers(tpChrg); err != nil {
		t.Error(err)
	}
	for i := range tpChrg {
		if cpps, err := db.GetTPChargers("TP1", "cgrates.org", fmt.Sprintf("Charger%d", i+1)); err != nil {
			t.Error(err)
		} else if !reflect.DeepEqual(cpps, tpChrg[i:i+1]) {
			t.Errorf("Expected %v,Received %v", utils.ToJSON(tpChrg[i:i+1]), utils.ToJSON(cpps))
		}
	}
}

func TestTPDispatcher(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, false, cfg.StorDbCfg().Items)
	tpDsp := []*utils.TPDispatcherProfile{
		{
			TPid:      "TP1",
			Tenant:    "cgrates.org",
			ID:        "Dsp1",
			FilterIDs: []string{"*string:Account:1002"},
			ActivationInterval: &utils.TPActivationInterval{
				ActivationTime: "2021-07-29T15:00:00Z",
				ExpiryTime:     "",
			},
			Strategy: utils.MetaFirst,
			Weight:   10,
		}, {

			TPid:       "TP1",
			Tenant:     "cgrates.org",
			ID:         "Dsp2",
			Subsystems: []string{"*any"},
			FilterIDs:  []string{},
			Strategy:   utils.MetaFirst,
			ActivationInterval: &utils.TPActivationInterval{
				ActivationTime: "2022-07-14T14:35:00Z",
				ExpiryTime:     "",
			},
			StrategyParams: []interface{}{},
			Weight:         20,
		},
	}
	if err := db.SetTPDispatcherProfiles(tpDsp); err != nil {
		t.Error(err)
	}
	for i := range tpDsp {
		if dpp, err := db.GetTPDispatcherProfiles("TP1", "cgrates.org", fmt.Sprintf("Dsp%d", i+1)); err != nil {
			t.Error(err)
		} else if !reflect.DeepEqual(dpp, tpDsp[i:i+1]) {
			t.Errorf("Expected %v,Received %v", utils.ToJSON(tpDsp[i:i+1]), utils.ToJSON(dpp))
		}
	}
}
