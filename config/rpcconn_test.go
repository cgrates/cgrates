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
	"time"

	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

func TestRPCConnsloadFromJsonCfgCase1(t *testing.T) {
	cfgJSON := &RPCConnsJson{
		Strategy: utils.StringPointer(utils.MetaFirst),
		PoolSize: utils.IntPointer(1),
		Conns: &[]*RemoteHostJson{
			{
				Address:         utils.StringPointer("127.0.0.1:2012"),
				Transport:       utils.StringPointer("*json"),
				Tls:             utils.BoolPointer(false),
				Key_path:        utils.StringPointer("key_path"),
				Cert_path:       utils.StringPointer("cert_path"),
				Ca_path:         utils.StringPointer("ca_path"),
				Conn_attempts:   utils.IntPointer(5),
				Reconnects:      utils.IntPointer(2),
				Connect_timeout: utils.StringPointer("1m"),
				Reply_timeout:   utils.StringPointer("1m"),
			},
		},
	}
	expected := RPCConns{
		utils.MetaBiJSONLocalHost: &RPCConn{
			Strategy: rpcclient.PoolFirst,
			PoolSize: 0,
			Conns: []*RemoteHost{{
				Address:   "127.0.0.1:2014",
				Transport: rpcclient.BiRPCJSON,
			},
			},
		},
		utils.MetaInternal: {
			Strategy: utils.MetaFirst,
			PoolSize: 0,
			Conns: []*RemoteHost{
				{
					Address:   utils.MetaInternal,
					Transport: utils.EmptyString,
					TLS:       false,
				},
			},
		},
		rpcclient.BiRPCInternal: {
			Strategy: utils.MetaFirst,
			PoolSize: 0,
			Conns: []*RemoteHost{
				{
					Address:   rpcclient.BiRPCInternal,
					Transport: utils.EmptyString,
					TLS:       false,
				},
			},
		},
		utils.MetaLocalHost: {
			Strategy: utils.MetaFirst,
			PoolSize: 1,
			Conns: []*RemoteHost{
				{
					Address:           "127.0.0.1:2012",
					Transport:         "*json",
					ConnectAttempts:   5,
					Reconnects:        2,
					ConnectTimeout:    1 * time.Minute,
					ReplyTimeout:      1 * time.Minute,
					TLS:               false,
					ClientKey:         "key_path",
					ClientCertificate: "cert_path",
					CaCertificate:     "ca_path",
				},
			},
		},
	}
	jsonCfg := NewDefaultCGRConfig()

	jsonCfg.rpcConns[utils.MetaLocalHost].loadFromJSONCfg(cfgJSON)
	if !reflect.DeepEqual(jsonCfg.rpcConns, expected) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(jsonCfg.rpcConns))
	}

}

func TestRPCConnsloadFromJsonCfgCase2(t *testing.T) {
	expected := RPCConns{
		utils.MetaBiJSONLocalHost: {
			Strategy: rpcclient.PoolFirst,
			PoolSize: 0,
			Conns: []*RemoteHost{{
				Address:   "127.0.0.1:2014",
				Transport: rpcclient.BiRPCJSON,
			},
			},
		},
		utils.MetaInternal: {
			Strategy: utils.MetaFirst,
			PoolSize: 0,
			Conns: []*RemoteHost{
				{
					Address:   utils.MetaInternal,
					Transport: utils.EmptyString,
					TLS:       false,
				},
			},
		},
		rpcclient.BiRPCInternal: {
			Strategy: utils.MetaFirst,
			PoolSize: 0,
			Conns: []*RemoteHost{
				{
					Address:   rpcclient.BiRPCInternal,
					Transport: utils.EmptyString,
					TLS:       false,
				},
			},
		},
		utils.MetaLocalHost: {
			Strategy: utils.MetaFirst,
			PoolSize: 0,
			Conns: []*RemoteHost{
				{
					Address:           "127.0.0.1:2012",
					Transport:         "*json",
					ConnectAttempts:   0,
					Reconnects:        0,
					ConnectTimeout:    0 * time.Minute,
					ReplyTimeout:      0 * time.Minute,
					TLS:               false,
					ClientKey:         "",
					ClientCertificate: "",
					CaCertificate:     "",
				},
			},
		},
	}
	jsonCfg := NewDefaultCGRConfig()
	jsonCfg.rpcConns[utils.MetaLocalHost].loadFromJSONCfg(nil)
	if !reflect.DeepEqual(expected, jsonCfg.rpcConns) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(expected), utils.ToJSON(jsonCfg.rpcConns))
	}
}

