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
	"fmt"
	"net"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

// NewMapEvent makes sure the content is not nil
func NewMapEvent(mp map[string]interface{}) (me MapEvent) {
	if mp == nil {
		mp = make(map[string]interface{})
	}
	return MapEvent(mp)
}

// MapEvent is a map[string]interface{} with convenience methods on top
type MapEvent map[string]interface{}

func (me MapEvent) String() string {
	return utils.ToJSON(me)
}

func (me MapEvent) FieldAsInterface(fldPath []string) (interface{}, error) {
	if len(fldPath) != 1 {
		return nil, utils.ErrNotFound
	}
	fldIface, has := me[fldPath[0]]
	if !has {
		return nil, utils.ErrNotFound
	}
	return fldIface, nil
}

func (me MapEvent) FieldAsString(fldPath []string) (string, error) {
	if len(fldPath) != 1 {
		return "", utils.ErrNotFound
	}
	return me.GetString(fldPath[0])
}

func (me MapEvent) RemoteHost() net.Addr {
	return utils.LocalAddr()
}

func (me MapEvent) HasField(fldName string) (has bool) {
	_, has = me[fldName]
	return
}

func (me MapEvent) GetString(fldName string) (out string, err error) {
	fldIface, has := me[fldName]
	if !has {
		return "", utils.ErrNotFound
	}
	return utils.IfaceAsString(fldIface), nil
}

func (me MapEvent) GetTInt64(fldName string) (out int64, err error) {
	fldIface, has := me[fldName]
	if !has {
		return 0, utils.ErrNotFound
	}
	return utils.IfaceAsTInt64(fldIface)
}

// GetFloat64 returns a field as float64 instance
func (me MapEvent) GetFloat64(fldName string) (f float64, err error) {
	iface, has := me[fldName]
	if !has {
		return f, utils.ErrNotFound
	}
	return utils.IfaceAsFloat64(iface)
}

func (me MapEvent) GetStringIgnoreErrors(fldName string) (out string) {
	out, _ = me.GetString(fldName)
	return
}

// GetDuration returns a field as Duration
func (me MapEvent) GetDuration(fldName string) (d time.Duration, err error) {
	fldIface, has := me[fldName]
	if !has {
		return d, utils.ErrNotFound
	}
	return utils.IfaceAsDuration(fldIface)
}

// GetDuration returns a field as Duration, ignoring errors
func (me MapEvent) GetDurationIgnoreErrors(fldName string) (d time.Duration) {
	d, _ = me.GetDuration(fldName)
	return
}

// GetDurationPointer returns pointer towards duration, useful to detect presence of duration
func (me MapEvent) GetDurationPtr(fldName string) (d *time.Duration, err error) {
	fldIface, has := me[fldName]
	if !has {
		return nil, utils.ErrNotFound
	}
	var dReal time.Duration
	if dReal, err = utils.IfaceAsDuration(fldIface); err != nil {
		return
	}
	return &dReal, nil
}

// GetDurationPointer returns pointer towards duration, useful to detect presence of duration
func (me MapEvent) GetDurationPtrIgnoreErrors(fldName string) (d *time.Duration) {
	d, _ = me.GetDurationPtr(fldName)
	return
}

// GetDurationPtrOrDefault returns pointer or default if fldName is missing
func (me MapEvent) GetDurationPtrOrDefault(fldName string, dflt *time.Duration) (d *time.Duration, err error) {
	if d, err = me.GetDurationPtr(fldName); err == utils.ErrNotFound {
		d = dflt
		err = nil
	}
	return
}

// GetTime returns a field as Time
func (me MapEvent) GetTime(fldName string, timezone string) (t time.Time, err error) {
	fldIface, has := me[fldName]
	if !has {
		return t, utils.ErrNotFound
	}
	return utils.IfaceAsTime(fldIface, timezone)
}

// GetTimeIgnoreErrors returns a field as Time instance, ignoring errors
func (me MapEvent) GetTimeIgnoreErrors(fldName string, tmz string) (t time.Time) {
	t, _ = me.GetTime(fldName, tmz)
	return
}

// GetTimePtr returns a pointer towards time or error
func (me MapEvent) GetTimePtr(fldName, tmz string) (t *time.Time, err error) {
	var tm time.Time
	if tm, err = me.GetTime(fldName, tmz); err != nil {
		return
	}
	return utils.TimePointer(tm), nil
}

