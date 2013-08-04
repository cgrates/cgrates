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

package history

import (
	"sort"
)

const (
	RATING_PROFILE_PREFIX = "rpf_"
	DESTINATION_PREFIX    = "dst_"
)

type Scribe interface {
	Record(record *Record, out *int) error
}

type Record struct {
	Key    string
	Object interface{}
}

type records []*Record

func (rs records) Len() int {
	return len(rs)
}

func (rs records) Swap(i, j int) {
	rs[i], rs[j] = rs[j], rs[i]
}

func (rs records) Less(i, j int) bool {
	return rs[i].Key < rs[j].Key
}

func (rs records) Sort() {
	sort.Sort(rs)
}

func (rs records) SetOrAdd(rec *Record) records {
	found := false
	for _, r := range rs {
		if r.Key == rec.Key {
			found = true
			r.Object = rec.Object
			return rs
		}
	}
	if !found {
		rs = append(rs, rec)
	}
	return rs
}
