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
	"strings"
	"testing"
)

func TestOrderedNavigableMap(t *testing.T) {
	onm := NewOrderedNavigableMap()

	onm.Set(&FullPath{
		Path:      "Field1",
		PathSlice: []string{"Field1"},
	}, NewLeafNode(10))
	expOrder := [][]string{{"Field1"}}
	if !reflect.DeepEqual(expOrder, onm.GetOrder()) {
		t.Errorf("Expected %s ,received: %s", expOrder, ToJSON(onm.GetOrder()))
	}

	onm.Set(&FullPath{
		Path:      "Field2[0]",
		PathSlice: []string{"Field2", "0"},
	}, NewLeafNode("1001"))
	expOrder = [][]string{
		{"Field1"},
		{"Field2", "0"},
	}
	if !reflect.DeepEqual(expOrder, onm.GetOrder()) {
		t.Errorf("Expected %s ,received: %s", expOrder, ToJSON(onm.GetOrder()))
	}

	onm.Set(&FullPath{
		Path:      "Field2[1].Account[0]",
		PathSlice: []string{"Field2", "1", "Account", "0"},
	}, NewLeafNode(10))
	expOrder = [][]string{
		{"Field1"},
		{"Field2", "0"},
		{"Field2", "1", "Account", "0"},
	}
	if !reflect.DeepEqual(expOrder, onm.GetOrder()) {
		t.Errorf("Expected %s ,received: %s", expOrder, ToJSON(onm.GetOrder()))
	}

	onm.Set(&FullPath{
		Path:      "Field2[1].Account[1]",
		PathSlice: []string{"Field2", "1", "Account", "1"},
	}, NewLeafNode(11))
	expOrder = [][]string{
		{"Field1"},
		{"Field2", "0"},
		{"Field2", "1", "Account", "0"},
		{"Field2", "1", "Account", "1"},
	}
	if !reflect.DeepEqual(expOrder, onm.GetOrder()) {
		t.Errorf("Expected %s ,received: %s", expOrder, ToJSON(onm.GetOrder()))
	}

	onm.Set(&FullPath{
		Path:      "Field2[2]",
		PathSlice: []string{"Field2", "2"},
	}, NewLeafNode(111))
	expOrder = [][]string{
		{"Field1"},
		{"Field2", "0"},
		{"Field2", "1", "Account", "0"},
		{"Field2", "1", "Account", "1"},
		{"Field2", "2"},
	}
	if !reflect.DeepEqual(expOrder, onm.GetOrder()) {
		t.Errorf("Expected %s ,received: %s", expOrder, ToJSON(onm.GetOrder()))
	}

	onm.Set(&FullPath{
		Path:      "Field3.Field4.Field5",
		PathSlice: []string{"Field3", "Field4", "Field5"},
	}, NewLeafNode(5))
	expOrder = [][]string{
		{"Field1"},
		{"Field2", "0"},
		{"Field2", "1", "Account", "0"},
		{"Field2", "1", "Account", "1"},
		{"Field2", "2"},
		{"Field3", "Field4", "Field5"},
	}
	if !reflect.DeepEqual(expOrder, onm.GetOrder()) {
		t.Errorf("Expected %s ,received: %s", expOrder, ToJSON(onm.GetOrder()))
	}

	expnm := &DataNode{Type: NMMapType, Map: map[string]*DataNode{
		"Field1": NewLeafNode(10),
		"Field2": {Type: NMSliceType, Slice: []*DataNode{
			NewLeafNode("1001"),
			{Type: NMMapType, Map: map[string]*DataNode{
				"Account": {Type: NMSliceType, Slice: []*DataNode{NewLeafNode(10), NewLeafNode(11)}},
			}},
			NewLeafNode(111),
		}},
		"Field3": {Type: NMMapType, Map: map[string]*DataNode{
			"Field4": {Type: NMMapType, Map: map[string]*DataNode{
				"Field5": NewLeafNode(5),
			},
			},
		}}}}
	if onm.Empty() {
		t.Error("Expected not empty type")
	}
	if onm.nm.Type != NMMapType {
		t.Errorf("Expected %v ,received: %v", NMDataType, onm.nm.Type)
	}
	if !reflect.DeepEqual(expnm, onm.nm) {
		t.Errorf("Expected %s ,received: %s", ToJSON(expnm), ToJSON(onm.nm))
	}

	// sliceDeNM
	exp := []*DataNode{NewLeafNode("500"), NewLeafNode("502")}
	path := []string{"Field2"}
	if err := onm.SetAsSlice(&FullPath{Path: path[0], PathSlice: path}, exp); err != nil {
		t.Error(err)
	}
	path = []string{"Field2"}
	if val, err := onm.FieldAsInterface(path); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(val, exp) {
		t.Errorf("Expected %q ,received: %q", ToJSON(exp), ToJSON(val))
	}
	expOrder = [][]string{
		{"Field1"},
		{"Field3", "Field4", "Field5"},
		{"Field2", "0"},
		{"Field2", "1"},
	}
	if !reflect.DeepEqual(expOrder, onm.GetOrder()) {
		t.Errorf("Expected %s ,received: %s", expOrder, onm.GetOrder())
	}

	path = []string{"Field2", "0"}
	if val, err := onm.Field(path); err != nil {
		t.Error(err)
	} else if val.Data != "500" {
		t.Errorf("Expected %q ,received: %q", "500", val.Data)
	}
	expnm = &DataNode{Type: NMMapType, Map: map[string]*DataNode{
		"Field1": NewLeafNode(10),
		"Field3": {Type: NMMapType, Map: map[string]*DataNode{
			"Field4": {Type: NMMapType, Map: map[string]*DataNode{
				"Field5": NewLeafNode(5),
			}},
		}},
		"Field2": {Type: NMSliceType, Slice: []*DataNode{
			NewLeafNode("500"),
			NewLeafNode("502"),
		}},
	}}
	if !reflect.DeepEqual(expnm, onm.nm) {
		t.Errorf("Expected %s ,received: %s", ToJSON(expnm), ToJSON(onm.nm))
	}
}

