package timespans

import (
	"time"
	"strings"
	"strconv"	
)

/*
Defines a time interval for which a certain set of prices will apply
*/
type Interval struct {
	Month time.Month
	MonthDay int
	WeekDays []time.Weekday
	StartHour, EndHour string // ##:## format	 
	Ponder, ConnectFee, Price, BillingUnit float64
}

/*
Returns true if the received time is inside the interval
*/
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

/*
Returns true if the timespan has at list one margin in the interval
*/
func (i *Interval) ContainsSpan(t *TimeSpan) bool {
	return i.Contains(t.TimeStart) || i.Contains(t.TimeEnd)
}

/*
Returns true if the timespan is fully enclosed in the interval
*/
func (i *Interval) ContainsFullSpan(t *TimeSpan) bool {
	return i.Contains(t.TimeStart) && i.Contains(t.TimeEnd)
}

/*
Returns a time object that represents the end of the interval realtive to the received time
*/
func (i *Interval) getRightMargin(t time.Time) (rigthtTime time.Time){
	year, month, day := t.Year(), t.Month(), t.Day() 
	hour, min, sec, nsec := 23,59,59,0
	loc := t.Location()	 
	if i.Month > 0 { month = i.Month }
	if i.MonthDay > 0 { day = i.MonthDay }
	if i.EndHour != "" {
		split := strings.Split(i.EndHour, ":")
		hour, _ = strconv.Atoi(split[0])
		min, _ = strconv.Atoi(split[1])
		sec = 0
	}
	return time.Date(year, month, day, hour, min, sec, nsec, loc)
}

/*
Returns a time object that represents the start of the interval realtive to the received time
*/
func (i *Interval) getLeftMargin(t time.Time) (rigthtTime time.Time){
	year, month, day := t.Year(), t.Month(), t.Day() 
	hour, min, sec, nsec := 0,0,0,0
	loc := t.Location()	 
	if i.Month > 0 { month = i.Month }
	if i.MonthDay > 0 { day = i.MonthDay }
	if i.StartHour != "" {
		split := strings.Split(i.StartHour, ":")
		hour, _ = strconv.Atoi(split[0])
		min, _ = strconv.Atoi(split[1])
		sec = 0
	}
	return time.Date(year, month, day, hour, min, sec, nsec, loc)
}

/*
Splits the given timespan according to how it relates to the interval.
It will modify the endtime of the received timespan and it will return
a new timespan starting from the end of the received one.
The interval will attach itself to the timespan that overlaps the interval.
*/
func (i *Interval) Split(ts *TimeSpan) (nts *TimeSpan) {
	// if the span is not in interval return nil
	if !i.ContainsSpan(ts) {		
		return
	}
	// if the span is enclosed in the interval try to set as new interval and return nil
	if i.ContainsFullSpan(ts){		
		ts.SetInterval(i)		
		return
	}
	// if only the start time is in the interval splitt he interval
	if i.Contains(ts.TimeStart){		
		splitTime := i.getRightMargin(ts.TimeStart)		
		oldTimeEnd := ts.TimeEnd
		ts.TimeEnd = splitTime		
		ts.SetInterval(i)		
		nts = &TimeSpan{TimeStart: splitTime, TimeEnd: oldTimeEnd}				
		return
	}
	// if only the end time is in the interval split the interval
	if i.Contains(ts.TimeEnd){			
		splitTime := i.getLeftMargin(ts.TimeEnd)
		oldTimeEnd := ts.TimeEnd
		ts.TimeEnd = splitTime		
		nts = &TimeSpan{TimeStart: splitTime, TimeEnd: oldTimeEnd}				
		nts.SetInterval(i)
		return
	}
	return 
}

