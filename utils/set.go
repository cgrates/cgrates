// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

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
