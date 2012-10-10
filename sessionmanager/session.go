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
	"github.com/cgrates/cgrates/rater"
	"strings"
	"time"
)

// Session type holding the call information fields, a session delegate for specific
// actions and a channel to signal end of the debit loop.
type Session struct {
	uuid           string
	callDescriptor *rater.CallDescriptor
	sessionManager SessionManager
	stopDebit      chan bool
	CallCosts      []*rater.CallCost
}

// Creates a new session and starts the debit loop
func NewSession(ev Event, sm SessionManager) (s *Session) {
	// if there is no account configured leave the call alone
	if strings.TrimSpace(ev.GetReqType()) == "" {
		return
	}
	startTime, err := ev.GetStartTime(START_TIME)
	if err != nil {
		rater.Logger.Err("Error parsing answer event start time, using time.Now!")
		startTime = time.Now()
	}

	cd := &rater.CallDescriptor{
		Direction:   ev.GetDirection(),
		Tenant:      ev.GetTenant(),
		TOR:         ev.GetTOR(),
		Subject:     ev.GetSubject(),
		Account:     ev.GetAccount(),
		Destination: ev.GetDestination(),
		TimeStart:   startTime}
	s = &Session{uuid: ev.GetUUID(),
		callDescriptor: cd,
		stopDebit:      make(chan bool, 2)} //buffer it for multiple close signals
	s.sessionManager = sm
	if ev.MissingParameter() {
		sm.DisconnectSession(s, MISSING_PARAMETER)
	} else {
		switch ev.GetReqType() {
		case REQTYPE_PREPAID:
			go s.startDebitLoop()
		case REQTYPE_POSTPAID:
			// do not loop, make only one debit at hangup
		}
	}
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
		if nextCd.TimeEnd != s.callDescriptor.TimeEnd { // first time use the session start time
			nextCd.TimeStart = time.Now()
		}
		nextCd.TimeEnd = time.Now().Add(s.sessionManager.GetDebitPeriod())
		s.sessionManager.LoopAction(s, &nextCd)
		time.Sleep(s.sessionManager.GetDebitPeriod())
	}
}

// Returns the session duration till the specified time
func (s *Session) getSessionDurationFrom(now time.Time) (d time.Duration) {
	seconds := now.Sub(s.callDescriptor.TimeStart).Seconds()
	d, err := time.ParseDuration(fmt.Sprintf("%ds", int(seconds)))
	if err != nil {
		rater.Logger.Err(fmt.Sprintf("Cannot parse session duration %v", seconds))
	}
	return
}

// Returns the session duration till now
func (s *Session) GetSessionDuration() time.Duration {
	return s.getSessionDurationFrom(time.Now())
}

// Stops the debit loop
func (s *Session) Close() {
	rater.Logger.Debug(fmt.Sprintf("Stopping debit for %s", s.uuid))
	if s == nil {
		return
	}
	s.stopDebit <- true
	s.callDescriptor.TimeEnd = time.Now()
	s.sessionManager.RemoveSession(s)
}

// Nice print for session
func (s *Session) String() string {
	return fmt.Sprintf("%v: %s(%s) -> %s", s.callDescriptor.TimeStart, s.callDescriptor.Subject, s.callDescriptor.Account, s.callDescriptor.Destination)
}

// 
func (s *Session) SaveOperations() {
	go func() {
		if s == nil || len(s.CallCosts) == 0 {
			return
		}
		firstCC := s.CallCosts[0]
		for _, cc := range s.CallCosts[1:] {
			firstCC.Merge(cc)
		}
		if s.sessionManager.GetDbLogger() != nil {
			s.sessionManager.GetDbLogger().LogCallCost(s.uuid, rater.SESSION_MANAGER_SOURCE, firstCC)
		}
		rater.Logger.Debug(firstCC.String())
	}()
}
