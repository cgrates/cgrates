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

	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

func TestLoadFromJSONCfgNil(t *testing.T) {
	var jc JanusConn
	err := jc.loadFromJSONCfg(nil)
	if err != nil {
		t.Errorf("Expected %v, received %v", nil, err)
	}

}

func TestJanusAgentLoadFromJSONCfg(t *testing.T) {

	tests := []struct {
		name     string
		jsonCFG  *JanusAgentJsonCfg
		expected *JanusAgentCfg
	}{
		{
			name: "Complete JanusAgentJsonCfg",
			jsonCFG: &JanusAgentJsonCfg{
				Enabled:        utils.BoolPointer(false),
				Sessions_conns: utils.SliceStringPointer([]string{"*internal:*sessions"}),
				Url:            utils.StringPointer("/janus"),
				Janus_conns: &[]*JanusConnJsonCfg{
					{
						Address:       utils.StringPointer("127.0.0.1:8088"),
						AdminAddress:  utils.StringPointer("localhost:7188"),
						AdminPassword: utils.StringPointer(""),
						Type:          utils.StringPointer("*ws"),
					},
				},
				RequestProcessors: &[]*ReqProcessorJsnCfg{
					{
						ID: utils.StringPointer("cgrates"),
					},
				},
			},
			expected: &JanusAgentCfg{
				Enabled:       false,
				URL:           "/janus",
				SessionSConns: []string{"*internal:*sessions"},
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
						ID: "cgrates",
					},
				},
			},
		},
		{
			name: "RequestProcessors not nil",
			jsonCFG: &JanusAgentJsonCfg{
				Enabled:        utils.BoolPointer(false),
				Sessions_conns: utils.SliceStringPointer([]string{"*internal:*sessions"}),
				Url:            utils.StringPointer("/janus"),
				Janus_conns: &[]*JanusConnJsonCfg{
					{
						Address:       utils.StringPointer("127.0.0.1:8088"),
						AdminAddress:  utils.StringPointer("localhost:7188"),
						AdminPassword: utils.StringPointer(""),
						Type:          utils.StringPointer("*ws"),
					},
				},
				RequestProcessors: &[]*ReqProcessorJsnCfg{
					{
						Filters:  utils.SliceStringPointer([]string{}),
						Flags:    utils.SliceStringPointer([]string(nil)),
						ID:       utils.StringPointer("cgrates"),
						Timezone: utils.StringPointer("Local"),
					},
				},
			},
			expected: &JanusAgentCfg{
				Enabled:       false,
				URL:           "/janus",
				SessionSConns: []string{"*internal:*sessions"},
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
						Filters:  []string{},
						Flags:    utils.FlagsWithParams{},
						ID:       "cgrates",
						Timezone: "Local",
					},
				},
			},
		},
		{
			name: "Nil RequestProcessors",
			jsonCFG: &JanusAgentJsonCfg{
				Enabled:        utils.BoolPointer(false),
				Sessions_conns: utils.SliceStringPointer([]string{"*internal:*sessions"}),
				Url:            utils.StringPointer("/janus"),
				Janus_conns: &[]*JanusConnJsonCfg{
					{
						Address:       utils.StringPointer("127.0.0.1:8088"),
						AdminAddress:  utils.StringPointer("localhost:7188"),
						AdminPassword: utils.StringPointer(""),
						Type:          utils.StringPointer("*ws"),
					},
				},
				RequestProcessors: nil,
			},
			expected: &JanusAgentCfg{
				Enabled:       false,
				URL:           "/janus",
				SessionSConns: []string{"*internal:*sessions"},
				JanusConns: []*JanusConn{
					{
						Address:       "127.0.0.1:8088",
						AdminAddress:  "localhost:7188",
						AdminPassword: "",
						Type:          "*ws",
					},
				},
			},
		},
		{
			name: "Without Janus_conns and RequestProcessors",
			jsonCFG: &JanusAgentJsonCfg{
				Enabled:        utils.BoolPointer(false),
				Sessions_conns: utils.SliceStringPointer([]string{"*internal:*sessions"}),
				Url:            utils.StringPointer("/janus"),
			},
			expected: &JanusAgentCfg{
				Enabled:       false,
				URL:           "/janus",
				SessionSConns: []string{"*internal:*sessions"},
				JanusConns: []*JanusConn{
					{
						Address:       "127.0.0.1:8088",
						AdminAddress:  "localhost:7188",
						AdminPassword: "",
						Type:          "*ws",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsnCfg := NewDefaultCGRConfig()

			if err := jsnCfg.janusAgentCfg.loadFromJSONCfg(tt.jsonCFG, jsnCfg.generalCfg.RSRSep); err != nil {
				t.Error(err)
			} else if !reflect.DeepEqual(tt.expected, jsnCfg.janusAgentCfg) {
				t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(tt.expected), utils.ToJSON(jsnCfg.janusAgentCfg))
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
			cfgJSONStr: `{"janus_agent": {
				"enabled": false,
				"url": "/janus",
				"sessions_conns": ["*internal"],
				"janus_conns":[{"address":"127.0.0.1:8088","type":"*ws","admin_address":"localhost:7188","admin_password":""}],
				"request_processors": [],
			},}`,

			eMap: map[string]any{
				utils.EnabledCfg:       false,
				utils.URLCfg:           "/janus",
				utils.SessionSConnsCfg: []string{utils.MetaInternal},
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
			name: "With MetaSessionS",
			cfgJSONStr: `{"janus_agent": {
				"enabled": false,
				"url": "/janus",
				"sessions_conns": ["*sessions"],
				"janus_conns":[{"address":"127.0.0.1:8088","type":"*ws","admin_address":"localhost:7188","admin_password":""}],
				"request_processors": [],
			},}`,

			eMap: map[string]any{
				utils.EnabledCfg:       false,
				utils.URLCfg:           "/janus",
				utils.SessionSConnsCfg: []string{utils.ConcatenatedKey(utils.MetaSessionS)},
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
			name: "With BiRPCInternal",
			cfgJSONStr: `{"janus_agent": {
				"enabled": false,
				"url": "/janus",
				"sessions_conns": ["*birpc_internal"],
				"janus_conns":[{"address":"127.0.0.1:8088","type":"*ws","admin_address":"localhost:7188","admin_password":""}],
				"request_processors": [],
			},}`,

			eMap: map[string]any{
				utils.EnabledCfg:       false,
				utils.URLCfg:           "/janus",
				utils.SessionSConnsCfg: []string{utils.ConcatenatedKey(rpcclient.BiRPCInternal)},
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
			cfgJSONStr: `{"janus_agent": {
				"enabled": false,
				"url": "/janus",
				"sessions_conns": ["*birpc_internal"],
				"janus_conns":[{"address":"127.0.0.1:8088","type":"*ws","admin_address":"localhost:7188","admin_password":""}],
				"request_processors": [
						{
							"filters": [],
							"flags":[],
							"id":  "cgrates",
							"timezone": "Local",
						},
					],
			},}`,

			eMap: map[string]any{
				utils.EnabledCfg:       false,
				utils.URLCfg:           "/janus",
				utils.SessionSConnsCfg: []string{utils.ConcatenatedKey(rpcclient.BiRPCInternal)},
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
			} else if rcv := cgrCfg.janusAgentCfg.AsMapInterface(cgrCfg.generalCfg.RSRSep); !reflect.DeepEqual(rcv, tt.eMap) {
				t.Errorf("Expected %+v, received %+v", utils.ToJSON(tt.eMap), utils.ToJSON(rcv))
			}
		})
	}
}

func TestJanusConnAsMapInterface(t *testing.T) {
	js := &JanusConn{
		Address: "127.001",
		Type:    "ws",
	}
	exp := map[string]any{
		utils.AddressCfg:       "127.001",
		utils.TypeCfg:          "ws",
		utils.AdminAddressCfg:  "",
		utils.AdminPasswordCfg: "",
	}
	val := js.AsMapInterface()
	if !reflect.DeepEqual(val, exp) {
		t.Errorf("expected %+v received %+v", exp, val)
	}

}

func TestJanusConnClone(t *testing.T) {
	tests := []struct {
		name   string
		janusC *JanusConn
	}{
		{
			name: "Complete JanusConn",
			janusC: &JanusConn{
				Address:       "127.0.0.1:8088",
				AdminAddress:  "localhost:7188",
				AdminPassword: "",
				Type:          "*ws",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.janusC.Clone()

			if !reflect.DeepEqual(result, tt.janusC) {
				t.Errorf("Clone() = %v, want %v", result, tt.janusC)
			}

			if result == tt.janusC {
				t.Errorf("Clone returned the same instance, expected a new instance")
			}
		})
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
				Enabled:       false,
				SessionSConns: []string{"*internal:*sessions"},
				URL:           "/janus",
				JanusConns: []*JanusConn{
					{
						Address:       "127.0.0.1:8088",
						Type:          "*ws",
						AdminAddress:  "localhost:7188",
						AdminPassword: "",
					},
				},
				RequestProcessors: nil,
			},
		},
		{
			name: "Nil SessionSConns, JanusConns, RequestProcessors",
			janusAgent: &JanusAgentCfg{
				Enabled:           false,
				SessionSConns:     nil,
				URL:               "/janus",
				JanusConns:        nil,
				RequestProcessors: nil,
			},
		},
		{
			name: "Empty url",
			janusAgent: &JanusAgentCfg{
				Enabled:       false,
				SessionSConns: []string{"*internal:*sessions"},
				URL:           "",
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
						ID:       "cgrates",
						Timezone: "Local",
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

			if result == tt.janusAgent {
				t.Errorf("Clone returned the same instance, expected a new instance")
			}
		})
	}
}

func TestJanusAgentCfgLoadFromJSONReusedProcessors(t *testing.T) {

	jsonCFG1 := &JanusAgentCfg{
		Enabled:       false,
		URL:           "/janus",
		SessionSConns: []string{"*internal:*sessions"},
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
				Filters:  []string{},
				Flags:    utils.FlagsWithParams{},
				ID:       "cgrates",
				Timezone: "Local",
			},
		},
	}

	existing := jsonCFG1.RequestProcessors[0]

	jsonCFG2 := &JanusAgentJsonCfg{
		Enabled:        utils.BoolPointer(false),
		Sessions_conns: utils.SliceStringPointer([]string{"*internal:*sessions"}),
		Url:            utils.StringPointer("/janus"),
		Janus_conns: &[]*JanusConnJsonCfg{
			{
				Address:       utils.StringPointer("127.0.0.1:8088"),
				AdminAddress:  utils.StringPointer("localhost:7188"),
				AdminPassword: utils.StringPointer(""),
				Type:          utils.StringPointer("*ws"),
			},
		},
		RequestProcessors: &[]*ReqProcessorJsnCfg{
			{
				Filters:  utils.SliceStringPointer([]string{}),
				Flags:    utils.SliceStringPointer([]string(nil)),
				ID:       utils.StringPointer("cgrates"),
				Timezone: utils.StringPointer("Local"),
			},
		},
	}

	err := jsonCFG1.loadFromJSONCfg(jsonCFG2, utils.InInFieldSep)
	if err != nil {
		t.Error("Unexpected error")
	}

	if len(jsonCFG1.RequestProcessors) != 1 {
		t.Errorf("Expected 1 RequestProcessor, got %d ", len(jsonCFG1.RequestProcessors))
	}
	if jsonCFG1.RequestProcessors[0] != existing {
		t.Error("Expecting existing RequestProcessors to be reused")
	}
}
