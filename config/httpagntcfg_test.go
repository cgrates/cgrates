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

func TestHttpAgentCfgsloadFromJsonCfgCase1(t *testing.T) {
	cfgJSON := &[]*HttpAgentJsonCfg{
		{
			Id:              utils.StringPointer("RandomID"),
			Url:             utils.StringPointer("/randomURL"),
			Sessions_conns:  &[]string{"*internal"},
			Reply_payload:   utils.StringPointer(utils.MetaXml),
			Request_payload: utils.StringPointer(utils.MetaUrl),
			Request_processors: &[]*ReqProcessorJsnCfg{
				{
					ID:             utils.StringPointer("OutboundAUTHDryRun"),
					Filters:        &[]string{"*string:*req.request_type:OutboundAUTH", "*string:*req.Msisdn:497700056231"},
					Tenant:         utils.StringPointer("cgrates.org"),
					Flags:          &[]string{utils.MetaDryRun},
					Request_fields: &[]*FcTemplateJsonCfg{},
					Reply_fields: &[]*FcTemplateJsonCfg{
						{
							Tag:       utils.StringPointer("Allow"),
							Path:      utils.StringPointer("response.Allow"),
							Type:      utils.StringPointer(utils.META_CONSTANT),
							Value:     utils.StringPointer("1"),
							Mandatory: utils.BoolPointer(true),
							Layout:    utils.StringPointer("2006-01-02T15:04:05Z07:00"),
						},
					},
				},
			},
		},
	}
	expected := HTTPAgentCfgs{
		{
			ID:             "RandomID",
			URL:            "/randomURL",
			SessionSConns:  []string{"*internal:*sessions"},
			RequestPayload: "*url",
			ReplyPayload:   "*xml",
			RequestProcessors: []*RequestProcessor{{
				ID:            "OutboundAUTHDryRun",
				Filters:       []string{"*string:*req.request_type:OutboundAUTH", "*string:*req.Msisdn:497700056231"},
				Tenant:        NewRSRParsersMustCompile("cgrates.org", utils.INFIELD_SEP),
				Flags:         utils.FlagsWithParams{utils.MetaDryRun: {}},
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
		},
	}
	expected[0].RequestProcessors[0].ReplyFields[0].ComputePath()
	jsnCfg := NewDefaultCGRConfig()
	if err = jsnCfg.httpAgentCfg.loadFromJSONCfg(cfgJSON, jsnCfg.generalCfg.RSRSep); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(&expected, &jsnCfg.httpAgentCfg) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(jsnCfg.httpAgentCfg))
	}
}

