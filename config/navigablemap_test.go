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

/*
func TestNavMapAsXMLElements(t *testing.T) {
	nM := utils.NewOrderedNavigableMap()
	order := [][]string{
		[]string{"FirstLevel2", "SecondLevel2", "Field2"},
		[]string{"FirstLevel", "SecondLevel", "ThirdLevel", "Fld1"},
		[]string{"FirstLevel2", "Field3"},
		[]string{"FirstLevel2", "Field5"},
		[]string{"Field4"},
		[]string{"FirstLevel2", "Field6"},
	}
	if err := nM.Set(order[0], []*NMItem{
		&NMItem{Path: order[0],
			Data:   "attrVal1",
			Config: &FCTemplate{Tag: "AttributeTest", AttributeID: "attribute1"}},
		&NMItem{Path: order[0],
			Data: "Value2"}}); err != nil {
		t.Error(err)
	}
	if err := nM.Set(order[1], []*NMItem{
		&NMItem{Path: order[1],
			Data: "Val1"}}); err != nil {
		t.Error(err)
	}
	if err := nM.Set(order[2], []*NMItem{
		&NMItem{Path: order[2],
			Data: "Value3"}}); err != nil {
		t.Error(err)
	}
	if err := nM.Set(order[3], []*NMItem{
		&NMItem{Path: order[3],
			Data: "Value5"},
		&NMItem{Path: order[3],
			Data:   "attrVal5",
			Config: &FCTemplate{Tag: "AttributeTest", AttributeID: "attribute5"}}}); err != nil {
		t.Error(err)
	}
	if err := nM.Set(order[4], []*NMItem{
		&NMItem{Path: order[4],
			Data: "Val4"},
		&NMItem{Path: order[4],
			Data:   "attrVal2",
			Config: &FCTemplate{Tag: "AttributeTest", AttributeID: "attribute2"}}}); err != nil {
		t.Error(err)
	}
	if err := nM.Set(order[5], []*NMItem{
		&NMItem{Path: order[5],
			Data:   "Value6",
			Config: &FCTemplate{Tag: "NewBranchTest", NewBranch: true}},
		&NMItem{Path: order[5],
			Data:   "attrVal6",
			Config: &FCTemplate{Tag: "AttributeTest", AttributeID: "attribute6"}}}); err != nil {
		t.Error(err)
	}
	eXMLElmnts := []*XMLElement{
		&XMLElement{
			XMLName: xml.Name{Local: order[0][0]},
			Elements: []*XMLElement{
				&XMLElement{
					XMLName: xml.Name{Local: order[0][1]},
					Elements: []*XMLElement{
						&XMLElement{
							XMLName: xml.Name{Local: order[0][2]},
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
					XMLName: xml.Name{Local: order[2][1]},
					Value:   "Value3",
				},
				&XMLElement{
					XMLName: xml.Name{Local: order[3][1]},
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
			XMLName: xml.Name{Local: order[1][0]},
			Elements: []*XMLElement{
				&XMLElement{
					XMLName: xml.Name{Local: order[1][1]},
					Elements: []*XMLElement{
						&XMLElement{
							XMLName: xml.Name{Local: order[1][2]},
							Elements: []*XMLElement{
								&XMLElement{
									XMLName: xml.Name{Local: order[1][3]},
									Value:   "Val1",
								},
							},
						},
					},
				},
			},
		},
		&XMLElement{
			XMLName: xml.Name{Local: order[4][0]},
			Attributes: []*xml.Attr{
				&xml.Attr{
					Name:  xml.Name{Local: "attribute2"},
					Value: "attrVal2",
				},
			},
			Value: "Val4",
		},
		&XMLElement{
			XMLName: xml.Name{Local: order[5][0]},
			Elements: []*XMLElement{
				&XMLElement{
					XMLName: xml.Name{Local: order[5][1]},
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

func TestNavMapAsCGREvent(t *testing.T) {
	nM := utils.NewOrderedNavigableMap()
	path := []string{"FirstLevel", "SecondLevel", "ThirdLevel", "Fld1"}
	if err := nM.Set(path, []*NMItem{{
		Path: path,
		Data: "Val1",
	}}); err != nil {
		t.Error(err)
	}

	path = []string{"FirstLevel2", "SecondLevel2", "Field2"}
	if err := nM.Set(path, []*NMItem{{
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
	if err := nM.Set(path, []*NMItem{{
		Path: path,
		Data: "Value3",
	}}); err != nil {
		t.Error(err)
	}

	path = []string{"FirstLevel2", "Field5"}
	if err := nM.Set(path, []*NMItem{{
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
	if err := nM.Set(path, []*NMItem{{
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
	if err := nM.Set(path, []*NMItem{{
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
