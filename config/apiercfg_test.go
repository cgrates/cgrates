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

func TestApierCfgloadFromJsonCfg(t *testing.T) {
	var aCfg, expected ApierCfg
	if err := aCfg.loadFromJsonCfg(nil); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(aCfg, expected) {
		t.Errorf("Expected: %+v ,recived: %+v", expected, aCfg)
	}
	if err := aCfg.loadFromJsonCfg(new(ApierJsonCfg)); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(aCfg, expected) {
		t.Errorf("Expected: %+v ,recived: %+v", expected, aCfg)
	}
	cfgJSONStr := `{
	"apiers": {
		"enabled": false,
		"caches_conns":["*internal"],
		"scheduler_conns": [],
		"attributes_conns": [],
	},
}`
	expected = ApierCfg{
		Enabled:         false,
		CachesConns:     []string{"*internal:*caches"},
		SchedulerConns:  []string{},
		AttributeSConns: []string{},
	}
	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnaCfg, err := jsnCfg.ApierCfgJson(); err != nil {
		t.Error(err)
	} else if err = aCfg.loadFromJsonCfg(jsnaCfg); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, aCfg) {
		t.Errorf("Expected: %+v , recived: %+v", expected, aCfg)
	}
}

func TestApierCfgAsMapInterface(t *testing.T) {
	var aCfg ApierCfg
	cfgJSONStr := `{
	"apiers": {
		"enabled": false,
		"caches_conns":[],
		"scheduler_conns": [],
		"attributes_conns": [],
	},
}`
	eMap := map[string]interface{}{
		"enabled":          false,
		"caches_conns":     []string{},
		"scheduler_conns":  []string{},
		"attributes_conns": []string{},
	}
	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnaCfg, err := jsnCfg.ApierCfgJson(); err != nil {
		t.Error(err)
	} else if err = aCfg.loadFromJsonCfg(jsnaCfg); err != nil {
		t.Error(err)
	} else if rcv := aCfg.AsMapInterface(); !reflect.DeepEqual(eMap, rcv) {
		t.Errorf("\nExpected: %+v\nRecived: %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}

	cfgJSONStr = `{
		"apiers": {
			"enabled": false,
			"caches_conns":["*internal"],
			"scheduler_conns": ["*internal"],
			"attributes_conns": ["*internal"],
		},
	}`
	eMap = map[string]interface{}{
		"enabled":          false,
		"caches_conns":     []string{"*internal"},
		"scheduler_conns":  []string{"*internal"},
		"attributes_conns": []string{"*internal"},
	}
	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnaCfg, err := jsnCfg.ApierCfgJson(); err != nil {
		t.Error(err)
	} else if err = aCfg.loadFromJsonCfg(jsnaCfg); err != nil {
		t.Error(err)
	} else if rcv := aCfg.AsMapInterface(); !reflect.DeepEqual(eMap, rcv) {
		t.Errorf("\nExpected: %+v\nRecived: %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}
}
