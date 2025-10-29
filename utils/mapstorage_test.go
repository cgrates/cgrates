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
	"reflect"
	"strings"
	"testing"
)

func TestNavMapGetFieldAsString(t *testing.T) {
	nM := MapStorage{
		"FirstLevel": map[string]any{
			"SecondLevel": map[string]any{
				"ThirdLevel": map[string]any{
					"Fld1": "Val1",
				},
			},
		},
		"AnotherFirstLevel": "ValAnotherFirstLevel",
	}
	eVal := "Val1"
	if strVal, err := nM.FieldAsString(
		strings.Split("FirstLevel>SecondLevel>ThirdLevel>Fld1", ">")); err != nil {
		t.Error(err)
	} else if strVal != eVal {
		t.Errorf("expecting: <%s> received: <%s>", eVal, strVal)
	}
	eVal = "ValAnotherFirstLevel"
	if strVal, err := nM.FieldAsString(
		strings.Split("AnotherFirstLevel", ">")); err != nil {
		t.Error(err)
	} else if strVal != eVal {
		t.Errorf("expecting: <%s> received: <%s>", eVal, strVal)
	}
	fPath := "NonExisting>AnotherFirstLevel"
	if _, err := nM.FieldAsString(strings.Split(fPath, ">")); err.Error() !=
		ErrNotFound.Error() {
		t.Error(err)
	}
}

type myEv map[string]any

func (ev myEv) AsMapStorage() (MapStorage, error) {
	return MapStorage(ev), nil
}

func TestNavMapAsMapStorage(t *testing.T) {
	myData := myEv{
		"FirstLevel": map[string]any{
			"SecondLevel": map[string]any{
				"ThirdLevel": map[string]any{
					"Fld1": 123.123,
				},
			},
		},
		"FistLever2": map[string]any{
			"SecondLevel2": map[string]any{
				"Field2": 123,
			},
			"Field3": "Value3",
		},
		"Field4": &testStruct{
			Item1: "Ten",
			Item2: 10,
		},
	}

	eNavMap := MapStorage{
		"FirstLevel": map[string]any{
			"SecondLevel": map[string]any{
				"ThirdLevel": map[string]any{
					"Fld1": 123.123,
				},
			},
		},
		"FistLever2": map[string]any{
			"SecondLevel2": map[string]any{
				"Field2": 123,
			},
			"Field3": "Value3",
		},
		"Field4": &testStruct{
			Item1: "Ten",
			Item2: 10,
		},
	}

	if rcv, err := myData.AsMapStorage(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eNavMap, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", eNavMap, rcv)
	}
}

type testStruct struct {
	Item1 string
	Item2 int
}

func TestMapStorageNavMapAdd2(t *testing.T) {
	nM := MapStorage{}
	path := []string{"FistLever2", "SecondLevel2", "Field2"}
	data := 123
	nM.Set(path, data)
	path = []string{"FirstLevel", "SecondLevel", "ThirdLevel", "Fld1"}
	data1 := 123.123
	nM.Set(path, data1)
	path = []string{"FistLever2", "Field3"}
	data2 := "Value3"
	nM.Set(path, data2)
	path = []string{"Field4"}
	data3 := &testStruct{
		Item1: "Ten",
		Item2: 10,
	}
	nM.Set(path, data3)
	eNavMap := MapStorage{
		"FirstLevel": MapStorage{
			"SecondLevel": MapStorage{
				"ThirdLevel": MapStorage{
					"Fld1": 123.123,
				},
			},
		},
		"FistLever2": MapStorage{
			"SecondLevel2": MapStorage{
				"Field2": 123,
			},
			"Field3": "Value3",
		},
		"Field4": &testStruct{
			Item1: "Ten",
			Item2: 10,
		},
	}
	if !reflect.DeepEqual(nM, eNavMap) {
		t.Errorf("Expecting: %+v, received: %+v", eNavMap, nM)
	}
}

