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
	"fmt"
	"regexp"
	"strings"
)

func NewRSRParser(parserRules, inRuleSep string) (rsrParser *RSRParser, err error) {
	if len(parserRules) == 0 {
		return
	}
	rsrParser = &RSRParser{Rules: parserRules}
	if strings.HasSuffix(parserRules, FILTER_VAL_END) { // Has filter, populate the var
		fltrStart := strings.LastIndex(parserRules, FILTER_VAL_START)
		if fltrStart < 1 {
			return nil, fmt.Errorf("invalid RSRFilter start rule in string: <%s>", parserRules)
		}
		fltrVal := parserRules[fltrStart+1 : len(parserRules)-1]
		rsrParser.filters, err = ParseRSRFilters(fltrVal, inRuleSep)
		if err != nil {
			return nil, fmt.Errorf("Invalid FilterValue in string: %s, err: %s", fltrVal, err.Error())
		}
		parserRules = parserRules[:fltrStart] // Take the filter part out before compiling further
	}
	if idxConverters := strings.Index(parserRules, "{*"); idxConverters != -1 { // converters in the string
		if !strings.HasSuffix(parserRules, "}") {
			return nil,
				fmt.Errorf("invalid converter terminator in rule: <%s>",
					parserRules)
		}
		convertersStr := parserRules[idxConverters+1 : len(parserRules)-1] // strip also {}
		convsSplt := strings.Split(convertersStr, inRuleSep)
		rsrParser.converters = make(DataConverters, len(convsSplt))
		for i, convStr := range convsSplt {
			if conv, err := NewDataConverter(convStr); err != nil {
				return nil,
					fmt.Errorf("invalid converter value in string: <%s>, err: %s",
						convStr, err.Error())
			} else {
				rsrParser.converters[i] = conv
			}
		}
		parserRules = parserRules[:idxConverters]
	}
	if !strings.HasPrefix(parserRules, DynamicDataPrefix) { // special case when RSR is defined as static attribute=value
		var staticHdr, staticVal string
		if splt := strings.Split(parserRules, AttrValueSep); len(splt) == 2 { // using '='' as separator since ':' is often use in date/time fields
			staticHdr, staticVal = splt[0], splt[1]         // strip the separator
			if strings.HasSuffix(staticVal, AttrValueSep) { // if value ends with sep, strip it since it is a part of the definition syntax
				staticVal = staticVal[:len(staticVal)-1]
			}
		} else if len(splt) > 2 {
			return nil, fmt.Errorf("invalid RSRField static rules: <%s>", parserRules)
		} else {
			staticVal = splt[0] // no attribute name
		}
		rsrParser.attrName = staticHdr
		rsrParser.attrValue = staticVal
		return
	}
	// dynamic content via attributeNames
	spltRgxp := regexp.MustCompile(`:s\/`)
	spltRules := spltRgxp.Split(parserRules, -1)
	rsrParser.attrName = spltRules[0][1:] // in form ~hdr_name
	if len(spltRules) > 1 {
		rulesRgxp := regexp.MustCompile(`(?:(.*[^\\])\/(.*[^\\])*\/){1,}`)
		for _, ruleStr := range spltRules[1:] { // :s/ already removed through split
			allMatches := rulesRgxp.FindStringSubmatch(ruleStr)
			if len(allMatches) != 3 {
				return nil, fmt.Errorf("not enough members in Search&Replace, ruleStr: <%s>, matches: %v, ", ruleStr, allMatches)
			}
			if srRegexp, err := regexp.Compile(allMatches[1]); err != nil {
				return nil, fmt.Errorf("invalid Search&Replace subfield rule: <%s>", allMatches[1])
			} else {
				rsrParser.rsrRules = append(rsrParser.rsrRules, &ReSearchReplace{SearchRegexp: srRegexp, ReplaceTemplate: allMatches[2]})
			}
		}
	}
	return
}

// RSRParser is a parser for data coming from various sources
type RSRParser struct {
	Rules string // Rules container holding the string rules, public so it can be stored

	attrName   string             // instruct extracting info out of header in event
	attrValue  string             // if populated, enforces parsing always to this value
	rsrRules   []*ReSearchReplace // rules to use when parsing value
	converters DataConverters     // set of converters to apply on output
	filters    RSRFilters         // The value to compare when used as filter
}

func NewRSRParsers(parsersRules, rlsSep, inRlSep string) (prsrs RSRParsers, err error) {
	if parsersRules == "" {
		return
	}
	return NewRSRParsersFromSlice(strings.Split(parsersRules, rlsSep), inRlSep)
}

func NewRSRParsersFromSlice(parsersRules []string, inRlSep string) (prsrs RSRParsers, err error) {
	prsrs = make(RSRParsers, len(parsersRules))
	for i, rlStr := range parsersRules {
		if rsrPrsr, err := NewRSRParser(rlStr, inRlSep); err != nil {
			return nil, err
		} else if rsrPrsr == nil {
			return nil, fmt.Errorf("emtpy RSRParser in rule: <%s>", rlStr)
		} else {
			prsrs[i] = rsrPrsr
		}
	}
	return
}

// RSRParsers is a set of RSRParser
type RSRParsers []*RSRParser
