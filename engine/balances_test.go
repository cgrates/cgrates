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
	"log"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

func TestBalanceSortPrecision(t *testing.T) {
	mb1 := &Balance{Weight: 1, precision: 2}
	mb2 := &Balance{Weight: 2, precision: 1}
	var bs Balances
	bs = append(bs, mb2, mb1)
	bs.Sort()
	if bs[0] != mb1 || bs[1] != mb2 {
		t.Error("Buckets not sorted by weight!")
	}
}

func TestBalanceSortPrecisionWeightEqual(t *testing.T) {
	mb1 := &Balance{Weight: 1, precision: 2}
	mb2 := &Balance{Weight: 1, precision: 1}
	var bs Balances
	bs = append(bs, mb2, mb1)
	bs.Sort()
	if bs[0] != mb1 || bs[1] != mb2 {
		t.Error("Buckets not sorted by precision!")
	}
}

func TestBalanceSortPrecisionWeightGreater(t *testing.T) {
	mb1 := &Balance{Weight: 2, precision: 2}
	mb2 := &Balance{Weight: 1, precision: 1}
	var bs Balances
	bs = append(bs, mb2, mb1)
	bs.Sort()
	if bs[0] != mb1 || bs[1] != mb2 {
		t.Error("Buckets not sorted by precision!")
	}
}

func TestBalanceSortWeight(t *testing.T) {
	mb1 := &Balance{Weight: 2, precision: 1}
	mb2 := &Balance{Weight: 1, precision: 1}
	var bs Balances
	bs = append(bs, mb2, mb1)
	bs.Sort()
	if bs[0] != mb1 || bs[1] != mb2 {
		t.Error("Buckets not sorted by precision!")
	}
}

func TestBalanceSortWeight2(t *testing.T) {
	bs := Balances{
		&Balance{ID: "B1", Weight: 2, precision: 1},
		&Balance{ID: "B2", Weight: 1, precision: 1},
	}
	bs.Sort()
	if bs[0].ID != "B1" && bs[1].ID != "B2" {
		t.Error("Buckets not sorted by precision!")
	}
}

func TestBalanceSortWeight3(t *testing.T) {
	bs := Balances{
		&Balance{ID: "B1", Weight: 2, Value: 10.0},
		&Balance{ID: "B2", Weight: 4, Value: 10.0},
	}
	bs.Sort()
	if bs[0].ID != "B2" && bs[1].ID != "B1" {
		t.Error(utils.ToJSON(bs))
	}
}

func TestBalanceSortWeightLess(t *testing.T) {
	mb1 := &Balance{Weight: 1, precision: 1}
	mb2 := &Balance{Weight: 2, precision: 1}
	var bs Balances
	bs = append(bs, mb2, mb1)
	bs.Sort()
	if bs[0] != mb2 || bs[1] != mb1 {
		t.Error("Buckets not sorted by precision!")
	}
}

func TestBalanceEqual(t *testing.T) {
	mb1 := &Balance{Weight: 1, precision: 1, RatingSubject: "1", DestinationIDs: utils.StringMap{}}
	mb2 := &Balance{Weight: 1, precision: 1, RatingSubject: "1", DestinationIDs: utils.StringMap{}}
	mb3 := &Balance{Weight: 1, precision: 1, RatingSubject: "2", DestinationIDs: utils.StringMap{}}
	if !mb1.Equal(mb2) || mb2.Equal(mb3) {
		t.Error("Equal failure!", mb1 == mb2, mb3)
	}
}

func TestBalanceMatchFilter(t *testing.T) {
	mb1 := &Balance{Weight: 1, precision: 1, RatingSubject: "1", DestinationIDs: utils.StringMap{}}
	mb2 := &BalanceFilter{Weight: utils.Float64Pointer(1), RatingSubject: nil, DestinationIDs: nil}
	if !mb1.MatchFilter(mb2, false, false) {
		t.Errorf("Match filter failure: %+v == %+v", mb1, mb2)
	}

	mb1.Uuid, mb2.Uuid = "id", utils.StringPointer("id")
	if !mb1.MatchFilter(mb2, false, false) {
		t.Errorf("Match filter failure: %+v == %+v", mb1, mb2)
	}

}

func TestBalanceMatchFilterEmpty(t *testing.T) {
	mb1 := &Balance{Weight: 1, precision: 1, RatingSubject: "1", DestinationIDs: utils.StringMap{}}
	mb2 := &BalanceFilter{}
	if !mb1.MatchFilter(mb2, false, false) {
		t.Errorf("Match filter failure: %+v == %+v", mb1, mb2)
	}
}

