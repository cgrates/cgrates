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
	cfgJSONStr := `{
		"dispatcherh":{
			"enabled": true,
			"dispatchers_conns": ["conn1","conn2"],
			"host_ids": ["HOST1","HOST2"],
			"register_interval": "5m",
			"register_transport": "*json",
		},
}`
	expected = DispatcherHCfg{
		Enabled:           true,
		DispatchersConns:  []string{"conn1", "conn2"},
		HostIDs:           []string{"HOST1", "HOST2"},
		RegisterInterval:  5 * time.Minute,
		RegisterTransport: utils.MetaJSON,
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
	cfgJSONStr := `{
		"dispatcherh":{
			"enabled": true,
			"dispatchers_conns": ["conn1","conn2"],
			"host_ids": ["HOST1","HOST2"],
			"register_interval": "5m",
			"register_transport": "*json",
		},		
}`
	eMap := map[string]interface{}{
		"enabled":            true,
		"dispatchers_conns":  []string{"conn1", "conn2"},
		"host_ids":           []string{"HOST1", "HOST2"},
		"register_interval":  5 * time.Minute,
		"register_transport": "*json",
	}
	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnDaCfg, err := jsnCfg.DispatcherHJsonCfg(); err != nil {
		t.Error(err)
	} else if err = daCfg.loadFromJsonCfg(jsnDaCfg); err != nil {
		t.Error(err)
	} else if rcv := daCfg.AsMapInterface(); !reflect.DeepEqual(eMap, rcv) {
		t.Errorf("\nExpected: %+v\nRecived: %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}
}
