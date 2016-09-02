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
	"reflect"
	"testing"

	"github.com/cgrates/cgrates/utils"
)

func TestSharedSetGet(t *testing.T) {
	id := "TEST_SG100"
	sg := &SharedGroup{
		Id: id,
		AccountParameters: map[string]*SharingParameters{
			"test": &SharingParameters{Strategy: STRATEGY_HIGHEST},
		},
		MemberIds: utils.NewStringMap("1", "2", "3"),
	}
	err := ratingStorage.SetSharedGroup(sg, utils.NonTransactional)
	if err != nil {
		t.Error("Error storing Shared groudp: ", err)
	}
	received, err := ratingStorage.GetSharedGroup(id, true, utils.NonTransactional)
	if err != nil || received == nil || !reflect.DeepEqual(sg, received) {
		t.Error("Error getting shared group: ", err, received)
	}
	received, err = ratingStorage.GetSharedGroup(id, false, utils.NonTransactional)
	if err != nil || received == nil || !reflect.DeepEqual(sg, received) {
		t.Error("Error getting cached shared group: ", err, received)
	}

}

func TestSharedPopBalanceByStrategyLow(t *testing.T) {
	bc := Balances{
		&Balance{Value: 2.0},
		&Balance{Uuid: "uuuu", Value: 1.0, account: &Account{ID: "test"}},
		&Balance{Value: 3.0},
	}
	sg := &SharedGroup{AccountParameters: map[string]*SharingParameters{
		"test": &SharingParameters{Strategy: STRATEGY_LOWEST}},
	}
	sbc := sg.SortBalancesByStrategy(bc[1], bc)
	if len(sbc) != 3 ||
		sbc[0].Value != 1.0 ||
		sbc[1].Value != 2.0 {
		t.Error("Error sorting balance chain: ", sbc[0].GetValue())
	}
}

func TestSharedPopBalanceByStrategyHigh(t *testing.T) {
	bc := Balances{
		&Balance{Uuid: "uuuu", Value: 2.0, account: &Account{ID: "test"}},
		&Balance{Value: 1.0},
		&Balance{Value: 3.0},
	}
	sg := &SharedGroup{AccountParameters: map[string]*SharingParameters{
		"test": &SharingParameters{Strategy: STRATEGY_HIGHEST}},
	}
	sbc := sg.SortBalancesByStrategy(bc[0], bc)
	if len(sbc) != 3 ||
		sbc[0].Value != 3.0 ||
		sbc[1].Value != 2.0 {
		t.Error("Error sorting balance chain: ", sbc)
	}
}

func TestSharedPopBalanceByStrategyMineHigh(t *testing.T) {
	bc := Balances{
		&Balance{Uuid: "uuuu", Value: 2.0, account: &Account{ID: "test"}},
		&Balance{Value: 1.0},
		&Balance{Value: 3.0},
	}
	sg := &SharedGroup{AccountParameters: map[string]*SharingParameters{
		"test": &SharingParameters{Strategy: STRATEGY_MINE_HIGHEST}},
	}
	sbc := sg.SortBalancesByStrategy(bc[0], bc)
	if len(sbc) != 3 ||
		sbc[0].Value != 2.0 ||
		sbc[1].Value != 3.0 {
		t.Error("Error sorting balance chain: ", sbc)
	}
}

/*func TestSharedPopBalanceByStrategyRandomHigh(t *testing.T) {
	bc := Balances{
		&Balance{Uuid: "uuuu", Value: 2.0, account: &Account{Id: "test"}},
		&Balance{Value: 1.0},
		&Balance{Value: 3.0},
	}
	sg := &SharedGroup{AccountParameters: map[string]*SharingParameters{
		"test": &SharingParameters{Strategy: STRATEGY_RANDOM}},
	}
	x := bc[0]
	sbc := sg.SortBalancesByStrategy(bc[0], bc)
	firstTest := (sbc[0].Uuid == x.Uuid)
	sbc = sg.SortBalancesByStrategy(bc[0], bc)
	secondTest := (sbc[0].Uuid == x.Uuid)
	sbc = sg.SortBalancesByStrategy(bc[0], bc)
	thirdTest := (sbc[0].Uuid == x.Uuid)
	sbc = sg.SortBalancesByStrategy(bc[0], bc)
	fourthTest := (sbc[0].Uuid == x.Uuid)
	if firstTest && secondTest && thirdTest && fourthTest {
		t.Error("Something is wrong with balance randomizer")
	}
}*/
