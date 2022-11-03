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

	"reflect"

	"github.com/cgrates/cgrates/utils"
)

func TestNewBalanceFilter(t *testing.T) {
	attrs := map[string]interface{}{}
	expected := &BalanceFilter{}
	if rply, err := NewBalanceFilter(attrs, ""); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expected: %s ,received: %s", utils.ToJSON(expected), utils.ToJSON(rply))
	}

	tNow := time.Now()
	attrs = map[string]interface{}{
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
	attrs[utils.ExpiryTime] = "NotTime"
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

func TestBalanceLoadFromBalance(t *testing.T) {
	bf := &BalanceFilter{
		Uuid: utils.StringPointer(""),
		ID:   utils.StringPointer(""),
		Value: &utils.ValueFormula{
			Static: 0},
		ExpirationDate: utils.TimePointer(time.Date(2022, 12, 21, 20, 0, 0, 0, time.UTC)),
		Weight:         utils.Float64Pointer(0),
		DestinationIDs: &utils.StringMap{},
		RatingSubject:  utils.StringPointer(""),
		Categories:     &utils.StringMap{},
		SharedGroups:   &utils.StringMap{},
		Timings:        []*RITiming{},
		TimingIDs:      &utils.StringMap{},
		Factor:         &ValueFactor{},
		Disabled:       utils.BoolPointer(true),
		Blocker:        utils.BoolPointer(true),
	}
	b := &Balance{
		Uuid:           "uuid",
		ID:             "id",
		Value:          20.4,
		ExpirationDate: time.Date(2022, 12, 21, 20, 0, 0, 0, time.UTC),
		Weight:         533.43,
		DestinationIDs: utils.StringMap{
			"key": true,
		},
		RatingSubject: "rate",
		Categories: utils.StringMap{
			"category": true,
		},
		SharedGroups: utils.StringMap{
			"sharedgroup": true,
		},
		Timings: []*RITiming{
			{ID: "id",
				Years:     utils.Years([]int{2, 2}),
				Months:    utils.Months([]time.Month{2, 2}),
				MonthDays: utils.MonthDays([]int{2, 22, 11}),
				WeekDays:  utils.WeekDays([]time.Weekday{0}),
			},
			{
				ID:        "id2",
				Years:     utils.Years([]int{1, 3, 2}),
				Months:    utils.Months([]time.Month{2, 5, 6}),
				MonthDays: utils.MonthDays([]int{2, 22, 11, 6, 4}),
				WeekDays:  utils.WeekDays([]time.Weekday{0, 2}),
			}},
		TimingIDs: utils.StringMap{
			"timing": true,
		},
		Factor: ValueFactor{
			"valfac": 22,
		},
		Disabled: true,
		Blocker:  true,
	}
	eOut := &BalanceFilter{
		Uuid: utils.StringPointer("uuid"),
		ID:   utils.StringPointer("id"),
		Value: &utils.ValueFormula{
			Static: 20.4,
		},
		ExpirationDate: utils.TimePointer(time.Date(2022, 12, 21, 20, 0, 0, 0, time.UTC)),
		Weight:         utils.Float64Pointer(533.43),
		DestinationIDs: &utils.StringMap{
			"key": true,
		},
		RatingSubject: utils.StringPointer("rate"),
		Categories: &utils.StringMap{
			"category": true,
		},
		SharedGroups: &utils.StringMap{
			"sharedgroup": true,
		},
		Timings: []*RITiming{
			{ID: "id",
				Years:     utils.Years([]int{2, 2}),
				Months:    utils.Months([]time.Month{2, 2}),
				MonthDays: utils.MonthDays([]int{2, 22, 11}),
				WeekDays:  utils.WeekDays([]time.Weekday{0}),
			},
			{
				ID:        "id2",
				Years:     utils.Years([]int{1, 3, 2}),
				Months:    utils.Months([]time.Month{2, 5, 6}),
				MonthDays: utils.MonthDays([]int{2, 22, 11, 6, 4}),
				WeekDays:  utils.WeekDays([]time.Weekday{0, 2}),
			},
		},
		TimingIDs: &utils.StringMap{
			"timing": true,
		},
		Factor: &ValueFactor{
			"valfac": 22,
		},
		Disabled: utils.BoolPointer(true),
		Blocker:  utils.BoolPointer(true),
	}
	if val := bf.LoadFromBalance(b); !reflect.DeepEqual(val, eOut) {
		t.Errorf("expected %v ,received %v", utils.ToJSON(eOut), utils.ToJSON(val))
	}

}

func TestBalanceFilterFieldAsInterface(t *testing.T) {
	bp := &BalanceFilter{}

	if _, err := bp.FieldAsInterface([]string{}); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	} else if _, err = bp.FieldAsInterface([]string{"DestinationIDs[key]"}); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	} else if _, err = bp.FieldAsInterface([]string{"Categories[indx]"}); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	} else if _, err = bp.FieldAsInterface([]string{"SharedGroups[indx]"}); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	} else if _, err = bp.FieldAsInterface([]string{"TimingIDs[indx]"}); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	} else if _, err = bp.FieldAsInterface([]string{"Timings[indx]"}); err == nil {
		t.Error(err)
	} else if _, err = bp.FieldAsInterface([]string{"Factor[indx]"}); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	} else if _, err = bp.FieldAsInterface([]string{"Uuid"}); err != nil {
		t.Error(err)
	} else if _, err = bp.FieldAsInterface([]string{"ID"}); err != nil {
		t.Error(err)
	} else if _, err = bp.FieldAsInterface([]string{"Type"}); err != nil {
		t.Error(err)
	} else if _, err = bp.FieldAsInterface([]string{"ExpirationDate"}); err != nil {
		t.Error(err)
	} else if _, err = bp.FieldAsInterface([]string{"Weight"}); err != nil {
		t.Error(err)
	} else if _, err = bp.FieldAsInterface([]string{"RatingSubject"}); err != nil {
		t.Error(err)
	} else if _, err = bp.FieldAsInterface([]string{"Disabled"}); err != nil {
		t.Error(err)
	} else if _, err = bp.FieldAsInterface([]string{"Blocker"}); err != nil {
		t.Error(err)
	} else if _, err = bp.FieldAsInterface([]string{"DestinationIDs"}); err != nil {
		t.Error(err)
	} else if _, err = bp.FieldAsInterface([]string{"Categories"}); err != nil {
		t.Error(err)
	} else if _, err = bp.FieldAsInterface([]string{"SharedGroups"}); err != nil {
		t.Error(err)
	} else if _, err = bp.FieldAsInterface([]string{"Timings"}); err != nil {
		t.Error(err)
	} else if _, err = bp.FieldAsInterface([]string{"TimingIDs"}); err != nil {
		t.Error(err)
	} else if _, err = bp.FieldAsInterface([]string{"Factor"}); err != nil {
		t.Error(err)
	} else if _, err = bp.FieldAsInterface([]string{"Value"}); err != nil {
		t.Error(err)
	}

	bp = &BalanceFilter{
		Uuid: utils.StringPointer("uuid"),
		ID:   utils.StringPointer("id"),
		Value: &utils.ValueFormula{
			Static: 20.4,
		},
		ExpirationDate: utils.TimePointer(time.Date(2022, 12, 21, 20, 0, 0, 0, time.UTC)),
		Weight:         utils.Float64Pointer(533.43),
		DestinationIDs: &utils.StringMap{
			"key": true,
		},
		RatingSubject: utils.StringPointer("rate"),
		Categories: &utils.StringMap{
			"category": true,
		},
		SharedGroups: &utils.StringMap{
			"sharedgroup": true,
		},
		Timings: []*RITiming{
			{ID: "id",
				Years:     utils.Years([]int{2, 2}),
				Months:    utils.Months([]time.Month{2, 2}),
				MonthDays: utils.MonthDays([]int{2, 22, 11}),
				WeekDays:  utils.WeekDays([]time.Weekday{0}),
			},
			{
				ID:        "id2",
				Years:     utils.Years([]int{1, 3, 2}),
				Months:    utils.Months([]time.Month{2, 5, 6}),
				MonthDays: utils.MonthDays([]int{2, 22, 11, 6, 4}),
				WeekDays:  utils.WeekDays([]time.Weekday{0, 2}),
			},
		},
		TimingIDs: &utils.StringMap{
			"timing": true,
		},
		Factor: &ValueFactor{
			"valfac": 22,
		},
		Disabled: utils.BoolPointer(false),
		Blocker:  utils.BoolPointer(false),
	}

	if _, err := bp.FieldAsInterface([]string{"DestinationIDs[indx]"}); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	} else if _, err := bp.FieldAsInterface([]string{"Categories[indx]"}); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	} else if _, err := bp.FieldAsInterface([]string{"SharedGroups[indx]"}); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	} else if _, err := bp.FieldAsInterface([]string{"TimingIDs[indx]"}); err == nil {
		t.Error(err)
	} else if _, err := bp.FieldAsInterface([]string{"Factor[indx]"}); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	}

	if _, err := bp.FieldAsInterface([]string{"DestinationIDs[key]"}); err != nil {
		t.Error(err)
	} else if _, err := bp.FieldAsInterface([]string{"Categories[category]"}); err != nil {
		t.Error(err)
	} else if _, err := bp.FieldAsInterface([]string{"SharedGroups[sharedgroup]"}); err != nil {
		t.Error(err)
	} else if _, err := bp.FieldAsInterface([]string{"TimingIDs[timing]"}); err != nil {
		t.Error(err)
	} else if _, err := bp.FieldAsInterface([]string{"Factor[valfac]"}); err != nil {
		t.Error(err)
	}

	if _, err = bp.FieldAsInterface([]string{"Timings[three]"}); err == nil {
		t.Error(err)
	} else if _, err = bp.FieldAsInterface([]string{"Timings[3]"}); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	} else if _, err = bp.FieldAsInterface([]string{"Timings[0]"}); err != nil {
		t.Error(err)
	}
	if _, err := bp.FieldAsInterface([]string{"Uuid"}); err != nil {
		t.Error(err)
	} else if _, err = bp.FieldAsInterface([]string{"Uuid", "test"}); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	} else if _, err = bp.FieldAsInterface([]string{"ID"}); err != nil {
		t.Error(err)
	} else if _, err = bp.FieldAsInterface([]string{"ID", "test"}); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	} else if _, err = bp.FieldAsInterface([]string{"Type"}); err != nil {
		t.Error(err)
	} else if _, err = bp.FieldAsInterface([]string{"Type", "test"}); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	} else if _, err = bp.FieldAsInterface([]string{"ExpirationDate"}); err != nil {
		t.Error(err)
	} else if _, err = bp.FieldAsInterface([]string{"ExpirationDate", "test"}); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	} else if _, err = bp.FieldAsInterface([]string{"Weight"}); err != nil {
		t.Error(err)
	} else if _, err = bp.FieldAsInterface([]string{"Weight", "test"}); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	} else if _, err = bp.FieldAsInterface([]string{"Type"}); err != nil {
		t.Error(err)
	} else if _, err = bp.FieldAsInterface([]string{"Type", "test"}); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	} else if _, err = bp.FieldAsInterface([]string{"RatingSubject"}); err != nil {
		t.Error(err)
	} else if _, err = bp.FieldAsInterface([]string{"RatingSubject", "test"}); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	} else if _, err = bp.FieldAsInterface([]string{"Disabled"}); err != nil {
		t.Error(err)
	} else if _, err = bp.FieldAsInterface([]string{"Disabled", "test"}); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	} else if _, err = bp.FieldAsInterface([]string{"RatingSubject"}); err != nil {
		t.Error(err)
	} else if _, err = bp.FieldAsInterface([]string{"RatingSubject", "test"}); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	} else if _, err = bp.FieldAsInterface([]string{"Blocker"}); err != nil {
		t.Error(err)
	} else if _, err = bp.FieldAsInterface([]string{"Blocker", "test"}); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	} else if _, err = bp.FieldAsInterface([]string{"DestinationIDs"}); err != nil {
		t.Error(err)
	} else if _, err = bp.FieldAsInterface([]string{"DestinationIDs", "test"}); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	} else if _, err = bp.FieldAsInterface([]string{"Categories"}); err != nil {
		t.Error(err)
	} else if _, err = bp.FieldAsInterface([]string{"Categories", "test"}); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	} else if _, err = bp.FieldAsInterface([]string{"SharedGroups"}); err != nil {
		t.Error(err)
	} else if _, err = bp.FieldAsInterface([]string{"SharedGroups", "test"}); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	} else if _, err = bp.FieldAsInterface([]string{"Timings"}); err != nil {
		t.Error(err)
	} else if _, err = bp.FieldAsInterface([]string{"Timings", "id2"}); err != nil {
		t.Error(err)
	} else if _, err = bp.FieldAsInterface([]string{"Timings", "test"}); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	}

}