func TestNavMapString(t *testing.T) {
	myData := map[string]any{
		"FirstLevel": map[string]any{
			"SecondLevel": map[string]any{
				"ThirdLevel": map[string]any{
					"Fld1": "Val1",
				},
			},
		},
		"FistLever2": map[string]any{
			"SecondLevel2": map[string]any{
				"Field2": "Value2",
			},
			"Field3": "Value3",
		},
		"Field4": "Val4",
	}
	nM := MapStorage(myData)
	eStr := ToJSON(myData)
	if !reflect.DeepEqual(nM.String(), eStr) {
		t.Errorf("Expecting: %+v, received: %+v", eStr, nM.String())
	}
}

func TestNavMapGetField(t *testing.T) {
	nM := MapStorage{
		"FirstLevel": map[string]any{
			"SecondLevel": map[string]any{
				"ThirdLevel": map[string]any{
					"Fld1": []any{"Val1", "Val2"},
				},
			},
		},
		"FirstLevel2": map[string]any{
			"SecondLevel2": []map[string]any{
				{
					"ThirdLevel2": map[string]any{
						"Fld1": "Val1",
					},
				},
				{
					"Count": 10,
					"ThirdLevel2": map[string]any{
						"Fld2": []string{"Val1", "Val2", "Val3"},
					},
				},
			},
		},
		"AnotherFirstLevel": "ValAnotherFirstLevel",
	}
	pth := []string{"FirstLevel", "SecondLevel", "ThirdLevel", "Fld1[0]"}
	eFld := "Val1"
	if fld, err := nM.FieldAsInterface(pth); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eFld, fld) {
		t.Errorf("expecting: %s, received: %s", ToIJSON(eFld), ToIJSON(fld))
	}
	eFld2 := map[string]any{"Fld1": "Val1"}
	pth = []string{"FirstLevel2", "SecondLevel2[0]", "ThirdLevel2"}
	if fld, err := nM.FieldAsInterface(pth); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eFld2, fld) {
		t.Errorf("expecting: %s, received: %s", ToIJSON(eFld2), ToIJSON(fld))
	}
	eFld3 := "ValAnotherFirstLevel"
	pth = []string{"AnotherFirstLevel"}
	if fld, err := nM.FieldAsInterface(pth); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eFld3, fld) {
		t.Errorf("expecting: %s, received: %s", ToIJSON(eFld3), ToIJSON(fld))
	}
	pth = []string{"AnotherFirstLevel2"}
	if _, err := nM.FieldAsInterface(pth); err == nil || err != ErrNotFound {
		t.Error(err)
	}
	pth = []string{"FirstLevel", "SecondLevel[1]", "ThirdLevel", "Fld1[0]"}
	if _, err := nM.FieldAsInterface(pth); err == nil || err != ErrNotFound {
		t.Error(err)
	}
}

