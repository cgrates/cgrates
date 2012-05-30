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
	"time"
	"strings"
	"strconv"
	"log"
)

// Defines months series
type Months []time.Month

// Return true if the specified date is inside the series
func (m Months) Contains(month time.Month) (result bool) {
	result = false
	for _, ms := range m {
		if ms == month {
			result = true
			break
		}
	}
	return
}

/*
Serializes the month for the storage. Used for key-value storages.
*/
func (m Months) store() (result string) {
	for _, ms := range m {
		result += strconv.Itoa(int(ms)) + ","
	}
	return
}

/*
De-serializes the month for the storage. Used for key-value storages.
*/
func (m *Months) restore(input string) {
	elements := strings.Split(input, ",")
	for _, ms := range elements {
		if month, err := strconv.Atoi(ms); err == nil {
			*m = append(*m, time.Month(month))
		}
	}
	log.Print("here: ", m)
}

// Defines month days series
type MonthDays []int

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

/*
Serializes the month days for the storage. Used for key-value storages.
*/
func (md MonthDays) store() (result string) {
	for _, mds := range md {
		result += strconv.Itoa(mds) + ","
	}
	return
}

/*
De-serializes the month days for the storage. Used for key-value storages.
*/
func (md *MonthDays) restore(input string) {
	elements := strings.Split(input, ",")
	for _, mds := range elements {
		if day, err := strconv.Atoi(mds); err == nil {
			*md = append(*md, day)
		}
	}
}

// Defines week days series
type WeekDays []time.Weekday

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

/*
Serializes the week days for the storage. Used for key-value storages.
*/
func (wd WeekDays) store() (result string) {
	for _, wds := range wd {
		result += strconv.Itoa(int(wds)) + ","
	}
	return
}

/*
De-serializes the week days for the storage. Used for key-value storages.
*/
func (wd *WeekDays) restore(input string) {
	elements := strings.Split(input, ",")
	for _, wds := range elements {
		if day, err := strconv.Atoi(wds); err == nil {
			*wd = append(*wd, time.Weekday(day))
		}
	}
}
