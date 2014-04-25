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
	"encoding/json"
	"fmt"
	"time"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// Session type holding the call information fields, a session delegate for specific
// actions and a channel to signal end of the debit loop.
type Session struct {
	cgrid          string
	uuid           string
	stopDebit      chan bool
	sessionManager SessionManager
	sessionRuns    []*SessionRun
}

// One individual run
type SessionRun struct {
	runId          string
	callDescriptor *engine.CallDescriptor
	callCosts      []*engine.CallCost
}

// Creates a new session and in case of prepaid starts the debit loop for each of the session runs individually
func NewSession(ev Event, sm SessionManager) *Session {
	s := &Session{cgrid: ev.GetCgrId(),
		uuid:           ev.GetUUID(),
		stopDebit:      make(chan bool),
		sessionManager: sm,
		sessionRuns:    make([]*SessionRun, 0),
	}
	runIds := append([]string{utils.DEFAULT_RUNID}, cfg.SMRunIds...) // Prepend default runid to extra configured for session manager
	for idx, runId := range runIds {                                 // Create the SessionRuns here
		var reqTypeFld, directionFld, tenantFld, torFld, actFld, subjFld, dstFld, aTimeFld string
		if idx != 0 { // Take fields out of config, default ones are automatically handled as empty
			idxCfg := idx - 1 // In configuration we did not prepend values
			reqTypeFld = cfg.SMReqTypeFields[idxCfg]
			directionFld = cfg.SMDirectionFields[idxCfg]
			tenantFld = cfg.SMTenantFields[idxCfg]
			torFld = cfg.SMTORFields[idxCfg]
			actFld = cfg.SMAccountFields[idxCfg]
			subjFld = cfg.SMSubjectFields[idxCfg]
			dstFld = cfg.SMDestFields[idxCfg]
			aTimeFld = cfg.SMAnswerTimeFields[idxCfg]
		}
		startTime, err := ev.GetAnswerTime(aTimeFld)
		if err != nil {
			engine.Logger.Err("Error parsing answer event start time, using time.Now!")
			startTime = time.Now()
		}
		cd := &engine.CallDescriptor{
			Direction:   ev.GetDirection(directionFld),
			Tenant:      ev.GetTenant(tenantFld),
			Category:    ev.GetTOR(torFld),
			Subject:     ev.GetSubject(subjFld),
			Account:     ev.GetAccount(actFld),
			Destination: ev.GetDestination(dstFld),
			TimeStart:   startTime}
		sr := &SessionRun{
			runId:          runId,
			callDescriptor: cd,
		}
		s.sessionRuns = append(s.sessionRuns, sr)
		if ev.GetReqType(reqTypeFld) == utils.PREPAID {
			go s.debitLoop(len(s.sessionRuns) - 1) // Send index of the just appended sessionRun
		}
	}
	if len(s.sessionRuns) == 0 {
		return nil
	}
	return s
}

// the debit loop method (to be stoped by sending somenthing on stopDebit channel)
func (s *Session) debitLoop(runIdx int) {
	nextCd := *s.sessionRuns[runIdx].callDescriptor
	index := 0.0
	debitPeriod := s.sessionManager.GetDebitPeriod()
	for {
		select {
		case <-s.stopDebit:
			return
		default:
		}
		if index > 0 { // first time use the session start time
			nextCd.TimeStart = nextCd.TimeEnd
		}
		nextCd.TimeEnd = nextCd.TimeStart.Add(debitPeriod)
		nextCd.LoopIndex = index
		nextCd.DurationIndex += debitPeriod // first presumed duration
		cc := &engine.CallCost{}
		if err := s.sessionManager.MaxDebit(&nextCd, cc); err != nil {
			engine.Logger.Err(fmt.Sprintf("Could not complete debit opperation: %v", err))
			// disconnect session
			s.sessionManager.DisconnectSession(s.uuid, SYSTEM_ERROR)
			return
		}
		if cc.GetDuration() == 0 {
			s.sessionManager.DisconnectSession(s.uuid, INSUFFICIENT_FUNDS)
			return
		}
		s.sessionRuns[runIdx].callCosts = append(s.sessionRuns[runIdx].callCosts, cc)
		nextCd.TimeEnd = cc.GetEndTime() // set debited timeEnd
		// update call duration with real debited duration
		nextCd.DurationIndex -= debitPeriod
		nextCd.DurationIndex += nextCd.GetDuration()
		time.Sleep(cc.GetDuration())
		index++
	}
}

// Stops the debit loop
func (s *Session) Close(ev Event) {
	// engine.Logger.Debug(fmt.Sprintf("Stopping debit for %s", s.uuid))
	if s == nil {
		return
	}
	close(s.stopDebit) // Close the channel so all the sessionRuns listening will be notified
	if _, err := ev.GetEndTime(); err != nil {
		engine.Logger.Err("Error parsing answer event stop time.")
		for idx := range s.sessionRuns {
			s.sessionRuns[idx].callDescriptor.TimeEnd = s.sessionRuns[idx].callDescriptor.TimeStart.Add(s.sessionRuns[idx].callDescriptor.DurationIndex)
		}
	}
	s.SaveOperations()
}

// Nice print for session
func (s *Session) String() string {
	sDump, _ := json.Marshal(s)
	return string(sDump)
}

// Saves call_costs for each session run
func (s *Session) SaveOperations() {
	if s == nil {
		return
	}
	go func() {
		for _, sr := range s.sessionRuns {
			if len(sr.callCosts) == 0 {
				break // There are no costs to save, ignore the operation
			}
			firstCC := sr.callCosts[0]
			for _, cc := range sr.callCosts[1:] {
				firstCC.Merge(cc)
			}
			if s.sessionManager.GetDbLogger() == nil {
				engine.Logger.Err("<SessionManager> Error: no connection to logger database, cannot save costs")
			}
			s.sessionManager.GetDbLogger().LogCallCost(s.cgrid, engine.SESSION_MANAGER_SOURCE, sr.runId, firstCC)
		}
	}()
}
