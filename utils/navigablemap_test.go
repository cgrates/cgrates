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
	"math/rand"
	"reflect"
	"testing"
)

func TestOrderedNavigableMap(t *testing.T) {
	// var err error
	onm := NewOrderedNavigableMap()

	onm.Set(PathItems{{Field: "Field1"}}, NewNMInterface(10))
	onm.Set(PathItems{{Field: "Field2", Index: IntPointer(0)}}, NewNMInterface("1001"))
	onm.Set(PathItems{{Field: "Field2", Index: IntPointer(1)}, {Field: "Account", Index: IntPointer(0)}}, NewNMInterface(10))
	onm.Set(PathItems{{Field: "Field2", Index: IntPointer(1)}, {Field: "Account", Index: IntPointer(1)}}, NewNMInterface(11))
	onm.Set(PathItems{{Field: "Field2", Index: IntPointer(2)}}, NewNMInterface(111))
	onm.Set(PathItems{{Field: "Field3"}, {Field: "Field4"}, {Field: "Field5"}}, NewNMInterface(5))
	var expnm NM = NavigableMap2{
		"Field1": NewNMInterface(10),
		"Field2": &NMSlice{
			NewNMInterface("1001"),
			NavigableMap2{
				"Account": &NMSlice{NewNMInterface(10), NewNMInterface(11)},
			},
			NewNMInterface(111),
		},
		"Field3": NavigableMap2{
			"Field4": NavigableMap2{
				"Field5": NewNMInterface(5),
			},
		},
	}
	if onm.Empty() {
		t.Error("Expected not empty type")
	}
	if onm.Type() != NMMapType {
		t.Errorf("Expected %v ,received: %v", NMInterfaceType, onm.Type())
	}
	if !reflect.DeepEqual(expnm, onm.nm) {
		t.Errorf("Expected %s ,received: %s", expnm, onm.nm)
	}
	expOrder := []PathItems{
		PathItems{{Field: "Field1"}},
		PathItems{{Field: "Field2", Index: IntPointer(0)}},
		PathItems{{Field: "Field2", Index: IntPointer(1)}, {Field: "Account", Index: IntPointer(0)}},
		PathItems{{Field: "Field2", Index: IntPointer(1)}, {Field: "Account", Index: IntPointer(1)}},
		PathItems{{Field: "Field2", Index: IntPointer(2)}},
		PathItems{{Field: "Field3"}, {Field: "Field4"}, {Field: "Field5"}},
	}
	if !reflect.DeepEqual(expOrder, onm.order) {
		t.Errorf("Expected %s ,received: %s", expOrder, onm.order)
	}

	path := PathItems{{Field: "Field2"}}
	//
	// sliceDeNM
	exp := &NMSlice{NewNMInterface("500"), NewNMInterface("502")}
	if err := onm.Set(path, exp); err != nil {
		t.Error(err)
	}
	path = PathItems{{Field: "Field2"}}
	if val, err := onm.Field(path); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(val, exp) {
		t.Errorf("Expected %q ,received: %q", exp, val.Interface())
	}

	expOrder = []PathItems{
		PathItems{{Field: "Field1"}},
		PathItems{{Field: "Field3"}, {Field: "Field4"}, {Field: "Field5"}},
		PathItems{{Field: "Field2", Index: IntPointer(0)}},
		PathItems{{Field: "Field2", Index: IntPointer(1)}},
	}
	if !reflect.DeepEqual(expOrder, onm.order) {
		t.Errorf("Expected %s ,received: %s", expOrder, onm.order)
	}

	path = PathItems{{Field: "Field2", Index: IntPointer(0)}}
	if val, err := onm.Field(path); err != nil {
		t.Error(err)
	} else if val.Interface() != "500" {
		t.Errorf("Expected %q ,received: %q", "500", val.Interface())
	}
}

