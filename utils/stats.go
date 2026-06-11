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

package utils

import (
	"encoding/json"
	"fmt"
	"slices"
	"sort"
	"strings"
	"time"
)

// StatQueueProfile the profile for a StatQueue
type StatQueueProfile struct {
	Tenant       string
	ID           string // QueueID
	FilterIDs    []string
	Weights      DynamicWeights
	Blockers     DynamicBlockers // blocker flag to stop processing on filters matched
	QueueLength  int
	TTL          time.Duration
	MinItems     int
	Stored       bool
	ThresholdIDs []string             // list of thresholds to be checked after changes
	Metrics      []*MetricWithFilters // list of metrics to build
}

// Clone clones *StatQueueProfile
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
	return ConcatenatedKey(sqp.Tenant, sqp.ID)
}

type MetricWithFilters struct {
	MetricID  string
	FilterIDs []string
	Blockers  DynamicBlockers // blocker flag to stop processing for next metric on filters matched
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

type StatQueueWithAPIOpts struct {
	StatQueue *StatQueue
	APIOpts   map[string]any
}

type SQItem struct {
	EventID    string     // Bounded to the original CGREvent
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
}

// TenantID will compose the unique identifier for the StatQueue out of Tenant and ID
func (sq *StatQueue) TenantID() string {
	return ConcatenatedKey(sq.Tenant, sq.ID)
}

func (sq *StatQueue) Compress(maxQL uint64) bool {
	if uint64(len(sq.SQItems)) < maxQL || maxQL == 0 {
		return false
	}
	var newSQItems []SQItem
	sqMap := make(map[string]*time.Time)
	idMap := make(StringSet)
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
	if slices.ContainsFunc(sq.SQItems, func(it SQItem) bool { return it.ExpiryTime != nil }) {
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
		metricSplit := strings.Split(metricID, HashtagSep)
		var metric StatMetric
		switch metricSplit[0] {
		case MetaASR:
			metric = new(StatASR)
		case MetaACD:
			metric = new(StatACD)
		case MetaTCD:
			metric = new(StatTCD)
		case MetaACC:
			metric = new(StatACC)
		case MetaTCC:
			metric = new(StatTCC)
		case MetaPDD:
			metric = new(StatPDD)
		case MetaDDC:
			metric = new(StatDDC)
		case MetaSum:
			metric = new(StatSum)
		case MetaAverage:
			metric = new(StatAverage)
		case MetaDistinct:
			metric = new(StatDistinct)
		case MetaHighest:
			metric = new(StatHighest)
		case MetaLowest:
			metric = new(StatLowest)
		case MetaREPSC:
			metric = new(StatREPSC)
		case MetaREPFC:
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
	return
}

// CacheClone returns a clone of StatQueue used by ltcache CacheCloner
func (sq *StatQueue) CacheClone() any {
	return sq.Clone()
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
		return ErrWrongPath
	case 1:
		if val == EmptyString {
			return
		}
		switch path[0] {
		default:
			return ErrWrongPath
		case Tenant:
			sqp.Tenant = IfaceAsString(val)
		case ID:
			sqp.ID = IfaceAsString(val)

		case QueueLength:
			sqp.QueueLength, err = IfaceAsInt(val)
		case MinItems:
			sqp.MinItems, err = IfaceAsInt(val)
		case TTL:
			sqp.TTL, err = IfaceAsDuration(val)
		case Stored:
			sqp.Stored, err = IfaceAsBool(val)
		case Blockers:
			if val != EmptyString {
				sqp.Blockers, err = NewDynamicBlockersFromString(IfaceAsString(val), InfieldSep, ANDSep)
			}
		case Weights:
			if val != EmptyString {
				sqp.Weights, err = NewDynamicWeightsFromString(IfaceAsString(val), InfieldSep, ANDSep)
			}

		case FilterIDs:
			var valA []string
			valA, err = IfaceAsStringSlice(val)
			sqp.FilterIDs = append(sqp.FilterIDs, valA...)
		case ThresholdIDs:
			var valA []string
			valA, err = IfaceAsStringSlice(val)
			sqp.ThresholdIDs = append(sqp.ThresholdIDs, valA...)
		}
	case 2:
		// path =[]string{Metrics, MetricID}
		if path[0] != Metrics {
			return ErrWrongPath
		}
		// val := *acd;*tcd;*asr
		if val != EmptyString {
			if len(sqp.Metrics) == 0 || newBranch {
				sqp.Metrics = append(sqp.Metrics, new(MetricWithFilters))
			}
			switch path[1] {
			case FilterIDs:
				var valA []string
				valA, err = IfaceAsStringSlice(val)
				sqp.Metrics[len(sqp.Metrics)-1].FilterIDs = append(sqp.Metrics[len(sqp.Metrics)-1].FilterIDs, valA...)
			case MetricID:
				valA := InfieldSplit(IfaceAsString(val))
				sqp.Metrics[len(sqp.Metrics)-1].MetricID = valA[0]
				for _, mID := range valA[1:] { // add the rest of the metrics
					sqp.Metrics = append(sqp.Metrics, &MetricWithFilters{MetricID: mID})
				}
			case Blockers:
				var blkrs DynamicBlockers
				if blkrs, err = NewDynamicBlockersFromString(IfaceAsString(val), InfieldSep, ANDSep); err != nil {
					return
				}
				sqp.Metrics[len(sqp.Metrics)-1].Blockers = append(sqp.Metrics[len(sqp.Metrics)-1].Blockers, blkrs...)
			default:
				return ErrWrongPath
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

func (sqp *StatQueueProfile) String() string { return ToJSON(sqp) }
func (sqp *StatQueueProfile) FieldAsString(fldPath []string) (_ string, err error) {
	var val any
	if val, err = sqp.FieldAsInterface(fldPath); err != nil {
		return
	}
	return IfaceAsString(val), nil
}
func (sqp *StatQueueProfile) FieldAsInterface(fldPath []string) (_ any, err error) {
	if len(fldPath) == 1 {
		switch fldPath[0] {
		default:
			fld, idx := GetPathIndex(fldPath[0])
			if idx != nil {
				switch fld {
				case ThresholdIDs:
					if *idx < len(sqp.ThresholdIDs) {
						return sqp.ThresholdIDs[*idx], nil
					}
				case FilterIDs:
					if *idx < len(sqp.FilterIDs) {
						return sqp.FilterIDs[*idx], nil
					}
				case Metrics:
					if *idx < len(sqp.Metrics) {
						return sqp.Metrics[*idx], nil
					}
				}
			}
			return nil, ErrNotFound
		case Tenant:
			return sqp.Tenant, nil
		case ID:
			return sqp.ID, nil
		case FilterIDs:
			return sqp.FilterIDs, nil
		case Weights:
			return sqp.Weights, nil
		case ThresholdIDs:
			return sqp.ThresholdIDs, nil
		case QueueLength:
			return sqp.QueueLength, nil
		case TTL:
			return sqp.TTL, nil
		case MinItems:
			return sqp.MinItems, nil
		case Metrics:
			return sqp.Metrics, nil
		case Stored:
			return sqp.Stored, nil
		case Blockers:
			return sqp.Blockers, nil
		}
	}
	if len(fldPath) == 0 {
		return nil, ErrNotFound
	}
	fld, idx := GetPathIndex(fldPath[0])
	if fld != Metrics ||
		idx == nil {
		return nil, ErrNotFound
	}
	if *idx >= len(sqp.Metrics) {
		return nil, ErrNotFound
	}
	return sqp.Metrics[*idx].FieldAsInterface(fldPath[1:])
}

func (mf *MetricWithFilters) String() string { return ToJSON(mf) }
func (mf *MetricWithFilters) FieldAsString(fldPath []string) (_ string, err error) {
	var val any
	if val, err = mf.FieldAsInterface(fldPath); err != nil {
		return
	}
	return IfaceAsString(val), nil
}
func (mf *MetricWithFilters) FieldAsInterface(fldPath []string) (_ any, err error) {
	if len(fldPath) != 1 {
		return nil, ErrNotFound
	}
	switch fldPath[0] {
	default:
		fld, idx := GetPathIndex(fldPath[0])
		if fld == FilterIDs &&
			idx != nil &&
			*idx < len(mf.FilterIDs) {
			return mf.FilterIDs[*idx], nil
		}
		return nil, ErrNotFound
	case MetricID:
		return mf.MetricID, nil
	case FilterIDs:
		return mf.FilterIDs, nil
	case Blockers:
		return mf.Blockers, nil
	}
}

// AsMapStringInterface converts StatQueueProfile struct to map[string]any
func (sqp *StatQueueProfile) AsMapStringInterface() map[string]any {
	if sqp == nil {
		return nil
	}
	return map[string]any{
		Tenant:       sqp.Tenant,
		ID:           sqp.ID,
		FilterIDs:    sqp.FilterIDs,
		Weights:      sqp.Weights,
		Blockers:     sqp.Blockers,
		QueueLength:  sqp.QueueLength,
		TTL:          sqp.TTL,
		MinItems:     sqp.MinItems,
		Stored:       sqp.Stored,
		ThresholdIDs: sqp.ThresholdIDs,
		Metrics:      sqp.Metrics,
	}
}

// MapStringInterfaceToStatQueueProfile converts map[string]any to StatQueueProfile struct
func MapStringInterfaceToStatQueueProfile(m map[string]any) (*StatQueueProfile, error) {
	sqp := &StatQueueProfile{}
	if v, ok := m[Tenant].(string); ok {
		sqp.Tenant = v
	}
	if v, ok := m[ID].(string); ok {
		sqp.ID = v
	}
	sqp.FilterIDs = InterfaceToStringSlice(m[FilterIDs])
	sqp.Weights = InterfaceToDynamicWeights(m[Weights])
	sqp.Blockers = InterfaceToDynamicBlockers(m[Blockers])
	if v, ok := m[QueueLength].(float64); ok {
		sqp.QueueLength = int(v)
	}
	if v, ok := m[TTL].(string); ok {
		if dur, err := time.ParseDuration(v); err != nil {
			return nil, err
		} else {
			sqp.TTL = dur
		}
	} else if v, ok := m[TTL].(float64); ok { // for -1 cases
		sqp.TTL = time.Duration(v)
	}
	if v, ok := m[MinItems].(float64); ok {
		sqp.MinItems = int(v)
	}
	if v, ok := m[Stored].(bool); ok {
		sqp.Stored = v
	}
	sqp.ThresholdIDs = InterfaceToStringSlice(m[ThresholdIDs])
	sqp.Metrics = InterfaceToMetrics(m[Metrics])
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
				if metricID, ok := itemMap[MetricID].(string); ok {
					metric.MetricID = metricID
				}
				metric.FilterIDs = InterfaceToStringSlice(itemMap[FilterIDs])
				metric.Blockers = InterfaceToDynamicBlockers(itemMap[Blockers])
				result = append(result, metric)
			}
		}
		return result
	}
	return nil
}

// StatQueueLockKey returns the ID used to lock a StatQueue with guardian.
func StatQueueLockKey(tnt, id string) string {
	return ConcatenatedKey(CacheStatQueues, tnt, id)
}
