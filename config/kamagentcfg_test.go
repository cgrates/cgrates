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

func TestKamAgentCfgloadFromJsonCfg(t *testing.T) {
	cfgJSON := &KamAgentJsonCfg{
		Enabled:        utils.BoolPointer(true),
		Sessions_conns: &[]string{"*internal"},
		Create_cdr:     utils.BoolPointer(true),
		Evapi_conns: &[]*KamConnJsonCfg{
			{
				Alias:      utils.StringPointer("randomAlias"),
				Address:    utils.StringPointer("127.0.0.1:8448"),
				Reconnects: utils.IntPointer(10),
			},
		},
		Timezone: utils.StringPointer("Local"),
	}
	expected := &KamAgentCfg{
		Enabled:       true,
		SessionSConns: []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaSessionS)},
		CreateCdr:     true,
		EvapiConns:    []*KamConnCfg{{Address: "127.0.0.1:8448", Reconnects: 10, Alias: "randomAlias"}},
		Timezone:      "Local",
	}
	if jsnCfg, err := NewDefaultCGRConfig(); err != nil {
		t.Error(err)
	} else if err = jsnCfg.kamAgentCfg.loadFromJSONCfg(cfgJSON); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, jsnCfg.kamAgentCfg) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(jsnCfg.kamAgentCfg))
	}
}

func TestKamConnCfgloadFromJsonCfg(t *testing.T) {
	var kamcocfg, expected KamConnCfg
	if err := kamcocfg.loadFromJSONCfg(nil); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(kamcocfg, expected) {
		t.Errorf("Expected: %+v ,received: %+v", expected, kamcocfg)
	}
	if err := kamcocfg.loadFromJSONCfg(new(KamConnJsonCfg)); err != nil {
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
	if err = kamcocfg.loadFromJSONCfg(json); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, kamcocfg) {
		t.Errorf("Expected: %+v , received: %+v", utils.ToJSON(expected), utils.ToJSON(kamcocfg))
	}
}

func TestKamAgentCfgAsMapInterface(t *testing.T) {
	cfgJSONStr := `{
		"kamailio_agent": {
			"sessions_conns": ["*conn1","*conn2"],
			"create_cdr": true,
			"timezone": "UTC",
			"evapi_conns":[
				{"address": "127.0.0.1:8448", "reconnects": 5, "alias": ""}
			],
		},
	}`
	eMap := map[string]interface{}{
		utils.EnabledCfg:       false,
		utils.SessionSConnsCfg: []string{"*conn1", "*conn2"},
		utils.CreateCdrCfg:     true,
		utils.TimezoneCfg:      "UTC",
		utils.EvapiConnsCfg: []map[string]interface{}{
			{utils.AddressCfg: "127.0.0.1:8448", utils.ReconnectsCfg: 5, utils.AliasCfg: ""},
		},
	}
	if cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cgrCfg.kamAgentCfg.AsMapInterface(); !reflect.DeepEqual(rcv, eMap) {
		t.Errorf("Expected %+v \n, received %+v", eMap, rcv)
	}
}

func TestKamAgentCfgAsMapInterface1(t *testing.T) {
	cfgJSONStr := `{
	"kamailio_agent": {},
}`
	eMap := map[string]interface{}{
		utils.EnabledCfg:       false,
		utils.SessionSConnsCfg: []string{"*internal"},
		utils.CreateCdrCfg:     false,
		utils.TimezoneCfg:      "",
		utils.EvapiConnsCfg: []map[string]interface{}{
			{utils.AddressCfg: "127.0.0.1:8448", utils.ReconnectsCfg: 5, utils.AliasCfg: ""},
		},
	}
	if cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cgrCfg.kamAgentCfg.AsMapInterface(); !reflect.DeepEqual(rcv, eMap) {
		t.Errorf("Expected %+v \n, received %+v", eMap, rcv)
	}
}

func TestKamAgentCfgClone(t *testing.T) {
	ban := &KamAgentCfg{
		Enabled:       true,
		SessionSConns: []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaSessionS), "*conn1"},
		CreateCdr:     true,
		EvapiConns:    []*KamConnCfg{{Address: "127.0.0.1:8448", Reconnects: 10, Alias: "randomAlias"}},
		Timezone:      "Local",
	}
	rcv := ban.Clone()
	if !reflect.DeepEqual(ban, rcv) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(ban), utils.ToJSON(rcv))
	}
	if rcv.SessionSConns[1] = ""; ban.SessionSConns[1] != "*conn1" {
		t.Errorf("Expected clone to not modify the cloned")
	}
	if rcv.EvapiConns[0].Alias = ""; ban.EvapiConns[0].Alias != "randomAlias" {
		t.Errorf("Expected clone to not modify the cloned")
	}
}
