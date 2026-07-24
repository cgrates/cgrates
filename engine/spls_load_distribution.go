// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package engine

import (
	"fmt"

	"github.com/cgrates/cgrates/utils"
)

func NewLoadDistributionSorter(spS *SupplierService) *LoadDistributionSorter {
	return &LoadDistributionSorter{spS: spS,
		sorting: utils.MetaLoad}
}

// ResourceAscendentSorter orders suppliers based on their Resource Usage
type LoadDistributionSorter struct {
	sorting string
	spS     *SupplierService
}

func (ws *LoadDistributionSorter) SortSuppliers(prflID string,
	suppls []*Supplier, suplEv *utils.CGREvent, extraOpts *optsGetSuppliers, argDsp *utils.ArgDispatcher) (sortedSuppls *SortedSuppliers, err error) {
	sortedSuppls = &SortedSuppliers{ProfileID: prflID,
		Sorting:         ws.sorting,
		SortedSuppliers: make([]*SortedSupplier, 0)}
	for _, s := range suppls {
		// we should have at least 1 statID defined for counting CDR (a.k.a *sum:1)
		if len(s.StatIDs) == 0 {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> supplier: <%s> - empty StatIDs",
					utils.SupplierS, s.ID))
			return nil, utils.NewErrMandatoryIeMissing("StatIDs")
		}
		if srtSpl, pass, err := ws.spS.populateSortingData(suplEv, s, extraOpts, argDsp); err != nil {
			return nil, err
		} else if pass && srtSpl != nil {
			// Add the ratio in SortingData so we can used it later in SortLoadDistribution
			srtSpl.SortingData[utils.Ratio] = s.cacheSupplier[utils.MetaRatio].(float64)
			sortedSuppls.SortedSuppliers = append(sortedSuppls.SortedSuppliers, srtSpl)
		}
	}
	sortedSuppls.SortLoadDistribution()
	return
}
