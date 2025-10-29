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
	"fmt"
	"net"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestReflectFieldAsStringOnStruct(t *testing.T) {
	mystruct := struct {
		Title       string
		Count       int
		Count64     int64
		Val         float64
		ExtraFields map[string]any
	}{"Title1", 5, 6, 7.3, map[string]any{"a": "Title2", "b": 15, "c": int64(16), "d": 17.3}}
	if strVal, err := ReflectFieldAsString(mystruct, "Title", "ExtraFields"); err != nil {
		t.Error(err)
	} else if strVal != "Title1" {
		t.Errorf("Received: %s", strVal)
	}
	if strVal, err := ReflectFieldAsString(mystruct, "Count", "ExtraFields"); err != nil {
		t.Error(err)
	} else if strVal != "5" {
		t.Errorf("Received: %s", strVal)
	}
	if strVal, err := ReflectFieldAsString(mystruct, "Count64", "ExtraFields"); err != nil {
		t.Error(err)
	} else if strVal != "6" {
		t.Errorf("Received: %s", strVal)
	}
	if strVal, err := ReflectFieldAsString(mystruct, "Val", "ExtraFields"); err != nil {
		t.Error(err)
	} else if strVal != "7.3" {
		t.Errorf("Received: %s", strVal)
	}
	if strVal, err := ReflectFieldAsString(mystruct, "a", "ExtraFields"); err != nil {
		t.Error(err)
	} else if strVal != "Title2" {
		t.Errorf("Received: %s", strVal)
	}
	if strVal, err := ReflectFieldAsString(mystruct, "b", "ExtraFields"); err != nil {
		t.Error(err)
	} else if strVal != "15" {
		t.Errorf("Received: %s", strVal)
	}
	if strVal, err := ReflectFieldAsString(mystruct, "c", "ExtraFields"); err != nil {
		t.Error(err)
	} else if strVal != "16" {
		t.Errorf("Received: %s", strVal)
	}
	if strVal, err := ReflectFieldAsString(mystruct, "d", "ExtraFields"); err != nil {
		t.Error(err)
	} else if strVal != "17.3" {
		t.Errorf("Received: %s", strVal)
	}
}

func TestReflectFieldInterface(t *testing.T) {
	type args struct {
		intf             any
		fldName          string
		extraFieldsLabel string
	}
	type want struct {
		retIf any
		err   error
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "pointer map with unmatching filed name",
			args: args{intf: &map[string]int{"test": 1}, fldName: "test1", extraFieldsLabel: ""},
			want: want{retIf: nil, err: ErrNotFound},
		},
		{
			name: "string as argument",
			args: args{intf: "testtest", fldName: "test1", extraFieldsLabel: ""},
			want: want{retIf: nil, err: fmt.Errorf("Unsupported field kind: string")},
		},
		{
			name: "struct with unmatching field name and extra field label as empty string",
			args: args{intf: struct{ Test string }{Test: "test"}, fldName: "Test1", extraFieldsLabel: ""},
			want: want{retIf: nil, err: ErrNotFound},
		},
		{
			name: "struct with unmatching field name and unmatching extra field label",
			args: args{intf: struct{ Test string }{Test: "test"}, fldName: "Test1", extraFieldsLabel: "Test2"},
			want: want{retIf: nil, err: ErrNotFound},
		},
		{
			name: "struct with unmatching field name and unmatching extra field label",
			args: args{intf: struct{ Test map[string]string }{Test: map[string]string{"Test": "test"}}, fldName: "Test1", extraFieldsLabel: "Test"},
			want: want{retIf: nil, err: ErrNotFound},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rcv, err := ReflectFieldInterface(tt.args.intf, tt.args.fldName, tt.args.extraFieldsLabel)

			if err.Error() != tt.want.err.Error() {
				t.Fatal("wrong error message or no error")
			}

			if rcv != tt.want.retIf {
				t.Errorf("reciving %v, expected %v", rcv, tt.want.retIf)
			}
		})
	}

}

func TestReflectFieldAsString(t *testing.T) {
	type args struct {
		intf             any
		fldName          string
		extraFieldsLabel string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "check error",
			args: args{intf: struct{ Test string }{Test: "test"}, fldName: "Test1", extraFieldsLabel: "Test2"},
			want: "",
		},
		{
			name: "check second error",
			args: args{struct{ Test bool }{Test: false}, "Test", ""},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rcv, err := ReflectFieldAsString(tt.args.intf, tt.args.fldName, tt.args.extraFieldsLabel)

			if err == nil {
				t.Fatal("was expecting an error")
			}

			if rcv != tt.want {
				t.Errorf("reciving %v, expected %v", rcv, tt.want)
			}
		})
	}
}

