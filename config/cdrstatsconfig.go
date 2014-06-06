/*
Real-time Charging System for Telecom & ISP environments
Copyright (C) 2012-2014 ITsysCOM GmbH

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

package config

import (
	"time"
)

type CdrStatsConfig struct {
	Id                string        // Config id, unique per config instance
	RatedCdrs         bool          // Build the stats for rated cdrs instead of raw ones
	QueuedItems       int64         // Number of items in the stats buffer
	TimeWindow        time.Duration // Will only keep the CDRs who's call setup time is not older than time.Now()-TimeWindow
	ProcessedStats    []string      // ASR, ACD, ACC
	SetupInterval     []time.Time   // 2 or less items (>= start interval,< stop_interval)
	TOR               []string
	CdrHost           []string
	CdrSource         []string
	ReqType           []string
	Direction         []string
	Tenant            []string
	Category          []string
	Account           []string
	Subject           []string
	DestinationPrefix []string
	UsageInterval     []time.Duration // 2 or less items (>= Usage, <Usage)
	MediationRunIds   []string
	CostInterval      []float64 // 2 or less items, (>=Cost, <Cost)
	CdrStatsTriggers  []*CdrStatsTrigger
}

type CdrStatsTrigger struct {
	ThresholdType  string // *min_asr, *max_asr, *min_acd, *max_acd, *min_acc, *max_acc
	ThresholdValue float64
	MinQueuedItems int           // Trigger actions only if this number is hit
	MinSleep       time.Duration // Minimum duration between two executions in case of recurrent triggers
	ActionsId      string        // Id of actions to be executed
	Recurrent      bool          // Re-enable automatically once executed
	Weight         float64
}
