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
	ruleStr := `Value1;Value2;~Header3;~Header4:s/a/${1}b/{*duration_seconds&*round:2};Value5{*duration_seconds&*round:2}`
	eRSRParsers := RSRParsers{
		&RSRParser{Rules: "Value1", Path: "Value1"},
		&RSRParser{Rules: "Value2", Path: "Value2"},
		&RSRParser{Rules: "~Header3", Path: "~Header3", rsrRules: make([]*ReSearchReplace, 0)},
		&RSRParser{Rules: "~Header4:s/a/${1}b/{*duration_seconds&*round:2}",
			Path: "~Header4",
			rsrRules: []*ReSearchReplace{{
				SearchRegexp:    regexp.MustCompile(`a`),
				ReplaceTemplate: "${1}b"}},
			converters: DataConverters{NewDataConverterMustCompile("*duration_seconds"),
				NewDataConverterMustCompile("*round:2")},
		},

		&RSRParser{Rules: "Value5{*duration_seconds&*round:2}",
			Path: "Value5",
			converters: DataConverters{NewDataConverterMustCompile("*duration_seconds"),
				NewDataConverterMustCompile("*round:2")},
		},
	}
	if rsrParsers, err := NewRSRParsers(ruleStr, InfieldSep); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if !reflect.DeepEqual(eRSRParsers, rsrParsers) {
		t.Errorf("expecting: %+v, received: %+v", eRSRParsers, rsrParsers)
	}
}

func TestRSRParserCompile(t *testing.T) {
	ePrsr := &RSRParser{
		Rules: "~Header4:s/a/${1}b/{*duration_seconds&*round:2}",
		Path:  "~Header4",
		rsrRules: []*ReSearchReplace{{
			SearchRegexp:    regexp.MustCompile(`a`),
			ReplaceTemplate: "${1}b"}},
		converters: DataConverters{NewDataConverterMustCompile("*duration_seconds"),
			NewDataConverterMustCompile("*round:2")},
	}
	prsr := &RSRParser{
		Rules: "~Header4:s/a/${1}b/{*duration_seconds&*round:2}",
	}
	if err := prsr.Compile(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(ePrsr, prsr) {
		t.Errorf("expecting: %+v, received: %+v", ePrsr, prsr)
	}

	prsr = &RSRParser{
		Rules: "~*req.Field{*}",
	}
	expErr := "invalid converter value in string: <*>, err: unsupported converter definition: <*>"
	if err := prsr.Compile(); err == nil || err.Error() != expErr {
		t.Fatal(err)
	}
}

func TestRSRParserConstant(t *testing.T) {
	rule := "cgrates.org"
	rsrParsers, err := NewRSRParsers(rule, InfieldSep)
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
	rsrParsers, err := NewRSRParsers(rule, InfieldSep)
	if err != nil {
		t.Error("Unexpected error: ", err.Error())
	}
	if out, err := rsrParsers.ParseValue(""); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if out != "" {
		t.Errorf("expecting: EmptyString , received: %+v", out)
	}
}

func TestNewRSRParsersConstant(t *testing.T) {
	ruleStr := "`>;q=0.7;expires=3600`"
	eRSRParsers := RSRParsers{
		&RSRParser{Rules: ">;q=0.7;expires=3600", Path: ">;q=0.7;expires=3600"},
	}
	if rsrParsers, err := NewRSRParsers(ruleStr, InfieldSep); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if !reflect.DeepEqual(eRSRParsers, rsrParsers) {
		t.Errorf("expecting: %+v, received: %+v", eRSRParsers, rsrParsers)
	} else if out, err := rsrParsers.ParseDataProvider(MapStorage{}); err != nil {
		t.Error(err)
	} else if expected := ">;q=0.7;expires=3600"; out != expected {
		t.Errorf("Expected %+v ,received %+v", expected, out)
	}
}

func TestNewRSRParsersConstant2(t *testing.T) {
	ruleStr := "constant;something`>;q=0.7;expires=3600`new;constant"
	if rsrParsers, err := NewRSRParsers(ruleStr, InfieldSep); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if out, err := rsrParsers.ParseDataProvider(MapStorage{}); err != nil {
		t.Error(err)
	} else if expected := "constantsomething>;q=0.7;expires=3600newconstant"; out != expected {
		t.Errorf("Expected %q ,received %q", expected, out)
	}

	ruleStr = "constant;`>;q=0.7;expires=3600`;constant"
	if rsrParsers, err := NewRSRParsers(ruleStr, InfieldSep); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if out, err := rsrParsers.ParseDataProvider(MapStorage{}); err != nil {
		t.Error(err)
	} else if expected := "constant>;q=0.7;expires=3600constant"; out != expected {
		t.Errorf("Expected %q ,received %q", expected, out)
	}

	ruleStr = "constant;`>;q=0.7;expires=3600`constant"
	if rsrParsers, err := NewRSRParsers(ruleStr, InfieldSep); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if out, err := rsrParsers.ParseDataProvider(MapStorage{}); err != nil {
		t.Error(err)
	} else if expected := "constant>;q=0.7;expires=3600constant"; out != expected {
		t.Errorf("Expected %q ,received %q", expected, out)
	}

	ruleStr = "constant;`>;q=0.7;expires=3600`;"
	if rsrParsers, err := NewRSRParsers(ruleStr, InfieldSep); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if out, err := rsrParsers.ParseDataProvider(MapStorage{}); err != nil {
		t.Error(err)
	} else if expected := "constant>;q=0.7;expires=3600"; out != expected {
		t.Errorf("Expected %q ,received %q", expected, out)
	}

	ruleStr = "constant;`>;q=0.7;expires=3600constant"
	if _, err := NewRSRParsers(ruleStr, InfieldSep); err == nil {
		t.Error("Unexpected error: ", err)
	}

	ruleStr = "constant;`>;q=0.7;expires=3600`;~*req.Account"
	if rsrParsers, err := NewRSRParsers(ruleStr, InfieldSep); err != nil {
		t.Error("Unexpected error: ", err)
	} else if _, err := rsrParsers.ParseDataProvider(MapStorage{}); err != ErrNotFound {
		t.Error(err)
	}

}

func TestRSRParserCompileConstant(t *testing.T) {
	ePrsr := &RSRParser{
		Rules: ":>;q=0.7;expires=3600",

		Path: ":>;q=0.7;expires=3600",
	}
	prsr := &RSRParser{
		Rules: ":>;q=0.7;expires=3600",
	}
	if err := prsr.Compile(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(ePrsr, prsr) {
		t.Errorf("expecting: %+v, received: %+v", ePrsr, prsr)
	}
}

func TestNewRSRParsersParseDataProviderWithInterfaces(t *testing.T) {
	ruleStr := "~;*accounts.;~*req.Account"
	if rsrParsers, err := NewRSRParsers(ruleStr, InfieldSep); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if out, err := rsrParsers.ParseDataProviderWithInterfaces(
		MapStorage{
			MetaReq: MapStorage{AccountField: "1001"},
		}); err != nil {
		t.Error(err)
	} else if expected := "~*accounts.1001"; out != expected {
		t.Errorf("Expected %q ,received %q", expected, out)
	}

	ruleStr = "constant;`>;q=0.7;expires=3600`;~*req.Account"
	if rsrParsers, err := NewRSRParsers(ruleStr, InfieldSep); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if _, err := rsrParsers.ParseDataProviderWithInterfaces(MapStorage{}); err != ErrNotFound {
		t.Error(err)
	}
}

func TestNewRSRParsersFromSlice(t *testing.T) {
	if _, err := NewRSRParsersFromSlice([]string{""}); err == nil {
		t.Error("Unexpected error: ", err)
	}

	if _, err := NewRSRParsersFromSlice([]string{"~*req.Account{*"}); err == nil {
		t.Error("Unexpected error: ", err)
	}
}

func TestNewRSRParsersMustCompile(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic on wrong rule")
		}
	}()
	NewRSRParsersMustCompile("~*req.Account{*", ";")
}