func TestOrderedNavigableMapString(t *testing.T) {
	var nm NM = &OrderedNavigableMap{nm: NavigableMap2{"Field1": NewNMInterface("1001")}}
	expected := `{"Field1":1001}`
	if rply := nm.String(); rply != expected {
		t.Errorf("Expected %q ,received: %q", expected, rply)
	}
	nm = &OrderedNavigableMap{nm: NavigableMap2{}}
	expected = `{}`
	if rply := nm.String(); rply != expected {
		t.Errorf("Expected %q ,received: %q", expected, rply)
	}
}

func TestOrderedNavigableMapInterface(t *testing.T) {
	nm := &OrderedNavigableMap{nm: NavigableMap2{"Field1": NewNMInterface("1001"), "Field2": NewNMInterface("1003")}}
	expected := NavigableMap2{"Field1": NewNMInterface("1001"), "Field2": NewNMInterface("1003")}
	if rply := nm.Interface(); !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expected %s ,received: %s", ToJSON(expected), ToJSON(rply))
	}
}

func TestOrderedNavigableMapField(t *testing.T) {
	nm := &OrderedNavigableMap{nm: NavigableMap2{}}
	if _, err := nm.Field(PathItems{{}}); err != ErrNotFound {
		t.Error(err)
	}
	nm = &OrderedNavigableMap{nm: NavigableMap2{
		"Field1": NewNMInterface("1001"),
		"Field2": NewNMInterface("1003"),
		"Field3": NavigableMap2{"Field4": NewNMInterface("Val")},
		"Field5": &NMSlice{NewNMInterface(10), NewNMInterface(101)},
	}}
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

	if _, err := nm.Field(PathItems{{Field: "Field1", Index: IntPointer(0)}}); err != ErrNotFound {
		t.Error(err)
	}
	if val, err := nm.Field(PathItems{{Field: "Field5", Index: IntPointer(0)}}); err != nil {
		t.Error(err)
	} else if val.Interface() != 10 {
		t.Errorf("Expected %q ,received: %q", 10, val.Interface())
	}
	if _, err := nm.Field(PathItems{{Field: "Field3", Index: IntPointer(0)}}); err != ErrNotFound {
		t.Error(err)
	}
	if val, err := nm.Field(PathItems{{Field: "Field3"}, {Field: "Field4"}}); err != nil {
		t.Error(err)
	} else if val.Interface() != "Val" {
		t.Errorf("Expected %q ,received: %q", "Val", val.Interface())
	}
}

func TestOrderedNavigableMapType(t *testing.T) {
	var nm NM = &OrderedNavigableMap{nm: NavigableMap2{}}
	if nm.Type() != NMMapType {
		t.Errorf("Expected %v ,received: %v", NMMapType, nm.Type())
	}
}

func TestOrderedNavigableMapEmpty(t *testing.T) {
	var nm NM = &OrderedNavigableMap{nm: NavigableMap2{}}
	if !nm.Empty() {
		t.Error("Expected empty type")
	}
	nm = &OrderedNavigableMap{nm: NavigableMap2{"Field1": NewNMInterface("1001")}}
	if nm.Empty() {
		t.Error("Expected not empty type")
	}
}

func TestOrderedNavigableMapLen(t *testing.T) {
	var nm NM = &OrderedNavigableMap{nm: NavigableMap2{}}
	if rply := nm.Len(); rply != 0 {
		t.Errorf("Expected 0 ,received: %v", rply)
	}
	nm = &OrderedNavigableMap{nm: NavigableMap2{"Field1": NewNMInterface("1001")}}
	if rply := nm.Len(); rply != 1 {
		t.Errorf("Expected 1 ,received: %v", rply)
	}
}

