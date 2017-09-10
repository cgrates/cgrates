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
	"errors"
	"sort"
	"time"

	"github.com/cgrates/cgrates/utils"
)

// StatsConfig represents the configuration of a  StatsInstance in StatS
type StatQueueProfile struct {
	Tenant             string
	ID                 string // QueueID
	Filters            []*RequestFilter
	ActivationInterval *utils.ActivationInterval // Activation interval
	QueueLength        int
	TTL                time.Duration
	Metrics            []string // list of metrics to build
	Store              bool     // store to DB
	Thresholds         []string // list of thresholds to be checked after changes
	Blocker            bool     // blocker flag to stop processing on filters matched
	Stored             bool
	Weight             float64
}

// StatEvent is an event processed by StatService
type StatEvent struct {
	Tenant string
	ID     string
	Fields map[string]interface{}
}

// AnswerTime returns the AnswerTime of StatEvent
func (se StatEvent) AnswerTime(timezone string) (at time.Time, err error) {
	atIf, has := se.Fields[utils.ANSWER_TIME]
	if !has {
		return at, utils.ErrNotFound
	}
	if at, canCast := atIf.(time.Time); canCast {
		return at, nil
	}
	atStr, canCast := atIf.(string)
	if !canCast {
		return at, errors.New("cannot cast to string")
	}
	return utils.ParseTimeDetectLayout(atStr, timezone)
}

// StatQueue represents an individual stats instance
type StatQueue struct {
	Tenant  string
	ID      string
	SQItems []struct {
		EventID    string     // Bounded to the original StatEvent
		ExpiryTime *time.Time // Used to auto-expire events
	}
	SQMetrics map[string]StatMetric
	sqPrfl    *StatQueueProfile
	dirty     *bool // needs save
}

/*
// GetSQStoredMetrics retrieves the data used for store to DB
func (sq *StatQueue) GetStoredMetrics() (sqSM *engine.SQStoredMetrics) {
	sq.RLock()
	defer sq.RUnlock()
	sEvents := make(map[string]engine.StatEvent)
	var sItems []*engine.SQItem
	for _, sqItem := range sq.sqItems { // make sure event is properly retrieved from cache
		ev := sq.sec.GetEvent(sqItem.EventID)
		if ev == nil {
			utils.Logger.Warning(fmt.Sprintf("<StatQueue> querying for storage eventID: %s, error: event not cached",
				sqItem.EventID))
			continue
		}
		sEvents[sqItem.EventID] = ev
		sItems = append(sItems, sqItem)
	}
	sqSM = &engine.SQStoredMetrics{
		SEvents:   sEvents,
		SQItems:   sItems,
		SQMetrics: make(map[string][]byte, len(sq.sqMetrics))}
	for metricID, metric := range sq.sqMetrics {
		var err error
		if sqSM.SQMetrics[metricID], err = metric.GetMarshaled(sq.ms); err != nil {
			utils.Logger.Warning(fmt.Sprintf("<StatQueue> querying for storage metricID: %s, error: %s",
				metricID, err.Error()))
			continue
		}
	}
	return
}

// ProcessEvent processes a StatEvent, returns true if processed
func (sq *StatQueue) ProcessEvent(ev engine.StatEvent) (err error) {
	sq.Lock()
	sq.remExpired()
	sq.remOnQueueLength()
	sq.addStatEvent(ev)
	sq.Unlock()
	return
}

// remExpired expires items in queue
func (sq *StatQueue) remExpired() {
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
func (sq *StatQueue) remOnQueueLength() {
	if sq.cfg.QueueLength == 0 {
		return
	}
	if len(sq.sqItems) == sq.cfg.QueueLength { // reached limit, rem first element
		itm := sq.sqItems[0]
		sq.remEventWithID(itm.EventID)
		itm = nil
		sq.sqItems = sq.sqItems[1:]
	}
}

// addStatEvent computes metrics for an event
func (sq *StatQueue) addStatEvent(ev engine.StatEvent) {
	evID := ev.ID()
	for metricID, metric := range sq.sqMetrics {
		if err := metric.AddEvent(ev); err != nil {
			utils.Logger.Warning(fmt.Sprintf("<StatQueue> metricID: %s, add eventID: %s, error: %s",
				metricID, evID, err.Error()))
		}
	}
}

// remStatEvent removes an event from metrics
func (sq *StatQueue) remEventWithID(evID string) {
	ev := sq.sec.GetEvent(evID)
	if ev == nil {
		utils.Logger.Warning(fmt.Sprintf("<StatQueue> removing eventID: %s, error: event not cached", evID))
		return
	}
	for metricID, metric := range sq.sqMetrics {
		if err := metric.RemEvent(ev); err != nil {
			utils.Logger.Warning(fmt.Sprintf("<StatQueue> metricID: %s, remove eventID: %s, error: %s", metricID, evID, err.Error()))
		}
	}
}
*/

// StatQueues is a sortable list of StatQueue
type StatQueues []*StatQueue

// Sort is part of sort interface, sort based on Weight
func (sis StatQueues) Sort() {
	sort.Slice(sis, func(i, j int) bool { return sis[i].sqPrfl.Weight > sis[j].sqPrfl.Weight })
}
