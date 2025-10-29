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
	"errors"
	"reflect"
	"regexp"
	"testing"

	"github.com/cgrates/cgrates/utils"
)

func TestNewRSRParsers(t *testing.T) {
	ruleStr := `Value1;Value2;~Header3(Val3&!Val4);~Header4:s/a/${1}b/{*duration_seconds&*round:2}(b&c);Value5{*duration_seconds&*round:2}`
	eRSRParsers := RSRParsers{
		&RSRParser{Rules: "Value1", AllFiltersMatch: true},
		&RSRParser{Rules: "Value2", AllFiltersMatch: true},
		&RSRParser{Rules: "~Header3(Val3&!Val4)", AllFiltersMatch: true, path: "Header3",
			filters: utils.RSRFilters{utils.NewRSRFilterMustCompile("Val3"),
				utils.NewRSRFilterMustCompile("!Val4")}},

		&RSRParser{Rules: "~Header4:s/a/${1}b/{*duration_seconds&*round:2}(b&c)", AllFiltersMatch: true,
			path: "Header4",
			rsrRules: []*utils.ReSearchReplace{
				{
					SearchRegexp:    regexp.MustCompile(`a`),
					ReplaceTemplate: "${1}b"}},
			converters: utils.DataConverters{utils.NewDataConverterMustCompile("*duration_seconds"),
				utils.NewDataConverterMustCompile("*round:2")},
			filters: utils.RSRFilters{utils.NewRSRFilterMustCompile("b"),
				utils.NewRSRFilterMustCompile("c")},
		},

		&RSRParser{Rules: "Value5{*duration_seconds&*round:2}", AllFiltersMatch: true,
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
		path:  "Header4",
		rsrRules: []*utils.ReSearchReplace{
			{
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
		&RSRParser{Rules: ">;q=0.7;expires=3600", AllFiltersMatch: true},
	}
	if rsrParsers, err := NewRSRParsers(ruleStr, true, utils.INFIELD_SEP); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if !reflect.DeepEqual(eRSRParsers, rsrParsers) {
		t.Errorf("expecting: %+v, received: %+v", eRSRParsers, rsrParsers)
	} else if out, err := rsrParsers.ParseDataProvider(utils.MapStorage{}, utils.NestingSep); err != nil {
		t.Error(err)
	} else if expected := ">;q=0.7;expires=3600"; out != expected {
		t.Errorf("Expected %+v ,received %+v", expected, out)
	}
}

func TestNewRSRParsersConstant2(t *testing.T) {
	ruleStr := "constant;something`>;q=0.7;expires=3600`new;constant"
	if rsrParsers, err := NewRSRParsers(ruleStr, true, utils.INFIELD_SEP); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if out, err := rsrParsers.ParseDataProvider(utils.MapStorage{}, utils.NestingSep); err != nil {
		t.Error(err)
	} else if expected := "constantsomething>;q=0.7;expires=3600newconstant"; out != expected {
		t.Errorf("Expected %q ,received %q", expected, out)
	}

	ruleStr = "constant;`>;q=0.7;expires=3600`;constant"
	if rsrParsers, err := NewRSRParsers(ruleStr, true, utils.INFIELD_SEP); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if out, err := rsrParsers.ParseDataProvider(utils.MapStorage{}, utils.NestingSep); err != nil {
		t.Error(err)
	} else if expected := "constant>;q=0.7;expires=3600constant"; out != expected {
		t.Errorf("Expected %q ,received %q", expected, out)
	}

	ruleStr = "constant;`>;q=0.7;expires=3600`constant"
	if rsrParsers, err := NewRSRParsers(ruleStr, true, utils.INFIELD_SEP); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if out, err := rsrParsers.ParseDataProvider(utils.MapStorage{}, utils.NestingSep); err != nil {
		t.Error(err)
	} else if expected := "constant>;q=0.7;expires=3600constant"; out != expected {
		t.Errorf("Expected %q ,received %q", expected, out)
	}
}

func TestRSRParserCompileConstant(t *testing.T) {
	ePrsr := &RSRParser{
		Rules:           "*constant:>;q=0.7;expires=3600",
		AllFiltersMatch: true,
	}
	prsr := &RSRParser{
		Rules:           "*constant:>;q=0.7;expires=3600",
		AllFiltersMatch: true,
	}
	if err := prsr.Compile(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(ePrsr, prsr) {
		t.Errorf("expecting: %+v, received: %+v", ePrsr, prsr)
	}
}

func TestRSRParserGetRule(t *testing.T) {
	p := RSRParser{
		Rules: "Test",
	}
	p2 := RSRParser{
		Rules: "Test2",
	}

	pp := RSRParsers{&p, &p2}

	rcv := pp.GetRule()
	exp := "Test.Test2"

	if rcv != exp {
		t.Errorf("expecting: %+v, received: %+v", exp, rcv)
	}

}

func TestRSRParsersCompile(t *testing.T) {

	p := RSRParser{
		Rules: "Test",
	}
	p2 := RSRParser{
		Rules: "Test2",
	}

	pp := RSRParsers{&p, &p2}

	rcv := pp.Compile()

	if rcv != nil {
		t.Errorf("expecting: %+v, received: %+v", nil, rcv)
	}
}

func TestRSRParserParseDataProviderWithInterface(t *testing.T) {

	p := RSRParser{
		Rules: "Test",
	}
	p2 := RSRParser{
		Rules: "Test2",
	}

	pp := RSRParsers{&p, &p2}

	rcv, err := pp.ParseDataProviderWithInterfaces(utils.MapStorage{}, ".")
	exp := "TestTest2"
	var expErr error = nil

	if err != expErr {
		t.Fatalf("recived %v, expected %v", err, expErr)
	}

	if rcv != exp {
		t.Errorf("recived %s, expected %s", rcv, exp)
	}

}

func TestRSRParserNewParserMustCompile(t *testing.T) {

	rcv := NewRSRParserMustCompile("test", false)
	exp, _ := NewRSRParser("test", false)

	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("recived %v, expected %v", rcv, exp)
	}
}

func TestRSRParserregexpMatched(t *testing.T) {

	p := RSRParser{
		Rules: "Test",
	}

	rcv := p.RegexpMatched()

	if rcv {
		t.Error("was expecting false")
	}
}

func TestRSRParserParseDataProviderAsFloat64(t *testing.T) {

	type args struct {
		dP        utils.DataProvider
		separator string
	}

	type exp struct {
		val float64
		err error
	}

	tests := []struct {
		name string
		args args
		exp  exp
	}{
		{
			name: "",
			args: args{dP: utils.MapStorage{}, separator: "."},
			exp:  exp{val: 0, err: errors.New("empty path in parser")},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			p := RSRParser{
				Rules: "Test",
			}

			rcv, err := p.ParseDataProviderAsFloat64(tt.args.dP, tt.args.separator)

			if err.Error() != tt.exp.err.Error() {
				t.Fatalf("recived %s, expected %s", err, tt.exp.err)
			}

			if rcv != tt.exp.val {
				t.Errorf("recived %v, expected %v", rcv, tt.exp.val)
			}
		})
	}
}

func TestRSRParserParseDataProvider(t *testing.T) {
	prsr := RSRParser{
		path: "test",
	}
	dP := &FWVProvider{
		req:   "test",
		cache: utils.MapStorage{"test": 1},
	}

	rcv, err := prsr.ParseDataProvider(dP, "")
	if err != nil {
		if err.Error() != "Invalid format for index : [t e s t]" {
			t.Error(err)
		}
	} else {
		t.Error("was expecting an error")
	}

	if rcv != "" {
		t.Error(err)
	}
}

func TestRSRParserParseDataProviderWithInterfaces(t *testing.T) {
	prsr := RSRParser{
		path: "test",
	}
	dP := &FWVProvider{
		req:   "test",
		cache: utils.MapStorage{"test": 1},
	}

	rcv, err := prsr.ParseDataProviderWithInterfaces(dP, "")
	if err != nil {
		if err.Error() != "Invalid format for index : [t e s t]" {
			t.Error(err)
		}
	} else {
		t.Error("was expecting an error")
	}

	if rcv != "" {
		t.Error(err)
	}
}

func TestRSRParserParseDataProviderAsFloat642(t *testing.T) {
	prsr := RSRParser{
		path: "test",
	}
	dP := &FWVProvider{
		req:   "test",
		cache: utils.MapStorage{"test": 1},
	}

	rcv, err := prsr.ParseDataProviderAsFloat64(dP, "")
	if err != nil {
		if err.Error() != "Invalid format for index : [t e s t]" {
			t.Error(err)
		}
	} else {
		t.Error("was expecting an error")
	}

	if rcv != 0.0 {
		t.Error(err)
	}
}

func TestRSRParsersParseValue(t *testing.T) {
	str := "test"
	prsrs := RSRParsers{{
		Rules:           str,
		AllFiltersMatch: false,
		path:            str,
		rsrRules:        []*utils.ReSearchReplace{},
		converters:      utils.DataConverters{&utils.MultiplyConverter{Value: 1.2}},
		filters:         utils.RSRFilters{},
	}}

	rcv, err := prsrs.ParseValue("test)")
	if err != nil {
		if err.Error() != `strconv.ParseFloat: parsing "test)": invalid syntax` {
			t.Error(err)
		}
	} else {
		t.Error("was expecting an error")
	}

	if rcv != "" {
		t.Error(rcv)
	}
}

func TestRSRParsersParseDataProvider(t *testing.T) {
	str := "test"
	prsrs := RSRParsers{{
		Rules:           str,
		AllFiltersMatch: false,
		path:            str,
		rsrRules:        []*utils.ReSearchReplace{},
		converters:      utils.DataConverters{&utils.MultiplyConverter{Value: 1.2}},
		filters:         utils.RSRFilters{},
	}}
	dP := &FWVProvider{
		req:   "test",
		cache: utils.MapStorage{"test": 1},
	}

	rcv, err := prsrs.ParseDataProvider(dP, "")
	if err != nil {
		if err.Error() != `Invalid format for index : [t e s t]` {
			t.Error(err)
		}
	} else {
		t.Error("was expecting an error")
	}

	if rcv != "" {
		t.Error(rcv)
	}
}

func TestRSRParsersParseDataProviderWithInterfaces(t *testing.T) {
	str := "test"
	prsrs := RSRParsers{{
		Rules:           str,
		AllFiltersMatch: false,
		path:            str,
		rsrRules:        []*utils.ReSearchReplace{},
		converters:      utils.DataConverters{&utils.MultiplyConverter{Value: 1.2}},
		filters:         utils.RSRFilters{},
	}}
	dP := &FWVProvider{
		req:   "test",
		cache: utils.MapStorage{"test": 1},
	}

	rcv, err := prsrs.ParseDataProviderWithInterfaces(dP, "")
	if err != nil {
		if err.Error() != `Invalid format for index : [t e s t]` {
			t.Error(err)
		}
	} else {
		t.Error("was expecting an error")
	}

	if rcv != "" {
		t.Error(rcv)
	}
}

func TestRSRParsersCompile2(t *testing.T) {
	str := "test)"
	prsrs := RSRParsers{{
		Rules:           str,
		AllFiltersMatch: false,
		path:            str,
		rsrRules:        []*utils.ReSearchReplace{},
		converters:      utils.DataConverters{&utils.MultiplyConverter{Value: 1.2}},
		filters:         utils.RSRFilters{},
	}}

	err := prsrs.Compile()

	if err != nil {
		if err.Error() != "invalid RSRFilter start rule in string: <test)>" {
			t.Error(err)
		}
	} else {
		t.Error("was expecting an error")
	}
}

func TestRSRParserparseValue(t *testing.T) {
	str := "test)"
	prsr := RSRParser{
		Rules:           str,
		AllFiltersMatch: false,
		path:            str,
		rsrRules: []*utils.ReSearchReplace{{
			ReplaceTemplate: "test",
			Matched:         true,
		},
		}}

	rcv := prsr.RegexpMatched()

	if rcv != true {
		t.Error(rcv)
	}
}

func TestRSRParsersNewRSRParsersFromSlice(t *testing.T) {
	rcv, err := NewRSRParsersFromSlice([]string{"test)"}, false)

	if err != nil {
		if err.Error() != "invalid RSRFilter start rule in string: <test)>" {
			t.Error(err)
		}
	}

	if rcv != nil {
		t.Error(rcv)
	}

	rcv, err = NewRSRParsersFromSlice([]string{""}, false)

	if err != nil {
		if err.Error() != "empty RSRParser in rule: <>" {
			t.Error(err)
		}
	}

	if rcv != nil {
		t.Error(rcv)
	}
}

func TestRSRParsersparseValue(t *testing.T) {
	str := "test)"
	prsr := RSRParser{
		Rules:           str,
		AllFiltersMatch: false,
		path:            str,
		rsrRules: []*utils.ReSearchReplace{{
			ReplaceTemplate: "test",
			Matched:         true,
		}},
	}

	rcv := prsr.parseValue("test")

	if rcv != "" {
		t.Error(rcv)
	}
}

func TestRSRParsersNewRSRParsersMustCompile(t *testing.T) {
	defer func() {
		_ = recover()
	}()

	rcv := NewRSRParsersMustCompile("test`", false, "")
	t.Error("should have panicked", rcv)
}

func TestRSRParsersNewRSRParserMustCompile(t *testing.T) {
	defer func() {
		_ = recover()
	}()

	rcv := NewRSRParserMustCompile("test)", false)
	t.Error("should have panicked, received:", rcv)
}
