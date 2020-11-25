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
	"reflect"
	"testing"
)

func TestRSRFilterPass(t *testing.T) {
	fltr, err := NewRSRFilter("") // Pass any
	if err != nil {
		t.Error(err)
	}
	if !fltr.Pass("") {
		t.Error("Not passing!")
	}
	if !fltr.Pass("any") {
		t.Error("Not passing!")
	}
	fltr, err = NewRSRFilter("!") // Pass nothing
	if err != nil {
		t.Error(err)
	}
	if fltr.Pass("") {
		t.Error("Passing!")
	}
	if fltr.Pass("any") {
		t.Error("Passing!")
	}
	fltr, err = NewRSRFilter("^full_match$") // Full string pass
	if err != nil {
		t.Error(err)
	}
	if !fltr.Pass("full_match") {
		t.Error("Not passing!")
	}
	if fltr.Pass("full_match1") {
		t.Error("Passing!")
	}
	if fltr.Pass("") {
		t.Error("Passing!")
	}
	fltr, err = NewRSRFilter("^prefixMatch") // Prefix pass
	if err != nil {
		t.Error(err)
	}
	if !fltr.Pass("prefixMatch") {
		t.Error("Not passing!")
	}
	if !fltr.Pass("prefixMatch12345") {
		t.Error("Not passing!")
	}
	if fltr.Pass("1prefixMatch") {
		t.Error("Passing!")
	}
	fltr, err = NewRSRFilter("suffixMatch$") // Suffix pass
	if err != nil {
		t.Error(err)
	}
	if !fltr.Pass("suffixMatch") {
		t.Error("Not passing!")
	}
	if !fltr.Pass("12345suffixMatch") {
		t.Error("Not passing!")
	}
	if fltr.Pass("suffixMatch1") {
		t.Error("Passing!")
	}
	fltr, err = NewRSRFilter("!fullMatch") // Negative full pass
	if err != nil {
		t.Error(err)
	}
	if !fltr.Pass("ShouldMatch") {
		t.Error("Not passing!")
	}
	if fltr.Pass("fullMatch") {
		t.Error("Passing!")
	}
	fltr, err = NewRSRFilter("!^prefixMatch") // Negative prefix pass
	if err != nil {
		t.Error(err)
	}
	if fltr.Pass("prefixMatch123") {
		t.Error("Passing!")
	}
	if !fltr.Pass("123prefixMatch") {
		t.Error("Not passing!")
	}
	fltr, err = NewRSRFilter("!suffixMatch$") // Negative suffix pass
	if err != nil {
		t.Error(err)
	}
	if fltr.Pass("123suffixMatch") {
		t.Error("Passing!")
	}
	if !fltr.Pass("suffixMatch123") {
		t.Error("Not passing!")
	}
	fltr, err = NewRSRFilter("~^C.+S$") // Regexp pass
	if err != nil {
		t.Error(err)
	}
	if !fltr.Pass("CGRateS") {
		t.Error("Not passing!")
	}
	if fltr.Pass("1CGRateS") {
		t.Error("Passing!")
	}
	fltr, err = NewRSRFilter("!~^C.*S$") // Negative regexp pass
	if err != nil {
		t.Error(err)
	}
	if fltr.Pass("CGRateS") {
		t.Error("Passing!")
	}
	fltr, err = NewRSRFilter("^$") // Empty value
	if err != nil {
		t.Error(err)
	}
	if fltr.Pass("CGRateS") {
		t.Error("Passing!")
	}
	if !fltr.Pass("") {
		t.Error("Not passing!")
	}
	fltr, err = NewRSRFilter("!^$") // Non-empty value
	if err != nil {
		t.Error(err)
	}
	if !fltr.Pass("CGRateS") {
		t.Error("Not passing!")
	}
	if fltr.Pass("") {
		t.Error("Passing!")
	}
	fltr, err = NewRSRFilter("indexed_match") // Indexed match
	if err != nil {
		t.Error(err)
	}
	if !fltr.Pass("indexed_match") {
		t.Error("Not passing!")
	}
	if !fltr.Pass("suf_indexed_match") {
		t.Error("Not passing!")
	}
	if !fltr.Pass("indexed_match_pref") {
		t.Error("Not passing!")
	}
	if !fltr.Pass("suf_indexed_match_pref") {
		t.Error("Not passing!")
	}
	if fltr.Pass("indexed_matc") {
		t.Error("Passing!")
	}
	if fltr.Pass("") {
		t.Error("Passing!")
	}
	fltr, err = NewRSRFilter("!indexed_match") // Negative indexed match
	if err != nil {
		t.Error(err)
	}
	if fltr.Pass("indexed_match") {
		t.Error("passing!")
	}
	if fltr.Pass("suf_indexed_match") {
		t.Error("passing!")
	}
	if fltr.Pass("indexed_match_pref") {
		t.Error("passing!")
	}
	if fltr.Pass("suf_indexed_match_pref") {
		t.Error("passing!")
	}
	if !fltr.Pass("indexed_matc") {
		t.Error("not passing!")
	}
	if !fltr.Pass("") {
		t.Error("Passing!")
	}

	// compare greaterThan
	fltr, err = NewRSRFilter(">0s")
	if err != nil {
		t.Error(err)
	}
	if fltr.Pass("0s") {
		t.Error("passing!")
	}
	if !fltr.Pass("13") {
		t.Error("not passing!")
	}
	if !fltr.Pass("12s") {
		t.Error("not passing!")
	}

	// compare not greaterThan
	fltr, err = NewRSRFilter("!>0s") // !(n>0s)
	if err != nil {
		t.Error(err)
	}
	if !fltr.Pass("0s") {
		t.Error("not passing!")
	}
	if fltr.Pass("13") {
		t.Error("passing!")
	}
	if fltr.Pass("12s") {
		t.Error("passing!")
	}

	// compare greaterThanOrEqual
	fltr, err = NewRSRFilter(">=0s")
	if err != nil {
		t.Error(err)
	}
	if fltr.Pass("-1s") {
		t.Error("passing!")
	}
	if !fltr.Pass("0s") {
		t.Error("not passing!")
	}
	if !fltr.Pass("13") {
		t.Error("not passing!")
	}
	if !fltr.Pass("12s") {
		t.Error("not passing!")
	}

	// compare not greaterThanOrEqual
	fltr, err = NewRSRFilter("!>=0s")
	if err != nil {
		t.Error(err)
	}
	if !fltr.Pass("-1s") {
		t.Error("not passing!")
	}
	if fltr.Pass("0s") {
		t.Error("passing!")
	}
	if fltr.Pass("13") {
		t.Error("passing!")
	}
	if fltr.Pass("12s") {
		t.Error("passing!")
	}

	// compare lessThan
	fltr, err = NewRSRFilter("<0s")
	if err != nil {
		t.Error(err)
	}
	if fltr.Pass("1ns") {
		t.Error("passing!")
	}
	if fltr.Pass("13") {
		t.Error("passing!")
	}
	if fltr.Pass("12s") {
		t.Error("passing!")
	}
	if !fltr.Pass("-12s") {
		t.Error("not passing!")
	}

	// compare not lessThan
	fltr, err = NewRSRFilter("!<0s")
	if err != nil {
		t.Error(err)
	}
	if !fltr.Pass("1ns") {
		t.Error("not passing!")
	}
	if !fltr.Pass("13") {
		t.Error("not passing!")
	}
	if !fltr.Pass("12s") {
		t.Error("not passing!")
	}
	if fltr.Pass("-12s") {
		t.Error("passing!")
	}

	// compare lessThanOrEqual
	fltr, err = NewRSRFilter("<=0s")
	if err != nil {
		t.Error(err)
	}
	if !fltr.Pass("-1s") {
		t.Error("not passing!")
	}
	if !fltr.Pass("0s") {
		t.Error("not passing!")
	}
	if fltr.Pass("13") {
		t.Error("passing!")
	}
	if fltr.Pass("12s") {
		t.Error("passing!")
	}

	// compare not lessThanOrEqual
	fltr, err = NewRSRFilter("!<=0s")
	if err != nil {
		t.Error(err)
	}
	if fltr.Pass("-1s") {
		t.Error("passing!")
	}
	if fltr.Pass("0s") {
		t.Error("passing!")
	}
	if !fltr.Pass("13") {
		t.Error("not passing!")
	}
	if !fltr.Pass("12s") {
		t.Error("not passing!")
	}
}