func TestReflectIfaceAsDuration(t *testing.T) {
	type want struct {
		d   time.Duration
		err error
	}
	var i8 int8 = 1
	var i16 int16 = 1
	var i32 int32 = 1
	var i64 int64 = 1
	var ui uint = 1
	var ui8 uint8 = 1
	var ui16 uint16 = 1
	var ui32 uint32 = 1
	var ui64 uint64 = 1
	var f32 float32 = 1.5
	var f64 float32 = 1.5
	tests := []struct {
		name string
		arg  any
		want want
	}{
		{
			name: "time.Duration",
			arg:  1 * time.Second,
			want: want{d: 1 * time.Second, err: nil},
		},
		{
			name: "int",
			arg:  1,
			want: want{d: time.Duration(int64(1)), err: nil},
		},
		{
			name: "int8",
			arg:  i8,
			want: want{d: time.Duration(int64(i8)), err: nil},
		},
		{
			name: "int16",
			arg:  i16,
			want: want{d: time.Duration(int64(i16)), err: nil},
		},
		{
			name: "int32",
			arg:  i32,
			want: want{d: time.Duration(int64(i32)), err: nil},
		},
		{
			name: "int64",
			arg:  i64,
			want: want{d: time.Duration(int64(i64)), err: nil},
		},
		{
			name: "uint",
			arg:  ui,
			want: want{d: time.Duration(int64(ui)), err: nil},
		},
		{
			name: "uint8",
			arg:  ui8,
			want: want{d: time.Duration(int64(ui8)), err: nil},
		},
		{
			name: "uint16",
			arg:  ui16,
			want: want{d: time.Duration(int64(ui16)), err: nil},
		},
		{
			name: "uint32",
			arg:  ui32,
			want: want{d: time.Duration(int64(ui32)), err: nil},
		},
		{
			name: "uint64",
			arg:  ui64,
			want: want{d: time.Duration(int64(ui64)), err: nil},
		},
		{
			name: "float32",
			arg:  f32,
			want: want{d: time.Duration(int64(f32)), err: nil},
		},
		{
			name: "float64",
			arg:  f64,
			want: want{d: time.Duration(int64(f64)), err: nil},
		},
		{
			name: "check error(default)",
			arg:  false,
			want: want{d: 0, err: fmt.Errorf("cannot convert field: %+v to time.Duration", false)},
		},
	}

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rcv, err := IfaceAsDuration(tt.arg)

			if i == len(tests)-1 {
				if err == nil {
					t.Fatal("no error received")
				}
			} else {
				if err != tt.want.err {
					t.Fatal("wrong error message or no error received")
				}
			}

			if rcv != tt.want.d {
				t.Errorf("received %v, expected %v", rcv, tt.want.d)
			}
		})
	}
}

func TestReflectFieldAsStringOnMap(t *testing.T) {
	myMap := map[string]any{"Title": "Title1", "Count": 5, "Count64": int64(6), "Val": 7.3,
		"a": "Title2", "b": 15, "c": int64(16), "d": 17.3}
	if strVal, err := ReflectFieldAsString(myMap, "Title", ""); err != nil {
		t.Error(err)
	} else if strVal != "Title1" {
		t.Errorf("Received: %s", strVal)
	}
	if strVal, err := ReflectFieldAsString(myMap, "Count", ""); err != nil {
		t.Error(err)
	} else if strVal != "5" {
		t.Errorf("Received: %s", strVal)
	}
	if strVal, err := ReflectFieldAsString(myMap, "Count64", ""); err != nil {
		t.Error(err)
	} else if strVal != "6" {
		t.Errorf("Received: %s", strVal)
	}
	if strVal, err := ReflectFieldAsString(myMap, "Val", ""); err != nil {
		t.Error(err)
	} else if strVal != "7.3" {
		t.Errorf("Received: %s", strVal)
	}
	if strVal, err := ReflectFieldAsString(myMap, "a", ""); err != nil {
		t.Error(err)
	} else if strVal != "Title2" {
		t.Errorf("Received: %s", strVal)
	}
	if strVal, err := ReflectFieldAsString(myMap, "b", ""); err != nil {
		t.Error(err)
	} else if strVal != "15" {
		t.Errorf("Received: %s", strVal)
	}
	if strVal, err := ReflectFieldAsString(myMap, "c", ""); err != nil {
		t.Error(err)
	} else if strVal != "16" {
		t.Errorf("Received: %s", strVal)
	}
	if strVal, err := ReflectFieldAsString(myMap, "d", ""); err != nil {
		t.Error(err)
	} else if strVal != "17.3" {
		t.Errorf("Received: %s", strVal)
	}
}

func TestReflectAsMapStringIface(t *testing.T) {
	mystruct := struct {
		Title       string
		Count       int
		Count64     int64
		Val         float64
		ExtraFields map[string]any
	}{"Title1", 5, 6, 7.3, map[string]any{"a": "Title2", "b": 15, "c": int64(16), "d": 17.3}}
	expectOutMp := map[string]any{
		"Title":       "Title1",
		"Count":       5,
		"Count64":     int64(6),
		"Val":         7.3,
		"ExtraFields": map[string]any{"a": "Title2", "b": 15, "c": int64(16), "d": 17.3},
	}
	if outMp, err := AsMapStringIface(mystruct); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectOutMp, outMp) {
		t.Errorf("Expecting: %+v, received: %+v", expectOutMp, outMp)
	}
}

