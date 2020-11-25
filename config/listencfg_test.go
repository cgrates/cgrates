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

func TestListenCfgloadFromJsonCfg(t *testing.T) {
	jsonCfg := &ListenJsonCfg{
		Rpc_json:     utils.StringPointer("127.0.0.1:2012"),
		Rpc_gob:      utils.StringPointer("127.0.0.1:2013"),
		Http:         utils.StringPointer("127.0.0.1:2080"),
		Rpc_json_tls: utils.StringPointer("127.0.0.1:2022"),
		Rpc_gob_tls:  utils.StringPointer("127.0.0.1:2023"),
		Http_tls:     utils.StringPointer("127.0.0.1:2280"),
	}
	expected := &ListenCfg{
		RPCJSONListen:    "127.0.0.1:2012",
		RPCGOBListen:     "127.0.0.1:2013",
		HTTPListen:       "127.0.0.1:2080",
		RPCJSONTLSListen: "127.0.0.1:2022",
		RPCGOBTLSListen:  "127.0.0.1:2023",
		HTTPTLSListen:    "127.0.0.1:2280",
	}
	if jsnCfg, err := NewDefaultCGRConfig(); err != nil {
		t.Error(err)
	} else if err = jsnCfg.listenCfg.loadFromJSONCfg(jsonCfg); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, jsnCfg.listenCfg) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(jsnCfg.listenCfg))
	}
}

func TestListenCfgAsMapInterface(t *testing.T) {
	cfgJSONStr := `{
        "listen": {},
}`
	eMap := map[string]interface{}{
		utils.RPCJSONListenCfg:    "127.0.0.1:2012",
		utils.RPCGOBListenCfg:     "127.0.0.1:2013",
		utils.HTTPListenCfg:       "127.0.0.1:2080",
		utils.RPCJSONTLSListenCfg: "127.0.0.1:2022",
		utils.RPCGOBTLSListenCfg:  "127.0.0.1:2023",
		utils.HTTPTLSListenCfg:    "127.0.0.1:2280",
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
	}
}`
	eMap := map[string]interface{}{
		utils.RPCJSONListenCfg:    "127.0.0.1:2010",
		utils.RPCGOBListenCfg:     "127.0.0.1:2018",
		utils.HTTPListenCfg:       "127.0.0.1:2080",
		utils.RPCJSONTLSListenCfg: "127.0.0.1:2025",
		utils.RPCGOBTLSListenCfg:  "127.0.0.1:2001",
		utils.HTTPTLSListenCfg:    "127.0.0.1:2288",
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
	}
	rcv := ban.Clone()
	if !reflect.DeepEqual(ban, rcv) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(ban), utils.ToJSON(rcv))
	}
	if rcv.RPCJSONListen = ""; ban.RPCJSONListen != "127.0.0.1:2012" {
		t.Errorf("Expected clone to not modify the cloned")
	}
}
