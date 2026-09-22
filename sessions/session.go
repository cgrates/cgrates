// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package sessions

import (
	"fmt"
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

	SRuns []*SRun          // forked based on ChargerS
	sRuns map[string]*SRun // new way of indexing SRuns, should replace SRuns

	lk          sync.RWMutex
	sTerminator *sTerminator // automatic timeout for the session
}

// Clone is a thread safe method to clone the sessions information
func (s *Session) Clone() (cln *Session) {
	s.lk.RLock()
	cln = &Session{
		ClientConnID: s.ClientConnID,
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
	if s.sRuns != nil {
		cln.sRuns = make(map[string]*SRun)
		for rID, sR := range s.sRuns {
			cln.sRuns[rID] = sR.Clone()
		}
	}
	s.lk.RUnlock()
	return
}

// AsExternalSessions returns the session as a list of ExternalSession using all SRuns (thread safe)
func (s *Session) AsExternalSessions(tmz, nodeID string) (aSs []*ExternalSession) {
	s.lk.RLock()
	aSs = make([]*ExternalSession, 0, len(s.sRuns))
	for _, sr := range s.sRuns {
		aSs = append(aSs, sr.AsExternalSession(s.ID, nodeID))
	}
	s.lk.RUnlock()
	return
}

// AsExternalSession returns SRun as an ExternalSession
func (sr *SRun) AsExternalSession(sID, nodeID string) (eS *ExternalSession) {
	eS = &ExternalSession{
		ID:       sID,
		RunID:    sr.ID,
		CGREvent: sr.CGREvent,
		NodeID:   nodeID,
		Charges:  sr.Charges.Clone(),
	}
	if sr.UsageAdjustment != nil {
		v, _ := sr.UsageAdjustment.Big.Int64()
		eS.UsageAdjustment = new(v)
	}
	if sr.InterimUsage != nil {
		v, _ := sr.InterimUsage.Big.Int64()
		eS.InterimUsage = new(v)
	}
	if sr.TotalUsage != nil {
		v, _ := sr.TotalUsage.Big.Int64()
		eS.TotalUsage = new(v)
	}
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

/*
// stopDebitLoops will stop all the active debits on the session
func (s *Session) stopDebitLoops() {
	if s.debitStop != nil {
		close(s.debitStop) // Stop automatic debits
		runtime.Gosched()
		s.debitStop = nil
	}
}
*/

func NewSRun(cgrEv *utils.CGREvent) *SRun {
	return &SRun{
		ID:       utils.IfaceAsString(cgrEv.APIOpts[utils.MetaRunID]),
		CGREvent: cgrEv,
	}
}

// SRun is one billing run for the Session
type SRun struct {
	ID                 string              // Identifier of the SRun, inherited from CGREvent.APIOpts[*runID]
	CGREvent           *utils.CGREvent     // Event received from ChargerS
	InterimUsage       *utils.Decimal      // last requested Usage
	UsageAdjustment    *utils.Decimal      // holds the extra usage either negative (ie. correction from consumed) or positive (ie. from roundingIncrements or correction)
	TotalUsage         *utils.Decimal      // sum of InterimUsage
	Charges            *utils.EventCharges // list of charges this session run has performed
	AutoChargeInterval *time.Duration      // Enable auto-charging
	NextAutoCharge     *time.Time          // Save here the next auto-charge so we can continue on failover

	autoChargeStop chan struct{} // stop the autoCharge from outside by closing this channel, not nil means it was already started

	lclDebit  *utils.Decimal // last positive adjustment done, treated as local debit
	nextDebit *utils.Decimal // this specifies the amount should be debitted on next run
	lk        sync.RWMutex   // protects the SRun on concurrency
}

// Clone returns the cloned version of SRun
func (sr *SRun) Clone() (clsr *SRun) {
	if sr == nil {
		return nil
	}
	clsr = &SRun{
		ID:              sr.ID,
		CGREvent:        sr.CGREvent.Clone(),
		InterimUsage:    sr.InterimUsage.Clone(),
		UsageAdjustment: sr.UsageAdjustment.Clone(),
		TotalUsage:      sr.TotalUsage.Clone(),
		Charges:         sr.Charges.Clone(),
	}
	if sr.AutoChargeInterval != nil {
		d := *sr.AutoChargeInterval
		clsr.AutoChargeInterval = &d
	}
	if sr.NextAutoCharge != nil {
		d := *sr.NextAutoCharge
		clsr.NextAutoCharge = &d
	}
	return
}

// computeUsage will consider all the usage opts and update SRun counters acordingly
func (sr *SRun) computeUsages(interimConsumed, interimUsage, totalUsage *utils.Decimal) {
	utils.Logger.Info(fmt.Sprintf("### IN: interimConsumed: %s, interimUsage: %s, totalUsage: %s, sRun: %s\n", utils.ToIJSON(interimConsumed), utils.ToIJSON(interimUsage), utils.ToIJSON(totalUsage), utils.ToIJSON(sr)))
	sr.nextDebit = utils.NewDecimal(0, 0)
	sr.lclDebit = nil // this are always valid for one run only

	// corect the UsageAdjustment out of consumed
	if interimConsumed != nil && sr.InterimUsage != nil { // correct if InterimUsage was previously recorded
		diffUsage := utils.SubstractDecimal(sr.InterimUsage, interimConsumed)
		if diffUsage.Compare(utils.NewDecimal(0, 0)) != 0 {
			sr.UsageAdjustment = utils.SumDecimal(sr.UsageAdjustment, diffUsage)
			sr.TotalUsage = utils.SumDecimal(utils.SubstractDecimal(sr.TotalUsage, sr.InterimUsage), interimConsumed)
			sr.InterimUsage = interimConsumed.Clone()
		}
	}
	// usage out of interimUsage
	if totalUsage != nil { // totalUsage should give us the interimUsage, if interimUsage also present, this will simply sum up since we want to debit also in advance
		interimUsage = utils.SubstractDecimal(utils.SumDecimal(totalUsage, interimUsage), sr.TotalUsage) // correct also interimUsage if totalUsage is present since this will be our debit
	}

	if interimUsage != nil {
		sr.nextDebit = interimUsage.Clone()
	}

	// Appying UsageAdjustment to Usage
	if sr.UsageAdjustment != nil && sr.UsageAdjustment.Compare(utils.NewDecimal(0, 0)) != 0 {
		sr.nextDebit = utils.SubstractDecimal(sr.nextDebit, sr.UsageAdjustment)
		sr.lclDebit = sr.UsageAdjustment.Clone()
		sr.UsageAdjustment = utils.NewDecimal(0, 0) // have consumed all out of UsageAdjustment
	}
	if sr.nextDebit.Compare(utils.NewDecimal(0, 0)) == -1 { // debit was done out of UsageAdjustment, no need of further debit
		sr.UsageAdjustment = utils.SumDecimal(sr.UsageAdjustment, utils.AbsoluteDecimal(sr.nextDebit)) // put back the extra units for next debit
		sr.nextDebit = utils.NewDecimal(0, 0)
		sr.lclDebit = utils.SubstractDecimal(sr.lclDebit, sr.UsageAdjustment) // correct the localDebit
	}
	// Save the interim and totalUsage
	sr.InterimUsage = interimUsage.Clone()
	sr.TotalUsage = utils.SumDecimal(sr.TotalUsage, interimUsage)
	utils.Logger.Info(fmt.Sprintf("### OUT: interimConsumed: %s, interimUsage: %s, totalUsage: %s, localDebit: %s, sRun: %s\n", utils.ToIJSON(interimConsumed), utils.ToIJSON(interimUsage), utils.ToIJSON(totalUsage), utils.ToIJSON(sr.lclDebit), utils.ToIJSON(sr)))

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

// setSRun will create/update a single run with the data received within CGREvent
func (s *Session) setSRun(runID string, cgrEv *utils.CGREvent, alterableFields utils.StringSet, cch map[string]any,
	interimConsumed, interimUsage, totalUsage *utils.Decimal) (has bool, err error) {

	if _, has = s.sRuns[runID]; !has {
		s.sRuns[runID] = &SRun{
			ID:       runID,
			CGREvent: cgrEv,
		}
	}

	s.sRuns[runID].computeUsages(interimConsumed, interimUsage, totalUsage)

	return
}
