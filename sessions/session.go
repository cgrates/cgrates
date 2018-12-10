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

package sessions

import (
	"errors"
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

// One session handled by SM
type SMGSession struct {
	sync.RWMutex                               // protects the SMGSession in places where is concurrently accessed
	stopDebit    chan struct{}                 // Channel to communicate with debit loops when closing the session
	clntConn     rpcclient.RpcClientConnection // Reference towards client connection on SMG side so we can disconnect.
	rals         rpcclient.RpcClientConnection // Connector to rals service
	cdrsrv       rpcclient.RpcClientConnection // Connector to CDRS service
	clientProto  float64

	Tenant     string // store original Tenant so we can use it in API calls
	CGRID      string // Unique identifier for this session
	RunID      string // Keep a reference for the derived run
	Timezone   string
	ResourceID string

	EventStart *engine.SafEvent       // Event which started the session
	CD         *engine.CallDescriptor // initial CD used for debits, updated on each debit
	EventCost  *engine.EventCost

	ExtraDuration time.Duration // keeps the current duration debited on top of what has been asked
	LastUsage     time.Duration // last requested Duration
	LastDebit     time.Duration // last real debited duration
	TotalUsage    time.Duration // sum of lastUsage
}

// Clone returns the cloned version of SMGSession
func (s *SMGSession) Clone() *SMGSession {
	return &SMGSession{CGRID: s.CGRID, RunID: s.RunID,
		Timezone: s.Timezone, ResourceID: s.ResourceID,
		EventStart:    s.EventStart.Clone(),
		CD:            s.CD.Clone(),
		EventCost:     s.EventCost.Clone(),
		ExtraDuration: s.ExtraDuration, LastUsage: s.LastUsage,
		LastDebit: s.LastDebit, TotalUsage: s.TotalUsage,
	}
}

type SessionID struct {
	OriginHost string
	OriginID   string
}

func (s *SessionID) CGRID() string {
	return utils.Sha1(s.OriginID, s.OriginHost)
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
				utils.Logger.Err(fmt.Sprintf("<%s> Could not complete debit operation on session: %s, error: %s", utils.SessionS, self.CGRID, err.Error()))
				disconnectReason := utils.ErrServerError.Error()
				if err.Error() == utils.ErrUnauthorizedDestination.Error() {
					disconnectReason = err.Error()
				}
				if err := self.disconnectSession(disconnectReason); err != nil {
					utils.Logger.Err(fmt.Sprintf("<%s> Could not disconnect session: %s, error: %s", utils.SessionS, self.CGRID, err.Error()))
				}
				return
			} else if maxDebit < debitInterval {
				time.Sleep(maxDebit)
				if err := self.disconnectSession(utils.ErrInsufficientCredit.Error()); err != nil {
					utils.Logger.Err(fmt.Sprintf("<%s> Could not disconnect session: %s, error: %s", utils.SessionS, self.CGRID, err.Error()))
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
	self.Lock()
	defer self.Unlock()
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
		self.ExtraDuration -= dur
		return requestedDuration, nil
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
	if self.clntConn == nil || reflect.ValueOf(self.clntConn).IsNil() {
		return errors.New("Calling SMGClientV1.DisconnectSession requires bidirectional JSON connection")
	}
	self.EventStart.Set(utils.Usage, self.TotalUsage) // Set the usage to total one debitted
	var reply string
	servMethod := "SessionSv1.DisconnectSession"
	if self.clientProto == 0 { // competibility with OpenSIPS
		servMethod = "SMGClientV1.DisconnectSession"
	}
	if err := self.clntConn.Call(servMethod,
		utils.AttrDisconnectSession{EventStart: self.EventStart.AsMapInterface(),
			Reason: reason},
		&reply); err != nil {
		return err
	} else if reply != utils.OK {
		return errors.New(fmt.Sprintf("Unexpected disconnect reply: %s", reply))
	}
	return nil
}

// Session has ended, check debits and refund the extra charged duration
func (self *SMGSession) close(usage time.Duration) (err error) {
	self.Lock()
	defer self.Unlock()
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
			if incr.BalanceInfo == nil ||
				(incr.BalanceInfo.Unit == nil &&
					incr.BalanceInfo.Monetary == nil) {
				continue // not enough information for refunds, most probably free units uncounted
			}
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
	var acnt engine.Account
	err = self.rals.Call("Responder.RefundIncrements", cd, &acnt)
	if acnt.ID != "" { // Account info updated, update also cached AccountSummary
		self.EventCost.AccountSummary = acnt.AsAccountSummary()
	}
	return
}

// storeSMCost will send the SMCost to CDRs for storing
func (self *SMGSession) storeSMCost() error {
	if self.EventCost == nil {
		return nil // There are no costs to save, ignore the operation
	}
	self.Lock()
	self.Unlock()
	smCost := &engine.V2SMCost{
		CGRID:       self.CGRID,
		CostSource:  utils.MetaSessionS,
		RunID:       self.RunID,
		OriginHost:  self.EventStart.GetStringIgnoreErrors(utils.OriginHost),
		OriginID:    self.EventStart.GetStringIgnoreErrors(utils.OriginID),
		Usage:       self.TotalUsage,
		CostDetails: self.EventCost,
	}
	var reply string
	if err := self.cdrsrv.Call("CdrsV2.StoreSMCost",
		engine.ArgsV2CDRSStoreSMCost{Cost: smCost,
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
	self.RLock()
	defer self.RUnlock()
	aSession := &ActiveSession{
		CGRID:       self.CGRID,
		TOR:         self.EventStart.GetStringIgnoreErrors(utils.ToR),
		RunID:       self.RunID,
		OriginID:    self.EventStart.GetStringIgnoreErrors(utils.OriginID),
		CdrHost:     self.EventStart.GetStringIgnoreErrors(utils.OriginHost),
		CdrSource:   utils.SessionS + "_" + self.EventStart.GetStringIgnoreErrors(utils.EVENT_NAME),
		ReqType:     self.EventStart.GetStringIgnoreErrors(utils.RequestType),
		Tenant:      self.EventStart.GetStringIgnoreErrors(utils.Tenant),
		Category:    self.EventStart.GetStringIgnoreErrors(utils.Category),
		Account:     self.EventStart.GetStringIgnoreErrors(utils.Account),
		Subject:     self.EventStart.GetStringIgnoreErrors(utils.Subject),
		Destination: self.EventStart.GetStringIgnoreErrors(utils.Destination),
		SetupTime:   self.EventStart.GetTimeIgnoreErrors(utils.SetupTime, self.Timezone),
		AnswerTime:  self.EventStart.GetTimeIgnoreErrors(utils.AnswerTime, self.Timezone),
		Usage:       self.TotalUsage,
		ExtraFields: self.EventStart.AsMapStringIgnoreErrors(utils.NewStringMap(utils.PrimaryCdrFields...)),
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

// Will be used when displaying active sessions via RPC
type ActiveSession struct {
	CGRID         string
	TOR           string            // type of record, meta-field, should map to one of the TORs hardcoded inside the server <*voice|*data|*sms|*generic>
	OriginID      string            // represents the unique accounting id given by the telecom switch generating the CDR
	CdrHost       string            // represents the IP address of the host generating the CDR (automatically populated by the server)
	CdrSource     string            // formally identifies the source of the CDR (free form field)
	ReqType       string            // matching the supported request types by the **CGRateS**, accepted values are hardcoded in the server <prepaid|postpaid|pseudoprepaid|rated>
	Tenant        string            // tenant whom this record belongs
	Category      string            // free-form filter for this record, matching the category defined in rating profiles.
	Account       string            // account id (accounting subsystem) the record should be attached to
	Subject       string            // rating subject (rating subsystem) this record should be attached to
	Destination   string            // destination to be charged
	SetupTime     time.Time         // set-up time of the event. Supported formats: datetime RFC3339 compatible, SQL datetime (eg: MySQL), unix timestamp.
	AnswerTime    time.Time         // answer time of the event. Supported formats: datetime RFC3339 compatible, SQL datetime (eg: MySQL), unix timestamp.
	Usage         time.Duration     // event usage information (eg: in case of tor=*voice this will represent the total duration of a call)
	ExtraFields   map[string]string // Extra fields to be stored in CDR
	SMId          string
	SMConnId      string
	RunID         string
	LoopIndex     float64       // indicates the position of this segment in a cost request loop
	DurationIndex time.Duration // the call duration so far (till TimeEnd)
	MaxRate       float64
	MaxRateUnit   time.Duration
	MaxCostSoFar  float64
}
