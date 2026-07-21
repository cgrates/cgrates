// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package utils

import (
	"strings"
)

// PrefixSliceItems iterates through slice and add a prefix before every element
func PrefixSliceItems(prfx string, slc []string) (out []string) {
	out = make([]string, 0, len(slc))
	for _, itm := range slc {
		if itm != EmptyString {
			out = append(out, prfx+itm)
		}
	}
	return
}

// SliceStringToIface converts slice of strings into a slice of interfaces
func SliceStringToIface(slc []string) (ifc []any) {
	ifc = make([]any, len(slc))
	for i, itm := range slc {
		ifc[i] = itm
	}
	return
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
