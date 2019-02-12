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
	"fmt"
	"reflect"
	"strconv"
	"time"
)

// CGREvent is a generic event processed by CGR services
type CGREvent struct {
	Tenant  string
	ID      string
	Context *string    // attach the event to a context
	Time    *time.Time // event time
	Event   map[string]interface{}
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

// AnswerTime returns a field as string instance
func (ev *CGREvent) FieldAsString(fldName string) (val string, err error) {
	iface, has := ev.Event[fldName]
	if !has {
		return "", ErrNotFound
	}
	val, err = IfaceAsString(iface)
	if err != nil {
		return "", fmt.Errorf("cannot cast %s to string", fldName)
	}
	return val, nil
}

// FieldAsTime returns a field as Time instance
func (ev *CGREvent) FieldAsTime(fldName string, timezone string) (t time.Time, err error) {
	iface, has := ev.Event[fldName]
	if !has {
		err = ErrNotFound
		return
	}
	var canCast bool
	if t, canCast = iface.(time.Time); canCast {
		return
	}
	s, canCast := iface.(string)
	if !canCast {
		err = fmt.Errorf("cannot cast %s to string", fldName)
		return
	}
	return ParseTimeDetectLayout(s, timezone)
}

// FieldAsDuration returns a field as Duration instance
func (ev *CGREvent) FieldAsDuration(fldName string) (d time.Duration, err error) {
	iface, has := ev.Event[fldName]
	if !has {
		err = ErrNotFound
		return
	}
	var canCast bool
	if d, canCast = iface.(time.Duration); canCast {
		return
	}
	if f, canCast := iface.(float64); canCast {
		return time.Duration(int64(f)), nil
	}
	s, canCast := iface.(string)
	if !canCast {
		err = fmt.Errorf("cannot cast %s to string", fldName)
		return
	}
	return ParseDurationWithNanosecs(s)
}

// FieldAsFloat64 returns a field as float64 instance
func (ev *CGREvent) FieldAsFloat64(fldName string) (f float64, err error) {
	iface, has := ev.Event[fldName]
	if !has {
		return f, ErrNotFound
	}
	if val, canCast := iface.(float64); canCast {
		return val, nil
	}
	csStr, canCast := iface.(string)
	if !canCast {
		err = fmt.Errorf("cannot cast %s to string", fldName)
		return
	}
	return strconv.ParseFloat(csStr, 64)
}

func (ev *CGREvent) TenantID() string {
	return ConcatenatedKey(ev.Tenant, ev.ID)
}

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

func (ev *CGREvent) Clone() (clned *CGREvent) {
	clned = &CGREvent{
		Tenant: ev.Tenant,
		ID:     ev.ID,
		Event:  make(map[string]interface{}), // a bit forced but safe
	}
	if ev.Context != nil {
		clned.Context = StringPointer(*ev.Context)
	}
	if ev.Time != nil {
		clned.Time = TimePointer(*ev.Time)
	}
	for k, v := range ev.Event {
		clned.Event[k] = v
	}
	return
}
