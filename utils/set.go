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

func NewStringSet(dataSlice []string) (s *StringSet) {
	s = &StringSet{data: make(map[string]struct{})}
	s.AddSlice(dataSlice)
	return s
}

// StringSet will manage data within a set
type StringSet struct {
	data map[string]struct{}
}

func (s *StringSet) Add(val string) {
	s.data[val] = struct{}{}
}

func (s *StringSet) Remove(val string) {
	delete(s.data, val)
}

func (s *StringSet) Has(val string) bool {
	_, has := s.data[val]
	return has
}

func (s *StringSet) AddSlice(dataSlice []string) {
	for _, val := range dataSlice {
		s.Add(val)
	}
}

func (s *StringSet) AsSlice() []string {
	result := make([]string, len(s.data))
	i := 0
	for k := range s.data {
		result[i] = k
		i++
	}
	return result
}