func TestRSRParserGetRule(t *testing.T) {
	ruleStr := "constant;~*req.Account"
	if rsrParsers, err := NewRSRParsers(ruleStr, InfieldSep); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if rule := rsrParsers.GetRule(); rule != ruleStr {
		t.Errorf("Expected: %q received: %q", ruleStr, rule)
	}
}

func TestRSRParsersCompile(t *testing.T) {
	prsrs := RSRParsers{&RSRParser{
		Rules: ":>;q=0.7;expires=3600",
	}}
	ePrsr := RSRParsers{&RSRParser{
		Rules: ":>;q=0.7;expires=3600",

		Path: ":>;q=0.7;expires=3600",
	}}
	if err := prsrs.Compile(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(prsrs, ePrsr) {
		t.Errorf("Expected %+v received %+v", ePrsr, prsrs)
	}
	prsrs = RSRParsers{&RSRParser{
		Rules: "~*req.Account{*unuportedConverter}",
	}}
	if err := prsrs.Compile(); err == nil {
		t.Error("Expected error received:", err)
	}
}

func TestRSRParsersParseValue(t *testing.T) {
	rsrParsers, err := NewRSRParsers("~*req.Account{*round}", InfieldSep)
	if err != nil {
		t.Error("Unexpected error: ", err.Error())
	}
	if _, err = rsrParsers.ParseValue("A"); err == nil {
		t.Error("Expected error received:", err)
	}
}

func TestNewRSRParserMustCompile(t *testing.T) {
	rsr := NewRSRParserMustCompile("~*req.Account")
	ePrsr := &RSRParser{
		Rules:    "~*req.Account",
		rsrRules: make([]*ReSearchReplace, 0),
		Path:     "~*req.Account",
	}
	if !reflect.DeepEqual(rsr, ePrsr) {
		t.Errorf("Expected %+v received %+v", ePrsr, rsr)
	}
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic on wrong rule")
		}
	}()
	NewRSRParserMustCompile("~*req.Account{*")
}

