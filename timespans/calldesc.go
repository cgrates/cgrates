package timespans

import (
	"fmt"
	"time"
	"strings"
	"strconv"
)

/*
The struture that is saved to storage.
*/
type ActivationPeriod struct {
	ActivationTime time.Time
	Intervals      []*Interval
}

func (ap *ActivationPeriod) AddInterval(is ...*Interval) {
	for _, i := range is {
		ap.Intervals = append(ap.Intervals, i)
	}
}

func (ap *ActivationPeriod) store() (result string){	
	result += strconv.FormatInt(ap.ActivationTime.Unix(), 10) + ";"
	var is string
	for _,i := range ap.Intervals {
		is = strconv.Itoa(int(i.Month)) + "|"
		is += strconv.Itoa(i.MonthDay) + "|"
		for _, wd := range i.WeekDays {
			is += strconv.Itoa(int(wd)) + ","
		}
		is = strings.TrimRight(is, ",")  + "|"
		is += i.StartTime + "|"
		is += i.EndTime + "|"
		is += strconv.FormatFloat(i.Ponder, 'f', -1, 64) + "|"
		is += strconv.FormatFloat(i.ConnectFee, 'f', -1, 64) + "|"
		is += strconv.FormatFloat(i.Price, 'f', -1, 64) + "|"
		is += strconv.FormatFloat(i.BillingUnit, 'f', -1, 64)
		result += is + ";"
	}	
	return 
}

func (ap *ActivationPeriod) restore(input string) {			
	elements := strings.Split(input, ";")	
	unix, _ := strconv.ParseInt(elements[0], 0, 64)		
	ap.ActivationTime = time.Unix(unix, 0)			
	for _, is := range elements[1:len(elements) - 1]{		
		i := &Interval{}
		ise := strings.Split(is, "|")		
		month, _ := strconv.Atoi(ise[0])
		i.Month = time.Month(month)
		i.MonthDay, _ = strconv.Atoi(ise[1])		 
		for _,d := range strings.Split(ise[2], ","){
			wd,_ :=  strconv.Atoi(d)
			i.WeekDays = append(i.WeekDays, time.Weekday(wd))
		}
		i.StartTime = ise[3]
		i.EndTime = ise[4]		
		i.Ponder, _ = strconv.ParseFloat(ise[5], 64)		 
		i.ConnectFee, _ = strconv.ParseFloat(ise[6], 64)
		i.Price, _ = strconv.ParseFloat(ise[7], 64)
		i.BillingUnit, _ = strconv.ParseFloat(ise[8], 64)
		
		ap.Intervals = append(ap.Intervals, i)
	}	
}

/*
The input stucture that contains call information.
*/
type CallDescriptor struct {
	TOR                                int
	CstmId, Subject, DestinationPrefix string
	TimeStart, TimeEnd                 time.Time
	ActivationPeriods                  []*ActivationPeriod
}

func (cd *CallDescriptor) AddActivationPeriod(aps ...*ActivationPeriod) {
	for _, ap := range aps {
		cd.ActivationPeriods = append(cd.ActivationPeriods, ap)
	}
}

func (cd *CallDescriptor) EncodeValues() (result string) {
	for _, ap := range cd.ActivationPeriods {
		result += ap.store() + "\n"
	}
	return 
}

func (cd *CallDescriptor) decodeValues(v string) {
	for _, aps := range strings.Split(v, "\n") {
		if(len(aps)>0){
			ap := &ActivationPeriod{}
			ap.restore(aps)
			cd.ActivationPeriods = append(cd.ActivationPeriods, ap)
		}
	}
}

func (cd *CallDescriptor) GetKey() string {
	return fmt.Sprintf("%s:%s:%s", cd.CstmId, cd.Subject, cd.DestinationPrefix)
}

func (cd *CallDescriptor) getActiveIntervals() (is []*Interval) {
	now := time.Now()
	// add a second in the future to be able to pick the active timestamp
	// from the very second it becomes active
	sec, _ := time.ParseDuration("1s")
	now.Add(sec)
	bestTime := time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC)
	for _, ap := range cd.ActivationPeriods {
		t := ap.ActivationTime
		if t.After(bestTime) && t.Before(now) {
			bestTime = t
			is = ap.Intervals
		}
	}
	return
}

func (cd *CallDescriptor) splitInTimeSpans(intervals []*Interval) (timespans []*TimeSpan) {
	ts1 := &TimeSpan{TimeStart: cd.TimeStart, TimeEnd: cd.TimeEnd}
	timespans = append(timespans, ts1)
	for _, interval := range intervals {
		for _, ts := range timespans {
			newTs := interval.Split(ts)
			if newTs != nil {
				timespans = append(timespans, newTs)
				break
			}
		}
	}
	return
}

/*
 */
func (cd *CallDescriptor) GetCost(sg StorageGetter) (result *CallCost, err error) {

	key := cd.GetKey()
	values, err := sg.Get(key)

	cd.decodeValues(values)

	intervals := cd.getActiveIntervals()
	timespans := cd.splitInTimeSpans(intervals)

	cost := 0.0
	for _, ts := range timespans {
		cost += ts.GetCost()
	}
	cc := &CallCost{TOR: cd.TOR,
		CstmId:            cd.CstmId,
		Subject:           cd.Subject,
		DestinationPrefix: cd.DestinationPrefix,
		Cost:              cost,
		ConnectFee:        timespans[0].Interval.ConnectFee}
	return cc, err
}

/*
The output structure that will be returned with the call cost information.
*/
type CallCost struct {
	TOR                                int
	CstmId, Subject, DestinationPrefix string
	Cost, ConnectFee                   float64
	//	ratesInfo *RatingProfile
}