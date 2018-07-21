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
package utils

import (
	"reflect"
	"regexp"
	"testing"
)

func TestNewRSRParsers(t *testing.T) {
	ruleStr := `Value1;Heade2=Value2;~Header3(Val3&!Val4);~Header4:s/a/${1}b/{*duration_seconds&*round:2}(b&c);Value5{*duration_seconds&*round:2}`
	eRSRParsers := RSRParsers{
		&RSRParser{Rules: "Value1", attrValue: "Value1"},
		&RSRParser{Rules: "Heade2=Value2", attrName: "Heade2", attrValue: "Value2"},
		&RSRParser{Rules: "~Header3(Val3&!Val4)", attrName: "Header3",
			filters: RSRFilters{NewRSRFilterMustCompile("Val3"),
				NewRSRFilterMustCompile("!Val4")}},

		&RSRParser{Rules: "~Header4:s/a/${1}b/{*duration_seconds&*round:2}(b&c)", attrName: "Header4",
			rsrRules: []*ReSearchReplace{
				&ReSearchReplace{
					SearchRegexp:    regexp.MustCompile(`a`),
					ReplaceTemplate: "${1}b"}},
			converters: DataConverters{NewDataConverterMustCompile("*duration_seconds"),
				NewDataConverterMustCompile("*round:2")},
			filters: RSRFilters{NewRSRFilterMustCompile("b"),
				NewRSRFilterMustCompile("c")},
		},

		&RSRParser{Rules: "Value5{*duration_seconds&*round:2}", attrValue: "Value5",
			converters: DataConverters{NewDataConverterMustCompile("*duration_seconds"),
				NewDataConverterMustCompile("*round:2")},
		},
	}
	if rsrParsers, err := NewRSRParsers(ruleStr, true); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if !reflect.DeepEqual(eRSRParsers, rsrParsers) {
		t.Errorf("expecting: %+v, received: %+v", eRSRParsers, rsrParsers)
	}
}

func TestRSRParserCompile(t *testing.T) {
	ePrsr := &RSRParser{
		Rules:    "~Header4:s/a/${1}b/{*duration_seconds&*round:2}(b&c)",
		attrName: "Header4",
		rsrRules: []*ReSearchReplace{
			&ReSearchReplace{
				SearchRegexp:    regexp.MustCompile(`a`),
				ReplaceTemplate: "${1}b"}},
		converters: DataConverters{NewDataConverterMustCompile("*duration_seconds"),
			NewDataConverterMustCompile("*round:2")},
		filters: RSRFilters{NewRSRFilterMustCompile("b"),
			NewRSRFilterMustCompile("c")},
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
	prsrs := NewRSRParsersMustCompile("~Header1;|;~Header2", true)
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