func TestBalanceMatchFilterId(t *testing.T) {
	mb1 := &Balance{ID: "T1", Weight: 2, precision: 2, RatingSubject: "2", DestinationIDs: utils.NewStringMap("NAT")}
	mb2 := &BalanceFilter{ID: utils.StringPointer("T1"), Weight: utils.Float64Pointer(1), RatingSubject: utils.StringPointer("1"), DestinationIDs: nil}
	if !mb1.MatchFilter(mb2, false, false) {
		t.Errorf("Match filter failure: %+v == %+v", mb1, mb2)
	}
}

func TestBalanceMatchFilterDiffId(t *testing.T) {
	mb1 := &Balance{ID: "T1", Weight: 1, precision: 1, RatingSubject: "1", DestinationIDs: utils.StringMap{}}
	mb2 := &BalanceFilter{ID: utils.StringPointer("T2"), Weight: utils.Float64Pointer(1), RatingSubject: utils.StringPointer("1"), DestinationIDs: nil}
	if mb1.MatchFilter(mb2, false, false) {
		t.Errorf("Match filter failure: %+v != %+v", mb1, mb2)
	}
}

func TestBalanceClone(t *testing.T) {
	var mb1 *Balance
	if mb2 := mb1.Clone(); mb2 != nil {
		t.Errorf("Balance should be %v", mb2)
	}

	mb1 = &Balance{Value: 1, Weight: 2, RatingSubject: "test", DestinationIDs: utils.NewStringMap("5")}
	mb2 := mb1.Clone()
	if mb1 == mb2 || !mb1.Equal(mb2) {
		t.Errorf("Cloning failure: \n%+v\n%+v", mb1, mb2)
	}
}

func TestBalanceMatchActionTriggerId(t *testing.T) {
	at := &ActionTrigger{Balance: &BalanceFilter{ID: utils.StringPointer("test")}}
	b := &Balance{ID: "test"}
	if !b.MatchActionTrigger(at) {
		t.Errorf("Error matching action trigger: %+v %+v", b, at)
	}
	b.ID = "test1"
	if b.MatchActionTrigger(at) {
		t.Errorf("Error matching action trigger: %+v %+v", b, at)
	}
	b.ID = ""
	if b.MatchActionTrigger(at) {
		t.Errorf("Error matching action trigger: %+v %+v", b, at)
	}
	b.ID = "test"
	at.Balance.ID = nil
	if !b.MatchActionTrigger(at) {
		t.Errorf("Error matching action trigger: %+v %+v", b, at)
	}
}

func TestBalanceMatchActionTriggerDestination(t *testing.T) {
	at := &ActionTrigger{Balance: &BalanceFilter{DestinationIDs: utils.StringMapPointer(utils.NewStringMap("test"))}}
	b := &Balance{DestinationIDs: utils.NewStringMap("test")}
	if !b.MatchActionTrigger(at) {
		t.Errorf("Error matching action trigger: %+v %+v", b, at)
	}
	b.DestinationIDs = utils.NewStringMap("test1")
	if b.MatchActionTrigger(at) {
		t.Errorf("Error matching action trigger: %+v %+v", b, at)
	}
	b.DestinationIDs = utils.NewStringMap("")
	if b.MatchActionTrigger(at) {
		t.Errorf("Error matching action trigger: %+v %+v", b, at)
	}
	b.DestinationIDs = utils.NewStringMap("test")
	at.Balance.DestinationIDs = nil
	if !b.MatchActionTrigger(at) {
		t.Errorf("Error matching action trigger: %+v %+v", b, at)
	}
}

func TestBalanceMatchActionTriggerWeight(t *testing.T) {
	at := &ActionTrigger{Balance: &BalanceFilter{Weight: utils.Float64Pointer(1)}}
	b := &Balance{Weight: 1}
	if !b.MatchActionTrigger(at) {
		t.Errorf("Error matching action trigger: %+v %+v", b, at)
	}
	b.Weight = 2
	if b.MatchActionTrigger(at) {
		t.Errorf("Error matching action trigger: %+v %+v", b, at)
	}
	b.Weight = 0
	if b.MatchActionTrigger(at) {
		t.Errorf("Error matching action trigger: %+v %+v", b, at)
	}
	b.Weight = 1
	at.Balance.Weight = nil
	if !b.MatchActionTrigger(at) {
		t.Errorf("Error matching action trigger: %+v %+v", b, at)
	}
}

