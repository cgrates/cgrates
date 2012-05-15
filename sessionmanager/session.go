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

// Session type holding the call information fields, a session delegate for specific
// actions and a channel to signal end of the debit loop.
type Session struct {
	uuid            string
	callDescriptor  *timespans.CallDescriptor
	sessionDelegate SessionDelegate
	stopDebit       chan byte
	CallCosts       []*timespans.CallCost
}

// Creates a new session and starts the debit loop
func NewSession(ev *Event, ed SessionDelegate) (s *Session) {
	startTime, err := time.Parse(time.RFC1123, ev.Fields[START_TIME])
	if err != nil {
		log.Print("Error parsing answer event start time, using time.Now!")
		startTime = time.Now()
	}
	cd := &timespans.CallDescriptor{TOR: ev.Fields[TOR],
		CstmId:            ev.Fields[CSTMID],
		Subject:           ev.Fields[SUBJECT],
		DestinationPrefix: ev.Fields[DESTINATION],
		TimeStart:         startTime}
	s = &Session{uuid: ev.Fields[UUID],
		callDescriptor: cd,
		stopDebit:      make(chan byte)}
	s.sessionDelegate = ed
	go s.startDebitLoop()
	return
}

// the debit loop method (to be stoped by sending somenting on stopDebit channel)
func (s *Session) startDebitLoop() {
	nextCd := *s.callDescriptor
	for {
		select {
		case <-s.stopDebit:
			return
		default:
		}
		if nextCd.TimeEnd == s.callDescriptor.TimeEnd { // first time use the session start time
			nextCd.TimeStart = time.Now()
		}
		nextCd.TimeEnd = time.Now().Add(s.sessionDelegate.GetDebitPeriod())
		s.sessionDelegate.LoopAction(s, &nextCd)
		time.Sleep(s.sessionDelegate.GetDebitPeriod())
	}
}

// Returns the session duration till the specified time
func (s *Session) getSessionDurationFrom(now time.Time) (d time.Duration) {
	seconds := now.Sub(s.callDescriptor.TimeStart).Seconds()
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

// Stops the debit loop
func (s *Session) Close() {
	s.stopDebit <- 1
	s.callDescriptor.TimeEnd = time.Now()
}
