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

func TestActionSCfgLoadFromJSONCfg(t *testing.T) {
	jsonCfg := &ActionSJsonCfg{
		Enabled:                   utils.BoolPointer(true),
		Ees_conns:                 &[]string{utils.MetaInternal},
		Cdrs_conns:                &[]string{utils.MetaInternal},
		Thresholds_conns:          &[]string{utils.MetaInternal},
		Stats_conns:               &[]string{utils.MetaInternal},
		Accounts_conns:            &[]string{utils.MetaInternal},
		Indexed_selects:           utils.BoolPointer(false),
		Tenants:                   &[]string{"itsyscom.com"},
		String_indexed_fields:     &[]string{"*req.index1"},
		Prefix_indexed_fields:     &[]string{"*req.index1", "*req.index2"},
		Suffix_indexed_fields:     &[]string{"*req.index1"},
		Nested_fields:             utils.BoolPointer(true),
		Dynaprepaid_actionprofile: &[]string{},
	}
	expected := &ActionSCfg{
		Enabled:                  true,
		EEsConns:                 []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaEEs)},
		CDRsConns:                []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCDRs)},
		ThresholdSConns:          []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds)},
		StatSConns:               []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaStats)},
		AccountSConns:            []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAccounts)},
		IndexedSelects:           false,
		Tenants:                  &[]string{"itsyscom.com"},
		StringIndexedFields:      &[]string{"*req.index1"},
		PrefixIndexedFields:      &[]string{"*req.index1", "*req.index2"},
		SuffixIndexedFields:      &[]string{"*req.index1"},
		NestedFields:             true,
		DynaprepaidActionProfile: []string{},
	}
	jsnCfg := NewDefaultCGRConfig()
	if err = jsnCfg.actionSCfg.loadFromJSONCfg(jsonCfg); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, jsnCfg.actionSCfg) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", utils.ToJSON(expected), utils.ToJSON(jsnCfg.actionSCfg))
	}
}

func TestActionSCfgAsMapInterface(t *testing.T) {
	cfgJSONStr := `{
"actions": {								
	"enabled": true,
    "cdrs_conns": ["*internal"],						
	"ees_conns": ["*internal"],						
	"thresholds_conns": ["*internal"],					
	"stats_conns": ["*internal"],						
	"accounts_conns": ["*internal"],						
	"tenants": ["itsyscom.com"],
	"indexed_selects": false,
	"string_indexed_fields": ["*req.index1"],			
	"prefix_indexed_fields": ["*req.index1","*req.index2"],		
    "suffix_indexed_fields": ["*req.index1"],
	"nested_fields": true,	
    "DynaprepaidActionProfile": [],
	},		
}`

	eMap := map[string]interface{}{
		utils.EnabledCfg:                true,
		utils.EEsConnsCfg:               []string{utils.MetaInternal},
		utils.ThresholdSConnsCfg:        []string{utils.MetaInternal},
		utils.StatSConnsCfg:             []string{utils.MetaInternal},
		utils.CDRsConnsCfg:              []string{utils.MetaInternal},
		utils.AccountSConnsCfg:          []string{utils.MetaInternal},
		utils.Tenants:                   []string{"itsyscom.com"},
		utils.IndexedSelectsCfg:         false,
		utils.StringIndexedFieldsCfg:    []string{"*req.index1"},
		utils.PrefixIndexedFieldsCfg:    []string{"*req.index1", "*req.index2"},
		utils.SuffixIndexedFieldsCfg:    []string{"*req.index1"},
		utils.NestedFieldsCfg:           true,
		utils.DynaprepaidActionplansCfg: []string{},
	}
	if cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cgrCfg.actionSCfg.AsMapInterface(); !reflect.DeepEqual(eMap, rcv) {
		t.Errorf("Expected: %+v\n Received: %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}
}

func TestActionSCfgClone(t *testing.T) {
	ban := &ActionSCfg{
		Enabled:             true,
		EEsConns:            []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaEEs)},
		CDRsConns:           []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCDRs)},
		ThresholdSConns:     []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds)},
		StatSConns:          []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaStats)},
		AccountSConns:       []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAccounts)},
		Tenants:             &[]string{"itsyscom.com"},
		IndexedSelects:      false,
		StringIndexedFields: &[]string{"*req.index1"},
		PrefixIndexedFields: &[]string{"*req.index1", "*req.index2"},
		SuffixIndexedFields: &[]string{"*req.index1"},
		NestedFields:        true,
	}
	rcv := ban.Clone()
	if !reflect.DeepEqual(ban, rcv) {
		t.Errorf("\nExpected: %+v\nReceived: %+v", utils.ToJSON(ban), utils.ToJSON(rcv))
	}
	if (*rcv.Tenants)[0] = utils.EmptyString; (*ban.Tenants)[0] != "itsyscom.com" {
		t.Errorf("Expected clone to not modify the cloned")
	}
	if (*rcv.StringIndexedFields)[0] = utils.EmptyString; (*ban.StringIndexedFields)[0] != "*req.index1" {
		t.Errorf("Expected clone to not modify the cloned")
	}
	if (*rcv.PrefixIndexedFields)[0] = utils.EmptyString; (*ban.PrefixIndexedFields)[0] != "*req.index1" {
		t.Errorf("Expected clone to not modify the cloned")
	}
	if (*rcv.SuffixIndexedFields)[0] = utils.EmptyString; (*ban.SuffixIndexedFields)[0] != "*req.index1" {
		t.Errorf("Expected clone to not modify the cloned")
	}
	if rcv.EEsConns[0] = utils.EmptyString; ban.EEsConns[0] != utils.ConcatenatedKey(utils.MetaInternal, utils.MetaEEs) {
		t.Errorf("Expected clone to not modify the cloned")
	}
	if rcv.CDRsConns[0] = utils.EmptyString; ban.CDRsConns[0] != utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCDRs) {
		t.Errorf("Expected clone to not modify the cloned")
	}
	if rcv.ThresholdSConns[0] = utils.EmptyString; ban.ThresholdSConns[0] != utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds) {
		t.Errorf("Expected clone to not modify the cloned")
	}
	if rcv.StatSConns[0] = utils.EmptyString; ban.StatSConns[0] != utils.ConcatenatedKey(utils.MetaInternal, utils.MetaStats) {
		t.Errorf("Expected clone to not modify the cloned")
	}
	if rcv.AccountSConns[0] = utils.EmptyString; ban.AccountSConns[0] != utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAccounts) {
		t.Errorf("Expected clone to not modify the cloned")
	}
}

