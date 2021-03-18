/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOev.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>
*/

package utils

import (
	"time"
)

// CGREvent is a generic event processed by CGR services
type CGREvent struct {
	Tenant string
	ID     string
	Time   *time.Time // event time
	Event  map[string]interface{}
	Opts   map[string]interface{}

	cache map[string]interface{}
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
func (ev *CGREvent) OptAsString(optName string) (val string, err error) {
	iface, has := ev.Opts[optName]
	if !has {
		return "", ErrNotFound
	}
	return IfaceAsString(iface), nil
}

// OptAsInt64 returns an option as int64
func (ev *CGREvent) OptAsInt64(optName string) (int64, error) {
	iface, has := ev.Opts[optName]
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

// OptAsDuration returns an option as Duration instance
func (ev *CGREvent) OptAsDuration(optName string) (d time.Duration, err error) {
	iface, has := ev.Event[optName]
	if !has {
		err = ErrNotFound
		return
	}
	return IfaceAsDuration(iface)
}

// FieldAsFloat64 returns a field as float64 instance
func (ev *CGREvent) FieldAsFloat64(fldName string) (f float64, err error) {
	iface, has := ev.Event[fldName]
	if !has {
		return f, ErrNotFound
	}
	return IfaceAsFloat64(iface)
}

// FieldAsInt64 returns a field as int64 instance
func (ev *CGREvent) FieldAsInt64(fldName string) (f int64, err error) {
	iface, has := ev.Event[fldName]
	if !has {
		return f, ErrNotFound
	}
	return IfaceAsInt64(iface)
}

func (ev *CGREvent) TenantID() string {
	return ConcatenatedKey(ev.Tenant, ev.ID)
}

// CacheInit will initialize the cache if not already done
func (ev *CGREvent) CacheInit() {
	if ev.cache == nil {
		ev.cache = make(map[string]interface{})
	}
}

// CacheClear will reset the cache
func (ev *CGREvent) CacheClear() {
	ev.cache = make(map[string]interface{})
}

// CacheGet will return a key from the cache
func (ev *CGREvent) CacheGet(key string) (itm interface{}, has bool) {
	itm, has = ev.cache[key]
	return
}

// CacheSet will set data into the event's cache
func (ev *CGREvent) CacheSet(key string, val interface{}) {
	ev.cache[key] = val
}

// CacheRemove will remove data from cache
func (ev *CGREvent) CacheRemove(key string) {
	delete(ev.cache, key)
}

func (ev *CGREvent) Clone() (clned *CGREvent) {
	clned = &CGREvent{
		Tenant: ev.Tenant,
		ID:     ev.ID,
		Event:  make(map[string]interface{}), // a bit forced but safe
		Opts:   make(map[string]interface{}),
	}
	if ev.Time != nil {
		clned.Time = TimePointer(*ev.Time)
	}
	for k, v := range ev.Event {
		clned.Event[k] = v
	}
	if ev.Opts != nil {
		for opt, val := range ev.Opts {
			clned.Opts[opt] = val
		}
	}
	return
}

// AsDataProvider returns the CGREvent as MapStorage with *opts and *req paths set
func (cgrEv *CGREvent) AsDataProvider() (ev DataProvider) {
	return MapStorage{
		MetaOpts: cgrEv.Opts,
		MetaReq:  cgrEv.Event,
	}
}

// CGREvents is a group of generic events processed by CGR services
// ie: derived CDRs
type CGREvents struct {
	Tenant string
	ID     string
	Time   *time.Time // event time
	Events []map[string]interface{}
}

// EventWithFlags is used where flags are needed to mark processing
type EventWithFlags struct {
	Flags []string
	Event map[string]interface{}
}

// GetRoutePaginatorFromOpts will consume supplierPaginator if present
func GetRoutePaginatorFromOpts(ev map[string]interface{}) (args Paginator, err error) {
	if ev == nil {
		return
	}
	//check if we have suppliersLimit in event and in case it has add it in args
	limitIface, hasRoutesLimit := ev[OptsRoutesLimit]
	if hasRoutesLimit {
		delete(ev, OptsRoutesLimit)
		var limit int64
		if limit, err = IfaceAsInt64(limitIface); err != nil {
			return
		}
		args = Paginator{
			Limit: IntPointer(int(limit)),
		}
	}
	//check if we have offset in event and in case it has add it in args
	offsetIface, hasRoutesOffset := ev[OptsRoutesOffset]
	if !hasRoutesOffset {
		return
	}
	delete(ev, OptsRoutesOffset)
	var offset int64
	if offset, err = IfaceAsInt64(offsetIface); err != nil {
		return
	}
	if !hasRoutesLimit { //in case we don't have limit, but we have offset we need to initialize the struct
		args = Paginator{
			Offset: IntPointer(int(offset)),
		}
		return
	}
	args.Offset = IntPointer(int(offset))
	return
}

// CGREventWithEeIDs is the CGREventWithOpts with EventExporterIDs
type CGREventWithEeIDs struct {
	EeIDs []string
	*CGREvent
}
