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
		ID:            utils.MetaDefault,
		Type:          "RandomType",
		FieldSep:      ",",
		SourceID:      "RandomSource",
		Filters:       []string{"Filter1", "Filter2"},
		Tenant:        NewRSRParsersMustCompile("cgrates.org", true, utils.INFIELD_SEP),
		Header_fields: []*FCTemplate{},
		Content_fields: []*FCTemplate{
			{
				Tag:       "TOR",
				FieldId:   "ToR",
				Type:      "*composed",
				Value:     NewRSRParsersMustCompile("~2", true, utils.INFIELD_SEP),
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
		Trailer_fields: []*FCTemplate{},
		Continue:       true,
	}
	cloned := orig.Clone()
	if !reflect.DeepEqual(cloned, orig) {
		t.Errorf("expected: %s \n,received: %s", utils.ToJSON(orig), utils.ToJSON(cloned))
	}
	initialOrig := &EventReaderCfg{
		ID:            utils.MetaDefault,
		Type:          "RandomType",
		FieldSep:      ",",
		SourceID:      "RandomSource",
		Filters:       []string{"Filter1", "Filter2"},
		Tenant:        NewRSRParsersMustCompile("cgrates.org", true, utils.INFIELD_SEP),
		Header_fields: []*FCTemplate{},
		Content_fields: []*FCTemplate{
			{
				Tag:       "TOR",
				FieldId:   "ToR",
				Type:      "*composed",
				Value:     NewRSRParsersMustCompile("~2", true, utils.INFIELD_SEP),
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
		Trailer_fields: []*FCTemplate{},
		Continue:       true,
	}
	orig.Filters = []string{"SingleFilter"}
	orig.Content_fields = []*FCTemplate{
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
		Enabled: true,
		SessionSConns: []*RemoteHost{
			{
				Address: utils.MetaInternal,
			},
		},
		Readers: []*EventReaderCfg{
			&EventReaderCfg{
				ID:             "file_reader1",
				Type:           utils.MetaFileCSV,
				FieldSep:       ",",
				RunDelay:       time.Duration(-1),
				ConcurrentReqs: 1024,
				SourcePath:     "/tmp/ers/in",
				ProcessedPath:  "/tmp/ers/out",
				XmlRootPath:    utils.EmptyString,
				SourceID:       "ers_csv",
				Tenant:         nil,
				Timezone:       utils.EmptyString,
				Filters:        nil,
				Flags:          utils.FlagsWithParams{},
				Header_fields:  make([]*FCTemplate, 0),
				Content_fields: []*FCTemplate{
					{Tag: "TOR", FieldId: "ToR", Type: utils.META_COMPOSED,
						Value: NewRSRParsersMustCompile("~2", true, utils.INFIELD_SEP), Mandatory: true},
					{Tag: "OriginID", FieldId: "OriginID", Type: utils.META_COMPOSED,
						Value: NewRSRParsersMustCompile("~3", true, utils.INFIELD_SEP), Mandatory: true},
					{Tag: "RequestType", FieldId: "RequestType", Type: utils.META_COMPOSED,
						Value: NewRSRParsersMustCompile("~4", true, utils.INFIELD_SEP), Mandatory: true},
					{Tag: "Tenant", FieldId: "Tenant", Type: utils.META_COMPOSED,
						Value: NewRSRParsersMustCompile("~6", true, utils.INFIELD_SEP), Mandatory: true},
					{Tag: "Category", FieldId: "Category", Type: utils.META_COMPOSED,
						Value: NewRSRParsersMustCompile("~7", true, utils.INFIELD_SEP), Mandatory: true},
					{Tag: "Account", FieldId: "Account", Type: utils.META_COMPOSED,
						Value: NewRSRParsersMustCompile("~8", true, utils.INFIELD_SEP), Mandatory: true},
					{Tag: "Subject", FieldId: "Subject", Type: utils.META_COMPOSED,
						Value: NewRSRParsersMustCompile("~9", true, utils.INFIELD_SEP), Mandatory: true},
					{Tag: "Destination", FieldId: "Destination", Type: utils.META_COMPOSED,
						Value: NewRSRParsersMustCompile("~10", true, utils.INFIELD_SEP), Mandatory: true},
					{Tag: "SetupTime", FieldId: "SetupTime", Type: utils.META_COMPOSED,
						Value: NewRSRParsersMustCompile("~11", true, utils.INFIELD_SEP), Mandatory: true},
					{Tag: "AnswerTime", FieldId: "AnswerTime", Type: utils.META_COMPOSED,
						Value: NewRSRParsersMustCompile("~12", true, utils.INFIELD_SEP), Mandatory: true},
					{Tag: "Usage", FieldId: "Usage", Type: utils.META_COMPOSED,
						Value: NewRSRParsersMustCompile("~13", true, utils.INFIELD_SEP), Mandatory: true},
				},
				Trailer_fields: make([]*FCTemplate, 0),
			},
		},
	}

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

	if cfg, err := NewCGRConfigFromJsonStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedERsCfg, cfg.ersCfg) {
		t.Errorf("Expected: %+v ,\n recived: %+v", utils.ToJSON(expectedERsCfg), utils.ToJSON(cfg.ersCfg))
	}

}
