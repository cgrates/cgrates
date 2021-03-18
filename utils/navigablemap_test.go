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

func TestNavigableMapString(t *testing.T) {
	var nm NMInterface = NavigableMap{"Field1": NewNMData("1001")}
	expected := `{"Field1":1001}`
	if rply := nm.String(); rply != expected {
		t.Errorf("Expected %q ,received: %q", expected, rply)
	}
	nm = NavigableMap{}
	expected = `{}`
	if rply := nm.String(); rply != expected {
		t.Errorf("Expected %q ,received: %q", expected, rply)
	}
}

func TestNavigableMapInterface(t *testing.T) {
	nm := NavigableMap{"Field1": NewNMData("1001"), "Field2": NewNMData("1003")}
	expected := NavigableMap{"Field1": NewNMData("1001"), "Field2": NewNMData("1003")}
	if rply := nm.Interface(); !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expected %s ,received: %s", ToJSON(expected), ToJSON(rply))
	}
}

func TestNavigableMapField(t *testing.T) {
	nm := NavigableMap{}
	if _, err := nm.Field(PathItems{{}}); err != ErrNotFound {
		t.Error(err)
	}
	nm = NavigableMap{
		"Field1": NewNMData("1001"),
		"Field2": NewNMData("1003"),
		"Field3": NavigableMap{"Field4": NewNMData("Val")},
		"Field5": &NMSlice{NewNMData(10), NewNMData(101)},
	}
	if _, err := nm.Field(nil); err != ErrWrongPath {
		t.Error(err)
	}
	if _, err := nm.Field(PathItems{{Field: "NaN"}}); err != ErrNotFound {
		t.Error(err)
	}

	if val, err := nm.Field(PathItems{{Field: "Field1"}}); err != nil {
		t.Error(err)
	} else if val.Interface() != "1001" {
		t.Errorf("Expected %q ,received: %q", "1001", val.Interface())
	}

	if _, err := nm.Field(PathItems{{Field: "Field1", Index: []string{"0"}}}); err != ErrNotFound {
		t.Error(err)
	}
	if val, err := nm.Field(PathItems{{Field: "Field5", Index: []string{"0"}}}); err != nil {
		t.Error(err)
	} else if val.Interface() != 10 {
		t.Errorf("Expected %q ,received: %q", 10, val.Interface())
	}
	if _, err := nm.Field(PathItems{{Field: "Field3", Index: []string{"0"}}}); err != ErrNotFound {
		t.Error(err)
	}
	if val, err := nm.Field(PathItems{{Field: "Field3"}, {Field: "Field4"}}); err != nil {
		t.Error(err)
	} else if val.Interface() != "Val" {
		t.Errorf("Expected %q ,received: %q", "Val", val.Interface())
	}
}

