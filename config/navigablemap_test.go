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
	"encoding/xml"
	"fmt"
	"reflect"
	"sort"
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
		utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

type myEv map[string]interface{}

func (ev myEv) AsNavigableMap() *NavigableMap {
	return NewNavigableMap(ev)
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

	if rcv := myData.AsNavigableMap(); err != nil {
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
	nM.Set(path, data, false, true)
	path = []string{"FirstLevel", "SecondLevel", "ThirdLevel", "Fld1"}
	data = "Val1"
	nM.Set(path, data, false, true)
	path = []string{"FistLever2", "Field3"}
	data = "Value3"
	nM.Set(path, data, false, true)
	path = []string{"Field4"}
	data = "Val4"
	nM.Set(path, data, false, true)
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
	eOrder := [][]string{
		[]string{"FistLever2", "SecondLevel2", "Field2"},
		[]string{"FirstLevel", "SecondLevel", "ThirdLevel", "Fld1"},
		[]string{"FistLever2", "Field3"},
		[]string{"Field4"},
	}
	if !reflect.DeepEqual(eOrder, nM.order) {
		t.Errorf("Expecting: %+v, received: %+v", eOrder, nM.order)
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
	nM.Set(path, data, false, true)
	path = []string{"FirstLevel", "SecondLevel", "ThirdLevel", "Fld1"}
	data1 := 123.123
	nM.Set(path, data1, false, true)
	path = []string{"FistLever2", "Field3"}
	data2 := "Value3"
	nM.Set(path, data2, false, true)
	path = []string{"Field4"}
	data3 := &testStruct{
		Item1: "Ten",
		Item2: 10,
	}
	nM.Set(path, data3, false, true)
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

func TestNavMapSetWithAppend(t *testing.T) {
	nM := NewNavigableMap(nil)
	itm := &NMItem{
		Path: []string{"FirstLevel", "SecondLevel", "ThirdLevel", "Fld1"},
		Data: "Val1",
	}
	nM.Set(itm.Path, []*NMItem{itm}, true, true)
	itm = &NMItem{
		Path: []string{"FirstLevel", "SecondLevel", "ThirdLevel", "Fld1"},
		Data: "Val2",
	}
	nM.Set(itm.Path, []*NMItem{itm}, true, true)
	eNavMap := &NavigableMap{
		data: map[string]interface{}{
			"FirstLevel": map[string]interface{}{
				"SecondLevel": map[string]interface{}{
					"ThirdLevel": map[string]interface{}{
						"Fld1": []*NMItem{
							{
								Path: []string{"FirstLevel", "SecondLevel", "ThirdLevel", "Fld1"},
								Data: "Val1",
							},
							{
								Path: []string{"FirstLevel", "SecondLevel", "ThirdLevel", "Fld1"},
								Data: "Val2",
							},
						},
					},
				},
			},
		},
		order: [][]string{
			{"FirstLevel", "SecondLevel", "ThirdLevel", "Fld1"},
			{"FirstLevel", "SecondLevel", "ThirdLevel", "Fld1"},
		},
	}
	if !reflect.DeepEqual(nM, eNavMap) {
		t.Errorf("expecting: %+v, received: %+v", eNavMap, nM)
	}
	nM = NewNavigableMap(nil)
	itm = &NMItem{
		Path: []string{"FirstLevel", "SecondLevel", "ThirdLevel", "Fld1"},
		Data: "Val1",
	}
	nM.Set(itm.Path, []*NMItem{itm}, false, true)
	itm = &NMItem{
		Path: []string{"FirstLevel", "SecondLevel", "ThirdLevel", "Fld1"},
		Data: "Val2",
	}
	nM.Set(itm.Path, []*NMItem{itm}, false, true)
	eNavMap = &NavigableMap{
		data: map[string]interface{}{
			"FirstLevel": map[string]interface{}{
				"SecondLevel": map[string]interface{}{
					"ThirdLevel": map[string]interface{}{
						"Fld1": []*NMItem{
							{
								Path: []string{"FirstLevel", "SecondLevel", "ThirdLevel", "Fld1"},
								Data: "Val2",
							},
						},
					},
				},
			},
		},
		order: [][]string{
			{"FirstLevel", "SecondLevel", "ThirdLevel", "Fld1"},
			{"FirstLevel", "SecondLevel", "ThirdLevel", "Fld1"},
		},
	}
	if !reflect.DeepEqual(nM, eNavMap) {
		t.Errorf("expecting: %+v, received: %+v", eNavMap, nM)
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
	if vals := nM.Values(); !reflect.DeepEqual(len(vals), len(eItems)) {
		t.Errorf("Expecting: %+v, received: %+v",
			utils.ToJSON(eItems), utils.ToJSON(vals))
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
	if vals := nM.Values(); !reflect.DeepEqual(len(vals), len(eItems)) {
		t.Errorf("Expecting: %+v, received: %+v",
			utils.ToJSON(eItems), utils.ToJSON(vals))
	}
}

func TestNavMapOrder(t *testing.T) {
	myData := map[string]interface{}{
		"FirstLevel": map[string]interface{}{
			"SecondLevel": map[string]interface{}{
				"ThirdLevel": map[string]interface{}{
					"Fld1": &NMItem{
						Path: []string{"FirstLevel", "SecondLevel", "ThirdLevel", "Fld1"},
						Data: "Val1",
					},
				},
			},
		},
		"FistLever2": map[string]interface{}{
			"SecondLevel2": map[string]interface{}{
				"Field2": &NMItem{
					Path: []string{"FistLever2", "SecondLevel2", "Field2"},
					Data: "Value2",
				},
			},
			"Field3": &NMItem{
				Path: []string{"FistLever2", "Field3"},
				Data: "Value3",
			},
		},
		"Field4": &NMItem{
			Path: []string{"Field4"},
			Data: "Val4",
		},
	}
	order := [][]string{
		[]string{"FirstLevel", "SecondLevel", "ThirdLevel", "Fld1"},
		[]string{"FistLever2", "SecondLevel2", "Field2"},
		[]string{"FistLever2", "Field3"},
		[]string{"Field4"},
	}
	nM := NewNavigableMap(myData)
	nM.order = order
	eItems := []interface{}{
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
	if vals := nM.Values(); !reflect.DeepEqual(vals, eItems) {
		t.Errorf("Expecting: %+v, received: %+v",
			utils.ToJSON(eItems), utils.ToJSON(vals))
	}
}

func TestNavMapOrder2(t *testing.T) {
	myData := map[string]interface{}{
		"FirstLevel": map[string]interface{}{
			"SecondLevel": map[string]interface{}{
				"ThirdLevel": map[string]interface{}{
					"Fld1": &NMItem{
						Path: []string{"FirstLevel", "SecondLevel", "ThirdLevel", "Fld1"},
						Data: "Val1",
					},
				},
			},
		},
		"FistLever2": map[string]interface{}{
			"SecondLevel2": map[string]interface{}{
				"Field2": &NMItem{
					Path: []string{"FistLever2", "SecondLevel2", "Field2"},
					Data: "Value2",
				},
			},
			"Field3": &NMItem{
				Path: []string{"FistLever2", "Field3"},
				Data: "Value3",
			},
		},
		"Field4": &NMItem{
			Path: []string{"Field4"},
			Data: "Val4",
		},
	}
	order := [][]string{
		[]string{"FistLever2", "SecondLevel2", "Field2"},
		[]string{"Field4"},
		[]string{"FistLever2", "Field3"},
		[]string{"FirstLevel", "SecondLevel", "ThirdLevel", "Fld1"},
	}
	nM := NewNavigableMap(myData)
	nM.order = order
	eItems := []interface{}{
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
	if vals := nM.Values(); !reflect.DeepEqual(eItems, vals) {
		t.Errorf("Expecting: %+v, received: %+v",
			utils.ToJSON(eItems), utils.ToJSON(vals))
	}
}

func TestNavMapIndexMapElementes(t *testing.T) {
	var elmsOut []interface{}
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
	eItems := []interface{}{"Val1", "Value2", "Value3", "Val4"}
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

func TestNavMapAsXMLElements(t *testing.T) {
	nM := &NavigableMap{
		data: map[string]interface{}{
			"FirstLevel": map[string]interface{}{
				"SecondLevel": map[string]interface{}{
					"ThirdLevel": map[string]interface{}{
						"Fld1": []*NMItem{
							&NMItem{Path: []string{"FirstLevel", "SecondLevel", "ThirdLevel", "Fld1"},
								Data: "Val1"}},
					},
				},
			},
			"FirstLevel2": map[string]interface{}{
				"SecondLevel2": map[string]interface{}{
					"Field2": []*NMItem{
						&NMItem{Path: []string{"FirstLevel2", "SecondLevel2", "Field2"},
							Data:   "attrVal1",
							Config: &FCTemplate{Tag: "AttributeTest", AttributeID: "attribute1"}},
						&NMItem{Path: []string{"FirstLevel2", "SecondLevel2", "Field2"},
							Data: "Value2"}},
				},
				"Field3": []*NMItem{
					&NMItem{Path: []string{"FirstLevel2", "Field3"},
						Data: "Value3"}},
				"Field5": []*NMItem{
					&NMItem{Path: []string{"FirstLevel2", "Field5"},
						Data: "Value5"},
					&NMItem{Path: []string{"FirstLevel2", "Field5"},
						Data:   "attrVal5",
						Config: &FCTemplate{Tag: "AttributeTest", AttributeID: "attribute5"}}},
				"Field6": []*NMItem{
					&NMItem{Path: []string{"FirstLevel2", "Field6"},
						Data:   "Value6",
						Config: &FCTemplate{Tag: "NewBranchTest", NewBranch: true}},
					&NMItem{Path: []string{"FirstLevel2", "Field6"},
						Data:   "attrVal6",
						Config: &FCTemplate{Tag: "AttributeTest", AttributeID: "attribute6"}},
				},
			},
			"Field4": []*NMItem{
				&NMItem{Path: []string{"Field4"},
					Data: "Val4"},
				&NMItem{Path: []string{"Field4"},
					Data:   "attrVal2",
					Config: &FCTemplate{Tag: "AttributeTest", AttributeID: "attribute2"}}},
		},
		order: [][]string{
			[]string{"FirstLevel2", "SecondLevel2", "Field2"},
			[]string{"FirstLevel", "SecondLevel", "ThirdLevel", "Fld1"},
			[]string{"FirstLevel2", "Field3"},
			[]string{"FirstLevel2", "Field5"},
			[]string{"Field4"},
			[]string{"FirstLevel2", "Field6"},
		},
	}
	eXMLElmnts := []*XMLElement{
		&XMLElement{
			XMLName: xml.Name{Local: nM.order[0][0]},
			Elements: []*XMLElement{
				&XMLElement{
					XMLName: xml.Name{Local: nM.order[0][1]},
					Elements: []*XMLElement{
						&XMLElement{
							XMLName: xml.Name{Local: nM.order[0][2]},
							Attributes: []*xml.Attr{
								&xml.Attr{
									Name:  xml.Name{Local: "attribute1"},
									Value: "attrVal1",
								},
							},
							Value: "Value2",
						},
					},
				},
				&XMLElement{
					XMLName: xml.Name{Local: nM.order[2][1]},
					Value:   "Value3",
				},
				&XMLElement{
					XMLName: xml.Name{Local: nM.order[3][1]},
					Attributes: []*xml.Attr{
						&xml.Attr{
							Name:  xml.Name{Local: "attribute5"},
							Value: "attrVal5",
						},
					},
					Value: "Value5",
				},
			},
		},
		&XMLElement{
			XMLName: xml.Name{Local: nM.order[1][0]},
			Elements: []*XMLElement{
				&XMLElement{
					XMLName: xml.Name{Local: nM.order[1][1]},
					Elements: []*XMLElement{
						&XMLElement{
							XMLName: xml.Name{Local: nM.order[1][2]},
							Elements: []*XMLElement{
								&XMLElement{
									XMLName: xml.Name{Local: nM.order[1][3]},
									Value:   "Val1",
								},
							},
						},
					},
				},
			},
		},
		&XMLElement{
			XMLName: xml.Name{Local: nM.order[4][0]},
			Attributes: []*xml.Attr{
				&xml.Attr{
					Name:  xml.Name{Local: "attribute2"},
					Value: "attrVal2",
				},
			},
			Value: "Val4",
		},
		&XMLElement{
			XMLName: xml.Name{Local: nM.order[5][0]},
			Elements: []*XMLElement{
				&XMLElement{
					XMLName: xml.Name{Local: nM.order[5][1]},
					Attributes: []*xml.Attr{
						&xml.Attr{
							Name:  xml.Name{Local: "attribute6"},
							Value: "attrVal6",
						},
					},
					Value: "Value6",
				},
			},
		},
	}
	xmlEnts, err := nM.AsXMLElements()
	if err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eXMLElmnts, xmlEnts) {
		t.Errorf("expecting: %s, received: %s", utils.ToJSON(eXMLElmnts), utils.ToJSON(xmlEnts))
	}
	eXML := []byte(`<FirstLevel2>
  <SecondLevel2>
    <Field2 attribute1="attrVal1">Value2</Field2>
  </SecondLevel2>
  <Field3>Value3</Field3>
  <Field5 attribute5="attrVal5">Value5</Field5>
</FirstLevel2>
<FirstLevel>
  <SecondLevel>
    <ThirdLevel>
      <Fld1>Val1</Fld1>
    </ThirdLevel>
  </SecondLevel>
</FirstLevel>
<Field4 attribute2="attrVal2">Val4</Field4>
<FirstLevel2>
  <Field6 attribute6="attrVal6">Value6</Field6>
</FirstLevel2>`)
	if output, err := xml.MarshalIndent(xmlEnts, "", "  "); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eXML, output) {
		t.Errorf("expecting: \n%s, received: \n%s\n", string(eXML), string(output))
	}
}

func TestIndexMapPaths(t *testing.T) {
	mp := make(map[string]interface{})
	parsedPaths := make([][]string, 0)
	eParsedPaths := make([][]string, 0)
	if indexMapPaths(mp, nil, &parsedPaths); !reflect.DeepEqual(eParsedPaths, parsedPaths) {
		t.Errorf("expecting: %+v, received: %+v", eParsedPaths, parsedPaths)
	}
	mp = map[string]interface{}{
		"a": "a",
	}
	parsedPaths = make([][]string, 0)
	eParsedPaths = [][]string{
		[]string{"a"},
	}
	if indexMapPaths(mp, nil, &parsedPaths); !reflect.DeepEqual(eParsedPaths, parsedPaths) {
		t.Errorf("expecting: %+v, received: %+v", eParsedPaths, parsedPaths)
	}
	mp = map[string]interface{}{
		"a": map[string]interface{}{
			"a1": "a",
		},
		"b": "b",
	}
	parsedPaths = make([][]string, 0)
	eParsedPaths = [][]string{
		[]string{"a", "a1"},
		[]string{"b"},
	}
	if indexMapPaths(mp, nil, &parsedPaths); len(eParsedPaths) != len(parsedPaths) {
		t.Errorf("expecting: %+v, received: %+v", eParsedPaths, parsedPaths)
	}
}

func TestNavMapAsCGREvent(t *testing.T) {
	nM := &NavigableMap{
		data: map[string]interface{}{
			"FirstLevel": map[string]interface{}{
				"SecondLevel": map[string]interface{}{
					"ThirdLevel": map[string]interface{}{
						"Fld1": []*NMItem{
							&NMItem{Path: []string{"FirstLevel", "SecondLevel", "ThirdLevel", "Fld1"},
								Data: "Val1"}},
					},
				},
			},
			"FirstLevel2": map[string]interface{}{
				"SecondLevel2": map[string]interface{}{
					"Field2": []*NMItem{
						&NMItem{Path: []string{"FirstLevel2", "SecondLevel2", "Field2"},
							Data: "attrVal1",
							Config: &FCTemplate{Tag: "AttributeTest",
								AttributeID: "attribute1"}},
						&NMItem{Path: []string{"FirstLevel2", "SecondLevel2", "Field2"},
							Data: "Value2"}},
				},
				"Field3": []*NMItem{
					&NMItem{Path: []string{"FirstLevel2", "Field3"},
						Data: "Value3"}},
				"Field5": []*NMItem{
					&NMItem{Path: []string{"FirstLevel2", "Field5"},
						Data: "Value5"},
					&NMItem{Path: []string{"FirstLevel2", "Field5"},
						Data: "attrVal5",
						Config: &FCTemplate{Tag: "AttributeTest",
							AttributeID: "attribute5"}}},
				"Field6": []*NMItem{
					&NMItem{Path: []string{"FirstLevel2", "Field6"},
						Data: "Value6",
						Config: &FCTemplate{Tag: "NewBranchTest",
							NewBranch: true}},
					&NMItem{Path: []string{"FirstLevel2", "Field6"},
						Data: "attrVal6",
						Config: &FCTemplate{Tag: "AttributeTest",
							AttributeID: "attribute6"}},
				},
			},
			"Field4": []*NMItem{
				&NMItem{Path: []string{"Field4"},
					Data: "Val4"},
				&NMItem{Path: []string{"Field4"},
					Data: "attrVal2",
					Config: &FCTemplate{Tag: "AttributeTest",
						AttributeID: "attribute2"}}},
		},
	}
	eEv := map[string]interface{}{
		"FirstLevel2.SecondLevel2.Field2":        "Value2",
		"FirstLevel.SecondLevel.ThirdLevel.Fld1": "Val1",
		"FirstLevel2.Field3":                     "Value3",
		"FirstLevel2.Field5":                     "Value5",
		"FirstLevel2.Field6":                     "Value6",
		"Field4":                                 "Val4",
	}
	if cgrEv := nM.AsCGREvent("cgrates.org",
		utils.NestingSep); cgrEv.Tenant != "cgrates.org" ||
		cgrEv.Time == nil ||
		!reflect.DeepEqual(eEv, cgrEv.Event) {
		t.Errorf("expecting: %+v, \nreceived: %+v", utils.ToJSON(eEv), utils.ToJSON(cgrEv.Event))
	}
}

func TestNavMapMerge(t *testing.T) {
	nM2 := &NavigableMap{
		data: map[string]interface{}{
			"FirstLevel": map[string]interface{}{
				"SecondLevel": map[string]interface{}{
					"ThirdLevel": map[string]interface{}{
						"Fld1": []*NMItem{
							{
								Path: []string{"FirstLevel", "SecondLevel", "ThirdLevel", "Fld1"},
								Data: "Val1",
							},
							{
								Path: []string{"FirstLevel", "SecondLevel", "ThirdLevel", "Fld1"},
								Data: "Val2",
							},
						},
					},
				},
			},
		},
		order: [][]string{
			{"FirstLevel", "SecondLevel", "ThirdLevel", "Fld1"},
			{"FirstLevel", "SecondLevel", "ThirdLevel", "Fld1"},
		},
	}
	nM := NewNavigableMap(nil)
	nM.Merge(nM2)
	if !reflect.DeepEqual(nM2, nM) {
		t.Errorf("expecting: %+v, received: %+v", nM2, nM)
	}
}

func TestNavMapGetIndex(t *testing.T) {
	nM := NewNavigableMap(nil)
	var expIndx *int
	var expPath string
	var path string = "Fld1"
	expPath = path
	if rplyPath, rplyIndx := nM.getIndex(path); rplyPath != expPath && rplyIndx != expIndx {
		t.Errorf("Expected: path=%s ,indx=%v, received: path=%s ,indx=%v", expPath, expIndx, rplyPath, rplyIndx)
	}

	path = "slice[5]"
	expPath = "slice"
	expIndx = utils.IntPointer(5)
	if rplyPath, rplyIndx := nM.getIndex(path); rplyPath != expPath && rplyIndx != expIndx {
		t.Errorf("Expected: path=%s ,indx=%v, received: path=%s ,indx=%v", expPath, expIndx, rplyPath, rplyIndx)
	}

	path = "slice[~cgreq.Count]"
	expPath = "slice"
	expIndx = nil
	if rplyPath, rplyIndx := nM.getIndex(path); rplyPath != expPath && rplyIndx != expIndx {
		t.Errorf("Expected: path=%s ,indx=%v, received: path=%s ,indx=%v", expPath, expIndx, rplyPath, rplyIndx)
	}
}

func TestNavMapGetNextMap(t *testing.T) {
	nM := NewNavigableMap(nil)
	mp := map[string]interface{}{
		"field1": 10,
		"field2": []string{"val1", "val2"},
		"field3": []int{1, 2, 3},
		"map1":   map[string]interface{}{"field1": 100000},
		"map2":   []map[string]interface{}{map[string]interface{}{"field2": 11}},
		"map3":   []NavigableMap{NavigableMap{data: map[string]interface{}{"field4": 112}}},
	}
	path := "map4"
	if _, err := nM.getNextMap(mp, path); err != nil && err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expected error: %s , received error %v", utils.ErrNotFound.Error(), err)
	}
	path = "map2[10]"
	if _, err := nM.getNextMap(mp, path); err != nil && err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expected error: %s , received error %v", utils.ErrNotFound.Error(), err)
	}
	path = "field1"
	experr := fmt.Errorf("cannot cast field: <%+v> type: %T with path: <%s> to map[string]interface{}",
		mp[path], mp[path], path)
	if _, err := nM.getNextMap(mp, path); err != nil && err.Error() != experr.Error() {
		t.Errorf("Expected error: %s , received error %v", experr.Error(), err)
	}
	path = "map1"
	expMap := map[string]interface{}{"field1": 100000}
	if rm, err := nM.getNextMap(mp, path); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rm, expMap) {
		t.Errorf("Expected: %s, received: %s", utils.ToJSON(expMap), utils.ToJSON(rm))
	}
	path = "map2[0]"
	expMap = map[string]interface{}{"field2": 11}
	if rm, err := nM.getNextMap(mp, path); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rm, expMap) {
		t.Errorf("Expected: %s, received: %s", utils.ToJSON(expMap), utils.ToJSON(rm))
	}
	path = "map3[0]"
	expMap = map[string]interface{}{"field4": 112}
	if rm, err := nM.getNextMap(mp, path); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rm, expMap) {
		t.Errorf("Expected: %s, received: %s", utils.ToJSON(expMap), utils.ToJSON(rm))
	}
}

func TestNavMapgetLastItem(t *testing.T) {
	nM := NewNavigableMap(nil)
	mp := map[string]interface{}{
		"field1": 10,
		"field2": []string{"val1", "val2"},
		"field3": []int{1, 2, 3},
		"map1":   map[string]interface{}{"field1": 100000},
		"map2":   []map[string]interface{}{map[string]interface{}{"field2": 11}},
		"map3":   []NavigableMap{NavigableMap{data: map[string]interface{}{"field4": 112}}},
	}
	path := "map4"
	if _, err := nM.getLastItem(mp, path); err != nil && err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expected error: %s , received error %v", utils.ErrNotFound.Error(), err)
	}
	path = "map2[10]"
	if _, err := nM.getLastItem(mp, path); err != nil && err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expected error: %s , received error %v", utils.ErrNotFound.Error(), err)
	}
	path = "field1"
	var expVal interface{} = 10
	if rplyVal, err := nM.getLastItem(mp, path); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expVal, rplyVal) {
		t.Errorf("Expected: %v ,received: %v", expVal, rplyVal)
	}

	path = "field2[1]"
	expVal = "val2"
	if rplyVal, err := nM.getLastItem(mp, path); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expVal, rplyVal) {
		t.Errorf("Expected: %v ,received: %v", expVal, rplyVal)
	}
	path = "field3[2]"
	expVal = 3
	if rplyVal, err := nM.getLastItem(mp, path); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expVal, rplyVal) {
		t.Errorf("Expected: %v ,received: %v", expVal, rplyVal)
	}
	path = "field2"
	expVal = []string{"val1", "val2"}
	if rplyVal, err := nM.getLastItem(mp, path); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expVal, rplyVal) {
		t.Errorf("Expected: %v ,received: %v", expVal, rplyVal)
	}
}

