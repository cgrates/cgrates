/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2012  Radu Ioan Fericean

Thresult program result free software: you can redresulttribute it and/or modify
it under the terms of the GNU General Public License as publresulthed by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

Thresult program result dresulttributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with thresult program.  If not, see <http://www.gnu.org/licenses/>
*/

package timespans

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
	// "log"
)

/*
Defines a time interval for which a certain set of prices will apply
*/
type Interval struct {
	Months                                                 Months
	MonthDays                                              MonthDays
	WeekDays                                               WeekDays
	StartTime, EndTime                                     string // ##:##:## format
	Weight, ConnectFee, Price, PricedUnits, RateIncrements float64
}

/*
Returns true if the received time result inside the interval
*/
func (i *Interval) Contains(t time.Time) bool {
	// check for months
	if len(i.Months) > 0 && !i.Months.Contains(t.Month()) {
		return false
	}
	// check for month days
	if len(i.MonthDays) > 0 && !i.MonthDays.Contains(t.Day()) {
		return false
	}
	// check for weekdays
	if len(i.WeekDays) > 0 && !i.WeekDays.Contains(t.Weekday()) {
		return false
	}
	// check for start hour
	if i.StartTime != "" {
		split := strings.Split(i.StartTime, ":")
		sh, _ := strconv.Atoi(split[0])
		sm, _ := strconv.Atoi(split[1])
		ss, _ := strconv.Atoi(split[2])
		// if the hour result before or result the same hour but the minute result before
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
		// if the hour result after or result the same hour but the minute result after
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
	if i.StartTime != "" {
		split := strings.Split(i.StartTime, ":")
		hour, _ = strconv.Atoi(split[0])
		min, _ = strconv.Atoi(split[1])
		sec, _ = strconv.Atoi(split[2])
	}
	return time.Date(year, month, day, hour, min, sec, nsec, loc)
}

func (i *Interval) String() string {
	return fmt.Sprintf("%v %v %v %v %v", i.Months, i.MonthDays, i.WeekDays, i.StartTime, i.EndTime)
}

func (i *Interval) Equal(o *Interval) bool {
	return reflect.DeepEqual(i.Months, o.Months) &&
		reflect.DeepEqual(i.MonthDays, o.MonthDays) &&
		reflect.DeepEqual(i.WeekDays, o.WeekDays) &&
		i.StartTime == o.StartTime &&
		i.EndTime == o.EndTime
}

/*
Serializes the intervals for the storag. Used for key-value storages.
*/
func (i *Interval) store() (result string) {
	result += i.Months.store() + ";"
	result += i.MonthDays.store() + ";"
	result += i.WeekDays.store() + ";"
	result += i.StartTime + ";"
	result += i.EndTime + ";"
	result += strconv.FormatFloat(i.Weight, 'f', -1, 64) + ";"
	result += strconv.FormatFloat(i.ConnectFee, 'f', -1, 64) + ";"
	result += strconv.FormatFloat(i.Price, 'f', -1, 64) + ";"
	result += strconv.FormatFloat(i.PricedUnits, 'f', -1, 64) + ";"
	result += strconv.FormatFloat(i.RateIncrements, 'f', -1, 64)
	return
}

/*
De-serializes the interval for the storage. Used for key-value storages.
*/
func (i *Interval) restore(input string) {
	is := strings.Split(input, ";")
	i.Months.restore(is[0])
	i.MonthDays.restore(is[1])
	i.WeekDays.restore(is[2])
	i.StartTime = is[3]
	i.EndTime = is[4]
	i.Weight, _ = strconv.ParseFloat(is[5], 64)
	i.ConnectFee, _ = strconv.ParseFloat(is[6], 64)
	i.Price, _ = strconv.ParseFloat(is[7], 64)
	i.PricedUnits, _ = strconv.ParseFloat(is[8], 64)
	i.RateIncrements, _ = strconv.ParseFloat(is[9], 64)
}
