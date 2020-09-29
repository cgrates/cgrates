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

func TestRouteSCfgloadFromJsonCfg(t *testing.T) {
	cfgJSON := &RouteSJsonCfg{
		Enabled:               utils.BoolPointer(true),
		Indexed_selects:       utils.BoolPointer(true),
		Prefix_indexed_fields: &[]string{"*req.index1", "*req.index2"},
		Suffix_indexed_fields: &[]string{"*req.index1", "*req.index2"},
		Attributes_conns:      &[]string{utils.MetaInternal},
		Resources_conns:       &[]string{utils.MetaInternal},
		Stats_conns:           &[]string{utils.MetaInternal},
		Rals_conns:            &[]string{utils.MetaInternal},
		Default_ratio:         utils.IntPointer(10),
		Nested_fields:         utils.BoolPointer(true),
	}
	expected := &RouteSCfg{
		Enabled:             true,
		IndexedSelects:      true,
		PrefixIndexedFields: &[]string{"*req.index1", "*req.index2"},
		SuffixIndexedFields: &[]string{"*req.index1", "*req.index2"},
		AttributeSConns:     []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes)},
		ResourceSConns:      []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaResources)},
		StatSConns:          []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaStatS)},
		RALsConns:           []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaResponder)},
		DefaultRatio:        10,
		NestedFields:        true,
	}
	if jsonCfg, err := NewDefaultCGRConfig(); err != nil {
		t.Error(err)
	} else if err = jsonCfg.routeSCfg.loadFromJsonCfg(cfgJSON); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, jsonCfg.routeSCfg) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(jsonCfg.routeSCfg))
	}
}

func TestRouteSCfgAsMapInterface(t *testing.T) {
	cfgJSONStr := `{
	"routes": {},
}`
	eMap := map[string]interface{}{
		utils.EnabledCfg:             false,
		utils.IndexedSelectsCfg:      true,
		utils.PrefixIndexedFieldsCfg: []string{},
		utils.SuffixIndexedFieldsCfg: []string{},
		utils.NestedFieldsCfg:        false,
		utils.AttributeSConnsCfg:     []string{},
		utils.ResourceSConnsCfg:      []string{},
		utils.StatSConnsCfg:          []string{},
		utils.RALsConnsCfg:           []string{},
		utils.DefaultRatioCfg:        1,
	}
	if cgrCfg, err := NewCGRConfigFromJsonStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cgrCfg.routeSCfg.AsMapInterface(); !reflect.DeepEqual(eMap, rcv) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}
}

func TestRouteSCfgAsMapInterface1(t *testing.T) {
	cfgJSONStr := `{
		"routes": {
			"enabled": true,
			"indexed_selects":false,
			"prefix_indexed_fields": ["*req.prefix","*req.indexed","*req.fields"],
            "suffix_indexed_fields": ["*req.prefix","*req.indexed"],
			"nested_fields": true,
			"attributes_conns": ["*internal"],
			"resources_conns": ["*internal"],
			"stats_conns": ["*internal"],
			"rals_conns": ["*internal"],
			"default_ratio":2,
		},
	}`
	eMap := map[string]interface{}{
		utils.EnabledCfg:             true,
		utils.IndexedSelectsCfg:      false,
		utils.PrefixIndexedFieldsCfg: []string{"*req.prefix", "*req.indexed", "*req.fields"},
		utils.SuffixIndexedFieldsCfg: []string{"*req.prefix", "*req.indexed"},
		utils.NestedFieldsCfg:        true,
		utils.AttributeSConnsCfg:     []string{"*internal"},
		utils.ResourceSConnsCfg:      []string{"*internal"},
		utils.StatSConnsCfg:          []string{"*internal"},
		utils.RALsConnsCfg:           []string{"*internal"},
		utils.DefaultRatioCfg:        2,
	}
	if cgrCfg, err := NewCGRConfigFromJsonStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cgrCfg.routeSCfg.AsMapInterface(); !reflect.DeepEqual(eMap, rcv) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}
}
