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
	"fmt"
	"regexp"
	"strings"
)

func NewRSRField(fldStr string) (*RSRField, error) {
	if len(fldStr) == 0 {
		return nil, nil
	}
	if strings.HasPrefix(fldStr, STATIC_VALUE_PREFIX) { // Special case when RSR is defined as static header/value
		var staticHdr, staticVal string
		if splt := strings.Split(fldStr, HDR_VAL_SEP); len(splt) == 2 { // Using | as separator since ':' is often use in date/time fields
			staticHdr, staticVal = splt[0][1:], splt[1]
		} else {
			staticHdr, staticVal = splt[0][1:], splt[0][1:] // If no split, header will remain as original, value as header without the prefix
		}
		return &RSRField{Id: staticHdr, staticValue: staticVal}, nil
	}
	if !strings.HasPrefix(fldStr, REGEXP_PREFIX) {
		return &RSRField{Id: fldStr}, nil
	}
	spltRgxp := regexp.MustCompile(`:s\/`)
	spltRules := spltRgxp.Split(fldStr, -1)
	if len(spltRules) < 2 {
		return nil, fmt.Errorf("Invalid Search&Replace field rule. %s", fldStr)
	}
	rsrField := &RSRField{Id: spltRules[0][1:]} // Original id in form ~hdr_name
	rulesRgxp := regexp.MustCompile(`(?:(.+[^\\])\/(.+[^\\])*\/){1,}`)
	for _, ruleStr := range spltRules[1:] { // :s/ already removed through split
		allMatches := rulesRgxp.FindStringSubmatch(ruleStr)
		if len(allMatches) != 3 {
			return nil, fmt.Errorf("Invalid Search&Replace field rule. %s", fldStr)
		}
		if srRegexp, err := regexp.Compile(allMatches[1]); err != nil {
			return nil, fmt.Errorf("Invalid Search&Replace field rule. %s", fldStr)
		} else {
			rsrField.RSRules = append(rsrField.RSRules, &ReSearchReplace{SearchRegexp: srRegexp, ReplaceTemplate: allMatches[2]})
		}
	}
	return rsrField, nil
}

type RSRField struct {
	Id          string             //  Identifier
	RSRules     []*ReSearchReplace // Rules to use when processing field value
	staticValue string             // If defined, enforces parsing always to this value
}

// Parse the field value from a string
func (rsrf *RSRField) ParseValue(value string) string {
	if len(rsrf.staticValue) != 0 { // Enforce parsing of static values
		return rsrf.staticValue
	}
	if len(value) == 0 {
		return value
	}
	for _, rsRule := range rsrf.RSRules {
		if rsRule != nil {
			value = rsRule.Process(value)
		}
	}
	return value
}

func (rsrf *RSRField) IsStatic() bool {
	return len(rsrf.staticValue) != 0
}

func (rsrf *RSRField) RegexpMatched() bool { // Investigate whether we had a regexp match through the rules
	for _, rsrule := range rsrf.RSRules {
		if rsrule.Matched {
			return true
		}
	}
	return false
}
