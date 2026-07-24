// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package engine

import (
	"fmt"

	"github.com/cgrates/cgrates/utils"
)

func NewHighestCostSorter(spS *SupplierService) *HightCostSorter {
	return &HightCostSorter{spS: spS,
		sorting: utils.MetaHC}
}

// HightCostSorter sorts suppliers based on their cost
type HightCostSorter struct {
	sorting string
	spS     *SupplierService
}

func (hcs *HightCostSorter) SortSuppliers(prflID string, suppls []*Supplier,
	ev *utils.CGREvent, extraOpts *optsGetSuppliers, argDsp *utils.ArgDispatcher) (sortedSuppls *SortedSuppliers, err error) {
	sortedSuppls = &SortedSuppliers{ProfileID: prflID,
		Sorting:         hcs.sorting,
		SortedSuppliers: make([]*SortedSupplier, 0)}
	for _, s := range suppls {
		if len(s.RatingPlanIDs) == 0 {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> supplier: <%s> - empty RatingPlanIDs",
					utils.SupplierS, s.ID))
			return nil, utils.NewErrMandatoryIeMissing("RatingPlanIDs")
		}
		if srtSpl, pass, err := hcs.spS.populateSortingData(ev, s, extraOpts, argDsp); err != nil {
			return nil, err
		} else if pass && srtSpl != nil {
			sortedSuppls.SortedSuppliers = append(sortedSuppls.SortedSuppliers, srtSpl)
		}
	}
	if len(sortedSuppls.SortedSuppliers) == 0 {
		return nil, utils.ErrNotFound
	}
	sortedSuppls.SortHighestCost()
	return
}
