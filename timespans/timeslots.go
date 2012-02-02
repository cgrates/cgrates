package timeslots

import (
	"time"
	"fmt"
	"log"
	"encoding/json"
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
If the first parameter is true then it adds the connection fee to the cost.
*/
func (ts *TimeSpan) GetCost(first bool) return (cost float32) {
	if ts.Interval.BillingUnit > 0 {
		cost = (ts.GetDuration().seconds() / ts.Interval.BillingUnit) * ts.Interval.Price
	} else {
		cost = ts.GetDuration().seconds() * ts.Interval.Price
	}
	if first {
		cost += ts.Interval.ConnectFee
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

/*
A structure containing the time intervals with the cost information and the 
ActivationTime when those intervals will be applied.
*/
type ActivationPeriod struct {
	ActivationTime time.Time
	Intervals []*Interval
}

func (c *ActivationPeriod) AddInterval(is ...*Interval) {
	for _, i := range is {
		c.Intervals = append(c.Intervals, i)
	}
}

/*
The input stucture that contains call information.
*/
type CallDescription struct {
	TOR int
	CstmId, Subject, Destination string
	TimeStart, TimeEnd time.Time
	ActivationPeriods []*ActivationPeriod
}

/*
Adds an activation period to the internal slice
*/
func (c *Customer) addActivationPeriod(ap ...*ActivationPeriod) {
	for _,a := range ap {
		c.ActivationPeriods = append(c.ActivationPeriods, a)
	}
}

func (c *Customer) getKey() string {	
	return fmt.Sprintf("%s%s%s", c.CstmId, c.Subject, c.DestinationPrefix)	
}

func (c *Customer) encodeValue() []byte {
	jo, err := json.Marshal(c.ActivationPeriods)
	if err != nil {
		log.Print("Cannot encode intervals: ", err)
	}
	return jo
}

func (c *Customer) decodeValue(v []byte) {
	err := json.Unmarshal(v, &c.ActivationPeriods)
	if err != nil {
		log.Print("Cannot decode intervals: ", err)
	}	
}

/*
*/
func (cd *CallDescription) GetCost() (result *CallCost) {
	ts := &TimeSpan{TimeStart: cd.TimeStart, TimeEnd: cd.TimeEnd}
	c := &Customer{CstmId:, Subject:, DestinationPrefix: }
}



/*
The output structure that will be returned with the call cost information.
*/
type CallCost struct {
	TOR int
	CstmId, Subject, Prefix string
	Cost, ConnectFee float32
//	ratesInfo *RatingProfile
}


