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
	ID            string
	RunID         string
	CGREvent      *utils.CGREvent
	NodeID        string
	TotalUsage    time.Duration // the call duration so far (till TimeEnd)
	TotalCost     float64
	DebitInterval time.Duration
	NextAutoDebit time.Time
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
	ID             string          // Unique identifier per Session, defaults to APIOpts[*originID]
	OriginCGREvent *utils.CGREvent // initial CGREvent received
	ClientConnID   string          // connection ID towards the client so we can recover from passive
	DebitInterval  *time.Duration  // execute debits for *prepaid runs
	SRuns          []*SRun         // forked based on ChargerS

	lk          sync.RWMutex
	debitStop   chan struct{}
	sTerminator *sTerminator // automatic timeout for the session
}

// Clone is a thread safe method to clone the sessions information
func (s *Session) Clone() (cln *Session) {
	s.lk.RLock()
	cln = &Session{
		OriginCGREvent: s.OriginCGREvent.Clone(),
		ClientConnID:   s.ClientConnID,
		DebitInterval:  s.DebitInterval,
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

		if sr.NextAutoDebit != nil {
			aSs[i].NextAutoDebit = *sr.NextAutoDebit
		}
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
	if s.SRuns[sRunIdx].NextAutoDebit != nil {
		aS.NextAutoDebit = *s.SRuns[sRunIdx].NextAutoDebit
	}
	return
}

// totalUsage returns the first session run total usage
// not thread save
func (s *Session) totalUsage() (tDur time.Duration) {
	if len(s.SRuns) == 0 {
		return
	}
	for _, sr := range s.SRuns {
		tDur = sr.TotalUsage
		break // only first
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
	ID       string          // Identifier of the SRun, inherited from CGREvent.APIOpts[*runID]
	CGREvent *utils.CGREvent // Event received from ChargerS

	ExtraUsage    time.Duration // keeps the extra usage debited on top of what has been asked
	LastUsage     time.Duration // last requested Duration
	TotalUsage    time.Duration // sum of lastUsage
	NextAutoDebit *time.Time
}

// Clone returns the cloned version of SRun
func (sr *SRun) Clone() (clsr *SRun) {
	clsr = &SRun{
		ID:         sr.ID,
		CGREvent:   sr.CGREvent.Clone(),
		ExtraUsage: sr.ExtraUsage,
		LastUsage:  sr.LastUsage,
		TotalUsage: sr.TotalUsage,
	}
	if sr.NextAutoDebit != nil {
		clsr.NextAutoDebit = utils.TimePointer(*sr.NextAutoDebit)
	}
	return
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
	}
}
