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
	ruleStr := `Value1;Value2;~Header3;~Header4:s/a/${1}b/{*duration_seconds&*round:2};Value5{*duration_seconds&*round:2}`
	eRSRParsers := RSRParsers{
		&RSRParser{Rules: "Value1", path: "Value1"},
		&RSRParser{Rules: "Value2", path: "Value2"},
		&RSRParser{Rules: "~Header3", path: "~Header3", rsrRules: make([]*utils.ReSearchReplace, 0)},
		&RSRParser{Rules: "~Header4:s/a/${1}b/{*duration_seconds&*round:2}",
			path: "~Header4",
			rsrRules: []*utils.ReSearchReplace{{
				SearchRegexp:    regexp.MustCompile(`a`),
				ReplaceTemplate: "${1}b"}},
			converters: utils.DataConverters{utils.NewDataConverterMustCompile("*duration_seconds"),
				utils.NewDataConverterMustCompile("*round:2")},
		},

		&RSRParser{Rules: "Value5{*duration_seconds&*round:2}",
			path: "Value5",
			converters: utils.DataConverters{utils.NewDataConverterMustCompile("*duration_seconds"),
				utils.NewDataConverterMustCompile("*round:2")},
		},
	}
	if rsrParsers, err := NewRSRParsers(ruleStr, utils.InfieldSep); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if !reflect.DeepEqual(eRSRParsers, rsrParsers) {
		t.Errorf("expecting: %+v, received: %+v", eRSRParsers, rsrParsers)
	}
}

