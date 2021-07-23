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

func TestSafeMapStorageString(t *testing.T) {
	ms := &SafeMapStorage{
		MapStorage: MapStorage{
			"field1": 2,
		},
	}
	expected := "{\"field1\":2}"
	if reply := ms.String(); reply != expected {
		t.Errorf("Expected %s \n but received \n %s", expected, reply)
	}
}

func TestSafeMapStorageFieldAsInterface(t *testing.T) {
	ms := &SafeMapStorage{
		MapStorage: MapStorage{
			"field1": 2,
		},
	}

	input := []string{"field1"}
	expected := 2
	if reply, err := ms.FieldAsInterface(input); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, reply) {
		t.Errorf("Expected %d \n but received \n %d", expected, reply)
	}
}

func TestSafeMapStorageFieldAsString(t *testing.T) {
	ms := &SafeMapStorage{
		MapStorage: MapStorage{
			"field1": 2,
		},
	}

	input := []string{"field1"}
	expected := "2"
	if reply, err := ms.FieldAsString(input); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, reply) {
		t.Errorf("Expected %s \n but received \n %s", expected, reply)
	}
}

func TestSafeMapStorageSet(t *testing.T) {
	ms := &SafeMapStorage{
		MapStorage: MapStorage{
			"field1": 2,
		},
	}

	expected := &SafeMapStorage{
		MapStorage: MapStorage{
			"field1": 2,
			"field2": 3,
		},
	}

	if err := ms.Set([]string{"field2"}, 3); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(ms, expected) {
		t.Errorf("Expected %v \n but received \n %v", expected, ms)
	}
}

func TestSafeMapStorageGetKeys(t *testing.T) {
	ms := &SafeMapStorage{
		MapStorage: MapStorage{
			"field1": 2,
			"field2": 3,
		},
	}

	expected := []string{"*req.field1", "*req.field2"}
	if reply := ms.GetKeys(false, 0, MetaReq); !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expected %v \n but received \n %v", expected, reply)
	}
}

func TestSafeMapStorageRemove(t *testing.T) {
	ms := &SafeMapStorage{
		MapStorage: MapStorage{
			"field1": 2,
			"field2": 3,
		},
	}

	expected := &SafeMapStorage{
		MapStorage: MapStorage{
			"field1": 2,
		},
	}

	if err := ms.Remove([]string{"field2"}); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(ms, expected) {
		t.Errorf("Expected %v \n but received \n %v", expected, ms)
	}
}

func TestSafeMapStorageClone(t *testing.T) {
	ms := &SafeMapStorage{
		MapStorage: MapStorage{
			"field1": 2,
			"field2": 3,
		},
	}

	expected := &SafeMapStorage{
		MapStorage: MapStorage{
			"field1": 2,
			"field2": 3,
		},
	}

	if reply := ms.Clone(); !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expected %v \n but received \n %v", expected, reply)
	}
}
