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

func TestEventRedearClone(t *testing.T) {
	orig := &EventReaderCfg{
		ID:       utils.MetaDefault,
		Type:     "RandomType",
		FieldSep: ",",
		Filters:  []string{"Filter1", "Filter2"},
		Tenant:   NewRSRParsersMustCompile("cgrates.org", utils.INFIELD_SEP),
		Fields: []*FCTemplate{
			{
				Tag:       "ToR",
				Path:      "ToR",
				Type:      "*composed",
				Value:     NewRSRParsersMustCompile("~*req.2", utils.INFIELD_SEP),
				Mandatory: true,
			},
			{
				Tag:       "RandomField",
				Path:      "RandomField",
				Type:      "*composed",
				Value:     NewRSRParsersMustCompile("Test", utils.INFIELD_SEP),
				Mandatory: true,
			},
		},
		CacheDumpFields: make([]*FCTemplate, 0),
		Opts:            make(map[string]interface{}),
	}
	for _, v := range orig.Fields {
		v.ComputePath()
	}
	cloned := orig.Clone()
	if !reflect.DeepEqual(cloned, orig) {
		t.Errorf("expected: %s \n,received: %s", utils.ToJSON(orig), utils.ToJSON(cloned))
	}
	initialOrig := &EventReaderCfg{
		ID:       utils.MetaDefault,
		Type:     "RandomType",
		FieldSep: ",",
		Filters:  []string{"Filter1", "Filter2"},
		Tenant:   NewRSRParsersMustCompile("cgrates.org", utils.INFIELD_SEP),
		Fields: []*FCTemplate{
			{
				Tag:       "ToR",
				Path:      "ToR",
				Type:      "*composed",
				Value:     NewRSRParsersMustCompile("~*req.2", utils.INFIELD_SEP),
				Mandatory: true,
			},
			{
				Tag:       "RandomField",
				Path:      "RandomField",
				Type:      "*composed",
				Value:     NewRSRParsersMustCompile("Test", utils.INFIELD_SEP),
				Mandatory: true,
			},
		},
		CacheDumpFields: make([]*FCTemplate, 0),
		Opts:            make(map[string]interface{}),
	}
	for _, v := range initialOrig.Fields {
		v.ComputePath()
	}
	orig.Filters = []string{"SingleFilter"}
	orig.Fields = []*FCTemplate{
		{
			Tag:       "ToR",
			Path:      "ToR",
			Type:      "*composed",
			Value:     NewRSRParsersMustCompile("~2", utils.INFIELD_SEP),
			Mandatory: true,
		},
	}
	if !reflect.DeepEqual(cloned, initialOrig) {
		t.Errorf("expected: %s \n,received: %s", utils.ToJSON(initialOrig), utils.ToJSON(cloned))
	}
}

