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

func TestMapStorageRemote(t *testing.T) {
	nm := MapStorage{}
	eOut := LocalAddr()
	if rcv := nm.RemoteHost(); !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", eOut, rcv)
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
		"FirstLevel": dataStorage(MapStorage{
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
		"Field8":  []dataStorage{MapStorage{"A": 1}},
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
