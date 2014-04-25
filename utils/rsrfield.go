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
	"errors"
	"regexp"
	"strings"
)

func ParseSearchReplaceFromFieldRule(fieldRule string) (string, *ReSearchReplace, error) {
	// String rule expected in the form ~hdr_name:s/match_rule/replace_rule/
	getRuleRgxp := regexp.MustCompile(`~(\w+):s\/(.+[^\\])\/(.+[^\\])\/`) // Make sure the separator / is not escaped in the rule
	allMatches := getRuleRgxp.FindStringSubmatch(fieldRule)
	if len(allMatches) != 4 { // Second and third groups are of interest to us
		return "", nil, errors.New("Invalid Search&Replace field rule.")
	}
	fieldName := allMatches[1]
	searchRegexp, err := regexp.Compile(allMatches[2])
	if err != nil {
		return fieldName, nil, err
	}
	return fieldName, &ReSearchReplace{searchRegexp, allMatches[3]}, nil
}

func NewRSRField(fldStr string) (*RSRField, error) {
	if len(fldStr) == 0 {
		return nil, nil
	}
	if !strings.HasPrefix(fldStr, REGEXP_PREFIX) {
		return &RSRField{Id: fldStr}, nil
	}
	if fldId, reSrcRepl, err := ParseSearchReplaceFromFieldRule(fldStr); err != nil {
		return nil, err
	} else {
		return &RSRField{fldId, reSrcRepl}, nil
	}
}

type RSRField struct {
	Id     string           //  Identifier
	RSRule *ReSearchReplace // Rule to use when processing field value
}

// Parse the field value from a string
func (rsrf *RSRField) ParseValue(value string) string {
	if len(value) == 0 {
		return value
	}
	if rsrf.RSRule != nil {
		value = rsrf.RSRule.Process(value)
	}
	return value
}
