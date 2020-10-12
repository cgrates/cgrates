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
	if jsnCfg, err := NewDefaultCGRConfig(); err != nil {
		t.Error(err)
	} else if err = jsnCfg.loaderCgrCfg.loadFromJsonCfg(cfgJSON); err != nil {
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
		"caches_conns":["*localhost"],
		"scheduler_conns": ["*localhost"],
		"gapi_credentials": ".gapi/credentials.json",
		"gapi_token": ".gapi/token.json"
	},
}`
	eMap := map[string]interface{}{
		utils.TpIDCfg:            "",
		utils.DataPathCfg:        "./",
		utils.DisableReverseCfg:  false,
		utils.FieldSeparatorCfg:  ",",
		utils.CachesConnsCfg:     []string{"*localhost"},
		utils.SchedulerConnsCfg:  []string{"*localhost"},
		utils.GapiCredentialsCfg: json.RawMessage(`".gapi/credentials.json"`),
		utils.GapiTokenCfg:       json.RawMessage(`".gapi/token.json"`),
	}
	if cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cgrCfg.loaderCgrCfg.AsMapInterface(); !reflect.DeepEqual(rcv, eMap) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}
}
