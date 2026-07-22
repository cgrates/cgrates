// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package config

import (
	"reflect"
	"testing"

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
		Keys: []string{"key1", "key2"},
	}
	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnalS, err := jsnCfg.ApiBanCfgJson(); err != nil {
		t.Error(err)
	} else if err = alS.loadFromJSONCfg(jsnalS); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, alS) {
		t.Errorf("Expected: %+v , received: %+v", expected, alS)
	}
}

func TestAPIBanCfgAsMapInterface(t *testing.T) {
	var alS APIBanCfg
	cfgJSONStr := `{
		"apiban":{
			"keys": ["key1","key2"]
		},
		
}`
	eMap := map[string]any{
		"keys": []string{"key1", "key2"},
	}
	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnalS, err := jsnCfg.ApiBanCfgJson(); err != nil {
		t.Error(err)
	} else if err = alS.loadFromJSONCfg(jsnalS); err != nil {
		t.Error(err)
	} else if rcv := alS.AsMapInterface(); !reflect.DeepEqual(eMap, rcv) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}
}

func TestAPIBanCfgClone(t *testing.T) {
	ban := &APIBanCfg{
		Keys: []string{"key1", "key2"},
	}
	rcv := ban.Clone()
	if !reflect.DeepEqual(ban, rcv) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(ban), utils.ToJSON(rcv))
	}
	if rcv.Keys[0] = ""; ban.Keys[0] != "key1" {
		t.Errorf("Expected clone to not modify the cloned")
	}

	ban = nil
	rcv = ban.Clone()
	if !reflect.DeepEqual(ban, rcv) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(ban), utils.ToJSON(rcv))
	}
}
