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
package utils

import (
	"reflect"
	"sort"
	"testing"
)

func TestIsSliceMember(t *testing.T) {
	if !IsSliceMember([]string{"1", "2", "3", "4", "5"}, "5") {
		t.Error("Expecting: true, received: false")
	}
	if IsSliceMember([]string{"1", "2", "3", "4", "5"}, "6") {
		t.Error("Expecting: true, received: false")
	}
}

func TestSliceHasMember(t *testing.T) {
	if !SliceHasMember([]string{"1", "2", "3", "4", "5"}, "5") {
		t.Error("Expecting: true, received: false")
	}
	if SliceHasMember([]string{"1", "2", "3", "4", "5"}, "6") {
		t.Error("Expecting: true, received: false")
	}
}

func TestFlaot64SliceHasMember(t *testing.T) {
	if !Float64SliceHasMember([]float64{1, 2, 3, 4, 5}, 5) {
		t.Error("Expecting: true, received: false")
	}
	if Float64SliceHasMember([]float64{1, 2, 3, 4, 5}, 6) {
		t.Error("Expecting: true, received: false")
	}
}

func TestHasPrefixSlice(t *testing.T) {
	if !HasPrefixSlice([]string{"1", "2", "3", "4", "5"}, "123") {
		t.Error("Expecting: true, received: false")
	}
	if HasPrefixSlice([]string{"1", "2", "3", "4", "5"}, "689") {
		t.Error("Expecting: true, received: false")
	}
}

func TestPrefixSliceItems(t *testing.T) {
	rcv := PrefixSliceItems([]string{"1", "2", "3", "4", "5"}, "*")
	sort.Strings(rcv)
	eOut := []string{"*1", "*2", "*3", "*4", "*5"}
	if !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", eOut, rcv)
	}
}

func TestSliceStringToIface(t *testing.T) {
	exp := []interface{}{"*default", "ToR", "*voice"}
	if rply := SliceStringToIface([]string{"*default", "ToR", "*voice"}); !reflect.DeepEqual(exp, rply) {
		t.Errorf("Expected: %s ,received: %s", ToJSON(exp), ToJSON(rply))
	}
}

func TestSliceStringEqual(t *testing.T) {
	v1 := []string{"*V1field1", "*V1field2"}
	v2 := []string{"*V2field1"}

	//When the length don't match
	if rcv := SliceStringEqual(v1, v2); rcv {
		t.Error("The length should not match")
	}

	v2 = append(v2, "*V2field2")
	if rcv := SliceStringEqual(v1, v2); rcv {
		t.Error("The values should not match")
	}

	v2 = v1
	if rcv := SliceStringEqual(v1, v2); !rcv {
		t.Error("The slices should match")
	}
}
