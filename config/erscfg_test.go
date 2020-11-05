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
				ID:               utils.MetaDefault,
				Type:             utils.META_NONE,
				FieldSep:         ",",
				HeaderDefineChar: ":",
				RunDelay:         0,
				ConcurrentReqs:   1024,
				SourcePath:       "/var/spool/cgrates/ers/in",
				ProcessedPath:    "/var/spool/cgrates/ers/out",
				XmlRootPath:      utils.HierarchyPath{utils.EmptyString},
				Tenant:           nil,
				Timezone:         utils.EmptyString,
				Filters:          nil,
				Flags:            utils.FlagsWithParams{},
				Fields: []*FCTemplate{
					{Tag: utils.ToR, Path: utils.MetaCgreq + utils.NestingSep + utils.ToR, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.2", utils.INFIELD_SEP), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.OriginID, Path: utils.MetaCgreq + utils.NestingSep + utils.OriginID, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.3", utils.INFIELD_SEP), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.RequestType, Path: utils.MetaCgreq + utils.NestingSep + utils.RequestType, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.4", utils.INFIELD_SEP), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.Tenant, Path: utils.MetaCgreq + utils.NestingSep + utils.Tenant, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.6", utils.INFIELD_SEP), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.Category, Path: utils.MetaCgreq + utils.NestingSep + utils.Category, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.7", utils.INFIELD_SEP), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.Account, Path: utils.MetaCgreq + utils.NestingSep + utils.Account, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.8", utils.INFIELD_SEP), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.Subject, Path: utils.MetaCgreq + utils.NestingSep + utils.Subject, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.9", utils.INFIELD_SEP), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.Destination, Path: utils.MetaCgreq + utils.NestingSep + utils.Destination, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.10", utils.INFIELD_SEP), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.SetupTime, Path: utils.MetaCgreq + utils.NestingSep + utils.SetupTime, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.11", utils.INFIELD_SEP), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.AnswerTime, Path: utils.MetaCgreq + utils.NestingSep + utils.AnswerTime, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.12", utils.INFIELD_SEP), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.Usage, Path: utils.MetaCgreq + utils.NestingSep + utils.Usage, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.13", utils.INFIELD_SEP), Mandatory: true, Layout: time.RFC3339},
				},
				CacheDumpFields: make([]*FCTemplate, 0),
				Opts:            make(map[string]interface{}),
			},
			{
				ID:                       "file_reader1",
				Type:                     "*file_csv",
				FieldSep:                 ",",
				HeaderDefineChar:         ":",
				RunDelay:                 -1,
				ConcurrentReqs:           1024,
				SourcePath:               "/tmp/ers/in",
				ProcessedPath:            "/tmp/ers/out",
				XmlRootPath:              utils.HierarchyPath{utils.EmptyString},
				Tenant:                   NewRSRParsersMustCompile("~*req.Destination1", utils.INFIELD_SEP),
				Timezone:                 utils.EmptyString,
				Filters:                  []string{"randomFiletrs"},
				Flags:                    utils.FlagsWithParams{},
				FailedCallsPrefix:        "randomPrefix",
				PartialRecordCache:       time.Second,
				PartialCacheExpiryAction: "randomAction",
				Fields: []*FCTemplate{
					{Tag: utils.ToR, Path: utils.MetaCgreq + utils.NestingSep + utils.ToR, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.2", utils.INFIELD_SEP), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.OriginID, Path: utils.MetaCgreq + utils.NestingSep + utils.OriginID, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.3", utils.INFIELD_SEP), Mandatory: true, Layout: time.RFC3339},
				},
				CacheDumpFields: []*FCTemplate{
					{Tag: utils.ToR, Path: utils.MetaCgreq + utils.NestingSep + utils.ToR, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.2", utils.INFIELD_SEP), Mandatory: true, Layout: time.RFC3339},
				},
				Opts: map[string]interface{}{
					utils.MetaDefault: "randomVal",
				},
			},
		},
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
	jsonCfg, err := NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
	}
	eventReader := new(EventReaderCfg)
	if err = eventReader.loadFromJsonCfg(nil, jsonCfg.templates, jsonCfg.generalCfg.RSRSep); err != nil {
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
	if jsoncfg, err := NewDefaultCGRConfig(); err != nil {
		t.Error(err)
	} else if err = jsoncfg.ersCfg.loadFromJsonCfg(cfgJSON, jsoncfg.templates, jsoncfg.generalCfg.RSRSep, jsoncfg.dfltEvRdr, jsoncfg.generalCfg.RSRSep); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestEventReaderloadFromJsonCase2(t *testing.T) {
	cfgJSON := &ERsJsonCfg{
		Readers: &[]*EventReaderJsonCfg{
			{
				Partial_record_cache: utils.StringPointer("1ss"),
			},
		},
	}
	expected := "time: unknown unit \"ss\" in duration \"1ss\""
	if jsoncfg, err := NewDefaultCGRConfig(); err != nil {
		t.Error(err)
	} else if err = jsoncfg.ersCfg.loadFromJsonCfg(cfgJSON, jsoncfg.templates, jsoncfg.generalCfg.RSRSep, jsoncfg.dfltEvRdr, jsoncfg.generalCfg.RSRSep); err == nil || err.Error() != expected {
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
	if jsoncfg, err := NewDefaultCGRConfig(); err != nil {
		t.Error(err)
	} else if err = jsoncfg.ersCfg.loadFromJsonCfg(cfgJSON, jsoncfg.templates, jsoncfg.generalCfg.RSRSep, jsoncfg.dfltEvRdr, jsoncfg.generalCfg.RSRSep); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestERSLoadFromjsonCfg(t *testing.T) {
	expectedERsCfg := &ERsCfg{
		Enabled:       true,
		SessionSConns: []string{"conn1", "conn3"},
		Readers: []*EventReaderCfg{
			{
				ID:               utils.MetaDefault,
				Type:             utils.META_NONE,
				FieldSep:         ",",
				HeaderDefineChar: ":",
				RunDelay:         0,
				ConcurrentReqs:   1024,
				SourcePath:       "/var/spool/cgrates/ers/in",
				ProcessedPath:    "/var/spool/cgrates/ers/out",
				XmlRootPath:      utils.HierarchyPath{utils.EmptyString},
				Tenant:           nil,
				Timezone:         utils.EmptyString,
				Filters:          []string{},
				Flags:            utils.FlagsWithParams{},
				Fields: []*FCTemplate{
					{Tag: utils.ToR, Path: utils.MetaCgreq + utils.NestingSep + utils.ToR, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.2", utils.INFIELD_SEP), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.OriginID, Path: utils.MetaCgreq + utils.NestingSep + utils.OriginID, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.3", utils.INFIELD_SEP), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.RequestType, Path: utils.MetaCgreq + utils.NestingSep + utils.RequestType, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.4", utils.INFIELD_SEP), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.Tenant, Path: utils.MetaCgreq + utils.NestingSep + utils.Tenant, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.6", utils.INFIELD_SEP), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.Category, Path: utils.MetaCgreq + utils.NestingSep + utils.Category, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.7", utils.INFIELD_SEP), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.Account, Path: utils.MetaCgreq + utils.NestingSep + utils.Account, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.8", utils.INFIELD_SEP), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.Subject, Path: utils.MetaCgreq + utils.NestingSep + utils.Subject, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.9", utils.INFIELD_SEP), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.Destination, Path: utils.MetaCgreq + utils.NestingSep + utils.Destination, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.10", utils.INFIELD_SEP), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.SetupTime, Path: utils.MetaCgreq + utils.NestingSep + utils.SetupTime, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.11", utils.INFIELD_SEP), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.AnswerTime, Path: utils.MetaCgreq + utils.NestingSep + utils.AnswerTime, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.12", utils.INFIELD_SEP), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.Usage, Path: utils.MetaCgreq + utils.NestingSep + utils.Usage, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.13", utils.INFIELD_SEP), Mandatory: true, Layout: time.RFC3339},
				},
				CacheDumpFields: make([]*FCTemplate, 0),
				Opts:            make(map[string]interface{}),
			},
			{
				ID:               "file_reader1",
				Type:             utils.MetaFileCSV,
				FieldSep:         ",",
				HeaderDefineChar: ":",
				RunDelay:         -1,
				ConcurrentReqs:   1024,
				SourcePath:       "/tmp/ers/in",
				ProcessedPath:    "/tmp/ers/out",
				XmlRootPath:      utils.HierarchyPath{utils.EmptyString},
				Tenant:           nil,
				Timezone:         utils.EmptyString,
				Filters:          nil,
				Flags:            utils.FlagsWithParams{},
				Fields: []*FCTemplate{
					{Tag: utils.ToR, Path: utils.MetaCgreq + utils.NestingSep + utils.ToR, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.2", utils.INFIELD_SEP), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.OriginID, Path: utils.MetaCgreq + utils.NestingSep + utils.OriginID, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.3", utils.INFIELD_SEP), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.RequestType, Path: utils.MetaCgreq + utils.NestingSep + utils.RequestType, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.4", utils.INFIELD_SEP), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.Tenant, Path: utils.MetaCgreq + utils.NestingSep + utils.Tenant, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.6", utils.INFIELD_SEP), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.Category, Path: utils.MetaCgreq + utils.NestingSep + utils.Category, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.7", utils.INFIELD_SEP), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.Account, Path: utils.MetaCgreq + utils.NestingSep + utils.Account, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.8", utils.INFIELD_SEP), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.Subject, Path: utils.MetaCgreq + utils.NestingSep + utils.Subject, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.9", utils.INFIELD_SEP), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.Destination, Path: utils.MetaCgreq + utils.NestingSep + utils.Destination, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.10", utils.INFIELD_SEP), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.SetupTime, Path: utils.MetaCgreq + utils.NestingSep + utils.SetupTime, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.11", utils.INFIELD_SEP), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.AnswerTime, Path: utils.MetaCgreq + utils.NestingSep + utils.AnswerTime, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.12", utils.INFIELD_SEP), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.Usage, Path: utils.MetaCgreq + utils.NestingSep + utils.Usage, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.13", utils.INFIELD_SEP), Mandatory: true, Layout: time.RFC3339},
				},
				CacheDumpFields: make([]*FCTemplate, 0),
				Opts:            make(map[string]interface{}),
			},
		},
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
	if jsonCfg, err := NewDefaultCGRConfig(); err != nil {
		t.Error(err)
	} else if err = jsonCfg.ersCfg.loadFromJsonCfg(cfgJSON, jsonCfg.templates, jsonCfg.generalCfg.RSRSep, jsonCfg.dfltEvRdr, jsonCfg.generalCfg.RSRSep); err != nil {
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
	if jsonCfg, err := NewDefaultCGRConfig(); err != nil {
		t.Error(err)
	} else if err = jsonCfg.ersCfg.loadFromJsonCfg(cfgJSON, jsonCfg.templates, jsonCfg.generalCfg.RSRSep, jsonCfg.dfltEvRdr, jsonCfg.generalCfg.RSRSep); err == nil || err.Error() != expected {
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
	if jsonCfg, err := NewDefaultCGRConfig(); err != nil {
		t.Error(err)
	} else if err = jsonCfg.ersCfg.loadFromJsonCfg(cfgJSON, jsonCfg.templates, jsonCfg.generalCfg.RSRSep, jsonCfg.dfltEvRdr, jsonCfg.generalCfg.RSRSep); err == nil || err.Error() != expected {
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
	if jsonCfg, err := NewDefaultCGRConfig(); err != nil {
		t.Error(err)
	} else if err = jsonCfg.ersCfg.loadFromJsonCfg(cfgJSON, jsonCfg.templates, jsonCfg.generalCfg.RSRSep, jsonCfg.dfltEvRdr, jsonCfg.generalCfg.RSRSep); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestERSloadFromJsonCase3(t *testing.T) {
	cfgJSON := &ERsJsonCfg{
		Enabled:        utils.BoolPointer(true),
		Sessions_conns: &[]string{"*conn1"},
		Readers: &[]*EventReaderJsonCfg{
			{
				Id:                          utils.StringPointer("file_reader1"),
				Type:                        utils.StringPointer(utils.MetaFileCSV),
				Row_length:                  utils.IntPointer(5),
				Field_separator:             utils.StringPointer(","),
				Header_define_character:     utils.StringPointer(":"),
				Run_delay:                   utils.StringPointer("-1"),
				Concurrent_requests:         utils.IntPointer(1024),
				Source_path:                 utils.StringPointer("/tmp/ers/in"),
				Processed_path:              utils.StringPointer("/tmp/ers/out"),
				Xml_root_path:               utils.StringPointer(""),
				Tenant:                      nil,
				Timezone:                    utils.StringPointer(""),
				Filters:                     nil,
				Flags:                       &[]string{},
				Failed_calls_prefix:         nil,
				Partial_cache_expiry_action: utils.StringPointer(""),
				Partial_record_cache:        utils.StringPointer("0"),
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
				ID:               utils.MetaDefault,
				Type:             utils.META_NONE,
				FieldSep:         ",",
				HeaderDefineChar: ":",
				RunDelay:         0,
				ConcurrentReqs:   1024,
				SourcePath:       "/var/spool/cgrates/ers/in",
				ProcessedPath:    "/var/spool/cgrates/ers/out",
				XmlRootPath:      utils.HierarchyPath{utils.EmptyString},
				Tenant:           nil,
				Timezone:         utils.EmptyString,
				Filters:          []string{},
				Flags:            utils.FlagsWithParams{},
				Fields: []*FCTemplate{
					{Tag: utils.ToR, Path: utils.MetaCgreq + utils.NestingSep + utils.ToR, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.2", utils.INFIELD_SEP), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.OriginID, Path: utils.MetaCgreq + utils.NestingSep + utils.OriginID, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.3", utils.INFIELD_SEP), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.RequestType, Path: utils.MetaCgreq + utils.NestingSep + utils.RequestType, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.4", utils.INFIELD_SEP), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.Tenant, Path: utils.MetaCgreq + utils.NestingSep + utils.Tenant, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.6", utils.INFIELD_SEP), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.Category, Path: utils.MetaCgreq + utils.NestingSep + utils.Category, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.7", utils.INFIELD_SEP), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.Account, Path: utils.MetaCgreq + utils.NestingSep + utils.Account, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.8", utils.INFIELD_SEP), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.Subject, Path: utils.MetaCgreq + utils.NestingSep + utils.Subject, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.9", utils.INFIELD_SEP), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.Destination, Path: utils.MetaCgreq + utils.NestingSep + utils.Destination, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.10", utils.INFIELD_SEP), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.SetupTime, Path: utils.MetaCgreq + utils.NestingSep + utils.SetupTime, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.11", utils.INFIELD_SEP), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.AnswerTime, Path: utils.MetaCgreq + utils.NestingSep + utils.AnswerTime, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.12", utils.INFIELD_SEP), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.Usage, Path: utils.MetaCgreq + utils.NestingSep + utils.Usage, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.13", utils.INFIELD_SEP), Mandatory: true, Layout: time.RFC3339},
				},
				CacheDumpFields: make([]*FCTemplate, 0),
				Opts:            make(map[string]interface{}),
			},
			{
				ID:                       "file_reader1",
				Type:                     utils.MetaFileCSV,
				RowLength:                5,
				FieldSep:                 ",",
				HeaderDefineChar:         ":",
				RunDelay:                 -1,
				ConcurrentReqs:           1024,
				SourcePath:               "/tmp/ers/in",
				ProcessedPath:            "/tmp/ers/out",
				XmlRootPath:              utils.HierarchyPath{utils.EmptyString},
				Tenant:                   nil,
				Timezone:                 utils.EmptyString,
				Filters:                  nil,
				Flags:                    utils.FlagsWithParams{},
				PartialCacheExpiryAction: "",
				PartialRecordCache:       0,
				Fields: []*FCTemplate{
					{
						Tag:    utils.CGRID,
						Path:   "*exp.CGRID",
						Type:   utils.MetaVariable,
						Layout: time.RFC3339,
					},
				},
				CacheDumpFields: make([]*FCTemplate, 0),
				Opts:            make(map[string]interface{}),
			},
		},
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
	if jsonCfg, err := NewDefaultCGRConfig(); err != nil {
		t.Error(err)
	} else if err = jsonCfg.ersCfg.loadFromJsonCfg(cfgJSON, msgTemplates, jsonCfg.generalCfg.RSRSep, jsonCfg.dfltEvRdr, jsonCfg.generalCfg.RSRSep); err != nil {
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
				Id:                          utils.StringPointer("file_reader1"),
				Type:                        utils.StringPointer(utils.MetaFileCSV),
				Row_length:                  utils.IntPointer(5),
				Field_separator:             utils.StringPointer(","),
				Header_define_character:     utils.StringPointer(":"),
				Run_delay:                   utils.StringPointer("-1"),
				Concurrent_requests:         utils.IntPointer(1024),
				Source_path:                 utils.StringPointer("/tmp/ers/in"),
				Processed_path:              utils.StringPointer("/tmp/ers/out"),
				Xml_root_path:               utils.StringPointer(""),
				Tenant:                      nil,
				Timezone:                    utils.StringPointer(""),
				Filters:                     nil,
				Flags:                       &[]string{},
				Failed_calls_prefix:         nil,
				Partial_cache_expiry_action: utils.StringPointer(""),
				Partial_record_cache:        utils.StringPointer("0"),
				Fields:                      &[]*FcTemplateJsonCfg{},
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
				ID:               utils.MetaDefault,
				Type:             utils.META_NONE,
				FieldSep:         ",",
				HeaderDefineChar: ":",
				RunDelay:         0,
				ConcurrentReqs:   1024,
				SourcePath:       "/var/spool/cgrates/ers/in",
				ProcessedPath:    "/var/spool/cgrates/ers/out",
				XmlRootPath:      utils.HierarchyPath{utils.EmptyString},
				Tenant:           nil,
				Timezone:         utils.EmptyString,
				Filters:          []string{},
				Flags:            utils.FlagsWithParams{},
				Fields: []*FCTemplate{
					{Tag: utils.ToR, Path: utils.MetaCgreq + utils.NestingSep + utils.ToR, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.2", utils.INFIELD_SEP), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.OriginID, Path: utils.MetaCgreq + utils.NestingSep + utils.OriginID, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.3", utils.INFIELD_SEP), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.RequestType, Path: utils.MetaCgreq + utils.NestingSep + utils.RequestType, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.4", utils.INFIELD_SEP), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.Tenant, Path: utils.MetaCgreq + utils.NestingSep + utils.Tenant, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.6", utils.INFIELD_SEP), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.Category, Path: utils.MetaCgreq + utils.NestingSep + utils.Category, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.7", utils.INFIELD_SEP), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.Account, Path: utils.MetaCgreq + utils.NestingSep + utils.Account, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.8", utils.INFIELD_SEP), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.Subject, Path: utils.MetaCgreq + utils.NestingSep + utils.Subject, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.9", utils.INFIELD_SEP), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.Destination, Path: utils.MetaCgreq + utils.NestingSep + utils.Destination, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.10", utils.INFIELD_SEP), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.SetupTime, Path: utils.MetaCgreq + utils.NestingSep + utils.SetupTime, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.11", utils.INFIELD_SEP), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.AnswerTime, Path: utils.MetaCgreq + utils.NestingSep + utils.AnswerTime, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.12", utils.INFIELD_SEP), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.Usage, Path: utils.MetaCgreq + utils.NestingSep + utils.Usage, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.13", utils.INFIELD_SEP), Mandatory: true, Layout: time.RFC3339},
				},
				CacheDumpFields: make([]*FCTemplate, 0),
				Opts:            make(map[string]interface{}),
			},
			{
				ID:                       "file_reader1",
				Type:                     utils.MetaFileCSV,
				RowLength:                5,
				FieldSep:                 ",",
				HeaderDefineChar:         ":",
				RunDelay:                 -1,
				ConcurrentReqs:           1024,
				SourcePath:               "/tmp/ers/in",
				ProcessedPath:            "/tmp/ers/out",
				XmlRootPath:              utils.HierarchyPath{utils.EmptyString},
				Tenant:                   nil,
				Timezone:                 utils.EmptyString,
				Filters:                  nil,
				Flags:                    utils.FlagsWithParams{},
				PartialCacheExpiryAction: "",
				PartialRecordCache:       0,
				Fields:                   []*FCTemplate{},
				CacheDumpFields: []*FCTemplate{
					{
						Tag:   "OrderID",
						Path:  "*exp.OrderID",
						Type:  "*variable",
						Value: NewRSRParsersMustCompile("~*req.OrderID", utils.INFIELD_SEP),
					},
				},
				Opts: make(map[string]interface{}),
			},
		},
	}
	msgTemplates := map[string][]*FCTemplate{
		"randomTemplate": {
			{
				Tag:   "OrderID",
				Path:  "*exp.OrderID",
				Type:  "*variable",
				Value: NewRSRParsersMustCompile("~*req.OrderID", utils.INFIELD_SEP),
			},
		},
	}
	for _, v := range expectedERsCfg.Readers[0].CacheDumpFields {
		v.ComputePath()
	}
	for _, v := range expectedERsCfg.Readers[0].Fields {
		v.ComputePath()
	}
	if jsonCfg, err := NewDefaultCGRConfig(); err != nil {
		t.Error(err)
	} else if err = jsonCfg.ersCfg.loadFromJsonCfg(cfgJSON, msgTemplates, jsonCfg.generalCfg.RSRSep, jsonCfg.dfltEvRdr, jsonCfg.generalCfg.RSRSep); err != nil {
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
	if jsonCfg, err := NewDefaultCGRConfig(); err != nil {
		t.Error(err)
	} else if err = jsonCfg.ersCfg.loadFromJsonCfg(cfgJSON, jsonCfg.templates, jsonCfg.generalCfg.RSRSep, jsonCfg.dfltEvRdr, jsonCfg.generalCfg.RSRSep); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestEventReaderSameID(t *testing.T) {
	expectedERsCfg := &ERsCfg{
		Enabled:       true,
		SessionSConns: []string{"conn1"},
		Readers: []*EventReaderCfg{
			{
				ID:               utils.MetaDefault,
				Type:             utils.META_NONE,
				FieldSep:         ",",
				HeaderDefineChar: ":",
				RunDelay:         0,
				ConcurrentReqs:   1024,
				SourcePath:       "/var/spool/cgrates/ers/in",
				ProcessedPath:    "/var/spool/cgrates/ers/out",
				XmlRootPath:      utils.HierarchyPath{utils.EmptyString},
				Tenant:           nil,
				Timezone:         utils.EmptyString,
				Filters:          []string{},
				Flags:            utils.FlagsWithParams{},
				Fields: []*FCTemplate{
					{Tag: utils.ToR, Path: utils.MetaCgreq + utils.NestingSep + utils.ToR, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.2", utils.INFIELD_SEP), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.OriginID, Path: utils.MetaCgreq + utils.NestingSep + utils.OriginID, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.3", utils.INFIELD_SEP), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.RequestType, Path: utils.MetaCgreq + utils.NestingSep + utils.RequestType, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.4", utils.INFIELD_SEP), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.Tenant, Path: utils.MetaCgreq + utils.NestingSep + utils.Tenant, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.6", utils.INFIELD_SEP), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.Category, Path: utils.MetaCgreq + utils.NestingSep + utils.Category, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.7", utils.INFIELD_SEP), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.Account, Path: utils.MetaCgreq + utils.NestingSep + utils.Account, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.8", utils.INFIELD_SEP), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.Subject, Path: utils.MetaCgreq + utils.NestingSep + utils.Subject, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.9", utils.INFIELD_SEP), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.Destination, Path: utils.MetaCgreq + utils.NestingSep + utils.Destination, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.10", utils.INFIELD_SEP), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.SetupTime, Path: utils.MetaCgreq + utils.NestingSep + utils.SetupTime, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.11", utils.INFIELD_SEP), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.AnswerTime, Path: utils.MetaCgreq + utils.NestingSep + utils.AnswerTime, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.12", utils.INFIELD_SEP), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.Usage, Path: utils.MetaCgreq + utils.NestingSep + utils.Usage, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.13", utils.INFIELD_SEP), Mandatory: true, Layout: time.RFC3339},
				},
				CacheDumpFields: make([]*FCTemplate, 0),
				Opts:            make(map[string]interface{}),
			},
			{
				ID:               "file_reader1",
				Type:             utils.MetaFileCSV,
				RowLength:        5,
				FieldSep:         ",",
				HeaderDefineChar: ":",
				RunDelay:         -1,
				ConcurrentReqs:   1024,
				SourcePath:       "/tmp/ers/in",
				ProcessedPath:    "/tmp/ers/out",
				XmlRootPath:      utils.HierarchyPath{utils.EmptyString},
				Tenant:           nil,
				Timezone:         utils.EmptyString,
				Filters:          nil,
				Flags:            utils.FlagsWithParams{},
				Fields: []*FCTemplate{
					{Tag: "CustomTag2", Path: "CustomPath2", Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("CustomValue2", utils.INFIELD_SEP), Mandatory: true, Layout: time.RFC3339},
				},
				CacheDumpFields: make([]*FCTemplate, 0),
				Opts:            make(map[string]interface{}),
			},
		},
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
				utils.FiltersCfg:                  []string{},
				utils.FlagsCfg:                    []string{},
				utils.IDCfg:                       "*default",
				utils.PartialRecordCacheCfg:       "0",
				utils.ProcessedPathCfg:            "/var/spool/cgrates/ers/out",
				utils.RowLengthCfg:                0,
				utils.RunDelayCfg:                 "0",
				utils.PartialCacheExpiryActionCfg: "",
				utils.SourcePathCfg:               "/var/spool/cgrates/ers/in",
				utils.TenantCfg:                   "",
				utils.TimezoneCfg:                 "",
				utils.XMLRootPathCfg:              []string{""},
				utils.CacheDumpFieldsCfg:          []map[string]interface{}{},
				utils.ConcurrentRequestsCfg:       1024,
				utils.TypeCfg:                     "*none",
				utils.FailedCallsPrefixCfg:        "",
				utils.FieldSepCfg:                 ",",
				utils.HeaderDefCharCfg:            ":",
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
				utils.OptsCfg: make(map[string]interface{}),
			},
			{
				utils.CacheDumpFieldsCfg:    []map[string]interface{}{},
				utils.ConcurrentRequestsCfg: 1024,
				utils.TypeCfg:               "*file_csv",
				utils.FailedCallsPrefixCfg:  "",
				utils.FieldSepCfg:           ",",
				utils.HeaderDefCharCfg:      ":",
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
				utils.FiltersCfg:                  []string{},
				utils.FlagsCfg:                    []string{},
				utils.IDCfg:                       "file_reader1",
				utils.PartialRecordCacheCfg:       "0",
				utils.ProcessedPathCfg:            "/tmp/ers/out",
				utils.RowLengthCfg:                0,
				utils.RunDelayCfg:                 "-1",
				utils.PartialCacheExpiryActionCfg: "",
				utils.SourcePathCfg:               "/tmp/ers/in",
				utils.TenantCfg:                   "~*req.Destination1",
				utils.TimezoneCfg:                 "",
				utils.XMLRootPathCfg:              []string{""},
				utils.OptsCfg:                     make(map[string]interface{}),
			},
		},
	}
	if cfg, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cfg.ersCfg.AsMapInterface(utils.EmptyString); !reflect.DeepEqual(eMap, rcv) {
		t.Errorf("\nExpected: %+v\nReceived: %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
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
			},
		],
	}
}`
	eMap := map[string]interface{}{
		utils.EnabledCfg:       true,
		utils.SessionSConnsCfg: []string{"conn1", "conn3"},
		utils.ReadersCfg: []map[string]interface{}{
			{
				utils.FiltersCfg:                  []string{},
				utils.FlagsCfg:                    []string{},
				utils.IDCfg:                       "*default",
				utils.PartialRecordCacheCfg:       "0",
				utils.ProcessedPathCfg:            "/var/spool/cgrates/ers/out",
				utils.RowLengthCfg:                0,
				utils.RunDelayCfg:                 "0",
				utils.PartialCacheExpiryActionCfg: "",
				utils.SourcePathCfg:               "/var/spool/cgrates/ers/in",
				utils.TenantCfg:                   "",
				utils.TimezoneCfg:                 "",
				utils.XMLRootPathCfg:              []string{""},
				utils.CacheDumpFieldsCfg:          []map[string]interface{}{},
				utils.ConcurrentRequestsCfg:       1024,
				utils.TypeCfg:                     "*none",
				utils.FailedCallsPrefixCfg:        "",
				utils.FieldSepCfg:                 ",",
				utils.HeaderDefCharCfg:            ":",
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
				utils.OptsCfg: make(map[string]interface{}),
			},
			{
				utils.CacheDumpFieldsCfg: []map[string]interface{}{
					{utils.MandatoryCfg: true, utils.PathCfg: "*cgreq.ToR", utils.TagCfg: "ToR", utils.TypeCfg: "*variable", utils.ValueCfg: "~*req.2"},
				},
				utils.ConcurrentRequestsCfg: 1024,
				utils.TypeCfg:               "*file_csv",
				utils.FailedCallsPrefixCfg:  "",
				utils.FieldSepCfg:           ",",
				utils.HeaderDefCharCfg:      ":",
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
				utils.FiltersCfg:                  []string{"randomFilter"},
				utils.FlagsCfg:                    []string{"randomFlag"},
				utils.IDCfg:                       "file_reader1",
				utils.ProcessedPathCfg:            "/tmp/ers/out",
				utils.RowLengthCfg:                0,
				utils.RunDelayCfg:                 "10s",
				utils.PartialRecordCacheCfg:       "1s",
				utils.PartialCacheExpiryActionCfg: "",
				utils.SourcePathCfg:               "/tmp/ers/in",
				utils.TenantCfg:                   "~*req.Destination1",
				utils.TimezoneCfg:                 "",
				utils.XMLRootPathCfg:              []string{""},
				utils.OptsCfg:                     make(map[string]interface{}),
			},
		},
	}
	if cfg, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cfg.ersCfg.AsMapInterface(utils.EmptyString); !reflect.DeepEqual(eMap, rcv) {
		t.Errorf("\nExpected: %+v\nReceived: %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}
}

func TestERsloadFromJsonCfg(t *testing.T) {
	cfgJSON := &ERsJsonCfg{
		Enabled:        utils.BoolPointer(true),
		Sessions_conns: &[]string{"*conn1"},
		Readers: &[]*EventReaderJsonCfg{
			{
				Id:                          utils.StringPointer("file_reader1"),
				Type:                        utils.StringPointer(utils.MetaFileCSV),
				Row_length:                  utils.IntPointer(5),
				Field_separator:             utils.StringPointer(","),
				Header_define_character:     utils.StringPointer(":"),
				Run_delay:                   utils.StringPointer("-1"),
				Concurrent_requests:         utils.IntPointer(1024),
				Source_path:                 utils.StringPointer("/tmp/ers/in"),
				Processed_path:              utils.StringPointer("/tmp/ers/out"),
				Xml_root_path:               utils.StringPointer(""),
				Tenant:                      nil,
				Timezone:                    utils.StringPointer(""),
				Filters:                     nil,
				Flags:                       &[]string{},
				Failed_calls_prefix:         nil,
				Partial_cache_expiry_action: utils.StringPointer(""),
				Partial_record_cache:        utils.StringPointer("0"),
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
				ID:               utils.MetaDefault,
				Type:             utils.META_NONE,
				FieldSep:         ",",
				HeaderDefineChar: ":",
				RunDelay:         0,
				ConcurrentReqs:   1024,
				SourcePath:       "/var/spool/cgrates/ers/in",
				ProcessedPath:    "/var/spool/cgrates/ers/out",
				XmlRootPath:      utils.HierarchyPath{utils.EmptyString},
				Tenant:           nil,
				Timezone:         utils.EmptyString,
				Filters:          []string{},
				Flags:            utils.FlagsWithParams{},
				Fields: []*FCTemplate{
					{Tag: utils.ToR, Path: utils.MetaCgreq + utils.NestingSep + utils.ToR, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.2", utils.INFIELD_SEP), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.OriginID, Path: utils.MetaCgreq + utils.NestingSep + utils.OriginID, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.3", utils.INFIELD_SEP), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.RequestType, Path: utils.MetaCgreq + utils.NestingSep + utils.RequestType, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.4", utils.INFIELD_SEP), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.Tenant, Path: utils.MetaCgreq + utils.NestingSep + utils.Tenant, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.6", utils.INFIELD_SEP), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.Category, Path: utils.MetaCgreq + utils.NestingSep + utils.Category, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.7", utils.INFIELD_SEP), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.Account, Path: utils.MetaCgreq + utils.NestingSep + utils.Account, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.8", utils.INFIELD_SEP), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.Subject, Path: utils.MetaCgreq + utils.NestingSep + utils.Subject, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.9", utils.INFIELD_SEP), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.Destination, Path: utils.MetaCgreq + utils.NestingSep + utils.Destination, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.10", utils.INFIELD_SEP), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.SetupTime, Path: utils.MetaCgreq + utils.NestingSep + utils.SetupTime, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.11", utils.INFIELD_SEP), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.AnswerTime, Path: utils.MetaCgreq + utils.NestingSep + utils.AnswerTime, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.12", utils.INFIELD_SEP), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.Usage, Path: utils.MetaCgreq + utils.NestingSep + utils.Usage, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.13", utils.INFIELD_SEP), Mandatory: true, Layout: time.RFC3339},
				},
				CacheDumpFields: make([]*FCTemplate, 0),
				Opts:            make(map[string]interface{}),
			},
			{
				ID:                       "file_reader1",
				Type:                     utils.MetaFileCSV,
				RowLength:                5,
				FieldSep:                 ",",
				HeaderDefineChar:         ":",
				RunDelay:                 -1,
				ConcurrentReqs:           1024,
				SourcePath:               "/tmp/ers/in",
				ProcessedPath:            "/tmp/ers/out",
				XmlRootPath:              utils.HierarchyPath{utils.EmptyString},
				Tenant:                   nil,
				Timezone:                 utils.EmptyString,
				Filters:                  nil,
				Flags:                    utils.FlagsWithParams{},
				PartialCacheExpiryAction: "",
				PartialRecordCache:       0,
				Fields: []*FCTemplate{
					{
						Tag:    utils.CGRID,
						Path:   "*exp.CGRID",
						Type:   utils.MetaVariable,
						Layout: time.RFC3339,
					},
					{Tag: "CustomTag2", Path: "CustomPath2", Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("CustomValue2", utils.INFIELD_SEP), Mandatory: true, Layout: time.RFC3339},
				},
				CacheDumpFields: make([]*FCTemplate, 0),
				Opts:            make(map[string]interface{}),
			},
		},
	}
	for _, profile := range expectedERsCfg.Readers {
		for _, v := range profile.Fields {
			v.ComputePath()
		}
		for _, v := range profile.CacheDumpFields {
			v.ComputePath()
		}
	}
	if cfgCgr, err := NewDefaultCGRConfig(); err != nil {
		t.Error(err)
	} else if err := cfgCgr.ersCfg.loadFromJsonCfg(cfgJSON, cfgCgr.templates, cfgCgr.generalCfg.RSRSep, cfgCgr.dfltEvRdr, cfgCgr.generalCfg.RSRSep); err != nil {
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
	if cfgCgr, err := NewDefaultCGRConfig(); err != nil {
		t.Error(err)
	} else if dft := cfgCgr.EEsCfg().GetDefaultExporter(); dft == nil || dft.ID != utils.MetaDefault {
		t.Fatalf("Unexpected default cfg returned: %s", utils.ToJSON(dft))
	}
}
