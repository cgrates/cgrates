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
	"fmt"
	"reflect"
)

/*
The output structure that will be returned with the call cost information.
*/
type CallCost struct {
	Direction, TOR, Tenant, Subject, Account, Destination string
	Cost, ConnectFee                                      float64
	Timespans                                             []*TimeSpan
}

// Pretty printing for call cost
func (cc *CallCost) String() (r string) {
	r = fmt.Sprintf("%v[%v] : %s(%s) -> %s (", cc.Cost, cc.ConnectFee, cc.Subject, cc.Account, cc.Destination)
	for _, ts := range cc.Timespans {
		r += fmt.Sprintf(" %v,", ts.GetDuration())
	}
	r += " )"
	return
}

// Merges the received timespan if they are similar (same activation period, same interval, same minute info.
func (cc *CallCost) Merge(other *CallCost) {
	if len(cc.Timespans)-1 < 0 {
		return
	}
	ts := cc.Timespans[len(cc.Timespans)-1]
	otherTs := other.Timespans[0]
	if reflect.DeepEqual(ts.ActivationPeriod, otherTs.ActivationPeriod) &&
		reflect.DeepEqual(ts.MinuteInfo, otherTs.MinuteInfo) && reflect.DeepEqual(ts.Interval, otherTs.Interval) {
		// extend the last timespan with
		ts.TimeEnd = ts.TimeEnd.Add(otherTs.GetDuration())
		// add the rest of the timspans
		cc.Timespans = append(cc.Timespans, other.Timespans[1:]...)
	} else {
		// just add all timespans
		cc.Timespans = append(cc.Timespans, other.Timespans...)
	}
	cc.Cost += other.Cost
}
