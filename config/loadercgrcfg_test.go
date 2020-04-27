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

	"github.com/cgrates/cgrates/utils"
)

func TestLoaderCgrCfgloadFromJsonCfg(t *testing.T) {
	var loadscfg, expected LoaderCgrCfg
	if err := loadscfg.loadFromJsonCfg(nil); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(loadscfg, expected) {
		t.Errorf("Expected: %+v ,recived: %+v", expected, loadscfg)
	}
	if err := loadscfg.loadFromJsonCfg(new(LoaderCfgJson)); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(loadscfg, expected) {
		t.Errorf("Expected: %+v ,recived: %+v", expected, loadscfg)
	}
	cfgJSONStr := `{
"loader": {									// loader for tariff plans out of .csv files
	"tpid": "",								// tariff plan identificator
	"data_path": "",						// path towards tariff plan files
	"disable_reverse": false,				// disable reverse computing
	"field_separator": ";",					// separator used in case of csv files
	"caches_conns":["*localhost"],
	"scheduler_conns": ["*localhost"],
},
}`
	expected = LoaderCgrCfg{
		FieldSeparator: rune(';'),
		CachesConns:    []string{utils.MetaLocalHost},
		SchedulerConns: []string{utils.MetaLocalHost},
	}
	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnLoadersCfg, err := jsnCfg.LoaderCfgJson(); err != nil {
		t.Error(err)
	} else if err = loadscfg.loadFromJsonCfg(jsnLoadersCfg); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, loadscfg) {
		t.Errorf("Expected: %+v , recived: %+v", utils.ToJSON(expected), utils.ToJSON(loadscfg))
	}
}

func TestLoaderCgrCfgAsMapInterface(t *testing.T) {
	var loadscfg LoaderCgrCfg
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
		"tpid":             "",
		"data_path":        "./",
		"disable_reverse":  false,
		"field_separator":  ",",
		"caches_conns":     []string{"*localhost"},
		"scheduler_conns":  []string{"*localhost"},
		"gapi_credentials": ".gapi/credentials.json",
		"gapi_token":       ".gapi/token.json",
	}
	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnLoadersCfg, err := jsnCfg.LoaderCfgJson(); err != nil {
		t.Error(err)
	} else if err = loadscfg.loadFromJsonCfg(jsnLoadersCfg); err != nil {
		t.Error(err)
	} else if rcv := loadscfg.AsMapInterface(); !reflect.DeepEqual(eMap["tpid"], rcv["tpid"]) {
		t.Errorf("Expected: %+v, Recived: %+v, at field: '%s'", utils.ToJSON(eMap["tpid"]), utils.ToJSON(rcv["tpid"]), "tpid")
	} else if !reflect.DeepEqual(eMap["data_path"], rcv["data_path"]) {
		t.Errorf("Expected: %+v, Recived: %+v, at field: %s", utils.ToJSON(eMap["data_path"]), utils.ToJSON(rcv["data_path"]), "data_path")
	} else if !reflect.DeepEqual(eMap["disable_reverse"], rcv["disable_reverse"]) {
		t.Errorf("Expected: %+v, Recived: %+v, at field: %s", utils.ToJSON(eMap["disable_reverse"]), utils.ToJSON(rcv["disable_reverse"]), "disable_reverse")
	} else if !reflect.DeepEqual(eMap["field_separator"], rcv["field_separator"]) {
		t.Errorf("Expected: %+v, Recived: %+v, at field: %s", utils.ToJSON(eMap["field_separator"]), utils.ToJSON(rcv["field_separator"]), "field_separator")
	} else if !reflect.DeepEqual(eMap["caches_conns"], rcv["caches_conns"]) {
		t.Errorf("Expected: %+v, Recived: %+v, at field: %s", utils.ToJSON(eMap["caches_conns"]), utils.ToJSON(rcv["caches_conns"]), "caches_conns")
	} else if !reflect.DeepEqual(eMap["scheduler_conns"], rcv["scheduler_conns"]) {
		t.Errorf("Expected: %+v, Recived: %+v, at field: %s", utils.ToJSON(eMap["scheduler_conns"]), utils.ToJSON(rcv["scheduler_conns"]), "scheduler_conns")
	}
}
