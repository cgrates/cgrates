/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2012-2015 ITsysCOM

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
	"errors"
	"fmt"
	"time"

	"github.com/cgrates/cgrates/engine"
)

// Session type holding the call information fields, a session delegate for specific
// actions and a channel to signal end of the debit loop.
type Session struct {
	eventStart     engine.Event // Store the original event who started this session so we can use it's info later (eg: disconnect, cgrid)
	stopDebit      chan bool    // Channel to communicate with debit loops when closing the session
	sessionManager SessionManager
	connId         string // Reference towards connection id on the session manager side.
	warnMinDur     time.Duration
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
func NewSession(ev engine.Event, connId string, sm SessionManager) *Session {
	s := &Session{eventStart: ev,
		stopDebit:      make(chan bool),
		sessionManager: sm,
		connId:         connId,
	}
	if err := sm.Rater().GetSessionRuns(*ev.AsStoredCdr(), &s.sessionRuns); err != nil || len(s.sessionRuns) == 0 {
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
	debitPeriod := s.sessionManager.DebitInterval()
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
		if err := s.sessionManager.Rater().MaxDebit(nextCd, cc); err != nil {
			engine.Logger.Err(fmt.Sprintf("Could not complete debit opperation: %v", err))
			s.sessionManager.DisconnectSession(s.eventStart, s.connId, SYSTEM_ERROR)
			return
		}
		if cc.GetDuration() == 0 {
			s.sessionManager.DisconnectSession(s.eventStart, s.connId, INSUFFICIENT_FUNDS)
			return
		}
		if s.warnMinDur != time.Duration(0) && cc.GetDuration() <= s.warnMinDur {
			s.sessionManager.WarnSessionMinDuration(s.eventStart.GetUUID(), s.connId)
		}
		s.sessionRuns[runIdx].CallCosts = append(s.sessionRuns[runIdx].CallCosts, cc)
		nextCd.TimeEnd = cc.GetEndTime() // set debited timeEnd
		// update call duration with real debited duration
		nextCd.DurationIndex -= debitPeriod
		nextCd.DurationIndex += nextCd.GetDuration()
		nextCd.MaxCostSoFar += cc.Cost
		time.Sleep(cc.GetDuration())
		index++
	}
}

// Stops the debit loop
func (s *Session) Close(ev engine.Event) error {
	close(s.stopDebit) // Close the channel so all the sessionRuns listening will be notified
	if _, err := ev.GetEndTime(); err != nil {
		engine.Logger.Err("Error parsing event stop time.")
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
		err = s.Refund(lastCC, hangupTime)
		if err != nil {
			return err
		}
	}
	go s.SaveOperations()
	return nil
}

func (s *Session) Refund(lastCC *engine.CallCost, hangupTime time.Time) error {
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
		var reply string
		err := s.sessionManager.CdrSrv().LogCallCost(&engine.CallCostLog{
			CgrId:    s.eventStart.GetCgrId(),
			Source:   engine.SESSION_MANAGER_SOURCE,
			RunId:    sr.DerivedCharger.RunId,
			CallCost: firstCC,
		}, &reply)
		// this is a protection against the case when the close event is missed for some reason
		// when the cdr arrives to cdrserver because our callcost is not there it will be rated
		// as postpaid. When the close event finally arives we have to refund everything
		if err != nil {
			if err == errors.New("unique violation ") { //FIXME: find the right error
				s.Refund(firstCC, firstCC.Timespans[0].TimeStart)
			} else {
				engine.Logger.Err(fmt.Sprintf("failed to log call cost: %v", err))
			}
		}
	}
}
