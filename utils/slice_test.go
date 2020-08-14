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
func TestSliceWithoutMember(t *testing.T) {
	rcv := SliceWithoutMember([]string{"1", "2", "3", "4", "5"}, "5")
	sort.Strings(rcv)
	eOut := []string{"1", "2", "3", "4"}
	if !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", eOut, rcv)
	}
	rcv = SliceWithoutMember([]string{"1", "2", "3", "4", "5"}, "6")
	sort.Strings(rcv)
	eOut = []string{"1", "2", "3", "4", "5"}
	if !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", eOut, rcv)
	}
}

func TestSliceMemberHasPrefix(t *testing.T) {
	if !SliceMemberHasPrefix([]string{"1", "*2", "3", "4", "5"}, "*") {
		t.Error("Expecting: true, received: false")
	}
	if SliceMemberHasPrefix([]string{"1", "2", "3", "4", "5"}, "*") {
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

func TestAvg(t *testing.T) {
	values := []float64{1, 2, 3}
	result := Avg(values)
	expected := 2.0
	if expected != result {
		t.Errorf("Wrong Avg: expected %v got %v", expected, result)
	}
}

func TestAvgEmpty(t *testing.T) {
	values := []float64{}
	result := Avg(values)
	expected := 0.0
	if expected != result {
		t.Errorf("Wrong Avg: expected %v got %v", expected, result)
	}
}

func TestStripSlicePrefix(t *testing.T) {
	eSlc := make([]string, 0)
	if retSlc := StripSlicePrefix([]string{}, 2); !reflect.DeepEqual(eSlc, retSlc) {
		t.Errorf("expecting: %+v, received: %+v", eSlc, retSlc)
	}
	eSlc = []string{"1", "2"}
	if retSlc := StripSlicePrefix([]string{"0", "1", "2"}, 1); !reflect.DeepEqual(eSlc, retSlc) {
		t.Errorf("expecting: %+v, received: %+v", eSlc, retSlc)
	}
	eSlc = []string{}
	if retSlc := StripSlicePrefix([]string{"0", "1", "2"}, 3); !reflect.DeepEqual(eSlc, retSlc) {
		t.Errorf("expecting: %+v, received: %+v", eSlc, retSlc)
	}
}

func TestSliceStringToIface(t *testing.T) {
	args := []string{"*default", "ToR", "*voice"}
	exp := []interface{}{"*default", "ToR", "*voice"}
	if rply := SliceStringToIface(args); !reflect.DeepEqual(exp, rply) {
		t.Errorf("Expected: %s ,received: %s", ToJSON(exp), ToJSON(rply))
	}
}