func TestOrderedNavigableMapGetField(t *testing.T) {
	nm := &OrderedNavigableMap{nm: NavigableMap2{}}
	if _, err := nm.GetField(PathItem{}); err != ErrNotFound {
		t.Error(err)
	}
	nm = &OrderedNavigableMap{nm: NavigableMap2{
		"Field1": NewNMInterface("1001"),
		"Field2": NewNMInterface("1003"),
		"Field3": NavigableMap2{"Field4": NewNMInterface("Val")},
		"Field5": &NMSlice{NewNMInterface(10), NewNMInterface(101)},
	}}
	if _, err := nm.GetField(PathItem{}); err != ErrNotFound {
		t.Error(err)
	}
	if _, err := nm.GetField(PathItem{Field: "NaN"}); err != ErrNotFound {
		t.Error(err)
	}
	// if _, err := nm.GetField(nil); err != ErrWrongPath {
	// 	t.Error(err)
	// }
	if val, err := nm.GetField(PathItem{Field: "Field1"}); err != nil {
		t.Error(err)
	} else if val.Interface() != "1001" {
		t.Errorf("Expected %q ,received: %q", "1001", val.Interface())
	}
	if _, err := nm.GetField(PathItem{Field: "Field1", Index: IntPointer(0)}); err != ErrNotFound {
		t.Error(err)
	}
	if val, err := nm.GetField(PathItem{Field: "Field5", Index: IntPointer(0)}); err != nil {
		t.Error(err)
	} else if val.Interface() != 10 {
		t.Errorf("Expected %q ,received: %q", 10, val.Interface())
	}
}

func TestOrderedNavigableMapGetSet(t *testing.T) {
	var nm NM = &OrderedNavigableMap{nm: NavigableMap2{
		"Field1": NewNMInterface(10),
		"Field2": &NMSlice{
			NewNMInterface("1001"),
			NavigableMap2{
				"Account": &NMSlice{NewNMInterface(10), NewNMInterface(11)},
			},
		},
		"Field3": NavigableMap2{
			"Field4": NavigableMap2{
				"Field5": NewNMInterface(5),
			},
		},
	}}
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

	path = PathItems{{Field: "Field2", Index: IntPointer(2)}}
	if err := nm.Set(path, NewNMInterface("500")); err != nil {
		t.Error(err)
	}
	if val, err := nm.Field(path); err != nil {
		t.Error(err)
	} else if val.Interface() != "500" {
		t.Errorf("Expected %q ,received: %q", "500", val.Interface())
	}

	path = PathItems{{Field: "Field2", Index: IntPointer(1)}, {Field: "Account"}}
	if err := nm.Set(path, NewNMInterface("5")); err != nil {
		t.Error(err)
	}
	path = PathItems{{Field: "Field2", Index: IntPointer(1)}, {Field: "Account"}}
	if val, err := nm.Field(path); err != nil {
		t.Error(err)
	} else if val.Interface() != "5" {
		t.Errorf("Expected %q ,received: %q", "5", val.Interface())
	}
	path = PathItems{{Field: "Field2", Index: IntPointer(1)}, {Field: "Account", Index: IntPointer(0)}}
	if _, err := nm.Field(path); err != ErrNotFound {
		t.Error(err)
	}
}

func TestOrderedNavigableMapFieldAsInterface(t *testing.T) {
	nm := &OrderedNavigableMap{nm: NavigableMap2{
		"Field1": NewNMInterface("1001"),
		"Field2": NewNMInterface("1003"),
		"Field3": NavigableMap2{"Field4": NewNMInterface("Val")},
		"Field5": &NMSlice{NewNMInterface(10), NewNMInterface(101)},
	}}
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

func TestOrderedNavigableMapFieldAsString(t *testing.T) {
	nm := &OrderedNavigableMap{nm: NavigableMap2{
		"Field1": NewNMInterface("1001"),
		"Field2": NewNMInterface("1003"),
		"Field3": NavigableMap2{"Field4": NewNMInterface("Val")},
		"Field5": &NMSlice{NewNMInterface(10), NewNMInterface(101)},
	}}
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
		t.Errorf("Expected %q ,received: %q", 10, val)
	}
}

