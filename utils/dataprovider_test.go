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

func TestDPDynamicInterface(t *testing.T) {
	nm := MapStorage{
		"Field1": "1001",
		"Field2": "1003",
		"Field3": MapStorage{"Field4": "Val"},
		"Field5": []interface{}{10, 101},
	}
	var expected interface{} = "Field5[1]"
	if rply, err := DPDynamicInterface("Field5[1]", nm); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expected %q ,received: %q", expected, rply)
	}

	expected = 101
	if rply, err := DPDynamicInterface("~Field5[1]", nm); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expected %v ,received: %v", expected, rply)
	}

}

func TestDPDynamicString(t *testing.T) {
	nm := MapStorage{
		"Field1": "1001",
		"Field2": "1003",
		"Field3": MapStorage{"Field4": "Val"},
		"Field5": []interface{}{10, 101},
	}
	var expected interface{} = "Field5[1]"
	if rply, err := DPDynamicString("Field5[1]", nm); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expected %q ,received: %q", expected, rply)
	}

	expected = "101"
	if rply, err := DPDynamicString("~Field5[1]", nm); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expected %v ,received: %v", expected, rply)
	}

}

func TestAppendNavMapVal(t *testing.T) {
	onm := NewOrderedNavigableMap()
	nm := &DataNode{Type: NMMapType, Map: map[string]*DataNode{
		"Field1": NewLeafNode("1001"),
		"Field2": NewLeafNode("1003"),
		"Field3": {Type: NMMapType, Map: map[string]*DataNode{"Field4": NewLeafNode("Val")}},
		"Field5": {Type: NMSliceType, Slice: []*DataNode{NewLeafNode(10), NewLeafNode(101)}},
	}}
	onm.nm = nm
	expected := &DataNode{Type: NMMapType, Map: map[string]*DataNode{
		"Field1": NewLeafNode("1001"),
		"Field2": NewLeafNode("1003"),
		"Field3": {Type: NMMapType, Map: map[string]*DataNode{"Field4": NewLeafNode("Val")}},
		"Field5": {Type: NMSliceType, Slice: []*DataNode{NewLeafNode(10), NewLeafNode(101), NewLeafNode(18)}},
	}}
	if err := onm.Append(&FullPath{Path: "Field5", PathSlice: []string{"Field5"}}, NewLeafNode(18).Value); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, onm.nm) {
		t.Errorf("Expected %v ,received: %v", expected, onm.nm)
	}

	if err := onm.Append(&FullPath{}, NewLeafNode(18).Value); err != ErrWrongPath {
		t.Errorf("Expected error: %s received: %v", ErrWrongPath, err)
	}
}

func TestComposeNavMapVal(t *testing.T) {
	onm := NewOrderedNavigableMap()
	nm := &DataNode{Type: NMMapType, Map: map[string]*DataNode{
		"Field4": {Type: NMSliceType, Slice: []*DataNode{}},
		"Field5": {Type: NMSliceType, Slice: []*DataNode{NewLeafNode(10), NewLeafNode(101)}},
	}}
	onm.nm = nm
	if err := onm.Compose(&FullPath{Path: "Field4", PathSlice: []string{"Field4", "10"}}, NewLeafNode(18).Value); err != ErrNotFound {
		t.Error(err)
	}
	expected := &DataNode{Type: NMMapType, Map: map[string]*DataNode{
		"Field4": {Type: NMSliceType, Slice: []*DataNode{}},
		"Field5": {Type: NMSliceType, Slice: []*DataNode{NewLeafNode(10), NewLeafNode("10118")}},
	}}
	if err := onm.Compose(&FullPath{Path: "Field5", PathSlice: []string{"Field5"}}, NewLeafNode(18).Value); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, nm) {
		t.Errorf("Expected %v ,received: %v", expected, nm)
	}

	expected = &DataNode{Type: NMMapType, Map: map[string]*DataNode{
		"Field4": {Type: NMSliceType, Slice: []*DataNode{}},
		"Field5": {Type: NMSliceType, Slice: []*DataNode{NewLeafNode(10), NewLeafNode("10118")}},
		"Field6": {Type: NMSliceType, Slice: []*DataNode{NewLeafNode(10)}},
	}}
	if err := onm.Compose(&FullPath{Path: "Field6", PathSlice: []string{"Field6"}}, NewLeafNode(10).Value); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, nm) {
		t.Errorf("Expected %v ,received: %v", expected, nm)
	}

	onm.nm = &DataNode{Type: NMMapType, Map: map[string]*DataNode{
		"Field4": NewLeafNode(1),
		"Field5": {Type: NMSliceType, Slice: []*DataNode{NewLeafNode(10), NewLeafNode(101)}},
	}}

	if err := onm.Compose(&FullPath{}, NewLeafNode(18).Value); err != ErrWrongPath {
		t.Errorf("Expected error: %s received: %v", ErrWrongPath, err)
	}
}

func TestIsPathValid(t *testing.T) {
	path := "Field1.Field2[0]"
	if err := IsPathValid(path); err != nil {
		t.Error(err)
	}

	///
	path = "~Field1"
	errExpect := "Path is missing "
	if err := IsPathValid(path); err == nil || err.Error() != errExpect {
		t.Errorf("Expected %v but received %v", errExpect, err)
	}

	///
	path = "~Field1.\n\t.Field2[0]"
	errExpect = "Empty field path "
	if err := IsPathValid(path); err == nil || err.Error() != errExpect {
		t.Errorf("Expected %v but received %v", errExpect, err)
	}

	///
	path = "~Field1.Field2[0]"
	if err := IsPathValid(path); err != nil {
		t.Error(err)
	}
}

func TestIsPathValidForExporters(t *testing.T) {
	path := "Field1.Field2[0]"
	if err := IsPathValidForExporters(path); err != nil {
		t.Error(err)
	}

	///
	path = "~Field1.\n\t.Field2[0]"
	errExpect := "Empty field path "
	if err := IsPathValidForExporters(path); err == nil || err.Error() != errExpect {
		t.Errorf("Expected %v but received %v", errExpect, err)
	}

	///
	path = "~Field1.Field2[0]"
	if err := IsPathValidForExporters(path); err != nil {
		t.Error(err)
	}
}

func TestCheckInLineFilter(t *testing.T) {
	fltrs := []string{"Test1", "Test2"}
	if err := CheckInLineFilter(fltrs); err != nil {
		t.Error(err)
	}

	///
	fltrs = []string{"*Test1", "*Test2"}
	errExpect := "inline parse error for string: <*Test1>"
	if err := CheckInLineFilter(fltrs); err == nil || err.Error() != errExpect {
		t.Errorf("Expected %v but received %v", errExpect, err)
	}

	///
	fltrs = []string{"*Test1:*Test2:*Test3:*Test4"}
	if err := CheckInLineFilter(fltrs); err != nil {
		t.Error(err)
	}

	///
	fltrs = []string{"*empty:~Field1..Field2[0]:*Test3:*Test4"}
	errExpect = "Empty field path  for <*empty:~Field1..Field2[0]:*Test3:*Test4>"
	if err := CheckInLineFilter(fltrs); err == nil || err.Error() != errExpect {
		t.Errorf("Expected %v but received %v", errExpect, err)
	}

	///
	fltrs = []string{"*empty:~Field1.Field2[0]:~Field1..Field2[0]|Test4"}
	errExpect = "Empty field path  for <*empty:~Field1.Field2[0]:~Field1..Field2[0]|Test4>"
	if err := CheckInLineFilter(fltrs); err == nil || err.Error() != errExpect {
		t.Errorf("Expected %v but received %v", errExpect, err)
	}
}
