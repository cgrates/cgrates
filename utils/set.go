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

import "sort"

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
	result := make([]string, len(s))
	i := 0
	for k := range s {
		result[i] = k
		i++
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
