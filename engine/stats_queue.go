/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2012-2015 ITsysCOM

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
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/cgrates/cgrates/utils"
)

type StatsQueue struct {
	Cdrs    []*QCdr
	conf    *CdrStats
	metrics map[string]Metric
	mux     sync.Mutex
	dirty   bool
}

var METRIC_TRIGGER_MAP = map[string]string{
	"*min_asr": ASR,
	"*max_asr": ASR,
	"*min_pdd": PDD,
	"*max_pdd": PDD,
	"*min_acd": ACD,
	"*max_acd": ACD,
	"*min_tcd": TCD,
	"*max_tcd": TCD,
	"*min_acc": ACC,
	"*max_acc": ACC,
	"*min_tcc": TCC,
	"*max_tcc": TCC,
	"*min_ddc": DDC,
	"*max_ddc": DDC,
}

// Simplified cdr structure containing only the necessary info
type QCdr struct {
	SetupTime  time.Time
	AnswerTime time.Time
	Pdd        time.Duration
	Usage      time.Duration
	Cost       float64
	Dest       string
}

func NewStatsQueue(conf *CdrStats) *StatsQueue {
	if conf == nil {
		return &StatsQueue{metrics: make(map[string]Metric)}
	}
	sq := &StatsQueue{}
	sq.UpdateConf(conf)
	return sq
}

func (sq *StatsQueue) UpdateConf(conf *CdrStats) {
	sq.mux.Lock()
	defer sq.mux.Unlock()
	// check if new conf asks for action trigger reset only
	if sq.conf != nil && (!conf.hasGeneralConfigs() || sq.conf.equalExceptTriggers(conf)) {
		sq.conf.Triggers = conf.Triggers
		return
	}
	sq.conf = conf
	sq.Cdrs = make([]*QCdr, 0)
	sq.metrics = make(map[string]Metric, len(conf.Metrics))
	sq.dirty = true
	for _, m := range conf.Metrics {
		if metric := CreateMetric(m); metric != nil {
			sq.metrics[m] = metric
		}
	}
}

func (sq *StatsQueue) Save(adb AccountingStorage) {
	sq.mux.Lock()
	defer sq.mux.Unlock()
	if sq.dirty {
		if err := adb.SetCdrStatsQueue(sq); err != nil {
			utils.Logger.Err(fmt.Sprintf("Error saving cdr stats queue id %s: %v", sq.GetId(), err))
			return
		}
		sq.dirty = false
	}
}

func (sq *StatsQueue) Load(saved *StatsQueue) {
	sq.mux.Lock()
	defer sq.mux.Unlock()
	sq.Cdrs = saved.Cdrs
	for _, qcdr := range saved.Cdrs {
		sq.appendQcdr(qcdr, false)
	}
}

func (sq *StatsQueue) AppendCDR(cdr *CDR) {
	sq.mux.Lock()
	defer sq.mux.Unlock()
	if sq.conf.AcceptCdr(cdr) {
		sq.appendQcdr(sq.simplifyCdr(cdr), true)
	}
}

func (sq *StatsQueue) appendQcdr(qcdr *QCdr, runTrigger bool) {
	sq.Cdrs = append(sq.Cdrs, qcdr)
	sq.addToMetrics(qcdr)
	sq.purgeObsoleteCdrs()
	sq.dirty = true
	// check for trigger
	if runTrigger {
		stats := sq.getStats()
		sq.conf.Triggers.Sort()
		for _, at := range sq.conf.Triggers {
			if at.MinQueuedItems > 0 && len(sq.Cdrs) < at.MinQueuedItems {
				continue
			}
			if strings.HasPrefix(at.ThresholdType, "*min_") {
				if value, ok := stats[METRIC_TRIGGER_MAP[at.ThresholdType]]; ok {
					if value > STATS_NA && value <= at.ThresholdValue {
						at.Execute(nil, sq.Triggered(at))
					}
				}
			}
			if strings.HasPrefix(at.ThresholdType, "*max_") {
				if value, ok := stats[METRIC_TRIGGER_MAP[at.ThresholdType]]; ok {
					if value > STATS_NA && value >= at.ThresholdValue {
						at.Execute(nil, sq.Triggered(at))
					}
				}
			}
		}
	}
}

func (sq *StatsQueue) addToMetrics(cdr *QCdr) {
	for _, metric := range sq.metrics {
		metric.AddCdr(cdr)
	}
}

func (sq *StatsQueue) removeFromMetrics(cdr *QCdr) {
	for _, metric := range sq.metrics {
		metric.RemoveCdr(cdr)
	}
}

func (sq *StatsQueue) simplifyCdr(cdr *CDR) *QCdr {
	return &QCdr{
		SetupTime:  cdr.SetupTime,
		AnswerTime: cdr.AnswerTime,
		Pdd:        cdr.PDD,
		Usage:      cdr.Usage,
		Cost:       cdr.Cost,
		Dest:       cdr.Destination,
	}
}

func (sq *StatsQueue) purgeObsoleteCdrs() {
	if sq.conf.QueueLength > 0 {
		currentLength := len(sq.Cdrs)
		if currentLength > sq.conf.QueueLength {
			for _, cdr := range sq.Cdrs[:currentLength-sq.conf.QueueLength] {
				sq.removeFromMetrics(cdr)
			}
			sq.Cdrs = sq.Cdrs[currentLength-sq.conf.QueueLength:]
		}
	}
	if sq.conf.TimeWindow > 0 {
		index := -1
		for i, cdr := range sq.Cdrs {
			if time.Now().Sub(cdr.SetupTime) > sq.conf.TimeWindow {
				sq.removeFromMetrics(cdr)
				index = i
				continue
			}
			break
		}
		if index != -1 {
			if index > 0 {
				sq.Cdrs = sq.Cdrs[index:]
			} else {
				sq.Cdrs = make([]*QCdr, 0)
			}
		}
	}
}

func (sq *StatsQueue) GetStats() map[string]float64 {
	sq.mux.Lock()
	defer sq.mux.Unlock()
	sq.purgeObsoleteCdrs()
	return sq.getStats()
}

func (sq *StatsQueue) getStats() map[string]float64 {
	stat := make(map[string]float64, len(sq.metrics))
	for key, metric := range sq.metrics {
		stat[key] = metric.GetValue()
	}
	return stat
}

func (sq *StatsQueue) GetId() string {
	return sq.conf.Id
}

// Convert data into a struct which can be used in actions based on triggers hit
func (sq *StatsQueue) Triggered(at *ActionTrigger) *StatsQueueTriggered {
	return &StatsQueueTriggered{Id: sq.conf.Id, Metrics: sq.getStats(), Trigger: at}
}

// Struct to be passed to triggered actions
type StatsQueueTriggered struct {
	Id      string // StatsQueueId
	Metrics map[string]float64
	Trigger *ActionTrigger
}
