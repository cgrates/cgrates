// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package config

import (
	"reflect"
	"testing"

	"github.com/cgrates/cgrates/utils"
)

func TestSchedulerCfgloadFromJsonCfg(t *testing.T) {
	cfgJSONS := &SchedulerJsonCfg{
		Enabled:                 utils.BoolPointer(true),
		Cdrs_conns:              &[]string{utils.MetaInternal, "*conn1"},
		Thresholds_conns:        &[]string{utils.MetaInternal, "*conn1"},
		Stats_conns:             &[]string{utils.MetaInternal, "*conn1"},
		Filters:                 &[]string{"randomFilter"},
		Dynaprepaid_actionplans: &[]string{"randomPlan"},
	}
	expected := &SchedulerCfg{
		Enabled:                true,
		CDRsConns:              []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCDRs), "*conn1"},
		ThreshSConns:           []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds), "*conn1"},
		StatSConns:             []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaStats), "*conn1"},
		Filters:                []string{"randomFilter"},
		DynaprepaidActionPlans: []string{"randomPlan"},
	}
	jsonCfg := NewDefaultCGRConfig()
	if err := jsonCfg.schedulerCfg.loadFromJSONCfg(cfgJSONS); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, jsonCfg.schedulerCfg) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(jsonCfg.schedulerCfg))
	}
}

func TestSchedulerCfgAsMapInterface(t *testing.T) {
	cfgJSONStr := `{
	"schedulers": {},
}`
	eMap := map[string]any{
		utils.EnabledCfg:                false,
		utils.CDRsConnsCfg:              []string{},
		utils.ThreshSConnsCfg:           []string{},
		utils.StatSConnsCfg:             []string{},
		utils.FiltersCfg:                []string{},
		utils.DynaprepaidActionplansCfg: []string{},
	}
	if cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cgrCfg.schedulerCfg.AsMapInterface(); !reflect.DeepEqual(eMap, rcv) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}

}

func TestSchedulerCfgAsMapInterface1(t *testing.T) {
	cfgJSONStr := `{
	"schedulers": {
       "enabled": true,
	   "cdrs_conns": ["*internal", "*conn1"],
	   "thresholds_conns": ["*internal", "*conn1"],
	   "stats_conns": ["*internal", "*conn1"],
       "filters": ["randomFilter"],
		"dynaprepaid_actionplans":["randomPlan"],
    },
}`
	eMap := map[string]any{
		utils.EnabledCfg:                true,
		utils.CDRsConnsCfg:              []string{utils.MetaInternal, "*conn1"},
		utils.ThreshSConnsCfg:           []string{utils.MetaInternal, "*conn1"},
		utils.StatSConnsCfg:             []string{utils.MetaInternal, "*conn1"},
		utils.FiltersCfg:                []string{"randomFilter"},
		utils.DynaprepaidActionplansCfg: []string{"randomPlan"},
	}
	if cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cgrCfg.schedulerCfg.AsMapInterface(); !reflect.DeepEqual(eMap, rcv) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}

}

func TestSchedulerCfgClone(t *testing.T) {
	ban := &SchedulerCfg{
		Enabled:                true,
		CDRsConns:              []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCDRs), "*conn1"},
		ThreshSConns:           []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds), "*conn1"},
		StatSConns:             []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaStats), "*conn1"},
		Filters:                []string{"randomFilter"},
		DynaprepaidActionPlans: []string{"plan"},
	}
	rcv := ban.Clone()
	if !reflect.DeepEqual(ban, rcv) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(ban), utils.ToJSON(rcv))
	}
	if rcv.CDRsConns[1] = ""; ban.CDRsConns[1] != "*conn1" {
		t.Errorf("Expected clone to not modify the cloned")
	}
	if rcv.ThreshSConns[1] = ""; ban.ThreshSConns[1] != "*conn1" {
		t.Errorf("Expected clone to not modify the cloned")
	}
	if rcv.StatSConns[1] = ""; ban.StatSConns[1] != "*conn1" {
		t.Errorf("Expected clone to not modify the cloned")
	}
	if rcv.Filters[0] = ""; ban.Filters[0] != "randomFilter" {
		t.Errorf("Expected clone to not modify the cloned")
	}

	if rcv.DynaprepaidActionPlans[0] = ""; ban.DynaprepaidActionPlans[0] != "plan" {
		t.Errorf("Expected clone to not modify the cloned")
	}

	ban = nil
	rcv = ban.Clone()
	if !reflect.DeepEqual(ban, rcv) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(ban), utils.ToJSON(rcv))
	}
}