func TestNavigableMapSet(t *testing.T) {
	nm := NavigableMap{}
	if _, err := nm.Set(nil, nil); err != ErrWrongPath {
		t.Error(err)
	}
	if _, err := nm.Set(PathItems{{Field: "Field1", Index: []string{"10"}}}, NewNMData("1001")); err != ErrWrongPath {
		t.Error(err)
	}
	expected := NavigableMap{"Field1": &NMSlice{NewNMData("1001")}}
	if _, err := nm.Set(PathItems{{Field: "Field1", Index: []string{"0"}}}, NewNMData("1001")); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(nm, expected) {
		t.Errorf("Expected %s ,received: %s", expected, nm)
	}
	expected = NavigableMap{
		"Field1": &NMSlice{NewNMData("1001")},
		"Field2": NewNMData("1002"),
	}
	if _, err := nm.Set(PathItems{{Field: "Field2"}}, NewNMData("1002")); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(nm, expected) {
		t.Errorf("Expected %s ,received: %s", expected, nm)
	}
	if _, err := nm.Set(PathItems{{Field: "Field2", Index: []string{"1"}}}, NewNMData("1003")); err != ErrWrongPath {
		t.Error(err)
	}
	expected = NavigableMap{
		"Field1": &NMSlice{NewNMData("1001"), NewNMData("1003")},
		"Field2": NewNMData("1002"),
	}
	if _, err := nm.Set(PathItems{{Field: "Field1", Index: []string{"1"}}}, NewNMData("1003")); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(nm, expected) {
		t.Errorf("Expected %s ,received: %s", expected, nm)
	}
	expected = NavigableMap{
		"Field1": &NMSlice{NewNMData("1001"), NewNMData("1003")},
		"Field2": NewNMData("1004"),
	}
	if _, err := nm.Set(PathItems{{Field: "Field2"}}, NewNMData("1004")); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(nm, expected) {
		t.Errorf("Expected %s ,received: %s", expected, nm)
	}

	if _, err := nm.Set(PathItems{{Field: "Field3", Index: []string{"10"}}, {}}, NewNMData("1001")); err != ErrWrongPath {
		t.Error(err)
	}
	expected = NavigableMap{
		"Field1": &NMSlice{NewNMData("1001"), NewNMData("1003")},
		"Field2": NewNMData("1004"),
		"Field3": &NMSlice{NavigableMap{"Field4": NewNMData("1005")}},
	}
	if _, err := nm.Set(PathItems{{Field: "Field3", Index: []string{"0"}}, {Field: "Field4"}}, NewNMData("1005")); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(nm, expected) {
		t.Errorf("Expected %s ,received: %s", expected, nm)
	}

	if _, err := nm.Set(PathItems{{Field: "Field5"}, {Field: "Field6", Index: []string{"10"}}}, NewNMData("1006")); err != ErrWrongPath {
		t.Error(err)
	}

	expected = NavigableMap{
		"Field1": &NMSlice{NewNMData("1001"), NewNMData("1003")},
		"Field2": NewNMData("1004"),
		"Field3": &NMSlice{NavigableMap{"Field4": NewNMData("1005")}},
		"Field5": NavigableMap{"Field6": NewNMData("1006")},
	}
	if _, err := nm.Set(PathItems{{Field: "Field5"}, {Field: "Field6"}}, NewNMData("1006")); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(nm, expected) {
		t.Errorf("Expected %s ,received: %s", expected, nm)
	}

	if _, err := nm.Set(PathItems{{Field: "Field2", Index: []string{"0"}}, {}}, NewNMData("1006")); err != ErrWrongPath {
		t.Error(err)
	}
	expected = NavigableMap{
		"Field1": &NMSlice{NewNMData("1001"), NewNMData("1003"), NavigableMap{"Field6": NewNMData("1006")}},
		"Field2": NewNMData("1004"),
		"Field3": &NMSlice{NavigableMap{"Field4": NewNMData("1005")}},
		"Field5": NavigableMap{"Field6": NewNMData("1006")},
	}
	if _, err := nm.Set(PathItems{{Field: "Field1", Index: []string{"2"}}, {Field: "Field6"}}, NewNMData("1006")); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(nm, expected) {
		t.Errorf("Expected %s ,received: %s", expected, nm)
	}
	if _, err := nm.Set(PathItems{{Field: "Field2"}, {}}, NewNMData("1006")); err != ErrWrongPath {
		t.Error(err)
	}

	expected = NavigableMap{
		"Field1": &NMSlice{NewNMData("1001"), NewNMData("1003"), NavigableMap{"Field6": NewNMData("1006")}},
		"Field2": NewNMData("1004"),
		"Field3": &NMSlice{NavigableMap{"Field4": NewNMData("1005")}},
		"Field5": NavigableMap{"Field6": NewNMData("1007")},
	}
	if _, err := nm.Set(PathItems{{Field: "Field5"}, {Field: "Field6"}}, NewNMData("1007")); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(nm, expected) {
		t.Errorf("Expected %s ,received: %s", expected, nm)
	}

	expected = NavigableMap{
		"Field1": &NMSlice{NewNMData("1001"), NewNMData("1003"), NavigableMap{"Field6": NewNMData("1006")}},
		"Field2": NewNMData("1004"),
		"Field3": &NMSlice{NavigableMap{"Field4": NewNMData("1005")}},
		"Field5": NavigableMap{"Field6": NewNMData("1007")},
		"Field7": NavigableMap{"Field8": &NMSlice{NavigableMap{"Field9": NewNMData("1005")}}},
	}
	if _, err := nm.Set(PathItems{{Field: "Field7", Index: []string{"Field8", "0"}}, {Field: "Field9"}}, NewNMData("1005")); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(nm, expected) {
		t.Errorf("Expected %s ,received: %s", expected, nm)
	}
}

func TestNavigableMapType(t *testing.T) {
	var nm NMInterface = NavigableMap{}
	if nm.Type() != NMMapType {
		t.Errorf("Expected %v ,received: %v", NMMapType, nm.Type())
	}
}

func TestNavigableMapEmpty(t *testing.T) {
	var nm NMInterface = NavigableMap{}
	if !nm.Empty() {
		t.Error("Expected empty type")
	}
	nm = NavigableMap{"Field1": NewNMData("1001")}
	if nm.Empty() {
		t.Error("Expected not empty type")
	}
}

