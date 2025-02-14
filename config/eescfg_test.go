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
	"time"

	"github.com/cgrates/cgrates/utils"
)

func TestEESClone(t *testing.T) {
	cfgJSONStr := `{
  "ees": {
     "enabled": true,						
	"attributes_conns":["*internal", "*conn1"],					
	"cache": {
		"*fileCSV": {"limit": -2, "ttl": "3s", "static_ttl": true},
	},
	"exporters": [
		{
			"id": "cgrates",									
			"type": "*none",									
			"export_path": "/var/spool/cgrates/ees",			
			"opts": {
              "csvFieldSeparator": ";"
             },											
			"timezone": "local",										
			"filters": ["randomFiletrs"],										
			"flags": [],										
			"attribute_ids": ["randomID"],								
			"attribute_context": "",							
			"synchronous": false,								
			"attempts": 2,										
			"field_separator": ",",								
			"fields":[											
				{"tag": "*originID", "path": "*exp.*originID", "type": "*variable", "value": "~*opts.*originID"},
                {"tag": "*originID", "path": "*hdr.*originID", "type": "*variable", "value": "~*opts.*originID"},
                {"tag": "*originID", "path": "*trl.*originID", "type": "*variable", "value": "~*opts.*originID"},
                {"tag": "*originID", "path": "*uch.*originID", "type": "*variable", "value": "~*opts.*originID"},
			],
		},
	],
},
}`
	expected := &EEsCfg{
		Enabled:         true,
		AttributeSConns: []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes), "*conn1"},
		Cache: map[string]*CacheParamCfg{
			utils.MetaFileCSV: {
				Limit:     -2,
				TTL:       3 * time.Second,
				StaticTTL: true,
				Precache:  false,
				Replicate: false,
			},
		},
		Exporters: []*EventExporterCfg{
			{
				ID:             utils.MetaDefault,
				Type:           utils.MetaNone,
				Synchronous:    false,
				ExportPath:     "/var/spool/cgrates/ees",
				Attempts:       1,
				Timezone:       utils.EmptyString,
				AttributeSCtx:  utils.EmptyString,
				Filters:        []string{},
				AttributeSIDs:  []string{},
				Flags:          utils.FlagsWithParams{},
				Fields:         []*FCTemplate{},
				contentFields:  []*FCTemplate{},
				headerFields:   []*FCTemplate{},
				trailerFields:  []*FCTemplate{},
				Opts:           &EventExporterOpts{},
				FailedPostsDir: "/var/spool/cgrates/failed_posts",
				EFsConns:       []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaEFs)},
			},
			{
				ID:             utils.CGRateSLwr,
				Type:           utils.MetaNone,
				Synchronous:    false,
				ExportPath:     "/var/spool/cgrates/ees",
				Attempts:       2,
				Timezone:       "local",
				Filters:        []string{"randomFiletrs"},
				AttributeSIDs:  []string{"randomID"},
				Flags:          utils.FlagsWithParams{},
				FailedPostsDir: "/var/spool/cgrates/failed_posts",
				EFsConns:       []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaEFs)},
				Fields: []*FCTemplate{
					{
						Tag:    utils.MetaOriginID,
						Path:   "*exp.*originID",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*opts.*originID", utils.InfieldSep),
						Layout: time.RFC3339,
					},
					{
						Tag:    utils.MetaOriginID,
						Path:   "*hdr.*originID",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*opts.*originID", utils.InfieldSep),
						Layout: time.RFC3339,
					},
					{
						Tag:    utils.MetaOriginID,
						Path:   "*trl.*originID",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*opts.*originID", utils.InfieldSep),
						Layout: time.RFC3339,
					},
					{
						Tag:    utils.MetaOriginID,
						Path:   "*uch.*originID",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*opts.*originID", utils.InfieldSep),
						Layout: time.RFC3339,
					},
				},
				contentFields: []*FCTemplate{
					{
						Tag:    utils.MetaOriginID,
						Path:   "*exp.*originID",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*opts.*originID", utils.InfieldSep),
						Layout: time.RFC3339,
					},
					{
						Tag:    utils.MetaOriginID,
						Path:   "*uch.*originID",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*opts.*originID", utils.InfieldSep),
						Layout: time.RFC3339,
					},
				},
				headerFields: []*FCTemplate{
					{
						Tag:    utils.MetaOriginID,
						Path:   "*hdr.*originID",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*opts.*originID", utils.InfieldSep),
						Layout: time.RFC3339,
					},
				},
				trailerFields: []*FCTemplate{
					{
						Tag:    utils.MetaOriginID,
						Path:   "*trl.*originID",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*opts.*originID", utils.InfieldSep),
						Layout: time.RFC3339,
					},
				},
				Opts: &EventExporterOpts{
					CSVFieldSeparator: utils.StringPointer(utils.InfieldSep),
				},
			},
		},
	}
	for _, profile := range expected.Exporters {
		for _, v := range profile.Fields {
			v.ComputePath()
		}
		for _, v := range profile.ContentFields() {
			v.ComputePath()
		}
		for _, v := range profile.HeaderFields() {
			v.ComputePath()
		}
		for _, v := range profile.TrailerFields() {
			v.ComputePath()
		}
	}
	if jsonCfg, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else {
		cloneCfg := jsonCfg.eesCfg.Clone()
		if !reflect.DeepEqual(cloneCfg, expected) {
			t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(cloneCfg))
		}
	}
}

func TestEventExporterFieldloadFromJsonCfg(t *testing.T) {
	eventExporterJSON := &EEsJsonCfg{
		Exporters: &[]*EventExporterJsonCfg{
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
	if err := jsonCfg.eesCfg.loadFromJSONCfg(eventExporterJSON, jsonCfg.templates); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestEventExporterFieldloadFromJsonCfg1(t *testing.T) {
	eventExporterJSON := &EEsJsonCfg{
		Exporters: &[]*EventExporterJsonCfg{
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
	if err := jsonCfg.eesCfg.loadFromJSONCfg(eventExporterJSON, jsonCfg.templates); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestEventExporterloadFromJsonCfg(t *testing.T) {
	jsonCfg := NewDefaultCGRConfig()

	eventExporter := new(EventExporterCfg)
	if err := eventExporter.loadFromJSONCfg(nil, jsonCfg.templates); err != nil {
		t.Error(err)
	}
}

func TestEESCacheloadFromJsonCfg(t *testing.T) {
	eesCfg := &EEsJsonCfg{
		Cache: map[string]*CacheParamJsonCfg{
			utils.MetaFileCSV: {
				Ttl: utils.StringPointer("1ss"),
			},
		},
	}
	expected := "time: unknown unit \"ss\" in duration \"1ss\""
	jsonCfg := NewDefaultCGRConfig()
	if err := jsonCfg.eesCfg.loadFromJSONCfg(eesCfg, jsonCfg.templates); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestEventExporterSameID(t *testing.T) {
	expectedEEsCfg := &EEsCfg{
		Enabled:         true,
		AttributeSConns: []string{"conn1"},
		Cache: map[string]*CacheParamCfg{
			utils.MetaFileCSV: {
				Limit:     -1,
				TTL:       5 * time.Second,
				StaticTTL: false,
			},
		},
		Exporters: []*EventExporterCfg{
			{
				ID:             utils.MetaDefault,
				Type:           utils.MetaNone,
				ExportPath:     "/var/spool/cgrates/ees",
				Attempts:       1,
				Timezone:       utils.EmptyString,
				Filters:        []string{},
				AttributeSIDs:  []string{},
				Flags:          utils.FlagsWithParams{},
				Fields:         []*FCTemplate{},
				contentFields:  []*FCTemplate{},
				headerFields:   []*FCTemplate{},
				trailerFields:  []*FCTemplate{},
				Opts:           &EventExporterOpts{},
				EFsConns:       []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaEFs)},
				FailedPostsDir: "/var/spool/cgrates/failed_posts",
			},
			{
				ID:            "file_exporter1",
				Type:          utils.MetaFileCSV,
				Timezone:      utils.EmptyString,
				Filters:       []string{},
				AttributeSIDs: []string{},
				ExportPath:    "/var/spool/cgrates/ees",
				Attempts:      1,
				Flags:         utils.FlagsWithParams{},
				Fields: []*FCTemplate{
					{Tag: "CustomTag2", Path: "*exp.CustomPath2", Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("CustomValue2", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
				},
				contentFields: []*FCTemplate{
					{Tag: "CustomTag2", Path: "*exp.CustomPath2", Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("CustomValue2", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
				},
				headerFields:   []*FCTemplate{},
				trailerFields:  []*FCTemplate{},
				Opts:           &EventExporterOpts{},
				FailedPostsDir: "/var/spool/cgrates/failed_posts",
				EFsConns:       []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaEFs)},
			},
		},
	}
	for _, profile := range expectedEEsCfg.Exporters {
		for _, v := range profile.Fields {
			v.ComputePath()
		}
		for _, v := range profile.contentFields {
			v.ComputePath()
		}
	}
	cfgJSONStr := `{
"ees": {
	"enabled": true,
	"attributes_conns":["conn1"],
	"exporters": [
		{
			"id": "file_exporter1",
			"type": "*fileCSV",
			"fields":[
				{"tag": "CustomTag1", "path": "*exp.CustomPath1", "type": "*variable", "value": "CustomValue1", "mandatory": true},
			],
		},
		{
			"id": "file_exporter1",
			"type": "*fileCSV",
			"fields":[
				{"tag": "CustomTag2", "path": "*exp.CustomPath2", "type": "*variable", "value": "CustomValue2", "mandatory": true},
			],
		},
	],
}
}`
	if cfg, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedEEsCfg, cfg.eesCfg) {
		t.Errorf("Expected: %+v ,\n received: %+v", utils.ToJSON(expectedEEsCfg), utils.ToJSON(cfg.eesCfg))
	}
}

func TestEEsCfgloadFromJsonCfgCase1(t *testing.T) {
	jsonCfg := &EEsJsonCfg{
		Enabled:          utils.BoolPointer(true),
		Attributes_conns: &[]string{"*conn1", "*conn2"},
		Cache: map[string]*CacheParamJsonCfg{
			utils.MetaFileCSV: {
				Limit:      utils.IntPointer(-2),
				Ttl:        utils.StringPointer("1s"),
				Static_ttl: utils.BoolPointer(false),
			},
		},
		Exporters: &[]*EventExporterJsonCfg{
			{
				Id:               utils.StringPointer("CSVExporter"),
				Type:             utils.StringPointer("*fileCSV"),
				Filters:          &[]string{},
				Attribute_ids:    &[]string{},
				Flags:            &[]string{"*dryRun"},
				Export_path:      utils.StringPointer("/tmp/testCSV"),
				Timezone:         utils.StringPointer("UTC"),
				Synchronous:      utils.BoolPointer(true),
				Attempts:         utils.IntPointer(1),
				Failed_posts_dir: utils.StringPointer("/var/spool/cgrates/failed_posts"),
				Fields: &[]*FcTemplateJsonCfg{
					{
						Tag:   utils.StringPointer(utils.MetaOriginID),
						Path:  utils.StringPointer("*exp.*originID"),
						Type:  utils.StringPointer(utils.MetaVariable),
						Value: utils.StringPointer("~*opts.*originID"),
					},
				},
			},
		},
	}
	expectedCfg := &EEsCfg{
		Enabled:         true,
		AttributeSConns: []string{"*conn1", "*conn2"},
		Cache: map[string]*CacheParamCfg{
			utils.MetaFileCSV: {
				Limit:     -2,
				TTL:       time.Second,
				StaticTTL: false,
			},
		},
		Exporters: []*EventExporterCfg{
			{
				ID:             utils.MetaDefault,
				Type:           utils.MetaNone,
				ExportPath:     "/var/spool/cgrates/ees",
				Attempts:       1,
				Timezone:       utils.EmptyString,
				Filters:        []string{},
				AttributeSIDs:  []string{},
				Flags:          utils.FlagsWithParams{},
				contentFields:  []*FCTemplate{},
				Fields:         []*FCTemplate{},
				headerFields:   []*FCTemplate{},
				trailerFields:  []*FCTemplate{},
				Opts:           &EventExporterOpts{},
				FailedPostsDir: "/var/spool/cgrates/failed_posts",
				EFsConns:       []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaEFs)},
			},
			{
				ID:            "CSVExporter",
				Type:          "*fileCSV",
				Filters:       []string{},
				AttributeSIDs: []string{},
				Flags:         utils.FlagsWithParamsFromSlice([]string{utils.MetaDryRun}),
				ExportPath:    "/tmp/testCSV",
				Timezone:      "UTC",
				Synchronous:   true,
				Attempts:      1,
				headerFields:  []*FCTemplate{},
				trailerFields: []*FCTemplate{},
				contentFields: []*FCTemplate{
					{
						Tag:    utils.MetaOriginID,
						Path:   "*exp.*originID",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*opts.*originID", utils.InfieldSep),
						Layout: time.RFC3339,
					},
				},
				Opts: &EventExporterOpts{},
				Fields: []*FCTemplate{
					{Tag: utils.MetaOriginID, Path: "*exp.*originID", Type: utils.MetaVariable, Value: NewRSRParsersMustCompile("~*opts.*originID", utils.InfieldSep), Layout: time.RFC3339},
				},
				FailedPostsDir: "/var/spool/cgrates/failed_posts",
				EFsConns:       []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaEFs)},
			},
		},
	}
	for _, profile := range expectedCfg.Exporters {
		for _, v := range profile.Fields {
			v.ComputePath()
		}
		for _, v := range profile.contentFields {
			v.ComputePath()
		}
	}
	cgrCfg := NewDefaultCGRConfig()
	if err := cgrCfg.eesCfg.loadFromJSONCfg(nil, cgrCfg.templates); err != nil {
		t.Error(err)
	} else if err := cgrCfg.eesCfg.loadFromJSONCfg(jsonCfg, cgrCfg.templates); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedCfg, cgrCfg.eesCfg) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expectedCfg), utils.ToJSON(cgrCfg.eesCfg))
	}
}

