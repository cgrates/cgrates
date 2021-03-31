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
	"strconv"
	"strings"
	"time"
)

// Defines years days series
type Years []int

func (ys Years) Sort() {
	sort.Sort(ys)
}

func (ys Years) Len() int {
	return len(ys)
}

func (ys Years) Swap(i, j int) {
	ys[i], ys[j] = ys[j], ys[i]
}

func (ys Years) Less(j, i int) bool {
	return ys[j] < ys[i]
}

// Parse Years elements from string separated by sep.
func (ys *Years) Parse(input, sep string) {
	switch input {
	case MetaAny, EmptyString:
		*ys = []int{}
	default:
		elements := strings.Split(input, sep)
		for _, yss := range elements {
			if year, err := strconv.Atoi(yss); err == nil {
				*ys = append(*ys, year)
			}
		}
	}
}

// Defines months series
type Months []time.Month

func (m Months) Sort() {
	sort.Sort(m)
}

func (m Months) Len() int {
	return len(m)
}

func (m Months) Swap(i, j int) {
	m[i], m[j] = m[j], m[i]
}

func (m Months) Less(j, i int) bool {
	return m[j] < m[i]
}

// Loades Month elemnents from a string separated by sep.
func (m *Months) Parse(input, sep string) {
	switch input {
	case MetaAny, EmptyString: // Apier cannot receive empty string, hence using meta-tag
		*m = []time.Month{}
	default:
		elements := strings.Split(input, sep)
		for _, ms := range elements {
			if month, err := strconv.Atoi(ms); err == nil {
				*m = append(*m, time.Month(month))
			}
		}
	}
}

// Defines month days series
type MonthDays []int

func (md MonthDays) Sort() {
	sort.Sort(md)
}

func (md MonthDays) Len() int {
	return len(md)
}

func (md MonthDays) Swap(i, j int) {
	md[i], md[j] = md[j], md[i]
}

func (md MonthDays) Less(j, i int) bool {
	return md[j] < md[i]
}

// Parse MonthDay elements from string separated by sep.
func (md *MonthDays) Parse(input, sep string) {
	switch input {
	case MetaAny, EmptyString:
		*md = []int{}
	default:
		elements := strings.Split(input, sep)
		for _, mds := range elements {
			if day, err := strconv.Atoi(mds); err == nil {
				*md = append(*md, day)
			}
		}
	}
}

// Defines week days series
type WeekDays []time.Weekday

func (wd WeekDays) Sort() {
	sort.Sort(wd)
}

func (wd WeekDays) Len() int {
	return len(wd)
}

func (wd WeekDays) Swap(i, j int) {
	wd[i], wd[j] = wd[j], wd[i]
}

func (wd WeekDays) Less(j, i int) bool {
	return wd[j] < wd[i]
}

func (wd *WeekDays) Parse(input, sep string) {
	switch input {
	case MetaAny, EmptyString:
		*wd = []time.Weekday{}
	default:
		elements := strings.Split(input, sep)
		for _, wds := range elements {
			if day, err := strconv.Atoi(wds); err == nil {
				*wd = append(*wd, time.Weekday(day%7)) // %7 for sunday = 7 normalization
			}
		}
	}
}
