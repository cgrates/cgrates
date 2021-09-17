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

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/utils"
)

func TestAPIBanCfgloadFromJsonCfg(t *testing.T) {
	var alS, expected APIBanCfg
	if err := alS.loadFromJSONCfg(nil); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(alS, expected) {
		t.Errorf("Expected: %+v ,received: %+v", expected, alS)
	}
	if err := alS.loadFromJSONCfg(new(APIBanJsonCfg)); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(alS, expected) {
		t.Errorf("Expected: %+v ,received: %+v", expected, alS)
	}
	cfgJSONStr := `{
		"apiban":{								// APIBan config
			"enabled":false,							// starts APIBan service: <true|false>.
			"keys": ["key1","key2"]
		},
		
}`
	expected = APIBanCfg{
		Enabled: false,
		Keys:    []string{"key1", "key2"},
	}
	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if err = alS.Load(context.Background(), jsnCfg, nil); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, alS) {
		t.Errorf("Expected: %+v , received: %+v", expected, alS)
	}
}

func TestAPIBanCfgAsMapInterface(t *testing.T) {
	var alS APIBanCfg
	cfgJSONStr := `{
		"apiban":{
			"enabled":false,
			"keys": ["key1","key2"]
		},
		
}`
	eMap := map[string]interface{}{
		"enabled": false,
		"keys":    []string{"key1", "key2"},
	}
	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if err = alS.Load(context.Background(), jsnCfg, nil); err != nil {
		t.Error(err)
	} else if rcv := alS.AsMapInterface(""); !reflect.DeepEqual(eMap, rcv) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}
}

func TestAPIBanCfgClone(t *testing.T) {
	ban := &APIBanCfg{
		Enabled: false,
		Keys:    []string{"key1", "key2"},
	}
	rcv := ban.Clone()
	if !reflect.DeepEqual(ban, rcv) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(ban), utils.ToJSON(rcv))
	}
	if rcv.Keys[0] = ""; ban.Keys[0] != "key1" {
		t.Errorf("Expected clone to not modify the cloned")
	}
}

func TestDiffAPIBanJsonCfg(t *testing.T) {
	var d *APIBanJsonCfg

	v1 := &APIBanCfg{
		Enabled: false,
		Keys:    []string{"key1", "key2"},
	}

	v2 := &APIBanCfg{
		Enabled: true,
		Keys:    []string{"key3", "key4"},
	}

	expected := &APIBanJsonCfg{
		Enabled: utils.BoolPointer(true),
		Keys:    &[]string{"key3", "key4"},
	}

	rcv := diffAPIBanJsonCfg(d, v1, v2)
	if !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}

	v2 = v1
	expected2 := &APIBanJsonCfg{}

	rcv = diffAPIBanJsonCfg(d, v1, v2)
	if !reflect.DeepEqual(rcv, expected2) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected2), utils.ToJSON(rcv))
	}
}
