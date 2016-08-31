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

import (
	"reflect"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

func NewCdrStatsFromCdrStatsCfg(csCfg *config.CdrStatsConfig) *CdrStats {
	return &CdrStats{
		Id:              csCfg.Id,
		QueueLength:     csCfg.QueueLength,
		TimeWindow:      csCfg.TimeWindow,
		Metrics:         csCfg.Metrics,
		SetupInterval:   csCfg.SetupInterval,
		TOR:             csCfg.TORs,
		CdrHost:         csCfg.CdrHosts,
		CdrSource:       csCfg.CdrSources,
		ReqType:         csCfg.ReqTypes,
		Direction:       csCfg.Directions,
		Tenant:          csCfg.Tenants,
		Category:        csCfg.Categories,
		Account:         csCfg.Accounts,
		Subject:         csCfg.Subjects,
		DestinationIds:  csCfg.DestinationIds,
		UsageInterval:   csCfg.UsageInterval,
		Supplier:        csCfg.Suppliers,
		DisconnectCause: csCfg.DisconnectCauses,
		MediationRunIds: csCfg.MediationRunIds,
		RatedAccount:    csCfg.RatedAccounts,
		RatedSubject:    csCfg.RatedSubjects,
		CostInterval:    csCfg.CostInterval,
	}
}

type CdrStats struct {
	Id              string        // Config id, unique per config instance
	QueueLength     int           // Number of items in the stats buffer
	TimeWindow      time.Duration // Will only keep the CDRs who's call setup time is not older than time.Now()-TimeWindow
	SaveInterval    time.Duration
	Metrics         []string        // ASR, ACD, ACC
	SetupInterval   []time.Time     // CDRFieldFilter on SetupInterval, 2 or less items (>= start interval,< stop_interval)
	TOR             []string        // CDRFieldFilter on TORs
	CdrHost         []string        // CDRFieldFilter on CdrHosts
	CdrSource       []string        // CDRFieldFilter on CdrSources
	ReqType         []string        // CDRFieldFilter on RequestTypes
	Direction       []string        // CDRFieldFilter on Directions
	Tenant          []string        // CDRFieldFilter on Tenants
	Category        []string        // CDRFieldFilter on Categories
	Account         []string        // CDRFieldFilter on Accounts
	Subject         []string        // CDRFieldFilter on Subjects
	DestinationIds  []string        // CDRFieldFilter on DestinationPrefixes
	UsageInterval   []time.Duration // CDRFieldFilter on UsageInterval, 2 or less items (>= Usage, <Usage)
	PddInterval     []time.Duration // CDRFieldFilter on PddInterval, 2 or less items (>= Pdd, <Pdd)
	Supplier        []string        // CDRFieldFilter on Suppliers
	DisconnectCause []string        // Filter on DisconnectCause
	MediationRunIds []string        // CDRFieldFilter on MediationRunIds
	RatedAccount    []string        // CDRFieldFilter on RatedAccounts
	RatedSubject    []string        // CDRFieldFilter on RatedSubjects
	CostInterval    []float64       // CDRFieldFilter on CostInterval, 2 or less items, (>=Cost, <Cost)
	Triggers        ActionTriggers
}

func (cs *CdrStats) AcceptCdr(cdr *CDR) bool {
	if cdr == nil {
		return false
	}
	if len(cs.SetupInterval) > 0 {
		if cdr.SetupTime.Before(cs.SetupInterval[0]) {
			return false
		}
		if len(cs.SetupInterval) > 1 && (cdr.SetupTime.Equal(cs.SetupInterval[1]) || cdr.SetupTime.After(cs.SetupInterval[1])) {
			return false
		}
	}
	if len(cs.TOR) > 0 && !utils.IsSliceMember(cs.TOR, cdr.ToR) {
		return false
	}
	if len(cs.CdrHost) > 0 && !utils.IsSliceMember(cs.CdrHost, cdr.OriginHost) {
		return false
	}
	if len(cs.CdrSource) > 0 && !utils.IsSliceMember(cs.CdrSource, cdr.Source) {
		return false
	}
	if len(cs.ReqType) > 0 && !utils.IsSliceMember(cs.ReqType, cdr.RequestType) {
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
	if len(cs.DestinationIds) > 0 {
		found := false
		for _, p := range utils.SplitPrefix(cdr.Destination, MIN_PREFIX_MATCH) {
			if destIDs, err := ratingStorage.GetReverseDestination(p, false, utils.NonTransactional); err == nil {
				for _, idID := range destIDs {
					if utils.IsSliceMember(cs.DestinationIds, idID) {
						found = true
						break
					}
				}
				if found {
					break
				}
			}
			if found {
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
	if len(cs.PddInterval) > 0 {
		if cdr.PDD < cs.PddInterval[0] {
			return false
		}
		if len(cs.PddInterval) > 1 && cdr.PDD >= cs.PddInterval[1] {
			return false
		}
	}
	if len(cs.Supplier) > 0 && !utils.IsSliceMember(cs.Supplier, cdr.Supplier) {
		return false
	}
	if len(cs.DisconnectCause) > 0 && !utils.IsSliceMember(cs.DisconnectCause, cdr.DisconnectCause) {
		return false
	}
	if len(cs.MediationRunIds) > 0 && !utils.IsSliceMember(cs.MediationRunIds, cdr.RunID) {
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
	return true
}

func (cs *CdrStats) hasGeneralConfigs() bool {
	return cs.QueueLength == 0 &&
		cs.TimeWindow == 0 &&
		cs.SaveInterval == 0 &&
		len(cs.Metrics) == 0
}

func (cs *CdrStats) equalExceptTriggers(other *CdrStats) bool {
	return cs.QueueLength == other.QueueLength &&
		cs.TimeWindow == other.TimeWindow &&
		cs.SaveInterval == other.SaveInterval &&
		reflect.DeepEqual(cs.Metrics, other.Metrics) &&
		reflect.DeepEqual(cs.SetupInterval, other.SetupInterval) &&
		reflect.DeepEqual(cs.TOR, other.TOR) &&
		reflect.DeepEqual(cs.CdrHost, other.CdrHost) &&
		reflect.DeepEqual(cs.CdrSource, other.CdrSource) &&
		reflect.DeepEqual(cs.ReqType, other.ReqType) &&
		reflect.DeepEqual(cs.Direction, other.Direction) &&
		reflect.DeepEqual(cs.Tenant, other.Tenant) &&
		reflect.DeepEqual(cs.Category, other.Category) &&
		reflect.DeepEqual(cs.Account, other.Account) &&
		reflect.DeepEqual(cs.Subject, other.Subject) &&
		reflect.DeepEqual(cs.DestinationIds, other.DestinationIds) &&
		reflect.DeepEqual(cs.UsageInterval, other.UsageInterval) &&
		reflect.DeepEqual(cs.PddInterval, other.PddInterval) &&
		reflect.DeepEqual(cs.Supplier, other.Supplier) &&
		reflect.DeepEqual(cs.DisconnectCause, other.DisconnectCause) &&
		reflect.DeepEqual(cs.MediationRunIds, other.MediationRunIds) &&
		reflect.DeepEqual(cs.RatedAccount, other.RatedAccount) &&
		reflect.DeepEqual(cs.RatedSubject, other.RatedSubject) &&
		reflect.DeepEqual(cs.CostInterval, other.CostInterval)
}
