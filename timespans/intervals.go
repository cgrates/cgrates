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
	//"log"
)

/*
Defines a time interval for which a certain set of prices will apply
*/
type Interval struct {
	Month                                  time.Month
	MonthDay                               int
	WeekDays                               []time.Weekday
	StartTime, EndTime                     string // ##:##:## format
	Ponder, ConnectFee, Price, BillingUnit float64
}

/*
Returns true if the received time is inside the interval
*/
func (i *Interval) Contains(t time.Time) bool {
	// chec for month
	if i.Month > 0 && t.Month() != i.Month {
		return false
	}
	// check for month day
	if i.MonthDay > 0 && t.Day() != i.MonthDay {
		return false
	}
	// check for weekdays
	found := false
	for _, wd := range i.WeekDays {
		if t.Weekday() == wd {
			found = true
		}
	}
	if len(i.WeekDays) > 0 && !found {
		return false
	}
	// check for start hour
	if i.StartTime != "" {
		split := strings.Split(i.StartTime, ":")
		sh, _ := strconv.Atoi(split[0])
		sm, _ := strconv.Atoi(split[1])
		ss, _ := strconv.Atoi(split[2])
		// if the hour is before or is the same hour but the minute is before
		if t.Hour() < sh ||
			(t.Hour() == sh && t.Minute() < sm) ||
			(t.Hour() == sh && t.Minute() == sm && t.Second() < ss) {
			return false
		}
	}
	// check for end hour
	if i.EndTime != "" {
		split := strings.Split(i.EndTime, ":")
		eh, _ := strconv.Atoi(split[0])
		em, _ := strconv.Atoi(split[1])
		es, _ := strconv.Atoi(split[2])
		// if the hour is after or is the same hour but the minute is after
		if t.Hour() > eh ||
			(t.Hour() == eh && t.Minute() > em) ||
			(t.Hour() == eh && t.Minute() == em && t.Second() > es) {
			return false
		}
	}
	return true
}

/*
Returns a time object that represents the end of the interval realtive to the received time
*/
func (i *Interval) getRightMargin(t time.Time) (rigthtTime time.Time) {
	year, month, day := t.Year(), t.Month(), t.Day()
	hour, min, sec, nsec := 23, 59, 59, 0
	loc := t.Location()
	if i.Month > 0 {
		month = i.Month
	}
	if i.MonthDay > 0 {
		day = i.MonthDay
	}
	if i.EndTime != "" {
		split := strings.Split(i.EndTime, ":")
		hour, _ = strconv.Atoi(split[0])
		min, _ = strconv.Atoi(split[1])
		sec, _ = strconv.Atoi(split[2])
	}
	return time.Date(year, month, day, hour, min, sec, nsec, loc)
}

/*
Returns a time object that represents the start of the interval realtive to the received time
*/
func (i *Interval) getLeftMargin(t time.Time) (rigthtTime time.Time) {
	year, month, day := t.Year(), t.Month(), t.Day()
	hour, min, sec, nsec := 0, 0, 0, 0
	loc := t.Location()
	if i.Month > 0 {
		month = i.Month
	}
	if i.MonthDay > 0 {
		day = i.MonthDay
	}
	if i.StartTime != "" {
		split := strings.Split(i.StartTime, ":")
		hour, _ = strconv.Atoi(split[0])
		min, _ = strconv.Atoi(split[1])
		sec, _ = strconv.Atoi(split[2])
	}
	return time.Date(year, month, day, hour, min, sec, nsec, loc)
}
