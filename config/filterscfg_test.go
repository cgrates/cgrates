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
		Stats_conns:     &[]string{utils.MetaInternal, "*conn1"},
		Resources_conns: &[]string{utils.MetaInternal, "*conn1"},
		Admins_conns:    &[]string{utils.MetaInternal, "*conn1"},
	}
	expected := &FilterSCfg{
		StatSConns:     []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaStats), "*conn1"},
		ResourceSConns: []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaResources), "*conn1"},
		AdminSConns:    []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAdminS), "*conn1"},
	}
	jsnCfg := NewDefaultCGRConfig()
	if err = jsnCfg.filterSCfg.loadFromJSONCfg(cfgJSONS); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, jsnCfg.filterSCfg) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(jsnCfg.filterSCfg))
	}
}

func TestFilterSCfgAsMapInterface(t *testing.T) {
	cfgJSONStr := `{
		"filters": {								
			"stats_conns": ["*internal:*stats", "*conn1"],						
			"resources_conns": ["*internal:*resources", "*conn1"],
            "admins_conns": ["*internal:*apier", "*conn1"],
	},
}`
	eMap := map[string]interface{}{
		utils.StatSConnsCfg:     []string{utils.MetaInternal, "*conn1"},
		utils.ResourceSConnsCfg: []string{utils.MetaInternal, "*conn1"},
		utils.AdminSConnsCfg:    []string{utils.MetaInternal, "*conn1"},
	}
	if cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err != nil {
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
		utils.AdminSConnsCfg:    []string{},
	}
	if cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cgrCfg.filterSCfg.AsMapInterface(); !reflect.DeepEqual(rcv, eMap) {
		t.Errorf("Expected %+v, received %+v", eMap, rcv)
	}
}

func TestFilterSCfgClone(t *testing.T) {
	ban := &FilterSCfg{
		StatSConns:     []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaStats), "*conn1"},
		ResourceSConns: []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaResources), "*conn1"},
		AdminSConns:    []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAdminS), "*conn1"},
	}
	rcv := ban.Clone()
	if !reflect.DeepEqual(ban, rcv) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(ban), utils.ToJSON(rcv))
	}
	if rcv.StatSConns[1] = ""; ban.StatSConns[1] != "*conn1" {
		t.Errorf("Expected clone to not modify the cloned")
	}
	if rcv.ResourceSConns[1] = ""; ban.ResourceSConns[1] != "*conn1" {
		t.Errorf("Expected clone to not modify the cloned")
	}
	if rcv.AdminSConns[1] = ""; ban.AdminSConns[1] != "*conn1" {
		t.Errorf("Expected clone to not modify the cloned")
	}
}

func TestDiffFilterSJsonCfg(t *testing.T) {
	var d *FilterSJsonCfg

	v1 := &FilterSCfg{
		StatSConns:     []string{},
		ResourceSConns: []string{},
		AdminSConns:    []string{},
	}

	v2 := &FilterSCfg{
		StatSConns:     []string{"*localhost"},
		ResourceSConns: []string{"*localhost"},
		AdminSConns:    []string{"*localhost"},
	}

	expected := &FilterSJsonCfg{
		Stats_conns:     &[]string{"*localhost"},
		Resources_conns: &[]string{"*localhost"},
		Admins_conns:    &[]string{"*localhost"},
	}

	rcv := diffFilterSJsonCfg(d, v1, v2)
	if !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}

	v1 = v2
	expected2 := &FilterSJsonCfg{}

	rcv = diffFilterSJsonCfg(d, v1, v2)
	if !reflect.DeepEqual(rcv, expected2) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected2), utils.ToJSON(rcv))
	}
}
