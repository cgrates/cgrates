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
	"github.com/cgrates/rpcclient"
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

func TestTpRLoadRatingProfilesFiltered(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	tpr, err := NewTpReader(db, db, "TP1", "UTC", nil, nil)
	if err != nil {
		t.Error(err)
	}
	rpl := &RatingPlan{

		Id: "TEST_RPLAN1",
		Timings: map[string]*RITiming{
			"TimingsId1": {
				Years:     utils.Years{},
				Months:    utils.Months{},
				MonthDays: utils.MonthDays{1, 2, 3, 4, 5},
			},
		},
		Ratings: map[string]*RIRate{
			"RateId": {
				ConnectFee: 0.0,
				Rates: RateGroups{
					&Rate{
						GroupIntervalStart: 0,
						Value:              0.3,
						RateIncrement:      15 * time.Second,
						RateUnit:           60 * time.Second,
					},
				},
				RoundingMethod:   utils.ROUNDING_MIDDLE,
				RoundingDecimals: 4,
			},
		},
		DestinationRates: map[string]RPRateList{
			"*any": []*RPRate{
				{
					Timing: "TimingsId1",
					Rating: "RateId",
					Weight: 10,
				},
			},
		}}

	db.SetRatingPlanDrv(rpl)
	qriedRpf := &utils.TPRatingProfile{
		TPid:     "TP1",
		LoadId:   "TEST_LOADID",
		Tenant:   "cgrates.org",
		Category: "call",
		Subject:  "1001",
		RatingPlanActivations: []*utils.TPRatingActivation{
			{
				ActivationTime:   "2014-01-14T00:00:00Z",
				RatingPlanId:     "TEST_RPLAN1",
				FallbackSubjects: "subj1;subj2"},
		},
	}
	tpRpf := []*utils.TPRatingProfile{
		qriedRpf,
	}
	db.SetTPRatingProfiles(tpRpf)
	exp := []string{utils.ConcatenatedKey(utils.META_OUT, "cgrates.org", "call", "1001")}
	if rpl, err := tpr.LoadRatingProfilesFiltered(qriedRpf); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(exp, rpl) {
		t.Errorf("Expected %v,Received %v", utils.ToJSON(exp), utils.ToJSON(rpl))
	}

}

func TestTprLoadRatingPlansFiltered(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	tpr, err := NewTpReader(db, db, "TP1", "UTC", nil, nil)
	if err != nil {
		t.Error(err)
	}
	dests := []*utils.TPDestination{
		{
			TPid: "TP1",
			ID:   "DEST1",
			Prefixes: []string{
				"+20", "+232",
			},
		},
	}
	db.SetTPDestinations(dests)
	rates := []*utils.TPRate{
		{
			TPid: "TP1",
			ID:   "RATE1",
			RateSlots: []*utils.RateSlot{
				{
					ConnectFee:         0.100,
					Rate:               0.200,
					RateUnit:           "60",
					RateIncrement:      "60",
					GroupIntervalStart: "0"},
				{
					ConnectFee:         0.0,
					Rate:               0.1,
					RateUnit:           "1",
					RateIncrement:      "60",
					GroupIntervalStart: "60"},
			},
		},
		{
			TPid: "TP1",
			ID:   "RT1",
			RateSlots: []*utils.RateSlot{
				{
					ConnectFee:         0.0,
					Rate:               0.1,
					RateUnit:           "1",
					RateIncrement:      "60",
					GroupIntervalStart: "60"},
			},
		},
	}
	db.SetTPRates(rates)
	dRates := []*utils.TPDestinationRate{
		{
			TPid: "TP1",
			ID:   "TEST_DSTRATE1",
			DestinationRates: []*utils.DestinationRate{
				{
					DestinationId:    "DEST1",
					RateId:           "RATE1",
					RoundingMethod:   "*up",
					RoundingDecimals: 4},
			},
		},
		{
			TPid: "TP1",
			ID:   "RateId",
			DestinationRates: []*utils.DestinationRate{
				{DestinationId: utils.ANY, RateId: "RT1", RoundingMethod: utils.ROUNDING_UP, RoundingDecimals: 1},
			},
		},
	}
	db.SetTPDestinationRates(dRates)
	tms := []*utils.ApierTPTiming{
		{
			TPid:      "TP1",
			ID:        "TEST_TIMING1",
			Years:     "*any",
			Months:    "*any",
			MonthDays: "*any",
			WeekDays:  "1;2;3;4;5",
			Time:      "19:00:00",
		},
	}
	db.SetTPTimings(tms)
	rPs := []*utils.TPRatingPlan{
		{
			TPid: "TP1",
			ID:   "RP_1",
			RatingPlanBindings: []*utils.TPRatingPlanBinding{
				{
					DestinationRatesId: "TEST_DSTRATE1",
					TimingId:           "TEST_TIMING1",
					Weight:             10.0},
			},
		}, {

			TPid: "TPP1",
			ID:   "RP_2",
			RatingPlanBindings: []*utils.TPRatingPlanBinding{
				{
					DestinationRatesId: "RateId",
					TimingId:           "TEST_TIMING1",
					Weight:             12,
				},
			},
		},
	}
	db.SetTPRatingPlans(rPs)
	if pass, err := tpr.LoadRatingPlansFiltered("RP"); err != nil || !pass {
		t.Error(err)
	}
}

