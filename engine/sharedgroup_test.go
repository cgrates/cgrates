/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2013 ITsysCOM

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
)

func TestSharedGroupGetMembersExcept(t *testing.T) {
	sg := &SharedGroup{
		Members: []string{"1", "2", "3"},
	}
	a1 := sg.GetMembersExceptUser("1")
	a2 := sg.GetMembersExceptUser("2")
	a3 := sg.GetMembersExceptUser("3")
	if !reflect.DeepEqual(a1, []string{"3", "2"}) ||
		!reflect.DeepEqual(a2, []string{"1", "3"}) ||
		!reflect.DeepEqual(a3, []string{"1", "2"}) {
		t.Error("Error getting shared group members: ", a1, a2, a3)
	}

}

func TestSharedPopBalanceByStrategyLow(t *testing.T) {
	bc := BalanceChain{
		&Balance{Value: 2.0},
		&Balance{Value: 1.0},
		&Balance{Value: 3.0},
	}
	sg := &SharedGroup{Strategy: STRATEGY_LOWEST_FIRST}
	b := sg.PopBalanceByStrategy(&bc)
	if b.Value != 1.0 {
		t.Error("Error popping the right balance according to strategy: ", b, bc)
	}
	if len(bc) != 2 ||
		bc[0].Value != 2.0 ||
		bc[1].Value != 3.0 {
		t.Error("Error removing balance from chain: ", bc)
	}
}

func TestSharedPopBalanceByStrategyHigh(t *testing.T) {
	bc := BalanceChain{
		&Balance{Value: 2.0},
		&Balance{Value: 1.0},
		&Balance{Value: 3.0},
	}
	sg := &SharedGroup{Strategy: STRATEGY_HIGHEST_FIRST}
	b := sg.PopBalanceByStrategy(&bc)
	if b.Value != 3.0 {
		t.Error("Error popping the right balance according to strategy: ", b, bc)
	}
	if len(bc) != 2 ||
		bc[0].Value != 2.0 ||
		bc[1].Value != 1.0 {
		t.Error("Error removing balance from chain: ", bc)
	}
}
