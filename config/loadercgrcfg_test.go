// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package config

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/cgrates/cgrates/utils"
)

func TestLoaderCgrCfgloadFromJsonCfg(t *testing.T) {
	cfgJSON := &LoaderCfgJson{
		Tpid:             utils.StringPointer("randomID"),
		Data_path:        utils.StringPointer("./"),
		Disable_reverse:  utils.BoolPointer(true),
		Field_separator:  utils.StringPointer(";"),
		Caches_conns:     &[]string{utils.MetaInternal},
		Scheduler_conns:  &[]string{utils.MetaInternal},
		Gapi_credentials: &json.RawMessage{12, 13, 60},
		Gapi_token:       &json.RawMessage{13, 16},
	}
	expected := &LoaderCgrCfg{
		TpID:            "randomID",
		DataPath:        "./",
		DisableReverse:  true,
		FieldSeparator:  rune(';'),
		CachesConns:     []string{"*internal:*caches"},
		SchedulerConns:  []string{"*internal:*scheduler"},
		GapiCredentials: json.RawMessage{12, 13, 60},
		GapiToken:       json.RawMessage{13, 16},
	}
	jsnCfg := NewDefaultCGRConfig()
	if err := jsnCfg.loaderCgrCfg.loadFromJSONCfg(cfgJSON); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, jsnCfg.loaderCgrCfg) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(jsnCfg.loaderCgrCfg))
	}
}

func TestLoaderCgrCfgAsMapInterface(t *testing.T) {
	cfgJSONStr := `{
	"loader": {
		"tpid": "",
		"data_path": "./",
		"disable_reverse": false,
		"field_separator": ",",
		"caches_conns":["*internal","*localhost"],
		"scheduler_conns": ["*internal","*localhost"],
		"gapi_credentials": ".gapi/credentials.json",
		"gapi_token": ".gapi/token.json"
	},
}`
	eMap := map[string]any{
		utils.TpIDCfg:            "",
		utils.DataPathCfg:        "./",
		utils.DisableReverseCfg:  false,
		utils.FieldSepCfg:        ",",
		utils.CachesConnsCfg:     []string{"*internal", "*localhost"},
		utils.SchedulerConnsCfg:  []string{"*internal", "*localhost"},
		utils.GapiCredentialsCfg: json.RawMessage(`".gapi/credentials.json"`),
		utils.GapiTokenCfg:       json.RawMessage(`".gapi/token.json"`),
	}
	if cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cgrCfg.loaderCgrCfg.AsMapInterface(); !reflect.DeepEqual(rcv, eMap) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}
}

func TestLoaderCgrCfgClone(t *testing.T) {
	ban := &LoaderCgrCfg{
		TpID:            "randomID",
		DataPath:        "./",
		DisableReverse:  true,
		FieldSeparator:  rune(';'),
		CachesConns:     []string{"*internal:*caches"},
		SchedulerConns:  []string{"*internal:*scheduler"},
		GapiCredentials: json.RawMessage{12, 13, 60},
		GapiToken:       json.RawMessage{13, 16},
	}
	rcv := ban.Clone()
	if !reflect.DeepEqual(ban, rcv) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(ban), utils.ToJSON(rcv))
	}
	if rcv.CachesConns[0] = ""; ban.CachesConns[0] != "*internal:*caches" {
		t.Errorf("Expected clone to not modify the cloned")
	}
	if rcv.SchedulerConns[0] = ""; ban.SchedulerConns[0] != "*internal:*scheduler" {
		t.Errorf("Expected clone to not modify the cloned")
	}
	if rcv.GapiCredentials[0] = 0; ban.GapiCredentials[0] != 12 {
		t.Errorf("Expected clone to not modify the cloned")
	}
	if rcv.GapiToken[0] = 0; ban.GapiToken[0] != 13 {
		t.Errorf("Expected clone to not modify the cloned")
	}

	ban = nil
	rcv = ban.Clone()
	if !reflect.DeepEqual(ban, rcv) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(ban), utils.ToJSON(rcv))
	}
}
