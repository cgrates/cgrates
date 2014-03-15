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

package config

import (
	"reflect"
	"regexp"
	"testing"

	"github.com/cgrates/cgrates/utils"
)

func TestParseSearchReplaceFromFieldRule(t *testing.T) {
	// Normal case
	fieldRule := `~sip_redirected_to:s/sip:\+49(\d+)@/0$1/`
	field, regSrchRplc, err := parseSearchReplaceFromFieldRule(fieldRule)
	if len(field) == 0 || regSrchRplc == nil || err != nil {
		t.Error("Failed parsing the field rule")
	} else if !reflect.DeepEqual(regSrchRplc, &utils.ReSearchReplace{regexp.MustCompile(`sip:\+49(\d+)@`), "0$1"}) {
		t.Error("Unexpected ReSearchReplace parsed")
	}
	// Missing ~ prefix
	fieldRule = `sip_redirected_to:s/sip:\+49(\d+)@/0$1/`
	if _, _, err := parseSearchReplaceFromFieldRule(fieldRule); err == nil {
		t.Error("Parse error, field rule does not start with ~")
	}
	// Separator escaped
	fieldRule = `~sip_redirected_to:s\/sip:\+49(\d+)@/0$1/`
	if _, _, err := parseSearchReplaceFromFieldRule(fieldRule); err == nil {
		t.Error("Parse error, field rule does not contain correct number of separators")
	}
	// One extra separator but escaped
	fieldRule = `~sip_redirected_to:s/sip:\+49(\d+)\/@/0$1/`
	field, regSrchRplc, err = parseSearchReplaceFromFieldRule(fieldRule)
	if len(field) == 0 || regSrchRplc == nil || err != nil {
		t.Error("Failed parsing the field rule")
	} else if !reflect.DeepEqual(regSrchRplc, &utils.ReSearchReplace{regexp.MustCompile(`sip:\+49(\d+)\/@`), "0$1"}) {
		t.Error("Unexpected ReSearchReplace parsed")
	}
}

func TestParseRSRFields(t *testing.T) {
	fields := `host,~sip_redirected_to:s/sip:\+49(\d+)@/0$1/,destination`
	expectParsedFields := []*utils.RSRField{&utils.RSRField{Id: "host"},
		&utils.RSRField{Id: "sip_redirected_to", RSRule: &utils.ReSearchReplace{regexp.MustCompile(`sip:\+49(\d+)@`), "0$1"}},
		&utils.RSRField{Id: "destination"}}
	if parsedFields, err := ParseRSRFields(fields); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if !reflect.DeepEqual(parsedFields, expectParsedFields) {
		t.Errorf("Unexpected value of parsed fields")
	}
}
