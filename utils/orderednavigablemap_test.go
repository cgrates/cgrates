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

func TestOrderedNavigableMap(t *testing.T) {
	onm := NewOrderedNavigableMap()

	onm.Set(&FullPath{
		Path:      "Field1",
		PathItems: PathItems{{Field: "Field1"}},
	}, NewNMData(10))
	expOrder := []PathItems{
		{{Field: "Field1"}},
	}
	if !reflect.DeepEqual(expOrder, onm.GetOrder()) {
		t.Errorf("Expected %s ,received: %s", expOrder, ToJSON(onm.GetOrder()))
	}

	onm.Set(&FullPath{
		Path:      "Field2[0]",
		PathItems: PathItems{{Field: "Field2", Index: StringPointer("0")}},
	}, NewNMData("1001"))
	expOrder = []PathItems{
		{{Field: "Field1"}},
		{{Field: "Field2", Index: StringPointer("0")}},
	}
	if !reflect.DeepEqual(expOrder, onm.GetOrder()) {
		t.Errorf("Expected %s ,received: %s", expOrder, ToJSON(onm.GetOrder()))
	}

	onm.Set(&FullPath{
		Path: "Field2[1].Account[0]",
		PathItems: PathItems{
			{Field: "Field2", Index: StringPointer("1")},
			{Field: "Account", Index: StringPointer("0")}},
	}, NewNMData(10))
	expOrder = []PathItems{
		{{Field: "Field1"}},
		{{Field: "Field2", Index: StringPointer("0")}},
		{
			{Field: "Field2", Index: StringPointer("1")},
			{Field: "Account", Index: StringPointer("0")}},
	}
	if !reflect.DeepEqual(expOrder, onm.GetOrder()) {
		t.Errorf("Expected %s ,received: %s", expOrder, ToJSON(onm.GetOrder()))
	}

	onm.Set(&FullPath{
		Path: "Field2[1].Account[1]",
		PathItems: PathItems{
			{Field: "Field2", Index: StringPointer("1")},
			{Field: "Account", Index: StringPointer("1")}},
	}, NewNMData(11))
	expOrder = []PathItems{
		{{Field: "Field1"}},
		{{Field: "Field2", Index: StringPointer("0")}},
		{
			{Field: "Field2", Index: StringPointer("1")},
			{Field: "Account", Index: StringPointer("0")}},
		{
			{Field: "Field2", Index: StringPointer("1")},
			{Field: "Account", Index: StringPointer("1")}},
	}
	if !reflect.DeepEqual(expOrder, onm.GetOrder()) {
		t.Errorf("Expected %s ,received: %s", expOrder, ToJSON(onm.GetOrder()))
	}

	onm.Set(&FullPath{
		Path:      "Field2[2]",
		PathItems: PathItems{{Field: "Field2", Index: StringPointer("2")}},
	}, NewNMData(111))
	expOrder = []PathItems{
		{{Field: "Field1"}},
		{{Field: "Field2", Index: StringPointer("0")}},
		{
			{Field: "Field2", Index: StringPointer("1")},
			{Field: "Account", Index: StringPointer("0")}},
		{
			{Field: "Field2", Index: StringPointer("1")},
			{Field: "Account", Index: StringPointer("1")}},
		{{Field: "Field2", Index: StringPointer("2")}},
	}
	if !reflect.DeepEqual(expOrder, onm.GetOrder()) {
		t.Errorf("Expected %s ,received: %s", expOrder, ToJSON(onm.GetOrder()))
	}

	onm.Set(&FullPath{
		Path: "Field3.Field4.Field5",
		PathItems: PathItems{
			{Field: "Field3"},
			{Field: "Field4"},
			{Field: "Field5"}},
	}, NewNMData(5))
	expOrder = []PathItems{
		{{Field: "Field1"}},
		{{Field: "Field2", Index: StringPointer("0")}},
		{
			{Field: "Field2", Index: StringPointer("1")},
			{Field: "Account", Index: StringPointer("0")}},
		{
			{Field: "Field2", Index: StringPointer("1")},
			{Field: "Account", Index: StringPointer("1")}},
		{{Field: "Field2", Index: StringPointer("2")}},
		{
			{Field: "Field3"},
			{Field: "Field4"},
			{Field: "Field5"}},
	}
	if !reflect.DeepEqual(expOrder, onm.GetOrder()) {
		t.Errorf("Expected %s ,received: %s", expOrder, ToJSON(onm.GetOrder()))
	}

	var expnm NMInterface = NavigableMap2{
		"Field1": NewNMData(10),
		"Field2": &NMSlice{
			NewNMData("1001"),
			NavigableMap2{
				"Account": &NMSlice{NewNMData(10), NewNMData(11)},
			},
			NewNMData(111),
		},
		"Field3": NavigableMap2{
			"Field4": NavigableMap2{
				"Field5": NewNMData(5),
			},
		},
	}
	if onm.Empty() {
		t.Error("Expected not empty type")
	}
	if onm.Type() != NMMapType {
		t.Errorf("Expected %v ,received: %v", NMDataType, onm.Type())
	}
	if !reflect.DeepEqual(expnm, onm.nm) {
		t.Errorf("Expected %s ,received: %s", expnm, onm.nm)
	}

	// sliceDeNM
	exp := &NMSlice{NewNMData("500"), NewNMData("502")}
	path := PathItems{{Field: "Field2"}}
	if _, err := onm.Set(&FullPath{Path: path.String(), PathItems: path}, exp); err != nil {
		t.Error(err)
	}
	path = PathItems{{Field: "Field2"}}
	if val, err := onm.Field(path); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(val, exp) {
		t.Errorf("Expected %q ,received: %q", exp, val.Interface())
	}
	expOrder = []PathItems{
		{{Field: "Field1"}},
		{
			{Field: "Field3"},
			{Field: "Field4"},
			{Field: "Field5"}},
		{{Field: "Field2", Index: StringPointer("0")}},
		{{Field: "Field2", Index: StringPointer("1")}},
	}
	if !reflect.DeepEqual(expOrder, onm.GetOrder()) {
		t.Errorf("Expected %s ,received: %s", expOrder, onm.GetOrder())
	}

	path = PathItems{{Field: "Field2", Index: StringPointer("0")}}
	if val, err := onm.Field(path); err != nil {
		t.Error(err)
	} else if val.Interface() != "500" {
		t.Errorf("Expected %q ,received: %q", "500", val.Interface())
	}
	expnm = NavigableMap2{
		"Field1": NewNMData(10),
		"Field3": NavigableMap2{
			"Field4": NavigableMap2{
				"Field5": NewNMData(5),
			},
		},
		"Field2": &NMSlice{
			NewNMData("500"),
			NewNMData("502"),
		},
	}
	if !reflect.DeepEqual(expnm, onm.nm) {
		t.Errorf("Expected %s ,received: %s", expnm, onm.nm)
	}
}

