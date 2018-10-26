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
)

func TestAttributeSCfgloadFromJsonCfg(t *testing.T) {
	var attscfg, expected AttributeSCfg
	if err := attscfg.loadFromJsonCfg(nil); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(attscfg, expected) {
		t.Errorf("Expected: %+v ,recived: %+v", expected, attscfg)
	}
	if err := attscfg.loadFromJsonCfg(new(AttributeSJsonCfg)); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(attscfg, expected) {
		t.Errorf("Expected: %+v ,recived: %+v", expected, attscfg)
	}
	cfgJSONStr := `{
"attributes": {								// Attribute service
	"enabled": true,						// starts attribute service: <true|false>.
	//"string_indexed_fields": [],			// query indexes based on these fields for faster processing
	"prefix_indexed_fields": ["index1","index2"],			// query indexes based on these fields for faster processing
	"process_runs": 1,						// number of run loops when processing event
	},		
}`
	expected = AttributeSCfg{
		Enabled:             true,
		PrefixIndexedFields: &[]string{"index1", "index2"},
		ProcessRuns:         1,
	}
	if jsnCfg, err := NewCgrJsonCfgFromReader(strings.NewReader(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnAttSCfg, err := jsnCfg.AttributeServJsonCfg(); err != nil {
		t.Error(err)
	} else if err = attscfg.loadFromJsonCfg(jsnAttSCfg); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, attscfg) {
		t.Errorf("Expected: %+v , recived: %+v", expected, attscfg)
	}
}
