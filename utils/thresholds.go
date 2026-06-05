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
	"time"
)

// ThresholdProfile represents the user configuration for the threshold
type ThresholdProfile struct {
	Tenant           string
	ID               string
	FilterIDs        []string
	MaxHits          int
	MinHits          int
	MinSleep         time.Duration
	Blocker          bool           // blocker flag to stop processing on filters matched
	Weights          DynamicWeights // Weight to sort the thresholds
	AttributeIDs     []string
	ActionProfileIDs []string
	Async            bool
	EeIDs            []string
}

// Clone clones *ThresholdProfile
func (tp *ThresholdProfile) Clone() *ThresholdProfile {
	if tp == nil {
		return nil
	}
	clone := &ThresholdProfile{
		Tenant:   tp.Tenant,
		ID:       tp.ID,
		MaxHits:  tp.MaxHits,
		MinHits:  tp.MinHits,
		MinSleep: tp.MinSleep,
		Blocker:  tp.Blocker,
		Async:    tp.Async,
	}
	if tp.FilterIDs != nil {
		clone.FilterIDs = make([]string, len(tp.FilterIDs))
		copy(clone.FilterIDs, tp.FilterIDs)
	}
	if tp.ActionProfileIDs != nil {
		clone.ActionProfileIDs = make([]string, len(tp.ActionProfileIDs))
		copy(clone.ActionProfileIDs, tp.ActionProfileIDs)
	}
	if tp.Weights != nil {
		clone.Weights = tp.Weights.Clone()
	}
	if tp.EeIDs != nil {
		clone.EeIDs = make([]string, len(tp.EeIDs))
		copy(clone.EeIDs, tp.EeIDs)
	}
	if tp.AttributeIDs != nil {
		clone.AttributeIDs = make([]string, len(tp.AttributeIDs))
		copy(clone.AttributeIDs, tp.AttributeIDs)
	}
	return clone
}

// CacheClone returns a clone of ThresholdProfile used by ltcache CacheCloner
func (tp *ThresholdProfile) CacheClone() any {
	return tp.Clone()
}

// TenantID returns the concatenated key between tenant and ID
func (tp *ThresholdProfile) TenantID() string {
	return ConcatenatedKey(tp.Tenant, tp.ID)
}

func (tp *ThresholdProfile) Set(path []string, val any, _ bool) error {
	if len(path) != 1 {
		return ErrWrongPath
	}
	var err error
	switch path[0] {
	default:
		return ErrWrongPath
	case Tenant:
		tp.Tenant = IfaceAsString(val)
	case ID:
		tp.ID = IfaceAsString(val)
	case Blocker:
		tp.Blocker, err = IfaceAsBool(val)
	case Weights:
		if val != "" {
			tp.Weights, err = NewDynamicWeightsFromString(IfaceAsString(val), InfieldSep, ANDSep)
		}
	case FilterIDs:
		var valA []string
		valA, err = IfaceAsStringSlice(val)
		tp.FilterIDs = append(tp.FilterIDs, valA...)
	case MaxHits:
		if val != "" {
			tp.MaxHits, err = IfaceAsInt(val)
		}
	case MinHits:
		if val != "" {
			tp.MinHits, err = IfaceAsInt(val)
		}
	case MinSleep:
		tp.MinSleep, err = IfaceAsDuration(val)
	case ActionProfileIDs:
		var valA []string
		valA, err = IfaceAsStringSlice(val)
		tp.ActionProfileIDs = append(tp.ActionProfileIDs, valA...)
	case EeIDs:
		var valA []string
		valA, err = IfaceAsStringSlice(val)
		tp.EeIDs = append(tp.EeIDs, valA...)
	case AttributeIDs:
		var valA []string
		valA, err = IfaceAsStringSlice(val)
		tp.AttributeIDs = append(tp.AttributeIDs, valA...)
	case Async:
		tp.Async, err = IfaceAsBool(val)
	}
	return err
}

func (tp *ThresholdProfile) Merge(v2 any) {
	vi := v2.(*ThresholdProfile)
	if len(vi.Tenant) != 0 {
		tp.Tenant = vi.Tenant
	}
	if len(vi.ID) != 0 {
		tp.ID = vi.ID
	}
	tp.FilterIDs = append(tp.FilterIDs, vi.FilterIDs...)
	tp.ActionProfileIDs = append(tp.ActionProfileIDs, vi.ActionProfileIDs...)
	tp.EeIDs = append(tp.EeIDs, vi.EeIDs...)
	tp.AttributeIDs = append(tp.AttributeIDs, vi.AttributeIDs...)
	if vi.Blocker {
		tp.Blocker = vi.Blocker
	}
	if vi.Async {
		tp.Async = vi.Async
	}
	tp.Weights = append(tp.Weights, vi.Weights...)
	if vi.MaxHits != 0 {
		tp.MaxHits = vi.MaxHits
	}
	if vi.MinHits != 0 {
		tp.MinHits = vi.MinHits
	}
	if vi.MinSleep != 0 {
		tp.MinSleep = vi.MinSleep
	}
}