func TestBalanceMatchActionTriggerRatingSubject(t *testing.T) {

	at := &ActionTrigger{Balance: &BalanceFilter{RatingSubject: utils.StringPointer("test")}}
	b := &Balance{RatingSubject: "test"}
	if !b.MatchActionTrigger(at) {
		t.Errorf("Error matching action trigger: %+v %+v", b, at)
	}
	b.RatingSubject = "test1"
	if b.MatchActionTrigger(at) {
		t.Errorf("Error matching action trigger: %+v %+v", b, at)
	}
	b.RatingSubject = ""
	if b.MatchActionTrigger(at) {
		t.Errorf("Error matching action trigger: %+v %+v", b, at)
	}
	b.RatingSubject = "test"
	at.Balance.RatingSubject = nil
	if !b.MatchActionTrigger(at) {
		t.Errorf("Error matching action trigger: %+v %+v", b, at)
	}
}

func TestBalanceMatchActionTriggerSharedGroup(t *testing.T) {
	at := &ActionTrigger{Balance: &BalanceFilter{SharedGroups: utils.StringMapPointer(utils.NewStringMap("test"))}}
	b := &Balance{SharedGroups: utils.NewStringMap("test")}
	if !b.MatchActionTrigger(at) {
		t.Errorf("Error matching action trigger: %+v %+v", b, at)
	}
	b.SharedGroups = utils.NewStringMap("test1")
	if b.MatchActionTrigger(at) {
		t.Errorf("Error matching action trigger: %+v %+v", b, at)
	}
	b.SharedGroups = utils.NewStringMap("")
	if b.MatchActionTrigger(at) {
		t.Errorf("Error matching action trigger: %+v %+v", b, at)
	}
	b.SharedGroups = utils.NewStringMap("test")
	at.Balance.SharedGroups = nil
	if !b.MatchActionTrigger(at) {
		t.Errorf("Error matching action trigger: %+v %+v", b, at)
	}
}

func TestBalanceIsDefault(t *testing.T) {
	b := &Balance{Weight: 0}
	if b.IsDefault() {
		t.Errorf("Balance should not be default: %+v", b)
	}
	b = &Balance{ID: utils.MetaDefault}
	if !b.IsDefault() {
		t.Errorf("Balance should be default: %+v", b)
	}
}

func TestBalanceIsExpiredAt(t *testing.T) {
	//expiration date is 0
	balance := &Balance{}
	var date2 time.Time
	if rcv := balance.IsExpiredAt(date2); rcv {
		t.Errorf("Expecting: false , received: %+v", rcv)
	}
	//expiration date before time t
	balance.ExpirationDate = time.Date(2020, time.April, 18, 23, 0, 3, 0, time.UTC)
	date2 = time.Date(2020, time.April, 18, 23, 0, 4, 0, time.UTC)
	if rcv := balance.IsExpiredAt(date2); !rcv {
		t.Errorf("Expecting: true , received: %+v", rcv)
	}
	//expiration date after time t
	date2 = time.Date(2020, time.April, 18, 23, 0, 2, 0, time.UTC)
	if rcv := balance.IsExpiredAt(date2); rcv {
		t.Errorf("Expecting: false , received: %+v", rcv)
	}
	//time t = 0
	var date3 time.Time
	if rcv := balance.IsExpiredAt(date3); rcv {
		t.Errorf("Expecting: false , received: %+v", rcv)
	}

}

