// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package engine

import (
	"github.com/cgrates/cgrates/utils"
)

func NewQOSSupplierSorter(spS *SupplierService) *QOSSupplierSorter {
	return &QOSSupplierSorter{spS: spS,
		sorting: utils.MetaQOS}
}

// QOSSorter sorts suppliers based on stats
type QOSSupplierSorter struct {
	sorting string
	spS     *SupplierService
}

func (qos *QOSSupplierSorter) SortSuppliers(prflID string, suppls []*Supplier,
	ev *utils.CGREvent, extraOpts *optsGetSuppliers, argDsp *utils.ArgDispatcher) (sortedSuppls *SortedSuppliers, err error) {
	sortedSuppls = &SortedSuppliers{ProfileID: prflID,
		Sorting:         qos.sorting,
		SortedSuppliers: make([]*SortedSupplier, 0)}
	for _, s := range suppls {
		if srtSpl, pass, err := qos.spS.populateSortingData(ev, s, extraOpts, argDsp); err != nil {
			return nil, err
		} else if pass && srtSpl != nil {
			sortedSuppls.SortedSuppliers = append(sortedSuppls.SortedSuppliers, srtSpl)
		}
	}
	if len(sortedSuppls.SortedSuppliers) == 0 {
		return nil, utils.ErrNotFound
	}
	sortedSuppls.SortQOS(extraOpts.sortingParameters)
	return
}
