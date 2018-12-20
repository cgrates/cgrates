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
	if i := sort.SearchStrings(ss, s); i < len(ss) && ss[i] == s {
		return true
	}
	return false
}

// SliceHasMember is a simpler mode to match inside a slice
// useful to search in shared vars (no slice sort)
func SliceHasMember(ss []string, s string) (has bool) {
	for _, mbr := range ss {
		if mbr == s {
			has = true
			break
		}
	}
	return
}

func SliceWithoutMember(ss []string, s string) []string {
	sort.Strings(ss)
	if i := sort.SearchStrings(ss, s); i < len(ss) && ss[i] == s {
		ss[i], ss = ss[len(ss)-1], ss[:len(ss)-1]
	}
	return ss
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

func Avg(values []float64) float64 {
	if len(values) == 0 {
		return 0.0
	}
	var sum float64
	for _, val := range values {
		sum += val
	}
	return sum / float64(len(values))
}

func AvgNegative(values []float64) float64 {
	if len(values) == 0 {
		return -1 // return -1 if no data
	}
	return Avg(values)
}

func PrefixSliceItems(slc []string, prfx string) (out []string) {
	out = make([]string, len(slc))
	for i, itm := range slc {
		out[i] = prfx + itm
	}
	return
}

// StripSlicePrefix will strip a number of items from the beginning of the slice
func StripSlicePrefix(slc []string, nrItems int) []string {
	if len(slc) < nrItems {
		return []string{}
	}
	return slc[nrItems:]
}

// SliceStringToIface converts slice of strings into a slice of interfaces
func SliceStringToIface(slc []string) (ifc []interface{}) {
	ifc = make([]interface{}, len(slc))
	for i, itm := range slc {
		ifc[i] = itm
	}
	return
}
