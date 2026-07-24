// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package engine

import (
	"github.com/cgrates/cgrates/utils"
)

func NewWeightSorter(spS *SupplierService) *WeightSorter {
	return &WeightSorter{spS: spS,
		sorting: utils.MetaWeight}
}

// WeightSorter orders suppliers based on their weight, no cost involved
type WeightSorter struct {
	sorting string
	spS     *SupplierService
}

func (ws *WeightSorter) SortSuppliers(prflID string,
	suppls []*Supplier, suplEv *utils.CGREvent, extraOpts *optsGetSuppliers, argDsp *utils.ArgDispatcher) (sortedSuppls *SortedSuppliers, err error) {
	sortedSuppls = &SortedSuppliers{ProfileID: prflID,
		Sorting:         ws.sorting,
		SortedSuppliers: make([]*SortedSupplier, 0)}
	for _, s := range suppls {
		if srtSpl, pass, err := ws.spS.populateSortingData(suplEv, s, extraOpts, argDsp); err != nil {
			return nil, err
		} else if pass && srtSpl != nil {
			sortedSuppls.SortedSuppliers = append(sortedSuppls.SortedSuppliers, srtSpl)
		}
	}
	sortedSuppls.SortWeight()
	return
}