func TestOrderedNavigableMapField(t *testing.T) {
	nm := &OrderedNavigableMap{nm: &DataNode{Type: NMMapType, Map: map[string]*DataNode{}}}
	if _, err := nm.Field([]string{"Field1"}); err != ErrNotFound {
		t.Error(err)
	}
	nm = &OrderedNavigableMap{nm: &DataNode{Type: NMMapType, Map: map[string]*DataNode{
		"Field1": NewLeafNode("1001"),
		"Field2": NewLeafNode("1003"),
		"Field3": {Type: NMMapType, Map: map[string]*DataNode{"Field4": NewLeafNode("Val")}},
		"Field5": {Type: NMSliceType, Slice: []*DataNode{NewLeafNode(10), NewLeafNode(101)}},
	}}}
	if _, err := nm.Field(nil); err != ErrWrongPath {
		t.Error(err)
	}
	if _, err := nm.Field([]string{"NaN"}); err != ErrNotFound {
		t.Error(err)
	}

	if val, err := nm.Field([]string{"Field1"}); err != nil {
		t.Error(err)
	} else if val.Data != "1001" {
		t.Errorf("Expected %q ,received: %q", "1001", val.Data)
	}

	if _, err := nm.Field([]string{"Field1", "0"}); err != ErrNotFound {
		t.Error(err)
	}
	if val, err := nm.Field([]string{"Field5", "0"}); err != nil {
		t.Error(err)
	} else if val.Data != 10 {
		t.Errorf("Expected %q ,received: %q", 10, val.Data)
	}
	if _, err := nm.Field([]string{"Field3", "0"}); err != ErrNotFound {
		t.Error(err)
	}
	if val, err := nm.Field([]string{"Field3", "Field4"}); err != nil {
		t.Error(err)
	} else if val.Data != "Val" {
		t.Errorf("Expected %q ,received: %q", "Val", val.Data)
	}
}

func TestOrderedNavigableMapEmpty(t *testing.T) {
	nm := &OrderedNavigableMap{nm: &DataNode{Type: NMMapType, Map: map[string]*DataNode{}}}
	if !nm.Empty() {
		t.Error("Expected empty type")
	}
	nm = &OrderedNavigableMap{nm: &DataNode{Type: NMMapType, Map: map[string]*DataNode{"Field1": NewLeafNode("1001")}}}
	if nm.Empty() {
		t.Error("Expected not empty type")
	}
}

