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
	"reflect"
	"testing"

	"github.com/cgrates/cgrates/utils"
)

func TestCdreCfgClone(t *testing.T) {
	cgrIDRsrs := NewRSRParsersMustCompile("cgrid", true, utils.INFIELD_SEP)
	runIDRsrs := NewRSRParsersMustCompile("runid", true, utils.INFIELD_SEP)
	initContentFlds := []*FCTemplate{
		{Tag: "CgrId",
			Type:  utils.META_COMPOSED,
			Path:  "cgrid",
			Value: cgrIDRsrs},
		{Tag: "RunId",
			Type:  utils.META_COMPOSED,
			Path:  "runid",
			Value: runIDRsrs},
	}
	for _, v := range initContentFlds {
		v.ComputePath()
	}
	initCdreCfg := &CdreCfg{
		ExportFormat:   utils.MetaFileCSV,
		ExportPath:     "/var/spool/cgrates/cdre",
		Synchronous:    true,
		Attempts:       2,
		FieldSeparator: rune(utils.CSV_SEP),
		Fields:         initContentFlds,
	}
	eClnContentFlds := []*FCTemplate{
		{Tag: "CgrId",
			Type:  utils.META_COMPOSED,
			Path:  "cgrid",
			Value: cgrIDRsrs},
		{Tag: "RunId",
			Type:  utils.META_COMPOSED,
			Path:  "runid",
			Value: runIDRsrs},
	}
	for _, v := range eClnContentFlds {
		v.ComputePath()
	}
	eClnCdreCfg := &CdreCfg{
		ExportFormat:   utils.MetaFileCSV,
		ExportPath:     "/var/spool/cgrates/cdre",
		Synchronous:    true,
		Attempts:       2,
		Filters:        []string{},
		FieldSeparator: rune(utils.CSV_SEP),
		Fields:         eClnContentFlds,
	}
	clnCdreCfg := initCdreCfg.Clone()
	if !reflect.DeepEqual(eClnCdreCfg, clnCdreCfg) {
		t.Errorf("Cloned result: %+v", clnCdreCfg)
	}
	initContentFlds[0].Tag = "Destination"
	if !reflect.DeepEqual(eClnCdreCfg, clnCdreCfg) { // MOdifying a field after clone should not affect cloned instance
		t.Errorf("Cloned result: %+v", clnCdreCfg)
	}
	clnCdreCfg.Fields[0].Path = "destination"
	if initCdreCfg.Fields[0].Path != "cgrid" {
		t.Error("Unexpected change of Path: ", initCdreCfg.Fields[0].Path)
	}

}

func TestCdreCfgloadFromJsonCfg(t *testing.T) {
	var lstcfg, expected CdreCfg
	if err := lstcfg.loadFromJsonCfg(nil, utils.INFIELD_SEP); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(lstcfg, expected) {
		t.Errorf("Expected: %+v ,received: %+v", expected, lstcfg)
	}
	if err := lstcfg.loadFromJsonCfg(new(CdreJsonCfg), utils.INFIELD_SEP); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(lstcfg, expected) {
		t.Errorf("Expected: %+v ,received: %+v", expected, lstcfg)
	}
	cfgJSONStr := `{
"cdre": {
	"*default": {
		"export_format": "*file_csv",					// exported CDRs format <*file_csv|*file_fwv|*http_post|*http_json_cdr|*http_json_map|*amqp_json_cdr|*amqp_json_map>
		"export_path": "/var/spool/cgrates/cdre",		// path where the exported CDRs will be placed
		"filters" :[],									// new filters for cdre
		"tenant": "cgrates.org",						// tenant used in filterS.Pass
		"synchronous": false,							// block processing until export has a result
		"attempts": 1,									// Number of attempts if not success
		"field_separator": ",",							// used field separator in some export formats, eg: *file_csv
		"fields": [								// template of the exported content fields
			{"path": "*exp.CGRID", "type": "*composed", "value": "~*req.CGRID"},
		],
	},
},
}`
	val, err := NewRSRParsers("~*req.CGRID", true, utils.INFIELD_SEP)
	if err != nil {
		t.Error(err)
	}
	expected = CdreCfg{
		ExportFormat:   "*file_csv",
		ExportPath:     "/var/spool/cgrates/cdre",
		Filters:        []string{},
		Tenant:         "cgrates.org",
		Attempts:       1,
		FieldSeparator: utils.CSV_SEP,
		Fields: []*FCTemplate{{
			Path:  "*exp.CGRID",
			Tag:   "*exp.CGRID",
			Type:  "*composed",
			Value: val,
		}},
	}
	expected.Fields[0].ComputePath()
	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnCdreCfg, err := jsnCfg.CdreJsonCfgs(); err != nil {
		t.Error(err)
	} else if err = lstcfg.loadFromJsonCfg(jsnCdreCfg[utils.MetaDefault], utils.INFIELD_SEP); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, lstcfg) {
		t.Errorf("Expected: %+v , received: %+v", utils.ToJSON(expected), utils.ToJSON(lstcfg))
	}
}

