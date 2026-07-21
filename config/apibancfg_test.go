// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

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
	eMap := map[string]any{
		"enabled": false,
		"keys":    []string{"key1", "key2"},
	}
	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if err = alS.Load(context.Background(), jsnCfg, nil); err != nil {
		t.Error(err)
	} else if rcv := alS.AsMapInterface(); !reflect.DeepEqual(eMap, rcv) {
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

func TestAPIBanCloneSection(t *testing.T) {
	apbCfg := &APIBanCfg{
		Enabled: false,
		Keys:    []string{"key1", "key2"},
	}

	exp := &APIBanCfg{
		Enabled: false,
		Keys:    []string{"key1", "key2"},
	}
	rcv := apbCfg.CloneSection()
	if !reflect.DeepEqual(exp, rcv) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(rcv))
	}
}
