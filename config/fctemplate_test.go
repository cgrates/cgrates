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
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/utils"
)

func TestNewFCTemplateFromFCTemplateJsonCfg(t *testing.T) {
	jsonCfg := &FcTemplateJsonCfg{
		Tag:     utils.StringPointer("Tenant"),
		Type:    utils.StringPointer("*composed"),
		Path:    utils.StringPointer("Tenant"),
		Filters: &[]string{"Filter1", "Filter2"},
		Value:   utils.StringPointer("cgrates.org"),
	}
	expected := &FCTemplate{
		Tag:     "Tenant",
		Type:    "*composed",
		Path:    "Tenant",
		Filters: []string{"Filter1", "Filter2"},
		Value:   NewRSRParsersMustCompile("cgrates.org", utils.InfieldSep),
		Layout:  time.RFC3339,
	}
	expected.ComputePath()
	if rcv, err := NewFCTemplateFromFCTemplateJSONCfg(jsonCfg, utils.InfieldSep); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rcv) {
		t.Errorf("expected: %s ,received: %s", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}

func TestFCTemplateInflateTemplate(t *testing.T) {
	fcTemplate := []*FCTemplate{
		{
			Type:  utils.MetaTemplate,
			Value: NewRSRParsersMustCompile("1sa{*duration}", utils.InfieldSep),
		},
	}
	expected := "time: unknown unit \"sa\" in duration \"1sa\""
	jsonCfg := NewDefaultCGRConfig()
	if _, err := InflateTemplates(fcTemplate, jsonCfg.templates); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestFCTemplatePathItems(t *testing.T) {
	fcTemplate := FCTemplate{
		Path: "*req.Account[1].Balance[*monetary].Value",
	}
	expected := []string{"*req", "Account", "1", "Balance", "*monetary", "Value"}
	fcTemplate.ComputePath()
	if !reflect.DeepEqual(expected, fcTemplate.GetPathItems()) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(expected), utils.ToJSON(fcTemplate))
	}
}

