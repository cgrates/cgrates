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
	"sort"
	"strings"
)

// Binary string search in slice
func IsSliceMember(ss []string, s string) bool {
	sort.Strings(ss)
	return SliceHasMember(ss, s)
}

// SliceHasMember searches within a *sorted* slice
// useful to search in shared vars (no slice sort)
func SliceHasMember(ss []string, s string) bool {
	i := sort.SearchStrings(ss, s)
	return i < len(ss) && ss[i] == s
}

// PrefixSliceItems iterates through slice and add a prefix before every element
func PrefixSliceItems(slc []string, prfx string) (out []string) {
	out = make([]string, len(slc))
	for i, itm := range slc {
		out[i] = prfx + itm
	}
	return
}

// SliceStringToIface converts slice of strings into a slice of interfaces
func SliceStringToIface(slc []string) (ifc []interface{}) {
	ifc = make([]interface{}, len(slc))
	for i, itm := range slc {
		ifc[i] = itm
	}
	return
}

// Float64SliceHasMember searches within a *sorted* slice
// useful to search in shared vars (no slice sort)
func Float64SliceHasMember(ss []float64, s float64) bool {
	i := sort.SearchFloat64s(ss, s)
	return i < len(ss) && ss[i] == s
}

// HasPrefixSlice iterates over slice members and returns true if one the element has that prefix
func HasPrefixSlice(prfxs []string, el string) bool {
	for _, prfx := range prfxs {
		if strings.HasPrefix(el, prfx) {
			return true
		}
	}
	return false
}

func CloneSlice(in []string) (cl []string) {
	cl = make([]string, len(in))
	for i, j := range in {
		cl[i] = j
	}
	return
}
