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

package rater

import (
	"github.com/cgrates/cgrates/cache2go"
	"strconv"
	"strings"
	"time"
)

/*
The struture that is saved to storage.
*/
type ActivationPeriod struct {
	ActivationTime time.Time
	Intervals      []*Interval
}

type xCachedActivationPeriods struct {
	destPrefix string
	aps        []*ActivationPeriod
	*cache2go.XEntry
}

/*
Adds one ore more intervals to the internal interval list.
*/
func (ap *ActivationPeriod) AddInterval(is ...*Interval) {
	ap.Intervals = append(ap.Intervals, is...)
}

/*
Adds one ore more intervals to the internal interval list only if it is not allready in the list.
*/
func (ap *ActivationPeriod) AddIntervalIfNotPresent(is ...*Interval) {
	for _, i := range is {
		found := false
		for _, ei := range ap.Intervals {
			if i.Equal(ei) {
				found = true
				break
			}
		}
		if !found {
			ap.Intervals = append(ap.Intervals, i)
		}
	}
}

func (ap *ActivationPeriod) Equal(o *ActivationPeriod) bool {
	return ap.ActivationTime == o.ActivationTime
}

/*
Serializes the activation periods for the storage. Used for key-value storages.
*/
func (ap *ActivationPeriod) store() (result string) {
	result += strconv.FormatInt(ap.ActivationTime.UnixNano(), 10) + "|"
	for _, i := range ap.Intervals {
		result += i.store() + "|"
	}
	result = strings.TrimRight(result, "|")
	return
}

/*
De-serializes the activation periods for the storage. Used for key-value storages.
*/
func (ap *ActivationPeriod) restore(input string) {
	elements := strings.Split(input, "|")
	unixNano, _ := strconv.ParseInt(elements[0], 10, 64)
	ap.ActivationTime = time.Unix(0, unixNano).In(time.UTC)
	els := elements[1:]
	if len(els) > 1 {
		els = elements[1 : len(elements)-1]
	}
	for _, is := range els {
		i := &Interval{}
		i.restore(is)
		ap.Intervals = append(ap.Intervals, i)
	}
}
