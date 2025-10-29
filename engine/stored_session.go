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
	ResourceID    string
	ClientConnID  string        // connection ID towards the client so we can recover from passive
	EventStart    MapEvent      // Event which started the session
	DebitInterval time.Duration // execute debits for *prepaid runs
	Chargeable    bool          // used in case of pausing debit
	SRuns         []*StoredSRun // forked based on ChargerS
	OptsStart     MapEvent
	UpdatedAt     time.Time // time when session was changed
}