func TestDiffActionSJsonCfg(t *testing.T) {
	var d *ActionSJsonCfg

	v1 := &ActionSCfg{
		Enabled:             false,
		CDRsConns:           []string{},
		EEsConns:            []string{},
		ThresholdSConns:     []string{},
		StatSConns:          []string{},
		AccountSConns:       []string{},
		Tenants:             &[]string{},
		IndexedSelects:      false,
		StringIndexedFields: &[]string{},
		PrefixIndexedFields: &[]string{},
		SuffixIndexedFields: &[]string{},
		NestedFields:        true,
	}

	v2 := &ActionSCfg{
		Enabled:             true,
		CDRsConns:           []string{"*localhost"},
		EEsConns:            []string{"*localhost"},
		ThresholdSConns:     []string{"*localhost"},
		StatSConns:          []string{"*localhost"},
		AccountSConns:       []string{"*localhost"},
		Tenants:             &[]string{"cgrates.org"},
		IndexedSelects:      true,
		StringIndexedFields: &[]string{"*req.Index1"},
		PrefixIndexedFields: nil,
		SuffixIndexedFields: nil,
		NestedFields:        false,
	}

	expected := &ActionSJsonCfg{
		Enabled:               utils.BoolPointer(true),
		Cdrs_conns:            &[]string{"*localhost"},
		Ees_conns:             &[]string{"*localhost"},
		Thresholds_conns:      &[]string{"*localhost"},
		Stats_conns:           &[]string{"*localhost"},
		Accounts_conns:        &[]string{"*localhost"},
		Tenants:               &[]string{"cgrates.org"},
		Indexed_selects:       utils.BoolPointer(true),
		String_indexed_fields: &[]string{"*req.Index1"},
		Prefix_indexed_fields: nil,
		Suffix_indexed_fields: nil,
		Nested_fields:         utils.BoolPointer(false),
	}

	rcv := diffActionSJsonCfg(d, v1, v2)
	if !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}

	//The output "d" should be nil when there isn't any difference between v1 and v2_2
	v2_2 := v1
	expected2 := &ActionSJsonCfg{
		Enabled:               nil,
		Cdrs_conns:            nil,
		Ees_conns:             nil,
		Thresholds_conns:      nil,
		Stats_conns:           nil,
		Accounts_conns:        nil,
		Tenants:               nil,
		Indexed_selects:       nil,
		String_indexed_fields: nil,
		Prefix_indexed_fields: nil,
		Suffix_indexed_fields: nil,
		Nested_fields:         nil,
	}
	rcv = diffActionSJsonCfg(d, v1, v2_2)
	if !reflect.DeepEqual(rcv, expected2) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected2), utils.ToJSON(rcv))
	}
}
