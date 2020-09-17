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

func TestHttpAgentCfgsloadFromJsonCfg(t *testing.T) {
	var httpcfg, expected HttpAgentCfgs
	if err := httpcfg.loadFromJsonCfg(nil, utils.INFIELD_SEP); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(httpcfg, expected) {
		t.Errorf("Expected: %+v ,recived: %+v", expected, httpcfg)
	}
	if err := httpcfg.loadFromJsonCfg(new([]*HttpAgentJsonCfg), utils.INFIELD_SEP); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(httpcfg, expected) {
		t.Errorf("Expected: %+v ,recived: %+v", expected, httpcfg)
	}
	cfgJSONStr := `{
"http_agent": [
	{
		"id": "conecto1",
		"url": "/conecto",
		"sessions_conns": ["*localhost"],
		"request_payload":	"*url",
		"reply_payload":	"*xml",
		"request_processors": [
			{
				"id": "OutboundAUTHDryRun",
				"filters": ["*string:*req.request_type:OutboundAUTH","*string:*req.Msisdn:497700056231"],
				"tenant": "cgrates.org",
				"flags": ["*dryrun"],
				"request_fields":[
				],
				"reply_fields":[
					{"tag": "Allow", "path": "response.Allow", "type": "*constant", 
						"value": "1", "mandatory": true},
				],
			},
		],
	},
	],
}`
	expected = HttpAgentCfgs{&HttpAgentCfg{
		ID:             "conecto1",
		Url:            "/conecto",
		SessionSConns:  []string{utils.MetaLocalHost},
		RequestPayload: "*url",
		ReplyPayload:   "*xml",
		RequestProcessors: []*RequestProcessor{{
			ID:            "OutboundAUTHDryRun",
			Filters:       []string{"*string:*req.request_type:OutboundAUTH", "*string:*req.Msisdn:497700056231"},
			Tenant:        NewRSRParsersMustCompile("cgrates.org", utils.INFIELD_SEP),
			Flags:         utils.FlagsWithParams{"*dryrun": {}},
			RequestFields: []*FCTemplate{},
			ReplyFields: []*FCTemplate{{
				Tag:       "Allow",
				Path:      "response.Allow",
				Type:      "*constant",
				Value:     NewRSRParsersMustCompile("1", utils.INFIELD_SEP),
				Mandatory: true,
				Layout:    time.RFC3339,
			}},
		}},
	}}
	expected[0].RequestProcessors[0].ReplyFields[0].ComputePath()
	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnhttpCfg, err := jsnCfg.HttpAgentJsonCfg(); err != nil {
		t.Error(err)
	} else if err = httpcfg.loadFromJsonCfg(jsnhttpCfg, utils.INFIELD_SEP); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, httpcfg) {
		t.Errorf("Expected: %+v , recived: %+v", utils.ToJSON(expected), utils.ToJSON(httpcfg))
	}
	cfgJSONStr = `{
"http_agent": [
	{
		"id": "conecto1",
		"url": "/conecto",
		"sessions_conns": ["*localhost"],
		"request_payload":	"*url",
		"reply_payload":	"*xml",
		"request_processors": [
			{
				"id": "mtcall_cdr",
				"filters": ["*string:*req.request_type:MTCALL_CDR"],
				"tenant": "cgrates.org",
				"flags": ["*cdrs"],
				"request_fields":[
					{"tag": "RequestType", "path": "RequestType", "type": "*constant", 
						"value": "*pseudoprepaid", "mandatory": true},	
				],
				"reply_fields":[
					{"tag": "CDR_ID", "path": "CDR_RESPONSE.CDR_ID", "type": "*composed", 
						"value": "~*req.CDR_ID", "mandatory": true},
				],
			
			},
		],
	},
	{
		"id": "conecto_xml",
		"url": "/conecto_xml",
		"sessions_conns": ["*localhost"],
		"request_payload":	"*xml",
		"reply_payload":	"*xml",
		"request_processors": [
			{
				"id": "cdr_from_xml",
				"tenant": "cgrates.org",
				"flags": ["*cdrs"],
				"request_fields":[],
				"reply_fields":[],
			}
		],
	},
	],
}`
	expected = HttpAgentCfgs{
		&HttpAgentCfg{
			ID:             "conecto1",
			Url:            "/conecto",
			SessionSConns:  []string{utils.MetaLocalHost},
			RequestPayload: "*url",
			ReplyPayload:   "*xml",
			RequestProcessors: []*RequestProcessor{
				{
					ID:            "OutboundAUTHDryRun",
					Filters:       []string{"*string:*req.request_type:OutboundAUTH", "*string:*req.Msisdn:497700056231"},
					Tenant:        NewRSRParsersMustCompile("cgrates.org", utils.INFIELD_SEP),
					Flags:         utils.FlagsWithParams{"*dryrun": {}},
					RequestFields: []*FCTemplate{},
					ReplyFields: []*FCTemplate{{
						Tag:       "Allow",
						Path:      "response.Allow",
						Type:      "*constant",
						Value:     NewRSRParsersMustCompile("1", utils.INFIELD_SEP),
						Mandatory: true,
						Layout:    time.RFC3339,
					}},
				},
				{
					ID:      "mtcall_cdr",
					Filters: []string{"*string:*req.request_type:MTCALL_CDR"},
					Tenant:  NewRSRParsersMustCompile("cgrates.org", utils.INFIELD_SEP),
					Flags:   utils.FlagsWithParams{"*cdrs": {}},
					RequestFields: []*FCTemplate{{
						Tag:       "RequestType",
						Path:      "RequestType",
						Type:      "*constant",
						Value:     NewRSRParsersMustCompile("*pseudoprepaid", utils.INFIELD_SEP),
						Mandatory: true,
						Layout:    time.RFC3339,
					}},
					ReplyFields: []*FCTemplate{{
						Tag:       "CDR_ID",
						Path:      "CDR_RESPONSE.CDR_ID",
						Type:      "*composed",
						Value:     NewRSRParsersMustCompile("~*req.CDR_ID", utils.INFIELD_SEP),
						Mandatory: true,
						Layout:    time.RFC3339,
					}},
				}},
		}, &HttpAgentCfg{
			ID:             "conecto_xml",
			Url:            "/conecto_xml",
			SessionSConns:  []string{utils.MetaLocalHost},
			RequestPayload: "*xml",
			ReplyPayload:   "*xml",
			RequestProcessors: []*RequestProcessor{{
				ID:            "cdr_from_xml",
				Tenant:        NewRSRParsersMustCompile("cgrates.org", utils.INFIELD_SEP),
				Flags:         utils.FlagsWithParams{"*cdrs": {}},
				RequestFields: []*FCTemplate{},
				ReplyFields:   []*FCTemplate{},
			}},
		}}
	expected[0].RequestProcessors[0].ReplyFields[0].ComputePath()
	expected[0].RequestProcessors[1].ReplyFields[0].ComputePath()
	expected[0].RequestProcessors[1].RequestFields[0].ComputePath()
	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnhttpCfg, err := jsnCfg.HttpAgentJsonCfg(); err != nil {
		t.Error(err)
	} else if err = httpcfg.loadFromJsonCfg(jsnhttpCfg, utils.INFIELD_SEP); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, httpcfg) {
		t.Errorf("Expected: %+v , recived: %+v", utils.ToJSON(expected), utils.ToJSON(httpcfg))
	}
}

