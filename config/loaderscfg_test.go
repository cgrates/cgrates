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
	"path"
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/utils"
)

func TestLoaderSCfgloadFromJsonCfgCase1(t *testing.T) {
	cfgJSONStr := `{
			"loaders": [
	{
		"id": "*default",
		"enabled": true,
		"tenant": "cgrates.org",
		"lockfile_path": ".cgr.lck",
		"caches_conns": ["*internal","*conn1"],
		"field_separator": ",",
		"tp_in_dir": "/var/spool/cgrates/loader/in",
		"tp_out_dir": "/var/spool/cgrates/loader/out",
		"data":[
			{
				"type": "*attributes",
				"file_name": "Attributes.csv",
                "flags": [],
				"fields": [
					{"tag": "TenantID", "path": "Tenant", "type": "*composed", "value": "~req.0", "mandatory": true,"layout": "2006-01-02T15:04:05Z07:00"},
					],
				},
			],
		},
	],
}`
	val, err := NewRSRParsers("~req.0", utils.InfieldSep)
	if err != nil {
		t.Error(err)
	}
	ten := "cgrates.org"

	expected := LoaderSCfgs{
		{
			Enabled:        true,
			ID:             utils.MetaDefault,
			Tenant:         ten,
			LockFilePath:   ".cgr.lck",
			CacheSConns:    []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches), "*conn1"},
			FieldSeparator: ",",
			TpInDir:        "/var/spool/cgrates/loader/in",
			TpOutDir:       "/var/spool/cgrates/loader/out",
			Data: []*LoaderDataType{
				{
					Type:     "*attributes",
					Filename: "Attributes.csv",
					Flags:    utils.FlagsWithParams{},
					Fields: []*FCTemplate{
						{
							Tag:       "TenantID",
							Path:      "Tenant",
							pathSlice: []string{"Tenant"},
							Type:      utils.MetaComposed,
							Value:     val,
							Mandatory: true,
							Layout:    time.RFC3339,
						},
					},
				},
			},
		},
	}
	newCfg := new(CGRConfig)
	newCfg.generalCfg = new(GeneralCfg)
	newCfg.generalCfg.RSRSep = ";"
	if jsonCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if err = newCfg.loadLoaderSCfg(jsonCfg); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, newCfg.loaderCfg) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(newCfg.loaderCfg))
	}
}

// func TestLoaderSCfgloadFromJsonCfgCase2(t *testing.T) {
// 	cfgJSON := &LoaderJsonCfg{
// 		Tenant: utils.StringPointer("a{*"),
// 	}
// 	expected := "invalid converter terminator in rule: <a{*>"
// 	jsonCfg := NewDefaultCGRConfig()
// 	if err = jsonCfg.loaderCfg[0].loadFromJSONCfg(nil, jsonCfg.templates, jsonCfg.generalCfg.RSRSep); err != nil {
// 		t.Error(err)
// 	} else if err = jsonCfg.loaderCfg[0].loadFromJSONCfg(cfgJSON, jsonCfg.templates, jsonCfg.generalCfg.RSRSep); err == nil || err.Error() != expected {
// 		t.Errorf("Expected %+v, received %+v", expected, err)
// 	}
// }

