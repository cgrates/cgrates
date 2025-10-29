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
	"slices"
	"testing"
	"time"

	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
	"github.com/google/go-cmp/cmp"
)

func TestRadiusAgentCfgloadFromJsonCfgCase1(t *testing.T) {
	cfgJSON := &RadiusAgentJsonCfg{
		Enabled: utils.BoolPointer(true),
		Listeners: &[]*RadiListenerJsnCfg{
			{
				Network:      utils.StringPointer(utils.UDP),
				Auth_Address: utils.StringPointer("127.0.0.1:1812"),
				Acct_Address: utils.StringPointer("127.0.0.1:1813"),
			},
		},
		ClientSecrets:      &map[string]string{utils.MetaDefault: "CGRateS.org"},
		ClientDictionaries: &map[string][]string{utils.MetaDefault: {"/usr/share/cgrates/radius/dict/"}},
		SessionSConns:      &[]string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaSessionS)},
		RequestProcessors: &[]*ReqProcessorJsnCfg{
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
						Type:      utils.StringPointer(utils.MetaConstant),
						Value:     utils.StringPointer("1"),
						Mandatory: utils.BoolPointer(true),
						Layout:    utils.StringPointer(string(time.RFC3339)),
					},
				},
			},
		},
		ClientDaAddresses: map[string]DAClientOptsJson{
			"fsfdsz": {
				Transport: utils.StringPointer("http"),
				Host:      utils.StringPointer("localhost"),
				Port:      utils.IntPointer(6768),
				Flags:     []string{"*sessions", "*routes"},
			},
		},
	}
	expected := &RadiusAgentCfg{
		Enabled: true,
		Listeners: []RadiusListener{
			{
				Network:  utils.UDP,
				AuthAddr: "127.0.0.1:1812",
				AcctAddr: "127.0.0.1:1813",
			},
		},
		ClientSecrets:      map[string]string{utils.MetaDefault: "CGRateS.org"},
		ClientDictionaries: map[string][]string{utils.MetaDefault: {"/usr/share/cgrates/radius/dict/"}},
		SessionSConns:      []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaSessionS)},
		StatSConns:         []string{},
		ThresholdSConns:    []string{},
		DMRTemplate:        "*dmr",
		CoATemplate:        "*coa",
		ClientDaAddresses: map[string]DAClientOpts{
			"fsfdsz": {
				Transport: "http",
				Host:      "localhost",
				Port:      6768,
				Flags: utils.FlagsWithParams{
					utils.MetaSessionS: utils.FlagParams{},
					utils.MetaRoutes:   utils.FlagParams{},
				},
			},
		},
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
						Type:      utils.MetaConstant,
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
	if err := cfg.radiusAgentCfg.loadFromJSONCfg(cfgJSON, cfg.generalCfg.RSRSep); err != nil {
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
		RequestProcessors: &[]*ReqProcessorJsnCfg{
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
	} else if err := jsonCfg.radiusAgentCfg.loadFromJSONCfg(cfgJSON, jsonCfg.generalCfg.RSRSep); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(jsonCfg.radiusAgentCfg.RequestProcessors[0].ID, expected.RequestProcessors[0].ID) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(jsonCfg.radiusAgentCfg.RequestProcessors[0].ID),
			utils.ToJSON(expected.RequestProcessors[0].ID))
	}
}

