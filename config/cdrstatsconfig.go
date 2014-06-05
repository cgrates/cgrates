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
	QueuedItems       int64       // Number of items in the stats buffer
	SetupInterval     []time.Time // 2 or less items (>= start interval,< stop_interval)
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
}
