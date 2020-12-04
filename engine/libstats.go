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
	"sort"
	"sync"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

// StatsConfig represents the configuration of a  StatsInstance in StatS
type StatQueueProfile struct {
	Tenant             string
	ID                 string // QueueID
	FilterIDs          []string
	ActivationInterval *utils.ActivationInterval // Activation interval
	QueueLength        int
	TTL                time.Duration
	MinItems           int
	Metrics            []*MetricWithFilters // list of metrics to build
	Stored             bool
	Blocker            bool // blocker flag to stop processing on filters matched
	Weight             float64
	ThresholdIDs       []string // list of thresholds to be checked after changes
}

// StatQueueProfileWithOpts is used in replicatorV1 for dispatcher
type StatQueueProfileWithOpts struct {
	*StatQueueProfile
	Opts map[string]interface{}
}

type StatQueueWithCache struct {
	*StatQueueProfile
	Cache *string
	Opts  map[string]interface{}
}

func (sqp *StatQueueProfile) TenantID() string {
	return utils.ConcatenatedKey(sqp.Tenant, sqp.ID)
}

type MetricWithFilters struct {
	FilterIDs []string
	MetricID  string
}

// NewStoredStatQueue initiates a StoredStatQueue out of StatQueue
func NewStoredStatQueue(sq *StatQueue, ms Marshaler) (sSQ *StoredStatQueue, err error) {
	sSQ = &StoredStatQueue{
		Tenant: sq.Tenant,
		ID:     sq.ID,
		Compressed: sq.Compress(int64(config.CgrConfig().StatSCfg().StoreUncompressedLimit),
			config.CgrConfig().GeneralCfg().RoundingDecimals),
		SQItems:   make([]SQItem, len(sq.SQItems)),
		SQMetrics: make(map[string][]byte, len(sq.SQMetrics)),
	}
	for i, sqItm := range sq.SQItems {
		sSQ.SQItems[i] = sqItm
	}
	for metricID, metric := range sq.SQMetrics {
		marshaled, err := metric.Marshal(ms)
		if err != nil {
			return nil, err
		}
		sSQ.SQMetrics[metricID] = marshaled
	}
	return
}

// StoredStatQueue differs from StatQueue due to serialization of SQMetrics
type StoredStatQueue struct {
	Tenant     string
	ID         string
	SQItems    []SQItem
	SQMetrics  map[string][]byte
	Compressed bool
}

type StoredStatQueueWithOpts struct {
	*StoredStatQueue
	Opts map[string]interface{}
}

// SqID will compose the unique identifier for the StatQueue out of Tenant and ID
func (ssq *StoredStatQueue) SqID() string {
	return utils.ConcatenatedKey(ssq.Tenant, ssq.ID)
}

// AsStatQueue converts into StatQueue unmarshaling SQMetrics
func (ssq *StoredStatQueue) AsStatQueue(ms Marshaler) (sq *StatQueue, err error) {
	if ssq == nil {
		return
	}
	sq = &StatQueue{
		Tenant:    ssq.Tenant,
		ID:        ssq.ID,
		SQItems:   make([]SQItem, len(ssq.SQItems)),
		SQMetrics: make(map[string]StatMetric, len(ssq.SQMetrics)),
	}
	for i, sqItm := range ssq.SQItems {
		sq.SQItems[i] = sqItm
	}
	for metricID, marshaled := range ssq.SQMetrics {
		if metric, err := NewStatMetric(metricID, 0, []string{}); err != nil {
			return nil, err
		} else if err := metric.LoadMarshaled(ms, marshaled); err != nil {
			return nil, err
		} else {
			sq.SQMetrics[metricID] = metric
		}
	}
	if ssq.Compressed {
		sq.Expand()
	}
	return
}

type SQItem struct {
	EventID    string     // Bounded to the original utils.CGREvent
	ExpiryTime *time.Time // Used to auto-expire events
}

// StatQueue represents an individual stats instance
type StatQueue struct {
	lk        sync.RWMutex // protect the elements from within
	Tenant    string
	ID        string
	SQItems   []SQItem
	SQMetrics map[string]StatMetric
	sqPrfl    *StatQueueProfile
	dirty     *bool          // needs save
	ttl       *time.Duration // timeToLeave, picked on each init
}

// RLock only to implement sync.RWMutex methods
func (sq *StatQueue) RLock() { sq.lk.RLock() }

// RUnlock only to implement sync.RWMutex methods
func (sq *StatQueue) RUnlock() { sq.lk.RUnlock() }

// Lock only to implement sync.RWMutex methods
func (sq *StatQueue) Lock() { sq.lk.Lock() }

// Unlock only to implement sync.RWMutex methods
func (sq *StatQueue) Unlock() { sq.lk.Unlock() }

// TenantID will compose the unique identifier for the StatQueue out of Tenant and ID
func (sq *StatQueue) TenantID() string {
	return utils.ConcatenatedKey(sq.Tenant, sq.ID)
}

// ProcessEvent processes a utils.CGREvent, returns true if processed
func (sq *StatQueue) ProcessEvent(tnt, evID string, filterS *FilterS, evNm utils.MapStorage) (err error) {
	if _, err = sq.remExpired(); err != nil {
		return
	}
	if err = sq.remOnQueueLength(); err != nil {
		return
	}
	return sq.addStatEvent(tnt, evID, filterS, evNm)
}

