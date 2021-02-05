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

// CGRID returns the CGRID formated using the SessionID
func (s *SessionID) CGRID() string {
	return utils.Sha1(s.OriginID, s.OriginHost)
}

// ExternalSession is used when displaying active sessions via RPC
type ExternalSession struct {
	CGRID         string
	RunID         string
	ToR           string            // type of record, meta-field, should map to one of the TORs hardcoded inside the server <*voice|*data|*sms|*generic>
	OriginID      string            // represents the unique accounting id given by the telecom switch generating the CDR
	OriginHost    string            // represents the IP address of the host generating the CDR (automatically populated by the server)
	Source        string            // formally identifies the source of the CDR (free form field)
	RequestType   string            // matching the supported request types by the **CGRateS**, accepted values are hardcoded in the server <prepaid|postpaid|pseudoprepaid|rated>
	Tenant        string            // tenant whom this record belongs
	Category      string            // free-form filter for this record, matching the category defined in rating profiles.
	Account       string            // account id (accounting subsystem) the record should be attached to
	Subject       string            // rating subject (rating subsystem) this record should be attached to
	Destination   string            // destination to be charged
	SetupTime     time.Time         // set-up time of the event. Supported formats: datetime RFC3339 compatible, SQL datetime (eg: MySQL), unix timestamp.
	AnswerTime    time.Time         // answer time of the event. Supported formats: datetime RFC3339 compatible, SQL datetime (eg: MySQL), unix timestamp.
	Usage         time.Duration     // event usage information (eg: in case of tor=*voice this will represent the total duration of a call)
	ExtraFields   map[string]string // Extra fields to be stored in CDR
	NodeID        string
	LoopIndex     float64       // indicates the position of this segment in a cost request loop
	DurationIndex time.Duration // the call duration so far (till TimeEnd)
	MaxRate       float64
	MaxRateUnit   time.Duration
	MaxCostSoFar  float64
	DebitInterval time.Duration
	NextAutoDebit time.Time
}

// Session is the main structure to describe a call
type Session struct {
	lk sync.RWMutex

	CGRID         string
	Tenant        string
	ResourceID    string
	ClientConnID  string          // connection ID towards the client so we can recover from passive
	EventStart    engine.MapEvent // Event which started the session
	DebitInterval time.Duration   // execute debits for *prepaid runs
	SRuns         []*SRun         // forked based on ChargerS
	OptsStart     engine.MapEvent

	debitStop   chan struct{}
	sTerminator *sTerminator // automatic timeout for the session
}

// Lock exported function from sync.RWMutex
func (s *Session) Lock() {
	s.lk.Lock()
}

// Unlock exported function from sync.RWMutex
func (s *Session) Unlock() {
	s.lk.Unlock()
}

// RLock exported function from sync.RWMutex
func (s *Session) RLock() {
	s.lk.RLock()
}

// RUnlock exported function from sync.RWMutex
func (s *Session) RUnlock() {
	s.lk.RUnlock()
}

// cgrID is method to return the CGRID of a session
// not thread safe
func (s *Session) cgrID() (cgrID string) {
	cgrID = s.CGRID
	return
}

// Clone is a thread safe method to clone the sessions information
func (s *Session) Clone() (cln *Session) {
	s.RLock()
	cln = &Session{
		CGRID:         s.CGRID,
		Tenant:        s.Tenant,
		ResourceID:    s.ResourceID,
		ClientConnID:  s.ClientConnID,
		EventStart:    s.EventStart.Clone(),
		DebitInterval: s.DebitInterval,
	}
	if s.SRuns != nil {
		cln.SRuns = make([]*SRun, len(s.SRuns))
		for i, sR := range s.SRuns {
			cln.SRuns[i] = sR.Clone()
		}
	}
	s.RUnlock()
	return
}

// AsExternalSessions returns the session as a list of ExternalSession using all SRuns (thread safe)
func (s *Session) AsExternalSessions(tmz, nodeID string) (aSs []*ExternalSession) {
	s.RLock()
	aSs = make([]*ExternalSession, len(s.SRuns))
	for i, sr := range s.SRuns {
		aSs[i] = &ExternalSession{
			CGRID:         s.CGRID,
			RunID:         sr.Event.GetStringIgnoreErrors(utils.RunID),
			ToR:           sr.Event.GetStringIgnoreErrors(utils.ToR),
			OriginID:      s.EventStart.GetStringIgnoreErrors(utils.OriginID),
			OriginHost:    s.EventStart.GetStringIgnoreErrors(utils.OriginHost),
			Source:        utils.SessionS + "_" + s.EventStart.GetStringIgnoreErrors(utils.EventName),
			RequestType:   sr.Event.GetStringIgnoreErrors(utils.RequestType),
			Tenant:        s.Tenant,
			Category:      sr.Event.GetStringIgnoreErrors(utils.Category),
			Account:       sr.Event.GetStringIgnoreErrors(utils.AccountField),
			Subject:       sr.Event.GetStringIgnoreErrors(utils.Subject),
			Destination:   sr.Event.GetStringIgnoreErrors(utils.Destination),
			SetupTime:     sr.Event.GetTimeIgnoreErrors(utils.SetupTime, tmz),
			AnswerTime:    sr.Event.GetTimeIgnoreErrors(utils.AnswerTime, tmz),
			Usage:         sr.TotalUsage,
			ExtraFields:   sr.Event.AsMapString(utils.MainCDRFields),
			NodeID:        nodeID,
			DebitInterval: s.DebitInterval,
		}
		if sr.NextAutoDebit != nil {
			aSs[i].NextAutoDebit = *sr.NextAutoDebit
		}
		if sr.CD != nil {
			aSs[i].LoopIndex = sr.CD.LoopIndex
			aSs[i].DurationIndex = sr.CD.DurationIndex
			aSs[i].MaxRate = sr.CD.MaxRate
			aSs[i].MaxRateUnit = sr.CD.MaxRateUnit
			aSs[i].MaxCostSoFar = sr.CD.MaxCostSoFar
		}
	}
	s.RUnlock()
	return
}

