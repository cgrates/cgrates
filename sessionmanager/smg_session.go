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
	lastUsage     time.Duration // Keep record of the last debit for LastUsed functionality
	totalUsage    time.Duration
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
		if maxDebit, err := self.debit(debitInterval, nilDuration); err != nil {
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
func (self *SMGSession) debit(dur time.Duration, lastUsed time.Duration) (time.Duration, error) {
	lastUsedCorrection := time.Duration(0) // Used if lastUsed influences the debit
	if self.cd.DurationIndex != 0 && lastUsed != 0 {
		if self.lastUsage > lastUsed { // We have debitted more than we have used, refund in the duration debitted
			lastUsedCorrection = -(self.lastUsage - lastUsed)
		} else { // We have debitted less than we have consumed, add the difference to duration debitted
			lastUsedCorrection = lastUsed - self.lastUsage
		}

		// apply the lastUsed correction
		dur += lastUsedCorrection
		self.totalUsage += lastUsed // Should reflect the total usage so far
	} else {
		// apply correction from previous run
		dur -= self.extraDuration
	}
	self.extraDuration = 0
	if self.cd.LoopIndex > 0 {
		self.cd.TimeStart = self.cd.TimeEnd
	}
	self.cd.TimeEnd = self.cd.TimeStart.Add(dur)
	self.cd.DurationIndex += dur
	cc := &engine.CallCost{}
	if err := self.rater.MaxDebit(self.cd, cc); err != nil {
		self.lastUsage = 0
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
	ccDuration -= lastUsedCorrection
	if ccDuration < 0 { // if correction has pushed ccDuration bellow 0
		ccDuration = 0
	}
	self.lastUsage = ccDuration // Reset the lastUsage for later reference
	return ccDuration, nil
}

// Attempts to refund a duration, error on failure
func (self *SMGSession) refund(refundDuration time.Duration) error {
	initialRefundDuration := refundDuration
	firstCC := self.callCosts[0] // use merged cc (from close function)
	firstCC.Timespans.Decompress()
	var refundIncrements engine.Increments
	for i := len(firstCC.Timespans) - 1; i >= 0; i-- {
		ts := firstCC.Timespans[i]
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
				firstCC.Timespans[i] = nil
				firstCC.Timespans = firstCC.Timespans[:i]
			} else {
				ts.SplitByIncrement(lastRefundedIncrementIndex)
				ts.Cost = ts.CalculateCost()
			}
			break // do not go to other timespans
		} else {
			refundIncrements = append(refundIncrements, ts.Increments...)
			// remove the timespan entirely
			firstCC.Timespans[i] = nil
			firstCC.Timespans = firstCC.Timespans[:i]
			// continue to the next timespan with what is left to refund
			refundDuration -= tsDuration
		}
	}
	// show only what was actualy refunded (stopped in timespan)
	// utils.Logger.Info(fmt.Sprintf("Refund duration: %v", initialRefundDuration-refundDuration))
	if len(refundIncrements) > 0 {
		cd := firstCC.CreateCallDescriptor()
		cd.Increments = refundIncrements
		cd.Increments.Compress()
		utils.Logger.Info(fmt.Sprintf("Refunding duration %v with cd: %s", initialRefundDuration, utils.ToJSON(cd)))
		var response float64
		err := self.rater.RefundIncrements(cd, &response)
		if err != nil {
			return err
		}
	}
	//firstCC.Cost -= refundIncrements.GetTotalCost() // use updateCost instead
	firstCC.UpdateCost()
	firstCC.UpdateRatedUsage()
	firstCC.Timespans.Compress()
	return nil
}

