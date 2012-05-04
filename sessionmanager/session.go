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
	uuid, cstmId, subject, destination string
	startTime                          time.Time // destination: startTime
}

func NewSession(ev *Event) *Session {
	startTime, err := time.Parse(time.RFC1123, ev.Fields[START_TIME])
	if err != nil {
		log.Print("Error parsing answer event start time, using time.Now!")
		startTime = time.Now()
	}
	return &Session{uuid: ev.Fields[UUID],
		cstmId:      ev.Fields[CSTMID],
		subject:     ev.Fields[SUBJECT],
		destination: ev.Fields[DESTINATION],
		startTime:   startTime}
}

func (s *Session) GetSessionDurationFrom(now time.Time) (d time.Duration) {
	seconds := now.Sub(s.startTime).Seconds()
	d, err := time.ParseDuration(fmt.Sprintf("%ds", int(seconds)))
	if err != nil {
		log.Printf("Cannot parse session duration %v", seconds)
	}
	return
}

func (s *Session) GetSessionDuration() time.Duration {
	return s.GetSessionDurationFrom(time.Now())
}

func (s *Session) GetSessionCostFrom(now time.Time) (callCosts *timespans.CallCost, err error) {
	cd := &timespans.CallDescriptor{TOR: 1, CstmId: s.cstmId, Subject: s.subject, DestinationPrefix: s.destination, TimeStart: s.startTime, TimeEnd: now}
	cd.SetStorageGetter(storageGetter)
	callCosts, err = cd.GetCost()
	if err != nil {
		log.Printf("Error getting call cost for session %v", s)
	}
	return
}

func (s *Session) GetSessionCost() (callCosts *timespans.CallCost, err error) {
	return s.GetSessionCostFrom(time.Now())
}
