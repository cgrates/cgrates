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
	expected := PathItems{{Field: MetaReq}, {Field: "Field1", Index: []string{"0"}}, {Field: AccountField}}
	if rply := NewPathItems(pathSlice); !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expected: %s, received: %s", ToJSON(expected), ToJSON(rply))
	}
	pathSlice = []string{}
	expected = PathItems{}
	if rply := NewPathItems(pathSlice); !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expected: %s, received: %s", ToJSON(expected), ToJSON(rply))
	}

	pathSlice = strings.Split("*req.Field1[*raw][0].Account", NestingSep)
	expected = PathItems{{Field: MetaReq}, {Field: "Field1", Index: []string{"*raw", "0"}}, {Field: AccountField}}
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
	path = PathItem{Field: MetaReq, Index: []string{"10"}}
	expected = MetaReq + "[10]"
	if rply := path.String(); expected != rply {
		t.Errorf("Expected: %q, received: %q", expected, rply)
	}
}

func TestPathItemClone(t *testing.T) {
	path := PathItem{Field: MetaReq, Index: []string{"0"}}
	expected := PathItem{Field: MetaReq, Index: []string{"0"}}
	rply := path.Clone()
	path.Index[0] = "1"
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
func TestGetPathWithoutIndex(t *testing.T) {
	expected := "field"
	if rply := GetPathWithoutIndex("field[index]"); expected != rply {
		t.Errorf("Expected: %s, received: %s", expected, rply)
	}
	expected = "field[index"
	if rply := GetPathWithoutIndex("field[index"); expected != rply {
		t.Errorf("Expected: %s, received: %s", expected, rply)
	}
}

func TestPathItemsSlice(t *testing.T) {
	expected := []string{"*req", "Field1[0]", "Account"}
	path := NewPathItems(expected)
	if rply := path.Slice(); !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expected: %q, received: %q", expected, rply)
	}
}

func TestNewFullPath(t *testing.T) {
	expected := &FullPath{
		PathItems: []PathItem{
			{
				Field: "test",
			},
			{
				Field: "path",
			},
		},
		Path: "test.path",
	}
	if rcv := NewFullPath("test.path", NestingSep); !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Expected %+v \n, received %+v", ToJSON(expected), ToJSON(rcv))
	}
}
