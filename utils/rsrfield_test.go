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

func TestParseSearchReplaceFromFieldRule(t *testing.T) {
	// Normal case
	fieldRule := `~sip_redirected_to:s/sip:\+49(\d+)@/0$1/`
	field, regSrchRplc, err := ParseSearchReplaceFromFieldRule(fieldRule)
	if len(field) == 0 || regSrchRplc == nil || err != nil {
		t.Error("Failed parsing the field rule")
	} else if !reflect.DeepEqual(regSrchRplc, &ReSearchReplace{regexp.MustCompile(`sip:\+49(\d+)@`), "0$1"}) {
		t.Error("Unexpected ReSearchReplace parsed")
	}
	// Missing ~ prefix
	fieldRule = `sip_redirected_to:s/sip:\+49(\d+)@/0$1/`
	if _, _, err := ParseSearchReplaceFromFieldRule(fieldRule); err == nil {
		t.Error("Parse error, field rule does not start with ~")
	}
	// Separator escaped
	fieldRule = `~sip_redirected_to:s\/sip:\+49(\d+)@/0$1/`
	if _, _, err := ParseSearchReplaceFromFieldRule(fieldRule); err == nil {
		t.Error("Parse error, field rule does not contain correct number of separators")
	}
	// One extra separator but escaped
	fieldRule = `~sip_redirected_to:s/sip:\+49(\d+)\/@/0$1/`
	field, regSrchRplc, err = ParseSearchReplaceFromFieldRule(fieldRule)
	if len(field) == 0 || regSrchRplc == nil || err != nil {
		t.Error("Failed parsing the field rule")
	} else if !reflect.DeepEqual(regSrchRplc, &ReSearchReplace{regexp.MustCompile(`sip:\+49(\d+)\/@`), "0$1"}) {
		t.Error("Unexpected ReSearchReplace parsed")
	}
}

func TestParseRSRField1(t *testing.T) {
	fieldRule := `~current_application_data:s/,origination_caller_id_number=(\+?\d+),/$1/`
	if field, regSrchRplc, err := ParseSearchReplaceFromFieldRule(fieldRule); err != nil {
		t.Error("Error parsing the filed rule: ", err.Error())
	} else if field != "current_application_data" {
		t.Error("Failed parsing field name")
	} else if regSrchRplc == nil {
		t.Error("Failed parsing regexp rule")
	}
}

func TestNewRSRField(t *testing.T) {
	expectRSRField := &RSRField{Id: "sip_redirected_to", RSRule: &ReSearchReplace{regexp.MustCompile(`sip:\+49(\d+)@`), "0$1"}}
	if rsrField, err := NewRSRField(`~sip_redirected_to:s/sip:\+49(\d+)@/0$1/`); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rsrField, expectRSRField) {
		t.Errorf("Unexpected RSRField received: %v", rsrField)
	}
	expectRSRField = &RSRField{Id: "sip_redirected_to"}
	if rsrField, err := NewRSRField(`sip_redirected_to`); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rsrField, expectRSRField) {
		t.Errorf("Unexpected RSRField received: %v", rsrField)
	}
	if rsrField, err := NewRSRField(""); err != nil {
		t.Error(err)
	} else if rsrField != nil {
		t.Errorf("Unexpected RSRField received: %v", rsrField)
	}
}

func TestNewRSRFieldDDz(t *testing.T) {
	expectRSRField := &RSRField{Id: "effective_caller_id_number", RSRule: &ReSearchReplace{regexp.MustCompile(`(\d+)`), "+$1"}}
	if rsrField, err := NewRSRField(`~effective_caller_id_number:s/(\d+)/+$1/`); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rsrField, expectRSRField) {
		t.Errorf("Unexpected RSRField received: %v", rsrField)
	}
}