func TestRPCConnsAsMapInterface(t *testing.T) {
	cfgJSONStr := `{
		"rpc_conns": {
			"*localhost": {
				"conns": [
					{
						"address": "127.0.0.1:2012", 
						"transport":"*json",
						"id": "id_example",
						"tls": true,
						"key_path": "path_to_key",
						"cert_path": "path_to_cert",
						"ca_path":	"path_to_ca",
						"connect_attempts": 5,
						"reconnects": 3,
						"connect_timeout": "1m",
						"reply_timeout": "1m"
					}
				],
			},
		},	
}`
	eMap := map[string]interface{}{
		utils.MetaBiJSONLocalHost: map[string]interface{}{
			utils.PoolSize:    0,
			utils.StrategyCfg: utils.MetaFirst,
			utils.Conns: []map[string]interface{}{
				{
					utils.AddressCfg:   "127.0.0.1:2014",
					utils.TransportCfg: rpcclient.BiRPCJSON,
				},
			},
		},
		utils.MetaLocalHost: map[string]interface{}{
			utils.PoolSize:    0,
			utils.StrategyCfg: utils.MetaFirst,
			utils.Conns: []map[string]interface{}{
				{
					utils.AddressCfg:        "127.0.0.1:2012",
					utils.TransportCfg:      "*json",
					utils.IDCfg:             "id_example",
					utils.TLSNoCaps:         true,
					utils.KeyPathCgr:        "path_to_key",
					utils.CertPathCgr:       "path_to_cert",
					utils.CAPathCgr:         "path_to_ca",
					utils.ReconnectsCfg:     3,
					utils.ConnectTimeoutCfg: 1 * time.Minute,
					utils.ReplyTimeoutCfg:   1 * time.Minute,
				},
			},
		},
		utils.MetaInternal: map[string]interface{}{
			utils.StrategyCfg: utils.MetaFirst,
			utils.PoolSize:    0,
			utils.Conns: []map[string]interface{}{
				{
					utils.AddressCfg:   utils.MetaInternal,
					utils.TransportCfg: utils.EmptyString,
				},
			},
		},
		rpcclient.BiRPCInternal: map[string]interface{}{
			utils.StrategyCfg: utils.MetaFirst,
			utils.PoolSize:    0,
			utils.Conns: []map[string]interface{}{
				{
					utils.AddressCfg:   rpcclient.BiRPCInternal,
					utils.TransportCfg: utils.EmptyString,
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
                  {"address": "127.0.0.1:2018", "tls": true, "synchronous": true, "transport": "*json"},
             ],
             "poolSize": 2,
	      },
     },		
}`
	eMap := map[string]interface{}{
		utils.MetaBiJSONLocalHost: map[string]interface{}{
			utils.Conns: []map[string]interface{}{
				{
					utils.AddressCfg:   "127.0.0.1:2014",
					utils.TransportCfg: rpcclient.BiRPCJSON,
				},
			},
			utils.PoolSize:    0,
			utils.StrategyCfg: utils.MetaFirst,
		},
		utils.MetaInternal: map[string]interface{}{
			utils.Conns: []map[string]interface{}{
				{
					utils.AddressCfg:   utils.MetaInternal,
					utils.TransportCfg: utils.EmptyString,
				},
			},
			utils.PoolSize:    0,
			utils.StrategyCfg: utils.MetaFirst,
		},
		rpcclient.BiRPCInternal: map[string]interface{}{
			utils.Conns: []map[string]interface{}{
				{
					utils.AddressCfg:   rpcclient.BiRPCInternal,
					utils.TransportCfg: utils.EmptyString,
				},
			},
			utils.PoolSize:    0,
			utils.StrategyCfg: utils.MetaFirst,
		},
		utils.MetaLocalHost: map[string]interface{}{
			utils.Conns: []map[string]interface{}{
				{
					utils.TLSNoCaps:    true,
					utils.AddressCfg:   "127.0.0.1:2018",
					utils.TransportCfg: "*json",
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
					Address:   utils.MetaInternal,
					Transport: utils.EmptyString,
					TLS:       false,
				},
			},
		},
		utils.MetaLocalHost: {
			Strategy: utils.MetaFirst,
			PoolSize: 1,
			Conns: []*RemoteHost{
				{
					Address:           "127.0.0.1:2012",
					Transport:         "*json",
					ConnectAttempts:   0,
					Reconnects:        0,
					ConnectTimeout:    1 * time.Minute,
					ReplyTimeout:      1 * time.Minute,
					TLS:               false,
					ClientKey:         "",
					ClientCertificate: "",
					CaCertificate:     "",
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

func TestUpdateRPCCons(t *testing.T) {
	rpc := RPCConns{
		utils.MetaInternal: {
			Strategy: utils.MetaFirst,
			PoolSize: 0,
			Conns: []*RemoteHost{
				{
					ID:        "RPC1",
					Address:   utils.MetaInternal,
					Transport: utils.EmptyString,
					TLS:       false,
				},
				{
					ID:        "RPC2",
					Address:   utils.MetaInternal,
					Transport: utils.EmptyString,
					TLS:       true,
				},
			},
		},
	}

	newHosts := map[string]*RemoteHost{
		"RPC1": {
			ID:                "RPC1",
			Address:           utils.MetaInternal,
			Transport:         utils.EmptyString,
			ConnectAttempts:   2,
			Reconnects:        2,
			ConnectTimeout:    1 * time.Minute,
			ReplyTimeout:      1 * time.Minute,
			TLS:               true,
			ClientKey:         "key",
			ClientCertificate: "cert",
			CaCertificate:     "ca",
		},
	}
	expectedID := utils.StringSet{utils.MetaInternal: {}}
	expectedRPCCons := RPCConns{
		utils.MetaInternal: {
			Strategy: utils.MetaFirst,
			PoolSize: 0,
			Conns: []*RemoteHost{
				{
					ID:                "RPC1",
					Address:           utils.MetaInternal,
					Transport:         utils.EmptyString,
					ConnectAttempts:   2,
					Reconnects:        2,
					ConnectTimeout:    1 * time.Minute,
					ReplyTimeout:      1 * time.Minute,
					TLS:               true,
					ClientKey:         "key",
					ClientCertificate: "cert",
					CaCertificate:     "ca",
				},
				{
					ID:        "RPC2",
					Address:   utils.MetaInternal,
					Transport: utils.EmptyString,
					TLS:       true,
				},
			},
		},
	}

	if rcv := UpdateRPCCons(rpc, newHosts); !reflect.DeepEqual(rcv, expectedID) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(expectedID), utils.ToJSON(rcv))
	} else if !reflect.DeepEqual(rpc, expectedRPCCons) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(expectedRPCCons), utils.ToJSON(rpc))
	}
}

func TestRemoveRPCCons(t *testing.T) {
	rpc := RPCConns{
		utils.MetaInternal: {
			Strategy: utils.MetaFirst,
			PoolSize: 0,
			Conns: []*RemoteHost{
				{
					ID:        "RPC1",
					Address:   utils.MetaInternal,
					Transport: utils.EmptyString,
					TLS:       false,
				},
				{
					ID:                "RPC2",
					Address:           utils.MetaInternal,
					Transport:         utils.EmptyString,
					ConnectAttempts:   2,
					Reconnects:        2,
					ConnectTimeout:    1 * time.Minute,
					ReplyTimeout:      1 * time.Minute,
					TLS:               false,
					ClientKey:         "key",
					ClientCertificate: "cert",
					CaCertificate:     "ca",
				},
			},
		},
	}

	expectedID := utils.StringSet{utils.MetaInternal: {}}
	expectedRPCCons := RPCConns{
		utils.MetaInternal: {
			Strategy: utils.MetaFirst,
			PoolSize: 0,
			Conns: []*RemoteHost{
				{
					ID:      "RPC1",
					Address: utils.MetaInternal,
				},
				{
					ID: "RPC2",
				},
			},
		},
	}
	host := utils.StringSet{"RPC2": {}}
	if rcv := RemoveRPCCons(rpc, host); !reflect.DeepEqual(rcv, expectedID) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(expectedID), utils.ToJSON(rcv))
	} else if !reflect.DeepEqual(rpc, expectedRPCCons) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(expectedRPCCons), utils.ToJSON(rpc))
	}
}
