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
	"strings"
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
	"stats_conns": [{"Address":"127.0.0.1","Transport":"","Synchronous":true}],		// address where to reach the stat service, empty to disable stats functionality: <""|*internal|x.y.z.y:1234>
	"indexed_selects":true,					// enable profile matching exclusively on indexes
	},
}`
	expected = FilterSCfg{
		IndexedSelects: true,
		StatSConns:     []*HaPoolConfig{{Address: "127.0.0.1", Transport: "", Synchronous: true}},
	}
	if jsnCfg, err := NewCgrJsonCfgFromReader(strings.NewReader(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnFsCfg, err := jsnCfg.FilterSJsonCfg(); err != nil {
		t.Error(err)
	} else if err = fscfg.loadFromJsonCfg(jsnFsCfg); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, fscfg) {
		t.Errorf("Expected: %+v , recived: %+v", utils.ToJSON(expected), utils.ToJSON(fscfg))
	}
}
