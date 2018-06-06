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

func TestStringMapParse(t *testing.T) {
	sm := ParseStringMap("1;2;3;4")
	if len(sm) != 4 {
		t.Error("Error pasring map: ", sm)
	}
}

func TestStringMapParseNegative(t *testing.T) {
	sm := ParseStringMap("1;2;!3;4")
	if len(sm) != 4 {
		t.Error("Error pasring map: ", sm)
	}
	if sm["3"] != false {
		t.Error("Error parsing negative: ", sm)
	}
}

func TestStringMapCompare(t *testing.T) {
	sm := ParseStringMap("1;2;!3;4")
	if include, found := sm["2"]; include != true && found != true {
		t.Error("Error detecting positive: ", sm)
	}
	if include, found := sm["3"]; include != false && found != true {
		t.Error("Error detecting negative: ", sm)
	}
	if include, found := sm["5"]; include != false && found != false {
		t.Error("Error detecting missing: ", sm)
	}
}

func TestMapMergeMapsStringIface(t *testing.T) {
	mp1 := map[string]interface{}{
		"Hdr1": "Val1",
		"Hdr2": "Val2",
		"Hdr3": "Val3",
	}
	mp2 := map[string]interface{}{
		"Hdr3": "Val4",
		"Hdr4": "Val4",
	}
	eMergedMap := map[string]interface{}{
		"Hdr1": "Val1",
		"Hdr2": "Val2",
		"Hdr3": "Val4",
		"Hdr4": "Val4",
	}
	if mergedMap := MergeMapsStringIface(mp1, mp2); !reflect.DeepEqual(eMergedMap, mergedMap) {
		t.Errorf("Expecting: %+v, received: %+v", eMergedMap, mergedMap)
	}
}

func TestEqual(t *testing.T) {
	t1 := NewStringMap("val1")
	t2 := NewStringMap("val2")
	result := t1.Equal(t2)
	expected := false
	if result != expected {
		t.Error("Expecting:", expected, ", received:", result)
	}
}

func TestIsEmpty(t *testing.T) {
	t1 := NewStringMap("val1")
	result := t1.IsEmpty()
	expected := false
	if result != expected {
		t.Error("Expecting:", expected, ", received:", result)
	}
}

func TestMapStringToInt64(t *testing.T) {
	t1 := map[string]int64{"test": int64(21)}
	t2 := map[string]string{"test": "21"}
	t3, err := MapStringToInt64(t2)
	if err != nil {
		t.Error("Got Error: ", err)
	}
	if !reflect.DeepEqual(t1, t3) {
		t.Errorf("Expecting: %+v, received: %+v", t1, t3)
	}
}

func TestMapHasKey(t *testing.T) {
	mp := ParseStringMap("Item1;Item2;Item3")
	if mp.HasKey("Item1") != true {
		t.Errorf("Expecting: true, received: %+v", mp.HasKey("Item1"))
	}
	if mp.HasKey("Item4") != false {
		t.Errorf("Expecting: true, received: %+v", mp.HasKey("Item4"))
	}

}
