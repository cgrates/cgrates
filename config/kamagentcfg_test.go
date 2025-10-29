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

func TestKamAgentCfgloadFromJsonCfg(t *testing.T) {
	var kamagcfg, expected KamAgentCfg
	if err := kamagcfg.loadFromJsonCfg(nil); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(kamagcfg, expected) {
		t.Errorf("Expected: %+v ,received: %+v", expected, kamagcfg)
	}
	if err := kamagcfg.loadFromJsonCfg(new(KamAgentJsonCfg)); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(kamagcfg, expected) {
		t.Errorf("Expected: %+v ,received: %+v", expected, kamagcfg)
	}
	cfgJSONStr := `{
"kamailio_agent": {
	"enabled": false,						// starts SessionManager service: <true|false>
	"sessions_conns": ["*internal"],
	"create_cdr": false,					// create CDR out of events and sends them to CDRS component
	"timezone": "",							// timezone of the Kamailio server
	"evapi_conns":[							// instantiate connections to multiple Kamailio servers
		{"address": "127.0.0.1:8448", "reconnects": 5}
	],
},
}`
	expected = KamAgentCfg{
		SessionSConns: []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaSessionS)},
		EvapiConns:    []*KamConnCfg{{Address: "127.0.0.1:8448", Reconnects: 5}},
	}
	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnKamAgCfg, err := jsnCfg.KamAgentJsonCfg(); err != nil {
		t.Error(err)
	} else if err = kamagcfg.loadFromJsonCfg(jsnKamAgCfg); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, kamagcfg) {
		t.Errorf("Expected: %+v , received: %+v", utils.ToJSON(expected), utils.ToJSON(kamagcfg))
	}
}

func TestKamConnCfgloadFromJsonCfg(t *testing.T) {
	var kamcocfg, expected KamConnCfg
	if err := kamcocfg.loadFromJsonCfg(nil); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(kamcocfg, expected) {
		t.Errorf("Expected: %+v ,received: %+v", expected, kamcocfg)
	}
	if err := kamcocfg.loadFromJsonCfg(new(KamConnJsonCfg)); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(kamcocfg, expected) {
		t.Errorf("Expected: %+v ,received: %+v", expected, kamcocfg)
	}
	json := &KamConnJsonCfg{
		Address:    utils.StringPointer("127.0.0.1:8448"),
		Reconnects: utils.IntPointer(5),
	}
	expected = KamConnCfg{
		Address:    "127.0.0.1:8448",
		Reconnects: 5,
	}
	if err = kamcocfg.loadFromJsonCfg(json); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, kamcocfg) {
		t.Errorf("Expected: %+v , received: %+v", utils.ToJSON(expected), utils.ToJSON(kamcocfg))
	}
}

func TestKamAgentCfgAsMapInterface(t *testing.T) {
	var kamagcfg KamAgentCfg
	cfgJSONStr := `{
		"kamailio_agent": {
			"enabled": false,
			"sessions_conns": [""],
			"create_cdr": false,
			"timezone": "",
			"evapi_conns":[
				{"address": "127.0.0.1:8448", "reconnects": 5}
			],
		},
	}`
	eMap := map[string]any{
		"enabled":        false,
		"sessions_conns": []string{""},
		"create_cdr":     false,
		"timezone":       "",
		"evapi_conns": []map[string]any{
			{"address": "127.0.0.1:8448", "reconnects": 5, "alias": ""},
		},
	}
	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnKamAgCfg, err := jsnCfg.KamAgentJsonCfg(); err != nil {
		t.Error(err)
	} else if err = kamagcfg.loadFromJsonCfg(jsnKamAgCfg); err != nil {
		t.Error(err)
	} else if rcv := kamagcfg.AsMapInterface(); !reflect.DeepEqual(eMap, rcv) {
		t.Errorf("\nExpected: %+v\nReceived: %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}

	cfgJSONStr = `{
	"kamailio_agent": {
		"enabled": false,
		"sessions_conns": ["*internal"],
		"create_cdr": false,
		"timezone": "",
		"evapi_conns":[
			{"address": "127.0.0.1:8448", "reconnects": 5}
		],
	},
}`
	eMap = map[string]any{
		"enabled":        false,
		"sessions_conns": []string{"*internal"},
		"create_cdr":     false,
		"timezone":       "",
		"evapi_conns": []map[string]any{
			{"address": "127.0.0.1:8448", "reconnects": 5, "alias": ""},
		},
	}
	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnKamAgCfg, err := jsnCfg.KamAgentJsonCfg(); err != nil {
		t.Error(err)
	} else if err = kamagcfg.loadFromJsonCfg(jsnKamAgCfg); err != nil {
		t.Error(err)
	} else if rcv := kamagcfg.AsMapInterface(); !reflect.DeepEqual(eMap, rcv) {
		t.Errorf("\nExpected: %+v\nReceived: %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}
}

func TestKamAgentCfgloadFromJsonCfg2(t *testing.T) {
	nm := 1
	str := "test"
	self := KamConnCfg{}

	js := KamConnJsonCfg{
		Alias:      &str,
		Address:    &str,
		Reconnects: &nm,
	}

	exp := KamConnCfg{
		Alias:      str,
		Address:    str,
		Reconnects: nm,
	}

	err := self.loadFromJsonCfg(&js)
	if err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(self, exp) {
		t.Errorf("\nExpected: %+v\nReceived: %+v", exp, self)
	}
}
