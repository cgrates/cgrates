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
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/guardian"
)

// StatsConfig represents the configuration of a  StatsInstance in StatS
type StatQueueProfile struct {
	Tenant       string
	ID           string // QueueID
	FilterIDs    []string
	Weights      utils.DynamicWeights
	Blockers     utils.DynamicBlockers // blocker flag to stop processing on filters matched
	QueueLength  int
	TTL          time.Duration
	MinItems     int
	Stored       bool
	ThresholdIDs []string             // list of thresholds to be checked after changes
	Metrics      []*MetricWithFilters // list of metrics to build

	lkID string // holds the reference towards guardian lock key
}

// Clone clones *StatQueueProfile (lkID excluded)
func (sqp *StatQueueProfile) Clone() *StatQueueProfile {
	if sqp == nil {
		return nil
	}
	result := &StatQueueProfile{
		Tenant:      sqp.Tenant,
		ID:          sqp.ID,
		QueueLength: sqp.QueueLength,
		TTL:         sqp.TTL,
		MinItems:    sqp.MinItems,
		Stored:      sqp.Stored,
	}
	if sqp.FilterIDs != nil {
		result.FilterIDs = make([]string, len(sqp.FilterIDs))
		copy(result.FilterIDs, sqp.FilterIDs)
	}
	if sqp.ThresholdIDs != nil {
		result.ThresholdIDs = make([]string, len(sqp.ThresholdIDs))
		copy(result.ThresholdIDs, sqp.ThresholdIDs)
	}
	if sqp.Weights != nil {
		result.Weights = sqp.Weights.Clone()
	}
	if sqp.Blockers != nil {
		result.Blockers = sqp.Blockers.Clone()
	}
	if sqp.Metrics != nil {
		result.Metrics = make([]*MetricWithFilters, len(sqp.Metrics))
		for i, metric := range sqp.Metrics {
			result.Metrics[i] = metric.Clone()
		}
	}
	return result
}

// CacheClone returns a clone of StatQueueProfile used by ltcache CacheCloner
func (sqp *StatQueueProfile) CacheClone() any {
	return sqp.Clone()
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
	tmp := sqp.lkID
	sqp.lkID = utils.EmptyString
	guardian.Guardian.UnguardIDs(tmp)
}

// isLocked returns the locks status of this StatQueueProfile
func (sqp *StatQueueProfile) isLocked() bool {
	return sqp.lkID != utils.EmptyString
}

type MetricWithFilters struct {
	MetricID  string
	FilterIDs []string
	Blockers  utils.DynamicBlockers // blocker flag to stop processing for next metric on filters matched
}

