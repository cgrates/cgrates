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
	"encoding/json"
	"strings"
	"time"

	"github.com/cgrates/cgrates/utils"
	"github.com/ericlagergren/decimal"
)

// CGREvent is a generic event processed by CGR services
type CGREvent struct {
	Tenant  string
	ID      string
	Time    *time.Time // event time
	Event   map[string]any
	APIOpts map[string]any
	clnb    bool //rpcclonable
}

// UnmarshalJSON ensures that CostDetails is of type *EventCost if it exists.
func (cgrEv *CGREvent) UnmarshalJSON(data []byte) (err error) {

	// Alias CGREvent to avoid recursion during json.Unmarshal.
	type Alias CGREvent
	aux := &struct{ *Alias }{Alias: (*Alias)(cgrEv)}

	// Use default unmarshaler to unmarshal the JSON data
	// into the auxiliary struct.
	if err = json.Unmarshal(data, aux); err != nil {
		return err
	}

	// Check if the Event map contains the "CostDetails" key,
	// which requires special handling.
	if ecEv, has := cgrEv.Event[utils.CostDetails]; has {
		var ecBytes []byte

		// CostDetails value can either be a JSON string (which is
		// the marshaled form of an EventCost) or a map representing
		// the EventCost directly.
		switch v := ecEv.(type) {
		case string:
			// If string, it's assumed to be the JSON
			// representation of EventCost.
			ecBytes = []byte(v)
		default:
			// Otherwise we assume it's a map and we marshal
			// it back to JSON to prepare for unmarshalling
			// into EventCost.
			ecBytes, err = json.Marshal(v)
			if err != nil {
				return err
			}
		}

		// Unmarshal the JSON (either directly from the string case
		// or from the marshaled map) into an EventCost struct.
		var ec EventCost
		if err := json.Unmarshal(ecBytes, &ec); err != nil {
			return err
		}

		// Update the Event map with the unmarshalled EventCost,
		// ensuring the type of CostDetails is *EventCost.
		cgrEv.Event[utils.CostDetails] = &ec
	}
	return nil
}

func (ev *CGREvent) HasField(fldName string) (has bool) {
	_, has = ev.Event[fldName]
	return
}

func (ev *CGREvent) CheckMandatoryFields(fldNames []string) error {
	for _, fldName := range fldNames {
		if _, has := ev.Event[fldName]; !has {
			return utils.NewErrMandatoryIeMissing(fldName)
		}
	}
	return nil
}

// FieldAsString returns a field as string instance
func (ev *CGREvent) FieldAsString(fldName string) (val string, err error) {
	iface, has := ev.Event[fldName]
	if !has {
		return "", utils.ErrNotFound
	}
	return utils.IfaceAsString(iface), nil
}

// OptAsString returns an option as string
func (ev *CGREvent) OptAsString(optName string) (val string, err error) {
	iface, has := ev.APIOpts[optName]
	if !has {
		return "", utils.ErrNotFound
	}
	return utils.IfaceAsString(iface), nil
}

// OptAsInt64 returns an option as int64
func (ev *CGREvent) OptAsInt64(optName string) (int64, error) {
	iface, has := ev.APIOpts[optName]
	if !has {
		return 0, utils.ErrNotFound
	}
	return utils.IfaceAsTInt64(iface)
}

// FieldAsTime returns a field as Time instance
func (ev *CGREvent) FieldAsTime(fldName string, timezone string) (t time.Time, err error) {
	iface, has := ev.Event[fldName]
	if !has {
		err = utils.ErrNotFound
		return
	}
	return utils.IfaceAsTime(iface, timezone)
}

// FieldAsDuration returns a field as Duration instance
func (ev *CGREvent) FieldAsDuration(fldName string) (d time.Duration, err error) {
	iface, has := ev.Event[fldName]
	if !has {
		err = utils.ErrNotFound
		return
	}
	return utils.IfaceAsDuration(iface)
}

// OptAsDuration returns an option as Duration instance
func (ev *CGREvent) OptAsDuration(optName string) (d time.Duration, err error) {
	iface, has := ev.APIOpts[optName]
	if !has {
		err = utils.ErrNotFound
		return
	}
	return utils.IfaceAsDuration(iface)
}

// FieldAsFloat64 returns a field as float64 instance
func (ev *CGREvent) FieldAsFloat64(fldName string) (f float64, err error) {
	iface, has := ev.Event[fldName]
	if !has {
		return f, utils.ErrNotFound
	}
	return utils.IfaceAsFloat64(iface)
}

