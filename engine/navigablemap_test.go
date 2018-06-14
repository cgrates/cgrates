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
package engine

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/cgrates/cgrates/utils"
)

func TestNavMapGetFieldAsString(t *testing.T) {
	nM := &NavigableMap{
		data: map[string]interface{}{
			"FirstLevel": map[string]interface{}{
				"SecondLevel": map[string]interface{}{
					"ThirdLevel": map[string]interface{}{
						"Fld1": "Val1",
					},
				},
			},
			"AnotherFirstLevel": "ValAnotherFirstLevel",
		},
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
		errors.New("no map at path: <NonExisting>").Error() {
		t.Error(err)
	}
}

type myEv map[string]interface{}

func (ev myEv) AsNavigableMap() (*NavigableMap, error) {
	return NewNavigableMap(ev), nil
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
		data: map[string]interface{}{
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
		},
	}

	if rcv, err := myData.AsNavigableMap(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eNavMap.data, rcv.data) {
		t.Errorf("Expecting: %+v, received: %+v", eNavMap.data, rcv.data)
	}
}

func TestNavMapNewNavigableMap(t *testing.T) {
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

	eNavMap := NavigableMap{
		data: myData,
	}

	nM := NewNavigableMap(myData)
	if !reflect.DeepEqual(nM.data, eNavMap.data) {
		t.Errorf("Expecting: %+v, received: %+v", eNavMap.data, nM.data)
	}
}

func TestNavMapAdd(t *testing.T) {
	nM := NewNavigableMap(nil)
	path := []string{"FistLever2", "SecondLevel2", "Field2"}
	data := "Value2"
	nM.Add(path, data)
	path = []string{"FirstLevel", "SecondLevel", "ThirdLevel", "Fld1"}
	data = "Val1"
	nM.Add(path, data)
	path = []string{"FistLever2", "Field3"}
	data = "Value3"
	nM.Add(path, data)
	path = []string{"Field4"}
	data = "Val4"
	nM.Add(path, data)
	eNavMap := NavigableMap{
		data: map[string]interface{}{
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
		},
	}
	if !reflect.DeepEqual(nM.data, eNavMap.data) {
		t.Errorf("Expecting: %+v, received: %+v", eNavMap.data, nM.data)
	}
}

type testStruct struct {
	Item1 string
	Item2 int
}

func TestNavMapAdd2(t *testing.T) {
	nM := NewNavigableMap(nil)
	path := []string{"FistLever2", "SecondLevel2", "Field2"}
	data := 123
	nM.Add(path, data)
	path = []string{"FirstLevel", "SecondLevel", "ThirdLevel", "Fld1"}
	data1 := 123.123
	nM.Add(path, data1)
	path = []string{"FistLever2", "Field3"}
	data2 := "Value3"
	nM.Add(path, data2)
	path = []string{"Field4"}
	data3 := &testStruct{
		Item1: "Ten",
		Item2: 10,
	}
	nM.Add(path, data3)
	eNavMap := NavigableMap{
		data: map[string]interface{}{
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
		},
	}
	if !reflect.DeepEqual(nM.data, eNavMap.data) {
		t.Errorf("Expecting: %+v, received: %+v", eNavMap.data, nM.data)
	}
}

func TestNavMapItems(t *testing.T) {
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
	nM := NewNavigableMap(myData)
	eItems := []*NMItem{
		&NMItem{
			Path: []string{"FirstLevel", "SecondLevel", "ThirdLevel", "Fld1"},
			Data: "Val1",
		},
		&NMItem{
			Path: []string{"FistLever2", "SecondLevel2", "Field2"},
			Data: "Value2",
		},
		&NMItem{
			Path: []string{"FistLever2", "Field3"},
			Data: "Value3",
		},
		&NMItem{
			Path: []string{"Field4"},
			Data: "Val4",
		},
	}
	if !reflect.DeepEqual(len(nM.Items()), len(eItems)) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(eItems), utils.ToJSON(nM.Items()))
	}
}

