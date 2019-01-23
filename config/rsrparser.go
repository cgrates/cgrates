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
	"fmt"
	"regexp"
	"strings"

	"github.com/cgrates/cgrates/utils"
)

func NewRSRParsers(parsersRules string, allFiltersMatch bool, rsrSeparator string) (prsrs RSRParsers, err error) {
	if parsersRules == "" {
		return
	}
	return NewRSRParsersFromSlice(strings.Split(parsersRules, rsrSeparator), allFiltersMatch)
}

func NewRSRParsersFromSlice(parsersRules []string, allFiltersMatch bool) (prsrs RSRParsers, err error) {
	prsrs = make(RSRParsers, len(parsersRules))
	for i, rlStr := range parsersRules {
		if rsrPrsr, err := NewRSRParser(rlStr, allFiltersMatch); err != nil {
			return nil, err
		} else if rsrPrsr == nil {
			return nil, fmt.Errorf("emtpy RSRParser in rule: <%s>", rlStr)
		} else {
			prsrs[i] = rsrPrsr
		}
	}
	return
}

func NewRSRParsersMustCompile(parsersRules string, allFiltersMatch bool, rsrSeparator string) (prsrs RSRParsers) {
	var err error
	if prsrs, err = NewRSRParsers(parsersRules, allFiltersMatch, rsrSeparator); err != nil {
		panic(fmt.Sprintf("rule: <%s>, error: %s", parsersRules, err.Error()))
	}
	return
}

// RSRParsers is a set of RSRParser
type RSRParsers []*RSRParser

func (prsrs RSRParsers) Compile() (err error) {
	for _, prsr := range prsrs {
		if err = prsr.Compile(); err != nil {
			return
		}
	}
	return
}

// ParseValue will parse the value out considering converters and filters
func (prsrs RSRParsers) ParseValue(value interface{}) (out string, err error) {
	for _, prsr := range prsrs {
		if outPrsr, err := prsr.ParseValue(value); err != nil {
			return "", err
		} else {
			out += outPrsr
		}
	}
	return
}

// ParseEvent will parse the event values into one output
func (prsrs RSRParsers) ParseEvent(ev map[string]interface{}) (out string, err error) {
	for _, prsr := range prsrs {
		if outPrsr, err := prsr.ParseEvent(ev); err != nil {
			return "", err
		} else {
			out += outPrsr
		}
	}
	return
}

func (prsrs RSRParsers) ParseDataProvider(dP DataProvider, separator string) (out string, err error) {
	for _, prsr := range prsrs {
		if outPrsr, err := prsr.ParseDataProvider(dP, separator); err != nil {
			return "", err
		} else {
			out += outPrsr
		}
	}
	return
}

func (prsrs RSRParsers) ParseDataProviderWithInterfaces(dP DataProvider, separator string) (out string, err error) {
	for _, prsr := range prsrs {
		if outPrsr, err := prsr.ParseDataProviderWithInterfaces(dP, separator); err != nil {
			return "", err
		} else {
			out += outPrsr
		}
	}
	return
}

func NewRSRParser(parserRules string, allFiltersMatch bool) (rsrParser *RSRParser, err error) {
	if len(parserRules) == 0 {
		return
	}
	rsrParser = &RSRParser{Rules: parserRules, AllFiltersMatch: allFiltersMatch}
	if strings.HasSuffix(parserRules, utils.FILTER_VAL_END) { // Has filter, populate the var
		fltrStart := strings.LastIndex(parserRules, utils.FILTER_VAL_START)
		if fltrStart < 1 {
			return nil, fmt.Errorf("invalid RSRFilter start rule in string: <%s>", parserRules)
		}
		fltrVal := parserRules[fltrStart+1 : len(parserRules)-1]
		rsrParser.filters, err = utils.ParseRSRFilters(fltrVal, utils.ANDSep)
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
		convsSplt := strings.Split(convertersStr, utils.ANDSep)
		rsrParser.converters = make(utils.DataConverters, len(convsSplt))
		for i, convStr := range convsSplt {
			if conv, err := utils.NewDataConverter(convStr); err != nil {
				return nil,
					fmt.Errorf("invalid converter value in string: <%s>, err: %s",
						convStr, err.Error())
			} else {
				rsrParser.converters[i] = conv
			}
		}
		parserRules = parserRules[:idxConverters]
	}
	if !strings.HasPrefix(parserRules, utils.DynamicDataPrefix) { // special case when RSR is defined as static attribute=value
		var staticHdr, staticVal string
		if splt := strings.Split(parserRules, utils.AttrValueSep); len(splt) == 2 { // using '='' as separator since ':' is often use in date/time fields
			staticHdr, staticVal = splt[0], splt[1]               // strip the separator
			if strings.HasSuffix(staticVal, utils.AttrValueSep) { // if value ends with sep, strip it since it is a part of the definition syntax
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
				rsrParser.rsrRules = append(rsrParser.rsrRules, &utils.ReSearchReplace{SearchRegexp: srRegexp, ReplaceTemplate: allMatches[2]})
			}
		}
	}
	return
}

