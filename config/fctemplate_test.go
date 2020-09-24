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
		Value:   NewRSRParsersMustCompile("cgrates.org", utils.INFIELD_SEP),
		Layout:  time.RFC3339,
	}
	expected.ComputePath()
	if rcv, err := NewFCTemplateFromFCTemplateJsonCfg(jsonCfg, utils.INFIELD_SEP); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rcv) {
		t.Errorf("expected: %s ,received: %s", utils.ToJSON(expected), utils.ToJSON(rcv))
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
			Value:   NewRSRParsersMustCompile("cgrates.org", utils.INFIELD_SEP),
			Layout:  time.RFC3339,
		},
		{
			Tag:     "RunID",
			Type:    "*composed",
			Path:    "RunID",
			Filters: []string{"Filter1_1", "Filter2_2"},
			Value:   NewRSRParsersMustCompile("SampleValue", utils.INFIELD_SEP),
			Layout:  time.RFC3339,
		},
	}
	for _, v := range expected {
		v.ComputePath()
	}
	if rcv, err := FCTemplatesFromFCTemplatesJsonCfg(jsnCfgs, utils.INFIELD_SEP); err != nil {
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
			Value:   NewRSRParsersMustCompile("cgrates.org", utils.INFIELD_SEP),
		},
		{
			Tag:     "RunID",
			Type:    "*composed",
			Path:    "RunID",
			Filters: []string{"Filter1_1", "Filter2_2"},
			Value:   NewRSRParsersMustCompile("SampleValue", utils.INFIELD_SEP),
		},
		{
			Tag:     "TmpMap",
			Type:    "*template",
			Filters: []string{"Filter1_1", "Filter2_2"},
			Value:   NewRSRParsersMustCompile("TmpMap", utils.INFIELD_SEP),
		},
	}
	fcTmpMp := map[string][]*FCTemplate{
		"TmpMap": {
			{
				Tag:     "Elem1",
				Type:    "*composed",
				Path:    "Elem1",
				Filters: []string{"Filter1", "Filter2"},
				Value:   NewRSRParsersMustCompile("Elem1", utils.INFIELD_SEP),
			},
			{
				Tag:     "Elem2",
				Type:    "*composed",
				Path:    "Elem2",
				Filters: []string{"Filter1_1", "Filter2_2"},
				Value:   NewRSRParsersMustCompile("Elem2", utils.INFIELD_SEP),
			},
		},
		"TmpMap2": {
			{
				Tag:     "Elem2.1",
				Type:    "*composed",
				Path:    "Elem2.1",
				Filters: []string{"Filter1", "Filter2"},
				Value:   NewRSRParsersMustCompile("Elem2.1", utils.INFIELD_SEP),
			},
			{
				Tag:     "Elem2.2",
				Type:    "*composed",
				Path:    "Elem2.2",
				Filters: []string{"Filter1_1", "Filter2_2"},
				Value:   NewRSRParsersMustCompile("Elem2.2", utils.INFIELD_SEP),
			},
		},
	}
	expFC := []*FCTemplate{
		{
			Tag:     "Tenant",
			Type:    "*composed",
			Path:    "Tenant",
			Filters: []string{"Filter1", "Filter2"},
			Value:   NewRSRParsersMustCompile("cgrates.org", utils.INFIELD_SEP),
		},
		{
			Tag:     "RunID",
			Type:    "*composed",
			Path:    "RunID",
			Filters: []string{"Filter1_1", "Filter2_2"},
			Value:   NewRSRParsersMustCompile("SampleValue", utils.INFIELD_SEP),
		},
		{
			Tag:     "Elem1",
			Type:    "*composed",
			Path:    "Elem1",
			Filters: []string{"Filter1", "Filter2"},
			Value:   NewRSRParsersMustCompile("Elem1", utils.INFIELD_SEP),
		},
		{
			Tag:     "Elem2",
			Type:    "*composed",
			Path:    "Elem2",
			Filters: []string{"Filter1_1", "Filter2_2"},
			Value:   NewRSRParsersMustCompile("Elem2", utils.INFIELD_SEP),
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
			Value:   NewRSRParsersMustCompile("cgrates.org", utils.INFIELD_SEP),
		},
		{
			Tag:     "RunID",
			Type:    "*composed",
			Path:    "RunID",
			Filters: []string{"Filter1_1", "Filter2_2"},
			Value:   NewRSRParsersMustCompile("SampleValue", utils.INFIELD_SEP),
		},
		{
			Tag:     "TmpMap3",
			Type:    "*template",
			Filters: []string{"Filter1_1", "Filter2_2"},
			Value:   NewRSRParsersMustCompile("TmpMap3", utils.INFIELD_SEP),
		},
	}
	fcTmpMp := map[string][]*FCTemplate{
		"TmpMap": {
			{
				Tag:     "Elem1",
				Type:    "*composed",
				Path:    "Elem1",
				Filters: []string{"Filter1", "Filter2"},
				Value:   NewRSRParsersMustCompile("Elem1", utils.INFIELD_SEP),
			},
			{
				Tag:     "Elem2",
				Type:    "*composed",
				Path:    "Elem2",
				Filters: []string{"Filter1_1", "Filter2_2"},
				Value:   NewRSRParsersMustCompile("Elem2", utils.INFIELD_SEP),
			},
		},
		"TmpMap2": {
			{
				Tag:     "Elem2.1",
				Type:    "*composed",
				Path:    "Elem2.1",
				Filters: []string{"Filter1", "Filter2"},
				Value:   NewRSRParsersMustCompile("Elem2.1", utils.INFIELD_SEP),
			},
			{
				Tag:     "Elem2.2",
				Type:    "*composed",
				Path:    "Elem2.2",
				Filters: []string{"Filter1_1", "Filter2_2"},
				Value:   NewRSRParsersMustCompile("Elem2.2", utils.INFIELD_SEP),
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
			Value:   NewRSRParsersMustCompile("cgrates.org", utils.INFIELD_SEP),
		},
		{
			Tag:     "RunID",
			Type:    "*composed",
			Path:    "RunID",
			Filters: []string{"Filter1_1", "Filter2_2"},
			Value:   NewRSRParsersMustCompile("SampleValue", utils.INFIELD_SEP),
		},
		{
			Tag:     "TmpMap",
			Type:    "*template",
			Filters: []string{"Filter1_1", "Filter2_2"},
			Value:   NewRSRParsersMustCompile("TmpMap", utils.INFIELD_SEP),
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
				Value:   NewRSRParsersMustCompile("Elem2.1", utils.INFIELD_SEP),
			},
			{
				Tag:     "Elem2.2",
				Type:    "*composed",
				Path:    "Elem2.2",
				Filters: []string{"Filter1_1", "Filter2_2"},
				Value:   NewRSRParsersMustCompile("Elem2.2", utils.INFIELD_SEP),
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
		Tag:     "Tenant",
		Type:    "*composed",
		Path:    "Tenant",
		Filters: []string{"Filter1", "Filter2"},
		Value:   NewRSRParsersMustCompile("cgrates.org", utils.INFIELD_SEP),
	}
	smpl.ComputePath()
	cloned := smpl.Clone()
	if !reflect.DeepEqual(cloned, smpl) {
		t.Errorf("expected: %s ,received: %s", utils.ToJSON(smpl), utils.ToJSON(cloned))
	}
	initialSmpl := &FCTemplate{
		Tag:     "Tenant",
		Type:    "*composed",
		Path:    "Tenant",
		Filters: []string{"Filter1", "Filter2"},
		Value:   NewRSRParsersMustCompile("cgrates.org", utils.INFIELD_SEP),
	}
	initialSmpl.ComputePath()
	smpl.Filters = []string{"SingleFilter"}
	smpl.Value = NewRSRParsersMustCompile("cgrates.com", utils.INFIELD_SEP)
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
	if cgrCfg, err := NewCGRConfigFromJsonStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cgrCfg.templates.AsMapInterface(cgrCfg.generalCfg.RSRSep); !reflect.DeepEqual(eMap["custom_template"], rcv["custom_template"]) {
		t.Errorf("Expected %+v \n, recieved %+v", utils.ToJSON(eMap["custom_template"]), utils.ToJSON(rcv["custom_template"]))
	}
}
