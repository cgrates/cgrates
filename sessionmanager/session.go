/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2012  Radu Ioan Fericean

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>
*/

package sessionmanager

import (
	"fmt"
	"github.com/rif/cgrates/timespans"
	"log"
	"time"
)

type Session struct {
	id, cstmId, subject string
	callsMap            map[string]time.Time
}

func NewSession(cstmId, subject string) *Session {
	return &Session{cstmId: cstmId, subject: subject, callsMap: make(map[string]time.Time)}
}

func (s *Session) AddCallToSession(destination string, startTime time.Time) {
	s.callsMap[destination] = startTime
}

func (s *Session) GetSessionDurationFrom(now time.Time) (d time.Duration) {
	seconds := 0.0

	for _, st := range s.callsMap {
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
	for dest, st := range s.callsMap {
		cd := &timespans.CallDescriptor{TOR: 1, CstmId: s.cstmId, Subject: s.subject, DestinationPrefix: dest, TimeStart: st, TimeEnd: now}
		cd.SetStorageGetter(storageGetter)
		if cc, err := cd.GetCost(); err == nil {
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