func TestRSRParserAttrName(t *testing.T) {
	rsr := NewRSRParserMustCompile("~*req.Account")
	expected := "*req.Account"
	if attr := rsr.AttrName(); attr != expected {
		t.Errorf("Expected: %q received: %q", expected, attr)
	}
}

func TestRSRParserCompile3(t *testing.T) {
	rsr := &RSRParser{Rules: "~*req.Account:s/(a+)/${1}s"}
	if err := rsr.Compile(); err == nil {
		t.Error("Expected error received:", err)
	}

	rsr = &RSRParser{Rules: "~*req.Account:s/*/${1}s/"}
	if err := rsr.Compile(); err == nil {
		t.Error("Expected error received:", err)
	}
}
func TestRSRParserDynamic(t *testing.T) {
	ePrsr := &RSRParser{
		Rules:       "~*opts.<~*opts.*originID;~*req.RunID;-Cost>",
		dynRules:    NewRSRParsersMustCompile("~*opts.*originID;~*req.RunID;-Cost", ";"),
		dynIdxStart: 7,
		dynIdxEnd:   43,
	}
	prsr := &RSRParser{
		Rules: "~*opts.<~*opts.*originID;~*req.RunID;-Cost>",
	}
	if err := prsr.Compile(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(ePrsr, prsr) {
		t.Errorf("expecting: %+v, received: %+v", ePrsr, prsr)
	}

	dP := MapStorage{
		MetaReq: MapStorage{

			RunID: MetaDefault,
		},
		MetaOpts: MapStorage{
			MetaOriginID:        "Uniq",
			"Uniq*default-Cost": 10,
		},
	}
	if out, err := prsr.ParseDataProvider(dP); err != nil {
		t.Error(err)
	} else if out != "10" {
		t.Errorf("Expected 10 received: %q", out)
	}

	prsr = &RSRParser{
		Rules: "~*opts.<~*opts.*originID;~*req.RunID;-Cost{*}>",
	}
	expErr := "invalid converter value in string: <*>, err: unsupported converter definition: <*>"
	if err := prsr.Compile(); err == nil || err.Error() != expErr {
		t.Fatal(err)
	}

}

func TestRSRParserDynamic2(t *testing.T) {
	prsr, err := NewRSRParsersFromSlice([]string{"~*opts.<~*opts.*originID;~*req.RunID;-Cos>t", "s"})
	if err != nil {
		t.Fatal(err)
	}

	dP := MapStorage{
		MetaReq: MapStorage{

			RunID: MetaDefault,
		},
		MetaOpts: MapStorage{
			MetaOriginID:                "originIDUniq",
			"originIDUniq*default-Cost": 10,
		},
	}
	if out, err := prsr.ParseDataProvider(dP); err != nil {
		t.Error(err)
	} else if out != "10s" {
		t.Errorf("Expected 10s received: %q", out)
	}

	prsr, err = NewRSRParsersFromSlice([]string{"2.", "~*opts.<~*opts.*originID;~*req.RunID;-Cos>t", "s"})
	if err != nil {
		t.Fatal(err)
	}

	if out, err := prsr.ParseDataProvider(dP); err != nil {
		t.Error(err)
	} else if out != "2.10s" {
		t.Errorf("Expected 2.10s received: %q", out)
	}

	prsr, err = NewRSRParsersFromSlice([]string{"2.", "~*opts.<~*opts.*originID;~*req.RunID;-Cost>"})
	if err != nil {
		t.Fatal(err)
	}

	if out, err := prsr.ParseDataProvider(dP); err != nil {
		t.Error(err)
	} else if out != "2.10" {
		t.Errorf("Expected 2.10 received: %q", out)
	}

	dP = MapStorage{
		MetaReq: MapStorage{},
		MetaOpts: MapStorage{
			MetaOriginID: "originIDUniq",
		},
	}
	if _, err := prsr.ParseDataProvider(dP); err != ErrNotFound {
		t.Errorf("Expected error %s, received: %v", ErrNotFound, err)
	}

	prsr, err = NewRSRParsersFromSlice([]string{"2.", "~*opts.*originID<~*opts.Converter>"})
	if err != nil {
		t.Fatal(err)
	}
	dP = MapStorage{
		MetaReq: MapStorage{},
		MetaOpts: MapStorage{
			"Converter":  "{*",
			MetaOriginID: "originIDUniq",
		},
	}
	if _, err := prsr.ParseDataProvider(dP); err == nil {
		t.Error(err)
	}
}

func TestRSRParserDynamic3(t *testing.T) {
	prsr, err := NewRSRParsersFromSlice([]string{"2.", "~*opts.<~*opts.*originID;~*req.RunID>-Cost", "-", "~*req.<~*req.UnitField>"})
	if err != nil {
		t.Fatal(err)
	}

	dP := MapStorage{
		MetaReq: MapStorage{

			RunID: MetaDefault,

			"UnitField": "Unit",
			"Unit":      "MB",
			"IP":        "127.0.0.1",
		},
		MetaOpts: MapStorage{

			MetaOriginID:                "originIDUniq",
			"originIDUniq*default-Cost": 10,
		},
	}

	if out, err := prsr.ParseDataProvider(dP); err != nil {
		t.Error(err)
	} else if out != "2.10-MB" {
		t.Errorf("Expected 2.10-MB received: %q", out)
	}

}

func TestRSRParserParseDataProviderWithInterfaces(t *testing.T) {
	prsr, err := NewRSRParsersFromSlice([]string{"~*opts.<~*opts.*originID;~*req.RunID;-Cos>t", "s"})
	if err != nil {
		t.Fatal(err)
	}

	dP := MapStorage{
		MetaReq: MapStorage{

			RunID: MetaDefault,
		},
		MetaOpts: MapStorage{
			MetaOriginID:                "originIDUniq",
			"originIDUniq*default-Cost": 10,
		},
	}
	if out, err := prsr.ParseDataProviderWithInterfaces(dP); err != nil {
		t.Error(err)
	} else if out != "10s" {
		t.Errorf("Expected 10s received: %q", out)
	}

	prsr, err = NewRSRParsersFromSlice([]string{"2.", "~*opts.<~*opts.*originID;~*req.RunID;-Cos>t", "s"})
	if err != nil {
		t.Fatal(err)
	}

	if out, err := prsr.ParseDataProviderWithInterfaces(dP); err != nil {
		t.Error(err)
	} else if out != "210s" {
		t.Errorf("Expected 210s received: %q", out)
	}

	prsr, err = NewRSRParsersFromSlice([]string{"2.", "~*opts.<~*opts.*originID;~*req.RunID;-Cost>"})
	if err != nil {
		t.Fatal(err)
	}

	if out, err := prsr.ParseDataProviderWithInterfaces(dP); err != nil {
		t.Error(err)
	} else if out != "210" {
		t.Errorf("Expected 210 received: %q", out)
	}

	dP = MapStorage{
		MetaReq: MapStorage{},
		MetaOpts: MapStorage{
			MetaOriginID: "originIDUniq",
		},
	}
	if _, err := prsr.ParseDataProviderWithInterfaces(dP); err != ErrNotFound {
		t.Errorf("Expected error %s, received: %v", ErrNotFound, err)
	}

	prsr, err = NewRSRParsersFromSlice([]string{"2.", "~*opts.*originID<~*opts.Converter>"})
	if err != nil {
		t.Fatal(err)
	}
	dP = MapStorage{
		MetaReq: MapStorage{},
		MetaOpts: MapStorage{
			"Converter":  "{*",
			MetaOriginID: "originIDUniq",
		},
	}
	if _, err := prsr.ParseDataProviderWithInterfaces(dP); err == nil {
		t.Error(err)
	}
}

func TestRSRParserCompileDynRule(t *testing.T) {
	prsr, err := NewRSRParser("~*opts.<~*opts.*originID;~*req.RunID;-Cos>t")
	if err != nil {
		t.Fatal(err)
	}

	dP := MapStorage{
		MetaReq: MapStorage{

			RunID: MetaDefault,
		},
		MetaOpts: MapStorage{
			MetaOriginID:                "originIDUniq",
			"originIDUniq*default-Cost": 10,
		},
	}
	if out, err := prsr.CompileDynRule(dP); err != nil {
		t.Error(err)
	} else if out != "~*opts.originIDUniq*default-Cost" {
		t.Errorf("Expected ~*opts.originIDUniq*default-Cost received: %q", out)
	}

	dP = MapStorage{
		MetaReq: MapStorage{},
		MetaOpts: MapStorage{
			MetaOriginID: "originIDUniq",
		},
	}
	if _, err := prsr.CompileDynRule(dP); err != ErrNotFound {
		t.Errorf("Expected error %s, received: %v", ErrNotFound, err)
	}

	prsr, err = NewRSRParser("~*opts.*originID")
	if err != nil {
		t.Fatal(err)
	}

	if out, err := prsr.CompileDynRule(dP); err != nil {
		t.Error(err)
	} else if out != "~*opts.*originID" {
		t.Errorf("Expected ~*opts.*originID received: %q", out)
	}
}

func TestRSRParsersGetIfaceFromValues(t *testing.T) {
	dp := MapStorage{
		MetaReq: MapStorage{
			Category: "call",
		},
	}
	exp := []any{"*rated", "call"}
	if rply, err := NewRSRParsersMustCompile("*rated;~*req.Category", InfieldSep).GetIfaceFromValues(dp); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(exp, rply) {
		t.Errorf("Expecting %q, received: %q", exp, rply)
	}
	if _, err := NewRSRParsersMustCompile("*rated;~req.Category", InfieldSep).GetIfaceFromValues(MapStorage{}); err != ErrNotFound {
		t.Error(err)
	}
}

func TestNewRSRParser(t *testing.T) {
	// Normal case
	rulesStr := `~sip_redirected_to:s/sip:\+49(\d+)@/0$1/`
	expRSRField1 := &RSRParser{
		Path:  "~sip_redirected_to",
		Rules: rulesStr,
		rsrRules: []*ReSearchReplace{
			{
				SearchRegexp:    regexp.MustCompile(`sip:\+49(\d+)@`),
				ReplaceTemplate: "0$1",
			},
		},
		converters: nil,
	}
	if rsrField, err := NewRSRParser(rulesStr); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if !reflect.DeepEqual(expRSRField1, rsrField) {
		t.Errorf("Expecting: %+v, received: %+v",
			expRSRField1, rsrField)
	}

	// with dataConverters
	rulesStr = `~sip_redirected_to:s/sip:\+49(\d+)@/0$1/{*duration_seconds&*round:5:*middle}`
	expRSRField := &RSRParser{
		Path:  "~sip_redirected_to",
		Rules: rulesStr,
		rsrRules: []*ReSearchReplace{{
			SearchRegexp:    regexp.MustCompile(`sip:\+49(\d+)@`),
			ReplaceTemplate: "0$1",
		}},
		converters: []DataConverter{
			new(DurationSecondsConverter),
			&RoundConverter{Decimals: 5, Method: "*middle"},
		},
	}
	if rsrField, err := NewRSRParser(rulesStr); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if !reflect.DeepEqual(expRSRField, rsrField) {
		t.Errorf("Expecting: %+v, received: %+v", expRSRField, rsrField)
	}
	// One extra separator but escaped
	rulesStr = `~sip_redirected_to:s/sip:\+49(\d+)\/@/0$1/`
	expRSRField3 := &RSRParser{
		Path:  "~sip_redirected_to",
		Rules: rulesStr,
		rsrRules: []*ReSearchReplace{{
			SearchRegexp:    regexp.MustCompile(`sip:\+49(\d+)\/@`),
			ReplaceTemplate: "0$1",
		}},
	}
	if rsrField, err := NewRSRParser(rulesStr); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if !reflect.DeepEqual(expRSRField3, rsrField) {
		t.Errorf("Expecting: %v, received: %v", expRSRField3, rsrField)
	}

}

func TestNewRSRParserDDz(t *testing.T) {
	rulesStr := `~effective_caller_id_number:s/(\d+)/+$1/`
	expectRSRField := &RSRParser{
		Path:  "~effective_caller_id_number",
		Rules: rulesStr,
		rsrRules: []*ReSearchReplace{{
			SearchRegexp:    regexp.MustCompile(`(\d+)`),
			ReplaceTemplate: "+$1",
		}},
	}
	if rsrField, err := NewRSRParser(rulesStr); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rsrField, expectRSRField) {
		t.Errorf("Unexpected RSRField received: %v", rsrField)
	}
}

