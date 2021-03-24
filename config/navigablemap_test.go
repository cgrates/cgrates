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

func TestNMAsXMLElementsConfigEmptyID(t *testing.T) {
	nM := utils.NewOrderedNavigableMap()
	order := [][]string{
		{"FirstLevel", "SecondLevel", "ThirdLevel", "Fld1"},
		{"FirstLevel2", "FirestLevel3", "Field3"},
		{"FirstLevel2", "FirestLevel3", "Field5"},
	}
	if err := nM.SetAsSlice(&utils.FullPath{Path: strings.Join(order[0], utils.NestingSep), PathItems: order[0]}, []*utils.DataNode{
		{Type: utils.NMDataType, Value: &utils.DataLeaf{Path: order[0],
			Data: "Val1"}}}); err != nil {
		t.Error(err)
	}
	if err := nM.SetAsSlice(&utils.FullPath{Path: strings.Join(order[1], utils.NestingSep), PathItems: order[1]}, []*utils.DataNode{
		{Type: utils.NMDataType, Value: &utils.DataLeaf{Path: order[1],
			Data: "Value3"}}}); err != nil {
		t.Error(err)
	}
	if err := nM.SetAsSlice(&utils.FullPath{Path: strings.Join(order[2], utils.NestingSep), PathItems: order[2]}, []*utils.DataNode{
		{Type: utils.NMDataType, Value: &utils.DataLeaf{Path: order[2],
			Data:        "Value5",
			AttributeID: "attribute5"}},
		{Type: utils.NMDataType, Value: &utils.DataLeaf{Path: order[2],
			Data:        "attrVal5",
			AttributeID: "attribute5"}}}); err != nil {
		t.Error(err)
	}
	if _, err := NMAsXMLElements(nM); err != nil {
		t.Error(err)
	}
}