func TestCdreCfgAsMapInterface(t *testing.T) {
	var cdre CdreCfg
	cfgJSONStr := `{
		"cdre": {												
		"*default": {
			"export_format": "*file_csv",					
			"export_path": "/var/spool/cgrates/cdre",		
			"filters" :[],									
			"tenant": "",									
			"synchronous": false,							
			"attempts": 1,									
			"field_separator": ",",							
			"attributes_context": "",						
			"fields": [										
				{"path": "*exp.CGRID", "type": "*variable", "value": "~*req.CGRID"},
			],
		},
	},
}`
	eMap := map[string]any{
		"export_format":      "*file_csv",
		"export_path":        "/var/spool/cgrates/cdre",
		"filters":            []string{},
		"tenant":             "",
		"synchronous":        false,
		"attempts":           1,
		"field_separator":    ",",
		"attributes_context": "",
		"fields": []map[string]any{
			{
				"path":  "*exp.CGRID",
				"tag":   "*exp.CGRID",
				"type":  "*variable",
				"value": "~*req.CGRID",
			},
		},
	}

	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if cdreCfg, err := jsnCfg.CdreJsonCfgs(); err != nil {
		t.Error(err)
	} else if err = cdre.loadFromJsonCfg(cdreCfg["*default"], utils.EmptyString); err != nil {
		t.Error(err)
	} else if rcv := cdre.AsMapInterface(""); !reflect.DeepEqual(eMap, rcv) {
		t.Errorf("\nExpected: %+v\nReceived: %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}

}

func TestCdreCfgloadFromJsonCfg2(t *testing.T) {
	type args struct {
		js  *CdreJsonCfg
		sep string
	}

	s := CdreCfg{}
	str := "test`"
	js := CdreJsonCfg{
		Filters: &[]string{"val1", "val2"},
		Fields: &[]*FcTemplateJsonCfg{
			{
				Value: &str,
			},
		},
	}

	tests := []struct {
		name string
		args args
		exp  string
	}{
		{
			name: "populate filters",
			args: args{js: &js, sep: ""},
			exp:  "Unclosed unspilit syntax",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := s.loadFromJsonCfg(tt.args.js, tt.args.sep)

			if err.Error() != tt.exp {
				t.Errorf("received %s, expected %s", err, tt.exp)
			}

			exp := []string{"val1", "val2"}
			if !reflect.DeepEqual(s.Filters, exp) {
				t.Errorf("received %v, expected %v", s.Filters, exp)
			}
		})
	}
}

func TestCdreCfgClone2(t *testing.T) {
	str := "test"
	slc := []string{"val1", "val2"}
	bl := false
	nm := 1
	rn := 'a'

	s := CdreCfg{
		ExportFormat:      str,
		ExportPath:        str,
		Filters:           slc,
		Tenant:            str,
		AttributeSContext: "",
		Synchronous:       bl,
		Attempts:          nm,
		FieldSeparator:    rn,
		Fields:            []*FCTemplate{},
	}

	rcv := s.Clone()

	if !reflect.DeepEqual(*rcv, s) {
		t.Errorf("received %v, expected %v", *rcv, s)
	}
}