func TestNewRSRParserIvo(t *testing.T) {
	rulesStr := `~cost_details:s/MatchedDestId":".+_(\s\s\s\s\s)"/$1/`
	expectRSRField := &RSRParser{
		Path:  "~cost_details",
		Rules: rulesStr,
		rsrRules: []*ReSearchReplace{{
			SearchRegexp:    regexp.MustCompile(`MatchedDestId":".+_(\s\s\s\s\s)"`),
			ReplaceTemplate: "$1",
		}},
	}
	if rsrField, err := NewRSRParser(rulesStr); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rsrField, expectRSRField) {
		t.Errorf("Unexpected RSRField received: %v", rsrField)
	}
	if _, err := NewRSRParser(`~account:s/^[A-Za-z0-9]*[c|a]\d{4}$/S/:s/^[A-Za-z0-9]*n\d{4}$/C/:s/^\d{10}$//`); err != nil {
		t.Error(err)
	}
}

func TestConvertPlusNationalAnd00(t *testing.T) {
	rulesStr := `~effective_caller_id_number:s/\+49(\d+)/0$1/:s/\+(\d+)/00$1/`
	expectRSRField := &RSRParser{
		Path:  "~effective_caller_id_number",
		Rules: rulesStr,
		rsrRules: []*ReSearchReplace{
			{
				SearchRegexp:    regexp.MustCompile(`\+49(\d+)`),
				ReplaceTemplate: "0$1",
			},
			{
				SearchRegexp:    regexp.MustCompile(`\+(\d+)`),
				ReplaceTemplate: "00$1",
			},
		},
	}
	rsrField, err := NewRSRParser(rulesStr)
	if err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rsrField, expectRSRField) {
		t.Errorf("Expecting: %v, received: %v", expectRSRField, rsrField)
	}
	if parsedVal, err := rsrField.ParseValue("+4986517174963"); err != nil {
		t.Error(err)
	} else if parsedVal != "086517174963" {
		t.Errorf("Expecting: 086517174963, received: %s", parsedVal)
	}
	if parsedVal, err := rsrField.ParseValue("+3186517174963"); err != nil {
		t.Error(err)
	} else if parsedVal != "003186517174963" {
		t.Errorf("Expecting: 003186517174963, received: %s", parsedVal)
	}
}