func (tp *ThresholdProfile) String() string { return ToJSON(tp) }
func (tp *ThresholdProfile) FieldAsString(fldPath []string) (string, error) {
	val, err := tp.FieldAsInterface(fldPath)
	if err != nil {
		return "", err
	}
	return IfaceAsString(val), nil
}
func (tp *ThresholdProfile) FieldAsInterface(fldPath []string) (any, error) {
	if len(fldPath) != 1 {
		return nil, ErrNotFound
	}
	switch fldPath[0] {
	default:
		fld, idx := GetPathIndex(fldPath[0])
		if idx != nil {
			switch fld {
			case ActionProfileIDs:
				if *idx < len(tp.ActionProfileIDs) {
					return tp.ActionProfileIDs[*idx], nil
				}
			case EeIDs:
				if *idx < len(tp.EeIDs) {
					return tp.EeIDs[*idx], nil
				}
			case FilterIDs:
				if *idx < len(tp.FilterIDs) {
					return tp.FilterIDs[*idx], nil
				}
			}
		}
		return nil, ErrNotFound
	case Tenant:
		return tp.Tenant, nil
	case ID:
		return tp.ID, nil
	case FilterIDs:
		return tp.FilterIDs, nil
	case Weights:
		return tp.Weights, nil
	case ActionProfileIDs:
		return tp.ActionProfileIDs, nil
	case MaxHits:
		return tp.MaxHits, nil
	case MinHits:
		return tp.MinHits, nil
	case MinSleep:
		return tp.MinSleep, nil
	case Blocker:
		return tp.Blocker, nil
	case Async:
		return tp.Async, nil
	case EeIDs:
		return tp.EeIDs, nil
	case AttributeIDs:
		return tp.AttributeIDs, nil
	}
}

// AsMapStringInterface converts ThresholdProfile struct to map[string]any
func (tp *ThresholdProfile) AsMapStringInterface() map[string]any {
	if tp == nil {
		return nil
	}
	return map[string]any{
		Tenant:           tp.Tenant,
		ID:               tp.ID,
		FilterIDs:        tp.FilterIDs,
		MaxHits:          tp.MaxHits,
		MinHits:          tp.MinHits,
		MinSleep:         tp.MinSleep,
		Blocker:          tp.Blocker,
		Weights:          tp.Weights,
		ActionProfileIDs: tp.ActionProfileIDs,
		Async:            tp.Async,
		EeIDs:            tp.EeIDs,
		AttributeIDs:     tp.AttributeIDs,
	}
}

// ThresholdProfileWithAPIOpts is used in replicatorV1 for dispatcher
type ThresholdProfileWithAPIOpts struct {
	*ThresholdProfile
	APIOpts map[string]any
}

// MapStringInterfaceToThresholdProfile converts map[string]any to ThresholdProfile struct
func MapStringInterfaceToThresholdProfile(m map[string]any) (*ThresholdProfile, error) {
	tp := &ThresholdProfile{}
	if v, ok := m[Tenant].(string); ok {
		tp.Tenant = v
	}
	if v, ok := m[ID].(string); ok {
		tp.ID = v
	}
	tp.FilterIDs = InterfaceToStringSlice(m[FilterIDs])
	if v, ok := m[MaxHits].(float64); ok {
		tp.MaxHits = int(v)
	}
	if v, ok := m[MinHits].(float64); ok {
		tp.MinHits = int(v)
	}
	if v, ok := m[MinSleep].(string); ok {
		dur, err := time.ParseDuration(v)
		if err != nil {
			return nil, err
		}
		tp.MinSleep = dur
	} else if v, ok := m[MinSleep].(float64); ok { // for -1 cases
		tp.MinSleep = time.Duration(v)
	}
	if v, ok := m[Blocker].(bool); ok {
		tp.Blocker = v
	}
	tp.Weights = InterfaceToDynamicWeights(m[Weights])
	tp.ActionProfileIDs = InterfaceToStringSlice(m[ActionProfileIDs])
	if v, ok := m[Async].(bool); ok {
		tp.Async = v
	}
	tp.EeIDs = InterfaceToStringSlice(m[EeIDs])
	tp.AttributeIDs = InterfaceToStringSlice(m[AttributeIDs])
	return tp, nil
}

// Threshold is the unit matched by filters
type Threshold struct {
	Tenant string
	ID     string
	Hits   int       // number of hits for this threshold
	Snooze time.Time // prevent threshold to run too early
}

// Clone clones *Threshold
func (t *Threshold) Clone() *Threshold {
	if t == nil {
		return nil
	}
	clone := &Threshold{
		Tenant: t.Tenant,
		ID:     t.ID,
		Hits:   t.Hits,
		Snooze: t.Snooze,
	}
	return clone
}

// CacheClone returns a clone of Threshold used by ltcache CacheCloner
func (t *Threshold) CacheClone() any {
	return t.Clone()
}

// TenantID returns the concatenated key between tenant and ID
func (t *Threshold) TenantID() string {
	return ConcatenatedKey(t.Tenant, t.ID)
}

// AsMapStringInterface converts Threshold struct to map[string]any
func (t *Threshold) AsMapStringInterface() map[string]any {
	if t == nil {
		return nil
	}
	return map[string]any{
		Tenant: t.Tenant,
		ID:     t.ID,
		Hits:   t.Hits,
		Snooze: t.Snooze,
	}
}

// ThresholdWithAPIOpts is used in replicatorV1 for dispatcher
type ThresholdWithAPIOpts struct {
	*Threshold
	APIOpts map[string]any
}

// MapStringInterfaceToThreshold converts map[string]any to Threshold struct
func MapStringInterfaceToThreshold(m map[string]any) (*Threshold, error) {
	th := &Threshold{}
	if v, ok := m[Tenant].(string); ok {
		th.Tenant = v
	}
	if v, ok := m[ID].(string); ok {
		th.ID = v
	}
	if v, ok := m[Hits].(float64); ok {
		th.Hits = int(v)
	}
	if v, ok := m[Snooze].(string); ok {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			th.Snooze = t
		}
	}
	return th, nil
}

// ThresholdLockKey returns the ID used to lock a threshold with guardian.
func ThresholdLockKey(tnt, id string) string {
	return ConcatenatedKey(CacheThresholds, tnt, id)
}
