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
)

func TestDNSAgentCfgloadFromJsonCfg(t *testing.T) {
	var dnsCfg, expected DNSAgentCfg
	if err := dnsCfg.loadFromJsonCfg(nil, utils.INFIELD_SEP); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(dnsCfg, expected) {
		t.Errorf("Expected: %+v ,received: %+v", expected, dnsCfg)
	}
	if err := dnsCfg.loadFromJsonCfg(new(DNSAgentJsonCfg), utils.INFIELD_SEP); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(dnsCfg, expected) {
		t.Errorf("Expected: %+v ,received: %+v", expected, dnsCfg)
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
		t.Errorf("Expected: %+v , received: %+v", utils.ToJSON(expected), utils.ToJSON(dnsCfg))
	}
}

func TestRequestProcessorloadFromJsonCfg(t *testing.T) {
	var dareq, expected RequestProcessor
	if err := dareq.loadFromJsonCfg(nil, utils.INFIELD_SEP); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(dareq, expected) {
		t.Errorf("Expected: %+v ,received: %+v", expected, dareq)
	}
	if err := dareq.loadFromJsonCfg(new(ReqProcessorJsnCfg), utils.INFIELD_SEP); err != nil {
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
		Tenant:  NewRSRParsersMustCompile("tenant", true, utils.INFIELD_SEP),
		Filters: []string{"filter1", "filter2"},
		Flags:   utils.FlagsWithParams{"flag1": []string{}, "flag2": []string{}},
	}
	if err = dareq.loadFromJsonCfg(json, utils.INFIELD_SEP); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, dareq) {
		t.Errorf("Expected: %+v , received: %+v", utils.ToJSON(expected), utils.ToJSON(dareq))
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
	eMap := map[string]any{
		"enabled":            false,
		"listen":             "127.0.0.1:2053",
		"listen_net":         "udp",
		"sessions_conns":     []string{"*internal"},
		"timezone":           "",
		"request_processors": []map[string]any{},
	}
	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnDaCfg, err := jsnCfg.DNSAgentJsonCfg(); err != nil {
		t.Error(err)
	} else if err = dnsCfg.loadFromJsonCfg(jsnDaCfg, utils.INFIELD_SEP); err != nil {
		t.Error(err)
	} else if rcv := dnsCfg.AsMapInterface(utils.EmptyString); !reflect.DeepEqual(eMap, rcv) {
		t.Errorf("\nExpected: %+v\nReceived: %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
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
	eMap = map[string]any{
		"enabled":        false,
		"listen":         "127.0.0.1:2053",
		"listen_net":     "udp",
		"sessions_conns": []string{"*internal"},
		"timezone":       "UTC",
		"request_processors": []map[string]any{
			{
				"id":             "OutboundAUTHDryRun",
				"filters":        []string{"*string:~*req.request_type:OutboundAUTH", "*string:~*req.Msisdn:497700056231"},
				"tenant":         "cgrates.org",
				"flags":          map[string][]string{"*dryrun": {}},
				"Timezone":       "",
				"request_fields": []map[string]any{},
				"reply_fields": []map[string]any{
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
		t.Errorf("\nExpected: %+v\nReceived: %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}

}

func TestDNSAgentCfgloadFromJsonCfg2(t *testing.T) {
	bl := false
	str := "test"
	slc := []string{"val1", "val2"}
	eslc := []string{}
	estr := ""

	da := DNSAgentCfg{
		RequestProcessors: []*RequestProcessor{{
			ID: str,
		}},
	}

	js := DNSAgentJsonCfg{
		Enabled:        &bl,
		Listen:         &str,
		Listen_net:     &str,
		Sessions_conns: &slc,
		Timezone:       &str,
		Request_processors: &[]*ReqProcessorJsnCfg{{
			ID:             &str,
			Filters:        &slc,
			Tenant:         &estr,
			Timezone:       &str,
			Flags:          &eslc,
			Request_fields: &[]*FcTemplateJsonCfg{},
			Reply_fields:   &[]*FcTemplateJsonCfg{},
		}},
	}

	exp := DNSAgentCfg{
		Enabled:       bl,
		Listen:        str,
		ListenNet:     str,
		SessionSConns: slc,
		Timezone:      str,
		RequestProcessors: []*RequestProcessor{{
			ID:            str,
			Tenant:        RSRParsers{},
			Filters:       slc,
			Flags:         utils.FlagsWithParams{},
			Timezone:      str,
			RequestFields: []*FCTemplate{},
			ReplyFields:   []*FCTemplate{},
		}},
	}

	err := da.loadFromJsonCfg(&js, "")
	if err != nil {
		t.Fatal(err)
	}

	if da.RequestProcessors == nil {
		t.Errorf("received %v, expected %v", da, exp)
	}
}

func TestDNSAgentCfgloadFromJsonCfgError(t *testing.T) {
	strErr := "test`"
	da := DNSAgentCfg{}

	js := DNSAgentJsonCfg{
		Request_processors: &[]*ReqProcessorJsnCfg{
			{
				Tenant: &strErr,
			},
		},
	}

	err := da.loadFromJsonCfg(&js, "")
	if err != nil {
		t.Error(err)
	}
}

func TestDNSAgentCfgAsMapInterface2(t *testing.T) {
	da := DNSAgentCfg{
		Enabled:           false,
		Listen:            "test",
		ListenNet:         "test",
		SessionSConns:     []string{"val1", "val2"},
		Timezone:          "test",
		RequestProcessors: []*RequestProcessor{},
	}

	exp := map[string]any{
		utils.EnabledCfg:           da.Enabled,
		utils.ListenCfg:            da.Listen,
		utils.ListenNetCfg:         da.ListenNet,
		utils.SessionSConnsCfg:     []string{"val1", "val2"},
		utils.TimezoneCfg:          da.Timezone,
		utils.RequestProcessorsCfg: []map[string]any{},
	}

	rcv := da.AsMapInterface("")

	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("received %v, expected %v", rcv, exp)
	}
}

func TestDNSAgentCfgloadFromJsonCfgRPErrors(t *testing.T) {
	strErr := "test`"
	type args struct {
		js  *ReqProcessorJsnCfg
		sep string
	}

	rp := RequestProcessor{}

	tests := []struct {
		name string
		args args
		err  string
	}{
		{
			name: "flags error",
			args: args{js: &ReqProcessorJsnCfg{
				Flags: &[]string{"test:test:test:test"},
			}, sep: ""},
			err: utils.ErrUnsupportedFormat.Error(),
		},
		{
			name: "Request fields error",
			args: args{js: &ReqProcessorJsnCfg{
				Request_fields: &[]*FcTemplateJsonCfg{
					{
						Value: &strErr,
					},
				},
			}, sep: ""},
			err: "Unclosed unspilit syntax",
		},
		{
			name: "Reply fields error",
			args: args{js: &ReqProcessorJsnCfg{
				Reply_fields: &[]*FcTemplateJsonCfg{
					{
						Value: &strErr,
					},
				},
			}, sep: ""},
			err: "Unclosed unspilit syntax",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := rp.loadFromJsonCfg(tt.args.js, tt.args.sep)

			if err.Error() != tt.err {
				t.Errorf("received %s, expected %s", err.Error(), tt.err)
			}
		})
	}
}

func TestDNSAgentCfgAsMapInterfaceRP(t *testing.T) {
	str := "test"
	slc := []string{"val1", "val2"}

	rp := RequestProcessor{
		ID:            str,
		Tenant:        RSRParsers{},
		Filters:       slc,
		Flags:         utils.FlagsWithParams{},
		Timezone:      str,
		RequestFields: []*FCTemplate{{}},
		ReplyFields:   []*FCTemplate{},
	}

	exp := map[string]any{
		utils.IDCfg:            rp.ID,
		utils.TenantCfg:        "",
		utils.FiltersCfg:       rp.Filters,
		utils.FlagsCfg:         map[string][]string{},
		utils.TimezoneCfgC:     rp.Timezone,
		utils.RequestFieldsCfg: []map[string]any{{}},
		utils.ReplyFieldsCfg:   []map[string]any{},
	}

	rcv := rp.AsMapInterface("")

	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("received %v, expected %v", rcv, exp)
	}
}