// Clone clones *MetricWithFilters
func (mwf *MetricWithFilters) Clone() *MetricWithFilters {
	if mwf == nil {
		return nil
	}
	result := &MetricWithFilters{
		MetricID: mwf.MetricID,
	}
	if mwf.FilterIDs != nil {
		result.FilterIDs = make([]string, len(mwf.FilterIDs))
		copy(result.FilterIDs, mwf.FilterIDs)
	}
	if mwf.Blockers != nil {
		result.Blockers = mwf.Blockers.Clone()
	}
	return result
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

type StatQueueWithAPIOpts struct {
	StatQueue *StatQueue
	APIOpts   map[string]any
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

type SQItem struct {
	EventID    string     // Bounded to the original utils.CGREvent
	ExpiryTime *time.Time // Used to auto-expire events
}

func NewStatQueue(tnt, id string, metrics []*MetricWithFilters, minItems uint64) (sq *StatQueue, err error) {
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

// TenantID will compose the unique identifier for the StatQueue out of Tenant and ID
func (sq *StatQueue) TenantID() string {
	return utils.ConcatenatedKey(sq.Tenant, sq.ID)
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
	dDP := NewDynamicDP(ctx, config.CgrConfig().FilterSCfg().ResourceSConns, config.CgrConfig().FilterSCfg().StatSConns,
		config.CgrConfig().FilterSCfg().AccountSConns, config.CgrConfig().FilterSCfg().TrendSConns, config.CgrConfig().FilterSCfg().RankingSConns, tnt, utils.MapStorage{utils.MetaReq: evNm[utils.MetaReq], utils.MetaOpts: evNm[utils.MetaOpts]})
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

	dDP := NewDynamicDP(ctx, config.CgrConfig().FilterSCfg().ResourceSConns, config.CgrConfig().FilterSCfg().StatSConns,
		config.CgrConfig().FilterSCfg().AccountSConns, config.CgrConfig().FilterSCfg().TrendSConns, config.CgrConfig().FilterSCfg().RankingSConns, tnt, utils.MapStorage{utils.MetaReq: evNm[utils.MetaReq], utils.MetaOpts: evNm[utils.MetaOpts]})

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

func (sq *StatQueue) Compress(maxQL uint64) bool {
	if uint64(len(sq.SQItems)) < maxQL || maxQL == 0 {
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
		for _, id := range m.Compress(maxQL, defaultCompressID) {
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
			return newSQItems[i].ExpiryTime.Before(*newSQItems[j].ExpiryTime)
		})
	}
	sq.SQItems = newSQItems
	return true
}

func (sq *StatQueue) Expand() {
	compressFactorMap := make(map[string]uint64)
	for _, m := range sq.SQMetrics {
		compressFactorMap = m.GetCompressFactor(compressFactorMap)
	}
	var newSQItems []SQItem
	for _, sqi := range sq.SQItems {
		cf, has := compressFactorMap[sqi.EventID]
		if !has {
			continue
		}
		for i := uint64(0); i < cf; i++ {
			newSQItems = append(newSQItems, sqi)
		}
	}
	sq.SQItems = newSQItems
}

func (sq *StatQueue) MarshalJSON() (rply []byte, err error) {
	type tmp StatQueue
	sq.lock(utils.EmptyString)
	rply, err = json.Marshal(tmp(*sq))
	sq.unlock()
	return
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
		case utils.MetaHighest:
			metric = new(StatHighest)
		case utils.MetaLowest:
			metric = new(StatLowest)
		case utils.MetaREPSC:
			metric = new(StatREPSC)
		case utils.MetaREPFC:
			metric = new(StatREPFC)
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

type sqEncode StatQueue

func (sq StatQueue) GobEncode() (rply []byte, err error) {
	buf := bytes.NewBuffer(rply)
	sq.lock(utils.EmptyString)
	err = gob.NewEncoder(buf).Encode(sqEncode(sq))
	sq.unlock()
	return buf.Bytes(), err
}

func (sq *StatQueue) GobDecode(rply []byte) (err error) {
	buf := bytes.NewBuffer(rply)
	var eSq sqEncode
	err = gob.NewDecoder(buf).Decode(&eSq)
	*sq = StatQueue(eSq)
	return err
}

func (sq *StatQueue) Clone() (cln *StatQueue) {
	if sq == nil {
		return nil
	}
	cln = &StatQueue{
		Tenant: sq.Tenant,
		ID:     sq.ID,
	}
	if sq.SQItems != nil {
		cln.SQItems = make([]SQItem, len(sq.SQItems))
		for i, itm := range sq.SQItems {
			var exp *time.Time
			if itm.ExpiryTime != nil {
				exp = new(time.Time)
				*exp = *itm.ExpiryTime
			}
			cln.SQItems[i] = SQItem{EventID: itm.EventID, ExpiryTime: exp}
		}
	}
	if sq.SQMetrics != nil {
		cln.SQMetrics = make(map[string]StatMetric, len(sq.SQMetrics))
		for k, m := range sq.SQMetrics {
			if m != nil {
				cln.SQMetrics[k] = m.Clone()
			}
		}
	}
	if sq.sqPrfl != nil {
		cln.sqPrfl = sq.sqPrfl.Clone()
	}
	if sq.dirty != nil {
		dirtyVal := *sq.dirty
		cln.dirty = &dirtyVal
	}
	if sq.ttl != nil {
		ttlVal := *sq.ttl
		cln.ttl = &ttlVal
	}
	return
}

// CacheClone returns a clone of StatQueue used by ltcache CacheCloner
func (sq *StatQueue) CacheClone() any {
	return sq.Clone()
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

func (ssq *StatQueueWithAPIOpts) MarshalJSON() (rply []byte, err error) {
	if ssq == nil {
		return []byte("null"), nil
	}
	type tmp struct {
		StatQueue
		APIOpts map[string]any
	}
	rply, err = json.Marshal(tmp{
		StatQueue: *ssq.StatQueue,
		APIOpts:   ssq.APIOpts,
	})
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

func (sqp *StatQueueProfile) Set(path []string, val any, newBranch bool) (err error) {
	switch len(path) {
	default:
		return utils.ErrWrongPath
	case 1:
		if val == utils.EmptyString {
			return
		}
		switch path[0] {
		default:
			return utils.ErrWrongPath
		case utils.Tenant:
			sqp.Tenant = utils.IfaceAsString(val)
		case utils.ID:
			sqp.ID = utils.IfaceAsString(val)

		case utils.QueueLength:
			sqp.QueueLength, err = utils.IfaceAsInt(val)
		case utils.MinItems:
			sqp.MinItems, err = utils.IfaceAsInt(val)
		case utils.TTL:
			sqp.TTL, err = utils.IfaceAsDuration(val)
		case utils.Stored:
			sqp.Stored, err = utils.IfaceAsBool(val)
		case utils.Blockers:
			if val != utils.EmptyString {
				sqp.Blockers, err = utils.NewDynamicBlockersFromString(utils.IfaceAsString(val), utils.InfieldSep, utils.ANDSep)
			}
		case utils.Weights:
			if val != utils.EmptyString {
				sqp.Weights, err = utils.NewDynamicWeightsFromString(utils.IfaceAsString(val), utils.InfieldSep, utils.ANDSep)
			}

		case utils.FilterIDs:
			var valA []string
			valA, err = utils.IfaceAsStringSlice(val)
			sqp.FilterIDs = append(sqp.FilterIDs, valA...)
		case utils.ThresholdIDs:
			var valA []string
			valA, err = utils.IfaceAsStringSlice(val)
			sqp.ThresholdIDs = append(sqp.ThresholdIDs, valA...)
		}
	case 2:
		// path =[]string{Metrics, MetricID}
		if path[0] != utils.Metrics {
			return utils.ErrWrongPath
		}
		// val := *acd;*tcd;*asr
		if val != utils.EmptyString {
			if len(sqp.Metrics) == 0 || newBranch {
				sqp.Metrics = append(sqp.Metrics, new(MetricWithFilters))
			}
			switch path[1] {
			case utils.FilterIDs:
				var valA []string
				valA, err = utils.IfaceAsStringSlice(val)
				sqp.Metrics[len(sqp.Metrics)-1].FilterIDs = append(sqp.Metrics[len(sqp.Metrics)-1].FilterIDs, valA...)
			case utils.MetricID:
				valA := utils.InfieldSplit(utils.IfaceAsString(val))
				sqp.Metrics[len(sqp.Metrics)-1].MetricID = valA[0]
				for _, mID := range valA[1:] { // add the rest of the metrics
					sqp.Metrics = append(sqp.Metrics, &MetricWithFilters{MetricID: mID})
				}
			case utils.Blockers:
				var blkrs utils.DynamicBlockers
				if blkrs, err = utils.NewDynamicBlockersFromString(utils.IfaceAsString(val), utils.InfieldSep, utils.ANDSep); err != nil {
					return
				}
				sqp.Metrics[len(sqp.Metrics)-1].Blockers = append(sqp.Metrics[len(sqp.Metrics)-1].Blockers, blkrs...)
			default:
				return utils.ErrWrongPath
			}
		}
	}
	return
}

func (sqp *StatQueueProfile) Merge(v2 any) {
	vi := v2.(*StatQueueProfile)
	if len(vi.Tenant) != 0 {
		sqp.Tenant = vi.Tenant
	}
	if len(vi.ID) != 0 {
		sqp.ID = vi.ID
	}
	sqp.FilterIDs = append(sqp.FilterIDs, vi.FilterIDs...)
	sqp.ThresholdIDs = append(sqp.ThresholdIDs, vi.ThresholdIDs...)
	sqp.Metrics = append(sqp.Metrics, vi.Metrics...)

	if vi.QueueLength != 0 {
		sqp.QueueLength = vi.QueueLength
	}
	if vi.TTL != 0 {
		sqp.TTL = vi.TTL
	}
	if vi.MinItems != 0 {
		sqp.MinItems = vi.MinItems
	}
	if vi.Stored {
		sqp.Stored = vi.Stored
	}
	sqp.Weights = append(sqp.Weights, vi.Weights...)
	sqp.Blockers = append(sqp.Blockers, vi.Blockers...)
}

func (sqp *StatQueueProfile) String() string { return utils.ToJSON(sqp) }
func (sqp *StatQueueProfile) FieldAsString(fldPath []string) (_ string, err error) {
	var val any
	if val, err = sqp.FieldAsInterface(fldPath); err != nil {
		return
	}
	return utils.IfaceAsString(val), nil
}
func (sqp *StatQueueProfile) FieldAsInterface(fldPath []string) (_ any, err error) {
	if len(fldPath) == 1 {
		switch fldPath[0] {
		default:
			fld, idx := utils.GetPathIndex(fldPath[0])
			if idx != nil {
				switch fld {
				case utils.ThresholdIDs:
					if *idx < len(sqp.ThresholdIDs) {
						return sqp.ThresholdIDs[*idx], nil
					}
				case utils.FilterIDs:
					if *idx < len(sqp.FilterIDs) {
						return sqp.FilterIDs[*idx], nil
					}
				case utils.Metrics:
					if *idx < len(sqp.Metrics) {
						return sqp.Metrics[*idx], nil
					}
				}
			}
			return nil, utils.ErrNotFound
		case utils.Tenant:
			return sqp.Tenant, nil
		case utils.ID:
			return sqp.ID, nil
		case utils.FilterIDs:
			return sqp.FilterIDs, nil
		case utils.Weights:
			return sqp.Weights, nil
		case utils.ThresholdIDs:
			return sqp.ThresholdIDs, nil
		case utils.QueueLength:
			return sqp.QueueLength, nil
		case utils.TTL:
			return sqp.TTL, nil
		case utils.MinItems:
			return sqp.MinItems, nil
		case utils.Metrics:
			return sqp.Metrics, nil
		case utils.Stored:
			return sqp.Stored, nil
		case utils.Blockers:
			return sqp.Blockers, nil
		}
	}
	if len(fldPath) == 0 {
		return nil, utils.ErrNotFound
	}
	fld, idx := utils.GetPathIndex(fldPath[0])
	if fld != utils.Metrics ||
		idx == nil {
		return nil, utils.ErrNotFound
	}
	if *idx >= len(sqp.Metrics) {
		return nil, utils.ErrNotFound
	}
	return sqp.Metrics[*idx].FieldAsInterface(fldPath[1:])
}

func (mf *MetricWithFilters) String() string { return utils.ToJSON(mf) }
func (mf *MetricWithFilters) FieldAsString(fldPath []string) (_ string, err error) {
	var val any
	if val, err = mf.FieldAsInterface(fldPath); err != nil {
		return
	}
	return utils.IfaceAsString(val), nil
}
func (mf *MetricWithFilters) FieldAsInterface(fldPath []string) (_ any, err error) {
	if len(fldPath) != 1 {
		return nil, utils.ErrNotFound
	}
	switch fldPath[0] {
	default:
		fld, idx := utils.GetPathIndex(fldPath[0])
		if fld == utils.FilterIDs &&
			idx != nil &&
			*idx < len(mf.FilterIDs) {
			return mf.FilterIDs[*idx], nil
		}
		return nil, utils.ErrNotFound
	case utils.MetricID:
		return mf.MetricID, nil
	case utils.FilterIDs:
		return mf.FilterIDs, nil
	case utils.Blockers:
		return mf.Blockers, nil
	}
}

// AsMapStringInterface converts StatQueueProfile struct to map[string]any
func (sqp *StatQueueProfile) AsMapStringInterface() map[string]any {
	if sqp == nil {
		return nil
	}
	return map[string]any{
		utils.Tenant:       sqp.Tenant,
		utils.ID:           sqp.ID,
		utils.FilterIDs:    sqp.FilterIDs,
		utils.Weights:      sqp.Weights,
		utils.Blockers:     sqp.Blockers,
		utils.QueueLength:  sqp.QueueLength,
		utils.TTL:          sqp.TTL,
		utils.MinItems:     sqp.MinItems,
		utils.Stored:       sqp.Stored,
		utils.ThresholdIDs: sqp.ThresholdIDs,
		utils.Metrics:      sqp.Metrics,
	}
}

// MapStringInterfaceToStatQueueProfile converts map[string]any to StatQueueProfile struct
func MapStringInterfaceToStatQueueProfile(m map[string]any) (*StatQueueProfile, error) {
	sqp := &StatQueueProfile{}
	if v, ok := m[utils.Tenant].(string); ok {
		sqp.Tenant = v
	}
	if v, ok := m[utils.ID].(string); ok {
		sqp.ID = v
	}
	sqp.FilterIDs = utils.InterfaceToStringSlice(m[utils.FilterIDs])
	sqp.Weights = utils.InterfaceToDynamicWeights(m[utils.Weights])
	sqp.Blockers = utils.InterfaceToDynamicBlockers(m[utils.Blockers])
	if v, ok := m[utils.QueueLength].(float64); ok {
		sqp.QueueLength = int(v)
	}
	if v, ok := m[utils.TTL].(string); ok {
		if dur, err := time.ParseDuration(v); err != nil {
			return nil, err
		} else {
			sqp.TTL = dur
		}
	} else if v, ok := m[utils.TTL].(float64); ok { // for -1 cases
		sqp.TTL = time.Duration(v)
	}
	if v, ok := m[utils.MinItems].(float64); ok {
		sqp.MinItems = int(v)
	}
	if v, ok := m[utils.Stored].(bool); ok {
		sqp.Stored = v
	}
	sqp.ThresholdIDs = utils.InterfaceToStringSlice(m[utils.ThresholdIDs])
	sqp.Metrics = InterfaceToMetrics(m[utils.Metrics])
	return sqp, nil
}

func InterfaceToMetrics(v any) []*MetricWithFilters {
	if v == nil {
		return nil
	}
	switch val := v.(type) {
	case []any:
		result := make([]*MetricWithFilters, 0, len(val))
		for _, item := range val {
			if itemMap, ok := item.(map[string]any); ok {
				metric := &MetricWithFilters{}
				if metricID, ok := itemMap[utils.MetricID].(string); ok {
					metric.MetricID = metricID
				}
				metric.FilterIDs = utils.InterfaceToStringSlice(itemMap[utils.FilterIDs])
				metric.Blockers = utils.InterfaceToDynamicBlockers(itemMap[utils.Blockers])
				result = append(result, metric)
			}
		}
		return result
	}
	return nil
}
