package timeslots

import (
	"time"
	"fmt"
	"log"
	"encoding/json"
)

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
A structure that contains the data extracted from the storage.
The CstmId and the Destination prefix represent the key and the 
ActivationPeriods slice is the value.
*/
type Customer struct {
	CstmId string
	Subject string
	DestinationPrefix string
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
The input stucture that contains call information.
*/
type CallDescription struct {
	TOR int
	CstmId, Subject, Destination string
	TimeStart, TimeEnd time.Time
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

func GetCost(in *CallDescription, sg StorageGetter) (result *CallCost, err error) {
	return &CallCost{TOR: 1, CstmId:"",Subject:"", Prefix:"", Cost:1, ConnectFee:1}, nil
}

