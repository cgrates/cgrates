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
		Tenant:   NewRSRParsersMustCompile("cgrates.org", true, utils.INFIELD_SEP),
		Fields: []*FCTemplate{
			{
				Tag:       "ToR",
				Path:      "ToR",
				Type:      "*composed",
				Value:     NewRSRParsersMustCompile("~*req.2", true, utils.INFIELD_SEP),
				Mandatory: true,
			},
			{
				Tag:       "RandomField",
				Path:      "RandomField",
				Type:      "*composed",
				Value:     NewRSRParsersMustCompile("Test", true, utils.INFIELD_SEP),
				Mandatory: true,
			},
		},
		CacheDumpFields: make([]*FCTemplate, 0),
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
		Tenant:   NewRSRParsersMustCompile("cgrates.org", true, utils.INFIELD_SEP),
		Fields: []*FCTemplate{
			{
				Tag:       "ToR",
				Path:      "ToR",
				Type:      "*composed",
				Value:     NewRSRParsersMustCompile("~*req.2", true, utils.INFIELD_SEP),
				Mandatory: true,
			},
			{
				Tag:       "RandomField",
				Path:      "RandomField",
				Type:      "*composed",
				Value:     NewRSRParsersMustCompile("Test", true, utils.INFIELD_SEP),
				Mandatory: true,
			},
		},
		CacheDumpFields: make([]*FCTemplate, 0),
	}
	orig.Filters = []string{"SingleFilter"}
	orig.Fields = []*FCTemplate{
		{
			Tag:       "ToR",
			Path:      "ToR",
			Type:      "*composed",
			Value:     NewRSRParsersMustCompile("~2", true, utils.INFIELD_SEP),
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
				ID:             utils.MetaDefault,
				Type:           utils.META_NONE,
				FieldSep:       ",",
				RunDelay:       time.Duration(0),
				ConcurrentReqs: 1024,
				SourcePath:     "/var/spool/cgrates/cdrc/in",
				ProcessedPath:  "/var/spool/cgrates/cdrc/out",
				XmlRootPath:    utils.HierarchyPath{utils.EmptyString},
				Tenant:         nil,
				Timezone:       utils.EmptyString,
				Filters:        []string{},
				Flags:          utils.FlagsWithParams{},
				Fields: []*FCTemplate{
					{Tag: utils.ToR, Path: utils.MetaCgreq + utils.NestingSep + utils.ToR, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.2", true, utils.INFIELD_SEP), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.OriginID, Path: utils.MetaCgreq + utils.NestingSep + utils.OriginID, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.3", true, utils.INFIELD_SEP), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.RequestType, Path: utils.MetaCgreq + utils.NestingSep + utils.RequestType, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.4", true, utils.INFIELD_SEP), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.Tenant, Path: utils.MetaCgreq + utils.NestingSep + utils.Tenant, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.6", true, utils.INFIELD_SEP), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.Category, Path: utils.MetaCgreq + utils.NestingSep + utils.Category, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.7", true, utils.INFIELD_SEP), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.Account, Path: utils.MetaCgreq + utils.NestingSep + utils.Account, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.8", true, utils.INFIELD_SEP), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.Subject, Path: utils.MetaCgreq + utils.NestingSep + utils.Subject, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.9", true, utils.INFIELD_SEP), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.Destination, Path: utils.MetaCgreq + utils.NestingSep + utils.Destination, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.10", true, utils.INFIELD_SEP), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.SetupTime, Path: utils.MetaCgreq + utils.NestingSep + utils.SetupTime, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.11", true, utils.INFIELD_SEP), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.AnswerTime, Path: utils.MetaCgreq + utils.NestingSep + utils.AnswerTime, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.12", true, utils.INFIELD_SEP), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.Usage, Path: utils.MetaCgreq + utils.NestingSep + utils.Usage, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.13", true, utils.INFIELD_SEP), Mandatory: true, Layout: time.RFC3339},
				},
				CacheDumpFields: make([]*FCTemplate, 0),
			},
			{
				ID:             "file_reader1",
				Type:           utils.MetaFileCSV,
				FieldSep:       ",",
				RunDelay:       time.Duration(-1),
				ConcurrentReqs: 1024,
				SourcePath:     "/tmp/ers/in",
				ProcessedPath:  "/tmp/ers/out",
				XmlRootPath:    utils.HierarchyPath{utils.EmptyString},
				Tenant:         nil,
				Timezone:       utils.EmptyString,
				Filters:        nil,
				Flags:          utils.FlagsWithParams{},
				Fields: []*FCTemplate{
					{Tag: utils.ToR, Path: utils.MetaCgreq + utils.NestingSep + utils.ToR, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.2", true, utils.INFIELD_SEP), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.OriginID, Path: utils.MetaCgreq + utils.NestingSep + utils.OriginID, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.3", true, utils.INFIELD_SEP), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.RequestType, Path: utils.MetaCgreq + utils.NestingSep + utils.RequestType, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.4", true, utils.INFIELD_SEP), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.Tenant, Path: utils.MetaCgreq + utils.NestingSep + utils.Tenant, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.6", true, utils.INFIELD_SEP), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.Category, Path: utils.MetaCgreq + utils.NestingSep + utils.Category, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.7", true, utils.INFIELD_SEP), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.Account, Path: utils.MetaCgreq + utils.NestingSep + utils.Account, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.8", true, utils.INFIELD_SEP), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.Subject, Path: utils.MetaCgreq + utils.NestingSep + utils.Subject, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.9", true, utils.INFIELD_SEP), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.Destination, Path: utils.MetaCgreq + utils.NestingSep + utils.Destination, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.10", true, utils.INFIELD_SEP), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.SetupTime, Path: utils.MetaCgreq + utils.NestingSep + utils.SetupTime, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.11", true, utils.INFIELD_SEP), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.AnswerTime, Path: utils.MetaCgreq + utils.NestingSep + utils.AnswerTime, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.12", true, utils.INFIELD_SEP), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.Usage, Path: utils.MetaCgreq + utils.NestingSep + utils.Usage, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.13", true, utils.INFIELD_SEP), Mandatory: true, Layout: time.RFC3339},
				},
				CacheDumpFields: make([]*FCTemplate, 0),
			},
		},
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
				ID:             utils.MetaDefault,
				Type:           utils.META_NONE,
				FieldSep:       ",",
				RunDelay:       time.Duration(0),
				ConcurrentReqs: 1024,
				SourcePath:     "/var/spool/cgrates/cdrc/in",
				ProcessedPath:  "/var/spool/cgrates/cdrc/out",
				XmlRootPath:    utils.HierarchyPath{utils.EmptyString},
				Tenant:         nil,
				Timezone:       utils.EmptyString,
				Filters:        []string{},
				Flags:          utils.FlagsWithParams{},
				Fields: []*FCTemplate{
					{Tag: utils.ToR, Path: utils.MetaCgreq + utils.NestingSep + utils.ToR, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.2", true, utils.INFIELD_SEP), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.OriginID, Path: utils.MetaCgreq + utils.NestingSep + utils.OriginID, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.3", true, utils.INFIELD_SEP), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.RequestType, Path: utils.MetaCgreq + utils.NestingSep + utils.RequestType, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.4", true, utils.INFIELD_SEP), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.Tenant, Path: utils.MetaCgreq + utils.NestingSep + utils.Tenant, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.6", true, utils.INFIELD_SEP), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.Category, Path: utils.MetaCgreq + utils.NestingSep + utils.Category, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.7", true, utils.INFIELD_SEP), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.Account, Path: utils.MetaCgreq + utils.NestingSep + utils.Account, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.8", true, utils.INFIELD_SEP), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.Subject, Path: utils.MetaCgreq + utils.NestingSep + utils.Subject, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.9", true, utils.INFIELD_SEP), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.Destination, Path: utils.MetaCgreq + utils.NestingSep + utils.Destination, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.10", true, utils.INFIELD_SEP), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.SetupTime, Path: utils.MetaCgreq + utils.NestingSep + utils.SetupTime, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.11", true, utils.INFIELD_SEP), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.AnswerTime, Path: utils.MetaCgreq + utils.NestingSep + utils.AnswerTime, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.12", true, utils.INFIELD_SEP), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.Usage, Path: utils.MetaCgreq + utils.NestingSep + utils.Usage, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.13", true, utils.INFIELD_SEP), Mandatory: true, Layout: time.RFC3339},
				},
				CacheDumpFields: make([]*FCTemplate, 0),
			},
			{
				ID:             "file_reader1",
				Type:           utils.MetaFileCSV,
				RowLength:      5,
				FieldSep:       ",",
				RunDelay:       time.Duration(-1),
				ConcurrentReqs: 1024,
				SourcePath:     "/tmp/ers/in",
				ProcessedPath:  "/tmp/ers/out",
				XmlRootPath:    utils.HierarchyPath{utils.EmptyString},
				Tenant:         nil,
				Timezone:       utils.EmptyString,
				Filters:        nil,
				Flags:          utils.FlagsWithParams{},
				Fields: []*FCTemplate{
					{Tag: "CustomTag2", Path: "CustomPath2", Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("CustomValue2", true, utils.INFIELD_SEP), Mandatory: true, Layout: time.RFC3339},
				},
				CacheDumpFields: make([]*FCTemplate, 0),
			},
		},
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
