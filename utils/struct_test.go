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

func TestStructToMapStringString(t *testing.T) {

	tests := []struct {
		name string
		arg  any
		exp  map[string]string
	}{
		{
			name: "struct to map",
			arg: &struct {
				name    string
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

func TestStructGetMapExtraFields(t *testing.T) {

	type in struct {
		mapAny map[string]string
	}

	type args struct {
		in          any
		extraFields string
	}

	tests := []struct {
		name string
		args args
		exp  map[string]string
	}{
		{
			name: "tests get map extra fields",
			args: args{&in{map[string]string{"test": "val1"}}, "mapAny"},
			exp:  map[string]string{"test": "val1"},
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

func TestStructSetMapExtraFields(t *testing.T) {

	type in struct {
		Field map[string]string
	}

	type args struct {
		Values      map[string]string
		extraFields string
	}

	tests := []struct {
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
			SetMapExtraFields(&argIn, tt.args.Values, tt.args.extraFields)

			if argIn.Field["test"] != "" {
				t.Errorf("map value didn't change: %s, expected %s", argIn.Field["test"], tt.args.Values["test"])
			}
		})
	}
}

func TestStructFromMapStringString(t *testing.T) {

	type args struct {
		m  map[string]string
		in any
	}

	type in struct {
		Field string
	}

	inArg := in{""}

	tests := []struct {
		name string
		args args
		exp  in
	}{
		{
			name: "test map string string",
			args: args{map[string]string{"Field": ""}, &inArg},
			exp:  in{Field: ""},
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

func TestStructFromMapStringInterface(t *testing.T) {

	type args struct {
		m  map[string]any
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

	tests := []struct {
		name string
		args args
		exp  error
	}{
		{
			name: "test from map string interface",
			args: args{map[string]any{"Field": ""}, &inArg},
			exp:  nil,
		},
		{
			name: "test from map string interface",
			args: args{map[string]any{"Field": 1}, &inArg},
			exp:  nil,
		},
		{
			name: "invalid field",
			args: args{map[string]any{"field": 1}, &inC},
			exp:  nil,
		},
	}

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			err := FromMapStringInterface(tt.args.m, tt.args.in)

			if i == 1 {
				if err == nil {
					t.Error("was expecting an error")
				}
			} else {
				if err != tt.exp {
					t.Errorf("recived %s, expected %s", err, tt.exp)
				}
			}
		})
	}
}

func TestStructUpdateStructWithStrMap(t *testing.T) {

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

	tests := []struct {
		name string
		args args
		exp  []string
	}{
		{
			name: "bool case",
			args: args{&arg, map[string]string{"Field1": "true", "Field2": "2", "Field3": "val2"}},
			exp:  []string{},
		},
		{
			name: "bool case",
			args: args{&arg, map[string]string{"Field": "1.8"}},
			exp:  []string{"Field"},
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

func TestStructUpdateStructWithIfaceMap(t *testing.T) {
	type Test struct {
		Bl bool
		Nm int
		Fl float64
		Df []byte
	}
	test := Test{Bl: true}
	test2 := Test{Nm: 1}
	test3 := Test{Fl: 1.2}
	test4 := Test{Df: []byte{}}
	type args struct {
		s  any
		mp map[string]any
	}
	tests := []struct {
		name string
		args args
		err  string
	}{
		{
			name: "auto populate bool",
			args: args{&test, map[string]any{"Bl": ""}},
			err:  "",
		},
		{
			name: "interface as bool error",
			args: args{&test, map[string]any{"Bl": []byte{}}},
			err:  "cannot convert field: [] to bool",
		},
		{
			name: "auto populate int",
			args: args{&test2, map[string]any{"Nm": ""}},
			err:  "",
		},
		{
			name: "interface as int error",
			args: args{&test2, map[string]any{"Nm": []byte{}}},
			err:  "cannot convert field: [] to int",
		},
		{
			name: "auto populate float64",
			args: args{&test3, map[string]any{"Fl": ""}},
			err:  "",
		},
		{
			name: "interface as flaot64 error",
			args: args{&test3, map[string]any{"Fl": []byte{}}},
			err:  "cannot convert field: [] to float64",
		},
		{
			name: "default",
			args: args{&test4, map[string]any{"Df": ""}},
			err:  "cannot update unsupported struct field: []",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := UpdateStructWithIfaceMap(tt.args.s, tt.args.mp)

			if tt.err != "" {
				if err != nil {
					if err.Error() != tt.err {
						t.Error(err)
					}
				} else {
					t.Error("was expecting an error")
				}
			}
		})
	}
}

func TestToMapMapStringInterface(t *testing.T) {

	type TestStruct struct {
		Field1 string
		Field2 int
		Field3 bool
	}
	input := TestStruct{
		Field1: "value1",
		Field2: 42,
		Field3: true,
	}
	expected := map[string]any{
		"Field1": "value1",
		"Field2": 42,
		"Field3": true,
	}
	output := ToMapMapStringInterface(input)
	if !reflect.DeepEqual(output, expected) {
		t.Errorf("expected %v, got %v", expected, output)
	}
}