func TestConvertDurToSecs(t *testing.T) {
	rulesStr := `~9:s/^(\d+)$/${1}s/`
	expectRSRField := &RSRParser{
		Path:  "~9",
		Rules: rulesStr,
		rsrRules: []*ReSearchReplace{{
			SearchRegexp:    regexp.MustCompile(`^(\d+)$`),
			ReplaceTemplate: "${1}s",
		}},
	}
	rsrField, err := NewRSRParser(rulesStr)
	if err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rsrField, expectRSRField) {
		t.Errorf("Expecting: %v, received: %v", expectRSRField, rsrField)
	}
	if parsedVal, err := rsrField.ParseValue("640113"); err != nil {
		t.Error(err)
	} else if parsedVal != "640113s" {
		t.Errorf("Expecting: 640113s, received: %s", parsedVal)
	}
}

func TestPrefix164(t *testing.T) {
	rulesStr := `~0:s/^([1-9]\d+)$/+$1/`
	expectRSRField := &RSRParser{
		Path:  "~0",
		Rules: rulesStr,
		rsrRules: []*ReSearchReplace{{
			SearchRegexp:    regexp.MustCompile(`^([1-9]\d+)$`),
			ReplaceTemplate: "+$1",
		}},
	}
	rsrField, err := NewRSRParser(rulesStr)
	if err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rsrField, expectRSRField) {
		t.Errorf("Expecting: %v, received: %v", expectRSRField, rsrField)
	}
	if parsedVal, err := rsrField.ParseValue("4986517174960"); err != nil {
		t.Error(err)
	} else if parsedVal != "+4986517174960" {
		t.Errorf("Expecting: +4986517174960, received: %s", parsedVal)
	}
}