func TestEventReaderLoadFromJSON(t *testing.T) {
	expectedERsCfg := &ERsCfg{
		Enabled:       true,
		SessionSConns: []string{"conn1", "conn3"},
		Readers: []*EventReaderCfg{
			{
				ID:               utils.MetaDefault,
				Type:             utils.META_NONE,
				FieldSep:         ",",
				HeaderDefineChar: ":",
				RunDelay:         time.Duration(0),
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
				RunDelay:         time.Duration(-1),
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

	if cfg, err := NewCGRConfigFromJsonStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedERsCfg, cfg.ersCfg) {
		t.Errorf("Expected: %+v ,\n recived: %+v", utils.ToJSON(expectedERsCfg), utils.ToJSON(cfg.ersCfg))
	}

}

func TestEventReaderSanitization(t *testing.T) {
	cfgJSONStr := `{
"ers": {
	"enabled": true,
	"readers": [
		{
			"id": "file_reader1",
			"run_delay":  "-1",
			"type": "*file_csv",
			"source_path": "/tmp/ers/in",
			"processed_path": "/tmp/ers/out",
		},
	],
}
}`

	if _, err := NewCGRConfigFromJsonStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
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
				RunDelay:         time.Duration(0),
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
				RunDelay:         time.Duration(-1),
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

	if cfg, err := NewCGRConfigFromJsonStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedERsCfg, cfg.ersCfg) {
		t.Errorf("Expected: %+v ,\n recived: %+v", utils.ToJSON(expectedERsCfg), utils.ToJSON(cfg.ersCfg))
	}

}

func TestERsCfgAsMapInterface(t *testing.T) {
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
	var filters []string
	eMap := map[string]interface{}{
		"enabled":        true,
		"sessions_conns": []string{"conn1", "conn3"},
		"readers": []map[string]interface{}{
			{
				"filters":                     []string{},
				"flags":                       filters,
				"id":                          "*default",
				"partial_record_cache":        "0",
				"processed_path":              "/var/spool/cgrates/ers/out",
				"row_length":                  0,
				"run_delay":                   "0",
				"partial_cache_expiry_action": "",
				"source_path":                 "/var/spool/cgrates/ers/in",
				"tenant":                      "",
				"timezone":                    "",
				"xml_root_path":               []string{""},
				"cache_dump_fields":           []map[string]interface{}{},
				"concurrent_requests":         1024,
				"type":                        "*none",
				"failed_calls_prefix":         "",
				"field_separator":             ",",
				utils.HeaderDefCharCfg:        ":",
				"fields": []map[string]interface{}{
					{"mandatory": true, "path": "*cgreq.ToR", "tag": "ToR", "type": "*variable", "value": "~*req.2"},
					{"mandatory": true, "path": "*cgreq.OriginID", "tag": "OriginID", "type": "*variable", "value": "~*req.3"},
					{"mandatory": true, "path": "*cgreq.RequestType", "tag": "RequestType", "type": "*variable", "value": "~*req.4"},
					{"mandatory": true, "path": "*cgreq.Tenant", "tag": "Tenant", "type": "*variable", "value": "~*req.6"},
					{"mandatory": true, "path": "*cgreq.Category", "tag": "Category", "type": "*variable", "value": "~*req.7"},
					{"mandatory": true, "path": "*cgreq.Account", "tag": "Account", "type": "*variable", "value": "~*req.8"},
					{"mandatory": true, "path": "*cgreq.Subject", "tag": "Subject", "type": "*variable", "value": "~*req.9"},
					{"mandatory": true, "path": "*cgreq.Destination", "tag": "Destination", "type": "*variable", "value": "~*req.10"},
					{"mandatory": true, "path": "*cgreq.SetupTime", "tag": "SetupTime", "type": "*variable", "value": "~*req.11"},
					{"mandatory": true, "path": "*cgreq.AnswerTime", "tag": "AnswerTime", "type": "*variable", "value": "~*req.12"},
					{"mandatory": true, "path": "*cgreq.Usage", "tag": "Usage", "type": "*variable", "value": "~*req.13"},
				},
				"opts": make(map[string]interface{}),
			},
			{
				"cache_dump_fields":    []map[string]interface{}{},
				"concurrent_requests":  1024,
				"type":                 "*file_csv",
				"failed_calls_prefix":  "",
				"field_separator":      ",",
				utils.HeaderDefCharCfg: ":",
				"fields": []map[string]interface{}{
					{"mandatory": true, "path": "*cgreq.ToR", "tag": "ToR", "type": "*variable", "value": "~*req.2"},
					{"mandatory": true, "path": "*cgreq.OriginID", "tag": "OriginID", "type": "*variable", "value": "~*req.3"},
					{"mandatory": true, "path": "*cgreq.RequestType", "tag": "RequestType", "type": "*variable", "value": "~*req.4"},
					{"mandatory": true, "path": "*cgreq.Tenant", "tag": "Tenant", "type": "*variable", "value": "~*req.6"},
					{"mandatory": true, "path": "*cgreq.Category", "tag": "Category", "type": "*variable", "value": "~*req.7"},
					{"mandatory": true, "path": "*cgreq.Account", "tag": "Account", "type": "*variable", "value": "~*req.8"},
					{"mandatory": true, "path": "*cgreq.Subject", "tag": "Subject", "type": "*variable", "value": "~*req.9"},
					{"mandatory": true, "path": "*cgreq.Destination", "tag": "Destination", "type": "*variable", "value": "~*req.10"},
					{"mandatory": true, "path": "*cgreq.SetupTime", "tag": "SetupTime", "type": "*variable", "value": "~*req.11"},
					{"mandatory": true, "path": "*cgreq.AnswerTime", "tag": "AnswerTime", "type": "*variable", "value": "~*req.12"},
					{"mandatory": true, "path": "*cgreq.Usage", "tag": "Usage", "type": "*variable", "value": "~*req.13"},
				},
				"filters":                     filters,
				"flags":                       filters,
				"id":                          "file_reader1",
				"partial_record_cache":        "0",
				"processed_path":              "/tmp/ers/out",
				"row_length":                  0,
				"run_delay":                   "-1",
				"partial_cache_expiry_action": "",
				"source_path":                 "/tmp/ers/in",
				"tenant":                      "",
				"timezone":                    "",
				"xml_root_path":               []string{""},
				"opts":                        make(map[string]interface{}),
			},
		},
	}
	if cfg, err := NewCGRConfigFromJsonStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cfg.ersCfg.AsMapInterface(utils.EmptyString); !reflect.DeepEqual(eMap, rcv) {
		t.Errorf("\nExpected: %+v\nRecived: %+v", utils.ToIJSON(eMap), utils.ToIJSON(rcv))
	}
}
