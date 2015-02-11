/*
Real-time Charging System for Telecom & ISP environments
Copyright (C) 2012-2014 ITsysCOM GmbH

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

func TestNewRSRField1(t *testing.T) {
	// Normal case
	expRSRField1 := &RSRField{Id: "sip_redirected_to",
		RSRules: []*ReSearchReplace{&ReSearchReplace{SearchRegexp: regexp.MustCompile(`sip:\+49(\d+)@`), ReplaceTemplate: "0$1"}}}
	if rsrField, err := NewRSRField(`~sip_redirected_to:s/sip:\+49(\d+)@/0$1/`); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if !reflect.DeepEqual(expRSRField1, rsrField) {
		t.Errorf("Expecting: %v, received: %v", expRSRField1, rsrField)
	}
	// With filter
	expRSRField2 := &RSRField{Id: "sip_redirected_to", filterValue: "086517174963",
		RSRules: []*ReSearchReplace{&ReSearchReplace{SearchRegexp: regexp.MustCompile(`sip:\+49(\d+)@`), ReplaceTemplate: "0$1"}}}
	if rsrField, err := NewRSRField(`~sip_redirected_to:s/sip:\+49(\d+)@/0$1/(086517174963)`); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if !reflect.DeepEqual(expRSRField2, rsrField) {
		t.Errorf("Expecting: %v, received: %v", expRSRField2, rsrField)
	}
	// Separator escaped
	if rsrField, err := NewRSRField(`~sip_redirected_to:s\/sip:\+49(\d+)@/0$1/`); err == nil {
		t.Errorf("Parse error, field rule does not contain correct number of separators, received: %v", rsrField)
	}
	// One extra separator but escaped
	expRSRField3 := &RSRField{Id: "sip_redirected_to",
		RSRules: []*ReSearchReplace{&ReSearchReplace{SearchRegexp: regexp.MustCompile(`sip:\+49(\d+)\/@`), ReplaceTemplate: "0$1"}}}
	if rsrField, err := NewRSRField(`~sip_redirected_to:s/sip:\+49(\d+)\/@/0$1/`); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if !reflect.DeepEqual(expRSRField3, rsrField) {
		t.Errorf("Expecting: %v, received: %v", expRSRField3, rsrField)
	}
}

func TestNewRSRFieldDDz(t *testing.T) {
	expectRSRField := &RSRField{Id: "effective_caller_id_number",
		RSRules: []*ReSearchReplace{&ReSearchReplace{SearchRegexp: regexp.MustCompile(`(\d+)`), ReplaceTemplate: "+$1"}}}
	if rsrField, err := NewRSRField(`~effective_caller_id_number:s/(\d+)/+$1/`); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rsrField, expectRSRField) {
		t.Errorf("Unexpected RSRField received: %v", rsrField)
	}
}

func TestNewRSRFieldIvo(t *testing.T) {
	expectRSRField := &RSRField{Id: "cost_details",
		RSRules: []*ReSearchReplace{&ReSearchReplace{SearchRegexp: regexp.MustCompile(`MatchedDestId":".+_(\s\s\s\s\s)"`), ReplaceTemplate: "$1"}}}
	if rsrField, err := NewRSRField(`~cost_details:s/MatchedDestId":".+_(\s\s\s\s\s)"/$1/`); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rsrField, expectRSRField) {
		t.Errorf("Unexpected RSRField received: %v", rsrField)
	}
	if _, err := NewRSRField(`~account:s/^[A-Za-z0-9]*[c|a]\d{4}$/S/:s/^[A-Za-z0-9]*n\d{4}$/C/:s/^\d{10}$//`); err != nil {
		t.Error(err)
	}
}

func TestConvertPlusNationalAnd00(t *testing.T) {
	expectRSRField := &RSRField{Id: "effective_caller_id_number", RSRules: []*ReSearchReplace{
		&ReSearchReplace{SearchRegexp: regexp.MustCompile(`\+49(\d+)`), ReplaceTemplate: "0$1"},
		&ReSearchReplace{SearchRegexp: regexp.MustCompile(`\+(\d+)`), ReplaceTemplate: "00$1"}}}
	rsrField, err := NewRSRField(`~effective_caller_id_number:s/\+49(\d+)/0$1/:s/\+(\d+)/00$1/`)
	if err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rsrField, expectRSRField) {
		t.Errorf("Expecting: %v, received: %v", expectRSRField, rsrField)
	}
	if parsedVal := rsrField.ParseValue("+4986517174963"); parsedVal != "086517174963" {
		t.Errorf("Expecting: 086517174963, received: %s", parsedVal)
	}
	if parsedVal := rsrField.ParseValue("+3186517174963"); parsedVal != "003186517174963" {
		t.Errorf("Expecting: 003186517174963, received: %s", parsedVal)
	}
}

func TestRSRParseStatic(t *testing.T) {
	if rsrField, err := NewRSRField("^static_header::static_value/"); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rsrField, &RSRField{Id: "static_header", staticValue: "static_value"}) {
		t.Errorf("Unexpected RSRField received: %v", rsrField)
	} else if parsed := rsrField.ParseValue("dynamic_value"); parsed != "static_value" {
		t.Errorf("Expected: %s, received: %s", "static_value", parsed)
	}
	if rsrField, err := NewRSRField(`^static_hdrvalue`); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rsrField, &RSRField{Id: "static_hdrvalue", staticValue: "static_hdrvalue"}) {
		t.Errorf("Unexpected RSRField received: %v", rsrField)
	} else if parsed := rsrField.ParseValue("dynamic_value"); parsed != "static_hdrvalue" {
		t.Errorf("Expected: %s, received: %s", "static_hdrvalue", parsed)
	}
}

func TestConvertDurToSecs(t *testing.T) {
	expectRSRField := &RSRField{Id: "9", RSRules: []*ReSearchReplace{
		&ReSearchReplace{SearchRegexp: regexp.MustCompile(`^(\d+)$`), ReplaceTemplate: "${1}s"}}}
	rsrField, err := NewRSRField(`~9:s/^(\d+)$/${1}s/`)
	if err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rsrField, expectRSRField) {
		t.Errorf("Expecting: %v, received: %v", expectRSRField, rsrField)
	}
	if parsedVal := rsrField.ParseValue("640113"); parsedVal != "640113s" {
		t.Errorf("Expecting: 640113s, received: %s", parsedVal)
	}
}

func TestPrefix164(t *testing.T) {
	expectRSRField := &RSRField{Id: "0", RSRules: []*ReSearchReplace{
		&ReSearchReplace{SearchRegexp: regexp.MustCompile(`^([1-9]\d+)$`), ReplaceTemplate: "+$1"}}}
	rsrField, err := NewRSRField(`~0:s/^([1-9]\d+)$/+$1/`)
	if err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rsrField, expectRSRField) {
		t.Errorf("Expecting: %v, received: %v", expectRSRField, rsrField)
	}
	if parsedVal := rsrField.ParseValue("4986517174960"); parsedVal != "+4986517174960" {
		t.Errorf("Expecting: +4986517174960, received: %s", parsedVal)
	}
}

func TestIsStatic(t *testing.T) {
	rsr1 := &RSRField{Id: "0", staticValue: "0"}
	if !rsr1.IsStatic() {
		t.Error("Failed to detect static value.")
	}
	rsr2 := &RSRField{Id: "0", RSRules: []*ReSearchReplace{&ReSearchReplace{SearchRegexp: regexp.MustCompile(`^([1-9]\d+)$`), ReplaceTemplate: "+$1"}}}
	if rsr2.IsStatic() {
		t.Error("Non static detected as static value")
	}
}

func TestParseRSRFields(t *testing.T) {
	fieldsStr1 := `~account:s/^\w+[mpls]\d{6}$//;~subject:s/^0\d{9}$//;^destination/+4912345/;~mediation_runid:s/^default$/default/`
	rsrFld1, _ := NewRSRField(`~account:s/^\w+[mpls]\d{6}$//`)
	rsrFld2, _ := NewRSRField(`~subject:s/^0\d{9}$//`)
	rsrFld3, _ := NewRSRField(`^destination/+4912345/`)
	rsrFld4, _ := NewRSRField(`~mediation_runid:s/^default$/default/`)
	eRSRFields := RSRFields{rsrFld1, rsrFld2, rsrFld3, rsrFld4}
	if rsrFlds, err := ParseRSRFields(fieldsStr1, INFIELD_SEP); err != nil {
		t.Error("Unexpected error: ", err)
	} else if !reflect.DeepEqual(eRSRFields, rsrFlds) {
		t.Errorf("Expecting: %v, received: %v", eRSRFields, rsrFlds)
	}
	fields := `host,~sip_redirected_to:s/sip:\+49(\d+)@/0$1/,destination`
	expectParsedFields := RSRFields{
		&RSRField{Id: "host"},
		&RSRField{Id: "sip_redirected_to",
			RSRules: []*ReSearchReplace{&ReSearchReplace{SearchRegexp: regexp.MustCompile(`sip:\+49(\d+)@`), ReplaceTemplate: "0$1"}}},
		&RSRField{Id: "destination"}}
	if parsedFields, err := ParseRSRFields(fields, FIELDS_SEP); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if !reflect.DeepEqual(parsedFields, expectParsedFields) {
		t.Errorf("Unexpected value of parsed fields")
	}
}

func TestParseCdrcDn1(t *testing.T) {
	if rl, err := NewRSRField(`~1:s/^00(\d+)(?:[a-zA-Z].{3})*0*([1-9]\d+)$/+$1$2/:s/^\+49(18\d{2})$/+491400$1/`); err != nil {
		t.Error("Unexpected error: ", err)
	} else if parsed := rl.ParseValue("0049ABOC0630415354"); parsed != "+49630415354" {
		t.Errorf("Expecting: +49630415354, received: %s", parsed)
	} else if parsed2 := rl.ParseValue("00491888"); parsed2 != "+4914001888" {
		t.Errorf("Expecting: +4914001888, received: %s", parsed2)
	}
}

func TestFilterPasses(t *testing.T) {
	rl, err := NewRSRField(`~1:s/^00(\d+)(?:[a-zA-Z].{3})*0*([1-9]\d+)$/+$1$2/:s/^\+49(18\d{2})$/+491400$1/(+49630415354)`)
	if err != nil {
		t.Error("Unexpected error: ", err)
	}
	if rl.FilterPasses("0031ABOC0630415354") {
		t.Error("Passing filter")
	}
	rl, err = NewRSRField(`~1:s/^$/_empty_/(_empty_)`)
	if err != nil {
		t.Error("Unexpected error: ", err)
	}
	if !rl.FilterPasses("") {
		t.Error("Not passing filter")
	}
	if rl.FilterPasses("Non empty") {
		t.Error("Passing filter")
	}
}

func TestRSRFieldsId(t *testing.T) {
	fieldsStr1 := `~account:s/^\w+[mpls]\d{6}$//;~subject:s/^0\d{9}$//;^destination/+4912345/;~mediation_runid:s/^default$/default/`
	if rsrFlds, err := ParseRSRFields(fieldsStr1, INFIELD_SEP); err != nil {
		t.Error("Unexpected error: ", err)
	} else if idRcv := rsrFlds.Id(); idRcv != "account" {
		t.Errorf("Received id: %s", idRcv)
	}
	fieldsStr2 := ""
	if rsrFlds, err := ParseRSRFields(fieldsStr2, INFIELD_SEP); err != nil {
		t.Error("Unexpected error: ", err)
	} else if idRcv := rsrFlds.Id(); idRcv != "" {
		t.Errorf("Received id: %s", idRcv)
	}
}
