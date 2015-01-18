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
	"github.com/cgrates/cgrates/utils"
	"time"
)

type CdrStatsConfig struct {
	Id                  string        // Config id, unique per config instance
	QueueLength         int           // Number of items in the stats buffer
	TimeWindow          time.Duration // Will only keep the CDRs who's call setup time is not older than time.Now()-TimeWindow
	Metrics             []string      // ASR, ACD, ACC
	SetupInterval       []time.Time   // 2 or less items (>= start interval,< stop_interval)
	TORs                []string
	CdrHosts            []string
	CdrSources          []string
	ReqTypes            []string
	Directions          []string
	Tenants             []string
	Categories          []string
	Accounts            []string
	Subjects            []string
	DestinationPrefixes []string
	UsageInterval       []time.Duration // 2 or less items (>= Usage, <Usage)
	MediationRunIds     []string
	RatedAccounts       []string
	RatedSubjects       []string
	CostInterval        []float64 // 2 or less items, (>=Cost, <Cost)
}

func (self *CdrStatsConfig) loadFromJsonCfg(jsnCfg *CdrStatsJsonCfg) error {
	var err error
	if jsnCfg.Queue_length != nil {
		self.QueueLength = *jsnCfg.Queue_length
	}
	if jsnCfg.Time_window != nil {
		if self.TimeWindow, err = utils.ParseDurationWithSecs(*jsnCfg.Time_window); err != nil {
			return err
		}
	}
	if jsnCfg.Metrics != nil {
		self.Metrics = *jsnCfg.Metrics
	}
	if jsnCfg.Setup_interval != nil {
		for _, setupTimeStr := range *jsnCfg.Setup_interval {
			if setupTime, err := utils.ParseTimeDetectLayout(setupTimeStr); err != nil {
				return err
			} else {
				self.SetupInterval = append(self.SetupInterval, setupTime)
			}
		}
	}
	if jsnCfg.Tors != nil {
		self.TORs = *jsnCfg.Tors
	}
	if jsnCfg.Cdr_hosts != nil {
		self.CdrHosts = *jsnCfg.Cdr_hosts
	}
	if jsnCfg.Cdr_sources != nil {
		self.CdrSources = *jsnCfg.Cdr_sources
	}
	if jsnCfg.Req_types != nil {
		self.ReqTypes = *jsnCfg.Req_types
	}
	if jsnCfg.Directions != nil {
		self.Directions = *jsnCfg.Directions
	}
	if jsnCfg.Tenants != nil {
		self.Tenants = *jsnCfg.Tenants
	}
	if jsnCfg.Categories != nil {
		self.Categories = *jsnCfg.Categories
	}
	if jsnCfg.Accounts != nil {
		self.Accounts = *jsnCfg.Accounts
	}
	if jsnCfg.Subjects != nil {
		self.Subjects = *jsnCfg.Subjects
	}
	if jsnCfg.Destination_prefixes != nil {
		self.DestinationPrefixes = *jsnCfg.Destination_prefixes
	}
	if jsnCfg.Usage_interval != nil {
		for _, usageDurStr := range *jsnCfg.Usage_interval {
			if usageDur, err := utils.ParseDurationWithSecs(usageDurStr); err != nil {
				return err
			} else {
				self.UsageInterval = append(self.UsageInterval, usageDur)
			}
		}
	}
	if jsnCfg.Mediation_run_ids != nil {
		self.MediationRunIds = *jsnCfg.Mediation_run_ids
	}
	if jsnCfg.Rated_accounts != nil {
		self.RatedAccounts = *jsnCfg.Rated_accounts
	}
	if jsnCfg.Rated_subjects != nil {
		self.RatedSubjects = *jsnCfg.Rated_subjects
	}
	if jsnCfg.Cost_interval != nil {
		self.CostInterval = *jsnCfg.Cost_interval
	}
	return nil
}
