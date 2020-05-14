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
)

func TestRateSConfigloadFromJsonCfg(t *testing.T) {
	var rateCfg, expected RateSCfg
	if err := rateCfg.loadFromJsonCfg(nil); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rateCfg, expected) {
		t.Errorf("Expected: %+v ,recived: %+v", expected, rateCfg)
	}
	if err := rateCfg.loadFromJsonCfg(new(RateSJsonCfg)); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rateCfg, expected) {
		t.Errorf("Expected: %+v ,recived: %+v", expected, rateCfg)
	}
	cfgJSONStr := `{
"rates": {					
	"enabled": true,		
},	
}`
	expected = RateSCfg{
		Enabled: true,
	}
	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnRateSCfg, err := jsnCfg.RateCfgJson(); err != nil {
		t.Error(err)
	} else if err = rateCfg.loadFromJsonCfg(jsnRateSCfg); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rateCfg) {
		t.Errorf("Expected: %+v , recived: %+v", expected, rateCfg)
	}
}
