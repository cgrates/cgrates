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
	"testing"
	"time"

	"reflect"

	"github.com/cgrates/cgrates/utils"
)

func TestNewBalanceFilter(t *testing.T) {
	attrs := map[string]any{}
	expected := &BalanceFilter{}
	if rply, err := NewBalanceFilter(attrs, ""); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expected: %s ,received: %s", utils.ToJSON(expected), utils.ToJSON(rply))
	}

	tNow := time.Now()
	attrs = map[string]any{
		utils.ID:             "ID",
		utils.UUID:           "UUID",
		utils.Value:          10.5,
		utils.ExpiryTime:     tNow,
		utils.Weight:         10,
		utils.DestinationIDs: "dst1;dst2",
		utils.RatingSubject:  "*zero",
		utils.Categories:     "call;voice",
		utils.SharedGroups:   "shrdGroup",
		utils.TimingIDs:      "*asap",
		utils.Disabled:       false,
		utils.Blocker:        true,
	}
	expected = &BalanceFilter{
		ID:             utils.StringPointer("ID"),
		Uuid:           utils.StringPointer("UUID"),
		Value:          &utils.ValueFormula{Static: 10.5},
		ExpirationDate: utils.TimePointer(tNow),
		Weight:         utils.Float64Pointer(10),
		DestinationIDs: utils.StringMapPointer(utils.NewStringMap("dst1", "dst2")),
		RatingSubject:  utils.StringPointer("*zero"),
		Categories:     utils.StringMapPointer(utils.NewStringMap("call", "voice")),
		SharedGroups:   utils.StringMapPointer(utils.NewStringMap("shrdGroup")),
		TimingIDs:      utils.StringMapPointer(utils.NewStringMap("*asap")),
		Disabled:       utils.BoolPointer(false),
		Blocker:        utils.BoolPointer(true),
	}
	if rply, err := NewBalanceFilter(attrs, ""); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expected: %s ,received: %s", utils.ToJSON(expected), utils.ToJSON(rply))
	}
	attrs[utils.Blocker] = "10"
	if _, err := NewBalanceFilter(attrs, ""); err == nil {
		t.Error("Expecxted error received nil")
	}
	attrs[utils.Disabled] = "10"
	if _, err := NewBalanceFilter(attrs, ""); err == nil {
		t.Error("Expecxted error received nil")
	}
	attrs[utils.Weight] = "NotFloat"
	if _, err := NewBalanceFilter(attrs, ""); err == nil {
		t.Error("Expecxted error received nil")
	}
	attrs[utils.ExpirationDate] = "NotTime"
	if _, err := NewBalanceFilter(attrs, ""); err == nil {
		t.Error("Expecxted error received nil")
	}
	attrs[utils.Value] = "NotFloat"
	if _, err := NewBalanceFilter(attrs, ""); err == nil {
		t.Error("Expecxted error received nil")
	}
}

