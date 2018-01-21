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
	// "time"

	"github.com/cgrates/cgrates/utils"
)

// SupplierReply represents one supplier in
type SortedSupplier struct {
	SupplierID         string
	SortingData        map[string]interface{} // store here extra info like cost or stats
	SupplierParameters string
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

// SortWeight is part of sort interface, sort based on Weight
func (sSpls *SortedSuppliers) SortWeight() {
	sort.Slice(sSpls.SortedSuppliers, func(i, j int) bool {
		return sSpls.SortedSuppliers[i].SortingData[utils.Weight].(float64) > sSpls.SortedSuppliers[j].SortingData[utils.Weight].(float64)
	})
}

// SortCost is part of sort interface,
// sort based on Cost with fallback on Weight
func (sSpls *SortedSuppliers) SortCost() {
	sort.Slice(sSpls.SortedSuppliers, func(i, j int) bool {
		if sSpls.SortedSuppliers[i].SortingData[utils.Cost].(float64) == sSpls.SortedSuppliers[j].SortingData[utils.Cost].(float64) {
			return sSpls.SortedSuppliers[i].SortingData[utils.Weight].(float64) > sSpls.SortedSuppliers[j].SortingData[utils.Weight].(float64)
		}
		return sSpls.SortedSuppliers[i].SortingData[utils.Cost].(float64) < sSpls.SortedSuppliers[j].SortingData[utils.Cost].(float64)
	})
}

// Digest returns list of supplierIDs + parameters for easier outside access
func (sSpls *SortedSuppliers) SuppliersDigest() (sDigest []string) {
	sDigest = make([]string, len(sSpls.SortedSuppliers))
	for i, sSpl := range sSpls.SortedSuppliers {
		sDigest[i] = sSpl.SupplierID
		if sSpl.SupplierParameters != "" {
			sDigest[i] += utils.InInFieldSep + sSpl.SupplierParameters
		}

	}
	return
}

// SuppliersSorter is the interface which needs to be implemented by supplier sorters
type SuppliersSorter interface {
	SortSuppliers(string, []*Supplier, *utils.CGREvent) (*SortedSuppliers, error)
}

// NewSupplierSortDispatcher constructs SupplierSortDispatcher
func NewSupplierSortDispatcher(lcrS *SupplierService) (ssd SupplierSortDispatcher, err error) {
	ssd = make(map[string]SuppliersSorter)
	ssd[utils.MetaWeight] = NewWeightSorter()
	ssd[utils.MetaLeastCost] = NewLeastCostSorter(lcrS)
	return
}

// SupplierStrategyHandler will initialize strategies
// and dispatch requests to them
type SupplierSortDispatcher map[string]SuppliersSorter

func (ssd SupplierSortDispatcher) SortSuppliers(prflID, strategy string,
	suppls []*Supplier, suplEv *utils.CGREvent) (sortedSuppls *SortedSuppliers, err error) {
	sd, has := ssd[strategy]
	if !has {
		return nil, fmt.Errorf("unsupported sorting strategy: %s", strategy)
	}
	return sd.SortSuppliers(prflID, suppls, suplEv)
}

func NewWeightSorter() *WeightSorter {
	return &WeightSorter{sorting: utils.MetaWeight}
}

// WeightSorter orders suppliers based on their weight, no cost involved
type WeightSorter struct {
	sorting string
}

func (ws *WeightSorter) SortSuppliers(prflID string,
	suppls []*Supplier, suplEv *utils.CGREvent) (sortedSuppls *SortedSuppliers, err error) {
	sortedSuppls = &SortedSuppliers{ProfileID: prflID,
		Sorting:         ws.sorting,
		SortedSuppliers: make([]*SortedSupplier, len(suppls))}
	for i, s := range suppls {
		sortedSuppls.SortedSuppliers[i] = &SortedSupplier{
			SupplierID:         s.ID,
			SortingData:        map[string]interface{}{utils.Weight: s.Weight},
			SupplierParameters: s.SupplierParameters}
	}
	sortedSuppls.SortWeight()
	return
}
