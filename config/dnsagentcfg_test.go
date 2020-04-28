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
		Tenant:  NewRSRParsersMustCompile("tenant", true, utils.INFIELD_SEP),
		Filters: []string{"filter1", "filter2"},
		Flags:   utils.FlagsWithParams{"flag1": []string{}, "flag2": []string{}},
	}
	if err = dareq.loadFromJsonCfg(json, utils.INFIELD_SEP); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, dareq) {
		t.Errorf("Expected: %+v , recived: %+v", utils.ToJSON(expected), utils.ToJSON(dareq))
	}
}

func TestDNSAgentCfgAsMapInterface(t *testing.T) {
	var dnsCfg DNSAgentCfg
	cfgJSONStr := `{
	"dns_agent": {
		"enabled": false,
		"listen": "127.0.0.1:2053",
		"listen_net": "udp",
		"sessions_conns": ["*internal"],
		"timezone": "",
		"request_processors": [
		],
	},
}`
	eMap := map[string]interface{}{
		"enabled":            false,
		"listen":             "127.0.0.1:2053",
		"listen_net":         "udp",
		"sessions_conns":     []string{"*internal"},
		"timezone":           "",
		"request_processors": []map[string]interface{}{},
	}
	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnDaCfg, err := jsnCfg.DNSAgentJsonCfg(); err != nil {
		t.Error(err)
	} else if err = dnsCfg.loadFromJsonCfg(jsnDaCfg, utils.INFIELD_SEP); err != nil {
		t.Error(err)
	} else if rcv := dnsCfg.AsMapInterface(utils.EmptyString); !reflect.DeepEqual(eMap, rcv) {
		t.Errorf("\nExpected: %+v\nRecived: %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}

	cfgJSONStr = `{
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
	eMap = map[string]interface{}{
		"enabled":        false,
		"listen":         "127.0.0.1:2053",
		"listen_net":     "udp",
		"sessions_conns": []string{"*internal"},
		"timezone":       "UTC",
		"request_processors": []map[string]interface{}{
			{
				"id":             "OutboundAUTHDryRun",
				"filters":        []string{"*string:~*req.request_type:OutboundAUTH", "*string:~*req.Msisdn:497700056231"},
				"tenant":         "cgrates.org",
				"flags":          map[string][]string{"*dryrun": {}},
				"Timezone":       "",
				"request_fields": []map[string]interface{}{},
				"reply_fields": []map[string]interface{}{
					{"tag": "Allow", "path": "*rep.response.Allow", "type": "*constant", "value": "1", "mandatory": true},
					{"tag": "Concatenated1", "path": "*rep.response.Concatenated", "type": "*composed", "value": "~*req.MCC;/", "mandatory": true},
					{"tag": "Concatenated2", "path": "*rep.response.Concatenated", "type": "*composed", "value": "Val1"},
					{"tag": "MaxDuration", "path": "*rep.response.MaxDuration", "type": "*constant", "value": "1200", "blocker": true},
					{"tag": "Unused", "path": "*rep.response.Unused", "type": "*constant", "value": "0"},
				},
			},
		},
	}
	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnDaCfg, err := jsnCfg.DNSAgentJsonCfg(); err != nil {
		t.Error(err)
	} else if err = dnsCfg.loadFromJsonCfg(jsnDaCfg, utils.INFIELD_SEP); err != nil {
		t.Error(err)
	} else if rcv := dnsCfg.AsMapInterface(";"); !reflect.DeepEqual(eMap, rcv) {
		t.Errorf("\nExpected: %+v\nRecived: %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}

}