func TestTPRLoadAccountActionsFiltered(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	tmpConn := connMgr
	defer func() {
		SetConnManager(tmpConn)
	}()
	aTriggers := []*utils.TPActionTriggers{
		{
			TPid: "TP1",
			ID:   "STANDARD_TRIGGERS",
			ActionTriggers: []*utils.TPActionTrigger{
				{
					Id:                    "STANDARD_TRIGGERS",
					UniqueID:              "1",
					ThresholdType:         "*min_balance",
					ThresholdValue:        2.0,
					Recurrent:             false,
					MinSleep:              "0",
					BalanceId:             "b1",
					BalanceType:           "*monetary",
					BalanceDestinationIds: "",
					BalanceWeight:         "0.0",
					BalanceExpirationDate: utils.UNLIMITED,
					BalanceTimingTags:     "T1",
					BalanceRatingSubject:  "special1",
					BalanceCategories:     "call",
					BalanceSharedGroups:   "",
					BalanceBlocker:        "false",
					BalanceDisabled:       "false",
					ActionsId:             "TOPUP_RST_10",
					Weight:                10},
			},
		},
	}
	db.SetTPActionTriggers(aTriggers)
	timings := []*utils.ApierTPTiming{
		{
			TPid:      "TP1",
			ID:        "ASAP",
			Years:     "*any",
			Months:    "*any",
			MonthDays: "*any",
			WeekDays:  "1;2;3;4;5",
		},
	}
	db.SetTPTimings(timings)
	acts := []*utils.TPActions{{
		TPid: "TP1",
		ID:   "TOPUP_RST_10",
		Actions: []*utils.TPAction{
			{
				Identifier:      "*topup_reset",
				BalanceType:     "*monetary",
				Units:           "5.0",
				ExpiryTime:      "*never",
				DestinationIds:  "*any",
				RatingSubject:   "special1",
				Categories:      "call",
				SharedGroups:    "GROUP1",
				BalanceWeight:   "10.0",
				ExtraParameters: "",
				Weight:          10.0},
		}},
	}
	db.SetTPActions(acts)
	aPlans := []*utils.TPActionPlan{
		{
			TPid: "TP1",
			ID:   "PACKAGE_10",
			ActionPlan: []*utils.TPActionTiming{
				{
					ActionsId: "TOPUP_RST_10",
					TimingId:  "ASAP",
					Weight:    10.0},
			},
		},
	}
	db.SetTPActionPlans(aPlans)
	qriedAA := &utils.TPAccountActions{
		TPid:             "TP1",
		LoadId:           "TEST_LOADID",
		Tenant:           "cgrates.org",
		Account:          "1001",
		ActionPlanId:     "PACKAGE_10",
		ActionTriggersId: "STANDARD_TRIGGERS",
		AllowNegative:    true,
		Disabled:         true,
	}
	db.SetTPAccountActions([]*utils.TPAccountActions{qriedAA})
	clientConn := make(chan rpcclient.ClientConnector, 1)
	clientConn <- clMock(func(serviceMethod string, _, reply interface{}) error {
		if serviceMethod == utils.CacheSv1ReloadCache {
			*reply.(*string) = utils.OK
			return nil
		}
		return utils.ErrNotImplemented
	},
	)
	connMgr := NewConnManager(cfg, map[string]chan rpcclient.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches): clientConn,
	})
	tpr, err := NewTpReader(db, db, "TP1", "UTC", []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches)}, nil)
	if err != nil {
		t.Error(err)
	}
	SetConnManager(connMgr)
	if err := tpr.LoadAccountActionsFiltered(qriedAA); err != nil {
		t.Error(err)
	}
}

