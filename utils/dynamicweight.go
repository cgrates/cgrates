// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package utils

import (
	"fmt"
	"strconv"
	"strings"
)

// NewDynamicWeightsFromString creates a DynamicWeight list based on the string received from .csv
func NewDynamicWeightsFromString(s, dWSep, fltrSep string) (dWs DynamicWeights, err error) {
	if len(s) == 0 {
		return DynamicWeights{{}}, nil
	}
	dwStrs := strings.Split(s, dWSep)
	lnDwStrs := len(dwStrs)
	if lnDwStrs%2 != 0 { // need to have multiples of number of fields in one DynamicWeight
		return nil, fmt.Errorf("invalid DynamicWeight format for string <%s>", s)
	}
	dWs = make([]*DynamicWeight, 0, lnDwStrs/2)
	for i := 0; i < lnDwStrs; i += 2 {
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

type DynamicWeights []*DynamicWeight

func (dWS DynamicWeights) String(dWSep, fltrsep string) (out string) {
	if len(dWS) == 0 {
		return
	}
	dwToString := make([]string, len(dWS))
	for i, value := range dWS {
		dwToString[i] = value.String(dWSep, fltrsep)
	}
	return strings.Join(dwToString, dWSep)
}

func (dW DynamicWeight) String(dWSep, fltrsep string) (out string) {
	return strings.Join(dW.FilterIDs, fltrsep) + dWSep + strconv.FormatFloat(dW.Weight, 'f', -1, 64)
}

func (dW *DynamicWeight) Equals(dnWg *DynamicWeight) (eq bool) {
	if dW.FilterIDs == nil && dnWg.FilterIDs != nil ||
		dW.FilterIDs != nil && dnWg.FilterIDs == nil ||
		len(dW.FilterIDs) != len(dnWg.FilterIDs) ||
		dW.Weight != dnWg.Weight {
		return
	}
	for i := range dW.FilterIDs {
		if dW.FilterIDs[i] != dnWg.FilterIDs[i] {
			return
		}
	}
	return true
}

// DynamicWeight returns Weight based on Filters
type DynamicWeight struct {
	FilterIDs []string
	Weight    float64
}

func (dW *DynamicWeight) Clone() (cln *DynamicWeight) {
	if dW == nil {
		return nil
	}
	cln = &DynamicWeight{
		Weight: dW.Weight,
	}
	if dW.FilterIDs != nil {
		cln.FilterIDs = make([]string, len(dW.FilterIDs))
		copy(cln.FilterIDs, dW.FilterIDs)
	}
	return cln
}

func (dW *DynamicWeight) FieldAsInterface(fldPath []string) (any, error) {
	if len(fldPath) != 1 {
		return nil, ErrNotFound
	}
	switch fldPath[0] {
	case FilterIDs:
		return dW.FilterIDs, nil
	case Weight:
		return dW.Weight, nil
	default:
		fld, idx := GetPathIndex(fldPath[0])
		if idx != nil &&
			fld == FilterIDs {
			if *idx < len(dW.FilterIDs) {
				return dW.FilterIDs[*idx], nil
			}
		}
	}
	return nil, ErrNotFound
}

func (dW DynamicWeights) Clone() (dinWeight DynamicWeights) {
	if dW == nil {
		return
	}
	dinWeight = make(DynamicWeights, len(dW))
	for i, value := range dW {
		dinWeight[i] = value.Clone()
	}
	return
}