func TestHttpAgentCfgsloadFromJsonCfgCase2(t *testing.T) {
	cfgJSON := &[]*HttpAgentJsonCfg{
		{
			Id:              utils.StringPointer("conecto1"),
			Url:             utils.StringPointer("/conecto"),
			Sessions_conns:  &[]string{utils.MetaLocalHost},
			Request_payload: utils.StringPointer(utils.MetaUrl),
			Reply_payload:   utils.StringPointer(utils.MetaXml),
			Request_processors: &[]*ReqProcessorJsnCfg{
				{
					ID:             utils.StringPointer("OutboundAUTHDryRun"),
					Filters:        &[]string{"*string:*req.request_type:OutboundAUTH", "*string:*req.Msisdn:497700056231"},
					Tenant:         utils.StringPointer("cgrates.org"),
					Flags:          &[]string{utils.MetaDryRun},
					Request_fields: &[]*FcTemplateJsonCfg{},
					Reply_fields: &[]*FcTemplateJsonCfg{
						{
							Tag:       utils.StringPointer("Allow"),
							Path:      utils.StringPointer("response.Allow"),
							Type:      utils.StringPointer(utils.META_CONSTANT),
							Value:     utils.StringPointer("1"),
							Mandatory: utils.BoolPointer(true),
							Layout:    utils.StringPointer("2006-01-02T15:04:05Z07:00"),
						},
					},
				},
				{
					ID:      utils.StringPointer("mtcall_cdr"),
					Filters: &[]string{"*string:*req.request_type:MTCALL_CDR"},
					Tenant:  utils.StringPointer("cgrates.org"),
					Flags:   &[]string{utils.MetaCDRs},
					Request_fields: &[]*FcTemplateJsonCfg{
						{
							Tag:       utils.StringPointer("RequestType"),
							Path:      utils.StringPointer("RequestType"),
							Type:      utils.StringPointer(utils.META_CONSTANT),
							Value:     utils.StringPointer("*pseudoprepaid"),
							Mandatory: utils.BoolPointer(true),
							Layout:    utils.StringPointer("2006-01-02T15:04:05Z07:00"),
						},
					},
					Reply_fields: &[]*FcTemplateJsonCfg{
						{
							Tag:       utils.StringPointer("CDR_ID"),
							Path:      utils.StringPointer("CDR_RESPONSE.CDR_ID"),
							Type:      utils.StringPointer(utils.META_COMPOSED),
							Value:     utils.StringPointer("~*req.CDR_ID"),
							Mandatory: utils.BoolPointer(true),
							Layout:    utils.StringPointer("2006-01-02T15:04:05Z07:00"),
						},
					},
				}},
		},
		{
			Id:              utils.StringPointer("conecto_xml"),
			Url:             utils.StringPointer("/conecto_xml"),
			Sessions_conns:  &[]string{utils.MetaLocalHost},
			Request_payload: utils.StringPointer("*xml"),
			Reply_payload:   utils.StringPointer("*xml"),
			Request_processors: &[]*ReqProcessorJsnCfg{
				{
					ID:             utils.StringPointer("cdr_from_xml"),
					Tenant:         utils.StringPointer("cgrates.org"),
					Flags:          &[]string{utils.MetaCDRs},
					Request_fields: &[]*FcTemplateJsonCfg{},
					Reply_fields:   &[]*FcTemplateJsonCfg{},
				},
			},
		},
	}
	expected := HTTPAgentCfgs{
		&HTTPAgentCfg{
			ID:             "conecto1",
			URL:            "/conecto",
			SessionSConns:  []string{utils.MetaLocalHost},
			RequestPayload: utils.MetaUrl,
			ReplyPayload:   utils.MetaXml,
			RequestProcessors: []*RequestProcessor{
				{
					ID:            "OutboundAUTHDryRun",
					Filters:       []string{"*string:*req.request_type:OutboundAUTH", "*string:*req.Msisdn:497700056231"},
					Tenant:        NewRSRParsersMustCompile("cgrates.org", utils.INFIELD_SEP),
					Flags:         utils.FlagsWithParams{utils.MetaDryRun: {}},
					RequestFields: []*FCTemplate{},
					ReplyFields: []*FCTemplate{{
						Tag:       "Allow",
						Path:      "response.Allow",
						Type:      utils.META_CONSTANT,
						Value:     NewRSRParsersMustCompile("1", utils.INFIELD_SEP),
						Mandatory: true,
						Layout:    time.RFC3339,
					}},
				},
				{
					ID:      "mtcall_cdr",
					Filters: []string{"*string:*req.request_type:MTCALL_CDR"},
					Tenant:  NewRSRParsersMustCompile("cgrates.org", utils.INFIELD_SEP),
					Flags:   utils.FlagsWithParams{utils.MetaCDRs: {}},
					RequestFields: []*FCTemplate{{
						Tag:       "RequestType",
						Path:      "RequestType",
						Type:      utils.META_CONSTANT,
						Value:     NewRSRParsersMustCompile("*pseudoprepaid", utils.INFIELD_SEP),
						Mandatory: true,
						Layout:    time.RFC3339,
					}},
					ReplyFields: []*FCTemplate{{
						Tag:       "CDR_ID",
						Path:      "CDR_RESPONSE.CDR_ID",
						Type:      utils.META_COMPOSED,
						Value:     NewRSRParsersMustCompile("~*req.CDR_ID", utils.INFIELD_SEP),
						Mandatory: true,
						Layout:    time.RFC3339,
					}},
				}},
		}, &HTTPAgentCfg{
			ID:             "conecto_xml",
			URL:            "/conecto_xml",
			SessionSConns:  []string{utils.MetaLocalHost},
			RequestPayload: utils.MetaXml,
			ReplyPayload:   utils.MetaXml,
			RequestProcessors: []*RequestProcessor{{
				ID:            "cdr_from_xml",
				Tenant:        NewRSRParsersMustCompile("cgrates.org", utils.INFIELD_SEP),
				Flags:         utils.FlagsWithParams{utils.MetaCDRs: {}},
				RequestFields: []*FCTemplate{},
				ReplyFields:   []*FCTemplate{},
			}},
		}}
	expected[0].RequestProcessors[0].ReplyFields[0].ComputePath()
	expected[0].RequestProcessors[1].ReplyFields[0].ComputePath()
	expected[0].RequestProcessors[1].RequestFields[0].ComputePath()
	cfgJsn := NewDefaultCGRConfig()
	if err = cfgJsn.httpAgentCfg.loadFromJSONCfg(cfgJSON, cfgJsn.generalCfg.RSRSep); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, cfgJsn.httpAgentCfg) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(cfgJsn.httpAgentCfg))
	}
}