func TestEEsCfgloadFromJsonCfgCase2(t *testing.T) {
	jsonCfg := &EEsJsonCfg{
		Enabled:          utils.BoolPointer(true),
		Attributes_conns: &[]string{"*conn1", "*conn2"},
		Cache: map[string]*CacheParamJsonCfg{
			utils.MetaFileCSV: {
				Limit:      utils.IntPointer(-2),
				Ttl:        utils.StringPointer("1s"),
				Static_ttl: utils.BoolPointer(false),
			},
		},
		Exporters: &[]*EventExporterJsonCfg{
			{
				Id:            utils.StringPointer("CSVExporter"),
				Type:          utils.StringPointer("*fileCSV"),
				Filters:       &[]string{},
				Attribute_ids: &[]string{},
				Flags:         &[]string{"*dryRun"},
				Export_path:   utils.StringPointer("/tmp/testCSV"),
				Timezone:      utils.StringPointer("UTC"),
				Synchronous:   utils.BoolPointer(true),
				Attempts:      utils.IntPointer(1),
				Fields: &[]*FcTemplateJsonCfg{
					{
						Tag:    utils.StringPointer(utils.AnswerTime),
						Path:   utils.StringPointer("*exp.AnswerTime"),
						Type:   utils.StringPointer(utils.MetaTemplate),
						Value:  utils.StringPointer("randomVal"),
						Layout: utils.StringPointer(time.RFC3339),
					},
					{
						Path:  utils.StringPointer("*opts.*originID"),
						Type:  utils.StringPointer(utils.MetaVariable),
						Value: utils.StringPointer("1"),
					},
				},
			},
		},
	}
	expectedCfg := &EEsCfg{
		Enabled:         true,
		AttributeSConns: []string{"*conn1", "*conn2"},
		Cache: map[string]*CacheParamCfg{
			utils.MetaFileCSV: {
				Limit:     -2,
				TTL:       time.Second,
				StaticTTL: false,
			},
		},
		Exporters: []*EventExporterCfg{
			{
				ID:             utils.MetaDefault,
				Type:           utils.MetaNone,
				ExportPath:     "/var/spool/cgrates/ees",
				Attempts:       1,
				Timezone:       utils.EmptyString,
				Filters:        []string{},
				AttributeSIDs:  []string{},
				Flags:          utils.FlagsWithParams{},
				contentFields:  []*FCTemplate{},
				Fields:         []*FCTemplate{},
				headerFields:   []*FCTemplate{},
				trailerFields:  []*FCTemplate{},
				Opts:           &EventExporterOpts{},
				FailedPostsDir: "/var/spool/cgrates/failed_posts",
				EFsConns:       []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaEFs)},
			},
			{
				ID:            "CSVExporter",
				Type:          "*fileCSV",
				Filters:       []string{},
				AttributeSIDs: []string{},
				Flags:         utils.FlagsWithParamsFromSlice([]string{utils.MetaDryRun}),
				ExportPath:    "/tmp/testCSV",
				Timezone:      "UTC",
				Synchronous:   true,
				Attempts:      1,
				EFsConns:      []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaEFs)},
				headerFields:  []*FCTemplate{},
				trailerFields: []*FCTemplate{},
				contentFields: []*FCTemplate{
					{
						Tag:    utils.MetaOriginID,
						Path:   "*exp.*originID",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*opts.*originID", utils.InfieldSep),
						Layout: time.RFC3339,
					},
				},
				FailedPostsDir: "/var/spool/cgrates/failed_posts",
				Opts:           &EventExporterOpts{},
				Fields: []*FCTemplate{
					{
						Tag:    utils.MetaOriginID,
						Path:   "*exp.*originID",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*opts.*originID", utils.InfieldSep),
						Layout: time.RFC3339,
					},
					{
						Tag:    "*opts.*originID",
						Path:   "*opts.*originID",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("1", utils.InfieldSep),
						Layout: time.RFC3339,
					},
				},
			},
		},
	}
	msgTemplates := map[string][]*FCTemplate{
		"randomVal": {
			{
				Tag:    utils.MetaOriginID,
				Path:   "*exp.*originID",
				Type:   utils.MetaVariable,
				Value:  NewRSRParsersMustCompile("~*opts.*originID", utils.InfieldSep),
				Layout: time.RFC3339,
			},
		},
	}
	for _, profile := range expectedCfg.Exporters {
		for _, v := range profile.Fields {
			v.ComputePath()
		}
		for _, v := range profile.contentFields {
			v.ComputePath()
		}
		for _, v := range msgTemplates["randomVal"] {
			v.ComputePath()
		}
	}
	jsnCfg := NewDefaultCGRConfig()
	if err := jsnCfg.eesCfg.loadFromJSONCfg(jsonCfg, msgTemplates); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedCfg, jsnCfg.eesCfg) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expectedCfg), utils.ToJSON(jsnCfg.eesCfg))
	}
}

