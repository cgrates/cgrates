package timespans

import (
	"time"	
)

/*
A unit in which a call will be split that has a specific price related interval attached to it.
*/
type TimeSpan struct {
	TimeStart, TimeEnd time.Time
	Interval *Interval
}

/*
Returns the duration of the timespan
*/
func (ts *TimeSpan) GetDuration() time.Duration {
	return ts.TimeEnd.Sub(ts.TimeStart)
}

/*
Returns the cost of the timespan according to the relevant cost interval.
*/
func (ts *TimeSpan) GetCost() (cost float64) {
	if ts.Interval.BillingUnit > 0 {
		cost = (ts.GetDuration().Seconds() / ts.Interval.BillingUnit) * ts.Interval.Price
	} else {
		cost = ts.GetDuration().Seconds() * ts.Interval.Price
	}
	return
}

/*
will set ne interval as spans's interval if new ponder is greater then span's interval ponder
or if the ponders are equal and new price is lower then spans's interval price
*/
func (ts *TimeSpan) SetInterval(i *Interval) {
	if ts.Interval == nil || ts.Interval.Ponder < i.Ponder {
		ts.Interval = i
	}
	if ts.Interval.Ponder == i.Ponder && i.Price < ts.Interval.Price {
		ts.Interval = i
	}
}
