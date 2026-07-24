// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package engine

import (
	"fmt"

	"github.com/cgrates/cgrates/utils"
)

func NewResourceDescendentSorter(spS *SupplierService) *ResourceDescendentSorter {
	return &ResourceDescendentSorter{spS: spS,
		sorting: utils.MetaReds}
}

// ResourceAscendentSorter orders suppliers based on their Resource Usage
type ResourceDescendentSorter struct {
	sorting string
	spS     *SupplierService
}

func (ws *ResourceDescendentSorter) SortSuppliers(prflID string,
	suppls []*Supplier, suplEv *utils.CGREvent, extraOpts *optsGetSuppliers, argDsp *utils.ArgDispatcher) (sortedSuppls *SortedSuppliers, err error) {
	sortedSuppls = &SortedSuppliers{ProfileID: prflID,
		Sorting:         ws.sorting,
		SortedSuppliers: make([]*SortedSupplier, 0)}
	for _, s := range suppls {
		if len(s.ResourceIDs) == 0 {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> supplier: <%s> - empty ResourceIDs",
					utils.SupplierS, s.ID))
			return nil, utils.NewErrMandatoryIeMissing("ResourceIDs")
		}
		if srtSpl, pass, err := ws.spS.populateSortingData(suplEv, s, extraOpts, argDsp); err != nil {
			return nil, err
		} else if pass && srtSpl != nil {
			sortedSuppls.SortedSuppliers = append(sortedSuppls.SortedSuppliers, srtSpl)
		}
	}
	sortedSuppls.SortResourceDescendent()
	return
}
