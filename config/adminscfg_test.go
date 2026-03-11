/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/
package config

import (
	"reflect"
	"testing"

	"github.com/cgrates/cgrates/utils"
)

func TestApierCfgloadFromJsonCfg(t *testing.T) {
	jsonCfg := &AdminSJsonCfg{
		Enabled: utils.BoolPointer(false),
		Conns: map[string][]*DynamicConns{
			utils.MetaCaches:     {{ConnIDs: []string{utils.MetaInternal, "*conn1"}}},
			utils.MetaActions:    {{ConnIDs: []string{utils.MetaInternal, "*conn1"}}},
			utils.MetaAttributes: {{ConnIDs: []string{utils.MetaInternal, "*conn1"}}},
			utils.MetaEEs:        {{ConnIDs: []string{utils.MetaInternal, "*conn1"}}},
		},
	}
	expected := &AdminSCfg{
		Enabled: false,
		Conns: map[string][]*DynamicConns{
			utils.MetaCaches:     {{ConnIDs: []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches), "*conn1"}}},
			utils.MetaActions:    {{ConnIDs: []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaActions), "*conn1"}}},
			utils.MetaAttributes: {{ConnIDs: []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes), "*conn1"}}},
			utils.MetaEEs:        {{ConnIDs: []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaEEs), "*conn1"}}},
		},
	}
	jsnCfg := NewDefaultCGRConfig()
	if err := jsnCfg.admS.loadFromJSONCfg(jsonCfg); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, jsnCfg.admS) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(jsnCfg.admS))
	}

	jsonCfg = nil
	if err := jsnCfg.admS.loadFromJSONCfg(jsonCfg); err != nil {
		t.Error(err)
	}
}

