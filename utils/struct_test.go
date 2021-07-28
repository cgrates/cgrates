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
	"math/cmplx"
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
	}{"bevoip.eu", "danconns0001", MetaPrepaid, "mama"}
	if missing := MissingStructFields(&attr,
		[]string{"Tenant", "Account", "Type", "ActionTimingsID"}); len(missing) != 0 {
		t.Error("Found missing field on correct struct", missing)
	}
}

func TestMissingStructFieldsNilCorporate(t *testing.T) {
	tst := &TenantIDWithAPIOpts{
		APIOpts: map[string]interface{}{
			OptsAPIKey: "attr1234",
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
		*TenantAccount
	}{
		APIOpts: map[string]interface{}{
			OptsAPIKey: "attr1234",
		},
		TenantID: &TenantID{
			Tenant: "cgrates.org",
		},
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

func TestMissingMapFields(t *testing.T) {
	var attr = map[string]interface{}{
		Tenant:            "cgrates.org",
		AccountField:      "1001",
		"Type":            MetaPrepaid,
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

func TestMissingStructFieldsAppend(t *testing.T) {
	var attr = struct {
		Tenant          string
		Account         string
		Type            string
		ActionTimingsID string
	}{"", "", MetaPrepaid, ""}
	missing := MissingStructFields(&attr,
		[]string{"Tenant", "Account", "Type", "ActionTimingsID"})
	if len(missing) == 0 {
		t.Error("Required missing field not found")
	}
}

func TestMissingMapFieldsTrim(t *testing.T) {
	var attr = map[string]interface{}{
		"Tenant":  "cgrates.org",
		"Account": "1001",
	}
	if missing := MissingMapFields(attr,
		[]string{"Tenant", "Account"}); len(missing) != 0 {
		t.Error("Found missing field on correct struct", missing)
	}
}

func TestMissingMapFieldsMissing(t *testing.T) {
	var attr = map[string]interface{}{
		"Tenant":  0,
		"Account": 0,
	}
	missing := MissingMapFields(attr, []string{"Tenant", "Account"})
	if len(missing) == 0 {
		t.Error("Required missing field not found")
	}
}

func TestUpdateStructWithIfaceMapValEmpty(t *testing.T) {
	type myStruct struct {
		String string
		Bool   bool
		Float  float64
		Int    int64
	}
	s := new(myStruct)
	mp := map[string]interface{}{
		"String": "",
		"Bool":   "",
		"Float":  "",
		"Int":    "",
	}
	expectedStruct := &myStruct{
		String: "",
		Bool:   false,
		Float:  0,
		Int:    0,
	}
	UpdateStructWithIfaceMap(s, mp)
	if !reflect.DeepEqual(s, expectedStruct) {
		t.Errorf("Expected <%+v> ,received: <%+v>", expectedStruct, s)
	}
}

func TestUpdateStructWithIfaceMapErrorBol(t *testing.T) {
	type myStruct struct {
		String string
		Bool   bool
		Float  float64
		Int    int64
	}
	s := new(myStruct)
	mp := map[string]interface{}{
		"String": "string",
		"Bool":   "cat",
		"Float":  1.2,
		"Int":    1,
	}
	err := UpdateStructWithIfaceMap(s, mp)
	if err == nil || err.Error() != "strconv.ParseBool: parsing \"cat\": invalid syntax" {
		t.Errorf("Expected <strconv.ParseBool: parsing \"cat\": invalid syntax> ,received: <%+v>", err)
	}
}

func TestUpdateStructWithIfaceMapErrorInt(t *testing.T) {
	type myStruct struct {
		String string
		Bool   bool
		Float  float64
		Int    int64
	}
	s := new(myStruct)
	mp := map[string]interface{}{
		"String": "string",
		"Bool":   true,
		"Float":  1.2,
		"Int":    "cat",
	}
	err := UpdateStructWithIfaceMap(s, mp)
	if err == nil || err.Error() != "strconv.ParseInt: parsing \"cat\": invalid syntax" {
		t.Errorf("Expected <strconv.ParseInt: parsing \"cat\": invalid syntax> ,received: <%+v>", err)
	}
}

func TestUpdateStructWithIfaceMapErrorFloat(t *testing.T) {
	type myStruct struct {
		String string
		Bool   bool
		Float  float64
		Int    int64
	}
	s := new(myStruct)
	mp := map[string]interface{}{
		"String": "string",
		"Bool":   true,
		"Float":  "cat",
		"Int":    2,
	}
	err := UpdateStructWithIfaceMap(s, mp)
	if err == nil || err.Error() != "strconv.ParseFloat: parsing \"cat\": invalid syntax" {
		t.Errorf("Expected <strconv.ParseFloat: parsing \"cat\": invalid syntax> ,received: <%+v>", err)
	}
}

func TestUpdateStructWithIfaceMapErrorDefault(t *testing.T) {
	type myStruct struct {
		wrongField1 complex128
	}
	s := new(myStruct)
	mp := map[string]interface{}{
		"wrongField1": cmplx.Sqrt(-5 + 12i),
	}
	err := UpdateStructWithIfaceMap(s, mp)
	if err == nil || err.Error() != "cannot update unsupported struct field: (0+0i)" {
		t.Errorf("Expected <cannot update unsupported struct field: (0+0i)> ,received: <%+v>", err)
	}
}
