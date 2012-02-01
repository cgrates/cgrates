package timeslots

import (
	"time"
)


type ActivationPeriod struct {
	ActivationTime time.Time
	Interval []*Interval
}

func (c *ActivationPeriod) AddInterval(is ...*Interval) {
	for _, i := range is {
		c.Interval = append(c.Interval, i)
	}
}

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

type CallDescription struct {
	TOR int
	CstmId, Subject, Destination string
	TimeStart, TimeEnd time.Time
}

type TimeSpan struct {
	TimeStart, TimeEnd time.Time
	Interval *Interval
}

func (ts *TimeSpan) GetDuration() time.Duration {
	return ts.TimeEnd.Sub(ts.TimeStart)
}


type CallCost struct {
	TOR int
	CstmId, Subject, Prefix string
	Cost, ConnectFee float32
//	ratesInfo *RatingProfile
}

func GetCost(in *CallDescription, sg StorageGetter) (result *CallCost, err error) {
	return &CallCost{TOR: 1, CstmId:"",Subject:"", Prefix:"", Cost:1, ConnectFee:1}, nil
}