func TestOrderedNavigableMapGetOrder(t *testing.T) {
	expected := []PathItems{
		{{Field: "Field1"}, {Field: "Field2", Index: IntPointer(0)}},
		{{Field: "Field1"}, {Field: "Field2", Index: IntPointer(1)}},
		{{Field: "Field3"}, {Field: "Field4"}, {Field: "Field5", Index: IntPointer(0)}},
		{{Field: "Field1"}, {Field: "Field2", Index: IntPointer(2)}},
	}
	nm := &OrderedNavigableMap{order: expected}
	if rply := nm.GetOrder(); !reflect.DeepEqual(rply, expected) {
		t.Errorf("Expected %s ,received: %s", expected, rply)
	}
}

//////////////////////////////////////////

func TestOrderedNavigableMapSet(t *testing.T) {
	nm := NewOrderedNavigableMap()
	if err := nm.Set(nil, nil); err != ErrWrongPath {
		t.Error(err)
	}
	if err := nm.Set(PathItems{{Field: "1"}}, &NavigableMap2{}); err != ErrNotImplemented {
		t.Error(err)
	}
	if err := nm.Set(PathItems{{Field: "Field1", Index: IntPointer(10)}}, NewNMInterface("1001")); err != ErrWrongPath {
		t.Error(err)
	}
	if err := nm.Set(PathItems{{Field: "Field1", Index: IntPointer(10)}, {Field: "Field2"}}, NewNMInterface("1001")); err != ErrWrongPath {
		t.Error(err)
	}
	path := PathItems{{Field: "Field1", Index: IntPointer(0)}}
	expected := &OrderedNavigableMap{
		nm:    NavigableMap2{"Field1": &NMSlice{NewNMInterface("1001")}},
		order: []PathItems{path},
	}
	if err := nm.Set(path, NewNMInterface("1001")); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(nm, expected) {
		t.Errorf("Expected %s ,received: %s", expected, nm)
	}
	if err := nm.Set(PathItems{{Field: "Field1", Index: IntPointer(0)}, {}}, NewNMInterface("1001")); err != ErrNotImplemented {
		t.Error(err)
	}
	if err := nm.Set(PathItems{{Field: "Field1", Index: IntPointer(10)}}, NewNMInterface("1001")); err != ErrWrongPath {
		t.Error(err)
	}

	if err := nm.Set(PathItems{{Field: "Field1", Index: IntPointer(0)}, {}}, &NMSlice{}); err != ErrNotImplemented {
		t.Error(err)
	}
	if err := nm.Set(PathItems{{Field: "Field1", Index: IntPointer(10)}}, &NMSlice{}); err != ErrWrongPath {
		t.Error(err)
	}

	expected = &OrderedNavigableMap{
		nm:    NavigableMap2{"Field1": &NMSlice{NewNMInterface("1002")}},
		order: []PathItems{path},
	}
	if err := nm.Set(path, NewNMInterface("1002")); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(nm, expected) {
		t.Errorf("Expected %s ,received: %s", expected, nm)
	}

	path = PathItems{{Field: "Field2"}}
	expected = &OrderedNavigableMap{
		nm: NavigableMap2{
			"Field1": &NMSlice{NewNMInterface("1002")},
			"Field2": NewNMInterface("1002"),
		},
		order: append(expected.order, path),
	}
	if err := nm.Set(path, NewNMInterface("1002")); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(nm, expected) {
		t.Errorf("Expected %s ,received: %s", expected, nm)
	}

	path = PathItems{{Field: "Field1", Index: IntPointer(1)}}
	expected = &OrderedNavigableMap{
		nm: NavigableMap2{
			"Field1": &NMSlice{NewNMInterface("1002"), NewNMInterface("1003")},
			"Field2": NewNMInterface("1002"),
		},
		order: append(expected.order, path),
	}
	if err := nm.Set(path, NewNMInterface("1003")); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(nm, expected) {
		t.Errorf("Expected %s ,received: %s", expected, nm)
	}

	path = PathItems{{Field: "Field3"}}
	obj := &NMSlice{NewNMInterface("1004"), NewNMInterface("1005")}
	expected = &OrderedNavigableMap{
		nm: NavigableMap2{
			"Field1": &NMSlice{NewNMInterface("1002"), NewNMInterface("1003")},
			"Field2": NewNMInterface("1002"),
			"Field3": obj,
		},
		order: append(expected.order, PathItems{{Field: "Field3", Index: IntPointer(0)}}, PathItems{{Field: "Field3", Index: IntPointer(1)}}),
	}
	if err := nm.Set(path, obj); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(nm, expected) {
		t.Errorf("Expected %s ,received: %s", expected, nm)
	}

	obj = &NMSlice{NewNMInterface("1005"), NewNMInterface("1006")}
	expected = &OrderedNavigableMap{
		nm: NavigableMap2{
			"Field1": &NMSlice{NewNMInterface("1005"), NewNMInterface("1006")},
			"Field2": NewNMInterface("1002"),
			"Field3": &NMSlice{NewNMInterface("1004"), NewNMInterface("1005")},
		},
		order: []PathItems{
			{{Field: "Field2"}},
			{{Field: "Field3", Index: IntPointer(0)}},
			{{Field: "Field3", Index: IntPointer(1)}},
			{{Field: "Field1", Index: IntPointer(0)}},
			{{Field: "Field1", Index: IntPointer(1)}},
		},
	}
	if err := nm.Set(PathItems{{Field: "Field1"}}, obj); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(nm, expected) {
		t.Errorf("Expected %s ,received: %s", expected, nm)
	}

	obj = &NMSlice{NewNMInterface("1005"), NewNMInterface("1006")}
	expected = &OrderedNavigableMap{
		nm: NavigableMap2{
			"Field1": &NMSlice{NewNMInterface("1005"), NewNMInterface("1006")},
			"Field2": NewNMInterface("1002"),
			"Field3": &NMSlice{NewNMInterface("1004"), NewNMInterface("1007")},
		},
		order: []PathItems{
			{{Field: "Field2"}},
			{{Field: "Field3", Index: IntPointer(0)}},
			{{Field: "Field1", Index: IntPointer(0)}},
			{{Field: "Field1", Index: IntPointer(1)}},
			{{Field: "Field3", Index: IntPointer(1)}},
		},
	}
	if err := nm.Set(PathItems{{Field: "Field3", Index: IntPointer(-1)}}, NewNMInterface("1007")); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(nm, expected) {
		t.Errorf("Expected %s ,received: %s", expected.order, nm.order)
	}
}

