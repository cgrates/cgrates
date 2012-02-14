package timespans

import (
	"fmt"
	"log"
	"strings"
	"time"
)

/*
The input stucture that contains call information.
*/
type CallDescriptor struct {
	TOR                                int
	CstmId, Subject, DestinationPrefix string
	TimeStart, TimeEnd                 time.Time
	ActivationPeriods                  []*ActivationPeriod
}

/*
Adds an activation period that applyes to current call descriptor.
*/
func (cd *CallDescriptor) AddActivationPeriod(aps ...*ActivationPeriod) {
	for _, ap := range aps {
		cd.ActivationPeriods = append(cd.ActivationPeriods, ap)
	}
}

/*
Creates a string ready for storage containing the serialization of all
activation periods held in the internal list.
*/
func (cd *CallDescriptor) EncodeValues() (result string) {
	for _, ap := range cd.ActivationPeriods {
		result += ap.store() + "\n"
	}
	return
}

/*
Constructs the key for the storage lookup.
The prefixLen is limiting the length of the destination prefix.
*/
func (cd *CallDescriptor) GetKey() string {
	return fmt.Sprintf("%s:%s:%s", cd.CstmId, cd.Subject, cd.DestinationPrefix)
}

/*
Splits the call timespan into sub time spans accordin to the activation periods intervals.
*/
func (cd *CallDescriptor) splitInTimeSpans() (timespans []*TimeSpan) {
	if len(cd.ActivationPeriods) == 0 {
		log.Print("Nothing to split, move along... ", cd)
		return
	}
	ts1 := &TimeSpan{TimeStart: cd.TimeStart, TimeEnd: cd.TimeEnd}
	ts1.ActivationPeriod = cd.ActivationPeriods[0]

	// split on activation periods
	timespans = append(timespans, ts1)
	afterStart, afterEnd := false, false //optimization for multiple activation periods
	for _, ap := range cd.ActivationPeriods {
		if !afterStart && !afterEnd && ap.ActivationTime.Before(cd.TimeStart) {
			ts1.ActivationPeriod = ap
		} else {
			afterStart = true
			for i := 0; i < len(timespans); i++ {
				newTs := timespans[i].SplitByActivationPeriod(ap)
				if newTs != nil {
					timespans = append(timespans, newTs)
				} else {
					afterEnd = true
					break
				}
			}
		}
	}
	// split on price intervals
	for i := 0; i < len(timespans); i++ {
		ap := timespans[i].ActivationPeriod
		//timespans[i].ActivationPeriod = nil
		for _, interval := range ap.Intervals {
			newTs := timespans[i].SplitByInterval(interval)
			if newTs != nil {
				newTs.ActivationPeriod = ap
				timespans = append(timespans, newTs)
			}
		}
	}
	return
}

func (cd *CallDescriptor) RestoreFromStorage(sg StorageGetter) (destPrefix string, err error) {
	cd.ActivationPeriods = make([]*ActivationPeriod, 0)
	base := fmt.Sprintf("%s:%s:", cd.CstmId, cd.Subject)
	destPrefix = cd.DestinationPrefix
	key := base + destPrefix
	values, err := sg.Get(key)
	for i := len(cd.DestinationPrefix); err != nil && i > 1; values, err = sg.Get(key) {
		i--
		destPrefix = cd.DestinationPrefix[:i]
		key = base + destPrefix
	}
	if err == nil {
		for _, aps := range strings.Split(values, "\n") {
			if len(aps) > 0 {
				ap := &ActivationPeriod{}
				ap.restore(aps)
				cd.ActivationPeriods = append(cd.ActivationPeriods, ap)
			}
		}
	}
	return
}

/*
Creates a CallCost structure with the cost nformation calculated for the received CallDescriptor.
*/
func (cd *CallDescriptor) GetCost(sg StorageGetter) (result *CallCost, err error) {
	destPrefix, err := cd.RestoreFromStorage(sg)

	timespans := cd.splitInTimeSpans()

	cost := 0.0
	for _, ts := range timespans {
		cost += ts.GetCost()
	}

	connectionFee := 0.0
	if len(timespans) > 0 {
		connectionFee = timespans[0].Interval.ConnectFee
	}
	
	cc := &CallCost{TOR: cd.TOR,
		CstmId:            cd.CstmId,
		Subject:           cd.Subject,
		DestinationPrefix: destPrefix,
		Cost:              cost,
		ConnectFee:        connectionFee,
		Timespans:         timespans}
	return cc, err
}

/*
The output structure that will be returned with the call cost information.
*/
type CallCost struct {
	TOR                                int
	CstmId, Subject, DestinationPrefix string
	Cost, ConnectFee                   float64
	Timespans                          []*TimeSpan
}
