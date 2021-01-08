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
		BalanceType:      utils.MetaMonetary,
		ExtraParameters:  "",
		ExpirationString: "",
		Balance: &v1Balance{
			Uuid:     "UUID1",
			Id:       utils.MetaDefault,
			Value:    10,
			Weight:   30,
			Category: utils.Call,
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
			Type:       utils.StringPointer(utils.MetaMonetary),
			Value:      &utils.ValueFormula{Static: 10},
			Weight:     utils.Float64Pointer(30),
			Categories: utils.StringMapPointer(utils.ParseStringMap(utils.Call)),
		},
	}
	newact := v1act.AsAction()
	if !reflect.DeepEqual(act, newact) {
		t.Errorf("Expecting: %+v, received: %+v", act, newact)
	}
}
