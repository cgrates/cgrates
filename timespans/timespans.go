package timespans

import (
	"time"
	//"log"
)

/*
A unit in which a call will be split that has a specific price related interval attached to it.
*/
type TimeSpan struct {
	TimeStart, TimeEnd time.Time
	ActivationPeriod   *ActivationPeriod
	Interval           *Interval
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
	if ts.Interval == nil {
		return 0
	}
	if ts.Interval.BillingUnit > 0 {
		cost = (ts.GetDuration().Seconds() / ts.Interval.BillingUnit) * ts.Interval.Price
	} else {
		cost = ts.GetDuration().Seconds() * ts.Interval.Price
	}
	return
}

/*
Returns true if the given time is inside timespan range.
*/
func (ts *TimeSpan) Contains(t time.Time) bool {
	return t.After(ts.TimeStart) && t.Before(ts.TimeEnd)
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

/*
Splits the given timespan according to how it relates to the interval.
It will modify the endtime of the received timespan and it will return
a new timespan starting from the end of the received one.
The interval will attach itself to the timespan that overlaps the interval.
*/
func (ts *TimeSpan) SplitByInterval(i *Interval) (nts *TimeSpan) {
	// if the span is not in interval return nil
	if !i.ContainsSpan(ts) {
		return
	}
	// if the span is enclosed in the interval try to set as new interval and return nil
	if i.ContainsFullSpan(ts) {
		ts.SetInterval(i)
		return
	}
	// if only the start time is in the interval split the interval
	if i.Contains(ts.TimeStart) {
		splitTime := i.getRightMargin(ts.TimeStart)
		ts.SetInterval(i)
		if splitTime == ts.TimeStart {
			return
		}
		nts = &TimeSpan{TimeStart: splitTime, TimeEnd: ts.TimeEnd}
		ts.TimeEnd = splitTime

		return
	}
	// if only the end time is in the interval split the interval
	if i.Contains(ts.TimeEnd) {
		splitTime := i.getLeftMargin(ts.TimeEnd)
		if splitTime == ts.TimeEnd {
			return
		}
		nts = &TimeSpan{TimeStart: splitTime, TimeEnd: ts.TimeEnd}
		ts.TimeEnd = splitTime

		nts.SetInterval(i)
		return
	}
	return
}

/*
Splits the given timespan on activation period's activation time.
*/
func (ts *TimeSpan) SplitByActivationPeriod(ap *ActivationPeriod) (newTs *TimeSpan) {
	if !ts.Contains(ap.ActivationTime) {
		return nil
	}
	newTs = &TimeSpan{TimeStart: ap.ActivationTime, TimeEnd: ts.TimeEnd, ActivationPeriod: ap}
	ts.TimeEnd = ap.ActivationTime
	return
}
