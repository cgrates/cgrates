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
	"testing"
)

func TestNMSliceString(t *testing.T) {
	var nm NMInterface = &NMSlice{NewNMData("1001"), NewNMData("1003")}
	expected := "[1001,1003]"
	if rply := nm.String(); rply != expected {
		t.Errorf("Expected %q ,received: %q", expected, rply)
	}
	nm = &NMSlice{}
	expected = `[]`
	if rply := nm.String(); rply != expected {
		t.Errorf("Expected %q ,received: %q", expected, rply)
	}
}

func TestNMSliceInterface(t *testing.T) {
	nm := &NMSlice{NewNMData("1001"), NewNMData("1003")}
	expected := &NMSlice{NewNMData("1001"), NewNMData("1003")}
	if rply := nm.Interface(); !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expected %s ,received: %s", ToJSON(expected), ToJSON(rply))
	}
}

func TestNMSliceField(t *testing.T) {
	nm := &NMSlice{}
	if _, err := nm.Field(PathItems{{}}); err != ErrNotFound {
		t.Error(err)
	}
	nm = &NMSlice{
		NewNMData("1001"),
		NewNMData("1003"),
		&NavigableMap2{"Field1": NewNMData("Val")},
	}
	if _, err := nm.Field(PathItems{{}}); err != ErrNotFound {
		t.Error(err)
	}
	if _, err := nm.Field(PathItems{{Index: IntPointer(4)}}); err != ErrNotFound {
		t.Error(err)
	}
	if _, err := nm.Field(nil); err != ErrWrongPath {
		t.Error(err)
	}
	if val, err := nm.Field(PathItems{{Field: "None", Index: IntPointer(-1)}, {Field: "Field1"}}); err != nil {
		t.Error(err)
	} else if val.Interface() != "Val" {
		t.Errorf("Expected %q ,received: %q", "Val", val.Interface())
	}
	if val, err := nm.Field(PathItems{{Field: "1234", Index: IntPointer(1)}}); err != nil {
		t.Error(err)
	} else if val.Interface() != "1003" {
		t.Errorf("Expected %q ,received: %q", "Val", val.Interface())
	}
}

func TestNMSliceSet(t *testing.T) {
	nm := &NMSlice{}
	if err := nm.Set(nil, nil); err != ErrWrongPath {
		t.Error(err)
	}
	expected := &NMSlice{NewNMData("1001")}
	if err := nm.Set(PathItems{{Field: "1234", Index: IntPointer(0)}}, NewNMData("1001")); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(nm, expected) {
		t.Errorf("Expected %s ,received: %s", expected, nm)
	}
	if err := nm.Set(PathItems{{Field: "1234", Index: IntPointer(1)}, {Field: "Field1", Index: IntPointer(1)}},
		NewNMData("1001")); err != ErrWrongPath {
		t.Error(err)
	}
	expected = &NMSlice{NewNMData("1001"), NavigableMap2{"Field1": NewNMData("1001")}}
	if err := nm.Set(PathItems{{Field: "1234", Index: IntPointer(1)}, {Field: "Field1"}},
		NewNMData("1001")); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(nm, expected) {
		t.Errorf("Expected %s ,received: %s", expected, nm)
	}
	expected = &NMSlice{NewNMData("1001"), NewNMData("1001")}
	if err := nm.Set(PathItems{{Field: "1234", Index: IntPointer(-1)}}, NewNMData("1001")); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(nm, expected) {
		t.Errorf("Expected %s ,received: %s", expected, nm)
	}

	nm = &NMSlice{&NMSlice{}}
	if err := nm.Set(PathItems{{Field: "1234", Index: IntPointer(0)}, {}}, NewNMData("1001")); err != ErrWrongPath {
		t.Error(err)
	}
}

func TestNMSliceType(t *testing.T) {
	var nm NMInterface = &NMSlice{}
	if nm.Type() != NMSliceType {
		t.Errorf("Expected %v ,received: %v", NMSliceType, nm.Type())
	}
}

func TestNMSliceEmpty(t *testing.T) {
	var nm NMInterface = &NMSlice{}
	if !nm.Empty() {
		t.Error("Expected empty type")
	}
	nm = &NMSlice{NewNMData("1001")}
	if nm.Empty() {
		t.Error("Expected not empty type")
	}
}

func TestNMSliceLen(t *testing.T) {
	var nm NMInterface = &NMSlice{}
	if rply := nm.Len(); rply != 0 {
		t.Errorf("Expected 0 ,received: %v", rply)
	}
	nm = &NMSlice{NewNMData("1001")}
	if rply := nm.Len(); rply != 1 {
		t.Errorf("Expected 1 ,received: %v", rply)
	}
}