func TestNavMapGetField(t *testing.T) {
	nM := &NavigableMap{
		data: map[string]interface{}{
			"FirstLevel": map[string]interface{}{
				"SecondLevel": map[string]interface{}{
					"ThirdLevel": map[string]interface{}{
						"Fld1": []*NMItem{
							{
								Path: []string{"FirstLevel", "SecondLevel", "ThirdLevel", "Fld1"},
								Data: "Val1",
							},
							{
								Path: []string{"FirstLevel", "SecondLevel", "ThirdLevel", "Fld1"},
								Data: "Val2",
							},
						},
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
		},
	}
	pth := []string{"FirstLevel", "SecondLevel", "ThirdLevel", "Fld1[0]"}
	eFld := &NMItem{
		Path: []string{"FirstLevel", "SecondLevel", "ThirdLevel", "Fld1"},
		Data: "Val1",
	}
	if fld, err := nM.GetField(pth); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eFld, fld) {
		t.Errorf("expecting: %s, received: %s", utils.ToIJSON(eFld), utils.ToIJSON(fld))
	}
	eFld2 := map[string]interface{}{"Fld1": "Val1"}
	pth = []string{"FirstLevel2", "SecondLevel2[0]", "ThirdLevel2"}
	if fld, err := nM.GetField(pth); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eFld2, fld) {
		t.Errorf("expecting: %s, received: %s", utils.ToIJSON(eFld2), utils.ToIJSON(fld))
	}
	eFld3 := "ValAnotherFirstLevel"
	pth = []string{"AnotherFirstLevel"}
	if fld, err := nM.GetField(pth); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eFld3, fld) {
		t.Errorf("expecting: %s, received: %s", utils.ToIJSON(eFld3), utils.ToIJSON(fld))
	}
	pth = []string{"AnotherFirstLevel2"}
	if _, err := nM.GetField(pth); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	}
	pth = []string{"FirstLevel", "SecondLevel[1]", "ThirdLevel", "Fld1[0]"}
	if _, err := nM.GetField(pth); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	}
}

