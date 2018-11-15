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
		utils.ErrNotFound.Error() {
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
