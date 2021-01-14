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
	"math/rand"
	"reflect"
	"strings"
	"testing"

	"github.com/cgrates/cgrates/utils"
)

func TestNMAsXMLElementsNilNMSlice(t *testing.T) {
	nM := utils.NewOrderedNavigableMap()
	order := []utils.PathItems{
		{{Field: "Field4"}},
	}
	if _, err := nM.Set(&utils.FullPath{
		Path:      order[0].String(),
		PathItems: order[0]},
		&utils.NMSlice{nil}); err != nil {
		t.Error(err)
	}
	expected := "value: Field4[0] is not []*NMItem"
	if _, err := NMAsXMLElements(nM); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestNMAsXMLElementsConfigEmptyID(t *testing.T) {
	nM := utils.NewOrderedNavigableMap()
	order := []utils.PathItems{
		{{Field: "FirstLevel"}, {Field: "SecondLevel"}, {Field: "ThirdLevel"}, {Field: "Fld1"}},
		{{Field: "FirstLevel2"}, {Field: "FirestLevel3"}, {Field: "Field3"}},
		{{Field: "FirstLevel2"}, {Field: "FirestLevel3"}, {Field: "Field5"}},
	}
	if _, err := nM.Set(&utils.FullPath{Path: order[0].String(), PathItems: order[0]}, &utils.NMSlice{
		&NMItem{Path: strings.Split(order[0].String(), utils.NestingSep),
			Data: "Val1"}}); err != nil {
		t.Error(err)
	}
	if _, err := nM.Set(&utils.FullPath{Path: order[1].String(), PathItems: order[1]}, &utils.NMSlice{
		&NMItem{Path: strings.Split(order[1].String(), utils.NestingSep),
			Data: "Value3"}}); err != nil {
		t.Error(err)
	}
	if _, err := nM.Set(&utils.FullPath{Path: order[2].String(), PathItems: order[2]}, &utils.NMSlice{
		&NMItem{Path: strings.Split(order[2].String(), utils.NestingSep),
			Data:   "Value5",
			Config: &FCTemplate{Tag: "AttributeTest", AttributeID: "attribute5"}},
		&NMItem{Path: strings.Split(order[2].String(), utils.NestingSep),
			Data:   "attrVal5",
			Config: &FCTemplate{Tag: "AttributeTest", AttributeID: "attribute5"}}}); err != nil {
		t.Error(err)
	}
	if _, err := NMAsXMLElements(nM); err != nil {
		t.Error(err)
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
		{
			XMLName: xml.Name{Local: order[0][0].String()},
			Elements: []*XMLElement{
				{
					XMLName: xml.Name{Local: order[0][1].String()},
					Elements: []*XMLElement{
						{
							XMLName: xml.Name{Local: order[0][2].String()},
							Attributes: []*xml.Attr{
								{
									Name:  xml.Name{Local: "attribute1"},
									Value: "attrVal1",
								},
							},
							Value: "Value2",
						},
					},
				},
				{
					XMLName: xml.Name{Local: "Field3"},
					Value:   "Value3",
				},
				{
					XMLName: xml.Name{Local: order[3][1].String()},
					Attributes: []*xml.Attr{
						{
							Name:  xml.Name{Local: "attribute5"},
							Value: "attrVal5",
						},
					},
					Value: "Value5",
				},
			},
		},
		{
			XMLName: xml.Name{Local: order[1][0].String()},
			Elements: []*XMLElement{
				{
					XMLName: xml.Name{Local: order[1][1].String()},
					Elements: []*XMLElement{
						{
							XMLName: xml.Name{Local: order[1][2].String()},
							Elements: []*XMLElement{
								{
									XMLName: xml.Name{Local: "Fld1"},
									Value:   "Val1",
								},
							},
						},
					},
				},
			},
		},
		{
			XMLName: xml.Name{Local: order[4][0].String()},
			Attributes: []*xml.Attr{
				{
					Name:  xml.Name{Local: "attribute2"},
					Value: "attrVal2",
				},
			},
			Value: "Val4",
		},
		{
			XMLName: xml.Name{Local: order[5][0].String()},
			Elements: []*XMLElement{
				{
					XMLName: xml.Name{Local: order[5][1].String()},
					Attributes: []*xml.Attr{
						{
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
		utils.NestingSep,nil); cgrEv != nil {
		t.Errorf("expecting: %+v, \nreceived: %+v", utils.ToJSON(nil), utils.ToJSON(cgrEv.Event))
	}

	nM := utils.NewOrderedNavigableMap()
	if cgrEv := NMAsCGREvent(nM, "cgrates.org",
		utils.NestingSep,nil); cgrEv != nil {
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
		utils.NestingSep,utils.NewOrderedNavigableMap()); cgrEv.Tenant != "cgrates.org" ||
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

/*
goos: linux
goarch: amd64
pkg: github.com/cgrates/cgrates/config
BenchmarkOrderdNavigableMapSet2-16    	    1738	   6463443 ns/op
BenchmarkOrderdNavigableMapSet2-16    	    1792	   6536313 ns/op
BenchmarkOrderdNavigableMapSet2-16    	    1744	   6554331 ns/op
BenchmarkNavigableMapOld1Set-16       	    2980	   3831743 ns/op
BenchmarkNavigableMapOld1Set-16       	    2758	   3789885 ns/op
BenchmarkNavigableMapOld1Set-16       	    2916	   3741273 ns/op
PASS
ok  	github.com/cgrates/cgrates/config	71.065s
*/
var generator = rand.New(rand.NewSource(42))
var gen = generateRandomTemplate(10_000)

type benchData struct {
	path      []string
	pathItems utils.PathItems
	strPath   string
	data      string
}

func generateRandomPath() (out []string) {
	size := generator.Intn(16) + 1
	out = make([]string, size)
	for i := 0; i < size; i++ {
		out[i] = utils.Sha1(utils.GenUUID())
	}
	return
}
func generateRandomTemplate(size int) (out []benchData) {
	out = make([]benchData, size)
	for i := 0; i < size; i++ {
		out[i].path = generateRandomPath()
		out[i].data = utils.UUIDSha1Prefix()
		out[i].pathItems = utils.NewPathItems(out[i].path)
		out[i].strPath = out[i].pathItems.String()
		// out[i].pathItems[len(out[i].pathItems)-1].Index = IntPointer(0)
	}
	return
}

func BenchmarkOrderdNavigableMapSet2(b *testing.B) {
	nm := utils.NewOrderedNavigableMap()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		for _, data := range gen {
			if _, err := nm.Set(&utils.FullPath{PathItems: data.pathItems, Path: data.strPath}, utils.NewNMData(data.data)); err != nil {
				b.Log(err, data.path)
			}
		}
	}
}

// func BenchmarkNavigableMapOld1Set(b *testing.B) {
// 	nm := NewNavigableMap(nil)
// 	b.ResetTimer()
// 	for n := 0; n < b.N; n++ {
// 		for _, data := range gen {
// 			nm.Set(data.path, data.data, false, true)
// 		}
// 	}
// }

func TestNMAsMapInterface(t *testing.T) {
	nM := utils.NewOrderedNavigableMap()
	if cgrEv := NMAsMapInterface(nM, utils.NestingSep); !reflect.DeepEqual(cgrEv, map[string]interface{}{}) {
		t.Errorf("expecting: %+v, \nreceived: %+v", utils.ToJSON(map[string]interface{}{}), utils.ToJSON(cgrEv))
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
	if cgrEv := NMAsMapInterface(nM, utils.NestingSep); !reflect.DeepEqual(eEv, cgrEv) {
		t.Errorf("expecting: %+v, \nreceived: %+v", utils.ToJSON(eEv), utils.ToJSON(cgrEv))
	}
}
