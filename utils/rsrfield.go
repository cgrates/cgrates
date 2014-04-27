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
	"fmt"
	"regexp"
	"strings"
)

func NewRSRField(fldStr string) (*RSRField, error) {
	if len(fldStr) == 0 {
		return nil, nil
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
	rulesRgxp := regexp.MustCompile(`(?:(.+[^\\])\/(.+[^\\])\/){1,}`)
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
	Id      string             //  Identifier
	RSRules []*ReSearchReplace // Rules to use when processing field value
}

// Parse the field value from a string
func (rsrf *RSRField) ParseValue(value string) string {
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
