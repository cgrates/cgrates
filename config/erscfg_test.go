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

func TestERSClone(t *testing.T) {
	cfgJSONStr := `{
"ers": {									
	"enabled": true,						
	"sessions_conns":["*internal"],			
	"readers": [
         {
            "id": "file_reader1",
			"run_delay": "-1",
			"type": "*file_csv",
			"flags": ["*dryrun"],
			"source_path": "/tmp/ers/in",
			"processed_path": "/tmp/ers/out",
			"opts": {
              "*default": "randomVal"
             },
			"xml_root_path": "",								
			"tenant": "~*req.Destination1",											
			"timezone": "",										
			"filters": ["randomFiletrs"],										
			"flags": [],										
			"fields":[											
				{"tag": "ToR", "path": "*cgreq.ToR", "type": "*variable", "value": "~*req.2", "mandatory": true},
				{"tag": "OriginID", "path": "*cgreq.OriginID", "type": "*variable", "value": "~*req.3", "mandatory": true},
			],
			"cache_dump_fields": [
               {"tag": "ToR", "path": "*cgreq.ToR", "type": "*variable", "value": "~*req.2", "mandatory": true},
            ],
			"partial_commit_fields": [{
				"mandatory": true,
				"path": "*cgreq.ToR",
				"tag": "ToR",
				"type": "*variable",
				"value": "~*req.2"
			}],
            "failed_calls_prefix": "randomPrefix",
            "partial_record_cache": "1s",
            "partial_cache_expiry_action": "randomAction",
         },
	],
},
}`
	expectedERsCfg := &ERsCfg{
		Enabled:       true,
		SessionSConns: []string{"*internal:*sessions"},
		Readers: []*EventReaderCfg{
			{
				ID:             utils.MetaDefault,
				Type:           utils.MetaNone,
				RunDelay:       0,
				ConcurrentReqs: 1024,
				SourcePath:     "/var/spool/cgrates/ers/in",
				ProcessedPath:  "/var/spool/cgrates/ers/out",
				Tenant:         nil,
				Timezone:       utils.EmptyString,
				Filters:        []string{},
				Flags:          utils.FlagsWithParams{},
				Fields: []*FCTemplate{
					{Tag: utils.ToR, Path: utils.MetaCgreq + utils.NestingSep + utils.ToR, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.2", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.OriginID, Path: utils.MetaCgreq + utils.NestingSep + utils.OriginID, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.3", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.RequestType, Path: utils.MetaCgreq + utils.NestingSep + utils.RequestType, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.4", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.Tenant, Path: utils.MetaCgreq + utils.NestingSep + utils.Tenant, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.6", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.Category, Path: utils.MetaCgreq + utils.NestingSep + utils.Category, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.7", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.AccountField, Path: utils.MetaCgreq + utils.NestingSep + utils.AccountField, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.8", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.Subject, Path: utils.MetaCgreq + utils.NestingSep + utils.Subject, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.9", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.Destination, Path: utils.MetaCgreq + utils.NestingSep + utils.Destination, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.10", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.SetupTime, Path: utils.MetaCgreq + utils.NestingSep + utils.SetupTime, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.11", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.AnswerTime, Path: utils.MetaCgreq + utils.NestingSep + utils.AnswerTime, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.12", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.Usage, Path: utils.MetaCgreq + utils.NestingSep + utils.Usage, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.13", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
				},
				CacheDumpFields:     make([]*FCTemplate, 0),
				PartialCommitFields: make([]*FCTemplate, 0),
				Opts: map[string]interface{}{
					"csvFieldSeparator":   ",",
					"csvHeaderDefineChar": ":",
					"csvRowLength":        0.,
					"xmlRootPath":         "",
					"partialOrderField":   "~*req.AnswerTime",
					"natsSubject":         "cgrates_cdrs",
				},
			},
			{
				ID:             "file_reader1",
				Type:           "*file_csv",
				RunDelay:       -1,
				ConcurrentReqs: 1024,
				SourcePath:     "/tmp/ers/in",
				ProcessedPath:  "/tmp/ers/out",
				Tenant:         NewRSRParsersMustCompile("~*req.Destination1", utils.InfieldSep),
				Timezone:       utils.EmptyString,
				Filters:        []string{"randomFiletrs"},
				Flags:          utils.FlagsWithParams{},
				Fields: []*FCTemplate{
					{Tag: utils.ToR, Path: utils.MetaCgreq + utils.NestingSep + utils.ToR, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.2", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.OriginID, Path: utils.MetaCgreq + utils.NestingSep + utils.OriginID, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.3", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
				},
				CacheDumpFields: []*FCTemplate{
					{Tag: utils.ToR, Path: utils.MetaCgreq + utils.NestingSep + utils.ToR, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.2", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
				},
				PartialCommitFields: []*FCTemplate{
					{
						Type:      utils.MetaVariable,
						Value:     NewRSRParsersMustCompile("~*req.2", utils.InfieldSep),
						Tag:       utils.ToR,
						Path:      utils.MetaCgreq + utils.NestingSep + utils.ToR,
						Mandatory: true,
						Layout:    time.RFC3339,
					},
				},
				Opts: map[string]interface{}{
					utils.MetaDefault:     "randomVal",
					"csvFieldSeparator":   ",",
					"csvHeaderDefineChar": ":",
					"csvRowLength":        0.,
					"xmlRootPath":         "",
					"partialOrderField":   "~*req.AnswerTime",
					"natsSubject":         "cgrates_cdrs",
				},
			},
		},
		PartialCacheTTL:    time.Second,
		PartialCacheAction: utils.MetaNone,
	}
	for _, profile := range expectedERsCfg.Readers {
		for _, v := range profile.Fields {
			v.ComputePath()
		}
		for _, v := range profile.CacheDumpFields {
			v.ComputePath()
		}
		for _, v := range profile.PartialCommitFields {
			v.ComputePath()
		}
	}
	if jsonCfg, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else {
		clonedErs := jsonCfg.ersCfg.Clone()
		if !reflect.DeepEqual(clonedErs, expectedERsCfg) {
			t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expectedERsCfg), utils.ToJSON(clonedErs))
		}
	}
}

func TestEventReaderloadFromJsonCfg(t *testing.T) {
	jsonCfg := NewDefaultCGRConfig()

	eventReader := new(EventReaderCfg)
	if err = eventReader.loadFromJSONCfg(nil, jsonCfg.templates, jsonCfg.generalCfg.RSRSep); err != nil {
		t.Error(err)
	}
}

func TestEventReaderloadFromJsonCase1(t *testing.T) {
	cfgJSON := &ERsJsonCfg{
		Readers: &[]*EventReaderJsonCfg{
			{
				Run_delay: utils.StringPointer("1ss"),
			},
		},
	}
	expected := "time: unknown unit \"ss\" in duration \"1ss\""
	jsoncfg := NewDefaultCGRConfig()
	if err = jsoncfg.ersCfg.loadFromJSONCfg(cfgJSON, jsoncfg.templates, jsoncfg.generalCfg.RSRSep, jsoncfg.dfltEvRdr); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestEventReaderloadFromJsonCase3(t *testing.T) {
	cfgJSON := &ERsJsonCfg{
		Readers: &[]*EventReaderJsonCfg{
			{
				Tenant: utils.StringPointer("a{*"),
			},
		},
	}
	expected := "invalid converter terminator in rule: <a{*>"
	jsoncfg := NewDefaultCGRConfig()
	if err = jsoncfg.ersCfg.loadFromJSONCfg(cfgJSON, jsoncfg.templates, jsoncfg.generalCfg.RSRSep, jsoncfg.dfltEvRdr); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestEventReaderloadFromJsonCase2(t *testing.T) {
	cfgJSON := &ERsJsonCfg{
		Partial_cache_ttl: utils.StringPointer("1ss"),
	}
	expected := `time: unknown unit "ss" in duration "1ss"`
	jsonCfg := NewDefaultCGRConfig()
	if err = jsonCfg.ersCfg.loadFromJSONCfg(cfgJSON, jsonCfg.templates, jsonCfg.generalCfg.RSRSep, jsonCfg.dfltEvRdr); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestERSLoadFromjsonCfg(t *testing.T) {
	expectedERsCfg := &ERsCfg{
		Enabled:       true,
		SessionSConns: []string{"conn1", "conn3"},
		Readers: []*EventReaderCfg{
			{
				ID:             utils.MetaDefault,
				Type:           utils.MetaNone,
				RunDelay:       0,
				ConcurrentReqs: 1024,
				SourcePath:     "/var/spool/cgrates/ers/in",
				ProcessedPath:  "/var/spool/cgrates/ers/out",
				Tenant:         nil,
				Timezone:       utils.EmptyString,
				Filters:        []string{},
				Flags:          utils.FlagsWithParams{},
				Fields: []*FCTemplate{
					{Tag: utils.ToR, Path: utils.MetaCgreq + utils.NestingSep + utils.ToR, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.2", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.OriginID, Path: utils.MetaCgreq + utils.NestingSep + utils.OriginID, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.3", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.RequestType, Path: utils.MetaCgreq + utils.NestingSep + utils.RequestType, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.4", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.Tenant, Path: utils.MetaCgreq + utils.NestingSep + utils.Tenant, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.6", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.Category, Path: utils.MetaCgreq + utils.NestingSep + utils.Category, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.7", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.AccountField, Path: utils.MetaCgreq + utils.NestingSep + utils.AccountField, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.8", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.Subject, Path: utils.MetaCgreq + utils.NestingSep + utils.Subject, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.9", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.Destination, Path: utils.MetaCgreq + utils.NestingSep + utils.Destination, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.10", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.SetupTime, Path: utils.MetaCgreq + utils.NestingSep + utils.SetupTime, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.11", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.AnswerTime, Path: utils.MetaCgreq + utils.NestingSep + utils.AnswerTime, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.12", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.Usage, Path: utils.MetaCgreq + utils.NestingSep + utils.Usage, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.13", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
				},
				CacheDumpFields:     make([]*FCTemplate, 0),
				PartialCommitFields: make([]*FCTemplate, 0),
				Opts: map[string]interface{}{
					"csvFieldSeparator":   ",",
					"csvHeaderDefineChar": ":",
					"csvRowLength":        0.,
					"xmlRootPath":         "",
					"partialOrderField":   "~*req.AnswerTime",
					"natsSubject":         "cgrates_cdrs",
				},
			},
			{
				ID:             "file_reader1",
				Type:           utils.MetaFileCSV,
				RunDelay:       -1,
				ConcurrentReqs: 1024,
				SourcePath:     "/tmp/ers/in",
				ProcessedPath:  "/tmp/ers/out",
				Tenant:         nil,
				Timezone:       utils.EmptyString,
				Filters:        []string{},
				Flags:          utils.FlagsWithParams{},
				Fields: []*FCTemplate{
					{Tag: utils.ToR, Path: utils.MetaCgreq + utils.NestingSep + utils.ToR, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.2", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.OriginID, Path: utils.MetaCgreq + utils.NestingSep + utils.OriginID, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.3", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.RequestType, Path: utils.MetaCgreq + utils.NestingSep + utils.RequestType, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.4", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.Tenant, Path: utils.MetaCgreq + utils.NestingSep + utils.Tenant, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.6", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.Category, Path: utils.MetaCgreq + utils.NestingSep + utils.Category, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.7", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.AccountField, Path: utils.MetaCgreq + utils.NestingSep + utils.AccountField, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.8", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.Subject, Path: utils.MetaCgreq + utils.NestingSep + utils.Subject, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.9", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.Destination, Path: utils.MetaCgreq + utils.NestingSep + utils.Destination, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.10", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.SetupTime, Path: utils.MetaCgreq + utils.NestingSep + utils.SetupTime, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.11", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.AnswerTime, Path: utils.MetaCgreq + utils.NestingSep + utils.AnswerTime, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.12", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.Usage, Path: utils.MetaCgreq + utils.NestingSep + utils.Usage, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.13", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
				},
				CacheDumpFields:     make([]*FCTemplate, 0),
				PartialCommitFields: make([]*FCTemplate, 0),
				Opts: map[string]interface{}{
					"csvFieldSeparator":   ",",
					"csvHeaderDefineChar": ":",
					"csvRowLength":        0.,
					"xmlRootPath":         "",
					"partialOrderField":   "~*req.AnswerTime",
					"natsSubject":         "cgrates_cdrs",
				},
			},
		},
		PartialCacheTTL:    time.Second,
		PartialCacheAction: utils.MetaNone,
	}
	for _, profile := range expectedERsCfg.Readers {
		for _, v := range profile.Fields {
			v.ComputePath()
		}
	}
	cfgJSONStr := `{
"ers": {
	"enabled": true,
	"sessions_conns":["conn1","conn3"],
	"readers": [
		{
			"id": "file_reader1",
			"run_delay":  "-1",
			"type": "*file_csv",
			"source_path": "/tmp/ers/in",
			"processed_path": "/tmp/ers/out",
			"cache_dump_fields": [],
		},
	],
}
}`

	if cfg, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedERsCfg, cfg.ersCfg) {
		t.Errorf("Expected: %+v ,\n received: %+v", utils.ToJSON(expectedERsCfg), utils.ToJSON(cfg.ersCfg))
	}

}

func TestERSloadFromJsonCfg(t *testing.T) {
	cfgJSON := &ERsJsonCfg{
		Readers: nil,
	}
	jsonCfg := NewDefaultCGRConfig()
	if err = jsonCfg.ersCfg.loadFromJSONCfg(cfgJSON, jsonCfg.templates, jsonCfg.generalCfg.RSRSep, jsonCfg.dfltEvRdr); err != nil {
		t.Error(err)
	}
}

func TestEventReaderloadFromJsonCfgErr1(t *testing.T) {
	erS := &EventReaderCfg{
		PartialCommitFields: []*FCTemplate{
			{
				Type:  utils.MetaTemplate,
				Value: NewRSRParsersMustCompile("1sa{*duration}", utils.InfieldSep),
			},
		},
	}
	jsnCfg := &EventReaderJsonCfg{
		Partial_commit_fields: &[]*FcTemplateJsonCfg{
			{
				Type:  utils.StringPointer(utils.MetaTemplate),
				Value: utils.StringPointer("1sa{*duration}"),
			},
		},
	}
	expected := "time: unknown unit \"sa\" in duration \"1sa\""
	cfg := NewDefaultCGRConfig()
	if err = erS.loadFromJSONCfg(jsnCfg, cfg.templates, cfg.generalCfg.RSRSep); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestEventReaderloadFromJsonCfgErr2(t *testing.T) {
	erS := &EventReaderCfg{
		PartialCommitFields: make([]*FCTemplate, 0),
	}
	jsnCfg := &EventReaderJsonCfg{
		Partial_commit_fields: &[]*FcTemplateJsonCfg{
			{
				Value: utils.StringPointer("a{*"),
			},
		},
	}
	expected := "invalid converter terminator in rule: <a{*>"
	cfg := NewDefaultCGRConfig()
	if err = erS.loadFromJSONCfg(jsnCfg, cfg.templates, cfg.generalCfg.RSRSep); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestEventReaderloadFromJsonCfgErr3(t *testing.T) {
	erS := &EventReaderCfg{
		PartialCommitFields: []*FCTemplate{
			{
				Type:  utils.MetaTemplate,
				Value: NewRSRParsersMustCompile("value", utils.InfieldSep),
			},
		},
	}
	jsnCfg := &EventReaderJsonCfg{
		Partial_commit_fields: &[]*FcTemplateJsonCfg{
			{
				Tag:   utils.StringPointer("tag2"),
				Type:  utils.StringPointer(utils.MetaTemplate),
				Value: utils.StringPointer("value"),
			},
		},
	}
	tmpl := FCTemplates{
		"value": {
			{
				Type: utils.MetaVariable,
				Tag:  "tag",
			},
		},
	}
	cfg := NewDefaultCGRConfig()
	if err = erS.loadFromJSONCfg(jsnCfg, tmpl, cfg.generalCfg.RSRSep); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(erS.PartialCommitFields, tmpl["value"]) {
		t.Errorf("Expected %v \n but received \n %v", erS.PartialCommitFields, tmpl["value"])
	}
}

func TestEventReaderFieldsloadFromJsonCfg(t *testing.T) {
	cfgJSON := &ERsJsonCfg{
		Readers: &[]*EventReaderJsonCfg{
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
	if err = jsonCfg.ersCfg.loadFromJSONCfg(cfgJSON, jsonCfg.templates, jsonCfg.generalCfg.RSRSep, jsonCfg.dfltEvRdr); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestERSloadFromJsonCase1(t *testing.T) {
	cfgJSON := &ERsJsonCfg{
		Readers: &[]*EventReaderJsonCfg{
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
	if err = jsonCfg.ersCfg.loadFromJSONCfg(cfgJSON, jsonCfg.templates, jsonCfg.generalCfg.RSRSep, jsonCfg.dfltEvRdr); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}
func TestERSloadFromJsonCase2(t *testing.T) {
	cfgJSON := &ERsJsonCfg{
		Readers: &[]*EventReaderJsonCfg{
			{
				Cache_dump_fields: &[]*FcTemplateJsonCfg{
					{

						Type: utils.StringPointer(utils.MetaTemplate),
					},
				},
			},
		},
	}
	expected := "no template with id: <>"
	jsonCfg := NewDefaultCGRConfig()
	if err = jsonCfg.ersCfg.loadFromJSONCfg(cfgJSON, jsonCfg.templates, jsonCfg.generalCfg.RSRSep, jsonCfg.dfltEvRdr); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestERSloadFromJsonCase3(t *testing.T) {
	cfgJSON := &ERsJsonCfg{
		Enabled:        utils.BoolPointer(true),
		Sessions_conns: &[]string{"*conn1"},
		Readers: &[]*EventReaderJsonCfg{
			{
				Id:                  utils.StringPointer("file_reader1"),
				Type:                utils.StringPointer(utils.MetaFileCSV),
				Run_delay:           utils.StringPointer("-1"),
				Concurrent_requests: utils.IntPointer(1024),
				Source_path:         utils.StringPointer("/tmp/ers/in"),
				Processed_path:      utils.StringPointer("/tmp/ers/out"),
				Tenant:              nil,
				Timezone:            utils.StringPointer(""),
				Filters:             nil,
				Flags:               &[]string{},
				Fields: &[]*FcTemplateJsonCfg{
					{
						Tag:    utils.StringPointer(utils.AnswerTime),
						Path:   utils.StringPointer("*cgreq.AnswerTime"),
						Type:   utils.StringPointer(utils.MetaTemplate),
						Value:  utils.StringPointer("randomTemplate"),
						Layout: utils.StringPointer("2006-01-02T15:04:05Z07:00"),
					},
				},
			},
		},
	}
	expectedERsCfg := &ERsCfg{
		Enabled:       true,
		SessionSConns: []string{"*conn1"},
		Readers: []*EventReaderCfg{
			{
				ID:             utils.MetaDefault,
				Type:           utils.MetaNone,
				RunDelay:       0,
				ConcurrentReqs: 1024,
				SourcePath:     "/var/spool/cgrates/ers/in",
				ProcessedPath:  "/var/spool/cgrates/ers/out",
				Tenant:         nil,
				Timezone:       utils.EmptyString,
				Filters:        []string{},
				Flags:          utils.FlagsWithParams{},
				Fields: []*FCTemplate{
					{Tag: utils.ToR, Path: utils.MetaCgreq + utils.NestingSep + utils.ToR, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.2", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.OriginID, Path: utils.MetaCgreq + utils.NestingSep + utils.OriginID, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.3", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.RequestType, Path: utils.MetaCgreq + utils.NestingSep + utils.RequestType, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.4", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.Tenant, Path: utils.MetaCgreq + utils.NestingSep + utils.Tenant, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.6", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.Category, Path: utils.MetaCgreq + utils.NestingSep + utils.Category, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.7", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.AccountField, Path: utils.MetaCgreq + utils.NestingSep + utils.AccountField, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.8", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.Subject, Path: utils.MetaCgreq + utils.NestingSep + utils.Subject, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.9", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.Destination, Path: utils.MetaCgreq + utils.NestingSep + utils.Destination, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.10", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.SetupTime, Path: utils.MetaCgreq + utils.NestingSep + utils.SetupTime, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.11", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.AnswerTime, Path: utils.MetaCgreq + utils.NestingSep + utils.AnswerTime, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.12", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.Usage, Path: utils.MetaCgreq + utils.NestingSep + utils.Usage, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.13", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
				},
				CacheDumpFields:     make([]*FCTemplate, 0),
				PartialCommitFields: make([]*FCTemplate, 0),
				Opts: map[string]interface{}{
					"csvFieldSeparator":   ",",
					"csvHeaderDefineChar": ":",
					"csvRowLength":        0.,
					"xmlRootPath":         "",
					"partialOrderField":   "~*req.AnswerTime",
					"natsSubject":         "cgrates_cdrs",
				},
			},
			{
				ID:             "file_reader1",
				Type:           utils.MetaFileCSV,
				RunDelay:       -1,
				ConcurrentReqs: 1024,
				SourcePath:     "/tmp/ers/in",
				ProcessedPath:  "/tmp/ers/out",
				Tenant:         nil,
				Timezone:       utils.EmptyString,
				Filters:        []string{},
				Flags:          utils.FlagsWithParams{},
				Fields: []*FCTemplate{
					{
						Tag:    utils.CGRID,
						Path:   "*exp.CGRID",
						Type:   utils.MetaVariable,
						Layout: time.RFC3339,
					},
				},
				CacheDumpFields:     make([]*FCTemplate, 0),
				PartialCommitFields: make([]*FCTemplate, 0),
				Opts: map[string]interface{}{
					"csvFieldSeparator":   ",",
					"csvHeaderDefineChar": ":",
					"csvRowLength":        0.,
					"xmlRootPath":         "",
					"partialOrderField":   "~*req.AnswerTime",
					"natsSubject":         "cgrates_cdrs",
				},
			},
		},

		PartialCacheTTL:    time.Second,
		PartialCacheAction: utils.MetaNone,
	}
	msgTemplates := map[string][]*FCTemplate{
		"randomTemplate": {
			{
				Tag:    utils.CGRID,
				Path:   "*exp.CGRID",
				Type:   utils.MetaVariable,
				Layout: time.RFC3339,
			},
		},
	}
	for _, v := range expectedERsCfg.Readers[0].Fields {
		v.ComputePath()
	}
	jsonCfg := NewDefaultCGRConfig()
	if err = jsonCfg.ersCfg.loadFromJSONCfg(cfgJSON, msgTemplates, jsonCfg.generalCfg.RSRSep, jsonCfg.dfltEvRdr); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedERsCfg, jsonCfg.ersCfg) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expectedERsCfg), utils.ToJSON(jsonCfg.ersCfg))
	}
}

func TestERSloadFromJsonCase4(t *testing.T) {
	cfgJSON := &ERsJsonCfg{
		Enabled:        utils.BoolPointer(true),
		Sessions_conns: &[]string{"*conn1"},
		Readers: &[]*EventReaderJsonCfg{
			{
				Id:                  utils.StringPointer("file_reader1"),
				Type:                utils.StringPointer(utils.MetaFileCSV),
				Run_delay:           utils.StringPointer("-1"),
				Concurrent_requests: utils.IntPointer(1024),
				Source_path:         utils.StringPointer("/tmp/ers/in"),
				Processed_path:      utils.StringPointer("/tmp/ers/out"),
				Tenant:              nil,
				Timezone:            utils.StringPointer(""),
				Filters:             nil,
				Flags:               &[]string{},
				Fields:              &[]*FcTemplateJsonCfg{},
				Cache_dump_fields: &[]*FcTemplateJsonCfg{
					{
						Tag:   utils.StringPointer("OriginID"),
						Path:  utils.StringPointer("*exp.OriginID"),
						Type:  utils.StringPointer(utils.MetaTemplate),
						Value: utils.StringPointer("randomTemplate"),
					},
				},
			},
		},
	}
	expectedERsCfg := &ERsCfg{
		Enabled:       true,
		SessionSConns: []string{"*conn1"},
		Readers: []*EventReaderCfg{
			{
				ID:             utils.MetaDefault,
				Type:           utils.MetaNone,
				RunDelay:       0,
				ConcurrentReqs: 1024,
				SourcePath:     "/var/spool/cgrates/ers/in",
				ProcessedPath:  "/var/spool/cgrates/ers/out",
				Tenant:         nil,
				Timezone:       utils.EmptyString,
				Filters:        []string{},
				Flags:          utils.FlagsWithParams{},
				Fields: []*FCTemplate{
					{Tag: utils.ToR, Path: utils.MetaCgreq + utils.NestingSep + utils.ToR, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.2", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.OriginID, Path: utils.MetaCgreq + utils.NestingSep + utils.OriginID, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.3", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.RequestType, Path: utils.MetaCgreq + utils.NestingSep + utils.RequestType, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.4", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.Tenant, Path: utils.MetaCgreq + utils.NestingSep + utils.Tenant, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.6", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.Category, Path: utils.MetaCgreq + utils.NestingSep + utils.Category, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.7", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.AccountField, Path: utils.MetaCgreq + utils.NestingSep + utils.AccountField, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.8", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.Subject, Path: utils.MetaCgreq + utils.NestingSep + utils.Subject, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.9", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.Destination, Path: utils.MetaCgreq + utils.NestingSep + utils.Destination, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.10", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.SetupTime, Path: utils.MetaCgreq + utils.NestingSep + utils.SetupTime, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.11", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.AnswerTime, Path: utils.MetaCgreq + utils.NestingSep + utils.AnswerTime, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.12", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.Usage, Path: utils.MetaCgreq + utils.NestingSep + utils.Usage, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.13", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
				},
				CacheDumpFields:     make([]*FCTemplate, 0),
				PartialCommitFields: make([]*FCTemplate, 0),
				Opts: map[string]interface{}{
					"csvFieldSeparator":   ",",
					"csvHeaderDefineChar": ":",
					"csvRowLength":        0.,
					"xmlRootPath":         "",
					"partialOrderField":   "~*req.AnswerTime",
					"natsSubject":         "cgrates_cdrs",
				},
			},
			{
				ID:             "file_reader1",
				Type:           utils.MetaFileCSV,
				RunDelay:       -1,
				ConcurrentReqs: 1024,
				SourcePath:     "/tmp/ers/in",
				ProcessedPath:  "/tmp/ers/out",
				Tenant:         nil,
				Timezone:       utils.EmptyString,
				Filters:        []string{},
				Flags:          utils.FlagsWithParams{},
				Fields:         []*FCTemplate{},
				CacheDumpFields: []*FCTemplate{
					{
						Tag:   "OrderID",
						Path:  "*exp.OrderID",
						Type:  "*variable",
						Value: NewRSRParsersMustCompile("~*req.OrderID", utils.InfieldSep),
					},
				},
				PartialCommitFields: make([]*FCTemplate, 0),
				Opts: map[string]interface{}{
					"csvFieldSeparator":   ",",
					"csvHeaderDefineChar": ":",
					"csvRowLength":        0.,
					"xmlRootPath":         "",
					"partialOrderField":   "~*req.AnswerTime",
					"natsSubject":         "cgrates_cdrs",
				},
			},
		},

		PartialCacheTTL:    time.Second,
		PartialCacheAction: utils.MetaNone,
	}
	msgTemplates := map[string][]*FCTemplate{
		"randomTemplate": {
			{
				Tag:   "OrderID",
				Path:  "*exp.OrderID",
				Type:  "*variable",
				Value: NewRSRParsersMustCompile("~*req.OrderID", utils.InfieldSep),
			},
		},
	}
	for _, v := range expectedERsCfg.Readers[0].CacheDumpFields {
		v.ComputePath()
	}
	for _, v := range expectedERsCfg.Readers[0].Fields {
		v.ComputePath()
	}
	jsonCfg := NewDefaultCGRConfig()
	if err = jsonCfg.ersCfg.loadFromJSONCfg(cfgJSON, msgTemplates, jsonCfg.generalCfg.RSRSep, jsonCfg.dfltEvRdr); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedERsCfg, jsonCfg.ersCfg) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expectedERsCfg), utils.ToJSON(jsonCfg.ersCfg))
	}
}

func TestEventReaderCacheDumpFieldsloadFromJsonCfg(t *testing.T) {
	cfgJSON := &ERsJsonCfg{
		Readers: &[]*EventReaderJsonCfg{
			{
				Cache_dump_fields: &[]*FcTemplateJsonCfg{
					{
						Value: utils.StringPointer("a{*"),
					},
				},
			},
		},
	}
	expected := "invalid converter terminator in rule: <a{*>"
	jsonCfg := NewDefaultCGRConfig()
	if err = jsonCfg.ersCfg.loadFromJSONCfg(cfgJSON, jsonCfg.templates, jsonCfg.generalCfg.RSRSep, jsonCfg.dfltEvRdr); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestEventReaderSameID(t *testing.T) {
	expectedERsCfg := &ERsCfg{
		Enabled:       true,
		SessionSConns: []string{"conn1"},
		Readers: []*EventReaderCfg{
			{
				ID:             utils.MetaDefault,
				Type:           utils.MetaNone,
				RunDelay:       0,
				ConcurrentReqs: 1024,
				SourcePath:     "/var/spool/cgrates/ers/in",
				ProcessedPath:  "/var/spool/cgrates/ers/out",
				Tenant:         nil,
				Timezone:       utils.EmptyString,
				Filters:        []string{},
				Flags:          utils.FlagsWithParams{},
				Fields: []*FCTemplate{
					{Tag: utils.ToR, Path: utils.MetaCgreq + utils.NestingSep + utils.ToR, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.2", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.OriginID, Path: utils.MetaCgreq + utils.NestingSep + utils.OriginID, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.3", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.RequestType, Path: utils.MetaCgreq + utils.NestingSep + utils.RequestType, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.4", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.Tenant, Path: utils.MetaCgreq + utils.NestingSep + utils.Tenant, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.6", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.Category, Path: utils.MetaCgreq + utils.NestingSep + utils.Category, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.7", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.AccountField, Path: utils.MetaCgreq + utils.NestingSep + utils.AccountField, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.8", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.Subject, Path: utils.MetaCgreq + utils.NestingSep + utils.Subject, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.9", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.Destination, Path: utils.MetaCgreq + utils.NestingSep + utils.Destination, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.10", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.SetupTime, Path: utils.MetaCgreq + utils.NestingSep + utils.SetupTime, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.11", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.AnswerTime, Path: utils.MetaCgreq + utils.NestingSep + utils.AnswerTime, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.12", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.Usage, Path: utils.MetaCgreq + utils.NestingSep + utils.Usage, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.13", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
				},
				CacheDumpFields:     make([]*FCTemplate, 0),
				PartialCommitFields: make([]*FCTemplate, 0),
				Opts: map[string]interface{}{
					"csvFieldSeparator":   ",",
					"csvHeaderDefineChar": ":",
					"csvRowLength":        0.,
					"xmlRootPath":         "",
					"partialOrderField":   "~*req.AnswerTime",
					"natsSubject":         "cgrates_cdrs",
				},
			},
			{
				ID:             "file_reader1",
				Type:           utils.MetaFileCSV,
				RunDelay:       -1,
				ConcurrentReqs: 1024,
				SourcePath:     "/tmp/ers/in",
				ProcessedPath:  "/tmp/ers/out",
				Tenant:         nil,
				Timezone:       utils.EmptyString,
				Filters:        []string{},
				Flags:          utils.FlagsWithParams{},
				Fields: []*FCTemplate{
					{Tag: "CustomTag2", Path: "CustomPath2", Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("CustomValue2", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
				},
				CacheDumpFields:     make([]*FCTemplate, 0),
				PartialCommitFields: make([]*FCTemplate, 0),
				Opts: map[string]interface{}{
					"csvFieldSeparator":   ",",
					"csvHeaderDefineChar": ":",
					"csvRowLength":        0.,
					"xmlRootPath":         "",
					"partialOrderField":   "~*req.AnswerTime",
					"natsSubject":         "cgrates_cdrs",
				},
			},
		},

		PartialCacheTTL:    time.Second,
		PartialCacheAction: utils.MetaNone,
	}
	for _, profile := range expectedERsCfg.Readers {
		for _, v := range profile.Fields {
			v.ComputePath()
		}
	}
	cfgJSONStr := `{
"ers": {
	"enabled": true,
	"sessions_conns":["conn1"],
	"readers": [
		{
			"id": "file_reader1",
			"run_delay":  "-1",
			"type": "*file_csv",
			"row_length" : 5,
			"source_path": "/tmp/ers/in",
			"processed_path": "/tmp/ers/out",
			"fields":[
				{"tag": "CustomTag1", "path": "CustomPath1", "type": "*variable", "value": "CustomValue1", "mandatory": true},
			],
			"cache_dump_fields": [],
		},
		{
			"id": "file_reader1",
			"run_delay":  "-1",
			"type": "*file_csv",
			"source_path": "/tmp/ers/in",
			"processed_path": "/tmp/ers/out",
			"fields":[
				{"tag": "CustomTag2", "path": "CustomPath2", "type": "*variable", "value": "CustomValue2", "mandatory": true},
			],
			"cache_dump_fields": [],
		},
	],
}
}`

	if cfg, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedERsCfg, cfg.ersCfg) {
		t.Errorf("Expected: %+v ,\n received: %+v", utils.ToJSON(expectedERsCfg), utils.ToJSON(cfg.ersCfg))
	}

}

func TestERsCfgAsMapInterfaceCase1(t *testing.T) {
	cfgJSONStr := `{
	"ers": {
		"enabled": true,
		"sessions_conns":["conn1","conn3"],
		"readers": [
			{
				"id": "file_reader1",
				"run_delay":  "-1",
                "tenant": "~*req.Destination1",
				"type": "*file_csv",
				"source_path": "/tmp/ers/in",
				"processed_path": "/tmp/ers/out",
				"cache_dump_fields": [],
			},
		],
	}
}`
	eMap := map[string]interface{}{
		utils.EnabledCfg:       true,
		utils.SessionSConnsCfg: []string{"conn1", "conn3"},
		utils.ReadersCfg: []map[string]interface{}{
			{
				utils.FiltersCfg:             []string{},
				utils.FlagsCfg:               []string{},
				utils.IDCfg:                  "*default",
				utils.ProcessedPathCfg:       "/var/spool/cgrates/ers/out",
				utils.RunDelayCfg:            "0",
				utils.SourcePathCfg:          "/var/spool/cgrates/ers/in",
				utils.TenantCfg:              "",
				utils.TimezoneCfg:            "",
				utils.CacheDumpFieldsCfg:     []map[string]interface{}{},
				utils.PartialCommitFieldsCfg: []map[string]interface{}{},
				utils.ConcurrentRequestsCfg:  1024,
				utils.TypeCfg:                "*none",
				utils.FieldsCfg: []map[string]interface{}{
					{utils.MandatoryCfg: true, utils.PathCfg: "*cgreq.ToR", utils.TagCfg: "ToR", utils.TypeCfg: "*variable", utils.ValueCfg: "~*req.2"},
					{utils.MandatoryCfg: true, utils.PathCfg: "*cgreq.OriginID", utils.TagCfg: "OriginID", utils.TypeCfg: "*variable", utils.ValueCfg: "~*req.3"},
					{utils.MandatoryCfg: true, utils.PathCfg: "*cgreq.RequestType", utils.TagCfg: "RequestType", utils.TypeCfg: "*variable", utils.ValueCfg: "~*req.4"},
					{utils.MandatoryCfg: true, utils.PathCfg: "*cgreq.Tenant", utils.TagCfg: "Tenant", utils.TypeCfg: "*variable", utils.ValueCfg: "~*req.6"},
					{utils.MandatoryCfg: true, utils.PathCfg: "*cgreq.Category", utils.TagCfg: "Category", utils.TypeCfg: "*variable", utils.ValueCfg: "~*req.7"},
					{utils.MandatoryCfg: true, utils.PathCfg: "*cgreq.Account", utils.TagCfg: "Account", utils.TypeCfg: "*variable", utils.ValueCfg: "~*req.8"},
					{utils.MandatoryCfg: true, utils.PathCfg: "*cgreq.Subject", utils.TagCfg: "Subject", utils.TypeCfg: "*variable", utils.ValueCfg: "~*req.9"},
					{utils.MandatoryCfg: true, utils.PathCfg: "*cgreq.Destination", utils.TagCfg: "Destination", utils.TypeCfg: "*variable", utils.ValueCfg: "~*req.10"},
					{utils.MandatoryCfg: true, utils.PathCfg: "*cgreq.SetupTime", utils.TagCfg: "SetupTime", utils.TypeCfg: "*variable", utils.ValueCfg: "~*req.11"},
					{utils.MandatoryCfg: true, utils.PathCfg: "*cgreq.AnswerTime", utils.TagCfg: "AnswerTime", utils.TypeCfg: "*variable", utils.ValueCfg: "~*req.12"},
					{utils.MandatoryCfg: true, utils.PathCfg: "*cgreq.Usage", utils.TagCfg: "Usage", utils.TypeCfg: "*variable", utils.ValueCfg: "~*req.13"},
				},
				utils.OptsCfg: map[string]interface{}{
					"csvFieldSeparator":   ",",
					"csvHeaderDefineChar": ":",
					"csvRowLength":        0.,
					"xmlRootPath":         "",
					"partialOrderField":   "~*req.AnswerTime",
					"natsSubject":         "cgrates_cdrs",
				},
			},
			{
				utils.CacheDumpFieldsCfg:     []map[string]interface{}{},
				utils.PartialCommitFieldsCfg: []map[string]interface{}{},
				utils.ConcurrentRequestsCfg:  1024,
				utils.TypeCfg:                "*file_csv",
				utils.FieldsCfg: []map[string]interface{}{
					{utils.MandatoryCfg: true, utils.PathCfg: "*cgreq.ToR", utils.TagCfg: "ToR", utils.TypeCfg: "*variable", utils.ValueCfg: "~*req.2"},
					{utils.MandatoryCfg: true, utils.PathCfg: "*cgreq.OriginID", utils.TagCfg: "OriginID", utils.TypeCfg: "*variable", utils.ValueCfg: "~*req.3"},
					{utils.MandatoryCfg: true, utils.PathCfg: "*cgreq.RequestType", utils.TagCfg: "RequestType", utils.TypeCfg: "*variable", utils.ValueCfg: "~*req.4"},
					{utils.MandatoryCfg: true, utils.PathCfg: "*cgreq.Tenant", utils.TagCfg: "Tenant", utils.TypeCfg: "*variable", utils.ValueCfg: "~*req.6"},
					{utils.MandatoryCfg: true, utils.PathCfg: "*cgreq.Category", utils.TagCfg: "Category", utils.TypeCfg: "*variable", utils.ValueCfg: "~*req.7"},
					{utils.MandatoryCfg: true, utils.PathCfg: "*cgreq.Account", utils.TagCfg: "Account", utils.TypeCfg: "*variable", utils.ValueCfg: "~*req.8"},
					{utils.MandatoryCfg: true, utils.PathCfg: "*cgreq.Subject", utils.TagCfg: "Subject", utils.TypeCfg: "*variable", utils.ValueCfg: "~*req.9"},
					{utils.MandatoryCfg: true, utils.PathCfg: "*cgreq.Destination", utils.TagCfg: "Destination", utils.TypeCfg: "*variable", utils.ValueCfg: "~*req.10"},
					{utils.MandatoryCfg: true, utils.PathCfg: "*cgreq.SetupTime", utils.TagCfg: "SetupTime", utils.TypeCfg: "*variable", utils.ValueCfg: "~*req.11"},
					{utils.MandatoryCfg: true, utils.PathCfg: "*cgreq.AnswerTime", utils.TagCfg: "AnswerTime", utils.TypeCfg: "*variable", utils.ValueCfg: "~*req.12"},
					{utils.MandatoryCfg: true, utils.PathCfg: "*cgreq.Usage", utils.TagCfg: "Usage", utils.TypeCfg: "*variable", utils.ValueCfg: "~*req.13"},
				},
				utils.FiltersCfg:       []string{},
				utils.FlagsCfg:         []string{},
				utils.IDCfg:            "file_reader1",
				utils.ProcessedPathCfg: "/tmp/ers/out",
				utils.RunDelayCfg:      "-1",
				utils.SourcePathCfg:    "/tmp/ers/in",
				utils.TenantCfg:        "~*req.Destination1",
				utils.TimezoneCfg:      "",
				utils.OptsCfg: map[string]interface{}{
					"csvFieldSeparator":   ",",
					"csvHeaderDefineChar": ":",
					"csvRowLength":        0.,
					"xmlRootPath":         "",
					"partialOrderField":   "~*req.AnswerTime",
					"natsSubject":         "cgrates_cdrs",
				},
			},
		},
		utils.PartialCacheTTLCfg:    "1s",
		utils.PartialCacheActionCfg: utils.MetaNone,
		utils.PartialPathCfg:        "",
	}
	if cfg, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cfg.ersCfg.AsMapInterface(utils.InfieldSep); !reflect.DeepEqual(eMap, rcv) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}
}

func TestERSCfgAsMapInterfaceCase2(t *testing.T) {
	cfgJSONStr := `{
	"ers": {
		"enabled": true,
		"sessions_conns":["conn1","conn3"],
		"readers": [
			{
                "id": "file_reader1",
				"run_delay":  "10s",
                "tenant": "~*req.Destination1",
				"type": "*file_csv",
                "flags": ["randomFlag"],
                "filters": ["randomFilter"],
				"source_path": "/tmp/ers/in",
                "partial_record_cache": "1s",
				"processed_path": "/tmp/ers/out",
				"cache_dump_fields": [
                    {"tag": "ToR", "path": "*cgreq.ToR", "type": "*variable", "value": "~*req.2", "mandatory": true}                
				],
				"partial_commit_fields": [{
					"mandatory": true,
					"path": "*cgreq.ToR",
					"tag": "ToR",
					"type": "*variable",
					"value": "~*req.2"
				}],
				"opts":{
					"kafkaGroupID": "test",
				},
			},
		],
	}
}`
	eMap := map[string]interface{}{
		utils.EnabledCfg:       true,
		utils.SessionSConnsCfg: []string{"conn1", "conn3"},
		utils.ReadersCfg: []map[string]interface{}{
			{
				utils.FiltersCfg:             []string{},
				utils.FlagsCfg:               []string{},
				utils.IDCfg:                  "*default",
				utils.ProcessedPathCfg:       "/var/spool/cgrates/ers/out",
				utils.RunDelayCfg:            "0",
				utils.SourcePathCfg:          "/var/spool/cgrates/ers/in",
				utils.TenantCfg:              "",
				utils.TimezoneCfg:            "",
				utils.CacheDumpFieldsCfg:     []map[string]interface{}{},
				utils.PartialCommitFieldsCfg: []map[string]interface{}{},
				utils.ConcurrentRequestsCfg:  1024,
				utils.TypeCfg:                "*none",
				utils.FieldsCfg: []map[string]interface{}{
					{utils.MandatoryCfg: true, utils.PathCfg: "*cgreq.ToR", utils.TagCfg: "ToR", utils.TypeCfg: "*variable", utils.ValueCfg: "~*req.2"},
					{utils.MandatoryCfg: true, utils.PathCfg: "*cgreq.OriginID", utils.TagCfg: "OriginID", utils.TypeCfg: "*variable", utils.ValueCfg: "~*req.3"},
					{utils.MandatoryCfg: true, utils.PathCfg: "*cgreq.RequestType", utils.TagCfg: "RequestType", utils.TypeCfg: "*variable", utils.ValueCfg: "~*req.4"},
					{utils.MandatoryCfg: true, utils.PathCfg: "*cgreq.Tenant", utils.TagCfg: "Tenant", utils.TypeCfg: "*variable", utils.ValueCfg: "~*req.6"},
					{utils.MandatoryCfg: true, utils.PathCfg: "*cgreq.Category", utils.TagCfg: "Category", utils.TypeCfg: "*variable", utils.ValueCfg: "~*req.7"},
					{utils.MandatoryCfg: true, utils.PathCfg: "*cgreq.Account", utils.TagCfg: "Account", utils.TypeCfg: "*variable", utils.ValueCfg: "~*req.8"},
					{utils.MandatoryCfg: true, utils.PathCfg: "*cgreq.Subject", utils.TagCfg: "Subject", utils.TypeCfg: "*variable", utils.ValueCfg: "~*req.9"},
					{utils.MandatoryCfg: true, utils.PathCfg: "*cgreq.Destination", utils.TagCfg: "Destination", utils.TypeCfg: "*variable", utils.ValueCfg: "~*req.10"},
					{utils.MandatoryCfg: true, utils.PathCfg: "*cgreq.SetupTime", utils.TagCfg: "SetupTime", utils.TypeCfg: "*variable", utils.ValueCfg: "~*req.11"},
					{utils.MandatoryCfg: true, utils.PathCfg: "*cgreq.AnswerTime", utils.TagCfg: "AnswerTime", utils.TypeCfg: "*variable", utils.ValueCfg: "~*req.12"},
					{utils.MandatoryCfg: true, utils.PathCfg: "*cgreq.Usage", utils.TagCfg: "Usage", utils.TypeCfg: "*variable", utils.ValueCfg: "~*req.13"},
				},
				utils.OptsCfg: map[string]interface{}{
					"csvFieldSeparator":   ",",
					"csvHeaderDefineChar": ":",
					"csvRowLength":        0.,
					"xmlRootPath":         "",
					"partialOrderField":   "~*req.AnswerTime",
					"natsSubject":         "cgrates_cdrs",
				},
			},
			{
				utils.CacheDumpFieldsCfg: []map[string]interface{}{
					{utils.MandatoryCfg: true, utils.PathCfg: "*cgreq.ToR", utils.TagCfg: "ToR", utils.TypeCfg: "*variable", utils.ValueCfg: "~*req.2"},
				},
				utils.PartialCommitFieldsCfg: []map[string]interface{}{
					{
						"mandatory": true,
						"path":      "*cgreq.ToR",
						"tag":       "ToR",
						"type":      "*variable",
						"value":     "~*req.2",
					},
				},
				utils.ConcurrentRequestsCfg: 1024,
				utils.TypeCfg:               "*file_csv",
				utils.FieldsCfg: []map[string]interface{}{
					{utils.MandatoryCfg: true, utils.PathCfg: "*cgreq.ToR", utils.TagCfg: "ToR", utils.TypeCfg: "*variable", utils.ValueCfg: "~*req.2"},
					{utils.MandatoryCfg: true, utils.PathCfg: "*cgreq.OriginID", utils.TagCfg: "OriginID", utils.TypeCfg: "*variable", utils.ValueCfg: "~*req.3"},
					{utils.MandatoryCfg: true, utils.PathCfg: "*cgreq.RequestType", utils.TagCfg: "RequestType", utils.TypeCfg: "*variable", utils.ValueCfg: "~*req.4"},
					{utils.MandatoryCfg: true, utils.PathCfg: "*cgreq.Tenant", utils.TagCfg: "Tenant", utils.TypeCfg: "*variable", utils.ValueCfg: "~*req.6"},
					{utils.MandatoryCfg: true, utils.PathCfg: "*cgreq.Category", utils.TagCfg: "Category", utils.TypeCfg: "*variable", utils.ValueCfg: "~*req.7"},
					{utils.MandatoryCfg: true, utils.PathCfg: "*cgreq.Account", utils.TagCfg: "Account", utils.TypeCfg: "*variable", utils.ValueCfg: "~*req.8"},
					{utils.MandatoryCfg: true, utils.PathCfg: "*cgreq.Subject", utils.TagCfg: "Subject", utils.TypeCfg: "*variable", utils.ValueCfg: "~*req.9"},
					{utils.MandatoryCfg: true, utils.PathCfg: "*cgreq.Destination", utils.TagCfg: "Destination", utils.TypeCfg: "*variable", utils.ValueCfg: "~*req.10"},
					{utils.MandatoryCfg: true, utils.PathCfg: "*cgreq.SetupTime", utils.TagCfg: "SetupTime", utils.TypeCfg: "*variable", utils.ValueCfg: "~*req.11"},
					{utils.MandatoryCfg: true, utils.PathCfg: "*cgreq.AnswerTime", utils.TagCfg: "AnswerTime", utils.TypeCfg: "*variable", utils.ValueCfg: "~*req.12"},
					{utils.MandatoryCfg: true, utils.PathCfg: "*cgreq.Usage", utils.TagCfg: "Usage", utils.TypeCfg: "*variable", utils.ValueCfg: "~*req.13"},
				},
				utils.FiltersCfg:       []string{"randomFilter"},
				utils.FlagsCfg:         []string{"randomFlag"},
				utils.IDCfg:            "file_reader1",
				utils.ProcessedPathCfg: "/tmp/ers/out",
				utils.RunDelayCfg:      "10s",
				utils.SourcePathCfg:    "/tmp/ers/in",
				utils.TenantCfg:        "~*req.Destination1",
				utils.TimezoneCfg:      "",
				utils.OptsCfg: map[string]interface{}{
					utils.KafkaGroupID:    "test",
					"csvFieldSeparator":   ",",
					"csvHeaderDefineChar": ":",
					"csvRowLength":        0.,
					"xmlRootPath":         "",
					"partialOrderField":   "~*req.AnswerTime",
					"natsSubject":         "cgrates_cdrs",
				},
			},
		},
		utils.PartialCacheTTLCfg:    "1s",
		utils.PartialCacheActionCfg: utils.MetaNone,
		utils.PartialPathCfg:        "",
	}
	if cfg, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cfg.ersCfg.AsMapInterface(utils.InfieldSep); !reflect.DeepEqual(eMap, rcv) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}
}

func TestERsloadFromJsonCfg(t *testing.T) {
	cfgJSON := &ERsJsonCfg{
		Enabled:        utils.BoolPointer(true),
		Sessions_conns: &[]string{"*conn1"},
		Readers: &[]*EventReaderJsonCfg{
			{
				Id:                  utils.StringPointer("file_reader1"),
				Type:                utils.StringPointer(utils.MetaFileCSV),
				Run_delay:           utils.StringPointer("-1"),
				Concurrent_requests: utils.IntPointer(1024),
				Source_path:         utils.StringPointer("/tmp/ers/in"),
				Processed_path:      utils.StringPointer("/tmp/ers/out"),
				Tenant:              nil,
				Timezone:            utils.StringPointer(""),
				Filters:             nil,
				Flags:               &[]string{},
				Fields: &[]*FcTemplateJsonCfg{
					{
						Tag:    utils.StringPointer(utils.CGRID),
						Path:   utils.StringPointer("*exp.CGRID"),
						Type:   utils.StringPointer(utils.MetaVariable),
						Layout: utils.StringPointer("2006-01-02T15:04:05Z07:00"),
					},
					{Tag: utils.StringPointer("CustomTag2"), Path: utils.StringPointer("CustomPath2"), Type: utils.StringPointer(utils.MetaVariable),
						Value: utils.StringPointer("CustomValue2"), Mandatory: utils.BoolPointer(true), Layout: utils.StringPointer("2006-01-02T15:04:05Z07:00")},
				},
			},
		},
	}
	expectedERsCfg := &ERsCfg{
		Enabled:       true,
		SessionSConns: []string{"*conn1"},
		Readers: []*EventReaderCfg{
			{
				ID:             utils.MetaDefault,
				Type:           utils.MetaNone,
				RunDelay:       0,
				ConcurrentReqs: 1024,
				SourcePath:     "/var/spool/cgrates/ers/in",
				ProcessedPath:  "/var/spool/cgrates/ers/out",
				Tenant:         nil,
				Timezone:       utils.EmptyString,
				Filters:        []string{},
				Flags:          utils.FlagsWithParams{},
				Fields: []*FCTemplate{
					{Tag: utils.ToR, Path: utils.MetaCgreq + utils.NestingSep + utils.ToR, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.2", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.OriginID, Path: utils.MetaCgreq + utils.NestingSep + utils.OriginID, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.3", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.RequestType, Path: utils.MetaCgreq + utils.NestingSep + utils.RequestType, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.4", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.Tenant, Path: utils.MetaCgreq + utils.NestingSep + utils.Tenant, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.6", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.Category, Path: utils.MetaCgreq + utils.NestingSep + utils.Category, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.7", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.AccountField, Path: utils.MetaCgreq + utils.NestingSep + utils.AccountField, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.8", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.Subject, Path: utils.MetaCgreq + utils.NestingSep + utils.Subject, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.9", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.Destination, Path: utils.MetaCgreq + utils.NestingSep + utils.Destination, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.10", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.SetupTime, Path: utils.MetaCgreq + utils.NestingSep + utils.SetupTime, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.11", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.AnswerTime, Path: utils.MetaCgreq + utils.NestingSep + utils.AnswerTime, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.12", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.Usage, Path: utils.MetaCgreq + utils.NestingSep + utils.Usage, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.13", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
				},
				CacheDumpFields:     make([]*FCTemplate, 0),
				PartialCommitFields: make([]*FCTemplate, 0),
				Opts: map[string]interface{}{
					"csvFieldSeparator":   ",",
					"csvHeaderDefineChar": ":",
					"csvRowLength":        0.,
					"xmlRootPath":         "",
					"partialOrderField":   "~*req.AnswerTime",
					"natsSubject":         "cgrates_cdrs",
				},
			},
			{
				ID:             "file_reader1",
				Type:           utils.MetaFileCSV,
				RunDelay:       -1,
				ConcurrentReqs: 1024,
				SourcePath:     "/tmp/ers/in",
				ProcessedPath:  "/tmp/ers/out",
				Tenant:         nil,
				Timezone:       utils.EmptyString,
				Filters:        []string{},
				Flags:          utils.FlagsWithParams{},
				Fields: []*FCTemplate{
					{
						Tag:    utils.CGRID,
						Path:   "*exp.CGRID",
						Type:   utils.MetaVariable,
						Layout: time.RFC3339,
					},
					{Tag: "CustomTag2", Path: "CustomPath2", Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("CustomValue2", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
				},
				CacheDumpFields:     make([]*FCTemplate, 0),
				PartialCommitFields: make([]*FCTemplate, 0),
				Opts: map[string]interface{}{
					"csvFieldSeparator":   ",",
					"csvHeaderDefineChar": ":",
					"csvRowLength":        0.,
					"xmlRootPath":         "",
					"partialOrderField":   "~*req.AnswerTime",
					"natsSubject":         "cgrates_cdrs",
				},
			},
		},
		PartialCacheTTL:    time.Second,
		PartialCacheAction: utils.MetaNone,
	}
	for _, profile := range expectedERsCfg.Readers {
		for _, v := range profile.Fields {
			v.ComputePath()
		}
		for _, v := range profile.CacheDumpFields {
			v.ComputePath()
		}
	}
	cfgCgr := NewDefaultCGRConfig()
	if err := cfgCgr.ersCfg.loadFromJSONCfg(cfgJSON, cfgCgr.templates, cfgCgr.generalCfg.RSRSep, cfgCgr.dfltEvRdr); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(cfgCgr.ersCfg, expectedERsCfg) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expectedERsCfg), utils.ToJSON(cgrCfg.ersCfg))
	}
}

func TestGetDefaultExporter(t *testing.T) {
	ees := new(EEsCfg)
	if dft := ees.GetDefaultExporter(); dft != nil {
		t.Fatalf("Expected no default cfg, received: %s", utils.ToJSON(dft))
	}
	cfgCgr := NewDefaultCGRConfig()
	if dft := cfgCgr.EEsCfg().GetDefaultExporter(); dft == nil || dft.ID != utils.MetaDefault {
		t.Fatalf("Unexpected default cfg returned: %s", utils.ToJSON(dft))
	}
}

func TestDiffEventReaderJsonCfg(t *testing.T) {
	var d *EventReaderJsonCfg

	v1 := &EventReaderCfg{
		ID:             "ERS_ID",
		Type:           "xml",
		RunDelay:       1 * time.Second,
		ConcurrentReqs: 2,
		SourcePath:     "/tmp/ers/in",
		ProcessedPath:  "/tmp/ers/out",
		Opts:           map[string]interface{}{},
		Tenant: RSRParsers{
			{
				Rules: "cgrates.org",
			},
		},
		Timezone:            "UTC",
		Filters:             []string{"Filter1"},
		Flags:               utils.FlagsWithParams{},
		Fields:              []*FCTemplate{},
		CacheDumpFields:     []*FCTemplate{},
		PartialCommitFields: []*FCTemplate{},
	}

	v2 := &EventReaderCfg{
		ID:             "ERS_ID2",
		Type:           "json",
		RunDelay:       3 * time.Second,
		ConcurrentReqs: 1,
		SourcePath:     "/var/tmp/ers/in",
		ProcessedPath:  "/var/tmp/ers/out",
		Opts: map[string]interface{}{
			"OPT": "opt",
		},
		Tenant: RSRParsers{
			{
				Rules: "itsyscom.com",
			},
		},
		Timezone: "EEST",
		Filters:  []string{"Filter2"},
		Flags: utils.FlagsWithParams{
			"FLAG1": {
				"PARAM_1": []string{"param1"},
			},
		},
		Fields: []*FCTemplate{
			{
				Type:   utils.MetaVariable,
				Layout: time.RFC3339,
			},
		},
		CacheDumpFields: []*FCTemplate{
			{
				Type:   utils.MetaConstant,
				Layout: time.RFC3339,
			},
		},
		PartialCommitFields: []*FCTemplate{
			{
				Type:   utils.MetaConstant,
				Layout: time.RFC3339,
			},
		},
	}

	expected := &EventReaderJsonCfg{
		Id:                  utils.StringPointer("ERS_ID2"),
		Type:                utils.StringPointer("json"),
		Run_delay:           utils.StringPointer("3s"),
		Concurrent_requests: utils.IntPointer(1),
		Source_path:         utils.StringPointer("/var/tmp/ers/in"),
		Processed_path:      utils.StringPointer("/var/tmp/ers/out"),
		Opts: map[string]interface{}{
			"OPT": "opt",
		},
		Tenant:   utils.StringPointer("itsyscom.com"),
		Timezone: utils.StringPointer("EEST"),
		Filters:  &[]string{"Filter2"},
		Flags:    &[]string{"FLAG1:PARAM_1:param1"},
		Fields: &[]*FcTemplateJsonCfg{
			{
				Type: utils.StringPointer(utils.MetaVariable),
			},
		},
		Cache_dump_fields: &[]*FcTemplateJsonCfg{
			{
				Type: utils.StringPointer(utils.MetaConstant),
			},
		},
		Partial_commit_fields: &[]*FcTemplateJsonCfg{
			{
				Type: utils.StringPointer(utils.MetaConstant),
			},
		},
	}

	rcv := diffEventReaderJsonCfg(d, v1, v2, ";")
	if !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}

	v1 = v2
	d = &EventReaderJsonCfg{
		Fields: &[]*FcTemplateJsonCfg{
			{
				Tag:   utils.StringPointer("ToR"),
				Path:  utils.StringPointer("*cgreq.ToR"),
				Type:  utils.StringPointer(utils.MetaVariable),
				Value: utils.StringPointer("~*req.2"),
			},
		},
		Cache_dump_fields: &[]*FcTemplateJsonCfg{
			{
				Tag:   utils.StringPointer("ToR2"),
				Path:  utils.StringPointer("*cgreq.ToR2"),
				Type:  utils.StringPointer(utils.MetaConstant),
				Value: utils.StringPointer("~*req.3"),
			},
		},
		Partial_commit_fields: &[]*FcTemplateJsonCfg{
			{
				Tag:   utils.StringPointer("ToR3"),
				Path:  utils.StringPointer("*cgreq.ToR3"),
				Type:  utils.StringPointer(utils.MetaTemplate),
				Value: utils.StringPointer("~*req.4"),
			},
		},
	}

	expected = &EventReaderJsonCfg{
		Opts: map[string]interface{}{},
		Fields: &[]*FcTemplateJsonCfg{
			{
				Tag:   utils.StringPointer("ToR"),
				Path:  utils.StringPointer("*cgreq.ToR"),
				Type:  utils.StringPointer(utils.MetaVariable),
				Value: utils.StringPointer("~*req.2"),
			},
		},
		Cache_dump_fields: &[]*FcTemplateJsonCfg{
			{
				Tag:   utils.StringPointer("ToR2"),
				Path:  utils.StringPointer("*cgreq.ToR2"),
				Type:  utils.StringPointer(utils.MetaConstant),
				Value: utils.StringPointer("~*req.3"),
			},
		},
		Partial_commit_fields: &[]*FcTemplateJsonCfg{
			{
				Tag:   utils.StringPointer("ToR3"),
				Path:  utils.StringPointer("*cgreq.ToR3"),
				Type:  utils.StringPointer(utils.MetaTemplate),
				Value: utils.StringPointer("~*req.4"),
			},
		},
	}
	rcv = diffEventReaderJsonCfg(d, v1, v2, ";")
	if !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}

func TestGetEventReaderJsonCfg(t *testing.T) {
	d := []*EventReaderJsonCfg{
		{
			Id: utils.StringPointer("ERS_ID"),
		},
	}

	expected := &EventReaderJsonCfg{
		Id: utils.StringPointer("ERS_ID"),
	}

	rcv, idx := getEventReaderJsonCfg(d, "ERS_ID")
	if !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	} else if idx != 0 {
		t.Errorf("Expected %v \n but received \n %v", 0, idx)
	}

	d = []*EventReaderJsonCfg{
		{
			Id: nil,
		},
	}
	rcv, idx = getEventReaderJsonCfg(d, "ERS_ID")
	if rcv != nil {
		t.Error("Received value should be null")
	} else if idx != -1 {
		t.Errorf("Expected %v \n but received \n %v", -1, idx)
	}
}

func TestGetEventReaderCfg(t *testing.T) {
	d := []*EventReaderCfg{
		{
			ID: "ERS_ID",
		},
	}

	expected := &EventReaderCfg{
		ID: "ERS_ID",
	}

	rcv := getEventReaderCfg(d, "ERS_ID")
	if !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}

	d = []*EventReaderCfg{
		{
			ID: "ERS_ID2",
		},
	}

	rcv = getEventReaderCfg(d, "ERS_ID")
	if !reflect.DeepEqual(rcv, new(EventReaderCfg)) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(new(EventReaderCfg)), utils.ToJSON(rcv))
	}
}

func TestDiffEventReadersJsonCfg(t *testing.T) {
	var d *[]*EventReaderJsonCfg

	v1 := []*EventReaderCfg{
		{
			ID:             "ERS_ID",
			Type:           "xml",
			RunDelay:       1 * time.Second,
			ConcurrentReqs: 2,
			SourcePath:     "/tmp/ers/in",
			ProcessedPath:  "/tmp/ers/out",
			Opts:           map[string]interface{}{},
			Tenant: RSRParsers{
				{
					Rules: "cgrates.org",
				},
			},
			Timezone:        "UTC",
			Filters:         []string{"Filter1"},
			Flags:           utils.FlagsWithParams{},
			Fields:          []*FCTemplate{},
			CacheDumpFields: []*FCTemplate{},
		},
	}

	v2 := []*EventReaderCfg{
		{
			ID:             "ERS_ID2",
			Type:           "json",
			RunDelay:       3 * time.Second,
			ConcurrentReqs: 1,
			SourcePath:     "/var/tmp/ers/in",
			ProcessedPath:  "/var/tmp/ers/out",
			Opts: map[string]interface{}{
				"OPT": "opt",
			},
			Tenant: RSRParsers{
				{
					Rules: "itsyscom.com",
				},
			},
			Timezone: "EEST",
			Filters:  []string{"Filter2"},
			Flags: utils.FlagsWithParams{
				"FLAG1": {
					"PARAM_1": []string{"param1"},
				},
			},
			Fields: []*FCTemplate{
				{
					Type: "*string",
				},
			},
			CacheDumpFields: []*FCTemplate{
				{
					Type: "*string",
				},
			},
		},
	}

	expected := &[]*EventReaderJsonCfg{
		{
			Id:                  utils.StringPointer("ERS_ID2"),
			Type:                utils.StringPointer("json"),
			Run_delay:           utils.StringPointer("3s"),
			Concurrent_requests: utils.IntPointer(1),
			Source_path:         utils.StringPointer("/var/tmp/ers/in"),
			Processed_path:      utils.StringPointer("/var/tmp/ers/out"),
			Opts: map[string]interface{}{
				"OPT": "opt",
			},
			Tenant:   utils.StringPointer("itsyscom.com"),
			Timezone: utils.StringPointer("EEST"),
			Filters:  &[]string{"Filter2"},
			Flags:    &[]string{"FLAG1:PARAM_1:param1"},
			Fields: &[]*FcTemplateJsonCfg{
				{
					Type:   utils.StringPointer("*string"),
					Layout: utils.StringPointer(""),
				},
			},
			Cache_dump_fields: &[]*FcTemplateJsonCfg{
				{
					Type:   utils.StringPointer("*string"),
					Layout: utils.StringPointer(""),
				},
			},
		},
	}

	rcv := diffEventReadersJsonCfg(d, v1, v2, ";")
	if !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}

	v1 = v2
	expected = &[]*EventReaderJsonCfg{
		{
			Opts: map[string]interface{}{},
		},
	}
	rcv = diffEventReadersJsonCfg(d, v1, v2, ";")
	if !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}

	(*expected)[0].Id = utils.StringPointer("ERS_ID2")
	d = &[]*EventReaderJsonCfg{
		{
			Id: utils.StringPointer("ERS_ID2"),
		},
	}

	rcv = diffEventReadersJsonCfg(d, v1, v2, ";")
	if !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}

func TestDiffERsJsonCfg(t *testing.T) {
	var d *ERsJsonCfg

	v1 := &ERsCfg{
		Enabled:       false,
		SessionSConns: []string{"*birpc"},
		Readers: []*EventReaderCfg{
			{
				ID: "ERS_ID",
			},
		},
		PartialCacheTTL:    24 * time.Hour,
		PartialCacheAction: "partial_cache_action",
		PartialPath:        "/partial/path",
	}

	v2 := &ERsCfg{
		Enabled:       true,
		SessionSConns: []string{"*localhost"},
		Readers: []*EventReaderCfg{
			{
				ID: "ERS_ID2",
			},
		},
		PartialCacheTTL:    12 * time.Hour,
		PartialCacheAction: "partial_cache_action2",
		PartialPath:        "/partial/path/2",
	}

	expected := &ERsJsonCfg{
		Enabled:        utils.BoolPointer(true),
		Sessions_conns: &[]string{"*localhost"},
		Readers: &[]*EventReaderJsonCfg{
			{
				Id:   utils.StringPointer("ERS_ID2"),
				Opts: map[string]interface{}{},
			},
		},
		Partial_cache_ttl:    utils.StringPointer("12h0m0s"),
		Partial_cache_action: utils.StringPointer("partial_cache_action2"),
		Partial_path:         utils.StringPointer("/partial/path/2"),
	}

	rcv := diffERsJsonCfg(d, v1, v2, ";")
	if !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Expected %+v \n but received \n %+v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}