func TestHttpAgentCfgloadFromJsonCfgCase3(t *testing.T) {
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
	expected := HTTPAgentCfg{
		ID:             "conecto1",
		URL:            "/conecto",
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
	var httpcfg HTTPAgentCfg
	if err = httpcfg.loadFromJSONCfg(jsnhttpCfg, utils.INFIELD_SEP); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, httpcfg) {
		t.Errorf("Expected: %+v \n, received: %+v", utils.ToJSON(expected), utils.ToJSON(httpcfg))
	}
}

func TestHttpAgentCfgloadFromJsonCfgCase4(t *testing.T) {
	cfgJSON := &[]*HttpAgentJsonCfg{
		{
			Id:              utils.StringPointer("conecto1"),
			Url:             utils.StringPointer("/conecto"),
			Sessions_conns:  &[]string{utils.MetaLocalHost},
			Request_payload: utils.StringPointer(utils.MetaUrl),
			Reply_payload:   utils.StringPointer(utils.MetaXml),
			Request_processors: &[]*ReqProcessorJsnCfg{
				{
					ID:             utils.StringPointer("OutboundAUTHDryRun"),
					Filters:        &[]string{"*string:*req.request_type:OutboundAUTH", "*string:*req.Msisdn:497700056231"},
					Tenant:         utils.StringPointer("cgrates.org"),
					Flags:          &[]string{utils.MetaDryRun},
					Request_fields: &[]*FcTemplateJsonCfg{},
					Reply_fields: &[]*FcTemplateJsonCfg{
						{
							Tag:       utils.StringPointer("Allow"),
							Path:      utils.StringPointer("response.Allow"),
							Type:      utils.StringPointer(utils.META_CONSTANT),
							Value:     utils.StringPointer("1"),
							Mandatory: utils.BoolPointer(true),
							Layout:    utils.StringPointer("2006-01-02T15:04:05Z07:00"),
						},
					},
				},
				{
					ID:      utils.StringPointer("mtcall_cdr"),
					Filters: &[]string{"*string:*req.request_type:MTCALL_CDR"},
					Tenant:  utils.StringPointer("cgrates.org"),
					Flags:   &[]string{utils.MetaCDRs},
					Request_fields: &[]*FcTemplateJsonCfg{
						{
							Tag:       utils.StringPointer("RequestType"),
							Path:      utils.StringPointer("RequestType"),
							Type:      utils.StringPointer(utils.META_CONSTANT),
							Value:     utils.StringPointer("*pseudoprepaid"),
							Mandatory: utils.BoolPointer(true),
							Layout:    utils.StringPointer("2006-01-02T15:04:05Z07:00"),
						},
					},
					Reply_fields: &[]*FcTemplateJsonCfg{
						{
							Tag:       utils.StringPointer("CDR_ID"),
							Path:      utils.StringPointer("CDR_RESPONSE.CDR_ID"),
							Type:      utils.StringPointer(utils.META_COMPOSED),
							Value:     utils.StringPointer("~*req.CDR_ID"),
							Mandatory: utils.BoolPointer(true),
							Layout:    utils.StringPointer("2006-01-02T15:04:05Z07:00"),
						},
					},
				}},
		},
		{
			Id:              utils.StringPointer("conecto_xml"),
			Url:             utils.StringPointer("/conecto_xml"),
			Sessions_conns:  &[]string{utils.MetaLocalHost},
			Request_payload: utils.StringPointer("*xml"),
			Reply_payload:   utils.StringPointer("*xml"),
			Request_processors: &[]*ReqProcessorJsnCfg{
				{
					ID:             utils.StringPointer("cdr_from_xml"),
					Tenant:         utils.StringPointer("a{*"),
					Flags:          &[]string{utils.MetaCDRs},
					Request_fields: &[]*FcTemplateJsonCfg{},
					Reply_fields:   &[]*FcTemplateJsonCfg{},
				},
			},
		},
	}
	expected := "invalid converter terminator in rule: <a{*>"
	jsonCfg := NewDefaultCGRConfig()
	if err = jsonCfg.httpAgentCfg.loadFromJSONCfg(cfgJSON, jsonCfg.generalCfg.RSRSep); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestHttpAgentCfgloadFromJsonCfgCase5(t *testing.T) {
	cfgJSON := &[]*HttpAgentJsonCfg{
		{
			Request_processors: nil,
		},
	}
	jsonCfg := NewDefaultCGRConfig()

	if err := jsonCfg.httpAgentCfg.loadFromJSONCfg(cfgJSON, jsonCfg.generalCfg.RSRSep); err != nil {
		t.Error(err)
	}
}

func TestHttpAgentCfgloadFromJsonCfgCase6(t *testing.T) {
	jsonCfg := NewDefaultCGRConfig()

	httpAgentCfg := new(HTTPAgentCfg)
	if err := httpAgentCfg.loadFromJSONCfg(nil, jsonCfg.generalCfg.RSRSep); err != nil {
		t.Error(err)
	}
}

func TestHttpAgentCfgloadFromJsonCfgCase7(t *testing.T) {
	cfgJSONStr := `{
"http_agent": [
	{
		"id": "RandomID",
		},
	],	
}`
	cfgJSON := &[]*HttpAgentJsonCfg{
		{
			Id: utils.StringPointer("RandomID"),
		},
	}
	expected := HTTPAgentCfgs{
		{
			ID: "RandomID",
		},
	}
	if jsnCfg, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if err = jsnCfg.httpAgentCfg.loadFromJSONCfg(cfgJSON, jsnCfg.generalCfg.RSRSep); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(&expected, &jsnCfg.httpAgentCfg) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(jsnCfg.httpAgentCfg))
	}
}

