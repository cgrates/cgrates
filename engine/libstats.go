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
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/guardian"
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

	lkID string // holds the reference towards guardian lock key
}

// StatQueueProfileWithAPIOpts is used in replicatorV1 for dispatcher
type StatQueueProfileWithAPIOpts struct {
	*StatQueueProfile
	APIOpts map[string]any
}

func (sqp *StatQueueProfile) TenantID() string {
	return utils.ConcatenatedKey(sqp.Tenant, sqp.ID)
}

// statQueueProfileLockKey returns the ID used to lock a StatQueueProfile with guardian
func statQueueProfileLockKey(tnt, id string) string {
	return utils.ConcatenatedKey(utils.CacheStatQueueProfiles, tnt, id)
}

// lock will lock the StatQueueProfile using guardian and store the lock within r.lkID
// if lkID is passed as argument, the lock is considered as executed
func (sqp *StatQueueProfile) lock(lkID string) {
	if lkID == utils.EmptyString {
		lkID = guardian.Guardian.GuardIDs("",
			config.CgrConfig().GeneralCfg().LockingTimeout,
			statQueueProfileLockKey(sqp.Tenant, sqp.ID))
	}
	sqp.lkID = lkID
}

// unlock will unlock the StatQueueProfile and clear rp.lkID
func (sqp *StatQueueProfile) unlock() {
	if sqp.lkID == utils.EmptyString {
		return
	}
	guardian.Guardian.UnguardIDs(sqp.lkID)
	sqp.lkID = utils.EmptyString
}

