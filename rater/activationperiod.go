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

package rater

import (
	"github.com/cgrates/cgrates/cache2go"
	"time"
)

/*
The struture that is saved to storage.
*/
type ActivationPeriod struct {
	ActivationTime time.Time
	Intervals      IntervalList
}

type xCachedActivationPeriods struct {
	destPrefix string
	aps        []*ActivationPeriod
	*cache2go.XEntry
}

/*
Adds one ore more intervals to the internal interval list only if it is not allready in the list.
*/
func (ap *ActivationPeriod) AddInterval(is ...*Interval) {
	for _, i := range is {
		found := false
		for _, ei := range ap.Intervals {
			if i.Equal(ei) {
				(&ei.Prices).AddPrice(i.Prices...)
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
