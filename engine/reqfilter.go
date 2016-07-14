/*
Real-time Charging System for Telecom & ISP environments
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
	"errors"
	"fmt"
	"strings"

	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

const (
	MetaStringPrefix = "*string_prefix"
	MetaTimings      = "*timings"
	MetaRSRFields    = "*rsr_fields"
	MetaCDRStats     = "*cdr_stats"
	MetaDestinations = "*destinations"
)

func NewRequestFilter(rfType, fieldName string, vals []string, cdrStats rpcclient.RpcClientConnection) (*RequestFilter, error) {
	if !utils.IsSliceMember([]string{MetaStringPrefix, MetaTimings, MetaRSRFields, MetaCDRStats, MetaDestinations}, rfType) {
		return nil, fmt.Errorf("Unsupported filter Type: %s", rfType)
	}
	if fieldName == "" && utils.IsSliceMember([]string{MetaStringPrefix, MetaTimings, MetaDestinations}, rfType) {
		return nil, fmt.Errorf("FieldName is mandatory for Type: %s", rfType)
	}
	if len(vals) == 0 && utils.IsSliceMember([]string{MetaStringPrefix, MetaTimings, MetaRSRFields, MetaDestinations}, rfType) {
		return nil, fmt.Errorf("Values is mandatory for Type: %s", rfType)
	}
	if rfType == MetaCDRStats && cdrStats == nil {
		return nil, errors.New("Missing cdrStats information")
	}
	rf := &RequestFilter{Type: rfType, FieldName: fieldName, Values: vals, cdrStats: cdrStats}
	if rfType == MetaRSRFields {
		var err error
		if rf.rsrFields, err = utils.ParseRSRFieldsFromSlice(vals); err != nil {
			return nil, err
		}
	}
	return rf, nil
}

// RequestFilter filters requests coming into various places
type RequestFilter struct {
	Type      string          // Filter type (*string, *timing, *rsr_filters, *cdr_stats)
	FieldName string          // Name of the field providing us the Values to check (used in case of some )
	Values    []string        // Filter definition
	rsrFields utils.RSRFields // Cache here the RSRFilter Values
	dataDB    AccountingStorage
	cdrStats  rpcclient.RpcClientConnection // Connection towards CDRStats service (eg: for *cdr_stats type)

}

func (fltr *RequestFilter) Pass(req interface{}, extraFieldsLabel string) (bool, error) {
	switch fltr.Type {
	case MetaStringPrefix:
		return fltr.passStringPrefix(req, extraFieldsLabel)
	case MetaTimings:
		return fltr.passTimings(req, extraFieldsLabel)
	case MetaDestinations:
		return fltr.passDestinations(req, extraFieldsLabel)
	case MetaRSRFields:
		return fltr.passRSRFields(req, extraFieldsLabel)
	case MetaCDRStats:
		return fltr.passCDRStats(req, extraFieldsLabel)
	default:
		return false, utils.ErrNotImplemented
	}
}

func (fltr *RequestFilter) passStringPrefix(req interface{}, extraFieldsLabel string) (bool, error) {
	strVal, err := utils.ReflectFieldAsString(req, fltr.FieldName, extraFieldsLabel)
	if err != nil {
		return false, err
	}
	matchedPrefix := false
	for _, prfx := range fltr.Values {
		if strings.HasPrefix(strVal, prfx) {
			matchedPrefix = true
			break
		}
	}
	return matchedPrefix, nil
}

// ToDo when Timings will be available in TPdb
func (fltr *RequestFilter) passTimings(req interface{}, extraFieldsLabel string) (bool, error) {
	return false, utils.ErrNotImplemented
}

// ToDo
func (fltr *RequestFilter) passDestinations(req interface{}, extraFieldsLabel string) (bool, error) {
	/*for _, p := range utils.SplitPrefix(cd.Destination, MIN_PREFIX_MATCH) {
	if x, err := CacheGet(utils.DESTINATION_PREFIX + p); err == nil {
		destIds := x.(map[string]struct{})
		var bestWeight float64
		for dID := range destIds {
			if _, ok := rpl.DestinationRates[dID]; ok {
				ril := rpl.RateIntervalList(dID)
				currentWeight := ril.GetWeight()
				if currentWeight > bestWeight {
					bestWeight = currentWeight
					rps = ril
					prefix = p
					destinationId = dID
				}
			}
		}
	}
	if rps != nil {
		break
	}
	*/
	return false, utils.ErrNotImplemented
}

func (fltr *RequestFilter) passRSRFields(req interface{}, extraFieldsLabel string) (bool, error) {
	for _, rsrFld := range fltr.rsrFields {
		if strVal, err := utils.ReflectFieldAsString(req, rsrFld.Id, extraFieldsLabel); err != nil {
			return false, err
		} else if !rsrFld.FilterPasses(strVal) {
			return false, nil
		}
	}
	return true, nil
}

// ToDo
func (fltr *RequestFilter) passCDRStats(req interface{}, extraFieldsLabel string) (bool, error) {
	return false, utils.ErrNotImplemented
}
