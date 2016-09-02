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