func TestBalanceAsInterface(t *testing.T) {

	b := &Balance{
		Uuid:           "uuid",
		ID:             "id",
		Value:          2.21,
		ExpirationDate: time.Date(2022, 11, 22, 9, 0, 0, 0, time.UTC),
		Weight:         2.88,
		DestinationIDs: utils.StringMap{
			"destId1": true,
			"destId2": true,
		},
		RatingSubject: "rating",
		Categories: utils.StringMap{
			"ctg1": true,
			"ctg2": false,
		},
		SharedGroups: utils.StringMap{
			"shgp1": false,
			"shgp2": true,
		},
		Timings: []*RITiming{
			{
				ID:    "id",
				Years: utils.Years{2, 3},
			},
		},
		TimingIDs: utils.StringMap{
			"timingid1": true,
			"timingid2": false,
		},
		Factor: ValueFactor{
			"factor1": 2.21,
			"factor2": 1.34,
		},
	}

	if _, err := b.FieldAsInterface([]string{}); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	} else if _, err = b.FieldAsInterface([]string{"value"}); err == nil {
		t.Error(err)
	} else if _, err = b.FieldAsInterface([]string{"DestinationIDs[destId1]", "secondVal"}); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	} else if _, err = b.FieldAsInterface([]string{"Categories[ctg1]", "secondVal"}); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	} else if _, err = b.FieldAsInterface([]string{"SharedGroups[shgp1]", "secondVal"}); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	} else if _, err = b.FieldAsInterface([]string{"TimingIDs[timingid1]", "secondVal"}); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	} else if _, err = b.FieldAsInterface([]string{"Timings[zero]"}); err == nil {
		t.Error(err)
	} else if _, err = b.FieldAsInterface([]string{"Timings[2]"}); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	} else if _, err = b.FieldAsInterface([]string{"Timings[2]", "val"}); err == nil {
		t.Error(err)
	} else if _, err = b.FieldAsInterface([]string{"Factor[factor1]", "secondVal"}); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	}

	if _, err = b.FieldAsInterface([]string{"DestinationIDs[destId1]"}); err != nil {
		t.Error(err)
	} else if _, err = b.FieldAsInterface([]string{"Categories[ctg1]"}); err != nil {
		t.Error(err)
	} else if _, err = b.FieldAsInterface([]string{"SharedGroups[shgp1]"}); err != nil {
		t.Error(err)
	} else if _, err = b.FieldAsInterface([]string{"TimingIDs[timingid1]"}); err != nil {
		t.Error(err)
	} else if _, err = b.FieldAsInterface([]string{"Timings[0]"}); err != nil {
		t.Error(err)
	} else if _, err = b.FieldAsInterface([]string{"Factor[factor1]"}); err != nil {
		t.Error(err)
	}
	if _, err = b.FieldAsInterface([]string{utils.Uuid}); err != nil {
		t.Error(err)
	} else if _, err = b.FieldAsInterface([]string{utils.ExpirationDate}); err != nil {
		t.Error(err)
	} else if _, err = b.FieldAsInterface([]string{utils.Weight}); err != nil {
		t.Error(err)
	} else if _, err = b.FieldAsInterface([]string{utils.DestinationIDs}); err != nil {
		t.Error(err)
	} else if _, err = b.FieldAsInterface([]string{utils.DestinationIDs}); err != nil {
		t.Error(err)
	} else if _, err = b.FieldAsInterface([]string{utils.RatingSubject}); err != nil {
		t.Error(err)
	} else if _, err = b.FieldAsInterface([]string{utils.Categories}); err != nil {
		t.Error(err)
	} else if _, err = b.FieldAsInterface([]string{utils.SharedGroups}); err != nil {
		t.Error(err)
	} else if _, err = b.FieldAsInterface([]string{utils.Timings}); err != nil {
		t.Error(err)
	} else if _, err = b.FieldAsInterface([]string{utils.Disabled}); err != nil {
		t.Error(err)
	} else if _, err = b.FieldAsInterface([]string{utils.Factor}); err != nil {
		t.Error(err)
	} else if _, err = b.FieldAsInterface([]string{utils.Blocker}); err != nil {
		t.Error(err)
	} else if _, err = b.FieldAsInterface([]string{utils.TimingIDs}); err != nil {
		t.Error(err)
	}

	if _, err = b.FieldAsInterface([]string{utils.TimingIDs, "val"}); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	} else if _, err = b.FieldAsInterface([]string{utils.Uuid, "val"}); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	} else if _, err = b.FieldAsInterface([]string{utils.ID, "val"}); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	} else if _, err = b.FieldAsInterface([]string{utils.Value, "val"}); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	} else if _, err = b.FieldAsInterface([]string{utils.ExpirationDate, "val"}); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	} else if _, err = b.FieldAsInterface([]string{utils.Weight, "val"}); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	} else if _, err = b.FieldAsInterface([]string{utils.DestinationIDs, "val"}); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	} else if _, err = b.FieldAsInterface([]string{utils.RatingSubject, "val"}); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	} else if _, err = b.FieldAsInterface([]string{utils.Categories, "val"}); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	} else if _, err = b.FieldAsInterface([]string{utils.SharedGroups, "val"}); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	}

}

func TestValueFactorFieldAsInterface(t *testing.T) {
	v := &ValueFactor{
		"FACT_VAL": 20.22,
	}
	if _, err := v.FieldAsInterface([]string{}); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	} else if _, err = v.FieldAsInterface([]string{"TEST"}); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	} else if _, err = v.FieldAsInterface([]string{"FACT_VAL"}); err != nil {
		t.Error(err)
	}
}
func TestValueFactorFieldAsString(t *testing.T) {
	v := &ValueFactor{
		"FACT_VAL": 20.22,
	}
	if _, err = v.FieldAsString([]string{"TEST"}); err == nil {
		t.Error(err)
	} else if _, err = v.FieldAsString([]string{"FACT_VAL"}); err != nil {
		t.Error(err)
	}
}

