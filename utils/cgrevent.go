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

/*
func (ev *CGREvent) FilterableEvent(fltredFields []string) (fEv map[string]interface{}) {
	fEv = make(map[string]interface{})
	if len(fltredFields) == 0 {
		i := 0
		fltredFields = make([]string, len(ev.Event))
		for k := range ev.Event {
			fltredFields[i] = k
			i++
		}
	}
	for _, fltrFld := range fltredFields {
		fldVal, has := ev.Event[fltrFld]
		if !has {
			continue // the field does not exist in map, ignore it
		}
		valOf := reflect.ValueOf(fldVal)
		if valOf.Kind() == reflect.String {
			fEv[fltrFld] = StringToInterface(valOf.String()) // attempt converting from string to comparable interface
		} else {
			fEv[fltrFld] = fldVal
		}
	}
	return
}
*/

func (ev *CGREvent) Clone() (clned *CGREvent) {
	clned = &CGREvent{
		Tenant: ev.Tenant,
		ID:     ev.ID,
		Event:  make(map[string]interface{}), // a bit forced but safe
	}
	if ev.Time != nil {
		clned.Time = TimePointer(*ev.Time)
	}
	for k, v := range ev.Event {
		clned.Event[k] = v
	}
	return
}

// CGREvents is a group of generic events processed by CGR services
// ie: derived CDRs
type CGREvents struct {
	Tenant string
	ID     string
	Time   *time.Time // event time
	Events []map[string]interface{}
}

// CGREventWithOpts is the event with Opts
type CGREventWithOpts struct {
	Opts map[string]interface{}
	*CGREvent

	cache map[string]interface{}
}

// Clone return a copy of the CGREventWithOpts
func (ev *CGREventWithOpts) Clone() (clned *CGREventWithOpts) {
	if ev == nil {
		return
	}
	clned = new(CGREventWithOpts)
	if ev.CGREvent != nil {
		clned.CGREvent = ev.CGREvent.Clone()
	}
	if ev.Opts != nil {
		clned.Opts = make(map[string]interface{})
		for opt, val := range ev.Opts {
			clned.Opts[opt] = val
		}
	}
	return
}

// CacheInit will initialize the cache if not already done
func (ev *CGREventWithOpts) CacheInit() {
	if ev.cache == nil {
		ev.cache = make(map[string]interface{})
	}
}

// CacheClear will reset the cache
func (ev *CGREventWithOpts) CacheClear() {
	ev.cache = make(map[string]interface{})
}

// CacheGet will return a key from the cache
func (ev *CGREventWithOpts) CacheGet(key string) (itm interface{}, has bool) {
	itm, has = ev.cache[key]
	return
}

// CacheSet will set data into the event's cache
func (ev *CGREventWithOpts) CacheSet(key string, val interface{}) {
	ev.cache[key] = val
}

// CacheRemove will remove data from cache
func (ev *CGREventWithOpts) CacheRemove(key string) {
	delete(ev.cache, key)
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
	*CGREventWithOpts
}