func TestEEsCfgAsMapInterface(t *testing.T) {
	cfgJSONStr := `{
      "ees": {									
	        "enabled": true,						
            "attributes_conns":["*internal","*conn2"],					
            "cache": {
		          "*fileCSV": {"limit": -2, "precache": false, "replicate": false, "ttl": "1s", "static_ttl": false}
            },
            "exporters": [
            {
                  "id": "CSVExporter",									
			      "type": "*fileCSV",									
                  "export_path": "/tmp/testCSV",			
			      "opts": {
					"awsSecret": "test",
					"mysqlDSNParams": {
						"allowOldPasswords": "true",
						"allowNativePasswords": "true",
					},
				  },											
			      "timezone": "UTC",										
			      "filters": [],										
			      "flags": ["randomFlag"],										
			      "attribute_ids": [],								
			      "attribute_context": "",							
			      "synchronous": false,								
			      "attempts": 1,										
			      "field_separator": ",",								
			      "fields":[
                      {"tag": "*originID", "path": "*exp.*originID", "type": "*variable", "value": "~*opts.*originID"}
                  ]
            }]
	  }
    }`
	eMap := map[string]any{
		utils.EnabledCfg:         true,
		utils.AttributeSConnsCfg: []string{utils.MetaInternal, "*conn2"},
		utils.CacheCfg: map[string]any{
			utils.MetaFileCSV: map[string]any{
				utils.LimitCfg:     -2,
				utils.PrecacheCfg:  false,
				utils.RemoteCfg:    false,
				utils.ReplicateCfg: false,
				utils.TTLCfg:       "1s",
				utils.StaticTTLCfg: false,
			},
		},
		utils.ExportersCfg: []map[string]any{
			{
				utils.IDCfg:         "CSVExporter",
				utils.TypeCfg:       "*fileCSV",
				utils.ExportPathCfg: "/tmp/testCSV",
				utils.OptsCfg: map[string]any{
					utils.AWSSecret: "test",
					utils.MYSQLDSNParams: map[string]string{
						"allowOldPasswords":    "true",
						"allowNativePasswords": "true",
					},
				},
				utils.TimezoneCfg:           "UTC",
				utils.FiltersCfg:            []string{},
				utils.FlagsCfg:              []string{"randomFlag"},
				utils.AttributeIDsCfg:       []string{},
				utils.AttributeContextCfg:   utils.EmptyString,
				utils.SynchronousCfg:        false,
				utils.AttemptsCfg:           1,
				utils.ConcurrentRequestsCfg: 0,
				utils.BlockerCfg:            false,
				utils.FieldsCfg: []map[string]any{
					{
						utils.TagCfg:   utils.MetaOriginID,
						utils.PathCfg:  "*exp.*originID",
						utils.TypeCfg:  utils.MetaVariable,
						utils.ValueCfg: "~*opts.*originID",
					},
				},
				utils.FailedPostsDirCfg: "/var/spool/cgrates/failed_posts",
				utils.EFsConnsCfg:       []string{utils.MetaInternal},
			},
		},
	}
	if cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else {
		rcv := cgrCfg.eesCfg.AsMapInterface().(map[string]any)
		if len(rcv[utils.ExportersCfg].([]map[string]any)) != 2 {
			t.Errorf("Expected %+v, received %+v", 2, len(rcv[utils.ExportersCfg].([]map[string]any)))
		} else if !reflect.DeepEqual(eMap[utils.ExportersCfg].([]map[string]any)[0][utils.FieldsCfg].([]map[string]any)[0][utils.ValueCfg],
			rcv[utils.ExportersCfg].([]map[string]any)[1][utils.FieldsCfg].([]map[string]any)[0][utils.ValueCfg]) {
			t.Errorf("Expected %+v, received %+v", eMap[utils.ExportersCfg].([]map[string]any)[0][utils.FieldsCfg].([]map[string]any)[0][utils.ValueCfg],
				rcv[utils.ExportersCfg].([]map[string]any)[1][utils.FieldsCfg].([]map[string]any)[0][utils.ValueCfg])
		}
		rcv[utils.ExportersCfg].([]map[string]any)[1][utils.FieldsCfg] = nil
		eMap[utils.ExportersCfg].([]map[string]any)[0][utils.FieldsCfg] = nil
		if !reflect.DeepEqual(rcv[utils.ExportersCfg].([]map[string]any)[1],
			eMap[utils.ExportersCfg].([]map[string]any)[0]) {
			t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(eMap[utils.ExportersCfg].([]map[string]any)[0]),
				utils.ToJSON(rcv[utils.ExportersCfg].([]map[string]any)[1]))
		}
		rcv[utils.ExportersCfg] = nil
		eMap[utils.ExportersCfg] = nil
		if !reflect.DeepEqual(rcv, eMap) {
			t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
		}
	}
}

func TestDiffEventExporterJsonCfg(t *testing.T) {
	var d *EventExporterJsonCfg

	v1 := &EventExporterCfg{
		ID:         "EES_ID",
		Type:       "xml",
		ExportPath: "/tmp/ees",
		Opts:       &EventExporterOpts{},
		Timezone:   "UTC",
		Filters:    []string{"Filter1"},
		Flags: utils.FlagsWithParams{
			"FLAG_1": {
				"PARAM_1": []string{"param1"},
			},
		},
		AttributeSIDs:      []string{"ATTR_PRF"},
		AttributeSCtx:      "*sessions",
		Synchronous:        false,
		Attempts:           2,
		ConcurrentRequests: 3,
		FailedPostsDir:     "/tmp/failedPosts",
		Fields: []*FCTemplate{
			{
				Type: "*string",
			},
		},
		headerFields: []*FCTemplate{
			{
				Type: "*string",
			},
		},
		contentFields: []*FCTemplate{
			{
				Type: "*string",
			},
		},
		trailerFields: []*FCTemplate{
			{
				Type: "*string",
			},
		},
		Blocker:  true,
		EFsConns: []string{"v1 efs test"},
	}

	v2 := &EventExporterCfg{
		ID:         "EES_ID2",
		Type:       "http",
		ExportPath: "/var/tmp/ees",
		Opts: &EventExporterOpts{
			CSVFieldSeparator: utils.StringPointer(utils.InfieldSep),
		},

		Timezone: "EEST",
		Filters:  []string{"Filter2"},
		Flags: utils.FlagsWithParams{
			"FLAG_2": {
				"PARAM_2": []string{"param2"},
			},
		},
		AttributeSIDs:      []string{"ATTR_PRF_2"},
		AttributeSCtx:      "*actions",
		Synchronous:        true,
		Attempts:           3,
		ConcurrentRequests: 4,
		FailedPostsDir:     "/tmp/failed",
		Fields: []*FCTemplate{
			{
				Type: "*prefix",
			},
		},
		headerFields: []*FCTemplate{
			{
				Type: "*prefix",
			},
		},
		contentFields: []*FCTemplate{
			{
				Type: "*prefix",
			},
		},
		trailerFields: []*FCTemplate{
			{
				Type: "*prefix",
			},
		},
		Blocker:  false,
		EFsConns: []string{"efs test"},
	}

	expected := &EventExporterJsonCfg{
		Id:          utils.StringPointer("EES_ID2"),
		Type:        utils.StringPointer("http"),
		Export_path: utils.StringPointer("/var/tmp/ees"),
		Opts: &EventExporterOptsJson{
			CSVFieldSeparator: utils.StringPointer(utils.InfieldSep),
		},
		Timezone:            utils.StringPointer("EEST"),
		Filters:             &[]string{"Filter2"},
		Flags:               &[]string{"FLAG_2:PARAM_2:param2"},
		Attribute_ids:       &[]string{"ATTR_PRF_2"},
		Attribute_context:   utils.StringPointer("*actions"),
		Synchronous:         utils.BoolPointer(true),
		Attempts:            utils.IntPointer(3),
		Concurrent_requests: utils.IntPointer(4),
		Failed_posts_dir:    utils.StringPointer("/tmp/failed"),
		Fields: &[]*FcTemplateJsonCfg{
			{
				Type:   utils.StringPointer("*prefix"),
				Layout: utils.StringPointer(""),
			},
		},
		Blocker:   utils.BoolPointer(false),
		Efs_conns: utils.SliceStringPointer([]string{"efs test"}),
	}

	rcv := diffEventExporterJsonCfg(d, v1, v2)
	if !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}

	v1 = v2
	expected = &EventExporterJsonCfg{
		Opts: &EventExporterOptsJson{},
	}
	rcv = diffEventExporterJsonCfg(d, v1, v2)
	if !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}

	d = &EventExporterJsonCfg{
		Fields: &[]*FcTemplateJsonCfg{
			{
				Type: utils.StringPointer("*prefix"),
			},
		},
	}

	expected = &EventExporterJsonCfg{
		Opts: &EventExporterOptsJson{},
		Fields: &[]*FcTemplateJsonCfg{
			{
				Type: utils.StringPointer("*prefix"),
			},
		},
	}

	rcv = diffEventExporterJsonCfg(d, v1, v2)
	if !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}

func TestGetEventExporterJsonCfg(t *testing.T) {
	d := []*EventExporterJsonCfg{
		{
			Id: utils.StringPointer("EES_ID"),
		},
	}

	expected := &EventExporterJsonCfg{
		Id: utils.StringPointer("EES_ID"),
	}

	rcv, idx := getEventExporterJsonCfg(d, "EES_ID")
	if !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	} else if idx != 0 {
		t.Errorf("Expected %v \n but received \n %v", 0, idx)
	}

	d = []*EventExporterJsonCfg{
		{
			Id: nil,
		},
	}
	rcv, idx = getEventExporterJsonCfg(d, "EES_ID")
	if rcv != nil {
		t.Error("Received value should be null")
	} else if idx != -1 {
		t.Errorf("Expected %v \n but received \n %v", -1, idx)
	}
}

func TestGetEventExporterCfg(t *testing.T) {
	d := []*EventExporterCfg{
		{
			ID:   "EES_ID",
			Opts: &EventExporterOpts{},
		},
	}

	expected := &EventExporterCfg{
		ID:   "EES_ID",
		Opts: &EventExporterOpts{},
	}

	rcv := getEventExporterCfg(d, "EES_ID")
	if !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}

	d = []*EventExporterCfg{
		{
			ID:   "EES_ID2",
			Opts: &EventExporterOpts{},
		},
	}

	rcv = getEventExporterCfg(d, "EES_ID")
	expected = &EventExporterCfg{
		Opts: &EventExporterOpts{},
	}
	if !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}

