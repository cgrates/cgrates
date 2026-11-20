// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package sessions

import (
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

type session struct {
	session *utils.Session

	lk          sync.RWMutex
	sTerminator *sTerminator // automatic timeout for the session
	runStates   map[string]*sRun
}

func (s *session) Clone() *utils.Session {
	s.lk.RLock()
	defer s.lk.RUnlock()
	return s.session.Clone()
}

// AsExternalSessions returns the session as a list of ExternalSession using all sRuns (thread safe)
func (s *session) AsExternalSessions(tmz, nodeID string) (aSs []*ExternalSession) {
	s.lk.RLock()
	aSs = make([]*ExternalSession, 0, len(s.session.SRuns))
	for runID := range s.session.SRuns {
		aSs = append(aSs, s.AsExternalSession(runID, nodeID))
	}
	s.lk.RUnlock()
	return
}

// AsExternalSession returns SRun as an ExternalSession
func (s *session) AsExternalSession(runID, nodeID string) (eS *ExternalSession) {
	sr := s.session.SRuns[runID]
	eS = &ExternalSession{
		ID:      s.session.ID,
		RunID:   sr.ID,
		NodeID:  nodeID,
		Charges: sr.Charges.Clone(),
	}
	if sr.CGREvent != nil {
		eS.CGREvent = sr.CGREvent.Clone()
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

// asCGREventsMap returns a map of all sRuns
// asCGREventsMap is not thread safe
func (s *session) asCGREventsMap() (cgrEvs map[string]*utils.CGREvent) {
	cgrEvs = make(map[string]*utils.CGREvent, len(s.session.SRuns)) // so we can gather all cdr info while under lock
	for runID, sr := range s.session.SRuns {
		cgrEvs[runID] = sr.CGREvent
	}
	return
}

// stopSTerminator keeps the deadline so expiry can resume if the session becomes active again.
func (s *session) stopSTerminator() {
	if s.sTerminator == nil ||
		s.sTerminator.endChan == nil {
		return
	}
	close(s.sTerminator.endChan)
	s.sTerminator.endChan = nil
}

// setSRun will create/update a single run with the data received within CGREvent
func (s *session) setSRun(runID string, cgrEv *utils.CGREvent, alterableFields, indexedFields utils.StringSet,
	cch map[string]any, interimConsumed, interimUsage, totalUsage *utils.Decimal) (has, reIdx bool, err error) {
	sr, has := s.session.SRuns[runID]
	if !has {
		sr = &utils.SRun{ID: runID, CGREvent: cgrEv.Clone()}
		s.session.SRuns[runID] = sr
	}
	// A restored run needs debit state only when it receives an update.
	run := s.runStates[runID]
	if run == nil {
		if s.runStates == nil {
			s.runStates = make(map[string]*sRun)
		}
		run = &sRun{run: sr}
		s.runStates[runID] = run
	}
	run.computeUsages(interimConsumed, interimUsage, totalUsage)
	if !has || len(alterableFields) == 0 {
		return has, false, nil
	}
	for k, v := range cgrEv.Event {
		if !alterableFields.Has(k) {
			continue
		}
		if indexedFields.Has(k) &&
			utils.IfaceAsString(sr.CGREvent.Event[k]) != utils.IfaceAsString(v) {
			reIdx = true
		}
		sr.CGREvent.Event[k] = v
	}
	return has, reIdx, nil
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

type sRun struct {
	run *utils.SRun

	autoChargeStop chan struct{} // stop the autoCharge from outside by closing this channel, not nil means it was already started

	lclDebit  *utils.Decimal // last positive adjustment done, treated as local debit
	nextDebit *utils.Decimal // this specifies the amount should be debitted on next run
	lk        sync.RWMutex   // protects the SRun on concurrency
}

// computeUsage will consider all the usage opts and update SRun counters acordingly
func (sr *sRun) computeUsages(interimConsumed, interimUsage, totalUsage *utils.Decimal) {
	// utils.Logger.Info(fmt.Sprintf("### IN: interimConsumed: %s, interimUsage: %s, totalUsage: %s, sRun: %s\n", utils.ToIJSON(interimConsumed), utils.ToIJSON(interimUsage), utils.ToIJSON(totalUsage), utils.ToIJSON(sr)))
	sr.nextDebit = utils.NewDecimal(0, 0)
	sr.lclDebit = nil // this are always valid for one run only

	// corect the UsageAdjustment out of consumed
	if interimConsumed != nil && sr.run.InterimUsage != nil { // correct if InterimUsage was previously recorded
		diffUsage := utils.SubstractDecimal(sr.run.InterimUsage, interimConsumed)
		if diffUsage.Compare(utils.NewDecimal(0, 0)) != 0 {
			sr.run.UsageAdjustment = utils.SumDecimal(sr.run.UsageAdjustment, diffUsage)
			sr.run.TotalUsage = utils.SumDecimal(utils.SubstractDecimal(sr.run.TotalUsage, sr.run.InterimUsage), interimConsumed)
			sr.run.InterimUsage = interimConsumed.Clone()
		}
	}
	// usage out of interimUsage
	if totalUsage != nil { // totalUsage should give us the interimUsage, if interimUsage also present, this will simply sum up since we want to debit also in advance
		interimUsage = utils.SubstractDecimal(utils.SumDecimal(totalUsage, interimUsage), sr.run.TotalUsage) // correct also interimUsage if totalUsage is present since this will be our debit
	}

	if interimUsage != nil {
		sr.nextDebit = interimUsage.Clone()
	}

	// Appying UsageAdjustment to Usage
	if sr.run.UsageAdjustment != nil && sr.run.UsageAdjustment.Compare(utils.NewDecimal(0, 0)) != 0 {
		sr.nextDebit = utils.SubstractDecimal(sr.nextDebit, sr.run.UsageAdjustment)
		sr.lclDebit = sr.run.UsageAdjustment.Clone()
		sr.run.UsageAdjustment = utils.NewDecimal(0, 0) // have consumed all out of UsageAdjustment
	}
	if sr.nextDebit.Compare(utils.NewDecimal(0, 0)) == -1 { // debit was done out of UsageAdjustment, no need of further debit
		sr.run.UsageAdjustment = utils.SumDecimal(sr.run.UsageAdjustment, utils.AbsoluteDecimal(sr.nextDebit)) // put back the extra units for next debit
		sr.nextDebit = utils.NewDecimal(0, 0)
		sr.lclDebit = utils.SubstractDecimal(sr.lclDebit, sr.run.UsageAdjustment) // correct the localDebit
	}
	// Save the interim and totalUsage
	sr.run.InterimUsage = interimUsage.Clone()
	sr.run.TotalUsage = utils.SumDecimal(sr.run.TotalUsage, interimUsage)
	// utils.Logger.Info(fmt.Sprintf("### OUT: interimConsumed: %s, interimUsage: %s, totalUsage: %s, localDebit: %s, sRun: %s\n", utils.ToIJSON(interimConsumed), utils.ToIJSON(interimUsage), utils.ToIJSON(totalUsage), utils.ToIJSON(sr.lclDebit), utils.ToIJSON(sr)))

}

// updateSRuns updates the sRuns event with the alterable fields (is not thread safe)
func (s *session) updateSRuns(updEv engine.MapEvent, alterableFields utils.StringSet) {
	if alterableFields.Size() == 0 {
		return
	}
	for k, v := range updEv {
		if !alterableFields.Has(k) {
			continue
		}
		for _, sr := range s.session.SRuns { // Update the *new* approach
			sr.CGREvent.Event[k] = v
		}
	}
}
