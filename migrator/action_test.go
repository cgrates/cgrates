// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package migrator

import (
	"reflect"
	"testing"

	"github.com/cgrates/cgrates/utils"

	"github.com/cgrates/cgrates/engine"
)

func TestV1ActionsAsActions(t *testing.T) {
	v1act := &v1Action{
		Id:               "",
		ActionType:       "",
		BalanceType:      "",
		Direction:        "INBOUND",
		ExtraParameters:  "",
		ExpirationString: "",
		Balance:          &v1Balance{},
	}
	act := &engine.Action{
		Id:               "",
		ActionType:       "",
		ExtraParameters:  "",
		ExpirationString: "",
		Weight:           0.00,
		Balance:          &engine.BalanceFilter{},
	}
	newact := v1act.AsAction()
	if !reflect.DeepEqual(act, newact) {
		t.Errorf("Expecting: %+v, received: %+v", act, newact)
	}
}

func TestV1ActionsAsActions2(t *testing.T) {
	v1act := &v1Action{
		Id:               "ID",
		ActionType:       "*log",
		BalanceType:      utils.MONETARY,
		ExtraParameters:  "",
		ExpirationString: "",
		Balance: &v1Balance{
			Uuid:     "UUID1",
			Id:       utils.MetaDefault,
			Value:    10,
			Weight:   30,
			Category: utils.CALL,
		},
	}

	act := &engine.Action{
		Id:               "ID",
		ActionType:       "*log",
		ExtraParameters:  "",
		ExpirationString: "",
		Weight:           0.00,
		Balance: &engine.BalanceFilter{
			Uuid:       utils.StringPointer("UUID1"),
			ID:         utils.StringPointer(utils.MetaDefault),
			Type:       utils.StringPointer(utils.MONETARY),
			Value:      &utils.ValueFormula{Static: 10},
			Weight:     utils.Float64Pointer(30),
			Categories: utils.StringMapPointer(utils.ParseStringMap(utils.CALL)),
		},
	}
	newact := v1act.AsAction()
	if !reflect.DeepEqual(act, newact) {
		t.Errorf("Expecting: %+v, received: %+v", act, newact)
	}
}