func TestBalanceFilterClone(t *testing.T) {
	bf := &BalanceFilter{}
	eOut := &BalanceFilter{}
	if rcv := bf.Clone(); !reflect.DeepEqual(rcv, eOut) {
		t.Errorf("Expecting: %+v,\n received: %+v", eOut, rcv)
	}
	bf = &BalanceFilter{
		Uuid: utils.StringPointer("Uuid_test"),
		ID:   utils.StringPointer("ID_test"),
		Type: utils.StringPointer("Type_test"),
		Value: &utils.ValueFormula{
			Method: "ValueMethod_test",
		},
		ExpirationDate: utils.TimePointer(time.Date(2020, time.April, 18, 23, 0, 4, 0, time.UTC)),
		Weight:         utils.Float64Pointer(0.7),
		DestinationIDs: &utils.StringMap{
			"DestinationIDs_true":  true,
			"DestinationIDs_false": false,
		},
		RatingSubject: utils.StringPointer("RatingSubject_test"),
		Categories: &utils.StringMap{
			"Categories_true":  true,
			"Categories_false": false,
		},
		SharedGroups: &utils.StringMap{
			"SharedGroups_true":  true,
			"SharedGroups_false": false,
		},
		TimingIDs: &utils.StringMap{
			"TimingIDs_true":  true,
			"TimingIDs_false": false,
		},
		Timings: []*RITiming{
			{Years: utils.Years{2019}},
			{Months: utils.Months{4}},
		},
		Disabled: utils.BoolPointer(true),
		Factor:   &ValueFactor{AccountActionsCSVContent: 0.7},
		Blocker:  utils.BoolPointer(true),
	}
	eOut = &BalanceFilter{
		Uuid: utils.StringPointer("Uuid_test"),
		ID:   utils.StringPointer("ID_test"),
		Type: utils.StringPointer("Type_test"),
		Value: &utils.ValueFormula{
			Method: "ValueMethod_test",
		},
		ExpirationDate: utils.TimePointer(time.Date(2020, time.April, 18, 23, 0, 4, 0, time.UTC)),
		Weight:         utils.Float64Pointer(0.7),
		DestinationIDs: &utils.StringMap{
			"DestinationIDs_true":  true,
			"DestinationIDs_false": false,
		},
		RatingSubject: utils.StringPointer("RatingSubject_test"),
		Categories: &utils.StringMap{
			"Categories_true":  true,
			"Categories_false": false,
		},
		SharedGroups: &utils.StringMap{
			"SharedGroups_true":  true,
			"SharedGroups_false": false,
		},
		TimingIDs: &utils.StringMap{
			"TimingIDs_true":  true,
			"TimingIDs_false": false,
		},
		Timings: []*RITiming{
			{Years: utils.Years{2019}},
			{Months: utils.Months{4}},
		},
		Disabled: utils.BoolPointer(true),
		Factor:   &ValueFactor{AccountActionsCSVContent: 0.7},
		Blocker:  utils.BoolPointer(true),
	}
	rcv := bf.Clone()
	if !reflect.DeepEqual(rcv, eOut) {
		t.Errorf("Expecting: %+v,\n received: %+v", eOut, rcv)
	}
	rcv.Weight = utils.Float64Pointer(0.8)
	if *bf.Weight != 0.7 {
		t.Errorf("Expecting: 0.7, received: %+v", *bf.Weight)
	}
}

func TestBalanceFilterLoadFromBalance(t *testing.T) {
	str := "test"
	fl := 1.2
	tm := time.Date(2009, 11, 17, 20, 34, 58, 651387237, time.UTC)
	sm := utils.StringMap{"test": true}
	rt := []*RITiming{{}}
	bl := true
	vf := ValueFactor{"test": 1.5}
	acc := Account{
		ID: "test",
	}
	nm := 2

	bf := BalanceFilter{
		Value: &utils.ValueFormula{},
	}

	b := Balance{
		Uuid:           str,
		ID:             str,
		Value:          fl,
		ExpirationDate: tm,
		Weight:         fl,
		DestinationIDs: sm,
		RatingSubject:  str,
		Categories:     sm,
		SharedGroups:   sm,
		Timings:        rt,
		TimingIDs:      sm,
		Disabled:       bl,
		Factor:         vf,
		Blocker:        bl,
		precision:      nm,
		account:        &acc,
		dirty:          bl,
	}

	exp := &BalanceFilter{
		Uuid: &str,
		ID:   &str,
		Value: &utils.ValueFormula{
			Static: fl,
		},
		ExpirationDate: &tm,
		Weight:         &fl,
		DestinationIDs: &sm,
		RatingSubject:  &str,
		Categories:     &sm,
		SharedGroups:   &sm,
		TimingIDs:      &sm,
		Timings:        rt,
		Disabled:       &bl,
		Factor:         &vf,
		Blocker:        &bl,
	}

	rcv := bf.LoadFromBalance(&b)

	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("expected %v, received %v", utils.ToJSON(exp), utils.ToJSON(rcv))
	}
}

func TestBalanceFilterGetType(t *testing.T) {
	bp := BalanceFilter{}

	rcv := bp.GetType()

	if rcv != "" {
		t.Error(rcv)
	}
}

func TestBalanceFilterGetValue(t *testing.T) {
	bp := BalanceFilter{
		Value: &utils.ValueFormula{
			Method: "test",
		},
	}

	rcv := bp.GetValue()

	if rcv != 0.0 {
		t.Error(rcv)
	}
}

