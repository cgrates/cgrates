// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package config

import (
	"reflect"
	"testing"

	"github.com/cgrates/cgrates/utils"
)

func TestListenCfgloadFromJsonCfg(t *testing.T) {
	jsonCfg := &ListenJsonCfg{
		Rpc_json:     utils.StringPointer("127.0.0.1:2012"),
		Rpc_gob:      utils.StringPointer("127.0.0.1:2013"),
		Http:         utils.StringPointer("127.0.0.1:2080"),
		Rpc_json_tls: utils.StringPointer("127.0.0.1:2022"),
		Rpc_gob_tls:  utils.StringPointer("127.0.0.1:2023"),
		Http_tls:     utils.StringPointer("127.0.0.1:2280"),
		Birpc_json:   utils.StringPointer("127.0.0.1:2018"),
	}
	expected := &ListenCfg{
		RPCJSONListen:    "127.0.0.1:2012",
		RPCGOBListen:     "127.0.0.1:2013",
		HTTPListen:       "127.0.0.1:2080",
		RPCJSONTLSListen: "127.0.0.1:2022",
		RPCGOBTLSListen:  "127.0.0.1:2023",
		HTTPTLSListen:    "127.0.0.1:2280",
		BiJSONListen:     "127.0.0.1:2018",
	}
	jsnCfg := NewDefaultCGRConfig()
	if err := jsnCfg.listenCfg.loadFromJSONCfg(jsonCfg); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, jsnCfg.listenCfg) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(jsnCfg.listenCfg))
	}
}

func TestListenCfgAsMapInterface(t *testing.T) {
	cfgJSONStr := `{
        "listen": {},
}`
	eMap := map[string]any{
		utils.RPCJSONListenCfg:    "127.0.0.1:2012",
		utils.RPCGOBListenCfg:     "127.0.0.1:2013",
		utils.HTTPListenCfg:       "127.0.0.1:2080",
		utils.RPCJSONTLSListenCfg: "127.0.0.1:2022",
		utils.RPCGOBTLSListenCfg:  "127.0.0.1:2023",
		utils.HTTPTLSListenCfg:    "127.0.0.1:2280",
		utils.BijsonListenCfg:     "127.0.0.1:2014",
		utils.BigobListenCfg:      "",
	}
	if cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cgrCfg.listenCfg.AsMapInterface(); !reflect.DeepEqual(eMap, rcv) {
		t.Errorf("Expected %+v, received %+v", eMap, rcv)
	}
}

func TestListenCfgAsMapInterface1(t *testing.T) {
	cfgJSONStr := `{
	"listen": {
		"rpc_json": "127.0.0.1:2010",			
        "rpc_gob": "127.0.0.1:2018",			
        "rpc_json_tls" : "127.0.0.1:2025",		
        "rpc_gob_tls": "127.0.0.1:2001",		
        "http_tls": "127.0.0.1:2288",			
        "birpc_json": "127.0.0.1:2014",			
	}
}`
	eMap := map[string]any{
		utils.RPCJSONListenCfg:    "127.0.0.1:2010",
		utils.RPCGOBListenCfg:     "127.0.0.1:2018",
		utils.HTTPListenCfg:       "127.0.0.1:2080",
		utils.RPCJSONTLSListenCfg: "127.0.0.1:2025",
		utils.RPCGOBTLSListenCfg:  "127.0.0.1:2001",
		utils.HTTPTLSListenCfg:    "127.0.0.1:2288",
		utils.BijsonListenCfg:     "127.0.0.1:2014",
		utils.BigobListenCfg:      "",
	}
	if cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cgrCfg.listenCfg.AsMapInterface(); !reflect.DeepEqual(eMap, rcv) {
		t.Errorf("Expected %+v, received %+v", eMap, rcv)
	}
}

func TestListenCfgClone(t *testing.T) {
	ban := &ListenCfg{
		RPCJSONListen:    "127.0.0.1:2012",
		RPCGOBListen:     "127.0.0.1:2013",
		HTTPListen:       "127.0.0.1:2080",
		RPCJSONTLSListen: "127.0.0.1:2022",
		RPCGOBTLSListen:  "127.0.0.1:2023",
		HTTPTLSListen:    "127.0.0.1:2280",
		BiJSONListen:     "127.0.0.1:2018",
	}
	rcv := ban.Clone()
	if !reflect.DeepEqual(ban, rcv) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(ban), utils.ToJSON(rcv))
	}
	if rcv.RPCJSONListen = ""; ban.RPCJSONListen != "127.0.0.1:2012" {
		t.Errorf("Expected clone to not modify the cloned")
	}

	ban = nil
	rcv = ban.Clone()
	if !reflect.DeepEqual(ban, rcv) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(ban), utils.ToJSON(rcv))
	}
}
