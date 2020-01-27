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
		ID:           utils.MetaDefault,
		Type:         "RandomType",
		FieldSep:     ",",
		Filters:      []string{"Filter1", "Filter2"},
		Tenant:       NewRSRParsersMustCompile("cgrates.org", true, utils.INFIELD_SEP),
		HeaderFields: []*FCTemplate{},
		ContentFields: []*FCTemplate{
			{
				Tag:       "TOR",
				FieldId:   "ToR",
				Type:      "*composed",
				Value:     NewRSRParsersMustCompile("~*req.2", true, utils.INFIELD_SEP),
				Mandatory: true,
			},
			{
				Tag:       "RandomField",
				FieldId:   "RandomField",
				Type:      "*composed",
				Value:     NewRSRParsersMustCompile("Test", true, utils.INFIELD_SEP),
				Mandatory: true,
			},
		},
		TrailerFields: []*FCTemplate{},
	}
	cloned := orig.Clone()
	if !reflect.DeepEqual(cloned, orig) {
		t.Errorf("expected: %s \n,received: %s", utils.ToJSON(orig), utils.ToJSON(cloned))
	}
	initialOrig := &EventReaderCfg{
		ID:           utils.MetaDefault,
		Type:         "RandomType",
		FieldSep:     ",",
		Filters:      []string{"Filter1", "Filter2"},
		Tenant:       NewRSRParsersMustCompile("cgrates.org", true, utils.INFIELD_SEP),
		HeaderFields: []*FCTemplate{},
		ContentFields: []*FCTemplate{
			{
				Tag:       "TOR",
				FieldId:   "ToR",
				Type:      "*composed",
				Value:     NewRSRParsersMustCompile("~*req.2", true, utils.INFIELD_SEP),
				Mandatory: true,
			},
			{
				Tag:       "RandomField",
				FieldId:   "RandomField",
				Type:      "*composed",
				Value:     NewRSRParsersMustCompile("Test", true, utils.INFIELD_SEP),
				Mandatory: true,
			},
		},
		TrailerFields: []*FCTemplate{},
	}
	orig.Filters = []string{"SingleFilter"}
	orig.ContentFields = []*FCTemplate{
		{
			Tag:       "TOR",
			FieldId:   "ToR",
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
			&EventReaderCfg{
				ID:             utils.MetaDefault,
				Type:           utils.MetaFileCSV,
				FieldSep:       ",",
				RunDelay:       time.ParseDuration("0s"),
				ConcurrentReqs: 1024,
				SourcePath:     "/var/spool/cgrates/cdrc/in",
				ProcessedPath:  "/var/spool/cgrates/cdrc/out",
				XmlRootPath:    utils.HierarchyPath{utils.EmptyString},
				Tenant:         nil,
				Timezone:       utils.EmptyString,
				Filters:        []string{},
				Flags:          utils.FlagsWithParams{},
				HeaderFields:   make([]*FCTemplate, 0),
				ContentFields: []*FCTemplate{
					{Tag: "TOR", FieldId: "ToR", Type: utils.META_COMPOSED,
						Value: NewRSRParsersMustCompile("~*req.2", true, utils.INFIELD_SEP), Mandatory: true},
					{Tag: "OriginID", FieldId: "OriginID", Type: utils.META_COMPOSED,
						Value: NewRSRParsersMustCompile("~*req.3", true, utils.INFIELD_SEP), Mandatory: true},
					{Tag: "RequestType", FieldId: "RequestType", Type: utils.META_COMPOSED,
						Value: NewRSRParsersMustCompile("~*req.4", true, utils.INFIELD_SEP), Mandatory: true},
					{Tag: "Tenant", FieldId: "Tenant", Type: utils.META_COMPOSED,
						Value: NewRSRParsersMustCompile("~*req.6", true, utils.INFIELD_SEP), Mandatory: true},
					{Tag: "Category", FieldId: "Category", Type: utils.META_COMPOSED,
						Value: NewRSRParsersMustCompile("~*req.7", true, utils.INFIELD_SEP), Mandatory: true},
					{Tag: "Account", FieldId: "Account", Type: utils.META_COMPOSED,
						Value: NewRSRParsersMustCompile("~*req.8", true, utils.INFIELD_SEP), Mandatory: true},
					{Tag: "Subject", FieldId: "Subject", Type: utils.META_COMPOSED,
						Value: NewRSRParsersMustCompile("~*req.9", true, utils.INFIELD_SEP), Mandatory: true},
					{Tag: "Destination", FieldId: "Destination", Type: utils.META_COMPOSED,
						Value: NewRSRParsersMustCompile("~*req.10", true, utils.INFIELD_SEP), Mandatory: true},
					{Tag: "SetupTime", FieldId: "SetupTime", Type: utils.META_COMPOSED,
						Value: NewRSRParsersMustCompile("~*req.11", true, utils.INFIELD_SEP), Mandatory: true},
					{Tag: "AnswerTime", FieldId: "AnswerTime", Type: utils.META_COMPOSED,
						Value: NewRSRParsersMustCompile("~*req.12", true, utils.INFIELD_SEP), Mandatory: true},
					{Tag: "Usage", FieldId: "Usage", Type: utils.META_COMPOSED,
						Value: NewRSRParsersMustCompile("~*req.13", true, utils.INFIELD_SEP), Mandatory: true},
				},
				TrailerFields: make([]*FCTemplate, 0),
			},
			&EventReaderCfg{
				ID:             "file_reader1",
				Type:           utils.MetaFileCSV,
				FieldSep:       ",",
				RunDelay:       time.ParseDuration("-1s"),
				ConcurrentReqs: 1024,
				SourcePath:     "/tmp/ers/in",
				ProcessedPath:  "/tmp/ers/out",
				XmlRootPath:    utils.HierarchyPath{utils.EmptyString},
				Tenant:         nil,
				Timezone:       utils.EmptyString,
				Filters:        nil,
				Flags:          utils.FlagsWithParams{},
				HeaderFields:   make([]*FCTemplate, 0),
				ContentFields: []*FCTemplate{
					{Tag: "TOR", FieldId: "ToR", Type: utils.META_COMPOSED,
						Value: NewRSRParsersMustCompile("~*req.2", true, utils.INFIELD_SEP), Mandatory: true},
					{Tag: "OriginID", FieldId: "OriginID", Type: utils.META_COMPOSED,
						Value: NewRSRParsersMustCompile("~*req.3", true, utils.INFIELD_SEP), Mandatory: true},
					{Tag: "RequestType", FieldId: "RequestType", Type: utils.META_COMPOSED,
						Value: NewRSRParsersMustCompile("~*req.4", true, utils.INFIELD_SEP), Mandatory: true},
					{Tag: "Tenant", FieldId: "Tenant", Type: utils.META_COMPOSED,
						Value: NewRSRParsersMustCompile("~*req.6", true, utils.INFIELD_SEP), Mandatory: true},
					{Tag: "Category", FieldId: "Category", Type: utils.META_COMPOSED,
						Value: NewRSRParsersMustCompile("~*req.7", true, utils.INFIELD_SEP), Mandatory: true},
					{Tag: "Account", FieldId: "Account", Type: utils.META_COMPOSED,
						Value: NewRSRParsersMustCompile("~*req.8", true, utils.INFIELD_SEP), Mandatory: true},
					{Tag: "Subject", FieldId: "Subject", Type: utils.META_COMPOSED,
						Value: NewRSRParsersMustCompile("~*req.9", true, utils.INFIELD_SEP), Mandatory: true},
					{Tag: "Destination", FieldId: "Destination", Type: utils.META_COMPOSED,
						Value: NewRSRParsersMustCompile("~*req.10", true, utils.INFIELD_SEP), Mandatory: true},
					{Tag: "SetupTime", FieldId: "SetupTime", Type: utils.META_COMPOSED,
						Value: NewRSRParsersMustCompile("~*req.11", true, utils.INFIELD_SEP), Mandatory: true},
					{Tag: "AnswerTime", FieldId: "AnswerTime", Type: utils.META_COMPOSED,
						Value: NewRSRParsersMustCompile("~*req.12", true, utils.INFIELD_SEP), Mandatory: true},
					{Tag: "Usage", FieldId: "Usage", Type: utils.META_COMPOSED,
						Value: NewRSRParsersMustCompile("~*req.13", true, utils.INFIELD_SEP), Mandatory: true},
				},
				TrailerFields: make([]*FCTemplate, 0),
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
			"run_delay": -1,
			"type": "*file_csv",
			"source_path": "/tmp/ers/in",
			"processed_path": "/tmp/ers/out",
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

func TestEventReaderSanitisation(t *testing.T) {
	cfgJSONStr := `{
"ers": {
	"enabled": true,
	"readers": [
		{
			"id": "file_reader1",
			"run_delay": -1,
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
