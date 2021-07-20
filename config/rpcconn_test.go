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
	cfgJSON := &RPCConnJson{
		Strategy: utils.StringPointer(utils.MetaFirst),
		PoolSize: utils.IntPointer(1),
		Conns: &[]*RemoteHostJson{
			{
				Address:            utils.StringPointer("127.0.0.1:2012"),
				Transport:          utils.StringPointer("*json"),
				Tls:                utils.BoolPointer(false),
				Client_key:         utils.StringPointer("key_path"),
				Client_certificate: utils.StringPointer("cert_path"),
				Ca_certificate:     utils.StringPointer("ca_path"),
				Connect_attempts:   utils.IntPointer(5),
				Reconnects:         utils.IntPointer(2),
				Connect_timeout:    utils.StringPointer("1m"),
				Reply_timeout:      utils.StringPointer("1m"),
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
					TLS:               false,
					ConnectAttempts:   0,
					Reconnects:        0,
					ConnectTimeout:    0 * time.Minute,
					ReplyTimeout:      0 * time.Minute,
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
					utils.AddressCfg:   "127.0.0.1:2012",
					utils.TransportCfg: "*json",
					utils.IDCfg:        "id_example",
					utils.TLSNoCaps:    true,
					// utils.KeyPathCgr:         "path_to_key",
					// utils.CertPathCgr:        "path_to_cert",
					// utils.CAPathCgr:          "path_to_ca",
					utils.ConnectAttemptsCfg: 5,
					utils.ReconnectsCfg:      3,
					utils.ConnectTimeoutCfg:  1 * time.Minute,
					utils.ReplyTimeoutCfg:    1 * time.Minute,
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
                  {"address": "127.0.0.1:2018", "tls": false, "transport": "*json"},
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

func TestDiffRPCConnJson(t *testing.T) {
	var d *RPCConnJson

	v1 := &RPCConn{
		Strategy: utils.MetaTopUpReset,
		PoolSize: 2,
		Conns: []*RemoteHost{
			{
				ID:                "host1_ID",
				Address:           "127.0.0.1:8080",
				Transport:         "tcp",
				ConnectAttempts:   2,
				Reconnects:        5,
				ConnectTimeout:    1 * time.Minute,
				ReplyTimeout:      1 * time.Minute,
				TLS:               false,
				ClientKey:         "key1",
				ClientCertificate: "path1",
				CaCertificate:     "ca_path1",
			},
		},
	}

	v2 := &RPCConn{
		Strategy: "*disconnect",
		PoolSize: 3,
		Conns: []*RemoteHost{
			{
				ID:                "host2_ID",
				Address:           "0.0.0.0:8080",
				Transport:         "udp",
				ConnectAttempts:   3,
				Reconnects:        4,
				ConnectTimeout:    2 * time.Minute,
				ReplyTimeout:      2 * time.Minute,
				TLS:               true,
				ClientKey:         "key2",
				ClientCertificate: "path2",
				CaCertificate:     "ca_path2",
			},
		},
	}

	expected := &RPCConnJson{
		Strategy: utils.StringPointer("*disconnect"),
		PoolSize: utils.IntPointer(3),
		Conns: &[]*RemoteHostJson{
			{
				Id:                 utils.StringPointer("host2_ID"),
				Address:            utils.StringPointer("0.0.0.0:8080"),
				Transport:          utils.StringPointer("udp"),
				Tls:                utils.BoolPointer(true),
				Client_key:         utils.StringPointer("key2"),
				Client_certificate: utils.StringPointer("path2"),
				Ca_certificate:     utils.StringPointer("ca_path2"),
				Connect_attempts:   utils.IntPointer(3),
				Reconnects:         utils.IntPointer(4),
				Connect_timeout:    utils.StringPointer("2m0s"),
				Reply_timeout:      utils.StringPointer("2m0s"),
			},
		},
	}

	rcv := diffRPCConnJson(d, v1, v2)
	if !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}

	v1 = v2
	expected = &RPCConnJson{
		Conns: &[]*RemoteHostJson{
			{
				Id:                 utils.StringPointer("host2_ID"),
				Address:            utils.StringPointer("0.0.0.0:8080"),
				Transport:          utils.StringPointer("udp"),
				Tls:                utils.BoolPointer(true),
				Client_key:         utils.StringPointer("key2"),
				Client_certificate: utils.StringPointer("path2"),
				Ca_certificate:     utils.StringPointer("ca_path2"),
				Connect_attempts:   utils.IntPointer(3),
				Reconnects:         utils.IntPointer(4),
				Connect_timeout:    utils.StringPointer("2m0s"),
				Reply_timeout:      utils.StringPointer("2m0s"),
			},
		},
	}
	rcv = diffRPCConnJson(d, v1, v2)
	if !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}

func TestEqualsRemoteHosts(t *testing.T) {
	v1 := []*RemoteHost{
		{
			ID:                "host1_ID",
			Address:           "127.0.0.1:8080",
			Transport:         "tcp",
			ConnectAttempts:   2,
			Reconnects:        5,
			ConnectTimeout:    1 * time.Minute,
			ReplyTimeout:      1 * time.Minute,
			TLS:               false,
			ClientKey:         "key1",
			ClientCertificate: "path1",
			CaCertificate:     "ca_path1",
		},
	}

	v2 := []*RemoteHost{
		{
			ID:                "host2_ID",
			Address:           "0.0.0.0:8080",
			Transport:         "udp",
			ConnectAttempts:   3,
			Reconnects:        4,
			ConnectTimeout:    2 * time.Minute,
			ReplyTimeout:      2 * time.Minute,
			TLS:               true,
			ClientKey:         "key2",
			ClientCertificate: "path2",
			CaCertificate:     "ca_path2",
		},
	}

	if equalsRemoteHosts(v1, v2) {
		t.Error("Hosts should not match")
	}

	v1 = v2
	if !equalsRemoteHosts(v1, v2) {
		t.Error("Hosts should match")
	}

	v2 = []*RemoteHost{}
	if equalsRemoteHosts(v1, v2) {
		t.Error("Hosts should not match")
	}
}

func TestEqualsRPCConn(t *testing.T) {
	v1 := &RPCConn{
		Strategy: utils.MetaTopUpReset,
		PoolSize: 2,
		Conns: []*RemoteHost{
			{
				ID:                "host1_ID",
				Address:           "127.0.0.1:8080",
				Transport:         "tcp",
				ConnectAttempts:   2,
				Reconnects:        5,
				ConnectTimeout:    1 * time.Minute,
				ReplyTimeout:      1 * time.Minute,
				TLS:               false,
				ClientKey:         "key1",
				ClientCertificate: "path1",
				CaCertificate:     "ca_path1",
			},
		},
	}

	v2 := &RPCConn{
		Strategy: "*disconnect",
		PoolSize: 3,
		Conns: []*RemoteHost{
			{
				ID:                "host2_ID",
				Address:           "0.0.0.0:8080",
				Transport:         "udp",
				ConnectAttempts:   3,
				Reconnects:        4,
				ConnectTimeout:    2 * time.Minute,
				ReplyTimeout:      2 * time.Minute,
				TLS:               true,
				ClientKey:         "key2",
				ClientCertificate: "path2",
				CaCertificate:     "ca_path2",
			},
		},
	}

	if equalsRPCConn(v1, v2) {
		t.Error("Conns should not match")
	}

	v1 = v2
	if !equalsRPCConn(v1, v2) {
		t.Error("Conns should match")
	}

	v2 = &RPCConn{}
	if equalsRPCConn(v1, v2) {
		t.Error("Conns should not match")
	}
}

func TestDiffRPCConnsJson(t *testing.T) {
	var d RPCConnsJson

	v1 := RPCConns{
		"CONN_1": {
			Strategy: utils.MetaTopUpReset,
			PoolSize: 2,
			Conns: []*RemoteHost{
				{
					ID:                "host1_ID",
					Address:           "127.0.0.1:8080",
					Transport:         "tcp",
					ConnectAttempts:   2,
					Reconnects:        5,
					ConnectTimeout:    1 * time.Minute,
					ReplyTimeout:      1 * time.Minute,
					TLS:               false,
					ClientKey:         "key1",
					ClientCertificate: "path1",
					CaCertificate:     "ca_path1",
				},
			},
		},
	}

	v2 := RPCConns{
		"CONN_1": {
			Strategy: "*disconnect",
			PoolSize: 3,
			Conns: []*RemoteHost{
				{
					ID:                "host2_ID",
					Address:           "0.0.0.0:8080",
					Transport:         "udp",
					ConnectAttempts:   3,
					Reconnects:        4,
					ConnectTimeout:    2 * time.Minute,
					ReplyTimeout:      2 * time.Minute,
					TLS:               true,
					ClientKey:         "key2",
					ClientCertificate: "path2",
					CaCertificate:     "ca_path2",
				},
			},
		},
	}

	expected := RPCConnsJson{
		"CONN_1": {
			Strategy: utils.StringPointer("*disconnect"),
			PoolSize: utils.IntPointer(3),
			Conns: &[]*RemoteHostJson{
				{
					Id:                 utils.StringPointer("host2_ID"),
					Address:            utils.StringPointer("0.0.0.0:8080"),
					Transport:          utils.StringPointer("udp"),
					Tls:                utils.BoolPointer(true),
					Client_key:         utils.StringPointer("key2"),
					Client_certificate: utils.StringPointer("path2"),
					Ca_certificate:     utils.StringPointer("ca_path2"),
					Connect_attempts:   utils.IntPointer(3),
					Reconnects:         utils.IntPointer(4),
					Connect_timeout:    utils.StringPointer("2m0s"),
					Reply_timeout:      utils.StringPointer("2m0s"),
				},
			},
		},
	}

	rcv := diffRPCConnsJson(d, v1, v2)
	if !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}

	v1 = v2
	expected = RPCConnsJson{}
	rcv = diffRPCConnsJson(d, v1, v2)
	if !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}

func TestNewDfltRemoteHost(t *testing.T) {
	dfltRemoteHost = nil
	rcv := NewDfltRemoteHost()
	if !reflect.DeepEqual(rcv, new(RemoteHost)) {
		t.Errorf("Expected %v \n but received \n %v", new(RemoteHost), rcv)
	}
}
