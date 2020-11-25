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

func TestRPCConnsloadFromJsonCfgCase1(t *testing.T) {
	cfgJSON := &RPCConnsJson{
		Strategy: utils.StringPointer(utils.MetaFirst),
		PoolSize: utils.IntPointer(1),
		Conns: &[]*RemoteHostJson{
			{
				Address:     utils.StringPointer("127.0.0.1:2012"),
				Transport:   utils.StringPointer("*json"),
				Synchronous: utils.BoolPointer(false),
				Tls:         utils.BoolPointer(false),
			},
		},
	}
	expected := RPCConns{
		utils.MetaInternal: {
			Strategy: utils.MetaFirst,
			PoolSize: 0,
			Conns: []*RemoteHost{
				{
					Address:     utils.MetaInternal,
					Transport:   utils.EmptyString,
					Synchronous: false,
					TLS:         false,
				},
			},
		},
		utils.MetaLocalHost: {
			Strategy: utils.MetaFirst,
			PoolSize: 1,
			Conns: []*RemoteHost{
				{
					Address:     "127.0.0.1:2012",
					Transport:   "*json",
					Synchronous: false,
					TLS:         false,
				},
			},
		},
	}
	if jsonCfg, err := NewDefaultCGRConfig(); err != nil {
		t.Error(err)
	} else {
		jsonCfg.rpcConns[utils.MetaLocalHost].loadFromJSONCfg(cfgJSON)
		if !reflect.DeepEqual(jsonCfg.rpcConns, expected) {
			t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(jsonCfg.rpcConns))
		}
	}
}

func TestRPCConnsloadFromJsonCfgCase2(t *testing.T) {
	expected := RPCConns{
		utils.MetaInternal: {
			Strategy: utils.MetaFirst,
			PoolSize: 0,
			Conns: []*RemoteHost{
				{
					Address:     utils.MetaInternal,
					Transport:   utils.EmptyString,
					Synchronous: false,
					TLS:         false,
				},
			},
		},
		utils.MetaLocalHost: {
			Strategy: utils.MetaFirst,
			PoolSize: 0,
			Conns: []*RemoteHost{
				{
					Address:     "127.0.0.1:2012",
					Transport:   "*json",
					Synchronous: false,
					TLS:         false,
				},
			},
		},
	}
	jsonCfg, err := NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
	}
	jsonCfg.rpcConns[utils.MetaLocalHost].loadFromJSONCfg(nil)
	if !reflect.DeepEqual(expected, jsonCfg.rpcConns) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(expected), utils.ToJSON(jsonCfg.rpcConns))
	}
}

func TestRPCConnsAsMapInterface(t *testing.T) {
	cfgJSONStr := `{
		"rpc_conns": {
			"*localhost": {
				"conns": [{"address": "127.0.0.1:2012", "transport":"*json"}],
			},
		},	
}`
	eMap := map[string]interface{}{
		utils.MetaLocalHost: map[string]interface{}{
			utils.PoolSize:    0,
			utils.StrategyCfg: utils.MetaFirst,
			utils.Conns: []map[string]interface{}{
				{
					utils.AddressCfg:     "127.0.0.1:2012",
					utils.TransportCfg:   "*json",
					utils.SynchronousCfg: false,
					utils.TLS:            false,
				},
			},
		},
		utils.MetaInternal: map[string]interface{}{
			utils.StrategyCfg: utils.MetaFirst,
			utils.PoolSize:    0,
			utils.Conns: []map[string]interface{}{
				{
					utils.AddressCfg:     utils.MetaInternal,
					utils.TransportCfg:   utils.EmptyString,
					utils.SynchronousCfg: false,
					utils.TLS:            false,
				},
			},
		},
	}
	if cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cgrCfg.rpcConns.AsMapInterface(); !reflect.DeepEqual(eMap, rcv) {
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
	if cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cgrCfg.rpcConns.AsMapInterface(); !reflect.DeepEqual(rcv, eMap) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}
}

func TestRPCConnsClone(t *testing.T) {
	ban := RPCConns{
		utils.MetaInternal: {
			Strategy: utils.MetaFirst,
			PoolSize: 0,
			Conns: []*RemoteHost{
				{
					Address:     utils.MetaInternal,
					Transport:   utils.EmptyString,
					Synchronous: false,
					TLS:         false,
				},
			},
		},
		utils.MetaLocalHost: {
			Strategy: utils.MetaFirst,
			PoolSize: 1,
			Conns: []*RemoteHost{
				{
					Address:     "127.0.0.1:2012",
					Transport:   "*json",
					Synchronous: false,
					TLS:         false,
				},
			},
		},
	}
	rcv := ban.Clone()
	if !reflect.DeepEqual(ban, rcv) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(ban), utils.ToJSON(rcv))
	}
	if rcv[utils.MetaInternal].Conns[0].Address = ""; ban[utils.MetaInternal].Conns[0].Address != utils.MetaInternal {
		t.Errorf("Expected clone to not modify the cloned")
	}
}
