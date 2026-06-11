/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/

package engine

import (
	"encoding/base64"
	"fmt"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/guardian"
)

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
	tmp := sqp.lkID
	sqp.lkID = utils.EmptyString
	guardian.Guardian.UnguardIDs(tmp)
}

// isLocked returns the locks status of this StatQueueProfile
func (sqp *StatQueueProfile) isLocked() bool {
	return sqp.lkID != utils.EmptyString
}

// NewStoredStatQueue initiates a StoredStatQueue out of StatQueue
func NewStoredStatQueue(sq *StatQueue, ms utils.Marshaler) (sSQ *StoredStatQueue, err error) {
	sSQ = &StoredStatQueue{
		Tenant:     sq.Tenant,
		ID:         sq.ID,
		Compressed: sq.Compress(uint64(config.CgrConfig().StatSCfg().StoreUncompressedLimit)),
		SQItems:    make([]SQItem, len(sq.SQItems)),
		SQMetrics:  make(map[string][]byte, len(sq.SQMetrics)),
	}
	copy(sSQ.SQItems, sq.SQItems)
	for metricID, metric := range sq.SQMetrics {
		marshaled, err := ms.Marshal(metric)
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

// SqID will compose the unique identifier for the StatQueue out of Tenant and ID
func (ssq *StoredStatQueue) SqID() string {
	return utils.ConcatenatedKey(ssq.Tenant, ssq.ID)
}

// AsStatQueue converts into StatQueue unmarshaling SQMetrics
func (ssq *StoredStatQueue) AsStatQueue(ms utils.Marshaler) (sq *StatQueue, err error) {
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
		metric, err := NewStatMetric(metricID, 0, []string{})
		if err != nil {
			return nil, err
		}
		if err := ms.Unmarshal(marshaled, metric); err != nil {
			return nil, err
		}
		sq.SQMetrics[metricID] = metric
	}
	if ssq.Compressed {
		sq.Expand()
	}
	return
}

// statQueueLockKey returns the ID used to lock a StatQueue with guardian
func statQueueLockKey(tnt, id string) string {
	return utils.ConcatenatedKey(utils.CacheStatQueues, tnt, id)
}

// lock will lock the StatQueue using guardian and store the lock within r.lkID
// if lkID is passed as argument, the lock is considered as executed
func (sq *StatQueue) lock(lkID string) {
	if lkID == utils.EmptyString {
		lkID = guardian.Guardian.GuardIDs(utils.EmptyString,
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
	tmp := sq.lkID
	sq.lkID = utils.EmptyString
	guardian.Guardian.UnguardIDs(tmp)
}

// isLocked returns the locks status of this StatQueue
func (sq *StatQueue) isLocked() bool {
	return sq.lkID != utils.EmptyString
}

// ProcessEvent processes a utils.CGREvent, returns true if processed
func (sq *StatQueue) ProcessEvent(ctx *context.Context, tnt, evID string, filterS *FilterS, evNm utils.MapStorage) (err error) {

	//processing metrics without storing in the queue
	if oneEv := sq.isOneEvent(); oneEv {
		return sq.addStatOneEvent(ctx, tnt, filterS, evNm)
	}
	if _, err = sq.remExpired(); err != nil {
		return
	}
	if err = sq.remOnQueueLength(); err != nil {
		return
	}
	return sq.addStatEvent(ctx, tnt, evID, filterS, evNm)
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
func (sq *StatQueue) addStatEvent(ctx *context.Context, tnt, evID string, filterS *FilterS, evNm utils.MapStorage) (err error) {
	var expTime *time.Time
	if sq.ttl != nil {
		expTime = utils.TimePointer(time.Now().Add(*sq.ttl))
	}
	sq.SQItems = append(sq.SQItems, SQItem{EventID: evID, ExpiryTime: expTime})
	var pass bool
	// recreate the request without *opts
	metricEvNm := utils.MapStorage{utils.MetaReq: evNm[utils.MetaReq], utils.MetaOpts: evNm[utils.MetaOpts]}

	dDP := NewDynamicDP(ctx, config.CgrConfig(), tnt, metricEvNm, filterS)
	for idx, metricCfg := range sq.sqPrfl.Metrics {
		if pass, err = filterS.Pass(ctx, tnt, metricCfg.FilterIDs,
			evNm); err != nil {
			return
		} else if !pass {
			continue
		}
		// in case of # metrics type
		if err = sq.SQMetrics[metricCfg.MetricID].AddEvent(evID, dDP); err != nil {
			utils.Logger.Warning(fmt.Sprintf("<StatQueue>: metric: %s, add eventID: %s, error: %s", metricCfg.MetricID,
				evID, err.Error()))
			return
		}
		// every metric has a blocker, verify them
		var blocker bool
		if blocker, err = BlockerFromDynamics(ctx, metricCfg.Blockers, filterS, tnt, evNm); err != nil {
			return
		}
		if blocker && idx != len(sq.sqPrfl.Metrics)-1 {
			break
		}
	}
	return
}

func (sq *StatQueue) isOneEvent() bool {
	return sq.ttl != nil && *sq.ttl == -1
}

func (sq *StatQueue) addStatOneEvent(ctx *context.Context, tnt string, filterS *FilterS, evNm utils.MapStorage) (err error) {
	var pass bool

	metricEvNm := utils.MapStorage{utils.MetaReq: evNm[utils.MetaReq], utils.MetaOpts: evNm[utils.MetaOpts]}
	dDP := NewDynamicDP(ctx, config.CgrConfig(), tnt, metricEvNm, filterS)

	for idx, metricCfg := range sq.sqPrfl.Metrics {
		if pass, err = filterS.Pass(ctx, tnt, metricCfg.FilterIDs,
			evNm); err != nil {
			return
		} else if !pass {
			continue
		}

		if err = sq.SQMetrics[metricCfg.MetricID].AddOneEvent(dDP); err != nil {
			utils.Logger.Warning(fmt.Sprintf("<StatQueue>: metric: %s, error: %s", metricCfg.MetricID, err.Error()))
			return
		}

		var blocker bool
		if blocker, err = BlockerFromDynamics(ctx, metricCfg.Blockers, filterS, tnt, evNm); err != nil {
			return
		}
		if blocker && idx != len(sq.sqPrfl.Metrics)-1 {
			break
		}
	}
	return
}

// AsMapStringInterface converts StoredStatQueue struct to map[string]any
func (ssq *StoredStatQueue) AsMapStringInterface() map[string]any {
	if ssq == nil {
		return nil
	}
	return map[string]any{
		utils.Tenant:     ssq.Tenant,
		utils.ID:         ssq.ID,
		utils.SQItems:    ssq.SQItems,
		utils.SQMetrics:  ssq.SQMetrics,
		utils.Compressed: ssq.Compressed,
	}
}

// MapStringInterfaceToStoredStatQueue converts map[string]any to StoredStatQueue struct
func MapStringInterfaceToStoredStatQueue(m map[string]any) (*StoredStatQueue, error) {
	ssq := &StoredStatQueue{}
	if v, ok := m[utils.Tenant].(string); ok {
		ssq.Tenant = v
	}
	if v, ok := m[utils.ID].(string); ok {
		ssq.ID = v
	}
	if items, ok := m[utils.SQItems].([]any); ok {
		for _, item := range items {
			if itemMap, ok := item.(map[string]any); ok {
				sqItem := SQItem{}
				if eventID, ok := itemMap[utils.EventID].(string); ok {
					sqItem.EventID = eventID
				}
				if expiryTime, ok := itemMap[utils.ExpiryTime].(*time.Time); ok {
					sqItem.ExpiryTime = expiryTime
				} else if expiryStr, ok := itemMap[utils.ExpiryTime].(string); ok {
					if parsedTime, err := time.Parse(time.RFC3339, expiryStr); err == nil {
						sqItem.ExpiryTime = &parsedTime
					} else {
						return nil, err
					}
				}
				ssq.SQItems = append(ssq.SQItems, sqItem)
			}
		}
	}
	if metrics, ok := m[utils.SQMetrics].(map[string]any); ok {
		ssq.SQMetrics = make(map[string][]byte)
		for key, value := range metrics {
			if metricBytes, ok := value.(string); ok {
				decodedBytes, err := base64.StdEncoding.DecodeString(metricBytes)
				if err != nil {
					return nil, fmt.Errorf("failed to decode base64 string: %v", err)
				}
				ssq.SQMetrics[key] = decodedBytes
			}
		}
	}
	if v, ok := m[utils.Compressed].(bool); ok {
		ssq.Compressed = v
	}
	return ssq, nil
}

// unlockStatQueues unlocks all locked StatQueues in the given slice.
func unlockStatQueues(sqs []*StatQueue) {
	for _, sq := range sqs {
		sq.unlock()
		if sq.sqPrfl != nil {
			sq.sqPrfl.unlock()
		}
	}
}

// getStatQueueIDs returns a slice of IDs from the given StatQueues
func getStatQueueIDs(sqs []*StatQueue) []string {
	ids := make([]string, len(sqs))
	for i, sq := range sqs {
		ids[i] = sq.ID
	}
	return ids
}
