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

func TestSchedulerCfgloadFromJsonCfg(t *testing.T) {
	cfgJSONS := &SchedulerJsonCfg{
		Enabled:    utils.BoolPointer(true),
		Cdrs_conns: &[]string{utils.MetaInternal},
		Filters:    &[]string{},
	}
	expected := &SchedulerCfg{
		Enabled:   true,
		CDRsConns: []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCDRs)},
		Filters:   []string{},
	}
	if jsonCfg, err := NewDefaultCGRConfig(); err != nil {
		t.Error(err)
	} else if err = jsonCfg.schedulerCfg.loadFromJsonCfg(cfgJSONS); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, jsonCfg.schedulerCfg) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(jsonCfg.schedulerCfg))
	}
}

func TestSchedulerCfgAsMapInterface(t *testing.T) {
	cfgJSONStr := `{
	"schedulers": {},
}`
	eMap := map[string]interface{}{
		utils.EnabledCfg:   false,
		utils.CDRsConnsCfg: []string{},
		utils.FiltersCfg:   []string{},
	}
	if cgrCfg, err := NewCGRConfigFromJsonStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cgrCfg.schedulerCfg.AsMapInterface(); !reflect.DeepEqual(eMap, rcv) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}

}

func TestSchedulerCfgAsMapInterface1(t *testing.T) {
	cfgJSONStr := `{
	"schedulers": {
       "enabled": true,
       "cdrs_conns": ["*internal"],
    },
}`
	eMap := map[string]interface{}{
		utils.EnabledCfg:   true,
		utils.CDRsConnsCfg: []string{"*internal"},
		utils.FiltersCfg:   []string{},
	}
	if cgrCfg, err := NewCGRConfigFromJsonStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cgrCfg.schedulerCfg.AsMapInterface(); !reflect.DeepEqual(eMap, rcv) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}

}