func NewRSRParserMustCompile(parserRules string, allFiltersMatch bool) (rsrPrsr *RSRParser) {
	var err error
	if rsrPrsr, err = NewRSRParser(parserRules, allFiltersMatch); err != nil {
		panic(fmt.Sprintf("compiling rules: <%s>, error: %s", parserRules, err.Error()))
	}
	return
}

// RSRParser is a parser for data coming from various sources
type RSRParser struct {
	Rules           string // Rules container holding the string rules, public so it can be stored
	AllFiltersMatch bool   // all filters must match policy

	attrName   string                   // instruct extracting info out of header in event
	attrValue  string                   // if populated, enforces parsing always to this value
	rsrRules   []*utils.ReSearchReplace // rules to use when parsing value
	converters utils.DataConverters     // set of converters to apply on output
	filters    utils.RSRFilters         // The value to compare when used as filter
}

// AttrName exports the attribute name of the RSRParser
func (prsr *RSRParser) AttrName() string {
	return prsr.attrName
}

// Compile parses Rules string and repopulates other fields
func (prsr *RSRParser) Compile() (err error) {
	var newPrsr *RSRParser
	if newPrsr, err = NewRSRParser(prsr.Rules, prsr.AllFiltersMatch); err != nil {
		return
	}
	if newPrsr != nil {
		*prsr = *newPrsr
	}
	return
}

// RegexpMatched will investigate whether we had at least one regexp match through the rules
func (prsr *RSRParser) RegexpMatched() bool {
	for _, rsrule := range prsr.rsrRules {
		if rsrule.Matched {
			return true
		}
	}
	return false
}

// parseValue the field value from a string
func (prsr *RSRParser) parseValue(value string) string {
	if prsr.attrValue != "" { // Enforce parsing of static values
		return prsr.attrValue
	}
	for _, rsRule := range prsr.rsrRules {
		value = rsRule.Process(value)
	}
	return value
}

// ParseValue will parse the value out considering converters and filters
func (prsr *RSRParser) ParseValue(value interface{}) (out string, err error) {
	if out, err = utils.IfaceAsString(value); err != nil {
		return
	}
	out = prsr.parseValue(out)
	if out, err = prsr.converters.ConvertString(out); err != nil {
		return
	}
	if !prsr.filters.Pass(out, prsr.AllFiltersMatch) {
		return "", utils.ErrFilterNotPassingNoCaps
	}
	return
}

// ParseEvent will parse the value out considering converters and filters
func (prsr *RSRParser) ParseEvent(ev map[string]interface{}) (out string, err error) {
	val, has := ev[prsr.attrName]
	if !has && prsr.attrValue == "" {
		return "", utils.ErrNotFound
	}
	return prsr.ParseValue(val)
}

func (prsr *RSRParser) ParseDataProvider(dP DataProvider, separator string) (out string, err error) {
	var outStr string
	if prsr.attrValue == "" {
		if outStr, err = dP.FieldAsString(
			strings.Split(prsr.attrName, separator)); err != nil &&
			(err != utils.ErrNotFound || prsr.filters.FilterRules() != "^$") {
			return
		}
	}
	return prsr.ParseValue(outStr)
}

func (prsr *RSRParser) ParseDataProviderWithInterfaces(dP DataProvider, separator string) (out string, err error) {
	var outIface interface{}
	if prsr.attrValue == "" {
		if outIface, err = dP.FieldAsInterface(
			strings.Split(prsr.attrName, separator)); err != nil &&
			(err != utils.ErrNotFound || prsr.filters.FilterRules() != "^$") {
			return
		}
	}
	return prsr.ParseValue(outIface)
}
