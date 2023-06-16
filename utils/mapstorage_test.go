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

func TestNavMapAdd2(t *testing.T) {
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

func TestNavMapFieldAsInterface(t *testing.T) {
	nM := MapStorage{
		"FirstLevel": map[string]any{
			"SecondLevel": []map[string]any{
				{
					"ThirdLevel": map[string]any{
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

	path := []string{"FirstLevel", "SecondLevel[0]", "Count"}
	expErr := ErrNotFound
	var eVal any = nil
	if _, err := nM.FieldAsInterface(path); err != nil && err.Error() != expErr.Error() {
		t.Errorf("Expected error: %s, received error: %v", expErr.Error(), err)
	}

	path = []string{"AnotherFirstLevel", "SecondLevel", "Count"}
	expErr = ErrWrongPath
	if _, err := nM.FieldAsInterface(path); err != nil && err.Error() != expErr.Error() {
		t.Errorf("Expected error: %s, received error: %v", expErr.Error(), err)
	}

	path = []string{"FirstLevel", "SecondLevel[1]", "Count"}
	eVal = 10
	if rplyVal, err := nM.FieldAsInterface(path); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eVal, rplyVal) {
		t.Errorf("Expected: %s , received: %s", ToJSON(eVal), ToJSON(rplyVal))
	}

	path = []string{"FirstLevel", "SecondLevel[1]", "ThirdLevel2", "Fld2"}
	eVal = []string{"Val1", "Val2", "Val3"}
	if rplyVal, err := nM.FieldAsInterface(path); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eVal, rplyVal) {
		t.Errorf("Expected: %s , received: %s", ToJSON(eVal), ToJSON(rplyVal))
	}

	path = []string{"FirstLevel", "SecondLevel[1]", "ThirdLevel2", "Fld2[2]"}
	eVal = "Val3"
	if rplyVal, err := nM.FieldAsInterface(path); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eVal, rplyVal) {
		t.Errorf("Expected: %s , received: %s", ToJSON(eVal), ToJSON(rplyVal))
	}
}

func TestNavMapGetKeys(t *testing.T) {
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

func TestMapRemove(t *testing.T) {
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

func TestGetPathFromValue(t *testing.T) {
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
		/*{
			name: "map",
			args: args{reflect.ValueOf(map[string]string{"test": "test"}), "test"},
			exp: []string{"testtest"},
		},*/
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

func TestRemoteHost(t *testing.T) {
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

func TestNavMapSet(t *testing.T) {
	tests := []struct{
		name string
		path []string
		val any
		err error
	}{
		{
			name: "empty path",
			path: []string{},
			val: 1,
			err: ErrWrongPath,
		},
		{
			name: "non supported data type",
			path: []string{"Test", "test"},
			val: 1,
			err: ErrWrongPath,
		},
		{
			name: "non supported data type",
			path: []string{"MP", "test"},
			val: 1,
			err: nil,
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
