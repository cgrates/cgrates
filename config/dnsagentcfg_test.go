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

func TestDNSAgentCfgloadFromJsonCfg(t *testing.T) {
	jsnCfg := &DNSAgentJsonCfg{
		Enabled:        utils.BoolPointer(true),
		Listen:         utils.StringPointer("127.0.0.1:2053"),
		Listen_net:     utils.StringPointer("udp"),
		Sessions_conns: &[]string{utils.MetaInternal, "*conn1"},
		Timezone:       utils.StringPointer("UTC"),
		Request_processors: &[]*ReqProcessorJsnCfg{
			{
				ID:             utils.StringPointer("OutboundAUTHDryRun"),
				Filters:        &[]string{"*string:~*req.request_type:OutboundAUTH", "*string:~*req.Msisdn:497700056231"},
				Flags:          &[]string{"*dryrun"},
				Timezone:       utils.StringPointer("UTC"),
				Request_fields: &[]*FcTemplateJsonCfg{},
				Reply_fields: &[]*FcTemplateJsonCfg{
					{Tag: utils.StringPointer("Allow"), Path: utils.StringPointer("*rep.response.Allow"), Type: utils.StringPointer("constant"),
						Mandatory: utils.BoolPointer(true), Layout: utils.StringPointer(utils.EmptyString)},
				},
			},
		},
	}
	expected := &DNSAgentCfg{
		Enabled:       true,
		Listen:        "127.0.0.1:2053",
		ListenNet:     "udp",
		SessionSConns: []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaSessionS), "*conn1"},
		Timezone:      "UTC",
		RequestProcessors: []*RequestProcessor{
			{
				ID:            "OutboundAUTHDryRun",
				Filters:       []string{"*string:~*req.request_type:OutboundAUTH", "*string:~*req.Msisdn:497700056231"},
				Flags:         utils.FlagsWithParamsFromSlice([]string{utils.MetaDryRun}),
				Timezone:      "UTC",
				RequestFields: []*FCTemplate{},
				ReplyFields: []*FCTemplate{
					{Tag: "Allow", Path: "*rep.response.Allow", Type: "constant", Mandatory: true, Layout: utils.EmptyString},
				},
			},
		},
	}
	for _, v := range expected.RequestProcessors[0].ReplyFields {
		v.ComputePath()
	}
	if jsonCfg, err := NewDefaultCGRConfig(); err != nil {
		t.Error(err)
	} else if err = jsonCfg.dnsAgentCfg.loadFromJsonCfg(jsnCfg, jsonCfg.generalCfg.RSRSep); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(jsonCfg.dnsAgentCfg, expected) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(jsonCfg.dnsAgentCfg))
	}
}

func TestRequestProcessorloadFromJsonCfg(t *testing.T) {
	var dareq, expected RequestProcessor
	if err := dareq.loadFromJSONCfg(nil, utils.INFIELD_SEP); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(dareq, expected) {
		t.Errorf("Expected: %+v ,received: %+v", expected, dareq)
	}
	if err := dareq.loadFromJSONCfg(new(ReqProcessorJsnCfg), utils.INFIELD_SEP); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(dareq, expected) {
		t.Errorf("Expected: %+v ,received: %+v", expected, dareq)
	}
	json := &ReqProcessorJsnCfg{
		ID:      utils.StringPointer("cgrates"),
		Tenant:  utils.StringPointer("tenant"),
		Filters: &[]string{"filter1", "filter2"},
		Flags:   &[]string{"flag1", "flag2"},
	}
	expected = RequestProcessor{
		ID:      "cgrates",
		Tenant:  NewRSRParsersMustCompile("tenant", utils.INFIELD_SEP),
		Filters: []string{"filter1", "filter2"},
		Flags:   utils.FlagsWithParams{"flag1": {}, "flag2": {}},
	}
	if err = dareq.loadFromJSONCfg(json, utils.INFIELD_SEP); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, dareq) {
		t.Errorf("Expected: %+v , received: %+v", utils.ToJSON(expected), utils.ToJSON(dareq))
	}
}

