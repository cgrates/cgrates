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
		"*file_csv": {"limit": -2, "ttl": "3s", "static_ttl": true},
	},
	"exporters": [
		{
			"id": "cgrates",									
			"type": "*none",									
			"export_path": "/var/spool/cgrates/ees",			
			"opts": {
              "*default": "randomVal"
             },											
			"tenant": "~*req.Destination1",										
			"timezone": "local",										
			"filters": ["randomFiletrs"],										
			"flags": [],										
			"attribute_ids": ["randomID"],								
			"attribute_context": "",							
			"synchronous": false,								
			"attempts": 2,										
			"field_separator": ",",								
			"fields":[											
				{"tag": "CGRID", "path": "*exp.CGRID", "type": "*variable", "value": "~*req.CGRID"},
                {"tag": "CGRID", "path": "*hdr.CGRID", "type": "*variable", "value": "~*req.CGRID"},
                {"tag": "CGRID", "path": "*trl.CGRID", "type": "*variable", "value": "~*req.CGRID"},
                {"tag": "CGRID", "path": "*uch.CGRID", "type": "*variable", "value": "~*req.CGRID"},
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
				TTL:       time.Duration(3 * time.Second),
				StaticTTL: true,
				Precache:  false,
				Replicate: false,
			},
		},
		Exporters: []*EventExporterCfg{
			{
				ID:            utils.MetaDefault,
				Type:          utils.META_NONE,
				FieldSep:      ",",
				Synchronous:   false,
				Tenant:        NewRSRParsersMustCompile("", utils.INFIELD_SEP),
				ExportPath:    "/var/spool/cgrates/ees",
				Attempts:      1,
				Timezone:      utils.EmptyString,
				AttributeSCtx: utils.EmptyString,
				Filters:       nil,
				AttributeSIDs: nil,
				Flags:         utils.FlagsWithParams{},
				Fields: []*FCTemplate{
					{
						Tag:    utils.CGRID,
						Path:   "*exp.CGRID",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.CGRID", utils.INFIELD_SEP),
						Layout: time.RFC3339,
					},
					{
						Tag:    utils.RunID,
						Path:   "*exp.RunID",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.RunID", utils.INFIELD_SEP),
						Layout: time.RFC3339,
					},
					{
						Tag:    utils.ToR,
						Path:   "*exp.ToR",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.ToR", utils.INFIELD_SEP),
						Layout: time.RFC3339,
					},
					{
						Tag:    utils.OriginID,
						Path:   "*exp.OriginID",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.OriginID", utils.INFIELD_SEP),
						Layout: time.RFC3339,
					},
					{
						Tag:    utils.RequestType,
						Path:   "*exp.RequestType",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.RequestType", utils.INFIELD_SEP),
						Layout: time.RFC3339,
					},
					{
						Tag:    utils.Tenant,
						Path:   "*exp.Tenant",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.Tenant", utils.INFIELD_SEP),
						Layout: time.RFC3339,
					},
					{
						Tag:    utils.Category,
						Path:   "*exp.Category",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.Category", utils.INFIELD_SEP),
						Layout: time.RFC3339,
					},
					{
						Tag:    utils.Account,
						Path:   "*exp.Account",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.Account", utils.INFIELD_SEP),
						Layout: time.RFC3339,
					},
					{
						Tag:    utils.Subject,
						Path:   "*exp.Subject",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.Subject", utils.INFIELD_SEP),
						Layout: time.RFC3339,
					},
					{
						Tag:    utils.Destination,
						Path:   "*exp.Destination",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.Destination", utils.INFIELD_SEP),
						Layout: time.RFC3339,
					},
					{
						Tag:    utils.SetupTime,
						Path:   "*exp.SetupTime",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.SetupTime", utils.INFIELD_SEP),
						Layout: "2006-01-02T15:04:05Z07:00",
					},
					{
						Tag:    utils.AnswerTime,
						Path:   "*exp.AnswerTime",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.AnswerTime", utils.INFIELD_SEP),
						Layout: "2006-01-02T15:04:05Z07:00",
					},
					{
						Tag:    utils.Usage,
						Path:   "*exp.Usage",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.Usage", utils.INFIELD_SEP),
						Layout: time.RFC3339,
					},
					{
						Tag:    utils.Cost,
						Path:   "*exp.Cost",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.Cost{*round:4}", utils.INFIELD_SEP),
						Layout: time.RFC3339,
					},
				},
				contentFields: []*FCTemplate{
					{
						Tag:    utils.CGRID,
						Path:   "*exp.CGRID",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.CGRID", utils.INFIELD_SEP),
						Layout: time.RFC3339,
					},
					{
						Tag:    utils.RunID,
						Path:   "*exp.RunID",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.RunID", utils.INFIELD_SEP),
						Layout: time.RFC3339,
					},
					{
						Tag:    utils.ToR,
						Path:   "*exp.ToR",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.ToR", utils.INFIELD_SEP),
						Layout: time.RFC3339,
					},
					{
						Tag:    utils.OriginID,
						Path:   "*exp.OriginID",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.OriginID", utils.INFIELD_SEP),
						Layout: time.RFC3339,
					},
					{
						Tag:    utils.RequestType,
						Path:   "*exp.RequestType",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.RequestType", utils.INFIELD_SEP),
						Layout: time.RFC3339,
					},
					{
						Tag:    utils.Tenant,
						Path:   "*exp.Tenant",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.Tenant", utils.INFIELD_SEP),
						Layout: time.RFC3339,
					},
					{
						Tag:    utils.Category,
						Path:   "*exp.Category",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.Category", utils.INFIELD_SEP),
						Layout: time.RFC3339,
					},
					{
						Tag:    utils.Account,
						Path:   "*exp.Account",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.Account", utils.INFIELD_SEP),
						Layout: time.RFC3339,
					},
					{
						Tag:    utils.Subject,
						Path:   "*exp.Subject",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.Subject", utils.INFIELD_SEP),
						Layout: time.RFC3339,
					},
					{
						Tag:    utils.Destination,
						Path:   "*exp.Destination",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.Destination", utils.INFIELD_SEP),
						Layout: time.RFC3339,
					},
					{
						Tag:    utils.SetupTime,
						Path:   "*exp.SetupTime",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.SetupTime", utils.INFIELD_SEP),
						Layout: "2006-01-02T15:04:05Z07:00",
					},
					{
						Tag:    utils.AnswerTime,
						Path:   "*exp.AnswerTime",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.AnswerTime", utils.INFIELD_SEP),
						Layout: "2006-01-02T15:04:05Z07:00",
					},
					{
						Tag:    utils.Usage,
						Path:   "*exp.Usage",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.Usage", utils.INFIELD_SEP),
						Layout: time.RFC3339,
					},
					{
						Tag:    utils.Cost,
						Path:   "*exp.Cost",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.Cost{*round:4}", utils.INFIELD_SEP),
						Layout: time.RFC3339,
					},
				},
				headerFields:  []*FCTemplate{},
				trailerFields: []*FCTemplate{},
				Opts:          make(map[string]interface{}),
			},
			{
				ID:            utils.CGRATES,
				Type:          utils.META_NONE,
				FieldSep:      ",",
				Synchronous:   false,
				Tenant:        NewRSRParsersMustCompile("~*req.Destination1", utils.INFIELD_SEP),
				ExportPath:    "/var/spool/cgrates/ees",
				Attempts:      2,
				Timezone:      "local",
				Filters:       []string{"randomFiletrs"},
				AttributeSIDs: []string{"randomID"},
				Flags:         utils.FlagsWithParams{},
				Fields: []*FCTemplate{
					{
						Tag:    utils.CGRID,
						Path:   "*exp.CGRID",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.CGRID", utils.INFIELD_SEP),
						Layout: time.RFC3339,
					},
					{
						Tag:    utils.CGRID,
						Path:   "*hdr.CGRID",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.CGRID", utils.INFIELD_SEP),
						Layout: time.RFC3339,
					},
					{
						Tag:    utils.CGRID,
						Path:   "*trl.CGRID",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.CGRID", utils.INFIELD_SEP),
						Layout: time.RFC3339,
					},
					{
						Tag:    utils.CGRID,
						Path:   "*uch.CGRID",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.CGRID", utils.INFIELD_SEP),
						Layout: time.RFC3339,
					},
				},
				contentFields: []*FCTemplate{
					{
						Tag:    utils.CGRID,
						Path:   "*exp.CGRID",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.CGRID", utils.INFIELD_SEP),
						Layout: time.RFC3339,
					},
					{
						Tag:    utils.CGRID,
						Path:   "*uch.CGRID",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.CGRID", utils.INFIELD_SEP),
						Layout: time.RFC3339,
					},
				},
				headerFields: []*FCTemplate{
					{
						Tag:    utils.CGRID,
						Path:   "*hdr.CGRID",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.CGRID", utils.INFIELD_SEP),
						Layout: time.RFC3339,
					},
				},
				trailerFields: []*FCTemplate{
					{
						Tag:    utils.CGRID,
						Path:   "*trl.CGRID",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.CGRID", utils.INFIELD_SEP),
						Layout: time.RFC3339,
					},
				},
				Opts: map[string]interface{}{
					utils.MetaDefault: "randomVal",
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
	if jsonCfg, err := NewCGRConfigFromJsonStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else {
		cloneCfg := jsonCfg.eesCfg.Clone()
		if !reflect.DeepEqual(cloneCfg, expected) {
			t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(cloneCfg))
		}
	}
}

func TestEventExporterFieldloadFromJsonCfg(t *testing.T) {
	eventExporterJson := &EEsJsonCfg{
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
	if jsonCfg, err := NewDefaultCGRConfig(); err != nil {
		t.Error(err)
	} else if err = jsonCfg.eesCfg.loadFromJsonCfg(eventExporterJson, jsonCfg.templates, jsonCfg.generalCfg.RSRSep, jsonCfg.dfltEvExp); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestEventExporterFieldloadFromJsonCfg1(t *testing.T) {
	eventExporterJson := &EEsJsonCfg{
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
	if jsonCfg, err := NewDefaultCGRConfig(); err != nil {
		t.Error(err)
	} else if err = jsonCfg.eesCfg.loadFromJsonCfg(eventExporterJson, jsonCfg.templates, jsonCfg.generalCfg.RSRSep, jsonCfg.dfltEvExp); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestEventExporterloadFromJsonCfg(t *testing.T) {
	jsonCfg, err := NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
	}
	eventExporter := new(EventExporterCfg)
	if err := eventExporter.loadFromJsonCfg(nil, jsonCfg.templates, jsonCfg.generalCfg.RSRSep); err != nil {
		t.Error(err)
	}
}

func TestEESCacheloadFromJsonCfg(t *testing.T) {
	eesCfg := &EEsJsonCfg{
		Cache: &map[string]*CacheParamJsonCfg{
			utils.MetaFileCSV: {
				Ttl: utils.StringPointer("1ss"),
			},
		},
	}
	expected := "time: unknown unit \"ss\" in duration \"1ss\""
	if jsonCfg, err := NewDefaultCGRConfig(); err != nil {
		t.Error(err)
	} else if err = jsonCfg.eesCfg.loadFromJsonCfg(eesCfg, jsonCfg.templates, jsonCfg.generalCfg.RSRSep, jsonCfg.dfltEvExp); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestEESExportersloadFromJsonCfg(t *testing.T) {
	eesCfg := &EEsJsonCfg{
		Exporters: &[]*EventExporterJsonCfg{
			{
				Tenant: utils.StringPointer("a{*"),
			},
		},
	}
	expected := "invalid converter terminator in rule: <a{*>"
	if jsonCfg, err := NewDefaultCGRConfig(); err != nil {
		t.Error(err)
	} else if err = jsonCfg.eesCfg.loadFromJsonCfg(eesCfg, jsonCfg.templates, jsonCfg.generalCfg.RSRSep, jsonCfg.dfltEvExp); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
	eesCfgExporter := &EEsJsonCfg{
		Exporters: nil,
	}
	if jsonCfg, err := NewDefaultCGRConfig(); err != nil {
		t.Error(err)
	} else if err = jsonCfg.eesCfg.loadFromJsonCfg(eesCfgExporter, jsonCfg.templates, jsonCfg.generalCfg.RSRSep, jsonCfg.dfltEvExp); err != nil {
		t.Error(err)
	}
}

func TestEventExporterSameID(t *testing.T) {
	expectedEEsCfg := &EEsCfg{
		Enabled:         true,
		AttributeSConns: []string{"conn1"},
		Cache: map[string]*CacheParamCfg{
			utils.MetaFileCSV: {
				Limit:     -1,
				TTL:       time.Duration(5 * time.Second),
				StaticTTL: false,
			},
		},
		Exporters: []*EventExporterCfg{
			{
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
						Value:  NewRSRParsersMustCompile("~*req.CGRID", utils.INFIELD_SEP),
						Layout: time.RFC3339,
					},
					{
						Tag:    utils.RunID,
						Path:   "*exp.RunID",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.RunID", utils.INFIELD_SEP),
						Layout: time.RFC3339,
					},
					{
						Tag:    utils.ToR,
						Path:   "*exp.ToR",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.ToR", utils.INFIELD_SEP),
						Layout: time.RFC3339,
					},
					{
						Tag:    utils.OriginID,
						Path:   "*exp.OriginID",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.OriginID", utils.INFIELD_SEP),
						Layout: time.RFC3339,
					},
					{
						Tag:    utils.RequestType,
						Path:   "*exp.RequestType",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.RequestType", utils.INFIELD_SEP),
						Layout: time.RFC3339,
					},
					{
						Tag:    utils.Tenant,
						Path:   "*exp.Tenant",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.Tenant", utils.INFIELD_SEP),
						Layout: time.RFC3339,
					},
					{
						Tag:    utils.Category,
						Path:   "*exp.Category",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.Category", utils.INFIELD_SEP),
						Layout: time.RFC3339,
					},
					{
						Tag:    utils.Account,
						Path:   "*exp.Account",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.Account", utils.INFIELD_SEP),
						Layout: time.RFC3339,
					},
					{
						Tag:    utils.Subject,
						Path:   "*exp.Subject",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.Subject", utils.INFIELD_SEP),
						Layout: time.RFC3339,
					},
					{
						Tag:    utils.Destination,
						Path:   "*exp.Destination",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.Destination", utils.INFIELD_SEP),
						Layout: time.RFC3339,
					},
					{
						Tag:    utils.SetupTime,
						Path:   "*exp.SetupTime",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.SetupTime", utils.INFIELD_SEP),
						Layout: "2006-01-02T15:04:05Z07:00",
					},
					{
						Tag:    utils.AnswerTime,
						Path:   "*exp.AnswerTime",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.AnswerTime", utils.INFIELD_SEP),
						Layout: "2006-01-02T15:04:05Z07:00",
					},
					{
						Tag:    utils.Usage,
						Path:   "*exp.Usage",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.Usage", utils.INFIELD_SEP),
						Layout: time.RFC3339,
					},
					{
						Tag:    utils.Cost,
						Path:   "*exp.Cost",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.Cost{*round:4}", utils.INFIELD_SEP),
						Layout: time.RFC3339,
					},
				},
				contentFields: []*FCTemplate{
					{
						Tag:    utils.CGRID,
						Path:   "*exp.CGRID",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.CGRID", utils.INFIELD_SEP),
						Layout: time.RFC3339,
					},
					{
						Tag:    utils.RunID,
						Path:   "*exp.RunID",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.RunID", utils.INFIELD_SEP),
						Layout: time.RFC3339,
					},
					{
						Tag:    utils.ToR,
						Path:   "*exp.ToR",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.ToR", utils.INFIELD_SEP),
						Layout: time.RFC3339,
					},
					{
						Tag:    utils.OriginID,
						Path:   "*exp.OriginID",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.OriginID", utils.INFIELD_SEP),
						Layout: time.RFC3339,
					},
					{
						Tag:    utils.RequestType,
						Path:   "*exp.RequestType",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.RequestType", utils.INFIELD_SEP),
						Layout: time.RFC3339,
					},
					{
						Tag:    utils.Tenant,
						Path:   "*exp.Tenant",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.Tenant", utils.INFIELD_SEP),
						Layout: time.RFC3339,
					},
					{
						Tag:    utils.Category,
						Path:   "*exp.Category",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.Category", utils.INFIELD_SEP),
						Layout: time.RFC3339,
					},
					{
						Tag:    utils.Account,
						Path:   "*exp.Account",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.Account", utils.INFIELD_SEP),
						Layout: time.RFC3339,
					},
					{
						Tag:    utils.Subject,
						Path:   "*exp.Subject",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.Subject", utils.INFIELD_SEP),
						Layout: time.RFC3339,
					},
					{
						Tag:    utils.Destination,
						Path:   "*exp.Destination",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.Destination", utils.INFIELD_SEP),
						Layout: time.RFC3339,
					},
					{
						Tag:    utils.SetupTime,
						Path:   "*exp.SetupTime",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.SetupTime", utils.INFIELD_SEP),
						Layout: "2006-01-02T15:04:05Z07:00",
					},
					{
						Tag:    utils.AnswerTime,
						Path:   "*exp.AnswerTime",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.AnswerTime", utils.INFIELD_SEP),
						Layout: "2006-01-02T15:04:05Z07:00",
					},
					{
						Tag:    utils.Usage,
						Path:   "*exp.Usage",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.Usage", utils.INFIELD_SEP),
						Layout: time.RFC3339,
					},
					{
						Tag:    utils.Cost,
						Path:   "*exp.Cost",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.Cost{*round:4}", utils.INFIELD_SEP),
						Layout: time.RFC3339,
					},
				},
				headerFields:  []*FCTemplate{},
				trailerFields: []*FCTemplate{},
				Opts:          make(map[string]interface{}),
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
						Value: NewRSRParsersMustCompile("CustomValue2", utils.INFIELD_SEP), Mandatory: true, Layout: time.RFC3339},
				},
				contentFields: []*FCTemplate{
					{Tag: "CustomTag2", Path: "*exp.CustomPath2", Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("CustomValue2", utils.INFIELD_SEP), Mandatory: true, Layout: time.RFC3339},
				},
				headerFields:  []*FCTemplate{},
				trailerFields: []*FCTemplate{},
				Opts:          make(map[string]interface{}),
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

func TestEEsCfgloadFromJsonCfgCase1(t *testing.T) {
	jsonCfg := &EEsJsonCfg{
		Enabled:          utils.BoolPointer(true),
		Attributes_conns: &[]string{"*conn1", "*conn2"},
		Cache: &map[string]*CacheParamJsonCfg{
			utils.MetaFileCSV: {
				Limit:      utils.IntPointer(-2),
				Ttl:        utils.StringPointer("1s"),
				Static_ttl: utils.BoolPointer(false),
			},
		},
		Exporters: &[]*EventExporterJsonCfg{
			{
				Id:              utils.StringPointer("CSVExporter"),
				Type:            utils.StringPointer("*file_csv"),
				Filters:         &[]string{},
				Attribute_ids:   &[]string{},
				Flags:           &[]string{"*dryrun"},
				Export_path:     utils.StringPointer("/tmp/testCSV"),
				Tenant:          nil,
				Timezone:        utils.StringPointer("UTC"),
				Synchronous:     utils.BoolPointer(true),
				Attempts:        utils.IntPointer(1),
				Field_separator: utils.StringPointer(","),
				Fields: &[]*FcTemplateJsonCfg{
					{
						Tag:   utils.StringPointer(utils.CGRID),
						Path:  utils.StringPointer("*exp.CGRID"),
						Type:  utils.StringPointer(utils.MetaVariable),
						Value: utils.StringPointer("~*req.CGRID"),
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
				TTL:       1 * time.Second,
				StaticTTL: false,
			},
		},
		Exporters: []*EventExporterCfg{
			{
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
				contentFields: []*FCTemplate{
					{
						Tag:    utils.CGRID,
						Path:   "*exp.CGRID",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.CGRID", utils.INFIELD_SEP),
						Layout: time.RFC3339,
					},
					{
						Tag:    utils.RunID,
						Path:   "*exp.RunID",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.RunID", utils.INFIELD_SEP),
						Layout: time.RFC3339,
					},
					{
						Tag:    utils.ToR,
						Path:   "*exp.ToR",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.ToR", utils.INFIELD_SEP),
						Layout: time.RFC3339,
					},
					{
						Tag:    utils.OriginID,
						Path:   "*exp.OriginID",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.OriginID", utils.INFIELD_SEP),
						Layout: time.RFC3339,
					},
					{
						Tag:    utils.RequestType,
						Path:   "*exp.RequestType",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.RequestType", utils.INFIELD_SEP),
						Layout: time.RFC3339,
					},
					{
						Tag:    utils.Tenant,
						Path:   "*exp.Tenant",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.Tenant", utils.INFIELD_SEP),
						Layout: time.RFC3339,
					},
					{
						Tag:    utils.Category,
						Path:   "*exp.Category",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.Category", utils.INFIELD_SEP),
						Layout: time.RFC3339,
					},
					{
						Tag:    utils.Account,
						Path:   "*exp.Account",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.Account", utils.INFIELD_SEP),
						Layout: time.RFC3339,
					},
					{
						Tag:    utils.Subject,
						Path:   "*exp.Subject",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.Subject", utils.INFIELD_SEP),
						Layout: time.RFC3339,
					},
					{
						Tag:    utils.Destination,
						Path:   "*exp.Destination",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.Destination", utils.INFIELD_SEP),
						Layout: time.RFC3339,
					},
					{
						Tag:    utils.SetupTime,
						Path:   "*exp.SetupTime",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.SetupTime", utils.INFIELD_SEP),
						Layout: "2006-01-02T15:04:05Z07:00",
					},
					{
						Tag:    utils.AnswerTime,
						Path:   "*exp.AnswerTime",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.AnswerTime", utils.INFIELD_SEP),
						Layout: "2006-01-02T15:04:05Z07:00",
					},
					{
						Tag:    utils.Usage,
						Path:   "*exp.Usage",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.Usage", utils.INFIELD_SEP),
						Layout: time.RFC3339,
					},
					{
						Tag:    utils.Cost,
						Path:   "*exp.Cost",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.Cost{*round:4}", utils.INFIELD_SEP),
						Layout: time.RFC3339,
					},
				},
				Fields: []*FCTemplate{
					{
						Tag:    utils.CGRID,
						Path:   "*exp.CGRID",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.CGRID", utils.INFIELD_SEP),
						Layout: time.RFC3339,
					},
					{
						Tag:    utils.RunID,
						Path:   "*exp.RunID",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.RunID", utils.INFIELD_SEP),
						Layout: time.RFC3339,
					},
					{
						Tag:    utils.ToR,
						Path:   "*exp.ToR",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.ToR", utils.INFIELD_SEP),
						Layout: time.RFC3339,
					},
					{
						Tag:    utils.OriginID,
						Path:   "*exp.OriginID",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.OriginID", utils.INFIELD_SEP),
						Layout: time.RFC3339,
					},
					{
						Tag:    utils.RequestType,
						Path:   "*exp.RequestType",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.RequestType", utils.INFIELD_SEP),
						Layout: time.RFC3339,
					},
					{
						Tag:    utils.Tenant,
						Path:   "*exp.Tenant",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.Tenant", utils.INFIELD_SEP),
						Layout: time.RFC3339,
					},
					{
						Tag:    utils.Category,
						Path:   "*exp.Category",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.Category", utils.INFIELD_SEP),
						Layout: time.RFC3339,
					},
					{
						Tag:    utils.Account,
						Path:   "*exp.Account",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.Account", utils.INFIELD_SEP),
						Layout: time.RFC3339,
					},
					{
						Tag:    utils.Subject,
						Path:   "*exp.Subject",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.Subject", utils.INFIELD_SEP),
						Layout: time.RFC3339,
					},
					{
						Tag:    utils.Destination,
						Path:   "*exp.Destination",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.Destination", utils.INFIELD_SEP),
						Layout: time.RFC3339,
					},
					{
						Tag:    utils.SetupTime,
						Path:   "*exp.SetupTime",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.SetupTime", utils.INFIELD_SEP),
						Layout: "2006-01-02T15:04:05Z07:00",
					},
					{
						Tag:    utils.AnswerTime,
						Path:   "*exp.AnswerTime",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.AnswerTime", utils.INFIELD_SEP),
						Layout: "2006-01-02T15:04:05Z07:00",
					},
					{
						Tag:    utils.Usage,
						Path:   "*exp.Usage",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.Usage", utils.INFIELD_SEP),
						Layout: time.RFC3339,
					},
					{
						Tag:    utils.Cost,
						Path:   "*exp.Cost",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.Cost{*round:4}", utils.INFIELD_SEP),
						Layout: time.RFC3339,
					},
				},
				headerFields:  []*FCTemplate{},
				trailerFields: []*FCTemplate{},
				Opts:          make(map[string]interface{}),
			},
			{
				ID:            "CSVExporter",
				Type:          "*file_csv",
				Filters:       []string{},
				AttributeSIDs: []string{},
				Flags:         utils.FlagsWithParamsFromSlice([]string{utils.MetaDryRun}),
				ExportPath:    "/tmp/testCSV",
				Tenant:        nil,
				Timezone:      "UTC",
				Synchronous:   true,
				Attempts:      1,
				FieldSep:      ",",
				headerFields:  []*FCTemplate{},
				trailerFields: []*FCTemplate{},
				contentFields: []*FCTemplate{
					{
						Tag:    utils.CGRID,
						Path:   "*exp.CGRID",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.CGRID", utils.INFIELD_SEP),
						Layout: time.RFC3339,
					},
				},
				Opts: make(map[string]interface{}),
				Fields: []*FCTemplate{
					{Tag: utils.CGRID, Path: "*exp.CGRID", Type: utils.MetaVariable, Value: NewRSRParsersMustCompile("~*req.CGRID", utils.INFIELD_SEP), Layout: time.RFC3339},
				},
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
	if cgrCfg, err := NewDefaultCGRConfig(); err != nil {
		t.Error(err)
	} else if err := cgrCfg.eesCfg.loadFromJsonCfg(nil, cgrCfg.templates, cgrCfg.generalCfg.RSRSep, cgrCfg.dfltEvExp); err != nil {
		t.Error(err)
	} else if err := cgrCfg.eesCfg.loadFromJsonCfg(jsonCfg, cgrCfg.templates, cgrCfg.generalCfg.RSRSep, cgrCfg.dfltEvExp); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedCfg, cgrCfg.eesCfg) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expectedCfg), utils.ToJSON(cgrCfg.eesCfg))
	}
}

func TestEEsCfgloadFromJsonCfgCase2(t *testing.T) {
	jsonCfg := &EEsJsonCfg{
		Enabled:          utils.BoolPointer(true),
		Attributes_conns: &[]string{"*conn1", "*conn2"},
		Cache: &map[string]*CacheParamJsonCfg{
			utils.MetaFileCSV: {
				Limit:      utils.IntPointer(-2),
				Ttl:        utils.StringPointer("1s"),
				Static_ttl: utils.BoolPointer(false),
			},
		},
		Exporters: &[]*EventExporterJsonCfg{
			{
				Id:              utils.StringPointer("CSVExporter"),
				Type:            utils.StringPointer("*file_csv"),
				Filters:         &[]string{},
				Attribute_ids:   &[]string{},
				Flags:           &[]string{"*dryrun"},
				Export_path:     utils.StringPointer("/tmp/testCSV"),
				Tenant:          nil,
				Timezone:        utils.StringPointer("UTC"),
				Synchronous:     utils.BoolPointer(true),
				Attempts:        utils.IntPointer(1),
				Field_separator: utils.StringPointer(","),
				Fields: &[]*FcTemplateJsonCfg{
					{
						Tag:    utils.StringPointer(utils.AnswerTime),
						Path:   utils.StringPointer("*exp.AnswerTime"),
						Type:   utils.StringPointer(utils.MetaTemplate),
						Value:  utils.StringPointer("randomVal"),
						Layout: utils.StringPointer(time.RFC3339),
					},
					{
						Path:  utils.StringPointer("*req.CGRID"),
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
				TTL:       1 * time.Second,
				StaticTTL: false,
			},
		},
		Exporters: []*EventExporterCfg{
			{
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
				contentFields: []*FCTemplate{
					{
						Tag:    utils.CGRID,
						Path:   "*exp.CGRID",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.CGRID", utils.INFIELD_SEP),
						Layout: time.RFC3339,
					},
					{
						Tag:    utils.RunID,
						Path:   "*exp.RunID",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.RunID", utils.INFIELD_SEP),
						Layout: time.RFC3339,
					},
					{
						Tag:    utils.ToR,
						Path:   "*exp.ToR",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.ToR", utils.INFIELD_SEP),
						Layout: time.RFC3339,
					},
					{
						Tag:    utils.OriginID,
						Path:   "*exp.OriginID",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.OriginID", utils.INFIELD_SEP),
						Layout: time.RFC3339,
					},
					{
						Tag:    utils.RequestType,
						Path:   "*exp.RequestType",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.RequestType", utils.INFIELD_SEP),
						Layout: time.RFC3339,
					},
					{
						Tag:    utils.Tenant,
						Path:   "*exp.Tenant",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.Tenant", utils.INFIELD_SEP),
						Layout: time.RFC3339,
					},
					{
						Tag:    utils.Category,
						Path:   "*exp.Category",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.Category", utils.INFIELD_SEP),
						Layout: time.RFC3339,
					},
					{
						Tag:    utils.Account,
						Path:   "*exp.Account",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.Account", utils.INFIELD_SEP),
						Layout: time.RFC3339,
					},
					{
						Tag:    utils.Subject,
						Path:   "*exp.Subject",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.Subject", utils.INFIELD_SEP),
						Layout: time.RFC3339,
					},
					{
						Tag:    utils.Destination,
						Path:   "*exp.Destination",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.Destination", utils.INFIELD_SEP),
						Layout: time.RFC3339,
					},
					{
						Tag:    utils.SetupTime,
						Path:   "*exp.SetupTime",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.SetupTime", utils.INFIELD_SEP),
						Layout: "2006-01-02T15:04:05Z07:00",
					},
					{
						Tag:    utils.AnswerTime,
						Path:   "*exp.AnswerTime",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.AnswerTime", utils.INFIELD_SEP),
						Layout: "2006-01-02T15:04:05Z07:00",
					},
					{
						Tag:    utils.Usage,
						Path:   "*exp.Usage",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.Usage", utils.INFIELD_SEP),
						Layout: time.RFC3339,
					},
					{
						Tag:    utils.Cost,
						Path:   "*exp.Cost",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.Cost{*round:4}", utils.INFIELD_SEP),
						Layout: time.RFC3339,
					},
				},
				Fields: []*FCTemplate{
					{
						Tag:    utils.CGRID,
						Path:   "*exp.CGRID",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.CGRID", utils.INFIELD_SEP),
						Layout: time.RFC3339,
					},
					{
						Tag:    utils.RunID,
						Path:   "*exp.RunID",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.RunID", utils.INFIELD_SEP),
						Layout: time.RFC3339,
					},
					{
						Tag:    utils.ToR,
						Path:   "*exp.ToR",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.ToR", utils.INFIELD_SEP),
						Layout: time.RFC3339,
					},
					{
						Tag:    utils.OriginID,
						Path:   "*exp.OriginID",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.OriginID", utils.INFIELD_SEP),
						Layout: time.RFC3339,
					},
					{
						Tag:    utils.RequestType,
						Path:   "*exp.RequestType",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.RequestType", utils.INFIELD_SEP),
						Layout: time.RFC3339,
					},
					{
						Tag:    utils.Tenant,
						Path:   "*exp.Tenant",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.Tenant", utils.INFIELD_SEP),
						Layout: time.RFC3339,
					},
					{
						Tag:    utils.Category,
						Path:   "*exp.Category",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.Category", utils.INFIELD_SEP),
						Layout: time.RFC3339,
					},
					{
						Tag:    utils.Account,
						Path:   "*exp.Account",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.Account", utils.INFIELD_SEP),
						Layout: time.RFC3339,
					},
					{
						Tag:    utils.Subject,
						Path:   "*exp.Subject",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.Subject", utils.INFIELD_SEP),
						Layout: time.RFC3339,
					},
					{
						Tag:    utils.Destination,
						Path:   "*exp.Destination",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.Destination", utils.INFIELD_SEP),
						Layout: time.RFC3339,
					},
					{
						Tag:    utils.SetupTime,
						Path:   "*exp.SetupTime",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.SetupTime", utils.INFIELD_SEP),
						Layout: "2006-01-02T15:04:05Z07:00",
					},
					{
						Tag:    utils.AnswerTime,
						Path:   "*exp.AnswerTime",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.AnswerTime", utils.INFIELD_SEP),
						Layout: "2006-01-02T15:04:05Z07:00",
					},
					{
						Tag:    utils.Usage,
						Path:   "*exp.Usage",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.Usage", utils.INFIELD_SEP),
						Layout: time.RFC3339,
					},
					{
						Tag:    utils.Cost,
						Path:   "*exp.Cost",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.Cost{*round:4}", utils.INFIELD_SEP),
						Layout: time.RFC3339,
					},
				},
				headerFields:  []*FCTemplate{},
				trailerFields: []*FCTemplate{},
				Opts:          make(map[string]interface{}),
			},
			{
				ID:            "CSVExporter",
				Type:          "*file_csv",
				Filters:       []string{},
				AttributeSIDs: []string{},
				Flags:         utils.FlagsWithParamsFromSlice([]string{utils.MetaDryRun}),
				ExportPath:    "/tmp/testCSV",
				Tenant:        nil,
				Timezone:      "UTC",
				Synchronous:   true,
				Attempts:      1,
				FieldSep:      ",",
				headerFields:  []*FCTemplate{},
				trailerFields: []*FCTemplate{},
				contentFields: []*FCTemplate{
					{
						Tag:    utils.CGRID,
						Path:   "*exp.CGRID",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.CGRID", utils.INFIELD_SEP),
						Layout: time.RFC3339,
					},
				},
				Opts: make(map[string]interface{}),
				Fields: []*FCTemplate{
					{
						Tag:    utils.CGRID,
						Path:   "*exp.CGRID",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.CGRID", utils.INFIELD_SEP),
						Layout: time.RFC3339,
					},
					{
						Tag:    "*req.CGRID",
						Path:   "*req.CGRID",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("1", utils.INFIELD_SEP),
						Layout: time.RFC3339,
					},
				},
			},
		},
	}
	msgTemplates := map[string][]*FCTemplate{
		"randomVal": {
			{
				Tag:    utils.CGRID,
				Path:   "*exp.CGRID",
				Type:   utils.MetaVariable,
				Value:  NewRSRParsersMustCompile("~*req.CGRID", utils.INFIELD_SEP),
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
	if jsnCfg, err := NewDefaultCGRConfig(); err != nil {
		t.Error(err)
	} else if err = jsnCfg.eesCfg.loadFromJsonCfg(jsonCfg, msgTemplates, jsnCfg.generalCfg.RSRSep, jsnCfg.dfltEvExp); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedCfg, jsnCfg.eesCfg) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expectedCfg), utils.ToJSON(jsnCfg.eesCfg))
	}
}

func TestEEsCfgAsMapInterface(t *testing.T) {
	cfgJSONStr := `{
      "ees": {									
	        "enabled": true,						
            "attributes_conns":["*conn1","*conn2"],					
            "cache": {
		          "*file_csv": {"limit": -2, "precache": false, "replicate": false, "ttl": "1s", "static_ttl": false}
            },
            "exporters": [
            {
                  "id": "CSVExporter",									
			      "type": "*file_csv",									
                  "export_path": "/tmp/testCSV",			
			      "opts": {},											
			      "tenant": "~*req.Destination1",										
			      "timezone": "UTC",										
			      "filters": [],										
			      "flags": [],										
			      "attribute_ids": [],								
			      "attribute_context": "",							
			      "synchronous": false,								
			      "attempts": 1,										
			      "field_separator": ",",								
			      "fields":[
                      {"tag": "CGRID", "path": "*exp.CGRID", "type": "*variable", "value": "~*req.CGRID"}
                  ]
            }]
	  }
    }`
	eMap := map[string]interface{}{
		utils.EnabledCfg:         true,
		utils.AttributeSConnsCfg: []string{"*conn1", "*conn2"},
		utils.CacheCfg: map[string]interface{}{
			utils.MetaFileCSV: map[string]interface{}{
				utils.LimitCfg:     -2,
				utils.PrecacheCfg:  false,
				utils.ReplicateCfg: false,
				utils.TTLCfg:       "1s",
				utils.StaticTTLCfg: false,
			},
		},
		utils.ExportersCfg: []map[string]interface{}{
			{
				utils.IdCfg:               "CSVExporter",
				utils.TypeCfg:             "*file_csv",
				utils.ExportPathCfg:       "/tmp/testCSV",
				utils.OptsCfg:             map[string]interface{}{},
				utils.TenantCfg:           NewRSRParsersMustCompile("~*req.Destination1", utils.INFIELD_SEP),
				utils.TimezoneCfg:         "UTC",
				utils.FiltersCfg:          []string{},
				utils.FlagsCfg:            utils.FlagsWithParams{},
				utils.AttributeIDsCfg:     []string{},
				utils.AttributeContextCfg: nil,
				utils.SynchronousCfg:      false,
				utils.AttemptsCfg:         1,
				utils.FieldSepCfg:         ",",
				utils.FieldsCfg: []map[string]interface{}{
					{
						utils.TagCfg:   utils.CGRID,
						utils.PathCfg:  "*exp.CGRID",
						utils.TypeCfg:  utils.MetaVariable,
						utils.ValueCfg: "~*req.CGRID",
					},
				},
			},
		},
	}
	if cgrCfg, err := NewCGRConfigFromJsonStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else {
		rcv := cgrCfg.eesCfg.AsMapInterface(cgrCfg.generalCfg.RSRSep)
		if !reflect.DeepEqual(eMap[utils.ExportersCfg].([]map[string]interface{})[0][utils.Flags],
			rcv[utils.ExportersCfg].([]map[string]interface{})[0][utils.Flags]) {
			t.Errorf("Expecetd %+v, received %+v", eMap[utils.ExportersCfg].([]map[string]interface{})[0][utils.Flags],
				rcv[utils.ExportersCfg].([]map[string]interface{})[0][utils.Flags])
		} else if !reflect.DeepEqual(eMap[utils.ExportersCfg].([]map[string]interface{})[0][utils.FieldsCfg].([]map[string]interface{})[0][utils.ValueCfg],
			rcv[utils.ExportersCfg].([]map[string]interface{})[0][utils.FieldsCfg].([]map[string]interface{})[0][utils.ValueCfg]) {
			t.Errorf("Expected %+v, received %+v", eMap[utils.ExportersCfg].([]map[string]interface{})[0][utils.FieldsCfg].([]map[string]interface{})[0][utils.ValueCfg],
				rcv[utils.ExportersCfg].([]map[string]interface{})[0][utils.FieldsCfg].([]map[string]interface{})[0][utils.ValueCfg])
		} else if !reflect.DeepEqual(eMap[utils.AttributeSConnsCfg], rcv[utils.AttributeSConnsCfg]) {
			t.Errorf("Expected %+v, received %+v", eMap[utils.AttributeSConnsCfg], rcv[utils.AttributeSConnsCfg])
		} else if !reflect.DeepEqual(eMap[utils.CacheCfg].(map[string]interface{})[utils.MetaFileCSV], rcv[utils.CacheCfg].(map[string]interface{})[utils.MetaFileCSV]) {
			t.Errorf("Expected %+v \n, received %+v", eMap[utils.CacheCfg].(map[string]interface{})[utils.MetaFileCSV], rcv[utils.CacheCfg].(map[string]interface{})[utils.MetaFileCSV])
		}
	}
}
