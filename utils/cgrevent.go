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
	"strings"
	"time"

	"github.com/ericlagergren/decimal"
)

// CGREvent is a generic event processed by CGR services
type CGREvent struct {
	Tenant  string
	ID      string
	Event   map[string]interface{}
	APIOpts map[string]interface{}
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

// OptsAsDecimal OptAsDecimal returns an option as decimal.Big instance
func (ev *CGREvent) OptsAsDecimal(optName string) (d *decimal.Big, err error) {
	iface, has := ev.APIOpts[optName]
	if !has {
		err = ErrNotFound
		return
	}
	return IfaceAsBig(iface)
}

func (ev *CGREvent) Clone() (clned *CGREvent) {
	clned = &CGREvent{
		Tenant:  ev.Tenant,
		ID:      ev.ID,
		Event:   make(map[string]interface{}), // a bit forced but safe
		APIOpts: make(map[string]interface{}),
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

// AsDataProvider returns the CGREvent as MapStorage with *opts and *req paths set
func (cgrEv *CGREvent) AsDataProvider() (ev DataProvider) {
	return MapStorage{
		MetaOpts: cgrEv.APIOpts,
		MetaReq:  cgrEv.Event,
	}
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

// NMAsCGREvent builds a CGREvent considering Time as time.Now()
// and Event as linear map[string]interface{} with joined paths
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
		Event:   make(map[string]interface{}),
		APIOpts: opts,
	}
	for ; el != nil; el = el.Next() {
		path := el.Value
		val, _ := nM.Field(path) // this should never return error cause we get the path from the order
		if val.AttributeID != "" {
			continue
		}
		path = path[:len(path)-1] // remove the last index
		opath := strings.Join(path, NestingSep)
		if _, has := cgrEv.Event[opath]; !has {
			cgrEv.Event[opath] = val.Data // first item which is not an attribute will become the value
		}
	}
	return
}

// StartTime returns the event time used to check active rate profiles
func (args *CGREvent) StartTime(configSTime, tmz string) (time.Time, error) {
	if tIface, has := args.APIOpts[OptsRatesStartTime]; has {
		return IfaceAsTime(tIface, tmz)
	}
	if tIface, has := args.APIOpts[MetaStartTime]; has {
		return IfaceAsTime(tIface, tmz)
	}
	return ParseTimeDetectLayout(configSTime, tmz)
}

// usage returns the event time used to check active rate profiles
func (args *CGREvent) Usage(configUsage string) (usage *decimal.Big, err error) {
	// first search for the rateUsage in opts
	if uIface, has := args.APIOpts[OptsRatesUsage]; has {
		return IfaceAsBig(uIface)
	}
	// second search for the usage in opts
	if uIface, has := args.APIOpts[MetaUsage]; has {
		return IfaceAsBig(uIface)
	}
	// if the usage is not found in the event populate with default value and overwrite the NOT_FOUND error with nil
	return StringAsBig(configUsage)
}

// IntervalStart returns the inerval start out of APIOpts received for the event
func (args *CGREvent) IntervalStart(configIvlStart string) (ivlStart *decimal.Big, err error) {
	if iface, has := args.APIOpts[OptsRatesIntervalStart]; has {
		return IfaceAsBig(iface)
	}
	return StringAsBig(configIvlStart)
}
