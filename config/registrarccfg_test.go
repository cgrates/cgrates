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

func TestDispatcherHCfgloadFromJsonCfg(t *testing.T) {
	jsonCfg := &RegistrarCJsonCfgs{
		RPC: &RegistrarCJsonCfg{
			Registrars_conns: &[]string{"*conn1", "*conn2"},
			Hosts: map[string][]*RemoteHostJson{
				utils.MetaDefault: {
					{
						Id:        utils.StringPointer("Host1"),
						Transport: utils.StringPointer(utils.MetaJSON),
					},
					{
						Id:        utils.StringPointer("Host2"),
						Transport: utils.StringPointer(utils.MetaGOB),
					},
				},
				"cgrates.net": {
					{
						Id:        utils.StringPointer("Host1"),
						Transport: utils.StringPointer(utils.MetaJSON),
						Tls:       utils.BoolPointer(true),
					},
					{
						Id:        utils.StringPointer("Host2"),
						Transport: utils.StringPointer(utils.MetaGOB),
						Tls:       utils.BoolPointer(true),
					},
				},
			},
			Refresh_interval: utils.StringPointer("5"),
		},
		Dispatchers: &RegistrarCJsonCfg{
			Registrars_conns: &[]string{"*conn1", "*conn2"},
			Hosts: map[string][]*RemoteHostJson{
				utils.MetaDefault: {
					{
						Id:        utils.StringPointer("Host1"),
						Transport: utils.StringPointer(utils.MetaJSON),
					},
					{
						Id:        utils.StringPointer("Host2"),
						Transport: utils.StringPointer(utils.MetaGOB),
					},
				},
				"cgrates.net": {
					{
						Id:        utils.StringPointer("Host1"),
						Transport: utils.StringPointer(utils.MetaJSON),
						Tls:       utils.BoolPointer(true),
					},
					{
						Id:        utils.StringPointer("Host2"),
						Transport: utils.StringPointer(utils.MetaGOB),
						Tls:       utils.BoolPointer(true),
					},
				},
			},
			Refresh_interval: utils.StringPointer("5"),
		},
	}
	expected := &RegistrarCCfgs{
		RPC: &RegistrarCCfg{
			RegistrarSConns: []string{"*conn1", "*conn2"},
			Hosts: map[string][]*RemoteHost{
				utils.MetaDefault: {
					{
						ID:        "Host1",
						Transport: utils.MetaJSON,
					},
					{
						ID:        "Host2",
						Transport: utils.MetaGOB,
					},
				},
				"cgrates.net": {
					{
						ID:        "Host1",
						Transport: utils.MetaJSON,
						TLS:       true,
					},
					{
						ID:        "Host2",
						Transport: utils.MetaGOB,
						TLS:       true,
					},
				},
			},
			RefreshInterval: 5,
		},
		Dispatchers: &RegistrarCCfg{
			RegistrarSConns: []string{"*conn1", "*conn2"},
			Hosts: map[string][]*RemoteHost{
				utils.MetaDefault: {
					{
						ID:        "Host1",
						Transport: utils.MetaJSON,
					},
					{
						ID:        "Host2",
						Transport: utils.MetaGOB,
					},
				},
				"cgrates.net": {
					{
						ID:        "Host1",
						Transport: utils.MetaJSON,
						TLS:       true,
					},
					{
						ID:        "Host2",
						Transport: utils.MetaGOB,
						TLS:       true,
					},
				},
			},
			RefreshInterval: 5,
		},
	}
	jsnCfg := NewDefaultCGRConfig()
	if err = jsnCfg.registrarCCfg.loadFromJSONCfg(jsonCfg); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, jsnCfg.registrarCCfg) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(jsnCfg.registrarCCfg))
	}
}

