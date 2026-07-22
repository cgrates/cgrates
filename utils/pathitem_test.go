// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

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
