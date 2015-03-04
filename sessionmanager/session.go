/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2012-2014 ITsysCOM

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
	"github.com/cgrates/fsock"
)

// Session type holding the call information fields, a session delegate for specific
// actions and a channel to signal end of the debit loop.
type Session struct {
	eventStart     utils.Event // Store the original event who started this session so we can use it's info later (eg: disconnect, cgrid)
	stopDebit      chan bool   // Channel to communicate with debit loops when closing the session
	sessionManager SessionManager
	sessionRuns    []*engine.SessionRun
}

func (s *Session) GetSessionRun(runid string) *engine.SessionRun {
	for _, sr := range s.sessionRuns {
		if sr.DerivedCharger.RunId == runid {
			return sr
		}
	}
	return nil
}

func (s *Session) SessionRuns() []*engine.SessionRun {
	return s.sessionRuns
}

// Creates a new session and in case of prepaid starts the debit loop for each of the session runs individually
func NewSession(ev utils.Event, sm SessionManager) *Session {
	s := &Session{eventStart: ev,
		stopDebit:      make(chan bool),
		sessionManager: sm,
	}
	//sRuns := make([]*engine.SessionRun, 0)
	if err := sm.Rater().GetSessionRuns(ev, &s.sessionRuns); err != nil || len(s.sessionRuns) == 0 {
		return nil
	}
	for runIdx := range s.sessionRuns {
		go s.debitLoop(runIdx) // Send index of the just appended sessionRun
	}
	return s
}

// the debit loop method (to be stoped by sending somenthing on stopDebit channel)
func (s *Session) debitLoop(runIdx int) {
	nextCd := *s.sessionRuns[runIdx].CallDescriptor
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
		cc := new(engine.CallCost)
		if err := s.sessionManager.MaxDebit(&nextCd, cc); err != nil {
			engine.Logger.Err(fmt.Sprintf("Could not complete debit opperation: %v", err))
			s.sessionManager.DisconnectSession(s.eventStart, SYSTEM_ERROR)
			return
		}
		if cc.GetDuration() == 0 {
			s.sessionManager.DisconnectSession(s.eventStart, INSUFFICIENT_FUNDS)
			return
		}
		if cc.GetDuration() <= cfg.FSMinDurLowBalance && len(cfg.FSLowBalanceAnnFile) != 0 {
			if _, err := fsock.FS.SendApiCmd(fmt.Sprintf("uuid_broadcast %s %s aleg\n\n", s.eventStart.GetUUID(), cfg.FSLowBalanceAnnFile)); err != nil {
				engine.Logger.Err(fmt.Sprintf("<SessionManager> Could not send uuid_broadcast to freeswitch: %s", err.Error()))
			}
		}
		s.sessionRuns[runIdx].CallCosts = append(s.sessionRuns[runIdx].CallCosts, cc)
		nextCd.TimeEnd = cc.GetEndTime() // set debited timeEnd
		// update call duration with real debited duration
		nextCd.DurationIndex -= debitPeriod
		nextCd.DurationIndex += nextCd.GetDuration()
		time.Sleep(cc.GetDuration())
		index++
	}
}

// Stops the debit loop
func (s *Session) Close(ev utils.Event) error {
	close(s.stopDebit) // Close the channel so all the sessionRuns listening will be notified
	if _, err := ev.GetEndTime(); err != nil {
		engine.Logger.Err("Error parsing answer event stop time.")
		for idx := range s.sessionRuns {
			s.sessionRuns[idx].CallDescriptor.TimeEnd = s.sessionRuns[idx].CallDescriptor.TimeStart.Add(s.sessionRuns[idx].CallDescriptor.DurationIndex)
		}
	}
	// Costs refunds
	for _, sr := range s.SessionRuns() {
		if len(sr.CallCosts) == 0 {
			continue // why would we have 0 callcosts
		}
		lastCC := sr.CallCosts[len(sr.CallCosts)-1]
		lastCC.Timespans.Decompress()
		// put credit back
		startTime, err := ev.GetAnswerTime(sr.DerivedCharger.AnswerTimeField)
		if err != nil {
			engine.Logger.Crit("Error parsing prepaid call start time from event")
			return err
		}
		duration, err := ev.GetDuration(sr.DerivedCharger.UsageField)
		if err != nil {
			engine.Logger.Crit(fmt.Sprintf("Error parsing call duration from event %s", err.Error()))
			return err
		}
		hangupTime := startTime.Add(duration)
		end := lastCC.Timespans[len(lastCC.Timespans)-1].TimeEnd
		refundDuration := end.Sub(hangupTime)
		var refundIncrements engine.Increments
		for i := len(lastCC.Timespans) - 1; i >= 0; i-- {
			ts := lastCC.Timespans[i]
			tsDuration := ts.GetDuration()
			if refundDuration <= tsDuration {
				lastRefundedIncrementIndex := 0
				for j := len(ts.Increments) - 1; j >= 0; j-- {
					increment := ts.Increments[j]
					if increment.Duration <= refundDuration {
						refundIncrements = append(refundIncrements, increment)
						refundDuration -= increment.Duration
						lastRefundedIncrementIndex = j
					}
				}
				ts.SplitByIncrement(lastRefundedIncrementIndex)
				break // do not go to other timespans
			} else {
				refundIncrements = append(refundIncrements, ts.Increments...)
				// remove the timespan entirely
				lastCC.Timespans[i] = nil
				lastCC.Timespans = lastCC.Timespans[:i]
				// continue to the next timespan with what is left to refund
				refundDuration -= tsDuration
			}
		}
		// show only what was actualy refunded (stopped in timespan)
		// engine.Logger.Info(fmt.Sprintf("Refund duration: %v", initialRefundDuration-refundDuration))
		if len(refundIncrements) > 0 {
			cd := &engine.CallDescriptor{
				Direction:   lastCC.Direction,
				Tenant:      lastCC.Tenant,
				Category:    lastCC.Category,
				Subject:     lastCC.Subject,
				Account:     lastCC.Account,
				Destination: lastCC.Destination,
				Increments:  refundIncrements,
			}
			var response float64
			err := s.sessionManager.Rater().RefundIncrements(*cd, &response)
			if err != nil {
				return err
			}
		}
		cost := refundIncrements.GetTotalCost()
		lastCC.Cost -= cost
		lastCC.Timespans.Compress()
	}
	go s.SaveOperations()
	return nil
}

// Nice print for session
func (s *Session) String() string {
	sDump, _ := json.Marshal(s)
	return string(sDump)
}

// Saves call_costs for each session run
func (s *Session) SaveOperations() {
	for _, sr := range s.sessionRuns {
		if len(sr.CallCosts) == 0 {
			break // There are no costs to save, ignore the operation
		}
		firstCC := sr.CallCosts[0]
		for _, cc := range sr.CallCosts[1:] {
			firstCC.Merge(cc)
		}
		if s.sessionManager.GetDbLogger() == nil {
			engine.Logger.Err("<SessionManager> Error: no connection to logger database, cannot save costs")
		}
		s.sessionManager.GetDbLogger().LogCallCost(s.eventStart.GetCgrId(), engine.SESSION_MANAGER_SOURCE, sr.DerivedCharger.RunId, firstCC)
	}
}