func TestRequestProcessorDNSAgentloadFromJsonCfg(t *testing.T) {
	cfgJSON := &DNSAgentJsonCfg{
		Request_processors: &[]*ReqProcessorJsnCfg{
			{
				Tenant: utils.StringPointer("a{*"),
			},
		},
	}
	expected := "invalid converter terminator in rule: <a{*>"
	if jsonCfg, err := NewDefaultCGRConfig(); err != nil {
		t.Error(err)
	} else if err := jsonCfg.dnsAgentCfg.loadFromJsonCfg(cfgJSON, jsonCfg.generalCfg.RSRSep); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestRequestProcessorDNSAgentloadFromJsonCfg1(t *testing.T) {
	cfgJSONStr := `{ 
      "dns_agent": {
        "request_processors": [
	        {
		       "id": "random",
            },
         ]
       }
}`
	cfgJSON := &DNSAgentJsonCfg{
		Request_processors: &[]*ReqProcessorJsnCfg{
			{
				ID: utils.StringPointer("random"),
			},
		},
	}
	if jsonCfg, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if err = jsonCfg.dnsAgentCfg.loadFromJsonCfg(cfgJSON, jsonCfg.generalCfg.RSRSep); err != nil {
		t.Error(err)
	}
}

func TestRequestProcessorReplyFieldsloadFromJsonCfg(t *testing.T) {
	cfgJSON := &DNSAgentJsonCfg{
		Request_processors: &[]*ReqProcessorJsnCfg{
			{
				Reply_fields: &[]*FcTemplateJsonCfg{
					{
						Value: utils.StringPointer("a{*"),
					},
				},
			},
		},
	}
	expected := "invalid converter terminator in rule: <a{*>"
	if jsonCfg, err := NewDefaultCGRConfig(); err != nil {
		t.Error(err)
	} else if err := jsonCfg.dnsAgentCfg.loadFromJsonCfg(cfgJSON, jsonCfg.generalCfg.RSRSep); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestRequestProcessorRequestFieldsloadFromJsonCfg(t *testing.T) {
	cfgJSON := &DNSAgentJsonCfg{
		Request_processors: &[]*ReqProcessorJsnCfg{
			{
				Request_fields: &[]*FcTemplateJsonCfg{
					{
						Value: utils.StringPointer("a{*"),
					},
				},
			},
		},
	}
	expected := "invalid converter terminator in rule: <a{*>"
	if jsonCfg, err := NewDefaultCGRConfig(); err != nil {
		t.Error(err)
	} else if err := jsonCfg.dnsAgentCfg.loadFromJsonCfg(cfgJSON, jsonCfg.generalCfg.RSRSep); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestDNSAgentCfgAsMapInterface(t *testing.T) {
	cfgJSONStr := `{
	"dns_agent": {
		"enabled": false,
		"listen": "127.0.0.1:2053",
		"listen_net": "udp",
		"sessions_conns": ["*internal"],
		"timezone": "",
		"request_processors": [],
	},
}`
	eMap := map[string]interface{}{

		utils.EnabledCfg:           false,
		utils.ListenCfg:            "127.0.0.1:2053",
		utils.ListenNetCfg:         "udp",
		utils.SessionSConnsCfg:     []string{"*internal"},
		utils.TimezoneCfg:          "",
		utils.RequestProcessorsCfg: []map[string]interface{}{},
	}
	if cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cgrCfg.dnsAgentCfg.AsMapInterface(utils.EmptyString); !reflect.DeepEqual(eMap, rcv) {
		t.Errorf("Expected %+v, received %+v", eMap, rcv)
	}
}

func TestDNSAgentCfgAsMapInterface1(t *testing.T) {
	cfgJSONStr := `{
		"dns_agent": {
			"enabled": false,
			"listen": "127.0.0.1:2053",
			"listen_net": "udp",
			"sessions_conns": ["*internal:*sessions", "*conn1"],
			"timezone": "UTC",
			"request_processors": [
			{
				"id": "OutboundAUTHDryRun",
				"filters": ["*string:~*req.request_type:OutboundAUTH","*string:~*req.Msisdn:497700056231"],
				"tenant": "cgrates.org",
				"flags": ["*dryrun"],
                "timezone": "UTC",
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
	}`
	eMap := map[string]interface{}{
		utils.EnabledCfg:       false,
		utils.ListenCfg:        "127.0.0.1:2053",
		utils.ListenNetCfg:     "udp",
		utils.SessionSConnsCfg: []string{utils.MetaInternal, "*conn1"},
		utils.TimezoneCfg:      "UTC",
		utils.RequestProcessorsCfg: []map[string]interface{}{
			{
				utils.IDCfg:            "OutboundAUTHDryRun",
				utils.FiltersCfg:       []string{"*string:~*req.request_type:OutboundAUTH", "*string:~*req.Msisdn:497700056231"},
				utils.TenantCfg:        "cgrates.org",
				utils.FlagsCfg:         []string{"*dryrun"},
				utils.TimezoneCfg:      "UTC",
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
	} else if rcv := cgrCfg.dnsAgentCfg.AsMapInterface(cgrCfg.generalCfg.RSRSep); !reflect.DeepEqual(eMap, rcv) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToIJSON(eMap), utils.ToIJSON(rcv))
	}
}
