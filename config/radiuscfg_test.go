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

func TestRadiusAgentCfgloadFromJsonCfg(t *testing.T) {
	var racfg, expected RadiusAgentCfg
	if err := racfg.loadFromJsonCfg(nil, utils.INFIELD_SEP); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(racfg, expected) {
		t.Errorf("Expected: %+v ,recived: %+v", expected, racfg)
	}
	if err := racfg.loadFromJsonCfg(new(RadiusAgentJsonCfg), utils.INFIELD_SEP); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(racfg, expected) {
		t.Errorf("Expected: %+v ,recived: %+v", expected, racfg)
	}
	cfgJSONStr := `{
"radius_agent": {
	"enabled": false,											// enables the radius agent: <true|false>
	"listen_net": "udp",										// network to listen on <udp|tcp>
	"listen_auth": "127.0.0.1:1812",							// address where to listen for radius authentication requests <x.y.z.y:1234>
	"listen_acct": "127.0.0.1:1813",							// address where to listen for radius accounting requests <x.y.z.y:1234>
	"client_secrets": {											// hash containing secrets for clients connecting here <*default|$client_ip>
		"*default": "CGRateS.org"
	},
	"client_dictionaries": {									// per client path towards directory holding additional dictionaries to load (extra to RFC)
		"*default": "/usr/share/cgrates/radius/dict/",			// key represents the client IP or catch-all <*default|$client_ip>
	},
	"sessions_conns": [
		{"address": "*internal"}								// connection towards SessionService
	],
	"request_processors": [],
},
}`
	expected = RadiusAgentCfg{
		ListenNet:          "udp",
		ListenAuth:         "127.0.0.1:1812",
		ListenAcct:         "127.0.0.1:1813",
		ClientSecrets:      map[string]string{"*default": "CGRateS.org"},
		ClientDictionaries: map[string]string{"*default": "/usr/share/cgrates/radius/dict/"},
		SessionSConns:      []*HaPoolConfig{{Address: "*internal"}},
	}
	if jsnCfg, err := NewCgrJsonCfgFromReader(strings.NewReader(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnRaCfg, err := jsnCfg.RadiusAgentJsonCfg(); err != nil {
		t.Error(err)
	} else if err = racfg.loadFromJsonCfg(jsnRaCfg, utils.INFIELD_SEP); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, racfg) {
		t.Errorf("Expected: %+v , recived: %+v", utils.ToJSON(expected), utils.ToJSON(racfg))
	}
}

func TestRARequestProcessorloadFromJsonCfg(t *testing.T) {
	var rareq, expected RARequestProcessor
	if err := rareq.loadFromJsonCfg(nil, utils.INFIELD_SEP); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rareq, expected) {
		t.Errorf("Expected: %+v ,recived: %+v", expected, rareq)
	}
	if err := rareq.loadFromJsonCfg(new(RAReqProcessorJsnCfg), utils.INFIELD_SEP); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rareq, expected) {
		t.Errorf("Expected: %+v ,recived: %+v", expected, rareq)
	}
	json := &RAReqProcessorJsnCfg{
		Id:                  utils.StringPointer("cgrates"),
		Tenant:              utils.StringPointer("tenant"),
		Filters:             &[]string{"filter1", "filter2"},
		Flags:               &[]string{"flag1", "flag2"},
		Timezone:            utils.StringPointer("Local"),
		Continue_on_success: utils.BoolPointer(true),
	}
	expected = RARequestProcessor{
		Id:                "cgrates",
		Tenant:            NewRSRParsersMustCompile("tenant", true, utils.INFIELD_SEP),
		Filters:           []string{"filter1", "filter2"},
		Flags:             utils.StringMap{"flag1": true, "flag2": true},
		Timezone:          "Local",
		ContinueOnSuccess: true,
	}
	if err = rareq.loadFromJsonCfg(json, utils.INFIELD_SEP); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rareq) {
		t.Errorf("Expected: %+v , recived: %+v", utils.ToJSON(expected), utils.ToJSON(rareq))
	}
}