// FieldAsInt64 returns a field as int64 instance
func (ev *CGREvent) FieldAsInt64(fldName string) (f int64, err error) {
	iface, has := ev.Event[fldName]
	if !has {
		return f, utils.ErrNotFound
	}
	return utils.IfaceAsInt64(iface)
}

func (ev *CGREvent) TenantID() string {
	return utils.ConcatenatedKey(ev.Tenant, ev.ID)
}

func (ev *CGREvent) Clone() (clned *CGREvent) {
	clned = &CGREvent{
		Tenant:  ev.Tenant,
		ID:      ev.ID,
		Event:   make(map[string]any),
		APIOpts: make(map[string]any),
	}
	if ev.Time != nil {
		clned.Time = new(time.Time)
		*clned.Time = *ev.Time
	}
	for k, v := range ev.Event {
		clned.Event[k] = v
	}
	for opt, val := range ev.APIOpts {
		clned.APIOpts[opt] = val
	}
	return
}

// AsDataProvider returns the CGREvent as MapStorage with *opts and *req paths set
func (cgrEv *CGREvent) AsDataProvider() (ev utils.DataProvider) {
	return utils.MapStorage{
		utils.MetaOpts: cgrEv.APIOpts,
		utils.MetaReq:  cgrEv.Event,
	}
}

// EventWithFlags is used where flags are needed to mark processing
type EventWithFlags struct {
	Flags []string
	Event map[string]any
}

// GetRoutePaginatorFromOpts will consume supplierPaginator if present
func GetRoutePaginatorFromOpts(ev map[string]any) (args utils.Paginator, err error) {
	if ev == nil {
		return
	}
	//check if we have suppliersLimit in event and in case it has add it in args
	limitIface, hasRoutesLimit := ev[utils.OptsRoutesLimit]
	if hasRoutesLimit {
		delete(ev, utils.OptsRoutesLimit)
		var limit int64
		if limit, err = utils.IfaceAsInt64(limitIface); err != nil {
			return
		}
		args = utils.Paginator{
			Limit: utils.IntPointer(int(limit)),
		}
	}
	//check if we have offset in event and in case it has add it in args
	offsetIface, hasRoutesOffset := ev[utils.OptsRoutesOffset]
	if !hasRoutesOffset {
		return
	}
	delete(ev, utils.OptsRoutesOffset)
	var offset int64
	if offset, err = utils.IfaceAsInt64(offsetIface); err != nil {
		return
	}
	if !hasRoutesLimit { //in case we don't have limit, but we have offset we need to initialize the struct
		args = utils.Paginator{
			Offset: utils.IntPointer(int(offset)),
		}
		return
	}
	args.Offset = utils.IntPointer(int(offset))
	return
}

