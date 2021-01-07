/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT MetaAny WARRANTY; without even the implied warranty of
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
