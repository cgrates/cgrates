package timeslots

import (
	"time"
)

type RatingProfile struct {
	StartTime time.Duration
	ConnectFee, Price, BillingUnit float32
}

type ActivationPeriod struct {
	ActivationTime time.Time
	RatingProfiles []*RatingProfile
}

func (cd *ActivationPeriod) SplitInTimeSpans(t []*ActivationPeriods) (timespans []*TimeSpans) {
	for i, ap := range aps {
		t := &TimeSpan{TimeStart: cd.TimeStart, TimeEnd: cdTimeEnd}
		timespans = append(timespans, aps.SplitInTimeSlots(t))
	}
	return
}

func (c *ActivationPeriod) AddRatingProfile(rp ...*RatingProfile) {
	for _, r := range rp {
		c.RatingProfiles = append(c.RatingProfiles, r)
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

func (c *Customer) SplitInTimeSpans(cd *CallDescription) (timespans []*TimeSpans) {
	t := &TimeSpan{TimeStart: cd.TimeStart, TimeEnd: cdTimeEnd}
	for i, ap := range c.ActivationPeriods {
		timespans = append(timespans, aps.SplitInTimeSlots(t))
	}
	return
}

func (c *Customer) CleanOldActivationPeriods() {
	now := time.Now()
	obsoleteIndex := -1
	for i, ap := range c.ActivationPeriods {
		if i > len(c.ActivationPeriods) - 2 {
			break
		}
		if a
	}
}


type CallDescription struct {
	TOR int
	CstmId, Subject, Destination string
	TimeStart, TimeEnd time.Time
}

type TimeSpan struct {
	TimeStart, TimeEnd time.Time
	RatingProfile *RatingProfile
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

