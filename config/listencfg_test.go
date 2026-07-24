// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package config

import (
	"reflect"
	"testing"

	"github.com/cgrates/cgrates/utils"
)

func TestListenCfgloadFromJsonCfg(t *testing.T) {
	var lstcfg, expected ListenCfg
	if err := lstcfg.loadFromJsonCfg(nil); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(lstcfg, expected) {
		t.Errorf("Expected: %+v ,received: %+v", expected, lstcfg)
	}
	if err := lstcfg.loadFromJsonCfg(new(ListenJsonCfg)); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(lstcfg, expected) {
		t.Errorf("Expected: %+v ,received: %+v", expected, lstcfg)
	}
	cfgJSONStr := `{
"listen": {
	"rpc_json": "127.0.0.1:2012",			// RPC JSON listening address
	"rpc_gob": "127.0.0.1:2013",			// RPC GOB listening address
	"http": "127.0.0.1:2080",				// HTTP listening address
	"rpc_json_tls" : "127.0.0.1:2022",		// RPC JSON TLS listening address
	"rpc_gob_tls": "127.0.0.1:2023",		// RPC GOB TLS listening address
	"http_tls": "127.0.0.1:2280",			// HTTP TLS listening address
	}
}`
	expected = ListenCfg{
		RPCJSONListen:    "127.0.0.1:2012",
		RPCGOBListen:     "127.0.0.1:2013",
		HTTPListen:       "127.0.0.1:2080",
		RPCJSONTLSListen: "127.0.0.1:2022",
		RPCGOBTLSListen:  "127.0.0.1:2023",
		HTTPTLSListen:    "127.0.0.1:2280",
	}
	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnLstCfg, err := jsnCfg.ListenJsonCfg(); err != nil {
		t.Error(err)
	} else if err = lstcfg.loadFromJsonCfg(jsnLstCfg); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, lstcfg) {
		t.Errorf("Expected: %+v , received: %+v", expected, lstcfg)
	}
}

func TestListenCfgAsMapInterface(t *testing.T) {
	str := "test"

	lstcfg := ListenCfg{
		RPCJSONListen:    str,
		RPCGOBListen:     str,
		HTTPListen:       str,
		RPCJSONTLSListen: str,
		RPCGOBTLSListen:  str,
		HTTPTLSListen:    str,
	}

	exp := map[string]any{
		utils.RPCJSONListenCfg:    lstcfg.RPCJSONListen,
		utils.RPCGOBListenCfg:     lstcfg.RPCGOBListen,
		utils.HTTPListenCfg:       lstcfg.HTTPListen,
		utils.RPCJSONTLSListenCfg: lstcfg.RPCJSONTLSListen,
		utils.RPCGOBTLSListenCfg:  lstcfg.RPCGOBTLSListen,
		utils.HTTPTLSListenCfg:    lstcfg.HTTPTLSListen,
	}

	rcv := lstcfg.AsMapInterface()

	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("Expected: %+v , received: %+v", exp, rcv)
	}
}
