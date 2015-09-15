/*
Real-time Charging System for Telecom & ISP environments
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

package config

import (
	"time"
)

type CdrStatsConfig struct {
	Id               string        // Config id, unique per config instance
	QueueLength      int           // Number of items in the stats buffer
	TimeWindow       time.Duration // Will only keep the CDRs who's call setup time is not older than time.Now()-TimeWindow
	SaveInterval     time.Duration
	Metrics          []string    // ASR, ACD, ACC
	SetupInterval    []time.Time // 2 or less items (>= start interval,< stop_interval)
	TORs             []string
	CdrHosts         []string
	CdrSources       []string
	ReqTypes         []string
	Directions       []string
	Tenants          []string
	Categories       []string
	Accounts         []string
	Subjects         []string
	DestinationIds   []string
	UsageInterval    []time.Duration // 2 or less items (>= Usage, <Usage)
	Suppliers        []string
	DisconnectCauses []string
	MediationRunIds  []string
	RatedAccounts    []string
	RatedSubjects    []string
	CostInterval     []float64 // 2 or less items, (>=Cost, <Cost)
}
