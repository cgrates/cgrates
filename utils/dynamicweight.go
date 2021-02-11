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

package utils

import (
	"fmt"
	"strconv"
	"strings"
)

// NewDynamicWeightsFromString creates a DynamicWeight list based on the string received from .csv/StorDB
func NewDynamicWeightsFromString(s, dWSep, fltrSep string) (dWs []*DynamicWeight, err error) {
	nrFlds := 2
	dwStrs := strings.Split(s, dWSep)
	lnDwStrs := len(dwStrs)
	if lnDwStrs%nrFlds != 0 { // need to have multiples of number of fields in one DynamicWeight
		return nil, fmt.Errorf("invalid DynamicWeight format for string <%s>", s)
	}
	dWs = make([]*DynamicWeight, 0, lnDwStrs/nrFlds)
	for i := 0; i < lnDwStrs; i += nrFlds {
		var fltrIDs []string
		if len(dwStrs[i]) != 0 {
			fltrIDs = strings.Split(dwStrs[i], fltrSep)
		}
		var weight float64
		if len(dwStrs[i+1]) != 0 {
			if weight, err = strconv.ParseFloat(dwStrs[i+1], 64); err != nil {
				return nil, fmt.Errorf("invalid Weight <%s> in string: <%s>", dwStrs[i+1], s)
			}
		}
		dWs = append(dWs, &DynamicWeight{FilterIDs: fltrIDs, Weight: weight})
	}
	return
}

// DynamicWeight returns Weight based on Filters
type DynamicWeight struct {
	FilterIDs []string
	Weight    float64
}
