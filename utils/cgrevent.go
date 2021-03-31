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

func (ev *CGREvent) consumeArgDispatcher() (arg *ArgDispatcher) {
	if ev == nil {
		return
	}
	//check if we have APIKey in event and in case it has add it in ArgDispatcher
	apiKeyIface, hasApiKey := ev.Event[MetaApiKey]
	if hasApiKey {
		delete(ev.Event, MetaApiKey)
		arg = &ArgDispatcher{
			APIKey: StringPointer(apiKeyIface.(string)),
		}
	}
	//check if we have RouteID in event and in case it has add it in ArgDispatcher
	routeIDIface, hasRouteID := ev.Event[MetaRouteID]
	if !hasRouteID {
		return
	}
	delete(ev.Event, MetaRouteID)
	if !hasApiKey { //in case we don't have APIKey, but we have RouteID we need to initialize the struct
		return &ArgDispatcher{
			RouteID: StringPointer(routeIDIface.(string)),
		}
	}
	arg.RouteID = StringPointer(routeIDIface.(string))
	return
}

// ConsumeSupplierPaginator will consume supplierPaginator if presented
func (ev *CGREvent) consumeSupplierPaginator() (args *Paginator) {
	args = new(Paginator)
	if ev == nil {
		return
	}
	//check if we have suppliersLimit in event and in case it has add it in args
	limitIface, hasSuppliersLimit := ev.Event[MetaSuppliersLimit]
	if hasSuppliersLimit {
		delete(ev.Event, MetaSuppliersLimit)
		limit, err := IfaceAsInt64(limitIface)
		if err != nil {
			Logger.Err(err.Error())
			return
		}
		args = &Paginator{
			Limit: IntPointer(int(limit)),
		}
	}
	//check if we have offset in event and in case it has add it in args
	offsetIface, hasSuppliersOffset := ev.Event[MetaSuppliersOffset]
	if hasSuppliersOffset {
		delete(ev.Event, MetaSuppliersOffset)
		offset, err := IfaceAsInt64(offsetIface)
		if err != nil {
			Logger.Err(err.Error())
			return
		}
		if !hasSuppliersLimit { //in case we don't have limit, but we have offset we need to initialize the struct
			args = &Paginator{
				Offset: IntPointer(int(offset)),
			}
		} else {
			args.Offset = IntPointer(int(offset))
		}
	}
	return
}

// ExtractedArgs stores the extracted arguments from CGREvent
type ExtractedArgs struct {
	ArgDispatcher     *ArgDispatcher
	SupplierPaginator *Paginator
}

// ExtractArgs extracts the ArgDispatcher and SupplierPaginator from the received event
func (ev *CGREvent) ExtractArgs(dispatcherFlag, consumeSupplierPaginator bool) (ca ExtractedArgs) {
	ca = ExtractedArgs{
		ArgDispatcher: ev.consumeArgDispatcher(),
	}
	if dispatcherFlag && ca.ArgDispatcher == nil {
		ca.ArgDispatcher = new(ArgDispatcher)
	}
	if consumeSupplierPaginator {
		ca.SupplierPaginator = ev.consumeSupplierPaginator()
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

func NewCGREventWithArgDispatcher() *CGREventWithArgDispatcher {
	return &CGREventWithArgDispatcher{
		CGREvent:      new(CGREvent),
		ArgDispatcher: new(ArgDispatcher),
	}
}

type CGREventWithArgDispatcher struct {
	*CGREvent
	*ArgDispatcher
}

func (ev *CGREventWithArgDispatcher) Clone() (clned *CGREventWithArgDispatcher) {
	if ev == nil {
		return
	}
	clned = new(CGREventWithArgDispatcher)
	if ev.CGREvent != nil {
		clned.CGREvent = ev.CGREvent.Clone()
	}
	if ev.ArgDispatcher != nil {
		clned.ArgDispatcher = new(ArgDispatcher)
		*clned.ArgDispatcher = *ev.ArgDispatcher
	}
	return
}