func TestMapStorageNavMapFieldAsInterface(t *testing.T) {
	nm := MapStorage{
		"SlcString": []string{"val1", "val2"},
		"SlcAny":    []any{"val1"},
		"MapPtr":    &map[string]any{"test1": 1},
		"Array":     [1]string{"val1"},
		"DP":        MapStorage{"test": "val1"},
		"MapInt":    map[string]int{"test": 1},
		"SlcDP":     []DataProvider{MapStorage{"test": "val1"}},
		"SlcMS":     []MapStorage{{"test": "val1"}},
		"SlcMapAny": []map[string]any{{"test": "val1"}},
		"SlcAnyMap": []any{map[string]any{"test": "val1"}},
		"SlcAnyDP":  []any{MapStorage{"test": "val1"}},
	}

	type exp struct {
		val any
		err error
	}

	tests := []struct {
		name string
		arg  []string
		exp  exp
	}{
		{
			name: "empty argument",
			arg:  []string{},
			exp:  exp{nil, nil},
		},
		{
			name: "field is slice of strings, index not found",
			arg:  []string{"SlcString[3]"},
			exp:  exp{nil, ErrNotFound},
		},
		{
			name: "field is slice of strings",
			arg:  []string{"SlcString[1]"},
			exp:  exp{"val2", nil},
		},
		{
			name: "field is slice of any",
			arg:  []string{"SlcAny[1]"},
			exp:  exp{nil, ErrNotFound},
		},
		{
			name: "not slice or array and ptr",
			arg:  []string{"MapPtr[1]"},
			exp:  exp{nil, ErrNotFound},
		},
		{
			name: "array",
			arg:  []string{"Array[1]"},
			exp:  exp{nil, ErrNotFound},
		},
		{
			name: "array",
			arg:  []string{"Array[0]"},
			exp:  exp{"val1", nil},
		},
		{
			name: "data storage",
			arg:  []string{"DP", "test"},
			exp:  exp{"val1", nil},
		},
		{
			name: "map of int",
			arg:  []string{"MapInt", "test"},
			exp:  exp{map[string]int{"test": 1}, ErrWrongPath},
		},
		{
			name: "slice of DataProvider not found",
			arg:  []string{"SlcDP[1]", "test"},
			exp:  exp{nil, ErrNotFound},
		},
		{
			name: "slice of DataProvider",
			arg:  []string{"SlcDP[0]", "test"},
			exp:  exp{"val1", nil},
		},
		{
			name: "slice of MapStorage not found",
			arg:  []string{"SlcMS[1]", "test"},
			exp:  exp{nil, ErrNotFound},
		},
		{
			name: "slice of MapStorage",
			arg:  []string{"SlcMS[0]", "test"},
			exp:  exp{"val1", nil},
		},
		{
			name: "slice of map[string]any not found",
			arg:  []string{"SlcMapAny[1]", "test"},
			exp:  exp{nil, ErrNotFound},
		},
		{
			name: "slice of any",
			arg:  []string{"SlcAnyMap[1]", "test"},
			exp:  exp{nil, ErrNotFound},
		},
		{
			name: "slice of any map",
			arg:  []string{"SlcAnyMap[0]", "test"},
			exp:  exp{"val1", nil},
		},
		{
			name: "slice of any data provider",
			arg:  []string{"SlcAnyDP[0]", "test"},
			exp:  exp{"val1", nil},
		},
	}

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rcv, err := nm.FieldAsInterface(tt.arg)

			if i == 0 {
				if err == nil {
					t.Fatal("was expecting an error")
				}
			} else {
				if err != tt.exp.err {
					t.Fatalf("recived %s, expected %s", err, tt.exp.err)
				}
			}

			if !reflect.DeepEqual(rcv, tt.exp.val) {
				t.Errorf("recived %v, expected %v", rcv, tt.exp.val)
			}
		})
	}
}

func TestMapStorageNavMapGetKeys(t *testing.T) {
	tests := []struct {
		name     string
		arg      bool
		expected []string
	}{
		{
			name: "only first layer of keywords",
			arg:  false,
			expected: []string{"SlcAny", "SlcString", "MS", "MP", "Test",
				"SlcMapStorage", "SlcDataStorage", "SlcMap", "Uint8"},
		},
		{
			name: "all layers",
			arg:  true,
			expected: []string{"SlcAny[0]", "SlcString[0]", "MS.test", "MP.test2", "Test",
				"SlcMapStorage[0].test3", "SlcDataStorage[0].test4", "SlcMap[0].test5", "Uint8"},
		},
	}

	var num uint8 = 1

	ms := MapStorage{
		"MS":             MapStorage{"test": 1},
		"MP":             map[string]any{"test2": 2},
		"Test":           "test string",
		"SlcMapStorage":  []MapStorage{{"test3": 3}},
		"SlcDataStorage": []dataStorage{MapStorage{"test4": 4}},
		"SlcMap":         []map[string]any{{"test5": 5}},
		"SlcAny":         []any{map[string]any{"test6": 6}},
		"SlcString":      []string{"test7"},
		"Uint8":          num,
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rcv := ms.GetKeys(tt.arg)
			var has bool

			for _, vRcv := range rcv {
				has = false
				for _, vExp := range tt.expected {
					if vRcv == vExp {
						has = true
					} else {
						continue
					}
				}
			}

			if !has {
				t.Errorf("recived %+v, expected %+v", rcv, tt.expected)
			}
		})
	}
}

