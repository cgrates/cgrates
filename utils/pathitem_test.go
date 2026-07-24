// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package utils

import (
	"reflect"
	"strings"
	"testing"
)

func TestStripIdxFromLastPathElm(t *testing.T) {
	str := ""
	if strp := stripIdxFromLastPathElm(str); strp != "" {
		t.Errorf("received: <%s>", strp)
	}
	str = "mypath[0]"
	if strp := stripIdxFromLastPathElm(str); strp != "mypath" {
		t.Errorf("received: <%s>", strp)
	}
	str = "mypath"
	if strp := stripIdxFromLastPathElm(str); strp != "mypath" {
		t.Errorf("received: <%s>", strp)
	}
	str = "mypath.mypath2[0]"
	if strp := stripIdxFromLastPathElm(str); strp != "mypath.mypath2" {
		t.Errorf("received: <%s>", strp)
	}
	str = "mypath.mypath2"
	if strp := stripIdxFromLastPathElm(str); strp != "mypath.mypath2" {
		t.Errorf("received: <%s>", strp)
	}
	str = "mypath[1].mypath2[0]"
	if strp := stripIdxFromLastPathElm(str); strp != "mypath[1].mypath2" {
		t.Errorf("received: <%s>", strp)
	}
	str = "mypath[1].mypath2"
	if strp := stripIdxFromLastPathElm(str); strp != "mypath[1].mypath2" {
		t.Errorf("received: <%s>", strp)
	}
}

func TestNewPathItems(t *testing.T) {
	pathSlice := strings.Split("*req.Field1[0].Account", NestingSep)
	expected := PathItems{{Field: MetaReq}, {Field: "Field1", Index: IntPointer(0)}, {Field: Account}}
	if rply := NewPathItems(pathSlice); !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expected: %s, received: %s", ToJSON(expected), ToJSON(rply))
	}
	pathSlice = []string{}
	expected = PathItems{}
	if rply := NewPathItems(pathSlice); !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expected: %s, received: %s", ToJSON(expected), ToJSON(rply))
	}
}

func TestPathItemString(t *testing.T) {
	path := PathItem{Field: MetaReq}
	expected := MetaReq
	if rply := path.String(); expected != rply {
		t.Errorf("Expected: %q, received: %q", expected, rply)
	}
	path = PathItem{Field: MetaReq, Index: IntPointer(10)}
	expected = MetaReq + "[10]"
	if rply := path.String(); expected != rply {
		t.Errorf("Expected: %q, received: %q", expected, rply)
	}
}

func TestPathItemEqual(t *testing.T) {
	path := PathItem{Field: MetaReq}
	p1 := PathItem{Field: MetaReq}
	if !path.Equal(p1) {
		t.Errorf("Expected %s to be equal to %s", ToJSON(path), ToJSON(p1))
	}
	p1 = PathItem{Field: MetaRep}
	if path.Equal(p1) {
		t.Errorf("Expected %s to not be equal to %s", ToJSON(path), ToJSON(p1))
	}
	p1 = PathItem{Field: MetaReq, Index: IntPointer(0)}
	if path.Equal(p1) {
		t.Errorf("Expected %s to not be equal to %s", ToJSON(path), ToJSON(p1))
	}
	path = PathItem{Field: MetaReq, Index: IntPointer(0)}
	if !path.Equal(p1) {
		t.Errorf("Expected %s to be equal to %s", ToJSON(path), ToJSON(p1))
	}
}

func TestPathItemClone(t *testing.T) {
	path := PathItem{Field: MetaReq, Index: IntPointer(0)}
	expected := PathItem{Field: MetaReq, Index: IntPointer(0)}
	rply := path.Clone()
	*path.Index = 1
	if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expected: %s, received: %s", ToJSON(expected), ToJSON(rply))
	}

	var path2 PathItem
	expected = PathItem{}
	rply = path2.Clone()
	if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expected: %s, received: %s", ToJSON(expected), ToJSON(rply))
	}
}

func TestPathItemsString(t *testing.T) {
	expected := "*req.Field1[0].Account"
	path := NewPathItems(strings.Split(expected, NestingSep))
	if rply := path.String(); expected != rply {
		t.Errorf("Expected: %q, received: %q", expected, rply)
	}
	expected = ""
	path = nil
	if rply := path.String(); expected != rply {
		t.Errorf("Expected: %q, received: %q", expected, rply)
	}
}

func TestPathItemsClone(t *testing.T) {
	path := NewPathItems(strings.Split("*req.Field1[0].Account", NestingSep))
	expected := NewPathItems(strings.Split("*req.Field1[0].Account", NestingSep))
	rply := path.Clone()
	path[0] = PathItem{}
	if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expected: %s, received: %s", ToJSON(expected), ToJSON(rply))
	}
	expected = nil
	path = nil
	rply = path.Clone()
	if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expected: %s, received: %s", ToJSON(expected), ToJSON(rply))
	}
}

func TestPathItemGetIndexPath(t *testing.T) {
	type want struct {
		opath string
		idx   int
	}
	tests := []struct {
		name      string
		arg       string
		want      want
		returnNil bool
	}{
		{
			name:      "argument does not have square brackets",
			arg:       "1",
			want:      want{opath: "1"},
			returnNil: true,
		},
		{
			name:      "check error from strings.Atoi",
			arg:       "[a, b, c]",
			want:      want{opath: "[a, b, c]"},
			returnNil: true,
		},
		{
			name:      "testing GetPathIndex",
			arg:       "test[1]",
			want:      want{opath: "test", idx: 1},
			returnNil: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opath, idx := GetPathIndex(tt.arg)

			if tt.returnNil {
				if opath != tt.want.opath || idx != nil {
					t.Errorf("received '%s' and %v, expexted '%s' and %v", opath, idx, tt.want.opath, tt.want.idx)
				}
			} else {
				expInx := *idx
				if opath != tt.want.opath || expInx != tt.want.idx {
					t.Errorf("received '%s' and %v, expexted '%s' and %v", opath, expInx, tt.want.opath, tt.want.idx)
				}
			}
		})
	}
}

func TestPathItemGetPathWithoutIndex(t *testing.T) {
	tests := []struct {
		name string
		arg  string
		want string
	}{
		{
			name: "argument does not have square brackets",
			arg:  "test",
			want: "test",
		},
		{
			name: "testing GetPathIndex",
			arg:  "test[1]",
			want: "test",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opath := GetPathWithoutIndex(tt.arg)

			if opath != tt.want {
				t.Errorf("received '%s', expexted '%s'", opath, tt.want)
			}
		})
	}
}
