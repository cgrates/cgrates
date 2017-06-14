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
	"strconv"
	"sync"
	"time"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

// One session handled by SM
type SMGSession struct {
	mux       sync.RWMutex                  // protects the SMGSession in places where is concurrently accessed
	stopDebit chan struct{}                 // Channel to communicate with debit loops when closing the session
	clntConn  rpcclient.RpcClientConnection // Reference towards client connection on SMG side so we can disconnect.
	rals      rpcclient.RpcClientConnection // Connector to rals service
	cdrsrv    rpcclient.RpcClientConnection // Connector to CDRS service

	CGRID      string // Unique identifier for this session
	RunID      string // Keep a reference for the derived run
	Timezone   string
	EventStart SMGenericEvent         // Event which started the session
	CD         *engine.CallDescriptor // initial CD used for debits, updated on each debit

	EventCost     *engine.EventCost
	ExtraDuration time.Duration // keeps the current duration debited on top of what heas been asked
	LastUsage     time.Duration // last requested Duration
	LastDebit     time.Duration // last real debited duration
	TotalUsage    time.Duration // sum of lastUsage

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
				utils.Logger.Err(fmt.Sprintf("<SMGeneric> Could not complete debit operation on session: %s, error: %s", self.CGRID, err.Error()))
				disconnectReason := SYSTEM_ERROR
				if err.Error() == utils.ErrUnauthorizedDestination.Error() {
					disconnectReason = err.Error()
				}
				if err := self.disconnectSession(disconnectReason); err != nil {
					utils.Logger.Err(fmt.Sprintf("<SMGeneric> Could not disconnect session: %s, error: %s", self.CGRID, err.Error()))
				}
				return
			} else if maxDebit < debitInterval {
				time.Sleep(maxDebit)
				if err := self.disconnectSession(INSUFFICIENT_FUNDS); err != nil {
					utils.Logger.Err(fmt.Sprintf("<SMGeneric> Could not disconnect session: %s, error: %s", self.CGRID, err.Error()))
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
	self.mux.Lock()
	defer self.mux.Unlock()
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
	if err := self.rals.Call("Responder.MaxDebit", self.CD, cc); err != nil || cc.GetDuration() == 0 {
		self.LastUsage = 0
		self.LastDebit = 0
		return 0, err
	}
	// cd corrections
	self.CD.TimeEnd = cc.GetEndTime() // set debited timeEnd
	// update call duration with real debited duration
	ccDuration := cc.GetDuration()
	if ccDuration > dur {
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
	self.LastDebit = initialExtraDuration + ccDuration
	self.TotalUsage += self.LastUsage
	ec := engine.NewEventCostFromCallCost(cc, self.CGRID, self.RunID)
	if self.EventCost == nil {
		self.EventCost = ec
	} else {
		self.EventCost.Merge(ec)
	}
	if ccDuration < dur {
		return initialExtraDuration + ccDuration, nil
	}
	return requestedDuration, nil
}

// Send disconnect order to remote connection
func (self *SMGSession) disconnectSession(reason string) error {
	self.EventStart[utils.USAGE] = strconv.FormatFloat(self.TotalUsage.Seconds(), 'f', -1, 64) // Set the usage to total one debitted
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

// Session has ended, check debits and refund the extra charged duration
func (self *SMGSession) close(usage time.Duration) (err error) {
	self.mux.Lock()
	defer self.mux.Unlock()
	if self.EventCost == nil {
		return
	}
	if notCharged := usage - self.EventCost.GetUsage(); notCharged > 0 { // we did not charge enough, make a manual debit here
		if self.CD.LoopIndex > 0 {
			self.CD.TimeStart = self.CD.TimeEnd
		}
		self.CD.TimeEnd = self.CD.TimeStart.Add(notCharged)
		self.CD.DurationIndex += notCharged
		cc := &engine.CallCost{}
		if err = self.rals.Call("Responder.Debit", self.CD, cc); err == nil {
			self.EventCost.Merge(
				engine.NewEventCostFromCallCost(cc, self.CGRID, self.RunID))
		}
	} else if notCharged < 0 { // charged too much, try refund
		err = self.refund(usage)
	}

	return
}

// Attempts to refund a duration, error on failure
// usage represents the real usage
func (self *SMGSession) refund(usage time.Duration) (err error) {
	if self.EventCost == nil {
		return
	}
	srplsEC, err := self.EventCost.Trim(usage)
	if err != nil {
		return err
	}
	if srplsEC == nil {
		return
	}

	cc := srplsEC.AsCallCost()
	var incrmts engine.Increments
	for _, tmspn := range cc.Timespans {
		for _, incr := range tmspn.Increments {
			incrmts = append(incrmts, incr)
		}
	}
	cd := &engine.CallDescriptor{
		CgrID:       self.CGRID,
		RunID:       self.RunID,
		Direction:   self.CD.Direction,
		Category:    self.CD.Category,
		Tenant:      self.CD.Tenant,
		Subject:     self.CD.Subject,
		Account:     self.CD.Account,
		Destination: self.CD.Destination,
		TOR:         self.CD.TOR,
		Increments:  incrmts,
	}
	var reply float64
	return self.rals.Call("Responder.RefundIncrements", cd, &reply)
}

// storeSMCost will send the SMCost to CDRs for storing
func (self *SMGSession) storeSMCost() error {
	if self.EventCost == nil {
		return nil // There are no costs to save, ignore the operation
	}
	self.mux.Lock()
	self.mux.Unlock()
	smCost := &engine.V2SMCost{
		CGRID:       self.CGRID,
		CostSource:  utils.SESSION_MANAGER_SOURCE,
		RunID:       self.RunID,
		OriginHost:  self.EventStart.GetOriginatorIP(utils.META_DEFAULT),
		OriginID:    self.EventStart.GetOriginID(utils.META_DEFAULT),
		Usage:       self.TotalUsage.Seconds(),
		CostDetails: self.EventCost,
	}
	var reply string
	if err := self.cdrsrv.Call("CdrsV2.StoreSMCost", engine.ArgsV2CDRSStoreSMCost{Cost: smCost,
		CheckDuplicate: true}, &reply); err != nil {
		if err == utils.ErrExists {
			self.refund(self.CD.GetDuration()) // Refund entire duration
		} else {
			return err
		}
	}
	return nil
}

func (self *SMGSession) AsActiveSession(timezone string) *ActiveSession {
	self.mux.RLock()
	defer self.mux.RUnlock()
	sTime, _ := self.EventStart.GetSetupTime(utils.META_DEFAULT, timezone)
	aTime, _ := self.EventStart.GetAnswerTime(utils.META_DEFAULT, timezone)
	pdd, _ := self.EventStart.GetPdd(utils.META_DEFAULT)
	aSession := &ActiveSession{
		CGRID:       self.CGRID,
		TOR:         self.EventStart.GetTOR(utils.META_DEFAULT),
		RunID:       self.RunID,
		OriginID:    self.EventStart.GetOriginID(utils.META_DEFAULT),
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
