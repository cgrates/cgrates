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

package engine

import (
	"fmt"
	"sort"
	"strings"

	"github.com/cgrates/cgrates/utils"
)

// SupplierReply represents one supplier in
type SortedSupplier struct {
	SupplierID         string
	SupplierParameters string
	SortingData        map[string]interface{} // store here extra info like cost or stats
	globalStats        map[string]float64
}

// SuppliersReply is returned as part of GetSuppliers call
type SortedSuppliers struct {
	ProfileID       string            // Profile matched
	Sorting         string            // Sorting algorithm
	SortedSuppliers []*SortedSupplier // list of supplier IDs and SortingData data
}

// SupplierIDs returns list of suppliers
func (sSpls *SortedSuppliers) SupplierIDs() (sIDs []string) {
	sIDs = make([]string, len(sSpls.SortedSuppliers))
	for i, spl := range sSpls.SortedSuppliers {
		sIDs[i] = spl.SupplierID
	}
	return
}

// SupplierIDs returns list of suppliers
func (sSpls *SortedSuppliers) SuppliersWithParams() (sPs []string) {
	sPs = make([]string, len(sSpls.SortedSuppliers))
	for i, spl := range sSpls.SortedSuppliers {
		sPs[i] = spl.SupplierID
		if spl.SupplierParameters != "" {
			sPs[i] += utils.InInFieldSep + spl.SupplierParameters
		}
	}
	return
}

// SortWeight is part of sort interface, sort based on Weight
func (sSpls *SortedSuppliers) SortWeight() {
	sort.Slice(sSpls.SortedSuppliers, func(i, j int) bool {
		return sSpls.SortedSuppliers[i].SortingData[utils.Weight].(float64) > sSpls.SortedSuppliers[j].SortingData[utils.Weight].(float64)
	})
}

// SortLeastCost is part of sort interface,
// sort ascendent based on Cost with fallback on Weight
func (sSpls *SortedSuppliers) SortLeastCost() {
	sort.Slice(sSpls.SortedSuppliers, func(i, j int) bool {
		if sSpls.SortedSuppliers[i].SortingData[utils.Cost].(float64) == sSpls.SortedSuppliers[j].SortingData[utils.Cost].(float64) {
			return sSpls.SortedSuppliers[i].SortingData[utils.Weight].(float64) > sSpls.SortedSuppliers[j].SortingData[utils.Weight].(float64)
		}
		return sSpls.SortedSuppliers[i].SortingData[utils.Cost].(float64) < sSpls.SortedSuppliers[j].SortingData[utils.Cost].(float64)
	})
}

// SortHighestCost is part of sort interface,
// sort descendent based on Cost with fallback on Weight
func (sSpls *SortedSuppliers) SortHighestCost() {
	sort.Slice(sSpls.SortedSuppliers, func(i, j int) bool {
		if sSpls.SortedSuppliers[i].SortingData[utils.Cost].(float64) == sSpls.SortedSuppliers[j].SortingData[utils.Cost].(float64) {
			return sSpls.SortedSuppliers[i].SortingData[utils.Weight].(float64) > sSpls.SortedSuppliers[j].SortingData[utils.Weight].(float64)
		}
		return sSpls.SortedSuppliers[i].SortingData[utils.Cost].(float64) > sSpls.SortedSuppliers[j].SortingData[utils.Cost].(float64)
	})
}

// SortQOS is part of sort interface,
// sort based on Stats
func (sSpls *SortedSuppliers) SortQOS(params []string) {
	sort.Slice(sSpls.SortedSuppliers, func(i, j int) bool {
		for _, param := range params {
			// skip to next param
			if sSpls.SortedSuppliers[i].globalStats[param] == sSpls.SortedSuppliers[j].globalStats[param] {
				continue
			}
			if (param != utils.MetaPDD && sSpls.SortedSuppliers[i].globalStats[param] == -1) ||
				(param == utils.MetaPDD && sSpls.SortedSuppliers[i].globalStats[param] == 1000000) {
				return false
			}
			switch param {
			default:
				return sSpls.SortedSuppliers[i].globalStats[param] > sSpls.SortedSuppliers[j].globalStats[param]
			case utils.MetaPDD:
				return sSpls.SortedSuppliers[i].globalStats[param] < sSpls.SortedSuppliers[j].globalStats[param]
			}
		}
		return sSpls.SortedSuppliers[i].SortingData[utils.Weight].(float64) > sSpls.SortedSuppliers[j].SortingData[utils.Weight].(float64)
	})
}

// Digest returns list of supplierIDs + parameters for easier outside access
// format suppl1:suppl1params,suppl2:suppl2params
func (sSpls *SortedSuppliers) Digest() string {
	return strings.Join(sSpls.SuppliersWithParams(), utils.FIELDS_SEP)
}

type SupplierWithParams struct {
	SupplierName   string
	SupplierParams string
}

// SuppliersSorter is the interface which needs to be implemented by supplier sorters
type SuppliersSorter interface {
	SortSuppliers(string, []*Supplier, *utils.CGREvent, *optsGetSuppliers) (*SortedSuppliers, error)
}

// NewSupplierSortDispatcher constructs SupplierSortDispatcher
func NewSupplierSortDispatcher(lcrS *SupplierService) (ssd SupplierSortDispatcher, err error) {
	ssd = make(map[string]SuppliersSorter)
	ssd[utils.MetaWeight] = NewWeightSorter(lcrS)
	ssd[utils.MetaLeastCost] = NewLeastCostSorter(lcrS)
	ssd[utils.MetaHighestCost] = NewHighestCostSorter(lcrS)
	ssd[utils.MetaQOS] = NewQOSSupplierSorter(lcrS)
	return
}

// SupplierStrategyHandler will initialize strategies
// and dispatch requests to them
type SupplierSortDispatcher map[string]SuppliersSorter

func (ssd SupplierSortDispatcher) SortSuppliers(prflID, strategy string,
	suppls []*Supplier, suplEv *utils.CGREvent, extraOpts *optsGetSuppliers) (sortedSuppls *SortedSuppliers, err error) {
	sd, has := ssd[strategy]
	if !has {
		return nil, fmt.Errorf("unsupported sorting strategy: %s", strategy)
	}
	return sd.SortSuppliers(prflID, suppls, suplEv, extraOpts)
}
