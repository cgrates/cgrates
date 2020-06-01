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

func TestEventExporterClone(t *testing.T) {
	orig := &EventExporterCfg{
		ID:       utils.MetaDefault,
		Type:     "RandomType",
		FieldSep: ",",
		Filters:  []string{"Filter1", "Filter2"},
		Tenant:   NewRSRParsersMustCompile("cgrates.org", true, utils.INFIELD_SEP),
		contentFields: []*FCTemplate{
			{
				Tag:       "ToR",
				Path:      "*exp.ToR",
				Type:      "*composed",
				Value:     NewRSRParsersMustCompile("~*req.2", true, utils.INFIELD_SEP),
				Mandatory: true,
			},
			{
				Tag:       "RandomField",
				Path:      "*exp.RandomField",
				Type:      "*composed",
				Value:     NewRSRParsersMustCompile("Test", true, utils.INFIELD_SEP),
				Mandatory: true,
			},
		},
		Fields: []*FCTemplate{
			{
				Tag:       "ToR",
				Path:      "*exp.ToR",
				Type:      "*composed",
				Value:     NewRSRParsersMustCompile("~*req.2", true, utils.INFIELD_SEP),
				Mandatory: true,
			},
			{
				Tag:       "RandomField",
				Path:      "*exp.RandomField",
				Type:      "*composed",
				Value:     NewRSRParsersMustCompile("Test", true, utils.INFIELD_SEP),
				Mandatory: true,
			},
		},
		headerFields:  []*FCTemplate{},
		trailerFields: []*FCTemplate{},
	}
	for _, v := range orig.Fields {
		v.ComputePath()
	}
	for _, v := range orig.contentFields {
		v.ComputePath()
	}
	cloned := orig.Clone()
	if !reflect.DeepEqual(cloned, orig) {
		t.Errorf("expected: %s \n,received: %s", utils.ToJSON(orig), utils.ToJSON(cloned))
	}
	initialOrig := &EventExporterCfg{
		ID:       utils.MetaDefault,
		Type:     "RandomType",
		FieldSep: ",",
		Filters:  []string{"Filter1", "Filter2"},
		Tenant:   NewRSRParsersMustCompile("cgrates.org", true, utils.INFIELD_SEP),
		Fields: []*FCTemplate{
			{
				Tag:       "ToR",
				Path:      "*exp.ToR",
				Type:      "*composed",
				Value:     NewRSRParsersMustCompile("~*req.2", true, utils.INFIELD_SEP),
				Mandatory: true,
			},
			{
				Tag:       "RandomField",
				Path:      "*exp.RandomField",
				Type:      "*composed",
				Value:     NewRSRParsersMustCompile("Test", true, utils.INFIELD_SEP),
				Mandatory: true,
			},
		},
		contentFields: []*FCTemplate{
			{
				Tag:       "ToR",
				Path:      "*exp.ToR",
				Type:      "*composed",
				Value:     NewRSRParsersMustCompile("~*req.2", true, utils.INFIELD_SEP),
				Mandatory: true,
			},
			{
				Tag:       "RandomField",
				Path:      "*exp.RandomField",
				Type:      "*composed",
				Value:     NewRSRParsersMustCompile("Test", true, utils.INFIELD_SEP),
				Mandatory: true,
			},
		},
		headerFields:  []*FCTemplate{},
		trailerFields: []*FCTemplate{},
	}
	for _, v := range initialOrig.Fields {
		v.ComputePath()
	}
	for _, v := range initialOrig.contentFields {
		v.ComputePath()
	}
	orig.Filters = []string{"SingleFilter"}
	orig.contentFields = []*FCTemplate{
		{
			Tag:       "ToR",
			Path:      "*exp.ToR",
			Type:      "*composed",
			Value:     NewRSRParsersMustCompile("~2", true, utils.INFIELD_SEP),
			Mandatory: true,
		},
	}
	if !reflect.DeepEqual(cloned, initialOrig) {
		t.Errorf("expected: %s \n,received: %s", utils.ToJSON(initialOrig), utils.ToJSON(cloned))
	}
}