// GetTimePtrIgnoreErrors returns a pointer towards time or nil if errors
func (me MapEvent) GetTimePtrIgnoreErrors(fldName, tmz string) (t *time.Time) {
	t, _ = me.GetTimePtr(fldName, tmz)
	return
}

// Clone returns the cloned map
func (me MapEvent) Clone() (mp MapEvent) {
	if me == nil {
		return
	}
	mp = make(MapEvent, len(me))
	for k, v := range me {
		mp[k] = v
	}
	return
}

// AsMapString returns a map[string]string out of mp, ignoring specific fields if needed
// most used when needing to export extraFields
func (me MapEvent) AsMapString(ignoredFlds utils.StringSet) (mp map[string]string) {
	mp = make(map[string]string)
	if ignoredFlds == nil {
		ignoredFlds = utils.NewStringSet(nil)
	}
	for k, v := range me {
		if ignoredFlds.Has(k) {
			continue
		}
		mp[k] = utils.IfaceAsString(v)
	}
	return
}

// AsCDR exports the MapEvent as CDR
func (me MapEvent) AsCDR(cfg *config.CGRConfig, tnt, tmz string) (cdr *CDR, err error) {
	cdr = &CDR{Tenant: tnt, Cost: -1.0, ExtraFields: make(map[string]string)}
	for k, v := range me {
		if !utils.MainCDRFields.Has(k) { // not primary field, populate extra ones
			cdr.ExtraFields[k] = utils.IfaceAsString(v)
			continue
		}
		switch k {
		default:
			// for the momment this return can not be reached because we implemented a case for every MainCDRField
			return nil, fmt.Errorf("unimplemented CDR field: <%s>", k)
		case utils.CGRID:
			cdr.CGRID = utils.IfaceAsString(v)
		case utils.RunID:
			cdr.RunID = utils.IfaceAsString(v)
		case utils.OriginHost:
			cdr.OriginHost = utils.IfaceAsString(v)
		case utils.Source:
			cdr.Source = utils.IfaceAsString(v)
		case utils.OriginID:
			cdr.OriginID = utils.IfaceAsString(v)
		case utils.ToR:
			cdr.ToR = utils.IfaceAsString(v)
		case utils.RequestType:
			cdr.RequestType = utils.IfaceAsString(v)
		case utils.Tenant:
			cdr.Tenant = utils.IfaceAsString(v)
		case utils.Category:
			cdr.Category = utils.IfaceAsString(v)
		case utils.AccountField:
			cdr.Account = utils.IfaceAsString(v)
		case utils.Subject:
			cdr.Subject = utils.IfaceAsString(v)
		case utils.Destination:
			cdr.Destination = utils.IfaceAsString(v)
		case utils.SetupTime:
			if cdr.SetupTime, err = utils.IfaceAsTime(v, tmz); err != nil {
				return nil, err
			}
		case utils.AnswerTime:
			if cdr.AnswerTime, err = utils.IfaceAsTime(v, tmz); err != nil {
				return nil, err
			}
		case utils.Usage:
			if cdr.Usage, err = utils.IfaceAsDuration(v); err != nil {
				return nil, err
			}
		case utils.Partial:
			if cdr.Partial, err = utils.IfaceAsBool(v); err != nil {
				return nil, err
			}
		case utils.PreRated:
			if cdr.PreRated, err = utils.IfaceAsBool(v); err != nil {
				return nil, err
			}
		case utils.CostSource:
			cdr.CostSource = utils.IfaceAsString(v)
		case utils.Cost:
			if cdr.Cost, err = utils.IfaceAsFloat64(v); err != nil {
				return nil, err
			}
		case utils.CostDetails:
			if cdr.CostDetails, err = IfaceAsEventCost(v); err != nil {
				return nil, err
			}
			cdr.CostDetails.initCache()
		case utils.ExtraInfo:
			cdr.ExtraInfo = utils.IfaceAsString(v)
		case utils.OrderID:
			if cdr.OrderID, err = utils.IfaceAsTInt64(v); err != nil {
				return nil, err
			}
		}
	}
	if cfg != nil {
		cdr.AddDefaults(cfg)
	}
	return
}

// Data returns the MapEvent as a map[string]interface{}
func (me MapEvent) Data() map[string]interface{} {
	return me
}

// GetBoolOrDefault returns the value as a bool or dflt if not present in map
func (me MapEvent) GetBoolOrDefault(fldName string, dflt bool) (out bool) {
	fldIface, has := me[fldName]
	if !has {
		return dflt
	}
	out, err := utils.IfaceAsBool(fldIface)
	if err != nil {
		return dflt
	}
	return out
}