func TestOrderedNavigableMapSetField(t *testing.T) {
	nm := &OrderedNavigableMap{nm: NavigableMap2{}}
	// if err := nm.SetField(nil, nil); err != ErrWrongPath {
	// 	t.Error(err)
	// }
	if err := nm.SetField(PathItem{Field: "1"}, &NavigableMap2{}); err != ErrNotImplemented {
		t.Error(err)
	}
	if err := nm.SetField(PathItem{Field: "Field1", Index: IntPointer(10)}, NewNMInterface("1001")); err != ErrWrongPath {
		t.Error(err)
	}
	if err := nm.SetField(PathItem{Field: "Field1", Index: IntPointer(10)}, &NMSlice{}); err != ErrWrongPath {
		t.Error(err)
	}
	expected := &OrderedNavigableMap{
		nm:    NavigableMap2{"Field1": NewNMInterface(10)},
		order: []PathItems{{{Field: "Field1"}}},
	}
	if err := nm.SetField(PathItem{Field: "Field1"}, NewNMInterface(10)); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(nm, expected) {
		t.Errorf("Expected %s ,received: %s", expected, nm)
	}

	expected = &OrderedNavigableMap{
		nm:    NavigableMap2{"Field1": NewNMInterface(11)},
		order: []PathItems{{{Field: "Field1"}}},
	}
	if err := nm.SetField(PathItem{Field: "Field1"}, NewNMInterface(11)); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(nm, expected) {
		t.Errorf("Expected %s ,received: %s", expected, nm)
	}
	expected = &OrderedNavigableMap{
		nm: NavigableMap2{
			"Field1": NewNMInterface(11),
			"Field2": &NMSlice{NewNMInterface("1005"), NewNMInterface("1006")},
		},
		order: []PathItems{{{Field: "Field1"}}, {{Field: "Field2", Index: IntPointer(0)}}, {{Field: "Field2", Index: IntPointer(1)}}},
	}
	if err := nm.SetField(PathItem{Field: "Field2"}, &NMSlice{NewNMInterface("1005"), NewNMInterface("1006")}); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(nm, expected) {
		t.Errorf("Expected %s ,received: %s", expected, nm)
	}
	expected = &OrderedNavigableMap{
		nm: NavigableMap2{
			"Field1": NewNMInterface(11),
			"Field2": &NMSlice{NewNMInterface("1007")},
		},
		order: []PathItems{{{Field: "Field1"}}, {{Field: "Field2", Index: IntPointer(0)}}},
	}
	if err := nm.SetField(PathItem{Field: "Field2"}, &NMSlice{NewNMInterface("1007")}); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(nm, expected) {
		t.Errorf("Expected %s ,received: %s", expected, nm)
	}
}