func TestBalancesHasBalance(t *testing.T) {
	bc := Balances{
		{
			Uuid:           "uuid",
			ID:             "id",
			Value:          12.22,
			ExpirationDate: time.Date(2022, 11, 1, 20, 0, 0, 0, time.UTC),
			Blocker:        true,
			Disabled:       true,
			precision:      2,
		},
		{
			Uuid:           "uuid2",
			ID:             "id2",
			Value:          133.22,
			ExpirationDate: time.Date(2023, 3, 21, 5, 0, 0, 0, time.UTC),
			Blocker:        true,
			Disabled:       true,
			precision:      2,
		},
	}
	balance := &Balance{
		Uuid:           "uuid",
		ID:             "id",
		Value:          12.22,
		ExpirationDate: time.Date(2022, 11, 1, 20, 0, 0, 0, time.UTC),
		Blocker:        true,
		Disabled:       true,
		precision:      2,
	}

	if !bc.HasBalance(balance) {
		t.Error("should be true")
	}

}

func TestBalanceDebitUnits(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, true, nil)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	cd := &CallDescriptor{
		Category:      "postpaid",
		ToR:           utils.MetaVoice,
		Tenant:        "foehn",
		Subject:       "foehn",
		Account:       "foehn",
		Destination:   "0034678096720",
		TimeStart:     time.Date(2015, 4, 24, 7, 59, 4, 0, time.UTC),
		TimeEnd:       time.Date(2015, 4, 24, 8, 2, 0, 0, time.UTC),
		LoopIndex:     0,
		DurationIndex: 17 * time.Second,

		FallbackSubject: "",
		RatingInfos: RatingInfos{
			&RatingInfo{
				MatchedSubject: "*out:foehn:postpaid:foehn",
				MatchedPrefix:  "0034678",
				MatchedDestId:  "SPN_MOB",
				ActivationTime: time.Date(2015, 4, 23, 0, 0, 0, 0, time.UTC),
				RateIntervals: []*RateInterval{
					{
						Timing: &RITiming{
							WeekDays:  []time.Weekday{time.Monday, time.Tuesday, time.Wednesday, time.Thursday, time.Friday},
							StartTime: "08:00:00",
						},
						Rating: &RIRate{
							ConnectFee:       0,
							RoundingMethod:   "*up",
							RoundingDecimals: 6,
							Rates: RateGroups{
								&RGRate{Value: 1, RateIncrement: time.Second, RateUnit: time.Second},
							},
						},
					},
				},
			},
		},
	}
	ub := &Account{
		ID: "vdf:broker",
		BalanceMap: map[string]Balances{
			utils.MetaVoice: {
				&Balance{Value: 20 * float64(time.Second),
					DestinationIDs: utils.NewStringMap("NAT"),
					Weight:         10, RatingSubject: "rif"},
				&Balance{Value: 100 * float64(time.Second),
					DestinationIDs: utils.NewStringMap("RET"), Weight: 20},
			}},
	}

	moneyBalances := Balances{{
		Uuid:           "uuid",
		ID:             "id",
		Value:          12.22,
		ExpirationDate: time.Date(2022, 11, 1, 20, 0, 0, 0, time.UTC),
		Blocker:        true,
		Disabled:       true,
		precision:      2,
	}}
	b := &Balance{
		Uuid:           "uuid",
		ID:             "id",
		Value:          12.22,
		ExpirationDate: time.Date(2022, 11, 1, 20, 0, 0, 0, time.UTC),
		Blocker:        true,
		Disabled:       false,
		precision:      2,
		RatingSubject:  "*zero34",
	}
	fltrs := FilterS{cfg, dm, nil}
	config.SetCgrConfig(cfg)
	exp := &CallCost{Category: "postpaid",
		Tenant:  "foehn",
		Subject: "foehn", Account: "foehn",
		Destination: "0034678096720", ToR: "*voice",
		Cost: 0,
		Timespans: TimeSpans{
			{TimeStart: time.Date(2015, 4, 24, 7, 59, 4, 0, time.UTC),
				TimeEnd: time.Date(2015, 4, 24, 8, 2, 0, 0, time.UTC),
				Cost:    0,
				RateInterval: &RateInterval{
					Rating: &RIRate{
						ConnectFee:       0,
						RoundingDecimals: 0,
						MaxCost:          0,
						Rates: RateGroups{
							{
								GroupIntervalStart: 0,
								Value:              0,
								RateIncrement:      34,
								RateUnit:           34},
						}},
					Weight: 0},
				DurationIndex:  26,
				MatchedSubject: "uuid",
				MatchedPrefix:  "0034678096720",
				MatchedDestId:  "*any",
				RatingPlanId:   "*none",
				CompressFactor: 0}},
		RatedUsage: 0,
	}

	if val, err := b.debitUnits(cd, ub, moneyBalances, true, false, true, &fltrs); err != nil {
		t.Errorf("received %v", err)
	} else if reflect.DeepEqual(val, exp) {
		t.Errorf("expected %+v ,received  %+v", utils.ToJSON(exp), utils.ToJSON(val))
	}
}
func TestBalanceDebitMoneyMaxCostFree(t *testing.T) {
	utils.Logger.SetLogLevel(3)
	utils.Logger.SetSyslog(nil)
	buf := new(bytes.Buffer)
	log.SetOutput(buf)
	defer func() {
		utils.Logger.SetLogLevel(0)
		log.SetOutput(os.Stderr)
	}()
	cd := &CallDescriptor{
		Category:     "postpaid",
		ToR:          utils.MetaVoice,
		Tenant:       "foehn",
		TimeStart:    time.Date(2015, 4, 24, 7, 59, 4, 0, time.UTC),
		TimeEnd:      time.Date(2015, 4, 24, 8, 2, 0, 0, time.UTC),
		MaxCostSoFar: 33,
		testCallcost: &CallCost{
			Category:         "generic",
			Tenant:           "cgrates.org",
			Subject:          "1001",
			Account:          "1001",
			Destination:      "data",
			ToR:              "*data",
			Cost:             0,
			deductConnectFee: true,
			Timespans: TimeSpans{
				{
					TimeStart:     time.Date(2013, 9, 24, 10, 48, 0, 0, time.UTC),
					TimeEnd:       time.Date(2013, 9, 24, 10, 48, 10, 0, time.UTC),
					DurationIndex: 0,
					Increments: Increments{
						{Cost: 2, BalanceInfo: &DebitInfo{
							Monetary: &MonetaryInfo{UUID: "moneya"}},
						}},
					RateInterval: &RateInterval{
						Rating: &RIRate{
							ConnectFee:      0.15,
							MaxCost:         23.2,
							MaxCostStrategy: utils.MetaMaxCostFree,
							Rates: RateGroups{&RGRate{GroupIntervalStart: 0,
								Value: 0.1, RateIncrement: time.Second,
								RateUnit: time.Second}}}},
				},
			},
		},
		FallbackSubject: "",
	}
	ub := &Account{
		ID: "vdf:broker",
		BalanceMap: map[string]Balances{
			utils.MetaVoice: {
				&Balance{Value: 20 * float64(time.Second),
					DestinationIDs: utils.NewStringMap("NAT"),
					Weight:         10, RatingSubject: "rif"},
				&Balance{Value: 100 * float64(time.Second),
					DestinationIDs: utils.NewStringMap("RET"), Weight: 20},
			}},
	}
	moneyBalances := Balances{}
	b := &Balance{
		Uuid:           "uuid",
		ID:             "id",
		Value:          12.22,
		ExpirationDate: time.Date(2022, 11, 1, 20, 0, 0, 0, time.UTC),
		Blocker:        true,
		Disabled:       false,
		precision:      2,
		RatingSubject:  "*val34",
		Factor: ValueFactor{
			"FACT_VAL": 20.22,
		},
	}

	if _, err := b.debitMoney(cd, ub, moneyBalances, true, true, true, nil); err != nil {
		t.Error(err)
	}
}
func TestBalanceDebitMoneyMaxCostDisconnect(t *testing.T) {
	utils.Logger.SetLogLevel(3)
	utils.Logger.SetSyslog(nil)
	buf := new(bytes.Buffer)
	log.SetOutput(buf)
	defer func() {
		utils.Logger.SetLogLevel(0)
		log.SetOutput(os.Stderr)
	}()
	cd := &CallDescriptor{
		Category:     "postpaid",
		ToR:          utils.MetaVoice,
		Tenant:       "foehn",
		TimeStart:    time.Date(2015, 4, 24, 7, 59, 4, 0, time.UTC),
		TimeEnd:      time.Date(2015, 4, 24, 8, 2, 0, 0, time.UTC),
		MaxCostSoFar: 33,
		testCallcost: &CallCost{
			Category:         "generic",
			Tenant:           "cgrates.org",
			Subject:          "1001",
			Account:          "1001",
			Destination:      "data",
			ToR:              "*data",
			Cost:             0,
			deductConnectFee: true,
			Timespans: TimeSpans{
				{
					TimeStart:     time.Date(2013, 9, 24, 10, 48, 0, 0, time.UTC),
					TimeEnd:       time.Date(2013, 9, 24, 10, 48, 10, 0, time.UTC),
					DurationIndex: 0,
					Increments: Increments{
						{Cost: 2, BalanceInfo: &DebitInfo{
							Monetary: &MonetaryInfo{UUID: "moneya"}},
						}},
				},
			},
		},
		FallbackSubject: "",
	}
	ub := &Account{
		ID: "vdf:broker",
		BalanceMap: map[string]Balances{
			utils.MetaVoice: {
				&Balance{Value: 20 * float64(time.Second),
					DestinationIDs: utils.NewStringMap("NAT"),
					Weight:         10, RatingSubject: "rif"},
				&Balance{Value: 100 * float64(time.Second),
					DestinationIDs: utils.NewStringMap("RET"), Weight: 20},
			}},
	}
	moneyBalances := Balances{}
	b := &Balance{
		Uuid:           "uuid",
		ID:             "id",
		Value:          12.22,
		ExpirationDate: time.Date(2022, 11, 1, 20, 0, 0, 0, time.UTC),
		Blocker:        true,
		Disabled:       false,
		precision:      2,
		RatingSubject:  "*val34",
		Factor: ValueFactor{
			"FACT_VAL": 20.22,
		},
	}
	expLog := `Nil RateInterval ERROR on TS:`
	if _, err := b.debitMoney(cd, ub, moneyBalances, true, true, true, nil); err == nil {
		t.Error(err)
	} else if rcvLog := buf.String(); !strings.Contains(rcvLog, expLog) {
		t.Errorf("Logger %v,doesn't contain %v", rcvLog, expLog)
	}
	cd.testCallcost.Timespans[0].RateInterval = &RateInterval{
		Rating: &RIRate{
			ConnectFee:      0.15,
			MaxCost:         23.2,
			MaxCostStrategy: utils.MetaMaxCostDisconnect,
			Rates: RateGroups{&RGRate{GroupIntervalStart: 0,
				Value: 0.1, RateIncrement: time.Second,
				RateUnit: time.Second}}}}
	if _, err := b.debitMoney(cd, ub, moneyBalances, true, true, true, nil); err != nil {
		t.Error(err)
	}
}
func TestBalanceDebitMoney(t *testing.T) {
	utils.Logger.SetLogLevel(3)
	utils.Logger.SetSyslog(nil)
	buf := new(bytes.Buffer)
	log.SetOutput(buf)
	defer func() {
		utils.Logger.SetLogLevel(0)
		log.SetOutput(os.Stderr)
	}()
	cd := &CallDescriptor{
		Category:     "postpaid",
		ToR:          utils.MetaVoice,
		Tenant:       "foehn",
		TimeStart:    time.Date(2015, 4, 24, 7, 59, 4, 0, time.UTC),
		TimeEnd:      time.Date(2015, 4, 24, 8, 2, 0, 0, time.UTC),
		MaxCostSoFar: 33,
		testCallcost: &CallCost{
			Category:         "generic",
			Tenant:           "cgrates.org",
			Subject:          "1001",
			Account:          "1001",
			Destination:      "data",
			ToR:              "*data",
			Cost:             0,
			deductConnectFee: true,
			Timespans: TimeSpans{
				{
					TimeStart:     time.Date(2013, 9, 24, 10, 48, 0, 0, time.UTC),
					TimeEnd:       time.Date(2013, 9, 24, 10, 48, 10, 0, time.UTC),
					DurationIndex: 0,
					Increments: Increments{
						{Cost: 2, BalanceInfo: &DebitInfo{
							Monetary: &MonetaryInfo{UUID: "moneya"}},
						}},
					RateInterval: &RateInterval{
						Rating: &RIRate{
							ConnectFee: 0.15,
							MaxCost:    23.2,

							Rates: RateGroups{&RGRate{GroupIntervalStart: 0,
								Value: 0.1, RateIncrement: time.Second,
								RateUnit: time.Second}}}},
				},
			},
		},
		FallbackSubject: "",
	}
	ub := &Account{
		ID: "vdf:broker",
		BalanceMap: map[string]Balances{
			utils.MetaVoice: {
				&Balance{Value: 20 * float64(time.Second),
					DestinationIDs: utils.NewStringMap("NAT"),
					Weight:         10, RatingSubject: "rif"},
				&Balance{Value: 100 * float64(time.Second),
					DestinationIDs: utils.NewStringMap("RET"), Weight: 20},
			}},
	}
	moneyBalances := Balances{}
	b := &Balance{
		ID:             "id",
		Value:          12.22,
		ExpirationDate: time.Date(2022, 11, 1, 20, 0, 0, 0, time.UTC),
		precision:      2,
		RatingSubject:  "*val34",
		Factor: ValueFactor{
			"FACT_VAL": 20.22,
		},
	}
	if _, err := b.debitMoney(cd, ub, moneyBalances, true, true, true, nil); err != nil {
		t.Error(err)
	}
}
func TestBalanceDebitUnits2(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, true, nil)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	cd := &CallDescriptor{
		Category:     "postpaid",
		ToR:          utils.MetaVoice,
		Tenant:       "foehn",
		TimeStart:    time.Date(2015, 4, 24, 7, 59, 4, 0, time.UTC),
		TimeEnd:      time.Date(2015, 4, 24, 8, 2, 0, 0, time.UTC),
		MaxCostSoFar: 23.8,
		testCallcost: &CallCost{
			Category: "generic",
			Tenant:   "cgrates.org",
			Subject:  "1001",
			Account:  "1001",

			Destination:      "data",
			ToR:              "*data",
			Cost:             0,
			deductConnectFee: true,
			Timespans: TimeSpans{
				{
					TimeStart:     time.Date(2013, 9, 24, 10, 48, 0, 0, time.UTC),
					TimeEnd:       time.Date(2013, 9, 24, 10, 48, 10, 0, time.UTC),
					DurationIndex: 0,
					Increments: Increments{
						{Cost: 2, BalanceInfo: &DebitInfo{
							Monetary: &MonetaryInfo{UUID: "moneya"}},
						}},
					RateInterval: &RateInterval{
						Rating: &RIRate{
							ConnectFee:      0.15,
							MaxCost:         23.2,
							MaxCostStrategy: utils.MetaMaxCostDisconnect,
							Rates: RateGroups{&RGRate{GroupIntervalStart: 0,
								Value: 0.1, RateIncrement: time.Second,
								RateUnit: time.Second}}}},
				},
			},
		},
		FallbackSubject: "",
	}
	ub := &Account{
		ID: "vdf:broker",
		BalanceMap: map[string]Balances{
			utils.MetaVoice: {
				&Balance{Value: 20 * float64(time.Second),
					DestinationIDs: utils.NewStringMap("NAT"),
					Weight:         10, RatingSubject: "rif"},
				&Balance{Value: 100 * float64(time.Second),
					DestinationIDs: utils.NewStringMap("RET"), Weight: 20},
			}},
	}

	moneyBalances := Balances{{
		Uuid:           "uuid",
		ID:             "id",
		Value:          12.22,
		ExpirationDate: time.Date(2022, 11, 1, 20, 0, 0, 0, time.UTC),
		Blocker:        true,
		Disabled:       true,
		precision:      2,
	}}
	b := &Balance{
		Uuid:           "uuid",
		ID:             "id",
		Value:          12.22,
		ExpirationDate: time.Date(2022, 11, 1, 20, 0, 0, 0, time.UTC),
		Blocker:        true,
		Disabled:       false,
		precision:      2,
		RatingSubject:  "*val34",
		Factor: ValueFactor{
			"FACT_VAL": 20.22,
		},
	}
	fltrs := FilterS{cfg, dm, nil}
	config.SetCgrConfig(cfg)

	if _, err := b.debitUnits(cd, ub, moneyBalances, true, true, true, &fltrs); err != nil {
		t.Errorf("received %v", err)
	}
}

