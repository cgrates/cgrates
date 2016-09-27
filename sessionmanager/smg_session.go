/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
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
	"reflect"
	"time"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

// One session handled by SM
type SMGSession struct {
	eventStart    SMGenericEvent // Event which started
	stopDebit     chan struct{}  // Channel to communicate with debit loops when closing the session
	runId         string         // Keep a reference for the derived run
	timezone      string
	rater         rpcclient.RpcClientConnection // Connector to Rater service
	cdrsrv        rpcclient.RpcClientConnection // Connector to CDRS service
	cd            *engine.CallDescriptor
	sessionCds    []*engine.CallDescriptor
	callCosts     []*engine.CallCost
	extraDuration time.Duration                 // keeps the current duration debited on top of what heas been asked
	lastUsage     time.Duration                 // last requested Duration
	lastDebit     time.Duration                 // last real debited duration
	totalUsage    time.Duration                 // sum of lastUsage
	clntConn      rpcclient.RpcClientConnection // Reference towards client connection on SMG side so we can disconnect.
}

// Called in case of automatic debits
func (self *SMGSession) debitLoop(debitInterval time.Duration) {
	loopIndex := 0
	sleepDur := time.Duration(0) // start with empty duration for debit
	for {
		select {
		case <-self.stopDebit:
			return
		case <-time.After(sleepDur):
			if maxDebit, err := self.debit(debitInterval, nil); err != nil {
				utils.Logger.Err(fmt.Sprintf("<SMGeneric> Could not complete debit operation on session: %s, error: %s", self.eventStart.GetUUID(), err.Error()))
				disconnectReason := SYSTEM_ERROR
				if err.Error() == utils.ErrUnauthorizedDestination.Error() {
					disconnectReason = err.Error()
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
			sleepDur = debitInterval
			loopIndex++
		}
	}
}

// Attempts to debit a duration, returns maximum duration which can be debitted or error
func (self *SMGSession) debit(dur time.Duration, lastUsed *time.Duration) (time.Duration, error) {
	//utils.Logger.Debug(fmt.Sprintf("### SMGSession.debit, dur: %+v, lastUsed: %+v, session: %+v", dur, lastUsed, self))
	requestedDuration := dur
	if lastUsed != nil {
		self.extraDuration = self.lastDebit - *lastUsed
		if *lastUsed != self.lastUsage {
			// total usage correction
			self.totalUsage -= self.lastUsage
			self.totalUsage += *lastUsed
		}
	}
	// apply correction from previous run
	if self.extraDuration < dur {
		dur -= self.extraDuration
	} else {
		self.lastUsage = requestedDuration
		self.totalUsage += self.lastUsage
		ccDuration := self.extraDuration // fake ccDuration
		self.extraDuration -= dur
		return ccDuration, nil
	}
	initialExtraDuration := self.extraDuration
	self.extraDuration = 0
	if self.cd.LoopIndex > 0 {
		self.cd.TimeStart = self.cd.TimeEnd
	}
	self.cd.TimeEnd = self.cd.TimeStart.Add(dur)
	self.cd.DurationIndex += dur
	cc := &engine.CallCost{}
	if err := self.rater.Call("Responder.MaxDebit", self.cd, cc); err != nil {
		self.lastUsage = 0
		self.lastDebit = 0
		return 0, err
	}
	// cd corrections
	self.cd.TimeEnd = cc.GetEndTime() // set debited timeEnd
	// update call duration with real debited duration
	ccDuration := cc.GetDuration()
	if ccDuration != dur {
		self.extraDuration = ccDuration - dur
	}
	if ccDuration >= dur {
		self.lastUsage = requestedDuration
	} else {
		self.lastUsage = ccDuration
	}
	self.cd.DurationIndex -= dur
	self.cd.DurationIndex += ccDuration
	self.cd.MaxCostSoFar += cc.Cost
	self.cd.LoopIndex += 1
	self.sessionCds = append(self.sessionCds, self.cd.Clone())
	self.callCosts = append(self.callCosts, cc)
	self.lastDebit = initialExtraDuration + ccDuration
	self.totalUsage += self.lastUsage
	if ccDuration >= dur { // we got what we asked to be debited
		return requestedDuration, nil
	}
	return initialExtraDuration + ccDuration, nil
}

// Attempts to refund a duration, error on failure
func (self *SMGSession) refund(refundDuration time.Duration) error {
	if refundDuration == 0 { // Nothing to refund
		return nil
	}
	firstCC := self.callCosts[0] // use merged cc (from close function)
	firstCC.Timespans.Decompress()
	defer firstCC.Timespans.Compress()
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
	if len(refundIncrements) > 0 {
		cd := firstCC.CreateCallDescriptor()
		cd.Increments = refundIncrements
		cd.CgrID = self.cd.CgrID
		cd.RunID = self.cd.RunID
		cd.Increments.Compress()
		var response float64
		err := self.rater.Call("Responder.RefundIncrements", cd, &response)
		if err != nil {
			return err
		}
	}
	//firstCC.Cost -= refundIncrements.GetTotalCost() // use updateCost instead
	firstCC.UpdateCost()
	firstCC.UpdateRatedUsage()
	return nil
}

// Session has ended, check debits and refund the extra charged duration
func (self *SMGSession) close(endTime time.Time) error {
	if len(self.callCosts) != 0 { // We have had at least one cost calculation
		firstCC := self.callCosts[0]
		for _, cc := range self.callCosts[1:] {
			firstCC.Merge(cc)
		}
		end := firstCC.GetEndTime()
		refundDuration := end.Sub(endTime)
		self.refund(refundDuration)
	}
	return nil
}

// Send disconnect order to remote connection
func (self *SMGSession) disconnectSession(reason string) error {
	if self.clntConn == nil || reflect.ValueOf(self.clntConn).IsNil() {
		return errors.New("Calling SMGClientV1.DisconnectSession requires bidirectional JSON connection")
	}
	var reply string
	if err := self.clntConn.Call("SMGClientV1.DisconnectSession", utils.AttrDisconnectSession{EventStart: self.eventStart, Reason: reason}, &reply); err != nil {
		return err
	} else if reply != utils.OK {
		return errors.New(fmt.Sprintf("Unexpected disconnect reply: %s", reply))
	}
	return nil
}

// Merge the sum of costs and sends it to CDRS for storage
// originID could have been changed from original event, hence passing as argument here
// pass cc as the clone of original to avoid concurrency issues
func (self *SMGSession) saveOperations(originID string) error {
	if len(self.callCosts) == 0 {
		return nil // There are no costs to save, ignore the operation
	}
	cc := self.callCosts[0] // was merged in close method
	cc.Round()
	roundIncrements := cc.GetRoundIncrements()
	if len(roundIncrements) != 0 {
		cd := cc.CreateCallDescriptor()
		cd.CgrID = self.cd.CgrID
		cd.RunID = self.cd.RunID
		cd.Increments = roundIncrements
		var response float64
		if err := self.rater.Call("Responder.RefundRounding", cd, &response); err != nil {
			return err
		}
	}
	smCost := &engine.SMCost{
		CGRID:       self.eventStart.GetCgrId(self.timezone),
		CostSource:  utils.SESSION_MANAGER_SOURCE,
		RunID:       self.runId,
		OriginHost:  self.eventStart.GetOriginatorIP(utils.META_DEFAULT),
		OriginID:    originID,
		Usage:       self.TotalUsage().Seconds(),
		CostDetails: cc,
	}
	if len(smCost.CostDetails.Timespans) > MaxTimespansInCost { // Merge since we will get a callCost too big
		if err := utils.Clone(cc, &smCost.CostDetails); err != nil { // Avoid concurrency on CC
			utils.Logger.Err(fmt.Sprintf("<SMGeneric> Could not clone callcost for sessionID: %s, runId: %s, error: %s", originID, self.runId, err.Error()))
		}
		go func(smCost *engine.SMCost) { // could take longer than the locked stage
			if err := self.storeSMCost(smCost); err != nil {
				utils.Logger.Err(fmt.Sprintf("<SMGeneric> Could not store callcost for sessionID: %s, runId: %s, error: %s", originID, self.runId, err.Error()))
			}
		}(smCost)
	} else {
		return self.storeSMCost(smCost)
	}
	return nil
}

func (self *SMGSession) storeSMCost(smCost *engine.SMCost) error {
	if len(smCost.CostDetails.Timespans) > MaxTimespansInCost { // Merge so we can compress the CostDetails
		smCost.CostDetails.Timespans.Decompress()
		smCost.CostDetails.Timespans.Merge()
		smCost.CostDetails.Timespans.Compress()
	}
	var reply string
	if err := self.cdrsrv.Call("CdrsV1.StoreSMCost", engine.AttrCDRSStoreSMCost{Cost: smCost, CheckDuplicate: true}, &reply); err != nil {
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
	pdd, _ := self.eventStart.GetPdd(utils.META_DEFAULT)
	aSession := &ActiveSession{
		CgrId:       self.eventStart.GetCgrId(timezone),
		TOR:         self.eventStart.GetTOR(utils.META_DEFAULT),
		RunId:       self.runId,
		OriginID:    self.eventStart.GetUUID(),
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
		Usage:       self.TotalUsage(),
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