func TestRSRFiltersPass(t *testing.T) {
	rlStr := "~^C.+S$;CGRateS;ateS$"
	fltrs, err := ParseRSRFilters(rlStr, INFIELD_SEP)
	if err != nil {
		t.Error(err)
	}
	if !fltrs.Pass("CGRateS", true) {
		t.Error("Not passing")
	}
	if fltrs.Pass("ateS", true) {
		t.Error("Passing")
	}
	if !fltrs.Pass("ateS", false) {
		t.Error("Not passing")
	}
	if fltrs.Pass("teS", false) {
		t.Error("Passing")
	}
}

func TestNewRSRFilterError(t *testing.T) {
	filterString := "~(^_^"
	_, err := NewRSRFilter(filterString)
	if err == nil || err.Error() != "error parsing regexp: missing closing ): `(^_^`" {
		t.Errorf("Expected <error parsing regexp: missing closing ): `(^_^`> ,received: <%+v>", err)
	}
}

func TestNewRSRFilterMustCompile(t *testing.T) {
	filterString := "~works"
	filterNew := NewRSRFilterMustCompile(filterString)
	expected, _ := NewRSRFilter(filterString)
	if !reflect.DeepEqual(expected, filterNew) {
		t.Errorf("Expected <%+v> ,received: <%+v>", expected, filterNew)
	}
}

func TestNewRSRFilterMustCompilePanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic on wrong rule")
		}
	}()
	filterString := "~(^_^"
	err := NewRSRFilterMustCompile(filterString)
	if err != nil {
		t.Errorf("Expected <%+v> ,received: <%+v>", nil, err)
	}
}

func TestFilterRuleTest(t *testing.T) {
	newRSRFilter, err := NewRSRFilter("^cat$")
	if err != nil {
		t.Errorf("Expected <nil> ,received: <%+v>", err)
	}
	rule := newRSRFilter.FilterRule()
	if !reflect.DeepEqual("^cat$", rule) {
		t.Errorf("Expected <^cat$> ,received: <%+v>", rule)
	}
}

func TestParseRSRFiltersNil(t *testing.T) {
	testFilter, err := ParseRSRFilters("", "notnil")
	if testFilter != nil {
		t.Errorf("Expected <nil> ,received: <%+v>", testFilter)
	}
	if err != nil {
		t.Errorf("Expected <nil> ,received: <%+v>", err)
	}

}

func TestParseRSRFilterssPassZeroLength(t *testing.T) {
	err := new(RSRFilters).Pass("", false)
	if !reflect.DeepEqual(err, true) {
		t.Errorf("Expected <true> ,received: <%+v>", err)
	}
}

func TestParseRSRFilterRules(t *testing.T) {
	filter1, _ := NewRSRFilter("^rule1$")
	filter2, _ := NewRSRFilter("^rule2$")
	filters := RSRFilters{filter1, filter2}
	result := filters.FilterRules()
	if !reflect.DeepEqual(result, "^rule1$^rule2$") {
		t.Errorf("Expected <^rule1$^rule2$> ,received: <%+v>", result)
	}
}

func TestParseRSRFiltersFromSliceError(t *testing.T) {
	stringSlice := []string{"~(^_^", "^rule2$"}
	_, err := ParseRSRFiltersFromSlice(stringSlice)
	if err == nil || err.Error() != "error parsing regexp: missing closing ): `(^_^`" {
		t.Errorf("Expected <error parsing regexp: missing closing ): `(^_^`> ,received: <%+v>", err)
	}
}

func TestRSRFilterPassMatchGreaterThanOrEqual(t *testing.T) {
	fltr, err := NewRSRFilter(">=0")
	if err != nil {
		t.Errorf("Expected <nil> ,received: <%+v>", err)
	}
	result := fltr.Pass("string")
	if !reflect.DeepEqual(false, result) {
		t.Errorf("Expected <false> ,received: <%+v>", result)
	}
}

func TestRSRFilterPassMatchLessThanOrEqual(t *testing.T) {
	fltr, err := NewRSRFilter("<=0")
	if err != nil {
		t.Errorf("Expected <nil> ,received: <%+v>", err)
	}
	result := fltr.Pass("string")
	if !reflect.DeepEqual(false, result) {
		t.Errorf("Expected <false> ,received: <%+v>", result)
	}
}

func TestRSRFilterPassMatchGreaterThan(t *testing.T) {
	fltr, err := NewRSRFilter(">0")
	if err != nil {
		t.Errorf("Expected <nil> ,received: <%+v>", err)
	}
	result := fltr.Pass("invalid")
	if !reflect.DeepEqual(false, result) {
		t.Errorf("Expected <false> ,received: <%+v>", result)
	}
}

func TestRSRFilterPassMatchLessThan(t *testing.T) {
	fltr, err := NewRSRFilter("<0")
	if err != nil {
		t.Errorf("Expected <nil> ,received: <%+v>", err)
	}
	result := fltr.Pass("string")
	if !reflect.DeepEqual(false, result) {
		t.Errorf("Expected <false> ,received: <%+v>", result)
	}
}
