/*
Real-time Charging System for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can Storagetribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITH*out ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>
*/

package engine

import "time"

type CdrStats struct {
	Id                string        // Config id, unique per config instance
	QueueLength       int           // Number of items in the stats buffer
	TimeWindow        time.Duration // Will only keep the CDRs who's call setup time is not older than time.Now()-TimeWindow
	Metrics           []string      // ASR, ACD, ACC
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
	RatedAccount      []string
	RatedSubject      []string
	CostInterval      []float64 // 2 or less items, (>=Cost, <Cost)
	Triggers          ActionTriggerPriotityList
}
