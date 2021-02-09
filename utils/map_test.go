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

func TestConvertMapValStrIf(t *testing.T) {
	var mapIn map[string]string
	var mapOut map[string]interface{}
	if rcv := ConvertMapValStrIf(mapIn); reflect.TypeOf(rcv) != reflect.TypeOf(mapOut) {
		t.Errorf("Expecting: %+v, received: %+v", reflect.TypeOf(mapOut), reflect.TypeOf(rcv))
	}
	mapIn = map[string]string{"test1": "_test1_", "test2": "_test2_"}
	if rcv := ConvertMapValStrIf(mapIn); reflect.TypeOf(rcv) != reflect.TypeOf(mapOut) {
		t.Errorf("Expecting: %+v, received: %+v", reflect.TypeOf(mapOut), reflect.TypeOf(rcv))
	} else if !reflect.DeepEqual(mapIn["test1"], rcv["test1"]) {
		t.Errorf("Expecting: %+v, received: %+v", mapIn["test1"], rcv["test1"])
	} else if len(rcv) != len(mapIn) {
		t.Errorf("Expecting: %+v, received: %+v", len(mapIn), len(rcv))
	}
}

func TestMirrorMap(t *testing.T) {
	var mapIn map[string]string
	if rcv := MirrorMap(mapIn); reflect.DeepEqual(rcv, mapIn) {
		t.Errorf("Expecting: %+v, received: %+v", reflect.TypeOf(mapIn), reflect.TypeOf(rcv))
	} else if len(rcv) != 0 {
		t.Errorf("Expecting: %+v, received: %+v", 0, len(rcv))
	}
	mapIn = map[string]string{"test1": "_test1_", "test2": "_test2_"}
	eOut := map[string]string{"_test1_": "test1", "_test2_": "test2"}
	if rcv := MirrorMap(mapIn); reflect.DeepEqual(rcv, mapIn) {
		t.Errorf("Expecting: %+v, received: %+v", reflect.TypeOf(mapIn), reflect.TypeOf(rcv))
	} else if !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", eOut, rcv)
	}
}

func TestMissingMapKeys(t *testing.T) {
	mapIn := map[string]string{}
	requiredKeys := []string{}
	if rcv := MissingMapKeys(mapIn, requiredKeys); len(rcv) != 0 {
		t.Errorf("Expecting: %+v, received: %+v", 0, len(rcv))
	}

	mapIn = map[string]string{"test1": "_test1_", "test2": "_test2_"}
	requiredKeys = []string{"test1", "test2"}
	if rcv := MissingMapKeys(mapIn, requiredKeys); len(rcv) != 0 {
		t.Errorf("Expecting: %+v, received: %+v", 0, len(rcv))
	}

	mapIn = map[string]string{"test1": "_test1_", "test2": "_test2_"}
	requiredKeys = []string{"test2", "test3"}
	if rcv := MissingMapKeys(mapIn, requiredKeys); len(rcv) != 1 {
		t.Errorf("Expecting: %+v, received: %+v", 1, len(rcv))
	} else if !reflect.DeepEqual([]string{"test3"}, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", []string{"test3"}, rcv)
	}

	requiredKeys = []string{"test3", "test4"}
	eOut := []string{"test3", "test4"}
	if rcv := MissingMapKeys(mapIn, requiredKeys); len(rcv) != 2 {
		t.Errorf("Expecting: %+v, received: %+v", 2, len(rcv))
	} else if !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", eOut, rcv)
	}
}

func TestMapKeys(t *testing.T) {
	mapIn := map[string]string{"test1": "_test1_", "test2": "_test2_"}
	eOut := []string{"test1", "test2"}
	rcv := MapKeys(mapIn)
	sort.Slice(rcv, func(i, j int) bool { return rcv[i] < rcv[j] })
	if !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", eOut, rcv)
	}
	mapIn = map[string]string{"test1": "_test1_", "test2": "_test2_", "test3": "_test3_", "test4": "_test4_"}
	eOut = []string{"test1", "test2", "test3", "test4"}
	rcv = MapKeys(mapIn)
	sort.Slice(rcv, func(i, j int) bool { return rcv[i] < rcv[j] })
	if !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", eOut, rcv)
	}
}

func MapKeysStringMapParse(t *testing.T) {
	if sm := ParseStringMap(EmptyString); len(sm) != 0 {
		t.Errorf("Expecting %+v, received %+v", 0, len(sm))
	}
	if sm := ParseStringMap(ZERO); len(sm) != 0 {
		t.Errorf("Expecting %+v, received %+v", 0, len(sm))
	}
	if sm := ParseStringMap("1;2;3;4"); len(sm) != 4 {
		t.Error("Error pasring map: ", sm)
	}
	if sm := ParseStringMap("1;2;!3;4"); len(sm) != 4 {
		t.Error("Error pasring map: ", sm)
	} else if sm["3"] != false {
		t.Error("Error parsing negative: ", sm)
	}
	sm := ParseStringMap("1;2;!3;4")
	if include, has := sm["2"]; include != true && has != true {
		t.Error("Error detecting positive: ", sm)
	}
	if include, has := sm["3"]; include != false && has != true {
		t.Error("Error detecting negative: ", sm)
	}
	if include, has := sm["5"]; include != false && has != false {
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

func TestFlagsWithParamsGetBool(t *testing.T) {
	flagsWithParams := &FlagsWithParams{
		"test":  []string{"string1", "string2"},
		"test2": []string{"true", "string2"},
		"empty": []string{},
	}
	key := "notpresent"
	if rcv := flagsWithParams.GetBool(key); rcv != false {
		t.Errorf("Expecting: false, received: %+v", ToJSON(rcv))
	}
	key = "empty"
	if rcv := flagsWithParams.GetBool(key); rcv != true {
		t.Errorf("Expecting: false, received: %+v", ToJSON(rcv))
	}
	key = "test"
	if rcv := flagsWithParams.GetBool(key); rcv != false {
		t.Errorf("Expecting: false, received: %+v", ToJSON(rcv))
	}
	key = "test2"
	if rcv := flagsWithParams.GetBool(key); rcv != true {
		t.Errorf("Expecting: true, received: %+v", ToJSON(rcv))
	}
}
func TestFlagsWithParamsValue(t *testing.T) {
	flagsWithParams := &FlagsWithParams{
		"test":  []string{"string2"},
		"empty": []string{},
	}
	key := "notpresent"
	if rcv := flagsWithParams.ParamValue(key); rcv != EmptyString {
		t.Errorf("Expecting: %q, received: %+v", EmptyString, rcv)
	}
	key = "empty"
	if rcv := flagsWithParams.ParamValue(key); rcv != EmptyString {
		t.Errorf("Expecting: %q, received: %+v", EmptyString, rcv)
	}
	key = "test"
	if rcv := flagsWithParams.ParamValue(key); rcv != "string2" {
		t.Errorf("Expecting: string2, received: %+v", rcv)
	}
}
