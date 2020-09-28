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

func TestFilterSCfgloadFromJsonCfg(t *testing.T) {
	cfgJSONS := &FilterSJsonCfg{
		Stats_conns:     &[]string{utils.MetaInternal},
		Resources_conns: &[]string{utils.MetaInternal},
		Apiers_conns:    &[]string{utils.MetaInternal},
	}
	expected := &FilterSCfg{
		StatSConns:     []string{"*internal:*stats"},
		ResourceSConns: []string{"*internal:*resources"},
		ApierSConns:    []string{"*internal:*apier"},
	}
	if jsnCfg, err := NewDefaultCGRConfig(); err != nil {
		t.Error(err)
	} else if err = jsnCfg.filterSCfg.loadFromJsonCfg(cfgJSONS); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, jsnCfg.filterSCfg) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(jsnCfg.filterSCfg))
	}
}

func TestFilterSCfgAsMapInterface(t *testing.T) {
	cfgJSONStr := `{
		"filters": {								
			"stats_conns": ["*localhost"],						
			"resources_conns": ["*conn1", "*conn2"],
	},
}`
	eMap := map[string]interface{}{
		utils.StatSConnsCfg:     []string{utils.MetaLocalHost},
		utils.ResourceSConnsCfg: []string{"*conn1", "*conn2"},
		utils.ApierSConnsCfg:    []string{},
	}
	if cgrCfg, err := NewCGRConfigFromJsonStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cgrCfg.filterSCfg.AsMapInterface(); !reflect.DeepEqual(rcv, eMap) {
		t.Errorf("Expected %+v, received %+v", eMap, rcv)
	}

}

func TestFilterSCfgAsMapInterface2(t *testing.T) {
	cfgJSONStr := `{
      "filters": {}
}`
	eMap := map[string]interface{}{
		utils.StatSConnsCfg:     []string{},
		utils.ResourceSConnsCfg: []string{},
		utils.ApierSConnsCfg:    []string{},
	}
	if cgrCfg, err := NewCGRConfigFromJsonStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cgrCfg.filterSCfg.AsMapInterface(); !reflect.DeepEqual(rcv, eMap) {
		t.Errorf("Expected %+v, received %+v", eMap, rcv)
	}
}
