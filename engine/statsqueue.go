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
package engine

import (
	"time"

	"github.com/cgrates/cgrates/utils"
)

// SQItem represents one item in the stats queue
type SQItem struct {
	EventID    string     // Bounded to the original StatsEvent
	ExpiryTime *time.Time // Used to auto-expire events
}

// SQStoredMetrics contains metrics saved in DB
type SQStoredMetrics struct {
	SqID      string                // StatsInstanceID
	SEvents   map[string]StatsEvent // Events used by SQItems
	SQItems   []*SQItem             // SQItems
	SQMetrics map[string][]byte
}

// StatsEvent is an event received by StatService
type StatsEvent map[string]interface{}

func (se StatsEvent) ID() (id string) {
	if sID, has := se[utils.ID]; has {
		id = sID.(string)
	}
	return
}

// StatsQueue represents the configuration of a  StatsInstance in StatS
type StatsQueue struct {
	ID                 string // QueueID
	Filters            []*RequestFilter
	ActivationInterval *utils.ActivationInterval // Activation interval
	QueueLength        int
	TTL                *time.Duration
	Metrics            []string // list of metrics to build
	Store              bool     // store to DB
	Thresholds         []string // list of thresholds to be checked after changes
	Weight             float64
}
