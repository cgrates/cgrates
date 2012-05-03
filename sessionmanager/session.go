package sessionmanager

import (
	"fmt"
	"github.com/rif/cgrates/timespans"
	"log"
	"time"
)

type Session struct {
	customer, subject string
	destinations      []string
	startTimes        []time.Time
}

func (s *Session) AddCallToSession(destination string, startTime time.Time) {
	s.destinations = append(s.destinations, destination)
	s.startTimes = append(s.startTimes, startTime)
}

func (s *Session) GetSessionDurationFrom(now time.Time) (d time.Duration) {
	seconds := 0.0

	for _, st := range s.startTimes {
		seconds += now.Sub(st).Seconds()
	}
	d, err := time.ParseDuration(fmt.Sprintf("%ds", int(seconds)))
	if err != nil {
		log.Printf("Cannot parse session duration %v", seconds)
	}
	return
}

func (s *Session) GetSessionDuration() time.Duration {
	return s.GetSessionDurationFrom(time.Now())
}

func (s *Session) GetSessionCostFrom(now time.Time) (callCosts []*timespans.CallCost, err error) {
	for i, st := range s.startTimes {
		cd := &timespans.CallDescriptor{TOR: 1, CstmId: s.customer, Subject: s.subject, DestinationPrefix: s.destinations[i], TimeStart: st, TimeEnd: now}
		cd.SetStorageGetter(storageGetter)
		if cc, err := cd.GetCost(); err != nil {
			callCosts = append(callCosts, cc)
		} else {
			break
		}
	}
	return
}

func (s *Session) GetSessionCost() (callCosts []*timespans.CallCost, err error) {
	return s.GetSessionCostFrom(time.Now())
}
