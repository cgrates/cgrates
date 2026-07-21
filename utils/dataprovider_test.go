// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

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
		"Field5": []any{10, 101},
	}
	var expected any = "Field5[1]"
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
		"Field5": []any{10, 101},
	}
	var expected any = "Field5[1]"
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
		"Field6": NewLeafNode(10),
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

func TestMapStringDPFieldAsInterface(t *testing.T) {
	tests := []struct {
		name    string
		mp      MapStringDP
		fldPath []string
		want    any
		expErr  error
	}{
		{
			name: "Path not found",
			mp: MapStringDP{
				"Field1": "1001",
				"Field2": "1003",
			},
			fldPath: []string{"1004"},
			want:    nil,
			expErr:  ErrNotFound,
		},
		{
			name:    "Empty fldPath",
			fldPath: []string{},
			want:    nil,
			expErr:  ErrNotFound,
		},
		{
			mp: MapStringDP{
				"Field1": "1001",
				"Field2": "1003",
			},
			fldPath: []string{"Field1"},
			want:    "1001",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.mp.FieldAsInterface(tt.fldPath)
			if err != nil && err != tt.expErr {
				t.Errorf("Expected %+v, recieved %+v", tt.expErr, err)
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Expected %+v, recieved %+v", tt.want, got)
			}
		})
	}
}

func TestMapStringDPFieldAsString(t *testing.T) {
	tests := []struct {
		name    string
		mp      MapStringDP
		fldPath []string
		want    any
		expErr  error
	}{
		{
			mp: MapStringDP{
				"Field1": "3001",
				"Field2": "3003",
			},
			fldPath: []string{"Field1"},
			want:    "3001",
		},
		{
			name:    "Empty fldPath",
			fldPath: []string{},
			want:    "",
			expErr:  ErrNotFound,
		},
		{
			name: "Path not found",
			mp: MapStringDP{
				"Field1": "2002",
				"Field2": "2003",
			},
			fldPath: []string{"Field1.Field2[0]"},
			want:    "",
			expErr:  ErrNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.mp.FieldAsString(tt.fldPath)
			if err != nil && err != tt.expErr {
				t.Errorf("Expected %+v, recieved %+v", tt.expErr, err)
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Expected %+v, recieved %+v", tt.want, got)
			}
		})
	}
}

func TestMapStringDPString(t *testing.T) {
	mp := MapStringDP{
		"Field1": "2002",
	}
	fldPath := ToJSON(mp)
	got := mp.String()
	if !reflect.DeepEqual(got, fldPath) {
		t.Errorf("Expected %#+v, recieved %#+v", fldPath, got)
	}
}
