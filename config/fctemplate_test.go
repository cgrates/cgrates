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

	"github.com/cgrates/cgrates/utils"
)

func TestNewFCTemplateFromFCTemplateJsonCfg(t *testing.T) {
	jsonCfg := &FcTemplateJsonCfg{
		Tag:      utils.StringPointer("Tenant"),
		Type:     utils.StringPointer("*composed"),
		Field_id: utils.StringPointer("Tenant"),
		Filters:  &[]string{"Filter1", "Filter2"},
		Value:    utils.StringPointer("cgrates.org"),
	}
	expected := &FCTemplate{
		Tag:     "Tenant",
		Type:    "*composed",
		FieldId: "Tenant",
		Filters: []string{"Filter1", "Filter2"},
		Value:   NewRSRParsersMustCompile("cgrates.org", true, utils.INFIELD_SEP),
	}
	if rcv, err := NewFCTemplateFromFCTemplateJsonCfg(jsonCfg, utils.INFIELD_SEP); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rcv) {
		t.Errorf("expected: %s ,received: %s", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}

func TestFCTemplatesFromFCTemplatesJsonCfg(t *testing.T) {
	jsnCfgs := []*FcTemplateJsonCfg{
		&FcTemplateJsonCfg{
			Tag:      utils.StringPointer("Tenant"),
			Type:     utils.StringPointer("*composed"),
			Field_id: utils.StringPointer("Tenant"),
			Filters:  &[]string{"Filter1", "Filter2"},
			Value:    utils.StringPointer("cgrates.org"),
		},
		&FcTemplateJsonCfg{
			Tag:      utils.StringPointer("RunID"),
			Type:     utils.StringPointer("*composed"),
			Field_id: utils.StringPointer("RunID"),
			Filters:  &[]string{"Filter1_1", "Filter2_2"},
			Value:    utils.StringPointer("SampleValue"),
		},
	}
	expected := []*FCTemplate{
		&FCTemplate{
			Tag:     "Tenant",
			Type:    "*composed",
			FieldId: "Tenant",
			Filters: []string{"Filter1", "Filter2"},
			Value:   NewRSRParsersMustCompile("cgrates.org", true, utils.INFIELD_SEP),
		},
		&FCTemplate{
			Tag:     "RunID",
			Type:    "*composed",
			FieldId: "RunID",
			Filters: []string{"Filter1_1", "Filter2_2"},
			Value:   NewRSRParsersMustCompile("SampleValue", true, utils.INFIELD_SEP),
		},
	}
	if rcv, err := FCTemplatesFromFCTemplatesJsonCfg(jsnCfgs, utils.INFIELD_SEP); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rcv) {
		t.Errorf("expected: %s ,received: %s", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}

func TestFCTemplateInflate1(t *testing.T) {
	fcTmp1 := []*FCTemplate{
		&FCTemplate{
			Tag:     "Tenant",
			Type:    "*composed",
			FieldId: "Tenant",
			Filters: []string{"Filter1", "Filter2"},
			Value:   NewRSRParsersMustCompile("cgrates.org", true, utils.INFIELD_SEP),
		},
		&FCTemplate{
			Tag:     "RunID",
			Type:    "*composed",
			FieldId: "RunID",
			Filters: []string{"Filter1_1", "Filter2_2"},
			Value:   NewRSRParsersMustCompile("SampleValue", true, utils.INFIELD_SEP),
		},
		&FCTemplate{
			Tag:     "TmpMap",
			Type:    "*template",
			Filters: []string{"Filter1_1", "Filter2_2"},
			Value:   NewRSRParsersMustCompile("TmpMap", true, utils.INFIELD_SEP),
		},
	}
	fcTmpMp := map[string][]*FCTemplate{
		"TmpMap": []*FCTemplate{
			&FCTemplate{
				Tag:     "Elem1",
				Type:    "*composed",
				FieldId: "Elem1",
				Filters: []string{"Filter1", "Filter2"},
				Value:   NewRSRParsersMustCompile("Elem1", true, utils.INFIELD_SEP),
			},
			&FCTemplate{
				Tag:     "Elem2",
				Type:    "*composed",
				FieldId: "Elem2",
				Filters: []string{"Filter1_1", "Filter2_2"},
				Value:   NewRSRParsersMustCompile("Elem2", true, utils.INFIELD_SEP),
			},
		},
		"TmpMap2": []*FCTemplate{
			&FCTemplate{
				Tag:     "Elem2.1",
				Type:    "*composed",
				FieldId: "Elem2.1",
				Filters: []string{"Filter1", "Filter2"},
				Value:   NewRSRParsersMustCompile("Elem2.1", true, utils.INFIELD_SEP),
			},
			&FCTemplate{
				Tag:     "Elem2.2",
				Type:    "*composed",
				FieldId: "Elem2.2",
				Filters: []string{"Filter1_1", "Filter2_2"},
				Value:   NewRSRParsersMustCompile("Elem2.2", true, utils.INFIELD_SEP),
			},
		},
	}
	expFC := []*FCTemplate{
		&FCTemplate{
			Tag:     "Tenant",
			Type:    "*composed",
			FieldId: "Tenant",
			Filters: []string{"Filter1", "Filter2"},
			Value:   NewRSRParsersMustCompile("cgrates.org", true, utils.INFIELD_SEP),
		},
		&FCTemplate{
			Tag:     "RunID",
			Type:    "*composed",
			FieldId: "RunID",
			Filters: []string{"Filter1_1", "Filter2_2"},
			Value:   NewRSRParsersMustCompile("SampleValue", true, utils.INFIELD_SEP),
		},
		&FCTemplate{
			Tag:     "Elem1",
			Type:    "*composed",
			FieldId: "Elem1",
			Filters: []string{"Filter1", "Filter2"},
			Value:   NewRSRParsersMustCompile("Elem1", true, utils.INFIELD_SEP),
		},
		&FCTemplate{
			Tag:     "Elem2",
			Type:    "*composed",
			FieldId: "Elem2",
			Filters: []string{"Filter1_1", "Filter2_2"},
			Value:   NewRSRParsersMustCompile("Elem2", true, utils.INFIELD_SEP),
		},
	}
	if rcv, err := InflateTemplates(fcTmp1, fcTmpMp); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expFC, rcv) {
		t.Errorf("expected: %s ,received: %s", utils.ToJSON(expFC), utils.ToJSON(rcv))
	}
}

func TestFCTemplateInflate2(t *testing.T) {
	fcTmp1 := []*FCTemplate{
		&FCTemplate{
			Tag:     "Tenant",
			Type:    "*composed",
			FieldId: "Tenant",
			Filters: []string{"Filter1", "Filter2"},
			Value:   NewRSRParsersMustCompile("cgrates.org", true, utils.INFIELD_SEP),
		},
		&FCTemplate{
			Tag:     "RunID",
			Type:    "*composed",
			FieldId: "RunID",
			Filters: []string{"Filter1_1", "Filter2_2"},
			Value:   NewRSRParsersMustCompile("SampleValue", true, utils.INFIELD_SEP),
		},
		&FCTemplate{
			Tag:     "TmpMap3",
			Type:    "*template",
			Filters: []string{"Filter1_1", "Filter2_2"},
			Value:   NewRSRParsersMustCompile("TmpMap3", true, utils.INFIELD_SEP),
		},
	}
	fcTmpMp := map[string][]*FCTemplate{
		"TmpMap": []*FCTemplate{
			&FCTemplate{
				Tag:     "Elem1",
				Type:    "*composed",
				FieldId: "Elem1",
				Filters: []string{"Filter1", "Filter2"},
				Value:   NewRSRParsersMustCompile("Elem1", true, utils.INFIELD_SEP),
			},
			&FCTemplate{
				Tag:     "Elem2",
				Type:    "*composed",
				FieldId: "Elem2",
				Filters: []string{"Filter1_1", "Filter2_2"},
				Value:   NewRSRParsersMustCompile("Elem2", true, utils.INFIELD_SEP),
			},
		},
		"TmpMap2": []*FCTemplate{
			&FCTemplate{
				Tag:     "Elem2.1",
				Type:    "*composed",
				FieldId: "Elem2.1",
				Filters: []string{"Filter1", "Filter2"},
				Value:   NewRSRParsersMustCompile("Elem2.1", true, utils.INFIELD_SEP),
			},
			&FCTemplate{
				Tag:     "Elem2.2",
				Type:    "*composed",
				FieldId: "Elem2.2",
				Filters: []string{"Filter1_1", "Filter2_2"},
				Value:   NewRSRParsersMustCompile("Elem2.2", true, utils.INFIELD_SEP),
			},
		},
	}
	if _, err := InflateTemplates(fcTmp1, fcTmpMp); err.Error() != "no template with id: <TmpMap3>" {
		t.Error(err)
	}
}

func TestFCTemplateInflate3(t *testing.T) {
	fcTmp1 := []*FCTemplate{
		&FCTemplate{
			Tag:     "Tenant",
			Type:    "*composed",
			FieldId: "Tenant",
			Filters: []string{"Filter1", "Filter2"},
			Value:   NewRSRParsersMustCompile("cgrates.org", true, utils.INFIELD_SEP),
		},
		&FCTemplate{
			Tag:     "RunID",
			Type:    "*composed",
			FieldId: "RunID",
			Filters: []string{"Filter1_1", "Filter2_2"},
			Value:   NewRSRParsersMustCompile("SampleValue", true, utils.INFIELD_SEP),
		},
		&FCTemplate{
			Tag:     "TmpMap",
			Type:    "*template",
			Filters: []string{"Filter1_1", "Filter2_2"},
			Value:   NewRSRParsersMustCompile("TmpMap", true, utils.INFIELD_SEP),
		},
	}
	fcTmpMp := map[string][]*FCTemplate{
		"TmpMap": []*FCTemplate{},
		"TmpMap2": []*FCTemplate{
			&FCTemplate{
				Tag:     "Elem2.1",
				Type:    "*composed",
				FieldId: "Elem2.1",
				Filters: []string{"Filter1", "Filter2"},
				Value:   NewRSRParsersMustCompile("Elem2.1", true, utils.INFIELD_SEP),
			},
			&FCTemplate{
				Tag:     "Elem2.2",
				Type:    "*composed",
				FieldId: "Elem2.2",
				Filters: []string{"Filter1_1", "Filter2_2"},
				Value:   NewRSRParsersMustCompile("Elem2.2", true, utils.INFIELD_SEP),
			},
		},
	}
	if _, err := InflateTemplates(fcTmp1, fcTmpMp); err == nil ||
		err.Error() != "empty template with id: <TmpMap>" {
		t.Error(err)
	}
}