func TestDiffEventExportersJsonCfg(t *testing.T) {
	var d *[]*EventExporterJsonCfg

	v1 := []*EventExporterCfg{
		{
			ID:         "EES_ID",
			Type:       "xml",
			ExportPath: "/tmp/ees",
			Opts:       &EventExporterOpts{},

			Timezone: "UTC",
			Filters:  []string{"Filter1"},
			Flags: utils.FlagsWithParams{
				"FLAG_1": {
					"PARAM_1": []string{"param1"},
				},
			},
			AttributeSIDs: []string{"ATTR_PRF"},
			AttributeSCtx: "*sessions",
			Synchronous:   false,
			Attempts:      2,
			Fields: []*FCTemplate{
				{
					Type: "*string",
				},
			},
			headerFields: []*FCTemplate{
				{
					Type: "*string",
				},
			},
			contentFields: []*FCTemplate{
				{
					Type: "*string",
				},
			},
			trailerFields: []*FCTemplate{
				{
					Type: "*string",
				},
			},
		},
	}

	v2 := []*EventExporterCfg{
		{
			ID:         "EES_ID2",
			Type:       "http",
			ExportPath: "/var/tmp/ees",
			Opts: &EventExporterOpts{
				CSVFieldSeparator: utils.StringPointer(utils.InfieldSep),
			},

			Timezone: "EEST",
			Filters:  []string{"Filter2"},
			Flags: utils.FlagsWithParams{
				"FLAG_2": {
					"PARAM_2": []string{"param2"},
				},
			},
			AttributeSIDs: []string{"ATTR_PRF_2"},
			AttributeSCtx: "*actions",
			Synchronous:   true,
			Attempts:      3,
			Fields: []*FCTemplate{
				{
					Type: "*prefix",
				},
			},
			headerFields: []*FCTemplate{
				{
					Type: "*prefix",
				},
			},
			contentFields: []*FCTemplate{
				{
					Type: "*prefix",
				},
			},
			trailerFields: []*FCTemplate{
				{
					Type: "*prefix",
				},
			},
		},
	}

	expected := &[]*EventExporterJsonCfg{
		{
			Id:          utils.StringPointer("EES_ID2"),
			Type:        utils.StringPointer("http"),
			Export_path: utils.StringPointer("/var/tmp/ees"),
			Opts: &EventExporterOptsJson{
				CSVFieldSeparator: utils.StringPointer(utils.InfieldSep),
			},
			Timezone:          utils.StringPointer("EEST"),
			Filters:           &[]string{"Filter2"},
			Flags:             &[]string{"FLAG_2:PARAM_2:param2"},
			Attribute_ids:     &[]string{"ATTR_PRF_2"},
			Attribute_context: utils.StringPointer("*actions"),
			Synchronous:       utils.BoolPointer(true),
			Attempts:          utils.IntPointer(3),
			Fields: &[]*FcTemplateJsonCfg{
				{
					Type:   utils.StringPointer("*prefix"),
					Layout: utils.StringPointer(""),
				},
			},
		},
	}

	rcv := diffEventExportersJsonCfg(d, v1, v2)
	if !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}

	v1 = v2
	expected = &[]*EventExporterJsonCfg{
		{
			Opts: &EventExporterOptsJson{},
		},
	}
	rcv = diffEventExportersJsonCfg(d, v1, v2)
	if !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}

	d = &[]*EventExporterJsonCfg{
		{
			Id: utils.StringPointer("EES_ID2"),
		},
	}

	expected = &[]*EventExporterJsonCfg{
		{
			Opts: &EventExporterOptsJson{},
			Id:   utils.StringPointer("EES_ID2"),
		},
	}

	rcv = diffEventExportersJsonCfg(d, v1, v2)
	if !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}

func TestDiffEEsJsonCfg(t *testing.T) {
	var d *EEsJsonCfg

	v1 := &EEsCfg{
		Enabled:         false,
		AttributeSConns: []string{"*localhost"},
		Cache:           map[string]*CacheParamCfg{},
		Exporters: []*EventExporterCfg{
			{
				Opts: &EventExporterOpts{},
			},
		},
	}

	v2 := &EEsCfg{
		Enabled:         true,
		AttributeSConns: []string{"*birpc"},
		Cache: map[string]*CacheParamCfg{
			"CACHE_1": {
				Limit: 1,
			},
		},
		Exporters: []*EventExporterCfg{
			{
				ID:   "EES_ID",
				Opts: &EventExporterOpts{},
			},
		},
	}

	expected := &EEsJsonCfg{
		Enabled:          utils.BoolPointer(true),
		Attributes_conns: &[]string{"*birpc"},
		Cache: map[string]*CacheParamJsonCfg{
			"CACHE_1": {
				Limit:      utils.IntPointer(1),
				Ttl:        utils.StringPointer("0s"),
				Static_ttl: utils.BoolPointer(false),
				Remote:     utils.BoolPointer(false),
				Precache:   utils.BoolPointer(false),
				Replicate:  utils.BoolPointer(false),
			},
		},
		Exporters: &[]*EventExporterJsonCfg{
			{
				Id:   utils.StringPointer("EES_ID"),
				Opts: &EventExporterOptsJson{},
			},
		},
	}

	rcv := diffEEsJsonCfg(d, v1, v2)
	if !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}

	v1 = v2
	expected = &EEsJsonCfg{
		Cache: map[string]*CacheParamJsonCfg{
			"CACHE_1": {
				Limit:      utils.IntPointer(1),
				Ttl:        utils.StringPointer("0s"),
				Static_ttl: utils.BoolPointer(false),
				Precache:   utils.BoolPointer(false),
				Remote:     utils.BoolPointer(false),
				Replicate:  utils.BoolPointer(false),
			},
		},
		Exporters: &[]*EventExporterJsonCfg{
			{
				Opts: &EventExporterOptsJson{},
			},
		},
	}
	rcv = diffEEsJsonCfg(d, v1, v2)
	if !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}

func TestEeSCloneSection(t *testing.T) {
	eeSCfg := &EEsCfg{
		Enabled:         true,
		AttributeSConns: []string{"*birpc"},
		Cache: map[string]*CacheParamCfg{
			"CACHE_1": {
				Limit: 1,
			},
		},
	}

	exp := &EEsCfg{
		Enabled:         true,
		AttributeSConns: []string{"*birpc"},
		Cache: map[string]*CacheParamCfg{
			"CACHE_1": {
				Limit: 1,
			},
		},
	}
	rcv := eeSCfg.CloneSection()
	rcv.(*EEsCfg).Exporters = nil
	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(rcv))
	}
}

func TestDiffEventExporterOptsJsonCfg(t *testing.T) {
	var d *EventExporterOptsJson

	v1 := &EventExporterOpts{
		ConnIDs: utils.SliceStringPointer([]string{"V1test"}),
	}

	v2 := &EventExporterOpts{
		CSVFieldSeparator:        utils.StringPointer(","),
		ElsIndex:                 utils.StringPointer("idx1"),
		ElsRefresh:               utils.StringPointer("true"),
		ElsOpType:                utils.StringPointer("op_type"),
		ElsPipeline:              utils.StringPointer("pipeline"),
		ElsRouting:               utils.StringPointer("routing"),
		ElsTimeout:               utils.DurationPointer(2 * time.Second),
		ElsWaitForActiveShards:   utils.StringPointer("wfas"),
		SQLMaxIdleConns:          utils.IntPointer(5),
		SQLMaxOpenConns:          utils.IntPointer(10),
		SQLConnMaxLifetime:       utils.DurationPointer(2 * time.Second),
		SQLTableName:             utils.StringPointer("cdrs"),
		SQLDBName:                utils.StringPointer("cgrates"),
		PgSSLMode:                utils.StringPointer("sslm"),
		KafkaTopic:               utils.StringPointer("topic1"),
		KafkaCAPath:              utils.StringPointer("kafkaCAPath"),
		KafkaSkipTLSVerify:       utils.BoolPointer(false),
		AMQPRoutingKey:           utils.StringPointer("routing_key"),
		AMQPQueueID:              utils.StringPointer("queue_id"),
		AMQPExchange:             utils.StringPointer("amqp_exchange"),
		AMQPExchangeType:         utils.StringPointer("amqp_exchange_type"),
		AWSRegion:                utils.StringPointer("utc"),
		AWSKey:                   utils.StringPointer("aws_key"),
		AWSSecret:                utils.StringPointer("aws_secret"),
		AWSToken:                 utils.StringPointer("aws_token"),
		SQSQueueID:               utils.StringPointer("sqs_queue_id"),
		S3BucketID:               utils.StringPointer("s3_bucket_id"),
		S3FolderPath:             utils.StringPointer("s3_folder_path"),
		NATSJetStream:            utils.BoolPointer(false),
		NATSSubject:              utils.StringPointer("ees_nats"),
		NATSJWTFile:              utils.StringPointer("/path/to/jwt"),
		NATSSeedFile:             utils.StringPointer("/path/to/seed"),
		NATSCertificateAuthority: utils.StringPointer("ca"),
		NATSClientCertificate:    utils.StringPointer("cc"),
		NATSClientKey:            utils.StringPointer("ck"),
		NATSJetStreamMaxWait:     utils.DurationPointer(2 * time.Second),
		RPCCodec:                 utils.StringPointer("rpccodec"),
		ServiceMethod:            utils.StringPointer("service_method"),
		KeyPath:                  utils.StringPointer("/path/to/key"),
		CertPath:                 utils.StringPointer("cp"),
		CAPath:                   utils.StringPointer("ca_path"),
		TLS:                      utils.BoolPointer(false),
		RPCConnTimeout:           utils.DurationPointer(2 * time.Second),
		RPCReplyTimeout:          utils.DurationPointer(2 * time.Second),
		MYSQLDSNParams:           map[string]string{},
		KafkaTLS:                 utils.BoolPointer(true),
		ConnIDs:                  utils.SliceStringPointer([]string{"test"}),
	}

	exp := &EventExporterOptsJson{
		CSVFieldSeparator:        utils.StringPointer(","),
		ElsIndex:                 utils.StringPointer("idx1"),
		ElsRefresh:               utils.StringPointer("true"),
		ElsOpType:                utils.StringPointer("op_type"),
		ElsPipeline:              utils.StringPointer("pipeline"),
		ElsRouting:               utils.StringPointer("routing"),
		ElsTimeout:               utils.StringPointer("2s"),
		ElsWaitForActiveShards:   utils.StringPointer("wfas"),
		SQLMaxIdleConns:          utils.IntPointer(5),
		SQLMaxOpenConns:          utils.IntPointer(10),
		SQLConnMaxLifetime:       utils.StringPointer("2s"),
		SQLTableName:             utils.StringPointer("cdrs"),
		SQLDBName:                utils.StringPointer("cgrates"),
		PgSSLMode:                utils.StringPointer("sslm"),
		KafkaTopic:               utils.StringPointer("topic1"),
		KafkaCAPath:              utils.StringPointer("kafkaCAPath"),
		KafkaSkipTLSVerify:       utils.BoolPointer(false),
		AMQPRoutingKey:           utils.StringPointer("routing_key"),
		AMQPQueueID:              utils.StringPointer("queue_id"),
		AMQPExchange:             utils.StringPointer("amqp_exchange"),
		AMQPExchangeType:         utils.StringPointer("amqp_exchange_type"),
		AWSRegion:                utils.StringPointer("utc"),
		AWSKey:                   utils.StringPointer("aws_key"),
		AWSSecret:                utils.StringPointer("aws_secret"),
		AWSToken:                 utils.StringPointer("aws_token"),
		SQSQueueID:               utils.StringPointer("sqs_queue_id"),
		S3BucketID:               utils.StringPointer("s3_bucket_id"),
		S3FolderPath:             utils.StringPointer("s3_folder_path"),
		NATSJetStream:            utils.BoolPointer(false),
		NATSSubject:              utils.StringPointer("ees_nats"),
		NATSJWTFile:              utils.StringPointer("/path/to/jwt"),
		NATSSeedFile:             utils.StringPointer("/path/to/seed"),
		NATSCertificateAuthority: utils.StringPointer("ca"),
		NATSClientCertificate:    utils.StringPointer("cc"),
		NATSClientKey:            utils.StringPointer("ck"),
		NATSJetStreamMaxWait:     utils.StringPointer("2s"),
		RPCCodec:                 utils.StringPointer("rpccodec"),
		ServiceMethod:            utils.StringPointer("service_method"),
		KeyPath:                  utils.StringPointer("/path/to/key"),
		CertPath:                 utils.StringPointer("cp"),
		CAPath:                   utils.StringPointer("ca_path"),
		TLS:                      utils.BoolPointer(false),
		RPCConnTimeout:           utils.StringPointer("2s"),
		RPCReplyTimeout:          utils.StringPointer("2s"),
		MYSQLDSNParams:           map[string]string{},
		KafkaTLS:                 utils.BoolPointer(true),
		ConnIDs:                  utils.SliceStringPointer([]string{"test"}),
	}

	rcv := diffEventExporterOptsJsonCfg(d, v1, v2)
	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(rcv))
	}
}

