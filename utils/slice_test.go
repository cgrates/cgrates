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
package utils

import (
	"reflect"
	"sort"
	"testing"
)

func TestHasPrefixSlice(t *testing.T) {
	if !HasPrefixSlice([]string{"1", "2", "3", "4", "5"}, "123") {
		t.Error("Expecting: true, received: false")
	}
	if HasPrefixSlice([]string{"1", "2", "3", "4", "5"}, "689") {
		t.Error("Expecting: true, received: false")
	}
}

func TestPrefixSliceItems(t *testing.T) {
	rcv := PrefixSliceItems("*", []string{"1", "2", "3", "", "5"})
	sort.Strings(rcv)
	eOut := []string{"*1", "*2", "*3", "*5"}
	if !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", eOut, rcv)
	}
}

func TestSliceStringToIface(t *testing.T) {
	exp := []any{"*default", "ToR", "*voice"}
	if rply := SliceStringToIface([]string{"*default", "ToR", "*voice"}); !reflect.DeepEqual(exp, rply) {
		t.Errorf("Expected: %s ,received: %s", ToJSON(exp), ToJSON(rply))
	}
}
