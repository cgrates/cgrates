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

func TestRadiusAgentCfgloadFromJsonCfgCase1(t *testing.T) {
	cfgJSON := &RadiusAgentJsonCfg{
		Enabled:             utils.BoolPointer(true),
		Listen_net:          utils.StringPointer(utils.UDP),
		Listen_auth:         utils.StringPointer("127.0.0.1:1812"),
		Listen_acct:         utils.StringPointer("127.0.0.1:1813"),
		Client_secrets:      &map[string]string{utils.MetaDefault: "CGRateS.org"},
		Client_dictionaries: &map[string]string{utils.MetaDefault: "/usr/share/cgrates/radius/dict/"},
		Sessions_conns:      &[]string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaSessionS)},
		Request_processors: &[]*ReqProcessorJsnCfg{
			{
				ID:             utils.StringPointer("OutboundAUTHDryRun"),
				Filters:        &[]string{"*string:~*req.request_type:OutboundAUTH", "*string:~*req.Msisdn:497700056231"},
				Flags:          &[]string{utils.MetaDryRun},
				Timezone:       utils.StringPointer(utils.EmptyString),
				Tenant:         utils.StringPointer("~*req.CGRID"),
				Request_fields: &[]*FcTemplateJsonCfg{},
				Reply_fields: &[]*FcTemplateJsonCfg{
					{
						Tag:       utils.StringPointer("Allow"),
						Path:      utils.StringPointer("*rep.response.Allow"),
						Type:      utils.StringPointer(utils.META_CONSTANT),
						Value:     utils.StringPointer("1"),
						Mandatory: utils.BoolPointer(true),
						Layout:    utils.StringPointer(string(time.RFC3339)),
					},
				},
			},
		},
	}
	expected := &RadiusAgentCfg{
		Enabled:            true,
		ListenNet:          "udp",
		ListenAuth:         "127.0.0.1:1812",
		ListenAcct:         "127.0.0.1:1813",
		ClientSecrets:      map[string]string{utils.MetaDefault: "CGRateS.org"},
		ClientDictionaries: map[string]string{utils.MetaDefault: "/usr/share/cgrates/radius/dict/"},
		SessionSConns:      []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaSessionS)},
		RequestProcessors: []*RequestProcessor{
			{
				ID:            "OutboundAUTHDryRun",
				Filters:       []string{"*string:~*req.request_type:OutboundAUTH", "*string:~*req.Msisdn:497700056231"},
				Flags:         utils.FlagsWithParams{utils.MetaDryRun: {}},
				Timezone:      utils.EmptyString,
				Tenant:        NewRSRParsersMustCompile("~*req.CGRID", utils.InfieldSep),
				RequestFields: []*FCTemplate{},
				ReplyFields: []*FCTemplate{
					{
						Tag:       "Allow",
						Path:      "*rep.response.Allow",
						Type:      utils.META_CONSTANT,
						Value:     NewRSRParsersMustCompile("1", utils.InfieldSep),
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
	cfg := NewDefaultCGRConfig()
	if err = cfg.radiusAgentCfg.loadFromJSONCfg(cfgJSON, cfg.generalCfg.RSRSep); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, cfg.radiusAgentCfg) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(cfg.radiusAgentCfg))
	}
}

func TestRadiusAgentCfgloadFromJsonCfgCase2(t *testing.T) {
	cfgJSONStr := `{
	"radius_agent": {
         "request_processors": [
			{
				"id": "OutboundAUTHDryRun",
			},
         ],									
     },
}`
	cfgJSON := &RadiusAgentJsonCfg{
		Request_processors: &[]*ReqProcessorJsnCfg{
			{
				ID: utils.StringPointer("OutboundAUTHDryRun"),
			},
		},
	}
	expected := &RadiusAgentCfg{
		RequestProcessors: []*RequestProcessor{
			{
				ID: "OutboundAUTHDryRun",
			},
		},
	}
	if jsonCfg, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if err = jsonCfg.radiusAgentCfg.loadFromJSONCfg(cfgJSON, jsonCfg.generalCfg.RSRSep); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(jsonCfg.radiusAgentCfg.RequestProcessors[0].ID, expected.RequestProcessors[0].ID) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(jsonCfg.radiusAgentCfg.RequestProcessors[0].ID),
			utils.ToJSON(expected.RequestProcessors[0].ID))
	}
}

