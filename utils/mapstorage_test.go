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
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"
)

func TestNavMapGetFieldAsString(t *testing.T) {
	nM := MapStorage{
		"FirstLevel": map[string]interface{}{
			"SecondLevel": map[string]interface{}{
				"ThirdLevel": map[string]interface{}{
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

type testStruct struct {
	Item1 string
	Item2 int
}

func TestNavMapAdd2(t *testing.T) {
	nM := MapStorage{}
	path := []string{"FistLever2", "SecondLevel2", "Field2"}
	data := 123
	if err := nM.Set(path, data); err != nil {
		t.Error(err)
	}
	path = []string{"FirstLevel", "SecondLevel", "ThirdLevel", "Fld1"}
	data1 := 123.123
	if err := nM.Set(path, data1); err != nil {
		t.Error(err)
	}
	path = []string{"FistLever2", "Field3"}
	data2 := "Value3"
	if err := nM.Set(path, data2); err != nil {
		t.Error(err)
	}
	path = []string{"Field4"}
	data3 := &testStruct{
		Item1: "Ten",
		Item2: 10,
	}
	if err := nM.Set(path, data3); err != nil {
		t.Error(err)
	}
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

	if err := nM.Set([]string{}, nil); err != ErrWrongPath {
		t.Errorf("Expected error: %s received: %v", ErrWrongPath, err)
	}

	if err := nM.Set([]string{"Field4", "Field2"}, nil); err != ErrWrongPath {
		t.Errorf("Expected error: %s received: %v", ErrWrongPath, err)
	}

	nM = MapStorage{"Field1": map[string]interface{}{}}
	path = []string{"Field1", "SecondLevel2", "Field2"}
	data = 123
	if err := nM.Set(path, data); err != nil {
		t.Error(err)
	}

	eNavMap = MapStorage{"Field1": map[string]interface{}{
		"SecondLevel2": MapStorage{
			"Field2": 123,
		},
	}}
	if !reflect.DeepEqual(nM, eNavMap) {
		t.Errorf("Expecting: %+v, received: %+v", eNavMap, nM)
	}
}

func TestCloneMapStorage(t *testing.T) {
	expected := MapStorage{
		"FirstLevel": map[string]interface{}{
			"SecondLevel": map[string]interface{}{
				"ThirdLevel": map[string]interface{}{
					"Fld1": "Val1",
				},
			},
		},
		"FistLever2": map[string]interface{}{
			"SecondLevel2": map[string]interface{}{
				"Field2": "Value2",
			},
			"Field3": "Value3",
		},
		"Field4": "Val4",
	}
	if received := expected.Clone(); !reflect.DeepEqual(received, expected) {
		t.Errorf("Expected %+v \n, received %+v", ToJSON(expected), ToJSON(received))
	}
}

func TestNavMapString(t *testing.T) {
	myData := map[string]interface{}{
		"FirstLevel": map[string]interface{}{
			"SecondLevel": map[string]interface{}{
				"ThirdLevel": map[string]interface{}{
					"Fld1": "Val1",
				},
			},
		},
		"FistLever2": map[string]interface{}{
			"SecondLevel2": map[string]interface{}{
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
		"FirstLevel": map[string]interface{}{
			"SecondLevel": map[string]interface{}{
				"ThirdLevel": map[string]interface{}{
					"Fld1": []interface{}{"Val1", "Val2"},
				},
			},
		},
		"FirstLevel2": map[string]interface{}{
			"SecondLevel2": []map[string]interface{}{
				{
					"ThirdLevel2": map[string]interface{}{
						"Fld1": "Val1",
					},
				},
				{
					"Count": 10,
					"ThirdLevel2": map[string]interface{}{
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
	eFld2 := map[string]interface{}{"Fld1": "Val1"}
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
		"FirstLevel": map[string]interface{}{
			"SecondLevel": []map[string]interface{}{
				{
					"ThirdLevel": map[string]interface{}{
						"Fld1": "Val1",
					},
				},
				{
					"Count": 10,
					"ThirdLevel2": map[string]interface{}{
						"Fld2": []string{"Val1", "Val2", "Val3"},
					},
				},
			},
		},
		"AnotherFirstLevel": "ValAnotherFirstLevel",
	}

	path := []string{"FirstLevel", "SecondLevel[0]", "Count"}
	expErr := ErrNotFound
	var eVal interface{} = nil
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
	navMp := MapStorage{
		"FirstLevel": MapStorage{
			"SecondLevel": map[string]interface{}{
				"ThirdLevel": map[string]interface{}{
					"Fld1": 123.123,
				},
			},
		},
		"FistLever2": map[string]interface{}{
			"SecondLevel2": map[string]interface{}{
				"Field2": 123,
			},
			"Field3": "Value3",
			"Field4": &testStruct{
				Item1: "Ten",
				Item2: 10,
			},
		},
		"Field5": &testStruct{
			Item1: "Ten",
			Item2: 10,
		},
		"Field6": []string{"1", "2"},
	}
	expKeys := []string{
		"FirstLevel.SecondLevel.ThirdLevel.Fld1",
		"FistLever2.SecondLevel2.Field2",
		"FistLever2.Field3",
		"FistLever2.Field4.Item1",
		"FistLever2.Field4.Item2",
		"Field5.Item1",
		"Field5.Item2",
		"Field6[0]",
		"Field6[1]",
	}
	keys := navMp.GetKeys(true, 0, EmptyString)
	sort.Strings(expKeys)
	sort.Strings(keys)
	if !reflect.DeepEqual(expKeys, keys) {
		t.Errorf("Expecting: %+v, received: %+v", ToJSON(expKeys), ToJSON(keys))
	}

	expKeys = []string{
		"*req.FirstLevel",
		"*req.FistLever2",
		"*req.Field5",
		"*req.Field6",
	}
	keys = navMp.GetKeys(false, 1, MetaReq)
	sort.Strings(expKeys)
	sort.Strings(keys)
	if !reflect.DeepEqual(expKeys, keys) {
		t.Errorf("Expecting: %+v, received: %+v", ToJSON(expKeys), ToJSON(keys))
	}

	expKeys = []string{
		"FirstLevel.SecondLevel",
		"FistLever2.SecondLevel2",
		"FistLever2.Field3",
		"FistLever2.Field4",
		"Field5",
		"Field6",
	}
	keys = navMp.GetKeys(false, 2, EmptyString)
	sort.Strings(expKeys)
	sort.Strings(keys)
	if !reflect.DeepEqual(expKeys, keys) {
		t.Errorf("Expecting: %+v, received: %+v", ToJSON(expKeys), ToJSON(keys))
	}
}

func TestNavMapFieldAsInterface2(t *testing.T) {
	nM := MapStorage{
		"AnotherFirstLevel": "ValAnotherFirstLevel",
		"Slice":             &[]struct{}{{}},
		"SliceString":       []string{"1", "2"},
		"SliceInterface":    []interface{}{1, "2"},
	}

	path := []string{"Slice[1]"}
	expErr := ErrNotFound
	var eVal interface{} = nil
	if _, err := nM.FieldAsInterface(path); err != nil && err.Error() != expErr.Error() {
		t.Errorf("Expected error: %s, received error: %v", expErr.Error(), err)
	}

	path = []string{"Slice[nan]"}
	expErr = fmt.Errorf(`strconv.Atoi: parsing "nan": invalid syntax`)
	if _, err := nM.FieldAsInterface(path); err != nil && err.Error() != expErr.Error() {
		t.Errorf("Expected error: %s, received error: %v", expErr.Error(), err)
	}

	path = []string{"Slice[0]"}
	eVal = struct{}{}
	if rplyVal, err := nM.FieldAsInterface(path); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eVal, rplyVal) {
		t.Errorf("Expected: %s , received: %s", ToJSON(eVal), ToJSON(rplyVal))
	}

	path = []string{"AnotherFirstLevel[1]"}
	expErr = ErrNotFound
	if _, err := nM.FieldAsInterface(path); err != nil && err.Error() != expErr.Error() {
		t.Errorf("Expected error: %s, received error: %v", expErr.Error(), err)
	}

	path = []string{"SliceString[nan]"}
	expErr = fmt.Errorf(`strconv.Atoi: parsing "nan": invalid syntax`)
	if _, err := nM.FieldAsInterface(path); err != nil && err.Error() != expErr.Error() {
		t.Errorf("Expected error: %s, received error: %v", expErr.Error(), err)
	}

	path = []string{"SliceInterface[nan]"}
	expErr = fmt.Errorf(`strconv.Atoi: parsing "nan": invalid syntax`)
	if _, err := nM.FieldAsInterface(path); err != nil && err.Error() != expErr.Error() {
		t.Errorf("Expected error: %s, received error: %v", expErr.Error(), err)
	}

	path = []string{"SliceString[4]"}
	expErr = ErrNotFound
	if _, err := nM.FieldAsInterface(path); err != nil && err.Error() != expErr.Error() {
		t.Errorf("Expected error: %s, received error: %v", expErr.Error(), err)
	}

	path = []string{"SliceInterface[4]"}
	expErr = ErrNotFound
	if _, err := nM.FieldAsInterface(path); err != nil && err.Error() != expErr.Error() {
		t.Errorf("Expected error: %s, received error: %v", expErr.Error(), err)
	}

	path = nil
	expErr = errors.New("empty field path")
	if _, err := nM.FieldAsInterface(path); err != nil && err.Error() != expErr.Error() {
		t.Errorf("Expected error: %s, received error: %v", expErr.Error(), err)
	}
}

func TestNavMapGetField2(t *testing.T) {
	nM := MapStorage{
		"FirstLevel": MapStorage{
			"SecondLevel": MapStorage{
				"ThirdLevel": MapStorage{
					"Fld1": []interface{}{"Val1", "Val2"},
				},
			},
		},
		"FirstLevel2": MapStorage{
			"SecondLevel2": []MapStorage{
				{
					"ThirdLevel2": MapStorage{
						"Fld1": "Val1",
					},
				},
				{
					"Count": 10,
					"ThirdLevel2": MapStorage{
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
	eFld2 := MapStorage{"Fld1": "Val1"}
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

func TestNavMapRemove(t *testing.T) {
	nM := MapStorage{
		"Field4": &testStruct{
			Item1: "Ten",
			Item2: 10,
		},
	}

	if err := nM.Remove([]string{}); err != ErrWrongPath {
		t.Errorf("Expected error: %s received: %v", ErrWrongPath, err)
	}

	if err := nM.Remove([]string{"Field4", "Field2"}); err != ErrWrongPath {
		t.Errorf("Expected error: %s received: %v", ErrWrongPath, err)
	}
	nM = MapStorage{
		"Field1": map[string]interface{}{
			"SecondLevel2": 1,
		},
	}

	path := []string{"Field1", "SecondLevel2"}
	if err := nM.Remove(path); err != nil {
		t.Error(err)
	}
	eNavMap := MapStorage{"Field1": map[string]interface{}{}}
	if !reflect.DeepEqual(nM, eNavMap) {
		t.Errorf("Expecting: %+v, received: %+v", eNavMap, nM)
	}
	nM = MapStorage{"Field1": MapStorage{
		"SecondLevel2": 1,
	}}
	path = []string{"Field1", "SecondLevel2"}
	if err := nM.Remove(path); err != nil {
		t.Error(err)
	}
	eNavMap = MapStorage{"Field1": MapStorage{}}
	if !reflect.DeepEqual(nM, eNavMap) {
		t.Errorf("Expecting: %+v, received: %+v", eNavMap, nM)
	}
	path = []string{"Field1", "SecondLevel2"}
	if err := nM.Remove(path); err != nil {
		t.Error(err)
	}
	eNavMap = MapStorage{"Field1": MapStorage{}}
	if !reflect.DeepEqual(nM, eNavMap) {
		t.Errorf("Expecting: %+v, received: %+v", eNavMap, nM)
	}
}

func TestNavMapFieldAsInterface3(t *testing.T) {
	nM := MapStorage{
		"AnotherFirstLevel": "ValAnotherFirstLevel",
		"Slice":             []MapStorage{{}},
		"Slice2":            []DataProvider{MapStorage{}},
		"SliceString":       []map[string]interface{}{{}},
		"SliceInterface":    []interface{}{MapStorage{"A": 0}, map[string]interface{}{"B": 1}},
	}

	path := []string{"Slice[1]", "A"}
	expErr := ErrNotFound
	if _, err := nM.FieldAsInterface(path); err != nil && err.Error() != expErr.Error() {
		t.Errorf("Expected error: %s, received error: %v", expErr.Error(), err)
	}

	path = []string{"Slice[nan]", "A"}
	expErr = fmt.Errorf(`strconv.Atoi: parsing "nan": invalid syntax`)
	if _, err := nM.FieldAsInterface(path); err != nil && err.Error() != expErr.Error() {
		t.Errorf("Expected error: %s, received error: %v", expErr.Error(), err)
	}

	path = []string{"AnotherFirstLevel[1]", "N"}
	expErr = ErrNotFound
	if _, err := nM.FieldAsInterface(path); err != nil && err.Error() != expErr.Error() {
		t.Errorf("Expected error: %s, received error: %v", expErr.Error(), err)
	}

	path = []string{"SliceString[1]", "A"}
	if _, err := nM.FieldAsInterface(path); err != nil && err.Error() != expErr.Error() {
		t.Errorf("Expected error: %s, received error: %v", expErr.Error(), err)
	}

	path = []string{"SliceString[nan]", "A"}
	expErr = fmt.Errorf(`strconv.Atoi: parsing "nan": invalid syntax`)
	if _, err := nM.FieldAsInterface(path); err != nil && err.Error() != expErr.Error() {
		t.Errorf("Expected error: %s, received error: %v", expErr.Error(), err)
	}
	path = []string{"SliceInterface[nan]", "A"}
	expErr = fmt.Errorf(`strconv.Atoi: parsing "nan": invalid syntax`)
	if _, err := nM.FieldAsInterface(path); err != nil && err.Error() != expErr.Error() {
		t.Errorf("Expected error: %s, received error: %v", expErr.Error(), err)
	}

	path = []string{"SliceInterface[4]", "A"}
	expErr = ErrNotFound
	if _, err := nM.FieldAsInterface(path); err != nil && err.Error() != expErr.Error() {
		t.Errorf("Expected error: %s, received error: %v", expErr.Error(), err)
	}

	path = []string{"SliceInterface[0]", "A"}
	expErr = fmt.Errorf(`strconv.Atoi: parsing "nan": invalid syntax`)
	var eVal interface{} = 0
	if rplyVal, err := nM.FieldAsInterface(path); err != nil && err.Error() != expErr.Error() {
		t.Errorf("Expected error: %s, received error: %v", expErr.Error(), err)
	} else if !reflect.DeepEqual(eVal, rplyVal) {
		t.Errorf("Expected: %s , received: %s", ToJSON(eVal), ToJSON(rplyVal))
	}

	path = []string{"SliceInterface[1]", "B"}
	expErr = ErrNotFound
	eVal = 1
	if rplyVal, err := nM.FieldAsInterface(path); err != nil && err.Error() != expErr.Error() {
		t.Errorf("Expected error: %s, received error: %v", expErr.Error(), err)
	} else if !reflect.DeepEqual(eVal, rplyVal) {
		t.Errorf("Expected: %s , received: %s", ToJSON(eVal), ToJSON(rplyVal))
	}

	path = []string{"Slice2[1]", "A"}
	expErr = ErrNotFound
	if _, err := nM.FieldAsInterface(path); err != nil && err.Error() != expErr.Error() {
		t.Errorf("Expected error: %s, received error: %v", expErr.Error(), err)
	}

	path = []string{"Slice2[nan]", "A"}
	expErr = fmt.Errorf(`strconv.Atoi: parsing "nan": invalid syntax`)
	if _, err := nM.FieldAsInterface(path); err != nil && err.Error() != expErr.Error() {
		t.Errorf("Expected error: %s, received error: %v", expErr.Error(), err)
	}

	path = []string{"Slice2[0]", "A"}
	expErr = ErrNotFound
	if _, err := nM.FieldAsInterface(path); err != nil && err.Error() != expErr.Error() {
		t.Errorf("Expected error: %s, received error: %v", expErr.Error(), err)
	}
}

func TestNavMapGetKeys2(t *testing.T) {
	navMp := MapStorage{
		"FirstLevel": DataStorage(MapStorage{
			"SecondLevel": map[string]interface{}{
				"ThirdLevel": MapStorage{
					"Fld1": 123.123,
				},
			},
		}),
		"FistLever2": MapStorage{
			"SecondLevel2": map[string]interface{}{
				"Field2": 123,
			},
			"Field3": "Value3",
			"Field4": &testStruct{
				Item1: "Ten",
				Item2: 10,
			},
		},
		"Field5": &testStruct{
			Item1: "Ten",
			Item2: 10,
		},
		"Field6":  []string{"1", "2"},
		"Field7":  []interface{}{"1", "2"},
		"Field8":  []DataStorage{MapStorage{"A": 1}},
		"Field9":  []MapStorage{{"A": 1}},
		"Field10": []map[string]interface{}{{"A": 1}},
	}
	expKeys := []string{
		"FirstLevel.SecondLevel.ThirdLevel.Fld1",
		"FistLever2.SecondLevel2.Field2",
		"FistLever2.Field3",
		"FistLever2.Field4.Item1",
		"FistLever2.Field4.Item2",
		"Field5.Item1",
		"Field5.Item2",
		"Field6[0]",
		"Field6[1]",
		"Field7[0]",
		"Field7[1]",
		"Field8[0].A",
		"Field9[0].A",
		"Field10[0].A",
	}
	keys := navMp.GetKeys(true, 0, EmptyString)
	sort.Strings(expKeys)
	sort.Strings(keys)
	if !reflect.DeepEqual(expKeys, keys) {
		t.Errorf("Expecting: %+v, received: %+v", ToJSON(expKeys), ToJSON(keys))
	}
}

func TestMapStorageCloneNil(t *testing.T) {
	var test MapStorage
	if !reflect.DeepEqual(test, test.Clone()) {
		t.Errorf("Expecting: <nil>, received: %+v", test.Clone())
	}
}

func TestNavMapGetFieldAsMapStringInterfaceError(t *testing.T) {
	nM := MapStorage{
		"AnotherFirstLevel": "ValAnotherFirstLevel",
		"Slice":             &[]struct{}{{}},
		"SliceString":       []string{"1", "2"},
		"SliceInterface":    map[string]interface{}{},
	}
	path := []string{"SliceInterface[4]"}
	_, err := nM.FieldAsInterface(path)
	if err == nil || err.Error() != "NOT_FOUND" {
		t.Errorf("Expecting: <NOT_FOUND>, received: %+v", err)
	}

}

func TestNavMapGetFieldAsMapStringInterface(t *testing.T) {
	nM := MapStorage{
		"FIELD": map[string]interface{}{
			"Field1": "Val1",
			"Field2": "Val2"},
	}
	path := []string{"FIELD[Field2]"}
	if result, err := nM.FieldAsInterface(path); err != nil {
		t.Errorf("Expecting: <nil>, received: %+v", err)
	} else if !reflect.DeepEqual(result, "Val2") {
		t.Errorf("Expecting: <Val2>, received: %+v", result)
	}

}

func TestNavMapGetFieldAsDataProvider(t *testing.T) {
	nM := MapStorage{
		"FIELD": MapStorage{
			"Field1": "Val1",
			"Field2": "Val2"},
	}
	path := []string{"FIELD[Field2]"}
	if result, err := nM.FieldAsInterface(path); err != nil {
		t.Errorf("Expecting: <nil>, received: %+v", err)
	} else if !reflect.DeepEqual(result, "Val2") {
		t.Errorf("Expecting: <Val2>, received: %+v", result)
	}
}

func TestMSGetKeys(t *testing.T) {
	m := MapStorage{
		"testKey1": false,
		"testKey2": true,
		"testKey3": false,
	}

	expected := []string{"testKey1", "testKey2", "testKey3"}
	received := m.GetKeys(false, 2, EmptyString)
	sort.Strings(received)

	if !reflect.DeepEqual(received, expected) {
		t.Errorf(
			"\nCase bool: \nReceived: <%+v>, \nExpected: <%+v>",
			received,
			expected,
		)
	}

	m = MapStorage{
		"testKey1": 1,
		"testKey2": 3,
		"testKey3": 2,
	}

	received = m.GetKeys(false, 2, EmptyString)
	sort.Strings(received)

	if !reflect.DeepEqual(received, expected) {
		t.Errorf(
			"\nCase Int: \nReceived: <%+v>, \nExpected: <%+v>",
			received,
			expected,
		)
	}

	m = MapStorage{
		"testKey1": []uint8{1},
		"testKey2": []uint8{3},
		"testKey3": []uint8{2},
	}

	received = m.GetKeys(false, 2, EmptyString)
	sort.Strings(received)

	if !reflect.DeepEqual(received, expected) {
		t.Errorf(
			"\nCase []uint8: \nReceived: <%+v>, \nExpected: <%+v>",
			received,
			expected,
		)
	}
	m = MapStorage{
		"testKey1": "testString1",
		"testKey2": "testString2",
		"testKey3": "testString3",
	}

	received = m.GetKeys(false, 2, EmptyString)
	sort.Strings(received)

	if !reflect.DeepEqual(received, expected) {
		t.Errorf(
			"\nCase string: \nReceived: <%+v>, \nExpected: <%+v>",
			received,
			expected,
		)
	}
	m = MapStorage{
		"testKey1": 2 * time.Second,
		"testKey2": time.Second,
		"testKey3": 3 * time.Second,
	}

	received = m.GetKeys(false, 2, EmptyString)
	sort.Strings(received)

	if !reflect.DeepEqual(received, expected) {
		t.Errorf(
			"\nCase time: \nReceived: <%+v>, \nExpected: <%+v>",
			received,
			expected,
		)
	}
	m = MapStorage{
		"testKey1": 3.0,
		"testKey2": 1.0,
		"testKey3": 2.0,
	}

	received = m.GetKeys(false, 2, EmptyString)
	sort.Strings(received)

	if !reflect.DeepEqual(received, expected) {
		t.Errorf(
			"\nCase float: \nReceived: <%+v>, \nExpected: <%+v>",
			received,
			expected,
		)
	}

	m = MapStorage{
		"testKey1": nil,
		"testKey2": nil,
		"testKey3": nil,
	}

	received = m.GetKeys(false, 2, EmptyString)
	sort.Strings(received)

	if !reflect.DeepEqual(received, expected) {
		t.Errorf(
			"\nCase nil: \nReceived: <%+v>, \nExpected: <%+v>",
			received,
			expected,
		)
	}
}

func TestMSgetPathFromValue(t *testing.T) {
	var v reflect.Value
	pref := "testPrefix."
	sl := []int{1, 2}
	v = reflect.ValueOf(sl)

	expected := []string{pref[:len(pref)-1] + "[0]", pref[:len(pref)-1] + "[1]"}
	received := getPathFromValue(v, pref)

	if !reflect.DeepEqual(expected, received) {
		t.Errorf(
			"\nCase reflect.Array: \nReceived: <%+v>, \nExpected: <%+v>",
			received,
			expected,
		)
	}

	arr := [2]int{1, 2}
	v = reflect.ValueOf(arr)

	received = getPathFromValue(v, pref)

	if !reflect.DeepEqual(expected, received) {
		t.Errorf(
			"\nCase reflect.Array: \nReceived: <%+v>, \nExpected: <%+v>",
			received,
			expected,
		)
	}

	m1 := map[string]int{
		"testKey1": 1,
		"testKey2": 2,
		"testKey3": 3,
	}
	v = reflect.ValueOf(m1)

	expected = []string{pref + "testKey1", pref + "testKey2", pref + "testKey3"}
	received = getPathFromValue(v, pref)
	sort.Strings(received)

	if !reflect.DeepEqual(expected, received) {
		t.Errorf(
			"\nCase string key values: \nReceived: %v, \nExpected: %v",
			received,
			expected,
		)
	}

	m2 := map[int]bool{
		1: false,
		2: true,
	}
	v = reflect.ValueOf(m2)
	expected = []string{pref + "<int Value>", pref + "<int Value>"}
	received = getPathFromValue(v, pref)
	if !reflect.DeepEqual(received, expected) {
		t.Errorf(
			"\nCase non-string key values: \nReceived: %v, \nExpected: %v",
			received,
			expected,
		)
	}
}

func TestSecureMapStorageString(t *testing.T) {
	sm := &SecureMapStorage{ms: MapStorage{
		"field1": 23,
		"field2": []string{"ms1", "ms2"},
	}}
	exp := "{\"field1\":23,\"field2\":[\"ms1\",\"ms2\"]}"
	if val := sm.String(); val != exp {
		t.Errorf("expected %+s,received %+s", exp, val)
	}
}

func TestSecureMapStorageFieldAsInterface(t *testing.T) {
	sm := &SecureMapStorage{ms: MapStorage{
		"field1": []string{"val1", "val2"},
	}}
	if val, err := sm.FieldAsInterface([]string{"field1[0]"}); err != nil {

		t.Error(err)
	} else if val.(string) != sm.ms["field1"].([]string)[0] {
		t.Errorf("expected %s,received %v", val, sm.ms["field1"].([]string)[0])
	}
}

func TestSecureMapStorageFieldAsString(t *testing.T) {
	sm := &SecureMapStorage{ms: MapStorage{
		"field1": []string{"val1", "val2"},
	}}
	if val, err := sm.FieldAsString([]string{"field1[0]"}); err != nil {
		t.Error(err)
	} else if val != sm.ms["field1"].([]string)[0] {
		t.Errorf("expected %s,received %v", val, sm.ms["field1"].([]string)[0])
	}

}
func TestSecureMapStorageSet(t *testing.T) {
	sm := &SecureMapStorage{ms: MapStorage{
		"field": map[string]interface{}{},
	}}
	exp := MapStorage{
		"test": "val",
	}
	if err := sm.Set([]string{"field2"}, MapStorage{"test": "val"}); err != nil {
		t.Error(err)
	}
	if v, has := sm.ms["field2"]; !has {
		t.Error("expected")
	} else if !reflect.DeepEqual(v, exp) {
		t.Errorf("expected %v,reeived %v", ToJSON(exp), ToJSON(v))
	}

}

func TestSecureMapStorageGetKeys(t *testing.T) {
	sm := &SecureMapStorage{
		ms: MapStorage{
			"field1": map[string]interface{}{
				"subfield1": []string{"subkey"},
			},
			"field2": MapStorage{
				"mapfield": map[string]interface{}{
					"submapfield2": []string{"subkey"},
				},
			},
		},
	}
	exp := []string{"*pre.field1.subfield1[0]", "*pre.field2.mapfield.submapfield2[0]"}
	val := sm.GetKeys(true, 3, "*pre")
	sort.Slice(val, func(i, j int) bool {
		return val[i] < val[j]
	})
	if !reflect.DeepEqual(val, exp) {
		t.Errorf("expected %+v,received %+v", ToJSON(exp), ToJSON(val))
	}
}

func TestSecureMapStorageRemove(t *testing.T) {
	sm := &SecureMapStorage{
		ms: MapStorage{
			"field": map[string]interface{}{
				"subfield": uint16(3),
			},
		},
	}
	if err := sm.Remove([]string{"field", "subfield"}); err != nil {
		t.Error(err)
	}
	if _, has := sm.ms["field"].(map[string]interface{})["subfield"]; has {
		t.Error("should been removed")
	}
}

func TestSecureMapStorageClone(t *testing.T) {
	sm := &SecureMapStorage{
		ms: MapStorage{
			"field1": []string{"val1", "val2"},
			"field2": "val",
		},
	}
	if val := sm.Clone(); !reflect.DeepEqual(val, sm) {
		t.Errorf("expected %+v,received %+v", ToJSON(sm.ms), ToJSON(val.ms))
	}
}

func TestNewSecureMapStorage(t *testing.T) {
	if sm := NewSecureMapStorage(); reflect.DeepEqual(sm, nil) {
		t.Error("should receive new secure map")
	}
}
