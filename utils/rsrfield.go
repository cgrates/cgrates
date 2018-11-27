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

func NewRSRField(fldStr string) (fld *RSRField, err error) {
	if len(fldStr) == 0 {
		return nil, nil
	}
	rsrField := &RSRField{Rules: fldStr}
	var filters []*RSRFilter
	if strings.HasSuffix(fldStr, FILTER_VAL_END) { // Has filter, populate the var
		fltrStart := strings.LastIndex(fldStr, FILTER_VAL_START)
		if fltrStart < 1 {
			return nil, fmt.Errorf("Invalid FilterStartValue in string: %s", fldStr)
		}
		fltrVal := fldStr[fltrStart+1 : len(fldStr)-1]
		filters, err = ParseRSRFilters(fltrVal, MetaPipe)
		if err != nil {
			return nil, fmt.Errorf("Invalid FilterValue in string: %s, err: %s", fltrVal, err.Error())
		}
		fldStr = fldStr[:fltrStart] // Take the filter part out before compiling further

	}
	if idxConverters := strings.Index(fldStr, "{*"); idxConverters != -1 { // converters in the string
		if !strings.HasSuffix(fldStr, "}") {
			return nil,
				fmt.Errorf("Invalid converter value in string: %s, err: invalid converter terminator",
					fldStr)
		}
		convertersStr := fldStr[idxConverters+1 : len(fldStr)-1] // strip also {}
		convsSplt := strings.Split(convertersStr, INFIELD_SEP)
		rsrField.converters = make([]DataConverter, len(convsSplt))
		for i, convStr := range convsSplt {
			if conv, err := NewDataConverter(convStr); err != nil {
				return nil,
					fmt.Errorf("Invalid converter value in string: %s, err: %s",
						convStr, err.Error())
			} else {
				rsrField.converters[i] = conv
			}
		}
		fldStr = fldStr[:idxConverters]
	}
	if strings.HasPrefix(fldStr, STATIC_VALUE_PREFIX) { // Special case when RSR is defined as static header/value
		var staticHdr, staticVal string
		if splt := strings.Split(fldStr, STATIC_HDRVAL_SEP); len(splt) == 2 { // Using / as separator since ':' is often use in date/time fields
			staticHdr, staticVal = splt[0][1:], splt[1] // Strip the / suffix
			if strings.HasSuffix(staticVal, "/") {      // If value ends with /, strip it since it is a part of the definition syntax
				staticVal = staticVal[:len(staticVal)-1]
			}
		} else if len(splt) > 2 {
			return nil, fmt.Errorf("Invalid RSRField string: %s", fldStr)
		} else {
			staticHdr, staticVal = splt[0][1:], splt[0][1:] // If no split, header will remain as original, value as header without the prefix
		}
		rsrField.Id = staticHdr
		rsrField.staticValue = staticVal
		rsrField.filters = filters
		return rsrField, nil
	}
	if !strings.HasPrefix(fldStr, REGEXP_PREFIX) {
		rsrField.Id = fldStr
		rsrField.filters = filters
		return rsrField, nil
	}
	spltRgxp := regexp.MustCompile(`:s\/`)
	spltRules := spltRgxp.Split(fldStr, -1)
	if len(spltRules) < 2 {
		return nil, fmt.Errorf("Invalid Split of Search&Replace field rule. %s", fldStr)
	}
	rsrField.Id = spltRules[0][1:]
	rsrField.filters = filters // Original id in form ~hdr_name
	rulesRgxp := regexp.MustCompile(`(?:(.+[^\\])\/(.*[^\\])*\/){1,}`)
	for _, ruleStr := range spltRules[1:] { // :s/ already removed through split
		allMatches := rulesRgxp.FindStringSubmatch(ruleStr)
		if len(allMatches) != 3 {
			return nil, fmt.Errorf("Not enough members in Search&Replace, ruleStr: %s, matches: %v, ", ruleStr, allMatches)
		}
		if srRegexp, err := regexp.Compile(allMatches[1]); err != nil {
			return nil, fmt.Errorf("Invalid Search&Replace subfield rule: %s", allMatches[1])
		} else {
			rsrField.RSRules = append(rsrField.RSRules, &ReSearchReplace{SearchRegexp: srRegexp, ReplaceTemplate: allMatches[2]})
		}
	}
	return rsrField, nil
}

func NewRSRFieldMustCompile(fldStr string) (rsrFld *RSRField) {
	var err error
	if rsrFld, err = NewRSRField(fldStr); err != nil {
		return nil
	}
	return
}

type RSRField struct {
	Id          string             // Identifier
	Rules       string             // Rules container holding the string rules to be able to restore it after DB
	staticValue string             // If defined, enforces parsing always to this value
	RSRules     []*ReSearchReplace // Rules to use when processing field value
	filters     []*RSRFilter       // The value to compare when used as filter
	converters  DataConverters     // set of converters to apply on output
}