func TestFCTemplatesFromFCTemplatesJsonCfg(t *testing.T) {
	jsnCfgs := []*FcTemplateJsonCfg{
		{
			Tag:     utils.StringPointer("Tenant"),
			Type:    utils.StringPointer("*composed"),
			Path:    utils.StringPointer("Tenant"),
			Filters: &[]string{"Filter1", "Filter2"},
			Value:   utils.StringPointer("cgrates.org"),
		},
		{
			Tag:     utils.StringPointer("RunID"),
			Type:    utils.StringPointer("*composed"),
			Path:    utils.StringPointer("RunID"),
			Filters: &[]string{"Filter1_1", "Filter2_2"},
			Value:   utils.StringPointer("SampleValue"),
		},
	}
	expected := []*FCTemplate{
		{
			Tag:     "Tenant",
			Type:    "*composed",
			Path:    "Tenant",
			Filters: []string{"Filter1", "Filter2"},
			Value:   NewRSRParsersMustCompile("cgrates.org", utils.InfieldSep),
			Layout:  time.RFC3339,
		},
		{
			Tag:     "RunID",
			Type:    "*composed",
			Path:    "RunID",
			Filters: []string{"Filter1_1", "Filter2_2"},
			Value:   NewRSRParsersMustCompile("SampleValue", utils.InfieldSep),
			Layout:  time.RFC3339,
		},
	}
	for _, v := range expected {
		v.ComputePath()
	}
	if rcv, err := FCTemplatesFromFCTemplatesJSONCfg(jsnCfgs, utils.InfieldSep); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rcv) {
		t.Errorf("expected: %s ,received: %s", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}

func TestFCTemplateInflate1(t *testing.T) {
	fcTmp1 := []*FCTemplate{
		{
			Tag:     "Tenant",
			Type:    "*composed",
			Path:    "Tenant",
			Filters: []string{"Filter1", "Filter2"},
			Value:   NewRSRParsersMustCompile("cgrates.org", utils.InfieldSep),
		},
		{
			Tag:     "RunID",
			Type:    "*composed",
			Path:    "RunID",
			Filters: []string{"Filter1_1", "Filter2_2"},
			Value:   NewRSRParsersMustCompile("SampleValue", utils.InfieldSep),
		},
		{
			Tag:     "TmpMap",
			Type:    "*template",
			Filters: []string{"Filter1_1", "Filter2_2"},
			Value:   NewRSRParsersMustCompile("TmpMap", utils.InfieldSep),
		},
	}
	fcTmpMp := map[string][]*FCTemplate{
		"TmpMap": {
			{
				Tag:     "Elem1",
				Type:    "*composed",
				Path:    "Elem1",
				Filters: []string{"Filter1", "Filter2"},
				Value:   NewRSRParsersMustCompile("Elem1", utils.InfieldSep),
			},
			{
				Tag:     "Elem2",
				Type:    "*composed",
				Path:    "Elem2",
				Filters: []string{"Filter1_1", "Filter2_2"},
				Value:   NewRSRParsersMustCompile("Elem2", utils.InfieldSep),
			},
		},
		"TmpMap2": {
			{
				Tag:     "Elem2.1",
				Type:    "*composed",
				Path:    "Elem2.1",
				Filters: []string{"Filter1", "Filter2"},
				Value:   NewRSRParsersMustCompile("Elem2.1", utils.InfieldSep),
			},
			{
				Tag:     "Elem2.2",
				Type:    "*composed",
				Path:    "Elem2.2",
				Filters: []string{"Filter1_1", "Filter2_2"},
				Value:   NewRSRParsersMustCompile("Elem2.2", utils.InfieldSep),
			},
		},
	}
	expFC := []*FCTemplate{
		{
			Tag:     "Tenant",
			Type:    "*composed",
			Path:    "Tenant",
			Filters: []string{"Filter1", "Filter2"},
			Value:   NewRSRParsersMustCompile("cgrates.org", utils.InfieldSep),
		},
		{
			Tag:     "RunID",
			Type:    "*composed",
			Path:    "RunID",
			Filters: []string{"Filter1_1", "Filter2_2"},
			Value:   NewRSRParsersMustCompile("SampleValue", utils.InfieldSep),
		},
		{
			Tag:     "Elem1",
			Type:    "*composed",
			Path:    "Elem1",
			Filters: []string{"Filter1", "Filter2"},
			Value:   NewRSRParsersMustCompile("Elem1", utils.InfieldSep),
		},
		{
			Tag:     "Elem2",
			Type:    "*composed",
			Path:    "Elem2",
			Filters: []string{"Filter1_1", "Filter2_2"},
			Value:   NewRSRParsersMustCompile("Elem2", utils.InfieldSep),
		},
	}
	if rcv, err := InflateTemplates(fcTmp1, fcTmpMp); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expFC, rcv) {
		t.Errorf("expected: %s\n ,received: %s", utils.ToJSON(expFC), utils.ToJSON(rcv))
	}
}

func TestFCTemplateInflate2(t *testing.T) {
	fcTmp1 := []*FCTemplate{
		{
			Tag:     "Tenant",
			Type:    "*composed",
			Path:    "Tenant",
			Filters: []string{"Filter1", "Filter2"},
			Value:   NewRSRParsersMustCompile("cgrates.org", utils.InfieldSep),
		},
		{
			Tag:     "RunID",
			Type:    "*composed",
			Path:    "RunID",
			Filters: []string{"Filter1_1", "Filter2_2"},
			Value:   NewRSRParsersMustCompile("SampleValue", utils.InfieldSep),
		},
		{
			Tag:     "TmpMap3",
			Type:    "*template",
			Filters: []string{"Filter1_1", "Filter2_2"},
			Value:   NewRSRParsersMustCompile("TmpMap3", utils.InfieldSep),
		},
	}
	fcTmpMp := map[string][]*FCTemplate{
		"TmpMap": {
			{
				Tag:     "Elem1",
				Type:    "*composed",
				Path:    "Elem1",
				Filters: []string{"Filter1", "Filter2"},
				Value:   NewRSRParsersMustCompile("Elem1", utils.InfieldSep),
			},
			{
				Tag:     "Elem2",
				Type:    "*composed",
				Path:    "Elem2",
				Filters: []string{"Filter1_1", "Filter2_2"},
				Value:   NewRSRParsersMustCompile("Elem2", utils.InfieldSep),
			},
		},
		"TmpMap2": {
			{
				Tag:     "Elem2.1",
				Type:    "*composed",
				Path:    "Elem2.1",
				Filters: []string{"Filter1", "Filter2"},
				Value:   NewRSRParsersMustCompile("Elem2.1", utils.InfieldSep),
			},
			{
				Tag:     "Elem2.2",
				Type:    "*composed",
				Path:    "Elem2.2",
				Filters: []string{"Filter1_1", "Filter2_2"},
				Value:   NewRSRParsersMustCompile("Elem2.2", utils.InfieldSep),
			},
		},
	}
	if _, err := InflateTemplates(fcTmp1, fcTmpMp); err.Error() != "no template with id: <TmpMap3>" {
		t.Error(err)
	}
}