func TestLoaderSCfgloadFromJsonCfgCase3(t *testing.T) {
	cfg := &LoaderJsonCfg{
		Data: &[]*LoaderJsonDataType{
			{
				Fields: &[]*FcTemplateJsonCfg{
					{
						Value: utils.StringPointer("a{*"),
					},
				},
			},
		},
	}
	expected := "invalid converter terminator in rule: <a{*>"
	jsonCfg := NewDefaultCGRConfig()
	if err := jsonCfg.loaderCfg[0].loadFromJSONCfg(cfg, jsonCfg.templates, jsonCfg.generalCfg.RSRSep); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestLoaderSCfgloadFromJsonCfgCase4(t *testing.T) {
	cfg := &LoaderJsonCfg{
		Data: &[]*LoaderJsonDataType{
			{
				Fields: &[]*FcTemplateJsonCfg{
					{
						Type: utils.StringPointer(utils.MetaTemplate),
					},
				},
			},
		},
	}
	expected := "no template with id: <>"
	jsonCfg := NewDefaultCGRConfig()
	if err = jsonCfg.loaderCfg[0].loadFromJSONCfg(cfg, jsonCfg.templates, jsonCfg.generalCfg.RSRSep); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestLoaderSCfgloadFromJsonCfgCase5(t *testing.T) {
	cfg := &LoaderJsonCfg{
		Data: &[]*LoaderJsonDataType{
			{
				Fields: &[]*FcTemplateJsonCfg{
					{
						Tag:   utils.StringPointer("randomTag"),
						Path:  utils.StringPointer("randomPath"),
						Type:  utils.StringPointer(utils.MetaTemplate),
						Value: utils.StringPointer("randomTemplate"),
					},
				},
			},
		},
	}
	expectedFields := LoaderSCfgs{
		{
			Data: []*LoaderDataType{
				{
					Fields: []*FCTemplate{
						{
							Tag:       "TenantID",
							Path:      "Tenant",
							Type:      utils.MetaVariable,
							Value:     NewRSRParsersMustCompile("~*req.0", utils.InfieldSep),
							Mandatory: true,
						},
					},
				},
			},
		},
	}
	msgTemplates := map[string][]*FCTemplate{
		"randomTemplate": {
			{
				Tag:       "TenantID",
				Path:      "Tenant",
				Type:      utils.MetaVariable,
				Value:     NewRSRParsersMustCompile("~*req.0", utils.InfieldSep),
				Mandatory: true,
			},
		},
	}
	jsonCfg := NewDefaultCGRConfig()
	if err = jsonCfg.loaderCfg[0].loadFromJSONCfg(cfg, msgTemplates, jsonCfg.generalCfg.RSRSep); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(jsonCfg.loaderCfg[0].Data[0].Fields[0], expectedFields[0].Data[0].Fields[0]) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expectedFields[0].Data[0].Fields[0]), utils.ToJSON(jsonCfg.loaderCfg[0].Data[0].Fields[0]))
	}

	if err := jsonCfg.loaderCfg[0].loadFromJSONCfg(nil, msgTemplates, jsonCfg.generalCfg.RSRSep); err != nil {
		t.Error(err)
	}
}

func TestLoaderSCfgloadFromJsonCfgCase6(t *testing.T) {
	cfg := &LoaderJsonCfg{
		Data: &[]*LoaderJsonDataType{nil},
	}
	jsonCfg := NewDefaultCGRConfig()
	if err = jsonCfg.loaderCfg[0].loadFromJSONCfg(cfg, jsonCfg.templates, jsonCfg.generalCfg.RSRSep); err != nil {
		t.Error(err)
	}
}

func TestEnabledCase1(t *testing.T) {
	jsonCfg := NewDefaultCGRConfig()

	if enabled := jsonCfg.loaderCfg.Enabled(); enabled {
		t.Errorf("Expected %+v", enabled)
	}
}
func TestEnabledCase2(t *testing.T) {
	cfgJSONStr := `{
			"loaders": [												
	{
		"enabled": true,
		},
	],	
}`
	if jsonCfg, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if enabled := jsonCfg.loaderCfg.Enabled(); !enabled {
		t.Errorf("Expected %+v", enabled)
	}
}

func TestLoaderCfgloadFromJsonCfg(t *testing.T) {
	cfgJSONStr := `{
			"loaders": [												
	{
		"enabled": true,
		"run_delay": "1sa",										
	},
	],	
}`
	expected := "time: unknown unit \"sa\" in duration \"1sa\""
	if _, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err == nil || err.Error() != expected {
		t.Errorf("Expected error: %s ,received: %v", expected, err)
	}
}

