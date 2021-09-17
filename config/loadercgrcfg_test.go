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
		Actions_conns:    &[]string{utils.MetaInternal},
		Gapi_credentials: &json.RawMessage{12, 13, 60},
		Gapi_token:       &json.RawMessage{13, 16},
	}
	expected := &LoaderCgrCfg{
		TpID:            "randomID",
		DataPath:        "./",
		DisableReverse:  true,
		FieldSeparator:  ';',
		CachesConns:     []string{"*internal:*caches"},
		ActionSConns:    []string{"*internal:*actions"},
		GapiCredentials: json.RawMessage{12, 13, 60},
		GapiToken:       json.RawMessage{13, 16},
	}
	jsnCfg := NewDefaultCGRConfig()
	if err = jsnCfg.loaderCgrCfg.loadFromJSONCfg(cfgJSON); err != nil {
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
		"actions_conns": ["*internal","*localhost"],
		"gapi_credentials": ".gapi/credentials.json",
		"gapi_token": ".gapi/token.json"
	},
}`
	eMap := map[string]interface{}{
		utils.TpIDCfg:            "",
		utils.DataPathCfg:        "./",
		utils.DisableReverseCfg:  false,
		utils.FieldSepCfg:        ",",
		utils.CachesConnsCfg:     []string{"*internal", "*localhost"},
		utils.ActionSConnsCfg:    []string{"*internal", "*localhost"},
		utils.GapiCredentialsCfg: json.RawMessage(`".gapi/credentials.json"`),
		utils.GapiTokenCfg:       json.RawMessage(`".gapi/token.json"`),
	}
	if cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cgrCfg.loaderCgrCfg.AsMapInterface(""); !reflect.DeepEqual(rcv, eMap) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}
}

func TestLoaderCgrCfgClone(t *testing.T) {
	ban := &LoaderCgrCfg{
		TpID:            "randomID",
		DataPath:        "./",
		DisableReverse:  true,
		FieldSeparator:  ';',
		CachesConns:     []string{"*internal:*caches"},
		ActionSConns:    []string{"*internal:*scheduler"},
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
	if rcv.ActionSConns[0] = ""; ban.ActionSConns[0] != "*internal:*scheduler" {
		t.Errorf("Expected clone to not modify the cloned")
	}
	if rcv.GapiCredentials[0] = 0; ban.GapiCredentials[0] != 12 {
		t.Errorf("Expected clone to not modify the cloned")
	}
	if rcv.GapiToken[0] = 0; ban.GapiToken[0] != 13 {
		t.Errorf("Expected clone to not modify the cloned")
	}
}

func TestDiffLoaderCfgJson(t *testing.T) {
	var d *LoaderCfgJson

	v1 := &LoaderCgrCfg{
		TpID:            "loaderID",
		DataPath:        "/data/path",
		DisableReverse:  false,
		FieldSeparator:  rune(22),
		CachesConns:     []string{"*localhost"},
		ActionSConns:    []string{"*localhost"},
		GapiCredentials: json.RawMessage(`{"field1":"value1"}`),
		GapiToken:       json.RawMessage(`{"field1":"value1"}`),
	}

	v2 := &LoaderCgrCfg{
		TpID:            "loaderID2",
		DataPath:        "/data/path/2",
		DisableReverse:  true,
		FieldSeparator:  rune(97),
		CachesConns:     []string{"*birpc"},
		ActionSConns:    []string{"*birpc"},
		GapiCredentials: nil,
		GapiToken:       nil,
	}

	expected := &LoaderCfgJson{
		Tpid:             utils.StringPointer("loaderID2"),
		Data_path:        utils.StringPointer("/data/path/2"),
		Disable_reverse:  utils.BoolPointer(true),
		Field_separator:  utils.StringPointer("a"),
		Caches_conns:     &[]string{"*birpc"},
		Actions_conns:    &[]string{"*birpc"},
		Gapi_credentials: nil,
		Gapi_token:       nil,
	}

	rcv := diffLoaderCfgJson(d, v1, v2)
	if !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Expected %+v \n but received \n %+v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}

	v1 = v2
	expected = &LoaderCfgJson{}
	rcv = diffLoaderCfgJson(d, v1, v2)
	if !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Expected %+v \n but received \n %+v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}

func TestDiffLoaderCfgJsonCase2(t *testing.T) {
	var d *LoaderCfgJson

	v1 := &LoaderCgrCfg{
		TpID:            "loaderID",
		DataPath:        "/data/path",
		DisableReverse:  false,
		FieldSeparator:  rune(22),
		CachesConns:     []string{"*localhost"},
		ActionSConns:    []string{"*localhost"},
		GapiCredentials: nil,
		GapiToken:       nil,
	}

	v2 := &LoaderCgrCfg{
		TpID:            "loaderID2",
		DataPath:        "/data/path/2",
		DisableReverse:  true,
		FieldSeparator:  rune(97),
		CachesConns:     []string{"*birpc"},
		ActionSConns:    []string{"*birpc"},
		GapiCredentials: json.RawMessage(`{"field1":"value1"}`),
		GapiToken:       json.RawMessage(`{"field1":"value1"}`),
	}

	gapiC := json.RawMessage(`{"field1":"value1"}`)
	gapiT := json.RawMessage(`{"field1":"value1"}`)

	expected := &LoaderCfgJson{
		Tpid:             utils.StringPointer("loaderID2"),
		Data_path:        utils.StringPointer("/data/path/2"),
		Disable_reverse:  utils.BoolPointer(true),
		Field_separator:  utils.StringPointer("a"),
		Caches_conns:     &[]string{"*birpc"},
		Actions_conns:    &[]string{"*birpc"},
		Gapi_credentials: &gapiC,
		Gapi_token:       &gapiT,
	}

	rcv := diffLoaderCfgJson(d, v1, v2)
	if !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Expected %+v \n but received \n %+v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}
