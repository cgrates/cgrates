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

func TestDiffReqProcessorJsnCfg(t *testing.T) {
	var d *ReqProcessorJsnCfg

	v1 := &RequestProcessor{
		ID:      "req_proc_id1",
		Filters: []string{"filter1"},
		Tenant: RSRParsers{
			{
				Rules: "cgrates.org",
			},
		},
		Timezone: "UTC",
		Flags: utils.FlagsWithParams{
			"FLAG_1": map[string][]string{
				"PARAM_1": {"param_1"},
			},
		},
		RequestFields: []*FCTemplate{
			{
				Type: "type",
				Tag:  "tag",
			},
		},
		ReplyFields: []*FCTemplate{
			{
				Type: "type",
				Tag:  "tag",
			},
		},
	}

	v2 := &RequestProcessor{
		ID:      "req_proc_id2",
		Filters: []string{"filter2"},
		Tenant: RSRParsers{
			{
				Rules: "itsyscom.com",
			},
		},
		Timezone: "Local",
		Flags: utils.FlagsWithParams{
			"FLAG_1": map[string][]string{
				"PARAM_2": {"param_2"},
			},
		},
		RequestFields: []*FCTemplate{
			{
				Type:   "type2",
				Tag:    "tag2",
				Layout: time.RFC3339,
			},
		},
		ReplyFields: []*FCTemplate{
			{
				Type:   "type2",
				Tag:    "tag2",
				Layout: time.RFC3339,
			},
		},
	}

	expected := &ReqProcessorJsnCfg{
		ID:       utils.StringPointer("req_proc_id2"),
		Filters:  &[]string{"filter2"},
		Tenant:   utils.StringPointer("itsyscom.com"),
		Timezone: utils.StringPointer("Local"),
		Flags:    &[]string{"FLAG_1:PARAM_2:param_2"},
		Request_fields: &[]*FcTemplateJsonCfg{
			{
				Type: utils.StringPointer("type2"),
				Tag:  utils.StringPointer("tag2"),
			},
		},
		Reply_fields: &[]*FcTemplateJsonCfg{
			{
				Type: utils.StringPointer("type2"),
				Tag:  utils.StringPointer("tag2"),
			},
		},
	}

	rcv := diffReqProcessorJsnCfg(d, v1, v2, ";")
	if !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}

	d = &ReqProcessorJsnCfg{
		Request_fields: &[]*FcTemplateJsonCfg{
			{
				Type: utils.StringPointer("type2"),
				Tag:  utils.StringPointer("tag2"),
			},
		},
		Reply_fields: &[]*FcTemplateJsonCfg{
			{
				Type: utils.StringPointer("type2"),
				Tag:  utils.StringPointer("tag2"),
			},
		},
	}

	expected = &ReqProcessorJsnCfg{
		Request_fields: &[]*FcTemplateJsonCfg{
			{
				Type: utils.StringPointer("type2"),
				Tag:  utils.StringPointer("tag2"),
			},
		},
		Reply_fields: &[]*FcTemplateJsonCfg{
			{
				Type: utils.StringPointer("type2"),
				Tag:  utils.StringPointer("tag2"),
			},
		},
	}

	v1 = v2
	rcv = diffReqProcessorJsnCfg(d, v1, v2, ";")
	if !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}

