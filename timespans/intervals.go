package timespans

import (
	"strconv"
	"strings"
	"time"
)

/*
Defines a time interval for which a certain set of prices will apply
*/
type Interval struct {
	Month                                  time.Month
	MonthDay                               int
	WeekDays                               []time.Weekday
	StartTime, EndTime                     string // ##:##:## format	 
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
	for _, wd := range i.WeekDays {
		if t.Weekday() == wd {
			found = true
		}
	}
	if len(i.WeekDays) > 0 && !found {
		return false
	}
	// check for start hour
	if i.StartTime != "" {
		split := strings.Split(i.StartTime, ":")
		sh, _ := strconv.Atoi(split[0])
		sm, _ := strconv.Atoi(split[1])
		ss, _ := strconv.Atoi(split[2])
		// if the hour is before or is the same hour but the minute is before 
		if t.Hour() < sh ||
			(t.Hour() == sh && t.Minute() < sm) ||
			(t.Hour() == sh && t.Minute() == sm && t.Second() < ss) {
			return false
		}
	}
	// check for end hour
	if i.EndTime != "" {
		split := strings.Split(i.EndTime, ":")
		eh, _ := strconv.Atoi(split[0])
		em, _ := strconv.Atoi(split[1])
		es, _ := strconv.Atoi(split[2])
		// if the hour is after or is the same hour but the minute is after 
		if t.Hour() > eh ||
			(t.Hour() == eh && t.Minute() > em) ||
			(t.Hour() == eh && t.Minute() == em && t.Second() > es) {
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
func (i *Interval) ContainsFullSpan(ts *TimeSpan) bool {
	return i.Contains(ts.TimeStart) && i.Contains(ts.TimeEnd)
}

/*
Returns a time object that represents the end of the interval realtive to the received time
*/
func (i *Interval) getRightMargin(t time.Time) (rigthtTime time.Time) {
	year, month, day := t.Year(), t.Month(), t.Day()
	hour, min, sec, nsec := 23, 59, 59, 0
	loc := t.Location()
	if i.Month > 0 {
		month = i.Month
	}
	if i.MonthDay > 0 {
		day = i.MonthDay
	}
	if i.EndTime != "" {
		split := strings.Split(i.EndTime, ":")
		hour, _ = strconv.Atoi(split[0])
		min, _ = strconv.Atoi(split[1])
		sec, _ = strconv.Atoi(split[2])
	}
	return time.Date(year, month, day, hour, min, sec, nsec, loc)
}

/*
Returns a time object that represents the start of the interval realtive to the received time
*/
func (i *Interval) getLeftMargin(t time.Time) (rigthtTime time.Time) {
	year, month, day := t.Year(), t.Month(), t.Day()
	hour, min, sec, nsec := 0, 0, 0, 0
	loc := t.Location()
	if i.Month > 0 {
		month = i.Month
	}
	if i.MonthDay > 0 {
		day = i.MonthDay
	}
	if i.StartTime != "" {
		split := strings.Split(i.StartTime, ":")
		hour, _ = strconv.Atoi(split[0])
		min, _ = strconv.Atoi(split[1])
		sec, _ = strconv.Atoi(split[2])
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
	if i.ContainsFullSpan(ts) {
		ts.SetInterval(i)
		return
	}
	// if only the start time is in the interval split he interval
	if i.Contains(ts.TimeStart) {
		splitTime := i.getRightMargin(ts.TimeStart)
		ts.SetInterval(i)
		if splitTime == ts.TimeStart {
			return
		}
		oldTimeEnd := ts.TimeEnd
		ts.TimeEnd = splitTime

		nts = &TimeSpan{TimeStart: splitTime, TimeEnd: oldTimeEnd}
		return
	}
	// if only the end time is in the interval split the interval
	if i.Contains(ts.TimeEnd) {
		splitTime := i.getLeftMargin(ts.TimeEnd)
		if splitTime == ts.TimeEnd {
			ts.SetInterval(i)
			return
		}
		oldTimeEnd := ts.TimeEnd
		ts.TimeEnd = splitTime

		nts = &TimeSpan{TimeStart: splitTime, TimeEnd: oldTimeEnd}
		nts.SetInterval(i)
		return
	}
	return
}
