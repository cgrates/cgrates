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
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/cgrates/cgrates/utils"
)

func TestLoaderSCfgloadFromJsonCfg(t *testing.T) {
	var loadscfg, expected LoaderSCfg
	if err := loadscfg.loadFromJsonCfg(nil, utils.INFIELD_SEP); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(loadscfg, expected) {
		t.Errorf("Expected: %+v ,received: %+v", expected, loadscfg)
	}
	if err := loadscfg.loadFromJsonCfg(new(LoaderJsonCfg), utils.INFIELD_SEP); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(loadscfg, expected) {
		t.Errorf("Expected: %+v ,received: %+v", expected, loadscfg)
	}
	cfgJSONStr := `{
"loaders": [
	{
		"id": "*default",									// identifier of the Loader
		"enabled": false,									// starts as service: <true|false>.
		"tenant": "cgrates.org",							// tenant used in filterS.Pass
		"dry_run": false,									// do not send the CDRs to CDRS, just parse them
		"run_delay": 0,										// sleep interval in seconds between consecutive runs, 0 to use automation via inotify
		"lock_filename": ".cgr.lck",						// Filename containing concurrency lock in case of delayed processing
		"caches_conns": ["*internal"],
		"field_separator": ",",								// separator used in case of csv files
		"tp_in_dir": "/var/spool/cgrates/loader/in",		// absolute path towards the directory where the CDRs are stored
		"tp_out_dir": "/var/spool/cgrates/loader/out",		// absolute path towards the directory where processed CDRs will be moved
		"data":[											// data profiles to load
			{
				"type": "*attributes",						// data source type
				"file_name": "Attributes.csv",				// file name in the tp_in_dir
				"fields": [
					{"tag": "TenantID", "path": "Tenant", "type": "*composed", "value": "~0", "mandatory": true},
				],
			},]
		}
	]
}`
	val, err := NewRSRParsers("~0", true, utils.INFIELD_SEP)
	if err != nil {
		t.Error(err)
	}
	ten, err := NewRSRParsers("cgrates.org", true, utils.INFIELD_SEP)
	if err != nil {
		t.Error(err)
	}
	expected = LoaderSCfg{
		Id:             utils.MetaDefault,
		Tenant:         ten,
		LockFileName:   ".cgr.lck",
		CacheSConns:    []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches)},
		FieldSeparator: ",",
		TpInDir:        "/var/spool/cgrates/loader/in",
		TpOutDir:       "/var/spool/cgrates/loader/out",
		Data: []*LoaderDataType{
			{
				Type:     "*attributes",
				Filename: "Attributes.csv",
				Fields: []*FCTemplate{
					{
						Tag:       "TenantID",
						Path:      "Tenant",
						pathSlice: []string{"Tenant"},
						pathItems: utils.PathItems{{Field: "Tenant"}},
						Type:      "*composed",
						Value:     val,
						Mandatory: true,
					},
				},
			},
		},
	}
	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnLoadersCfg, err := jsnCfg.LoaderJsonCfg(); err != nil {
		t.Error(err)
	} else if err = loadscfg.loadFromJsonCfg(jsnLoadersCfg[0], utils.INFIELD_SEP); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, loadscfg) {
		t.Errorf("Expected: %+v , received: %+v", utils.ToJSON(expected), utils.ToJSON(loadscfg))
	}
}