func TestNavMapFieldAsInterface(t *testing.T) {
	nM := &NavigableMap{
		data: map[string]interface{}{
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
		},
	}

	path := []string{"FirstLevel", "SecondLevel[0]", "Count"}
	expErr := utils.ErrNotFound
	var eVal interface{} = nil
	if _, err := nM.FieldAsInterface(path); err != nil && err.Error() != expErr.Error() {
		t.Errorf("Expected error: %s, received error: %v", expErr.Error(), err)
	}

	path = []string{"AnotherFirstLevel", "SecondLevel", "Count"}
	expErr = fmt.Errorf("cannot cast field: <%+v> type: %T with path: <%s> to map[string]interface{}",
		nM.data["AnotherFirstLevel"], nM.data["AnotherFirstLevel"], "AnotherFirstLevel")
	if _, err := nM.FieldAsInterface(path); err != nil && err.Error() != expErr.Error() {
		t.Errorf("Expected error: %s, received error: %v", expErr.Error(), err)
	}

	path = []string{"FirstLevel", "SecondLevel[1]", "Count"}
	eVal = 10
	if rplyVal, err := nM.FieldAsInterface(path); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eVal, rplyVal) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(eVal), utils.ToJSON(rplyVal))
	}

	path = []string{"FirstLevel", "SecondLevel[1]", "ThirdLevel2", "Fld2"}
	eVal = []string{"Val1", "Val2", "Val3"}
	if rplyVal, err := nM.FieldAsInterface(path); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eVal, rplyVal) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(eVal), utils.ToJSON(rplyVal))
	}

	path = []string{"FirstLevel", "SecondLevel[1]", "ThirdLevel2", "Fld2[2]"}
	eVal = "Val3"
	if rplyVal, err := nM.FieldAsInterface(path); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eVal, rplyVal) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(eVal), utils.ToJSON(rplyVal))
	}
}

