/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/

package config

import (
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/utils"
)

func TestJanusAgentCfgLoadFromJSONCfg(t *testing.T) {
	tests := []struct {
		name     string
		jsonCFG  *JanusAgentJsonCfg
		expected *JanusAgentCfg
	}{
		{
			name: "Complete JanusAgentJsonCfg",
			jsonCFG: &JanusAgentJsonCfg{
				Enabled: utils.BoolPointer(false),
				Url:     utils.StringPointer("/janus"),
				Conns: map[string][]*DynamicConns{
					utils.MetaSessionS: {{ConnIDs: []string{"*internal"}}},
				},
				Janus_conns: &[]*JanusConnJsonCfg{
					{
						Address:       utils.StringPointer("127.0.0.1:8088"),
						AdminAddress:  utils.StringPointer("localhost:7188"),
						AdminPassword: utils.StringPointer(""),
						Type:          utils.StringPointer("*ws"),
					},
				},
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
								Type:      utils.StringPointer(utils.MetaConstant),
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
								Type:      utils.StringPointer(utils.MetaConstant),
								Value:     utils.StringPointer("*pseudoprepaid"),
								Mandatory: utils.BoolPointer(true),
								Layout:    utils.StringPointer("2006-01-02T15:04:05Z07:00"),
							},
						},
						Reply_fields: &[]*FcTemplateJsonCfg{
							{
								Tag:       utils.StringPointer("CDR_ID"),
								Path:      utils.StringPointer("CDR_RESPONSE.CDR_ID"),
								Type:      utils.StringPointer(utils.MetaComposed),
								Value:     utils.StringPointer("~*req.CDR_ID"),
								Mandatory: utils.BoolPointer(true),
								Layout:    utils.StringPointer("2006-01-02T15:04:05Z07:00"),
							},
						},
					},
				},
			},
			expected: &JanusAgentCfg{
				Enabled: false,
				URL:     "/janus",
				Conns: map[string][]*DynamicConns{
					utils.MetaSessionS: {{ConnIDs: []string{"*internal:*sessions"}}},
				},
				JanusConns: []*JanusConn{
					{
						Address:       "127.0.0.1:8088",
						AdminAddress:  "localhost:7188",
						AdminPassword: "",
						Type:          "*ws",
					},
				},
				RequestProcessors: []*RequestProcessor{
					{
						ID:            "OutboundAUTHDryRun",
						Filters:       []string{"*string:*req.request_type:OutboundAUTH", "*string:*req.Msisdn:497700056231"},
						Tenant:        utils.NewRSRParsersMustCompile("cgrates.org", utils.InfieldSep),
						Flags:         utils.FlagsWithParams{utils.MetaDryRun: {}},
						RequestFields: []*FCTemplate{},
						ReplyFields: []*FCTemplate{{
							Tag:       "Allow",
							Path:      "response.Allow",
							Type:      utils.MetaConstant,
							Value:     utils.NewRSRParsersMustCompile("1", utils.InfieldSep),
							Mandatory: true,
							Layout:    time.RFC3339,
						}},
					},
					{
						ID:      "mtcall_cdr",
						Filters: []string{"*string:*req.request_type:MTCALL_CDR"},
						Tenant:  utils.NewRSRParsersMustCompile("cgrates.org", utils.InfieldSep),
						Flags:   utils.FlagsWithParams{utils.MetaCDRs: {}},
						RequestFields: []*FCTemplate{{
							Tag:       "RequestType",
							Path:      "RequestType",
							Type:      utils.MetaConstant,
							Value:     utils.NewRSRParsersMustCompile("*pseudoprepaid", utils.InfieldSep),
							Mandatory: true,
							Layout:    time.RFC3339,
						}},
						ReplyFields: []*FCTemplate{{
							Tag:       "CDR_ID",
							Path:      "CDR_RESPONSE.CDR_ID",
							Type:      utils.MetaComposed,
							Value:     utils.NewRSRParsersMustCompile("~*req.CDR_ID", utils.InfieldSep),
							Mandatory: true,
							Layout:    time.RFC3339,
						}},
					},
				},
			},
		},
		{
			name:    "Nil JanusAgentJsonCfg",
			jsonCFG: nil,
			expected: &JanusAgentCfg{
				Enabled: false,
				URL:     "/janus",
				Conns: map[string][]*DynamicConns{
					utils.MetaSessionS: {{ConnIDs: []string{"*internal:*sessions"}}},
				},
				JanusConns: []*JanusConn{
					{
						Address:       "127.0.0.1:8088",
						AdminAddress:  "localhost:7188",
						AdminPassword: "",
						Type:          "*ws",
					},
				},
				RequestProcessors: nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsnCfg := NewDefaultCGRConfig()
			if err := jsnCfg.janusAgentCfg.loadFromJSONCfg(tt.jsonCFG); err != nil {
				t.Error(err)
			} else if !reflect.DeepEqual(utils.ToJSON(tt.expected), utils.ToJSON(jsnCfg.janusAgentCfg)) {
				t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(tt.expected), utils.ToJSON(jsnCfg.janusAgentCfg))
			}
		})
		t.Run("nil janusAgentCfg", func(t *testing.T) {
			jsnCfg := NewDefaultCGRConfig()
			jsnCfg.janusAgentCfg = nil
			if err := jsnCfg.janusAgentCfg.loadFromJSONCfg(tt.jsonCFG); err != nil {
				t.Error(err)
			}
		})
	}
}

