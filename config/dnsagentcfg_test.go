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
	var dnsCfg, expected DNSAgentCfg
	if err := dnsCfg.loadFromJsonCfg(nil, utils.INFIELD_SEP); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(dnsCfg, expected) {
		t.Errorf("Expected: %+v ,recived: %+v", expected, dnsCfg)
	}
	if err := dnsCfg.loadFromJsonCfg(new(DNSAgentJsonCfg), utils.INFIELD_SEP); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(dnsCfg, expected) {
		t.Errorf("Expected: %+v ,recived: %+v", expected, dnsCfg)
	}
	cfgJSONStr := `{
"dns_agent": {
	"enabled": false,											// enables the DNS agent: <true|false>
	"listen": "127.0.0.1:2053",									// address where to listen for DNS requests <x.y.z.y:1234>
	"listen_net": "udp",										// network to listen on <udp|tcp|tcp-tls>
	"sessions_conns": ["*internal"],
	"timezone": "UTC",												// timezone of the events if not specified  <UTC|Local|$IANA_TZ_DB>
	"request_processors": [										// request processors to be applied to DNS messages
	],
},
}`
	expected = DNSAgentCfg{
		Listen:        "127.0.0.1:2053",
		ListenNet:     "udp",
		SessionSConns: []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaSessionS)},
		Timezone:      "UTC",
	}
	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnDaCfg, err := jsnCfg.DNSAgentJsonCfg(); err != nil {
		t.Error(err)
	} else if err = dnsCfg.loadFromJsonCfg(jsnDaCfg, utils.INFIELD_SEP); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, dnsCfg) {
		t.Errorf("Expected: %+v , recived: %+v", utils.ToJSON(expected), utils.ToJSON(dnsCfg))
	}
}

func TestRequestProcessorloadFromJsonCfg(t *testing.T) {
	var dareq, expected RequestProcessor
	if err := dareq.loadFromJsonCfg(nil, utils.INFIELD_SEP); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(dareq, expected) {
		t.Errorf("Expected: %+v ,recived: %+v", expected, dareq)
	}
	if err := dareq.loadFromJsonCfg(new(ReqProcessorJsnCfg), utils.INFIELD_SEP); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(dareq, expected) {
		t.Errorf("Expected: %+v ,recived: %+v", expected, dareq)
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
	if err = dareq.loadFromJsonCfg(json, utils.INFIELD_SEP); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, dareq) {
		t.Errorf("Expected: %+v , recived: %+v", utils.ToJSON(expected), utils.ToJSON(dareq))
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
	if cgrCfg, err := NewCGRConfigFromJsonStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cgrCfg.dnsAgentCfg.AsMapInterface(utils.EmptyString); !reflect.DeepEqual(eMap, rcv) {
		t.Errorf("Expecetd %+v, received %+v", eMap, rcv)
	}
}

func TestDNSAgentCfgAsMapInterface1(t *testing.T) {
	cfgJSONStr := `{
		"dns_agent": {
			"enabled": false,
			"listen": "127.0.0.1:2053",
			"listen_net": "udp",
			"sessions_conns": ["*internal"],
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
		utils.SessionSConnsCfg: []string{"*internal"},
		utils.TimezoneCfg:      "UTC",
		utils.RequestProcessorsCfg: []map[string]interface{}{
			{
				utils.IdCfg:            "OutboundAUTHDryRun",
				utils.FilterSCfg:       []string{"*string:~*req.request_type:OutboundAUTH", "*string:~*req.Msisdn:497700056231"},
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
	if cgrCfg, err := NewCGRConfigFromJsonStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cgrCfg.dnsAgentCfg.AsMapInterface(cgrCfg.generalCfg.RSRSep); !reflect.DeepEqual(eMap, rcv) {
		t.Errorf("Expecetd %+v \n, received %+v", utils.ToIJSON(eMap), utils.ToIJSON(rcv))
	}
}
