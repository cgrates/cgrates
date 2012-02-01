package timeslots

import (
	"time"
)

/*
A structure containing the time intervals with the cost information and the 
ActivationTime when those intervals will be applied.
*/
type ActivationPeriod struct {
	ActivationTime time.Time
	Interval []*Interval
}

func (c *ActivationPeriod) AddInterval(is ...*Interval) {
	for _, i := range is {
		c.Interval = append(c.Interval, i)
	}
}

/*
A structure that contains the data extracted from the storage.
The CstmId and the Destination prefix represent the key and the 
ActivationPeriods slice is the value.
*/
type Customer struct {
	CstmId string
	DestinationPrefix string
	ActivationPeriods []*ActivationPeriod
}

func (c *Customer) AddActivationPeriod(ap ...*ActivationPeriod) {
	for _,a := range ap {
		c.ActivationPeriods = append(c.ActivationPeriods, a)
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

