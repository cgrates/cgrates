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
	"sync"
	"time"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

type SessionID struct {
	OriginHost string
	OriginID   string
}

func (s *SessionID) CGRID() string {
	return utils.Sha1(s.OriginID, s.OriginHost)
}

// Will be used when displaying active sessions via RPC
type ActiveSession struct {
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
}

type Session struct {
	sync.RWMutex

	CGRID         string
	Tenant        string
	ResourceID    string
	ClientConnID  string           // connection ID towards the client so we can recover from passive
	EventStart    *engine.SafEvent // Event which started the session
	DebitInterval time.Duration    // execute debits for *prepaid runs
	SRuns         []*SRun          // forked based on ChargerS

	debitStop   chan struct{}
	sTerminator *sTerminator // automatic timeout for the session
	*utils.ArgDispatcher
}

// CGRid is a thread-safe method to return the CGRID of a session
func (s *Session) CGRid() (cgrID string) {
	s.RLock()
	cgrID = s.CGRID
	s.RUnlock()
	return
}

// DebitStopChan reads the debit stop
func (s *Session) DebitStopChan() (dbtStop chan struct{}) {
	s.RLock()
	dbtStop = s.debitStop
	s.RUnlock()
	return
}

// Clone is a thread safe method to clone the sessions information
func (s Session) Clone() (cln *Session) {
	s.RLock()
	cln = &Session{
		CGRID:        s.CGRID,
		Tenant:       s.Tenant,
		ResourceID:   s.ResourceID,
		EventStart:   s.EventStart.Clone(),
		ClientConnID: s.ClientConnID,
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

func (s *Session) AsActiveSessions(tmz, nodeID string) (aSs []*ActiveSession) {
	s.RLock()
	aSs = make([]*ActiveSession, len(s.SRuns))
	for i, sr := range s.SRuns {
		aSs[i] = &ActiveSession{
			CGRID:       s.CGRID,
			RunID:       sr.Event.GetStringIgnoreErrors(utils.RunID),
			ToR:         sr.Event.GetStringIgnoreErrors(utils.ToR),
			OriginID:    s.EventStart.GetStringIgnoreErrors(utils.OriginID),
			OriginHost:  s.EventStart.GetStringIgnoreErrors(utils.OriginHost),
			Source:      utils.SessionS + "_" + s.EventStart.GetStringIgnoreErrors(utils.EVENT_NAME),
			RequestType: sr.Event.GetStringIgnoreErrors(utils.RequestType),
			Tenant:      s.Tenant,
			Category:    sr.Event.GetStringIgnoreErrors(utils.Category),
			Account:     sr.Event.GetStringIgnoreErrors(utils.Account),
			Subject:     sr.Event.GetStringIgnoreErrors(utils.Subject),
			Destination: sr.Event.GetStringIgnoreErrors(utils.Destination),
			SetupTime:   sr.Event.GetTimeIgnoreErrors(utils.SetupTime, tmz),
			AnswerTime:  sr.Event.GetTimeIgnoreErrors(utils.AnswerTime, tmz),
			Usage:       sr.TotalUsage,
			ExtraFields: sr.Event.AsMapStringIgnoreErrors(
				utils.NewStringMap(utils.MainCDRFields...)),
			NodeID:        nodeID,
			DebitInterval: s.DebitInterval,
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
func (s *Session) asActiveSessions(sr *SRun, tmz, nodeID string) (aS *ActiveSession) {
	s.RLock()
	aS = &ActiveSession{
		CGRID:       s.CGRID,
		RunID:       sr.Event.GetStringIgnoreErrors(utils.RunID),
		ToR:         sr.Event.GetStringIgnoreErrors(utils.ToR),
		OriginID:    s.EventStart.GetStringIgnoreErrors(utils.OriginID),
		OriginHost:  s.EventStart.GetStringIgnoreErrors(utils.OriginHost),
		Source:      utils.SessionS + "_" + s.EventStart.GetStringIgnoreErrors(utils.EVENT_NAME),
		RequestType: sr.Event.GetStringIgnoreErrors(utils.RequestType),
		Tenant:      s.Tenant,
		Category:    sr.Event.GetStringIgnoreErrors(utils.Category),
		Account:     sr.Event.GetStringIgnoreErrors(utils.Account),
		Subject:     sr.Event.GetStringIgnoreErrors(utils.Subject),
		Destination: sr.Event.GetStringIgnoreErrors(utils.Destination),
		SetupTime:   sr.Event.GetTimeIgnoreErrors(utils.SetupTime, tmz),
		AnswerTime:  sr.Event.GetTimeIgnoreErrors(utils.AnswerTime, tmz),
		Usage:       sr.TotalUsage,
		ExtraFields: sr.Event.AsMapStringIgnoreErrors(
			utils.NewStringMap(utils.MainCDRFields...)),
		NodeID:        nodeID,
		DebitInterval: s.DebitInterval,
	}
	if sr.CD != nil {
		aS.LoopIndex = sr.CD.LoopIndex
		aS.DurationIndex = sr.CD.DurationIndex
		aS.MaxRate = sr.CD.MaxRate
		aS.MaxRateUnit = sr.CD.MaxRateUnit
		aS.MaxCostSoFar = sr.CD.MaxCostSoFar
	}
	s.RUnlock()
	return
}

// TotalUsage returns the first session run total usage
func (s *Session) TotalUsage() (tDur time.Duration) {
	if len(s.SRuns) == 0 {
		return
	}
	s.RLock()
	for _, sr := range s.SRuns {
		tDur = sr.TotalUsage
		break // only first
	}
	s.RUnlock()
	return
}

// AsCGREvents is a  method to return the Session as CGREvents
// there will be one CGREvent for each SRun plus one *raw for EventStart
// AsCGREvents is not thread safe since it is supposed to run by the time Session is closed
func (s *Session) asCGREvents() (cgrEvs []*utils.CGREvent, err error) {
	cgrEvs = make([]*utils.CGREvent, len(s.SRuns)+1) // so we can gather all cdr info while under lock
	rawEv := s.EventStart.MapEvent()
	rawEv[utils.RunID] = utils.MetaRaw
	cgrEvs[0] = &utils.CGREvent{
		Tenant: s.Tenant,
		ID:     utils.UUIDSha1Prefix(),
		Event:  rawEv,
	}
	for i, sr := range s.SRuns {
		cgrEvs[i+1] = &utils.CGREvent{
			Tenant: s.Tenant,
			ID:     utils.UUIDSha1Prefix(),
			Event:  sr.Event,
		}
	}
	return
}

// SRun is one billing run for the Session
type SRun struct {
	Event     engine.MapEvent        // Event received from ChargerS
	CD        *engine.CallDescriptor // initial CD used for debits, updated on each debit
	EventCost *engine.EventCost

	ExtraDuration time.Duration // keeps the current duration debited on top of what has been asked
	LastUsage     time.Duration // last requested Duration
	TotalUsage    time.Duration // sum of lastUsage
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
