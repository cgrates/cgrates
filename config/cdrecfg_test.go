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

	"github.com/cgrates/cgrates/utils"
)

func TestCdreCfgClone(t *testing.T) {
	cgrIDRsrs := NewRSRParsersMustCompile("cgrid", true, utils.INFIELD_SEP)
	runIDRsrs := NewRSRParsersMustCompile("runid", true, utils.INFIELD_SEP)
	initContentFlds := []*FCTemplate{
		{Tag: "CgrId",
			Type:      utils.META_COMPOSED,
			Path:      utils.PathItems{{Field: "cgrid"}},
			PathSlice: []string{"cgrid"},
			Value:     cgrIDRsrs},
		{Tag: "RunId",
			Type:      utils.META_COMPOSED,
			Path:      utils.PathItems{{Field: "runid"}},
			PathSlice: []string{"runid"},
			Value:     runIDRsrs},
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
			Type:      utils.META_COMPOSED,
			Path:      utils.PathItems{{Field: "cgrid"}},
			PathSlice: []string{"cgrid"},
			Value:     cgrIDRsrs},
		{Tag: "RunId",
			Type:      utils.META_COMPOSED,
			Path:      utils.PathItems{{Field: "runid"}},
			PathSlice: []string{"runid"},
			Value:     runIDRsrs},
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
	clnCdreCfg.Fields[0].Path[0].Field = "destination"
	if initCdreCfg.Fields[0].Path[0].Field != "cgrid" {
		t.Error("Unexpected change of Path:  ", initCdreCfg.Fields[0].Path.String())
	}

}

func TestCdreCfgloadFromJsonCfg(t *testing.T) {
	var lstcfg, expected CdreCfg
	if err := lstcfg.loadFromJsonCfg(nil, utils.INFIELD_SEP); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(lstcfg, expected) {
		t.Errorf("Expected: %+v ,recived: %+v", expected, lstcfg)
	}
	if err := lstcfg.loadFromJsonCfg(new(CdreJsonCfg), utils.INFIELD_SEP); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(lstcfg, expected) {
		t.Errorf("Expected: %+v ,recived: %+v", expected, lstcfg)
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
			Path:      utils.PathItems{{Field: utils.MetaExp}, {Field: "CGRID"}},
			PathSlice: []string{utils.MetaExp, "CGRID"},
			Tag:       "*exp.CGRID",
			Type:      "*composed",
			Value:     val,
		}},
	}
	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnCdreCfg, err := jsnCfg.CdreJsonCfgs(); err != nil {
		t.Error(err)
	} else if err = lstcfg.loadFromJsonCfg(jsnCdreCfg[utils.MetaDefault], utils.INFIELD_SEP); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, lstcfg) {
		t.Errorf("Expected: %+v , recived: %+v", utils.ToJSON(expected), utils.ToJSON(lstcfg))
	}
}