func TestTprLoadDestinationsFiltered(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	tpr, err := NewTpReader(db, db, "TP1", "UTC", []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches)}, nil)
	if err != nil {
		t.Error(err)
	}
	dests := []*utils.TPDestination{
		{TPid: "TP1", ID: "DST_1002", Prefixes: []string{"1002"}},
		{TPid: "TP1", ID: "DST_1003", Prefixes: []string{"1003"}},
		{TPid: "TP1", ID: "DST_1007", Prefixes: []string{"1007"}},
	}
	db.SetTPDestinations(dests)
	if pass, err := tpr.LoadDestinationsFiltered("DST"); err != nil || !pass {
		t.Error(err)
	}
}

func TestTprRealoadSched(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	db2 := NewInternalDB(nil, nil, false, cfg.StorDbCfg().Items)
	tmpConn := connMgr
	defer func() {
		SetConnManager(tmpConn)
	}()
	tpr, err := NewTpReader(db, db2, "TPID1", "UTC", nil, []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaScheduler)})
	if err != nil {
		t.Error(err)
	}
	timings := []*utils.ApierTPTiming{
		{
			TPid:      "TPID1",
			ID:        "TEST_TIMING",
			Years:     "*any",
			Months:    "*any",
			MonthDays: "*any",
			WeekDays:  "1;2;4",
			Time:      "00:00:01",
		},
	}
	db2.SetTPTimings(timings)
	if err := tpr.LoadTimings(); err != nil {
		t.Error(err)
	}
	acP := &Action{
		Id:         "TOPUP_RST_10",
		ActionType: utils.TOPUP_RESET,
	}
	db.SetActionsDrv("TOPUP_RST_10", Actions{acP})
	aPlans := []*utils.TPActionPlan{
		{
			TPid: "TPID1",
			ID:   "PACKAGE_10",
			ActionPlan: []*utils.TPActionTiming{
				{
					ActionsId: "TOPUP_RST_10",
					TimingId:  utils.ASAP,
					Weight:    10.0},
			},
		},
	}
	db2.SetTPActionPlans(aPlans)
	if err := tpr.LoadActionPlans(); err != nil {
		t.Error(err)
	}
	clientConn := make(chan rpcclient.ClientConnector, 1)
	clientConn <- clMock(func(serviceMethod string, _, _ interface{}) error {
		if serviceMethod == utils.SchedulerSv1Reload {
			return nil
		}
		return utils.ErrNotImplemented
	})
	connMgr := NewConnManager(cfg, map[string]chan rpcclient.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaScheduler): clientConn,
	})
	SetConnManager(connMgr)
	if err := tpr.ReloadScheduler(false); err != nil {
		t.Error(err)
	}
}
func TestTprReloadCache(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	dataDb := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	storDb := NewInternalDB(nil, nil, false, cfg.StorDbCfg().Items)
	tmpConn := connMgr
	defer func() {
		SetConnManager(tmpConn)
	}()
	clientConn := make(chan rpcclient.ClientConnector, 1)
	clientConn <- clMock(func(serviceMethod string, args, _ interface{}) error {
		if serviceMethod == utils.CacheSv1LoadCache {
			return nil
		} else if serviceMethod == utils.CacheSv1Clear {
			return nil
		}
		return utils.ErrNotImplemented
	})
	connMgr := NewConnManager(cfg, map[string]chan rpcclient.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches): clientConn,
	})

	SetConnManager(connMgr)
	tpr, err := NewTpReader(dataDb, storDb, "TEST_TP", "UTC", []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches)}, nil)
	if err != nil {
		t.Error(err)
	}
	dests := []*utils.TPDestination{
		{
			TPid: "TEST_TP",
			ID:   "DEST1",
			Prefixes: []string{
				"1001", "1002",
			},
		},
		{
			TPid: "TEST_TP",
			ID:   "DEST2",
			Prefixes: []string{
				"1003", "1004",
			},
		},
	}
	storDb.SetTPDestinations(dests)
	if err := tpr.LoadDestinations(); err != nil {
		t.Error(err)
	}
	filters := []*utils.TPFilterProfile{
		{
			TPid:   "TEST_TP",
			ID:     "FLT_1",
			Tenant: "cgrates.org",
			Filters: []*utils.TPFilter{
				{
					Type:    utils.MetaPrefix,
					Element: "Account",
					Values:  []string{"1001", "1002"},
				},
			},
		},
		{
			TPid:   "TEST_TP",
			ID:     "FLT_2",
			Tenant: "cgrates.org",
			Filters: []*utils.TPFilter{
				{
					Type:    utils.MetaGreaterOrEqual,
					Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Weight,
					Values:  []string{"15.0"},
				},
			},
		},
	}
	storDb.SetTPFilters(filters)
	if err := tpr.LoadFilters(); err != nil {
		t.Error(err)
	}
	if err := tpr.ReloadCache(utils.MetaLoad, false, nil, "cgrates.org"); err == nil {
		t.Error(err)
	}
}

