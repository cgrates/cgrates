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

package config

import (
	"reflect"
	"testing"

	"github.com/cgrates/cgrates/utils"
)

func TestNewObjectDP(t *testing.T) {
	object := "cgrates.org"
	objDp := &ObjectDP{
		obj:   "cgrates.org",
		cache: make(map[string]interface{}),
	}
	if received := NewObjectDP(object); !reflect.DeepEqual(objDp, received) {
		t.Errorf("Expected %+v, received %+v", objDp, received)
	}
}

func TestRemoteHostObjDP(t *testing.T) {
	expected := utils.LocalAddr()
	objDp := &ObjectDP{
		obj:   "cgrates.org",
		cache: make(map[string]interface{}),
	}
	if received := objDp.RemoteHost(); !reflect.DeepEqual(expected, received) {
		t.Errorf("Expected %+v, received %+v", expected, received)
	}
}

func TestStringObjDP(t *testing.T) {
	objDp := &ObjectDP{
		obj:   "cgrates.org",
		cache: make(map[string]interface{}),
	}
	expected := `"cgrates.org"`
	if received := objDp.String(); !reflect.DeepEqual(expected, received) {
		t.Errorf("Expected %+v, received %+v", expected, received)
	}
}

func TestFieldAsInterfaceObjDPSliceOfInt(t *testing.T) {
	object := []string{"1"}
	objDp := &ObjectDP{
		obj:   []int{12, 13},
		cache: make(map[string]interface{}),
	}
	expected := 13
	if received, err := objDp.FieldAsInterface(object); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(received, expected) {
		t.Errorf("Expected %+v, received %+v", expected, received)
	}
}

func TestFieldAsInterfaceObjDPInvalidSyntax(t *testing.T) {
	object := []string{"1]"}
	objDp := &ObjectDP{
		obj:   []int{12, 13},
		cache: make(map[string]interface{}),
	}
	expected := "strconv.Atoi: parsing \"1]\": invalid syntax"
	if _, err := objDp.FieldAsInterface(object); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestFieldAsInterfaceObjDPInvalidFormat(t *testing.T) {
	object := []string{"invalid[path"}
	objDp := &ObjectDP{
		obj:   []int{12, 13},
		cache: make(map[string]interface{}),
	}
	expected := "filter rule <path> needs to end in ]"
	if _, err := objDp.FieldAsInterface(object); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestFieldAsInterfaceObjDPCache(t *testing.T) {
	object := []string{"validPath"}
	objDp := &ObjectDP{
		cache: map[string]interface{}{
			"validPath": "cgrates.org",
		},
	}
	expected := "cgrates.org"
	if rcv, err := objDp.FieldAsInterface(object); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rcv) {
		t.Errorf("Expected %+v, received %+v", expected, rcv)
	}
}

func TestFieldAsInterfaceObjDPChangedObject(t *testing.T) {
	object := []string{"0[1]"}
	objDp := &ObjectDP{
		obj:   []int{1},
		cache: map[string]interface{}{},
	}
	expected := "unsupported field kind: int"
	if _, err := objDp.FieldAsInterface(object); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestFieldAsInterfaceObjDPValid1(t *testing.T) {
	object := []string{"0[1]"}
	objDp := &ObjectDP{
		obj: []map[string]interface{}{
			{
				"1": 1,
				"2": 2,
			},
		},
		cache: map[string]interface{}{},
	}
	if rcv, err := objDp.FieldAsInterface(object); err != nil {
		t.Error(err)
	} else if rcv != 1 {
		t.Errorf("Expected %+v, received %+v", 1, rcv)
	}
}

func TestFieldAsStringObjDP(t *testing.T) {
	object := []string{"0[1]"}
	objDp := &ObjectDP{
		obj: []map[string]interface{}{
			{
				"1": 1,
				"2": 2,
			},
		},
		cache: map[string]interface{}{},
	}
	if rcv, err := objDp.FieldAsString(object); err != nil {
		t.Error(err)
	} else if rcv != "1" {
		t.Errorf("Expected %+v, received %+v", 1, rcv)
	}
}

func TestFieldAsStringError(t *testing.T) {
	object := []string{"0[1]"}
	objDp := &ObjectDP{
		obj:   []int{1},
		cache: map[string]interface{}{},
	}
	expected := "unsupported field kind: int"
	if _, err := objDp.FieldAsString(object); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestFieldAsInterfaceObjDPMultiplePaths(t *testing.T) {
	type aNewStruct struct {
		Field1 int
		Field2 int
	}
	type pNewStruct struct {
		Field3 aNewStruct
		Field4 int
		Field5 []string
	}
	objDp := &ObjectDP{
		obj: pNewStruct{
			Field3: aNewStruct{
				Field1: 2,
				Field2: 4,
			},
			Field4: 2,
			Field5: []string{"1", "2"},
		},
		cache: map[string]interface{}{},
	}
	if rcv, err := objDp.FieldAsInterface([]string{"Field3", "Field2"}); err != nil {
		t.Error(err)
	} else if rcv != 4 {
		t.Errorf("Expected %+v, received %+v", 4, rcv)
	}

	if rcv, err := objDp.FieldAsInterface([]string{"Field5[0]"}); err != nil {
		t.Error(err)
	} else if rcv != "1" {
		t.Errorf("Expected %+v, received %+v", "1", rcv)
	}
}

func TestFieldAsInterface(t *testing.T) {
	type aNewStruct struct {
		Field1 int
		Field2 int
	}
	type pNewStruct struct {
		Field3 aNewStruct
		Field4 int
		Field5 []string
	}
	objDp := &ObjectDP{
		obj: pNewStruct{
			Field3: aNewStruct{
				Field1: 2,
				Field2: 4,
			},
			Field4: 2,
			Field5: []string{""},
		},
		cache: map[string]interface{}{
			"field1": nil,
		},
	}
	_, err := objDp.FieldAsInterface([]string{"field1"})
	if err == nil || err != utils.ErrNotFound {
		t.Error(err)
	}
}