func TestOrderedNavigableMapGetSet(t *testing.T) {
	nm := NewOrderedNavigableMap()
	nm.Set(&FullPath{
		PathSlice: []string{"Account", "0"},
		Path:      "Account",
	}, NewLeafNode(1001))
	nm.Set(&FullPath{
		PathSlice: []string{"Account", "1"},
		Path:      "Account",
	}, NewLeafNode("account_on_new_branch"))

	expectedOrder := [][]string{
		{"Account", "0"},
		{"Account", "1"},
	}

	if receivedOrder := nm.GetOrder(); !reflect.DeepEqual(expectedOrder, receivedOrder) {
		t.Errorf("Expected %s ,received: %s", expectedOrder, receivedOrder)
	}
	nm = &OrderedNavigableMap{
		nm: &DataNode{Type: NMMapType, Map: map[string]*DataNode{
			"Field1": NewLeafNode(10),
			"Field2": {Type: NMSliceType, Slice: []*DataNode{
				NewLeafNode("1001"),
				{Type: NMMapType, Map: map[string]*DataNode{
					"Account": {Type: NMSliceType, Slice: []*DataNode{NewLeafNode(10), NewLeafNode(11)}},
				}},
			}},
			"Field3": {Type: NMMapType, Map: map[string]*DataNode{
				"Field4": {Type: NMMapType, Map: map[string]*DataNode{
					"Field5": NewLeafNode(5),
				}},
			}},
		}},
		orderIdx: NewPathItemList(),
		orderRef: make(map[string][]*PathItemElement),
	}
	path := []string{"Field1"}
	if val, err := nm.Field(path); err != nil {
		t.Error(err)
	} else if val.Data != 10 {
		t.Errorf("Expected %q ,received: %q", 10, val.Data)
	}

	path = []string{"Field3", "Field4", "Field5"}
	if val, err := nm.Field(path); err != nil {
		t.Error(err)
	} else if val.Data != 5 {
		t.Errorf("Expected %q ,received: %q", 5, val.Data)
	}

	path = []string{"Field2", "2"}
	if err := nm.Set(&FullPath{Path: strings.Join(path, NestingSep), PathSlice: path}, NewLeafNode("500")); err != nil {
		t.Error(err)
	}
	if val, err := nm.Field(path); err != nil {
		t.Error(err)
	} else if val.Data != "500" {
		t.Errorf("Expected %q ,received: %q", "500", val.Data)
	}

	path = []string{"Field2", "1", "Account"}
	if err := nm.Set(&FullPath{Path: strings.Join(path, NestingSep), PathSlice: path}, NewLeafNode("5")); err != nil {
		t.Error(err)
	}
	path = []string{"Field2", "1", "Account"}
	if val, err := nm.Field(path); err != nil {
		t.Error(err)
	} else if val.Data != "5" {
		t.Errorf("Expected %q ,received: %q", "5", val.Data)
	}
	path = []string{"Field2", "1", "Account", "0"}
	if _, err := nm.Field(path); err != ErrNotFound {
		t.Error(err)
	}
}

