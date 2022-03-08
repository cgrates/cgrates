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

func TestNewFullPath(t *testing.T) {
	expected := &FullPath{
		PathSlice: []string{"test", "path"},
		Path:      "test.path",
	}
	if rcv := NewFullPath("test.path"); !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Expected %+v \n, received %+v", ToJSON(expected), ToJSON(rcv))
	}
}
