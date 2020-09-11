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

func TestApierCfgAsMapInterface1(t *testing.T) {
	cfgJSONStr := `{
	"apiers": {
		"caches_conns":[],
	},
}`
	sls := make([]string, 0)
	eMap := map[string]interface{}{
		utils.EnabledCfg:         false,
		utils.CachesConnsCfg:     sls,
		utils.SchedulerConnsCfg:  sls,
		utils.AttributeSConnsCfg: sls,
		utils.EEsConnsCfg:        sls,
	}
	if cgrCfg, err := NewCGRConfigFromJsonStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if newMap := cgrCfg.apier.AsMapInterface(); !reflect.DeepEqual(newMap, eMap) {
		t.Errorf("Expected %+v, recived %+v", eMap, newMap)
	}
}

func TestApierCfgAsMapInterface2(t *testing.T) {
	myJSONStr := `{
    "apiers": {
       "enabled": true,
       "attributes_conns": ["conn1", "conn2"],
       "ees_conns": ["*internal"],
    },
}`
	expectedMap := map[string]interface{}{
		utils.EnabledCfg:         true,
		utils.CachesConnsCfg:     []string{"*internal"},
		utils.SchedulerConnsCfg:  []string{},
		utils.AttributeSConnsCfg: []string{"conn1", "conn2"},
		utils.EEsConnsCfg:        []string{"*internal"},
	}
	if cgrCfg, err := NewCGRConfigFromJsonStringWithDefaults(myJSONStr); err != nil {
		t.Error(err)
	} else if newMap := cgrCfg.apier.AsMapInterface(); !reflect.DeepEqual(expectedMap, newMap) {
		t.Errorf("Expected %+v, recived %+v", expectedMap, newMap)
	}
}
