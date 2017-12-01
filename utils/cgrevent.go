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
	"time"
)

// CGREvent is a generic event processed by CGR services
type CGREvent struct {
	Tenant string
	ID     string
	Event  map[string]interface{}
}

func (ev *CGREvent) CheckMandatoryFields(fldNames []string) error {
	for _, fldName := range fldNames {
		if _, has := ev.Event[fldName]; !has {
			return NewErrMandatoryIeMissing(fldName)
		}
	}
	return nil
}

// AnswerTime returns the AnswerTime of CGREvent
func (ev *CGREvent) FieldAsString(fldName string) (val string, err error) {
	iface, has := ev.Event[fldName]
	if !has {
		return "", ErrNotFound
	}
	val, canCast := CastFieldIfToString(iface)
	if !canCast {
		return "", fmt.Errorf("cannot cast %s to string", fldName)
	}
	return val, nil
}

// FieldAsTime returns the a field as Time instance
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

// FieldAsTime returns the a field as Time instance
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
	s, canCast := iface.(string)
	if !canCast {
		err = fmt.Errorf("cannot cast %s to string", fldName)
		return
	}
	return ParseDurationWithNanosecs(s)
}
