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
	"strconv"
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
	MetaMinCapPrefix = "*MIN_"
	MetaMaxCapPrefix = "*MAX_"
)

func NewRequestFilter(rfType, fieldName string, vals []string, cdrStats rpcclient.RpcClientConnection) (*RequestFilter, error) {
	if !utils.IsSliceMember([]string{MetaStringPrefix, MetaTimings, MetaRSRFields, MetaCDRStats, MetaDestinations}, rfType) {
		return nil, fmt.Errorf("Unsupported filter Type: %s", rfType)
	}
	if fieldName == "" && utils.IsSliceMember([]string{MetaStringPrefix, MetaTimings, MetaDestinations}, rfType) {
		return nil, fmt.Errorf("FieldName is mandatory for Type: %s", rfType)
	}
	if len(vals) == 0 && utils.IsSliceMember([]string{MetaStringPrefix, MetaTimings, MetaRSRFields, MetaDestinations, MetaDestinations}, rfType) {
		return nil, fmt.Errorf("Values is mandatory for Type: %s", rfType)
	}
	rf := &RequestFilter{Type: rfType, FieldName: fieldName, Values: vals, cdrStats: cdrStats, cdrStatSThresholds: make([]*RFStatSThreshold, len(vals))}
	if rfType == MetaCDRStats {
		if cdrStats == nil {
			return nil, errors.New("Missing cdrStats information")
		}
		for i, val := range vals {
			valSplt := strings.Split(val, utils.InInFieldSep)
			if len(valSplt) != 3 {
				return nil, fmt.Errorf("Value %s needs to contain at least 3 items", val)
			}
			st := &RFStatSThreshold{QueueID: valSplt[0], ThresholdType: strings.ToUpper(valSplt[1])}
			if len(st.ThresholdType) < len(MetaMinCapPrefix)+1 {
				return nil, fmt.Errorf("Value %s contains a unsupported ThresholdType format", val)
			} else if !strings.HasPrefix(st.ThresholdType, MetaMinCapPrefix) && !strings.HasPrefix(st.ThresholdType, MetaMaxCapPrefix) {
				return nil, fmt.Errorf("Value %s contains unsupported ThresholdType prefix", val)
			}
			if tv, err := strconv.ParseFloat(valSplt[2], 64); err != nil {
				return nil, err
			} else {
				st.ThresholdValue = tv
			}
			rf.cdrStatSThresholds[i] = st
		}
	}
	if rfType == MetaRSRFields {
		var err error
		if rf.rsrFields, err = utils.ParseRSRFieldsFromSlice(vals); err != nil {
			return nil, err
		}
	}
	return rf, nil
}

type RFStatSThreshold struct {
	QueueID        string
	ThresholdType  string
	ThresholdValue float64
}

// RequestFilter filters requests coming into various places
// Pass rule: default negative, one mathing rule should pass the filter
type RequestFilter struct {
	Type               string                        // Filter type (*string, *timing, *rsr_filters, *cdr_stats)
	FieldName          string                        // Name of the field providing us the Values to check (used in case of some )
	Values             []string                      // Filter definition
	rsrFields          utils.RSRFields               // Cache here the RSRFilter Values
	cdrStats           rpcclient.RpcClientConnection // Connection towards CDRStats service (eg: for *cdr_stats type)
	cdrStatSThresholds []*RFStatSThreshold           // Cached compiled RFStatsThreshold out of Values
}

// Pass is the method which should be used from outside.
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
	for _, prfx := range fltr.Values {
		if strings.HasPrefix(strVal, prfx) {
			return true, nil
		}
	}
	return false, nil
}

// ToDo when Timings will be available in TPdb
func (fltr *RequestFilter) passTimings(req interface{}, extraFieldsLabel string) (bool, error) {
	return false, utils.ErrNotImplemented
}

func (fltr *RequestFilter) passDestinations(req interface{}, extraFieldsLabel string) (bool, error) {
	dst, err := utils.ReflectFieldAsString(req, fltr.FieldName, extraFieldsLabel)
	if err != nil {
		return false, err
	}
	for _, p := range utils.SplitPrefix(dst, MIN_PREFIX_MATCH) {
		if x, ok := CacheGet(utils.DESTINATION_PREFIX + p); ok {
			destIds := x.(map[string]struct{})
			for dID := range destIds {
				for _, valDstID := range fltr.Values {
					if valDstID == dID {
						return true, nil
					}
				}
			}
		}
	}
	return false, nil
}

func (fltr *RequestFilter) passRSRFields(req interface{}, extraFieldsLabel string) (bool, error) {
	for _, rsrFld := range fltr.rsrFields {
		if strVal, err := utils.ReflectFieldAsString(req, rsrFld.Id, extraFieldsLabel); err != nil {
			return false, err
		} else if rsrFld.FilterPasses(strVal) {
			return true, nil
		}
	}
	return false, nil
}

func (fltr *RequestFilter) passCDRStats(req interface{}, extraFieldsLabel string) (bool, error) {
	for _, threshold := range fltr.cdrStatSThresholds {
		statValues := make(map[string]float64)
		if err := fltr.cdrStats.Call("CDRStatsV1.GetValues", threshold.QueueID, &statValues); err != nil {
			return false, err
		}
		if val, hasIt := statValues[threshold.ThresholdType[len(MetaMinCapPrefix):]]; !hasIt {
			continue
		} else if strings.HasPrefix(threshold.ThresholdType, MetaMinCapPrefix) && val >= threshold.ThresholdValue {
			return true, nil
		} else if strings.HasPrefix(threshold.ThresholdType, MetaMaxCapPrefix) && val < threshold.ThresholdValue {
			return true, nil
		}
	}
	return false, nil
}