// IsCompiled finds out whether this RSRField was already parsed or RAW state
func (rsrf *RSRField) IsCompiled() bool {
	return rsrf.staticValue != "" ||
		rsrf.RSRules != nil ||
		rsrf.filters != nil ||
		rsrf.converters != nil
}

// Compile parses Rules string and repopulates other fields
func (rsrf *RSRField) Compile() error {
	if newRSRFld, err := NewRSRField(rsrf.Rules); err != nil {
		return err
	} else if newRSRFld != nil {
		rsrf.staticValue = newRSRFld.staticValue
		rsrf.RSRules = newRSRFld.RSRules
		rsrf.filters = newRSRFld.filters
	}
	return nil
}

// IsStatic detects if a RSRField is a static value
func (rsrf *RSRField) IsStatic() bool {
	return len(rsrf.staticValue) != 0
}

// RegexpMatched will investigate whether we had a regexp match through the rules
func (rsrf *RSRField) RegexpMatched() bool {
	for _, rsrule := range rsrf.RSRules {
		if rsrule.Matched {
			return true
		}
	}
	return false
}

// parseValue the field value from a string
func (rsrf *RSRField) parseValue(value string) string {
	if len(rsrf.staticValue) != 0 { // Enforce parsing of static values
		return rsrf.staticValue
	}
	for _, rsRule := range rsrf.RSRules {
		if rsRule != nil {
			value = rsRule.Process(value)
		}
	}
	return value
}

func (rsrf *RSRField) filtersPassing(value string) bool {
	for _, fltr := range rsrf.filters {
		if !fltr.Pass(value) {
			return false
		}
	}
	return true
}

// Parse will parse the value out considering converters and filters
func (rsrf *RSRField) Parse(value interface{}) (out string, err error) {
	if out, err = IfaceAsString(value); err != nil {
		return
	}
	out = rsrf.parseValue(out)
	if out, err = rsrf.converters.ConvertString(out); err != nil {
		return
	}
	if !rsrf.filtersPassing(out) {
		return "", ErrFilterNotPassingNoCaps
	}
	return
}

// NewRSRFilter instantiates a new RSRFilter, setting it's properties
func NewRSRFilter(fltrVal string) (rsrFltr *RSRFilter, err error) {
	rsrFltr = new(RSRFilter)
	if fltrVal == "" {
		return rsrFltr, nil
	}
	if fltrVal[:1] == NegativePrefix {
		rsrFltr.negative = true
		fltrVal = fltrVal[1:]
		if fltrVal == "" {
			return rsrFltr, nil
		}
	}
	rsrFltr.filterRule = fltrVal
	if fltrVal[:1] == REGEXP_PREFIX {
		if rsrFltr.fltrRgxp, err = regexp.Compile(fltrVal[1:]); err != nil {
			return nil, err
		}
	}
	return rsrFltr, nil
}

// NewRSRFilterMustCompile is used mostly in tests
func NewRSRFilterMustCompile(fltrVal string) (rsrFltr *RSRFilter) {
	var err error
	if rsrFltr, err = NewRSRFilter(fltrVal); err != nil {
		panic(fmt.Sprintf("parsing <%s>, err: %s", fltrVal, err.Error()))
	}
	return
}

// One filter rule
type RSRFilter struct {
	filterRule string // Value in raw format
	fltrRgxp   *regexp.Regexp
	negative   bool // Rule should not match
}

func (rsrFltr *RSRFilter) FilterRule() string {
	return rsrFltr.filterRule
}