func TestNMAsXMLElements(t *testing.T) {
	nM := utils.NewOrderedNavigableMap()
	order := [][]string{
		{"FirstLevel2", "SecondLevel2", "Field2"},
		{"FirstLevel", "SecondLevel", "ThirdLevel", "Fld1"},
		{"FirstLevel2", "Field3"},
		{"FirstLevel2", "Field5"},
		{"Field4"},
		{"FirstLevel2", "Field6"},
	}
	if err := nM.SetAsSlice(&utils.FullPath{Path: strings.Join(order[0], utils.NestingSep), PathItems: order[0]}, []*utils.DataNode{
		{Type: utils.NMDataType, Value: &utils.DataLeaf{Path: order[0],
			Data:        "attrVal1",
			AttributeID: "attribute1"}},
		{Type: utils.NMDataType, Value: &utils.DataLeaf{Path: order[0],
			Data: "Value2"}}}); err != nil {
		t.Error(err)
	}
	if err := nM.SetAsSlice(&utils.FullPath{Path: strings.Join(order[1], utils.NestingSep), PathItems: order[1]}, []*utils.DataNode{
		{Type: utils.NMDataType, Value: &utils.DataLeaf{Path: order[1],
			Data: "Val1"}}}); err != nil {
		t.Error(err)
	}
	if err := nM.SetAsSlice(&utils.FullPath{Path: strings.Join(order[2], utils.NestingSep), PathItems: order[2]}, []*utils.DataNode{
		{Type: utils.NMDataType, Value: &utils.DataLeaf{Path: order[2],
			Data: "Value3"}}}); err != nil {
		t.Error(err)
	}
	if err := nM.SetAsSlice(&utils.FullPath{Path: strings.Join(order[3], utils.NestingSep), PathItems: order[3]}, []*utils.DataNode{
		{Type: utils.NMDataType, Value: &utils.DataLeaf{Path: order[3],
			Data: "Value5"}},
		{Type: utils.NMDataType, Value: &utils.DataLeaf{Path: order[3],
			Data:        "attrVal5",
			AttributeID: "attribute5"}}}); err != nil {
		t.Error(err)
	}
	if err := nM.SetAsSlice(&utils.FullPath{Path: strings.Join(order[4], utils.NestingSep), PathItems: order[4]}, []*utils.DataNode{
		{Type: utils.NMDataType, Value: &utils.DataLeaf{Path: order[4],
			Data: "Val4"}},
		{Type: utils.NMDataType, Value: &utils.DataLeaf{Path: order[4],
			Data:        "attrVal2",
			AttributeID: "attribute2"}}}); err != nil {
		t.Error(err)
	}
	if err := nM.SetAsSlice(&utils.FullPath{Path: strings.Join(order[5], utils.NestingSep), PathItems: order[5]}, []*utils.DataNode{
		{Type: utils.NMDataType, Value: &utils.DataLeaf{Path: order[5],
			Data:      "Value6",
			NewBranch: true}},
		{Type: utils.NMDataType, Value: &utils.DataLeaf{Path: order[5],
			Data:        "attrVal6",
			AttributeID: "attribute6"}}}); err != nil {
		t.Error(err)
	}

	eXMLElmnts := []*XMLElement{
		{
			XMLName: xml.Name{Local: order[0][0]},
			Elements: []*XMLElement{
				{
					XMLName: xml.Name{Local: order[0][1]},
					Elements: []*XMLElement{
						{
							XMLName: xml.Name{Local: order[0][2]},
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
					XMLName: xml.Name{Local: order[3][1]},
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
			XMLName: xml.Name{Local: order[1][0]},
			Elements: []*XMLElement{
				{
					XMLName: xml.Name{Local: order[1][1]},
					Elements: []*XMLElement{
						{
							XMLName: xml.Name{Local: order[1][2]},
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
			XMLName: xml.Name{Local: order[4][0]},
			Attributes: []*xml.Attr{
				{
					Name:  xml.Name{Local: "attribute2"},
					Value: "attrVal2",
				},
			},
			Value: "Val4",
		},
		{
			XMLName: xml.Name{Local: order[5][0]},
			Elements: []*XMLElement{
				{
					XMLName: xml.Name{Local: order[5][1]},
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
		utils.NestingSep, nil); cgrEv != nil {
		t.Errorf("expecting: %+v, \nreceived: %+v", utils.ToJSON(nil), utils.ToJSON(cgrEv.Event))
	}

	nM := utils.NewOrderedNavigableMap()
	if cgrEv := NMAsCGREvent(nM, "cgrates.org",
		utils.NestingSep, nil); cgrEv != nil {
		t.Errorf("expecting: %+v, \nreceived: %+v", utils.ToJSON(nil), utils.ToJSON(cgrEv.Event))
	}

	path := []string{"FirstLevel", "SecondLevel", "ThirdLevel", "Fld1"}
	if err := nM.SetAsSlice(&utils.FullPath{Path: strings.Join(path, utils.NestingSep), PathItems: path}, []*utils.DataNode{
		{Type: utils.NMDataType, Value: &utils.DataLeaf{
			Path: path,
			Data: "Val1",
		}}}); err != nil {
		t.Error(err)
	}

	path = []string{"FirstLevel2", "SecondLevel2", "Field2"}
	if err := nM.SetAsSlice(&utils.FullPath{Path: strings.Join(path, utils.NestingSep), PathItems: path}, []*utils.DataNode{
		{Type: utils.NMDataType, Value: &utils.DataLeaf{
			Path:        path,
			Data:        "attrVal1",
			AttributeID: "attribute1",
		}},
		{Type: utils.NMDataType, Value: &utils.DataLeaf{
			Path: path,
			Data: "Value2",
		}}}); err != nil {
		t.Error(err)
	}

	path = []string{"FirstLevel2", "Field3"}
	if err := nM.SetAsSlice(&utils.FullPath{Path: strings.Join(path, utils.NestingSep), PathItems: path}, []*utils.DataNode{
		{Type: utils.NMDataType, Value: &utils.DataLeaf{
			Path: path,
			Data: "Value3",
		}}}); err != nil {
		t.Error(err)
	}

	path = []string{"FirstLevel2", "Field5"}
	if err := nM.SetAsSlice(&utils.FullPath{Path: strings.Join(path, utils.NestingSep), PathItems: path}, []*utils.DataNode{
		{Type: utils.NMDataType, Value: &utils.DataLeaf{
			Path: path,
			Data: "Value5",
		}},
		{Type: utils.NMDataType, Value: &utils.DataLeaf{
			Path:        path,
			Data:        "attrVal5",
			AttributeID: "attribute5",
		}}}); err != nil {
		t.Error(err)
	}

	path = []string{"FirstLevel2", "Field6"}
	if err := nM.SetAsSlice(&utils.FullPath{Path: strings.Join(path, utils.NestingSep), PathItems: path}, []*utils.DataNode{
		{Type: utils.NMDataType, Value: &utils.DataLeaf{
			Path:      path,
			Data:      "Value6",
			NewBranch: true,
		}},
		{Type: utils.NMDataType, Value: &utils.DataLeaf{
			Path:        path,
			Data:        "attrVal6",
			AttributeID: "attribute6",
		}}}); err != nil {
		t.Error(err)
	}

	path = []string{"Field4"}
	if err := nM.SetAsSlice(&utils.FullPath{Path: strings.Join(path, utils.NestingSep), PathItems: path}, []*utils.DataNode{
		{Type: utils.NMDataType, Value: &utils.DataLeaf{
			Path: path,
			Data: "Val4",
		}},
		{Type: utils.NMDataType, Value: &utils.DataLeaf{
			Path:        path,
			Data:        "attrVal2",
			AttributeID: "attribute2",
		}}}); err != nil {
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
		utils.NestingSep, utils.MapStorage{}); cgrEv.Tenant != "cgrates.org" ||
		cgrEv.Time == nil ||
		!reflect.DeepEqual(eEv, cgrEv.Event) {
		t.Errorf("expecting: %+v, \nreceived: %+v", utils.ToJSON(eEv), utils.ToJSON(cgrEv.Event))
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
	path    []string
	strPath string
	data    string
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
		out[i].strPath = strings.Join(out[i].path, utils.NestingSep)
	}
	return
}

func BenchmarkOrderdNavigableMapSet2(b *testing.B) {
	nm := utils.NewOrderedNavigableMap()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		for _, data := range gen {
			if err := nm.Set(&utils.FullPath{PathItems: data.path, Path: data.strPath}, data.data); err != nil {
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

	path := []string{"FirstLevel", "SecondLevel", "ThirdLevel", "Fld1"}
	if err := nM.SetAsSlice(&utils.FullPath{Path: strings.Join(path, utils.NestingSep), PathItems: path}, []*utils.DataNode{
		{Type: utils.NMDataType, Value: &utils.DataLeaf{
			Path: path,
			Data: "Val1",
		}}}); err != nil {
		t.Error(err)
	}

	path = []string{"FirstLevel2", "SecondLevel2", "Field2"}
	if err := nM.SetAsSlice(&utils.FullPath{Path: strings.Join(path, utils.NestingSep), PathItems: path}, []*utils.DataNode{
		{Type: utils.NMDataType, Value: &utils.DataLeaf{
			Path:        path,
			Data:        "attrVal1",
			AttributeID: "attribute1",
		}},
		{Type: utils.NMDataType, Value: &utils.DataLeaf{
			Path: path,
			Data: "Value2",
		}}}); err != nil {
		t.Error(err)
	}

	path = []string{"FirstLevel2", "Field3"}
	if err := nM.SetAsSlice(&utils.FullPath{Path: strings.Join(path, utils.NestingSep), PathItems: path}, []*utils.DataNode{
		{Type: utils.NMDataType, Value: &utils.DataLeaf{
			Path: path,
			Data: "Value3",
		}}}); err != nil {
		t.Error(err)
	}

	path = []string{"FirstLevel2", "Field5"}
	if err := nM.SetAsSlice(&utils.FullPath{Path: strings.Join(path, utils.NestingSep), PathItems: path}, []*utils.DataNode{
		{Type: utils.NMDataType, Value: &utils.DataLeaf{
			Path: path,
			Data: "Value5",
		}},
		{Type: utils.NMDataType, Value: &utils.DataLeaf{
			Path:        path,
			Data:        "attrVal5",
			AttributeID: "attribute5"},
		}}); err != nil {
		t.Error(err)
	}

	path = []string{"FirstLevel2", "Field6"}
	if err := nM.SetAsSlice(&utils.FullPath{Path: strings.Join(path, utils.NestingSep), PathItems: path}, []*utils.DataNode{
		{Type: utils.NMDataType, Value: &utils.DataLeaf{
			Path:      path,
			Data:      "Value6",
			NewBranch: true},
		},
		{Type: utils.NMDataType, Value: &utils.DataLeaf{
			Path:        path,
			Data:        "attrVal6",
			AttributeID: "attribute6"},
		}}); err != nil {
		t.Error(err)
	}

	path = []string{"Field4"}
	if err := nM.SetAsSlice(&utils.FullPath{Path: strings.Join(path, utils.NestingSep), PathItems: path}, []*utils.DataNode{
		{Type: utils.NMDataType, Value: &utils.DataLeaf{
			Path: path,
			Data: "Val4",
		}},
		{Type: utils.NMDataType, Value: &utils.DataLeaf{
			Path:        path,
			Data:        "attrVal2",
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