func TestApierCfgAsMapInterface1(t *testing.T) {
	cfgJSONStr := `{
	"admins": {
		"conns": {},
	},
}`
	eMap := map[string]any{
		utils.EnabledCfg: false,
		utils.ConnsCfg: map[string][]*DynamicConns{
			utils.MetaCaches: {{ConnIDs: []string{utils.MetaInternal}}},
		},
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
       "conns": {
           "*attributes": [{"ConnIDs": ["*internal:*attributes", "*conn1"]}],
           "*ees":        [{"ConnIDs": ["*internal:*ees", "*conn1"]}],
           "*caches":     [{"ConnIDs": ["*internal:*caches", "*conn1"]}],
           "*actions":    [{"ConnIDs": ["*internal:*actions", "*conn1"]}]
       },
    },
}`
	expectedMap := map[string]any{
		utils.EnabledCfg: true,
		utils.ConnsCfg: map[string][]*DynamicConns{
			utils.MetaAttributes: {{ConnIDs: []string{utils.MetaInternal, "*conn1"}}},
			utils.MetaEEs:        {{ConnIDs: []string{utils.MetaInternal, "*conn1"}}},
			utils.MetaCaches:     {{ConnIDs: []string{utils.MetaInternal, "*conn1"}}},
			utils.MetaActions:    {{ConnIDs: []string{utils.MetaInternal, "*conn1"}}},
		},
	}
	if cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(myJSONStr); err != nil {
		t.Error(err)
	} else if newMap := cgrCfg.admS.AsMapInterface(); !reflect.DeepEqual(expectedMap, newMap) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expectedMap), utils.ToJSON(newMap))
	}
}

func TestApierCfgClone(t *testing.T) {
	sa := &AdminSCfg{
		Enabled: false,
		Conns: map[string][]*DynamicConns{
			utils.MetaCaches:     {{ConnIDs: []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches), "*conn1"}}},
			utils.MetaActions:    {{ConnIDs: []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaActions), "*conn1"}}},
			utils.MetaAttributes: {{ConnIDs: []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes), "*conn1"}}},
			utils.MetaEEs:        {{ConnIDs: []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaEEs), "*conn1"}}},
		},
	}
	rcv := sa.Clone()
	if !reflect.DeepEqual(sa, rcv) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(sa), utils.ToJSON(rcv))
	}
	if rcv.Conns[utils.MetaCaches][0].ConnIDs[1] = ""; sa.Conns[utils.MetaCaches][0].ConnIDs[1] != "*conn1" {
		t.Errorf("Expected clone to not modify the cloned")
	}
	if rcv.Conns[utils.MetaActions][0].ConnIDs[1] = ""; sa.Conns[utils.MetaActions][0].ConnIDs[1] != "*conn1" {
		t.Errorf("Expected clone to not modify the cloned")
	}
	if rcv.Conns[utils.MetaAttributes][0].ConnIDs[1] = ""; sa.Conns[utils.MetaAttributes][0].ConnIDs[1] != "*conn1" {
		t.Errorf("Expected clone to not modify the cloned")
	}
	if rcv.Conns[utils.MetaEEs][0].ConnIDs[1] = ""; sa.Conns[utils.MetaEEs][0].ConnIDs[1] != "*conn1" {
		t.Errorf("Expected clone to not modify the cloned")
	}
}

func TestApierCfgDiffAdminSJsonCfg(t *testing.T) {
	var d *AdminSJsonCfg

	v1 := &AdminSCfg{
		Enabled: false,
		Conns: map[string][]*DynamicConns{
			utils.MetaCaches:     {{ConnIDs: []string{"*localhost"}}},
			utils.MetaActions:    {{ConnIDs: []string{"*localhost"}}},
			utils.MetaAttributes: {{ConnIDs: []string{"*localhost"}}},
			utils.MetaEEs:        {{ConnIDs: []string{"*localhost"}}},
		},
	}

	v2 := &AdminSCfg{
		Enabled: true,
		Conns: map[string][]*DynamicConns{
			utils.MetaCaches:     {{ConnIDs: []string{"*birpc"}}},
			utils.MetaActions:    {{ConnIDs: []string{"*birpc"}}},
			utils.MetaAttributes: {{ConnIDs: []string{"*birpc"}}},
			utils.MetaEEs:        {{ConnIDs: []string{"*birpc"}}},
		},
	}

	expected := &AdminSJsonCfg{
		Enabled: utils.BoolPointer(true),
		Conns: map[string][]*DynamicConns{
			utils.MetaCaches:     {{ConnIDs: []string{"*birpc"}}},
			utils.MetaActions:    {{ConnIDs: []string{"*birpc"}}},
			utils.MetaAttributes: {{ConnIDs: []string{"*birpc"}}},
			utils.MetaEEs:        {{ConnIDs: []string{"*birpc"}}},
		},
	}

	rcv := diffAdminSJsonCfg(d, v1, v2)
	if !reflect.DeepEqual(expected, rcv) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}

	v2_2 := v1
	expected2 := &AdminSJsonCfg{}

	rcv = diffAdminSJsonCfg(d, v1, v2_2)
	if !reflect.DeepEqual(expected2, rcv) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected2), utils.ToJSON(rcv))
	}
}

func TestAdminSCloneSection(t *testing.T) {
	admCfg := &AdminSCfg{
		Enabled: false,
		Conns: map[string][]*DynamicConns{
			utils.MetaCaches:     {{ConnIDs: []string{"*localhost"}}},
			utils.MetaActions:    {{ConnIDs: []string{"*localhost"}}},
			utils.MetaAttributes: {{ConnIDs: []string{"*localhost"}}},
			utils.MetaEEs:        {{ConnIDs: []string{"*localhost"}}},
		},
	}

	exp := &AdminSCfg{
		Enabled: false,
		Conns: map[string][]*DynamicConns{
			utils.MetaCaches:     {{ConnIDs: []string{"*localhost"}}},
			utils.MetaActions:    {{ConnIDs: []string{"*localhost"}}},
			utils.MetaAttributes: {{ConnIDs: []string{"*localhost"}}},
			utils.MetaEEs:        {{ConnIDs: []string{"*localhost"}}},
		},
	}

	rcv := admCfg.CloneSection()
	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(rcv))
	}
}