func TestNMSliceGetField(t *testing.T) {
	nm := &NMSlice{}
	if _, err := nm.GetField(PathItem{}); err != ErrNotFound {
		t.Error(err)
	}
	nm = &NMSlice{
		NewNMData("1001"),
		NewNMData("1003"),
	}
	if _, err := nm.GetField(PathItem{}); err != ErrNotFound {
		t.Error(err)
	}
	if _, err := nm.GetField(PathItem{Index: IntPointer(4)}); err != ErrNotFound {
		t.Error(err)
	}
	// if _, err := nm.GetField(nil); err != ErrWrongPath {
	// 	t.Error(err)
	// }
	if val, err := nm.GetField(PathItem{Field: "None", Index: IntPointer(-1)}); err != nil {
		t.Error(err)
	} else if val.Interface() != "1003" {
		t.Errorf("Expected %q ,received: %q", "Val", val.Interface())
	}
	if val, err := nm.GetField(PathItem{Field: "1234", Index: IntPointer(0)}); err != nil {
		t.Error(err)
	} else if val.Interface() != "1001" {
		t.Errorf("Expected %q ,received: %q", "Val", val.Interface())
	}
}

func TestNMSliceSetField(t *testing.T) {
	nm := &NMSlice{}
	if err := nm.SetField(PathItem{}, NewNMData(10)); err != ErrWrongPath {
		t.Error(err)
	}
	// if err := nm.SetField(nil, nil); err != ErrWrongPath {
	// 	t.Error(err)
	// }
	expected := &NMSlice{NewNMData(10)}
	if err := nm.SetField(PathItem{Index: IntPointer(0)}, NewNMData(10)); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(nm, expected) {
		t.Errorf("Expected %s ,received: %s", expected, nm)
	}

	expected = &NMSlice{NewNMData(10), NewNMData(11)}
	if err := nm.SetField(PathItem{Index: IntPointer(1)}, NewNMData(11)); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(nm, expected) {
		t.Errorf("Expected %s ,received: %s", expected, nm)
	}

	expected = &NMSlice{NewNMData(10), NewNMData(20)}
	if err := nm.SetField(PathItem{Index: IntPointer(-1)}, NewNMData(20)); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(nm, expected) {
		t.Errorf("Expected %s ,received: %s", expected, nm)
	}

	expected = &NMSlice{NewNMData(0), NewNMData(20)}
	if err := nm.SetField(PathItem{Index: IntPointer(0)}, NewNMData(0)); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(nm, expected) {
		t.Errorf("Expected %s ,received: %s", expected, nm)
	}

	if err := nm.SetField(PathItem{Index: IntPointer(5)}, NewNMData(0)); err != ErrWrongPath {
		t.Error(err)
	}
}

func TestNMSliceRemove(t *testing.T) {
	nm := &NMSlice{
		NewNMData("1001"),
		NewNMData("1003"),
		&NavigableMap2{"Field1": NewNMData("Val")},
		&NMSlice{},
	}
	if err := nm.Remove(nil); err != ErrWrongPath {
		t.Error(err)
	}
	if err := nm.Remove(PathItems{}); err != ErrWrongPath {
		t.Error(err)
	}
	if err := nm.Remove(PathItems{{Field: "field"}}); err != ErrWrongPath {
		t.Error(err)
	}

	if err := nm.Remove(PathItems{{Index: IntPointer(-1)}, {}}); err != ErrWrongPath {
		t.Error(err)
	}
	expected := &NMSlice{
		NewNMData("1001"),
		NewNMData("1003"),
		&NavigableMap2{"Field1": NewNMData("Val")},
	}

	if err := nm.Remove(PathItems{{Index: IntPointer(-1)}}); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(nm, expected) {
		t.Errorf("Expected %s ,received: %s", expected, nm)
	}

	if err := nm.Remove(PathItems{{Index: IntPointer(1)}, {}}); err != ErrWrongPath {
		t.Error(err)
	}

	expected = &NMSlice{
		NewNMData("1001"),
		&NavigableMap2{"Field1": NewNMData("Val")},
	}
	if err := nm.Remove(PathItems{{Index: IntPointer(1)}}); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(nm, expected) {
		t.Errorf("Expected %s ,received: %s", expected, nm)
	}

	if err := nm.Remove(PathItems{{Index: IntPointer(10)}}); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(nm, expected) {
		t.Errorf("Expected %s ,received: %s", expected, nm)
	}

	if err := nm.Remove(PathItems{{Index: IntPointer(1)}, {Field: "Field1", Index: IntPointer(1)}}); err != ErrWrongPath {
		t.Error(err)
	}
	expected = &NMSlice{
		NewNMData("1001"),
	}
	if err := nm.Remove(PathItems{{Index: IntPointer(1)}, {Field: "Field1"}}); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(nm, expected) {
		t.Errorf("Expected %s ,received: %s", expected, nm)
	}

}