func TestDiffEventExporterOptsJsonCfgConnIDsAreEqual(t *testing.T) {
	var d *EventExporterOptsJson

	v1 := &EventExporterOpts{
		ConnIDs: utils.SliceStringPointer([]string{"test"}),
	}

	v2 := &EventExporterOpts{
		ConnIDs: utils.SliceStringPointer([]string{"test"}),
	}

	exp := &EventExporterOptsJson{}

	rcv := diffEventExporterOptsJsonCfg(d, v1, v2)
	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(rcv))
	}
}

func TestEventExporterOptsClone(t *testing.T) {
	eeOpts := &EventExporterOpts{
		CSVFieldSeparator:        utils.StringPointer(","),
		ElsIndex:                 utils.StringPointer("idx1"),
		ElsRefresh:               utils.StringPointer("true"),
		ElsOpType:                utils.StringPointer("op_type"),
		ElsPipeline:              utils.StringPointer("pipeline"),
		ElsRouting:               utils.StringPointer("routing"),
		ElsTimeout:               utils.DurationPointer(2 * time.Second),
		ElsWaitForActiveShards:   utils.StringPointer("wfas"),
		SQLMaxIdleConns:          utils.IntPointer(5),
		SQLMaxOpenConns:          utils.IntPointer(10),
		SQLConnMaxLifetime:       utils.DurationPointer(2 * time.Second),
		SQLTableName:             utils.StringPointer("cdrs"),
		SQLDBName:                utils.StringPointer("cgrates"),
		PgSSLMode:                utils.StringPointer("sslm"),
		KafkaTopic:               utils.StringPointer("topic1"),
		KafkaBatchSize:           utils.IntPointer(50),
		KafkaCAPath:              utils.StringPointer("kafkaCAPath"),
		KafkaSkipTLSVerify:       utils.BoolPointer(false),
		AMQPRoutingKey:           utils.StringPointer("routing_key"),
		AMQPQueueID:              utils.StringPointer("queue_id"),
		AMQPExchange:             utils.StringPointer("amqp_exchange"),
		AMQPExchangeType:         utils.StringPointer("amqp_exchange_type"),
		AWSRegion:                utils.StringPointer("utc"),
		AWSKey:                   utils.StringPointer("aws_key"),
		AWSSecret:                utils.StringPointer("aws_secret"),
		AWSToken:                 utils.StringPointer("aws_token"),
		SQSQueueID:               utils.StringPointer("sqs_queue_id"),
		S3BucketID:               utils.StringPointer("s3_bucket_id"),
		S3FolderPath:             utils.StringPointer("s3_folder_path"),
		NATSJetStream:            utils.BoolPointer(false),
		NATSSubject:              utils.StringPointer("ees_nats"),
		NATSJWTFile:              utils.StringPointer("/path/to/jwt"),
		NATSSeedFile:             utils.StringPointer("/path/to/seed"),
		NATSCertificateAuthority: utils.StringPointer("ca"),
		NATSClientCertificate:    utils.StringPointer("cc"),
		NATSClientKey:            utils.StringPointer("ck"),
		NATSJetStreamMaxWait:     utils.DurationPointer(2 * time.Second),
		RPCCodec:                 utils.StringPointer("rpccodec"),
		ServiceMethod:            utils.StringPointer("service_method"),
		KeyPath:                  utils.StringPointer("/path/to/key"),
		CertPath:                 utils.StringPointer("cp"),
		CAPath:                   utils.StringPointer("ca_path"),
		TLS:                      utils.BoolPointer(false),
		RPCConnTimeout:           utils.DurationPointer(2 * time.Second),
		RPCReplyTimeout:          utils.DurationPointer(2 * time.Second),
		MYSQLDSNParams:           make(map[string]string),
		KafkaTLS:                 utils.BoolPointer(false),
		ConnIDs:                  utils.SliceStringPointer([]string{"testID"}),
		RPCAPIOpts:               make(map[string]any),
	}

	exp := &EventExporterOpts{
		CSVFieldSeparator:        utils.StringPointer(","),
		ElsIndex:                 utils.StringPointer("idx1"),
		ElsRefresh:               utils.StringPointer("true"),
		ElsOpType:                utils.StringPointer("op_type"),
		ElsPipeline:              utils.StringPointer("pipeline"),
		ElsRouting:               utils.StringPointer("routing"),
		ElsTimeout:               utils.DurationPointer(2 * time.Second),
		ElsWaitForActiveShards:   utils.StringPointer("wfas"),
		SQLMaxIdleConns:          utils.IntPointer(5),
		SQLMaxOpenConns:          utils.IntPointer(10),
		SQLConnMaxLifetime:       utils.DurationPointer(2 * time.Second),
		SQLTableName:             utils.StringPointer("cdrs"),
		SQLDBName:                utils.StringPointer("cgrates"),
		PgSSLMode:                utils.StringPointer("sslm"),
		KafkaTopic:               utils.StringPointer("topic1"),
		KafkaBatchSize:           utils.IntPointer(50),
		KafkaCAPath:              utils.StringPointer("kafkaCAPath"),
		KafkaSkipTLSVerify:       utils.BoolPointer(false),
		AMQPRoutingKey:           utils.StringPointer("routing_key"),
		AMQPQueueID:              utils.StringPointer("queue_id"),
		AMQPExchange:             utils.StringPointer("amqp_exchange"),
		AMQPExchangeType:         utils.StringPointer("amqp_exchange_type"),
		AWSRegion:                utils.StringPointer("utc"),
		AWSKey:                   utils.StringPointer("aws_key"),
		AWSSecret:                utils.StringPointer("aws_secret"),
		AWSToken:                 utils.StringPointer("aws_token"),
		SQSQueueID:               utils.StringPointer("sqs_queue_id"),
		S3BucketID:               utils.StringPointer("s3_bucket_id"),
		S3FolderPath:             utils.StringPointer("s3_folder_path"),
		NATSJetStream:            utils.BoolPointer(false),
		NATSSubject:              utils.StringPointer("ees_nats"),
		NATSJWTFile:              utils.StringPointer("/path/to/jwt"),
		NATSSeedFile:             utils.StringPointer("/path/to/seed"),
		NATSCertificateAuthority: utils.StringPointer("ca"),
		NATSClientCertificate:    utils.StringPointer("cc"),
		NATSClientKey:            utils.StringPointer("ck"),
		NATSJetStreamMaxWait:     utils.DurationPointer(2 * time.Second),
		RPCCodec:                 utils.StringPointer("rpccodec"),
		ServiceMethod:            utils.StringPointer("service_method"),
		KeyPath:                  utils.StringPointer("/path/to/key"),
		CertPath:                 utils.StringPointer("cp"),
		CAPath:                   utils.StringPointer("ca_path"),
		TLS:                      utils.BoolPointer(false),
		RPCConnTimeout:           utils.DurationPointer(2 * time.Second),
		RPCReplyTimeout:          utils.DurationPointer(2 * time.Second),
		MYSQLDSNParams:           make(map[string]string),
		KafkaTLS:                 utils.BoolPointer(false),
		ConnIDs:                  utils.SliceStringPointer([]string{"testID"}),
		RPCAPIOpts:               make(map[string]any),
	}

	if rcv := eeOpts.Clone(); !reflect.DeepEqual(exp, rcv) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(rcv))
	}
}