func TestBalanceFilterSetValue(t *testing.T) {
	bp := BalanceFilter{}

	bp.SetValue(1.5)

	if bp.Value.Static != 1.5 {
		t.Error(bp.Value.Static)
	}
}

func TestBalanceFilterGetUuid(t *testing.T) {
	str := "test"

	bp := BalanceFilter{
		Uuid: &str,
	}

	rcv := bp.GetUuid()

	if rcv != str {
		t.Error(rcv)
	}
}

func TestBalanceFilterGetCategories(t *testing.T) {
	bp := BalanceFilter{
		Categories: &utils.StringMap{"test": true},
	}

	rcv := bp.GetCategories()

	if !reflect.DeepEqual(rcv, utils.StringMap{"test": true}) {
		t.Error(rcv)
	}
}

func TestBalanceFilterGetTimings(t *testing.T) {
	bp := BalanceFilter{
		TimingIDs: &utils.StringMap{"test": true},
	}

	rcv := bp.GetTimingIDs()

	if !reflect.DeepEqual(rcv, utils.StringMap{"test": true}) {
		t.Error(rcv)
	}
}

func TestBalanceFilterGetFactor(t *testing.T) {
	bp := BalanceFilter{
		Factor: &ValueFactor{"test": 1.5},
	}

	rcv := bp.GetFactor()

	if !reflect.DeepEqual(rcv, ValueFactor{"test": 1.5}) {
		t.Error(rcv)
	}
}

func TestBalanceFilterEmptyExpirationDate(t *testing.T) {
	tm := time.Date(2009, 11, 17, 20, 34, 58, 651387237, time.UTC)

	bp := BalanceFilter{
		ExpirationDate: &tm,
	}

	rcv := bp.EmptyExpirationDate()

	if rcv != false {
		t.Error(rcv)
	}
}

func TestBalanceFilterModifyBalance(t *testing.T) {
	str := "test"
	fl := 1.2
	tm := time.Date(2009, 11, 17, 20, 34, 58, 651387237, time.UTC)
	sm := utils.StringMap{"test": true}
	rt := []*RITiming{{}}
	bl := true
	acc := Account{
		ID: "test",
	}
	nm := 2

	str2 := "testt"
	fl2 := 1.5
	tm2 := time.Date(2010, 11, 17, 20, 34, 58, 651387237, time.UTC)
	sm2 := utils.StringMap{"testt": false}
	rt2 := []*RITiming{{}}
	bl2 := true
	acc2 := Account{
		ID: "test",
	}
	nm2 := 2

	bf := BalanceFilter{
		ID: &str2,
		Value: &utils.ValueFormula{
			Static: fl2,
		},
		ExpirationDate: &tm2,
		Weight:         &fl2,
		DestinationIDs: &sm2,
		RatingSubject:  &str2,
		Categories:     &sm2,
		SharedGroups:   &sm2,
		TimingIDs:      &sm2,
		Timings:        rt2,
		Disabled:       &bl2,
		Blocker:        &bl2,
	}

	b := Balance{
		ID:             str,
		Value:          fl,
		ExpirationDate: tm,
		Weight:         fl,
		DestinationIDs: sm,
		RatingSubject:  str,
		Categories:     sm,
		SharedGroups:   sm,
		Timings:        rt,
		TimingIDs:      sm,
		Disabled:       bl,
		Blocker:        bl,
		precision:      nm,
		account:        &acc,
		dirty:          bl,
	}

	exp := Balance{
		ID:             str2,
		Value:          fl2,
		ExpirationDate: tm2,
		Weight:         fl2,
		DestinationIDs: sm2,
		RatingSubject:  str2,
		Categories:     sm2,
		SharedGroups:   sm2,
		Timings:        rt2,
		TimingIDs:      sm2,
		Disabled:       bl2,
		Blocker:        bl2,
		precision:      nm2,
		account:        &acc2,
		dirty:          bl2,
	}

	bf.ModifyBalance(&b)

	if !reflect.DeepEqual(b, exp) {
		t.Error(exp, b)
	}

	bf.ModifyBalance(nil)
}