func TestNewRSRParsers2(t *testing.T) {
	fieldsStr1 := `~account:s/^\w+[mpls]\d{6}$//;~subject:s/^0\d{9}$//;~mediation_runid:s/^default$/default/`
	rsrFld1, err := NewRSRParser(`~account:s/^\w+[mpls]\d{6}$//`)
	if err != nil {
		t.Fatal(err)
	}
	rsrFld2, err := NewRSRParser(`~subject:s/^0\d{9}$//`)
	if err != nil {
		t.Fatal(err)
	}
	rsrFld4, err := NewRSRParser(`~mediation_runid:s/^default$/default/`)
	if err != nil {
		t.Fatal(err)
	}
	eRSRFields := RSRParsers{rsrFld1, rsrFld2, rsrFld4}
	if rsrFlds, err := NewRSRParsers(fieldsStr1, InfieldSep); err != nil {
		t.Error("Unexpected error: ", err)
	} else if !reflect.DeepEqual(eRSRFields, rsrFlds) {
		t.Errorf("Expecting: %v, received: %v", eRSRFields, rsrFlds)
	}
	fields := `host,~sip_redirected_to:s/sip:\+49(\d+)@/0$1/,destination`
	expectParsedFields := RSRParsers{
		{
			Path:  "host",
			Rules: "host",
		},
		{
			Path:  "~sip_redirected_to",
			Rules: `~sip_redirected_to:s/sip:\+49(\d+)@/0$1/`,
			rsrRules: []*ReSearchReplace{{
				SearchRegexp:    regexp.MustCompile(`sip:\+49(\d+)@`),
				ReplaceTemplate: "0$1",
			}},
		},
		{
			Path:  "destination",
			Rules: "destination",
		},
	}
	if parsedFields, err := NewRSRParsers(fields, FieldsSep); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if !reflect.DeepEqual(parsedFields, expectParsedFields) {
		t.Errorf("Expected: %s ,received: %s ", ToJSON(expectParsedFields), ToJSON(parsedFields))
	}
}

