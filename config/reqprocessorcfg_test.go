// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

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
		Tenant: utils.RSRParsers{
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
		Tenant: utils.RSRParsers{
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

	rcv := diffReqProcessorJsnCfg(d, v1, v2)
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
	rcv = diffReqProcessorJsnCfg(d, v1, v2)
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
			Tenant: utils.RSRParsers{
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
			Tenant: utils.RSRParsers{
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

	rcv := diffReqProcessorsJsnCfg(d, v1, v2)
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
	rcv = diffReqProcessorsJsnCfg(d, v1, v2)
	if !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}

	d = nil

	v1 = v2
	expected = &[]*ReqProcessorJsnCfg{
		{},
	}
	rcv = diffReqProcessorsJsnCfg(d, v1, v2)
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
			Tenant: utils.RSRParsers{
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
			Tenant: utils.RSRParsers{
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
