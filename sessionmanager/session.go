/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2013 ITsysCOM

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
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"time"
)

// Session type holding the call information fields, a session delegate for specific
// actions and a channel to signal end of the debit loop.
type Session struct {
	uuid           string
	callDescriptor *engine.CallDescriptor
	sessionManager SessionManager
	stopDebit      chan bool
	CallCosts      []*engine.CallCost
}

// Creates a new session and starts the debit loop
func NewSession(ev Event, sm SessionManager) (s *Session) {
	// SesionManager only handles prepaid and postpaid calls
	if ev.GetReqType() != utils.PREPAID && ev.GetReqType() != utils.POSTPAID {
		return
	}
	startTime, err := ev.GetStartTime(START_TIME)
	if err != nil {
		engine.Logger.Err("Error parsing answer event start time, using time.Now!")
		startTime = time.Now()
	}

	cd := &engine.CallDescriptor{
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
		case utils.PREPAID:
			go s.startDebitLoop()
		case utils.POSTPAID:
			// do not loop, make only one debit at hangup
		}
	}
	return
}

// the debit loop method (to be stoped by sending somenthing on stopDebit channel)
func (s *Session) startDebitLoop() {
	nextCd := *s.callDescriptor
	index := 0.0
	for {
		select {
		case <-s.stopDebit:
			return
		default:
		}
		if index > 0 { // first time use the session start time
			nextCd.TimeStart = nextCd.TimeEnd
		}
		nextCd.TimeEnd = nextCd.TimeStart.Add(s.sessionManager.GetDebitPeriod())
		cc := s.sessionManager.LoopAction(s, &nextCd, index)
		time.Sleep(cc.GetDuration())
		index++
	}
}

// Returns the session duration till the specified time
func (s *Session) getSessionDurationFrom(now time.Time) (d time.Duration) {
	seconds := now.Sub(s.callDescriptor.TimeStart).Seconds()
	d, err := time.ParseDuration(fmt.Sprintf("%ds", int(seconds)))
	if err != nil {
		engine.Logger.Err(fmt.Sprintf("Cannot parse session duration %v", seconds))
	}
	return
}

// Stops the debit loop
func (s *Session) Close(ev Event) {
	engine.Logger.Debug(fmt.Sprintf("Stopping debit for %s", s.uuid))
	if s == nil {
		return
	}
	s.stopDebit <- true
	//s.callDescriptor.TimeEnd = time.Now()
	if endTime, err := ev.GetEndTime(); err != nil {
		engine.Logger.Err("Error parsing answer event stop time.")
		endTime = s.callDescriptor.TimeStart.Add(s.callDescriptor.CallDuration)
		s.callDescriptor.TimeEnd = endTime
	}
	s.SaveOperations()
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
		if s.sessionManager.GetDbLogger() == nil {
			engine.Logger.Err("<SessionManager> Error: no connection to logger database, cannot save costs")
		}
		s.sessionManager.GetDbLogger().LogCallCost(s.uuid, engine.SESSION_MANAGER_SOURCE, firstCC)
		engine.Logger.Debug(fmt.Sprintf("<SessionManager> End of call, having costs: %v", firstCC.String()))
	}()
}
