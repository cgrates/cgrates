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
	"sort"
	"strings"
)

// Binary string search in slice
func IsSliceMember(ss []string, s string) bool {
	sort.Strings(ss)
	if i := sort.SearchStrings(ss, s); i < len(ss) && ss[i] == s {
		return true
	}
	return false
}

// Binary string search in slice
// returns true if found and the index
func GetSliceMemberIndex(ss []string, s string) (int, bool) {
	sort.Strings(ss)
	if i := sort.SearchStrings(ss, s); i < len(ss) && ss[i] == s {
		return i, true
	}
	return len(ss), false
}

//Iterates over slice members and returns true if one starts with prefix
func SliceMemberHasPrefix(ss []string, prfx string) bool {
	for _, mbr := range ss {
		if strings.HasPrefix(mbr, prfx) {
			return true
		}
	}
	return false
}

type InterfaceStrings []interface{}

func (a InterfaceStrings) Len() int           { return len(a) }
func (a InterfaceStrings) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a InterfaceStrings) Less(i, j int) bool { return a[i].(string) < a[j].(string) }

// Binary string search in slice
// returns true if found and the index
func GetSliceInterfaceIndex(ss []interface{}, s interface{}) (int, bool) {
	if i := sort.Search(len(ss), func(i int) bool { return ss[i].(string) >= s.(string) }); i < len(ss) && ss[i] == s {
		return i, true
	}
	return len(ss), false
}