func TestOrderedNavigableMapString(t *testing.T) {
	nm := &OrderedNavigableMap{nm: NavigableMap2{"Field1": NewNMData("1001")}}
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
	nm := &OrderedNavigableMap{nm: NavigableMap2{"Field1": NewNMData("1001"), "Field2": NewNMData("1003")}}
	expected := NavigableMap2{"Field1": NewNMData("1001"), "Field2": NewNMData("1003")}
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
		"Field1": NewNMData("1001"),
		"Field2": NewNMData("1003"),
		"Field3": NavigableMap2{"Field4": NewNMData("Val")},
		"Field5": &NMSlice{NewNMData(10), NewNMData(101)},
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

	if _, err := nm.Field(PathItems{{Field: "Field1", Index: StringPointer("0")}}); err != ErrNotFound {
		t.Error(err)
	}
	if val, err := nm.Field(PathItems{{Field: "Field5", Index: StringPointer("0")}}); err != nil {
		t.Error(err)
	} else if val.Interface() != 10 {
		t.Errorf("Expected %q ,received: %q", 10, val.Interface())
	}
	if _, err := nm.Field(PathItems{{Field: "Field3", Index: StringPointer("0")}}); err != ErrNotFound {
		t.Error(err)
	}
	if val, err := nm.Field(PathItems{{Field: "Field3"}, {Field: "Field4"}}); err != nil {
		t.Error(err)
	} else if val.Interface() != "Val" {
		t.Errorf("Expected %q ,received: %q", "Val", val.Interface())
	}
}

