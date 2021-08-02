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

func TestMissingStructFieldsCorrect(t *testing.T) {
	var attr = struct {
		Tenant          string
		Account         string
		Type            string
		ActionTimingsID string
	}{"bevoip.eu", "danconns0001", META_PREPAID, "mama"}
	if missing := MissingStructFields(&attr,
		[]string{"Tenant", "Account", "Type", "ActionTimingsID"}); len(missing) != 0 {
		t.Error("Found missing field on correct struct", missing)
	}
}

func TestMissingStructFieldsNilCorporate(t *testing.T) {
	tst := &TenantArgWithPaginator{
		Paginator: Paginator{
			Limit: IntPointer(1),
		},
	}
	if missing := MissingStructFields(tst,
		[]string{Tenant}); len(missing) != 1 {
		t.Errorf("TenantIDWithAPIOpts is missing from my struct: %v", missing)
	}
}

func TestMissingStructFieldsNilCorporateTwoStructs(t *testing.T) {
	tst := &struct {
		APIOpts map[string]interface{}
		*TenantID
		*TenantArg
	}{
		TenantID: &TenantID{
			Tenant: "cgrates.org",
		},
		TenantArg: &TenantArg{},
	}
	if missing := MissingStructFields(tst,
		[]string{Tenant}); len(missing) != 1 {
		t.Errorf("TenantIDWithAPIOpts is missing from my struct: %v", missing)
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

func TestMissingMapFields(t *testing.T) {
	var attr = map[string]interface{}{
		Tenant:            "cgrates.org",
		Account:           "1001",
		"Type":            META_PREPAID,
		"ActionTimingsID": "*asap",
	}
	if missing := MissingMapFields(attr,
		[]string{"Tenant", "Account", "Type", "ActionTimingsID"}); len(missing) != 0 {
		t.Error("Found missing field on correct struct", missing)
	}
	attr["ActionTimingsID"] = ""
	delete(attr, "Type")
	expected := []string{"ActionTimingsID", "Type"}
	missing := MissingMapFields(attr,
		[]string{"Tenant", "Account", "Type", "ActionTimingsID"})
	sort.Strings(missing)
	if !reflect.DeepEqual(expected, missing) {
		t.Errorf("Expected %s ,received: %s", expected, missing)
	}
}
