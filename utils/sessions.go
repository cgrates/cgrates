// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package utils

import "time"

// Session holds the data for a call and its billing runs.
type Session struct {
	ID             string    // CGRID
	OriginCGREvent *CGREvent // initial event for the session
	ClientConnID   string    // connection towards the client for passive sessions
	SRuns          map[string]*SRun
}

// Clone copies the session and its runs. Nested event values are not copied.
func (s *Session) Clone() *Session {
	cln := &Session{
		ClientConnID: s.ClientConnID,
	}
	if s.OriginCGREvent != nil {
		cln.OriginCGREvent = s.OriginCGREvent.Clone()
	}
	if s.SRuns != nil {
		cln.SRuns = make(map[string]*SRun, len(s.SRuns))
		for id, run := range s.SRuns {
			cln.SRuns[id] = run.Clone()
		}
	}
	return cln
}

// SRun holds the data for one billing run of a session.
type SRun struct {
	ID                 string
	CGREvent           *CGREvent
	InterimUsage       *Decimal // last requested usage
	UsageAdjustment    *Decimal // unused grant or additional usage to debit
	TotalUsage         *Decimal // sum of interim usage
	Charges            *EventCharges
	AutoChargeInterval *time.Duration
	NextAutoCharge     *time.Time
}

// Clone copies the run. Nested event values are not copied.
func (sr *SRun) Clone() *SRun {
	if sr == nil {
		return nil
	}
	cln := &SRun{
		ID:              sr.ID,
		CGREvent:        sr.CGREvent.Clone(),
		InterimUsage:    sr.InterimUsage.Clone(),
		UsageAdjustment: sr.UsageAdjustment.Clone(),
		TotalUsage:      sr.TotalUsage.Clone(),
		Charges:         sr.Charges.Clone(),
	}
	if sr.AutoChargeInterval != nil {
		cln.AutoChargeInterval = new(*sr.AutoChargeInterval)
	}
	if sr.NextAutoCharge != nil {
		cln.NextAutoCharge = new(*sr.NextAutoCharge)
	}
	return cln
}
