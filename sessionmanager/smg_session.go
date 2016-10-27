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
	EventStart    SMGenericEvent // Event which started
	stopDebit     chan struct{}  // Channel to communicate with debit loops when closing the session
	RunID         string         // Keep a reference for the derived run
	Timezone      string
	CD            *engine.CallDescriptor
	SessionCDs    []*engine.CallDescriptor
	CallCosts     []*engine.CallCost
	ExtraDuration time.Duration                 // keeps the current duration debited on top of what heas been asked
	LastUsage     time.Duration                 // last requested Duration
	LastDebit     time.Duration                 // last real debited duration
	TotalUsage    time.Duration                 // sum of lastUsage
	clntConn      rpcclient.RpcClientConnection // Reference towards client connection on SMG side so we can disconnect.
	rater         rpcclient.RpcClientConnection // Connector to Rater service
	cdrsrv        rpcclient.RpcClientConnection // Connector to CDRS service
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
				utils.Logger.Err(fmt.Sprintf("<SMGeneric> Could not complete debit operation on session: %s, error: %s", self.EventStart.GetUUID(), err.Error()))
				disconnectReason := SYSTEM_ERROR
				if err.Error() == utils.ErrUnauthorizedDestination.Error() {
					disconnectReason = err.Error()
				}
				if err := self.disconnectSession(disconnectReason); err != nil {
					utils.Logger.Err(fmt.Sprintf("<SMGeneric> Could not disconnect session: %s, error: %s", self.EventStart.GetUUID(), err.Error()))
				}
				return
			} else if maxDebit < debitInterval {
				time.Sleep(maxDebit)
				if err := self.disconnectSession(INSUFFICIENT_FUNDS); err != nil {
					utils.Logger.Err(fmt.Sprintf("<SMGeneric> Could not disconnect session: %s, error: %s", self.EventStart.GetUUID(), err.Error()))
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
		self.ExtraDuration = self.LastDebit - *lastUsed
		if *lastUsed != self.LastUsage {
			// total usage correction
			self.TotalUsage -= self.LastUsage
			self.TotalUsage += *lastUsed
		}
	}
	// apply correction from previous run
	if self.ExtraDuration < dur {
		dur -= self.ExtraDuration
	} else {
		self.LastUsage = requestedDuration
		self.TotalUsage += self.LastUsage
		ccDuration := self.ExtraDuration // fake ccDuration
		self.ExtraDuration -= dur
		return ccDuration, nil
	}
	initialExtraDuration := self.ExtraDuration
	self.ExtraDuration = 0
	if self.CD.LoopIndex > 0 {
		self.CD.TimeStart = self.CD.TimeEnd
	}
	self.CD.TimeEnd = self.CD.TimeStart.Add(dur)
	self.CD.DurationIndex += dur
	cc := &engine.CallCost{}
	if err := self.rater.Call("Responder.MaxDebit", self.CD, cc); err != nil {
		self.LastUsage = 0
		self.LastDebit = 0
		return 0, err
	}
	// cd corrections
	self.CD.TimeEnd = cc.GetEndTime() // set debited timeEnd
	// update call duration with real debited duration
	ccDuration := cc.GetDuration()
	if ccDuration != dur {
		self.ExtraDuration = ccDuration - dur
	}
	if ccDuration >= dur {
		self.LastUsage = requestedDuration
	} else {
		self.LastUsage = ccDuration
	}
	self.CD.DurationIndex -= dur
	self.CD.DurationIndex += ccDuration
	self.CD.MaxCostSoFar += cc.Cost
	self.CD.LoopIndex += 1
	self.SessionCDs = append(self.SessionCDs, self.CD.Clone())
	self.CallCosts = append(self.CallCosts, cc)
	self.LastDebit = initialExtraDuration + ccDuration
	self.TotalUsage += self.LastUsage
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
	firstCC := self.CallCosts[0] // use merged cc (from close function)
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
		cd.CgrID = self.CD.CgrID
		cd.RunID = self.CD.RunID
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
	if len(self.CallCosts) != 0 { // We have had at least one cost calculation
		firstCC := self.CallCosts[0]
		for _, cc := range self.CallCosts[1:] {
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
	if err := self.clntConn.Call("SMGClientV1.DisconnectSession", utils.AttrDisconnectSession{EventStart: self.EventStart, Reason: reason}, &reply); err != nil {
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
	if len(self.CallCosts) == 0 {
		return nil // There are no costs to save, ignore the operation
	}
	cc := self.CallCosts[0] // was merged in close method
	cc.Round()
	roundIncrements := cc.GetRoundIncrements()
	if len(roundIncrements) != 0 {
		cd := cc.CreateCallDescriptor()
		cd.CgrID = self.CD.CgrID
		cd.RunID = self.CD.RunID
		cd.Increments = roundIncrements
		var response float64
		if err := self.rater.Call("Responder.RefundRounding", cd, &response); err != nil {
			return err
		}
	}
	smCost := &engine.SMCost{
		CGRID:       self.EventStart.GetCgrId(self.Timezone),
		CostSource:  utils.SESSION_MANAGER_SOURCE,
		RunID:       self.RunID,
		OriginHost:  self.EventStart.GetOriginatorIP(utils.META_DEFAULT),
		OriginID:    originID,
		Usage:       self.TotalUsage.Seconds(),
		CostDetails: cc,
	}
	if len(smCost.CostDetails.Timespans) > MaxTimespansInCost { // Merge since we will get a callCost too big
		if err := utils.Clone(cc, &smCost.CostDetails); err != nil { // Avoid concurrency on CC
			utils.Logger.Err(fmt.Sprintf("<SMGeneric> Could not clone callcost for sessionID: %s, RunID: %s, error: %s", originID, self.RunID, err.Error()))
		}
		go func(smCost *engine.SMCost) { // could take longer than the locked stage
			if err := self.storeSMCost(smCost); err != nil {
				utils.Logger.Err(fmt.Sprintf("<SMGeneric> Could not store callcost for sessionID: %s, RunID: %s, error: %s", originID, self.RunID, err.Error()))
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
			self.refund(self.CD.GetDuration()) // Refund entire duration
		} else {
			return err
		}
	}
	return nil
}

func (self *SMGSession) AsActiveSession(timezone string) *ActiveSession {
	sTime, _ := self.EventStart.GetSetupTime(utils.META_DEFAULT, timezone)
	aTime, _ := self.EventStart.GetAnswerTime(utils.META_DEFAULT, timezone)
	pdd, _ := self.EventStart.GetPdd(utils.META_DEFAULT)
	aSession := &ActiveSession{
		CgrId:       self.EventStart.GetCgrId(timezone),
		TOR:         self.EventStart.GetTOR(utils.META_DEFAULT),
		RunID:       self.RunID,
		OriginID:    self.EventStart.GetUUID(),
		CdrHost:     self.EventStart.GetOriginatorIP(utils.META_DEFAULT),
		CdrSource:   self.EventStart.GetCdrSource(),
		ReqType:     self.EventStart.GetReqType(utils.META_DEFAULT),
		Direction:   self.EventStart.GetDirection(utils.META_DEFAULT),
		Tenant:      self.EventStart.GetTenant(utils.META_DEFAULT),
		Category:    self.EventStart.GetCategory(utils.META_DEFAULT),
		Account:     self.EventStart.GetAccount(utils.META_DEFAULT),
		Subject:     self.EventStart.GetSubject(utils.META_DEFAULT),
		Destination: self.EventStart.GetDestination(utils.META_DEFAULT),
		SetupTime:   sTime,
		AnswerTime:  aTime,
		Usage:       self.TotalUsage,
		Pdd:         pdd,
		ExtraFields: self.EventStart.GetExtraFields(),
		Supplier:    self.EventStart.GetSupplier(utils.META_DEFAULT),
		SMId:        "CGR-DA",
	}
	if self.CD != nil {
		aSession.LoopIndex = self.CD.LoopIndex
		aSession.DurationIndex = self.CD.DurationIndex
		aSession.MaxRate = self.CD.MaxRate
		aSession.MaxRateUnit = self.CD.MaxRateUnit
		aSession.MaxCostSoFar = self.CD.MaxCostSoFar
	}
	return aSession
}
