// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package engine

import (
	"time"
)

// used to evade import cycle of the real sessions.SRun struct
type StoredSRun struct {
	Event     MapEvent        // Event received from ChargerS
	CD        *CallDescriptor // initial CD used for debits, updated on each debit
	EventCost *EventCost

	ExtraDuration time.Duration // keeps the current duration debited on top of what has been asked
	LastUsage     time.Duration // last requested Duration
	TotalUsage    time.Duration // sum of lastUsage
	NextAutoDebit *time.Time
}

// Holds a Session for storing in DataDB
type StoredSession struct {
	CGRID         string
	Tenant        string
	IPAllocID     string
	ResourceID    string
	ClientConnID  string        // connection ID towards the client so we can recover from passive
	EventStart    MapEvent      // Event which started the session
	DebitInterval time.Duration // execute debits for *prepaid runs
	Chargeable    bool          // used in case of pausing debit
	SRuns         []*StoredSRun // forked based on ChargerS
	OptsStart     MapEvent
	UpdatedAt     time.Time // time when session was changed
}
