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

func TestAnalyzerSCfgloadFromJsonCfg(t *testing.T) {
	jsonCfg := &AnalyzerSJsonCfg{
		Enabled: utils.BoolPointer(false),
	}
	expected := &AnalyzerSCfg{
		Enabled: false,
	}
	if jsnCfg, err := NewDefaultCGRConfig(); err != nil {
		t.Error(err)
	} else if err = jsnCfg.analyzerSCfg.loadFromJsonCfg(jsonCfg); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(jsnCfg.analyzerSCfg, expected) {
		t.Errorf("Expected %+v \n, received %+v", expected, jsnCfg.analyzerSCfg)
	}
}

func TestAnalyzerSCfgAsMapInterface(t *testing.T) {
	cfgJSONStr := `{
		"analyzers":{},
    }
}`
	eMap := map[string]interface{}{
		utils.EnabledCfg: false,
	}
	if cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cgrCfg.analyzerSCfg.AsMapInterface(); !reflect.DeepEqual(rcv, eMap) {
		t.Errorf("Expected: %+v , recived: %+v", eMap, rcv)
	}
}

func TestAnalyzerSCfgAsMapInterface1(t *testing.T) {
	cfgJSONStr := `{
		"analyzers":{
            "enabled": true,  
        },
    }
}`
	eMap := map[string]interface{}{
		utils.EnabledCfg: true,
	}
	if cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cgrCfg.analyzerSCfg.AsMapInterface(); !reflect.DeepEqual(rcv, eMap) {
		t.Errorf("Expected: %+v , recived: %+v", eMap, rcv)
	}
}
