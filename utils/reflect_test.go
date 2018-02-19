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

func TestReflectAsMapStringIface(t *testing.T) {
	mystruct := struct {
		Title       string
		Count       int
		Count64     int64
		Val         float64
		ExtraFields map[string]interface{}
	}{"Title1", 5, 6, 7.3, map[string]interface{}{"a": "Title2", "b": 15, "c": int64(16), "d": 17.3}}
	expectOutMp := map[string]interface{}{
		"Title":       "Title1",
		"Count":       5,
		"Count64":     int64(6),
		"Val":         7.3,
		"ExtraFields": map[string]interface{}{"a": "Title2", "b": 15, "c": int64(16), "d": 17.3},
	}
	if outMp, err := AsMapStringIface(mystruct); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectOutMp, outMp) {
		t.Errorf("Expecting: %+v, received: %+v", expectOutMp, outMp)
	}
}

func TestGreaterThan(t *testing.T) {
	if _, err := GreaterThan(1, 1.2, false); err == nil || err.Error() != "incomparable" {
		t.Error(err)
	}
	if _, err := GreaterThan(struct{}{},
		map[string]interface{}{"a": "a"}, false); err == nil || err.Error() != "incomparable" {
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
	if gte, err := GreaterThan(time.Duration(2*time.Second),
		time.Duration(1*time.Second), false); err != nil {
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
}

func TestStringToInterface(t *testing.T) {
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

func TestCastFieldIfToString(t *testing.T) {
	val := interface{}("string1")
	if resVal, converted := CastFieldIfToString(val); !converted || resVal != "string1" {
		t.Error(resVal, converted)
	}
	val = interface{}(123)
	if resVal, converted := CastFieldIfToString(val); !converted || resVal != "123" {
		t.Error(resVal, converted)
	}
	val = interface{}([]byte("byte_val"))
	if resVal, converted := CastFieldIfToString(val); !converted || resVal != "byte_val" {
		t.Error(resVal, converted)
	}
	val = interface{}(true)
	if resVal, converted := CastFieldIfToString(val); !converted || resVal != "true" {
		t.Error(resVal, converted)
	}
	if strVal, cast := CastFieldIfToString(time.Duration(1 * time.Second)); !cast {
		t.Error("cannot cast time.Duration")
	} else if strVal != "1s" {
		t.Errorf("received: %s", strVal)
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
	eItm := time.Duration(time.Second)
	if itmConvert, err := IfaceAsDuration(interface{}(time.Duration(time.Second))); err != nil {
		t.Error(err)
	} else if eItm != itmConvert {
		t.Errorf("received: %+v", itmConvert)
	}
	if itmConvert, err := IfaceAsDuration(interface{}(float64(1.0))); err != nil {
		t.Error(err)
	} else if eItm != itmConvert {
		t.Errorf("received: %+v", itmConvert)
	}
	if itmConvert, err := IfaceAsDuration(interface{}(int64(1000000000))); err != nil {
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
	val = interface{}(time.Duration(6))
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
