package timespans

import ("time"; "encoding/json"; "fmt"; "log")

/*
The output structure that will be returned with the call cost information.
*/
type CallCost struct {
	TOR int
	CstmId, Subject, DestinationPrefix string
	Cost, ConnectFee float64
//	ratesInfo *RatingProfile
}

type ActivationPeriod struct {
	ActivationTime time.Time
	Intervals []*Interval
}

func (ap *ActivationPeriod) AddInterval( is ...*Interval) {	
	for _, i := range is {
		ap.Intervals = append(ap.Intervals, i)
	}	
}

/*
The input stucture that contains call information.
*/
type CallDescriptor struct {
	TOR int
	CstmId, Subject, DestinationPrefix string
	TimeStart, TimeEnd time.Time
	ActivationPeriods []*ActivationPeriod
}

func (cd *CallDescriptor) AddActivationPeriod(aps ...*ActivationPeriod) {	
	for _, ap := range aps {
		cd.ActivationPeriods = append(cd.ActivationPeriods, ap)
	}	
}

func (cd *CallDescriptor) EncodeValues() []byte {
	jo, err := json.Marshal(cd.ActivationPeriods)
	if err != nil {
		log.Print("Cannot encode intervals: ", err)
	}
	return jo
}

func (cd *CallDescriptor) GetKey() string {
	return fmt.Sprintf("%s:%s:%s", cd.CstmId, cd.Subject, cd.DestinationPrefix)
}

func (cd *CallDescriptor) decodeValues(v []byte) {
	err := json.Unmarshal(v, &cd.ActivationPeriods)
	if err != nil {
		log.Print("Cannot decode intervals: ", err)
	}	
}

func (cd *CallDescriptor) getActiveIntervals() (is []*Interval) {	
	now := time.Now()
	// add a second in the future to be able to pick the active timestamp
	// from the very second it becomes active
	sec,_ := time.ParseDuration("1s")
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
	
	cd.decodeValues([]byte(values))
	
	intervals := cd.getActiveIntervals()
	timespans := cd.splitInTimeSpans(intervals)

	cost := 0.0
	for _, ts := range timespans {
		cost += ts.GetCost()
	}
	cc := &CallCost{TOR: cd.TOR,
			CstmId: cd.CstmId,
			Subject: cd.Subject,
			DestinationPrefix: cd.DestinationPrefix,
			Cost: cost,
			ConnectFee: timespans[0].Interval.ConnectFee}
	return cc, err
}