func TestRadiusAgentCfgloadFromJsonCfgCase3(t *testing.T) {
	cfgJSON := &RadiusAgentJsonCfg{
		RequestProcessors: &[]*ReqProcessorJsnCfg{
			{
				Tenant: utils.StringPointer("a{*"),
			},
		},
	}
	expected := "invalid converter terminator in rule: <a{*>"
	jsonCfg := NewDefaultCGRConfig()
	if err := jsonCfg.radiusAgentCfg.loadFromJSONCfg(cfgJSON, jsonCfg.generalCfg.RSRSep); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestRadiusAgentCfgAsMapInterface(t *testing.T) {
	cfgJSONStr := `{
	"radius_agent": {
         "enabled": true,		
		 "listeners":[
			{
				"auth_address": "127.0.0.1:1816",					
				"acct_address": "127.0.0.1:1892"
			}
		],														
	     "client_dictionaries": {									
	    	"*default": [
				"/usr/share/cgrates/",
			],			
	     },
	     "sessions_conns": ["*birpc_internal", "*conn1","*conn2"],
	     "stats_conns": ["*internal", "*conn1","*conn2"],
	     "thresholds_conns": ["*internal", "*conn1","*conn2"],
		 "dmr_template": "*dmr",
		 "coa_template": "*coa",
		 "requests_cache_key": "~*req.Acc-Session-Id",
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
	eMap := map[string]any{
		utils.EnabledCfg: true,
		utils.ListenersCfg: []map[string]any{
			{
				utils.NetworkCfg:  utils.EmptyString,
				utils.AuthAddrCfg: "127.0.0.1:1816",
				utils.AcctAddrCfg: "127.0.0.1:1892",
			},
		},
		utils.ClientSecretsCfg: map[string]string{
			utils.MetaDefault: "CGRateS.org",
		},
		utils.ClientDictionariesCfg: map[string][]string{
			utils.MetaDefault: {"/usr/share/cgrates/"},
		},
		utils.SessionSConnsCfg:    []string{rpcclient.BiRPCInternal, "*conn1", "*conn2"},
		utils.StatSConnsCfg:       []string{rpcclient.InternalRPC, "*conn1", "*conn2"},
		utils.ThresholdSConnsCfg:  []string{rpcclient.InternalRPC, "*conn1", "*conn2"},
		utils.DMRTemplateCfg:      "*dmr",
		utils.CoATemplateCfg:      "*coa",
		utils.RequestsCacheKeyCfg: "~*req.Acc-Session-Id",
		utils.RequestProcessorsCfg: []map[string]any{
			{
				utils.IDCfg:            "OutboundAUTHDryRun",
				utils.FiltersCfg:       []string{"*string:~*req.request_type:OutboundAUTH", "*string:~*req.Msisdn:497700056231"},
				utils.TenantCfg:        "cgrates.org",
				utils.FlagsCfg:         []string{"*dryrun"},
				utils.TimezoneCfg:      "",
				utils.RequestFieldsCfg: []map[string]any{},
				utils.ReplyFieldsCfg: []map[string]any{
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
	eMap := map[string]any{
		utils.EnabledCfg: false,
		utils.ListenersCfg: []map[string]any{
			{
				utils.NetworkCfg:  utils.UDP,
				utils.AuthAddrCfg: "127.0.0.1:1812",
				utils.AcctAddrCfg: "127.0.0.1:1813",
			},
		},
		utils.ClientSecretsCfg: map[string]string{
			utils.MetaDefault: "CGRateS.org",
		},
		utils.ClientDictionariesCfg: map[string][]string{
			utils.MetaDefault: {"/usr/share/cgrates/radius/dict/"},
		},
		utils.SessionSConnsCfg:     []string{"*internal"},
		utils.StatSConnsCfg:        []string{},
		utils.ThresholdSConnsCfg:   []string{},
		utils.DMRTemplateCfg:       "*dmr",
		utils.CoATemplateCfg:       "*coa",
		utils.RequestsCacheKeyCfg:  "",
		utils.RequestProcessorsCfg: []map[string]any{},
	}
	if cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cgrCfg.radiusAgentCfg.AsMapInterface(cgrCfg.generalCfg.RSRSep); !reflect.DeepEqual(rcv, eMap) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}
}

func TestRadiusAgentCfgClone(t *testing.T) {
	ban := &RadiusAgentCfg{
		Enabled: true,
		Listeners: []RadiusListener{
			{
				Network:  utils.UDP,
				AuthAddr: "127.0.0.1:1812",
				AcctAddr: "127.0.0.1:1813",
			},
		},
		ClientSecrets:      map[string]string{utils.MetaDefault: "CGRateS.org"},
		ClientDictionaries: map[string][]string{utils.MetaDefault: {"/usr/share/cgrates/radius/dict/"}},
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
						Type:      utils.MetaConstant,
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
	rcv.ClientDictionaries[utils.MetaDefault] = []string{""}
	if !reflect.DeepEqual(ban.ClientDictionaries[utils.MetaDefault],
		[]string{"/usr/share/cgrates/radius/dict/"}) {
		t.Errorf("Expected clone to not modify the cloned")
	}
}

func TestDAClientOptsClone(t *testing.T) {
	originalOpts := &DAClientOpts{
		Transport: "udp",
		Host:      "localhost",
		Port:      6768,
		Flags: utils.FlagsWithParams{
			utils.MetaRoutes:   utils.FlagParams{},
			utils.MetaSessionS: utils.FlagParams{},
		},
	}

	got := originalOpts.Clone()

	if diff := cmp.Diff(originalOpts, got); diff != "" {
		t.Errorf("Clone() returned an unexpected value(-want +got): \n%s", diff)
	}
}

func TestDAClientOptsAsMapInterface(t *testing.T) {
	expectedMap := map[string]any{
		utils.TransportCfg: "udp",
		utils.HostCfg:      "localhost",
		utils.PortCfg:      6768,
		utils.FlagsCfg: []string{utils.MetaRoutes,
			utils.MetaSessionS},
	}
	opts := &DAClientOpts{
		Transport: "udp",
		Host:      "localhost",
		Port:      6768,
		Flags: utils.FlagsWithParams{
			utils.MetaSessionS: utils.FlagParams{},
			utils.MetaRoutes:   utils.FlagParams{},
		},
	}

	got := opts.AsMapInterface()
	slices.Sort(got[utils.FlagsCfg].([]string))

	if diff := cmp.Diff(expectedMap, got); diff != "" {
		t.Errorf("opts.AsMapInterface() returned an unexpected value(-want +got): \n%s", diff)
	}
}

func TestDiffMapStringSlice(t *testing.T) {
	tests := []struct {
		name string
		d    map[string][]string
		v1   map[string][]string
		v2   map[string][]string
		want map[string][]string
	}{
		{
			name: "Empty maps",
			d:    nil,
			v1:   map[string][]string{},
			v2:   map[string][]string{},
			want: map[string][]string{},
		},
		{
			name: "v1 has no key from v2",
			d:    nil,
			v1:   map[string][]string{},
			v2:   map[string][]string{"tenant1": {"id1", "id2"}},
			want: map[string][]string{"tenant1": {"id1", "id2"}},
		},
		{
			name: "Different values for the same key",
			d:    nil,
			v1:   map[string][]string{"tenant1": {"id1", "id2"}},
			v2:   map[string][]string{"tenant1": {"id3", "id4"}},
			want: map[string][]string{"tenant1": {"id3", "id4"}},
		},
		{
			name: "Same key with equal values",
			d:    nil,
			v1:   map[string][]string{"tenant1": {"id1", "id2"}},
			v2:   map[string][]string{"tenant1": {"id1", "id2"}},
			want: map[string][]string{},
		},
		{
			name: "d is not nil, adds differences from v2",
			d:    map[string][]string{"existing": {"val3"}},
			v1:   map[string][]string{"tenant1": {"id1", "id2"}},
			v2:   map[string][]string{"tenant1": {"id3", "id4"}, "tenant2": {"id4", "id5"}},
			want: map[string][]string{"existing": {"val3"}, "tenant1": {"id3", "id4"}, "tenant2": {"id4", "id5"}},
		},
		{
			name: "Adding multiple tenants",
			d:    nil,
			v1:   map[string][]string{"tenant1": {"id1", "id2"}},
			v2:   map[string][]string{"tenant1": {"id1", "id2"}, "tenant2": {"id4", "id5"}},
			want: map[string][]string{"tenant2": {"id4", "id5"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := diffMapStringSlice(tt.d, tt.v1, tt.v2)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("diffMapStringSlice() = %v, want %v", got, tt.want)
			}
		})
	}
}
