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

	"github.com/cgrates/cgrates/utils"
)

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
	suppls Suppliers) (sortedSuppls *SortedSuppliers, err error) {
	fmt.Printf("Sort strategy: %s, suppliers: %s\n", strategy, utils.ToJSON(suppls))
	sd, has := ssd[strategy]
	if !has {
		return nil, fmt.Errorf("unsupported sorting strategy: %s", strategy)
	}
	return sd.SortSuppliers(prflID, suppls)
}

type SuppliersSorter interface {
	SortSuppliers(string, Suppliers) (*SortedSuppliers, error)
}

// NewLeastCostSorter constructs LeastCostSorter
func NewLeastCostSorter(lcrS *SupplierService) *LeastCostSorter {
	return &LeastCostSorter{lcrS: lcrS}
}

// LeastCostSorter orders suppliers based on lowest cost
type LeastCostSorter struct {
	lcrS *SupplierService
}

func (lcs *LeastCostSorter) SortSuppliers(prflID string,
	suppls Suppliers) (sortedSuppls *SortedSuppliers, err error) {
	return
}

func NewWeightSorter() *WeightSorter {
	return &WeightSorter{Sorting: utils.MetaWeight}
}

// WeightSorter orders suppliers based on their weight, no cost involved
type WeightSorter struct {
	Sorting string
}

func (ws *WeightSorter) SortSuppliers(prflID string,
	suppls Suppliers) (sortedSuppls *SortedSuppliers, err error) {
	fmt.Printf("Sort suppliers: %s\n", utils.ToJSON(suppls))
	suppls.Sort()
	sortedSuppls = &SortedSuppliers{ProfileID: prflID,
		Sorting:         ws.Sorting,
		SortedSuppliers: make([]*SortedSupplier, len(suppls))}
	for i, s := range suppls {
		sortedSuppls.SortedSuppliers[i] = &SortedSupplier{
			SupplierID:  s.ID,
			SortingData: map[string]interface{}{"Weight": s.Weight}}
	}
	return
}