func TestJanusAgentAsMapInterface(t *testing.T) {
	tests := []struct {
		name       string
		cfgJSONStr string
		eMap       map[string]any
	}{
		{
			name: "Mapping JanusAgent",
			cfgJSONStr: `{"janusAgent": {
				"enabled": false,
				"url": "/janus",
				"conns": {"*sessions": [{"connIDs": ["*internal"]}]},
				"janusConns": [{
					"address": "127.0.0.1:8088",
					"adminAddress": "localhost:7188",	
					"adminPassword": "",			
					"type": "*ws"
				}],
				"requestProcessors": []
			},}`,

			eMap: map[string]any{
				utils.EnabledCfg: false,
				utils.URLCfg:     "/janus",
				utils.Conns: map[string][]*DynamicConns{
					utils.MetaSessionS: {{ConnIDs: []string{"*internal"}}},
				},
				utils.JanusConnsCfg: []map[string]any{
					{
						utils.AddressCfg:       "127.0.0.1:8088",
						utils.TypeCfg:          "*ws",
						utils.AdminAddressCfg:  "localhost:7188",
						utils.AdminPasswordCfg: "",
					},
				},
				utils.RequestProcessorsCfg: []map[string]any{},
			},
		},
		{
			name: "RequestProcessors with values",
			cfgJSONStr: `{"janusAgent": {
				"enabled": false,
				"url": "/janus",
				"conns": {"*sessions": [{"connIDs": ["*internal"]}]},
				"janusConns": [{
					"address": "127.0.0.1:8088",
					"adminAddress": "localhost:7188",	
					"adminPassword": "",			
					"type": "*ws"
				}],
				"requestProcessors": [
					{
						"filters": [],
						"flags":[],
						"id":  "cgrates",
						"timezone": "Local",
					},
				]
			},}`,

			eMap: map[string]any{
				utils.EnabledCfg: false,
				utils.URLCfg:     "/janus",
				utils.Conns: map[string][]*DynamicConns{
					utils.MetaSessionS: {{ConnIDs: []string{"*internal"}}},
				},
				utils.JanusConnsCfg: []map[string]any{
					{
						utils.AddressCfg:       "127.0.0.1:8088",
						utils.TypeCfg:          "*ws",
						utils.AdminAddressCfg:  "localhost:7188",
						utils.AdminPasswordCfg: "",
					},
				},
				utils.RequestProcessorsCfg: []map[string]any{
					{
						utils.FiltersCfg:  []string{},
						utils.FlagsCfg:    []string(nil),
						utils.IDCfg:       "cgrates",
						utils.TimezoneCfg: "Local",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(tt.cfgJSONStr); err != nil {
				t.Error(err)
			} else if rcv := cgrCfg.janusAgentCfg.AsMapInterface(); !reflect.DeepEqual(rcv, tt.eMap) {
				t.Errorf("Expected %+v, received %+v", utils.ToJSON(tt.eMap), utils.ToJSON(rcv))
			}
		})
	}
}

func TestDiffJanusAgentSJsonCfg(t *testing.T) {
	var d *JanusAgentJsonCfg

	v1 := &JanusAgentCfg{
		Enabled: true,
		URL:     "/janusagent",
		Conns: map[string][]*DynamicConns{
			utils.MetaSessionS: {
				{
					ConnIDs: []string{"*localhost"},
				},
			},
		},
		JanusConns: []*JanusConn{
			{
				Address:       "127.0.0.1:8087",
				AdminAddress:  "localhost:7187",
				AdminPassword: "tst",
				Type:          "*ws",
			},
		},
		RequestProcessors: []*RequestProcessor{
			{
				ID: "reqProcessors",
			},
		},
	}

	v2 := &JanusAgentCfg{
		Enabled: false,
		URL:     "/janus",
		Conns: map[string][]*DynamicConns{
			utils.MetaSessionS: {
				{
					ConnIDs: []string{"*internal"},
				},
			},
		},
		JanusConns: []*JanusConn{
			{
				Address:       "127.0.0.1:8088",
				AdminAddress:  "localhost:7188",
				AdminPassword: "test",
				Type:          "*ws",
			},
		},
		RequestProcessors: []*RequestProcessor{},
	}

	expected := &JanusAgentJsonCfg{
		Enabled: utils.BoolPointer(false),
		Url:     utils.StringPointer("/janus"),
		Conns: map[string][]*DynamicConns{
			utils.MetaSessionS: {
				{
					ConnIDs: []string{"*internal"},
				},
			},
		},
		Janus_conns: &[]*JanusConnJsonCfg{
			{
				Address:       utils.StringPointer("127.0.0.1:8088"),
				Type:          utils.StringPointer("*ws"),
				AdminAddress:  utils.StringPointer("localhost:7188"),
				AdminPassword: utils.StringPointer("test"),
			},
		},
		Request_processors: &[]*ReqProcessorJsnCfg{},
	}

	rcv := diffJanusAgentSJsonCfg(d, v1, v2)
	if !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}

func TestJanusAgentCfgClone(t *testing.T) {
	tests := []struct {
		name       string
		janusAgent *JanusAgentCfg
	}{
		{
			name: "Complete JanusAgentCfg",
			janusAgent: &JanusAgentCfg{
				Enabled: false,
				Conns: map[string][]*DynamicConns{
					utils.MetaSessionS: {{ConnIDs: []string{"*internal"}}},
				},
				URL: "/janus",
				JanusConns: []*JanusConn{
					{
						Address:       "127.0.0.1:8088",
						Type:          "*ws",
						AdminAddress:  "localhost:7188",
						AdminPassword: "",
					},
				},
				RequestProcessors: []*RequestProcessor{
					{
						ID:            "OutboundAUTHDryRun",
						Filters:       []string{"*string:*req.request_type:OutboundAUTH", "*string:*req.Msisdn:497700056231"},
						Tenant:        utils.NewRSRParsersMustCompile("cgrates.org", utils.InfieldSep),
						Flags:         utils.FlagsWithParams{utils.MetaDryRun: {}},
						RequestFields: []*FCTemplate{},
						ReplyFields: []*FCTemplate{
							{
								Tag:       "Allow",
								Path:      "response.Allow",
								Type:      utils.MetaConstant,
								Value:     utils.NewRSRParsersMustCompile("1", utils.InfieldSep),
								Mandatory: true,
								Layout:    time.RFC3339,
							},
						},
					},
				},
			},
		},
		{
			name: "Nil Conns, JanusConns, and RequestProcessors",
			janusAgent: &JanusAgentCfg{
				Enabled:           false,
				Conns:             nil,
				URL:               "/janus",
				JanusConns:        nil,
				RequestProcessors: nil,
			},
		},
		{
			name: "Empty url",
			janusAgent: &JanusAgentCfg{
				Enabled: false,
				Conns: map[string][]*DynamicConns{
					utils.MetaSessionS: {{ConnIDs: []string{"*internal"}}},
				},
				URL: "",
				JanusConns: []*JanusConn{
					{
						Address:       "127.0.0.1:8088",
						Type:          "*ws",
						AdminAddress:  "localhost:7188",
						AdminPassword: "",
					},
				},
				RequestProcessors: []*RequestProcessor{
					{
						ID:            "OutboundAUTHDryRun",
						Filters:       []string{"*string:*req.request_type:OutboundAUTH", "*string:*req.Msisdn:497700056231"},
						Tenant:        utils.NewRSRParsersMustCompile("cgrates.org", utils.InfieldSep),
						Flags:         utils.FlagsWithParams{utils.MetaDryRun: {}},
						RequestFields: []*FCTemplate{},
						ReplyFields: []*FCTemplate{
							{
								Tag:       "Allow",
								Path:      "response.Allow",
								Type:      utils.MetaConstant,
								Value:     utils.NewRSRParsersMustCompile("1", utils.InfieldSep),
								Mandatory: true,
								Layout:    time.RFC3339,
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.janusAgent.Clone()

			if !reflect.DeepEqual(result, tt.janusAgent) {
				t.Errorf("Clone() = %v, want %v", result, tt.janusAgent)
			}

			if tt.janusAgent != nil && result == tt.janusAgent {
				t.Errorf("Clone returned the same instance, expected a new instance")
			}
		})
	}
}