func TestNavMapGetKeys(t *testing.T) {
	navMp := NewNavigableMap(
		map[string]interface{}{
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
		},
	)
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
	keys := navMp.GetKeys()
	sort.Strings(expKeys)
	sort.Strings(keys)
	if !reflect.DeepEqual(expKeys, keys) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(expKeys), utils.ToJSON(keys))
	}
}

func TestNMAsXMLElements(t *testing.T) {
	nM := utils.NewOrderedNavigableMap()
	order := []utils.PathItems{
		{{Field: "FirstLevel2"}, {Field: "SecondLevel2"}, {Field: "Field2"}},
		{{Field: "FirstLevel"}, {Field: "SecondLevel"}, {Field: "ThirdLevel"}, {Field: "Fld1"}},
		{{Field: "FirstLevel2"}, {Field: "Field3"}},
		{{Field: "FirstLevel2"}, {Field: "Field5"}},
		{{Field: "Field4"}},
		{{Field: "FirstLevel2"}, {Field: "Field6"}},
	}
	if _, err := nM.Set(&utils.FullPath{Path: order[0].String(), PathItems: order[0]}, &utils.NMSlice{
		&NMItem{Path: strings.Split(order[0].String(), utils.NestingSep),
			Data:   "attrVal1",
			Config: &FCTemplate{Tag: "AttributeTest", AttributeID: "attribute1"}},
		&NMItem{Path: strings.Split(order[0].String(), utils.NestingSep),
			Data: "Value2"}}); err != nil {
		t.Error(err)
	}
	if _, err := nM.Set(&utils.FullPath{Path: order[1].String(), PathItems: order[1]}, &utils.NMSlice{
		&NMItem{Path: strings.Split(order[1].String(), utils.NestingSep),
			Data: "Val1"}}); err != nil {
		t.Error(err)
	}
	if _, err := nM.Set(&utils.FullPath{Path: order[2].String(), PathItems: order[2]}, &utils.NMSlice{
		&NMItem{Path: strings.Split(order[2].String(), utils.NestingSep),
			Data: "Value3"}}); err != nil {
		t.Error(err)
	}
	if _, err := nM.Set(&utils.FullPath{Path: order[3].String(), PathItems: order[3]}, &utils.NMSlice{
		&NMItem{Path: strings.Split(order[3].String(), utils.NestingSep),
			Data: "Value5"},
		&NMItem{Path: strings.Split(order[3].String(), utils.NestingSep),
			Data:   "attrVal5",
			Config: &FCTemplate{Tag: "AttributeTest", AttributeID: "attribute5"}}}); err != nil {
		t.Error(err)
	}
	if _, err := nM.Set(&utils.FullPath{Path: order[4].String(), PathItems: order[4]}, &utils.NMSlice{
		&NMItem{Path: strings.Split(order[4].String(), utils.NestingSep),
			Data: "Val4"},
		&NMItem{Path: strings.Split(order[4].String(), utils.NestingSep),
			Data:   "attrVal2",
			Config: &FCTemplate{Tag: "AttributeTest", AttributeID: "attribute2"}}}); err != nil {
		t.Error(err)
	}
	if _, err := nM.Set(&utils.FullPath{Path: order[5].String(), PathItems: order[5]}, &utils.NMSlice{
		&NMItem{Path: strings.Split(order[5].String(), utils.NestingSep),
			Data:   "Value6",
			Config: &FCTemplate{Tag: "NewBranchTest", NewBranch: true}},
		&NMItem{Path: strings.Split(order[5].String(), utils.NestingSep),
			Data:   "attrVal6",
			Config: &FCTemplate{Tag: "AttributeTest", AttributeID: "attribute6"}}}); err != nil {
		t.Error(err)
	}
	eXMLElmnts := []*XMLElement{
		&XMLElement{
			XMLName: xml.Name{Local: order[0][0].String()},
			Elements: []*XMLElement{
				&XMLElement{
					XMLName: xml.Name{Local: order[0][1].String()},
					Elements: []*XMLElement{
						&XMLElement{
							XMLName: xml.Name{Local: order[0][2].String()},
							Attributes: []*xml.Attr{
								&xml.Attr{
									Name:  xml.Name{Local: "attribute1"},
									Value: "attrVal1",
								},
							},
							Value: "Value2",
						},
					},
				},
				&XMLElement{
					XMLName: xml.Name{Local: "Field3"},
					Value:   "Value3",
				},
				&XMLElement{
					XMLName: xml.Name{Local: order[3][1].String()},
					Attributes: []*xml.Attr{
						&xml.Attr{
							Name:  xml.Name{Local: "attribute5"},
							Value: "attrVal5",
						},
					},
					Value: "Value5",
				},
			},
		},
		&XMLElement{
			XMLName: xml.Name{Local: order[1][0].String()},
			Elements: []*XMLElement{
				&XMLElement{
					XMLName: xml.Name{Local: order[1][1].String()},
					Elements: []*XMLElement{
						&XMLElement{
							XMLName: xml.Name{Local: order[1][2].String()},
							Elements: []*XMLElement{
								&XMLElement{
									XMLName: xml.Name{Local: "Fld1"},
									Value:   "Val1",
								},
							},
						},
					},
				},
			},
		},
		&XMLElement{
			XMLName: xml.Name{Local: order[4][0].String()},
			Attributes: []*xml.Attr{
				&xml.Attr{
					Name:  xml.Name{Local: "attribute2"},
					Value: "attrVal2",
				},
			},
			Value: "Val4",
		},
		&XMLElement{
			XMLName: xml.Name{Local: order[5][0].String()},
			Elements: []*XMLElement{
				&XMLElement{
					XMLName: xml.Name{Local: order[5][1].String()},
					Attributes: []*xml.Attr{
						&xml.Attr{
							Name:  xml.Name{Local: "attribute6"},
							Value: "attrVal6",
						},
					},
					Value: "Value6",
				},
			},
		},
	}
	xmlEnts, err := NMAsXMLElements(nM)
	if err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eXMLElmnts, xmlEnts) {
		t.Errorf("expecting: %s, received: %s", utils.ToJSON(eXMLElmnts), utils.ToJSON(xmlEnts))
	}
	eXML := []byte(`<FirstLevel2>
  <SecondLevel2>
    <Field2 attribute1="attrVal1">Value2</Field2>
  </SecondLevel2>
  <Field3>Value3</Field3>
  <Field5 attribute5="attrVal5">Value5</Field5>
</FirstLevel2>
<FirstLevel>
  <SecondLevel>
    <ThirdLevel>
      <Fld1>Val1</Fld1>
    </ThirdLevel>
  </SecondLevel>
</FirstLevel>
<Field4 attribute2="attrVal2">Val4</Field4>
<FirstLevel2>
  <Field6 attribute6="attrVal6">Value6</Field6>
</FirstLevel2>`)
	if output, err := xml.MarshalIndent(xmlEnts, "", "  "); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eXML, output) {
		t.Errorf("expecting: \n%s, received: \n%s\n", string(eXML), string(output))
	}
}