func TestFCTemplateInflate3(t *testing.T) {
	fcTmp1 := []*FCTemplate{
		{
			Tag:     "Tenant",
			Type:    "*composed",
			Path:    "Tenant",
			Filters: []string{"Filter1", "Filter2"},
			Value:   NewRSRParsersMustCompile("cgrates.org", utils.InfieldSep),
		},
		{
			Tag:     "RunID",
			Type:    "*composed",
			Path:    "RunID",
			Filters: []string{"Filter1_1", "Filter2_2"},
			Value:   NewRSRParsersMustCompile("SampleValue", utils.InfieldSep),
		},
		{
			Tag:     "TmpMap",
			Type:    "*template",
			Filters: []string{"Filter1_1", "Filter2_2"},
			Value:   NewRSRParsersMustCompile("TmpMap", utils.InfieldSep),
		},
	}
	fcTmpMp := map[string][]*FCTemplate{
		"TmpMap": {},
		"TmpMap2": {
			{
				Tag:     "Elem2.1",
				Type:    "*composed",
				Path:    "Elem2.1",
				Filters: []string{"Filter1", "Filter2"},
				Value:   NewRSRParsersMustCompile("Elem2.1", utils.InfieldSep),
			},
			{
				Tag:     "Elem2.2",
				Type:    "*composed",
				Path:    "Elem2.2",
				Filters: []string{"Filter1_1", "Filter2_2"},
				Value:   NewRSRParsersMustCompile("Elem2.2", utils.InfieldSep),
			},
		},
	}
	for _, v := range fcTmp1 {
		v.ComputePath()
	}
	for _, tmpl := range fcTmpMp {
		for _, v := range tmpl {
			v.ComputePath()
		}
	}
	if _, err := InflateTemplates(fcTmp1, fcTmpMp); err == nil ||
		err.Error() != "empty template with id: <TmpMap>" {
		t.Error(err)
	}
}

func TestFCTemplateClone(t *testing.T) {
	smpl := &FCTemplate{
		Tag:              "Tenant",
		Type:             "*composed",
		Path:             "Tenant",
		Filters:          []string{"Filter1", "Filter2"},
		Value:            NewRSRParsersMustCompile("cgrates.org", utils.InfieldSep),
		RoundingDecimals: utils.IntPointer(2),
	}
	smpl.ComputePath()
	cloned := smpl.Clone()
	if !reflect.DeepEqual(cloned, smpl) {
		t.Errorf("expected: %s ,received: %s", utils.ToJSON(smpl), utils.ToJSON(cloned))
	}
	initialSmpl := &FCTemplate{
		Tag:              "Tenant",
		Type:             "*composed",
		Path:             "Tenant",
		Filters:          []string{"Filter1", "Filter2"},
		Value:            NewRSRParsersMustCompile("cgrates.org", utils.InfieldSep),
		RoundingDecimals: utils.IntPointer(2),
	}
	initialSmpl.ComputePath()
	smpl.Filters = []string{"SingleFilter"}
	smpl.Value = NewRSRParsersMustCompile("cgrates.com", utils.InfieldSep)
	if !reflect.DeepEqual(cloned, initialSmpl) {
		t.Errorf("expected: %s ,received: %s", utils.ToJSON(initialSmpl), utils.ToJSON(cloned))
	}
}

