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
package engine

import (
	"reflect"
	"time"
)

// The output structure that will be returned with the call cost information.
type CallCost struct {
	Direction, TOR, Tenant, Subject, Account, Destination, Type string
	Cost                                                        float64
	Timespans                                                   TimeSpans
	deductConnectFee                                            bool
}

// Pretty printing for call cost
/*func (cc *CallCost) String() (r string) {
	connectFee := 0.0
	if cc.deductConnectFee {
		connectFee = cc.GetConnectFee()
	}
	r = fmt.Sprintf("%v[%v] : %s(%s) - > %s (", cc.Cost, connectFee, cc.Subject, cc.Account, cc.Destination)
	for _, ts := range cc.Timespans {
		r += fmt.Sprintf(" %v,", ts.GetDuration())
	}
	r += " )"
	return
}*/

// Merges the received timespan if they are similar (same activation period, same interval, same minute info.
func (cc *CallCost) Merge(other *CallCost) {
	if len(cc.Timespans)-1 < 0 || len(other.Timespans) == 0 {
		return
	}
	ts := cc.Timespans[len(cc.Timespans)-1]
	otherTs := other.Timespans[0]
	if reflect.DeepEqual(ts.ratingInfo, otherTs.ratingInfo) &&
		reflect.DeepEqual(ts.RateInterval, otherTs.RateInterval) {
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

func (cc *CallCost) GetStartTime() time.Time {
	if len(cc.Timespans) == 0 {
		return time.Now()
	}
	return cc.Timespans[0].TimeStart
}

func (cc *CallCost) GetEndTime() time.Time {
	if len(cc.Timespans) == 0 {
		return time.Now()
	}
	return cc.Timespans[len(cc.Timespans)-1].TimeEnd
}

func (cc *CallCost) GetDuration() (td time.Duration) {
	for _, ts := range cc.Timespans {
		td += ts.GetDuration()
	}
	return
}

func (cc *CallCost) GetConnectFee() float64 {
	if len(cc.Timespans) == 0 ||
		cc.Timespans[0].RateInterval == nil ||
		cc.Timespans[0].RateInterval.Rating == nil {
		return 0
	}
	return cc.Timespans[0].RateInterval.Rating.ConnectFee
}

// Creates a CallDescriptor structure copying related data from CallCost
func (cc *CallCost) CreateCallDescriptor() *CallDescriptor {
	return &CallDescriptor{
		Direction:   cc.Direction,
		TOR:         cc.TOR,
		Tenant:      cc.Tenant,
		Subject:     cc.Subject,
		Account:     cc.Account,
		Destination: cc.Destination,
	}
}

func (cc *CallCost) IsPaid() bool {
	for _, ts := range cc.Timespans {
		if paid, _ := ts.IsPaid(); !paid {
			return false
		}
	}
	return true
}
