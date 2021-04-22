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
)

// NewStringSet returns a new StringSet
func NewStringSet(dataSlice []string) (s StringSet) {
	s = make(StringSet)
	s.AddSlice(dataSlice)
	return s
}

// StringSet will manage data within a set
type StringSet map[string]struct{}

// Add adds a key in set
func (s StringSet) Add(val string) {
	s[val] = struct{}{}
}

// Remove removes a key from set
func (s StringSet) Remove(val string) {
	delete(s, val)
}

// Has returns if the key is in set
func (s StringSet) Has(val string) bool {
	_, has := s[val]
	return has
}

// AddSlice adds all the element of a slice
func (s StringSet) AddSlice(dataSlice []string) {
	for _, val := range dataSlice {
		s.Add(val)
	}
}

// AsSlice returns the keys as string slice
func (s StringSet) AsSlice() []string {
	if s == nil {
		return nil
	}
	result := make([]string, 0, len(s))
	for k := range s {
		result = append(result, k)
	}
	return result
}

// AsOrderedSlice returns the keys as ordered string slice
func (s StringSet) AsOrderedSlice() (ss []string) {
	ss = s.AsSlice()
	sort.Strings(ss)
	return
}

// Sha1 returns the Sha1 on top of ordered slice
func (s StringSet) Sha1() string {
	return Sha1(s.AsOrderedSlice()...)
}

// Size returns the size of the set
func (s StringSet) Size() int {
	return len(s)
}

// Intersect removes all key s2 do not have
func (s StringSet) Intersect(s2 StringSet) {
	for k := range s {
		if !s2.Has(k) {
			s.Remove(k)
		}
	}
}

// Clone creates a clone of the set
func (s StringSet) Clone() (cln StringSet) {
	if s == nil {
		return
	}
	cln = make(StringSet)
	for k := range s {
		cln.Add(k)
	}
	return
}

// GetOne returns a key from set
func (s StringSet) GetOne() string {
	for k := range s {
		return k
	}
	return EmptyString
}

// JoinStringSet intersect multiple StringSets in one
func JoinStringSet(s ...StringSet) (conc StringSet) {
	conc = make(StringSet)
	for _, k := range s {
		for key := range k {
			conc.Add(key)
		}
	}
	return
}

// Equals returns true if the set are the same
func (s StringSet) Equals(s2 StringSet) bool {
	if s.Size() != s2.Size() {
		return false
	}
	for k := range s2 {
		if !s.Has(k) {
			return false
		}
	}
	return true
}
