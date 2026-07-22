// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package config

import (
	"reflect"
	"testing"

	"github.com/cgrates/cgrates/utils"
)

func TestApierCfgloadFromJsonCfg(t *testing.T) {
	jsonCfg := &ApierJsonCfg{
		Enabled:          utils.BoolPointer(false),
		Caches_conns:     &[]string{utils.MetaInternal, "*conn1"},
		Scheduler_conns:  &[]string{utils.MetaInternal, "*conn1"},
		Attributes_conns: &[]string{utils.MetaInternal, "*conn1"},
		Ees_conns:        &[]string{utils.MetaInternal, "*conn1"},
	}
	expected := &ApierCfg{
		Enabled:         false,
		CachesConns:     []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches), "*conn1"},
		SchedulerConns:  []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaScheduler), "*conn1"},
		AttributeSConns: []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes), "*conn1"},
		EEsConns:        []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaEEs), "*conn1"},
	}
	jsnCfg := NewDefaultCGRConfig()
	if err := jsnCfg.apier.loadFromJSONCfg(jsonCfg); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, jsnCfg.apier) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(jsnCfg.apier))
	}
}

func TestApierCfgAsMapInterface1(t *testing.T) {
	cfgJSONStr := `{
	"apiers": {
		"caches_conns":[],
	},
}`
	sls := make([]string, 0)
	eMap := map[string]any{
		utils.EnabledCfg:         false,
		utils.CachesConnsCfg:     sls,
		utils.SchedulerConnsCfg:  sls,
		utils.AttributeSConnsCfg: sls,
		utils.EEsConnsCfg:        sls,
	}
	if cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if newMap := cgrCfg.apier.AsMapInterface(); !reflect.DeepEqual(newMap, eMap) {
		t.Errorf("Expected %+v, received %+v", eMap, newMap)
	}
}

func TestApierCfgAsMapInterface2(t *testing.T) {
	myJSONStr := `{
    "apiers": {
       "enabled": true,
       "attributes_conns": ["*internal:*attributes", "*conn1"],
       "ees_conns": ["*internal:*ees", "*conn1"],
       "caches_conns": ["*internal:*caches", "*conn1"],
       "scheduler_conns": ["*internal:*scheduler", "*conn1"],
    },
}`
	expectedMap := map[string]any{
		utils.EnabledCfg:         true,
		utils.CachesConnsCfg:     []string{utils.MetaInternal, "*conn1"},
		utils.SchedulerConnsCfg:  []string{utils.MetaInternal, "*conn1"},
		utils.AttributeSConnsCfg: []string{utils.MetaInternal, "*conn1"},
		utils.EEsConnsCfg:        []string{utils.MetaInternal, "*conn1"},
	}
	if cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(myJSONStr); err != nil {
		t.Error(err)
	} else if newMap := cgrCfg.apier.AsMapInterface(); !reflect.DeepEqual(expectedMap, newMap) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expectedMap), utils.ToJSON(newMap))
	}
}

func TestApierCfgClone(t *testing.T) {
	sa := &ApierCfg{
		Enabled:         false,
		CachesConns:     []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches), "*conn1"},
		SchedulerConns:  []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaScheduler), "*conn1"},
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
	if rcv.SchedulerConns[1] = ""; sa.SchedulerConns[1] != "*conn1" {
		t.Errorf("Expected clone to not modify the cloned")
	}
	if rcv.AttributeSConns[1] = ""; sa.AttributeSConns[1] != "*conn1" {
		t.Errorf("Expected clone to not modify the cloned")
	}
	if rcv.EEsConns[1] = ""; sa.EEsConns[1] != "*conn1" {
		t.Errorf("Expected clone to not modify the cloned")
	}

	sa = nil
	rcv = sa.Clone()
	if !reflect.DeepEqual(sa, rcv) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(sa), utils.ToJSON(rcv))
	}
}