func TestNavMapItems2(t *testing.T) {
	myData := map[string]interface{}{
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
	nM := NewNavigableMap(myData)
	eItems := []*NMItem{
		&NMItem{
			Path: []string{"FirstLevel", "SecondLevel", "ThirdLevel", "Fld1"},
			Data: 123.123,
		},
		&NMItem{
			Path: []string{"FistLever2", "SecondLevel2", "Field2"},
			Data: 123,
		},
		&NMItem{
			Path: []string{"FistLever2", "Field3"},
			Data: "Value3",
		},
		&NMItem{
			Path: []string{"Field4"},
			Data: &testStruct{
				Item1: "Ten",
				Item2: 10,
			},
		},
	}
	if !reflect.DeepEqual(len(nM.Items()), len(eItems)) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(eItems), utils.ToJSON(nM.Items()))
	}
}

func TestNavMapOrder(t *testing.T) {
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
	order := [][]string{
		[]string{"FirstLevel", "SecondLevel", "ThirdLevel", "Fld1"},
		[]string{"FistLever2", "SecondLevel2", "Field2"},
		[]string{"FistLever2", "Field3"},
		[]string{"Field4"},
	}
	nM := NewNavigableMap(myData)
	nM.order = order
	eItems := []*NMItem{
		&NMItem{
			Path: []string{"FirstLevel", "SecondLevel", "ThirdLevel", "Fld1"},
			Data: "Val1",
		},
		&NMItem{
			Path: []string{"FistLever2", "SecondLevel2", "Field2"},
			Data: "Value2",
		},
		&NMItem{
			Path: []string{"FistLever2", "Field3"},
			Data: "Value3",
		},
		&NMItem{
			Path: []string{"Field4"},
			Data: "Val4",
		},
	}
	if !reflect.DeepEqual(nM.Items(), eItems) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(eItems), utils.ToJSON(nM.Items()))
	}
}

func TestNavMapOrder2(t *testing.T) {
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
	order := [][]string{
		[]string{"FistLever2", "SecondLevel2", "Field2"},
		[]string{"Field4"},
		[]string{"FistLever2", "Field3"},
		[]string{"FirstLevel", "SecondLevel", "ThirdLevel", "Fld1"},
	}
	nM := NewNavigableMap(myData)
	nM.order = order
	eItems := []*NMItem{
		&NMItem{
			Path: []string{"FistLever2", "SecondLevel2", "Field2"},
			Data: "Value2",
		},
		&NMItem{
			Path: []string{"Field4"},
			Data: "Val4",
		},
		&NMItem{
			Path: []string{"FistLever2", "Field3"},
			Data: "Value3",
		},
		&NMItem{
			Path: []string{"FirstLevel", "SecondLevel", "ThirdLevel", "Fld1"},
			Data: "Val1",
		},
	}
	if !reflect.DeepEqual(nM.Items(), eItems) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(eItems), utils.ToJSON(nM.Items()))
	}
}

func TestNavMapIndexMapElementes(t *testing.T) {
	var elmsOut []*NMItem
	ifaceMap := map[string]interface{}{
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
	eItems := []*NMItem{
		&NMItem{
			Path: []string{"FirstLevel", "SecondLevel", "ThirdLevel", "Fld1"},
			Data: "Val1",
		},
		&NMItem{
			Path: []string{"FistLever2", "SecondLevel2", "Field2"},
			Data: "Value2",
		},
		&NMItem{
			Path: []string{"FistLever2", "Field3"},
			Data: "Value3",
		},
		&NMItem{
			Path: []string{"Field4"},
			Data: "Val4",
		},
	}
	indexMapElements(ifaceMap, []string{}, &elmsOut)
	if !reflect.DeepEqual(len(elmsOut), len(eItems)) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(eItems), utils.ToJSON(elmsOut))
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
	nM := NewNavigableMap(myData)
	eStr := utils.ToJSON(myData)
	if !reflect.DeepEqual(nM.String(), eStr) {
		t.Errorf("Expecting: %+v, received: %+v", eStr, nM.String())
	}
}
