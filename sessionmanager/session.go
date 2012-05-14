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

const (
	DEBIT_PERIOD = 10 * time.Second
)

// Session type holding the call information fields, a session delegate for specific
// actions and a channel to signal end of the debit loop.
type Session struct {
	uuid, cstmId, subject, destination string
	startTime                          time.Time // destination: startTime
	sessionDelegate                    SessionDelegate
	stopDebit                          chan byte
}

// Creates a new session and starts the debit loop
func NewSession(ev *Event, ed SessionDelegate) (s *Session) {
	startTime, err := time.Parse(time.RFC1123, ev.Fields[START_TIME])
	if err != nil {
		log.Print("Error parsing answer event start time, using time.Now!")
		startTime = time.Now()
	}
	s = &Session{uuid: ev.Fields[UUID],
		cstmId:      ev.Fields[CSTMID],
		subject:     ev.Fields[SUBJECT],
		destination: ev.Fields[DESTINATION],
		startTime:   startTime,
		stopDebit:   make(chan byte)}
	s.sessionDelegate = ed
	go s.startDebitLoop()
	return
}

// the debit loop method (to be stoped by sending somenting on stopDebit channel)
func (s *Session) startDebitLoop() {
	for {
		select {
		case <-s.stopDebit:
			return
		default:
		}
		s.sessionDelegate.LoopAction()
		time.Sleep(DEBIT_PERIOD)
	}
}

// Returns the session duration till the specified time
func (s *Session) getSessionDurationFrom(now time.Time) (d time.Duration) {
	seconds := now.Sub(s.startTime).Seconds()
	d, err := time.ParseDuration(fmt.Sprintf("%ds", int(seconds)))
	if err != nil {
		log.Printf("Cannot parse session duration %v", seconds)
	}
	return
}

// Returns the session duration till now
func (s *Session) GetSessionDuration() time.Duration {
	return s.getSessionDurationFrom(time.Now())
}

// Returns the session cost till the specified time
func (s *Session) getSessionCostFrom(now time.Time) (callCosts *timespans.CallCost, err error) {
	cd := &timespans.CallDescriptor{TOR: 1, CstmId: s.cstmId, Subject: s.subject, DestinationPrefix: s.destination, TimeStart: s.startTime, TimeEnd: now}
	cd.SetStorageGetter(s.sessionDelegate.GetStorageGetter())
	callCosts, err = cd.GetCost()
	if err != nil {
		log.Printf("Error getting call cost for session %v", s)
	}
	return
}

// Returns the session duration till now
func (s *Session) GetSessionCost() (callCosts *timespans.CallCost, err error) {
	return s.getSessionCostFrom(time.Now())
}

// Stops the debit loop
func (s *Session) Close() {
	s.stopDebit <- 1
}