func TestOrderedNavigableMapRemove(t *testing.T) {
	nm := &OrderedNavigableMap{
		nm: NavigableMap2{
			"Field1": NewNMInterface("1001"),
			"Field2": NewNMInterface("1003"),
			"Field3": NavigableMap2{"Field4": NewNMInterface("Val")},
			"Field5": &NMSlice{NewNMInterface(10), NewNMInterface(101)},
		},
		order: []PathItems{
			{{Field: "Field2"}},
			{{Field: "Field3"}, {Field: "Field4"}},
			{{Field: "Field5", Index: IntPointer(0)}},
			{{Field: "Field1"}},
			{{Field: "Field5", Index: IntPointer(1)}},
		},
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

	if err := nm.Remove(PathItems{{Index: IntPointer(-1)}, {}}); err != nil {
		t.Error(err)
	}
	expected := &OrderedNavigableMap{
		nm: NavigableMap2{
			"Field1": NewNMInterface("1001"),
			"Field2": NewNMInterface("1003"),
			"Field3": NavigableMap2{"Field4": NewNMInterface("Val")},
			"Field5": &NMSlice{NewNMInterface(10), NewNMInterface(101)},
		},
		order: []PathItems{
			{{Field: "Field2"}},
			{{Field: "Field3"}, {Field: "Field4"}},
			{{Field: "Field5", Index: IntPointer(0)}},
			{{Field: "Field1"}},
			{{Field: "Field5", Index: IntPointer(1)}},
		},
	}
	if !reflect.DeepEqual(nm, expected) {
		t.Errorf("Expected %s ,received: %s", expected, nm)
	}
	expected = &OrderedNavigableMap{
		nm: NavigableMap2{
			"Field1": NewNMInterface("1001"),
			"Field3": NavigableMap2{"Field4": NewNMInterface("Val")},
			"Field5": &NMSlice{NewNMInterface(10), NewNMInterface(101)},
		},
		order: []PathItems{
			{{Field: "Field3"}, {Field: "Field4"}},
			{{Field: "Field5", Index: IntPointer(0)}},
			{{Field: "Field1"}},
			{{Field: "Field5", Index: IntPointer(1)}},
		},
	}

	if err := nm.Remove(PathItems{{Field: "Field2"}}); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(nm, expected) {
		t.Errorf("Expected %s ,received: %s", expected, nm)
	}

	expected = &OrderedNavigableMap{
		nm: NavigableMap2{
			"Field1": NewNMInterface("1001"),
			"Field3": NavigableMap2{"Field4": NewNMInterface("Val")},
			"Field5": &NMSlice{NewNMInterface(101)},
		},
		order: []PathItems{
			{{Field: "Field3"}, {Field: "Field4"}},
			{{Field: "Field1"}},
			{{Field: "Field5", Index: IntPointer(0)}},
		},
	}

	if err := nm.Remove(PathItems{{Field: "Field5", Index: IntPointer(0)}}); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(nm, expected) {
		t.Errorf("Expected %s ,received: %s", expected, nm)
	}

	expected = &OrderedNavigableMap{
		nm: NavigableMap2{
			"Field1": NewNMInterface("1001"),
			"Field3": NavigableMap2{"Field4": NewNMInterface("Val")},
		},
		order: []PathItems{
			{{Field: "Field3"}, {Field: "Field4"}},
			{{Field: "Field1"}},
		},
	}

	if err := nm.Remove(PathItems{{Field: "Field5", Index: IntPointer(-1)}}); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(nm, expected) {
		t.Errorf("Expected %s ,received: %s", expected.order, nm.order)
	}

	if err := nm.Remove(PathItems{{Field: "Field1", Index: IntPointer(0)}, {}}); err != ErrWrongPath {
		t.Error(err)
	}

	nm = &OrderedNavigableMap{
		nm: NavigableMap2{
			"Field1": NewNMInterface("1001"),
			"Field2": NewNMInterface("1003"),
			"Field6": NavigableMap2{"Field8": NewNMInterface("Val")},
			"Field3": NavigableMap2{
				"Field4": NewNMInterface("Val"),
				"Field7": &NMSlice{NewNMInterface(103), NewNMInterface(102)},
			},
			"Field5": &NMSlice{NewNMInterface(10), NewNMInterface(101)},
		},
		order: []PathItems{
			{{Field: "Field2"}},
			{{Field: "Field3"}, {Field: "Field4"}},
			{{Field: "Field5", Index: IntPointer(0)}},
			{{Field: "Field1"}},
			{{Field: "Field6"}, {Field: "Field8"}},
			{{Field: "Field3"}, {Field: "Field7", Index: IntPointer(0)}},
			{{Field: "Field5", Index: IntPointer(1)}},
			{{Field: "Field3"}, {Field: "Field7", Index: IntPointer(1)}},
		},
	}

	expected = &OrderedNavigableMap{
		nm: NavigableMap2{
			"Field1": NewNMInterface("1001"),
			"Field2": NewNMInterface("1003"),
			"Field6": NavigableMap2{"Field8": NewNMInterface("Val")},
			"Field3": NavigableMap2{
				"Field4": NewNMInterface("Val"),
				"Field7": &NMSlice{NewNMInterface(102)},
			},
			"Field5": &NMSlice{NewNMInterface(10), NewNMInterface(101)},
		},
		order: []PathItems{
			{{Field: "Field2"}},
			{{Field: "Field3"}, {Field: "Field4"}},
			{{Field: "Field5", Index: IntPointer(0)}},
			{{Field: "Field1"}},
			{{Field: "Field6"}, {Field: "Field8"}},
			{{Field: "Field5", Index: IntPointer(1)}},
			{{Field: "Field3"}, {Field: "Field7", Index: IntPointer(0)}},
		},
	}

	if err := nm.Remove(PathItems{{Field: "Field3"}, {Field: "Field7", Index: IntPointer(0)}}); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(nm, expected) {
		t.Errorf("Expected %s ,received: %s", expected.order, nm.order)
	}

}

func TestOrderedNavigableRemote(t *testing.T) {
	nm := &OrderedNavigableMap{nm: NavigableMap2{"Field1": NewNMInterface("1001")}}
	eOut := LocalAddr()
	if rcv := nm.RemoteHost(); !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", eOut, rcv)
	}
}

