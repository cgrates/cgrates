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

func TestAccountSCfgLoadFromJSONCfg(t *testing.T) {
	jsonCfg := &AccountSJsonCfg{
		Enabled:               utils.BoolPointer(true),
		Attributes_conns:      &[]string{"*req.index1"},
		Rates_conns:           &[]string{"*req.index1"},
		Thresholds_conns:      &[]string{"*req.index1"},
		Indexed_selects:       utils.BoolPointer(false),
		String_indexed_fields: &[]string{"*req.index1"},
		Prefix_indexed_fields: &[]string{"*req.index1"},
		Suffix_indexed_fields: &[]string{"*req.index1"},
		Nested_fields:         utils.BoolPointer(true),
	}
	expected := &AccountSCfg{
		Enabled:             true,
		AttributeSConns:     []string{"*req.index1"},
		RateSConns:          []string{"*req.index1"},
		ThresholdSConns:     []string{"*req.index1"},
		IndexedSelects:      false,
		StringIndexedFields: &[]string{"*req.index1"},
		PrefixIndexedFields: &[]string{"*req.index1"},
		SuffixIndexedFields: &[]string{"*req.index1"},
		NestedFields:        true,
	}
	jsnCfg := NewDefaultCGRConfig()
	if err = jsnCfg.accountSCfg.loadFromJSONCfg(jsonCfg); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, jsnCfg.accountSCfg) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", utils.ToJSON(expected), utils.ToJSON(jsnCfg.accountSCfg))
	}
}

func TestAccountSCfgAsMapInterface(t *testing.T) {
	cfgJSONStr := `{
"accounts": {								
	"enabled": true,						
	"indexed_selects": false,			
	"attributes_conns": ["*req.index1"],
	"rates_conns": ["*req.index1"],
	"thresholds_conns": ["*req.index1"],					
	"string_indexed_fields": ["*req.index1"],			
	"prefix_indexed_fields": ["*req.index1"],			
	"suffix_indexed_fields": ["*req.index1"],			
	"nested_fields": true,					
},	
}`

	eMap := map[string]interface{}{
		utils.EnabledCfg:             true,
		utils.IndexedSelectsCfg:      false,
		utils.AttributeSConnsCfg:     []string{"*req.index1"},
		utils.RateSConnsCfg:          []string{"*req.index1"},
		utils.ThresholdSConnsCfg:     []string{"*req.index1"},
		utils.StringIndexedFieldsCfg: []string{"*req.index1"},
		utils.PrefixIndexedFieldsCfg: []string{"*req.index1"},
		utils.SuffixIndexedFieldsCfg: []string{"*req.index1"},
		utils.NestedFieldsCfg:        true,
	}
	if cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cgrCfg.accountSCfg.AsMapInterface(); !reflect.DeepEqual(eMap, rcv) {
		t.Errorf("Expected: %+v\n Received: %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}
}

func TestAccountSCfgClone(t *testing.T) {
	ban := &AccountSCfg{
		Enabled:             true,
		IndexedSelects:      false,
		AttributeSConns:     []string{"*req.index1"},
		RateSConns:          []string{"*req.index1"},
		ThresholdSConns:     []string{"*req.index1"},
		StringIndexedFields: &[]string{"*req.index1"},
		PrefixIndexedFields: &[]string{"*req.index1", "*req.index2"},
		SuffixIndexedFields: &[]string{"*req.index1"},
		NestedFields:        true,
	}
	rcv := ban.Clone()
	if !reflect.DeepEqual(ban, rcv) {
		t.Errorf("\nExpected: %+v\nReceived: %+v", utils.ToJSON(ban), utils.ToJSON(rcv))
	}
	if (rcv.AttributeSConns)[0] = ""; (ban.AttributeSConns)[0] != "*req.index1" {
		t.Errorf("Expected clone to not modify the cloned")
	}
	if (rcv.RateSConns)[0] = ""; (ban.RateSConns)[0] != "*req.index1" {
		t.Errorf("Expected clone to not modify the cloned")
	}
	if (rcv.ThresholdSConns)[0] = ""; (ban.ThresholdSConns)[0] != "*req.index1" {
		t.Errorf("Expected clone to not modify the cloned")
	}
	if (*rcv.StringIndexedFields)[0] = ""; (*ban.StringIndexedFields)[0] != "*req.index1" {
		t.Errorf("Expected clone to not modify the cloned")
	}
	if (*rcv.PrefixIndexedFields)[0] = ""; (*ban.PrefixIndexedFields)[0] != "*req.index1" {
		t.Errorf("Expected clone to not modify the cloned")
	}
	if (*rcv.SuffixIndexedFields)[0] = ""; (*ban.SuffixIndexedFields)[0] != "*req.index1" {
		t.Errorf("Expected clone to not modify the cloned")
	}
}
