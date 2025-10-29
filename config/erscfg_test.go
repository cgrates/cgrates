/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/
package config

import (
	"fmt"
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
	for _, v := range initialOrig.Fields {
		v.ComputePath()
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
				SourcePath:     "/var/spool/cgrates/ers/in",
				ProcessedPath:  "/var/spool/cgrates/ers/out",
				XmlRootPath:    nil,
				Tenant:         nil,
				Timezone:       utils.EmptyString,
				Filters:        []string{},
				Flags:          utils.FlagsWithParams{},
				Fields: []*FCTemplate{
					{Tag: utils.ToR, Path: utils.MetaCgreq + utils.NestingSep + utils.ToR, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.2", true, utils.INFIELD_SEP), Mandatory: true},
					{Tag: utils.OriginID, Path: utils.MetaCgreq + utils.NestingSep + utils.OriginID, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.3", true, utils.INFIELD_SEP), Mandatory: true},
					{Tag: utils.RequestType, Path: utils.MetaCgreq + utils.NestingSep + utils.RequestType, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.4", true, utils.INFIELD_SEP), Mandatory: true},
					{Tag: utils.Tenant, Path: utils.MetaCgreq + utils.NestingSep + utils.Tenant, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.6", true, utils.INFIELD_SEP), Mandatory: true},
					{Tag: utils.Category, Path: utils.MetaCgreq + utils.NestingSep + utils.Category, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.7", true, utils.INFIELD_SEP), Mandatory: true},
					{Tag: utils.Account, Path: utils.MetaCgreq + utils.NestingSep + utils.Account, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.8", true, utils.INFIELD_SEP), Mandatory: true},
					{Tag: utils.Subject, Path: utils.MetaCgreq + utils.NestingSep + utils.Subject, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.9", true, utils.INFIELD_SEP), Mandatory: true},
					{Tag: utils.Destination, Path: utils.MetaCgreq + utils.NestingSep + utils.Destination, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.10", true, utils.INFIELD_SEP), Mandatory: true},
					{Tag: utils.SetupTime, Path: utils.MetaCgreq + utils.NestingSep + utils.SetupTime, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.11", true, utils.INFIELD_SEP), Mandatory: true},
					{Tag: utils.AnswerTime, Path: utils.MetaCgreq + utils.NestingSep + utils.AnswerTime, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.12", true, utils.INFIELD_SEP), Mandatory: true},
					{Tag: utils.Usage, Path: utils.MetaCgreq + utils.NestingSep + utils.Usage, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.13", true, utils.INFIELD_SEP), Mandatory: true},
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
				XmlRootPath:    nil,
				Tenant:         nil,
				Timezone:       utils.EmptyString,
				Filters:        nil,
				Flags:          utils.FlagsWithParams{},
				Fields: []*FCTemplate{
					{Tag: utils.ToR, Path: utils.MetaCgreq + utils.NestingSep + utils.ToR, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.2", true, utils.INFIELD_SEP), Mandatory: true},
					{Tag: utils.OriginID, Path: utils.MetaCgreq + utils.NestingSep + utils.OriginID, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.3", true, utils.INFIELD_SEP), Mandatory: true},
					{Tag: utils.RequestType, Path: utils.MetaCgreq + utils.NestingSep + utils.RequestType, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.4", true, utils.INFIELD_SEP), Mandatory: true},
					{Tag: utils.Tenant, Path: utils.MetaCgreq + utils.NestingSep + utils.Tenant, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.6", true, utils.INFIELD_SEP), Mandatory: true},
					{Tag: utils.Category, Path: utils.MetaCgreq + utils.NestingSep + utils.Category, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.7", true, utils.INFIELD_SEP), Mandatory: true},
					{Tag: utils.Account, Path: utils.MetaCgreq + utils.NestingSep + utils.Account, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.8", true, utils.INFIELD_SEP), Mandatory: true},
					{Tag: utils.Subject, Path: utils.MetaCgreq + utils.NestingSep + utils.Subject, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.9", true, utils.INFIELD_SEP), Mandatory: true},
					{Tag: utils.Destination, Path: utils.MetaCgreq + utils.NestingSep + utils.Destination, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.10", true, utils.INFIELD_SEP), Mandatory: true},
					{Tag: utils.SetupTime, Path: utils.MetaCgreq + utils.NestingSep + utils.SetupTime, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.11", true, utils.INFIELD_SEP), Mandatory: true},
					{Tag: utils.AnswerTime, Path: utils.MetaCgreq + utils.NestingSep + utils.AnswerTime, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.12", true, utils.INFIELD_SEP), Mandatory: true},
					{Tag: utils.Usage, Path: utils.MetaCgreq + utils.NestingSep + utils.Usage, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.13", true, utils.INFIELD_SEP), Mandatory: true},
				},
				CacheDumpFields: make([]*FCTemplate, 0),
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
		t.Errorf("Expected: %+v ,\n received: %+v", utils.ToJSON(expectedERsCfg), utils.ToJSON(cfg.ersCfg))
	}

}

func TestEventReaderSanitisation(t *testing.T) {
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
	eMap := map[string]any{
		"enabled":        true,
		"sessions_conns": []string{"conn1", "conn3"},
		"readers": []map[string]any{
			{
				"filters":                     []string{},
				"flags":                       map[string][]any{},
				"id":                          "*default",
				"partial_record_cache":        "0",
				"processed_path":              "/var/spool/cgrates/ers/out",
				"row_length":                  0,
				"run_delay":                   "0",
				"partial_cache_expiry_action": "",
				"source_path":                 "/var/spool/cgrates/ers/in",
				"tenant":                      "",
				"timezone":                    "",
				"cache_dump_fields":           []map[string]any{},
				"concurrent_requests":         1024,
				"type":                        "*none",
				"failed_calls_prefix":         "",
				"field_separator":             ",",
				"fields": []map[string]any{
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
			},
			{
				"cache_dump_fields":   []map[string]any{},
				"concurrent_requests": 1024,
				"type":                "*file_csv",
				"failed_calls_prefix": "",
				"field_separator":     ",",
				"fields": []map[string]any{
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
				"flags":                       map[string][]any{},
				"id":                          "file_reader1",
				"partial_record_cache":        "0",
				"processed_path":              "/tmp/ers/out",
				"row_length":                  0,
				"run_delay":                   "-1",
				"partial_cache_expiry_action": "",
				"source_path":                 "/tmp/ers/in",
				"tenant":                      "",
				"timezone":                    "",
			},
		},
	}
	if cfg, err := NewCGRConfigFromJsonStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cfg.ersCfg.AsMapInterface(utils.EmptyString); !reflect.DeepEqual(eMap, rcv) {
		t.Errorf("\nExpected: %+v\nReceived: %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}
}

func TestERSCFGClone(t *testing.T) {
	ev := EventReaderCfg{
		CacheDumpFields: []*FCTemplate{
			{Tag: "test"},
		},
	}
	ev2 := EventReaderCfg{}

	e := ERsCfg{
		Enabled:       true,
		SessionSConns: []string{"val1", "val2"},
		Readers:       []*EventReaderCfg{&ev, &ev2},
	}

	rcv := e.Clone()
	exp := &e

	if rcv.Enabled != exp.Enabled && !reflect.DeepEqual(rcv.SessionSConns, exp.SessionSConns) && !reflect.DeepEqual(rcv.Readers, exp.Readers) {
		t.Errorf("recived %v, expected %v", rcv, exp)
	}

}

func TestEventReadersCFGAppendERsReaders(t *testing.T) {

	er := EventReaderCfg{
		ID: "test",
	}

	e := ERsCfg{
		Enabled:       false,
		SessionSConns: []string{},
		Readers:       []*EventReaderCfg{&er},
	}

	id := "test"
	ten := "`test"

	ej := EventReaderJsonCfg{
		Id:     &id,
		Tenant: &ten,
	}

	erj := []*EventReaderJsonCfg{&ej}

	type args struct {
		jsnReaders *[]*EventReaderJsonCfg
		sep        string
		dfltRdrCfg *EventReaderCfg
	}

	tests := []struct {
		name string
		args args
		exp  error
	}{
		{
			name: "nil return",
			args: args{jsnReaders: nil, sep: "", dfltRdrCfg: nil},
			exp:  nil,
		},
		{
			name: "nil return",
			args: args{jsnReaders: &erj, sep: "", dfltRdrCfg: nil},
			exp:  fmt.Errorf("Unclosed unspilit syntax"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rcv := e.appendERsReaders(tt.args.jsnReaders, tt.args.sep, tt.args.dfltRdrCfg)

			if err != nil {
				if rcv.Error() != tt.exp.Error() {
					t.Errorf("recived %v, expected %v", rcv, tt.exp)
				}
			}
		})
	}
}

func TestEventReaderLoadFromJSON2(t *testing.T) {

	er := EventReaderCfg{
		ID: "test",
	}

	rn := "test"

	ej1 := EventReaderJsonCfg{
		Run_delay: &rn,
	}

	flt := []string{"val1", "val2"}
	flg := []string{"1:2:3:4:5"}

	ej2 := EventReaderJsonCfg{
		Filters: &flt,
		Flags:   &flg,
	}

	fcp := "test"
	prc := "test"

	ej3 := EventReaderJsonCfg{
		Failed_calls_prefix:  &fcp,
		Partial_record_cache: &prc,
	}

	vl := "`test"

	fcT := FcTemplateJsonCfg{
		Value: &vl,
	}

	pcea := "test"
	flds := []*FcTemplateJsonCfg{&fcT}

	ej4 := EventReaderJsonCfg{
		Partial_cache_expiry_action: &pcea,
		Fields:                      &flds,
	}

	ej5 := EventReaderJsonCfg{
		Cache_dump_fields: &flds,
	}

	type args struct {
		jsnCfg *EventReaderJsonCfg
		sep    string
	}

	tests := []struct {
		name string
		args args
		exp  error
	}{
		{
			name: "nil return",
			args: args{jsnCfg: nil, sep: ""},
			exp:  nil,
		},
		{
			name: "check error invalid duration",
			args: args{jsnCfg: &ej1, sep: ""},
			exp:  fmt.Errorf(`time: invalid duration "test"`),
		},
		{
			name: "check error unsupported format",
			args: args{jsnCfg: &ej2, sep: ""},
			exp:  utils.ErrUnsupportedFormat,
		},
		{
			name: "check error invalid duration",
			args: args{jsnCfg: &ej3, sep: ""},
			exp:  fmt.Errorf(`time: invalid duration "test"`),
		},
		{
			name: "check error in Fields",
			args: args{jsnCfg: &ej4, sep: ""},
			exp:  fmt.Errorf("Unclosed unspilit syntax"),
		},
		{
			name: "check error in cache dump fields",
			args: args{jsnCfg: &ej5, sep: ""},
			exp:  fmt.Errorf("Unclosed unspilit syntax"),
		},
	}

	for _, tt := range tests {
		rcv := er.loadFromJsonCfg(tt.args.jsnCfg, tt.args.sep)

		if rcv != nil {
			if rcv.Error() != tt.exp.Error() {
				t.Errorf("recived %v, expected %v", rcv, tt.exp)
			}
		}
	}
}

func TestEventReeadersCFGAsMapInterface(t *testing.T) {

	fct := FCTemplate{
		Tag: "test",
	}

	er := EventReaderCfg{
		Tenant: RSRParsers{&RSRParser{
			Rules: "test",
		}},
		Flags: utils.FlagsWithParams{"test": []string{"val1", "val2"}},
		CacheDumpFields: []*FCTemplate{
			&fct,
		},
		RunDelay:                 1 * time.Second,
		PartialRecordCache:       1 * time.Millisecond,
		ID:                       "test",
		Type:                     "test",
		RowLength:                1,
		FieldSep:                 "test",
		ConcurrentReqs:           1,
		SourcePath:               "test",
		ProcessedPath:            "test",
		XmlRootPath:              utils.HierarchyPath{"item1", "item2"},
		Timezone:                 "test",
		Filters:                  []string{},
		FailedCallsPrefix:        "!",
		PartialCacheExpiryAction: "test",
	}

	mp := map[string]any{
		utils.IDCfg:                       er.ID,
		utils.TypeCfg:                     er.Type,
		utils.RowLengthCfg:                er.RowLength,
		utils.FieldSepCfg:                 er.FieldSep,
		utils.RunDelayCfg:                 "1s",
		utils.ConcurrentReqsCfg:           er.ConcurrentReqs,
		utils.SourcePathCfg:               er.SourcePath,
		utils.ProcessedPathCfg:            er.ProcessedPath,
		utils.TenantCfg:                   "test",
		utils.TimezoneCfg:                 er.Timezone,
		utils.FiltersCfg:                  er.Filters,
		utils.FlagsCfg:                    map[string][]any{"test": {"val1", "val2"}},
		utils.FailedCallsPrefixCfg:        er.FailedCallsPrefix,
		utils.PartialRecordCacheCfg:       "1ms",
		utils.PartialCacheExpiryActionCfg: er.PartialCacheExpiryAction,
		utils.FieldsCfg:                   []map[string]any{},
		utils.CacheDumpFieldsCfg:          []map[string]any{fct.AsMapInterface("")},
		utils.XmlRootPathCfg:              []string{"item1", "item2"},
	}

	tests := []struct {
		name string
		arg  string
		exp  map[string]any
	}{
		{
			name: "test ErsCFG as map interface",
			arg:  "",
			exp:  mp,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rcv := er.AsMapInterface(tt.arg)

			if !reflect.DeepEqual(rcv, tt.exp) {
				t.Errorf("expected %v,\nreceived %v", tt.exp, rcv)
			}
		})
	}
}