func TestOrderedNavigableMapFieldAsInterface(t *testing.T) {
	nm := &OrderedNavigableMap{nm: &DataNode{Type: NMMapType, Map: map[string]*DataNode{
		"Field1": NewLeafNode("1001"),
		"Field2": NewLeafNode("1003"),
		"Field3": {Type: NMMapType, Map: map[string]*DataNode{"Field4": NewLeafNode("Val")}},
		"Field5": {Type: NMSliceType, Slice: []*DataNode{NewLeafNode(10), NewLeafNode(101)}},
	}}}

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
	nm := &OrderedNavigableMap{nm: &DataNode{Type: NMMapType, Map: map[string]*DataNode{
		"Field1": NewLeafNode("1001"),
		"Field2": NewLeafNode("1003"),
		"Field3": {Type: NMMapType, Map: map[string]*DataNode{"Field4": NewLeafNode("Val")}},
		"Field5": {Type: NMSliceType, Slice: []*DataNode{NewLeafNode(10), NewLeafNode(101)}},
	}}}

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
		PathSlice: []string{"Field1", "Field2", "0"},
		Path:      "Field1.Field2[0]",
	}, NewLeafNode("1003"))
	nm.Set(&FullPath{
		PathSlice: []string{"Field1", "Field2", "1"},
		Path:      "Field1.Field2[1]",
	}, NewLeafNode("Val"))
	nm.Set(&FullPath{
		PathSlice: []string{"Field3", "Field4", "Field5", "0"},
		Path:      "Field3.Field4.Field5",
	}, NewLeafNode("1001"))
	nm.Set(&FullPath{
		PathSlice: []string{"Field1", "Field2", "2"},
		Path:      "Field1.Field2[2]",
	}, NewLeafNode(101))
	expected := [][]string{
		{"Field1", "Field2", "0"},
		{"Field1", "Field2", "1"},
		{"Field3", "Field4", "Field5", "0"},
		{"Field1", "Field2", "2"},
	}
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

	path := []string{"Field1", "0"}
	if err := nm.Set(&FullPath{
		PathSlice: path,
		Path:      strings.Join(path, NestingSep),
	}, NewLeafNode("1001")); err != nil {
		t.Error(err)
	}
	nMap := &DataNode{Type: NMMapType, Map: map[string]*DataNode{"Field1": {Type: NMSliceType, Slice: []*DataNode{NewLeafNode("1001")}}}}
	order := [][]string{path}
	if !reflect.DeepEqual(nm.nm, nMap) {
		t.Errorf("Expected %s ,received: %s", ToJSON(nMap), ToJSON(nm))
	}
	if !reflect.DeepEqual(nm.GetOrder(), order) {
		t.Errorf("Expected %s ,received: %s", order, nm.GetOrder())
	}
	if err := nm.Set(&FullPath{
		PathSlice: []string{"Field1", "0", ""},
		Path:      "Field1[0]",
	}, NewLeafNode("1001")); err != ErrWrongPath {
		t.Error(err)
	}

	if err := nm.SetAsSlice(&FullPath{
		PathSlice: []string{"Field1", "0", ""},
		Path:      "Field1[0]",
	}, []*DataNode{}); err != ErrWrongPath {
		t.Error(err)
	}
	if err := nm.SetAsSlice(&FullPath{
		PathSlice: []string{"Field1", "10"},
		Path:      "Field[10]",
	}, []*DataNode{}); err != ErrNotFound {
		t.Error(err)
	}

	nMap = &DataNode{Type: NMMapType, Map: map[string]*DataNode{"Field1": {Type: NMSliceType, Slice: []*DataNode{NewLeafNode("1002")}}}}
	order = [][]string{path}
	if err := nm.Set(&FullPath{
		PathSlice: path,
		Path:      strings.Join(path, NestingSep),
	}, NewLeafNode("1002")); err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(nm.nm, nMap) {
		t.Errorf("Expected %s ,received: %s", ToJSON(nMap), ToJSON(nm))
	}
	if !reflect.DeepEqual(nm.GetOrder(), order) {
		t.Errorf("Expected %s ,received: %s", order, nm.GetOrder())
	}
	path = []string{"Field2"}
	nMap = &DataNode{Type: NMMapType, Map: map[string]*DataNode{
		"Field1": {Type: NMSliceType, Slice: []*DataNode{NewLeafNode("1002")}},
		"Field2": NewLeafNode("1002"),
	}}
	order = append(order, path)
	if err := nm.Set(&FullPath{
		PathSlice: path,
		Path:      strings.Join(path, NestingSep),
	}, NewLeafNode("1002")); err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(nm.nm, nMap) {
		t.Errorf("Expected %s ,received: %s", ToJSON(nMap), ToJSON(nm))
	}
	if !reflect.DeepEqual(nm.GetOrder(), order) {
		t.Errorf("Expected %s ,received: %s", order, nm.GetOrder())
	}
	path = []string{"Field1", "1"}
	nMap = &DataNode{Type: NMMapType, Map: map[string]*DataNode{
		"Field1": {Type: NMSliceType, Slice: []*DataNode{NewLeafNode("1002"), NewLeafNode("1003")}},
		"Field2": NewLeafNode("1002"),
	}}
	order = append(order, path)
	if err := nm.Set(&FullPath{
		PathSlice: path,
		Path:      strings.Join(path, NestingSep),
	}, NewLeafNode("1003")); err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(nm.nm, nMap) {
		t.Errorf("Expected %s ,received: %s", ToJSON(nMap), ToJSON(nm))
	}
	if !reflect.DeepEqual(nm.GetOrder(), order) {
		t.Errorf("Expected %s ,received: %s", order, nm.GetOrder())
	}
	path = []string{"Field3"}
	obj := &DataNode{Type: NMSliceType, Slice: []*DataNode{NewLeafNode("1004"), NewLeafNode("1005")}}
	nMap = &DataNode{Type: NMMapType, Map: map[string]*DataNode{
		"Field1": {Type: NMSliceType, Slice: []*DataNode{NewLeafNode("1002"), NewLeafNode("1003")}},
		"Field2": NewLeafNode("1002"),
		"Field3": obj,
	}}
	order = append(order, []string{"Field3", "0"}, []string{"Field3", "1"})
	if err := nm.SetAsSlice(&FullPath{
		PathSlice: path,
		Path:      strings.Join(path, NestingSep),
	}, obj.Slice); err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(nm.nm, nMap) {
		t.Errorf("Expected %s ,received: %s", ToJSON(nMap), ToJSON(nm))
	}
	if !reflect.DeepEqual(nm.GetOrder(), order) {
		t.Errorf("Expected %s ,received: %s", order, nm.GetOrder())
	}
	obj = &DataNode{Type: NMSliceType, Slice: []*DataNode{NewLeafNode("1005"), NewLeafNode("1006")}}
	nMap = &DataNode{Type: NMMapType, Map: map[string]*DataNode{
		"Field1": {Type: NMSliceType, Slice: []*DataNode{NewLeafNode("1005"), NewLeafNode("1006")}},
		"Field2": NewLeafNode("1002"),
		"Field3": {Type: NMSliceType, Slice: []*DataNode{NewLeafNode("1004"), NewLeafNode("1005")}},
	}}
	order = [][]string{
		{"Field2"},
		{"Field3", "0"},
		{"Field3", "1"},
		{"Field1", "0"},
		{"Field1", "1"},
	}
	if err := nm.SetAsSlice(&FullPath{
		PathSlice: []string{"Field1"},
		Path:      "Field1",
	}, obj.Slice); err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(nm.nm, nMap) {
		t.Errorf("Expected %s ,received: %s", ToJSON(nMap), ToJSON(nm))
	}
	if !reflect.DeepEqual(nm.GetOrder(), order) {
		t.Errorf("Expected %s ,received: %s", order, nm.GetOrder())
	}
	// try dynamic path
	// obj = &DataNode{Type: NMSliceType, Slice: []*DataNode{NewLeafNode("1005"), NewLeafNode("1006")}}
	// nMap = &DataNode{Type: NMMapType, Map: map[string]*DataNode{
	// 	"Field1": {Type: NMSliceType, Slice: []*DataNode{NewLeafNode("1005"), NewLeafNode("1006")}},
	// 	"Field2": NewLeafNode("1002"),
	// 	"Field3": {Type: NMSliceType, Slice: []*DataNode{NewLeafNode("1004"), NewLeafNode("1007")}},
	// }}
	// order = [][]string{
	// 	{"Field2"},
	// 	{"Field3", "0"},
	// 	{"Field1", "0"},
	// 	{"Field1", "1"},
	// 	{"Field3", "1"},
	// }
	// if err := nm.Set(&FullPath{
	// 	PathSlice: []string{"Field3", "-1"},
	// 	Path:      "Field3[-1]",
	// }, NewLeafNode("1007")); err != nil {
	// 	t.Error(err)
	// }
	// if !reflect.DeepEqual(nm.nm, nMap) {
	// 	t.Errorf("Expected %s ,received: %s", ToJSON(nMap), ToJSON(nm))
	// }
	// if !reflect.DeepEqual(nm.GetOrder(), order) {
	// 	t.Errorf("Expected %s ,received: %s", ToJSON(order), ToJSON(nm.GetOrder()))
	// }
}

