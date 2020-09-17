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
)

func TestDispatcherHCfgloadFromJsonCfg(t *testing.T) {
	var daCfg, expected DispatcherHCfg
	if err := daCfg.loadFromJsonCfg(nil); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(daCfg, expected) {
		t.Errorf("Expected: %+v ,recived: %+v", expected, daCfg)
	}
	if err := daCfg.loadFromJsonCfg(new(DispatcherHJsonCfg)); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(daCfg, expected) {
		t.Errorf("Expected: %+v ,recived: %+v", expected, daCfg)
	}
	daCfg.Hosts = make(map[string][]*DispatcherHRegistarCfg)
	cfgJSONStr := `{
		"dispatcherh":{
			"enabled": true,
			"dispatchers_conns": ["conn1","conn2"],
			"hosts": {
				"*default": [
					{
						"ID": "Host1",
						"register_transport": "*json",
						"register_tls": false
					},
					{
						"ID": "Host2",
						"register_transport": "*gob",
						"register_tls": false
					}
				],
				"cgrates.net": [
					{
						"ID": "Host1",
						"register_transport": "*json",
						"register_tls": true
					},
					{
						"ID": "Host2",
						"register_transport": "*gob",
						"register_tls": true
					}
				]
			},
			"register_interval": "5m",
		},
}`
	expected = DispatcherHCfg{
		Enabled:          true,
		DispatchersConns: []string{"conn1", "conn2"},
		Hosts: map[string][]*DispatcherHRegistarCfg{
			utils.MetaDefault: {
				{
					ID:                "Host1",
					RegisterTransport: utils.MetaJSON,
				},
				{
					ID:                "Host2",
					RegisterTransport: utils.MetaGOB,
				},
			},
			"cgrates.net": {
				{
					ID:                "Host1",
					RegisterTransport: utils.MetaJSON,
					RegisterTLS:       true,
				},
				{
					ID:                "Host2",
					RegisterTransport: utils.MetaGOB,
					RegisterTLS:       true,
				},
			},
		},
		RegisterInterval: 5 * time.Minute,
	}
	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnDaCfg, err := jsnCfg.DispatcherHJsonCfg(); err != nil {
		t.Error(err)
	} else if err = daCfg.loadFromJsonCfg(jsnDaCfg); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, daCfg) {
		t.Errorf("Expected: %+v,\nRecived: %+v", utils.ToJSON(expected), utils.ToJSON(daCfg))
	}
}

func TestDispatcherHCfgAsMapInterface(t *testing.T) {
	cfgJSONStr := `{
		"dispatcherh":{
			"enabled": true,
			"dispatchers_conns": ["conn1","conn2"],
			"hosts": {
				"*default": [
					{
						"ID": "Host1",
						"register_transport": "*json",
						"register_tls": false
					},
					{
						"ID": "Host2",
						"register_transport": "*gob",
						"register_tls": false
					}
				]
			},
			"register_interval": "5m",
		},		
}`
	eMap := map[string]interface{}{
		utils.EnabledCfg:          true,
		utils.DispatchersConnsCfg: []string{"conn1", "conn2"},
		utils.HostsCfg: map[string][]map[string]interface{}{
			utils.MetaDefault: {
				{
					utils.IdCfg:                "Host1",
					utils.RegisterTransportCfg: "*json",
					utils.RegisterTLSCfg:       false,
				},
				{
					utils.IdCfg:                "Host2",
					utils.RegisterTransportCfg: "*gob",
					utils.RegisterTLSCfg:       false,
				},
			},
		},
		utils.RegisterIntervalCfg: 5 * time.Minute,
	}
	if cgrCfg, err := NewCGRConfigFromJsonStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cgrCfg.dispatcherHCfg.AsMapInterface(); !reflect.DeepEqual(eMap, rcv) {
		t.Errorf("Expected %+v, received %+v", eMap, rcv)
	}
}

func TestDispatcherHCfgAsMapInterface1(t *testing.T) {
	cfgJSONStr := `{
     "dispatcherh":{
          "enabled": true,
          "dispatchers_conns":["conn1"],
          "hosts": {
             "*default": [
             {
                  "ID":"",
                  "register_transport": "*json",
                  "register_tls":false,
             },
             {
                  "ID":"host2",
                  "register_transport": "",
                  "register_tls":true,
             },
          ]
          },
          "register_interval": "1m",
     },

}`
	eMap := map[string]interface{}{
		utils.EnabledCfg:          true,
		utils.DispatchersConnsCfg: []string{"conn1"},
		utils.HostsCfg: map[string][]map[string]interface{}{
			utils.MetaDefault: {
				{
					utils.IDCfg:                utils.EmptyString,
					utils.RegisterTransportCfg: utils.MetaJSON,
					utils.RegisterTLSCfg:       false,
				},
				{
					utils.IDCfg:                "host2",
					utils.RegisterTransportCfg: utils.EmptyString,
					utils.RegisterTLSCfg:       true,
				},
			},
		},
		utils.RegisterIntervalCfg: 1 * time.Minute,
	}
	if cgrCfg, err := NewCGRConfigFromJsonStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else {
		rcv := cgrCfg.dispatcherHCfg.AsMapInterface()
		if !reflect.DeepEqual(eMap[utils.HostsCfg].(map[string][]map[string]interface{})[utils.IdCfg],
			rcv[utils.HostsCfg].(map[string][]map[string]interface{})[utils.IdCfg]) {
			t.Errorf("Expected %+v, received %+v", eMap[utils.HostsCfg].(map[string][]map[string]interface{})[utils.IdCfg],
				rcv[utils.HostsCfg].(map[string][]map[string]interface{})[utils.IdCfg])
		} else if !reflect.DeepEqual(eMap[utils.HostsCfg], rcv[utils.HostsCfg]) {
			t.Errorf("Expected %+v, received %+v", eMap[utils.HostsCfg], rcv[utils.HostsCfg])
		} else if !reflect.DeepEqual(eMap, rcv) {
			t.Errorf("Expected %+v, received %+v", eMap, rcv)
		}
	}
}

func TestDispatcherHCfgAsMapInterface2(t *testing.T) {
	cfgJSONStr := `{
      "dispatcherh": {},
}`
	eMap := map[string]interface{}{
		utils.EnabledCfg:          false,
		utils.DispatchersConnsCfg: []string{},
		utils.HostsCfg:            map[string][]map[string]interface{}{},
		utils.RegisterIntervalCfg: 5 * time.Minute,
	}
	if cgrCfg, err := NewCGRConfigFromJsonStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cgrCfg.dispatcherHCfg.AsMapInterface(); !reflect.DeepEqual(eMap, rcv) {
		t.Errorf("Expected %+v, received %+v", eMap, rcv)
	}
}
