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
	"strings"
)

// CGRReplier is the interface supported by replies convertible to CGRReply
type CGRReplier interface {
	AsCGRReply() (CGRReply, error)
}

func NewCGRReply(rply CGRReplier, errRply error) (cgrReply CGRReply, err error) {
	if errRply != nil {
		return CGRReply{Error: errRply.Error()}, nil
	}
	if cgrReply, err = rply.AsCGRReply(); err != nil {
		return
	}
	cgrReply[Error] = "" // enforce empty error
	return
}

// CGRReply represents the CGRateS answer which can be used in templates
// it can be layered, case when interface{} will be castable into map[string]interface{}
type CGRReply map[string]interface{}

// GetFieldAsString returns the field value as string for the path specified
func (cgrReply CGRReply) GetFieldAsString(fldPath string, sep string) (fldVal string, err error) {
	path := strings.Split(fldPath, sep)
	lenPath := len(path)
	if lenPath == 0 {
		return "", errors.New("empty field path")
	}
	if path[0] == MetaCGRReply {
		path = path[1:]
		lenPath -= 1
	}
	lastMp := cgrReply // last map when layered
	var canCast bool
	for i, spath := range path {
		if i == lenPath-1 { // lastElement
			fldValIf, has := lastMp[spath]
			if !has {
				return "", fmt.Errorf("no field with path: <%s>", fldPath)
			}

			if fldVal, canCast = CastFieldIfToString(fldValIf); !canCast {
				return "", fmt.Errorf("cannot cast field: %s to string", ToJSON(fldValIf))
			}
			return
		} else {
			lastMp, canCast = lastMp[spath].(map[string]interface{})
			if !canCast {
				return "", fmt.Errorf("cannot cast field: %s to map[string]interface{}", ToJSON(lastMp[spath]))
			}
		}
	}
	return "", errors.New("end of function")
}