func TestLoaderCfgAsMapInterface(t *testing.T) {
	var loadscfg LoaderSCfg
	cfgJSONStr := `{
			"loaders": [												
	{
		"id": "*default",									
		"enabled": false,									
		"tenant": "",										
		"dry_run": false,									
		"run_delay": 0,										
		"lock_filename": ".cgr.lck",						
		"caches_conns": ["*internal"],
		"field_separator": ",",								
		"tp_in_dir": "/var/spool/cgrates/loader/in",		
		"tp_out_dir": "/var/spool/cgrates/loader/out",		
		"data":[											
			{
				"type": "*attributes",						
				"file_name": "Attributes.csv",				
				"fields": [
					{"tag": "TenantID", "path": "Tenant", "type": "*variable", "value": "~0", "mandatory": true},
					{"tag": "ProfileID", "path": "ID", "type": "*variable", "value": "~1", "mandatory": true},
					],
				},
			],
		},
	],
	
}`
	eMap := map[string]any{
		"id":              "*default",
		"enabled":         false,
		"tenant":          "",
		"dry_run":         false,
		"run_delay":       "0",
		"lock_filename":   ".cgr.lck",
		"caches_conns":    []string{"*internal"},
		"field_separator": ",",
		"tp_in_dir":       "/var/spool/cgrates/loader/in",
		"tp_out_dir":      "/var/spool/cgrates/loader/out",
		"data": []map[string]any{
			{
				"type":      "*attributes",
				"file_name": "Attributes.csv",
				"fields": []map[string]any{
					{
						"tag":       "TenantID",
						"path":      "Tenant",
						"type":      "*variable",
						"value":     "~0",
						"mandatory": true,
					}, {
						"tag":       "ProfileID",
						"path":      "ID",
						"type":      "*variable",
						"value":     "~1",
						"mandatory": true,
					},
				},
			},
		},
	}

	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnLoadersCfg, err := jsnCfg.LoaderJsonCfg(); err != nil {
		t.Error(err)
	} else if err = loadscfg.loadFromJsonCfg(jsnLoadersCfg[0], utils.INFIELD_SEP); err != nil {
		t.Error(err)
	} else if rcv := loadscfg.AsMapInterface(""); !reflect.DeepEqual(eMap, rcv) {
		t.Errorf("\nExpected: %+v\nReceived: %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}

}

func TestLoadersCFGEnable(t *testing.T) {

	lt := LoaderSCfg{
		Enabled: true,
	}
	lf := LoaderSCfg{
		Enabled: false,
	}

	lst := LoaderSCfgs{&lt}
	lsf := LoaderSCfgs{&lf}

	tests := []struct {
		name   string
		loader LoaderSCfgs
		exp    bool
	}{
		{
			name:   "return true",
			loader: lst,
			exp:    true,
		},
		{
			name:   "return false",
			loader: lsf,
			exp:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rcv := tt.loader.Enabled()

			if rcv != tt.exp {
				t.Errorf("recived %v, expected %v", rcv, tt.exp)
			}
		})
	}
}

func TestLoadersCFGLoadFromJsonCfg(t *testing.T) {

	ld := LoaderDataType{
		Type:     "test",
		Filename: "test",
		Fields:   []*FCTemplate{},
	}

	str := "`test"

	ljd := LoaderJsonDataType{
		Fields: &[]*FcTemplateJsonCfg{
			{Value: &str},
		},
	}

	type args struct {
		jsnCfg    *LoaderJsonDataType
		separator string
	}

	tests := []struct {
		name string
		args args
		exp  error
	}{
		{
			name: "nil return",
			args: args{jsnCfg: nil, separator: ""},
			exp:  nil,
		},
		{
			name: "error check",
			args: args{&ljd, ""},
			exp:  fmt.Errorf("Unclosed unspilit syntax"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rcv := ld.loadFromJsonCfg(tt.args.jsnCfg, tt.args.separator)

			if rcv != nil {
				if rcv.Error() != tt.exp.Error() {
					t.Errorf("recived %v, expected %v", rcv, tt.exp)
				}
			}
		})
	}
}

func TestLoaderSCfgloadFromJsonCfg2(t *testing.T) {
	str := "test)"
	self := &LoaderSCfg{}
	jsnCfg := &LoaderJsonCfg{
		Tenant: &str,
	}

	err := self.loadFromJsonCfg(jsnCfg, "")

	if err != nil {
		if err.Error() != "invalid RSRFilter start rule in string: <)>" {
			t.Error(err)
		}
	}

	jsnCfg2 := &LoaderJsonCfg{
		Caches_conns: &[]string{str},
	}

	err = self.loadFromJsonCfg(jsnCfg2, "")
	if err != nil {
		t.Error(err)
	}

	if self.CacheSConns[0] != str {
		t.Error(self.CacheSConns[0])
	}
}

func TestLoaderSCfgAsMapInterface(t *testing.T) {
	l := &LoaderSCfg{
		Tenant: RSRParsers{{
			Rules:           "test",
			AllFiltersMatch: true,
		}},
	}

	exp := map[string]any{
		utils.TenantCfg:         strings.Join([]string{"test"}, utils.EmptyString),
		utils.IdCfg:             l.Id,
		utils.EnabledCfg:        l.Enabled,
		utils.DryRunCfg:         l.DryRun,
		utils.RunDelayCfg:       "0",
		utils.LockFileNameCfg:   l.LockFileName,
		utils.CacheSConnsCfg:    []string{},
		utils.FieldSeparatorCfg: l.FieldSeparator,
		utils.TpInDirCfg:        l.TpInDir,
		utils.TpOutDirCfg:       l.TpOutDir,
		utils.DataCfg:           []map[string]any{},
	}
	rcv := l.AsMapInterface("")

	if !reflect.DeepEqual(exp, rcv) {
		t.Errorf("\nexpeting: %s\n received: %s\n", utils.ToJSON(exp), utils.ToJSON(rcv))
	}
}

func TestLoadersCfgNewDfltLoaderDataTypeConfig(t *testing.T) {
	dfltLoaderDataTypeConfig = nil
	result := NewDfltLoaderDataTypeConfig()
	if reflect.TypeOf(result) != reflect.TypeOf(&LoaderDataType{}) {
		t.Errorf("Expected type *LoaderDataType, got %T", result)
	}
	tData := &LoaderDataType{}
	dfltLoaderDataTypeConfig = tData
	result = NewDfltLoaderDataTypeConfig()
	if !reflect.DeepEqual(result, tData) {
		t.Errorf("Expected %+v, got %+v", tData, result)
	}
}
