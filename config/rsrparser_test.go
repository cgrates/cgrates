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
	if rsrParsers, err := NewRSRParsers(ruleStr, utils.INFIELD_SEP); err != nil {
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
	rsrParsers, err := NewRSRParsers(rule, utils.INFIELD_SEP)
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
	rsrParsers, err := NewRSRParsers(rule, utils.INFIELD_SEP)
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
	if rsrParsers, err := NewRSRParsers(ruleStr, utils.INFIELD_SEP); err != nil {
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
	if rsrParsers, err := NewRSRParsers(ruleStr, utils.INFIELD_SEP); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if out, err := rsrParsers.ParseDataProvider(utils.MapStorage{}); err != nil {
		t.Error(err)
	} else if expected := "constantsomething>;q=0.7;expires=3600newconstant"; out != expected {
		t.Errorf("Expected %q ,received %q", expected, out)
	}

	ruleStr = "constant;`>;q=0.7;expires=3600`;constant"
	if rsrParsers, err := NewRSRParsers(ruleStr, utils.INFIELD_SEP); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if out, err := rsrParsers.ParseDataProvider(utils.MapStorage{}); err != nil {
		t.Error(err)
	} else if expected := "constant>;q=0.7;expires=3600constant"; out != expected {
		t.Errorf("Expected %q ,received %q", expected, out)
	}

	ruleStr = "constant;`>;q=0.7;expires=3600`constant"
	if rsrParsers, err := NewRSRParsers(ruleStr, utils.INFIELD_SEP); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if out, err := rsrParsers.ParseDataProvider(utils.MapStorage{}); err != nil {
		t.Error(err)
	} else if expected := "constant>;q=0.7;expires=3600constant"; out != expected {
		t.Errorf("Expected %q ,received %q", expected, out)
	}

	ruleStr = "constant;`>;q=0.7;expires=3600`;"
	if rsrParsers, err := NewRSRParsers(ruleStr, utils.INFIELD_SEP); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if out, err := rsrParsers.ParseDataProvider(utils.MapStorage{}); err != nil {
		t.Error(err)
	} else if expected := "constant>;q=0.7;expires=3600"; out != expected {
		t.Errorf("Expected %q ,received %q", expected, out)
	}

	ruleStr = "constant;`>;q=0.7;expires=3600constant"
	if _, err := NewRSRParsers(ruleStr, utils.INFIELD_SEP); err == nil {
		t.Error("Unexpected error: ", err.Error())
	}

	ruleStr = "constant;`>;q=0.7;expires=3600`;~*req.Account"
	if rsrParsers, err := NewRSRParsers(ruleStr, utils.INFIELD_SEP); err != nil {
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
	if rsrParsers, err := NewRSRParsers(ruleStr, utils.INFIELD_SEP); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if out, err := rsrParsers.ParseDataProviderWithInterfaces(
		utils.MapStorage{
			utils.MetaReq: utils.MapStorage{utils.Account: "1001"},
		}); err != nil {
		t.Error(err)
	} else if expected := "~*accounts.1001"; out != expected {
		t.Errorf("Expected %q ,received %q", expected, out)
	}

	ruleStr = "constant;`>;q=0.7;expires=3600`;~*req.Account"
	if rsrParsers, err := NewRSRParsers(ruleStr, utils.INFIELD_SEP); err != nil {
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
	if rsrParsers, err := NewRSRParsers(ruleStr, utils.INFIELD_SEP); err != nil {
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
	rsrParsers, err := NewRSRParsers("~*req.Account{*round}", utils.INFIELD_SEP)
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

func TestRSRParserRegexpMatched(t *testing.T) {
	rsr := NewRSRParserMustCompile("~*req.Time:s/(.*)/${1}s/")
	expected := "1ss"
	if val, err := rsr.parseValue("1s"); err != nil {
		t.Error(err)
	} else if val != expected {
		t.Errorf("Expected: %q received: %q", expected, val)
	}
	if !rsr.RegexpMatched() {
		t.Error("Expected the regex to match")
	}
	rsr = NewRSRParserMustCompile("~*req.Time:s/(a+)/${1}s/")
	expected = "1s"
	if val, err := rsr.parseValue("1s"); err != nil {
		t.Error(err)
	} else if val != expected {
		t.Errorf("Expected: %q received: %q", expected, val)
	}
	if rsr.RegexpMatched() {
		t.Error("Expected the regex to not match")
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
	if rply, err := NewRSRParsersMustCompile("*rated;~*req.Category", utils.INFIELD_SEP).GetIfaceFromValues(dp); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(exp, rply) {
		t.Errorf("Expecting %q, received: %q", exp, rply)
	}
	if _, err := NewRSRParsersMustCompile("*rated;~req.Category", utils.INFIELD_SEP).GetIfaceFromValues(utils.MapStorage{}); err != utils.ErrNotFound {
		t.Error(err)
	}
}
