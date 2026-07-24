// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package utils

import (
	"regexp"
)

// Regexp Search/Replace, used for example for field formatting
type ReSearchReplace struct {
	SearchRegexp    *regexp.Regexp
	ReplaceTemplate string
	Matched         bool
}

func (rsr *ReSearchReplace) Process(source string) string {
	if rsr.SearchRegexp == nil {
		return ""
	}
	res := []byte{}
	match := rsr.SearchRegexp.FindStringSubmatchIndex(source)
	if match == nil {
		return source // No match returns unaltered source, so we can play with national vs international dialing
	} else {
		rsr.Matched = true
	}
	res = rsr.SearchRegexp.ExpandString(res, rsr.ReplaceTemplate, source, match)
	return string(res)
}