func TestNavigableMapLen(t *testing.T) {
	var nm NMInterface = NavigableMap{}
	if rply := nm.Len(); rply != 0 {
		t.Errorf("Expected 0 ,received: %v", rply)
	}
	nm = NavigableMap{"Field1": NewNMData("1001")}
	if rply := nm.Len(); rply != 1 {
		t.Errorf("Expected 1 ,received: %v", rply)
	}
}

func TestNavigableMapRemove(t *testing.T) {
	nm := NavigableMap{
		"Field1": NewNMData("1001"),
		"Field2": NewNMData("1003"),
		"Field3": NavigableMap{"Field4": NewNMData("Val")},
		"Field5": &NMSlice{NewNMData(10), NewNMData(101)},
	}
	if err := nm.Remove(nil); err != ErrWrongPath {
		t.Error(err)
	}
	if err := nm.Remove(PathItems{}); err != ErrWrongPath {
		t.Error(err)
	}
	if err := nm.Remove(PathItems{{Field: "field"}}); err != nil {
		t.Error(err)
	}

	if err := nm.Remove(PathItems{{Index: []string{"-1"}}, {}}); err != nil {
		t.Error(err)
	}
	expected := NavigableMap{
		"Field1": NewNMData("1001"),
		"Field2": NewNMData("1003"),
		"Field3": NavigableMap{"Field4": NewNMData("Val")},
		"Field5": &NMSlice{NewNMData(10), NewNMData(101)},
	}
	if !reflect.DeepEqual(nm, expected) {
		t.Errorf("Expected %s ,received: %s", expected, nm)
	}

	expected = NavigableMap{
		"Field1": NewNMData("1001"),
		"Field3": NavigableMap{"Field4": NewNMData("Val")},
		"Field5": &NMSlice{NewNMData(10), NewNMData(101)},
	}

	if err := nm.Remove(PathItems{{Field: "Field2"}}); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(nm, expected) {
		t.Errorf("Expected %s ,received: %s", expected, nm)
	}

	expected = NavigableMap{
		"Field1": NewNMData("1001"),
		"Field3": NavigableMap{"Field4": NewNMData("Val")},
		"Field5": &NMSlice{NewNMData(101)},
	}

	if err := nm.Remove(PathItems{{Field: "Field5", Index: []string{"0"}}}); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(nm, expected) {
		t.Errorf("Expected %s ,received: %s", expected, nm)
	}

	if err := nm.Remove(PathItems{{Field: "Field1", Index: []string{"0"}}, {}}); err != ErrWrongPath {
		t.Error(err)
	}

	expected = NavigableMap{
		"Field1": NewNMData("1001"),
		"Field3": NavigableMap{"Field4": NewNMData("Val")},
	}

	if err := nm.Remove(PathItems{{Field: "Field5", Index: []string{"0"}}}); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(nm, expected) {
		t.Errorf("Expected %s ,received: %s", expected, nm)
	}
	nm = NavigableMap{
		"Field1": NewNMData("1001"),
		"Field2": NewNMData("1003"),
		"Field3": NavigableMap{"Field4": NewNMData("Val")},
		"Field5": &NMSlice{NavigableMap{"Field42": NewNMData("Val2")}},
	}
	if err := nm.Remove(PathItems{{Field: "Field5", Index: []string{"0"}}, {Field: "Field42", Index: []string{"0"}}}); err != ErrWrongPath {
		t.Error(err)
	}

	expected = NavigableMap{
		"Field1": NewNMData("1001"),
		"Field2": NewNMData("1003"),
		"Field3": NavigableMap{"Field4": NewNMData("Val")},
	}

	if err := nm.Remove(PathItems{{Field: "Field5", Index: []string{"0"}}, {Field: "Field42"}}); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(nm, expected) {
		t.Errorf("Expected %s ,received: %s", expected, nm)
	}

	if err := nm.Remove(PathItems{{Field: "Field1"}, {}}); err != ErrWrongPath {
		t.Error(err)
	}
	if err := nm.Remove(PathItems{{Field: "Field3"}, {Field: "Field4", Index: []string{"0"}}}); err != ErrWrongPath {
		t.Error(err)
	}
	expected = NavigableMap{
		"Field1": NewNMData("1001"),
		"Field2": NewNMData("1003"),
	}
	if err := nm.Remove(PathItems{{Field: "Field3"}, {Field: "Field4"}}); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(nm, expected) {
		t.Errorf("Expected %s ,received: %s", expected, nm)
	}

	nm = NavigableMap{
		"Field1": NewNMData("1001"),
		"Field2": NewNMData("1003"),
		"Field3": NavigableMap{"Field4": &NMSlice{NewNMData("Val")}},
	}
	expected = NavigableMap{
		"Field1": NewNMData("1001"),
		"Field2": NewNMData("1003"),
	}
	if err := nm.Remove(PathItems{{Field: "Field3", Index: []string{"Field4", "0"}}}); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(nm, expected) {
		t.Errorf("Expected %s ,received: %s", expected, nm)
	}

}