func TestFCTemplateAsMapInterface(t *testing.T) {
	cfgJSONStr := `{
     "templates": {
           "custom_template": [
              {
                "tag": "Tenant",
                "type": "*composed",
                "path": "Tenant",
                "filters" : ["val1","val2"],
                "value": "cgrates.org",
                "width": 10,
                "strip": "strp", 
                "padding": "pdding",
                "mandatory": true,
                "attribute_id": "random.val",
                "new_branch": true,
                "timezone": "Local",
                "blocker": true,
                "layout": "",
                "cost_shift_digits": 10,
                "rounding_decimals": 1,
                "mask_destinationd_id": "randomVal",
                "mask_length": 10,
            },
           ],
     }
}`
	eMap := map[string]interface{}{
		"custom_template": []map[string]interface{}{
			{
				utils.TagCfg:              "Tenant",
				utils.TypeCfg:             "*composed",
				utils.PathCfg:             "Tenant",
				utils.FiltersCfg:          []string{"val1", "val2"},
				utils.ValueCfg:            "cgrates.org",
				utils.WidthCfg:            10,
				utils.StripCfg:            "strp",
				utils.PaddingCfg:          "pdding",
				utils.MandatoryCfg:        true,
				utils.AttributeIDCfg:      "random.val",
				utils.NewBranchCfg:        true,
				utils.TimezoneCfg:         "Local",
				utils.BlockerCfg:          true,
				utils.LayoutCfg:           "",
				utils.CostShiftDigitsCfg:  10,
				utils.RoundingDecimalsCfg: 1,
				utils.MaskDestIDCfg:       "randomVal",
				utils.MaskLenCfg:          10,
			},
		},
	}
	if cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cgrCfg.templates.AsMapInterface(cgrCfg.generalCfg.RSRSep); !reflect.DeepEqual(eMap["custom_template"], rcv["custom_template"]) {
		t.Errorf("Expected %+v \n, recieved %+v", utils.ToJSON(eMap["custom_template"]), utils.ToJSON(rcv["custom_template"]))
	}
}

