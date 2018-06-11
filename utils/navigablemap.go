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
	"errors"
	"fmt"
)

// CGRReplier is the interface supported by replies convertible to CGRReply
type NavigableMapper interface {
	AsNavigableMap() (map[string]interface{}, error)
}

// NavigableMap is a map who's values can be navigated via path
type NavigableMap map[string]interface{}

// FieldAsInterface returns the field value as interface{} for the path specified
// implements DataProvider
func (nM NavigableMap) FieldAsInterface(fldPath []string) (fldVal interface{}, err error) {
	lenPath := len(fldPath)
	if lenPath == 0 {
		return nil, errors.New("empty field path")
	}
	lastMp := nM // last map when layered
	var canCast bool
	for i, spath := range fldPath {
		if i == lenPath-1 { // lastElement
			var has bool
			fldVal, has = lastMp[spath]
			if !has {
				return nil, ErrNotFound
			}
			return
		} else {
			elmnt, has := lastMp[spath]
			if !has {
				err = fmt.Errorf("no map at path: <%s>", spath)
				return
			}
			lastMp, canCast = elmnt.(map[string]interface{})
			if !canCast {
				err = fmt.Errorf("cannot cast field: %s to map[string]interface{}",
					ToJSON(elmnt))
				return
			}
		}
	}
	err = errors.New("end of function")
	return
}

// FieldAsString returns the field value as string for the path specified
// implements DataProvider
func (nM NavigableMap) FieldAsString(fldPath []string) (fldVal string, err error) {
	var valIface interface{}
	valIface, err = nM.FieldAsInterface(fldPath)
	if err != nil {
		return
	}
	var canCast bool
	if fldVal, canCast = CastFieldIfToString(valIface); !canCast {
		return "", fmt.Errorf("cannot cast field: %s to string", ToJSON(valIface))
	}
	return
}

func (nM NavigableMap) String() string {
	return ToJSON(nM)
}

// NewCGRReply is specific to replies coming from CGRateS
func NewCGRReply(rply NavigableMapper, errRply error) (nM map[string]interface{}, err error) {
	if errRply != nil {
		return NavigableMap{Error: errRply.Error()}, nil
	}
	if nM, err = rply.AsNavigableMap(); err != nil {
		return
	}
	nM[Error] = "" // enforce empty error
	return
}
