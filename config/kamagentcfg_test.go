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
	"github.com/cgrates/rpcclient"
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
	jsnCfg := NewDefaultCGRConfig()
	if err = jsnCfg.kamAgentCfg.loadFromJSONCfg(cfgJSON); err != nil {
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
			"sessions_conns": ["*birpc_internal", "*conn1","*conn2", "*internal"],
			"create_cdr": true,
			"timezone": "UTC",
			"evapi_conns":[
				{"address": "127.0.0.1:8448", "reconnects": 5, "alias": ""}
			],
		},
	}`
	eMap := map[string]interface{}{
		utils.EnabledCfg:       false,
		utils.SessionSConnsCfg: []string{rpcclient.BiRPCInternal, "*conn1", "*conn2", utils.MetaInternal},
		utils.CreateCdrCfg:     true,
		utils.TimezoneCfg:      "UTC",
		utils.EvapiConnsCfg: []map[string]interface{}{
			{utils.AddressCfg: "127.0.0.1:8448", utils.ReconnectsCfg: 5, utils.AliasCfg: ""},
		},
	}
	if cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cgrCfg.kamAgentCfg.AsMapInterface(""); !reflect.DeepEqual(rcv, eMap) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}
}

func TestKamAgentCfgAsMapInterface1(t *testing.T) {
	cfgJSONStr := `{
	"kamailio_agent": {},
}`
	eMap := map[string]interface{}{
		utils.EnabledCfg:       false,
		utils.SessionSConnsCfg: []string{rpcclient.BiRPCInternal},
		utils.CreateCdrCfg:     false,
		utils.TimezoneCfg:      "",
		utils.EvapiConnsCfg: []map[string]interface{}{
			{utils.AddressCfg: "127.0.0.1:8448", utils.ReconnectsCfg: 5, utils.AliasCfg: ""},
		},
	}
	if cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cgrCfg.kamAgentCfg.AsMapInterface(""); !reflect.DeepEqual(rcv, eMap) {
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

func TestDiffKamConnJsonCfg(t *testing.T) {
	v1 := &KamConnCfg{
		Alias:      "KAM",
		Address:    "localhost:8080",
		Reconnects: 2,
	}

	v2 := &KamConnCfg{
		Alias:      "KAM_2",
		Address:    "localhost:8037",
		Reconnects: 5,
	}

	expected := &KamConnJsonCfg{
		Alias:      utils.StringPointer("KAM_2"),
		Address:    utils.StringPointer("localhost:8037"),
		Reconnects: utils.IntPointer(5),
	}

	rcv := diffKamConnJsonCfg(v1, v2)
	if !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}

	v1 = v2
	expected = &KamConnJsonCfg{}

	rcv = diffKamConnJsonCfg(v1, v2)
	if !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}

func TestEqualsKamConnsCfg(t *testing.T) {
	v1 := []*KamConnCfg{
		{
			Alias:      "KAM",
			Address:    "localhost:8080",
			Reconnects: 2,
		},
	}

	v2 := []*KamConnCfg{
		{
			Alias:      "KAM_2",
			Address:    "localhost:8037",
			Reconnects: 5,
		},
	}

	if equalsKamConnsCfg(v1, v2) {
		t.Error("Conns should not match")
	}

	v2 = []*KamConnCfg{
		{
			Alias:      "KAM",
			Address:    "localhost:8080",
			Reconnects: 2,
		},
	}

	if !equalsKamConnsCfg(v1, v2) {
		t.Error("Conns should match")
	}

	v2 = []*KamConnCfg{}
	if equalsKamConnsCfg(v1, v2) {
		t.Error("Conns should not match")
	}
}

func TestDiffKamAgentJsonCfg(t *testing.T) {
	var d *KamAgentJsonCfg

	v1 := &KamAgentCfg{
		Enabled:       false,
		SessionSConns: []string{"*localhost"},
		CreateCdr:     false,
		EvapiConns: []*KamConnCfg{
			{
				Alias:      "KAM_2",
				Address:    "localhost:8037",
				Reconnects: 5,
			},
		},
		Timezone: "UTC",
	}

	v2 := &KamAgentCfg{
		Enabled:       true,
		SessionSConns: []string{"*birpc"},
		CreateCdr:     true,
		EvapiConns: []*KamConnCfg{
			{
				Alias:      "KAM_1",
				Address:    "localhost:8080",
				Reconnects: 2,
			},
		},
		Timezone: "EEST",
	}

	expected := &KamAgentJsonCfg{
		Enabled:        utils.BoolPointer(true),
		Sessions_conns: &[]string{"*birpc"},
		Create_cdr:     utils.BoolPointer(true),
		Evapi_conns: &[]*KamConnJsonCfg{
			{
				Alias:      utils.StringPointer("KAM_1"),
				Address:    utils.StringPointer("localhost:8080"),
				Reconnects: utils.IntPointer(2),
			},
		},
		Timezone: utils.StringPointer("EEST"),
	}

	rcv := diffKamAgentJsonCfg(d, v1, v2)
	if !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}

	v1 = v2
	expected = &KamAgentJsonCfg{}

	rcv = diffKamAgentJsonCfg(d, v1, v2)
	if !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}
