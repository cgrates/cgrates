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
			"type": "*fileCSV",
			"flags": ["*dryRun"],
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
				Opts: &EventReaderOpts{
					CSVFieldSeparator:   utils.StringPointer(","),
					CSVHeaderDefineChar: utils.StringPointer(":"),
					CSVRowLength:        utils.IntPointer(0),
					XMLRootPath:         utils.StringPointer(utils.EmptyString),
					PartialOrderField:   utils.StringPointer("~*req.AnswerTime"),
					PartialCacheAction:  utils.StringPointer(utils.MetaNone),
					NATSSubject:         utils.StringPointer("cgrates_cdrs"),
				},
			},
			{
				ID:             "file_reader1",
				Type:           "*fileCSV",
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
				Opts: &EventReaderOpts{
					CSVFieldSeparator:   utils.StringPointer(","),
					CSVHeaderDefineChar: utils.StringPointer(":"),
					CSVRowLength:        utils.IntPointer(0),
					XMLRootPath:         utils.StringPointer(utils.EmptyString),
					PartialOrderField:   utils.StringPointer("~*req.AnswerTime"),
					PartialCacheAction:  utils.StringPointer(utils.MetaNone),
					NATSSubject:         utils.StringPointer("cgrates_cdrs"),
				},
			},
		},
		PartialCacheTTL: time.Second,
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
	if err = jsoncfg.ersCfg.loadFromJSONCfg(cfgJSON, jsoncfg.templates, jsoncfg.generalCfg.RSRSep); err == nil || err.Error() != expected {
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
	if err = jsoncfg.ersCfg.loadFromJSONCfg(cfgJSON, jsoncfg.templates, jsoncfg.generalCfg.RSRSep); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestEventReaderloadFromJsonCase2(t *testing.T) {
	cfgJSON := &ERsJsonCfg{
		Partial_cache_ttl: utils.StringPointer("1ss"),
	}
	expected := `time: unknown unit "ss" in duration "1ss"`
	jsonCfg := NewDefaultCGRConfig()
	if err = jsonCfg.ersCfg.loadFromJSONCfg(cfgJSON, jsonCfg.templates, jsonCfg.generalCfg.RSRSep); err == nil || err.Error() != expected {
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
				Opts: &EventReaderOpts{
					CSVFieldSeparator:   utils.StringPointer(","),
					CSVHeaderDefineChar: utils.StringPointer(":"),
					CSVRowLength:        utils.IntPointer(0),
					XMLRootPath:         utils.StringPointer(utils.EmptyString),
					PartialOrderField:   utils.StringPointer("~*req.AnswerTime"),
					PartialCacheAction:  utils.StringPointer(utils.MetaNone),
					NATSSubject:         utils.StringPointer("cgrates_cdrs"),
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
				Opts: &EventReaderOpts{
					CSVFieldSeparator:   utils.StringPointer(","),
					CSVHeaderDefineChar: utils.StringPointer(":"),
					CSVRowLength:        utils.IntPointer(0),
					XMLRootPath:         utils.StringPointer(utils.EmptyString),
					PartialOrderField:   utils.StringPointer("~*req.AnswerTime"),
					PartialCacheAction:  utils.StringPointer(utils.MetaNone),
					NATSSubject:         utils.StringPointer("cgrates_cdrs"),
				},
			},
		},
		PartialCacheTTL: time.Second,
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
			"type": "*fileCSV",
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
	if err = jsonCfg.ersCfg.loadFromJSONCfg(cfgJSON, jsonCfg.templates, jsonCfg.generalCfg.RSRSep); err != nil {
		t.Error(err)
	}
	cfgJSON = nil
	if err = jsonCfg.ersCfg.loadFromJSONCfg(cfgJSON, jsonCfg.templates, jsonCfg.generalCfg.RSRSep); err != nil {
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
	if err = jsonCfg.ersCfg.loadFromJSONCfg(cfgJSON, jsonCfg.templates, jsonCfg.generalCfg.RSRSep); err == nil || err.Error() != expected {
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
	if err = jsonCfg.ersCfg.loadFromJSONCfg(cfgJSON, jsonCfg.templates, jsonCfg.generalCfg.RSRSep); err == nil || err.Error() != expected {
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
	if err = jsonCfg.ersCfg.loadFromJSONCfg(cfgJSON, jsonCfg.templates, jsonCfg.generalCfg.RSRSep); err == nil || err.Error() != expected {
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
				Opts: &EventReaderOpts{
					CSVFieldSeparator:   utils.StringPointer(","),
					CSVHeaderDefineChar: utils.StringPointer(":"),
					CSVRowLength:        utils.IntPointer(0),
					XMLRootPath:         utils.StringPointer(utils.EmptyString),
					PartialOrderField:   utils.StringPointer("~*req.AnswerTime"),
					PartialCacheAction:  utils.StringPointer(utils.MetaNone),
					NATSSubject:         utils.StringPointer("cgrates_cdrs"),
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
						Tag:    utils.MetaOriginID,
						Path:   "*opts.*originID",
						Type:   utils.MetaVariable,
						Layout: time.RFC3339,
					},
				},
				CacheDumpFields:     make([]*FCTemplate, 0),
				PartialCommitFields: make([]*FCTemplate, 0),
				Opts: &EventReaderOpts{
					CSVFieldSeparator:   utils.StringPointer(","),
					CSVHeaderDefineChar: utils.StringPointer(":"),
					CSVRowLength:        utils.IntPointer(0),
					XMLRootPath:         utils.StringPointer(utils.EmptyString),
					PartialOrderField:   utils.StringPointer("~*req.AnswerTime"),
					PartialCacheAction:  utils.StringPointer(utils.MetaNone),
					NATSSubject:         utils.StringPointer("cgrates_cdrs"),
				},
			},
		},

		PartialCacheTTL: time.Second,
	}
	msgTemplates := map[string][]*FCTemplate{
		"randomTemplate": {
			{
				Tag:    utils.MetaOriginID,
				Path:   "*opts.*originID",
				Type:   utils.MetaVariable,
				Layout: time.RFC3339,
			},
		},
	}
	for _, v := range expectedERsCfg.Readers[0].Fields {
		v.ComputePath()
	}
	jsonCfg := NewDefaultCGRConfig()
	if err = jsonCfg.ersCfg.loadFromJSONCfg(cfgJSON, msgTemplates, jsonCfg.generalCfg.RSRSep); err != nil {
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
				Opts: &EventReaderOpts{
					CSVFieldSeparator:   utils.StringPointer(","),
					CSVHeaderDefineChar: utils.StringPointer(":"),
					CSVRowLength:        utils.IntPointer(0),
					XMLRootPath:         utils.StringPointer(utils.EmptyString),
					PartialOrderField:   utils.StringPointer("~*req.AnswerTime"),
					PartialCacheAction:  utils.StringPointer(utils.MetaNone),
					NATSSubject:         utils.StringPointer("cgrates_cdrs"),
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
				Opts: &EventReaderOpts{
					CSVFieldSeparator:   utils.StringPointer(","),
					CSVHeaderDefineChar: utils.StringPointer(":"),
					CSVRowLength:        utils.IntPointer(0),
					XMLRootPath:         utils.StringPointer(utils.EmptyString),
					PartialOrderField:   utils.StringPointer("~*req.AnswerTime"),
					PartialCacheAction:  utils.StringPointer(utils.MetaNone),
					NATSSubject:         utils.StringPointer("cgrates_cdrs"),
				},
			},
		},

		PartialCacheTTL: time.Second,
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
	if err = jsonCfg.ersCfg.loadFromJSONCfg(cfgJSON, msgTemplates, jsonCfg.generalCfg.RSRSep); err != nil {
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
	if err = jsonCfg.ersCfg.loadFromJSONCfg(cfgJSON, jsonCfg.templates, jsonCfg.generalCfg.RSRSep); err == nil || err.Error() != expected {
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
				Opts: &EventReaderOpts{
					CSVFieldSeparator:   utils.StringPointer(","),
					CSVHeaderDefineChar: utils.StringPointer(":"),
					CSVRowLength:        utils.IntPointer(0),
					XMLRootPath:         utils.StringPointer(utils.EmptyString),
					PartialOrderField:   utils.StringPointer("~*req.AnswerTime"),
					PartialCacheAction:  utils.StringPointer(utils.MetaNone),
					NATSSubject:         utils.StringPointer("cgrates_cdrs"),
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
				Opts: &EventReaderOpts{
					CSVFieldSeparator:   utils.StringPointer(","),
					CSVHeaderDefineChar: utils.StringPointer(":"),
					CSVRowLength:        utils.IntPointer(0),
					XMLRootPath:         utils.StringPointer(utils.EmptyString),
					PartialOrderField:   utils.StringPointer("~*req.AnswerTime"),
					PartialCacheAction:  utils.StringPointer(utils.MetaNone),
					NATSSubject:         utils.StringPointer("cgrates_cdrs"),
				},
			},
		},

		PartialCacheTTL: time.Second,
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
			"type": "*fileCSV",
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
			"type": "*fileCSV",
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
				"type": "*fileCSV",
				"source_path": "/tmp/ers/in",
				"processed_path": "/tmp/ers/out",
				"cache_dump_fields": [],
			},
		],
	}
}`
	eMap := map[string]any{
		utils.EnabledCfg:       true,
		utils.SessionSConnsCfg: []string{"conn1", "conn3"},
		utils.ReadersCfg: []map[string]any{
			{
				utils.FiltersCfg:             []string{},
				utils.FlagsCfg:               []string{},
				utils.IDCfg:                  "*default",
				utils.ProcessedPathCfg:       "/var/spool/cgrates/ers/out",
				utils.RunDelayCfg:            "0",
				utils.SourcePathCfg:          "/var/spool/cgrates/ers/in",
				utils.TenantCfg:              "",
				utils.TimezoneCfg:            "",
				utils.CacheDumpFieldsCfg:     []map[string]any{},
				utils.PartialCommitFieldsCfg: []map[string]any{},
				utils.ConcurrentRequestsCfg:  1024,
				utils.TypeCfg:                "*none",
				utils.FieldsCfg: []map[string]any{
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
				utils.OptsCfg: map[string]any{
					utils.CSVFieldSepOpt:        ",",
					utils.HeaderDefineCharOpt:   ":",
					utils.CSVRowLengthOpt:       0,
					utils.XMLRootPathOpt:        "",
					utils.PartialOrderFieldOpt:  "~*req.AnswerTime",
					utils.PartialCacheActionOpt: utils.MetaNone,
					utils.NatsSubject:           "cgrates_cdrs",
				},
			},
			{
				utils.CacheDumpFieldsCfg:     []map[string]any{},
				utils.PartialCommitFieldsCfg: []map[string]any{},
				utils.ConcurrentRequestsCfg:  1024,
				utils.TypeCfg:                "*fileCSV",
				utils.FieldsCfg: []map[string]any{
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
				utils.OptsCfg: map[string]any{
					utils.CSVFieldSepOpt:        ",",
					utils.HeaderDefineCharOpt:   ":",
					utils.CSVRowLengthOpt:       0,
					utils.XMLRootPathOpt:        "",
					utils.PartialOrderFieldOpt:  "~*req.AnswerTime",
					utils.PartialCacheActionOpt: utils.MetaNone,
					utils.NatsSubject:           "cgrates_cdrs",
				},
			},
		},
		utils.PartialCacheTTLCfg: "1s",
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
				"type": "*fileCSV",
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
	eMap := map[string]any{
		utils.EnabledCfg:       true,
		utils.SessionSConnsCfg: []string{"conn1", "conn3"},
		utils.ReadersCfg: []map[string]any{
			{
				utils.FiltersCfg:             []string{},
				utils.FlagsCfg:               []string{},
				utils.IDCfg:                  "*default",
				utils.ProcessedPathCfg:       "/var/spool/cgrates/ers/out",
				utils.RunDelayCfg:            "0",
				utils.SourcePathCfg:          "/var/spool/cgrates/ers/in",
				utils.TenantCfg:              "",
				utils.TimezoneCfg:            "",
				utils.CacheDumpFieldsCfg:     []map[string]any{},
				utils.PartialCommitFieldsCfg: []map[string]any{},
				utils.ConcurrentRequestsCfg:  1024,
				utils.TypeCfg:                "*none",
				utils.FieldsCfg: []map[string]any{
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
				utils.OptsCfg: map[string]any{
					utils.CSVFieldSepOpt:        ",",
					utils.HeaderDefineCharOpt:   ":",
					utils.CSVRowLengthOpt:       0,
					utils.XMLRootPathOpt:        "",
					utils.PartialOrderFieldOpt:  "~*req.AnswerTime",
					utils.PartialCacheActionOpt: utils.MetaNone,
					utils.NatsSubject:           "cgrates_cdrs",
				},
			},
			{
				utils.CacheDumpFieldsCfg: []map[string]any{
					{utils.MandatoryCfg: true, utils.PathCfg: "*cgreq.ToR", utils.TagCfg: "ToR", utils.TypeCfg: "*variable", utils.ValueCfg: "~*req.2"},
				},
				utils.PartialCommitFieldsCfg: []map[string]any{
					{
						"mandatory": true,
						"path":      "*cgreq.ToR",
						"tag":       "ToR",
						"type":      "*variable",
						"value":     "~*req.2",
					},
				},
				utils.ConcurrentRequestsCfg: 1024,
				utils.TypeCfg:               "*fileCSV",
				utils.FieldsCfg: []map[string]any{
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
				utils.OptsCfg: map[string]any{
					utils.KafkaGroupID:          "test",
					utils.CSVFieldSepOpt:        ",",
					utils.HeaderDefineCharOpt:   ":",
					utils.CSVRowLengthOpt:       0,
					utils.XMLRootPathOpt:        "",
					utils.PartialOrderFieldOpt:  "~*req.AnswerTime",
					utils.PartialCacheActionOpt: utils.MetaNone,
					utils.NatsSubject:           "cgrates_cdrs",
				},
			},
		},
		utils.PartialCacheTTLCfg: "1s",
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
						Tag:    utils.StringPointer(utils.MetaOriginID),
						Path:   utils.StringPointer("*opts.*originID"),
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
				Opts: &EventReaderOpts{
					CSVFieldSeparator:   utils.StringPointer(","),
					CSVHeaderDefineChar: utils.StringPointer(":"),
					CSVRowLength:        utils.IntPointer(0),
					XMLRootPath:         utils.StringPointer(utils.EmptyString),
					PartialOrderField:   utils.StringPointer("~*req.AnswerTime"),
					PartialCacheAction:  utils.StringPointer(utils.MetaNone),
					NATSSubject:         utils.StringPointer("cgrates_cdrs"),
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
						Tag:    utils.MetaOriginID,
						Path:   "*opts.*originID",
						Type:   utils.MetaVariable,
						Layout: time.RFC3339,
					},
					{Tag: "CustomTag2", Path: "CustomPath2", Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("CustomValue2", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
				},
				CacheDumpFields:     make([]*FCTemplate, 0),
				PartialCommitFields: make([]*FCTemplate, 0),
				Opts: &EventReaderOpts{
					CSVFieldSeparator:   utils.StringPointer(","),
					CSVHeaderDefineChar: utils.StringPointer(":"),
					CSVRowLength:        utils.IntPointer(0),
					XMLRootPath:         utils.StringPointer(utils.EmptyString),
					PartialOrderField:   utils.StringPointer("~*req.AnswerTime"),
					PartialCacheAction:  utils.StringPointer(utils.MetaNone),
					NATSSubject:         utils.StringPointer("cgrates_cdrs"),
				},
			},
		},
		PartialCacheTTL: time.Second,
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
	if err := cfgCgr.ersCfg.loadFromJSONCfg(cfgJSON, cfgCgr.templates, cfgCgr.generalCfg.RSRSep); err != nil {
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
		Opts:           &EventReaderOpts{},
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
		Opts: &EventReaderOpts{
			CSVRowLength: utils.IntPointer(5),
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
		Opts: &EventReaderOptsJson{
			CSVRowLength: utils.IntPointer(5),
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
		Opts: &EventReaderOptsJson{},
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
			ID:   "ERS_ID",
			Opts: &EventReaderOpts{},
		},
	}

	expected := &EventReaderCfg{
		ID:   "ERS_ID",
		Opts: &EventReaderOpts{},
	}

	rcv := getEventReaderCfg(d, "ERS_ID")
	if !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}

	d = []*EventReaderCfg{
		{
			ID:   "ERS_ID2",
			Opts: &EventReaderOpts{},
		},
	}

	rcv = getEventReaderCfg(d, "ERS_ID")
	expected = &EventReaderCfg{
		Opts: &EventReaderOpts{},
	}
	if !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
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
			Opts:           &EventReaderOpts{},
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
			Opts: &EventReaderOpts{
				CSVRowLength: utils.IntPointer(5),
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
			Opts: &EventReaderOptsJson{
				CSVRowLength: utils.IntPointer(5),
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
			Opts: &EventReaderOptsJson{},
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
				ID:   "ERS_ID",
				Opts: &EventReaderOpts{},
			},
		},
		PartialCacheTTL: 24 * time.Hour,
	}

	v2 := &ERsCfg{
		Enabled:       true,
		SessionSConns: []string{"*localhost"},
		Readers: []*EventReaderCfg{
			{
				ID:   "ERS_ID2",
				Opts: &EventReaderOpts{},
			},
		},
		PartialCacheTTL: 12 * time.Hour,
	}

	expected := &ERsJsonCfg{
		Enabled:        utils.BoolPointer(true),
		Sessions_conns: &[]string{"*localhost"},
		Readers: &[]*EventReaderJsonCfg{
			{
				Id:   utils.StringPointer("ERS_ID2"),
				Opts: &EventReaderOptsJson{},
			},
		},
		Partial_cache_ttl: utils.StringPointer("12h0m0s"),
	}

	rcv := diffERsJsonCfg(d, v1, v2, ";")
	if !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Expected %+v \n but received \n %+v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}

func TestErSCloneSection(t *testing.T) {
	erSCfg := &ERsCfg{
		Enabled:       false,
		SessionSConns: []string{"*birpc"},
		Readers: []*EventReaderCfg{
			{
				ID:   "ERS_ID",
				Opts: &EventReaderOpts{},
			},
		},
		PartialCacheTTL: 24 * time.Hour,
	}

	exp := &ERsCfg{
		Enabled:       false,
		SessionSConns: []string{"*birpc"},
		Readers: []*EventReaderCfg{
			{
				ID:   "ERS_ID",
				Opts: &EventReaderOpts{},
			},
		},
		PartialCacheTTL: 24 * time.Hour,
	}

	rcv := erSCfg.CloneSection()
	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("Expected %+v \n but received \n %+v", utils.ToJSON(exp), utils.ToJSON(rcv))
	}
}

func TestERsLoadFromJSONCfg(t *testing.T) {
	erOpts := &EventReaderOpts{}

	erJson := &EventReaderOptsJson{
		PartialPath:                       utils.StringPointer("/tmp/path"),
		PartialCacheAction:                utils.StringPointer("partial_cache_action"),
		PartialOrderField:                 utils.StringPointer("partial_order_field"),
		PartialCSVFieldSeparator:          utils.StringPointer(";"),
		CSVRowLength:                      utils.IntPointer(2),
		CSVFieldSeparator:                 utils.StringPointer(","),
		CSVHeaderDefineChar:               utils.StringPointer("header_define_char"),
		CSVLazyQuotes:                     utils.BoolPointer(false),
		XMLRootPath:                       utils.StringPointer("xml_root_path"),
		AMQPQueueID:                       utils.StringPointer("queue_id"),
		AMQPQueueIDProcessed:              utils.StringPointer("queue_id_processed"),
		AMQPConsumerTag:                   utils.StringPointer("consumer_tag"),
		AMQPExchange:                      utils.StringPointer("exchange"),
		AMQPExchangeType:                  utils.StringPointer("exchange_type"),
		AMQPRoutingKey:                    utils.StringPointer("routing_key"),
		AMQPExchangeProcessed:             utils.StringPointer("exchange_processed"),
		AMQPExchangeTypeProcessed:         utils.StringPointer("excange_type_processed"),
		AMQPRoutingKeyProcessed:           utils.StringPointer("routing_key_processed"),
		KafkaTopic:                        utils.StringPointer("topic"),
		KafkaGroupID:                      utils.StringPointer("group_id"),
		KafkaMaxWait:                      utils.StringPointer("2s"),
		KafkaTopicProcessed:               utils.StringPointer("topic_processed"),
		SQLDBName:                         utils.StringPointer("cgrates"),
		SQLTableName:                      utils.StringPointer("cgrates_t1"),
		PgSSLMode:                         utils.StringPointer("ssl_mode"),
		SQLDBNameProcessed:                utils.StringPointer("cgrates_processed"),
		SQLTableNameProcessed:             utils.StringPointer("cgrates_t1_processed"),
		PgSSLModeProcessed:                utils.StringPointer("ssl_mode_processed"),
		AWSRegion:                         utils.StringPointer("us-west"),
		AWSKey:                            utils.StringPointer("aws_key"),
		AWSSecret:                         utils.StringPointer("aws_secret"),
		AWSToken:                          utils.StringPointer("aws_token"),
		AWSRegionProcessed:                utils.StringPointer("region_processed"),
		AWSKeyProcessed:                   utils.StringPointer("aws_key_processed"),
		AWSSecretProcessed:                utils.StringPointer("aws_secret_processed"),
		AWSTokenProcessed:                 utils.StringPointer("aws_token_processed"),
		SQSQueueID:                        utils.StringPointer("queue_id"),
		SQSQueueIDProcessed:               utils.StringPointer("queue_id_processed"),
		S3BucketID:                        utils.StringPointer("bucket_id"),
		S3FolderPathProcessed:             utils.StringPointer("folder_path_processed"),
		S3BucketIDProcessed:               utils.StringPointer("bucket_id_processed"),
		NATSJetStream:                     utils.BoolPointer(false),
		NATSConsumerName:                  utils.StringPointer("consumer_name"),
		NATSSubject:                       utils.StringPointer("subject"),
		NATSQueueID:                       utils.StringPointer("queue_id"),
		NATSJWTFile:                       utils.StringPointer("jsw_file"),
		NATSSeedFile:                      utils.StringPointer("seed_file"),
		NATSCertificateAuthority:          utils.StringPointer("ca"),
		NATSClientCertificate:             utils.StringPointer("cc"),
		NATSClientKey:                     utils.StringPointer("ck"),
		NATSJetStreamMaxWait:              utils.StringPointer("2s"),
		NATSJetStreamProcessed:            utils.BoolPointer(false),
		NATSSubjectProcessed:              utils.StringPointer("subject_processed"),
		NATSJWTFileProcessed:              utils.StringPointer("jwt_file_processed"),
		NATSSeedFileProcessed:             utils.StringPointer("seed_file_processed"),
		NATSCertificateAuthorityProcessed: utils.StringPointer("ca_processed"),
		NATSClientCertificateProcessed:    utils.StringPointer("cc_processed"),
		NATSClientKeyProcessed:            utils.StringPointer("ck_processed"),
		NATSJetStreamMaxWaitProcessed:     utils.StringPointer("2s"),
		KafkaTLS:                          utils.BoolPointer(true),
		KafkaCAPath:                       utils.StringPointer("/tmp/path"),
		KafkaSkipTLSVerify:                utils.BoolPointer(true),
		KafkaTLSProcessed:                 utils.BoolPointer(true),
		KafkaCAPathProcessed:              utils.StringPointer("/tmp/path"),
		KafkaSkipTLSVerifyProcessed:       utils.BoolPointer(true),
	}

	exp := &EventReaderOpts{
		PartialPath:                       utils.StringPointer("/tmp/path"),
		PartialCacheAction:                utils.StringPointer("partial_cache_action"),
		PartialOrderField:                 utils.StringPointer("partial_order_field"),
		PartialCSVFieldSeparator:          utils.StringPointer(";"),
		CSVRowLength:                      utils.IntPointer(2),
		CSVFieldSeparator:                 utils.StringPointer(","),
		CSVHeaderDefineChar:               utils.StringPointer("header_define_char"),
		CSVLazyQuotes:                     utils.BoolPointer(false),
		XMLRootPath:                       utils.StringPointer("xml_root_path"),
		AMQPQueueID:                       utils.StringPointer("queue_id"),
		AMQPQueueIDProcessed:              utils.StringPointer("queue_id_processed"),
		AMQPConsumerTag:                   utils.StringPointer("consumer_tag"),
		AMQPExchange:                      utils.StringPointer("exchange"),
		AMQPExchangeType:                  utils.StringPointer("exchange_type"),
		AMQPRoutingKey:                    utils.StringPointer("routing_key"),
		AMQPExchangeProcessed:             utils.StringPointer("exchange_processed"),
		AMQPExchangeTypeProcessed:         utils.StringPointer("excange_type_processed"),
		AMQPRoutingKeyProcessed:           utils.StringPointer("routing_key_processed"),
		KafkaTopic:                        utils.StringPointer("topic"),
		KafkaGroupID:                      utils.StringPointer("group_id"),
		KafkaMaxWait:                      utils.DurationPointer(2 * time.Second),
		KafkaTopicProcessed:               utils.StringPointer("topic_processed"),
		SQLDBName:                         utils.StringPointer("cgrates"),
		SQLTableName:                      utils.StringPointer("cgrates_t1"),
		PgSSLMode:                         utils.StringPointer("ssl_mode"),
		SQLDBNameProcessed:                utils.StringPointer("cgrates_processed"),
		SQLTableNameProcessed:             utils.StringPointer("cgrates_t1_processed"),
		PgSSLModeProcessed:                utils.StringPointer("ssl_mode_processed"),
		AWSRegion:                         utils.StringPointer("us-west"),
		AWSKey:                            utils.StringPointer("aws_key"),
		AWSSecret:                         utils.StringPointer("aws_secret"),
		AWSToken:                          utils.StringPointer("aws_token"),
		AWSRegionProcessed:                utils.StringPointer("region_processed"),
		AWSKeyProcessed:                   utils.StringPointer("aws_key_processed"),
		AWSSecretProcessed:                utils.StringPointer("aws_secret_processed"),
		AWSTokenProcessed:                 utils.StringPointer("aws_token_processed"),
		SQSQueueID:                        utils.StringPointer("queue_id"),
		SQSQueueIDProcessed:               utils.StringPointer("queue_id_processed"),
		S3BucketID:                        utils.StringPointer("bucket_id"),
		S3FolderPathProcessed:             utils.StringPointer("folder_path_processed"),
		S3BucketIDProcessed:               utils.StringPointer("bucket_id_processed"),
		NATSJetStream:                     utils.BoolPointer(false),
		NATSConsumerName:                  utils.StringPointer("consumer_name"),
		NATSSubject:                       utils.StringPointer("subject"),
		NATSQueueID:                       utils.StringPointer("queue_id"),
		NATSJWTFile:                       utils.StringPointer("jsw_file"),
		NATSSeedFile:                      utils.StringPointer("seed_file"),
		NATSCertificateAuthority:          utils.StringPointer("ca"),
		NATSClientCertificate:             utils.StringPointer("cc"),
		NATSClientKey:                     utils.StringPointer("ck"),
		NATSJetStreamMaxWait:              utils.DurationPointer(2 * time.Second),
		NATSJetStreamProcessed:            utils.BoolPointer(false),
		NATSSubjectProcessed:              utils.StringPointer("subject_processed"),
		NATSJWTFileProcessed:              utils.StringPointer("jwt_file_processed"),
		NATSSeedFileProcessed:             utils.StringPointer("seed_file_processed"),
		NATSCertificateAuthorityProcessed: utils.StringPointer("ca_processed"),
		NATSClientCertificateProcessed:    utils.StringPointer("cc_processed"),
		NATSClientKeyProcessed:            utils.StringPointer("ck_processed"),
		NATSJetStreamMaxWaitProcessed:     utils.DurationPointer(2 * time.Second),
		KafkaTLS:                          utils.BoolPointer(true),
		KafkaCAPath:                       utils.StringPointer("/tmp/path"),
		KafkaSkipTLSVerify:                utils.BoolPointer(true),
		KafkaTLSProcessed:                 utils.BoolPointer(true),
		KafkaCAPathProcessed:              utils.StringPointer("/tmp/path"),
		KafkaSkipTLSVerifyProcessed:       utils.BoolPointer(true),
	}

	if err := erOpts.loadFromJSONCfg(erJson); err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(erOpts, exp) {
		t.Errorf("Expected %+v \n but received \n %+v", utils.ToJSON(exp), utils.ToJSON(erOpts))
	}

	erJson = nil
	if err := erOpts.loadFromJSONCfg(erJson); err != nil {
		t.Error(err)
	}
}

func TestERsLoadFromJsonCfgParseError(t *testing.T) {
	erOpts := &EventReaderOpts{}

	erJson := &EventReaderOptsJson{
		PartialPath:                       utils.StringPointer("/tmp/path"),
		PartialCacheAction:                utils.StringPointer("partial_cache_action"),
		PartialOrderField:                 utils.StringPointer("partial_order_field"),
		PartialCSVFieldSeparator:          utils.StringPointer(";"),
		CSVRowLength:                      utils.IntPointer(2),
		CSVFieldSeparator:                 utils.StringPointer(","),
		CSVHeaderDefineChar:               utils.StringPointer("header_define_char"),
		CSVLazyQuotes:                     utils.BoolPointer(false),
		XMLRootPath:                       utils.StringPointer("xml_root_path"),
		AMQPQueueID:                       utils.StringPointer("queue_id"),
		AMQPQueueIDProcessed:              utils.StringPointer("queue_id_processed"),
		AMQPConsumerTag:                   utils.StringPointer("consumer_tag"),
		AMQPExchange:                      utils.StringPointer("exchange"),
		AMQPExchangeType:                  utils.StringPointer("exchange_type"),
		AMQPRoutingKey:                    utils.StringPointer("routing_key"),
		AMQPExchangeProcessed:             utils.StringPointer("exchange_processed"),
		AMQPExchangeTypeProcessed:         utils.StringPointer("excange_type_processed"),
		AMQPRoutingKeyProcessed:           utils.StringPointer("routing_key_processed"),
		KafkaTopic:                        utils.StringPointer("topic"),
		KafkaGroupID:                      utils.StringPointer("group_id"),
		KafkaMaxWait:                      utils.StringPointer("2s"),
		KafkaTopicProcessed:               utils.StringPointer("topic_processed"),
		SQLDBName:                         utils.StringPointer("cgrates"),
		SQLTableName:                      utils.StringPointer("cgrates_t1"),
		PgSSLMode:                         utils.StringPointer("ssl_mode"),
		SQLDBNameProcessed:                utils.StringPointer("cgrates_processed"),
		SQLTableNameProcessed:             utils.StringPointer("cgrates_t1_processed"),
		PgSSLModeProcessed:                utils.StringPointer("ssl_mode_processed"),
		AWSRegion:                         utils.StringPointer("us-west"),
		AWSKey:                            utils.StringPointer("aws_key"),
		AWSSecret:                         utils.StringPointer("aws_secret"),
		AWSToken:                          utils.StringPointer("aws_token"),
		AWSRegionProcessed:                utils.StringPointer("region_processed"),
		AWSKeyProcessed:                   utils.StringPointer("aws_key_processed"),
		AWSSecretProcessed:                utils.StringPointer("aws_secret_processed"),
		AWSTokenProcessed:                 utils.StringPointer("aws_token_processed"),
		SQSQueueID:                        utils.StringPointer("queue_id"),
		SQSQueueIDProcessed:               utils.StringPointer("queue_id_processed"),
		S3BucketID:                        utils.StringPointer("bucket_id"),
		S3FolderPathProcessed:             utils.StringPointer("folder_path_processed"),
		S3BucketIDProcessed:               utils.StringPointer("bucket_id_processed"),
		NATSJetStream:                     utils.BoolPointer(false),
		NATSConsumerName:                  utils.StringPointer("consumer_name"),
		NATSSubject:                       utils.StringPointer("subject"),
		NATSQueueID:                       utils.StringPointer("queue_id"),
		NATSJWTFile:                       utils.StringPointer("jsw_file"),
		NATSSeedFile:                      utils.StringPointer("seed_file"),
		NATSCertificateAuthority:          utils.StringPointer("ca"),
		NATSClientCertificate:             utils.StringPointer("cc"),
		NATSClientKey:                     utils.StringPointer("ck"),
		NATSJetStreamMaxWait:              utils.StringPointer("2s"),
		NATSJetStreamProcessed:            utils.BoolPointer(false),
		NATSSubjectProcessed:              utils.StringPointer("subject_processed"),
		NATSJWTFileProcessed:              utils.StringPointer("jwt_file_processed"),
		NATSSeedFileProcessed:             utils.StringPointer("seed_file_processed"),
		NATSCertificateAuthorityProcessed: utils.StringPointer("ca_processed"),
		NATSClientCertificateProcessed:    utils.StringPointer("cc_processed"),
		NATSClientKeyProcessed:            utils.StringPointer("ck_processed"),
		NATSJetStreamMaxWaitProcessed:     utils.StringPointer("2s"),
	}

	errExp := `time: unknown unit "c" in duration "2c"`

	erJson.KafkaMaxWait = utils.StringPointer("2c")
	if err := erOpts.loadFromJSONCfg(erJson); err == nil || err.Error() != errExp {
		t.Errorf("Expected %v \n but received \n %v", errExp, err.Error())
	}
	erJson.KafkaMaxWait = utils.StringPointer("2s")

	/////

	erJson.NATSJetStreamMaxWait = utils.StringPointer("2c")
	if err := erOpts.loadFromJSONCfg(erJson); err == nil || err.Error() != errExp {
		t.Errorf("Expected %v \n but received \n %v", errExp, err.Error())
	}
	erJson.NATSJetStreamMaxWait = utils.StringPointer("2s")

	/////

	erJson.NATSJetStreamMaxWaitProcessed = utils.StringPointer("2c")
	if err := erOpts.loadFromJSONCfg(erJson); err == nil || err.Error() != errExp {
		t.Errorf("Expected %v \n but received \n %v", errExp, err.Error())
	}
	erJson.NATSJetStreamMaxWaitProcessed = utils.StringPointer("2s")
}

func TestERsClone(t *testing.T) {
	erOpts := &EventReaderOpts{
		PartialPath:                       utils.StringPointer("/tmp/path"),
		PartialCacheAction:                utils.StringPointer("partial_cache_action"),
		PartialOrderField:                 utils.StringPointer("partial_order_field"),
		PartialCSVFieldSeparator:          utils.StringPointer(";"),
		CSVRowLength:                      utils.IntPointer(2),
		CSVFieldSeparator:                 utils.StringPointer(","),
		CSVHeaderDefineChar:               utils.StringPointer("header_define_char"),
		CSVLazyQuotes:                     utils.BoolPointer(false),
		XMLRootPath:                       utils.StringPointer("xml_root_path"),
		AMQPQueueID:                       utils.StringPointer("queue_id"),
		AMQPQueueIDProcessed:              utils.StringPointer("queue_id_processed"),
		AMQPConsumerTag:                   utils.StringPointer("consumer_tag"),
		AMQPExchange:                      utils.StringPointer("exchange"),
		AMQPExchangeType:                  utils.StringPointer("exchange_type"),
		AMQPRoutingKey:                    utils.StringPointer("routing_key"),
		AMQPExchangeProcessed:             utils.StringPointer("exchange_processed"),
		AMQPExchangeTypeProcessed:         utils.StringPointer("excange_type_processed"),
		AMQPRoutingKeyProcessed:           utils.StringPointer("routing_key_processed"),
		KafkaTopic:                        utils.StringPointer("topic"),
		KafkaGroupID:                      utils.StringPointer("group_id"),
		KafkaMaxWait:                      utils.DurationPointer(2 * time.Second),
		KafkaTopicProcessed:               utils.StringPointer("topic_processed"),
		SQLDBName:                         utils.StringPointer("cgrates"),
		SQLTableName:                      utils.StringPointer("cgrates_t1"),
		PgSSLMode:                         utils.StringPointer("ssl_mode"),
		SQLDBNameProcessed:                utils.StringPointer("cgrates_processed"),
		SQLTableNameProcessed:             utils.StringPointer("cgrates_t1_processed"),
		PgSSLModeProcessed:                utils.StringPointer("ssl_mode_processed"),
		AWSRegion:                         utils.StringPointer("us-west"),
		AWSKey:                            utils.StringPointer("aws_key"),
		AWSSecret:                         utils.StringPointer("aws_secret"),
		AWSToken:                          utils.StringPointer("aws_token"),
		AWSRegionProcessed:                utils.StringPointer("region_processed"),
		AWSKeyProcessed:                   utils.StringPointer("aws_key_processed"),
		AWSSecretProcessed:                utils.StringPointer("aws_secret_processed"),
		AWSTokenProcessed:                 utils.StringPointer("aws_token_processed"),
		SQSQueueID:                        utils.StringPointer("queue_id"),
		SQSQueueIDProcessed:               utils.StringPointer("queue_id_processed"),
		S3BucketID:                        utils.StringPointer("bucket_id"),
		S3FolderPathProcessed:             utils.StringPointer("folder_path_processed"),
		S3BucketIDProcessed:               utils.StringPointer("bucket_id_processed"),
		NATSJetStream:                     utils.BoolPointer(false),
		NATSConsumerName:                  utils.StringPointer("consumer_name"),
		NATSSubject:                       utils.StringPointer("subject"),
		NATSQueueID:                       utils.StringPointer("queue_id"),
		NATSJWTFile:                       utils.StringPointer("jsw_file"),
		NATSSeedFile:                      utils.StringPointer("seed_file"),
		NATSCertificateAuthority:          utils.StringPointer("ca"),
		NATSClientCertificate:             utils.StringPointer("cc"),
		NATSClientKey:                     utils.StringPointer("ck"),
		NATSJetStreamMaxWait:              utils.DurationPointer(2 * time.Second),
		NATSJetStreamProcessed:            utils.BoolPointer(false),
		NATSSubjectProcessed:              utils.StringPointer("subject_processed"),
		NATSJWTFileProcessed:              utils.StringPointer("jwt_file_processed"),
		NATSSeedFileProcessed:             utils.StringPointer("seed_file_processed"),
		NATSCertificateAuthorityProcessed: utils.StringPointer("ca_processed"),
		NATSClientCertificateProcessed:    utils.StringPointer("cc_processed"),
		NATSClientKeyProcessed:            utils.StringPointer("ck_processed"),
		NATSJetStreamMaxWaitProcessed:     utils.DurationPointer(2 * time.Second),
		KafkaTLS:                          utils.BoolPointer(false),
		KafkaCAPath:                       utils.StringPointer("/tmp/path"),
		KafkaSkipTLSVerify:                utils.BoolPointer(false),
		KafkaTLSProcessed:                 utils.BoolPointer(false),
		KafkaCAPathProcessed:              utils.StringPointer("/tmp/path"),
		KafkaSkipTLSVerifyProcessed:       utils.BoolPointer(false),
	}

	exp := &EventReaderOpts{
		PartialPath:                       utils.StringPointer("/tmp/path"),
		PartialCacheAction:                utils.StringPointer("partial_cache_action"),
		PartialOrderField:                 utils.StringPointer("partial_order_field"),
		PartialCSVFieldSeparator:          utils.StringPointer(";"),
		CSVRowLength:                      utils.IntPointer(2),
		CSVFieldSeparator:                 utils.StringPointer(","),
		CSVHeaderDefineChar:               utils.StringPointer("header_define_char"),
		CSVLazyQuotes:                     utils.BoolPointer(false),
		XMLRootPath:                       utils.StringPointer("xml_root_path"),
		AMQPQueueID:                       utils.StringPointer("queue_id"),
		AMQPQueueIDProcessed:              utils.StringPointer("queue_id_processed"),
		AMQPConsumerTag:                   utils.StringPointer("consumer_tag"),
		AMQPExchange:                      utils.StringPointer("exchange"),
		AMQPExchangeType:                  utils.StringPointer("exchange_type"),
		AMQPRoutingKey:                    utils.StringPointer("routing_key"),
		AMQPExchangeProcessed:             utils.StringPointer("exchange_processed"),
		AMQPExchangeTypeProcessed:         utils.StringPointer("excange_type_processed"),
		AMQPRoutingKeyProcessed:           utils.StringPointer("routing_key_processed"),
		KafkaTopic:                        utils.StringPointer("topic"),
		KafkaGroupID:                      utils.StringPointer("group_id"),
		KafkaMaxWait:                      utils.DurationPointer(2 * time.Second),
		KafkaTopicProcessed:               utils.StringPointer("topic_processed"),
		SQLDBName:                         utils.StringPointer("cgrates"),
		SQLTableName:                      utils.StringPointer("cgrates_t1"),
		PgSSLMode:                         utils.StringPointer("ssl_mode"),
		SQLDBNameProcessed:                utils.StringPointer("cgrates_processed"),
		SQLTableNameProcessed:             utils.StringPointer("cgrates_t1_processed"),
		PgSSLModeProcessed:                utils.StringPointer("ssl_mode_processed"),
		AWSRegion:                         utils.StringPointer("us-west"),
		AWSKey:                            utils.StringPointer("aws_key"),
		AWSSecret:                         utils.StringPointer("aws_secret"),
		AWSToken:                          utils.StringPointer("aws_token"),
		AWSRegionProcessed:                utils.StringPointer("region_processed"),
		AWSKeyProcessed:                   utils.StringPointer("aws_key_processed"),
		AWSSecretProcessed:                utils.StringPointer("aws_secret_processed"),
		AWSTokenProcessed:                 utils.StringPointer("aws_token_processed"),
		SQSQueueID:                        utils.StringPointer("queue_id"),
		SQSQueueIDProcessed:               utils.StringPointer("queue_id_processed"),
		S3BucketID:                        utils.StringPointer("bucket_id"),
		S3FolderPathProcessed:             utils.StringPointer("folder_path_processed"),
		S3BucketIDProcessed:               utils.StringPointer("bucket_id_processed"),
		NATSJetStream:                     utils.BoolPointer(false),
		NATSConsumerName:                  utils.StringPointer("consumer_name"),
		NATSSubject:                       utils.StringPointer("subject"),
		NATSQueueID:                       utils.StringPointer("queue_id"),
		NATSJWTFile:                       utils.StringPointer("jsw_file"),
		NATSSeedFile:                      utils.StringPointer("seed_file"),
		NATSCertificateAuthority:          utils.StringPointer("ca"),
		NATSClientCertificate:             utils.StringPointer("cc"),
		NATSClientKey:                     utils.StringPointer("ck"),
		NATSJetStreamMaxWait:              utils.DurationPointer(2 * time.Second),
		NATSJetStreamProcessed:            utils.BoolPointer(false),
		NATSSubjectProcessed:              utils.StringPointer("subject_processed"),
		NATSJWTFileProcessed:              utils.StringPointer("jwt_file_processed"),
		NATSSeedFileProcessed:             utils.StringPointer("seed_file_processed"),
		NATSCertificateAuthorityProcessed: utils.StringPointer("ca_processed"),
		NATSClientCertificateProcessed:    utils.StringPointer("cc_processed"),
		NATSClientKeyProcessed:            utils.StringPointer("ck_processed"),
		NATSJetStreamMaxWaitProcessed:     utils.DurationPointer(2 * time.Second),
		KafkaTLS:                          utils.BoolPointer(false),
		KafkaCAPath:                       utils.StringPointer("/tmp/path"),
		KafkaSkipTLSVerify:                utils.BoolPointer(false),
		KafkaTLSProcessed:                 utils.BoolPointer(false),
		KafkaCAPathProcessed:              utils.StringPointer("/tmp/path"),
		KafkaSkipTLSVerifyProcessed:       utils.BoolPointer(false),
	}

	rcv := erOpts.Clone()
	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("Expected %+v \n but received \n %+v", utils.ToJSON(exp), utils.ToJSON(rcv))
	}
}

func TestERsAsMapInterface(t *testing.T) {
	erCfg := &EventReaderCfg{
		Opts: &EventReaderOpts{
			PartialPath:                       utils.StringPointer("/tmp/path"),
			PartialCacheAction:                utils.StringPointer("partial_cache_action"),
			PartialOrderField:                 utils.StringPointer("partial_order_field"),
			PartialCSVFieldSeparator:          utils.StringPointer(";"),
			CSVRowLength:                      utils.IntPointer(2),
			CSVFieldSeparator:                 utils.StringPointer(","),
			CSVHeaderDefineChar:               utils.StringPointer("header_define_char"),
			CSVLazyQuotes:                     utils.BoolPointer(false),
			XMLRootPath:                       utils.StringPointer("xml_root_path"),
			AMQPQueueID:                       utils.StringPointer("queue_id"),
			AMQPQueueIDProcessed:              utils.StringPointer("queue_id_processed"),
			AMQPConsumerTag:                   utils.StringPointer("consumer_tag"),
			AMQPExchange:                      utils.StringPointer("exchange"),
			AMQPExchangeType:                  utils.StringPointer("exchange_type"),
			AMQPRoutingKey:                    utils.StringPointer("routing_key"),
			AMQPExchangeProcessed:             utils.StringPointer("exchange_processed"),
			AMQPExchangeTypeProcessed:         utils.StringPointer("excange_type_processed"),
			AMQPRoutingKeyProcessed:           utils.StringPointer("routing_key_processed"),
			KafkaTopic:                        utils.StringPointer("topic"),
			KafkaGroupID:                      utils.StringPointer("group_id"),
			KafkaMaxWait:                      utils.DurationPointer(2 * time.Second),
			KafkaTopicProcessed:               utils.StringPointer("topic_processed"),
			SQLDBName:                         utils.StringPointer("cgrates"),
			SQLTableName:                      utils.StringPointer("cgrates_t1"),
			PgSSLMode:                         utils.StringPointer("ssl_mode"),
			SQLDBNameProcessed:                utils.StringPointer("cgrates_processed"),
			SQLTableNameProcessed:             utils.StringPointer("cgrates_t1_processed"),
			PgSSLModeProcessed:                utils.StringPointer("ssl_mode_processed"),
			AWSRegion:                         utils.StringPointer("us-west"),
			AWSKey:                            utils.StringPointer("aws_key"),
			AWSSecret:                         utils.StringPointer("aws_secret"),
			AWSToken:                          utils.StringPointer("aws_token"),
			AWSRegionProcessed:                utils.StringPointer("region_processed"),
			AWSKeyProcessed:                   utils.StringPointer("aws_key_processed"),
			AWSSecretProcessed:                utils.StringPointer("aws_secret_processed"),
			AWSTokenProcessed:                 utils.StringPointer("aws_token_processed"),
			SQSQueueID:                        utils.StringPointer("queue_id"),
			SQSQueueIDProcessed:               utils.StringPointer("queue_id_processed"),
			S3BucketID:                        utils.StringPointer("bucket_id"),
			S3FolderPathProcessed:             utils.StringPointer("folder_path_processed"),
			S3BucketIDProcessed:               utils.StringPointer("bucket_id_processed"),
			NATSJetStream:                     utils.BoolPointer(false),
			NATSConsumerName:                  utils.StringPointer("consumer_name"),
			NATSSubject:                       utils.StringPointer("subject"),
			NATSQueueID:                       utils.StringPointer("queue_id"),
			NATSJWTFile:                       utils.StringPointer("jsw_file"),
			NATSSeedFile:                      utils.StringPointer("seed_file"),
			NATSCertificateAuthority:          utils.StringPointer("ca"),
			NATSClientCertificate:             utils.StringPointer("cc"),
			NATSClientKey:                     utils.StringPointer("ck"),
			NATSJetStreamMaxWait:              utils.DurationPointer(2 * time.Second),
			NATSJetStreamProcessed:            utils.BoolPointer(false),
			NATSSubjectProcessed:              utils.StringPointer("subject_processed"),
			NATSJWTFileProcessed:              utils.StringPointer("jwt_file_processed"),
			NATSSeedFileProcessed:             utils.StringPointer("seed_file_processed"),
			NATSCertificateAuthorityProcessed: utils.StringPointer("ca_processed"),
			NATSClientCertificateProcessed:    utils.StringPointer("cc_processed"),
			NATSClientKeyProcessed:            utils.StringPointer("ck_processed"),
			NATSJetStreamMaxWaitProcessed:     utils.DurationPointer(2 * time.Second),
			KafkaTLS:                          utils.BoolPointer(false),
			KafkaCAPath:                       utils.StringPointer("/tmp/path"),
			KafkaSkipTLSVerify:                utils.BoolPointer(false),
			KafkaTLSProcessed:                 utils.BoolPointer(false),
			KafkaCAPathProcessed:              utils.StringPointer("/tmp/path"),
			KafkaSkipTLSVerifyProcessed:       utils.BoolPointer(false),
		},
	}

	exp := map[string]any{
		"opts": map[string]any{
			"amqpConsumerTag":                   "consumer_tag",
			"amqpExchange":                      "exchange",
			"amqpExchangeProcessed":             "exchange_processed",
			"amqpExchangeType":                  "exchange_type",
			"amqpExchangeTypeProcessed":         "excange_type_processed",
			"amqpQueueID":                       "queue_id",
			"amqpQueueIDProcessed":              "queue_id_processed",
			"amqpRoutingKey":                    "routing_key",
			"amqpRoutingKeyProcessed":           "routing_key_processed",
			"awsKey":                            "aws_key",
			"awsKeyProcessed":                   "aws_key_processed",
			"awsRegion":                         "us-west",
			"awsRegionProcessed":                "region_processed",
			"awsSecret":                         "aws_secret",
			"awsSecretProcessed":                "aws_secret_processed",
			"awsToken":                          "aws_token",
			"awsTokenProcessed":                 "aws_token_processed",
			"csvFieldSeparator":                 ",",
			"csvHeaderDefineChar":               "header_define_char",
			"csvLazyQuotes":                     false,
			"csvRowLength":                      2,
			"kafkaGroupID":                      "group_id",
			"kafkaMaxWait":                      "2s",
			"kafkaTopic":                        "topic",
			"kafkaTopicProcessed":               "topic_processed",
			"natsCertificateAuthority":          "ca",
			"natsCertificateAuthorityProcessed": "ca_processed",
			"natsClientCertificate":             "cc",
			"natsClientCertificateProcessed":    "cc_processed",
			"natsClientKey":                     "ck",
			"natsClientKeyProcessed":            "ck_processed",
			"natsConsumerName":                  "consumer_name",
			"natsJWTFile":                       "jsw_file",
			"natsJWTFileProcessed":              "jwt_file_processed",
			"natsJetStream":                     false,
			"natsJetStreamMaxWait":              "2s",
			"natsJetStreamMaxWaitProcessed":     "2s",
			"natsJetStreamProcessed":            false,
			"natsQueueID":                       "queue_id",
			"natsSeedFile":                      "seed_file",
			"natsSeedFileProcessed":             "seed_file_processed",
			"natsSubject":                       "subject",
			"natsSubjectProcessed":              "subject_processed",
			"partialCacheAction":                "partial_cache_action",
			"partialOrderField":                 "partial_order_field",
			"partialPath":                       "/tmp/path",
			"partialcsvFieldSeparator":          ";",
			"s3BucketID":                        "bucket_id",
			"s3BucketIDProcessed":               "bucket_id_processed",
			"s3FolderPathProcessed":             "folder_path_processed",
			"sqlDBName":                         "cgrates",
			"sqlDBNameProcessed":                "cgrates_processed",
			"sqlTableName":                      "cgrates_t1",
			"sqlTableNameProcessed":             "cgrates_t1_processed",
			"sqsQueueID":                        "queue_id",
			"sqsQueueIDProcessed":               "queue_id_processed",
			"pgSSLMode":                         "ssl_mode",
			"pgSSLModeProcessed":                "ssl_mode_processed",
			"xmlRootPath":                       "xml_root_path",
			"kafkaTLS":                          false,
			"kafkaCAPath":                       "/tmp/path",
			"kafkaSkipTLSVerify":                false,
			"kafkaTLSProcessed":                 false,
			"kafkaCAPathProcessed":              "/tmp/path",
			"kafkaSkipTLSVerifyProcessed":       false,
		},
	}

	rcv := erCfg.AsMapInterface("")
	if !reflect.DeepEqual(rcv[utils.OptsCfg], exp[utils.OptsCfg]) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp[utils.OptsCfg]), utils.ToJSON(rcv[utils.OptsCfg]))
	}
}

func TestDiffEventReaderOptsJsonCfg(t *testing.T) {
	var d *EventReaderOptsJson

	v1 := &EventReaderOpts{
		PartialPath:                       utils.StringPointer("/tmp/path/diff"),
		PartialCacheAction:                utils.StringPointer("partial_cache_action_diff"),
		PartialOrderField:                 utils.StringPointer("partial_order_field_diff"),
		PartialCSVFieldSeparator:          utils.StringPointer(";_diff"),
		CSVRowLength:                      utils.IntPointer(3),
		CSVFieldSeparator:                 utils.StringPointer(",_diff"),
		CSVHeaderDefineChar:               utils.StringPointer("header_define_char_diff"),
		CSVLazyQuotes:                     utils.BoolPointer(true),
		XMLRootPath:                       utils.StringPointer("xml_root_path_diff"),
		AMQPQueueID:                       utils.StringPointer("queue_id_diff"),
		AMQPQueueIDProcessed:              utils.StringPointer("queue_id_processed_diff"),
		AMQPConsumerTag:                   utils.StringPointer("consumer_tag_diff"),
		AMQPExchange:                      utils.StringPointer("exchange_diff"),
		AMQPExchangeType:                  utils.StringPointer("exchange_type_diff"),
		AMQPRoutingKey:                    utils.StringPointer("routing_key_diff"),
		AMQPExchangeProcessed:             utils.StringPointer("exchange_processed_diff"),
		AMQPExchangeTypeProcessed:         utils.StringPointer("excange_type_processed_diff"),
		AMQPRoutingKeyProcessed:           utils.StringPointer("routing_key_processed_diff"),
		KafkaTopic:                        utils.StringPointer("topic_diff"),
		KafkaGroupID:                      utils.StringPointer("group_id_diff"),
		KafkaMaxWait:                      utils.DurationPointer(3 * time.Second),
		KafkaTopicProcessed:               utils.StringPointer("topic_processed_diff"),
		SQLDBName:                         utils.StringPointer("cgrates_diff"),
		SQLTableName:                      utils.StringPointer("cgrates_t1_diff"),
		PgSSLMode:                         utils.StringPointer("ssl_mode_diff"),
		SQLDBNameProcessed:                utils.StringPointer("cgrates_processed_diff"),
		SQLTableNameProcessed:             utils.StringPointer("cgrates_t1_processed_diff"),
		PgSSLModeProcessed:                utils.StringPointer("ssl_mode_processed_diff"),
		AWSRegion:                         utils.StringPointer("us-west_diff"),
		AWSKey:                            utils.StringPointer("aws_key_diff"),
		AWSSecret:                         utils.StringPointer("aws_secret_diff"),
		AWSToken:                          utils.StringPointer("aws_token_diff"),
		AWSRegionProcessed:                utils.StringPointer("region_processed_diff"),
		AWSKeyProcessed:                   utils.StringPointer("aws_key_processed_diff"),
		AWSSecretProcessed:                utils.StringPointer("aws_secret_processed_diff"),
		AWSTokenProcessed:                 utils.StringPointer("aws_token_processed_diff"),
		SQSQueueID:                        utils.StringPointer("queue_id_diff"),
		SQSQueueIDProcessed:               utils.StringPointer("queue_id_processed_diff"),
		S3BucketID:                        utils.StringPointer("bucket_id_diff"),
		S3FolderPathProcessed:             utils.StringPointer("folder_path_processed_diff"),
		S3BucketIDProcessed:               utils.StringPointer("bucket_id_processed_diff"),
		NATSJetStream:                     utils.BoolPointer(true),
		NATSConsumerName:                  utils.StringPointer("consumer_name_diff"),
		NATSSubject:                       utils.StringPointer("subject_diff"),
		NATSQueueID:                       utils.StringPointer("queue_id_diff"),
		NATSJWTFile:                       utils.StringPointer("jsw_file_diff"),
		NATSSeedFile:                      utils.StringPointer("seed_file_diff"),
		NATSCertificateAuthority:          utils.StringPointer("ca_diff"),
		NATSClientCertificate:             utils.StringPointer("cc_diff"),
		NATSClientKey:                     utils.StringPointer("ck_diff"),
		NATSJetStreamMaxWait:              utils.DurationPointer(5 * time.Second),
		NATSJetStreamProcessed:            utils.BoolPointer(true),
		NATSSubjectProcessed:              utils.StringPointer("subject_processed_diff"),
		NATSJWTFileProcessed:              utils.StringPointer("jwt_file_processed_diff"),
		NATSSeedFileProcessed:             utils.StringPointer("seed_file_processed_diff"),
		NATSCertificateAuthorityProcessed: utils.StringPointer("ca_processed_diff"),
		NATSClientCertificateProcessed:    utils.StringPointer("cc_processed_diff"),
		NATSClientKeyProcessed:            utils.StringPointer("ck_processed_diff"),
		NATSJetStreamMaxWaitProcessed:     utils.DurationPointer(6 * time.Second),
		KafkaTLS:                          utils.BoolPointer(true),
		KafkaCAPath:                       utils.StringPointer("/tmp/path/diff"),
		KafkaSkipTLSVerify:                utils.BoolPointer(true),
		KafkaTLSProcessed:                 utils.BoolPointer(true),
		KafkaCAPathProcessed:              utils.StringPointer("/tmp/path/diff"),
		KafkaSkipTLSVerifyProcessed:       utils.BoolPointer(true),
	}

	v2 := &EventReaderOpts{
		PartialPath:                       utils.StringPointer("/tmp/path"),
		PartialCacheAction:                utils.StringPointer("partial_cache_action"),
		PartialOrderField:                 utils.StringPointer("partial_order_field"),
		PartialCSVFieldSeparator:          utils.StringPointer(";"),
		CSVRowLength:                      utils.IntPointer(2),
		CSVFieldSeparator:                 utils.StringPointer(","),
		CSVHeaderDefineChar:               utils.StringPointer("header_define_char"),
		CSVLazyQuotes:                     utils.BoolPointer(false),
		XMLRootPath:                       utils.StringPointer("xml_root_path"),
		AMQPQueueID:                       utils.StringPointer("queue_id"),
		AMQPQueueIDProcessed:              utils.StringPointer("queue_id_processed"),
		AMQPConsumerTag:                   utils.StringPointer("consumer_tag"),
		AMQPExchange:                      utils.StringPointer("exchange"),
		AMQPExchangeType:                  utils.StringPointer("exchange_type"),
		AMQPRoutingKey:                    utils.StringPointer("routing_key"),
		AMQPExchangeProcessed:             utils.StringPointer("exchange_processed"),
		AMQPExchangeTypeProcessed:         utils.StringPointer("excange_type_processed"),
		AMQPRoutingKeyProcessed:           utils.StringPointer("routing_key_processed"),
		KafkaTopic:                        utils.StringPointer("topic"),
		KafkaGroupID:                      utils.StringPointer("group_id"),
		KafkaMaxWait:                      utils.DurationPointer(2 * time.Second),
		KafkaTopicProcessed:               utils.StringPointer("topic_processed"),
		SQLDBName:                         utils.StringPointer("cgrates"),
		SQLTableName:                      utils.StringPointer("cgrates_t1"),
		PgSSLMode:                         utils.StringPointer("ssl_mode"),
		SQLDBNameProcessed:                utils.StringPointer("cgrates_processed"),
		SQLTableNameProcessed:             utils.StringPointer("cgrates_t1_processed"),
		PgSSLModeProcessed:                utils.StringPointer("ssl_mode_processed"),
		AWSRegion:                         utils.StringPointer("us-west"),
		AWSKey:                            utils.StringPointer("aws_key"),
		AWSSecret:                         utils.StringPointer("aws_secret"),
		AWSToken:                          utils.StringPointer("aws_token"),
		AWSRegionProcessed:                utils.StringPointer("region_processed"),
		AWSKeyProcessed:                   utils.StringPointer("aws_key_processed"),
		AWSSecretProcessed:                utils.StringPointer("aws_secret_processed"),
		AWSTokenProcessed:                 utils.StringPointer("aws_token_processed"),
		SQSQueueID:                        utils.StringPointer("queue_id"),
		SQSQueueIDProcessed:               utils.StringPointer("queue_id_processed"),
		S3BucketID:                        utils.StringPointer("bucket_id"),
		S3FolderPathProcessed:             utils.StringPointer("folder_path_processed"),
		S3BucketIDProcessed:               utils.StringPointer("bucket_id_processed"),
		NATSJetStream:                     utils.BoolPointer(false),
		NATSConsumerName:                  utils.StringPointer("consumer_name"),
		NATSSubject:                       utils.StringPointer("subject"),
		NATSQueueID:                       utils.StringPointer("queue_id"),
		NATSJWTFile:                       utils.StringPointer("jsw_file"),
		NATSSeedFile:                      utils.StringPointer("seed_file"),
		NATSCertificateAuthority:          utils.StringPointer("ca"),
		NATSClientCertificate:             utils.StringPointer("cc"),
		NATSClientKey:                     utils.StringPointer("ck"),
		NATSJetStreamMaxWait:              utils.DurationPointer(2 * time.Second),
		NATSJetStreamProcessed:            utils.BoolPointer(false),
		NATSSubjectProcessed:              utils.StringPointer("subject_processed"),
		NATSJWTFileProcessed:              utils.StringPointer("jwt_file_processed"),
		NATSSeedFileProcessed:             utils.StringPointer("seed_file_processed"),
		NATSCertificateAuthorityProcessed: utils.StringPointer("ca_processed"),
		NATSClientCertificateProcessed:    utils.StringPointer("cc_processed"),
		NATSClientKeyProcessed:            utils.StringPointer("ck_processed"),
		NATSJetStreamMaxWaitProcessed:     utils.DurationPointer(2 * time.Second),
		KafkaTLS:                          utils.BoolPointer(false),
		KafkaCAPath:                       utils.StringPointer("/tmp/path"),
		KafkaSkipTLSVerify:                utils.BoolPointer(false),
		KafkaTLSProcessed:                 utils.BoolPointer(false),
		KafkaCAPathProcessed:              utils.StringPointer("/tmp/path"),
		KafkaSkipTLSVerifyProcessed:       utils.BoolPointer(false),
	}

	exp := &EventReaderOptsJson{
		PartialPath:                       utils.StringPointer("/tmp/path"),
		PartialCacheAction:                utils.StringPointer("partial_cache_action"),
		PartialOrderField:                 utils.StringPointer("partial_order_field"),
		PartialCSVFieldSeparator:          utils.StringPointer(";"),
		CSVRowLength:                      utils.IntPointer(2),
		CSVFieldSeparator:                 utils.StringPointer(","),
		CSVHeaderDefineChar:               utils.StringPointer("header_define_char"),
		CSVLazyQuotes:                     utils.BoolPointer(false),
		XMLRootPath:                       utils.StringPointer("xml_root_path"),
		AMQPQueueID:                       utils.StringPointer("queue_id"),
		AMQPQueueIDProcessed:              utils.StringPointer("queue_id_processed"),
		AMQPConsumerTag:                   utils.StringPointer("consumer_tag"),
		AMQPExchange:                      utils.StringPointer("exchange"),
		AMQPExchangeType:                  utils.StringPointer("exchange_type"),
		AMQPRoutingKey:                    utils.StringPointer("routing_key"),
		AMQPExchangeProcessed:             utils.StringPointer("exchange_processed"),
		AMQPExchangeTypeProcessed:         utils.StringPointer("excange_type_processed"),
		AMQPRoutingKeyProcessed:           utils.StringPointer("routing_key_processed"),
		KafkaTopic:                        utils.StringPointer("topic"),
		KafkaGroupID:                      utils.StringPointer("group_id"),
		KafkaMaxWait:                      utils.StringPointer("2s"),
		KafkaTopicProcessed:               utils.StringPointer("topic_processed"),
		SQLDBName:                         utils.StringPointer("cgrates"),
		SQLTableName:                      utils.StringPointer("cgrates_t1"),
		PgSSLMode:                         utils.StringPointer("ssl_mode"),
		SQLDBNameProcessed:                utils.StringPointer("cgrates_processed"),
		SQLTableNameProcessed:             utils.StringPointer("cgrates_t1_processed"),
		PgSSLModeProcessed:                utils.StringPointer("ssl_mode_processed"),
		AWSRegion:                         utils.StringPointer("us-west"),
		AWSKey:                            utils.StringPointer("aws_key"),
		AWSSecret:                         utils.StringPointer("aws_secret"),
		AWSToken:                          utils.StringPointer("aws_token"),
		AWSRegionProcessed:                utils.StringPointer("region_processed"),
		AWSKeyProcessed:                   utils.StringPointer("aws_key_processed"),
		AWSSecretProcessed:                utils.StringPointer("aws_secret_processed"),
		AWSTokenProcessed:                 utils.StringPointer("aws_token_processed"),
		SQSQueueID:                        utils.StringPointer("queue_id"),
		SQSQueueIDProcessed:               utils.StringPointer("queue_id_processed"),
		S3BucketID:                        utils.StringPointer("bucket_id"),
		S3FolderPathProcessed:             utils.StringPointer("folder_path_processed"),
		S3BucketIDProcessed:               utils.StringPointer("bucket_id_processed"),
		NATSJetStream:                     utils.BoolPointer(false),
		NATSConsumerName:                  utils.StringPointer("consumer_name"),
		NATSSubject:                       utils.StringPointer("subject"),
		NATSQueueID:                       utils.StringPointer("queue_id"),
		NATSJWTFile:                       utils.StringPointer("jsw_file"),
		NATSSeedFile:                      utils.StringPointer("seed_file"),
		NATSCertificateAuthority:          utils.StringPointer("ca"),
		NATSClientCertificate:             utils.StringPointer("cc"),
		NATSClientKey:                     utils.StringPointer("ck"),
		NATSJetStreamMaxWait:              utils.StringPointer("2s"),
		NATSJetStreamProcessed:            utils.BoolPointer(false),
		NATSSubjectProcessed:              utils.StringPointer("subject_processed"),
		NATSJWTFileProcessed:              utils.StringPointer("jwt_file_processed"),
		NATSSeedFileProcessed:             utils.StringPointer("seed_file_processed"),
		NATSCertificateAuthorityProcessed: utils.StringPointer("ca_processed"),
		NATSClientCertificateProcessed:    utils.StringPointer("cc_processed"),
		NATSClientKeyProcessed:            utils.StringPointer("ck_processed"),
		NATSJetStreamMaxWaitProcessed:     utils.StringPointer("2s"),
		KafkaTLS:                          utils.BoolPointer(false),
		KafkaCAPath:                       utils.StringPointer("/tmp/path"),
		KafkaSkipTLSVerify:                utils.BoolPointer(false),
		KafkaTLSProcessed:                 utils.BoolPointer(false),
		KafkaCAPathProcessed:              utils.StringPointer("/tmp/path"),
		KafkaSkipTLSVerifyProcessed:       utils.BoolPointer(false),
	}

	if rcv := diffEventReaderOptsJsonCfg(d, v1, v2); !reflect.DeepEqual(rcv, exp) {
		t.Errorf("Expected %v \n but received \n %v", exp, rcv)
	}
}
