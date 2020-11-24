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

func TestSIPAgentCfgloadFromJsonCfgCase1(t *testing.T) {
	cfgJSONS := &SIPAgentJsonCfg{
		Enabled:              utils.BoolPointer(true),
		Listen:               utils.StringPointer("127.0.0.1:5060"),
		Listen_net:           utils.StringPointer("udp"),
		Sessions_conns:       &[]string{utils.MetaInternal},
		Timezone:             utils.StringPointer("local"),
		Retransmission_timer: utils.StringPointer("1"),
		Request_processors: &[]*ReqProcessorJsnCfg{
			{
				ID:             utils.StringPointer("OutboundAUTHDryRun"),
				Filters:        &[]string{"*string:~*req.request_type:OutboundAUTH", "*string:~*req.Msisdn:497700056231"},
				Flags:          &[]string{utils.MetaDryRun},
				Timezone:       utils.StringPointer("local"),
				Request_fields: &[]*FcTemplateJsonCfg{},
				Reply_fields: &[]*FcTemplateJsonCfg{
					{
						Tag:       utils.StringPointer("SessionId"),
						Path:      utils.StringPointer("*rep.Session-Id"),
						Type:      utils.StringPointer(utils.MetaVariable),
						Mandatory: utils.BoolPointer(true),
					},
				},
			},
		},
	}
	expected := &SIPAgentCfg{
		Enabled:             true,
		Listen:              "127.0.0.1:5060",
		ListenNet:           "udp",
		SessionSConns:       []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaSessionS)},
		Timezone:            "local",
		RetransmissionTimer: 1,
		RequestProcessors: []*RequestProcessor{
			{
				ID:            "OutboundAUTHDryRun",
				Filters:       []string{"*string:~*req.request_type:OutboundAUTH", "*string:~*req.Msisdn:497700056231"},
				Flags:         utils.FlagsWithParams{utils.MetaDryRun: {}},
				Timezone:      "local",
				RequestFields: []*FCTemplate{},
				ReplyFields: []*FCTemplate{
					{
						Tag:       "SessionId",
						Path:      "*rep.Session-Id",
						Type:      utils.MetaVariable,
						Mandatory: true,
						Layout:    time.RFC3339,
					},
				},
			},
		},
	}
	for _, r := range expected.RequestProcessors[0].ReplyFields {
		r.ComputePath()
	}
	if jsonCfg, err := NewDefaultCGRConfig(); err != nil {
		t.Error(err)
	} else if err = jsonCfg.sipAgentCfg.loadFromJSONCfg(cfgJSONS, jsonCfg.generalCfg.RSRSep); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, jsonCfg.sipAgentCfg) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(jsonCfg.sipAgentCfg))
	}
}

func TestSIPAgentCfgloadFromJsonCfgCase2(t *testing.T) {
	cfgJSON := &SIPAgentJsonCfg{
		Retransmission_timer: utils.StringPointer("1ss"),
	}
	expected := "time: unknown unit \"ss\" in duration \"1ss\""
	if jsonCfg, err := NewDefaultCGRConfig(); err != nil {
		t.Error(err)
	} else if err = jsonCfg.sipAgentCfg.loadFromJSONCfg(cfgJSON, jsonCfg.generalCfg.RSRSep); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestSIPAgentCfgloadFromJsonCfgCase4(t *testing.T) {
	cfgJSONStr := `{
	"sip_agent": {
		"request_processors": [
             {
               "id": "randomID",
             },
		],
	},
}`
	sipAgent := &SIPAgentJsonCfg{
		Request_processors: &[]*ReqProcessorJsnCfg{{
			ID: utils.StringPointer("randomID"),
		}},
	}
	expected := &SIPAgentCfg{
		RequestProcessors: []*RequestProcessor{
			{
				ID: "randomID",
			},
		},
	}
	if jsonCfg, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if err = jsonCfg.sipAgentCfg.loadFromJSONCfg(sipAgent, jsonCfg.generalCfg.RSRSep); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected.RequestProcessors[0].ID, jsonCfg.sipAgentCfg.RequestProcessors[0].ID) {
		t.Errorf("Expected %+v, received %+v", expected.RequestProcessors[0].ID, jsonCfg.sipAgentCfg.RequestProcessors[0].ID)
	}
}

