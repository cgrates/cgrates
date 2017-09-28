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
	"fmt"
	"github.com/cgrates/cgrates/utils"
	"sort"
	"strconv"
	"time"
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
	Thresholds         []string // list of thresholds to be checked after changes
	Blocker            bool     // blocker flag to stop processing on filters matched
	Stored             bool
	Weight             float64
}

func (sqp *StatQueueProfile) TenantID() string {
	return utils.ConcatenatedKey(sqp.Tenant, sqp.ID)
}

// StatEvent is an event processed by StatService
type StatEvent struct {
	Tenant string
	ID     string
	Fields map[string]interface{}
}

// TenantID returns the unique identifier based on Tenant and ID
func (se StatEvent) TenantID() string {
	return utils.ConcatenatedKey(se.Tenant, se.ID)
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

// Usage returns the Usage of StatEvent
func (se StatEvent) Usage(timezone string) (at time.Duration, err error) {
	usIf, has := se.Fields[utils.USAGE]
	if !has {
		return at, utils.ErrNotFound
	}
	if us, canCast := usIf.(time.Duration); canCast {
		return us, nil
	}
	if us, canCast := usIf.(float64); canCast {
		return time.Duration(int64(us)), nil
	}
	usStr, canCast := usIf.(string)
	if !canCast {
		return at, errors.New("cannot cast to string")
	}
	return utils.ParseDurationWithSecs(usStr)
}

// Cost returns the Cost of StatEvent
func (se StatEvent) Cost(timezone string) (cs float64, err error) {
	csIf, has := se.Fields[utils.COST]
	if !has {
		return cs, utils.ErrNotFound
	}
	if val, canCast := csIf.(float64); canCast {
		return val, nil
	}
	csStr, canCast := csIf.(string)
	if !canCast {
		return cs, errors.New("cannot cast to string")
	}
	return strconv.ParseFloat(csStr, 64)
}

// Pdd returns the Pdd of StatEvent
func (se StatEvent) Pdd(timezone string) (pdd time.Duration, err error) {
	pddIf, has := se.Fields[utils.PDD]
	if !has {
		return pdd, utils.ErrNotFound
	}
	if pdd, canCast := pddIf.(time.Duration); canCast {
		return pdd, nil
	}
	if pdd, canCast := pddIf.(float64); canCast {
		return time.Duration(int64(pdd)), nil
	}
	pddStr, canCast := pddIf.(string)
	if !canCast {
		return pdd, errors.New("cannot cast to string")
	}
	return utils.ParseDurationWithSecs(pddStr)
}

// Destination returns the Destination of StatEvent
func (se StatEvent) Destination(timezone string) (ddc string, err error) {
	ddcIf, has := se.Fields[utils.DESTINATION]
	if !has {
		return ddc, utils.ErrNotFound
	}
	if ddcInt, canCast := ddcIf.(int64); canCast {
		return strconv.FormatInt(ddcInt, 64), nil
	}
	ddcStr, canCast := ddcIf.(string)
	if !canCast {
		return ddc, errors.New("cannot cast to string")
	}
	return ddcStr, nil
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
		EventID    string     // Bounded to the original StatEvent
		ExpiryTime *time.Time // Used to auto-expire events
	}
	SQMetrics map[string][]byte
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
	}
	for i, sqItm := range ssq.SQItems {
		sq.SQItems[i] = sqItm
	}
	for metricID, marshaled := range ssq.SQMetrics {
		if metric, err := NewStatMetric(metricID); err != nil {
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
		EventID    string     // Bounded to the original StatEvent
		ExpiryTime *time.Time // Used to auto-expire events
	}
	SQMetrics map[string]StatMetric
	sqPrfl    *StatQueueProfile
	dirty     *bool          // needs save
	ttl       *time.Duration // timeToLeave, picked on each init
}

// SqID will compose the unique identifier for the StatQueue out of Tenant and ID
func (sq *StatQueue) TenantID() string {
	return utils.ConcatenatedKey(sq.Tenant, sq.ID)
}

// ProcessEvent processes a StatEvent, returns true if processed
func (sq *StatQueue) ProcessEvent(ev *StatEvent) (err error) {
	sq.remExpired()
	sq.remOnQueueLength()
	sq.addStatEvent(ev)
	return
}

// remStatEvent removes an event from metrics
func (sq *StatQueue) remEventWithID(evTenantID string) {
	for metricID, metric := range sq.SQMetrics {
		if err := metric.RemEvent(evTenantID); err != nil {
			utils.Logger.Warning(fmt.Sprintf("<StatQueue> metricID: %s, remove eventID: %s, error: %s", metricID, evTenantID, err.Error()))
		}
	}
}

// remExpired expires items in queue
func (sq *StatQueue) remExpired() {
	var expIdx *int // index of last item to be expired
	for i, item := range sq.SQItems {
		if item.ExpiryTime == nil {
			break
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
func (sq *StatQueue) addStatEvent(ev *StatEvent) {
	for metricID, metric := range sq.SQMetrics {
		if err := metric.AddEvent(ev); err != nil {
			utils.Logger.Warning(fmt.Sprintf("<StatQueue> metricID: %s, add eventID: %s, error: %s",
				metricID, ev.TenantID(), err.Error()))
		}
	}
}

// StatQueues is a sortable list of StatQueue
type StatQueues []*StatQueue

// Sort is part of sort interface, sort based on Weight
func (sis StatQueues) Sort() {
	sort.Slice(sis, func(i, j int) bool { return sis[i].sqPrfl.Weight > sis[j].sqPrfl.Weight })
}