func TestNMAsCGREvent(t *testing.T) {
	if cgrEv := NMAsCGREvent(nil, "cgrates.org",
		utils.NestingSep); cgrEv != nil {
		t.Errorf("expecting: %+v, \nreceived: %+v", utils.ToJSON(nil), utils.ToJSON(cgrEv.Event))
	}

	nM := utils.NewOrderedNavigableMap()
	if cgrEv := NMAsCGREvent(nM, "cgrates.org",
		utils.NestingSep); cgrEv != nil {
		t.Errorf("expecting: %+v, \nreceived: %+v", utils.ToJSON(nil), utils.ToJSON(cgrEv.Event))
	}

	path := utils.PathItems{{Field: "FirstLevel"}, {Field: "SecondLevel"}, {Field: "ThirdLevel"}, {Field: "Fld1"}}
	if _, err := nM.Set(&utils.FullPath{Path: path.String(), PathItems: path}, &utils.NMSlice{&NMItem{
		Path: strings.Split(path.String(), utils.NestingSep),
		Data: "Val1",
	}}); err != nil {
		t.Error(err)
	}

	path = utils.PathItems{{Field: "FirstLevel2"}, {Field: "SecondLevel2"}, {Field: "Field2"}}
	if _, err := nM.Set(&utils.FullPath{Path: path.String(), PathItems: path}, &utils.NMSlice{&NMItem{
		Path: strings.Split(path.String(), utils.NestingSep),
		Data: "attrVal1",
		Config: &FCTemplate{Tag: "AttributeTest",
			AttributeID: "attribute1"},
	}, &NMItem{
		Path: strings.Split(path.String(), utils.NestingSep),
		Data: "Value2",
	}}); err != nil {
		t.Error(err)
	}

	path = utils.PathItems{{Field: "FirstLevel2"}, {Field: "Field3"}}
	if _, err := nM.Set(&utils.FullPath{Path: path.String(), PathItems: path}, &utils.NMSlice{&NMItem{
		Path: strings.Split(path.String(), utils.NestingSep),
		Data: "Value3",
	}}); err != nil {
		t.Error(err)
	}

	path = utils.PathItems{{Field: "FirstLevel2"}, {Field: "Field5"}}
	if _, err := nM.Set(&utils.FullPath{Path: path.String(), PathItems: path}, &utils.NMSlice{&NMItem{
		Path: strings.Split(path.String(), utils.NestingSep),
		Data: "Value5",
	}, &NMItem{
		Path: strings.Split(path.String(), utils.NestingSep),
		Data: "attrVal5",
		Config: &FCTemplate{Tag: "AttributeTest",
			AttributeID: "attribute5"},
	}}); err != nil {
		t.Error(err)
	}

	path = utils.PathItems{{Field: "FirstLevel2"}, {Field: "Field6"}}
	if _, err := nM.Set(&utils.FullPath{Path: path.String(), PathItems: path}, &utils.NMSlice{&NMItem{
		Path: strings.Split(path.String(), utils.NestingSep),
		Data: "Value6",
		Config: &FCTemplate{Tag: "NewBranchTest",
			NewBranch: true},
	}, &NMItem{
		Path: strings.Split(path.String(), utils.NestingSep),
		Data: "attrVal6",
		Config: &FCTemplate{Tag: "AttributeTest",
			AttributeID: "attribute6"},
	}}); err != nil {
		t.Error(err)
	}

	path = utils.PathItems{{Field: "Field4"}}
	if _, err := nM.Set(&utils.FullPath{Path: path.String(), PathItems: path}, &utils.NMSlice{&NMItem{
		Path: strings.Split(path.String(), utils.NestingSep),
		Data: "Val4",
	}, &NMItem{
		Path: strings.Split(path.String(), utils.NestingSep),
		Data: "attrVal2",
		Config: &FCTemplate{Tag: "AttributeTest",
			AttributeID: "attribute2"},
	}}); err != nil {
		t.Error(err)
	}
	eEv := map[string]interface{}{
		"FirstLevel2.SecondLevel2.Field2":        "Value2",
		"FirstLevel.SecondLevel.ThirdLevel.Fld1": "Val1",
		"FirstLevel2.Field3":                     "Value3",
		"FirstLevel2.Field5":                     "Value5",
		"FirstLevel2.Field6":                     "Value6",
		"Field4":                                 "Val4",
	}
	if cgrEv := NMAsCGREvent(nM, "cgrates.org",
		utils.NestingSep); cgrEv.Tenant != "cgrates.org" ||
		cgrEv.Time == nil ||
		!reflect.DeepEqual(eEv, cgrEv.Event) {
		t.Errorf("expecting: %+v, \nreceived: %+v", utils.ToJSON(eEv), utils.ToJSON(cgrEv.Event))
	}
}

