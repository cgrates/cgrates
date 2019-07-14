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
	"sort"
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

func TestMapSubsystemIDsFromSlice(t *testing.T) {
	sls := []string{"*event", "*thresholds:ID1;ID2;ID3", "*attributes", "*stats:ID"}
	eMp := FlagsWithParams{
		"*event":      []string{},
		"*thresholds": []string{"ID1", "ID2", "ID3"},
		"*attributes": []string{},
		"*stats":      []string{"ID"},
	}
	if mp, err := FlagsWithParamsFromSlice(sls); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(mp, eMp) {
		t.Errorf("Expecting: %+v, received: %+v", eMp, mp)
	}
}

func TestMapSubsystemIDsFromSliceWithErr(t *testing.T) {
	sls := []string{"*event", "*thresholds:ID1;ID2;ID3:error:", "*attributes", "*stats:ID"}

	if _, err := FlagsWithParamsFromSlice(sls); err != ErrUnsupportedFormat {
		t.Error(err)
	}
}

func TestMapSubsystemIDsHasKey(t *testing.T) {
	sls := []string{"*event", "*thresholds:ID1;ID2;ID3", "*attributes", "*stats:ID"}
	eMp := FlagsWithParams{
		"*event":      []string{},
		"*thresholds": []string{"ID1", "ID2", "ID3"},
		"*attributes": []string{},
		"*stats":      []string{"ID"},
	}
	mp, err := FlagsWithParamsFromSlice(sls)
	if err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(mp, eMp) {
		t.Errorf("Expecting: %+v, received: %+v", eMp, mp)
	}
	if has := mp.HasKey("*event"); !has {
		t.Errorf("Expecting: true, received: %+v", has)
	}
	if has := mp.HasKey("*thresholds"); !has {
		t.Errorf("Expecting: true, received: %+v", has)
	}
	if has := mp.HasKey("*resources"); has {
		t.Errorf("Expecting: false, received: %+v", has)
	}
}

func TestMapSubsystemIDsGetIDs(t *testing.T) {
	sls := []string{"*event", "*thresholds:ID1;ID2;ID3", "*attributes", "*stats:ID"}
	eMp := FlagsWithParams{
		"*event":      []string{},
		"*thresholds": []string{"ID1", "ID2", "ID3"},
		"*attributes": []string{},
		"*stats":      []string{"ID"},
	}
	mp, err := FlagsWithParamsFromSlice(sls)
	if err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(mp, eMp) {
		t.Errorf("Expecting: %+v, received: %+v", eMp, mp)
	}
	eIDs := []string{"ID1", "ID2", "ID3"}
	if ids := mp.ParamsSlice("*thresholds"); !reflect.DeepEqual(ids, eIDs) {
		t.Errorf("Expecting: %+v, received: %+v", eIDs, ids)
	}
	eIDs = []string{}
	if ids := mp.ParamsSlice("*event"); !reflect.DeepEqual(ids, eIDs) {
		t.Errorf("Expecting: %+v, received: %+v", eIDs, ids)
	}
	eIDs = nil
	if ids := mp.ParamsSlice("*test"); !reflect.DeepEqual(ids, eIDs) {
		t.Errorf("Expecting: %+v, received: %+v", eIDs, ids)
	}
}

func TestFlagsToSlice(t *testing.T) {
	sls := []string{"*event", "*thresholds:ID1;ID2;ID3", "*attributes", "*stats:ID"}
	eMp := FlagsWithParams{
		"*event":      []string{},
		"*thresholds": []string{"ID1", "ID2", "ID3"},
		"*attributes": []string{},
		"*stats":      []string{"ID"},
	}
	mp, err := FlagsWithParamsFromSlice(sls)
	if err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(mp, eMp) {
		t.Errorf("Expecting: %+v, received: %+v", eMp, mp)
	}
	sort.Strings(sls)
	flgSls := mp.SliceFlags()
	sort.Strings(flgSls)
	if !reflect.DeepEqual(flgSls, sls) {
		t.Errorf("Expecting: %+v, received: %+v", sls, flgSls)
	}
}