func TestOrderedNavigableMapRemove(t *testing.T) {
	nm := NewOrderedNavigableMap()
	nm.Set(&FullPath{
		PathSlice: []string{"Field2"},
		Path:      "Field2",
	}, NewLeafNode("1003"))
	nm.Set(&FullPath{
		PathSlice: []string{"Field3", "Field4"},
		Path:      "Field3.Field4",
	}, NewLeafNode("Val"))
	nm.Set(&FullPath{
		PathSlice: []string{"Field1"},
		Path:      "Field1",
	}, NewLeafNode("1001"))
	nm.SetAsSlice(&FullPath{
		PathSlice: []string{"Field5"},
		Path:      "Field5",
	}, []*DataNode{NewLeafNode(10), NewLeafNode(101)})

	if err := nm.Remove(&FullPath{}); err != ErrWrongPath {
		t.Error(err)
	}
	if err := nm.Remove(&FullPath{PathSlice: []string{}}); err != ErrWrongPath {
		t.Error(err)
	}
	if err := nm.Remove(&FullPath{PathSlice: []string{"field"}, Path: "field"}); err != nil {
		t.Error(err)
	}

	if err := nm.Remove(&FullPath{PathSlice: []string{"-1", ""}}); err != ErrWrongPath {
		t.Error(err)
	}
	nMap := &DataNode{Type: NMMapType, Map: map[string]*DataNode{
		"Field1": NewLeafNode("1001"),
		"Field2": NewLeafNode("1003"),
		"Field3": {Type: NMMapType, Map: map[string]*DataNode{"Field4": NewLeafNode("Val")}},
		"Field5": {Type: NMSliceType, Slice: []*DataNode{NewLeafNode(10), NewLeafNode(101)}},
	}}
	order := [][]string{
		{"Field2"},
		{"Field3", "Field4"},
		{"Field1"},
		{"Field5", "0"},
		{"Field5", "1"},
	}
	if !reflect.DeepEqual(nm.nm, nMap) {
		t.Errorf("Expected %s ,received: %s", ToJSON(nMap), ToJSON(nm))
	}
	if !reflect.DeepEqual(nm.GetOrder(), order) {
		t.Errorf("Expected %s ,received: %s", order, nm.GetOrder())
	}
	nMap = &DataNode{Type: NMMapType, Map: map[string]*DataNode{
		"Field1": NewLeafNode("1001"),
		"Field3": {Type: NMMapType, Map: map[string]*DataNode{"Field4": NewLeafNode("Val")}},
		"Field5": {Type: NMSliceType, Slice: []*DataNode{NewLeafNode(10), NewLeafNode(101)}},
	}}
	order = [][]string{
		{"Field3", "Field4"},
		{"Field1"},
		{"Field5", "0"},
		{"Field5", "1"},
	}

	if err := nm.Remove(&FullPath{PathSlice: []string{"Field2"}, Path: "Field2"}); err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(nm.nm, nMap) {
		t.Errorf("Expected %s ,received: %s", ToJSON(nMap), ToJSON(nm))
	}
	if !reflect.DeepEqual(nm.GetOrder(), order) {
		t.Errorf("Expected %s ,received: %s", order, nm.GetOrder())
	}
	nMap = &DataNode{Type: NMMapType, Map: map[string]*DataNode{
		"Field1": NewLeafNode("1001"),
		"Field3": {Type: NMMapType, Map: map[string]*DataNode{"Field4": NewLeafNode("Val")}},
	}}
	order = [][]string{
		{"Field3", "Field4"},
		{"Field1"},
	}

	if err := nm.Remove(&FullPath{PathSlice: []string{"Field5"}, Path: "Field5"}); err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(nm.nm, nMap) {
		t.Errorf("Expected %s ,received: %s", ToJSON(nMap), ToJSON(nm))
	}
	if !reflect.DeepEqual(nm.GetOrder(), order) {
		t.Errorf("Expected %s ,received: %s", order, nm.GetOrder())
	}
	if err := nm.Remove(&FullPath{PathSlice: []string{"Field1", "0", ""}}); err != ErrWrongPath {
		t.Error(err)
	}
}

