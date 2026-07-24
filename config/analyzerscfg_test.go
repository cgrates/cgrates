// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package config

import (
	"reflect"
	"testing"

	"github.com/cgrates/cgrates/utils"
)

func TestAnalyzerSCfgloadFromJsonCfg(t *testing.T) {
	var alS, expected AnalyzerSCfg
	if err := alS.loadFromJsonCfg(nil); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(alS, expected) {
		t.Errorf("Expected: %+v ,received: %+v", expected, alS)
	}
	if err := alS.loadFromJsonCfg(new(AnalyzerSJsonCfg)); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(alS, expected) {
		t.Errorf("Expected: %+v ,received: %+v", expected, alS)
	}
	cfgJSONStr := `{
		"analyzers":{								// AnalyzerS config
			"enabled":false							// starts AnalyzerS service: <true|false>.
		},
		
}`
	expected = AnalyzerSCfg{
		Enabled: false,
	}
	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnalS, err := jsnCfg.AnalyzerCfgJson(); err != nil {
		t.Error(err)
	} else if err = alS.loadFromJsonCfg(jsnalS); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, alS) {
		t.Errorf("Expected: %+v , received: %+v", expected, alS)
	}
}

func TestAnalyzerSCfgAsMapInterface(t *testing.T) {
	var alS AnalyzerSCfg
	cfgJSONStr := `{
		"analyzers":{
			"enabled":false
		},
		
}`
	eMap := map[string]any{
		"enabled": false,
	}
	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnalS, err := jsnCfg.AnalyzerCfgJson(); err != nil {
		t.Error(err)
	} else if err = alS.loadFromJsonCfg(jsnalS); err != nil {
		t.Error(err)
	} else if rcv := alS.AsMapInterface(); !reflect.DeepEqual(eMap, rcv) {
		t.Errorf("\nExpected: %+v\nReceived: %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}
}
