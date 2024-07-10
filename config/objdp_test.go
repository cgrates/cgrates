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
		cache: make(map[string]any),
	}
	if received := NewObjectDP(object); !reflect.DeepEqual(objDp, received) {
		t.Errorf("Expected %+v, received %+v", objDp, received)
	}
}

func TestStringObjDP(t *testing.T) {
	objDp := &ObjectDP{
		obj:   "cgrates.org",
		cache: make(map[string]any),
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
		cache: make(map[string]any),
	}
	objDp2 := &ObjectDP{
		obj:   []any{},
		cache: make(map[string]any),
	}
	expected := 13
	if received, err := objDp.FieldAsInterface(object); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(received, expected) {
		t.Errorf("Expected %+v, received %+v", expected, received)
	} else if _, err = objDp2.FieldAsInterface([]string{"test"}); err != utils.ErrNotFound {
		t.Error(err)
	}
}

func TestFieldAsInterfaceObjDPInvalidSyntax(t *testing.T) {
	object := []string{"1]"}
	objDp := &ObjectDP{
		obj:   []int{12, 13},
		cache: make(map[string]any),
	}
	expected := utils.ErrNotFound.Error()
	if _, err := objDp.FieldAsInterface(object); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestFieldAsInterfaceObjDPInvalidFormat(t *testing.T) {
	object := []string{"invalid[path"}
	objDp := &ObjectDP{
		obj:   []int{12, 13},
		cache: make(map[string]any),
	}
	expected := "filter rule <path> needs to end in ]"
	if _, err := objDp.FieldAsInterface(object); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestFieldAsInterfaceObjDPCache(t *testing.T) {
	object := []string{"validPath"}
	objDp := &ObjectDP{
		cache: map[string]any{
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
		cache: map[string]any{},
	}
	expected := "unsupported field kind: int"
	if _, err := objDp.FieldAsInterface(object); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestFieldAsInterfaceObjDPValid1(t *testing.T) {
	object := []string{"0[1]"}
	objDp := &ObjectDP{
		obj: []map[string]any{
			{
				"1": 1,
				"2": 2,
			},
		},
		cache: map[string]any{},
	}
	if rcv, err := objDp.FieldAsInterface(object); err != nil {
		t.Error(err)
	} else if rcv != 1 {
		t.Errorf("Expected %+v, received %+v", 1, rcv)
	}
	if _, err := objDp.FieldAsInterface([]string{"val"}); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	}
}

func TestFieldAsStringObjDP(t *testing.T) {
	object := []string{"0[1]"}
	objDp := &ObjectDP{
		obj: []map[string]any{
			{
				"1": 1,
				"2": 2,
			},
		},
		cache: map[string]any{},
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
		cache: map[string]any{},
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
		cache: map[string]any{},
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

func TestObjectDPFieldAsInterfaceError(t *testing.T) {
	type test struct {
		Field string
	}
	tst := test{
		Field: "test",
	}
	objDP := &ObjectDP{
		obj:   tst,
		cache: map[string]any{"Field": nil},
	}

	rcv, err := objDP.FieldAsInterface([]string{"Field"})

	if err != nil {
		if err.Error() != "NOT_FOUND" {
			t.Error(err)
		}
	} else {
		t.Error("was expecting an error")
	}

	if rcv != nil {
		t.Error(rcv)
	}
}
