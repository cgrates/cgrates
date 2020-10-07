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
		String_indexed_fields: &[]string{"*req.index1"},
		Prefix_indexed_fields: &[]string{"*req.index1", "*req.index2"},
		Suffix_indexed_fields: &[]string{"*req.index1", "*req.index2"},
		Attributes_conns:      &[]string{utils.MetaInternal, "*conn1"},
		Resources_conns:       &[]string{utils.MetaInternal, "*conn1"},
		Stats_conns:           &[]string{utils.MetaInternal, "*conn1"},
		Rals_conns:            &[]string{utils.MetaInternal, "*conn1"},
		Default_ratio:         utils.IntPointer(10),
		Nested_fields:         utils.BoolPointer(true),
	}
	expected := &RouteSCfg{
		Enabled:             true,
		IndexedSelects:      true,
		StringIndexedFields: &[]string{"*req.index1"},
		PrefixIndexedFields: &[]string{"*req.index1", "*req.index2"},
		SuffixIndexedFields: &[]string{"*req.index1", "*req.index2"},
		AttributeSConns:     []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes), "*conn1"},
		ResourceSConns:      []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaResources), "*conn1"},
		StatSConns:          []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaStatS), "*conn1"},
		RALsConns:           []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaResponder), "*conn1"},
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
            "string_indexed_fields": ["*req.string"],
			"prefix_indexed_fields": ["*req.prefix","*req.indexed","*req.fields"],
            "suffix_indexed_fields": ["*req.prefix","*req.indexed"],
			"nested_fields": true,
			"attributes_conns": ["*internal:*attributes", "*conn1"],
			"resources_conns": ["*internal:*resources", "*conn1"],
			"stats_conns": ["*internal:*stats", "*conn1"],
			"rals_conns": ["*internal:*responder", "*conn1"],
			"default_ratio":2,
		},
	}`
	eMap := map[string]interface{}{
		utils.EnabledCfg:             true,
		utils.IndexedSelectsCfg:      false,
		utils.StringIndexedFieldsCfg: []string{"*req.string"},
		utils.PrefixIndexedFieldsCfg: []string{"*req.prefix", "*req.indexed", "*req.fields"},
		utils.SuffixIndexedFieldsCfg: []string{"*req.prefix", "*req.indexed"},
		utils.NestedFieldsCfg:        true,
		utils.AttributeSConnsCfg:     []string{utils.MetaInternal, "*conn1"},
		utils.ResourceSConnsCfg:      []string{utils.MetaInternal, "*conn1"},
		utils.StatSConnsCfg:          []string{utils.MetaInternal, "*conn1"},
		utils.RALsConnsCfg:           []string{utils.MetaInternal, "*conn1"},
		utils.DefaultRatioCfg:        2,
	}
	if cgrCfg, err := NewCGRConfigFromJsonStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cgrCfg.routeSCfg.AsMapInterface(); !reflect.DeepEqual(eMap, rcv) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}
}
