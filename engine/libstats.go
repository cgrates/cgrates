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
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/guardian"
	"github.com/cgrates/cgrates/utils"
)

// StatsConfig represents the configuration of a  StatsInstance in StatS
type StatQueueProfile struct {
	Tenant       string
	ID           string // QueueID
	FilterIDs    []string
	QueueLength  int
	TTL          time.Duration
	MinItems     int
	Metrics      []*MetricWithFilters // list of metrics to build
	Stored       bool
	Blocker      bool // blocker flag to stop processing on filters matched
	Weights      utils.DynamicWeights
	ThresholdIDs []string // list of thresholds to be checked after changes

	lkID string // holds the reference towards guardian lock key
}

type statprfWithWeight struct {
	*StatQueueProfile
	weight float64
}

// StatQueueProfileWithAPIOpts is used in replicatorV1 for dispatcher
type StatQueueProfileWithAPIOpts struct {
	*StatQueueProfile
	APIOpts map[string]interface{}
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
func NewStoredStatQueue(sq *StatQueue, ms utils.Marshaler) (sSQ *StoredStatQueue, err error) {
	sSQ = &StoredStatQueue{
		Tenant:     sq.Tenant,
		ID:         sq.ID,
		Compressed: sq.Compress(uint64(config.CgrConfig().StatSCfg().StoreUncompressedLimit)),
		SQItems:    make([]SQItem, len(sq.SQItems)),
		SQMetrics:  make(map[string][]byte, len(sq.SQMetrics)),
	}
	for i, sqItm := range sq.SQItems {
		sSQ.SQItems[i] = sqItm
	}
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
	APIOpts   map[string]interface{}
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
	for i, sqItm := range ssq.SQItems {
		sq.SQItems[i] = sqItm
	}
	for metricID, marshaled := range ssq.SQMetrics {
		if metric, err := NewStatMetric(metricID, 0, []string{}); err != nil {
			return nil, err
		} else if err := ms.Unmarshal(marshaled, metric); err != nil {
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
	weight    float64
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
func (sq *StatQueue) ProcessEvent(ctx *context.Context, tnt, evID string, filterS *FilterS, evNm utils.MapStorage) (err error) {
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
	dDP := newDynamicDP(ctx, config.CgrConfig().FilterSCfg().ResourceSConns, config.CgrConfig().FilterSCfg().StatSConns,
		config.CgrConfig().FilterSCfg().AccountSConns, tnt, utils.MapStorage{utils.MetaReq: evNm[utils.MetaReq]})
	for metricID, metric := range sq.SQMetrics {
		if pass, err = filterS.Pass(ctx, tnt, metric.GetFilterIDs(),
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
			return newSQItems[i].ExpiryTime.Before(*(newSQItems[j].ExpiryTime))
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

// StatQueues is a sortable list of StatQueue
type StatQueues []*StatQueue

// Sort is part of sort interface, sort based on Weight
func (sis StatQueues) Sort() {
	sort.Slice(sis, func(i, j int) bool {
		return sis[i].weight > sis[j].weight
	})
}

func (sq *StatQueue) MarshalJSON() (rply []byte, err error) {
	type tmp StatQueue
	sq.lock(utils.EmptyString)
	rply, err = json.Marshal(tmp(*sq))
	sq.unlock()
	return
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
	cln = &StatQueue{
		Tenant:    sq.Tenant,
		ID:        sq.ID,
		SQItems:   make([]SQItem, len(sq.SQItems)),
		SQMetrics: make(map[string]StatMetric),
	}
	for i, itm := range sq.SQItems {
		var exp *time.Time
		if itm.ExpiryTime != nil {
			exp = new(time.Time)
			*exp = *itm.ExpiryTime
		}
		cln.SQItems[i] = SQItem{EventID: itm.EventID, ExpiryTime: exp}
	}
	for k, m := range sq.SQMetrics {
		cln.SQMetrics[k] = m.Clone()
	}
	return
}

func (ssq *StatQueueWithAPIOpts) MarshalJSON() (rply []byte, err error) {
	if ssq == nil {
		return []byte("null"), nil
	}
	type tmp struct {
		StatQueue
		APIOpts map[string]interface{}
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
		APIOpts map[string]interface{}
	}{}
	if err = json.Unmarshal(data, &i); err != nil {
		return
	}
	ssq.StatQueue = sq
	ssq.APIOpts = i.APIOpts
	return
}

func (sqp *StatQueueProfile) Set(path []string, val interface{}, newBranch bool, _ string) (err error) {
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
			sqp.QueueLength, err = utils.IfaceAsTInt(val)
		case utils.MinItems:
			sqp.MinItems, err = utils.IfaceAsTInt(val)
		case utils.TTL:
			sqp.TTL, err = utils.IfaceAsDuration(val)
		case utils.Stored:
			sqp.Stored, err = utils.IfaceAsBool(val)
		case utils.Blocker:
			sqp.Blocker, err = utils.IfaceAsBool(val)
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
		if path[0] != utils.Metrics {
			return utils.ErrWrongPath
		}
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

			default:
				return utils.ErrWrongPath
			}
		}
	}
	return
}

func (sqp *StatQueueProfile) Merge(v2 interface{}) {
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
	if vi.Blocker {
		sqp.Blocker = vi.Blocker
	}
	if vi.Stored {
		sqp.Stored = vi.Stored
	}
	sqp.Weights = append(sqp.Weights, vi.Weights...)
}

func (sqp *StatQueueProfile) String() string { return utils.ToJSON(sqp) }
func (sqp *StatQueueProfile) FieldAsString(fldPath []string) (_ string, err error) {
	var val interface{}
	if val, err = sqp.FieldAsInterface(fldPath); err != nil {
		return
	}
	return utils.IfaceAsString(val), nil
}
func (sqp *StatQueueProfile) FieldAsInterface(fldPath []string) (_ interface{}, err error) {
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
		case utils.Weight:
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
		case utils.Blocker:
			return sqp.Blocker, nil
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
	var val interface{}
	if val, err = mf.FieldAsInterface(fldPath); err != nil {
		return
	}
	return utils.IfaceAsString(val), nil
}
func (mf *MetricWithFilters) FieldAsInterface(fldPath []string) (_ interface{}, err error) {
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
	}
}