func TestBalanceFilterFieldAsString(t *testing.T) {
	bp := &BalanceFilter{
		Uuid: utils.StringPointer("uuid"),
		ID:   utils.StringPointer("id"),
	}

	if _, err := bp.FieldAsString([]string{}); err == nil {
		t.Error(err)
	} else if _, err = bp.FieldAsString([]string{"Uuid"}); err != nil {
		t.Error(err)
	}
}

func TestBalanceFilterModifyBalance(t *testing.T) {

	bf := &BalanceFilter{
		ID: utils.StringPointer("id"),

		ExpirationDate: utils.TimePointer(time.Date(2022, 12, 24, 10, 0, 0, 0, time.UTC)),
		RatingSubject:  utils.StringPointer("rating"),
		Categories: &utils.StringMap{
			"exp": true,
		},
		SharedGroups: &utils.StringMap{
			"shared": false,
		},
		TimingIDs: &utils.StringMap{
			"one": true,
		},
		Blocker: utils.BoolPointer(true),
		Timings: []*RITiming{
			{
				ID:    "tId",
				Years: utils.Years{2, 1},
			},
		},
		Disabled: utils.BoolPointer(true),
	}
	b := &Balance{}

	exp := &Balance{
		ID:             "id",
		ExpirationDate: time.Date(2022, 12, 24, 10, 0, 0, 0, time.UTC),
		Weight:         0,
		RatingSubject:  "rating",
		Categories:     utils.StringMap{"exp": true},
		SharedGroups: utils.StringMap{
			"shared": false},
		Timings: []*RITiming{
			{
				ID:    "tId",
				Years: utils.Years{2, 1},
			},
		},
		TimingIDs: utils.StringMap{
			"one": true},
		Disabled: true,

		Blocker: true}

	bf.ModifyBalance(b)
	if reflect.DeepEqual(b, exp) {
		t.Errorf("expected %v ,received %v", utils.ToJSON(exp), utils.ToJSON(b))
	}

}
