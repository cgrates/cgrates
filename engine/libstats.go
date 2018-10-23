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
	"github.com/cgrates/cgrates/utils"
	"sort"
	"time"
)

// StatsConfig represents the configuration of a  StatsInstance in StatS
type StatQueueProfile struct {
	Tenant             string
	ID                 string // QueueID
	FilterIDs          []string
	ActivationInterval *utils.ActivationInterval // Activation interval
	QueueLength        int
	TTL                time.Duration
	Metrics            []*utils.MetricWithParams // list of metrics to build
	ThresholdIDs       []string                  // list of thresholds to be checked after changes
	Blocker            bool                      // blocker flag to stop processing on filters matched
	Stored             bool
	Weight             float64
	MinItems           int
}

func (sqp *StatQueueProfile) TenantID() string {
	return utils.ConcatenatedKey(sqp.Tenant, sqp.ID)
}

// NewStoredStatQueue initiates a StoredStatQueue out of StatQueue
func NewStoredStatQueue(sq *StatQueue, ms Marshaler) (sSQ *StoredStatQueue, err error) {
	sSQ = &StoredStatQueue{
		Tenant: sq.Tenant,
		ID:     sq.ID,
		SQItems: make([]struct {
			EventID    string
			ExpiryTime *time.Time
		}, len(sq.SQItems)),
		SQMetrics: make(map[string][]byte, len(sq.SQMetrics)),
		MinItems:  sq.MinItems,
	}
	for i, sqItm := range sq.SQItems {
		sSQ.SQItems[i] = sqItm
	}
	for metricID, metric := range sq.SQMetrics {
		if marshaled, err := metric.Marshal(ms); err != nil {
			return nil, err
		} else {
			sSQ.SQMetrics[metricID] = marshaled
		}
	}
	return
}

// StoredStatQueue differs from StatQueue due to serialization of SQMetrics
type StoredStatQueue struct {
	Tenant  string
	ID      string
	SQItems []struct {
		EventID    string     // Bounded to the original utils.CGREvent
		ExpiryTime *time.Time // Used to auto-expire events
	}
	SQMetrics map[string][]byte
	MinItems  int
}

// SqID will compose the unique identifier for the StatQueue out of Tenant and ID
func (ssq *StoredStatQueue) SqID() string {
	return utils.ConcatenatedKey(ssq.Tenant, ssq.ID)
}

// AsStatQueue converts into StatQueue unmarshaling SQMetrics
func (ssq *StoredStatQueue) AsStatQueue(ms Marshaler) (sq *StatQueue, err error) {
	sq = &StatQueue{
		Tenant: ssq.Tenant,
		ID:     ssq.ID,
		SQItems: make([]struct {
			EventID    string
			ExpiryTime *time.Time
		}, len(ssq.SQItems)),
		SQMetrics: make(map[string]StatMetric, len(ssq.SQMetrics)),
		MinItems:  ssq.MinItems,
	}
	for i, sqItm := range ssq.SQItems {
		sq.SQItems[i] = sqItm
	}
	for metricID, marshaled := range ssq.SQMetrics {
		if metric, err := NewStatMetric(metricID, ssq.MinItems, ""); err != nil {
			return nil, err
		} else if err := metric.LoadMarshaled(ms, marshaled); err != nil {
			return nil, err
		} else {
			sq.SQMetrics[metricID] = metric
		}
	}
	return
}

// StatQueue represents an individual stats instance
type StatQueue struct {
	Tenant  string
	ID      string
	SQItems []struct {
		EventID    string     // Bounded to the original utils.CGREvent
		ExpiryTime *time.Time // Used to auto-expire events
	}
	SQMetrics map[string]StatMetric
	MinItems  int
	sqPrfl    *StatQueueProfile
	dirty     *bool          // needs save
	ttl       *time.Duration // timeToLeave, picked on each init
}

// SqID will compose the unique identifier for the StatQueue out of Tenant and ID
func (sq *StatQueue) TenantID() string {
	return utils.ConcatenatedKey(sq.Tenant, sq.ID)
}

// ProcessEvent processes a utils.CGREvent, returns true if processed
func (sq *StatQueue) ProcessEvent(ev *utils.CGREvent) (err error) {
	sq.remExpired()
	sq.remOnQueueLength()
	sq.addStatEvent(ev)
	return
}

// remStatEvent removes an event from metrics
func (sq *StatQueue) remEventWithID(evID string) {
	for metricID, metric := range sq.SQMetrics {
		if err := metric.RemEvent(evID); err != nil {
			utils.Logger.Warning(fmt.Sprintf("<StatQueue> metricID: %s, remove eventID: %s, error: %s", metricID, evID, err.Error()))
		}
	}
}

// remExpired expires items in queue
func (sq *StatQueue) remExpired() {
	var expIdx *int // index of last item to be expired
	for i, item := range sq.SQItems {
		if item.ExpiryTime == nil {
			break // items are ordered, so no need to look further
		}
		if item.ExpiryTime.After(time.Now()) {
			break
		}
		sq.remEventWithID(item.EventID)
		expIdx = utils.IntPointer(i)
	}
	if expIdx == nil {
		return
	}
	sq.SQItems = sq.SQItems[*expIdx+1:]
}

// remOnQueueLength removes elements based on QueueLength setting
func (sq *StatQueue) remOnQueueLength() {
	if sq.sqPrfl.QueueLength <= 0 { // infinite length
		return
	}
	if len(sq.SQItems) == sq.sqPrfl.QueueLength { // reached limit, rem first element
		itm := sq.SQItems[0]
		sq.remEventWithID(itm.EventID)
		sq.SQItems = sq.SQItems[1:]
	}
}

// addStatEvent computes metrics for an event
func (sq *StatQueue) addStatEvent(ev *utils.CGREvent) {
	var expTime *time.Time
	if sq.ttl != nil {
		expTime = utils.TimePointer(time.Now().Add(*sq.ttl))
	}
	sq.SQItems = append(sq.SQItems,
		struct {
			EventID    string
			ExpiryTime *time.Time
		}{ev.ID, expTime})

	for metricID, metric := range sq.SQMetrics {
		if err := metric.AddEvent(ev); err != nil {
			utils.Logger.Warning(fmt.Sprintf("<StatQueue> metricID: %s, add eventID: %s, error: %s",
				metricID, ev.ID, err.Error()))
		}
	}
}

// StatQueues is a sortable list of StatQueue
type StatQueues []*StatQueue

// Sort is part of sort interface, sort based on Weight
func (sis StatQueues) Sort() {
	sort.Slice(sis, func(i, j int) bool { return sis[i].sqPrfl.Weight > sis[j].sqPrfl.Weight })
}
