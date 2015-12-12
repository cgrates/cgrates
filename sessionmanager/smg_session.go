/*
Real-time Charging System for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

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
	"errors"
	"fmt"
	"time"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// One session handled by SM
type SMGSession struct {
	eventStart    SMGenericEvent // Event which started
	stopDebit     chan struct{}  // Channel to communicate with debit loops when closing the session
	connId        string         // Reference towards connection id on the session manager side.
	runId         string         // Keep a reference for the derived run
	timezone      string
	rater         engine.Connector // Connector to Rater service
	cdrsrv        engine.Connector // Connector to CDRS service
	extconns      *SMGExternalConnections
	cd            *engine.CallDescriptor
	sessionCds    []*engine.CallDescriptor
	callCosts     []*engine.CallCost
	extraDuration time.Duration // keeps the current duration debited on top of what heas been asked
}

// Called in case of automatic debits
func (self *SMGSession) debitLoop(debitInterval time.Duration) {
	loopIndex := 0
	for {
		select {
		case <-self.stopDebit:
			return
		default:
		}
		if maxDebit, err := self.debit(debitInterval); err != nil {
			utils.Logger.Err(fmt.Sprintf("<SMGeneric> Could not complete debit opperation on session: %s, error: %s", self.eventStart.GetUUID(), err.Error()))
			disconnectReason := SYSTEM_ERROR
			if err.Error() == utils.ErrUnauthorizedDestination.Error() {
				disconnectReason = UNAUTHORIZED_DESTINATION
			}
			if err := self.disconnectSession(disconnectReason); err != nil {
				utils.Logger.Err(fmt.Sprintf("<SMGeneric> Could not disconnect session: %s, error: %s", self.eventStart.GetUUID(), err.Error()))
			}
			return
		} else if maxDebit < debitInterval {
			time.Sleep(maxDebit)
			if err := self.disconnectSession(INSUFFICIENT_FUNDS); err != nil {
				utils.Logger.Err(fmt.Sprintf("<SMGeneric> Could not disconnect session: %s, error: %s", self.eventStart.GetUUID(), err.Error()))
			}
			return
		}
		time.Sleep(debitInterval)
		loopIndex++
	}
}

// Attempts to debit a duration, returns maximum duration which can be debitted or error
func (self *SMGSession) debit(dur time.Duration) (time.Duration, error) {
	// apply correction from previous run
	dur -= self.extraDuration
	self.extraDuration = 0

	if self.cd.LoopIndex > 0 {
		self.cd.TimeStart = self.cd.TimeEnd
	}
	self.cd.TimeEnd = self.cd.TimeStart.Add(dur)
	self.cd.DurationIndex += dur
	cc := &engine.CallCost{}
	if err := self.rater.MaxDebit(self.cd, cc); err != nil {
		return 0, err
	}
	// cd corrections
	self.cd.TimeEnd = cc.GetEndTime() // set debited timeEnd
	// update call duration with real debited duration
	ccDuration := cc.GetDuration()
	if ccDuration != dur {
		self.extraDuration = ccDuration - dur
	}
	self.cd.DurationIndex -= dur
	self.cd.DurationIndex += ccDuration
	self.cd.MaxCostSoFar += cc.Cost
	self.cd.LoopIndex += 1
	self.sessionCds = append(self.sessionCds, self.cd.Clone())
	self.callCosts = append(self.callCosts, cc)
	return ccDuration, nil
}

// Attempts to refund a duration, error on failure
func (self *SMGSession) refund(refundDuration time.Duration) error {
	lastCC := self.callCosts[len(self.callCosts)-1]
	lastCC.Timespans.Decompress()
	var refundIncrements engine.Increments
	for i := len(lastCC.Timespans) - 1; i >= 0; i-- {
		ts := lastCC.Timespans[i]
		tsDuration := ts.GetDuration()
		if refundDuration <= tsDuration {

			lastRefundedIncrementIndex := -1
			for j := len(ts.Increments) - 1; j >= 0; j-- {
				increment := ts.Increments[j]
				if increment.Duration <= refundDuration {
					refundIncrements = append(refundIncrements, increment)
					refundDuration -= increment.Duration
					lastRefundedIncrementIndex = j
				} else {
					break //increment duration is larger, cannot refund increment
				}
			}
			if lastRefundedIncrementIndex == 0 {
				lastCC.Timespans[i] = nil
				lastCC.Timespans = lastCC.Timespans[:i]
			} else {
				ts.SplitByIncrement(lastRefundedIncrementIndex)
			}
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
	// utils.Logger.Info(fmt.Sprintf("Refund duration: %v", initialRefundDuration-refundDuration))
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
		err := self.rater.RefundIncrements(cd, &response)
		if err != nil {
			return err
		}
	}
	//utils.Logger.Debug(fmt.Sprintf("REFUND INCR: %s", utils.ToJSON(refundIncrements)))
	lastCC.Cost -= refundIncrements.GetTotalCost()
	lastCC.Timespans.Compress()
	return nil
}

// Session has ended, check debits and refund the extra charged duration
func (self *SMGSession) close(endTime time.Time) error {
	lastCC := self.callCosts[len(self.callCosts)-1]
	end := lastCC.GetEndTime()
	refundDuration := end.Sub(endTime)
	self.refund(refundDuration)
	return nil
}

// Send disconnect order to remote connection
func (self *SMGSession) disconnectSession(reason string) error {
	type AttrDisconnectSession struct {
		EventStart map[string]interface{}
		Reason     string
	}
	conn := self.extconns.GetConnection(self.connId)
	if conn == nil {
		return ErrConnectionNotFound
	}
	var reply string
	if err := conn.Call("SMGClientV1.DisconnectSession", AttrDisconnectSession{EventStart: self.eventStart, Reason: reason}, &reply); err != nil {
		return err
	} else if reply != utils.OK {
		return errors.New(fmt.Sprintf("Unexpected disconnect reply: %s", reply))
	}
	return nil
}

// Merge the sum of costs and sends it to CDRS for storage
func (self *SMGSession) saveOperations() error {
	if len(self.callCosts) == 0 {
		return nil // There are no costs to save, ignore the operation
	}
	firstCC := self.callCosts[0]
	for _, cc := range self.callCosts[1:] {
		firstCC.Merge(cc)
	}
	var reply string
	err := self.cdrsrv.LogCallCost(&engine.CallCostLog{
		CgrId:          self.eventStart.GetCgrId(self.timezone),
		Source:         utils.SESSION_MANAGER_SOURCE,
		RunId:          self.runId,
		CallCost:       firstCC,
		CheckDuplicate: true,
	}, &reply)
	// this is a protection against the case when the close event is missed for some reason
	// when the cdr arrives to cdrserver because our callcost is not there it will be rated
	// as postpaid. When the close event finally arives we have to refund everything
	if err != nil {
		if err == utils.ErrExists {
			self.refund(self.cd.GetDuration()) // Refund entire duration
		} else {
			return err
		}
	}
	return nil
}