// remStatEvent removes an event from metrics
func (sq *StatQueue) remEventWithID(evID string) (err error) {
	for metricID, metric := range sq.SQMetrics {
		if err = metric.RemEvent(evID); err != nil {
			if err.Error() == utils.ErrNotFound.Error() {
				err = nil
				continue
			}
			utils.Logger.Warning(fmt.Sprintf("<StatQueue> metricID: %s, remove eventID: %s, error: %s", metricID, evID, err.Error()))
			return
		}
	}
	return
}

// remExpired expires items in queue
func (sq *StatQueue) remExpired() (removed int, err error) {
	var expIdx *int // index of last item to be expired
	for i, item := range sq.SQItems {
		if item.ExpiryTime == nil {
			break // items are ordered, so no need to look further
		}
		if item.ExpiryTime.After(time.Now()) {
			break
		}
		if err = sq.remEventWithID(item.EventID); err != nil {
			return
		}
		expIdx = utils.IntPointer(i)
	}
	if expIdx == nil {
		return
	}
	removed = *expIdx + 1
	sq.SQItems = sq.SQItems[removed:]
	return
}

// remOnQueueLength removes elements based on QueueLength setting
func (sq *StatQueue) remOnQueueLength() (err error) {
	if sq.sqPrfl.QueueLength <= 0 { // infinite length
		return
	}
	if len(sq.SQItems) == sq.sqPrfl.QueueLength { // reached limit, rem first element
		item := sq.SQItems[0]
		if err = sq.remEventWithID(item.EventID); err != nil {
			return
		}
		sq.SQItems = sq.SQItems[1:]
	}
	return
}

// addStatEvent computes metrics for an event
func (sq *StatQueue) addStatEvent(tnt, evID string, filterS *FilterS, evNm utils.MapStorage) (err error) {
	var expTime *time.Time
	if sq.ttl != nil {
		expTime = utils.TimePointer(time.Now().Add(*sq.ttl))
	}
	sq.SQItems = append(sq.SQItems, SQItem{EventID: evID, ExpiryTime: expTime})
	var pass bool
	// recreate the request without *opts
	dDP := newDynamicDP(config.CgrConfig().FilterSCfg().ResourceSConns, config.CgrConfig().FilterSCfg().StatSConns,
		config.CgrConfig().FilterSCfg().ApierSConns, tnt, utils.MapStorage{utils.MetaReq: evNm[utils.MetaReq]})
	for metricID, metric := range sq.SQMetrics {
		if pass, err = filterS.Pass(tnt, metric.GetFilterIDs(),
			evNm); err != nil {
			return
		} else if !pass {
			continue
		}
		if err = metric.AddEvent(evID, dDP); err != nil {
			utils.Logger.Warning(fmt.Sprintf("<StatQueue> metricID: %s, add eventID: %s, error: %s",
				metricID, evID, err.Error()))
			return
		}
	}
	return
}

func (sq *StatQueue) Compress(maxQL int64, roundDec int) bool {
	if int64(len(sq.SQItems)) < maxQL || maxQL == 0 {
		return false
	}
	var newSQItems []SQItem
	sqMap := make(map[string]*time.Time)
	idMap := make(utils.StringSet)
	defaultCompressID := sq.SQItems[len(sq.SQItems)-1].EventID
	defaultTTL := sq.SQItems[len(sq.SQItems)-1].ExpiryTime

	for _, sqitem := range sq.SQItems {
		sqMap[sqitem.EventID] = sqitem.ExpiryTime
	}

	for _, m := range sq.SQMetrics {
		for _, id := range m.Compress(maxQL, defaultCompressID, roundDec) {
			idMap.Add(id)
		}
	}
	for k := range idMap {
		ttl, has := sqMap[k]
		if !has { // log warning
			ttl = defaultTTL
		}
		newSQItems = append(newSQItems, SQItem{
			EventID:    k,
			ExpiryTime: ttl,
		})
	}
	if sq.ttl != nil {
		sort.Slice(newSQItems, func(i, j int) bool {
			if newSQItems[i].ExpiryTime == nil {
				return false
			}
			if newSQItems[j].ExpiryTime == nil {
				return true
			}
			return newSQItems[i].ExpiryTime.Before(*(newSQItems[j].ExpiryTime))
		})
	}
	sq.SQItems = newSQItems
	return true
}

func (sq *StatQueue) Expand() {
	compressFactorMap := make(map[string]int)
	for _, m := range sq.SQMetrics {
		compressFactorMap = m.GetCompressFactor(compressFactorMap)
	}
	var newSQItems []SQItem
	for _, sqi := range sq.SQItems {
		cf, has := compressFactorMap[sqi.EventID]
		if !has {
			continue
		}
		for i := 0; i < cf; i++ {
			newSQItems = append(newSQItems, sqi)
		}
	}
	sq.SQItems = newSQItems
}

// StatQueues is a sortable list of StatQueue
type StatQueues []*StatQueue

// Sort is part of sort interface, sort based on Weight
func (sis StatQueues) Sort() {
	sort.Slice(sis, func(i, j int) bool { return sis[i].sqPrfl.Weight > sis[j].sqPrfl.Weight })
}
