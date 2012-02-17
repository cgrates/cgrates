package timespans

import (
	"time"
	//"log"
)

/*
The struture that is saved to storage.
*/
type ActivationPeriod struct {
	ActivationTime time.Time
	Intervals      []*Interval
}

/*
Adds one ore more intervals to the internal interval list.
*/
func (ap *ActivationPeriod) AddInterval(is ...*Interval) {
	for _, i := range is {
		ap.Intervals = append(ap.Intervals, i)
	}
}