func TestReflectGreaterThan(t *testing.T) {
	if gte, err := GreaterThan(1, 1.2, false); err != nil {
		t.Error(err)
	} else if gte {
		t.Error("should be not greater than")
	}
	if _, err := GreaterThan(struct{}{},
		map[string]any{"a": false}, false); err == nil ||
		!strings.HasPrefix(err.Error(), "incomparable") {
		t.Error(err)
	}
	if gte, err := GreaterThan(1.2, 1.2, true); err != nil {
		t.Error(err)
	} else if !gte {
		t.Error("should be equal")
	}
	if gte, err := GreaterThan(1.3, 1.2, false); err != nil {
		t.Error(err)
	} else if !gte {
		t.Error("should be greater than")
	}
	if gte, err := GreaterThan(1.3, int(1), false); err != nil {
		t.Error(err)
	} else if !gte {
		t.Error("should be greater than")
	}
	if gte, err := GreaterThan(1.2, 1.3, false); err != nil {
		t.Error(err)
	} else if gte {
		t.Error("should be greater than")
	}
	if gte, err := GreaterThan(2, 1, false); err != nil {
		t.Error(err)
	} else if !gte {
		t.Error("should be greater than")
	}
	if gte, err := GreaterThan(2, float64(1.5), false); err != nil {
		t.Error(err)
	} else if !gte {
		t.Error("should be greater than")
	}
	if gte, err := GreaterThan(time.Duration(2*time.Second),
		time.Duration(1*time.Second), false); err != nil {
		t.Error(err)
	} else if !gte {
		t.Error("should be greater than")
	}
	if gte, err := GreaterThan(time.Duration(2*time.Second),
		20, false); err != nil {
		t.Error(err)
	} else if !gte {
		t.Error("should be greater than")
	}
	if gte, err := GreaterThan(time.Duration(2*time.Second),
		float64(1*time.Second), false); err != nil {
		t.Error(err)
	} else if !gte {
		t.Error("should be greater than")
	}
	if gte, err := GreaterThan(time.Duration(1*time.Second),
		time.Duration(2*time.Second), false); err != nil {
		t.Error(err)
	} else if gte {
		t.Error("should be less than")
	}
	if gte, err := GreaterThan(time.Duration(2*time.Second),
		time.Duration(1*time.Second), false); err != nil {
		t.Error(err)
	} else if !gte {
		t.Error("should be greater than")
	}
	now := time.Now()
	if gte, err := GreaterThan(now.Add(time.Second), now, false); err != nil {
		t.Error(err)
	} else if !gte {
		t.Error("should be greater than")
	}
	if gte, err := GreaterThan(now, now, true); err != nil {
		t.Error(err)
	} else if !gte {
		t.Error("should be equal")
	}
	if gte, err := GreaterThan(uint(1), uint(2), false); err != nil {
		t.Error(err)
	} else if gte {
		t.Error("should be greater than")
	}
	if gte, err := GreaterThan(uint(1), uint(1), true); err != nil {
		t.Error(err)
	} else if !gte {
		t.Error("should be equal")
	}
	if _, err := GreaterThan(true, true, true); err == nil {
		t.Error("was expecting an error")
	}
}

func TestReflectStringToInterface(t *testing.T) {
	if res := StringToInterface("1"); res != int64(1) {
		t.Error("not parsing int")
	}
	if res := StringToInterface("true"); res != true {
		t.Error("not parsing bool")
	}
	if res := StringToInterface("1.2"); res != 1.2 {
		t.Error("not parsing float64")
	}
	if res := StringToInterface("1.2"); res != 1.2 {
		t.Error("not parsing float64")
	}
	if res := StringToInterface("45s"); res != time.Duration(45*time.Second) {
		t.Error("not parsing time.Duration")
	}
	res := StringToInterface("+24h")
	resTime := res.(time.Time)
	now := time.Now()
	if resTime.Hour() != now.Hour() && resTime.Minute() != now.Minute() {
		t.Error("not parsing time.Time")
	}
}

func TestReflectIfaceAsString(t *testing.T) {
	val := any("string1")
	if rply := IfaceAsString(val); rply != "string1" {
		t.Errorf("Expected string1 ,received %+v", rply)
	}
	val = any(123)
	if rply := IfaceAsString(val); rply != "123" {
		t.Errorf("Expected 123 ,received %+v", rply)
	}
	val = any([]byte("byte_val"))
	if rply := IfaceAsString(val); rply != "byte_val" {
		t.Errorf("Expected byte_val ,received %+v", rply)
	}
	val = any(true)
	if rply := IfaceAsString(val); rply != "true" {
		t.Errorf("Expected true ,received %+v", rply)
	}
	if rply := IfaceAsString(time.Duration(1 * time.Second)); rply != "1s" {
		t.Errorf("Expected 1s ,received %+v", rply)
	}
	if rply := IfaceAsString(nil); rply != "" {
		t.Errorf("Expected  ,received %+v", rply)
	}
	val = any(net.ParseIP("127.0.0.1"))
	if rply := IfaceAsString(val); rply != "127.0.0.1" {
		t.Errorf("Expected  ,received %+v", rply)
	}
	val = any(10.23)
	if rply := IfaceAsString(val); rply != "10.23" {
		t.Errorf("Expected  ,received %+v", rply)
	}
	val = any(time.Date(2009, 11, 10, 23, 0, 0, 0, time.UTC))
	if rply := IfaceAsString(val); rply != "2009-11-10T23:00:00Z" {
		t.Errorf("Expected  ,received %+v", rply)
	}
	val = any(int32(123))
	if rply := IfaceAsString(val); rply != "123" {
		t.Errorf("Expected 123 ,received %+v", rply)
	}
	val = any(int64(123))
	if rply := IfaceAsString(val); rply != "123" {
		t.Errorf("Expected 123 ,received %+v", rply)
	}
	val = any(uint32(123))
	if rply := IfaceAsString(val); rply != "123" {
		t.Errorf("Expected 123 ,received %+v", rply)
	}
	val = any(uint64(123))
	if rply := IfaceAsString(val); rply != "123" {
		t.Errorf("Expected 123 ,received %+v", rply)
	}
	val = any(float32(123.5))
	if rply := IfaceAsString(val); rply != "123.5" {
		t.Errorf("Expected 123 ,received %+v", rply)
	}
	val = any(uint8(1))
	if rply := IfaceAsString(val); rply != "1" {
		t.Errorf("Expected 123 ,received %+v", rply)
	}
	val = NewNMData(123)
	if rply := IfaceAsString(val); rply != "123" {
		t.Errorf("Expected 123 ,received %+v", rply)
	}
}

