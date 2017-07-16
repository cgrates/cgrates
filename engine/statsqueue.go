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
	"fmt"
	"sync"
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
	SqID      string                // StatsQueueID
	SEvents   map[string]StatsEvent // Events used by SQItems
	SQItems   []*SQItem             // SQItems
	SQMetrics map[string][]byte
}

// StatsQueue represents an individual stats instance
type StatsQueue struct {
	sync.RWMutex
	dirty     bool // needs save
	sec       *StatsEventCache
	sqItems   []*SQItem
	sqMetrics map[string]StatsMetric
	ms        Marshaler // used to get/set Metrics

	ID                 string                    // QueueID
	ActivationInterval *utils.ActivationInterval // Activation interval
	Filters            []*RequestFilter
	QueueLength        int
	TTL                *time.Duration
	Metrics            []string // list of metrics to build
	Store              bool     // store to DB
	Thresholds         []string // list of thresholds to be checked after changes
}

// Init prepares a StatsQueue for operations
// Should be executed at server start
func (sq *StatsQueue) Init(sec *StatsEventCache, ms Marshaler, sqSM *SQStoredMetrics) (err error) {
	sq.sec = sec
	if sqSM == nil {
		return
	}
	for evID, ev := range sqSM.SEvents {
		sq.sec.Cache(evID, ev, sq.ID)
	}
	sq.sqItems = sqSM.SQItems
	for metricID := range sq.sqMetrics {
		if sq.sqMetrics[metricID], err = NewStatsMetric(metricID); err != nil {
			return
		}
		if stored, has := sqSM.SQMetrics[metricID]; !has {
			continue
		} else if err = sq.sqMetrics[metricID].SetFromMarshaled(stored, ms); err != nil {
			return
		}
	}
	return
}

// GetSQStoredMetrics retrieves the data used for store to DB
func (sq *StatsQueue) GetStoredMetrics() (sqSM *SQStoredMetrics) {
	sq.RLock()
	defer sq.RUnlock()
	sEvents := make(map[string]StatsEvent)
	var sItems []*SQItem
	for _, sqItem := range sq.sqItems { // make sure event is properly retrieved from cache
		ev := sq.sec.GetEvent(sqItem.EventID)
		if ev == nil {
			utils.Logger.Warning(fmt.Sprintf("<StatsQueue> querying for storage eventID: %s, error: event not cached",
				sqItem.EventID))
			continue
		}
		sEvents[sqItem.EventID] = ev
		sItems = append(sItems, sqItem)
	}
	sqSM = &SQStoredMetrics{
		SEvents:   sEvents,
		SQItems:   sItems,
		SQMetrics: make(map[string][]byte, len(sq.sqMetrics))}
	for metricID, metric := range sq.sqMetrics {
		var err error
		if sqSM.SQMetrics[metricID], err = metric.GetMarshaled(sq.ms); err != nil {
			utils.Logger.Warning(fmt.Sprintf("<StatsQueue> querying for storage metricID: %s, error: %s",
				metricID, err.Error()))
			continue
		}
	}
	return
}

// ProcessEvent processes a StatsEvent, returns true if processed
func (sq *StatsQueue) ProcessEvent(ev StatsEvent) (err error) {
	sq.Lock()
	sq.remExpired()
	sq.remOnQueueLength()
	sq.addStatsEvent(ev)
	sq.Unlock()
	return
}

// remExpired expires items in queue
func (sq *StatsQueue) remExpired() {
	var expIdx *int // index of last item to be expired
	for i, item := range sq.sqItems {
		if item.ExpiryTime == nil {
			break
		}
		if item.ExpiryTime.After(time.Now()) {
			break
		}
		sq.remEventWithID(item.EventID)
		item = nil // garbage collected asap
		expIdx = &i
	}
	if expIdx == nil {
		return
	}
	nextValidIdx := *expIdx + 1
	sq.sqItems = sq.sqItems[nextValidIdx:]
}

// remOnQueueLength rems elements based on QueueLength setting
func (sq *StatsQueue) remOnQueueLength() {
	if sq.QueueLength == 0 {
		return
	}
	if len(sq.sqItems) == sq.QueueLength { // reached limit, rem first element
		itm := sq.sqItems[0]
		sq.remEventWithID(itm.EventID)
		itm = nil
		sq.sqItems = sq.sqItems[1:]
	}
}

// addStatsEvent computes metrics for an event
func (sq *StatsQueue) addStatsEvent(ev StatsEvent) {
	evID := ev.ID()
	for metricID, metric := range sq.sqMetrics {
		if err := metric.AddEvent(ev); err != nil {
			utils.Logger.Warning(fmt.Sprintf("<StatsQueue> metricID: %s, add eventID: %s, error: %s",
				metricID, evID, err.Error()))
		}
	}
}

// remStatsEvent removes an event from metrics
func (sq *StatsQueue) remEventWithID(evID string) {
	ev := sq.sec.GetEvent(evID)
	if ev == nil {
		utils.Logger.Warning(fmt.Sprintf("<StatsQueue> removing eventID: %s, error: event not cached", evID))
		return
	}
	for metricID, metric := range sq.sqMetrics {
		if err := metric.RemEvent(ev); err != nil {
			utils.Logger.Warning(fmt.Sprintf("<StatsQueue> metricID: %s, remove eventID: %s, error: %s", metricID, evID, err.Error()))
		}
	}
}