func TestLoadFromJSONCfg(t *testing.T) {
	eeOpts := &EventExporterOpts{}

	eeSJson := &EventExporterOptsJson{
		CSVFieldSeparator:        utils.StringPointer(","),
		ElsIndex:                 utils.StringPointer("idx1"),
		ElsRefresh:               utils.StringPointer("true"),
		ElsOpType:                utils.StringPointer("op_type"),
		ElsPipeline:              utils.StringPointer("pipeline"),
		ElsRouting:               utils.StringPointer("routing"),
		ElsTimeout:               utils.StringPointer("2s"),
		ElsWaitForActiveShards:   utils.StringPointer("wfas"),
		SQLMaxIdleConns:          utils.IntPointer(5),
		SQLMaxOpenConns:          utils.IntPointer(10),
		SQLConnMaxLifetime:       utils.StringPointer("2s"),
		SQLTableName:             utils.StringPointer("cdrs"),
		SQLDBName:                utils.StringPointer("cgrates"),
		PgSSLMode:                utils.StringPointer("sslm"),
		KafkaTopic:               utils.StringPointer("topic1"),
		KafkaBatchSize:           utils.IntPointer(50),
		KafkaCAPath:              utils.StringPointer("kafkaCAPath"),
		KafkaSkipTLSVerify:       utils.BoolPointer(false),
		AMQPRoutingKey:           utils.StringPointer("routing_key"),
		AMQPQueueID:              utils.StringPointer("queue_id"),
		AMQPExchange:             utils.StringPointer("amqp_exchange"),
		AMQPExchangeType:         utils.StringPointer("amqp_exchange_type"),
		AWSRegion:                utils.StringPointer("utc"),
		AWSKey:                   utils.StringPointer("aws_key"),
		AWSSecret:                utils.StringPointer("aws_secret"),
		AWSToken:                 utils.StringPointer("aws_token"),
		SQSQueueID:               utils.StringPointer("sqs_queue_id"),
		S3BucketID:               utils.StringPointer("s3_bucket_id"),
		S3FolderPath:             utils.StringPointer("s3_folder_path"),
		NATSJetStream:            utils.BoolPointer(false),
		NATSSubject:              utils.StringPointer("ees_nats"),
		NATSJWTFile:              utils.StringPointer("/path/to/jwt"),
		NATSSeedFile:             utils.StringPointer("/path/to/seed"),
		NATSCertificateAuthority: utils.StringPointer("ca"),
		NATSClientCertificate:    utils.StringPointer("cc"),
		NATSClientKey:            utils.StringPointer("ck"),
		NATSJetStreamMaxWait:     utils.StringPointer("2s"),
		RPCCodec:                 utils.StringPointer("rpccodec"),
		ServiceMethod:            utils.StringPointer("service_method"),
		KeyPath:                  utils.StringPointer("/path/to/key"),
		CertPath:                 utils.StringPointer("cp"),
		CAPath:                   utils.StringPointer("ca_path"),
		TLS:                      utils.BoolPointer(false),
		RPCConnTimeout:           utils.StringPointer("2s"),
		RPCReplyTimeout:          utils.StringPointer("2s"),
		KafkaTLS:                 utils.BoolPointer(false),
		ConnIDs:                  utils.SliceStringPointer([]string{"testID"}),
		RPCAPIOpts:               make(map[string]any),
	}

	exp := &EventExporterOpts{
		CSVFieldSeparator:        utils.StringPointer(","),
		ElsIndex:                 utils.StringPointer("idx1"),
		ElsRefresh:               utils.StringPointer("true"),
		ElsOpType:                utils.StringPointer("op_type"),
		ElsPipeline:              utils.StringPointer("pipeline"),
		ElsRouting:               utils.StringPointer("routing"),
		ElsTimeout:               utils.DurationPointer(2 * time.Second),
		ElsWaitForActiveShards:   utils.StringPointer("wfas"),
		SQLMaxIdleConns:          utils.IntPointer(5),
		SQLMaxOpenConns:          utils.IntPointer(10),
		SQLConnMaxLifetime:       utils.DurationPointer(2 * time.Second),
		SQLTableName:             utils.StringPointer("cdrs"),
		SQLDBName:                utils.StringPointer("cgrates"),
		PgSSLMode:                utils.StringPointer("sslm"),
		KafkaTopic:               utils.StringPointer("topic1"),
		KafkaBatchSize:           utils.IntPointer(50),
		KafkaCAPath:              utils.StringPointer("kafkaCAPath"),
		KafkaSkipTLSVerify:       utils.BoolPointer(false),
		AMQPRoutingKey:           utils.StringPointer("routing_key"),
		AMQPQueueID:              utils.StringPointer("queue_id"),
		AMQPExchange:             utils.StringPointer("amqp_exchange"),
		AMQPExchangeType:         utils.StringPointer("amqp_exchange_type"),
		AWSRegion:                utils.StringPointer("utc"),
		AWSKey:                   utils.StringPointer("aws_key"),
		AWSSecret:                utils.StringPointer("aws_secret"),
		AWSToken:                 utils.StringPointer("aws_token"),
		SQSQueueID:               utils.StringPointer("sqs_queue_id"),
		S3BucketID:               utils.StringPointer("s3_bucket_id"),
		S3FolderPath:             utils.StringPointer("s3_folder_path"),
		NATSJetStream:            utils.BoolPointer(false),
		NATSSubject:              utils.StringPointer("ees_nats"),
		NATSJWTFile:              utils.StringPointer("/path/to/jwt"),
		NATSSeedFile:             utils.StringPointer("/path/to/seed"),
		NATSCertificateAuthority: utils.StringPointer("ca"),
		NATSClientCertificate:    utils.StringPointer("cc"),
		NATSClientKey:            utils.StringPointer("ck"),
		NATSJetStreamMaxWait:     utils.DurationPointer(2 * time.Second),
		RPCCodec:                 utils.StringPointer("rpccodec"),
		ServiceMethod:            utils.StringPointer("service_method"),
		KeyPath:                  utils.StringPointer("/path/to/key"),
		CertPath:                 utils.StringPointer("cp"),
		CAPath:                   utils.StringPointer("ca_path"),
		TLS:                      utils.BoolPointer(false),
		RPCConnTimeout:           utils.DurationPointer(2 * time.Second),
		RPCReplyTimeout:          utils.DurationPointer(2 * time.Second),
		KafkaTLS:                 utils.BoolPointer(false),
		ConnIDs:                  utils.SliceStringPointer([]string{"testID"}),
		RPCAPIOpts:               make(map[string]any),
	}

	if err := eeOpts.loadFromJSONCfg(eeSJson); err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(exp, eeOpts) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(eeOpts))
	}

	//check with empty json config
	eeSJson = nil
	if err := eeOpts.loadFromJSONCfg(eeSJson); err != nil {
		t.Error(err)
	}
}

func TestLoadFromJsonParseErrors(t *testing.T) {
	eeOpts := &EventExporterOpts{}

	eeSJson := &EventExporterOptsJson{
		CSVFieldSeparator:        utils.StringPointer(","),
		ElsIndex:                 utils.StringPointer("idx1"),
		ElsOpType:                utils.StringPointer("op_type"),
		ElsPipeline:              utils.StringPointer("pipeline"),
		ElsRouting:               utils.StringPointer("routing"),
		ElsTimeout:               utils.StringPointer("2c"),
		ElsWaitForActiveShards:   utils.StringPointer("wfas"),
		SQLMaxIdleConns:          utils.IntPointer(5),
		SQLMaxOpenConns:          utils.IntPointer(10),
		SQLConnMaxLifetime:       utils.StringPointer("2s"),
		SQLTableName:             utils.StringPointer("cdrs"),
		SQLDBName:                utils.StringPointer("cgrates"),
		PgSSLMode:                utils.StringPointer("sslm"),
		KafkaTopic:               utils.StringPointer("topic1"),
		AMQPRoutingKey:           utils.StringPointer("routing_key"),
		AMQPQueueID:              utils.StringPointer("queue_id"),
		AMQPExchange:             utils.StringPointer("amqp_exchange"),
		AMQPExchangeType:         utils.StringPointer("amqp_exchange_type"),
		AWSRegion:                utils.StringPointer("utc"),
		AWSKey:                   utils.StringPointer("aws_key"),
		AWSSecret:                utils.StringPointer("aws_secret"),
		AWSToken:                 utils.StringPointer("aws_token"),
		SQSQueueID:               utils.StringPointer("sqs_queue_id"),
		S3BucketID:               utils.StringPointer("s3_bucket_id"),
		S3FolderPath:             utils.StringPointer("s3_folder_path"),
		NATSJetStream:            utils.BoolPointer(false),
		NATSSubject:              utils.StringPointer("ees_nats"),
		NATSJWTFile:              utils.StringPointer("/path/to/jwt"),
		NATSSeedFile:             utils.StringPointer("/path/to/seed"),
		NATSCertificateAuthority: utils.StringPointer("ca"),
		NATSClientCertificate:    utils.StringPointer("cc"),
		NATSClientKey:            utils.StringPointer("ck"),
		NATSJetStreamMaxWait:     utils.StringPointer("2s"),
		RPCCodec:                 utils.StringPointer("rpccodec"),
		ServiceMethod:            utils.StringPointer("service_method"),
		KeyPath:                  utils.StringPointer("/path/to/key"),
		CertPath:                 utils.StringPointer("cp"),
		CAPath:                   utils.StringPointer("ca_path"),
		TLS:                      utils.BoolPointer(false),
		RPCConnTimeout:           utils.StringPointer("2s"),
		RPCReplyTimeout:          utils.StringPointer("2s"),
	}

	errExp := `time: unknown unit "c" in duration "2c"`
	if err := eeOpts.loadFromJSONCfg(eeSJson); err == nil || err.Error() != errExp {
		t.Errorf("Expected %v \n but received \n %v", errExp, err.Error())
	}
	eeSJson.ElsTimeout = utils.StringPointer("2s")

	///////

	eeSJson.SQLConnMaxLifetime = utils.StringPointer("2c")
	if err := eeOpts.loadFromJSONCfg(eeSJson); err == nil || err.Error() != errExp {
		t.Errorf("Expected %v \n but received \n %v", errExp, err.Error())
	}
	eeSJson.SQLConnMaxLifetime = utils.StringPointer("2s")

	//////

	eeSJson.NATSJetStreamMaxWait = utils.StringPointer("2c")
	if err := eeOpts.loadFromJSONCfg(eeSJson); err == nil || err.Error() != errExp {
		t.Errorf("Expected %v \n but received \n %v", errExp, err.Error())
	}
	eeSJson.NATSJetStreamMaxWait = utils.StringPointer("2s")

	/////

	eeSJson.RPCConnTimeout = utils.StringPointer("2c")
	if err := eeOpts.loadFromJSONCfg(eeSJson); err == nil || err.Error() != errExp {
		t.Errorf("Expected %v \n but received \n %v", errExp, err.Error())
	}
	eeSJson.RPCConnTimeout = utils.StringPointer("2s")

	/////

	eeSJson.RPCReplyTimeout = utils.StringPointer("2c")
	if err := eeOpts.loadFromJSONCfg(eeSJson); err == nil || err.Error() != errExp {
		t.Errorf("Expected %v \n but received \n %v", errExp, err.Error())
	}
	eeSJson.RPCReplyTimeout = utils.StringPointer("2s")
}

