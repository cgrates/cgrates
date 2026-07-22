// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

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
	} else if jsnalS, err := jsnCfg.CoreSCfgJson(); err != nil {
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
	if err := alS.loadFromJSONCfg(coresJSONCfg); err == nil || err.Error() != expErr {
		t.Errorf("Expected error: %s,received: %v", expErr, err)
	}
	coresJSONCfg = &CoreSJsonCfg{
		Shutdown_timeout: utils.StringPointer("1ss"),
	}
	if err := alS.loadFromJSONCfg(coresJSONCfg); err == nil || err.Error() != expErr {
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
	eMap := map[string]any{
		utils.CapsCfg:              0,
		utils.CapsStrategyCfg:      utils.MetaBusy,
		utils.CapsStatsIntervalCfg: "0",
		utils.ShutdownTimeoutCfg:   "0",
	}
	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnalS, err := jsnCfg.CoreSCfgJson(); err != nil {
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

	cS = nil
	rcv = cS.Clone()
	if !reflect.DeepEqual(cS, rcv) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(cS), utils.ToJSON(rcv))
	}
}