func TestDispatcherHCfgAsMapInterface(t *testing.T) {
	cfgJSONStr := `{
		"registrarc":{
			"rpc":{
				"registrars_conns": ["*conn1","*conn2"],
				"hosts": {
					"*default": [
						{
							"ID": "Host1",
							"transport": "*json",
							"tls": false
						},
						{
							"ID": "Host2",
							"transport": "*gob",
							"tls": false
						}
					]
				},
				"refresh_interval": "0",
			},
			"dispatchers":{
				"enabled": true,
				"registrars_conns": ["*conn1","*conn2"],
				"hosts": {
					"*default": [
						{
							"ID": "Host1",
							"transport": "*json",
							"tls": false
						},
						{
							"ID": "Host2",
							"transport": "*gob",
							"tls": false
						}
					]
				},
				"refresh_interval": "0",
			},
		},		
}`
	eMap := map[string]interface{}{
		utils.RPCCfg: map[string]interface{}{
			utils.RegistrarsConnsCfg: []string{"*conn1", "*conn2"},
			utils.HostsCfg: map[string][]map[string]interface{}{
				utils.MetaDefault: {
					{
						utils.IDCfg:        "Host1",
						utils.TransportCfg: "*json",
					},
					{
						utils.IDCfg:        "Host2",
						utils.TransportCfg: "*gob",
					},
				},
			},
			utils.RefreshIntervalCfg: "0",
		},
		utils.DispatcherCfg: map[string]interface{}{
			utils.RegistrarsConnsCfg: []string{"*conn1", "*conn2"},
			utils.HostsCfg: map[string][]map[string]interface{}{
				utils.MetaDefault: {
					{
						utils.IDCfg:        "Host1",
						utils.TransportCfg: "*json",
					},
					{
						utils.IDCfg:        "Host2",
						utils.TransportCfg: "*gob",
					},
				},
			},
			utils.RefreshIntervalCfg: "0",
		},
	}
	if cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cgrCfg.registrarCCfg.AsMapInterface(); !reflect.DeepEqual(eMap, rcv) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}
}

func TestDispatcherCfgParseWithNanoSec(t *testing.T) {
	jsonCfg := &RegistrarCJsonCfgs{
		RPC: &RegistrarCJsonCfg{
			Refresh_interval: utils.StringPointer("1ss"),
		},
	}
	expErrMessage := "time: unknown unit \"ss\" in duration \"1ss\""
	jsnCfg := NewDefaultCGRConfig()
	if err = jsnCfg.registrarCCfg.loadFromJSONCfg(jsonCfg); err == nil || err.Error() != expErrMessage {
		t.Errorf("Expected %+v \n, recevied %+v", expErrMessage, err)
	}
}

func TestDispatcherCfgParseWithNanoSec2(t *testing.T) {
	jsonCfg := &RegistrarCJsonCfgs{
		Dispatchers: &RegistrarCJsonCfg{
			Refresh_interval: utils.StringPointer("1ss"),
		},
	}
	expErrMessage := "time: unknown unit \"ss\" in duration \"1ss\""
	jsnCfg := NewDefaultCGRConfig()
	if err = jsnCfg.registrarCCfg.loadFromJSONCfg(jsonCfg); err == nil || err.Error() != expErrMessage {
		t.Errorf("Expected %+v \n, recevied %+v", expErrMessage, err)
	}
}

func TestDispatcherHCfgAsMapInterface2(t *testing.T) {
	cfgJSONStr := `{
      "registrarc": {},
}`
	eMap := map[string]interface{}{
		utils.DispatcherCfg: map[string]interface{}{
			utils.RegistrarsConnsCfg: []string{},
			utils.HostsCfg:           map[string][]map[string]interface{}{},
			utils.RefreshIntervalCfg: "5m0s",
		},
		utils.RPCCfg: map[string]interface{}{
			utils.RegistrarsConnsCfg: []string{},
			utils.HostsCfg:           map[string][]map[string]interface{}{},
			utils.RefreshIntervalCfg: "5m0s",
		},
	}
	if cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cgrCfg.registrarCCfg.AsMapInterface(); !reflect.DeepEqual(eMap, rcv) {
		t.Errorf("Expected %+v, received %+v", eMap, rcv)
	}
}

func TestDispatcherHCfgClone(t *testing.T) {
	ban := &RegistrarCCfg{
		RegistrarSConns: []string{"*conn1", "*conn2"},
		Hosts: map[string][]*RemoteHost{
			utils.MetaDefault: {
				{
					ID:        "Host1",
					Transport: utils.MetaJSON,
				},
				{
					ID:        "Host2",
					Transport: utils.MetaGOB,
				},
			},
			"cgrates.net": {
				{
					ID:        "Host1",
					Transport: utils.MetaJSON,
					TLS:       true,
				},
				{
					ID:        "Host2",
					Transport: utils.MetaGOB,
					TLS:       true,
				},
			},
		},
		RefreshInterval: 5,
	}
	rcv := ban.Clone()
	if !reflect.DeepEqual(ban, rcv) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(ban), utils.ToJSON(rcv))
	}
	if rcv.RegistrarSConns[0] = ""; ban.RegistrarSConns[0] != "*conn1" {
		t.Errorf("Expected clone to not modify the cloned")
	}
	if rcv.Hosts[utils.MetaDefault][0].ID = ""; ban.Hosts[utils.MetaDefault][0].ID != "Host1" {
		t.Errorf("Expected clone to not modify the cloned")
	}
}