func TestHttpAgentCfgloadFromJsonCfg(t *testing.T) {
	var httpcfg, expected HttpAgentCfg
	if err := httpcfg.loadFromJsonCfg(nil, utils.INFIELD_SEP); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(httpcfg, expected) {
		t.Errorf("Expected: %+v ,recived: %+v", expected, httpcfg)
	}
	if err := httpcfg.loadFromJsonCfg(new(HttpAgentJsonCfg), utils.INFIELD_SEP); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(httpcfg, expected) {
		t.Errorf("Expected: %+v ,recived: %+v", expected, httpcfg)
	}

	jsnhttpCfg := &HttpAgentJsonCfg{
		Id:              utils.StringPointer("conecto1"),
		Url:             utils.StringPointer("/conecto"),
		Sessions_conns:  &[]string{utils.MetaLocalHost},
		Request_payload: utils.StringPointer("*url"),
		Reply_payload:   utils.StringPointer("*xml"),
		Request_processors: &[]*ReqProcessorJsnCfg{
			{
				ID:             utils.StringPointer("OutboundAUTHDryRun"),
				Filters:        &[]string{"*string:*req.request_type:OutboundAUTH", "*string:*req.Msisdn:497700056231"},
				Tenant:         utils.StringPointer("cgrates.org"),
				Flags:          &[]string{"*dryrun"},
				Request_fields: &[]*FcTemplateJsonCfg{},
				Reply_fields:   &[]*FcTemplateJsonCfg{},
			},
		},
	}
	expected = HttpAgentCfg{
		ID:             "conecto1",
		Url:            "/conecto",
		SessionSConns:  []string{utils.MetaLocalHost},
		RequestPayload: "*url",
		ReplyPayload:   "*xml",
		RequestProcessors: []*RequestProcessor{{
			ID:            "OutboundAUTHDryRun",
			Filters:       []string{"*string:*req.request_type:OutboundAUTH", "*string:*req.Msisdn:497700056231"},
			Tenant:        NewRSRParsersMustCompile("cgrates.org", utils.INFIELD_SEP),
			Flags:         utils.FlagsWithParams{"*dryrun": {}},
			RequestFields: []*FCTemplate{},
			ReplyFields:   []*FCTemplate{},
		}},
	}

	if err = httpcfg.loadFromJsonCfg(jsnhttpCfg, utils.INFIELD_SEP); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, httpcfg) {
		t.Errorf("Expected: %+v , recived: %+v", utils.ToJSON(expected), utils.ToJSON(httpcfg))
	}
}