// AsExternalSession returns the session as an ExternalSession using the SRuns given
func (s *Session) AsExternalSession(sr *SRun, tmz, nodeID string) (aS *ExternalSession) {
	aS = &ExternalSession{
		CGRID:         s.CGRID,
		RunID:         sr.Event.GetStringIgnoreErrors(utils.RunID),
		ToR:           sr.Event.GetStringIgnoreErrors(utils.ToR),
		OriginID:      s.EventStart.GetStringIgnoreErrors(utils.OriginID),
		OriginHost:    s.EventStart.GetStringIgnoreErrors(utils.OriginHost),
		Source:        utils.SessionS + "_" + s.EventStart.GetStringIgnoreErrors(utils.EventName),
		RequestType:   sr.Event.GetStringIgnoreErrors(utils.RequestType),
		Tenant:        s.Tenant,
		Category:      sr.Event.GetStringIgnoreErrors(utils.Category),
		Account:       sr.Event.GetStringIgnoreErrors(utils.AccountField),
		Subject:       sr.Event.GetStringIgnoreErrors(utils.Subject),
		Destination:   sr.Event.GetStringIgnoreErrors(utils.Destination),
		SetupTime:     sr.Event.GetTimeIgnoreErrors(utils.SetupTime, tmz),
		AnswerTime:    sr.Event.GetTimeIgnoreErrors(utils.AnswerTime, tmz),
		Usage:         sr.TotalUsage,
		ExtraFields:   sr.Event.AsMapString(utils.MainCDRFields),
		NodeID:        nodeID,
		DebitInterval: s.DebitInterval,
	}
	if sr.NextAutoDebit != nil {
		aS.NextAutoDebit = *sr.NextAutoDebit
	}
	if sr.CD != nil {
		aS.LoopIndex = sr.CD.LoopIndex
		aS.DurationIndex = sr.CD.DurationIndex
		aS.MaxRate = sr.CD.MaxRate
		aS.MaxRateUnit = sr.CD.MaxRateUnit
		aS.MaxCostSoFar = sr.CD.MaxCostSoFar
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
		cgrEvs[i] = &utils.CGREvent{
			Tenant: s.Tenant,
			ID:     utils.UUIDSha1Prefix(),
			Event:  sr.Event,
			Opts:   s.OptsStart,
		}
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

// SRun is one billing run for the Session
type SRun struct {
	Event     engine.MapEvent        // Event received from ChargerS
	CD        *engine.CallDescriptor // initial CD used for debits, updated on each debit
	EventCost *engine.EventCost

	ExtraDuration time.Duration // keeps the current duration debited on top of what has been asked
	LastUsage     time.Duration // last requested Duration
	TotalUsage    time.Duration // sum of lastUsage
	NextAutoDebit *time.Time
}

// Clone returns the cloned version of SRun
func (sr *SRun) Clone() (clsr *SRun) {
	clsr = &SRun{
		Event:         sr.Event.Clone(),
		ExtraDuration: sr.ExtraDuration,
		LastUsage:     sr.LastUsage,
		TotalUsage:    sr.TotalUsage,
	}
	if sr.CD != nil {
		clsr.CD = sr.CD.Clone()
	}
	if sr.EventCost != nil {
		clsr.EventCost = sr.EventCost.Clone()
	}
	if sr.NextAutoDebit != nil {
		clsr.NextAutoDebit = utils.TimePointer(*sr.NextAutoDebit)
	}
	return
}

// debitReserve attempty to debit from ExtraDuration and returns remaining duration
// if lastUsage is not nil, the ExtraDuration is corrected
func (sr *SRun) debitReserve(dur time.Duration, lastUsage *time.Duration) (rDur time.Duration) {
	if lastUsage != nil &&
		sr.LastUsage != *lastUsage {
		diffUsage := sr.LastUsage - *lastUsage
		sr.ExtraDuration += diffUsage
		sr.TotalUsage -= sr.LastUsage
		sr.TotalUsage += *lastUsage
		sr.LastUsage = *lastUsage
	}
	// debit from reserved
	if sr.ExtraDuration >= dur {
		sr.ExtraDuration -= dur
		sr.LastUsage = dur
		sr.TotalUsage += dur
	} else {
		rDur = dur - sr.ExtraDuration
		sr.ExtraDuration = 0
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
			sr.Event[k] = v
		}
	}
}

// UpdateSRuns updates the SRuns event with the alterable fields (is thread safe)
func (s *Session) UpdateSRuns(updEv engine.MapEvent, alterableFields utils.StringSet) {
	if alterableFields.Size() == 0 { // do not lock if we can't update any field
		return
	}
	s.Lock()
	s.updateSRuns(updEv, alterableFields)
	s.Unlock()
}
