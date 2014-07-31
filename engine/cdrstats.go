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

package engine

import (
	"strings"
	"time"

	"github.com/cgrates/cgrates/utils"
)

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

func (cs *CdrStats) AcceptCDR(cdr *utils.StoredCdr) bool {
	if len(cs.SetupInterval) > 0 {
		if cdr.SetupTime.Before(cs.SetupInterval[0]) {
			return false
		}
		if len(cs.SetupInterval) > 1 && (cdr.SetupTime.Equal(cs.SetupInterval[1]) || cdr.SetupTime.After(cs.SetupInterval[1])) {
			return false
		}
	}
	if len(cs.TOR) > 0 && !utils.IsSliceMember(cs.TOR, cdr.TOR) {
		return false
	}
	if len(cs.CdrHost) > 0 && !utils.IsSliceMember(cs.CdrHost, cdr.CdrHost) {
		return false
	}
	if len(cs.CdrSource) > 0 && !utils.IsSliceMember(cs.CdrSource, cdr.CdrSource) {
		return false
	}
	if len(cs.ReqType) > 0 && !utils.IsSliceMember(cs.ReqType, cdr.ReqType) {
		return false
	}
	if len(cs.Direction) > 0 && !utils.IsSliceMember(cs.Direction, cdr.Direction) {
		return false
	}
	if len(cs.Tenant) > 0 && !utils.IsSliceMember(cs.Tenant, cdr.Tenant) {
		return false
	}
	if len(cs.Category) > 0 && !utils.IsSliceMember(cs.Category, cdr.Category) {
		return false
	}
	if len(cs.Account) > 0 && !utils.IsSliceMember(cs.Account, cdr.Account) {
		return false
	}
	if len(cs.Subject) > 0 && !utils.IsSliceMember(cs.Subject, cdr.Subject) {
		return false
	}
	if len(cs.DestinationPrefix) > 0 {
		found := false
		for _, prefix := range cs.DestinationPrefix {
			if strings.HasPrefix(cdr.Destination, prefix) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	if len(cs.UsageInterval) > 0 {
		if cdr.Usage < cs.UsageInterval[0] {
			return false
		}
		if len(cs.UsageInterval) > 1 && cdr.Usage >= cs.UsageInterval[1] {
			return false
		}
	}
	if len(cs.MediationRunIds) > 0 && !utils.IsSliceMember(cs.MediationRunIds, cdr.MediationRunId) {
		return false
	}
	if len(cs.CostInterval) > 0 {
		if cdr.Cost < cs.CostInterval[0] {
			return false
		}
		if len(cs.CostInterval) > 1 && cdr.Cost >= cs.CostInterval[1] {
			return false
		}
	}
	if len(cs.RatedAccount) > 0 && !utils.IsSliceMember(cs.RatedAccount, cdr.RatedAccount) {
		return false
	}
	if len(cs.RatedSubject) > 0 && !utils.IsSliceMember(cs.RatedSubject, cdr.RatedSubject) {
		return false
	}
	return true
}