func TestHttpAgentCfgappendHttpAgntProcCfgs(t *testing.T) {
	initial := &HttpAgentCfg{
		ID:             "conecto1",
		Url:            "/conecto",
		SessionSConns:  []string{utils.MetaLocalHost},
		RequestPayload: "*url",
		ReplyPayload:   "*xml",
		RequestProcessors: []*RequestProcessor{{
			ID:            "OutboundAUTHDryRun",
			Filters:       []string{"*string:*req.request_type:OutboundAUTH", "*string:*req.Msisdn:497700056231"},
			Tenant:        NewRSRParsersMustCompile("cgrates.org", utils.INFIELD_SEP),
			Flags:         utils.FlagsWithParams{"*dryrun": {}},
			RequestFields: []*FCTemplate{},
			ReplyFields: []*FCTemplate{{
				Tag:       "Allow",
				Path:      "response.Allow",
				Type:      "*constant",
				Value:     NewRSRParsersMustCompile("1", utils.INFIELD_SEP),
				Mandatory: true,
			}},
		}},
	}
	initial.RequestProcessors[0].ReplyFields[0].ComputePath()
	proceses := &[]*ReqProcessorJsnCfg{{
		ID:             utils.StringPointer("OutboundAUTHDryRun1"),
		Filters:        &[]string{"*string:*req.request_type:OutboundAUTH", "*string:*req.Msisdn:497700056231"},
		Tenant:         utils.StringPointer("cgrates.org"),
		Flags:          &[]string{"*dryrun"},
		Request_fields: &[]*FcTemplateJsonCfg{},
		Reply_fields: &[]*FcTemplateJsonCfg{{
			Tag:       utils.StringPointer("Allow"),
			Path:      utils.StringPointer("response.Allow"),
			Type:      utils.StringPointer("*constant"),
			Value:     utils.StringPointer("1"),
			Mandatory: utils.BoolPointer(true),
		}},
	}, {
		ID:             utils.StringPointer("OutboundAUTHDryRun"),
		Filters:        &[]string{"*string:*req.request_type:OutboundAUTH", "*string:*req.Msisdn:497700056231"},
		Tenant:         utils.StringPointer("cgrates.org"),
		Flags:          &[]string{"*dryrun"},
		Request_fields: &[]*FcTemplateJsonCfg{},
		Reply_fields: &[]*FcTemplateJsonCfg{{
			Tag:       utils.StringPointer("Allow"),
			Path:      utils.StringPointer("response.Allow"),
			Type:      utils.StringPointer("*constant"),
			Value:     utils.StringPointer("1"),
			Mandatory: utils.BoolPointer(false),
		}},
	},
	}
	expected := &HttpAgentCfg{
		ID:             "conecto1",
		Url:            "/conecto",
		SessionSConns:  []string{utils.MetaLocalHost},
		RequestPayload: "*url",
		ReplyPayload:   "*xml",
		RequestProcessors: []*RequestProcessor{{
			ID:            "OutboundAUTHDryRun",
			Filters:       []string{"*string:*req.request_type:OutboundAUTH", "*string:*req.Msisdn:497700056231"},
			Tenant:        NewRSRParsersMustCompile("cgrates.org", utils.INFIELD_SEP),
			Flags:         utils.FlagsWithParams{"*dryrun": {}},
			RequestFields: []*FCTemplate{},
			ReplyFields: []*FCTemplate{{
				Tag:       "Allow",
				Path:      "response.Allow",
				Type:      "*constant",
				Value:     NewRSRParsersMustCompile("1", utils.INFIELD_SEP),
				Mandatory: false,
				Layout:    time.RFC3339,
			}},
		}, {
			ID:            "OutboundAUTHDryRun1",
			Filters:       []string{"*string:*req.request_type:OutboundAUTH", "*string:*req.Msisdn:497700056231"},
			Tenant:        NewRSRParsersMustCompile("cgrates.org", utils.INFIELD_SEP),
			Flags:         utils.FlagsWithParams{"*dryrun": {}},
			RequestFields: []*FCTemplate{},
			ReplyFields: []*FCTemplate{{
				Tag:       "Allow",
				Path:      "response.Allow",
				Type:      "*constant",
				Value:     NewRSRParsersMustCompile("1", utils.INFIELD_SEP),
				Mandatory: true,
				Layout:    time.RFC3339,
			}},
		}},
	}
	expected.RequestProcessors[0].ReplyFields[0].ComputePath()
	expected.RequestProcessors[1].ReplyFields[0].ComputePath()
	if err = initial.appendHttpAgntProcCfgs(proceses, utils.INFIELD_SEP); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, initial) {
		t.Errorf("Expected: %+v , recived: %+v", utils.ToJSON(expected), utils.ToJSON(initial))
	}
}

