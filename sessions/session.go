/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/

package sessions

import (
	"runtime"
	"sync"
	"time"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// SessionID is given by an agent as the answer to GetActiveSessionIDs API
type SessionID struct {
	OriginHost string
	OriginID   string
}

// OptsOriginID returns the OptsOriginID formated using the SessionID
func (s *SessionID) OptsOriginID() string {
	return utils.Sha1(s.OriginID, s.OriginHost)
}

// ExternalSession is used when displaying active sessions via RPC
type ExternalSession struct {
	ID                 string
	RunID              string
	CGREvent           *utils.CGREvent
	NodeID             string
	UsageAdjustment    *int64 // holds the extra usage either negative (ie. correction from consumed) or positive (ie. from roundingIncrements or correction)
	InterimUsage       *int64 // last requested Usage
	TotalUsage         *int64 // sum of InterimUsage
	TotalCost          float64
	AutoChargeInterval time.Duration // Enable auto-charging
	NextAutoCharge     *time.Time
	Charges            *utils.EventCharges
}

// NewSession is the constructor for one Session
func NewSession(origCGREv *utils.CGREvent, clientConnID string, runEvents []*utils.CGREvent) (s *Session) {
	s = &Session{
		ID:             utils.IfaceAsString(origCGREv.APIOpts[utils.MetaOriginID]),
		OriginCGREvent: origCGREv,
		ClientConnID:   clientConnID,
	}
	if runEvents != nil {
		s.SRuns = make([]*SRun, len(runEvents))
		for i, runEv := range runEvents {
			s.SRuns[i] = NewSRun(runEv)
		}
	}
	return
}

// Session is the main structure to describe a call
type Session struct {
	ID             string          // Unique identifier per Session, defaults to APIOpts[*cgrID]
	OriginCGREvent *utils.CGREvent // initial CGREvent received
	ClientConnID   string          // connection ID towards the client so we can recover from passive

	AutoChargeInterval time.Duration // Enable auto-charging
	NextAutoCharge     *time.Time

	SRuns []*SRun          // forked based on ChargerS
	sRuns map[string]*SRun // new way of indexing SRuns, should replace SRuns

	lk          sync.RWMutex
	debitStop   chan struct{}
	sTerminator *sTerminator // automatic timeout for the session

}

// Clone is a thread safe method to clone the sessions information
func (s *Session) Clone() (cln *Session) {
	s.lk.RLock()
	cln = &Session{
		ClientConnID:       s.ClientConnID,
		AutoChargeInterval: s.AutoChargeInterval,
	}
	if s.OriginCGREvent != nil {
		cln.OriginCGREvent = s.OriginCGREvent.Clone()
	}
	if s.SRuns != nil {
		cln.SRuns = make([]*SRun, len(s.SRuns))
		for i, sR := range s.SRuns {
			cln.SRuns[i] = sR.Clone()
		}
	}
	s.lk.RUnlock()
	return
}

// AsExternalSessions returns the session as a list of ExternalSession using all SRuns (thread safe)
func (s *Session) AsExternalSessions(tmz, nodeID string) (aSs []*ExternalSession) {
	s.lk.RLock()
	aSs = make([]*ExternalSession, len(s.SRuns))
	for i, sr := range s.SRuns {
		aSs[i] = &ExternalSession{
			ID:       s.ID,
			RunID:    sr.ID,
			CGREvent: sr.CGREvent,
			NodeID:   utils.EmptyString,
		}
	}
	for _, sr := range s.sRuns {
		eS := &ExternalSession{
			ID:                 s.ID,
			RunID:              sr.ID,
			CGREvent:           sr.CGREvent,
			NodeID:             utils.EmptyString,
			AutoChargeInterval: s.AutoChargeInterval,
			NextAutoCharge:     s.NextAutoCharge,
			Charges:            sr.Charges.Clone(),
		}
		if sr.UsageAdjustment != nil {
			i, _ := sr.UsageAdjustment.Big.Int64()
			eS.UsageAdjustment = utils.Int64Pointer(i)
		}
		if sr.InterimUsage != nil {
			i, _ := sr.InterimUsage.Big.Int64()
			eS.InterimUsage = utils.Int64Pointer(i)
		}
		if sr.TotalUsage != nil {
			i, _ := sr.TotalUsage.Big.Int64()
			eS.TotalUsage = utils.Int64Pointer(i)
		}
		aSs = append(aSs, eS)
	}
	s.lk.RUnlock()
	return
}

// AsExternalSession returns the session as an ExternalSession using the SRuns given
func (s *Session) AsExternalSession(sRunIdx int, nodeID string) (aS *ExternalSession) {
	aS = &ExternalSession{
		ID:       s.ID,
		RunID:    s.SRuns[sRunIdx].ID,
		CGREvent: s.SRuns[sRunIdx].CGREvent,
		NodeID:   nodeID,
	}
	return
}

// AsCGREvents is a  method to return the Session as CGREvents
// AsCGREvents is not thread safe since it is supposed to run by the time Session is closed
func (s *Session) asCGREvents() (cgrEvs []*utils.CGREvent) {
	cgrEvs = make([]*utils.CGREvent, len(s.SRuns)) // so we can gather all cdr info while under lock
	for i, sr := range s.SRuns {
		cgrEvs[i] = sr.CGREvent
	}
	return
}

// asCGREventsMap returns a map of all SRuns
// asCGREventsMap is not thread safe
func (s *Session) asCGREventsMap() (cgrEvs map[string]*utils.CGREvent) {
	cgrEvs = make(map[string]*utils.CGREvent, len(s.sRuns)) // so we can gather all cdr info while under lock
	for runID, sr := range s.sRuns {
		cgrEvs[runID] = sr.CGREvent
	}
	return
}

// stopSTerminator clears the session terminator
func (s *Session) stopSTerminator() {
	if s.sTerminator == nil ||
		s.sTerminator.endChan == nil {
		return
	}
	close(s.sTerminator.endChan)
	s.sTerminator.endChan = nil
}

// stopDebitLoops will stop all the active debits on the session
func (s *Session) stopDebitLoops() {
	if s.debitStop != nil {
		close(s.debitStop) // Stop automatic debits
		runtime.Gosched()
		s.debitStop = nil
	}
}

func NewSRun(cgrEv *utils.CGREvent) *SRun {
	return &SRun{
		ID:       utils.IfaceAsString(cgrEv.APIOpts[utils.MetaRunID]),
		CGREvent: cgrEv,
	}
}

// SRun is one billing run for the Session
type SRun struct {
	ID              string              // Identifier of the SRun, inherited from CGREvent.APIOpts[*runID]
	CGREvent        *utils.CGREvent     // Event received from ChargerS
	InterimUsage    *utils.Decimal      // last requested Usage
	UsageAdjustment *utils.Decimal      // holds the extra usage either negative (ie. correction from consumed) or positive (ie. from roundingIncrements or correction)
	TotalUsage      *utils.Decimal      // sum of InterimUsage
	Charges         *utils.EventCharges // list of charges this session run has performed

	lclDebit *utils.Decimal // last positive adjustment done, treated as local debit
}

// Clone returns the cloned version of SRun
func (sr *SRun) Clone() (clsr *SRun) {
	clsr = &SRun{
		ID:       sr.ID,
		CGREvent: sr.CGREvent.Clone(),
	}
	return
}

// updateUsages will consider all the usage opts and update SRun counters acordingly
func (sr *SRun) updateUsages(interimConsumed, interimUsage, totalUsage *utils.Decimal) error {
	sr.lclDebit = nil // this is always valid for one run only

	// corect the UsageAdjustment out of consumed
	if interimConsumed != nil && sr.InterimUsage != nil { // correct if InterimUsage was previously recorded
		diffUsage := utils.SubstractDecimal(sr.InterimUsage, interimConsumed)
		if diffUsage.Compare(utils.NewDecimal(0, 0)) != 0 {
			sr.UsageAdjustment = utils.SumDecimal(sr.UsageAdjustment, diffUsage)
			sr.TotalUsage = utils.SumDecimal(utils.SubstractDecimal(sr.TotalUsage, sr.InterimUsage), interimConsumed)
			sr.InterimUsage = interimConsumed
		}
	}
	// usage out of interimUsage
	if totalUsage != nil { // totalUsage should give us the interimUsage, if interimUsage also present, this will simply sum up since we want to debit also in advance
		interimUsage = utils.SubstractDecimal(utils.SumDecimal(totalUsage, interimUsage), sr.TotalUsage) // correct also interimUsage if totalUsage is present since this will be our debit
	}
	usage := utils.NewDecimal(0, 0)
	if interimUsage != nil {
		usage = interimUsage
	}

	// Appying UsageAdjustment to Usage
	if sr.UsageAdjustment != nil && sr.UsageAdjustment.Compare(utils.NewDecimal(0, 0)) != 0 {
		usage = utils.SubstractDecimal(usage, sr.UsageAdjustment)
		sr.lclDebit = sr.UsageAdjustment
		sr.UsageAdjustment = utils.NewDecimal(0, 0) // have consumed all out of UsageAdjustment
	}
	if usage.Compare(utils.NewDecimal(0, 0)) == -1 { // debit was done out of UsageAdjustment, no need of further debit
		sr.UsageAdjustment = utils.SumDecimal(sr.UsageAdjustment, utils.AbsoluteDecimal(usage)) // put back the extra units for next debit
		usage = utils.NewDecimal(0, 0)
		sr.lclDebit = utils.SubstractDecimal(sr.lclDebit, sr.UsageAdjustment) // correct the localDebit
	}
	// Save the interim and totalUsage
	sr.InterimUsage = interimUsage
	sr.TotalUsage = utils.SumDecimal(sr.TotalUsage, interimUsage)
	// Save the usage in SRuns so they can be debitted
	sr.CGREvent.APIOpts[utils.MetaUsage] = usage
	return nil
}

// updateSRuns updates the SRuns event with the alterable fields (is not thread safe)
func (s *Session) updateSRuns(updEv engine.MapEvent, alterableFields utils.StringSet) {
	if alterableFields.Size() == 0 {
		return
	}
	for k, v := range updEv {
		if !alterableFields.Has(k) {
			continue
		}
		for _, sr := range s.SRuns {
			sr.CGREvent.Event[k] = v
		}
		for _, sr := range s.sRuns { // Update the *new* approach
			sr.CGREvent.Event[k] = v
		}
	}
}