func TestGetMinutesForCredi(t *testing.T) {
	utils.Logger.SetLogLevel(4)
	utils.Logger.SetSyslog(nil)
	buf := new(bytes.Buffer)
	log.SetOutput(buf)
	defer func() {
		utils.Logger.SetLogLevel(0)
		log.SetOutput(os.Stderr)
	}()
	b := &Balance{
		Value:          20 * float64(time.Second),
		DestinationIDs: utils.NewStringMap("NAT"),
		Weight:         10, RatingSubject: "rif",
	}
	cd := &CallDescriptor{
		Category:      "postpaid",
		ToR:           utils.MetaVoice,
		Tenant:        "foehn",
		Subject:       "foehn",
		Account:       "foehn",
		Destination:   "0034678096720",
		TimeStart:     time.Date(2015, 4, 24, 7, 59, 4, 0, time.UTC),
		TimeEnd:       time.Date(2015, 4, 24, 8, 2, 0, 0, time.UTC),
		LoopIndex:     0,
		DurationIndex: 176 * time.Second,
	}
	if dur, _ := b.GetMinutesForCredit(cd, 12); dur != 0 {
		t.Error(err)
	}
	expLog := `Error getting new cost for balance subject:`
	if rcvLog := buf.String(); !strings.Contains(rcvLog, expLog) {
		t.Errorf("expected %v,received %v", utils.ToJSON(expLog), utils.ToJSON(rcvLog))
	}
}
