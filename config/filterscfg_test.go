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

func TestFilterSCfgloadFromJsonCfg(t *testing.T) {
	var fscfg, expected FilterSCfg
	if err := fscfg.loadFromJsonCfg(nil); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(fscfg, expected) {
		t.Errorf("Expected: %+v ,recived: %+v", expected, fscfg)
	}
	if err := fscfg.loadFromJsonCfg(new(FilterSJsonCfg)); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(fscfg, expected) {
		t.Errorf("Expected: %+v ,recived: %+v", expected, fscfg)
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
		t.Errorf("Expected: %+v , recived: %+v", utils.ToJSON(expected), utils.ToJSON(fscfg))
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
	eMap := map[string]interface{}{
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
		t.Errorf("Expected: %+v ,\n recived: %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}
}
