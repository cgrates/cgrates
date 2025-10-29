/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/
package config

import (
	"reflect"
	"testing"

	"github.com/cgrates/cgrates/utils"
)

func TestAttributeSCfgloadFromJsonCfg(t *testing.T) {
	var attscfg, expected AttributeSCfg
	if err := attscfg.loadFromJsonCfg(nil); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(attscfg, expected) {
		t.Errorf("Expected: %+v ,received: %+v", expected, attscfg)
	}
	if err := attscfg.loadFromJsonCfg(new(AttributeSJsonCfg)); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(attscfg, expected) {
		t.Errorf("Expected: %+v ,received: %+v", expected, attscfg)
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
	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnAttSCfg, err := jsnCfg.AttributeServJsonCfg(); err != nil {
		t.Error(err)
	} else if err = attscfg.loadFromJsonCfg(jsnAttSCfg); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, attscfg) {
		t.Errorf("Expected: %+v , received: %+v", expected, attscfg)
	}
}

func TestAttributeSCfgAsMapInterface(t *testing.T) {
	var attscfg AttributeSCfg
	cfgJSONStr := `{
"attributes": {								
	"enabled": true,									
	"prefix_indexed_fields": ["index1","index2"],			
	"process_runs": 3,						
	},		
}`
	eMap := map[string]any{
		"enabled":               true,
		"prefix_indexed_fields": []string{"index1", "index2"},
		"process_runs":          3,
		"indexed_selects":       false,
		"nested_fields":         false,
		"string_indexed_fields": []string{},
	}
	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnAttSCfg, err := jsnCfg.AttributeServJsonCfg(); err != nil {
		t.Error(err)
	} else if err = attscfg.loadFromJsonCfg(jsnAttSCfg); err != nil {
		t.Error(err)
	} else if rcv := attscfg.AsMapInterface(); !reflect.DeepEqual(eMap, rcv) {
		t.Errorf("\nExpected: %+v\nReceived: %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}
}

func TestAttributeSCfgloadFromJsonCfg2(t *testing.T) {
	alS := AttributeSCfg{}

	js := AttributeSJsonCfg{
		String_indexed_fields: &[]string{"val1", "val2"},
	}

	err := alS.loadFromJsonCfg(&js)
	if err != nil {
		t.Error(err)
	}

	exp := &[]string{"val1", "val2"}

	if !reflect.DeepEqual(alS.StringIndexedFields, exp) {
		t.Errorf("recived %v, expected %v", alS.StringIndexedFields, exp)
	}
}

func TestAttributeSCfgAsMapInterface2(t *testing.T) {
	alS := AttributeSCfg{
		StringIndexedFields: &[]string{"val1", "val2"},
	}

	rcv := alS.AsMapInterface()
	exp := []string{"val1", "val2"}

	if !reflect.DeepEqual(rcv["string_indexed_fields"], exp) {
		t.Errorf("recived %v, expected %v", rcv["string_indexed_fields"], exp)
	}
}
