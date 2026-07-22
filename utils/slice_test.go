// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package utils

import (
	"reflect"
	"sort"
	"testing"
)

func TestHasPrefixSlice(t *testing.T) {
	if !HasPrefixSlice([]string{"1", "2", "3", "4", "5"}, "123") {
		t.Error("Expecting: true, received: false")
	}
	if HasPrefixSlice([]string{"1", "2", "3", "4", "5"}, "689") {
		t.Error("Expecting: true, received: false")
	}
}

func TestPrefixSliceItems(t *testing.T) {
	rcv := PrefixSliceItems("*", []string{"1", "2", "3", "", "5"})
	sort.Strings(rcv)
	eOut := []string{"*1", "*2", "*3", "*5"}
	if !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", eOut, rcv)
	}
}

func TestSliceStringToIface(t *testing.T) {
	exp := []any{"*default", "ToR", "*voice"}
	if rply := SliceStringToIface([]string{"*default", "ToR", "*voice"}); !reflect.DeepEqual(exp, rply) {
		t.Errorf("Expected: %s ,received: %s", ToJSON(exp), ToJSON(rply))
	}
}
