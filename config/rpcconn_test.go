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
