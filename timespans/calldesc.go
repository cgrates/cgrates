package timespans

import (
	"fmt"
	"strings"
	"time"
	//"log"
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
Finds the activation periods applicable to the call descriptior.
*/
func (cd *CallDescriptor) getActivePeriods() (is []*ActivationPeriod) {
	bestTime := cd.ActivationPeriods[0].ActivationTime
	is = append(is, cd.ActivationPeriods[0])

	for _, ap := range cd.ActivationPeriods {
		if ap.ActivationTime.After(bestTime) && ap.ActivationTime.Before(cd.TimeStart) {
			bestTime = ap.ActivationTime
			is[0] = ap
		}
		if ap.ActivationTime.After(cd.TimeStart) && ap.ActivationTime.Before(cd.TimeEnd) {
			is = append(is, ap)
		}
	}
	return
}

/*
Splits the call timespan into sub time spans accordin to the activation periods intervals.
*/
func (cd *CallDescriptor) splitInTimeSpans(aps []*ActivationPeriod) (timespans []*TimeSpan) {
	ts1 := &TimeSpan{TimeStart: cd.TimeStart, TimeEnd: cd.TimeEnd}
	ts1.ActivationPeriod = aps[0] // first activation period starts before the timespan

	timespans = append(timespans, ts1)

	for _, ap := range aps {
		for i := 0; i < len(timespans); i++ {
			ts := timespans[i]
			newTs := ts.SplitByActivationPeriod(ap)
			if newTs != nil {
				timespans = append(timespans, newTs)
			}
		}
	}

	for i := 0; i < len(timespans); i++ {
		ts := timespans[i]
		for _, interval := range ts.ActivationPeriod.Intervals {
			newTs := ts.SplitByInterval(interval)
			if newTs != nil {
				newTs.ActivationPeriod = ts.ActivationPeriod
				timespans = append(timespans, newTs)
			}
		}
	}
	return
}

func (cd *CallDescriptor) RestoreFromStorage(sg StorageGetter) (destPrefix string, err error) {
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
	timespans := cd.splitInTimeSpans(cd.getActivePeriods())

	cost := 0.0
	for _, ts := range timespans {
		cost += ts.GetCost()
	}

	cc := &CallCost{TOR: cd.TOR,
		CstmId:            cd.CstmId,
		Subject:           cd.Subject,
		DestinationPrefix: destPrefix,
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
