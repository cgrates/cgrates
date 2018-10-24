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
	"time"

	"github.com/cgrates/cgrates/utils"
)

func TestCdrStatsCfgloadFromJsonCfg(t *testing.T) {
	var cdrscfg, expected CdrStatsCfg
	if err := cdrscfg.loadFromJsonCfg(nil); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(cdrscfg, expected) {
		t.Errorf("Expected: %+v ,recived: %+v", expected, cdrscfg)
	}
	if err := cdrscfg.loadFromJsonCfg(new(CdrStatsJsonCfg)); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(cdrscfg, expected) {
		t.Errorf("Expected: %+v ,recived: %+v", expected, cdrscfg)
	}
	cfgJSONStr := `{
"cdrstats": {
	"enabled": false,						// starts the cdrstats service: <true|false>
	"save_interval": "1m",					// interval to save changed stats into dataDb storage
	},
}`
	expected = CdrStatsCfg{
		CDRStatsEnabled:      false,
		CDRStatsSaveInterval: time.Duration(time.Minute),
	}
	if jsnCfg, err := NewCgrJsonCfgFromReader(strings.NewReader(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnCdrstatsCfg, err := jsnCfg.CdrStatsJsonCfg(); err != nil {
		t.Error(err)
	} else if err = cdrscfg.loadFromJsonCfg(jsnCdrstatsCfg); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, cdrscfg) {
		t.Errorf("Expected: %+v , recived: %+v", utils.ToJSON(expected), utils.ToJSON(cdrscfg))
	}
}
