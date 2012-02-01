package timeslots

import (
	"time"
	"strings"
	"strconv"	
)

type RatingProfile struct {
	StartTime time.Duration
	ConnectFee, Price, BillingUnit float32
}

type Interval struct {
	Month time.Month
	MonthDay int
	WeekDays []time.Weekday
	StartHour, EndHour string // ##:## format
}

func (i *Interval) Contains(t time.Time) bool {
	// chec for month
	if i.Month > 0 && t.Month() != i.Month {		
		return false
	}	
	// check for month day
	if i.MonthDay > 0 && t.Day() != i.MonthDay {		
		return false
	}
	// check for weekdays
	found := false
	for _,wd := range i.WeekDays {
		if t.Weekday() == wd {
			found = true
		}		
	}	
	if len(i.WeekDays) > 0 && !found {
		return false
	}
	// check for start hour
	if i.StartHour != ""{
		split:= strings.Split(i.StartHour, ":")
		sh, _ := strconv.Atoi(split[0])
		sm, _ := strconv.Atoi(split[1])
		// if the hour is before or is the same hour but the minute is before 
		if t.Hour() < sh || (t.Hour() == sh && t.Minute() < sm) { 			
			return false
		}
	}
	// check for end hour
	if i.EndHour != ""{
		split := strings.Split(i.EndHour, ":")
		eh, _ := strconv.Atoi(split[0])
		em, _ := strconv.Atoi(split[1])
		// if the hour is after or is the same hour but the minute is after 
		if t.Hour() > eh || (t.Hour() == eh && t.Minute() > em) { 			
			return false
		}
	}
	return true
}

func (i *Interval) ContainsSpan(t *TimeSpan) bool {
	return i.Contains(t.TimeStart) || i.Contains(t.TimeEnd)
}

func (i *Interval) ContainsFullSpan(t *TimeSpan) bool {
	return i.Contains(t.TimeStart) && i.Contains(t.TimeEnd)
}

func (i *Interval) Split(t *TimeSpan) (spans []*TimeSpan) {
	if !i.ContainsSpan(t) {
		return
	}
	if !i.ContainsFullSpan(t){
		spans = append(spans, t)
	}
	if !i.Contains(t.TimeStart){

		spans = append(spans, t)
	}
	return 
}

type ActivationPeriod struct {
	ActivationTime time.Time
	RatingProfiles []*RatingProfile
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

func (c *Customer) CleanOldActivationPeriods() {
	/*now := time.Now()
	obsoleteIndex := -1
	for i, ap := range c.ActivationPeriods {
		if i > len(c.ActivationPeriods) - 2 {
			break
		}
		//if a
	}*/
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