var generator = rand.New(rand.NewSource(42))
var gen = generateRandomTemplate(1_000)

type benchData struct {
	path      []string
	pathItems PathItems
	data      string
}

func generateRandomPath() (out []string) {
	size := generator.Intn(16) + 1
	out = make([]string, size)
	for i := 0; i < size; i++ {
		out[i] = Sha1(GenUUID())
	}
	return
}
func generateRandomTemplate(size int) (out []benchData) {
	out = make([]benchData, size)
	for i := 0; i < size; i++ {
		out[i].path = generateRandomPath()
		out[i].data = UUIDSha1Prefix()
		out[i].pathItems = NewPathToItem(out[i].path)
	}
	return
}

func BenchmarkOrderdNavigableMapSet(b *testing.B) {
	nm := NewOrderedNavigableMap()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		for _, data := range gen {
			if err := nm.Set(data.pathItems, NewNMInterface(data.data)); err != nil {
				b.Log(err, data.path)
			}
		}
	}
}

func BenchmarkNavigableMapSet(b *testing.B) {
	nm := NavigableMap2{}
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		for _, data := range gen {
			if err := nm.Set(data.pathItems, NewNMInterface(data.data)); err != nil {
				b.Log(err, data.path)
			}
		}
	}
}

func BenchmarkNavigableMapOldSet(b *testing.B) {
	nm := NavigableMap{}
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		for _, data := range gen {
			if err := nm.Set(data.path, data.data); err != nil {
				b.Log(err, data.path)
			}
		}
	}
}

