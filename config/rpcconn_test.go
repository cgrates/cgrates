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

func TestRPCConnsAsMapInterface(t *testing.T) {
	var cfg RPCConn
	cfgJSONStr := `{
		"rpc_conns": {
			"*localhost": {
				"conns": [{"address": "127.0.0.1:2012", "transport":"*json"}],
			},
		},	
}`
	eMap := map[string]interface{}{
		"poolSize": 0,
		"strategy": "",
		"conns": []map[string]interface{}{
			{
				"address":     "127.0.0.1:2012",
				"transport":   "*json",
				"synchronous": false,
				"TLS":         false,
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
		t.Errorf("\nExpected: %+v\nRecived: %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}
}

func TestRpcConnAsMapInterface(t *testing.T) {
	cfgJSONStr := `{
      "rpc_conns": {}
}`
	eMap := map[string]interface{}{
		utils.MetaInternal: map[string]interface{}{
			utils.Conns: []map[string]interface{}{
				{
					utils.TLS:            false,
					utils.AddressCfg:     utils.MetaInternal,
					utils.SynchronousCfg: false,
					utils.TransportCfg:   utils.EmptyString,
				},
			},
			utils.PoolSize:    0,
			utils.StrategyCfg: utils.MetaFirst,
		},
		utils.MetaLocalHost: map[string]interface{}{
			utils.Conns: []map[string]interface{}{
				{
					utils.TLS:            false,
					utils.AddressCfg:     "127.0.0.1:2012",
					utils.SynchronousCfg: false,
					utils.TransportCfg:   "*json",
				},
			},
			utils.PoolSize:    0,
			utils.StrategyCfg: utils.MetaFirst,
		},
	}
	if cgrCfg, err := NewCGRConfigFromJsonStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cgrCfg.rpcConns.AsMapInterface(); !reflect.DeepEqual(rcv, eMap) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}
}

func TestRpcConnAsMapInterface1(t *testing.T) {
	cfgJSONStr := `{
     "rpc_conns": {
	     "*localhost": {
		     "conns": [
                  {"address": "127.0.0.1:2018", "TLS": true, "synchronous": true, "transport": "*json"},
             ],
             "poolSize": 2,
	      },
     },		
}`
	eMap := map[string]interface{}{
		utils.MetaInternal: map[string]interface{}{
			utils.Conns: []map[string]interface{}{
				{
					utils.TLS:            false,
					utils.AddressCfg:     utils.MetaInternal,
					utils.SynchronousCfg: false,
					utils.TransportCfg:   utils.EmptyString,
				},
			},
			utils.PoolSize:    0,
			utils.StrategyCfg: utils.MetaFirst,
		},
		utils.MetaLocalHost: map[string]interface{}{
			utils.Conns: []map[string]interface{}{
				{
					utils.TLS:            true,
					utils.AddressCfg:     "127.0.0.1:2018",
					utils.SynchronousCfg: true,
					utils.TransportCfg:   "*json",
				},
			},
			utils.PoolSize:    2,
			utils.StrategyCfg: utils.MetaFirst,
		},
	}
	if cgrCfg, err := NewCGRConfigFromJsonStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cgrCfg.rpcConns.AsMapInterface(); !reflect.DeepEqual(rcv, eMap) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}
}