// isLocked returns the locks status of this StatQueueProfile
func (sqp *StatQueueProfile) isLocked() bool {
	return sqp.lkID != utils.EmptyString
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

	copy(sSQ.SQItems, sq.SQItems)

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

type StatQueueWithAPIOpts struct {
	*StatQueue
	APIOpts map[string]any
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

	copy(sq.SQItems, ssq.SQItems)

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

func NewStatQueue(tnt, id string, metrics []*MetricWithFilters, minItems int) (sq *StatQueue, err error) {
	sq = &StatQueue{
		Tenant:    tnt,
		ID:        id,
		SQMetrics: make(map[string]StatMetric),
	}

	for _, metric := range metrics {
		if sq.SQMetrics[metric.MetricID], err = NewStatMetric(metric.MetricID,
			minItems, metric.FilterIDs); err != nil {
			return
		}
	}
	return
}

// StatQueue represents an individual stats instance
type StatQueue struct {
	Tenant    string
	ID        string
	SQItems   []SQItem
	SQMetrics map[string]StatMetric
	lkID      string // ID of the lock used when matching the stat
	sqPrfl    *StatQueueProfile
	dirty     *bool          // needs save
	ttl       *time.Duration // timeToLeave, picked on each init
}

// statQueueLockKey returns the ID used to lock a StatQueue with guardian
func statQueueLockKey(tnt, id string) string {
	return utils.ConcatenatedKey(utils.CacheStatQueues, tnt, id)
}

// lock will lock the StatQueue using guardian and store the lock within r.lkID
// if lkID is passed as argument, the lock is considered as executed
func (sq *StatQueue) lock(lkID string) {
	if lkID == utils.EmptyString {
		lkID = guardian.Guardian.GuardIDs("",
			config.CgrConfig().GeneralCfg().LockingTimeout,
			statQueueLockKey(sq.Tenant, sq.ID))
	}
	sq.lkID = lkID
}

// unlock will unlock the StatQueue and clear r.lkID
func (sq *StatQueue) unlock() {
	if sq.lkID == utils.EmptyString {
		return
	}
	guardian.Guardian.UnguardIDs(sq.lkID)
	sq.lkID = utils.EmptyString
}

// isLocked returns the locks status of this StatQueue
func (sq *StatQueue) isLocked() bool {
	return sq.lkID != utils.EmptyString
}

// TenantID will compose the unique identifier for the StatQueue out of Tenant and ID
func (sq *StatQueue) TenantID() string {
	return utils.ConcatenatedKey(sq.Tenant, sq.ID)
}

// ProcessEvent processes a utils.CGREvent, returns true if processed
func (sq *StatQueue) ProcessEvent(tnt, evID string, filterS *FilterS, evNm utils.MapStorage) error {
	if oneEv := sq.isOneEvent(); oneEv {
		return sq.addOneEvent(tnt, filterS, evNm)
	}
	sq.remExpired()
	sq.remOnQueueLength()
	return sq.addStatEvent(tnt, evID, filterS, evNm)
}

func (sq *StatQueue) isOneEvent() bool {
	return sq.ttl != nil && *sq.ttl == -1
}

func (sq *StatQueue) addOneEvent(tnt string, filterS *FilterS, evNm utils.MapStorage) (err error) {
	var pass bool
	dDP := newDynamicDP(config.CgrConfig().FilterSCfg().ResourceSConns, config.CgrConfig().FilterSCfg().StatSConns,
		config.CgrConfig().FilterSCfg().ApierSConns, config.CgrConfig().FilterSCfg().TrendSConns, tnt, utils.MapStorage{utils.MetaReq: evNm[utils.MetaReq]})
	for metricID, metric := range sq.SQMetrics {
		if pass, err = filterS.Pass(tnt, metric.GetFilterIDs(),
			evNm); err != nil {
			return
		} else if !pass {
			continue
		}
		if err = metric.AddOneEvent(dDP); err != nil {
			utils.Logger.Warning(fmt.Sprintf("<StatQueue> metricID: %s, OneEvent, error: %s",
				metricID, err.Error()))
			return
		}
	}
	return
}

// remStatEvent removes an event from metrics
func (sq *StatQueue) remEventWithID(evID string) {
	for _, metric := range sq.SQMetrics {
		metric.RemEvent(evID)
	}
}

// remExpired expires items in queue
func (sq *StatQueue) remExpired() (removed int) {
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
	removed = *expIdx + 1
	sq.SQItems = sq.SQItems[removed:]
	return
}

// remOnQueueLength removes elements based on QueueLength setting
func (sq *StatQueue) remOnQueueLength() {
	if sq.sqPrfl.QueueLength <= 0 { // infinite length
		return
	}
	if len(sq.SQItems) == sq.sqPrfl.QueueLength { // reached limit, remove first element
		item := sq.SQItems[0]
		sq.remEventWithID(item.EventID)
		sq.SQItems = sq.SQItems[1:]
	}
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
		config.CgrConfig().FilterSCfg().ApierSConns, config.CgrConfig().FilterSCfg().TrendSConns, tnt, utils.MapStorage{utils.MetaReq: evNm[utils.MetaReq]})
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

// unlock will unlock StatQueues part of this slice
func (sis StatQueues) unlock() {
	for _, s := range sis {
		s.unlock()
		if s.sqPrfl != nil {
			s.sqPrfl.unlock()
		}
	}
}

func (sis StatQueues) IDs() []string {
	ids := make([]string, len(sis))
	for i, s := range sis {
		ids[i] = s.ID
	}
	return ids
}

// UnmarshalJSON here only to fully support json for StatQueue
func (sq *StatQueue) UnmarshalJSON(data []byte) (err error) {
	var tmp struct {
		Tenant    string
		ID        string
		SQItems   []SQItem
		SQMetrics map[string]json.RawMessage
	}
	if err = json.Unmarshal(data, &tmp); err != nil {
		return
	}
	sq.Tenant = tmp.Tenant
	sq.ID = tmp.ID
	sq.SQItems = tmp.SQItems
	sq.SQMetrics = make(map[string]StatMetric)
	for metricID, val := range tmp.SQMetrics {
		metricSplit := strings.Split(metricID, utils.HashtagSep)
		var metric StatMetric
		switch metricSplit[0] {
		case utils.MetaASR:
			metric = new(StatASR)
		case utils.MetaACD:
			metric = new(StatACD)
		case utils.MetaTCD:
			metric = new(StatTCD)
		case utils.MetaACC:
			metric = new(StatACC)
		case utils.MetaTCC:
			metric = new(StatTCC)
		case utils.MetaPDD:
			metric = new(StatPDD)
		case utils.MetaDDC:
			metric = new(StatDDC)
		case utils.MetaSum:
			metric = new(StatSum)
		case utils.MetaAverage:
			metric = new(StatAverage)
		case utils.MetaDistinct:
			metric = new(StatDistinct)
		default:
			return fmt.Errorf("unsupported metric type <%s>", metricSplit[0])
		}
		if err = json.Unmarshal([]byte(val), metric); err != nil {
			return
		}
		sq.SQMetrics[metricID] = metric
	}
	return
}

// UnmarshalJSON here only to fully support json for StatQueue
func (ssq *StatQueueWithAPIOpts) UnmarshalJSON(data []byte) (err error) {
	sq := new(StatQueue)
	if err = json.Unmarshal(data, &sq); err != nil {
		return
	}
	i := struct {
		APIOpts map[string]any
	}{}
	if err = json.Unmarshal(data, &i); err != nil {
		return
	}
	ssq.StatQueue = sq
	ssq.APIOpts = i.APIOpts
	return
}
