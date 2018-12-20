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

func (me MapEvent) AsNavigableMap([]*config.FCTemplate) (*config.NavigableMap, error) {
	return config.NewNavigableMap(me), nil
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
	return utils.IfaceAsString(fldIface)
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

// Clone returns the cloned map
func (me MapEvent) Clone() (mp MapEvent) {
	mp = make(MapEvent, len(me))
	for k, v := range me {
		mp[k] = v
	}
	return
}

// AsMapString returns a map[string]string out of mp, ignoring specific fields if needed
// most used when needing to export extraFields
func (me MapEvent) AsMapString(ignoredFlds utils.StringMap) (mp map[string]string, err error) {
	mp = make(map[string]string)
	for k, v := range me {
		if ignoredFlds.HasKey(k) {
			continue
		}
		var out string
		if out, err = utils.IfaceAsString(v); err != nil {
			return nil, err
		}
		mp[k] = out
	}
	return
}

func (me MapEvent) AsMapStringIgnoreErrors(ignoredFlds utils.StringMap) (mp map[string]string) {
	mp = make(map[string]string)
	for k, v := range me {
		if ignoredFlds.HasKey(k) {
			continue
		}
		if out, err := utils.IfaceAsString(v); err == nil {
			mp[k] = out
		}
	}
	return
}

// AsCDR exports the SafEvent as CDR
func (me MapEvent) AsCDR(cfg *config.CGRConfig, tnt, tmz string) (cdr *CDR, err error) {
	cdr = &CDR{Tenant: tnt, Cost: -1.0, ExtraFields: make(map[string]string)}
	for k, v := range me {
		if !utils.IsSliceMember(utils.NotExtraCDRFields, k) { // not primary field, populate extra ones
			utils.Logger.Debug(fmt.Sprintf("field <%s> as extra since is not present in %s", k, utils.ToJSON(utils.NotExtraCDRFields)))
			if cdr.ExtraFields[k], err = utils.IfaceAsString(v); err != nil {
				return nil, err
			}
			continue
		}
		switch k {
		default:
			return nil, fmt.Errorf("unimplemented CDR field: <%s>", k)
		case utils.CGRID:
			if cdr.CGRID, err = utils.IfaceAsString(v); err != nil {
				return nil, err
			}
		case utils.RunID:
			if cdr.RunID, err = utils.IfaceAsString(v); err != nil {
				return nil, err
			}
		case utils.OriginHost:
			if cdr.OriginHost, err = utils.IfaceAsString(v); err != nil {
				return nil, err
			}
		case utils.Source:
			if cdr.Source, err = utils.IfaceAsString(v); err != nil {
				return nil, err
			}
		case utils.OriginID:
			if cdr.OriginID, err = utils.IfaceAsString(v); err != nil {
				return nil, err
			}
		case utils.ToR:
			if cdr.ToR, err = utils.IfaceAsString(v); err != nil {
				return nil, err
			}
		case utils.RequestType:
			if cdr.RequestType, err = utils.IfaceAsString(v); err != nil {
				return nil, err
			}
		case utils.Tenant:
			if cdr.Tenant, err = utils.IfaceAsString(v); err != nil {
				return nil, err
			}
		case utils.Category:
			if cdr.Category, err = utils.IfaceAsString(v); err != nil {
				return nil, err
			}
		case utils.Account:
			if cdr.Account, err = utils.IfaceAsString(v); err != nil {
				return nil, err
			}
		case utils.Subject:
			if cdr.Subject, err = utils.IfaceAsString(v); err != nil {
				return nil, err
			}
		case utils.Destination:
			if cdr.Destination, err = utils.IfaceAsString(v); err != nil {
				return nil, err
			}
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
			if cdr.CostSource, err = utils.IfaceAsString(v); err != nil {
				return nil, err
			}
		case utils.Cost:
			if cdr.Cost, err = utils.IfaceAsFloat64(v); err != nil {
				return nil, err
			}
		}
	}
	if cfg != nil {
		cdr.AddDefaults(cfg)
	}
	return
}