func TestOrderedNavigableMapType(t *testing.T) {
	nm := &OrderedNavigableMap{nm: NavigableMap2{}}
	if nm.Type() != NMMapType {
		t.Errorf("Expected %v ,received: %v", NMMapType, nm.Type())
	}
}

func TestOrderedNavigableMapEmpty(t *testing.T) {
	nm := &OrderedNavigableMap{nm: NavigableMap2{}}
	if !nm.Empty() {
		t.Error("Expected empty type")
	}
	nm = &OrderedNavigableMap{nm: NavigableMap2{"Field1": NewNMData("1001")}}
	if nm.Empty() {
		t.Error("Expected not empty type")
	}
}

func TestOrderedNavigableMapLen(t *testing.T) {
	nm := &OrderedNavigableMap{nm: NavigableMap2{}}
	if rply := nm.Len(); rply != 0 {
		t.Errorf("Expected 0 ,received: %v", rply)
	}
	nm = &OrderedNavigableMap{nm: NavigableMap2{"Field1": NewNMData("1001")}}
	if rply := nm.Len(); rply != 1 {
		t.Errorf("Expected 1 ,received: %v", rply)
	}
}

func TestOrderedNavigableMapGetSet(t *testing.T) {
	nm := NewOrderedNavigableMap()
	nm.Set(&FullPath{
		PathItems: PathItems{{Field: "Account", Index: StringPointer("0")}},
		Path:      "Account",
	}, NewNMData(1001))
	nm.Set(&FullPath{
		PathItems: PathItems{{Field: "Account", Index: StringPointer("1")}},
		Path:      "Account",
	}, NewNMData("account_on_new_branch"))

	expectedOrder := []PathItems{
		{{Field: "Account", Index: StringPointer("0")}},
		{{Field: "Account", Index: StringPointer("1")}},
	}

	if receivedOrder := nm.GetOrder(); !reflect.DeepEqual(expectedOrder, receivedOrder) {
		t.Errorf("Expected %s ,received: %s", expectedOrder, receivedOrder)
	}
	nm = &OrderedNavigableMap{
		nm: NavigableMap2{
			"Field1": NewNMData(10),
			"Field2": &NMSlice{
				NewNMData("1001"),
				NavigableMap2{
					"Account": &NMSlice{NewNMData(10), NewNMData(11)},
				},
			},
			"Field3": NavigableMap2{
				"Field4": NavigableMap2{
					"Field5": NewNMData(5),
				},
			},
		},
		orderIdx: NewPathItemList(),
		orderRef: make(map[string][]*PathItemElement),
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

	path = PathItems{{Field: "Field2", Index: StringPointer("2")}}
	if _, err := nm.Set(&FullPath{Path: path.String(), PathItems: path}, NewNMData("500")); err != nil {
		t.Error(err)
	}
	if val, err := nm.Field(path); err != nil {
		t.Error(err)
	} else if val.Interface() != "500" {
		t.Errorf("Expected %q ,received: %q", "500", val.Interface())
	}

	path = PathItems{{Field: "Field2", Index: StringPointer("1")}, {Field: "Account"}}
	if _, err := nm.Set(&FullPath{Path: path.String(), PathItems: path}, NewNMData("5")); err != nil {
		t.Error(err)
	}
	path = PathItems{{Field: "Field2", Index: StringPointer("1")}, {Field: "Account"}}
	if val, err := nm.Field(path); err != nil {
		t.Error(err)
	} else if val.Interface() != "5" {
		t.Errorf("Expected %q ,received: %q", "5", val.Interface())
	}
	path = PathItems{{Field: "Field2", Index: StringPointer("1")}, {Field: "Account", Index: StringPointer("0")}}
	if _, err := nm.Field(path); err != ErrNotFound {
		t.Error(err)
	}
}

func TestOrderedNavigableMapFieldAsInterface(t *testing.T) {
	nm := &OrderedNavigableMap{nm: NavigableMap2{
		"Field1": NewNMData("1001"),
		"Field2": NewNMData("1003"),
		"Field3": NavigableMap2{"Field4": NewNMData("Val")},
		"Field5": &NMSlice{NewNMData(10), NewNMData(101)},
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
		"Field1": NewNMData("1001"),
		"Field2": NewNMData("1003"),
		"Field3": NavigableMap2{"Field4": NewNMData("Val")},
		"Field5": &NMSlice{NewNMData(10), NewNMData(101)},
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
	nm := NewOrderedNavigableMap()
	nm.Set(&FullPath{
		PathItems: PathItems{{Field: "Field1"}, {Field: "Field2", Index: StringPointer("0")}},
		Path:      "Field1.Field2[0]",
	}, NewNMData("1003"))
	nm.Set(&FullPath{
		PathItems: PathItems{{Field: "Field1"}, {Field: "Field2", Index: StringPointer("1")}},
		Path:      "Field1.Field2[1]",
	}, NewNMData("Val"))
	nm.Set(&FullPath{
		PathItems: PathItems{{Field: "Field3"}, {Field: "Field4"}, {Field: "Field5", Index: StringPointer("0")}},
		Path:      "Field3.Field4.Field5",
	}, NewNMData("1001"))
	nm.Set(&FullPath{
		PathItems: PathItems{{Field: "Field1"}, {Field: "Field2", Index: StringPointer("2")}},
		Path:      "Field1.Field2[2]",
	}, NewNMData(101))
	expected := []PathItems{
		{{Field: "Field1"}, {Field: "Field2", Index: StringPointer("0")}},
		{{Field: "Field1"}, {Field: "Field2", Index: StringPointer("1")}},
		{{Field: "Field3"}, {Field: "Field4"}, {Field: "Field5", Index: StringPointer("0")}},
		{{Field: "Field1"}, {Field: "Field2", Index: StringPointer("2")}},
	}
	if rply := nm.GetOrder(); !reflect.DeepEqual(rply, expected) {
		t.Errorf("Expected %s ,received: %s", expected, rply)
	}
}

//////////////////////////////////////////

func TestOrderedNavigableMapSet(t *testing.T) {
	nm := NewOrderedNavigableMap()
	if _, err := nm.Set(nil, nil); err != ErrWrongPath {
		t.Error(err)
	}
	if _, err := nm.Set(&FullPath{
		PathItems: PathItems{{Field: "Field1", Index: StringPointer("10")}},
		Path:      "Field1[10]",
	}, NewNMData("1001")); err != ErrWrongPath {
		t.Error(err)
	}
	if _, err := nm.Set(&FullPath{
		PathItems: PathItems{{Field: "Field1", Index: StringPointer("10")}, {Field: "Field2"}},
		Path:      "Field1[10].Field2",
	}, NewNMData("1001")); err != ErrWrongPath {
		t.Error(err)
	}
	path := PathItems{{Field: "Field1", Index: StringPointer("0")}}
	if addedNew, err := nm.Set(&FullPath{
		PathItems: path,
		Path:      path.String(),
	}, NewNMData("1001")); err != nil {
		t.Error(err)
	} else if !addedNew {
		t.Error("Expected the field to be added new")
	}
	nMap := NavigableMap2{"Field1": &NMSlice{NewNMData("1001")}}
	order := []PathItems{path}
	if !reflect.DeepEqual(nm.nm, nMap) {
		t.Errorf("Expected %s ,received: %s", nMap, nm)
	}
	if !reflect.DeepEqual(nm.GetOrder(), order) {
		t.Errorf("Expected %s ,received: %s", order, nm.GetOrder())
	}
	if _, err := nm.Set(&FullPath{
		PathItems: PathItems{{Field: "Field1", Index: StringPointer("0")}, {}},
		Path:      "Field1[0]",
	}, NewNMData("1001")); err != ErrWrongPath {
		t.Error(err)
	}
	if _, err := nm.Set(&FullPath{
		PathItems: PathItems{{Field: "Field1", Index: StringPointer("10")}},
		Path:      "Field1[10]",
	}, NewNMData("1001")); err != ErrWrongPath {
		t.Error(err)
	}

	if _, err := nm.Set(&FullPath{
		PathItems: PathItems{{Field: "Field1", Index: StringPointer("0")}, {}},
		Path:      "Field1[0]",
	}, &NMSlice{}); err != ErrWrongPath {
		t.Error(err)
	}
	if _, err := nm.Set(&FullPath{
		PathItems: PathItems{{Field: "Field1", Index: StringPointer("10")}},
		Path:      "Field[10]",
	}, &NMSlice{}); err != ErrWrongPath {
		t.Error(err)
	}

	nMap = NavigableMap2{"Field1": &NMSlice{NewNMData("1002")}}
	order = []PathItems{path}
	if addedNew, err := nm.Set(&FullPath{
		PathItems: path,
		Path:      path.String(),
	}, NewNMData("1002")); err != nil {
		t.Error(err)
	} else if addedNew {
		t.Error("Expected the field to be only updated")
	}
	if !reflect.DeepEqual(nm.nm, nMap) {
		t.Errorf("Expected %s ,received: %s", nMap, nm)
	}
	if !reflect.DeepEqual(nm.GetOrder(), order) {
		t.Errorf("Expected %s ,received: %s", order, nm.GetOrder())
	}
	path = PathItems{{Field: "Field2"}}
	nMap = NavigableMap2{
		"Field1": &NMSlice{NewNMData("1002")},
		"Field2": NewNMData("1002"),
	}
	order = append(order, path)
	if addedNew, err := nm.Set(&FullPath{
		PathItems: path,
		Path:      path.String(),
	}, NewNMData("1002")); err != nil {
		t.Error(err)
	} else if !addedNew {
		t.Error("Expected the field to be added new")
	}
	if !reflect.DeepEqual(nm.nm, nMap) {
		t.Errorf("Expected %s ,received: %s", nMap, nm)
	}
	if !reflect.DeepEqual(nm.GetOrder(), order) {
		t.Errorf("Expected %s ,received: %s", order, nm.GetOrder())
	}
	path = PathItems{{Field: "Field1", Index: StringPointer("1")}}
	nMap = NavigableMap2{
		"Field1": &NMSlice{NewNMData("1002"), NewNMData("1003")},
		"Field2": NewNMData("1002"),
	}
	order = append(order, path)
	if addedNew, err := nm.Set(&FullPath{
		PathItems: path,
		Path:      path.String(),
	}, NewNMData("1003")); err != nil {
		t.Error(err)
	} else if !addedNew {
		t.Error("Expected the field to be added new")
	}
	if !reflect.DeepEqual(nm.nm, nMap) {
		t.Errorf("Expected %s ,received: %s", nMap, nm)
	}
	if !reflect.DeepEqual(nm.GetOrder(), order) {
		t.Errorf("Expected %s ,received: %s", order, nm.GetOrder())
	}
	path = PathItems{{Field: "Field3"}}
	obj := &NMSlice{NewNMData("1004"), NewNMData("1005")}
	nMap = NavigableMap2{
		"Field1": &NMSlice{NewNMData("1002"), NewNMData("1003")},
		"Field2": NewNMData("1002"),
		"Field3": obj,
	}
	order = append(order, PathItems{{Field: "Field3", Index: StringPointer("0")}}, PathItems{{Field: "Field3", Index: StringPointer("1")}})
	if addedNew, err := nm.Set(&FullPath{
		PathItems: path,
		Path:      path.String(),
	}, obj); err != nil {
		t.Error(err)
	} else if !addedNew {
		t.Error("Expected the field to be added new")
	}
	if !reflect.DeepEqual(nm.nm, nMap) {
		t.Errorf("Expected %s ,received: %s", nMap, nm)
	}
	if !reflect.DeepEqual(nm.GetOrder(), order) {
		t.Errorf("Expected %s ,received: %s", order, nm.GetOrder())
	}
	obj = &NMSlice{NewNMData("1005"), NewNMData("1006")}
	nMap = NavigableMap2{
		"Field1": &NMSlice{NewNMData("1005"), NewNMData("1006")},
		"Field2": NewNMData("1002"),
		"Field3": &NMSlice{NewNMData("1004"), NewNMData("1005")},
	}
	order = []PathItems{
		{{Field: "Field2"}},
		{{Field: "Field3", Index: StringPointer("0")}},
		{{Field: "Field3", Index: StringPointer("1")}},
		{{Field: "Field1", Index: StringPointer("0")}},
		{{Field: "Field1", Index: StringPointer("1")}},
	}
	if addedNew, err := nm.Set(&FullPath{
		PathItems: PathItems{{Field: "Field1"}},
		Path:      "Field1",
	}, obj); err != nil {
		t.Error(err)
	} else if addedNew {
		t.Error("Expected the field to be only updated")
	}
	if !reflect.DeepEqual(nm.nm, nMap) {
		t.Errorf("Expected %s ,received: %s", nMap, nm)
	}
	if !reflect.DeepEqual(nm.GetOrder(), order) {
		t.Errorf("Expected %s ,received: %s", order, nm.GetOrder())
	}
	obj = &NMSlice{NewNMData("1005"), NewNMData("1006")}
	nMap = NavigableMap2{
		"Field1": &NMSlice{NewNMData("1005"), NewNMData("1006")},
		"Field2": NewNMData("1002"),
		"Field3": &NMSlice{NewNMData("1004"), NewNMData("1007")},
	}
	order = []PathItems{
		{{Field: "Field2"}},
		{{Field: "Field3", Index: StringPointer("0")}},
		{{Field: "Field1", Index: StringPointer("0")}},
		{{Field: "Field1", Index: StringPointer("1")}},
		{{Field: "Field3", Index: StringPointer("1")}},
	}
	if addedNew, err := nm.Set(&FullPath{
		PathItems: PathItems{{Field: "Field3", Index: StringPointer("-1")}},
		Path:      "Field3[-1]",
	}, NewNMData("1007")); err != nil {
		t.Error(err)
	} else if addedNew {
		t.Error("Expected the field to be only updated")
	}
	if !reflect.DeepEqual(nm.nm, nMap) {
		t.Errorf("Expected %s ,received: %s", nMap, nm)
	}
	if !reflect.DeepEqual(nm.GetOrder(), order) {
		t.Errorf("Expected %s ,received: %s", order, nm.GetOrder())
	}
}

func TestOrderedNavigableMapRemove(t *testing.T) {
	nm := NewOrderedNavigableMap()
	nm.Set(&FullPath{
		PathItems: PathItems{{Field: "Field2"}},
		Path:      "Field2",
	}, NewNMData("1003"))
	nm.Set(&FullPath{
		PathItems: PathItems{{Field: "Field3"}, {Field: "Field4"}},
		Path:      "Field3.Field4",
	}, NewNMData("Val"))
	nm.Set(&FullPath{
		PathItems: PathItems{{Field: "Field1"}},
		Path:      "Field1",
	}, NewNMData("1001"))
	nm.Set(&FullPath{
		PathItems: PathItems{{Field: "Field5"}},
		Path:      "Field5",
	}, &NMSlice{NewNMData(10), NewNMData(101)})

	if err := nm.Remove(&FullPath{}); err != ErrWrongPath {
		t.Error(err)
	}
	if err := nm.Remove(&FullPath{PathItems: PathItems{}}); err != ErrWrongPath {
		t.Error(err)
	}
	if err := nm.Remove(&FullPath{PathItems: PathItems{{Field: "field"}}, Path: "field"}); err != nil {
		t.Error(err)
	}

	if err := nm.Remove(&FullPath{PathItems: PathItems{{Index: StringPointer("-1")}, {}}}); err != ErrWrongPath {
		t.Error(err)
	}
	nMap := NavigableMap2{
		"Field1": NewNMData("1001"),
		"Field2": NewNMData("1003"),
		"Field3": NavigableMap2{"Field4": NewNMData("Val")},
		"Field5": &NMSlice{NewNMData(10), NewNMData(101)},
	}
	order := []PathItems{
		{{Field: "Field2"}},
		{{Field: "Field3"}, {Field: "Field4"}},
		{{Field: "Field1"}},
		{{Field: "Field5", Index: StringPointer("0")}},
		{{Field: "Field5", Index: StringPointer("1")}},
	}
	if !reflect.DeepEqual(nm.nm, nMap) {
		t.Errorf("Expected %s ,received: %s", nMap, nm)
	}
	if !reflect.DeepEqual(nm.GetOrder(), order) {
		t.Errorf("Expected %s ,received: %s", order, nm.GetOrder())
	}
	nMap = NavigableMap2{
		"Field1": NewNMData("1001"),
		"Field3": NavigableMap2{"Field4": NewNMData("Val")},
		"Field5": &NMSlice{NewNMData(10), NewNMData(101)},
	}
	order = []PathItems{
		{{Field: "Field3"}, {Field: "Field4"}},
		{{Field: "Field1"}},
		{{Field: "Field5", Index: StringPointer("0")}},
		{{Field: "Field5", Index: StringPointer("1")}},
	}

	if err := nm.Remove(&FullPath{PathItems: PathItems{{Field: "Field2"}}, Path: "Field2"}); err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(nm.nm, nMap) {
		t.Errorf("Expected %s ,received: %s", nMap, nm)
	}
	if !reflect.DeepEqual(nm.GetOrder(), order) {
		t.Errorf("Expected %s ,received: %s", order, nm.GetOrder())
	}
	nMap = NavigableMap2{
		"Field1": NewNMData("1001"),
		"Field3": NavigableMap2{"Field4": NewNMData("Val")},
	}
	order = []PathItems{
		{{Field: "Field3"}, {Field: "Field4"}},
		{{Field: "Field1"}},
	}

	if err := nm.Remove(&FullPath{PathItems: PathItems{{Field: "Field5"}}, Path: "Field5"}); err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(nm.nm, nMap) {
		t.Errorf("Expected %s ,received: %s", nMap, nm)
	}
	if !reflect.DeepEqual(nm.GetOrder(), order) {
		t.Errorf("Expected %s ,received: %s", order, nm.GetOrder())
	}
	if err := nm.Remove(&FullPath{PathItems: PathItems{{Field: "Field1", Index: StringPointer("0")}, {}}}); err != ErrWrongPath {
		t.Error(err)
	}
}

func TestOrderedNavigableRemote(t *testing.T) {
	nm := &OrderedNavigableMap{nm: NavigableMap2{"Field1": NewNMData("1001")}}
	eOut := LocalAddr()
	if rcv := nm.RemoteHost(); !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", eOut, rcv)
	}
}

/*
func BenchmarkOrderdNavigableMapSet(b *testing.B) {
	nm := NewOrderedNavigableMap()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		for _, data := range gen {
			if _,err := nm.Set(data.pathItems, NewNMData(data.data)); err != nil {
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
			if _,err := nm.Set(data.pathItems, NewNMData(data.data)); err != nil {
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
			if _,err := nm.Set(data.path, data.data); err != nil {
				b.Log(err, data.path)
			}
		}
	}
}

func BenchmarkOrderdNavigableMapFieldAsInterface(b *testing.B) {
	nm := NewOrderedNavigableMap()
	for _, data := range gen {
		if _,err := nm.Set(data.pathItems, NewNMData(data.data)); err != nil {
			b.Log(err, data.path)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, data := range gen {
			if val, err := nm.FieldAsInterface(data.path); err != nil {
				b.Log(err)
			} else if (*(val.(*NMSlice)))[0].Interface() != data.data {
				b.Errorf("Expected %q ,received: %q", data.data, val)
			}
		}
	}
}

func BenchmarkNavigableMapFieldAsInterface(b *testing.B) {
	nm := NavigableMap2{}
	for _, data := range gen {
		if _,err := nm.Set(data.pathItems, NewNMData(data.data)); err != nil {
			b.Log(err, data.path)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, data := range gen {
			if val, err := nm.FieldAsInterface(data.path); err != nil {
				b.Log(err)
			} else if (*(val.(*NMSlice)))[0].Interface() != data.data {
				b.Errorf("Expected %q ,received: %q", data.data, val)
			}
		}
	}
}

func BenchmarkNavigableMapOldFieldAsInterface(b *testing.B) {
	nm := NavigableMap{}
	for _, data := range gen {
		if _,err := nm.Set(data.path, data.data); err != nil {
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

func BenchmarkNavigableMapOld1FieldAsInterface(b *testing.B) {
	nm := NewNavigableMapOld1(nil)
	for _, data := range gen {
		nm.Set(data.path, data.data, true)
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
		if _,err := nm.Set(data.pathItems, NewNMData(data.data)); err != nil {
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
		if _,err := nm.Set(data.pathItems, NewNMData(data.data)); err != nil {
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
//*/

func TestOrderedNavigableMapRemoveAll(t *testing.T) {
	nm := NewOrderedNavigableMap()
	nm.Set(&FullPath{
		PathItems: PathItems{{Field: "Field2"}},
		Path:      "Field2",
	}, NewNMData("1003"))
	nm.Set(&FullPath{
		PathItems: PathItems{{Field: "Field3"}, {Field: "Field4"}},
		Path:      "Field3.Field4",
	}, NewNMData("Val"))
	nm.Set(&FullPath{
		PathItems: PathItems{{Field: "Field1"}},
		Path:      "Field1",
	}, NewNMData("1001"))
	nm.Set(&FullPath{
		PathItems: PathItems{{Field: "Field5"}},
		Path:      "Field5",
	}, &NMSlice{NewNMData(10), NewNMData(101)})
	expected := NewOrderedNavigableMap()
	nm.RemoveAll()
	if !reflect.DeepEqual(nm, expected) {
		t.Errorf("Expected %s ,received: %s", expected, nm)
	}
}

func TestOrderedNavigableMapRemove2(t *testing.T) {
	nm := &OrderedNavigableMap{
		nm: NavigableMap2{
			"Field1": &NMSlice{},
		},
	}
	expErr := `strconv.Atoi: parsing "nan": invalid syntax`
	if err := nm.Remove(&FullPath{PathItems: PathItems{{Field: "Field1", Index: StringPointer("nan")}, {}}, Path: "Field1[nan]"}); err == nil || err.Error() != expErr {
		t.Errorf("Expected error: %s,received: %v", expErr, err)
	}
}

func TestOrderedNavigableMapOrderedFields(t *testing.T) {
	nm := NewOrderedNavigableMap()
	nm.Set(&FullPath{
		PathItems: PathItems{{Field: "Field1"}, {Field: "Field2", Index: StringPointer("0")}},
		Path:      "Field1.Field2[0]",
	}, NewNMData("1003"))
	nm.Set(&FullPath{
		PathItems: PathItems{{Field: "Field1"}, {Field: "Field3", Index: StringPointer("0")}},
		Path:      "Field1.Field3[0]",
	}, NewNMData("1004"))
	nm.Set(&FullPath{
		PathItems: PathItems{{Field: "Field5"}},
		Path:      "Field5",
	}, NewNMData("1005"))
	nm.Set(&FullPath{
		PathItems: PathItems{{Field: "Field6"}},
		Path:      "Field6",
	}, NewNMData("1006"))
	nm.Remove(&FullPath{
		PathItems: PathItems{{Field: "Field5"}},
		Path:      "Field5",
	})
	exp := []interface{}{"1003", "1004", "1006"}
	rcv := nm.OrderedFields()
	if !reflect.DeepEqual(exp, rcv) {
		t.Errorf("Expected %+v, received %+v", exp, rcv)
	}
}
