/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2013 ITsysCOM

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
	expRSRField1 := &RSRField{Id: "sip_redirected_to", RSRules: []*ReSearchReplace{&ReSearchReplace{regexp.MustCompile(`sip:\+49(\d+)@`), "0$1"}}}
	if rsrField, err := NewRSRField(`~sip_redirected_to:s/sip:\+49(\d+)@/0$1/`); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if !reflect.DeepEqual(expRSRField1, rsrField) {
		t.Errorf("Expecting: %v, received: %v", expRSRField1, rsrField)
	}
	// Separator escaped
	if rsrField, err := NewRSRField(`~sip_redirected_to:s\/sip:\+49(\d+)@/0$1/`); err == nil {
		t.Error("Parse error, field rule does not contain correct number of separators, received: %v", rsrField)
	}
	// One extra separator but escaped
	expRSRField3 := &RSRField{Id: "sip_redirected_to", RSRules: []*ReSearchReplace{&ReSearchReplace{regexp.MustCompile(`sip:\+49(\d+)\/@`), "0$1"}}}
	if rsrField, err := NewRSRField(`~sip_redirected_to:s/sip:\+49(\d+)\/@/0$1/`); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if !reflect.DeepEqual(expRSRField3, rsrField) {
		t.Errorf("Expecting: %v, received: %v", expRSRField3, rsrField)
	}
}

func TestNewRSRFieldDDz(t *testing.T) {
	expectRSRField := &RSRField{Id: "effective_caller_id_number", RSRules: []*ReSearchReplace{&ReSearchReplace{regexp.MustCompile(`(\d+)`), "+$1"}}}
	if rsrField, err := NewRSRField(`~effective_caller_id_number:s/(\d+)/+$1/`); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rsrField, expectRSRField) {
		t.Errorf("Unexpected RSRField received: %v", rsrField)
	}
}

func TestNewRSRFieldIvo(t *testing.T) {
	expectRSRField := &RSRField{Id: "cost_details", RSRules: []*ReSearchReplace{&ReSearchReplace{regexp.MustCompile(`MatchedDestId":".+_(\s\s\s\s\s)"`), "$1"}}}
	if rsrField, err := NewRSRField(`~cost_details:s/MatchedDestId":".+_(\s\s\s\s\s)"/$1/`); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rsrField, expectRSRField) {
		t.Errorf("Unexpected RSRField received: %v", rsrField)
	}
}

func TestConvertPlusNationalAnd00(t *testing.T) {
	expectRSRField := &RSRField{Id: "effective_caller_id_number", RSRules: []*ReSearchReplace{
		&ReSearchReplace{regexp.MustCompile(`\+49(\d+)`), "0$1"},
		&ReSearchReplace{regexp.MustCompile(`\+(\d+)`), "00$1"}}}
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
	if rsrField, err := NewRSRField("^static_header/static_value"); err != nil {
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
