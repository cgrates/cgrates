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
	"time"

	"github.com/cgrates/cgrates/utils"
)

func TestCoreSloadFromJsonCfg(t *testing.T) {
	var alS, expected CoreSCfg
	if err := alS.loadFromJSONCfg(nil); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(alS, expected) {
		t.Errorf("Expected: %+v ,received: %+v", expected, alS)
	}
	if err := alS.loadFromJSONCfg(new(CoreSJsonCfg)); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(alS, expected) {
		t.Errorf("Expected: %+v ,received: %+v", expected, alS)
	}
	cfgJSONStr := `{
		"cores": {
			"caps": 10,							// maximum concurrent request allowed ( 0 to disabled )
			"caps_strategy": "*busy",			// strategy in case in case of concurrent requests reached	
			"caps_stats_interval": "0"			// the interval we sample for caps stats ( 0 to disabled )
		},
}`
	expected = CoreSCfg{
		Caps:              10,
		CapsStrategy:      utils.MetaBusy,
		CapsStatsInterval: 0,
	}
	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnalS, err := jsnCfg.CoreSJSON(); err != nil {
		t.Error(err)
	} else if err = alS.loadFromJSONCfg(jsnalS); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, alS) {
		t.Errorf("Expected: %+v , received: %+v", expected, alS)
	}

	expErr := "time: unknown unit \"ss\" in duration \"1ss\""
	coresJSONCfg := &CoreSJsonCfg{
		Caps_stats_interval: utils.StringPointer("1ss"),
	}
	if err = alS.loadFromJSONCfg(coresJSONCfg); err == nil || err.Error() != expErr {
		t.Errorf("Expected error: %s,received: %v", expErr, err)
	}
	coresJSONCfg = &CoreSJsonCfg{
		Shutdown_timeout: utils.StringPointer("1ss"),
	}
	if err = alS.loadFromJSONCfg(coresJSONCfg); err == nil || err.Error() != expErr {
		t.Errorf("Expected error: %s,received: %v", expErr, err)
	}
}

func TestCoreSAsMapInterface(t *testing.T) {
	var alS CoreSCfg
	cfgJSONStr := `{
		"cores": {
			"caps": 0,							// maximum concurrent request allowed ( 0 to disabled )
			"caps_strategy": "*busy",			// strategy in case in case of concurrent requests reached	
			"caps_stats_interval": "0",			// the interval we sample for caps stats ( 0 to disabled )
			"shutdown_timeout": "0"				// the interval we sample for caps stats ( 0 to disabled )
		},
}`
	eMap := map[string]interface{}{
		utils.CapsCfg:              0,
		utils.CapsStrategyCfg:      utils.MetaBusy,
		utils.CapsStatsIntervalCfg: "0",
		utils.ShutdownTimeoutCfg:   "0",
	}
	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnalS, err := jsnCfg.CoreSJSON(); err != nil {
		t.Error(err)
	} else if err = alS.loadFromJSONCfg(jsnalS); err != nil {
		t.Error(err)
	} else if rcv := alS.AsMapInterface(); !reflect.DeepEqual(eMap, rcv) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}
	eMap[utils.CapsStatsIntervalCfg] = "1s"
	eMap[utils.ShutdownTimeoutCfg] = "1s"
	alS = CoreSCfg{
		Caps:              0,
		CapsStatsInterval: time.Second,
		ShutdownTimeout:   time.Second,
		CapsStrategy:      utils.MetaBusy,
	}
	if rcv := alS.AsMapInterface(); !reflect.DeepEqual(eMap, rcv) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}
}

func TestCoreSCfgClone(t *testing.T) {
	cS := &CoreSCfg{
		Caps:              0,
		CapsStatsInterval: time.Second,
		ShutdownTimeout:   time.Second,
		CapsStrategy:      utils.MetaBusy,
	}
	rcv := cS.Clone()
	if !reflect.DeepEqual(cS, rcv) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(cS), utils.ToJSON(rcv))
	}
	if rcv.Caps = 1; cS.Caps != 0 {
		t.Errorf("Expected clone to not modify the cloned")
	}
}

func TestDiffCoreSJsonCfg(t *testing.T) {
	var d *CoreSJsonCfg

	v1 := &CoreSCfg{
		Caps:              2,
		CapsStrategy:      utils.MetaTopUpReset,
		CapsStatsInterval: 3 * time.Second,
		ShutdownTimeout:   5 * time.Minute,
	}

	v2 := &CoreSCfg{
		Caps:              3,
		CapsStrategy:      utils.MetaMaxCostDisconnect,
		CapsStatsInterval: 1 * time.Second,
		ShutdownTimeout:   2 * time.Minute,
	}

	expected := &CoreSJsonCfg{
		Caps:                utils.IntPointer(3),
		Caps_strategy:       utils.StringPointer(utils.MetaMaxCostDisconnect),
		Caps_stats_interval: utils.StringPointer("1s"),
		Shutdown_timeout:    utils.StringPointer("2m0s"),
	}

	rcv := diffCoreSJsonCfg(d, v1, v2)
	if !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}

	v2_2 := v1
	expected2 := &CoreSJsonCfg{}

	rcv = diffCoreSJsonCfg(d, v1, v2_2)
	if !reflect.DeepEqual(rcv, expected2) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected2), utils.ToJSON(rcv))
	}
}