func TestReflectIfaceAsTime(t *testing.T) {
	timeDate := time.Date(2009, 11, 10, 23, 0, 0, 0, time.UTC)
	val := any("2009-11-10T23:00:00Z")
	if itmConvert, err := IfaceAsTime(val, "UTC"); err != nil {
		t.Error(err)
	} else if itmConvert != timeDate {
		t.Errorf("received: %+v", itmConvert)
	}
	val = any(timeDate)
	if itmConvert, err := IfaceAsTime(val, "UTC"); err != nil {
		t.Error(err)
	} else if itmConvert != timeDate {
		t.Errorf("received: %+v", itmConvert)
	}
	val = any(1)
	if _, err := IfaceAsTime(val, "test"); err == nil {
		t.Error("There should be error")
	}
}

func TestReflectIfaceAsFloat64(t *testing.T) {
	eFloat := 6.0
	val := any(6.0)
	if itmConvert, err := IfaceAsFloat64(val); err != nil {
		t.Error(err)
	} else if itmConvert != eFloat {
		t.Errorf("received: %+v", itmConvert)
	}
	val = any(time.Duration(6))
	if itmConvert, err := IfaceAsFloat64(val); err != nil {
		t.Error(err)
	} else if itmConvert != eFloat {
		t.Errorf("received: %+v", itmConvert)
	}
	val = any("6")
	if itmConvert, err := IfaceAsFloat64(val); err != nil {
		t.Error(err)
	} else if itmConvert != eFloat {
		t.Errorf("received: %+v", itmConvert)
	}
	val = any(int64(6))
	if itmConvert, err := IfaceAsFloat64(val); err != nil {
		t.Error(err)
	} else if itmConvert != eFloat {
		t.Errorf("received: %+v", itmConvert)
	}
	val = any("This is not a float")
	if _, err := IfaceAsFloat64(val); err == nil {
		t.Error("expecting error")
	}
}

func TestReflectIfaceAsInt64(t *testing.T) {
	eInt := int64(3)
	val := any(int32(3))
	if itmConvert, err := IfaceAsInt64(val); err != nil {
		t.Error(err)
	} else if itmConvert != eInt {
		t.Errorf("received: %+v", itmConvert)
	}
	val = any(time.Duration(3))
	if itmConvert, err := IfaceAsInt64(val); err != nil {
		t.Error(err)
	} else if itmConvert != eInt {
		t.Errorf("received: %+v", itmConvert)
	}
	val = any("3")
	if itmConvert, err := IfaceAsInt64(val); err != nil {
		t.Error(err)
	} else if itmConvert != eInt {
		t.Errorf("received: %+v", itmConvert)
	}
	val = any(int64(3))
	if itmConvert, err := IfaceAsInt64(val); err != nil {
		t.Error(err)
	} else if itmConvert != eInt {
		t.Errorf("received: %+v", itmConvert)
	}
	val = any("This is not an integer")
	if _, err := IfaceAsInt64(val); err == nil {
		t.Error("expecting error")
	}
}