func TestParseCdrcDn1(t *testing.T) {
	rl, err := NewRSRParser(`~1:s/^00(\d+)(?:[a-zA-Z].{3})*0*([1-9]\d+)$/+$1$2/:s/^\+49(18\d{2})$/+491400$1/`)
	if err != nil {
		t.Error("Unexpected error: ", err)
	}
	if parsed, err := rl.ParseValue("0049ABOC0630415354"); err != nil {
		t.Error(err)
	} else if parsed != "+49630415354" {
		t.Errorf("Expecting: +49630415354, received: %s", parsed)
	}
	if parsed2, err := rl.ParseValue("00491888"); err != nil {
		t.Error(err)
	} else if parsed2 != "+4914001888" {
		t.Errorf("Expecting: +4914001888, received: %s", parsed2)
	}
}

func TestRSRCostDetails(t *testing.T) {
	fieldsStr1 := `{"Category":"default_route","Tenant":"demo.cgrates.org","Subject":"voxbeam_premium","Account":"6335820713","Destination":"15143606781","ToR":"*voice","Cost":0.0007,"Timespans":[{"TimeStart":"2015-08-30T21:46:54Z","TimeEnd":"2015-08-30T21:47:06Z","Cost":0.00072,"RateInterval":{"Rating":{"ConnectFee":0,"RoundingMethod":"*middle","RoundingDecimals":5,"MaxCost":0,"MaxCostStrategy":"0","Rates":[{"GroupIntervalStart":0,"Value":0.0036,"RateIncrement":6000000000,"RateUnit":60000000000}]},"Weight":10},"DurationIndex":12000000000,"Increments":[{"Duration":6000000000,"Cost":0.00036,"BalanceInfo":{"UnitBalanceUuid":"","MoneyBalanceUuid":"40adda88-25d3-4009-b928-f39d61590439","AccountId":"*out:demo.cgrates.org:6335820713"},"BalanceRateInterval":null,"UnitInfo":null,"CompressFactor":2}],"MatchedSubject":"*out:demo.cgrates.org:default_route:voxbeam_premium","MatchedPrefix":"1514","MatchedDestId":"Canada","RatingPlanId":"RP_VOXBEAM_PREMIUM"}]}`
	rsrField, err := NewRSRParser(`~cost_details:s/"MatchedDestId":"(\w+)"/${1}/`)
	if err != nil {
		t.Error(err)
	}
	if parsedVal, err := rsrField.ParseValue(fieldsStr1); err != nil {
		t.Error(err)
	} else if parsedVal != "Canada" {
		t.Errorf("Expecting: Canada, received: %s", parsedVal)
	}
	fieldsStr2 := `{"Category":"call","Tenant":"sip.test.cgrates.org","Subject":"dan","Account":"dan","Destination":"+4986517174963","ToR":"*voice","Cost":0,"Timespans":[{"TimeStart":"2015-05-13T15:03:34+02:00","TimeEnd":"2015-05-13T15:03:38+02:00","Cost":0,"RateInterval":{"Rating":{"ConnectFee":0,"RoundingMethod":"*middle","RoundingDecimals":4,"MaxCost":0,"MaxCostStrategy":"","Rates":[{"GroupIntervalStart":0,"Value":0,"RateIncrement":1000000000,"RateUnit":60000000000}]},"Weight":10},"DurationIndex":4000000000,"Increments":[{"Duration":1000000000,"Cost":0,"BalanceInfo":{"Unit":null,"Monetary":null,"AccountID":""},"CompressFactor":4}],"RoundIncrement":null,"MatchedSubject":"*out:sip.test.cgrates.org:call:*any","MatchedPrefix":"+31800","MatchedDestId":"CST_49800_DE080","RatingPlanId":"ISC_V","CompressFactor":1}],"RatedUsage":4}`
	rsrField, err = NewRSRParser(`~CostDetails:s/"MatchedDestId":.*_(\w{5})/${1}/:s/"MatchedDestId":"INTERNAL"/ON010/`)
	if err != nil {
		t.Error(err)
	}
	eMatch := "DE080"
	if parsedVal, err := rsrField.ParseValue(fieldsStr2); err != nil {
		t.Error(err)
	} else if parsedVal != eMatch {
		t.Errorf("Expecting: <%s>, received: <%s>", eMatch, parsedVal)
	}
}