func TestTpRLoadAll(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	dataDb := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	storDb := NewInternalDB(nil, nil, false, cfg.StorDbCfg().Items)
	tpId := "TP1"
	tpr, err := NewTpReader(dataDb, storDb, tpId, "UTC", nil, nil)
	if err != nil {
		t.Error(err)
	}
	dests := []*utils.TPDestination{
		{
			TPid: tpId,
			ID:   "DEST",
			Prefixes: []string{
				"1001", "1002", "1003",
			},
		},
	}
	dest := &Destination{
		Id: "DEST",
		Prefixes: []string{
			"1001", "1002", "1003",
		},
	}
	rates := []*utils.TPRate{
		&utils.TPRate{
			TPid: tpId,
			ID:   "RATE1",
			RateSlots: []*utils.RateSlot{
				{ConnectFee: 12,
					Rate:               3,
					RateUnit:           "4s",
					RateIncrement:      "6s",
					GroupIntervalStart: "1s"},
			},
		}}
	destRates := []*utils.TPDestinationRate{
		{
			TPid: tpId,
			ID:   "DR_FREESWITCH_USERS",
			DestinationRates: []*utils.DestinationRate{
				{
					DestinationId:    "DEST",
					RateId:           "RATE1",
					RoundingMethod:   "*up",
					RoundingDecimals: 4},
			},
		},
	}
	timings := []*utils.ApierTPTiming{
		{
			TPid:      tpId,
			ID:        "ALWAYS",
			Years:     "*any",
			Months:    "*any",
			MonthDays: "*any",
			WeekDays:  "1;2;3;4;5",
			Time:      "19:00:00",
		},
	}

	ratingPlans := []*utils.TPRatingPlan{
		{
			TPid: tpId,
			ID:   "RP_1",
			RatingPlanBindings: []*utils.TPRatingPlanBinding{
				{
					DestinationRatesId: "DR_FREESWITCH_USERS",
					TimingId:           "ALWAYS",
					Weight:             10,
				},
			},
		},
	}

	if err := storDb.SetTPDestinations(dests); err != nil {
		t.Error(err)
	}

	if err := dataDb.SetDestinationDrv(dest, utils.NonTransactional); err != nil {
		t.Error(err)
	}

	if err := storDb.SetTPRates(rates); err != nil {
		t.Error(err)
	}

	if err := storDb.SetTPDestinationRates(destRates); err != nil {
		t.Error(err)
	}

	if err := storDb.SetTPTimings(timings); err != nil {
		t.Error(err)
	}
	if err := storDb.SetTPRatingPlans(ratingPlans); err != nil {
		t.Error(err)
	}

	if err := tpr.LoadAll(); err != nil {
		t.Error(err)
	}

	if err := tpr.RemoveFromDatabase(false, false); err != nil {
		t.Error(err)
	}
}