func TestReflectIfaceAsTInt64(t *testing.T) {
	eInt := int64(3)
	val := any(3)
	if itmConvert, err := IfaceAsTInt64(val); err != nil {
		t.Error(err)
	} else if itmConvert != eInt {
		t.Errorf("received: %+v", itmConvert)
	}
	val = any(time.Duration(3))
	if itmConvert, err := IfaceAsTInt64(val); err != nil {
		t.Error(err)
	} else if itmConvert != eInt {
		t.Errorf("received: %+v", itmConvert)
	}
	val = any("3")
	if itmConvert, err := IfaceAsTInt64(val); err != nil {
		t.Error(err)
	} else if itmConvert != eInt {
		t.Errorf("received: %+v", itmConvert)
	}
	val = any(int64(3))
	if itmConvert, err := IfaceAsTInt64(val); err != nil {
		t.Error(err)
	} else if itmConvert != eInt {
		t.Errorf("received: %+v", itmConvert)
	}
	val = any(int32(3))
	if itmConvert, err := IfaceAsTInt64(val); err != nil {
		t.Error(err)
	} else if itmConvert != eInt {
		t.Errorf("received: %+v", itmConvert)
	}
	val = any(float32(3.14))
	if itmConvert, err := IfaceAsTInt64(val); err != nil {
		t.Error(err)
	} else if itmConvert != eInt {
		t.Errorf("received: %+v", itmConvert)
	}
	val = any(float64(3.14))
	if itmConvert, err := IfaceAsTInt64(val); err != nil {
		t.Error(err)
	} else if itmConvert != eInt {
		t.Errorf("received: %+v", itmConvert)
	}
	val = any(false)
	if _, err := IfaceAsTInt64(val); err == nil {
		t.Error("expecting error")
	}
}

func TestReflectIfaceAsBool(t *testing.T) {
	val := any(true)
	if itmConvert, err := IfaceAsBool(val); err != nil {
		t.Error(err)
	} else if itmConvert != true {
		t.Errorf("received: %+v", itmConvert)
	}
	val = any("true")
	if itmConvert, err := IfaceAsBool(val); err != nil {
		t.Error(err)
	} else if itmConvert != true {
		t.Errorf("received: %+v", itmConvert)
	}
	val = any(0)
	if itmConvert, err := IfaceAsBool(val); err != nil {
		t.Error(err)
	} else if itmConvert != false {
		t.Errorf("received: %+v", itmConvert)
	}
	val = any(int64(1))
	if itmConvert, err := IfaceAsBool(val); err != nil {
		t.Error(err)
	} else if itmConvert != true {
		t.Errorf("received: %+v", itmConvert)
	}
	val = any(0.0)
	if itmConvert, err := IfaceAsBool(val); err != nil {
		t.Error(err)
	} else if itmConvert != false {
		t.Errorf("received: %+v", itmConvert)
	}
	val = any(1.0)
	if itmConvert, err := IfaceAsBool(val); err != nil {
		t.Error(err)
	} else if itmConvert != true {
		t.Errorf("received: %+v", itmConvert)
	}
	val = any(int8(1))
	if _, err := IfaceAsBool(val); err == nil {
		t.Error("expecting error")
	}
}

func TestReflectSum(t *testing.T) {
	if _, err := Sum(1); err == nil || err != ErrNotEnoughParameters {
		t.Error(err)
	}
	if _, err := Sum(1, 1.2, false); err == nil || err.Error() != "cannot convert field: 1.2 to int" {
		t.Error(err)
	}
	if sum, err := Sum(1.2, 1.2, 1.2, 1.2); err != nil {
		t.Error(err)
	} else if sum != 4.8 {
		t.Errorf("Expecting: 4.8, received: %+v", sum)
	}
	if sum, err := Sum(2, 4, 6, 8); err != nil {
		t.Error(err)
	} else if sum != int64(20) {
		t.Errorf("Expecting: 20, received: %+v", sum)
	}
	if sum, err := Sum(0.5, 1.23, 1.456, 2.234, 11.2, 0.45); err != nil {
		t.Error(err)
	} else if sum != 17.069999999999997 {
		t.Errorf("Expecting: 17.069999999999997, received: %+v", sum)
	}
	now := time.Now()
	then := time.Date(2009, 11, 17, 20, 34, 58, 651387237, time.UTC)
	if _, err := Sum(now, then); err == nil {
		t.Fatal("was expecting an error")
	}
	sum, err := Sum(then, time.Duration(1*time.Second))
	str := sum.(time.Time).String()
	if err != nil {
		t.Error(err)
	} else if str != "2009-11-17 20:34:59.651387237 +0000 UTC" {
		t.Errorf("Expecting: 2009-11-17 20:34:59.651387237 +0000 UTC , received: %v", str)
	}
	if sum, err := Sum(time.Duration(2*time.Second), time.Duration(10*time.Millisecond)); err != nil {
		t.Error(err)
	} else if sum != time.Duration(2*time.Second+10*time.Millisecond) {
		t.Errorf("Expecting: 2s10ms, received: %+v", sum)
	}
	if _, err := Sum(time.Duration(2*time.Second), false); err == nil {
		t.Fatal("was expecting an error")
	}
	if _, err := Sum(1.2, 1.2, 1.2, ""); err == nil {
		t.Error("was expecting an error")
	}
	if sum, err := Sum(int64(2), int64(4)); err != nil {
		t.Error(err)
	} else if sum != int64(6) {
		t.Errorf("Expecting: 20, received: %+v", sum)
	}
	if _, err := Sum(int64(2), "int64(4)"); err == nil {
		t.Error("was expecting an error")
	}
	if _, err := Sum(int(2), "int64(4)"); err == nil {
		t.Error("was expecting an error")
	}
}

