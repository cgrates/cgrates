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
		Id:       utils.StringPointer("Tenant"),
		Type:     utils.StringPointer("*composed"),
		Field_id: utils.StringPointer("Tenant"),
		Filters:  &[]string{"Filter1", "Filter2"},
		Value:    utils.StringPointer("cgrates.org"),
	}
	expected := &FCTemplate{
		ID:      "Tenant",
		Type:    "*composed",
		FieldId: "Tenant",
		Filters: []string{"Filter1", "Filter2"},
		Value:   NewRSRParsersMustCompile("cgrates.org", true),
	}
	rcv := NewFCTemplateFromFCTemplateJsonCfg(jsonCfg)
	if !reflect.DeepEqual(expected, rcv) {
		t.Errorf("expected: %s ,received: %s", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}

func TestFCTemplatesFromFCTemapltesJsonCfg(t *testing.T) {
	jsnCfgs := []*FcTemplateJsonCfg{
		&FcTemplateJsonCfg{
			Id:       utils.StringPointer("Tenant"),
			Type:     utils.StringPointer("*composed"),
			Field_id: utils.StringPointer("Tenant"),
			Filters:  &[]string{"Filter1", "Filter2"},
			Value:    utils.StringPointer("cgrates.org"),
		},
		&FcTemplateJsonCfg{
			Id:       utils.StringPointer("RunID"),
			Type:     utils.StringPointer("*composed"),
			Field_id: utils.StringPointer("RunID"),
			Filters:  &[]string{"Filter1_1", "Filter2_2"},
			Value:    utils.StringPointer("SampleValue"),
		},
	}
	expected := []*FCTemplate{
		&FCTemplate{
			ID:      "Tenant",
			Type:    "*composed",
			FieldId: "Tenant",
			Filters: []string{"Filter1", "Filter2"},
			Value:   NewRSRParsersMustCompile("cgrates.org", true),
		},
		&FCTemplate{
			ID:      "RunID",
			Type:    "*composed",
			FieldId: "RunID",
			Filters: []string{"Filter1_1", "Filter2_2"},
			Value:   NewRSRParsersMustCompile("SampleValue", true),
		},
	}
	rcv := FCTemplatesFromFCTemapltesJsonCfg(jsnCfgs)
	if !reflect.DeepEqual(expected, rcv) {
		t.Errorf("expected: %s ,received: %s", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}