func TestHttpAgentCfgAsMapInterface(t *testing.T) {
	cfgJSONStr := `{
		
"http_agent": [
	{
		"id": "conecto1",
		"url": "/conecto",
		"sessions_conns": ["*localhost"],
		"request_payload":	"*url",
		"reply_payload":	"*xml",
		"request_processors": [
			{
				"id": "OutboundAUTHDryRun",
				"filters": ["*string:~*req.request_type:OutboundAUTH","*string:~*req.Msisdn:497700056231"],
				"tenant": "cgrates.org",
				"flags": ["*dryrun"],
                "timezone": "",
				"request_fields":[],
				"reply_fields":[
					{"tag": "Allow", "path": "*rep.response.Allow", "type": "*constant",
						"value": "1", "mandatory": true},
					{"tag": "Concatenated1", "path": "*rep.response.Concatenated", "type": "*composed",
                    	"value": "~*req.MCC;/", "mandatory": true},
                    {"tag": "Concatenated2", "path": "*rep.response.Concatenated", "type": "*composed",
                    	"value": "Val1"},
					{"tag": "MaxDuration", "path": "*rep.response.MaxDuration", "type": "*constant",
						"value": "1200", "blocker": true},
					{"tag": "Unused", "path": "*rep.response.Unused", "type": "*constant",
						"value": "0"},
					],
				},
			],
		},
	],	
}`
	eMap := []map[string]interface{}{
		{
			utils.IdCfg:             "conecto1",
			utils.UrlCfg:            "/conecto",
			utils.SessionSConnsCfg:  []string{"*localhost"},
			utils.RequestPayloadCfg: "*url",
			utils.ReplyPayloadCfg:   "*xml",
			utils.RequestProcessorsCfg: []map[string]interface{}{
				{
					utils.IdCfg:            "OutboundAUTHDryRun",
					utils.FilterSCfg:       []string{"*string:~*req.request_type:OutboundAUTH", "*string:~*req.Msisdn:497700056231"},
					utils.TenantCfg:        "cgrates.org",
					utils.FlagsCfg:         []string{"*dryrun"},
					utils.TimezoneCfg:      "",
					utils.RequestFieldsCfg: []map[string]interface{}{},
					utils.ReplyFieldsCfg: []map[string]interface{}{
						{utils.TagCfg: "Allow", utils.PathCfg: "*rep.response.Allow", utils.TypeCfg: "*constant", utils.ValueCfg: "1", utils.MandatoryCfg: true},
						{utils.TagCfg: "Concatenated1", utils.PathCfg: "*rep.response.Concatenated", utils.TypeCfg: "*composed", utils.ValueCfg: "~*req.MCC;/", utils.MandatoryCfg: true},
						{utils.TagCfg: "Concatenated2", utils.PathCfg: "*rep.response.Concatenated", utils.TypeCfg: "*composed", utils.ValueCfg: "Val1"},
						{utils.TagCfg: "MaxDuration", utils.PathCfg: "*rep.response.MaxDuration", utils.TypeCfg: "*constant", utils.ValueCfg: "1200", utils.BlockerCfg: true},
						{utils.TagCfg: "Unused", utils.PathCfg: "*rep.response.Unused", utils.TypeCfg: "*constant", utils.ValueCfg: "0"},
					},
				},
			},
		},
	}
	if cgrCfg, err := NewCGRConfigFromJsonStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cgrCfg.httpAgentCfg.AsMapInterface(cgrCfg.generalCfg.RSRSep); !reflect.DeepEqual(eMap, rcv) {
		t.Errorf("Expected %+v, recieved %+v", eMap, rcv)
	}
}
