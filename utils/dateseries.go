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
	"fmt"
	"reflect"
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

// Return true if the specified date is inside the series
func (ys Years) Contains(year int) (result bool) {
	result = false
	for _, yss := range ys {
		if yss == year {
			result = true
			break
		}
	}
	return
}

// Parse Years elements from string separated by sep.
func (ys *Years) Parse(input, sep string) {
	switch input {
	case "*any", "":
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

func (ys Years) Serialize(sep string) string {
	if len(ys) == 0 {
		return "*any"
	}
	var yStr string
	for idx, yr := range ys {
		if idx != 0 {
			yStr = fmt.Sprintf("%s%s%d", yStr, sep, yr)
		} else {
			yStr = strconv.Itoa(yr)
		}
	}
	return yStr
}

// Equals implies that Years are already sorted
func (ys Years) Equals(oYS Years) bool {
	if len(ys) != len(oYS) {
		return false
	}
	for i := range ys {
		if ys[i] != oYS[i] {
			return false
		}
	}
	return true
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
	case "*any", "": // Apier cannot receive empty string, hence using meta-tag
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

// Dumps the months in a serialized string, similar to the one parsed
func (m Months) Serialize(sep string) string {
	if len(m) == 0 {
		return "*any"
	}
	var mStr string
	for idx, mt := range m {
		if idx != 0 {
			mStr = fmt.Sprintf("%s%s%d", mStr, sep, mt)
		} else {
			mStr = strconv.Itoa(int(mt))
		}
	}
	return mStr
}

func (m Months) IsComplete() bool {
	allMonths := Months{time.January, time.February, time.March, time.April, time.May, time.June, time.July, time.August, time.September, time.October, time.November, time.December}
	m.Sort()
	return reflect.DeepEqual(m, allMonths)
}

// Equals implies that Months are already sorted
func (m Months) Equals(oM Months) bool {
	if len(m) != len(oM) {
		return false
	}
	for i := range m {
		if m[i] != oM[i] {
			return false
		}
	}
	return true
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
	case "*any", "":
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

// Dumps the month days in a serialized string, similar to the one parsed
func (md MonthDays) Serialize(sep string) string {
	if len(md) == 0 {
		return "*any"
	}
	var mdsStr string
	for idx, mDay := range md {
		if idx != 0 {
			mdsStr = fmt.Sprintf("%s%s%d", mdsStr, sep, mDay)
		} else {
			mdsStr = strconv.Itoa(mDay)
		}
	}
	return mdsStr
}

// Equals implies that MonthDays are already sorted
func (md MonthDays) Equals(oMD MonthDays) bool {
	if len(md) != len(oMD) {
		return false
	}
	for i := range md {
		if md[i] != oMD[i] {
			return false
		}
	}
	return true
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
	case "*any", "":
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

// Dumps the week days in a serialized string, similar to the one parsed
func (wd WeekDays) Serialize(sep string) string {
	if len(wd) == 0 {
		return "*any"
	}
	var wdStr string
	for idx, d := range wd {
		if idx != 0 {
			wdStr = fmt.Sprintf("%s%s%d", wdStr, sep, d)
		} else {
			wdStr = strconv.Itoa(int(d))
		}
	}
	return wdStr
}

// Equals implies that WeekDays are already sorted
func (wd WeekDays) Equals(oWD WeekDays) bool {
	if len(wd) != len(oWD) {
		return false
	}
	for i := range wd {
		if wd[i] != oWD[i] {
			return false
		}
	}
	return true
}

func DaysInMonth(year int, month time.Month) float64 {
	return float64(time.Date(year, month, 1, 0, 0, 0, 0, time.UTC).AddDate(0, 1, -1).Day())
}

func DaysInYear(year int) float64 {
	first := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
	last := first.AddDate(1, 0, 0)
	return float64(last.Sub(first).Hours() / 24)
}

type LocalAddr struct{}

func (lc *LocalAddr) Network() string {
	return Local
}

func (lc *LocalAddr) String() string {
	return Local
}
