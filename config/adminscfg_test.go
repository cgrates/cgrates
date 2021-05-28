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
	jsonCfg := &AdminSJsonCfg{
		Enabled:          utils.BoolPointer(false),
		Caches_conns:     &[]string{utils.MetaInternal, "*conn1"},
		Actions_conns:    &[]string{utils.MetaInternal, "*conn1"},
		Attributes_conns: &[]string{utils.MetaInternal, "*conn1"},
		Ees_conns:        &[]string{utils.MetaInternal, "*conn1"},
	}
	expected := &AdminSCfg{
		Enabled:         false,
		CachesConns:     []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches), "*conn1"},
		ActionSConns:    []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaActions), "*conn1"},
		AttributeSConns: []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes), "*conn1"},
		EEsConns:        []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaEEs), "*conn1"},
	}
	jsnCfg := NewDefaultCGRConfig()
	if err = jsnCfg.admS.loadFromJSONCfg(jsonCfg); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, jsnCfg.admS) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(jsnCfg.admS))
	}
}

func TestApierCfgAsMapInterface1(t *testing.T) {
	cfgJSONStr := `{
	"admins": {
		"caches_conns":[],
	},
}`
	sls := make([]string, 0)
	eMap := map[string]interface{}{
		utils.EnabledCfg:         false,
		utils.CachesConnsCfg:     sls,
		utils.ActionSConnsCfg:    sls,
		utils.AttributeSConnsCfg: sls,
		utils.EEsConnsCfg:        sls,
	}
	if cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if newMap := cgrCfg.admS.AsMapInterface(); !reflect.DeepEqual(newMap, eMap) {
		t.Errorf("Expected %+v, received %+v", eMap, newMap)
	}
}

func TestApierCfgAsMapInterface2(t *testing.T) {
	myJSONStr := `{
    "admins": {
       "enabled": true,
       "attributes_conns": ["*internal:*attributes", "*conn1"],
       "ees_conns": ["*internal:*ees", "*conn1"],
       "caches_conns": ["*internal:*caches", "*conn1"],
       "actions_conns": ["*internal:*actions", "*conn1"],
    },
}`
	expectedMap := map[string]interface{}{
		utils.EnabledCfg:         true,
		utils.CachesConnsCfg:     []string{utils.MetaInternal, "*conn1"},
		utils.ActionSConnsCfg:    []string{utils.MetaInternal, "*conn1"},
		utils.AttributeSConnsCfg: []string{utils.MetaInternal, "*conn1"},
		utils.EEsConnsCfg:        []string{utils.MetaInternal, "*conn1"},
	}
	if cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(myJSONStr); err != nil {
		t.Error(err)
	} else if newMap := cgrCfg.admS.AsMapInterface(); !reflect.DeepEqual(expectedMap, newMap) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expectedMap), utils.ToJSON(newMap))
	}
}

func TestApierCfgClone(t *testing.T) {
	sa := &AdminSCfg{
		Enabled:         false,
		CachesConns:     []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches), "*conn1"},
		ActionSConns:    []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaActions), "*conn1"},
		AttributeSConns: []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes), "*conn1"},
		EEsConns:        []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaEEs), "*conn1"},
	}
	rcv := sa.Clone()
	if !reflect.DeepEqual(sa, rcv) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(sa), utils.ToJSON(rcv))
	}
	if rcv.CachesConns[1] = ""; sa.CachesConns[1] != "*conn1" {
		t.Errorf("Expected clone to not modify the cloned")
	}
	if rcv.ActionSConns[1] = ""; sa.ActionSConns[1] != "*conn1" {
		t.Errorf("Expected clone to not modify the cloned")
	}
	if rcv.AttributeSConns[1] = ""; sa.AttributeSConns[1] != "*conn1" {
		t.Errorf("Expected clone to not modify the cloned")
	}
	if rcv.EEsConns[1] = ""; sa.EEsConns[1] != "*conn1" {
		t.Errorf("Expected clone to not modify the cloned")
	}
}

func TestApierCfgDiffAdminSJsonCfg(t *testing.T) {
	var d *AdminSJsonCfg

	v1 := &AdminSCfg{
		Enabled:         false,
		CachesConns:     []string{"*localhost"},
		ActionSConns:    []string{"*localhost"},
		AttributeSConns: []string{"*localhost"},
		EEsConns:        []string{"*localhost"},
	}

	v2 := &AdminSCfg{
		Enabled:         true,
		CachesConns:     []string{"*birpc"},
		ActionSConns:    []string{"*birpc"},
		AttributeSConns: []string{"*birpc"},
		EEsConns:        []string{"*birpc"},
	}

	expected := &AdminSJsonCfg{
		Enabled:          utils.BoolPointer(true),
		Caches_conns:     &[]string{"*birpc"},
		Actions_conns:    &[]string{"*birpc"},
		Attributes_conns: &[]string{"*birpc"},
		Ees_conns:        &[]string{"*birpc"},
	}

	rcv := diffAdminSJsonCfg(d, v1, v2)
	if !reflect.DeepEqual(expected, rcv) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}

	v2_2 := v1
	expected2 := &AdminSJsonCfg{
		Enabled:          nil,
		Caches_conns:     nil,
		Actions_conns:    nil,
		Attributes_conns: nil,
		Ees_conns:        nil,
	}

	rcv = diffAdminSJsonCfg(d, v1, v2_2)
	if !reflect.DeepEqual(expected2, rcv) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected2), utils.ToJSON(rcv))
	}
}
