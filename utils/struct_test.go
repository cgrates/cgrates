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
		APIOpts map[string]any
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
	mp := map[string]any{
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
	mp = map[string]any{
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
		//Amount			int
	}{"bevoip.eu", true, "testaccount", META_PREPAID, ""}
	mapStruct := NonemptyStructFields(&attr)
	expMapStruct := map[string]any{
		"Tenant":    "bevoip.eu",
		"Direction": true,
		"Account":   "testaccount",
		"Type":      META_PREPAID,
		//"Amount":	 10,
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
	expMapStruct := map[string]any{
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
	var attr = map[string]any{
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

func TestToMapStringString(t *testing.T) {

	tests := []struct{
		name string
		arg any
		exp map[string]string
	}{
		{
			name: "struct to map",
			arg: &struct{
				name string
				surname string
			}{"test", "test2"},
			exp: map[string]string{"name": "test", "surname": "test2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rcv := ToMapStringString(tt.arg)

			if !reflect.DeepEqual(rcv, tt.exp) {
				t.Errorf("expected %v, recived %v", tt.exp, rcv)
			}
		})
	}
}

func TestGetMapExtraFields(t *testing.T) {

	type in struct {
		mapAny map[string]string
	}

	type args struct {
		in any
		extraFields string
	}

	tests := []struct{
		name string
		args args
		exp map[string]string
	}{
		{
			name: "tests get map extra fields",
			args: args{&in{map[string]string{"test": "val1"}}, "mapAny"},
			exp: map[string]string{"test": "val1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rcv := GetMapExtraFields(tt.args.in, tt.args.extraFields)

			if !reflect.DeepEqual(rcv, tt.exp) {
				t.Errorf("recived %v, expected %v", rcv, tt.exp)
			}
		})
	}
}

func TestSetMapExtraFields(t *testing.T) {

	type in struct {
		Field map[string]string
	}
	
	type args struct {
		Values map[string]string
		extraFields string
	}

	tests := []struct{
		name string
		args args
	}{
		{
			name: "set map extra fields",
			args: args{map[string]string{"test": ""}, "Field"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			argIn := in{map[string]string{"test": "val1"}}
			SetMapExtraFields(&argIn , tt.args.Values, tt.args.extraFields)


			if argIn.Field["test"] != "" {
				t.Errorf("map value didn't change: %s, expected %s", argIn.Field["test"], tt.args.Values["test"])
			}
		})
	}
}

func TestFromMapStringString(t *testing.T) {

	type args struct {
		m map[string]string
		in any
	}

	type in struct {
		Field string
	}

	inArg := in{""}

	tests := []struct{
		name string
		args args
		exp in
	}{
		{
			name: "test map string string",
			args: args{map[string]string{"Field": ""}, &inArg},
			exp: in{Field: ""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			FromMapStringString(tt.args.m, tt.args.in)

			//needs fix
			if !reflect.DeepEqual(tt.exp, inArg) {
				t.Errorf("expected %v, reciving %v", tt.exp, inArg)
			}
		})
	}
}

func TestFromMapStringInterface(t *testing.T) {

	type args struct {
		m map[string]any
		in any
	}

	type in struct {
		Field string
	}

	type inCantSet struct {
		field string
	}

	inArg := in{""}

	inC := inCantSet{""}

	tests := []struct{
		name string
		args args
		exp error
	}{
		{
			name: "test from map string interface",
			args: args{map[string]any{"Field": ""}, &inArg},
			exp: nil,
		},
		{
			name: "test from map string interface",
			args: args{map[string]any{"Field": 1}, &inArg},
			exp: ErrTypeDidntMatch,
		},
		{
			name: "invalid field",
			args: args{map[string]any{"field": 1}, &inC},
			exp: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			err := FromMapStringInterface(tt.args.m, tt.args.in)

			if err != tt.exp {
				t.Errorf("recived %s, expected %s", err, tt.exp)
			}  
		})
	}
}

func TestUpdateStructWithStrMap(t *testing.T) {

	type argStruct struct {
		Field1 bool
		Field2 int
		Field3 string
		Field4 float64
	}

	arg := argStruct{false, 1, "val1", 1.5}

	type args struct {
		s any
		m map[string]string
	}

	tests := []struct{
		name string
		args args
		exp []string
	}{
		{
			name: "bool case",
			args: args{&arg, map[string]string{"Field1": "true", "Field2": "2", "Field3": "val2"}},
			exp: []string{},
		},
		{
			name: "bool case",
			args: args{&arg, map[string]string{"Field": "1.8"}},
			exp: []string{"Field"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			
			rcv := UpdateStructWithStrMap(tt.args.s, tt.args.m)

			if !reflect.DeepEqual(rcv, tt.exp) {
				t.Errorf("recived %v, expecte %v", rcv, tt.exp)
			}
		})
	}
}