func TestHttpAgentCfgappendHttpAgntProcCfgs(t *testing.T) {
	initial := &HTTPAgentCfg{
		ID:             "conecto1",
		URL:            "/conecto",
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
	expected := &HTTPAgentCfg{
		ID:             "conecto1",
		URL:            "/conecto",
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
	if err = initial.appendHTTPAgntProcCfgs(proceses, utils.INFIELD_SEP); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, initial) {
		t.Errorf("Expected: %+v , received: %+v", utils.ToJSON(expected), utils.ToJSON(initial))
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
			utils.IDCfg:             "conecto1",
			utils.URLCfg:            "/conecto",
			utils.SessionSConnsCfg:  []string{"*localhost"},
			utils.RequestPayloadCfg: "*url",
			utils.ReplyPayloadCfg:   "*xml",
			utils.RequestProcessorsCfg: []map[string]interface{}{
				{
					utils.IDCfg:            "OutboundAUTHDryRun",
					utils.FiltersCfg:       []string{"*string:~*req.request_type:OutboundAUTH", "*string:~*req.Msisdn:497700056231"},
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
	if cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cgrCfg.httpAgentCfg.AsMapInterface(cgrCfg.generalCfg.RSRSep); !reflect.DeepEqual(eMap, rcv) {
		t.Errorf("Expected %+v, recieved %+v", eMap, rcv)
	}
}

func TestHTTPAgentCfgsClone(t *testing.T) {
	ban := HTTPAgentCfgs{
		{
			ID:             "RandomID",
			URL:            "/randomURL",
			SessionSConns:  []string{"*internal:*sessions", "*conn1"},
			RequestPayload: "*url",
			ReplyPayload:   "*xml",
			RequestProcessors: []*RequestProcessor{{
				ID:            "OutboundAUTHDryRun",
				Filters:       []string{"*string:*req.request_type:OutboundAUTH", "*string:*req.Msisdn:497700056231"},
				Tenant:        NewRSRParsersMustCompile("cgrates.org", utils.INFIELD_SEP),
				Flags:         utils.FlagsWithParams{utils.MetaDryRun: {}},
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
		},
	}
	rcv := ban.Clone()
	if !reflect.DeepEqual(ban, rcv) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(ban), utils.ToJSON(rcv))
	}
	if rcv[0].SessionSConns[1] = ""; ban[0].SessionSConns[1] != "*conn1" {
		t.Errorf("Expected clone to not modify the cloned")
	}
	if rcv[0].RequestProcessors[0].ID = ""; ban[0].RequestProcessors[0].ID != "OutboundAUTHDryRun" {
		t.Errorf("Expected clone to not modify the cloned")
	}
}
