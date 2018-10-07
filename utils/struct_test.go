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

func TestMissingStructFieldsCorrect(t *testing.T) {
	var attr = struct {
		Tenant          string
		Direction       string
		Account         string
		Type            string
		ActionTimingsId string
	}{"bevoip.eu", "OUT", "danconns0001", META_PREPAID, "mama"}
	if missing := MissingStructFields(&attr,
		[]string{"Tenant", "Direction", "Account", "Type", "ActionTimingsId"}); len(missing) != 0 {
		t.Error("Found missing field on correct struct", missing)
	}
}

func TestStructMapStruct(t *testing.T) {
	type TestStruct struct {
		Name    string
		Surname string
		Address string
		Other   string
	}
	ts := &TestStruct{
		Name:    "1",
		Surname: "2",
		Address: "3",
		Other:   "",
	}
	nts := &TestStruct{
		Name:    "1",
		Surname: "2",
		Address: "3",
		Other:   "",
	}
	m := ToMapStringString(ts)

	FromMapStringString(m, ts)
	if !reflect.DeepEqual(ts, nts) {
		t.Log(m)
		t.Errorf("Expected: %+v got: %+v", ts, nts)
	}
}

func TestMapStructAddStructs(t *testing.T) {
	type TestStruct struct {
		Name    string
		Surname string
		Address string
		Other   string
	}
	ts := &TestStruct{
		Name:    "1",
		Surname: "2",
		Address: "3",
		Other:   "",
	}
	nts := &TestStruct{
		Name:    "1",
		Surname: "2",
		Address: "3",
		Other:   "",
	}
	m := ToMapStringString(ts)
	m["Test"] = "4"
	FromMapStringString(m, ts)

	if !reflect.DeepEqual(ts, nts) {
		t.Log(m)
		t.Errorf("Expected: %+v got: %+v", ts, nts)
	}
}

func TestStructExtraFields(t *testing.T) {
	ts := struct {
		Name        string
		Surname     string
		Address     string
		ExtraFields map[string]string
	}{
		Name:    "1",
		Surname: "2",
		Address: "3",
		ExtraFields: map[string]string{
			"k1": "v1",
			"k2": "v2",
			"k3": "v3",
		},
	}
	efMap := GetMapExtraFields(ts, "ExtraFields")

	if !reflect.DeepEqual(efMap, ts.ExtraFields) {
		t.Errorf("expected: %v got: %v", ts.ExtraFields, efMap)
	}
}

func TestSetStructExtraFields(t *testing.T) {
	ts := struct {
		Name        string
		Surname     string
		Address     string
		ExtraFields map[string]string
	}{
		Name:        "1",
		Surname:     "2",
		Address:     "3",
		ExtraFields: make(map[string]string),
	}
	s := "ExtraFields"
	m := map[string]string{
		"k1": "v1",
		"k2": "v2",
		"k3": "v3",
	}
	SetMapExtraFields(ts, m, s)
	efMap := GetMapExtraFields(ts, "ExtraFields")
	if !reflect.DeepEqual(efMap, ts.ExtraFields) {
		t.Errorf("expected: %v got: %v", ts.ExtraFields, efMap)
	}
}

func TestStructFromMapStringInterface(t *testing.T) {
	ts := &struct {
		Name     string
		Class    *string
		List     []string
		Elements struct {
			Type  string
			Value float64
		}
	}{}
	s := "test2"
	m := map[string]interface{}{
		"Name":  "test1",
		"Class": &s,
		"List":  []string{"test3", "test4"},
		"Elements": struct {
			Type  string
			Value float64
		}{
			Type:  "test5",
			Value: 9.8,
		},
	}
	if err := FromMapStringInterface(m, ts); err != nil {
		t.Logf("ts: %+v", ToJSON(ts))
		t.Error("Error converting map to struct: ", err)
	}
}

func TestStructFromMapStringInterfaceValue(t *testing.T) {
	type T struct {
		Name     string
		Disabled *bool
		Members  []string
	}
	ts := &T{}
	vts := reflect.ValueOf(ts)
	x, err := FromMapStringInterfaceValue(map[string]interface{}{
		"Name":     "test",
		"Disabled": true,
		"Members":  []string{"1", "2", "3"},
	}, vts)
	rt := x.(T)
	if err != nil {
		t.Fatalf("error converting structure value: %v", err)
	}
	if rt.Name != "test" ||
		*rt.Disabled != true ||
		!reflect.DeepEqual(rt.Members, []string{"1", "2", "3"}) {
		t.Errorf("error converting structure value: %s", ToIJSON(rt))
	}
}

func TestUpdateStructWithIfaceMap(t *testing.T) {
	type myStruct struct {
		String string
		Bool   bool
		Float  float64
		Int    int64
	}
	s := new(myStruct)
	mp := map[string]interface{}{
		"String": "s",
		"Bool":   true,
		"Float":  6.4,
		"Int":    2,
	}
	eStruct := &myStruct{
		String: "s",
		Bool:   true,
		Float:  6.4,
		Int:    2,
	}
	if err := UpdateStructWithIfaceMap(s, mp); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eStruct, s) {
		t.Errorf("expecting: %+v, received: %+v", eStruct, s)
	}
	mp = map[string]interface{}{
		"String": "aaa",
		"Bool":   false,
	}
	eStruct = &myStruct{
		String: "aaa",
		Bool:   false,
		Float:  6.4,
		Int:    2,
	}
	if err := UpdateStructWithIfaceMap(s, mp); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eStruct, s) {
		t.Errorf("expecting: %+v, received: %+v", eStruct, s)
	}
}

func TestNonemptyStructFields(t *testing.T) {
	var attr = struct {
		Tenant          string
		Direction       bool
		Account         string
		Type            string
		ActionTimingsId string
	}{"bevoip.eu", true, "testaccount", META_PREPAID, ""}
	mapStruct := NonemptyStructFields(&attr)
	expMapStruct := map[string]interface{}{
		"Tenant":    "bevoip.eu",
		"Direction": true,
		"Account":   "testaccount",
		"Type":      META_PREPAID,
	}
	if !reflect.DeepEqual(expMapStruct, mapStruct) {
		t.Errorf("expecting: %+v, received: %+v", expMapStruct, mapStruct)
	}
}

/*
func TestToMapMapStringInterface(t *testing.T) {
	var attr = struct {
		Tenant    string
		Direction bool
		Account   string
		Type      string
	}{"bevoip.eu", true, "testaccount", META_PREPAID}
	mapStruct := ToMapMapStringInterface(&attr)
	expMapStruct := map[string]interface{}{
		"Tenant":    "bevoip.eu",
		"Direction": true,
		"Account":   "testaccount",
		"Type":      META_PREPAID,
	}
	if !reflect.DeepEqual(expMapStruct, mapStruct) {
		t.Errorf("expecting: %+v, received: %+v", expMapStruct, mapStruct)
	}
}*/