func TestRSRParserCompile(t *testing.T) {
	ePrsr := &RSRParser{
		Rules: "~Header4:s/a/${1}b/{*duration_seconds&*round:2}",
		path:  "~Header4",
		rsrRules: []*utils.ReSearchReplace{{
			SearchRegexp:    regexp.MustCompile(`a`),
			ReplaceTemplate: "${1}b"}},
		converters: utils.DataConverters{utils.NewDataConverterMustCompile("*duration_seconds"),
			utils.NewDataConverterMustCompile("*round:2")},
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
	rsrParsers, err := NewRSRParsers(rule, utils.InfieldSep)
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
	rsrParsers, err := NewRSRParsers(rule, utils.InfieldSep)
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
		&RSRParser{Rules: ">;q=0.7;expires=3600", path: ">;q=0.7;expires=3600"},
	}
	if rsrParsers, err := NewRSRParsers(ruleStr, utils.InfieldSep); err != nil {
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
	if rsrParsers, err := NewRSRParsers(ruleStr, utils.InfieldSep); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if out, err := rsrParsers.ParseDataProvider(utils.MapStorage{}); err != nil {
		t.Error(err)
	} else if expected := "constantsomething>;q=0.7;expires=3600newconstant"; out != expected {
		t.Errorf("Expected %q ,received %q", expected, out)
	}

	ruleStr = "constant;`>;q=0.7;expires=3600`;constant"
	if rsrParsers, err := NewRSRParsers(ruleStr, utils.InfieldSep); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if out, err := rsrParsers.ParseDataProvider(utils.MapStorage{}); err != nil {
		t.Error(err)
	} else if expected := "constant>;q=0.7;expires=3600constant"; out != expected {
		t.Errorf("Expected %q ,received %q", expected, out)
	}

	ruleStr = "constant;`>;q=0.7;expires=3600`constant"
	if rsrParsers, err := NewRSRParsers(ruleStr, utils.InfieldSep); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if out, err := rsrParsers.ParseDataProvider(utils.MapStorage{}); err != nil {
		t.Error(err)
	} else if expected := "constant>;q=0.7;expires=3600constant"; out != expected {
		t.Errorf("Expected %q ,received %q", expected, out)
	}

	ruleStr = "constant;`>;q=0.7;expires=3600`;"
	if rsrParsers, err := NewRSRParsers(ruleStr, utils.InfieldSep); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if out, err := rsrParsers.ParseDataProvider(utils.MapStorage{}); err != nil {
		t.Error(err)
	} else if expected := "constant>;q=0.7;expires=3600"; out != expected {
		t.Errorf("Expected %q ,received %q", expected, out)
	}

	ruleStr = "constant;`>;q=0.7;expires=3600constant"
	if _, err := NewRSRParsers(ruleStr, utils.InfieldSep); err == nil {
		t.Error("Unexpected error: ", err.Error())
	}

	ruleStr = "constant;`>;q=0.7;expires=3600`;~*req.Account"
	if rsrParsers, err := NewRSRParsers(ruleStr, utils.InfieldSep); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if _, err := rsrParsers.ParseDataProvider(utils.MapStorage{}); err != utils.ErrNotFound {
		t.Error(err)
	}

}

func TestRSRParserCompileConstant(t *testing.T) {
	ePrsr := &RSRParser{
		Rules: ":>;q=0.7;expires=3600",

		path: ":>;q=0.7;expires=3600",
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
	if rsrParsers, err := NewRSRParsers(ruleStr, utils.InfieldSep); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if out, err := rsrParsers.ParseDataProviderWithInterfaces(
		utils.MapStorage{
			utils.MetaReq: utils.MapStorage{utils.AccountField: "1001"},
		}); err != nil {
		t.Error(err)
	} else if expected := "~*accounts.1001"; out != expected {
		t.Errorf("Expected %q ,received %q", expected, out)
	}

	ruleStr = "constant;`>;q=0.7;expires=3600`;~*req.Account"
	if rsrParsers, err := NewRSRParsers(ruleStr, utils.InfieldSep); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if _, err := rsrParsers.ParseDataProviderWithInterfaces(utils.MapStorage{}); err != utils.ErrNotFound {
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
	if rsrParsers, err := NewRSRParsers(ruleStr, utils.InfieldSep); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if rule := rsrParsers.GetRule(utils.InfieldSep); rule != ruleStr {
		t.Errorf("Expected: %q received: %q", ruleStr, rule)
	}
}

func TestRSRParsersCompile(t *testing.T) {
	prsrs := RSRParsers{&RSRParser{
		Rules: ":>;q=0.7;expires=3600",
	}}
	ePrsr := RSRParsers{&RSRParser{
		Rules: ":>;q=0.7;expires=3600",

		path: ":>;q=0.7;expires=3600",
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
	rsrParsers, err := NewRSRParsers("~*req.Account{*round}", utils.InfieldSep)
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
		rsrRules: make([]*utils.ReSearchReplace, 0),
		path:     "~*req.Account",
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
		Rules: "~*req.<~*req.CGRID;~*req.RunID;-Cost>",

		dynRules:    NewRSRParsersMustCompile("~*req.CGRID;~*req.RunID;-Cost", ";"),
		dynIdxStart: 6,
		dynIdxEnd:   37,
	}
	prsr := &RSRParser{
		Rules: "~*req.<~*req.CGRID;~*req.RunID;-Cost>",
	}
	if err := prsr.Compile(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(ePrsr, prsr) {
		t.Errorf("expecting: %+v, received: %+v", ePrsr, prsr)
	}

	dP := utils.MapStorage{
		utils.MetaReq: utils.MapStorage{
			utils.CGRID:              "cgridUniq",
			utils.RunID:              utils.MetaDefault,
			"cgridUniq*default-Cost": 10,
		},
	}
	if out, err := prsr.ParseDataProvider(dP); err != nil {
		t.Error(err)
	} else if out != "10" {
		t.Errorf("Expected 10 received: %q", out)
	}

	prsr = &RSRParser{
		Rules: "~*req.<~*req.CGRID;~*req.RunID;-Cost{*}>",
	}
	expErr := "invalid converter value in string: <*>, err: unsupported converter definition: <*>"
	if err := prsr.Compile(); err == nil || err.Error() != expErr {
		t.Fatal(err)
	}

}

func TestRSRParserDynamic2(t *testing.T) {
	prsr, err := NewRSRParsersFromSlice([]string{"~*req.<~*req.CGRID;~*req.RunID;-Cos>t", "s"})
	if err != nil {
		t.Fatal(err)
	}

	dP := utils.MapStorage{
		utils.MetaReq: utils.MapStorage{
			utils.CGRID:              "cgridUniq",
			utils.RunID:              utils.MetaDefault,
			"cgridUniq*default-Cost": 10,
		},
	}
	if out, err := prsr.ParseDataProvider(dP); err != nil {
		t.Error(err)
	} else if out != "10s" {
		t.Errorf("Expected 10s received: %q", out)
	}

	prsr, err = NewRSRParsersFromSlice([]string{"2.", "~*req.<~*req.CGRID;~*req.RunID;-Cos>t", "s"})
	if err != nil {
		t.Fatal(err)
	}

	if out, err := prsr.ParseDataProvider(dP); err != nil {
		t.Error(err)
	} else if out != "2.10s" {
		t.Errorf("Expected 2.10s received: %q", out)
	}

	prsr, err = NewRSRParsersFromSlice([]string{"2.", "~*req.<~*req.CGRID;~*req.RunID;-Cost>"})
	if err != nil {
		t.Fatal(err)
	}

	if out, err := prsr.ParseDataProvider(dP); err != nil {
		t.Error(err)
	} else if out != "2.10" {
		t.Errorf("Expected 2.10 received: %q", out)
	}

	dP = utils.MapStorage{
		utils.MetaReq: utils.MapStorage{
			utils.CGRID: "cgridUniq",
		},
	}
	if _, err := prsr.ParseDataProvider(dP); err != utils.ErrNotFound {
		t.Errorf("Expected error %s, received: %v", utils.ErrNotFound, err)
	}

	prsr, err = NewRSRParsersFromSlice([]string{"2.", "~*req.CGRID<~*opts.Converter>"})
	if err != nil {
		t.Fatal(err)
	}
	dP = utils.MapStorage{
		utils.MetaReq: utils.MapStorage{
			utils.CGRID: "cgridUniq",
		},
		utils.MetaOpts: utils.MapStorage{
			"Converter": "{*",
		},
	}
	if _, err := prsr.ParseDataProvider(dP); err == nil {
		t.Error(err)
	}
}

func TestRSRParserDynamic3(t *testing.T) {
	prsr, err := NewRSRParsersFromSlice([]string{"2.", "~*req.<~*req.CGRID;~*req.RunID>-Cost", "-", "~*req.<~*req.UnitField>"})
	if err != nil {
		t.Fatal(err)
	}

	dP := utils.MapStorage{
		utils.MetaReq: utils.MapStorage{
			utils.CGRID:              "cgridUniq",
			utils.RunID:              utils.MetaDefault,
			"cgridUniq*default-Cost": 10,
			"UnitField":              "Unit",
			"Unit":                   "MB",
			"IP":                     "127.0.0.1",
		},
	}

	if out, err := prsr.ParseDataProvider(dP); err != nil {
		t.Error(err)
	} else if out != "2.10-MB" {
		t.Errorf("Expected 2.10-MB received: %q", out)
	}

}

func TestRSRParserParseDataProviderWithInterfaces(t *testing.T) {
	prsr, err := NewRSRParsersFromSlice([]string{"~*req.<~*req.CGRID;~*req.RunID;-Cos>t", "s"})
	if err != nil {
		t.Fatal(err)
	}

	dP := utils.MapStorage{
		utils.MetaReq: utils.MapStorage{
			utils.CGRID:              "cgridUniq",
			utils.RunID:              utils.MetaDefault,
			"cgridUniq*default-Cost": 10,
		},
	}
	if out, err := prsr.ParseDataProviderWithInterfaces(dP); err != nil {
		t.Error(err)
	} else if out != "10s" {
		t.Errorf("Expected 10s received: %q", out)
	}

	prsr, err = NewRSRParsersFromSlice([]string{"2.", "~*req.<~*req.CGRID;~*req.RunID;-Cos>t", "s"})
	if err != nil {
		t.Fatal(err)
	}

	if out, err := prsr.ParseDataProviderWithInterfaces(dP); err != nil {
		t.Error(err)
	} else if out != "210s" {
		t.Errorf("Expected 210s received: %q", out)
	}

	prsr, err = NewRSRParsersFromSlice([]string{"2.", "~*req.<~*req.CGRID;~*req.RunID;-Cost>"})
	if err != nil {
		t.Fatal(err)
	}

	if out, err := prsr.ParseDataProviderWithInterfaces(dP); err != nil {
		t.Error(err)
	} else if out != "210" {
		t.Errorf("Expected 210 received: %q", out)
	}

	dP = utils.MapStorage{
		utils.MetaReq: utils.MapStorage{
			utils.CGRID: "cgridUniq",
		},
	}
	if _, err := prsr.ParseDataProviderWithInterfaces(dP); err != utils.ErrNotFound {
		t.Errorf("Expected error %s, received: %v", utils.ErrNotFound, err)
	}

	prsr, err = NewRSRParsersFromSlice([]string{"2.", "~*req.CGRID<~*opts.Converter>"})
	if err != nil {
		t.Fatal(err)
	}
	dP = utils.MapStorage{
		utils.MetaReq: utils.MapStorage{
			utils.CGRID: "cgridUniq",
		},
		utils.MetaOpts: utils.MapStorage{
			"Converter": "{*",
		},
	}
	if _, err := prsr.ParseDataProviderWithInterfaces(dP); err == nil {
		t.Error(err)
	}
}

func TestRSRParserCompileDynRule(t *testing.T) {
	prsr, err := NewRSRParser("~*req.<~*req.CGRID;~*req.RunID;-Cos>t")
	if err != nil {
		t.Fatal(err)
	}

	dP := utils.MapStorage{
		utils.MetaReq: utils.MapStorage{
			utils.CGRID:              "cgridUniq",
			utils.RunID:              utils.MetaDefault,
			"cgridUniq*default-Cost": 10,
		},
	}
	if out, err := prsr.CompileDynRule(dP); err != nil {
		t.Error(err)
	} else if out != "~*req.cgridUniq*default-Cost" {
		t.Errorf("Expected ~*req.cgridUniq*default-Cost received: %q", out)
	}

	dP = utils.MapStorage{
		utils.MetaReq: utils.MapStorage{
			utils.CGRID: "cgridUniq",
		},
	}
	if _, err := prsr.CompileDynRule(dP); err != utils.ErrNotFound {
		t.Errorf("Expected error %s, received: %v", utils.ErrNotFound, err)
	}

	prsr, err = NewRSRParser("~*req.CGRID")
	if err != nil {
		t.Fatal(err)
	}

	if out, err := prsr.CompileDynRule(dP); err != nil {
		t.Error(err)
	} else if out != "~*req.CGRID" {
		t.Errorf("Expected ~*req.CGRID received: %q", out)
	}
}

func TestRSRParsersGetIfaceFromValues(t *testing.T) {
	dp := utils.MapStorage{
		utils.MetaReq: utils.MapStorage{
			utils.Category: "call",
		},
	}
	exp := []interface{}{"*rated", "call"}
	if rply, err := NewRSRParsersMustCompile("*rated;~*req.Category", utils.InfieldSep).GetIfaceFromValues(dp); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(exp, rply) {
		t.Errorf("Expecting %q, received: %q", exp, rply)
	}
	if _, err := NewRSRParsersMustCompile("*rated;~req.Category", utils.InfieldSep).GetIfaceFromValues(utils.MapStorage{}); err != utils.ErrNotFound {
		t.Error(err)
	}
}

func TestNewRSRParser(t *testing.T) {
	// Normal case
	rulesStr := `~sip_redirected_to:s/sip:\+49(\d+)@/0$1/`
	expRSRField1 := &RSRParser{
		path:  "~sip_redirected_to",
		Rules: rulesStr,
		rsrRules: []*utils.ReSearchReplace{
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
		path:  "~sip_redirected_to",
		Rules: rulesStr,
		rsrRules: []*utils.ReSearchReplace{{
			SearchRegexp:    regexp.MustCompile(`sip:\+49(\d+)@`),
			ReplaceTemplate: "0$1",
		}},
		converters: []utils.DataConverter{
			new(utils.DurationSecondsConverter),
			&utils.RoundConverter{Decimals: 5, Method: "*middle"},
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
		path:  "~sip_redirected_to",
		Rules: rulesStr,
		rsrRules: []*utils.ReSearchReplace{{
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
		path:  "~effective_caller_id_number",
		Rules: rulesStr,
		rsrRules: []*utils.ReSearchReplace{{
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
		path:  "~cost_details",
		Rules: rulesStr,
		rsrRules: []*utils.ReSearchReplace{{
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
		path:  "~effective_caller_id_number",
		Rules: rulesStr,
		rsrRules: []*utils.ReSearchReplace{
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
		path:  "~9",
		Rules: rulesStr,
		rsrRules: []*utils.ReSearchReplace{{
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
		path:  "~0",
		Rules: rulesStr,
		rsrRules: []*utils.ReSearchReplace{{
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
	if rsrFlds, err := NewRSRParsers(fieldsStr1, utils.InfieldSep); err != nil {
		t.Error("Unexpected error: ", err)
	} else if !reflect.DeepEqual(eRSRFields, rsrFlds) {
		t.Errorf("Expecting: %v, received: %v", eRSRFields, rsrFlds)
	}
	fields := `host,~sip_redirected_to:s/sip:\+49(\d+)@/0$1/,destination`
	expectParsedFields := RSRParsers{
		{
			path:  "host",
			Rules: "host",
		},
		{
			path:  "~sip_redirected_to",
			Rules: `~sip_redirected_to:s/sip:\+49(\d+)@/0$1/`,
			rsrRules: []*utils.ReSearchReplace{{
				SearchRegexp:    regexp.MustCompile(`sip:\+49(\d+)@`),
				ReplaceTemplate: "0$1",
			}},
		},
		{
			path:  "destination",
			Rules: "destination",
		},
	}
	if parsedFields, err := NewRSRParsers(fields, utils.FIELDS_SEP); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if !reflect.DeepEqual(parsedFields, expectParsedFields) {
		t.Errorf("Expected: %s ,received: %s ", utils.ToJSON(expectParsedFields), utils.ToJSON(parsedFields))
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
	fieldsStr1 := `{"Category":"default_route","Tenant":"demo.cgrates.org","Subject":"voxbeam_premium","Account":"6335820713","Destination":"15143606781","ToR":"*voice","Cost":0.0007,"Timespans":[{"TimeStart":"2015-08-30T21:46:54Z","TimeEnd":"2015-08-30T21:47:06Z","Cost":0.00072,"RateInterval":{"Timing":{"Years":[],"Months":[],"MonthDays":[],"WeekDays":[],"StartTime":"00:00:00","EndTime":""},"Rating":{"ConnectFee":0,"RoundingMethod":"*middle","RoundingDecimals":5,"MaxCost":0,"MaxCostStrategy":"0","Rates":[{"GroupIntervalStart":0,"Value":0.0036,"RateIncrement":6000000000,"RateUnit":60000000000}]},"Weight":10},"DurationIndex":12000000000,"Increments":[{"Duration":6000000000,"Cost":0.00036,"BalanceInfo":{"UnitBalanceUuid":"","MoneyBalanceUuid":"40adda88-25d3-4009-b928-f39d61590439","AccountId":"*out:demo.cgrates.org:6335820713"},"BalanceRateInterval":null,"UnitInfo":null,"CompressFactor":2}],"MatchedSubject":"*out:demo.cgrates.org:default_route:voxbeam_premium","MatchedPrefix":"1514","MatchedDestId":"Canada","RatingPlanId":"RP_VOXBEAM_PREMIUM"}]}`
	rsrField, err := NewRSRParser(`~cost_details:s/"MatchedDestId":"(\w+)"/${1}/`)
	if err != nil {
		t.Error(err)
	}
	if parsedVal, err := rsrField.ParseValue(fieldsStr1); err != nil {
		t.Error(err)
	} else if parsedVal != "Canada" {
		t.Errorf("Expecting: Canada, received: %s", parsedVal)
	}
	fieldsStr2 := `{"Category":"call","Tenant":"sip.test.cgrates.org","Subject":"dan","Account":"dan","Destination":"+4986517174963","ToR":"*voice","Cost":0,"Timespans":[{"TimeStart":"2015-05-13T15:03:34+02:00","TimeEnd":"2015-05-13T15:03:38+02:00","Cost":0,"RateInterval":{"Timing":{"Years":[],"Months":[],"MonthDays":[],"WeekDays":[],"StartTime":"00:00:00","EndTime":""},"Rating":{"ConnectFee":0,"RoundingMethod":"*middle","RoundingDecimals":4,"MaxCost":0,"MaxCostStrategy":"","Rates":[{"GroupIntervalStart":0,"Value":0,"RateIncrement":1000000000,"RateUnit":60000000000}]},"Weight":10},"DurationIndex":4000000000,"Increments":[{"Duration":1000000000,"Cost":0,"BalanceInfo":{"Unit":null,"Monetary":null,"AccountID":""},"CompressFactor":4}],"RoundIncrement":null,"MatchedSubject":"*out:sip.test.cgrates.org:call:*any","MatchedPrefix":"+31800","MatchedDestId":"CST_49800_DE080","RatingPlanId":"ISC_V","CompressFactor":1}],"RatedUsage":4}`
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
	rsrs, err := NewRSRParsers(`~subject:s/^0\d{9}$//{*duration}`, utils.InfieldSep)
	if err != nil {
		t.Fatal(err)
	}
	cln := rsrs.Clone()
	if !reflect.DeepEqual(rsrs, cln) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(rsrs), utils.ToJSON(cln))
	}
	cln[0].converters[0] = utils.NewDataConverterMustCompile(utils.MetaIP2Hex)
	if reflect.DeepEqual(cln[0].converters[0], rsrs[0].converters[0]) {
		t.Errorf("Expected clone to not modify the cloned")
	}
}