func (rsrFltr *RSRFilter) Pass(val string) bool {
	if rsrFltr.filterRule == "" {
		return !rsrFltr.negative
	}
	if rsrFltr.filterRule[:1] == REGEXP_PREFIX {
		return rsrFltr.fltrRgxp.MatchString(val) != rsrFltr.negative
	}
	if rsrFltr.filterRule == "^$" { // Special case to test empty value
		return len(val) == 0 != rsrFltr.negative
	}
	if rsrFltr.filterRule[:1] == MatchStartPrefix {
		if rsrFltr.filterRule[len(rsrFltr.filterRule)-1:] == MatchEndPrefix { // starts with ^ and ends with $, exact match
			return val == rsrFltr.filterRule[1:len(rsrFltr.filterRule)-1] != rsrFltr.negative
		}
		return strings.HasPrefix(val, rsrFltr.filterRule[1:]) != rsrFltr.negative
	}
	lastIdx := len(rsrFltr.filterRule) - 1
	if rsrFltr.filterRule[lastIdx:] == MatchEndPrefix {
		return strings.HasSuffix(val, rsrFltr.filterRule[:lastIdx]) != rsrFltr.negative
	}
	if len(rsrFltr.filterRule) > 2 && rsrFltr.filterRule[:2] == MatchGreaterThanOrEqual {
		gt, err := GreaterThan(StringToInterface(val),
			StringToInterface(rsrFltr.filterRule[2:]), true)
		if err != nil {
			Logger.Warning(fmt.Sprintf("<RSRFilter> rule: <%s>, err: <%s>", rsrFltr.filterRule, err.Error()))
			return false
		}
		return gt != rsrFltr.negative
	}

	if len(rsrFltr.filterRule) > 2 && rsrFltr.filterRule[:2] == MatchLessThanOrEqual {
		gt, err := GreaterThan(StringToInterface(rsrFltr.filterRule[2:]), // compare the rule with the val
			StringToInterface(val),
			true)
		if err != nil {
			Logger.Warning(fmt.Sprintf("<RSRFilter> rule: <%s>, err: <%s>", rsrFltr.filterRule, err.Error()))
			return false
		}
		return gt != rsrFltr.negative
	}

	if rsrFltr.filterRule[:1] == MatchGreaterThan {
		gt, err := GreaterThan(StringToInterface(val),
			StringToInterface(rsrFltr.filterRule[1:]), false)
		if err != nil {
			Logger.Warning(fmt.Sprintf("<RSRFilter> rule: <%s>, err: <%s>", rsrFltr.filterRule, err.Error()))
			return false
		}
		return gt != rsrFltr.negative
	}

	if rsrFltr.filterRule[:1] == MatchLessThan {
		gt, err := GreaterThan(StringToInterface(rsrFltr.filterRule[1:]), // compare the rule with the val
			StringToInterface(val),
			false)
		if err != nil {
			Logger.Warning(fmt.Sprintf("<RSRFilter> rule: <%s>, err: <%s>", rsrFltr.filterRule, err.Error()))
			return false
		}
		return gt != rsrFltr.negative
	}
	return (strings.Index(val, rsrFltr.filterRule) != -1) != rsrFltr.negative // default is string index
}

func ParseRSRFilters(fldsStr, sep string) (RSRFilters, error) {
	if fldsStr == "" {
		return nil, nil
	}
	fltrSplt := strings.Split(fldsStr, sep)
	return ParseRSRFiltersFromSlice(fltrSplt)
}

func ParseRSRFiltersFromSlice(fltrStrs []string) (RSRFilters, error) {
	rsrFltrs := make(RSRFilters, len(fltrStrs))
	for i, rlStr := range fltrStrs {
		if rsrFltr, err := NewRSRFilter(rlStr); err != nil {
			return nil, err
		} else if rsrFltr == nil {
			return nil, fmt.Errorf("Empty RSRFilter in rule: %s", rlStr)
		} else {
			rsrFltrs[i] = rsrFltr
		}
	}
	return rsrFltrs, nil
}

type RSRFilters []*RSRFilter

func (fltrs RSRFilters) FilterRules() (rls string) {
	for _, fltr := range fltrs {
		rls += fltr.FilterRule()
	}
	return
}

// @all: specifies whether all filters should match or at least one
func (fltrs RSRFilters) Pass(val string, allMustMatch bool) (matched bool) {
	if len(fltrs) == 0 {
		return true
	}
	for _, fltr := range fltrs {
		matched = fltr.Pass(val)
		if allMustMatch {
			if !matched {
				return
			}
		} else if matched {
			return
		}
	}
	return
}

func ParseRSRFieldsFromSlice(flds []string) (RSRFields, error) {
	if len(flds) == 0 {
		return nil, nil
	}
	rsrFields := make(RSRFields, len(flds))
	for idx, ruleStr := range flds {
		if rsrField, err := NewRSRField(ruleStr); err != nil {
			return nil, err
		} else if rsrField == nil {
			return nil, fmt.Errorf("Empty RSRField in rule: %s", ruleStr)
		} else {
			rsrFields[idx] = rsrField
		}
	}
	return rsrFields, nil

}

// Parses list of RSRFields, used for example as multiple filters in derived charging
func ParseRSRFields(fldsStr, sep string) (RSRFields, error) {
	//rsrRlsPattern := regexp.MustCompile(`^(~\w+:s/.+/.*/)|(\^.+(/.+/)?)(;(~\w+:s/.+/.*/)|(\^.+(/.+/)?))*$`) //ToDo:Fix here rule able to confirm the content
	if len(fldsStr) == 0 {
		return nil, nil
	}
	rulesSplt := strings.Split(fldsStr, sep)
	return ParseRSRFieldsFromSlice(rulesSplt)

}

func ParseRSRFieldsMustCompile(fldsStr, sep string) RSRFields {
	if flds, err := ParseRSRFields(fldsStr, sep); err != nil {
		return nil
	} else {
		return flds
	}
}

type RSRFields []*RSRField

// Return first Id of the rsrFields, used in cdre
func (flds RSRFields) Id() string {
	if len(flds) == 0 {
		return ""
	}
	return flds[0].Id
}

func (flds RSRFields) Compile() (err error) {
	for _, rsrFld := range flds {
		if err = rsrFld.Compile(); err != nil {
			break
		}
	}
	return
}
