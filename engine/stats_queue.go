/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2013 ITsysCOM

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
	"sync"
	"time"

	"github.com/cgrates/cgrates/utils"
)

type StatsQueue struct {
	cdrs    []*QCDR
	conf    *CdrStats
	metrics map[string]Metric
	mux     sync.RWMutex
}

// Simplified cdr structure containing only the necessary info
type QCDR struct {
	SetupTime  time.Time
	AnswerTime time.Time
	Usage      time.Duration
	Cost       float64
}

func NewStatsQueue(conf *CdrStats) *StatsQueue {
	if conf == nil {
		return &StatsQueue{metrics: make(map[string]Metric)}
	}
	sq := &StatsQueue{
		conf:    conf,
		metrics: make(map[string]Metric, len(conf.Metrics)),
	}
	for _, m := range conf.Metrics {
		metric := CreateMetric(m)
		if metric != nil {
			sq.metrics[m] = metric
		}
	}
	return sq
}

func (sq *StatsQueue) AppendCDR(cdr *utils.StoredCdr) {
	sq.mux.Lock()
	defer sq.mux.Unlock()
	if sq.acceptCDR(cdr) {
		qcdr := sq.simplifyCDR(cdr)
		sq.cdrs = append(sq.cdrs, qcdr)
		sq.addToMetrics(qcdr)
		sq.purgeObsoleteCDRs()
		// check for trigger
		stats := sq.getStats()
		sq.conf.Triggers.Sort()
		for _, at := range sq.conf.Triggers {
			if at.MinQueuedItems > 0 && len(sq.cdrs) < at.MinQueuedItems {
				continue
			}
			if strings.HasPrefix(at.ThresholdType, "*min_") {
				if value, ok := stats[at.ThresholdType[len("*min_"):]]; ok {
					if value <= at.ThresholdValue {
						at.Execute(nil, sq)
					}
				}
			}
			if strings.HasPrefix(at.ThresholdType, "*max_") {
				if value, ok := stats[at.ThresholdType[len("*max_"):]]; ok {
					if value >= at.ThresholdValue {
						at.Execute(nil, sq)
					}
				}
			}
		}
	}
}

func (sq *StatsQueue) addToMetrics(cdr *QCDR) {
	for _, metric := range sq.metrics {
		metric.AddCDR(cdr)
	}
}

func (sq *StatsQueue) removeFromMetrics(cdr *QCDR) {
	for _, metric := range sq.metrics {
		metric.RemoveCDR(cdr)
	}
}

func (sq *StatsQueue) simplifyCDR(cdr *utils.StoredCdr) *QCDR {
	return &QCDR{
		SetupTime:  cdr.SetupTime,
		AnswerTime: cdr.AnswerTime,
		Usage:      cdr.Usage,
		Cost:       cdr.Cost,
	}
}

func (sq *StatsQueue) purgeObsoleteCDRs() {
	if sq.conf.QueueLength > 0 {
		currentLength := len(sq.cdrs)
		if currentLength > sq.conf.QueueLength {
			for _, cdr := range sq.cdrs[:currentLength-sq.conf.QueueLength] {
				sq.removeFromMetrics(cdr)
			}
			sq.cdrs = sq.cdrs[currentLength-sq.conf.QueueLength:]
		}
	}
	if sq.conf.TimeWindow > 0 {
		for i, cdr := range sq.cdrs {
			if time.Now().Sub(cdr.SetupTime) > sq.conf.TimeWindow {
				sq.removeFromMetrics(cdr)
				continue
			} else {
				if i > 0 {
					sq.cdrs = sq.cdrs[i:]
				}
				break
			}
		}
	}
}

func (sq *StatsQueue) GetStats() map[string]float64 {
	sq.mux.RLock()
	defer sq.mux.RUnlock()
	return sq.getStats()
}

func (sq *StatsQueue) getStats() map[string]float64 {
	stat := make(map[string]float64, len(sq.metrics))
	for key, metric := range sq.metrics {
		stat[key] = metric.GetValue()
	}
	return stat
}

func (sq *StatsQueue) acceptCDR(cdr *utils.StoredCdr) bool {
	if len(sq.conf.SetupInterval) > 0 {
		if cdr.SetupTime.Before(sq.conf.SetupInterval[0]) {
			return false
		}
		if len(sq.conf.SetupInterval) > 1 && (cdr.SetupTime.Equal(sq.conf.SetupInterval[1]) || cdr.SetupTime.After(sq.conf.SetupInterval[1])) {
			return false
		}
	}
	if len(sq.conf.TOR) > 0 && !utils.IsSliceMember(sq.conf.TOR, cdr.TOR) {
		return false
	}
	if len(sq.conf.CdrHost) > 0 && !utils.IsSliceMember(sq.conf.CdrHost, cdr.CdrHost) {
		return false
	}
	if len(sq.conf.CdrSource) > 0 && !utils.IsSliceMember(sq.conf.CdrSource, cdr.CdrSource) {
		return false
	}
	if len(sq.conf.ReqType) > 0 && !utils.IsSliceMember(sq.conf.ReqType, cdr.ReqType) {
		return false
	}
	if len(sq.conf.Direction) > 0 && !utils.IsSliceMember(sq.conf.Direction, cdr.Direction) {
		return false
	}
	if len(sq.conf.Tenant) > 0 && !utils.IsSliceMember(sq.conf.Tenant, cdr.Tenant) {
		return false
	}
	if len(sq.conf.Category) > 0 && !utils.IsSliceMember(sq.conf.Category, cdr.Category) {
		return false
	}
	if len(sq.conf.Account) > 0 && !utils.IsSliceMember(sq.conf.Account, cdr.Account) {
		return false
	}
	if len(sq.conf.Subject) > 0 && !utils.IsSliceMember(sq.conf.Subject, cdr.Subject) {
		return false
	}
	if len(sq.conf.DestinationPrefix) > 0 {
		found := false
		for _, prefix := range sq.conf.DestinationPrefix {
			if strings.HasPrefix(cdr.Destination, prefix) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	if len(sq.conf.UsageInterval) > 0 {
		if cdr.Usage < sq.conf.UsageInterval[0] {
			return false
		}
		if len(sq.conf.UsageInterval) > 1 && cdr.Usage >= sq.conf.UsageInterval[1] {
			return false
		}
	}
	if len(sq.conf.MediationRunIds) > 0 && !utils.IsSliceMember(sq.conf.MediationRunIds, cdr.MediationRunId) {
		return false
	}
	if len(sq.conf.CostInterval) > 0 {
		if cdr.Cost < sq.conf.CostInterval[0] {
			return false
		}
		if len(sq.conf.CostInterval) > 1 && cdr.Cost >= sq.conf.CostInterval[1] {
			return false
		}
	}
	return true
}
