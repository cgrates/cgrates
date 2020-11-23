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
		ExtraFields map[string]interface{}
	}{"Title1", 5, 6, 7.3, map[string]interface{}{"a": "Title2", "b": 15, "c": int64(16), "d": 17.3}}
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

func TestReflectFieldAsStringOnMap(t *testing.T) {
	myMap := map[string]interface{}{"Title": "Title1", "Count": 5, "Count64": int64(6), "Val": 7.3,
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

func TestGreaterThan(t *testing.T) {
	if gte, err := GreaterThan(1, 1.2, false); err != nil {
		t.Error(err)
	} else if gte {
		t.Error("should be not greater than")
	}
	if _, err := GreaterThan(struct{}{},
		map[string]interface{}{"a": "a"}, false); err == nil ||
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
	if gte, err := GreaterThan(2*time.Second,
		time.Second, false); err != nil {
		t.Error(err)
	} else if !gte {
		t.Error("should be greater than")
	}
	if gte, err := GreaterThan(2*time.Second,
		20, false); err != nil {
		t.Error(err)
	} else if !gte {
		t.Error("should be greater than")
	}
	if gte, err := GreaterThan(2*time.Second,
		float64(time.Second), false); err != nil {
		t.Error(err)
	} else if !gte {
		t.Error("should be greater than")
	}
	if gte, err := GreaterThan(time.Second,
		2*time.Second, false); err != nil {
		t.Error(err)
	} else if gte {
		t.Error("should be less than")
	}
	if gte, err := GreaterThan(2*time.Second,
		time.Second, false); err != nil {
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
}

func TestStringToInterface(t *testing.T) {
	if res := StringToInterface("1"); res != int64(1) {
		t.Error("not parsing int")
	}
	if res := StringToInterface(""); res != "" {
		t.Error("not parsing string")
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
	if res := StringToInterface("45s"); res != 45*time.Second {
		t.Error("not parsing time.Duration")
	}
	res := StringToInterface("+24h")
	resTime := res.(time.Time)
	now := time.Now()
	if resTime.Hour() != now.Hour() && resTime.Minute() != now.Minute() {
		t.Error("not parsing time.Time")
	}
}

func TestIfaceAsString(t *testing.T) {
	val := interface{}("string1")
	if rply := IfaceAsString(val); rply != "string1" {
		t.Errorf("Expected string1 ,received %+v", rply)
	}
	val = interface{}(123)
	if rply := IfaceAsString(val); rply != "123" {
		t.Errorf("Expected 123 ,received %+v", rply)
	}
	val = interface{}([]byte("byte_val"))
	if rply := IfaceAsString(val); rply != "byte_val" {
		t.Errorf("Expected byte_val ,received %+v", rply)
	}
	val = interface{}(true)
	if rply := IfaceAsString(val); rply != "true" {
		t.Errorf("Expected true ,received %+v", rply)
	}
	if rply := IfaceAsString(time.Second); rply != "1s" {
		t.Errorf("Expected 1s ,received %+v", rply)
	}
	if rply := IfaceAsString(nil); rply != "" {
		t.Errorf("Expected  ,received %+v", rply)
	}
	val = interface{}(net.ParseIP("127.0.0.1"))
	if rply := IfaceAsString(val); rply != "127.0.0.1" {
		t.Errorf("Expected  ,received %+v", rply)
	}
	val = interface{}(10.23)
	if rply := IfaceAsString(val); rply != "10.23" {
		t.Errorf("Expected  ,received %+v", rply)
	}
	val = interface{}(time.Date(2009, 11, 10, 23, 0, 0, 0, time.UTC))
	if rply := IfaceAsString(val); rply != "2009-11-10T23:00:00Z" {
		t.Errorf("Expected  ,received %+v", rply)
	}
}

func TestIfaceAsTime(t *testing.T) {
	timeDate := time.Date(2009, 11, 10, 23, 0, 0, 0, time.UTC)
	val := interface{}("2009-11-10T23:00:00Z")
	if itmConvert, err := IfaceAsTime(val, "UTC"); err != nil {
		t.Error(err)
	} else if itmConvert != timeDate {
		t.Errorf("received: %+v", itmConvert)
	}
	val = interface{}(timeDate)
	if itmConvert, err := IfaceAsTime(val, "UTC"); err != nil {
		t.Error(err)
	} else if itmConvert != timeDate {
		t.Errorf("received: %+v", itmConvert)
	}
	val = interface{}("This is not a time")
	if _, err := IfaceAsTime(val, "UTC"); err == nil {
		t.Error("There should be error")
	}
}

func TestIfaceAsDuration(t *testing.T) {
	eItm := time.Second
	if itmConvert, err := IfaceAsDuration(interface{}(time.Second)); err != nil {
		t.Error(err)
	} else if eItm != itmConvert {
		t.Errorf("received: %+v", itmConvert)
	}
	if itmConvert, err := IfaceAsDuration(interface{}(float64(1000000000.0))); err != nil {
		t.Error(err)
	} else if eItm != itmConvert {
		t.Errorf("received: %+v", itmConvert)
	}
	if itmConvert, err := IfaceAsDuration(interface{}(int64(1000000000))); err != nil {
		t.Error(err)
	} else if eItm != itmConvert {
		t.Errorf("received: %+v", itmConvert)
	}
	if itmConvert, err := IfaceAsDuration(interface{}(int(1000000000))); err != nil {
		t.Error(err)
	} else if eItm != itmConvert {
		t.Errorf("received: %+v", itmConvert)
	}
	if itmConvert, err := IfaceAsDuration(interface{}(string("1s"))); err != nil {
		t.Error(err)
	} else if eItm != itmConvert {
		t.Errorf("received: %+v", itmConvert)
	}
	if _, err := IfaceAsDuration(interface{}(string("s1s"))); err == nil {
		t.Error("empty error")
	}
}

func TestIfaceAsFloat64(t *testing.T) {
	eFloat := 6.0
	val := interface{}(6.0)
	if itmConvert, err := IfaceAsFloat64(val); err != nil {
		t.Error(err)
	} else if itmConvert != eFloat {
		t.Errorf("received: %+v", itmConvert)
	}
	val = interface{}(6)
	if itmConvert, err := IfaceAsFloat64(val); err != nil {
		t.Error(err)
	} else if itmConvert != eFloat {
		t.Errorf("received: %+v", itmConvert)
	}
	val = interface{}("6")
	if itmConvert, err := IfaceAsFloat64(val); err != nil {
		t.Error(err)
	} else if itmConvert != eFloat {
		t.Errorf("received: %+v", itmConvert)
	}
	val = interface{}(int64(6))
	if itmConvert, err := IfaceAsFloat64(val); err != nil {
		t.Error(err)
	} else if itmConvert != eFloat {
		t.Errorf("received: %+v", itmConvert)
	}
	val = interface{}("This is not a float")
	if _, err := IfaceAsFloat64(val); err == nil {
		t.Error("expecting error")
	}
}

func TestIfaceAsInt64(t *testing.T) {
	eInt := int64(3)
	val := interface{}(3)
	if itmConvert, err := IfaceAsInt64(val); err != nil {
		t.Error(err)
	} else if itmConvert != eInt {
		t.Errorf("received: %+v", itmConvert)
	}
	val = interface{}(3)
	if itmConvert, err := IfaceAsInt64(val); err != nil {
		t.Error(err)
	} else if itmConvert != eInt {
		t.Errorf("received: %+v", itmConvert)
	}
	val = interface{}("3")
	if itmConvert, err := IfaceAsInt64(val); err != nil {
		t.Error(err)
	} else if itmConvert != eInt {
		t.Errorf("received: %+v", itmConvert)
	}
	val = interface{}(int64(3))
	if itmConvert, err := IfaceAsInt64(val); err != nil {
		t.Error(err)
	} else if itmConvert != eInt {
		t.Errorf("received: %+v", itmConvert)
	}
	val = interface{}("This is not an integer")
	if _, err := IfaceAsInt64(val); err == nil {
		t.Error("expecting error")
	}
}

func TestIfaceAsTInt64(t *testing.T) {
	eInt := int64(3)
	val := interface{}(3)
	if itmConvert, err := IfaceAsTInt64(val); err != nil {
		t.Error(err)
	} else if itmConvert != eInt {
		t.Errorf("received: %+v", itmConvert)
	}
	val = interface{}(3)
	if itmConvert, err := IfaceAsTInt64(val); err != nil {
		t.Error(err)
	} else if itmConvert != eInt {
		t.Errorf("received: %+v", itmConvert)
	}
	val = interface{}("3")
	if itmConvert, err := IfaceAsTInt64(val); err != nil {
		t.Error(err)
	} else if itmConvert != eInt {
		t.Errorf("received: %+v", itmConvert)
	}
	val = interface{}(int64(3))
	if itmConvert, err := IfaceAsTInt64(val); err != nil {
		t.Error(err)
	} else if itmConvert != eInt {
		t.Errorf("received: %+v", itmConvert)
	}
	val = interface{}(int32(3))
	if itmConvert, err := IfaceAsTInt64(val); err != nil {
		t.Error(err)
	} else if itmConvert != eInt {
		t.Errorf("received: %+v", itmConvert)
	}
	val = interface{}(float32(3.14))
	if itmConvert, err := IfaceAsTInt64(val); err != nil {
		t.Error(err)
	} else if itmConvert != eInt {
		t.Errorf("received: %+v", itmConvert)
	}
	val = interface{}(float64(3.14))
	if itmConvert, err := IfaceAsTInt64(val); err != nil {
		t.Error(err)
	} else if itmConvert != eInt {
		t.Errorf("received: %+v", itmConvert)
	}
	val = interface{}("This is not an integer")
	if _, err := IfaceAsTInt64(val); err == nil {
		t.Error("expecting error")
	}
}

func TestIfaceAsBool(t *testing.T) {
	val := interface{}(true)
	if itmConvert, err := IfaceAsBool(val); err != nil {
		t.Error(err)
	} else if itmConvert != true {
		t.Errorf("received: %+v", itmConvert)
	}
	val = interface{}("true")
	if itmConvert, err := IfaceAsBool(val); err != nil {
		t.Error(err)
	} else if itmConvert != true {
		t.Errorf("received: %+v", itmConvert)
	}
	val = interface{}(0)
	if itmConvert, err := IfaceAsBool(val); err != nil {
		t.Error(err)
	} else if itmConvert != false {
		t.Errorf("received: %+v", itmConvert)
	}
	val = interface{}(1)
	if itmConvert, err := IfaceAsBool(val); err != nil {
		t.Error(err)
	} else if itmConvert != true {
		t.Errorf("received: %+v", itmConvert)
	}
	val = interface{}(0.0)
	if itmConvert, err := IfaceAsBool(val); err != nil {
		t.Error(err)
	} else if itmConvert != false {
		t.Errorf("received: %+v", itmConvert)
	}
	val = interface{}(1.0)
	if itmConvert, err := IfaceAsBool(val); err != nil {
		t.Error(err)
	} else if itmConvert != true {
		t.Errorf("received: %+v", itmConvert)
	}
	val = interface{}("This is not a bool")
	if _, err := IfaceAsBool(val); err == nil {
		t.Error("expecting error")
	}
}

func TestSum(t *testing.T) {
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
	if sum, err := Sum(2*time.Second, time.Second, 2*time.Second,
		5*time.Second, 4*time.Millisecond); err != nil {
		t.Error(err)
	} else if sum != 10*time.Second+4*time.Millisecond {
		t.Errorf("Expecting: 10.004s, received: %+v", sum)
	}
	if sum, err := Sum(2*time.Second,
		10*time.Millisecond); err != nil {
		t.Error(err)
	} else if sum != 2*time.Second+10*time.Millisecond {
		t.Errorf("Expecting: 2s10ms, received: %+v", sum)
	}
}

func TestGetUniformType(t *testing.T) {
	var arg, expected interface{}
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

func TestDifference(t *testing.T) {
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
	if diff, err := Difference(10*time.Second, time.Second, 2*time.Second,
		4*time.Millisecond); err != nil {
		t.Error(err)
	} else if diff != 6*time.Second+996*time.Millisecond {
		t.Errorf("Expecting: 6.996ms, received: %+v", diff)
	}
	if diff, err := Difference(2*time.Second,
		10*time.Millisecond); err != nil {
		t.Error(err)
	} else if diff != time.Second+990*time.Millisecond {
		t.Errorf("Expecting: 1.99s, received: %+v", diff)
	}

	if diff, err := Difference(time.Date(2009, 11, 10, 23, 0, 0, 0, time.UTC),
		10*time.Second); err != nil {
		t.Error(err)
	} else if diff != time.Date(2009, 11, 10, 22, 59, 50, 0, time.UTC) {
		t.Errorf("Expecting: %+v, received: %+v", time.Date(2009, 11, 10, 22, 59, 50, 0, time.UTC), diff)
	}

	if diff, err := Difference(time.Date(2009, 11, 10, 23, 0, 0, 0, time.UTC),
		10*time.Second, 10000000000); err != nil {
		t.Error(err)
	} else if diff != time.Date(2009, 11, 10, 22, 59, 40, 0, time.UTC) {
		t.Errorf("Expecting: %+v, received: %+v", time.Date(2009, 11, 10, 22, 59, 40, 0, time.UTC), diff)
	}

}

func TestMultiply(t *testing.T) {
	if _, err := Multiply(10); err == nil || err != ErrNotEnoughParameters {
		t.Error(err)
	}
	if _, err := Multiply(10, 1.2, false); err == nil || err.Error() != "cannot convert field: 1.2 to int" {
		t.Error(err)
	}
	if diff, err := Multiply(12, 1, 2, 3); err != nil {
		t.Error(err)
	} else if diff != int64(72) {
		t.Errorf("Expecting: 72, received: %+v", diff)
	}
	if diff, err := Multiply(8.0, 4.0, 2.0, 1.0); err != nil {
		t.Error(err)
	} else if diff != 64.0 {
		t.Errorf("Expecting: 64.0, received: %+v", diff)
	}

	if diff, err := Multiply(8.0, 4, 6.0, 1.0); err != nil {
		t.Error(err)
	} else if diff != 192.0 {
		t.Errorf("Expecting: 192.0, received: %+v", diff)
	}
}

func TestDivide(t *testing.T) {
	if _, err := Divide(10); err == nil || err != ErrNotEnoughParameters {
		t.Error(err)
	}
	if _, err := Divide(10, 1.2, false); err == nil || err.Error() != "cannot convert field: 1.2 to int" {
		t.Error(err)
	}
	if diff, err := Divide(12, 1, 2, 3); err != nil {
		t.Error(err)
	} else if diff != int64(2) {
		t.Errorf("Expecting: 2, received: %+v", diff)
	}
	if diff, err := Divide(8.0, 4.0, 2.0, 1.0); err != nil {
		t.Error(err)
	} else if diff != 1.0 {
		t.Errorf("Expecting: 1.0, received: %+v", diff)
	}

	if diff, err := Divide(8.0, 4, 6.0, 1.0); err != nil {
		t.Error(err)
	} else if diff != 0.3333333333333333 {
		t.Errorf("Expecting: 0.3333333333333333, received: %+v", diff)
	}
}

func TestEqualTo(t *testing.T) {
	if gte, err := EqualTo(1, 1.2); err != nil {
		t.Error(err)
	} else if gte {
		t.Error("should be not greater than")
	}
	if _, err := EqualTo(struct{}{},
		map[string]interface{}{"a": "a"}); err == nil ||
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
	if gte, err := EqualTo(2*time.Second,
		2*time.Second); err != nil {
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
func (_ *TestA) TestFuncWithError2() (string, error) {
	return "TestFuncWithError2", ErrPartiallyExecuted
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
	ifValue, err = ReflectFieldMethodInterface(a, "TestFuncWithError2")
	if err == nil || err != ErrPartiallyExecuted {
		t.Error(err)
	}
}

func TestIfaceAsSliceString(t *testing.T) {
	var attrs interface{}
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
	attrs = []interface{}{true, 10, "two"}
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

func TestReflectFieldInterfaceBadType(t *testing.T) {
	_, err := ReflectFieldInterface(22, "22", "22")
	if err == nil || err.Error() != "Unsupported field kind: int" {
		t.Errorf("Expected <Unsupported field kind: int> ,received: <%+v>", err)
	}
}

func TestReflectFieldInterfaceFieldNotValid(t *testing.T) {
	var intf = map[string]int{
		"field1": 22,
		"field2": 35,
	}
	_, err := ReflectFieldInterface(intf, "", "")
	if err == nil || err.Error() != "NOT_FOUND" {
		t.Errorf("NOT_FOUND> ,received: <%+v>", err)
	}
}

func TestReflectFieldInterfaceStructCase1(t *testing.T) {
	type vStruct struct {
		X int
		Y int
	}
	intf := vStruct{22, 35}
	_, err := ReflectFieldInterface(intf, "x", "")
	if err == nil || err.Error() != "NOT_FOUND" {
		t.Errorf("<NOT_FOUND> ,received: <%+v>", err)
	}
}

func TestReflectFieldInterfaceStructCase2(t *testing.T) {
	type v struct {
		X int
		Y int
	}
	intf := v{22, 35}
	_, err := ReflectFieldInterface(intf, "x", "y")
	if err == nil || err.Error() != "NOT_FOUND" {
		t.Errorf("<NOT_FOUND> ,received: <%+v>", err)
	}
}

func TestReflectFieldInterfaceStruct(t *testing.T) {
	type Test struct {
		StrField  string
		StrField2 string
	}
	structTest := Test{
		StrField:  "TestStructField",
		StrField2: "TestStructField2",
	}
	result, _ := ReflectFieldInterface(structTest, "StrField", "StrField2")
	if !reflect.DeepEqual(result, "TestStructField") {
		t.Errorf("Expected <TestStructField> ,received: <%+v>", result)
	}
}

func TestReflectFieldInterfaceStructError3(t *testing.T) {
	type Test struct {
		StrField  string
		StrField2 string
		StrField3 map[string]string
	}
	structTest := Test{
		StrField:  "TestStructField",
		StrField2: "TestStructField2",
	}
	_, err := ReflectFieldInterface(structTest, "StrField4", "StrField3")
	if err != ErrNotFound {
		t.Errorf("<NOT_FOUND> ,received: <%+v>", err)
	}
}

func TestReflectFieldAsStringError(t *testing.T) {
	var test bool
	_, err := IfaceAsTime(test, "")
	if err == nil || err.Error() != "cannot convert field: false to time.Time" {
		t.Errorf("Expecting <cannot convert field: false to time.Time> ,received: <%+v>", err)
	}
}

func TestIfaceAsDurationDefaultError(t *testing.T) {
	var test bool
	_, err := IfaceAsDuration(test)
	if err == nil || err.Error() != "cannot convert field: false to time.Duration" {
		t.Errorf("Expecting <cannot convert field: false to time.Duration> ,received: <%+v>", err)
	}
}

func TestIfaceAsDurationCaseUInt(t *testing.T) {
	var test uint
	test = 127
	response, _ := IfaceAsDuration(test)
	if !reflect.DeepEqual(response.String(), "127ns") {
		t.Errorf("Expected <127ns> ,received: <%+v>", response)
	}
}

func TestIfaceAsDurationCaseInt8(t *testing.T) {
	var test int8
	test = 127
	response, _ := IfaceAsDuration(test)
	if !reflect.DeepEqual(response.String(), "127ns") {
		t.Errorf("Expected <127ns> ,received: <%+v>", response)
	}
}

func TestIfaceAsDurationCaseNegInt8(t *testing.T) {
	var test int8
	test = -127
	response, _ := IfaceAsDuration(test)
	if !reflect.DeepEqual(response.String(), "-127ns") {
		t.Errorf("Expected <-127ns> ,received: <%+v>", response)
	}
}

func TestIfaceAsDurationCaseUInt8(t *testing.T) {
	var test uint8
	test = 127
	response, _ := IfaceAsDuration(test)
	if !reflect.DeepEqual(response.String(), "127ns") {
		t.Errorf("Expected <127ns> ,received: <%+v>", response)
	}
}

func TestIfaceAsDurationCaseInt16(t *testing.T) {
	var test int16
	test = 32767
	response, _ := IfaceAsDuration(test)
	if !reflect.DeepEqual(response.String(), "32.767µs") {
		t.Errorf("Expected <32.767µs> ,received: <%+v>", response)
	}
}

func TestIfaceAsDurationCaseNegInt16(t *testing.T) {
	var test int16
	test = -32767
	response, _ := IfaceAsDuration(test)
	if !reflect.DeepEqual(response.String(), "-32.767µs") {
		t.Errorf("Expected <-32.767µs> ,received: <%+v>", response)
	}
}

func TestIfaceAsDurationCaseUInt16(t *testing.T) {
	var test uint16
	test = 32767
	response, _ := IfaceAsDuration(test)
	if !reflect.DeepEqual(response.String(), "32.767µs") {
		t.Errorf("Expected <32.767µs> ,received: <%+v>", response)
	}
}

func TestIfaceAsDurationCaseInt32(t *testing.T) {
	var test int32
	test = 2147483647
	response, _ := IfaceAsDuration(test)
	if !reflect.DeepEqual(response.String(), "2.147483647s") {
		t.Errorf("Expected <2.147483647s> ,received: <%+v>", response)
	}
}

func TestIfaceAsDurationCaseNegInt32(t *testing.T) {
	var test int32
	test = -2147483647
	response, _ := IfaceAsDuration(test)
	if !reflect.DeepEqual(response.String(), "-2.147483647s") {
		t.Errorf("Expected <-2.147483647s> ,received: <%+v>", response)
	}
}

func TestIfaceAsDurationCaseUInt32(t *testing.T) {
	var test uint32
	test = 2147483647
	response, _ := IfaceAsDuration(test)
	if !reflect.DeepEqual(response.String(), "2.147483647s") {
		t.Errorf("Expected <2.147483647s> ,received: <%+v>", response)
	}
}

func TestIfaceAsDurationCaseInt64(t *testing.T) {
	var test int64
	test = 9223372036854775807
	response, _ := IfaceAsDuration(test)
	if !reflect.DeepEqual(response.String(), "2562047h47m16.854775807s") {
		t.Errorf("Expected <2562047h47m16.854775807s> ,received: <%+v>", response)
	}
}

func TestIfaceAsDurationCaseNegInt64(t *testing.T) {
	var test int64
	test = -9223372036854775807
	response, _ := IfaceAsDuration(test)
	if !reflect.DeepEqual(response.String(), "-2562047h47m16.854775807s") {
		t.Errorf("Expected <-2562047h47m16.854775807s> ,received: <%+v>", response)
	}
}

func TestIfaceAsDurationCaseUInt64(t *testing.T) {
	var test uint64
	test = 9223372036854775807
	response, _ := IfaceAsDuration(test)
	if !reflect.DeepEqual(response.String(), "2562047h47m16.854775807s") {
		t.Errorf("Expected <2562047h47m16.854775807s> ,received: <%+v>", response)
	}
}

func TestIfaceAsDurationCaseFloat32(t *testing.T) {
	var test float32
	test = 9.5555555
	response, _ := IfaceAsDuration(test)
	if !reflect.DeepEqual(response.String(), "9ns") {
		t.Errorf("Expected <9ns> ,received: <%+v>", response)
	}
}

func TestIfaceAsInt6432to64(t *testing.T) {
	var test int32
	var expected int64
	test = 2147483647
	expected = 2147483647
	response, _ := IfaceAsInt64(test)
	if !reflect.DeepEqual(response, expected) {
		t.Errorf("Expected <%+v> ,received: <%+v>", expected, response)
	}
}

func TestIfaceAsInt64Default(t *testing.T) {
	var test bool
	test = true
	_, err := IfaceAsInt64(test)
	if err == nil || err.Error() != "cannot convert field: true to int" {
		t.Errorf("Expecting <cannot convert field: true to int> ,received: <%+v>", err)
	}
}

func TestIfaceAsInt64Nanosecs(t *testing.T) {
	var test time.Duration
	test = 2147483647
	response, _ := IfaceAsInt64(test)
	if !reflect.DeepEqual(response, test.Nanoseconds()) {
		t.Errorf("Expected <%+v> ,received: <%+v>", test.Nanoseconds(), response)
	}
}

func TestIfaceAsTInt64Default(t *testing.T) {
	var test bool
	test = true
	_, err := IfaceAsTInt64(test)
	if err == nil || err.Error() != "cannot convert field<bool>: true to int" {
		t.Errorf("Expecting <cannot convert field<bool>: true to int> ,received: <%+v>", err)
	}
}

func TestIfaceAsTInt64Nanosecs(t *testing.T) {
	var test time.Duration
	test = 2147483647
	response, _ := IfaceAsTInt64(test)
	if !reflect.DeepEqual(response, test.Nanoseconds()) {
		t.Errorf("Expected <%+v> ,received: <%+v>", test.Nanoseconds(), response)
	}
}

func TestIfaceAsBoolInt64(t *testing.T) {
	var test int64
	test = 2147483647
	response, _ := IfaceAsBool(test)
	if !reflect.DeepEqual(response, true) {
		t.Errorf("Expected <%+v> ,received: <%+v>", true, response)
	}
}

func TestIfaceAsBoolDefault(t *testing.T) {
	var test uint64
	test = 2147483647
	_, err := IfaceAsBool(test)
	if err == nil || err.Error() != "cannot convert field: 2147483647 to bool" {
		t.Errorf("Expecting <cannot convert field: 2147483647 to bool> ,received: <%+v>", err)
	}
}

func TestIfaceAsStringInt32(t *testing.T) {
	var test int32
	var expected string
	test = 2147483647
	expected = "2147483647"
	response := IfaceAsString(test)
	if !reflect.DeepEqual(response, expected) {
		t.Errorf("Expected <%+v> ,received: <%+v>", expected, response)
	}
}

func TestIfaceAsStringInt32Neg(t *testing.T) {
	var test int32
	var expected string
	test = -2147483647
	expected = "-2147483647"
	response := IfaceAsString(test)
	if !reflect.DeepEqual(response, expected) {
		t.Errorf("Expected <%+v> ,received: <%+v>", expected, response)
	}
}

func TestIfaceAsStringInt64Neg(t *testing.T) {
	var test int64
	var expected string
	test = -9223372036854775807
	expected = "-9223372036854775807"
	response := IfaceAsString(test)
	if !reflect.DeepEqual(response, expected) {
		t.Errorf("Expected <%+v> ,received: <%+v>", expected, response)
	}
}

func TestIfaceAsStringUInt32(t *testing.T) {
	var test uint32
	var expected string
	test = 2147483647
	expected = "2147483647"
	response := IfaceAsString(test)
	if !reflect.DeepEqual(response, expected) {
		t.Errorf("Expected <%+v> ,received: <%+v>", expected, response)
	}
}

func TestIfaceAsStringUInt64(t *testing.T) {
	var test uint64
	var expected string
	test = 9223372036854775807
	expected = "9223372036854775807"
	response := IfaceAsString(test)
	if !reflect.DeepEqual(response, expected) {
		t.Errorf("Expected <%+v> ,received: <%+v>", expected, response)
	}
}

func TestIfaceAsStringFloat32(t *testing.T) {
	var test float32
	var expected string
	test = 2.5
	expected = "2.5"
	response := IfaceAsString(test)
	if !reflect.DeepEqual(response, expected) {
		t.Errorf("Expected <%+v> ,received: <%+v>", expected, response)
	}
}

func TestIfaceAsStringFloat32Neg(t *testing.T) {
	var test float32
	var expected string
	test = -2.5
	expected = "-2.5"
	response := IfaceAsString(test)
	if !reflect.DeepEqual(response, expected) {
		t.Errorf("Expected <%+v> ,received: <%+v>", expected, response)
	}
}

func TestIfaceAsStringNMInterface(t *testing.T) {
	var test NMInterface = NewNMData("1001")
	response := IfaceAsString(test)
	if !reflect.DeepEqual(response, "1001") {
		t.Errorf("Expected <1001> ,received: <%+v>", response)
	}
}

func TestGetBasicTypeUint(t *testing.T) {
	var test interface{}
	test = uint8(123)
	valItm := reflect.ValueOf(test)
	response := GetBasicType(test)
	if !reflect.DeepEqual(valItm.Uint(), response) {
		t.Errorf("Expected <%+v> ,received: <%+v>", valItm.Uint(), response)
	}
}

func TestReflectFieldInterfacePointer(t *testing.T) {
	type Test struct {
		StrField  string
		StrField2 string
	}
	structTest := Test{
		StrField:  "TestStructField",
		StrField2: "TestStructField2",
	}
	result, _ := ReflectFieldInterface(&structTest, "StrField", "StrField2")
	if !reflect.DeepEqual(result, "TestStructField") {
		t.Errorf("Expected <TestStructField> ,received: <%+v>", result)
	}
}

func TestReflectFieldAsStringErrorCase(t *testing.T) {
	type Test struct {
		StrField  string
		StrField2 string
		StrField3 map[string]string
	}
	structTest := Test{
		StrField:  "TestStructField",
		StrField2: "TestStructField2",
	}
	_, err := ReflectFieldAsString(structTest, "StrField4", "StrField3")
	if err != ErrNotFound {
		t.Errorf("<NOT_FOUND> ,received: <%+v>", err)
	}
}

func TestReflectFieldAsStringDefaultError(t *testing.T) {
	type Test struct {
		StrField  bool
		StrField2 bool
	}
	structTest := Test{
		StrField:  true,
		StrField2: true,
	}
	_, err := ReflectFieldAsString(structTest, "StrField", "StrField2")
	if err == nil || err.Error() != "Cannot convert to string field type: bool" {
		t.Errorf("Expected <Cannot convert to string field type: bool> ,received: <%+v>", err)
	}
}
func TestGreaterThanUint64(t *testing.T) {
	var firstUint64 uint64
	var secondUint64 uint64
	firstUint64 = 1
	secondUint64 = 2
	if gte, err := GreaterThan(firstUint64, secondUint64, false); err != nil {
		t.Error(err)
	} else if gte {
		t.Error("should be not greater than")
	}
}

func TestGreaterThanUint64Equal(t *testing.T) {
	var firstUint64 uint64
	var secondUint64 uint64
	firstUint64 = 2
	secondUint64 = 2
	if gte, err := GreaterThan(firstUint64, secondUint64, true); err != nil {
		t.Error(err)
	} else if !gte {
		t.Error("should be equal")
	}
}

func TestGreaterThanDefaultError(t *testing.T) {
	var firstUint64 bool
	var secondUint64 bool
	firstUint64 = true
	secondUint64 = false
	_, err := GreaterThan(firstUint64, secondUint64, true)
	if err == nil || err.Error() != "unsupported comparison type: bool, kind: bool" {
		t.Errorf("Expected <unsupported comparison type: bool, kind: bool> ,received: <%+v>", err)
	}
}

func TestEqualToUint64(t *testing.T) {
	var firstUint64 uint64
	var secondUint64 uint64
	firstUint64 = 2
	secondUint64 = 2
	if gte, err := EqualTo(firstUint64, secondUint64); err != nil {
		t.Error(err)
	} else if !gte {
		t.Error("should be equal")
	}
}

func TestEqualToString(t *testing.T) {
	var firstUint64 string
	var secondUint64 string
	firstUint64 = "2"
	secondUint64 = "2"
	if gte, err := EqualTo(firstUint64, secondUint64); err != nil {
		t.Error(err)
	} else if !gte {
		t.Error("should be equal")
	}
}

func TestEqualToError(t *testing.T) {
	var firstUint64 bool
	var secondUint64 bool
	firstUint64 = true
	secondUint64 = true
	_, err := EqualTo(firstUint64, secondUint64)
	if err == nil || err.Error() != "unsupported comparison type: bool, kind: bool" {
		t.Errorf("Expected <unsupported comparison type: bool, kind: bool> ,received: <%+v>", err)
	}
}

func TestIfaceAsStringDefault(t *testing.T) {
	var test int8
	test = 2
	response := IfaceAsString(test)
	if !reflect.DeepEqual(response, "2") {
		t.Errorf("Expected <2> ,received: <%+v>", response)
	}

}

func TestReflectSumTimeDurationError(t *testing.T) {
	var time1 time.Duration
	var time2 bool
	time1 = 2
	time2 = true
	_, err := Sum(time1, time2)
	if err == nil || err.Error() != "cannot convert field: true to time.Duration" {
		t.Errorf("Expected <cannot convert field: true to time.Duration> ,received: <%+v>", err)
	}
}

func TestReflectSumInt64(t *testing.T) {
	var testInt64 int64
	var test2Int64 int64
	var expected int64
	testInt64 = 2
	test2Int64 = 3
	expected = 5
	sum, _ := Sum(testInt64, test2Int64)
	if !reflect.DeepEqual(sum, expected) {
		t.Errorf("Expected <%+v> ,received: <%+v>", expected, sum)
	}
}

func TestReflectSumFloat64Error(t *testing.T) {
	var testFloat64 float64
	var test2Float64 bool
	testFloat64 = 2.56
	test2Float64 = true
	_, err := Sum(testFloat64, test2Float64)
	if err == nil || err.Error() != "cannot convert field: true to float64" {
		t.Errorf("Expected <cannot convert field: true to float64> ,received: <%+v>", err)
	}
}

func TestReflectSumInt64Error(t *testing.T) {
	var testVar int64
	var test2Var bool
	testVar = 25354
	test2Var = true
	_, err := Sum(testVar, test2Var)
	if err == nil || err.Error() != "cannot convert field: true to int" {
		t.Errorf("Expected <cannot convert field: true to int> ,received: <%+v>", err)
	}
}

func TestReflectDifferenceTimeDurationError(t *testing.T) {
	var testVar time.Duration
	var test2Var bool
	testVar = 25354
	test2Var = true
	_, err := Difference(testVar, test2Var)
	if err == nil || err.Error() != "cannot convert field: true to time.Duration" {
		t.Errorf("Expected <cannot convert field: true to time.Duration> ,received: <%+v>", err)
	}
}

func TestReflectDifferenceFloat64Error(t *testing.T) {
	var testVar float64
	var test2Var bool
	testVar = 2.5
	test2Var = true
	_, err := Difference(testVar, test2Var)
	if err == nil || err.Error() != "cannot convert field: true to float64" {
		t.Errorf("Expected <cannot convert field: true to float64> ,received: <%+v>", err)
	}
}

func TestReflectDifferenceInt64Error(t *testing.T) {
	var testVar int64
	var test2Var int64
	var expected int64
	testVar = 6
	test2Var = 5
	expected = 1
	dif, _ := Difference(testVar, test2Var)
	if !reflect.DeepEqual(dif, expected) {
		t.Errorf("Expected <%+v> ,received: <%+v>", expected, dif)
	}
}

func TestReflectDifferenceDefault(t *testing.T) {
	var testVar bool
	var test2Var bool
	testVar = true
	test2Var = true
	_, err := Difference(testVar, test2Var)
	if err == nil || err.Error() != "unsupported type" {
		t.Errorf("Expected <unsupported type> ,received: <%+v>", err)
	}
}

func TestReflectMultiplyDefault(t *testing.T) {
	var testVar bool
	var test2Var bool
	testVar = true
	test2Var = true
	_, err := Multiply(testVar, test2Var)
	if err == nil || err.Error() != "unsupported type" {
		t.Errorf("Expected <unsupported type> ,received: <%+v>", err)
	}
}

func TestReflectMultiplyFloat64Error(t *testing.T) {
	var testVar float64
	var test2Var bool
	testVar = 2.5
	test2Var = true
	_, err := Multiply(testVar, test2Var)
	if err == nil || err.Error() != "cannot convert field: true to float64" {
		t.Errorf("Expected <cannot convert field: true to float64> ,received: <%+v>", err)
	}
}

func TestReflectMultiplyInt64(t *testing.T) {
	var testInt64 int64
	var test2Int64 int64
	var expected int64
	testInt64 = 2
	test2Int64 = 3
	expected = 6
	mul, _ := Multiply(testInt64, test2Int64)
	if !reflect.DeepEqual(mul, expected) {
		t.Errorf("Expected <%+v> ,received: <%+v>", expected, mul)
	}
}

func TestReflectDivideInt64(t *testing.T) {
	var testInt64 int64
	var test2Int64 int64
	var expected int64
	testInt64 = 4
	test2Int64 = 2
	expected = 2
	div, _ := Divide(testInt64, test2Int64)
	if !reflect.DeepEqual(div, expected) {
		t.Errorf("Expected <%+v> ,received: <%+v>", expected, div)
	}
}

func TestReflectDivideDefault(t *testing.T) {
	var testVar bool
	var test2Var bool
	testVar = true
	test2Var = true
	_, err := Divide(testVar, test2Var)
	if err == nil || err.Error() != "unsupported type" {
		t.Errorf("Expected <unsupported type> ,received: <%+v>", err)
	}
}

func TestReflectDivideFloat64Error(t *testing.T) {
	var testVar float64
	var test2Var bool
	testVar = 2.5
	test2Var = true
	_, err := Divide(testVar, test2Var)
	if err == nil || err.Error() != "cannot convert field: true to float64" {
		t.Errorf("Expected <cannot convert field: true to float64> ,received: <%+v>", err)
	}
}

func TestReflectFieldMethodInterfaceArray(t *testing.T) {
	obj := []int{2, 3, 5, 7, 11, 13}
	fldName := "0"
	result, _ := ReflectFieldMethodInterface(obj, fldName)
	if !reflect.DeepEqual(result, 2) {
		t.Errorf("Expected <2> ,received: <%+v>", result)
	}
}

func TestReflectFieldMethodInterfaceMap(t *testing.T) {
	var obj = map[string]string{"field1": "val1", "field2": "val2"}
	fldName := "field1"
	result, _ := ReflectFieldMethodInterface(obj, fldName)
	if !reflect.DeepEqual(result, "val1") {
		t.Errorf("Expected <val1> ,received: <%+v>", result)
	}

}

func TestReflectFieldMethodInterfaceArrayError(t *testing.T) {
	obj := []int{2, 3, 5, 7, 11, 13}
	fldName := "test"
	_, err := ReflectFieldMethodInterface(obj, fldName)
	if err == nil || err.Error() != "strconv.Atoi: parsing \"test\": invalid syntax" {
		t.Errorf("Expected <strconv.Atoi: parsing \"test\": invalid syntax> ,received: <%+v>", err)
	}
}

func TestReflectFieldMethodInterfaceArrayError2(t *testing.T) {
	obj := []int{2, 3, 5, 7, 11, 13}
	fldName := "6"
	_, err := ReflectFieldMethodInterface(obj, fldName)
	if err == nil || err.Error() != "index out of range" {
		t.Errorf("Expected <index out of range> ,received: <%+v>", err)
	}
}

func TestReflectFieldMethodInterfaceArrayDefault(t *testing.T) {
	obj := "test"
	fldName := "test2"
	_, err := ReflectFieldMethodInterface(obj, fldName)
	if err == nil || err.Error() != "unsupported field kind: string" {
		t.Errorf("Expected <unsupported field kind: string> ,received: <%+v>", err)
	}
}

func (_ *TestA) TestFuncWithParamError2() (string, string, string) {
	return "Invalid", "invalid2", "invalid3"
}

func TestReflectFieldMethodInterfaceElseError1(t *testing.T) {
	a := &TestA{StrField: "TestStructField"}
	_, err := ReflectFieldMethodInterface(a, "TestFuncWithParam")
	if err == nil || err.Error() != "invalid function called" {
		t.Errorf("Expected <invalid function called> ,received: <%+v>", err)
	}
}

func TestReflectFieldMethodInterfaceElseError2(t *testing.T) {
	a := &TestA{StrField: "TestStructField"}
	_, err := ReflectFieldMethodInterface(a, "TestFuncWithParamError2")
	if err == nil || err.Error() != "invalid function called" {
		t.Errorf("Expected <invalid function called> ,received: <%+v>", err)
	}
}

func (_ *TestA) TestFuncWithParamError3() (string, string) {
	return "Invalid", "invalid2"
}

func TestReflectFieldMethodInterfaceElseError3(t *testing.T) {
	a := &TestA{StrField: "TestStructField"}
	_, err := ReflectFieldMethodInterface(a, "TestFuncWithParamError3")
	if err == nil || err.Error() != "invalid function called" {
		t.Errorf("Expected <invalid function called> ,received: <%+v>", err)
	}
}

func TestSumTimeTimeError(t *testing.T) {
	day1 := time.Now()
	day2 := "testValue"
	_, err := Sum(day1, day2)
	if err == nil || err.Error() != "time: invalid duration \"testValue\"" {
		t.Errorf("Expected <time: invalid duration testValue> ,received: <%+v>", err)
	}

}

func TestSumTimeTime(t *testing.T) {
	day1 := time.Now()
	day2 := time.Hour
	expected := day1.Add(day2)
	sum, _ := Sum(day1, day2)
	if !reflect.DeepEqual(sum, expected) {
		t.Errorf("Expected <%+v> ,received: <%+v>", expected, sum)
	}
}

func TestDifferenceTimeTimeError(t *testing.T) {
	_, err := Difference(time.Now(), "cat")
	if err == nil || err.Error() != "time: invalid duration \"cat\"" {
		t.Errorf("Expected <time: invalid duration \"cat\"> ,received: <%+v>", err)
	}
}

func TestDifferenceInt64Error(t *testing.T) {
	_, err := Difference(int64(2), "cat")
	if err == nil || err.Error() != "strconv.ParseInt: parsing \"cat\": invalid syntax" {
		t.Errorf("Expected <strconv.ParseInt: parsing \"cat\": invalid syntax> ,received: <%+v>", err)
	}
}

func TestDivideInt64Error(t *testing.T) {
	_, err := Divide(int64(2), "cat")
	if err == nil || err.Error() != "strconv.ParseInt: parsing \"cat\": invalid syntax" {
		t.Errorf("Expected <strconv.ParseInt: parsing \"cat\": invalid syntax> ,received: <%+v>", err)
	}
}

func TestMultiplyInt64Error(t *testing.T) {
	_, err := Multiply(int64(2), "cat")
	if err == nil || err.Error() != "strconv.ParseInt: parsing \"cat\": invalid syntax" {
		t.Errorf("Expected <strconv.ParseInt: parsing \"cat\": invalid syntax> ,received: <%+v>", err)
	}
}