func TestSIPAgentCfgloadFromJsonCfgCase5(t *testing.T) {
	sipAgent := &SIPAgentJsonCfg{
		Request_processors: &[]*ReqProcessorJsnCfg{{
			Tenant: utils.StringPointer("a{*"),
		}},
	}
	expected := "invalid converter terminator in rule: <a{*>"
	if jsonCfg, err := NewDefaultCGRConfig(); err != nil {
		t.Error(err)
	} else if err = jsonCfg.sipAgentCfg.loadFromJSONCfg(sipAgent, jsonCfg.generalCfg.RSRSep); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestSIPAgentCfgAsMapInterface(t *testing.T) {
	cfgJSONStr := `{
	"sip_agent": {
		"enabled": false,
		"listen": "127.0.0.1:5060",
		"listen_net": "udp",
		"sessions_conns": ["*internal"],
		"timezone": "",
        "retransmission_timer": "2s",
		"request_processors": [
		],
	},
}`
	eMap := map[string]interface{}{
		utils.EnabledCfg:             false,
		utils.ListenCfg:              "127.0.0.1:5060",
		utils.ListenNetCfg:           "udp",
		utils.SessionSConnsCfg:       []string{"*internal"},
		utils.TimezoneCfg:            "",
		utils.RetransmissionTimerCfg: 2 * time.Second,
		utils.RequestProcessorsCfg:   []map[string]interface{}{},
	}
	if cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cgrCfg.sipAgentCfg.AsMapInterface(cgrCfg.generalCfg.RSRSep); !reflect.DeepEqual(eMap, rcv) {
		t.Errorf("Expected %+v, received %+v", eMap, rcv)
	}
}

func TestSIPAgentCfgAsMapInterface1(t *testing.T) {
	cfgJSONStr := `{
		"sip_agent": {
			"enabled": false,
			"listen": "127.0.0.1:5060",
			"listen_net": "udp",
			"sessions_conns": ["*internal"],
			"timezone": "UTC",
            "retransmission_timer": "5s",
			"request_processors": [
			{
				"id": "OutboundAUTHDryRun",
				"filters": ["*string:~*req.request_type:OutboundAUTH","*string:~*req.Msisdn:497700056231"],
				"tenant": "cgrates.org",
				"flags": ["*dryrun"],
                "timezone":       "",
				"request_fields":[
				],
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
	}`
	eMap := map[string]interface{}{
		utils.EnabledCfg:             false,
		utils.ListenCfg:              "127.0.0.1:5060",
		utils.ListenNetCfg:           "udp",
		utils.SessionSConnsCfg:       []string{"*internal"},
		utils.TimezoneCfg:            "UTC",
		utils.RetransmissionTimerCfg: 5 * time.Second,
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
	}
	if cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cgrCfg.sipAgentCfg.AsMapInterface(cgrCfg.generalCfg.RSRSep); !reflect.DeepEqual(rcv, eMap) {
		t.Errorf("Expected %+v \n, received %+v", eMap, rcv)
	}
}

func TestSIPAgentCfgAsMapInterface2(t *testing.T) {
	cfgJSONStr := `{
	"sip_agent": {
		"enabled": true,
		"listen": "",
		"sessions_conns": ["*conn1", "*conn2"],
		"request_processors": [
         {
			"id": "Register",
			"filters": ["*notstring:~*vars.Method:INVITE"],
            "tenant": "cgrates.org",
			"flags": ["*none"],
            "timezone": "",
			"request_fields": [],
			"reply_fields": [
               {"tag": "Request","path": "*rep.Request","type": "*constant","value": "SIP/2.0 405 Method Not Allowed",},
			],
		},
    ],

	}
}`
	eMap := map[string]interface{}{
		utils.EnabledCfg:             true,
		utils.ListenCfg:              "",
		utils.ListenNetCfg:           "udp",
		utils.SessionSConnsCfg:       []string{"*conn1", "*conn2"},
		utils.TimezoneCfg:            "",
		utils.RetransmissionTimerCfg: time.Second,
		utils.RequestProcessorsCfg: []map[string]interface{}{
			{
				utils.IDCfg:            "Register",
				utils.FiltersCfg:       []string{"*notstring:~*vars.Method:INVITE"},
				utils.TenantCfg:        "cgrates.org",
				utils.FlagsCfg:         []string{"*none"},
				utils.TimezoneCfg:      "",
				utils.RequestFieldsCfg: []map[string]interface{}{},
				utils.ReplyFieldsCfg: []map[string]interface{}{
					{utils.TagCfg: "Request", utils.PathCfg: "*rep.Request", utils.TypeCfg: "*constant", utils.ValueCfg: "SIP/2.0 405 Method Not Allowed"},
				},
			},
		},
	}
	if cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cgrCfg.sipAgentCfg.AsMapInterface(cgrCfg.generalCfg.RSRSep); !reflect.DeepEqual(rcv, eMap) {
		t.Errorf("Expected %+v \n, received %+v", eMap, rcv)
	}
}

func TestSIPAgentCfgClone(t *testing.T) {
	sa := &SIPAgentCfg{
		Enabled:             true,
		Listen:              "127.0.0.1:5060",
		ListenNet:           "udp",
		SessionSConns:       []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaSessionS)},
		Timezone:            "UTC",
		RetransmissionTimer: 1,
		RequestProcessors: []*RequestProcessor{
			{
				ID:            "OutboundAUTHDryRun",
				Filters:       []string{"*string:~*req.request_type:OutboundAUTH", "*string:~*req.Msisdn:497700056231"},
				Flags:         utils.FlagsWithParams{utils.MetaDryRun: {}},
				Timezone:      "UTC",
				RequestFields: []*FCTemplate{},
				ReplyFields: []*FCTemplate{
					{
						Tag:       "SessionId",
						Path:      "*rep.Session-Id",
						Type:      utils.MetaVariable,
						Mandatory: true,
						Layout:    time.RFC3339,
					},
				},
			},
		},
	}
	rcv := sa.Clone()
	if !reflect.DeepEqual(sa, rcv) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(sa), utils.ToJSON(rcv))
	}
	if rcv.RequestProcessors[0].ID = ""; sa.RequestProcessors[0].ID != "OutboundAUTHDryRun" {
		t.Errorf("Expected clone to not modify the cloned")
	}
	if rcv.SessionSConns[0] = ""; sa.SessionSConns[0] != utils.ConcatenatedKey(utils.MetaInternal, utils.MetaSessionS) {
		t.Errorf("Expected clone to not modify the cloned")
	}
}
