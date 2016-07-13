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
	"strings"

	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

const (
	MetaStringPrefix = "*string_prefix"
	MetaTiming       = "*timing"
	MetaRSRFilters   = "*rsr_filters"
	MetaCDRStats     = "*cdr_stats"
	MetaDestinations = "*destinations"
)

// RequestFilter filters requests coming into various places
type RequestFilter struct {
	Type        string                        // Filter type (*string, *timing, *rsr_filters, *cdr_stats)
	FieldName   string                        // Name of the field providing us the value to check
	Values      string                        // Filter definition
	rsrFltrVals []*utils.RSRFilter            // Cache here the RSRFilter values
	cdrStats    rpcclient.RpcClientConnection // Connection towards CDRStats service (eg: for *cdr_stats type)
}

func (fltr *RequestFilter) Pass(req interface{}, extraFieldsLabel string) (bool, error) {
	switch fltr.Type {
	case MetaStringPrefix:
		return fltr.passStringPrefix(req, extraFieldsLabel)
	case MetaTiming:
		return fltr.passTiming(req, extraFieldsLabel)
	case MetaDestinations:
		return fltr.passDestinations(req, extraFieldsLabel)
	case MetaRSRFilters:
		return fltr.passRSRFilters(req, extraFieldsLabel)
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
	for _, prfx := range strings.Split(fltr.Values, utils.INFIELD_SEP) {
		if strings.HasPrefix(strVal, prfx) {
			matchedPrefix = true
			break
		}
	}
	return matchedPrefix, nil
}

// ToDo
func (fltr *RequestFilter) passTiming(req interface{}, extraFieldsLabel string) (bool, error) {
	return false, utils.ErrNotImplemented
}

// ToDo
func (fltr *RequestFilter) passDestinations(req interface{}, extraFieldsLabel string) (bool, error) {
	return false, utils.ErrNotImplemented
}

// ToDo
func (fltr *RequestFilter) passRSRFilters(req interface{}, extraFieldsLabel string) (bool, error) {
	return false, utils.ErrNotImplemented
}

// ToDo
func (fltr *RequestFilter) passCDRStats(req interface{}, extraFieldsLabel string) (bool, error) {
	return false, utils.ErrNotImplemented
}