func TestMapStorageMapRemove(t *testing.T) {
	tests := []struct {
		name string
		arg  []string
		want error
	}{
		{
			name: "empty path",
			arg:  []string{},
			want: ErrWrongPath,
		},
		{
			name: "non existing path",
			arg:  []string{"abc"},
			want: nil,
		},
		{
			name: "one argument in path",
			arg:  []string{"Test"},
			want: nil,
		},
		{
			name: "case dataStorage",
			arg:  []string{"MS", "test"},
			want: nil,
		},
		{
			name: "case map[string]any",
			arg:  []string{"MP", "test2"},
			want: nil,
		},
		{
			name: "wrong path",
			arg:  []string{"Test1", "test3"},
			want: ErrWrongPath,
		},
	}

	ms := MapStorage{
		"MS":    MapStorage{"test": 1},
		"MP":    map[string]any{"test2": 2},
		"Test":  "test string",
		"Test1": "test2 string",
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			err := ms.Remove(tt.arg)

			if err != tt.want {
				t.Errorf("expected %s, recived %s", tt.want, err)
			}
		})
	}
}

func TestMapStorageGetPathFromValue(t *testing.T) {
	type args struct {
		in     reflect.Value
		prefix string
	}
	tests := []struct {
		name string
		args args
		exp  []string
	}{
		{
			name: "ponter slice",
			args: args{reflect.ValueOf(&[]string{"test"}), "test"},
			exp:  []string{"test[0]"},
		},
		{
			name: "map",
			args: args{reflect.ValueOf(map[string]string{"test": "test"}), "test"},
			exp:  []string{"testtest"},
		},
		{
			name: "struct",
			args: args{reflect.ValueOf(struct{ test string }{"test"}), "test"},
			exp:  []string{"testtest"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rcv := getPathFromValue(tt.args.in, tt.args.prefix)
			var has bool

			for _, vRcv := range rcv {
				has = false
				for _, vExp := range tt.exp {
					if vRcv == vExp {
						has = true
					} else {
						continue
					}
				}
			}

			if !has {
				t.Errorf("recived %+v, expected %+v", rcv, tt.exp)
			}
		})
	}
}

func TestMapStorageRemoteHost(t *testing.T) {
	ms := MapStorage{
		"MS":    MapStorage{"test": 1},
		"MP":    map[string]any{"test2": 2},
		"Test":  "test string",
		"Test1": "test2 string",
	}
	rcv := ms.RemoteHost()
	rcvStr := fmt.Sprintf("%T/", rcv)

	if rcvStr != "*utils.NetAddr/" {
		t.Errorf("wrong return %s", rcvStr)
	}
}

func TestMapStorageNavMapSet(t *testing.T) {
	tests := []struct {
		name string
		path []string
		val  any
		err  error
	}{
		{
			name: "empty path",
			path: []string{},
			val:  1,
			err:  ErrWrongPath,
		},
		{
			name: "non supported data type",
			path: []string{"Test", "test"},
			val:  1,
			err:  ErrWrongPath,
		},
		{
			name: "non supported data type",
			path: []string{"MP", "test"},
			val:  1,
			err:  nil,
		},
	}

	ms := MapStorage{
		"MS":    MapStorage{"test": 1},
		"MP":    map[string]any{"test2": 2},
		"Test":  "test",
		"Test1": "test2 string",
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ms.Set(tt.path, tt.val)

			if err != tt.err {
				t.Errorf("recived %s, expected %s", err, tt.err)
			}
		})
	}
}
