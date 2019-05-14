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

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

// SupplierReply represents one supplier in
type SortedSupplier struct {
	SupplierID         string
	SupplierParameters string
	SortingData        map[string]interface{} // store here extra info like cost or stats
}

// SuppliersReply is returned as part of GetSuppliers call
type SortedSuppliers struct {
	ProfileID       string            // Profile matched
	Sorting         string            // Sorting algorithm
	Count           int               // number of suppliers returned
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
	//sort suppliers
	sort.Slice(sSpls.SortedSuppliers, func(i, j int) bool {
		for _, param := range params {
			//in case we have the same value for the current param we skip to the next one
			if sSpls.SortedSuppliers[i].SortingData[param].(float64) == sSpls.SortedSuppliers[j].SortingData[param].(float64) {
				continue
			}
			switch param {
			default:
				if sSpls.SortedSuppliers[i].SortingData[param].(float64) > sSpls.SortedSuppliers[j].SortingData[param].(float64) {
					return true
				}
				return false
			case utils.MetaPDD: //in case of pdd the smalles value if the best
				if sSpls.SortedSuppliers[i].SortingData[param].(float64) < sSpls.SortedSuppliers[j].SortingData[param].(float64) {
					return true
				}
				return false
			}

		}
		//in case that we have the same value for all params we sort base on weight
		return sSpls.SortedSuppliers[i].SortingData[utils.Weight].(float64) > sSpls.SortedSuppliers[j].SortingData[utils.Weight].(float64)
	})
}

// SortResourceAscendent is part of sort interface,
// sort ascendent based on ResourceUsage with fallback on Weight
func (sSpls *SortedSuppliers) SortResourceAscendent() {
	sort.Slice(sSpls.SortedSuppliers, func(i, j int) bool {
		if sSpls.SortedSuppliers[i].SortingData[utils.ResourceUsage].(float64) == sSpls.SortedSuppliers[j].SortingData[utils.ResourceUsage].(float64) {
			return sSpls.SortedSuppliers[i].SortingData[utils.Weight].(float64) > sSpls.SortedSuppliers[j].SortingData[utils.Weight].(float64)
		}
		return sSpls.SortedSuppliers[i].SortingData[utils.ResourceUsage].(float64) < sSpls.SortedSuppliers[j].SortingData[utils.ResourceUsage].(float64)
	})
}

// SortResourceDescendent is part of sort interface,
// sort descendent based on ResourceUsage with fallback on Weight
func (sSpls *SortedSuppliers) SortResourceDescendent() {
	sort.Slice(sSpls.SortedSuppliers, func(i, j int) bool {
		if sSpls.SortedSuppliers[i].SortingData[utils.ResourceUsage].(float64) == sSpls.SortedSuppliers[j].SortingData[utils.ResourceUsage].(float64) {
			return sSpls.SortedSuppliers[i].SortingData[utils.Weight].(float64) > sSpls.SortedSuppliers[j].SortingData[utils.Weight].(float64)
		}
		return sSpls.SortedSuppliers[i].SortingData[utils.ResourceUsage].(float64) > sSpls.SortedSuppliers[j].SortingData[utils.ResourceUsage].(float64)
	})
}

// Digest returns list of supplierIDs + parameters for easier outside access
// format suppl1:suppl1params,suppl2:suppl2params
func (sSpls *SortedSuppliers) Digest() string {
	return strings.Join(sSpls.SuppliersWithParams(), utils.FIELDS_SEP)
}

func (sSpls *SortedSuppliers) AsNavigableMap() (nm *config.NavigableMap) {
	mp := map[string]interface{}{
		"ProfileID": sSpls.ProfileID,
		"Sorting":   sSpls.Sorting,
		"Count":     sSpls.Count,
	}
	sm := make([]map[string]interface{}, len(sSpls.SortedSuppliers))
	for i, ss := range sSpls.SortedSuppliers {
		sm[i] = map[string]interface{}{
			"SupplierID":         ss.SupplierID,
			"SupplierParameters": ss.SupplierParameters,
			"SortingData":        ss.SortingData,
		}
	}
	mp["SortedSuppliers"] = sm
	return config.NewNavigableMap(mp)
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
	ssd[utils.MetaReas] = NewResourceAscendetSorter(lcrS)
	ssd[utils.MetaReds] = NewResourceDescendentSorter(lcrS)
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