func TestReflectGetUniformType(t *testing.T) {
	var arg, expected any
	arg = time.Second
	expected = float64(time.Second)
	if rply, err := GetUniformType(arg); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rply, expected) {
		t.Errorf("Expected: %v of type %T, received: %v of type %T", expected, expected, rply, rply)
	}
	arg = uint(10)
	expected = float64(10)
	if rply, err := GetUniformType(arg); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rply, expected) {
		t.Errorf("Expected: %v of type %T, received: %v of type %T", expected, expected, rply, rply)
	}
	arg = int64(10)
	if rply, err := GetUniformType(arg); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rply, expected) {
		t.Errorf("Expected: %v of type %T, received: %v of type %T", expected, expected, rply, rply)
	}

	arg = time.Now()
	expected = arg
	if rply, err := GetUniformType(arg); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rply, expected) {
		t.Errorf("Expected: %v of type %T, received: %v of type %T", expected, expected, rply, rply)
	}
	arg = struct{ b int }{b: 10}
	expected = arg
	if rply, err := GetUniformType(arg); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rply, expected) {
		t.Errorf("Expected: %v of type %T, received: %v of type %T", expected, expected, rply, rply)
	}

	arg = time.Now()
	if _, err := GetUniformType(&arg); err == nil || err.Error() != "incomparable" {
		t.Errorf("Exppected \"incomparable\" error received:%v ", err)
	}
	arg = uint(10)
	if _, err := GetUniformType(&arg); err == nil || err.Error() != "incomparable" {
		t.Errorf("Exppected \"incomparable\" error received:%v ", err)
	}
	arg = true
	if _, err := GetUniformType(arg); err == nil || err.Error() != "incomparable" {
		t.Errorf("Exppected \"incomparable\" error received:%v ", err)
	}
	arg = "String"
	if _, err := GetUniformType(arg); err == nil || err.Error() != "incomparable" {
		t.Errorf("Exppected \"incomparable\" error received:%v ", err)
	}
}

func TestReflectGetBasicType(t *testing.T) {
	var wantInt64 int64 = 1
	var argu uint = 1
	var wantu64 uint64 = 1
	tests := []struct {
		name string
		arg  any
		want any
	}{
		{
			name: "int argument",
			arg:  1,
			want: wantInt64,
		},
		{
			name: "uint argument",
			arg:  argu,
			want: wantu64,
		},
		{
			name: "float argument",
			arg:  1.5,
			want: 1.5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rcv := GetBasicType(tt.arg)

			if rcv != tt.want {
				t.Errorf("received %T, expected %T", rcv, tt.want)
			}
		})
	}
}

func TestReflectDifference(t *testing.T) {
	if _, err := Difference(10); err == nil || err != ErrNotEnoughParameters {
		t.Error(err)
	}
	if _, err := Difference(10, 1.2, false); err == nil || err.Error() != "cannot convert field: 1.2 to int" {
		t.Error(err)
	}
	if diff, err := Difference(12, 1, 2, 3); err != nil {
		t.Error(err)
	} else if diff != int64(6) {
		t.Errorf("Expecting: 6, received: %+v", diff)
	}
	if diff, err := Difference(8.0, 4.0, 2.0, -1.0); err != nil {
		t.Error(err)
	} else if diff != 3.0 {
		t.Errorf("Expecting: 3.0, received: %+v", diff)
	}

	if diff, err := Difference(8.0, 4, 2.0, -1.0); err != nil {
		t.Error(err)
	} else if diff != 3.0 {
		t.Errorf("Expecting: 3.0, received: %+v", diff)
	}
	if diff, err := Difference(10*time.Second, 1*time.Second, 2*time.Second,
		4*time.Millisecond); err != nil {
		t.Error(err)
	} else if diff != time.Duration(6*time.Second+996*time.Millisecond) {
		t.Errorf("Expecting: 6.996ms, received: %+v", diff)
	}
	if diff, err := Difference(time.Duration(2*time.Second),
		time.Duration(10*time.Millisecond)); err != nil {
		t.Error(err)
	} else if diff != time.Duration(1*time.Second+990*time.Millisecond) {
		t.Errorf("Expecting: 1.99s, received: %+v", diff)
	}
	if _, err := Difference(time.Duration(2*time.Second), false); err == nil {
		t.Fatal("was expecting an error")
	}
	if diff, err := Difference(time.Date(2009, 11, 10, 23, 0, 0, 0, time.UTC),
		time.Duration(10*time.Second)); err != nil {
		t.Error(err)
	} else if diff != time.Date(2009, 11, 10, 22, 59, 50, 0, time.UTC) {
		t.Errorf("Expecting: %+v, received: %+v", time.Date(2009, 11, 10, 22, 59, 50, 0, time.UTC), diff)
	}

	if diff, err := Difference(time.Date(2009, 11, 10, 23, 0, 0, 0, time.UTC),
		time.Duration(10*time.Second), 10000000000); err != nil {
		t.Error(err)
	} else if diff != time.Date(2009, 11, 10, 22, 59, 40, 0, time.UTC) {
		t.Errorf("Expecting: %+v, received: %+v", time.Date(2009, 11, 10, 22, 59, 40, 0, time.UTC), diff)
	}
	if _, err := Difference(time.Date(2009, 11, 10, 23, 0, 0, 0, time.UTC), false); err == nil {
		t.Fatal("was expecting an error")
	}
	if _, err := Difference(1.5, false); err == nil {
		t.Fatal("was expecting an error")
	}
	if diff, err := Difference(int64(2), int64(1)); err != nil {
		t.Error(err)
	} else if diff != int64(1) {
		t.Errorf("Expecting: 1, received: %+v", diff)
	}
	if _, err := Difference(int64(1), false); err == nil {
		t.Fatal("was expecting an error")
	}
	if _, err := Difference(uint8(1), false); err == nil {
		t.Fatal("was expecting an error")
	}
}