func TestRadiusAgentCfgloadFromJsonCfgCase3(t *testing.T) {
	cfgJSON := &RadiusAgentJsonCfg{
		Request_processors: &[]*ReqProcessorJsnCfg{
			{
				Tenant: utils.StringPointer("a{*"),
			},
		},
	}
	expected := "invalid converter terminator in rule: <a{*>"
	jsonCfg := NewDefaultCGRConfig()
	if err = jsonCfg.radiusAgentCfg.loadFromJSONCfg(cfgJSON, jsonCfg.generalCfg.RSRSep); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestRadiusAgentCfgAsMapInterface(t *testing.T) {
	cfgJSONStr := `{
	"radius_agent": {
         "enabled": true,												
         "listen_auth": "127.0.0.1:1816",							
         "listen_acct": "127.0.0.1:1892",							

	     "client_dictionaries": {									
	    	"*default": "/usr/share/cgrates/",			
	     },
	     "sessions_conns": ["*conn1","*conn2"],
         "request_processors": [
			{
				"id": "OutboundAUTHDryRun",
				"filters": ["*string:~*req.request_type:OutboundAUTH","*string:~*req.Msisdn:497700056231"],
				"tenant": "cgrates.org",
				"flags": ["*dryrun"],
				"request_fields":[],
				"reply_fields":[
					{"tag": "Allow", "path": "*rep.response.Allow", "type": "*constant", 
						"value": "1", "mandatory": true},
				],
			},],									
     },
}`
	eMap := map[string]interface{}{
		utils.EnabledCfg:    true,
		utils.ListenNetCfg:  "udp",
		utils.ListenAuthCfg: "127.0.0.1:1816",
		utils.ListenAcctCfg: "127.0.0.1:1892",
		utils.ClientSecretsCfg: map[string]string{
			utils.MetaDefault: "CGRateS.org",
		},
		utils.ClientDictionariesCfg: map[string]string{
			utils.MetaDefault: "/usr/share/cgrates/",
		},
		utils.SessionSConnsCfg: []string{"*conn1", "*conn2"},
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
				},
			},
		},
	}
	if cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cgrCfg.radiusAgentCfg.AsMapInterface(cgrCfg.generalCfg.RSRSep); !reflect.DeepEqual(rcv, eMap) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}
}

func TestRadiusAgentCfgAsMapInterface1(t *testing.T) {
	cfgJSONStr := `{
	"radius_agent": {},
}`
	eMap := map[string]interface{}{
		utils.EnabledCfg:    false,
		utils.ListenNetCfg:  "udp",
		utils.ListenAuthCfg: "127.0.0.1:1812",
		utils.ListenAcctCfg: "127.0.0.1:1813",
		utils.ClientSecretsCfg: map[string]string{
			utils.MetaDefault: "CGRateS.org",
		},
		utils.ClientDictionariesCfg: map[string]string{
			utils.MetaDefault: "/usr/share/cgrates/radius/dict/",
		},
		utils.SessionSConnsCfg:     []string{"*internal"},
		utils.RequestProcessorsCfg: []map[string]interface{}{},
	}
	if cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cgrCfg.radiusAgentCfg.AsMapInterface(cgrCfg.generalCfg.RSRSep); !reflect.DeepEqual(rcv, eMap) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}
}

func TestRadiusAgentCfgClone(t *testing.T) {
	ban := &RadiusAgentCfg{
		Enabled:            true,
		ListenNet:          "udp",
		ListenAuth:         "127.0.0.1:1812",
		ListenAcct:         "127.0.0.1:1813",
		ClientSecrets:      map[string]string{utils.MetaDefault: "CGRateS.org"},
		ClientDictionaries: map[string]string{utils.MetaDefault: "/usr/share/cgrates/radius/dict/"},
		SessionSConns:      []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaSessionS), "*conn1"},
		RequestProcessors: []*RequestProcessor{
			{
				ID:            "OutboundAUTHDryRun",
				Filters:       []string{"*string:~*req.request_type:OutboundAUTH", "*string:~*req.Msisdn:497700056231"},
				Flags:         utils.FlagsWithParams{utils.MetaDryRun: {}},
				Timezone:      utils.EmptyString,
				Tenant:        NewRSRParsersMustCompile("~*req.CGRID", utils.InfieldSep),
				RequestFields: []*FCTemplate{},
				ReplyFields: []*FCTemplate{
					{
						Tag:       "Allow",
						Path:      "*rep.response.Allow",
						Type:      utils.META_CONSTANT,
						Value:     NewRSRParsersMustCompile("1", utils.InfieldSep),
						Mandatory: true,
						Layout:    time.RFC3339,
					},
				},
			},
		},
	}
	rcv := ban.Clone()
	if !reflect.DeepEqual(ban, rcv) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(ban), utils.ToJSON(rcv))
	}
	if rcv.SessionSConns[1] = ""; ban.SessionSConns[1] != "*conn1" {
		t.Errorf("Expected clone to not modify the cloned")
	}
	if rcv.RequestProcessors[0].ID = ""; ban.RequestProcessors[0].ID != "OutboundAUTHDryRun" {
		t.Errorf("Expected clone to not modify the cloned")
	}
	if rcv.ClientSecrets[utils.MetaDefault] = ""; ban.ClientSecrets[utils.MetaDefault] != "CGRateS.org" {
		t.Errorf("Expected clone to not modify the cloned")
	}
	if rcv.ClientDictionaries[utils.MetaDefault] = ""; ban.ClientDictionaries[utils.MetaDefault] != "/usr/share/cgrates/radius/dict/" {
		t.Errorf("Expected clone to not modify the cloned")
	}
}