func TestEventExporterSameID(t *testing.T) {
	expectedEEsCfg := &EEsCfg{
		Enabled:         true,
		AttributeSConns: []string{"conn1"},
		Cache: map[string]*CacheParamCfg{
			utils.MetaFileCSV: &CacheParamCfg{
				Limit:     -1,
				TTL:       time.Duration(5 * time.Second),
				StaticTTL: false,
			},
		},
		Exporters: []*EventExporterCfg{
			&EventExporterCfg{
				ID:            utils.MetaDefault,
				Type:          utils.META_NONE,
				FieldSep:      ",",
				Tenant:        nil,
				ExportPath:    "/var/spool/cgrates/ees",
				Attempts:      1,
				Timezone:      utils.EmptyString,
				Filters:       []string{},
				AttributeSIDs: []string{},
				Flags:         utils.FlagsWithParams{},
				Fields: []*FCTemplate{
					{
						Tag:    utils.CGRID,
						Path:   "*exp.CGRID",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.CGRID", true, utils.INFIELD_SEP),
						Layout: time.RFC3339,
					},
					{
						Tag:    utils.RunID,
						Path:   "*exp.RunID",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.RunID", true, utils.INFIELD_SEP),
						Layout: time.RFC3339,
					},
					{
						Tag:    utils.ToR,
						Path:   "*exp.ToR",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.ToR", true, utils.INFIELD_SEP),
						Layout: time.RFC3339,
					},
					{
						Tag:    utils.OriginID,
						Path:   "*exp.OriginID",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.OriginID", true, utils.INFIELD_SEP),
						Layout: time.RFC3339,
					},
					{
						Tag:    utils.RequestType,
						Path:   "*exp.RequestType",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.RequestType", true, utils.INFIELD_SEP),
						Layout: time.RFC3339,
					},
					{
						Tag:    utils.Tenant,
						Path:   "*exp.Tenant",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.Tenant", true, utils.INFIELD_SEP),
						Layout: time.RFC3339,
					},
					{
						Tag:    utils.Category,
						Path:   "*exp.Category",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.Category", true, utils.INFIELD_SEP),
						Layout: time.RFC3339,
					},
					{
						Tag:    utils.Account,
						Path:   "*exp.Account",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.Account", true, utils.INFIELD_SEP),
						Layout: time.RFC3339,
					},
					{
						Tag:    utils.Subject,
						Path:   "*exp.Subject",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.Subject", true, utils.INFIELD_SEP),
						Layout: time.RFC3339,
					},
					{
						Tag:    utils.Destination,
						Path:   "*exp.Destination",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.Destination", true, utils.INFIELD_SEP),
						Layout: time.RFC3339,
					},
					{
						Tag:    utils.SetupTime,
						Path:   "*exp.SetupTime",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.SetupTime", true, utils.INFIELD_SEP),
						Layout: "2006-01-02T15:04:05Z07:00",
					},
					{
						Tag:    utils.AnswerTime,
						Path:   "*exp.AnswerTime",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.AnswerTime", true, utils.INFIELD_SEP),
						Layout: "2006-01-02T15:04:05Z07:00",
					},
					{
						Tag:    utils.Usage,
						Path:   "*exp.Usage",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.Usage", true, utils.INFIELD_SEP),
						Layout: time.RFC3339,
					},
					{
						Tag:    utils.Cost,
						Path:   "*exp.Cost",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.Cost{*round:4}", true, utils.INFIELD_SEP),
						Layout: time.RFC3339,
					},
				},
				contentFields: []*FCTemplate{
					{
						Tag:    utils.CGRID,
						Path:   "*exp.CGRID",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.CGRID", true, utils.INFIELD_SEP),
						Layout: time.RFC3339,
					},
					{
						Tag:    utils.RunID,
						Path:   "*exp.RunID",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.RunID", true, utils.INFIELD_SEP),
						Layout: time.RFC3339,
					},
					{
						Tag:    utils.ToR,
						Path:   "*exp.ToR",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.ToR", true, utils.INFIELD_SEP),
						Layout: time.RFC3339,
					},
					{
						Tag:    utils.OriginID,
						Path:   "*exp.OriginID",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.OriginID", true, utils.INFIELD_SEP),
						Layout: time.RFC3339,
					},
					{
						Tag:    utils.RequestType,
						Path:   "*exp.RequestType",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.RequestType", true, utils.INFIELD_SEP),
						Layout: time.RFC3339,
					},
					{
						Tag:    utils.Tenant,
						Path:   "*exp.Tenant",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.Tenant", true, utils.INFIELD_SEP),
						Layout: time.RFC3339,
					},
					{
						Tag:    utils.Category,
						Path:   "*exp.Category",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.Category", true, utils.INFIELD_SEP),
						Layout: time.RFC3339,
					},
					{
						Tag:    utils.Account,
						Path:   "*exp.Account",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.Account", true, utils.INFIELD_SEP),
						Layout: time.RFC3339,
					},
					{
						Tag:    utils.Subject,
						Path:   "*exp.Subject",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.Subject", true, utils.INFIELD_SEP),
						Layout: time.RFC3339,
					},
					{
						Tag:    utils.Destination,
						Path:   "*exp.Destination",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.Destination", true, utils.INFIELD_SEP),
						Layout: time.RFC3339,
					},
					{
						Tag:    utils.SetupTime,
						Path:   "*exp.SetupTime",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.SetupTime", true, utils.INFIELD_SEP),
						Layout: "2006-01-02T15:04:05Z07:00",
					},
					{
						Tag:    utils.AnswerTime,
						Path:   "*exp.AnswerTime",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.AnswerTime", true, utils.INFIELD_SEP),
						Layout: "2006-01-02T15:04:05Z07:00",
					},
					{
						Tag:    utils.Usage,
						Path:   "*exp.Usage",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.Usage", true, utils.INFIELD_SEP),
						Layout: time.RFC3339,
					},
					{
						Tag:    utils.Cost,
						Path:   "*exp.Cost",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.Cost{*round:4}", true, utils.INFIELD_SEP),
						Layout: time.RFC3339,
					},
				},
				headerFields:  []*FCTemplate{},
				trailerFields: []*FCTemplate{},
			},
			{
				ID:         "file_exporter1",
				Type:       utils.MetaFileCSV,
				FieldSep:   ",",
				Tenant:     nil,
				Timezone:   utils.EmptyString,
				Filters:    nil,
				ExportPath: "/var/spool/cgrates/ees",
				Attempts:   1,
				Flags:      utils.FlagsWithParams{},
				Fields: []*FCTemplate{
					{Tag: "CustomTag2", Path: "*exp.CustomPath2", Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("CustomValue2", true, utils.INFIELD_SEP), Mandatory: true, Layout: time.RFC3339},
				},
				contentFields: []*FCTemplate{
					{Tag: "CustomTag2", Path: "*exp.CustomPath2", Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("CustomValue2", true, utils.INFIELD_SEP), Mandatory: true, Layout: time.RFC3339},
				},
				headerFields:  []*FCTemplate{},
				trailerFields: []*FCTemplate{},
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
			"type": "*file_csv",
			"fields":[
				{"tag": "CustomTag1", "path": "*exp.CustomPath1", "type": "*variable", "value": "CustomValue1", "mandatory": true},
			],
		},
		{
			"id": "file_exporter1",
			"type": "*file_csv",
			"fields":[
				{"tag": "CustomTag2", "path": "*exp.CustomPath2", "type": "*variable", "value": "CustomValue2", "mandatory": true},
			],
		},
	],
}
}`

	if cfg, err := NewCGRConfigFromJsonStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedEEsCfg, cfg.eesCfg) {
		t.Errorf("Expected: %+v ,\n recived: %+v", utils.ToJSON(expectedEEsCfg), utils.ToJSON(cfg.eesCfg))
	}

}
