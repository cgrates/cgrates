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
			"opts": {},
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
            "failed_calls_prefix": "randomPrefix",
            "partial_record_cache": "1s",
            "partial_cache_expiry_action": "randomAction"
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
					CSVFieldSeparator:   utils.StringPointer(utils.FieldsSep),
					CSVHeaderDefineChar: utils.StringPointer(utils.InInFieldSep),
					CSVRowLength:        utils.IntPointer(0),
					XMLRootPath:         utils.StringPointer(utils.EmptyString),
					PartialOrderField:   utils.StringPointer("~*req.AnswerTime"),
					PartialCacheAction:  utils.StringPointer(utils.MetaNone),
					NATSSubject:         utils.StringPointer("cgrates_cdrs"),
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
				PartialCommitFields: make([]*FCTemplate, 0),
				Opts: &EventReaderOpts{
					CSVFieldSeparator:   utils.StringPointer(utils.FieldsSep),
					CSVHeaderDefineChar: utils.StringPointer(utils.InInFieldSep),
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
	if err = jsoncfg.ersCfg.loadFromJSONCfg(cfgJSON, jsoncfg.templates, jsoncfg.generalCfg.RSRSep, jsoncfg.dfltEvRdr, jsoncfg.generalCfg.RSRSep); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
	cfgJson := &ERsJsonCfg{

		Partial_cache_ttl: utils.StringPointer("test"),
	}

	if err := jsoncfg.ersCfg.loadFromJSONCfg(cfgJson, jsoncfg.templates, jsoncfg.generalCfg.RSRSep, jsoncfg.dfltEvRdr, jsoncfg.generalCfg.RSRSep); err == nil {
		t.Error(err)
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
	if err = jsoncfg.ersCfg.loadFromJSONCfg(cfgJSON, jsoncfg.templates, jsoncfg.generalCfg.RSRSep, jsoncfg.dfltEvRdr, jsoncfg.generalCfg.RSRSep); err == nil || err.Error() != expected {
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
					CSVFieldSeparator:   utils.StringPointer(utils.FieldsSep),
					CSVHeaderDefineChar: utils.StringPointer(utils.InInFieldSep),
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
					CSVFieldSeparator:   utils.StringPointer(utils.FieldsSep),
					CSVHeaderDefineChar: utils.StringPointer(utils.InInFieldSep),
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
	if err = jsonCfg.ersCfg.loadFromJSONCfg(cfgJSON, jsonCfg.templates, jsonCfg.generalCfg.RSRSep, jsonCfg.dfltEvRdr, jsonCfg.generalCfg.RSRSep); err != nil {
		t.Error(err)
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
	if err = jsonCfg.ersCfg.loadFromJSONCfg(cfgJSON, jsonCfg.templates, jsonCfg.generalCfg.RSRSep, jsonCfg.dfltEvRdr, jsonCfg.generalCfg.RSRSep); err == nil || err.Error() != expected {
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
	if err = jsonCfg.ersCfg.loadFromJSONCfg(cfgJSON, jsonCfg.templates, jsonCfg.generalCfg.RSRSep, jsonCfg.dfltEvRdr, jsonCfg.generalCfg.RSRSep); err == nil || err.Error() != expected {
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
	if err = jsonCfg.ersCfg.loadFromJSONCfg(cfgJSON, jsonCfg.templates, jsonCfg.generalCfg.RSRSep, jsonCfg.dfltEvRdr, jsonCfg.generalCfg.RSRSep); err == nil || err.Error() != expected {
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
					CSVFieldSeparator:   utils.StringPointer(utils.FieldsSep),
					CSVHeaderDefineChar: utils.StringPointer(utils.InInFieldSep),
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
						Tag:    utils.CGRID,
						Path:   "*exp.CGRID",
						Type:   utils.MetaVariable,
						Layout: time.RFC3339,
					},
				},
				CacheDumpFields:     make([]*FCTemplate, 0),
				PartialCommitFields: make([]*FCTemplate, 0),
				Opts: &EventReaderOpts{
					CSVFieldSeparator:   utils.StringPointer(utils.FieldsSep),
					CSVHeaderDefineChar: utils.StringPointer(utils.InInFieldSep),
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
	if err = jsonCfg.ersCfg.loadFromJSONCfg(cfgJSON, msgTemplates, jsonCfg.generalCfg.RSRSep, jsonCfg.dfltEvRdr, jsonCfg.generalCfg.RSRSep); err != nil {
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
					CSVFieldSeparator:   utils.StringPointer(utils.FieldsSep),
					CSVHeaderDefineChar: utils.StringPointer(utils.InInFieldSep),
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
					CSVFieldSeparator:   utils.StringPointer(utils.FieldsSep),
					CSVHeaderDefineChar: utils.StringPointer(utils.InInFieldSep),
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
	if err = jsonCfg.ersCfg.loadFromJSONCfg(cfgJSON, msgTemplates, jsonCfg.generalCfg.RSRSep, jsonCfg.dfltEvRdr, jsonCfg.generalCfg.RSRSep); err != nil {
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
	if err = jsonCfg.ersCfg.loadFromJSONCfg(cfgJSON, jsonCfg.templates, jsonCfg.generalCfg.RSRSep, jsonCfg.dfltEvRdr, jsonCfg.generalCfg.RSRSep); err == nil || err.Error() != expected {
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
					CSVFieldSeparator:   utils.StringPointer(utils.FieldsSep),
					CSVHeaderDefineChar: utils.StringPointer(utils.InInFieldSep),
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
					CSVFieldSeparator:   utils.StringPointer(utils.FieldsSep),
					CSVHeaderDefineChar: utils.StringPointer(utils.InInFieldSep),
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
					"csvRowLength":        0,
					"xmlRootPath":         "",
					"partialOrderField":   "~*req.AnswerTime",
					"partialCacheAction":  utils.MetaNone,
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
					"csvRowLength":        0,
					"xmlRootPath":         "",
					"partialOrderField":   "~*req.AnswerTime",
					"partialCacheAction":  utils.MetaNone,
					"natsSubject":         "cgrates_cdrs",
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
				"type": "*file_csv",
                "flags": ["randomFlag"],
                "filters": ["randomFilter"],
				"source_path": "/tmp/ers/in",
                "partial_record_cache": "1s",
				"processed_path": "/tmp/ers/out",
				"cache_dump_fields": [
                           {"tag": "ToR", "path": "*cgreq.ToR", "type": "*variable", "value": "~*req.2", "mandatory": true}                
				],
				"opts":{
					"kafkaGroupID": "test",
					"csvLazyQuotes": false,
					"amqpQueueID": "id",
					"amqpQueueIDProcessed":"queue",
					"amqpConsumerTag": "tag",
					"amqpExchange":"exchange",
					"amqpExchangeType":"type",
					"amqpRoutingKey":"key",
					"amqpExchangeProcessed":"amqprocess",
					"amqpExchangeTypeProcessed":"amqtype",
					"amqpRoutingKeyProcessed":"routekey",
					"partialPath":"optpath",
					"partialcsvFieldSeparator":",",
					"kafkaTopic":"topic",
					"kafkaMaxWait":"1m",
					"kafkaTopicProcessed":"process",
					"sqlDBName":"db",
					"sqlTableName":"table",
					"pgSSLMode":"pg",
					"sqlDBNameProcessed":"dbname",
					"sqlTableNameProcessed":"tablename",
					"pgSSLModeProcessed":"sslprocess",
					"awsRegion":"eu",
					"awsKey":"key",
					"awsSecret":"secret",
					"awsToken":"token",
					"awsRegionProcessed":"awsRegion",
					"awsKeyProcessed":"keyprocess",
					"awsSecretProcessed":"secretkey",
					"awsTokenProcessed": "tokenprocess",
					"sqsQueueID":"sqs",
					"sqsQueueIDProcessed":"sqsid",
					"s3BucketID":"s3bucket",
					"s3FolderPathProcessed": "s3folder",
					"s3BucketIDProcessed":"s3bucketid",
					"natsJetStream":true,
					"natsConsumerName": "NATConsumer",
					"natsQueueID":"NATid",
					"natsJWTFile":"jwt",
					"natsCertificateAuthority":"auth",
					"natsClientCertificate":"certificate",
					"natsSeedFile":"seed",
					"natsClientKey":"clientkey",
					"natsJetStreamMaxWait":"1m",
					"natsJetStreamProcessed":false,
					"natsSubjectProcessed":"NATProcess",
					"natsJWTFileProcessed":"jwtfile",
					"natsSeedFileProcessed":"NATSeed",
					"natsCertificateAuthorityProcessed":"CertificateAuth",
					"natsClientCertificateProcessed":"ClientCertificate",
					"natsClientKeyProcessed":"ClientKey",
					"natsJetStreamMaxWaitProcessed":"1m",
					
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
					"csvRowLength":        0,
					"xmlRootPath":         "",
					"partialOrderField":   "~*req.AnswerTime",
					"partialCacheAction":  utils.MetaNone,
					"natsSubject":         "cgrates_cdrs",
				},
			},
			{
				utils.CacheDumpFieldsCfg: []map[string]interface{}{
					{utils.MandatoryCfg: true, utils.PathCfg: "*cgreq.ToR", utils.TagCfg: "ToR", utils.TypeCfg: "*variable", utils.ValueCfg: "~*req.2"},
				},
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
				utils.FiltersCfg:       []string{"randomFilter"},
				utils.FlagsCfg:         []string{"randomFlag"},
				utils.IDCfg:            "file_reader1",
				utils.ProcessedPathCfg: "/tmp/ers/out",
				utils.RunDelayCfg:      "10s",
				utils.SourcePathCfg:    "/tmp/ers/in",
				utils.TenantCfg:        "~*req.Destination1",
				utils.TimezoneCfg:      "",
				utils.OptsCfg: map[string]interface{}{
					utils.CSVLazyQuotes:                        false,
					utils.KafkaGroupID:                         "test",
					utils.CSVFieldSepOpt:                       ",",
					utils.XMLRootPathOpt:                       "",
					"csvHeaderDefineChar":                      ":",
					"csvRowLength":                             0,
					"partialOrderField":                        "~*req.AnswerTime",
					"partialCacheAction":                       utils.MetaNone,
					"natsSubject":                              "cgrates_cdrs",
					utils.AMQPQueueID:                          "id",
					utils.AMQPQueueIDProcessedCfg:              "queue",
					utils.AMQPConsumerTag:                      "tag",
					utils.AMQPExchange:                         "exchange",
					utils.AMQPExchangeType:                     "type",
					utils.AMQPRoutingKey:                       "key",
					utils.AMQPExchangeProcessedCfg:             "amqprocess",
					utils.AMQPExchangeTypeProcessedCfg:         "amqtype",
					utils.AMQPRoutingKeyProcessedCfg:           "routekey",
					utils.PartialPathOpt:                       "optpath",
					utils.PartialCSVFieldSepartorOpt:           ",",
					utils.KafkaTopic:                           "topic",
					utils.KafkaMaxWait:                         "1m0s",
					utils.KafkaTopicProcessedCfg:               "process",
					utils.SQLDBNameOpt:                         "db",
					utils.SQLTableNameOpt:                      "table",
					utils.PgSSLModeCfg:                         "pg",
					utils.SQLDBNameProcessedCfg:                "dbname",
					utils.SQLTableNameProcessedCfg:             "tablename",
					utils.PgSSLModeProcessedCfg:                "sslprocess",
					utils.AWSRegion:                            "eu",
					utils.AWSKey:                               "key",
					utils.AWSSecret:                            "secret",
					utils.AWSToken:                             "token",
					utils.AWSRegionProcessedCfg:                "awsRegion",
					utils.AWSKeyProcessedCfg:                   "keyprocess",
					utils.AWSSecretProcessedCfg:                "secretkey",
					utils.AWSTokenProcessedCfg:                 "tokenprocess",
					utils.SQSQueueID:                           "sqs",
					utils.SQSQueueIDProcessedCfg:               "sqsid",
					utils.S3Bucket:                             "s3bucket",
					utils.S3FolderPathProcessedCfg:             "s3folder",
					utils.S3BucketIDProcessedCfg:               "s3bucketid",
					utils.NatsJetStream:                        true,
					utils.NatsConsumerName:                     "NATConsumer",
					utils.NatsQueueID:                          "NATid",
					utils.NatsJWTFile:                          "jwt",
					utils.NatsSeedFile:                         "seed",
					utils.NatsCertificateAuthority:             "auth",
					utils.NatsClientCertificate:                "certificate",
					utils.NatsClientKey:                        "clientkey",
					utils.NatsJetStreamMaxWait:                 "1m0s",
					utils.NATSJetStreamProcessedCfg:            false,
					utils.NATSSubjectProcessedCfg:              "NATProcess",
					utils.NATSJWTFileProcessedCfg:              "jwtfile",
					utils.NATSSeedFileProcessedCfg:             "NATSeed",
					utils.NATSCertificateAuthorityProcessedCfg: "CertificateAuth",
					utils.NATSClientCertificateProcessed:       "ClientCertificate",
					utils.NATSClientKeyProcessedCfg:            "ClientKey",
					utils.NATSJetStreamMaxWaitProcessedCfg:     "1m0s",
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
				Opts: &EventReaderOpts{
					CSVFieldSeparator:   utils.StringPointer(utils.FieldsSep),
					CSVHeaderDefineChar: utils.StringPointer(utils.InInFieldSep),
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
				Opts: &EventReaderOpts{
					CSVFieldSeparator:   utils.StringPointer(utils.FieldsSep),
					CSVHeaderDefineChar: utils.StringPointer(utils.InInFieldSep),
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
	if err := cfgCgr.ersCfg.loadFromJSONCfg(cfgJSON, cfgCgr.templates, cfgCgr.generalCfg.RSRSep, cfgCgr.dfltEvRdr, cfgCgr.generalCfg.RSRSep); err != nil {
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
func TestEventReaderOptsCfg(t *testing.T) {
	erCfg := new(EventReaderCfg)
	if err := erCfg.Opts.loadFromJSONCfg(nil); err != nil {
		t.Error(err)
	}
	eventReaderOptsJson := &EventReaderOptsJson{
		PartialPath:                       utils.StringPointer("path"),
		PartialCSVFieldSeparator:          utils.StringPointer("/"),
		CSVLazyQuotes:                     utils.BoolPointer(false),
		AMQPQueueID:                       utils.StringPointer("id"),
		AMQPQueueIDProcessed:              utils.StringPointer("id"),
		AMQPConsumerTag:                   utils.StringPointer("tag"),
		AMQPExchange:                      utils.StringPointer("exchange"),
		AMQPExchangeType:                  utils.StringPointer("type"),
		AMQPRoutingKey:                    utils.StringPointer("key1"),
		AMQPExchangeProcessed:             utils.StringPointer("amq"),
		AMQPExchangeTypeProcessed:         utils.StringPointer("amqtype"),
		AMQPRoutingKeyProcessed:           utils.StringPointer("key"),
		KafkaTopic:                        utils.StringPointer("kafka"),
		KafkaMaxWait:                      utils.StringPointer("1m"),
		KafkaTopicProcessed:               utils.StringPointer("kafkaproc"),
		SQLDBName:                         utils.StringPointer("dbname"),
		SQLTableName:                      utils.StringPointer("tablename"),
		PgSSLMode:                         utils.StringPointer("sslmode"),
		SQLDBNameProcessed:                utils.StringPointer("dbnameproc"),
		SQLTableNameProcessed:             utils.StringPointer("tablenameproc"),
		PgSSLModeProcessed:                utils.StringPointer("sslproc"),
		AWSRegion:                         utils.StringPointer("eu"),
		AWSKey:                            utils.StringPointer("key"),
		AWSSecret:                         utils.StringPointer("secret"),
		AWSToken:                          utils.StringPointer("token"),
		AWSKeyProcessed:                   utils.StringPointer("secret"),
		AWSSecretProcessed:                utils.StringPointer("secret"),
		AWSTokenProcessed:                 utils.StringPointer("token"),
		SQSQueueID:                        utils.StringPointer("SQSQueue"),
		SQSQueueIDProcessed:               utils.StringPointer("SQSQueueId"),
		S3BucketID:                        utils.StringPointer("S3BucketID"),
		S3FolderPathProcessed:             utils.StringPointer("S3Path"),
		S3BucketIDProcessed:               utils.StringPointer("S3BucketProc"),
		NATSJetStream:                     utils.BoolPointer(false),
		NATSConsumerName:                  utils.StringPointer("user"),
		NATSQueueID:                       utils.StringPointer("id"),
		NATSJWTFile:                       utils.StringPointer("jwt"),
		NATSSeedFile:                      utils.StringPointer("seed"),
		NATSCertificateAuthority:          utils.StringPointer("authority"),
		NATSClientCertificate:             utils.StringPointer("certificate"),
		NATSClientKey:                     utils.StringPointer("key5"),
		NATSJetStreamMaxWait:              utils.StringPointer("1m"),
		NATSJetStreamProcessed:            utils.BoolPointer(true),
		NATSJWTFileProcessed:              utils.StringPointer("file"),
		NATSSeedFileProcessed:             utils.StringPointer("natseed"),
		NATSCertificateAuthorityProcessed: utils.StringPointer("natsauth"),
		NATSClientCertificateProcessed:    utils.StringPointer("natcertificate"),
		NATSClientKeyProcessed:            utils.StringPointer("natsprocess"),
		NATSJetStreamMaxWaitProcessed:     utils.StringPointer("1m"),
		AWSRegionProcessed:                utils.StringPointer("eu"),
		NATSSubjectProcessed:              utils.StringPointer("process"),
		KafkaGroupID:                      utils.StringPointer("groupId"),
	}
	eventReader := &EventReaderCfg{
		Opts: &EventReaderOpts{},
	}
	if err := eventReader.Opts.loadFromJSONCfg(eventReaderOptsJson); err != nil {
		t.Error(err)
	}
	if err := eventReader.Opts.loadFromJSONCfg(&EventReaderOptsJson{
		KafkaMaxWait: utils.StringPointer("test"),
	}); err == nil {
		t.Error(err)
	} else if err := eventReader.Opts.loadFromJSONCfg(&EventReaderOptsJson{
		NATSJetStreamMaxWait: utils.StringPointer("nil"),
	}); err == nil {
		t.Error(err)
	} else if err := eventReader.Opts.loadFromJSONCfg(&EventReaderOptsJson{
		NATSJetStreamMaxWaitProcessed: utils.StringPointer("nil"),
	}); err == nil {
		t.Error(err)
	}

}

func TestEventReaderCfgClone(t *testing.T) {
	ban := &EventReaderCfg{
		ID:             "2",
		Type:           "type",
		RunDelay:       1 * time.Minute,
		ConcurrentReqs: 5,
		PartialCommitFields: []*FCTemplate{
			{
				Tag:  "tag1",
				Type: "type1",
			},
			{
				Tag:  "tag2",
				Type: "type2",
			},
		},
		SourcePath:    "/",
		ProcessedPath: "/path",
		Tenant:        RSRParsers{},
		Timezone:      "time.Utc",
		Flags:         utils.FlagsWithParams{},
		Opts: &EventReaderOpts{
			PartialPath:                       utils.StringPointer("path"),
			PartialCSVFieldSeparator:          utils.StringPointer("/"),
			CSVLazyQuotes:                     utils.BoolPointer(false),
			AMQPQueueID:                       utils.StringPointer("id"),
			AMQPQueueIDProcessed:              utils.StringPointer("id"),
			AMQPConsumerTag:                   utils.StringPointer("tag"),
			AMQPExchange:                      utils.StringPointer("exchange"),
			AMQPExchangeType:                  utils.StringPointer("type"),
			AMQPRoutingKey:                    utils.StringPointer("key1"),
			AMQPExchangeProcessed:             utils.StringPointer("amq"),
			AMQPExchangeTypeProcessed:         utils.StringPointer("amqtype"),
			AMQPRoutingKeyProcessed:           utils.StringPointer("key"),
			KafkaTopic:                        utils.StringPointer("kafka"),
			KafkaMaxWait:                      utils.DurationPointer(1 * time.Minute),
			KafkaTopicProcessed:               utils.StringPointer("kafkaproc"),
			SQLDBName:                         utils.StringPointer("dbname"),
			SQLTableName:                      utils.StringPointer("tablename"),
			PgSSLMode:                         utils.StringPointer("sslmode"),
			SQLDBNameProcessed:                utils.StringPointer("dbnameproc"),
			SQLTableNameProcessed:             utils.StringPointer("tablenameproc"),
			PgSSLModeProcessed:                utils.StringPointer("sslproc"),
			AWSRegion:                         utils.StringPointer("eu"),
			AWSKey:                            utils.StringPointer("key"),
			AWSSecret:                         utils.StringPointer("secret"),
			AWSToken:                          utils.StringPointer("token"),
			AWSKeyProcessed:                   utils.StringPointer("secret"),
			AWSSecretProcessed:                utils.StringPointer("secret"),
			AWSTokenProcessed:                 utils.StringPointer("token"),
			SQSQueueID:                        utils.StringPointer("SQSQueue"),
			SQSQueueIDProcessed:               utils.StringPointer("SQSQueueId"),
			S3BucketID:                        utils.StringPointer("S3BucketID"),
			S3FolderPathProcessed:             utils.StringPointer("S3Path"),
			S3BucketIDProcessed:               utils.StringPointer("S3BucketProc"),
			NATSJetStream:                     utils.BoolPointer(false),
			NATSConsumerName:                  utils.StringPointer("user"),
			NATSQueueID:                       utils.StringPointer("id"),
			NATSJWTFile:                       utils.StringPointer("jwt"),
			NATSSeedFile:                      utils.StringPointer("seed"),
			NATSCertificateAuthority:          utils.StringPointer("authority"),
			NATSClientCertificate:             utils.StringPointer("certificate"),
			NATSClientKey:                     utils.StringPointer("key5"),
			NATSJetStreamMaxWait:              utils.DurationPointer(1 * time.Minute),
			NATSJetStreamProcessed:            utils.BoolPointer(true),
			NATSJWTFileProcessed:              utils.StringPointer("file"),
			NATSSeedFileProcessed:             utils.StringPointer("natseed"),
			NATSCertificateAuthorityProcessed: utils.StringPointer("natsauth"),
			NATSClientCertificateProcessed:    utils.StringPointer("natcertificate"),
			NATSClientKeyProcessed:            utils.StringPointer("natsprocess"),
			NATSJetStreamMaxWaitProcessed:     utils.DurationPointer(1 * time.Minute),
			AWSRegionProcessed:                utils.StringPointer("eu"),
			NATSSubjectProcessed:              utils.StringPointer("process"),
			KafkaGroupID:                      utils.StringPointer("groupId"),
		},
	}
	rcv := ban.Clone()
	if !reflect.DeepEqual(rcv, ban) {
		t.Errorf("expected %v received %v", utils.ToJSON(ban), utils.ToJSON(rcv))
	}

}