func TestNMItemLen(t *testing.T) {
	var nm utils.NMInterface = &NMItem{Data: "1001"}
	if rply := nm.Len(); rply != 0 {
		t.Errorf("Expected 0 ,received: %v", rply)
	}
}

func TestNMItemString(t *testing.T) {
	var nm utils.NMInterface = &NMItem{Data: "1001"}
	expected := "{\"Path\":null,\"Data\":\"1001\",\"Config\":null}"
	if rply := nm.String(); rply != expected {
		t.Errorf("Expected %q ,received: %q", expected, rply)
	}
}

func TestNMItemInterface(t *testing.T) {
	var nm utils.NMInterface = &NMItem{Data: "1001"}
	expected := "1001"
	if rply := nm.Interface(); rply != expected {
		t.Errorf("Expected %q ,received: %q", expected, rply)
	}
}

func TestNMItemField(t *testing.T) {
	var nm utils.NMInterface = &NMItem{Data: "1001"}
	if _, err := nm.Field(nil); err != utils.ErrNotImplemented {
		t.Error(err)
	}
}

func TestNMItemRemove(t *testing.T) {
	var nm utils.NMInterface = &NMItem{Data: "1001"}
	if err := nm.Remove(nil); err != utils.ErrNotImplemented {
		t.Error(err)
	}
}

func TestNMItemEmpty(t *testing.T) {
	var nm utils.NMInterface = &NMItem{Data: "1001"}
	if nm.Empty() {
		t.Error("Expected not empty type")
	}
	nm = &NMItem{Data: nil}
	if !nm.Empty() {
		t.Error("Expected empty type")
	}
}

func TestNMItemType(t *testing.T) {
	var nm utils.NMInterface = &NMItem{Data: "1001"}
	if nm.Type() != utils.NMDataType {
		t.Errorf("Expected %v ,received: %v", utils.NMDataType, nm.Type())
	}
}

func TestNMItemSet(t *testing.T) {
	var nm utils.NMInterface = &NMItem{Data: "1001"}
	if _, err := nm.Set(utils.PathItems{{}}, nil); err != utils.ErrWrongPath {
		t.Error(err)
	}
	if _, err := nm.Set(nil, &NMItem{Data: "1002"}); err != nil {
		t.Error(err)
	}
	expected := "1002"
	if rply := nm.Interface(); rply != expected {
		t.Errorf("Expected %q ,received: %q", expected, rply)
	}
}