// NMAsCGREvent builds a CGREvent considering Time as time.Now()
// and Event as linear map[string]any with joined paths
// treats particular case when the value of map is []*NMItem - used in agents/AgentRequest
func NMAsCGREvent(nM *utils.OrderedNavigableMap, tnt string, pathSep string, opts utils.MapStorage) (cgrEv *CGREvent) {
	if nM == nil {
		return
	}
	el := nM.GetFirstElement()
	if el == nil {
		return
	}
	cgrEv = &CGREvent{
		Tenant:  tnt,
		ID:      utils.UUIDSha1Prefix(),
		Time:    utils.TimePointer(time.Now()),
		Event:   make(map[string]any),
		APIOpts: opts,
	}
	for ; el != nil; el = el.Next() {
		path := el.Value
		val, _ := nM.Field(path) // this should never return error cause we get the path from the order
		if val.AttributeID != "" {
			continue
		}
		path = path[:len(path)-1] // remove the last index
		opath := strings.Join(path, utils.NestingSep)
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

// GetFloat64Opts checks the specified option names in order among the keys in APIOpts returning the first value it finds as float64, otherwise it
// returns the default option (usually the value specified in config)
func GetFloat64Opts(ev *CGREvent, dftOpt float64, optNames ...string) (cfgOpt float64, err error) {
	for _, optName := range optNames {
		if opt, has := ev.APIOpts[optName]; has {
			return utils.IfaceAsFloat64(opt)
		}
	}
	return dftOpt, nil
}

// GetDurationOpts checks the specified option names in order among the keys in APIOpts returning the first value it finds as time.Duration, otherwise it
// returns the default option (usually the value specified in config)
func GetDurationOpts(ev *CGREvent, dftOpt time.Duration, optNames ...string) (cfgOpt time.Duration, err error) {
	for _, optName := range optNames {
		if opt, has := ev.APIOpts[optName]; has {
			return utils.IfaceAsDuration(opt)
		}
	}
	return dftOpt, nil
}

// GetStringOpts checks the specified option names in order among the keys in APIOpts returning the first value it finds as string, otherwise it
// returns the default option (usually the value specified in config)
func GetStringOpts(ev *CGREvent, dftOpt string, optNames ...string) (cfgOpt string) {
	for _, optName := range optNames {
		if opt, has := ev.APIOpts[optName]; has {
			return utils.IfaceAsString(opt)
		}
	}
	return dftOpt
}

// GetStringSliceOpts checks the specified option names in order among the keys in APIOpts returning the first value it finds as []string, otherwise it
// returns the default option (usually the value specified in config)
func GetStringSliceOpts(ev *CGREvent, dftOpt []string, optNames ...string) (cfgOpt []string, err error) {
	for _, optName := range optNames {
		if opt, has := ev.APIOpts[optName]; has {
			return utils.IfaceAsSliceString(opt)
		}
	}
	return dftOpt, nil
}

// GetIntOpts checks the specified option names in order among the keys in APIOpts returning the first value it finds as int, otherwise it
// returns the default option (usually the value specified in config)
func GetIntOpts(ev *CGREvent, dftOpt int, optNames ...string) (cfgOpt int, err error) {
	for _, optName := range optNames {
		if opt, has := ev.APIOpts[optName]; has {
			var value int64
			if value, err = utils.IfaceAsTInt64(opt); err != nil {
				return
			}
			return int(value), nil
		}
	}
	return dftOpt, nil
}

// GetBoolOpts checks the specified option names in order among the keys in APIOpts returning the first value it finds as bool, otherwise it
// returns the default option (usually the value specified in config)
func GetBoolOpts(ev *CGREvent, dftOpt bool, optNames ...string) (cfgOpt bool, err error) {
	for _, optName := range optNames {
		if opt, has := ev.APIOpts[optName]; has {
			return utils.IfaceAsBool(opt)
		}
	}
	return dftOpt, nil
}

// GetDecimalBigOpts checks the specified option names in order among the keys in APIOpts returning the first value it finds as *decimal.Big, otherwise it
// returns the default option (usually the value specified in config)
func GetDecimalBigOpts(ev *CGREvent, dftOpt *decimal.Big, optNames ...string) (cfgOpt *decimal.Big, err error) {
	for _, optName := range optNames {
		if opt, has := ev.APIOpts[optName]; has {
			return utils.IfaceAsBig(opt)
		}
	}
	return dftOpt, nil
}

// GetInterfaceOpts checks the specified option names in order among the keys in APIOpts returning the first value it finds as any, otherwise it
// returns the default option (usually the value specified in config)
func GetInterfaceOpts(ev *CGREvent, dftOpt any, optNames ...string) (cfgOpt any) {
	for _, optName := range optNames {
		if opt, has := ev.APIOpts[optName]; has {
			return opt
		}
	}
	return dftOpt
}

// GetIntPointerOpts checks the specified option names in order among the keys in APIOpts returning the first value it finds as *int, otherwise it
// returns the default option (usually the value specified in config)
func GetIntPointerOpts(ev *CGREvent, dftOpt *int, optNames ...string) (cfgOpt *int, err error) {
	for _, optName := range optNames {
		if opt, has := ev.APIOpts[optName]; has {
			var value int64
			if value, err = utils.IfaceAsTInt64(opt); err != nil {
				return
			}
			return utils.IntPointer(int(value)), nil
		}
	}
	return dftOpt, nil
}

// GetDurationPointerOpts checks the specified option names in order among the keys in APIOpts returning the first value it finds as *time.Duration, otherwise it
// returns the default option (usually the value specified in config)
func GetDurationPointerOpts(ev *CGREvent, dftOpt *time.Duration, optNames ...string) (cfgOpt *time.Duration, err error) {
	for _, optName := range optNames {
		if opt, has := ev.APIOpts[optName]; has {
			var value time.Duration
			if value, err = utils.IfaceAsDuration(opt); err != nil {
				return
			}
			return utils.DurationPointer(value), nil
		}
	}
	return dftOpt, nil
}
