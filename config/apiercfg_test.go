// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package config

import (
	"reflect"
	"testing"

	"github.com/cgrates/cgrates/utils"
)

func TestApierCfgloadFromJsonCfg(t *testing.T) {
	var aCfg, expected ApierCfg
	if err := aCfg.loadFromJsonCfg(nil); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(aCfg, expected) {
		t.Errorf("Expected: %+v ,received: %+v", expected, aCfg)
	}
	if err := aCfg.loadFromJsonCfg(new(ApierJsonCfg)); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(aCfg, expected) {
		t.Errorf("Expected: %+v ,received: %+v", expected, aCfg)
	}
	cfgJSONStr := `{
	"apiers": {
		"enabled": false,
		"caches_conns":["*internal"],
		"scheduler_conns": [],
		"attributes_conns": [],
	},
}`
	expected = ApierCfg{
		Enabled:         false,
		CachesConns:     []string{"*internal:*caches"},
		SchedulerConns:  []string{},
		AttributeSConns: []string{},
	}
	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnaCfg, err := jsnCfg.ApierCfgJson(); err != nil {
		t.Error(err)
	} else if err = aCfg.loadFromJsonCfg(jsnaCfg); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, aCfg) {
		t.Errorf("Expected: %+v , received: %+v", expected, aCfg)
	}
}

func TestApierCfgAsMapInterface(t *testing.T) {
	var aCfg ApierCfg
	cfgJSONStr := `{
	"apiers": {
		"enabled": false,
		"caches_conns":[],
		"scheduler_conns": [],
		"attributes_conns": [],
	},
}`
	eMap := map[string]any{
		"enabled":          false,
		"caches_conns":     []string{},
		"scheduler_conns":  []string{},
		"attributes_conns": []string{},
	}
	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnaCfg, err := jsnCfg.ApierCfgJson(); err != nil {
		t.Error(err)
	} else if err = aCfg.loadFromJsonCfg(jsnaCfg); err != nil {
		t.Error(err)
	} else if rcv := aCfg.AsMapInterface(); !reflect.DeepEqual(eMap, rcv) {
		t.Errorf("\nExpected: %+v\nReceived: %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}

	cfgJSONStr = `{
		"apiers": {
			"enabled": false,
			"caches_conns":["*internal"],
			"scheduler_conns": ["*internal"],
			"attributes_conns": ["*internal"],
		},
	}`
	eMap = map[string]any{
		"enabled":          false,
		"caches_conns":     []string{"*internal"},
		"scheduler_conns":  []string{"*internal"},
		"attributes_conns": []string{"*internal"},
	}
	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnaCfg, err := jsnCfg.ApierCfgJson(); err != nil {
		t.Error(err)
	} else if err = aCfg.loadFromJsonCfg(jsnaCfg); err != nil {
		t.Error(err)
	} else if rcv := aCfg.AsMapInterface(); !reflect.DeepEqual(eMap, rcv) {
		t.Errorf("\nExpected: %+v\nReceived: %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}
}

func TestApierCfgloadFromJsonCfg2(t *testing.T) {
	bl := true
	slc := []string{"val1", "val2"}

	a := ApierCfg{}

	js := ApierJsonCfg{
		Enabled:          &bl,
		Caches_conns:     &slc,
		Scheduler_conns:  &slc,
		Attributes_conns: &slc,
	}

	exp := ApierCfg{
		Enabled:         bl,
		CachesConns:     slc,
		SchedulerConns:  slc,
		AttributeSConns: slc,
	}

	err := a.loadFromJsonCfg(&js)

	if err != nil {
		t.Fatal("was not expecting an error", err)
	}

	if !reflect.DeepEqual(a, exp) {
		t.Errorf("recived %v, expected %v", a, exp)
	}

}

func TestApierCfgAsMapInterface2(t *testing.T) {
	bl := true
	slc := []string{"val1", "val2"}

	a := ApierCfg{
		Enabled:         bl,
		CachesConns:     slc,
		SchedulerConns:  slc,
		AttributeSConns: slc,
	}

	exp := map[string]any{
		utils.EnabledCfg:         a.Enabled,
		utils.CachesConnsCfg:     slc,
		utils.SchedulerConnsCfg:  slc,
		utils.AttributeSConnsCfg: slc,
	}

	rcv := a.AsMapInterface()

	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("recived %v, expected %v", rcv, exp)
	}
}