func TestReflectEqualTo(t *testing.T) {
	if gte, err := EqualTo(1, 1.2); err != nil {
		t.Error(err)
	} else if gte {
		t.Error("should be not greater than")
	}
	if _, err := EqualTo(struct{}{},
		map[string]any{"a": "a"}); err == nil ||
		!strings.HasPrefix(err.Error(), "incomparable") {
		t.Error(err)
	}
	if gte, err := EqualTo(1.2, 1.2); err != nil {
		t.Error(err)
	} else if !gte {
		t.Error("should be equal")
	}
	if gte, err := EqualTo(1.3, 1.2); err != nil {
		t.Error(err)
	} else if gte {
		t.Error("should not be equal")
	}
	if gte, err := EqualTo(1.3, int(1)); err != nil {
		t.Error(err)
	} else if gte {
		t.Error("should not be equal")
	}
	if gte, err := EqualTo(1.2, 1.3); err != nil {
		t.Error(err)
	} else if gte {
		t.Error("should be equal")
	}
	if gte, err := EqualTo(2.0, int64(2)); err != nil {
		t.Error(err)
	} else if !gte {
		t.Error("should be equal")
	}
	if gte, err := EqualTo(time.Duration(2*time.Second),
		time.Duration(2*time.Second)); err != nil {
		t.Error(err)
	} else if !gte {
		t.Error("should be equal")
	}
	now := time.Now()
	if gte, err := EqualTo(now.Add(time.Second), now); err != nil {
		t.Error(err)
	} else if gte {
		t.Error("should not be equal")
	}
	if gte, err := EqualTo(now, now); err != nil {
		t.Error(err)
	} else if !gte {
		t.Error("should be equal")
	}
	if gte, err := EqualTo("test", "test"); err != nil {
		t.Error(err)
	} else if !gte {
		t.Error("should be equal")
	}
	if gte, err := EqualTo(uint64(1), uint64(1)); err != nil {
		t.Error(err)
	} else if !gte {
		t.Error("should be equal")
	}
	if _, err := EqualTo(true, true); err == nil {
		t.Fatal("expected an error")
	}
}

type TestA struct {
	StrField string
}

type TestASlice []*TestA

func (_ *TestA) TestFunc() string {
	return "This is a test function on a structure"
}

func (_ *TestA) TestFuncWithParam(param string) string {
	return "Invalid"
}

func (_ *TestA) TestFuncWithError() (string, error) {
	return "TestFuncWithError", nil
}

func (_ *TestA) TestFuncWithPtrError() (string, *error) {
	return "TestFuncWithError", nil
}

func (_ *TestA) TestFuncWithError2() (string, error) {
	return "TestFuncWithError2", ErrPartiallyExecuted
}

func (_ *TestA) TestFuncWithThree() (string, int, error) {
	return "TestFuncWithError", 1, nil
}

func TestReflectFieldMethodInterface(t *testing.T) {
	a := &TestA{StrField: "TestStructField"}
	ifValue, err := ReflectFieldMethodInterface(a, "StrField")
	if err != nil {
		t.Error(err)
	} else if ifValue != "TestStructField" {
		t.Errorf("Expecting: TestStructField, received: %+v", ifValue)
	}
	ifValue, err = ReflectFieldMethodInterface(a, "InexistentField")
	if err != ErrNotFound {
		t.Error(err)
	}
	ifValue, err = ReflectFieldMethodInterface(a, "TestFunc")
	if err != nil {
		t.Error(err)
	} else if ifValue != "This is a test function on a structure" {
		t.Errorf("Expecting: This is a test function on a structure, received: %+v", ifValue)
	}
	ifValue, err = ReflectFieldMethodInterface(a, "TestFuncWithError")
	if err != nil {
		t.Error(err)
	} else if ifValue != "TestFuncWithError" {
		t.Errorf("Expecting: TestFuncWithError, received: %+v", ifValue)
	}
	ifValue, err = ReflectFieldMethodInterface(map[string]string{}, "TestFuncWithError2")
	if err == nil || err != ErrNotFound {
		t.Error(err)
	}
	ifValue, err = ReflectFieldMethodInterface(a, "TestFunc")
	if err != nil {
		t.Error(err)
	} else if ifValue != "This is a test function on a structure" {
		t.Errorf("Expecting: This is a test function on a structure, received: %+v", ifValue)
	}
	ifValue, err = ReflectFieldMethodInterface([]string{}, "TestFuncWithError2")
	if err == nil {
		t.Error(err)
	}
	ifValue, err = ReflectFieldMethodInterface([]string{"test"}, "1")
	if err == nil {
		t.Error(err)
	}
	ifValue, err = ReflectFieldMethodInterface([]string{"test"}, "0")
	if err != nil {
		t.Error(err)
	} else if ifValue != "test" {
		t.Errorf("Expecting: test, received: %+v", ifValue)
	}
	ifValue, err = ReflectFieldMethodInterface(1, "1")
	if err == nil {
		t.Error(err)
	}
	ifValue, err = ReflectFieldMethodInterface(a, "TestFuncWithParam")
	if err == nil {
		t.Fatal("was expecting an error")
	}
	ifValue, err = ReflectFieldMethodInterface(a, "TestFuncWithThree")
	if err == nil {
		t.Fatal("was expecting an error")
	}
	ifValue, err = ReflectFieldMethodInterface(a, "TestFuncWithPtrError")
	if err == nil {
		t.Fatal("was expecting an error")
	}
	ifValue, err = ReflectFieldMethodInterface(a, "TestFuncWithError2")
	if err == nil {
		t.Fatal("was expecting a error")
	}
	if ifValue != "TestFuncWithError2" {
		t.Errorf("Expecting: TestFuncWithError2, received: %+v", ifValue)
	}
}

