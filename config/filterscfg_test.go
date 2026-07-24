// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package config

import (
	"reflect"
	"testing"

	"github.com/cgrates/cgrates/utils"
)

func TestFilterSCfgloadFromJsonCfg(t *testing.T) {
	var fscfg, expected FilterSCfg
	if err := fscfg.loadFromJsonCfg(nil); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(fscfg, expected) {
		t.Errorf("Expected: %+v ,received: %+v", expected, fscfg)
	}
	if err := fscfg.loadFromJsonCfg(new(FilterSJsonCfg)); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(fscfg, expected) {
		t.Errorf("Expected: %+v ,received: %+v", expected, fscfg)
	}
	cfgJSONStr := `{
"filters": {								// Filters configuration (*new)
	"stats_conns": ["*localhost"],		// address where to reach the stat service, empty to disable stats functionality: <""|*internal|x.y.z.y:1234>
	},
}`
	expected = FilterSCfg{
		StatSConns: []string{utils.MetaLocalHost},
	}
	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnFsCfg, err := jsnCfg.FilterSJsonCfg(); err != nil {
		t.Error(err)
	} else if err = fscfg.loadFromJsonCfg(jsnFsCfg); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, fscfg) {
		t.Errorf("Expected: %+v , received: %+v", utils.ToJSON(expected), utils.ToJSON(fscfg))
	}
}

func TestFilterSCfgAsMapInterface(t *testing.T) {
	var fscfg FilterSCfg
	cfgJSONStr := `{
		"filters": {								
			"stats_conns": ["*localhost"],						
			"resources_conns": [],					
			"apiers_conns": [],						
	},
}`
	var emptySlice []string
	eMap := map[string]any{
		"stats_conns":     []string{"*localhost"},
		"resources_conns": emptySlice,
		"apiers_conns":    emptySlice,
	}
	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnFsCfg, err := jsnCfg.FilterSJsonCfg(); err != nil {
		t.Error(err)
	} else if err = fscfg.loadFromJsonCfg(jsnFsCfg); err != nil {
		t.Error(err)
	} else if rcv := fscfg.AsMapInterface(); reflect.DeepEqual(eMap, rcv) {
		t.Errorf("Expected: %+v ,\n received: %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}
}

func TestFilterSCfgloadFromJsonCfg2(t *testing.T) {
	fSCfg := FilterSCfg{
		StatSConns:     []string{},
		ResourceSConns: []string{},
	}
	jsnCfg := &FilterSJsonCfg{
		Stats_conns:     &[]string{utils.MetaInternal},
		Resources_conns: &[]string{utils.MetaInternal},
	}

	err := fSCfg.loadFromJsonCfg(jsnCfg)
	exp := FilterSCfg{
		StatSConns:     []string{"*internal:*stats"},
		ResourceSConns: []string{"*internal:*resources"},
	}
	if err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(exp, fSCfg) {
		t.Errorf("expected %s, received %s", utils.ToJSON(exp), utils.ToJSON(fSCfg))
	}

	jsnCfg2 := &FilterSJsonCfg{
		Stats_conns:     &[]string{utils.MetaInternal},
		Resources_conns: &[]string{"test"},
	}

	err = fSCfg.loadFromJsonCfg(jsnCfg2)
	exp = FilterSCfg{
		StatSConns:     []string{"*internal:*stats"},
		ResourceSConns: []string{"test"},
	}
	if err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(exp, fSCfg) {
		t.Errorf("expected %s, received %s", utils.ToJSON(exp), utils.ToJSON(fSCfg))
	}
}
