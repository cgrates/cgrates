// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package utils

import (
	"fmt"
	"strconv"
	"strings"
)

// NewDynamicWeightsFromString creates a DynamicWeight list based on the string received from .csv/StorDB
func NewDynamicWeightsFromString(s, dWSep, fltrSep string) (dWs DynamicWeights, err error) {
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

// DynamicWeight returns Weight based on Filters
type DynamicWeight struct {
	FilterIDs []string
	Weight    float64
}

func (dW *DynamicWeight) Clone() (dinWeight *DynamicWeight) {
	dinWeight = &DynamicWeight{
		Weight: dW.Weight,
	}
	if dW.FilterIDs != nil {
		dinWeight.FilterIDs = make([]string, len(dW.FilterIDs))
		copy(dinWeight.FilterIDs, dW.FilterIDs)
	}
	return dinWeight
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