// Session has ended, check debits and refund the extra charged duration
func (self *SMGSession) close(endTime time.Time) error {
	if len(self.callCosts) != 0 { // We have had at least one cost calculation
		firstCC := self.callCosts[0]
		for _, cc := range self.callCosts[1:] {
			firstCC.Merge(cc)
		}
		//utils.Logger.Debug("MergedCC: " + utils.ToJSON(firstCC))
		end := firstCC.GetEndTime()
		refundDuration := end.Sub(endTime)
		self.refund(refundDuration)
	}
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
	firstCC := self.callCosts[0] // was merged in close method
	firstCC.Round()
	//utils.Logger.Debug("Saved CC: " + utils.ToJSON(firstCC))
	roundIncrements := firstCC.GetRoundIncrements()
	if len(roundIncrements) != 0 {
		cd := firstCC.CreateCallDescriptor()
		cd.Increments = roundIncrements
		var response float64
		if err := self.rater.RefundRounding(cd, &response); err != nil {
			return err
		}
	}
	var reply string
	smCost := &engine.SMCost{
		CGRID:       self.eventStart.GetCgrId(self.timezone),
		CostSource:  utils.SESSION_MANAGER_SOURCE,
		RunID:       self.runId,
		OriginHost:  self.eventStart.GetOriginatorIP(utils.META_DEFAULT),
		OriginID:    self.eventStart.GetUUID(),
		CostDetails: firstCC,
	}
	if err := self.cdrsrv.StoreSMCost(engine.AttrCDRSStoreSMCost{SMCost: smCost, CheckDuplicate: true}, &reply); err != nil {
		// this is a protection against the case when the close event is missed for some reason
		// when the cdr arrives to cdrserver because our callcost is not there it will be rated
		// as postpaid. When the close event finally arives we have to refund everything
		if err == utils.ErrExists {
			self.refund(self.cd.GetDuration()) // Refund entire duration
		} else {
			return err
		}
	}
	return nil
}

func (self *SMGSession) TotalUsage() time.Duration {
	return self.totalUsage
}

func (self *SMGSession) AsActiveSession(timezone string) *ActiveSession {
	sTime, _ := self.eventStart.GetSetupTime(utils.META_DEFAULT, timezone)
	aTime, _ := self.eventStart.GetAnswerTime(utils.META_DEFAULT, timezone)
	usage, _ := self.eventStart.GetUsage(utils.META_DEFAULT)
	pdd, _ := self.eventStart.GetPdd(utils.META_DEFAULT)
	aSession := &ActiveSession{
		CgrId:       self.eventStart.GetCgrId(timezone),
		TOR:         utils.VOICE,
		RunId:       self.runId,
		AccId:       self.eventStart.GetUUID(),
		CdrHost:     self.eventStart.GetOriginatorIP(utils.META_DEFAULT),
		CdrSource:   self.eventStart.GetCdrSource(),
		ReqType:     self.eventStart.GetReqType(utils.META_DEFAULT),
		Direction:   self.eventStart.GetDirection(utils.META_DEFAULT),
		Tenant:      self.eventStart.GetTenant(utils.META_DEFAULT),
		Category:    self.eventStart.GetCategory(utils.META_DEFAULT),
		Account:     self.eventStart.GetAccount(utils.META_DEFAULT),
		Subject:     self.eventStart.GetSubject(utils.META_DEFAULT),
		Destination: self.eventStart.GetDestination(utils.META_DEFAULT),
		SetupTime:   sTime,
		AnswerTime:  aTime,
		Usage:       usage,
		Pdd:         pdd,
		ExtraFields: self.eventStart.GetExtraFields(),
		Supplier:    self.eventStart.GetSupplier(utils.META_DEFAULT),
		SMId:        "CGR-DA",
	}
	if self.cd != nil {
		aSession.LoopIndex = self.cd.LoopIndex
		aSession.DurationIndex = self.cd.DurationIndex
		aSession.MaxRate = self.cd.MaxRate
		aSession.MaxRateUnit = self.cd.MaxRateUnit
		aSession.MaxCostSoFar = self.cd.MaxCostSoFar
	}
	return aSession
}
