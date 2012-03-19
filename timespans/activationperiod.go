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
	//"log"
	"strconv"
	"strings"
)

/*
The struture that is saved to storage.
*/
type ActivationPeriod struct {
	ActivationTime time.Time
	Intervals      []*Interval
}

/*
Adds one ore more intervals to the internal interval list.
*/
func (ap *ActivationPeriod) AddInterval(is ...*Interval) {
	for _, i := range is {
		ap.Intervals = append(ap.Intervals, i)
	}
}

/*
Serializes the activation periods for the storage. Used for key-value storages.
*/
func (ap *ActivationPeriod) store() (result string) {
	result += strconv.FormatInt(ap.ActivationTime.UnixNano(), 10) + ";"
	for _, i := range ap.Intervals {
		var is string
		is = strconv.Itoa(int(i.Month)) + "|"
		is += strconv.Itoa(i.MonthDay) + "|"
		for _, wd := range i.WeekDays {
			is += strconv.Itoa(int(wd)) + ","
		}
		is = strings.TrimRight(is, ",") + "|"
		is += i.StartTime + "|"
		is += i.EndTime + "|"
		is += strconv.FormatFloat(i.Ponder, 'f', -1, 64) + "|"
		is += strconv.FormatFloat(i.ConnectFee, 'f', -1, 64) + "|"
		is += strconv.FormatFloat(i.Price, 'f', -1, 64) + "|"
		is += strconv.FormatFloat(i.BillingUnit, 'f', -1, 64)
		result += is + ";"
	}
	return
}

/*
De-serializes the activation periods for the storage. Used for key-value storages.
*/
func (ap *ActivationPeriod) restore(input string) {
	elements := strings.Split(input, ";")
	unixNano, _ := strconv.ParseInt(elements[0], 10, 64)
	ap.ActivationTime = time.Unix(0, unixNano).In(time.UTC)
	ap.Intervals = make([]*Interval, 0)
	for _, is := range elements[1 : len(elements)-1] {
		i := &Interval{}
		ise := strings.Split(is, "|")
		month, _ := strconv.Atoi(ise[0])
		i.Month = time.Month(month)
		i.MonthDay, _ = strconv.Atoi(ise[1])
		for _, d := range strings.Split(ise[2], ",") {
			if d != "" {
				wd, _ := strconv.Atoi(d)
				i.WeekDays = append(i.WeekDays, time.Weekday(wd))
			}
		}
		i.StartTime = ise[3]
		i.EndTime = ise[4]
		i.Ponder, _ = strconv.ParseFloat(ise[5], 64)
		i.ConnectFee, _ = strconv.ParseFloat(ise[6], 64)
		i.Price, _ = strconv.ParseFloat(ise[7], 64)
		i.BillingUnit, _ = strconv.ParseFloat(ise[8], 64)

		ap.Intervals = append(ap.Intervals, i)
	}
}
