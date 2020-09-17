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

func TestSIPAgentCfgloadFromJsonCfg(t *testing.T) {
	var dnsCfg, expected SIPAgentCfg
	if err := dnsCfg.loadFromJsonCfg(nil, utils.INFIELD_SEP); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(dnsCfg, expected) {
		t.Errorf("Expected: %+v ,recived: %+v", expected, dnsCfg)
	}
	if err := dnsCfg.loadFromJsonCfg(new(SIPAgentJsonCfg), utils.INFIELD_SEP); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(dnsCfg, expected) {
		t.Errorf("Expected: %+v ,recived: %+v", expected, dnsCfg)
	}
	cfgJSONStr := `{
"sip_agent": {
	"enabled": false,											// enables the DNS agent: <true|false>
	"listen": "127.0.0.1:5060",									// address where to listen for DNS requests <x.y.z.y:1234>
	"listen_net": "udp",										// network to listen on <udp|tcp|tcp-tls>
	"sessions_conns": ["*internal"],
	"timezone": "UTC",												// timezone of the events if not specified  <UTC|Local|$IANA_TZ_DB>
	"request_processors": [										// request processors to be applied to DNS messages
	],
},
}`
	expected = SIPAgentCfg{
		Listen:        "127.0.0.1:5060",
		ListenNet:     "udp",
		SessionSConns: []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaSessionS)},
		Timezone:      "UTC",
	}
	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnDaCfg, err := jsnCfg.SIPAgentJsonCfg(); err != nil {
		t.Error(err)
	} else if err = dnsCfg.loadFromJsonCfg(jsnDaCfg, utils.INFIELD_SEP); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, dnsCfg) {
		t.Errorf("Expected: %+v , recived: %+v", utils.ToJSON(expected), utils.ToJSON(dnsCfg))
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
	if cgrCfg, err := NewCGRConfigFromJsonStringWithDefaults(cfgJSONStr); err != nil {
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
	}
	if cgrCfg, err := NewCGRConfigFromJsonStringWithDefaults(cfgJSONStr); err != nil {
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
		utils.RetransmissionTimerCfg: 1 * time.Second,
		utils.RequestProcessorsCfg: []map[string]interface{}{
			{
				utils.IdCfg:            "Register",
				utils.FilterSCfg:       []string{"*notstring:~*vars.Method:INVITE"},
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
	if cgrCfg, err := NewCGRConfigFromJsonStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cgrCfg.sipAgentCfg.AsMapInterface(cgrCfg.generalCfg.RSRSep); !reflect.DeepEqual(rcv, eMap) {
		t.Errorf("Expected %+v \n, received %+v", eMap, rcv)
	}
}