func TestDiffReqProcessorsJsnCfg(t *testing.T) {
	var d *[]*ReqProcessorJsnCfg

	v1 := []*RequestProcessor{
		{
			ID:      "req_proc_id1",
			Filters: []string{"filter1"},
			Tenant: RSRParsers{
				{
					Rules: "cgrates.org",
				},
			},
			Timezone: "UTC",
			Flags: utils.FlagsWithParams{
				"FLAG_1": map[string][]string{
					"PARAM_1": {"param_1"},
				},
			},
			RequestFields: []*FCTemplate{
				{
					Type: "type",
					Tag:  "tag",
				},
			},
			ReplyFields: []*FCTemplate{
				{
					Type: "type",
					Tag:  "tag",
				},
			},
		},
	}

	v2 := []*RequestProcessor{
		{
			ID:      "req_proc_id2",
			Filters: []string{"filter2"},
			Tenant: RSRParsers{
				{
					Rules: "itsyscom.com",
				},
			},
			Timezone: "Local",
			Flags: utils.FlagsWithParams{
				"FLAG_1": map[string][]string{
					"PARAM_2": {"param_2"},
				},
			},
			RequestFields: []*FCTemplate{
				{
					Type:   "type2",
					Tag:    "tag2",
					Layout: time.RFC3339,
				},
			},
			ReplyFields: []*FCTemplate{
				{
					Type:   "type2",
					Tag:    "tag2",
					Layout: time.RFC3339,
				},
			},
		},
	}

	expected := &[]*ReqProcessorJsnCfg{
		{
			ID:       utils.StringPointer("req_proc_id2"),
			Filters:  &[]string{"filter2"},
			Tenant:   utils.StringPointer("itsyscom.com"),
			Timezone: utils.StringPointer("Local"),
			Flags:    &[]string{"FLAG_1:PARAM_2:param_2"},
			Request_fields: &[]*FcTemplateJsonCfg{
				{
					Type: utils.StringPointer("type2"),
					Tag:  utils.StringPointer("tag2"),
				},
			},
			Reply_fields: &[]*FcTemplateJsonCfg{
				{
					Type: utils.StringPointer("type2"),
					Tag:  utils.StringPointer("tag2"),
				},
			},
		},
	}

	rcv := diffReqProcessorsJsnCfg(d, v1, v2, ";")
	if !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}

	d = &[]*ReqProcessorJsnCfg{
		{
			ID: utils.StringPointer("req_proc_id2"),
		},
	}
	expected = &[]*ReqProcessorJsnCfg{
		{
			ID:       utils.StringPointer("req_proc_id2"),
			Filters:  &[]string{"filter2"},
			Tenant:   utils.StringPointer("itsyscom.com"),
			Timezone: utils.StringPointer("Local"),
			Flags:    &[]string{"FLAG_1:PARAM_2:param_2"},
			Request_fields: &[]*FcTemplateJsonCfg{
				{
					Type: utils.StringPointer("type2"),
					Tag:  utils.StringPointer("tag2"),
				},
			},
			Reply_fields: &[]*FcTemplateJsonCfg{
				{
					Type: utils.StringPointer("type2"),
					Tag:  utils.StringPointer("tag2"),
				},
			},
		},
	}
	rcv = diffReqProcessorsJsnCfg(d, v1, v2, ";")
	if !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}

	d = nil

	v1 = v2
	expected = &[]*ReqProcessorJsnCfg{
		{},
	}
	rcv = diffReqProcessorsJsnCfg(d, v1, v2, ";")
	if !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}

func TestGetReqProcessorJsnCfg(t *testing.T) {

	d := []*ReqProcessorJsnCfg{
		{
			ID:       utils.StringPointer("req_id"),
			Timezone: utils.StringPointer("Local"),
		},
	}

	expected := &ReqProcessorJsnCfg{
		ID:       utils.StringPointer("req_id"),
		Timezone: utils.StringPointer("Local"),
	}

	rcv, idx := getReqProcessorJsnCfg(d, "req_id")
	if !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Expected %+v \n but recevied \n %+v", expected, rcv)
	} else if idx != 0 {
		t.Errorf("Expected %v \n but recevied \n %v", 0, idx)
	}
}

func TestEqualsRequestProcessors(t *testing.T) {
	v1 := []*RequestProcessor{
		{
			ID:      "req_proc_id1",
			Filters: []string{"filter1"},
			Tenant: RSRParsers{
				{
					Rules: "cgrates.org",
				},
			},
			Timezone: "UTC",
			Flags: utils.FlagsWithParams{
				"FLAG_1": map[string][]string{
					"PARAM_1": {"param_1"},
				},
			},
			RequestFields: []*FCTemplate{
				{
					Type: "type",
					Tag:  "tag",
				},
			},
			ReplyFields: []*FCTemplate{
				{
					Type: "type",
					Tag:  "tag",
				},
			},
		},
	}

	v2 := []*RequestProcessor{
		{
			ID:      "req_proc_id2",
			Filters: []string{"filter2"},
			Tenant: RSRParsers{
				{
					Rules: "itsyscom.com",
				},
			},
			Timezone: "Local",
			Flags: utils.FlagsWithParams{
				"FLAG_1": map[string][]string{
					"PARAM_2": {"param_2"},
				},
			},
			RequestFields: []*FCTemplate{
				{
					Type:   "type2",
					Tag:    "tag2",
					Layout: time.RFC3339,
				},
			},
			ReplyFields: []*FCTemplate{
				{
					Type:   "type2",
					Tag:    "tag2",
					Layout: time.RFC3339,
				},
			},
		},
	}

	if equalsRequestProcessors(v1, v2) {
		t.Error("Reqs should not match")
	}

	v1 = nil

	if equalsRequestProcessors(v1, v2) {
		t.Error("Reqs should not match")
	}
}