func TestLoaderCfgAsMapInterfaceCase1(t *testing.T) {
	cfgJSONStr := `{
			"loaders": [												
	{
		"id": "*default",									
		"enabled": false,									
		"tenant": "~*req.Destination1",										
		"dry_run": false,									
		"run_delay": "0",										
		"lockfile_path": ".cgr.lck",						
		"caches_conns": ["*internal:*caches"],
		"field_separator": ",",								
		"tp_in_dir": "/var/spool/cgrates/loader/in",		
		"tp_out_dir": "/var/spool/cgrates/loader/out",		
		"data":[											
			{
				"type": "*attributes",						
				"file_name": "Attributes.csv",				
				"fields": [
					{"tag": "TenantID", "path": "Tenant", "type": "*variable", "value": "~req.0", "mandatory": true},
					{"tag": "ProfileID", "path": "ID", "type": "*variable", "value": "~*req.1", "mandatory": true},
					],
				},
			],
		},
	],	
}`
	eMap := []map[string]interface{}{
		{
			utils.IDCfg:           "*default",
			utils.EnabledCfg:      false,
			utils.TenantCfg:       "~*req.Destination1",
			utils.DryRunCfg:       false,
			utils.RunDelayCfg:     "0",
			utils.LockFilePathCfg: ".cgr.lck",
			utils.CachesConnsCfg:  []string{utils.MetaInternal},
			utils.FieldSepCfg:     ",",
			utils.TpInDirCfg:      "/var/spool/cgrates/loader/in",
			utils.TpOutDirCfg:     "/var/spool/cgrates/loader/out",
			utils.DataCfg: []map[string]interface{}{
				{
					utils.TypeCfg:     "*attributes",
					utils.FilenameCfg: "Attributes.csv",
					utils.FieldsCfg: []map[string]interface{}{
						{
							utils.TagCfg:       "TenantID",
							utils.PathCfg:      "Tenant",
							utils.TypeCfg:      "*variable",
							utils.ValueCfg:     "~*req.0",
							utils.MandatoryCfg: true,
						}, {
							utils.TagCfg:       "ProfileID",
							utils.PathCfg:      "ID",
							utils.TypeCfg:      "*variable",
							utils.ValueCfg:     "~*req.1",
							utils.MandatoryCfg: true,
						},
					},
				},
			},
		},
	}
	if cfgCgr, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else {
		rcv := cgrCfg.loaderCfg.AsMapInterface(cfgCgr.generalCfg.RSRSep)
		if !reflect.DeepEqual(eMap[0][utils.DataCfg].([]map[string]interface{})[0][utils.FieldsCfg].([]map[string]interface{})[0],
			rcv[0][utils.DataCfg].([]map[string]interface{})[0][utils.FieldsCfg].([]map[string]interface{})[0]) {
			t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(eMap[0][utils.DataCfg].([]map[string]interface{})[0][utils.FieldsCfg].([]map[string]interface{})[0]),
				utils.ToJSON(rcv[0][utils.DataCfg].([]map[string]interface{})[0][utils.FieldsCfg].([]map[string]interface{})[0]))
		} else if !reflect.DeepEqual(eMap[0][utils.DataCfg].([]map[string]interface{})[0][utils.FieldsCfg].([]map[string]interface{})[1],
			rcv[0][utils.DataCfg].([]map[string]interface{})[0][utils.FieldsCfg].([]map[string]interface{})[1]) {
			t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(eMap[0][utils.DataCfg].([]map[string]interface{})[0][utils.FieldsCfg].([]map[string]interface{})[1]),
				utils.ToJSON(rcv[0][utils.DataCfg].([]map[string]interface{})[0][utils.FieldsCfg].([]map[string]interface{})[1]))
		} else if !reflect.DeepEqual(eMap[0][utils.CachesConnsCfg], rcv[0][utils.CachesConnsCfg]) {
			t.Errorf("Expected %+v, received %+v", eMap[0][utils.CachesConnsCfg], rcv[0][utils.CachesConnsCfg])
		} else if !reflect.DeepEqual(eMap[0][utils.TpInDirCfg], rcv[0][utils.TpInDirCfg]) {
			t.Errorf("Expected %+v, received %+v", eMap[0][utils.TpInDirCfg], rcv[0][utils.TpInDirCfg])
		} else if !reflect.DeepEqual(eMap[0][utils.LockFilePathCfg], rcv[0][utils.LockFilePathCfg]) {
			t.Errorf("Expected %+v, received %+v", eMap[0][utils.LockFilePathCfg], rcv[0][utils.LockFilePathCfg])
		}
	}
}