func TestEEsAsMapInterface(t *testing.T) {
	eeCfg := &EventExporterCfg{
		Opts: &EventExporterOpts{
			CSVFieldSeparator:        utils.StringPointer(","),
			ElsIndex:                 utils.StringPointer("idx1"),
			ElsRefresh:               utils.StringPointer("true"),
			ElsOpType:                utils.StringPointer("op_type"),
			ElsPipeline:              utils.StringPointer("pipeline"),
			ElsRouting:               utils.StringPointer("routing"),
			ElsTimeout:               utils.DurationPointer(2 * time.Second),
			ElsWaitForActiveShards:   utils.StringPointer("wfas"),
			SQLMaxIdleConns:          utils.IntPointer(5),
			SQLMaxOpenConns:          utils.IntPointer(10),
			SQLConnMaxLifetime:       utils.DurationPointer(2 * time.Second),
			SQLTableName:             utils.StringPointer("cdrs"),
			SQLDBName:                utils.StringPointer("cgrates"),
			PgSSLMode:                utils.StringPointer("sslm"),
			KafkaTopic:               utils.StringPointer("topic1"),
			KafkaBatchSize:           utils.IntPointer(50),
			KafkaCAPath:              utils.StringPointer("kafkaCAPath"),
			KafkaSkipTLSVerify:       utils.BoolPointer(false),
			AMQPRoutingKey:           utils.StringPointer("routing_key"),
			AMQPQueueID:              utils.StringPointer("queue_id"),
			AMQPExchange:             utils.StringPointer("amqp_exchange"),
			AMQPExchangeType:         utils.StringPointer("amqp_exchange_type"),
			AWSRegion:                utils.StringPointer("utc"),
			AWSKey:                   utils.StringPointer("aws_key"),
			AWSSecret:                utils.StringPointer("aws_secret"),
			AWSToken:                 utils.StringPointer("aws_token"),
			SQSQueueID:               utils.StringPointer("sqs_queue_id"),
			S3BucketID:               utils.StringPointer("s3_bucket_id"),
			S3FolderPath:             utils.StringPointer("s3_folder_path"),
			NATSJetStream:            utils.BoolPointer(false),
			NATSSubject:              utils.StringPointer("ees_nats"),
			NATSJWTFile:              utils.StringPointer("/path/to/jwt"),
			NATSSeedFile:             utils.StringPointer("/path/to/seed"),
			NATSCertificateAuthority: utils.StringPointer("ca"),
			NATSClientCertificate:    utils.StringPointer("cc"),
			NATSClientKey:            utils.StringPointer("ck"),
			NATSJetStreamMaxWait:     utils.DurationPointer(2 * time.Second),
			RPCCodec:                 utils.StringPointer("rpccodec"),
			ServiceMethod:            utils.StringPointer("service_method"),
			KeyPath:                  utils.StringPointer("/path/to/key"),
			CertPath:                 utils.StringPointer("cp"),
			CAPath:                   utils.StringPointer("ca_path"),
			TLS:                      utils.BoolPointer(false),
			RPCConnTimeout:           utils.DurationPointer(2 * time.Second),
			RPCReplyTimeout:          utils.DurationPointer(2 * time.Second),
			KafkaTLS:                 utils.BoolPointer(false),
			ConnIDs:                  utils.SliceStringPointer([]string{"testID"}),
			RPCAPIOpts:               make(map[string]any),
		},
	}

	exp := map[string]any{
		"opts": map[string]any{
			"tls":                      false,
			"amqpExchange":             "amqp_exchange",
			"amqpExchangeType":         "amqp_exchange_type",
			"amqpQueueID":              "queue_id",
			"amqpRoutingKey":           "routing_key",
			"awsKey":                   "aws_key",
			"awsRegion":                "utc",
			"awsSecret":                "aws_secret",
			"awsToken":                 "aws_token",
			"caPath":                   "ca_path",
			"certPath":                 "cp",
			"csvFieldSeparator":        ",",
			"elsIndex":                 "idx1",
			"elsRefresh":               "true",
			"elsOpType":                "op_type",
			"elsPipeline":              "pipeline",
			"elsRouting":               "routing",
			"elsTimeout":               "2s",
			"elsWaitForActiveShards":   "wfas",
			"kafkaTopic":               "topic1",
			utils.KafkaBatchSize:       50,
			"kafkaCAPath":              "kafkaCAPath",
			"kafkaSkipTLSVerify":       false,
			"keyPath":                  "/path/to/key",
			"natsCertificateAuthority": "ca",
			"natsClientCertificate":    "cc",
			"natsClientKey":            "ck",
			"natsJWTFile":              "/path/to/jwt",
			"natsJetStream":            false,
			"natsJetStreamMaxWait":     "2s",
			"natsSeedFile":             "/path/to/seed",
			"natsSubject":              "ees_nats",
			"rpcCodec":                 "rpccodec",
			"rpcConnTimeout":           "2s",
			"rpcReplyTimeout":          "2s",
			"s3BucketID":               "s3_bucket_id",
			"s3FolderPath":             "s3_folder_path",
			"serviceMethod":            "service_method",
			"sqlConnMaxLifetime":       "2s",
			"sqlDBName":                "cgrates",
			"sqlMaxIdleConns":          5,
			"sqlMaxOpenConns":          10,
			"sqlTableName":             "cdrs",
			"sqsQueueID":               "sqs_queue_id",
			"pgSSLMode":                "sslm",
			"kafkaTLS":                 false,
			"connIDs":                  []string{"testID"},
			"rpcAPIOpts":               make(map[string]any),
		},
	}

	rcv := eeCfg.AsMapInterface()

	if !reflect.DeepEqual(exp[utils.OptsCfg], rcv[utils.OptsCfg]) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp["opts"]), utils.ToJSON(rcv["opts"]))
	}
}

func TestEescfgNewEventExporterCfg(t *testing.T) {
	str := "test"
	bl := true
	tm := 1 * time.Second
	nm := 1
	eeo := &EventExporterOpts{
		CSVFieldSeparator:           &str,
		ElsIndex:                    &str,
		ElsRefresh:                  &str,
		ElsDiscoverNodesOnStart:     &bl,
		ElsDiscoverNodeInterval:     &tm,
		ElsCloud:                    &bl,
		ElsAPIKey:                   &str,
		ElsCertificateFingerprint:   &str,
		ElsServiceToken:             &str,
		ElsUsername:                 &str,
		ElsPassword:                 &str,
		ElsEnableDebugLogger:        &bl,
		ElsLogger:                   &str,
		ElsCompressRequestBody:      &bl,
		ElsCompressRequestBodyLevel: &nm,
		ElsRetryOnStatus:            &[]int{nm},
		ElsMaxRetries:               &nm,
		ElsDisableRetry:             &bl,
		ElsOpType:                   &str,
		ElsPipeline:                 &str,
		ElsRouting:                  &str,
		ElsTimeout:                  &tm,
		ElsWaitForActiveShards:      &str,
		SQLMaxIdleConns:             &nm,
		SQLMaxOpenConns:             &nm,
		SQLConnMaxLifetime:          &tm,
		MYSQLDSNParams:              map[string]string{str: str},
		SQLTableName:                &str,
		SQLDBName:                   &str,
		PgSSLMode:                   &str,
		KafkaTopic:                  &str,
		KafkaBatchSize:              &nm,
		KafkaTLS:                    &bl,
		KafkaCAPath:                 &str,
		KafkaSkipTLSVerify:          &bl,
		AMQPRoutingKey:              &str,
		AMQPQueueID:                 &str,
		AMQPExchange:                &str,
		AMQPExchangeType:            &str,
		AMQPUsername:                &str,
		AMQPPassword:                &str,
		AWSRegion:                   &str,
		AWSKey:                      &str,
		AWSSecret:                   &str,
		AWSToken:                    &str,
		SQSQueueID:                  &str,
		S3BucketID:                  &str,
		S3FolderPath:                &str,
		NATSJetStream:               &bl,
		NATSSubject:                 &str,
		NATSJWTFile:                 &str,
		NATSSeedFile:                &str,
		NATSCertificateAuthority:    &str,
		NATSClientCertificate:       &str,
		NATSClientKey:               &str,
		NATSJetStreamMaxWait:        &tm,
		RPCCodec:                    &str,
		ServiceMethod:               &str,
		KeyPath:                     &str,
		CertPath:                    &str,
		CAPath:                      &str,
		TLS:                         &bl,
		ConnIDs:                     &[]string{str},
		RPCConnTimeout:              &tm,
		RPCReplyTimeout:             &tm,
		RPCAPIOpts:                  map[string]any{str: bl},
	}
	rcv := NewEventExporterCfg(str, str, str, str, 1, eeo)
	exp := &EventExporterCfg{
		ID:             str,
		Type:           str,
		ExportPath:     str,
		FailedPostsDir: str,
		Attempts:       1,
		Opts:           eeo,
	}

	if !reflect.DeepEqual(exp, rcv) {
		t.Errorf("\nexpected: %s\n received: %s\n", utils.ToJSON(exp), utils.ToJSON(rcv))
	}

	rcv = NewEventExporterCfg(str, str, str, str, 1, nil)
	exp = &EventExporterCfg{
		ID:             str,
		Type:           str,
		ExportPath:     str,
		FailedPostsDir: str,
		Attempts:       1,
		Opts:           new(EventExporterOpts),
	}

	if !reflect.DeepEqual(exp, rcv) {
		t.Errorf("\nexpected: %s\n received: %s\n", utils.ToJSON(exp), utils.ToJSON(rcv))
	}
}

