// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package utils

import (
	"strings"
	"time"
)

// CGREvent is a generic event processed by CGR services
type CGREvent struct {
	Tenant  string
	ID      string
	Event   map[string]any
	APIOpts map[string]any
	clnb    bool //rpcclonable
}

func (ev *CGREvent) HasField(fldName string) (has bool) {
	_, has = ev.Event[fldName]
	return
}

func (ev *CGREvent) CheckMandatoryFields(fldNames []string) error {
	for _, fldName := range fldNames {
		if _, has := ev.Event[fldName]; !has {
			return NewErrMandatoryIeMissing(fldName)
		}
	}
	return nil
}

// FieldAsString returns a field as string instance
func (ev *CGREvent) FieldAsString(fldName string) (val string, err error) {
	iface, has := ev.Event[fldName]
	if !has {
		return "", ErrNotFound
	}
	return IfaceAsString(iface), nil
}

// OptAsString returns an option as string
func (ev *CGREvent) OptAsString(optName string) (string, error) {
	v, exists := ev.APIOpts[optName]
	if !exists {
		return "", ErrNotFound
	}
	return IfaceAsString(v), nil
}

// OptAsInt64 returns an option as int64
func (ev *CGREvent) OptAsInt64(optName string) (int64, error) {
	iface, has := ev.APIOpts[optName]
	if !has {
		return 0, ErrNotFound
	}
	return IfaceAsTInt64(iface)
}

// FieldAsTime returns a field as Time instance
func (ev *CGREvent) FieldAsTime(fldName string, timezone string) (t time.Time, err error) {
	iface, has := ev.Event[fldName]
	if !has {
		err = ErrNotFound
		return
	}
	return IfaceAsTime(iface, timezone)
}

// FieldAsDuration returns a field as Duration instance
func (ev *CGREvent) FieldAsDuration(fldName string) (d time.Duration, err error) {
	iface, has := ev.Event[fldName]
	if !has {
		err = ErrNotFound
		return
	}
	return IfaceAsDuration(iface)
}

func (ev *CGREvent) Clone() (clned *CGREvent) {
	clned = &CGREvent{
		Tenant:  ev.Tenant,
		ID:      ev.ID,
		Event:   make(map[string]any),
		APIOpts: make(map[string]any),
	}
	for k, v := range ev.Event {
		clned.Event[k] = v
	}
	if ev.APIOpts != nil {
		for opt, val := range ev.APIOpts {
			clned.APIOpts[opt] = val
		}
	}
	return
}

// CacheClone returns a clone of CGREvent used by ltcache CacheCloner
func (ev *CGREvent) CacheClone() any {
	return ev.Clone()
}

// AsDataProvider returns the CGREvent as MapStorage with *opts and *req paths set
func (cgrEv *CGREvent) AsDataProvider() (ev MapStorage) {
	return MapStorage{
		MetaOpts: cgrEv.APIOpts,
		MetaReq:  cgrEv.Event,
	}
}

// CGREventWithRateProfile is used to get the rates prom a specific RatePRofileID that is matching our Event
type CGREventWithRateProfile struct {
	RateProfileID string
	*CGREvent
}

type EventsWithOpts struct {
	Event map[string]any
	Opts  map[string]any
}

// CGREventWithEeIDs is the CGREventWithOpts with EventExporterIDs
type CGREventWithEeIDs struct {
	EeIDs []string
	*CGREvent
}

func (CGREventWithEeIDs) RPCClone() {} // disable rpcClonable from CGREvent

// NMAsCGREvent builds a CGREvent considering Time as time.Now()
// and Event as linear map[string]any with joined paths
// treats particular case when the value of map is []*NMItem - used in agents/AgentRequest
func NMAsCGREvent(nM *OrderedNavigableMap, tnt string, pathSep string, opts MapStorage) (cgrEv *CGREvent) {
	if nM == nil {
		return
	}
	el := nM.GetFirstElement()
	if el == nil {
		return
	}

	cgrEv = &CGREvent{
		Tenant:  tnt,
		ID:      UUIDSha1Prefix(),
		Event:   make(map[string]any),
		APIOpts: opts,
	}
	for ; el != nil; el = el.Next() {
		path := el.Value
		val, _ := nM.Field(path) // this should never return error cause we get the path from the order
		if val.AttributeID != "" {
			continue
		}
		path = StripTrailingIndex(path)
		opath := strings.Join(path, NestingSep)
		if _, has := cgrEv.Event[opath]; !has {
			cgrEv.Event[opath] = val.Data // first item which is not an attribute will become the value
		}
	}
	return
}

// SetCloneable sets if the args should be clonned on internal connections
func (attr *CGREvent) SetCloneable(rpcCloneable bool) {
	attr.clnb = rpcCloneable
}

// RPCClone implements rpcclient.RPCCloner interface
func (attr *CGREvent) RPCClone() (any, error) {
	if !attr.clnb {
		return attr, nil
	}
	return attr.Clone(), nil
}