func BenchmarkOrderdNavigableMapFieldAsInterface(b *testing.B) {
	nm := NewOrderedNavigableMap()
	for _, data := range gen {
		if err := nm.Set(data.pathItems, NewNMInterface(data.data)); err != nil {
			b.Log(err, data.path)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, data := range gen {
			if val, err := nm.FieldAsInterface(data.path); err != nil {
				b.Log(err)
			} else if val != data.data {
				b.Errorf("Expected %q ,received: %q", data.data, val)
			}
		}
	}
}

func BenchmarkNavigableMapFieldAsInterface(b *testing.B) {
	nm := NavigableMap2{}
	for _, data := range gen {
		if err := nm.Set(data.pathItems, NewNMInterface(data.data)); err != nil {
			b.Log(err, data.path)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, data := range gen {
			if val, err := nm.FieldAsInterface(data.path); err != nil {
				b.Log(err)
			} else if val != data.data {
				b.Errorf("Expected %q ,received: %q", data.data, val)
			}
		}
	}
}

func BenchmarkNavigableMapOldFieldAsInterface(b *testing.B) {
	nm := NavigableMap{}
	for _, data := range gen {
		if err := nm.Set(data.path, data.data); err != nil {
			b.Log(err, data.path)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, data := range gen {
			if val, err := nm.FieldAsInterface(data.path); err != nil {
				b.Log(err)
			} else if val != data.data {
				b.Errorf("Expected %q ,received: %q", data.data, val)
			}
		}
	}
}

func BenchmarkOrderdNavigableMapField(b *testing.B) {
	nm := NewOrderedNavigableMap()
	for _, data := range gen {
		if err := nm.Set(data.pathItems, NewNMInterface(data.data)); err != nil {
			b.Log(err, data.path)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, data := range gen {
			if val, err := nm.Field(data.pathItems); err != nil {
				b.Log(err)
			} else if val.Interface() != data.data {
				b.Errorf("Expected %q ,received: %q", data.data, val.Interface())
			}
		}
	}
}

func BenchmarkNavigableMapField(b *testing.B) {
	nm := NavigableMap2{}
	for _, data := range gen {
		if err := nm.Set(data.pathItems, NewNMInterface(data.data)); err != nil {
			b.Log(err, data.path)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, data := range gen {
			if val, err := nm.Field(data.pathItems); err != nil {
				b.Log(err)
			} else if val.Interface() != data.data {
				b.Errorf("Expected %q ,received: %q", data.data, val.Interface())
			}
		}
	}
}