func TestNavigableMapGetSet(t *testing.T) {
	var nm NMInterface = NavigableMap{
		"Field1": NewNMData(10),
		"Field2": &NMSlice{
			NewNMData("1001"),
			NavigableMap{
				"Account": &NMSlice{NewNMData(10), NewNMData(11)},
			},
		},
		"Field3": NavigableMap{
			"Field4": NavigableMap{
				"Field5": NewNMData(5),
			},
		},
	}
	path := PathItems{{Field: "Field1"}}
	if val, err := nm.Field(path); err != nil {
		t.Error(err)
	} else if val.Interface() != 10 {
		t.Errorf("Expected %q ,received: %q", 10, val.Interface())
	}

	path = PathItems{{Field: "Field3"}, {Field: "Field4"}, {Field: "Field5"}}
	if val, err := nm.Field(path); err != nil {
		t.Error(err)
	} else if val.Interface() != 5 {
		t.Errorf("Expected %q ,received: %q", 5, val.Interface())
	}

	path = PathItems{{Field: "Field2", Index: []string{"2"}}}
	if _, err := nm.Set(path, NewNMData("500")); err != nil {
		t.Error(err)
	}
	if val, err := nm.Field(path); err != nil {
		t.Error(err)
	} else if val.Interface() != "500" {
		t.Errorf("Expected %q ,received: %q", "500", val.Interface())
	}

	path = PathItems{{Field: "Field2", Index: []string{"1"}}, {Field: "Account"}}
	if _, err := nm.Set(path, NewNMData("5")); err != nil {
		t.Error(err)
	}
	path = PathItems{{Field: "Field2", Index: []string{"1"}}, {Field: "Account"}}
	if val, err := nm.Field(path); err != nil {
		t.Error(err)
	} else if val.Interface() != "5" {
		t.Errorf("Expected %q ,received: %q", "5", val.Interface())
	}
	path = PathItems{{Field: "Field2", Index: []string{"1"}}, {Field: "Account", Index: []string{"0"}}}
	if _, err := nm.Field(path); err != ErrNotFound {
		t.Error(err)
	}
}

func TestNavigableMapFieldAsInterface(t *testing.T) {
	nm := NavigableMap{
		"Field1": NewNMData("1001"),
		"Field2": NewNMData("1003"),
		"Field3": NavigableMap{"Field4": NewNMData("Val")},
		"Field5": &NMSlice{NewNMData(10), NewNMData(101)},
	}
	if _, err := nm.FieldAsInterface(nil); err != ErrWrongPath {
		t.Error(err)
	}

	if val, err := nm.FieldAsInterface([]string{"Field3", "Field4"}); err != nil {
		t.Error(err)
	} else if val != "Val" {
		t.Errorf("Expected %q ,received: %q", "Val", val)
	}

	if val, err := nm.FieldAsInterface([]string{"Field5[0]"}); err != nil {
		t.Error(err)
	} else if val != 10 {
		t.Errorf("Expected %q ,received: %q", 10, val)
	}
}

func TestNavigableMapFieldAsString(t *testing.T) {
	nm := NavigableMap{
		"Field1": NewNMData("1001"),
		"Field2": NewNMData("1003"),
		"Field3": NavigableMap{"Field4": NewNMData("Val")},
		"Field5": &NMSlice{NewNMData(10), NewNMData(101)},
	}
	if _, err := nm.FieldAsString(nil); err != ErrWrongPath {
		t.Error(err)
	}

	if val, err := nm.FieldAsString([]string{"Field3", "Field4"}); err != nil {
		t.Error(err)
	} else if val != "Val" {
		t.Errorf("Expected %q ,received: %q", "Val", val)
	}

	if val, err := nm.FieldAsString([]string{"Field5[0]"}); err != nil {
		t.Error(err)
	} else if val != "10" {
		t.Errorf("Expected %q ,received: %q", "10", val)
	}
}

func TestNavigableMapRemote(t *testing.T) {
	nm := NavigableMap{"Field1": NewNMData("1001")}
	eOut := LocalAddr()
	if rcv := nm.RemoteHost(); !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", eOut, rcv)
	}
}
