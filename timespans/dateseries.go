/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2012  Radu Ioan Fericean

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

package timespans

import (
	"strconv"
	"strings"
	"time"
	// "log"
)

// Defines months series
type Months []time.Month

func (m Months) Len() int {
	return len(m)
}

func (m Months) Swap(i, j int) {
	m[i], m[j] = m[j], m[i]
}

func (m Months) Less(j, i int) bool {
	return m[j] < m[i]
}

// Return true if the specified date is inside the series
func (m Months) Contains(month time.Month) (result bool) {
	for _, ms := range m {
		if ms == month {
			result = true
			break
		}
	}
	return
}

// Loades Month elemnents from a string separated by sep.
func (m *Months) Parse(input, sep string) {
	switch input {
	case "*all":
		*m = []time.Month{time.January, time.February, time.March, time.April, time.May, time.June,
			time.July, time.August, time.September, time.October, time.November, time.December}
	case "*none":
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

func (md MonthDays) Len() int {
	return len(md)
}

func (md MonthDays) Swap(i, j int) {
	md[i], md[j] = md[j], md[i]
}

func (md MonthDays) Less(j, i int) bool {
	return md[j] < md[i]
}

// Return true if the specified date is inside the series
func (md MonthDays) Contains(monthDay int) (result bool) {
	result = false
	for _, mds := range md {
		if mds == monthDay {
			result = true
			break
		}
	}
	return
}

// Parse MonthDay elements from string separated by sep.
func (md *MonthDays) Parse(input, sep string) {
	switch input {
	case "*all":
		*md = []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31}
	case "*none":
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

func (wd WeekDays) Len() int {
	return len(wd)
}

func (wd WeekDays) Swap(i, j int) {
	wd[i], wd[j] = wd[j], wd[i]
}

func (wd WeekDays) Less(j, i int) bool {
	return wd[j] < wd[i]
}

// Return true if the specified date is inside the series
func (wd WeekDays) Contains(weekDay time.Weekday) (result bool) {
	result = false
	for _, wds := range wd {
		if wds == weekDay {
			result = true
			break
		}
	}
	return
}

func (wd *WeekDays) Parse(input, sep string) {
	switch input {
	case "*all":
		*wd = []time.Weekday{time.Monday, time.Tuesday, time.Wednesday, time.Thursday, time.Friday, time.Saturday, time.Sunday}
	case "*none":
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