func TestFCTemplateAsMapInterface1(t *testing.T) {
	cfgJSONStr := `{
      "templates": {
        "*err": [
			{"tag": "SessionId", "path": "*rep.Session-Id", "type": "*variable",
				"value": "~*req.Session-Id", "mandatory": true},
			{"tag": "OriginHost", "path": "*rep.Origin-Host", "type": "*variable",
				"value": "~*vars.OriginHost", "mandatory": true},
			{"tag": "OriginRealm", "path": "*rep.Origin-Realm", "type": "*variable",
				"value": "~*vars.OriginRealm", "mandatory": true},
	    ],
         "*asr": [
			{"tag": "SessionId", "path": "*diamreq.Session-Id", "type": "*variable",
				"value": "~*req.Session-Id", "mandatory": true},
			{"tag": "OriginHost", "path": "*diamreq.Origin-Host", "type": "*variable",
				"value": "~*req.Destination-Host", "mandatory": true},
			{"tag": "OriginRealm", "path": "*diamreq.Origin-Realm", "type": "*variable",
				"value": "~*req.Destination-Realm", "mandatory": true},
			{"tag": "DestinationRealm", "path": "*diamreq.Destination-Realm", "type": "*variable",
				"value": "~*req.Origin-Realm", "mandatory": true},
			{"tag": "DestinationHost", "path": "*diamreq.Destination-Host", "type": "*variable",
				"value": "~*req.Origin-Host", "mandatory": true},
			{"tag": "AuthApplicationId", "path": "*diamreq.Auth-Application-Id", "type": "*variable",
				 "value": "~*vars.*appid", "mandatory": true},
	     ],
    }
}`
	eMap := map[string]interface{}{
		utils.MetaErr: []map[string]interface{}{
			{utils.TagCfg: "SessionId", utils.PathCfg: "*rep.Session-Id", utils.TypeCfg: "*variable",
				utils.ValueCfg: "~*req.Session-Id", utils.MandatoryCfg: true},
			{utils.TagCfg: "OriginHost", utils.PathCfg: "*rep.Origin-Host", utils.TypeCfg: "*variable",
				utils.ValueCfg: "~*vars.OriginHost", utils.MandatoryCfg: true},
			{utils.TagCfg: "OriginRealm", utils.PathCfg: "*rep.Origin-Realm", utils.TypeCfg: "*variable",
				utils.ValueCfg: "~*vars.OriginRealm", utils.MandatoryCfg: true},
		},
		utils.MetaASR: []map[string]interface{}{
			{utils.TagCfg: "SessionId", utils.PathCfg: "*diamreq.Session-Id", utils.TypeCfg: "*variable",
				utils.ValueCfg: "~*req.Session-Id", utils.MandatoryCfg: true},
			{utils.TagCfg: "OriginHost", utils.PathCfg: "*diamreq.Origin-Host", utils.TypeCfg: "*variable",
				utils.ValueCfg: "~*req.Destination-Host", utils.MandatoryCfg: true},
			{utils.TagCfg: "OriginRealm", utils.PathCfg: "*diamreq.Origin-Realm", utils.TypeCfg: "*variable",
				utils.ValueCfg: "~*req.Destination-Realm", utils.MandatoryCfg: true},
			{utils.TagCfg: "DestinationRealm", utils.PathCfg: "*diamreq.Destination-Realm", utils.TypeCfg: "*variable",
				utils.ValueCfg: "~*req.Origin-Realm", utils.MandatoryCfg: true},
			{utils.TagCfg: "DestinationHost", utils.PathCfg: "*diamreq.Destination-Host", utils.TypeCfg: "*variable",
				utils.ValueCfg: "~*req.Origin-Host", utils.MandatoryCfg: true},
			{utils.TagCfg: "AuthApplicationId", utils.PathCfg: "*diamreq.Auth-Application-Id", utils.TypeCfg: "*variable",
				utils.ValueCfg: "~*vars.*appid", utils.MandatoryCfg: true},
		},
	}
	if cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else {
		rcv := cgrCfg.templates.AsMapInterface(cgrCfg.generalCfg.RSRSep)
		if !reflect.DeepEqual(eMap[utils.MetaErr], rcv[utils.MetaErr]) {
			t.Errorf("Expected %+v \n, recieved %+v", utils.ToJSON(eMap[utils.MetaErr]), utils.ToJSON(rcv[utils.MetaErr]))
		} else if !reflect.DeepEqual(eMap[utils.MetaASR], rcv[utils.MetaASR]) {
			t.Errorf("Expected %+v \n, recieved %+v", utils.ToJSON(eMap[utils.MetaASR]), utils.ToJSON(rcv[utils.MetaASR]))
		}
	}
}

func TestFCTemplatesClone(t *testing.T) {
	smpl := FcTemplates{
		utils.MetaErr: {{
			Tag:              "Tenant",
			Type:             "*composed",
			Path:             "Tenant",
			Filters:          []string{"Filter1", "Filter2"},
			Value:            NewRSRParsersMustCompile("cgrates.org", utils.InfieldSep),
			RoundingDecimals: utils.IntPointer(2),
		}},
	}
	smpl[utils.MetaErr][0].ComputePath()
	cloned := smpl.Clone()
	if !reflect.DeepEqual(cloned, smpl) {
		t.Errorf("expected: %s ,received: %s", utils.ToJSON(smpl), utils.ToJSON(cloned))
	}
	initialSmpl := FcTemplates{
		utils.MetaErr: {{
			Tag:              "Tenant",
			Type:             "*composed",
			Path:             "Tenant",
			Filters:          []string{"Filter1", "Filter2"},
			Value:            NewRSRParsersMustCompile("cgrates.org", utils.InfieldSep),
			RoundingDecimals: utils.IntPointer(2),
		}},
	}
	initialSmpl[utils.MetaErr][0].ComputePath()
	smpl[utils.MetaErr] = nil
	if !reflect.DeepEqual(cloned, initialSmpl) {
		t.Errorf("expected: %s ,received: %s", utils.ToJSON(initialSmpl), utils.ToJSON(cloned))
	}

	smpl = nil
	cloned = smpl.Clone()
	if !reflect.DeepEqual(cloned, smpl) {
		t.Errorf("expected: %s ,received: %s", utils.ToJSON(smpl), utils.ToJSON(cloned))
	}
}
