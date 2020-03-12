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
	"sort"
	"strings"
	"testing"
)

func TestNavMapGetFieldAsString(t *testing.T) {
	nM := NavigableMap{
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

type myEv map[string]interface{}

func (ev myEv) AsNavigableMap() (NavigableMap, error) {
	return NavigableMap(ev), nil
}

func TestNavMapAsNavigableMap(t *testing.T) {
	myData := myEv{
		"FirstLevel": map[string]interface{}{
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
		},
		"Field4": &testStruct{
			Item1: "Ten",
			Item2: 10,
		},
	}

	eNavMap := NavigableMap{
		"FirstLevel": map[string]interface{}{
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
		},
		"Field4": &testStruct{
			Item1: "Ten",
			Item2: 10,
		},
	}

	if rcv, err := myData.AsNavigableMap(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eNavMap, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", eNavMap, rcv)
	}
}

/*
func TestNavMapAdd(t *testing.T) {
	nM := NewOrderedNavigableMap()
	path := []string{"FistLever2", "SecondLevel2", "Field2"}
	data := "Value2"
	nM.Set(path, data)
	path = []string{"FirstLevel", "SecondLevel", "ThirdLevel", "Fld1"}
	data = "Val1"
	nM.Set(path, data)
	path = []string{"FistLever2", "Field3"}
	data = "Value3"
	nM.Set(path, data)
	path = []string{"Field4"}
	data = "Val4"
	nM.Set(path, data)
	eNavMap := NavigableMap{
		"FirstLevel": NavigableMap{
			"SecondLevel": NavigableMap{
				"ThirdLevel": NavigableMap{
					"Fld1": "Val1",
				},
			},
		},
		"FistLever2": NavigableMap{
			"SecondLevel2": NavigableMap{
				"Field2": "Value2",
			},
			"Field3": "Value3",
		},
		"Field4": "Val4",
	}
	if !reflect.DeepEqual(nM.nm, eNavMap) {
		t.Errorf("Expecting: %+v, received: %+v", eNavMap, nM.nm)
	}
	eOrder := [][]string{
		[]string{"FistLever2", "SecondLevel2", "Field2"},
		[]string{"FirstLevel", "SecondLevel", "ThirdLevel", "Fld1"},
		[]string{"FistLever2", "Field3"},
		[]string{"Field4"},
	}
	if !reflect.DeepEqual(eOrder, nM.order) {
		t.Errorf("Expecting: %+v, received: %+v", eOrder, nM.order)
	}

}*/

type testStruct struct {
	Item1 string
	Item2 int
}

func TestNavMapAdd2(t *testing.T) {
	nM := NavigableMap{}
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
	eNavMap := NavigableMap{
		"FirstLevel": NavigableMap{
			"SecondLevel": NavigableMap{
				"ThirdLevel": NavigableMap{
					"Fld1": 123.123,
				},
			},
		},
		"FistLever2": NavigableMap{
			"SecondLevel2": NavigableMap{
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

/*
func TestNavMapItems(t *testing.T) {
	nM := NewOrderedNavigableMap()
	if err := nM.Set([]string{"FirstLevel", "SecondLevel", "ThirdLevel", "Fld1"}, "Val1"); err != nil {
		t.Error(err)
	}
	if err := nM.Set([]string{"FistLever2", "SecondLevel2", "Field2"}, "Value2"); err != nil {
		t.Error(err)
	}
	if err := nM.Set([]string{"FistLever2", "Field3"}, "Value3"); err != nil {
		t.Error(err)
	}
	if err := nM.Set([]string{"Field4"}, "Val4"); err != nil {
		t.Error(err)
	}
	eItems := []interface{}{"Val1", "Value2", "Value3", "Val4"}
	if vals := nM.Values(); len(vals) != len(eItems) {
		t.Errorf("Expecting: %+v, received: %+v",
			ToJSON(eItems), ToJSON(vals))
	}
}

func TestNavMapItems2(t *testing.T) {
	nM := NewOrderedNavigableMap()
	if err := nM.Set([]string{"FirstLevel", "SecondLevel", "ThirdLevel", "Fld1"}, 123.123); err != nil {
		t.Error(err)
	}
	if err := nM.Set([]string{"FistLever2", "SecondLevel2", "Field2"}, 123); err != nil {
		t.Error(err)
	}
	if err := nM.Set([]string{"FistLever2", "Field3"}, "Value3"); err != nil {
		t.Error(err)
	}
	if err := nM.Set([]string{"Field4"}, &testStruct{
		Item1: "Ten",
		Item2: 10,
	}); err != nil {
		t.Error(err)
	}
	eItems := []interface{}{123.123, 123, "Value3", &testStruct{
		Item1: "Ten",
		Item2: 10,
	}}
	if vals := nM.Values(); len(vals) != len(eItems) {
		t.Errorf("Expecting: %+v, received: %+v",
			ToJSON(eItems), ToJSON(vals))
	}
}

func TestNavMapOrder(t *testing.T) {
	myData := NavigableMap{
		"FirstLevel": NavigableMap{
			"SecondLevel": NavigableMap{
				"ThirdLevel": NavigableMap{
					"Fld1": "Val1",
				},
			},
		},
		"FistLever2": NavigableMap{
			"SecondLevel2": NavigableMap{
				"Field2": "Value2",
			},
			"Field3": "Value3",
		},
		"Field4": "Val4",
	}
	order := [][]string{
		[]string{"FirstLevel", "SecondLevel", "ThirdLevel", "Fld1"},
		[]string{"FistLever2", "SecondLevel2", "Field2"},
		[]string{"FistLever2", "Field3"},
		[]string{"Field4"},
	}
	nM := &OrderedNavigableMap{
		nm:    myData,
		order: order,
	}
	eItems := []interface{}{"Val1", "Value2", "Value3", "Val4"}
	if vals := nM.Values(); !reflect.DeepEqual(vals, eItems) {
		t.Errorf("Expecting: %+v, received: %+v",
			ToJSON(eItems), ToJSON(vals))
	}
}

func TestNavMapOrder2(t *testing.T) {
	myData := NavigableMap{
		"FirstLevel": NavigableMap{
			"SecondLevel": NavigableMap{
				"ThirdLevel": NavigableMap{
					"Fld1": "Val1",
				},
			},
		},
		"FistLever2": NavigableMap{
			"SecondLevel2": NavigableMap{
				"Field2": "Value2",
			},
			"Field3": "Value3",
		},
		"Field4": "Val4",
	}
	order := [][]string{
		[]string{"FistLever2", "SecondLevel2", "Field2"},
		[]string{"Field4"},
		[]string{"FistLever2", "Field3"},
		[]string{"FirstLevel", "SecondLevel", "ThirdLevel", "Fld1"},
	}
	nM := &OrderedNavigableMap{
		nm:    myData,
		order: order,
	}
	eItems := []interface{}{"Value2", "Val4", "Value3", "Val1"}
	if vals := nM.Values(); !reflect.DeepEqual(eItems, vals) {
		t.Errorf("Expecting: %+v, received: %+v",
			ToJSON(eItems), ToJSON(vals))
	}
}
*/
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
	nM := NavigableMap(myData)
	eStr := ToJSON(myData)
	if !reflect.DeepEqual(nM.String(), eStr) {
		t.Errorf("Expecting: %+v, received: %+v", eStr, nM.String())
	}
}

func TestNavMapGetField(t *testing.T) {
	nM := NavigableMap{
		"FirstLevel": map[string]interface{}{
			"SecondLevel": map[string]interface{}{
				"ThirdLevel": map[string]interface{}{
					"Fld1": []interface{}{"Val1", "Val2"},
				},
			},
		},
		"FirstLevel2": map[string]interface{}{
			"SecondLevel2": []map[string]interface{}{
				map[string]interface{}{
					"ThirdLevel2": map[string]interface{}{
						"Fld1": "Val1",
					},
				},
				map[string]interface{}{
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
	nM := NavigableMap{
		"FirstLevel": map[string]interface{}{
			"SecondLevel": []map[string]interface{}{
				map[string]interface{}{
					"ThirdLevel": map[string]interface{}{
						"Fld1": "Val1",
					},
				},
				map[string]interface{}{
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
	navMp := NavigableMap{
		"FirstLevel": map[string]interface{}{
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
		"FirstLevel",
		"FirstLevel.SecondLevel",
		"FirstLevel.SecondLevel.ThirdLevel",
		"FirstLevel.SecondLevel.ThirdLevel.Fld1",
		"FistLever2",
		"FistLever2.SecondLevel2",
		"FistLever2.SecondLevel2.Field2",
		"FistLever2.Field3",
		"FistLever2.Field4",
		"FistLever2.Field4.Item1",
		"FistLever2.Field4.Item2",
		"Field5",
		"Field5.Item1",
		"Field5.Item2",
		"Field6",
		"Field6[0]",
		"Field6[1]",
	}
	keys := navMp.GetKeys(true)
	sort.Strings(expKeys)
	sort.Strings(keys)
	if !reflect.DeepEqual(expKeys, keys) {
		t.Errorf("Expecting: %+v, received: %+v", ToJSON(expKeys), ToJSON(keys))
	}
}