func TestReflectIfaceAsSliceString(t *testing.T) {
	var attrs any
	var expected []string
	if rply, err := IfaceAsSliceString(attrs); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting: %s ,received: %s", expected, rply)
	}

	attrs = []int{1, 2, 3}
	expected = []string{"1", "2", "3"}
	if rply, err := IfaceAsSliceString(attrs); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting: %s ,received: %s", expected, rply)
	}
	attrs = []int32{1, 2, 3}
	if rply, err := IfaceAsSliceString(attrs); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting: %s ,received: %s", expected, rply)
	}
	attrs = []int64{1, 2, 3}
	if rply, err := IfaceAsSliceString(attrs); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting: %s ,received: %s", expected, rply)
	}
	attrs = []uint{1, 2, 3}
	if rply, err := IfaceAsSliceString(attrs); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting: %s ,received: %s", expected, rply)
	}
	attrs = []uint{1, 2, 3}
	if rply, err := IfaceAsSliceString(attrs); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting: %s ,received: %s", expected, rply)
	}
	attrs = []uint32{1, 2, 3}
	if rply, err := IfaceAsSliceString(attrs); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting: %s ,received: %s", expected, rply)
	}
	attrs = []uint64{1, 2, 3}
	if rply, err := IfaceAsSliceString(attrs); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting: %s ,received: %s", expected, rply)
	}
	attrs = []float32{1, 2, 3}
	if rply, err := IfaceAsSliceString(attrs); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting: %s ,received: %s", expected, rply)
	}
	attrs = []float64{1, 2, 3}
	if rply, err := IfaceAsSliceString(attrs); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting: %s ,received: %s", expected, rply)
	}
	attrs = []string{"1", "2", "3"}
	if rply, err := IfaceAsSliceString(attrs); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting: %s ,received: %s", expected, rply)
	}
	attrs = [][]byte{[]byte("1"), []byte("2"), []byte("3")}
	if rply, err := IfaceAsSliceString(attrs); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting: %s ,received: %s", expected, rply)
	}
	attrs = []bool{true, false}
	expected = []string{"true", "false"}
	if rply, err := IfaceAsSliceString(attrs); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting: %s ,received: %s", expected, rply)
	}

	attrs = []time.Duration{time.Second, time.Minute}
	expected = []string{"1s", "1m0s"}
	if rply, err := IfaceAsSliceString(attrs); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting: %s ,received: %s", expected, rply)
	}
	tNow := time.Now()
	attrs = []time.Time{tNow}
	expected = []string{tNow.Format(time.RFC3339)}
	if rply, err := IfaceAsSliceString(attrs); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting: %s ,received: %s", expected, rply)
	}
	attrs = []net.IP{net.ParseIP("127.0.0.1")}
	expected = []string{"127.0.0.1"}
	if rply, err := IfaceAsSliceString(attrs); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting: %s ,received: %s", expected, rply)
	}
	attrs = []any{true, 10, "two"}
	expected = []string{"true", "10", "two"}
	if rply, err := IfaceAsSliceString(attrs); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting: %s ,received: %s", expected, rply)
	}
	attrs = "notSlice"
	expError := fmt.Errorf("cannot convert field: %T to []string", attrs)
	if _, err := IfaceAsSliceString(attrs); err == nil || err.Error() != expError.Error() {
		t.Errorf("Expected error %s ,received: %v", expError, err)
	}
}

func TestReflecttAsMapStringIface(t *testing.T) {
	str := "arg"
	type want struct {
		out map[string]any
		err error
	}
	tests := []struct {
		name string
		arg  any
		want want
	}{
		{
			name: "pointer non struct argument",
			arg:  &str,
			want: want{out: nil, err: fmt.Errorf("AsMapStringIface only accepts structs; got %T", str)},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rcv, err := AsMapStringIface(tt.arg)

			if err == nil {
				t.Fatalf("wrong error message, expected %s, received %s", tt.want.err, err)
			}

			if rcv != nil {
				t.Errorf("expected %v, received %v", tt.want.out, rcv)
			}
		})
	}
}