func TestLoaderCfgAsMapInterfaceCase2(t *testing.T) {
	cfgJSONStr := `{
			"loaders": [												
	{
		"id": "*default",									
		"enabled": false,									
		"tenant": "~*req.Destination1",										
		"dry_run": false,									
		"run_delay": "1",										
		"lockfile_path": ".cgr.lck",						
		"caches_conns": ["*conn1"],
		"field_separator": ",",								
		"tp_in_dir": "/var/spool/cgrates/loader/in",		
		"tp_out_dir": "/var/spool/cgrates/loader/out",		
		"data":[											
			{
				"type": "*attributes",						
				"file_name": "Attributes.csv",				
				"fields": [
					{"tag": "TenantID", "path": "Tenant", "type": "*variable", "value": "~req.0", "mandatory": true},
					{"tag": "ProfileID", "path": "ID", "type": "*variable", "value": "~*req.1", "mandatory": true},
					],
				},
			],
		},
	],	
}`
	eMap := []map[string]interface{}{
		{
			utils.IDCfg:           "*default",
			utils.EnabledCfg:      false,
			utils.TenantCfg:       "~*req.Destination1",
			utils.DryRunCfg:       false,
			utils.RunDelayCfg:     "0",
			utils.LockFilePathCfg: ".cgr.lck",
			utils.CachesConnsCfg:  []string{"*conn1"},
			utils.FieldSepCfg:     ",",
			utils.TpInDirCfg:      "/var/spool/cgrates/loader/in",
			utils.TpOutDirCfg:     "/var/spool/cgrates/loader/out",
			utils.DataCfg: []map[string]interface{}{
				{
					utils.TypeCfg:     "*attributes",
					utils.FilenameCfg: "Attributes.csv",
					utils.FieldsCfg: []map[string]interface{}{
						{
							utils.TagCfg:       "TenantID",
							utils.PathCfg:      "Tenant",
							utils.TypeCfg:      "*variable",
							utils.ValueCfg:     "~*req.0",
							utils.MandatoryCfg: true,
						}, {
							utils.TagCfg:       "ProfileID",
							utils.PathCfg:      "ID",
							utils.TypeCfg:      "*variable",
							utils.ValueCfg:     "~*req.1",
							utils.MandatoryCfg: true,
						},
					},
				},
			},
		},
	}
	if jsonCfg, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if rcv := jsonCfg.loaderCfg.AsMapInterface(jsonCfg.generalCfg.RSRSep); !reflect.DeepEqual(rcv[0][utils.Tenant], eMap[0][utils.Tenant]) {
		t.Errorf("Expected %+v, received %+v", rcv[0][utils.Tenant], eMap[0][utils.Tenant])
	}
}

func TestLoaderSCfgsClone(t *testing.T) {
	ban := LoaderSCfgs{{
		Enabled:        true,
		ID:             utils.MetaDefault,
		Tenant:         "cgrate.org",
		LockFilePath:   ".cgr.lck",
		CacheSConns:    []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches), "*conn1"},
		FieldSeparator: ",",
		TpInDir:        "/var/spool/cgrates/loader/in",
		TpOutDir:       "/var/spool/cgrates/loader/out",
		Data: []*LoaderDataType{{
			Type:     "*attributes",
			Filename: "Attributes.csv",
			Flags:    utils.FlagsWithParams{},
			Fields: []*FCTemplate{
				{
					Tag:       "TenantID",
					Path:      "Tenant",
					pathSlice: []string{"Tenant"},
					Type:      utils.MetaComposed,
					Value:     NewRSRParsersMustCompile("cgrate.org", utils.InfieldSep),
					Mandatory: true,
					Layout:    time.RFC3339,
				},
			}},
		},
	}}
	rcv := ban.Clone()
	if !reflect.DeepEqual(ban, rcv) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(ban), utils.ToJSON(rcv))
	}
	if rcv[0].CacheSConns[1] = ""; ban[0].CacheSConns[1] != "*conn1" {
		t.Errorf("Expected clone to not modify the cloned")
	}
	if rcv[0].Data[0].Type = ""; ban[0].Data[0].Type != "*attributes" {
		t.Errorf("Expected clone to not modify the cloned")
	}
}

