// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package config

import (
	"reflect"
	"testing"

	"github.com/cgrates/cgrates/utils"
)

func TestRPCConnsAsMapInterface(t *testing.T) {
	var cfg RPCConn
	cfgJSONStr := `{
		"rpc_conns": {
			"*localhost": {
				"conns": [{"address": "127.0.0.1:2012", "transport":"*json"}],
			},
		},	
}`
	eMap := map[string]any{
		"poolSize": 0,
		"strategy": "",
		"conns": []map[string]any{
			{
				"address":     "127.0.0.1:2012",
				"transport":   "*json",
				"synchronous": false,
				"tls":         false,
			},
		},
	}
	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnRPCCfg, err := jsnCfg.RPCConnJsonCfg(); err != nil {
		t.Error(err)
	} else if err = cfg.loadFromJsonCfg(jsnRPCCfg["*localhost"]); err != nil {
		t.Error(err)
	} else if rcv := cfg.AsMapInterface(); !reflect.DeepEqual(eMap, rcv) {
		t.Errorf("\nExpected: %+v\nReceived: %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}
}

func TestRPCConnloadFromJsonCfg(t *testing.T) {
	str := "test"
	str2 := "test2"
	nm := 1
	nm2 := 2
	bl := false
	rh := &RemoteHost{
		ID:          str,
		Address:     str,
		Transport:   str,
		Synchronous: true,
		TLS:         true,
	}
	rC := &RPCConn{
		Strategy: str,
		PoolSize: nm,
		Conns:    []*RemoteHost{rh},
	}
	jsnCfg := &RPCConnsJson{
		Strategy: &str2,
		PoolSize: &nm2,
		Conns: &[]*RemoteHostJson{{
			Id:          &str2,
			Address:     &str2,
			Transport:   &str2,
			Synchronous: &bl,
			Tls:         &bl,
		}},
	}

	err := rC.loadFromJsonCfg(jsnCfg)
	rh2 := &RemoteHost{
		ID:          str2,
		Address:     str2,
		Transport:   str2,
		Synchronous: false,
		TLS:         false,
	}
	exp := &RPCConn{
		Strategy: str2,
		PoolSize: nm2,
		Conns:    []*RemoteHost{rh2},
	}
	if err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(rC, exp) {
		t.Errorf("expected %s, received %s", utils.ToJSON(exp), utils.ToJSON(rC))
	}

	err = rC.loadFromJsonCfg(nil)
	if err != nil {
		t.Error(err)
	}
}
