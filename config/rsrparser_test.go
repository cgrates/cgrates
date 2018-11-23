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
	"regexp"
	"testing"

	"github.com/cgrates/cgrates/utils"
)

func TestNewRSRParsers(t *testing.T) {
	ruleStr := `Value1;Heade2=Value2;~Header3(Val3&!Val4);~Header4:s/a/${1}b/{*duration_seconds&*round:2}(b&c);Value5{*duration_seconds&*round:2}`
	eRSRParsers := RSRParsers{
		&RSRParser{Rules: "Value1", AllFiltersMatch: true, attrValue: "Value1"},
		&RSRParser{Rules: "Heade2=Value2", AllFiltersMatch: true, attrName: "Heade2", attrValue: "Value2"},
		&RSRParser{Rules: "~Header3(Val3&!Val4)", AllFiltersMatch: true, attrName: "Header3",
			filters: utils.RSRFilters{utils.NewRSRFilterMustCompile("Val3"),
				utils.NewRSRFilterMustCompile("!Val4")}},

		&RSRParser{Rules: "~Header4:s/a/${1}b/{*duration_seconds&*round:2}(b&c)", AllFiltersMatch: true,
			attrName: "Header4",
			rsrRules: []*utils.ReSearchReplace{
				&utils.ReSearchReplace{
					SearchRegexp:    regexp.MustCompile(`a`),
					ReplaceTemplate: "${1}b"}},
			converters: utils.DataConverters{utils.NewDataConverterMustCompile("*duration_seconds"),
				utils.NewDataConverterMustCompile("*round:2")},
			filters: utils.RSRFilters{utils.NewRSRFilterMustCompile("b"),
				utils.NewRSRFilterMustCompile("c")},
		},

		&RSRParser{Rules: "Value5{*duration_seconds&*round:2}", AllFiltersMatch: true,
			attrValue: "Value5",
			converters: utils.DataConverters{utils.NewDataConverterMustCompile("*duration_seconds"),
				utils.NewDataConverterMustCompile("*round:2")},
		},
	}
	if rsrParsers, err := NewRSRParsers(ruleStr, true, utils.INFIELD_SEP); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if !reflect.DeepEqual(eRSRParsers, rsrParsers) {
		t.Errorf("expecting: %+v, received: %+v", eRSRParsers, rsrParsers)
	}
}

func TestRSRParserCompile(t *testing.T) {
	ePrsr := &RSRParser{
		Rules:    "~Header4:s/a/${1}b/{*duration_seconds&*round:2}(b&c)",
		attrName: "Header4",
		rsrRules: []*utils.ReSearchReplace{
			&utils.ReSearchReplace{
				SearchRegexp:    regexp.MustCompile(`a`),
				ReplaceTemplate: "${1}b"}},
		converters: utils.DataConverters{utils.NewDataConverterMustCompile("*duration_seconds"),
			utils.NewDataConverterMustCompile("*round:2")},
		filters: utils.RSRFilters{utils.NewRSRFilterMustCompile("b"),
			utils.NewRSRFilterMustCompile("c")},
	}
	prsr := &RSRParser{
		Rules: "~Header4:s/a/${1}b/{*duration_seconds&*round:2}(b&c)",
	}
	if err := prsr.Compile(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(ePrsr, prsr) {
		t.Errorf("expecting: %+v, received: %+v", ePrsr, prsr)
	}
}

func TestRSRParsersParseEvent(t *testing.T) {
	prsrs := NewRSRParsersMustCompile("~Header1;|;~Header2", true, utils.INFIELD_SEP)
	ev := map[string]interface{}{
		"Header1": "Value1",
		"Header2": "Value2",
	}
	eOut := "Value1|Value2"
	if out, err := prsrs.ParseEvent(ev); err != nil {
		t.Error(err)
	} else if eOut != out {
		t.Errorf("expecting: %s, received: %s", eOut, out)
	}
}

func TestRSRParserConstant(t *testing.T) {
	rule := "cgrates.org"
	rsrParsers, err := NewRSRParsers(rule, true, utils.INFIELD_SEP)
	if err != nil {
		t.Error("Unexpected error: ", err.Error())
	}
	if out, err := rsrParsers.ParseValue(""); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if out != "cgrates.org" {
		t.Errorf("expecting: cgrates.org , received: %+v", out)
	}
}

func TestRSRParserNotConstant(t *testing.T) {
	rule := "~Header1;~Header2"
	rsrParsers, err := NewRSRParsers(rule, true, utils.INFIELD_SEP)
	if err != nil {
		t.Error("Unexpected error: ", err.Error())
	}
	if out, err := rsrParsers.ParseValue(""); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if out != "" {
		t.Errorf("expecting: EmptyString , received: %+v", out)
	}
}

func TestRSRParsersParseEvent2(t *testing.T) {
	prsrs := NewRSRParsersMustCompile("~Header1.Test;|;~Header2.Test", true, utils.INFIELD_SEP)
	ev := map[string]interface{}{
		"Header1.Test": "Value1",
		"Header2.Test": "Value2",
	}
	eOut := "Value1|Value2"
	if out, err := prsrs.ParseEvent(ev); err != nil {
		t.Error(err)
	} else if eOut != out {
		t.Errorf("expecting: %s, received: %s", eOut, out)
	}
}