func TestOrderedNavigableRemote(t *testing.T) {
	nm := &OrderedNavigableMap{nm: &DataNode{Type: NMMapType, Map: map[string]*DataNode{"Field1": NewLeafNode("1001")}}}
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
			if _, err := nm.Set(data.pathItems, NewLeafNode(data.data)); err != nil {
				b.Log(err, data.path)
			}
		}
	}
}

func BenchmarkNavigableMapSet(b *testing.B) {
	nm := &DataNode{Type: NMMapType, Map: map[string]*DataNode{}}
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		for _, data := range gen {
			if _, err := nm.Set(data.pathItems, NewLeafNode(data.data)); err != nil {
				b.Log(err, data.path)
			}
		}
	}
}
func BenchmarkNavigableMapOldSet(b *testing.B) {
	nm := &DataNode{Type: NMMapType, Map: map[string]*DataNode{}}
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		for _, data := range gen {
			if _, err := nm.Set(data.path, data.data); err != nil {
				b.Log(err, data.path)
			}
		}
	}
}

func BenchmarkOrderdNavigableMapFieldAsInterface(b *testing.B) {
	nm := NewOrderedNavigableMap()
	for _, data := range gen {
		if _, err := nm.Set(data.pathItems, NewLeafNode(data.data)); err != nil {
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
	nm := &DataNode{Type: NMMapType, Map: map[string]*DataNode{}}
	for _, data := range gen {
		if _, err := nm.Set(data.pathItems, NewLeafNode(data.data)); err != nil {
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
	nm := &DataNode{Type: NMMapType, Map: map[string]*DataNode{}}
	for _, data := range gen {
		if _, err := nm.Set(data.path, data.data); err != nil {
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
		if _, err := nm.Set(data.pathItems, NewLeafNode(data.data)); err != nil {
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
	nm := &DataNode{Type: NMMapType, Map: map[string]*DataNode{}}
	for _, data := range gen {
		if _, err := nm.Set(data.pathItems, NewLeafNode(data.data)); err != nil {
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
		PathSlice: []string{"Field2"},
		Path:      "Field2",
	}, NewLeafNode("1003"))
	nm.Set(&FullPath{
		PathSlice: []string{"Field3", "Field4"},
		Path:      "Field3.Field4",
	}, NewLeafNode("Val"))
	nm.Set(&FullPath{
		PathSlice: []string{"Field1"},
		Path:      "Field1",
	}, NewLeafNode("1001"))
	nm.SetAsSlice(&FullPath{
		PathSlice: []string{"Field5"},
		Path:      "Field5",
	}, []*DataNode{NewLeafNode(10), NewLeafNode(101)})
	expected := NewOrderedNavigableMap()
	nm.RemoveAll()
	if !reflect.DeepEqual(nm, expected) {
		t.Errorf("Expected %s ,received: %s", expected, nm)
	}
}

func TestOrderedNavigableMapRemove2(t *testing.T) {
	nm := &OrderedNavigableMap{
		nm: &DataNode{Type: NMMapType, Map: map[string]*DataNode{
			"Field1": {Type: NMSliceType, Slice: []*DataNode{}},
		}},
	}
	expErr := `strconv.Atoi: parsing "nan": invalid syntax`
	if err := nm.Remove(&FullPath{PathSlice: []string{"Field1", "nan", ""}, Path: "Field1[nan]"}); err == nil || err.Error() != expErr {
		t.Errorf("Expected error: %s,received: %v", expErr, err)
	}
}

func TestOrderedNavigableMapOrderedFields(t *testing.T) {
	nm := NewOrderedNavigableMap()
	nm.Set(&FullPath{
		PathSlice: []string{"Field1", "Field2", "0"},
		Path:      "Field1.Field2[0]",
	}, NewLeafNode("1003"))
	nm.Set(&FullPath{
		PathSlice: []string{"Field1", "Field3", "0"},
		Path:      "Field1.Field3[0]",
	}, NewLeafNode("1004"))
	nm.Set(&FullPath{
		PathSlice: []string{"Field5"},
		Path:      "Field5",
	}, NewLeafNode("1005"))
	nm.Set(&FullPath{
		PathSlice: []string{"Field6"},
		Path:      "Field6",
	}, NewLeafNode("1006"))
	nm.Remove(&FullPath{
		PathSlice: []string{"Field5"},
		Path:      "Field5",
	})
	exp := []interface{}{"1003", "1004", "1006"}
	rcv := nm.OrderedFields()
	if !reflect.DeepEqual(exp, rcv) {
		t.Errorf("Expected %+v<%T>, received %+v<%T>", exp, exp[0], rcv, rcv[0])
	}
	exp2 := []string{"1003", "1004", "1006"}
	rcv2 := nm.OrderedFieldsAsStrings()
	if !reflect.DeepEqual(exp2, rcv2) {
		t.Errorf("Expected %+v, received %+v", exp2, rcv2)
	}
}
