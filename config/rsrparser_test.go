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
	ruleStr := `Value1;Value2;~Header3(Val3&!Val4);~Header4:s/a/${1}b/{*duration_seconds&*round:2}(b&c);Value5{*duration_seconds&*round:2}`
	eRSRParsers := RSRParsers{
		&RSRParser{Rules: "Value1", AllFiltersMatch: true, path: "Value1"},
		&RSRParser{Rules: "Value2", AllFiltersMatch: true, path: "Value2"},
		&RSRParser{Rules: "~Header3(Val3&!Val4)", AllFiltersMatch: true, path: "~Header3",
			filters: utils.RSRFilters{utils.NewRSRFilterMustCompile("Val3"),
				utils.NewRSRFilterMustCompile("!Val4")}},

		&RSRParser{Rules: "~Header4:s/a/${1}b/{*duration_seconds&*round:2}(b&c)", AllFiltersMatch: true,
			path: "~Header4",
			rsrRules: []*utils.ReSearchReplace{{
				SearchRegexp:    regexp.MustCompile(`a`),
				ReplaceTemplate: "${1}b"}},
			converters: utils.DataConverters{utils.NewDataConverterMustCompile("*duration_seconds"),
				utils.NewDataConverterMustCompile("*round:2")},
			filters: utils.RSRFilters{utils.NewRSRFilterMustCompile("b"),
				utils.NewRSRFilterMustCompile("c")},
		},

		&RSRParser{Rules: "Value5{*duration_seconds&*round:2}", AllFiltersMatch: true,
			path: "Value5",
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
		Rules: "~Header4:s/a/${1}b/{*duration_seconds&*round:2}(b&c)",
		path:  "~Header4",
		rsrRules: []*utils.ReSearchReplace{{
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

// TestRSRParsersParseInnerBraces makes sure the inner braces are allowed in a filter rule
func TestRSRParsersParseInnerBracket(t *testing.T) {
	rule := "~*req.Service-Information.IN-Information.CalledPartyAddress(~^(00)*(33|0)890240004$)"
	prsr, err := NewRSRParser(rule, true)
	if err != nil {
		t.Error(err)
	}
	expAttrName := "*req.Service-Information.IN-Information.CalledPartyAddress"
	if prsr.AttrName() != expAttrName {
		t.Errorf("expecting: %s, received: %s", expAttrName, prsr.AttrName())
	}
}

func TestNewRSRParsersConstant(t *testing.T) {
	ruleStr := "`>;q=0.7;expires=3600`"
	eRSRParsers := RSRParsers{
		&RSRParser{Rules: ">;q=0.7;expires=3600", AllFiltersMatch: true, path: ">;q=0.7;expires=3600"},
	}
	if rsrParsers, err := NewRSRParsers(ruleStr, true, utils.INFIELD_SEP); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if !reflect.DeepEqual(eRSRParsers, rsrParsers) {
		t.Errorf("expecting: %+v, received: %+v", eRSRParsers, rsrParsers)
	} else if out, err := rsrParsers.ParseDataProvider(utils.MapStorage{}); err != nil {
		t.Error(err)
	} else if expected := ">;q=0.7;expires=3600"; out != expected {
		t.Errorf("Expected %+v ,received %+v", expected, out)
	}
}

func TestNewRSRParsersConstant2(t *testing.T) {
	ruleStr := "constant;something`>;q=0.7;expires=3600`new;constant"
	if rsrParsers, err := NewRSRParsers(ruleStr, true, utils.INFIELD_SEP); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if out, err := rsrParsers.ParseDataProvider(utils.MapStorage{}); err != nil {
		t.Error(err)
	} else if expected := "constantsomething>;q=0.7;expires=3600newconstant"; out != expected {
		t.Errorf("Expected %q ,received %q", expected, out)
	}

	ruleStr = "constant;`>;q=0.7;expires=3600`;constant"
	if rsrParsers, err := NewRSRParsers(ruleStr, true, utils.INFIELD_SEP); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if out, err := rsrParsers.ParseDataProvider(utils.MapStorage{}); err != nil {
		t.Error(err)
	} else if expected := "constant>;q=0.7;expires=3600constant"; out != expected {
		t.Errorf("Expected %q ,received %q", expected, out)
	}

	ruleStr = "constant;`>;q=0.7;expires=3600`constant"
	if rsrParsers, err := NewRSRParsers(ruleStr, true, utils.INFIELD_SEP); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if out, err := rsrParsers.ParseDataProvider(utils.MapStorage{}); err != nil {
		t.Error(err)
	} else if expected := "constant>;q=0.7;expires=3600constant"; out != expected {
		t.Errorf("Expected %q ,received %q", expected, out)
	}
}

func TestRSRParserCompileConstant(t *testing.T) {
	ePrsr := &RSRParser{
		Rules:           ":>;q=0.7;expires=3600",
		AllFiltersMatch: true,
		path:            ":>;q=0.7;expires=3600",
	}
	prsr := &RSRParser{
		Rules:           ":>;q=0.7;expires=3600",
		AllFiltersMatch: true,
	}
	if err := prsr.Compile(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(ePrsr, prsr) {
		t.Errorf("expecting: %+v, received: %+v", ePrsr, prsr)
	}
}

func TestNewRSRParsersParseDataProviderWithInterfaces(t *testing.T) {
	ruleStr := "~;*accounts.;~*req.Account"
	if rsrParsers, err := NewRSRParsers(ruleStr, true, utils.INFIELD_SEP); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if out, err := rsrParsers.ParseDataProviderWithInterfaces(
		utils.MapStorage{
			utils.MetaReq: utils.MapStorage{utils.Account: "1001"},
		}); err != nil {
		t.Error(err)
	} else if expected := "~*accounts.1001"; out != expected {
		t.Errorf("Expected %q ,received %q", expected, out)
	}
}