func TestRSRFldParse(t *testing.T) {
	// with dataConverters
	rulesStr := `~Usage:s/(\d+)/${1}ms/{*duration_seconds&*round:1:*middle}`
	rsrField, err := NewRSRParser(rulesStr)
	if err != nil {
		t.Fatal(err)
	}
	eOut := "2.2"
	if out, err := rsrField.ParseValue("2210"); err != nil {
		t.Error(err)
	} else if out != eOut {
		t.Errorf("expecting: %s, received: %s", eOut, out)
	}
	rulesStr = `~Usage:s/(\d+)/${1}ms/{*duration_seconds&*round}`
	if rsrField, err = NewRSRParser(rulesStr); err != nil {
		t.Error(err)
	}
	eOut = "2"
	if out, err := rsrField.ParseValue("2210"); err != nil {
		t.Error(err)
	} else if out != eOut {
		t.Errorf("expecting: %s, received: %s", eOut, out)
	}
	rulesStr = `~Usage{*duration_seconds}`
	rsrField, err = NewRSRParser(rulesStr)
	if err != nil {
		t.Error(err)
	}
	eOut = "10"
	if out, err := rsrField.ParseValue("10000000000"); err != nil {
		t.Error(err)
	} else if out != eOut {
		t.Errorf("expecting: %s, received: %s", eOut, out)
	}
}

func TestRSRParsersClone(t *testing.T) {
	rsrs, err := NewRSRParsers(`~subject:s/^0\d{9}$//{*duration}`, InfieldSep)
	if err != nil {
		t.Fatal(err)
	}
	cln := rsrs.Clone()
	if !reflect.DeepEqual(rsrs, cln) {
		t.Errorf("Expected: %+v\nReceived: %+v", ToJSON(rsrs), ToJSON(cln))
	}
	cln[0].converters[0] = NewDataConverterMustCompile(MetaIP2Hex)
	if reflect.DeepEqual(cln[0].converters[0], rsrs[0].converters[0]) {
		t.Errorf("Expected clone to not modify the cloned")
	}
}

func TestRSRParsersParseDataProviderWithInterfaces2(t *testing.T) {
	ruleStr := "~;*accounts.;~*req.Account"
	if prsrs, err := NewRSRParsers(ruleStr, InfieldSep); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if rcv, err := prsrs.ParseDataProviderWithInterfaces2(MapStorage{
		MetaReq: MapStorage{AccountField: "1001"},
	}); err != nil {
		t.Errorf("Expected error <nil>, Received error <%v>", err)

	} else if expected := "~*accounts.1001"; rcv != expected {
		t.Errorf("Expected %q ,received %q", expected, rcv)
	}
	ruleStr = "constant;`>;q=0.7;expires=3600`;~*req.Account"
	if prsrs, err := NewRSRParsers(ruleStr, InfieldSep); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if _, err := prsrs.ParseDataProviderWithInterfaces2(MapStorage{}); err != ErrNotFound {
		t.Errorf("Expected error <%v>, Received error <%v>", ErrNotFound, err)
	}

}

func TestRSRParserParseValueInterface(t *testing.T) {

	prsr := &RSRParser{
		rsrRules: []*ReSearchReplace{
			{
				SearchRegexp:    regexp.MustCompile(`a`),
				ReplaceTemplate: "${1}b",
			},
		},
	}

	if rcv, err := prsr.parseValueInterface(""); err != nil {
		t.Errorf("Expected error <nil>, Received error <%v>", err)

	} else if expected := ""; rcv != expected {
		t.Errorf("Expected %q ,received %q", expected, rcv)
	}
}

func TestRSRParserParseDataProviderWithInterfaces2(t *testing.T) {
	prsr := &RSRParser{
		dynRules: NewRSRParsersMustCompile("~*opts.*originID;~*req.RunID;-Cost", ";"),
	}

	if _, err := prsr.ParseDataProviderWithInterfaces2(MapStorage{
		MetaReq: MapStorage{AccountField: "1001"},
	}); err != ErrNotFound {
		t.Errorf("Expected error <%v>, Received error <%v>", ErrNotFound, err)

	}

	rsrParser, _ := NewRSRParsers("~*opts.*originI;;;D;~*req.RunID;-Cost", "")
	prsr = &RSRParser{
		Rules:    "~*opts.<~*opts.*originID;~*req.RunID;-Cost{*}>",
		dynRules: rsrParser,
	}
	expErr := "invalid converter value in string: <*>, err: unsupported converter definition: <*>"
	if _, err := prsr.ParseDataProviderWithInterfaces2(MapStorage{
		MetaReq: MapStorage{AccountField: "1001"},
	}); err.Error() != expErr {
		t.Errorf("Expected error <%v>, Received error <%v>", expErr, err.Error())

	}

	prsr = &RSRParser{
		dynRules: NewRSRParsersMustCompile("~*opts.*originID;~*req.RunID;-Cost", ";"),
	}
	dP := MapStorage{
		MetaReq: MapStorage{

			RunID: MetaDefault,
		},
		MetaOpts: MapStorage{
			MetaOriginID:        "Uniq",
			"Uniq*default-Cost": 10,
		},
	}
	exp := "Uniq*default-Cost"
	if rcv, err := prsr.ParseDataProviderWithInterfaces2(dP); err != nil {
		t.Error(err)
	} else if rcv != exp {
		t.Errorf("Expected <%v>, \nReceived <%v>", exp, rcv)
	}
}