func TestEescfgloadFromJSONCfg(t *testing.T) {
	str := "test"
	bl := true
	tm := 1 * time.Second
	nm := 1
	tms := "1s"
	eeOpts := &EventExporterOpts{}
	jsnCfg := &EventExporterOptsJson{
		CSVFieldSeparator:           &str,
		ElsCloud:                    &bl,
		ElsAPIKey:                   &str,
		ElsServiceToken:             &str,
		ElsCertificateFingerprint:   &str,
		ElsUsername:                 &str,
		ElsPassword:                 &str,
		ElsDiscoverNodesOnStart:     &bl,
		ElsDiscoverNodesInterval:    &tms,
		ElsEnableDebugLogger:        &bl,
		ElsLogger:                   &str,
		ElsCompressRequestBody:      &bl,
		ElsCompressRequestBodyLevel: &nm,
		ElsRetryOnStatus:            &[]int{nm},
		ElsMaxRetries:               &nm,
		ElsDisableRetry:             &bl,
		ElsIndex:                    &str,
		ElsRefresh:                  &str,
		ElsOpType:                   &str,
		ElsPipeline:                 &str,
		ElsRouting:                  &str,
		ElsTimeout:                  &tms,
		ElsWaitForActiveShards:      &str,
		SQLMaxIdleConns:             &nm,
		SQLMaxOpenConns:             &nm,
		SQLConnMaxLifetime:          &tms,
		MYSQLDSNParams:              map[string]string{str: str},
		SQLTableName:                &str,
		SQLDBName:                   &str,
		PgSSLMode:                   &str,
		KafkaTopic:                  &str,
		KafkaBatchSize:              &nm,
		KafkaTLS:                    &bl,
		KafkaCAPath:                 &str,
		KafkaSkipTLSVerify:          &bl,
		AMQPQueueID:                 &str,
		AMQPRoutingKey:              &str,
		AMQPExchange:                &str,
		AMQPExchangeType:            &str,
		AMQPUsername:                &str,
		AMQPPassword:                &str,
		AWSRegion:                   &str,
		AWSKey:                      &str,
		AWSSecret:                   &str,
		AWSToken:                    &str,
		SQSQueueID:                  &str,
		S3BucketID:                  &str,
		S3FolderPath:                &str,
		NATSJetStream:               &bl,
		NATSSubject:                 &str,
		NATSJWTFile:                 &str,
		NATSSeedFile:                &str,
		NATSCertificateAuthority:    &str,
		NATSClientCertificate:       &str,
		NATSClientKey:               &str,
		NATSJetStreamMaxWait:        &tms,
		RPCCodec:                    &str,
		ServiceMethod:               &str,
		KeyPath:                     &str,
		CertPath:                    &str,
		CAPath:                      &str,
		ConnIDs:                     &[]string{str},
		TLS:                         &bl,
		RPCConnTimeout:              &tms,
		RPCReplyTimeout:             &tms,
		RPCAPIOpts:                  map[string]any{str: bl},
	}

	err := eeOpts.loadFromJSONCfg(jsnCfg)
	if err != nil {
		t.Error(err)
	}

	exp := &EventExporterOpts{
		CSVFieldSeparator:           &str,
		ElsIndex:                    &str,
		ElsRefresh:                  &str,
		ElsDiscoverNodesOnStart:     &bl,
		ElsDiscoverNodeInterval:     &tm,
		ElsCloud:                    &bl,
		ElsAPIKey:                   &str,
		ElsCertificateFingerprint:   &str,
		ElsServiceToken:             &str,
		ElsUsername:                 &str,
		ElsPassword:                 &str,
		ElsEnableDebugLogger:        &bl,
		ElsLogger:                   &str,
		ElsCompressRequestBody:      &bl,
		ElsCompressRequestBodyLevel: &nm,
		ElsRetryOnStatus:            &[]int{nm},
		ElsMaxRetries:               &nm,
		ElsDisableRetry:             &bl,
		ElsOpType:                   &str,
		ElsPipeline:                 &str,
		ElsRouting:                  &str,
		ElsTimeout:                  &tm,
		ElsWaitForActiveShards:      &str,
		SQLMaxIdleConns:             &nm,
		SQLMaxOpenConns:             &nm,
		SQLConnMaxLifetime:          &tm,
		MYSQLDSNParams:              map[string]string{str: str},
		SQLTableName:                &str,
		SQLDBName:                   &str,
		PgSSLMode:                   &str,
		KafkaTopic:                  &str,
		KafkaBatchSize:              &nm,
		KafkaTLS:                    &bl,
		KafkaCAPath:                 &str,
		KafkaSkipTLSVerify:          &bl,
		AMQPRoutingKey:              &str,
		AMQPQueueID:                 &str,
		AMQPExchange:                &str,
		AMQPExchangeType:            &str,
		AMQPUsername:                &str,
		AMQPPassword:                &str,
		AWSRegion:                   &str,
		AWSKey:                      &str,
		AWSSecret:                   &str,
		AWSToken:                    &str,
		SQSQueueID:                  &str,
		S3BucketID:                  &str,
		S3FolderPath:                &str,
		NATSJetStream:               &bl,
		NATSSubject:                 &str,
		NATSJWTFile:                 &str,
		NATSSeedFile:                &str,
		NATSCertificateAuthority:    &str,
		NATSClientCertificate:       &str,
		NATSClientKey:               &str,
		NATSJetStreamMaxWait:        &tm,
		RPCCodec:                    &str,
		ServiceMethod:               &str,
		KeyPath:                     &str,
		CertPath:                    &str,
		CAPath:                      &str,
		TLS:                         &bl,
		ConnIDs:                     &[]string{str},
		RPCConnTimeout:              &tm,
		RPCReplyTimeout:             &tm,
		RPCAPIOpts:                  map[string]any{str: bl},
	}

	if !reflect.DeepEqual(exp, eeOpts) {
		t.Errorf("\nexpected: %s\nreceived: %s\n", utils.ToJSON(exp), utils.ToJSON(eeOpts))
	}

	jsnCfg2 := &EventExporterOptsJson{
		ElsDiscoverNodesInterval: &str,
	}

	err = eeOpts.loadFromJSONCfg(jsnCfg2)
	if err != nil {
		if err.Error() != `time: invalid duration "test"` {
			t.Error(err)
		}
	} else {
		t.Error("was expecting an error")
	}
}

func TestEescfgClone(t *testing.T) {
	str := "test"
	eeOpts := &EventExporterOpts{
		AMQPUsername: &str,
		AMQPPassword: &str,
	}

	rcv := eeOpts.Clone()

	if !reflect.DeepEqual(eeOpts, rcv) {
		t.Errorf("\nexpected: %s\nreceived: %s\n", utils.ToJSON(eeOpts), utils.ToJSON(rcv))
	}
}

func TestEescfgAsMapInterface(t *testing.T) {
	str := "test"
	eeOpts := &EventExporterOpts{
		AMQPUsername: &str,
		AMQPPassword: &str,
	}

	rcv := eeOpts.AsMapInterface()
	exp := map[string]any{
		utils.AMQPUsername: *eeOpts.AMQPUsername,
		utils.AMQPPassword: *eeOpts.AMQPPassword,
	}

	if !reflect.DeepEqual(exp, rcv) {
		t.Errorf("\nexpected: %s\nreceived: %s\n", utils.ToJSON(exp), utils.ToJSON(rcv))
	}
}

func TestEescfgdiffEventExporterOptsJsonCfg(t *testing.T) {
	str := "test"
	v2 := &EventExporterOpts{
		AMQPUsername: &str,
		AMQPPassword: &str,
	}
	v1 := &EventExporterOpts{}
	d := &EventExporterOptsJson{}
	rcv := diffEventExporterOptsJsonCfg(d, v1, v2)

	expD := &EventExporterOptsJson{
		AMQPUsername: &str,
		AMQPPassword: &str,
	}

	if !reflect.DeepEqual(expD, d) {
		t.Errorf("\nexpected: %s\nreceived: %s\n", utils.ToJSON(expD), utils.ToJSON(rcv))
	}
}
