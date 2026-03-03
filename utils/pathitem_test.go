/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
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

func TestStripTrailingIndex(t *testing.T) {
	tests := []struct {
		name string
		in   []string
		want []string
	}{
		{name: "empty", in: nil, want: nil},
		{name: "scalar", in: []string{"Account"}, want: []string{"Account"}},
		{name: "trailing index", in: []string{"Account", "0"}, want: []string{"Account"}},
		{name: "nested with index", in: []string{"billing", "Amount", "2"}, want: []string{"billing", "Amount"}},
		{name: "no trailing index", in: []string{"billing", "Amount"}, want: []string{"billing", "Amount"}},
		{name: "negative index stripped", in: []string{"Field", "-1"}, want: []string{"Field"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := StripTrailingIndex(tc.in)
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("StripTrailingIndex(%v) = %v, want %v", tc.in, got, tc.want)
			}
		})
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
