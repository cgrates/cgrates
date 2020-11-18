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
	if fltrVal[:1] == DynamicDataPrefix {
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
	if rsrFltr.filterRule[:1] == DynamicDataPrefix {
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
