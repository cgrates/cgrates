// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package utils

import (
	"regexp"
)

// ReSearchReplace id regexp Search/Replace, used for example for field formatting
type ReSearchReplace struct {
	SearchRegexp    *regexp.Regexp
	ReplaceTemplate string
}

// Process process the string with the regex
func (rsr *ReSearchReplace) Process(source string) string {
	if rsr.SearchRegexp == nil {
		return EmptyString
	}
	match := rsr.SearchRegexp.FindStringSubmatchIndex(source)
	if match == nil {
		return source // No match returns unaltered source, so we can play with national vs international dialing
	}
	return string(rsr.SearchRegexp.ExpandString(nil, rsr.ReplaceTemplate, source, match))
}

// Clone returns a deep copy of ReSearchReplace
func (rsr *ReSearchReplace) Clone() (cln *ReSearchReplace) {
	if rsr == nil {
		return nil
	}
	cln = &ReSearchReplace{
		ReplaceTemplate: rsr.ReplaceTemplate,
	}
	if rsr.SearchRegexp != nil {
		cln.SearchRegexp = rsr.SearchRegexp.Copy()
	}
	return
}