func TestLockFolderRelativePath(t *testing.T) {
	ldr := &LoaderSCfg{
		TpInDir:      "/var/spool/cgrates/loader/in/",
		TpOutDir:     "/var/spool/cgrates/loader/out/",
		LockFilePath: utils.ResourcesCsv,
	}

	jsonCfg := &LoaderJsonCfg{
		ID:              utils.StringPointer("loaderid"),
		Enabled:         utils.BoolPointer(true),
		Tenant:          utils.StringPointer("cgrates.org"),
		Dry_run:         utils.BoolPointer(false),
		Lockfile_path:   utils.StringPointer(utils.ResourcesCsv),
		Field_separator: utils.StringPointer(utils.InfieldSep),
		Tp_in_dir:       utils.StringPointer("/var/spool/cgrates/loader/in/"),
		Tp_out_dir:      utils.StringPointer("/var/spool/cgrates/loader/out/"),
	}
	expPath := path.Join(ldr.LockFilePath)
	if err = ldr.loadFromJSONCfg(jsonCfg, map[string][]*FCTemplate{}, utils.InfieldSep); err != nil {
		t.Error(err)
	} else if ldr.LockFilePath != expPath {
		t.Errorf("Expected %v \n but received \n %v", expPath, ldr.LockFilePath)
	}
}
func TestLockFolderNonRelativePath(t *testing.T) {
	ldr := &LoaderSCfg{
		TpInDir:      "/var/spool/cgrates/loader/in/",
		TpOutDir:     "/var/spool/cgrates/loader/out/",
		LockFilePath: utils.ResourcesCsv,
	}

	jsonCfg := &LoaderJsonCfg{
		ID:              utils.StringPointer("loaderid"),
		Enabled:         utils.BoolPointer(true),
		Tenant:          utils.StringPointer("cgrates.org"),
		Dry_run:         utils.BoolPointer(false),
		Lockfile_path:   utils.StringPointer(path.Join("/tmp/", utils.ResourcesCsv)),
		Field_separator: utils.StringPointer(utils.InfieldSep),
		Tp_in_dir:       utils.StringPointer("/var/spool/cgrates/loader/in/"),
		Tp_out_dir:      utils.StringPointer("/var/spool/cgrates/loader/out/"),
	}
	expPath := path.Join("/tmp/", utils.ResourcesCsv)
	if err = ldr.loadFromJSONCfg(jsonCfg, map[string][]*FCTemplate{}, utils.InfieldSep); err != nil {
		t.Error(err)
	} else if ldr.LockFilePath != expPath {
		t.Errorf("Expected %v \n but received \n %v", expPath, ldr.LockFilePath)
	}
}

func TestLockFolderIsDir(t *testing.T) {
	ldr := &LoaderSCfg{
		LockFilePath: "test",
	}

	jsonCfg := &LoaderJsonCfg{
		ID:              utils.StringPointer("loaderid"),
		Enabled:         utils.BoolPointer(true),
		Tenant:          utils.StringPointer("cgrates.org"),
		Dry_run:         utils.BoolPointer(false),
		Lockfile_path:   utils.StringPointer("/tmp"),
		Field_separator: utils.StringPointer(utils.InfieldSep),
		Tp_in_dir:       utils.StringPointer("/var/spool/cgrates/loader/in/"),
		Tp_out_dir:      utils.StringPointer("/var/spool/cgrates/loader/out/"),
	}
	expPath := path.Join("/tmp")

	if err = ldr.loadFromJSONCfg(jsonCfg, map[string][]*FCTemplate{}, utils.InfieldSep); err != nil {
		t.Error(err)
	} else if ldr.LockFilePath != expPath {
		t.Errorf("Expected %v \n but received \n %v", expPath, ldr.LockFilePath)
	}
}

func TestLockGetLockFilePath(t *testing.T) {

	l := LoaderSCfg{
		LockFilePath: "folder",
		TpInDir:      "/dir",
		ID:           "4"}
	expected := "/dir/folder"
	if val := l.GetLockFilePath(); val != expected {
		t.Error(val)
	}

}
