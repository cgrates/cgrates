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
	"strings"
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
	"sessions_conns": [											// connections to SessionS for session management and CDR posting
		{"address": "*internal"}
	],
	"timezone": "UTC",												// timezone of the events if not specified  <UTC|Local|$IANA_TZ_DB>
	"request_processors": [										// request processors to be applied to DNS messages
	],
},
}`
	expected = DNSAgentCfg{
		Listen:        "127.0.0.1:2053",
		ListenNet:     "udp",
		SessionSConns: []*RemoteHost{{Address: "*internal"}},
		Timezone:      "UTC",
	}
	if jsnCfg, err := NewCgrJsonCfgFromReader(strings.NewReader(cfgJSONStr)); err != nil {
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
		ID:                  utils.StringPointer("cgrates"),
		Tenant:              utils.StringPointer("tenant"),
		Filters:             &[]string{"filter1", "filter2"},
		Flags:               &[]string{"flag1", "flag2"},
		Continue_on_success: utils.BoolPointer(true),
	}
	expected = RequestProcessor{
		ID:                "cgrates",
		Tenant:            NewRSRParsersMustCompile("tenant", true, utils.INFIELD_SEP),
		Filters:           []string{"filter1", "filter2"},
		Flags:             utils.FlagsWithParams{"flag1": []string{}, "flag2": []string{}},
		ContinueOnSuccess: true,
	}
	if err = dareq.loadFromJsonCfg(json, utils.INFIELD_SEP); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, dareq) {
		t.Errorf("Expected: %+v , recived: %+v", utils.ToJSON(expected), utils.ToJSON(dareq))
	}
}
