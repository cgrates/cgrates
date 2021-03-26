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
	"encoding/xml"
	"math/rand"
	"reflect"
	"strings"
	"testing"
)

func TestNMAsXMLElementsConfigEmptyID(t *testing.T) {
	nM := NewOrderedNavigableMap()
	order := [][]string{
		{"FirstLevel", "SecondLevel", "ThirdLevel", "Fld1"},
		{"FirstLevel2", "FirestLevel3", "Field3"},
		{"FirstLevel2", "FirestLevel3", "Field5"},
	}
	if err := nM.SetAsSlice(&FullPath{Path: strings.Join(order[0], NestingSep), PathSlice: order[0]}, []*DataNode{
		{Type: NMDataType, Value: &DataLeaf{
			Data: "Val1"}}}); err != nil {
		t.Error(err)
	}
	if err := nM.SetAsSlice(&FullPath{Path: strings.Join(order[1], NestingSep), PathSlice: order[1]}, []*DataNode{
		{Type: NMDataType, Value: &DataLeaf{
			Data: "Value3"}}}); err != nil {
		t.Error(err)
	}
	if err := nM.SetAsSlice(&FullPath{Path: strings.Join(order[2], NestingSep), PathSlice: order[2]}, []*DataNode{
		{Type: NMDataType, Value: &DataLeaf{
			Data:        "Value5",
			AttributeID: "attribute5"}},
		{Type: NMDataType, Value: &DataLeaf{
			Data:        "attrVal5",
			AttributeID: "attribute5"}}}); err != nil {
		t.Error(err)
	}
	if _, err := NMAsXMLElements(nM); err != nil {
		t.Error(err)
	}
}

func TestNMAsXMLElements(t *testing.T) {
	nM := NewOrderedNavigableMap()
	order := [][]string{
		{"FirstLevel2", "SecondLevel2", "Field2"},
		{"FirstLevel", "SecondLevel", "ThirdLevel", "Fld1"},
		{"FirstLevel2", "Field3"},
		{"FirstLevel2", "Field5"},
		{"Field4"},
		{"FirstLevel2", "Field6"},
	}
	if err := nM.SetAsSlice(&FullPath{Path: strings.Join(order[0], NestingSep), PathSlice: order[0]}, []*DataNode{
		{Type: NMDataType, Value: &DataLeaf{
			Data:        "attrVal1",
			AttributeID: "attribute1"}},
		{Type: NMDataType, Value: &DataLeaf{
			Data: "Value2"}}}); err != nil {
		t.Error(err)
	}
	if err := nM.SetAsSlice(&FullPath{Path: strings.Join(order[1], NestingSep), PathSlice: order[1]}, []*DataNode{
		{Type: NMDataType, Value: &DataLeaf{
			Data: "Val1"}}}); err != nil {
		t.Error(err)
	}
	if err := nM.SetAsSlice(&FullPath{Path: strings.Join(order[2], NestingSep), PathSlice: order[2]}, []*DataNode{
		{Type: NMDataType, Value: &DataLeaf{
			Data: "Value3"}}}); err != nil {
		t.Error(err)
	}
	if err := nM.SetAsSlice(&FullPath{Path: strings.Join(order[3], NestingSep), PathSlice: order[3]}, []*DataNode{
		{Type: NMDataType, Value: &DataLeaf{
			Data: "Value5"}},
		{Type: NMDataType, Value: &DataLeaf{
			Data:        "attrVal5",
			AttributeID: "attribute5"}}}); err != nil {
		t.Error(err)
	}
	if err := nM.SetAsSlice(&FullPath{Path: strings.Join(order[4], NestingSep), PathSlice: order[4]}, []*DataNode{
		{Type: NMDataType, Value: &DataLeaf{
			Data: "Val4"}},
		{Type: NMDataType, Value: &DataLeaf{
			Data:        "attrVal2",
			AttributeID: "attribute2"}}}); err != nil {
		t.Error(err)
	}
	if err := nM.SetAsSlice(&FullPath{Path: strings.Join(order[5], NestingSep), PathSlice: order[5]}, []*DataNode{
		{Type: NMDataType, Value: &DataLeaf{
			Data:      "Value6",
			NewBranch: true}},
		{Type: NMDataType, Value: &DataLeaf{
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
		t.Errorf("expecting: %s, received: %s", ToJSON(eXMLElmnts), ToJSON(xmlEnts))
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
		out[i] = Sha1(GenUUID())
	}
	return
}
func generateRandomTemplate(size int) (out []benchData) {
	out = make([]benchData, size)
	for i := 0; i < size; i++ {
		out[i].path = generateRandomPath()
		out[i].data = UUIDSha1Prefix()
		out[i].strPath = strings.Join(out[i].path, NestingSep)
	}
	return
}

func BenchmarkOrderdNavigableMapSet2(b *testing.B) {
	nm := NewOrderedNavigableMap()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		for _, data := range gen {
			if err := nm.Set(&FullPath{PathSlice: data.path, Path: data.strPath}, data.data); err != nil {
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
