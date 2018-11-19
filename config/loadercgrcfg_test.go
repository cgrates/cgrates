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

func TestLoaderCgrCfgloadFromJsonCfg(t *testing.T) {
	var loadscfg, expected LoaderCgrCfg
	if err := loadscfg.loadFromJsonCfg(nil); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(loadscfg, expected) {
		t.Errorf("Expected: %+v ,recived: %+v", expected, loadscfg)
	}
	if err := loadscfg.loadFromJsonCfg(new(LoaderCfgJson)); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(loadscfg, expected) {
		t.Errorf("Expected: %+v ,recived: %+v", expected, loadscfg)
	}
	cfgJSONStr := `{
"loader": {									// loader for tariff plans out of .csv files
	"tpid": "",								// tariff plan identificator
	"data_path": "",						// path towards tariff plan files
	"disable_reverse": false,				// disable reverse computing
	"field_separator": ";",					// separator used in case of csv files
	"caches_conns":[						// addresses towards cacheS components for reloads
		{"address": "127.0.0.1:2012", "transport": "*json"}
	],
	"scheduler_conns": [
		{"address": "127.0.0.1:2012"}
	],
},
}`
	expected = LoaderCgrCfg{
		FieldSeparator: rune(';'),
		CachesConns:    []*HaPoolConfig{{Address: "127.0.0.1:2012", Transport: "*json"}},
		SchedulerConns: []*HaPoolConfig{{Address: "127.0.0.1:2012"}},
	}
	if jsnCfg, err := NewCgrJsonCfgFromReader(strings.NewReader(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnLoadersCfg, err := jsnCfg.LoaderCfgJson(); err != nil {
		t.Error(err)
	} else if err = loadscfg.loadFromJsonCfg(jsnLoadersCfg); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, loadscfg) {
		t.Errorf("Expected: %+v , recived: %+v", utils.ToJSON(expected), utils.ToJSON(loadscfg))
	}
}
