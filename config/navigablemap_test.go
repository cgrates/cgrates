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

func TestNavMapAsXMLElements(t *testing.T) {
	nM := utils.NewOrderedNavigableMap()
	order := []utils.PathItems{
		{{Field: "FirstLevel2"}, {Field: "SecondLevel2"}, {Field: "Field2"}},
		{{Field: "FirstLevel"}, {Field: "SecondLevel"}, {Field: "ThirdLevel"}, {Field: "Fld1"}},
		{{Field: "FirstLevel2"}, {Field: "Field3"}},
		{{Field: "FirstLevel2"}, {Field: "Field5"}},
		{{Field: "Field4"}},
		{{Field: "FirstLevel2"}, {Field: "Field6"}},
	}
	if err := nM.Set(order[0], &utils.NMSlice{
		&NMItem{Path: strings.Split(order[0].String(), utils.NestingSep),
			Data:   "attrVal1",
			Config: &FCTemplate{Tag: "AttributeTest", AttributeID: "attribute1"}},
		&NMItem{Path: strings.Split(order[0].String(), utils.NestingSep),
			Data: "Value2"}}); err != nil {
		t.Error(err)
	}
	if err := nM.Set(order[1], &utils.NMSlice{
		&NMItem{Path: strings.Split(order[1].String(), utils.NestingSep),
			Data: "Val1"}}); err != nil {
		t.Error(err)
	}
	if err := nM.Set(order[2], &utils.NMSlice{
		&NMItem{Path: strings.Split(order[2].String(), utils.NestingSep),
			Data: "Value3"}}); err != nil {
		t.Error(err)
	}
	if err := nM.Set(order[3], &utils.NMSlice{
		&NMItem{Path: strings.Split(order[3].String(), utils.NestingSep),
			Data: "Value5"},
		&NMItem{Path: strings.Split(order[3].String(), utils.NestingSep),
			Data:   "attrVal5",
			Config: &FCTemplate{Tag: "AttributeTest", AttributeID: "attribute5"}}}); err != nil {
		t.Error(err)
	}
	if err := nM.Set(order[4], &utils.NMSlice{
		&NMItem{Path: strings.Split(order[4].String(), utils.NestingSep),
			Data: "Val4"},
		&NMItem{Path: strings.Split(order[4].String(), utils.NestingSep),
			Data:   "attrVal2",
			Config: &FCTemplate{Tag: "AttributeTest", AttributeID: "attribute2"}}}); err != nil {
		t.Error(err)
	}
	if err := nM.Set(order[5], &utils.NMSlice{
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
					XMLName: xml.Name{Local: order[2][1].String()},
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
									XMLName: xml.Name{Local: order[1][3].String()},
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

/*
func TestNavMapAsCGREvent(t *testing.T) {
	nM := utils.NewOrderedNavigableMap()
	path := []string{"FirstLevel", "SecondLevel", "ThirdLevel", "Fld1"}
	if err := nM.Set(path, &utils.NMSlice{{
		Path: path,
		Data: "Val1",
	}}); err != nil {
		t.Error(err)
	}

	path = []string{"FirstLevel2", "SecondLevel2", "Field2"}
	if err := nM.Set(path, &utils.NMSlice{{
		Path: path,
		Data: "attrVal1",
		Config: &FCTemplate{Tag: "AttributeTest",
			AttributeID: "attribute1"},
	}, {
		Path: path,
		Data: "Value2",
	}}); err != nil {
		t.Error(err)
	}

	path = []string{"FirstLevel2", "Field3"}
	if err := nM.Set(path, &utils.NMSlice{{
		Path: path,
		Data: "Value3",
	}}); err != nil {
		t.Error(err)
	}

	path = []string{"FirstLevel2", "Field5"}
	if err := nM.Set(path, &utils.NMSlice{{
		Path: path,
		Data: "Value5",
	}, {
		Path: path,
		Data: "attrVal5",
		Config: &FCTemplate{Tag: "AttributeTest",
			AttributeID: "attribute5"},
	}}); err != nil {
		t.Error(err)
	}

	path = []string{"FirstLevel2", "Field6"}
	if err := nM.Set(path, &utils.NMSlice{{
		Path: path,
		Data: "Value6",
		Config: &FCTemplate{Tag: "NewBranchTest",
			NewBranch: true},
	}, {
		Path: path,
		Data: "attrVal6",
		Config: &FCTemplate{Tag: "AttributeTest",
			AttributeID: "attribute6"},
	}}); err != nil {
		t.Error(err)
	}

	path = []string{"Field4"}
	if err := nM.Set(path, &utils.NMSlice{{
		Path: path,
		Data: "Val4",
	}, {
		Path: path,
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
*/

func TestNMItemLen(t *testing.T) {
	var nm utils.NM = &NMItem{Data: "1001"}
	if rply := nm.Len(); rply != 0 {
		t.Errorf("Expected 0 ,received: %v", rply)
	}
}

func TestNMItemString(t *testing.T) {
	var nm utils.NM = &NMItem{Data: "1001"}
	expected := "{\"Path\":null,\"Data\":\"1001\",\"Config\":null}"
	if rply := nm.String(); rply != expected {
		t.Errorf("Expected %q ,received: %q", expected, rply)
	}
}

func TestNMItemInterface(t *testing.T) {
	var nm utils.NM = &NMItem{Data: "1001"}
	expected := "1001"
	if rply := nm.Interface(); rply != expected {
		t.Errorf("Expected %q ,received: %q", expected, rply)
	}
}

func TestNMItemField(t *testing.T) {
	var nm utils.NM = &NMItem{Data: "1001"}
	if _, err := nm.Field(nil); err != utils.ErrNotImplemented {
		t.Error(err)
	}
}

func TestNMItemGetField(t *testing.T) {
	var nm utils.NM = &NMItem{Data: "1001"}
	if _, err := nm.GetField(nil); err != utils.ErrNotImplemented {
		t.Error(err)
	}
}

func TestNMItemRemove(t *testing.T) {
	var nm utils.NM = &NMItem{Data: "1001"}
	if err := nm.Remove(nil); err != utils.ErrNotImplemented {
		t.Error(err)
	}
}

func TestNMItemEmpty(t *testing.T) {
	var nm utils.NM = &NMItem{Data: "1001"}
	if nm.Empty() {
		t.Error("Expected not empty type")
	}
	nm = &NMItem{Data: nil}
	if !nm.Empty() {
		t.Error("Expected empty type")
	}
}

func TestNMItemType(t *testing.T) {
	var nm utils.NM = &NMItem{Data: "1001"}
	if nm.Type() != utils.NMInterfaceType {
		t.Errorf("Expected %v ,received: %v", utils.NMInterfaceType, nm.Type())
	}
}

func TestNMItemSet(t *testing.T) {
	var nm utils.NM = &NMItem{Data: "1001"}
	if err := nm.Set(utils.PathItems{{}}, nil); err != utils.ErrWrongPath {
		t.Error(err)
	}
	if err := nm.Set(nil, &NMItem{Data: "1002"}); err != nil {
		t.Error(err)
	}
	expected := "1002"
	if rply := nm.Interface(); rply != expected {
		t.Errorf("Expected %q ,received: %q", expected, rply)
	}
}

func TestNMItemSetField(t *testing.T) {
	var nm utils.NM = &NMItem{Data: "1001"}
	if err := nm.SetField(&utils.PathItem{}, nil); err != utils.ErrNotImplemented {
		t.Error(err)
	}
	// if err := nm.SetField(nil, &NMItem{Data: "1002"}); err != nil {
	// 	t.Error(err)
	// }
	// expected := "1002"
	// if rply := nm.Interface(); rply != expected {
	// 	t.Errorf("Expected %q ,received: %q", expected, rply)
	// }
}
